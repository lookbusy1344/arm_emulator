namespace ARMEmulator.Tests.Integration;

/// <summary>Probes the backend once so integration tests skip when it is not running.</summary>
internal static class BackendAvailability
{
	private const string HealthUrl = "http://localhost:8080/health";
	private static readonly TimeSpan ProbeTimeout = TimeSpan.FromSeconds(1);

	private static readonly Lazy<bool> Probe = new(() => {
		try {
			using var http = new HttpClient { Timeout = ProbeTimeout };
			using var response = http.GetAsync(HealthUrl).GetAwaiter().GetResult();
			return response.IsSuccessStatusCode;
		}
		catch (HttpRequestException) {
			return false;
		}
		catch (TaskCanceledException) {
			return false;
		}
	});

	public const string SkipReason = "Requires running backend at localhost:8080 (./arm-emulator -api-server)";

	public static bool IsRunning => Probe.Value;
}
