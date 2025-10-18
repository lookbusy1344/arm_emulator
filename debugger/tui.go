package debugger

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// TUI represents the text user interface for the debugger
type TUI struct {
	// Core components
	Debugger *Debugger
	App      *tview.Application
	Pages    *tview.Pages

	// Layout containers
	MainLayout *tview.Flex
	LeftPanel  *tview.Flex
	RightPanel *tview.Flex

	// View panels
	SourceView      *tview.TextView
	RegisterView    *tview.TextView
	MemoryView      *tview.TextView
	StackView       *tview.TextView
	DisassemblyView *tview.TextView
	BreakpointsView *tview.TextView
	OutputView      *tview.TextView
	CommandInput    *tview.InputField

	// State
	CurrentAddress uint32
	MemoryAddress  uint32
	StackAddress   uint32
	Running        bool

	// Source code cache
	SourceLines []string
	SourceFile  string
}

// tuiWriter redirects VM output to the TUI OutputView
type tuiWriter struct {
	tui *TUI
}

// Write implements io.Writer interface
func (w *tuiWriter) Write(p []byte) (n int, err error) {
	w.tui.App.QueueUpdateDraw(func() {
		_, _ = w.tui.OutputView.Write(p) // Ignore write errors in TUI
		w.tui.OutputView.ScrollToEnd()
	})
	return len(p), nil
}

// NewTUI creates a new text user interface
func NewTUI(debugger *Debugger) *TUI {
	return NewTUIWithScreen(debugger, nil)
}

// NewTUIWithScreen creates a new text user interface with an optional screen
// If screen is nil, a default screen will be created by tview
func NewTUIWithScreen(debugger *Debugger, screen tcell.Screen) *TUI {
	app := tview.NewApplication()
	if screen != nil {
		app.SetScreen(screen)
	}

	tui := &TUI{
		Debugger:       debugger,
		App:            app,
		CurrentAddress: 0,
		MemoryAddress:  0,
		StackAddress:   0,
		Running:        false,
	}

	tui.initializeViews()
	tui.buildLayout()
	tui.setupKeyBindings()

	// Redirect VM output to TUI OutputView
	debugger.VM.OutputWriter = &tuiWriter{tui: tui}

	return tui
}

// initializeViews creates all the view panels
func (t *TUI) initializeViews() {
	// Source View
	t.SourceView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)
	t.SourceView.SetBorder(true).SetTitle(" Source ")

	// Register View
	t.RegisterView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false)
	t.RegisterView.SetBorder(true).SetTitle(" Registers ")

	// Memory View
	t.MemoryView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)
	t.MemoryView.SetBorder(true).SetTitle(" Memory ")

	// Stack View
	t.StackView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)
	t.StackView.SetBorder(true).SetTitle(" Stack ")

	// Disassembly View
	t.DisassemblyView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)
	t.DisassemblyView.SetBorder(true).SetTitle(" Disassembly ")

	// Breakpoints View
	t.BreakpointsView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)
	t.BreakpointsView.SetBorder(true).SetTitle(" Breakpoints/Watchpoints ")

	// Output View
	t.OutputView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	t.OutputView.SetBorder(true).SetTitle(" Output ")

	// Command Input
	t.CommandInput = tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0)
	t.CommandInput.SetBorder(true).SetTitle(" Command ")
	t.CommandInput.SetDoneFunc(t.handleCommand)
}

// buildLayout constructs the TUI layout
func (t *TUI) buildLayout() {
	// Left panel: Source and Disassembly
	t.LeftPanel = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.SourceView, 0, 3, false).
		AddItem(t.DisassemblyView, 0, 2, false)

	// Right panel top: Registers, Memory, Stack
	rightTop := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.RegisterView, 10, 0, false).
		AddItem(t.MemoryView, 0, 1, false).
		AddItem(t.StackView, 0, 1, false)

	// Right panel: Top + Breakpoints
	t.RightPanel = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(rightTop, 0, 3, false).
		AddItem(t.BreakpointsView, 8, 0, false)

	// Main content: Left and Right panels
	mainContent := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(t.LeftPanel, 0, 2, false).
		AddItem(t.RightPanel, 0, 1, false)

	// Main layout: Content + Output + Command
	t.MainLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(mainContent, 0, 4, false).
		AddItem(t.OutputView, 8, 0, false).
		AddItem(t.CommandInput, 3, 0, true)

	// Create pages for potential dialogs/modals
	t.Pages = tview.NewPages().
		AddPage("main", t.MainLayout, true, true)
}

// setupKeyBindings sets up keyboard shortcuts
func (t *TUI) setupKeyBindings() {
	// Global key handler
	t.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			go t.executeCommand("help")
			return nil
		case tcell.KeyF5:
			go t.executeCommand("continue")
			return nil
		case tcell.KeyF9:
			go t.executeCommand("break")
			return nil
		case tcell.KeyF10:
			go t.executeCommand("next")
			return nil
		case tcell.KeyF11:
			go t.executeCommand("step")
			return nil
		case tcell.KeyCtrlC:
			t.App.Stop()
			return nil
		case tcell.KeyCtrlL:
			t.App.QueueUpdateDraw(func() {
				t.RefreshAll()
			})
			return nil
		}
		return event
	})
}

// handleCommand processes command input
func (t *TUI) handleCommand(key tcell.Key) {
	if key == tcell.KeyEnter {
		cmd := t.CommandInput.GetText()
		if cmd != "" {
			go t.executeCommand(cmd)
			t.CommandInput.SetText("")
		}
	}
}

// executeCommand executes a debugger command
func (t *TUI) executeCommand(cmd string) {
	// Clear previous output
	t.Debugger.Output.Reset()

	// Check for quit/exit commands
	cmdLower := strings.ToLower(strings.TrimSpace(cmd))
	if cmdLower == "quit" || cmdLower == "q" || cmdLower == "exit" {
		go t.App.QueueUpdateDraw(func() {
			t.WriteOutput("[yellow]Exiting debugger...[white]\n")
		})
		t.App.Stop()
		return
	}

	// Execute command
	err := t.Debugger.ExecuteCommand(cmd)

	// Get output
	output := t.Debugger.GetOutput()

	// Display output and refresh (in goroutine to avoid blocking)
	go t.App.QueueUpdateDraw(func() {
		if err != nil {
			t.WriteOutput(fmt.Sprintf("[red]Error:[white] %v\n", err))
		}
		if output != "" {
			t.WriteOutput(output)
		}
		t.RefreshAll()
	})

	// If running, execute until breakpoint or halt
	if t.Debugger.Running {
		t.executeUntilBreak()
	}
}

// executeUntilBreak runs the VM until a breakpoint is hit or the program halts
func (t *TUI) executeUntilBreak() {
	// Run execution in the background to keep TUI responsive
	go func() {
		for t.Debugger.Running {
			// Check for breakpoint before execution
			if shouldBreak, reason := t.Debugger.ShouldBreak(); shouldBreak {
				t.Debugger.Running = false
				t.App.QueueUpdateDraw(func() {
					t.WriteOutput(fmt.Sprintf("[yellow]Stopped:[white] %s at PC=0x%08X\n", reason, t.Debugger.VM.CPU.PC))
					t.RefreshAll()
				})
				break
			}

			// Execute one step
			if err := t.Debugger.VM.Step(); err != nil {
				if t.Debugger.VM.State == vm.StateHalted {
					t.Debugger.Running = false
					t.App.QueueUpdateDraw(func() {
						t.WriteOutput(fmt.Sprintf("[green]Program exited with code %d[white]\n", t.Debugger.VM.ExitCode))
						t.RefreshAll()
					})
					break
				}
				t.Debugger.Running = false
				t.App.QueueUpdateDraw(func() {
					t.WriteOutput(fmt.Sprintf("[red]Runtime error:[white] %v\n", err))
					t.RefreshAll()
				})
				break
			}

			// Update display periodically during long runs
			// (every 100 instructions to keep display responsive)
			if t.Debugger.VM.CPU.Cycles%100 == 0 {
				t.App.QueueUpdateDraw(func() {
					t.RefreshAll()
				})
			}
		}

		// Final refresh when execution stops
		t.App.QueueUpdateDraw(func() {
			t.RefreshAll()
		})
	}()
}

// WriteOutput writes to the output view
func (t *TUI) WriteOutput(text string) {
	_, _ = t.OutputView.Write([]byte(text)) // Ignore write errors in TUI
	t.OutputView.ScrollToEnd()
}

// RefreshAll refreshes all view panels
// Note: This should be called from within App.QueueUpdateDraw or the main event loop
func (t *TUI) RefreshAll() {
	t.UpdateSourceView()
	t.UpdateRegisterView()
	t.UpdateMemoryView()
	t.UpdateStackView()
	t.UpdateDisassemblyView()
	t.UpdateBreakpointsView()
	// Note: App.Draw() is not called here - caller must use QueueUpdateDraw
}

// UpdateSourceView updates the source code view
func (t *TUI) UpdateSourceView() {
	t.SourceView.Clear()

	// If no source map, show message
	if len(t.Debugger.SourceMap) == 0 {
		t.SourceView.SetText("[yellow]No source code available[white]")
		return
	}

	// Get current PC
	pc := t.Debugger.VM.CPU.PC

	// Find source lines around current PC
	var lines []string
	startAddr := pc - 20 // Show 10 instructions before
	if startAddr > pc {  // Handle underflow
		startAddr = 0
	}

	for addr := startAddr; addr < pc+40; addr += 4 {
		if sourceLine, exists := t.Debugger.SourceMap[addr]; exists {
			// Highlight current line
			marker := "  "
			color := "white"
			if addr == pc {
				marker = "->"
				color = "yellow"
			}

			// Check for breakpoint
			if t.Debugger.Breakpoints.GetBreakpoint(addr) != nil {
				marker = "* "
			}

			line := fmt.Sprintf("[%s]%s 0x%08X: %s[white]", color, marker, addr, sourceLine)
			lines = append(lines, line)
		}
	}

	t.SourceView.SetText(strings.Join(lines, "\n"))
}

// UpdateRegisterView updates the register view
func (t *TUI) UpdateRegisterView() {
	t.RegisterView.Clear()

	cpu := t.Debugger.VM.CPU
	var lines []string

	// General purpose registers (4 columns)
	for i := 0; i < 4; i++ {
		var cols []string
		for j := 0; j < 4; j++ {
			reg := i*4 + j
			name := fmt.Sprintf("R%-2d", reg)
			var value uint32
			if reg == 15 {
				name = "PC "
				value = cpu.PC
			} else if reg == 13 {
				name = "SP "
				value = cpu.R[reg]
			} else if reg == 14 {
				name = "LR "
				value = cpu.R[reg]
			} else {
				value = cpu.R[reg]
			}
			cols = append(cols, fmt.Sprintf("%s: 0x%08X", name, value))
		}
		lines = append(lines, strings.Join(cols, "  "))
	}

	lines = append(lines, "")

	// CPSR flags
	flags := ""
	if cpu.CPSR.N {
		flags += "[red]N[white]"
	} else {
		flags += "n"
	}
	if cpu.CPSR.Z {
		flags += "[blue]Z[white]"
	} else {
		flags += "z"
	}
	if cpu.CPSR.C {
		flags += "[green]C[white]"
	} else {
		flags += "c"
	}
	if cpu.CPSR.V {
		flags += "[yellow]V[white]"
	} else {
		flags += "v"
	}

	// Calculate CPSR value manually
	cpsrValue := uint32(0)
	if cpu.CPSR.N {
		cpsrValue |= 0x80000000
	}
	if cpu.CPSR.Z {
		cpsrValue |= 0x40000000
	}
	if cpu.CPSR.C {
		cpsrValue |= 0x20000000
	}
	if cpu.CPSR.V {
		cpsrValue |= 0x10000000
	}

	lines = append(lines, fmt.Sprintf("CPSR: 0x%08X  Flags: %s", cpsrValue, flags))
	lines = append(lines, fmt.Sprintf("Cycles: %d", cpu.Cycles))

	t.RegisterView.SetText(strings.Join(lines, "\n"))
}

// UpdateMemoryView updates the memory view
func (t *TUI) UpdateMemoryView() {
	t.MemoryView.Clear()

	// Use current memory address or PC if not set
	addr := t.MemoryAddress
	if addr == 0 {
		addr = t.Debugger.VM.CPU.PC
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("[yellow]Address: 0x%08X[white]", addr))

	// Show 16 rows of 16 bytes each
	for row := 0; row < 16; row++ {
		rowOffset, err := vm.SafeIntToUint32(row * 16)
		if err != nil {
			break // Should never happen
		}
		rowAddr := addr + rowOffset

		// Address
		line := fmt.Sprintf("0x%08X: ", rowAddr)

		// Hex bytes
		var hexBytes []string
		var asciiBytes []byte

		for col := 0; col < 16; col++ {
			colOffset, err := vm.SafeIntToUint32(col)
			if err != nil {
				break // Should never happen
			}
			byteAddr := rowAddr + colOffset
			b, err := t.Debugger.VM.Memory.ReadByteAt(byteAddr)
			if err != nil {
				hexBytes = append(hexBytes, "??")
				asciiBytes = append(asciiBytes, '.')
			} else {
				hexBytes = append(hexBytes, fmt.Sprintf("%02X", b))
				if b >= 32 && b < 127 {
					asciiBytes = append(asciiBytes, b)
				} else {
					asciiBytes = append(asciiBytes, '.')
				}
			}
		}

		line += strings.Join(hexBytes, " ") + "  " + string(asciiBytes)
		lines = append(lines, line)
	}

	t.MemoryView.SetText(strings.Join(lines, "\n"))
}

// UpdateStackView updates the stack view
func (t *TUI) UpdateStackView() {
	t.StackView.Clear()

	sp := t.Debugger.VM.CPU.R[13] // Stack pointer

	var lines []string
	lines = append(lines, fmt.Sprintf("[yellow]Stack Pointer: 0x%08X[white]", sp))

	// Show 16 words (64 bytes) from stack
	for i := 0; i < 16; i++ {
		offset, err := vm.SafeIntToUint32(i * 4)
		if err != nil {
			break // Should never happen
		}
		addr := sp + offset

		// Read word
		word, err := t.Debugger.VM.Memory.ReadWord(addr)
		if err != nil {
			lines = append(lines, fmt.Sprintf("0x%08X: ????????", addr))
			continue
		}

		// Mark current SP
		marker := "  "
		if addr == sp {
			marker = "->"
		}

		line := fmt.Sprintf("%s 0x%08X: 0x%08X", marker, addr, word)

		// Try to resolve as symbol
		if sym := t.findSymbolForAddress(word); sym != "" {
			line += fmt.Sprintf(" <%s>", sym)
		}

		lines = append(lines, line)
	}

	t.StackView.SetText(strings.Join(lines, "\n"))
}

// UpdateDisassemblyView updates the disassembly view
func (t *TUI) UpdateDisassemblyView() {
	t.DisassemblyView.Clear()

	pc := t.Debugger.VM.CPU.PC

	var lines []string

	// Show 16 instructions around PC
	startAddr := pc - 32 // 8 instructions before
	if startAddr > pc {  // Handle underflow
		startAddr = 0
	}

	for i := 0; i < 16; i++ {
		offset, err := vm.SafeIntToUint32(i * 4)
		if err != nil {
			break // Should never happen
		}
		addr := startAddr + offset

		// Read instruction
		instr, err := t.Debugger.VM.Memory.ReadWord(addr)
		if err != nil {
			continue
		}

		// Highlight current instruction
		marker := "  "
		color := "white"
		if addr == pc {
			marker = "->"
			color = "yellow"
		}

		// Check for breakpoint
		if t.Debugger.Breakpoints.GetBreakpoint(addr) != nil {
			marker = "* "
		}

		// Simple disassembly (just show hex for now)
		line := fmt.Sprintf("[%s]%s 0x%08X: 0x%08X[white]", color, marker, addr, instr)

		// Try to add symbol
		if sym := t.findSymbolForAddress(addr); sym != "" {
			line = fmt.Sprintf("[%s]%s 0x%08X: 0x%08X  <%s>[white]", color, marker, addr, instr, sym)
		}

		lines = append(lines, line)
	}

	t.DisassemblyView.SetText(strings.Join(lines, "\n"))
}

// UpdateBreakpointsView updates the breakpoints and watchpoints view
func (t *TUI) UpdateBreakpointsView() {
	t.BreakpointsView.Clear()

	var lines []string

	// Breakpoints
	bps := t.Debugger.Breakpoints.GetAllBreakpoints()
	if len(bps) > 0 {
		lines = append(lines, "[yellow]Breakpoints:[white]")
		for _, bp := range bps {
			status := "enabled"
			color := "green"
			if !bp.Enabled {
				status = "disabled"
				color = "red"
			}

			line := fmt.Sprintf("  %d: [%s]%s[white] 0x%08X", bp.ID, color, status, bp.Address)

			// Add symbol if available
			if sym := t.findSymbolForAddress(bp.Address); sym != "" {
				line += fmt.Sprintf(" <%s>", sym)
			}

			// Add condition if present
			if bp.Condition != "" {
				line += fmt.Sprintf(" if %s", bp.Condition)
			}

			// Add hit count
			line += fmt.Sprintf(" (hits: %d)", bp.HitCount)

			lines = append(lines, line)
		}
	} else {
		lines = append(lines, "[yellow]No breakpoints set[white]")
	}

	lines = append(lines, "")

	// Watchpoints
	wps := t.Debugger.Watchpoints.GetAllWatchpoints()
	if len(wps) > 0 {
		lines = append(lines, "[yellow]Watchpoints:[white]")
		for _, wp := range wps {
			typeStr := "watch"
			if wp.Type == WatchRead {
				typeStr = "rwatch"
			} else if wp.Type == WatchReadWrite {
				typeStr = "awatch"
			}

			line := fmt.Sprintf("  %d: %s %s = 0x%08X", wp.ID, typeStr, wp.Expression, wp.LastValue)
			lines = append(lines, line)
		}
	}

	t.BreakpointsView.SetText(strings.Join(lines, "\n"))
}

// findSymbolForAddress finds a symbol name for an address
func (t *TUI) findSymbolForAddress(addr uint32) string {
	for sym, symAddr := range t.Debugger.Symbols {
		if symAddr == addr {
			return sym
		}
	}
	return ""
}

// Run starts the TUI application
func (t *TUI) Run() error {
	// Set up initial content (but don't call Draw yet - app isn't running)
	t.UpdateSourceView()
	t.UpdateRegisterView()
	t.UpdateMemoryView()
	t.UpdateStackView()
	t.UpdateDisassemblyView()
	t.UpdateBreakpointsView()

	// Show welcome message
	t.WriteOutput("[green]ARM Emulator Debugger TUI[white]\n")
	t.WriteOutput("Press F1 for help, F5 to continue, F10 to step over, F11 to step\n")
	t.WriteOutput("Type 'help' for command list\n\n")

	// Run the application (this will handle drawing)
	return t.App.SetRoot(t.Pages, true).SetFocus(t.CommandInput).Run()
}

// Stop stops the TUI application
func (t *TUI) Stop() {
	t.App.Stop()
}

// LoadSource loads source code for display
func (t *TUI) LoadSource(filename string, lines []string) {
	t.SourceFile = filename
	t.SourceLines = lines
	t.UpdateSourceView()
}
