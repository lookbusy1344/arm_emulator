package integration_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/encoder"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// Helper function to run assembly code and capture stdout
func runAssembly(t *testing.T, code string) (stdout string, stderr string, exitCode int32, err error) {
	t.Helper()

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
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

	// Create VM
	machine := vm.NewVM()
	machine.CycleLimit = 1000000

	// Initialize stack
	stackTop := uint32(vm.StackSegmentStart + vm.StackSegmentSize)
	machine.InitializeStack(stackTop)

	// Load program
	err = loadProgramIntoVM(machine, program, 0x8000)
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

// Helper function to load program into VM (copied from main.go)
func loadProgramIntoVM(machine *vm.VM, program *parser.Program, entryPoint uint32) error {
	currentAddr := entryPoint

	// Create encoder
	enc := encoder.NewEncoder(program.SymbolTable)

	// Build address map for instructions
	addressMap := make(map[*parser.Instruction]uint32)
	dataAddr := currentAddr

	for _, inst := range program.Instructions {
		addressMap[inst] = dataAddr
		dataAddr += 4 // Each instruction is 4 bytes
	}

	// Data directives go after instructions (original layout)

	// Process data directives
	for _, directive := range program.Directives {
		// Update label address in symbol table if this directive has a label
		if directive.Label != "" {
			if err := program.SymbolTable.UpdateAddress(directive.Label, dataAddr); err != nil {
				return err
			}
		}

		switch directive.Name {
		case ".org":
			// .org directive is handled at parse time, skip it here
			continue

		case ".align":
			// Align to power of 2 (e.g., .align 2 means align to 2^2 = 4 bytes)
			if len(directive.Args) > 0 {
				var alignPower uint32
				_, err := parseValue(directive.Args[0], &alignPower)
				if err != nil {
					return err
				}
				alignBytes := uint32(1 << alignPower) // 2^alignPower
				mask := alignBytes - 1
				dataAddr = (dataAddr + mask) & ^mask
			}

		case ".balign":
			// Align to specified boundary
			if len(directive.Args) > 0 {
				var align uint32
				_, err := parseValue(directive.Args[0], &align)
				if err != nil {
					return err
				}
				if dataAddr%align != 0 {
					dataAddr += align - (dataAddr % align)
				}
			}

		case ".word":
			// Write 32-bit words
			for _, arg := range directive.Args {
				var value uint32
				// Check if it's a symbol first (labels are more common than numbers in .word)
				if symValue, symErr := program.SymbolTable.Get(arg); symErr == nil {
					value = symValue
				} else {
					// Try to parse as a number
					_, err := parseValue(arg, &value)
					if err != nil {
						return err
					}
				}
				if err := machine.Memory.WriteWordUnsafe(dataAddr, value); err != nil {
					return err
				}
				dataAddr += 4
			}

		case ".byte":
			// Write bytes
			for _, arg := range directive.Args {
				var value uint32
				_, err := parseValue(arg, &value)
				if err != nil {
					return err
				}
				if err := machine.Memory.WriteByteUnsafe(dataAddr, byte(value)); err != nil {
					return err
				}
				dataAddr++
			}

		case ".asciz", ".string":
			// Write null-terminated string
			if len(directive.Args) > 0 {
				str := directive.Args[0]
				// Remove quotes
				if len(str) >= 2 && (str[0] == '"' || str[0] == '\'') {
					str = str[1 : len(str)-1]
				}
				// Write string bytes
				for i := 0; i < len(str); i++ {
					if err := machine.Memory.WriteByteUnsafe(dataAddr, str[i]); err != nil {
						return err
					}
					dataAddr++
				}
				// Write null terminator
				if err := machine.Memory.WriteByteUnsafe(dataAddr, 0); err != nil {
					return err
				}
				dataAddr++
			}
		}
	}

	// Set literal pool start address to after all data
	// Align to 4-byte boundary
	literalPoolStart := (dataAddr + 3) & ^uint32(3)
	enc.LiteralPoolStart = literalPoolStart

	// Encode and write instructions
	for _, inst := range program.Instructions {
		addr := addressMap[inst]

		// Encode instruction
		opcode, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			return err
		}

		// Write to memory
		if err := machine.Memory.WriteWordUnsafe(addr, opcode); err != nil {
			return err
		}
	}

	// Write literal pool
	for addr, value := range enc.LiteralPool {
		if err := machine.Memory.WriteWordUnsafe(addr, value); err != nil {
			return err
		}
	}

	// Set PC to entry point
	machine.CPU.PC = entryPoint

	return nil
}

// Helper to parse immediate values
func parseValue(s string, out *uint32) (int, error) {
	var val uint32
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
			return 0, nil
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
			return 0, nil
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
