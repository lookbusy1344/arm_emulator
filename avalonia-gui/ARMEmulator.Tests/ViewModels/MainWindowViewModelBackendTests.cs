using System.Diagnostics.CodeAnalysis;
using System.Reactive;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using System.Windows.Input;
using ARMEmulator.Collections;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using AwesomeAssertions;
using NSubstitute;
using NSubstitute.ExceptionExtensions;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// Tests for backend status display and the Restart backend command.
/// </summary>
public sealed class MainWindowViewModelBackendTests : IDisposable
{
	private const string OldSessionId = "old";
	private const string NewSessionId = "new";
	private const string Program = "_start:\n  MOV R0, #5\n";

	private readonly IApiClient api = Substitute.For<IApiClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient ws = Substitute.For<IWebSocketClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IFileService files = Substitute.For<IFileService>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IBackendManager backend = Substitute.For<IBackendManager>();

	private readonly Subject<EmulatorEvent> events = new();
	private readonly Subject<BackendStatus> backendStatus = new();

	public MainWindowViewModelBackendTests()
	{
		ws.Events.Returns(events);
		files.RecentFilesChanged.Returns(Observable.Never<Unit>());
		files.RecentFiles.Returns([]);
		backend.StatusChanged.Returns(backendStatus);
		backend.Status.Returns(BackendStatus.Running);
		api.GetMemoryAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<byte>.Empty);
		api.GetDisassemblyAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<DisassemblyInstruction>.Empty);
		api.CreateSessionAsync(Arg.Any<CancellationToken>()).Returns(new SessionInfo(OldSessionId), new SessionInfo(NewSessionId));
		api.GetRegistersAsync(Arg.Any<string>(), Arg.Any<CancellationToken>()).Returns(RegisterState.Create());
		api.GetSourceMapAsync(Arg.Any<string>(), Arg.Any<CancellationToken>()).Returns([new SourceMapEntry(0x8000, 2, "  MOV R0, #5")]);
		api.LoadProgramAsync(Arg.Any<string>(), Arg.Any<string>(), Arg.Any<CancellationToken>())
			.Returns(new LoadProgramResponse(EquatableDictionaryFactory.CopyOf(new Dictionary<string, uint>())));
	}

	public void Dispose()
	{
		events.Dispose();
		backendStatus.Dispose();
	}

	private static CancellationToken Ct => TestContext.Current.CancellationToken;

	private async Task<MainWindowViewModel> StartedViewModelAsync()
	{
		var vm = new MainWindowViewModel(api, ws, files);
		await vm.StartAsync(backend, Ct);
		api.ClearReceivedCalls();
		ws.ClearReceivedCalls();
		backend.ClearReceivedCalls();
		return vm;
	}

	private static void Execute(ICommand command) => command.Execute(null);

	private static async Task RestartAsync(MainWindowViewModel vm)
	{
		await vm.RestartBackendCommand.Execute();
	}

	// Status display

	[Fact]
	public async Task BackendStatus_FollowsBackendManagerNotifications()
	{
		using var vm = await StartedViewModelAsync();

		backendStatus.OnNext(BackendStatus.Error);

		vm.BackendStatus.Should().Be(BackendStatus.Error);
	}

	[Fact]
	public async Task StatusText_IncludesBackendState()
	{
		using var vm = await StartedViewModelAsync();

		vm.StatusText.Should().Be("Idle (backend: Running)");

		backendStatus.OnNext(BackendStatus.Error);

		vm.StatusText.Should().Be("Idle (backend: Error)");
	}

	// Restart backend

	[Fact]
	public async Task RestartBackend_StopsStartsAndCreatesSessionInOrder()
	{
		using var vm = await StartedViewModelAsync();

		await RestartAsync(vm);

		Received.InOrder(() => {
			_ = ws.DisconnectAsync();
			_ = backend.StopAsync();
			_ = backend.StartAsync(Arg.Any<CancellationToken>());
			_ = api.CreateSessionAsync(Arg.Any<CancellationToken>());
			_ = ws.ConnectAsync(NewSessionId, Arg.Any<CancellationToken>());
		});
		vm.SessionId.Should().Be(NewSessionId);
		vm.IsConnected.Should().BeTrue();
		vm.ErrorMessage.Should().BeNull();
	}

	[Fact]
	public async Task RestartBackend_DoesNotDestroyTheOldSessionOnTheDeadBackend()
	{
		using var vm = await StartedViewModelAsync();

		await RestartAsync(vm);

		_ = api.DidNotReceiveWithAnyArgs().DestroySessionAsync(default!, TestContext.Current.CancellationToken);
	}

	[Fact]
	public async Task RestartBackend_ReloadsTheLoadedSourceIntoTheNewSession()
	{
		using var vm = await StartedViewModelAsync();
		vm.SourceCode = Program;
		Execute(vm.AssembleCommand);
		api.ClearReceivedCalls();

		await RestartAsync(vm);

		Received.InOrder(() => {
			_ = api.CreateSessionAsync(Arg.Any<CancellationToken>());
			_ = api.LoadProgramAsync(NewSessionId, Program, Arg.Any<CancellationToken>());
		});
		vm.AddressToLine.Should().ContainKey(0x8000u);
	}

	[Fact]
	public async Task RestartBackend_WithNoProgramLoaded_LoadsNothing()
	{
		using var vm = await StartedViewModelAsync();
		vm.SourceCode = Program;

		await RestartAsync(vm);

		_ = api.DidNotReceiveWithAnyArgs().LoadProgramAsync(default!, default!, TestContext.Current.CancellationToken);
	}

	[Fact]
	public async Task RestartBackend_WhenStopFails_ReportsErrorAndStartsNothing()
	{
		backend.StopAsync().ThrowsAsync(new InvalidOperationException("kill failed"));
		using var vm = await StartedViewModelAsync();
		backend.StopAsync().ThrowsAsync(new InvalidOperationException("kill failed"));

		await RestartAsync(vm);

		vm.ErrorMessage.Should().Be("Failed to stop backend: kill failed");
		_ = backend.DidNotReceive().StartAsync(Arg.Any<CancellationToken>());
	}

	[Fact]
	public async Task RestartBackend_WhenStartFails_ReportsErrorAndCreatesNoSession()
	{
		using var vm = await StartedViewModelAsync();
		backend.StartAsync(Arg.Any<CancellationToken>()).ThrowsAsync(new BackendStartException("port in use"));

		await RestartAsync(vm);

		vm.ErrorMessage.Should().Be("Failed to start backend: port in use");
		_ = api.DidNotReceive().CreateSessionAsync(Arg.Any<CancellationToken>());
		vm.SessionId.Should().BeNull();
		vm.IsConnected.Should().BeFalse();
	}

	[Fact]
	public async Task RestartBackend_WhenSessionCreationFails_ReportsErrorAndLoadsNothing()
	{
		using var vm = await StartedViewModelAsync();
		vm.SourceCode = Program;
		Execute(vm.AssembleCommand);
		api.ClearReceivedCalls();
		api.CreateSessionAsync(Arg.Any<CancellationToken>()).ThrowsAsync(new BackendUnavailableException("refused"));

		await RestartAsync(vm);

		vm.ErrorMessage.Should().Be("Failed to connect to backend: refused");
		_ = api.DidNotReceiveWithAnyArgs().LoadProgramAsync(default!, default!, TestContext.Current.CancellationToken);
	}

	[Fact]
	public void RestartBackend_BeforeAnyStart_ReportsNoBackend()
	{
		using var vm = new MainWindowViewModel(api, ws, files);

		Execute(vm.RestartBackendCommand);

		vm.ErrorMessage.Should().Be("Restart backend failed: no backend to restart");
	}
}
