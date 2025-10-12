package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ============================================================================
// Priority 4, Section 8: Immediate Value Encoding Edge Cases
// ============================================================================
//
// ARM immediate values are encoded as an 8-bit value rotated right by
// a 4-bit rotation amount (in steps of 2).
// Format: imm = 8-bit value ROR (rotation * 2)
//
// This allows representing various常用 values compactly:
// - 0-255 (rotation = 0)
// - 0x100, 0x200, 0x400, 0x800 (rotation != 0)
// - 0xFF000000, 0x00FF0000, etc.
//

func TestImmediate_ZeroRotation(t *testing.T) {
	// Test simple values (0-255) with rotation = 0
	testCases := []struct {
		immediate uint32
		expected  uint32
	}{
		{0, 0},
		{1, 1},
		{42, 42},
		{128, 128},
		{255, 255},
	}

	for _, tc := range testCases {
		v := vm.NewVM()
		v.CPU.PC = 0x8000

		// MOV R0, #immediate
		// Opcode: MOV (1101), I=1, S=0, Rn=0, Rd=R0
		// Immediate encoding: rotation (bits 11-8) = 0, value (bits 7-0)
		opcode := uint32(0xE3A00000) | tc.immediate
		setupCodeWrite(v)
		v.Memory.WriteWord(0x8000, opcode)
		v.Step()

		if v.CPU.R[0] != tc.expected {
			t.Errorf("immediate %d: expected R0=%d, got R0=%d", tc.immediate, tc.expected, v.CPU.R[0])
		}
	}
}

func TestImmediate_CommonValues(t *testing.T) {
	// Test common powers of 2 and rotated values
	testCases := []struct {
		name     string
		opcode   uint32
		expected uint32
		comment  string
	}{
		{"0x100", 0xE3A00C01, 0x100, "1 ROR 24 = 0x100"},
		{"0x200", 0xE3A00C02, 0x200, "2 ROR 24 = 0x200"},
		{"0x400", 0xE3A00C04, 0x400, "4 ROR 24 = 0x400"},
		{"0x800", 0xE3A00C08, 0x800, "8 ROR 24 = 0x800"},
		{"0x1000", 0xE3A00A01, 0x1000, "1 ROR 20 = 0x1000"},
		{"0x10000", 0xE3A00801, 0x10000, "1 ROR 16 = 0x10000"},
		{"0xFF", 0xE3A000FF, 0xFF, "255 ROR 0 = 0xFF"},
		{"0xFF00", 0xE3A00CFF, 0xFF00, "255 ROR 24 = 0xFF00"},
		{"0xFF0000", 0xE3A008FF, 0xFF0000, "255 ROR 16 = 0xFF0000"},
		{"0xFF000000", 0xE3A004FF, 0xFF000000, "255 ROR 8 = 0xFF000000"},
	}

	for _, tc := range testCases {
		v := vm.NewVM()
		v.CPU.PC = 0x8000

		setupCodeWrite(v)
		v.Memory.WriteWord(0x8000, tc.opcode)
		v.Step()

		if v.CPU.R[0] != tc.expected {
			t.Errorf("%s: expected R0=0x%X, got R0=0x%X (%s)", tc.name, tc.expected, v.CPU.R[0], tc.comment)
		}
	}
}

func TestImmediate_AllRotations(t *testing.T) {
	// Test all 16 possible rotation values
	// Rotation is encoded in bits 11-8, actual rotation = rotation * 2
	for rotation := uint32(0); rotation < 16; rotation++ {
		v := vm.NewVM()
		v.CPU.PC = 0x8000

		// Use base value 0x80 (bit pattern: 10000000)
		// MOV R0, #(0x80 ROR (rotation * 2))
		immediate := uint32(0x80)
		rotateField := rotation << 8
		opcode := uint32(0xE3A00000) | rotateField | immediate

		setupCodeWrite(v)
		v.Memory.WriteWord(0x8000, opcode)
		v.Step()

		// Calculate expected value: 0x80 rotated right by (rotation * 2) bits
		rotateAmount := (rotation * 2) % 32
		expected := (immediate >> rotateAmount) | (immediate << (32 - rotateAmount))

		if v.CPU.R[0] != expected {
			t.Errorf("rotation %d: expected R0=0x%X, got R0=0x%X", rotation, expected, v.CPU.R[0])
		}
	}
}

func TestImmediate_MaxValue(t *testing.T) {
	// Test maximum 8-bit value (0xFF) with various rotations
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MOV R0, #0xFF (no rotation)
	opcode := uint32(0xE3A000FF)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFF {
		t.Errorf("expected R0=0xFF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestImmediate_InArithmetic(t *testing.T) {
	// Test that immediate values work correctly in arithmetic operations
	v := vm.NewVM()
	v.CPU.R[1] = 100
	v.CPU.PC = 0x8000

	// ADD R0, R1, #0x100
	// Use rotation to encode 0x100 (1 ROR 24)
	opcode := uint32(0xE2810C01) // ADD, I=1, Rn=R1, Rd=R0, rotation=12, imm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 100 + 0x100 = 356
	expected := uint32(100 + 0x100)
	if v.CPU.R[0] != expected {
		t.Errorf("expected R0=%d, got R0=%d", expected, v.CPU.R[0])
	}
}

func TestImmediate_NegativePattern(t *testing.T) {
	// Test patterns that create "negative-looking" values
	// These are actually positive but have the sign bit set
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MVN R0, #0 creates 0xFFFFFFFF (-1 in two's complement)
	// MVN (1111), I=1, S=0, Rn=0, Rd=R0, immediate=0
	opcode := uint32(0xE3E00000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestImmediate_BitwisePatterns(t *testing.T) {
	// Test useful bitwise patterns
	testCases := []struct {
		name     string
		opcode   uint32
		expected uint32
	}{
		{"alternate_bits", 0xE3A00CAA, 0xAA00},
		{"checkerboard", 0xE3A00055, 0x55},
		{"single_bit_high", 0xE3A00401, 0x1000000}, // 1 ROR 8 = 0x01000000
		{"all_but_one", 0xE3E00401, 0xFEFFFFFF},    // NOT(1 ROR 8) = 0xFEFFFFFF
	}

	for _, tc := range testCases {
		v := vm.NewVM()
		v.CPU.PC = 0x8000

		setupCodeWrite(v)
		v.Memory.WriteWord(0x8000, tc.opcode)
		v.Step()

		if v.CPU.R[0] != tc.expected {
			t.Errorf("%s: expected R0=0x%X, got R0=0x%X", tc.name, tc.expected, v.CPU.R[0])
		}
	}
}

func TestImmediate_EdgeRotations(t *testing.T) {
	// Test specific edge case rotations
	testCases := []struct {
		rotation uint32
		value    uint32
		expected uint32
	}{
		{0, 0xFF, 0xFF},         // No rotation
		{1, 0xFF, 0xC000003F},   // 0xFF ROR 2
		{2, 0xFF, 0xF000000F},   // 0xFF ROR 4
		{15, 0xFF, 0x000003FC},  // 0xFF ROR 30
		{8, 0x80, 0x00800000},   // 0x80 ROR 16
	}

	for _, tc := range testCases {
		v := vm.NewVM()
		v.CPU.PC = 0x8000

		// MOV R0, #(value ROR (rotation * 2))
		rotateField := tc.rotation << 8
		opcode := uint32(0xE3A00000) | rotateField | tc.value

		setupCodeWrite(v)
		v.Memory.WriteWord(0x8000, opcode)
		v.Step()

		if v.CPU.R[0] != tc.expected {
			t.Errorf("rotation=%d, value=0x%X: expected R0=0x%X, got R0=0x%X",
				tc.rotation, tc.value, tc.expected, v.CPU.R[0])
		}
	}
}

func TestImmediate_CompareOperations(t *testing.T) {
	// Test immediate values in comparison operations
	v := vm.NewVM()
	v.CPU.R[1] = 0x100
	v.CPU.PC = 0x8000

	// CMP R1, #0x100 (should set Z flag)
	// Use rotation to encode 0x100
	opcode := uint32(0xE3510C01) // CMP (1010), I=1, S=1, Rn=R1, rotation=12, imm=1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: R1 - 0x100 = 0, so Z flag should be set
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}
	if v.CPU.CPSR.N {
		t.Error("expected N flag to be clear")
	}
}

func TestImmediate_SubtractLarge(t *testing.T) {
	// Test subtraction with large immediate
	v := vm.NewVM()
	v.CPU.R[1] = 0xFF000000
	v.CPU.PC = 0x8000

	// SUB R0, R1, #0xFF000000
	// Use rotation=4 (ROR 8) to encode 0xFF000000
	opcode := uint32(0xE24104FF) // SUB, I=1, Rn=R1, Rd=R0, rotation=4, imm=0xFF
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: 0xFF000000 - 0xFF000000 = 0
	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
}
