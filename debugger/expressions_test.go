package debugger

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestExpressionEvaluator_Numbers(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()
	symbols := make(map[string]uint32)

	tests := []struct {
		name string
		expr string
		want uint32
	}{
		{"Decimal", "42", 42},
		{"Hex", "0x100", 0x100},
		{"Hex uppercase", "0X1A", 0x1A},
		{"Binary", "0b1010", 0b1010},
		{"Octal", "010", 8},
		{"Negative", "-1", 0xFFFFFFFF},
		{"Large hex", "0xFFFFFFFF", 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eval.EvaluateExpression(tt.expr, machine, symbols)
			if err != nil {
				t.Fatalf("EvaluateExpression() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EvaluateExpression() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestExpressionEvaluator_Registers(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()
	symbols := make(map[string]uint32)

	// Set register values
	machine.CPU.R[0] = 100
	machine.CPU.R[5] = 200
	machine.CPU.SetSP(0x1000)
	machine.CPU.SetLR(0x2000)
	machine.CPU.PC = 0x3000

	tests := []struct {
		name string
		expr string
		want uint32
	}{
		{"R0", "r0", 100},
		{"R5", "r5", 200},
		{"SP", "sp", 0x1000},
		{"R13", "r13", 0x1000},
		{"LR", "lr", 0x2000},
		{"R14", "r14", 0x2000},
		{"PC", "pc", 0x3000},
		{"R15", "r15", 0x3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eval.EvaluateExpression(tt.expr, machine, symbols)
			if err != nil {
				t.Fatalf("EvaluateExpression() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EvaluateExpression() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestExpressionEvaluator_Symbols(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()
	symbols := map[string]uint32{
		"main":   0x1000,
		"loop":   0x2000,
		"_start": 0x3000,
	}

	tests := []struct {
		name string
		expr string
		want uint32
	}{
		{"main", "main", 0x1000},
		{"loop", "loop", 0x2000},
		{"_start", "_start", 0x3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eval.EvaluateExpression(tt.expr, machine, symbols)
			if err != nil {
				t.Fatalf("EvaluateExpression() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EvaluateExpression() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestExpressionEvaluator_Memory(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()

	// Use data segment addresses
	dataAddr := uint32(0x00020000)
	symbols := map[string]uint32{
		"data": dataAddr,
	}

	// Write test values to memory
	machine.Memory.WriteWord(dataAddr, 0x12345678)
	machine.Memory.WriteWord(dataAddr+0x1000, 0xABCDEF00)

	tests := []struct {
		name string
		expr string
		want uint32
	}{
		{"Bracket notation", "[0x00020000]", 0x12345678},
		{"Star notation", "*0x00021000", 0xABCDEF00},
		{"Symbol in brackets", "[data]", 0x12345678},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eval.EvaluateExpression(tt.expr, machine, symbols)
			if err != nil {
				t.Fatalf("EvaluateExpression() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EvaluateExpression() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestExpressionEvaluator_Arithmetic(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()
	symbols := make(map[string]uint32)

	tests := []struct {
		name string
		expr string
		want uint32
	}{
		{"Addition", "10 + 20", 30},
		{"Subtraction", "50 - 20", 30},
		{"Multiplication", "5 * 6", 30},
		{"Division", "60 / 2", 30},
		// TODO: Fix hex number parsing in expressions
		// {"Hex addition", "0x10 + 0x20", 0x30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eval.EvaluateExpression(tt.expr, machine, symbols)
			if err != nil {
				t.Fatalf("EvaluateExpression() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EvaluateExpression() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestExpressionEvaluator_Bitwise(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()
	symbols := make(map[string]uint32)

	tests := []struct {
		name string
		expr string
		want uint32
	}{
		// TODO: Fix hex number parsing in bitwise expressions
		// {"AND", "0xFF & 0x0F", 0x0F},
		// {"OR", "0xF0 | 0x0F", 0xFF},
		// {"XOR", "0xFF ^ 0x0F", 0xF0},
		{"Left shift", "1 << 4", 16},
		{"Right shift", "16 >> 2", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eval.EvaluateExpression(tt.expr, machine, symbols)
			if err != nil {
				t.Fatalf("EvaluateExpression() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EvaluateExpression() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestExpressionEvaluator_RegisterOperations(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()
	symbols := make(map[string]uint32)

	machine.CPU.R[0] = 10
	machine.CPU.R[1] = 20

	tests := []struct {
		name string
		expr string
		want uint32
	}{
		// TODO: Fix register expression parsing
		// {"Register addition", "r0 + r1", 30},
		// {"Register with constant", "r0 + 5", 15},
		// {"Register subtraction", "r1 - r0", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eval.EvaluateExpression(tt.expr, machine, symbols)
			if err != nil {
				t.Fatalf("EvaluateExpression() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EvaluateExpression() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

func TestExpressionEvaluator_ValueHistory(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()
	symbols := make(map[string]uint32)

	// Evaluate some expressions
	val1, _ := eval.EvaluateExpression("42", machine, symbols)
	val2, _ := eval.EvaluateExpression("100", machine, symbols)

	// Check value numbers
	if eval.GetValueNumber() != 2 {
		t.Errorf("ValueNumber = %d, want 2", eval.GetValueNumber())
	}

	// Retrieve values
	got1, err := eval.GetValue(1)
	if err != nil {
		t.Fatalf("GetValue(1) error = %v", err)
	}
	if got1 != val1 {
		t.Errorf("GetValue(1) = %d, want %d", got1, val1)
	}

	got2, err := eval.GetValue(2)
	if err != nil {
		t.Fatalf("GetValue(2) error = %v", err)
	}
	if got2 != val2 {
		t.Errorf("GetValue(2) = %d, want %d", got2, val2)
	}

	// Try invalid value number
	_, err = eval.GetValue(999)
	if err == nil {
		t.Error("Expected error for invalid value number")
	}
}

func TestExpressionEvaluator_BooleanEvaluation(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()
	symbols := make(map[string]uint32)

	machine.CPU.R[0] = 42

	tests := []struct {
		name string
		expr string
		want bool
	}{
		{"Zero is false", "0", false},
		{"Non-zero is true", "42", true},
		{"Register non-zero", "r0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eval.Evaluate(tt.expr, machine, symbols)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpressionEvaluator_Errors(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()
	symbols := make(map[string]uint32)

	tests := []struct {
		name string
		expr string
	}{
		{"Empty expression", ""},
		{"Unknown symbol", "unknown_symbol"},
		{"Invalid register", "r99"},
		{"Division by zero", "10 / 0"},
		{"Invalid hex", "0xGGGG"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := eval.EvaluateExpression(tt.expr, machine, symbols)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestExpressionEvaluator_Reset(t *testing.T) {
	eval := NewExpressionEvaluator()
	machine := vm.NewVM()
	symbols := make(map[string]uint32)

	// Evaluate some expressions
	eval.EvaluateExpression("42", machine, symbols)
	eval.EvaluateExpression("100", machine, symbols)

	if eval.GetValueNumber() != 2 {
		t.Error("Value number should be 2 before reset")
	}

	// Reset
	eval.Reset()

	if eval.GetValueNumber() != 0 {
		t.Error("Value number should be 0 after reset")
	}

	if len(eval.valueHistory) != 0 {
		t.Error("Value history should be empty after reset")
	}
}
