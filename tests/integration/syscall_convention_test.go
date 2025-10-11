package integration_test

import (
	"strings"
	"testing"
)

// TestSyscallConvention_TraditionalExit tests traditional SWI #0x00 (EXIT)
func TestSyscallConvention_TraditionalExit(t *testing.T) {
	code := `
        .org    0x8000
_start:
        MOV     R0, #42
        SWI     #0x00           ; Traditional EXIT with code 42
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil && !strings.Contains(err.Error(), "program exited with code 42") {
		t.Fatalf("Unexpected error: %v\nStderr: %s", err, stderr)
	}

	if exitCode != 42 {
		t.Errorf("Expected exit code 42, got %d\nStdout: %s\nStderr: %s", exitCode, stdout, stderr)
	}
}

// TestSyscallConvention_TraditionalExitWithR7 tests that SWI #0x00 works
// regardless of R7 contents (R7 is now just a general-purpose register)
func TestSyscallConvention_TraditionalExitWithR7(t *testing.T) {
	code := `
        .org    0x8000
_start:
        LDR     R7, =0xDEADBEEF ; Put any value in R7
        MOV     R0, #0
        SWI     #0x00           ; Should exit cleanly (R7 is irrelevant)
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil && !strings.Contains(err.Error(), "program exited with code 0") {
		t.Errorf("Program should exit cleanly regardless of R7 value")
		t.Errorf("Error: %v\nStderr: %s", err, stderr)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStdout: %s", exitCode, stdout)
	}
}

// TestSyscallConvention_AllTraditionalSyscalls tests all traditional syscalls still work
func TestSyscallConvention_AllTraditionalSyscalls(t *testing.T) {
	code := `
        .org    0x8000
_start:
        ; Test WRITE_CHAR (0x01)
        MOV     R0, #88         ; 'X' in ASCII
        SWI     #0x01

        ; Test WRITE_STRING (0x02)
        LDR     R0, =msg
        SWI     #0x02

        ; Test WRITE_INT (0x03)
        MOV     R0, #42
        MOV     R1, #10
        SWI     #0x03

        ; Test WRITE_NEWLINE (0x07)
        SWI     #0x07

        ; Test EXIT (0x00)
        MOV     R0, #0
        SWI     #0x00

msg:    .asciz "Y"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil && !strings.Contains(err.Error(), "program exited with code 0") {
		t.Fatalf("Unexpected error: %v\nStderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "X") {
		t.Errorf("Missing WRITE_CHAR output 'X' in: %s", stdout)
	}
	if !strings.Contains(stdout, "Y") {
		t.Errorf("Missing WRITE_STRING output 'Y' in: %s", stdout)
	}
	if !strings.Contains(stdout, "42") {
		t.Errorf("Missing WRITE_INT output '42' in: %s", stdout)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

// TestSyscallConvention_ComplexProgramWithR7 tests a complex program that uses R7
// for calculations but still exits properly
func TestSyscallConvention_ComplexProgramWithR7(t *testing.T) {
	code := `
        .org    0x8000
_start:
        ; Use R7 for calculations (this was causing the bug)
        MOV     R7, #100
        MOV     R8, #200
        ADD     R7, R7, R8      ; R7 = 300

        ; Multiply by 1000
        MOV     R9, #1000
        MUL     R7, R9, R7      ; R7 = 300000 (garbage for syscall)

        ; Print a message
        LDR     R0, =msg
        SWI     #0x02

        ; Exit - should work even though R7 has garbage
        MOV     R0, #0
        SWI     #0x00

msg:    .asciz "Test passed"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil && !strings.Contains(err.Error(), "program exited with code 0") {
		t.Errorf("Program should exit cleanly regardless of R7 value")
		t.Errorf("Error: %v\nStderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "Test passed") {
		t.Errorf("Expected 'Test passed' in output, got: %s", stdout)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
