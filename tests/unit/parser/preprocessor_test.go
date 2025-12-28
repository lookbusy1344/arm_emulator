package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// TestPreprocessor_Define verifies symbol definition
func TestPreprocessor_Define(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// Initially not defined
	if pp.IsDefined("TEST_SYMBOL") {
		t.Error("Symbol should not be defined initially")
	}

	// Define it
	pp.Define("TEST_SYMBOL")

	// Now should be defined
	if !pp.IsDefined("TEST_SYMBOL") {
		t.Error("Symbol should be defined after Define()")
	}
}

// TestPreprocessor_Undefine verifies symbol undefinition
func TestPreprocessor_Undefine(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// Define a symbol
	pp.Define("TEST_SYMBOL")
	if !pp.IsDefined("TEST_SYMBOL") {
		t.Error("Symbol should be defined")
	}

	// Undefine it
	pp.Undefine("TEST_SYMBOL")

	// Should no longer be defined
	if pp.IsDefined("TEST_SYMBOL") {
		t.Error("Symbol should not be defined after Undefine()")
	}
}

// TestPreprocessor_IsDefined verifies symbol checking
func TestPreprocessor_IsDefined(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// Check multiple symbols
	symbols := []string{"SYM1", "SYM2", "SYM3"}

	// None should be defined initially
	for _, sym := range symbols {
		if pp.IsDefined(sym) {
			t.Errorf("Symbol '%s' should not be defined initially", sym)
		}
	}

	// Define some
	pp.Define("SYM1")
	pp.Define("SYM3")

	// Check again
	if !pp.IsDefined("SYM1") {
		t.Error("SYM1 should be defined")
	}
	if pp.IsDefined("SYM2") {
		t.Error("SYM2 should not be defined")
	}
	if !pp.IsDefined("SYM3") {
		t.Error("SYM3 should be defined")
	}
}

// TestPreprocessor_Errors verifies error list access
func TestPreprocessor_Errors(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	errors := pp.Errors()
	if errors == nil {
		t.Error("Errors() should not return nil")
	}

	// Should initially have no errors
	if errors.HasErrors() {
		t.Error("Should have no errors initially")
	}
}

// TestPreprocessor_Reset verifies reset functionality
func TestPreprocessor_Reset(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// Define some symbols
	pp.Define("SYM1")
	pp.Define("SYM2")

	// Reset
	pp.Reset()

	// Errors should be cleared (create new error list)
	errors := pp.Errors()
	if errors.HasErrors() {
		t.Error("Should have no errors after reset")
	}

	// Include stack should be cleared
	stack := pp.GetIncludeStack()
	if len(stack) != 0 {
		t.Errorf("Include stack should be empty after reset, got %d entries", len(stack))
	}
}

// TestPreprocessor_GetIncludeStack verifies include stack access
func TestPreprocessor_GetIncludeStack(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// Initially empty
	stack := pp.GetIncludeStack()
	if len(stack) != 0 {
		t.Errorf("Include stack should be empty initially, got %d entries", len(stack))
	}

	// The stack is internal, so we can't easily add to it without processing files
	// This test just verifies the function exists and returns a slice
}

// TestPreprocessor_ProcessContent_Simple verifies basic content processing
func TestPreprocessor_ProcessContent_Simple(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// Simple content without any directives
	content := "MOV R0, #1\nMOV R1, #2\n"
	result, err := pp.ProcessContent(content, "test.s")

	if err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	if result != content {
		t.Errorf("Expected content to be unchanged, got: %s", result)
	}
}

// TestPreprocessor_ProcessContent_Comments verifies comment handling
func TestPreprocessor_ProcessContent_Comments(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// Content with comments
	content := "; This is a comment\nMOV R0, #1 ; inline comment\n"
	result, err := pp.ProcessContent(content, "test.s")

	if err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	// Comments should be preserved
	if result != content {
		t.Errorf("Expected content to be unchanged, got: %s", result)
	}
}

// TestPreprocessor_MultipleDefines verifies multiple symbol definitions
func TestPreprocessor_MultipleDefines(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	symbols := []string{"DEBUG", "FEATURE_X", "PLATFORM_ARM"}

	// Define all
	for _, sym := range symbols {
		pp.Define(sym)
	}

	// Verify all are defined
	for _, sym := range symbols {
		if !pp.IsDefined(sym) {
			t.Errorf("Symbol '%s' should be defined", sym)
		}
	}
}

// TestPreprocessor_DefineUndefineMultiple verifies define/undefine combinations
func TestPreprocessor_DefineUndefineMultiple(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// Define, undefine, redefine
	pp.Define("TEST")
	if !pp.IsDefined("TEST") {
		t.Error("TEST should be defined")
	}

	pp.Undefine("TEST")
	if pp.IsDefined("TEST") {
		t.Error("TEST should not be defined after undefine")
	}

	pp.Define("TEST")
	if !pp.IsDefined("TEST") {
		t.Error("TEST should be defined after redefine")
	}
}

// TestPreprocessor_UndefineNonexistent verifies undefining non-existent symbol
func TestPreprocessor_UndefineNonexistent(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// Undefine a symbol that was never defined (should not panic)
	pp.Undefine("NONEXISTENT")

	// Verify it's still not defined
	if pp.IsDefined("NONEXISTENT") {
		t.Error("NONEXISTENT should not be defined")
	}
}

// TestPreprocessor_ProcessFile_Nonexistent verifies error on non-existent file
func TestPreprocessor_ProcessFile_Nonexistent(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	_, err := pp.ProcessFile("nonexistent_file_12345.s")
	if err == nil {
		t.Error("Expected error when processing non-existent file")
	}
}

// TestPreprocessor_ProcessFile_Simple verifies basic file processing
func TestPreprocessor_ProcessFile_Simple(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "preproc_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple file
	testFile := filepath.Join(tmpDir, "simple.s")
	content := "MOV R0, #1\nMOV R1, #2\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Process the file
	pp := parser.NewPreprocessor(tmpDir)
	result, err := pp.ProcessFile("simple.s")

	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	if result != content {
		t.Errorf("Expected content to match, got: %s", result)
	}
}

// TestPreprocessor_BaseDir verifies base directory handling
func TestPreprocessor_BaseDir(t *testing.T) {
	// Empty base dir should default to "."
	pp := parser.NewPreprocessor("")
	if pp == nil {
		t.Fatal("NewPreprocessor returned nil")
	}

	// Non-empty base dir should be accepted
	pp = parser.NewPreprocessor("/tmp")
	if pp == nil {
		t.Fatal("NewPreprocessor returned nil")
	}
}

// TestPreprocessor_NestedIfdef_ElseInSkippedBlock verifies that .else inside
// a skipped parent block doesn't enable output (issue 3.3)
func TestPreprocessor_NestedIfdef_ElseInSkippedBlock(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// FOO is NOT defined, so the outer block is skipped
	// The inner .else should NOT enable output since parent is skipped
	content := `.ifdef FOO
; A - should be skipped (FOO not defined)
.ifdef BAR
; B - should be skipped
.else
; C - should STILL be skipped (parent FOO block is skipped!)
.endif
.endif
; D - should be included (outside all conditionals)
`
	result, err := pp.ProcessContent(content, "test.s")
	if err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	// Only line D should be in the output
	expected := "; D - should be included (outside all conditionals)\n"
	if result != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
	}
}

// TestPreprocessor_NestedIfdef_ElseInActiveBlock verifies that .else inside
// an active parent block works correctly
func TestPreprocessor_NestedIfdef_ElseInActiveBlock(t *testing.T) {
	pp := parser.NewPreprocessor(".")

	// Define FOO so outer block is active
	pp.Define("FOO")

	// FOO is defined, so outer block is active
	// BAR is NOT defined, so inner .else should enable output
	content := `.ifdef FOO
; A - should be included (FOO is defined)
.ifdef BAR
; B - should be skipped (BAR not defined)
.else
; C - should be included (BAR not defined, but FOO block is active)
.endif
.endif
; D - should be included (outside all conditionals)
`
	result, err := pp.ProcessContent(content, "test.s")
	if err != nil {
		t.Fatalf("ProcessContent failed: %v", err)
	}

	// Lines A, C, and D should be in the output
	expected := `; A - should be included (FOO is defined)
; C - should be included (BAR not defined, but FOO block is active)
; D - should be included (outside all conditionals)
`
	if result != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
	}
}
