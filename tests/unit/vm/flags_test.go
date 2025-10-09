package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ================================================================================
// N Flag (Negative) Tests
// ================================================================================

func TestNFlag_Set_WhenResultNegative(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFFFFFF // -1 in two's complement
	v.CPU.PC = 0x8000

	// MOVS R1, R0 (set flags)
	opcode := uint32(0xE1B01000) // MOV with S bit
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.N {
		t.Error("N flag should be set when bit 31 is 1")
	}
}

func TestNFlag_Clear_WhenResultPositive(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x7FFFFFFF
	v.CPU.PC = 0x8000

	// MOVS R1, R0
	opcode := uint32(0xE1B01000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.N {
		t.Error("N flag should be clear when bit 31 is 0")
	}
}

func TestNFlag_AfterSubtraction(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 5
	v.CPU.R[1] = 10
	v.CPU.PC = 0x8000

	// SUBS R2, R0, R1 (5 - 10 = -5)
	opcode := uint32(0xE0502001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.N {
		t.Error("N flag should be set for negative result")
	}
}

func TestNFlag_AfterAddition(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x7FFFFFFF
	v.CPU.R[1] = 1
	v.CPU.PC = 0x8000

	// ADDS R2, R0, R1 (overflow to negative)
	opcode := uint32(0xE0902001) // Fixed: opcode 0x4 for ADD, not 0x5 (ADC)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.N {
		t.Error("N flag should be set when addition overflows to negative")
	}
}

func TestNFlag_AfterAND(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFFFFFF
	v.CPU.R[1] = 0x80000000
	v.CPU.PC = 0x8000

	// ANDS R2, R0, R1
	opcode := uint32(0xE0102001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.N {
		t.Error("N flag should be set when AND result has bit 31 set")
	}
}

// ================================================================================
// Z Flag (Zero) Tests
// ================================================================================

func TestZFlag_Set_WhenResultZero(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 10
	v.CPU.R[1] = 10
	v.CPU.PC = 0x8000

	// SUBS R2, R0, R1
	opcode := uint32(0xE0502001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.Z {
		t.Error("Z flag should be set when result is zero")
	}
}

func TestZFlag_Clear_WhenResultNonZero(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 10
	v.CPU.R[1] = 5
	v.CPU.PC = 0x8000

	// SUBS R2, R0, R1
	opcode := uint32(0xE0502001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.Z {
		t.Error("Z flag should be clear when result is non-zero")
	}
}

func TestZFlag_AfterAND_AllBitsClear(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0xAAAAAAAA
	v.CPU.R[1] = 0x55555555
	v.CPU.PC = 0x8000

	// ANDS R2, R0, R1 (alternating bits = 0)
	opcode := uint32(0xE0120001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.Z {
		t.Error("Z flag should be set when AND result is zero")
	}
}

func TestZFlag_AfterEOR_SameValue(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0x12345678
	v.CPU.PC = 0x8000

	// EORS R2, R0, R1 (same value XOR = 0)
	opcode := uint32(0xE0302001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.Z {
		t.Error("Z flag should be set when EOR of identical values is zero")
	}
}

func TestZFlag_AfterMOV_Zero(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MOVS R0, #0
	opcode := uint32(0xE3B00000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.Z {
		t.Error("Z flag should be set when moving zero")
	}
}

// ================================================================================
// C Flag (Carry) Tests - Addition
// ================================================================================

func TestCFlag_Set_OnUnsignedAdditionOverflow(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFFFFFF
	v.CPU.R[1] = 1
	v.CPU.PC = 0x8000

	// ADDS R2, R0, R1
	opcode := uint32(0xE0902001) // Fixed: opcode 0x4 for ADD
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.C {
		t.Error("C flag should be set on unsigned overflow")
	}
}

func TestCFlag_Clear_OnNoCarryAddition(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.R[1] = 200
	v.CPU.PC = 0x8000

	// ADDS R2, R0, R1
	opcode := uint32(0xE0902001) // Fixed: opcode 0x4 for ADD
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.C {
		t.Error("C flag should be clear when no carry occurs")
	}
}

func TestCFlag_AdditionMaxValues(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFFFFFF
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.PC = 0x8000

	// ADDS R2, R0, R1
	opcode := uint32(0xE0902001) // Fixed: opcode 0x4 for ADD
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.C {
		t.Error("C flag should be set when adding max values")
	}
}

func TestCFlag_ADC_WithCarryIn(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFFFFFF
	v.CPU.R[1] = 0
	v.CPU.CPSR.C = true // Carry in
	v.CPU.PC = 0x8000

	// ADCS R2, R0, R1
	opcode := uint32(0xE0B02001) // Fixed: opcode 0x5 for ADC, Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.C {
		t.Error("C flag should be set when ADC overflows")
	}
}

// ================================================================================
// C Flag (Carry) Tests - Subtraction
// ================================================================================

func TestCFlag_Subtraction_NoBorrow(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.R[1] = 50
	v.CPU.PC = 0x8000

	// SUBS R2, R0, R1
	opcode := uint32(0xE0502001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.C {
		t.Error("C flag should be set (no borrow) when minuend >= subtrahend")
	}
}

func TestCFlag_Subtraction_WithBorrow(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 50
	v.CPU.R[1] = 100
	v.CPU.PC = 0x8000

	// SUBS R2, R0, R1
	opcode := uint32(0xE0502001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.C {
		t.Error("C flag should be clear (borrow) when minuend < subtrahend")
	}
}

func TestCFlag_Subtraction_Equal(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.R[1] = 100
	v.CPU.PC = 0x8000

	// SUBS R2, R0, R1
	opcode := uint32(0xE0502001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.C {
		t.Error("C flag should be set (no borrow) when subtracting equal values")
	}
}

func TestCFlag_RSB_ReverseSub(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.R[1] = 50
	v.CPU.PC = 0x8000

	// RSBS R2, R0, R1 (50 - 100)
	opcode := uint32(0xE0702001) // Fixed: opcode 0x3 for RSB, Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.C {
		t.Error("C flag should be clear (borrow) in reverse subtraction")
	}
}

func TestCFlag_CMP_Greater(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 200
	v.CPU.R[1] = 100
	v.CPU.PC = 0x8000

	// CMP R0, R1
	opcode := uint32(0xE1500001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.C {
		t.Error("C flag should be set when comparing greater value")
	}
}

func TestCFlag_CMP_Less(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.R[1] = 200
	v.CPU.PC = 0x8000

	// CMP R0, R1
	opcode := uint32(0xE1500001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.C {
		t.Error("C flag should be clear when comparing lesser value")
	}
}

// ================================================================================
// C Flag (Carry) Tests - Shifts
// ================================================================================

func TestCFlag_LSL_CarryOut(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x80000000
	v.CPU.PC = 0x8000

	// MOVS R1, R0, LSL #1
	opcode := uint32(0xE1B01080)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.C {
		t.Error("C flag should be set from bit shifted out")
	}
}

func TestCFlag_LSR_CarryOut(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x00000001
	v.CPU.PC = 0x8000

	// MOVS R1, R0, LSR #1
	opcode := uint32(0xE1B010A0)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.C {
		t.Error("C flag should be set from bit shifted out")
	}
}

func TestCFlag_ASR_CarryOut(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x00000003
	v.CPU.PC = 0x8000

	// MOVS R1, R0, ASR #1
	opcode := uint32(0xE1B010C0)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.C {
		t.Error("C flag should be set from bit shifted out")
	}
}

func TestCFlag_ROR_CarryOut(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x00000001
	v.CPU.PC = 0x8000

	// MOVS R1, R0, ROR #1
	opcode := uint32(0xE1B010E0)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.C {
		t.Error("C flag should be set from bit rotated out")
	}
}

// ================================================================================
// V Flag (Overflow) Tests - Addition
// ================================================================================

func TestVFlag_Set_PositivePlusPositiveToNegative(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x7FFFFFFF // Max positive int32
	v.CPU.R[1] = 1
	v.CPU.PC = 0x8000

	// ADDS R2, R0, R1
	opcode := uint32(0xE0902001) // Fixed: opcode 0x4 for ADD
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.V {
		t.Error("V flag should be set: pos + pos = neg overflow")
	}
}

func TestVFlag_Set_NegativePlusNegativeToPositive(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x80000000 // Min negative int32
	v.CPU.R[1] = 0xFFFFFFFF // -1
	v.CPU.PC = 0x8000

	// ADDS R2, R0, R1
	opcode := uint32(0xE0902001) // Fixed: opcode 0x4 for ADD
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.V {
		t.Error("V flag should be set: neg + neg = pos overflow")
	}
}

func TestVFlag_Clear_PositivePlusPositive(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.R[1] = 200
	v.CPU.PC = 0x8000

	// ADDS R2, R0, R1
	opcode := uint32(0xE0902001) // Fixed: opcode 0x4 for ADD
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.V {
		t.Error("V flag should be clear: no signed overflow")
	}
}

func TestVFlag_Clear_PositivePlusNegative(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.R[1] = 0xFFFFFFFF // -1
	v.CPU.PC = 0x8000

	// ADDS R2, R0, R1
	opcode := uint32(0xE0902001) // Fixed: opcode 0x4 for ADD
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.V {
		t.Error("V flag should be clear: different signs cannot overflow in addition")
	}
}

// ================================================================================
// V Flag (Overflow) Tests - Subtraction
// ================================================================================

func TestVFlag_Subtraction_PositiveMinusNegativeToNegative(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x7FFFFFFF // Max positive
	v.CPU.R[1] = 0x80000000 // Min negative
	v.CPU.PC = 0x8000

	// SUBS R2, R0, R1
	opcode := uint32(0xE0502001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.V {
		t.Error("V flag should be set: pos - neg overflow")
	}
}

func TestVFlag_Subtraction_NegativeMinusPositiveToPositive(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x80000000 // Min negative
	v.CPU.R[1] = 1          // Positive
	v.CPU.PC = 0x8000

	// SUBS R2, R0, R1
	opcode := uint32(0xE0502001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.V {
		t.Error("V flag should be set: neg - pos overflow")
	}
}

func TestVFlag_Subtraction_NoOverflow(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.R[1] = 50
	v.CPU.PC = 0x8000

	// SUBS R2, R0, R1
	opcode := uint32(0xE0502001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.V {
		t.Error("V flag should be clear: no overflow")
	}
}

func TestVFlag_Subtraction_SameSign(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x7FFFFFFF
	v.CPU.R[1] = 100
	v.CPU.PC = 0x8000

	// SUBS R2, R0, R1
	opcode := uint32(0xE0502001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.V {
		t.Error("V flag should be clear: same sign subtraction cannot overflow")
	}
}

// ================================================================================
// Combined Flag Tests
// ================================================================================

func TestFlags_AllSet(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x80000000
	v.CPU.R[1] = 0x80000000
	v.CPU.PC = 0x8000

	// ADDS R2, R0, R1 (0x80000000 + 0x80000000 = 0, so N=0, Z=1, C=1, V=1)
	opcode := uint32(0xE0902001) // Fixed: opcode 0x4 for ADD
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.N {
		t.Error("N flag should be clear (result is 0)")
	}
	if !v.CPU.CPSR.Z {
		t.Error("Z flag should be set")
	}
	if !v.CPU.CPSR.C {
		t.Error("C flag should be set")
	}
	if !v.CPU.CPSR.V {
		t.Error("V flag should be set")
	}
}

func TestFlags_AllClear(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.R[1] = 50
	v.CPU.PC = 0x8000

	// ADDS R2, R0, R1 (simple addition, all flags clear)
	opcode := uint32(0xE0902001) // Fixed: opcode 0x4 for ADD
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.N {
		t.Error("N flag should be clear")
	}
	if v.CPU.CPSR.Z {
		t.Error("Z flag should be clear")
	}
	if v.CPU.CPSR.C {
		t.Error("C flag should be clear")
	}
	if v.CPU.CPSR.V {
		t.Error("V flag should be clear")
	}
}

func TestFlags_PreservedWithoutSBit(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = true
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = true
	v.CPU.R[0] = 10
	v.CPU.R[1] = 20
	v.CPU.PC = 0x8000

	// ADD R2, R0, R1 (without S bit)
	opcode := uint32(0xE0820001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.N || !v.CPU.CPSR.Z || !v.CPU.CPSR.C || !v.CPU.CPSR.V {
		t.Error("Flags should be preserved when S bit is not set")
	}
}

// ================================================================================
// Flag Tests with Logical Operations
// ================================================================================

func TestFlags_AND_NoCarryOrOverflow(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFFFFFF
	v.CPU.R[1] = 0x0000000F
	v.CPU.PC = 0x8000

	// ANDS R2, R0, R1
	opcode := uint32(0xE0102001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Logical operations don't affect C and V (they preserve them)
	// Only N and Z are affected
	if v.CPU.CPSR.N {
		t.Error("N flag should be clear")
	}
	if v.CPU.CPSR.Z {
		t.Error("Z flag should be clear")
	}
}

func TestFlags_ORR_SetN(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x80000000
	v.CPU.R[1] = 0x00000001
	v.CPU.PC = 0x8000

	// ORRS R2, R0, R1
	opcode := uint32(0xE1902001) // Fixed: Rn=0, Rd=2, Rm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.N {
		t.Error("N flag should be set")
	}
	if v.CPU.CPSR.Z {
		t.Error("Z flag should be clear")
	}
}

func TestFlags_MVN_InvertBits(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x00000000
	v.CPU.PC = 0x8000

	// MVNS R1, R0 (invert 0 = 0xFFFFFFFF)
	opcode := uint32(0xE1F01000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.N {
		t.Error("N flag should be set")
	}
	if v.CPU.CPSR.Z {
		t.Error("Z flag should be clear")
	}
	if v.CPU.R[1] != 0xFFFFFFFF {
		t.Errorf("R1 should be 0xFFFFFFFF, got 0x%X", v.CPU.R[1])
	}
}

// ================================================================================
// Edge Case Flag Tests
// ================================================================================

func TestFlags_MaxIntPlusOne(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x7FFFFFFF
	v.CPU.PC = 0x8000

	// ADDS R0, R0, #1
	opcode := uint32(0xE2B00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("Expected 0x80000000, got 0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.N {
		t.Error("N flag should be set")
	}
	if v.CPU.CPSR.Z {
		t.Error("Z flag should be clear")
	}
	if v.CPU.CPSR.C {
		t.Error("C flag should be clear (no unsigned overflow)")
	}
	if !v.CPU.CPSR.V {
		t.Error("V flag should be set (signed overflow)")
	}
}

func TestFlags_MinIntMinusOne(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x80000000
	v.CPU.PC = 0x8000

	// SUBS R0, R0, #1
	opcode := uint32(0xE2500001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x7FFFFFFF {
		t.Errorf("Expected 0x7FFFFFFF, got 0x%X", v.CPU.R[0])
	}
	if v.CPU.CPSR.N {
		t.Error("N flag should be clear")
	}
	if v.CPU.CPSR.Z {
		t.Error("Z flag should be clear")
	}
	if !v.CPU.CPSR.C {
		t.Error("C flag should be set (no borrow)")
	}
	if !v.CPU.CPSR.V {
		t.Error("V flag should be set (signed overflow)")
	}
}

func TestFlags_ZeroMinusOne(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// SUBS R0, R0, #1
	opcode := uint32(0xE2500001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("Expected 0xFFFFFFFF, got 0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.N {
		t.Error("N flag should be set")
	}
	if v.CPU.CPSR.Z {
		t.Error("Z flag should be clear")
	}
	if v.CPU.CPSR.C {
		t.Error("C flag should be clear (borrow occurred)")
	}
	if v.CPU.CPSR.V {
		t.Error("V flag should be clear (no signed overflow)")
	}
}
