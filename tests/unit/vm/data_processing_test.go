package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// Helper function to create a VM and execute an instruction
func executeInstruction(t *testing.T, opcode uint32) *vm.VM {
	t.Helper()
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// Code segment needs write permission for tests
	for _, seg := range v.Memory.Segments {
		if seg.Name == "code" {
			seg.Permissions = vm.PermRead | vm.PermWrite | vm.PermExecute
		}
	}

	// Write instruction to memory
	setupCodeWrite(v)
	err := v.Memory.WriteWord(0x8000, opcode)
	if err != nil {
		t.Fatalf("failed to write instruction: %v", err)
	}

	// Execute one step
	err = v.Step()
	if err != nil {
		t.Fatalf("failed to execute instruction: %v", err)
	}

	return v
}

func TestMOV_Immediate(t *testing.T) {
	// MOV R0, #42 (E3A0002A)
	// Condition: AL (1110), Opcode: MOV (1101), I=1, S=0
	// Rd=0, rotation=0, immediate=42
	opcode := uint32(0xE3A0002A)
	v := executeInstruction(t, opcode)

	if v.CPU.R[0] != 42 {
		t.Errorf("expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestMOV_Register(t *testing.T) {
	// MOV R1, R0
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.PC = 0x8000

	// MOV R1, R0 (E1A01000)
	opcode := uint32(0xE1A01000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[1] != 100 {
		t.Errorf("expected R1=100, got R1=%d", v.CPU.R[1])
	}
}

func TestADD_Immediate(t *testing.T) {
	// ADD R2, R1, #10
	v := vm.NewVM()
	v.CPU.R[1] = 5
	v.CPU.PC = 0x8000

	// ADD R2, R1, #10 (E281200A)
	opcode := uint32(0xE281200A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[2] != 15 {
		t.Errorf("expected R2=15, got R2=%d", v.CPU.R[2])
	}
}

func TestADD_WithFlags(t *testing.T) {
	// ADDS R0, R1, R2 (set flags)
	v := vm.NewVM()
	v.CPU.R[1] = 5
	v.CPU.R[2] = 10
	v.CPU.PC = 0x8000

	// ADDS R0, R1, R2 (E0910002 with S bit)
	opcode := uint32(0xE0B10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 15 {
		t.Errorf("expected R0=15, got R0=%d", v.CPU.R[0])
	}

	// Result is positive, so N should be false, Z should be false
	if v.CPU.CPSR.N {
		t.Error("expected N flag to be clear")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag to be clear")
	}
}

func TestSUB_Immediate(t *testing.T) {
	// SUB R0, R1, #5
	v := vm.NewVM()
	v.CPU.R[1] = 20
	v.CPU.PC = 0x8000

	// SUB R0, R1, #5 (E2410005)
	opcode := uint32(0xE2410005)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 15 {
		t.Errorf("expected R0=15, got R0=%d", v.CPU.R[0])
	}
}

func TestSUB_ZeroFlag(t *testing.T) {
	// SUBS R0, R1, R2 (should set Z flag when result is 0)
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 10
	v.CPU.PC = 0x8000

	// SUBS R0, R1, R2 (E0510002 with S bit)
	opcode := uint32(0xE0510002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=%d", v.CPU.R[0])
	}

	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}
}

func TestAND_Immediate(t *testing.T) {
	// AND R0, R1, #0xFF
	v := vm.NewVM()
	v.CPU.R[1] = 0x12345678
	v.CPU.PC = 0x8000

	// AND R0, R1, #0xFF (E20100FF)
	opcode := uint32(0xE20100FF)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x78 {
		t.Errorf("expected R0=0x78, got R0=0x%X", v.CPU.R[0])
	}
}

func TestORR_Immediate(t *testing.T) {
	// ORR R0, R1, #0x0F
	v := vm.NewVM()
	v.CPU.R[1] = 0xF0
	v.CPU.PC = 0x8000

	// ORR R0, R1, #0x0F (E381000F)
	opcode := uint32(0xE381000F)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFF {
		t.Errorf("expected R0=0xFF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestEOR_Immediate(t *testing.T) {
	// EOR R0, R1, #0xFF
	v := vm.NewVM()
	v.CPU.R[1] = 0xAA
	v.CPU.PC = 0x8000

	// EOR R0, R1, #0xFF (E22100FF)
	opcode := uint32(0xE22100FF)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x55 {
		t.Errorf("expected R0=0x55, got R0=0x%X", v.CPU.R[0])
	}
}

func TestMVN_Immediate(t *testing.T) {
	// MVN R0, #0 (should give 0xFFFFFFFF)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MVN R0, #0 (E3E00000)
	opcode := uint32(0xE3E00000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestCMP_Instruction(t *testing.T) {
	// CMP R0, R1 (should set flags but not write result)
	v := vm.NewVM()
	v.CPU.R[0] = 10
	v.CPU.R[1] = 5
	v.CPU.PC = 0x8000

	// CMP R0, R1 (E1500001)
	opcode := uint32(0xE1500001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// R0 should be unchanged
	if v.CPU.R[0] != 10 {
		t.Errorf("CMP should not modify R0, got R0=%d", v.CPU.R[0])
	}

	// 10 - 5 = 5, so Z should be clear, C should be set (no borrow)
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag to be clear")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set (no borrow)")
	}
}

func TestTST_Instruction(t *testing.T) {
	// TST R0, #0xFF (test if any of lower 8 bits are set)
	v := vm.NewVM()
	v.CPU.R[0] = 0x100
	v.CPU.PC = 0x8000

	// TST R0, #0xFF (E31000FF)
	opcode := uint32(0xE31000FF)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0x100 & 0xFF = 0, so Z should be set
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}
}

func TestADD_Overflow(t *testing.T) {
	// ADDS R0, R1, R2 (test signed overflow)
	v := vm.NewVM()
	v.CPU.R[1] = 0x7FFFFFFF // INT32_MAX
	v.CPU.R[2] = 1
	v.CPU.PC = 0x8000

	// ADDS R0, R1, R2 (E0B10002)
	opcode := uint32(0xE0B10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}

	// Should set overflow flag
	if !v.CPU.CPSR.V {
		t.Error("expected V flag to be set (signed overflow)")
	}

	// Should set negative flag
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set (result is negative)")
	}
}

func TestADD_Carry(t *testing.T) {
	// ADDS R0, R1, R2 (test unsigned overflow/carry)
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 1
	v.CPU.PC = 0x8000

	// ADDS R0, R1, R2 (E0B10002)
	opcode := uint32(0xE0B10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=%d", v.CPU.R[0])
	}

	// Should set carry flag
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set (unsigned overflow)")
	}

	// Should set zero flag
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}
}

func TestConditionalExecution_EQ(t *testing.T) {
	// MOVEQ R0, #1 (should execute when Z flag is set)
	v := vm.NewVM()
	v.CPU.CPSR.Z = true // Set Z flag
	v.CPU.PC = 0x8000

	// MOVEQ R0, #1 (03A00001) - condition code 0000 (EQ)
	opcode := uint32(0x03A00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 1 {
		t.Errorf("expected R0=1 when Z flag set, got R0=%d", v.CPU.R[0])
	}
}

func TestConditionalExecution_NE(t *testing.T) {
	// MOVNE R0, #1 (should NOT execute when Z flag is set)
	v := vm.NewVM()
	v.CPU.CPSR.Z = true // Set Z flag
	v.CPU.PC = 0x8000

	// MOVNE R0, #1 (13A00001) - condition code 0001 (NE)
	opcode := uint32(0x13A00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0 when condition fails, got R0=%d", v.CPU.R[0])
	}
}

func TestShift_LSL(t *testing.T) {
	// MOV R0, R1, LSL #2 (shift left by 2)
	v := vm.NewVM()
	v.CPU.R[1] = 5
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSL #2 (E1A00101)
	opcode := uint32(0xE1A00101)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 20 {
		t.Errorf("expected R0=20, got R0=%d", v.CPU.R[0])
	}
}

func TestShift_LSR(t *testing.T) {
	// MOV R0, R1, LSR #2 (shift right by 2)
	v := vm.NewVM()
	v.CPU.R[1] = 20
	v.CPU.PC = 0x8000

	// MOV R0, R1, LSR #2 (E1A00121)
	opcode := uint32(0xE1A00121)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 5 {
		t.Errorf("expected R0=5, got R0=%d", v.CPU.R[0])
	}
}
