package parser_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

func TestLexer_BasicTokens(t *testing.T) {
	input := "MOV R0, #42"
	lexer := parser.NewLexer(input, "test.s")

	expectedTokens := []parser.TokenType{
		parser.TokenIdentifier, // MOV
		parser.TokenRegister,   // R0
		parser.TokenComma,      // ,
		parser.TokenHash,       // #
		parser.TokenNumber,     // 42
		parser.TokenEOF,
	}

	for i, expected := range expectedTokens {
		tok := lexer.NextToken()
		if tok.Type != expected {
			t.Errorf("token %d: expected %v, got %v", i, expected, tok.Type)
		}
	}
}

func TestLexer_Labels(t *testing.T) {
	input := "loop: ADD R1, R1, #1"
	lexer := parser.NewLexer(input, "test.s")

	tok := lexer.NextToken()
	if tok.Type != parser.TokenIdentifier || tok.Literal != "loop" {
		t.Errorf("expected label 'loop', got %v %q", tok.Type, tok.Literal)
	}

	tok = lexer.NextToken()
	if tok.Type != parser.TokenColon {
		t.Errorf("expected colon, got %v", tok.Type)
	}
}

func TestLexer_Comments(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"; line comment", " line comment"},
		{"// line comment", " line comment"},
		{"/* block comment */", " block comment "},
	}

	for _, tt := range tests {
		lexer := parser.NewLexer(tt.input, "test.s")
		tok := lexer.NextToken()
		if tok.Type != parser.TokenComment {
			t.Errorf("input %q: expected comment, got %v", tt.input, tok.Type)
		}
	}
}

func TestLexer_Numbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"0x2A", "0x2A"},
		{"0b101010", "0b101010"},
		{"0o52", "0o52"},
	}

	for _, tt := range tests {
		lexer := parser.NewLexer(tt.input, "test.s")
		tok := lexer.NextToken()
		if tok.Type != parser.TokenNumber {
			t.Errorf("input %q: expected number, got %v", tt.input, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, tok.Literal)
		}
	}
}

func TestLexer_Registers(t *testing.T) {
	registers := []string{"R0", "R1", "R15", "SP", "LR", "PC"}

	for _, reg := range registers {
		lexer := parser.NewLexer(reg, "test.s")
		tok := lexer.NextToken()
		if tok.Type != parser.TokenRegister {
			t.Errorf("register %q: expected TokenRegister, got %v", reg, tok.Type)
		}
	}
}

func TestLexer_Directives(t *testing.T) {
	directives := []string{".org", ".equ", ".word", ".byte", ".ascii"}

	for _, dir := range directives {
		lexer := parser.NewLexer(dir, "test.s")
		tok := lexer.NextToken()
		if tok.Type != parser.TokenDirective {
			t.Errorf("directive %q: expected TokenDirective, got %v", dir, tok.Type)
		}
		if tok.Literal != dir {
			t.Errorf("directive: expected %q, got %q", dir, tok.Literal)
		}
	}
}

func TestLexer_MemoryOperands(t *testing.T) {
	input := "[R1, #4]!"
	lexer := parser.NewLexer(input, "test.s")

	expectedTokens := []parser.TokenType{
		parser.TokenLBracket, // [
		parser.TokenRegister, // R1
		parser.TokenComma,    // ,
		parser.TokenHash,     // #
		parser.TokenNumber,   // 4
		parser.TokenRBracket, // ]
		parser.TokenExclaim,  // !
		parser.TokenEOF,
	}

	for i, expected := range expectedTokens {
		tok := lexer.NextToken()
		if tok.Type != expected {
			t.Errorf("token %d: expected %v, got %v", i, expected, tok.Type)
		}
	}
}

func TestLexer_Strings(t *testing.T) {
	input := `"Hello, World!"`
	lexer := parser.NewLexer(input, "test.s")

	tok := lexer.NextToken()
	if tok.Type != parser.TokenString {
		t.Errorf("expected string, got %v", tok.Type)
	}
	if tok.Literal != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", tok.Literal)
	}
}

func TestLexer_ComplexInstruction(t *testing.T) {
	input := "ADDSEQ R0, R1, R2, LSL #2 ; Add with shift"
	lexer := parser.NewLexer(input, "test.s")

	tok := lexer.NextToken()
	if tok.Type != parser.TokenIdentifier || tok.Literal != "ADDSEQ" {
		t.Errorf("expected identifier 'ADDSEQ', got %v %q", tok.Type, tok.Literal)
	}
}

func TestLexer_Newlines(t *testing.T) {
	input := "MOV R0, #1\nMOV R1, #2"
	lexer := parser.NewLexer(input, "test.s")

	// MOV
	lexer.NextToken()
	// R0
	lexer.NextToken()
	// ,
	lexer.NextToken()
	// #
	lexer.NextToken()
	// 1
	lexer.NextToken()

	// Should get newline
	tok := lexer.NextToken()
	if tok.Type != parser.TokenNewline {
		t.Errorf("expected newline, got %v", tok.Type)
	}
}
