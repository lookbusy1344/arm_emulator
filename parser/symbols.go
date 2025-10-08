package parser

import (
	"fmt"
)

// SymbolType represents the type of a symbol
type SymbolType int

const (
	SymbolLabel SymbolType = iota
	SymbolConstant
	SymbolVariable
)

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name    string
	Type    SymbolType
	Value   uint32
	Defined bool
	Pos     Position
	// Forward references: list of positions where this symbol is used
	References []Position
}

// SymbolTable manages symbols during assembly
type SymbolTable struct {
	symbols map[string]*Symbol
	// Relocation entries for forward references
	relocations []*Relocation
}

// Relocation represents a location that needs address resolution
type Relocation struct {
	Pos         Position
	SymbolName  string
	Address     uint32 // Address where the relocation needs to be applied
	Type        RelocationType
	Instruction string // The instruction that needs relocation
}

// RelocationType specifies how to apply the relocation
type RelocationType int

const (
	RelocationAbsolute RelocationType = iota // Full 32-bit address
	RelocationBranch                         // 24-bit signed offset for branch
	RelocationRelative                       // PC-relative offset
)

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols:     make(map[string]*Symbol),
		relocations: make([]*Relocation, 0),
	}
}

// Define defines a new symbol or updates an existing one
func (st *SymbolTable) Define(name string, symType SymbolType, value uint32, pos Position) error {
	if sym, exists := st.symbols[name]; exists {
		if sym.Defined {
			return fmt.Errorf("symbol %q already defined at %s", name, sym.Pos)
		}
		// Update forward reference with actual value
		sym.Value = value
		sym.Defined = true
		sym.Pos = pos
		return nil
	}

	st.symbols[name] = &Symbol{
		Name:       name,
		Type:       symType,
		Value:      value,
		Defined:    true,
		Pos:        pos,
		References: make([]Position, 0),
	}
	return nil
}

// Reference marks a symbol as referenced at a position
func (st *SymbolTable) Reference(name string, pos Position) {
	if sym, exists := st.symbols[name]; exists {
		sym.References = append(sym.References, pos)
	} else {
		// Create forward reference
		st.symbols[name] = &Symbol{
			Name:       name,
			Type:       SymbolLabel,
			Value:      0,
			Defined:    false,
			Pos:        pos,
			References: []Position{pos},
		}
	}
}

// Lookup looks up a symbol by name
func (st *SymbolTable) Lookup(name string) (*Symbol, bool) {
	sym, exists := st.symbols[name]
	return sym, exists
}

// Get returns a symbol's value, or error if undefined
func (st *SymbolTable) Get(name string) (uint32, error) {
	sym, exists := st.symbols[name]
	if !exists {
		return 0, fmt.Errorf("undefined symbol: %q", name)
	}
	if !sym.Defined {
		return 0, fmt.Errorf("symbol %q used but not defined", name)
	}
	return sym.Value, nil
}

// AddRelocation adds a relocation entry
func (st *SymbolTable) AddRelocation(rel *Relocation) {
	st.relocations = append(st.relocations, rel)
	st.Reference(rel.SymbolName, rel.Pos)
}

// GetRelocations returns all relocation entries
func (st *SymbolTable) GetRelocations() []*Relocation {
	return st.relocations
}

// GetUndefinedSymbols returns a list of all undefined symbols
func (st *SymbolTable) GetUndefinedSymbols() []*Symbol {
	undefined := make([]*Symbol, 0)
	for _, sym := range st.symbols {
		if !sym.Defined {
			undefined = append(undefined, sym)
		}
	}
	return undefined
}

// GetUnusedSymbols returns a list of all defined but unreferenced symbols
func (st *SymbolTable) GetUnusedSymbols() []*Symbol {
	unused := make([]*Symbol, 0)
	for _, sym := range st.symbols {
		if sym.Defined && len(sym.References) == 0 {
			unused = append(unused, sym)
		}
	}
	return unused
}

// ResolveForwardReferences resolves all forward references
// Returns an error if any symbols remain undefined
func (st *SymbolTable) ResolveForwardReferences() error {
	undefined := st.GetUndefinedSymbols()
	if len(undefined) > 0 {
		// Return error for the first undefined symbol with its references
		sym := undefined[0]
		if len(sym.References) > 0 {
			return fmt.Errorf("undefined symbol %q referenced at %s", sym.Name, sym.References[0])
		}
		return fmt.Errorf("undefined symbol %q", sym.Name)
	}
	return nil
}

// GetAllSymbols returns all symbols in the table
func (st *SymbolTable) GetAllSymbols() map[string]*Symbol {
	return st.symbols
}

// Clear clears the symbol table
func (st *SymbolTable) Clear() {
	st.symbols = make(map[string]*Symbol)
	st.relocations = make([]*Relocation, 0)
}

// NumericLabelTable manages numeric labels (1:, 2:, etc.) with forward/backward references
type NumericLabelTable struct {
	// Map from label number to list of addresses (allows multiple definitions)
	labels map[int][]uint32
	// Map from label number to list of positions where defined
	positions map[int][]Position
}

// NewNumericLabelTable creates a new numeric label table
func NewNumericLabelTable() *NumericLabelTable {
	return &NumericLabelTable{
		labels:    make(map[int][]uint32),
		positions: make(map[int][]Position),
	}
}

// Define defines a numeric label at an address
func (nlt *NumericLabelTable) Define(num int, address uint32, pos Position) {
	nlt.labels[num] = append(nlt.labels[num], address)
	nlt.positions[num] = append(nlt.positions[num], pos)
}

// LookupBackward finds the most recent definition of a numeric label
// (e.g., 1b looks for the most recent 1:)
func (nlt *NumericLabelTable) LookupBackward(num int, currentAddr uint32) (uint32, bool) {
	addresses, exists := nlt.labels[num]
	if !exists || len(addresses) == 0 {
		return 0, false
	}

	// Find the most recent label before or at currentAddr
	for i := len(addresses) - 1; i >= 0; i-- {
		if addresses[i] <= currentAddr {
			return addresses[i], true
		}
	}

	return 0, false
}

// LookupForward finds the next definition of a numeric label
// (e.g., 1f looks for the next 1:)
func (nlt *NumericLabelTable) LookupForward(num int, currentAddr uint32) (uint32, bool) {
	addresses, exists := nlt.labels[num]
	if !exists || len(addresses) == 0 {
		return 0, false
	}

	// Find the first label after currentAddr
	for _, addr := range addresses {
		if addr > currentAddr {
			return addr, true
		}
	}

	return 0, false
}

// Clear clears all numeric labels
func (nlt *NumericLabelTable) Clear() {
	nlt.labels = make(map[int][]uint32)
	nlt.positions = make(map[int][]Position)
}
