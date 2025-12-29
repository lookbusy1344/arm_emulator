package parser_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/encoder"
	"github.com/lookbusy1344/arm-emulator/parser"
)

// TestCharacterLiteral_Basic tests basic character literal parsing
func TestCharacterLiteral_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint32
	}{
		{"space character", "MOV R0, #' '\n", 32},
		{"letter A", "MOV R0, #'A'\n", 65},
		{"letter Z", "MOV R0, #'Z'\n", 90},
		{"digit 0", "MOV R0, #'0'\n", 48},
		{"digit 9", "MOV R0, #'9'\n", 57},
		{"comma", "MOV R0, #','\n", 44},
		{"semicolon", "MOV R0, #';'\n", 59},
		{"exclamation", "MOV R0, #'!'\n", 33},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]
			if inst.Mnemonic != "MOV" {
				t.Errorf("expected MOV, got %s", inst.Mnemonic)
			}

			// Encode and verify the immediate value
			enc := encoder.NewEncoder(program.SymbolTable)
			machineCode, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			// Extract immediate value from encoded instruction
			// MOV R0, #imm is a data processing instruction
			// Format: [cond:4][00][I:1][opcode:4][S:1][Rn:4][Rd:4][operand2:12]
			// For immediate: operand2 = [rotate:4][imm8:8]
			operand2 := machineCode & 0xFF // Get lower 8 bits (unrotated immediate)

			if operand2 != tt.expected {
				t.Errorf("expected character value %d (0x%X), got %d (0x%X)",
					tt.expected, tt.expected, operand2, operand2)
			}
		})
	}
}

// TestCharacterLiteral_EscapeSequences tests escape sequence parsing
func TestCharacterLiteral_EscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint32
	}{
		{"newline", "MOV R0, #'\\n'\n", 10},
		{"tab", "MOV R0, #'\\t'\n", 9},
		{"carriage return", "MOV R0, #'\\r'\n", 13},
		{"null", "MOV R0, #'\\0'\n", 0},
		{"backslash", "MOV R0, #'\\\\'\n", 92},
		{"single quote", "MOV R0, #'\\''\n", 39},
		{"double quote", "MOV R0, #'\\\"'\n", 34},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]

			// Encode and verify
			enc := encoder.NewEncoder(program.SymbolTable)
			machineCode, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			operand2 := machineCode & 0xFF
			if operand2 != tt.expected {
				t.Errorf("expected escape sequence value %d (0x%X), got %d (0x%X)",
					tt.expected, tt.expected, operand2, operand2)
			}
		})
	}
}

// TestCharacterLiteral_InComparisons tests character literals in CMP instructions
func TestCharacterLiteral_InComparisons(t *testing.T) {
	input := `
		CMP R0, #'A'
		CMP R1, #'Z'
		CMP R2, #' '
	`

	p := parser.NewParser(input, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Instructions) != 3 {
		t.Fatalf("expected 3 instructions, got %d", len(program.Instructions))
	}

	expectedValues := []uint32{65, 90, 32} // 'A', 'Z', ' '

	enc := encoder.NewEncoder(program.SymbolTable)
	for i, inst := range program.Instructions {
		if inst.Mnemonic != "CMP" {
			t.Errorf("instruction %d: expected CMP, got %s", i, inst.Mnemonic)
			continue
		}

		// Safe conversion: i is from range, always >= 0 and bounded by instruction count
		// #nosec G115 -- i is loop index, guaranteed non-negative and within bounds
		addr := uint32(i * 4)
		machineCode, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			t.Fatalf("instruction %d encode error: %v", i, err)
		}

		operand2 := machineCode & 0xFF
		if operand2 != expectedValues[i] {
			t.Errorf("instruction %d: expected value %d, got %d",
				i, expectedValues[i], operand2)
		}
	}
}

// TestCharacterLiteral_InDataProcessing tests character literals in various instructions
func TestCharacterLiteral_InDataProcessing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		mnemonic string
		expected uint32
	}{
		{"ADD with char", "ADD R0, R1, #'0'\n", "ADD", 48},
		{"SUB with char", "SUB R2, R3, #' '\n", "SUB", 32},
		{"AND with char", "AND R4, R5, #'\\n'\n", "AND", 10},
		{"ORR with char", "ORR R6, R7, #'A'\n", "ORR", 65},
		{"EOR with char", "EOR R8, R9, #'\\t'\n", "EOR", 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]
			if inst.Mnemonic != tt.mnemonic {
				t.Errorf("expected %s, got %s", tt.mnemonic, inst.Mnemonic)
			}

			enc := encoder.NewEncoder(program.SymbolTable)
			machineCode, err := enc.EncodeInstruction(inst, 0)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			operand2 := machineCode & 0xFF
			if operand2 != tt.expected {
				t.Errorf("expected value %d (0x%X), got %d (0x%X)",
					tt.expected, tt.expected, operand2, operand2)
			}
		})
	}
}

// TestCharacterLiteral_InvalidEscapes tests error handling for invalid escape sequences
func TestCharacterLiteral_InvalidEscapes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid escape x", "MOV R0, #'\\x'\n"},
		{"invalid escape z", "MOV R0, #'\\z'\n"},
		{"invalid escape 8", "MOV R0, #'\\8'\n"}, // 8 and 9 are not valid octal digits
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()
			if err != nil {
				// Parse might succeed, encoding should fail
				return
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			enc := encoder.NewEncoder(program.SymbolTable)
			_, err = enc.EncodeInstruction(program.Instructions[0], 0)
			if err == nil {
				t.Errorf("expected encoding error for invalid escape sequence, got none")
			}
		})
	}
}

// TestCharacterLiteral_Lexer tests the lexer's handling of character literals
func TestCharacterLiteral_Lexer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple char", "#'A'", "A"},
		{"space char", "#' '", " "},
		{"escaped tab", "#'\\t'", "\\t"},
		{"escaped newline", "#'\\n'", "\\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := parser.NewLexer(tt.input, "test.s")

			// Should get # token
			tok := lexer.NextToken()
			if tok.Type != parser.TokenHash {
				t.Fatalf("expected TokenHash, got %v", tok.Type)
			}

			// Should get string token with character literal
			tok = lexer.NextToken()
			if tok.Type != parser.TokenString {
				t.Fatalf("expected TokenString, got %v", tok.Type)
			}

			if tok.Literal != tt.expected {
				t.Errorf("expected literal %q, got %q", tt.expected, tok.Literal)
			}
		})
	}
}

// TestCharacterLiteral_MultipleInSameProgram tests multiple character literals in one program
func TestCharacterLiteral_MultipleInSameProgram(t *testing.T) {
	input := `
		MOV R0, #'H'
		MOV R1, #'e'
		MOV R2, #'l'
		MOV R3, #'l'
		MOV R4, #'o'
		MOV R5, #' '
		MOV R6, #'!'
	`

	p := parser.NewParser(input, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Instructions) != 7 {
		t.Fatalf("expected 7 instructions, got %d", len(program.Instructions))
	}

	expectedChars := []uint32{72, 101, 108, 108, 111, 32, 33} // "Hello !"

	enc := encoder.NewEncoder(program.SymbolTable)
	for i, inst := range program.Instructions {
		// Safe conversion: i is from range, always >= 0 and bounded by instruction count
		// #nosec G115 -- i is loop index, guaranteed non-negative and within bounds
		addr := uint32(i * 4)
		machineCode, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			t.Fatalf("instruction %d encode error: %v", i, err)
		}

		operand2 := machineCode & 0xFF
		if operand2 != expectedChars[i] {
			t.Errorf("instruction %d: expected char value %d ('%c'), got %d",
				i, expectedChars[i], rune(expectedChars[i]), operand2)
		}
	}
}
