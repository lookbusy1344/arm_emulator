package vm

import (
	"fmt"
)

// ExecuteLoadStore executes load/store instructions (LDR, STR, LDRB, STRB, LDRH, STRH)
func ExecuteLoadStore(v *VM, inst *Instruction) error {
	vm := v
	load := (inst.Opcode >> 20) & 0x1             // L bit: 1=load, 0=store
	byteTransfer := (inst.Opcode >> 22) & 0x1     // B bit: 1=byte, 0=word
	writeBack := (inst.Opcode >> 21) & 0x1        // W bit: write address back to base
	preIndexed := (inst.Opcode >> 24) & 0x1       // P bit: 1=pre-indexed, 0=post-indexed
	addOffset := (inst.Opcode >> 23) & 0x1        // U bit: 1=add offset, 0=subtract
	immediate := ((inst.Opcode >> 25) & 0x1) == 0 // I bit inverted: 0=immediate, 1=register

	rd := int((inst.Opcode >> 12) & 0xF) // Data register
	rn := int((inst.Opcode >> 16) & 0xF) // Base register

	baseAddr := vm.CPU.GetRegister(rn)

	// Calculate offset
	var offset uint32
	if immediate {
		// Immediate offset
		offset = inst.Opcode & 0xFFF
	} else {
		// Register offset with optional shift
		rm := int(inst.Opcode & 0xF)
		offsetReg := vm.CPU.GetRegister(rm)

		shiftType := ShiftType((inst.Opcode >> 5) & 0x3)
		shiftAmount := int((inst.Opcode >> 7) & 0x1F)

		offset = PerformShift(offsetReg, shiftAmount, shiftType, vm.CPU.CPSR.C)
	}

	// Apply sign of offset
	var effectiveAddr uint32
	if addOffset == 1 {
		effectiveAddr = baseAddr + offset
	} else {
		effectiveAddr = baseAddr - offset
	}

	// Determine which address to use
	var accessAddr uint32
	if preIndexed == 1 {
		// Pre-indexed: use effective address
		accessAddr = effectiveAddr
	} else {
		// Post-indexed: use base address
		accessAddr = baseAddr
	}

	// Check for halfword transfer (ARM2a extension)
	// Identified by bits [7:4] == 0b1011 (load) or 0b1001 (store) and bit 22 == 0
	halfwordBits := (inst.Opcode >> 4) & 0xF
	isHalfword := byteTransfer == 0 && (halfwordBits == 0xB || halfwordBits == 0x9)

	// Perform load or store
	if load == 1 {
		// Load instruction
		var value uint32
		var err error

		if isHalfword {
			// Load halfword
			halfValue, err2 := vm.Memory.ReadHalfword(accessAddr)
			value = uint32(halfValue)
			err = err2
		} else if byteTransfer == 1 {
			// Load byte
			byteValue, err2 := vm.Memory.ReadByteAt(accessAddr)
			value = uint32(byteValue)
			err = err2
		} else {
			// Load word
			value, err = vm.Memory.ReadWord(accessAddr)
		}

		if err != nil {
			return fmt.Errorf("load failed at 0x%08X: %w", accessAddr, err)
		}

		vm.CPU.SetRegister(rd, value)
	} else {
		// Store instruction
		value := vm.CPU.GetRegister(rd)
		var err error

		if isHalfword {
			// Store halfword - ARM architecture truncates to lower 16 bits
			//nolint:gosec // G115: Intentional truncation for STRH instruction
			err = vm.Memory.WriteHalfword(accessAddr, uint16(value&0xFFFF))
		} else if byteTransfer == 1 {
			// Store byte - ARM architecture truncates to lower 8 bits
			//nolint:gosec // G115: Intentional truncation for STRB instruction
			err = vm.Memory.WriteByteAt(accessAddr, uint8(value&0xFF))
		} else {
			// Store word
			err = vm.Memory.WriteWord(accessAddr, value)
		}

		if err != nil {
			return fmt.Errorf("store failed at 0x%08X: %w", accessAddr, err)
		}
	}

	// Write back effective address to base register if requested
	if (preIndexed == 1 && writeBack == 1) || preIndexed == 0 {
		// Pre-indexed with writeback or post-indexed always writes back
		if rn != 15 { // Don't write back to PC
			vm.CPU.SetRegister(rn, effectiveAddr)
		}
	}

	// Increment PC (unless we loaded into PC)
	if !(load == 1 && rd == 15) {
		vm.CPU.IncrementPC()
	}

	return nil
}
