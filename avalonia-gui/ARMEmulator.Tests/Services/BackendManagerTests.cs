using System.Net;
using ARMEmulator.Services;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

/// <summary>
/// Tests for BackendManager. Process spawning is covered by running the app;
/// these tests cover argument construction and reuse of an already running backend.
/// </summary>
public sealed class BackendManagerTests
{
	[Fact]
	public void Constructor_InitializesWithStoppedStatus()
	{
		using var manager = new BackendManager();
		manager.Status.Should().Be(BackendStatus.Stopped);
		manager.BaseUrl.Should().Be("http://localhost:8080");
	}

	[Theory]
	[InlineData("http://localhost:8080", "8080")]
	[InlineData("http://127.0.0.1:18080", "18080")]
	[InlineData("http://localhost", "80")]
	public void ServerArguments_StartApiServerOnTheUrlPort(string baseUrl, string port)
	{
		BackendManager.ServerArguments(new Uri(baseUrl)).Should().Equal("-api-server", "-port", port);
	}

	[Fact]
	public async Task StartAsync_WhenBackendAlreadyHealthy_ReportsRunningWithoutSpawning()
	{
		using var handler = new TestHttpMessageHandler();
		handler.SetResponse(HttpStatusCode.OK, "");
		using var manager = new BackendManager("http://localhost:18080", handler);

		await manager.StartAsync(TestContext.Current.CancellationToken);

		manager.Status.Should().Be(BackendStatus.Running);
		handler.LastRequest!.RequestUri.Should().Be(new Uri("http://localhost:18080/health"));
	}

	[Fact]
	public async Task StartAsync_WhenBackendUnhealthyAndBinaryMissing_ThrowsAndReportsError()
	{
		using var handler = new TestHttpMessageHandler();
		handler.SetException(new HttpRequestException("Connection refused"));
		using var manager = new BackendManager("http://localhost:18080", handler, findBinary: static () => null);

		var act = async () => await manager.StartAsync(TestContext.Current.CancellationToken);

		await act.Should().ThrowAsync<BackendStartException>().WithMessage("Backend binary not found*");
		manager.Status.Should().Be(BackendStatus.Error);
	}
}
