package vm_test

import (
	"errors"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

const (
	ldrR0FromR1 = 0xE5910000 // LDR R0, [R1]
	swiExit     = 0xEF000000
	swiBreak    = 0xEF0000F1
)

// A debugger single-steps from Halted (after load) or Breakpoint (after a stop).
var steppingStates = map[string]vm.ExecutionState{
	"halted":     vm.StateHalted,
	"breakpoint": vm.StateBreakpoint,
	"running":    vm.StateRunning,
}

func TestSteppedFaultEntersErrorState(t *testing.T) {
	for name, state := range steppingStates {
		t.Run(name, func(t *testing.T) {
			v := vm.NewVM()
			v.State = state
			v.CPU.R[1] = 0 // unmapped
			v.CPU.PC = vm.CodeSegmentStart
			if err := v.Memory.WriteWord(vm.CodeSegmentStart, ldrR0FromR1); err != nil {
				t.Fatal(err)
			}
			if err := v.Step(); err == nil {
				t.Fatal("Step() succeeded, want fault")
			}
			if v.State != vm.StateError {
				t.Errorf("State = %v, want StateError", v.State)
			}
			if v.LastError == nil {
				t.Error("LastError not set")
			}
			if err := v.Step(); err == nil {
				t.Error("second Step() after a fault succeeded")
			}
		})
	}
}

func TestSteppedExitAndBreakpointKeepTheirStates(t *testing.T) {
	for name, state := range steppingStates {
		t.Run(name, func(t *testing.T) {
			v := vm.NewVM()
			v.State = state
			v.CPU.R[0] = 7
			v.CPU.PC = vm.CodeSegmentStart
			if err := v.Memory.WriteWord(vm.CodeSegmentStart, swiExit); err != nil {
				t.Fatal(err)
			}
			err := v.Step()
			if !errors.Is(err, vm.ErrProgramExited) {
				t.Fatalf("Step() error = %v, want ErrProgramExited", err)
			}
			if v.State != vm.StateHalted || v.ExitCode != 7 || v.LastError != nil {
				t.Errorf("State=%v ExitCode=%d LastError=%v, want Halted, 7, nil", v.State, v.ExitCode, v.LastError)
			}

			v = vm.NewVM()
			v.State = state
			v.CPU.PC = vm.CodeSegmentStart
			if err := v.Memory.WriteWord(vm.CodeSegmentStart, swiBreak); err != nil {
				t.Fatal(err)
			}
			err = v.Step()
			if !errors.Is(err, vm.ErrBreakpointHit) {
				t.Fatalf("Step() error = %v, want ErrBreakpointHit", err)
			}
			if v.State != vm.StateBreakpoint || v.LastError != nil {
				t.Errorf("State=%v LastError=%v, want Breakpoint, nil", v.State, v.LastError)
			}
		})
	}
}
