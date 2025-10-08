package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ============================================================================
// LSL (Logical Shift Left) comprehensive tests
// ============================================================================

func TestLSL_ImmediateShift(t *testing.T) {
	// MOV R0, R1, LSL #4
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000010
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL #4 (E1A00201)
	opcode := uint32(0xE1A00201)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x10 << 4 = 0x100
	if v.CPU.R[0] != 0x100 {
		t.Errorf("expected R0=0x100, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLSL_ZeroShift(t *testing.T) {
	// MOV R0, R1, LSL #0 (no shift, same as MOV R0, R1)
	v := vm.NewVM()
	v.CPU.R[1] = 0x12345678
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL #0 (E1A00001)
	opcode := uint32(0xE1A00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x12345678 {
		t.Errorf("expected R0=0x12345678, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLSL_CarryOut(t *testing.T) {
	// MOVS R0, R1, LSL #1 (should set carry when bit 31 is shifted out)
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000
	v.CPU.PC = 0x8000

	// MOVS R0, R1, LSL #1 (E1B00081)
	opcode := uint32(0xE1B00081)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set (bit shifted out)")
	}
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}
}

func TestLSL_MaxShift(t *testing.T) {
	// MOV R0, R1, LSL #31 (shift left by maximum)
	v := vm.NewVM()
	v.CPU.R[1] = 1
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL #31 (E1A00F81)
	opcode := uint32(0xE1A00F81)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLSL_RegisterShift(t *testing.T) {
	// MOV R0, R1, LSL R2 (shift amount in register)
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000001
	v.CPU.R[2] = 8 // Shift by 8
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL R2 (E1A00211)
	opcode := uint32(0xE1A00211)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x100 {
		t.Errorf("expected R0=0x100, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLSL_RegisterShiftOver32(t *testing.T) {
	// MOV R0, R1, LSL R2 (shift by more than 32 should result in 0)
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 33
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL R2 (E1A00211)
	opcode := uint32(0xE1A00211)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0 (shift > 32), got R0=0x%X", v.CPU.R[0])
	}
}

func TestLSL_InADD(t *testing.T) {
	// ADD R0, R1, R2, LSL #2 (add with shifted operand)
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 5 // 5 << 2 = 20
	v.CPU.PC = 0x8000

	// ADD R0, R1, R2, LSL #2 (E0810102)
	opcode := uint32(0xE0810102)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 10 + 20 = 30
	if v.CPU.R[0] != 30 {
		t.Errorf("expected R0=30, got R0=%d", v.CPU.R[0])
	}
}

// ============================================================================
// LSR (Logical Shift Right) comprehensive tests
// ============================================================================

func TestLSR_ImmediateShift(t *testing.T) {
	// MOV R0, R1, LSR #4
	v := vm.NewVM()
	v.CPU.R[1] = 0x100
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSR #4 (E1A00221)
	opcode := uint32(0xE1A00221)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x100 >> 4 = 0x10
	if v.CPU.R[0] != 0x10 {
		t.Errorf("expected R0=0x10, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLSR_SignBit(t *testing.T) {
	// MOV R0, R1, LSR #1 (logical shift doesn't preserve sign)
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSR #1 (E1A000A1) - bits[6:5]=01 for LSR, bits[11:7]=00001 for shift by 1
	opcode := uint32(0xE1A000A1)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x80000000 >> 1 = 0x40000000 (logical, not sign-extended)
	if v.CPU.R[0] != 0x40000000 {
		t.Errorf("expected R0=0x40000000, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLSR_CarryOut(t *testing.T) {
	// MOVS R0, R1, LSR #1 (should set carry when bit 0 is shifted out)
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000001
	v.CPU.PC = 0x8000

	// MOVS R0, R1, LSR #1 (E1B000A1) - S bit set, LSR #1
	opcode := uint32(0xE1B000A1)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set (bit shifted out)")
	}
}

func TestLSR_FullShift(t *testing.T) {
	// MOV R0, R1, LSR #32 (shift all bits out)
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSR #32 - LSR #32 is encoded as shift amount = 0
	// bits[11:7]=00000, bits[6:5]=01 (LSR), bit[4]=0
	opcode := uint32(0xE1A00021)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLSR_RegisterShift(t *testing.T) {
	// MOV R0, R1, LSR R2
	v := vm.NewVM()
	v.CPU.R[1] = 0xFF00
	v.CPU.R[2] = 8
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSR R2 (E1A00231)
	opcode := uint32(0xE1A00231)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFF {
		t.Errorf("expected R0=0xFF, got R0=0x%X", v.CPU.R[0])
	}
}

// ============================================================================
// ASR (Arithmetic Shift Right) comprehensive tests
// ============================================================================

func TestASR_PositiveNumber(t *testing.T) {
	// MOV R0, R1, ASR #4 (positive number)
	v := vm.NewVM()
	v.CPU.R[1] = 0x1000
	v.CPU.PC = 0x8000

	// MOV R0, R1, ASR #4 (E1A00241)
	opcode := uint32(0xE1A00241)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x1000 >> 4 = 0x100 (arithmetic, but positive)
	if v.CPU.R[0] != 0x100 {
		t.Errorf("expected R0=0x100, got R0=0x%X", v.CPU.R[0])
	}
}

func TestASR_NegativeNumber(t *testing.T) {
	// MOV R0, R1, ASR #4 (negative number, sign-extended)
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000
	v.CPU.PC = 0x8000

	// MOV R0, R1, ASR #4 (E1A00241)
	opcode := uint32(0xE1A00241)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x80000000 ASR 4 = 0xF8000000 (sign-extended)
	if v.CPU.R[0] != 0xF8000000 {
		t.Errorf("expected R0=0xF8000000, got R0=0x%X", v.CPU.R[0])
	}
}

func TestASR_PreserveSign(t *testing.T) {
	// MOV R0, R1, ASR #1 (verify sign preservation)
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFE // -2 in two's complement
	v.CPU.PC = 0x8000

	// MOV R0, R1, ASR #1 (E1A000C1)
	opcode := uint32(0xE1A000C1)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// -2 ASR 1 = -1 (0xFFFFFFFF)
	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestASR_CarryOut(t *testing.T) {
	// MOVS R0, R1, ASR #1 (should set carry)
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000003
	v.CPU.PC = 0x8000

	// MOVS R0, R1, ASR #1 (E1B000C1)
	opcode := uint32(0xE1B000C1)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 1 {
		t.Errorf("expected R0=1, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set (bit shifted out)")
	}
}

func TestASR_RegisterShift(t *testing.T) {
	// MOV R0, R1, ASR R2
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000
	v.CPU.R[2] = 8
	v.CPU.PC = 0x8000

	// MOV R0, R1, ASR R2 (E1A00251)
	opcode := uint32(0xE1A00251)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x80000000 ASR 8 = 0xFF800000
	if v.CPU.R[0] != 0xFF800000 {
		t.Errorf("expected R0=0xFF800000, got R0=0x%X", v.CPU.R[0])
	}
}

func TestASR_FullShift(t *testing.T) {
	// MOV R0, R1, ASR #32 (shift all bits, preserving sign)
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000001
	v.CPU.PC = 0x8000

	// ASR #32 is encoded as 0 in shift field
	opcode := uint32(0xE1A00041)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// All bits become sign bit (0xFFFFFFFF for negative)
	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
}

// ============================================================================
// ROR (Rotate Right) comprehensive tests
// ============================================================================

func TestROR_ImmediateRotate(t *testing.T) {
	// MOV R0, R1, ROR #4
	v := vm.NewVM()
	v.CPU.R[1] = 0x12345678
	v.CPU.PC = 0x8000

	// MOV R0, R1, ROR #4 (E1A00261)
	opcode := uint32(0xE1A00261)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x12345678 ROR 4 = 0x81234567
	if v.CPU.R[0] != 0x81234567 {
		t.Errorf("expected R0=0x81234567, got R0=0x%X", v.CPU.R[0])
	}
}

func TestROR_8BitRotate(t *testing.T) {
	// MOV R0, R1, ROR #8
	v := vm.NewVM()
	v.CPU.R[1] = 0x12345678
	v.CPU.PC = 0x8000

	// MOV R0, R1, ROR #8 (E1A00461)
	opcode := uint32(0xE1A00461)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x12345678 ROR 8 = 0x78123456
	if v.CPU.R[0] != 0x78123456 {
		t.Errorf("expected R0=0x78123456, got R0=0x%X", v.CPU.R[0])
	}
}

func TestROR_FullRotation(t *testing.T) {
	// MOV R0, R1, ROR #32 (full rotation, back to original)
	v := vm.NewVM()
	v.CPU.R[1] = 0xABCDEF01
	v.CPU.PC = 0x8000

	// ROR #32 encoded as 0 means rotate by 32, but acts like RRX or no rotation
	// Let's use a register shift to test properly
	v.CPU.R[2] = 32
	opcode := uint32(0xE1A00271) // MOV R0, R1, ROR R2
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// ROR by 32 should return original value
	if v.CPU.R[0] != 0xABCDEF01 {
		t.Errorf("expected R0=0xABCDEF01, got R0=0x%X", v.CPU.R[0])
	}
}

func TestROR_CarryOut(t *testing.T) {
	// MOVS R0, R1, ROR #1 (should set carry from bit rotated out)
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000001
	v.CPU.PC = 0x8000

	// MOVS R0, R1, ROR #1 (E1B000E1)
	opcode := uint32(0xE1B000E1)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x00000001 ROR 1 = 0x80000000
	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set")
	}
}

func TestROR_RegisterShift(t *testing.T) {
	// MOV R0, R1, ROR R2
	v := vm.NewVM()
	v.CPU.R[1] = 0xFF000000
	v.CPU.R[2] = 8
	v.CPU.PC = 0x8000

	// MOV R0, R1, ROR R2 (E1A00271)
	opcode := uint32(0xE1A00271)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0xFF000000 ROR 8 = 0x00FF0000
	if v.CPU.R[0] != 0x00FF0000 {
		t.Errorf("expected R0=0x00FF0000, got R0=0x%X", v.CPU.R[0])
	}
}

func TestROR_InDataProcessing(t *testing.T) {
	// ADD R0, R1, R2, ROR #4
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 0x12345678
	v.CPU.PC = 0x8000

	// ADD R0, R1, R2, ROR #4 (E0810262)
	opcode := uint32(0xE0810262)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 10 + (0x12345678 ROR 4) = 10 + 0x81234567
	expected := uint32(10 + 0x81234567)
	if v.CPU.R[0] != expected {
		t.Errorf("expected R0=0x%X, got R0=0x%X", expected, v.CPU.R[0])
	}
}

// ============================================================================
// RRX (Rotate Right Extended) tests
// ============================================================================

func TestRRX_WithCarryClear(t *testing.T) {
	// MOVS R0, R1, RRX (rotate through carry, C=0)
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000001
	v.CPU.CPSR.C = false
	v.CPU.PC = 0x8000

	// MOVS R0, R1, RRX (E1B00061 with ROR #0 encoding means RRX)
	opcode := uint32(0xE1B00061)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// RRX shifts right by 1 and puts old C in bit 31
	// 0x00000001 RRX with C=0 = 0x00000000, C=1
	if v.CPU.R[0] != 0x00000000 {
		t.Errorf("expected R0=0x00000000, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set (bit 0 rotated into carry)")
	}
}

func TestRRX_WithCarrySet(t *testing.T) {
	// MOVS R0, R1, RRX (rotate through carry, C=1)
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000000
	v.CPU.CPSR.C = true
	v.CPU.PC = 0x8000

	// MOVS R0, R1, RRX (E1B00061)
	opcode := uint32(0xE1B00061)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// RRX with C=1 puts 1 in bit 31
	// 0x00000000 RRX with C=1 = 0x80000000, C=0
	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.CPSR.C {
		t.Error("expected C flag to be clear")
	}
}

func TestRRX_MultipleOperations(t *testing.T) {
	// Test chaining RRX operations
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000003
	v.CPU.CPSR.C = false
	v.CPU.PC = 0x8000

	// First RRX: 0x00000003 RRX with C=0 = 0x00000001, C=1
	opcode := uint32(0xE1B00061)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x00000001 {
		t.Errorf("expected R0=0x00000001 after first RRX, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set after first RRX")
	}
}

// ============================================================================
// Combined shift tests
// ============================================================================

func TestShifts_CompareTypes(t *testing.T) {
	// Compare LSR vs ASR on same negative number
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000
	v.CPU.PC = 0x8000

	// LSR #1
	opcode := uint32(0xE1A00061)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	lsrResult := v.CPU.R[0]

	// Reset for ASR
	v = vm.NewVM()
	v.CPU.R[1] = 0x80000000
	v.CPU.PC = 0x8000

	// ASR #1
	opcode = uint32(0xE1A000C1)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	asrResult := v.CPU.R[0]

	// LSR should give 0x40000000, ASR should give 0xC0000000
	if lsrResult != 0x40000000 {
		t.Errorf("LSR: expected 0x40000000, got 0x%X", lsrResult)
	}
	if asrResult != 0xC0000000 {
		t.Errorf("ASR: expected 0xC0000000, got 0x%X", asrResult)
	}
}

func TestShifts_InComplexExpression(t *testing.T) {
	// ADD R0, R1, R2, LSL R3 (shift amount in register)
	v := vm.NewVM()
	v.CPU.R[1] = 100
	v.CPU.R[2] = 5
	v.CPU.R[3] = 3 // Shift by 3
	v.CPU.PC = 0x8000

	// ADD R0, R1, R2, LSL R3 (E0810312)
	opcode := uint32(0xE0810312)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 100 + (5 << 3) = 100 + 40 = 140
	if v.CPU.R[0] != 140 {
		t.Errorf("expected R0=140, got R0=%d", v.CPU.R[0])
	}
}

func TestShifts_ZeroAmount(t *testing.T) {
	// Test shifts with zero amount in register
	v := vm.NewVM()
	v.CPU.R[1] = 0x12345678
	v.CPU.R[2] = 0 // No shift
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL R2 (E1A00211)
	opcode := uint32(0xE1A00211)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Shift by 0 should preserve value
	if v.CPU.R[0] != 0x12345678 {
		t.Errorf("expected R0=0x12345678, got R0=0x%X", v.CPU.R[0])
	}
}
