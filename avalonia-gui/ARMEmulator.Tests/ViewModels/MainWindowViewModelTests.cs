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
}
