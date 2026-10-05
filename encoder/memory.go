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

	// A bare label or expression is a PC-relative address: LDR Rd, label
	if !strings.HasPrefix(strings.TrimSpace(addrMode), "[") && len(inst.Operands) == 2 {
		addrMode, err = e.pcRelativeAddress(addrMode, strings.HasSuffix(mnemonic, "H"))
		if err != nil {
			return 0, err
		}
		rewritten := *inst
		rewritten.Operands = []string{inst.Operands[0], addrMode}
		inst = &rewritten
	}

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

	// W bit: pre-indexed writeback. Post-indexed always writes back; W=1 there
	// would select the user-mode (LDRT/STRT) form.
	var wBit uint32
	if writeBack && !postIndexed {
		wBit = 1
	}

	// Default: no offset, add direction
	var iBit, uBit, offsetField uint32 = 0, 1, 0

	if len(parts) > 1 {
		// Has offset
		offsetStr, subtract := splitOffsetSign(strings.Join(parts[1:], ","))
		if subtract {
			uBit = 0
		}

		// Check if it's a register or immediate
		if strings.HasPrefix(offsetStr, "#") || isNumeric(offsetStr) {
			// Immediate offset
			iBit = 0
			offset, err := e.parseImmediate(offsetStr)
			if err != nil {
				return 0, err
			}

			// 12-bit offset
			if offset > MaxOffset12Bit {
				return 0, fmt.Errorf("offset too large: %d (max %d)", offset, MaxOffset12Bit)
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
				offsetField = (shiftAmount << ShiftAmount) | (shiftType << ShiftType) | rm
			} else {
				// No shift
				offsetField = rm
			}
		}
	}

	// Format: cccc 01IP UBWL nnnn dddd oooo oooo oooo
	instruction := (cond << ConditionShift) | (1 << TypeShift26) | (iBit << TypeShift25) | (pBit << PBitShift) |
		(uBit << UBitShift) | (bBit << BBitShift) | (wBit << WBitShift) | (lBit << LBitShift) |
		(rn << RnShift) | (rd << RdShift) | offsetField

	return instruction, nil
}

// pcRelativeAddress turns a label expression into [PC, #offset] for the instruction
// being encoded. Halfword transfers reach ±255 bytes, the others ±4095.
func (e *Encoder) pcRelativeAddress(expr string, halfword bool) (string, error) {
	target, err := e.evaluateExpression(expr)
	if err != nil {
		return "", fmt.Errorf("invalid addressing mode: %s", expr)
	}
	limit := int64(MaxOffset12Bit)
	if halfword {
		limit = MaxOffsetHalfword
	}
	offset := int64(target) - int64(e.currentAddr+vm.ARMPipelineOffset)
	if offset < -limit || offset > limit {
		return "", fmt.Errorf("label %s out of range for PC-relative access: offset %d (max ±%d)", expr, offset, limit)
	}
	return fmt.Sprintf("[PC, #%d]", offset), nil
}

// splitOffsetSign separates the sign from a load/store offset. It accepts #-4, -#4,
// #+4, -4, -R2 and +R2, and returns the magnitude ("#4", "4" or "R2") and whether the
// offset subtracts. A second sign stays in the magnitude and fails to parse.
func splitOffsetSign(offset string) (magnitude string, subtract bool) {
	s := strings.TrimSpace(offset)
	hash := strings.HasPrefix(s, "#")
	s = strings.TrimSpace(strings.TrimPrefix(s, "#"))
	switch {
	case strings.HasPrefix(s, "-"):
		subtract, s = true, s[1:]
	case strings.HasPrefix(s, "+"):
		s = s[1:]
	}
	s = strings.TrimSpace(s)
	if !hash && strings.HasPrefix(s, "#") {
		hash, s = true, strings.TrimSpace(s[1:])
	}
	if hash {
		return "#" + s, subtract
	}
	return s, subtract
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
		instruction := (cond << ConditionShift) | (1 << TypeShift25) | (opMOV << OpcodeShift) | (rd << RdShift) | encoded
		return instruction, nil
	}

	// Try MVN (move not) if ~value fits
	if encoded, ok := e.encodeImmediate(^value); ok {
		// Can use MVN
		instruction := (cond << ConditionShift) | (1 << TypeShift25) | (opMVN << OpcodeShift) | (rd << RdShift) | encoded
		return instruction, nil
	}

	// Need to use literal pool - generate PC-relative LDR
	pc := e.currentAddr + vm.ARMPipelineOffset // PC = current instruction + pipeline offset

	// Reuse an existing literal with this value if this LDR can reach it
	var literalAddr uint32
	var found bool
	for addr, val := range e.LiteralPool {
		if val == value && withinLiteralReach(pc, addr) {
			literalAddr = addr
			found = true
			break
		}
	}

	if !found {
		// Find the nearest literal pool location that's within ±MaxOffset12Bit bytes
		literalAddr = e.findNearestLiteralPoolLocation(pc, value)

		if literalAddr == 0 {
			// No suitable pool found - this shouldn't happen if .ltorg is properly placed
			// Fall back to old behavior
			if e.LiteralPoolStart > 0 {
				poolSize, err := vm.SafeIntToUint32(len(e.LiteralPool) * WordSize)
				if err != nil {
					return 0, fmt.Errorf("literal pool too large: %v", err)
				}
				literalAddr = e.LiteralPoolStart + poolSize
			} else {
				poolSize, err := vm.SafeIntToUint32(len(e.LiteralPool) * WordSize)
				if err != nil {
					return 0, fmt.Errorf("literal pool too large: %v", err)
				}
				literalOffset := LiteralPoolOffset + poolSize
				literalAddr = (e.currentAddr & LiteralPoolAlignmentMask) + literalOffset
			}
		}

		// Store value in literal pool
		e.LiteralPool[literalAddr] = value
		e.pendingLiterals[value] = literalAddr
	}

	// Check addresses are in int32 range
	if literalAddr > math.MaxInt32 || pc > math.MaxInt32 {
		return 0, fmt.Errorf("address out of int32 range for PC-relative addressing")
	}
	offset := int32(literalAddr) - int32(pc) // Safe: both values checked

	// Check if offset fits in 12 bits (max MaxOffset12Bit bytes)
	absOffset := offset
	if absOffset < 0 {
		absOffset = -absOffset
	}
	if absOffset > MaxOffset12Bit {
		return 0, fmt.Errorf("literal pool offset too large: %d bytes (max %d) - literal at 0x%08X, PC=0x%08X", absOffset, MaxOffset12Bit, literalAddr, pc)
	}

	if offset < 0 {
		offset = -offset
		// Encode as LDR Rd, [PC, #-offset]
		instruction := (cond << ConditionShift) | (1 << TypeShift26) | (1 << PBitShift) | (0 << UBitShift) | (1 << LBitShift) |
			(RegisterPC << RnShift) | (rd << RdShift) | uint32(offset)
		return instruction, nil
	}

	// Encode as LDR Rd, [PC, #offset]
	instruction := (cond << ConditionShift) | (1 << TypeShift26) | (1 << PBitShift) | (1 << UBitShift) | (1 << LBitShift) |
		(RegisterPC << RnShift) | (rd << RdShift) | uint32(offset)
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

		offsetStr, subtract := splitOffsetSign(parts[1])
		if subtract {
			uBit = 0
		}
		if strings.HasPrefix(offsetStr, "#") || isNumeric(offsetStr) {
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
			offsetStr, subtract := splitOffsetSign(parts[1])
			if subtract {
				uBit = 0
			}
			if strings.HasPrefix(offsetStr, "#") || isNumeric(offsetStr) {
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

		opcode = (cond << ConditionShift) |
			(pBit << PBitShift) |
			(uBit << UBitShift) |
			(wBit << WBitShift) |
			(lBit << LBitShift) |
			(rn << RnShift) |
			(rd << RdShift) |
			(hBit << HalfwordHBitShift) |
			(sBit << HalfwordSBitShift) |
			(1 << HalfwordBit7) | (1 << HalfwordBit4) | // bits 7 and 4 mark a halfword transfer
			offset // Rm in lower 4 bits
	} else {
		// Immediate offset: split into high (bits[11:8]) and low (bits[3:0])
		if offset > MaxOffsetHalfword {
			return 0, fmt.Errorf("halfword immediate offset too large: %d (max %d)", offset, MaxOffsetHalfword)
		}

		offsetHigh := (offset >> Bit4) & vm.Mask4Bit
		offsetLow := offset & vm.Mask4Bit

		hBit := uint32(1) // H=1 for halfword
		sBit := uint32(0) // S=0 for unsigned halfword

		opcode = (cond << ConditionShift) |
			(pBit << PBitShift) |
			(uBit << UBitShift) |
			(1 << HalfwordIBitShift) | // I bit = 1 for immediate in halfword encoding
			(wBit << WBitShift) |
			(lBit << LBitShift) |
			(rn << RnShift) |
			(rd << RdShift) |
			(offsetHigh << RsShift) |
			(1 << HalfwordBit7) | (1 << HalfwordBit4) | // bits 7 and 4 mark a halfword transfer
			(hBit << HalfwordHBitShift) |
			(sBit << HalfwordSBitShift) |
			offsetLow
	}

	return opcode, nil
}

// withinLiteralReach reports whether a PC-relative LDR at pc can address addr.
func withinLiteralReach(pc, addr uint32) bool {
	if addr >= pc {
		return addr-pc <= MaxOffset12Bit
	}
	return pc-addr <= MaxOffset12Bit
}

// findNearestLiteralPoolLocation returns the next free slot in the .ltorg pool that the
// parser reserved space for this instruction: the first pool after it, or the last pool
// when none follows. It returns 0 when there are no .ltorg pools or the pool is full.
func (e *Encoder) findNearestLiteralPoolLocation(pc uint32, value uint32) uint32 {
	if len(e.LiteralPoolLocs) == 0 {
		return 0
	}

	if addr, ok := e.pendingLiterals[value]; ok {
		if withinLiteralReach(pc, addr) {
			return addr
		}
		delete(e.pendingLiterals, value)
	}

	pool := len(e.LiteralPoolLocs) - 1
	for i, poolLoc := range e.LiteralPoolLocs {
		if poolLoc > e.currentAddr {
			pool = i
			break
		}
	}

	poolLoc, capacity := e.LiteralPoolLocs[pool], e.poolCapacity(pool)
	used := e.countLiteralsAtPool(poolLoc, capacity)
	if used >= capacity {
		return 0
	}
	return poolLoc + uint32(used)*WordSize // #nosec G115 -- used < capacity, a small count
}

// poolCapacity is the number of literal slots the parser reserved for pool i.
func (e *Encoder) poolCapacity(i int) int {
	if i < len(e.LiteralPoolCounts) {
		return e.LiteralPoolCounts[i]
	}
	return parser.EstimatedLiteralsPerPool
}

// countLiteralsAtPool counts the literals already placed in a pool's reserved slots.
func (e *Encoder) countLiteralsAtPool(poolLoc uint32, capacity int) int {
	end := poolLoc + uint32(capacity)*WordSize // #nosec G115 -- capacity is a small count
	count := 0
	for addr := range e.LiteralPool {
		if addr >= poolLoc && addr < end {
			count++
		}
	}
	return count
}
