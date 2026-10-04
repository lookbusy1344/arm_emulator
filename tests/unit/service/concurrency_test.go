package service_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
)

const (
	spinProgram = `.org 0x8000
_start:
	MOV R0, #0
	LDR R2, =buf
loop:
	ADD R0, R0, #1
	STR R0, [R2]
	B loop
buf: .word 0
`
	readIntProgram = `.org 0x8000
_start:
	SWI #0x06
	MOV R1, #10
	SWI #0x03
	SWI #0x07
	MOV R0, #0
	SWI #0x00
`
	waitTimeout = 2 * time.Second
	pollPeriod  = time.Millisecond
)

func newLoadedService(t *testing.T, source string) *service.DebuggerService {
	t.Helper()
	svc := service.NewDebuggerService(vm.NewVM())
	loadSource(t, svc, source)
	return svc
}

func loadSource(t *testing.T, svc *service.DebuggerService, source string) {
	t.Helper()
	program, err := parser.NewParser(source, "test.s").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if err := svc.LoadProgram(program, 0x8000); err != nil {
		t.Fatalf("LoadProgram failed: %v", err)
	}
}

// startRun mirrors the API run handler: set running, then execute in a goroutine.
func startRun(svc *service.DebuggerService) <-chan error {
	svc.SetRunning(true)
	done := make(chan error, 1)
	go func() { done <- svc.RunUntilHalt() }()
	return done
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(waitTimeout)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(pollPeriod)
	}
}

func waitDone(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(waitTimeout):
		t.Fatal("execution did not finish")
		return nil
	}
}

func cyclesAdvanced(svc *service.DebuggerService) bool {
	return svc.GetRegisterState().Cycles > 100
}

// Run with -race: state queries must not race with the executing goroutine.
func TestRunConcurrentQueries(t *testing.T) {
	svc := newLoadedService(t, spinProgram)
	done := startRun(svc)
	waitFor(t, "execution to start", func() bool { return cyclesAdvanced(svc) })

	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 200 {
				_ = svc.GetRegisterState()
				_, _ = svc.GetMemory(0x8000, 64)
				_ = svc.GetStack(0, 8)
				_ = svc.GetDisassembly(0x8000, 8)
				_ = svc.GetExecutionState()
				_ = svc.GetLastMemoryWrite()
			}
		})
	}
	wg.Wait()

	svc.Pause()
	if err := waitDone(t, done); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}
}

func TestRunRejectsSecondExecution(t *testing.T) {
	svc := newLoadedService(t, spinProgram)
	done := startRun(svc)
	waitFor(t, "execution to start", func() bool { return cyclesAdvanced(svc) })

	if err := svc.RunUntilHalt(); !errors.Is(err, service.ErrExecutionInProgress) {
		t.Errorf("second RunUntilHalt: expected ErrExecutionInProgress, got %v", err)
	}
	if err := svc.Step(); !errors.Is(err, service.ErrExecutionInProgress) {
		t.Errorf("Step while running: expected ErrExecutionInProgress, got %v", err)
	}
	if err := svc.StepOver(); !errors.Is(err, service.ErrExecutionInProgress) {
		t.Errorf("StepOver while running: expected ErrExecutionInProgress, got %v", err)
	}

	svc.Pause()
	if err := waitDone(t, done); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}
}

func TestPauseStopsExecutionBeforeReturning(t *testing.T) {
	svc := newLoadedService(t, spinProgram)
	done := startRun(svc)
	waitFor(t, "execution to start", func() bool { return cyclesAdvanced(svc) })

	svc.Pause()

	if state := svc.GetExecutionState(); state != service.StateHalted {
		t.Errorf("expected halted after Pause, got %s", state)
	}
	before := svc.GetRegisterState().Cycles
	time.Sleep(10 * time.Millisecond)
	if after := svc.GetRegisterState().Cycles; after != before {
		t.Errorf("VM kept executing after Pause returned: cycles %d -> %d", before, after)
	}
	if err := waitDone(t, done); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}
}

func TestPauseInterruptsBlockedInputRead(t *testing.T) {
	svc := newLoadedService(t, readIntProgram)
	done := startRun(svc)
	waitFor(t, "input wait", func() bool { return svc.GetExecutionState() == service.StateWaitingForInput })

	paused := make(chan struct{})
	go func() {
		svc.Pause()
		close(paused)
	}()
	select {
	case <-paused:
	case <-time.After(waitTimeout):
		t.Fatal("Pause blocked while the guest waited for input")
	}

	if err := waitDone(t, done); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}
	if state := svc.GetExecutionState(); state != service.StateHalted {
		t.Errorf("expected halted, got %s", state)
	}
	// The read did not complete, so PC stays on the SWI.
	if pc := svc.GetRegisterState().PC; pc != 0x8000 {
		t.Errorf("expected PC to stay on the read SWI at 0x8000, got 0x%08X", pc)
	}
}

func TestStepBlockedOnInputCanBeInterrupted(t *testing.T) {
	svc := newLoadedService(t, readIntProgram)

	stepped := make(chan error, 1)
	go func() { stepped <- svc.Step() }()
	waitFor(t, "input wait", func() bool { return svc.GetExecutionState() == service.StateWaitingForInput })

	svc.Pause()
	if err := waitDone(t, stepped); !errors.Is(err, vm.ErrInputInterrupted) {
		t.Errorf("expected ErrInputInterrupted from interrupted Step, got %v", err)
	}
}

func TestResetWhileRunning(t *testing.T) {
	svc := newLoadedService(t, spinProgram)
	done := startRun(svc)
	waitFor(t, "execution to start", func() bool { return cyclesAdvanced(svc) })

	if err := svc.Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
	if err := waitDone(t, done); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}

	regs := svc.GetRegisterState()
	if regs.PC != 0 || regs.Cycles != 0 {
		t.Errorf("expected pristine VM after Reset, got PC=0x%08X cycles=%d", regs.PC, regs.Cycles)
	}
	if svc.IsRunning() {
		t.Error("expected not running after Reset")
	}
}

func TestLoadProgramWhileRunning(t *testing.T) {
	svc := newLoadedService(t, spinProgram)
	done := startRun(svc)
	waitFor(t, "execution to start", func() bool { return cyclesAdvanced(svc) })

	loadSource(t, svc, ".org 0x8000\n_start:\nMOV R0, #42\nSWI #0x00\n")
	if err := waitDone(t, done); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}

	done = startRun(svc)
	if err := waitDone(t, done); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}
	if r0 := svc.GetRegisterState().Registers[0]; r0 != 42 {
		t.Errorf("expected R0=42 from the new program, got %d", r0)
	}
}

func TestInputDeliveredWhileWaiting(t *testing.T) {
	svc := newLoadedService(t, readIntProgram)
	done := startRun(svc)
	waitFor(t, "input wait", func() bool { return svc.GetExecutionState() == service.StateWaitingForInput })

	if err := svc.SendInput("42"); err != nil {
		t.Fatalf("SendInput failed: %v", err)
	}
	if err := waitDone(t, done); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}
	// Live input is echoed, then the program prints the value.
	if out := svc.GetOutput(); out != "42\n42\n" {
		t.Errorf("unexpected output %q", out)
	}
}

func TestInputBufferedBeforeRun(t *testing.T) {
	svc := newLoadedService(t, readIntProgram)

	if err := svc.SendInput("7\n"); err != nil {
		t.Fatalf("SendInput failed: %v", err)
	}
	if err := waitDone(t, startRun(svc)); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}
	if out := svc.GetOutput(); out != "7\n" {
		t.Errorf("unexpected output %q", out)
	}
}

func TestInputAfterReset(t *testing.T) {
	svc := service.NewDebuggerService(vm.NewVM())
	if err := svc.Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
	loadSource(t, svc, readIntProgram)

	done := startRun(svc)
	waitFor(t, "input wait", func() bool { return svc.GetExecutionState() == service.StateWaitingForInput })
	if err := svc.SendInput("42"); err != nil {
		t.Fatalf("SendInput failed: %v", err)
	}
	if err := waitDone(t, done); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}
	if out := svc.GetOutput(); out != "42\n42\n" {
		t.Errorf("unexpected output %q", out)
	}
}

func TestCloseStopsExecution(t *testing.T) {
	svc := newLoadedService(t, spinProgram)
	done := startRun(svc)
	waitFor(t, "execution to start", func() bool { return cyclesAdvanced(svc) })

	svc.Close()
	if err := waitDone(t, done); err != nil {
		t.Fatalf("RunUntilHalt failed: %v", err)
	}
	if err := svc.Step(); !errors.Is(err, service.ErrClosed) {
		t.Errorf("Step after Close: expected ErrClosed, got %v", err)
	}
}
