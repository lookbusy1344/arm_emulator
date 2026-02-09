using System.Collections.Immutable;
using ARMEmulator.Models;
using ARMEmulator.Services;
using ARMEmulator.ViewModels;
using FluentAssertions;
using NSubstitute;
using Xunit;

namespace ARMEmulator.Tests.ViewModels;

public class MemoryViewModelTests : IDisposable
{
	private readonly IApiClient apiClient;
	private readonly MemoryViewModel viewModel;
	private readonly string sessionId = "test-session";

	public MemoryViewModelTests()
	{
		apiClient = Substitute.For<IApiClient>();
		viewModel = new MemoryViewModel(apiClient) { SessionId = sessionId };
	}

	[Fact]
	public void InitialState_ShouldBeDefaultValues()
	{
		viewModel.CurrentAddress.Should().Be(0);
		viewModel.AutoScrollToWrites.Should().BeTrue();
		viewModel.MemoryData.Should().BeEmpty();
		viewModel.LastWriteAddress.Should().BeNull();
	}

	[Fact]
	public async Task LoadMemoryAsync_ShouldFetchMemoryFromApi()
	{
		// Arrange
		var expectedData = ImmutableArray.Create<byte>(0x01, 0x02, 0x03, 0x04);
		apiClient.GetMemoryAsync(sessionId, 0x8000, 256, Arg.Any<CancellationToken>())
			.Returns(expectedData);

		// Act
		await viewModel.LoadMemoryAsync(0x8000);

		// Assert
		viewModel.MemoryData.Should().Equal(expectedData);
		viewModel.CurrentAddress.Should().Be(0x8000);
	}

	[Fact]
	public async Task NavigateToAddressCommand_ShouldLoadMemoryAtAddress()
	{
		// Arrange
		var expectedData = ImmutableArray.Create<byte>(0xAA, 0xBB, 0xCC, 0xDD);
		apiClient.GetMemoryAsync(sessionId, 0x1000, 256, Arg.Any<CancellationToken>())
			.Returns(expectedData);

		viewModel.AddressInput = "0x1000";

		// Act
		await viewModel.NavigateToAddressCommand.Execute();

		// Assert
		viewModel.CurrentAddress.Should().Be(0x1000);
		viewModel.MemoryData.Should().Equal(expectedData);
	}

	[Fact]
	public async Task NavigateToAddressCommand_ShouldParseDecimalAddress()
	{
		// Arrange
		apiClient.GetMemoryAsync(sessionId, 4096, 256, Arg.Any<CancellationToken>())
			.Returns([]);

		viewModel.AddressInput = "4096";

		// Act
		await viewModel.NavigateToAddressCommand.Execute();

		// Assert
		viewModel.CurrentAddress.Should().Be(4096);
	}

	[Fact]
	public async Task NavigateToAddressCommand_ShouldHandleInvalidAddress()
	{
		// Arrange
		viewModel.AddressInput = "invalid";

		// Act
		await viewModel.NavigateToAddressCommand.Execute();

		// Assert
		viewModel.ErrorMessage.Should().NotBeNullOrEmpty();
		viewModel.ErrorMessage.Should().Contain("Invalid address");
	}

	[Fact]
	public async Task JumpToPCCommand_ShouldLoadMemoryAtPCAddress()
	{
		// Arrange
		var registers = RegisterState.Create(pc: 0x8000);
		apiClient.GetMemoryAsync(sessionId, 0x8000, 256, Arg.Any<CancellationToken>())
			.Returns([]);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.JumpToPCCommand.Execute();

		// Assert
		viewModel.CurrentAddress.Should().Be(0x8000);
	}

	[Fact]
	public async Task JumpToSPCommand_ShouldLoadMemoryAtSPAddress()
	{
		// Arrange
		var registers = RegisterState.Create(sp: 0x50000);
		apiClient.GetMemoryAsync(sessionId, 0x50000, 256, Arg.Any<CancellationToken>())
			.Returns([]);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.JumpToSPCommand.Execute();

		// Assert
		viewModel.CurrentAddress.Should().Be(0x50000);
	}

	[Fact]
	public async Task JumpToRegisterCommand_ShouldLoadMemoryAtRegisterValue()
	{
		// Arrange
		var registers = RegisterState.Create(r0: 0x2000);
		apiClient.GetMemoryAsync(sessionId, 0x2000, 256, Arg.Any<CancellationToken>())
			.Returns([]);

		viewModel.UpdateRegisters(registers);

		// Act
		await viewModel.JumpToRegisterCommand.Execute(0); // R0

		// Assert
		viewModel.CurrentAddress.Should().Be(0x2000);
	}

	[Fact]
	public void UpdateMemoryWrite_ShouldSetLastWriteAddress()
	{
		// Act
		viewModel.UpdateMemoryWrite(new MemoryWrite(0x1234, 4));

		// Assert
		viewModel.LastWriteAddress.Should().Be(0x1234);
	}

	[Fact]
	public async Task UpdateMemoryWrite_WithAutoScroll_ShouldNavigateToWriteAddress()
	{
		// Arrange
		viewModel.AutoScrollToWrites = true;
		apiClient.GetMemoryAsync(sessionId, 0x5000, 256, Arg.Any<CancellationToken>())
			.Returns([]);

		// Act
		viewModel.UpdateMemoryWrite(new MemoryWrite(0x5000, 4));
		await Task.Delay(50); // Allow async navigation to complete

		// Assert
		viewModel.CurrentAddress.Should().Be(0x5000);
	}

	[Fact]
	public void UpdateMemoryWrite_WithoutAutoScroll_ShouldNotNavigate()
	{
		// Arrange
		viewModel.AutoScrollToWrites = false;
		var originalAddress = viewModel.CurrentAddress;

		// Act
		viewModel.UpdateMemoryWrite(new MemoryWrite(0x5000, 4));

		// Assert
		viewModel.CurrentAddress.Should().Be(originalAddress);
	}

	[Fact]
	public void FormatMemoryRows_ShouldCreateHexDumpRows()
	{
		// Arrange
		var data = ImmutableArray.Create<byte>(
			0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F,
			0x72, 0x6C, 0x64, 0x00, 0x00, 0x00, 0x00, 0x00
		);
		viewModel.MemoryData = data;
		viewModel.CurrentAddress = 0x8000;

		// Act
		var rows = viewModel.FormattedRows;

		// Assert
		rows.Should().HaveCount(1);
		rows[0].Address.Should().Be("00008000");
		rows[0].HexBytes.Should().Contain("48 65 6C 6C");
		rows[0].AsciiText.Should().StartWith("Hello");
	}

	public void Dispose()
	{
		viewModel.Dispose();
		GC.SuppressFinalize(this);
	}
}
