using System.Diagnostics.CodeAnalysis;
using System.Net;
using System.Reactive;
using System.Reactive.Linq;
using System.Reactive.Subjects;
using System.Windows.Input;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using AwesomeAssertions;
using NSubstitute;
using NSubstitute.ExceptionExtensions;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// Tests that command failures surface in <see cref="MainWindowViewModel.ErrorMessage"/> rather than ending the process.
/// Commands run through <see cref="ICommand.Execute"/>, the path the UI bindings use.
/// </summary>
public sealed class MainWindowViewModelErrorTests : IDisposable
{
	private const string SessionId = "s1";

	private readonly IApiClient api = Substitute.For<IApiClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient ws = Substitute.For<IWebSocketClient>();

	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IFileService files = Substitute.For<IFileService>();

	private readonly Subject<EmulatorEvent> events = new();

	public MainWindowViewModelErrorTests()
	{
		ws.Events.Returns(events);
		files.RecentFilesChanged.Returns(Observable.Never<Unit>());
		files.RecentFiles.Returns([]);
		api.GetMemoryAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<byte>.Empty);
		api.GetDisassemblyAsync(default!, default, default, default).ReturnsForAnyArgs(ImmutableArray<DisassemblyInstruction>.Empty);
	}

	public void Dispose() => events.Dispose();

	private MainWindowViewModel CreateViewModel() => new(api, ws, files) { SessionId = SessionId };

	private static void Execute(ICommand command, object? parameter = null) => command.Execute(parameter);

	// Error bar

	[Fact]
	public void DismissErrorCommand_ClearsErrorMessage()
	{
		using var vm = CreateViewModel();
		vm.ErrorMessage = "Failed to load program: boom";

		Execute(vm.DismissErrorCommand);

		vm.ErrorMessage.Should().BeNull();
	}

	// Execution command failures

	[Theory]
	[InlineData("Run")]
	[InlineData("Pause")]
	[InlineData("Step")]
	[InlineData("Step over")]
	[InlineData("Step out")]
	public void ExecutionCommand_WhenApiFails_ReportsOperationAndMessage(string operation)
	{
		var failure = new ApiException("API error: boom", HttpStatusCode.InternalServerError);
		api.RunAsync(SessionId, Arg.Any<CancellationToken>()).ThrowsAsync(failure);
		api.StopAsync(SessionId, Arg.Any<CancellationToken>()).ThrowsAsync(failure);
		api.StepAsync(SessionId, Arg.Any<CancellationToken>()).ThrowsAsync(failure);
		api.StepOverAsync(SessionId, Arg.Any<CancellationToken>()).ThrowsAsync(failure);
		api.StepOutAsync(SessionId, Arg.Any<CancellationToken>()).ThrowsAsync(failure);
		using var vm = CreateViewModel();
		vm.Status = operation == "Pause" ? VMState.Running : VMState.Breakpoint;

		Execute(operation switch {
			"Run" => vm.RunCommand,
			"Pause" => vm.PauseCommand,
			"Step" => vm.StepCommand,
			"Step over" => vm.StepOverCommand,
			"Step out" => vm.StepOutCommand,
			_ => throw new ArgumentOutOfRangeException(nameof(operation))
		});

		vm.ErrorMessage.Should().Be($"{operation} failed: API error: boom");
	}

	[Fact]
	public void Step_WhenBackendConnectionIsLost_ReportsTransportMessageAndKeepsRegisters()
	{
		api.StepAsync(SessionId, Arg.Any<CancellationToken>()).ThrowsAsync(new HttpRequestException("Connection refused"));
		using var vm = CreateViewModel();
		var registers = RegisterState.Create(r0: 7, pc: 0x8004);
		vm.UpdateRegisters(registers);

		Execute(vm.StepCommand);

		vm.ErrorMessage.Should().Be("Step failed: Connection refused");
		vm.Registers.Should().Be(registers);
	}

	[Fact]
	public void Step_WhenSessionExpired_ReportsSessionMessage()
	{
		api.StepAsync(SessionId, Arg.Any<CancellationToken>()).ThrowsAsync(new SessionNotFoundException(SessionId));
		using var vm = CreateViewModel();

		Execute(vm.StepCommand);

		vm.ErrorMessage.Should().Be("Step failed: Session 's1' not found or expired");
	}

	[Fact]
	public void Step_WhenCancelled_ReportsNothing()
	{
		api.StepAsync(SessionId, Arg.Any<CancellationToken>()).ThrowsAsync(new OperationCanceledException());
		using var vm = CreateViewModel();

		Execute(vm.StepCommand);

		vm.ErrorMessage.Should().BeNull();
	}

	[Fact]
	public void CommandFailure_LeavesCommandUsable()
	{
		api.StepAsync(SessionId, Arg.Any<CancellationToken>())
			.Returns(
				_ => throw new ApiException("API error: first"),
				_ => RegisterState.Create(pc: 0x8008));
		using var vm = CreateViewModel();

		Execute(vm.StepCommand);
		Execute(vm.StepCommand);

		vm.Registers.PC.Should().Be(0x8008u);
	}

	// Breakpoint toggling by source line

	[Fact]
	public void ToggleBreakpoint_OnLineWithoutBreakpoint_AddsBreakpointAtMappedAddress()
	{
		using var vm = CreateViewModel();
		vm.LineToAddress = ImmutableDictionary<int, uint>.Empty.Add(3, 0x8000);

		Execute(vm.ToggleBreakpointCommand, 3);

		_ = api.Received(1).AddBreakpointAsync(SessionId, 0x8000, Arg.Any<CancellationToken>());
		vm.Breakpoints.Should().BeEquivalentTo([0x8000u]);
	}

	[Fact]
	public void ToggleBreakpoint_OnLineWithBreakpoint_RemovesIt()
	{
		using var vm = CreateViewModel();
		vm.LineToAddress = ImmutableDictionary<int, uint>.Empty.Add(3, 0x8000);
		vm.Breakpoints = [0x8000];

		Execute(vm.ToggleBreakpointCommand, 3);

		_ = api.Received(1).RemoveBreakpointAsync(SessionId, 0x8000, Arg.Any<CancellationToken>());
		vm.Breakpoints.Should().BeEmpty();
	}

	[Fact]
	public void ToggleBreakpoint_OnLineWithoutInstruction_ReportsLineAndCallsNoApi()
	{
		using var vm = CreateViewModel();
		vm.LineToAddress = ImmutableDictionary<int, uint>.Empty.Add(3, 0x8000);

		Execute(vm.ToggleBreakpointCommand, 7);

		vm.ErrorMessage.Should().Be("Line 7 has no instruction for a breakpoint");
		_ = api.DidNotReceiveWithAnyArgs().AddBreakpointAsync(default!, default, TestContext.Current.CancellationToken);
	}

	[Fact]
	public void ToggleBreakpoint_WhenApiFails_ReportsErrorAndKeepsBreakpoints()
	{
		api.AddBreakpointAsync(SessionId, 0x8000, Arg.Any<CancellationToken>())
			.ThrowsAsync(new ApiException("API error: Failed to add breakpoint: invalid breakpoint address"));
		using var vm = CreateViewModel();
		vm.LineToAddress = ImmutableDictionary<int, uint>.Empty.Add(3, 0x8000);

		Execute(vm.ToggleBreakpointCommand, 3);

		vm.ErrorMessage.Should().Be("Toggle breakpoint failed: API error: Failed to add breakpoint: invalid breakpoint address");
		vm.Breakpoints.Should().BeEmpty();
	}

	// Breakpoint toggling by address (disassembly marker column)

	[Fact]
	public void ToggleBreakpointAtCaret_UsesTheEditorCaretLine()
	{
		using var vm = CreateViewModel();
		vm.LineToAddress = ImmutableDictionary<int, uint>.Empty.Add(3, 0x8000).Add(5, 0x8004);
		vm.CaretLine = 5;

		Execute(vm.ToggleBreakpointAtCaretCommand);

		_ = api.Received(1).AddBreakpointAsync(SessionId, 0x8004, Arg.Any<CancellationToken>());
		vm.Breakpoints.Should().BeEquivalentTo([0x8004u]);
	}

	[Fact]
	public void ToggleBreakpointAtCaret_OnLineWithoutInstruction_ReportsLine()
	{
		using var vm = CreateViewModel();
		vm.CaretLine = 7;

		Execute(vm.ToggleBreakpointAtCaretCommand);

		vm.ErrorMessage.Should().Be("Line 7 has no instruction for a breakpoint");
	}

	[Fact]
	public void ToggleBreakpointAtAddress_WithoutBreakpoint_AddsIt()
	{
		using var vm = CreateViewModel();

		Execute(vm.ToggleBreakpointAtAddressCommand, 0x8004u);

		_ = api.Received(1).AddBreakpointAsync(SessionId, 0x8004, Arg.Any<CancellationToken>());
		vm.Breakpoints.Should().BeEquivalentTo([0x8004u]);
	}

	[Fact]
	public void ToggleBreakpointAtAddress_WithBreakpoint_RemovesIt()
	{
		using var vm = CreateViewModel();
		vm.Breakpoints = [0x8004];

		Execute(vm.ToggleBreakpointAtAddressCommand, 0x8004u);

		_ = api.Received(1).RemoveBreakpointAsync(SessionId, 0x8004, Arg.Any<CancellationToken>());
		vm.Breakpoints.Should().BeEmpty();
	}

	[Fact]
	public void ToggleBreakpointAtAddress_WhenApiFails_ReportsErrorAndKeepsBreakpoints()
	{
		api.AddBreakpointAsync(SessionId, 0x8004, Arg.Any<CancellationToken>())
			.ThrowsAsync(new ApiException("API error: invalid breakpoint address"));
		using var vm = CreateViewModel();

		Execute(vm.ToggleBreakpointAtAddressCommand, 0x8004u);

		vm.ErrorMessage.Should().Be("Toggle breakpoint failed: API error: invalid breakpoint address");
		vm.Breakpoints.Should().BeEmpty();
	}
}
