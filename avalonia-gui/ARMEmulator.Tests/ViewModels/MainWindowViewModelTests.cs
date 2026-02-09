using System.Diagnostics.CodeAnalysis;
using System.Reactive.Subjects;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using FluentAssertions;
using NSubstitute;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

/// <summary>
/// Tests for MainWindowViewModel following TDD principles.
/// Tests define expected behavior before implementation.
/// </summary>
public class MainWindowViewModelTests : IDisposable
{
	private readonly IApiClient _mockApi;

	// NSubstitute mocks don't need disposal - they're lightweight test doubles
	[SuppressMessage("IDisposableAnalyzers.Correctness", "CA2213:Disposable fields should be disposed", Justification = "NSubstitute mock doesn't require disposal")]
	private readonly IWebSocketClient _mockWs;

	private readonly Subject<EmulatorEvent> _eventsSubject;

	public MainWindowViewModelTests()
	{
		_mockApi = Substitute.For<IApiClient>();
		_mockWs = Substitute.For<IWebSocketClient>();
		_eventsSubject = new Subject<EmulatorEvent>();
		_mockWs.Events.Returns(_eventsSubject);
	}

	public void Dispose()
	{
		_eventsSubject.Dispose();
		GC.SuppressFinalize(this);
	}

	[Fact]
	public void Constructor_InitializesWithDefaultState()
	{
		// Arrange & Act
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Assert
		viewModel.Status.Should().Be(VMState.Idle);
		viewModel.Registers.Should().NotBeNull();
		viewModel.Registers.R0.Should().Be(0);
		viewModel.PreviousRegisters.Should().BeNull();
		viewModel.ChangedRegisters.Should().BeEmpty();
		viewModel.ConsoleOutput.Should().BeEmpty();
		viewModel.ErrorMessage.Should().BeNull();
		viewModel.Breakpoints.Should().BeEmpty();
		viewModel.Watchpoints.Should().BeEmpty();
		viewModel.SourceCode.Should().BeEmpty();
		viewModel.MemoryData.Should().BeEmpty();
		viewModel.Disassembly.Should().BeEmpty();
		viewModel.IsConnected.Should().BeFalse();
		viewModel.SessionId.Should().BeNull();
	}

	[Fact]
	public void Constructor_InitializesCommands()
	{
		// Arrange & Act
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Assert
		viewModel.RunCommand.Should().NotBeNull();
		viewModel.PauseCommand.Should().NotBeNull();
		viewModel.StepCommand.Should().NotBeNull();
		viewModel.StepOverCommand.Should().NotBeNull();
		viewModel.StepOutCommand.Should().NotBeNull();
		viewModel.ResetCommand.Should().NotBeNull();
		viewModel.LoadProgramCommand.Should().NotBeNull();
		viewModel.ShowPcCommand.Should().NotBeNull();
	}

	[Fact]
	public void StatusColor_ReturnsGray_WhenDisconnected()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act & Assert
		viewModel.IsConnected = false;
		viewModel.StatusColor.Should().Be("Gray");
	}

	[Fact]
	public void StatusColor_ReturnsGreen_WhenIdleAndConnected()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		viewModel.IsConnected = true;
		viewModel.Status = VMState.Idle;

		// Assert
		viewModel.StatusColor.Should().Be("Green");
	}

	[Fact]
	public void StatusColor_ReturnsDodgerBlue_WhenRunning()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		viewModel.IsConnected = true;
		viewModel.Status = VMState.Running;

		// Assert
		viewModel.StatusColor.Should().Be("DodgerBlue");
	}

	[Fact]
	public void StatusColor_ReturnsOrange_WhenBreakpoint()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		viewModel.IsConnected = true;
		viewModel.Status = VMState.Breakpoint;

		// Assert
		viewModel.StatusColor.Should().Be("Orange");
	}

	[Fact]
	public void StatusColor_ReturnsPurple_WhenHalted()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		viewModel.IsConnected = true;
		viewModel.Status = VMState.Halted;

		// Assert
		viewModel.StatusColor.Should().Be("Purple");
	}

	[Fact]
	public void StatusColor_ReturnsRed_WhenError()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		viewModel.IsConnected = true;
		viewModel.Status = VMState.Error;

		// Assert
		viewModel.StatusColor.Should().Be("Red");
	}

	[Fact]
	public void StatusText_ReturnsDisconnected_WhenNotConnected()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act & Assert
		viewModel.IsConnected = false;
		viewModel.StatusText.Should().Be("Disconnected");
	}

	[Fact]
	public void StatusText_ReturnsIdle_WhenIdleAndConnected()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		viewModel.IsConnected = true;
		viewModel.Status = VMState.Idle;

		// Assert
		viewModel.StatusText.Should().Be("Idle");
	}

	[Fact]
	public void StatusText_ReturnsRunning_WhenRunning()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		viewModel.IsConnected = true;
		viewModel.Status = VMState.Running;

		// Assert
		viewModel.StatusText.Should().Be("Running");
	}

	[Fact]
	public void CanPause_ReturnsFalse_WhenStateIsIdle()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act & Assert
		viewModel.CanPause.Should().BeFalse();
	}

	[Fact]
	public void CanStep_ReturnsTrue_WhenStateIsIdle()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act & Assert
		viewModel.CanStep.Should().BeTrue();
	}

	[Fact]
	public void IsEditorEditable_ReturnsTrue_WhenStateIsIdle()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act & Assert
		viewModel.IsEditorEditable.Should().BeTrue();
	}

	[Fact]
	public void UpdateRegisters_TracksChangedRegisters()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		var initialRegisters = RegisterState.Create(r0: 0, r1: 100);
		var updatedRegisters = RegisterState.Create(r0: 42, r1: 100);

		// Act
		viewModel.UpdateRegisters(initialRegisters);  // Set initial state
		viewModel.UpdateRegisters(updatedRegisters);  // Update R0

		// Assert
		viewModel.Registers.Should().Be(updatedRegisters);
		viewModel.PreviousRegisters.Should().Be(initialRegisters);
		viewModel.ChangedRegisters.Should().Contain("R0");
		viewModel.ChangedRegisters.Should().NotContain("R1");  // R1 didn't change
	}

	[Fact]
	public void UpdateRegisters_DetectsMultipleChanges()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		var initialRegisters = RegisterState.Create(r0: 0, r1: 0, r2: 0);
		var updatedRegisters = RegisterState.Create(r0: 10, r1: 20, r2: 0);

		// Act
		viewModel.UpdateRegisters(initialRegisters);
		viewModel.UpdateRegisters(updatedRegisters);

		// Assert
		viewModel.ChangedRegisters.Should().Contain("R0");
		viewModel.ChangedRegisters.Should().Contain("R1");
		viewModel.ChangedRegisters.Should().NotContain("R2");
	}

	[Fact]
	public void UpdateRegisters_DetectsCPSRChange()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		var initialRegisters = RegisterState.Create(cpsr: new CPSRFlags(false, false, false, false));
		var updatedRegisters = RegisterState.Create(cpsr: new CPSRFlags(true, false, false, false));

		// Act
		viewModel.UpdateRegisters(initialRegisters);
		viewModel.UpdateRegisters(updatedRegisters);

		// Assert
		viewModel.ChangedRegisters.Should().Contain("CPSR");
	}

	[Fact]
	public void UpdateRegisters_FirstUpdate_DoesNotHighlightAnything()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		var registers = RegisterState.Create(r0: 42);

		// Act
		viewModel.UpdateRegisters(registers);

		// Assert - first update has no "previous" to compare against (PreviousRegisters was null)
		viewModel.ChangedRegisters.Should().BeEmpty();
		viewModel.PreviousRegisters.Should().NotBeNull();  // Now holds the default RegisterState
		viewModel.Registers.Should().Be(registers);
	}

	[Fact]
	public async Task UpdateRegisters_HighlightRemoval_RemovesAfter1500ms()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		var initialRegisters = RegisterState.Create(r0: 0);
		var updatedRegisters = RegisterState.Create(r0: 42);

		// Act
		viewModel.UpdateRegisters(initialRegisters);
		viewModel.UpdateRegisters(updatedRegisters);

		// Assert - highlight is immediately added
		viewModel.ChangedRegisters.Should().Contain("R0");

		// Wait for highlight to be removed (1.5s + buffer)
		await Task.Delay(1700);

		// Assert - highlight should be automatically removed
		viewModel.ChangedRegisters.Should().NotContain("R0");
	}

	[Fact]
	public async Task UpdateRegisters_MultipleChanges_EachHighlightTimedIndependently()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		viewModel.UpdateRegisters(RegisterState.Create(r0: 0, r1: 0));

		// Act - change R0 first
		viewModel.UpdateRegisters(RegisterState.Create(r0: 10, r1: 0));
		viewModel.ChangedRegisters.Should().Contain("R0");

		// Wait 800ms, then change R1
		await Task.Delay(800);
		viewModel.UpdateRegisters(RegisterState.Create(r0: 10, r1: 20));

		// Assert - both should be highlighted now
		viewModel.ChangedRegisters.Should().Contain("R0");
		viewModel.ChangedRegisters.Should().Contain("R1");

		// Wait 800ms more (R0 should expire at ~1600ms total, R1 at ~2400ms)
		await Task.Delay(800);

		// R0 should be gone, R1 should still be visible
		viewModel.ChangedRegisters.Should().NotContain("R0");
		viewModel.ChangedRegisters.Should().Contain("R1");

		// Wait another 800ms for R1 to expire (with buffer for scheduling overhead)
		await Task.Delay(800);
		viewModel.ChangedRegisters.Should().NotContain("R1");
	}

	[Fact]
	public void WebSocket_StateEvent_UpdatesRegistersAndStatus()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		var newRegisters = RegisterState.Create(r0: 42, r1: 100);
		var status = new VMStatus(VMState.Running, PC: 0x8000, Cycles: 100);

		// Act
		_eventsSubject.OnNext(new StateEvent("session1", status, newRegisters));

		// Assert
		viewModel.Registers.Should().Be(newRegisters);
		viewModel.Status.Should().Be(VMState.Running);
	}

	[Fact]
	public void WebSocket_OutputEvent_AppendsToConsole()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		_eventsSubject.OnNext(new OutputEvent("session1", OutputStreamType.Stdout, "Hello\n"));
		_eventsSubject.OnNext(new OutputEvent("session1", OutputStreamType.Stdout, "World\n"));

		// Assert
		viewModel.ConsoleOutput.Should().Be("Hello\nWorld\n");
	}

	[Fact]
	public void WebSocket_ExecutionEvent_BreakpointHit_UpdatesStatus()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		_eventsSubject.OnNext(new ExecutionEvent("session1", ExecutionEventType.BreakpointHit));

		// Assert
		viewModel.Status.Should().Be(VMState.Breakpoint);
	}

	[Fact]
	public void WebSocket_ExecutionEvent_Halted_UpdatesStatus()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		_eventsSubject.OnNext(new ExecutionEvent("session1", ExecutionEventType.Halted));

		// Assert
		viewModel.Status.Should().Be(VMState.Halted);
	}

	[Fact]
	public void WebSocket_ExecutionEvent_Error_UpdatesStatusAndMessage()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act
		_eventsSubject.OnNext(new ExecutionEvent("session1", ExecutionEventType.Error, Message: "Test error"));

		// Assert
		viewModel.Status.Should().Be(VMState.Error);
		viewModel.ErrorMessage.Should().Be("Test error");
	}

	[Fact]
	public void WebSocket_StateEventWhenHalted_IsIgnored()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		viewModel.Status = VMState.Halted;
		var originalRegisters = viewModel.Registers;

		// Act - send state event while halted
		var newRegisters = RegisterState.Create(r0: 999);
		var status = new VMStatus(VMState.Running, PC: 0x8000, Cycles: 100);
		_eventsSubject.OnNext(new StateEvent("session1", status, newRegisters));

		// Assert - state should not change when already halted
		viewModel.Status.Should().Be(VMState.Halted);
		viewModel.Registers.Should().Be(originalRegisters);
	}

	[Fact]
	public async Task CreateSession_SetsSessionIdAndConnects()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		var sessionInfo = new SessionInfo("test-session-123");
		_mockApi.CreateSessionAsync(Arg.Any<CancellationToken>()).Returns(sessionInfo);

		// Act
		await viewModel.CreateSessionAsync();

		// Assert
		viewModel.SessionId.Should().Be("test-session-123");
		viewModel.IsConnected.Should().BeTrue();
		await _mockWs.Received(1).ConnectAsync("test-session-123", Arg.Any<CancellationToken>());
	}

	[Fact]
	public async Task DestroySession_ClearsSessionIdAndDisconnects()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		var sessionInfo = new SessionInfo("test-session-123");
		_mockApi.CreateSessionAsync(Arg.Any<CancellationToken>()).Returns(sessionInfo);
		await viewModel.CreateSessionAsync();

		// Act
		await viewModel.DestroySessionAsync();

		// Assert
		viewModel.SessionId.Should().BeNull();
		viewModel.IsConnected.Should().BeFalse();
		await _mockWs.Received(1).DisconnectAsync();
		await _mockApi.Received(1).DestroySessionAsync("test-session-123", Arg.Any<CancellationToken>());
	}

	[Fact]
	public async Task CreateSession_WhenAlreadyConnected_DestroysOldSessionFirst()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		var session1 = new SessionInfo("session-1");
		var session2 = new SessionInfo("session-2");
		_mockApi.CreateSessionAsync(Arg.Any<CancellationToken>()).Returns(session1, session2);

		// Act
		await viewModel.CreateSessionAsync();  // Create first session
		await viewModel.CreateSessionAsync();  // Create second session

		// Assert - old session should be destroyed first
		await _mockApi.Received(1).DestroySessionAsync("session-1", Arg.Any<CancellationToken>());
		viewModel.SessionId.Should().Be("session-2");
	}

	[Fact]
	public async Task RunCommand_CallsApiAndUpdatesState()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		viewModel.SessionId = "test-session";

		// Act
		await viewModel.RunCommand.Execute();

		// Assert
		await _mockApi.Received(1).RunAsync("test-session", Arg.Any<CancellationToken>());
	}

	[Fact]
	public async Task PauseCommand_CallsApiStop()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		viewModel.SessionId = "test-session";
		viewModel.Status = VMState.Running;

		// Act
		await viewModel.PauseCommand.Execute();

		// Assert
		await _mockApi.Received(1).StopAsync("test-session", Arg.Any<CancellationToken>());
	}

	[Fact]
	public async Task StepCommand_CallsApiAndUpdatesRegisters()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		viewModel.SessionId = "test-session";
		var newRegisters = RegisterState.Create(r0: 42);
		_mockApi.StepAsync("test-session", Arg.Any<CancellationToken>()).Returns(newRegisters);

		// Act
		await viewModel.StepCommand.Execute();

		// Assert
		await _mockApi.Received(1).StepAsync("test-session", Arg.Any<CancellationToken>());
		viewModel.Registers.Should().Be(newRegisters);
	}

	[Fact]
	public async Task StepOverCommand_CallsApiAndUpdatesRegisters()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		viewModel.SessionId = "test-session";
		var newRegisters = RegisterState.Create(r0: 99);
		_mockApi.StepOverAsync("test-session", Arg.Any<CancellationToken>()).Returns(newRegisters);

		// Act
		await viewModel.StepOverCommand.Execute();

		// Assert
		await _mockApi.Received(1).StepOverAsync("test-session", Arg.Any<CancellationToken>());
		viewModel.Registers.Should().Be(newRegisters);
	}

	[Fact]
	public async Task StepOutCommand_CallsApiAndUpdatesRegisters()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		viewModel.SessionId = "test-session";
		var newRegisters = RegisterState.Create(r0: 123);
		_mockApi.StepOutAsync("test-session", Arg.Any<CancellationToken>()).Returns(newRegisters);

		// Act
		await viewModel.StepOutCommand.Execute();

		// Assert
		await _mockApi.Received(1).StepOutAsync("test-session", Arg.Any<CancellationToken>());
		viewModel.Registers.Should().Be(newRegisters);
	}

	[Fact]
	public async Task ResetCommand_CallsApi()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);
		viewModel.SessionId = "test-session";

		// Act
		await viewModel.ResetCommand.Execute();

		// Assert
		await _mockApi.Received(1).ResetAsync("test-session", Arg.Any<CancellationToken>());
	}

	[Fact]
	public async Task ShowPCCommand_CanExecuteWithoutSession()
	{
		// Arrange
		using var viewModel = new MainWindowViewModel(_mockApi, _mockWs);

		// Act & Assert - should not throw
		await viewModel.ShowPcCommand.Execute();
	}
}
