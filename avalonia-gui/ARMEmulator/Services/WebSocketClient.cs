using System.Diagnostics.CodeAnalysis;
using System.Net.WebSockets;
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
	private readonly string wsUrl;
	private readonly IWebSocketFactory factory;
	private readonly Subject<EmulatorEvent> eventsSubject = new();
	private readonly BehaviorSubject<bool> connectionStateSubject = new(false);
	private readonly CancellationTokenSource disposeCts = new();

	private WebSocket? ws;
	private Task? receiveTask;
	private string currentSessionId = string.Empty;

	/// <summary>
	/// Creates a new WebSocket client.
	/// </summary>
	/// <param name="wsUrl">WebSocket URL (e.g., "ws://localhost:8080/ws")</param>
	/// <param name="factory">Factory for creating WebSocket instances (injectable for testing)</param>
	public WebSocketClient(string wsUrl, IWebSocketFactory? factory = null)
	{
		this.wsUrl = wsUrl;
		this.factory = factory ?? new DefaultWebSocketFactory();
	}

	public IObservable<EmulatorEvent> Events => eventsSubject.AsObservable();

	public bool IsConnected => ws?.State == WebSocketState.Open;

	public IObservable<bool> ConnectionState => connectionStateSubject.AsObservable();

	public async Task ConnectAsync(string sessionId, CancellationToken ct = default)
	{
		if (IsConnected) {
			await DisconnectAsync();
		}

		currentSessionId = sessionId;

		try {
			ws = factory.CreateWebSocket();

			if (ws is ClientWebSocket clientWs) {
				await clientWs.ConnectAsync(new Uri(wsUrl), ct);
			}

			connectionStateSubject.OnNext(true);

			// Send subscription message
			await SendSubscriptionAsync(sessionId, ct);

			// Start receive loop
			receiveTask = Task.Run(() => ReceiveLoopAsync(disposeCts.Token), disposeCts.Token);
		}
		catch (Exception ex) {
			connectionStateSubject.OnNext(false);
			throw new WebSocketConnectionException($"Failed to connect to {wsUrl}", ex);
		}
	}

	public async Task DisconnectAsync()
	{
		if (ws is null) {
			return;
		}

		try {
			if (ws.State == WebSocketState.Open) {
				await ws.CloseAsync(
					WebSocketCloseStatus.NormalClosure,
					"Client disconnecting",
					CancellationToken.None);
			}

			ws.Dispose();
			ws = null;

			connectionStateSubject.OnNext(false);

			if (receiveTask is not null) {
				await receiveTask.ConfigureAwait(false);
				receiveTask = null;
			}
		}
		catch {
			// Ignore disconnect errors
		}
	}

	[SuppressMessage("Usage", "VSTHRD002:Avoid problematic synchronous waits", Justification = "Dispose must be synchronous; ConfigureAwait(false) prevents deadlock")]
	public void Dispose()
	{
		disposeCts.Cancel();
		try {
			DisconnectAsync().ConfigureAwait(false).GetAwaiter().GetResult();
		}
		catch {
			// Ignore dispose errors - may occur if connection already closed
		}
		disposeCts.Dispose();
		eventsSubject.Dispose();
		connectionStateSubject.Dispose();
	}

	private async Task SendSubscriptionAsync(string sessionId, CancellationToken ct)
	{
		// Manual JSON construction to avoid reflection
		var json = $$"""{"type":"subscribe","sessionId":"{{sessionId}}","events":[]}""";
		var bytes = Encoding.UTF8.GetBytes(json);
		await ws!.SendAsync(
			new ArraySegment<byte>(bytes),
			WebSocketMessageType.Text,
			endOfMessage: true,
			ct);
	}

	private async Task ReceiveLoopAsync(CancellationToken ct)
	{
		var buffer = new byte[8192];

		try {
			while (!ct.IsCancellationRequested && ws?.State == WebSocketState.Open) {
				var result = await ws.ReceiveAsync(new ArraySegment<byte>(buffer), ct);

				if (result.MessageType == WebSocketMessageType.Close) {
					connectionStateSubject.OnNext(false);
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
			eventsSubject.OnError(new WebSocketConnectionException("WebSocket error", ex));
			connectionStateSubject.OnNext(false);
		}
	}

	private void ProcessMessage(string message)
	{
		if (ParseMessage(message) is EmulatorEvent evt) {
			eventsSubject.OnNext(evt);
		}
	}

	/// <summary>
	/// Parses one broadcast message (api/broadcaster.go). Returns null for messages the GUI does not use
	/// or cannot interpret, so a malformed payload never terminates the event stream.
	/// </summary>
	internal static EmulatorEvent? ParseMessage(string message)
	{
		try {
			var json = JsonNode.Parse(message);
			var data = json?["data"];
			if (json is null || data is null) {
				return null;
			}

			var sessionId = json["sessionId"]?.GetValue<string>() ?? string.Empty;
			return json["type"]?.GetValue<string>() switch {
				"state" => ParseStateEvent(sessionId, data),
				"output" => ParseOutputEvent(sessionId, data),
				"event" => ParseExecutionEvent(sessionId, data),
				_ => null
			};
		}
		catch (Exception ex) when (ex is JsonException or FormatException or InvalidOperationException) {
			System.Diagnostics.Debug.WriteLine($"Ignoring WebSocket message: {ex.Message}");
			return null;
		}
	}

	// Full updates (handlers.go broadcastStateChange) carry pc and registers;
	// status-only updates (session_manager.go OnStateChange) carry just the status.
	private static StateEvent ParseStateEvent(string sessionId, JsonNode data)
	{
		var state = WireFormat.ParseState(data["status"]?.GetValue<string>() ?? throw new FormatException("State event has no status."));
		var pc = data["pc"]?.GetValue<uint>() ?? 0;
		var registersNode = data["registers"];
		var registers = registersNode is null
			? null
			: JsonSerializer.Deserialize(registersNode, ApiJsonContext.Default.RegistersResponse)?.ToRegisterState();

		return new StateEvent(sessionId, new VMStatus(state, pc, Cycles: 0), registers);
	}

	private static OutputEvent ParseOutputEvent(string sessionId, JsonNode data)
	{
		var stream = data["stream"]?.GetValue<string>() == "stderr" ? OutputStreamType.Stderr : OutputStreamType.Stdout;
		var content = data["content"]?.GetValue<string>() ?? string.Empty;

		return new OutputEvent(sessionId, stream, content);
	}

	private static ExecutionEvent ParseExecutionEvent(string sessionId, JsonNode data)
	{
		var eventType = WireFormat.ParseExecutionEvent(data["event"]?.GetValue<string>() ?? throw new FormatException("Execution event has no name."));

		return new ExecutionEvent(
			sessionId,
			eventType,
			data["address"]?.GetValue<uint>(),
			data["symbol"]?.GetValue<string>(),
			data["message"]?.GetValue<string>());
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
