package service_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/service"
)

const nestedCalls = `.org 0x8000
_start:
	BL outer
	MOV R0, #1
	SWI #0x00
outer:
	STMFD SP!, {LR}
	MOV R1, #2
	BL leaf
	MOV R2, #3
	LDMFD SP!, {PC}
leaf:
	MOV R3, #4
	MOV PC, LR
bxleaf:
	MOV R4, #5
	BX LR
`

const (
	afterOuterCall = 0x8004
	outerAddr      = 0x800C
	afterLeafCall  = 0x8018
	leafAddr       = 0x8020
)

func stepN(t *testing.T, svc *service.DebuggerService, n int) {
	t.Helper()
	for range n {
		if err := svc.Step(); err != nil {
			t.Fatalf("Step failed: %v", err)
		}
	}
}

func TestStepOutPassesOverNestedCall(t *testing.T) {
	svc := newLoadedService(t, nestedCalls)
	stepN(t, svc, 1) // into outer
	if pc := svc.GetRegisterState().PC; pc != outerAddr {
		t.Fatalf("expected to be in outer at 0x%X, got 0x%08X", outerAddr, pc)
	}

	if err := svc.StepOut(); err != nil {
		t.Fatalf("StepOut failed: %v", err)
	}

	regs := svc.GetRegisterState()
	if regs.PC != afterOuterCall {
		t.Fatalf("expected PC=0x%X after step out, got 0x%08X", afterOuterCall, regs.PC)
	}
	if regs.Registers[1] != 2 || regs.Registers[2] != 3 || regs.Registers[3] != 4 {
		t.Errorf("expected outer and leaf to finish, R1=%d R2=%d R3=%d",
			regs.Registers[1], regs.Registers[2], regs.Registers[3])
	}
	if regs.Registers[0] != 0 {
		t.Error("step out ran past the return address")
	}
	if state := svc.GetExecutionState(); state != service.StateBreakpoint {
		t.Errorf("expected breakpoint state after step out, got %s", state)
	}
}

func TestStepOutOfLeafReturnsToCaller(t *testing.T) {
	svc := newLoadedService(t, nestedCalls)
	stepN(t, svc, 4) // BL outer, STMFD, MOV R1, BL leaf
	if pc := svc.GetRegisterState().PC; pc != leafAddr {
		t.Fatalf("expected to be in leaf at 0x%X, got 0x%08X", leafAddr, pc)
	}

	if err := svc.StepOut(); err != nil {
		t.Fatalf("StepOut failed: %v", err)
	}
	regs := svc.GetRegisterState()
	if regs.PC != afterLeafCall {
		t.Fatalf("expected PC=0x%X, got 0x%08X", afterLeafCall, regs.PC)
	}
	if regs.Registers[2] != 0 {
		t.Error("step out ran past the return address")
	}
}

func TestStepOutRecognisesBXReturn(t *testing.T) {
	const program = `.org 0x8000
_start:
	BL bxleaf
	MOV R0, #1
	SWI #0x00
bxleaf:
	MOV R4, #5
	BX LR
`
	svc := newLoadedService(t, program)
	stepN(t, svc, 1)

	if err := svc.StepOut(); err != nil {
		t.Fatalf("StepOut failed: %v", err)
	}
	if regs := svc.GetRegisterState(); regs.PC != 0x8004 || regs.Registers[4] != 5 {
		t.Fatalf("expected PC=0x8004 with R4=5, got PC=0x%08X R4=%d", regs.PC, regs.Registers[4])
	}
}

func TestStepOutStopsAtBreakpointInsideFunction(t *testing.T) {
	svc := newLoadedService(t, nestedCalls)
	stepN(t, svc, 1)
	if err := svc.AddBreakpoint(leafAddr); err != nil {
		t.Fatal(err)
	}

	if err := svc.StepOut(); err != nil {
		t.Fatalf("StepOut failed: %v", err)
	}
	if pc := svc.GetRegisterState().PC; pc != leafAddr {
		t.Fatalf("expected stop at breakpoint 0x%X, got 0x%08X", leafAddr, pc)
	}
}

func TestStepOutFromTopLevelRunsToExit(t *testing.T) {
	svc := newLoadedService(t, nestedCalls)

	if err := svc.StepOut(); err != nil {
		t.Fatalf("StepOut failed: %v", err)
	}
	if state := svc.GetExecutionState(); state != service.StateHalted {
		t.Errorf("expected program to finish, got %s", state)
	}
}
