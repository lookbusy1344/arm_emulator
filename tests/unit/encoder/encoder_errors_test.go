package encoder_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/encoder"
	"github.com/lookbusy1344/arm-emulator/parser"
)

// TestEncodingError_ErrorFormat tests the error message formatting
func TestEncodingError_ErrorFormat(t *testing.T) {
	tests := []struct {
		name     string
		inst     *parser.Instruction
		message  string
		wrapped  error
		wantSubs []string // substrings that must appear in error
	}{
		{
			name: "with full position info",
			inst: &parser.Instruction{
				Mnemonic: "BADOP",
				Operands: []string{"R0", "R1"},
				Pos: parser.Position{
					Filename: "test.s",
					Line:     42,
					Column:   5,
				},
				RawLine: "    BADOP R0, R1",
			},
			message:  "unknown instruction: BADOP",
			wantSubs: []string{"test.s:42:5:", "unknown instruction: BADOP", "source:", "BADOP R0, R1"},
		},
		{
			name: "with line number only",
			inst: &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{"R16", "#0"},
				Pos: parser.Position{
					Line: 10,
				},
				RawLine: "MOV R16, #0",
			},
			message:  "invalid register: R16",
			wantSubs: []string{"line 10:", "invalid register: R16", "source:", "MOV R16, #0"},
		},
		{
			name: "with wrapped error",
			inst: &parser.Instruction{
				Mnemonic: "LDR",
				Operands: []string{"R0", "=invalid"},
				Pos: parser.Position{
					Filename: "program.s",
					Line:     99,
					Column:   1,
				},
			},
			message:  "failed to encode instruction",
			wrapped:  errors.New("undefined symbol: invalid"),
			wantSubs: []string{"program.s:99:1:", "failed to encode instruction", "undefined symbol: invalid"},
		},
		{
			name:     "nil instruction",
			inst:     nil,
			message:  "encoding failed",
			wantSubs: []string{"encoding error:", "encoding failed"},
		},
		{
			name:     "nil instruction with wrapped error",
			inst:     nil,
			message:  "internal error",
			wrapped:  errors.New("something went wrong"),
			wantSubs: []string{"encoding error:", "internal error", "something went wrong"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var encErr *encoder.EncodingError
			if tt.wrapped != nil {
				encErr = &encoder.EncodingError{
					Instruction: tt.inst,
					Message:     tt.message,
					Wrapped:     tt.wrapped,
				}
			} else {
				encErr = encoder.NewEncodingError(tt.inst, tt.message)
			}

			errMsg := encErr.Error()

			for _, sub := range tt.wantSubs {
				if !strings.Contains(errMsg, sub) {
					t.Errorf("error message missing %q\ngot: %s", sub, errMsg)
				}
			}
		})
	}
}

// TestEncodingError_Unwrap tests error unwrapping
func TestEncodingError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	encErr := &encoder.EncodingError{
		Message: "wrapper",
		Wrapped: originalErr,
	}

	if encErr.Unwrap() != originalErr {
		t.Error("Unwrap() should return the wrapped error")
	}

	// Test with errors.Is
	if !errors.Is(encErr, originalErr) {
		t.Error("errors.Is should find the wrapped error")
	}
}

// TestWrapEncodingError tests the WrapEncodingError helper
func TestWrapEncodingError(t *testing.T) {
	inst := &parser.Instruction{
		Mnemonic: "MOV",
		Pos:      parser.Position{Line: 5},
	}

	t.Run("wraps regular error", func(t *testing.T) {
		originalErr := errors.New("some error")
		wrapped := encoder.WrapEncodingError(inst, originalErr)

		encErr, ok := wrapped.(*encoder.EncodingError)
		if !ok {
			t.Fatal("expected *EncodingError")
		}

		if encErr.Instruction != inst {
			t.Error("instruction should be preserved")
		}

		if !errors.Is(wrapped, originalErr) {
			t.Error("should wrap the original error")
		}
	})

	t.Run("does not double-wrap EncodingError", func(t *testing.T) {
		original := encoder.NewEncodingError(inst, "first error")
		wrapped := encoder.WrapEncodingError(inst, original)

		if wrapped != original {
			t.Error("should not double-wrap EncodingError")
		}
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrapped := encoder.WrapEncodingError(inst, nil)
		if wrapped != nil {
			t.Error("should return nil for nil error")
		}
	})
}

// TestEncodeInstruction_ErrorContext verifies that EncodeInstruction returns errors with context
func TestEncodeInstruction_ErrorContext(t *testing.T) {
	enc := encoder.NewEncoder(parser.NewSymbolTable())

	tests := []struct {
		name      string
		inst      *parser.Instruction
		wantSubs  []string
		wantNoErr bool
	}{
		{
			name: "unknown instruction includes context",
			inst: &parser.Instruction{
				Mnemonic: "INVALID",
				Operands: []string{"R0", "R1"},
				Pos: parser.Position{
					Filename: "test.s",
					Line:     15,
					Column:   1,
				},
				RawLine: "INVALID R0, R1",
			},
			wantSubs: []string{"test.s:15:1:", "unknown instruction", "INVALID"},
		},
		{
			name: "missing operands includes context",
			inst: &parser.Instruction{
				Mnemonic: "ADD",
				Operands: []string{"R0", "R1"}, // Missing third operand
				Pos: parser.Position{
					Filename: "program.s",
					Line:     20,
					Column:   4,
				},
				RawLine: "    ADD R0, R1",
			},
			wantSubs: []string{"program.s:20:4:", "requires 3 operands"},
		},
		{
			name: "invalid register includes context",
			inst: &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{"R20", "#0"}, // Invalid register
				Pos: parser.Position{
					Line: 5,
				},
			},
			wantSubs: []string{"line 5:", "invalid register"},
		},
		{
			name: "valid instruction returns no error",
			inst: &parser.Instruction{
				Mnemonic: "MOV",
				Operands: []string{"R0", "#0"},
				Pos: parser.Position{
					Filename: "good.s",
					Line:     1,
				},
			},
			wantNoErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := enc.EncodeInstruction(tt.inst, 0x8000)

			if tt.wantNoErr {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			errMsg := err.Error()
			for _, sub := range tt.wantSubs {
				if !strings.Contains(errMsg, sub) {
					t.Errorf("error message missing %q\ngot: %s", sub, errMsg)
				}
			}
		})
	}
}

// TestEncodingError_UsableWithErrorsAs tests that EncodingError works with errors.As
func TestEncodingError_UsableWithErrorsAs(t *testing.T) {
	enc := encoder.NewEncoder(parser.NewSymbolTable())

	inst := &parser.Instruction{
		Mnemonic: "BADOP",
		Operands: []string{},
		Pos: parser.Position{
			Filename: "test.s",
			Line:     10,
		},
		RawLine: "BADOP",
	}

	_, err := enc.EncodeInstruction(inst, 0x8000)
	if err == nil {
		t.Fatal("expected error")
	}

	var encErr *encoder.EncodingError
	if !errors.As(err, &encErr) {
		t.Fatal("error should be extractable as *EncodingError")
	}

	if encErr.Instruction != inst {
		t.Error("instruction should be preserved in extracted error")
	}

	if encErr.Instruction.Pos.Line != 10 {
		t.Error("position should be preserved")
	}
}
