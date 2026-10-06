namespace ARMEmulator.Services;

/// <summary>
/// Derives backend endpoint addresses from the HTTP base URL.
/// </summary>
public static class BackendEndpoints
{
	private const string WebSocketPath = "/api/v1/ws";

	/// <summary>
	/// Returns the WebSocket event endpoint for a backend served at <paramref name="baseUri"/>.
	/// </summary>
	/// <exception cref="ArgumentException"><paramref name="baseUri"/> is not an http or https URI.</exception>
	public static Uri WebSocketUri(Uri baseUri)
	{
		ArgumentNullException.ThrowIfNull(baseUri);

		var scheme = baseUri.Scheme switch {
			"http" => "ws",
			"https" => "wss",
			_ => throw new ArgumentException($"Backend URL must use http or https, not '{baseUri.Scheme}'.", nameof(baseUri))
		};

		return new UriBuilder(baseUri) { Scheme = scheme, Path = WebSocketPath }.Uri;
	}
}
