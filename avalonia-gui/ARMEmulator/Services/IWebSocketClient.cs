namespace ARMEmulator.Services;

/// <summary>
/// WebSocket client for real-time event streaming from the ARM Emulator backend.
/// Automatically reconnects on disconnect with exponential backoff.
/// </summary>
public interface IWebSocketClient : IDisposable
{
	/// <summary>
	/// Observable stream of emulator events from the backend.
	/// Events are delivered on the main thread scheduler.
	/// </summary>
	IObservable<Models.EmulatorEvent> Events { get; }

	/// <summary>
	/// Current connection state.
	/// </summary>
	bool IsConnected { get; }

	/// <summary>
	/// Observable stream of connection state changes.
	/// </summary>
	IObservable<bool> ConnectionState { get; }

	/// <summary>
	/// Connects to the WebSocket endpoint and subscribes to events for the specified session.
	/// </summary>
	/// <param name="sessionId">Session ID to filter events (empty string = all sessions)</param>
	/// <param name="ct">Cancellation token</param>
	/// <exception cref="WebSocketConnectionException">Connection failed</exception>
	Task ConnectAsync(string sessionId, CancellationToken ct = default);

	/// <summary>
	/// Disconnects from the WebSocket endpoint and stops event streaming.
	/// </summary>
	Task DisconnectAsync();
}
