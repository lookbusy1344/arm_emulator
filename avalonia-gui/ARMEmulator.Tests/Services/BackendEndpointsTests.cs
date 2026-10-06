using ARMEmulator.Services;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

public sealed class BackendEndpointsTests
{
	[Theory]
	[InlineData("http://localhost:8080", "ws://localhost:8080/api/v1/ws")]
	[InlineData("http://localhost:8080/", "ws://localhost:8080/api/v1/ws")]
	[InlineData("https://emulator.example:9443", "wss://emulator.example:9443/api/v1/ws")]
	public void WebSocketUri_MapsSchemeAndAppendsWebSocketPath(string baseUrl, string expected)
	{
		BackendEndpoints.WebSocketUri(new Uri(baseUrl)).Should().Be(new Uri(expected));
	}

	[Fact]
	public void WebSocketUri_WithNonHttpScheme_ThrowsArgumentException()
	{
		var act = () => BackendEndpoints.WebSocketUri(new Uri("ftp://localhost:8080"));

		act.Should().Throw<ArgumentException>().WithParameterName("baseUri");
	}
}
