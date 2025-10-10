package encoder

import (
	"fmt"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// encodeMemory encodes LDR, STR, LDRB, STRB, LDRH, STRH instructions
func (e *Encoder) encodeMemory(inst *parser.Instruction, cond uint32) (uint32, error) {
	if len(inst.Operands) < 2 {
		return 0, fmt.Errorf("%s requires at least 2 operands, got %d (operands: %v)", inst.Mnemonic, len(inst.Operands), inst.Operands)
	}

	mnemonic := strings.ToUpper(inst.Mnemonic)

	// Parse destination/source register
	rd, err := e.parseRegister(inst.Operands[0])
	if err != nil {
		return 0, err
	}

	// Check for pseudo-instruction: LDR Rd, =value or =label
	// The parser might give us "=" and "label" as separate operands or "=label" as one
	if strings.HasPrefix(inst.Operands[1], "=") {
		return e.encodeLDRPseudo(inst, cond, rd)
	}

	// Check if operand is just "=" and the label is in the next operand
	if inst.Operands[1] == "=" && len(inst.Operands) > 2 {
		// Combine them
		combined := "=" + inst.Operands[2]
		// Create a temporary instruction with combined operand
		tempInst := *inst
		tempInst.Operands = []string{inst.Operands[0], combined}
		return e.encodeLDRPseudo(&tempInst, cond, rd)
	}

	// Parse addressing mode
	addrMode := inst.Operands[1]

	// Determine L bit (1 for load, 0 for store)
	var lBit uint32
	if strings.HasPrefix(mnemonic, "LDR") {
		lBit = 1
	}

	// Determine B bit (1 for byte, 0 for word)
	var bBit uint32
	if strings.HasSuffix(mnemonic, "B") {
		bBit = 1
	}

	// For halfword, use different encoding (not implemented yet for simplicity)
	if strings.HasSuffix(mnemonic, "H") {
		return e.encodeMemoryHalfword(inst, cond, rd, lBit)
	}

	// Parse addressing mode
	return e.encodeAddressingMode(cond, lBit, bBit, rd, addrMode)
}

// encodeAddressingMode parses and encodes various addressing modes
func (e *Encoder) encodeAddressingMode(cond, lBit, bBit, rd uint32, addrMode string) (uint32, error) {
	addrMode = strings.TrimSpace(addrMode)

	// Check format: [Rn] or [Rn, offset] or [Rn, offset]! or [Rn], offset
	if !strings.HasPrefix(addrMode, "[") {
		return 0, fmt.Errorf("invalid addressing mode: %s", addrMode)
	}

	// Check for post-indexed: [Rn], offset
	postIndexed := strings.Contains(addrMode, "]") && !strings.HasSuffix(addrMode, "]") && !strings.HasSuffix(addrMode, "]!")

	// Check for pre-indexed with writeback: [Rn, offset]!
	writeBack := strings.HasSuffix(addrMode, "]!")
	if writeBack {
		addrMode = strings.TrimSuffix(addrMode, "!")
	}

	// Remove brackets
	addrMode = strings.TrimPrefix(addrMode, "[")
	addrMode = strings.TrimSuffix(addrMode, "]")

	// Split into base register and offset
	parts := strings.Split(addrMode, ",")
	rn, err := e.parseRegister(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, err
	}

	// P bit: 1 for pre-indexed (or offset), 0 for post-indexed
	var pBit uint32 = 1
	if postIndexed {
		pBit = 0
	}

	// W bit: 1 for writeback (also set for post-indexed)
	var wBit uint32
	if writeBack || postIndexed {
		wBit = 1
	}

	// Default: no offset, add direction
	var iBit, uBit, offsetField uint32 = 0, 1, 0

	if len(parts) > 1 {
		// Has offset
		offsetStr := strings.TrimSpace(strings.Join(parts[1:], ","))

		// Check if offset is negative
		uBit = 1 // Default: add
		if strings.HasPrefix(offsetStr, "-") {
			uBit = 0 // Subtract
			offsetStr = strings.TrimPrefix(offsetStr, "-")
		} else {
			offsetStr = strings.TrimPrefix(offsetStr, "+") // Remove optional +
		}

		offsetStr = strings.TrimSpace(offsetStr)

		// Check if it's a register or immediate
		if strings.HasPrefix(offsetStr, "#") || isNumeric(offsetStr) {
			// Immediate offset
			iBit = 0
			offset, err := e.parseImmediate(offsetStr)
			if err != nil {
				return 0, err
			}

			// 12-bit offset
			if offset > 0xFFF {
				return 0, fmt.Errorf("offset too large: %d (max 4095)", offset)
			}
			offsetField = offset

		} else {
			// Register offset (with optional shift)
			iBit = 1
			regParts := strings.Split(offsetStr, ",")
			rm, err := e.parseRegister(strings.TrimSpace(regParts[0]))
			if err != nil {
				return 0, err
			}

			if len(regParts) > 1 {
				// Has shift
				shiftStr := strings.TrimSpace(strings.Join(regParts[1:], ","))
				shiftType, shiftAmount, _, err := e.parseShift(shiftStr)
				if err != nil {
					return 0, err
				}
				offsetField = (shiftAmount << 7) | (shiftType << 5) | rm
			} else {
				// No shift
				offsetField = rm
			}
		}
	}

	// Format: cccc 01IP UBWL nnnn dddd oooo oooo oooo
	instruction := (cond << 28) | (1 << 26) | (iBit << 25) | (pBit << 24) |
		(uBit << 23) | (bBit << 22) | (wBit << 21) | (lBit << 20) |
		(rn << 16) | (rd << 12) | offsetField

	return instruction, nil
}

// encodeLDRPseudo encodes LDR Rd, =value or =label (pseudo-instruction)
func (e *Encoder) encodeLDRPseudo(inst *parser.Instruction, cond, rd uint32) (uint32, error) {
	operand := strings.TrimSpace(inst.Operands[1])
	valueStr := strings.TrimPrefix(operand, "=")
	valueStr = strings.TrimSpace(valueStr)

	var value uint32
	var err error

	if valueStr == "" {
		return 0, fmt.Errorf("empty pseudo-instruction value in operand: '%s'", inst.Operands[1])
	}

	// Try to resolve as symbol
	if sym, exists := e.symbolTable.Lookup(valueStr); exists && sym.Defined {
		value = sym.Value
	} else {
		// Parse as immediate
		value, err = e.parseImmediate(valueStr)
		if err != nil {
			return 0, fmt.Errorf("invalid pseudo-instruction value '%s': %w", valueStr, err)
		}
	}

	// Try to encode as MOV Rd, #value if it fits
	if encoded, ok := e.encodeImmediate(value); ok {
		// Can use MOV
		instruction := (cond << 28) | (1 << 25) | (opMOV << 21) | (rd << 12) | encoded
		return instruction, nil
	}

	// Try MVN (move not) if ~value fits
	if encoded, ok := e.encodeImmediate(^value); ok {
		// Can use MVN
		instruction := (cond << 28) | (1 << 25) | (opMVN << 21) | (rd << 12) | encoded
		return instruction, nil
	}

	// Need to use literal pool - generate PC-relative LDR
	// Place literals in a pool at a fixed offset to avoid overwriting instructions
	// Use a large offset (4KB) to ensure it's after all code and data
	literalOffset := uint32(0x1000 + (len(e.LiteralPool) * 4))
	literalAddr := (e.currentAddr & 0xFFFFF000) + literalOffset

	// Store value in literal pool
	e.LiteralPool[literalAddr] = value

	// Calculate PC-relative offset
	// PC = current instruction + 8
	pc := e.currentAddr + 8
	offset := int32(literalAddr) - int32(pc)

	if offset < 0 {
		offset = -offset
		// Encode as LDR Rd, [PC, #-offset]
		instruction := (cond << 28) | (1 << 26) | (1 << 24) | (0 << 23) | (1 << 20) |
			(15 << 16) | (rd << 12) | uint32(offset)
		return instruction, nil
	}

	// Encode as LDR Rd, [PC, #offset]
	instruction := (cond << 28) | (1 << 26) | (1 << 24) | (1 << 23) | (1 << 20) |
		(15 << 16) | (rd << 12) | uint32(offset)
	return instruction, nil
}

// encodeMemoryHalfword encodes halfword load/store (simplified)
func (e *Encoder) encodeMemoryHalfword(inst *parser.Instruction, cond, rd, lBit uint32) (uint32, error) {
	// Simplified halfword encoding
	// Format is different from word/byte
	// For now, return an error as this is complex
	return 0, fmt.Errorf("halfword operations not fully implemented yet")
}
