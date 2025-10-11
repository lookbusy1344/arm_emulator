package integration_test

import (
	"testing"
)

// TestOffset8WithLabel tests storing and loading at offset 8 using a label
func TestOffset8WithLabel(t *testing.T) {
	code := `.org 0x8000
start:
    LDR R4, =array
    MOV R2, #100
    STR R2, [R4]
    MOV R2, #200
    STR R2, [R4, #4]
    MOV R2, #300
    STR R2, [R4, #8]
    MOV R2, #400
    STR R2, [R4, #12]

    ; Read offset 0
    LDR R4, =array
    LDR R3, [R4]
    CMP R3, #100
    BNE fail

    ; Read offset 4
    LDR R4, =array
    LDR R3, [R4, #4]
    CMP R3, #200
    BNE fail

    ; Read offset 8
    LDR R4, =array
    LDR R3, [R4, #8]
    CMP R3, #300
    BNE fail

    ; Read offset 12
    LDR R4, =array
    LDR R3, [R4, #12]
    CMP R3, #400
    BNE fail

    ; Success
    MOV R0, #0
    SWI 0x00

fail:
    MOV R0, #1
    SWI 0x00

    .align 2
array:
    .word 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0

msg1:
    .asciz "Hello"
msg2:
    .asciz "World"
`

	_, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	if exitCode != 0 {
		t.Errorf("Test failed with exit code %d (expected 0)\nStderr: %s", exitCode, stderr)
	}
}
