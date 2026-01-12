package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestSTRB_TracksWriteSize verifies that STRB instruction tracks 1-byte write size
func TestSTRB_TracksWriteSize(t *testing.T) {
	// STRB R0, [R1] - store byte to R1
	v := vm.NewVM()
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STRB R0, [R1] (E5C10000)
	opcode := uint32(0xE5C10000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Verify the byte was written
	value, _ := v.Memory.ReadByteAt(0x20000)
	if value != 0x78 {
		t.Errorf("expected memory[0x20000]=0x78, got 0x%X", value)
	}

	// Verify write tracking
	if !v.HasMemoryWrite {
		t.Error("expected HasMemoryWrite=true")
	}
	if v.LastMemoryWrite != 0x20000 {
		t.Errorf("expected LastMemoryWrite=0x20000, got 0x%X", v.LastMemoryWrite)
	}

	// NEW: Verify write SIZE is tracked
	if v.LastMemoryWriteSize != 1 {
		t.Errorf("expected LastMemoryWriteSize=1 for STRB, got %d", v.LastMemoryWriteSize)
	}
}

// TestSTRH_TracksWriteSize verifies that STRH instruction tracks 2-byte write size
func TestSTRH_TracksWriteSize(t *testing.T) {
	// STRH R0, [R1] - store halfword to R1
	v := vm.NewVM()
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STRH R0, [R1] (E1C100B0) - store halfword, immediate offset 0
	opcode := uint32(0xE1C100B0)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Verify the halfword was written
	value, _ := v.Memory.ReadHalfword(0x20000)
	if value != 0x5678 {
		t.Errorf("expected memory[0x20000]=0x5678, got 0x%X", value)
	}

	// Verify write tracking
	if !v.HasMemoryWrite {
		t.Error("expected HasMemoryWrite=true")
	}
	if v.LastMemoryWrite != 0x20000 {
		t.Errorf("expected LastMemoryWrite=0x20000, got 0x%X", v.LastMemoryWrite)
	}

	// NEW: Verify write SIZE is tracked
	if v.LastMemoryWriteSize != 2 {
		t.Errorf("expected LastMemoryWriteSize=2 for STRH, got %d", v.LastMemoryWriteSize)
	}
}

// TestSTR_TracksWriteSize verifies that STR instruction tracks 4-byte write size
func TestSTR_TracksWriteSize(t *testing.T) {
	// STR R0, [R1] - store word to R1
	v := vm.NewVM()
	v.CPU.R[0] = 0xDEADBEEF
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STR R0, [R1] (E5810000)
	opcode := uint32(0xE5810000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Verify the word was written
	value, _ := v.Memory.ReadWord(0x20000)
	if value != 0xDEADBEEF {
		t.Errorf("expected memory[0x20000]=0xDEADBEEF, got 0x%X", value)
	}

	// Verify write tracking
	if !v.HasMemoryWrite {
		t.Error("expected HasMemoryWrite=true")
	}
	if v.LastMemoryWrite != 0x20000 {
		t.Errorf("expected LastMemoryWrite=0x20000, got 0x%X", v.LastMemoryWrite)
	}

	// NEW: Verify write SIZE is tracked
	if v.LastMemoryWriteSize != 4 {
		t.Errorf("expected LastMemoryWriteSize=4 for STR, got %d", v.LastMemoryWriteSize)
	}
}

// TestMemoryWriteSize_ClearedOnReset verifies write size is cleared on reset
func TestMemoryWriteSize_ClearedOnReset(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// Execute STRB to set write tracking
	opcode := uint32(0xE5C10000) // STRB R0, [R1]
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Verify write was tracked
	if !v.HasMemoryWrite || v.LastMemoryWriteSize != 1 {
		t.Fatal("write tracking not set up correctly")
	}

	// Reset VM
	v.Reset()

	// Verify write tracking is cleared
	if v.HasMemoryWrite {
		t.Error("expected HasMemoryWrite=false after reset")
	}
	if v.LastMemoryWriteSize != 0 {
		t.Errorf("expected LastMemoryWriteSize=0 after reset, got %d", v.LastMemoryWriteSize)
	}
}
