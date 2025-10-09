package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test actual example files from the examples directory
func TestExamplePrograms_Hello(t *testing.T) {
	examplePath := filepath.Join("..", "..", "examples", "hello.s")
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skip("examples/hello.s not found")
	}

	code, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("failed to read hello.s: %v", err)
	}

	stdout, _, exitCode, err := runAssembly(t, string(code))
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Hello, World!") {
		t.Errorf("expected output to contain 'Hello, World!', got %q", stdout)
	}
}

// Test arithmetic operations
func TestProgram_SimpleArithmetic(t *testing.T) {
	code := `
		.org 0x8000
_start:
		; Test addition
		MOV R0, #10
		MOV R1, #5
		ADD R2, R0, R1   ; R2 = 15

		; Print result
		MOV R0, R2
		MOV R1, #10
		SWI #0x03
		SWI #0x07

		; Test subtraction
		MOV R0, #20
		MOV R1, #8
		SUB R2, R0, R1   ; R2 = 12

		; Print result
		MOV R0, R2
		MOV R1, #10
		SWI #0x03

		MOV R0, #0
		SWI #0x00
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines of output, got %d", len(lines))
	}

	if lines[0] != "15" {
		t.Errorf("expected first line '15', got %q", lines[0])
	}

	if lines[1] != "12" {
		t.Errorf("expected second line '12', got %q", lines[1])
	}
}

// Test simple loop
func TestProgram_Loop(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #0       ; counter
		MOV R1, #3       ; limit

loop:
		CMP R0, R1
		BGE end

		; Print counter
		PUSH {R0, R1}
		MOV R2, R0
		MOV R0, R2
		MOV R1, #10
		SWI #0x03
		SWI #0x07
		POP {R0, R1}

		ADD R0, R0, #1
		B loop

end:
		MOV R0, #0
		SWI #0x00
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (0, 1, 2), got %d lines: %v", len(lines), lines)
	}

	expected := []string{"0", "1", "2"}
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("line %d: expected %q, got %q", i, exp, lines[i])
		}
	}
}

// Test conditional execution
func TestProgram_Conditionals(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #10
		MOV R1, #20

		CMP R0, R1
		BLT less_than
		B not_less

less_than:
		LDR R0, =msg_lt
		SWI #0x02
		B end

not_less:
		LDR R0, =msg_ge
		SWI #0x02

end:
		MOV R0, #0
		SWI #0x00

msg_lt:
		.asciz "Less"
msg_ge:
		.asciz "Greater or Equal"
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if stdout != "Less" {
		t.Errorf("expected 'Less', got %q", stdout)
	}
}

// Test function call and return
func TestProgram_FunctionCall(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #5
		BL multiply_by_two
		; R0 now contains 10

		MOV R1, #10
		SWI #0x03
		MOV R0, #0
		SWI #0x00

multiply_by_two:
		ADD R0, R0, R0
		MOV PC, LR
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if strings.TrimSpace(stdout) != "10" {
		t.Errorf("expected '10', got %q", stdout)
	}
}

// Test data processing with immediate values
func TestProgram_ImmediateValues(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #100
		ADD R0, R0, #50
		SUB R0, R0, #25
		; R0 = 125

		MOV R1, #10
		SWI #0x03

		MOV R0, #0
		SWI #0x00
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if strings.TrimSpace(stdout) != "125" {
		t.Errorf("expected '125', got %q", stdout)
	}
}

// Test logical operations
func TestProgram_LogicalOps(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #15      ; 0b1111
		MOV R1, #7       ; 0b0111
		AND R2, R0, R1   ; R2 = 7 (0b0111)

		MOV R0, R2
		MOV R1, #10
		SWI #0x03
		SWI #0x07

		MOV R0, #12      ; 0b1100
		MOV R1, #3       ; 0b0011
		ORR R2, R0, R1   ; R2 = 15 (0b1111)

		MOV R0, R2
		MOV R1, #10
		SWI #0x03

		MOV R0, #0
		SWI #0x00
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	if lines[0] != "7" {
		t.Errorf("AND result: expected '7', got %q", lines[0])
	}

	if lines[1] != "15" {
		t.Errorf("ORR result: expected '15', got %q", lines[1])
	}
}

// Test memory operations
func TestProgram_MemoryOps(t *testing.T) {
	code := `
		.org 0x8000
_start:
		LDR R0, =data_val
		LDR R1, [R0]     ; Load the value 42

		MOV R0, R1
		MOV R1, #10
		SWI #0x03

		MOV R0, #0
		SWI #0x00

data_val:
		.word 42
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if strings.TrimSpace(stdout) != "42" {
		t.Errorf("expected '42', got %q", stdout)
	}
}

// Test shift operations
func TestProgram_Shifts(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #4
		MOV R1, R0, LSL #2   ; R1 = 4 << 2 = 16

		MOV R0, R1
		MOV R1, #10
		SWI #0x03
		SWI #0x07

		MOV R0, #32
		MOV R1, R0, LSR #2   ; R1 = 32 >> 2 = 8

		MOV R0, R1
		MOV R1, #10
		SWI #0x03

		MOV R0, #0
		SWI #0x00
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	if lines[0] != "16" {
		t.Errorf("LSL result: expected '16', got %q", lines[0])
	}

	if lines[1] != "8" {
		t.Errorf("LSR result: expected '8', got %q", lines[1])
	}
}

// Test stack operations
func TestProgram_Stack(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #10
		MOV R1, #20
		MOV R2, #30

		PUSH {R0, R1, R2}

		MOV R0, #0
		MOV R1, #0
		MOV R2, #0

		POP {R0, R1, R2}

		; Print R0 (should be 10)
		PUSH {R1, R2}
		MOV R1, #10
		SWI #0x03
		SWI #0x07
		POP {R1, R2}

		; Print R1 (should be 20)
		MOV R0, R1
		PUSH {R2}
		MOV R1, #10
		SWI #0x03
		SWI #0x07
		POP {R2}

		; Print R2 (should be 30)
		MOV R0, R2
		MOV R1, #10
		SWI #0x03

		MOV R0, #0
		SWI #0x00
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	expected := []string{"10", "20", "30"}
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("line %d: expected %q, got %q", i, exp, lines[i])
		}
	}
}

// Test multiply instruction
func TestProgram_Multiply(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #6
		MOV R1, #7
		MUL R2, R0, R1   ; R2 = 42

		MOV R0, R2
		MOV R1, #10
		SWI #0x03

		MOV R0, #0
		SWI #0x00
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if strings.TrimSpace(stdout) != "42" {
		t.Errorf("expected '42', got %q", stdout)
	}
}

// Test negative numbers
func TestProgram_NegativeNumbers(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #10
		MOV R1, #20
		SUB R2, R0, R1   ; R2 = -10

		MOV R0, R2
		MOV R1, #10
		SWI #0x03

		MOV R0, #0
		SWI #0x00
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if strings.TrimSpace(stdout) != "-10" {
		t.Errorf("expected '-10', got %q", stdout)
	}
}

// Test compare and flags
func TestProgram_CompareFlags(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #5
		MOV R1, #5
		CMP R0, R1
		BEQ equal
		LDR R0, =msg_ne
		B print

equal:
		LDR R0, =msg_eq

print:
		SWI #0x02
		MOV R0, #0
		SWI #0x00

msg_eq:
		.asciz "Equal"
msg_ne:
		.asciz "Not Equal"
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if stdout != "Equal" {
		t.Errorf("expected 'Equal', got %q", stdout)
	}
}
