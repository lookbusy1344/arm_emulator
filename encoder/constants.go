package encoder

import "github.com/lookbusy1344/arm-emulator/vm"

// ARM architectural constants (bit positions, register numbers) are in vm/arch_constants.go
// They are imported here for convenience
const (
	ConditionShift  = vm.ConditionShift
	OpcodeShift     = vm.OpcodeShift
	SBitShift       = vm.SBitShift
	RnShift         = vm.RnShift
	RdShift         = vm.RdShift
	RsShift         = vm.RsShift
	PBitShift       = vm.PBitShift
	UBitShift       = vm.UBitShift
	BBitShift       = vm.BBitShift
	WBitShift       = vm.WBitShift
	LBitShift       = vm.LBitShift
	BranchLinkShift = vm.BranchLinkShift
	RegisterPC      = vm.ARMRegisterPC
	RegisterLR      = vm.ARMRegisterLR
	RegisterSP      = vm.ARMRegisterSP
)

// Encoder-specific bit shift positions
const (
	TypeShift25 = 25 // Bit 25: often used for I bit or instruction type
	TypeShift26 = 26 // Bit 26: instruction type bit

	ShiftAmount = 7 // Bits 11-7: shift amount
	ShiftType   = 5 // Bits 6-5: shift type
	Bit4        = 4 // Bit 4: register/immediate shift indicator

	// Multiply instruction shifts
	MultiplyABitShift = 21 // Bit 21: accumulate bit for MLA

	// Halfword instruction shifts
	HalfwordHBitShift = 5  // Bit 5: halfword bit
	HalfwordSBitShift = 6  // Bit 6: signed bit
	HalfwordBit7      = 7  // Bit 7: always 1 for halfword
	HalfwordIBitShift = 22 // Bit 22: immediate bit for halfword
)

// Bit masks are imported from vm package
// Use vm.Mask4Bit, vm.Mask12Bit, vm.Mask16Bit, vm.Mask24Bit as needed

// Immediate Value Limits
const (
	MaxOffset12Bit     = 4095      // Maximum 12-bit offset (0xFFF)
	MaxOffsetHalfword  = 255       // Maximum 8-bit halfword offset
	MaxBranchOffsetPos = 0x7FFFFF  // Maximum positive 24-bit branch offset
	MinBranchOffsetNeg = -0x800000 // Minimum negative 24-bit branch offset
)

// Instruction Type Values (for bits shifted into position)
const (
	BranchTypeValue = 5   // Value 0b101 (5) in bits 27-25 for branch (before shift)
	LDMSTMTypeValue = 4   // Value 0b100 (4) in bits 27-25 for LDM/STM (before shift)
	MultiplyMarker  = 9   // Value 0b1001 (9) in bits 7-4 for multiply instructions
	MOVOpcodeValue  = 0xD // MOV opcode (13, 0b1101)
	SWITypeValue    = 0xF // SWI type bits (15, 0b1111)
	// Note: MOVW (ARMv6T2+) is intentionally not supported for ARM2 compatibility
)

// Word Size
const (
	WordSize = 4 // ARM instructions and words are 4 bytes
)

// Literal Pool Address Calculation
// These constants control how literal pool addresses are calculated when no explicit
// .ltorg directive is present. The assembler places literals at a 4KB-aligned boundary
// to ensure they're within the 4KB range accessible by PC-relative addressing.
const (
	LiteralPoolOffset        = 0x1000     // 4KB offset for automatic literal pool placement
	LiteralPoolAlignmentMask = 0xFFFFF000 // Mask to align addresses to 4KB boundaries (clears bottom 12 bits)
)
