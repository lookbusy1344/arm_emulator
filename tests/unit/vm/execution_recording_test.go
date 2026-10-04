package vm_test

import (
	"bytes"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestMnemonic(t *testing.T) {
	tests := []struct {
		opcode uint32
		want   string
	}{
		{0xE3A0002A, "MOV"},   // MOV R0, #42
		{0xE0810002, "ADD"},   // ADD R0, R1, R2
		{0xE2511001, "SUB"},   // SUBS R1, R1, #1
		{0xE3500000, "CMP"},   // CMP R0, #0
		{0xE1C00001, "BIC"},   // BIC R0, R0, R1
		{0xE1E00001, "MVN"},   // MVN R0, R1
		{0xE0000291, "MUL"},   // MUL R0, R1, R2
		{0xE0203291, "MLA"},   // MLA R0, R1, R2, R3
		{0xE0810392, "UMULL"}, // UMULL R0, R1, R2, R3
		{0xE0C10392, "SMULL"}, // SMULL R0, R1, R2, R3
		{0xE5910000, "LDR"},   // LDR R0, [R1]
		{0xE5810000, "STR"},   // STR R0, [R1]
		{0xE5D10000, "LDRB"},  // LDRB R0, [R1]
		{0xE5C10000, "STRB"},  // STRB R0, [R1]
		{0xE1D100B0, "LDRH"},  // LDRH R0, [R1]
		{0xE1C100B0, "STRH"},  // STRH R0, [R1]
		{0xE1D100D0, "LDRSB"}, // LDRSB R0, [R1]
		{0xE1D100F0, "LDRSH"}, // LDRSH R0, [R1]
		{0xE8BD8000, "LDM"},   // LDMFD SP!, {PC}
		{0xE92D4000, "STM"},   // STMFD SP!, {LR}
		{0xEA000000, "B"},     // B .+8
		{0x1AFFFFFE, "B"},     // BNE .
		{0xEB000000, "BL"},    // BL .+8
		{0xE12FFF1E, "BX"},    // BX LR
		{0xE12FFF33, "BLX"},   // BLX R3
		{0xEF000000, "SWI"},   // SWI #0
		{0xE10F0000, "MRS"},   // MRS R0, CPSR
		{0xE128F000, "MSR"},   // MSR CPSR_f, R0
		{0xEE000000, "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := vm.Mnemonic(tt.opcode); got != tt.want {
			t.Errorf("Mnemonic(0x%08X) = %q, want %q", tt.opcode, got, tt.want)
		}
	}
}

// loadWords writes instructions at the code segment start and points PC there.
func loadWords(t *testing.T, v *vm.VM, words ...uint32) {
	t.Helper()
	setupCodeWrite(v)
	for i, w := range words {
		if err := v.Memory.WriteWord(vm.CodeSegmentStart+uint32(i)*4, w); err != nil {
			t.Fatal(err)
		}
	}
	v.CPU.PC = vm.CodeSegmentStart
}

func stepN(t *testing.T, v *vm.VM, n int) {
	t.Helper()
	for range n {
		if err := v.Step(); err != nil {
			t.Fatalf("Step failed: %v", err)
		}
	}
}

func TestStepRecordsStatistics(t *testing.T) {
	v := vm.NewVM()
	setupDataWrite(v)
	loadWords(t, v,
		0xE3A01802, // MOV R1, #0x20000
		0xE3A0002A, // MOV R0, #42
		0xE5810000, // STR R0, [R1]
		0xE5D12000, // LDRB R2, [R1]
		0xE3500000, // CMP R0, #0
		0x0A000000, // BEQ +8 (not taken)
		0x1A000000, // BNE +8 (taken)
	)
	v.Statistics = vm.NewPerformanceStatistics()
	v.Statistics.Start()

	stepN(t, v, 7)

	s := v.Statistics
	if s.TotalInstructions != 7 {
		t.Errorf("TotalInstructions = %d, want 7", s.TotalInstructions)
	}
	if s.InstructionCounts["MOV"] != 2 || s.InstructionCounts["B"] != 2 || s.InstructionCounts["LDRB"] != 1 {
		t.Errorf("unexpected instruction counts %v", s.InstructionCounts)
	}
	if s.BranchCount != 2 || s.BranchTakenCount != 1 || s.BranchMissedCount != 1 {
		t.Errorf("branches: count=%d taken=%d missed=%d, want 2/1/1",
			s.BranchCount, s.BranchTakenCount, s.BranchMissedCount)
	}
	if s.MemoryWrites != 1 || s.BytesWritten != 4 {
		t.Errorf("writes: %d ops %d bytes, want 1/4", s.MemoryWrites, s.BytesWritten)
	}
	if s.MemoryReads != 1 || s.BytesRead != 1 {
		t.Errorf("reads: %d ops %d bytes, want 1/1", s.MemoryReads, s.BytesRead)
	}
	if s.HotPath[vm.CodeSegmentStart] != 1 {
		t.Errorf("hot path missing entry address: %v", s.HotPath)
	}
}

func TestStepRecordsFunctionCalls(t *testing.T) {
	v := vm.NewVM()
	loadWords(t, v,
		0xEB000000, // BL +8 -> 0x8008
		0xE1A00000, // NOP
		0xE1A00000, // NOP (callee)
	)
	v.Statistics = vm.NewPerformanceStatistics()
	v.Statistics.Start()

	stepN(t, v, 1)

	const callee = vm.CodeSegmentStart + 8
	call, ok := v.Statistics.FunctionCalls[callee]
	if !ok || call.CallCount != 1 {
		t.Fatalf("expected one call to 0x%X, got %v", callee, v.Statistics.FunctionCalls)
	}
}

func TestStepRecordsExecutionTrace(t *testing.T) {
	v := vm.NewVM()
	loadWords(t, v,
		0xE3A0002A, // MOV R0, #42
		0xEA000000, // B +8 -> 0x800C
		0xE1A00000, // skipped
		0xE3A01001, // MOV R1, #1
	)
	var out bytes.Buffer
	v.ExecutionTrace = vm.NewExecutionTrace(&out)
	v.ExecutionTrace.LoadSourceMap(map[uint32]string{vm.CodeSegmentStart: "    MOV R0, #42   ; answer"})
	v.ExecutionTrace.Start()

	stepN(t, v, 3)

	entries := v.ExecutionTrace.GetEntries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 trace entries, got %d", len(entries))
	}
	wantAddr := []uint32{vm.CodeSegmentStart, vm.CodeSegmentStart + 4, vm.CodeSegmentStart + 12}
	for i, e := range entries {
		if e.Address != wantAddr[i] {
			t.Errorf("entry %d: address 0x%08X, want 0x%08X", i, e.Address, wantAddr[i])
		}
	}
	if entries[0].Opcode != 0xE3A0002A {
		t.Errorf("entry 0 opcode 0x%08X, want 0xE3A0002A", entries[0].Opcode)
	}
	if entries[0].Disassembly != "MOV R0, #42" {
		t.Errorf("entry 0 text %q, want source line without comment", entries[0].Disassembly)
	}
	if entries[1].Disassembly != "B" {
		t.Errorf("entry 1 text %q, want mnemonic fallback", entries[1].Disassembly)
	}
	if entries[0].RegisterChanges["R0"] != 42 {
		t.Errorf("entry 0 should record R0=42, got %v", entries[0].RegisterChanges)
	}
}
