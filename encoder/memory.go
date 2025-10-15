package encoder

import (
	"fmt"
	"math"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
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

	// Check for post-indexed addressing: [Rn], offset
	// Parser splits this into two operands: "[Rn]" and "offset"
	if len(inst.Operands) > 2 && strings.HasSuffix(addrMode, "]") && !strings.HasSuffix(addrMode, "]!") {
		// Combine the bracket part with the offset: "[Rn]" + "," + "#offset"
		addrMode = addrMode + "," + inst.Operands[2]
	}

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
	postIndexed := strings.Contains(addrMode, "],")

	// Check for pre-indexed with writeback: [Rn, offset]!
	writeBack := strings.HasSuffix(addrMode, "]!")
	if writeBack {
		addrMode = strings.TrimSuffix(addrMode, "!")
	}

	// Remove brackets and split
	var parts []string
	if postIndexed {
		// Post-indexed: "[Rn],offset" → split on "]," then clean up
		addrMode = strings.TrimPrefix(addrMode, "[")
		parts = strings.Split(addrMode, "],")
		// parts[0] is "Rn", parts[1] is "offset"
	} else {
		// Pre-indexed or offset: "[Rn,offset]" or "[Rn]"
		addrMode = strings.TrimPrefix(addrMode, "[")
		addrMode = strings.TrimSuffix(addrMode, "]")
		parts = strings.Split(addrMode, ",")
	}
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

	// Evaluate the expression (handles both simple symbols and expressions like "label+12")
	value, err = e.evaluateExpression(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid pseudo-instruction value '%s': %w", valueStr, err)
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
	// Check if this value already exists in the literal pool (deduplication)
	var literalAddr uint32
	var found bool
	for addr, val := range e.LiteralPool {
		if val == value {
			literalAddr = addr
			found = true
			break
		}
	}

	if !found {
		// Find the nearest literal pool location that's within ±4095 bytes
		pc := e.currentAddr + 8 // PC = current instruction + 8
		literalAddr = e.findNearestLiteralPoolLocation(pc, value)

		if literalAddr == 0 {
			// No suitable pool found - this shouldn't happen if .ltorg is properly placed
			// Fall back to old behavior
			if e.LiteralPoolStart > 0 {
				poolSize, err := vm.SafeIntToUint32(len(e.LiteralPool) * 4)
				if err != nil {
					return 0, fmt.Errorf("literal pool too large: %v", err)
				}
				literalAddr = e.LiteralPoolStart + poolSize
			} else {
				poolSize, err := vm.SafeIntToUint32(len(e.LiteralPool) * 4)
				if err != nil {
					return 0, fmt.Errorf("literal pool too large: %v", err)
				}
				literalOffset := 0x1000 + poolSize
				literalAddr = (e.currentAddr & 0xFFFFF000) + literalOffset
			}
		}

		// Store value in literal pool
		e.LiteralPool[literalAddr] = value
		e.pendingLiterals[value] = literalAddr
	}

	// Calculate PC-relative offset
	// PC = current instruction + 8
	pc := e.currentAddr + 8
	// Check addresses are in int32 range
	if literalAddr > math.MaxInt32 || pc > math.MaxInt32 {
		return 0, fmt.Errorf("address out of int32 range for PC-relative addressing")
	}
	offset := int32(literalAddr) - int32(pc) // Safe: both values checked

	// Check if offset fits in 12 bits (max 4095 bytes)
	absOffset := offset
	if absOffset < 0 {
		absOffset = -absOffset
	}
	if absOffset > 4095 {
		return 0, fmt.Errorf("literal pool offset too large: %d bytes (max 4095) - literal at 0x%08X, PC=0x%08X", absOffset, literalAddr, pc)
	}

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

// encodeMemoryHalfword encodes halfword load/store (LDRH/STRH)
// ARM halfword format: cond 000P UBWL Rn Rd offsetH 1SH1 offsetL
// P=1 for pre-indexed, P=0 for post-indexed
// U=1 for add offset, U=0 for subtract offset
// B=0 for halfword (always 0 for LDRH/STRH)
// W=1 for writeback (pre-indexed only)
// L=1 for load, L=0 for store
// S=0, H=1 for unsigned halfword (bits[6:5] = 01)
// bits[7:4] = 1011 for load, 1001 for store (actually controlled by S,H,L bits)
func (e *Encoder) encodeMemoryHalfword(inst *parser.Instruction, cond, rd, lBit uint32) (uint32, error) {
	if len(inst.Operands) < 2 {
		return 0, fmt.Errorf("halfword instruction requires at least 2 operands")
	}

	// Parse addressing mode
	addrMode := inst.Operands[1]

	// Check for post-indexed: combine with third operand if present
	if len(inst.Operands) > 2 && strings.HasSuffix(addrMode, "]") && !strings.HasSuffix(addrMode, "]!") {
		addrMode = addrMode + "," + inst.Operands[2]
	}

	addrMode = strings.TrimSpace(addrMode)

	// Check for post-indexed: [Rn], offset
	postIndexed := strings.Contains(addrMode, "],")

	// Check for pre-indexed with writeback: [Rn, offset]!
	writeBack := strings.HasSuffix(addrMode, "]!")
	if writeBack {
		addrMode = strings.TrimSuffix(addrMode, "!")
	}

	// Extract base register and offset
	if !strings.HasPrefix(addrMode, "[") {
		return 0, fmt.Errorf("invalid addressing mode for halfword: %s", addrMode)
	}

	addrMode = strings.TrimPrefix(addrMode, "[")
	addrMode = strings.TrimSuffix(addrMode, "]")

	var rn uint32
	var offset uint32
	var uBit uint32 = 1 // Default to add
	var isRegisterOffset bool

	if postIndexed {
		// Post-indexed: [Rn], offset
		parts := strings.Split(addrMode, "],")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid post-indexed addressing for halfword")
		}

		rnReg, err := e.parseRegister(strings.TrimSpace(parts[0]))
		if err != nil {
			return 0, err
		}
		rn = rnReg

		offsetStr := strings.TrimSpace(parts[1])
		if strings.HasPrefix(offsetStr, "#") || isNumeric(offsetStr) {
			// Check if offset is negative
			if strings.HasPrefix(offsetStr, "-") {
				uBit = 0 // Subtract
				offsetStr = strings.TrimPrefix(offsetStr, "-")
			}
			offsetVal, err := e.parseImmediate(offsetStr)
			if err != nil {
				return 0, err
			}
			offset = offsetVal
		} else {
			// Register offset
			offsetReg, err := e.parseRegister(offsetStr)
			if err != nil {
				return 0, err
			}
			offset = offsetReg
			isRegisterOffset = true
		}
	} else {
		// Pre-indexed or simple: [Rn] or [Rn, offset]
		parts := strings.Split(addrMode, ",")

		rnReg, err := e.parseRegister(strings.TrimSpace(parts[0]))
		if err != nil {
			return 0, err
		}
		rn = rnReg

		if len(parts) > 1 {
			offsetStr := strings.TrimSpace(parts[1])
			if strings.HasPrefix(offsetStr, "#") || isNumeric(offsetStr) {
				// Check if offset is negative
				if strings.HasPrefix(offsetStr, "-") {
					uBit = 0 // Subtract
					offsetStr = strings.TrimPrefix(offsetStr, "-")
				}
				offsetVal, err := e.parseImmediate(offsetStr)
				if err != nil {
					return 0, err
				}
				offset = offsetVal
			} else {
				// Register offset
				offsetReg, err := e.parseRegister(offsetStr)
				if err != nil {
					return 0, err
				}
				offset = offsetReg
				isRegisterOffset = true
			}
		}
	}

	// Build the instruction
	// Format: cond 000P UBWL Rn Rd offsetH 1SH1 offsetL
	// For LDRH/STRH: S=0, H=1 (bits[6:5] = 01), so bits[7:4] = ?0?1
	// Combined with bits[7:4], we get 1011 for load, 1001 for store

	pBit := uint32(1) // Pre-indexed by default
	if postIndexed {
		pBit = 0
	}

	wBit := uint32(0)
	if writeBack {
		wBit = 1
	}

	var opcode uint32

	if isRegisterOffset {
		// Register offset: bits[7:4] = 1001 for store, 1011 for load
		// Rm in bits[3:0], offset high bits in [11:8]
		hBit := uint32(1) // H=1 for halfword
		sBit := uint32(0) // S=0 for unsigned halfword

		opcode = (cond << 28) |
			(pBit << 24) |
			(uBit << 23) |
			(wBit << 21) |
			(lBit << 20) |
			(rn << 16) |
			(rd << 12) |
			(hBit << 5) |
			(sBit << 6) |
			(1 << 7) | // Always 1 for halfword
			offset // Rm in lower 4 bits
	} else {
		// Immediate offset: split into high (bits[11:8]) and low (bits[3:0])
		if offset > 255 {
			return 0, fmt.Errorf("halfword immediate offset too large: %d (max 255)", offset)
		}

		offsetHigh := (offset >> 4) & 0xF
		offsetLow := offset & 0xF

		hBit := uint32(1) // H=1 for halfword
		sBit := uint32(0) // S=0 for unsigned halfword

		opcode = (cond << 28) |
			(pBit << 24) |
			(uBit << 23) |
			(1 << 22) | // I bit = 1 for immediate in halfword encoding
			(wBit << 21) |
			(lBit << 20) |
			(rn << 16) |
			(rd << 12) |
			(offsetHigh << 8) |
			(1 << 7) | // Always 1 for halfword misc
			(hBit << 5) |
			(sBit << 6) |
			offsetLow
	}

	return opcode, nil
}

// findNearestLiteralPoolLocation finds the nearest literal pool location within ±4095 bytes
// Returns 0 if no suitable location is found
func (e *Encoder) findNearestLiteralPoolLocation(pc uint32, value uint32) uint32 {
	// If no .ltorg directives specified, return 0 to use fallback behavior
	if len(e.LiteralPoolLocs) == 0 {
		return 0
	}

	// Check if this value already has a pending location
	if addr, ok := e.pendingLiterals[value]; ok {
		// Verify it's still within range
		if addr > pc {
			if addr-pc > 4095 {
				// Out of range, need to find a new location
				delete(e.pendingLiterals, value)
			} else {
				return addr
			}
		} else {
			if pc-addr > 4095 {
				// Out of range, need to find a new location
				delete(e.pendingLiterals, value)
			} else {
				return addr
			}
		}
	}

	// Find nearest pool location within ±4095 bytes
	var bestAddr uint32
	var bestDistance uint32 = 0xFFFFFFFF

	for _, poolLoc := range e.LiteralPoolLocs {
		var distance uint32
		if poolLoc > pc {
			distance = poolLoc - pc
			if distance <= 4095 && distance < bestDistance {
				// Count how many literals are already assigned to this pool
				literalsAtPool := e.countLiteralsAtPool(poolLoc)
				// Calculate where this literal would go
				candidateAddr := poolLoc + uint32(literalsAtPool*4)
				// Check if it's still within range from PC
				if candidateAddr > pc && candidateAddr-pc <= 4095 {
					bestAddr = candidateAddr
					bestDistance = distance
				}
			}
		} else {
			distance = pc - poolLoc
			if distance <= 4095 && distance < bestDistance {
				// For backward references, we need to be more careful
				// Count existing literals
				literalsAtPool := e.countLiteralsAtPool(poolLoc)
				candidateAddr := poolLoc + uint32(literalsAtPool*4)
				// Check distance from PC to candidate address
				if candidateAddr <= pc && pc-candidateAddr <= 4095 {
					bestAddr = candidateAddr
					bestDistance = distance
				}
			}
		}
	}

	return bestAddr
}

// countLiteralsAtPool counts how many literals are already assigned to start at or near a pool location
func (e *Encoder) countLiteralsAtPool(poolLoc uint32) int {
	count := 0
	// Check all assigned literals to see how many are in this pool region
	// Literals within 1024 bytes of the pool location are considered part of the same pool
	for addr := range e.LiteralPool {
		if addr >= poolLoc && addr < poolLoc+1024 {
			count++
		}
	}
	return count
}
