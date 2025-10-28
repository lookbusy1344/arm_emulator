package service

import "github.com/lookbusy1344/arm-emulator/vm"

// RegisterState represents a snapshot of CPU registers
type RegisterState struct {
	Registers [16]uint32
	CPSR      CPSRState
	PC        uint32
	Cycles    uint64
}

// CPSRState represents CPSR flags for serialization
type CPSRState struct {
	N bool // Negative
	Z bool // Zero
	C bool // Carry
	V bool // Overflow
}

// BreakpointInfo represents a breakpoint for UI display
type BreakpointInfo struct {
	Address uint32
	Enabled bool
}

// WatchpointInfo represents a watchpoint for UI display
type WatchpointInfo struct {
	Address uint32
	Type    string // "read", "write", "readwrite"
	Enabled bool
}

// MemoryRegion represents a contiguous memory region
type MemoryRegion struct {
	Address uint32
	Data    []byte
	Size    uint32
}

// ExecutionState represents the current state of execution
type ExecutionState string

const (
	StateRunning    ExecutionState = "running"
	StateHalted     ExecutionState = "halted"
	StateBreakpoint ExecutionState = "breakpoint"
	StateError      ExecutionState = "error"
)

// VMStateToExecution converts vm.ExecutionState to service.ExecutionState
func VMStateToExecution(state vm.ExecutionState) ExecutionState {
	switch state {
	case vm.StateRunning:
		return StateRunning
	case vm.StateHalted:
		return StateHalted
	case vm.StateBreakpoint:
		return StateBreakpoint
	case vm.StateError:
		return StateError
	default:
		return StateHalted
	}
}
