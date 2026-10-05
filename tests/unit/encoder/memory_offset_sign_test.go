package encoder_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

func TestEncodeMemoryOffsetSign(t *testing.T) {
	tests := []struct {
		mnemonic string
		operands []string
		want     uint32
	}{
		// cccc 01IP UBWL nnnn dddd oooooooooooo
		{"LDR", []string{"R0", "[R1, #4]"}, 0xE5910004},
		{"LDR", []string{"R0", "[R1, #-4]"}, 0xE5110004},
		{"LDR", []string{"R0", "[R1, -#4]"}, 0xE5110004},
		{"LDR", []string{"R0", "[R1, #+4]"}, 0xE5910004},
		{"LDR", []string{"R0", "[R1, #-4]!"}, 0xE5310004},
		{"LDR", []string{"R0", "[R1]", "#-4"}, 0xE4110004},
		{"STR", []string{"R0", "[R1, #-8]"}, 0xE5010008},
		{"LDRB", []string{"R0", "[R1, #-1]"}, 0xE5510001},
		{"LDR", []string{"R0", "[R1, #-K]"}, 0xE5110004},
		{"LDR", []string{"R0", "[R1, -R2]"}, 0xE7110002},
		{"LDR", []string{"R0", "[R1, +R2]"}, 0xE7910002},
		{"LDR", []string{"R0", "[R1, #-4095]"}, 0xE5110FFF},
		// cccc 000P U1WL nnnn dddd hhhh 1SH1 llll
		{"LDRH", []string{"R0", "[R1, #-2]"}, 0xE15100B2},
		{"LDRH", []string{"R0", "[R1, -#2]"}, 0xE15100B2},
		{"LDRH", []string{"R0", "[R1, #-2]!"}, 0xE17100B2},
		{"LDRH", []string{"R0", "[R1]", "#-2"}, 0xE05100B2},
		{"STRH", []string{"R0", "[R1, #-0x12]"}, 0xE14101B2},
		{"LDRH", []string{"R0", "[R1, -R2]"}, 0xE11100B2},
	}
	for _, tt := range tests {
		t.Run(tt.mnemonic+" "+tt.operands[1], func(t *testing.T) {
			enc := newTestEncoderWithSymbols(map[string]uint32{"K": 4})
			got := encodeInstruction(t, enc, tt.mnemonic, tt.operands, 0x8000)
			if got != tt.want {
				t.Errorf("%s %v = 0x%08X, want 0x%08X", tt.mnemonic, tt.operands, got, tt.want)
			}
		})
	}
}

func TestEncodeMemoryOffsetOutOfRange(t *testing.T) {
	tests := []struct {
		mnemonic string
		operands []string
	}{
		{"LDR", []string{"R0", "[R1, #-4096]"}},
		{"LDR", []string{"R0", "[R1, #4096]"}},
		{"LDRH", []string{"R0", "[R1, #-256]"}},
		{"LDR", []string{"R0", "[R1, #--4]"}},
	}
	for _, tt := range tests {
		t.Run(tt.mnemonic+" "+tt.operands[1], func(t *testing.T) {
			inst := &parser.Instruction{Mnemonic: tt.mnemonic, Operands: tt.operands}
			if got, err := newTestEncoder().EncodeInstruction(inst, 0x8000); err == nil {
				t.Errorf("%v encoded as 0x%08X, want error", tt.operands, got)
			}
		})
	}
}
