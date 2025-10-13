package tools

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// FormatStyle defines formatting options
type FormatStyle int

const (
	FormatDefault  FormatStyle = iota // Standard formatting
	FormatCompact                     // Minimal whitespace
	FormatExpanded                    // Extra whitespace for readability
)

// FormatOptions controls formatter behavior
type FormatOptions struct {
	Style              FormatStyle
	LabelColumn        int  // Column for labels (default: 0)
	InstructionColumn  int  // Column for instructions (default: 8)
	OperandColumn      int  // Column for operands (default: 16)
	CommentColumn      int  // Column for comments (default: 40)
	AlignOperands      bool // Align operands in columns
	AlignComments      bool // Align comments in columns
	IndentSize         int  // Spaces for indentation
	PreserveEmptyLines bool // Keep empty lines
	TabWidth           int  // Tab width (for expanding tabs)
}

// DefaultFormatOptions returns default formatter options
func DefaultFormatOptions() *FormatOptions {
	return &FormatOptions{
		Style:              FormatDefault,
		LabelColumn:        0,
		InstructionColumn:  8,
		OperandColumn:      16,
		CommentColumn:      40,
		AlignOperands:      true,
		AlignComments:      true,
		IndentSize:         8,
		PreserveEmptyLines: true,
		TabWidth:           8,
	}
}

// CompactFormatOptions returns options for compact formatting
func CompactFormatOptions() *FormatOptions {
	opts := DefaultFormatOptions()
	opts.Style = FormatCompact
	opts.InstructionColumn = 0
	opts.OperandColumn = 0
	opts.CommentColumn = 0
	opts.AlignOperands = false
	opts.AlignComments = false
	return opts
}

// ExpandedFormatOptions returns options for expanded formatting
func ExpandedFormatOptions() *FormatOptions {
	opts := DefaultFormatOptions()
	opts.Style = FormatExpanded
	opts.InstructionColumn = 12
	opts.OperandColumn = 24
	opts.CommentColumn = 50
	return opts
}

// Formatter formats assembly source code
type Formatter struct {
	options *FormatOptions
	parser  *parser.Parser
	program *parser.Program
	output  strings.Builder
}

// NewFormatter creates a new formatter
func NewFormatter(options *FormatOptions) *Formatter {
	if options == nil {
		options = DefaultFormatOptions()
	}
	return &Formatter{
		options: options,
	}
}

// Format formats the given assembly source code
func (f *Formatter) Format(input, filename string) (string, error) {
	// Parse the code
	f.parser = parser.NewParser(input, filename)
	prog, err := f.parser.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	if prog == nil {
		return "", fmt.Errorf("failed to parse program")
	}

	f.program = prog
	f.output.Reset()

	// Format the program
	f.formatProgram()

	return f.output.String(), nil
}

// formatProgram formats the entire program
func (f *Formatter) formatProgram() {
	// Collect labels already attached to instructions/directives
	attachedLabels := make(map[string]bool)
	for _, inst := range f.program.Instructions {
		if inst.Label != "" {
			attachedLabels[inst.Label] = true
		}
	}
	for _, dir := range f.program.Directives {
		if dir.Label != "" {
			attachedLabels[dir.Label] = true
		}
	}

	// Collect standalone labels from symbol table with their positions
	type standaloneLabel struct {
		name string
		line int
	}
	var standaloneLabels []standaloneLabel
	if f.program.SymbolTable != nil {
		allSymbols := f.program.SymbolTable.GetAllSymbols()
		for name, sym := range allSymbols {
			if !attachedLabels[name] && sym.Type == parser.SymbolLabel {
				standaloneLabels = append(standaloneLabels, standaloneLabel{
					name: name,
					line: sym.Pos.Line,
				})
			}
		}
	}

	// Sort standalone labels by line number to ensure deterministic order
	sort.Slice(standaloneLabels, func(i, j int) bool {
		return standaloneLabels[i].line < standaloneLabels[j].line
	})

	// Interleave instructions, directives, and standalone labels in source order
	instructions := make([]*parser.Instruction, len(f.program.Instructions))
	copy(instructions, f.program.Instructions)
	directives := make([]*parser.Directive, len(f.program.Directives))
	copy(directives, f.program.Directives)

	instIdx := 0
	dirIdx := 0
	labelIdx := 0

	for instIdx < len(instructions) || dirIdx < len(directives) || labelIdx < len(standaloneLabels) {
		// Determine which comes first by comparing line numbers
		var nextInstLine, nextDirLine, nextLabelLine int = 1<<31 - 1, 1<<31 - 1, 1<<31 - 1

		if instIdx < len(instructions) {
			nextInstLine = instructions[instIdx].Pos.Line
		}
		if dirIdx < len(directives) {
			nextDirLine = directives[dirIdx].Pos.Line
		}
		if labelIdx < len(standaloneLabels) {
			nextLabelLine = standaloneLabels[labelIdx].line
		}

		// Output whichever comes first
		if nextLabelLine <= nextInstLine && nextLabelLine <= nextDirLine {
			// Standalone label comes first
			f.output.WriteString(standaloneLabels[labelIdx].name)
			f.output.WriteString(":\n")
			labelIdx++
		} else if nextInstLine <= nextDirLine {
			// Instruction comes first
			f.formatInstruction(instructions[instIdx])
			instIdx++
		} else {
			// Directive comes first
			f.formatDirective(directives[dirIdx])
			dirIdx++
		}
	}
}

// formatInstruction formats a single instruction
func (f *Formatter) formatInstruction(inst *parser.Instruction) {
	line := strings.Builder{}

	// Format label
	if inst.Label != "" {
		if f.options.Style == FormatCompact {
			line.WriteString(inst.Label)
			line.WriteString(":")
		} else {
			line.WriteString(inst.Label)
			line.WriteString(":")
			// Pad to instruction column
			f.padToColumn(&line, f.options.InstructionColumn)
		}
	} else {
		// No label, indent to instruction column
		if f.options.Style != FormatCompact {
			f.padToColumn(&line, f.options.InstructionColumn)
		}
	}

	// Format mnemonic
	mnemonic := strings.ToUpper(inst.Mnemonic)
	if inst.Condition != "" {
		mnemonic += strings.ToUpper(inst.Condition)
	}
	if inst.SetFlags {
		mnemonic += "S"
	}

	if f.options.Style == FormatCompact {
		if inst.Label != "" {
			line.WriteString(" ")
		}
		line.WriteString(mnemonic)
	} else {
		line.WriteString(mnemonic)
		// Pad to operand column if we have operands
		if len(inst.Operands) > 0 && f.options.AlignOperands {
			f.padToColumn(&line, f.options.OperandColumn)
		} else if len(inst.Operands) > 0 {
			line.WriteString("\t")
		}
	}

	// Format operands
	if len(inst.Operands) > 0 {
		if f.options.Style == FormatCompact && inst.Label == "" && inst.Mnemonic != "" {
			line.WriteString(" ")
		}
		operands := f.formatOperands(inst.Operands)
		line.WriteString(operands)
	}

	// Format comment
	if inst.Comment != "" {
		comment := strings.TrimSpace(inst.Comment)
		if f.options.Style == FormatCompact {
			line.WriteString(" ; ")
			line.WriteString(comment)
		} else if f.options.AlignComments {
			f.padToColumn(&line, f.options.CommentColumn)
			line.WriteString("; ")
			line.WriteString(comment)
		} else {
			line.WriteString("\t; ")
			line.WriteString(comment)
		}
	}

	f.output.WriteString(line.String())
	f.output.WriteString("\n")
}

// formatDirective formats a single directive
func (f *Formatter) formatDirective(dir *parser.Directive) {
	line := strings.Builder{}

	// Format label if present
	if dir.Label != "" {
		line.WriteString(dir.Label)
		line.WriteString(":")
		if f.options.Style != FormatCompact {
			f.padToColumn(&line, f.options.InstructionColumn)
		} else {
			line.WriteString(" ")
		}
	} else {
		if f.options.Style != FormatCompact {
			f.padToColumn(&line, f.options.InstructionColumn)
		}
	}

	// Format directive name
	directiveName := strings.ToLower(dir.Name)
	if !strings.HasPrefix(directiveName, ".") {
		directiveName = "." + directiveName
	}
	line.WriteString(directiveName)

	// Format arguments
	if len(dir.Args) > 0 {
		if f.options.Style == FormatCompact {
			line.WriteString(" ")
		} else {
			line.WriteString("\t")
		}
		line.WriteString(strings.Join(dir.Args, ", "))
	}

	// Format comment
	if dir.Comment != "" {
		comment := strings.TrimSpace(dir.Comment)
		if f.options.Style == FormatCompact {
			line.WriteString(" ; ")
			line.WriteString(comment)
		} else if f.options.AlignComments {
			f.padToColumn(&line, f.options.CommentColumn)
			line.WriteString("; ")
			line.WriteString(comment)
		} else {
			line.WriteString("\t; ")
			line.WriteString(comment)
		}
	}

	f.output.WriteString(line.String())
	f.output.WriteString("\n")
}

// formatOperands formats a list of operands
func (f *Formatter) formatOperands(operands []string) string {
	// Join operands with proper spacing
	result := strings.Builder{}

	for i, op := range operands {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(strings.TrimSpace(op))
	}

	return result.String()
}

// padToColumn pads the string builder to the specified column
func (f *Formatter) padToColumn(sb *strings.Builder, column int) {
	current := sb.Len()
	if current < column {
		spaces := column - current
		sb.WriteString(strings.Repeat(" ", spaces))
	} else if current == column {
		// Already at column
	} else {
		// Already past column, add one space
		sb.WriteString(" ")
	}
}

// FormatString is a convenience function to format a string with default options
func FormatString(input, filename string) (string, error) {
	formatter := NewFormatter(DefaultFormatOptions())
	return formatter.Format(input, filename)
}

// FormatStringWithStyle formats a string with the specified style
func FormatStringWithStyle(input, filename string, style FormatStyle) (string, error) {
	var options *FormatOptions
	switch style {
	case FormatCompact:
		options = CompactFormatOptions()
	case FormatExpanded:
		options = ExpandedFormatOptions()
	default:
		options = DefaultFormatOptions()
	}
	formatter := NewFormatter(options)
	return formatter.Format(input, filename)
}
