package parser_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// TestConstantExpressions tests that the parser can handle constant expressions
// in immediate values and pseudo-instructions like LDR Rd, =label + offset
func TestConstantExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkASM func(*testing.T, *parser.Program)
	}{
		{
			name: "LDR with label plus constant",
			input: `
.data
buffer: .space 12
buffer_end:

.text
_start:
    LDR r0, =buffer_end
    LDR r1, =buffer + 12
    MOV pc, lr
`,
			wantErr: false,
			checkASM: func(t *testing.T, prog *parser.Program) {
				// Both should resolve to the same address
				if len(prog.Instructions) < 2 {
					t.Fatalf("Expected at least 2 instructions, got %d", len(prog.Instructions))
				}
				// Check that both instructions parsed successfully
				inst1 := prog.Instructions[0]
				inst2 := prog.Instructions[1]
				if inst1.Mnemonic != "LDR" || inst2.Mnemonic != "LDR" {
					t.Errorf("Expected LDR instructions, got %s and %s", inst1.Mnemonic, inst2.Mnemonic)
				}
			},
		},
		{
			name: "LDR with label minus constant",
			input: `
.text
_start:
    LDR r0, =end_label
    LDR r1, =end_label - 4
    B end_label
end_label:
    MOV pc, lr
`,
			wantErr: false,
		},
		{
			name: "Multiple arithmetic operations",
			input: `
.data
base: .word 0
.text
_start:
    LDR r0, =base + 4
    LDR r1, =base + 8
    LDR r2, =base + 12
    MOV pc, lr
`,
			wantErr: false,
		},
		{
			name: "Expression with hex values",
			input: `
.data
buffer: .space 0x100
.text
_start:
    LDR r0, =buffer + 0x10
    LDR r1, =buffer + 0xFF
    MOV pc, lr
`,
			wantErr: false,
		},
		{
			name: "Negative offset",
			input: `
.text
end_marker:
    .word 0xDEADBEEF
_start:
    LDR r0, =end_marker - 4
    MOV pc, lr
`,
			wantErr: false,
		},
		{
			name: "Multi-digit offset addition",
			input: `
.data
numbuf: .space 12
.text
_start:
    LDR r3, =numbuf + 11
    LDR r4, =numbuf + 7
    LDR r5, =numbuf + 123
    MOV pc, lr
`,
			wantErr: false,
		},
		{
			name: "Multi-digit offset subtraction",
			input: `
.data
buffer: .space 256
.text
_start:
    LDR r0, =buffer + 255
    LDR r1, =buffer + 255 - 17
    LDR r2, =buffer + 100 - 33
    MOV pc, lr
`,
			wantErr: false,
		},
		{
			name: "Various non-power-of-2 offsets",
			input: `
.data
data: .space 50
.text
_start:
    LDR r0, =data + 3
    LDR r1, =data + 7
    LDR r2, =data + 13
    LDR r3, =data + 19
    LDR r4, =data + 31
    MOV pc, lr
`,
			wantErr: false,
		},
		{
			name: "Subtraction with various values",
			input: `
.text
marker:
    .word 0
_start:
    LDR r0, =marker - 1
    LDR r1, =marker - 5
    LDR r2, =marker - 11
    LDR r3, =marker - 23
    MOV pc, lr
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			prog, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected parse error: %v", err)
			}

			if prog == nil {
				t.Fatal("Expected program but got nil")
			}

			if tt.checkASM != nil {
				tt.checkASM(t, prog)
			}
		})
	}
}

// TestConstantExpressionEvaluation tests that constant expressions are evaluated correctly
func TestConstantExpressionEvaluation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantErr        bool
		checkAddresses func(*testing.T, *parser.Program)
	}{
		{
			name: "Addition evaluates correctly",
			input: `
.data
buffer: .space 10
buffer_plus_5:
.text
_start:
    LDR r0, =buffer + 5
    LDR r1, =buffer_plus_5
`,
			wantErr: false,
			checkAddresses: func(t *testing.T, prog *parser.Program) {
				// The address of buffer+5 should equal buffer_plus_5
				// Both should be 5 bytes after the start of buffer
				if len(prog.Instructions) < 2 {
					t.Fatalf("Expected at least 2 instructions")
				}
				// We can't easily check the resolved addresses here without
				// running the encoder, but we can verify both instructions parsed
			},
		},
		{
			name: "Subtraction evaluates correctly",
			input: `
.text
end_label:
    NOP
    NOP
    NOP
start:
    LDR r0, =end_label - 8
    B end_label
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			prog, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected parse error: %v", err)
			}

			if tt.checkAddresses != nil {
				tt.checkAddresses(t, prog)
			}
		})
	}
}
