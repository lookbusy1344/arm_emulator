package debugger

import (
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lookbusy1344/arm-emulator/debugger"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// Tests for TUI thread safety (CODE_REVIEW_OPUS.md section 4.1 and 7 Phase 1)
// These tests verify that concurrent access to TUI shared state is properly synchronized

// createTestTUIForConcurrency creates a TUI suitable for concurrency testing
func createTestTUIForConcurrency(t *testing.T) (*debugger.TUI, tcell.SimulationScreen) {
	t.Helper()
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	tui := debugger.NewTUIWithScreen(dbg, screen)
	return tui, screen
}

// TestTUI_ConcurrentRegisterCapture tests concurrent calls to CaptureRegisterState
func TestTUI_ConcurrentRegisterCapture(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	const numGoroutines = 10
	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				// Capture state (uses Lock) - tests TUI synchronization
				tui.CaptureRegisterState()
			}
		}()
	}

	wg.Wait()
}

// TestTUI_ConcurrentRegisterDetect tests concurrent calls to DetectRegisterChanges
func TestTUI_ConcurrentRegisterDetect(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	const numGoroutines = 10
	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				// Detect changes (uses Lock) - tests TUI synchronization
				tui.DetectRegisterChanges()
			}
		}()
	}

	wg.Wait()
}

// TestTUI_ConcurrentMemoryCapture tests concurrent calls to CaptureMemoryTraceState
func TestTUI_ConcurrentMemoryCapture(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	const numGoroutines = 10
	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				tui.CaptureMemoryTraceState()
			}
		}()
	}

	wg.Wait()
}

// TestTUI_ConcurrentMemoryDetect tests concurrent calls to DetectMemoryWrites
func TestTUI_ConcurrentMemoryDetect(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	const numGoroutines = 10
	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				tui.DetectMemoryWrites()
			}
		}()
	}

	wg.Wait()
}

// TestTUI_ConcurrentCaptureAndDetect tests concurrent capture and detect operations
func TestTUI_ConcurrentCaptureAndDetect(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(4)

	// Goroutine 1: CaptureRegisterState
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.CaptureRegisterState()
		}
	}()

	// Goroutine 2: DetectRegisterChanges
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.DetectRegisterChanges()
		}
	}()

	// Goroutine 3: CaptureMemoryTraceState
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.CaptureMemoryTraceState()
		}
	}()

	// Goroutine 4: DetectMemoryWrites
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.DetectMemoryWrites()
		}
	}()

	wg.Wait()
}

// TestTUI_ConcurrentUpdateAndCapture tests concurrent update (read) and capture (write) operations
// This simulates the real TUI scenario where:
// - Background goroutine captures/detects state (writers)
// - Main UI thread updates views (readers)
// Note: VM state is accessed sequentially from setup, not concurrently
func TestTUI_ConcurrentUpdateAndCapture(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(4)

	// Background goroutine 1: Capture and detect (writers)
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.CaptureRegisterState()
			tui.DetectRegisterChanges()
		}
	}()

	// Background goroutine 2: Memory capture and detect (writers)
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.CaptureMemoryTraceState()
			tui.DetectMemoryWrites()
		}
	}()

	// Main UI thread simulation 1: UpdateRegisterView (reader)
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.UpdateRegisterView()
		}
	}()

	// Main UI thread simulation 2: UpdateMemoryView (reader)
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.UpdateMemoryView()
		}
	}()

	wg.Wait()
}

// TestTUI_ConcurrentStackView tests concurrent stack view updates
func TestTUI_ConcurrentStackView(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	const numGoroutines = 5
	const numIterations = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				tui.UpdateStackView()
			}
		}()
	}

	wg.Wait()
}

// TestTUI_SimulateExecutionLoop simulates the actual execution loop pattern
// from executeUntilBreak where a background goroutine updates state while
// the main thread refreshes views
// Note: Focuses on TUI state synchronization, VM state is set up beforehand
func TestTUI_SimulateExecutionLoop(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	// Set up initial VM state (sequential, no races)
	tui.Debugger.VM.CPU.R[0] = 0x12345678
	tui.Debugger.VM.CPU.R[1] = 0xABCDEF00
	tui.Debugger.VM.CPU.PC = 0x8000

	const numCycles = 50

	var wg sync.WaitGroup
	wg.Add(2)

	// Simulated background execution goroutine (like executeUntilBreak)
	go func() {
		defer wg.Done()
		for cycle := 0; cycle < numCycles; cycle++ {
			// Capture and detect like executeUntilBreak does
			tui.CaptureRegisterState()
			tui.DetectRegisterChanges()
			tui.CaptureMemoryTraceState()
			tui.DetectMemoryWrites()

			// Small sleep to simulate instruction execution
			time.Sleep(time.Microsecond)
		}
	}()

	// Simulated main UI thread (RefreshAll pattern)
	go func() {
		defer wg.Done()
		for cycle := 0; cycle < numCycles; cycle++ {
			// Update all views like RefreshAll does
			tui.UpdateRegisterView()
			tui.UpdateMemoryView()
			tui.UpdateStackView()
			tui.UpdateDisassemblyView()
			tui.UpdateSourceView()
			tui.UpdateBreakpointsView()

			// Small sleep to simulate UI refresh rate
			time.Sleep(time.Microsecond)
		}
	}()

	wg.Wait()
}

// TestTUI_RapidRegisterChanges tests rapid register changes with concurrent reads
// Note: This test focuses on TUI state synchronization, not VM state
func TestTUI_RapidRegisterChanges(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	const numIterations = 200

	var wg sync.WaitGroup
	wg.Add(3)

	// Writer goroutine: rapidly capture and detect (TUI operations)
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.CaptureRegisterState()
			tui.DetectRegisterChanges()
		}
	}()

	// Reader goroutine 1
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.UpdateRegisterView()
		}
	}()

	// Reader goroutine 2
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.UpdateRegisterView()
		}
	}()

	wg.Wait()
}

// TestTUI_CPSRConcurrency tests concurrent CPSR-related TUI state access
// Note: Focuses on TUI state synchronization for ChangedCPSR tracking
func TestTUI_CPSRConcurrency(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	// Set up initial CPSR state (sequential, no races)
	tui.Debugger.VM.CPU.CPSR.N = true
	tui.Debugger.VM.CPU.CPSR.Z = false
	tui.Debugger.VM.CPU.CPSR.C = true
	tui.Debugger.VM.CPU.CPSR.V = false

	const numIterations = 200

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer: capture and detect (accesses TUI's PrevCPSR and ChangedCPSR)
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.CaptureRegisterState()
			tui.DetectRegisterChanges()
		}
	}()

	// Reader: update register view (reads ChangedCPSR state)
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.UpdateRegisterView()
		}
	}()

	wg.Wait()
}

// TestTUI_MemoryWritesConcurrency tests concurrent memory write detection
// Note: Focuses on TUI state synchronization for RecentWrites tracking
func TestTUI_MemoryWritesConcurrency(t *testing.T) {
	tui, screen := createTestTUIForConcurrency(t)
	defer screen.Fini()

	const numIterations = 200

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer: capture and detect memory writes (TUI operations)
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.CaptureMemoryTraceState()
			tui.DetectMemoryWrites()
		}
	}()

	// Reader: update memory and stack views
	go func() {
		defer wg.Done()
		for i := 0; i < numIterations; i++ {
			tui.UpdateMemoryView()
			tui.UpdateStackView()
		}
	}()

	wg.Wait()
}
