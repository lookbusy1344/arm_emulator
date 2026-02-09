using System.Collections.Immutable;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using FluentAssertions;
using NSubstitute;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

public class DisassemblyViewModelTests : IDisposable
{
	private readonly IApiClient apiClient;
	private readonly DisassemblyViewModel viewModel;
	private readonly string sessionId = "test-session";

	public DisassemblyViewModelTests()
	{
		apiClient = Substitute.For<IApiClient>();
		viewModel = new DisassemblyViewModel(apiClient) { SessionId = sessionId };
	}

	[Fact]
	public void InitialState_ShouldBeDefaultValues()
	{
		viewModel.ProgramCounter.Should().Be(0);
		viewModel.Instructions.Should().BeEmpty();
		viewModel.Breakpoints.Should().BeEmpty();
	}

	[Fact]
	public async Task RefreshDisassemblyAsync_ShouldLoadInstructionsAroundPC()
	{
		// Arrange
		var registers = RegisterState.Create(pc: 0x8000);
		var instructions = ImmutableArray.Create(
			new DisassemblyInstruction(Address: 0x7FF0, MachineCode: 0xE3A00001, Mnemonic: "MOV R0, #1", Symbol: null),
			new DisassemblyInstruction(Address: 0x8000, MachineCode: 0xE2811001, Mnemonic: "ADD R1, R1, #1", Symbol: null)
		);
		apiClient.GetDisassemblyAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(instructions);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshDisassemblyAsync();

		// Assert
		viewModel.Instructions.Should().HaveCount(2);
		viewModel.ProgramCounter.Should().Be(0x8000);
	}

	[Fact]
	public async Task RefreshDisassemblyAsync_ShouldLoadWindowAroundPC()
	{
		// Arrange
		var registers = RegisterState.Create(pc: 0x8000);
		uint capturedAddress = 0;
		int capturedCount = 0;

		apiClient.GetDisassemblyAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(callInfo => {
				capturedAddress = callInfo.ArgAt<uint>(1);
				capturedCount = callInfo.ArgAt<int>(2);
				return ImmutableArray<DisassemblyInstruction>.Empty;
			});

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshDisassemblyAsync();

		// Assert - Should load window centered around PC (±32 instructions = 64 total)
		capturedCount.Should().Be(64);
		capturedAddress.Should().BeLessThan(0x8000); // Should start before PC
	}

	[Fact]
	public void UpdateRegisters_ShouldUpdateProgramCounter()
	{
		// Arrange
		var registers = RegisterState.Create(pc: 0x8100);

		// Act
		viewModel.UpdateRegisters(registers);

		// Assert
		viewModel.ProgramCounter.Should().Be(0x8100);
	}

	[Fact]
	public void UpdateBreakpoints_ShouldSetBreakpoints()
	{
		// Arrange
		var breakpoints = ImmutableHashSet.Create<uint>(0x8000, 0x8004, 0x8008);

		// Act
		viewModel.UpdateBreakpoints(breakpoints);

		// Assert
		viewModel.Breakpoints.Should().BeEquivalentTo(breakpoints);
	}

	[Fact]
	public async Task FormatInstructions_ShouldIncludePCIndicator()
	{
		// Arrange
		var registers = RegisterState.Create(pc: 0x8000);
		var instructions = ImmutableArray.Create(
			new DisassemblyInstruction(Address: 0x7FFC, MachineCode: 0xE3A00001, Mnemonic: "MOV R0, #1", Symbol: null),
			new DisassemblyInstruction(Address: 0x8000, MachineCode: 0xE2811001, Mnemonic: "ADD R1, R1, #1", Symbol: null),
			new DisassemblyInstruction(Address: 0x8004, MachineCode: 0xE1510002, Mnemonic: "CMP R1, R2", Symbol: null)
		);
		apiClient.GetDisassemblyAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(instructions);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshDisassemblyAsync();

		// Assert
		var formattedInstructions = viewModel.FormattedInstructions;
		formattedInstructions.Should().Contain(i => i.IsCurrentPC && i.Address == "00008000");
	}

	[Fact]
	public async Task FormatInstructions_ShouldIncludeBreakpointIndicator()
	{
		// Arrange
		var registers = RegisterState.Create(pc: 0x8000);
		var breakpoints = ImmutableHashSet.Create<uint>(0x8000);
		var instructions = ImmutableArray.Create(
			new DisassemblyInstruction(Address: 0x8000, MachineCode: 0xE3A00001, Mnemonic: "MOV R0, #1", Symbol: null)
		);
		apiClient.GetDisassemblyAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(instructions);

		viewModel.UpdateRegisters(registers);
		viewModel.UpdateBreakpoints(breakpoints);

		// Act
		await viewModel.RefreshDisassemblyAsync();

		// Assert
		var formattedInstructions = viewModel.FormattedInstructions;
		formattedInstructions.Should().Contain(i => i.HasBreakpoint && i.Address == "00008000");
	}

	[Fact]
	public async Task RefreshCommand_ShouldReloadDisassembly()
	{
		// Arrange
		var registers = RegisterState.Create(pc: 0x8000);
		apiClient.GetDisassemblyAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns([]);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshCommand.Execute();

		// Assert
		await apiClient.Received(1).GetDisassemblyAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>());
	}

	[Fact]
	public async Task FormatInstructions_ShouldIncludeSymbols()
	{
		// Arrange
		var registers = RegisterState.Create(pc: 0x8000);
		var instructions = ImmutableArray.Create(
			new DisassemblyInstruction(Address: 0x8000, MachineCode: 0xE3A00001, Mnemonic: "MOV R0, #1", Symbol: "main")
		);
		apiClient.GetDisassemblyAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(instructions);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshDisassemblyAsync();

		// Assert
		var formattedInstructions = viewModel.FormattedInstructions;
		formattedInstructions.Should().Contain(i => i.Symbol == "main");
	}

	public void Dispose()
	{
		viewModel.Dispose();
		GC.SuppressFinalize(this);
	}
}
