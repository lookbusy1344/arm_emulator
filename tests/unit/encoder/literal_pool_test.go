package encoder_test

import (
	"fmt"
	"testing"

	"github.com/lookbusy1344/arm-emulator/encoder"
	"github.com/lookbusy1344/arm-emulator/parser"
)

// Tests for literal pool handling (CODE_REVIEW_OPUS.md section 3.3 and 4.5)
// These tests verify literal pool stress scenarios and edge cases

// TestLiteralPool_ManyLiterals tests adding many literals (>16) to a pool
func TestLiteralPool_ManyLiterals(t *testing.T) {
	enc := encoder.NewEncoder(parser.NewSymbolTable())

	const numLiterals = 50
	baseAddr := uint32(0x8000)

	for i := 0; i < numLiterals; i++ {
		// Each literal is a unique value that can't be encoded as immediate
		value := uint32(0x12340000 + i)
		inst := &parser.Instruction{
			Mnemonic: "LDR",
			Operands: []string{"R0", fmt.Sprintf("=0x%08X", value)},
		}

		addr := baseAddr + uint32(i*4)
		result, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			t.Fatalf("Failed to encode literal %d at 0x%X: %v", i, addr, err)
		}

		// Verify it's a valid LDR instruction (bits 27:26 = 01)
		if (result>>26)&0x3 != 0x1 {
			t.Errorf("Literal %d: expected LDR encoding, got 0x%08X", i, result)
		}
	}

	// Verify all literals are in the pool
	if len(enc.LiteralPool) != numLiterals {
		t.Errorf("Expected %d literals in pool, got %d", numLiterals, len(enc.LiteralPool))
	}
}

// TestLiteralPool_Deduplication tests that identical values are deduplicated
func TestLiteralPool_Deduplication(t *testing.T) {
	enc := encoder.NewEncoder(parser.NewSymbolTable())

	const duplicateValue = uint32(0xDEADBEEF)
	const numReferences = 10
	baseAddr := uint32(0x8000)

	for i := 0; i < numReferences; i++ {
		inst := &parser.Instruction{
			Mnemonic: "LDR",
			Operands: []string{"R0", fmt.Sprintf("=0x%08X", duplicateValue)},
		}

		addr := baseAddr + uint32(i*4)
		_, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			t.Fatalf("Failed to encode duplicate literal %d: %v", i, err)
		}
	}

	// Count occurrences of the duplicate value in the pool
	count := 0
	for _, val := range enc.LiteralPool {
		if val == duplicateValue {
			count++
		}
	}

	// Deduplication means only one entry should exist
	if count != 1 {
		t.Errorf("Expected 1 entry for deduplicated value, got %d", count)
	}
}

// TestLiteralPool_MixedUniqueAndDuplicate tests a mix of unique and duplicate values
func TestLiteralPool_MixedUniqueAndDuplicate(t *testing.T) {
	enc := encoder.NewEncoder(parser.NewSymbolTable())

	values := []uint32{
		0x12345678, // unique
		0xABCDEF00, // unique
		0x12345678, // duplicate of first
		0xFEDCBA98, // unique
		0xABCDEF00, // duplicate of second
		0x12345678, // duplicate of first again
		0x11111111, // unique
	}

	baseAddr := uint32(0x8000)
	for i, val := range values {
		inst := &parser.Instruction{
			Mnemonic: "LDR",
			Operands: []string{"R0", fmt.Sprintf("=0x%08X", val)},
		}

		addr := baseAddr + uint32(i*4)
		_, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			t.Fatalf("Failed to encode literal %d (0x%08X): %v", i, val, err)
		}
	}

	// Should have 4 unique values in the pool
	expectedUnique := 4
	if len(enc.LiteralPool) != expectedUnique {
		t.Errorf("Expected %d unique literals in pool, got %d", expectedUnique, len(enc.LiteralPool))
	}
}

// TestLiteralPool_LargeValues tests various large values that need literal pool
// Note: Some values can be encoded as MOV (rotated imm8) or MVN (~value), and won't use literal pool
func TestLiteralPool_LargeValues(t *testing.T) {
	tests := []struct {
		name  string
		value uint32
	}{
		// These values CANNOT be encoded as immediate and WILL use literal pool
		{"typical address", 0x80008000},
		{"odd pattern", 0xABCDEF12},
		{"alternating bits", 0x55555555},
		{"scattered bits", 0x12481248},
		{"prime-like pattern", 0x17171717},
		{"complex pattern", 0x12345678},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := encoder.NewEncoder(parser.NewSymbolTable())

			inst := &parser.Instruction{
				Mnemonic: "LDR",
				Operands: []string{"R0", fmt.Sprintf("=0x%08X", tt.value)},
			}

			_, err := enc.EncodeInstruction(inst, 0x8000)
			if err != nil {
				t.Fatalf("Failed to encode value 0x%08X: %v", tt.value, err)
			}

			// Verify value is in literal pool
			found := false
			for _, v := range enc.LiteralPool {
				if v == tt.value {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Value 0x%08X not found in literal pool", tt.value)
			}
		})
	}
}

// TestLiteralPool_SmallValuesUseMOV tests that small values don't use literal pool
func TestLiteralPool_SmallValuesUseMOV(t *testing.T) {
	tests := []struct {
		name  string
		value uint32
	}{
		{"zero", 0},
		{"small positive", 42},
		{"byte value", 0xFF},
		{"rotatable", 0x100}, // 1 rotated left by 8
		{"rotatable FF00", 0xFF00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := encoder.NewEncoder(parser.NewSymbolTable())

			inst := &parser.Instruction{
				Mnemonic: "LDR",
				Operands: []string{"R0", fmt.Sprintf("=0x%X", tt.value)},
			}

			result, err := enc.EncodeInstruction(inst, 0x8000)
			if err != nil {
				t.Fatalf("Failed to encode value 0x%X: %v", tt.value, err)
			}

			// For small values, should use MOV (data processing) not LDR
			// Data processing: bits 27:26 = 00
			// Memory (LDR): bits 27:26 = 01
			instrType := (result >> 26) & 0x3

			if instrType == 0x1 {
				t.Errorf("Value 0x%X used LDR (literal pool) instead of MOV", tt.value)
			}

			// Literal pool should be empty for encodable immediates
			if len(enc.LiteralPool) > 0 {
				t.Errorf("Literal pool should be empty for MOV-encodable value 0x%X, got %d entries",
					tt.value, len(enc.LiteralPool))
			}
		})
	}
}

// TestLiteralPool_InvertibleValuesUseMVN tests that invertible values use MVN
func TestLiteralPool_InvertibleValuesUseMVN(t *testing.T) {
	tests := []struct {
		name  string
		value uint32
	}{
		{"inverted zero", 0xFFFFFFFF}, // ~0
		{"inverted small", 0xFFFFFF00}, // ~0xFF
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := encoder.NewEncoder(parser.NewSymbolTable())

			inst := &parser.Instruction{
				Mnemonic: "LDR",
				Operands: []string{"R0", fmt.Sprintf("=0x%X", tt.value)},
			}

			result, err := enc.EncodeInstruction(inst, 0x8000)
			if err != nil {
				t.Fatalf("Failed to encode value 0x%X: %v", tt.value, err)
			}

			// Should use data processing (MOV/MVN) not LDR
			instrType := (result >> 26) & 0x3
			if instrType == 0x1 {
				t.Errorf("Value 0x%X used LDR instead of MVN", tt.value)
			}

			// Literal pool should be empty
			if len(enc.LiteralPool) > 0 {
				t.Errorf("Literal pool should be empty for MVN-encodable value 0x%X", tt.value)
			}
		})
	}
}

// TestLiteralPool_SequentialAddresses tests literals at sequential addresses
func TestLiteralPool_SequentialAddresses(t *testing.T) {
	enc := encoder.NewEncoder(parser.NewSymbolTable())

	// Use 20 different unencodable values
	const numLiterals = 20
	baseAddr := uint32(0x8000)

	for i := 0; i < numLiterals; i++ {
		value := uint32(0x87654321 + i*0x01010101) // Ensure each is unique and unencodable
		inst := &parser.Instruction{
			Mnemonic: "LDR",
			Operands: []string{fmt.Sprintf("R%d", i%16), fmt.Sprintf("=0x%08X", value)},
		}

		addr := baseAddr + uint32(i*4)
		result, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			t.Fatalf("Failed to encode literal %d at 0x%X: %v", i, addr, err)
		}

		// Verify the offset is within 12-bit limit (4095 bytes)
		offset := result & 0xFFF
		if offset > 4095 {
			t.Errorf("Literal %d: offset %d exceeds 12-bit limit", i, offset)
		}
	}

	if len(enc.LiteralPool) != numLiterals {
		t.Errorf("Expected %d literals, got %d", numLiterals, len(enc.LiteralPool))
	}
}

// TestLiteralPool_AllRegisters tests literal loads to all registers
func TestLiteralPool_AllRegisters(t *testing.T) {
	enc := encoder.NewEncoder(parser.NewSymbolTable())

	const litValue = uint32(0x12345678)

	for reg := 0; reg < 16; reg++ {
		t.Run(fmt.Sprintf("R%d", reg), func(t *testing.T) {
			regName := fmt.Sprintf("R%d", reg)
			inst := &parser.Instruction{
				Mnemonic: "LDR",
				Operands: []string{regName, fmt.Sprintf("=0x%08X", litValue+uint32(reg))},
			}

			addr := uint32(0x8000 + reg*4)
			result, err := enc.EncodeInstruction(inst, addr)
			if err != nil {
				t.Fatalf("Failed to encode LDR %s: %v", regName, err)
			}

			// Extract Rd field (bits 15:12)
			rd := (result >> 12) & 0xF
			if rd != uint32(reg) {
				t.Errorf("Expected Rd=%d, got %d", reg, rd)
			}
		})
	}
}

// TestLiteralPool_WithSymbols tests literal pool with symbolic expressions
func TestLiteralPool_WithSymbols(t *testing.T) {
	st := parser.NewSymbolTable()
	_ = st.Define("DATA_START", parser.SymbolLabel, 0x80001000, parser.Position{})
	_ = st.Define("OFFSET", parser.SymbolConstant, 0x100, parser.Position{})

	enc := encoder.NewEncoder(st)

	tests := []struct {
		name    string
		operand string
	}{
		{"simple label", "=DATA_START"},
		{"label plus immediate", "=DATA_START+4"},
		{"label minus immediate", "=DATA_START-4"},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "LDR",
				Operands: []string{"R0", tt.operand},
			}

			addr := uint32(0x8000 + i*4)
			_, err := enc.EncodeInstruction(inst, addr)
			if err != nil {
				t.Fatalf("Failed to encode %s: %v", tt.operand, err)
			}
		})
	}
}

// TestLiteralPool_ZeroPoolState tests initial pool state
func TestLiteralPool_ZeroPoolState(t *testing.T) {
	enc := encoder.NewEncoder(parser.NewSymbolTable())

	// New encoder should have empty literal pool
	if len(enc.LiteralPool) != 0 {
		t.Errorf("Expected empty literal pool on new encoder, got %d entries", len(enc.LiteralPool))
	}
}

// TestLiteralPool_Capacity tests pool capacity with many unique values
func TestLiteralPool_Capacity(t *testing.T) {
	enc := encoder.NewEncoder(parser.NewSymbolTable())

	// Add 100 unique values - this tests pool capacity handling
	const numValues = 100
	baseAddr := uint32(0x8000)

	for i := 0; i < numValues; i++ {
		// Generate a value that definitely can't be encoded as immediate
		// Use different bits at different positions
		value := uint32((i << 24) | ((255 - i) << 16) | ((i * 3) << 8) | (i ^ 0xAB))
		inst := &parser.Instruction{
			Mnemonic: "LDR",
			Operands: []string{"R0", fmt.Sprintf("=0x%08X", value)},
		}

		addr := baseAddr + uint32(i*4)
		_, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			t.Fatalf("Failed to encode literal %d: %v", i, err)
		}
	}

	// All values should be in the pool
	if len(enc.LiteralPool) != numValues {
		t.Errorf("Expected %d literals in pool, got %d", numValues, len(enc.LiteralPool))
	}
}
