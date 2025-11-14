package vm

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// Stack Trace Tests
//
// These tests verify the StackTrace component's diagnostic and analysis capabilities.
// Tests call RecordPush/RecordPop/RecordSPMove directly to test the StackTrace's
// independent detection logic, bypassing the CPU's proactive bounds validation.
//
// In production:
// - SetSPWithTrace validates SP bounds BEFORE setting SP
// - Invalid SP values are rejected and never reach StackTrace.Record* functions
// - StackTrace's overflow/underflow detection serves as analysis/diagnostic layer
//
// In these tests:
// - We call StackTrace.Record* methods directly to test detection in isolation
// - This allows testing the StackTrace component's capabilities independently
// - Tests intentionally use invalid SP values to verify detection logic works

func TestStackTraceBasic(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)

	if !stackTrace.Enabled {
		t.Error("Stack trace should be enabled by default")
	}

	stackTrace.Start(0x50000)

	// Record a push operation
	stackTrace.RecordPush(1, 0x8000, 0x50000, 0x4FFFC, 0x12345678, 0x4FFFC, "R0")

	entries := stackTrace.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Operation != vm.StackPush {
		t.Errorf("Expected PUSH operation, got %v", entry.Operation)
	}
	if entry.OldSP != 0x50000 {
		t.Errorf("Expected old SP 0x50000, got 0x%X", entry.OldSP)
	}
	if entry.NewSP != 0x4FFFC {
		t.Errorf("Expected new SP 0x4FFFC, got 0x%X", entry.NewSP)
	}
	if entry.Value != 0x12345678 {
		t.Errorf("Expected value 0x12345678, got 0x%X", entry.Value)
	}
}

func TestStackTracePushPop(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)
	stackTrace.Start(0x50000)

	// Push
	stackTrace.RecordPush(1, 0x8000, 0x50000, 0x4FFFC, 0x11111111, 0x4FFFC, "R1")
	stackTrace.RecordPush(2, 0x8004, 0x4FFFC, 0x4FFF8, 0x22222222, 0x4FFF8, "R2")

	// Pop
	stackTrace.RecordPop(3, 0x8008, 0x4FFF8, 0x4FFFC, 0x22222222, 0x4FFF8, "R2")
	stackTrace.RecordPop(4, 0x800C, 0x4FFFC, 0x50000, 0x11111111, 0x4FFFC, "R1")

	entries := stackTrace.GetEntries()
	if len(entries) != 4 {
		t.Fatalf("Expected 4 entries, got %d", len(entries))
	}

	// Check push entries
	if entries[0].Operation != vm.StackPush {
		t.Error("First entry should be PUSH")
	}
	if entries[1].Operation != vm.StackPush {
		t.Error("Second entry should be PUSH")
	}

	// Check pop entries
	if entries[2].Operation != vm.StackPop {
		t.Error("Third entry should be POP")
	}
	if entries[3].Operation != vm.StackPop {
		t.Error("Fourth entry should be POP")
	}

	// Check final stack depth
	depth := stackTrace.GetStackDepth()
	if depth != 0 {
		t.Errorf("Expected stack depth 0 after balanced push/pop, got %d", depth)
	}
}

func TestStackTraceMaxUsage(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)
	stackTrace.Start(0x50000)

	// Push multiple values
	stackTrace.RecordPush(1, 0x8000, 0x50000, 0x4FFFC, 0x11111111, 0x4FFFC, "R1")
	stackTrace.RecordPush(2, 0x8004, 0x4FFFC, 0x4FFF8, 0x22222222, 0x4FFF8, "R2")
	stackTrace.RecordPush(3, 0x8008, 0x4FFF8, 0x4FFF4, 0x33333333, 0x4FFF4, "R3")

	maxUsage := stackTrace.GetStackUsage()
	if maxUsage != 12 {
		t.Errorf("Expected max usage 12 bytes, got %d", maxUsage)
	}

	// Pop one value
	stackTrace.RecordPop(4, 0x800C, 0x4FFF4, 0x4FFF8, 0x33333333, 0x4FFF4, "R3")

	// Max usage should still be 12
	maxUsage = stackTrace.GetStackUsage()
	if maxUsage != 12 {
		t.Errorf("Expected max usage still 12 bytes, got %d", maxUsage)
	}

	// Current depth should be 8
	depth := stackTrace.GetStackDepth()
	if depth != 8 {
		t.Errorf("Expected current depth 8 bytes, got %d", depth)
	}
}

func TestStackTraceOverflow(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)
	stackTrace.Start(0x50000)

	// Push past the stack limit (below StackTop)
	// NOTE: This test calls RecordPush directly to test StackTrace's detection logic.
	// In production, SetSPWithTrace validates bounds proactively, so invalid SP values
	// never reach RecordPush. This test verifies the StackTrace component's independent
	// detection capabilities for diagnostic/analysis purposes.
	stackTrace.RecordPush(1, 0x8000, 0x40004, 0x3FFFC, 0x11111111, 0x3FFFC, "R1")

	if !stackTrace.HasOverflow() {
		t.Error("Should detect stack overflow")
	}

	entries := stackTrace.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
}

func TestStackTraceUnderflow(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)
	stackTrace.Start(0x50000)

	// Pop past the stack base (above StackBase)
	// NOTE: This test calls RecordPop directly to test StackTrace's detection logic.
	// In production, SetSPWithTrace validates bounds proactively, so invalid SP values
	// never reach RecordPop. This test verifies the StackTrace component's independent
	// detection capabilities for diagnostic/analysis purposes.
	stackTrace.RecordPop(1, 0x8000, 0x50000, 0x50004, 0x11111111, 0x50000, "R1")

	if !stackTrace.HasUnderflow() {
		t.Error("Should detect stack underflow")
	}
}

func TestStackTraceSPMove(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)
	stackTrace.Start(0x50000)

	// Record SP modification
	stackTrace.RecordSPMove(1, 0x8000, 0x50000, 0x4FFF0)

	entries := stackTrace.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Operation != vm.StackMove {
		t.Errorf("Expected MOVE operation, got %v", entry.Operation)
	}
	if entry.OldSP != 0x50000 {
		t.Errorf("Expected old SP 0x50000, got 0x%X", entry.OldSP)
	}
	if entry.NewSP != 0x4FFF0 {
		t.Errorf("Expected new SP 0x4FFF0, got 0x%X", entry.NewSP)
	}
	if entry.Size != 16 {
		t.Errorf("Expected size 16 bytes, got %d", entry.Size)
	}
}

func TestStackTraceJSON(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)
	stackTrace.Start(0x50000)

	stackTrace.RecordPush(1, 0x8000, 0x50000, 0x4FFFC, 0x12345678, 0x4FFFC, "R0")

	err := stackTrace.ExportJSON(&buf)
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
	if data["stack_base"].(float64) != 0x50000 {
		t.Error("JSON should contain stack_base")
	}
	if data["stack_top"].(float64) != 0x40000 {
		t.Error("JSON should contain stack_top")
	}
	if data["total_pushes"].(float64) != 1 {
		t.Error("JSON should show 1 push")
	}
}

func TestStackTraceFlush(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)
	stackTrace.Start(0x50000)

	stackTrace.RecordPush(1, 0x8000, 0x50000, 0x4FFFC, 0x12345678, 0x4FFFC, "R0")
	stackTrace.RecordPop(2, 0x8004, 0x4FFFC, 0x50000, 0x12345678, 0x4FFFC, "R0")

	err := stackTrace.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Stack Trace Report") {
		t.Error("Output should contain title")
	}
	if !strings.Contains(output, "PUSH") {
		t.Error("Output should contain PUSH operation")
	}
	if !strings.Contains(output, "POP") {
		t.Error("Output should contain POP operation")
	}
}

func TestStackTraceString(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)
	stackTrace.Start(0x50000)

	stackTrace.RecordPush(1, 0x8000, 0x50000, 0x4FFFC, 0x12345678, 0x4FFFC, "R0")

	str := stackTrace.String()
	if !strings.Contains(str, "Stack Trace Summary") {
		t.Error("String output should contain title")
	}
	if !strings.Contains(str, "Max Stack Usage") {
		t.Error("String output should contain max usage")
	}
	if !strings.Contains(str, "Total Pushes:       1") {
		t.Error("String output should show 1 push")
	}
}

func TestStackTraceDisabled(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)
	stackTrace.Enabled = false
	stackTrace.Start(0x50000)

	// Record when disabled
	stackTrace.RecordPush(1, 0x8000, 0x50000, 0x4FFFC, 0x12345678, 0x4FFFC, "R0")

	entries := stackTrace.GetEntries()
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries when disabled, got %d", len(entries))
	}
}

func TestStackTraceStatistics(t *testing.T) {
	var buf bytes.Buffer
	stackTrace := vm.NewStackTrace(&buf, 0x50000, 0x40000)
	stackTrace.Start(0x50000)

	// Perform various operations
	stackTrace.RecordPush(1, 0x8000, 0x50000, 0x4FFFC, 0x11111111, 0x4FFFC, "R1")
	stackTrace.RecordPush(2, 0x8004, 0x4FFFC, 0x4FFF8, 0x22222222, 0x4FFF8, "R2")
	stackTrace.RecordPush(3, 0x8008, 0x4FFF8, 0x4FFF4, 0x33333333, 0x4FFF4, "R3")
	stackTrace.RecordPop(4, 0x800C, 0x4FFF4, 0x4FFF8, 0x33333333, 0x4FFF4, "R3")
	stackTrace.RecordPop(5, 0x8010, 0x4FFF8, 0x4FFFC, 0x22222222, 0x4FFF8, "R2")
	stackTrace.RecordSPMove(6, 0x8014, 0x4FFFC, 0x4FFF0)

	// Check statistics via String()
	str := stackTrace.String()
	if !strings.Contains(str, "Total Pushes:       3") {
		t.Error("Should show 3 pushes")
	}
	if !strings.Contains(str, "Total Pops:         2") {
		t.Error("Should show 2 pops")
	}
}
