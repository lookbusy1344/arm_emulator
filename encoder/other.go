package encoder

import (
	"fmt"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// Operand counts for MUL Rd, Rm, Rs and MLA Rd, Rm, Rs, Rn.
const (
	mulOperands = 3
	mlaOperands = 4
)

// encodeMultiply encodes MUL and MLA.
//
//	cccc 0000 00AS dddd nnnn ssss 1001 mmmm
//
// ARM2 requires Rd != Rm. The product commutes, so Rd == Rm swaps Rm and Rs.
func (e *Encoder) encodeMultiply(inst *parser.Instruction, cond uint32) (uint32, error) {
	mnemonic := strings.ToUpper(inst.Mnemonic)
	accumulate := mnemonic == "MLA"
	want := mulOperands
	if accumulate {
		want = mlaOperands
	}
	if len(inst.Operands) != want {
		return 0, fmt.Errorf("%s requires %d operands, got %d", mnemonic, want, len(inst.Operands))
	}

	regs := make([]uint32, want)
	for i, op := range inst.Operands {
		reg, err := e.parseRegister(op)
		if err != nil {
			return 0, err
		}
		if reg == RegisterPC {
			return 0, fmt.Errorf("%s: PC cannot be used", mnemonic)
		}
		regs[i] = reg
	}
	rd, rm, rs := regs[0], regs[1], regs[2]
	if rd == rm {
		if rd == rs {
			return 0, fmt.Errorf("%s: Rd must differ from Rm or Rs", mnemonic)
		}
		rm, rs = rs, rm
	}

	var bits uint32
	if accumulate {
		bits |= 1<<MultiplyABitShift | regs[3]<<RdShift
	}
	if inst.SetFlags {
		bits |= 1 << SBitShift
	}
	return cond<<ConditionShift | bits | rd<<RnShift | rs<<RsShift | MultiplyMarker<<Bit4 | rm, nil
}

// encodeLoadStoreMultiple encodes LDM/STM instructions
func (e *Encoder) encodeLoadStoreMultiple(inst *parser.Instruction, cond uint32, isStore bool) (uint32, error) {
	if len(inst.Operands) < 2 {
		return 0, fmt.Errorf("%s requires at least 2 operands, got %d", inst.Mnemonic, len(inst.Operands))
	}

	// Parse base register
	baseReg := inst.Operands[0]
	writeBack := strings.HasSuffix(baseReg, "!")
	if writeBack {
		baseReg = strings.TrimSuffix(baseReg, "!")
	}

	rn, err := e.parseRegister(baseReg)
	if err != nil {
		return 0, err
	}

	// Parse register list
	regList := inst.Operands[1]
	regMask, err := e.parseRegisterList(regList)
	if err != nil {
		return 0, err
	}

	// Determine addressing mode from mnemonic
	var pBit, uBit uint32
	mnemonic := strings.ToUpper(inst.Mnemonic)

	switch {
	case strings.Contains(mnemonic, "IA"): // Increment After
		pBit, uBit = 0, 1
	case strings.Contains(mnemonic, "IB"): // Increment Before
		pBit, uBit = 1, 1
	case strings.Contains(mnemonic, "DA"): // Decrement After
		pBit, uBit = 0, 0
	case strings.Contains(mnemonic, "DB"): // Decrement Before
		pBit, uBit = 1, 0
	case strings.Contains(mnemonic, "FD"): // Full Descending (LDMFD = LDMIA, STMFD = STMDB)
		if isStore {
			pBit, uBit = 1, 0 // DB
		} else {
			pBit, uBit = 0, 1 // IA
		}
	case strings.Contains(mnemonic, "ED"): // Empty Descending
		if isStore {
			pBit, uBit = 0, 0 // DA
		} else {
			pBit, uBit = 1, 1 // IB
		}
	case strings.Contains(mnemonic, "FA"): // Full Ascending
		if isStore {
			pBit, uBit = 0, 1 // IA
		} else {
			pBit, uBit = 1, 0 // DB
		}
	case strings.Contains(mnemonic, "EA"): // Empty Ascending
		if isStore {
			pBit, uBit = 1, 1 // IB
		} else {
			pBit, uBit = 0, 0 // DA
		}
	default:
		// Default to IA
		pBit, uBit = 0, 1
	}

	// L bit: 1 for load, 0 for store
	var lBit uint32
	if !isStore {
		lBit = 1
	}

	// W bit for writeback
	var wBit uint32
	if writeBack {
		wBit = 1
	}

	// Format: cccc 100P USWL nnnn rrrr rrrr rrrr rrrr
	instruction := (cond << ConditionShift) | (LDMSTMTypeValue << TypeShift25) | (pBit << PBitShift) | (uBit << UBitShift) |
		(wBit << WBitShift) | (lBit << LBitShift) | (rn << RnShift) | regMask

	return instruction, nil
}

// encodePush encodes PUSH {reglist} as STMDB SP!, {reglist}
func (e *Encoder) encodePush(inst *parser.Instruction, cond uint32) (uint32, error) {
	if len(inst.Operands) < 1 {
		return 0, fmt.Errorf("PUSH requires 1 operand, got %d", len(inst.Operands))
	}

	regMask, err := e.parseRegisterList(inst.Operands[0])
	if err != nil {
		return 0, err
	}

	// PUSH = STMDB SP!, {reglist}
	// P=1, U=0 (decrement before), S=0, W=1 (writeback), L=0 (store)
	instruction := (cond << ConditionShift) | (LDMSTMTypeValue << TypeShift25) | (1 << PBitShift) | (0 << UBitShift) |
		(1 << WBitShift) | (0 << LBitShift) | (RegisterSP << RnShift) | regMask

	return instruction, nil
}

// encodePop encodes POP {reglist} as LDMIA SP!, {reglist}
func (e *Encoder) encodePop(inst *parser.Instruction, cond uint32) (uint32, error) {
	if len(inst.Operands) < 1 {
		return 0, fmt.Errorf("POP requires 1 operand, got %d", len(inst.Operands))
	}

	regMask, err := e.parseRegisterList(inst.Operands[0])
	if err != nil {
		return 0, err
	}

	// POP = LDMIA SP!, {reglist}
	// P=0, U=1 (increment after), S=0, W=1 (writeback), L=1 (load)
	instruction := (cond << ConditionShift) | (LDMSTMTypeValue << TypeShift25) | (0 << PBitShift) | (1 << UBitShift) |
		(1 << WBitShift) | (1 << LBitShift) | (RegisterSP << RnShift) | regMask

	return instruction, nil
}

// parseRegisterList parses a register list like {R0, R1, R2-R5, LR}
func (e *Encoder) parseRegisterList(list string) (uint32, error) {
	list = strings.TrimSpace(list)
	list = strings.TrimPrefix(list, "{")
	list = strings.TrimSuffix(list, "}")

	var mask uint32

	parts := strings.Split(list, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check for range (R2-R5)
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return 0, fmt.Errorf("invalid register range: %s", part)
			}

			start, err := e.parseRegister(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return 0, err
			}

			end, err := e.parseRegister(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return 0, err
			}

			if start > end {
				return 0, fmt.Errorf("invalid register range: %s (start > end)", part)
			}

			for r := start; r <= end; r++ {
				mask |= (1 << r)
			}
		} else {
			// Single register
			reg, err := e.parseRegister(part)
			if err != nil {
				return 0, err
			}
			mask |= (1 << reg)
		}
	}

	return mask, nil
}

// encodeNOP encodes NOP as MOV R0, R0
func (e *Encoder) encodeNOP(cond uint32) uint32 {
	// NOP = MOV R0, R0
	// Format: cccc 0001 101S dddd 0000 0000 0000 mmmm
	// MOV (S=0, Rd=0, Rm=0)
	instruction := (cond << ConditionShift) | (MOVOpcodeValue << OpcodeShift) | (0 << RnShift) | 0
	return instruction
}

// encodeSWI encodes software interrupt instruction
// ARM SWI instruction format (32 bits):
//
//	Bits 31-28: Condition code (cccc)
//	Bits 27-24: 1111 (SWI opcode identifier - SWITypeValue=0xF shifted by 24)
//	Bits 23-0:  24-bit immediate value (syscall number)
//
// Example: SWI #0x01 with condition AL (0xE) encodes as 0xEF000001
func (e *Encoder) encodeSWI(inst *parser.Instruction, cond uint32) (uint32, error) {
	if len(inst.Operands) < 1 {
		return 0, fmt.Errorf("SWI requires 1 operand, got %d", len(inst.Operands))
	}

	// Parse 24-bit immediate value
	imm, err := e.parseImmediate(inst.Operands[0])
	if err != nil {
		return 0, err
	}

	// Check if it fits in 24 bits
	if imm > vm.Mask24Bit {
		return 0, fmt.Errorf("SWI immediate too large: 0x%X (max 0x%X)", imm, vm.Mask24Bit)
	}

	// Build instruction: condition in bits 31-28, 0xF in bits 27-24, immediate in bits 23-0
	// SWITypeValue (0xF) << PBitShift (24) = 0x0F000000 (bits 27-24 = 1111)
	instruction := (cond << ConditionShift) | (SWITypeValue << PBitShift) | imm

	return instruction, nil
}

// Long multiply bits (bits 22 and 21)
const (
	longMultiplySignedShift = 22 // 1 = SMULL/SMLAL, 0 = UMULL/UMLAL
	longMultiplyAccumShift  = 21 // 1 = xMLAL
	longMultiplyOperands    = 4
)

// encodeMultiplyLong encodes UMULL, UMLAL, SMULL and SMLAL.
//
//	xxxLL{cond}{S} RdLo, RdHi, Rm, Rs    cccc 0000 1UAS hhhh llll ssss 1001 mmmm
func (e *Encoder) encodeMultiplyLong(inst *parser.Instruction, cond uint32) (uint32, error) {
	mnemonic := strings.ToUpper(inst.Mnemonic)
	if len(inst.Operands) != longMultiplyOperands {
		return 0, fmt.Errorf("%s requires %d operands, got %d", mnemonic, longMultiplyOperands, len(inst.Operands))
	}

	var regs [longMultiplyOperands]uint32
	for i, op := range inst.Operands {
		reg, err := e.parseRegister(op)
		if err != nil {
			return 0, err
		}
		if reg == RegisterPC {
			return 0, fmt.Errorf("%s: PC cannot be used", mnemonic)
		}
		regs[i] = reg
	}
	rdLo, rdHi, rm, rs := regs[0], regs[1], regs[2], regs[3]
	if rdLo == rdHi {
		return 0, fmt.Errorf("%s: RdLo and RdHi must be different registers", mnemonic)
	}
	if rdLo == rm || rdHi == rm {
		return 0, fmt.Errorf("%s: RdLo and RdHi must differ from Rm", mnemonic)
	}

	var bits uint32
	if strings.HasPrefix(mnemonic, "S") {
		bits |= 1 << longMultiplySignedShift
	}
	if strings.HasSuffix(mnemonic, "MLAL") {
		bits |= 1 << longMultiplyAccumShift
	}
	if inst.SetFlags {
		bits |= 1 << SBitShift
	}

	return cond<<ConditionShift | vm.LongMultiplyPattern | bits |
		rdHi<<RnShift | rdLo<<RdShift | rs<<RsShift | rm, nil
}
