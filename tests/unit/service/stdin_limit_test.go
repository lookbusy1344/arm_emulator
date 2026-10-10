package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/service"
)

func TestSendInputQueueLimit(t *testing.T) {
	svc := newLoadedService(t, readerProgram)

	if err := svc.SendInput(strings.Repeat("a", service.MaxQueuedInput)); err != nil {
		t.Fatalf("input at the limit: %v", err)
	}
	if err := svc.SendInput("b"); !errors.Is(err, service.ErrInputQueueFull) {
		t.Fatalf("input over the limit: err = %v, want %v", err, service.ErrInputQueueFull)
	}

	// Loading a program drops queued input, so the queue accepts input again.
	loadSource(t, svc, readerProgram)
	if err := svc.SendInput(strings.Repeat("a", service.MaxQueuedInput)); err != nil {
		t.Fatalf("input after reload: %v", err)
	}
}

// Input sent during execution gains a newline, which counts against the limit.
func TestSendInputLimitCountsAddedNewline(t *testing.T) {
	svc := newLoadedService(t, readerProgram)
	svc.SetRunning(true)
	t.Cleanup(func() { svc.SetRunning(false) })

	if err := svc.SendInput(strings.Repeat("a", service.MaxQueuedInput)); !errors.Is(err, service.ErrInputQueueFull) {
		t.Fatalf("input filling the queue before its newline: err = %v, want %v", err, service.ErrInputQueueFull)
	}
	if err := svc.SendInput(strings.Repeat("a", service.MaxQueuedInput-1)); err != nil {
		t.Fatalf("input plus newline at the limit: %v", err)
	}
}
