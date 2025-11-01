package encoder

import "github.com/lookbusy1344/arm-emulator/vm"

// Instruction Encoding Bit Shift Positions
// These constants define the bit positions used in ARM instruction encoding
const (
	// Common shift positions used across all instruction types
	ConditionShift = 28 // Bits 31-28: condition code
	TypeShift25    = 25 // Bit 25: often used for I bit or instruction type
	TypeShift26    = 26 // Bit 26: instruction type bit

	// Data processing and memory instruction shifts
	OpcodeShift = 21 // Bits 24-21: opcode for data processing
	SBitShift   = 20 // Bit 20: set flags bit
	RnShift     = 16 // Bits 19-16: first operand register
	RdShift     = 12 // Bits 15-12: destination register
	RsShift     = 8  // Bits 11-8: shift register
	ShiftAmount = 7  // Bits 11-7: shift amount
	ShiftType   = 5  // Bits 6-5: shift type
	Bit4        = 4  // Bit 4: register/immediate shift indicator

	// Memory instruction specific shifts
	PBitShift = 24 // Bit 24: pre/post indexing
	UBitShift = 23 // Bit 23: up/down (add/subtract offset)
	BBitShift = 22 // Bit 22: byte/word
	WBitShift = 21 // Bit 21: writeback
	LBitShift = 20 // Bit 20: load/store

	// Branch instruction shifts
	BranchLinkShift = 24 // Bit 24: link bit for BL

	// Multiply instruction shifts
	MultiplyABitShift = 21 // Bit 21: accumulate bit for MLA

	// Halfword instruction shifts
	HalfwordHBitShift = 5 // Bit 5: halfword bit
	HalfwordSBitShift = 6 // Bit 6: signed bit
	HalfwordBit7      = 7 // Bit 7: always 1 for halfword
	HalfwordIBitShift = 22 // Bit 22: immediate bit for halfword
)

// Note: Bit masks (Mask4Bit, Mask12Bit, Mask16Bit, Mask24Bit) are imported from vm package

// Immediate Value Limits
const (
	MaxOffset12Bit      = 4095   // Maximum 12-bit offset (0xFFF)
	MaxOffsetHalfword   = 255    // Maximum 8-bit halfword offset
	MaxBranchOffsetPos  = 0x7FFFFF  // Maximum positive 24-bit branch offset
	MinBranchOffsetNeg  = -0x800000 // Minimum negative 24-bit branch offset
)

// Note: Register numbers (RegisterPC, RegisterLR, RegisterSP) are available as ARMRegisterPC in vm package
// For encoder-specific use, we define local aliases:
const (
	RegisterPC = 15 // Program Counter (R15)
	RegisterLR = 14 // Link Register (R14)
	RegisterSP = 13 // Stack Pointer (R13)
)

// Instruction Type Values (for bits shifted into position)
const (
	BranchTypeValue    = 5    // Value 0b101 (5) in bits 27-25 for branch (before shift)
	LDMSTMTypeValue    = 4    // Value 0b100 (4) in bits 27-25 for LDM/STM (before shift)
	MultiplyMarker     = 9    // Value 0b1001 (9) in bits 7-4 for multiply instructions
	MOVOpcodeValue     = 0xD  // MOV opcode (13, 0b1101)
	SWITypeValue       = 0xF  // SWI type bits (15, 0b1111)
	MOVWOpcodeValue    = 0x30 // MOVW opcode pattern (0b110000)
)

// Word Size
const (
	WordSize = 4 // ARM instructions and words are 4 bytes
)
