package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestLDM_WithSBit_RestoresCPSR tests that LDM with S bit and PC restores CPSR from SPSR
func TestLDM_WithSBit_RestoresCPSR(t *testing.T) {
	v := vm.NewVM()
	stackAddr := uint32(vm.StackSegmentStart + 0x1000) // 0x00041000
	v.CPU.R[13] = stackAddr                            // SP
	v.CPU.PC = 0x8000

	// Set current CPSR flags (these should be replaced)
	v.CPU.CPSR.N = false
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.C = false
	v.CPU.CPSR.V = false

	// Set SPSR flags (these should be restored)
	v.CPU.SPSR.N = true
	v.CPU.SPSR.Z = true
	v.CPU.SPSR.C = true
	v.CPU.SPSR.V = true

	setupCodeWrite(v)
	v.Memory.WriteWord(stackAddr, 0xAAAA0000)   // R0
	v.Memory.WriteWord(stackAddr+4, 0x00009000) // PC (return address)

	// LDMIA SP!, {R0, PC}^ - S bit set (bit 22)
	// Base encoding: LDMIA SP!, {R0, PC} = 0xE8BD8001
	// With S bit: 0xE8FD8001 (bit 22 set)
	opcode := uint32(0xE8FD8001) // LDMIA SP!, {R0, PC}^
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Verify R0 was loaded
	if v.CPU.R[0] != 0xAAAA0000 {
		t.Errorf("expected R0=0xAAAA0000, got R0=0x%X", v.CPU.R[0])
	}

	// Verify PC was loaded
	if v.CPU.PC != 0x00009000 {
		t.Errorf("expected PC=0x00009000, got PC=0x%X", v.CPU.PC)
	}

	// Verify CPSR was restored from SPSR
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set (restored from SPSR)")
	}
	if !v.CPU.CPSR.Z {
		t.Error("expected Z flag to be set (restored from SPSR)")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set (restored from SPSR)")
	}
	if !v.CPU.CPSR.V {
		t.Error("expected V flag to be set (restored from SPSR)")
	}
}

// TestLDM_WithoutSBit_PreservesCPSR tests that LDM without S bit doesn't affect CPSR
func TestLDM_WithoutSBit_PreservesCPSR(t *testing.T) {
	v := vm.NewVM()
	stackAddr := uint32(vm.StackSegmentStart + 0x1000); v.CPU.R[13] = stackAddr // SP
	v.CPU.PC = 0x8000

	// Set current CPSR flags (these should be preserved)
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = false

	// Set SPSR flags (these should NOT be restored)
	v.CPU.SPSR.N = false
	v.CPU.SPSR.Z = true
	v.CPU.SPSR.C = false
	v.CPU.SPSR.V = true

	setupCodeWrite(v)
	v.Memory.WriteWord(stackAddr, 0xAAAA0000) // R0
	v.Memory.WriteWord(stackAddr+4, 0x9000)     // PC

	// LDMIA SP!, {R0, PC} - NO S bit
	opcode := uint32(0xE8BD8001) // LDMIA SP!, {R0, PC}
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Verify R0 and PC were loaded
	if v.CPU.R[0] != 0xAAAA0000 {
		t.Errorf("expected R0=0xAAAA0000, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.PC != 0x9000 {
		t.Errorf("expected PC=0x9000, got PC=0x%X", v.CPU.PC)
	}

	// Verify CPSR was NOT changed (should still have original values)
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to remain set (not restored from SPSR)")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag to remain clear (not restored from SPSR)")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to remain set (not restored from SPSR)")
	}
	if v.CPU.CPSR.V {
		t.Error("expected V flag to remain clear (not restored from SPSR)")
	}
}

// TestLDM_WithSBit_NoPCLoad_NoCPSRRestore tests that S bit without PC load doesn't restore CPSR
func TestLDM_WithSBit_NoPCLoad_NoCPSRRestore(t *testing.T) {
	v := vm.NewVM()
	stackAddr := uint32(vm.StackSegmentStart + 0x1000); v.CPU.R[13] = stackAddr // SP
	v.CPU.PC = 0x8000

	// Set current CPSR flags (these should be preserved)
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = false

	// Set SPSR flags (these should NOT be restored since we're not loading PC)
	v.CPU.SPSR.N = false
	v.CPU.SPSR.Z = true
	v.CPU.SPSR.C = false
	v.CPU.SPSR.V = true

	setupCodeWrite(v)
	v.Memory.WriteWord(stackAddr, 0xAAAA0000) // R0
	v.Memory.WriteWord(stackAddr+4, 0xBBBB0001) // R1

	// LDMIA SP!, {R0, R1}^ - S bit set but NO PC in list
	// Base encoding: LDMIA SP!, {R0, R1} = 0xE8BD0003
	// With S bit: 0xE8FD0003 (bit 22 set)
	opcode := uint32(0xE8FD0003) // LDMIA SP!, {R0, R1}^
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Verify registers were loaded
	if v.CPU.R[0] != 0xAAAA0000 {
		t.Errorf("expected R0=0xAAAA0000, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0xBBBB0001 {
		t.Errorf("expected R1=0xBBBB0001, got R1=0x%X", v.CPU.R[1])
	}

	// Verify CPSR was NOT changed (S bit only affects CPSR when loading PC)
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to remain set (S bit without PC load)")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag to remain clear (S bit without PC load)")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to remain set (S bit without PC load)")
	}
	if v.CPU.CPSR.V {
		t.Error("expected V flag to remain clear (S bit without PC load)")
	}
}

// TestLDM_AllFlagCombinations tests all 16 combinations of flag values
func TestLDM_AllFlagCombinations(t *testing.T) {
	testCases := []struct {
		name string
		n    bool
		z    bool
		c    bool
		v    bool
	}{
		{"all_clear", false, false, false, false},
		{"only_v", false, false, false, true},
		{"only_c", false, false, true, false},
		{"c_v", false, false, true, true},
		{"only_z", false, true, false, false},
		{"z_v", false, true, false, true},
		{"z_c", false, true, true, false},
		{"z_c_v", false, true, true, true},
		{"only_n", true, false, false, false},
		{"n_v", true, false, false, true},
		{"n_c", true, false, true, false},
		{"n_c_v", true, false, true, true},
		{"n_z", true, true, false, false},
		{"n_z_v", true, true, false, true},
		{"n_z_c", true, true, true, false},
		{"all_set", true, true, true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := vm.NewVM()
			stackAddr := uint32(vm.StackSegmentStart + 0x1000); v.CPU.R[13] = stackAddr // SP
			v.CPU.PC = 0x8000

			// Set CPSR to opposite of test values (should be overwritten)
			v.CPU.CPSR.N = !tc.n
			v.CPU.CPSR.Z = !tc.z
			v.CPU.CPSR.C = !tc.c
			v.CPU.CPSR.V = !tc.v

			// Set SPSR to test values (should be restored)
			v.CPU.SPSR.N = tc.n
			v.CPU.SPSR.Z = tc.z
			v.CPU.SPSR.C = tc.c
			v.CPU.SPSR.V = tc.v

			setupCodeWrite(v)
			v.Memory.WriteWord(stackAddr, 0x9000) // PC

			// LDMIA SP!, {PC}^ - S bit set
			opcode := uint32(0xE8FD8000) // LDMIA SP!, {PC}^
			v.Memory.WriteWord(0x8000, opcode)
			v.Step()

			// Verify PC was loaded
			if v.CPU.PC != 0x9000 {
				t.Errorf("expected PC=0x9000, got PC=0x%X", v.CPU.PC)
			}

			// Verify all flags were restored correctly
			if v.CPU.CPSR.N != tc.n {
				t.Errorf("expected N=%v, got N=%v", tc.n, v.CPU.CPSR.N)
			}
			if v.CPU.CPSR.Z != tc.z {
				t.Errorf("expected Z=%v, got Z=%v", tc.z, v.CPU.CPSR.Z)
			}
			if v.CPU.CPSR.C != tc.c {
				t.Errorf("expected C=%v, got C=%v", tc.c, v.CPU.CPSR.C)
			}
			if v.CPU.CPSR.V != tc.v {
				t.Errorf("expected V=%v, got V=%v", tc.v, v.CPU.CPSR.V)
			}
		})
	}
}

// TestLDM_SBit_MultipleRegisters tests S bit with multiple registers including PC
func TestLDM_SBit_MultipleRegisters(t *testing.T) {
	v := vm.NewVM()
	stackAddr := uint32(vm.StackSegmentStart + 0x1000); v.CPU.R[13] = stackAddr // SP
	v.CPU.PC = 0x8000

	// Set current CPSR (should be replaced)
	v.CPU.CPSR.N = false
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.C = false
	v.CPU.CPSR.V = false

	// Set SPSR (should be restored)
	v.CPU.SPSR.N = true
	v.CPU.SPSR.Z = false
	v.CPU.SPSR.C = true
	v.CPU.SPSR.V = false

	setupCodeWrite(v)
	v.Memory.WriteWord(stackAddr, 0x11111111) // R0
	v.Memory.WriteWord(stackAddr+4, 0x22222222) // R1
	v.Memory.WriteWord(stackAddr+8, 0x33333333) // R2
	v.Memory.WriteWord(stackAddr+12, 0x44444444) // R3
	v.Memory.WriteWord(stackAddr+16, 0x9000)     // PC

	// LDMIA SP!, {R0-R3, PC}^ - S bit set
	// Register list: R0-R3, PC = 0x800F
	// With S bit: 0xE8FD800F
	opcode := uint32(0xE8FD800F) // LDMIA SP!, {R0-R3, PC}^
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Verify all registers were loaded
	if v.CPU.R[0] != 0x11111111 {
		t.Errorf("expected R0=0x11111111, got R0=0x%X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0x22222222 {
		t.Errorf("expected R1=0x22222222, got R1=0x%X", v.CPU.R[1])
	}
	if v.CPU.R[2] != 0x33333333 {
		t.Errorf("expected R2=0x33333333, got R2=0x%X", v.CPU.R[2])
	}
	if v.CPU.R[3] != 0x44444444 {
		t.Errorf("expected R3=0x44444444, got R3=0x%X", v.CPU.R[3])
	}
	if v.CPU.PC != 0x9000 {
		t.Errorf("expected PC=0x9000, got PC=0x%X", v.CPU.PC)
	}

	// Verify CPSR was restored
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to be set")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag to be clear")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to be set")
	}
	if v.CPU.CPSR.V {
		t.Error("expected V flag to be clear")
	}
}

// TestSTM_WithSBit_NoEffect tests that STM with S bit has no special CPSR behavior
func TestSTM_WithSBit_NoEffect(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[0] = 0xAAAA0000
	v.CPU.R[1] = 0xBBBB0001
	stackAddr := uint32(vm.StackSegmentStart + 0x1000); v.CPU.R[13] = stackAddr // SP
	v.CPU.PC = 0x8000

	// Set CPSR (should remain unchanged)
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = false

	// Set SPSR (should remain unchanged)
	v.CPU.SPSR.N = false
	v.CPU.SPSR.Z = true
	v.CPU.SPSR.C = false
	v.CPU.SPSR.V = true

	setupCodeWrite(v)

	// STMDB SP!, {R0, R1}^ - S bit set (but no special behavior for STM)
	// Base encoding: STMDB SP!, {R0, R1} = 0xE92D0003
	// With S bit: 0xE96D0003 (bit 22 set)
	opcode := uint32(0xE96D0003) // STMDB SP!, {R0, R1}^
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Verify registers were stored (SP decremented by 8 for 2 registers)
	expectedAddr := stackAddr - 8
	val0, _ := v.Memory.ReadWord(expectedAddr)
	val1, _ := v.Memory.ReadWord(expectedAddr + 4)
	if val0 != 0xAAAA0000 {
		t.Errorf("expected memory[0x%X]=0xAAAA0000, got 0x%X", expectedAddr, val0)
	}
	if val1 != 0xBBBB0001 {
		t.Errorf("expected memory[0x%X]=0xBBBB0001, got 0x%X", expectedAddr+4, val1)
	}

	// Verify CPSR was not affected
	if !v.CPU.CPSR.N {
		t.Error("expected N flag to remain set")
	}
	if v.CPU.CPSR.Z {
		t.Error("expected Z flag to remain clear")
	}
	if !v.CPU.CPSR.C {
		t.Error("expected C flag to remain set")
	}
	if v.CPU.CPSR.V {
		t.Error("expected V flag to remain clear")
	}

	// Verify SPSR was not affected
	if v.CPU.SPSR.N {
		t.Error("expected SPSR N flag to remain clear")
	}
	if !v.CPU.SPSR.Z {
		t.Error("expected SPSR Z flag to remain set")
	}
	if v.CPU.SPSR.C {
		t.Error("expected SPSR C flag to remain clear")
	}
	if !v.CPU.SPSR.V {
		t.Error("expected SPSR V flag to remain set")
	}
}

// TestSaveCPSR_HelperMethod tests the SaveCPSR helper method
func TestSaveCPSR_HelperMethod(t *testing.T) {
	cpu := vm.NewCPU()

	// Set CPSR flags
	cpu.CPSR.N = true
	cpu.CPSR.Z = false
	cpu.CPSR.C = true
	cpu.CPSR.V = false

	// Save CPSR to SPSR
	cpu.SaveCPSR()

	// Verify SPSR now has the same values
	if cpu.SPSR.N != true {
		t.Error("expected SPSR.N=true after SaveCPSR()")
	}
	if cpu.SPSR.Z != false {
		t.Error("expected SPSR.Z=false after SaveCPSR()")
	}
	if cpu.SPSR.C != true {
		t.Error("expected SPSR.C=true after SaveCPSR()")
	}
	if cpu.SPSR.V != false {
		t.Error("expected SPSR.V=false after SaveCPSR()")
	}

	// Verify CPSR is unchanged
	if cpu.CPSR.N != true {
		t.Error("expected CPSR.N=true (unchanged)")
	}
	if cpu.CPSR.Z != false {
		t.Error("expected CPSR.Z=false (unchanged)")
	}
	if cpu.CPSR.C != true {
		t.Error("expected CPSR.C=true (unchanged)")
	}
	if cpu.CPSR.V != false {
		t.Error("expected CPSR.V=false (unchanged)")
	}
}

// TestRestoreCPSR_HelperMethod tests the RestoreCPSR helper method
func TestRestoreCPSR_HelperMethod(t *testing.T) {
	cpu := vm.NewCPU()

	// Set SPSR flags
	cpu.SPSR.N = true
	cpu.SPSR.Z = false
	cpu.SPSR.C = true
	cpu.SPSR.V = false

	// Set CPSR to different values
	cpu.CPSR.N = false
	cpu.CPSR.Z = true
	cpu.CPSR.C = false
	cpu.CPSR.V = true

	// Restore CPSR from SPSR
	cpu.RestoreCPSR()

	// Verify CPSR now has SPSR values
	if cpu.CPSR.N != true {
		t.Error("expected CPSR.N=true after RestoreCPSR()")
	}
	if cpu.CPSR.Z != false {
		t.Error("expected CPSR.Z=false after RestoreCPSR()")
	}
	if cpu.CPSR.C != true {
		t.Error("expected CPSR.C=true after RestoreCPSR()")
	}
	if cpu.CPSR.V != false {
		t.Error("expected CPSR.V=false after RestoreCPSR()")
	}

	// Verify SPSR is unchanged
	if cpu.SPSR.N != true {
		t.Error("expected SPSR.N=true (unchanged)")
	}
	if cpu.SPSR.Z != false {
		t.Error("expected SPSR.Z=false (unchanged)")
	}
	if cpu.SPSR.C != true {
		t.Error("expected SPSR.C=true (unchanged)")
	}
	if cpu.SPSR.V != false {
		t.Error("expected SPSR.V=false (unchanged)")
	}
}

// TestIntegration_ExceptionHandlerSimulation tests a complete exception handler flow
func TestIntegration_ExceptionHandlerSimulation(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000
	stackAddr := uint32(vm.StackSegmentStart + 0x1000); v.CPU.R[13] = stackAddr // SP

	// Set initial CPSR state (normal execution)
	v.CPU.CPSR.N = false
	v.CPU.CPSR.Z = true
	v.CPU.CPSR.C = false
	v.CPU.CPSR.V = false

	setupCodeWrite(v)

	// Step 1: Save CPSR to SPSR (simulating exception entry)
	v.CPU.SaveCPSR()

	// Step 2: Change CPSR (simulating exception handler modifying flags)
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = true

	// Step 3: Prepare stack for exception return
	v.Memory.WriteWord(stackAddr, 0x9000) // Return address (PC)

	// Step 4: Execute LDM with S bit to return from exception
	// LDMIA SP!, {PC}^
	opcode := uint32(0xE8FD8000)
	v.Memory.WriteWord(0x8000, opcode)
	v.Step()

	// Verify PC was restored
	if v.CPU.PC != 0x9000 {
		t.Errorf("expected PC=0x9000, got PC=0x%X", v.CPU.PC)
	}

	// Verify CPSR was restored to pre-exception state
	if v.CPU.CPSR.N != false {
		t.Error("expected N flag restored to false")
	}
	if v.CPU.CPSR.Z != true {
		t.Error("expected Z flag restored to true")
	}
	if v.CPU.CPSR.C != false {
		t.Error("expected C flag restored to false")
	}
	if v.CPU.CPSR.V != false {
		t.Error("expected V flag restored to false")
	}
}
