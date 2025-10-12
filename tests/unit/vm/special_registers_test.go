package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ============================================================================
// Priority 4, Section 7: PC-Relative and Special Register Operations
// ============================================================================
//
// These tests verify that the special ARM registers (PC, SP, LR) behave
// correctly in various contexts:
// - PC (R15): Program Counter - points to current instruction + 8
// - SP (R13): Stack Pointer - used for function call stack
// - LR (R14): Link Register - stores return addresses
//

// ====== PC (R15) as Source Operand ======

func TestADD_PC_AsSource(t *testing.T) {
	// ADD R0, PC, #8
	// PC should read as current instruction + 8
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// ADD R0, PC, #8
	// Opcode: ADD (0100), I=1 (immediate), S=0
	// Rn=R15 (PC), Rd=R0, immediate=8
	opcode := uint32(0xE28F0008) // 1110 0010 1000 1111 0000 0000 0000 1000
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: PC reads as 0x8000 + 8 = 0x8008, then + 8 = 0x8010
	if v.CPU.R[0] != 0x8010 {
		t.Errorf("expected R0=0x8010, got R0=0x%X", v.CPU.R[0])
	}
}

func TestMOV_PC_AsSource(t *testing.T) {
	// MOV R0, PC
	// Should read PC + 8
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// MOV R0, PC (register move)
	// Opcode: MOV (1101), I=0, S=0, Rn=0 (ignored), Rd=R0, Rm=R15
	opcode := uint32(0xE1A0000F) // 1110 0001 1010 0000 0000 0000 0000 1111
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: PC + 8 = 0x8008
	if v.CPU.R[0] != 0x8008 {
		t.Errorf("expected R0=0x8008, got R0=0x%X", v.CPU.R[0])
	}
}

func TestLDR_PC_Relative(t *testing.T) {
	// LDR R0, [PC, #4]
	// Load from PC-relative address (common pattern for literal pools)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// Write test data at PC + 8 + 4 = 0x800C
	setupCodeWrite(v)
	v.Memory.WriteWord(0x800C, 0x12345678)

	// LDR R0, [PC, #4]
	// Opcode: LDR, I=0, P=1, U=1, W=0, Rn=PC, Rd=R0, offset=4
	opcode := uint32(0xE59F0004) // 1110 0101 1001 1111 0000 0000 0000 0100
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: load from (PC + 8) + 4 = 0x8008 + 4 = 0x800C
	if v.CPU.R[0] != 0x12345678 {
		t.Errorf("expected R0=0x12345678, got R0=0x%X", v.CPU.R[0])
	}
}

func TestSTR_PC_AsSource(t *testing.T) {
	// STR PC, [R1]
	// Store PC value to memory
	v := vm.NewVM()
	v.CPU.PC = 0x8000
	v.CPU.R[1] = 0x9000

	// STR PC, [R1]
	// Opcode: STR, I=0, P=1, U=1, W=0, Rn=R1, Rd=PC, offset=0
	opcode := uint32(0xE581F000) // 1110 0101 1000 0001 1111 0000 0000 0000
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: PC + 8 stored (ARM2 stores PC+8 for STR)
	// Note: Different ARM implementations vary
	val, err := v.Memory.ReadWord(0x9000)
	if err != nil {
		t.Fatalf("failed to read stored value: %v", err)
	}

	// PC + 8 = 0x8008
	if val != 0x8008 {
		t.Errorf("expected stored value 0x8008, got 0x%X", val)
	}
}

// ====== PC (R15) as Destination (Branch Operations) ======

func TestMOV_PC_AsBranch(t *testing.T) {
	// MOV PC, LR (common return from subroutine)
	v := vm.NewVM()
	v.CPU.PC = 0x8000
	v.CPU.R[14] = 0x9000 // LR contains return address

	// MOV PC, LR
	// Opcode: MOV (1101), I=0, S=0, Rn=0, Rd=PC, Rm=LR
	opcode := uint32(0xE1A0F00E) // 1110 0001 1010 0000 1111 0000 0000 1110
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: PC set to 0x9000
	if v.CPU.PC != 0x9000 {
		t.Errorf("expected PC=0x9000, got PC=0x%X", v.CPU.PC)
	}
}

func TestADD_PC_AsBranch(t *testing.T) {
	// ADD PC, PC, R0 (computed branch)
	v := vm.NewVM()
	v.CPU.PC = 0x8000
	v.CPU.R[0] = 0x100 // Branch offset

	// ADD PC, PC, R0
	// Opcode: ADD (0100), I=0, S=0, Rn=PC, Rd=PC, Rm=R0
	opcode := uint32(0xE08FF000) // 1110 0000 1000 1111 1111 0000 0000 0000
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: PC = (PC + 8) + R0 = 0x8008 + 0x100 = 0x8108
	if v.CPU.PC != 0x8108 {
		t.Errorf("expected PC=0x8108, got PC=0x%X", v.CPU.PC)
	}
}

func TestLDM_WithPC(t *testing.T) {
	// LDMIA SP!, {R0-R3, PC} (return from function with register restore)
	v := vm.NewVM()
	v.CPU.R[13] = 0x9000 // SP
	v.CPU.PC = 0x8000

	// Setup stack with values to load
	setupCodeWrite(v)
	v.Memory.WriteWord(0x9000, 0xAAAAAAAA) // R0
	v.Memory.WriteWord(0x9004, 0xBBBBBBBB) // R1
	v.Memory.WriteWord(0x9008, 0xCCCCCCCC) // R2
	v.Memory.WriteWord(0x900C, 0xDDDDDDDD) // R3
	v.Memory.WriteWord(0x9010, 0x8100)     // PC (return address)

	// LDMIA SP!, {R0-R3, PC}
	// Opcode: LDM, P=0, U=1, S=0, W=1 (writeback), Rn=SP, register list includes R0-R3, PC
	// Register list: bits 0-3 for R0-R3, bit 15 for PC = 0x800F
	opcode := uint32(0xE8BD800F) // 1110 1000 1011 1101 1000 0000 0000 1111
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: R0-R3 loaded, PC set to 0x8100, SP incremented by 20 (5 words)
	if v.CPU.R[0] != 0xAAAAAAAA {
		t.Errorf("expected R0=0xAAAAAAAA, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0xBBBBBBBB {
		t.Errorf("expected R1=0xBBBBBBBB, got R1=0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0xCCCCCCCC {
		t.Errorf("expected R2=0xCCCCCCCC, got R2=0x%X", v.CPU.R[2])
	}
	if v.CPU.R[3] != 0xDDDDDDDD {
		t.Errorf("expected R3=0xDDDDDDDD, got R3=0x%X", v.CPU.R[3])
	}
	if v.CPU.PC != 0x8100 {
		t.Errorf("expected PC=0x8100, got PC=0x%X", v.CPU.PC)
	}
	if v.CPU.R[13] != 0x9014 {
		t.Errorf("expected SP=0x9014, got SP=0x%X", v.CPU.R[13])
	}
}

// ====== SP (R13) Operations ======

func TestADD_SP_Adjustment(t *testing.T) {
	// ADD SP, SP, #16 (allocate stack space)
	v := vm.NewVM()
	v.CPU.R[13] = 0x10000 // SP
	v.CPU.PC = 0x8000

	// ADD SP, SP, #16
	// Opcode: ADD (0100), I=1, S=0, Rn=SP, Rd=SP, immediate=16
	opcode := uint32(0xE28DD010) // 1110 0010 1000 1101 1101 0000 0001 0000
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: SP = 0x10000 + 16 = 0x10010
	if v.CPU.R[13] != 0x10010 {
		t.Errorf("expected SP=0x10010, got SP=0x%X", v.CPU.R[13])
	}
}

func TestSUB_SP_Adjustment(t *testing.T) {
	// SUB SP, SP, #32 (deallocate stack space)
	v := vm.NewVM()
	v.CPU.R[13] = 0x10000 // SP
	v.CPU.PC = 0x8000

	// SUB SP, SP, #32
	// Opcode: SUB (0010), I=1, S=0, Rn=SP, Rd=SP, immediate=32 (0x20)
	opcode := uint32(0xE24DD020) // 1110 0010 0100 1101 1101 0000 0010 0000
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: SP = 0x10000 - 32 = 0xFFE0
	if v.CPU.R[13] != 0xFFE0 {
		t.Errorf("expected SP=0xFFE0, got SP=0x%X", v.CPU.R[13])
	}
}

func TestMOV_SP_Copy(t *testing.T) {
	// MOV R0, SP (save stack pointer)
	v := vm.NewVM()
	v.CPU.R[13] = 0x10000
	v.CPU.PC = 0x8000

	// MOV R0, SP
	// Opcode: MOV (1101), I=0, S=0, Rn=0, Rd=R0, Rm=SP
	opcode := uint32(0xE1A0000D) // 1110 0001 1010 0000 0000 0000 0000 1101
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: R0 = SP = 0x10000
	if v.CPU.R[0] != 0x10000 {
		t.Errorf("expected R0=0x10000, got R0=0x%X", v.CPU.R[0])
	}
}

func TestMOV_SP_Set(t *testing.T) {
	// MOV SP, R0 (restore stack pointer)
	v := vm.NewVM()
	v.CPU.R[0] = 0x20000
	v.CPU.R[13] = 0x10000
	v.CPU.PC = 0x8000

	// MOV SP, R0
	// Opcode: MOV (1101), I=0, S=0, Rn=0, Rd=SP, Rm=R0
	opcode := uint32(0xE1A0D000) // 1110 0001 1010 0000 1101 0000 0000 0000
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: SP = R0 = 0x20000
	if v.CPU.R[13] != 0x20000 {
		t.Errorf("expected SP=0x20000, got SP=0x%X", v.CPU.R[13])
	}
}

// ====== LR (R14) Operations ======

func TestMOV_LR_Save(t *testing.T) {
	// MOV R0, LR (save return address)
	v := vm.NewVM()
	v.CPU.R[14] = 0x8100 // LR
	v.CPU.PC = 0x8000

	// MOV R0, LR
	// Opcode: MOV (1101), I=0, S=0, Rn=0, Rd=R0, Rm=LR
	opcode := uint32(0xE1A0000E) // 1110 0001 1010 0000 0000 0000 0000 1110
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: R0 = LR = 0x8100
	if v.CPU.R[0] != 0x8100 {
		t.Errorf("expected R0=0x8100, got R0=0x%X", v.CPU.R[0])
	}
}

func TestMOV_LR_Restore(t *testing.T) {
	// MOV LR, R0 (restore return address)
	v := vm.NewVM()
	v.CPU.R[0] = 0x8200
	v.CPU.R[14] = 0x8100
	v.CPU.PC = 0x8000

	// MOV LR, R0
	// Opcode: MOV (1101), I=0, S=0, Rn=0, Rd=LR, Rm=R0
	opcode := uint32(0xE1A0E000) // 1110 0001 1010 0000 1110 0000 0000 0000
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: LR = R0 = 0x8200
	if v.CPU.R[14] != 0x8200 {
		t.Errorf("expected LR=0x8200, got LR=0x%X", v.CPU.R[14])
	}
}

func TestSTR_LR_Save(t *testing.T) {
	// STR LR, [SP, #-4]! (push LR onto stack with pre-decrement)
	v := vm.NewVM()
	v.CPU.R[13] = 0x10000 // SP
	v.CPU.R[14] = 0x8100  // LR
	v.CPU.PC = 0x8000

	// STR LR, [SP, #-4]!
	// Opcode: STR, I=0, P=1 (pre-indexed), U=0 (subtract), W=1 (writeback), Rn=SP, Rd=LR, offset=4
	opcode := uint32(0xE52DE004) // 1110 0101 0010 1101 1110 0000 0000 0100
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: LR stored at SP - 4 = 0xFFFC, SP updated to 0xFFFC
	val, err := v.Memory.ReadWord(0xFFFC)
	if err != nil {
		t.Fatalf("failed to read stored value: %v", err)
	}
	if val != 0x8100 {
		t.Errorf("expected stored value 0x8100, got 0x%X", val)
	}
	if v.CPU.R[13] != 0xFFFC {
		t.Errorf("expected SP=0xFFFC, got SP=0x%X", v.CPU.R[13])
	}
}

func TestLDR_LR_Restore(t *testing.T) {
	// LDR LR, [SP], #4 (pop LR from stack with post-increment)
	v := vm.NewVM()
	v.CPU.R[13] = 0x10000 // SP
	v.CPU.PC = 0x8000

	// Write return address to stack
	setupCodeWrite(v)
	v.Memory.WriteWord(0x10000, 0x8200)

	// LDR LR, [SP], #4
	// Opcode: LDR, I=0, P=0 (post-indexed), U=1 (add), W=0, Rn=SP, Rd=LR, offset=4
	opcode := uint32(0xE49DE004) // 1110 0100 1001 1101 1110 0000 0000 0100
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Expected: LR loaded from 0x10000, SP incremented to 0x10004
	if v.CPU.R[14] != 0x8200 {
		t.Errorf("expected LR=0x8200, got LR=0x%X", v.CPU.R[14])
	}
	if v.CPU.R[13] != 0x10004 {
		t.Errorf("expected SP=0x10004, got SP=0x%X", v.CPU.R[13])
	}
}
