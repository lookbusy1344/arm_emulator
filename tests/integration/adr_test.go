package integration_test

import (
	"strings"
	"testing"
)

// TestADRBasic tests basic ADR instruction (forward reference)
func TestADRBasic(t *testing.T) {
	program := `
		.org 0x8000
_start:	ADR R0, message
		; Print the address in R0
		MOV R1, #16
		SWI #0x03
		SWI #0x07
		MOV R0, #0
		SWI #0x00
message:
		.word 0x12345678
	`

	stdout, _, exitCode, err := runAssembly(t, program)
	if err != nil {
		t.Fatalf("Failed to run program: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Check that output contains a valid address (should be in 0x8000 range)
	// The message label is at some offset from start, should print like 8010 or similar
	if !strings.Contains(stdout, "80") {
		t.Errorf("Expected address output to contain '80' (0x8000 range), got: %q", stdout)
	}
}

// TestADRBackward tests ADR with backward reference
func TestADRBackward(t *testing.T) {
	program := `
		.org 0x8000
data:	.word 0xDEADBEEF
_start:	ADR R1, data
		; Load the word at data address
		LDR R2, [R1]
		; Print R2 (should be 0xDEADBEEF)
		MOV R0, R2
		MOV R1, #16
		SWI #0x03
		MOV R0, #0
		SWI #0x00
	`

	stdout, _, exitCode, err := runAssembly(t, program)
	if err != nil {
		t.Fatalf("Failed to run program: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Check that output contains deadbeef
	if !strings.Contains(strings.ToLower(stdout), "deadbeef") {
		t.Errorf("Expected output to contain 'deadbeef', got: %q", stdout)
	}
}

// TestADRMultiple tests multiple ADR instructions loading different addresses
func TestADRMultiple(t *testing.T) {
	program := `
		.org 0x8000
_start:	ADR R0, label1
		ADR R1, label2
		ADR R2, label3
		; Verify all three loaded correctly by dereferencing
		LDR R3, [R0]    ; Should be 0x11111111
		LDR R4, [R1]    ; Should be 0x22222222
		LDR R5, [R2]    ; Should be 0x33333333
		; Print R3 to verify
		MOV R0, R3
		MOV R1, #16
		SWI #0x03
		MOV R0, #0
		SWI #0x00
label1:	.word 0x11111111
label2:	.word 0x22222222
label3:	.word 0x33333333
	`

	stdout, _, exitCode, err := runAssembly(t, program)
	if err != nil {
		t.Fatalf("Failed to run program: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Check output contains the first value
	if !strings.Contains(strings.ToLower(stdout), "11111111") {
		t.Errorf("Expected output to contain '11111111', got: %q", stdout)
	}
}

// TestADRLoadAndDereference tests using ADR to load address then dereferencing
func TestADRLoadAndDereference(t *testing.T) {
	program := `
		.org 0x8000
_start:	ADR R0, data
		LDR R1, [R0]        ; Load value at data address
		; Print the value
		MOV R0, R1
		MOV R1, #16
		SWI #0x03
		MOV R0, #0
		SWI #0x00
data:	.word 0xCAFEBABE
	`

	stdout, _, exitCode, err := runAssembly(t, program)
	if err != nil {
		t.Fatalf("Failed to run program: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Check output
	if !strings.Contains(strings.ToLower(stdout), "cafebabe") {
		t.Errorf("Expected output to contain 'cafebabe', got: %q", stdout)
	}
}

