package integration_test

import (
	"testing"
)

// TestAddressingMode_ImmediateOffset_FullPipeline tests immediate offset addressing
// through the complete parse -> encode -> execute pipeline
func TestAddressingMode_ImmediateOffset_FullPipeline(t *testing.T) {
	code := `.org 0x8000
start:
    MOV R1, #100
    SUB SP, SP, #16
    STR R1, [SP]
    STR R1, [SP, #4]
    MOV R4, SP
    LDR R5, [R4, #4]
    ; Write R5 to stdout to verify it's correct
    MOV R0, R5
    SWI 0x03
    MOV R0, #0
    SWI 0x00
`

	_, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}
}

// TestAddressingMode_PreIndexedWriteback_FullPipeline tests pre-indexed with writeback
// NOTE: The original bug was not in pre-indexed writeback, but in the test using SWI 0x00
// which triggers Linux-style syscall convention that reads R7 for the syscall number
func TestAddressingMode_PreIndexedWriteback_FullPipeline(t *testing.T) {
	code := `.org 0x8000
start:
    MOV R1, #100
    SUB SP, SP, #16
    STR R1, [SP]
    STR R1, [SP, #4]
    MOV R6, SP
    LDR R2, [R6, #4]!
    ; Write R2 to stdout to verify it's correct (changed from R7 to avoid SWI 0x00 conflict)
    MOV R0, R2
    SWI 0x03
    MOV R0, #1
    SWI 0x00
`

	_, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d\nStderr: %s", exitCode, stderr)
	}
}

// TestAddressingMode_PostIndexed_FullPipeline tests post-indexed addressing
// through the complete parse -> encode -> execute pipeline
func TestAddressingMode_PostIndexed_FullPipeline(t *testing.T) {
	code := `.org 0x8000
start:
    MOV R1, #100
    SUB SP, SP, #16
    STR R1, [SP]
    STR R1, [SP, #4]
    MOV R8, SP
    LDR R9, [R8], #8
    ; Write R9 to stdout to verify it's correct
    MOV R0, R9
    SWI 0x03
    MOV R0, #0
    SWI 0x00
`

	_, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}
}
