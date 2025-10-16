package integration_test

import (
	"os"
	"strings"
	"testing"
)

func TestRegisterTrace_Basic(t *testing.T) {
	// Create a simple test program
	code := `.org 0x8000
start:
    MOV R0, #1
    MOV R1, #10
    MOV R2, #20
    ADD R3, R1, R2
    MOV R0, #0
    SWI #0x00
`

	progPath := createTestProgram(t, code)
	defer os.Remove(progPath)

	// Create temp file for register trace output
	traceFile, err := os.CreateTemp("", "register_trace_*.txt")
	if err != nil {
		t.Fatalf("Failed to create trace file: %v", err)
	}
	traceFile.Close()
	tracePath := traceFile.Name()
	defer os.Remove(tracePath)

	// Run with --register-trace flag
	_, stderr, exitCode := runEmulatorWithFlags(t, progPath,
		"--register-trace",
		"--register-trace-file", tracePath)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	// Read trace file
	traceData, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("Failed to read trace file: %v", err)
	}

	output := string(traceData)

	// Verify trace output contains expected information
	if !strings.Contains(output, "Register Access Pattern Analysis") {
		t.Error("Missing header in trace output")
	}

	if !strings.Contains(output, "Total Reads:") {
		t.Error("Missing total reads in trace output")
	}

	if !strings.Contains(output, "Total Writes:") {
		t.Error("Missing total writes in trace output")
	}

	if !strings.Contains(output, "Hot Registers") {
		t.Error("Missing hot registers section in trace output")
	}

	// Verify that R0, R1, R2, R3 were used (from our program)
	if !strings.Contains(output, "R0") {
		t.Error("R0 should appear in trace output")
	}
	if !strings.Contains(output, "R1") {
		t.Error("R1 should appear in trace output")
	}
	if !strings.Contains(output, "R2") {
		t.Error("R2 should appear in trace output")
	}
	if !strings.Contains(output, "R3") {
		t.Error("R3 should appear in trace output")
	}
}

func TestRegisterTrace_JSONOutput(t *testing.T) {
	// Create a simple test program
	code := `.org 0x8000
start:
    MOV R0, #10
    MOV R1, #20
    ADD R2, R0, R1
    MOV R0, #0
    SWI #0x00
`

	progPath := createTestProgram(t, code)
	defer os.Remove(progPath)

	// Create temp file for JSON register trace output
	traceFile, err := os.CreateTemp("", "register_trace_*.json")
	if err != nil {
		t.Fatalf("Failed to create trace file: %v", err)
	}
	traceFile.Close()
	tracePath := traceFile.Name()
	defer os.Remove(tracePath)

	// Run with --register-trace flag and JSON format
	_, stderr, exitCode := runEmulatorWithFlags(t, progPath,
		"--register-trace",
		"--register-trace-file", tracePath,
		"--register-trace-format", "json")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	// Read trace file
	traceData, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("Failed to read trace file: %v", err)
	}

	output := string(traceData)

	// Verify JSON contains expected fields
	if !strings.Contains(output, "\"total_reads\"") {
		t.Error("JSON missing total_reads field")
	}
	if !strings.Contains(output, "\"total_writes\"") {
		t.Error("JSON missing total_writes field")
	}
	if !strings.Contains(output, "\"register_stats\"") {
		t.Error("JSON missing register_stats field")
	}
	if !strings.Contains(output, "\"hot_registers\"") {
		t.Error("JSON missing hot_registers field")
	}
	if !strings.Contains(output, "\"unused_registers\"") {
		t.Error("JSON missing unused_registers field")
	}
	if !strings.Contains(output, "\"read_before_write\"") {
		t.Error("JSON missing read_before_write field")
	}
}
