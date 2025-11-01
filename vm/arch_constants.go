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

	// Sign bit for overflow calculations
	SignBitPos = 31 // Position of sign bit in 32-bit word
	SignBitMask = 0x80000000 // Mask for sign bit

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

// Instruction decoding bit shift positions
// These are used to extract fields from encoded ARM instructions
const (
	// Condition code field
	ConditionShift  = 28 // Bits 31-28: condition code

	// Common instruction field positions
	OpcodeShift     = 21 // Bits 24-21: opcode field
	SBitShift       = 20 // Bit 20: S bit (set flags)
	RnShift         = 16 // Bits 19-16: Rn (first operand register)
	RdShift         = 12 // Bits 15-12: Rd (destination register)
	RsShift         = 8  // Bits 11-8: Rs (shift register)
	ShiftAmountPos  = 7  // Bits 11-7: shift amount
	ShiftTypePos    = 5  // Bits 6-5: shift type
	Bit4Pos         = 4  // Bit 4: various uses
	Bit7Pos         = 7  // Bit 7: various uses

	// Memory instruction bit positions
	PBitShift       = 24 // Bit 24: P (pre/post indexing)
	UBitShift       = 23 // Bit 23: U (up/down - add/subtract offset)
	BBitShift       = 22 // Bit 22: B (byte/word)
	WBitShift       = 21 // Bit 21: W (writeback)
	LBitShift       = 20 // Bit 20: L (load/store)
	IBitShift       = 25 // Bit 25: I (immediate/register)

	// Branch/multiply field positions
	BranchLinkShift = 24 // Bit 24: L bit for BL
	MultiplyAShift  = 21 // Bit 21: A bit (accumulate)

	// Bit ranges for multi-bit fields
	Bits27_26Shift  = 26 // Bits 27-26 starting position
	Bits27_25Shift  = 25 // Bits 27-25 starting position
	Bits27_23Shift  = 23 // Bits 27-23 starting position
)

// Instruction decoding bit masks
const (
	// Field extraction masks (applied after shifting)
	Mask1Bit  = 0x1
	Mask2Bit  = 0x3
	Mask3Bit  = 0x7
	Mask5Bit  = 0x1F

	// Pre-shifted masks for common patterns
	BXPatternMask      = 0x0FFFFFF0 // Mask for BX/BLX detection
	LongMultiplyMask5  = 0x1F       // 5-bit mask for long multiply detection

	// Offset masks
	Offset12BitMask    = 0xFFF     // 12-bit immediate offset
	Offset24BitMask    = 0xFFFFFF  // 24-bit branch offset
	Offset24BitSignBit = 0x800000  // Sign bit for 24-bit offset
	Offset24BitSignExt = 0xFF000000 // Sign extension mask for 24-bit offset

	// Halfword transfer field masks
	HalfwordOffsetHighMask = 0xF   // High nibble of halfword offset (bits 11-8)
	HalfwordOffsetLowMask  = 0xF   // Low nibble of halfword offset (bits 3-0)
	HalfwordHighShift      = 8     // Shift for high nibble
	HalfwordLowShift       = 4     // Shift for assembling halfword offset

	// Register list mask (for LDM/STM)
	RegisterListMask = 0xFFFF // Bits 0-15: register list

	// Immediate value masks
	ImmediateValueMask = 0xFF  // 8-bit immediate value
	RotationMask       = 0xF   // 4-bit rotation value
	RotationShift      = 8     // Position of rotation field

	// Value truncation masks
	ByteValueMask     = 0xFF   // Mask for byte values
	HalfwordValueMask = 0xFFFF // Mask for halfword values

	// Multiply cycle calculation
	MultiplyBit2Mask = 0x3 // 2-bit mask for multiply timing
)

// Special values
const (
	// PC offset adjustments
	PCStoreOffset = 12 // PC+12 when storing PC in STM
	PCBranchBase  = 8  // PC+8 base for branch calculations

	// Register boundaries
	PCRegister = 15 // R15 is the PC
	SPRegister = 13 // R13 is the SP
	LRRegister = 14 // R14 is the LR

	// Bit shift for word-to-byte offset conversion
	WordToByteShift = 2 // Shift left by 2 to convert word offset to byte offset

	// Thumb mode bit (for BX instruction, though not fully supported in ARM2)
	ThumbModeClearMask = 0xFFFFFFFE // Mask to clear bit 0 (Thumb mode indicator)

	// Rotate constants
	RotationMultiplier = 2  // Rotation field is multiplied by 2
	BitsInWord         = 32 // 32 bits in a word

	// Multiply timing constants
	MultiplyBaseCycles  = 2  // Base cycle count for multiply
	MultiplyMaxCycles   = 16 // Maximum cycle count for multiply
	MultiplyBitPairs    = 16 // Number of 2-bit pairs to check
	MultiplyBitShift    = 2  // Shift for each iteration

	// Long multiply timing
	LongMultiplyBaseCycles       = 3 // Base cycles for UMULL/SMULL
	LongMultiplyAccumulateCycles = 4 // Cycles for UMLAL/SMLAL

	// Address alignment for LDM/STM
	MultiRegisterWordSize = 4 // 4 bytes per register in multi-register transfers
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
