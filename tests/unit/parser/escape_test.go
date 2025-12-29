package parser_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// TestProcessEscapeSequences_Octal tests octal escape sequence parsing
func TestProcessEscapeSequences_Octal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"octal 101 (A)", "\\101", "A"},
		{"octal 132 (Z)", "\\132", "Z"},
		{"octal 000 (null)", "\\000", "\x00"},
		{"octal 012 (newline)", "\\012", "\n"},
		{"octal 011 (tab)", "\\011", "\t"},
		{"octal 040 (space)", "\\040", " "},
		{"octal 377 (255)", "\\377", "\xff"},
		{"octal in string", "Hello\\012World", "Hello\nWorld"},
		{"mixed escapes", "\\101\\n\\102", "A\nB"},
		{"octal with regular text", "X\\101Y", "XAY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ProcessEscapeSequences(tt.input)
			if result != tt.expected {
				t.Errorf("ProcessEscapeSequences(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseEscapeChar_Octal tests single octal escape character parsing
func TestParseEscapeChar_Octal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected byte
		consumed int
		wantErr  bool
	}{
		{"octal 101 (A)", "\\101", 'A', 4, false},
		{"octal 132 (Z)", "\\132", 'Z', 4, false},
		{"octal 000 (null)", "\\000", 0, 4, false},
		{"octal 012 (newline)", "\\012", '\n', 4, false},
		{"octal 377 (255)", "\\377", 0xff, 4, false},
		{"octal 1 digit", "\\1", 1, 2, false},
		{"octal 2 digits", "\\12", '\n', 3, false},
		{"octal overflow 400", "\\400", 0, 0, true}, // > 255
		{"octal with trailing", "\\101ABC", 'A', 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, consumed, err := parser.ParseEscapeChar(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseEscapeChar(%q) expected error, got none", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseEscapeChar(%q) unexpected error: %v", tt.input, err)
				return
			}
			if b != tt.expected {
				t.Errorf("ParseEscapeChar(%q) byte = %d, want %d", tt.input, b, tt.expected)
			}
			if consumed != tt.consumed {
				t.Errorf("ParseEscapeChar(%q) consumed = %d, want %d", tt.input, consumed, tt.consumed)
			}
		})
	}
}

// TestProcessEscapeSequences_Hex tests hex escape sequences are still working
func TestProcessEscapeSequences_Hex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"hex 41 (A)", "\\x41", "A"},
		{"hex 5A (Z)", "\\x5A", "Z"},
		{"hex 00 (null)", "\\x00", "\x00"},
		{"hex 0A (newline)", "\\x0A", "\n"},
		{"hex FF (255)", "\\xFF", "\xff"},
		{"hex lowercase", "\\xff", "\xff"},
		{"hex in string", "Hello\\x0AWorld", "Hello\nWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ProcessEscapeSequences(tt.input)
			if result != tt.expected {
				t.Errorf("ProcessEscapeSequences(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestProcessEscapeSequences_Standard tests standard escapes still work
func TestProcessEscapeSequences_Standard(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"newline", "\\n", "\n"},
		{"tab", "\\t", "\t"},
		{"carriage return", "\\r", "\r"},
		{"backslash", "\\\\", "\\"},
		{"null", "\\0", "\x00"},
		{"double quote", "\\\"", "\""},
		{"single quote", "\\'", "'"},
		{"alert", "\\a", "\a"},
		{"backspace", "\\b", "\b"},
		{"form feed", "\\f", "\f"},
		{"vertical tab", "\\v", "\v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ProcessEscapeSequences(tt.input)
			if result != tt.expected {
				t.Errorf("ProcessEscapeSequences(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestOctalEdgeCases tests edge cases for octal parsing
func TestOctalEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Octal should only consume up to 3 digits
		{"octal stops at 3 digits", "\\1234", "\x534"}, // \123 = 83 = 'S', then '4'
		// Octal should stop at non-octal digit
		{"octal stops at 8", "\\128", "\n8"},   // \12 = 10 = '\n', then '8'
		{"octal stops at 9", "\\019", "\x019"}, // \01 = 1, then '9'
		// Single digit octal
		{"single digit octal", "\\7", "\x07"},
		// Two digit octal
		{"two digit octal", "\\77", "?"},
		// Three digit octal
		{"three digit octal", "\\101", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ProcessEscapeSequences(tt.input)
			if result != tt.expected {
				t.Errorf("ProcessEscapeSequences(%q) = %q (bytes: %v), want %q (bytes: %v)",
					tt.input, result, []byte(result), tt.expected, []byte(tt.expected))
			}
		})
	}
}
