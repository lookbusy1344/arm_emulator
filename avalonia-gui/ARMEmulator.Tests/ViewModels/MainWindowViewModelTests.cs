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
public class MainWindowViewModelTests
{
	private readonly IApiClient _mockApi;
	private readonly IWebSocketClient _mockWs;

	public MainWindowViewModelTests()
	{
		_mockApi = Substitute.For<IApiClient>();
		_mockWs = Substitute.For<IWebSocketClient>();
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
}
