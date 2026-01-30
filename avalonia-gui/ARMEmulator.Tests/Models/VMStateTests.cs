using ARMEmulator.Models;
using FluentAssertions;

namespace ARMEmulator.Tests.Models;

public sealed class VMStateTests
{
	[Theory]
	[InlineData(VMState.Idle, true)]
	[InlineData(VMState.Halted, true)]
	[InlineData(VMState.Error, true)]
	[InlineData(VMState.Running, false)]
	[InlineData(VMState.Breakpoint, false)]
	[InlineData(VMState.WaitingForInput, false)]
	public void IsEditorEditable_ReturnsCorrectValue(VMState state, bool expected)
	{
		state.IsEditorEditable().Should().Be(expected);
	}

	[Theory]
	[InlineData(VMState.Idle, true)]
	[InlineData(VMState.Breakpoint, true)]
	[InlineData(VMState.Running, false)]
	[InlineData(VMState.Halted, false)]
	[InlineData(VMState.Error, false)]
	[InlineData(VMState.WaitingForInput, false)]
	public void CanStep_ReturnsCorrectValue(VMState state, bool expected)
	{
		state.CanStep().Should().Be(expected);
	}

	[Theory]
	[InlineData(VMState.Running, true)]
	[InlineData(VMState.WaitingForInput, true)]
	[InlineData(VMState.Idle, false)]
	[InlineData(VMState.Breakpoint, false)]
	[InlineData(VMState.Halted, false)]
	[InlineData(VMState.Error, false)]
	public void CanPause_ReturnsCorrectValue(VMState state, bool expected)
	{
		state.CanPause().Should().Be(expected);
	}
}
