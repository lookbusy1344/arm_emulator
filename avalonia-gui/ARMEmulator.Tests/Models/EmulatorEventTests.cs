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

		_ = evt.SessionId.Should().Be("session-123");
		_ = evt.Status.Should().Be(status);
		_ = evt.Registers.Should().Be(registers);
	}

	[Fact]
	public void OutputEvent_StoresAllProperties()
	{
		var evt = new OutputEvent("session-123", OutputStreamType.Stdout, "Hello, World!\n");

		_ = evt.SessionId.Should().Be("session-123");
		_ = evt.Stream.Should().Be(OutputStreamType.Stdout);
		_ = evt.Content.Should().Be("Hello, World!\n");
	}

	[Fact]
	public void OutputEvent_WithStderr_StoresCorrectStream()
	{
		var evt = new OutputEvent("session-123", OutputStreamType.Stderr, "Error occurred\n");

		_ = evt.Stream.Should().Be(OutputStreamType.Stderr);
	}

	[Fact]
	public void ExecutionEvent_WithBreakpoint_StoresCorrectType()
	{
		var evt = new ExecutionEvent(
			"session-123",
			ExecutionEventType.BreakpointHit,
			Address: 0x8000
		);

		_ = evt.SessionId.Should().Be("session-123");
		_ = evt.EventType.Should().Be(ExecutionEventType.BreakpointHit);
		_ = evt.Address.Should().Be(0x8000u);
		_ = evt.Symbol.Should().BeNull();
		_ = evt.Message.Should().BeNull();
	}

	[Fact]
	public void ExecutionEvent_WithHalted_StoresMessage()
	{
		var evt = new ExecutionEvent(
			"session-123",
			ExecutionEventType.Halted,
			Message: "Program completed successfully"
		);

		_ = evt.EventType.Should().Be(ExecutionEventType.Halted);
		_ = evt.Message.Should().Be("Program completed successfully");
		_ = evt.Address.Should().BeNull();
	}

	[Fact]
	public void ExecutionEvent_WithError_StoresErrorMessage()
	{
		var evt = new ExecutionEvent(
			"session-123",
			ExecutionEventType.Error,
			Message: "Invalid instruction"
		);

		_ = evt.EventType.Should().Be(ExecutionEventType.Error);
		_ = evt.Message.Should().Be("Invalid instruction");
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

		_ = result.Should().Be("State: Running");
	}

	[Fact]
	public void EmulatorEvent_BaseClassProvidesSessionId()
	{
		EmulatorEvent evt = new OutputEvent("session-456", OutputStreamType.Stdout, "test");

		_ = evt.SessionId.Should().Be("session-456");
	}

	[Theory]
	[InlineData(ExecutionEventType.BreakpointHit)]
	[InlineData(ExecutionEventType.Halted)]
	[InlineData(ExecutionEventType.Error)]
	public void ExecutionEventType_HasAllValues(ExecutionEventType type)
	{
		// Verify enum values exist
		_ = type.Should().BeOneOf(
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
		_ = stream.Should().BeOneOf(OutputStreamType.Stdout, OutputStreamType.Stderr);
	}
}
