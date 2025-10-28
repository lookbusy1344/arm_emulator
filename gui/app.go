package main

import (
	"context"
	"fmt"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
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
	return a.service.Step()
}

// Continue runs until breakpoint or halt
func (a *App) Continue() error {
	return a.service.RunUntilHalt()
}

// Pause pauses execution
func (a *App) Pause() {
	a.service.Pause()
}

// Reset resets VM to initial state
func (a *App) Reset() error {
	return a.service.Reset()
}

// AddBreakpoint adds a breakpoint at address
func (a *App) AddBreakpoint(address uint32) error {
	return a.service.AddBreakpoint(address)
}

// RemoveBreakpoint removes a breakpoint
func (a *App) RemoveBreakpoint(address uint32) error {
	return a.service.RemoveBreakpoint(address)
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
