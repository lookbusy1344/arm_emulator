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

func TestFinishCommandReturnsToCaller(t *testing.T) {
	const (
		caller    = 0x8000
		callee    = 0x8100
		returnTo  = caller + 4
		blCallee  = 0xEB00003E // BL 0x8100 from 0x8000
		movR0     = 0xE3A00001 // MOV R0, #1
		movPCLR   = 0xE1A0F00E // MOV PC, LR
		calleeNop = 0xE1A00000 // MOV R0, R0
	)
	machine := vm.NewVM()
	for _, seg := range machine.Memory.Segments {
		if seg.Name == "code" {
			seg.Permissions |= vm.PermWrite
		}
	}
	for addr, word := range map[uint32]uint32{
		caller: blCallee, returnTo: movR0, callee: calleeNop, callee + 4: movPCLR,
	} {
		if err := machine.Memory.WriteWord(addr, word); err != nil {
			t.Fatal(err)
		}
	}
	machine.CPU.PC = caller
	if err := machine.Step(); err != nil {
		t.Fatal(err)
	}
	if machine.CPU.PC != callee {
		t.Fatalf("expected BL to reach 0x%X, got 0x%08X", callee, machine.CPU.PC)
	}

	dbg := debugger.NewDebugger(machine)
	if err := dbg.ExecuteCommand("finish"); err != nil {
		t.Fatal(err)
	}
	const maxSteps = 10
	for range maxSteps {
		if stop, _ := dbg.ShouldBreak(); stop {
			break
		}
		if err := machine.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if machine.CPU.PC != returnTo {
		t.Fatalf("expected finish to stop at 0x%X, got 0x%08X", returnTo, machine.CPU.PC)
	}
}
