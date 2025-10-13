package integration_test

import (
	"os"
	"path/filepath"
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

// loadExpectedOutput reads the expected output file for a given example
func loadExpectedOutput(t *testing.T, baseName string) string {
	t.Helper()

	expectedPath := filepath.Join("expected_outputs", baseName+".txt")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected output %s: %v", expectedPath, err)
	}

	return string(content)
}

func TestExampleProgram_Quicksort(t *testing.T) {
	stdout, _, exitCode, err := runExampleProgram(t, "quicksort.s")
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	expected := loadExpectedOutput(t, "quicksort")
	if stdout != expected {
		t.Errorf("output mismatch\nExpected (%d bytes):\n%q\nGot (%d bytes):\n%q",
			len(expected), expected, len(stdout), stdout)
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

	expected := loadExpectedOutput(t, "division")
	if stdout != expected {
		t.Errorf("output mismatch\nExpected (%d bytes):\n%q\nGot (%d bytes):\n%q",
			len(expected), expected, len(stdout), stdout)
	}
}
