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
/// Tests for assembling the editor source on demand: before Run, Step, Reset and gutter breakpoint toggles.
/// </summary>
public sealed class MainWindowViewModelAssembleTests : IDisposable
{
	private const string SessionId = "s1";

	// Instructions on lines 2, 3 and 4
	private const string Program = "_start:\n  MOV R0, #1\n  MOV R1, #2\n  SWI #0\n";
	private const string Edited = "_start:\n  MOV R0, #1\n  MOV R1, #3\n  SWI #0\n";

	private const uint Line2Address = 0x8000;
	private const uint Line3Address = 0x8004;
	private const uint Line4Address = 0x8008;
	private const uint MovedLine4Address = 0x800C;

	private readonly IApiClient api = Substitute.For<IApiClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient ws = Substitute.For<IWebSocketClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IFileService files = Substitute.For<IFileService>();

	private readonly Subject<EmulatorEvent> events = new();

	public MainWindowViewModelAssembleTests()
	{
		ws.Events.Returns(events);
		files.RecentFilesChanged.Returns(Observable.Never<Unit>());
		files.RecentFiles.Returns([]);
		api.GetMemoryAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<byte>.Empty);
		api.GetDisassemblyAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<DisassemblyInstruction>.Empty);
		api.GetRegistersAsync(SessionId, Arg.Any<CancellationToken>()).Returns(RegisterState.Create(pc: Line2Address));
		api.StepAsync(SessionId, Arg.Any<CancellationToken>()).Returns(RegisterState.Create(pc: Line3Address));
		api.StepOverAsync(SessionId, Arg.Any<CancellationToken>()).Returns(RegisterState.Create(pc: Line3Address));
		api.StepOutAsync(SessionId, Arg.Any<CancellationToken>()).Returns(RegisterState.Create(pc: Line3Address));
		api.LoadProgramAsync(SessionId, Arg.Any<string>(), Arg.Any<CancellationToken>())
			.Returns(new LoadProgramResponse(EquatableDictionaryFactory.CopyOf(new Dictionary<string, uint>())));
		StubSourceMap(Line4Address);
	}

	public void Dispose() => events.Dispose();

	private static CancellationToken Ct => TestContext.Current.CancellationToken;

	private MainWindowViewModel CreateViewModel() => new(api, ws, files) { SessionId = SessionId };

	private void StubSourceMap(uint line4Address) =>
		api.GetSourceMapAsync(SessionId, Arg.Any<CancellationToken>())
			.Returns([new SourceMapEntry(Line2Address, 2, ""), new SourceMapEntry(Line3Address, 3, ""), new SourceMapEntry(line4Address, 4, "")]);

	private void StubAssemblerError(string source) =>
		api.LoadProgramAsync(SessionId, source, Arg.Any<CancellationToken>())
			.ThrowsAsync(new ProgramLoadException(["line 3: bad"]));

	/// <summary>A view model with <see cref="Program"/> assembled and the API call log cleared.</summary>
	private async Task<MainWindowViewModel> CreateAssembledViewModelAsync()
	{
		var vm = CreateViewModel();
		vm.SourceCode = Program;
		await vm.AssembleCommand.Execute();
		api.ClearReceivedCalls();
		return vm;
	}

	// Needs assembly

	[Fact]
	public void NeedsAssembly_EmptyEditor_IsFalse()
	{
		using var vm = CreateViewModel();

		vm.NeedsAssembly.Should().BeFalse();
	}

	[Fact]
	public async Task NeedsAssembly_AfterAnEdit_IsTrue()
	{
		using var vm = await CreateAssembledViewModelAsync();

		vm.SourceCode = Edited;

		vm.NeedsAssembly.Should().BeTrue();
	}

	[Fact]
	public async Task NeedsAssembly_AfterEditingBackToTheAssembledSource_IsFalse()
	{
		using var vm = await CreateAssembledViewModelAsync();

		vm.SourceCode = Edited;
		vm.SourceCode = Program;

		vm.NeedsAssembly.Should().BeFalse();
	}

	[Fact]
	public async Task NeedsAssembly_AfterFailedAssembly_IsTrue()
	{
		StubAssemblerError(Program);
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.AssembleCommand.Execute();

		vm.AssembledSource.Should().BeNull();
		vm.NeedsAssembly.Should().BeTrue();
	}

	// Run and step

	[Fact]
	public async Task Run_WithEditedSource_AssemblesThenRuns()
	{
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.RunCommand.Execute();

		Received.InOrder(() => {
			_ = api.LoadProgramAsync(SessionId, Program, Arg.Any<CancellationToken>());
			_ = api.RunAsync(SessionId, Arg.Any<CancellationToken>());
		});
		vm.AssembledSource.Should().Be(Program);
	}

	[Fact]
	public async Task Run_WithUnchangedSource_DoesNotReassemble()
	{
		using var vm = await CreateAssembledViewModelAsync();

		await vm.RunCommand.Execute();

		_ = api.DidNotReceiveWithAnyArgs().LoadProgramAsync(default!, default!, Ct);
		_ = api.Received(1).RunAsync(SessionId, Arg.Any<CancellationToken>());
	}

	[Fact]
	public async Task Run_WithEmptyEditor_DoesNotAssemble()
	{
		using var vm = CreateViewModel();

		await vm.RunCommand.Execute();

		_ = api.DidNotReceiveWithAnyArgs().LoadProgramAsync(default!, default!, Ct);
		_ = api.Received(1).RunAsync(SessionId, Arg.Any<CancellationToken>());
	}

	[Fact]
	public async Task Run_WhenAssemblyFails_ReportsErrorsAndDoesNotRun()
	{
		StubAssemblerError(Program);
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.RunCommand.Execute();

		_ = api.DidNotReceiveWithAnyArgs().RunAsync(default!, Ct);
		vm.ErrorMessage.Should().Be("Failed to assemble program:\nline 3: bad");
	}

	[Fact]
	public async Task Run_WithEditedSourceAndNoSession_ReportsErrorAndDoesNotRun()
	{
		using var vm = new MainWindowViewModel(api, ws, files);
		vm.SourceCode = Program;

		await vm.RunCommand.Execute();

		_ = api.DidNotReceiveWithAnyArgs().LoadProgramAsync(default!, default!, Ct);
		_ = api.DidNotReceiveWithAnyArgs().RunAsync(default!, Ct);
		vm.ErrorMessage.Should().Be("No active session");
	}

	[Fact]
	public async Task Step_WithEditedSource_AssemblesThenSteps()
	{
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.StepCommand.Execute();

		Received.InOrder(() => {
			_ = api.LoadProgramAsync(SessionId, Program, Arg.Any<CancellationToken>());
			_ = api.StepAsync(SessionId, Arg.Any<CancellationToken>());
		});
	}

	[Fact]
	public async Task StepOver_WithEditedSource_AssemblesThenSteps()
	{
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.StepOverCommand.Execute();

		Received.InOrder(() => {
			_ = api.LoadProgramAsync(SessionId, Program, Arg.Any<CancellationToken>());
			_ = api.StepOverAsync(SessionId, Arg.Any<CancellationToken>());
		});
	}

	[Fact]
	public async Task StepOut_WithEditedSource_AssemblesThenSteps()
	{
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.StepOutCommand.Execute();

		Received.InOrder(() => {
			_ = api.LoadProgramAsync(SessionId, Program, Arg.Any<CancellationToken>());
			_ = api.StepOutAsync(SessionId, Arg.Any<CancellationToken>());
		});
	}

	[Fact]
	public async Task Step_WhenAssemblyFails_DoesNotStep()
	{
		StubAssemblerError(Program);
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.StepCommand.Execute();

		_ = api.DidNotReceiveWithAnyArgs().StepAsync(default!, Ct);
	}

	[Theory]
	[InlineData(VMState.Halted)]
	[InlineData(VMState.Error)]
	public async Task CanStep_WhenStoppedWithEditedSource_IsTrue(VMState state)
	{
		using var vm = await CreateAssembledViewModelAsync();
		vm.Status = state;

		vm.SourceCode = Edited;

		vm.CanStep.Should().BeTrue();
		(await vm.StepCommand.CanExecute.FirstAsync()).Should().BeTrue();
	}

	[Fact]
	public async Task CanStep_WhenHaltedWithUnchangedSource_IsFalse()
	{
		using var vm = await CreateAssembledViewModelAsync();

		vm.Status = VMState.Halted;

		vm.CanStep.Should().BeFalse();
		(await vm.StepCommand.CanExecute.FirstAsync()).Should().BeFalse();
	}

	[Fact]
	public async Task CanStep_WhenRunningWithEditedSource_IsFalse()
	{
		using var vm = await CreateAssembledViewModelAsync();
		vm.SourceCode = Edited;

		vm.Status = VMState.Running;

		vm.CanStep.Should().BeFalse();
	}

	// Reset

	[Fact]
	public async Task Reset_WithEditedSource_AssemblesInsteadOfRestarting()
	{
		using var vm = await CreateAssembledViewModelAsync();
		vm.SourceCode = Edited;

		await vm.ResetCommand.Execute();

		_ = api.Received(1).LoadProgramAsync(SessionId, Edited, Arg.Any<CancellationToken>());
		_ = api.DidNotReceiveWithAnyArgs().RestartAsync(default!, Ct);
	}

	[Fact]
	public async Task Reset_WithUnchangedSource_Restarts()
	{
		using var vm = await CreateAssembledViewModelAsync();

		await vm.ResetCommand.Execute();

		_ = api.Received(1).RestartAsync(SessionId, Arg.Any<CancellationToken>());
		_ = api.DidNotReceiveWithAnyArgs().LoadProgramAsync(default!, default!, Ct);
	}

	// Assemble command

	[Fact]
	public async Task Assemble_WithUnchangedSource_AssemblesAgain()
	{
		using var vm = await CreateAssembledViewModelAsync();

		await vm.AssembleCommand.Execute();

		_ = api.Received(1).LoadProgramAsync(SessionId, Program, Arg.Any<CancellationToken>());
	}

	[Theory]
	[InlineData(VMState.Running, false)]
	[InlineData(VMState.Breakpoint, false)]
	[InlineData(VMState.WaitingForInput, false)]
	[InlineData(VMState.Idle, true)]
	[InlineData(VMState.Halted, true)]
	public async Task Assemble_IsAvailableOnlyWhileTheEditorIsEditable(VMState state, bool expected)
	{
		using var vm = CreateViewModel();

		vm.Status = state;

		(await vm.AssembleCommand.CanExecute.FirstAsync()).Should().Be(expected);
	}

	// Breakpoints

	[Fact]
	public async Task Reassembly_MovesBreakpointsToTheNewAddressOfTheirLine()
	{
		using var vm = await CreateAssembledViewModelAsync();
		await vm.ToggleBreakpointCommand.Execute(4);
		StubSourceMap(MovedLine4Address);
		vm.SourceCode = Edited;

		await vm.RunCommand.Execute();

		Received.InOrder(() => {
			_ = api.AddBreakpointAsync(SessionId, Line4Address, Arg.Any<CancellationToken>());
			_ = api.LoadProgramAsync(SessionId, Edited, Arg.Any<CancellationToken>());
			_ = api.RemoveBreakpointAsync(SessionId, Line4Address, Arg.Any<CancellationToken>());
			_ = api.AddBreakpointAsync(SessionId, MovedLine4Address, Arg.Any<CancellationToken>());
			_ = api.RunAsync(SessionId, Arg.Any<CancellationToken>());
		});
		vm.Breakpoints.Should().BeEquivalentTo([MovedLine4Address]);
	}

	[Fact]
	public async Task Reassembly_DropsBreakpointsOnLinesWithoutAnInstruction()
	{
		using var vm = await CreateAssembledViewModelAsync();
		await vm.ToggleBreakpointCommand.Execute(4);
		api.GetSourceMapAsync(SessionId, Arg.Any<CancellationToken>())
			.Returns([new SourceMapEntry(Line2Address, 2, ""), new SourceMapEntry(Line3Address, 3, "")]);
		vm.SourceCode = Edited;
		api.ClearReceivedCalls();

		await vm.RunCommand.Execute();

		_ = api.Received(1).RemoveBreakpointAsync(SessionId, Line4Address, Arg.Any<CancellationToken>());
		_ = api.DidNotReceiveWithAnyArgs().AddBreakpointAsync(default!, default, Ct);
		vm.Breakpoints.Should().BeEmpty();
	}

	[Fact]
	public async Task FailedReassembly_KeepsBreakpoints()
	{
		using var vm = await CreateAssembledViewModelAsync();
		await vm.ToggleBreakpointCommand.Execute(4);
		StubAssemblerError(Edited);
		vm.SourceCode = Edited;

		await vm.RunCommand.Execute();

		_ = api.DidNotReceiveWithAnyArgs().RemoveBreakpointAsync(default!, default, Ct);
		vm.Breakpoints.Should().BeEquivalentTo([Line4Address]);
	}

	[Fact]
	public async Task ReassemblyAfterAFailure_RestoresBreakpointsOnTheirLines()
	{
		const string broken = "_start:\n  MOV R0, #1\n  FOO\n  SWI #0\n";
		using var vm = await CreateAssembledViewModelAsync();
		await vm.ToggleBreakpointCommand.Execute(4);
		StubAssemblerError(broken);
		vm.SourceCode = broken;
		await vm.RunCommand.Execute();
		StubSourceMap(MovedLine4Address);
		vm.SourceCode = Edited;

		await vm.RunCommand.Execute();

		_ = api.Received(1).AddBreakpointAsync(SessionId, MovedLine4Address, Arg.Any<CancellationToken>());
		vm.Breakpoints.Should().BeEquivalentTo([MovedLine4Address]);
	}

	[Fact]
	public async Task ToggleBreakpointAtLine_WithEditedSource_AssemblesFirst()
	{
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.ToggleBreakpointCommand.Execute(3);

		Received.InOrder(() => {
			_ = api.LoadProgramAsync(SessionId, Program, Arg.Any<CancellationToken>());
			_ = api.AddBreakpointAsync(SessionId, Line3Address, Arg.Any<CancellationToken>());
		});
		vm.Breakpoints.Should().BeEquivalentTo([Line3Address]);
	}

	[Fact]
	public async Task ToggleBreakpointAtLine_WhenAssemblyFails_AddsNothing()
	{
		StubAssemblerError(Program);
		using var vm = CreateViewModel();
		vm.SourceCode = Program;

		await vm.ToggleBreakpointCommand.Execute(3);

		_ = api.DidNotReceiveWithAnyArgs().AddBreakpointAsync(default!, default, Ct);
		vm.ErrorMessage.Should().Be("Failed to assemble program:\nline 3: bad");
	}
}
