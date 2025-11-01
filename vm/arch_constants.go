package vm

// ============================================================================
// ARM Instruction Encoding Architecture Constants
// ============================================================================
// These constants define the ARM instruction encoding format as specified
// by the ARM architecture. They are shared between encoder and decoder.

// Instruction Field Bit Positions
// These define where fields appear in the 32-bit instruction encoding
const (
	// Condition code field (bits 31-28)
	ConditionShift = 28

	// Common instruction field positions
	OpcodeShift = 21 // Bits 24-21: opcode field
	SBitShift   = 20 // Bit 20: S bit (set flags)
	RnShift     = 16 // Bits 19-16: Rn (first operand register)
	RdShift     = 12 // Bits 15-12: Rd (destination register)
	RsShift     = 8  // Bits 11-8: Rs (shift register)

	// Memory instruction bit positions
	PBitShift = 24 // Bit 24: P (pre/post indexing)
	UBitShift = 23 // Bit 23: U (up/down - add/subtract offset)
	BBitShift = 22 // Bit 22: B (byte/word)
	WBitShift = 21 // Bit 21: W (writeback)
	LBitShift = 20 // Bit 20: L (load/store)

	// Branch instruction
	BranchLinkShift = 24 // Bit 24: L bit for BL
)

// ARM Register Numbers
const (
	ARMRegisterPC = 15 // Program Counter (R15)
	ARMRegisterLR = 14 // Link Register (R14)
	ARMRegisterSP = 13 // Stack Pointer (R13)
)
