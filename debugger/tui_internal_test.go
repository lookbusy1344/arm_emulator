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
