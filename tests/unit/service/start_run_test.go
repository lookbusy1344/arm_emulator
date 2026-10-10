package service_test

import (
	"errors"
	"testing"

	"github.com/lookbusy1344/arm-emulator/service"
)

func TestStartRunRejectsSecondRun(t *testing.T) {
	svc := newLoadedService(t, spinProgram)
	done := make(chan error, 1)
	if err := svc.StartRun(func() {}, func(err error) { done <- err }); err != nil {
		t.Fatal(err)
	}

	err := svc.StartRun(func() { t.Error("second run started") }, func(error) {})

	if !errors.Is(err, service.ErrExecutionInProgress) {
		t.Errorf("second StartRun = %v, want ErrExecutionInProgress", err)
	}
	svc.Pause()
	if err := <-done; err != nil {
		t.Errorf("first run: %v", err)
	}
}

func TestStartRunAfterCloseFails(t *testing.T) {
	svc := newLoadedService(t, spinProgram)
	svc.Close()

	err := svc.StartRun(func() { t.Error("run started") }, func(error) {})

	if !errors.Is(err, service.ErrClosed) {
		t.Errorf("StartRun after Close = %v, want ErrClosed", err)
	}
}

func TestStartRunRestartsHaltedProgram(t *testing.T) {
	svc := newLoadedService(t, counterProgram)
	for run := 1; run <= 2; run++ {
		var order []string
		done := make(chan struct{})
		err := svc.StartRun(
			func() { order = append(order, "start") },
			func(err error) {
				order = append(order, "done")
				if err != nil {
					t.Errorf("run %d: %v", run, err)
				}
				close(done)
			})
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		<-done

		if code := svc.GetExitCode(); code != 1 {
			t.Errorf("run %d: exit code %d, want 1", run, code)
		}
		if len(order) != 2 || order[0] != "start" || order[1] != "done" {
			t.Errorf("run %d: callbacks %v, want [start done]", run, order)
		}
	}
}
