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

// ============================================================================
// LDRH (Load Halfword) instruction tests - ARM2a extension
// ============================================================================

func TestLDRH_ImmediateOffset(t *testing.T) {
	// LDRH R0, [R1, #4] - load halfword from R1+4
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// Write test data to memory
	setupCodeWrite(v)
	v.Memory.WriteHalfword(0x20004, 0xABCD)

	// LDRH R0, [R1, #4] - opcode pattern for halfword load
	// Bits: cond=1110, 000P=0001, U=1, B=0, W=0, L=1, Rn=0001, Rd=0000, offset=0100, 1011, offset=0100
	// Format: 1110 000P UBWL Rn Rd offsetH 1011 offsetL
	// Pre-indexed (P=1), Add offset (U=1), No writeback (W=0), Load (L=1)
	opcode := uint32(0xE1D100B4) // LDRH R0, [R1, #4]
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xABCD {
		t.Errorf("expected R0=0xABCD, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDRH_PreIndexed(t *testing.T) {
	// LDRH R0, [R1, #4]! - load halfword and update R1
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteHalfword(0x20004, 0x1234)

	// LDRH R0, [R1, #4]! - with writeback (W=1)
	opcode := uint32(0xE1F100B4)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x1234 {
		t.Errorf("expected R0=0x1234, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (updated), got R1=0x%X", v.CPU.R[1])
	}
}

func TestLDRH_PostIndexed(t *testing.T) {
	// LDRH R0, [R1], #4 - load halfword then update R1
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteHalfword(0x20000, 0x5678)

	// LDRH R0, [R1], #4 - post-indexed (P=0)
	opcode := uint32(0xE0D100B4)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x5678 {
		t.Errorf("expected R0=0x5678, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (updated), got R1=0x%X", v.CPU.R[1])
	}
}

func TestLDRH_RegisterOffset(t *testing.T) {
	// LDRH R0, [R1, R2] - load halfword with register offset
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 6
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteHalfword(0x20006, 0x9ABC)

	// LDRH R0, [R1, R2] - register offset
	opcode := uint32(0xE19100B2)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x9ABC {
		t.Errorf("expected R0=0x9ABC, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDRH_NegativeOffset(t *testing.T) {
	// LDRH R0, [R1, #-4] - load halfword with negative offset
	v := vm.NewVM()
	v.CPU.R[1] = 0x20008
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteHalfword(0x20004, 0xFEDC)

	// LDRH R0, [R1, #-4] - subtract offset (U=0)
	opcode := uint32(0xE15100B4)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFEDC {
		t.Errorf("expected R0=0xFEDC, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDRH_ZeroExtend(t *testing.T) {
	// Verify LDRH zero-extends the loaded value
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFFFFFF // Pre-fill with all 1s
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteHalfword(0x20000, 0x00FF)

	// LDRH R0, [R1]
	opcode := uint32(0xE1D100B0)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Should be zero-extended to 0x000000FF, not sign-extended
	if v.CPU.R[0] != 0x000000FF {
		t.Errorf("expected R0=0x000000FF (zero-extended), got R0=0x%X", v.CPU.R[0])
	}
}

// ============================================================================
// STRH (Store Halfword) instruction tests - ARM2a extension
// ============================================================================

func TestSTRH_ImmediateOffset(t *testing.T) {
	// STRH R0, [R1, #4] - store halfword to R1+4
	v := vm.NewVM()
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STRH R0, [R1, #4] - store only lower 16 bits
	opcode := uint32(0xE1C100B4)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadHalfword(0x20004)
	if value != 0x5678 {
		t.Errorf("expected memory[0x20004]=0x5678, got 0x%X", value)
	}
}

func TestSTRH_PreIndexed(t *testing.T) {
	// STRH R0, [R1, #4]! - store halfword and update R1
	v := vm.NewVM()
	v.CPU.R[0] = 0xABCDEF01
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STRH R0, [R1, #4]! - with writeback (W=1)
	opcode := uint32(0xE1E100B4)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadHalfword(0x20004)
	if value != 0xEF01 {
		t.Errorf("expected memory[0x20004]=0xEF01, got 0x%X", value)
	}
	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (updated), got R1=0x%X", v.CPU.R[1])
	}
}

func TestSTRH_PostIndexed(t *testing.T) {
	// STRH R0, [R1], #4 - store halfword then update R1
	v := vm.NewVM()
	v.CPU.R[0] = 0x11223344
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STRH R0, [R1], #4 - post-indexed (P=0)
	opcode := uint32(0xE0C100B4)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadHalfword(0x20000)
	if value != 0x3344 {
		t.Errorf("expected memory[0x20000]=0x3344, got 0x%X", value)
	}
	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (updated), got R1=0x%X", v.CPU.R[1])
	}
}

func TestSTRH_RegisterOffset(t *testing.T) {
	// STRH R0, [R1, R2] - store halfword with register offset
	v := vm.NewVM()
	v.CPU.R[0] = 0x9999AAAA
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 8
	v.CPU.PC = 0x8000

	// STRH R0, [R1, R2]
	opcode := uint32(0xE18100B2)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadHalfword(0x20008)
	if value != 0xAAAA {
		t.Errorf("expected memory[0x20008]=0xAAAA, got 0x%X", value)
	}
}

func TestSTRH_NegativeOffset(t *testing.T) {
	// STRH R0, [R1, #-4] - store halfword with negative offset
	v := vm.NewVM()
	v.CPU.R[0] = 0xBBBBCCCC
	v.CPU.R[1] = 0x20008
	v.CPU.PC = 0x8000

	// STRH R0, [R1, #-4] - subtract offset (U=0)
	opcode := uint32(0xE14100B4)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadHalfword(0x20004)
	if value != 0xCCCC {
		t.Errorf("expected memory[0x20004]=0xCCCC, got 0x%X", value)
	}
}

func TestSTRH_TruncateUpper16Bits(t *testing.T) {
	// Verify STRH only stores lower 16 bits
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFF0000
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STRH R0, [R1]
	opcode := uint32(0xE1C100B0)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadHalfword(0x20000)
	if value != 0x0000 {
		t.Errorf("expected memory[0x20000]=0x0000 (lower 16 bits), got 0x%X", value)
	}
}
