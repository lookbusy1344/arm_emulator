package vm

// CPU represents the ARM2 processor state
type CPU struct {
	// General purpose registers R0-R14
	R [ARMGeneralRegisterCount]uint32

	// Program Counter (R15)
	PC uint32

	// Current Program Status Register
	CPSR CPSR

	// Saved Program Status Register (for exception handling)
	// Used when LDM with S bit restores PC (simulates exception return)
	SPSR CPSR

	// Cycle counter for statistics
	Cycles uint64
}

// CPSR represents the Current Program Status Register with condition flags
type CPSR struct {
	N bool // Negative flag (bit 31 of result)
	Z bool // Zero flag (result == 0)
	C bool // Carry flag (unsigned overflow for arithmetic, last bit shifted out for shifts)
	V bool // Overflow flag (signed overflow)
}

// ToUint32 converts CPSR flags to a 32-bit value
// ARM CPSR format: NZCV flags are in bits 31-28
func (c *CPSR) ToUint32() uint32 {
	var result uint32
	if c.N {
		result |= 1 << CPSRBitN // N flag in bit 31
	}
	if c.Z {
		result |= 1 << CPSRBitZ // Z flag in bit 30
	}
	if c.C {
		result |= 1 << CPSRBitC // C flag in bit 29
	}
	if c.V {
		result |= 1 << CPSRBitV // V flag in bit 28
	}
	// Bits 27-0 are reserved/unused in basic ARM2 CPSR
	return result
}

// FromUint32 sets CPSR flags from a 32-bit value
// ARM CPSR format: NZCV flags are in bits 31-28
func (c *CPSR) FromUint32(value uint32) {
	c.N = (value & (1 << CPSRBitN)) != 0 // N flag in bit 31
	c.Z = (value & (1 << CPSRBitZ)) != 0 // Z flag in bit 30
	c.C = (value & (1 << CPSRBitC)) != 0 // C flag in bit 29
	c.V = (value & (1 << CPSRBitV)) != 0 // V flag in bit 28
	// Bits 27-0 are ignored (reserved/unused in basic ARM2)
}

// Register aliases for convenience
const (
	R0  = 0
	R1  = 1
	R2  = 2
	R3  = 3
	R4  = 4
	R5  = 5
	R6  = 6
	R7  = 7
	R8  = 8
	R9  = 9
	R10 = 10
	R11 = 11
	R12 = 12
	SP  = 13 // Stack Pointer
	LR  = 14 // Link Register
	// PC is stored separately as a field
)

// NewCPU creates and initializes a new CPU instance
func NewCPU() *CPU {
	return &CPU{
		R:      [ARMGeneralRegisterCount]uint32{},
		PC:     0,
		CPSR:   CPSR{},
		SPSR:   CPSR{},
		Cycles: 0,
	}
}

// Reset resets the CPU to initial state
func (c *CPU) Reset() {
	for i := range c.R {
		c.R[i] = 0
	}
	c.PC = 0
	c.CPSR = CPSR{}
	c.SPSR = CPSR{}
	c.Cycles = 0
}

// GetSP returns the stack pointer value
func (c *CPU) GetSP() uint32 {
	return c.R[SP]
}

// SetSP sets the stack pointer (R13).
// Note: Like real ARM hardware, this function does not validate bounds.
// SP can be set to any value, and actual memory protection occurs when
// memory is accessed. This allows advanced use cases like cooperative
// multitasking with multiple stacks (see examples/task_scheduler.s).
func (c *CPU) SetSP(value uint32) error {
	c.R[SP] = value
	return nil
}

// SetSPWithTrace sets the stack pointer with tracing support.
// Note: Like real ARM hardware, this function does not validate bounds.
// Stack trace monitoring (when enabled) will detect overflow/underflow conditions.
// Memory protection occurs when memory is accessed, allowing advanced use cases
// like cooperative multitasking with multiple stacks.
func (c *CPU) SetSPWithTrace(vm *VM, value uint32, pc uint32) error {
	oldSP := c.R[SP]
	c.R[SP] = value

	// Record stack trace if enabled
	if vm.StackTrace != nil {
		vm.StackTrace.RecordSPMove(vm.CPU.Cycles, pc, oldSP, value)
	}

	return nil
}

// GetLR returns the link register value
func (c *CPU) GetLR() uint32 {
	return c.R[LR]
}

// SetLR sets the link register value
func (c *CPU) SetLR(value uint32) {
	c.R[LR] = value
}

// GetRegister returns the value of a register (R0-R14 or PC)
// When reading R15 (PC), returns PC+8 to simulate ARM pipeline effect
func (c *CPU) GetRegister(reg int) uint32 {
	if reg == ARMRegisterPC {
		return c.PC + ARMPipelineOffset
	}
	if reg < 0 || reg >= ARMGeneralRegisterCount {
		return 0
	}
	return c.R[reg]
}

// SetRegister sets the value of a register (R0-R14 or PC)
func (c *CPU) SetRegister(reg int, value uint32) {
	if reg == ARMRegisterPC {
		c.PC = value
	} else if reg >= 0 && reg < ARMGeneralRegisterCount {
		c.R[reg] = value
	}
}

// IncrementPC increments the program counter by 4 (one instruction)
func (c *CPU) IncrementPC() {
	c.PC += ARMInstructionSize
}

// Branch sets the program counter to a new address
func (c *CPU) Branch(address uint32) {
	c.PC = address
}

// BranchWithLink saves the return address in LR and branches
func (c *CPU) BranchWithLink(address uint32) {
	c.SetLR(c.PC + ARMInstructionSize) // Save return address
	c.PC = address
}

// IncrementCycles increments the cycle counter
func (c *CPU) IncrementCycles(cycles uint64) {
	c.Cycles += cycles
}

// GetRegisterWithTrace gets a register value and records the read
func (c *CPU) GetRegisterWithTrace(vm *VM, reg int, pc uint32) uint32 {
	value := c.GetRegister(reg)

	// Record register read if tracing enabled
	if vm.RegisterTrace != nil && vm.RegisterTrace.Enabled {
		regName := getRegisterName(reg)
		vm.RegisterTrace.RecordRead(c.Cycles, pc, regName, value)
	}

	return value
}

// SetRegisterWithTrace sets a register value and records the write
func (c *CPU) SetRegisterWithTrace(vm *VM, reg int, value uint32, pc uint32) {
	oldValue := c.GetRegister(reg)

	// Set the register
	c.SetRegister(reg, value)

	// Record register write if tracing enabled
	if vm.RegisterTrace != nil && vm.RegisterTrace.Enabled {
		regName := getRegisterName(reg)
		vm.RegisterTrace.RecordWrite(c.Cycles, pc, regName, oldValue, value)
	}
}

// getRegisterName returns the name of a register
func getRegisterName(reg int) string {
	switch reg {
	case R0:
		return "R0"
	case R1:
		return "R1"
	case R2:
		return "R2"
	case R3:
		return "R3"
	case R4:
		return "R4"
	case R5:
		return "R5"
	case R6:
		return "R6"
	case R7:
		return "R7"
	case R8:
		return "R8"
	case R9:
		return "R9"
	case R10:
		return "R10"
	case R11:
		return "R11"
	case R12:
		return "R12"
	case SP:
		return "SP"
	case LR:
		return "LR"
	case ARMRegisterPC:
		return "PC"
	default:
		return "UNKNOWN"
	}
}

// SaveCPSR copies the current CPSR to SPSR
// This is typically done when entering an exception handler
func (c *CPU) SaveCPSR() {
	c.SPSR = c.CPSR
}

// RestoreCPSR copies SPSR back to CPSR
// This is done when returning from an exception handler
// (e.g., LDM with S bit loading PC)
func (c *CPU) RestoreCPSR() {
	c.CPSR = c.SPSR
}
