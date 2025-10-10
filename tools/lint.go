package tools

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// LintLevel represents the severity of a lint issue
type LintLevel int

const (
	LintError   LintLevel = iota // Syntax errors, undefined references
	LintWarning                  // Best practice violations, potential issues
	LintInfo                     // Suggestions, style recommendations
)

func (l LintLevel) String() string {
	switch l {
	case LintError:
		return "error"
	case LintWarning:
		return "warning"
	case LintInfo:
		return "info"
	default:
		return "unknown"
	}
}

// LintIssue represents a single lint finding
type LintIssue struct {
	Level   LintLevel
	Line    int
	Column  int
	Message string
	Code    string // Issue code like "UNDEF_LABEL", "UNREACHABLE_CODE"
}

func (i *LintIssue) String() string {
	return fmt.Sprintf("line %d:%d: %s: %s [%s]", i.Line, i.Column, i.Level, i.Message, i.Code)
}

// LintOptions controls linter behavior
type LintOptions struct {
	Strict       bool // Treat warnings as errors
	CheckUnused  bool // Check for unused labels
	CheckReach   bool // Check for unreachable code
	CheckRegUse  bool // Check register usage
	SuggestFixes bool // Suggest fixes for common issues
}

// DefaultLintOptions returns default linter options
func DefaultLintOptions() *LintOptions {
	return &LintOptions{
		Strict:       false,
		CheckUnused:  true,
		CheckReach:   true,
		CheckRegUse:  true,
		SuggestFixes: true,
	}
}

// Linter analyzes assembly code for issues
type Linter struct {
	options *LintOptions
	issues  []*LintIssue
	program *parser.Program
	parser  *parser.Parser

	// Analysis state
	definedLabels    map[string]int   // label -> line number
	referencedLabels map[string][]int // label -> line numbers where used
	instructions     []*parser.Instruction
	directives       []*parser.Directive
}

// NewLinter creates a new linter
func NewLinter(options *LintOptions) *Linter {
	if options == nil {
		options = DefaultLintOptions()
	}
	return &Linter{
		options:          options,
		issues:           make([]*LintIssue, 0),
		definedLabels:    make(map[string]int),
		referencedLabels: make(map[string][]int),
	}
}

// Lint analyzes the given assembly source code
func (l *Linter) Lint(input, filename string) []*LintIssue {
	// Parse the code
	l.parser = parser.NewParser(input, filename)
	prog, err := l.parser.Parse()

	// Check for parse errors
	if err != nil {
		l.issues = append(l.issues, &LintIssue{
			Level:   LintError,
			Line:    1,
			Column:  1,
			Message: fmt.Sprintf("Parse error: %v", err),
			Code:    "PARSE_ERROR",
		})
	}

	// Add parser errors to lint issues
	if l.parser.Errors() != nil {
		for _, perr := range l.parser.Errors().Errors {
			l.issues = append(l.issues, &LintIssue{
				Level:   LintError,
				Line:    perr.Pos.Line,
				Column:  perr.Pos.Column,
				Message: perr.Message,
				Code:    "PARSE_ERROR",
			})
		}
	}

	if prog == nil {
		return l.issues
	}

	l.program = prog
	l.instructions = prog.Instructions
	l.directives = prog.Directives

	// Run analysis passes
	l.collectLabels()
	l.checkUndefinedLabels()

	if l.options.CheckUnused {
		l.checkUnusedLabels()
	}

	if l.options.CheckReach {
		l.checkUnreachableCode()
	}

	if l.options.CheckRegUse {
		l.checkRegisterUsage()
	}

	// Check directives
	l.checkDirectives()

	// Sort issues by line number
	sort.Slice(l.issues, func(i, j int) bool {
		if l.issues[i].Line == l.issues[j].Line {
			return l.issues[i].Column < l.issues[j].Column
		}
		return l.issues[i].Line < l.issues[j].Line
	})

	return l.issues
}

// collectLabels builds a map of all defined labels
func (l *Linter) collectLabels() {
	// Collect from instructions
	for _, inst := range l.instructions {
		if inst.Label != "" {
			if _, exists := l.definedLabels[inst.Label]; exists {
				l.issues = append(l.issues, &LintIssue{
					Level:   LintWarning,
					Line:    inst.Pos.Line,
					Column:  inst.Pos.Column,
					Message: fmt.Sprintf("Duplicate label '%s'", inst.Label),
					Code:    "DUPLICATE_LABEL",
				})
			} else {
				l.definedLabels[inst.Label] = inst.Pos.Line
			}
		}
	}

	// Collect from directives
	for _, dir := range l.directives {
		if dir.Label != "" {
			if _, exists := l.definedLabels[dir.Label]; exists {
				l.issues = append(l.issues, &LintIssue{
					Level:   LintWarning,
					Line:    dir.Pos.Line,
					Column:  dir.Pos.Column,
					Message: fmt.Sprintf("Duplicate label '%s'", dir.Label),
					Code:    "DUPLICATE_LABEL",
				})
			} else {
				l.definedLabels[dir.Label] = dir.Pos.Line
			}
		}
	}

	// Collect from symbol table
	if l.program != nil && l.program.SymbolTable != nil {
		for name := range l.program.SymbolTable.GetAllSymbols() {
			if _, exists := l.definedLabels[name]; !exists {
				// Symbol defined via .equ or similar
				l.definedLabels[name] = 0 // No specific line
			}
		}
	}
}

// checkUndefinedLabels checks for references to undefined labels
func (l *Linter) checkUndefinedLabels() {
	for _, inst := range l.instructions {
		// Check branch instructions
		mnem := strings.ToUpper(inst.Mnemonic)
		if mnem == "B" || mnem == "BL" || mnem == "BX" {
			if len(inst.Operands) > 0 {
				target := inst.Operands[0]
				// Skip register operands (BX Rn)
				if !strings.HasPrefix(strings.ToUpper(target), "R") &&
					!strings.HasPrefix(strings.ToUpper(target), "LR") &&
					!strings.HasPrefix(strings.ToUpper(target), "PC") {
					l.checkLabelReference(target, inst.Pos.Line, inst.Pos.Column)
				}
			}
		}

		// Check LDR/STR with label operands (e.g., LDR R0, =label)
		if (mnem == "LDR" || mnem == "STR") && len(inst.Operands) > 1 {
			operand := inst.Operands[1]
			if strings.HasPrefix(operand, "=") {
				label := strings.TrimPrefix(operand, "=")
				if !isNumeric(label) {
					l.checkLabelReference(label, inst.Pos.Line, inst.Pos.Column)
				}
			}
		}
	}
}

// checkLabelReference verifies a label exists and records usage
func (l *Linter) checkLabelReference(label string, line, column int) {
	label = strings.TrimSpace(label)

	// Record reference
	l.referencedLabels[label] = append(l.referencedLabels[label], line)

	// Check if defined
	if _, exists := l.definedLabels[label]; !exists {
		// Try to suggest similar labels
		suggestion := l.findSimilarLabel(label)
		msg := fmt.Sprintf("Undefined label '%s'", label)
		if suggestion != "" && l.options.SuggestFixes {
			msg += fmt.Sprintf(" (did you mean '%s'?)", suggestion)
		}
		l.issues = append(l.issues, &LintIssue{
			Level:   LintError,
			Line:    line,
			Column:  column,
			Message: msg,
			Code:    "UNDEF_LABEL",
		})
	}
}

// checkUnusedLabels warns about defined but unused labels
func (l *Linter) checkUnusedLabels() {
	for label, defLine := range l.definedLabels {
		if defLine == 0 {
			// Symbol from .equ, skip
			continue
		}

		// Skip special labels
		if isSpecialLabel(label) {
			continue
		}

		if _, used := l.referencedLabels[label]; !used {
			l.issues = append(l.issues, &LintIssue{
				Level:   LintWarning,
				Line:    defLine,
				Column:  1,
				Message: fmt.Sprintf("Label '%s' defined but never referenced", label),
				Code:    "UNUSED_LABEL",
			})
		}
	}
}

// checkUnreachableCode detects code after unconditional branches
func (l *Linter) checkUnreachableCode() {
	for i, inst := range l.instructions {
		mnem := strings.ToUpper(inst.Mnemonic)
		cond := strings.ToUpper(inst.Condition)

		// Check for unconditional branch or exit syscall
		isUnconditionalBranch := (mnem == "B" || mnem == "BL") && (cond == "" || cond == "AL")
		isExitSyscall := false
		if mnem == "SWI" && len(inst.Operands) > 0 {
			operand := strings.TrimSpace(inst.Operands[0])
			// Check for various forms of 0: #0, #0x00, 0, 0x00, etc.
			isExitSyscall = (operand == "#0" || operand == "#0x00" || operand == "0" || operand == "0x00" ||
				operand == "# 0" || operand == "# 0x00") // Parser may add spaces
		}

		if isUnconditionalBranch || isExitSyscall {
			// Check if there's code after this instruction
			if i+1 < len(l.instructions) {
				nextInst := l.instructions[i+1]
				// Only warn if next instruction doesn't have a label (not a branch target)
				if nextInst.Label == "" {
					l.issues = append(l.issues, &LintIssue{
						Level:   LintWarning,
						Line:    nextInst.Pos.Line,
						Column:  nextInst.Pos.Column,
						Message: "Unreachable code detected",
						Code:    "UNREACHABLE_CODE",
					})
					break // Only report once per unreachable block
				}
			}
		}
	}
}

// checkRegisterUsage checks for common register usage issues
func (l *Linter) checkRegisterUsage() {
	for _, inst := range l.instructions {
		mnem := strings.ToUpper(inst.Mnemonic)

		// Check MUL restrictions (Rd and Rm must be different)
		if mnem == "MUL" || mnem == "MLA" {
			if len(inst.Operands) >= 2 {
				rd := normalizeRegister(inst.Operands[0])
				rm := normalizeRegister(inst.Operands[1])
				if rd == rm {
					l.issues = append(l.issues, &LintIssue{
						Level:   LintError,
						Line:    inst.Pos.Line,
						Column:  inst.Pos.Column,
						Message: fmt.Sprintf("%s: destination register Rd and source register Rm must be different", mnem),
						Code:    "INVALID_MUL_REGS",
					})
				}
			}
		}

		// Warn about using PC in certain operations
		for idx, operand := range inst.Operands {
			reg := normalizeRegister(operand)
			if reg == "PC" || reg == "R15" {
				// PC is problematic in many instructions
				if mnem != "MOV" && mnem != "LDR" && mnem != "STR" &&
					mnem != "B" && mnem != "BL" && mnem != "BX" {
					if idx == 0 { // Destination
						l.issues = append(l.issues, &LintIssue{
							Level:   LintWarning,
							Line:    inst.Pos.Line,
							Column:  inst.Pos.Column,
							Message: fmt.Sprintf("Using PC as destination in %s may cause unexpected behavior", mnem),
							Code:    "PC_DEST_WARNING",
						})
					}
				}
			}
		}
	}
}

// checkDirectives validates assembler directives
func (l *Linter) checkDirectives() {
	for _, dir := range l.directives {
		name := strings.ToUpper(dir.Name)

		switch name {
		case ".ORG":
			if len(dir.Args) != 1 {
				l.issues = append(l.issues, &LintIssue{
					Level:   LintError,
					Line:    dir.Pos.Line,
					Column:  dir.Pos.Column,
					Message: ".org directive requires exactly one argument",
					Code:    "INVALID_DIRECTIVE",
				})
			}

		case ".WORD", ".HALF", ".BYTE":
			if len(dir.Args) == 0 {
				l.issues = append(l.issues, &LintIssue{
					Level:   LintError,
					Line:    dir.Pos.Line,
					Column:  dir.Pos.Column,
					Message: fmt.Sprintf("%s directive requires at least one argument", dir.Name),
					Code:    "INVALID_DIRECTIVE",
				})
			}

		case ".ALIGN", ".BALIGN":
			if len(dir.Args) != 1 {
				l.issues = append(l.issues, &LintIssue{
					Level:   LintError,
					Line:    dir.Pos.Line,
					Column:  dir.Pos.Column,
					Message: fmt.Sprintf("%s directive requires exactly one argument", dir.Name),
					Code:    "INVALID_DIRECTIVE",
				})
			}

		case ".INCLUDE":
			if len(dir.Args) != 1 {
				l.issues = append(l.issues, &LintIssue{
					Level:   LintError,
					Line:    dir.Pos.Line,
					Column:  dir.Pos.Column,
					Message: ".include directive requires exactly one argument",
					Code:    "INVALID_DIRECTIVE",
				})
			}
		}
	}
}

// Helper functions

// findSimilarLabel finds a label with a similar name (for suggestions)
func (l *Linter) findSimilarLabel(target string) string {
	target = strings.ToLower(target)
	bestMatch := ""
	bestDistance := 999

	for label := range l.definedLabels {
		dist := levenshteinDistance(strings.ToLower(label), target)
		if dist < bestDistance && dist <= 3 { // Max 3 character difference
			bestMatch = label
			bestDistance = dist
		}
	}

	return bestMatch
}

// levenshteinDistance calculates edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// isSpecialLabel checks if a label is a special entry point or system label
func isSpecialLabel(label string) bool {
	special := []string{"_start", "main", "__start", "start", "_exit", "_main"}
	for _, s := range special {
		if strings.EqualFold(label, s) {
			return true
		}
	}
	return false
}

// normalizeRegister normalizes register names for comparison
func normalizeRegister(operand string) string {
	operand = strings.TrimSpace(operand)
	operand = strings.ToUpper(operand)

	// Handle register names
	switch operand {
	case "SP", "R13":
		return "SP"
	case "LR", "R14":
		return "LR"
	case "PC", "R15":
		return "PC"
	}

	// Handle Rn format
	if strings.HasPrefix(operand, "R") && len(operand) >= 2 {
		return operand[:2] // Return R0-R9 or full for R10-R15
	}

	return operand
}

// isNumeric checks if a string represents a number
func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return true
	}
	if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") {
		return true
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
