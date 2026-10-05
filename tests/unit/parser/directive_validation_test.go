package parser_test

import (
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

func TestInvalidSizeAndAlignmentDirectivesFail(t *testing.T) {
	tests := []struct {
		name, src, wantErr string
	}{
		{"negative space", "buf: .space -4\n", ".space"},
		{"negative skip", "buf: .skip -1\n", ".skip"},
		{"negative space via constant", ".equ N, -4\nbuf: .space N\n", ".space"},
		{"space past end of address space", ".org 0xFFFFFFF0\nbuf: .space 0x20\n", ".space"},
		{"space to end of address space", ".org 0xFFFFFFF0\nbuf: .space 0x10\n", ".space"},
		{"balign zero", ".balign 0\n", ".balign"},
		{"balign not a power of two", ".balign 3\n", ".balign"},
		{"balign not a number", ".balign x\n", ".balign"},
		{"align too large", ".align 32\n", ".align"},
		{"align not a number", ".align x\n", ".align"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.NewParser(tt.src, "test.s").Parse()
			if err == nil {
				t.Fatalf("Parse(%q) succeeded, want error", tt.src)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not mention %s", err, tt.wantErr)
			}
		})
	}
}

func TestValidSizeAndAlignmentDirectives(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		wantNext uint32 // address of label "next"
	}{
		{"space zero", ".org 0x100\n .space 0\nnext:\n", 0x100},
		{"space via constant", ".equ N, 12\n.org 0x100\n .space N\nnext:\n", 0x10C},
		{"align 0 is no-op", ".org 0x101\n .align 0\nnext:\n", 0x101},
		{"align 2", ".org 0x101\n .align 2\nnext:\n", 0x104},
		{"align on boundary", ".org 0x100\n .align 4\nnext:\n", 0x100},
		{"balign 1", ".org 0x101\n .balign 1\nnext:\n", 0x101},
		{"balign 8", ".org 0x101\n .balign 8\nnext:\n", 0x108},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.NewParser(tt.src, "test.s").Parse()
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			next, ok := program.SymbolTable.Lookup("next")
			if !ok {
				t.Fatal("label next not defined")
			}
			if next.Value != tt.wantNext {
				t.Errorf("next = 0x%X, want 0x%X", next.Value, tt.wantNext)
			}
		})
	}
}
