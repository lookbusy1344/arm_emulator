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

// ============================================================================
// Priority 2: LDR addressing mode completeness
// ============================================================================

func TestLDR_RegisterOffset_Negative(t *testing.T) {
	// LDR R0, [R1, -R2] - load with negative register offset
	v := vm.NewVM()
	v.CPU.R[1] = 0x20010
	v.CPU.R[2] = 0x10
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x87654321)

	// LDR R0, [R1, -R2]
	// Format: cccc 011P UBWL nnnn dddd oooo oooo oooo
	// cond=E, I=1, P=1, U=0 (subtract), B=0, W=0, L=1
	// Rn=R1, Rd=R0, Rm=R2 (no shift)
	opcode := uint32(0xE7110002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x87654321 {
		t.Errorf("expected R0=0x87654321, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDR_ScaledRegisterOffset_LSL(t *testing.T) {
	// LDR R0, [R1, R2, LSL #2] - load with scaled register offset
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 3 // Will be shifted left by 2 = 12
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x2000C, 0xABCDEF12)

	// LDR R0, [R1, R2, LSL #2]
	// Format: cccc 011P UBWL nnnn dddd ssss s00t mmmm
	// shift amount=2 (bits 11:7), shift type=00 (LSL, bits 6:5), Rm=R2 (bits 3:0)
	// offset = (2 << 7) | (0 << 5) | 2 = 0x102
	opcode := uint32(0xE7910102)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xABCDEF12 {
		t.Errorf("expected R0=0xABCDEF12, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDR_ScaledRegisterOffset_LSR(t *testing.T) {
	// LDR R0, [R1, R2, LSR #2] - load with logical shift right
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 16 // Will be shifted right by 2 = 4
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20004, 0x11223344)

	// LDR R0, [R1, R2, LSR #2]
	// shift amount=2 (bits 11:7), shift type=01 (LSR, bits 6:5), Rm=R2
	// offset = (2 << 7) | (1 << 5) | 2 = 0x122
	opcode := uint32(0xE7910122)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x11223344 {
		t.Errorf("expected R0=0x11223344, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDR_ScaledRegisterOffset_ASR(t *testing.T) {
	// LDR R0, [R1, R2, ASR #2] - load with arithmetic shift right
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 32 // Will be shifted right by 2 = 8
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20008, 0x55667788)

	// LDR R0, [R1, R2, ASR #2]
	// shift amount=2 (bits 11:7), shift type=10 (ASR, bits 6:5), Rm=R2
	// offset = (2 << 7) | (2 << 5) | 2 = 0x142
	opcode := uint32(0xE7910142)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x55667788 {
		t.Errorf("expected R0=0x55667788, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDR_ScaledRegisterOffset_ROR(t *testing.T) {
	// LDR R0, [R1, R2, ROR #2] - load with rotate right
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 0x40000001 // Will be rotated right by 2 = 0x50000000
	v.CPU.PC = 0x8000

	// For this test, let's use a simpler value
	// R2 = 16, ROR #2 = 4
	v.CPU.R[2] = 16

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20004, 0x99AABBCC)

	// LDR R0, [R1, R2, ROR #2]
	// shift amount=2 (bits 11:7), shift type=11 (ROR, bits 6:5), Rm=R2
	// offset = (2 << 7) | (3 << 5) | 2 = 0x162
	opcode := uint32(0xE7910162)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x99AABBCC {
		t.Errorf("expected R0=0x99AABBCC, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDR_PreIndexedRegisterOffset(t *testing.T) {
	// LDR R0, [R1, R2]! - load with register offset and writeback
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 8
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20008, 0xFEDCBA98)

	// LDR R0, [R1, R2]!
	// Format: P=1, U=1, W=1 (writeback)
	opcode := uint32(0xE7B10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFEDCBA98 {
		t.Errorf("expected R0=0xFEDCBA98, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20008 {
		t.Errorf("expected R1=0x20008 (writeback), got R1=0x%X", v.CPU.R[1])
	}
}

func TestLDR_PreIndexedScaledOffset(t *testing.T) {
	// LDR R0, [R1, R2, LSL #2]! - load with scaled offset and writeback
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 2 // Will be shifted left by 2 = 8
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20008, 0x13579BDF)

	// LDR R0, [R1, R2, LSL #2]!
	// P=1, U=1, W=1, offset = (2 << 7) | (0 << 5) | 2
	opcode := uint32(0xE7B10102)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x13579BDF {
		t.Errorf("expected R0=0x13579BDF, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20008 {
		t.Errorf("expected R1=0x20008 (writeback), got R1=0x%X", v.CPU.R[1])
	}
}

func TestLDR_PostIndexedRegisterOffset(t *testing.T) {
	// LDR R0, [R1], R2 - load then update with register offset
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 12
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x2468ACE0)

	// LDR R0, [R1], R2
	// Format: P=0, U=1, W=0 (post-indexed always has W implicitly)
	opcode := uint32(0xE6910002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x2468ACE0 {
		t.Errorf("expected R0=0x2468ACE0, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x2000C {
		t.Errorf("expected R1=0x2000C (post-indexed), got R1=0x%X", v.CPU.R[1])
	}
}

func TestLDR_PostIndexedScaledOffset(t *testing.T) {
	// LDR R0, [R1], R2, LSL #1 - load then update with scaled offset
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 4 // Will be shifted left by 1 = 8
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x369CF258)

	// LDR R0, [R1], R2, LSL #1
	// P=0, U=1, W=0, offset = (1 << 7) | (0 << 5) | 2
	opcode := uint32(0xE6910082)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x369CF258 {
		t.Errorf("expected R0=0x369CF258, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20008 {
		t.Errorf("expected R1=0x20008 (post-indexed with scaled offset), got R1=0x%X", v.CPU.R[1])
	}
}

// ============================================================================
// Priority 2: STR addressing mode completeness
// ============================================================================

func TestSTR_PreIndexed(t *testing.T) {
	// STR R0, [R1, #8]! - store with writeback
	v := vm.NewVM()
	v.CPU.R[0] = 0xCAFEBABE
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STR R0, [R1, #8]!
	// Format: P=1, U=1, B=0, W=1, L=0, immediate offset=8
	opcode := uint32(0xE5A10008)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20008)
	if value != 0xCAFEBABE {
		t.Errorf("expected memory[0x20008]=0xCAFEBABE, got 0x%X", value)
	}
	if v.CPU.R[1] != 0x20008 {
		t.Errorf("expected R1=0x20008 (writeback), got R1=0x%X", v.CPU.R[1])
	}
}

func TestSTR_PostIndexed(t *testing.T) {
	// STR R0, [R1], #4 - store then update
	v := vm.NewVM()
	v.CPU.R[0] = 0x11223344
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STR R0, [R1], #4
	// Format: P=0, U=1, B=0, W=0, L=0
	opcode := uint32(0xE4810004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20000)
	if value != 0x11223344 {
		t.Errorf("expected memory[0x20000]=0x11223344, got 0x%X", value)
	}
	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (post-indexed), got R1=0x%X", v.CPU.R[1])
	}
}

func TestSTR_RegisterOffset(t *testing.T) {
	// STR R0, [R1, R2] - store with register offset
	v := vm.NewVM()
	v.CPU.R[0] = 0xDEADBEEF
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 12
	v.CPU.PC = 0x8000

	// STR R0, [R1, R2]
	// Format: I=1, P=1, U=1, B=0, W=0, L=0
	opcode := uint32(0xE7810002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x2000C)
	if value != 0xDEADBEEF {
		t.Errorf("expected memory[0x2000C]=0xDEADBEEF, got 0x%X", value)
	}
}

func TestSTR_RegisterOffset_Negative(t *testing.T) {
	// STR R0, [R1, -R2] - store with negative register offset
	v := vm.NewVM()
	v.CPU.R[0] = 0xABCDEF01
	v.CPU.R[1] = 0x20010
	v.CPU.R[2] = 0x10
	v.CPU.PC = 0x8000

	// STR R0, [R1, -R2]
	// Format: I=1, P=1, U=0 (subtract), B=0, W=0, L=0
	opcode := uint32(0xE7010002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20000)
	if value != 0xABCDEF01 {
		t.Errorf("expected memory[0x20000]=0xABCDEF01, got 0x%X", value)
	}
}

func TestSTR_ScaledRegisterOffset_LSL(t *testing.T) {
	// STR R0, [R1, R2, LSL #2] - store with scaled offset
	v := vm.NewVM()
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 4 // Will be shifted left by 2 = 16
	v.CPU.PC = 0x8000

	// STR R0, [R1, R2, LSL #2]
	// offset = (2 << 7) | (0 << 5) | 2 = 0x102
	opcode := uint32(0xE7810102)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20010)
	if value != 0x12345678 {
		t.Errorf("expected memory[0x20010]=0x12345678, got 0x%X", value)
	}
}

func TestSTR_ScaledRegisterOffset_LSR(t *testing.T) {
	// STR R0, [R1, R2, LSR #2] - store with logical shift right
	v := vm.NewVM()
	v.CPU.R[0] = 0x55AA55AA
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 16 // Will be shifted right by 2 = 4
	v.CPU.PC = 0x8000

	// STR R0, [R1, R2, LSR #2]
	// offset = (2 << 7) | (1 << 5) | 2 = 0x122
	opcode := uint32(0xE7810122)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20004)
	if value != 0x55AA55AA {
		t.Errorf("expected memory[0x20004]=0x55AA55AA, got 0x%X", value)
	}
}

func TestSTR_ScaledRegisterOffset_ASR(t *testing.T) {
	// STR R0, [R1, R2, ASR #2] - store with arithmetic shift right
	v := vm.NewVM()
	v.CPU.R[0] = 0xBBCCDDEE
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 32 // Will be shifted right by 2 = 8
	v.CPU.PC = 0x8000

	// STR R0, [R1, R2, ASR #2]
	// offset = (2 << 7) | (2 << 5) | 2 = 0x142
	opcode := uint32(0xE7810142)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20008)
	if value != 0xBBCCDDEE {
		t.Errorf("expected memory[0x20008]=0xBBCCDDEE, got 0x%X", value)
	}
}

func TestSTR_ScaledRegisterOffset_ROR(t *testing.T) {
	// STR R0, [R1, R2, ROR #2] - store with rotate right
	v := vm.NewVM()
	v.CPU.R[0] = 0xFF00FF00
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 16 // ROR #2 gives 4
	v.CPU.PC = 0x8000

	// STR R0, [R1, R2, ROR #2]
	// offset = (2 << 7) | (3 << 5) | 2 = 0x162
	opcode := uint32(0xE7810162)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20004)
	if value != 0xFF00FF00 {
		t.Errorf("expected memory[0x20004]=0xFF00FF00, got 0x%X", value)
	}
}

func TestSTR_PreIndexedRegisterOffset(t *testing.T) {
	// STR R0, [R1, R2]! - store with register offset and writeback
	v := vm.NewVM()
	v.CPU.R[0] = 0x98765432
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 8
	v.CPU.PC = 0x8000

	// STR R0, [R1, R2]!
	// Format: P=1, U=1, W=1
	opcode := uint32(0xE7A10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20008)
	if value != 0x98765432 {
		t.Errorf("expected memory[0x20008]=0x98765432, got 0x%X", value)
	}
	if v.CPU.R[1] != 0x20008 {
		t.Errorf("expected R1=0x20008 (writeback), got R1=0x%X", v.CPU.R[1])
	}
}

func TestSTR_PostIndexedRegisterOffset(t *testing.T) {
	// STR R0, [R1], R2 - store then update with register offset
	v := vm.NewVM()
	v.CPU.R[0] = 0x13572468
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 12
	v.CPU.PC = 0x8000

	// STR R0, [R1], R2
	// Format: P=0, U=1
	opcode := uint32(0xE6810002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20000)
	if value != 0x13572468 {
		t.Errorf("expected memory[0x20000]=0x13572468, got 0x%X", value)
	}
	if v.CPU.R[1] != 0x2000C {
		t.Errorf("expected R1=0x2000C (post-indexed), got R1=0x%X", v.CPU.R[1])
	}
}

// ============================================================================
// Priority 2: LDRB/STRB addressing mode completeness
// ============================================================================

func TestLDRB_ImmediateOffset_Negative(t *testing.T) {
	// LDRB R0, [R1, #-4] - load byte with negative offset
	v := vm.NewVM()
	v.CPU.R[1] = 0x20008
	v.CPU.PC = 0x8000

	v.Memory.WriteByteAt(0x20004, 0x7F)

	// LDRB R0, [R1, #-4]
	// Format: P=1, U=0 (subtract), B=1, W=0, L=1, offset=4
	opcode := uint32(0xE5510004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x7F {
		t.Errorf("expected R0=0x7F, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDRB_PreIndexed(t *testing.T) {
	// LDRB R0, [R1, #4]! - load byte with writeback
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	v.Memory.WriteByteAt(0x20004, 0x9A)

	// LDRB R0, [R1, #4]!
	// Format: P=1, U=1, B=1, W=1, L=1
	opcode := uint32(0xE5F10004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x9A {
		t.Errorf("expected R0=0x9A, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (writeback), got R1=0x%X", v.CPU.R[1])
	}
}

func TestLDRB_PostIndexed(t *testing.T) {
	// LDRB R0, [R1], #4 - load byte then update
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	v.Memory.WriteByteAt(0x20000, 0xBC)

	// LDRB R0, [R1], #4
	// Format: P=0, U=1, B=1, W=0, L=1
	opcode := uint32(0xE4D10004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xBC {
		t.Errorf("expected R0=0xBC, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (post-indexed), got R1=0x%X", v.CPU.R[1])
	}
}

func TestLDRB_RegisterOffset(t *testing.T) {
	// LDRB R0, [R1, R2] - load byte with register offset
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 7
	v.CPU.PC = 0x8000

	v.Memory.WriteByteAt(0x20007, 0xDE)

	// LDRB R0, [R1, R2]
	// Format: I=1, P=1, U=1, B=1, W=0, L=1
	opcode := uint32(0xE7D10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xDE {
		t.Errorf("expected R0=0xDE, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDRB_ScaledRegisterOffset(t *testing.T) {
	// LDRB R0, [R1, R2, LSL #2] - load byte with scaled offset
	v := vm.NewVM()
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 3 // Will be shifted left by 2 = 12
	v.CPU.PC = 0x8000

	v.Memory.WriteByteAt(0x2000C, 0xEF)

	// LDRB R0, [R1, R2, LSL #2]
	// offset = (2 << 7) | (0 << 5) | 2 = 0x102
	opcode := uint32(0xE7D10102)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xEF {
		t.Errorf("expected R0=0xEF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestSTRB_ImmediateOffset_Negative(t *testing.T) {
	// STRB R0, [R1, #-4] - store byte with negative offset
	v := vm.NewVM()
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0x20008
	v.CPU.PC = 0x8000

	// STRB R0, [R1, #-4]
	// Format: P=1, U=0 (subtract), B=1, W=0, L=0, offset=4
	opcode := uint32(0xE5410004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadByteAt(0x20004)
	if value != 0x78 {
		t.Errorf("expected memory[0x20004]=0x78, got 0x%X", value)
	}
}

func TestSTRB_PreIndexed(t *testing.T) {
	// STRB R0, [R1, #4]! - store byte with writeback
	v := vm.NewVM()
	v.CPU.R[0] = 0xAABBCCDD
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STRB R0, [R1, #4]!
	// Format: P=1, U=1, B=1, W=1, L=0
	opcode := uint32(0xE5E10004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadByteAt(0x20004)
	if value != 0xDD {
		t.Errorf("expected memory[0x20004]=0xDD, got 0x%X", value)
	}
	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (writeback), got R1=0x%X", v.CPU.R[1])
	}
}

func TestSTRB_PostIndexed(t *testing.T) {
	// STRB R0, [R1], #4 - store byte then update
	v := vm.NewVM()
	v.CPU.R[0] = 0xFEDCBA98
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	// STRB R0, [R1], #4
	// Format: P=0, U=1, B=1, W=0, L=0
	opcode := uint32(0xE4C10004)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadByteAt(0x20000)
	if value != 0x98 {
		t.Errorf("expected memory[0x20000]=0x98, got 0x%X", value)
	}
	if v.CPU.R[1] != 0x20004 {
		t.Errorf("expected R1=0x20004 (post-indexed), got R1=0x%X", v.CPU.R[1])
	}
}

func TestSTRB_RegisterOffset(t *testing.T) {
	// STRB R0, [R1, R2] - store byte with register offset
	v := vm.NewVM()
	v.CPU.R[0] = 0x11223344
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 5
	v.CPU.PC = 0x8000

	// STRB R0, [R1, R2]
	// Format: I=1, P=1, U=1, B=1, W=0, L=0
	opcode := uint32(0xE7C10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadByteAt(0x20005)
	if value != 0x44 {
		t.Errorf("expected memory[0x20005]=0x44, got 0x%X", value)
	}
}

func TestSTRB_ScaledRegisterOffset(t *testing.T) {
	// STRB R0, [R1, R2, LSL #2] - store byte with scaled offset
	v := vm.NewVM()
	v.CPU.R[0] = 0x55667788
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 2 // Will be shifted left by 2 = 8
	v.CPU.PC = 0x8000

	// STRB R0, [R1, R2, LSL #2]
	// offset = (2 << 7) | (0 << 5) | 2 = 0x102
	opcode := uint32(0xE7C10102)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadByteAt(0x20008)
	if value != 0x88 {
		t.Errorf("expected memory[0x20008]=0x88, got 0x%X", value)
	}
}

// ============================================================================
// Priority 2: STM/LDM addressing mode variants
// ============================================================================

func TestSTM_IB_IncrementBefore(t *testing.T) {
	// STMIB R0, {R1, R2} - store multiple, increment before
	v := vm.NewVM()
	v.CPU.R[0] = 0x20000
	v.CPU.R[1] = 0xAAAAAAAA
	v.CPU.R[2] = 0xBBBBBBBB
	v.CPU.PC = 0x8000

	// STMIB R0, {R1, R2}
	// Format: cccc 100P USWL nnnn rrrr rrrr rrrr rrrr
	// P=1 (increment before), U=1 (increment), S=0, W=0, L=0 (store)
	// Register list: bits 1,2 set = 0x06
	opcode := uint32(0xE9800006)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	val1, _ := v.Memory.ReadWord(0x20004)
	val2, _ := v.Memory.ReadWord(0x20008)

	if val1 != 0xAAAAAAAA {
		t.Errorf("expected memory[0x20004]=0xAAAAAAAA, got 0x%X", val1)
	}
	if val2 != 0xBBBBBBBB {
		t.Errorf("expected memory[0x20008]=0xBBBBBBBB, got 0x%X", val2)
	}
}

func TestSTM_DA_DecrementAfter(t *testing.T) {
	// STMDA R0, {R1, R2} - store multiple, decrement after
	// DA: stores start at (base - 4*n + 4), then increment through registers
	// With base=0x20008 and 2 regs: start at 0x20008 - 8 + 4 = 0x20004
	// R1 at 0x20004, R2 at 0x20008
	v := vm.NewVM()
	v.CPU.R[0] = 0x20008
	v.CPU.R[1] = 0x11111111
	v.CPU.R[2] = 0x22222222
	v.CPU.PC = 0x8000

	// STMDA R0, {R1, R2}
	// P=0 (decrement after), U=0 (decrement), S=0, W=0, L=0
	opcode := uint32(0xE8000006)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	val1, _ := v.Memory.ReadWord(0x20004)
	val2, _ := v.Memory.ReadWord(0x20008)

	if val1 != 0x11111111 {
		t.Errorf("expected memory[0x20004]=0x11111111, got 0x%X", val1)
	}
	if val2 != 0x22222222 {
		t.Errorf("expected memory[0x20008]=0x22222222, got 0x%X", val2)
	}
}

func TestSTM_DB_DecrementBefore(t *testing.T) {
	// STMDB R0, {R1, R2} - store multiple, decrement before (push)
	v := vm.NewVM()
	v.CPU.R[0] = 0x20008
	v.CPU.R[1] = 0x33333333
	v.CPU.R[2] = 0x44444444
	v.CPU.PC = 0x8000

	// STMDB R0, {R1, R2}
	// P=1 (decrement before), U=0 (decrement), S=0, W=0, L=0
	opcode := uint32(0xE9000006)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	val1, _ := v.Memory.ReadWord(0x20000)
	val2, _ := v.Memory.ReadWord(0x20004)

	if val1 != 0x33333333 {
		t.Errorf("expected memory[0x20000]=0x33333333, got 0x%X", val1)
	}
	if val2 != 0x44444444 {
		t.Errorf("expected memory[0x20004]=0x44444444, got 0x%X", val2)
	}
}

func TestSTM_WithWriteback(t *testing.T) {
	// STMIA R0!, {R1, R2} - store multiple with writeback
	v := vm.NewVM()
	v.CPU.R[0] = 0x20000
	v.CPU.R[1] = 0x55555555
	v.CPU.R[2] = 0x66666666
	v.CPU.PC = 0x8000

	// STMIA R0!, {R1, R2}
	// P=0, U=1, S=0, W=1 (writeback), L=0
	opcode := uint32(0xE8A00006)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	val1, _ := v.Memory.ReadWord(0x20000)
	val2, _ := v.Memory.ReadWord(0x20004)

	if val1 != 0x55555555 {
		t.Errorf("expected memory[0x20000]=0x55555555, got 0x%X", val1)
	}
	if val2 != 0x66666666 {
		t.Errorf("expected memory[0x20004]=0x66666666, got 0x%X", val2)
	}
	if v.CPU.R[0] != 0x20008 {
		t.Errorf("expected R0=0x20008 (writeback), got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDM_IB_IncrementBefore(t *testing.T) {
	// LDMIB R0, {R1, R2} - load multiple, increment before
	v := vm.NewVM()
	v.CPU.R[0] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20004, 0x77777777)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x20008, 0x88888888)

	// LDMIB R0, {R1, R2}
	// P=1, U=1, S=0, W=0, L=1 (load)
	opcode := uint32(0xE9900006)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[1] != 0x77777777 {
		t.Errorf("expected R1=0x77777777, got R1=0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0x88888888 {
		t.Errorf("expected R2=0x88888888, got R2=0x%X", v.CPU.R[2])
	}
}

func TestLDM_DB_DecrementBefore(t *testing.T) {
	// LDMDB R0, {R1, R2} - load multiple, decrement before (pop)
	v := vm.NewVM()
	v.CPU.R[0] = 0x20008
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0x99999999)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x20004, 0xAAAAAAAA)

	// LDMDB R0, {R1, R2}
	// P=1, U=0 (decrement), S=0, W=0, L=1
	opcode := uint32(0xE9100006)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[1] != 0x99999999 {
		t.Errorf("expected R1=0x99999999, got R1=0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0xAAAAAAAA {
		t.Errorf("expected R2=0xAAAAAAAA, got R2=0x%X", v.CPU.R[2])
	}
}
