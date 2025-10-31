package service_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestNewDebuggerService(t *testing.T) {
	machine := vm.NewVM()
	svc := service.NewDebuggerService(machine)

	if svc == nil {
		t.Fatal("expected service instance, got nil")
	}

	if svc.GetVM() != machine {
		t.Error("service VM mismatch")
	}
}

func TestDebuggerService_LoadProgram(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	// Parse simple program with .org directive
	p := parser.NewParser(".org 0x8000\n_start:\nMOV R0, #42\nSWI #0", "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Load into service
	err = svc.LoadProgram(program, 0x8000)
	if err != nil {
		t.Fatalf("LoadProgram failed: %v", err)
	}

	// Verify PC set correctly
	if machine.CPU.PC != 0x8000 {
		t.Errorf("expected PC=0x8000, got 0x%08X", machine.CPU.PC)
	}
}

func TestDebuggerService_GetSourceMap(t *testing.T) {
	// Create service with VM
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	// Load a simple program
	program := `
.org 0x8000
main:
    MOV R0, #42
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Get source map
	sourceMap := svc.GetSourceMap()

	// Should have entries for the instructions
	if len(sourceMap) == 0 {
		t.Error("Expected non-empty source map")
	}

	// Check that main label exists at 0x8000
	if source, ok := sourceMap[0x00008000]; ok {
		if source != "    MOV R0, #42" {
			t.Errorf("Expected '    MOV R0, #42', got '%s'", source)
		}
	} else {
		t.Error("Expected source line at address 0x00008000")
	}
}

func TestDebuggerService_GetSourceMap_DefensiveCopy(t *testing.T) {
	// Create service with VM
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	// Load a simple program
	program := `
.org 0x8000
main:
    MOV R0, #42
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Get source map and store original value
	sourceMap := svc.GetSourceMap()
	originalLine := sourceMap[0x00008000]

	// Modify the returned map
	sourceMap[0x00008000] = "MODIFIED LINE"
	sourceMap[0x00009999] = "NEW ENTRY"

	// Get source map again and verify it's unchanged
	sourceMap2 := svc.GetSourceMap()
	if sourceMap2[0x00008000] != originalLine {
		t.Errorf("Source map was modified externally - defensive copy failed. Expected '%s', got '%s'",
			originalLine, sourceMap2[0x00008000])
	}
	if _, exists := sourceMap2[0x00009999]; exists {
		t.Error("New entry added to external map affected internal state - defensive copy failed")
	}
}

func TestDebuggerService_GetSymbolForAddress(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #1
loop:
    ADD R0, R0, #1
    B loop
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Get symbol for main (should be at 0x8000)
	symbol := svc.GetSymbolForAddress(0x00008000)
	if symbol != "main" {
		t.Errorf("Expected symbol 'main', got '%s'", symbol)
	}

	// Get symbol for loop (should be at 0x8004)
	symbol = svc.GetSymbolForAddress(0x00008004)
	if symbol != "loop" {
		t.Errorf("Expected symbol 'loop', got '%s'", symbol)
	}

	// Get symbol for address without label
	symbol = svc.GetSymbolForAddress(0x00008008)
	if symbol != "" {
		t.Errorf("Expected empty string, got '%s'", symbol)
	}
}

func TestDebuggerService_GetDisassembly(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #42
    MOV R1, #10
    ADD R2, R0, R1
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Get disassembly starting at main
	lines := svc.GetDisassembly(0x00008000, 3)

	if len(lines) != 3 {
		t.Errorf("Expected 3 disassembly lines, got %d", len(lines))
	}

	// Check first line is at main
	if lines[0].Address != 0x00008000 {
		t.Errorf("Expected address 0x00008000, got 0x%08X", lines[0].Address)
	}
	if lines[0].Symbol != "main" {
		t.Errorf("Expected symbol 'main', got '%s'", lines[0].Symbol)
	}

	// Check opcodes are valid (non-zero)
	for i, line := range lines {
		if line.Opcode == 0 {
			t.Errorf("Line %d has zero opcode", i)
		}
	}
}

func TestDebuggerService_GetDisassembly_MemoryError(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #1
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Try to disassemble from an invalid address (should handle gracefully)
	lines := svc.GetDisassembly(0x99999000, 5)

	// Should return empty or partial results (graceful handling)
	if len(lines) > 0 {
		t.Errorf("Expected 0 lines from invalid address, got %d", len(lines))
	}
}

func TestDebuggerService_GetDisassembly_EdgeCases(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #42
    MOV R1, #10
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	tests := []struct {
		name        string
		address     uint32
		count       int
		expectedLen int
		description string
	}{
		{
			name:        "count_zero",
			address:     0x00008000,
			count:       0,
			expectedLen: 0,
			description: "count=0 should return empty slice",
		},
		{
			name:        "count_one",
			address:     0x00008000,
			count:       1,
			expectedLen: 1,
			description: "count=1 should return 1 line",
		},
		{
			name:        "negative_count",
			address:     0x00008000,
			count:       -1,
			expectedLen: 0,
			description: "negative count should return empty slice",
		},
		{
			name:        "count_exceeds_max",
			address:     0x00008000,
			count:       1001,
			expectedLen: 0,
			description: "count > 1000 should return empty slice",
		},
		{
			name:        "misaligned_address_plus_1",
			address:     0x00008001,
			count:       5,
			expectedLen: 0,
			description: "misaligned address (addr+1) should return empty slice",
		},
		{
			name:        "misaligned_address_plus_2",
			address:     0x00008002,
			count:       5,
			expectedLen: 0,
			description: "misaligned address (addr+2) should return empty slice",
		},
		{
			name:        "misaligned_address_plus_3",
			address:     0x00008003,
			count:       5,
			expectedLen: 0,
			description: "misaligned address (addr+3) should return empty slice",
		},
		{
			name:        "aligned_address_valid",
			address:     0x00008000,
			count:       2,
			expectedLen: 2,
			description: "properly aligned address should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := svc.GetDisassembly(tt.address, tt.count)
			if len(lines) != tt.expectedLen {
				t.Errorf("%s: expected %d lines, got %d", tt.description, tt.expectedLen, len(lines))
			}

			// For valid cases, verify opcodes are non-zero
			if tt.expectedLen > 0 && len(lines) > 0 {
				for i, line := range lines {
					if line.Opcode == 0 {
						t.Errorf("Line %d has zero opcode", i)
					}
				}
			}
		})
	}
}

func TestDebuggerService_GetStack(t *testing.T) {
	machine := vm.NewVM()
	// Use stack within valid range (0x00040000-0x0004FFFF, stack grows down)
	machine.InitializeStack(0x0004FFF0)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #0x12
    MOV R2, SP
    STR R0, [R2]
    MOV R1, #0x56
    STR R1, [R2, #4]
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Execute until we've stored values (5 instructions)
	for i := 0; i < 5; i++ {
		svc.Step()
	}

	// Get stack contents
	stack := svc.GetStack(0, 4)

	if len(stack) == 0 {
		t.Error("Expected non-empty stack")
	}

	// Stack should contain stored values
	foundValue := false
	for _, entry := range stack {
		if entry.Value == 0x56 || entry.Value == 0x12 {
			foundValue = true
			break
		}
	}

	if !foundValue {
		t.Errorf("Expected to find pushed values on stack, got %d entries", len(stack))
		for i, entry := range stack {
			t.Logf("Stack[%d]: Addr=0x%08X Value=0x%08X Symbol=%s", i, entry.Address, entry.Value, entry.Symbol)
		}
	}
}

func TestDebuggerService_GetStack_EdgeCases(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x0004FFF0)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #42
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	tests := []struct {
		name        string
		offset      int
		count       int
		expectedLen int
		description string
	}{
		{
			name:        "large_positive_offset",
			offset:      100001,
			count:       10,
			expectedLen: 0,
			description: "offset > 100000 should return empty (wraparound protection)",
		},
		{
			name:        "large_negative_offset",
			offset:      -100001,
			count:       10,
			expectedLen: 0,
			description: "offset < -100000 should return empty (wraparound protection)",
		},
		{
			name:        "offset_zero_count_zero",
			offset:      0,
			count:       0,
			expectedLen: 0,
			description: "count=0 should return empty",
		},
		{
			name:        "count_exceeds_max",
			offset:      0,
			count:       1001,
			expectedLen: 0,
			description: "count > 1000 should return empty",
		},
		{
			name:        "negative_count",
			offset:      0,
			count:       -1,
			expectedLen: 0,
			description: "negative count should return empty",
		},
		{
			name:        "max_valid_offset_positive",
			offset:      100000,
			count:       1,
			expectedLen: 0,
			description: "offset=100000 should be accepted but likely fail on memory read",
		},
		{
			name:        "max_valid_offset_negative",
			offset:      -100000,
			count:       1,
			expectedLen: 0,
			description: "offset=-100000 should be accepted but likely fail on memory read",
		},
		{
			name:        "valid_offset_and_count",
			offset:      0,
			count:       1,
			expectedLen: 1,
			description: "valid offset=0, count=1 should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := svc.GetStack(tt.offset, tt.count)
			if len(stack) != tt.expectedLen {
				t.Errorf("%s: expected %d entries, got %d", tt.description, tt.expectedLen, len(stack))
			}
		})
	}
}

func TestDebuggerService_GetStack_IntegerOverflow(t *testing.T) {
	machine := vm.NewVM()
	// Set SP near the top of address space to test wraparound
	machine.CPU.R[13] = 0xFFFFFFF0
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #42
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	tests := []struct {
		name        string
		offset      int
		count       int
		expectedLen int
		description string
	}{
		{
			name:        "wraparound_positive_offset",
			offset:      100,
			count:       10,
			expectedLen: 0,
			description: "SP=0xFFFFFFF0 + offset=400 should wraparound and return empty",
		},
		{
			name:        "wraparound_negative_offset_large_sp",
			offset:      -1073741824, // This would cause wraparound with SP near top
			count:       10,
			expectedLen: 0,
			description: "Large negative offset with high SP should be rejected",
		},
		{
			name:        "small_positive_offset_near_max",
			offset:      1,
			count:       1,
			expectedLen: 0,
			description: "SP near max + small offset should wraparound and return empty or fail on read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := svc.GetStack(tt.offset, tt.count)
			if len(stack) != tt.expectedLen {
				t.Errorf("%s: expected %d entries, got %d", tt.description, tt.expectedLen, len(stack))
			}
		})
	}
}

func TestDebuggerService_GetOutput(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #42
    SWI #0x03  ; Write integer
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Execute program
	err = svc.RunUntilHalt()
	if err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}

	// Get output
	output := svc.GetOutput()

	if output == "" {
		t.Error("Expected non-empty output")
	}

	// Second call should return empty (buffer cleared)
	output2 := svc.GetOutput()
	if output2 != "" {
		t.Errorf("Expected empty output after clear, got '%s'", output2)
	}
}

func TestDebuggerService_StepOver(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    BL function
    MOV R0, #1
    SWI #0x00
function:
    MOV R1, #2
    MOV PC, LR
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Step over should set the debugger to step over mode
	err = svc.StepOver()
	if err != nil {
		t.Errorf("StepOver failed: %v", err)
	}

	// Verify debugger is in step over mode (we can't easily test execution here)
	// The method should at least not error out
}

func TestDebuggerService_StepOut(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    BL function
    MOV R0, #1
    SWI #0x00
function:
    MOV R1, #2
    MOV R2, #3
    MOV PC, LR
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Step into function first
	err = svc.Step()
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}

	// Now step out should set the debugger to step out mode
	err = svc.StepOut()
	if err != nil {
		t.Errorf("StepOut failed: %v", err)
	}
}

func TestDebuggerService_AddWatchpoint(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #0x10000
    MOV R1, #42
    STR R1, [R0]
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Add watchpoint
	err = svc.AddWatchpoint(0x10000, "write")
	if err != nil {
		t.Errorf("AddWatchpoint failed: %v", err)
	}

	// Get watchpoints
	watchpoints := svc.GetWatchpoints()
	if len(watchpoints) == 0 {
		t.Error("Expected watchpoint to be added")
	}
}

func TestDebuggerService_RemoveWatchpoint(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Add watchpoint first
	err = svc.AddWatchpoint(0x10000, "write")
	if err != nil {
		t.Fatalf("AddWatchpoint failed: %v", err)
	}

	// Get watchpoints to find the ID
	watchpoints := svc.GetWatchpoints()
	if len(watchpoints) == 0 {
		t.Fatal("Expected watchpoint to be added")
	}

	// Remove watchpoint by ID
	err = svc.RemoveWatchpoint(watchpoints[0].ID)
	if err != nil {
		t.Errorf("RemoveWatchpoint failed: %v", err)
	}

	// Verify removed
	watchpoints = svc.GetWatchpoints()
	if len(watchpoints) != 0 {
		t.Error("Expected watchpoint to be removed")
	}
}

func TestDebuggerService_AddWatchpoint_InvalidType(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Try to add watchpoint with invalid type
	err = svc.AddWatchpoint(0x10000, "invalid")
	if err == nil {
		t.Error("Expected error for invalid watchpoint type")
	}
}

func TestDebuggerService_StepOver_NoProgramLoaded(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	// Try to step over without loading a program
	err := svc.StepOver()
	if err == nil {
		t.Error("Expected error when stepping over with no program loaded")
	}

	expectedMsg := "no program loaded"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestDebuggerService_StepOut_NoProgramLoaded(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	// Try to step out without loading a program
	err := svc.StepOut()
	if err == nil {
		t.Error("Expected error when stepping out with no program loaded")
	}

	expectedMsg := "no program loaded"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestDebuggerService_RemoveWatchpoint_InvalidID(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Try to remove watchpoint with invalid ID (no watchpoints exist)
	err = svc.RemoveWatchpoint(999)
	if err == nil {
		t.Error("Expected error when removing watchpoint with invalid ID")
	}
}

func TestDebuggerService_ExecuteCommand(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #42
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Execute "info registers" command
	output, err := svc.ExecuteCommand("info registers")
	if err != nil {
		t.Errorf("ExecuteCommand failed: %v", err)
	}

	if output == "" {
		t.Error("Expected non-empty command output")
	}

	// Verify output contains register information
	if !containsAny(output, []string{"R0", "R1", "R2", "register"}) {
		t.Errorf("Expected output to contain register information, got: %s", output)
	}
}

// Helper function to check if string contains any of the provided substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

func TestDebuggerService_EvaluateExpression(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #42
    MOV R1, #10
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Execute first two instructions to set R0=42, R1=10
	svc.Step()
	svc.Step()

	// Evaluate "R0 + R1" and verify the actual result value
	result, err := svc.EvaluateExpression("R0 + R1")
	if err != nil {
		t.Errorf("EvaluateExpression failed: %v", err)
	}

	expected := uint32(52) // 42 + 10
	if result != expected {
		t.Errorf("Expected result %d (42 + 10), got %d", expected, result)
	}
}

func TestDebuggerService_ExecuteCommand_NoProgram(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	// Try to execute command without loading a program
	// Note: ExecuteCommand currently doesn't check for program loaded, so this will succeed
	// but may produce unexpected results. This test documents current behavior.
	output, err := svc.ExecuteCommand("info registers")

	// Command should work even without program loaded (debugger exists)
	if err != nil {
		t.Logf("ExecuteCommand returned error (acceptable): %v", err)
	}

	// Should still produce some output (even if it's just register dump)
	if output == "" && err == nil {
		t.Error("Expected either output or error when executing command")
	}
}

func TestDebuggerService_EvaluateExpression_NoProgram(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	// Try to evaluate expression without loading a program
	// The evaluator might not be initialized without a program loaded
	result, err := svc.EvaluateExpression("R0 + R1")

	// If evaluator is not initialized, we should get an error
	// Otherwise the expression should evaluate (registers start at 0)
	if err != nil {
		// Expected behavior - no evaluator without program
		if result != 0 {
			t.Errorf("Expected result 0 on error, got %d", result)
		}
	} else {
		// Alternative behavior - evaluator exists but registers are 0
		// This is acceptable - documents actual behavior
		t.Logf("EvaluateExpression succeeded without program loaded, result: %d", result)
	}
}

func TestDebuggerService_EvaluateExpression_InvalidExpression(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #42
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Test various invalid expressions
	invalidExpressions := []struct {
		expr        string
		description string
		mustFail    bool // If true, MUST return error. If false, either error or 0 is acceptable
	}{
		{"R0 +", "Incomplete expression", true},
		{"+ R1", "Missing left operand", true},
		{"R0 + + R1", "Double operator", true},
		{"INVALID_REG", "Invalid register name", true},
		{"R0 & R1", "Bitwise AND operator", false}, // May or may not be supported
		{"", "Empty expression", true},
		{"R0 R1", "Missing operator", true},
		{"(R0 + R1", "Unclosed parenthesis", true},
		{"R99", "Invalid register number", true},
	}

	for _, tc := range invalidExpressions {
		result, err := svc.EvaluateExpression(tc.expr)
		if err == nil {
			if tc.mustFail {
				t.Errorf("Expected error for invalid expression '%s' (%s), but got result %d",
					tc.expr, tc.description, result)
			} else {
				// Some expressions might be valid in the implementation
				t.Logf("Expression '%s' (%s) succeeded with result %d (acceptable)",
					tc.expr, tc.description, result)
			}
		}
	}
}

func TestDebuggerService_EvaluateExpression_ComplexExpressions(t *testing.T) {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000)
	svc := service.NewDebuggerService(machine)

	program := `
.org 0x8000
main:
    MOV R0, #10
    MOV R1, #20
    MOV R2, #5
    SWI #0x00
`
	p := parser.NewParser(program, "test.s")
	parsed, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse program: %v", err)
	}

	err = svc.LoadProgram(parsed, 0x8000)
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Execute first three instructions to set R0=10, R1=20, R2=5
	svc.Step()
	svc.Step()
	svc.Step()

	tests := []struct {
		name       string
		expression string
		expected   uint32
	}{
		{
			name:       "simple_addition",
			expression: "R0 + R1",
			expected:   30, // 10 + 20
		},
		{
			name:       "simple_subtraction",
			expression: "R1 - R2",
			expected:   15, // 20 - 5
		},
		{
			name:       "multiple_operations",
			expression: "R0 + R1 + R2",
			expected:   35, // 10 + 20 + 5
		},
		{
			name:       "single_register",
			expression: "R0",
			expected:   10,
		},
		{
			name:       "hex_literal",
			expression: "0x10",
			expected:   16,
		},
		{
			name:       "decimal_literal",
			expression: "42",
			expected:   42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.EvaluateExpression(tt.expression)
			if err != nil {
				t.Errorf("EvaluateExpression('%s') failed: %v", tt.expression, err)
			}

			if result != tt.expected {
				t.Errorf("EvaluateExpression('%s'): expected %d, got %d", tt.expression, tt.expected, result)
			}
		})
	}
}
