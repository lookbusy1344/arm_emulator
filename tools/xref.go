package tools

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// ReferenceType indicates how a symbol is used
type ReferenceType int

const (
	RefDefinition ReferenceType = iota // Symbol defined here
	RefBranch                          // Branch target
	RefLoad                            // Load from address
	RefStore                           // Store to address
	RefData                            // Data reference
	RefCall                            // Function call (BL)
)

func (r ReferenceType) String() string {
	switch r {
	case RefDefinition:
		return "definition"
	case RefBranch:
		return "branch"
	case RefLoad:
		return "load"
	case RefStore:
		return "store"
	case RefData:
		return "data"
	case RefCall:
		return "call"
	default:
		return "unknown"
	}
}

// Reference represents a single reference to a symbol
type Reference struct {
	Type   ReferenceType
	Line   int
	Column int
	Source string // Source line text
}

// Symbol represents a symbol and all its references
type Symbol struct {
	Name        string
	Definition  *Reference   // Where it's defined
	References  []*Reference // Where it's used
	Value       uint32       // Symbol value (if constant)
	IsConstant  bool         // True for .equ symbols
	IsFunction  bool         // True if it's a function (has BL references)
	IsDataLabel bool         // True if it's a data label
}

// XRefGenerator generates cross-reference information
type XRefGenerator struct {
	parser  *parser.Parser
	program *parser.Program
	symbols map[string]*Symbol
}

// NewXRefGenerator creates a new cross-reference generator
func NewXRefGenerator() *XRefGenerator {
	return &XRefGenerator{
		symbols: make(map[string]*Symbol),
	}
}

// Generate generates cross-reference information from source code
func (x *XRefGenerator) Generate(input, filename string) (map[string]*Symbol, error) {
	// Parse the code
	x.parser = parser.NewParser(input, filename)
	prog, err := x.parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	if prog == nil {
		return nil, fmt.Errorf("failed to parse program")
	}

	x.program = prog

	// Collect symbol definitions
	x.collectDefinitions()

	// Collect symbol references
	x.collectReferences()

	// Analyze call graph
	x.analyzeCallGraph()

	return x.symbols, nil
}

// collectDefinitions collects all symbol definitions
func (x *XRefGenerator) collectDefinitions() {
	// Collect from instructions with labels
	for _, inst := range x.program.Instructions {
		if inst.Label != "" {
			if _, exists := x.symbols[inst.Label]; !exists {
				x.symbols[inst.Label] = &Symbol{
					Name:       inst.Label,
					Definition: nil,
					References: make([]*Reference, 0),
				}
			}
			x.symbols[inst.Label].Definition = &Reference{
				Type:   RefDefinition,
				Line:   inst.Pos.Line,
				Column: inst.Pos.Column,
				Source: inst.RawLine,
			}
		}
	}

	// Collect from directives with labels
	for _, dir := range x.program.Directives {
		if dir.Label != "" {
			if _, exists := x.symbols[dir.Label]; !exists {
				x.symbols[dir.Label] = &Symbol{
					Name:       dir.Label,
					Definition: nil,
					References: make([]*Reference, 0),
				}
			}
			x.symbols[dir.Label].Definition = &Reference{
				Type:   RefDefinition,
				Line:   dir.Pos.Line,
				Column: dir.Pos.Column,
				Source: dir.RawLine,
			}
			x.symbols[dir.Label].IsDataLabel = true
		}
	}

	// Collect from symbol table (.equ, .set)
	if x.program.SymbolTable != nil {
		for name, sym := range x.program.SymbolTable.GetAllSymbols() {
			if _, exists := x.symbols[name]; !exists {
				x.symbols[name] = &Symbol{
					Name:       name,
					Definition: nil,
					References: make([]*Reference, 0),
				}
			}
			x.symbols[name].IsConstant = true
			x.symbols[name].Value = sym.Value
		}
	}
}

// collectReferences collects all symbol references
func (x *XRefGenerator) collectReferences() {
	for _, inst := range x.program.Instructions {
		mnem := strings.ToUpper(inst.Mnemonic)

		// Branch instructions
		if mnem == "B" || mnem == "BL" || mnem == "BX" {
			if len(inst.Operands) > 0 {
				target := inst.Operands[0]
				// Skip register operands
				if !isRegisterOperand(target) {
					refType := RefBranch
					if mnem == "BL" {
						refType = RefCall
					}
					x.addReference(target, refType, inst.Pos.Line, inst.Pos.Column, inst.RawLine)
				}
			}
		}

		// Load/Store instructions with label references
		if mnem == "LDR" || mnem == "STR" || mnem == "LDRB" || mnem == "STRB" || mnem == "LDRH" || mnem == "STRH" {
			if len(inst.Operands) > 1 {
				operand := inst.Operands[1]
				// Check for =label syntax
				if strings.HasPrefix(operand, "=") {
					label := strings.TrimPrefix(operand, "=")
					if !isNumeric(label) {
						refType := RefLoad
						if strings.HasPrefix(mnem, "ST") {
							refType = RefStore
						}
						x.addReference(label, refType, inst.Pos.Line, inst.Pos.Column, inst.RawLine)
					}
				}
			}
		}

		// Check all operands for symbolic constants
		for _, operand := range inst.Operands {
			// Strip immediate prefix
			operand = strings.TrimPrefix(operand, "#")
			// Check if it's a symbol (not a number, not a register)
			if !isNumeric(operand) && !isRegisterOperand(operand) &&
				!strings.Contains(operand, "[") && !strings.Contains(operand, "]") {
				// Could be a symbolic constant
				if x.isSymbol(operand) {
					x.addReference(operand, RefData, inst.Pos.Line, inst.Pos.Column, inst.RawLine)
				}
			}
		}
	}
}

// addReference adds a reference to a symbol
func (x *XRefGenerator) addReference(name string, refType ReferenceType, line, column int, source string) {
	name = strings.TrimSpace(name)

	// Create symbol if it doesn't exist
	if _, exists := x.symbols[name]; !exists {
		x.symbols[name] = &Symbol{
			Name:       name,
			Definition: nil,
			References: make([]*Reference, 0),
		}
	}

	// Add reference
	ref := &Reference{
		Type:   refType,
		Line:   line,
		Column: column,
		Source: source,
	}
	x.symbols[name].References = append(x.symbols[name].References, ref)
}

// analyzeCallGraph determines which symbols are functions
func (x *XRefGenerator) analyzeCallGraph() {
	for _, symbol := range x.symbols {
		// Check if symbol is called with BL
		for _, ref := range symbol.References {
			if ref.Type == RefCall {
				symbol.IsFunction = true
				break
			}
		}
	}
}

// isSymbol checks if a name is a known symbol
func (x *XRefGenerator) isSymbol(name string) bool {
	_, exists := x.symbols[name]
	return exists
}

// isRegisterOperand checks if operand is a register
func isRegisterOperand(operand string) bool {
	operand = strings.ToUpper(strings.TrimSpace(operand))
	if operand == "SP" || operand == "LR" || operand == "PC" {
		return true
	}
	if strings.HasPrefix(operand, "R") && len(operand) >= 2 {
		return true
	}
	return false
}

// XRefReport generates a formatted cross-reference report
type XRefReport struct {
	symbols []*Symbol
}

// NewXRefReport creates a new cross-reference report
func NewXRefReport(symbols map[string]*Symbol) *XRefReport {
	// Sort symbols by name
	sortedSymbols := make([]*Symbol, 0, len(symbols))
	for _, sym := range symbols {
		sortedSymbols = append(sortedSymbols, sym)
	}
	sort.Slice(sortedSymbols, func(i, j int) bool {
		return sortedSymbols[i].Name < sortedSymbols[j].Name
	})

	return &XRefReport{
		symbols: sortedSymbols,
	}
}

// String generates a text report
func (r *XRefReport) String() string {
	var sb strings.Builder

	sb.WriteString("Symbol Cross-Reference\n")
	sb.WriteString("======================\n\n")

	for _, sym := range r.symbols {
		// Symbol name
		sb.WriteString(fmt.Sprintf("%-30s", sym.Name))

		// Symbol type
		if sym.IsConstant {
			sb.WriteString(fmt.Sprintf(" [constant=0x%08X]", sym.Value))
		} else if sym.IsFunction {
			sb.WriteString(" [function]")
		} else if sym.IsDataLabel {
			sb.WriteString(" [data]")
		} else {
			sb.WriteString(" [label]")
		}
		sb.WriteString("\n")

		// Definition
		if sym.Definition != nil {
			sb.WriteString(fmt.Sprintf("  Defined:     line %d\n", sym.Definition.Line))
		} else {
			sb.WriteString("  Defined:     (undefined)\n")
		}

		// References
		if len(sym.References) == 0 {
			sb.WriteString("  Referenced:  (never)\n")
		} else {
			sb.WriteString(fmt.Sprintf("  Referenced:  %d time(s)\n", len(sym.References)))

			// Group references by type
			refsByType := make(map[ReferenceType][]*Reference)
			for _, ref := range sym.References {
				refsByType[ref.Type] = append(refsByType[ref.Type], ref)
			}

			// Sort reference types for consistent output
			types := []ReferenceType{RefCall, RefBranch, RefLoad, RefStore, RefData}
			for _, refType := range types {
				refs := refsByType[refType]
				if len(refs) > 0 {
					lines := make([]string, len(refs))
					for i, ref := range refs {
						lines[i] = fmt.Sprintf("%d", ref.Line)
					}
					sb.WriteString(fmt.Sprintf("    %-10s: line(s) %s\n", refType.String(), strings.Join(lines, ", ")))
				}
			}
		}

		sb.WriteString("\n")
	}

	// Summary
	totalSymbols := len(r.symbols)
	definedSymbols := 0
	undefinedSymbols := 0
	unusedSymbols := 0
	functionCount := 0

	for _, sym := range r.symbols {
		if sym.Definition != nil {
			definedSymbols++
		} else {
			undefinedSymbols++
		}
		if len(sym.References) == 0 {
			unusedSymbols++
		}
		if sym.IsFunction {
			functionCount++
		}
	}

	sb.WriteString("Summary\n")
	sb.WriteString("=======\n")
	sb.WriteString(fmt.Sprintf("Total symbols:     %d\n", totalSymbols))
	sb.WriteString(fmt.Sprintf("Defined:           %d\n", definedSymbols))
	sb.WriteString(fmt.Sprintf("Undefined:         %d\n", undefinedSymbols))
	sb.WriteString(fmt.Sprintf("Unused:            %d\n", unusedSymbols))
	sb.WriteString(fmt.Sprintf("Functions:         %d\n", functionCount))

	return sb.String()
}

// GenerateXRef is a convenience function to generate a cross-reference report
func GenerateXRef(input, filename string) (string, error) {
	gen := NewXRefGenerator()
	symbols, err := gen.Generate(input, filename)
	if err != nil {
		return "", err
	}

	report := NewXRefReport(symbols)
	return report.String(), nil
}

// GetSymbols returns all symbols found in the source
func (x *XRefGenerator) GetSymbols() map[string]*Symbol {
	return x.symbols
}

// GetSymbol returns a specific symbol by name
func (x *XRefGenerator) GetSymbol(name string) (*Symbol, bool) {
	sym, exists := x.symbols[name]
	return sym, exists
}

// GetFunctions returns all symbols that are functions
func (x *XRefGenerator) GetFunctions() []*Symbol {
	functions := make([]*Symbol, 0)
	for _, sym := range x.symbols {
		if sym.IsFunction {
			functions = append(functions, sym)
		}
	}
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Name < functions[j].Name
	})
	return functions
}

// GetDataLabels returns all symbols that are data labels
func (x *XRefGenerator) GetDataLabels() []*Symbol {
	dataLabels := make([]*Symbol, 0)
	for _, sym := range x.symbols {
		if sym.IsDataLabel {
			dataLabels = append(dataLabels, sym)
		}
	}
	sort.Slice(dataLabels, func(i, j int) bool {
		return dataLabels[i].Name < dataLabels[j].Name
	})
	return dataLabels
}

// GetUndefinedSymbols returns all symbols that are referenced but not defined
func (x *XRefGenerator) GetUndefinedSymbols() []*Symbol {
	undefined := make([]*Symbol, 0)
	for _, sym := range x.symbols {
		if sym.Definition == nil && len(sym.References) > 0 {
			undefined = append(undefined, sym)
		}
	}
	sort.Slice(undefined, func(i, j int) bool {
		return undefined[i].Name < undefined[j].Name
	})
	return undefined
}

// GetUnusedSymbols returns all symbols that are defined but never referenced
func (x *XRefGenerator) GetUnusedSymbols() []*Symbol {
	unused := make([]*Symbol, 0)
	for _, sym := range x.symbols {
		if sym.Definition != nil && len(sym.References) == 0 {
			// Skip special entry point symbols
			if !isSpecialLabel(sym.Name) {
				unused = append(unused, sym)
			}
		}
	}
	sort.Slice(unused, func(i, j int) bool {
		return unused[i].Name < unused[j].Name
	})
	return unused
}
