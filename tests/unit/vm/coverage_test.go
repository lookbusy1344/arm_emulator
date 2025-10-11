package vm

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestCodeCoverageBasic(t *testing.T) {
	var buf bytes.Buffer
	coverage := vm.NewCodeCoverage(&buf)

	if !coverage.Enabled {
		t.Error("Coverage should be enabled by default")
	}

	// Set code range
	coverage.SetCodeRange(0x8000, 0x8010)
	coverage.Start()

	// Record some executions
	coverage.RecordExecution(0x8000, 1)
	coverage.RecordExecution(0x8004, 2)
	coverage.RecordExecution(0x8008, 3)
	coverage.RecordExecution(0x8000, 4) // Execute 0x8000 again

	// Check executed addresses
	executed := coverage.GetExecutedAddresses()
	if len(executed) != 3 {
		t.Errorf("Expected 3 executed addresses, got %d", len(executed))
	}

	// Check unexecuted addresses
	unexecuted := coverage.GetUnexecutedAddresses()
	if len(unexecuted) != 1 {
		t.Errorf("Expected 1 unexecuted address, got %d", len(unexecuted))
	}
	if unexecuted[0] != 0x800C {
		t.Errorf("Expected unexecuted address 0x800C, got 0x%X", unexecuted[0])
	}

	// Check coverage percentage
	coveragePct := coverage.GetCoverage()
	expectedPct := 75.0 // 3 out of 4 instructions
	if coveragePct != expectedPct {
		t.Errorf("Expected coverage %.2f%%, got %.2f%%", expectedPct, coveragePct)
	}
}

func TestCodeCoverageExecutionCounts(t *testing.T) {
	var buf bytes.Buffer
	coverage := vm.NewCodeCoverage(&buf)

	coverage.SetCodeRange(0x8000, 0x8008)
	coverage.Start()

	// Execute address multiple times
	coverage.RecordExecution(0x8000, 1)
	coverage.RecordExecution(0x8000, 2)
	coverage.RecordExecution(0x8000, 3)
	coverage.RecordExecution(0x8004, 4)

	entry := coverage.GetEntry(0x8000)
	if entry == nil {
		t.Fatal("Entry for 0x8000 should exist")
	}

	if entry.ExecutionCount != 3 {
		t.Errorf("Expected execution count 3, got %d", entry.ExecutionCount)
	}

	if entry.FirstExecution != 1 {
		t.Errorf("Expected first execution at cycle 1, got %d", entry.FirstExecution)
	}

	if entry.LastExecution != 3 {
		t.Errorf("Expected last execution at cycle 3, got %d", entry.LastExecution)
	}
}

func TestCodeCoverageWithSymbols(t *testing.T) {
	var buf bytes.Buffer
	coverage := vm.NewCodeCoverage(&buf)

	// Load symbols
	symbols := map[string]uint32{
		"main": 0x8000,
		"loop": 0x8004,
		"end":  0x8008,
	}
	coverage.LoadSymbols(symbols)

	coverage.SetCodeRange(0x8000, 0x800C)
	coverage.Start()

	coverage.RecordExecution(0x8000, 1)
	coverage.RecordExecution(0x8004, 2)

	// Flush and check output contains symbol names
	err := coverage.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[main]") {
		t.Error("Output should contain [main] symbol")
	}
	if !strings.Contains(output, "[loop]") {
		t.Error("Output should contain [loop] symbol")
	}
	if !strings.Contains(output, "[end]") {
		t.Error("Output should contain [end] symbol for unexecuted address")
	}
}

func TestCodeCoverageJSON(t *testing.T) {
	var buf bytes.Buffer
	coverage := vm.NewCodeCoverage(&buf)

	coverage.SetCodeRange(0x8000, 0x8010)
	coverage.Start()

	coverage.RecordExecution(0x8000, 1)
	coverage.RecordExecution(0x8004, 2)

	err := coverage.ExportJSON(&buf)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Parse JSON
	var data map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &data)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check fields
	if data["code_start"].(float64) != 0x8000 {
		t.Error("JSON should contain code_start")
	}
	if data["code_end"].(float64) != 0x8010 {
		t.Error("JSON should contain code_end")
	}
	if data["executed_count"].(float64) != 2 {
		t.Error("JSON should show 2 executed addresses")
	}
}

func TestCodeCoverageNoRange(t *testing.T) {
	var buf bytes.Buffer
	coverage := vm.NewCodeCoverage(&buf)

	// Don't set code range
	coverage.Start()

	// Record some executions
	coverage.RecordExecution(0x8000, 1)
	coverage.RecordExecution(0x9000, 2)

	// Without range, should still track executions
	executed := coverage.GetExecutedAddresses()
	if len(executed) != 2 {
		t.Errorf("Expected 2 executed addresses, got %d", len(executed))
	}

	// Coverage should be 0 without range
	coveragePct := coverage.GetCoverage()
	if coveragePct != 0.0 {
		t.Errorf("Expected 0%% coverage without range, got %.2f%%", coveragePct)
	}
}

func TestCodeCoverageOutOfRange(t *testing.T) {
	var buf bytes.Buffer
	coverage := vm.NewCodeCoverage(&buf)

	coverage.SetCodeRange(0x8000, 0x8010)
	coverage.Start()

	// Record execution outside range
	coverage.RecordExecution(0x7000, 1) // Before range
	coverage.RecordExecution(0x8000, 2) // In range
	coverage.RecordExecution(0x9000, 3) // After range

	executed := coverage.GetExecutedAddresses()
	if len(executed) != 1 {
		t.Errorf("Expected 1 executed address (in range), got %d", len(executed))
	}
	if executed[0] != 0x8000 {
		t.Errorf("Expected executed address 0x8000, got 0x%X", executed[0])
	}
}

func TestCodeCoverageString(t *testing.T) {
	var buf bytes.Buffer
	coverage := vm.NewCodeCoverage(&buf)

	coverage.SetCodeRange(0x8000, 0x8010)
	coverage.Start()

	coverage.RecordExecution(0x8000, 1)
	coverage.RecordExecution(0x8004, 2)

	str := coverage.String()
	if !strings.Contains(str, "Code Coverage Summary") {
		t.Error("String output should contain title")
	}
	if !strings.Contains(str, "50.00%") {
		t.Error("String output should show 50% coverage")
	}
}

func TestCodeCoverageDisabled(t *testing.T) {
	var buf bytes.Buffer
	coverage := vm.NewCodeCoverage(&buf)
	coverage.Enabled = false

	coverage.SetCodeRange(0x8000, 0x8010)
	coverage.Start()

	// Record execution when disabled
	coverage.RecordExecution(0x8000, 1)

	// Should not track anything
	executed := coverage.GetExecutedAddresses()
	if len(executed) != 0 {
		t.Errorf("Expected 0 executed addresses when disabled, got %d", len(executed))
	}
}
