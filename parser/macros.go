package parser

import (
	"fmt"
	"strings"
)

// Macro represents a macro definition
type Macro struct {
	Name       string
	Parameters []string
	Body       []string // Lines of macro body
	Pos        Position
}

// MacroTable manages macro definitions
type MacroTable struct {
	macros map[string]*Macro
}

// NewMacroTable creates a new macro table
func NewMacroTable() *MacroTable {
	return &MacroTable{
		macros: make(map[string]*Macro),
	}
}

// Define defines a new macro
func (mt *MacroTable) Define(macro *Macro) error {
	if _, exists := mt.macros[macro.Name]; exists {
		return fmt.Errorf("macro %q already defined", macro.Name)
	}
	mt.macros[macro.Name] = macro
	return nil
}

// Lookup looks up a macro by name
func (mt *MacroTable) Lookup(name string) (*Macro, bool) {
	macro, exists := mt.macros[name]
	return macro, exists
}

// Expand expands a macro invocation with the given arguments
func (mt *MacroTable) Expand(name string, args []string, pos Position) ([]string, error) {
	macro, exists := mt.macros[name]
	if !exists {
		return nil, fmt.Errorf("undefined macro: %q", name)
	}

	if len(args) != len(macro.Parameters) {
		return nil, fmt.Errorf("macro %q expects %d arguments, got %d at %s",
			name, len(macro.Parameters), len(args), pos)
	}

	// Create substitution map
	substitutions := make(map[string]string)
	for i, param := range macro.Parameters {
		substitutions[param] = args[i]
	}

	// Expand macro body with parameter substitution
	expanded := make([]string, 0, len(macro.Body))
	for _, line := range macro.Body {
		expandedLine := substituteParameters(line, substitutions)
		expanded = append(expanded, expandedLine)
	}

	return expanded, nil
}

// substituteParameters replaces parameter references in a line
func substituteParameters(line string, substitutions map[string]string) string {
	// Simple parameter substitution: replace \param with its value
	// Parameters can be referenced as \param or \{param}
	result := line

	for param, value := range substitutions {
		// Replace \{param}
		result = strings.ReplaceAll(result, "\\{"+param+"}", value)
		// Replace \param (but only whole words)
		result = strings.ReplaceAll(result, "\\"+param, value)
	}

	return result
}

// GetAllMacros returns all defined macros
func (mt *MacroTable) GetAllMacros() map[string]*Macro {
	return mt.macros
}

// Clear clears all macros
func (mt *MacroTable) Clear() {
	mt.macros = make(map[string]*Macro)
}

// MacroExpander handles macro expansion during parsing
type MacroExpander struct {
	macroTable *MacroTable
	// Track expansion depth to prevent infinite recursion
	expansionDepth int
	maxDepth       int
	// Track macro call stack for error reporting
	callStack []string
}

// NewMacroExpander creates a new macro expander
func NewMacroExpander(macroTable *MacroTable) *MacroExpander {
	return &MacroExpander{
		macroTable:     macroTable,
		expansionDepth: 0,
		maxDepth:       MaxMacroNestingDepth,
		callStack:      make([]string, 0),
	}
}

// Expand expands a macro call, checking for recursion
func (me *MacroExpander) Expand(name string, args []string, pos Position) ([]string, error) {
	// Check recursion depth
	if me.expansionDepth >= me.maxDepth {
		return nil, fmt.Errorf("macro expansion too deep (possible recursion) at %s: %s",
			pos, strings.Join(me.callStack, " -> "))
	}

	// Check for direct recursion
	for _, caller := range me.callStack {
		if caller == name {
			return nil, fmt.Errorf("recursive macro call detected at %s: %s -> %s",
				pos, strings.Join(me.callStack, " -> "), name)
		}
	}

	me.expansionDepth++
	me.callStack = append(me.callStack, name)
	defer func() {
		me.expansionDepth--
		me.callStack = me.callStack[:len(me.callStack)-1]
	}()

	return me.macroTable.Expand(name, args, pos)
}

// Reset resets the expander state
func (me *MacroExpander) Reset() {
	me.expansionDepth = 0
	me.callStack = make([]string, 0)
}
