using System.Diagnostics.CodeAnalysis;
using System.Reactive;
using System.Reactive.Linq;
using System.Reactive.Subjects;
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
/// Tests for MainWindowViewModel startup, program loading and child view model wiring.
/// </summary>
public sealed class MainWindowViewModelLoadTests : IDisposable
{
	private const string SessionId = "s1";
	private const int MemoryWindowBytes = 256;
	private const int DisassemblyWindowInstructions = 64;
	private const uint DisassemblyLeadBytes = 32 * 4;

	private const string Program = """
		.org 0x8000
		_start:
		  MOV R0, #5
		  SWI #0
		""";

	private readonly IApiClient api = Substitute.For<IApiClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient ws = Substitute.For<IWebSocketClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IFileService files = Substitute.For<IFileService>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IBackendManager backend = Substitute.For<IBackendManager>();

	private readonly Subject<EmulatorEvent> events = new();

	public MainWindowViewModelLoadTests()
	{
		ws.Events.Returns(events);
		files.RecentFilesChanged.Returns(Observable.Never<Unit>());
		files.RecentFiles.Returns([]);
		api.GetMemoryAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<byte>.Empty);
		api.GetDisassemblyAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<DisassemblyInstruction>.Empty);
	}

	public void Dispose() => events.Dispose();

	private static CancellationToken Ct => TestContext.Current.CancellationToken;

	private MainWindowViewModel CreateViewModel(string? sessionId = SessionId) =>
		new(api, ws, files) { SessionId = sessionId };

	private void StubSuccessfulLoad(string source, RegisterState registers)
	{
		api.LoadProgramAsync(SessionId, source, Arg.Any<CancellationToken>())
			.Returns(new LoadProgramResponse(EquatableDictionaryFactory.CopyOf(new Dictionary<string, uint> { ["_start"] = 0x8000 })));
		api.GetSourceMapAsync(SessionId, Arg.Any<CancellationToken>())
			.Returns([new SourceMapEntry(0x8000, 3, "  MOV R0, #5"), new SourceMapEntry(0x8004, 4, "  SWI #0")]);
		api.GetRegistersAsync(SessionId, Arg.Any<CancellationToken>()).Returns(registers);
	}

	// Startup

	[Fact]
	public async Task StartAsync_StartsBackendThenCreatesSessionAndConnects()
	{
		api.CreateSessionAsync(Arg.Any<CancellationToken>()).Returns(new SessionInfo(SessionId));
		using var vm = CreateViewModel(sessionId: null);

		await vm.StartAsync(backend, Ct);

		Received.InOrder(() => {
			_ = backend.StartAsync(Arg.Any<CancellationToken>());
			_ = api.CreateSessionAsync(Arg.Any<CancellationToken>());
			_ = ws.ConnectAsync(SessionId, Arg.Any<CancellationToken>());
		});
		vm.SessionId.Should().Be(SessionId);
		vm.IsConnected.Should().BeTrue();
		vm.ErrorMessage.Should().BeNull();
	}

	[Fact]
	public async Task StartAsync_WhenBackendFailsToStart_ReportsErrorAndCreatesNoSession()
	{
		backend.StartAsync(Arg.Any<CancellationToken>()).ThrowsAsync(new BackendStartException("Backend binary not found"));
		using var vm = CreateViewModel(sessionId: null);

		await vm.StartAsync(backend, Ct);

		vm.ErrorMessage.Should().Be("Failed to start backend: Backend binary not found");
		_ = api.DidNotReceive().CreateSessionAsync(Arg.Any<CancellationToken>());
		vm.SessionId.Should().BeNull();
		vm.IsConnected.Should().BeFalse();
	}

	[Fact]
	public async Task StartAsync_WhenSessionCreationFails_ReportsError()
	{
		api.CreateSessionAsync(Arg.Any<CancellationToken>())
			.ThrowsAsync(new BackendUnavailableException("Cannot connect to backend - is the emulator running?"));
		using var vm = CreateViewModel(sessionId: null);

		await vm.StartAsync(backend, Ct);

		vm.ErrorMessage.Should().Be("Failed to connect to backend: Cannot connect to backend - is the emulator running?");
		vm.SessionId.Should().BeNull();
		vm.IsConnected.Should().BeFalse();
	}

	[Fact]
	public async Task StartAsync_WhenWebSocketFails_ReportsErrorAndStaysDisconnected()
	{
		api.CreateSessionAsync(Arg.Any<CancellationToken>()).Returns(new SessionInfo(SessionId));
		ws.ConnectAsync(SessionId, Arg.Any<CancellationToken>())
			.ThrowsAsync(new WebSocketConnectionException("Failed to connect to ws://localhost:8080/api/v1/ws"));
		using var vm = CreateViewModel(sessionId: null);

		await vm.StartAsync(backend, Ct);

		vm.ErrorMessage.Should().Be("Failed to connect to backend: Failed to connect to ws://localhost:8080/api/v1/ws");
		vm.IsConnected.Should().BeFalse();
	}

	// Program loading

	[Fact]
	public async Task LoadProgram_WithoutSession_ReportsErrorAndCallsNothing()
	{
		using var vm = CreateViewModel(sessionId: null);
		vm.SourceCode = Program;

		await vm.LoadProgramCommand.Execute();

		vm.ErrorMessage.Should().Be("No active session");
		_ = api.DidNotReceiveWithAnyArgs().LoadProgramAsync(default!, default!, Ct);
	}

	[Fact]
	public async Task LoadProgram_Success_LoadsSourceBuildsLineMapsAndResetsState()
	{
		var registers = RegisterState.Create(sp: 0x50000, pc: 0x8000);
		StubSuccessfulLoad(Program, registers);
		using var vm = CreateViewModel();
		vm.SourceCode = Program;
		vm.Status = VMState.Halted;
		vm.ErrorMessage = "stale error";
		vm.LastMemoryWrite = new MemoryWrite(0x9000, 4);

		await vm.LoadProgramCommand.Execute();

		_ = api.Received(1).LoadProgramAsync(SessionId, Program, Arg.Any<CancellationToken>());
		vm.AddressToLine.Should().BeEquivalentTo(new Dictionary<uint, int> { [0x8000] = 3, [0x8004] = 4 });
		vm.LineToAddress.Should().BeEquivalentTo(new Dictionary<int, uint> { [3] = 0x8000, [4] = 0x8004 });
		vm.ValidBreakpointLines.Should().BeEquivalentTo([3, 4]);
		vm.Registers.Should().Be(registers);
		vm.Status.Should().Be(VMState.Idle);
		vm.ErrorMessage.Should().BeNull();
		vm.LastMemoryWrite.Should().BeNull();
	}

	[Fact]
	public async Task LoadProgram_Success_DoesNotHighlightRegistersCarriedFromPreviousProgram()
	{
		StubSuccessfulLoad(Program, RegisterState.Create(r0: 0, pc: 0x8000));
		using var vm = CreateViewModel();
		vm.UpdateRegisters(RegisterState.Create(r0: 1));
		vm.UpdateRegisters(RegisterState.Create(r0: 2, pc: 0x9000));
		vm.SourceCode = Program;

		await vm.LoadProgramCommand.Execute();

		vm.PreviousRegisters.Should().BeNull();
		vm.ChangedRegisters.Should().BeEmpty();
	}

	[Fact]
	public async Task LoadProgram_WithAssemblerErrors_ReportsEveryErrorAndClearsLineMaps()
	{
		StubSuccessfulLoad(Program, RegisterState.Create(pc: 0x8000));
		using var vm = CreateViewModel();
		vm.SourceCode = Program;
		await vm.LoadProgramCommand.Execute();

		const string broken = "_start:\n  FOO R0\n";
		api.LoadProgramAsync(SessionId, broken, Arg.Any<CancellationToken>())
			.ThrowsAsync(new ProgramLoadException(["api:2:2: unknown instruction: FOO", "api:3:1: undefined label: bar"]));
		vm.SourceCode = broken;

		await vm.LoadProgramCommand.Execute();

		vm.ErrorMessage.Should().Be("Failed to load program:\napi:2:2: unknown instruction: FOO\napi:3:1: undefined label: bar");
		vm.AddressToLine.Should().BeEmpty();
		vm.LineToAddress.Should().BeEmpty();
		vm.ValidBreakpointLines.Should().BeEmpty();
		_ = api.Received(1).GetSourceMapAsync(SessionId, Arg.Any<CancellationToken>());
	}

	[Fact]
	public async Task LoadProgram_WhenSessionExpired_ReportsApiMessage()
	{
		api.LoadProgramAsync(SessionId, Program, Arg.Any<CancellationToken>()).ThrowsAsync(new SessionNotFoundException(SessionId));
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.LoadProgramCommand.Execute();

		vm.ErrorMessage.Should().Be("Failed to load program: Session 's1' not found or expired");
	}

	[Fact]
	public async Task LoadProgram_WhenSourceMapFetchFails_ReportsErrorAndLeavesLineMapsEmpty()
	{
		StubSuccessfulLoad(Program, RegisterState.Create(pc: 0x8000));
		api.GetSourceMapAsync(SessionId, Arg.Any<CancellationToken>()).ThrowsAsync(new ApiException("API error: boom"));
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.LoadProgramCommand.Execute();

		vm.ErrorMessage.Should().Be("Failed to load program: API error: boom");
		vm.AddressToLine.Should().BeEmpty();
		vm.LineToAddress.Should().BeEmpty();
	}

	// WebSocket state without registers

	[Fact]
	public void WebSocket_StatusOnlyStateEvent_UpdatesStatusAndKeepsRegisters()
	{
		using var vm = CreateViewModel();
		var registers = RegisterState.Create(r0: 7, pc: 0x8004);
		vm.UpdateRegisters(registers);

		events.OnNext(new StateEvent(SessionId, new VMStatus(VMState.WaitingForInput, 0, 0), null));

		vm.Status.Should().Be(VMState.WaitingForInput);
		vm.Registers.Should().Be(registers);
	}

	// Child view models

	[Fact]
	public void SessionId_IsPropagatedToChildViewModels()
	{
		using var vm = CreateViewModel(sessionId: null);

		vm.SessionId = SessionId;

		vm.Memory.SessionId.Should().Be(SessionId);
		vm.Stack.SessionId.Should().Be(SessionId);
		vm.Disassembly.SessionId.Should().Be(SessionId);
	}

	[Fact]
	public async Task UpdateRegisters_PushesRegistersToChildViewModels()
	{
		using var vm = CreateViewModel();

		vm.UpdateRegisters(RegisterState.Create(sp: 0x4FFF0, lr: 0x8010, pc: 0x8008));
		await vm.Memory.JumpToPCCommand.Execute();

		vm.Stack.StackPointer.Should().Be(0x4FFF0u);
		vm.Stack.LinkRegister.Should().Be(0x8010u);
		vm.Disassembly.ProgramCounter.Should().Be(0x8008u);
		_ = api.Received(1).GetMemoryAsync(SessionId, 0x8008, MemoryWindowBytes, Arg.Any<CancellationToken>());
	}

	[Fact]
	public void Breakpoints_ArePushedToDisassembly()
	{
		using var vm = CreateViewModel();

		vm.Breakpoints = [0x8004, 0x800C];

		vm.Disassembly.Breakpoints.Should().BeEquivalentTo([0x8004u, 0x800Cu]);
	}

	[Fact]
	public void LastMemoryWrite_IsPushedToMemoryAndScrollsToIt()
	{
		using var vm = CreateViewModel();

		vm.LastMemoryWrite = new MemoryWrite(0x9000, 4);

		vm.Memory.LastWriteAddress.Should().Be(0x9000u);
		_ = api.Received(1).GetMemoryAsync(SessionId, 0x9000, MemoryWindowBytes, Arg.Any<CancellationToken>());
	}

	[Fact]
	public void RegisterChange_WhenStopped_RefreshesStackAndDisassembly()
	{
		using var vm = CreateViewModel();
		vm.Status = VMState.Breakpoint;

		vm.UpdateRegisters(RegisterState.Create(sp: 0x4FF00, pc: 0x8100));

		_ = api.Received(1).GetDisassemblyAsync(SessionId, 0x8100 - DisassemblyLeadBytes, DisassemblyWindowInstructions, Arg.Any<CancellationToken>());
		_ = api.Received(1).GetMemoryAsync(SessionId, 0x4FF00, MemoryWindowBytes, Arg.Any<CancellationToken>());
	}

	[Fact]
	public void RegisterChange_WhileRunning_DefersRefreshUntilStopped()
	{
		using var vm = CreateViewModel();
		vm.Status = VMState.Running;

		vm.UpdateRegisters(RegisterState.Create(sp: 0x4FF00, pc: 0x8100));

		_ = api.DidNotReceiveWithAnyArgs().GetDisassemblyAsync(default!, default, default, Ct);

		vm.Status = VMState.Breakpoint;

		_ = api.Received(1).GetDisassemblyAsync(SessionId, 0x8100 - DisassemblyLeadBytes, DisassemblyWindowInstructions, Arg.Any<CancellationToken>());
	}

	[Fact]
	public void RegisterChange_WithIdenticalRegisters_RefreshesOnce()
	{
		using var vm = CreateViewModel();
		vm.Status = VMState.Breakpoint;

		vm.UpdateRegisters(RegisterState.Create(pc: 0x8100));
		vm.UpdateRegisters(RegisterState.Create(pc: 0x8100));
		vm.Status = VMState.Idle;

		_ = api.Received(1).GetDisassemblyAsync(SessionId, 0x8100 - DisassemblyLeadBytes, DisassemblyWindowInstructions, Arg.Any<CancellationToken>());
	}

	[Fact]
	public void RegisterChange_WithoutSession_DoesNotRefresh()
	{
		using var vm = CreateViewModel(sessionId: null);
		vm.Status = VMState.Breakpoint;

		vm.UpdateRegisters(RegisterState.Create(pc: 0x8100));

		_ = api.DidNotReceiveWithAnyArgs().GetDisassemblyAsync(default!, default, default, Ct);
	}

	// Console

	[Fact]
	public async Task LoadProgram_Success_ClearsTheConsole()
	{
		StubSuccessfulLoad(Program, RegisterState.Create());
		using var vm = CreateViewModel();
		vm.SourceCode = Program;
		events.OnNext(new OutputEvent(SessionId, OutputStreamType.Stdout, "old output\n"));

		await vm.LoadProgramCommand.Execute();

		vm.ConsoleOutput.Should().BeEmpty();
	}

	[Fact]
	public async Task LoadProgram_Failure_KeepsTheConsole()
	{
		api.LoadProgramAsync(SessionId, Program, Arg.Any<CancellationToken>())
			.ThrowsAsync(new ProgramLoadException(["line 2: bad"]));
		using var vm = CreateViewModel();
		vm.SourceCode = Program;
		events.OnNext(new OutputEvent(SessionId, OutputStreamType.Stdout, "old output\n"));

		await vm.LoadProgramCommand.Execute();

		vm.ConsoleOutput.Should().Be("old output\n");
	}

	[Fact]
	public void ConsoleOutput_BelowTheCap_KeepsEverything()
	{
		using var vm = CreateViewModel();

		events.OnNext(new OutputEvent(SessionId, OutputStreamType.Stdout, "one\n"));
		events.OnNext(new OutputEvent(SessionId, OutputStreamType.Stdout, "two\n"));

		vm.ConsoleOutput.Should().Be("one\ntwo\n");
	}

	[Fact]
	public void ConsoleOutput_AboveTheCap_KeepsTheNewestText()
	{
		const int extra = 10;
		using var vm = CreateViewModel();
		var older = new string('a', MainWindowViewModel.MaxConsoleCharacters);
		var newer = new string('b', extra);

		events.OnNext(new OutputEvent(SessionId, OutputStreamType.Stdout, older));
		events.OnNext(new OutputEvent(SessionId, OutputStreamType.Stdout, newer));

		vm.ConsoleOutput.Length.Should().Be(MainWindowViewModel.MaxConsoleCharacters);
		vm.ConsoleOutput.Should().EndWith(newer);
		vm.ConsoleOutput.Should().StartWith(new string('a', MainWindowViewModel.MaxConsoleCharacters - extra));
	}

	[Fact]
	public void ConsoleOutput_CutThroughASurrogatePair_DropsTheOrphanedHalf()
	{
		const string emoji = "\U0001F600";
		using var vm = CreateViewModel();
		var tail = new string('z', MainWindowViewModel.MaxConsoleCharacters - 1);

		// The text is 6 characters over the cap, so the cut falls between the two halves of the pair
		events.OnNext(new OutputEvent(SessionId, OutputStreamType.Stdout, "aaaaa" + emoji + tail));

		vm.ConsoleOutput.Should().Be(tail);
	}

	[Fact]
	public async Task FirstStepAfterLoad_HighlightsTheRegistersItChanged()
	{
		StubSuccessfulLoad(Program, RegisterState.Create(pc: 0x8000));
		api.StepAsync(SessionId, Arg.Any<CancellationToken>()).Returns(RegisterState.Create(r0: 5, pc: 0x8004));
		using var vm = CreateViewModel();
		vm.SourceCode = Program;
		await vm.LoadProgramCommand.Execute();

		await vm.StepCommand.Execute();

		vm.ChangedRegisters.Should().BeEquivalentTo(["R0", "PC"]);
	}
}
