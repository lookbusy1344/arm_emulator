package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runExampleProgram loads and runs an example program file
func runExampleProgram(t *testing.T, filename string) (stdout string, stderr string, exitCode int32, err error) {
	t.Helper()

	examplePath := filepath.Join("..", "..", "examples", filename)
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skipf("examples/%s not found", filename)
	}

	code, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", filename, err)
	}

	return runAssembly(t, string(code))
}

func TestExampleProgram_Quicksort(t *testing.T) {
	stdout, _, exitCode, err := runExampleProgram(t, "quicksort.s")
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Verify expected output
	expectedOutputs := []string{
		"Quicksort Algorithm",
		"Original array:",
		"Sorted array:",
		"Verification: Array is correctly sorted!",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("expected output to contain %q\nActual output:\n%s", expected, stdout)
		}
	}

	// Verify array values appear in sorted order in output
	if !strings.Contains(stdout, "4, 8, 12, 14, 17") {
		t.Errorf("expected sorted array to contain '4, 8, 12, 14, 17' in output:\n%s", stdout)
	}
}

func TestExampleProgram_Division(t *testing.T) {
	stdout, _, exitCode, err := runExampleProgram(t, "division.s")
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Verify test results
	expectedResults := []string{
		"100 / 7 = 14 remainder 2",
		"1000 / 17 = 58 remainder 14",
		"144 / 12 = 12 remainder 0",
		"42 / 1 = 42 remainder 0",
		"5 / 10 = 0 remainder 5",
		"0 / 5 = 0 remainder 0",
	}

	for _, expected := range expectedResults {
		if !strings.Contains(stdout, expected) {
			t.Errorf("expected output to contain %q\nActual output:\n%s", expected, stdout)
		}
	}
}
