package debugger

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestGUICreation tests that the GUI can be created without errors
func TestGUICreation(t *testing.T) {
	// Create a simple test program
	source := `
_start:
    MOV R0, #42
    SWI #0x00
`
	p := parser.NewParser(source, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse test program: %v", err)
	}

	// Create VM
	machine := vm.NewVM()
	if err := machine.LoadProgram(program, 0x8000); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Create debugger
	dbg := NewDebugger(machine)

	// Create GUI (this should not panic or error)
	gui := newGUI(dbg)
	if gui == nil {
		t.Fatal("GUI creation returned nil")
	}

	// Verify GUI components are initialized
	if gui.SourceView == nil {
		t.Error("SourceView not initialized")
	}
	if gui.RegisterView == nil {
		t.Error("RegisterView not initialized")
	}
	if gui.MemoryView == nil {
		t.Error("MemoryView not initialized")
	}
	if gui.StackView == nil {
		t.Error("StackView not initialized")
	}
	if gui.BreakpointsList == nil {
		t.Error("BreakpointsList not initialized")
	}
	if gui.ConsoleOutput == nil {
		t.Error("ConsoleOutput not initialized")
	}
	if gui.Toolbar == nil {
		t.Error("Toolbar not initialized")
	}

	// Clean up
	if gui.App != nil {
		gui.App.Quit()
	}
}

// TestGUIViewUpdates tests that views can be updated
func TestGUIViewUpdates(t *testing.T) {
	// Create test program
	source := `
_start:
    MOV R0, #5
    MOV R1, #10
    ADD R2, R0, R1
    SWI #0x00
`
	p := parser.NewParser(source, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse test program: %v", err)
	}

	// Create VM
	machine := vm.NewVM()
	if err := machine.LoadProgram(program, 0x8000); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Create debugger and GUI
	dbg := NewDebugger(machine)
	gui := newGUI(dbg)
	defer gui.App.Quit()

	// Update views (should not panic)
	gui.updateRegisters()
	gui.updateMemory()
	gui.updateStack()
	gui.updateBreakpoints()
	gui.updateSource()

	// Verify register view has content
	registerText := gui.RegisterView.Text()
	if len(registerText) == 0 {
		t.Error("Register view is empty")
	}

	// Verify memory view has content
	memoryText := gui.MemoryView.Text()
	if len(memoryText) == 0 {
		t.Error("Memory view is empty")
	}

	// Verify stack view has content
	stackText := gui.StackView.Text()
	if len(stackText) == 0 {
		t.Error("Stack view is empty")
	}
}

// TestGUIBreakpointManagement tests breakpoint operations
func TestGUIBreakpointManagement(t *testing.T) {
	// Create test program
	source := `
_start:
    MOV R0, #1
    MOV R1, #2
    MOV R2, #3
    SWI #0x00
`
	p := parser.NewParser(source, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse test program: %v", err)
	}

	// Create VM
	machine := vm.NewVM()
	if err := machine.LoadProgram(program, 0x8000); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Create debugger and GUI
	dbg := NewDebugger(machine)
	gui := newGUI(dbg)
	defer gui.App.Quit()

	// Initially no breakpoints
	if len(gui.breakpoints) != 0 {
		t.Errorf("Expected 0 breakpoints, got %d", len(gui.breakpoints))
	}

	// Add a breakpoint
	gui.addBreakpoint()
	gui.updateBreakpoints()

	// Should have one breakpoint now
	if len(gui.breakpoints) != 1 {
		t.Errorf("Expected 1 breakpoint after adding, got %d", len(gui.breakpoints))
	}

	// Clear all breakpoints
	gui.clearBreakpoints()

	// Should have zero breakpoints again
	if len(gui.breakpoints) != 0 {
		t.Errorf("Expected 0 breakpoints after clearing, got %d", len(gui.breakpoints))
	}
}

// TestGUIStepExecution tests single-step execution
func TestGUIStepExecution(t *testing.T) {
	// Create test program
	source := `
_start:
    MOV R0, #42
    MOV R1, #100
    SWI #0x00
`
	p := parser.NewParser(source, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse test program: %v", err)
	}

	// Create VM
	machine := vm.NewVM()
	if err := machine.LoadProgram(program, 0x8000); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Create debugger and GUI
	dbg := NewDebugger(machine)
	gui := newGUI(dbg)
	defer gui.App.Quit()

	// Record initial PC
	initialPC := machine.CPU.PC

	// Execute one step
	gui.stepProgram()

	// PC should have advanced
	if machine.CPU.PC == initialPC {
		t.Error("PC did not advance after step")
	}

	// R0 should be 42 after first instruction
	if machine.CPU.R[0] != 42 {
		t.Errorf("Expected R0=42, got R0=%d", machine.CPU.R[0])
	}
}

// TestGUIWithTestDriver demonstrates using Fyne's test driver
func TestGUIWithTestDriver(t *testing.T) {
	// Create test program
	source := `
_start:
    MOV R0, #1
    SWI #0x00
`
	p := parser.NewParser(source, "test.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse test program: %v", err)
	}

	// Create VM
	machine := vm.NewVM()
	if err := machine.LoadProgram(program, 0x8000); err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Create debugger
	dbg := NewDebugger(machine)

	// Use Fyne's test app instead of real app
	testApp := test.NewApp()
	defer testApp.Quit()

	// Create GUI components manually with test app
	gui := &GUI{
		Debugger:    dbg,
		App:         testApp,
		breakpoints: []string{},
	}

	gui.initializeViews()

	// Verify views are created
	if gui.SourceView == nil {
		t.Error("SourceView not created")
	}
	if gui.RegisterView == nil {
		t.Error("RegisterView not created")
	}

	// Test view updates
	gui.updateRegisters()
	text := gui.RegisterView.Text()
	if len(text) == 0 {
		t.Error("Register view has no content")
	}

	// Verify register values are shown
	if !containsString(text, "R0:") {
		t.Error("Register view does not contain R0")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
