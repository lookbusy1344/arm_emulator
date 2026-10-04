package service_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/service"
)

const countingLoop = `.org 0x8000
_start:
	MOV R0, #0
loop:
	ADD R0, R0, #1
	CMP R0, #3
	BNE loop
	SWI #0x00
`

const loopAddr = 0x8004

func runToStop(t *testing.T, svc *service.DebuggerService) {
	t.Helper()
	if err := waitDone(t, startRun(svc)); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}
}

func TestRunResumesPastBreakpoint(t *testing.T) {
	svc := newLoadedService(t, countingLoop)
	if err := svc.AddBreakpoint(loopAddr); err != nil {
		t.Fatal(err)
	}

	for want := uint32(0); want < 3; want++ {
		runToStop(t, svc)
		regs := svc.GetRegisterState()
		if svc.GetExecutionState() != service.StateBreakpoint || regs.PC != loopAddr {
			t.Fatalf("iteration %d: expected breakpoint at 0x%X, got %s at 0x%08X",
				want, loopAddr, svc.GetExecutionState(), regs.PC)
		}
		if regs.Registers[0] != want {
			t.Fatalf("iteration %d: expected R0=%d, got %d", want, want, regs.Registers[0])
		}
	}

	runToStop(t, svc)
	if state := svc.GetExecutionState(); state != service.StateHalted {
		t.Errorf("expected program to finish, got %s", state)
	}
}

func TestRunFromEntryStopsAtEntryBreakpoint(t *testing.T) {
	svc := newLoadedService(t, countingLoop)
	const entry = 0x8000
	if err := svc.AddBreakpoint(entry); err != nil {
		t.Fatal(err)
	}

	runToStop(t, svc)
	regs := svc.GetRegisterState()
	if regs.PC != entry || regs.Cycles != 0 {
		t.Fatalf("expected stop at entry before executing, got PC=0x%08X cycles=%d", regs.PC, regs.Cycles)
	}

	// Restarting returns to the entry point and stops there again.
	if err := svc.ResetToEntryPoint(); err != nil {
		t.Fatal(err)
	}
	runToStop(t, svc)
	if regs := svc.GetRegisterState(); regs.PC != entry || regs.Cycles != 0 {
		t.Fatalf("after restart expected stop at entry, got PC=0x%08X cycles=%d", regs.PC, regs.Cycles)
	}
}

func TestRunAfterStepOntoBreakpointExecutesIt(t *testing.T) {
	svc := newLoadedService(t, countingLoop)
	if err := svc.AddBreakpoint(loopAddr); err != nil {
		t.Fatal(err)
	}
	if err := svc.Step(); err != nil {
		t.Fatal(err)
	}

	runToStop(t, svc)
	if r0 := svc.GetRegisterState().Registers[0]; r0 != 1 {
		t.Errorf("expected the breakpoint instruction to execute before stopping, R0=%d", r0)
	}
}
