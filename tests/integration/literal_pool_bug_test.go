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

// TestLiteralPoolBug_WithSubroutines tests literals in programs with subroutines
// This is closer to real-world usage and should reproduce the bug
func TestLiteralPoolBug_WithSubroutines(t *testing.T) {
	code := `
        .org    0x8000
_start:
        LDR     R0, =msg1
        SWI     #0x02

        BL      sub1
        BL      sub2
        BL      sub3

        LDR     R0, =msg_end
        SWI     #0x02
        MOV     R0, #0
        SWI     #0x00

sub1:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg2
        SWI     #0x02
        LDMFD   SP!, {R0, PC}

sub2:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg3
        SWI     #0x02
        LDMFD   SP!, {R0, PC}

sub3:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg4
        SWI     #0x02
        LDR     R0, =msg5
        SWI     #0x02
        LDMFD   SP!, {R0, PC}

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
msg_end:
        .asciz  "F"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	expectedOutput := "ABCDEF"
	if !strings.Contains(stdout, expectedOutput) {
		t.Errorf("Expected '%s' in output, got: %s", expectedOutput, stdout)
	}

	if exitCode != 0 {
		t.Errorf("BUG REPRODUCED: Exit code %d (expected 0)\nStderr: %s", exitCode, stderr)
		if strings.Contains(stderr, "unimplemented SWI") {
			t.Logf("BUG CONFIRMED: Got 'unimplemented SWI' error after correct execution")
		}
	}
}

// TestLiteralPoolBug_ComplexProgram tests a complex program similar to arrays.s
// This should definitively reproduce the bug
func TestLiteralPoolBug_ComplexProgram(t *testing.T) {
	code := `
        .org    0x8000
_start:
        LDR     R0, =msg_start
        SWI     #0x02
        SWI     #0x07           ; WRITE_NEWLINE

        ; Call multiple subroutines
        BL      process1
        BL      process2
        BL      process3

        LDR     R0, =msg_done
        SWI     #0x02
        SWI     #0x07

        MOV     R0, #0
        SWI     #0x00           ; EXIT

process1:
        STMFD   SP!, {R0-R2, LR}
        LDR     R0, =msg_p1
        SWI     #0x02
        LDR     R1, =data1
        LDR     R1, [R1]
        MOV     R0, R1
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07
        LDMFD   SP!, {R0-R2, PC}

process2:
        STMFD   SP!, {R0-R2, LR}
        LDR     R0, =msg_p2
        SWI     #0x02
        LDR     R1, =data2
        LDR     R1, [R1]
        MOV     R0, R1
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07
        LDMFD   SP!, {R0-R2, PC}

process3:
        STMFD   SP!, {R0-R2, LR}
        LDR     R0, =msg_p3
        SWI     #0x02
        LDR     R1, =data3
        LDR     R1, [R1]
        MOV     R0, R1
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07
        LDMFD   SP!, {R0-R2, PC}

        .align  2
data1:
        .word   100
data2:
        .word   200
data3:
        .word   300

msg_start:
        .asciz  "Complex Program Test"
msg_p1:
        .asciz  "Process 1: "
msg_p2:
        .asciz  "Process 2: "
msg_p3:
        .asciz  "Process 3: "
msg_done:
        .asciz  "All done!"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	// Check for expected output
	if !strings.Contains(stdout, "Complex Program Test") {
		t.Errorf("Missing 'Complex Program Test' in output: %s", stdout)
	}
	if !strings.Contains(stdout, "Process 1: 100") {
		t.Errorf("Missing 'Process 1: 100' in output: %s", stdout)
	}
	if !strings.Contains(stdout, "Process 2: 200") {
		t.Errorf("Missing 'Process 2: 200' in output: %s", stdout)
	}
	if !strings.Contains(stdout, "Process 3: 300") {
		t.Errorf("Missing 'Process 3: 300' in output: %s", stdout)
	}

	if exitCode != 0 {
		t.Errorf("BUG REPRODUCED: Exit code %d (expected 0)\nStderr: %s", exitCode, stderr)
		if strings.Contains(stderr, "unimplemented SWI") {
			t.Logf("BUG CONFIRMED: Got 'unimplemented SWI' error - this is the literal pool bug!")
			t.Logf("The program produces correct output but fails to halt properly")
		}
	}
}

// TestLiteralPoolBug_ManyLiterals tests a program with 15+ literals
// This should trigger the bug based on TODO.md documentation
func TestLiteralPoolBug_ManyLiterals(t *testing.T) {
	code := `
        .org    0x8000
_start:
        LDR     R0, =msg1
        SWI     #0x02
        BL      sub1
        LDR     R0, =msg2
        SWI     #0x02
        BL      sub2
        LDR     R0, =msg3
        SWI     #0x02
        BL      sub3
        LDR     R0, =msg4
        SWI     #0x02
        MOV     R0, #0
        SWI     #0x00

sub1:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg5
        SWI     #0x02
        LDR     R0, =msg6
        SWI     #0x02
        LDR     R0, =msg7
        SWI     #0x02
        LDMFD   SP!, {R0, PC}

sub2:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg8
        SWI     #0x02
        LDR     R0, =msg9
        SWI     #0x02
        LDR     R0, =msg10
        SWI     #0x02
        LDMFD   SP!, {R0, PC}

sub3:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg11
        SWI     #0x02
        LDR     R0, =msg12
        SWI     #0x02
        LDR     R0, =msg13
        SWI     #0x02
        LDR     R0, =msg14
        SWI     #0x02
        LDR     R0, =msg15
        SWI     #0x02
        LDMFD   SP!, {R0, PC}

msg1:   .asciz "1"
msg2:   .asciz "2"
msg3:   .asciz "3"
msg4:   .asciz "4"
msg5:   .asciz "5"
msg6:   .asciz "6"
msg7:   .asciz "7"
msg8:   .asciz "8"
msg9:   .asciz "9"
msg10:  .asciz "A"
msg11:  .asciz "B"
msg12:  .asciz "C"
msg13:  .asciz "D"
msg14:  .asciz "E"
msg15:  .asciz "F"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("Execution error: %v\nStderr: %s", err, stderr)
	}

	expectedOutput := "1567289A3BCDEF4"
	if !strings.Contains(stdout, expectedOutput) {
		t.Errorf("Expected '%s' in output, got: %s", expectedOutput, stdout)
	}

	if exitCode != 0 {
		t.Errorf("BUG REPRODUCED: Exit code %d (expected 0)\nStderr: %s", exitCode, stderr)
		if strings.Contains(stderr, "unimplemented SWI") {
			t.Logf("BUG CONFIRMED: 15+ literals with subroutines triggers the bug")
		}
	}
}

// TestLiteralPoolBug_AddressingModes tests the addressing_modes.s example program
// This is a real-world case that FAILS with "unimplemented SWI" error
func TestLiteralPoolBug_AddressingModes(t *testing.T) {
	// This test reproduces the exact bug from examples/addressing_modes.s
	// It has 11 LDR Rx, =label instructions and produces correct output
	// but then fails with "Runtime error at PC=0x000080D8: unimplemented SWI: 0x04FFC4"

	code := `
        .org    0x8000
_start:
        LDR     R0, =msg_intro
        SWI     #0x02
        SWI     #0x07
        SWI     #0x07

        SUB     SP, SP, #64
        MOV     R2, #100
        STR     R2, [SP]
        MOV     R2, #200
        STR     R2, [SP, #4]
        MOV     R2, #300
        STR     R2, [SP, #8]

        LDR     R3, [SP, #4]
        CMP     R3, #200
        BNE     fail1

        MOV     R4, SP
        LDR     R5, [R4, #8]!
        CMP     R5, #300
        BNE     fail2

        SUB     R0, R4, SP
        CMP     R0, #8
        BNE     fail2

        MOV     R7, SP
        LDR     R8, [R7], #4
        CMP     R8, #100
        BNE     fail3

        SUB     R0, R7, SP
        CMP     R0, #4
        BNE     fail3

        MOV     R9, SP
        MOV     R10, #8
        LDR     R11, [R9, R10]
        CMP     R11, #300
        BNE     fail4

        MOV     R12, SP
        MOV     R1, #1
        LDR     R2, [R12, R1, LSL #2]
        CMP     R2, #200
        BNE     fail5

        SUB     SP, SP, #16
        MOV     R3, SP
        MOV     R4, #'A'
        STRB    R4, [R3], #1
        MOV     R4, #'R'
        STRB    R4, [R3], #1
        MOV     R4, #'M'
        STRB    R4, [R3]

        MOV     R5, SP
        LDRB    R6, [R5], #1
        CMP     R6, #'A'
        BNE     fail6
        LDRB    R6, [R5], #1
        CMP     R6, #'R'
        BNE     fail6
        LDRB    R6, [R5]
        CMP     R6, #'M'
        BNE     fail6

        LDR     R0, =msg_success
        SWI     #0x02
        SWI     #0x07
        MOV     R0, #0
        SWI     #0x00

fail1:
        LDR     R0, =msg_fail1
        B       fail_exit
fail2:
        LDR     R0, =msg_fail2
        B       fail_exit
fail3:
        LDR     R0, =msg_fail3
        B       fail_exit
fail4:
        LDR     R0, =msg_fail4
        B       fail_exit
fail5:
        LDR     R0, =msg_fail5
        B       fail_exit
fail6:
        LDR     R0, =msg_fail6

fail_exit:
        SWI     #0x02
        SWI     #0x07
        MOV     R0, #1
        SWI     #0x00

msg_intro:
        .asciz  "Testing ARM Addressing Modes..."
msg_success:
        .asciz  "All addressing mode tests passed!"
msg_fail1:
        .asciz  "FAIL: Immediate offset [Rn, #offset]"
msg_fail2:
        .asciz  "FAIL: Pre-indexed with writeback [Rn, #offset]!"
msg_fail3:
        .asciz  "FAIL: Post-indexed [Rn], #offset"
msg_fail4:
        .asciz  "FAIL: Register offset [Rn, Rm]"
msg_fail5:
        .asciz  "FAIL: Scaled register offset [Rn, Rm, LSL #shift]"
msg_fail6:
        .asciz  "FAIL: Byte access"
`

	stdout, stderr, exitCode, err := runAssembly(t, code)

	// Check for expected successful output (should be present even if bug occurs)
	if !strings.Contains(stdout, "All addressing mode tests passed!") {
		if err != nil {
			t.Logf("Execution error occurred: %v", err)
		}
		t.Errorf("Expected success message in output, got: %s", stdout)
	}

	// THIS IS THE BUG: Program executes correctly but fails to halt
	if err != nil || exitCode != 0 {
		// BUG REPRODUCED!
		t.Errorf("BUG REPRODUCED: Program produced correct output but failed to halt properly")
		t.Errorf("Exit code: %d, Error: %v", exitCode, err)
		t.Errorf("Stderr: %s", stderr)

		if err != nil && strings.Contains(err.Error(), "unimplemented SWI") {
			t.Logf("BUG CONFIRMED: Got 'unimplemented SWI' error after correct execution")
			t.Logf("This is the classic literal pool bug - execution continues past SWI #0x00")
			t.Logf("The PC runs into literal pool data which is interpreted as an invalid SWI")
		}
	}
}
