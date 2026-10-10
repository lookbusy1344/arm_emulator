using System.Net.Http.Headers;
using System.Net.WebSockets;

namespace ARMEmulator.Services;

/// <summary>
/// Locates and reads the per-launch token the Go backend writes for its port.
/// Every request except <c>/health</c> must carry it as a bearer token.
/// </summary>
public static class BackendToken
{
	internal const string Scheme = "Bearer";
	private const string AppDirectory = "arm-emu";
	private const string FilePrefix = "api-token-";

	/// <summary>
	/// Returns the per-user configuration root the backend uses: %APPDATA% on Windows, ~/.config elsewhere.
	/// </summary>
	public static string GetDefaultConfigDirectory() =>
		OperatingSystem.IsWindows()
			? Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData)
			: Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".config");

	/// <summary>
	/// Returns the token file for a backend listening on <paramref name="port"/>.
	/// </summary>
	public static string FilePath(string configDirectory, int port)
	{
		ArgumentException.ThrowIfNullOrEmpty(configDirectory);
		return Path.Combine(configDirectory, AppDirectory, FilePrefix + port.ToString(System.Globalization.CultureInfo.InvariantCulture));
	}

	/// <summary>
	/// Reads the token at <paramref name="path"/>. Returns null when no backend has written one.
	/// </summary>
	public static string? Read(string path)
	{
		ArgumentException.ThrowIfNullOrEmpty(path);
		string text;
		try {
			text = File.ReadAllText(path);
		}
		catch (Exception ex) when (ex is FileNotFoundException or DirectoryNotFoundException) {
			return null;
		}

		var token = text.Trim();
		return token.Length == 0 ? null : token;
	}
}

/// <summary>
/// Adds the backend token to each request. The token is read per request, so a restarted backend's new token is used.
/// </summary>
public sealed class BackendTokenHandler(Func<string?> readToken) : DelegatingHandler
{
	protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
	{
		ArgumentNullException.ThrowIfNull(request);
		var token = readToken();
		if (token is not null) {
			request.Headers.Authorization = new AuthenticationHeaderValue(BackendToken.Scheme, token);
		}

		return base.SendAsync(request, cancellationToken);
	}
}

/// <summary>
/// Creates WebSockets whose upgrade request carries the backend token.
/// </summary>
public sealed class AuthorizedWebSocketFactory(Func<string?> readToken) : IWebSocketFactory
{
	public WebSocket CreateWebSocket()
	{
		var socket = new ClientWebSocket();
		var token = readToken();
		if (token is not null) {
			socket.Options.SetRequestHeader("Authorization", $"{BackendToken.Scheme} {token}");
		}

		return socket;
	}
}
