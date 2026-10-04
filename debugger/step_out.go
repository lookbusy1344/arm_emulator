package debugger

import "github.com/lookbusy1344/arm-emulator/vm"

// Instruction patterns that return from a function. Masks ignore the condition field.
const (
	// MOV{S} PC, LR
	movPCLRMask    = 0x0FEFFFFF
	movPCLRPattern = 0x01A0F00E
	// LDM with PC in the register list (e.g. LDMFD SP!, {..., PC})
	ldmPCMask    = 0x0E108000
	ldmPCPattern = 0x08108000
	// LDR PC, [SP...] (single-register pop into PC)
	ldrPCFromSPMask    = 0x0C1FF000
	ldrPCFromSPPattern = 0x041DF000
)

func isCall(opcode uint32) bool {
	return opcode&vm.BranchLinkMask == vm.BranchLinkPattern ||
		opcode&vm.BXPatternMask == vm.BLXEncodingBase
}

func isReturn(opcode uint32) bool {
	return opcode&vm.BXPatternMask == vm.BXEncodingBase ||
		opcode&movPCLRMask == movPCLRPattern ||
		opcode&ldmPCMask == ldmPCPattern ||
		opcode&ldrPCFromSPMask == ldrPCFromSPPattern
}

// stepOutState follows calls and returns while stepping out. ShouldBreak runs before
// each instruction, so it judges the previous instruction by where the PC went.
type stepOutState struct {
	depth      int
	prevPC     uint32
	prevOpcode uint32
	havePrev   bool
}

// stepOutReturnedLocked records the instruction about to run at pc and reports whether the
// function active when step-out began has returned. The caller holds d.mu.
func (d *Debugger) stepOutReturnedLocked(pc uint32) bool {
	s := &d.stepOut
	if s.havePrev && pc != s.prevPC+vm.ARMInstructionSize {
		switch {
		case isCall(s.prevOpcode) && d.VM.CPU.GetLR() == s.prevPC+vm.ARMInstructionSize:
			s.depth++
		case isReturn(s.prevOpcode):
			s.depth--
		}
	}
	if s.depth < 0 {
		return true
	}
	opcode, err := d.VM.Memory.ReadWord(pc)
	s.prevPC, s.prevOpcode, s.havePrev = pc, opcode, err == nil
	return false
}
