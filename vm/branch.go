package vm

import ()

// ExecuteBranch executes branch instructions (B, BL, BX, BLX)
func ExecuteBranch(vm *VM, inst *Instruction) error {
	// Check for BX (Branch and Exchange): bits [27:4] = 0x12FFF1
	if (inst.Opcode & BXPatternMask) == BXEncodingBase {
		return ExecuteBranchExchange(vm, inst)
	}

	// Check for BLX register form: bits [27:4] = 0x12FFF3
	if (inst.Opcode & BXPatternMask) == BLXEncodingBase {
		return ExecuteBranchLinkExchange(vm, inst)
	}

	link := (inst.Opcode >> BranchLinkShift) & Mask1Bit // L bit: 1=BL (branch with link), 0=B (branch)

	// Extract 24-bit signed offset and sign-extend to 32 bits
	offset := inst.Opcode & Offset24BitMask

	// Sign extend if bit 23 is set
	if (offset & Offset24BitSignBit) != 0 {
		offset |= Offset24BitSignExt
	}

	// Offset is in words, shift left by 2 to get byte offset
	// Add 8 to account for PC being 2 instructions ahead (ARM pipeline)
	targetAddr := vm.CPU.PC + PCBranchBase + (offset << WordToByteShift)

	// If this is a branch with link, save return address
	if link == 1 {
		vm.CPU.BranchWithLink(targetAddr)
	} else {
		vm.CPU.Branch(targetAddr)
	}

	return nil
}

// ExecuteBranchExchange executes BX (branch and exchange) instruction
// This is primarily for ARM/Thumb interworking, but in ARM2 we just branch
func ExecuteBranchExchange(vm *VM, inst *Instruction) error {
	rm := int(inst.Opcode & Mask4Bit) // Register containing target address
	targetAddr := vm.CPU.GetRegister(rm)

	// In a full ARM implementation, bit 0 would indicate Thumb mode
	// For ARM2 emulation, we just branch to the address (clearing bit 0)
	vm.CPU.Branch(targetAddr & ThumbModeClearMask)

	return nil
}

// ExecuteBranchLinkExchange executes BLX register form (branch with link and exchange)
func ExecuteBranchLinkExchange(vm *VM, inst *Instruction) error {
	rm := int(inst.Opcode & Mask4Bit) // Register containing target address
	targetAddr := vm.CPU.GetRegister(rm)

	// Save return address and branch
	vm.CPU.BranchWithLink(targetAddr & ThumbModeClearMask)

	return nil
}
