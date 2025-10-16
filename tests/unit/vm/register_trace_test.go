package vm_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestRegisterTrace_Basic(t *testing.T) {
	buf := &bytes.Buffer{}
	trace := vm.NewRegisterTrace(buf)

	if !trace.Enabled {
		t.Error("Register trace should be enabled by default")
	}

	trace.Start()

	// Record some register operations
	trace.RecordWrite(1, 0x8000, "R0", 0, 42)
	trace.RecordWrite(2, 0x8004, "R1", 0, 100)
	trace.RecordRead(3, 0x8008, "R0", 42)
	trace.RecordWrite(4, 0x800C, "R2", 0, 142)

	// Check statistics
	if trace.GetStats("R0").WriteCount != 1 {
		t.Errorf("R0 write count = %d, want 1", trace.GetStats("R0").WriteCount)
	}
	if trace.GetStats("R0").ReadCount != 1 {
		t.Errorf("R0 read count = %d, want 1", trace.GetStats("R0").ReadCount)
	}
	if trace.GetStats("R1").WriteCount != 1 {
		t.Errorf("R1 write count = %d, want 1", trace.GetStats("R1").WriteCount)
	}
	if trace.GetStats("R2").WriteCount != 1 {
		t.Errorf("R2 write count = %d, want 1", trace.GetStats("R2").WriteCount)
	}
}

func TestRegisterTrace_HotRegisters(t *testing.T) {
	buf := &bytes.Buffer{}
	trace := vm.NewRegisterTrace(buf)
	trace.Start()

	// Create access patterns
	// R0: 10 accesses (5 reads, 5 writes)
	for i := 0; i < 5; i++ {
		trace.RecordWrite(uint64(i), 0x8000, "R0", 0, uint32(i))
		trace.RecordRead(uint64(i+5), 0x8004, "R0", uint32(i))
	}

	// R1: 4 accesses (2 reads, 2 writes)
	for i := 0; i < 2; i++ {
		trace.RecordWrite(uint64(i+10), 0x8008, "R1", 0, uint32(i))
		trace.RecordRead(uint64(i+12), 0x800C, "R1", uint32(i))
	}

	// R2: 1 write
	trace.RecordWrite(14, 0x8010, "R2", 0, 99)

	hotRegs := trace.GetHotRegisters(3)
	if len(hotRegs) != 3 {
		t.Fatalf("Expected 3 hot registers, got %d", len(hotRegs))
	}

	// R0 should be first (most accesses)
	if hotRegs[0].RegisterName != "R0" {
		t.Errorf("First hot register = %s, want R0", hotRegs[0].RegisterName)
	}
	if hotRegs[0].ReadCount+hotRegs[0].WriteCount != 10 {
		t.Errorf("R0 total accesses = %d, want 10", hotRegs[0].ReadCount+hotRegs[0].WriteCount)
	}

	// R1 should be second
	if hotRegs[1].RegisterName != "R1" {
		t.Errorf("Second hot register = %s, want R1", hotRegs[1].RegisterName)
	}

	// R2 should be third
	if hotRegs[2].RegisterName != "R2" {
		t.Errorf("Third hot register = %s, want R2", hotRegs[2].RegisterName)
	}
}

func TestRegisterTrace_UnusedRegisters(t *testing.T) {
	buf := &bytes.Buffer{}
	trace := vm.NewRegisterTrace(buf)
	trace.Start()

	// Only use R0, R1, R2
	trace.RecordWrite(1, 0x8000, "R0", 0, 42)
	trace.RecordWrite(2, 0x8004, "R1", 0, 100)
	trace.RecordWrite(3, 0x8008, "R2", 0, 200)

	unused := trace.GetUnusedRegisters()

	// Should have unused registers (R3-R15)
	if len(unused) != 13 {
		t.Errorf("Unused register count = %d, want 13", len(unused))
	}

	// R0, R1, R2 should not be in unused list
	for _, reg := range unused {
		if reg == "R0" || reg == "R1" || reg == "R2" {
			t.Errorf("Register %s should not be in unused list", reg)
		}
	}
}

func TestRegisterTrace_ReadBeforeWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	trace := vm.NewRegisterTrace(buf)
	trace.Start()

	// R0: read before write
	trace.RecordRead(1, 0x8000, "R0", 0)
	trace.RecordWrite(2, 0x8004, "R0", 0, 42)

	// R1: write before read
	trace.RecordWrite(3, 0x8008, "R1", 0, 100)
	trace.RecordRead(4, 0x800C, "R1", 100)

	// R2: read but never written
	trace.RecordRead(5, 0x8010, "R2", 0)

	rbw := trace.DetectReadBeforeWrite()

	// R0 and R2 should be in the list
	if len(rbw) != 2 {
		t.Fatalf("Read-before-write count = %d, want 2", len(rbw))
	}

	found := make(map[string]bool)
	for _, reg := range rbw {
		found[reg] = true
	}

	if !found["R0"] {
		t.Error("R0 should be in read-before-write list")
	}
	if !found["R2"] {
		t.Error("R2 should be in read-before-write list")
	}
	if found["R1"] {
		t.Error("R1 should not be in read-before-write list")
	}
}

func TestRegisterTrace_UniqueValues(t *testing.T) {
	buf := &bytes.Buffer{}
	trace := vm.NewRegisterTrace(buf)
	trace.Start()

	// Write different values to R0
	trace.RecordWrite(1, 0x8000, "R0", 0, 10)
	trace.RecordWrite(2, 0x8004, "R0", 10, 20)
	trace.RecordWrite(3, 0x8008, "R0", 20, 30)
	trace.RecordWrite(4, 0x800C, "R0", 30, 10) // Repeat value

	stats := trace.GetStats("R0")
	if stats.UniqueValues != 3 {
		t.Errorf("R0 unique values = %d, want 3 (10, 20, 30)", stats.UniqueValues)
	}
	if stats.LastValue != 10 {
		t.Errorf("R0 last value = %d, want 10", stats.LastValue)
	}
}

func TestRegisterTrace_Flush(t *testing.T) {
	buf := &bytes.Buffer{}
	trace := vm.NewRegisterTrace(buf)
	trace.Start()

	// Record some operations
	trace.RecordWrite(1, 0x8000, "R0", 0, 42)
	trace.RecordWrite(2, 0x8004, "R1", 0, 100)
	trace.RecordRead(3, 0x8008, "R0", 42)
	trace.RecordWrite(4, 0x800C, "R2", 0, 200)

	// Flush to buffer
	err := trace.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	output := buf.String()

	// Check for expected content
	if !strings.Contains(output, "Register Access Pattern Analysis") {
		t.Error("Output missing header")
	}
	if !strings.Contains(output, "Total Reads:") {
		t.Error("Output missing total reads")
	}
	if !strings.Contains(output, "Total Writes:") {
		t.Error("Output missing total writes")
	}
	if !strings.Contains(output, "Hot Registers") {
		t.Error("Output missing hot registers section")
	}
	if !strings.Contains(output, "R0") {
		t.Error("Output missing R0 statistics")
	}
}

func TestRegisterTrace_JSON(t *testing.T) {
	buf := &bytes.Buffer{}
	trace := vm.NewRegisterTrace(buf)
	trace.Start()

	// Record some operations
	trace.RecordWrite(1, 0x8000, "R0", 0, 42)
	trace.RecordWrite(2, 0x8004, "R1", 0, 100)
	trace.RecordRead(3, 0x8008, "R0", 42)

	// Export as JSON
	jsonBuf := &bytes.Buffer{}
	err := trace.ExportJSON(jsonBuf)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Parse JSON to verify structure
	var data map[string]interface{}
	if err := json.Unmarshal(jsonBuf.Bytes(), &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check for expected fields
	if _, ok := data["total_reads"]; !ok {
		t.Error("JSON missing total_reads field")
	}
	if _, ok := data["total_writes"]; !ok {
		t.Error("JSON missing total_writes field")
	}
	if _, ok := data["register_stats"]; !ok {
		t.Error("JSON missing register_stats field")
	}
	if _, ok := data["hot_registers"]; !ok {
		t.Error("JSON missing hot_registers field")
	}
	if _, ok := data["unused_registers"]; !ok {
		t.Error("JSON missing unused_registers field")
	}
	if _, ok := data["read_before_write"]; !ok {
		t.Error("JSON missing read_before_write field")
	}

	// Verify totals
	if data["total_reads"].(float64) != 1 {
		t.Errorf("total_reads = %v, want 1", data["total_reads"])
	}
	if data["total_writes"].(float64) != 2 {
		t.Errorf("total_writes = %v, want 2", data["total_writes"])
	}
}

func TestRegisterTrace_Disabled(t *testing.T) {
	buf := &bytes.Buffer{}
	trace := vm.NewRegisterTrace(buf)
	trace.Enabled = false
	trace.Start()

	// Record operations (should be ignored)
	trace.RecordWrite(1, 0x8000, "R0", 0, 42)
	trace.RecordRead(2, 0x8004, "R0", 42)

	// Should have no statistics
	if trace.GetStats("R0") != nil {
		t.Error("Disabled trace should not record statistics")
	}
}

func TestRegisterTrace_String(t *testing.T) {
	buf := &bytes.Buffer{}
	trace := vm.NewRegisterTrace(buf)
	trace.Start()

	// Record some operations
	trace.RecordWrite(1, 0x8000, "R0", 0, 42)
	trace.RecordRead(2, 0x8004, "R0", 42)

	str := trace.String()

	if !strings.Contains(str, "Register Access Summary") {
		t.Error("String() missing header")
	}
	if !strings.Contains(str, "Total Reads:") {
		t.Error("String() missing total reads")
	}
	if !strings.Contains(str, "Total Writes:") {
		t.Error("String() missing total writes")
	}
}

func TestRegisterTrace_SequenceTracking(t *testing.T) {
	buf := &bytes.Buffer{}
	trace := vm.NewRegisterTrace(buf)
	trace.Start()

	// Record operations with sequence numbers
	trace.RecordWrite(10, 0x8000, "R0", 0, 42)
	trace.RecordRead(20, 0x8004, "R0", 42)
	trace.RecordWrite(30, 0x8008, "R0", 42, 100)
	trace.RecordRead(40, 0x800C, "R0", 100)

	stats := trace.GetStats("R0")
	if stats.FirstWrite != 10 {
		t.Errorf("FirstWrite = %d, want 10", stats.FirstWrite)
	}
	if stats.FirstRead != 20 {
		t.Errorf("FirstRead = %d, want 20", stats.FirstRead)
	}
	if stats.LastWrite != 30 {
		t.Errorf("LastWrite = %d, want 30", stats.LastWrite)
	}
	if stats.LastRead != 40 {
		t.Errorf("LastRead = %d, want 40", stats.LastRead)
	}
}
