package service_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// The first program dirties registers, flags, a data buffer, the heap, the exit code
// and the cycle count.
const dirtyProgram = `.org 0x8000
_start:
	MOV R5, #99
	LDR R1, =buffer
	MOV R2, #77
	STR R2, [R1]
	MOV R0, #16
	SWI #0x20
	MOV R6, R0
	CMP R0, #0
	MOV R0, #3
	SWI #0x00
buffer:
	.space 4
`

// The second program has the same layout and only exits.
const readerProgram = `.org 0x8000
_start:
	NOP
	NOP
	NOP
	NOP
	NOP
	NOP
	NOP
	NOP
	NOP
	MOV R0, #0
	SWI #0x00
buffer:
	.space 4
`

func runToHalt(t *testing.T, svc *service.DebuggerService) {
	t.Helper()
	if err := <-startRun(svc); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestReloadStartsFromCleanState(t *testing.T) {
	svc := newLoadedService(t, dirtyProgram)
	runToHalt(t, svc)
	if svc.GetExitCode() != 3 {
		t.Fatalf("setup: exit code %d, want 3", svc.GetExitCode())
	}

	loadSource(t, svc, readerProgram)

	regs := svc.GetRegisterState()
	for i, r := range regs.Registers[:vm.SP] {
		if r != 0 {
			t.Errorf("R%d = 0x%08X after reload, want 0", i, r)
		}
	}
	if regs.CPSR != (service.CPSRState{}) {
		t.Errorf("CPSR = %+v after reload, want clear", regs.CPSR)
	}
	if regs.Cycles != 0 {
		t.Errorf("cycles = %d after reload, want 0", regs.Cycles)
	}
	if regs.PC != 0x8000 || regs.Registers[vm.SP] != vm.StackSegmentStart+vm.StackSegmentSize {
		t.Errorf("PC=0x%08X SP=0x%08X, want entry and stack top", regs.PC, regs.Registers[vm.SP])
	}
	if svc.GetExitCode() != 0 {
		t.Errorf("exit code = %d after reload, want 0", svc.GetExitCode())
	}
	symbols := svc.GetSymbols()
	buf, err := svc.GetMemory(symbols["buffer"], 4)
	if err != nil {
		t.Fatal(err)
	}
	if buf[0] != 0 {
		t.Errorf("buffer = %v after reload, want zeroed", buf)
	}
	if heap := svc.GetVM().Memory.NextHeapAddress; heap != vm.HeapSegmentStart {
		t.Errorf("next heap address = 0x%08X after reload, want 0x%08X", heap, uint32(vm.HeapSegmentStart))
	}
}

func TestReloadKeepsDiagnosticsEnabledAndEmpty(t *testing.T) {
	svc := newLoadedService(t, dirtyProgram)
	if err := svc.EnableExecutionTrace(); err != nil {
		t.Fatal(err)
	}
	if err := svc.EnableStatistics(); err != nil {
		t.Fatal(err)
	}
	runToHalt(t, svc)

	loadSource(t, svc, readerProgram)

	entries, err := svc.GetExecutionTraceData()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("trace has %d entries after reload, want 0", len(entries))
	}
	stats, err := svc.GetStatistics()
	if err != nil || stats == nil || !stats.Enabled || stats.TotalInstructions != 0 {
		t.Fatalf("statistics after reload = %+v, %v; want enabled and empty", stats, err)
	}

	runToHalt(t, svc)
	if entries, _ = svc.GetExecutionTraceData(); len(entries) == 0 {
		t.Error("trace recorded nothing for the reloaded program")
	}
}

func TestReloadLowMemoryProgramAddsNoSegments(t *testing.T) {
	const lowProgram = ".org 0x0000\n_start:\n\tMOV R0, #0\n\tSWI #0x00\n"
	svc := service.NewDebuggerService(vm.NewVM())
	load := func() {
		program := parse(t, lowProgram)
		if err := svc.LoadProgram(program, 0); err != nil {
			t.Fatal(err)
		}
	}
	load()
	segments := len(svc.GetVM().Memory.Segments)
	load()
	load()
	if got := len(svc.GetVM().Memory.Segments); got != segments {
		t.Errorf("segments = %d after reloads, want %d", got, segments)
	}
}

func parse(t *testing.T, source string) *parser.Program {
	t.Helper()
	program, err := parser.NewParser(source, "test.s").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return program
}

func TestResetClearsBreakpointsAndWatchpoints(t *testing.T) {
	svc := newLoadedService(t, readerProgram)
	if err := svc.AddBreakpoint(0x8004); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddWatchpoint(svc.GetSymbols()["buffer"], "write"); err != nil {
		t.Fatal(err)
	}

	if err := svc.Reset(); err != nil {
		t.Fatal(err)
	}

	if bps := svc.GetBreakpoints(); len(bps) != 0 {
		t.Errorf("breakpoints after Reset = %v, want none", bps)
	}
	if wps := svc.GetWatchpoints(); len(wps) != 0 {
		t.Errorf("watchpoints after Reset = %v, want none", wps)
	}
}

func TestReloadAfterLowMemoryProgramUnmapsLowMemory(t *testing.T) {
	const lowProgram = ".org 0x0000\n_start:\n\tMOV R0, #0\n\tSWI #0x00\n"
	const nullRead = ".org 0x8000\n_start:\n\tMOV R1, #0\n\tLDR R0, [R1]\n\tSWI #0x00\n"
	svc := service.NewDebuggerService(vm.NewVM())
	if err := svc.LoadProgram(parse(t, lowProgram), 0); err != nil {
		t.Fatal(err)
	}

	loadSource(t, svc, nullRead)
	err := <-startRun(svc)

	if err == nil {
		t.Fatal("read of address 0 succeeded after reload, want a memory access violation")
	}
	if state := svc.GetExecutionState(); state != service.StateError {
		t.Errorf("state = %s, want error", state)
	}
}
