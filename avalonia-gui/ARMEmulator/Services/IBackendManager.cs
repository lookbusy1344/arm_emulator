namespace ARMEmulator.Services;

/// <summary>
/// Manages the lifecycle of the ARM Emulator backend process.
/// Handles platform-specific binary discovery, process start/stop, and health monitoring.
/// </summary>
public interface IBackendManager : IDisposable
{
	/// <summary>
	/// Current status of the backend process.
	/// </summary>
	BackendStatus Status { get; }

	/// <summary>
	/// Observable stream of status changes.
	/// </summary>
	IObservable<BackendStatus> StatusChanged { get; }

	/// <summary>
	/// Base URL of the backend HTTP API (e.g., "http://localhost:8080").
	/// </summary>
	string BaseUrl { get; }

	/// <summary>
	/// Starts the backend process if not already running.
	/// Automatically discovers the binary location based on platform.
	/// </summary>
	/// <param name="ct">Cancellation token</param>
	/// <exception cref="BackendStartException">Failed to start backend</exception>
	Task StartAsync(CancellationToken ct = default);

	/// <summary>
	/// Stops the backend process gracefully.
	/// </summary>
	Task StopAsync();

	/// <summary>
	/// Performs a health check against the backend API.
	/// </summary>
	/// <param name="ct">Cancellation token</param>
	/// <returns>True if backend is healthy and responding</returns>
	Task<bool> HealthCheckAsync(CancellationToken ct = default);
}

/// <summary>
/// Status of the backend process.
/// </summary>
public enum BackendStatus
{
	/// <summary>Status is unknown or not yet initialized.</summary>
	Unknown,

	/// <summary>Backend is starting up.</summary>
	Starting,

	/// <summary>Backend is running and healthy.</summary>
	Running,

	/// <summary>Backend is stopped.</summary>
	Stopped,

	/// <summary>Backend encountered an error.</summary>
	Error
}
