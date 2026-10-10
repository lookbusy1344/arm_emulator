package service_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// restartAndRun mirrors the API run handler on a halted program.
func restartAndRun(t *testing.T, svc *service.DebuggerService) {
	t.Helper()
	if err := svc.ResetToEntryPoint(); err != nil {
		t.Fatalf("restart: %v", err)
	}
	runToHalt(t, svc)
}

// Each run increments a counter in program memory and exits with its value.
const counterProgram = `.org 0x8000
_start:
	LDR R1, =count
	LDR R0, [R1]
	ADD R0, R0, #1
	STR R0, [R1]
	SWI #0x00
count:
	.word 0
`

func TestRestartRestoresProgramMemory(t *testing.T) {
	svc := newLoadedService(t, counterProgram)
	runToHalt(t, svc)
	if code := svc.GetExitCode(); code != 1 {
		t.Fatalf("first run: exit code %d, want 1", code)
	}

	restartAndRun(t, svc)

	if code := svc.GetExitCode(); code != 1 {
		t.Errorf("second run: exit code %d, want 1 (counter kept from the first run)", code)
	}
}

// Each run allocates 24 KB and exits with the address.
const allocProgram = `.org 0x8000
_start:
	LDR R0, =0x6000
	SWI #0x20
	SWI #0x00
`

func TestRestartResetsHeap(t *testing.T) {
	svc := newLoadedService(t, allocProgram)
	for run := 1; run <= 3; run++ {
		if run == 1 {
			runToHalt(t, svc)
		} else {
			restartAndRun(t, svc)
		}
		if code := uint32(svc.GetExitCode()); code != vm.HeapSegmentStart {
			t.Errorf("run %d: allocation at 0x%08X, want 0x%08X", run, code, uint32(vm.HeapSegmentStart))
		}
	}
}

// Each run opens a file without closing it and exits with the descriptor.
const openProgram = `.org 0x8000
_start:
	LDR R0, =name
	MOV R1, #1
	SWI #0x10
	SWI #0x00
name:
	.asciz "out.txt"
`

func TestRestartClosesGuestFiles(t *testing.T) {
	machine := vm.NewVM()
	machine.FilesystemRoot = t.TempDir()
	svc := service.NewDebuggerService(machine)
	loadSource(t, svc, openProgram)

	runToHalt(t, svc)
	if code := uint32(svc.GetExitCode()); code != vm.FirstUserFD {
		t.Fatalf("first run: fd %d, want %d", code, vm.FirstUserFD)
	}

	restartAndRun(t, svc)

	if code := uint32(svc.GetExitCode()); code != vm.FirstUserFD {
		t.Errorf("second run: fd %d, want %d (descriptor from the first run still open)", code, vm.FirstUserFD)
	}
}

// Reads an integer and exits with it.
const readIntExitProgram = `.org 0x8000
_start:
	SWI #0x06
	SWI #0x00
`

func TestRestartKeepsQueuedInput(t *testing.T) {
	svc := newLoadedService(t, readIntExitProgram)
	if err := svc.SendInput("5\n"); err != nil {
		t.Fatal(err)
	}
	runToHalt(t, svc)
	if code := svc.GetExitCode(); code != 5 {
		t.Fatalf("first run: exit code %d, want 5", code)
	}

	if err := svc.SendInput("9\n"); err != nil {
		t.Fatal(err)
	}
	restartAndRun(t, svc)

	if code := svc.GetExitCode(); code != 9 {
		t.Errorf("second run: exit code %d, want 9", code)
	}
}

func TestRestartKeepsBreakpoints(t *testing.T) {
	svc := newLoadedService(t, counterProgram)
	const addOffset = 8 // third instruction: ADD
	bp := uint32(0x8000 + addOffset)
	if err := svc.AddBreakpoint(bp); err != nil {
		t.Fatal(err)
	}

	if err := svc.ResetToEntryPoint(); err != nil {
		t.Fatal(err)
	}

	bps := svc.GetBreakpoints()
	if len(bps) != 1 || bps[0].Address != bp {
		t.Errorf("breakpoints after restart = %+v, want one at 0x%08X", bps, bp)
	}
}
