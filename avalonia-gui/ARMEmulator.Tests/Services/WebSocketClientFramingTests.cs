using System.Net.WebSockets;
using System.Text;
using ARMEmulator.Models;
using ARMEmulator.Services;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

/// <summary>
/// The client against a scripted socket: frame reassembly, connection failure and shutdown.
/// </summary>
public sealed class WebSocketClientFramingTests
{
	private const string Url = "ws://localhost:8080/api/v1/ws";
	private const int ReceiveBufferBytes = 8192;
	private static readonly TimeSpan CloseTimeout = TimeSpan.FromMilliseconds(100);
	private static readonly TimeSpan Guard = TimeSpan.FromSeconds(10);

	private static string OutputMessage(string content) =>
		$$$"""{"type":"output","sessionId":"s1","data":{"stream":"stdout","content":"{{{content}}}"}}""";

	private static async Task<IReadOnlyList<EmulatorEvent>> ReceiveAsync(int expectedEvents, params Frame[] frames)
	{
		var received = new List<EmulatorEvent>();
		var allReceived = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
		using var socket = new ScriptedWebSocket(frames);
		using var client = new WebSocketClient(Url, new SingleSocketFactory(socket));
		using var subscription = client.Events.Subscribe(evt => {
			lock (received) {
				received.Add(evt);
				if (received.Count == expectedEvents) {
					allReceived.SetResult();
				}
			}
		});

		await client.ConnectAsync("s1", TestContext.Current.CancellationToken);
		await allReceived.Task.WaitAsync(Guard, TestContext.Current.CancellationToken);

		lock (received) {
			return [.. received];
		}
	}

	[Fact]
	public async Task MessageSplitAcrossFrames_IsDeliveredOnce()
	{
		var bytes = Encoding.UTF8.GetBytes(OutputMessage("Hello"));
		var third = bytes.Length / 3;

		var events = await ReceiveAsync(
			expectedEvents: 1,
			new Frame(bytes[..third], EndOfMessage: false),
			new Frame(bytes[third..(2 * third)], EndOfMessage: false),
			new Frame(bytes[(2 * third)..], EndOfMessage: true));

		events.Should().Equal(new OutputEvent("s1", OutputStreamType.Stdout, "Hello"));
	}

	[Fact]
	public async Task MultiByteCharacterSplitAtFrameBoundary_DecodesIntact()
	{
		const string content = "café ✓ \U0001F600";
		var bytes = Encoding.UTF8.GetBytes(OutputMessage(content));
		var insideEAcute = Array.IndexOf(bytes, (byte)0xC3) + 1;

		var events = await ReceiveAsync(
			expectedEvents: 1,
			new Frame(bytes[..insideEAcute], EndOfMessage: false),
			new Frame(bytes[insideEAcute..], EndOfMessage: true));

		events.Should().Equal(new OutputEvent("s1", OutputStreamType.Stdout, content));
	}

	[Fact]
	public async Task MessageLargerThanTheReceiveBuffer_IsDeliveredIntact()
	{
		var content = new string('x', (3 * ReceiveBufferBytes) + 100);

		var events = await ReceiveAsync(
			expectedEvents: 1,
			new Frame(Encoding.UTF8.GetBytes(OutputMessage(content)), EndOfMessage: true));

		events.Should().Equal(new OutputEvent("s1", OutputStreamType.Stdout, content));
	}

	[Fact]
	public async Task ConsecutiveMessages_AreDeliveredSeparatelyInOrder()
	{
		var events = await ReceiveAsync(
			expectedEvents: 2,
			new Frame(Encoding.UTF8.GetBytes(OutputMessage("one")), EndOfMessage: true),
			new Frame(Encoding.UTF8.GetBytes(OutputMessage("two")), EndOfMessage: true));

		events.Should().Equal(
			new OutputEvent("s1", OutputStreamType.Stdout, "one"),
			new OutputEvent("s1", OutputStreamType.Stdout, "two"));
	}

	[Fact]
	public async Task Dispose_WhenServerNeverAnswersTheCloseHandshake_ReturnsAndAbortsTheSocket()
	{
		using var socket = new ScriptedWebSocket([], hangOnClose: true);
#pragma warning disable CA2000 // Disposed on a worker thread below
		var client = new WebSocketClient(Url, new SingleSocketFactory(socket), CloseTimeout);
#pragma warning restore CA2000
		await client.ConnectAsync("s1", TestContext.Current.CancellationToken);

		await Task.Run(() => DisposeOnBlockedUiThread(client), TestContext.Current.CancellationToken)
			.WaitAsync(Guard, TestContext.Current.CancellationToken);

		socket.WasAborted.Should().BeTrue();
	}

	[Fact]
	public async Task Dispose_OnAThreadWhoseSynchronizationContextIsNotPumping_DoesNotDeadlock()
	{
		using var socket = new ScriptedWebSocket([]);
#pragma warning disable CA2000 // Disposed on a worker thread below
		var client = new WebSocketClient(Url, new SingleSocketFactory(socket), CloseTimeout);
#pragma warning restore CA2000
		await client.ConnectAsync("s1", TestContext.Current.CancellationToken);

		await Task.Run(() => DisposeOnBlockedUiThread(client), TestContext.Current.CancellationToken)
			.WaitAsync(Guard, TestContext.Current.CancellationToken);
	}

	[Fact]
	public async Task ReceiveError_ReportsDisconnectedWithoutFaultingTheEventStream()
	{
		using var socket = new ScriptedWebSocket([], failWhenDrained: true);
		using var client = new WebSocketClient(Url, new SingleSocketFactory(socket), CloseTimeout);
		Exception? streamError = null;
		using var events = client.Events.Subscribe(_ => { }, ex => streamError = ex);
		var disconnected = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
		using var state = client.ConnectionState.Subscribe(connected => {
			if (!connected && socket.ReceiveCalls > 0) {
				disconnected.TrySetResult();
			}
		});

		await client.ConnectAsync("s1", TestContext.Current.CancellationToken);
		await disconnected.Task.WaitAsync(Guard, TestContext.Current.CancellationToken);

		streamError.Should().BeNull();
	}

	[Fact]
	public async Task ConnectAfterReceiveError_DisposesTheFailedSocketAndDeliversEventsToExistingSubscribers()
	{
		using var failing = new ScriptedWebSocket([], failWhenDrained: true);
		using var healthy = new ScriptedWebSocket([new Frame(Encoding.UTF8.GetBytes(OutputMessage("again")), EndOfMessage: true)]);
		using var client = new WebSocketClient(Url, new SequenceSocketFactory(failing, healthy), CloseTimeout);
		var received = new TaskCompletionSource<EmulatorEvent>(TaskCreationOptions.RunContinuationsAsynchronously);
		using var events = client.Events.Subscribe(evt => received.TrySetResult(evt), ex => received.TrySetException(ex));
		var failed = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
		using var state = client.ConnectionState.Subscribe(connected => {
			if (!connected && failing.ReceiveCalls > 0) {
				failed.TrySetResult();
			}
		});

		await client.ConnectAsync("s1", TestContext.Current.CancellationToken);
		await failed.Task.WaitAsync(Guard, TestContext.Current.CancellationToken);
		await client.ConnectAsync("s1", TestContext.Current.CancellationToken);

		var evt = await received.Task.WaitAsync(Guard, TestContext.Current.CancellationToken);
		evt.Should().Be(new OutputEvent("s1", OutputStreamType.Stdout, "again"));
		failing.WasDisposed.Should().BeTrue();
	}

	[Fact]
	public async Task ConnectAsync_WhenCancelled_ThrowsOperationCanceledAndReportsDisconnected()
	{
		using var socket = new ScriptedWebSocket([]);
		using var client = new WebSocketClient(Url, new SingleSocketFactory(socket), CloseTimeout);
		var states = new List<bool>();
		using var state = client.ConnectionState.Subscribe(states.Add);
		using var cancelled = new CancellationTokenSource();
		await cancelled.CancelAsync();

		var act = async () => await client.ConnectAsync("s1", cancelled.Token);

		await act.Should().ThrowExactlyAsync<OperationCanceledException>();
		states[^1].Should().BeFalse();
	}

	/// <summary>Disposes under a context that never runs posted work, as the UI thread does while it blocks in Dispose.</summary>
	private static void DisposeOnBlockedUiThread(WebSocketClient client)
	{
		SynchronizationContext.SetSynchronizationContext(new NonPumpingContext());
		client.Dispose();
	}

	private sealed class NonPumpingContext : SynchronizationContext
	{
		public override void Post(SendOrPostCallback d, object? state)
		{
			// Posted continuations never run: the owning thread is blocked
		}
	}

	private sealed record Frame(byte[] Bytes, bool EndOfMessage);

	private sealed class SingleSocketFactory(WebSocket socket) : IWebSocketFactory
	{
		public WebSocket CreateWebSocket() => socket;
	}

	private sealed class SequenceSocketFactory(params WebSocket[] sockets) : IWebSocketFactory
	{
		private readonly Queue<WebSocket> remaining = new(sockets);

		public WebSocket CreateWebSocket() => remaining.Dequeue();
	}

	/// <summary>
	/// Hands out scripted frames, splitting any frame larger than the caller's buffer. When the frames run out it waits for
	/// cancellation or, with <paramref name="failWhenDrained"/>, aborts and throws as a dropped connection does.
	/// </summary>
	private sealed class ScriptedWebSocket(IEnumerable<Frame> frames, bool hangOnClose = false, bool failWhenDrained = false) : WebSocket
	{
		private readonly Queue<Frame> pending = new(frames);
		private WebSocketState state = WebSocketState.Open;
		private int offsetInFrame;

		public override WebSocketCloseStatus? CloseStatus => null;

		public override string? CloseStatusDescription => null;

		public override WebSocketState State => state;

		public override string? SubProtocol => null;

		public bool WasAborted { get; private set; }

		public bool WasDisposed { get; private set; }

		public int ReceiveCalls => Volatile.Read(ref receiveCalls);

		private int receiveCalls;

		public override void Abort()
		{
			WasAborted = true;
			state = WebSocketState.Aborted;
		}

		public override async Task CloseAsync(WebSocketCloseStatus closeStatus, string? statusDescription, CancellationToken cancellationToken)
		{
			if (hangOnClose) {
				await Task.Delay(Timeout.Infinite, cancellationToken).ConfigureAwait(false);
			}

			state = WebSocketState.Closed;
		}

		public override Task CloseOutputAsync(WebSocketCloseStatus closeStatus, string? statusDescription, CancellationToken cancellationToken) =>
			CloseAsync(closeStatus, statusDescription, cancellationToken);

		public override void Dispose()
		{
			WasDisposed = true;
			state = WebSocketState.Closed;
		}

		public override async Task<WebSocketReceiveResult> ReceiveAsync(ArraySegment<byte> buffer, CancellationToken cancellationToken)
		{
			_ = Interlocked.Increment(ref receiveCalls);
			if (!pending.TryPeek(out var frame)) {
				if (failWhenDrained) {
					state = WebSocketState.Aborted;
					throw new WebSocketException(WebSocketError.ConnectionClosedPrematurely);
				}

				await Task.Delay(Timeout.Infinite, cancellationToken);
				throw new InvalidOperationException("unreachable");
			}

			var count = Math.Min(buffer.Count, frame.Bytes.Length - offsetInFrame);
			frame.Bytes.AsSpan(offsetInFrame, count).CopyTo(buffer.AsSpan());
			offsetInFrame += count;
			var frameDone = offsetInFrame == frame.Bytes.Length;
			if (frameDone) {
				_ = pending.Dequeue();
				offsetInFrame = 0;
			}

			return new WebSocketReceiveResult(count, WebSocketMessageType.Text, frameDone && frame.EndOfMessage);
		}

		public override Task SendAsync(ArraySegment<byte> buffer, WebSocketMessageType messageType, bool endOfMessage, CancellationToken cancellationToken)
		{
			cancellationToken.ThrowIfCancellationRequested();
			return Task.CompletedTask;
		}
	}
}
