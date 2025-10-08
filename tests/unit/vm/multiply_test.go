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
