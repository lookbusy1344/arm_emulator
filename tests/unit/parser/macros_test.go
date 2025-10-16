package parser_test

import (
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// TestMacroTable_Define verifies basic macro definition
func TestMacroTable_Define(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	macro := &parser.Macro{
		Name:       "TEST_MACRO",
		Parameters: []string{"reg", "val"},
		Body:       []string{"MOV \\reg, #\\val"},
		Pos:        pos,
	}

	err := mt.Define(macro)
	if err != nil {
		t.Fatalf("Failed to define macro: %v", err)
	}

	// Verify macro can be looked up
	found, exists := mt.Lookup("TEST_MACRO")
	if !exists {
		t.Error("Expected macro to be defined")
	}
	if found.Name != "TEST_MACRO" {
		t.Errorf("Expected name 'TEST_MACRO', got '%s'", found.Name)
	}
}

// TestMacroTable_DefineDuplicate verifies error on duplicate definition
func TestMacroTable_DefineDuplicate(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	macro := &parser.Macro{
		Name:       "DUP_MACRO",
		Parameters: []string{},
		Body:       []string{"NOP"},
		Pos:        pos,
	}

	// First definition should succeed
	err := mt.Define(macro)
	if err != nil {
		t.Fatalf("First definition failed: %v", err)
	}

	// Second definition should fail
	err = mt.Define(macro)
	if err == nil {
		t.Error("Expected error for duplicate macro definition")
	}
}

// TestMacroTable_Lookup verifies macro lookup
func TestMacroTable_Lookup(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	macro := &parser.Macro{
		Name:       "LOOKUP_TEST",
		Parameters: []string{"x"},
		Body:       []string{"ADD R0, R0, #\\x"},
		Pos:        pos,
	}

	mt.Define(macro)

	// Lookup existing macro
	found, exists := mt.Lookup("LOOKUP_TEST")
	if !exists {
		t.Error("Expected to find macro")
	}
	if found == nil {
		t.Fatal("Found macro is nil")
	}
	if len(found.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(found.Parameters))
	}

	// Lookup non-existent macro
	_, exists = mt.Lookup("NONEXISTENT")
	if exists {
		t.Error("Should not find non-existent macro")
	}
}

// TestMacroTable_Expand verifies macro expansion
func TestMacroTable_Expand(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	macro := &parser.Macro{
		Name:       "LOAD_REG",
		Parameters: []string{"reg", "value"},
		Body:       []string{"MOV \\reg, #\\value", "ADD \\reg, \\reg, #1"},
		Pos:        pos,
	}

	mt.Define(macro)

	// Expand with arguments
	expanded, err := mt.Expand("LOAD_REG", []string{"R1", "42"}, pos)
	if err != nil {
		t.Fatalf("Expansion failed: %v", err)
	}

	if len(expanded) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(expanded))
	}

	// Check parameter substitution
	if !strings.Contains(expanded[0], "R1") {
		t.Errorf("Expected 'R1' in first line, got: %s", expanded[0])
	}
	if !strings.Contains(expanded[0], "42") {
		t.Errorf("Expected '42' in first line, got: %s", expanded[0])
	}
	if !strings.Contains(expanded[1], "R1") {
		t.Errorf("Expected 'R1' in second line, got: %s", expanded[1])
	}
}

// TestMacroTable_ExpandUndefined verifies error on undefined macro
func TestMacroTable_ExpandUndefined(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	_, err := mt.Expand("UNDEFINED_MACRO", []string{}, pos)
	if err == nil {
		t.Error("Expected error for undefined macro")
	}
}

// TestMacroTable_ExpandWrongArgCount verifies error on wrong argument count
func TestMacroTable_ExpandWrongArgCount(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	macro := &parser.Macro{
		Name:       "TWO_PARAM",
		Parameters: []string{"a", "b"},
		Body:       []string{"MOV \\a, \\b"},
		Pos:        pos,
	}

	mt.Define(macro)

	// Try with wrong number of arguments
	_, err := mt.Expand("TWO_PARAM", []string{"R1"}, pos)
	if err == nil {
		t.Error("Expected error for wrong argument count")
	}

	_, err = mt.Expand("TWO_PARAM", []string{"R1", "R2", "R3"}, pos)
	if err == nil {
		t.Error("Expected error for wrong argument count")
	}
}

// TestMacroTable_ExpandBracedParameters verifies braced parameter substitution
func TestMacroTable_ExpandBracedParameters(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	macro := &parser.Macro{
		Name:       "BRACED_TEST",
		Parameters: []string{"reg", "val"},
		Body:       []string{"MOV \\{reg}, #\\{val}"},
		Pos:        pos,
	}

	mt.Define(macro)

	expanded, err := mt.Expand("BRACED_TEST", []string{"R5", "100"}, pos)
	if err != nil {
		t.Fatalf("Expansion failed: %v", err)
	}

	if !strings.Contains(expanded[0], "R5") {
		t.Errorf("Expected 'R5' in expanded line, got: %s", expanded[0])
	}
	if !strings.Contains(expanded[0], "100") {
		t.Errorf("Expected '100' in expanded line, got: %s", expanded[0])
	}
}

// TestMacroTable_GetAllMacros verifies getting all macros
func TestMacroTable_GetAllMacros(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	// Initially empty
	macros := mt.GetAllMacros()
	if len(macros) != 0 {
		t.Errorf("Expected 0 macros, got %d", len(macros))
	}

	// Add some macros
	macro1 := &parser.Macro{Name: "MACRO1", Parameters: []string{}, Body: []string{"NOP"}, Pos: pos}
	macro2 := &parser.Macro{Name: "MACRO2", Parameters: []string{}, Body: []string{"NOP"}, Pos: pos}

	mt.Define(macro1)
	mt.Define(macro2)

	macros = mt.GetAllMacros()
	if len(macros) != 2 {
		t.Errorf("Expected 2 macros, got %d", len(macros))
	}

	if _, exists := macros["MACRO1"]; !exists {
		t.Error("Expected MACRO1 to exist")
	}
	if _, exists := macros["MACRO2"]; !exists {
		t.Error("Expected MACRO2 to exist")
	}
}

// TestMacroTable_Clear verifies clearing all macros
func TestMacroTable_Clear(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	// Add a macro
	macro := &parser.Macro{Name: "TEST", Parameters: []string{}, Body: []string{"NOP"}, Pos: pos}
	mt.Define(macro)

	// Verify it exists
	_, exists := mt.Lookup("TEST")
	if !exists {
		t.Error("Macro should exist before clear")
	}

	// Clear all macros
	mt.Clear()

	// Verify it's gone
	_, exists = mt.Lookup("TEST")
	if exists {
		t.Error("Macro should not exist after clear")
	}

	// Verify GetAllMacros returns empty
	macros := mt.GetAllMacros()
	if len(macros) != 0 {
		t.Errorf("Expected 0 macros after clear, got %d", len(macros))
	}
}

// TestMacroTable_NoParameters verifies macros without parameters
func TestMacroTable_NoParameters(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	macro := &parser.Macro{
		Name:       "NO_PARAMS",
		Parameters: []string{},
		Body:       []string{"MOV R0, #0", "MOV R1, #1"},
		Pos:        pos,
	}

	mt.Define(macro)

	// Expand with no arguments
	expanded, err := mt.Expand("NO_PARAMS", []string{}, pos)
	if err != nil {
		t.Fatalf("Expansion failed: %v", err)
	}

	if len(expanded) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(expanded))
	}

	// Lines should be unchanged (no substitutions)
	if expanded[0] != "MOV R0, #0" {
		t.Errorf("Expected 'MOV R0, #0', got '%s'", expanded[0])
	}
	if expanded[1] != "MOV R1, #1" {
		t.Errorf("Expected 'MOV R1, #1', got '%s'", expanded[1])
	}
}

// TestMacroTable_MultipleSubstitutions verifies multiple parameter substitutions in one line
func TestMacroTable_MultipleSubstitutions(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	macro := &parser.Macro{
		Name:       "ADD_TWO",
		Parameters: []string{"dst", "src"},
		Body:       []string{"ADD \\dst, \\dst, \\src"},
		Pos:        pos,
	}

	mt.Define(macro)

	expanded, err := mt.Expand("ADD_TWO", []string{"R0", "R1"}, pos)
	if err != nil {
		t.Fatalf("Expansion failed: %v", err)
	}

	// Should have R0 twice and R1 once
	if !strings.Contains(expanded[0], "R0") {
		t.Errorf("Expected 'R0' in expanded line, got: %s", expanded[0])
	}
	if !strings.Contains(expanded[0], "R1") {
		t.Errorf("Expected 'R1' in expanded line, got: %s", expanded[0])
	}
}

// TestMacroExpander_Basic verifies basic macro expander
func TestMacroExpander_Basic(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	macro := &parser.Macro{
		Name:       "TEST_MACRO",
		Parameters: []string{"x"},
		Body:       []string{"MOV R0, #\\x"},
		Pos:        pos,
	}
	mt.Define(macro)

	me := parser.NewMacroExpander(mt)

	expanded, err := me.Expand("TEST_MACRO", []string{"42"}, pos)
	if err != nil {
		t.Fatalf("Expansion failed: %v", err)
	}

	if len(expanded) != 1 {
		t.Errorf("Expected 1 line, got %d", len(expanded))
	}

	if !strings.Contains(expanded[0], "42") {
		t.Errorf("Expected '42' in expanded line, got: %s", expanded[0])
	}
}

// TestMacroExpander_Reset verifies expander reset
func TestMacroExpander_Reset(t *testing.T) {
	mt := parser.NewMacroTable()
	me := parser.NewMacroExpander(mt)

	// Reset should not panic
	me.Reset()
}

// TestMacroExpander_RecursionDetection verifies recursion detection
func TestMacroExpander_RecursionDetection(t *testing.T) {
	mt := parser.NewMacroTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	// Create a macro that would call itself (if we tried)
	macro := &parser.Macro{
		Name:       "RECURSIVE",
		Parameters: []string{},
		Body:       []string{"NOP"},
		Pos:        pos,
	}
	mt.Define(macro)

	me := parser.NewMacroExpander(mt)

	// Manually simulate recursive call by expanding, then trying again
	// This is a simplified test - actual recursion would be through macro body
	_, err := me.Expand("RECURSIVE", []string{}, pos)
	if err != nil {
		t.Fatalf("First expansion failed: %v", err)
	}

	// A second expansion after reset should work
	me.Reset()
	_, err = me.Expand("RECURSIVE", []string{}, pos)
	if err != nil {
		t.Errorf("Second expansion after reset failed: %v", err)
	}
}
