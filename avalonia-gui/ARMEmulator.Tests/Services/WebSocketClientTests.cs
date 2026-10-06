using ARMEmulator.Models;
using ARMEmulator.Services;
using AwesomeAssertions;
using Xunit;

namespace ARMEmulator.Tests.Services;

/// <summary>
/// WebSocketClient tests. Message fixtures follow the payloads built in api/broadcaster.go,
/// api/handlers.go (broadcastStateChange) and api/session_manager.go (OnStateChange).
/// </summary>
public sealed class WebSocketClientTests
{
	[Fact]
	public void Constructor_CreatesDisconnectedClient()
	{
		using var client = new WebSocketClient("ws://localhost:8080/api/v1/ws");
		client.IsConnected.Should().BeFalse();
	}

	[Fact]
	public void ParseMessage_FullStateUpdate_ReturnsStatusAndRegisters()
	{
		const string message = """
			{"type":"state","sessionId":"s1","data":{"status":"breakpoint","pc":32772,
			"registers":{"r0":5,"r1":1,"r2":2,"r3":3,"r4":4,"r5":5,"r6":6,"r7":7,"r8":8,"r9":9,"r10":10,"r11":11,"r12":12,
			"sp":327680,"lr":32800,"pc":32772,"cpsr":{"n":false,"z":true,"c":false,"v":true}},
			"flags":{"n":false,"z":true,"c":false,"v":true}}}
			""";

		var evt = WebSocketClient.ParseMessage(message);

		var state = evt.Should().BeOfType<StateEvent>().Which;
		state.SessionId.Should().Be("s1");
		state.Status.State.Should().Be(VMState.Breakpoint);
		state.Status.PC.Should().Be(0x8004u);
		state.Registers.Should().Be(RegisterState.Create(
			r0: 5, r1: 1, r2: 2, r3: 3, r4: 4, r5: 5, r6: 6, r7: 7, r8: 8, r9: 9, r10: 10, r11: 11, r12: 12,
			sp: 0x50000, lr: 0x8020, pc: 0x8004, cpsr: new CPSRFlags(N: false, Z: true, C: false, V: true)));
	}

	[Fact]
	public void ParseMessage_StatusOnlyUpdate_ReturnsStateWithoutRegisters()
	{
		const string message = """{"type":"state","sessionId":"s1","data":{"status":"waiting_for_input"}}""";

		var evt = WebSocketClient.ParseMessage(message);

		var state = evt.Should().BeOfType<StateEvent>().Which;
		state.Status.State.Should().Be(VMState.WaitingForInput);
		state.Registers.Should().BeNull();
	}

	[Fact]
	public void ParseMessage_StateWithUnknownStatus_ReturnsNull()
	{
		const string message = """{"type":"state","sessionId":"s1","data":{"status":"exploded"}}""";

		WebSocketClient.ParseMessage(message).Should().BeNull();
	}

	[Fact]
	public void ParseMessage_Output_ReturnsStreamAndContent()
	{
		const string message = """{"type":"output","sessionId":"s1","data":{"stream":"stderr","content":"oops\n"}}""";

		var evt = WebSocketClient.ParseMessage(message);

		evt.Should().Be(new OutputEvent("s1", OutputStreamType.Stderr, "oops\n"));
	}

	[Theory]
	[InlineData("breakpoint_hit", ExecutionEventType.BreakpointHit)]
	[InlineData("halted", ExecutionEventType.Halted)]
	[InlineData("error", ExecutionEventType.Error)]
	public void ParseMessage_ExecutionEvent_MapsBackendEventNames(string wire, ExecutionEventType expected)
	{
		var message = $$$"""{"type":"event","sessionId":"s1","data":{"event":"{{{wire}}}","address":32772,"symbol":"loop","message":"m"}}""";

		var evt = WebSocketClient.ParseMessage(message);

		evt.Should().Be(new ExecutionEvent("s1", expected, 0x8004, "loop", "m"));
	}

	[Fact]
	public void ParseMessage_ExecutionEventWithUnknownName_ReturnsNull()
	{
		const string message = """{"type":"event","sessionId":"s1","data":{"event":"teleported"}}""";

		WebSocketClient.ParseMessage(message).Should().BeNull();
	}

	[Theory]
	[InlineData("""{"type":"mystery","sessionId":"s1","data":{}}""")]
	[InlineData("""{"type":"state","sessionId":"s1"}""")]
	[InlineData("not json")]
	public void ParseMessage_UnusableMessage_ReturnsNull(string message)
	{
		WebSocketClient.ParseMessage(message).Should().BeNull();
	}
}
