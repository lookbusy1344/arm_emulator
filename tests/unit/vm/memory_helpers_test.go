package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestInitializeStack verifies stack initialization
func TestInitializeStack(t *testing.T) {
	v := vm.NewVM()

	stackTop := uint32(0x00030000)
	v.InitializeStack(stackTop)

	if v.CPU.GetSP() != stackTop {
		t.Errorf("Expected SP=0x%08X after InitializeStack, got 0x%08X", stackTop, v.CPU.GetSP())
	}
}

// TestMemorySegmentNaming verifies segment name retrieval
func TestMemorySegmentNaming(t *testing.T) {
	v := vm.NewVM()

	// Verify we have expected segments
	segmentNames := make(map[string]bool)
	for _, seg := range v.Memory.Segments {
		segmentNames[seg.Name] = true
	}

	expectedSegments := []string{"code", "data", "stack"}
	for _, name := range expectedSegments {
		if !segmentNames[name] {
			t.Errorf("Expected segment '%s' to exist", name)
		}
	}
}

// TestCPURegistersRange verifies register access bounds
func TestCPURegistersRange(t *testing.T) {
	v := vm.NewVM()

	// Test setting and getting all general purpose registers
	for i := 0; i < 15; i++ {
		expectedValue := uint32(i * 100)
		v.CPU.SetRegister(i, expectedValue)
		gotValue := v.CPU.GetRegister(i)

		if gotValue != expectedValue {
			t.Errorf("Register R%d: expected 0x%08X, got 0x%08X", i, expectedValue, gotValue)
		}
	}
}

// TestCPURegistersInvalidRange verifies register access with invalid indices
func TestCPURegistersInvalidRange(t *testing.T) {
	v := vm.NewVM()

	// Test getting invalid register returns 0
	value := v.CPU.GetRegister(-1)
	if value != 0 {
		t.Errorf("Expected 0 for invalid register -1, got 0x%08X", value)
	}

	value = v.CPU.GetRegister(20)
	if value != 0 {
		t.Errorf("Expected 0 for invalid register 20, got 0x%08X", value)
	}

	// Test setting invalid register doesn't panic
	v.CPU.SetRegister(-1, 0x12345678)
	v.CPU.SetRegister(20, 0x12345678)
	// If we get here without panic, test passes
}

// TestCPUPCReadOffset verifies PC read returns PC+8
func TestCPUPCReadOffset(t *testing.T) {
	v := vm.NewVM()

	v.CPU.PC = 0x00008000

	// Reading R15 (PC) should return PC+8 due to ARM pipeline
	value := v.CPU.GetRegister(15)
	expected := uint32(0x00008008)

	if value != expected {
		t.Errorf("Expected GetRegister(15) to return PC+8=0x%08X, got 0x%08X", expected, value)
	}
}

// TestCPUSetPC verifies PC can be set via SetRegister
func TestCPUSetPC(t *testing.T) {
	v := vm.NewVM()

	newPC := uint32(0x00009000)
	v.CPU.SetRegister(15, newPC)

	if v.CPU.PC != newPC {
		t.Errorf("Expected PC=0x%08X after SetRegister(15), got 0x%08X", newPC, v.CPU.PC)
	}
}

// TestCPUIncrementPC verifies PC increment
func TestCPUIncrementPC(t *testing.T) {
	v := vm.NewVM()

	initialPC := uint32(0x00008000)
	v.CPU.PC = initialPC

	v.CPU.IncrementPC()

	expected := initialPC + 4
	if v.CPU.PC != expected {
		t.Errorf("Expected PC=0x%08X after IncrementPC, got 0x%08X", expected, v.CPU.PC)
	}
}

// TestCPUBranch verifies branch operation
func TestCPUBranch(t *testing.T) {
	v := vm.NewVM()

	v.CPU.PC = 0x00008000
	targetAddr := uint32(0x00009000)

	v.CPU.Branch(targetAddr)

	if v.CPU.PC != targetAddr {
		t.Errorf("Expected PC=0x%08X after Branch, got 0x%08X", targetAddr, v.CPU.PC)
	}
}

// TestCPUBranchWithLink verifies branch with link saves return address
func TestCPUBranchWithLink(t *testing.T) {
	v := vm.NewVM()

	initialPC := uint32(0x00008000)
	v.CPU.PC = initialPC
	targetAddr := uint32(0x00009000)

	v.CPU.BranchWithLink(targetAddr)

	// LR should contain return address (PC + 4)
	expectedLR := initialPC + 4
	if v.CPU.GetLR() != expectedLR {
		t.Errorf("Expected LR=0x%08X after BranchWithLink, got 0x%08X", expectedLR, v.CPU.GetLR())
	}

	// PC should be at target
	if v.CPU.PC != targetAddr {
		t.Errorf("Expected PC=0x%08X after BranchWithLink, got 0x%08X", targetAddr, v.CPU.PC)
	}
}

// TestCPUIncrementCycles verifies cycle counting
func TestCPUIncrementCycles(t *testing.T) {
	v := vm.NewVM()

	initialCycles := v.CPU.Cycles

	v.CPU.IncrementCycles(1)
	if v.CPU.Cycles != initialCycles+1 {
		t.Errorf("Expected %d cycles after IncrementCycles(1), got %d", initialCycles+1, v.CPU.Cycles)
	}

	v.CPU.IncrementCycles(10)
	if v.CPU.Cycles != initialCycles+11 {
		t.Errorf("Expected %d cycles after IncrementCycles(10), got %d", initialCycles+11, v.CPU.Cycles)
	}
}

// TestCPSRFlags verifies CPSR flag manipulation
func TestCPSRFlags(t *testing.T) {
	var cpsr vm.CPSR

	// Test all flags initially false
	if cpsr.N || cpsr.Z || cpsr.C || cpsr.V {
		t.Error("Expected all CPSR flags to be false initially")
	}

	// Set each flag
	cpsr.N = true
	cpsr.Z = true
	cpsr.C = true
	cpsr.V = true

	if !cpsr.N || !cpsr.Z || !cpsr.C || !cpsr.V {
		t.Error("Expected all CPSR flags to be true after setting")
	}

	// Clear flags
	cpsr.N = false
	cpsr.Z = false
	cpsr.C = false
	cpsr.V = false

	if cpsr.N || cpsr.Z || cpsr.C || cpsr.V {
		t.Error("Expected all CPSR flags to be false after clearing")
	}
}

// TestCPSRToFromUint32 verifies CPSR serialization
func TestCPSRToFromUint32(t *testing.T) {
	cpsr := vm.CPSR{
		N: true,
		Z: false,
		C: true,
		V: false,
	}

	// Convert to uint32
	value := cpsr.ToUint32()

	// Create new CPSR from uint32
	newCPSR := vm.CPSR{}
	newCPSR.FromUint32(value)

	// Verify flags match
	if newCPSR.N != cpsr.N {
		t.Errorf("N flag mismatch: expected %v, got %v", cpsr.N, newCPSR.N)
	}
	if newCPSR.Z != cpsr.Z {
		t.Errorf("Z flag mismatch: expected %v, got %v", cpsr.Z, newCPSR.Z)
	}
	if newCPSR.C != cpsr.C {
		t.Errorf("C flag mismatch: expected %v, got %v", cpsr.C, newCPSR.C)
	}
	if newCPSR.V != cpsr.V {
		t.Errorf("V flag mismatch: expected %v, got %v", cpsr.V, newCPSR.V)
	}
}

// TestMemoryLoadBytesUnsafe verifies unsafe byte loading
func TestMemoryLoadBytesUnsafe(t *testing.T) {
	v := vm.NewVM()

	data := []byte{0x01, 0x02, 0x03, 0x04}
	startAddr := uint32(vm.CodeSegmentStart)

	err := v.Memory.LoadBytesUnsafe(startAddr, data)
	if err != nil {
		t.Fatalf("LoadBytesUnsafe failed: %v", err)
	}

	// Verify data was loaded
	for i, expected := range data {
		addr := startAddr + uint32(i)
		actual, err := v.Memory.ReadByteAt(addr)
		if err != nil {
			t.Fatalf("Failed to read byte at 0x%08X: %v", addr, err)
		}
		if actual != expected {
			t.Errorf("At 0x%08X: expected 0x%02X, got 0x%02X", addr, expected, actual)
		}
	}
}

// TestMemoryGetBytes verifies byte retrieval
func TestMemoryGetBytes(t *testing.T) {
	v := vm.NewVM()

	// Load some data
	data := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	startAddr := uint32(vm.CodeSegmentStart)
	v.Memory.LoadBytesUnsafe(startAddr, data)

	// Get bytes back
	retrieved, err := v.Memory.GetBytes(startAddr, uint32(len(data)))
	if err != nil {
		t.Fatalf("GetBytes failed: %v", err)
	}

	// Verify data matches
	for i, expected := range data {
		if retrieved[i] != expected {
			t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, expected, retrieved[i])
		}
	}
}

// TestMemoryMakeCodeReadOnly verifies making code segment read-only
func TestMemoryMakeCodeReadOnly(t *testing.T) {
	v := vm.NewVM()

	// Load some data
	data := []byte{0x01, 0x02, 0x03, 0x04}
	startAddr := uint32(vm.CodeSegmentStart)
	v.Memory.LoadBytesUnsafe(startAddr, data)

	// Make code read-only
	v.Memory.MakeCodeReadOnly()

	// Try to write (should fail)
	err := v.Memory.WriteByteAt(startAddr, 0xFF)
	if err == nil {
		t.Error("Expected write to fail on read-only code segment")
	}

	// Reading should still work
	_, err = v.Memory.ReadByteAt(startAddr)
	if err != nil {
		t.Errorf("Reading from read-only code should work: %v", err)
	}
}

// TestMemoryResetHeap verifies heap reset
func TestMemoryResetHeap(t *testing.T) {
	v := vm.NewVM()

	// Call ResetHeap (it resets internal heap state)
	v.Memory.ResetHeap()

	// If we get here without panic, the function works
	// The heap management uses package-level variables that we can't easily inspect
}
