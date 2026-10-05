package encoder_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

func TestEncodeMultiplyRegisters(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		operands []string
		want     uint32
	}{
		// cccc 0000 00AS dddd nnnn ssss 1001 mmmm
		{"MUL", "MUL", []string{"R0", "R1", "R2"}, 0xE0000291},
		{"MLA", "MLA", []string{"R0", "R1", "R2", "R3"}, 0xE0203291},
		// Rd == Rm is unpredictable on ARM2; the product commutes, so Rm and Rs swap.
		{"MUL Rd==Rm swaps", "MUL", []string{"R0", "R0", "R1"}, 0xE0000091},
		{"MLA Rd==Rm swaps", "MLA", []string{"R2", "R2", "R3", "R4"}, 0xE0224293},
		{"MUL Rd==Rs needs no swap", "MUL", []string{"R0", "R1", "R0"}, 0xE0000091},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeInstruction(t, newTestEncoder(), tt.mnemonic, tt.operands, 0x8000)
			if got != tt.want {
				t.Errorf("%s %v = 0x%08X, want 0x%08X", tt.mnemonic, tt.operands, got, tt.want)
			}
		})
	}
}

func TestEncodeMultiplyErrors(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		operands []string
	}{
		{"MUL all same", "MUL", []string{"R0", "R0", "R0"}},
		{"MLA all same", "MLA", []string{"R1", "R1", "R1", "R2"}},
		{"MUL PC destination", "MUL", []string{"PC", "R1", "R2"}},
		{"MUL PC operand", "MUL", []string{"R0", "R15", "R2"}},
		{"MLA PC accumulator", "MLA", []string{"R0", "R1", "R2", "PC"}},
		{"MUL too few", "MUL", []string{"R0", "R1"}},
		{"MUL too many", "MUL", []string{"R0", "R1", "R2", "R3"}},
		{"MLA too many", "MLA", []string{"R0", "R1", "R2", "R3", "R4"}},
		{"MUL immediate", "MUL", []string{"R0", "R1", "#2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{Mnemonic: tt.mnemonic, Operands: tt.operands}
			if got, err := newTestEncoder().EncodeInstruction(inst, 0x8000); err == nil {
				t.Errorf("%s %v encoded as 0x%08X, want error", tt.mnemonic, tt.operands, got)
			}
		})
	}
}
