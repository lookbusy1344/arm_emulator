package debugger_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/debugger"
	"github.com/lookbusy1344/arm-emulator/vm"
)

const resumeBreakAddr = 0x8004

func debuggerAtBreakpoint(t *testing.T) *debugger.Debugger {
	t.Helper()
	dbg := debugger.NewDebugger(vm.NewVM())
	dbg.VM.CPU.PC = resumeBreakAddr
	dbg.Breakpoints.AddBreakpoint(resumeBreakAddr, false, "")
	if stop, _ := dbg.ShouldBreak(); !stop {
		t.Fatal("expected to stop at breakpoint")
	}
	return dbg
}

func TestResumeSkipsBreakpointAtCurrentPCOnce(t *testing.T) {
	dbg := debuggerAtBreakpoint(t)

	dbg.ResumeFromCurrentPC()
	if stop, reason := dbg.ShouldBreak(); stop {
		t.Fatalf("resume stopped again at the same breakpoint: %s", reason)
	}
	// The skip applies to one check only; reaching the address again stops.
	if stop, _ := dbg.ShouldBreak(); !stop {
		t.Error("expected breakpoint to fire on the next visit")
	}
}

func TestResumeDoesNotSkipBreakpointElsewhere(t *testing.T) {
	dbg := debuggerAtBreakpoint(t)
	dbg.ResumeFromCurrentPC()

	const other = 0x8010
	dbg.Breakpoints.AddBreakpoint(other, false, "")
	dbg.VM.CPU.PC = other
	if stop, _ := dbg.ShouldBreak(); !stop {
		t.Error("expected breakpoint at a different address to fire")
	}
}

func TestContinueCommandResumesPastBreakpoint(t *testing.T) {
	dbg := debuggerAtBreakpoint(t)
	dbg.VM.State = vm.StateBreakpoint

	if err := dbg.ExecuteCommand("continue"); err != nil {
		t.Fatalf("continue failed: %v", err)
	}
	if stop, reason := dbg.ShouldBreak(); stop {
		t.Fatalf("continue stopped again at the same breakpoint: %s", reason)
	}
}

func TestRunCommandStopsAtEntryBreakpoint(t *testing.T) {
	dbg := debugger.NewDebugger(vm.NewVM())
	dbg.VM.EntryPoint = resumeBreakAddr
	dbg.Breakpoints.AddBreakpoint(resumeBreakAddr, false, "")

	if err := dbg.ExecuteCommand("run"); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if stop, _ := dbg.ShouldBreak(); !stop {
		t.Error("expected run to stop at a breakpoint on the entry point")
	}
}
