package integration_test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExamplePrograms runs integration tests for example programs by comparing
// their output against expected output files
func TestExamplePrograms(t *testing.T) {
	tests := []struct {
		name           string // Test name
		programFile    string // Assembly file in examples/ directory
		expectedOutput string // Expected output file in expected_outputs/ directory
		stdin          string // Optional stdin input for the program
	}{
		{
			name:           "Hello",
			programFile:    "hello.s",
			expectedOutput: "hello.txt",
		},
		{
			name:           "Loops",
			programFile:    "loops.s",
			expectedOutput: "loops.txt",
		},
		{
			name:           "MatrixMultiply",
			programFile:    "matrix_multiply.s",
			expectedOutput: "matrix_multiply.txt",
		},
		{
			name:           "MemoryStress",
			programFile:    "memory_stress.s",
			expectedOutput: "memory_stress.txt",
		},
		{
			name:           "GCD",
			programFile:    "gcd.s",
			expectedOutput: "gcd.txt",
			stdin:          "48\n18\n",
		},
		{
			name:           "StateMachine",
			programFile:    "state_machine.s",
			expectedOutput: "state_machine.txt",
		},
		{
			name:           "StringReverse",
			programFile:    "string_reverse.s",
			expectedOutput: "string_reverse.txt",
			stdin:          "Hello World\n",
		},
		{
			name:           "Strings",
			programFile:    "strings.s",
			expectedOutput: "strings.txt",
		},
		{
			name:           "Stack",
			programFile:    "stack.s",
			expectedOutput: "stack.txt",
		},
		{
			name:           "NestedCalls",
			programFile:    "nested_calls.s",
			expectedOutput: "nested_calls.txt",
		},
		{
			name:           "HashTable",
			programFile:    "hash_table.s",
			expectedOutput: "hash_table.txt",
		},
		{
			name:           "ConstExpressions",
			programFile:    "const_expressions.s",
			expectedOutput: "const_expressions.txt",
		},
		{
			name:           "RecursiveFactorial",
			programFile:    "recursive_factorial.s",
			expectedOutput: "recursive_factorial.txt",
		},
		{
			name:           "RecursiveFibonacci",
			programFile:    "recursive_fib.s",
			expectedOutput: "recursive_fib.txt",
		},
		{
			name:           "SieveOfEratosthenes",
			programFile:    "sieve_of_eratosthenes.s",
			expectedOutput: "sieve_of_eratosthenes.txt",
		},
		{
			name:           "StandaloneLabels",
			programFile:    "standalone_labels.s",
			expectedOutput: "standalone_labels.txt",
		},
		{
			name:           "XORCipher",
			programFile:    "xor_cipher.s",
			expectedOutput: "xor_cipher.txt",
		},
		{
			name:           "FileIO",
			programFile:    "file_io.s",
			expectedOutput: "file_io.txt",
		},
		{
			name:           "MultiPrecisionArith",
			programFile:    "multi_precision_arith.s",
			expectedOutput: "multi_precision_arith.txt",
		},
		{
			name:           "TaskScheduler",
			programFile:    "task_scheduler.s",
			expectedOutput: "task_scheduler.txt",
		},
		{
			name:           "ADRDemo",
			programFile:    "adr_demo.s",
			expectedOutput: "adr_demo.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load and run the example program
			examplePath := filepath.Join("..", "..", "examples", tt.programFile)
			if _, err := os.Stat(examplePath); os.IsNotExist(err) {
				t.Skipf("examples/%s not found", tt.programFile)
			}

			code, err := os.ReadFile(examplePath)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.programFile, err)
			}

			stdout, _, exitCode, err := runAssemblyWithInput(t, string(code), tt.stdin)
			if err != nil {
				t.Fatalf("execution failed: %v", err)
			}

			if exitCode != 0 {
				t.Errorf("expected exit code 0, got %d", exitCode)
			}

			// Load expected output
			expectedPath := filepath.Join("expected_outputs", tt.expectedOutput)
			expectedBytes, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("failed to read expected output %s: %v", expectedPath, err)
			}
			expected := string(expectedBytes)

			// Compare output
			if stdout != expected {
				t.Errorf("output mismatch\nExpected (%d bytes):\n%q\nGot (%d bytes):\n%q",
					len(expected), expected, len(stdout), stdout)
			}
		})
	}
}
