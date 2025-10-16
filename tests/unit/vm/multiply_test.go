package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestMUL_Basic(t *testing.T) {
	// MUL R0, R1, R2 - R0 = R1 * R2
	v := vm.NewVM()
	v.CPU.R[1] = 5
	v.CPU.R[2] = 6
	v.CPU.PC = 0x8000

	// MUL R0, R1, R2 (E0000291)
	// Bits: cond=1110, 000000, S=0, Rd=0000, Rs=0010, 1001, Rm=0001
	opcode := uint32(0xE0000291)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 30 {
		t.Errorf("expected R0=30, got R0=%d", v.CPU.R[0])
	}
}

func TestMUL_WithFlags(t *testing.T) {
	// MULS R0, R1, R2 - multiply and set flags
	v := vm.NewVM()
	v.CPU.R[1] = 0
	v.CPU.R[2] = 5
	v.CPU.PC = 0x8000

	// MULS R0, R1, R2 (E0100291) - with S bit
	opcode := uint32(0xE0100291)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=%d", v.CPU.R[0])
	}

	// Should set Z flag
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}

	// N should be clear
	if v.CPU.CPSR.N {
		t.Error("expected N flag to be clear")
	}
}

func TestMUL_Negative(t *testing.T) {
	// MULS with negative result
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF // -1 in two's complement
	v.CPU.R[2] = 5
	v.CPU.PC = 0x8000

	// MULS R0, R1, R2 (E0100291)
	opcode := uint32(0xE0100291)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// -1 * 5 = -5 = 0xFFFFFFFB
	if v.CPU.R[0] != 0xFFFFFFFB {
		t.Errorf("expected R0=0xFFFFFFFB, got R0=0x%X", v.CPU.R[0])
	}

	// Should set N flag (negative result)
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set")
	}
}

func TestMLA_Basic(t *testing.T) {
	// MLA R0, R1, R2, R3 - R0 = R1 * R2 + R3
	v := vm.NewVM()
	v.CPU.R[1] = 3
	v.CPU.R[2] = 4
	v.CPU.R[3] = 10
	v.CPU.PC = 0x8000

	// MLA R0, R1, R2, R3 (E0203291)
	// Bits: cond=1110, 000000, A=1, S=0, Rd=0000, Rn=0011, Rs=0010, 1001, Rm=0001
	opcode := uint32(0xE0203291)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 3 * 4 + 10 = 22
	if v.CPU.R[0] != 22 {
		t.Errorf("expected R0=22, got R0=%d", v.CPU.R[0])
	}
}

func TestMLA_WithFlags(t *testing.T) {
	// MLAS R0, R1, R2, R3
	v := vm.NewVM()
	v.CPU.R[1] = 2
	v.CPU.R[2] = 3
	v.CPU.R[3] = 0xFFFFFFF6 // -10
	v.CPU.PC = 0x8000

	// MLAS R0, R1, R2, R3 (E0303291) - with S bit
	opcode := uint32(0xE0303291)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 2 * 3 + (-10) = 6 - 10 = -4 = 0xFFFFFFFC
	if v.CPU.R[0] != 0xFFFFFFFC {
		t.Errorf("expected R0=0xFFFFFFFC, got R0=0x%X", v.CPU.R[0])
	}

	// Should set N flag
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set")
	}
}

func TestMUL_Overflow(t *testing.T) {
	// Test that result is only lower 32 bits
	v := vm.NewVM()
	v.CPU.R[1] = 0x10000
	v.CPU.R[2] = 0x10000
	v.CPU.PC = 0x8000

	// MUL R0, R1, R2 (E0000291)
	opcode := uint32(0xE0000291)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x10000 * 0x10000 = 0x100000000, lower 32 bits = 0
	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0 (overflow), got R0=0x%X", v.CPU.R[0])
	}
}

func TestMUL_InvalidRegisters(t *testing.T) {
	// MUL R0, R0, R2 - Rd and Rm must be different
	v := vm.NewVM()
	v.CPU.R[0] = 5
	v.CPU.R[2] = 6
	v.CPU.PC = 0x8000

	// MUL R0, R0, R2 (E0000290) - Rd=Rm=R0
	opcode := uint32(0xE0000290)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	// Should fail with error
	if err == nil {
		t.Error("expected error when Rd == Rm")
	}
}

func TestMUL_LargeNumbers(t *testing.T) {
	// Test with large numbers
	v := vm.NewVM()
	v.CPU.R[1] = 1000
	v.CPU.R[2] = 2000
	v.CPU.PC = 0x8000

	// MUL R0, R1, R2 (E0000291)
	opcode := uint32(0xE0000291)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 2000000 {
		t.Errorf("expected R0=2000000, got R0=%d", v.CPU.R[0])
	}
}

func TestMLA_Zero(t *testing.T) {
	// MLA with zero multiplier
	v := vm.NewVM()
	v.CPU.R[1] = 0
	v.CPU.R[2] = 100
	v.CPU.R[3] = 50
	v.CPU.PC = 0x8000

	// MLA R0, R1, R2, R3 (E0203291)
	opcode := uint32(0xE0203291)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0 * 100 + 50 = 50
	if v.CPU.R[0] != 50 {
		t.Errorf("expected R0=50, got R0=%d", v.CPU.R[0])
	}
}

// Long Multiply Tests

func TestUMULL_Basic(t *testing.T) {
	// UMULL R0, R1, R2, R3 - R1:R0 = R2 * R3 (unsigned)
	v := vm.NewVM()
	v.CPU.R[2] = 0x10000
	v.CPU.R[3] = 0x10000
	v.CPU.PC = 0x8000

	// UMULL R0, R1, R2, R3 (E0C10392)
	// Bits: cond=1110, 0000100, U=1, A=0, S=0, RdHi=0001, RdLo=0000, Rs=0011, 1001, Rm=0010
	opcode := uint32(0xE0C10392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x10000 * 0x10000 = 0x100000000
	// Lo = 0x00000000, Hi = 0x00000001
	if v.CPU.R[0] != 0x00000000 {
		t.Errorf("expected R0=0x00000000, got R0=0x%08X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x00000001 {
		t.Errorf("expected R1=0x00000001, got R1=0x%08X", v.CPU.R[1])
	}
}

func TestUMULL_WithFlags(t *testing.T) {
	// UMULLS with zero result
	v := vm.NewVM()
	v.CPU.R[2] = 0
	v.CPU.R[3] = 0x12345678
	v.CPU.PC = 0x8000

	// UMULLS R0, R1, R2, R3 (E0D10392) - with S bit
	opcode := uint32(0xE0D10392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Result should be 0
	if v.CPU.R[0] != 0 || v.CPU.R[1] != 0 {
		t.Errorf("expected R1:R0=0:0, got R1:R0=0x%08X:0x%08X", v.CPU.R[1], v.CPU.R[0])
	}

	// Z flag should be set
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}

	// N flag should be clear
	if v.CPU.CPSR.N {
		t.Error("expected N flag to be clear")
	}
}

func TestUMULL_LargeNumbers(t *testing.T) {
	// UMULL with large unsigned numbers
	v := vm.NewVM()
	v.CPU.R[2] = 0xFFFFFFFF
	v.CPU.R[3] = 0xFFFFFFFF
	v.CPU.PC = 0x8000

	// UMULL R4, R5, R2, R3 (E0C54392)
	opcode := uint32(0xE0C54392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0xFFFFFFFF * 0xFFFFFFFF = 0xFFFFFFFE00000001
	// Lo = 0x00000001, Hi = 0xFFFFFFFE
	if v.CPU.R[4] != 0x00000001 {
		t.Errorf("expected R4=0x00000001, got R4=0x%08X", v.CPU.R[4])
	}
	if v.CPU.R[5] != 0xFFFFFFFE {
		t.Errorf("expected R5=0xFFFFFFFE, got R5=0x%08X", v.CPU.R[5])
	}
}

func TestUMLAL_Basic(t *testing.T) {
	// UMLAL R0, R1, R2, R3 - R1:R0 = R2 * R3 + R1:R0 (unsigned)
	v := vm.NewVM()
	v.CPU.R[0] = 10 // accumulator low
	v.CPU.R[1] = 0  // accumulator high
	v.CPU.R[2] = 5
	v.CPU.R[3] = 6
	v.CPU.PC = 0x8000

	// UMLAL R0, R1, R2, R3 (E0E10392)
	// Bits: cond=1110, 0000101, U=1, A=1, S=0, RdHi=0001, RdLo=0000, Rs=0011, 1001, Rm=0010
	opcode := uint32(0xE0E10392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 5 * 6 + 10 = 40
	if v.CPU.R[0] != 40 {
		t.Errorf("expected R0=40, got R0=%d", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0 {
		t.Errorf("expected R1=0, got R1=%d", v.CPU.R[1])
	}
}

func TestUMLAL_WithCarry(t *testing.T) {
	// UMLAL with carry into high word
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFFFFFF // accumulator low (max)
	v.CPU.R[1] = 0          // accumulator high
	v.CPU.R[2] = 2
	v.CPU.R[3] = 1
	v.CPU.PC = 0x8000

	// UMLAL R0, R1, R2, R3 (E0E10392)
	opcode := uint32(0xE0E10392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 2 * 1 + 0xFFFFFFFF = 0x100000001
	// Lo = 0x00000001, Hi = 0x00000001
	if v.CPU.R[0] != 0x00000001 {
		t.Errorf("expected R0=0x00000001, got R0=0x%08X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x00000001 {
		t.Errorf("expected R1=0x00000001, got R1=0x%08X", v.CPU.R[1])
	}
}

func TestSMULL_Positive(t *testing.T) {
	// SMULL R0, R1, R2, R3 - R1:R0 = R2 * R3 (signed)
	v := vm.NewVM()
	v.CPU.R[2] = 1000
	v.CPU.R[3] = 2000
	v.CPU.PC = 0x8000

	// SMULL R0, R1, R2, R3 (E0810392)
	// Bits: cond=1110, 0000100, U=0, A=0, S=0, RdHi=0001, RdLo=0000, Rs=0011, 1001, Rm=0010
	opcode := uint32(0xE0810392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 1000 * 2000 = 2000000
	if v.CPU.R[0] != 2000000 {
		t.Errorf("expected R0=2000000, got R0=%d", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0 {
		t.Errorf("expected R1=0, got R1=%d", v.CPU.R[1])
	}
}

func TestSMULL_Negative(t *testing.T) {
	// SMULL with negative numbers
	v := vm.NewVM()
	v.CPU.R[2] = 0xFFFFFFFF // -1
	v.CPU.R[3] = 1000
	v.CPU.PC = 0x8000

	// SMULL R0, R1, R2, R3 (E0810392)
	opcode := uint32(0xE0810392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// -1 * 1000 = -1000 = 0xFFFFFFFFFFFFFC18
	// Lo = 0xFFFFFC18, Hi = 0xFFFFFFFF
	if v.CPU.R[0] != 0xFFFFFC18 {
		t.Errorf("expected R0=0xFFFFFC18, got R0=0x%08X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0xFFFFFFFF {
		t.Errorf("expected R1=0xFFFFFFFF, got R1=0x%08X", v.CPU.R[1])
	}
}

func TestSMULL_BothNegative(t *testing.T) {
	// SMULL with both operands negative
	v := vm.NewVM()
	v.CPU.R[2] = 0xFFFFFFFF // -1
	v.CPU.R[3] = 0xFFFFFFFE // -2
	v.CPU.PC = 0x8000

	// SMULL R0, R1, R2, R3 (E0810392)
	opcode := uint32(0xE0810392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// -1 * -2 = 2
	if v.CPU.R[0] != 2 {
		t.Errorf("expected R0=2, got R0=%d", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0 {
		t.Errorf("expected R1=0, got R1=%d", v.CPU.R[1])
	}
}

func TestSMULL_WithFlags(t *testing.T) {
	// SMULLS with negative result
	v := vm.NewVM()
	v.CPU.R[2] = 0x80000000 // -2147483648 (most negative int32)
	v.CPU.R[3] = 2
	v.CPU.PC = 0x8000

	// SMULLS R0, R1, R2, R3 (E0910392) - with S bit
	opcode := uint32(0xE0910392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// -2147483648 * 2 = -4294967296 = 0xFFFFFFFF00000000
	// Lo = 0x00000000, Hi = 0xFFFFFFFF
	if v.CPU.R[0] != 0x00000000 {
		t.Errorf("expected R0=0x00000000, got R0=0x%08X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0xFFFFFFFF {
		t.Errorf("expected R1=0xFFFFFFFF, got R1=0x%08X", v.CPU.R[1])
	}

	// N flag should be set (bit 63 of result)
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set")
	}

	// Z flag should be clear
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag to be clear")
	}
}

func TestSMLAL_Basic(t *testing.T) {
	// SMLAL R0, R1, R2, R3 - R1:R0 = R2 * R3 + R1:R0 (signed)
	v := vm.NewVM()
	v.CPU.R[0] = 100 // accumulator low
	v.CPU.R[1] = 0   // accumulator high
	v.CPU.R[2] = 10
	v.CPU.R[3] = 20
	v.CPU.PC = 0x8000

	// SMLAL R0, R1, R2, R3 (E0A10392)
	// Bits: cond=1110, 0000101, U=0, A=1, S=0, RdHi=0001, RdLo=0000, Rs=0011, 1001, Rm=0010
	opcode := uint32(0xE0A10392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 10 * 20 + 100 = 300
	if v.CPU.R[0] != 300 {
		t.Errorf("expected R0=300, got R0=%d", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0 {
		t.Errorf("expected R1=0, got R1=%d", v.CPU.R[1])
	}
}

func TestSMLAL_NegativeAccumulator(t *testing.T) {
	// SMLAL with negative accumulator
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFFFF9C // -100 in lower 32 bits
	v.CPU.R[1] = 0xFFFFFFFF // sign extension
	v.CPU.R[2] = 10
	v.CPU.R[3] = 5
	v.CPU.PC = 0x8000

	// SMLAL R0, R1, R2, R3 (E0A10392)
	opcode := uint32(0xE0A10392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 10 * 5 + (-100) = -50 = 0xFFFFFFFFFFFFFFCE
	// Lo = 0xFFFFFFCE, Hi = 0xFFFFFFFF
	if v.CPU.R[0] != 0xFFFFFFCE {
		t.Errorf("expected R0=0xFFFFFFCE, got R0=0x%08X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0xFFFFFFFF {
		t.Errorf("expected R1=0xFFFFFFFF, got R1=0x%08X", v.CPU.R[1])
	}
}

func TestLongMultiply_InvalidRegisters(t *testing.T) {
	// UMULL R0, R0, R2, R3 - RdHi and RdLo must be different
	v := vm.NewVM()
	v.CPU.R[2] = 5
	v.CPU.R[3] = 6
	v.CPU.PC = 0x8000

	// UMULL R0, R0, R2, R3 (E0C00392) - RdHi=RdLo=R0
	opcode := uint32(0xE0C00392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	// Should fail with error
	if err == nil {
		t.Error("expected error when RdHi == RdLo")
	}
}

func TestLongMultiply_RdLoEqualsRm(t *testing.T) {
	// UMULL R2, R1, R2, R3 - RdLo and Rm must be different
	v := vm.NewVM()
	v.CPU.R[2] = 5
	v.CPU.R[3] = 6
	v.CPU.PC = 0x8000

	// UMULL R2, R1, R2, R3 (E0C12392) - RdLo=Rm=R2
	opcode := uint32(0xE0C12392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	// Should fail with error
	if err == nil {
		t.Error("expected error when RdLo == Rm")
	}
}

func TestLongMultiply_UsePC(t *testing.T) {
	// UMULL R0, R15, R2, R3 - R15 (PC) cannot be used
	v := vm.NewVM()
	v.CPU.R[2] = 5
	v.CPU.R[3] = 6
	v.CPU.PC = 0x8000

	// UMULL R0, R15, R2, R3 (E0CF0392) - RdHi=R15
	opcode := uint32(0xE0CF0392)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	// Should fail with error
	if err == nil {
		t.Error("expected error when using R15")
	}
}
