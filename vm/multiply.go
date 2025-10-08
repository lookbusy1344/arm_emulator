package vm

import (
	"fmt"
)

// ExecuteMultiply executes multiply instructions (MUL, MLA)
func ExecuteMultiply(vm *VM, inst *Instruction) error {
	accumulate := (inst.Opcode >> 21) & 0x1 // A bit: 1=MLA, 0=MUL
	setFlags := inst.SetFlags               // S bit

	rd := int((inst.Opcode >> 16) & 0xF) // Destination register
	rn := int((inst.Opcode >> 12) & 0xF) // Accumulate register (for MLA)
	rs := int((inst.Opcode >> 8) & 0xF)  // Operand register 1
	rm := int(inst.Opcode & 0xF)         // Operand register 2

	// Validate: Rd and Rm must be different registers (ARM2 restriction)
	if rd == rm {
		return fmt.Errorf("multiply: Rd and Rm must be different registers (Rd=%d, Rm=%d)", rd, rm)
	}

	// Validate: R15 (PC) cannot be used
	if rd == 15 || rm == 15 || rs == 15 || (accumulate == 1 && rn == 15) {
		return fmt.Errorf("multiply: R15 (PC) cannot be used in multiply instructions")
	}

	// Get operands
	op1 := vm.CPU.GetRegister(rm)
	op2 := vm.CPU.GetRegister(rs)

	// Perform multiplication (lower 32 bits only)
	result := op1 * op2

	// Add accumulator if MLA
	if accumulate == 1 {
		result += vm.CPU.GetRegister(rn)
	}

	// Store result
	vm.CPU.SetRegister(rd, result)

	// Update flags if requested
	if setFlags {
		// Multiply only updates N and Z flags
		// C flag is meaningless, V flag is unaffected
		vm.CPU.CPSR.UpdateFlagsNZ(result)
	}

	// Increment PC
	vm.CPU.IncrementPC()

	// Increment cycle count (multiply takes variable cycles: 2-16)
	// For simplicity, we use a fixed count based on the value
	cycles := calculateMultiplyCycles(op2)
	vm.CPU.IncrementCycles(uint64(cycles - 1)) // -1 because Step() already adds 1

	return nil
}

// calculateMultiplyCycles calculates the number of cycles for a multiply operation
// ARM2 multiply timing varies from 2 to 16 cycles based on the multiplier value
func calculateMultiplyCycles(multiplier uint32) int {
	// Count the number of significant bits in the multiplier
	// Each group of 2 bits adds a cycle
	cycles := 2 // Base cycles

	value := multiplier
	for i := 0; i < 16; i++ {
		if value&0x3 != 0 {
			cycles++
		}
		value >>= 2
		if value == 0 {
			break
		}
	}

	if cycles > 16 {
		cycles = 16
	}

	return cycles
}
