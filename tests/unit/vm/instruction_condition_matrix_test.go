package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestInstructionConditionMatrix tests every key instruction with every condition code
// This ensures that conditional execution works correctly across all instruction types
// Covers: 16 condition codes × 5 key instructions = 80 tests

// Condition codes: EQ, NE, CS/HS, CC/LO, MI, PL, VS, VC, HI, LS, GE, LT, GT, LE, AL, NV

// Helper function to convert uint32 CPSR value to vm.CPSR struct
func makeCPSR(flags uint32) vm.CPSR {
	return vm.CPSR{
		N: (flags & 0x80000000) != 0, // Bit 31
		Z: (flags & 0x40000000) != 0, // Bit 30
		C: (flags & 0x20000000) != 0, // Bit 29
		V: (flags & 0x10000000) != 0, // Bit 28
	}
}

// MOV instruction tests with all condition codes
func TestMOV_AllConditions(t *testing.T) {
	tests := []struct {
		name       string
		cond       uint32 // Condition code (bits 31:28)
		setupCPSR  uint32 // CPSR value to set before execution
		shouldExec bool   // Should instruction execute?
	}{
		// 0000 = EQ - Equal (Z set)
		{"MOV_EQ_Taken", 0x0, 0x40000000, true},     // Z=1
		{"MOV_EQ_NotTaken", 0x0, 0x00000000, false}, // Z=0

		// 0001 = NE - Not equal (Z clear)
		{"MOV_NE_Taken", 0x1, 0x00000000, true},     // Z=0
		{"MOV_NE_NotTaken", 0x1, 0x40000000, false}, // Z=1

		// 0010 = CS/HS - Carry set / unsigned higher or same (C set)
		{"MOV_CS_Taken", 0x2, 0x20000000, true},     // C=1
		{"MOV_CS_NotTaken", 0x2, 0x00000000, false}, // C=0

		// 0011 = CC/LO - Carry clear / unsigned lower (C clear)
		{"MOV_CC_Taken", 0x3, 0x00000000, true},     // C=0
		{"MOV_CC_NotTaken", 0x3, 0x20000000, false}, // C=1

		// 0100 = MI - Minus / negative (N set)
		{"MOV_MI_Taken", 0x4, 0x80000000, true},     // N=1
		{"MOV_MI_NotTaken", 0x4, 0x00000000, false}, // N=0

		// 0101 = PL - Plus / positive or zero (N clear)
		{"MOV_PL_Taken", 0x5, 0x00000000, true},     // N=0
		{"MOV_PL_NotTaken", 0x5, 0x80000000, false}, // N=1

		// 0110 = VS - Overflow set (V set)
		{"MOV_VS_Taken", 0x6, 0x10000000, true},     // V=1
		{"MOV_VS_NotTaken", 0x6, 0x00000000, false}, // V=0

		// 0111 = VC - Overflow clear (V clear)
		{"MOV_VC_Taken", 0x7, 0x00000000, true},     // V=0
		{"MOV_VC_NotTaken", 0x7, 0x10000000, false}, // V=1

		// 1000 = HI - Unsigned higher (C set and Z clear)
		{"MOV_HI_Taken", 0x8, 0x20000000, true},             // C=1, Z=0
		{"MOV_HI_NotTaken_NoCarry", 0x8, 0x00000000, false}, // C=0, Z=0
		{"MOV_HI_NotTaken_Zero", 0x8, 0x60000000, false},    // C=1, Z=1

		// 1001 = LS - Unsigned lower or same (C clear or Z set)
		{"MOV_LS_Taken_NoCarry", 0x9, 0x00000000, true}, // C=0, Z=0
		{"MOV_LS_Taken_Zero", 0x9, 0x60000000, true},    // C=1, Z=1
		{"MOV_LS_NotTaken", 0x9, 0x20000000, false},     // C=1, Z=0

		// 1010 = GE - Signed greater than or equal (N == V)
		{"MOV_GE_Taken_BothSet", 0xA, 0x90000000, true},   // N=1, V=1
		{"MOV_GE_Taken_BothClear", 0xA, 0x00000000, true}, // N=0, V=0
		{"MOV_GE_NotTaken_NDiff", 0xA, 0x80000000, false}, // N=1, V=0
		{"MOV_GE_NotTaken_VDiff", 0xA, 0x10000000, false}, // N=0, V=1

		// 1011 = LT - Signed less than (N != V)
		{"MOV_LT_Taken_NSet", 0xB, 0x80000000, true},          // N=1, V=0
		{"MOV_LT_Taken_VSet", 0xB, 0x10000000, true},          // N=0, V=1
		{"MOV_LT_NotTaken_BothSet", 0xB, 0x90000000, false},   // N=1, V=1
		{"MOV_LT_NotTaken_BothClear", 0xB, 0x00000000, false}, // N=0, V=0

		// 1100 = GT - Signed greater than (Z clear and N == V)
		{"MOV_GT_Taken_Positive", 0xC, 0x00000000, true}, // Z=0, N=0, V=0
		{"MOV_GT_Taken_Negative", 0xC, 0x90000000, true}, // Z=0, N=1, V=1
		{"MOV_GT_NotTaken_Zero", 0xC, 0x40000000, false}, // Z=1, N=0, V=0
		{"MOV_GT_NotTaken_LT", 0xC, 0x80000000, false},   // Z=0, N=1, V=0

		// 1101 = LE - Signed less than or equal (Z set or N != V)
		{"MOV_LE_Taken_Zero", 0xD, 0x40000000, true}, // Z=1, N=0, V=0
		{"MOV_LE_Taken_LT", 0xD, 0x80000000, true},   // Z=0, N=1, V=0
		{"MOV_LE_NotTaken", 0xD, 0x00000000, false},  // Z=0, N=0, V=0

		// 1110 = AL - Always execute
		{"MOV_AL_Always", 0xE, 0x00000000, true},
		{"MOV_AL_AllFlags", 0xE, 0xF0000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			v.CPU.R[0] = 0x11111111 // Initial R0 value
			v.CPU.CPSR = makeCPSR(tt.setupCPSR)
			v.CPU.PC = 0x8000

			// MOV R0, #0x42 with specified condition
			opcode := (tt.cond << 28) | (0x3A << 20) | (0 << 12) | 0x42

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

			if tt.shouldExec {
				if v.CPU.R[0] != 0x42 {
					t.Errorf("Expected R0=0x42, got R0=0x%08X (instruction should have executed)", v.CPU.R[0])
				}
			} else {
				if v.CPU.R[0] != 0x11111111 {
					t.Errorf("Expected R0=0x11111111 (unchanged), got R0=0x%08X (instruction should NOT have executed)", v.CPU.R[0])
				}
			}
		})
	}
}

// ADD instruction tests with all condition codes
func TestADD_AllConditions(t *testing.T) {
	tests := []struct {
		name       string
		cond       uint32
		setupCPSR  uint32
		shouldExec bool
	}{
		// EQ - Equal (Z set)
		{"ADD_EQ_Taken", 0x0, 0x40000000, true},
		{"ADD_EQ_NotTaken", 0x0, 0x00000000, false},

		// NE - Not equal (Z clear)
		{"ADD_NE_Taken", 0x1, 0x00000000, true},
		{"ADD_NE_NotTaken", 0x1, 0x40000000, false},

		// CS/HS - Carry set (C set)
		{"ADD_CS_Taken", 0x2, 0x20000000, true},
		{"ADD_CS_NotTaken", 0x2, 0x00000000, false},

		// CC/LO - Carry clear (C clear)
		{"ADD_CC_Taken", 0x3, 0x00000000, true},
		{"ADD_CC_NotTaken", 0x3, 0x20000000, false},

		// MI - Minus (N set)
		{"ADD_MI_Taken", 0x4, 0x80000000, true},
		{"ADD_MI_NotTaken", 0x4, 0x00000000, false},

		// PL - Plus (N clear)
		{"ADD_PL_Taken", 0x5, 0x00000000, true},
		{"ADD_PL_NotTaken", 0x5, 0x80000000, false},

		// VS - Overflow set (V set)
		{"ADD_VS_Taken", 0x6, 0x10000000, true},
		{"ADD_VS_NotTaken", 0x6, 0x00000000, false},

		// VC - Overflow clear (V clear)
		{"ADD_VC_Taken", 0x7, 0x00000000, true},
		{"ADD_VC_NotTaken", 0x7, 0x10000000, false},

		// HI - Unsigned higher
		{"ADD_HI_Taken", 0x8, 0x20000000, true},
		{"ADD_HI_NotTaken", 0x8, 0x00000000, false},

		// LS - Unsigned lower or same
		{"ADD_LS_Taken", 0x9, 0x00000000, true},
		{"ADD_LS_NotTaken", 0x9, 0x20000000, false},

		// GE - Signed greater than or equal
		{"ADD_GE_Taken", 0xA, 0x90000000, true},
		{"ADD_GE_NotTaken", 0xA, 0x80000000, false},

		// LT - Signed less than
		{"ADD_LT_Taken", 0xB, 0x80000000, true},
		{"ADD_LT_NotTaken", 0xB, 0x90000000, false},

		// GT - Signed greater than
		{"ADD_GT_Taken", 0xC, 0x00000000, true},
		{"ADD_GT_NotTaken", 0xC, 0x40000000, false},

		// LE - Signed less than or equal
		{"ADD_LE_Taken", 0xD, 0x40000000, true},
		{"ADD_LE_NotTaken", 0xD, 0x00000000, false},

		// AL - Always
		{"ADD_AL_Always", 0xE, 0x00000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			v.CPU.R[0] = 10
			v.CPU.R[1] = 5
			v.CPU.CPSR = makeCPSR(tt.setupCPSR)
			v.CPU.PC = 0x8000

			// ADD R2, R0, R1 with specified condition
			opcode := (tt.cond << 28) | (0x08 << 20) | (0 << 16) | (2 << 12) | 1

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

			if tt.shouldExec {
				if v.CPU.R[2] != 15 {
					t.Errorf("Expected R2=15, got R2=%d (instruction should have executed)", v.CPU.R[2])
				}
			} else {
				if v.CPU.R[2] != 0 {
					t.Errorf("Expected R2=0 (unchanged), got R2=%d (instruction should NOT have executed)", v.CPU.R[2])
				}
			}
		})
	}
}

// LDR instruction tests with all condition codes
func TestLDR_AllConditions(t *testing.T) {
	tests := []struct {
		name       string
		cond       uint32
		setupCPSR  uint32
		shouldExec bool
	}{
		// EQ
		{"LDR_EQ_Taken", 0x0, 0x40000000, true},
		{"LDR_EQ_NotTaken", 0x0, 0x00000000, false},

		// NE
		{"LDR_NE_Taken", 0x1, 0x00000000, true},
		{"LDR_NE_NotTaken", 0x1, 0x40000000, false},

		// CS
		{"LDR_CS_Taken", 0x2, 0x20000000, true},
		{"LDR_CS_NotTaken", 0x2, 0x00000000, false},

		// CC
		{"LDR_CC_Taken", 0x3, 0x00000000, true},
		{"LDR_CC_NotTaken", 0x3, 0x20000000, false},

		// MI
		{"LDR_MI_Taken", 0x4, 0x80000000, true},
		{"LDR_MI_NotTaken", 0x4, 0x00000000, false},

		// PL
		{"LDR_PL_Taken", 0x5, 0x00000000, true},
		{"LDR_PL_NotTaken", 0x5, 0x80000000, false},

		// VS
		{"LDR_VS_Taken", 0x6, 0x10000000, true},
		{"LDR_VS_NotTaken", 0x6, 0x00000000, false},

		// VC
		{"LDR_VC_Taken", 0x7, 0x00000000, true},
		{"LDR_VC_NotTaken", 0x7, 0x10000000, false},

		// HI
		{"LDR_HI_Taken", 0x8, 0x20000000, true},
		{"LDR_HI_NotTaken", 0x8, 0x00000000, false},

		// LS
		{"LDR_LS_Taken", 0x9, 0x00000000, true},
		{"LDR_LS_NotTaken", 0x9, 0x20000000, false},

		// GE
		{"LDR_GE_Taken", 0xA, 0x90000000, true},
		{"LDR_GE_NotTaken", 0xA, 0x80000000, false},

		// LT
		{"LDR_LT_Taken", 0xB, 0x80000000, true},
		{"LDR_LT_NotTaken", 0xB, 0x90000000, false},

		// GT
		{"LDR_GT_Taken", 0xC, 0x00000000, true},
		{"LDR_GT_NotTaken", 0xC, 0x40000000, false},

		// LE
		{"LDR_LE_Taken", 0xD, 0x40000000, true},
		{"LDR_LE_NotTaken", 0xD, 0x00000000, false},

		// AL
		{"LDR_AL_Always", 0xE, 0x00000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			addr := uint32(0x20000) // Data segment start
			setupDataWrite(v)
			v.Memory.WriteWord(addr, 0xDEADBEEF)
			v.CPU.R[1] = addr
			v.CPU.R[0] = 0x11111111 // Initial value
			v.CPU.CPSR = makeCPSR(tt.setupCPSR)
			v.CPU.PC = 0x8000

			// LDR R0, [R1] with specified condition
			opcode := (tt.cond << 28) | (0x59 << 20) | (1 << 16) | (0 << 12) | 0

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

			if tt.shouldExec {
				if v.CPU.R[0] != 0xDEADBEEF {
					t.Errorf("Expected R0=0xDEADBEEF, got R0=0x%08X (instruction should have executed)", v.CPU.R[0])
				}
			} else {
				if v.CPU.R[0] != 0x11111111 {
					t.Errorf("Expected R0=0x11111111 (unchanged), got R0=0x%08X (instruction should NOT have executed)", v.CPU.R[0])
				}
			}
		})
	}
}

// STR instruction tests with all condition codes
func TestSTR_AllConditions(t *testing.T) {
	tests := []struct {
		name       string
		cond       uint32
		setupCPSR  uint32
		shouldExec bool
	}{
		// EQ
		{"STR_EQ_Taken", 0x0, 0x40000000, true},
		{"STR_EQ_NotTaken", 0x0, 0x00000000, false},

		// NE
		{"STR_NE_Taken", 0x1, 0x00000000, true},
		{"STR_NE_NotTaken", 0x1, 0x40000000, false},

		// CS
		{"STR_CS_Taken", 0x2, 0x20000000, true},
		{"STR_CS_NotTaken", 0x2, 0x00000000, false},

		// CC
		{"STR_CC_Taken", 0x3, 0x00000000, true},
		{"STR_CC_NotTaken", 0x3, 0x20000000, false},

		// MI
		{"STR_MI_Taken", 0x4, 0x80000000, true},
		{"STR_MI_NotTaken", 0x4, 0x00000000, false},

		// PL
		{"STR_PL_Taken", 0x5, 0x00000000, true},
		{"STR_PL_NotTaken", 0x5, 0x80000000, false},

		// VS
		{"STR_VS_Taken", 0x6, 0x10000000, true},
		{"STR_VS_NotTaken", 0x6, 0x00000000, false},

		// VC
		{"STR_VC_Taken", 0x7, 0x00000000, true},
		{"STR_VC_NotTaken", 0x7, 0x10000000, false},

		// HI
		{"STR_HI_Taken", 0x8, 0x20000000, true},
		{"STR_HI_NotTaken", 0x8, 0x00000000, false},

		// LS
		{"STR_LS_Taken", 0x9, 0x00000000, true},
		{"STR_LS_NotTaken", 0x9, 0x20000000, false},

		// GE
		{"STR_GE_Taken", 0xA, 0x90000000, true},
		{"STR_GE_NotTaken", 0xA, 0x80000000, false},

		// LT
		{"STR_LT_Taken", 0xB, 0x80000000, true},
		{"STR_LT_NotTaken", 0xB, 0x90000000, false},

		// GT
		{"STR_GT_Taken", 0xC, 0x00000000, true},
		{"STR_GT_NotTaken", 0xC, 0x40000000, false},

		// LE
		{"STR_LE_Taken", 0xD, 0x40000000, true},
		{"STR_LE_NotTaken", 0xD, 0x00000000, false},

		// AL
		{"STR_AL_Always", 0xE, 0x00000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			addr := uint32(0x20000) // Data segment start
			setupDataWrite(v)
			v.Memory.WriteWord(addr, 0xFFFFFFFF) // Pre-fill with pattern
			v.CPU.R[0] = 0x12345678
			v.CPU.R[1] = addr
			v.CPU.CPSR = makeCPSR(tt.setupCPSR)
			v.CPU.PC = 0x8000

			// STR R0, [R1] with specified condition
			opcode := (tt.cond << 28) | (0x58 << 20) | (1 << 16) | (0 << 12) | 0

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

			result, err := v.Memory.ReadWord(addr)
			if err != nil {
				t.Fatalf("failed to read memory: %v", err)
			}
			if tt.shouldExec {
				if result != 0x12345678 {
					t.Errorf("Expected memory[0x%X]=0x12345678, got 0x%08X (instruction should have executed)", addr, result)
				}
			} else {
				if result != 0xFFFFFFFF {
					t.Errorf("Expected memory[0x%X]=0xFFFFFFFF (unchanged), got 0x%08X (instruction should NOT have executed)", addr, result)
				}
			}
		})
	}
}

// B (Branch) instruction tests with all condition codes
func TestB_AllConditions(t *testing.T) {
	tests := []struct {
		name       string
		cond       uint32
		setupCPSR  uint32
		shouldExec bool
	}{
		// EQ
		{"B_EQ_Taken", 0x0, 0x40000000, true},
		{"B_EQ_NotTaken", 0x0, 0x00000000, false},

		// NE
		{"B_NE_Taken", 0x1, 0x00000000, true},
		{"B_NE_NotTaken", 0x1, 0x40000000, false},

		// CS
		{"B_CS_Taken", 0x2, 0x20000000, true},
		{"B_CS_NotTaken", 0x2, 0x00000000, false},

		// CC
		{"B_CC_Taken", 0x3, 0x00000000, true},
		{"B_CC_NotTaken", 0x3, 0x20000000, false},

		// MI
		{"B_MI_Taken", 0x4, 0x80000000, true},
		{"B_MI_NotTaken", 0x4, 0x00000000, false},

		// PL
		{"B_PL_Taken", 0x5, 0x00000000, true},
		{"B_PL_NotTaken", 0x5, 0x80000000, false},

		// VS
		{"B_VS_Taken", 0x6, 0x10000000, true},
		{"B_VS_NotTaken", 0x6, 0x00000000, false},

		// VC
		{"B_VC_Taken", 0x7, 0x00000000, true},
		{"B_VC_NotTaken", 0x7, 0x10000000, false},

		// HI
		{"B_HI_Taken", 0x8, 0x20000000, true},
		{"B_HI_NotTaken", 0x8, 0x00000000, false},

		// LS
		{"B_LS_Taken", 0x9, 0x00000000, true},
		{"B_LS_NotTaken", 0x9, 0x20000000, false},

		// GE
		{"B_GE_Taken", 0xA, 0x90000000, true},
		{"B_GE_NotTaken", 0xA, 0x80000000, false},

		// LT
		{"B_LT_Taken", 0xB, 0x80000000, true},
		{"B_LT_NotTaken", 0xB, 0x90000000, false},

		// GT
		{"B_GT_Taken", 0xC, 0x00000000, true},
		{"B_GT_NotTaken", 0xC, 0x40000000, false},

		// LE
		{"B_LE_Taken", 0xD, 0x40000000, true},
		{"B_LE_NotTaken", 0xD, 0x00000000, false},

		// AL
		{"B_AL_Always", 0xE, 0x00000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			v.CPU.PC = 0x8000
			v.CPU.CPSR = makeCPSR(tt.setupCPSR)

			// B +8 (skip 2 instructions) with specified condition
			// Offset is in words, relative to PC+8
			// offset = (target - (PC + 8)) / 4 = (0x8010 - 0x8008) / 4 = 2
			offset := int32(2) & 0x00FFFFFF // 24-bit signed offset
			opcode := (tt.cond << 28) | (0x0A << 24) | uint32(offset)

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

			// PC after instruction should be PC+4 normally, or target if branch taken
			expectedPC := uint32(0x8004) // PC+4 (not taken)
			if tt.shouldExec {
				expectedPC = 0x8010 // Branch target (0x8000 + 8 + 2*4)
			}

			if v.CPU.PC != expectedPC {
				t.Errorf("Expected PC=0x%X, got PC=0x%X (branch should%s have been taken)",
					expectedPC, v.CPU.PC, map[bool]string{true: "", false: " NOT"}[tt.shouldExec])
			}
		})
	}
}
