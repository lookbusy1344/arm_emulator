package vm

// ARM2 Architecture Constants
// These values are defined by the ARM2 specification and should not be modified

const (
	// Instruction encoding
	ARMInstructionSize = 4 // bytes
	ARMPipelineOffset  = 8 // PC is instruction address + 8

	// Register counts
	ARMGeneralRegisterCount = 15 // R0-R14
	ARMTotalRegisterCount   = 16 // Including PC (R15)
	ARMRegisterPC           = 15

	// CPSR flag bit positions (bits 31-28)
	CPSRBitN = 31 // Negative flag
	CPSRBitZ = 30 // Zero flag
	CPSRBitC = 29 // Carry flag
	CPSRBitV = 28 // Overflow flag

	// Instruction encoding bit positions
	InstructionConditionShift      = 28 // bits 31-28
	InstructionTypeShift           = 26 // bits 27-26
	InstructionImmediateShift      = 25 // bit 25
	InstructionPreIndexShift       = 24 // bit 24 (P bit)
	InstructionUpShift             = 23 // bit 23 (U bit)
	InstructionByteShift           = 22 // bit 22 (B bit)
	InstructionWritebackShift      = 21 // bit 21 (W bit)
	InstructionLoadShift           = 20 // bit 20 (L bit)
	InstructionRnShift             = 16 // bits 19-16
	InstructionRdShift             = 12 // bits 15-12
	InstructionRsShift             = 8  // bits 11-8
	InstructionShiftTypeShift      = 5  // bits 6-5
	InstructionShiftAmountShift    = 7  // bits 11-7

	// Bit masks
	Mask4Bit  = 0xF
	Mask8Bit  = 0xFF
	Mask12Bit = 0xFFF
	Mask16Bit = 0xFFFF
	Mask24Bit = 0xFFFFFF
	Mask32Bit = 0xFFFFFFFF

	// Alignment constants (grouped together for discoverability)
	AlignmentWord        = 4          // 4-byte word alignment
	AlignmentHalfword    = 2          // 2-byte halfword alignment
	AlignmentByte        = 1          // no alignment required
	AlignMaskWord        = 0x3        // mask for word alignment check (address & mask == 0 means aligned)
	AlignMaskHalfword    = 0x1        // mask for halfword alignment check
	AlignRoundUpMaskWord = 0xFFFFFFFC // mask to round up to word alignment (~0x3)

	// Signed integer ranges (for branch offsets, etc.)
	Int24Max = 0x7FFFFF  // Maximum positive 24-bit signed value
	Int24Min = -0x800000 // Minimum negative 24-bit signed value
)

// Special instruction encoding patterns
const (
	// BX/BLX patterns are pre-shifted by 4 bits for direct use in encoding
	// Usage: instruction := (cond << 28) | BXEncodingBase | rm
	BXEncodingBase  = 0x12FFF10  // BX instruction base (0x12FFF1 << 4)
	BLXEncodingBase = 0x12FFF30  // BLX instruction base (0x12FFF3 << 4)
	NOPEncoding     = 0xE1A00000 // MOV R0, R0
)
