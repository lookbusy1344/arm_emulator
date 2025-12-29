package service_test

import (
	"testing"
	"time"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestDebuggerService_StepExecution(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	// Load simple program: MOV R0, #42; SWI #0
	p := parser.NewParser(".org 0x8000\n_start:\nMOV R0, #42\nSWI #0", "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if err := svc.LoadProgram(program, 0x8000); err != nil {
		t.Fatalf("LoadProgram failed: %v", err)
	}

	// Initial state should be halted
	state := svc.GetExecutionState()
	if state != service.StateHalted {
		t.Errorf("expected StateHalted, got %s", state)
	}

	// Execute one step
	if err := svc.Step(); err != nil {
		t.Fatalf("Step failed: %v", err)
	}

	// Check register changed
	regs := svc.GetRegisterState()
	if regs.Registers[0] != 42 {
		t.Errorf("expected R0=42, got %d", regs.Registers[0])
	}
}

func TestDebuggerService_ContinueExecution(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	// Load program with loop
	code := `.org 0x8000
_start:
	MOV R0, #0
loop:
	ADD R0, R0, #1
	CMP R0, #10
	BLT loop
	SWI #0`

	p := parser.NewParser(code, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if err := svc.LoadProgram(program, 0x8000); err != nil {
		t.Fatalf("LoadProgram failed: %v", err)
	}

	// Start execution in background (must set running state first)
	svc.SetRunning(true)
	errChan := make(chan error, 1)
	go func() {
		errChan <- svc.RunUntilHalt()
	}()

	// Wait a bit for execution
	time.Sleep(10 * time.Millisecond)

	// Wait for completion
	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("RunUntilHalt failed: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("execution timeout")
	}

	// Check final state
	regs := svc.GetRegisterState()
	if regs.Registers[0] != 10 {
		t.Errorf("expected R0=10, got %d", regs.Registers[0])
	}
}
