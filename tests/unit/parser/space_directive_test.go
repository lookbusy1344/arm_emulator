package parser_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// TestParser_SpaceDirective_LabelAfterSpace tests that labels after .space directives
// get the correct address (after the space allocation, not before).
func TestParser_SpaceDirective_LabelAfterSpace(t *testing.T) {
	input := `.data
buffer:     .space 12
buffer_end:`

	p := parser.NewParser(input, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Check symbol table
	bufferSym, exists := program.SymbolTable.Lookup("buffer")
	if !exists {
		t.Fatalf("label 'buffer' not found in symbol table")
	}

	bufferEndSym, exists := program.SymbolTable.Lookup("buffer_end")
	if !exists {
		t.Fatalf("label 'buffer_end' not found in symbol table")
	}

	// buffer should be at address 0 (or whatever .data sets)
	expectedBufferAddr := bufferSym.Value
	expectedBufferEndAddr := expectedBufferAddr + 12 // 12 bytes after buffer

	t.Logf("DEBUG: buffer=0x%X, buffer_end=0x%X, expected buffer_end=0x%X",
		bufferSym.Value, bufferEndSym.Value, expectedBufferEndAddr)

	if bufferEndSym.Value != expectedBufferEndAddr {
		t.Errorf("BUG: buffer_end has wrong address. buffer=0x%X, buffer_end=0x%X, expected buffer_end=0x%X",
			bufferSym.Value, bufferEndSym.Value, expectedBufferEndAddr)
	}
}

// TestParser_SpaceDirective_LabelOnSameLineAsSpace tests labels that appear
// BEFORE the .space directive on the SAME line (e.g., "buffer: .space 12")
func TestParser_SpaceDirective_LabelOnSameLineAsSpace(t *testing.T) {
	input := `.data
start_data: .space 0
array:      .space 16
end_array:  .space 0`

	p := parser.NewParser(input, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	startData, exists := program.SymbolTable.Lookup("start_data")
	if !exists {
		t.Fatalf("label 'start_data' not found")
	}

	array, exists := program.SymbolTable.Lookup("array")
	if !exists {
		t.Fatalf("label 'array' not found")
	}

	endArray, exists := program.SymbolTable.Lookup("end_array")
	if !exists {
		t.Fatalf("label 'end_array' not found")
	}

	// With .space on same line as label:
	// start_data: at 0, .space 0 -> next address is 0
	// array: at 0, .space 16 -> next address is 16
	// end_array: at 16, .space 0 -> next address is 16

	t.Logf("DEBUG: start_data=0x%X, array=0x%X, end_array=0x%X",
		startData.Value, array.Value, endArray.Value)

	if array.Value != startData.Value {
		t.Errorf("array should be at same address as start_data (both have .space 0): start_data=0x%X, array=0x%X",
			startData.Value, array.Value)
	}

	if endArray.Value != array.Value+16 {
		t.Errorf("end_array should be 16 bytes after array: array=0x%X, end_array=0x%X",
			array.Value, endArray.Value)
	}
}

// TestParser_SpaceDirective_MultipleLabeledSpaces tests multiple .space directives
// with labels to ensure address tracking works correctly.
func TestParser_SpaceDirective_MultipleLabeledSpaces(t *testing.T) {
	input := `.data
array1:     .space 16
array2:     .space 8
array3:     .space 4
end_marker:`

	p := parser.NewParser(input, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Check all labels exist
	array1, exists := program.SymbolTable.Lookup("array1")
	if !exists {
		t.Fatalf("label 'array1' not found")
	}

	array2, exists := program.SymbolTable.Lookup("array2")
	if !exists {
		t.Fatalf("label 'array2' not found")
	}

	array3, exists := program.SymbolTable.Lookup("array3")
	if !exists {
		t.Fatalf("label 'array3' not found")
	}

	endMarker, exists := program.SymbolTable.Lookup("end_marker")
	if !exists {
		t.Fatalf("label 'end_marker' not found")
	}

	// Verify addresses (assuming .data starts at 0)
	baseAddr := array1.Value

	expectedArray2 := baseAddr + 16
	expectedArray3 := baseAddr + 16 + 8
	expectedEnd := baseAddr + 16 + 8 + 4

	if array2.Value != expectedArray2 {
		t.Errorf("array2 address: got 0x%X, expected 0x%X", array2.Value, expectedArray2)
	}

	if array3.Value != expectedArray3 {
		t.Errorf("array3 address: got 0x%X, expected 0x%X", array3.Value, expectedArray3)
	}

	if endMarker.Value != expectedEnd {
		t.Errorf("end_marker address: got 0x%X, expected 0x%X", endMarker.Value, expectedEnd)
	}
}

// TestParser_SpaceDirective_LabelBeforeSpace tests that labels BEFORE .space
// work correctly (this should already work).
func TestParser_SpaceDirective_LabelBeforeSpace(t *testing.T) {
	input := `.data
buffer:
    .space 12`

	p := parser.NewParser(input, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Check symbol table
	bufferSym, exists := program.SymbolTable.Lookup("buffer")
	if !exists {
		t.Fatalf("label 'buffer' not found in symbol table")
	}

	// This should work fine - label on line before directive
	if bufferSym.Value != 0 {
		t.Errorf("buffer should be at address 0, got 0x%X", bufferSym.Value)
	}
}

// TestParser_SpaceDirective_StandaloneLabelFollowedByLabeledDirective tests the BUG:
// When a standalone label (with nothing after it) is followed by another label with a directive,
// the second label is not found.
// This is THE ACTUAL BUG described in TODO.md!
func TestParser_SpaceDirective_StandaloneLabelFollowedByLabeledDirective(t *testing.T) {
	input := `.data
label1:     .space 4
label2:
label3:     .space 4`

	p := parser.NewParser(input, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Check all labels
	label1, exists1 := program.SymbolTable.Lookup("label1")
	if !exists1 {
		t.Fatalf("label 'label1' not found")
	}

	label2, exists2 := program.SymbolTable.Lookup("label2")
	if !exists2 {
		t.Fatalf("label 'label2' not found")
	}

	label3, exists3 := program.SymbolTable.Lookup("label3")
	if !exists3 {
		t.Fatalf("BUG CONFIRMED: label 'label3' not found - this is the bug described in TODO.md!")
	}

	// Verify addresses
	expectedLabel1 := uint32(0)
	expectedLabel2 := uint32(4) // After .space 4
	expectedLabel3 := uint32(4) // Same as label2 (standalone label)

	t.Logf("DEBUG: label1=0x%X, label2=0x%X, label3=0x%X", label1.Value, label2.Value, label3.Value)

	if label1.Value != expectedLabel1 {
		t.Errorf("label1: got 0x%X, expected 0x%X", label1.Value, expectedLabel1)
	}

	if label2.Value != expectedLabel2 {
		t.Errorf("label2: got 0x%X, expected 0x%X", label2.Value, expectedLabel2)
	}

	if label3.Value != expectedLabel3 {
		t.Errorf("label3: got 0x%X, expected 0x%X", label3.Value, expectedLabel3)
	}
}

// TestParser_SpaceDirective_ZeroSpace tests edge case of .space 0
func TestParser_SpaceDirective_ZeroSpace(t *testing.T) {
	input := `.data
marker1:    .space 0
marker2:`

	p := parser.NewParser(input, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	marker1, exists := program.SymbolTable.Lookup("marker1")
	if !exists {
		t.Fatalf("label 'marker1' not found")
	}

	marker2, exists := program.SymbolTable.Lookup("marker2")
	if !exists {
		t.Fatalf("label 'marker2' not found")
	}

	// With .space 0, both markers should have the same address
	if marker1.Value != marker2.Value {
		t.Errorf("With .space 0, marker2 should equal marker1: marker1=0x%X, marker2=0x%X",
			marker1.Value, marker2.Value)
	}
}

// TestParser_SpaceDirective_MixedSizes tests .space with various sizes
func TestParser_SpaceDirective_MixedSizes(t *testing.T) {
	input := `.data
start_data:
small:      .space 1
medium:     .space 10
large:      .space 100
end_data:`

	p := parser.NewParser(input, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	startData, exists := program.SymbolTable.Lookup("start_data")
	if !exists {
		t.Fatalf("label 'start_data' not found")
	}
	small, exists := program.SymbolTable.Lookup("small")
	if !exists {
		t.Fatalf("label 'small' not found")
	}
	medium, exists := program.SymbolTable.Lookup("medium")
	if !exists {
		t.Fatalf("label 'medium' not found")
	}
	large, exists := program.SymbolTable.Lookup("large")
	if !exists {
		t.Fatalf("label 'large' not found")
	}
	endData, exists := program.SymbolTable.Lookup("end_data")
	if !exists {
		t.Fatalf("label 'end_data' not found")
	}

	baseAddr := startData.Value

	// Verify layout:
	// small: base + 0
	// medium: base + 1 (after 1 byte)
	// large: base + 11 (after 1 + 10)
	// end_data: base + 111 (after 1 + 10 + 100)

	expectedSmall := baseAddr
	expectedMedium := baseAddr + 1
	expectedLarge := baseAddr + 11
	expectedEndData := baseAddr + 111

	if small.Value != expectedSmall {
		t.Errorf("small: got 0x%X, expected 0x%X", small.Value, expectedSmall)
	}

	if medium.Value != expectedMedium {
		t.Errorf("medium: got 0x%X, expected 0x%X", medium.Value, expectedMedium)
	}

	if large.Value != expectedLarge {
		t.Errorf("large: got 0x%X, expected 0x%X", large.Value, expectedLarge)
	}

	if endData.Value != expectedEndData {
		t.Errorf("end_data: got 0x%X, expected 0x%X", endData.Value, expectedEndData)
	}
}
