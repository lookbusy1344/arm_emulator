package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ============================================================================
// Condition Code Tests - All 15 condition codes
// ============================================================================

// EQ - Equal (Z set)
func TestCondition_EQ_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVEQ R0, #42 (03A0002A)
	opcode := uint32(0x03A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("EQ condition true: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_EQ_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVEQ R0, #42 (03A0002A)
	opcode := uint32(0x03A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("EQ condition false: expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// NE - Not Equal (Z clear)
func TestCondition_NE_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVNE R0, #42 (13A0002A)
	opcode := uint32(0x13A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("NE condition true: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_NE_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVNE R0, #42 (13A0002A)
	opcode := uint32(0x13A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("NE condition false: expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// CS/HS - Carry Set / Unsigned Higher or Same (C set)
func TestCondition_CS_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.C = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVCS R0, #42 (23A0002A)
	opcode := uint32(0x23A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("CS condition true: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_CS_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.C = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVCS R0, #42 (23A0002A)
	opcode := uint32(0x23A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("CS condition false: expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// CC/LO - Carry Clear / Unsigned Lower (C clear)
func TestCondition_CC_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.C = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVCC R0, #42 (33A0002A)
	opcode := uint32(0x33A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("CC condition true: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_CC_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.C = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVCC R0, #42 (33A0002A)
	opcode := uint32(0x33A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("CC condition false: expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// MI - Minus/Negative (N set)
func TestCondition_MI_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVMI R0, #42 (43A0002A)
	opcode := uint32(0x43A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("MI condition true: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_MI_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVMI R0, #42 (43A0002A)
	opcode := uint32(0x43A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("MI condition false: expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// PL - Plus/Positive or Zero (N clear)
func TestCondition_PL_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVPL R0, #42 (53A0002A)
	opcode := uint32(0x53A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("PL condition true: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_PL_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVPL R0, #42 (53A0002A)
	opcode := uint32(0x53A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("PL condition false: expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// VS - Overflow Set (V set)
func TestCondition_VS_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.V = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVVS R0, #42 (63A0002A)
	opcode := uint32(0x63A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("VS condition true: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_VS_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVVS R0, #42 (63A0002A)
	opcode := uint32(0x63A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("VS condition false: expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// VC - Overflow Clear (V clear)
func TestCondition_VC_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVVC R0, #42 (73A0002A)
	opcode := uint32(0x73A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("VC condition true: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_VC_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.V = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVVC R0, #42 (73A0002A)
	opcode := uint32(0x73A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("VC condition false: expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// HI - Unsigned Higher (C set and Z clear)
func TestCondition_HI_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.C = true
	v.CPU.CPSR.Z = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVHI R0, #42 (83A0002A)
	opcode := uint32(0x83A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("HI condition true: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_HI_False_ClearZ(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.C = false
	v.CPU.CPSR.Z = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVHI R0, #42 (83A0002A)
	opcode := uint32(0x83A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("HI condition false (C=0): expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_HI_False_SetZ(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.C = true
	v.CPU.CPSR.Z = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVHI R0, #42 (83A0002A)
	opcode := uint32(0x83A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("HI condition false (Z=1): expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// LS - Unsigned Lower or Same (C clear or Z set)
func TestCondition_LS_True_ClearC(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.C = false
	v.CPU.CPSR.Z = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVLS R0, #42 (93A0002A)
	opcode := uint32(0x93A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("LS condition true (C=0): expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_LS_True_SetZ(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.C = true
	v.CPU.CPSR.Z = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVLS R0, #42 (93A0002A)
	opcode := uint32(0x93A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("LS condition true (Z=1): expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_LS_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.C = true
	v.CPU.CPSR.Z = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVLS R0, #42 (93A0002A)
	opcode := uint32(0x93A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("LS condition false: expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// GE - Signed Greater or Equal (N == V)
func TestCondition_GE_True_BothSet(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = true
	v.CPU.CPSR.V = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVGE R0, #42 (A3A0002A)
	opcode := uint32(0xA3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("GE condition true (N=V=1): expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_GE_True_BothClear(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = false
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVGE R0, #42 (A3A0002A)
	opcode := uint32(0xA3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("GE condition true (N=V=0): expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_GE_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = true
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVGE R0, #42 (A3A0002A)
	opcode := uint32(0xA3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("GE condition false (N≠V): expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// LT - Signed Less Than (N != V)
func TestCondition_LT_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = true
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVLT R0, #42 (B3A0002A)
	opcode := uint32(0xB3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("LT condition true (N≠V): expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_LT_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.N = true
	v.CPU.CPSR.V = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVLT R0, #42 (B3A0002A)
	opcode := uint32(0xB3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("LT condition false (N=V): expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// GT - Signed Greater Than (Z clear and N == V)
func TestCondition_GT_True(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.N = false
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVGT R0, #42 (C3A0002A)
	opcode := uint32(0xC3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("GT condition true: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_GT_False_ZSet(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = true
	v.CPU.CPSR.N = false
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVGT R0, #42 (C3A0002A)
	opcode := uint32(0xC3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("GT condition false (Z=1): expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_GT_False_NDiffV(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.N = true
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVGT R0, #42 (C3A0002A)
	opcode := uint32(0xC3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("GT condition false (N≠V): expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// LE - Signed Less or Equal (Z set or N != V)
func TestCondition_LE_True_ZSet(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = true
	v.CPU.CPSR.N = false
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVLE R0, #42 (D3A0002A)
	opcode := uint32(0xD3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("LE condition true (Z=1): expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_LE_True_NDiffV(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.N = true
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVLE R0, #42 (D3A0002A)
	opcode := uint32(0xD3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("LE condition true (N≠V): expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_LE_False(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.N = false
	v.CPU.CPSR.V = false
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVLE R0, #42 (D3A0002A)
	opcode := uint32(0xD3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0 {
		t.Errorf("LE condition false: expected R0=0, got R0=%d", v.CPU.R[0])
	}
}

// AL - Always (unconditional)
func TestCondition_AL(t *testing.T) {
	v := vm.NewVM()
	// Set all flags randomly
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = true
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = true
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	// MOVAL R0, #42 (same as MOV) (E3A0002A)
	opcode := uint32(0xE3A0002A)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 42 {
		t.Errorf("AL condition: expected R0=42, got R0=%d", v.CPU.R[0])
	}
}

// ============================================================================
// Complex condition tests with actual operations
// ============================================================================

func TestCondition_AfterCMP(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 10
	v.CPU.R[1] = 5
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	setupCodeWrite(v)

	// CMP R0, R1 (should set C and clear Z since 10 > 5)
	opcode := uint32(0xE1500001)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Now MOVGT should execute (Z=0 and N=V)
	v.CPU.PC = 0x8004
	opcode = uint32(0xC3A02064) // MOVGT R2, #100
	v.Memory.WriteWord(0x8004, opcode)
	v.Step()

	if v.CPU.R[2] != 100 {
		t.Errorf("expected R2=100 after CMP and MOVGT, got R2=%d", v.CPU.R[2])
	}
}

func TestCondition_AfterADDS_Overflow(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0x7FFFFFFF
	v.CPU.R[1] = 1
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	setupCodeWrite(v)

	// ADDS R0, R0, R1 (should set V flag for overflow)
	opcode := uint32(0xE0B00001)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// MOVVS should execute
	v.CPU.PC = 0x8004
	opcode = uint32(0x63A02001) // MOVVS R2, #1
	v.Memory.WriteWord(0x8004, opcode)
	v.Step()

	if v.CPU.R[2] != 1 {
		t.Errorf("expected R2=1 after overflow and MOVVS, got R2=%d", v.CPU.R[2])
	}
}

func TestCondition_ConditionalBranch(t *testing.T) {
	v := vm.NewVM()
	v.CPU.CPSR.Z = true // Set zero flag
	v.CPU.PC = 0x8000

	setupCodeWrite(v)

	// BEQ forward (should branch since Z=1)
	// Branch offset of +2 instructions (8 bytes)
	opcode := uint32(0x0A000001) // BEQ +8
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// PC should be at 0x8000 + 8 + 8 (instruction + 8, +2 words)
	expectedPC := uint32(0x8000 + 4 + 8)
	if v.CPU.PC != expectedPC {
		t.Errorf("expected PC=0x%X after BEQ, got PC=0x%X", expectedPC, v.CPU.PC)
	}
}

func TestCondition_MultipleConditions(t *testing.T) {
	// Test that multiple conditional instructions can be chained
	v := vm.NewVM()
	v.CPU.R[0] = 10
	v.CPU.R[1] = 10
	v.CPU.R[2] = 0
	v.CPU.R[3] = 0
	v.CPU.PC = 0x8000

	setupCodeWrite(v)

	// CMP R0, R1 (should set Z flag since equal)
	opcode := uint32(0xE1500001)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// MOVEQ R2, #1 (should execute)
	v.CPU.PC = 0x8004
	opcode = uint32(0x03A02001)
	v.Memory.WriteWord(0x8004, opcode)
	v.Step()

	// MOVNE R3, #1 (should NOT execute)
	v.CPU.PC = 0x8008
	opcode = uint32(0x13A03001)
	v.Memory.WriteWord(0x8008, opcode)
	v.Step()

	if v.CPU.R[2] != 1 {
		t.Errorf("expected R2=1 after MOVEQ, got R2=%d", v.CPU.R[2])
	}
	if v.CPU.R[3] != 0 {
		t.Errorf("expected R3=0 after MOVNE (should not execute), got R3=%d", v.CPU.R[3])
	}
}

// ============================================================================
// Conditional execution with different instruction types
// ============================================================================

func TestCondition_ADD_EQ(t *testing.T) {
	// Test ADD with EQ condition
	v := vm.NewVM()
	v.CPU.CPSR.Z = true // Equal condition
	v.CPU.R[0] = 10
	v.CPU.R[1] = 5
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	// ADDEQ R2, R0, R1 (02810001)
	opcode := uint32(0x02810001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[2] != 15 {
		t.Errorf("ADDEQ: expected R2=15, got R2=%d", v.CPU.R[2])
	}
}

func TestCondition_SUB_NE(t *testing.T) {
	// Test SUB with NE condition
	v := vm.NewVM()
	v.CPU.CPSR.Z = false // Not equal condition
	v.CPU.R[0] = 20
	v.CPU.R[1] = 7
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	// SUBNE R2, R0, R1 (12410001)
	opcode := uint32(0x12410001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[2] != 13 {
		t.Errorf("SUBNE: expected R2=13, got R2=%d", v.CPU.R[2])
	}
}

func TestCondition_AND_CS(t *testing.T) {
	// Test AND with CS (carry set) condition
	v := vm.NewVM()
	v.CPU.CPSR.C = true // Carry set
	v.CPU.R[0] = 0xFF00
	v.CPU.R[1] = 0x0FFF
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	// ANDCS R2, R0, R1 (22010001)
	opcode := uint32(0x22010001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[2] != 0x0F00 {
		t.Errorf("ANDCS: expected R2=0x0F00, got R2=0x%X", v.CPU.R[2])
	}
}

func TestCondition_ORR_MI(t *testing.T) {
	// Test ORR with MI (negative/minus) condition
	v := vm.NewVM()
	v.CPU.CPSR.N = true // Negative flag set
	v.CPU.R[0] = 0xF000
	v.CPU.R[1] = 0x000F
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	// ORRMI R2, R0, R1 (41810001)
	opcode := uint32(0x41810001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[2] != 0xF00F {
		t.Errorf("ORRMI: expected R2=0xF00F, got R2=0x%X", v.CPU.R[2])
	}
}

func TestCondition_LDR_GT(t *testing.T) {
	// Test LDR with GT (greater than) condition
	v := vm.NewVM()
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.N = false
	v.CPU.CPSR.V = false // Z=0 and N=V (GT condition)
	v.CPU.R[0] = 0
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x20000, 0xDEADBEEF)

	// LDRGT R0, [R1] (C5910000)
	opcode := uint32(0xC5910000)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xDEADBEEF {
		t.Errorf("LDRGT: expected R0=0xDEADBEEF, got R0=0x%X", v.CPU.R[0])
	}
}

func TestCondition_STR_LE(t *testing.T) {
	// Test STR with LE (less or equal) condition
	v := vm.NewVM()
	v.CPU.CPSR.Z = true // LE satisfied with Z=1
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0x20000
	v.CPU.PC = 0x8000

	setupCodeWrite(v)

	// STRLE R0, [R1] (D5810000)
	opcode := uint32(0xD5810000)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	value, _ := v.Memory.ReadWord(0x20000)
	if value != 0x12345678 {
		t.Errorf("STRLE: expected memory=0x12345678, got 0x%X", value)
	}
}

func TestCondition_CMP_AL(t *testing.T) {
	// Test CMP with AL (always) condition - should always execute
	v := vm.NewVM()
	// Set random flags
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = true
	v.CPU.R[0] = 10
	v.CPU.R[1] = 10
	v.CPU.PC = 0x8000

	// CMPAL R0, R1 (same as CMP) (E1500001)
	opcode := uint32(0xE1500001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Should set Z flag since values are equal
	if !v.CPU.CPSR.Z {
		t.Error("CMPAL: expected Z flag to be set")
	}
}

func TestCondition_EOR_VC(t *testing.T) {
	// Test EOR with VC (overflow clear) condition
	v := vm.NewVM()
	v.CPU.CPSR.V = false // Overflow clear
	v.CPU.R[0] = 0xAAAA
	v.CPU.R[1] = 0x5555
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	// EORVC R2, R0, R1 (72210001)
	opcode := uint32(0x72210001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[2] != 0xFFFF {
		t.Errorf("EORVC: expected R2=0xFFFF, got R2=0x%X", v.CPU.R[2])
	}
}

func TestCondition_BIC_HI(t *testing.T) {
	// Test BIC with HI (unsigned higher) condition
	v := vm.NewVM()
	v.CPU.CPSR.C = true
	v.CPU.CPSR.Z = false // C=1 and Z=0 (HI condition)
	v.CPU.R[0] = 0xFFFF
	v.CPU.R[1] = 0x00FF
	v.CPU.R[2] = 0
	v.CPU.PC = 0x8000

	// BICHI R2, R0, R1 (81C10001)
	opcode := uint32(0x81C10001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[2] != 0xFF00 {
		t.Errorf("BICHI: expected R2=0xFF00, got R2=0x%X", v.CPU.R[2])
	}
}

func TestCondition_MUL_LS(t *testing.T) {
	// Test MUL with LS (unsigned lower or same) condition
	v := vm.NewVM()
	v.CPU.CPSR.C = false // C=0 satisfies LS
	v.CPU.R[0] = 0
	v.CPU.R[1] = 5
	v.CPU.R[2] = 6
	v.CPU.PC = 0x8000

	// MULLS R0, R1, R2 (90000291) - need different encoding for conditional
	// MULS R0, R1, R2 with LS condition
	opcode := uint32(0x90000291)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 30 {
		t.Errorf("MULLS: expected R0=30, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_RSB_GE(t *testing.T) {
	// Test RSB (Reverse Subtract) with GE condition
	v := vm.NewVM()
	v.CPU.CPSR.N = false
	v.CPU.CPSR.V = false // N=V (GE condition)
	v.CPU.R[0] = 0
	v.CPU.R[1] = 10
	v.CPU.R[2] = 30
	v.CPU.PC = 0x8000

	// RSBGE R0, R1, R2 (A0610002)
	opcode := uint32(0xA0610002)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// R0 = R2 - R1 = 30 - 10 = 20
	if v.CPU.R[0] != 20 {
		t.Errorf("RSBGE: expected R0=20, got R0=%d", v.CPU.R[0])
	}
}

func TestCondition_MVN_LT(t *testing.T) {
	// Test MVN (Move Not) with LT condition
	v := vm.NewVM()
	v.CPU.CPSR.N = true
	v.CPU.CPSR.V = false // N!=V (LT condition)
	v.CPU.R[0] = 0
	v.CPU.R[1] = 0x0000FFFF
	v.CPU.PC = 0x8000

	// MVNLT R0, R1 (B1E00001)
	opcode := uint32(0xB1E00001)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	if v.CPU.R[0] != 0xFFFF0000 {
		t.Errorf("MVNLT: expected R0=0xFFFF0000, got R0=0x%X", v.CPU.R[0])
	}
}
