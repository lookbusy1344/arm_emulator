package encoder_test

import "testing"

func TestEncodeHalfwordTransfers(t *testing.T) {
	tests := []struct {
		mnemonic string
		operands []string
		want     uint32
	}{
		// cccc 000P U1WL nnnn dddd hhhh 1011 llll (immediate), 0000 1011 mmmm (register)
		{"LDRH", []string{"R0", "[R1]"}, 0xE1D100B0},
		{"LDRH", []string{"R0", "[R1, #2]"}, 0xE1D100B2},
		{"LDRH", []string{"R0", "[R1, #0xFE]"}, 0xE1D10FBE},
		{"LDRH", []string{"R0", "[R1, #2]!"}, 0xE1F100B2},
		{"LDRH", []string{"R0", "[R1, R2]"}, 0xE19100B2},
		{"STRH", []string{"R3", "[R1, #2]"}, 0xE1C130B2},
		{"STRH", []string{"R3", "[R1, R2]"}, 0xE18130B2},
	}
	for _, tt := range tests {
		t.Run(tt.mnemonic+" "+tt.operands[1], func(t *testing.T) {
			got := encodeInstruction(t, newTestEncoder(), tt.mnemonic, tt.operands, 0x8000)
			if got != tt.want {
				t.Errorf("%s %v = 0x%08X, want 0x%08X", tt.mnemonic, tt.operands, got, tt.want)
			}
		})
	}
}
