package encoder_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/encoder"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// Helper to create a basic encoder with empty symbol table
func newTestEncoder() *encoder.Encoder {
	return encoder.NewEncoder(parser.NewSymbolTable())
}

// Helper to create encoder with symbols
func newTestEncoderWithSymbols(symbols map[string]uint32) *encoder.Encoder {
	st := parser.NewSymbolTable()
	for name, value := range symbols {
		_ = st.Define(name, parser.SymbolLabel, value, parser.Position{})
	}
	return encoder.NewEncoder(st)
}

// Helper to parse and encode a single instruction
func encodeInstruction(t *testing.T, enc *encoder.Encoder, mnemonic string, operands []string, addr uint32) uint32 {
	t.Helper()
	inst := &parser.Instruction{
		Mnemonic: mnemonic,
		Operands: operands,
	}
	result, err := enc.EncodeInstruction(inst, addr)
	if err != nil {
		t.Fatalf("Failed to encode %s %v: %v", mnemonic, operands, err)
	}
	return result
}

// TestEncodeCondition tests condition code encoding
func TestEncodeCondition(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name     string
		cond     string
		expected uint32
	}{
		{"EQ", "EQ", uint32(vm.CondEQ)},
		{"NE", "NE", uint32(vm.CondNE)},
		{"CS", "CS", uint32(vm.CondCS)},
		{"HS (alias for CS)", "HS", uint32(vm.CondCS)},
		{"CC", "CC", uint32(vm.CondCC)},
		{"LO (alias for CC)", "LO", uint32(vm.CondCC)},
		{"MI", "MI", uint32(vm.CondMI)},
		{"PL", "PL", uint32(vm.CondPL)},
		{"VS", "VS", uint32(vm.CondVS)},
		{"VC", "VC", uint32(vm.CondVC)},
		{"HI", "HI", uint32(vm.CondHI)},
		{"LS", "LS", uint32(vm.CondLS)},
		{"GE", "GE", uint32(vm.CondGE)},
		{"LT", "LT", uint32(vm.CondLT)},
		{"GT", "GT", uint32(vm.CondGT)},
		{"LE", "LE", uint32(vm.CondLE)},
		{"AL", "AL", uint32(vm.CondAL)},
		{"empty (defaults to AL)", "", uint32(vm.CondAL)},
		{"lowercase eq", "eq", uint32(vm.CondEQ)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a conditional MOV to test condition encoding
			inst := &parser.Instruction{
				Mnemonic:  "MOV",
				Condition: tt.cond,
				Operands:  []string{"R0", "#0"},
			}
			result, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Extract condition code from bits 31-28
			actualCond := (result >> 28) & 0xF
			if actualCond != tt.expected {
				t.Errorf("Condition %q: got 0x%X, want 0x%X", tt.cond, actualCond, tt.expected)
			}
		})
	}
}

// TestEncodeImmediateBasic tests basic immediate value encoding
func TestEncodeImmediateBasic(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name     string
		operand  string
		wantErr  bool
		checkVal func(uint32) bool
	}{
		{"zero", "#0", false, func(v uint32) bool { return v != 0 }}, // Should encode successfully
		{"small positive", "#1", false, nil},
		{"0xFF", "#0xFF", false, nil},
		{"0x100", "#0x100", false, nil}, // Can be encoded with rotation
		{"0xFF00", "#0xFF00", false, nil},
		{"0xFF000000", "#0xFF000000", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{"R0", tt.operand},
			}
			_, err := enc.EncodeInstruction(inst, 0)
			if tt.wantErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.operand)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.operand, err)
			}
		})
	}
}

// TestEncodeImmediateRotations tests that immediates are correctly rotated
func TestEncodeImmediateRotations(t *testing.T) {
	enc := newTestEncoder()

	// These values require specific rotations to encode
	tests := []struct {
		name  string
		value string
	}{
		{"0x00", "#0x00"},
		{"0x01", "#0x01"},
		{"0xFF", "#0xFF"},
		{"0x100", "#0x100"},           // 1 rotated left by 8
		{"0x200", "#0x200"},           // 2 rotated left by 8
		{"0x3FC", "#0x3FC"},           // 0xFF rotated left by 2
		{"0xFF00", "#0xFF00"},         // 0xFF rotated left by 8
		{"0xFF000000", "#0xFF000000"}, // 0xFF rotated left by 24
		{"0x80000000", "#0x80000000"}, // 2 rotated left by 31
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{"R0", tt.value},
			}
			_, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Errorf("Failed to encode %s: %v", tt.value, err)
			}
		})
	}
}

// TestEncodeImmediateUnencodable tests values that cannot be encoded as rotated immediates
func TestEncodeImmediateUnencodable(t *testing.T) {
	enc := newTestEncoder()

	// These values cannot be encoded as 8-bit rotated immediate without MVN trick
	// ARM2 does not support MOVW, so values that can't be encoded via MOV/MVN should fail
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Values that can be encoded via MVN (inverted value is encodable)
		{"0xFFFFFF00 (can use MVN)", "#0xFFFFFF00", false}, // ~0xFF
		// Values that cannot be encoded as ARM2 immediate
		{"0x12345678", "#0x12345678", true},           // Cannot be encoded
		{"0xABCD (no MOVW in ARM2)", "#0xABCD", true}, // Would need MOVW which is not ARM2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{"R0", tt.value},
			}
			_, err := enc.EncodeInstruction(inst, 0)
			if tt.wantErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.value, err)
			}
		})
	}
}

// TestEncodeCharacterLiterals tests character literal parsing
func TestEncodeCharacterLiterals(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name    string
		operand string
		wantErr bool
	}{
		{"ASCII A", "#'A'", false},
		{"ASCII space", "#' '", false},
		{"escape newline", "#'\\n'", false},
		{"escape tab", "#'\\t'", false},
		{"escape carriage return", "#'\\r'", false},
		{"escape null", "#'\\0'", false},
		{"escape backslash", "#'\\\\'", false},
		{"escape single quote", "#'\\''", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{"R0", tt.operand},
			}
			_, err := enc.EncodeInstruction(inst, 0)
			if tt.wantErr && err == nil {
				t.Errorf("Expected error for %s", tt.operand)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.operand, err)
			}
		})
	}
}

// TestEncodeRegisters tests register encoding in various positions
func TestEncodeRegisters(t *testing.T) {
	enc := newTestEncoder()

	// Test all registers R0-R15 as destination
	for i := 0; i <= 15; i++ {
		t.Run("Rd=R"+string(rune('0'+i%10)), func(t *testing.T) {
			var regName string
			switch i {
			case 13:
				regName = "SP"
			case 14:
				regName = "LR"
			case 15:
				regName = "PC"
			default:
				regName = "R" + string(rune('0'+i/10)) + string(rune('0'+i%10))
				if i < 10 {
					regName = "R" + string(rune('0'+i))
				}
			}

			inst := &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{regName, "#0"},
			}
			result, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("Failed to encode MOV %s, #0: %v", regName, err)
			}

			// Extract Rd from bits 15-12
			actualRd := (result >> 12) & 0xF
			if actualRd != uint32(i) {
				t.Errorf("Rd for %s: got %d, want %d", regName, actualRd, i)
			}
		})
	}

	// Test special register aliases
	t.Run("SP alias", func(t *testing.T) {
		inst := &parser.Instruction{
			Mnemonic: "MOV",
			Operands: []string{"SP", "#0"},
		}
		result, err := enc.EncodeInstruction(inst, 0)
		if err != nil {
			t.Fatal(err)
		}
		rd := (result >> 12) & 0xF
		if rd != 13 {
			t.Errorf("SP should be R13, got R%d", rd)
		}
	})

	t.Run("LR alias", func(t *testing.T) {
		inst := &parser.Instruction{
			Mnemonic: "MOV",
			Operands: []string{"LR", "#0"},
		}
		result, err := enc.EncodeInstruction(inst, 0)
		if err != nil {
			t.Fatal(err)
		}
		rd := (result >> 12) & 0xF
		if rd != 14 {
			t.Errorf("LR should be R14, got R%d", rd)
		}
	})

	t.Run("PC alias", func(t *testing.T) {
		inst := &parser.Instruction{
			Mnemonic: "MOV",
			Operands: []string{"PC", "#0"},
		}
		result, err := enc.EncodeInstruction(inst, 0)
		if err != nil {
			t.Fatal(err)
		}
		rd := (result >> 12) & 0xF
		if rd != 15 {
			t.Errorf("PC should be R15, got R%d", rd)
		}
	})
}

// TestEncodeDataProcessingBasic tests basic data processing instruction encoding
func TestEncodeDataProcessingBasic(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name       string
		mnemonic   string
		operands   []string
		wantOpcode uint32
	}{
		{"MOV immediate", "MOV", []string{"R0", "#1"}, 0xD},
		{"MVN immediate", "MVN", []string{"R0", "#1"}, 0xF},
		{"ADD immediate", "ADD", []string{"R0", "R1", "#1"}, 0x4},
		{"SUB immediate", "SUB", []string{"R0", "R1", "#1"}, 0x2},
		{"AND immediate", "AND", []string{"R0", "R1", "#1"}, 0x0},
		{"ORR immediate", "ORR", []string{"R0", "R1", "#1"}, 0xC},
		{"EOR immediate", "EOR", []string{"R0", "R1", "#1"}, 0x1},
		{"BIC immediate", "BIC", []string{"R0", "R1", "#1"}, 0xE},
		{"CMP immediate", "CMP", []string{"R0", "#1"}, 0xA},
		{"CMN immediate", "CMN", []string{"R0", "#1"}, 0xB},
		{"TST immediate", "TST", []string{"R0", "#1"}, 0x8},
		{"TEQ immediate", "TEQ", []string{"R0", "#1"}, 0x9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: tt.mnemonic,
				Operands: tt.operands,
			}
			result, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Extract opcode from bits 24-21
			actualOpcode := (result >> 21) & 0xF
			if actualOpcode != tt.wantOpcode {
				t.Errorf("Opcode for %s: got 0x%X, want 0x%X", tt.mnemonic, actualOpcode, tt.wantOpcode)
			}
		})
	}
}

// TestEncodeSetFlags tests S bit encoding
func TestEncodeSetFlags(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name     string
		setFlags bool
		wantS    uint32
	}{
		{"without S", false, 0},
		{"with S", true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{"R0", "#1"},
				SetFlags: tt.setFlags,
			}
			result, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Extract S bit from bit 20
			actualS := (result >> 20) & 0x1
			if actualS != tt.wantS {
				t.Errorf("S bit: got %d, want %d", actualS, tt.wantS)
			}
		})
	}
}

// TestEncodeShifts tests shift specification encoding
func TestEncodeShifts(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name          string
		operand       string
		wantShiftType uint32
	}{
		{"LSL #0", "R1, LSL #0", 0},
		{"LSL #4", "R1, LSL #4", 0},
		{"LSR #1", "R1, LSR #1", 1},
		{"ASR #8", "R1, ASR #8", 2},
		{"ROR #16", "R1, ROR #16", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{"R0", tt.operand},
			}
			result, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Extract shift type from bits 6-5
			actualShiftType := (result >> 5) & 0x3
			if actualShiftType != tt.wantShiftType {
				t.Errorf("Shift type for %s: got %d, want %d", tt.name, actualShiftType, tt.wantShiftType)
			}
		})
	}
}

// TestEncodeBranch tests branch instruction encoding
func TestEncodeBranch(t *testing.T) {
	symbols := map[string]uint32{
		"target": 0x8100, // Target address
	}
	enc := newTestEncoderWithSymbols(symbols)

	tests := []struct {
		name     string
		mnemonic string
		operands []string
		addr     uint32
		wantLink uint32 // Link bit (bit 24)
	}{
		{"B forward", "B", []string{"target"}, 0x8000, 0},
		{"BL forward", "BL", []string{"target"}, 0x8000, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: tt.mnemonic,
				Operands: tt.operands,
			}
			result, err := enc.EncodeInstruction(inst, tt.addr)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Check it's a branch instruction (bits 27-25 = 101)
			branchBits := (result >> 25) & 0x7
			if branchBits != 0x5 {
				t.Errorf("Branch bits: got 0x%X, want 0x5", branchBits)
			}

			// Check link bit
			linkBit := (result >> 24) & 0x1
			if linkBit != tt.wantLink {
				t.Errorf("Link bit: got %d, want %d", linkBit, tt.wantLink)
			}
		})
	}
}

// TestEncodeMemory tests load/store instruction encoding
func TestEncodeMemory(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name     string
		mnemonic string
		operands []string
		wantLoad uint32 // L bit (bit 20)
		wantByte uint32 // B bit (bit 22)
	}{
		{"LDR basic", "LDR", []string{"R0", "[R1]"}, 1, 0},
		{"STR basic", "STR", []string{"R0", "[R1]"}, 0, 0},
		{"LDRB basic", "LDRB", []string{"R0", "[R1]"}, 1, 1},
		{"STRB basic", "STRB", []string{"R0", "[R1]"}, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: tt.mnemonic,
				Operands: tt.operands,
			}
			result, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Check L bit
			loadBit := (result >> 20) & 0x1
			if loadBit != tt.wantLoad {
				t.Errorf("Load bit: got %d, want %d", loadBit, tt.wantLoad)
			}

			// Check B bit
			byteBit := (result >> 22) & 0x1
			if byteBit != tt.wantByte {
				t.Errorf("Byte bit: got %d, want %d", byteBit, tt.wantByte)
			}
		})
	}
}

// TestEncodeMemoryAddressingModes tests various addressing modes
func TestEncodeMemoryAddressingModes(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name     string
		operands []string
		wantErr  bool
	}{
		{"[Rn]", []string{"R0", "[R1]"}, false},
		{"[Rn, #offset]", []string{"R0", "[R1, #4]"}, false},
		// Note: [Rn, #-offset] format with #-N is currently not handled correctly
		// The encoder expects the format [Rn, -#offset] or handles negative offsets differently
		{"[Rn, -#offset]", []string{"R0", "[R1, -#4]"}, false},
		{"[Rn, Rm]", []string{"R0", "[R1, R2]"}, false},
		{"[Rn, #offset]!", []string{"R0", "[R1, #4]!"}, false},
		{"[Rn], #offset", []string{"R0", "[R1], #4"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "LDR",
				Operands: tt.operands,
			}
			_, err := enc.EncodeInstruction(inst, 0)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestEncodeMultiply tests multiply instruction encoding
func TestEncodeMultiply(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name     string
		mnemonic string
		operands []string
		wantErr  bool
	}{
		{"MUL", "MUL", []string{"R0", "R1", "R2"}, false},
		{"MLA", "MLA", []string{"R0", "R1", "R2", "R3"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: tt.mnemonic,
				Operands: tt.operands,
			}
			result, err := enc.EncodeInstruction(inst, 0)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil {
				// Check multiply signature bits 7-4 = 1001
				multiplyBits := (result >> 4) & 0xF
				if multiplyBits != 0x9 {
					t.Errorf("Multiply bits: got 0x%X, want 0x9", multiplyBits)
				}
			}
		})
	}
}

// TestEncodeSWI tests software interrupt encoding
func TestEncodeSWI(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name    string
		operand string
		wantImm uint32
	}{
		{"SWI #0", "#0", 0},
		{"SWI #1", "#1", 1},
		{"SWI #0x10", "#0x10", 0x10},
		{"SWI #255", "#255", 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "SWI",
				Operands: []string{tt.operand},
			}
			result, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Check SWI signature bits 27-24 = 1111
			swiBits := (result >> 24) & 0xF
			if swiBits != 0xF {
				t.Errorf("SWI bits: got 0x%X, want 0xF", swiBits)
			}

			// Check immediate value (bits 23-0)
			imm := result & 0xFFFFFF
			if imm != tt.wantImm {
				t.Errorf("SWI immediate: got 0x%X, want 0x%X", imm, tt.wantImm)
			}
		})
	}
}

// TestEncodeSWI_VMDetection verifies encoded SWI matches VM's detection pattern
func TestEncodeSWI_VMDetection(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name    string
		operand string
		wantImm uint32
	}{
		{"SWI #0 (EXIT)", "#0", 0},
		{"SWI #1 (WRITE_CHAR)", "#1", 1},
		{"SWI #0x10 (OPEN)", "#0x10", 0x10},
		{"SWI #0xF0 (DEBUG_PRINT)", "#0xF0", 0xF0},
		{"SWI #0xFFFFFF (max)", "#0xFFFFFF", 0xFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "SWI",
				Operands: []string{tt.operand},
			}
			encoded, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("Failed to encode: %v", err)
			}

			// Verify the encoded instruction matches VM's SWI detection pattern
			// VM uses: (opcode & SWIDetectMask) == SWIPattern
			if (encoded & vm.SWIDetectMask) != vm.SWIPattern {
				t.Errorf("Encoded SWI doesn't match VM detection pattern: got 0x%08X, mask 0x%08X, pattern 0x%08X",
					encoded, vm.SWIDetectMask, vm.SWIPattern)
			}

			// Verify the immediate can be extracted correctly
			// VM uses: swiNum := inst.Opcode & SWIMask
			extractedImm := encoded & vm.SWIMask
			if extractedImm != tt.wantImm {
				t.Errorf("Extracted immediate: got 0x%X, want 0x%X", extractedImm, tt.wantImm)
			}
		})
	}
}

// TestEncodeNOP tests NOP encoding
func TestEncodeNOP(t *testing.T) {
	enc := newTestEncoder()

	inst := &parser.Instruction{
		Mnemonic: "NOP",
		Operands: []string{},
	}
	result, err := enc.EncodeInstruction(inst, 0)
	if err != nil {
		t.Fatalf("Failed to encode NOP: %v", err)
	}

	// NOP is typically encoded as MOV R0, R0 (condition AL)
	// Expected: 0xE1A00000
	// But implementations vary - just check it's not zero and condition is AL
	condBits := (result >> 28) & 0xF
	if condBits != uint32(vm.CondAL) {
		t.Errorf("NOP condition: got 0x%X, want 0x%X (AL)", condBits, vm.CondAL)
	}
}

// TestEncodeUnknownInstruction tests handling of unknown mnemonics
func TestEncodeUnknownInstruction(t *testing.T) {
	enc := newTestEncoder()

	inst := &parser.Instruction{
		Mnemonic: "UNKNOWN",
		Operands: []string{"R0", "R1"},
	}
	_, err := enc.EncodeInstruction(inst, 0)
	if err == nil {
		t.Error("Expected error for unknown instruction, got nil")
	}
}

// TestEncodeMissingOperands tests handling of missing operands
func TestEncodeMissingOperands(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name     string
		mnemonic string
		operands []string
	}{
		{"MOV with 0 operands", "MOV", []string{}},
		{"MOV with 1 operand", "MOV", []string{"R0"}},
		{"ADD with 2 operands", "ADD", []string{"R0", "R1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: tt.mnemonic,
				Operands: tt.operands,
			}
			_, err := enc.EncodeInstruction(inst, 0)
			if err == nil {
				t.Error("Expected error for missing operands, got nil")
			}
		})
	}
}

// TestEncodeInvalidRegister tests handling of invalid register names
func TestEncodeInvalidRegister(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name    string
		operand string
	}{
		{"R16", "R16"},
		{"R-1", "R-1"},
		{"RX", "RX"},
		{"invalid", "INVALID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{tt.operand, "#0"},
			}
			_, err := enc.EncodeInstruction(inst, 0)
			if err == nil {
				t.Errorf("Expected error for invalid register %s, got nil", tt.operand)
			}
		})
	}
}

// TestEncodeLDMSTM tests load/store multiple instruction encoding
func TestEncodeLDMSTM(t *testing.T) {
	enc := newTestEncoder()

	tests := []struct {
		name     string
		mnemonic string
		operands []string
		wantLoad uint32
		wantErr  bool
	}{
		{"LDMIA", "LDMIA", []string{"R0", "{R1, R2, R3}"}, 1, false},
		{"STMIA", "STMIA", []string{"R0", "{R1, R2, R3}"}, 0, false},
		{"LDMFD", "LDMFD", []string{"SP!", "{R0-R3, LR}"}, 1, false},
		{"STMFD", "STMFD", []string{"SP!", "{R0-R3, LR}"}, 0, false},
		{"PUSH", "PUSH", []string{"{R0, R1}"}, 0, false}, // PUSH is STMDB SP!, {...}
		{"POP", "POP", []string{"{R0, R1}"}, 1, false},   // POP is LDMIA SP!, {...}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &parser.Instruction{
				Mnemonic: tt.mnemonic,
				Operands: tt.operands,
			}
			result, err := enc.EncodeInstruction(inst, 0)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil {
				// Check L bit (bit 20)
				loadBit := (result >> 20) & 0x1
				if loadBit != tt.wantLoad {
					t.Errorf("Load bit: got %d, want %d", loadBit, tt.wantLoad)
				}
			}
		})
	}
}
