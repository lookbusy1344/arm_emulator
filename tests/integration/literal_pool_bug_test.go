package integration_test

import (
	"strings"
	"testing"
)

// TestLiteralPoolBug_Simple tests a simple case with 2 literals
// This should PASS
func TestLiteralPoolBug_Simple(t *testing.T) {
	code := `
        .org    0x8000
_start:
        LDR     R0, =msg1
        SWI     #0x02
        LDR     R0, =msg2
        SWI     #0x02
        MOV     R0, #0
        SWI     #0x00

msg1:
        .asciz  "Hello "
msg2:
        .asciz  "World"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	if exitCode != 0 {
		t.Errorf("Test failed with exit code %d (expected 0)\nStderr: %s\nStdout: %s", exitCode, stderr, stdout)
	}

	if !strings.Contains(stdout, "Hello World") {
		t.Errorf("Expected 'Hello World' in output, got: %s", stdout)
	}
}

// TestLiteralPoolBug_Medium tests with 5 literals
// This may or may not fail depending on the bug
func TestLiteralPoolBug_Medium(t *testing.T) {
	code := `
        .org    0x8000
_start:
        LDR     R0, =msg1
        SWI     #0x02
        LDR     R0, =msg2
        SWI     #0x02
        LDR     R0, =msg3
        SWI     #0x02
        LDR     R0, =msg4
        SWI     #0x02
        LDR     R0, =msg5
        SWI     #0x02
        MOV     R0, #0
        SWI     #0x00

msg1:
        .asciz  "A"
msg2:
        .asciz  "B"
msg3:
        .asciz  "C"
msg4:
        .asciz  "D"
msg5:
        .asciz  "E"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	if exitCode != 0 {
		t.Logf("WARNING: Test failed with exit code %d (bug reproduced?)\nStderr: %s", exitCode, stderr)
		t.Logf("Output: %s", stdout)
		// Don't fail the test - we're documenting the bug
		if strings.Contains(stderr, "unimplemented SWI") {
			t.Logf("BUG CONFIRMED: Got 'unimplemented SWI' error")
		}
	}

	if !strings.Contains(stdout, "ABCDE") {
		t.Logf("Expected 'ABCDE' in output, got: %s", stdout)
	}
}

// TestLiteralPoolBug_Many tests with 10 literals
// This should reproduce the bug based on the issue description
func TestLiteralPoolBug_Many(t *testing.T) {
	code := `
        .org    0x8000
_start:
        LDR     R0, =msg1
        SWI     #0x02
        LDR     R0, =msg2
        SWI     #0x02
        LDR     R0, =msg3
        SWI     #0x02
        LDR     R0, =msg4
        SWI     #0x02
        LDR     R0, =msg5
        SWI     #0x02
        LDR     R0, =msg6
        SWI     #0x02
        LDR     R0, =msg7
        SWI     #0x02
        LDR     R0, =msg8
        SWI     #0x02
        LDR     R0, =msg9
        SWI     #0x02
        LDR     R0, =msg10
        SWI     #0x02
        MOV     R0, #0
        SWI     #0x00

msg1:
        .asciz  "1"
msg2:
        .asciz  "2"
msg3:
        .asciz  "3"
msg4:
        .asciz  "4"
msg5:
        .asciz  "5"
msg6:
        .asciz  "6"
msg7:
        .asciz  "7"
msg8:
        .asciz  "8"
msg9:
        .asciz  "9"
msg10:
        .asciz  "0"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	expectedOutput := "1234567890"
	if !strings.Contains(stdout, expectedOutput) {
		t.Errorf("Expected '%s' in output, got: %s", expectedOutput, stdout)
	}

	// Document the bug
	if exitCode != 0 {
		t.Logf("BUG REPRODUCED: Exit code %d (expected 0)", exitCode)
		t.Logf("Stderr: %s", stderr)
		if strings.Contains(stderr, "unimplemented SWI") {
			t.Logf("BUG CONFIRMED: Got 'unimplemented SWI' error after correct execution")
		}
		// Don't fail - we're documenting the known bug
	} else {
		t.Logf("UNEXPECTED: Test passed! Bug may be fixed.")
	}
}

// TestLiteralPoolBug_WithBranches tests literals with branches
// This tests if branches are affected by literal pool placement
func TestLiteralPoolBug_WithBranches(t *testing.T) {
	code := `
        .org    0x8000
_start:
        MOV     R5, #0
        LDR     R0, =msg1
        SWI     #0x02

        CMP     R5, #0
        BEQ     path1
        B       path2

path1:
        LDR     R0, =msg2
        SWI     #0x02
        B       end

path2:
        LDR     R0, =msg3
        SWI     #0x02

end:
        LDR     R0, =msg4
        SWI     #0x02
        LDR     R0, =msg5
        SWI     #0x02
        MOV     R0, #0
        SWI     #0x00

msg1:
        .asciz  "A"
msg2:
        .asciz  "B"
msg3:
        .asciz  "C"
msg4:
        .asciz  "D"
msg5:
        .asciz  "E"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	expectedOutput := "ABDE"
	if !strings.Contains(stdout, expectedOutput) {
		t.Errorf("Expected '%s' in output, got: %s", expectedOutput, stdout)
	}

	if exitCode != 0 {
		t.Logf("BUG REPRODUCED: Exit code %d with branches\nStderr: %s", exitCode, stderr)
	}
}

// TestLiteralPoolBug_WithLoops tests literals inside a loop
// This tests if repeated literal loads cause issues
func TestLiteralPoolBug_WithLoops(t *testing.T) {
	code := `
        .org    0x8000
_start:
        MOV     R5, #0          ; Counter

loop:
        CMP     R5, #3
        BGE     done

        LDR     R0, =msg1
        SWI     #0x02
        LDR     R0, =msg2
        SWI     #0x02

        ADD     R5, R5, #1
        B       loop

done:
        LDR     R0, =msg3
        SWI     #0x02
        MOV     R0, #0
        SWI     #0x00

msg1:
        .asciz  "X"
msg2:
        .asciz  "Y"
msg3:
        .asciz  "Z"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	expectedOutput := "XYXYXYZ"
	if !strings.Contains(stdout, expectedOutput) {
		t.Errorf("Expected '%s' in output, got: %s", expectedOutput, stdout)
	}

	if exitCode != 0 {
		t.Logf("BUG REPRODUCED: Exit code %d with loops\nStderr: %s", exitCode, stderr)
	}
}

// TestLiteralPoolBug_RepeatedLabel tests loading the same label multiple times
// This checks if the literal pool correctly handles duplicates
func TestLiteralPoolBug_RepeatedLabel(t *testing.T) {
	code := `
        .org    0x8000
_start:
        LDR     R0, =msg
        SWI     #0x02
        LDR     R0, =msg
        SWI     #0x02
        LDR     R0, =msg
        SWI     #0x02
        LDR     R0, =msg
        SWI     #0x02
        LDR     R0, =msg
        SWI     #0x02
        LDR     R0, =msg
        SWI     #0x02
        LDR     R0, =msg
        SWI     #0x02
        LDR     R0, =msg
        SWI     #0x02
        MOV     R0, #0
        SWI     #0x00

msg:
        .asciz  "OK"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	// Should print "OK" 8 times
	okCount := strings.Count(stdout, "OK")
	if okCount != 8 {
		t.Errorf("Expected 8 'OK' in output, got %d: %s", okCount, stdout)
	}

	if exitCode != 0 {
		t.Logf("BUG REPRODUCED: Exit code %d with repeated labels\nStderr: %s", exitCode, stderr)
	}
}
