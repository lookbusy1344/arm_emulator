package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lookbusy1344/arm-emulator/debugger"
	"github.com/lookbusy1344/arm-emulator/loader"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

const (
	// Validator limits for API safety
	maxDisassemblyCount = 1000   // Maximum number of instructions to disassemble
	maxStackCount       = 1000   // Maximum number of stack entries to return
	maxStackOffset      = 100000 // Maximum stack offset to prevent wraparound attacks
	stepsBeforeYield    = 1000   // Yield every N steps during execution
)

var serviceLog *log.Logger

func init() {
	// Check if debug logging is enabled via environment variable
	if os.Getenv("ARM_EMULATOR_DEBUG") != "" {
		// Create debug log file.
		// Note: File handle intentionally not closed - kept open for process lifetime.
		// This is acceptable for debug logging; the OS cleans up on process exit.
		logPath := filepath.Join(os.TempDir(), "arm-emulator-service-debug.log")
		f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600) // #nosec G304 -- fixed filename in temp dir
		if err != nil {
			serviceLog = log.New(os.Stderr, "SERVICE: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
		} else {
			serviceLog = log.New(f, "SERVICE: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
		}
	} else {
		// Disable logging by default
		serviceLog = log.New(io.Discard, "", 0)
	}
}

// DebuggerService provides a thread-safe interface to debugger functionality
// This service is shared by TUI, GUI, and CLI interfaces
//
// Lock Ordering:
// The service uses its own sync.RWMutex (s.mu) to protect all field access,
// including access to the debugger. When calling Debugger methods that have
// their own internal mutex (like ShouldBreak), the lock order is:
// s.mu -> debugger.mu
//
// This is safe because:
// - The TUI uses the Debugger's internal mutex directly (no service mutex)
// - The service always acquires s.mu before any Debugger method that uses d.mu
// - The GUI only accesses debugger state through the service
//
// Do NOT acquire locks in the reverse order (debugger.mu -> s.mu) as this
// would create a deadlock risk.
type DebuggerService struct {
	mu                   sync.RWMutex
	vm                   *vm.VM
	debugger             *debugger.Debugger
	symbols              map[string]uint32
	sourceMap            []SourceMapEntry  // Address to source line mapping with line numbers
	sourceMapByAddr      map[uint32]string // Quick lookup by address (for debugger)
	program              *parser.Program
	entryPoint           uint32
	outputWriter         *EventEmittingWriter
	ctx                  context.Context
	stateChangedCallback func() // Callback for GUI state updates

	// stdin redirection for guest programs (GUI)
	stdinPipeReader *io.PipeReader
	stdinPipeWriter *io.PipeWriter
	stdinBuffer     strings.Builder // Buffer for stdin sent before execution starts
}

// NewDebuggerService creates a new debugger service
func NewDebuggerService(machine *vm.VM) *DebuggerService {
	// Setup stdin pipe for guest program input (GUI)
	stdinReader, stdinWriter := io.Pipe()
	machine.SetStdinReader(stdinReader)

	return &DebuggerService{
		vm:              machine,
		debugger:        debugger.NewDebugger(machine),
		symbols:         make(map[string]uint32),
		sourceMap:       nil,
		sourceMapByAddr: make(map[uint32]string),
		stdinPipeReader: stdinReader,
		stdinPipeWriter: stdinWriter,
	}
}

// GetVM returns the underlying VM (for testing)
func (s *DebuggerService) GetVM() *vm.VM {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.vm
}

// SetContext sets the Wails context for event emission
func (s *DebuggerService) SetContext(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx = ctx
}

// SetStateChangedCallback sets a callback for GUI state updates during execution
func (s *DebuggerService) SetStateChangedCallback(callback func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stateChangedCallback = callback
}

// LoadProgram loads and initializes a parsed program
func (s *DebuggerService) LoadProgram(program *parser.Program, entryPoint uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.program = program
	s.entryPoint = entryPoint

	// Extract symbols
	s.symbols = make(map[string]uint32)
	for name, symbol := range program.SymbolTable.GetAllSymbols() {
		if symbol.Type == parser.SymbolLabel {
			s.symbols[name] = symbol.Value
		}
	}

	// Build source map with line numbers
	s.sourceMap = nil
	s.sourceMapByAddr = make(map[uint32]string)
	for _, inst := range program.Instructions {
		entry := SourceMapEntry{
			Address:    inst.Address,
			LineNumber: inst.Pos.Line,
			Line:       inst.RawLine,
		}
		s.sourceMap = append(s.sourceMap, entry)
		s.sourceMapByAddr[inst.Address] = inst.RawLine
	}
	// Note: Data directives are excluded from breakpoint-valid locations
	// but kept in sourceMapByAddr for debugger display
	for _, dir := range program.Directives {
		if dir.Name == ".word" || dir.Name == ".byte" || dir.Name == ".ascii" ||
			dir.Name == ".asciz" || dir.Name == ".space" {
			s.sourceMapByAddr[dir.Address] = "[DATA]" + dir.RawLine
		}
	}

	// Create output buffer with event emission
	// IMPORTANT: Only set OutputWriter if it hasn't been configured already.
	// The API server sets up EventWriter for WebSocket broadcasting before calling LoadProgram.
	// The GUI (Wails) doesn't pre-configure OutputWriter, so we set up EventEmittingWriter for it.
	if s.vm.OutputWriter == os.Stdout {
		// OutputWriter is still the default (os.Stdout), so set up event emission
		outputBuffer := &bytes.Buffer{}
		s.outputWriter = NewEventEmittingWriter(outputBuffer, s.ctx)
		s.vm.OutputWriter = s.outputWriter
	}
	// else: OutputWriter was already configured (e.g., by API layer), leave it alone

	// Load into debugger
	s.debugger.LoadSymbols(s.symbols)
	s.debugger.LoadSourceMap(s.sourceMapByAddr)

	// Load into VM memory
	if err := loader.LoadProgramIntoVM(s.vm, program, entryPoint); err != nil {
		return err
	}

	// Initialize stack pointer only if not already set (preserve InitializeStack value)
	// Stack grows downward from top of stack segment
	if s.vm.StackTop == 0 {
		s.vm.StackTop = vm.StackSegmentStart + vm.StackSegmentSize
		if err := s.vm.CPU.SetSP(s.vm.StackTop); err != nil {
			return fmt.Errorf("failed to initialize stack pointer: %w", err)
		}
	}

	// Reset execution state to halted (not running until execution begins)
	s.vm.State = vm.StateHalted
	s.debugger.Running = false

	return nil
}

// GetRegisterState returns current register state (thread-safe)
func (s *DebuggerService) GetRegisterState() RegisterState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Build 16-register array: R0-R14 + PC at R15
	var regs [16]uint32
	copy(regs[:15], s.vm.CPU.R[:])
	regs[15] = s.vm.CPU.PC

	return RegisterState{
		Registers: regs,
		CPSR: CPSRState{
			N: s.vm.CPU.CPSR.N,
			Z: s.vm.CPU.CPSR.Z,
			C: s.vm.CPU.CPSR.C,
			V: s.vm.CPU.CPSR.V,
		},
		PC:     s.vm.CPU.PC,
		Cycles: s.vm.CPU.Cycles,
	}
}

// Step executes a single instruction
func (s *DebuggerService) Step() error {
	s.mu.Lock()
	// Release lock BEFORE Step() because Step() may block on stdin read.
	// This allows SendInput() to acquire RLock and write to the stdin pipe.
	s.mu.Unlock()

	return s.vm.Step()
}

// Continue runs until breakpoint or halt
func (s *DebuggerService) Continue() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.debugger.Running = true
	s.debugger.StepMode = debugger.StepNone

	return nil
}

// Pause pauses execution and sets VM state to halted
func (s *DebuggerService) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debugger.Running = false
	s.vm.State = vm.StateHalted
}

// Reset performs a complete reset to initial state
// This clears the loaded program, all breakpoints, and resets the VM to pristine state
// Use ResetToEntryPoint() if you want to restart the current program instead
func (s *DebuggerService) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Full VM reset: clears all registers (PC=0), memory, and execution state
	s.vm.Reset()

	// Reset stdin reader to prevent hangs when stdin was redirected by GUI
	// This ensures clean stdin state for the next program
	s.vm.ResetStdinReader()

	// Clear loaded program and associated metadata
	s.program = nil
	s.entryPoint = 0
	s.vm.EntryPoint = 0
	s.vm.StackTop = 0
	s.symbols = make(map[string]uint32)
	s.sourceMap = nil
	s.sourceMapByAddr = make(map[uint32]string)

	// Clear all breakpoints and watchpoints
	s.debugger.Breakpoints.Clear()
	// Note: WatchpointManager doesn't have Clear() - could add if needed

	// Reset execution control
	s.debugger.Running = false
	s.vm.State = vm.StateHalted

	return nil
}

// ResetToEntryPoint resets VM to program entry point without clearing the loaded program
// This is useful for restarting execution of the current program
func (s *DebuggerService) ResetToEntryPoint() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.program == nil {
		// No program loaded, perform full reset
		s.vm.Reset()
		s.vm.State = vm.StateHalted
		s.debugger.Running = false
		return nil
	}

	// Reset registers and execution state but preserve memory contents
	if err := s.vm.ResetRegisters(); err != nil {
		return fmt.Errorf("failed to reset registers: %w", err)
	}
	s.debugger.Running = false

	return nil
}

// GetExecutionState returns current execution state
func (s *DebuggerService) GetExecutionState() ExecutionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return VMStateToExecution(s.vm.State)
}

// AddBreakpoint adds a breakpoint at the specified address
func (s *DebuggerService) AddBreakpoint(address uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate that the address corresponds to actual code (not data)
	// Use sourceMapByAddr which contains both code and data entries
	line, exists := s.sourceMapByAddr[address]
	if !exists {
		return fmt.Errorf("invalid breakpoint address: 0x%X does not correspond to executable code", address)
	}
	// Reject data locations (prefixed with [DATA])
	if len(line) >= 6 && line[:6] == "[DATA]" {
		return fmt.Errorf("invalid breakpoint address: 0x%X is a data location, not executable code", address)
	}

	s.debugger.Breakpoints.AddBreakpoint(address, false, "")
	return nil
}

// RemoveBreakpoint removes a breakpoint
func (s *DebuggerService) RemoveBreakpoint(address uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.debugger.Breakpoints.DeleteBreakpointAt(address)
}

// GetBreakpoints returns all breakpoints
func (s *DebuggerService) GetBreakpoints() []BreakpointInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bps := s.debugger.Breakpoints.GetAllBreakpoints()
	result := make([]BreakpointInfo, len(bps))
	for i, bp := range bps {
		result[i] = BreakpointInfo{
			Address: bp.Address,
			Enabled: bp.Enabled,
		}
	}
	return result
}

// ClearAllBreakpoints removes all breakpoints
func (s *DebuggerService) ClearAllBreakpoints() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debugger.Breakpoints.Clear()
}

// GetMemory returns memory contents for a region
func (s *DebuggerService) GetMemory(address uint32, size uint32) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	serviceLog.Printf("GetMemory: address=0x%08X, size=%d", address, size)
	data := make([]byte, size)
	for i := uint32(0); i < size; i++ {
		b, err := s.vm.Memory.ReadByteAt(address + i)
		if err != nil {
			serviceLog.Printf("GetMemory: ReadByteAt failed at offset %d: %v", i, err)
			// Return 0 for unmapped or unreadable memory instead of failing the whole request
			// This allows the memory view to show partial results at segment boundaries
			data[i] = 0
			continue
		}
		data[i] = b
	}
	serviceLog.Printf("GetMemory: success, returning %d bytes", len(data))
	return data, nil
}

// GetLastMemoryWrite returns the address of the last memory write and clears the flag
func (s *DebuggerService) GetLastMemoryWrite() MemoryWriteInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := MemoryWriteInfo{
		Address:  s.vm.LastMemoryWrite,
		Size:     s.vm.LastMemoryWriteSize,
		HasWrite: s.vm.HasMemoryWrite,
	}
	serviceLog.Printf("GetLastMemoryWrite: address=0x%08X, size=%d, hasWrite=%v", result.Address, result.Size, result.HasWrite)
	s.vm.HasMemoryWrite = false
	return result
}

// GetSourceLine returns the source line for an address
func (s *DebuggerService) GetSourceLine(address uint32) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sourceMapByAddr[address]
}

// GetSourceMap returns the source map entries with line numbers
func (s *DebuggerService) GetSourceMap() []SourceMapEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return copy of source map to prevent external modification
	result := make([]SourceMapEntry, len(s.sourceMap))
	copy(result, s.sourceMap)
	return result
}

// GetSourceMapByAddr returns address-to-line lookup (for debugger display)
func (s *DebuggerService) GetSourceMapByAddr() map[uint32]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return copy to prevent external modification
	result := make(map[uint32]string, len(s.sourceMapByAddr))
	for addr, line := range s.sourceMapByAddr {
		result[addr] = line
	}
	return result
}

// GetSymbols returns all symbols
func (s *DebuggerService) GetSymbols() map[string]uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	symbols := make(map[string]uint32, len(s.symbols))
	for k, v := range s.symbols {
		symbols[k] = v
	}
	return symbols
}

// GetSymbolForAddress resolves an address to a symbol name
func (s *DebuggerService) GetSymbolForAddress(addr uint32) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if there's a symbol at this address
	for name, symbolAddr := range s.symbols {
		if symbolAddr == addr {
			return name
		}
	}

	return ""
}

// RunUntilHalt runs program until halt or breakpoint
// If Running is already false (e.g., paused before goroutine started), returns immediately.
// This handles the race where Pause() is called between Continue() setting Running=true
// and this function starting execution.
func (s *DebuggerService) RunUntilHalt() error {
	serviceLog.Println("RunUntilHalt() called")
	s.mu.Lock()
	// Check if already paused before we started (handles race with Pause())
	if !s.debugger.Running {
		serviceLog.Println("RunUntilHalt() - already paused, exiting early")
		s.mu.Unlock()
		return nil
	}

	// Flush any buffered stdin to the pipe in a background goroutine
	// This supports the batch stdin pattern where input is sent before calling run
	// We use a goroutine because pipe writes block until there's a reader,
	// but the reader only starts when the VM execution loop begins
	if s.stdinBuffer.Len() > 0 {
		buffered := s.stdinBuffer.String()
		s.stdinBuffer.Reset()
		serviceLog.Printf("Flushing %d bytes of buffered stdin in background", len(buffered))

		// Launch goroutine to write to pipe (won't block RunUntilHalt)
		go func() {
			if _, err := s.stdinPipeWriter.Write([]byte(buffered)); err != nil {
				serviceLog.Printf("Error writing buffered stdin to pipe: %v", err)
			}
		}()
	}

	s.vm.State = vm.StateRunning
	s.mu.Unlock()

	stepCount := 0

	for {
		s.mu.Lock()
		if !s.debugger.Running || s.vm.State != vm.StateRunning {
			serviceLog.Printf("Exiting loop: Running=%v, State=%v", s.debugger.Running, s.vm.State)
			s.mu.Unlock()
			break
		}

		// Check breakpoints
		if shouldBreak, _ := s.debugger.ShouldBreak(); shouldBreak {
			serviceLog.Println("Breakpoint hit")
			s.debugger.Running = false
			s.vm.State = vm.StateBreakpoint
			s.mu.Unlock()
			break
		}

		// Capture values needed for step
		pc := s.vm.CPU.PC

		// Release lock BEFORE Step() because Step() may block on stdin read.
		// This allows SendInput() to acquire RLock and write to the stdin pipe.
		s.mu.Unlock()

		// Execute step (without holding lock - Step may block on stdin)
		err := s.vm.Step()

		// Reacquire lock to check state
		s.mu.Lock()
		halted := s.vm.State == vm.StateHalted
		s.mu.Unlock()

		if stepCount == 0 {
			serviceLog.Printf("Executing at PC=0x%08X", pc)
		}

		// If error but VM is halted, it's normal program termination (SWI #0)
		if err != nil && !halted {
			serviceLog.Printf("Step error: %v", err)
			s.mu.Lock()
			s.debugger.Running = false
			s.mu.Unlock()
			return err
		}

		if halted {
			serviceLog.Println("VM halted")
			s.mu.Lock()
			s.debugger.Running = false
			s.mu.Unlock()
			break
		}

		// Periodically yield to allow GUI to query state
		stepCount++
		if stepCount >= stepsBeforeYield {
			serviceLog.Printf("Yielding after %d steps", stepCount)
			stepCount = 0
			// Brief sleep to yield to scheduler and allow GUI queries
			time.Sleep(1 * time.Millisecond)
		}
	}

	serviceLog.Println("RunUntilHalt() completed")
	return nil
}

// IsRunning returns whether execution is in progress
func (s *DebuggerService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.debugger.Running
}

// SetRunning sets the running state synchronously
// Used by async execution methods to set state before launching goroutines
func (s *DebuggerService) SetRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debugger.Running = running
	if running {
		s.vm.State = vm.StateRunning
	} else {
		// Don't override other states (halted, error, breakpoint)
		if s.vm.State == vm.StateRunning {
			s.vm.State = vm.StateHalted
		}
	}
}

// GetExitCode returns the program exit code
func (s *DebuggerService) GetExitCode() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.vm.ExitCode
}

// GetOutput returns captured program output (clears buffer)
func (s *DebuggerService) GetOutput() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.outputWriter == nil {
		return ""
	}

	return s.outputWriter.GetBufferAndClear()
}

// GetDisassembly returns disassembled instructions starting at address.
// Returns an empty slice if inputs are invalid or memory reads fail.
// Truncates the result if memory errors occur before count is reached.
//
// Parameters:
//   - startAddr: must be 4-byte aligned (ARM requirement)
//   - count: must be positive and <= maxDisassemblyCount
func (s *DebuggerService) GetDisassembly(startAddr uint32, count int) []DisassemblyLine {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate inputs
	if count <= 0 || count > maxDisassemblyCount {
		return []DisassemblyLine{}
	}
	if startAddr&0x3 != 0 { // Check 4-byte alignment
		return []DisassemblyLine{}
	}

	if s.vm == nil {
		return []DisassemblyLine{}
	}

	lines := make([]DisassemblyLine, 0, count)
	addr := startAddr
	// Clamp to first mapped code address to avoid empty results when startAddr is below code segment
	if addr < vm.CodeSegmentStart {
		addr = vm.CodeSegmentStart
	}

	for i := 0; i < count; i++ {
		// Read instruction from memory
		opcode, err := s.vm.Memory.ReadWord(addr)
		if err != nil {
			// Memory read error - return what we have so far (truncated result)
			break
		}

		// Get symbol at this address if any (use unsafe version since we already hold RLock)
		symbol := s.getSymbolForAddressUnsafe(addr)

		// Get mnemonic from source map if available
		mnemonic := ""
		if sourceLine, ok := s.sourceMapByAddr[addr]; ok {
			mnemonic = sourceLine
		}

		line := DisassemblyLine{
			Address:  addr,
			Opcode:   opcode,
			Mnemonic: mnemonic,
			Symbol:   symbol,
		}

		lines = append(lines, line)
		addr += 4 // ARM instructions are 4 bytes
	}

	return lines
}

// GetStack returns stack contents from SP+offset.
// Returns an empty slice if inputs are invalid or memory reads fail.
//
// Parameters:
//   - offset: stack offset in words (multiplied by 4 for byte offset).
//     Must be in range [-maxStackOffset, maxStackOffset] to prevent wraparound attacks.
//   - count: number of stack entries to read. Must be positive and <= maxStackCount.
//
// The function performs safe arithmetic with overflow detection to prevent
// integer wraparound vulnerabilities.
func (s *DebuggerService) GetStack(offset int, count int) []StackEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate inputs
	if count <= 0 || count > maxStackCount {
		return []StackEntry{}
	}

	// Validate offset to prevent wraparound attacks
	if offset < -maxStackOffset || offset > maxStackOffset {
		return []StackEntry{}
	}

	if s.vm == nil {
		return []StackEntry{}
	}

	entries := make([]StackEntry, 0, count)
	sp := s.vm.CPU.R[13] // R13 is SP

	// Safe calculation with overflow check
	offsetBytes := int64(offset) * 4
	newAddr := int64(sp) + offsetBytes

	// Check for wraparound
	if newAddr < 0 || newAddr > 0xFFFFFFFF {
		return []StackEntry{}
	}

	startAddr := uint32(newAddr)

	for i := 0; i < count; i++ {
		// Safe calculation with overflow check for loop iteration
		addrOffset := int64(i) * 4
		nextAddr := int64(startAddr) + addrOffset

		// Check for wraparound during iteration
		if nextAddr < 0 || nextAddr > 0xFFFFFFFF {
			// Address wrapped around - return what we have so far
			break
		}

		addr := uint32(nextAddr)

		// Read value from memory
		value, err := s.vm.Memory.ReadWord(addr)
		if err != nil {
			// Memory read error - return what we have so far (truncated result)
			break
		}

		// Check if value points to a symbol
		symbol := s.getSymbolForAddressUnsafe(value)

		entry := StackEntry{
			Address: addr,
			Value:   value,
			Symbol:  symbol,
		}

		entries = append(entries, entry)
	}

	return entries
}

// getSymbolForAddressUnsafe is the internal version without locking
func (s *DebuggerService) getSymbolForAddressUnsafe(addr uint32) string {
	// Check if there's a symbol at this address
	for name, symbolAddr := range s.symbols {
		if symbolAddr == addr {
			return name
		}
	}
	return ""
}

// StepOver executes one instruction, stepping over function calls
func (s *DebuggerService) StepOver() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil || s.program == nil {
		return fmt.Errorf("no program loaded")
	}

	// Use debugger's SetStepOver to configure mode
	s.debugger.SetStepOver()

	// Execute until step completes
	for s.debugger.Running {
		// Check if we should break
		if s.debugger.StepMode != debugger.StepSingle {
			if shouldBreak, _ := s.debugger.ShouldBreak(); shouldBreak {
				s.debugger.Running = false
				break
			}
		}

		// Release lock BEFORE Step() because Step() may block on stdin read.
		s.mu.Unlock()

		// Execute one instruction
		err := s.vm.Step()

		// Re-acquire lock
		s.mu.Lock()

		if err != nil {
			s.debugger.Running = false
			return err
		}

		// For single-step mode, check after execution
		if s.debugger.StepMode == debugger.StepSingle {
			if shouldBreak, _ := s.debugger.ShouldBreak(); shouldBreak {
				s.debugger.Running = false
				break
			}
		}
	}

	return nil
}

// StepOut executes until the current function returns
func (s *DebuggerService) StepOut() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil || s.program == nil {
		return fmt.Errorf("no program loaded")
	}

	// Use debugger's public method instead of accessing fields directly
	s.debugger.SetStepOut()

	return nil
}

// AddWatchpoint adds a watchpoint at the specified address
func (s *DebuggerService) AddWatchpoint(address uint32, watchType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil {
		return fmt.Errorf("no program loaded")
	}

	// Convert string type to debugger.WatchType
	var wpType debugger.WatchType
	switch watchType {
	case "read":
		wpType = debugger.WatchRead
	case "write":
		wpType = debugger.WatchWrite
	case "readwrite":
		wpType = debugger.WatchReadWrite
	default:
		return fmt.Errorf("invalid watchpoint type: %s", watchType)
	}

	// Add watchpoint (address watchpoint, not register)
	// expression is the formatted address, isRegister=false, register=0
	expression := fmt.Sprintf("[0x%08X]", address)
	s.debugger.Watchpoints.AddWatchpoint(wpType, expression, address, false, 0)

	return nil
}

// RemoveWatchpoint removes a watchpoint by ID
func (s *DebuggerService) RemoveWatchpoint(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil {
		return fmt.Errorf("no program loaded")
	}

	return s.debugger.Watchpoints.DeleteWatchpoint(id)
}

// GetWatchpoints returns all watchpoints
func (s *DebuggerService) GetWatchpoints() []WatchpointInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.debugger == nil {
		return []WatchpointInfo{}
	}

	wps := s.debugger.Watchpoints.GetAllWatchpoints()
	result := make([]WatchpointInfo, len(wps))
	for i, wp := range wps {
		// Convert debugger.WatchType to string
		var wpType string
		switch wp.Type {
		case debugger.WatchRead:
			wpType = "read"
		case debugger.WatchWrite:
			wpType = "write"
		case debugger.WatchReadWrite:
			wpType = "readwrite"
		}

		result[i] = WatchpointInfo{
			ID:      wp.ID,
			Address: wp.Address,
			Type:    wpType,
			Enabled: wp.Enabled,
		}
	}
	return result
}

// ExecuteCommand executes a debugger command and returns output
func (s *DebuggerService) ExecuteCommand(command string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil {
		return "", fmt.Errorf("no program loaded")
	}

	// Execute command (debugger writes to its Output buffer)
	err := s.debugger.ExecuteCommand(command)

	// Get output and clear buffer
	output := s.debugger.GetOutput()

	return output, err
}

// EvaluateExpression evaluates an expression and returns the result
func (s *DebuggerService) EvaluateExpression(expr string) (uint32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil || s.debugger.Evaluator == nil {
		return 0, fmt.Errorf("no program loaded")
	}

	return s.debugger.Evaluator.EvaluateExpression(expr, s.vm, s.symbols)
}

// SendInput sends user input to the guest program's stdin
// This is called from the GUI frontend when the user provides input
func (s *DebuggerService) SendInput(input string) error {
	if s.stdinPipeWriter == nil {
		return fmt.Errorf("stdin pipe not initialized")
	}

	// Check if VM is running or waiting for input
	// If not running and not waiting, buffer the input for later (batch stdin pattern)
	s.mu.RLock()
	running := s.debugger.Running
	waiting := s.vm.State == vm.StateWaitingForInput
	s.mu.RUnlock()

	if !running && !waiting {
		s.mu.Lock()
		// Note: input should already include newline from API layer
		s.stdinBuffer.WriteString(input)
		s.mu.Unlock()
		serviceLog.Printf("SendInput: Buffered %d bytes for later", len(input))
		return nil
	}

	// VM is running or waiting for input - echo to output and write to pipe
	// NOTE: No mutex lock for pipe write! io.Pipe is already thread-safe.
	// Taking a lock here causes deadlock when RunUntilHalt holds the lock while blocked on stdin read.

	// Echo the input to the output window so the user can see what they typed
	// Use RLock to safely access OutputWriter
	s.mu.RLock()
	outputWriter := s.vm.OutputWriter
	s.mu.RUnlock()

	if outputWriter != nil {
		_, _ = outputWriter.Write([]byte(input + "\n"))
	}

	// Write input + newline to the stdin pipe (io.Pipe.Write is thread-safe)
	_, err := s.stdinPipeWriter.Write([]byte(input + "\n"))
	return err
}

// EnableExecutionTrace enables execution tracing
func (s *DebuggerService) EnableExecutionTrace() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create execution trace if it doesn't exist
	if s.vm.ExecutionTrace == nil {
		// Use a bytes buffer for the trace output
		var buf bytes.Buffer
		s.vm.ExecutionTrace = vm.NewExecutionTrace(&buf)
		// Load symbols if available
		if len(s.symbols) > 0 {
			s.vm.ExecutionTrace.LoadSymbols(s.symbols)
		}
	}

	s.vm.ExecutionTrace.Enabled = true
	s.vm.ExecutionTrace.Start()
	return nil
}

// DisableExecutionTrace disables execution tracing
func (s *DebuggerService) DisableExecutionTrace() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vm.ExecutionTrace != nil {
		s.vm.ExecutionTrace.Enabled = false
	}
}

// GetExecutionTraceData returns execution trace entries
func (s *DebuggerService) GetExecutionTraceData() ([]vm.TraceEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.vm.ExecutionTrace == nil {
		return []vm.TraceEntry{}, nil
	}

	return s.vm.ExecutionTrace.GetEntries(), nil
}

// ClearExecutionTrace clears execution trace entries
func (s *DebuggerService) ClearExecutionTrace() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vm.ExecutionTrace != nil {
		s.vm.ExecutionTrace.Clear()
	}
}

// EnableStatistics enables performance statistics collection
func (s *DebuggerService) EnableStatistics() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create statistics collector if it doesn't exist
	if s.vm.Statistics == nil {
		s.vm.Statistics = vm.NewPerformanceStatistics()
	}

	s.vm.Statistics.Enabled = true
	s.vm.Statistics.Start()
	return nil
}

// DisableStatistics disables performance statistics collection
func (s *DebuggerService) DisableStatistics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vm.Statistics != nil {
		s.vm.Statistics.Enabled = false
	}
}

// GetStatistics returns performance statistics
func (s *DebuggerService) GetStatistics() (*vm.PerformanceStatistics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.vm.Statistics == nil {
		return nil, fmt.Errorf("statistics not enabled")
	}

	// Finalize statistics before returning
	s.vm.Statistics.Finalize()

	return s.vm.Statistics, nil
}
