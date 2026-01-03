package api

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestProcessMonitor_Initialization(t *testing.T) {
	shutdownCalled := false
	shutdown := func() { shutdownCalled = true }

	monitor := NewProcessMonitor(shutdown)

	if monitor.parentPID != os.Getppid() {
		t.Errorf("Expected parent PID %d, got %d", os.Getppid(), monitor.parentPID)
	}

	if monitor.checkInterval != 2*time.Second {
		t.Errorf("Expected check interval 2s, got %v", monitor.checkInterval)
	}

	if monitor.shutdownFunc == nil {
		t.Error("Expected shutdown function to be set")
	}

	if monitor.stopChan == nil {
		t.Error("Expected stop channel to be initialized")
	}

	if shutdownCalled {
		t.Error("Shutdown should not be called during initialization")
	}
}

func TestProcessMonitor_GracefulStop(t *testing.T) {
	shutdownCalled := false
	shutdown := func() { shutdownCalled = true }

	monitor := NewProcessMonitor(shutdown)
	monitor.Start()

	// Give monitor time to start
	time.Sleep(100 * time.Millisecond)

	// Stop the monitor
	monitor.Stop()

	// Give monitor time to stop
	time.Sleep(100 * time.Millisecond)

	if shutdownCalled {
		t.Error("Shutdown should not be called when stopping gracefully")
	}
}

func TestProcessMonitor_ShutdownCallback(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	shutdownCalled := false
	var mu sync.Mutex

	shutdown := func() {
		mu.Lock()
		shutdownCalled = true
		mu.Unlock()
		wg.Done()
	}

	monitor := NewProcessMonitor(shutdown)

	// Override check interval for faster testing
	monitor.checkInterval = 10 * time.Millisecond

	// Simulate parent death by changing the stored parent PID
	// In real scenarios, the OS changes the PPID when parent dies
	monitor.parentPID = 99999 // Non-existent PID

	monitor.Start()

	// Wait for shutdown to be called (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - shutdown was called
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for shutdown callback")
	}

	mu.Lock()
	defer mu.Unlock()
	if !shutdownCalled {
		t.Error("Expected shutdown to be called when parent PID changes")
	}
}

func TestProcessMonitor_MultipleStops(t *testing.T) {
	shutdown := func() {}

	monitor := NewProcessMonitor(shutdown)
	monitor.Start()

	time.Sleep(50 * time.Millisecond)

	// Stop multiple times should not panic
	monitor.Stop()
	monitor.Stop()
	monitor.Stop()
}

func TestProcessMonitor_StopBeforeStart(t *testing.T) {
	shutdown := func() {}

	monitor := NewProcessMonitor(shutdown)

	// Stop before start should not panic
	monitor.Stop()
}
