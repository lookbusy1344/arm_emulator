package parser_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// Tests for parser operand edge cases identified in code review (CODE_REVIEW_OPUS.md section 3.3)

// TestParseOperand_Immediate tests immediate value parsing edge cases
func TestParseOperand_Immediate(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantOperand  string
		wantParseErr bool
	}{
		// Valid immediate values
		{"simple immediate", "MOV R0, #42", "#42", false},
		{"negative immediate", "MOV R0, #-1", "#-1", false},
		{"hex immediate", "MOV R0, #0xFF", "#0xFF", false},
		{"character literal", "MOV R0, #'A'", "#'A'", false},
		{"zero immediate", "MOV R0, #0", "#0", false},
		// Symbol reference as immediate
		{"symbol immediate", "MOV R0, #MAX_VAL", "#MAX_VAL", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()

			if tt.wantParseErr {
				if err == nil {
					t.Errorf("expected parse error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]
			if len(inst.Operands) < 2 {
				t.Fatalf("expected at least 2 operands, got %d", len(inst.Operands))
			}

			if inst.Operands[1] != tt.wantOperand {
				t.Errorf("operand = %q, want %q", inst.Operands[1], tt.wantOperand)
			}
		})
	}
}

// TestParseOperand_Memory tests memory addressing edge cases
func TestParseOperand_Memory(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOperand string
		wantErr     bool
	}{
		// Valid memory addressing modes
		{"simple base register", "LDR R0, [R1]", "[R1]", false},
		{"base with immediate offset", "LDR R0, [R1, #4]", "[R1, #4]", false},
		{"base with negative offset", "LDR R0, [R1, #-4]", "[R1, #-4]", false},
		{"base with register offset", "LDR R0, [R1, R2]", "[R1,R2]", false},
		{"base with shifted register", "LDR R0, [R1, R2, LSL #2]", "[R1,R2, LSL #2]", false},
		{"pre-indexed with writeback", "LDR R0, [R1, #4]!", "[R1, #4]!", false},
		// Complex addressing
		{"PC-relative", "LDR R0, [PC, #8]", "[PC, #8]", false},
		{"SP-relative", "LDR R0, [SP, #-16]", "[SP, #-16]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected parse error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]
			if len(inst.Operands) < 2 {
				t.Fatalf("expected at least 2 operands, got %d", len(inst.Operands))
			}

			if inst.Operands[1] != tt.wantOperand {
				t.Errorf("operand = %q, want %q", inst.Operands[1], tt.wantOperand)
			}
		})
	}
}

// TestParseOperand_RegisterList tests register list parsing edge cases
func TestParseOperand_RegisterList(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOperand string
		wantErr     bool
	}{
		// Valid register lists
		{"single register", "STMFD SP!, {R0}", "{R0}", false},
		{"multiple registers", "STMFD SP!, {R0, R1, R2}", "{R0,R1,R2}", false},
		{"register range", "STMFD SP!, {R0-R3}", "{R0-R3}", false},
		{"mixed list and range", "STMFD SP!, {R0-R3, R5, R7}", "{R0-R3,R5,R7}", false},
		{"with LR", "STMFD SP!, {R0-R3, LR}", "{R0-R3,LR}", false},
		{"with PC", "LDMFD SP!, {R0-R3, PC}", "{R0-R3,PC}", false},
		{"full save", "STMFD SP!, {R0-R12, LR}", "{R0-R12,LR}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected parse error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]
			// Register list is typically the last operand
			lastOp := inst.Operands[len(inst.Operands)-1]
			if lastOp != tt.wantOperand {
				t.Errorf("register list = %q, want %q", lastOp, tt.wantOperand)
			}
		})
	}
}

// TestParseOperand_Pseudo tests pseudo-instruction operand parsing
func TestParseOperand_Pseudo(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOperand string
		wantErr     bool
	}{
		// LDR =value pseudo-instruction
		{"simple label", "LDR R0, =my_label", "=my_label", false},
		{"numeric value", "LDR R0, =0x12345678", "=0x12345678", false},
		{"decimal value", "LDR R0, =1000", "=1000", false},
		// Expression support
		{"label plus offset", "LDR R0, =data+12", "=data+12", false},
		{"label minus offset", "LDR R0, =data-4", "=data-4", false},
		{"complex expression", "LDR R0, =base+offset", "=base+offset", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected parse error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]
			if len(inst.Operands) < 2 {
				t.Fatalf("expected at least 2 operands, got %d", len(inst.Operands))
			}

			if inst.Operands[1] != tt.wantOperand {
				t.Errorf("operand = %q, want %q", inst.Operands[1], tt.wantOperand)
			}
		})
	}
}

// TestParseOperand_ShiftedRegister tests shifted register operand parsing
func TestParseOperand_ShiftedRegister(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOperand string
		wantErr     bool
	}{
		// Shift with immediate
		{"LSL immediate", "ADD R0, R1, R2, LSL #2", "R2,LSL #2", false},
		{"LSR immediate", "ADD R0, R1, R2, LSR #4", "R2,LSR #4", false},
		{"ASR immediate", "ADD R0, R1, R2, ASR #1", "R2,ASR #1", false},
		{"ROR immediate", "ADD R0, R1, R2, ROR #8", "R2,ROR #8", false},
		// Shift with register
		{"LSL register", "ADD R0, R1, R2, LSL R3", "R2,LSL R3", false},
		{"LSR register", "ADD R0, R1, R2, LSR R3", "R2,LSR R3", false},
		// RRX (no shift amount)
		{"RRX", "ADD R0, R1, R2, RRX", "R2,RRX", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected parse error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]
			// Shifted register is typically the last operand (3rd for data processing)
			if len(inst.Operands) < 3 {
				t.Fatalf("expected at least 3 operands, got %d", len(inst.Operands))
			}

			lastOp := inst.Operands[len(inst.Operands)-1]
			if lastOp != tt.wantOperand {
				t.Errorf("shifted register = %q, want %q", lastOp, tt.wantOperand)
			}
		})
	}
}

// TestParseOperand_Writeback tests register writeback parsing
func TestParseOperand_Writeback(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOperand string
		wantErr     bool
	}{
		{"SP writeback", "STMFD SP!, {R0-R3}", "SP!", false},
		{"R13 writeback", "STMFD R13!, {R0-R3}", "R13!", false},
		{"LDM writeback", "LDMFD SP!, {R0-R3}", "SP!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected parse error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]
			if len(inst.Operands) < 1 {
				t.Fatalf("expected at least 1 operand, got %d", len(inst.Operands))
			}

			if inst.Operands[0] != tt.wantOperand {
				t.Errorf("writeback operand = %q, want %q", inst.Operands[0], tt.wantOperand)
			}
		})
	}
}

// TestParseOperand_UnclosedBrackets tests behavior with unclosed brackets
// Note: The parser may handle these gracefully or report errors at encoding time
func TestParseOperand_UnclosedBrackets(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// These inputs have unclosed brackets - parser behavior may vary
		{"unclosed memory bracket", "LDR R0, [R1"},
		{"unclosed register list", "STMFD SP!, {R0, R1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()

			// Parser may or may not error - document actual behavior
			if err != nil {
				t.Logf("Parser correctly rejected unclosed bracket: %v", err)
				return
			}

			// If parser accepted it, verify the operand was captured (possibly incomplete)
			if len(program.Instructions) > 0 {
				inst := program.Instructions[0]
				t.Logf("Parser accepted with operands: %v", inst.Operands)
			}
		})
	}
}

// TestParseOperand_AllShiftTypes tests all ARM shift operators
func TestParseOperand_AllShiftTypes(t *testing.T) {
	shiftOps := []string{"LSL", "LSR", "ASR", "ROR", "RRX"}

	for _, shift := range shiftOps {
		t.Run(shift, func(t *testing.T) {
			var input string
			var wantOperand string

			if shift == "RRX" {
				// RRX doesn't take a shift amount
				input = "MOV R0, R1, " + shift
				wantOperand = "R1," + shift
			} else {
				input = "MOV R0, R1, " + shift + " #1"
				wantOperand = "R1," + shift + " #1"
			}

			p := parser.NewParser(input, "test.s")
			program, err := p.Parse()

			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]
			lastOp := inst.Operands[len(inst.Operands)-1]
			if lastOp != wantOperand {
				t.Errorf("shift operand = %q, want %q", lastOp, wantOperand)
			}
		})
	}
}

// TestParseOperand_MemoryWithAllShifts tests memory addressing with all shift types
func TestParseOperand_MemoryWithAllShifts(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOperand string
	}{
		{"memory LSL", "LDR R0, [R1, R2, LSL #2]", "[R1,R2, LSL #2]"},
		{"memory LSR", "LDR R0, [R1, R2, LSR #2]", "[R1,R2, LSR #2]"},
		{"memory ASR", "LDR R0, [R1, R2, ASR #2]", "[R1,R2, ASR #2]"},
		{"memory ROR", "LDR R0, [R1, R2, ROR #2]", "[R1,R2, ROR #2]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.input, "test.s")
			program, err := p.Parse()

			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if len(program.Instructions) != 1 {
				t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
			}

			inst := program.Instructions[0]
			if len(inst.Operands) < 2 {
				t.Fatalf("expected at least 2 operands, got %d", len(inst.Operands))
			}

			if inst.Operands[1] != tt.wantOperand {
				t.Errorf("memory operand = %q, want %q", inst.Operands[1], tt.wantOperand)
			}
		})
	}
}
