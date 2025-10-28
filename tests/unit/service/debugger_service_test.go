package service_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestNewDebuggerService(t *testing.T) {
	machine := vm.NewVM()
	svc := service.NewDebuggerService(machine)

	if svc == nil {
		t.Fatal("expected service instance, got nil")
	}

	if svc.GetVM() != machine {
		t.Error("service VM mismatch")
	}
}

func TestDebuggerService_LoadProgram(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	// Parse simple program with .org directive
	p := parser.NewParser(".org 0x8000\n_start:\nMOV R0, #42\nSWI #0", "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Load into service
	err = svc.LoadProgram(program, 0x8000)
	if err != nil {
		t.Fatalf("LoadProgram failed: %v", err)
	}

	// Verify PC set correctly
	if machine.CPU.PC != 0x8000 {
		t.Errorf("expected PC=0x8000, got 0x%08X", machine.CPU.PC)
	}
}
