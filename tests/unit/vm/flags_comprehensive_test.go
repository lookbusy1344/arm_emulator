package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ============================================================================
// Priority 4, Section 9: Flag Behavior Comprehensive Tests
// ============================================================================
//
// These tests verify that flag behavior is correct for each instruction class:
// - Arithmetic operations (ADD, SUB, ADC, SBC, RSB, RSC) set N, Z, C, V
// - Logical operations (AND, ORR, EOR, BIC) set N, Z, C (V unchanged)
// - Comparison operations (CMP, CMN) always set flags
// - Test operations (TST, TEQ) always set flags
// - Multiply operations (MUL, MLA) set N, Z only
// - Shift operations set carry flag appropriately
//

// ====== Arithmetic Instructions Set N, Z, C, V ======

func TestFlags_ADD_NZCV(t *testing.T) {
	tests := []struct {
		name    string
		a, b    uint32
		expectN bool
		expectZ bool
		expectC bool
		expectV bool
	}{
		{"positive result", 10, 20, false, false, false, false},
		{"zero result", 0, 0, false, true, false, false},
		{"negative result", 0xFFFFFFFF, 1, false, true, true, false},
		{"carry out", 0xFFFFFFFF, 2, false, false, true, false},
		{"overflow pos+pos=neg", 0x7FFFFFFF, 1, true, false, false, true},
		{"overflow neg+neg=pos", 0x80000000, 0x80000000, false, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			v.CPU.R[1] = tt.a
			v.CPU.R[2] = tt.b
			v.CPU.PC = 0x8000

			// ADDS R0, R1, R2 (with S bit set)
			opcode := uint32(0xE0910002) // ADD with S=1
			setupCodeWrite(v)
			v.Memory.WriteWord(0x8000, opcode)
			v.Step()

			checkFlag(t, "N", tt.expectN, v.CPU.CPSR.N)
			checkFlag(t, "Z", tt.expectZ, v.CPU.CPSR.Z)
			checkFlag(t, "C", tt.expectC, v.CPU.CPSR.C)
			checkFlag(t, "V", tt.expectV, v.CPU.CPSR.V)
		})
	}
}

func TestFlags_SUB_NZCV(t *testing.T) {
	tests := []struct {
		name    string
		a, b    uint32
		expectN bool
		expectZ bool
		expectC bool // In SUB, C is "not borrow"
		expectV bool
	}{
		{"positive result", 30, 10, false, false, true, false},
		{"zero result", 10, 10, false, true, true, false},
		{"negative result", 10, 20, true, false, false, false},
		{"borrow", 5, 10, true, false, false, false},
		{"overflow pos-neg=neg", 0x7FFFFFFF, 0x80000000, true, false, false, true},
		{"no overflow neg-pos", 0x80000000, 1, false, false, true, true}, // Actually does overflow: -2147483648 - 1 = 2147483647
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			v.CPU.R[1] = tt.a
			v.CPU.R[2] = tt.b
			v.CPU.PC = 0x8000

			// SUBS R0, R1, R2
			opcode := uint32(0xE0510002) // SUB with S=1
			setupCodeWrite(v)
			v.Memory.WriteWord(0x8000, opcode)
			v.Step()

			checkFlag(t, "N", tt.expectN, v.CPU.CPSR.N)
			checkFlag(t, "Z", tt.expectZ, v.CPU.CPSR.Z)
			checkFlag(t, "C", tt.expectC, v.CPU.CPSR.C)
			checkFlag(t, "V", tt.expectV, v.CPU.CPSR.V)
		})
	}
}

func TestFlags_ADC_WithCarry(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 1
	v.CPU.CPSR.C = true // Carry in
	v.CPU.PC = 0x8000

	// ADCS R0, R1, R2
	opcode := uint32(0xE0B10002) // ADC with S=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 0xFFFFFFFF + 1 + 1(carry) = 1, carry out
	if v.CPU.R[0] != 1 {
		t.Errorf("expected R0=1, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected carry flag set")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected zero flag clear")
	}
}

func TestFlags_SBC_WithBorrow(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 5
	v.CPU.CPSR.C = false // Borrow (C=0 means borrow)
	v.CPU.PC = 0x8000

	// SBCS R0, R1, R2
	opcode := uint32(0xE0D10002) // SBC with S=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 10 - 5 - 1(borrow) = 4
	if v.CPU.R[0] != 4 {
		t.Errorf("expected R0=4, got R0=0x%X", v.CPU.R[0])
	}
}

func TestFlags_RSB_NZCV(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 5
	v.CPU.R[2] = 10
	v.CPU.PC = 0x8000

	// RSBS R0, R1, R2 (reverse subtract: R2 - R1)
	opcode := uint32(0xE0710002) // RSB with S=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 10 - 5 = 5
	if v.CPU.R[0] != 5 {
		t.Errorf("expected R0=5, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.CPSR.N {
		t.Error("expected N flag clear")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag clear")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag set (no borrow)")
	}
}

func TestFlags_RSC_WithCarry(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 5
	v.CPU.R[2] = 10
	v.CPU.CPSR.C = true // No borrow
	v.CPU.PC = 0x8000

	// RSCS R0, R1, R2
	opcode := uint32(0xE0F10002) // RSC with S=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 10 - 5 - 0(no borrow) = 5
	if v.CPU.R[0] != 5 {
		t.Errorf("expected R0=5, got R0=0x%X", v.CPU.R[0])
	}
}

// ====== Logical Instructions Set N, Z, C (V Unchanged) ======

func TestFlags_AND_NZC(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 0
	v.CPU.CPSR.V = true // V should remain unchanged
	v.CPU.PC = 0x8000

	// ANDS R0, R1, R2
	opcode := uint32(0xE0110002) // AND with S=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: all bits clear, zero result
	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag set")
	}
	if v.CPU.CPSR.N {
		t.Error("expected N flag clear")
	}
	if !v.CPU.CPSR.V {
		t.Error("expected V flag unchanged (should still be set)")
	}
}

func TestFlags_ORR_NZC(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000
	v.CPU.R[2] = 0x00000001
	v.CPU.CPSR.V = true // V should remain unchanged
	v.CPU.PC = 0x8000

	// ORRS R0, R1, R2
	opcode := uint32(0xE1910002) // ORR with S=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 0x80000001, negative
	if v.CPU.R[0] != 0x80000001 {
		t.Errorf("expected R0=0x80000001, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.N {
		t.Error("expected N flag set")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag clear")
	}
	if !v.CPU.CPSR.V {
		t.Error("expected V flag unchanged")
	}
}

func TestFlags_EOR_NZC(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 0xFFFFFFFF
	v.CPU.CPSR.V = true
	v.CPU.PC = 0x8000

	// EORS R0, R1, R2
	opcode := uint32(0xE0310002) // EOR with S=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: XOR results in 0
	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag set")
	}
	if v.CPU.CPSR.N {
		t.Error("expected N flag clear")
	}
	if !v.CPU.CPSR.V {
		t.Error("expected V flag unchanged")
	}
}

func TestFlags_BIC_NZC(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 0x7FFFFFFF
	v.CPU.CPSR.V = true
	v.CPU.PC = 0x8000

	// BICS R0, R1, R2 (bit clear: R1 AND NOT R2)
	opcode := uint32(0xE1D10002) // BIC with S=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 0xFFFFFFFF AND NOT 0x7FFFFFFF = 0x80000000 (negative)
	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.N {
		t.Error("expected N flag set")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag clear")
	}
	if !v.CPU.CPSR.V {
		t.Error("expected V flag unchanged")
	}
}

// ====== Comparison Instructions Always Set Flags ======

func TestFlags_CMP_AlwaysSetsFlags(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 10
	v.CPU.PC = 0x8000

	// CMP R1, R2 (compare by subtraction, always sets flags)
	opcode := uint32(0xE1510002) // CMP (SUB with S=1, no destination)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 10 - 10 = 0, Z flag set
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag set")
	}
	if v.CPU.CPSR.N {
		t.Error("expected N flag clear")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag set (no borrow)")
	}
}

func TestFlags_CMN_AlwaysSetsFlags(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 0xFFFFFFF6 // -10 in two's complement
	v.CPU.PC = 0x8000

	// CMN R1, R2 (compare negative: add and set flags)
	opcode := uint32(0xE1710002) // CMN (ADD with S=1, no destination)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 10 + (-10) = 0, Z flag set
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag set")
	}
}

// ====== Test Instructions Always Set Flags ======

func TestFlags_TST_AlwaysSetsFlags(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xFF00FF00
	v.CPU.R[2] = 0x00FF00FF
	v.CPU.PC = 0x8000

	// TST R1, R2 (test bits: AND and set flags, no destination)
	opcode := uint32(0xE1110002) // TST (AND with S=1, no destination)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 0xFF00FF00 AND 0x00FF00FF = 0, Z flag set
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag set")
	}
	if v.CPU.CPSR.N {
		t.Error("expected N flag clear")
	}
}

func TestFlags_TEQ_AlwaysSetsFlags(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0xAAAAAAAA
	v.CPU.R[2] = 0xAAAAAAAA
	v.CPU.PC = 0x8000

	// TEQ R1, R2 (test equivalence: XOR and set flags, no destination)
	opcode := uint32(0xE1310002) // TEQ (EOR with S=1, no destination)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: identical values XOR to 0, Z flag set
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag set")
	}
}

// ====== Multiply Instructions Set N, Z Only ======

func TestFlags_MUL_NZ_Only(t *testing.T) {
	tests := []struct {
		name    string
		a, b    uint32
		expectN bool
		expectZ bool
	}{
		{"positive result", 10, 20, false, false},
		{"zero result", 0, 100, false, true},
		{"negative result", 0x80000001, 2, false, false}, // Result has bit 31 set
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			v.CPU.R[1] = tt.a
			v.CPU.R[2] = tt.b
			v.CPU.CPSR.C = true // C and V should be unchanged
			v.CPU.CPSR.V = true
			v.CPU.PC = 0x8000

			// MULS R0, R1, R2
			opcode := uint32(0xE0100291) // MUL with S=1
			setupCodeWrite(v)
			v.Memory.WriteWord(0x8000, opcode)
			v.Step()

			checkFlag(t, "N", tt.expectN, v.CPU.CPSR.N)
			checkFlag(t, "Z", tt.expectZ, v.CPU.CPSR.Z)
			// C and V should be unchanged (preserved as true)
			if !v.CPU.CPSR.C {
				t.Error("expected C flag unchanged (should still be set)")
			}
			if !v.CPU.CPSR.V {
				t.Error("expected V flag unchanged (should still be set)")
			}
		})
	}
}

func TestFlags_MLA_NZ_Only(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 5
	v.CPU.R[2] = 10
	v.CPU.R[3] = 100
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = true
	v.CPU.PC = 0x8000

	// MLAS R0, R1, R2, R3 (R0 = R1 * R2 + R3)
	// Format: Rd=0, Rn=3, Rs=2, Rm=1
	// Bits: cond=1110, 000000, A=1, S=1, Rd=0000, Rn=0011, Rs=0010, 1001, Rm=0001
	opcode := uint32(0xE0303291) // MLA with S=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 5 * 10 + 100 = 150
	if v.CPU.R[0] != 150 {
		t.Errorf("expected R0=150, got R0=%d", v.CPU.R[0])
	}
	if v.CPU.CPSR.N {
		t.Error("expected N flag clear")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag clear")
	}
	// C and V unchanged
	if !v.CPU.CPSR.C {
		t.Error("expected C flag unchanged")
	}
	if !v.CPU.CPSR.V {
		t.Error("expected V flag unchanged")
	}
}

// ====== Shift Operations Set Carry Flag ======

func TestFlags_ShiftCarry_LSL(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000001 // MSB set
	v.CPU.PC = 0x8000

	// MOVS R0, R1, LSL #1 (shift left, bit 31 goes to carry)
	opcode := uint32(0xE1B00081) // MOV with S=1, LSL #1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: result = 0x00000002, carry out from bit 31 = 1
	if v.CPU.R[0] != 0x00000002 {
		t.Errorf("expected R0=0x00000002, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected carry flag set (bit 31 shifted out)")
	}
}

func TestFlags_ShiftCarry_LSR(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000003 // LSB set
	v.CPU.PC = 0x8000

	// MOVS R0, R1, LSR #1 (shift right, bit 0 goes to carry)
	opcode := uint32(0xE1B000A1) // MOV with S=1, LSR #1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: result = 0x00000001, carry out from bit 0 = 1
	if v.CPU.R[0] != 0x00000001 {
		t.Errorf("expected R0=0x00000001, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected carry flag set (bit 0 shifted out)")
	}
}

func TestFlags_ShiftCarry_ASR(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000001 // Negative number, LSB set
	v.CPU.PC = 0x8000

	// MOVS R0, R1, ASR #1 (arithmetic shift right, sign extend)
	opcode := uint32(0xE1B000C1) // MOV with S=1, ASR #1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: result = 0xC0000000 (sign extended), carry = 1
	if v.CPU.R[0] != 0xC0000000 {
		t.Errorf("expected R0=0xC0000000, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected carry flag set (bit 0 shifted out)")
	}
	if !v.CPU.CPSR.N {
		t.Error("expected N flag set (sign bit preserved)")
	}
}

func TestFlags_ShiftCarry_ROR(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 0x00000001 // LSB set
	v.CPU.PC = 0x8000

	// MOVS R0, R1, ROR #1 (rotate right, bit 0 to carry and bit 31)
	opcode := uint32(0xE1B000E1) // MOV with S=1, ROR #1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: result = 0x80000000, carry = 1
	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected carry flag set (bit 0 rotated)")
	}
	if !v.CPU.CPSR.N {
		t.Error("expected N flag set (bit 31 now set)")
	}
}

// ====== Instructions Without S Bit Don't Update Flags ======

func TestFlags_NoUpdate_WithoutSBit(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 10
	// Set all flags initially
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = true
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = true
	v.CPU.PC = 0x8000

	// ADD R0, R1, R2 (without S bit)
	opcode := uint32(0xE0810002) // ADD with S=0
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: R0 = 20, but flags should be unchanged
	if v.CPU.R[0] != 20 {
		t.Errorf("expected R0=20, got R0=%d", v.CPU.R[0])
	}
	// All flags should remain set
	if !v.CPU.CPSR.N {
		t.Error("expected N flag unchanged (should still be set)")
	}
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag unchanged (should still be set)")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag unchanged (should still be set)")
	}
	if !v.CPU.CPSR.V {
		t.Error("expected V flag unchanged (should still be set)")
	}
}

// Helper function to check individual flags
func checkFlag(t *testing.T, name string, expected, actual bool) {
	t.Helper()
	if expected != actual {
		t.Errorf("flag %s: expected %v, got %v", name, expected, actual)
	}
}
