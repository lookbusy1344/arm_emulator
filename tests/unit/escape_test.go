package unit

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

func TestProcessEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic escapes
		{"newline", "hello\\nworld", "hello\nworld"},
		{"tab", "hello\\tworld", "hello\tworld"},
		{"carriage return", "hello\\rworld", "hello\rworld"},
		{"backslash", "hello\\\\world", "hello\\world"},
		{"null", "hello\\0world", "hello\x00world"},
		{"double quote", "hello\\\"world", "hello\"world"},
		{"single quote", "hello\\'world", "hello'world"},
		{"alert", "hello\\aworld", "hello\aworld"},
		{"backspace", "hello\\bworld", "hello\bworld"},
		{"form feed", "hello\\fworld", "hello\fworld"},
		{"vertical tab", "hello\\vworld", "hello\vworld"},

		// Hex escapes
		{"hex 0x00", "hello\\x00world", "hello\x00world"},
		{"hex 0x41 (A)", "hello\\x41world", "helloAworld"},
		{"hex 0xFF", "\\xFF", "\xFF"},
		{"hex lowercase", "\\x0a", "\n"},
		{"hex uppercase", "\\x0A", "\n"},
		{"hex mixed case", "\\xAb", "\xAB"},

		// Multiple escapes
		{"multiple", "\\n\\t\\r", "\n\t\r"},
		{"mixed", "hello\\nworld\\t!", "hello\nworld\t!"},

		// Unknown escapes preserved
		{"unknown z", "hello\\zworld", "hello\\zworld"},
		{"unknown q", "\\q", "\\q"},

		// Edge cases
		{"empty string", "", ""},
		{"no escapes", "hello world", "hello world"},
		{"trailing backslash", "hello\\", "hello\\"},
		{"just backslash n", "\\n", "\n"},

		// Real-world examples
		{"C string", "Hello, World!\\n", "Hello, World!\n"},
		{"prompt", "Enter value: \\t", "Enter value: \t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ProcessEscapeSequences(tt.input)
			if result != tt.expected {
				t.Errorf("ProcessEscapeSequences(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseEscapeChar(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    byte
		consumed    int
		expectError bool
	}{
		// Basic escapes
		{"newline", "\\n", '\n', 2, false},
		{"tab", "\\t", '\t', 2, false},
		{"carriage return", "\\r", '\r', 2, false},
		{"backslash", "\\\\", '\\', 2, false},
		{"null", "\\0", 0, 2, false},
		{"double quote", "\\\"", '"', 2, false},
		{"single quote", "\\'", '\'', 2, false},
		{"alert", "\\a", '\a', 2, false},
		{"backspace", "\\b", '\b', 2, false},
		{"form feed", "\\f", '\f', 2, false},
		{"vertical tab", "\\v", '\v', 2, false},

		// Hex escapes
		{"hex 00", "\\x00", 0, 4, false},
		{"hex 41", "\\x41", 'A', 4, false},
		{"hex FF", "\\xFF", 0xFF, 4, false},
		{"hex 0a lowercase", "\\x0a", '\n', 4, false},

		// With trailing content
		{"newline with extra", "\\nrest", '\n', 2, false},
		{"hex with extra", "\\x41rest", 'A', 4, false},

		// Error cases
		{"empty", "", 0, 0, true},
		{"no backslash", "n", 0, 0, true},
		{"unknown escape", "\\z", 0, 0, true},
		{"incomplete hex", "\\x0", 0, 0, true},
		{"invalid hex", "\\xGG", 0, 0, true},
		{"just backslash", "\\", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, consumed, err := parser.ParseEscapeChar(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseEscapeChar(%q) expected error but got none", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseEscapeChar(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseEscapeChar(%q) = %d, want %d", tt.input, result, tt.expected)
			}
			if consumed != tt.consumed {
				t.Errorf("ParseEscapeChar(%q) consumed %d, want %d", tt.input, consumed, tt.consumed)
			}
		})
	}
}

func TestHexEscapeEdgeCases(t *testing.T) {
	// Test that hex escapes require exactly 2 digits
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Hex needs exactly 2 chars after \x
		{"hex followed by non-hex", "\\xGH", "\\xGH"},
		{"single hex digit only", "\\x0", "\\x0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ProcessEscapeSequences(tt.input)
			if result != tt.expected {
				t.Errorf("ProcessEscapeSequences(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
