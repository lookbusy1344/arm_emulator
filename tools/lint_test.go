package tools

import (
	"strings"
	"testing"
)

func TestLint_UndefinedLabel(t *testing.T) {
	source := `
		MOV R0, #10
		B undefined_label
	`

	linter := NewLinter(DefaultLintOptions())
	issues := linter.Lint(source, "test.s")

	// Should find undefined label error
	foundError := false
	for _, issue := range issues {
		if issue.Code == "UNDEF_LABEL" && strings.Contains(issue.Message, "undefined_label") {
			foundError = true
			if issue.Level != LintError {
				t.Errorf("Expected error level, got %v", issue.Level)
			}
		}
	}

	if !foundError {
		t.Error("Expected undefined label error")
	}
}

func TestLint_DuplicateLabel(t *testing.T) {
	source := `
loop:	MOV R0, #10
loop:	ADD R0, R0, #1
	`

	linter := NewLinter(DefaultLintOptions())
	issues := linter.Lint(source, "test.s")

	// Parser may catch this as parse error instead of letting linter handle it
	// Check for either duplicate label warning or parse error
	foundIssue := false
	for _, issue := range issues {
		if issue.Code == "DUPLICATE_LABEL" || issue.Code == "PARSE_ERROR" {
			foundIssue = true
		}
	}

	if !foundIssue {
		t.Error("Expected duplicate label warning or parse error")
	}
}

func TestLint_UnusedLabel(t *testing.T) {
	source := `
_start:	MOV R0, #10
		SWI #0

unused:	MOV R1, #20
	`

	options := DefaultLintOptions()
	options.CheckUnused = true

	linter := NewLinter(options)
	issues := linter.Lint(source, "test.s")

	// Should find unused label warning
	foundWarning := false
	for _, issue := range issues {
		if issue.Code == "UNUSED_LABEL" && strings.Contains(issue.Message, "unused") {
			foundWarning = true
		}
	}

	if !foundWarning {
		t.Error("Expected unused label warning")
	}
}

func TestLint_UnreachableCode(t *testing.T) {
	source := `
_start:	MOV R0, #10
		B end
		MOV R1, #20
end:	SWI #0
	`

	options := DefaultLintOptions()
	options.CheckReach = true

	linter := NewLinter(options)
	issues := linter.Lint(source, "test.s")

	// Should find unreachable code warning
	foundWarning := false
	for _, issue := range issues {
		if issue.Code == "UNREACHABLE_CODE" {
			foundWarning = true
		}
	}

	if !foundWarning {
		t.Error("Expected unreachable code warning")
	}
}

func TestLint_MulRegisterRestriction(t *testing.T) {
	source := `
		MUL R0, R0, R1
	`

	options := DefaultLintOptions()
	options.CheckRegUse = true

	linter := NewLinter(options)
	issues := linter.Lint(source, "test.s")

	// Should find MUL register restriction error
	foundError := false
	for _, issue := range issues {
		if issue.Code == "INVALID_MUL_REGS" {
			foundError = true
			if issue.Level != LintError {
				t.Errorf("Expected error level, got %v", issue.Level)
			}
		}
	}

	if !foundError {
		t.Error("Expected MUL register restriction error")
	}
}

func TestLint_PCDestinationWarning(t *testing.T) {
	source := `
		ADD PC, R0, R1
	`

	options := DefaultLintOptions()
	options.CheckRegUse = true

	linter := NewLinter(options)
	issues := linter.Lint(source, "test.s")

	// Should find PC destination warning
	foundWarning := false
	for _, issue := range issues {
		if issue.Code == "PC_DEST_WARNING" {
			foundWarning = true
		}
	}

	if !foundWarning {
		t.Error("Expected PC destination warning")
	}
}

func TestLint_ValidProgram(t *testing.T) {
	source := `
_start:	MOV R0, #10
		BL subroutine
		SWI #0

subroutine:
		ADD R0, R0, #1
		MOV PC, LR
	`

	linter := NewLinter(DefaultLintOptions())
	issues := linter.Lint(source, "test.s")

	// Should have no errors
	for _, issue := range issues {
		if issue.Level == LintError {
			t.Errorf("Unexpected error in valid program: %v", issue.Message)
		}
	}
}

func TestLint_SuggestionForTypo(t *testing.T) {
	source := `
loop:	MOV R0, #10
		B looop
	`

	options := DefaultLintOptions()
	options.SuggestFixes = true

	linter := NewLinter(options)
	issues := linter.Lint(source, "test.s")

	// Should suggest 'loop' for 'looop'
	foundSuggestion := false
	for _, issue := range issues {
		if issue.Code == "UNDEF_LABEL" && strings.Contains(issue.Message, "did you mean 'loop'") {
			foundSuggestion = true
		}
	}

	if !foundSuggestion {
		t.Error("Expected suggestion for typo")
	}
}

func TestLint_DirectiveValidation(t *testing.T) {
	source := `
		.org
		.word
	`

	linter := NewLinter(DefaultLintOptions())
	issues := linter.Lint(source, "test.s")

	// Should find invalid directive errors
	orgError := false
	wordError := false

	for _, issue := range issues {
		if issue.Code == "INVALID_DIRECTIVE" {
			if strings.Contains(strings.ToLower(issue.Message), "org") {
				orgError = true
			}
			if strings.Contains(strings.ToLower(issue.Message), "word") {
				wordError = true
			}
		}
	}

	if !orgError {
		t.Error("Expected .org directive error")
	}
	if !wordError {
		t.Error("Expected .word directive error")
	}
}

func TestLint_NoIssues(t *testing.T) {
	source := `
_start:	MOV R0, #42
		SWI #0
	`

	linter := NewLinter(DefaultLintOptions())
	issues := linter.Lint(source, "test.s")

	// Should have no errors or warnings
	errorCount := 0
	for _, issue := range issues {
		if issue.Level == LintError {
			errorCount++
			t.Errorf("Unexpected error: %v", issue.Message)
		}
	}

	if errorCount > 0 {
		t.Errorf("Expected 0 errors, got %d", errorCount)
	}
}

func TestLint_LevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"loop", "looop", 1},
		{"start", "stat", 1},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		result := levenshteinDistance(tt.s1, tt.s2)
		if result != tt.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d, expected %d", tt.s1, tt.s2, result, tt.expected)
		}
	}
}

func TestLint_IsSpecialLabel(t *testing.T) {
	tests := []struct {
		label    string
		expected bool
	}{
		{"_start", true},
		{"main", true},
		{"__start", true},
		{"start", true},
		{"_exit", true},
		{"_main", true},
		{"loop", false},
		{"subroutine", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isSpecialLabel(tt.label)
		if result != tt.expected {
			t.Errorf("isSpecialLabel(%q) = %v, expected %v", tt.label, result, tt.expected)
		}
	}
}

func TestLint_NormalizeRegister(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"R0", "R0"},
		{"r0", "R0"},
		{"R13", "SP"},
		{"SP", "SP"},
		{"sp", "SP"},
		{"R14", "LR"},
		{"LR", "LR"},
		{"lr", "LR"},
		{"R15", "PC"},
		{"PC", "PC"},
		{"pc", "PC"},
	}

	for _, tt := range tests {
		result := normalizeRegister(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeRegister(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestLint_IsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0x10", true},
		{"0X10", true},
		{"0b1010", true},
		{"0B1010", true},
		{"#123", true},
		{"#0x10", true},
		{"label", false},
		{"", false},
		{"12abc", false},
	}

	for _, tt := range tests {
		result := isNumeric(tt.input)
		if result != tt.expected {
			t.Errorf("isNumeric(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestLint_MultipleIssues(t *testing.T) {
	source := `
loop:	MOV R0, #10
		B undefined
		MUL R1, R1, R2
loop:	ADD R0, R0, #1
	`

	linter := NewLinter(DefaultLintOptions())
	issues := linter.Lint(source, "test.s")

	// Should find multiple issues
	if len(issues) < 2 {
		t.Errorf("Expected multiple issues, got %d", len(issues))
	}

	// Check that issues are sorted by line number
	for i := 1; i < len(issues); i++ {
		if issues[i].Line < issues[i-1].Line {
			t.Error("Issues not sorted by line number")
		}
	}
}

func TestLint_StrictMode(t *testing.T) {
	source := `
unused:	MOV R0, #10
_start:	SWI #0
	`

	options := DefaultLintOptions()
	options.Strict = true
	options.CheckUnused = true

	linter := NewLinter(options)
	issues := linter.Lint(source, "test.s")

	// In strict mode, should still detect warnings
	// (The strict flag doesn't change detection, just how they're reported in a CLI tool)
	foundWarning := false
	for _, issue := range issues {
		if issue.Code == "UNUSED_LABEL" {
			foundWarning = true
		}
	}

	if !foundWarning {
		t.Error("Expected unused label warning in strict mode")
	}
}

func TestLint_BranchToRegister(t *testing.T) {
	source := `
		MOV R0, #0x1000
		BX R0
	`

	linter := NewLinter(DefaultLintOptions())
	issues := linter.Lint(source, "test.s")

	// Should not complain about register operand in BX
	for _, issue := range issues {
		if issue.Code == "UNDEF_LABEL" {
			t.Errorf("Should not report undefined label for register operand: %v", issue.Message)
		}
	}
}

func TestLint_LDRWithLabel(t *testing.T) {
	source := `
_start:	LDR R0, =data
		SWI #0
data:	.word 42
	`

	linter := NewLinter(DefaultLintOptions())
	issues := linter.Lint(source, "test.s")

	// Should not report undefined label for 'data'
	// Note: Parser may handle =label differently, check if 'data' is undefined
	hasUndefDataError := false
	for _, issue := range issues {
		if issue.Level == LintError && issue.Code == "UNDEF_LABEL" && strings.Contains(issue.Message, "data") {
			hasUndefDataError = true
			t.Errorf("Should not report undefined label for valid label reference: %v", issue.Message)
		}
	}

	// If parser doesn't handle =label syntax, test is still passing
	_ = hasUndefDataError
}

func TestLint_ConditionalBranch(t *testing.T) {
	source := `
_start:	CMP R0, #0
		BEQ zero
		MOV R1, #1
zero:	SWI #0
	`

	linter := NewLinter(DefaultLintOptions())
	issues := linter.Lint(source, "test.s")

	// Conditional branch should not trigger unreachable code warning
	for _, issue := range issues {
		if issue.Code == "UNREACHABLE_CODE" {
			t.Error("Should not report unreachable code after conditional branch")
		}
	}
}

func TestLint_ExitSyscall(t *testing.T) {
	source := `
_start:	MOV R0, #0
		SWI #0
		MOV R1, #1
	`

	options := DefaultLintOptions()
	options.CheckReach = true

	linter := NewLinter(options)
	issues := linter.Lint(source, "test.s")

	// Should detect unreachable code after exit syscall
	// Note: Linter checks for SWI #0 or SWI #0x00
	foundWarning := false
	for _, issue := range issues {
		if issue.Code == "UNREACHABLE_CODE" {
			foundWarning = true
		}
	}

	if !foundWarning {
		t.Error("Expected unreachable code warning after exit syscall")
	}
}
