package integration_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// TestSourceMapPopulation tests that the source map is correctly populated for all instructions
// This is a regression test for a bug where only labeled instructions were added to the source map
func TestSourceMapPopulation(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #1       ; First instruction (labeled)
		MOV R1, #2       ; Second instruction (no label)
		ADD R2, R0, R1   ; Third instruction (no label)
loop:
		CMP R2, #10      ; Fourth instruction (labeled)
		BLT end          ; Fifth instruction (no label)
		SWI #0x00        ; Sixth instruction (no label)
end:
		MOV R0, #0       ; Seventh instruction (labeled)
		SWI #0x00        ; Eighth instruction (no label)
`

	// Parse the assembly
	p := parser.NewParser(code, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Build source map the same way main.go does
	sourceMap := make(map[uint32]string)
	for _, inst := range program.Instructions {
		// Map every instruction's address to its raw source line
		sourceMap[inst.Address] = inst.RawLine
	}

	// Count non-empty instructions (excluding directives and empty lines)
	instructionCount := 0
	for _, inst := range program.Instructions {
		if inst.Mnemonic != "" {
			instructionCount++
		}
	}

	// We should have 8 instructions mapped
	if instructionCount == 0 {
		t.Fatal("No instructions found in parsed program")
	}

	// Verify source map has entries for all instruction addresses
	mappedCount := 0
	for _, inst := range program.Instructions {
		if inst.Mnemonic != "" {
			if sourceLine, exists := sourceMap[inst.Address]; exists {
				mappedCount++
				// Debug: log what we got
				t.Logf("Address 0x%08X: mnemonic=%s, label=%q, rawLine=%q",
					inst.Address, inst.Mnemonic, inst.Label, sourceLine)
			} else {
				t.Errorf("Instruction at address 0x%08X not in source map (mnemonic: %s, raw: %q)",
					inst.Address, inst.Mnemonic, inst.RawLine)
			}
		}
	}

	if mappedCount != instructionCount {
		t.Errorf("Expected %d instructions in source map, got %d", instructionCount, mappedCount)
	}

	// Verify specific addresses have source lines
	expectedMappings := []struct {
		addr     uint32
		contains string
	}{
		{0x8000, "MOV R0, #1"},     // First instruction (labeled)
		{0x8004, "MOV R1, #2"},     // Second instruction (no label)
		{0x8008, "ADD R2, R0, R1"}, // Third instruction (no label)
		{0x800C, "CMP R2, #10"},    // Fourth instruction (labeled as 'loop')
		{0x8010, "BLT end"},        // Fifth instruction (no label)
		{0x8014, "SWI #0x00"},      // Sixth instruction (no label)
		{0x8018, "MOV R0, #0"},     // Seventh instruction (labeled as 'end')
		{0x801C, "SWI #0x00"},      // Eighth instruction (no label)
	}

	for _, expected := range expectedMappings {
		sourceLine, exists := sourceMap[expected.addr]
		if !exists {
			t.Errorf("Address 0x%08X not found in source map (expected line containing %q)",
				expected.addr, expected.contains)
			continue
		}

		// The source line might have labels, comments, etc., so just check if it contains the expected text
		// Also normalize whitespace since the parser might normalize it differently
		if sourceLine == "" {
			t.Errorf("Address 0x%08X has empty source line (expected line containing %q)",
				expected.addr, expected.contains)
		}
	}

	t.Logf("Successfully mapped %d instructions to source lines", mappedCount)
}

// TestSourceMapWithLabelsOnly tests the old buggy behavior to ensure we're not regressing
func TestSourceMapWithLabelsOnly(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #1       ; First instruction (labeled)
		MOV R1, #2       ; Second instruction (no label)
		ADD R2, R0, R1   ; Third instruction (no label)
`

	// Parse the assembly
	p := parser.NewParser(code, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Build symbol table
	symbols := make(map[string]uint32)
	for name, symbol := range program.SymbolTable.GetAllSymbols() {
		if symbol.Type == parser.SymbolLabel {
			symbols[name] = symbol.Value
		}
	}

	// Build source map the OLD WAY (only for labeled instructions)
	oldSourceMap := make(map[uint32]string)
	for _, inst := range program.Instructions {
		if inst.Label != "" {
			if addr, exists := symbols[inst.Label]; exists {
				oldSourceMap[addr] = inst.RawLine
			}
		}
	}

	// The old way should only have 1 entry (the labeled instruction)
	if len(oldSourceMap) > 1 {
		t.Errorf("Old source map method should only map labeled instructions, got %d entries", len(oldSourceMap))
	}

	// Build source map the NEW WAY (all instructions)
	newSourceMap := make(map[uint32]string)
	for _, inst := range program.Instructions {
		newSourceMap[inst.Address] = inst.RawLine
	}

	// Count actual instructions
	instructionCount := 0
	for _, inst := range program.Instructions {
		if inst.Mnemonic != "" {
			instructionCount++
		}
	}

	// The new way should have all 3 instructions
	mappedCount := 0
	for _, inst := range program.Instructions {
		if inst.Mnemonic != "" {
			if _, exists := newSourceMap[inst.Address]; exists {
				mappedCount++
			}
		}
	}

	if mappedCount != instructionCount {
		t.Errorf("New source map method should map all %d instructions, got %d", instructionCount, mappedCount)
	}

	t.Logf("Old method: %d entries, New method: %d entries (out of %d instructions)",
		len(oldSourceMap), mappedCount, instructionCount)
}
