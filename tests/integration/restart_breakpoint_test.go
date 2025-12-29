package integration

import (
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestRestartWithBreakpoint tests the exact scenario that's failing in E2E tests:
// 1. Load program
// 2. Step 3 times
// 3. Set breakpoint at current PC
// 4. Restart (should reset PC to entry point but preserve program and breakpoints)
// 5. RunUntilHalt (should execute until hitting the breakpoint)
// 6. Verify PC stopped at breakpoint, not at entry point
func TestRestartWithBreakpoint(t *testing.T) {
	// Create VM and service
	machine := vm.NewVM()
	stackTop := uint32(vm.StackSegmentStart + vm.StackSegmentSize)
	machine.InitializeStack(stackTop)
	svc := service.NewDebuggerService(machine)

	// Load fibonacci program (same as E2E test)
	entryPoint := uint32(0x00008000)
	source := `.org 0x8000
    .text
    .global _start
_start:
    MOV R0, #10        ; Calculate 10 Fibonacci numbers
    MOV R1, #0         ; First number
    MOV R2, #1         ; Second number
loop:
    CMP R0, #0
    BEQ done
    MOV R3, R1
    ADD R1, R1, R2
    MOV R2, R3
    SUB R0, R0, #1
    B loop
done:
    SWI #0x00          ; EXIT
`
	p := parser.NewParser(source, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	err = svc.LoadProgram(program, entryPoint)
	if err != nil {
		t.Fatalf("LoadProgram failed: %v", err)
	}

	// Verify PC is at entry point
	state := svc.GetRegisterState()
	if state.PC != entryPoint {
		t.Fatalf("After load, PC=0x%08X, expected 0x%08X", state.PC, entryPoint)
	}
	t.Logf("✓ After load: PC=0x%08X (entry point)", state.PC)

	// Step 3 times
	for i := 0; i < 3; i++ {
		err = svc.Step()
		if err != nil {
			t.Fatalf("Step %d failed: %v", i+1, err)
		}
	}

	// Get current PC - this is where we'll set the breakpoint
	state = svc.GetRegisterState()
	breakpointAddr := state.PC
	t.Logf("✓ After 3 steps: PC=0x%08X (breakpoint location)", breakpointAddr)

	if breakpointAddr == entryPoint {
		t.Fatalf("After 3 steps, PC still at entry point - program didn't execute")
	}

	// Set breakpoint at current PC
	err = svc.AddBreakpoint(breakpointAddr)
	if err != nil {
		t.Fatalf("AddBreakpoint failed: %v", err)
	}
	t.Logf("✓ Breakpoint set at 0x%08X", breakpointAddr)

	// Restart - should reset PC to entry point but preserve program and breakpoints
	err = svc.ResetToEntryPoint()
	if err != nil {
		t.Fatalf("ResetToEntryPoint failed: %v", err)
	}

	state = svc.GetRegisterState()
	if state.PC != entryPoint {
		t.Fatalf("After restart, PC=0x%08X, expected 0x%08X (entry point)", state.PC, entryPoint)
	}
	t.Logf("✓ After restart: PC=0x%08X (back at entry point)", state.PC)

	// Verify breakpoint still exists
	breakpoints := svc.GetBreakpoints()
	if len(breakpoints) != 1 {
		t.Fatalf("After restart, found %d breakpoints, expected 1", len(breakpoints))
	}
	if breakpoints[0].Address != breakpointAddr {
		t.Fatalf("Breakpoint address changed from 0x%08X to 0x%08X", breakpointAddr, breakpoints[0].Address)
	}
	t.Logf("✓ Breakpoint preserved at 0x%08X", breakpointAddr)

	// Check VM state before running
	vmState := svc.GetVM()
	t.Logf("Before RunUntilHalt: vm.State=%v, vm.EntryPoint=0x%08X, vm.StackTop=0x%08X",
		vmState.State, vmState.EntryPoint, vmState.StackTop)

	// RunUntilHalt - should execute until hitting the breakpoint
	t.Logf("Calling RunUntilHalt()...")
	svc.SetRunning(true) // Must set running state before RunUntilHalt
	err = svc.RunUntilHalt()
	if err != nil {
		// Some error is expected if we hit a breakpoint, but not other errors
		if !strings.Contains(err.Error(), "breakpoint") {
			t.Logf("RunUntilHalt error (may be normal): %v", err)
		}
	}

	// Check execution state
	execState := svc.GetExecutionState()
	t.Logf("After RunUntilHalt: execution state=%s", execState)

	// Verify PC stopped at breakpoint
	state = svc.GetRegisterState()
	t.Logf("Final PC=0x%08X, expected 0x%08X (breakpoint)", state.PC, breakpointAddr)

	if state.PC == entryPoint {
		t.Fatalf("FAILURE: PC=0x%08X (entry point), program never executed! Expected PC=0x%08X (breakpoint)",
			state.PC, breakpointAddr)
	}

	if state.PC != breakpointAddr {
		t.Fatalf("FAILURE: PC=0x%08X, expected 0x%08X (breakpoint)", state.PC, breakpointAddr)
	}

	if execState != service.StateBreakpoint {
		t.Fatalf("FAILURE: Execution state=%s, expected %s", execState, service.StateBreakpoint)
	}

	t.Logf("✓ SUCCESS: Stopped at breakpoint 0x%08X with state=%s", state.PC, execState)
}
