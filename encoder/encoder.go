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
	symbolTable      *parser.SymbolTable
	currentAddr      uint32
	LiteralPool      map[uint32]uint32 // address -> value for literal pool (exported)
	literalCounter   uint32            // Counter for generating unique literal addresses
	LiteralPoolStart uint32            // Start address for literal pool (set externally)
}

// NewEncoder creates a new encoder instance
func NewEncoder(symbolTable *parser.SymbolTable) *Encoder {
	return &Encoder{
		symbolTable: symbolTable,
		LiteralPool: make(map[uint32]uint32),
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
	case "B", "BL", "BX":
		return e.encodeBranch(inst, cond)

	// Multiply instructions
	case "MUL", "MLA":
		return e.encodeMultiply(inst, cond)

	// Load/Store multiple
	case "LDM", "STM", "LDMIA", "LDMIB", "LDMDA", "LDMDB":
		return e.encodeLoadStoreMultiple(inst, cond, false)
	case "STMIA", "STMIB", "STMDA", "STMDB":
		return e.encodeLoadStoreMultiple(inst, cond, true)
	case "PUSH":
		return e.encodePush(inst, cond)
	case "POP":
		return e.encodePop(inst, cond)

	// Software interrupt
	case "SWI":
		return e.encodeSWI(inst, cond)

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
	case "NV":
		return 0xF // Never (rarely used)
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
	if strings.HasPrefix(imm, "#") {
		imm = imm[1:]
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
		if result > uint32(math.MaxInt32)+1 {
			return 0, fmt.Errorf("immediate value out of valid signed 32-bit range: %s", imm)
		}
		result = uint32(-int32(result))
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
			// Rotation is stored as rotate/2 in bits 11-8
			return ((rotate / 2) << 8) | rotated, true
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
		return shiftType, 0, int32(reg), nil
	}
}
