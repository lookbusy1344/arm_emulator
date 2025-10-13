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
	}{
		{
			name:           "Quicksort",
			programFile:    "quicksort.s",
			expectedOutput: "quicksort.txt",
		},
		{
			name:           "Division",
			programFile:    "division.s",
			expectedOutput: "division.txt",
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

			stdout, _, exitCode, err := runAssembly(t, string(code))
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
