using System.Collections.Immutable;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using AwesomeAssertions;
using NSubstitute;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

public class StackViewModelTests : IDisposable
{
	private readonly IApiClient apiClient;
	private readonly StackViewModel viewModel;
	private readonly string sessionId = "test-session";

	public StackViewModelTests()
	{
		apiClient = Substitute.For<IApiClient>();
		viewModel = new StackViewModel(apiClient) { SessionId = sessionId };
	}

	[Fact]
	public void InitialState_ShouldBeDefaultValues()
	{
		viewModel.StackPointer.Should().Be(0);
		viewModel.StackEntries.Should().BeEmpty();
		viewModel.StackSize.Should().Be(0);
	}

	[Fact]
	public async Task RefreshStackAsync_ShouldLoadStackMemory()
	{
		// Arrange
		var registers = RegisterState.Create(sp: 0x4FFF0, lr: 0x8100, pc: 0x8000);
		var memoryData = ImmutableArray.Create<byte>(
			0x01, 0x02, 0x03, 0x04,  // First word
			0x05, 0x06, 0x07, 0x08   // Second word
		);
		apiClient.GetMemoryAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(memoryData);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshStackAsync();

		// Assert
		viewModel.StackPointer.Should().Be(0x4FFF0);
		viewModel.StackEntries.Should().NotBeEmpty();
	}

	[Fact]
	public void UpdateRegisters_ShouldUpdateStackPointer()
	{
		// Arrange
		var registers = RegisterState.Create(sp: 0x48000);

		// Act
		viewModel.UpdateRegisters(registers);

		// Assert
		viewModel.StackPointer.Should().Be(0x48000);
	}

	[Fact]
	public async Task FormatStackEntries_ShouldCreateStackEntriesWithOffsets()
	{
		// Arrange
		var registers = RegisterState.Create(sp: 0x4FFF0);
		var memoryData = ImmutableArray.Create<byte>(
			0x00, 0x80, 0x00, 0x00,  // 0x8000 (code address)
			0xFF, 0xFF, 0xFF, 0xFF,  // 0xFFFFFFFF
			0x00, 0x00, 0x05, 0x00   // 0x50000 (stack address)
		);
		apiClient.GetMemoryAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(memoryData);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshStackAsync();

		// Assert
		viewModel.StackEntries.Should().HaveCountGreaterThan(0);
		viewModel.StackEntries[0].Offset.Should().Be("SP+0");
	}

	[Fact]
	public async Task FormatStackEntries_ShouldAnnotateCodeAddresses()
	{
		// Arrange
		var registers = RegisterState.Create(sp: 0x4FFF0);
		var memoryData = ImmutableArray.Create<byte>(
			0x00, 0x80, 0x00, 0x00  // 0x8000 (code address in typical range)
		);
		apiClient.GetMemoryAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(memoryData);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshStackAsync();

		// Assert
		viewModel.StackEntries.Should().Contain(e => e.Annotation == "code address");
	}

	[Fact]
	public async Task FormatStackEntries_ShouldAnnotateStackAddresses()
	{
		// Arrange
		var registers = RegisterState.Create(sp: 0x4FFF0);
		var memoryData = ImmutableArray.Create<byte>(
			0x00, 0x00, 0x05, 0x00  // 0x50000 (stack address)
		);
		apiClient.GetMemoryAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(memoryData);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshStackAsync();

		// Assert
		viewModel.StackEntries.Should().Contain(e => e.Annotation == "stack address");
	}

	[Fact]
	public async Task FormatStackEntries_ShouldAnnotateLinkRegister()
	{
		// Arrange
		var registers = RegisterState.Create(sp: 0x4FFF0, lr: 0x8100);
		var memoryData = ImmutableArray.Create<byte>(
			0x00, 0x81, 0x00, 0x00  // 0x8100 (matches LR)
		);
		apiClient.GetMemoryAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(memoryData);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshStackAsync();

		// Assert
		viewModel.StackEntries.Should().Contain(e => e.Annotation != null && e.Annotation.Contains("LR"));
	}

	[Fact]
	public async Task RefreshStackAsync_ReadsExactlyTheUsedStack()
	{
		viewModel.UpdateRegisters(RegisterState.Create(sp: 0x4FF00));
		apiClient.GetMemoryAsync(sessionId, 0x4FF00, 0x100, Arg.Any<CancellationToken>()).Returns(new byte[0x100].ToImmutableArray());

		await viewModel.RefreshStackAsync();

		await apiClient.Received(1).GetMemoryAsync(sessionId, 0x4FF00u, 0x100, Arg.Any<CancellationToken>());
		viewModel.StackEntries.Should().HaveCount(0x40);
	}

	[Fact]
	public async Task RefreshStackAsync_WithSpAtStackTop_ReadsNothingAndClearsEntries()
	{
		viewModel.StackMemory = [1, 0, 0, 0];
		viewModel.UpdateRegisters(RegisterState.Create(sp: 0x50000));

		await viewModel.RefreshStackAsync();

		await apiClient.DidNotReceive().GetMemoryAsync(Arg.Any<string>(), Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>());
		viewModel.StackEntries.Should().BeEmpty();
	}

	[Fact]
	public async Task RefreshStackAsync_WithStackLargerThanLimit_ReportsErrorAndReadsNothing()
	{
		viewModel.UpdateRegisters(RegisterState.Create(sp: 0x50000 - 0x10004));

		await viewModel.RefreshStackAsync();

		await apiClient.DidNotReceive().GetMemoryAsync(Arg.Any<string>(), Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>());
		viewModel.ErrorMessage.Should().Be("Stack of 65540 bytes exceeds the 65536 byte limit; SP may be corrupt");
	}

	[Fact]
	public async Task FormatStackEntries_ShouldAnnotateRegisterValues()
	{
		viewModel.UpdateRegisters(RegisterState.Create(sp: 0x4FFF0, r4: 0x1234, r0: 0));
		apiClient.GetMemoryAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns(ImmutableArray.Create<byte>(0x34, 0x12, 0, 0, 0, 0, 0, 0));

		await viewModel.RefreshStackAsync();

		viewModel.StackEntries.Select(e => e.Annotation).Should().Equal("R4", null);
	}

	[Fact]
	public void CalculateStackSize_ShouldReturnDifferenceFromStackTop()
	{
		// Arrange
		var registers = RegisterState.Create(sp: 0x4FF00);

		// Act
		viewModel.UpdateRegisters(registers);

		// Assert
		viewModel.StackSize.Should().Be(0x100); // 0x50000 - 0x4FF00 = 256 bytes
	}

	[Fact]
	public async Task RefreshCommand_ShouldReloadStackData()
	{
		// Arrange
		var registers = RegisterState.Create(sp: 0x4FFF0);
		apiClient.GetMemoryAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>())
			.Returns([0x01, 0x02, 0x03, 0x04]);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.RefreshCommand.Execute();

		// Assert
		await apiClient.Received(1).GetMemoryAsync(sessionId, Arg.Any<uint>(), Arg.Any<int>(), Arg.Any<CancellationToken>());
	}

	public void Dispose()
	{
		viewModel.Dispose();
		GC.SuppressFinalize(this);
	}
}
