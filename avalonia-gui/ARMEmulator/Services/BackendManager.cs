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

	private readonly BehaviorSubject<BackendStatus> statusSubject = new(BackendStatus.Stopped);
	private readonly string baseUrl;
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
		this.baseUrl = baseUrl;
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

		// Reuse a backend that is already serving this URL, e.g. one started from a terminal
		if (await HealthCheckAsync(ct)) {
			statusSubject.OnNext(BackendStatus.Running);
			return;
		}

		try {
			var binaryPath = findBinary();
			if (binaryPath is null) {
				statusSubject.OnNext(BackendStatus.Error);
				throw new BackendStartException("Backend binary not found");
			}

			var startInfo = new ProcessStartInfo {
				FileName = binaryPath,
				UseShellExecute = false,
				CreateNoWindow = true
			};
			foreach (var argument in ServerArguments(new Uri(baseUrl))) {
				startInfo.ArgumentList.Add(argument);
			}

			process = new Process { StartInfo = startInfo };

			if (!process.Start()) {
				statusSubject.OnNext(BackendStatus.Error);
				throw new BackendStartException("Failed to start backend process");
			}

			// Wait for backend to be ready
			for (var attempt = 0; attempt < StartupPollAttempts; ++attempt) {
				if (await HealthCheckAsync(ct)) {
					statusSubject.OnNext(BackendStatus.Running);
					return;
				}

				await Task.Delay(StartupPollInterval, ct);
			}

			statusSubject.OnNext(BackendStatus.Error);
			throw new BackendStartException("Backend started but health check failed");
		}
		catch (Exception ex) when (ex is not BackendStartException) {
			statusSubject.OnNext(BackendStatus.Error);
			throw new BackendStartException("Failed to start backend", ex);
		}
	}

	public async Task StopAsync()
	{
		if (process is null || process.HasExited) {
			statusSubject.OnNext(BackendStatus.Stopped);
			return;
		}

		try {
			process.Kill(entireProcessTree: true);
			await process.WaitForExitAsync();
			process.Dispose();
			process = null;
			statusSubject.OnNext(BackendStatus.Stopped);
		}
		catch {
			// Ignore stop errors
			statusSubject.OnNext(BackendStatus.Stopped);
		}
	}

	public async Task<bool> HealthCheckAsync(CancellationToken ct = default)
	{
		try {
			using var cts = CancellationTokenSource.CreateLinkedTokenSource(ct);
			cts.CancelAfter(TimeSpan.FromSeconds(1));

			var response = await http.GetAsync($"{baseUrl}/health", cts.Token);
			return response.IsSuccessStatusCode;
		}
		catch {
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

	private static string? FindBackendBinary()
	{
		// Platform-specific binary discovery
		if (RuntimeInformation.IsOSPlatform(OSPlatform.Windows)) {
			return FindBinaryWindows();
		} else if (RuntimeInformation.IsOSPlatform(OSPlatform.OSX)) {
			return FindBinaryMacOS();
		} else if (RuntimeInformation.IsOSPlatform(OSPlatform.Linux)) {
			return FindBinaryLinux();
		}

		return null;
	}

	private static string? FindBinaryWindows()
	{
		// Check app directory
		var appDir = AppContext.BaseDirectory;
		var binaryPath = Path.Combine(appDir, "arm-emulator.exe");
		if (File.Exists(binaryPath)) {
			return binaryPath;
		}

		// Check parent directory (for development)
		var parentDir = Directory.GetParent(appDir)?.FullName;
		if (parentDir is not null) {
			binaryPath = Path.Combine(parentDir, "arm-emulator.exe");
			if (File.Exists(binaryPath)) {
				return binaryPath;
			}
		}

		return null;
	}

	private static string? FindBinaryMacOS()
	{
		// Check if running from .app bundle
		var appDir = AppContext.BaseDirectory;
		if (appDir.Contains(".app/Contents/")) {
			// Running from .app bundle - check Contents/Resources
			var bundleContents = appDir[..appDir.IndexOf(".app/Contents/", StringComparison.Ordinal)] + ".app/Contents";
			var resourcesPath = Path.Combine(bundleContents, "Resources", "arm-emulator");
			if (File.Exists(resourcesPath)) {
				return resourcesPath;
			}
		}

		// Check app directory
		var binaryPath = Path.Combine(appDir, "arm-emulator");
		if (File.Exists(binaryPath)) {
			return binaryPath;
		}

		// Check parent directory (for development)
		var parentDir = Directory.GetParent(appDir)?.FullName;
		if (parentDir is not null) {
			binaryPath = Path.Combine(parentDir, "arm-emulator");
			if (File.Exists(binaryPath)) {
				return binaryPath;
			}
		}

		return null;
	}

	private static string? FindBinaryLinux()
	{
		// Check app directory
		var appDir = AppContext.BaseDirectory;
		var binaryPath = Path.Combine(appDir, "arm-emulator");
		if (File.Exists(binaryPath)) {
			return binaryPath;
		}

		// Check /usr/local/bin
		binaryPath = "/usr/local/bin/arm-emulator";
		if (File.Exists(binaryPath)) {
			return binaryPath;
		}

		// Check /usr/share/arm-emulator
		binaryPath = "/usr/share/arm-emulator/arm-emulator";
		if (File.Exists(binaryPath)) {
			return binaryPath;
		}

		return null;
	}
}
