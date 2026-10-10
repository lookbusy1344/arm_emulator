using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text;
using ARMEmulator.Services;
using AwesomeAssertions;

namespace ARMEmulator.Tests.Services;

public sealed class BackendTokenTests : IDisposable
{
	private const int Port = 8080;
	private const string Token = "abc123";

	private readonly string configDirectory = Directory.CreateTempSubdirectory("arm-token-tests-").FullName;

	public void Dispose() => Directory.Delete(configDirectory, recursive: true);

	private string WriteToken(string text, int port = Port)
	{
		var path = BackendToken.FilePath(configDirectory, port);
		_ = Directory.CreateDirectory(Path.GetDirectoryName(path)!);
		File.WriteAllText(path, text);
		return path;
	}

	[Fact]
	public void FilePath_IsPerPortInArmEmuDirectory()
	{
		BackendToken.FilePath(configDirectory, Port).Should().Be(Path.Combine(configDirectory, "arm-emu", "api-token-8080"));
	}

	[Fact]
	public void GetDefaultConfigDirectory_MatchesBackendLayout()
	{
		var expected = OperatingSystem.IsWindows()
			? Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData)
			: Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".config");

		BackendToken.GetDefaultConfigDirectory().Should().Be(expected);
	}

	[Fact]
	public void Read_ReturnsTokenWithoutTrailingNewline()
	{
		BackendToken.Read(WriteToken(Token + "\n")).Should().Be(Token);
	}

	[Fact]
	public void Read_MissingFile_ReturnsNull()
	{
		BackendToken.Read(BackendToken.FilePath(configDirectory, Port)).Should().BeNull();
	}

	[Fact]
	public void Read_BlankFile_ReturnsNull()
	{
		BackendToken.Read(WriteToken(" \n")).Should().BeNull();
	}

	[Fact]
	public void Read_OtherPortsFile_IsNotUsed()
	{
		_ = WriteToken(Token, port: 9090);

		BackendToken.Read(BackendToken.FilePath(configDirectory, Port)).Should().BeNull();
	}
}

public sealed class BackendTokenHandlerTests : IDisposable
{
	private readonly TestHttpMessageHandler inner = new();
	private readonly BackendTokenHandler handler;
	private readonly HttpClient http;
	private string? token;

	public BackendTokenHandlerTests()
	{
		handler = new BackendTokenHandler(() => token) { InnerHandler = inner };
		http = new HttpClient(handler, disposeHandler: false) { BaseAddress = new Uri("http://localhost:8080") };
	}

	public void Dispose()
	{
		http.Dispose();
		handler.Dispose();
		inner.Dispose();
	}

	private static CancellationToken Ct => TestContext.Current.CancellationToken;

	[Fact]
	public async Task SendAsync_WithToken_AddsBearerHeader()
	{
		token = "abc123";

		using var _ = await http.GetAsync("/api/v1/session", Ct);

		inner.LastRequest!.Headers.Authorization.Should().Be(new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", "abc123"));
	}

	[Fact]
	public async Task SendAsync_WithoutToken_SendsNoAuthorization()
	{
		using var _ = await http.GetAsync("/api/v1/session", Ct);

		inner.LastRequest!.Headers.Authorization.Should().BeNull();
	}

	// A restarted backend writes a new token; the next request must use it.
	[Fact]
	public async Task SendAsync_ReadsTokenForEachRequest()
	{
		token = "first";
		using (var _ = await http.GetAsync("/api/v1/session", Ct)) { }
		token = "second";

		using var response = await http.GetAsync("/api/v1/session", Ct);

		inner.LastRequest!.Headers.Authorization!.Parameter.Should().Be("second");
	}
}

public sealed class AuthorizedWebSocketFactoryTests
{
	private const int TimeoutMs = 5000;

	// Closing without a response makes the HTTP stack retry the upgrade on a new connection, so answer it.
	private const string RefusalResponse = "HTTP/1.1 401 Unauthorized\r\nContent-Length: 0\r\n\r\n";

	/// <summary>Accepts one connection, refuses it, and returns the HTTP upgrade request the client sent.</summary>
	private static async Task<string> CaptureUpgradeRequestAsync(AuthorizedWebSocketFactory factory, CancellationToken ct)
	{
		using var listener = new TcpListener(IPAddress.Loopback, 0);
		listener.Start();
		var port = ((IPEndPoint)listener.LocalEndpoint).Port;

		using var socket = (ClientWebSocket)factory.CreateWebSocket();
		var connect = socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/api/v1/ws"), ct);

		using var server = await listener.AcceptTcpClientAsync(ct);
		var stream = server.GetStream();
		var request = new StringBuilder();
		var buffer = new byte[1024];
		while (!request.ToString().Contains("\r\n\r\n", StringComparison.Ordinal)) {
			var read = await stream.ReadAsync(buffer, ct);
			if (read == 0) {
				break;
			}

			_ = request.Append(Encoding.ASCII.GetString(buffer, 0, read));
		}

		await stream.WriteAsync(Encoding.ASCII.GetBytes(RefusalResponse), ct);
		await connect.ContinueWith(static _ => { }, ct, TaskContinuationOptions.None, TaskScheduler.Default);
		return request.ToString();
	}

	[Fact(Timeout = TimeoutMs)]
	public async Task CreateWebSocket_WithToken_SendsBearerHeaderOnUpgrade()
	{
		var request = await CaptureUpgradeRequestAsync(new AuthorizedWebSocketFactory(() => "abc123"), TestContext.Current.CancellationToken);

		request.Should().Contain("\r\nAuthorization: Bearer abc123\r\n");
	}

	[Fact(Timeout = TimeoutMs)]
	public async Task CreateWebSocket_WithoutToken_SendsNoAuthorization()
	{
		var request = await CaptureUpgradeRequestAsync(new AuthorizedWebSocketFactory(() => null), TestContext.Current.CancellationToken);

		request.Should().NotContain("Authorization:");
	}
}
