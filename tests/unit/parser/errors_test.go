package parser_test

import (
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// TestPositionString verifies Position.String() formatting
func TestPositionString(t *testing.T) {
	pos := parser.Position{
		Filename: "test.s",
		Line:     10,
		Column:   5,
	}

	expected := "test.s:10:5"
	if pos.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, pos.String())
	}
}

// TestNewError verifies error creation
func TestNewError(t *testing.T) {
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}
	err := parser.NewError(pos, parser.ErrorSyntax, "syntax error")

	if err.Pos != pos {
		t.Errorf("Expected pos %v, got %v", pos, err.Pos)
	}
	if err.Kind != parser.ErrorSyntax {
		t.Errorf("Expected ErrorSyntax, got %v", err.Kind)
	}
	if err.Message != "syntax error" {
		t.Errorf("Expected 'syntax error', got '%s'", err.Message)
	}
}

// TestNewErrorWithContext verifies error creation with context
func TestNewErrorWithContext(t *testing.T) {
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}
	context := "MOV R0, #invalid"
	err := parser.NewErrorWithContext(pos, parser.ErrorSyntax, "invalid operand", context)

	if err.Pos != pos {
		t.Errorf("Expected pos %v, got %v", pos, err.Pos)
	}
	if err.Kind != parser.ErrorSyntax {
		t.Errorf("Expected ErrorSyntax, got %v", err.Kind)
	}
	if err.Message != "invalid operand" {
		t.Errorf("Expected 'invalid operand', got '%s'", err.Message)
	}
	if err.Context != context {
		t.Errorf("Expected context '%s', got '%s'", context, err.Context)
	}
}

// TestErrorString verifies Error.Error() formatting
func TestErrorString(t *testing.T) {
	pos := parser.Position{Filename: "test.s", Line: 5, Column: 10}
	err := parser.NewErrorWithContext(pos, parser.ErrorSyntax, "unexpected token", "MOV R0 #1")

	result := err.Error()

	expectedSubstrings := []string{
		"test.s:5:10",
		"error:",
		"unexpected token",
		"MOV R0 #1",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(result, substr) {
			t.Errorf("Expected error string to contain '%s', got: %s", substr, result)
		}
	}
}

// TestErrorStringWithoutContext verifies Error.Error() without context
func TestErrorStringWithoutContext(t *testing.T) {
	pos := parser.Position{Filename: "test.s", Line: 5, Column: 10}
	err := parser.NewError(pos, parser.ErrorSyntax, "unexpected token")

	result := err.Error()

	expectedSubstrings := []string{
		"test.s:5:10",
		"error:",
		"unexpected token",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(result, substr) {
			t.Errorf("Expected error string to contain '%s', got: %s", substr, result)
		}
	}
}

// TestWarningString verifies Warning.String() formatting
func TestWarningString(t *testing.T) {
	pos := parser.Position{Filename: "test.s", Line: 3, Column: 7}
	warn := &parser.Warning{
		Pos:     pos,
		Message: "unused label",
	}

	result := warn.String()

	expectedSubstrings := []string{
		"test.s:3:7",
		"warning:",
		"unused label",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(result, substr) {
			t.Errorf("Expected warning string to contain '%s', got: %s", substr, result)
		}
	}
}

// TestErrorListAddError verifies adding errors to error list
func TestErrorListAddError(t *testing.T) {
	el := &parser.ErrorList{}

	if el.HasErrors() {
		t.Error("Expected empty ErrorList to have no errors")
	}

	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}
	err1 := parser.NewError(pos, parser.ErrorSyntax, "error 1")
	err2 := parser.NewError(pos, parser.ErrorSyntax, "error 2")

	el.AddError(err1)
	if !el.HasErrors() {
		t.Error("Expected ErrorList to have errors after AddError")
	}
	if len(el.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(el.Errors))
	}

	el.AddError(err2)
	if len(el.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(el.Errors))
	}
}

// TestErrorListAddWarning verifies adding warnings to error list
func TestErrorListAddWarning(t *testing.T) {
	el := &parser.ErrorList{}

	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}
	warn1 := &parser.Warning{Pos: pos, Message: "warning 1"}
	warn2 := &parser.Warning{Pos: pos, Message: "warning 2"}

	el.AddWarning(warn1)
	if len(el.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(el.Warnings))
	}

	el.AddWarning(warn2)
	if len(el.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(el.Warnings))
	}
}

// TestErrorListError verifies ErrorList.Error() formatting
func TestErrorListError(t *testing.T) {
	el := &parser.ErrorList{}

	// Empty list should return empty string
	if el.Error() != "" {
		t.Errorf("Expected empty string for empty ErrorList, got '%s'", el.Error())
	}

	// Add errors
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}
	err1 := parser.NewError(pos, parser.ErrorSyntax, "error 1")
	err2 := parser.NewError(pos, parser.ErrorSyntax, "error 2")

	el.AddError(err1)
	el.AddError(err2)

	result := el.Error()

	if !strings.Contains(result, "error 1") {
		t.Errorf("Expected result to contain 'error 1', got: %s", result)
	}
	if !strings.Contains(result, "error 2") {
		t.Errorf("Expected result to contain 'error 2', got: %s", result)
	}
}

// TestErrorListPrintWarnings verifies warning printing
func TestErrorListPrintWarnings(t *testing.T) {
	el := &parser.ErrorList{}

	// Empty list should return empty string
	if el.PrintWarnings() != "" {
		t.Errorf("Expected empty string for empty warnings, got '%s'", el.PrintWarnings())
	}

	// Add warnings
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}
	warn1 := &parser.Warning{Pos: pos, Message: "warning 1"}
	warn2 := &parser.Warning{Pos: pos, Message: "warning 2"}

	el.AddWarning(warn1)
	el.AddWarning(warn2)

	result := el.PrintWarnings()

	if !strings.Contains(result, "warning 1") {
		t.Errorf("Expected result to contain 'warning 1', got: %s", result)
	}
	if !strings.Contains(result, "warning 2") {
		t.Errorf("Expected result to contain 'warning 2', got: %s", result)
	}
	if !strings.Contains(result, "test.s:1:1") {
		t.Errorf("Expected result to contain position, got: %s", result)
	}
}

// TestErrorListHasErrors verifies error checking
func TestErrorListHasErrors(t *testing.T) {
	el := &parser.ErrorList{}

	if el.HasErrors() {
		t.Error("Expected new ErrorList to have no errors")
	}

	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	// Add warning only - should not have errors
	warn := &parser.Warning{Pos: pos, Message: "warning"}
	el.AddWarning(warn)
	if el.HasErrors() {
		t.Error("Expected ErrorList with only warnings to have no errors")
	}

	// Add error - should have errors
	err := parser.NewError(pos, parser.ErrorSyntax, "error")
	el.AddError(err)
	if !el.HasErrors() {
		t.Error("Expected ErrorList with errors to return true from HasErrors")
	}
}

// TestAllErrorKinds verifies all error kinds can be created
func TestAllErrorKinds(t *testing.T) {
	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	kinds := []parser.ErrorKind{
		parser.ErrorSyntax,
		parser.ErrorUndefinedLabel,
		parser.ErrorDuplicateLabel,
		parser.ErrorInvalidDirective,
		parser.ErrorInvalidInstruction,
		parser.ErrorInvalidOperand,
		parser.ErrorCircularInclude,
		parser.ErrorMacroExpansion,
		parser.ErrorFileIO,
	}

	for _, kind := range kinds {
		err := parser.NewError(pos, kind, "test error")
		if err.Kind != kind {
			t.Errorf("Expected error kind %v, got %v", kind, err.Kind)
		}
	}
}

// TestErrorListMultipleErrors verifies handling multiple errors
func TestErrorListMultipleErrors(t *testing.T) {
	el := &parser.ErrorList{}

	// Add multiple errors of different kinds
	pos1 := parser.Position{Filename: "test.s", Line: 1, Column: 1}
	pos2 := parser.Position{Filename: "test.s", Line: 5, Column: 10}
	pos3 := parser.Position{Filename: "test.s", Line: 10, Column: 1}

	el.AddError(parser.NewError(pos1, parser.ErrorSyntax, "syntax error"))
	el.AddError(parser.NewError(pos2, parser.ErrorUndefinedLabel, "undefined label"))
	el.AddError(parser.NewError(pos3, parser.ErrorDuplicateLabel, "duplicate label"))

	if len(el.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(el.Errors))
	}

	result := el.Error()

	// Verify all errors are in the output
	expectedSubstrings := []string{
		"test.s:1:1",
		"syntax error",
		"test.s:5:10",
		"undefined label",
		"test.s:10:1",
		"duplicate label",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(result, substr) {
			t.Errorf("Expected error output to contain '%s', got: %s", substr, result)
		}
	}
}

// TestErrorListMixedErrorsAndWarnings verifies handling both errors and warnings
func TestErrorListMixedErrorsAndWarnings(t *testing.T) {
	el := &parser.ErrorList{}

	pos := parser.Position{Filename: "test.s", Line: 1, Column: 1}

	// Add both errors and warnings
	el.AddError(parser.NewError(pos, parser.ErrorSyntax, "error message"))
	el.AddWarning(&parser.Warning{Pos: pos, Message: "warning message"})

	// Should have errors
	if !el.HasErrors() {
		t.Error("Expected HasErrors to be true")
	}

	// Error output should only contain errors
	errOutput := el.Error()
	if !strings.Contains(errOutput, "error message") {
		t.Error("Expected error output to contain error message")
	}
	if strings.Contains(errOutput, "warning message") {
		t.Error("Expected error output to not contain warning message")
	}

	// Warning output should only contain warnings
	warnOutput := el.PrintWarnings()
	if !strings.Contains(warnOutput, "warning message") {
		t.Error("Expected warning output to contain warning message")
	}
	if strings.Contains(warnOutput, "error message") {
		t.Error("Expected warning output to not contain error message")
	}
}
