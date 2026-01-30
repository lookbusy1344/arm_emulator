using ARMEmulator.Models;
using FluentAssertions;

namespace ARMEmulator.Tests.Models;

public sealed class EmulatorEventTests
{
	[Fact]
	public void StateEvent_StoresAllProperties()
	{
		var status = new VMStatus(VMState.Running, 0x8004, 100);
		var registers = RegisterState.Create(r0: 42);
		var evt = new StateEvent("session-123", status, registers);

		evt.SessionId.Should().Be("session-123");
		evt.Status.Should().Be(status);
		evt.Registers.Should().Be(registers);
	}

	[Fact]
	public void OutputEvent_StoresAllProperties()
	{
		var evt = new OutputEvent("session-123", OutputStreamType.Stdout, "Hello, World!\n");

		evt.SessionId.Should().Be("session-123");
		evt.Stream.Should().Be(OutputStreamType.Stdout);
		evt.Content.Should().Be("Hello, World!\n");
	}

	[Fact]
	public void OutputEvent_WithStderr_StoresCorrectStream()
	{
		var evt = new OutputEvent("session-123", OutputStreamType.Stderr, "Error occurred\n");

		evt.Stream.Should().Be(OutputStreamType.Stderr);
	}

	[Fact]
	public void ExecutionEvent_WithBreakpoint_StoresCorrectType()
	{
		var evt = new ExecutionEvent(
			"session-123",
			ExecutionEventType.BreakpointHit,
			Address: 0x8000
		);

		evt.SessionId.Should().Be("session-123");
		evt.EventType.Should().Be(ExecutionEventType.BreakpointHit);
		evt.Address.Should().Be(0x8000u);
		evt.Symbol.Should().BeNull();
		evt.Message.Should().BeNull();
	}

	[Fact]
	public void ExecutionEvent_WithHalted_StoresMessage()
	{
		var evt = new ExecutionEvent(
			"session-123",
			ExecutionEventType.Halted,
			Message: "Program completed successfully"
		);

		evt.EventType.Should().Be(ExecutionEventType.Halted);
		evt.Message.Should().Be("Program completed successfully");
		evt.Address.Should().BeNull();
	}

	[Fact]
	public void ExecutionEvent_WithError_StoresErrorMessage()
	{
		var evt = new ExecutionEvent(
			"session-123",
			ExecutionEventType.Error,
			Message: "Invalid instruction"
		);

		evt.EventType.Should().Be(ExecutionEventType.Error);
		evt.Message.Should().Be("Invalid instruction");
	}

	[Fact]
	public void EmulatorEvent_CanBePatternMatched()
	{
		EmulatorEvent evt = new StateEvent(
			"session-123",
			new VMStatus(VMState.Running, 0x8000, 0),
			RegisterState.Create()
		);

		var result = evt switch {
			StateEvent se => $"State: {se.Status.State}",
			OutputEvent oe => $"Output: {oe.Content}",
			ExecutionEvent ee => $"Execution: {ee.EventType}",
			_ => "Unknown"
		};

		result.Should().Be("State: Running");
	}

	[Fact]
	public void EmulatorEvent_BaseClassProvidesSessionId()
	{
		EmulatorEvent evt = new OutputEvent("session-456", OutputStreamType.Stdout, "test");

		evt.SessionId.Should().Be("session-456");
	}

	[Theory]
	[InlineData(ExecutionEventType.BreakpointHit)]
	[InlineData(ExecutionEventType.Halted)]
	[InlineData(ExecutionEventType.Error)]
	public void ExecutionEventType_HasAllValues(ExecutionEventType type)
	{
		// Verify enum values exist
		type.Should().BeOneOf(
			ExecutionEventType.BreakpointHit,
			ExecutionEventType.Halted,
			ExecutionEventType.Error
		);
	}

	[Theory]
	[InlineData(OutputStreamType.Stdout)]
	[InlineData(OutputStreamType.Stderr)]
	public void OutputStreamType_HasAllValues(OutputStreamType stream)
	{
		// Verify enum values exist
		stream.Should().BeOneOf(OutputStreamType.Stdout, OutputStreamType.Stderr);
	}
}
