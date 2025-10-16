package vm

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestNewSymbolResolver(t *testing.T) {
	t.Run("creates resolver with symbols", func(t *testing.T) {
		symbols := map[string]uint32{
			"main":      0x8000,
			"loop":      0x8010,
			"end":       0x8020,
			"calculate": 0x8100,
		}

		resolver := vm.NewSymbolResolver(symbols)
		if resolver == nil {
			t.Fatal("expected non-nil resolver")
		}

		if !resolver.HasSymbols() {
			t.Error("expected resolver to have symbols")
		}

		if resolver.GetSymbolCount() != 4 {
			t.Errorf("expected 4 symbols, got %d", resolver.GetSymbolCount())
		}
	})

	t.Run("creates empty resolver with nil map", func(t *testing.T) {
		resolver := vm.NewSymbolResolver(nil)
		if resolver == nil {
			t.Fatal("expected non-nil resolver")
		}

		if resolver.HasSymbols() {
			t.Error("expected resolver to have no symbols")
		}

		if resolver.GetSymbolCount() != 0 {
			t.Errorf("expected 0 symbols, got %d", resolver.GetSymbolCount())
		}
	})

	t.Run("creates empty resolver with empty map", func(t *testing.T) {
		resolver := vm.NewSymbolResolver(make(map[string]uint32))
		if resolver == nil {
			t.Fatal("expected non-nil resolver")
		}

		if resolver.HasSymbols() {
			t.Error("expected resolver to have no symbols")
		}
	})
}

func TestLookupAddress(t *testing.T) {
	symbols := map[string]uint32{
		"main": 0x8000,
		"loop": 0x8010,
		"end":  0x8020,
	}
	resolver := vm.NewSymbolResolver(symbols)

	tests := []struct {
		name     string
		address  uint32
		expected string
	}{
		{"exact match main", 0x8000, "main"},
		{"exact match loop", 0x8010, "loop"},
		{"exact match end", 0x8020, "end"},
		{"no match before symbols", 0x7FFF, ""},
		{"no match between symbols", 0x8004, ""},
		{"no match after symbols", 0x9000, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.LookupAddress(tt.address)
			if result != tt.expected {
				t.Errorf("LookupAddress(0x%x) = %q, want %q", tt.address, result, tt.expected)
			}
		})
	}
}

func TestLookupSymbol(t *testing.T) {
	symbols := map[string]uint32{
		"main": 0x8000,
		"loop": 0x8010,
		"end":  0x8020,
	}
	resolver := vm.NewSymbolResolver(symbols)

	tests := []struct {
		name          string
		symbol        string
		expectedAddr  uint32
		expectedFound bool
	}{
		{"found main", "main", 0x8000, true},
		{"found loop", "loop", 0x8010, true},
		{"found end", "end", 0x8020, true},
		{"not found", "notfound", 0, false},
		{"empty string", "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, found := resolver.LookupSymbol(tt.symbol)
			if found != tt.expectedFound {
				t.Errorf("LookupSymbol(%q) found = %v, want %v", tt.symbol, found, tt.expectedFound)
			}
			if found && addr != tt.expectedAddr {
				t.Errorf("LookupSymbol(%q) addr = 0x%x, want 0x%x", tt.symbol, addr, tt.expectedAddr)
			}
		})
	}
}

func TestResolveAddress(t *testing.T) {
	symbols := map[string]uint32{
		"main":      0x8000,
		"loop":      0x8010,
		"end":       0x8020,
		"calculate": 0x8100,
	}
	resolver := vm.NewSymbolResolver(symbols)

	tests := []struct {
		name           string
		address        uint32
		expectedSymbol string
		expectedOffset uint32
		expectedFound  bool
	}{
		// Exact matches
		{"exact main", 0x8000, "main", 0, true},
		{"exact loop", 0x8010, "loop", 0, true},
		{"exact end", 0x8020, "end", 0, true},
		{"exact calculate", 0x8100, "calculate", 0, true},

		// Offsets within symbols
		{"main+4", 0x8004, "main", 4, true},
		{"main+8", 0x8008, "main", 8, true},
		{"main+12", 0x800C, "main", 12, true},
		{"loop+4", 0x8014, "loop", 4, true},
		{"loop+8", 0x8018, "loop", 8, true},
		{"end+4", 0x8024, "end", 4, true},
		{"end+100", 0x8084, "end", 100, true},
		{"calculate+16", 0x8110, "calculate", 16, true},

		// Before first symbol
		{"before all symbols", 0x7FFF, "", 0, false},
		{"way before symbols", 0x1000, "", 0, false},

		// Large offsets (between widely spaced symbols)
		{"between end and calculate", 0x8050, "end", 48, true},
		{"just before calculate", 0x80FF, "end", 223, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbol, offset, found := resolver.ResolveAddress(tt.address)

			if found != tt.expectedFound {
				t.Errorf("ResolveAddress(0x%x) found = %v, want %v", tt.address, found, tt.expectedFound)
			}

			if found {
				if symbol != tt.expectedSymbol {
					t.Errorf("ResolveAddress(0x%x) symbol = %q, want %q", tt.address, symbol, tt.expectedSymbol)
				}
				if offset != tt.expectedOffset {
					t.Errorf("ResolveAddress(0x%x) offset = %d, want %d", tt.address, offset, tt.expectedOffset)
				}
			}
		})
	}
}

func TestResolveAddressEmptyResolver(t *testing.T) {
	resolver := vm.NewSymbolResolver(nil)

	symbol, offset, found := resolver.ResolveAddress(0x8000)
	if found {
		t.Error("expected not found with empty resolver")
	}
	if symbol != "" {
		t.Errorf("expected empty symbol, got %q", symbol)
	}
	if offset != 0 {
		t.Errorf("expected offset 0, got %d", offset)
	}
}

func TestFormatAddress(t *testing.T) {
	symbols := map[string]uint32{
		"main":      0x8000,
		"loop":      0x8010,
		"calculate": 0x8100,
	}
	resolver := vm.NewSymbolResolver(symbols)

	tests := []struct {
		name     string
		address  uint32
		expected string
	}{
		// Exact matches
		{"exact main", 0x8000, "main (0x00008000)"},
		{"exact loop", 0x8010, "loop (0x00008010)"},
		{"exact calculate", 0x8100, "calculate (0x00008100)"},

		// With offsets
		{"main+4", 0x8004, "main+4 (0x00008004)"},
		{"main+12", 0x800C, "main+12 (0x0000800c)"},
		{"loop+8", 0x8018, "loop+8 (0x00008018)"},
		{"calculate+32", 0x8120, "calculate+32 (0x00008120)"},

		// No symbols
		{"before symbols", 0x7FFF, "0x00007fff"},
		{"way after symbols", 0x9000, "calculate+3840 (0x00009000)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.FormatAddress(tt.address)
			if result != tt.expected {
				t.Errorf("FormatAddress(0x%x) = %q, want %q", tt.address, result, tt.expected)
			}
		})
	}
}

func TestFormatAddressCompact(t *testing.T) {
	symbols := map[string]uint32{
		"main":      0x8000,
		"loop":      0x8010,
		"calculate": 0x8100,
	}
	resolver := vm.NewSymbolResolver(symbols)

	tests := []struct {
		name     string
		address  uint32
		expected string
	}{
		// Exact matches
		{"exact main", 0x8000, "main"},
		{"exact loop", 0x8010, "loop"},
		{"exact calculate", 0x8100, "calculate"},

		// With offsets
		{"main+4", 0x8004, "main+4"},
		{"main+12", 0x800C, "main+12"},
		{"loop+8", 0x8018, "loop+8"},
		{"calculate+32", 0x8120, "calculate+32"},

		// No symbols
		{"before symbols", 0x7FFF, "0x00007fff"},
		{"way after symbols", 0x9000, "calculate+3840"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.FormatAddressCompact(tt.address)
			if result != tt.expected {
				t.Errorf("FormatAddressCompact(0x%x) = %q, want %q", tt.address, result, tt.expected)
			}
		})
	}
}

func TestFormatAddressEmptyResolver(t *testing.T) {
	resolver := vm.NewSymbolResolver(nil)

	tests := []struct {
		name     string
		address  uint32
		expected string
	}{
		{"format 0x8000", 0x8000, "0x00008000"},
		{"format 0x1234", 0x1234, "0x00001234"},
		{"format 0xABCD", 0xABCD, "0x0000abcd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.FormatAddress(tt.address)
			if result != tt.expected {
				t.Errorf("FormatAddress(0x%x) = %q, want %q", tt.address, result, tt.expected)
			}

			// Compact should be same with no symbols
			resultCompact := resolver.FormatAddressCompact(tt.address)
			if resultCompact != tt.expected {
				t.Errorf("FormatAddressCompact(0x%x) = %q, want %q", tt.address, resultCompact, tt.expected)
			}
		})
	}
}

func TestGetAllSymbols(t *testing.T) {
	original := map[string]uint32{
		"main": 0x8000,
		"loop": 0x8010,
		"end":  0x8020,
	}
	resolver := vm.NewSymbolResolver(original)

	// Get copy
	copy := resolver.GetAllSymbols()

	// Verify contents match
	if len(copy) != len(original) {
		t.Errorf("GetAllSymbols() returned %d symbols, want %d", len(copy), len(original))
	}

	for name, addr := range original {
		if copyAddr, ok := copy[name]; !ok {
			t.Errorf("symbol %q not found in copy", name)
		} else if copyAddr != addr {
			t.Errorf("symbol %q address = 0x%x, want 0x%x", name, copyAddr, addr)
		}
	}

	// Verify it's a copy (modifying it doesn't affect resolver)
	copy["new"] = 0x9000
	if resolver.GetSymbolCount() != 3 {
		t.Error("modifying copy affected original resolver")
	}
}

func TestSymbolResolverEdgeCases(t *testing.T) {
	t.Run("single symbol", func(t *testing.T) {
		symbols := map[string]uint32{"only": 0x8000}
		resolver := vm.NewSymbolResolver(symbols)

		// Before symbol
		_, _, found := resolver.ResolveAddress(0x7FFF)
		if found {
			t.Error("expected not found before only symbol")
		}

		// At symbol
		symbol, offset, found := resolver.ResolveAddress(0x8000)
		if !found || symbol != "only" || offset != 0 {
			t.Errorf("ResolveAddress(0x8000) = (%q, %d, %v), want (only, 0, true)", symbol, offset, found)
		}

		// After symbol
		symbol, offset, found = resolver.ResolveAddress(0x8004)
		if !found || symbol != "only" || offset != 4 {
			t.Errorf("ResolveAddress(0x8004) = (%q, %d, %v), want (only, 4, true)", symbol, offset, found)
		}
	})

	t.Run("overlapping addresses not possible", func(t *testing.T) {
		// If two symbols have same address, map will only keep one
		// This is expected behavior
		symbols := map[string]uint32{
			"label1": 0x8000,
			"label2": 0x8000,
		}
		resolver := vm.NewSymbolResolver(symbols)

		// Should resolve to one of them (which one is not deterministic)
		symbol, offset, found := resolver.ResolveAddress(0x8000)
		if !found {
			t.Error("expected to find symbol at 0x8000")
		}
		if offset != 0 {
			t.Errorf("expected offset 0, got %d", offset)
		}
		if symbol != "label1" && symbol != "label2" {
			t.Errorf("expected label1 or label2, got %q", symbol)
		}
	})

	t.Run("address at 0x0", func(t *testing.T) {
		symbols := map[string]uint32{
			"start": 0x0,
			"main":  0x8000,
		}
		resolver := vm.NewSymbolResolver(symbols)

		symbol, offset, found := resolver.ResolveAddress(0x0)
		if !found || symbol != "start" || offset != 0 {
			t.Errorf("ResolveAddress(0x0) = (%q, %d, %v), want (start, 0, true)", symbol, offset, found)
		}

		symbol, offset, found = resolver.ResolveAddress(0x4)
		if !found || symbol != "start" || offset != 4 {
			t.Errorf("ResolveAddress(0x4) = (%q, %d, %v), want (start, 4, true)", symbol, offset, found)
		}
	})

	t.Run("maximum uint32 address", func(t *testing.T) {
		symbols := map[string]uint32{
			"maxaddr": 0xFFFFFFF0,
		}
		resolver := vm.NewSymbolResolver(symbols)

		symbol, offset, found := resolver.ResolveAddress(0xFFFFFFFF)
		if !found || symbol != "maxaddr" || offset != 15 {
			t.Errorf("ResolveAddress(0xFFFFFFFF) = (%q, %d, %v), want (maxaddr, 15, true)", symbol, offset, found)
		}
	})
}
