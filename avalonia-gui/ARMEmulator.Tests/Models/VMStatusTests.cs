using ARMEmulator.Models;
using FluentAssertions;

namespace ARMEmulator.Tests.Models;

public sealed class VMStatusTests
{
	[Fact]
	public void Constructor_WithMinimalParameters_CreatesInstance()
	{
		var status = new VMStatus(
			State: VMState.Idle,
			PC: 0x8000,
			Cycles: 0
		);

		_ = status.State.Should().Be(VMState.Idle);
		_ = status.PC.Should().Be(0x8000u);
		_ = status.Cycles.Should().Be(0ul);
		_ = status.Error.Should().BeNull();
		_ = status.LastWrite.Should().BeNull();
	}

	[Fact]
	public void Constructor_WithError_StoresErrorMessage()
	{
		var status = new VMStatus(
			State: VMState.Error,
			PC: 0x8000,
			Cycles: 100,
			Error: "Division by zero"
		);

		_ = status.State.Should().Be(VMState.Error);
		_ = status.Error.Should().Be("Division by zero");
	}

	[Fact]
	public void Constructor_WithMemoryWrite_StoresWriteInfo()
	{
		var write = new MemoryWrite(0x10000, 4);
		var status = new VMStatus(
			State: VMState.Running,
			PC: 0x8004,
			Cycles: 50,
			LastWrite: write
		);

		_ = status.LastWrite.Should().NotBeNull();
		_ = status.LastWrite!.Address.Should().Be(0x10000u);
		_ = status.LastWrite.Size.Should().Be(4u);
	}

	[Fact]
	public void WithExpression_CreatesModifiedCopy()
	{
		var original = new VMStatus(VMState.Idle, 0x8000, 0);
		var modified = original with { State = VMState.Running, Cycles = 10 };

		_ = original.State.Should().Be(VMState.Idle);
		_ = original.Cycles.Should().Be(0ul);
		_ = modified.State.Should().Be(VMState.Running);
		_ = modified.Cycles.Should().Be(10ul);
		_ = modified.PC.Should().Be(0x8000u);
	}

	[Fact]
	public void Equality_WithSameValues_AreEqual()
	{
		var status1 = new VMStatus(VMState.Idle, 0x8000, 100);
		var status2 = new VMStatus(VMState.Idle, 0x8000, 100);

		_ = status1.Should().Be(status2);
	}

	[Fact]
	public void Equality_WithDifferentValues_AreNotEqual()
	{
		var status1 = new VMStatus(VMState.Idle, 0x8000, 100);
		var status2 = new VMStatus(VMState.Running, 0x8000, 100);

		_ = status1.Should().NotBe(status2);
	}
}

public sealed class MemoryWriteTests
{
	[Fact]
	public void Constructor_StoresAddressAndSize()
	{
		var write = new MemoryWrite(0x10000, 4);

		_ = write.Address.Should().Be(0x10000u);
		_ = write.Size.Should().Be(4u);
	}

	[Fact]
	public void Equality_WithSameValues_AreEqual()
	{
		var write1 = new MemoryWrite(0x10000, 4);
		var write2 = new MemoryWrite(0x10000, 4);

		_ = write1.Should().Be(write2);
	}
}
