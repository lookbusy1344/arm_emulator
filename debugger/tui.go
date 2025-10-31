package debugger

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// helper: small max for ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

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
	StatusView      *tview.TextView // Status messages (breakpoints, stepping, errors)
	OutputView      *tview.TextView // Program output only
	CommandInput    *tview.InputField

	// Focus management
	focusables []tview.Primitive
	focusIndex int

	// State
	CurrentAddress uint32
	MemoryAddress  uint32
	StackAddress   uint32
	Running        bool

	// Source code cache
	SourceLines []string
	SourceFile  string

	// Register tracking for highlighting changes
	PrevRegisters [16]uint32   // Previous values of R0-R15 before last step
	PrevCPSR      vm.CPSR      // Previous CPSR flags before last step
	ChangedRegs   map[int]bool // Registers that changed in the last step
	ChangedCPSR   bool         // CPSR changed in the last step

	// Memory write tracking for highlighting changes
	RecentWrites        map[uint32]bool // Memory addresses written in the last step
	LastTraceEntryCount int             // Number of memory trace entries before last step
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
	// Enable mouse to allow scrolling in scrollable views (e.g., Source)
	app.EnableMouse(true)

	tui := &TUI{
		Debugger:       debugger,
		App:            app,
		CurrentAddress: 0,
		MemoryAddress:  0,
		StackAddress:   0,
		Running:        false,
		ChangedRegs:    make(map[int]bool),
		RecentWrites:   make(map[uint32]bool),
	}

	tui.initializeViews()
	tui.buildLayout()
	// Setup focus chain before key bindings
	tui.initFocusChain()
	tui.setupKeyBindings()

	// Redirect VM output to TUI OutputView
	debugger.VM.OutputWriter = &tuiWriter{tui: tui}

	// Enable MemoryTrace for tracking memory writes in the TUI
	if debugger.VM.MemoryTrace == nil {
		debugger.VM.MemoryTrace = vm.NewMemoryTrace(nil) // nil writer - we only need tracking, not output
	}
	debugger.VM.MemoryTrace.Enabled = true
	debugger.VM.MemoryTrace.Start()

	return tui
}

// initializeViews creates all the view panels
func (t *TUI) initializeViews() {
	// Source View
	t.SourceView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true)
	t.SourceView.SetBorder(true).SetTitle(" Source ")
	// Keyboard scrolling for Source view
	t.SourceView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			row, col := t.SourceView.GetScrollOffset()
			if row > 0 {
				row--
			}
			t.SourceView.ScrollTo(row, col)
			return nil
		case tcell.KeyDown:
			row, col := t.SourceView.GetScrollOffset()
			row++
			t.SourceView.ScrollTo(row, col)
			return nil
		case tcell.KeyPgUp:
			_, _, _, h := t.SourceView.GetRect()
			row, col := t.SourceView.GetScrollOffset()
			row -= max(1, h-1)
			if row < 0 {
				row = 0
			}
			t.SourceView.ScrollTo(row, col)
			return nil
		case tcell.KeyPgDn:
			_, _, _, h := t.SourceView.GetRect()
			row, col := t.SourceView.GetScrollOffset()
			row += max(1, h-1)
			t.SourceView.ScrollTo(row, col)
			return nil
		case tcell.KeyHome:
			t.SourceView.ScrollToBeginning()
			return nil
		case tcell.KeyEnd:
			t.SourceView.ScrollToEnd()
			return nil
		}
		return event
	})

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

	// Status View - for debugger messages
	t.StatusView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	t.StatusView.SetBorder(true).SetTitle(" Status ")

	// Output View - for program output only
	t.OutputView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	t.OutputView.SetBorder(true).SetTitle(" Program Output ")

	// Command Input
	t.CommandInput = tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0)
	t.CommandInput.SetBorder(true).SetTitle(" Command ")
	t.CommandInput.SetDoneFunc(t.handleCommand)
}

// buildLayout constructs the TUI layout
func (t *TUI) buildLayout() {
	// Left panel: Source, Disassembly, and Status
	t.LeftPanel = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.SourceView, 0, 3, false).      // Source gets flex weight 3 (more space)
		AddItem(t.DisassemblyView, 0, 1, false). // Disassembly gets flex weight 1 (less space)
		AddItem(t.StatusView, 4, 0, false)       // Fixed height for status messages

	// Right panel top: Registers, Memory, Stack
	rightTop := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.RegisterView, 9, 0, false). // Fixed height: 5 rows of regs + blank + status line + border = 9
		AddItem(t.MemoryView, 0, 3, false).   // Memory gets flex weight 3
		AddItem(t.StackView, 0, 2, false)     // Stack gets flex weight 2

	// Right panel: Top + Breakpoints (dynamic height based on content)
	t.RightPanel = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(rightTop, 0, 3, false).
		AddItem(t.BreakpointsView, 4, 0, false) // Start with minimal height, updated dynamically

	// Main content: Left and Right panels
	// Left panel (Source) is wider (flex 3), Right panel (Registers) is narrower (flex 2)
	mainContent := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(t.LeftPanel, 0, 3, false). // Source window gets more width (flex weight 3)
		AddItem(t.RightPanel, 0, 2, false) // Registers window gets less width (flex weight 2)

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

// initFocusChain sets the order of focusable widgets for Tab navigation
func (t *TUI) initFocusChain() {
	// Only views we want to focus with Tab (source, disasm, memory, stack, breakpoints, output, command)
	t.focusables = []tview.Primitive{
		t.SourceView,
		t.DisassemblyView,
		t.MemoryView,
		t.StackView,
		t.BreakpointsView,
		t.OutputView,
		t.CommandInput,
	}
	// Start focusing command input by default (matches Run())
	t.focusIndex = len(t.focusables) - 1
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
		case tcell.KeyF6:
			// Center current PC in Source and Disassembly views
			t.scrollPCIntoView()
			return nil
		case tcell.KeyCtrlC:
			t.App.Stop()
			return nil
		case tcell.KeyCtrlL:
			t.App.QueueUpdateDraw(func() {
				t.RefreshAll()
			})
			return nil
		case tcell.KeyTAB:
			// Cycle focus forward
			t.focusIndex = (t.focusIndex + 1) % len(t.focusables)
			t.App.SetFocus(t.focusables[t.focusIndex])
			return nil
		case tcell.KeyBacktab:
			// Shift+Tab cycles backward
			t.focusIndex = (t.focusIndex - 1 + len(t.focusables)) % len(t.focusables)
			t.App.SetFocus(t.focusables[t.focusIndex])
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
			t.WriteStatus("[yellow]Exiting debugger...[white]\n")
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
			t.WriteStatus(fmt.Sprintf("[red]Error:[white] %v\n", err))
		}
		if output != "" {
			t.WriteStatus(output)
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
			// For single-step mode, execute instruction first before checking if we should break
			// For other modes, check breakpoints before execution
			if t.Debugger.StepMode != StepSingle {
				if shouldBreak, reason := t.Debugger.ShouldBreak(); shouldBreak {
					t.Debugger.Running = false
					t.App.QueueUpdateDraw(func() {
						t.WriteStatus(fmt.Sprintf("[yellow]Stopped:[white] %s at PC=0x%08X\n", reason, t.Debugger.VM.CPU.PC))
						t.DetectRegisterChanges()
						t.DetectMemoryWrites()
						t.RefreshAll()
					})
					break
				}
			}

			// Capture register state before executing
			t.CaptureRegisterState()
			t.CaptureMemoryTraceState()

			// Execute one step
			if err := t.Debugger.VM.Step(); err != nil {
				if t.Debugger.VM.State == vm.StateHalted {
					t.Debugger.Running = false
					t.App.QueueUpdateDraw(func() {
						t.WriteStatus(fmt.Sprintf("[green]Program exited with code %d[white]\n", t.Debugger.VM.ExitCode))
						t.DetectRegisterChanges()
						t.DetectMemoryWrites()
						t.RefreshAll()
					})
					break
				}
				t.Debugger.Running = false
				t.App.QueueUpdateDraw(func() {
					t.WriteStatus(fmt.Sprintf("[red]Runtime error:[white] %v\n", err))
					t.DetectRegisterChanges()
					t.DetectMemoryWrites()
					t.RefreshAll()
				})
				break
			}

			// Detect register and memory changes after execution
			t.DetectRegisterChanges()
			t.DetectMemoryWrites()

			// For single-step mode, check if we should break after execution
			if t.Debugger.StepMode == StepSingle {
				if shouldBreak, reason := t.Debugger.ShouldBreak(); shouldBreak {
					t.Debugger.Running = false
					t.App.QueueUpdateDraw(func() {
						t.WriteStatus(fmt.Sprintf("[yellow]Stopped:[white] %s at PC=0x%08X\n", reason, t.Debugger.VM.CPU.PC))
						t.DetectRegisterChanges()
						t.DetectMemoryWrites()
						t.RefreshAll()
					})
					break
				}
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

// WriteOutput writes to the output view (program output only)
func (t *TUI) WriteOutput(text string) {
	_, _ = t.OutputView.Write([]byte(text)) // Ignore write errors in TUI
	t.OutputView.ScrollToEnd()
}

// WriteStatus writes to the status view (debugger messages)
func (t *TUI) WriteStatus(text string) {
	_, _ = t.StatusView.Write([]byte(text)) // Ignore write errors in TUI
	t.StatusView.ScrollToEnd()
}

// RefreshAll refreshes all view panels
// Note: This should be called from within App.QueueUpdateDraw or the main event loop

// scrollPCIntoView scrolls Source and Disassembly views so current PC line is visible and roughly centered.
func (t *TUI) scrollPCIntoView() {
	// Source view
	src := t.SourceView.GetText(true)
	if src != "" {
		rows := strings.Split(src, "\n")
		pcRow := -1
		for i, r := range rows {
			if strings.Contains(r, "->") {
				pcRow = i
				break
			}
		}
		if pcRow >= 0 {
			_, _, _, h := t.SourceView.GetRect()
			row := pcRow - h/2
			if row < 0 {
				row = 0
			}
			_, col := t.SourceView.GetScrollOffset()
			t.SourceView.ScrollTo(row, col)
		}
	}
	// Disassembly view
	dis := t.DisassemblyView.GetText(true)
	if dis != "" {
		drows := strings.Split(dis, "\n")
		dpcRow := -1
		for i, r := range drows {
			if strings.Contains(r, "->") {
				dpcRow = i
				break
			}
		}
		if dpcRow >= 0 {
			_, _, _, h := t.DisassemblyView.GetRect()
			row := dpcRow - h/2
			if row < 0 {
				row = 0
			}
			_, col := t.DisassemblyView.GetScrollOffset()
			t.DisassemblyView.ScrollTo(row, col)
		}
	}
}

func (t *TUI) RefreshAll() {
	t.UpdateSourceView()
	t.UpdateRegisterView()
	t.UpdateMemoryView()
	t.UpdateStackView()
	t.UpdateDisassemblyView()
	t.UpdateBreakpointsView()
	t.scrollPCIntoView() // Auto-scroll to keep PC visible
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

	// Target: show more instructions to fill the window
	// Show instructions before and after PC to provide context
	const targetBefore = 20
	const targetAfter = 80

	// Count how many valid instructions exist before PC
	actualBefore := 0
	for offset := uint32(4); offset <= targetBefore*4; offset += 4 {
		checkAddr := pc - offset
		if checkAddr > pc { // Handle underflow
			break
		}
		if _, exists := t.Debugger.SourceMap[checkAddr]; exists {
			actualBefore++
		}
	}

	// Calculate how many to show after based on what we found before
	// If we have fewer than target before, show more after to fill the window
	showAfter := targetAfter + (targetBefore - actualBefore)

	// Determine start address
	beforeBytes, err := vm.SafeIntToUint32(actualBefore * 4)
	if err != nil {
		beforeBytes = 0 // Should never happen with small counts
	}
	startAddr := pc - beforeBytes
	if startAddr > pc { // Handle underflow
		startAddr = 0
	}

	// Build the display
	var lines []string
	afterBytes, err := vm.SafeIntToUint32(showAfter * 4)
	if err != nil {
		afterBytes = 40 // Fallback to default
	}
	endAddr := pc + afterBytes
	for addr := startAddr; addr <= endAddr; addr += 4 {
		if sourceLine, exists := t.Debugger.SourceMap[addr]; exists {
			// Check if there's a label at this address and prepend it
			if label := t.findSymbolForAddress(addr); label != "" {
				// Show label on its own line with distinctive marker in green
				labelLine := fmt.Sprintf("[green]>> %s:[white]", label)
				lines = append(lines, labelLine)
			}

			// Check if this is a data directive (prefixed with [DATA])
			isData := false
			displayLine := sourceLine
			if strings.HasPrefix(sourceLine, "[DATA]") {
				isData = true
				displayLine = strings.TrimPrefix(sourceLine, "[DATA]")
			}

			// Highlight current line
			marker := "  "
			color := "white"
			if addr == pc {
				marker = "->"
				color = "yellow"
			} else if isData {
				// Data directives display in green when not at PC
				color = "green"
			}

			// Check for breakpoint
			if t.Debugger.Breakpoints.GetBreakpoint(addr) != nil {
				marker = "* "
			}

			// Escape square brackets in displayLine so tview doesn't interpret them as color tags
			escapedLine := tview.Escape(displayLine)
			line := fmt.Sprintf("[%s]%s 0x%08X: %s[white]", color, marker, addr, escapedLine)
			lines = append(lines, line)
		}
	}

	t.SourceView.SetText(strings.Join(lines, "\n"))
}

// UpdateRegisterView updates the register view
func (t *TUI) UpdateRegisterView() {
	t.RegisterView.Clear()
	t.RegisterView.ScrollToBeginning()

	cpu := t.Debugger.VM.CPU
	var lines []string

	// General purpose registers (3 columns, R0-R14, PC shown separately below)
	for i := 0; i < 5; i++ {
		var cols []string
		for j := 0; j < 3; j++ {
			reg := i*3 + j
			if reg >= 15 {
				break // Only show R0-R14 here, PC shown on status line
			}
			name := fmt.Sprintf("R%-2d", reg)
			var value uint32
			if reg == 13 {
				name = "SP "
				value = cpu.R[reg]
			} else if reg == 14 {
				name = "LR "
				value = cpu.R[reg]
			} else {
				value = cpu.R[reg]
			}

			// Check if this register changed in the last step
			if t.ChangedRegs[reg] {
				cols = append(cols, fmt.Sprintf("[green]%-3s: 0x%08X[white]", name, value))
			} else {
				cols = append(cols, fmt.Sprintf("%-3s: 0x%08X", name, value))
			}
		}
		lines = append(lines, strings.Join(cols, " "))
	}

	// CPSR flags - uppercase yellow when set, lowercase white when clear
	flags := ""
	if cpu.CPSR.N {
		flags += "[yellow]N[white]"
	} else {
		flags += "n"
	}
	if cpu.CPSR.Z {
		flags += "[yellow]Z[white]"
	} else {
		flags += "z"
	}
	if cpu.CPSR.C {
		flags += "[yellow]C[white]"
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

	// Blank line separator
	lines = append(lines, "")

	// Put PC, CPSR, flags, and cycles all on one compact line
	pcColor := "white"
	if t.ChangedRegs[15] {
		pcColor = "green"
	}
	cpsrColor := "white"
	if t.ChangedCPSR {
		cpsrColor = "green"
	}
	statusLine := fmt.Sprintf("[%s]PC:0x%08X[white] [%s]CPSR:0x%08X[white] Flags:%s Cyc:%d",
		pcColor, cpu.PC, cpsrColor, cpsrValue, flags, cpu.Cycles)
	lines = append(lines, statusLine)

	t.RegisterView.SetText(strings.Join(lines, "\n"))
}

// UpdateMemoryView updates the memory view
func (t *TUI) UpdateMemoryView() {
	// Use current memory address or PC if not set
	addr := t.MemoryAddress
	if addr == 0 {
		addr = t.Debugger.VM.CPU.PC
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("[yellow]Address: %08X (Lines end with .)[white]", addr))

	// Show 16 rows of 16 bytes each
	for row := 0; row < 16; row++ {
		rowOffset, err := vm.SafeIntToUint32(row * 16)
		if err != nil {
			break // Should never happen
		}
		rowAddr := addr + rowOffset

		// Address (without 0x prefix to save space)
		line := fmt.Sprintf("%08X: ", rowAddr)

		// Hex bytes - build manually to handle color tags properly
		var hexPart string

		for col := 0; col < 16; col++ {
			colOffset, err := vm.SafeIntToUint32(col)
			if err != nil {
				break // Should never happen
			}
			byteAddr := rowAddr + colOffset
			b, err := t.Debugger.VM.Memory.ReadByteAt(byteAddr)
			if err != nil {
				if col > 0 {
					hexPart += " "
				}
				hexPart += "??"
			} else {
				// Add space before byte (except first)
				if col > 0 {
					hexPart += " "
				}
				// Highlight recently written bytes in green
				if t.RecentWrites[byteAddr] {
					hexPart += fmt.Sprintf("[green]%02X[white]", b)
				} else {
					hexPart += fmt.Sprintf("%02X", b)
				}
			}
		}

		// Add end-of-line marker
		line += hexPart + "."

		lines = append(lines, line)
	}

	t.MemoryView.SetText(strings.Join(lines, "\n"))
}

// UpdateStackView updates the stack view
func (t *TUI) UpdateStackView() {
	t.StackView.Clear()

	sp := t.Debugger.VM.CPU.R[13] // Stack pointer

	var lines []string
	lines = append(lines, fmt.Sprintf("[yellow]Stack Pointer: 0x%08X (Lines end with .)[white]", sp))

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
			lines = append(lines, fmt.Sprintf("0x%08X: ????????.", addr))
			continue
		}

		// Mark current SP
		marker := "  "
		if addr == sp {
			marker = "->"
		}

		// Check if this stack location was recently written
		isRecentWrite := t.RecentWrites[addr] || t.RecentWrites[addr+1] ||
			t.RecentWrites[addr+2] || t.RecentWrites[addr+3]

		var line string
		if isRecentWrite {
			line = fmt.Sprintf("%s 0x%08X: [green]0x%08X[white]", marker, addr, word)
		} else {
			line = fmt.Sprintf("%s 0x%08X: 0x%08X", marker, addr, word)
		}

		// Try to resolve as symbol
		if sym := t.findSymbolForAddress(word); sym != "" {
			line += fmt.Sprintf(" <%s>", sym)
		}

		// Add end-of-line marker
		line += "."

		lines = append(lines, line)
	}

	t.StackView.SetText(strings.Join(lines, "\n"))
}

// UpdateDisassemblyView updates the disassembly view
func (t *TUI) UpdateDisassemblyView() {
	t.DisassemblyView.Clear()

	pc := t.Debugger.VM.CPU.PC

	var lines []string

	// Strategy: Collect instructions before and after PC separately
	// This ensures we always show PC and instructions after it, even if there are gaps before
	const targetBefore = 5
	const targetAfter = 10

	// Collect instructions BEFORE PC (up to targetBefore)
	var beforeLines []string
	for offset := targetBefore * 4; offset > 0; offset -= 4 {
		offsetU32, err := vm.SafeIntToUint32(offset)
		if err != nil {
			continue
		}
		addr := pc - offsetU32
		if addr > pc { // Handle underflow
			continue
		}

		// Read instruction
		instr, err := t.Debugger.VM.Memory.ReadWord(addr)
		if err != nil {
			continue // Skip invalid addresses
		}

		// Check for breakpoint
		marker := "  "
		if t.Debugger.Breakpoints.GetBreakpoint(addr) != nil {
			marker = "* "
		}

		// Simple disassembly
		line := fmt.Sprintf("[white]%s 0x%08X: 0x%08X[white]", marker, addr, instr)

		// Try to add symbol
		if sym := t.findSymbolForAddress(addr); sym != "" {
			line = fmt.Sprintf("[white]%s 0x%08X: 0x%08X  <%s>[white]", marker, addr, instr, sym)
		}

		beforeLines = append(beforeLines, line)
	}

	// Append before lines in ascending address order (they're already collected in that order)
	lines = append(lines, beforeLines...)

	// Collect instructions AT and AFTER PC (current + targetAfter)
	for i := 0; i <= targetAfter; i++ {
		offset, err := vm.SafeIntToUint32(i * 4)
		if err != nil {
			break
		}
		addr := pc + offset

		// Read instruction
		instr, err := t.Debugger.VM.Memory.ReadWord(addr)
		if err != nil {
			continue // Skip invalid addresses
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

		// Simple disassembly
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

	// Dynamically adjust the height based on content
	// Calculate needed height: borders (2) + content lines, with min of 4 and max of 12
	numLines := len(lines)
	height := numLines + 2 // Add 2 for border
	if height < 4 {
		height = 4 // Minimum height when no breakpoints
	}
	if height > 12 {
		height = 12 // Maximum height to prevent taking too much space
	}

	// Update the layout with new height
	t.RightPanel.ResizeItem(t.BreakpointsView, height, 0)
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

// CaptureRegisterState captures the current register state before stepping
func (t *TUI) CaptureRegisterState() {
	cpu := t.Debugger.VM.CPU

	// Save current register values
	for i := 0; i < 15; i++ {
		t.PrevRegisters[i] = cpu.R[i]
	}
	t.PrevRegisters[15] = cpu.PC

	// Save CPSR
	t.PrevCPSR = cpu.CPSR
}

// DetectRegisterChanges compares current registers with previous state
func (t *TUI) DetectRegisterChanges() {
	cpu := t.Debugger.VM.CPU

	// Clear previous changes
	t.ChangedRegs = make(map[int]bool)
	t.ChangedCPSR = false

	// Check each register
	for i := 0; i < 15; i++ {
		if cpu.R[i] != t.PrevRegisters[i] {
			t.ChangedRegs[i] = true
		}
	}

	// Check PC (R15)
	if cpu.PC != t.PrevRegisters[15] {
		t.ChangedRegs[15] = true
	}

	// Check CPSR flags
	if cpu.CPSR.N != t.PrevCPSR.N || cpu.CPSR.Z != t.PrevCPSR.Z ||
		cpu.CPSR.C != t.PrevCPSR.C || cpu.CPSR.V != t.PrevCPSR.V {
		t.ChangedCPSR = true
	}
}

// CaptureMemoryTraceState captures the current memory trace entry count before stepping
func (t *TUI) CaptureMemoryTraceState() {
	if t.Debugger.VM.MemoryTrace != nil && t.Debugger.VM.MemoryTrace.Enabled {
		entries := t.Debugger.VM.MemoryTrace.GetEntries()
		t.LastTraceEntryCount = len(entries)
	}
}

// DetectMemoryWrites tracks memory writes from the last step using MemoryTrace
func (t *TUI) DetectMemoryWrites() {
	// Clear previous writes
	t.RecentWrites = make(map[uint32]bool)

	// If MemoryTrace is enabled, check for new writes since last step
	if t.Debugger.VM.MemoryTrace != nil && t.Debugger.VM.MemoryTrace.Enabled {
		entries := t.Debugger.VM.MemoryTrace.GetEntries()
		var firstWriteAddr uint32
		foundWrite := false

		// Only look at new entries since last step
		for i := t.LastTraceEntryCount; i < len(entries); i++ {
			if entries[i].Type == "WRITE" {
				// Mark this address and surrounding bytes as recently written
				// This handles word writes (4 bytes) and byte writes
				addr := entries[i].Address
				t.RecentWrites[addr] = true
				t.RecentWrites[addr+1] = true
				t.RecentWrites[addr+2] = true
				t.RecentWrites[addr+3] = true

				// Track first write address to auto-focus memory view
				if !foundWrite {
					firstWriteAddr = addr
					foundWrite = true
				}
			}
		}

		// Auto-focus Memory window on first written address
		// Only update if it's not in the stack region (let stack view handle that)
		if foundWrite && firstWriteAddr < vm.StackSegmentStart {
			// Align to 16-byte boundary for better display
			t.MemoryAddress = firstWriteAddr & 0xFFFFFFF0
		}
	}
}
