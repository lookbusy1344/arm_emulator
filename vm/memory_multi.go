package vm

import (
	"fmt"
)

// ExecuteLoadStoreMultiple executes load/store multiple instructions (LDM, STM)
func ExecuteLoadStoreMultiple(vm *VM, inst *Instruction) error {
	load := (inst.Opcode >> LBitShift) & Mask1Bit      // L bit: 1=load, 0=store
	writeBack := (inst.Opcode >> WBitShift) & Mask1Bit // W bit: write address back to base
	psr := (inst.Opcode >> BBitShift) & Mask1Bit       // S bit: load/store PSR or force user mode
	increment := (inst.Opcode >> UBitShift) & Mask1Bit // U bit: 1=increment, 0=decrement
	preIndex := (inst.Opcode >> PBitShift) & Mask1Bit  // P bit: 1=pre-increment/decrement, 0=post

	rn := int((inst.Opcode >> RnShift) & Mask4Bit) // Base register
	regList := inst.Opcode & RegisterListMask      // Register list (bits 0-15)

	baseAddr := vm.CPU.GetRegister(rn)

	// Count number of registers in list
	numRegs := 0
	for i := 0; i < 16; i++ {
		if (regList & (1 << i)) != 0 {
			numRegs++
		}
	}

	if numRegs == 0 {
		return fmt.Errorf("load/store multiple with empty register list")
	}

	// Calculate starting address based on addressing mode
	var addr uint32
	if increment == 1 {
		// Incrementing
		if preIndex == 1 {
			// Pre-increment (IB - Increment Before)
			addr = baseAddr + MultiRegisterWordSize
		} else {
			// Post-increment (IA - Increment After)
			addr = baseAddr
		}
	} else {
		// Decrementing
		offset, err := SafeIntToUint32(numRegs * MultiRegisterWordSize)
		if err != nil {
			return fmt.Errorf("register count too large: %w", err)
		}
		// Security: check for underflow before subtraction
		if offset > baseAddr {
			return fmt.Errorf("address underflow: baseAddr=0x%08X, offset=0x%08X", baseAddr, offset)
		}
		if preIndex == 1 {
			// Pre-decrement (DB - Decrement Before)
			addr = baseAddr - offset
		} else {
			// Post-decrement (DA - Decrement After)
			addr = baseAddr - offset + MultiRegisterWordSize
		}
	}

	// Save the start address for writeback calculation
	regOffset, err := SafeIntToUint32(numRegs * MultiRegisterWordSize)
	if err != nil {
		return fmt.Errorf("register count too large: %w", err)
	}
	var newBase uint32
	if increment == 1 {
		newBase = baseAddr + regOffset
	} else {
		// Security: check for underflow in writeback calculation
		if regOffset > baseAddr {
			return fmt.Errorf("address underflow in writeback: baseAddr=0x%08X, offset=0x%08X", baseAddr, regOffset)
		}
		newBase = baseAddr - regOffset
	}

	// Process each register in the list
	pcLoaded := false
	for i := 0; i < 16; i++ {
		if (regList & (1 << i)) == 0 {
			continue
		}

		if load == 1 {
			// Load register
			value, err := vm.Memory.ReadWord(addr)
			if err != nil {
				return fmt.Errorf("load multiple failed at 0x%08X: %w", addr, err)
			}

			// Record memory trace if enabled
			if vm.MemoryTrace != nil {
				vm.MemoryTrace.RecordRead(vm.CPU.Cycles, vm.CPU.PC, addr, value, "WORD")
			}

			// If loading to SP (R13), use SetSPWithTrace for bounds validation
			if i == SP {
				if err := vm.CPU.SetSPWithTrace(vm, value, inst.Address); err != nil {
					vm.State = StateError
					vm.LastError = err
					return err
				}
			} else {
				vm.CPU.SetRegister(i, value)
			}

			if i == PCRegister {
				pcLoaded = true
			}
		} else {
			// Store register
			value := vm.CPU.GetRegister(i)

			// If storing R15 (PC), store PC+12 (current instruction + 8 + 4)
			if i == PCRegister {
				value = vm.CPU.PC + PCStoreOffset
			}

			err := vm.Memory.WriteWord(addr, value)
			if err != nil {
				return fmt.Errorf("store multiple failed at 0x%08X: %w", addr, err)
			}

			// Track last memory write for GUI
			vm.LastMemoryWrite = addr
			vm.HasMemoryWrite = true

			// Record memory trace if enabled
			if vm.MemoryTrace != nil {
				vm.MemoryTrace.RecordWrite(vm.CPU.Cycles, vm.CPU.PC, addr, value, "WORD")
			}
		}

		addr += MultiRegisterWordSize
	}

	// Write back to base register if requested
	if writeBack == 1 && rn != PCRegister {
		// If modifying SP (R13), use SetSPWithTrace for bounds validation
		if rn == SPRegister {
			if err := vm.CPU.SetSPWithTrace(vm, newBase, inst.Address); err != nil {
				vm.State = StateError
				vm.LastError = err
				return err
			}
		} else {
			vm.CPU.SetRegister(rn, newBase)
		}
	}

	// Also check if SP was loaded (but not base register)
	if load == 1 && (regList&(1<<SPRegister)) != 0 && rn != SPRegister && vm.StackTrace != nil {
		// SP was loaded from memory, record as SP move
		vm.StackTrace.RecordSPMove(vm.CPU.Cycles, inst.Address, baseAddr, vm.CPU.GetSP())
	}

	// Handle S bit (PSR transfer for LDM with PC)
	// ARM6+ behavior: When loading PC with S bit set, restore CPSR from SPSR
	// This simulates returning from an exception handler
	if psr == 1 && load == 1 && pcLoaded {
		// LDM with S bit and PC loaded: restore CPSR from SPSR (exception return)
		vm.CPU.RestoreCPSR()
	}
	// Note: STM with S bit has no special behavior in this implementation
	// (storing PC+12 is sufficient for exception handling)

	// Increment PC (unless we loaded into PC)
	if !pcLoaded {
		vm.CPU.IncrementPC()
	}

	return nil
}
