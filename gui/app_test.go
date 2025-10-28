package main

import (
	"testing"
)

func TestApp_LoadProgram(t *testing.T) {
	app := NewApp()

	// Parse simple program
	source := ".org 0x8000\n_start:\nMOV R0, #42\nSWI #0"
	err := app.LoadProgramFromSource(source, "test.s", 0x8000)
	if err != nil {
		t.Fatalf("LoadProgramFromSource failed: %v", err)
	}

	// Get registers
	regs := app.GetRegisters()
	if regs.PC != 0x8000 {
		t.Errorf("expected PC=0x8000, got 0x%08X", regs.PC)
	}
}

func TestApp_StepExecution(t *testing.T) {
	app := NewApp()

	source := ".org 0x8000\n_start:\nMOV R0, #42\nSWI #0"
	if err := app.LoadProgramFromSource(source, "test.s", 0x8000); err != nil {
		t.Fatalf("LoadProgramFromSource failed: %v", err)
	}

	// Step once
	err := app.Step()
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}

	// Check R0 changed
	regs := app.GetRegisters()
	if regs.Registers[0] != 42 {
		t.Errorf("expected R0=42, got %d", regs.Registers[0])
	}
}
