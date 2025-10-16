package parser_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

func TestSymbolTable_Define(t *testing.T) {
	st := parser.NewSymbolTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	err := st.Define("test_label", parser.SymbolLabel, 0x8000, pos)
	if err != nil {
		t.Fatalf("failed to define symbol: %v", err)
	}

	sym, exists := st.Lookup("test_label")
	if !exists {
		t.Errorf("symbol 'test_label' not found")
	}
	if sym.Value != 0x8000 {
		t.Errorf("expected value 0x8000, got 0x%X", sym.Value)
	}
}

func TestSymbolTable_DuplicateDefine(t *testing.T) {
	st := parser.NewSymbolTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	st.Define("test_label", parser.SymbolLabel, 0x8000, pos)
	err := st.Define("test_label", parser.SymbolLabel, 0x8004, pos)

	if err == nil {
		t.Errorf("expected error for duplicate symbol definition")
	}
}

func TestSymbolTable_ForwardReference(t *testing.T) {
	st := parser.NewSymbolTable()
	pos1 := parser.Position{Filename: "test.s", Line: 1, Column: 1}
	pos2 := parser.Position{Filename: "test.s", Line: 5, Column: 1}

	// Reference before definition
	st.Reference("forward_label", pos1)

	sym, exists := st.Lookup("forward_label")
	if !exists {
		t.Errorf("forward reference not created")
	}
	if sym.Defined {
		t.Errorf("forward reference should not be defined yet")
	}

	// Define the label
	err := st.Define("forward_label", parser.SymbolLabel, 0x8010, pos2)
	if err != nil {
		t.Fatalf("failed to define forward referenced symbol: %v", err)
	}

	sym, _ = st.Lookup("forward_label")
	if !sym.Defined {
		t.Errorf("symbol should be defined now")
	}
	if sym.Value != 0x8010 {
		t.Errorf("expected value 0x8010, got 0x%X", sym.Value)
	}
}

func TestSymbolTable_GetUndefinedSymbols(t *testing.T) {
	st := parser.NewSymbolTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	st.Reference("undefined1", pos)
	st.Reference("undefined2", pos)
	st.Define("defined", parser.SymbolLabel, 0x8000, pos)

	undefined := st.GetUndefinedSymbols()
	if len(undefined) != 2 {
		t.Errorf("expected 2 undefined symbols, got %d", len(undefined))
	}
}

func TestSymbolTable_ResolveForwardReferences(t *testing.T) {
	st := parser.NewSymbolTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	st.Reference("label1", pos)
	st.Define("label1", parser.SymbolLabel, 0x8000, pos)

	err := st.ResolveForwardReferences()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestSymbolTable_ResolveForwardReferences_Fail(t *testing.T) {
	st := parser.NewSymbolTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	st.Reference("undefined_label", pos)

	err := st.ResolveForwardReferences()
	if err == nil {
		t.Errorf("expected error for undefined symbol")
	}
}

func TestSymbolTable_Constants(t *testing.T) {
	st := parser.NewSymbolTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	err := st.Define("MAX_COUNT", parser.SymbolConstant, 100, pos)
	if err != nil {
		t.Fatalf("failed to define constant: %v", err)
	}

	value, err := st.Get("MAX_COUNT")
	if err != nil {
		t.Fatalf("failed to get constant: %v", err)
	}
	if value != 100 {
		t.Errorf("expected value 100, got %d", value)
	}
}

func TestNumericLabelTable_BackwardReference(t *testing.T) {
	nlt := parser.NewNumericLabelTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	// Define 1: at address 0x8000
	nlt.Define(1, 0x8000, pos)

	// Define 1: again at address 0x8020
	nlt.Define(1, 0x8020, pos)

	// Looking backward from 0x8024 should find 0x8020
	addr, found := nlt.LookupBackward(1, 0x8024)
	if !found {
		t.Errorf("backward reference not found")
	}
	if addr != 0x8020 {
		t.Errorf("expected 0x8020, got 0x%X", addr)
	}

	// Looking backward from 0x8010 should find 0x8000
	addr, found = nlt.LookupBackward(1, 0x8010)
	if !found {
		t.Errorf("backward reference not found")
	}
	if addr != 0x8000 {
		t.Errorf("expected 0x8000, got 0x%X", addr)
	}
}

func TestNumericLabelTable_ForwardReference(t *testing.T) {
	nlt := parser.NewNumericLabelTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	// Define 1: at address 0x8000
	nlt.Define(1, 0x8000, pos)

	// Define 1: again at address 0x8020
	nlt.Define(1, 0x8020, pos)

	// Looking forward from 0x7FF0 should find 0x8000
	addr, found := nlt.LookupForward(1, 0x7FF0)
	if !found {
		t.Errorf("forward reference not found")
	}
	if addr != 0x8000 {
		t.Errorf("expected 0x8000, got 0x%X", addr)
	}

	// Looking forward from 0x8010 should find 0x8020
	addr, found = nlt.LookupForward(1, 0x8010)
	if !found {
		t.Errorf("forward reference not found")
	}
	if addr != 0x8020 {
		t.Errorf("expected 0x8020, got 0x%X", addr)
	}
}

// TestSymbolTable_UpdateAddress verifies updating symbol addresses
func TestSymbolTable_UpdateAddress(t *testing.T) {
	st := parser.NewSymbolTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	// Define a symbol
	err := st.Define("test_symbol", parser.SymbolLabel, 0x8000, pos)
	if err != nil {
		t.Fatalf("Failed to define symbol: %v", err)
	}

	// Update its address
	err = st.UpdateAddress("test_symbol", 0x9000)
	if err != nil {
		t.Errorf("Failed to update symbol address: %v", err)
	}

	// Verify new address
	value, err := st.Get("test_symbol")
	if err != nil {
		t.Errorf("Failed to get symbol: %v", err)
	}
	if value != 0x9000 {
		t.Errorf("Expected address 0x9000, got 0x%X", value)
	}
}

// TestSymbolTable_UpdateAddress_NotFound verifies error on updating non-existent symbol
func TestSymbolTable_UpdateAddress_NotFound(t *testing.T) {
	st := parser.NewSymbolTable()

	err := st.UpdateAddress("nonexistent", 0x8000)
	if err == nil {
		t.Error("Expected error when updating non-existent symbol")
	}
}

// TestSymbolTable_Relocations verifies relocation tracking
func TestSymbolTable_Relocations(t *testing.T) {
	st := parser.NewSymbolTable()
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	// Add a relocation
	rel := &parser.Relocation{
		Pos:        pos,
		SymbolName: "target",
		Address:    0x8000,
		Type:       parser.RelocationBranch,
	}

	st.AddRelocation(rel)

	// Get relocations
	relocs := st.GetRelocations()
	if len(relocs) != 1 {
		t.Errorf("Expected 1 relocation, got %d", len(relocs))
	}

	if relocs[0].SymbolName != "target" {
		t.Errorf("Expected symbol name 'target', got '%s'", relocs[0].SymbolName)
	}

	// Verify the symbol was also referenced
	sym, exists := st.Lookup("target")
	if !exists {
		t.Error("Expected symbol 'target' to be created via relocation")
	}
	if len(sym.References) != 1 {
		t.Errorf("Expected 1 reference, got %d", len(sym.References))
	}
}

// TestSymbolTable_GetRelocations_Empty verifies empty relocation list
func TestSymbolTable_GetRelocations_Empty(t *testing.T) {
	st := parser.NewSymbolTable()

	relocs := st.GetRelocations()
	if len(relocs) != 0 {
		t.Errorf("Expected 0 relocations, got %d", len(relocs))
	}
}
