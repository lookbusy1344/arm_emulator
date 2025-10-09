package vm_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestPerformanceStatistics_RecordInstruction(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record some instructions
	stats.RecordInstruction("MOV", 0x8000, 1)
	stats.RecordInstruction("ADD", 0x8004, 1)
	stats.RecordInstruction("MOV", 0x8008, 1)

	// Check totals
	if stats.TotalInstructions != 3 {
		t.Errorf("Expected 3 instructions, got %d", stats.TotalInstructions)
	}
	if stats.TotalCycles != 3 {
		t.Errorf("Expected 3 cycles, got %d", stats.TotalCycles)
	}

	// Check instruction counts
	if stats.InstructionCounts["MOV"] != 2 {
		t.Errorf("Expected 2 MOV instructions, got %d", stats.InstructionCounts["MOV"])
	}
	if stats.InstructionCounts["ADD"] != 1 {
		t.Errorf("Expected 1 ADD instruction, got %d", stats.InstructionCounts["ADD"])
	}
}

func TestPerformanceStatistics_RecordBranch(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record some branches
	stats.RecordBranch(true)  // taken
	stats.RecordBranch(true)  // taken
	stats.RecordBranch(false) // not taken

	if stats.BranchCount != 3 {
		t.Errorf("Expected 3 branches, got %d", stats.BranchCount)
	}
	if stats.BranchTakenCount != 2 {
		t.Errorf("Expected 2 taken branches, got %d", stats.BranchTakenCount)
	}
	if stats.BranchMissedCount != 1 {
		t.Errorf("Expected 1 not-taken branch, got %d", stats.BranchMissedCount)
	}
}

func TestPerformanceStatistics_RecordFunctionCall(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record function calls
	stats.RecordFunctionCall(0x8100, "factorial")
	stats.RecordFunctionCall(0x8200, "fibonacci")
	stats.RecordFunctionCall(0x8100, "factorial") // second call

	if len(stats.FunctionCalls) != 2 {
		t.Errorf("Expected 2 unique functions, got %d", len(stats.FunctionCalls))
	}

	// Check factorial was called twice
	if stats.FunctionCalls[0x8100].CallCount != 2 {
		t.Errorf("Expected factorial called 2 times, got %d", stats.FunctionCalls[0x8100].CallCount)
	}

	// Check fibonacci was called once
	if stats.FunctionCalls[0x8200].CallCount != 1 {
		t.Errorf("Expected fibonacci called 1 time, got %d", stats.FunctionCalls[0x8200].CallCount)
	}
}

func TestPerformanceStatistics_RecordMemoryAccess(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record memory accesses
	stats.RecordMemoryRead(4)  // word
	stats.RecordMemoryRead(4)  // word
	stats.RecordMemoryWrite(4) // word
	stats.RecordMemoryRead(1)  // byte

	if stats.MemoryReads != 3 {
		t.Errorf("Expected 3 reads, got %d", stats.MemoryReads)
	}
	if stats.MemoryWrites != 1 {
		t.Errorf("Expected 1 write, got %d", stats.MemoryWrites)
	}
	if stats.BytesRead != 9 {
		t.Errorf("Expected 9 bytes read (4+4+1), got %d", stats.BytesRead)
	}
	if stats.BytesWritten != 4 {
		t.Errorf("Expected 4 bytes written, got %d", stats.BytesWritten)
	}
}

func TestPerformanceStatistics_HotPath(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record instructions at various addresses
	for i := 0; i < 10; i++ {
		stats.RecordInstruction("NOP", 0x8000, 1) // executed 10 times
	}
	for i := 0; i < 5; i++ {
		stats.RecordInstruction("NOP", 0x8004, 1) // executed 5 times
	}
	stats.RecordInstruction("NOP", 0x8008, 1) // executed 1 time

	// Get hot path
	hotPath := stats.GetTopHotPath(3)

	if len(hotPath) != 3 {
		t.Fatalf("Expected 3 hot path entries, got %d", len(hotPath))
	}

	// Should be sorted by count descending
	if hotPath[0].Address != 0x8000 || hotPath[0].Count != 10 {
		t.Errorf("Expected first entry to be 0x8000 with count 10, got 0x%04X with count %d",
			hotPath[0].Address, hotPath[0].Count)
	}
	if hotPath[1].Address != 0x8004 || hotPath[1].Count != 5 {
		t.Errorf("Expected second entry to be 0x8004 with count 5, got 0x%04X with count %d",
			hotPath[1].Address, hotPath[1].Count)
	}
	if hotPath[2].Address != 0x8008 || hotPath[2].Count != 1 {
		t.Errorf("Expected third entry to be 0x8008 with count 1, got 0x%04X with count %d",
			hotPath[2].Address, hotPath[2].Count)
	}
}

func TestPerformanceStatistics_GetTopInstructions(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record various instructions
	for i := 0; i < 100; i++ {
		stats.RecordInstruction("MOV", 0x8000, 1)
	}
	for i := 0; i < 50; i++ {
		stats.RecordInstruction("ADD", 0x8004, 1)
	}
	for i := 0; i < 25; i++ {
		stats.RecordInstruction("SUB", 0x8008, 1)
	}

	// Get top 2 instructions
	topInsts := stats.GetTopInstructions(2)

	if len(topInsts) != 2 {
		t.Fatalf("Expected 2 instructions, got %d", len(topInsts))
	}

	// Should be sorted by count
	if topInsts[0].Mnemonic != "MOV" || topInsts[0].Count != 100 {
		t.Errorf("Expected MOV with 100, got %s with %d", topInsts[0].Mnemonic, topInsts[0].Count)
	}
	if topInsts[1].Mnemonic != "ADD" || topInsts[1].Count != 50 {
		t.Errorf("Expected ADD with 50, got %s with %d", topInsts[1].Mnemonic, topInsts[1].Count)
	}
}

func TestPerformanceStatistics_ExportJSON(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record some data
	stats.RecordInstruction("MOV", 0x8000, 1)
	stats.RecordInstruction("ADD", 0x8004, 1)
	stats.RecordBranch(true)

	// Export to JSON
	var buf bytes.Buffer
	if err := stats.ExportJSON(&buf); err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var data map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Check some fields exist
	if _, ok := data["total_instructions"]; !ok {
		t.Error("JSON missing total_instructions field")
	}
	if _, ok := data["total_cycles"]; !ok {
		t.Error("JSON missing total_cycles field")
	}
	if _, ok := data["branch_count"]; !ok {
		t.Error("JSON missing branch_count field")
	}
}

func TestPerformanceStatistics_ExportCSV(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record some data
	stats.RecordInstruction("MOV", 0x8000, 1)
	stats.RecordInstruction("ADD", 0x8004, 1)

	// Export to CSV
	var buf bytes.Buffer
	if err := stats.ExportCSV(&buf); err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("CSV export produced no output")
	}

	// Should contain headers
	if !strings.Contains(output, "Metric") {
		t.Error("CSV missing Metric header")
	}
	if !strings.Contains(output, "Value") {
		t.Error("CSV missing Value header")
	}

	// Should contain data
	if !strings.Contains(output, "Total Instructions") {
		t.Error("CSV missing Total Instructions")
	}
}

func TestPerformanceStatistics_ExportHTML(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record some data
	stats.RecordInstruction("MOV", 0x8000, 1)
	stats.RecordBranch(true)

	// Export to HTML
	var buf bytes.Buffer
	if err := stats.ExportHTML(&buf); err != nil {
		t.Fatalf("ExportHTML failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("HTML export produced no output")
	}

	// Should be HTML
	if !strings.Contains(output, "<html>") {
		t.Error("Output doesn't look like HTML")
	}
	if !strings.Contains(output, "Performance Statistics") {
		t.Error("HTML missing title")
	}
	if !strings.Contains(output, "Total Instructions") {
		t.Error("HTML missing instruction count")
	}
}

func TestPerformanceStatistics_String(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record some data
	stats.RecordInstruction("MOV", 0x8000, 1)
	stats.RecordInstruction("ADD", 0x8004, 1)
	stats.RecordBranch(true)
	stats.RecordMemoryRead(4)

	// Get string representation
	output := stats.String()

	if output == "" {
		t.Error("String() produced no output")
	}

	// Should contain key metrics
	if !strings.Contains(output, "Total Instructions") {
		t.Error("String output missing Total Instructions")
	}
	if !strings.Contains(output, "Branch Count") {
		t.Error("String output missing Branch Count")
	}
	if !strings.Contains(output, "Memory Reads") {
		t.Error("String output missing Memory Reads")
	}
}

func TestPerformanceStatistics_Finalize(t *testing.T) {
	stats := vm.NewPerformanceStatistics()
	stats.Start()

	// Record some instructions
	stats.RecordInstruction("MOV", 0x8000, 1)
	stats.RecordInstruction("ADD", 0x8004, 1)

	// Finalize calculates instructions per second
	stats.Finalize()

	// Should have non-zero execution time
	if stats.ExecutionTime == 0 {
		t.Error("ExecutionTime is zero after Finalize")
	}

	// InstructionsPerSec should be calculated
	// (might be 0 if execution was too fast, but should be set)
	if stats.InstructionsPerSec < 0 {
		t.Error("InstructionsPerSec is negative")
	}
}
