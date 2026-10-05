package parser_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

func TestEquExpressions(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want uint32
	}{
		{"literal", ".equ X, 42\n", 42},
		{"negative literal", ".equ X, -4\n", 0xFFFFFFFC},
		{"symbol", ".equ A, 4\n.equ X, A\n", 4},
		{"symbol plus", ".equ A, 4\n.equ X, A+1\n", 5},
		{"spaces", ".equ A, 4\n.equ X, A + 1\n", 5},
		{"precedence", ".equ A, 4\n.equ X, A*2+1\n", 9},
		{"parentheses", ".equ A, 4\n.equ X, (A+1)*2\n", 10},
		{"left-associative minus", ".equ X, 10-3-2\n", 5},
		{"division", ".equ X, 17/5\n", 3},
		{"modulo", ".equ X, 17%5\n", 2},
		{"shift", ".equ X, 1<<5\n", 32},
		{"shift right", ".equ X, 0x80>>3\n", 16},
		{"mask", ".equ X, (1<<4)|(1<<1)\n", 0x12},
		{"and", ".equ X, 0xFF&0x0F\n", 0x0F},
		{"xor", ".equ X, 0xF0^0xFF\n", 0x0F},
		{"unary minus symbol", ".equ A, 4\n.equ X, -A\n", 0xFFFFFFFC},
		{"bitwise not", ".equ X, ~0\n", 0xFFFFFFFF},
		{"hex and binary", ".equ X, 0x10+0b11\n", 19},
		{"label", ".org 0x8000\nstart:\n NOP\n.equ X, start+4\n", 0x8004},
		{"set alias", ".set X, 3*3\n", 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.NewParser(tt.src, "test.s").Parse()
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			got, err := program.SymbolTable.Get("X")
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("X = 0x%X, want 0x%X", got, tt.want)
			}
		})
	}
}

func TestEquExpressionErrors(t *testing.T) {
	tests := []struct{ name, src string }{
		{"undefined symbol", ".equ X, NOPE+1\n"},
		{"division by zero", ".equ X, 4/0\n"},
		{"modulo by zero", ".equ X, 4%0\n"},
		{"dangling operator", ".equ X, 4+\n"},
		{"unbalanced parenthesis", ".equ X, (4+1\n"},
		{"extra parenthesis", ".equ X, 4+1)\n"},
		{"missing value", ".equ X\n"},
		{"two values", ".equ X, 4 5\n"},
		{"shift too far", ".equ X, 1<<32\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parser.NewParser(tt.src, "test.s").Parse(); err == nil {
				t.Errorf("Parse(%q) succeeded, want error", tt.src)
			}
		})
	}
}
