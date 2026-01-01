package api

import (
	"time"

	"github.com/lookbusy1344/arm-emulator/service"
)

// SessionCreateRequest represents a request to create a new session
type SessionCreateRequest struct {
	MemorySize uint32 `json:"memorySize,omitempty"` // Memory size in bytes (default: 1MB)
	StackSize  uint32 `json:"stackSize,omitempty"`  // Stack size in bytes (default: 64KB)
	HeapSize   uint32 `json:"heapSize,omitempty"`   // Heap size in bytes (default: 256KB)
	FSRoot     string `json:"fsRoot,omitempty"`     // Filesystem root directory
}

// SessionCreateResponse represents the response from creating a session
type SessionCreateResponse struct {
	SessionID string    `json:"sessionId"`
	CreatedAt time.Time `json:"createdAt"`
}

// SessionStatusResponse represents the current status of a session
type SessionStatusResponse struct {
	SessionID string                 `json:"sessionId"`
	State     string                 `json:"state"`
	PC        uint32                 `json:"pc"`
	Cycles    uint64                 `json:"cycles"`
	Error     string                 `json:"error,omitempty"`
	HasWrite  bool                   `json:"hasWrite"`
	WriteAddr uint32                 `json:"writeAddr,omitempty"`
}

// LoadProgramRequest represents a request to load a program
type LoadProgramRequest struct {
	Source string `json:"source"` // Assembly source code
}

// LoadProgramResponse represents the response from loading a program
type LoadProgramResponse struct {
	Success bool              `json:"success"`
	Errors  []string          `json:"errors,omitempty"`
	Symbols map[string]uint32 `json:"symbols,omitempty"`
}

// RegistersResponse represents the current register state
type RegistersResponse struct {
	R0     uint32    `json:"r0"`
	R1     uint32    `json:"r1"`
	R2     uint32    `json:"r2"`
	R3     uint32    `json:"r3"`
	R4     uint32    `json:"r4"`
	R5     uint32    `json:"r5"`
	R6     uint32    `json:"r6"`
	R7     uint32    `json:"r7"`
	R8     uint32    `json:"r8"`
	R9     uint32    `json:"r9"`
	R10    uint32    `json:"r10"`
	R11    uint32    `json:"r11"`
	R12    uint32    `json:"r12"`
	SP     uint32    `json:"sp"`
	LR     uint32    `json:"lr"`
	PC     uint32    `json:"pc"`
	CPSR   CPSRFlags `json:"cpsr"`
	Cycles uint64    `json:"cycles"`
}

// CPSRFlags represents the CPSR flags
type CPSRFlags struct {
	N bool `json:"n"` // Negative
	Z bool `json:"z"` // Zero
	C bool `json:"c"` // Carry
	V bool `json:"v"` // Overflow
}

// MemoryRequest represents a request for memory data
type MemoryRequest struct {
	Address uint32 `json:"address"`
	Length  uint32 `json:"length"`
}

// MemoryResponse represents memory data
type MemoryResponse struct {
	Address uint32 `json:"address"`
	Data    []byte `json:"data"`
	Length  uint32 `json:"length"`
}

// DisassemblyRequest represents a request for disassembly
type DisassemblyRequest struct {
	Address uint32 `json:"address"`
	Count   uint32 `json:"count"`
}

// DisassemblyResponse represents disassembled instructions
type DisassemblyResponse struct {
	Instructions []InstructionInfo `json:"instructions"`
}

// InstructionInfo represents a disassembled instruction
type InstructionInfo struct {
	Address     uint32 `json:"address"`
	MachineCode uint32 `json:"machineCode"`
	Disassembly string `json:"disassembly"`
	Symbol      string `json:"symbol,omitempty"`
}

// BreakpointRequest represents a request to add/remove a breakpoint
type BreakpointRequest struct {
	Address uint32 `json:"address"`
}

// BreakpointsResponse represents a list of breakpoints
type BreakpointsResponse struct {
	Breakpoints []uint32 `json:"breakpoints"`
}

// StdinRequest represents a request to send stdin data
type StdinRequest struct {
	Data string `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse represents a simple success response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Event represents a WebSocket event
type Event struct {
	Type      string      `json:"type"`
	SessionID string      `json:"sessionId"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// StateEvent represents a state change event
type StateEvent struct {
	State     string        `json:"state"`
	PC        uint32        `json:"pc"`
	Registers [16]uint32    `json:"registers"`
	CPSR      CPSRFlags     `json:"cpsr"`
	Cycles    uint64        `json:"cycles"`
}

// OutputEvent represents console output
type OutputEvent struct {
	Stream  string `json:"stream"`  // "stdout" or "stderr"
	Content string `json:"content"` // Output content
}

// ExecutionEvent represents execution events like breakpoints
type ExecutionEvent struct {
	Event   string `json:"event"`            // "breakpoint_hit", "error", "halted"
	Address uint32 `json:"address,omitempty"`
	Symbol  string `json:"symbol,omitempty"`
	Message string `json:"message,omitempty"`
}

// ToRegisterResponse converts service.RegisterState to API response
func ToRegisterResponse(regs *service.RegisterState) *RegistersResponse {
	return &RegistersResponse{
		R0:     regs.Registers[0],
		R1:     regs.Registers[1],
		R2:     regs.Registers[2],
		R3:     regs.Registers[3],
		R4:     regs.Registers[4],
		R5:     regs.Registers[5],
		R6:     regs.Registers[6],
		R7:     regs.Registers[7],
		R8:     regs.Registers[8],
		R9:     regs.Registers[9],
		R10:    regs.Registers[10],
		R11:    regs.Registers[11],
		R12:    regs.Registers[12],
		SP:     regs.Registers[13],
		LR:     regs.Registers[14],
		PC:     regs.PC,
		CPSR: CPSRFlags{
			N: regs.CPSR.N,
			Z: regs.CPSR.Z,
			C: regs.CPSR.C,
			V: regs.CPSR.V,
		},
		Cycles: regs.Cycles,
	}
}

// ToInstructionInfo converts service.DisassemblyLine to API response
func ToInstructionInfo(line *service.DisassemblyLine) InstructionInfo {
	return InstructionInfo{
		Address:     line.Address,
		MachineCode: line.Opcode,
		Disassembly: line.Mnemonic,
		Symbol:      line.Symbol,
	}
}
