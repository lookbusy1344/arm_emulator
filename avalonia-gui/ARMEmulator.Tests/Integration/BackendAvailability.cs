namespace ARMEmulator.Tests.Integration;

/// <summary>Probes the backend once so integration tests skip when it is not running.</summary>
internal static class BackendAvailability
{
	private const string UrlVariable = "ARM_EMULATOR_URL";
	private const string DefaultBaseUrl = "http://localhost:8080";
	private static readonly TimeSpan ProbeTimeout = TimeSpan.FromSeconds(1);

	/// <summary>The backend under test: <c>ARM_EMULATOR_URL</c>, or localhost:8080.</summary>
	public static Uri BaseUri { get; } = new(Environment.GetEnvironmentVariable(UrlVariable) is { Length: > 0 } url ? url : DefaultBaseUrl);

	private static readonly Lazy<bool> Probe = new(() => {
		try {
			using var http = new HttpClient { Timeout = ProbeTimeout };
			using var response = http.GetAsync(new Uri(BaseUri.AbsoluteUri.TrimEnd('/') + "/health")).GetAwaiter().GetResult();
			return response.IsSuccessStatusCode;
		}
		catch (HttpRequestException) {
			return false;
		}
		catch (TaskCanceledException) {
			return false;
		}
	});

	public const string SkipReason = "Requires running backend at localhost:8080 (./arm-emulator -api-server), or set ARM_EMULATOR_URL";

	public static bool IsRunning => Probe.Value;
}
