package vm

// CPU represents the ARM2 processor state
type CPU struct {
	// General purpose registers R0-R14
	R [15]uint32

	// Program Counter (R15)
	PC uint32

	// Current Program Status Register
	CPSR CPSR

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
		result |= 1 << 31 // N flag in bit 31
	}
	if c.Z {
		result |= 1 << 30 // Z flag in bit 30
	}
	if c.C {
		result |= 1 << 29 // C flag in bit 29
	}
	if c.V {
		result |= 1 << 28 // V flag in bit 28
	}
	// Bits 27-0 are reserved/unused in basic ARM2 CPSR
	return result
}

// FromUint32 sets CPSR flags from a 32-bit value
// ARM CPSR format: NZCV flags are in bits 31-28
func (c *CPSR) FromUint32(value uint32) {
	c.N = (value & (1 << 31)) != 0 // N flag in bit 31
	c.Z = (value & (1 << 30)) != 0 // Z flag in bit 30
	c.C = (value & (1 << 29)) != 0 // C flag in bit 29
	c.V = (value & (1 << 28)) != 0 // V flag in bit 28
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
		R:      [15]uint32{},
		PC:     0,
		CPSR:   CPSR{},
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
	c.Cycles = 0
}

// GetSP returns the stack pointer value
func (c *CPU) GetSP() uint32 {
	return c.R[SP]
}

// SetSP sets the stack pointer value
func (c *CPU) SetSP(value uint32) {
	c.R[SP] = value
}

// SetSPWithTrace sets the stack pointer value and records it for stack tracing
func (c *CPU) SetSPWithTrace(vm *VM, value uint32, pc uint32) {
	oldSP := c.R[SP]
	c.R[SP] = value

	// Record stack trace if enabled
	if vm.StackTrace != nil {
		vm.StackTrace.RecordSPMove(vm.CPU.Cycles, pc, oldSP, value)
	}
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
	if reg == 15 {
		return c.PC + 8
	}
	if reg < 0 || reg > 14 {
		return 0
	}
	return c.R[reg]
}

// SetRegister sets the value of a register (R0-R14 or PC)
func (c *CPU) SetRegister(reg int, value uint32) {
	if reg == 15 {
		c.PC = value
	} else if reg >= 0 && reg <= 14 {
		c.R[reg] = value
	}
}

// IncrementPC increments the program counter by 4 (one instruction)
func (c *CPU) IncrementPC() {
	c.PC += 4
}

// Branch sets the program counter to a new address
func (c *CPU) Branch(address uint32) {
	c.PC = address
}

// BranchWithLink saves the return address in LR and branches
func (c *CPU) BranchWithLink(address uint32) {
	c.SetLR(c.PC + 4) // Save return address
	c.PC = address
}

// IncrementCycles increments the cycle counter
func (c *CPU) IncrementCycles(cycles uint64) {
	c.Cycles += cycles
}
