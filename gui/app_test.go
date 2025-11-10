package main

import (
	"testing"
)

func TestApp_LoadProgram(t *testing.T) {
	app := NewApp()

	// Parse simple program
	source := ".org 0x8000\n_start:\nMOV R0, #42\nSWI #0"
	err := app.LoadProgramFromSource(source, "test.s", 0x8000)
	if err != nil {
		t.Fatalf("LoadProgramFromSource failed: %v", err)
	}

	// Get registers
	regs := app.GetRegisters()
	if regs.PC != 0x8000 {
		t.Errorf("expected PC=0x8000, got 0x%08X", regs.PC)
	}
}

func TestApp_StepExecution(t *testing.T) {
	app := NewApp()

	source := ".org 0x8000\n_start:\nMOV R0, #42\nSWI #0"
	if err := app.LoadProgramFromSource(source, "test.s", 0x8000); err != nil {
		t.Fatalf("LoadProgramFromSource failed: %v", err)
	}

	// Step once
	err := app.Step()
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}

	// Check R0 changed
	regs := app.GetRegisters()
	if regs.Registers[0] != 42 {
		t.Errorf("expected R0=42, got %d", regs.Registers[0])
	}
}

// TestStripComments tests the stripComments helper function
func TestStripComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no comments",
			input:    "MOV R0, R1",
			expected: "MOV R0, R1",
		},
		{
			name:     "semicolon line comment",
			input:    "MOV R0, R1 ; this is a comment",
			expected: "MOV R0, R1",
		},
		{
			name:     "at sign line comment",
			input:    "MOV R0, R1 @ this is a comment",
			expected: "MOV R0, R1",
		},
		{
			name:     "double slash line comment",
			input:    "MOV R0, R1 // this is a comment",
			expected: "MOV R0, R1",
		},
		{
			name:     "block comment",
			input:    "MOV R0, /* comment */ R1",
			expected: "MOV R0, R1",
		},
		{
			name:     "multiple block comments",
			input:    "MOV /* first */ R0, /* second */ R1",
			expected: "MOV R0, R1",
		},
		{
			name:     "unclosed block comment",
			input:    "MOV R0, /* comment without end",
			expected: "MOV R0,",
		},
		{
			name:     "line comment at start",
			input:    "; this is a comment",
			expected: "",
		},
		{
			name:     "block comment at start",
			input:    "/* comment */ MOV R0, R1",
			expected: "MOV R0, R1",
		},
		{
			name:     "mixed comments",
			input:    "MOV /* block */ R0, R1 ; line comment",
			expected: "MOV R0, R1",
		},
		{
			name:     "whitespace only",
			input:    "   \t   ",
			expected: "",
		},
		{
			name:     ".org directive with comment",
			input:    ".org 0x8000 ; start address",
			expected: ".org 0x8000",
		},
		{
			name:     ".org directive with block comment",
			input:    ".org /* address */ 0x8000",
			expected: ".org 0x8000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripComments(tt.input)
			if result != tt.expected {
				t.Errorf("stripComments(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestApp_LoadProgramFromSource_InputValidation tests input validation
func TestApp_LoadProgramFromSource_InputValidation(t *testing.T) {
	app := NewApp()

	t.Run("source too large", func(t *testing.T) {
		// Create source larger than 1MB
		largeSource := make([]byte, 1024*1024+1)
		for i := range largeSource {
			largeSource[i] = 'A'
		}
		err := app.LoadProgramFromSource(string(largeSource), "test.s", 0x8000)
		if err == nil {
			t.Fatal("expected error for source too large, got nil")
		}
		if !contains(err.Error(), "source code too large") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("entry point too low", func(t *testing.T) {
		source := "MOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x7FFF)
		if err == nil {
			t.Fatal("expected error for entry point too low, got nil")
		}
		if !contains(err.Error(), "invalid entry point") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("entry point too high", func(t *testing.T) {
		source := "MOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x18000)
		if err == nil {
			t.Fatal("expected error for entry point too high, got nil")
		}
		if !contains(err.Error(), "invalid entry point") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("valid entry point at lower boundary", func(t *testing.T) {
		source := "MOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Errorf("unexpected error for valid entry point: %v", err)
		}
	})

	t.Run("valid entry point at upper boundary", func(t *testing.T) {
		source := "MOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x17FFF)
		if err != nil {
			t.Errorf("unexpected error for valid entry point: %v", err)
		}
	})
}

// TestApp_LoadProgramFromSource_OrgDirectiveDetection tests .org directive detection
func TestApp_LoadProgramFromSource_OrgDirectiveDetection(t *testing.T) {
	app := NewApp()

	t.Run("detects .org directive", func(t *testing.T) {
		source := ".org 0x8000\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}
	})

	t.Run("detects .org with inline semicolon comment", func(t *testing.T) {
		source := ".org 0x8000 ; start address\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}
	})

	t.Run("detects .org with inline @ comment", func(t *testing.T) {
		source := ".org 0x8000 @ start address\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}
	})

	t.Run("detects .org with inline // comment", func(t *testing.T) {
		source := ".org 0x8000 // start address\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}
	})

	t.Run("detects .org with block comment", func(t *testing.T) {
		source := ".org /* address */ 0x8000\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}
	})

	t.Run("does not match .organize", func(t *testing.T) {
		source := "; .organize is not .org\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}
		// Should have added .org directive since .organize doesn't match
	})

	t.Run("handles .org in comment", func(t *testing.T) {
		source := "; .org in comment should be ignored\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}
		// Should have added .org directive since comment .org doesn't count
	})

	t.Run("handles block comment containing .org", func(t *testing.T) {
		source := "/* .org in block comment */ MOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}
		// Should have added .org directive since comment .org doesn't count
	})

	t.Run("auto-inserts .org when missing and loads at correct address", func(t *testing.T) {
		// Source without .org directive
		source := "MOV R0, #42\nMOV R1, #1\nSWI #0"
		entryPoint := uint32(0x8000)

		err := app.LoadProgramFromSource(source, "test.s", entryPoint)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}

		// Verify PC is at entry point
		regs := app.GetRegisters()
		if regs.PC != entryPoint {
			t.Errorf("expected PC=0x%X, got 0x%X", entryPoint, regs.PC)
		}

		// Step once and verify we can execute from the correct location
		err = app.Step()
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}

		// Verify R0 was set to 42 (first instruction executed)
		regs = app.GetRegisters()
		if regs.Registers[0] != 42 {
			t.Errorf("expected R0=42 after first instruction, got %d", regs.Registers[0])
		}
	})

	t.Run("auto-inserts .org with custom entry point", func(t *testing.T) {
		// Source without .org directive
		source := "MOV R0, #99\nSWI #0"
		entryPoint := uint32(0x9000) // Custom entry point

		err := app.LoadProgramFromSource(source, "test.s", entryPoint)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}

		// Verify PC is at custom entry point
		regs := app.GetRegisters()
		if regs.PC != entryPoint {
			t.Errorf("expected PC=0x%X, got 0x%X", entryPoint, regs.PC)
		}
	})
}

// TestApp_GetSymbolsForAddresses tests the batch symbol lookup API
func TestApp_GetSymbolsForAddresses(t *testing.T) {
	app := NewApp()

	// Load a program with labels
	source := `.org 0x8000
_start:
	MOV R0, #42
loop:
	ADD R1, R1, #1
	CMP R1, #10
	BNE loop
done:
	SWI #0`

	err := app.LoadProgramFromSource(source, "test.s", 0x8000)
	if err != nil {
		t.Fatalf("LoadProgramFromSource failed: %v", err)
	}

	t.Run("empty address list", func(t *testing.T) {
		result := app.GetSymbolsForAddresses([]uint32{})
		if len(result) != 0 {
			t.Errorf("expected empty map, got %d entries", len(result))
		}
	})

	t.Run("single address with symbol", func(t *testing.T) {
		result := app.GetSymbolsForAddresses([]uint32{0x8000})
		if len(result) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(result))
		}
		symbol, ok := result[0x8000]
		if !ok {
			t.Fatal("expected symbol for 0x8000")
		}
		if symbol != "_start" {
			t.Errorf("expected symbol '_start', got %q", symbol)
		}
	})

	t.Run("multiple addresses with symbols", func(t *testing.T) {
		// Get symbols for multiple addresses
		symbols := app.GetSymbols()
		addrs := make([]uint32, 0, len(symbols))
		for _, addr := range symbols {
			addrs = append(addrs, addr)
		}

		result := app.GetSymbolsForAddresses(addrs)
		if len(result) != len(symbols) {
			t.Errorf("expected %d entries, got %d", len(symbols), len(result))
		}

		// Verify all symbols are present
		for name, addr := range symbols {
			symbol, ok := result[addr]
			if !ok {
				t.Errorf("missing symbol for address 0x%X", addr)
				continue
			}
			if symbol != name {
				t.Errorf("expected symbol %q for 0x%X, got %q", name, addr, symbol)
			}
		}
	})

	t.Run("address without symbol", func(t *testing.T) {
		// Use an address that doesn't have a symbol
		result := app.GetSymbolsForAddresses([]uint32{0x9000})
		// Should return empty map since no symbol exists
		if len(result) != 0 {
			t.Errorf("expected no entries for address without symbol, got %d", len(result))
		}
	})

	t.Run("mixed addresses with and without symbols", func(t *testing.T) {
		result := app.GetSymbolsForAddresses([]uint32{0x8000, 0x9000, 0x8004})
		// Should have entries for addresses with symbols, but not for 0x9000
		if _, ok := result[0x9000]; ok {
			t.Error("unexpected symbol for address 0x9000")
		}
		if _, ok := result[0x8000]; !ok {
			t.Error("missing symbol for address 0x8000")
		}
	})
}

// TestExtractFilename tests extracting base filename from path
func TestExtractFilename(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "unix absolute path",
			path:     "/home/user/programs/stack.s",
			expected: "stack.s",
		},
		{
			name:     "unix relative path",
			path:     "examples/hello.s",
			expected: "hello.s",
		},
		{
			name:     "windows path",
			path:     "C:\\Users\\John\\code\\test.asm",
			expected: "test.asm",
		},
		{
			name:     "just filename",
			path:     "program.s",
			expected: "program.s",
		},
		{
			name:     "empty string",
			path:     "",
			expected: "",
		},
		{
			name:     "path with multiple extensions",
			path:     "/path/to/file.test.s",
			expected: "file.test.s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFilename(tt.path)
			if result != tt.expected {
				t.Errorf("extractFilename(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

// TestApp_GetCurrentFilename tests getting the current loaded filename
func TestApp_GetCurrentFilename(t *testing.T) {
	app := NewApp()

	t.Run("no file loaded initially", func(t *testing.T) {
		filename := app.GetCurrentFilename()
		if filename != "" {
			t.Errorf("expected empty filename, got %q", filename)
		}
	})

	t.Run("filename set after loading program", func(t *testing.T) {
		source := ".org 0x8000\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "/home/user/test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}

		filename := app.GetCurrentFilename()
		if filename != "test.s" {
			t.Errorf("expected filename 'test.s', got %q", filename)
		}
	})

	t.Run("filename updates with new program", func(t *testing.T) {
		// Load first program
		source1 := ".org 0x8000\nMOV R0, #1\nSWI #0"
		err := app.LoadProgramFromSource(source1, "first.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}

		// Load second program
		source2 := ".org 0x8000\nMOV R0, #2\nSWI #0"
		err = app.LoadProgramFromSource(source2, "examples/second.asm", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}

		filename := app.GetCurrentFilename()
		if filename != "second.asm" {
			t.Errorf("expected filename 'second.asm', got %q", filename)
		}
	})

	t.Run("filename cleared after reset", func(t *testing.T) {
		// Load a program
		source := ".org 0x8000\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}

		// Reset
		err = app.Reset()
		if err != nil {
			t.Fatalf("Reset failed: %v", err)
		}

		filename := app.GetCurrentFilename()
		if filename != "" {
			t.Errorf("expected empty filename after reset, got %q", filename)
		}
	})
}

// TestApp_GetWindowTitle tests getting the formatted window title
func TestApp_GetWindowTitle(t *testing.T) {
	app := NewApp()

	t.Run("default title with no file", func(t *testing.T) {
		title := app.GetWindowTitle()
		if title != "ARM Emulator" {
			t.Errorf("expected 'ARM Emulator', got %q", title)
		}
	})

	t.Run("title with loaded file", func(t *testing.T) {
		source := ".org 0x8000\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "/path/to/stack.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}

		title := app.GetWindowTitle()
		expected := "ARM Emulator - stack.s"
		if title != expected {
			t.Errorf("expected %q, got %q", expected, title)
		}
	})

	t.Run("title returns to default after reset", func(t *testing.T) {
		// Load a program
		source := ".org 0x8000\nMOV R0, #42\nSWI #0"
		err := app.LoadProgramFromSource(source, "test.s", 0x8000)
		if err != nil {
			t.Fatalf("LoadProgramFromSource failed: %v", err)
		}

		// Reset
		err = app.Reset()
		if err != nil {
			t.Fatalf("Reset failed: %v", err)
		}

		title := app.GetWindowTitle()
		if title != "ARM Emulator" {
			t.Errorf("expected 'ARM Emulator' after reset, got %q", title)
		}
	})
}

// contains checks if a string contains a substring
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
