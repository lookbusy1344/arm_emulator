package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	service *service.DebuggerService
	machine *vm.VM
}

// NewApp creates a new App application struct
func NewApp() *App {
	machine := vm.NewVM()
	machine.InitializeStack(0x30001000) // Default stack

	return &App{
		machine: machine,
		service: service.NewDebuggerService(machine),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.service.SetContext(ctx)
}

// LoadProgramFromSource parses and loads assembly source code
func (a *App) LoadProgramFromSource(source string, filename string, entryPoint uint32) error {
	p := parser.NewParser(source, filename)
	program, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	return a.service.LoadProgram(program, entryPoint)
}

// GetRegisters returns current register state
func (a *App) GetRegisters() service.RegisterState {
	return a.service.GetRegisterState()
}

// Step executes a single instruction
func (a *App) Step() error {
	err := a.service.Step()
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	} else {
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
	}
	return err
}

// Continue runs until breakpoint or halt
func (a *App) Continue() error {
	err := a.service.RunUntilHalt()
	runtime.EventsEmit(a.ctx, "vm:state-changed")

	if err != nil {
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
	}

	// Check if stopped at breakpoint
	if a.service.GetExecutionState() == service.StateBreakpoint {
		runtime.EventsEmit(a.ctx, "vm:breakpoint-hit")
	}

	return err
}

// Pause pauses execution
func (a *App) Pause() {
	a.service.Pause()
	runtime.EventsEmit(a.ctx, "vm:state-changed")
}

// Reset resets VM to initial state
func (a *App) Reset() error {
	err := a.service.Reset()
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	} else {
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
	}
	return err
}

// AddBreakpoint adds a breakpoint at address
func (a *App) AddBreakpoint(address uint32) error {
	err := a.service.AddBreakpoint(address)
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}
	return err
}

// RemoveBreakpoint removes a breakpoint
func (a *App) RemoveBreakpoint(address uint32) error {
	err := a.service.RemoveBreakpoint(address)
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}
	return err
}

// GetBreakpoints returns all breakpoints
func (a *App) GetBreakpoints() []service.BreakpointInfo {
	return a.service.GetBreakpoints()
}

// GetMemory returns memory contents
func (a *App) GetMemory(address uint32, size uint32) ([]byte, error) {
	return a.service.GetMemory(address, size)
}

// GetSourceLine returns source for address
func (a *App) GetSourceLine(address uint32) string {
	return a.service.GetSourceLine(address)
}

// GetSymbols returns all symbols
func (a *App) GetSymbols() map[string]uint32 {
	return a.service.GetSymbols()
}

// GetExecutionState returns current state
func (a *App) GetExecutionState() string {
	return string(a.service.GetExecutionState())
}

// IsRunning returns whether execution is active
func (a *App) IsRunning() bool {
	return a.service.IsRunning()
}

// ToggleBreakpoint toggles a breakpoint at the specified address
func (a *App) ToggleBreakpoint(address uint32) error {
	bps := a.service.GetBreakpoints()
	exists := false

	for _, bp := range bps {
		if bp.Address == address {
			exists = true
			break
		}
	}

	var err error
	if exists {
		err = a.service.RemoveBreakpoint(address)
	} else {
		err = a.service.AddBreakpoint(address)
	}

	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}

	return err
}

// GetSourceMap returns the complete source map
func (a *App) GetSourceMap() map[uint32]string {
	return a.service.GetSourceMap()
}

// GetDisassembly returns disassembled instructions
func (a *App) GetDisassembly(startAddr uint32, count int) []service.DisassemblyLine {
	return a.service.GetDisassembly(startAddr, count)
}

// GetStack returns stack contents
func (a *App) GetStack(offset int, count int) []service.StackEntry {
	return a.service.GetStack(offset, count)
}

// GetSymbolForAddress resolves address to symbol
func (a *App) GetSymbolForAddress(addr uint32) string {
	return a.service.GetSymbolForAddress(addr)
}

// GetOutput returns captured output
func (a *App) GetOutput() string {
	return a.service.GetOutput()
}

// StepOver steps over function calls
func (a *App) StepOver() error {
	err := a.service.StepOver()
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	} else {
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
	}
	return err
}

// StepOut steps out of current function
func (a *App) StepOut() error {
	err := a.service.StepOut()
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	} else {
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
	}
	return err
}

// AddWatchpoint adds a watchpoint
func (a *App) AddWatchpoint(address uint32, watchType string) error {
	err := a.service.AddWatchpoint(address, watchType)
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}
	return err
}

// RemoveWatchpoint removes a watchpoint
func (a *App) RemoveWatchpoint(id int) error {
	err := a.service.RemoveWatchpoint(id)
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}
	return err
}

// ExecuteCommand executes a debugger command
func (a *App) ExecuteCommand(command string) (string, error) {
	output, err := a.service.ExecuteCommand(command)

	// Check if command modified state
	if isStateModifyingCommand(command) {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}

	return output, err
}

// EvaluateExpression evaluates an expression
func (a *App) EvaluateExpression(expr string) (uint32, error) {
	return a.service.EvaluateExpression(expr)
}

// isStateModifyingCommand checks if command modifies VM state
func isStateModifyingCommand(command string) bool {
	stateCommands := []string{"step", "next", "finish", "continue", "set", "break", "delete"}
	for _, cmd := range stateCommands {
		if strings.HasPrefix(strings.ToLower(command), cmd) {
			return true
		}
	}
	return false
}
