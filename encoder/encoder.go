package encoder

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// Encoder converts parsed instructions into ARM machine code
type Encoder struct {
	symbolTable       *parser.SymbolTable
	currentAddr       uint32
	LiteralPool       map[uint32]uint32 // address -> value for literal pool (exported)
	LiteralPoolStart  uint32            // Start address for literal pool (set externally)
	LiteralPoolLocs   []uint32          // Addresses of .ltorg directives (multiple pools)
	LiteralPoolCounts []int             // Expected literal counts for each pool (from parser)
	pendingLiterals   map[uint32]uint32 // value -> preferred address mapping for dedup
	PoolWarnings      []string          // Warnings about pool capacity issues
}

// NewEncoder creates a new encoder instance
func NewEncoder(symbolTable *parser.SymbolTable) *Encoder {
	return &Encoder{
		symbolTable:       symbolTable,
		LiteralPool:       make(map[uint32]uint32),
		LiteralPoolLocs:   make([]uint32, 0),
		LiteralPoolCounts: make([]int, 0),
		pendingLiterals:   make(map[uint32]uint32),
		PoolWarnings:      make([]string, 0),
	}
}

// EncodeInstruction converts a single parsed instruction into ARM machine code
func (e *Encoder) EncodeInstruction(inst *parser.Instruction, address uint32) (uint32, error) {
	e.currentAddr = address

	// Get condition code (default to AL if not specified)
	cond := e.encodeCondition(inst.Condition)

	mnemonic := strings.ToUpper(inst.Mnemonic)

	// Route to appropriate encoder based on instruction type
	switch mnemonic {
	// Data processing instructions
	case "MOV", "MVN":
		return e.encodeDataProcessingMove(inst, cond)
	case "ADD", "ADC", "SUB", "SBC", "RSB", "RSC":
		return e.encodeDataProcessingArithmetic(inst, cond)
	case "AND", "ORR", "EOR", "BIC":
		return e.encodeDataProcessingLogical(inst, cond)
	case "CMP", "CMN", "TST", "TEQ":
		return e.encodeDataProcessingCompare(inst, cond)

	// Memory instructions
	case "LDR", "STR", "LDRB", "STRB", "LDRH", "STRH":
		return e.encodeMemory(inst, cond)

	// Branch instructions
	case "B", "BL", "BX", "BLX":
		return e.encodeBranch(inst, cond)

	// Multiply instructions
	case "MUL", "MLA":
		return e.encodeMultiply(inst, cond)

	// Load/Store multiple
	case "LDM", "STM", "LDMIA", "LDMIB", "LDMDA", "LDMDB":
		return e.encodeLoadStoreMultiple(inst, cond, false)
	case "STMIA", "STMIB", "STMDA", "STMDB":
		return e.encodeLoadStoreMultiple(inst, cond, true)
	// Load/Store multiple aliases (FD=Full Descending, FA=Full Ascending, EA=Empty Ascending, ED=Empty Descending)
	case "LDMFD", "LDMFA", "LDMEA", "LDMED":
		return e.encodeLoadStoreMultiple(inst, cond, false)
	case "STMFD", "STMFA", "STMEA", "STMED":
		return e.encodeLoadStoreMultiple(inst, cond, true)
	case "PUSH":
		return e.encodePush(inst, cond)
	case "POP":
		return e.encodePop(inst, cond)
	case "NOP":
		return e.encodeNOP(cond), nil

	// Software interrupt
	case "SWI", "SVC": // SVC is ARM7+ name for SWI
		return e.encodeSWI(inst, cond)

	// ADR pseudo-instruction
	case "ADR":
		return e.encodeADR(inst, cond)

	default:
		return 0, fmt.Errorf("unknown instruction: %s", mnemonic)
	}
}

// encodeCondition converts condition string to 4-bit code
func (e *Encoder) encodeCondition(cond string) uint32 {
	switch strings.ToUpper(cond) {
	case "EQ":
		return uint32(vm.CondEQ)
	case "NE":
		return uint32(vm.CondNE)
	case "CS", "HS":
		return uint32(vm.CondCS)
	case "CC", "LO":
		return uint32(vm.CondCC)
	case "MI":
		return uint32(vm.CondMI)
	case "PL":
		return uint32(vm.CondPL)
	case "VS":
		return uint32(vm.CondVS)
	case "VC":
		return uint32(vm.CondVC)
	case "HI":
		return uint32(vm.CondHI)
	case "LS":
		return uint32(vm.CondLS)
	case "GE":
		return uint32(vm.CondGE)
	case "LT":
		return uint32(vm.CondLT)
	case "GT":
		return uint32(vm.CondGT)
	case "LE":
		return uint32(vm.CondLE)
	case "", "AL":
		return uint32(vm.CondAL)
	default:
		return uint32(vm.CondAL) // Default to always
	}
}

// parseRegister parses a register name and returns its number
func (e *Encoder) parseRegister(reg string) (uint32, error) {
	reg = strings.ToUpper(strings.TrimSpace(reg))

	// Handle special register names
	switch reg {
	case "SP", "R13":
		return 13, nil
	case "LR", "R14":
		return 14, nil
	case "PC", "R15":
		return 15, nil
	}

	// Parse Rn format
	if strings.HasPrefix(reg, "R") {
		numStr := reg[1:]
		num, err := strconv.ParseUint(numStr, 10, 32)
		if err != nil || num > 15 {
			return 0, fmt.Errorf("invalid register: %s", reg)
		}
		return uint32(num), nil
	}

	return 0, fmt.Errorf("invalid register: %s", reg)
}

// parseImmediate parses an immediate value
func (e *Encoder) parseImmediate(imm string) (uint32, error) {
	imm = strings.TrimSpace(imm)

	if imm == "" {
		return 0, fmt.Errorf("empty immediate value")
	}

	// Remove leading # if present
	imm = strings.TrimPrefix(imm, "#")

	// Handle character literals like 'A' or ' ' or '\t' or '\x41'
	if strings.HasPrefix(imm, "'") && strings.HasSuffix(imm, "'") && len(imm) >= 3 {
		charLiteral := imm[1 : len(imm)-1] // Remove quotes

		// Handle escape sequences using shared parser utility
		if strings.HasPrefix(charLiteral, "\\") {
			b, consumed, err := parser.ParseEscapeChar(charLiteral)
			if err != nil {
				return 0, fmt.Errorf("invalid escape sequence in character literal: %s", imm)
			}
			// Ensure the entire escape was consumed
			if consumed != len(charLiteral) {
				return 0, fmt.Errorf("invalid character literal: %s", imm)
			}
			return uint32(b), nil
		}

		// Regular character literal
		if len(charLiteral) != 1 {
			return 0, fmt.Errorf("character literal must contain exactly one character: %s", imm)
		}
		return uint32(charLiteral[0]), nil
	}

	// Handle negative numbers
	negative := false
	if strings.HasPrefix(imm, "-") {
		negative = true
		imm = imm[1:]
	}

	// Try to parse as symbol first
	if !strings.HasPrefix(imm, "0x") && !strings.HasPrefix(imm, "0X") {
		if sym, exists := e.symbolTable.Lookup(imm); exists && sym.Defined {
			return sym.Value, nil
		}
	}

	var value uint64
	var err error

	// Parse based on prefix
	if strings.HasPrefix(imm, "0x") || strings.HasPrefix(imm, "0X") {
		value, err = strconv.ParseUint(imm[2:], 16, 32)
	} else if strings.HasPrefix(imm, "0b") || strings.HasPrefix(imm, "0B") {
		value, err = strconv.ParseUint(imm[2:], 2, 32)
	} else if strings.HasPrefix(imm, "0") && len(imm) > 1 {
		value, err = strconv.ParseUint(imm[1:], 8, 32)
	} else {
		value, err = strconv.ParseUint(imm, 10, 32)
	}

	if err != nil {
		return 0, fmt.Errorf("invalid immediate value: %s", imm)
	}

	result := uint32(value)
	if negative {
		// Bounds check before casting to int32 and negating
		if result < 1 || result > uint32(math.MaxInt32)+1 {
			return 0, fmt.Errorf("immediate value out of valid signed 32-bit range: %s", imm)
		}
		// Safe: value checked to be in valid range for signed negation
		result = uint32(-int32(result)) // #nosec G115 -- bounds checked above
	}

	return result, nil
}

// encodeImmediate encodes an 8-bit immediate value with 4-bit rotation
// Returns encoded value and success flag
func (e *Encoder) encodeImmediate(value uint32) (uint32, bool) {
	// Try each rotation (0, 2, 4, 6, ..., 30)
	for rotate := uint32(0); rotate < 32; rotate += 2 {
		// Rotate right by rotate bits
		rotated := (value >> rotate) | (value << (32 - rotate))

		// Check if it fits in 8 bits
		if rotated <= 0xFF {
			// Encode as: rotation (4 bits) | immediate (8 bits)
			// The rotation field specifies how much to ROR the immediate value
			// We rotated RIGHT by 'rotate' bits to find the 8-bit value
			// So the CPU needs to rotate RIGHT by (32-rotate) bits to reconstruct the original
			// This is because we rotated right to compress, CPU rotates right to decompress
			decodeRotate := (32 - rotate) % 32
			return ((decodeRotate / 2) << 8) | rotated, true
		}
	}

	return 0, false
}

// parseShift parses a shift specification like "LSL #2" or "LSL R3"
func (e *Encoder) parseShift(shift string) (shiftType, shiftAmount uint32, shiftReg int32, err error) {
	shift = strings.TrimSpace(shift)
	if shift == "" {
		return 0, 0, -1, nil // No shift
	}

	parts := strings.Fields(shift)
	if len(parts) < 2 {
		return 0, 0, -1, fmt.Errorf("invalid shift: %s", shift)
	}

	// Parse shift type
	switch strings.ToUpper(parts[0]) {
	case "LSL":
		shiftType = 0
	case "LSR":
		shiftType = 1
	case "ASR":
		shiftType = 2
	case "ROR":
		shiftType = 3
	case "RRX":
		return 3, 0, -1, nil // RRX is encoded as ROR #0
	default:
		return 0, 0, -1, fmt.Errorf("unknown shift type: %s", parts[0])
	}

	// Parse shift amount (register or immediate)
	if strings.HasPrefix(parts[1], "#") {
		// Immediate shift
		amount, err := e.parseImmediate(parts[1])
		if err != nil {
			return 0, 0, -1, err
		}
		return shiftType, amount, -1, nil
	} else {
		// Register shift
		reg, err := e.parseRegister(parts[1])
		if err != nil {
			return 0, 0, -1, err
		}
		// Safe: register number is 0-15, well within int32 range
		return shiftType, 0, int32(reg), nil // #nosec G115 -- register is 0-15
	}
}

// evaluateExpression evaluates a constant expression like "label+12" or "symbol-4"
// Returns the evaluated value or an error if the expression is invalid
func (e *Encoder) evaluateExpression(expr string) (uint32, error) {
	expr = strings.TrimSpace(expr)

	// Look for + or - operators (scanning from left to right, skip first char for potential minus)
	for i := 1; i < len(expr); i++ {
		if expr[i] == '+' || expr[i] == '-' {
			left := strings.TrimSpace(expr[:i])
			right := strings.TrimSpace(expr[i+1:])
			op := expr[i]

			// Evaluate left side
			leftVal, err := e.evaluateTerm(left)
			if err != nil {
				return 0, err
			}

			// Evaluate right side
			rightVal, err := e.evaluateTerm(right)
			if err != nil {
				return 0, err
			}

			// Perform operation
			if op == '+' {
				return leftVal + rightVal, nil
			} else {
				return leftVal - rightVal, nil
			}
		}
	}

	// No operator found, evaluate as single term
	return e.evaluateTerm(expr)
} // evaluateTerm evaluates a single term (symbol or number)
func (e *Encoder) evaluateTerm(term string) (uint32, error) {
	term = strings.TrimSpace(term)

	// Try to resolve as symbol first
	if sym, exists := e.symbolTable.Lookup(term); exists && sym.Defined {
		return sym.Value, nil
	}

	// Otherwise parse as immediate number
	return e.parseImmediate(term)
}

// ValidatePoolCapacity checks if actual literal pool usage matches expected capacity
// This method should be called after encoding all instructions
func (e *Encoder) ValidatePoolCapacity() {
	if len(e.LiteralPoolLocs) == 0 {
		return
	}

	// Count actual literals in each pool region
	actualCounts := make(map[uint32]int) // pool location -> count of literals in that region

	for addr := range e.LiteralPool {
		// Find which pool this literal belongs to
		for i, poolLoc := range e.LiteralPoolLocs {
			if i+1 < len(e.LiteralPoolLocs) {
				// Check if literal is between this pool and the next
				if addr >= poolLoc && addr < e.LiteralPoolLocs[i+1] {
					actualCounts[poolLoc]++
					break
				}
			} else {
				// Last pool - all remaining literals belong to it
				if addr >= poolLoc {
					actualCounts[poolLoc]++
					break
				}
			}
		}
	}

	// Check each pool against expected capacity
	for i, poolLoc := range e.LiteralPoolLocs {
		expectedCount := parser.EstimatedLiteralsPerPool
		if i < len(e.LiteralPoolCounts) {
			expectedCount = e.LiteralPoolCounts[i]
		}

		actualCount := actualCounts[poolLoc]

		// Warn if actual count exceeds expected
		if actualCount > expectedCount {
			warning := fmt.Sprintf(
				"Literal pool at 0x%08X: actual count (%d) exceeds expected (%d)",
				poolLoc, actualCount, expectedCount,
			)
			e.PoolWarnings = append(e.PoolWarnings, warning)
		}

		// Also warn if we're using more than half the reserved space for pools with large margins
		if expectedCount >= parser.EstimatedLiteralsPerPool && actualCount > parser.EstimatedLiteralsPerPool/2 {
			warning := fmt.Sprintf(
				"Literal pool at 0x%08X: using %d of %d estimated literals (%.1f%%)",
				poolLoc, actualCount, parser.EstimatedLiteralsPerPool,
				float64(actualCount)/float64(parser.EstimatedLiteralsPerPool)*100,
			)
			e.PoolWarnings = append(e.PoolWarnings, warning)
		}
	}
}

// GetPoolWarnings returns all collected pool capacity warnings
func (e *Encoder) GetPoolWarnings() []string {
	return e.PoolWarnings
}

// HasPoolWarnings returns true if any warnings were collected
func (e *Encoder) HasPoolWarnings() bool {
	return len(e.PoolWarnings) > 0
}
