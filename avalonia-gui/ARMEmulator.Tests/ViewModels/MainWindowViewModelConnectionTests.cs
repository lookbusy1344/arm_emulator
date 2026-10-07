using System.Diagnostics.CodeAnalysis;
using System.Reactive;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using AwesomeAssertions;
using NSubstitute;
using NSubstitute.ExceptionExtensions;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// The connection state behind the "starting" and "failed" view.
/// </summary>
public sealed class MainWindowViewModelConnectionTests : IDisposable
{
	private readonly IApiClient api = Substitute.For<IApiClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient ws = Substitute.For<IWebSocketClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IFileService files = Substitute.For<IFileService>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IBackendManager backend = Substitute.For<IBackendManager>();

	private readonly Subject<EmulatorEvent> events = new();
	private readonly Subject<BackendStatus> backendStatus = new();

	public MainWindowViewModelConnectionTests()
	{
		ws.Events.Returns(events);
		files.RecentFilesChanged.Returns(Observable.Never<Unit>());
		files.RecentFiles.Returns([]);
		backend.StatusChanged.Returns(backendStatus);
		backend.Status.Returns(BackendStatus.Running);
		api.CreateSessionAsync(Arg.Any<CancellationToken>()).Returns(new SessionInfo("s1"));
		api.GetMemoryAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<byte>.Empty);
		api.GetDisassemblyAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<DisassemblyInstruction>.Empty);
		api.GetRegistersAsync(Arg.Any<string>(), Arg.Any<CancellationToken>()).Returns(RegisterState.Create());
	}

	public void Dispose()
	{
		events.Dispose();
		backendStatus.Dispose();
	}

	private static CancellationToken Ct => TestContext.Current.CancellationToken;

	[Fact]
	public void Connection_BeforeStart_IsConnecting()
	{
		using var vm = new MainWindowViewModel(api, ws, files);

		vm.Connection.Should().Be(ConnectionState.Connecting);
		vm.ConnectionFailure.Should().BeNull();
	}

	[Fact]
	public async Task Connection_AfterASuccessfulStart_IsConnected()
	{
		using var vm = new MainWindowViewModel(api, ws, files);

		await vm.StartAsync(backend, Ct);

		vm.Connection.Should().Be(ConnectionState.Connected);
	}

	[Fact]
	public async Task Connection_WhenTheBackendFailsToStart_IsFailedWithTheReason()
	{
		backend.StartAsync(Arg.Any<CancellationToken>()).ThrowsAsync(new BackendStartException("binary not found"));
		using var vm = new MainWindowViewModel(api, ws, files);

		await vm.StartAsync(backend, Ct);

		vm.Connection.Should().Be(ConnectionState.Failed);
		vm.ConnectionFailure.Should().Be("Failed to start backend: binary not found");
	}

	[Fact]
	public async Task Connection_WhenTheSessionCannotBeCreated_IsFailedWithTheReason()
	{
		api.CreateSessionAsync(Arg.Any<CancellationToken>()).ThrowsAsync(new ApiException("refused"));
		using var vm = new MainWindowViewModel(api, ws, files);

		await vm.StartAsync(backend, Ct);

		vm.Connection.Should().Be(ConnectionState.Failed);
		vm.ConnectionFailure.Should().Be("Failed to connect to backend: refused");
	}

	[Fact]
	public async Task Connection_StaysFailedWhenTheErrorBarIsDismissed()
	{
		backend.StartAsync(Arg.Any<CancellationToken>()).ThrowsAsync(new BackendStartException("binary not found"));
		using var vm = new MainWindowViewModel(api, ws, files);
		await vm.StartAsync(backend, Ct);

		await vm.DismissErrorCommand.Execute();

		vm.Connection.Should().Be(ConnectionState.Failed);
	}

	[Fact]
	public async Task Retry_AfterAFailureThatNowSucceeds_IsConnectedWithNoFailure()
	{
		backend.StartAsync(Arg.Any<CancellationToken>()).Returns(
			_ => Task.FromException(new BackendStartException("binary not found")),
			_ => Task.CompletedTask);
		using var vm = new MainWindowViewModel(api, ws, files);
		await vm.StartAsync(backend, Ct);

		await vm.RestartBackendCommand.Execute();

		vm.Connection.Should().Be(ConnectionState.Connected);
		vm.ConnectionFailure.Should().BeNull();
	}

	[Fact]
	public async Task Connection_RaisesPropertyChangedWhenItChanges()
	{
		using var vm = new MainWindowViewModel(api, ws, files);
		var changed = new List<string?>();
		vm.PropertyChanged += (_, args) => changed.Add(args.PropertyName);

		await vm.StartAsync(backend, Ct);

		changed.Should().Contain(nameof(MainWindowViewModel.Connection));
	}
}
