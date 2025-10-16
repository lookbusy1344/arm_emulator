package vm_test

import (
	"bytes"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestSetSPWithTrace verifies stack pointer tracing
func TestSetSPWithTrace(t *testing.T) {
	v := vm.NewVM()

	// Enable stack tracing
	var buf bytes.Buffer
	trace := vm.NewStackTrace(&buf, vm.StackSegmentStart, vm.StackSegmentSize)
	trace.Start(uint32(vm.StackSegmentStart + vm.StackSegmentSize))
	v.StackTrace = trace

	// Set SP with trace
	newSP := uint32(0x00020000)
	pc := uint32(0x00008000)
	v.CPU.SetSPWithTrace(v, newSP, pc)

	// Verify SP was set
	if v.CPU.GetSP() != newSP {
		t.Errorf("Expected SP=0x%08X, got 0x%08X", newSP, v.CPU.GetSP())
	}

	// Verify trace recorded the operation
	entries := trace.GetEntries()
	if len(entries) == 0 {
		t.Error("Expected stack trace to record SP modification")
	}
}

// TestSetSPWithTraceDisabled verifies SetSPWithTrace when tracing is disabled
func TestSetSPWithTraceDisabled(t *testing.T) {
	v := vm.NewVM()

	// No stack trace enabled
	newSP := uint32(0x00020000)
	pc := uint32(0x00008000)
	v.CPU.SetSPWithTrace(v, newSP, pc)

	// Verify SP was set
	if v.CPU.GetSP() != newSP {
		t.Errorf("Expected SP=0x%08X, got 0x%08X", newSP, v.CPU.GetSP())
	}
}

// TestGetRegisterWithTrace verifies register read tracing
func TestGetRegisterWithTrace(t *testing.T) {
	v := vm.NewVM()

	// Enable register tracing
	var buf bytes.Buffer
	trace := vm.NewRegisterTrace(&buf)
	trace.Start()
	v.RegisterTrace = trace

	// Set a known value
	v.CPU.R[5] = 0x12345678

	// Read with trace
	pc := uint32(0x00008000)
	value := v.CPU.GetRegisterWithTrace(v, 5, pc)

	// Verify value
	if value != 0x12345678 {
		t.Errorf("Expected R5=0x12345678, got 0x%08X", value)
	}

	// Verify trace recorded the read
	stats := trace.GetStats("R5")
	if stats.ReadCount != 1 {
		t.Errorf("Expected 1 read for R5, got %d", stats.ReadCount)
	}
}

// TestGetRegisterWithTraceDisabled verifies GetRegisterWithTrace when tracing is disabled
func TestGetRegisterWithTraceDisabled(t *testing.T) {
	v := vm.NewVM()

	// Set a known value
	v.CPU.R[5] = 0x12345678

	// Read without trace
	pc := uint32(0x00008000)
	value := v.CPU.GetRegisterWithTrace(v, 5, pc)

	// Verify value
	if value != 0x12345678 {
		t.Errorf("Expected R5=0x12345678, got 0x%08X", value)
	}
}

// TestSetRegisterWithTrace verifies register write tracing
func TestSetRegisterWithTrace(t *testing.T) {
	v := vm.NewVM()

	// Enable register tracing
	var buf bytes.Buffer
	trace := vm.NewRegisterTrace(&buf)
	trace.Start()
	v.RegisterTrace = trace

	// Write with trace
	pc := uint32(0x00008000)
	v.CPU.SetRegisterWithTrace(v, 7, 0xABCDEF00, pc)

	// Verify value was set
	if v.CPU.R[7] != 0xABCDEF00 {
		t.Errorf("Expected R7=0xABCDEF00, got 0x%08X", v.CPU.R[7])
	}

	// Verify trace recorded the write
	stats := trace.GetStats("R7")
	if stats.WriteCount != 1 {
		t.Errorf("Expected 1 write for R7, got %d", stats.WriteCount)
	}
}

// TestSetRegisterWithTraceDisabled verifies SetRegisterWithTrace when tracing is disabled
func TestSetRegisterWithTraceDisabled(t *testing.T) {
	v := vm.NewVM()

	// Write without trace
	pc := uint32(0x00008000)
	v.CPU.SetRegisterWithTrace(v, 7, 0xABCDEF00, pc)

	// Verify value was set
	if v.CPU.R[7] != 0xABCDEF00 {
		t.Errorf("Expected R7=0xABCDEF00, got 0x%08X", v.CPU.R[7])
	}
}

// TestGetRegisterWithTracePC verifies tracing PC reads
func TestGetRegisterWithTracePC(t *testing.T) {
	v := vm.NewVM()

	// Enable register tracing
	var buf bytes.Buffer
	trace := vm.NewRegisterTrace(&buf)
	trace.Start()
	v.RegisterTrace = trace

	// Set PC
	v.CPU.PC = 0x00008000

	// Read PC (R15) with trace
	pc := uint32(0x00008000)
	value := v.CPU.GetRegisterWithTrace(v, 15, pc)

	// ARM convention: reading PC returns PC+8
	expected := uint32(0x00008008)
	if value != expected {
		t.Errorf("Expected PC+8=0x%08X, got 0x%08X", expected, value)
	}

	// Verify trace recorded the read
	stats := trace.GetStats("PC")
	if stats.ReadCount != 1 {
		t.Errorf("Expected 1 read for PC, got %d", stats.ReadCount)
	}
}

// TestSetRegisterWithTracePC verifies tracing PC writes
func TestSetRegisterWithTracePC(t *testing.T) {
	v := vm.NewVM()

	// Enable register tracing
	var buf bytes.Buffer
	trace := vm.NewRegisterTrace(&buf)
	trace.Start()
	v.RegisterTrace = trace

	// Write PC with trace
	pc := uint32(0x00008000)
	v.CPU.SetRegisterWithTrace(v, 15, 0x00009000, pc)

	// Verify PC was set
	if v.CPU.PC != 0x00009000 {
		t.Errorf("Expected PC=0x00009000, got 0x%08X", v.CPU.PC)
	}

	// Verify trace recorded the write
	stats := trace.GetStats("PC")
	if stats.WriteCount != 1 {
		t.Errorf("Expected 1 write for PC, got %d", stats.WriteCount)
	}
}

// TestGetRegisterWithTraceSP verifies tracing SP reads
func TestGetRegisterWithTraceSP(t *testing.T) {
	v := vm.NewVM()

	// Enable register tracing
	var buf bytes.Buffer
	trace := vm.NewRegisterTrace(&buf)
	trace.Start()
	v.RegisterTrace = trace

	// Set SP
	v.CPU.R[13] = 0x00020000

	// Read SP (R13) with trace
	pc := uint32(0x00008000)
	value := v.CPU.GetRegisterWithTrace(v, 13, pc)

	if value != 0x00020000 {
		t.Errorf("Expected SP=0x00020000, got 0x%08X", value)
	}

	// Verify trace recorded the read
	stats := trace.GetStats("SP")
	if stats.ReadCount != 1 {
		t.Errorf("Expected 1 read for SP, got %d", stats.ReadCount)
	}
}

// TestSetRegisterWithTraceSP verifies tracing SP writes
func TestSetRegisterWithTraceSP(t *testing.T) {
	v := vm.NewVM()

	// Enable register tracing
	var buf bytes.Buffer
	trace := vm.NewRegisterTrace(&buf)
	trace.Start()
	v.RegisterTrace = trace

	// Write SP with trace
	pc := uint32(0x00008000)
	v.CPU.SetRegisterWithTrace(v, 13, 0x00021000, pc)

	if v.CPU.R[13] != 0x00021000 {
		t.Errorf("Expected SP=0x00021000, got 0x%08X", v.CPU.R[13])
	}

	// Verify trace recorded the write
	stats := trace.GetStats("SP")
	if stats.WriteCount != 1 {
		t.Errorf("Expected 1 write for SP, got %d", stats.WriteCount)
	}
}

// TestMultipleRegisterTraces verifies tracking multiple register operations
func TestMultipleRegisterTraces(t *testing.T) {
	v := vm.NewVM()

	// Enable register tracing
	var buf bytes.Buffer
	trace := vm.NewRegisterTrace(&buf)
	trace.Start()
	v.RegisterTrace = trace

	pc := uint32(0x00008000)

	// Perform multiple operations
	v.CPU.SetRegisterWithTrace(v, 0, 100, pc)
	v.CPU.SetRegisterWithTrace(v, 1, 200, pc)
	_ = v.CPU.GetRegisterWithTrace(v, 0, pc)
	_ = v.CPU.GetRegisterWithTrace(v, 0, pc)
	v.CPU.SetRegisterWithTrace(v, 0, 300, pc)

	// Verify stats
	statsR0 := trace.GetStats("R0")
	statsR1 := trace.GetStats("R1")

	// R0: 2 writes, 2 reads
	if statsR0.WriteCount != 2 {
		t.Errorf("Expected 2 writes for R0, got %d", statsR0.WriteCount)
	}
	if statsR0.ReadCount != 2 {
		t.Errorf("Expected 2 reads for R0, got %d", statsR0.ReadCount)
	}

	// R1: 1 write, 0 reads
	if statsR1.WriteCount != 1 {
		t.Errorf("Expected 1 write for R1, got %d", statsR1.WriteCount)
	}
	if statsR1.ReadCount != 0 {
		t.Errorf("Expected 0 reads for R1, got %d", statsR1.ReadCount)
	}
}

// TestConditionCodeString verifies ConditionCode.String() method
func TestConditionCodeString(t *testing.T) {
	tests := []struct {
		cond     vm.ConditionCode
		expected string
	}{
		{vm.CondEQ, "EQ"},
		{vm.CondNE, "NE"},
		{vm.CondCS, "CS"},
		{vm.CondCC, "CC"},
		{vm.CondMI, "MI"},
		{vm.CondPL, "PL"},
		{vm.CondVS, "VS"},
		{vm.CondVC, "VC"},
		{vm.CondHI, "HI"},
		{vm.CondLS, "LS"},
		{vm.CondGE, "GE"},
		{vm.CondLT, "LT"},
		{vm.CondGT, "GT"},
		{vm.CondLE, "LE"},
		{vm.CondAL, "AL"},
	}

	for _, tt := range tests {
		result := tt.cond.String()
		if result != tt.expected {
			t.Errorf("Expected %s.String()='%s', got '%s'", tt.expected, tt.expected, result)
		}
	}
}

// TestParseConditionCode verifies ParseConditionCode function
func TestParseConditionCode(t *testing.T) {
	tests := []struct {
		input    string
		expected vm.ConditionCode
		wantErr  bool
	}{
		{"EQ", vm.CondEQ, false},
		{"NE", vm.CondNE, false},
		{"CS", vm.CondCS, false},
		{"HS", vm.CondCS, false}, // HS is alias for CS
		{"CC", vm.CondCC, false},
		{"LO", vm.CondCC, false}, // LO is alias for CC
		{"MI", vm.CondMI, false},
		{"PL", vm.CondPL, false},
		{"VS", vm.CondVS, false},
		{"VC", vm.CondVC, false},
		{"HI", vm.CondHI, false},
		{"LS", vm.CondLS, false},
		{"GE", vm.CondGE, false},
		{"LT", vm.CondLT, false},
		{"GT", vm.CondGT, false},
		{"LE", vm.CondLE, false},
		{"AL", vm.CondAL, false},
		{"", vm.CondAL, false},       // Empty defaults to AL
		{"INVALID", vm.CondAL, true}, // Invalid code
		{"eq", vm.CondEQ, true},      // Case sensitive - lowercase fails
		{"nE", vm.CondNE, true},      // Case sensitive - mixed case fails
	}

	for _, tt := range tests {
		result, ok := vm.ParseConditionCode(tt.input)

		if ok == tt.wantErr {
			// If we expected an error (!ok), but got ok=true, or vice versa
			t.Errorf("ParseConditionCode(%q) ok = %v, wantErr %v", tt.input, ok, tt.wantErr)
			continue
		}

		if ok && result != tt.expected {
			t.Errorf("ParseConditionCode(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
