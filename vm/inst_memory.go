package vm

import (
	"fmt"
	"math"
)

// ExecuteLoadStore executes load/store instructions (LDR, STR, LDRB, STRB, LDRH, STRH)
func ExecuteLoadStore(v *VM, inst *Instruction) error {
	vm := v
	load := (inst.Opcode >> LBitShift) & Mask1Bit         // L bit: 1=load, 0=store
	byteTransfer := (inst.Opcode >> BBitShift) & Mask1Bit // B bit: 1=byte, 0=word
	writeBack := (inst.Opcode >> WBitShift) & Mask1Bit    // W bit: write address back to base
	preIndexed := (inst.Opcode >> PBitShift) & Mask1Bit   // P bit: 1=pre-indexed, 0=post-indexed
	addOffset := (inst.Opcode >> UBitShift) & Mask1Bit    // U bit: 1=add offset, 0=subtract

	rd := int((inst.Opcode >> RdShift) & Mask4Bit) // Data register
	rn := int((inst.Opcode >> RnShift) & Mask4Bit) // Base register

	baseAddr := vm.CPU.GetRegister(rn)

	// Check for halfword transfer (ARM2a extension) first
	// LDRH/STRH: bits[27:25]=000, bit7=1, bit4=1
	// LDR/STR:   bits[27:26]=01
	bits27_25 := (inst.Opcode >> Bits27_25Shift) & Mask3Bit
	bit7 := (inst.Opcode >> Bit7Pos) & Mask1Bit
	bit4 := (inst.Opcode >> Bit4Pos) & Mask1Bit
	isHalfword := bits27_25 == 0 && bit7 == 1 && bit4 == 1

	// Calculate offset
	var offset uint32
	if isHalfword {
		// Halfword instructions use different encoding
		// I bit is at position 22 for halfword (1=immediate, 0=register)
		immediate := (inst.Opcode >> BBitShift) & Mask1Bit

		if immediate == 1 {
			// Immediate offset: split into high[11:8] and low[3:0]
			offsetHigh := (inst.Opcode >> HalfwordHighShift) & HalfwordOffsetHighMask
			offsetLow := inst.Opcode & HalfwordOffsetLowMask
			offset = (offsetHigh << HalfwordLowShift) | offsetLow
		} else {
			// Register offset
			rm := int(inst.Opcode & Mask4Bit)
			offset = vm.CPU.GetRegister(rm)
		}
	} else {
		// Standard word/byte transfer
		// I bit at position 25 (inverted: 0=immediate, 1=register)
		immediate := ((inst.Opcode >> IBitShift) & Mask1Bit) == 0

		if immediate {
			// Immediate offset
			offset = inst.Opcode & Offset12BitMask
		} else {
			// Register offset with optional shift
			rm := int(inst.Opcode & Mask4Bit)
			offsetReg := vm.CPU.GetRegister(rm)

			shiftType := ShiftType((inst.Opcode >> ShiftTypePos) & Mask2Bit)
			shiftAmount := int((inst.Opcode >> ShiftAmountPos) & Mask5Bit)

			offset = PerformShift(offsetReg, shiftAmount, shiftType, vm.CPU.CPSR.C)
		}
	}

	// Apply sign of offset with overflow/underflow detection
	var effectiveAddr uint32
	if addOffset == 1 {
		// Check for unsigned overflow: baseAddr + offset > MaxUint32
		if offset > 0 && baseAddr > math.MaxUint32-offset {
			return fmt.Errorf("address overflow: base 0x%08X + offset 0x%08X wraps around", baseAddr, offset)
		}
		effectiveAddr = baseAddr + offset
	} else {
		// Check for unsigned underflow: baseAddr - offset < 0
		if offset > baseAddr {
			return fmt.Errorf("address underflow: base 0x%08X - offset 0x%08X wraps around", baseAddr, offset)
		}
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

	// Perform load or store
	if load == 1 {
		// Load instruction
		var value uint32
		var err error
		var sizeStr string

		if isHalfword {
			// Load halfword
			halfValue, err2 := vm.Memory.ReadHalfword(accessAddr)
			value = uint32(halfValue)
			err = err2
			sizeStr = "HALF"
		} else if byteTransfer == 1 {
			// Load byte
			byteValue, err2 := vm.Memory.ReadByteAt(accessAddr)
			value = uint32(byteValue)
			err = err2
			sizeStr = "BYTE"
		} else {
			// Load word
			value, err = vm.Memory.ReadWord(accessAddr)
			sizeStr = "WORD"
		}

		if err != nil {
			return fmt.Errorf("load failed at 0x%08X: %w", accessAddr, err)
		}

		// Record memory trace if enabled
		if vm.MemoryTrace != nil {
			vm.MemoryTrace.RecordRead(vm.CPU.Cycles, vm.CPU.PC, accessAddr, value, sizeStr)
		}

		// If loading to SP (R13), use SetSPWithTrace for bounds validation
		if rd == SP {
			if err := vm.CPU.SetSPWithTrace(vm, value, vm.CPU.PC); err != nil {
				vm.State = StateError
				vm.LastError = err
				return err
			}
		} else {
			vm.CPU.SetRegister(rd, value)
		}
	} else {
		// Store instruction
		value := vm.CPU.GetRegister(rd)
		var err error
		var sizeStr string

		if isHalfword {
			// Store halfword - ARM architecture truncates to lower 16 bits
			//nolint:gosec // G115: Intentional truncation for STRH instruction
			err = vm.Memory.WriteHalfword(accessAddr, uint16(value&HalfwordValueMask))
			sizeStr = "HALF"
		} else if byteTransfer == 1 {
			// Store byte - ARM architecture truncates to lower 8 bits
			//nolint:gosec // G115: Intentional truncation for STRB instruction
			err = vm.Memory.WriteByteAt(accessAddr, uint8(value&ByteValueMask))
			sizeStr = "BYTE"
		} else {
			// Store word
			err = vm.Memory.WriteWord(accessAddr, value)
			sizeStr = "WORD"
		}

		if err != nil {
			return fmt.Errorf("store failed at 0x%08X: %w", accessAddr, err)
		}

		// Track last memory write for GUI
		vm.LastMemoryWrite = accessAddr
		vm.HasMemoryWrite = true

		// Record memory trace if enabled
		if vm.MemoryTrace != nil {
			vm.MemoryTrace.RecordWrite(vm.CPU.Cycles, vm.CPU.PC, accessAddr, value, sizeStr)
		}
	}

	// Write back effective address to base register if requested
	if (preIndexed == 1 && writeBack == 1) || preIndexed == 0 {
		// Pre-indexed with writeback or post-indexed always writes back
		if rn != PCRegister { // Don't write back to PC
			vm.CPU.SetRegister(rn, effectiveAddr)
		}
	}

	// Increment PC (unless we loaded into PC)
	if !(load == 1 && rd == PCRegister) {
		vm.CPU.IncrementPC()
	}

	return nil
}
