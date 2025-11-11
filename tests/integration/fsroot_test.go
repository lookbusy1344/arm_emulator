package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFilesystemRestrictionAllowedAccess tests that files within fsroot can be accessed
func TestFilesystemRestrictionAllowedAccess(t *testing.T) {
	// Build the emulator first
	buildCmd := exec.Command("go", "build", "-o", "arm-emulator-test")
	buildCmd.Dir = "../.."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build emulator: %v", err)
	}
	defer os.Remove("../../arm-emulator-test")

	// Get absolute paths
	testDir, err := filepath.Abs("fsroot_test/allowed")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	asmFile, err := filepath.Abs("fsroot_test/allowed_access.s")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Run the emulator with fsroot restriction
	cmd := exec.Command("../../arm-emulator-test", "-fsroot", testDir, asmFile)
	output, err := cmd.CombinedOutput()

	// Should succeed (exit code 0)
	if err != nil {
		t.Errorf("Expected success, got error: %v\nOutput: %s", err, string(output))
	}

	// Verify success message
	if !strings.Contains(string(output), "File access succeeded") {
		t.Errorf("Expected success message, got: %s", string(output))
	}
}

// TestFilesystemRestrictionEscapeAttempt tests that escape attempts halt the VM
func TestFilesystemRestrictionEscapeAttempt(t *testing.T) {
	// Build the emulator first
	buildCmd := exec.Command("go", "build", "-o", "arm-emulator-test")
	buildCmd.Dir = "../.."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build emulator: %v", err)
	}
	defer os.Remove("../../arm-emulator-test")

	// Get absolute paths
	testDir, err := filepath.Abs("fsroot_test/allowed")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	asmFile, err := filepath.Abs("fsroot_test/escape_attempt.s")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Run the emulator with fsroot restriction
	cmd := exec.Command("../../arm-emulator-test", "-fsroot", testDir, asmFile)
	output, err := cmd.CombinedOutput()

	// Should fail with exit code != 0
	if err == nil {
		t.Errorf("Expected error for escape attempt, but succeeded\nOutput: %s", string(output))
	}

	// Verify error message mentions filesystem access
	outputStr := string(output)
	if !strings.Contains(outputStr, "filesystem access denied") &&
		!strings.Contains(outputStr, "..") {
		t.Errorf("Expected filesystem access error message, got: %s", outputStr)
	}
}
