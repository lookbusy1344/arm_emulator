using System.Diagnostics;
using System.Net;
using System.Runtime.Versioning;
using ARMEmulator.Services;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

/// <summary>
/// Tests for BackendManager. Spawning tests run a shell script that records its process ID and then sleeps, standing in
/// for a backend that never becomes healthy.
/// </summary>
public sealed class BackendManagerTests : IDisposable
{
	private readonly DirectoryInfo directory = Directory.CreateTempSubdirectory("arm-backend-tests");

	public void Dispose() => directory.Delete(recursive: true);

	private string PidFile => Path.Combine(directory.FullName, "backend.pid");

	/// <summary>Writes a stand-in backend that records its PID and then sleeps until killed.</summary>
	[UnsupportedOSPlatform("windows")]
	private string WriteFakeBackend()
	{
		var script = Path.Combine(directory.FullName, "fake-backend.sh");
		File.WriteAllText(script, $"#!/bin/sh\necho $$ > '{PidFile}'\nexec sleep 600\n");
		File.SetUnixFileMode(script, UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute);
		return script;
	}

	private string TermMarker => Path.Combine(directory.FullName, "backend.term");

	/// <summary>
	/// Writes a stand-in backend that records its PID and runs until signalled. With <paramref name="handlesTerm"/> it
	/// records SIGTERM and exits; otherwise it ignores SIGTERM, like a backend stuck in shutdown.
	/// </summary>
	[UnsupportedOSPlatform("windows")]
	private string WriteSignalAwareBackend(bool handlesTerm)
	{
		var script = Path.Combine(directory.FullName, "signal-backend.sh");
		var trap = handlesTerm ? $"trap 'echo term > \"{TermMarker}\"; exit 0' TERM" : "trap '' TERM";
		File.WriteAllText(script, $"#!/bin/sh\n{trap}\necho $$ > '{PidFile}'\nwhile :; do sleep 0.05; done\n");
		File.SetUnixFileMode(script, UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute);
		return script;
	}

	private bool FakeBackendIsRunning()
	{
		var pid = int.Parse(File.ReadAllText(PidFile).Trim(), System.Globalization.CultureInfo.InvariantCulture);
		try {
			using var process = Process.GetProcessById(pid);
			return !process.HasExited;
		}
		catch (ArgumentException) {
			return false;
		}
	}

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

	[Fact]
	public async Task StartAsync_WhenCancelled_ThrowsOperationCanceledAndReportsStopped()
	{
		using var handler = new TestHttpMessageHandler();
		handler.SetException(new HttpRequestException("Connection refused"));
		using var manager = new BackendManager("http://localhost:18080", handler, findBinary: static () => null);
		using var cancelled = new CancellationTokenSource();
		await cancelled.CancelAsync();

		var act = async () => await manager.StartAsync(cancelled.Token);

		await act.Should().ThrowAsync<OperationCanceledException>();
		manager.Status.Should().Be(BackendStatus.Stopped);
	}

	[Fact]
	public void BackendStartException_IsNotAnApiError()
	{
		// Handlers that catch ApiException report "API error"; a process that failed to start is not one
		new BackendStartException("binary not found").Should().NotBeAssignableTo<ApiException>();
	}

	[Fact]
	public void Constructor_WithRelativeOrMalformedUrl_ThrowsArgumentException()
	{
		var act = () => new BackendManager("localhost:8080/api");

		act.Should().ThrowExactly<ArgumentException>().WithParameterName("baseUrl");
	}

	[Fact]
	[UnsupportedOSPlatform("windows")]
	public async Task StartAsync_WhenHealthCheckNeverPasses_StopsTheSpawnedProcess()
	{
		Assert.SkipWhen(OperatingSystem.IsWindows(), "Uses a POSIX shell script as the backend.");
		using var handler = new TestHttpMessageHandler();
		handler.SetException(new HttpRequestException("Connection refused"));
		var backend = WriteFakeBackend();
		using var manager = new BackendManager("http://localhost:18080", handler, findBinary: () => backend);

		var act = async () => await manager.StartAsync(TestContext.Current.CancellationToken);

		await act.Should().ThrowAsync<BackendStartException>().WithMessage("Backend started but health check failed*");
		FakeBackendIsRunning().Should().BeFalse();
		manager.Status.Should().Be(BackendStatus.Error);
	}

	[Fact]
	[UnsupportedOSPlatform("windows")]
	public async Task StartAsync_WhenCancelledWhileWaitingForHealth_StopsTheSpawnedProcess()
	{
		Assert.SkipWhen(OperatingSystem.IsWindows(), "Uses a POSIX shell script as the backend.");
		using var cancellation = new CancellationTokenSource();
		using var handler = new CancellingHandler(cancellation, () => File.Exists(PidFile));
		var backend = WriteFakeBackend();
		using var manager = new BackendManager("http://localhost:18080", handler, findBinary: () => backend);

		var act = async () => await manager.StartAsync(cancellation.Token);

		await act.Should().ThrowAsync<OperationCanceledException>();
		FakeBackendIsRunning().Should().BeFalse();
		manager.Status.Should().Be(BackendStatus.Stopped);
	}

	[Fact]
	public async Task HealthCheckAsync_WhenCancelled_ThrowsOperationCanceled()
	{
		using var handler = new TestHttpMessageHandler();
		handler.SetException(new HttpRequestException("Connection refused"));
		using var manager = new BackendManager("http://localhost:18080", handler);
		using var cancelled = new CancellationTokenSource();
		await cancelled.CancelAsync();

		var act = async () => await manager.HealthCheckAsync(cancelled.Token);

		await act.Should().ThrowAsync<OperationCanceledException>();
	}

	[Fact]
	public async Task HealthCheckAsync_WhenBackendUnreachable_ReturnsFalse()
	{
		using var handler = new TestHttpMessageHandler();
		handler.SetException(new HttpRequestException("Connection refused"));
		using var manager = new BackendManager("http://localhost:18080", handler);

		var healthy = await manager.HealthCheckAsync(TestContext.Current.CancellationToken);

		healthy.Should().BeFalse();
	}

	private const int ProcessTestTimeoutMs = 5000;

	[Fact(Timeout = ProcessTestTimeoutMs)]
	[UnsupportedOSPlatform("windows")]
	public async Task StopAsync_SendsSigtermSoTheBackendCanCleanUp()
	{
		Assert.SkipWhen(OperatingSystem.IsWindows(), "Signals are POSIX only.");
		var ct = TestContext.Current.CancellationToken;
		using var handler = new HealthyWhenHandler(() => File.Exists(PidFile));
		var backend = WriteSignalAwareBackend(handlesTerm: true);
		using var manager = new BackendManager("http://localhost:18080", handler, findBinary: () => backend);
		await manager.StartAsync(ct);

		await manager.StopAsync();

		(await File.ReadAllTextAsync(TermMarker, ct)).Should().Be("term\n");
		FakeBackendIsRunning().Should().BeFalse();
		manager.Status.Should().Be(BackendStatus.Stopped);
	}

	[Fact(Timeout = ProcessTestTimeoutMs)]
	[UnsupportedOSPlatform("windows")]
	public async Task StopAsync_KillsABackendThatIgnoresSigterm()
	{
		Assert.SkipWhen(OperatingSystem.IsWindows(), "Signals are POSIX only.");
		var ct = TestContext.Current.CancellationToken;
		using var handler = new HealthyWhenHandler(() => File.Exists(PidFile));
		var backend = WriteSignalAwareBackend(handlesTerm: false);
		using var manager = new BackendManager("http://localhost:18080", handler, findBinary: () => backend) {
			StopGracePeriod = TimeSpan.FromMilliseconds(200)
		};
		await manager.StartAsync(ct);

		await manager.StopAsync();

		FakeBackendIsRunning().Should().BeFalse();
		manager.Status.Should().Be(BackendStatus.Stopped);
	}

	/// <summary>Refuses connections until <paramref name="healthy"/> holds, then answers 200.</summary>
	private sealed class HealthyWhenHandler(Func<bool> healthy) : HttpMessageHandler
	{
		protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken) =>
			healthy()
				? Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK))
				: throw new HttpRequestException("Connection refused");
	}

	/// <summary>Refuses every connection, and cancels <paramref name="cancellation"/> on the first request after <paramref name="when"/> holds.</summary>
	private sealed class CancellingHandler(CancellationTokenSource cancellation, Func<bool> when) : HttpMessageHandler
	{
		protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
		{
			if (when()) {
				cancellation.Cancel();
			}

			throw new HttpRequestException("Connection refused");
		}
	}
}
