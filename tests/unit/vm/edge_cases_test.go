package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ============================================================================
// Edge Case Tests for Data Processing Instructions
// ============================================================================

// Test zero operands
func TestEdge_ZeroOperands(t *testing.T) {
	// ADD R0, R1, R2 where both are zero
	v := vm.NewVM()
	v.CPU.R[1] = 0
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	opcode := uint32(0xE0810002) // ADD R0, R1, R2
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// Test maximum positive value (INT32_MAX)
func TestEdge_MaxPositive(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x7FFFFFFF
	v.CPU.PC = 0x8000

	// MOV R0, R1
	opcode := uint32(0xE1A00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x7FFFFFFF {
		t.Errorf("expected R0=0x7FFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
}

// Test maximum negative value (INT32_MIN)
func TestEdge_MaxNegative(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000
	v.CPU.PC = 0x8000

	// MOV R0, R1
	opcode := uint32(0xE1A00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}
}

// Test signed overflow (positive + positive = negative)
func TestEdge_SignedOverflowPosPos(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x7FFFFFFF // INT32_MAX
	v.CPU.R[2] = 1
	v.CPU.PC = 0x8000

	// ADDS R0, R1, R2
	opcode := uint32(0xE0B10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.V {
		t.Error("expected V flag set for signed overflow")
	}
}

// Test signed overflow (negative + negative = positive)
func TestEdge_SignedOverflowNegNeg(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000 // INT32_MIN
	v.CPU.R[2] = 0xFFFFFFFF // -1
	v.CPU.PC = 0x8000

	// ADDS R0, R1, R2
	opcode := uint32(0xE0B10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x7FFFFFFF {
		t.Errorf("expected R0=0x7FFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.V {
		t.Error("expected V flag set for signed overflow")
	}
}

// Test unsigned overflow (carry out)
func TestEdge_UnsignedOverflow(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 1
	v.CPU.PC = 0x8000

	// ADDS R0, R1, R2
	opcode := uint32(0xE0B10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag set for unsigned overflow")
	}
}

// Test underflow
func TestEdge_Underflow(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0
	v.CPU.R[2] = 1
	v.CPU.PC = 0x8000

	// SUBS R0, R1, R2
	opcode := uint32(0xE0510002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0 - 1 = -1 (0xFFFFFFFF)
	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.CPSR.C {
		t.Error("expected C flag clear (borrow occurred)")
	}
}

// Test subtraction with signed underflow
func TestEdge_SignedUnderflow(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000 // INT32_MIN
	v.CPU.R[2] = 1
	v.CPU.PC = 0x8000

	// SUBS R0, R1, R2 (INT32_MIN - 1 should overflow)
	opcode := uint32(0xE0510002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x7FFFFFFF {
		t.Errorf("expected R0=0x7FFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.V {
		t.Error("expected V flag set for signed underflow")
	}
}

// Test all ones
func TestEdge_AllOnes(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.PC = 0x8000

	// MOV R0, R1
	opcode := uint32(0xE1A00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
}

// Test alternating bit patterns
func TestEdge_AlternatingBits1(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xAAAAAAAA
	v.CPU.R[2] = 0x55555555
	v.CPU.PC = 0x8000

	// ORR R0, R1, R2 (should result in all ones)
	opcode := uint32(0xE1810002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestEdge_AlternatingBits2(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xAAAAAAAA
	v.CPU.R[2] = 0x55555555
	v.CPU.PC = 0x8000

	// AND R0, R1, R2 (should result in zero)
	opcode := uint32(0xE0010002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
}

// Test single bit operations
func TestEdge_SingleBitSet(t *testing.T) {
	// Test each bit position
	for i := uint32(0); i < 32; i++ {
		v := vm.NewVM()
		v.CPU.R[1] = 1 << i
		v.CPU.PC = 0x8000

		// MOV R0, R1
		opcode := uint32(0xE1A00001)
		setupCodeWrite(v)
		v.Memory.WriteWord(0x8000, opcode)
		v.Step()

		expected := uint32(1 << i)
		if v.CPU.R[0] != expected {
			t.Errorf("bit %d: expected R0=0x%X, got R0=0x%X", i, expected, v.CPU.R[0])
		}
	}
}

// Test register aliases (SP, LR, PC)
func TestEdge_StackPointer(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[13] = 0x10000 // SP
	v.CPU.PC = 0x8000

	// MOV R0, R13 (MOV R0, SP)
	opcode := uint32(0xE1A0000D)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x10000 {
		t.Errorf("expected R0=0x10000, got R0=0x%X", v.CPU.R[0])
	}
}

func TestEdge_LinkRegister(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[14] = 0x8100 // LR
	v.CPU.PC = 0x8000

	// MOV R0, R14 (MOV R0, LR)
	opcode := uint32(0xE1A0000E)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x8100 {
		t.Errorf("expected R0=0x8100, got R0=0x%X", v.CPU.R[0])
	}
}

func TestEdge_ProgramCounter(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MOV R0, R15 (MOV R0, PC)
	// PC reads as current instruction + 8
	opcode := uint32(0xE1A0000F)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// PC should be current + 8 (pipeline effect)
	expected := uint32(0x8000 + 8)
	if v.CPU.R[0] != expected {
		t.Errorf("expected R0=0x%X (PC+8), got R0=0x%X", expected, v.CPU.R[0])
	}
}

// Test boundary values for shifts
func TestEdge_ShiftByZero(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x12345678
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL R2 (shift by 0)
	opcode := uint32(0xE1A00211)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x12345678 {
		t.Errorf("expected R0=0x12345678, got R0=0x%X", v.CPU.R[0])
	}
}

func TestEdge_ShiftBy32(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 32
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL R2 (shift by 32 = 0)
	opcode := uint32(0xE1A00211)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0 (shift by 32), got R0=0x%X", v.CPU.R[0])
	}
}

func TestEdge_ShiftGreaterThan32(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 100
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL R2 (shift > 32 = 0)
	opcode := uint32(0xE1A00211)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0 (shift > 32), got R0=0x%X", v.CPU.R[0])
	}
}

// Test flag combinations
func TestEdge_AllFlagsSet(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = true
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = true
	v.CPU.R[1] = 42
	v.CPU.PC = 0x8000

	// MOV R0, R1 (should not affect flags)
	opcode := uint32(0xE1A00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.N || !v.CPU.CPSR.Z || !v.CPU.CPSR.C || !v.CPU.CPSR.V {
		t.Error("MOV without S should not affect flags")
	}
}

func TestEdge_AllFlagsClear(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = false
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.C = false
	v.CPU.CPSR.V = false
	v.CPU.R[1] = 42
	v.CPU.PC = 0x8000

	// MOV R0, R1 (should not affect flags)
	opcode := uint32(0xE1A00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.N || v.CPU.CPSR.Z || v.CPU.CPSR.C || v.CPU.CPSR.V {
		t.Error("MOV without S should not affect flags")
	}
}

// Test operations that should not write results
func TestEdge_CMP_NoWrite(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 42
	v.CPU.R[1] = 10
	v.CPU.PC = 0x8000

	// CMP R0, R1 (should not modify R0)
	opcode := uint32(0xE1500001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("CMP should not modify R0, expected 42, got %d", v.CPU.R[0])
	}
}

func TestEdge_TST_NoWrite(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0xFF
	v.CPU.R[1] = 0xAA
	v.CPU.PC = 0x8000

	// TST R0, R1 (should not modify R0)
	opcode := uint32(0xE1100001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFF {
		t.Errorf("TST should not modify R0, expected 0xFF, got 0x%X", v.CPU.R[0])
	}
}

// Test same source and destination register
func TestEdge_SameRegisterSrcDst(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 10
	v.CPU.PC = 0x8000

	// ADD R0, R0, R0 (R0 = R0 + R0)
	opcode := uint32(0xE0800000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 20 {
		t.Errorf("expected R0=20, got R0=%d", v.CPU.R[0])
	}
}

func TestEdge_TripleSameRegister(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[5] = 3
	v.CPU.PC = 0x8000

	// ADD R5, R5, R5 (R5 = R5 + R5)
	opcode := uint32(0xE0855005)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[5] != 6 {
		t.Errorf("expected R5=6, got R5=%d", v.CPU.R[5])
	}
}

// Test immediate value rotations (ARM immediate encoding)
func TestEdge_ImmediateRotation(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MOV R0, #0xFF000000 (requires rotation)
	// This is encoded as rotate right of 0xFF by 8 positions
	// Encoded as: rotation=4 (4*2=8), value=0xFF
	opcode := uint32(0xE3A004FF) // MOV R0, #0xFF000000
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFF000000 {
		t.Errorf("expected R0=0xFF000000, got R0=0x%X", v.CPU.R[0])
	}
}

// Test RSB with zero (negation)
func TestEdge_RSB_Negate(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 42
	v.CPU.PC = 0x8000

	// RSB R0, R1, #0 (R0 = 0 - R1, negate R1)
	opcode := uint32(0xE2610000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// -42 in two's complement
	expected := uint32(^uint32(42) + 1)
	if v.CPU.R[0] != expected {
		t.Errorf("expected R0=0x%X (-42), got R0=0x%X", expected, v.CPU.R[0])
	}
}

// Test MVN of all zeros
func TestEdge_MVN_Zeros(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0
	v.CPU.PC = 0x8000

	// MVN R0, R1
	opcode := uint32(0xE1E00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
}

// Test BIC clearing all bits
func TestEdge_BIC_ClearAll(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 0xFFFFFFFF
	v.CPU.PC = 0x8000

	// BIC R0, R1, R2 (clear all bits)
	opcode := uint32(0xE1C10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
}

// Test carry propagation through ADC
func TestEdge_ADC_CarryPropagation(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 0
	v.CPU.CPSR.C = true
	v.CPU.PC = 0x8000

	// ADCS R0, R1, R2 (0xFFFFFFFF + 0 + 1 = 0 with carry)
	opcode := uint32(0xE0B10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected carry flag to be set")
	}
	if !v.CPU.CPSR.Z {
		t.Error("expected zero flag to be set")
	}
}

// Test SBC borrow chain
func TestEdge_SBC_BorrowChain(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0
	v.CPU.R[2] = 0
	v.CPU.CPSR.C = false // Borrow from previous operation
	v.CPU.PC = 0x8000

	// SBCS R0, R1, R2 (0 - 0 - 1 = -1)
	opcode := uint32(0xE0D10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
}
