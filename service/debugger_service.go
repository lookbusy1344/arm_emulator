package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/lookbusy1344/arm-emulator/debugger"
	"github.com/lookbusy1344/arm-emulator/encoder"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

var serviceLog *log.Logger

func init() {
	// Check if debug logging is enabled via environment variable
	if os.Getenv("ARM_EMULATOR_DEBUG") != "" {
		// Create debug log file
		f, err := os.OpenFile("/tmp/arm-emulator-service-debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
type DebuggerService struct {
	mu                   sync.RWMutex
	vm                   *vm.VM
	debugger             *debugger.Debugger
	symbols              map[string]uint32
	sourceMap            map[uint32]string
	program              *parser.Program
	entryPoint           uint32
	outputWriter         *EventEmittingWriter
	ctx                  context.Context
	stateChangedCallback func() // Callback for GUI state updates
}

// NewDebuggerService creates a new debugger service
func NewDebuggerService(machine *vm.VM) *DebuggerService {
	return &DebuggerService{
		vm:        machine,
		debugger:  debugger.NewDebugger(machine),
		symbols:   make(map[string]uint32),
		sourceMap: make(map[uint32]string),
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

	// Build source map
	s.sourceMap = make(map[uint32]string)
	for _, inst := range program.Instructions {
		s.sourceMap[inst.Address] = inst.RawLine
	}
	for _, dir := range program.Directives {
		if dir.Name == ".word" || dir.Name == ".byte" || dir.Name == ".ascii" ||
			dir.Name == ".asciz" || dir.Name == ".space" {
			s.sourceMap[dir.Address] = "[DATA]" + dir.RawLine
		}
	}

	// Create output buffer with event emission
	outputBuffer := &bytes.Buffer{}
	s.outputWriter = NewEventEmittingWriter(outputBuffer, s.ctx)
	s.vm.OutputWriter = s.outputWriter

	// Load into debugger
	s.debugger.LoadSymbols(s.symbols)
	s.debugger.LoadSourceMap(s.sourceMap)

	// Load into VM memory (simplified version of main.go logic)
	return s.loadProgramIntoVM(program, entryPoint)
}

// processEscapeSequences processes escape sequences in strings
func processEscapeSequences(s string) string {
	result := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			// Process escape sequence
			switch s[i+1] {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '0':
				result = append(result, '\x00')
			case '"':
				result = append(result, '"')
			case '\'':
				result = append(result, '\'')
			case 'a':
				result = append(result, '\a')
			case 'b':
				result = append(result, '\b')
			case 'f':
				result = append(result, '\f')
			case 'v':
				result = append(result, '\v')
			default:
				// Unknown escape sequence, keep as is
				result = append(result, s[i])
				result = append(result, s[i+1])
			}
			i += 2
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}

// loadProgramIntoVM loads a parsed program into the VM's memory (matches main.go implementation)
func (s *DebuggerService) loadProgramIntoVM(program *parser.Program, entryPoint uint32) error {
	// Create low memory segment if needed
	if entryPoint < vm.CodeSegmentStart {
		segmentSize := uint32(vm.CodeSegmentStart)
		s.vm.Memory.AddSegment("low-memory", 0, segmentSize, vm.PermRead|vm.PermWrite|vm.PermExecute)
	}

	// Create encoder
	enc := encoder.NewEncoder(program.SymbolTable)
	enc.LiteralPoolLocs = append([]uint32(nil), program.LiteralPoolLocs...)
	enc.LiteralPoolCounts = append([]int(nil), program.LiteralPoolCounts...)

	// Build address map
	addressMap := make(map[*parser.Instruction]uint32)
	maxAddr := entryPoint
	for _, inst := range program.Instructions {
		addressMap[inst] = inst.Address
		instEnd := inst.Address + 4
		if instEnd > maxAddr {
			maxAddr = instEnd
		}
	}

	// Process data directives (CRITICAL: this was missing!)
	for _, directive := range program.Directives {
		dataAddr := directive.Address

		switch directive.Name {
		case ".org", ".align", ".balign", ".ltorg":
			// These are handled at parse time or during encoding
			continue

		case ".word":
			// Write 32-bit words
			for _, arg := range directive.Args {
				var value uint32
				if _, err := fmt.Sscanf(arg, "0x%x", &value); err != nil {
					if _, err := fmt.Sscanf(arg, "%d", &value); err != nil {
						// Try to look up as a symbol
						symValue, symErr := program.SymbolTable.Get(arg)
						if symErr != nil {
							return fmt.Errorf("invalid .word value %q: %w", arg, symErr)
						}
						value = symValue
					}
				}
				if err := s.vm.Memory.WriteWordUnsafe(dataAddr, value); err != nil {
					return fmt.Errorf(".word write error at 0x%08X: %w", dataAddr, err)
				}
				dataAddr += 4
			}
			if dataAddr > maxAddr {
				maxAddr = dataAddr
			}

		case ".byte":
			// Write bytes
			for _, arg := range directive.Args {
				var value uint32
				// Check for character literal
				if len(arg) >= 3 && arg[0] == '\'' && arg[len(arg)-1] == '\'' {
					if len(arg) == 3 {
						value = uint32(arg[1])
					} else {
						return fmt.Errorf("invalid .byte character literal: %s", arg)
					}
				} else if _, err := fmt.Sscanf(arg, "0x%x", &value); err != nil {
					if _, err := fmt.Sscanf(arg, "%d", &value); err != nil {
						return fmt.Errorf("invalid .byte value: %s", arg)
					}
				}
				if err := s.vm.Memory.WriteByteUnsafe(dataAddr, byte(value)); err != nil {
					return fmt.Errorf(".byte write error at 0x%08X: %w", dataAddr, err)
				}
				dataAddr++
			}
			if dataAddr > maxAddr {
				maxAddr = dataAddr
			}

		case ".ascii":
			// Write string without null terminator
			if len(directive.Args) > 0 {
				str := directive.Args[0]
				// Remove quotes
				if len(str) >= 2 && (str[0] == '"' || str[0] == '\'') {
					str = str[1 : len(str)-1]
				}
				// Process escape sequences
				processedStr := processEscapeSequences(str)
				// Write string bytes
				for i := 0; i < len(processedStr); i++ {
					if err := s.vm.Memory.WriteByteUnsafe(dataAddr, processedStr[i]); err != nil {
						return fmt.Errorf(".ascii write error at 0x%08X: %w", dataAddr, err)
					}
					dataAddr++
				}
			}
			if dataAddr > maxAddr {
				maxAddr = dataAddr
			}

		case ".asciz", ".string":
			// Write null-terminated string
			if len(directive.Args) > 0 {
				str := directive.Args[0]
				// Remove quotes
				if len(str) >= 2 && (str[0] == '"' || str[0] == '\'') {
					str = str[1 : len(str)-1]
				}
				// Process escape sequences
				processedStr := processEscapeSequences(str)
				// Write string bytes
				for i := 0; i < len(processedStr); i++ {
					if err := s.vm.Memory.WriteByteUnsafe(dataAddr, processedStr[i]); err != nil {
						return fmt.Errorf(".asciz write error at 0x%08X: %w", dataAddr, err)
					}
					dataAddr++
				}
				// Write null terminator
				if err := s.vm.Memory.WriteByteUnsafe(dataAddr, 0); err != nil {
					return fmt.Errorf(".asciz null terminator write error at 0x%08X: %w", dataAddr, err)
				}
				dataAddr++
			}
			if dataAddr > maxAddr {
				maxAddr = dataAddr
			}

		case ".space", ".skip":
			// Track space reservation
			if len(directive.Args) > 0 {
				var size uint32
				if _, err := fmt.Sscanf(directive.Args[0], "0x%x", &size); err != nil {
					_, _ = fmt.Sscanf(directive.Args[0], "%d", &size)
				}
				endAddr := dataAddr + size
				if endAddr > maxAddr {
					maxAddr = endAddr
				}
			}
		}
	}

	// Set literal pool start
	literalPoolStart := (maxAddr + 3) & ^uint32(3)
	enc.LiteralPoolStart = literalPoolStart

	// Encode and write instructions
	for _, inst := range program.Instructions {
		addr := addressMap[inst]
		opcode, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			return fmt.Errorf("encode error at 0x%08X: %w", addr, err)
		}
		if err := s.vm.Memory.WriteWordUnsafe(addr, opcode); err != nil {
			return fmt.Errorf("write error at 0x%08X: %w", addr, err)
		}
	}

	// Write literal pool
	for addr, value := range enc.LiteralPool {
		if err := s.vm.Memory.WriteWordUnsafe(addr, value); err != nil {
			return fmt.Errorf("literal write error at 0x%08X: %w", addr, err)
		}
	}

	// Set PC
	s.vm.CPU.PC = entryPoint
	s.vm.EntryPoint = entryPoint

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
	defer s.mu.Unlock()
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

// Pause pauses execution
func (s *DebuggerService) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debugger.Running = false
}

// Reset resets VM to entry point
func (s *DebuggerService) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.vm.CPU.PC = s.entryPoint
	s.vm.State = vm.StateHalted
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
			return nil, err
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
		HasWrite: s.vm.HasMemoryWrite,
	}
	serviceLog.Printf("GetLastMemoryWrite: address=0x%08X, hasWrite=%v", result.Address, result.HasWrite)
	s.vm.HasMemoryWrite = false
	return result
}

// GetSourceLine returns the source line for an address
func (s *DebuggerService) GetSourceLine(address uint32) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sourceMap[address]
}

// GetSourceMap returns the complete source map (address -> source line)
func (s *DebuggerService) GetSourceMap() map[uint32]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return copy of source map to prevent external modification
	sourceMap := make(map[uint32]string, len(s.sourceMap))
	for addr, line := range s.sourceMap {
		sourceMap[addr] = line
	}

	return sourceMap
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
func (s *DebuggerService) RunUntilHalt() error {
	serviceLog.Println("RunUntilHalt() called")
	s.mu.Lock()
	s.debugger.Running = true
	s.vm.State = vm.StateRunning
	s.mu.Unlock()

	stepCount := 0
	const stepsBeforeYield = 1000 // Yield every 1000 steps

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

		// Execute step (while holding lock)
		pc := s.vm.CPU.PC
		err := s.vm.Step()
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
//   - count: must be positive and <= 1000
func (s *DebuggerService) GetDisassembly(startAddr uint32, count int) []DisassemblyLine {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate inputs
	if count <= 0 || count > 1000 {
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

		// Get symbol at this address if any
		symbol := s.GetSymbolForAddress(addr)

		// Get mnemonic from source map if available
		mnemonic := ""
		if sourceLine, ok := s.sourceMap[addr]; ok {
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
//     Must be in range [-100000, 100000] to prevent wraparound attacks.
//   - count: number of stack entries to read. Must be positive and <= 1000.
//
// The function performs safe arithmetic with overflow detection to prevent
// integer wraparound vulnerabilities.
func (s *DebuggerService) GetStack(offset int, count int) []StackEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate inputs
	if count <= 0 || count > 1000 {
		return []StackEntry{}
	}

	// Validate offset to prevent wraparound attacks
	if offset < -100000 || offset > 100000 {
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

	// Use debugger's public method instead of accessing fields directly
	s.debugger.SetStepOver()

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
