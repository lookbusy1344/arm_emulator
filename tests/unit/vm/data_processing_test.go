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

// ============================================================================
// Additional MOV instruction tests
// ============================================================================

func TestMOV_NegativeFlag(t *testing.T) {
	// MOVS R0, #0x80000000 (should set N flag)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MOVS with immediate 0x80000000
	// Need to use MVN to get 0x80000000 or use register
	v.CPU.R[1] = 0x80000000
	opcode := uint32(0xE1B00001) // MOVS R0, R1
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set")
	}
}

func TestMOV_ZeroResult(t *testing.T) {
	// MOVS R0, #0 (should set Z flag)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MOVS R0, #0 (E3B00000)
	opcode := uint32(0xE3B00000)
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

// ============================================================================
// MVN (Move Not) instruction tests
// ============================================================================

func TestMVN_Register(t *testing.T) {
	// MVN R0, R1 (bitwise NOT of R1)
	v := vm.NewVM()
	v.CPU.R[1] = 0x0F0F0F0F
	v.CPU.PC = 0x8000

	// MVN R0, R1 (E1E00001)
	opcode := uint32(0xE1E00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xF0F0F0F0 {
		t.Errorf("expected R0=0xF0F0F0F0, got R0=0x%X", v.CPU.R[0])
	}
}

func TestMVN_WithFlags(t *testing.T) {
	// MVNS R0, #0x7FFFFFFF (should result in 0x80000000 and set N flag)
	v := vm.NewVM()
	v.CPU.R[1] = 0x7FFFFFFF
	v.CPU.PC = 0x8000

	// MVNS R0, R1 (E1F00001)
	opcode := uint32(0xE1F00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set")
	}
}

// ============================================================================
// ADD instruction comprehensive tests
// ============================================================================

func TestADD_Register(t *testing.T) {
	// ADD R0, R1, R2
	v := vm.NewVM()
	v.CPU.R[1] = 100
	v.CPU.R[2] = 50
	v.CPU.PC = 0x8000

	// ADD R0, R1, R2 (E0810002)
	opcode := uint32(0xE0810002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 150 {
		t.Errorf("expected R0=150, got R0=%d", v.CPU.R[0])
	}
}

func TestADD_ZeroResult(t *testing.T) {
	// ADDS R0, R1, R2 (0 + 0 should set Z flag)
	v := vm.NewVM()
	v.CPU.R[1] = 0
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	// ADDS R0, R1, R2 (E0B10002)
	opcode := uint32(0xE0B10002)
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

func TestADD_NegativeResult(t *testing.T) {
	// ADDS R0, R1, R2 (result should set N flag)
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000
	v.CPU.R[2] = 0x10000000
	v.CPU.PC = 0x8000

	// ADDS R0, R1, R2 (E0B10002)
	opcode := uint32(0xE0B10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x90000000 {
		t.Errorf("expected R0=0x90000000, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set")
	}
}

// ============================================================================
// ADC (Add with Carry) instruction tests
// ============================================================================

func TestADC_WithCarryClear(t *testing.T) {
	// ADC R0, R1, R2 (with C flag clear)
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 5
	v.CPU.CPSR.C = false
	v.CPU.PC = 0x8000

	// ADC R0, R1, R2 (E0A10002)
	opcode := uint32(0xE0A10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 10 + 5 + 0 = 15
	if v.CPU.R[0] != 15 {
		t.Errorf("expected R0=15, got R0=%d", v.CPU.R[0])
	}
}

func TestADC_WithCarrySet(t *testing.T) {
	// ADC R0, R1, R2 (with C flag set)
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 5
	v.CPU.CPSR.C = true
	v.CPU.PC = 0x8000

	// ADC R0, R1, R2 (E0A10002)
	opcode := uint32(0xE0A10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 10 + 5 + 1 = 16
	if v.CPU.R[0] != 16 {
		t.Errorf("expected R0=16, got R0=%d", v.CPU.R[0])
	}
}

func TestADC_CarryChain(t *testing.T) {
	// ADCS R0, R1, R2 (should propagate carry)
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 0
	v.CPU.CPSR.C = true // Add with carry set
	v.CPU.PC = 0x8000

	// ADCS R0, R1, R2 (E0B10002)
	opcode := uint32(0xE0B10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0xFFFFFFFF + 0 + 1 = 0, with carry out
	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set")
	}
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}
}

// ============================================================================
// SUB instruction comprehensive tests
// ============================================================================

func TestSUB_Register(t *testing.T) {
	// SUB R0, R1, R2
	v := vm.NewVM()
	v.CPU.R[1] = 100
	v.CPU.R[2] = 30
	v.CPU.PC = 0x8000

	// SUB R0, R1, R2 (E0410002)
	opcode := uint32(0xE0410002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 70 {
		t.Errorf("expected R0=70, got R0=%d", v.CPU.R[0])
	}
}

func TestSUB_NegativeResult(t *testing.T) {
	// SUBS R0, R1, R2 (result should be negative)
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 20
	v.CPU.PC = 0x8000

	// SUBS R0, R1, R2 (E0510002)
	opcode := uint32(0xE0510002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 10 - 20 = -10 (0xFFFFFFF6 in two's complement)
	if v.CPU.R[0] != 0xFFFFFFF6 {
		t.Errorf("expected R0=0xFFFFFFF6, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set")
	}
	if v.CPU.CPSR.C {
		t.Error("expected C flag to be clear (borrow occurred)")
	}
}

func TestSUB_Borrow(t *testing.T) {
	// SUBS R0, R1, R2 (test borrow flag)
	v := vm.NewVM()
	v.CPU.R[1] = 5
	v.CPU.R[2] = 10
	v.CPU.PC = 0x8000

	// SUBS R0, R1, R2 (E0510002)
	opcode := uint32(0xE0510002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// In ARM, C=0 means borrow occurred
	if v.CPU.CPSR.C {
		t.Error("expected C flag to be clear (borrow occurred)")
	}
}

// ============================================================================
// SBC (Subtract with Carry) instruction tests
// ============================================================================

func TestSBC_WithCarrySet(t *testing.T) {
	// SBC R0, R1, R2 (with C flag set, no borrow)
	v := vm.NewVM()
	v.CPU.R[1] = 20
	v.CPU.R[2] = 5
	v.CPU.CPSR.C = true
	v.CPU.PC = 0x8000

	// SBC R0, R1, R2 (E0C10002)
	opcode := uint32(0xE0C10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 20 - 5 - 0 = 15 (C=1 means subtract 0)
	if v.CPU.R[0] != 15 {
		t.Errorf("expected R0=15, got R0=%d", v.CPU.R[0])
	}
}

func TestSBC_WithCarryClear(t *testing.T) {
	// SBC R0, R1, R2 (with C flag clear, borrow 1)
	v := vm.NewVM()
	v.CPU.R[1] = 20
	v.CPU.R[2] = 5
	v.CPU.CPSR.C = false
	v.CPU.PC = 0x8000

	// SBC R0, R1, R2 (E0C10002)
	opcode := uint32(0xE0C10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 20 - 5 - 1 = 14 (C=0 means subtract 1)
	if v.CPU.R[0] != 14 {
		t.Errorf("expected R0=14, got R0=%d", v.CPU.R[0])
	}
}

// ============================================================================
// RSB (Reverse Subtract) instruction tests
// ============================================================================

func TestRSB_Immediate(t *testing.T) {
	// RSB R0, R1, #100 (R0 = 100 - R1)
	v := vm.NewVM()
	v.CPU.R[1] = 30
	v.CPU.PC = 0x8000

	// RSB R0, R1, #100 (E2610064)
	opcode := uint32(0xE2610064)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 70 {
		t.Errorf("expected R0=70, got R0=%d", v.CPU.R[0])
	}
}

func TestRSB_Register(t *testing.T) {
	// RSB R0, R1, R2 (R0 = R2 - R1)
	v := vm.NewVM()
	v.CPU.R[1] = 25
	v.CPU.R[2] = 100
	v.CPU.PC = 0x8000

	// RSB R0, R1, R2 (E0610002)
	opcode := uint32(0xE0610002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 75 {
		t.Errorf("expected R0=75, got R0=%d", v.CPU.R[0])
	}
}

func TestRSB_NegateRegister(t *testing.T) {
	// RSB R0, R1, #0 (R0 = 0 - R1, effectively negates R1)
	v := vm.NewVM()
	v.CPU.R[1] = 42
	v.CPU.PC = 0x8000

	// RSB R0, R1, #0 (E2610000)
	opcode := uint32(0xE2610000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0 - 42 = -42 (0xFFFFFFD6)
	expected := uint32(0xFFFFFFD6)
	if v.CPU.R[0] != expected {
		t.Errorf("expected R0=0x%X, got R0=0x%X", expected, v.CPU.R[0])
	}
}

// ============================================================================
// RSC (Reverse Subtract with Carry) instruction tests
// ============================================================================

func TestRSC_WithCarrySet(t *testing.T) {
	// RSC R0, R1, R2 (R0 = R2 - R1 - !C, with C set)
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 50
	v.CPU.CPSR.C = true
	v.CPU.PC = 0x8000

	// RSC R0, R1, R2 (E0E10002)
	opcode := uint32(0xE0E10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 50 - 10 - 0 = 40
	if v.CPU.R[0] != 40 {
		t.Errorf("expected R0=40, got R0=%d", v.CPU.R[0])
	}
}

func TestRSC_WithCarryClear(t *testing.T) {
	// RSC R0, R1, R2 (R0 = R2 - R1 - 1, with C clear)
	v := vm.NewVM()
	v.CPU.R[1] = 10
	v.CPU.R[2] = 50
	v.CPU.CPSR.C = false
	v.CPU.PC = 0x8000

	// RSC R0, R1, R2 (E0E10002)
	opcode := uint32(0xE0E10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 50 - 10 - 1 = 39
	if v.CPU.R[0] != 39 {
		t.Errorf("expected R0=39, got R0=%d", v.CPU.R[0])
	}
}

// ============================================================================
// AND instruction comprehensive tests
// ============================================================================

func TestAND_Register(t *testing.T) {
	// AND R0, R1, R2
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFF
	v.CPU.R[2] = 0xFF00
	v.CPU.PC = 0x8000

	// AND R0, R1, R2 (E0010002)
	opcode := uint32(0xE0010002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFF00 {
		t.Errorf("expected R0=0xFF00, got R0=0x%X", v.CPU.R[0])
	}
}

func TestAND_ClearBits(t *testing.T) {
	// ANDS R0, R1, #0 (should result in 0 and set Z flag)
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.PC = 0x8000

	// ANDS R0, R1, #0 (E2110000)
	opcode := uint32(0xE2110000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}
}

func TestAND_MaskOperation(t *testing.T) {
	// AND R0, R1, #0x00FF0000 (extract specific byte)
	v := vm.NewVM()
	v.CPU.R[1] = 0x12345678
	v.CPU.PC = 0x8000

	// For immediate, need to construct valid rotated immediate
	// AND R0, R1, R2 with R2 = 0x00FF0000
	v.CPU.R[2] = 0x00FF0000
	opcode := uint32(0xE0010002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x00340000 {
		t.Errorf("expected R0=0x00340000, got R0=0x%X", v.CPU.R[0])
	}
}

// ============================================================================
// ORR instruction comprehensive tests
// ============================================================================

func TestORR_Register(t *testing.T) {
	// ORR R0, R1, R2
	v := vm.NewVM()
	v.CPU.R[1] = 0xFF00
	v.CPU.R[2] = 0x00FF
	v.CPU.PC = 0x8000

	// ORR R0, R1, R2 (E1810002)
	opcode := uint32(0xE1810002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFF {
		t.Errorf("expected R0=0xFFFF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestORR_SetBits(t *testing.T) {
	// ORR R0, R1, #0xFF (set lower byte)
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFF00
	v.CPU.PC = 0x8000

	// ORR R0, R1, #0xFF (E38100FF)
	opcode := uint32(0xE38100FF)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestORR_NoChange(t *testing.T) {
	// ORRS R0, R1, #0 (should not change R1, but set flags)
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000000
	v.CPU.PC = 0x8000

	// ORRS R0, R1, #0 (E3910000)
	opcode := uint32(0xE3910000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x80000000 {
		t.Errorf("expected R0=0x80000000, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set")
	}
}

// ============================================================================
// EOR (Exclusive OR) instruction comprehensive tests
// ============================================================================

func TestEOR_Register(t *testing.T) {
	// EOR R0, R1, R2
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFF
	v.CPU.R[2] = 0xFF00
	v.CPU.PC = 0x8000

	// EOR R0, R1, R2 (E0210002)
	opcode := uint32(0xE0210002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 0xFFFF XOR 0xFF00 = 0x00FF
	if v.CPU.R[0] != 0x00FF {
		t.Errorf("expected R0=0x00FF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestEOR_ToggleBits(t *testing.T) {
	// EOR R0, R1, #0xFFFFFFFF (invert all bits)
	v := vm.NewVM()
	v.CPU.R[1] = 0xAAAAAAAA
	v.CPU.R[2] = 0xFFFFFFFF
	v.CPU.PC = 0x8000

	// EOR R0, R1, R2 (E0210002)
	opcode := uint32(0xE0210002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x55555555 {
		t.Errorf("expected R0=0x55555555, got R0=0x%X", v.CPU.R[0])
	}
}

func TestEOR_SelfXOR(t *testing.T) {
	// EORS R0, R1, R1 (should result in 0 and set Z flag)
	v := vm.NewVM()
	v.CPU.R[1] = 0x12345678
	v.CPU.PC = 0x8000

	// EORS R0, R1, R1 (E0310001)
	opcode := uint32(0xE0310001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0, got R0=0x%X", v.CPU.R[0])
	}
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}
}

// ============================================================================
// BIC (Bit Clear) instruction tests
// ============================================================================

func TestBIC_Immediate(t *testing.T) {
	// BIC R0, R1, #0xFF (clear lower byte)
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.PC = 0x8000

	// BIC R0, R1, #0xFF (E3C100FF)
	opcode := uint32(0xE3C100FF)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFFFF00 {
		t.Errorf("expected R0=0xFFFFFF00, got R0=0x%X", v.CPU.R[0])
	}
}

func TestBIC_Register(t *testing.T) {
	// BIC R0, R1, R2 (clear bits in R1 that are set in R2)
	v := vm.NewVM()
	v.CPU.R[1] = 0xFFFFFFFF
	v.CPU.R[2] = 0x0F0F0F0F
	v.CPU.PC = 0x8000

	// BIC R0, R1, R2 (E1C10002)
	opcode := uint32(0xE1C10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xF0F0F0F0 {
		t.Errorf("expected R0=0xF0F0F0F0, got R0=0x%X", v.CPU.R[0])
	}
}

func TestBIC_ClearSpecificBits(t *testing.T) {
	// BICS R0, R1, #0x80000000 (clear sign bit)
	v := vm.NewVM()
	v.CPU.R[1] = 0x80000001
	v.CPU.R[2] = 0x80000000
	v.CPU.PC = 0x8000

	// BICS R0, R1, R2 (E1D10002)
	opcode := uint32(0xE1D10002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0x00000001 {
		t.Errorf("expected R0=0x00000001, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.CPSR.N {
		t.Error("expected N flag to be clear")
	}
}

// ============================================================================
// CMP (Compare) instruction comprehensive tests
// ============================================================================

func TestCMP_Equal(t *testing.T) {
	// CMP R0, R1 (when equal, should set Z flag)
	v := vm.NewVM()
	v.CPU.R[0] = 42
	v.CPU.R[1] = 42
	v.CPU.PC = 0x8000

	// CMP R0, R1 (E1500001)
	opcode := uint32(0xE1500001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set when values are equal")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set (no borrow)")
	}
}

func TestCMP_Greater(t *testing.T) {
	// CMP R0, R1 (R0 > R1)
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.R[1] = 50
	v.CPU.PC = 0x8000

	// CMP R0, R1 (E1500001)
	opcode := uint32(0xE1500001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.Z {
		t.Error("expected Z flag to be clear")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set (no borrow)")
	}
	if v.CPU.CPSR.N {
		t.Error("expected N flag to be clear (positive result)")
	}
}

func TestCMP_Less(t *testing.T) {
	// CMP R0, R1 (R0 < R1, should set N and clear C)
	v := vm.NewVM()
	v.CPU.R[0] = 50
	v.CPU.R[1] = 100
	v.CPU.PC = 0x8000

	// CMP R0, R1 (E1500001)
	opcode := uint32(0xE1500001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set (negative result)")
	}
	if v.CPU.CPSR.C {
		t.Error("expected C flag to be clear (borrow occurred)")
	}
}

func TestCMP_Immediate(t *testing.T) {
	// CMP R0, #100
	v := vm.NewVM()
	v.CPU.R[0] = 100
	v.CPU.PC = 0x8000

	// CMP R0, #100 (E3500064)
	opcode := uint32(0xE3500064)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set when equal")
	}
}

// ============================================================================
// CMN (Compare Negative) instruction tests
// ============================================================================

func TestCMN_Instruction(t *testing.T) {
	// CMN R0, R1 (compare R0 with -R1, i.e., R0 + R1)
	v := vm.NewVM()
	v.CPU.R[0] = 10
	v.CPU.R[1] = 0xFFFFFFF6 // -10 in two's complement
	v.CPU.PC = 0x8000

	// CMN R0, R1 (E1700001)
	opcode := uint32(0xE1700001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// 10 + (-10) = 0, should set Z flag
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}
}

func TestCMN_Overflow(t *testing.T) {
	// CMN R0, R1 (test overflow detection)
	v := vm.NewVM()
	v.CPU.R[0] = 0x7FFFFFFF
	v.CPU.R[1] = 1
	v.CPU.PC = 0x8000

	// CMN R0, R1 (E1700001)
	opcode := uint32(0xE1700001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.V {
		t.Error("expected V flag to be set (overflow)")
	}
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set (negative result)")
	}
}

// ============================================================================
// TEQ (Test Equivalence) instruction tests
// ============================================================================

func TestTEQ_Equal(t *testing.T) {
	// TEQ R0, R1 (when equal, should set Z flag)
	v := vm.NewVM()
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0x12345678
	v.CPU.PC = 0x8000

	// TEQ R0, R1 (E1300001)
	opcode := uint32(0xE1300001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// XOR of equal values is 0
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set when values are equal")
	}
}

func TestTEQ_Different(t *testing.T) {
	// TEQ R0, R1 (when different)
	v := vm.NewVM()
	v.CPU.R[0] = 0xFF
	v.CPU.R[1] = 0xAA
	v.CPU.PC = 0x8000

	// TEQ R0, R1 (E1300001)
	opcode := uint32(0xE1300001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.CPSR.Z {
		t.Error("expected Z flag to be clear when values are different")
	}
}

func TestTEQ_SignBit(t *testing.T) {
	// TEQ R0, #0x80000000 (test sign bit specifically)
	v := vm.NewVM()
	v.CPU.R[0] = 0x80000000
	v.CPU.R[1] = 0x80000000
	v.CPU.PC = 0x8000

	// TEQ R0, R1 (E1300001)
	opcode := uint32(0xE1300001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set")
	}
}
