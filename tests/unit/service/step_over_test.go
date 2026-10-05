package service_test

import "testing"

// fact(n) = n * fact(n-1) with fact(1) = 1. Each frame keeps n in R4.
const recursiveFactorial = `.org 0x8000
_start:
	MOV R0, #3
	BL fact
	MOV R5, R0
	SWI #0x00
fact:
	STMFD SP!, {R4, LR}
	MOV R4, R0
	CMP R0, #1
	MOVLE R0, #1
	BLE done
	SUB R0, R0, #1
	BL fact
	MUL R0, R4, R0
done:
	LDMFD SP!, {R4, PC}
`

const (
	recursiveCallAddr  = 0x8028 // BL fact inside fact
	afterRecursiveCall = 0x802C
	stepsToInnerCall   = 8 // MOV, BL, STMFD, MOV, CMP, MOVLE, BLE, SUB
	topFrameN          = 3
	factorialOfTwo     = 2
)

func TestStepOverRecursiveCallReturnsToSameFrame(t *testing.T) {
	svc := newLoadedService(t, recursiveFactorial)
	stepN(t, svc, stepsToInnerCall)
	regs := svc.GetRegisterState()
	if regs.PC != recursiveCallAddr || regs.Registers[4] != topFrameN {
		t.Fatalf("setup: PC=0x%08X R4=%d, want PC=0x%X R4=%d", regs.PC, regs.Registers[4], recursiveCallAddr, topFrameN)
	}

	if err := svc.StepOver(); err != nil {
		t.Fatalf("StepOver: %v", err)
	}

	regs = svc.GetRegisterState()
	if regs.PC != afterRecursiveCall {
		t.Fatalf("PC = 0x%08X, want 0x%X", regs.PC, afterRecursiveCall)
	}
	if regs.Registers[4] != topFrameN {
		t.Errorf("stopped in a nested frame: R4 = %d, want %d", regs.Registers[4], topFrameN)
	}
	if regs.Registers[0] != factorialOfTwo {
		t.Errorf("R0 = %d, want fact(2) = %d", regs.Registers[0], factorialOfTwo)
	}
}

func TestStepOverNotTakenConditionalCallStopsAtNextInstruction(t *testing.T) {
	const src = `.org 0x8000
_start:
	MOVS R0, #0
	BLNE never
	MOV R1, #1
	SWI #0x00
never:
	MOV R2, #2
	MOV PC, LR
`
	const afterCall = 0x8008
	svc := newLoadedService(t, src)
	stepN(t, svc, 1)
	if err := svc.StepOver(); err != nil {
		t.Fatalf("StepOver: %v", err)
	}
	regs := svc.GetRegisterState()
	if regs.PC != afterCall || regs.Registers[1] != 0 || regs.Registers[2] != 0 {
		t.Errorf("PC=0x%08X R1=%d R2=%d, want PC=0x%X with nothing executed", regs.PC, regs.Registers[1], regs.Registers[2], afterCall)
	}
}
