package vm

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestFlagTraceBasic(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)

	if !flagTrace.Enabled {
		t.Error("Flag trace should be enabled by default")
	}

	initialFlags := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.Start(initialFlags)

	// Record flag change
	newFlags := vm.CPSR{N: true, Z: false, C: false, V: false}
	flagTrace.RecordFlags(1, 0x8000, "SUBS R0, R0, #1", newFlags)

	entries := flagTrace.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", entry.Sequence)
	}
	if entry.PC != 0x8000 {
		t.Errorf("Expected PC 0x8000, got 0x%X", entry.PC)
	}
	if entry.Changed != "N" {
		t.Errorf("Expected changed flags 'N', got '%s'", entry.Changed)
	}
}

func TestFlagTraceNoChange(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)

	initialFlags := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.Start(initialFlags)

	// Record same flags (no change)
	flagTrace.RecordFlags(1, 0x8000, "MOV R0, R1", initialFlags)

	entries := flagTrace.GetEntries()
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries when no flags change, got %d", len(entries))
	}
}

func TestFlagTraceMultipleChanges(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)

	initialFlags := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.Start(initialFlags)

	// Change N and Z
	flags1 := vm.CPSR{N: true, Z: true, C: false, V: false}
	flagTrace.RecordFlags(1, 0x8000, "SUBS R0, R0, R0", flags1)

	// Change C
	flags2 := vm.CPSR{N: true, Z: true, C: true, V: false}
	flagTrace.RecordFlags(2, 0x8004, "ADDS R1, R1, R1", flags2)

	// Change V
	flags3 := vm.CPSR{N: true, Z: true, C: true, V: true}
	flagTrace.RecordFlags(3, 0x8008, "ADDS R2, R2, R2", flags3)

	entries := flagTrace.GetEntries()
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Check first entry
	if entries[0].Changed != "NZ" {
		t.Errorf("Expected 'NZ' changed, got '%s'", entries[0].Changed)
	}

	// Check second entry
	if entries[1].Changed != "C" {
		t.Errorf("Expected 'C' changed, got '%s'", entries[1].Changed)
	}

	// Check third entry
	if entries[2].Changed != "V" {
		t.Errorf("Expected 'V' changed, got '%s'", entries[2].Changed)
	}
}

func TestFlagTraceAllFlagsChange(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)

	initialFlags := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.Start(initialFlags)

	// Change all flags at once
	newFlags := vm.CPSR{N: true, Z: true, C: true, V: true}
	flagTrace.RecordFlags(1, 0x8000, "ADDS R0, R1, R2", newFlags)

	entries := flagTrace.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Changed != "NZCV" {
		t.Errorf("Expected 'NZCV' changed, got '%s'", entry.Changed)
	}
}

func TestFlagTraceFlagToggle(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)

	flags1 := vm.CPSR{N: false, Z: true, C: false, V: false}
	flagTrace.Start(flags1)

	// Toggle Z flag off
	flags2 := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.RecordFlags(1, 0x8000, "SUBS R0, R0, #1", flags2)

	// Toggle Z flag on
	flags3 := vm.CPSR{N: false, Z: true, C: false, V: false}
	flagTrace.RecordFlags(2, 0x8004, "SUBS R0, R0, #0", flags3)

	entries := flagTrace.GetEntries()
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// Both should show Z changed
	if entries[0].Changed != "Z" {
		t.Errorf("First entry should show Z changed, got '%s'", entries[0].Changed)
	}
	if entries[1].Changed != "Z" {
		t.Errorf("Second entry should show Z changed, got '%s'", entries[1].Changed)
	}
}

func TestFlagTraceJSON(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)

	initialFlags := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.Start(initialFlags)

	newFlags := vm.CPSR{N: true, Z: false, C: false, V: false}
	flagTrace.RecordFlags(1, 0x8000, "SUBS R0, R0, #1", newFlags)

	err := flagTrace.ExportJSON(&buf)
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
	if data["total_changes"].(float64) != 1 {
		t.Error("JSON should show 1 total change")
	}
	if data["n_changes"].(float64) != 1 {
		t.Error("JSON should show 1 N flag change")
	}
	if data["z_changes"].(float64) != 0 {
		t.Error("JSON should show 0 Z flag changes")
	}
}

func TestFlagTraceFlush(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)

	initialFlags := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.Start(initialFlags)

	newFlags := vm.CPSR{N: true, Z: true, C: false, V: false}
	flagTrace.RecordFlags(1, 0x8000, "SUBS R0, R0, #1", newFlags)

	err := flagTrace.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Flag Change Trace Report") {
		t.Error("Output should contain title")
	}
	if !strings.Contains(output, "Total Changes:") {
		t.Error("Output should contain statistics")
	}
	if !strings.Contains(output, "changed: NZ") {
		t.Error("Output should show NZ changed")
	}
}

func TestFlagTraceString(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)

	initialFlags := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.Start(initialFlags)

	// Change N flag
	flags1 := vm.CPSR{N: true, Z: false, C: false, V: false}
	flagTrace.RecordFlags(1, 0x8000, "SUBS R0, R0, #1", flags1)

	// Change Z flag (N stays true)
	flags2 := vm.CPSR{N: true, Z: true, C: false, V: false}
	flagTrace.RecordFlags(2, 0x8004, "SUBS R0, R0, R0", flags2)

	str := flagTrace.String()
	if !strings.Contains(str, "Flag Change Summary") {
		t.Error("String output should contain title")
	}
	if !strings.Contains(str, "Total Changes:      2") {
		t.Error("String output should show 2 total changes")
	}
	if !strings.Contains(str, "N flag changes:     1") {
		t.Error("String output should show 1 N flag change")
	}
	if !strings.Contains(str, "Z flag changes:     1") {
		t.Error("String output should show 1 Z flag change")
	}
}

func TestFlagTraceDisabled(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)
	flagTrace.Enabled = false

	initialFlags := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.Start(initialFlags)

	// Record when disabled
	newFlags := vm.CPSR{N: true, Z: false, C: false, V: false}
	flagTrace.RecordFlags(1, 0x8000, "SUBS R0, R0, #1", newFlags)

	entries := flagTrace.GetEntries()
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries when disabled, got %d", len(entries))
	}
}

func TestFlagTraceStatistics(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)

	initialFlags := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.Start(initialFlags)

	// Change N flag 3 times
	flags1 := vm.CPSR{N: true, Z: false, C: false, V: false}
	flagTrace.RecordFlags(1, 0x8000, "SUBS R0, R0, #1", flags1)

	flags2 := vm.CPSR{N: false, Z: false, C: false, V: false}
	flagTrace.RecordFlags(2, 0x8004, "ADDS R0, R0, #1", flags2)

	flags3 := vm.CPSR{N: true, Z: false, C: false, V: false}
	flagTrace.RecordFlags(3, 0x8008, "SUBS R0, R0, #1", flags3)

	// Change Z flag once
	flags4 := vm.CPSR{N: true, Z: true, C: false, V: false}
	flagTrace.RecordFlags(4, 0x800C, "SUBS R0, R0, R0", flags4)

	str := flagTrace.String()
	if !strings.Contains(str, "Total Changes:      4") {
		t.Error("Should show 4 total changes")
	}
	if !strings.Contains(str, "N flag changes:     3") {
		t.Error("Should show 3 N flag changes")
	}
	if !strings.Contains(str, "Z flag changes:     1") {
		t.Error("Should show 1 Z flag change")
	}
}

func TestFlagTracePartialChanges(t *testing.T) {
	var buf bytes.Buffer
	flagTrace := vm.NewFlagTrace(&buf)

	// Start with some flags set
	initialFlags := vm.CPSR{N: true, Z: false, C: true, V: false}
	flagTrace.Start(initialFlags)

	// Change only N and V
	newFlags := vm.CPSR{N: false, Z: false, C: true, V: true}
	flagTrace.RecordFlags(1, 0x8000, "ADDS R0, R1, R2", newFlags)

	entries := flagTrace.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Changed != "NV" {
		t.Errorf("Expected 'NV' changed, got '%s'", entry.Changed)
	}

	// Verify old flags
	if !entry.OldFlags.N || entry.OldFlags.Z || !entry.OldFlags.C || entry.OldFlags.V {
		t.Error("Old flags don't match initial state")
	}

	// Verify new flags
	if entry.NewFlags.N || entry.NewFlags.Z || !entry.NewFlags.C || !entry.NewFlags.V {
		t.Error("New flags don't match expected state")
	}
}
