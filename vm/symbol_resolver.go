package vm

import (
	"fmt"
	"sort"
)

// SymbolResolver provides address-to-symbol lookup functionality for trace output.
// It maintains both forward (name->address) and reverse (address->name) mappings
// and can resolve addresses to the nearest symbol with offset.
type SymbolResolver struct {
	// Forward mapping: symbol name -> address
	symbols map[string]uint32

	// Reverse mapping: address -> symbol name
	addressToSymbol map[uint32]string

	// Sorted list of all symbol addresses for nearest-symbol lookup
	sortedAddresses []uint32
}

// NewSymbolResolver creates a new symbol resolver from a symbol table.
// The symbols map should contain label names mapped to their addresses.
func NewSymbolResolver(symbols map[string]uint32) *SymbolResolver {
	if symbols == nil {
		symbols = make(map[string]uint32)
	}

	// Build reverse mapping
	addressToSymbol := make(map[uint32]string)
	for name, addr := range symbols {
		addressToSymbol[addr] = name
	}

	// Build sorted address list for nearest-symbol lookup
	sortedAddresses := make([]uint32, 0, len(addressToSymbol))
	for addr := range addressToSymbol {
		sortedAddresses = append(sortedAddresses, addr)
	}
	sort.Slice(sortedAddresses, func(i, j int) bool {
		return sortedAddresses[i] < sortedAddresses[j]
	})

	return &SymbolResolver{
		symbols:         symbols,
		addressToSymbol: addressToSymbol,
		sortedAddresses: sortedAddresses,
	}
}

// LookupAddress returns the exact symbol name for an address, or empty string if not found.
func (sr *SymbolResolver) LookupAddress(address uint32) string {
	return sr.addressToSymbol[address]
}

// LookupSymbol returns the address for a symbol name, or 0 and false if not found.
func (sr *SymbolResolver) LookupSymbol(name string) (uint32, bool) {
	addr, ok := sr.symbols[name]
	return addr, ok
}

// ResolveAddress resolves an address to the nearest symbol with offset.
// Returns the symbol name, offset, and whether a symbol was found.
//
// Examples:
//   - Address 0x8000 with symbol "main" at 0x8000 -> ("main", 0, true)
//   - Address 0x8004 with symbol "main" at 0x8000 -> ("main", 4, true)
//   - Address 0x7FFC with no symbols before it -> ("", 0, false)
func (sr *SymbolResolver) ResolveAddress(address uint32) (symbolName string, offset uint32, found bool) {
	// Fast path: exact match
	if name, ok := sr.addressToSymbol[address]; ok {
		return name, 0, true
	}

	// No symbols available
	if len(sr.sortedAddresses) == 0 {
		return "", 0, false
	}

	// Find the nearest symbol at or before this address using binary search
	idx := sort.Search(len(sr.sortedAddresses), func(i int) bool {
		return sr.sortedAddresses[i] > address
	})

	// If idx is 0, address is before all symbols
	if idx == 0 {
		return "", 0, false
	}

	// The symbol at idx-1 is the nearest one at or before our address
	nearestAddr := sr.sortedAddresses[idx-1]
	symbolName = sr.addressToSymbol[nearestAddr]
	offset = address - nearestAddr

	return symbolName, offset, true
}

// FormatAddress formats an address with optional symbol annotation.
// If a symbol is found, returns "symbol+offset (0xADDRESS)", otherwise just "0xADDRESS".
//
// Examples:
//   - Address 0x8000 with symbol "main" at 0x8000 -> "main (0x00008000)"
//   - Address 0x8004 with symbol "main" at 0x8000 -> "main+4 (0x00008004)"
//   - Address 0x8004 with no symbols -> "0x00008004"
func (sr *SymbolResolver) FormatAddress(address uint32) string {
	symbolName, offset, found := sr.ResolveAddress(address)

	if !found {
		return fmt.Sprintf("0x%08x", address)
	}

	if offset == 0 {
		return fmt.Sprintf("%s (0x%08x)", symbolName, address)
	}

	return fmt.Sprintf("%s+%d (0x%08x)", symbolName, offset, address)
}

// FormatAddressCompact formats an address with symbol annotation in compact form.
// If a symbol is found, returns "symbol+offset", otherwise just "0xADDRESS".
//
// Examples:
//   - Address 0x8000 with symbol "main" at 0x8000 -> "main"
//   - Address 0x8004 with symbol "main" at 0x8000 -> "main+4"
//   - Address 0x8004 with no symbols -> "0x00008004"
func (sr *SymbolResolver) FormatAddressCompact(address uint32) string {
	symbolName, offset, found := sr.ResolveAddress(address)

	if !found {
		return fmt.Sprintf("0x%08x", address)
	}

	if offset == 0 {
		return symbolName
	}

	return fmt.Sprintf("%s+%d", symbolName, offset)
}

// HasSymbols returns true if the resolver has any symbols loaded.
func (sr *SymbolResolver) HasSymbols() bool {
	return len(sr.symbols) > 0
}

// GetSymbolCount returns the number of symbols in the resolver.
func (sr *SymbolResolver) GetSymbolCount() int {
	return len(sr.symbols)
}

// GetAllSymbols returns a copy of the symbol map.
func (sr *SymbolResolver) GetAllSymbols() map[string]uint32 {
	result := make(map[string]uint32, len(sr.symbols))
	for name, addr := range sr.symbols {
		result[name] = addr
	}
	return result
}
