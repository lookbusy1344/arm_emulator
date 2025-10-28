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
