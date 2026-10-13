package debugger

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestExecuteCommandAsync tests that executeCommand doesn't block
// This is an internal test that can access unexported methods
func TestExecuteCommandAsync(t *testing.T) {
	machine := vm.NewVM()
	dbg := NewDebugger(machine)
	screen := tcell.NewSimulationScreen("UTF-8")
	err := screen.Init()
	if err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	defer screen.Fini()

	tui := NewTUIWithScreen(dbg, screen)

	// Execute a command in a goroutine (like the real TUI does)
	done := make(chan bool, 1)
	go func() {
		tui.executeCommand("help")
		done <- true
	}()

	// Wait a reasonable time for command to complete
	// If it blocks, this will timeout
	select {
	case <-done:
		// Success - command completed
	case <-time.After(time.Second * 2):
		t.Fatal("executeCommand blocked for more than 2 seconds - deadlock detected")
	}
}

// TestHandleCommandAsync tests that handleCommand doesn't block
// This is an internal test that can access unexported methods
func TestHandleCommandAsync(t *testing.T) {
	machine := vm.NewVM()
	dbg := NewDebugger(machine)
	screen := tcell.NewSimulationScreen("UTF-8")
	err := screen.Init()
	if err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	defer screen.Fini()

	tui := NewTUIWithScreen(dbg, screen)

	// Set a command in the input field
	tui.CommandInput.SetText("help")

	// Call handleCommand (which should spawn a goroutine)
	done := make(chan bool, 1)
	go func() {
		tui.handleCommand(tcell.KeyEnter)
		done <- true
	}()

	// handleCommand itself should not block - just spawn the goroutine
	select {
	case <-done:
		// Success - handleCommand returned immediately
	case <-time.After(time.Millisecond * 100):
		t.Fatal("handleCommand blocked for more than 100ms - should return immediately")
	}
}

// TestHandleProgramInputEchoing tests that program input is echoed to OutputView
func TestHandleProgramInputEchoing(t *testing.T) {
	machine := vm.NewVM()
	dbg := NewDebugger(machine)
	screen := tcell.NewSimulationScreen("UTF-8")
	err := screen.Init()
	if err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	defer screen.Fini()

	tui := NewTUIWithScreen(dbg, screen)

	// Set input text
	testInput := "test input 123"
	tui.ProgramInput.SetText(testInput)

	// Call handleProgramInput
	tui.handleProgramInput(tcell.KeyEnter)

	// Wait a moment for the goroutine to write
	time.Sleep(10 * time.Millisecond)

	// Check that the input was echoed to OutputView
	outputText := tui.OutputView.GetText(false)
	expectedEcho := testInput + "\n"
	if outputText != expectedEcho {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedEcho, outputText)
	}

	// Verify the input field was cleared
	if tui.ProgramInput.GetText() != "" {
		t.Error("ProgramInput should be cleared after handling input")
	}
}

func newSimulatedTUI(t *testing.T) *TUI {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	t.Cleanup(screen.Fini)
	return NewTUIWithScreen(NewDebugger(vm.NewVM()), screen)
}

func pressKey(tui *TUI, key tcell.Key) {
	tui.App.GetInputCapture()(tcell.NewEventKey(key, 0, tcell.ModNone))
}

// TestF9TogglesBreakpointAtPC tests that F9 sets, then clears, a breakpoint at the PC
func TestF9TogglesBreakpointAtPC(t *testing.T) {
	const pc = 0x8010
	tui := newSimulatedTUI(t)
	tui.Debugger.VM.CPU.PC = pc

	pressKey(tui, tcell.KeyF9)
	if !tui.Debugger.Breakpoints.HasBreakpoint(pc) {
		t.Fatalf("F9 should set a breakpoint at PC 0x%08X", pc)
	}

	pressKey(tui, tcell.KeyF9)
	if tui.Debugger.Breakpoints.HasBreakpoint(pc) {
		t.Fatalf("second F9 should clear the breakpoint at PC 0x%08X", pc)
	}
}

// TestCtrlCPausesRunningProgram tests that Ctrl+C pauses a running program instead of quitting
func TestCtrlCPausesRunningProgram(t *testing.T) {
	tui := newSimulatedTUI(t)
	tui.Debugger.SetRunning(true)
	tui.Debugger.SetStepMode(StepOut)

	pressKey(tui, tcell.KeyCtrlC)

	if tui.Debugger.IsRunning() {
		t.Fatal("Ctrl+C should pause a running program")
	}
	if mode := tui.Debugger.GetStepMode(); mode != StepNone {
		t.Fatalf("Ctrl+C should clear the step mode, got %d", mode)
	}
}

// TestEmptyEnterRepeatsLastCommand tests that Enter on an empty command line is passed on,
// so the debugger repeats the last command
func TestEmptyEnterRepeatsLastCommand(t *testing.T) {
	tui := newSimulatedTUI(t)
	var dispatched []string
	tui.dispatch = func(cmd string) { dispatched = append(dispatched, cmd) }

	tui.CommandInput.SetText("")
	tui.handleCommand(tcell.KeyEnter)

	if len(dispatched) != 1 || dispatched[0] != "" {
		t.Fatalf("expected one empty command dispatched, got %q", dispatched)
	}
}

// TestExecuteEmptyCommandRepeatsLast tests that an empty command runs the previous one again
func TestExecuteEmptyCommandRepeatsLast(t *testing.T) {
	tui := newSimulatedTUI(t)

	tui.executeCommand("set r0 = r0+1")
	tui.executeCommand("")

	if got := tui.Debugger.VM.CPU.GetRegister(0); got != 2 {
		t.Fatalf("expected r0 = 2 after repeating the increment, got %d", got)
	}
}

// TestCommandHistoryRecall tests that Up and Down walk the command history in the command field
func TestCommandHistoryRecall(t *testing.T) {
	tui := newSimulatedTUI(t)
	tui.executeCommand("print 1")
	tui.executeCommand("print 2")
	capture := tui.CommandInput.GetInputCapture()
	if capture == nil {
		t.Fatal("command input has no key handler")
	}
	press := func(key tcell.Key) string {
		capture(tcell.NewEventKey(key, 0, tcell.ModNone))
		return tui.CommandInput.GetText()
	}

	steps := []struct {
		key  tcell.Key
		want string
	}{
		{tcell.KeyUp, "print 2"},
		{tcell.KeyUp, "print 1"},
		{tcell.KeyUp, "print 1"}, // oldest entry stays
		{tcell.KeyDown, "print 2"},
		{tcell.KeyDown, ""}, // past the newest entry clears the field
	}
	for i, step := range steps {
		if got := press(step.key); got != step.want {
			t.Fatalf("step %d: expected %q, got %q", i, step.want, got)
		}
	}
}
