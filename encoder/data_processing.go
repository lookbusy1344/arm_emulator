package encoder

import (
	"fmt"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// Data processing instruction opcodes
const (
	opAND = 0x0
	opEOR = 0x1
	opSUB = 0x2
	opRSB = 0x3
	opADD = 0x4
	opADC = 0x5
	opSBC = 0x6
	opRSC = 0x7
	opTST = 0x8
	opTEQ = 0x9
	opCMP = 0xA
	opCMN = 0xB
	opORR = 0xC
	opMOV = 0xD
	opBIC = 0xE
	opMVN = 0xF
)

// encodeDataProcessingMove encodes MOV and MVN instructions
func (e *Encoder) encodeDataProcessingMove(inst *parser.Instruction, cond uint32) (uint32, error) {
	if len(inst.Operands) < 2 {
		return 0, fmt.Errorf("MOV/MVN requires 2 operands, got %d", len(inst.Operands))
	}

	// Get destination register
	rd, err := e.parseRegister(inst.Operands[0])
	if err != nil {
		return 0, err
	}

	// Determine opcode
	var opcode uint32
	if strings.ToUpper(inst.Mnemonic) == "MOV" {
		opcode = opMOV
	} else {
		opcode = opMVN
	}

	// S bit
	var sBit uint32
	if inst.SetFlags {
		sBit = 1
	}

	// Parse second operand
	operand2 := inst.Operands[1]
	return e.encodeOperand2(cond, opcode, 0, rd, sBit, operand2)
}

// encodeDataProcessingArithmetic encodes arithmetic instructions (ADD, SUB, etc.)
func (e *Encoder) encodeDataProcessingArithmetic(inst *parser.Instruction, cond uint32) (uint32, error) {
	if len(inst.Operands) < 3 {
		return 0, fmt.Errorf("%s requires 3 operands, got %d", inst.Mnemonic, len(inst.Operands))
	}

	// Get registers
	rd, err := e.parseRegister(inst.Operands[0])
	if err != nil {
		return 0, err
	}

	rn, err := e.parseRegister(inst.Operands[1])
	if err != nil {
		return 0, err
	}

	// Determine opcode
	var opcode uint32
	switch strings.ToUpper(inst.Mnemonic) {
	case "ADD":
		opcode = opADD
	case "ADC":
		opcode = opADC
	case "SUB":
		opcode = opSUB
	case "SBC":
		opcode = opSBC
	case "RSB":
		opcode = opRSB
	case "RSC":
		opcode = opRSC
	default:
		return 0, fmt.Errorf("unknown arithmetic instruction: %s", inst.Mnemonic)
	}

	// S bit
	var sBit uint32
	if inst.SetFlags {
		sBit = 1
	}

	// Parse third operand
	operand2 := inst.Operands[2]
	return e.encodeOperand2(cond, opcode, rn, rd, sBit, operand2)
}

// encodeDataProcessingLogical encodes logical instructions (AND, ORR, etc.)
func (e *Encoder) encodeDataProcessingLogical(inst *parser.Instruction, cond uint32) (uint32, error) {
	if len(inst.Operands) < 3 {
		return 0, fmt.Errorf("%s requires 3 operands, got %d", inst.Mnemonic, len(inst.Operands))
	}

	// Get registers
	rd, err := e.parseRegister(inst.Operands[0])
	if err != nil {
		return 0, err
	}

	rn, err := e.parseRegister(inst.Operands[1])
	if err != nil {
		return 0, err
	}

	// Determine opcode
	var opcode uint32
	switch strings.ToUpper(inst.Mnemonic) {
	case "AND":
		opcode = opAND
	case "ORR":
		opcode = opORR
	case "EOR":
		opcode = opEOR
	case "BIC":
		opcode = opBIC
	default:
		return 0, fmt.Errorf("unknown logical instruction: %s", inst.Mnemonic)
	}

	// S bit
	var sBit uint32
	if inst.SetFlags {
		sBit = 1
	}

	// Parse third operand
	operand2 := inst.Operands[2]
	return e.encodeOperand2(cond, opcode, rn, rd, sBit, operand2)
}

// encodeDataProcessingCompare encodes comparison instructions (CMP, TST, etc.)
func (e *Encoder) encodeDataProcessingCompare(inst *parser.Instruction, cond uint32) (uint32, error) {
	if len(inst.Operands) < 2 {
		return 0, fmt.Errorf("%s requires 2 operands, got %d", inst.Mnemonic, len(inst.Operands))
	}

	// Get first register
	rn, err := e.parseRegister(inst.Operands[0])
	if err != nil {
		return 0, err
	}

	// Determine opcode
	var opcode uint32
	switch strings.ToUpper(inst.Mnemonic) {
	case "CMP":
		opcode = opCMP
	case "CMN":
		opcode = opCMN
	case "TST":
		opcode = opTST
	case "TEQ":
		opcode = opTEQ
	default:
		return 0, fmt.Errorf("unknown comparison instruction: %s", inst.Mnemonic)
	}

	// Comparison instructions always set flags (S bit always 1)
	sBit := uint32(1)

	// Parse second operand
	operand2 := inst.Operands[1]
	// Rd is ignored for comparison instructions, set to 0
	return e.encodeOperand2(cond, opcode, rn, 0, sBit, operand2)
}

// encodeOperand2 encodes operand2 field for data processing instructions
func (e *Encoder) encodeOperand2(cond, opcode, rn, rd, sBit uint32, operand string) (uint32, error) {
	operand = strings.TrimSpace(operand)

	// Check if it's an immediate value
	if strings.HasPrefix(operand, "#") || isNumeric(operand) {
		// Immediate operand
		value, err := e.parseImmediate(operand)
		if err != nil {
			return 0, err
		}

		// Try to encode as rotated immediate
		encoded, ok := e.encodeImmediate(value)
		if !ok {
			return 0, fmt.Errorf("immediate value 0x%08X cannot be encoded as ARM immediate", value)
		}

		// Format: cccc 001o oooo Srrr rddd iiii iiii iiii
		// I=1 for immediate
		instruction := (cond << 28) | (1 << 25) | (opcode << 21) | (sBit << 20) |
			(rn << 16) | (rd << 12) | encoded

		return instruction, nil
	}

	// Parse as register with optional shift
	parts := strings.Split(operand, ",")
	rm, err := e.parseRegister(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, err
	}

	var shiftField uint32

	if len(parts) > 1 {
		// Has shift
		shiftStr := strings.TrimSpace(strings.Join(parts[1:], ","))
		shiftType, shiftAmount, shiftReg, err := e.parseShift(shiftStr)
		if err != nil {
			return 0, err
		}

		if shiftReg >= 0 {
			// Register shift: bit 4 = 1
			shiftField = (uint32(shiftReg) << 8) | (shiftType << 5) | (1 << 4) | rm
		} else {
			// Immediate shift: bit 4 = 0
			shiftField = (shiftAmount << 7) | (shiftType << 5) | rm
		}
	} else {
		// No shift
		shiftField = rm
	}

	// Format: cccc 000o oooo Srrr rddd ssss ssss mmmm
	// I=0 for register
	instruction := (cond << 28) | (0 << 25) | (opcode << 21) | (sBit << 20) |
		(rn << 16) | (rd << 12) | shiftField

	return instruction, nil
}

// isNumeric checks if a string looks like a number
func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "-") {
		s = s[1:]
	}
	return strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") ||
		strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") ||
		(s[0] >= '0' && s[0] <= '9')
}
