package vm

import (
	"fmt"
)

// Data processing operation codes
const (
	OpAND = 0x0 // AND - Bitwise AND
	OpEOR = 0x1 // EOR - Bitwise Exclusive OR
	OpSUB = 0x2 // SUB - Subtract
	OpRSB = 0x3 // RSB - Reverse Subtract
	OpADD = 0x4 // ADD - Add
	OpADC = 0x5 // ADC - Add with Carry
	OpSBC = 0x6 // SBC - Subtract with Carry
	OpRSC = 0x7 // RSC - Reverse Subtract with Carry
	OpTST = 0x8 // TST - Test (AND without storing result)
	OpTEQ = 0x9 // TEQ - Test Equivalence (EOR without storing result)
	OpCMP = 0xA // CMP - Compare (SUB without storing result)
	OpCMN = 0xB // CMN - Compare Negative (ADD without storing result)
	OpORR = 0xC // ORR - Bitwise OR
	OpMOV = 0xD // MOV - Move
	OpBIC = 0xE // BIC - Bit Clear
	OpMVN = 0xF // MVN - Move Not
)

// ExecuteDataProcessing executes a data processing instruction
func ExecuteDataProcessing(vm *VM, inst *Instruction) error {
	opcode := (inst.Opcode >> OpcodeShift) & Mask4Bit
	immediate := (inst.Opcode >> IBitShift) & Mask1Bit
	setFlags := inst.SetFlags

	rd := int((inst.Opcode >> RdShift) & Mask4Bit) // Destination register
	rn := int((inst.Opcode >> RnShift) & Mask4Bit) // First operand register

	// A register-specified shift takes an extra cycle, so R15 reads as the
	// instruction address plus 12 instead of plus 8.
	shiftByReg := immediate == 0 && (inst.Opcode>>Bit4Pos)&Mask1Bit == 1
	readOperand := func(reg int) uint32 {
		if shiftByReg && reg == ARMRegisterPC {
			return vm.CPU.PC + PCStoreOffset
		}
		return vm.CPU.GetRegister(reg)
	}

	// Get first operand
	op1 := readOperand(rn)

	// Get second operand (either immediate or register with shift)
	var op2 uint32
	var shiftCarry bool

	if immediate == 1 {
		// Immediate value with rotation
		imm := inst.Opcode & ImmediateValueMask
		rotation := ((inst.Opcode >> RotationShift) & RotationMask) * RotationMultiplier
		op2 = (imm >> rotation) | (imm << (BitsInWord - rotation))

		// Carry from rotation
		if rotation == 0 {
			shiftCarry = vm.CPU.CPSR.C
		} else {
			shiftCarry = (op2 & SignBitMask) != 0
		}
	} else {
		// Register with optional shift
		rm := int(inst.Opcode & Mask4Bit)
		op2Value := readOperand(rm)

		shiftType := ShiftType((inst.Opcode >> ShiftTypePos) & Mask2Bit)

		var shiftAmount int
		if shiftByReg {
			// Shift amount in the bottom byte of Rs
			rs := int((inst.Opcode >> RsShift) & Mask4Bit)
			shiftAmount = int(vm.CPU.GetRegister(rs) & ImmediateValueMask)
		} else {
			shiftType, shiftAmount = immediateShift(shiftType, int((inst.Opcode>>ShiftAmountPos)&Mask5Bit))
		}

		if shiftByReg && shiftAmount == 0 {
			// A register shift by zero leaves the operand and carry unchanged
			op2, shiftCarry = op2Value, vm.CPU.CPSR.C
		} else {
			shiftCarry = CalculateShiftCarry(op2Value, shiftAmount, shiftType, vm.CPU.CPSR.C)
			op2 = PerformShift(op2Value, shiftAmount, shiftType, vm.CPU.CPSR.C)
		}
	}

	// Execute operation
	var result uint32
	var carry, overflow bool
	writeResult := true
	updateFlags := setFlags

	switch opcode {
	case OpAND:
		result = op1 & op2
		carry = shiftCarry

	case OpEOR:
		result = op1 ^ op2
		carry = shiftCarry

	case OpSUB:
		result, carry, overflow = AddWithCarry(op1, ^op2, true)

	case OpRSB:
		result, carry, overflow = AddWithCarry(op2, ^op1, true)

	case OpADD:
		result, carry, overflow = AddWithCarry(op1, op2, false)

	case OpADC:
		result, carry, overflow = AddWithCarry(op1, op2, vm.CPU.CPSR.C)

	case OpSBC:
		result, carry, overflow = AddWithCarry(op1, ^op2, vm.CPU.CPSR.C)

	case OpRSC:
		result, carry, overflow = AddWithCarry(op2, ^op1, vm.CPU.CPSR.C)

	case OpTST:
		result = op1 & op2
		carry = shiftCarry
		writeResult = false
		updateFlags = true // TST always sets flags

	case OpTEQ:
		result = op1 ^ op2
		carry = shiftCarry
		writeResult = false
		updateFlags = true // TEQ always sets flags

	case OpCMP:
		result, carry, overflow = AddWithCarry(op1, ^op2, true)
		writeResult = false
		updateFlags = true // CMP always sets flags

	case OpCMN:
		result, carry, overflow = AddWithCarry(op1, op2, false)
		writeResult = false
		updateFlags = true // CMN always sets flags

	case OpORR:
		result = op1 | op2
		carry = shiftCarry

	case OpMOV:
		result = op2
		carry = shiftCarry

	case OpBIC:
		result = op1 & ^op2
		carry = shiftCarry

	case OpMVN:
		result = ^op2
		carry = shiftCarry

	default:
		return fmt.Errorf("unknown data processing opcode: 0x%X", opcode)
	}

	// Write result to destination register
	if writeResult {
		// If writing to SP (R13), use SetSPWithTrace for bounds validation
		if rd == SP {
			if err := vm.CPU.SetSPWithTrace(vm, result, inst.Address); err != nil {
				vm.State = StateError
				vm.LastError = err
				return err
			}
		} else {
			vm.CPU.SetRegister(rd, result)
		}
	}

	// Update flags if requested. An S-suffixed write to PC is an exception return,
	// which needs an SPSR; user mode has none, so CPSR keeps its value.
	exceptionReturn := writeResult && rd == ARMRegisterPC
	if updateFlags && !exceptionReturn {
		// Logical operations update N, Z, C (not V)
		// Arithmetic operations update all flags
		if opcode == OpAND || opcode == OpEOR || opcode == OpTST || opcode == OpTEQ ||
			opcode == OpORR || opcode == OpMOV || opcode == OpBIC || opcode == OpMVN {
			vm.CPU.CPSR.UpdateFlagsNZC(result, carry)
		} else {
			vm.CPU.CPSR.UpdateFlagsNZCV(result, carry, overflow)
		}
	}

	// Increment PC (CMP/TST/TEQ/CMN never write result so always advance)
	if rd != ARMRegisterPC || opcode == OpCMP || opcode == OpCMN || opcode == OpTST || opcode == OpTEQ {
		vm.CPU.IncrementPC()
	}

	return nil
}
