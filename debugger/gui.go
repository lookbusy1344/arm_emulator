package debugger

import (
	"fmt"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// GUI represents the graphical user interface for the debugger
type GUI struct {
	// Core components
	Debugger *Debugger
	App      fyne.App
	Window   fyne.Window

	// View panels
	SourceView      *widget.TextGrid
	RegisterView    *widget.TextGrid
	MemoryView      *widget.TextGrid
	StackView       *widget.TextGrid
	BreakpointsList *widget.List
	ConsoleOutput   *widget.TextGrid
	StatusLabel     *widget.Label

	// Controls
	Toolbar *widget.Toolbar

	// State
	CurrentAddress uint32
	MemoryAddress  uint32
	StackAddress   uint32
	Running        bool

	// Source code cache
	SourceLines []string
	SourceFile  string

	// Breakpoints data
	breakpoints []string

	// Console output buffer
	consoleBuffer strings.Builder
	consoleMutex  sync.Mutex
}

// guiWriter redirects VM output to the GUI console
type guiWriter struct {
	gui *GUI
}

// Write implements io.Writer interface
func (w *guiWriter) Write(p []byte) (n int, err error) {
	w.gui.consoleMutex.Lock()
	defer w.gui.consoleMutex.Unlock()

	w.gui.consoleBuffer.Write(p)
	w.gui.updateConsole()
	return len(p), nil
}

// RunGUI runs the GUI (Graphical User Interface) debugger
func RunGUI(dbg *Debugger) error {
	gui := newGUI(dbg)
	gui.Window.ShowAndRun()
	return nil
}

// newGUI creates a new graphical user interface
func newGUI(debugger *Debugger) *GUI {
	myApp := app.New()
	myWindow := myApp.NewWindow("ARM2 Emulator Debugger")

	gui := &GUI{
		Debugger:       debugger,
		App:            myApp,
		Window:         myWindow,
		CurrentAddress: 0,
		MemoryAddress:  0,
		StackAddress:   0,
		Running:        false,
		breakpoints:    []string{},
	}

	gui.initializeViews()
	gui.buildLayout()
	gui.setupToolbar()

	// Redirect VM output to GUI console
	debugger.VM.OutputWriter = &guiWriter{gui: gui}

	// Set window size
	myWindow.Resize(fyne.NewSize(1400, 900))

	return gui
}

// initializeViews creates all the view panels
func (g *GUI) initializeViews() {
	// Source view
	g.SourceView = widget.NewTextGrid()
	g.SourceView.SetText("No source file loaded")

	// Register view
	g.RegisterView = widget.NewTextGrid()
	g.updateRegisters()

	// Memory view
	g.MemoryView = widget.NewTextGrid()
	g.updateMemory()

	// Stack view
	g.StackView = widget.NewTextGrid()
	g.updateStack()

	// Breakpoints list
	g.breakpoints = []string{}
	g.BreakpointsList = widget.NewList(
		func() int {
			return len(g.breakpoints)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(g.breakpoints[id])
		},
	)

	// Console output
	g.ConsoleOutput = widget.NewTextGrid()
	g.ConsoleOutput.SetText("")

	// Status label
	g.StatusLabel = widget.NewLabel("Ready")
}

// buildLayout creates the main layout
func (g *GUI) buildLayout() {
	// Create bordered panels for better visual separation
	sourcePanel := container.NewBorder(
		widget.NewLabel("📄 Source Code"),
		nil, nil, nil,
		container.NewScroll(g.SourceView),
	)

	registerPanel := container.NewBorder(
		widget.NewLabel("📊 Registers"),
		nil, nil, nil,
		container.NewScroll(g.RegisterView),
	)

	memoryPanel := container.NewBorder(
		widget.NewLabel("💾 Memory"),
		nil, nil, nil,
		container.NewScroll(g.MemoryView),
	)

	stackPanel := container.NewBorder(
		widget.NewLabel("📚 Stack"),
		nil, nil, nil,
		container.NewScroll(g.StackView),
	)

	breakpointsPanel := container.NewBorder(
		widget.NewLabel("🔴 Breakpoints"),
		nil, nil, nil,
		container.NewScroll(g.BreakpointsList),
	)

	consolePanel := container.NewBorder(
		widget.NewLabel("💻 Console Output"),
		nil, nil, nil,
		container.NewScroll(g.ConsoleOutput),
	)

	// Left side: source code (larger)
	leftPanel := container.NewMax(sourcePanel)

	// Right side: registers and breakpoints
	rightTop := container.NewVSplit(registerPanel, breakpointsPanel)
	rightTop.SetOffset(0.6) // 60% registers, 40% breakpoints

	// Bottom right: memory, stack, console
	bottomTabs := container.NewAppTabs(
		container.NewTabItem("Memory", memoryPanel),
		container.NewTabItem("Stack", stackPanel),
		container.NewTabItem("Console", consolePanel),
	)

	rightPanel := container.NewVSplit(rightTop, bottomTabs)
	rightPanel.SetOffset(0.5)

	// Main split: left (source) and right (info panels)
	mainSplit := container.NewHSplit(leftPanel, rightPanel)
	mainSplit.SetOffset(0.55) // 55% source, 45% info

	// Add status bar at bottom
	statusBar := container.NewBorder(nil, nil, nil, nil, g.StatusLabel)

	// Complete layout with toolbar at top
	content := container.NewBorder(
		g.Toolbar,    // top
		statusBar,    // bottom
		nil,          // left
		nil,          // right
		mainSplit,    // center
	)

	g.Window.SetContent(content)
}

// setupToolbar creates the debugger control toolbar
func (g *GUI) setupToolbar() {
	g.Toolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.MediaPlayIcon(), func() {
			g.runProgram()
		}),
		widget.NewToolbarAction(theme.MediaSkipNextIcon(), func() {
			g.stepProgram()
		}),
		widget.NewToolbarAction(theme.MediaFastForwardIcon(), func() {
			g.continueProgram()
		}),
		widget.NewToolbarAction(theme.MediaStopIcon(), func() {
			g.stopProgram()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			g.addBreakpoint()
		}),
		widget.NewToolbarAction(theme.ContentClearIcon(), func() {
			g.clearBreakpoints()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			g.refreshViews()
		}),
	)
}

// updateViews refreshes all view panels
func (g *GUI) updateViews() {
	g.updateSource()
	g.updateRegisters()
	g.updateMemory()
	g.updateStack()
	g.updateBreakpoints()
	g.updateConsole()
}

// updateSource updates the source code view
func (g *GUI) updateSource() {
	if len(g.SourceLines) == 0 {
		// Try to load source from debugger
		if g.Debugger.SourceMap != nil && len(g.Debugger.SourceMap) > 0 {
			// Get all source lines from the SourceMap
			maxLine := 0
			for _, line := range g.Debugger.SourceMap {
				if len(line) > maxLine {
					maxLine = len(line)
				}
			}
			// Just show the source map as text for now
		}
	}

	if len(g.SourceLines) > 0 {
		var sb strings.Builder
		currentPC := g.Debugger.VM.CPU.PC

		// Find current line from PC
		currentSourceLine := ""
		if g.Debugger.SourceMap != nil {
			if line, ok := g.Debugger.SourceMap[currentPC]; ok {
				currentSourceLine = line
			}
		}

		for i, line := range g.SourceLines {
			prefix := "  "
			if line == currentSourceLine {
				prefix = "→ "
			}
			sb.WriteString(fmt.Sprintf("%s%4d: %s\n", prefix, i+1, line))
		}
		g.SourceView.SetText(sb.String())
	} else {
		// Show simple disassembly view
		var sb strings.Builder
		currentPC := g.Debugger.VM.CPU.PC
		
		sb.WriteString(fmt.Sprintf("Current PC: 0x%08X\n\n", currentPC))
		if source, ok := g.Debugger.SourceMap[currentPC]; ok {
			sb.WriteString(fmt.Sprintf("→ %s\n", source))
		} else {
			sb.WriteString("No source mapping available\n")
		}
		g.SourceView.SetText(sb.String())
	}
}

// updateRegisters updates the register view
func (g *GUI) updateRegisters() {
	var sb strings.Builder

	cpu := g.Debugger.VM.CPU

	sb.WriteString("General Purpose Registers:\n")
	sb.WriteString("──────────────────────────\n")
	for i := 0; i < 13; i++ {
		sb.WriteString(fmt.Sprintf("R%-2d: 0x%08X  (%d)\n", i, cpu.R[i], cpu.R[i]))
	}

	sb.WriteString("\nSpecial Registers:\n")
	sb.WriteString("──────────────────────────\n")
	sb.WriteString(fmt.Sprintf("SP:  0x%08X  (%d)\n", cpu.R[13], cpu.R[13]))
	sb.WriteString(fmt.Sprintf("LR:  0x%08X  (%d)\n", cpu.R[14], cpu.R[14]))
	sb.WriteString(fmt.Sprintf("PC:  0x%08X  (%d)\n", cpu.PC, cpu.PC))

	sb.WriteString("\nStatus Flags (CPSR):\n")
	sb.WriteString("──────────────────────────\n")
	flags := ""
	if cpu.CPSR.N {
		flags += "N"
	} else {
		flags += "-"
	}
	if cpu.CPSR.Z {
		flags += "Z"
	} else {
		flags += "-"
	}
	if cpu.CPSR.C {
		flags += "C"
	} else {
		flags += "-"
	}
	if cpu.CPSR.V {
		flags += "V"
	} else {
		flags += "-"
	}
	sb.WriteString(fmt.Sprintf("Flags: %s\n", flags))

	g.RegisterView.SetText(sb.String())
}

// updateMemory updates the memory view
func (g *GUI) updateMemory() {
	var sb strings.Builder

	// Show memory around PC or a specific address
	addr := g.MemoryAddress
	if addr == 0 {
		addr = g.Debugger.VM.CPU.PC
	}

	// Round down to 16-byte boundary
	addr = addr & 0xFFFFFFF0

	sb.WriteString(fmt.Sprintf("Memory at 0x%08X:\n", addr))
	sb.WriteString("──────────────────────────────────────────────────\n")

	// Show 16 lines of 16 bytes each
	for i := uint32(0); i < 16; i++ {
		lineAddr := addr + (i * 16)
		sb.WriteString(fmt.Sprintf("%08X: ", lineAddr))

		// Hex view
		for j := uint32(0); j < 16; j++ {
			byteAddr := lineAddr + j
			b, err := g.Debugger.VM.Memory.ReadByteAt(byteAddr)
			if err == nil {
				sb.WriteString(fmt.Sprintf("%02X ", b))
			} else {
				sb.WriteString("?? ")
			}
		}

		// ASCII view
		sb.WriteString(" ")
		for j := uint32(0); j < 16; j++ {
			byteAddr := lineAddr + j
			b, err := g.Debugger.VM.Memory.ReadByteAt(byteAddr)
			if err == nil {
				if b >= 32 && b < 127 {
					sb.WriteString(string(b))
				} else {
					sb.WriteString(".")
				}
			} else {
				sb.WriteString("?")
			}
		}
		sb.WriteString("\n")
	}

	g.MemoryView.SetText(sb.String())
}

// updateStack updates the stack view
func (g *GUI) updateStack() {
	var sb strings.Builder

	sp := g.Debugger.VM.CPU.R[13] // SP

	sb.WriteString(fmt.Sprintf("Stack at SP=0x%08X:\n", sp))
	sb.WriteString("──────────────────────────────\n")

	// Show 16 words above and below SP
	for i := int32(-8); i < 24; i++ {
		addr := uint32(int32(sp) + (i * 4))
		prefix := "  "
		if i == 0 {
			prefix = "→ "
		}

		word, err := g.Debugger.VM.Memory.ReadWord(addr)
		if err == nil {
			sb.WriteString(fmt.Sprintf("%s%08X: %08X  (%d)\n", prefix, addr, word, word))
		}
	}

	g.StackView.SetText(sb.String())
}

// updateBreakpoints updates the breakpoints list
func (g *GUI) updateBreakpoints() {
	breakpoints := g.Debugger.Breakpoints.GetAllBreakpoints()
	g.breakpoints = make([]string, 0, len(breakpoints))

	for _, bp := range breakpoints {
		// Try to resolve symbol name
		symbol := ""
		if g.Debugger.Symbols != nil {
			for name, addr := range g.Debugger.Symbols {
				if addr == bp.Address {
					symbol = fmt.Sprintf(" [%s]", name)
					break
				}
			}
		}
		
		status := "enabled"
		if !bp.Enabled {
			status = "disabled"
		}
		
		g.breakpoints = append(g.breakpoints, fmt.Sprintf("0x%08X%s (%s)", bp.Address, symbol, status))
	}

	g.BreakpointsList.Refresh()
}

// updateConsole updates the console output view
func (g *GUI) updateConsole() {
	g.consoleMutex.Lock()
	defer g.consoleMutex.Unlock()

	g.ConsoleOutput.SetText(g.consoleBuffer.String())
}

// runProgram starts/restarts program execution
func (g *GUI) runProgram() {
	g.StatusLabel.SetText("Running...")
	g.Debugger.VM.State = vm.StateRunning

	// Execute program in goroutine to keep UI responsive
	go func() {
		for g.Debugger.VM.State == vm.StateRunning {
			if err := g.Debugger.VM.Step(); err != nil {
				g.StatusLabel.SetText(fmt.Sprintf("Error: %v", err))
				break
			}

			// Check for breakpoints
			if shouldBreak, reason := g.Debugger.ShouldBreak(); shouldBreak {
				g.StatusLabel.SetText(fmt.Sprintf("Stopped: %s at PC=0x%08X", reason, g.Debugger.VM.CPU.PC))
				g.Debugger.VM.State = vm.StateBreakpoint
				g.updateViews()
				break
			}

			// Check if halted
			if g.Debugger.VM.State == vm.StateHalted {
				g.StatusLabel.SetText(fmt.Sprintf("Program exited with code %d", g.Debugger.VM.ExitCode))
				g.updateViews()
				break
			}
		}
	}()
}

// stepProgram executes one instruction
func (g *GUI) stepProgram() {
	if g.Debugger.VM.State == vm.StateHalted {
		g.StatusLabel.SetText("Program has halted")
		return
	}

	g.Debugger.VM.State = vm.StateRunning
	if err := g.Debugger.VM.Step(); err != nil {
		g.StatusLabel.SetText(fmt.Sprintf("Error: %v", err))
		return
	}

	if g.Debugger.VM.State == vm.StateHalted {
		g.StatusLabel.SetText(fmt.Sprintf("Program exited with code %d", g.Debugger.VM.ExitCode))
	} else {
		g.StatusLabel.SetText(fmt.Sprintf("Stepped to PC=0x%08X", g.Debugger.VM.CPU.PC))
	}

	g.updateViews()
}

// continueProgram continues execution until breakpoint
func (g *GUI) continueProgram() {
	g.runProgram()
}

// stopProgram stops execution
func (g *GUI) stopProgram() {
	g.Debugger.VM.State = vm.StateBreakpoint
	g.StatusLabel.SetText("Stopped")
	g.updateViews()
}

// addBreakpoint adds a breakpoint at current PC
func (g *GUI) addBreakpoint() {
	pc := g.Debugger.VM.CPU.PC
	g.Debugger.Breakpoints.AddBreakpoint(pc, false, "")
	g.updateBreakpoints()
	g.StatusLabel.SetText(fmt.Sprintf("Breakpoint added at 0x%08X", pc))
}

// clearBreakpoints removes all breakpoints
func (g *GUI) clearBreakpoints() {
	g.Debugger.Breakpoints.Clear()
	g.updateBreakpoints()
	g.StatusLabel.SetText("All breakpoints cleared")
}

// refreshViews manually refreshes all views
func (g *GUI) refreshViews() {
	g.updateViews()
	g.StatusLabel.SetText("Views refreshed")
}
