package encoder

import (
	"fmt"
	"math"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// encodeBranch encodes B, BL, and BX instructions
func (e *Encoder) encodeBranch(inst *parser.Instruction, cond uint32) (uint32, error) {
	if len(inst.Operands) < 1 {
		return 0, fmt.Errorf("%s requires 1 operand, got %d", inst.Mnemonic, len(inst.Operands))
	}

	mnemonic := strings.ToUpper(inst.Mnemonic)

	// BX is different - it's encoded as a special data processing instruction
	if mnemonic == "BX" {
		return e.encodeBX(inst, cond)
	}

	// Get target address
	target := strings.TrimSpace(inst.Operands[0])

	var targetAddr uint32
	var err error

	// Try to resolve as symbol/label
	if sym, exists := e.symbolTable.Lookup(target); exists && sym.Defined {
		targetAddr = sym.Value
	} else {
		// Try to parse as immediate address
		targetAddr, err = e.parseImmediate(target)
		if err != nil {
			return 0, fmt.Errorf("undefined label or invalid address: %s", target)
		}
	}

	// Calculate offset: (target - PC - 8) / 4
	// PC is current instruction address + 8 (ARM pipeline)
	pc := e.currentAddr + 8
	// Ensure both targetAddr and pc are safely convertible to int32
	if targetAddr > math.MaxInt32 {
		return 0, fmt.Errorf("branch target address out of int32 range: 0x%X", targetAddr)
	}
	if pc > math.MaxInt32 {
		return 0, fmt.Errorf("PC out of int32 range: 0x%X", pc)
	}
	offset := int32(targetAddr) - int32(pc) // Safe: both values checked

	// Check if offset is word-aligned
	if offset&0x3 != 0 {
		return 0, fmt.Errorf("branch target not word-aligned: offset=%d", offset)
	}

	// Divide by 4 to get word offset
	wordOffset := offset / 4

	// Check if offset fits in 24 bits (signed)
	if wordOffset < -0x800000 || wordOffset > 0x7FFFFF {
		return 0, fmt.Errorf("branch offset out of range: %d (max ±32MB)", offset)
	}

	// Encode 24-bit offset (sign-extended)
	// Safe: wordOffset bounds checked above to be in range [-0x800000, 0x7FFFFF]
	// Intentional conversion for bit pattern encoding
	encodedOffset := uint32(wordOffset) & 0xFFFFFF // #nosec G115 -- bounds checked, intentional bit encoding

	// L bit: 1 for BL (link), 0 for B
	var lBit uint32
	if mnemonic == "BL" {
		lBit = 1
	}

	// Format: cccc 101L oooo oooo oooo oooo oooo oooo
	instruction := (cond << 28) | (5 << 25) | (lBit << 24) | encodedOffset

	return instruction, nil
}

// encodeBX encodes BX (Branch and Exchange) instruction
func (e *Encoder) encodeBX(inst *parser.Instruction, cond uint32) (uint32, error) {
	if len(inst.Operands) < 1 {
		return 0, fmt.Errorf("BX requires 1 operand, got %d", len(inst.Operands))
	}

	// Parse register
	rm, err := e.parseRegister(inst.Operands[0])
	if err != nil {
		return 0, err
	}

	// BX format: cccc 0001 0010 1111 1111 1111 0001 mmmm
	instruction := (cond << 28) | (0x12FFF1 << 4) | rm

	return instruction, nil
}
