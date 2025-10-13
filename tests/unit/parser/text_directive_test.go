package parser_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// TestTextDirective tests that .text directive properly sets the origin to 0
func TestTextDirective(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantErr         bool
		expectOrigin    uint32
		expectOriginSet bool
	}{
		{
			name: ".text sets origin to 0",
			input: `.text
.global _start
_start:
    MOV r0, #42
    MOV pc, lr
`,
			wantErr:         false,
			expectOrigin:    0,
			expectOriginSet: true,
		},
		{
			name: ".text after .org preserves .org",
			input: `.org 0x8000
.text
_start:
    MOV r0, #42
    MOV pc, lr
`,
			wantErr:         false,
			expectOrigin:    0x8000,
			expectOriginSet: true,
		},
		{
			name: ".text before .org, .org takes precedence",
			input: `.text
.org 0x1000
_start:
    MOV r0, #42
    MOV pc, lr
`,
			wantErr:         false,
			expectOrigin:    0, // .text sets it first
			expectOriginSet: true,
		},
		{
			name: ".global without .text",
			input: `.global _start
_start:
    MOV r0, #42
    MOV pc, lr
`,
			wantErr:         false,
			expectOrigin:    0,     // Default origin
			expectOriginSet: false, // No explicit origin directive
		},
		{
			name: ".data directive",
			input: `.text
_start:
    MOV r0, #42
.data
msg: .asciz "Hello"
`,
			wantErr:         false,
			expectOrigin:    0,
			expectOriginSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test")
			prog, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if prog.Origin != tt.expectOrigin {
				t.Errorf("Expected origin 0x%X, got 0x%X", tt.expectOrigin, prog.Origin)
			}

			if prog.OriginSet != tt.expectOriginSet {
				t.Errorf("Expected OriginSet %v, got %v", tt.expectOriginSet, prog.OriginSet)
			}
		})
	}
}
