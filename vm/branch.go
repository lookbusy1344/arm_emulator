package vm

import ()

// ExecuteBranch executes branch instructions (B, BL, BX)
func ExecuteBranch(vm *VM, inst *Instruction) error {
	link := (inst.Opcode >> 24) & 0x1 // L bit: 1=BL (branch with link), 0=B (branch)

	// Extract 24-bit signed offset and sign-extend to 32 bits
	offset := inst.Opcode & 0x00FFFFFF

	// Sign extend if bit 23 is set
	if (offset & 0x00800000) != 0 {
		offset |= 0xFF000000
	}

	// Offset is in words, shift left by 2 to get byte offset
	// Add 8 to account for PC being 2 instructions ahead (ARM pipeline)
	targetAddr := vm.CPU.PC + 8 + (offset << 2)

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
	rm := int(inst.Opcode & 0xF) // Register containing target address
	targetAddr := vm.CPU.GetRegister(rm)

	// In a full ARM implementation, bit 0 would indicate Thumb mode
	// For ARM2 emulation, we just branch to the address (clearing bit 0)
	vm.CPU.Branch(targetAddr & 0xFFFFFFFE)

	return nil
}
