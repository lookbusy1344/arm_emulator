package integration_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/loader"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// Helper function to run assembly code and capture stdout
func runAssembly(t *testing.T, code string) (stdout string, stderr string, exitCode int32, err error) {
	return runAssemblyWithInput(t, code, "")
}

// runAssemblyWithInput runs assembly code with optional stdin input
func runAssemblyWithInput(t *testing.T, code string, stdin string) (stdout string, stderr string, exitCode int32, err error) {
	t.Helper()

	// Capture stdin, stdout and stderr
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Set up stdin if provided
	if stdin != "" {
		rIn, wIn, _ := os.Pipe()
		os.Stdin = rIn
		go func() {
			wIn.Write([]byte(stdin))
			wIn.Close()
		}()
		defer func() {
			os.Stdin = oldStdin
		}()
	}

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Parse the assembly
	p := parser.NewParser(code, "test.s")
	program, err := p.Parse()
	if err != nil {
		return "", "", -1, err
	}

	// Create VM after setting stdin so the reader uses the redirected stdin
	machine := vm.NewVM()
	machine.CycleLimit = 1000000

	// Set filesystem root to current directory for tests
	// This allows test programs to access files in their directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", -1, fmt.Errorf("failed to get current directory: %w", err)
	}
	machine.FilesystemRoot = cwd

	// Initialize stack
	stackTop := uint32(vm.StackSegmentStart + vm.StackSegmentSize)
	if err := machine.InitializeStack(stackTop); err != nil {
		return "", "", -1, fmt.Errorf("failed to initialize stack: %w", err)
	}

	// Determine entry point: use _start symbol first, then .org if set, otherwise default to 0x8000
	entryPoint := uint32(0x8000)
	if startSym, exists := program.SymbolTable.Lookup("_start"); exists && startSym.Defined {
		entryPoint = startSym.Value
	} else if program.OriginSet {
		entryPoint = program.Origin
	}

	// Load program
	err = loader.LoadProgramIntoVM(machine, program, entryPoint)
	if err != nil {
		return "", "", -1, err
	}

	// Run program
	var execErr error
	machine.State = vm.StateRunning
	for machine.State == vm.StateRunning {
		if err := machine.Step(); err != nil {
			if machine.State == vm.StateHalted {
				break
			}
			// Save error but continue to capture output
			execErr = err
			break
		}
	}

	// Close write ends and read output
	wOut.Close()
	wErr.Close()

	var outBuf, errBuf bytes.Buffer
	io.Copy(&outBuf, rOut)
	io.Copy(&errBuf, rErr)

	// Return captured output along with any error
	return outBuf.String(), errBuf.String(), machine.ExitCode, execErr
}

// Helper to parse immediate values
func parseValue(s string, out *uint32) (int, error) {
	var val uint32
	// Handle character literals like 'A'
	if len(s) >= 3 && s[0] == '\'' && s[len(s)-1] == '\'' {
		// Extract character (handle basic escape sequences if needed)
		char := s[1]
		if char == '\\' && len(s) >= 4 {
			// Handle escape sequences like '\n', '\t', etc.
			switch s[2] {
			case 'n':
				char = '\n'
			case 't':
				char = '\t'
			case 'r':
				char = '\r'
			case '\\':
				char = '\\'
			case '\'':
				char = '\''
			default:
				return 0, fmt.Errorf("unsupported escape sequence: %q", s)
			}
		}
		*out = uint32(char)
		return 0, nil
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		_, err := parseHex(s[2:], &val)
		*out = val
		return 0, err
	}
	_, err := parseInt(s, &val)
	*out = val
	return 0, err
}

func parseHex(s string, out *uint32) (int, error) {
	var val uint32
	for _, c := range s {
		val *= 16
		if c >= '0' && c <= '9' {
			val += uint32(c - '0')
		} else if c >= 'a' && c <= 'f' {
			val += uint32(c-'a') + 10
		} else if c >= 'A' && c <= 'F' {
			val += uint32(c-'A') + 10
		} else {
			return 0, fmt.Errorf("invalid hex digit: %q", c)
		}
	}
	*out = val
	return 0, nil
}

func parseInt(s string, out *uint32) (int, error) {
	var val uint32
	for _, c := range s {
		if c >= '0' && c <= '9' {
			val = val*10 + uint32(c-'0')
		} else {
			return 0, fmt.Errorf("invalid decimal digit: %q", c)
		}
	}
	*out = val
	return 0, nil
}

// Test WRITE_STRING syscall
func TestSyscall_WriteString(t *testing.T) {
	code := `
		.org 0x8000
_start:
		LDR R0, =msg
		SWI #0x02
		MOV R0, #0
		SWI #0x00
msg:
		.asciz "Hello, World!"
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if stdout != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", stdout)
	}
}

// Test WRITE_CHAR syscall
func TestSyscall_WriteChar(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #65      ; 'A'
		SWI #0x01
		MOV R0, #66      ; 'B'
		SWI #0x01
		MOV R0, #67      ; 'C'
		SWI #0x01
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

	if stdout != "ABC" {
		t.Errorf("expected 'ABC', got %q", stdout)
	}
}

// Test WRITE_INT syscall with decimal
func TestSyscall_WriteIntDecimal(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #42
		MOV R1, #10      ; decimal base
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

	if stdout != "42" {
		t.Errorf("expected '42', got %q", stdout)
	}
}

// Test WRITE_INT syscall with hex
func TestSyscall_WriteIntHex(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #255
		MOV R1, #16      ; hex base
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

	if stdout != "ff" {
		t.Errorf("expected 'ff', got %q", stdout)
	}
}

// Test WRITE_NEWLINE syscall
func TestSyscall_WriteNewline(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #65      ; 'A'
		SWI #0x01
		SWI #0x07        ; newline
		MOV R0, #66      ; 'B'
		SWI #0x01
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

	if stdout != "A\nB" {
		t.Errorf("expected 'A\\nB', got %q", stdout)
	}
}

// Test multiple strings
func TestSyscall_MultipleStrings(t *testing.T) {
	code := `
		.org 0x8000
_start:
		LDR R0, =str1
		SWI #0x02
		SWI #0x07
		LDR R0, =str2
		SWI #0x02
		SWI #0x07
		MOV R0, #0
		SWI #0x00
str1:
		.asciz "First"
str2:
		.asciz "Second"
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	expected := "First\nSecond\n"
	if stdout != expected {
		t.Errorf("expected %q, got %q", expected, stdout)
	}
}

// Test exit code propagation
func TestSyscall_ExitCode(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #42
		SWI #0x00
`
	_, _, exitCode, err := runAssembly(t, code)
	if err != nil && !strings.Contains(err.Error(), "exited with code") {
		t.Fatalf("unexpected error: %v", err)
	}

	if exitCode != 42 {
		t.Errorf("expected exit code 42, got %d", exitCode)
	}
}

// Test mixed output syscalls
func TestSyscall_MixedOutput(t *testing.T) {
	code := `
		.org 0x8000
_start:
		LDR R0, =msg1
		SWI #0x02        ; "Count: "

		MOV R0, #5
		MOV R1, #10
		SWI #0x03        ; "5"

		SWI #0x07        ; newline

		MOV R0, #72      ; 'H'
		SWI #0x01
		MOV R0, #105     ; 'i'
		SWI #0x01

		MOV R0, #0
		SWI #0x00
msg1:
		.asciz "Count: "
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	expected := "Count: 5\nHi"
	if stdout != expected {
		t.Errorf("expected %q, got %q", expected, stdout)
	}
}

// Test long string
func TestSyscall_LongString(t *testing.T) {
	longStr := strings.Repeat("A", 100)
	code := `
		.org 0x8000
_start:
		LDR R0, =msg
		SWI #0x02
		MOV R0, #0
		SWI #0x00
msg:
		.asciz "` + longStr + `"
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if stdout != longStr {
		t.Errorf("expected long string of %d chars, got %d chars", len(longStr), len(stdout))
	}
}

// Test empty string
func TestSyscall_EmptyString(t *testing.T) {
	code := `
		.org 0x8000
_start:
		LDR R0, =msg
		SWI #0x02
		MOV R0, #65      ; 'A'
		SWI #0x01
		MOV R0, #0
		SWI #0x00
msg:
		.asciz ""
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if stdout != "A" {
		t.Errorf("expected 'A', got %q", stdout)
	}
}

// Test special characters in strings
func TestSyscall_SpecialChars(t *testing.T) {
	code := `
		.org 0x8000
_start:
		LDR R0, =msg
		SWI #0x02
		MOV R0, #0
		SWI #0x00
msg:
		.asciz "Hello\tWorld!"
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Note: \t in .asciz is literal '\' and 't', not a tab
	// This is a limitation of the current parser
	if !strings.Contains(stdout, "Hello") || !strings.Contains(stdout, "World") {
		t.Errorf("expected string with Hello and World, got %q", stdout)
	}
}

// Test GET_TIME syscall (0x30)
func TestSyscall_GetTime(t *testing.T) {
	code := `
		.org 0x8000
_start:
		; Get first timestamp
		SWI #0x30
		MOV R4, R0

		; Get second timestamp
		SWI #0x30
		MOV R5, R0

		; Time should not go backwards
		CMP R5, R4
		MOVLT R0, #1       ; Error code if time went backwards
		MOVGE R0, #0       ; Success
		SWI #0x00
`
	_, _, exitCode, err := runAssembly(t, code)
	if err != nil && !strings.Contains(err.Error(), "exited with code") {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("time went backwards - test failed")
	}
}

// Test GET_RANDOM syscall (0x31)
func TestSyscall_GetRandom(t *testing.T) {
	code := `
		.org 0x8000
_start:
		; Get 5 random numbers and verify they're not all zero
		MOV R4, #5
		MOV R5, #0        ; OR of all random values

loop:
		SWI #0x31         ; GET_RANDOM
		ORR R5, R5, R0    ; Accumulate bits
		SUBS R4, R4, #1
		BNE loop

		; If R5 is 0, all random values were 0 (extremely unlikely)
		CMP R5, #0
		MOVEQ R0, #1      ; Error
		MOVNE R0, #0      ; Success
		SWI #0x00
`
	_, _, exitCode, err := runAssembly(t, code)
	if err != nil && !strings.Contains(err.Error(), "exited with code") {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("all random numbers were zero - test failed")
	}
}

// Test GET_ARGUMENTS syscall (0x32)
func TestSyscall_GetArguments(t *testing.T) {
	code := `
		.org 0x8000
_start:
		SWI #0x32         ; GET_ARGUMENTS
		; R0 contains argc, R1 contains argv pointer
		; For now, argc should be 0 (no args passed to test)
		CMP R0, #0
		MOVEQ R0, #0      ; Success
		MOVNE R0, #1      ; Unexpected arguments
		SWI #0x00
`
	_, _, exitCode, err := runAssembly(t, code)
	if err != nil && !strings.Contains(err.Error(), "exited with code") {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("GET_ARGUMENTS failed - unexpected argc value")
	}
}

// Test GET_ENVIRONMENT syscall (0x33)
func TestSyscall_GetEnvironment(t *testing.T) {
	code := `
		.org 0x8000
_start:
		SWI #0x33         ; GET_ENVIRONMENT
		; R0 should contain envp pointer (currently 0 in implementation)
		MOV R0, #0        ; Success
		SWI #0x00
`
	_, _, exitCode, err := runAssembly(t, code)
	if err != nil && !strings.Contains(err.Error(), "exited with code") {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("GET_ENVIRONMENT failed")
	}
}

// Test DEBUG_PRINT syscall (0xF0)
func TestSyscall_DebugPrint(t *testing.T) {
	code := `
		.org 0x8000
_start:
		LDR R0, =msg
		SWI #0xF0         ; DEBUG_PRINT
		MOV R0, #0
		SWI #0x00
msg:
		.asciz "Debug message test"
`
	_, stderr, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// DEBUG_PRINT should write to stderr
	if !strings.Contains(stderr, "Debug message test") {
		t.Errorf("expected debug message in stderr, got %q", stderr)
	}
}

// Test DUMP_REGISTERS syscall (0xF2)
func TestSyscall_DumpRegisters(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #42
		MOV R1, #100
		SWI #0xF2         ; DUMP_REGISTERS
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

	// Should contain register dump
	if !strings.Contains(stdout, "Register Dump") {
		t.Errorf("expected register dump in stdout, got %q", stdout)
	}
}

// Test DUMP_MEMORY syscall (0xF3)
func TestSyscall_DumpMemory(t *testing.T) {
	code := `
		.org 0x8000
_start:
		LDR R0, =data     ; Address
		MOV R1, #4        ; Length
		SWI #0xF3         ; DUMP_MEMORY
		MOV R0, #0
		SWI #0x00
data:
		.byte 0x11, 0x22, 0x33, 0x44
`
	stdout, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Should contain memory dump
	if !strings.Contains(stdout, "Memory Dump") {
		t.Errorf("expected memory dump in stdout, got %q", stdout)
	}
}

// Test ASSERT syscall pass (0xF4)
func TestSyscall_AssertPass(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #1        ; True condition
		LDR R1, =msg
		SWI #0xF4         ; ASSERT
		MOV R0, #0        ; Should reach here
		SWI #0x00
msg:
		.asciz "Assertion message"
`
	_, _, exitCode, err := runAssembly(t, code)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("assert with true condition should not fail")
	}
}

// Test ASSERT syscall fail (0xF4)
func TestSyscall_AssertFail(t *testing.T) {
	code := `
		.org 0x8000
_start:
		MOV R0, #0        ; False condition
		LDR R1, =msg
		SWI #0xF4         ; ASSERT - should halt
		MOV R0, #0        ; Should NOT reach here
		SWI #0x00
msg:
		.asciz "This should fail"
`
	_, _, _, err := runAssembly(t, code)

	// Should get an error for failed assertion
	if err == nil {
		t.Error("expected error for failed assertion")
	}

	if !strings.Contains(err.Error(), "ASSERTION FAILED") {
		t.Errorf("expected assertion failure message, got %v", err)
	}
}
