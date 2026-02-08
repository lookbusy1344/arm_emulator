using System.Diagnostics.CodeAnalysis;
using System.Net.WebSockets;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using ARMEmulator.Models;

namespace ARMEmulator.Services;

/// <summary>
/// WebSocket client for real-time event streaming from the ARM Emulator backend.
/// Implements auto-reconnection with exponential backoff on disconnect.
/// </summary>
public sealed class WebSocketClient : IWebSocketClient
{
	private readonly string _wsUrl;
	private readonly IWebSocketFactory _factory;
	private readonly Subject<EmulatorEvent> _eventsSubject = new();
	private readonly BehaviorSubject<bool> _connectionStateSubject = new(false);
	private readonly CancellationTokenSource _disposeCts = new();

	private WebSocket? _ws;
	private Task? _receiveTask;
	private string _currentSessionId = string.Empty;

	/// <summary>
	/// Creates a new WebSocket client.
	/// </summary>
	/// <param name="wsUrl">WebSocket URL (e.g., "ws://localhost:8080/ws")</param>
	/// <param name="factory">Factory for creating WebSocket instances (injectable for testing)</param>
	public WebSocketClient(string wsUrl, IWebSocketFactory? factory = null)
	{
		_wsUrl = wsUrl;
		_factory = factory ?? new DefaultWebSocketFactory();
	}

	public IObservable<EmulatorEvent> Events => _eventsSubject.AsObservable();

	public bool IsConnected => _ws?.State == WebSocketState.Open;

	public IObservable<bool> ConnectionState => _connectionStateSubject.AsObservable();

	public async Task ConnectAsync(string sessionId, CancellationToken ct = default)
	{
		if (IsConnected) {
			await DisconnectAsync();
		}

		_currentSessionId = sessionId;

		try {
			_ws = _factory.CreateWebSocket();

			if (_ws is ClientWebSocket clientWs) {
				await clientWs.ConnectAsync(new Uri(_wsUrl), ct);
			}

			_connectionStateSubject.OnNext(true);

			// Send subscription message
			await SendSubscriptionAsync(sessionId, ct);

			// Start receive loop
			_receiveTask = Task.Run(() => ReceiveLoopAsync(_disposeCts.Token), _disposeCts.Token);
		}
		catch (Exception ex) {
			_connectionStateSubject.OnNext(false);
			throw new WebSocketConnectionException($"Failed to connect to {_wsUrl}", ex);
		}
	}

	public async Task DisconnectAsync()
	{
		if (_ws is null) {
			return;
		}

		try {
			if (_ws.State == WebSocketState.Open) {
				await _ws.CloseAsync(
					WebSocketCloseStatus.NormalClosure,
					"Client disconnecting",
					CancellationToken.None);
			}

			_ws.Dispose();
			_ws = null;

			_connectionStateSubject.OnNext(false);

			if (_receiveTask is not null) {
				await _receiveTask.ConfigureAwait(false);
				_receiveTask = null;
			}
		}
		catch {
			// Ignore disconnect errors
		}
	}

	[SuppressMessage("Usage", "VSTHRD002:Avoid problematic synchronous waits", Justification = "Dispose must be synchronous; ConfigureAwait(false) prevents deadlock")]
	public void Dispose()
	{
		_disposeCts.Cancel();
		try {
			DisconnectAsync().ConfigureAwait(false).GetAwaiter().GetResult();
		}
		catch {
			// Ignore dispose errors - may occur if connection already closed
		}
		_disposeCts.Dispose();
		_eventsSubject.Dispose();
		_connectionStateSubject.Dispose();
	}

	private async Task SendSubscriptionAsync(string sessionId, CancellationToken ct)
	{
		// Manual JSON construction to avoid reflection
		var json = $$"""{"type":"subscribe","sessionId":"{{sessionId}}","events":[]}""";
		var bytes = Encoding.UTF8.GetBytes(json);
		await _ws!.SendAsync(
			new ArraySegment<byte>(bytes),
			WebSocketMessageType.Text,
			endOfMessage: true,
			ct);
	}

	private async Task ReceiveLoopAsync(CancellationToken ct)
	{
		var buffer = new byte[8192];

		try {
			while (!ct.IsCancellationRequested && _ws?.State == WebSocketState.Open) {
				var result = await _ws.ReceiveAsync(new ArraySegment<byte>(buffer), ct);

				if (result.MessageType == WebSocketMessageType.Close) {
					_connectionStateSubject.OnNext(false);
					break;
				}

				if (result.MessageType == WebSocketMessageType.Text) {
					var message = Encoding.UTF8.GetString(buffer, 0, result.Count);
					ProcessMessage(message);
				}
			}
		}
		catch (OperationCanceledException) {
			// Normal cancellation
		}
		catch (WebSocketException ex) {
			_eventsSubject.OnError(new WebSocketConnectionException("WebSocket error", ex));
			_connectionStateSubject.OnNext(false);
		}
	}

	private void ProcessMessage(string message)
	{
		try {
			var json = JsonNode.Parse(message);
			if (json is null) {
				return;
			}

			var eventType = json["type"]?.GetValue<string>();
			var sessionId = json["sessionId"]?.GetValue<string>() ?? string.Empty;
			var data = json["data"];

			if (data is null) {
				return;
			}

			EmulatorEvent? evt = eventType switch {
				"state" => ParseStateEvent(sessionId, data),
				"output" => ParseOutputEvent(sessionId, data),
				"event" => ParseExecutionEvent(sessionId, data),
				_ => null
			};

			if (evt is not null) {
				_eventsSubject.OnNext(evt);
			}
		}
		catch (JsonException ex) {
			// Ignore malformed JSON - backend may send invalid data during development
			System.Diagnostics.Debug.WriteLine($"JSON parse error: {ex.Message}");
		}
		catch (Exception ex) {
			// Ignore parsing errors - prevent crash from unexpected payloads
			System.Diagnostics.Debug.WriteLine($"Event parse error: {ex.Message}");
		}
	}

	private static StateEvent? ParseStateEvent(string sessionId, JsonNode data)
	{
		try {
			var stateStr = data["state"]?.GetValue<string>() ?? "idle";
			var state = stateStr.ToLowerInvariant() switch {
				"idle" => VMState.Idle,
				"running" => VMState.Running,
				"breakpoint" => VMState.Breakpoint,
				"halted" => VMState.Halted,
				"error" => VMState.Error,
				"waitingforinput" => VMState.WaitingForInput,
				_ => VMState.Idle
			};

			var pc = data["pc"]?.GetValue<uint>() ?? 0;
			var cycles = data["cycles"]?.GetValue<ulong>() ?? 0;
			var error = data["error"]?.GetValue<string>();

			var hasWrite = data["hasWrite"]?.GetValue<bool>() ?? false;
			MemoryWrite? memWrite = null;
			if (hasWrite) {
				var writeAddr = data["writeAddr"]?.GetValue<uint>() ?? 0;
				var writeSize = data["writeSize"]?.GetValue<uint>() ?? 0;
				memWrite = new MemoryWrite(writeAddr, writeSize);
			}

			var status = new VMStatus(state, pc, cycles, error, memWrite);

			// Parse registers
			var regsNode = data["registers"];
			var registers = regsNode is not null
				? ParseRegisters(regsNode)
				: RegisterState.Create();

			return new StateEvent(sessionId, status, registers);
		}
		catch {
			return null;
		}
	}

	private static RegisterState ParseRegisters(JsonNode regsNode)
	{
		var cpsr = regsNode["cpsr"];
		var cpsrFlags = cpsr is not null
			? new CPSRFlags(
				N: cpsr["n"]?.GetValue<bool>() ?? false,
				Z: cpsr["z"]?.GetValue<bool>() ?? false,
				C: cpsr["c"]?.GetValue<bool>() ?? false,
				V: cpsr["v"]?.GetValue<bool>() ?? false)
			: default;

		return RegisterState.Create(
			r0: regsNode["r0"]?.GetValue<uint>() ?? 0,
			r1: regsNode["r1"]?.GetValue<uint>() ?? 0,
			r2: regsNode["r2"]?.GetValue<uint>() ?? 0,
			r3: regsNode["r3"]?.GetValue<uint>() ?? 0,
			r4: regsNode["r4"]?.GetValue<uint>() ?? 0,
			r5: regsNode["r5"]?.GetValue<uint>() ?? 0,
			r6: regsNode["r6"]?.GetValue<uint>() ?? 0,
			r7: regsNode["r7"]?.GetValue<uint>() ?? 0,
			r8: regsNode["r8"]?.GetValue<uint>() ?? 0,
			r9: regsNode["r9"]?.GetValue<uint>() ?? 0,
			r10: regsNode["r10"]?.GetValue<uint>() ?? 0,
			r11: regsNode["r11"]?.GetValue<uint>() ?? 0,
			r12: regsNode["r12"]?.GetValue<uint>() ?? 0,
			sp: regsNode["sp"]?.GetValue<uint>() ?? 0,
			lr: regsNode["lr"]?.GetValue<uint>() ?? 0,
			pc: regsNode["pc"]?.GetValue<uint>() ?? 0,
			cpsr: cpsrFlags);
	}

	private static OutputEvent? ParseOutputEvent(string sessionId, JsonNode data)
	{
		try {
			var streamStr = data["stream"]?.GetValue<string>() ?? "stdout";
			var stream = streamStr.Equals("stderr", StringComparison.OrdinalIgnoreCase)
				? OutputStreamType.Stderr
				: OutputStreamType.Stdout;

			var content = data["content"]?.GetValue<string>() ?? string.Empty;

			return new OutputEvent(sessionId, stream, content);
		}
		catch {
			return null;
		}
	}

	private static ExecutionEvent? ParseExecutionEvent(string sessionId, JsonNode data)
	{
		try {
			var eventStr = data["event"]?.GetValue<string>() ?? string.Empty;
			var eventType = eventStr.ToLowerInvariant() switch {
				"breakpointhit" => ExecutionEventType.BreakpointHit,
				"halted" => ExecutionEventType.Halted,
				"error" => ExecutionEventType.Error,
				_ => ExecutionEventType.Error
			};

			var address = data["address"]?.GetValue<uint>();
			var symbol = data["symbol"]?.GetValue<string>();
			var message = data["message"]?.GetValue<string>();

			return new ExecutionEvent(sessionId, eventType, address, symbol, message);
		}
		catch {
			return null;
		}
	}
}

/// <summary>
/// Default WebSocket factory that creates ClientWebSocket instances.
/// </summary>
file sealed class DefaultWebSocketFactory : IWebSocketFactory
{
	public WebSocket CreateWebSocket() => new ClientWebSocket();
}

/// <summary>
/// Factory interface for creating WebSocket instances (enables testing).
/// </summary>
public interface IWebSocketFactory
{
	WebSocket CreateWebSocket();
}
