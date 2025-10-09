package vm_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestExecutionTrace_BasicRecording(t *testing.T) {
	var buf bytes.Buffer
	trace := vm.NewExecutionTrace(&buf)

	// Create VM instance
	machine := vm.NewVM()
	machine.CPU.R[0] = 10
	machine.CPU.R[1] = 20
	machine.CPU.PC = 0x8000

	// Record an instruction
	trace.RecordInstruction(machine, "MOV R0, #10")

	// Verify entry was recorded
	entries := trace.GetEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Address != 0x8000-4 { // PC already advanced
		t.Errorf("Expected address 0x7FFC, got 0x%04X", entry.Address)
	}
	if entry.Disassembly != "MOV R0, #10" {
		t.Errorf("Expected disassembly 'MOV R0, #10', got '%s'", entry.Disassembly)
	}
}

func TestExecutionTrace_RegisterFiltering(t *testing.T) {
	var buf bytes.Buffer
	trace := vm.NewExecutionTrace(&buf)

	// Set filter to only track R0 and PC
	trace.SetFilterRegisters([]string{"R0", "PC"})

	// Create VM instance
	machine := vm.NewVM()
	machine.CPU.R[0] = 10
	machine.CPU.R[1] = 20
	machine.CPU.R[2] = 30

	// Record instruction
	trace.RecordInstruction(machine, "ADD R0, R1, R2")

	// Get entry
	entries := trace.GetEntries()
	if len(entries) == 0 {
		t.Fatal("No entries recorded")
	}

	entry := entries[0]
	// Should only have R0 and PC in changes (if they changed)
	for reg := range entry.RegisterChanges {
		if reg != "R0" && reg != "PC" {
			t.Errorf("Unexpected register %s in filtered trace", reg)
		}
	}
}

func TestExecutionTrace_Flush(t *testing.T) {
	var buf bytes.Buffer
	trace := vm.NewExecutionTrace(&buf)

	// Create VM
	machine := vm.NewVM()
	machine.CPU.PC = 0x8000

	// Record some instructions
	trace.RecordInstruction(machine, "MOV R0, #1")
	machine.CPU.PC += 4
	trace.RecordInstruction(machine, "MOV R1, #2")
	machine.CPU.PC += 4
	trace.RecordInstruction(machine, "ADD R2, R0, R1")

	// Flush to buffer
	if err := trace.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Flush produced no output")
	}

	// Should contain all three instructions
	if !strings.Contains(output, "MOV R0, #1") {
		t.Error("Output missing first instruction")
	}
	if !strings.Contains(output, "MOV R1, #2") {
		t.Error("Output missing second instruction")
	}
	if !strings.Contains(output, "ADD R2, R0, R1") {
		t.Error("Output missing third instruction")
	}
}

func TestExecutionTrace_MaxEntries(t *testing.T) {
	var buf bytes.Buffer
	trace := vm.NewExecutionTrace(&buf)
	trace.MaxEntries = 10

	machine := vm.NewVM()

	// Record more than max entries
	for i := 0; i < 20; i++ {
		trace.RecordInstruction(machine, "NOP")
		machine.CPU.PC += 4
	}

	// Should only have MaxEntries
	entries := trace.GetEntries()
	if len(entries) != 10 {
		t.Errorf("Expected 10 entries (max), got %d", len(entries))
	}
}

func TestExecutionTrace_Clear(t *testing.T) {
	var buf bytes.Buffer
	trace := vm.NewExecutionTrace(&buf)

	machine := vm.NewVM()

	// Record some entries
	trace.RecordInstruction(machine, "MOV R0, #1")
	trace.RecordInstruction(machine, "MOV R1, #2")

	// Verify entries exist
	if len(trace.GetEntries()) != 2 {
		t.Error("Expected 2 entries before clear")
	}

	// Clear
	trace.Clear()

	// Verify cleared
	if len(trace.GetEntries()) != 0 {
		t.Error("Expected 0 entries after clear")
	}
}

func TestMemoryTrace_RecordRead(t *testing.T) {
	var buf bytes.Buffer
	trace := vm.NewMemoryTrace(&buf)

	// Record a read
	trace.RecordRead(1, 0x8000, 0x20000, 0x12345678, "WORD")

	// Verify entry
	entries := trace.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Type != "READ" {
		t.Errorf("Expected type READ, got %s", entry.Type)
	}
	if entry.Address != 0x20000 {
		t.Errorf("Expected address 0x20000, got 0x%08X", entry.Address)
	}
	if entry.Value != 0x12345678 {
		t.Errorf("Expected value 0x12345678, got 0x%08X", entry.Value)
	}
	if entry.Size != "WORD" {
		t.Errorf("Expected size WORD, got %s", entry.Size)
	}
}

func TestMemoryTrace_RecordWrite(t *testing.T) {
	var buf bytes.Buffer
	trace := vm.NewMemoryTrace(&buf)

	// Record a write
	trace.RecordWrite(1, 0x8000, 0x20000, 0xDEADBEEF, "WORD")

	// Verify entry
	entries := trace.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Type != "WRITE" {
		t.Errorf("Expected type WRITE, got %s", entry.Type)
	}
	if entry.Value != 0xDEADBEEF {
		t.Errorf("Expected value 0xDEADBEEF, got 0x%08X", entry.Value)
	}
}

func TestMemoryTrace_Flush(t *testing.T) {
	var buf bytes.Buffer
	trace := vm.NewMemoryTrace(&buf)

	// Record some accesses
	trace.RecordRead(1, 0x8000, 0x20000, 0x11111111, "WORD")
	trace.RecordWrite(2, 0x8004, 0x20004, 0x22222222, "WORD")
	trace.RecordRead(3, 0x8008, 0x20008, 0x33, "BYTE")

	// Flush
	if err := trace.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Flush produced no output")
	}

	// Should contain READ and WRITE markers
	if !strings.Contains(output, "READ") {
		t.Error("Output missing READ marker")
	}
	if !strings.Contains(output, "WRITE") {
		t.Error("Output missing WRITE marker")
	}
}

func TestMemoryTrace_MaxEntries(t *testing.T) {
	var buf bytes.Buffer
	trace := vm.NewMemoryTrace(&buf)
	trace.MaxEntries = 5

	// Record more than max
	for i := 0; i < 10; i++ {
		trace.RecordRead(uint64(i), 0x8000, 0x20000, 0, "WORD")
	}

	// Should only have MaxEntries
	entries := trace.GetEntries()
	if len(entries) != 5 {
		t.Errorf("Expected 5 entries (max), got %d", len(entries))
	}
}
