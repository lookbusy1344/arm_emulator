package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestB_Forward(t *testing.T) {
	// B forward (branch forward by 4 instructions)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// B +16 (EA000003) - offset of 3 words = 12 bytes, +8 for pipeline = 20
	opcode := uint32(0xEA000003)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// PC should be 0x8000 + 8 (pipeline) + 12 (offset*4) = 0x8014
	expected := uint32(0x8014)
	if v.CPU.PC != expected {
		t.Errorf("expected PC=0x%X, got PC=0x%X", expected, v.CPU.PC)
	}
}

func TestB_Backward(t *testing.T) {
	// B backward (branch back by 2 instructions)
	v := vm.NewVM()
	v.CPU.PC = 0x8010

	// B -8 (EAFFFFFE) - offset of -2 words
	opcode := uint32(0xEAFFFFFE)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8010, opcode)
	v.Step()

	// PC should be 0x8010 + 8 (pipeline) - 8 (offset*4) = 0x8010
	expected := uint32(0x8010)
	if v.CPU.PC != expected {
		t.Errorf("expected PC=0x%X, got PC=0x%X", expected, v.CPU.PC)
	}
}

func TestBL_BranchWithLink(t *testing.T) {
	// BL (branch with link) - should save return address in LR
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// BL +8 (EB000001)
	opcode := uint32(0xEB000001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// LR should contain return address (PC + 4)
	expectedLR := uint32(0x8004)
	if v.CPU.GetLR() != expectedLR {
		t.Errorf("expected LR=0x%X, got LR=0x%X", expectedLR, v.CPU.GetLR())
	}

	// PC should be updated
	expectedPC := uint32(0x800C) // 0x8000 + 8 + 4
	if v.CPU.PC != expectedPC {
		t.Errorf("expected PC=0x%X, got PC=0x%X", expectedPC, v.CPU.PC)
	}
}

func TestB_ConditionalEQ(t *testing.T) {
	// BEQ (branch if equal) - should only branch if Z flag is set
	v := vm.NewVM()
	v.CPU.PC = 0x8000
	v.CPU.CPSR.Z = true // Set Z flag

	// BEQ +4 (0A000000)
	opcode := uint32(0x0A000000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Should branch because Z is set
	expectedPC := uint32(0x8008)
	if v.CPU.PC != expectedPC {
		t.Errorf("expected PC=0x%X, got PC=0x%X", expectedPC, v.CPU.PC)
	}
}

func TestB_ConditionalNE_NotTaken(t *testing.T) {
	// BNE (branch if not equal) - should NOT branch if Z flag is set
	v := vm.NewVM()
	v.CPU.PC = 0x8000
	v.CPU.CPSR.Z = true // Set Z flag

	// BNE +4 (1A000000)
	opcode := uint32(0x1A000000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Should NOT branch because Z is set (condition fails)
	expectedPC := uint32(0x8004) // Just PC+4
	if v.CPU.PC != expectedPC {
		t.Errorf("expected PC=0x%X (not branched), got PC=0x%X", expectedPC, v.CPU.PC)
	}
}

func TestB_ConditionalGT(t *testing.T) {
	// BGT (branch if greater than) - Z==0 and N==V
	v := vm.NewVM()
	v.CPU.PC = 0x8000
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.N = false
	v.CPU.CPSR.V = false // N == V and Z == 0

	// BGT +8 (CA000001)
	opcode := uint32(0xCA000001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Should branch
	expectedPC := uint32(0x800C)
	if v.CPU.PC != expectedPC {
		t.Errorf("expected PC=0x%X, got PC=0x%X", expectedPC, v.CPU.PC)
	}
}

func TestB_LongOffset(t *testing.T) {
	// Test maximum positive offset
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// B with offset 0x7FFFFF (max positive 24-bit signed)
	// This is (EA7FFFFF)
	opcode := uint32(0xEA7FFFFF)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Offset is 0x7FFFFF << 2 = 0x1FFFFFC
	// PC = 0x8000 + 8 + 0x1FFFFFC = 0x2008004
	expectedPC := uint32(0x2008004)
	if v.CPU.PC != expectedPC {
		t.Errorf("expected PC=0x%X, got PC=0x%X", expectedPC, v.CPU.PC)
	}
}
