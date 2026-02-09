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
	private readonly BehaviorSubject<BackendStatus> statusSubject = new(BackendStatus.Stopped);
	private readonly string baseUrl;
	private readonly HttpClient http = new();
	private Process? process;

	public BackendManager(string baseUrl = "http://localhost:8080")
	{
		this.baseUrl = baseUrl;
	}

	public BackendStatus Status => statusSubject.Value;

	public IObservable<BackendStatus> StatusChanged => statusSubject;

	public string BaseUrl => baseUrl;

	public async Task StartAsync(CancellationToken ct = default)
	{
		if (process is not null && !process.HasExited) {
			return; // Already running
		}

		statusSubject.OnNext(BackendStatus.Starting);

		try {
			var binaryPath = FindBackendBinary();
			if (binaryPath is null) {
				statusSubject.OnNext(BackendStatus.Error);
				throw new BackendStartException("Backend binary not found");
			}

			process = new Process {
				StartInfo = new ProcessStartInfo {
					FileName = binaryPath,
					UseShellExecute = false,
					RedirectStandardOutput = true,
					RedirectStandardError = true,
					CreateNoWindow = true
				}
			};

			if (!process.Start()) {
				statusSubject.OnNext(BackendStatus.Error);
				throw new BackendStartException("Failed to start backend process");
			}

			// Wait for backend to be ready
			for (int i = 0; i < 30; i++) // 3 second timeout
			{
				if (await HealthCheckAsync(ct)) {
					statusSubject.OnNext(BackendStatus.Running);
					return;
				}

				await Task.Delay(100, ct);
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
