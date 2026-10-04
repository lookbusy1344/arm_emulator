package encoder_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/encoder"
	"github.com/lookbusy1344/arm-emulator/parser"
)

// encodeSource parses one instruction and encodes it.
func encodeSource(t *testing.T, line string) (uint32, error) {
	t.Helper()
	program, err := parser.NewParser(".org 0x8000\n_start:\n\t"+line+"\n", "test.s").Parse()
	if err != nil {
		t.Fatalf("parse %q: %v", line, err)
	}
	if len(program.Instructions) != 1 {
		t.Fatalf("expected one instruction from %q", line)
	}
	return encoder.NewEncoder(program.SymbolTable).EncodeInstruction(program.Instructions[0], 0x8000)
}

func TestEncodeShiftEdgeAmounts(t *testing.T) {
	tests := []struct {
		line string
		want uint32
	}{
		{"MOV R0, R1, LSR #32", 0xE1A00021},
		{"MOV R0, R1, ASR #32", 0xE1A00041},
		{"MOVS R0, R1, RRX", 0xE1B00061},
		{"ADD R0, R1, R2, RRX", 0xE0810062},
		{"MOV R0, R1, LSL #0", 0xE1A00001},
		// A zero LSR/ASR/ROR amount is no shift; its encoding would mean #32 or RRX.
		{"MOV R0, R1, LSR #0", 0xE1A00001},
		{"MOV R0, R1, ASR #0", 0xE1A00001},
		{"MOV R0, R1, ROR #0", 0xE1A00001},
		{"MOV R0, R1, LSR #31", 0xE1A00FA1},
		{"LDR R0, [R1, R2, LSR #32]", 0xE7910022},
		{"LDR R0, [R1, R2, RRX]", 0xE7910062},
	}
	for _, tt := range tests {
		got, err := encodeSource(t, tt.line)
		if err != nil {
			t.Errorf("%s: %v", tt.line, err)
			continue
		}
		if got != tt.want {
			t.Errorf("%s = 0x%08X, want 0x%08X", tt.line, got, tt.want)
		}
	}
}

func TestEncodeShiftRejectsOutOfRange(t *testing.T) {
	for _, line := range []string{
		"MOV R0, R1, LSL #32",
		"MOV R0, R1, ROR #32",
		"MOV R0, R1, LSR #33",
		"MOV R0, R1, RRX #1",
	} {
		if _, err := encodeSource(t, line); err == nil {
			t.Errorf("%s: expected an error", line)
		}
	}
}
