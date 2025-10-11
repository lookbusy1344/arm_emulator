package integration_test

import (
	"testing"
)

// TestStackFile tests the same code as test_stack.s
func TestStackFile(t *testing.T) {
	code := `; Test using stack memory
        .org    0x8000

_start:
        ; Allocate space on stack
        SUB     SP, SP, #64

        ; Store values
        MOV     R2, #100
        STR     R2, [SP]
        MOV     R2, #200
        STR     R2, [SP, #4]
        MOV     R2, #300
        STR     R2, [SP, #8]
        MOV     R2, #400
        STR     R2, [SP, #12]

        ; Read all values and check them
        LDR     R5, [SP]
        CMP     R5, #100
        BNE     fail

        LDR     R5, [SP, #4]
        CMP     R5, #200
        BNE     fail

        LDR     R5, [SP, #8]
        CMP     R5, #300
        BNE     fail

        LDR     R5, [SP, #12]
        CMP     R5, #400
        BNE     fail

        B       success

fail:
        MOV     R0, #1
        SWI     #0x00

success:

        ; Exit
        MOV     R0, #0
        SWI     #0x00

`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	if exitCode != 0 {
		t.Errorf("Test failed - values at offsets 8 and/or 12 are incorrect\nStderr: %s\nStdout: %s", stderr, stdout)
	}

	t.Logf("Test passed - all values match")
}
