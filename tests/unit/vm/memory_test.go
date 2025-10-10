package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestLDR_ImmediateOffset(t *testing.T) {
	// LDR R0, [R1, #4] - load word from R1+4
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000 // Data segment
	v.CPU.PC = 0x8000

	// Write test data to memory
	setupCodeWrite(v)
	v.Memory.WriteWord(0x20004, 0x12345678)

	// LDR R0, [R1, #4] (E5910004)
	opcode := uint32(0xE5910004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x12345678 {
		t.Errorf("expected R0=0x12345678, got R0=0x%X", v.CPU.R[0])
	}
}

func TestSTR_ImmediateOffset(t *testing.T) {
	// STR R0, [R1, #4] - store word to R1+4
	v := vm.NewVM()
	v.CPU.R[0] = 0xDEADBEEF
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STR R0, [R1, #4] (E5810004)
	opcode := uint32(0xE5810004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20004)
	if value != 0xDEADBEEF {
		t.Errorf("expected memory[0x20004]=0xDEADBEEF, got 0x%X", value)
	}
}

func TestLDRB_LoadByte(t *testing.T) {
	// LDRB R0, [R1] - load byte from R1
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// Write test data
	v.Memory.WriteByteAt(0x20000, 0xAB)

	// LDRB R0, [R1] (E5D10000)
	opcode := uint32(0xE5D10000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xAB {
		t.Errorf("expected R0=0xAB, got R0=0x%X", v.CPU.R[0])
	}
}

func TestSTRB_StoreByte(t *testing.T) {
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

	value, _ := v.Memory.ReadByteAt(0x20000)
	if value != 0x78 {
		t.Errorf("expected memory[0x20000]=0x78, got 0x%X", value)
	}
}

func TestLDR_PreIndexed(t *testing.T) {
	// LDR R0, [R1, #4]! - load word and update R1
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20004, 0xCAFEBABE)

	// LDR R0, [R1, #4]! (E5B10004) - pre-indexed with writeback
	opcode := uint32(0xE5B10004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xCAFEBABE {
		t.Errorf("expected R0=0xCAFEBABE, got R0=0x%X", v.CPU.R[0])
	}

	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (updated), got R1=0x%X", v.CPU.R[1])
	}
}

func TestLDR_PostIndexed(t *testing.T) {
	// LDR R0, [R1], #4 - load word then update R1
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x11223344)

	// LDR R0, [R1], #4 (E4910004) - post-indexed
	opcode := uint32(0xE4910004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x11223344 {
		t.Errorf("expected R0=0x11223344, got R0=0x%X", v.CPU.R[0])
	}

	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (updated), got R1=0x%X", v.CPU.R[1])
	}
}

func TestLDM_MultipleRegisters(t *testing.T) {
	// LDMIA R0, {R1, R2, R3} - load multiple registers
	v := vm.NewVM()
	v.CPU.R[0] = 0x20000
	v.CPU.PC = 0x8000

	// Write test data
	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x11111111)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x20004, 0x22222222)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x20008, 0x33333333)

	// LDMIA R0, {R1, R2, R3} (E890000E)
	// Register list: bits 1,2,3 set = 0x0E
	opcode := uint32(0xE890000E)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[1] != 0x11111111 {
		t.Errorf("expected R1=0x11111111, got R1=0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0x22222222 {
		t.Errorf("expected R2=0x22222222, got R2=0x%X", v.CPU.R[2])
	}
	if v.CPU.R[3] != 0x33333333 {
		t.Errorf("expected R3=0x33333333, got R3=0x%X", v.CPU.R[3])
	}
}

func TestSTM_MultipleRegisters(t *testing.T) {
	// STMIA R0, {R1, R2} - store multiple registers
	v := vm.NewVM()
	v.CPU.R[0] = 0x20000
	v.CPU.R[1] = 0xAAAAAAAA
	v.CPU.R[2] = 0xBBBBBBBB
	v.CPU.PC = 0x8000

	// STMIA R0, {R1, R2} (E8800006)
	// Register list: bits 1,2 set = 0x06
	opcode := uint32(0xE8800006)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	val1, _ := v.Memory.ReadWord(0x20000)
	val2, _ := v.Memory.ReadWord(0x20004)

	if val1 != 0xAAAAAAAA {
		t.Errorf("expected memory[0x20000]=0xAAAAAAAA, got 0x%X", val1)
	}
	if val2 != 0xBBBBBBBB {
		t.Errorf("expected memory[0x20004]=0xBBBBBBBB, got 0x%X", val2)
	}
}

func TestLDM_WithWriteback(t *testing.T) {
	// LDMIA R0!, {R1, R2} - load and update R0
	v := vm.NewVM()
	v.CPU.R[0] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x11111111)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x20004, 0x22222222)

	// LDMIA R0!, {R1, R2} (E8B00006) - with writeback
	opcode := uint32(0xE8B00006)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[1] != 0x11111111 {
		t.Errorf("expected R1=0x11111111, got R1=0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0x22222222 {
		t.Errorf("expected R2=0x22222222, got R2=0x%X", v.CPU.R[2])
	}
	if v.CPU.R[0] != 0x20008 {
		t.Errorf("expected R0=0x20008 (updated by 8 bytes), got R0=0x%X", v.CPU.R[0])
	}
}

func TestMemory_Alignment(t *testing.T) {
	// Test that unaligned word access fails
	v := vm.NewVM()
	v.Memory.StrictAlign = true

	// Try to read word from unaligned address
	_, err := v.Memory.ReadWord(0x20001)
	if err == nil {
		t.Error("expected error for unaligned word access")
	}
}

func TestMemory_Bounds(t *testing.T) {
	// Test that out-of-bounds access fails
	v := vm.NewVM()

	// Try to read from unmapped memory
	_, err := v.Memory.ReadWord(0xFFFFFFFF)
	if err == nil {
		t.Error("expected error for out-of-bounds access")
	}
}
