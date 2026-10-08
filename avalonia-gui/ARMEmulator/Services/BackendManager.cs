using System.Diagnostics;
using System.Diagnostics.CodeAnalysis;
using System.Reactive.Subjects;
using System.Runtime.InteropServices;

namespace ARMEmulator.Services;

/// <summary>
/// Manages the ARM Emulator backend process lifecycle.
/// Discovers platform-specific binary locations and manages process start/stop.
/// </summary>
public sealed class BackendManager : IBackendManager
{
	private const string DefaultBaseUrl = "http://localhost:8080";
	private const int StartupPollAttempts = 30;
	private static readonly TimeSpan StartupPollInterval = TimeSpan.FromMilliseconds(100);
	private static readonly TimeSpan HealthCheckTimeout = TimeSpan.FromSeconds(1);

	private readonly BehaviorSubject<BackendStatus> statusSubject = new(BackendStatus.Stopped);
	private readonly string baseUrl;
	private readonly Uri baseUri;
	private readonly HttpClient http;
	private readonly Func<string?> findBinary;
	private Process? process;

	/// <summary>Creates a manager for the backend served at <paramref name="baseUrl"/>.</summary>
	/// <param name="baseUrl">Backend HTTP base URL. A spawned backend listens on this URL's port.</param>
	/// <param name="healthCheckHandler">HTTP handler for health checks; the default handler when null.</param>
	/// <param name="findBinary">Locates the backend executable; platform discovery when null.</param>
	public BackendManager(
		string baseUrl = DefaultBaseUrl,
		HttpMessageHandler? healthCheckHandler = null,
		Func<string?>? findBinary = null)
	{
		if (!Uri.TryCreate(baseUrl, UriKind.Absolute, out var uri) || uri.Scheme is not ("http" or "https")) {
			throw new ArgumentException($"'{baseUrl}' is not an absolute HTTP URL.", nameof(baseUrl));
		}

		this.baseUrl = baseUrl;
		baseUri = uri;
		http = healthCheckHandler is null ? new HttpClient() : new HttpClient(healthCheckHandler, disposeHandler: false);
		this.findBinary = findBinary ?? FindBackendBinary;
	}

	public BackendStatus Status => statusSubject.Value;

	public IObservable<BackendStatus> StatusChanged => statusSubject;

	public string BaseUrl => baseUrl;

	/// <summary>
	/// Command-line arguments that start the backend as an API server on <paramref name="baseUrl"/>'s port.
	/// </summary>
	internal static ImmutableArray<string> ServerArguments(Uri baseUrl) =>
		["-api-server", "-port", baseUrl.Port.ToString(System.Globalization.CultureInfo.InvariantCulture)];

	public async Task StartAsync(CancellationToken ct = default)
	{
		if (process is not null && !process.HasExited) {
			return; // Already running
		}

		statusSubject.OnNext(BackendStatus.Starting);

		try {
			// Reuse a backend that is already serving this URL, e.g. one started from a terminal
			if (await HealthCheckAsync(ct)) {
				statusSubject.OnNext(BackendStatus.Running);
				return;
			}

			process = SpawnBackend(findBinary() ?? throw new BackendStartException("Backend binary not found"));

			for (var attempt = 0; attempt < StartupPollAttempts; ++attempt) {
				if (await HealthCheckAsync(ct)) {
					statusSubject.OnNext(BackendStatus.Running);
					return;
				}

				await Task.Delay(StartupPollInterval, ct);
			}

			await StopAsync().ConfigureAwait(false);
			throw new BackendStartException("Backend started but health check failed");
		}
		catch (OperationCanceledException) {
			// A cancelled start leaves no process behind
			await StopAsync().ConfigureAwait(false);
			throw;
		}
		catch (BackendStartException) {
			statusSubject.OnNext(BackendStatus.Error);
			throw;
		}
		catch (Exception ex) {
			statusSubject.OnNext(BackendStatus.Error);
			throw new BackendStartException("Failed to start backend", ex);
		}
	}

	private Process SpawnBackend(string binaryPath)
	{
		var startInfo = new ProcessStartInfo {
			FileName = binaryPath,
			UseShellExecute = false,
			CreateNoWindow = true
		};
		foreach (var argument in ServerArguments(baseUri)) {
			startInfo.ArgumentList.Add(argument);
		}

		var spawned = new Process { StartInfo = startInfo };
		try {
			return spawned.Start() ? spawned : throw new BackendStartException("Failed to start backend process");
		}
		catch {
			spawned.Dispose();
			throw;
		}
	}

	public async Task StopAsync()
	{
		var running = process;
		if (running is null) {
			statusSubject.OnNext(BackendStatus.Stopped);
			return;
		}

		try {
			running.Kill(entireProcessTree: true);
		}
		catch (InvalidOperationException) {
			// The process has already exited
		}

		await running.WaitForExitAsync().ConfigureAwait(false);
		running.Dispose();
		process = null;
		statusSubject.OnNext(BackendStatus.Stopped);
	}

	public async Task<bool> HealthCheckAsync(CancellationToken ct = default)
	{
		try {
			using var timeout = CancellationTokenSource.CreateLinkedTokenSource(ct);
			timeout.CancelAfter(HealthCheckTimeout);

			using var response = await http.GetAsync($"{baseUrl}/health", timeout.Token);
			return response.IsSuccessStatusCode;
		}
		catch (Exception ex) when (ex is HttpRequestException or OperationCanceledException) {
			ct.ThrowIfCancellationRequested();
			return false;
		}
	}

	[SuppressMessage("Usage", "VSTHRD002:Avoid problematic synchronous waits",
		Justification = "Dispose must be synchronous; ConfigureAwait(false) prevents deadlock")]
	public void Dispose()
	{
		StopAsync().ConfigureAwait(false).GetAwaiter().GetResult();
		http.Dispose();
		statusSubject.Dispose();
	}

	private static string? FindBackendBinary() =>
		BackendLocator.Current is BackendPlatform platform
			? BackendLocator.Find(platform, AppContext.BaseDirectory, File.Exists)
			: null;
}
