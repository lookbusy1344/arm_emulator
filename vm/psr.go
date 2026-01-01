package vm

import (
	"fmt"
)

// ExecutePSRTransfer executes PSR transfer instructions (MRS, MSR)
func ExecutePSRTransfer(vm *VM, inst *Instruction) error {
	// MRS/MSR instruction format:
	// Bits [27:26] = 00
	// Bit [25] = 1 (distinguishes from other instructions)
	// Bit [22] = PSR type (0=CPSR, 1=SPSR) - we only support CPSR for now
	// Bit [21] = Direction (0=MRS read PSR, 1=MSR write PSR)

	isMSR := (inst.Opcode >> MultiplyAShift) & Mask1Bit // 1=MSR, 0=MRS

	if isMSR == 0 {
		return executeMRS(vm, inst)
	}
	return executeMSR(vm, inst)
}

// executeMRS implements MRS (Move PSR to Register)
// Syntax: MRS{cond} Rd, PSR
// Reads CPSR into a general-purpose register
func executeMRS(vm *VM, inst *Instruction) error {
	rd := int((inst.Opcode >> RdShift) & Mask4Bit) // Destination register

	// R15 (PC) should not be used as destination
	if rd == PCRegister {
		return fmt.Errorf("MRS: R15 (PC) cannot be used as destination register")
	}

	// Read CPSR value
	cpsrValue := vm.CPU.CPSR.ToUint32()

	// Store in destination register - if destination is SP, use SetSPWithTrace for bounds validation
	if rd == SP {
		if err := vm.CPU.SetSPWithTrace(vm, cpsrValue, inst.Address); err != nil {
			vm.State = StateError
			vm.LastError = err
			return err
		}
	} else {
		vm.CPU.SetRegister(rd, cpsrValue)
	}

	// Increment PC
	vm.CPU.IncrementPC()
	// Note: IncrementCycles is called by Step() in executor.go

	return nil
}

// executeMSR implements MSR (Move Register/Immediate to PSR)
// Syntax: MSR{cond} PSR, Rm
// Writes a general-purpose register value to CPSR
func executeMSR(vm *VM, inst *Instruction) error {
	// Check if immediate or register source
	immediateBit := (inst.Opcode >> IBitShift) & Mask1Bit

	var sourceValue uint32

	if immediateBit == 1 {
		// Immediate value (rare for MSR, but supported)
		immediate := inst.Opcode & ImmediateValueMask
		rotate := ((inst.Opcode >> RotationShift) & RotationMask) * RotationMultiplier
		// Rotate right
		if rotate == 0 {
			sourceValue = immediate
		} else {
			sourceValue = (immediate >> rotate) | (immediate << (BitsInWord - rotate))
		}
	} else {
		// Register source
		rm := int(inst.Opcode & Mask4Bit)

		// R15 (PC) should not be used as source
		if rm == PCRegister {
			return fmt.Errorf("MSR: R15 (PC) cannot be used as source register")
		}

		sourceValue = vm.CPU.GetRegister(rm)
	}

	// Check which fields to update (bit 16-19 specify field mask)
	// For ARM2/basic implementation, we only update the flag bits (bits 31-28)
	// More advanced implementations might support updating mode bits, etc.

	// Update CPSR from source value
	// Only update the flag bits (NZCV) in bits 31-28
	// This is a safety measure to prevent mode changes in this basic implementation
	vm.CPU.CPSR.FromUint32(sourceValue)

	// Increment PC
	vm.CPU.IncrementPC()
	// Note: IncrementCycles is called by Step() in executor.go

	return nil
}
