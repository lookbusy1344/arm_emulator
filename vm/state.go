package vm

// RegisterSnapshot captures the state of CPU registers for change detection
type RegisterSnapshot struct {
	R    [16]uint32 // R0-R15 (PC is R15)
	CPSR CPSR
}

// Capture captures the current state of the CPU
func (s *RegisterSnapshot) Capture(cpu *CPU) {
	copy(s.R[:15], cpu.R[:])
	s.R[15] = cpu.PC
	s.CPSR = cpu.CPSR
}

// ChangedRegisters returns a list of registers that have changed compared to another snapshot
// Returns indices of changed registers (0-15)
func (s *RegisterSnapshot) ChangedRegisters(other *RegisterSnapshot) []int {
	var changed []int
	for i := 0; i < 16; i++ {
		if s.R[i] != other.R[i] {
			changed = append(changed, i)
		}
	}
	return changed
}

// CPSRChanged returns true if CPSR flags have changed
func (s *RegisterSnapshot) CPSRChanged(other *RegisterSnapshot) bool {
	return s.CPSR != other.CPSR
}

// GetRegister returns the value of a register from the snapshot
func (s *RegisterSnapshot) GetRegister(reg int) uint32 {
	if reg >= 0 && reg < 16 {
		return s.R[reg]
	}
	return 0
}
