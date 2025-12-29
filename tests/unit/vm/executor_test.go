package vm_test

import (
	"fmt"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestResetRegisters verifies that ResetRegisters preserves memory while resetting CPU state
func TestResetRegisters(t *testing.T) {
	v := vm.NewVM()

	// Load some data into memory
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	startAddr := uint32(vm.CodeSegmentStart)
	if err := v.LoadProgram(testData, startAddr); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Modify CPU state
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0xABCDEF00
	v.CPU.PC = startAddr + 8
	v.CPU.CPSR.Z = true
	v.CPU.Cycles = 100

	// Reset registers
	if err := v.ResetRegisters(); err != nil {
		t.Fatalf("ResetRegisters failed: %v", err)
	}

	// Verify CPU state is reset
	if v.CPU.R[0] != 0 {
		t.Errorf("Expected R0=0 after ResetRegisters, got 0x%08X", v.CPU.R[0])
	}
	if v.CPU.R[1] != 0 {
		t.Errorf("Expected R1=0 after ResetRegisters, got 0x%08X", v.CPU.R[1])
	}
	if v.CPU.CPSR.Z {
		t.Error("Expected Z flag to be cleared after ResetRegisters")
	}
	if v.CPU.Cycles != 0 {
		t.Errorf("Expected Cycles=0 after ResetRegisters, got %d", v.CPU.Cycles)
	}

	// Verify memory is preserved
	word, err := v.Memory.ReadWord(startAddr)
	if err != nil {
		t.Fatalf("Failed to read memory: %v", err)
	}
	expected := uint32(0x04030201) // Little-endian
	if word != expected {
		t.Errorf("Expected memory to be preserved: got 0x%08X, want 0x%08X", word, expected)
	}

	// Verify state is halted
	if v.State != vm.StateHalted {
		t.Errorf("Expected StateHalted after ResetRegisters, got %v", v.State)
	}
}

// TestReset verifies that Reset performs complete cleanup of all VM state
func TestReset(t *testing.T) {
	v := vm.NewVM()

	// Set up comprehensive state to verify everything gets cleared

	// 1. Load program into memory
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	startAddr := uint32(vm.CodeSegmentStart)
	if err := v.LoadProgram(testData, startAddr); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// 2. Modify CPU state
	v.CPU.R[0] = 0x12345678
	v.CPU.R[13] = 0x00050000 // SP
	v.CPU.PC = startAddr + 8
	v.CPU.CPSR.Z = true
	v.CPU.Cycles = 100

	// 3. Set entry point and stack top
	v.EntryPoint = startAddr
	v.StackTop = 0x00050000

	// 4. Set exit code
	v.ExitCode = 42

	// 5. Set program arguments
	v.ProgramArguments = []string{"arg1", "arg2"}

	// 6. Add to instruction log
	v.InstructionLog = []uint32{startAddr, startAddr + 4}

	// 7. Set error
	v.LastError = fmt.Errorf("test error")

	// 8. Set execution state
	v.State = vm.StateRunning

	// 9. Set memory write markers
	v.LastMemoryWrite = 0x00010000
	v.HasMemoryWrite = true

	// 10. Enable tracing (if we can access these fields)
	// Note: CodeCoverage, StackTrace, FlagTrace, RegisterTrace are pointers
	// that may not be directly testable, but Reset should nil them out

	// Now reset the VM
	v.Reset()

	// Verify all state is cleared

	// CPU should be reset
	if v.CPU.R[0] != 0 {
		t.Errorf("Expected R0=0 after Reset, got 0x%08X", v.CPU.R[0])
	}
	if v.CPU.R[13] != 0 {
		t.Errorf("Expected SP=0 after Reset, got 0x%08X", v.CPU.R[13])
	}
	if v.CPU.PC != 0 {
		t.Errorf("Expected PC=0 after Reset, got 0x%08X", v.CPU.PC)
	}
	if v.CPU.CPSR.Z {
		t.Error("Expected Z flag cleared after Reset")
	}
	if v.CPU.Cycles != 0 {
		t.Errorf("Expected Cycles=0 after Reset, got %d", v.CPU.Cycles)
	}

	// Memory should be cleared
	word, err := v.Memory.ReadWord(startAddr)
	if err != nil {
		t.Fatalf("Failed to read memory: %v", err)
	}
	if word != 0 {
		t.Errorf("Expected memory cleared after Reset: got 0x%08X, want 0x00000000", word)
	}

	// State should be halted
	if v.State != vm.StateHalted {
		t.Errorf("Expected StateHalted after Reset, got %v", v.State)
	}

	// Entry point and stack top should be cleared
	if v.EntryPoint != 0 {
		t.Errorf("Expected EntryPoint=0 after Reset, got 0x%08X", v.EntryPoint)
	}
	if v.StackTop != 0 {
		t.Errorf("Expected StackTop=0 after Reset, got 0x%08X", v.StackTop)
	}

	// Exit code should be cleared
	if v.ExitCode != 0 {
		t.Errorf("Expected ExitCode=0 after Reset, got %d", v.ExitCode)
	}

	// Program arguments should be cleared
	if v.ProgramArguments != nil {
		t.Errorf("Expected ProgramArguments=nil after Reset, got %v", v.ProgramArguments)
	}

	// Instruction log should be empty
	if len(v.InstructionLog) != 0 {
		t.Errorf("Expected InstructionLog empty after Reset, got length %d", len(v.InstructionLog))
	}

	// Last error should be cleared
	if v.LastError != nil {
		t.Errorf("Expected LastError=nil after Reset, got %v", v.LastError)
	}

	// Memory write markers should be cleared
	if v.LastMemoryWrite != 0 {
		t.Errorf("Expected LastMemoryWrite=0 after Reset, got 0x%08X", v.LastMemoryWrite)
	}
	if v.HasMemoryWrite {
		t.Error("Expected HasMemoryWrite=false after Reset")
	}

	// Note: We cannot easily test file descriptor cleanup without exposing internal fields
	// or adding accessor methods. The implementation will handle this.
}

// TestLoadProgram verifies program loading functionality
func TestLoadProgram(t *testing.T) {
	v := vm.NewVM()

	testData := []byte{
		0xE3, 0xA0, 0x00, 0x01, // MOV R0, #1
		0xE3, 0xA0, 0x10, 0x02, // MOV R1, #2
	}
	startAddr := uint32(vm.CodeSegmentStart)

	err := v.LoadProgram(testData, startAddr)
	if err != nil {
		t.Fatalf("LoadProgram failed: %v", err)
	}

	// Verify PC is set to start address
	if v.CPU.PC != startAddr {
		t.Errorf("Expected PC=0x%08X, got 0x%08X", startAddr, v.CPU.PC)
	}

	// Verify state is halted
	if v.State != vm.StateHalted {
		t.Errorf("Expected StateHalted after LoadProgram, got %v", v.State)
	}

	// Verify memory contents
	for i, b := range testData {
		addr := startAddr + uint32(i)
		readByte, err := v.Memory.ReadByteAt(addr)
		if err != nil {
			t.Fatalf("Failed to read byte at 0x%08X: %v", addr, err)
		}
		if readByte != b {
			t.Errorf("At address 0x%08X: expected byte 0x%02X, got 0x%02X", addr, b, readByte)
		}
	}
}

// TestSetEntryPoint verifies entry point setting
func TestSetEntryPoint(t *testing.T) {
	v := vm.NewVM()

	entryPoint := uint32(0x00008000)
	v.SetEntryPoint(entryPoint)

	if v.CPU.PC != entryPoint {
		t.Errorf("Expected PC=0x%08X after SetEntryPoint, got 0x%08X", entryPoint, v.CPU.PC)
	}
}

// TestGetState verifies state getter
func TestGetState(t *testing.T) {
	v := vm.NewVM()

	// Test initial state
	state := v.GetState()
	if state != vm.StateHalted {
		t.Errorf("Expected initial state to be StateHalted, got %v", state)
	}

	// Change state and verify
	v.SetState(vm.StateRunning)
	state = v.GetState()
	if state != vm.StateRunning {
		t.Errorf("Expected state to be StateRunning, got %v", state)
	}
}

// TestSetState verifies state setter
func TestSetState(t *testing.T) {
	v := vm.NewVM()

	states := []vm.ExecutionState{
		vm.StateRunning,
		vm.StateHalted,
		vm.StateBreakpoint,
		vm.StateError,
	}

	for _, expectedState := range states {
		v.SetState(expectedState)
		if v.State != expectedState {
			t.Errorf("Expected state %v, got %v", expectedState, v.State)
		}
	}
}

// TestGetInstructionHistory verifies instruction logging
func TestGetInstructionHistory(t *testing.T) {
	v := vm.NewVM()

	// Load a simple program
	program := []byte{
		0xE3, 0xA0, 0x00, 0x01, // MOV R0, #1
		0xE3, 0xA0, 0x10, 0x02, // MOV R1, #2
		0xEF, 0x00, 0x00, 0x01, // SWI #1 (exit)
	}

	startAddr := uint32(vm.CodeSegmentStart)
	if err := v.LoadProgram(program, startAddr); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	v.CPU.PC = startAddr

	// Execute a few steps
	for i := 0; i < 3; i++ {
		if err := v.Step(); err != nil {
			// SWI will cause halt
			break
		}
	}

	// Get instruction history
	history := v.GetInstructionHistory()
	if len(history) == 0 {
		t.Error("Expected non-empty instruction history")
	}

	// First instruction should be at start address
	if len(history) > 0 && history[0] != startAddr {
		t.Errorf("Expected first instruction at 0x%08X, got 0x%08X", startAddr, history[0])
	}
}

// TestDumpState verifies state dumping functionality
func TestDumpState(t *testing.T) {
	v := vm.NewVM()

	// Set some known values
	v.CPU.PC = 0x00008000
	validSP := uint32(vm.StackSegmentStart + 0x2000) // 0x00042000
	v.CPU.R[13] = validSP                            // SP
	v.CPU.R[14] = 0x00008100                         // LR
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = false
	v.CPU.CPSR.C = true
	v.CPU.CPSR.V = false
	v.CPU.Cycles = 42

	dump := v.DumpState()

	// Verify dump contains key information
	expectedSubstrings := []string{
		"0x00008000", // PC
		"0x00042000", // SP (updated to valid stack address)
		"0x00008100", // LR
		"N",          // N flag set
		"C",          // C flag set
		"42",         // Cycles
	}

	for _, substr := range expectedSubstrings {
		if !contains(dump, substr) {
			t.Errorf("Expected DumpState to contain '%s', got: %s", substr, dump)
		}
	}
}

// TestSetProgramArguments verifies argument setting
func TestSetProgramArguments(t *testing.T) {
	v := vm.NewVM()

	args := []string{"program", "arg1", "arg2"}
	v.SetProgramArguments(args)

	if len(v.ProgramArguments) != len(args) {
		t.Errorf("Expected %d arguments, got %d", len(args), len(v.ProgramArguments))
	}

	for i, arg := range args {
		if v.ProgramArguments[i] != arg {
			t.Errorf("Argument %d: expected '%s', got '%s'", i, arg, v.ProgramArguments[i])
		}
	}
}

// TestGetExitCode verifies exit code getter
func TestGetExitCode(t *testing.T) {
	v := vm.NewVM()

	// Test default exit code
	if v.GetExitCode() != 0 {
		t.Errorf("Expected default exit code 0, got %d", v.GetExitCode())
	}

	// Set and verify exit code
	v.ExitCode = 42
	if v.GetExitCode() != 42 {
		t.Errorf("Expected exit code 42, got %d", v.GetExitCode())
	}

	// Test negative exit code
	v.ExitCode = -1
	if v.GetExitCode() != -1 {
		t.Errorf("Expected exit code -1, got %d", v.GetExitCode())
	}
}

// TestBootstrap verifies VM bootstrap process
func TestBootstrap(t *testing.T) {
	v := vm.NewVM()

	args := []string{"program", "arg1"}
	entryPoint := uint32(0x00008000)
	v.EntryPoint = entryPoint

	err := v.Bootstrap(args)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}

	// Verify program arguments are stored
	if len(v.ProgramArguments) != len(args) {
		t.Errorf("Expected %d arguments, got %d", len(args), len(v.ProgramArguments))
	}

	// Verify stack pointer is initialized
	expectedStackTop := uint32(vm.StackSegmentStart + vm.StackSegmentSize)
	if v.CPU.GetSP() != expectedStackTop {
		t.Errorf("Expected SP=0x%08X, got 0x%08X", expectedStackTop, v.CPU.GetSP())
	}

	// Verify link register is set to halt address
	if v.CPU.GetLR() != 0xFFFFFFFF {
		t.Errorf("Expected LR=0xFFFFFFFF, got 0x%08X", v.CPU.GetLR())
	}

	// Verify PC is set to entry point
	if v.CPU.PC != entryPoint {
		t.Errorf("Expected PC=0x%08X, got 0x%08X", entryPoint, v.CPU.PC)
	}

	// Verify state is halted
	if v.State != vm.StateHalted {
		t.Errorf("Expected StateHalted, got %v", v.State)
	}

	// Verify exit code is 0
	if v.ExitCode != 0 {
		t.Errorf("Expected ExitCode=0, got %d", v.ExitCode)
	}
}

// TestFindEntryPoint verifies entry point detection
func TestFindEntryPoint(t *testing.T) {
	tests := []struct {
		name     string
		symbols  map[string]uint32
		expected uint32
		wantErr  bool
	}{
		{
			name:     "_start label",
			symbols:  map[string]uint32{"_start": 0x00008000, "main": 0x00008100},
			expected: 0x00008000,
			wantErr:  false,
		},
		{
			name:     "main label (no _start)",
			symbols:  map[string]uint32{"main": 0x00008100, "func": 0x00008200},
			expected: 0x00008100,
			wantErr:  false,
		},
		{
			name:     "__start label",
			symbols:  map[string]uint32{"__start": 0x00008000},
			expected: 0x00008000,
			wantErr:  false,
		},
		{
			name:     "start label",
			symbols:  map[string]uint32{"start": 0x00008000},
			expected: 0x00008000,
			wantErr:  false,
		},
		{
			name:     "no entry point (default)",
			symbols:  map[string]uint32{"func1": 0x00008000, "func2": 0x00008100},
			expected: vm.CodeSegmentStart,
			wantErr:  true,
		},
		{
			name:     "empty symbols",
			symbols:  map[string]uint32{},
			expected: vm.CodeSegmentStart,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()

			addr, err := v.FindEntryPoint(tt.symbols)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindEntryPoint() error = %v, wantErr %v", err, tt.wantErr)
			}

			if addr != tt.expected {
				t.Errorf("Expected entry point 0x%08X, got 0x%08X", tt.expected, addr)
			}

			if v.EntryPoint != tt.expected {
				t.Errorf("Expected VM.EntryPoint=0x%08X, got 0x%08X", tt.expected, v.EntryPoint)
			}
		})
	}
}

// TestRun verifies basic execution loop
func TestRun(t *testing.T) {
	v := vm.NewVM()

	// Load a simple program
	program := []byte{
		0xE3, 0xA0, 0x00, 0x01, // MOV R0, #1
		0xE3, 0xA0, 0x10, 0x02, // MOV R1, #2
	}

	startAddr := uint32(vm.CodeSegmentStart)
	if err := v.LoadProgram(program, startAddr); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	v.CPU.PC = startAddr
	v.CycleLimit = 10 // Set a limit so Run terminates

	// Run will execute until cycle limit
	_ = v.Run()

	// Should be in error state after cycle limit
	if v.State != vm.StateError {
		t.Errorf("Expected StateError after cycle limit, got %v", v.State)
	}

	// Should have executed at least one instruction
	if v.CPU.Cycles == 0 {
		t.Error("Expected some cycles to be executed")
	}
}

// TestRunWithCycleLimit verifies cycle limit enforcement
func TestRunWithCycleLimit(t *testing.T) {
	v := vm.NewVM()

	// Load multiple NOP-like instructions
	program := []byte{
		0xE3, 0xA0, 0x00, 0x00, // MOV R0, #0
		0xE3, 0xA0, 0x10, 0x00, // MOV R1, #0
		0xE3, 0xA0, 0x20, 0x00, // MOV R2, #0
		0xE3, 0xA0, 0x30, 0x00, // MOV R3, #0
		0xE3, 0xA0, 0x40, 0x00, // MOV R4, #0
	}

	startAddr := uint32(vm.CodeSegmentStart)
	if err := v.LoadProgram(program, startAddr); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	v.CPU.PC = startAddr
	v.CycleLimit = 3 // Small limit

	// Run should hit cycle limit
	_ = v.Run()

	// Should be in error state after cycle limit
	if v.State != vm.StateError {
		t.Errorf("Expected StateError after cycle limit, got %v", v.State)
	}

	// Should have executed some cycles
	if v.CPU.Cycles == 0 {
		t.Error("Expected some cycles to be executed")
	}
}

// TestStepWithCycleLimit verifies Step respects cycle limit
func TestStepWithCycleLimit(t *testing.T) {
	v := vm.NewVM()

	// Load a simple instruction
	program := []byte{
		0xE3, 0xA0, 0x00, 0x01, // MOV R0, #1
	}

	startAddr := uint32(vm.CodeSegmentStart)
	if err := v.LoadProgram(program, startAddr); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	v.CPU.PC = startAddr
	v.CycleLimit = 1 // Allow only 1 cycle
	v.CPU.Cycles = 1 // Already at limit

	err := v.Step()
	if err == nil {
		t.Error("Expected error due to cycle limit, got nil")
	}

	if v.State != vm.StateError {
		t.Errorf("Expected StateError after cycle limit in Step, got %v", v.State)
	}
}

// TestStepInErrorState verifies Step fails in error state
func TestStepInErrorState(t *testing.T) {
	v := vm.NewVM()

	// Set VM to error state with a sample error
	v.State = vm.StateError
	v.LastError = fmt.Errorf("test error")

	err := v.Step()
	if err == nil {
		t.Error("Expected Step to fail in error state, got nil")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestNewVM_CycleLimitDefault verifies that NewVM sets CycleLimit to DefaultMaxCycles
// to prevent infinite loops by default (CODE_REVIEW.md §4.2.3)
func TestNewVM_CycleLimitDefault(t *testing.T) {
	v := vm.NewVM()

	// CycleLimit should default to DefaultMaxCycles (1,000,000), not 0 (unlimited)
	if v.CycleLimit != vm.DefaultMaxCycles {
		t.Errorf("Expected CycleLimit=%d (DefaultMaxCycles), got %d", vm.DefaultMaxCycles, v.CycleLimit)
	}

	// Verify the constant has the expected value
	if vm.DefaultMaxCycles != 1000000 {
		t.Errorf("Expected DefaultMaxCycles=1000000, got %d", vm.DefaultMaxCycles)
	}
}
