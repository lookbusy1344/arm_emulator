package encoder

import (
	"fmt"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// MSR field mask bits (bits 19-16). The VM applies only the flags field.
const (
	psrFieldControl   = 1 << 0 // c: bits 7-0
	psrFieldExtension = 1 << 1 // x: bits 15-8
	psrFieldStatus    = 1 << 2 // s: bits 23-16
	psrFieldFlags     = 1 << 3 // f: bits 31-24
	psrFieldsAll      = psrFieldControl | psrFieldFlags
	psrFieldMaskShift = 16
)

// encodePSRTransfer encodes MRS and MSR. The assembler accepts only CPSR.
//
//	MRS Rd, CPSR               cccc 0001 0000 1111 dddd 0000 0000 0000
//	MSR CPSR_<fields>, Rm      cccc 0001 0010 ffff 1111 0000 0000 mmmm
//	MSR CPSR_<fields>, #imm    cccc 0011 0010 ffff 1111 rrrr iiii iiii
func (e *Encoder) encodePSRTransfer(inst *parser.Instruction, cond uint32) (uint32, error) {
	mnemonic := strings.ToUpper(inst.Mnemonic)
	if len(inst.Operands) != 2 {
		return 0, fmt.Errorf("%s requires 2 operands, got %d", mnemonic, len(inst.Operands))
	}
	if mnemonic == "MRS" {
		return e.encodeMRS(inst.Operands, cond)
	}
	return e.encodeMSR(inst.Operands, cond)
}

func (e *Encoder) encodeMRS(operands []string, cond uint32) (uint32, error) {
	rd, err := e.parseRegister(operands[0])
	if err != nil {
		return 0, err
	}
	if rd == RegisterPC {
		return 0, fmt.Errorf("MRS: PC cannot be the destination")
	}
	if !strings.EqualFold(strings.TrimSpace(operands[1]), "CPSR") {
		return 0, fmt.Errorf("MRS: unsupported status register %q (only CPSR)", operands[1])
	}
	return cond<<ConditionShift | vm.MRSPattern | rd<<RdShift, nil
}

func (e *Encoder) encodeMSR(operands []string, cond uint32) (uint32, error) {
	fields, err := parsePSRFields(operands[0])
	if err != nil {
		return 0, err
	}
	base := cond<<ConditionShift | fields<<psrFieldMaskShift | RegisterPC<<RdShift

	source := strings.TrimSpace(operands[1])
	if strings.HasPrefix(source, "#") {
		value, err := e.parseImmediate(source)
		if err != nil {
			return 0, err
		}
		encoded, ok := e.encodeImmediate(value)
		if !ok {
			return 0, fmt.Errorf("MSR: immediate 0x%X cannot be encoded as a rotated 8-bit value", value)
		}
		return base | vm.MSRImmPattern | encoded, nil
	}

	rm, err := e.parseRegister(source)
	if err != nil {
		return 0, err
	}
	if rm == RegisterPC {
		return 0, fmt.Errorf("MSR: PC cannot be the source")
	}
	return base | vm.MSRRegPattern | rm, nil
}

// parsePSRFields parses CPSR, CPSR_all, CPSR_flg or CPSR_<fsxc> into a field mask.
func parsePSRFields(operand string) (uint32, error) {
	name, suffix, _ := strings.Cut(strings.ToUpper(strings.TrimSpace(operand)), "_")
	if name != "CPSR" {
		return 0, fmt.Errorf("MSR: unsupported status register %q (only CPSR)", operand)
	}
	switch suffix {
	case "", "ALL":
		return psrFieldsAll, nil
	case "FLG":
		return psrFieldFlags, nil
	}
	var mask uint32
	for _, c := range suffix {
		switch c {
		case 'C':
			mask |= psrFieldControl
		case 'X':
			mask |= psrFieldExtension
		case 'S':
			mask |= psrFieldStatus
		case 'F':
			mask |= psrFieldFlags
		default:
			return 0, fmt.Errorf("MSR: invalid field %q in %q", c, operand)
		}
	}
	return mask, nil
}
