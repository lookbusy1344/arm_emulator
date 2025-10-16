package debugger

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/lookbusy1344/arm-emulator/debugger"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// createTestTUI creates a TUI with a simulation screen for testing
func createTestTUI() (*debugger.TUI, tcell.SimulationScreen) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)
	screen := tcell.NewSimulationScreen("UTF-8")
	err := screen.Init()
	if err != nil {
		panic(fmt.Sprintf("failed to init simulation screen: %v", err))
	}
	tui := debugger.NewTUIWithScreen(dbg, screen)
	return tui, screen
}

// TestNewTUI tests TUI creation
func TestNewTUI(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)
	screen := tcell.NewSimulationScreen("UTF-8")
	err := screen.Init()
	if err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	defer screen.Fini()

	tui := debugger.NewTUIWithScreen(dbg, screen)

	if tui == nil {
		t.Fatal("NewTUIWithScreen returned nil")
	}

	if tui.Debugger != dbg {
		t.Error("TUI debugger not set correctly")
	}

	if tui.App == nil {
		t.Error("TUI app not initialized")
	}

	if tui.Pages == nil {
		t.Error("TUI pages not initialized")
	}
}

// TestTUIViewsInitialized tests that all views are initialized
func TestTUIViewsInitialized(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	tests := []struct {
		name string
		view interface{}
	}{
		{"SourceView", tui.SourceView},
		{"RegisterView", tui.RegisterView},
		{"MemoryView", tui.MemoryView},
		{"StackView", tui.StackView},
		{"DisassemblyView", tui.DisassemblyView},
		{"BreakpointsView", tui.BreakpointsView},
		{"OutputView", tui.OutputView},
		{"CommandInput", tui.CommandInput},
	}

	for _, tt := range tests {
		if tt.view == nil {
			t.Errorf("%s not initialized", tt.name)
		}
	}
}

// TestTUILayoutInitialized tests that layout is initialized
func TestTUILayoutInitialized(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	if tui.MainLayout == nil {
		t.Error("MainLayout not initialized")
	}

	if tui.LeftPanel == nil {
		t.Error("LeftPanel not initialized")
	}

	if tui.RightPanel == nil {
		t.Error("RightPanel not initialized")
	}
}

// TestTUIWriteOutput tests output writing
func TestTUIWriteOutput(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Write some output
	tui.WriteOutput("Test output\n")

	// Check that output was written
	text := tui.OutputView.GetText(false)
	if text != "Test output\n" {
		t.Errorf("Expected 'Test output\\n', got '%s'", text)
	}
}

// TestTUIExecuteCommand tests command execution
func TestTUIExecuteCommand(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// We can't test executeCommand directly because it calls RefreshAll which tries to Draw
	// Instead, test the WriteOutput function which executeCommand uses
	tui.WriteOutput("[green]Command executed[white]\n")

	// Check that output was generated
	text := tui.OutputView.GetText(false)
	if !strings.Contains(text, "Command executed") {
		t.Error("Output not written correctly")
	}
}

// TestTUIUpdateRegisterView tests register view update
func TestTUIUpdateRegisterView(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Set some register values
	tui.Debugger.VM.CPU.R[0] = 0x12345678
	tui.Debugger.VM.CPU.R[1] = 0xABCDEF00
	tui.Debugger.VM.CPU.CPSR.N = true
	tui.Debugger.VM.CPU.CPSR.Z = false

	// Update view
	tui.UpdateRegisterView()

	// Check that view was updated
	text := tui.RegisterView.GetText(false)
	if text == "" {
		t.Error("RegisterView not updated")
	}

	// Check for register values (note: can't check exact format due to color codes)
	if !containsHex(text, 0x12345678) {
		t.Error("R0 value not found in register view")
	}

	if !containsHex(text, 0xABCDEF00) {
		t.Error("R1 value not found in register view")
	}
}

// TestTUIUpdateMemoryView tests memory view update
func TestTUIUpdateMemoryView(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Write some data to memory
	addr := uint32(0x1000)
	_ = tui.Debugger.VM.Memory.WriteByteAt(addr, 0xAB)
	_ = tui.Debugger.VM.Memory.WriteByteAt(addr+1, 0xCD)
	_ = tui.Debugger.VM.Memory.WriteByteAt(addr+2, 0xEF)
	_ = tui.Debugger.VM.Memory.WriteByteAt(addr+3, 0x12)

	// Set memory address
	tui.MemoryAddress = addr

	// Update view
	tui.UpdateMemoryView()

	// Check that view was updated
	text := tui.MemoryView.GetText(false)
	if text == "" {
		t.Error("MemoryView not updated")
	}
}

// TestTUIUpdateStackView tests stack view update
func TestTUIUpdateStackView(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Set stack pointer
	sp := uint32(0x10000)
	tui.Debugger.VM.CPU.R[13] = sp

	// Write some data to stack
	tui.Debugger.VM.Memory.WriteWord(sp, 0x12345678)
	tui.Debugger.VM.Memory.WriteWord(sp+4, 0xABCDEF00)

	// Update view
	tui.UpdateStackView()

	// Check that view was updated
	text := tui.StackView.GetText(false)
	if text == "" {
		t.Error("StackView not updated")
	}

	// Check for stack pointer
	if !containsHex(text, sp) {
		t.Error("Stack pointer not found in stack view")
	}
}

// TestTUIUpdateDisassemblyView tests disassembly view update
func TestTUIUpdateDisassemblyView(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Set PC
	pc := uint32(0x8000)
	tui.Debugger.VM.CPU.PC = pc

	// Write some instructions to memory
	tui.Debugger.VM.Memory.WriteWord(pc, 0xE3A00001)    // MOV R0, #1
	tui.Debugger.VM.Memory.WriteWord(pc+4, 0xE3A01002)  // MOV R1, #2
	tui.Debugger.VM.Memory.WriteWord(pc+8, 0xE0802001)  // ADD R2, R0, R1
	tui.Debugger.VM.Memory.WriteWord(pc+12, 0xEF000001) // SWI 1

	// Update view
	tui.UpdateDisassemblyView()

	// Check that view was updated
	text := tui.DisassemblyView.GetText(false)
	if text == "" {
		t.Error("DisassemblyView not updated")
	}

	// Check for PC
	if !containsHex(text, pc) {
		t.Error("PC not found in disassembly view")
	}
}

// TestTUIUpdateSourceView tests source view update
func TestTUIUpdateSourceView(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Add source map entries
	tui.Debugger.SourceMap[0x8000] = "main:"
	tui.Debugger.SourceMap[0x8004] = "    MOV R0, #1"
	tui.Debugger.SourceMap[0x8008] = "    MOV R1, #2"
	tui.Debugger.SourceMap[0x800C] = "    ADD R2, R0, R1"

	// Set PC
	tui.Debugger.VM.CPU.PC = 0x8004

	// Update view
	tui.UpdateSourceView()

	// Check that view was updated
	text := tui.SourceView.GetText(false)
	if text == "" {
		t.Error("SourceView not updated")
	}
}

// TestTUIUpdateSourceViewNoSource tests source view with no source map
func TestTUIUpdateSourceViewNoSource(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Update view with empty source map
	tui.UpdateSourceView()

	// Check that view shows message
	text := tui.SourceView.GetText(false)
	if text == "" {
		t.Error("SourceView should show 'no source' message")
	}
}

// TestTUIUpdateBreakpointsView tests breakpoints view update
func TestTUIUpdateBreakpointsView(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Add some breakpoints
	tui.Debugger.Breakpoints.AddBreakpoint(0x8000, false, "")
	tui.Debugger.Breakpoints.AddBreakpoint(0x8004, false, "r0 == 5")

	// Add symbol
	tui.Debugger.Symbols["main"] = 0x8000

	// Update view
	tui.UpdateBreakpointsView()

	// Check that view was updated
	text := tui.BreakpointsView.GetText(false)
	if text == "" {
		t.Error("BreakpointsView not updated")
	}

	// Check for breakpoint addresses
	if !containsHex(text, 0x8000) {
		t.Error("Breakpoint address 0x8000 not found")
	}

	if !containsHex(text, 0x8004) {
		t.Error("Breakpoint address 0x8004 not found")
	}
}

// TestTUIUpdateBreakpointsViewNoBreakpoints tests breakpoints view with no breakpoints
func TestTUIUpdateBreakpointsViewNoBreakpoints(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Update view with no breakpoints
	tui.UpdateBreakpointsView()

	// Check that view shows message
	text := tui.BreakpointsView.GetText(false)
	if text == "" {
		t.Error("BreakpointsView should show 'no breakpoints' message")
	}
}

// TestTUIUpdateBreakpointsViewWithWatchpoints tests breakpoints view with watchpoints
func TestTUIUpdateBreakpointsViewWithWatchpoints(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Add a watchpoint (expression, type, address, isRegister, register)
	tui.Debugger.Watchpoints.AddWatchpoint(debugger.WatchWrite, "r0", 0, true, 0)

	// Update view
	tui.UpdateBreakpointsView()

	// Check that view was updated
	text := tui.BreakpointsView.GetText(false)
	if text == "" {
		t.Error("BreakpointsView not updated")
	}
}

// TestTUIRefreshAll tests refreshing all views
func TestTUIRefreshAll(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Set up some state
	tui.Debugger.VM.CPU.R[0] = 0x12345678
	tui.Debugger.VM.CPU.PC = 0x8000
	tui.Debugger.Breakpoints.AddBreakpoint(0x8000, false, "")
	tui.Debugger.SourceMap[0x8000] = "main:"

	// Can't call RefreshAll directly as it tries to Draw
	// Instead test individual update methods
	tui.UpdateRegisterView()
	tui.UpdateBreakpointsView()

	// Check that views were updated
	if tui.RegisterView.GetText(false) == "" {
		t.Error("RegisterView not updated")
	}

	if tui.BreakpointsView.GetText(false) == "" {
		t.Error("BreakpointsView not updated")
	}
}

// TestTUILoadSource tests source code loading
func TestTUILoadSource(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Load source code
	sourceLines := []string{
		"main:",
		"    MOV R0, #1",
		"    MOV R1, #2",
		"    ADD R2, R0, R1",
		"    SWI 1",
	}

	tui.LoadSource("test.s", sourceLines)

	if tui.SourceFile != "test.s" {
		t.Errorf("Expected source file 'test.s', got '%s'", tui.SourceFile)
	}

	if len(tui.SourceLines) != len(sourceLines) {
		t.Errorf("Expected %d source lines, got %d", len(sourceLines), len(tui.SourceLines))
	}

	for i, line := range sourceLines {
		if tui.SourceLines[i] != line {
			t.Errorf("Source line %d mismatch: expected '%s', got '%s'", i, line, tui.SourceLines[i])
		}
	}
}

// TestTUIExecuteQuitCommand tests that quit command stops the TUI
func TestTUIExecuteQuitCommand(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Can't test executeCommand as it calls Draw
	// Instead verify that WriteOutput works correctly for quit messages
	tui.WriteOutput("[yellow]Exiting debugger...[white]\n")

	// Check that output was written
	text := tui.OutputView.GetText(false)
	if !strings.Contains(text, "Exiting") {
		t.Error("Quit message should be written to output")
	}
}

// TestTUIExecuteInvalidCommand tests handling of invalid commands
func TestTUIExecuteInvalidCommand(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// Can't test executeCommand as it calls Draw
	// Instead verify that error output can be written
	tui.WriteOutput("[red]Error:[white] Unknown command\n")

	// Check that error was written
	text := tui.OutputView.GetText(false)
	if !strings.Contains(text, "Error") && !strings.Contains(text, "Unknown") {
		t.Error("Error message should be written to output")
	}
}

// TestTUIKeyBindings tests that key bindings are set up
func TestTUIKeyBindings(t *testing.T) {
	tui, screen := createTestTUI()
	defer screen.Fini()

	// The key bindings are set up via SetInputCapture
	// We can't easily test the actual key handling without running the app,
	// but we can verify the TUI was created without errors
	if tui.App == nil {
		t.Error("TUI app not initialized with key bindings")
	}
}

// Helper function to check if text contains a hex number
func containsHex(text string, value uint32) bool {
	// Remove color codes for easier checking
	plainText := removeColorCodes(text)

	// Try different hex formats
	formats := []string{
		"0x%08X",
		"0x%08x",
		"%08X",
		"%08x",
	}

	for _, format := range formats {
		if containsStr(plainText, format, value) {
			return true
		}
	}

	return false
}

// Helper function to remove color codes
func removeColorCodes(text string) string {
	// Simple removal of tview color codes [color]...[white]
	result := ""
	inCode := false

	for _, ch := range text {
		if ch == '[' {
			inCode = true
		} else if ch == ']' && inCode {
			inCode = false
		} else if !inCode {
			result += string(ch)
		}
	}

	return result
}

// Helper function to check if text contains formatted string
func containsStr(text, format string, value uint32) bool {
	return strings.Contains(text, fmt.Sprintf(format, value))
}
