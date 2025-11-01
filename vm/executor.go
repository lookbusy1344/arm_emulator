package vm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

// ExecutionMode represents the execution mode of the VM
type ExecutionMode int

const (
	ModeRun      ExecutionMode = iota // Run until halt or breakpoint
	ModeStep                          // Execute single instruction
	ModeStepOver                      // Execute until next instruction at same call level
	ModeStepInto                      // Execute single instruction, following branches
)

// ExecutionState represents the current state of execution
type ExecutionState int

const (
	StateRunning ExecutionState = iota
	StateHalted
	StateBreakpoint
	StateError
)

// Instruction represents a decoded ARM instruction
type Instruction struct {
	Address   uint32
	Opcode    uint32
	Type      InstructionType
	Condition ConditionCode
	SetFlags  bool // S bit
	// Operands will be added as we implement instructions
}

// InstructionType represents the type of instruction
type InstructionType int

const (
	InstUnknown InstructionType = iota
	InstDataProcessing
	InstMultiply
	InstLoadStore
	InstLoadStoreMultiple
	InstBranch
	InstSWI
	InstPSRTransfer
)

// VM represents the complete virtual machine
type VM struct {
	CPU    *CPU
	Memory *Memory
	State  ExecutionState
	Mode   ExecutionMode

	// Execution limits and statistics
	MaxCycles      uint64
	CycleLimit     uint64
	InstructionLog []uint32 // History of executed instruction addresses

	// Error handling
	LastError error

	// Runtime environment
	EntryPoint       uint32
	StackTop         uint32 // Initial stack pointer value for reset
	ProgramArguments []string
	ExitCode         int32

	// I/O redirection (for TUI and testing)
	OutputWriter io.Writer // Writer for program output (defaults to os.Stdout)

	// Tracing and statistics (Phase 10)
	ExecutionTrace *ExecutionTrace
	MemoryTrace    *MemoryTrace
	Statistics     *PerformanceStatistics

	// Additional diagnostic modes (Phase 11)
	CodeCoverage  *CodeCoverage
	StackTrace    *StackTrace
	FlagTrace     *FlagTrace
	RegisterTrace *RegisterTrace

	// File descriptor table (simple)
	files []*os.File
	fdMu  sync.Mutex

	// Per-instance stdin reader to avoid race conditions when multiple VMs
	// run concurrently. Previously this was a global variable shared across
	// all VM instances, causing data corruption during parallel execution.
	stdinReader *bufio.Reader

	// Last memory write address for GUI highlighting
	LastMemoryWrite uint32
	HasMemoryWrite  bool
}

// NewVM creates a new virtual machine instance
func NewVM() *VM {
	return &VM{
		CPU:              NewCPU(),
		Memory:           NewMemory(),
		State:            StateHalted,
		Mode:             ModeRun,
		MaxCycles:        DefaultMaxCycles, // Default 1M instruction limit
		CycleLimit:       0,
		InstructionLog:   make([]uint32, 0, DefaultLogCapacity),
		EntryPoint:       CodeSegmentStart,
		ProgramArguments: make([]string, 0),
		ExitCode:         0,
		OutputWriter:     os.Stdout,                            // Default to stdout
		files:            make([]*os.File, DefaultFDTableSize), // Will be lazily initialized to stdin/stdout/stderr
		stdinReader:      bufio.NewReader(os.Stdin),            // Per-instance stdin reader
	}
}

// Reset resets the VM to initial state
func (vm *VM) Reset() {
	vm.CPU.Reset()
	vm.Memory.Reset()
	vm.State = StateHalted
	vm.InstructionLog = vm.InstructionLog[:0]
	vm.LastError = nil
}

// ResetRegisters resets only CPU registers and state, preserving memory contents
// This is useful for debugger operations that need to restart execution without
// losing the loaded program
func (vm *VM) ResetRegisters() {
	vm.CPU.Reset()
	// Restore PC to entry point after reset
	vm.CPU.PC = vm.EntryPoint
	// Restore stack pointer to initial value
	if vm.StackTop != 0 {
		vm.CPU.SetSP(vm.StackTop)
	}
	vm.State = StateHalted
	vm.InstructionLog = vm.InstructionLog[:0]
	vm.LastError = nil
}

// LoadProgram loads program bytes into code memory
func (vm *VM) LoadProgram(data []byte, startAddress uint32) error {
	if err := vm.Memory.LoadBytes(startAddress, data); err != nil {
		return fmt.Errorf("failed to load program: %w", err)
	}

	vm.CPU.PC = startAddress
	vm.State = StateHalted
	return nil
}

// SetEntryPoint sets the program counter to the entry point
func (vm *VM) SetEntryPoint(address uint32) {
	vm.CPU.PC = address
}

// InitializeStack initializes the stack pointer
func (vm *VM) InitializeStack(stackTop uint32) {
	vm.StackTop = stackTop
	vm.CPU.SetSP(stackTop)
}

// Step executes a single instruction
func (vm *VM) Step() error {
	if vm.State == StateError {
		return fmt.Errorf("VM is in error state: %w", vm.LastError)
	}

	// Check cycle limit
	if vm.CycleLimit > 0 && vm.CPU.Cycles >= vm.CycleLimit {
		vm.State = StateError
		vm.LastError = fmt.Errorf("cycle limit exceeded (%d cycles)", vm.CycleLimit)
		return vm.LastError
	}

	// Check execute permission for current PC
	if err := vm.Memory.CheckExecutePermission(vm.CPU.PC); err != nil {
		vm.State = StateError
		vm.LastError = err
		return err
	}

	// Log instruction address
	vm.InstructionLog = append(vm.InstructionLog, vm.CPU.PC)

	// Fetch instruction
	instruction, err := vm.Fetch()
	if err != nil {
		vm.State = StateError
		vm.LastError = fmt.Errorf("fetch failed at PC=0x%08X: %w", vm.CPU.PC, err)
		return vm.LastError
	}

	// Decode instruction
	decoded, err := vm.Decode(instruction)
	if err != nil {
		vm.State = StateError
		vm.LastError = fmt.Errorf("decode failed at PC=0x%08X: %w", vm.CPU.PC, err)
		return vm.LastError
	}

	// Check condition code
	condResult := vm.CPU.CPSR.EvaluateCondition(decoded.Condition)

	if !condResult {
		// Condition not met, skip instruction
		vm.CPU.IncrementPC()
		vm.CPU.IncrementCycles(1)
		return nil
	}

	// Snapshot registers before execution for register trace
	var regsBefore [16]uint32
	if vm.RegisterTrace != nil && vm.RegisterTrace.Enabled {
		// R0-R14
		copy(regsBefore[:15], vm.CPU.R[:])
		// PC (R15)
		regsBefore[15] = vm.CPU.PC
	}

	// Execute instruction
	if err := vm.Execute(decoded); err != nil {
		// Don't overwrite terminal states (Halted, Breakpoint) set by syscalls
		if vm.State != StateHalted && vm.State != StateBreakpoint {
			vm.State = StateError
			vm.LastError = fmt.Errorf("execute failed at PC=0x%08X: %w", decoded.Address, err)
		}
		return err
	}

	vm.CPU.IncrementCycles(1)

	// Record diagnostic information after instruction execution
	currentPC := decoded.Address

	// Code coverage tracking
	if vm.CodeCoverage != nil {
		vm.CodeCoverage.RecordExecution(currentPC, vm.CPU.Cycles)
	}

	// Flag change tracking
	if vm.FlagTrace != nil {
		// Get simple instruction name for trace (we'll enhance this later with proper disassembly)
		instName := fmt.Sprintf("0x%08X", decoded.Opcode)
		vm.FlagTrace.RecordFlags(vm.CPU.Cycles, currentPC, instName, vm.CPU.CPSR)
	}

	// Register change tracking
	if vm.RegisterTrace != nil && vm.RegisterTrace.Enabled {
		// Check each register for changes
		for i := 0; i < 15; i++ {
			if vm.CPU.R[i] != regsBefore[i] {
				vm.RegisterTrace.RecordWrite(vm.CPU.Cycles, currentPC, getRegisterName(i), regsBefore[i], vm.CPU.R[i])
			}
		}
		// Check PC (R15)
		if vm.CPU.PC != regsBefore[15] {
			vm.RegisterTrace.RecordWrite(vm.CPU.Cycles, currentPC, "PC", regsBefore[15], vm.CPU.PC)
		}
	}

	return nil
}

// Fetch fetches the instruction at the current PC
func (vm *VM) Fetch() (uint32, error) {
	instruction, err := vm.Memory.ReadWord(vm.CPU.PC)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch instruction: %w", err)
	}
	return instruction, nil
}

// Decode decodes a raw instruction word
func (vm *VM) Decode(opcode uint32) (*Instruction, error) {
	inst := &Instruction{
		Address:   vm.CPU.PC,
		Opcode:    opcode,
		Condition: ConditionCode((opcode >> 28) & 0xF),
		SetFlags:  (opcode & (1 << 20)) != 0, // S bit
	}

	// Determine instruction type based on bits 27-26
	bits2726 := (opcode >> 26) & 0x3

	switch bits2726 {
	case 0: // 00 - Could be data processing, multiply, BX, BLX, or load/store halfword
		// Check for BX (Branch and Exchange) first: bits [27:4] = 0x12FFF1
		if (opcode & 0x0FFFFFF0) == 0x012FFF10 {
			inst.Type = InstBranch
		} else if (opcode & 0x0FFFFFF0) == 0x012FFF30 {
			// BLX register form: bits [27:4] = 0x12FFF3
			inst.Type = InstBranch
		} else if (opcode & 0x0FC000F0) == 0x00000090 {
			// Multiply instruction pattern (MUL, MLA)
			inst.Type = InstMultiply
		} else if (opcode & 0x0F8000F0) == 0x00800090 {
			// Long multiply instruction pattern (UMULL, UMLAL, SMULL, SMLAL)
			// Bits [27:23] = 0b00001, bits [7:4] = 0b1001
			inst.Type = InstMultiply
		} else if (opcode & 0x0FBF0FFF) == 0x010F0000 {
			// MRS instruction: bits [27:23]=00010, [22]=PSR, [21]=0, [20]=0, [19:16]=1111, [11:0]=0
			// Pattern: cccc 00010 x 00 1111 dddd 0000 0000 0000
			inst.Type = InstPSRTransfer
		} else if (opcode & 0x0FB000F0) == 0x01200000 {
			// MSR instruction (register): bits [27:23]=00010, [21]=1, [20]=0, [7:4]=0000
			// Pattern: cccc 00010 x 10 xxxx 1111 0000 0000 mmmm
			inst.Type = InstPSRTransfer
		} else if (opcode & 0x0FB00000) == 0x03200000 {
			// MSR instruction (immediate): bits [27:23]=00110, [21]=1, [20]=0
			// Pattern: cccc 00110 x 10 xxxx 1111 rrrr iiii iiii
			inst.Type = InstPSRTransfer
		} else {
			// Check for halfword load/store: bit 25 = 0, bit 7 = 1, bit 4 = 1
			// This distinguishes from data processing with immediate (bit 25 = 1)
			bit25 := (opcode >> 25) & 1
			bit7 := (opcode >> 7) & 1
			bit4 := (opcode >> 4) & 1
			if bit25 == 0 && bit7 == 1 && bit4 == 1 {
				// This is a halfword/signed transfer (LDRH, STRH, LDRSB, LDRSH)
				inst.Type = InstLoadStore
			} else {
				// Data processing
				inst.Type = InstDataProcessing
			}
		}

	case 1: // 01 - Load/Store
		inst.Type = InstLoadStore

	case 2: // 10 - Could be branch or load/store multiple
		if (opcode & 0x02000000) != 0 {
			// Branch
			inst.Type = InstBranch
		} else {
			// Load/Store Multiple
			inst.Type = InstLoadStoreMultiple
		}

	case 3: // 11 - Coprocessor or SWI
		if (opcode & 0x0F000000) == 0x0F000000 {
			// SWI
			inst.Type = InstSWI
		} else {
			return nil, fmt.Errorf("coprocessor instructions not supported")
		}
	}

	return inst, nil
}

// Execute executes a decoded instruction
func (vm *VM) Execute(inst *Instruction) error {
	switch inst.Type {
	case InstDataProcessing:
		return ExecuteDataProcessing(vm, inst)
	case InstMultiply:
		return ExecuteMultiply(vm, inst)
	case InstLoadStore:
		return ExecuteLoadStore(vm, inst)
	case InstLoadStoreMultiple:
		return ExecuteLoadStoreMultiple(vm, inst)
	case InstBranch:
		return ExecuteBranch(vm, inst)
	case InstSWI:
		return ExecuteSWI(vm, inst)
	case InstPSRTransfer:
		return ExecutePSRTransfer(vm, inst)
	default:
		return fmt.Errorf("unknown instruction type at 0x%08X: opcode=0x%08X", inst.Address, inst.Opcode)
	}
}

// Instruction implementations are in separate files:
// - data_processing.go
// - multiply.go
// - inst_memory.go
// - memory_multi.go
// - branch.go
// - syscall.go
// - psr.go

// Run executes instructions until halt, error, or breakpoint
func (vm *VM) Run() error {
	vm.State = StateRunning

	for vm.State == StateRunning {
		if err := vm.Step(); err != nil {
			return err
		}

		// Check for halt conditions
		// This is a placeholder - will be enhanced with proper halt detection
		if vm.CPU.Cycles > vm.MaxCycles {
			vm.State = StateHalted
			return fmt.Errorf("maximum cycles exceeded")
		}
	}

	return nil
}

// GetState returns the current execution state
func (vm *VM) GetState() ExecutionState {
	return vm.State
}

// SetState sets the execution state
func (vm *VM) SetState(state ExecutionState) {
	vm.State = state
}

// GetInstructionHistory returns the history of executed instruction addresses
func (vm *VM) GetInstructionHistory() []uint32 {
	return vm.InstructionLog
}

// DumpState returns a string representation of the VM state for debugging
func (vm *VM) DumpState() string {
	return fmt.Sprintf(
		"PC=0x%08X SP=0x%08X LR=0x%08X CPSR=[%s%s%s%s] Cycles=%d State=%v",
		vm.CPU.PC,
		vm.CPU.GetSP(),
		vm.CPU.GetLR(),
		map[bool]string{true: "N", false: "-"}[vm.CPU.CPSR.N],
		map[bool]string{true: "Z", false: "-"}[vm.CPU.CPSR.Z],
		map[bool]string{true: "C", false: "-"}[vm.CPU.CPSR.C],
		map[bool]string{true: "V", false: "-"}[vm.CPU.CPSR.V],
		vm.CPU.Cycles,
		vm.State,
	)
}

// Bootstrap initializes the VM runtime environment
func (vm *VM) Bootstrap(args []string) error {
	// Store program arguments
	vm.ProgramArguments = args

	// Initialize stack pointer to top of stack
	stackTop := uint32(StackSegmentStart + StackSegmentSize)
	vm.InitializeStack(stackTop)

	// Set link register to a halt address (so returning from main halts)
	vm.CPU.SetLR(0xFFFFFFFF)

	// Set program counter to entry point
	vm.CPU.PC = vm.EntryPoint

	// Initialize state
	vm.State = StateHalted
	vm.ExitCode = 0

	return nil
}

// FindEntryPoint searches for common entry point labels in symbol table
// Common entry points: _start, main, __start
func (vm *VM) FindEntryPoint(symbols map[string]uint32) (uint32, error) {
	// Try common entry point names in order of preference
	entryPoints := []string{"_start", "main", "__start", "start"}

	for _, name := range entryPoints {
		if addr, exists := symbols[name]; exists {
			vm.EntryPoint = addr
			return addr, nil
		}
	}

	// If no entry point found, default to code segment start
	vm.EntryPoint = CodeSegmentStart
	return CodeSegmentStart, fmt.Errorf("no entry point found, using default 0x%08X", CodeSegmentStart)
}

// SetProgramArguments sets command-line arguments for the program
func (vm *VM) SetProgramArguments(args []string) {
	vm.ProgramArguments = args
}

// GetExitCode returns the program exit code
func (vm *VM) GetExitCode() int32 {
	return vm.ExitCode
}
