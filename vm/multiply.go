package vm

import (
	"fmt"
)

// ExecuteMultiply executes multiply instructions (MUL, MLA, UMULL, UMLAL, SMULL, SMLAL)
func ExecuteMultiply(vm *VM, inst *Instruction) error {
	// Check if this is a long multiply instruction
	// Long multiply: bits [27:23] = 0b00001
	if ((inst.Opcode >> Bits27_23Shift) & LongMultiplyMask5) == 1 {
		return ExecuteMultiplyLong(vm, inst)
	}

	// Standard multiply (MUL, MLA)
	accumulate := (inst.Opcode >> MultiplyAShift) & Mask1Bit // A bit: 1=MLA, 0=MUL
	setFlags := inst.SetFlags                                // S bit

	rd := int((inst.Opcode >> RnShift) & Mask4Bit) // Destination register
	rn := int((inst.Opcode >> RdShift) & Mask4Bit) // Accumulate register (for MLA)
	rs := int((inst.Opcode >> RsShift) & Mask4Bit) // Operand register 1
	rm := int(inst.Opcode & Mask4Bit)              // Operand register 2

	// Validate: Rd and Rm must be different registers (ARM2 restriction)
	if rd == rm {
		return fmt.Errorf("multiply: Rd and Rm must be different registers (Rd=%d, Rm=%d)", rd, rm)
	}

	// Validate: R15 (PC) cannot be used
	if rd == PCRegister || rm == PCRegister || rs == PCRegister || (accumulate == 1 && rn == PCRegister) {
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
	// Safe: cycles is in range [2, 16], well within uint64 range
	vm.CPU.IncrementCycles(uint64(cycles - 1)) // #nosec G115 -- cycles is 2-16, -1 because Step() already adds 1

	return nil
}

// calculateMultiplyCycles calculates the number of cycles for a multiply operation
// ARM2 multiply timing varies from 2 to 16 cycles based on the multiplier value
func calculateMultiplyCycles(multiplier uint32) int {
	// Count the number of significant bits in the multiplier
	// Each group of 2 bits adds a cycle
	cycles := MultiplyBaseCycles

	value := multiplier
	for i := 0; i < MultiplyBitPairs; i++ {
		if value&MultiplyBit2Mask != 0 {
			cycles++
		}
		value >>= MultiplyBitShift
		if value == 0 {
			break
		}
	}

	if cycles > MultiplyMaxCycles {
		cycles = MultiplyMaxCycles
	}

	return cycles
}

// ExecuteMultiplyLong executes long multiply instructions (UMULL, UMLAL, SMULL, SMLAL)
func ExecuteMultiplyLong(vm *VM, inst *Instruction) error {
	// Decode instruction fields
	// Bit [22] = U (1=unsigned UMULL/UMLAL, 0=signed SMULL/SMLAL)
	// Bit [21] = A (1=accumulate xMLAL, 0=multiply xMULL)
	// Bit [20] = S (set flags)
	unsignedOp := (inst.Opcode >> BBitShift) & Mask1Bit
	accumulate := (inst.Opcode >> MultiplyAShift) & Mask1Bit
	setFlags := inst.SetFlags

	rdHi := int((inst.Opcode >> RnShift) & Mask4Bit) // Destination high register
	rdLo := int((inst.Opcode >> RdShift) & Mask4Bit) // Destination low register
	rs := int((inst.Opcode >> RsShift) & Mask4Bit)   // Operand register 1
	rm := int(inst.Opcode & Mask4Bit)                // Operand register 2

	// Validate registers
	// RdHi, RdLo, Rm must all be different
	if rdHi == rdLo {
		return fmt.Errorf("long multiply: RdHi and RdLo must be different registers")
	}
	if rdHi == rm || rdLo == rm {
		return fmt.Errorf("long multiply: RdHi/RdLo and Rm must be different registers")
	}

	// R15 (PC) cannot be used
	if rdHi == PCRegister || rdLo == PCRegister || rm == PCRegister || rs == PCRegister {
		return fmt.Errorf("long multiply: R15 (PC) cannot be used")
	}

	// Get operands
	op1 := vm.CPU.GetRegister(rm)
	op2 := vm.CPU.GetRegister(rs)

	var resultHi, resultLo uint32

	if unsignedOp == 1 {
		// Unsigned multiply
		result64 := uint64(op1) * uint64(op2)

		// Add accumulator if UMLAL
		if accumulate == 1 {
			accHi := uint64(vm.CPU.GetRegister(rdHi))
			accLo := uint64(vm.CPU.GetRegister(rdLo))
			accumulator := (accHi << 32) | accLo
			result64 += accumulator
		}

		// Safe: extracting 32-bit words from 64-bit result
		resultHi = uint32(result64 >> BitsInWord) // #nosec G115 -- extracting high 32 bits
		resultLo = uint32(result64 & Mask32Bit)   // #nosec G115 -- extracting low 32 bits
	} else {
		// Signed multiply
		// Convert to signed 64-bit
		signedOp1 := int64(int32(op1)) // #nosec G115 -- intentional reinterpret cast for signed arithmetic
		signedOp2 := int64(int32(op2)) // #nosec G115 -- intentional reinterpret cast for signed arithmetic
		result64 := signedOp1 * signedOp2

		// Add accumulator if SMLAL
		if accumulate == 1 {
			accHi := uint64(vm.CPU.GetRegister(rdHi))
			accLo := uint64(vm.CPU.GetRegister(rdLo))
			accumulator := int64((accHi << 32) | accLo) // #nosec G115 -- reinterpreting 64-bit unsigned as signed
			result64 += accumulator
		}

		// Safe: extracting 32-bit words from 64-bit result
		resultHi = uint32(uint64(result64) >> BitsInWord) // #nosec G115 -- reinterpret signed to unsigned for bit extraction
		resultLo = uint32(uint64(result64) & Mask32Bit)   // #nosec G115 -- reinterpret signed to unsigned for bit extraction
	}

	// Store results
	vm.CPU.SetRegister(rdHi, resultHi)
	vm.CPU.SetRegister(rdLo, resultLo)

	// Update flags if requested
	if setFlags {
		// Long multiply updates N and Z flags based on 64-bit result
		// N = bit 63 of result
		// Z = result == 0
		n := (resultHi & SignBitMask) != 0
		z := (resultHi == 0) && (resultLo == 0)

		vm.CPU.CPSR.N = n
		vm.CPU.CPSR.Z = z
		// C and V are unaffected (or meaningless)
	}

	// Increment PC
	vm.CPU.IncrementPC()

	// Long multiply takes more cycles (typically 3-5 cycles for UMULL/SMULL, +1 for accumulate)
	cycles := LongMultiplyBaseCycles
	if accumulate == 1 {
		cycles = LongMultiplyAccumulateCycles
	}
	// Safe: cycles is 3 or 4, -1 = 2 or 3, well within uint64 range
	vm.CPU.IncrementCycles(uint64(cycles - 1)) // #nosec G115 -- cycles is 3-4, -1 because Step() already adds 1

	return nil
}
