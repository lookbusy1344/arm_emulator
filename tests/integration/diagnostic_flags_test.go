package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to create a temporary test program
func createTestProgram(t *testing.T, code string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test_*.s")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(code); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	tmpFile.Close()
	return tmpFile.Name()
}

// Helper to run the emulator binary with flags
func runEmulatorWithFlags(t *testing.T, progPath string, flags ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	// Build the emulator if needed
	binaryPath := filepath.Join("..", "..", "arm-emulator")

	// Prepare command
	args := append(flags, progPath)
	cmd := exec.Command(binaryPath, args...)

	// Capture output
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Run
	err := cmd.Run()

	// Get exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("Failed to run emulator: %v", err)
		}
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// TestMemTraceFlag tests the --mem-trace and --mem-trace-file flags
func TestMemTraceFlag(t *testing.T) {
	code := `.org 0x8000
start:
    SUB SP, SP, #16
    MOV R1, #42
    STR R1, [SP]
    STR R1, [SP, #4]
    LDR R2, [SP]
    LDR R3, [SP, #4]
    ADD SP, SP, #16
    MOV R0, #0
    SWI #0x00
`

	progPath := createTestProgram(t, code)
	defer os.Remove(progPath)

	// Create temp file for trace output
	traceFile, err := os.CreateTemp("", "trace_*.log")
	if err != nil {
		t.Fatalf("Failed to create trace file: %v", err)
	}
	traceFile.Close()
	tracePath := traceFile.Name()
	defer os.Remove(tracePath)

	// Run with --mem-trace flag
	_, stderr, exitCode := runEmulatorWithFlags(t, progPath,
		"--mem-trace",
		"--mem-trace-file", tracePath)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	// Read trace file
	traceData, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("Failed to read trace file: %v", err)
	}

	traceOutput := string(traceData)

	// Verify trace output
	if traceOutput == "" {
		t.Fatal("Trace file is empty")
	}

	// Check for READ operations
	if !strings.Contains(traceOutput, "[READ ]") {
		t.Error("Trace should contain [READ ] operations")
	}

	// Check for WRITE operations
	if !strings.Contains(traceOutput, "[WRITE]") {
		t.Error("Trace should contain [WRITE] operations")
	}

	// Check for WORD size indicator
	if !strings.Contains(traceOutput, "(WORD)") {
		t.Error("Trace should indicate (WORD) size")
	}

	// Count entries - should have at least 4 operations (2 STR + 2 LDR)
	lines := strings.Split(strings.TrimSpace(traceOutput), "\n")
	if len(lines) < 4 {
		t.Errorf("Expected at least 4 trace entries, got %d", len(lines))
	}

	t.Logf("Memory trace generated %d entries", len(lines))
}

// TestCoverageFlag tests the --coverage flag
func TestCoverageFlag(t *testing.T) {
	code := `.org 0x8000
start:
    MOV R0, #10
    MOV R1, #20
    CMP R0, R1
    BEQ skip
    ADD R2, R0, R1
skip:
    MOV R0, #0
    SWI #0x00
`

	progPath := createTestProgram(t, code)
	defer os.Remove(progPath)

	// Create temp file for coverage output
	coverageFile, err := os.CreateTemp("", "coverage_*.txt")
	if err != nil {
		t.Fatalf("Failed to create coverage file: %v", err)
	}
	coverageFile.Close()
	coveragePath := coverageFile.Name()
	defer os.Remove(coveragePath)

	// Run with --coverage flag
	_, stderr, exitCode := runEmulatorWithFlags(t, progPath,
		"--coverage",
		"--coverage-file", coveragePath,
		"--coverage-format", "text")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	// Read coverage file
	coverageData, err := os.ReadFile(coveragePath)
	if err != nil {
		t.Fatalf("Failed to read coverage file: %v", err)
	}

	coverageOutput := string(coverageData)

	// Verify coverage output
	if coverageOutput == "" {
		t.Fatal("Coverage file is empty")
	}

	// Check for expected content
	if !strings.Contains(coverageOutput, "Code Coverage") {
		t.Error("Coverage output should contain 'Code Coverage'")
	}

	// Should show coverage percentage
	if !strings.Contains(coverageOutput, "%") {
		t.Error("Coverage output should show percentage")
	}

	// Should show executed instructions
	if !strings.Contains(coverageOutput, "Executed") {
		t.Error("Coverage output should show Executed instructions")
	}

	t.Logf("Coverage output generated successfully")
}

// TestCoverageFlagJSON tests the --coverage flag with JSON format
func TestCoverageFlagJSON(t *testing.T) {
	code := `.org 0x8000
start:
    MOV R0, #5
    ADD R1, R0, #3
    MOV R0, #0
    SWI #0x00
`

	progPath := createTestProgram(t, code)
	defer os.Remove(progPath)

	// Create temp file for coverage output
	coverageFile, err := os.CreateTemp("", "coverage_*.json")
	if err != nil {
		t.Fatalf("Failed to create coverage file: %v", err)
	}
	coverageFile.Close()
	coveragePath := coverageFile.Name()
	defer os.Remove(coveragePath)

	// Run with --coverage flag and JSON format
	_, stderr, exitCode := runEmulatorWithFlags(t, progPath,
		"--coverage",
		"--coverage-file", coveragePath,
		"--coverage-format", "json")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	// Read coverage file
	coverageData, err := os.ReadFile(coveragePath)
	if err != nil {
		t.Fatalf("Failed to read coverage file: %v", err)
	}

	coverageOutput := string(coverageData)

	// Verify JSON output
	if coverageOutput == "" {
		t.Fatal("Coverage file is empty")
	}

	// Check for JSON structure
	if !strings.Contains(coverageOutput, "{") {
		t.Error("Coverage output should be JSON format")
	}

	if !strings.Contains(coverageOutput, "code_start") {
		t.Error("JSON should contain 'code_start' field")
	}

	if !strings.Contains(coverageOutput, "coverage_percent") {
		t.Error("JSON should contain 'coverage_percent' field")
	}

	t.Logf("JSON coverage output generated successfully")
}

// TestStackTraceFlag tests the --stack-trace flag
func TestStackTraceFlag(t *testing.T) {
	code := `.org 0x8000
start:
    SUB SP, SP, #8
    MOV R0, #100
    STR R0, [SP]
    STR R0, [SP, #4]
    LDR R1, [SP]
    LDR R2, [SP, #4]
    ADD SP, SP, #8
    MOV R0, #0
    SWI #0x00
`

	progPath := createTestProgram(t, code)
	defer os.Remove(progPath)

	// Create temp file for stack trace output
	stackTraceFile, err := os.CreateTemp("", "stack_trace_*.txt")
	if err != nil {
		t.Fatalf("Failed to create stack trace file: %v", err)
	}
	stackTraceFile.Close()
	stackTracePath := stackTraceFile.Name()
	defer os.Remove(stackTracePath)

	// Run with --stack-trace flag
	_, stderr, exitCode := runEmulatorWithFlags(t, progPath,
		"--stack-trace",
		"--stack-trace-file", stackTracePath,
		"--stack-trace-format", "text")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	// Read stack trace file
	stackTraceData, err := os.ReadFile(stackTracePath)
	if err != nil {
		t.Fatalf("Failed to read stack trace file: %v", err)
	}

	stackTraceOutput := string(stackTraceData)

	// Verify stack trace output
	if stackTraceOutput == "" {
		t.Fatal("Stack trace file is empty")
	}

	// Check for expected content
	if !strings.Contains(stackTraceOutput, "Stack") {
		t.Error("Stack trace output should contain 'Stack'")
	}

	// Should show stack operations
	if !strings.Contains(stackTraceOutput, "SP") {
		t.Error("Stack trace should reference SP (stack pointer)")
	}

	t.Logf("Stack trace output generated successfully")
}

// TestStackTraceFlagJSON tests the --stack-trace flag with JSON format
func TestStackTraceFlagJSON(t *testing.T) {
	code := `.org 0x8000
start:
    SUB SP, SP, #16
    MOV R0, #42
    STR R0, [SP]
    LDR R1, [SP]
    ADD SP, SP, #16
    MOV R0, #0
    SWI #0x00
`

	progPath := createTestProgram(t, code)
	defer os.Remove(progPath)

	// Create temp file for stack trace output
	stackTraceFile, err := os.CreateTemp("", "stack_trace_*.json")
	if err != nil {
		t.Fatalf("Failed to create stack trace file: %v", err)
	}
	stackTraceFile.Close()
	stackTracePath := stackTraceFile.Name()
	defer os.Remove(stackTracePath)

	// Run with --stack-trace flag and JSON format
	_, stderr, exitCode := runEmulatorWithFlags(t, progPath,
		"--stack-trace",
		"--stack-trace-file", stackTracePath,
		"--stack-trace-format", "json")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	// Read stack trace file
	stackTraceData, err := os.ReadFile(stackTracePath)
	if err != nil {
		t.Fatalf("Failed to read stack trace file: %v", err)
	}

	stackTraceOutput := string(stackTraceData)

	// Verify JSON output
	if stackTraceOutput == "" {
		t.Fatal("Stack trace file is empty")
	}

	// Check for JSON structure
	if !strings.Contains(stackTraceOutput, "{") {
		t.Error("Stack trace output should be JSON format")
	}

	if !strings.Contains(stackTraceOutput, "stack_base") || !strings.Contains(stackTraceOutput, "stack_size") {
		t.Error("JSON should contain 'stack_base' and 'stack_size' fields")
	}

	t.Logf("JSON stack trace output generated successfully")
}

// TestFlagTraceFlag tests the --flag-trace flag
func TestFlagTraceFlag(t *testing.T) {
	code := `.org 0x8000
start:
    MOV R0, #10
    MOV R1, #5
    SUBS R2, R0, R1
    ADDS R3, R0, R1
    CMP R0, R1
    MOV R0, #0
    SWI #0x00
`

	progPath := createTestProgram(t, code)
	defer os.Remove(progPath)

	// Create temp file for flag trace output
	flagTraceFile, err := os.CreateTemp("", "flag_trace_*.txt")
	if err != nil {
		t.Fatalf("Failed to create flag trace file: %v", err)
	}
	flagTraceFile.Close()
	flagTracePath := flagTraceFile.Name()
	defer os.Remove(flagTracePath)

	// Run with --flag-trace flag
	_, stderr, exitCode := runEmulatorWithFlags(t, progPath,
		"--flag-trace",
		"--flag-trace-file", flagTracePath,
		"--flag-trace-format", "text")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	// Read flag trace file
	flagTraceData, err := os.ReadFile(flagTracePath)
	if err != nil {
		t.Fatalf("Failed to read flag trace file: %v", err)
	}

	flagTraceOutput := string(flagTraceData)

	// Verify flag trace output
	if flagTraceOutput == "" {
		t.Fatal("Flag trace file is empty")
	}

	// Check for expected content - should show flag changes
	if !strings.Contains(flagTraceOutput, "Flag") && !strings.Contains(flagTraceOutput, "flag") {
		t.Error("Flag trace output should contain flag information")
	}

	// Should reference flag bits (N, Z, C, V)
	hasFlags := strings.Contains(flagTraceOutput, "N") ||
		strings.Contains(flagTraceOutput, "Z") ||
		strings.Contains(flagTraceOutput, "C") ||
		strings.Contains(flagTraceOutput, "V")

	if !hasFlags {
		t.Error("Flag trace should reference N, Z, C, or V flags")
	}

	t.Logf("Flag trace output generated successfully")
}

// TestFlagTraceFlagJSON tests the --flag-trace flag with JSON format
func TestFlagTraceFlagJSON(t *testing.T) {
	code := `.org 0x8000
start:
    MOV R0, #15
    SUBS R1, R0, #10
    ADDS R2, R0, #5
    MOV R0, #0
    SWI #0x00
`

	progPath := createTestProgram(t, code)
	defer os.Remove(progPath)

	// Create temp file for flag trace output
	flagTraceFile, err := os.CreateTemp("", "flag_trace_*.json")
	if err != nil {
		t.Fatalf("Failed to create flag trace file: %v", err)
	}
	flagTraceFile.Close()
	flagTracePath := flagTraceFile.Name()
	defer os.Remove(flagTracePath)

	// Run with --flag-trace flag and JSON format
	_, stderr, exitCode := runEmulatorWithFlags(t, progPath,
		"--flag-trace",
		"--flag-trace-file", flagTracePath,
		"--flag-trace-format", "json")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	// Read flag trace file
	flagTraceData, err := os.ReadFile(flagTracePath)
	if err != nil {
		t.Fatalf("Failed to read flag trace file: %v", err)
	}

	flagTraceOutput := string(flagTraceData)

	// Verify JSON output
	if flagTraceOutput == "" {
		t.Fatal("Flag trace file is empty")
	}

	// Check for JSON structure
	if !strings.Contains(flagTraceOutput, "{") {
		t.Error("Flag trace output should be JSON format")
	}

	// Check for key fields in JSON output
	if !strings.Contains(flagTraceOutput, "total_changes") && !strings.Contains(flagTraceOutput, "entries") {
		t.Error("JSON should contain flag trace data fields")
	}

	t.Logf("JSON flag trace output generated successfully")
}

// TestMultipleDiagnosticFlags tests using multiple diagnostic flags together
func TestMultipleDiagnosticFlags(t *testing.T) {
	code := `.org 0x8000
start:
    SUB SP, SP, #8
    MOV R0, #25
    STR R0, [SP]
    SUBS R1, R0, #10
    LDR R2, [SP]
    ADD SP, SP, #8
    MOV R0, #0
    SWI #0x00
`

	progPath := createTestProgram(t, code)
	defer os.Remove(progPath)

	// Create temp files for all outputs
	memTraceFile, _ := os.CreateTemp("", "mem_*.log")
	memTraceFile.Close()
	memTracePath := memTraceFile.Name()
	defer os.Remove(memTracePath)

	coverageFile, _ := os.CreateTemp("", "cov_*.txt")
	coverageFile.Close()
	coveragePath := coverageFile.Name()
	defer os.Remove(coveragePath)

	stackTraceFile, _ := os.CreateTemp("", "stack_*.txt")
	stackTraceFile.Close()
	stackTracePath := stackTraceFile.Name()
	defer os.Remove(stackTracePath)

	flagTraceFile, _ := os.CreateTemp("", "flags_*.txt")
	flagTraceFile.Close()
	flagTracePath := flagTraceFile.Name()
	defer os.Remove(flagTracePath)

	// Run with all diagnostic flags enabled
	_, stderr, exitCode := runEmulatorWithFlags(t, progPath,
		"--mem-trace", "--mem-trace-file", memTracePath,
		"--coverage", "--coverage-file", coveragePath,
		"--stack-trace", "--stack-trace-file", stackTracePath,
		"--flag-trace", "--flag-trace-file", flagTracePath)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	// Verify all output files were created and are non-empty
	files := map[string]string{
		"memory trace": memTracePath,
		"coverage":     coveragePath,
		"stack trace":  stackTracePath,
		"flag trace":   flagTracePath,
	}

	for name, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read %s file: %v", name, err)
			continue
		}

		if len(data) == 0 {
			t.Errorf("%s file is empty", name)
		} else {
			t.Logf("%s file generated successfully (%d bytes)", name, len(data))
		}
	}
}
