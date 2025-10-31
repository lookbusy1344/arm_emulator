package debugger

import (
	"fmt"
	"strings"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// Debugger represents the debugger state and functionality
type Debugger struct {
	VM *vm.VM

	// Breakpoint management
	Breakpoints *BreakpointManager

	// Watchpoint management
	Watchpoints *WatchpointManager

	// Command history
	History *CommandHistory

	// Expression evaluator
	Evaluator *ExpressionEvaluator

	// Execution control
	Running           bool
	StepMode          StepMode
	StepOverCallDepth int    // Track call depth for step over
	StepOverPC        uint32 // PC to return to after step over

	// Symbol table (for label/symbol resolution)
	Symbols map[string]uint32

	// Source code mapping (address -> source line)
	SourceMap map[uint32]string

	// Last command (for repeat on empty input)
	LastCommand string

	// Output buffer
	Output strings.Builder
}

// StepMode represents different stepping modes
type StepMode int

const (
	StepNone   StepMode = iota // Not stepping
	StepSingle                 // Step one instruction
	StepOver                   // Step over function calls
	StepOut                    // Step out of current function
)

// NewDebugger creates a new debugger instance
func NewDebugger(machine *vm.VM) *Debugger {
	return &Debugger{
		VM:          machine,
		Breakpoints: NewBreakpointManager(),
		Watchpoints: NewWatchpointManager(),
		History:     NewCommandHistory(),
		Evaluator:   NewExpressionEvaluator(),
		Running:     false,
		StepMode:    StepNone,
		Symbols:     make(map[string]uint32),
		SourceMap:   make(map[uint32]string),
	}
}

// LoadSymbols loads the symbol table for label resolution
func (d *Debugger) LoadSymbols(symbols map[string]uint32) {
	d.Symbols = symbols
}

// LoadSourceMap loads the source code mapping
func (d *Debugger) LoadSourceMap(sourceMap map[uint32]string) {
	d.SourceMap = sourceMap
}

// ResolveAddress resolves a label to an address, or parses a numeric address
func (d *Debugger) ResolveAddress(addrStr string) (uint32, error) {
	// Try to resolve as symbol first
	if addr, exists := d.Symbols[addrStr]; exists {
		return addr, nil
	}

	// Try to parse as numeric address
	var addr uint32
	if strings.HasPrefix(addrStr, "0x") || strings.HasPrefix(addrStr, "0X") {
		_, err := fmt.Sscanf(addrStr, "0x%x", &addr)
		if err != nil {
			return 0, fmt.Errorf("invalid address: %s", addrStr)
		}
	} else {
		_, err := fmt.Sscanf(addrStr, "%d", &addr)
		if err != nil {
			return 0, fmt.Errorf("invalid address: %s", addrStr)
		}
	}

	return addr, nil
}

// ExecuteCommand processes and executes a debugger command
func (d *Debugger) ExecuteCommand(cmdLine string) error {
	// Trim whitespace
	cmdLine = strings.TrimSpace(cmdLine)

	// Empty command repeats last command (for step, next, etc.)
	if cmdLine == "" {
		cmdLine = d.LastCommand
	}

	// Don't store empty commands
	if cmdLine != "" {
		d.History.Add(cmdLine)
		d.LastCommand = cmdLine
	}

	// Parse command
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	// Execute command
	return d.handleCommand(cmd, args)
}

// handleCommand dispatches commands to appropriate handlers
func (d *Debugger) handleCommand(cmd string, args []string) error {
	switch cmd {
	// Execution control
	case "run", "r":
		return d.cmdRun(args)
	case "continue", "c":
		return d.cmdContinue(args)
	case "step", "s", "si":
		return d.cmdStep(args)
	case "next", "n":
		return d.cmdNext(args)
	case "finish", "fin":
		return d.cmdFinish(args)

	// Breakpoints
	case "break", "b":
		return d.cmdBreak(args)
	case "tbreak", "tb":
		return d.cmdTBreak(args)
	case "delete", "d":
		return d.cmdDelete(args)
	case "enable":
		return d.cmdEnable(args)
	case "disable":
		return d.cmdDisable(args)

	// Watchpoints
	case "watch", "w":
		return d.cmdWatch(args)
	case "rwatch":
		return d.cmdRWatch(args)
	case "awatch":
		return d.cmdAWatch(args)

	// Inspection
	case "print", "p":
		return d.cmdPrint(args)
	case "x":
		return d.cmdExamine(args)
	case "info", "i":
		return d.cmdInfo(args)
	case "backtrace", "bt", "where":
		return d.cmdBacktrace(args)
	case "list", "l":
		return d.cmdList(args)

	// State modification
	case "set":
		return d.cmdSet(args)

	// Program control
	case "load":
		return d.cmdLoad(args)
	case "reset":
		return d.cmdReset(args)

	// Help
	case "help", "h", "?":
		return d.cmdHelp(args)

	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmd)
	}
}

// ShouldBreak checks if execution should pause at the current PC
func (d *Debugger) ShouldBreak() (bool, string) {
	pc := d.VM.CPU.PC

	// Check step mode
	switch d.StepMode {
	case StepSingle:
		d.StepMode = StepNone
		return true, "single step"

	case StepOver:
		// Continue until we return to the same call depth
		if pc == d.StepOverPC {
			d.StepMode = StepNone
			return true, "step over complete"
		}

	case StepOut:
		// This would require call stack tracking
		// For now, simplified implementation
	}

	// Check breakpoints
	if bp := d.Breakpoints.GetBreakpoint(pc); bp != nil {
		if !bp.Enabled {
			return false, ""
		}

		// Evaluate condition if present
		if bp.Condition != "" {
			result, err := d.Evaluator.Evaluate(bp.Condition, d.VM, d.Symbols)
			if err != nil {
				return true, fmt.Sprintf("breakpoint %d (condition error: %v)", bp.ID, err)
			}
			if !result {
				return false, ""
			}
		}

		// Increment hit count
		bp.HitCount++

		// Check if temporary breakpoint
		if bp.Temporary {
			_ = d.Breakpoints.DeleteBreakpoint(bp.ID) // Ignore error on cleanup
		}

		return true, fmt.Sprintf("breakpoint %d", bp.ID)
	}

	// Check watchpoints
	if wp, changed := d.Watchpoints.CheckWatchpoints(d.VM); wp != nil && changed {
		return true, fmt.Sprintf("watchpoint %d: %s", wp.ID, wp.Expression)
	}

	return false, ""
}

// GetOutput returns and clears the output buffer
func (d *Debugger) GetOutput() string {
	output := d.Output.String()
	d.Output.Reset()
	return output
}

// Printf writes formatted output to the output buffer
func (d *Debugger) Printf(format string, args ...interface{}) {
	d.Output.WriteString(fmt.Sprintf(format, args...))
}

// Println writes a line to the output buffer
func (d *Debugger) Println(args ...interface{}) {
	d.Output.WriteString(fmt.Sprintln(args...))
}

// SetStepOver configures the debugger to step over function calls
// This should be called while holding the appropriate locks in the calling code
func (d *Debugger) SetStepOver() {
	// Read instruction at current PC
	instr, err := d.VM.Memory.ReadWord(d.VM.CPU.PC)
	if err != nil {
		// If we can't read the instruction, fall back to single step
		d.StepMode = StepSingle
		d.Running = true
		return
	}

	// Check if this is a BL (Branch with Link) instruction
	// BL: bits[31:28] = condition, bits[27:24] = 1011
	isBL := (instr & 0x0F000000) == 0x0B000000

	if isBL {
		// This is a function call - set up step over
		d.StepOverPC = d.VM.CPU.PC + 4
		d.StepMode = StepOver
		d.Running = true
	} else {
		// Not a function call - just single step
		d.StepMode = StepSingle
		d.Running = true
	}
}

// SetStepOut configures the debugger to step out of the current function
// This should be called while holding the appropriate locks in the calling code
func (d *Debugger) SetStepOut() {
	d.StepMode = StepOut
	d.Running = true
}
