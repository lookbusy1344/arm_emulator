package vm

import "fmt"

// Instruction field positions used only to name instructions.
const (
	longMultiplySignedShift = 22 // Bit 22: U (signed) in long multiply
	halfwordSignShift       = 6  // Bit 6: S (signed) in halfword transfers
	halfwordHalfShift       = 5  // Bit 5: H (halfword) in halfword transfers
)

var dataProcessingMnemonics = [...]string{
	OpAND: "AND", OpEOR: "EOR", OpSUB: "SUB", OpRSB: "RSB",
	OpADD: "ADD", OpADC: "ADC", OpSBC: "SBC", OpRSC: "RSC",
	OpTST: "TST", OpTEQ: "TEQ", OpCMP: "CMP", OpCMN: "CMN",
	OpORR: "ORR", OpMOV: "MOV", OpBIC: "BIC", OpMVN: "MVN",
}

func bitSet(opcode uint32, shift uint) bool {
	return (opcode>>shift)&Mask1Bit != 0
}

// Mnemonic names the base operation of an instruction word, without condition or
// S suffix. It follows the same decode order as VM.Decode.
func Mnemonic(opcode uint32) string {
	switch (opcode >> Bits27_26Shift) & Mask2Bit {
	case 0:
		return mnemonicGroup0(opcode)
	case 1:
		return pick(bitSet(opcode, LBitShift), "LDR", "STR") + pick(bitSet(opcode, BBitShift), "B", "")
	case 2:
		if opcode&BranchBitMask != 0 {
			return pick(bitSet(opcode, BranchLinkShift), "BL", "B")
		}
		return pick(bitSet(opcode, LBitShift), "LDM", "STM")
	default:
		if opcode&SWIDetectMask == SWIPattern {
			return "SWI"
		}
		return "UNKNOWN"
	}
}

func mnemonicGroup0(opcode uint32) string {
	switch {
	case opcode&BXPatternMask == BXEncodingBase:
		return "BX"
	case opcode&BXPatternMask == BLXEncodingBase:
		return "BLX"
	case opcode&MultiplyMask == MultiplyPattern:
		return pick(bitSet(opcode, MultiplyAShift), "MLA", "MUL")
	case opcode&LongMultiplyMask == LongMultiplyPattern:
		return pick(bitSet(opcode, longMultiplySignedShift), "S", "U") +
			pick(bitSet(opcode, MultiplyAShift), "MLAL", "MULL")
	case opcode&MRSMask == MRSPattern:
		return "MRS"
	case opcode&MSRRegMask == MSRRegPattern, opcode&MSRImmMask == MSRImmPattern:
		return "MSR"
	case !bitSet(opcode, IBitShift) && bitSet(opcode, Bit7Pos) && bitSet(opcode, Bit4Pos):
		return halfwordMnemonic(opcode)
	default:
		return dataProcessingMnemonics[(opcode>>OpcodeShift)&Mask4Bit]
	}
}

func halfwordMnemonic(opcode uint32) string {
	signed, half := bitSet(opcode, halfwordSignShift), bitSet(opcode, halfwordHalfShift)
	switch {
	case !signed && !half:
		return "SWP"
	case !signed:
		return pick(bitSet(opcode, LBitShift), "LDRH", "STRH")
	default:
		return pick(half, "LDRSH", "LDRSB")
	}
}

// checkHalfwordTransfer rejects the encodings in the halfword group that the VM does
// not execute: SWP/SWPB (S=0, H=0) and signed stores (S=1, L=0).
func checkHalfwordTransfer(opcode uint32) error {
	signed, half := bitSet(opcode, halfwordSignShift), bitSet(opcode, halfwordHalfShift)
	switch {
	case !signed && !half:
		return fmt.Errorf("SWP is not supported (opcode 0x%08X)", opcode)
	case signed && !bitSet(opcode, LBitShift):
		return fmt.Errorf("signed store is undefined (opcode 0x%08X)", opcode)
	}
	return nil
}

func pick(cond bool, ifTrue, ifFalse string) string {
	if cond {
		return ifTrue
	}
	return ifFalse
}
