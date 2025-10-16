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

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		token    parser.TokenType
		expected string
	}{
		{parser.TokenEOF, "EOF"},
		{parser.TokenIdentifier, "IDENTIFIER"},
		{parser.TokenNumber, "NUMBER"},
		{parser.TokenRegister, "REGISTER"},
		{parser.TokenComma, ","},
		{parser.TokenColon, ":"},
		{parser.TokenHash, "#"},
		{parser.TokenLBracket, "["},
		{parser.TokenRBracket, "]"},
		{parser.TokenNewline, "NEWLINE"},
		{parser.TokenComment, "COMMENT"},
		{parser.TokenDirective, "DIRECTIVE"},
		{parser.TokenCondition, "CONDITION"},
	}

	for _, tt := range tests {
		result := tt.token.String()
		if result != tt.expected {
			t.Errorf("Expected %v.String()='%s', got '%s'", tt.token, tt.expected, result)
		}
	}
}

func TestTokenTypeStringUnknown(t *testing.T) {
	// Test with an undefined token type
	unknownToken := parser.TokenType(999)
	result := unknownToken.String()

	// Should return a formatted string like "TokenType(999)"
	expected := "TokenType(999)"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestTokenString(t *testing.T) {
	// Test Token.String() method
	token := parser.Token{
		Type:    parser.TokenIdentifier,
		Literal: "MOV",
		Pos: parser.Position{
			Filename: "test.s",
			Line:     1,
			Column:   1,
		},
	}

	result := token.String()

	// Should contain the token type, literal, and position
	expectedSubstrings := []string{
		"IDENTIFIER",
		"MOV",
		"test.s:1:1",
	}

	for _, substr := range expectedSubstrings {
		if !contains(result, substr) {
			t.Errorf("Expected token string to contain '%s', got: %s", substr, result)
		}
	}
}

func TestTokenStringWithNumber(t *testing.T) {
	token := parser.Token{
		Type:    parser.TokenNumber,
		Literal: "42",
		Pos: parser.Position{
			Filename: "test.s",
			Line:     5,
			Column:   10,
		},
	}

	result := token.String()

	expectedSubstrings := []string{
		"NUMBER",
		"42",
		"test.s:5:10",
	}

	for _, substr := range expectedSubstrings {
		if !contains(result, substr) {
			t.Errorf("Expected token string to contain '%s', got: %s", substr, result)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
