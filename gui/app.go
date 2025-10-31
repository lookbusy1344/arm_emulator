package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var debugLog *log.Logger
var debugEnabled bool

func init() {
	// Check if debug logging is enabled via environment variable
	debugEnabled = os.Getenv("ARM_EMULATOR_DEBUG") != ""
	
	if debugEnabled {
		// Create debug log file with restrictive permissions (0600 = owner read/write only)
		f, err := os.OpenFile("/tmp/arm-emulator-gui-debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open debug log: %v\n", err)
			debugLog = log.New(os.Stderr, "GUI: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
		} else {
			debugLog = log.New(f, "GUI: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
		}
	} else {
		// Disable logging by default
		debugLog = log.New(io.Discard, "", 0)
	}
}

// App struct
type App struct {
	ctx     context.Context
	service *service.DebuggerService
	machine *vm.VM
}

// NewApp creates a new App application struct
func NewApp() *App {
	machine := vm.NewVM()
	// Use correct stack top (stack segment is already created by NewVM)
	stackTop := uint32(vm.StackSegmentStart + vm.StackSegmentSize)
	machine.InitializeStack(stackTop)

	return &App{
		machine: machine,
		service: service.NewDebuggerService(machine),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	debugLog.Println("startup() called")
	a.ctx = ctx
	a.service.SetContext(ctx)
	debugLog.Println("startup() completed")
}

// stripComments removes all comments from a line (inline and block)
// Supports: ; @ // line comments and /* */ block comments
func stripComments(line string) string {
	// Handle block comments first (/* */)
	for {
		start := strings.Index(line, "/*")
		if start == -1 {
			break
		}
		end := strings.Index(line[start:], "*/")
		if end == -1 {
			// Unclosed block comment, treat rest of line as comment
			line = line[:start]
			break
		}
		// Remove block comment
		line = line[:start] + line[start+end+2:]
	}

	// Find first line comment marker (; @ //)
	// This ensures we only cut at the first comment, not process each marker separately
	firstComment := len(line)
	for _, marker := range []string{";", "@", "//"} {
		if idx := strings.Index(line, marker); idx != -1 && idx < firstComment {
			firstComment = idx
		}
	}
	if firstComment < len(line) {
		line = line[:firstComment]
	}

	return strings.TrimSpace(line)
}

// LoadProgramFromSource parses and loads assembly source code
func (a *App) LoadProgramFromSource(source string, filename string, entryPoint uint32) error {
	// Input validation
	const maxSourceSize = 1024 * 1024 // 1MB limit
	if len(source) > maxSourceSize {
		return fmt.Errorf("source code too large: %d bytes (maximum %d bytes)", len(source), maxSourceSize)
	}

	// Validate entry point is within valid code segment range
	// Code segment: 0x8000 to 0x18000 (64KB)
	if entryPoint < vm.CodeSegmentStart || entryPoint >= vm.CodeSegmentStart+vm.CodeSegmentSize {
		return fmt.Errorf("invalid entry point: 0x%X (must be between 0x%X and 0x%X)",
			entryPoint, vm.CodeSegmentStart, vm.CodeSegmentStart+vm.CodeSegmentSize-1)
	}

	// Check if .org directive is already present by parsing, not just searching
	// This avoids false positives from comments or strings containing ".org"
	hasOrgDirective := false
	for _, line := range strings.Split(source, "\n") {
		// Strip all comments (inline and block)
		stripped := stripComments(line)
		if stripped == "" {
			continue
		}
		// Check for .org directive with word boundary (not .organize, etc.)
		// Must be followed by whitespace or end of line
		if strings.HasPrefix(stripped, ".org") {
			rest := stripped[4:]
			if len(rest) == 0 || rest[0] == ' ' || rest[0] == '\t' {
				hasOrgDirective = true
				break
			}
		}
	}

	if !hasOrgDirective {
		source = fmt.Sprintf(".org 0x%X\n%s", entryPoint, source)
	}

	p := parser.NewParser(source, filename)
	program, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	return a.service.LoadProgram(program, entryPoint)
}

// LoadProgramFromFile opens a file dialog and loads an ARM assembly program
func (a *App) LoadProgramFromFile() error {
	// Open file dialog
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Load ARM Assembly Program",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "ARM Assembly Files (*.s, *.asm)",
				Pattern:     "*.s;*.asm",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to open file dialog: %w", err)
	}

	// User cancelled
	if filePath == "" {
		return nil
	}

	// Validate file size before reading (1MB limit)
	const maxSourceSize = 1024 * 1024
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	if info.Size() > maxSourceSize {
		return fmt.Errorf("file too large: %d bytes (maximum %d bytes)", info.Size(), maxSourceSize)
	}

	// Read file contents
	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse and load program with default entry point (code segment start)
	err = a.LoadProgramFromSource(string(source), filePath, vm.CodeSegmentStart)
	if err != nil {
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
		return err
	}

	runtime.EventsEmit(a.ctx, "vm:state-changed")
	runtime.EventsEmit(a.ctx, "vm:program-loaded", filePath)
	return nil
}

// GetRegisters returns current register state
func (a *App) GetRegisters() service.RegisterState {
	return a.service.GetRegisterState()
}

// Step executes a single instruction
func (a *App) Step() error {
	debugLog.Println("Step() called")
	err := a.service.Step()
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	} else {
		debugLog.Printf("Step() error: %v", err)
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
	}
	debugLog.Println("Step() completed")
	return err
}

// Continue runs until breakpoint or halt (asynchronously)
func (a *App) Continue() error {
	debugLog.Println("Continue() called - starting goroutine")
	// Capture context to avoid race condition
	ctx := a.ctx
	// Run in background to avoid blocking GUI
	go func() {
		debugLog.Println("Goroutine started, calling RunUntilHalt")
		start := time.Now()
		err := a.service.RunUntilHalt()
		elapsed := time.Since(start)
		debugLog.Printf("RunUntilHalt completed in %v, err: %v", elapsed, err)

		debugLog.Println("Emitting vm:state-changed")
		runtime.EventsEmit(ctx, "vm:state-changed")

		if err != nil {
			debugLog.Printf("Emitting error: %v", err)
			runtime.EventsEmit(ctx, "vm:error", err.Error())
		}

		// Check if stopped at breakpoint
		state := a.service.GetExecutionState()
		debugLog.Printf("Execution state: %s", state)
		if state == service.StateBreakpoint {
			debugLog.Println("Emitting breakpoint-hit")
			runtime.EventsEmit(ctx, "vm:breakpoint-hit")
		}
		debugLog.Println("Goroutine completed")
	}()

	debugLog.Println("Continue() returning")
	return nil
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
	debugLog.Printf("GetMemory called: address=0x%08X, size=%d", address, size)
	data, err := a.service.GetMemory(address, size)
	if err != nil {
		debugLog.Printf("GetMemory error: %v", err)
	} else {
		debugLog.Printf("GetMemory success: returned %d bytes", len(data))
	}
	return data, err
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

// GetLastMemoryWrite returns the address of the last memory write
func (a *App) GetLastMemoryWrite() service.MemoryWriteInfo {
	result := a.service.GetLastMemoryWrite()
	debugLog.Printf("GetLastMemoryWrite: address=0x%08X, hasWrite=%v", result.Address, result.HasWrite)
	return result
}

// GetSymbolForAddress resolves address to symbol
func (a *App) GetSymbolForAddress(addr uint32) string {
	return a.service.GetSymbolForAddress(addr)
}

// GetSymbolsForAddresses resolves multiple addresses to symbols in one call
func (a *App) GetSymbolsForAddresses(addrs []uint32) map[uint32]string {
	result := make(map[uint32]string, len(addrs))
	for _, addr := range addrs {
		symbol := a.service.GetSymbolForAddress(addr)
		if symbol != "" {
			result[addr] = symbol
		}
	}
	return result
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

// GetWatchpoints returns all watchpoints
func (a *App) GetWatchpoints() []service.WatchpointInfo {
	return a.service.GetWatchpoints()
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
