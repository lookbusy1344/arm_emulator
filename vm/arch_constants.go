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

	// Note: Instruction encoding bit shift positions are defined in encoder/constants.go
	// They are encoder-specific and not part of the core ARM2 architecture specification

	// Bit masks
	Mask4Bit  = 0xF
	Mask8Bit  = 0xFF
	Mask12Bit = 0xFFF
	Mask16Bit = 0xFFFF
	Mask24Bit = 0xFFFFFF
	Mask32Bit = 0xFFFFFFFF

	// Byte shift positions for endianness conversions
	ByteShift8  = 8  // Shift for byte 1 in multibyte values
	ByteShift16 = 16 // Shift for byte 2 in multibyte values
	ByteShift24 = 24 // Shift for byte 3 in multibyte values

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
	// BX/BLX patterns are pre-positioned (NOT shift results - the trailing 0 is part of the hex value)
	// These are the actual bit patterns used directly in instruction encoding
	// Usage: instruction := (cond << 28) | BXEncodingBase | rm
	// The lower 4 bits (0000) are reserved for the rm register operand
	BXEncodingBase  = 0x012FFF10 // BX instruction base pattern (binary: 0b0000_0001_0010_1111_1111_1111_0001_0000)
	BLXEncodingBase = 0x012FFF30 // BLX instruction base pattern (binary: 0b0000_0001_0010_1111_1111_1111_0011_0000)
	NOPEncoding     = 0xE1A00000 // MOV R0, R0 (unconditional)
)
