package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

var (
	// ErrExecutionInProgress is returned when guest code is already executing.
	ErrExecutionInProgress = errors.New("execution in progress")
	// ErrClosed is returned after Close.
	ErrClosed = errors.New("debugger service closed")
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
// This service is shared by TUI, API/GUI frontends, and CLI interfaces
//
// Concurrency:
// mu guards every field and all access to the VM. At most one goroutine executes guest
// code at a time (Step, StepOver, RunUntilHalt); it holds mu while an instruction runs
// and releases it between instructions and while the guest waits for input. Calls that
// replace VM state (LoadProgram, Reset, ResetToEntryPoint, Pause, Close) first stop that
// goroutine and wait for it to finish.
//
// Lock order: s.mu -> debugger.mu. Never acquire them in the reverse order.
type DebuggerService struct {
	mu                   sync.Mutex
	vm                   *vm.VM
	debugger             *debugger.Debugger
	symbols              map[string]uint32
	sourceMap            []SourceMapEntry  // Address to source line mapping with line numbers
	sourceMapByAddr      map[uint32]string // Quick lookup by address (for debugger)
	program              *parser.Program
	entryPoint           uint32
	outputBuffer         *bytes.Buffer // Output buffer for VM output (when not using API)
	stateChangedCallback func()        // Callback for GUI state updates

	executing bool       // a goroutine is executing guest code
	atEntry   bool       // no instruction has run since load or reset; run stops at an entry breakpoint
	closed    bool       // Close was called; no further execution
	execDone  *sync.Cond // signalled when executing becomes false

	// Guest stdin. SendInput appends to stdin; the VM reads it through stdinSource.
	stdin            []byte
	stdinInterrupted bool       // pending reads fail with vm.ErrInputInterrupted
	stdinReady       *sync.Cond // signalled when stdin grows or reads are interrupted
}

// stdinSource feeds queued input to the guest. The VM reads it inside vm.Step, which
// runs with s.mu held. Waiting for input releases s.mu so other calls can proceed.
type stdinSource struct{ s *DebuggerService }

func (r stdinSource) Read(p []byte) (int, error) {
	s := r.s
	for len(s.stdin) == 0 && !s.stdinInterrupted {
		s.stdinReady.Wait()
	}
	if s.stdinInterrupted {
		return 0, vm.ErrInputInterrupted
	}
	n := copy(p, s.stdin)
	s.stdin = s.stdin[n:]
	return n, nil
}

// NewDebuggerService creates a new debugger service
func NewDebuggerService(machine *vm.VM) *DebuggerService {
	s := &DebuggerService{
		vm:              machine,
		debugger:        debugger.NewDebugger(machine),
		symbols:         make(map[string]uint32),
		sourceMap:       nil,
		sourceMapByAddr: make(map[uint32]string),
	}
	s.execDone = sync.NewCond(&s.mu)
	s.stdinReady = sync.NewCond(&s.mu)
	machine.SetStdinReader(stdinSource{s})
	return s
}

// beginExecutionLocked claims the VM for guest execution. The caller holds s.mu and
// must call endExecutionLocked when done.
func (s *DebuggerService) beginExecutionLocked() error {
	if s.closed {
		return ErrClosed
	}
	if s.executing {
		return ErrExecutionInProgress
	}
	s.executing = true
	s.atEntry = false
	s.stdinInterrupted = false
	return nil
}

func (s *DebuggerService) endExecutionLocked() {
	s.executing = false
	s.execDone.Broadcast()
}

// stopExecutionLocked stops guest execution and waits for it to end. A blocked input
// read is abandoned without completing. The caller holds s.mu.
func (s *DebuggerService) stopExecutionLocked() {
	s.debugger.Running = false
	if !s.executing {
		return
	}
	s.stdinInterrupted = true
	s.stdinReady.Broadcast()
	for s.executing {
		s.execDone.Wait()
	}
}

// Close stops guest execution and rejects further execution.
func (s *DebuggerService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.stopExecutionLocked()
}

// GetVM returns the underlying VM (for testing)
func (s *DebuggerService) GetVM() *vm.VM {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vm
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
	s.stopExecutionLocked()

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

	// Note: OutputWriter should be configured by the API layer before calling LoadProgram.
	// The API server sets up EventWriter for WebSocket broadcasting.
	// If OutputWriter is still os.Stdout, set up a buffer to capture output for GetOutput().
	if s.vm.OutputWriter == os.Stdout {
		s.outputBuffer = &bytes.Buffer{}
		s.vm.OutputWriter = s.outputBuffer
	}

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
	s.atEntry = true

	return nil
}

// GetRegisterState returns current register state (thread-safe)
func (s *DebuggerService) GetRegisterState() RegisterState {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	defer s.mu.Unlock()
	if err := s.beginExecutionLocked(); err != nil {
		return err
	}
	defer s.endExecutionLocked()
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

// Pause stops execution, waits for it to end, and transitions to halted.
func (s *DebuggerService) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopExecutionLocked()
	s.vm.State = vm.StateHalted
}

// Reset performs a complete reset to initial state
// This clears the loaded program, all breakpoints, and resets the VM to pristine state
// Use ResetToEntryPoint() if you want to restart the current program instead
func (s *DebuggerService) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopExecutionLocked()

	// Full VM reset: clears all registers (PC=0), memory, and execution state
	s.vm.Reset()

	// Drop queued input, including bytes the VM has already buffered
	s.stdin = nil
	s.vm.SetStdinReader(stdinSource{s})

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
	s.atEntry = true

	return nil
}

// ResetToEntryPoint resets VM to program entry point without clearing the loaded program
// This is useful for restarting execution of the current program
func (s *DebuggerService) ResetToEntryPoint() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopExecutionLocked()

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
	s.atEntry = true

	return nil
}

// GetExecutionState returns current execution state
func (s *DebuggerService) GetExecutionState() ExecutionState {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sourceMapByAddr[address]
}

// GetSourceMap returns the source map entries with line numbers
func (s *DebuggerService) GetSourceMap() []SourceMapEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return copy of source map to prevent external modification
	result := make([]SourceMapEntry, len(s.sourceMap))
	copy(result, s.sourceMap)
	return result
}

// GetSourceMapByAddr returns address-to-line lookup (for debugger display)
func (s *DebuggerService) GetSourceMapByAddr() map[uint32]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return copy to prevent external modification
	result := make(map[uint32]string, len(s.sourceMapByAddr))
	for addr, line := range s.sourceMapByAddr {
		result[addr] = line
	}
	return result
}

// GetSymbols returns all symbols
func (s *DebuggerService) GetSymbols() map[string]uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return a copy to prevent external modification
	symbols := make(map[string]uint32, len(s.symbols))
	for k, v := range s.symbols {
		symbols[k] = v
	}
	return symbols
}

// GetSymbolForAddress resolves an address to a symbol name
func (s *DebuggerService) GetSymbolForAddress(addr uint32) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if there's a symbol at this address
	for name, symbolAddr := range s.symbols {
		if symbolAddr == addr {
			return name
		}
	}

	return ""
}

// RunUntilHalt runs the program until it halts, hits a breakpoint, or is stopped.
// It returns nil at once if Pause ran after SetRunning(true) but before this call.
func (s *DebuggerService) RunUntilHalt() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.debugger.Running {
		return nil
	}
	resuming := !s.atEntry
	if err := s.beginExecutionLocked(); err != nil {
		return err
	}
	defer s.endExecutionLocked()
	if resuming {
		s.debugger.ResumeFromCurrentPC()
	}
	return s.runLocked()
}

// runLocked executes until the debugger stops, the guest halts, or an error occurs.
// The caller holds s.mu and has claimed execution. The lock is released between
// instructions so queries and Pause can interleave.
func (s *DebuggerService) runLocked() error {
	s.vm.State = vm.StateRunning
	for steps := 1; s.debugger.Running && s.vm.State == vm.StateRunning; steps++ {
		// Single-step completes after one instruction; other modes check before it.
		if s.debugger.StepMode != debugger.StepSingle && s.breakLocked() {
			return nil
		}

		err := s.vm.Step()
		if errors.Is(err, vm.ErrInputInterrupted) {
			break
		}
		// A halt (exit) or SWI breakpoint also returns an error; the state says which.
		if err != nil && s.vm.State != vm.StateHalted && s.vm.State != vm.StateBreakpoint {
			s.debugger.Running = false
			return err
		}

		if s.debugger.StepMode == debugger.StepSingle && s.breakLocked() {
			return nil
		}

		s.mu.Unlock()
		if steps%stepsBeforeYield == 0 {
			time.Sleep(time.Millisecond)
		}
		s.mu.Lock()
	}

	s.debugger.Running = false
	// Still running means Pause, Reset or Close stopped it.
	if s.vm.State == vm.StateRunning {
		s.vm.State = vm.StateHalted
	}
	return nil
}

// breakLocked stops at a breakpoint or completed step when the debugger asks for it.
func (s *DebuggerService) breakLocked() bool {
	stop, reason := s.debugger.ShouldBreak()
	if !stop {
		return false
	}
	serviceLog.Printf("Stopped: %s at PC=0x%08X", reason, s.vm.CPU.PC)
	s.debugger.Running = false
	s.vm.State = vm.StateBreakpoint
	return true
}

// IsRunning returns whether execution is in progress
func (s *DebuggerService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.debugger.Running
}

// SetRunning sets the running state synchronously
// Used by async execution methods to set state before launching goroutines
func (s *DebuggerService) SetRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debugger.Running = running
	// The executing goroutine owns vm.State.
	if s.executing {
		return
	}
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
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vm.ExitCode
}

// GetOutput returns captured program output (clears buffer)
// Note: With the API server, output is broadcast via WebSocket events.
// This method is primarily used by tests and direct service usage (not via API).
func (s *DebuggerService) GetOutput() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.outputBuffer == nil {
		return ""
	}

	output := s.outputBuffer.String()
	s.outputBuffer.Reset()
	return output
}

// GetDisassembly returns disassembled instructions starting at address.
// Returns an empty slice if inputs are invalid or memory reads fail.
// Truncates the result if memory errors occur before count is reached.
//
// Parameters:
//   - startAddr: must be 4-byte aligned (ARM requirement)
//   - count: must be positive and <= maxDisassemblyCount
func (s *DebuggerService) GetDisassembly(startAddr uint32, count int) []DisassemblyLine {
	s.mu.Lock()
	defer s.mu.Unlock()

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

		// Get symbol at this address if any (lock already held)
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
	s.mu.Lock()
	defer s.mu.Unlock()

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
	if err := s.beginExecutionLocked(); err != nil {
		return err
	}
	defer s.endExecutionLocked()

	s.debugger.SetStepOver()
	return s.runLocked()
}

// StepOut executes until the current function returns
func (s *DebuggerService) StepOut() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil || s.program == nil {
		return fmt.Errorf("no program loaded")
	}

	if err := s.beginExecutionLocked(); err != nil {
		return err
	}
	defer s.endExecutionLocked()

	s.debugger.SetStepOut()
	return s.runLocked()
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
	s.mu.Lock()
	defer s.mu.Unlock()

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

// SendInput sends user input to the guest program's stdin.
// Input sent while nothing runs is queued as-is for the next run (batch pattern).
// Input sent during execution is echoed to the output and terminated with a newline.
func (s *DebuggerService) SendInput(input string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger.Running || s.executing {
		if s.vm.OutputWriter != nil {
			_, _ = s.vm.OutputWriter.Write([]byte(input + "\n"))
		}
		input += "\n"
	}
	s.stdin = append(s.stdin, input...)
	s.stdinReady.Broadcast()
	return nil
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
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vm.Statistics == nil {
		return nil, fmt.Errorf("statistics not enabled")
	}

	// Finalize statistics before returning
	s.vm.Statistics.Finalize()

	return s.vm.Statistics, nil
}
