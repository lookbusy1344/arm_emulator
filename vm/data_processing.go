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

	// Get first operand
	op1 := vm.CPU.GetRegister(rn)

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
			shiftCarry = (op2 & 0x80000000) != 0
		}
	} else {
		// Register with optional shift
		rm := int(inst.Opcode & Mask4Bit)
		op2Value := vm.CPU.GetRegister(rm)

		shiftType := ShiftType((inst.Opcode >> ShiftTypePos) & Mask2Bit)
		shiftByReg := (inst.Opcode >> Bit4Pos) & Mask1Bit

		var shiftAmount int
		if shiftByReg == 1 {
			// Shift amount in register
			rs := int((inst.Opcode >> RsShift) & Mask4Bit)
			shiftAmount = int(vm.CPU.GetRegister(rs) & ImmediateValueMask)
		} else {
			// Shift amount in instruction
			shiftAmount = int((inst.Opcode >> ShiftAmountPos) & Mask5Bit)
		}

		// In ARM, ROR #0 means RRX (rotate right extended through carry)
		if shiftType == ShiftROR && shiftAmount == 0 && shiftByReg == 0 {
			shiftType = ShiftRRX
		}

		shiftCarry = CalculateShiftCarry(op2Value, shiftAmount, shiftType, vm.CPU.CPSR.C)
		op2 = PerformShift(op2Value, shiftAmount, shiftType, vm.CPU.CPSR.C)
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
		result = op1 - op2
		carry = CalculateSubCarry(op1, op2)
		overflow = CalculateSubOverflow(op1, op2, result)

	case OpRSB:
		result = op2 - op1
		carry = CalculateSubCarry(op2, op1)
		overflow = CalculateSubOverflow(op2, op1, result)

	case OpADD:
		result = op1 + op2
		carry = CalculateAddCarry(op1, op2, result)
		overflow = CalculateAddOverflow(op1, op2, result)

	case OpADC:
		carryIn := uint32(0)
		if vm.CPU.CPSR.C {
			carryIn = 1
		}
		result = op1 + op2 + carryIn
		// Check if carry occurred from either addition
		temp := op1 + op2
		carry = CalculateAddCarry(op1, op2, temp) || CalculateAddCarry(temp, carryIn, result)
		overflow = CalculateAddOverflow(op1, op2, result)

	case OpSBC:
		carryIn := uint32(1)
		if !vm.CPU.CPSR.C {
			carryIn = 0
		}
		result = op1 - op2 - (1 - carryIn)
		carry = CalculateSubCarry(op1, op2+1-carryIn)
		overflow = CalculateSubOverflow(op1, op2+(1-carryIn), result)

	case OpRSC:
		carryIn := uint32(1)
		if !vm.CPU.CPSR.C {
			carryIn = 0
		}
		result = op2 - op1 - (1 - carryIn)
		carry = CalculateSubCarry(op2, op1+1-carryIn)
		overflow = CalculateSubOverflow(op2, op1+(1-carryIn), result)

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
		result = op1 - op2
		carry = CalculateSubCarry(op1, op2)
		overflow = CalculateSubOverflow(op1, op2, result)
		writeResult = false
		updateFlags = true // CMP always sets flags

	case OpCMN:
		result = op1 + op2
		carry = CalculateAddCarry(op1, op2, result)
		overflow = CalculateAddOverflow(op1, op2, result)
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
		// If writing to SP (R13), record stack trace
		if rd == SP && vm.StackTrace != nil {
			oldSP := vm.CPU.GetSP()
			vm.CPU.SetRegister(rd, result)
			vm.StackTrace.RecordSPMove(vm.CPU.Cycles, inst.Address, oldSP, result)
		} else {
			vm.CPU.SetRegister(rd, result)
		}
	}

	// Update flags if requested
	if updateFlags {
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
	if rd != PCRegister || opcode == OpCMP || opcode == OpCMN || opcode == OpTST || opcode == OpTEQ {
		vm.CPU.IncrementPC()
	}

	return nil
}
