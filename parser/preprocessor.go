package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Preprocessor handles file inclusion and conditional assembly
type Preprocessor struct {
	// Track included files to detect circular includes
	includeStack []string
	// Track defined symbols for conditional assembly
	defines map[string]bool
	// Base directory for resolving relative includes
	baseDir string
	// Error list
	errors *ErrorList
}

// NewPreprocessor creates a new preprocessor
func NewPreprocessor(baseDir string) *Preprocessor {
	if baseDir == "" {
		baseDir = "."
	}
	return &Preprocessor{
		includeStack: make([]string, 0),
		defines:      make(map[string]bool),
		baseDir:      baseDir,
		errors:       &ErrorList{},
	}
}

// Define defines a symbol for conditional assembly
func (p *Preprocessor) Define(symbol string) {
	p.defines[symbol] = true
}

// Undefine removes a symbol definition
func (p *Preprocessor) Undefine(symbol string) {
	delete(p.defines, symbol)
}

// IsDefined checks if a symbol is defined
func (p *Preprocessor) IsDefined(symbol string) bool {
	return p.defines[symbol]
}

// ProcessFile processes a file with includes and conditionals
func (p *Preprocessor) ProcessFile(filename string) (string, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(filepath.Join(p.baseDir, filename))
	if err != nil {
		return "", err
	}

	// Check for circular includes
	for _, included := range p.includeStack {
		if included == absPath {
			return "", fmt.Errorf("circular include detected: %s", absPath)
		}
	}

	// Read file
	content, err := os.ReadFile(absPath) // #nosec G304 -- user-provided include file path
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Push onto include stack
	p.includeStack = append(p.includeStack, absPath)
	defer func() {
		p.includeStack = p.includeStack[:len(p.includeStack)-1]
	}()

	// Process the content
	return p.ProcessContent(string(content), filename)
}

// ProcessContent processes content with includes and conditionals
func (p *Preprocessor) ProcessContent(content, filename string) (string, error) {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	// State for conditional assembly
	conditionalStack := make([]bool, 0) // Stack of condition states
	skip := false                       // Whether we're currently skipping lines

	for lineNum, line := range lines {
		pos := Position{Filename: filename, Line: lineNum + 1, Column: 1}
		trimmed := strings.TrimSpace(line)

		// Handle preprocessor directives
		if strings.HasPrefix(trimmed, ".include") {
			if skip {
				continue
			}

			// Parse .include directive
			includeFile := parseIncludeDirective(trimmed)
			if includeFile == "" {
				p.errors.AddError(NewError(pos, ErrorSyntax, "invalid .include directive"))
				continue
			}

			// Process included file
			includedContent, err := p.ProcessFile(includeFile)
			if err != nil {
				p.errors.AddError(NewError(pos, ErrorFileIO, fmt.Sprintf("failed to include %s: %v", includeFile, err)))
				continue
			}

			result = append(result, includedContent)

		} else if strings.HasPrefix(trimmed, ".ifdef") {
			// .ifdef SYMBOL
			parts := strings.Fields(trimmed)
			if len(parts) < 2 {
				p.errors.AddError(NewError(pos, ErrorSyntax, ".ifdef requires a symbol name"))
				continue
			}
			symbol := parts[1]
			condition := p.IsDefined(symbol) && !skip
			conditionalStack = append(conditionalStack, skip)
			skip = !condition

		} else if strings.HasPrefix(trimmed, ".ifndef") {
			// .ifndef SYMBOL
			parts := strings.Fields(trimmed)
			if len(parts) < 2 {
				p.errors.AddError(NewError(pos, ErrorSyntax, ".ifndef requires a symbol name"))
				continue
			}
			symbol := parts[1]
			condition := !p.IsDefined(symbol) && !skip
			conditionalStack = append(conditionalStack, skip)
			skip = !condition

		} else if strings.HasPrefix(trimmed, ".if") && !strings.HasPrefix(trimmed, ".ifdef") && !strings.HasPrefix(trimmed, ".ifndef") {
			// .if expression (not implemented yet, just skip)
			p.errors.AddError(NewError(pos, ErrorSyntax, ".if directive not yet implemented"))
			conditionalStack = append(conditionalStack, skip)
			skip = true

		} else if strings.HasPrefix(trimmed, ".else") {
			// .else - flip the skip state
			if len(conditionalStack) == 0 {
				p.errors.AddError(NewError(pos, ErrorSyntax, ".else without matching .ifdef/.ifndef/.if"))
				continue
			}
			parentSkip := conditionalStack[len(conditionalStack)-1]
			skip = !skip || parentSkip

		} else if strings.HasPrefix(trimmed, ".endif") {
			// .endif - pop conditional stack
			if len(conditionalStack) == 0 {
				p.errors.AddError(NewError(pos, ErrorSyntax, ".endif without matching .ifdef/.ifndef/.if"))
				continue
			}
			skip = conditionalStack[len(conditionalStack)-1]
			conditionalStack = conditionalStack[:len(conditionalStack)-1]

		} else {
			// Regular line - include if not skipping
			if !skip {
				result = append(result, line)
			}
		}
	}

	// Check for unclosed conditionals
	if len(conditionalStack) > 0 {
		p.errors.AddError(NewError(
			Position{Filename: filename, Line: len(lines), Column: 1},
			ErrorSyntax,
			fmt.Sprintf("unclosed conditional directive (%d unmatched)", len(conditionalStack)),
		))
	}

	return strings.Join(result, "\n"), nil
}

// parseIncludeDirective parses a .include directive and returns the filename
func parseIncludeDirective(line string) string {
	// .include "filename" or .include <filename>
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, ".include") {
		return ""
	}

	line = strings.TrimPrefix(line, ".include")
	line = strings.TrimSpace(line)

	// Remove quotes or angle brackets
	if len(line) >= 2 {
		if (line[0] == '"' && line[len(line)-1] == '"') ||
			(line[0] == '<' && line[len(line)-1] == '>') {
			return line[1 : len(line)-1]
		}
	}

	return ""
}

// Errors returns the error list
func (p *Preprocessor) Errors() *ErrorList {
	return p.errors
}

// Reset resets the preprocessor state
func (p *Preprocessor) Reset() {
	p.includeStack = make([]string, 0)
	p.errors = &ErrorList{}
}

// GetIncludeStack returns the current include stack (for debugging)
func (p *Preprocessor) GetIncludeStack() []string {
	stack := make([]string, len(p.includeStack))
	copy(stack, p.includeStack)
	return stack
}
