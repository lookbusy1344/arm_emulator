package vm_test

import (
	"fmt"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// Operand values at the unsigned and signed boundaries, where carry and overflow change.
var aluBoundaryValues = []uint32{
	0, 1, 2, 0x7FFFFFFE, 0x7FFFFFFF, 0x80000000, 0x80000001, 0xFFFFFFFE, 0xFFFFFFFF,
}

type nzcv struct{ N, Z, C, V bool }

// refAddWithCarry is the ARM AddWithCarry pseudocode, computed in 64 bits.
func refAddWithCarry(a, b uint32, carryIn bool) (uint32, nzcv) {
	c := uint64(0)
	if carryIn {
		c = 1
	}
	unsigned := uint64(a) + uint64(b) + c
	signed := int64(int32(a)) + int64(int32(b)) + int64(c)
	result := uint32(unsigned)
	return result, nzcv{
		N: result>>31 == 1,
		Z: result == 0,
		C: unsigned > uint64(^uint32(0)),
		V: signed != int64(int32(result)),
	}
}

// refArithmetic returns the expected result and flags of an arithmetic data-processing
// opcode with Rn=a and Operand2=b.
func refArithmetic(opcode uint32, a, b uint32, carryIn bool) (uint32, nzcv) {
	switch opcode {
	case vm.OpSUB, vm.OpCMP:
		return refAddWithCarry(a, ^b, true)
	case vm.OpRSB:
		return refAddWithCarry(b, ^a, true)
	case vm.OpADD, vm.OpCMN:
		return refAddWithCarry(a, b, false)
	case vm.OpADC:
		return refAddWithCarry(a, b, carryIn)
	case vm.OpSBC:
		return refAddWithCarry(a, ^b, carryIn)
	case vm.OpRSC:
		return refAddWithCarry(b, ^a, carryIn)
	}
	panic(fmt.Sprintf("not an arithmetic opcode: %d", opcode))
}

// encodeDPRegister encodes <op>S R0, R1, R2.
func encodeDPRegister(opcode uint32) uint32 {
	const (
		condAL  = 0xE0000000
		opShift = 21
		sBit    = 1 << 20
		rnShift = 16
		rnR1    = 1
		rmR2    = 2
		rdShift = 12
		rdR0    = 0
	)
	return condAL | opcode<<opShift | sBit | rnR1<<rnShift | rdR0<<rdShift | rmR2
}

func stepOne(t *testing.T, v *vm.VM, opcode uint32) {
	t.Helper()
	v.CPU.PC = vm.CodeSegmentStart
	if err := v.Memory.WriteWord(vm.CodeSegmentStart, opcode); err != nil {
		t.Fatalf("write opcode: %v", err)
	}
	if err := v.Step(); err != nil {
		t.Fatalf("step 0x%08X: %v", opcode, err)
	}
}

func flagsOf(v *vm.VM) nzcv {
	return nzcv{v.CPU.CPSR.N, v.CPU.CPSR.Z, v.CPU.CPSR.C, v.CPU.CPSR.V}
}

func TestArithmeticMatchesReferenceModel(t *testing.T) {
	ops := map[string]uint32{
		"ADD": vm.OpADD, "ADC": vm.OpADC, "SUB": vm.OpSUB, "SBC": vm.OpSBC,
		"RSB": vm.OpRSB, "RSC": vm.OpRSC, "CMP": vm.OpCMP, "CMN": vm.OpCMN,
	}
	for name, op := range ops {
		writesResult := op != vm.OpCMP && op != vm.OpCMN
		for _, a := range aluBoundaryValues {
			for _, b := range aluBoundaryValues {
				for _, carryIn := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%08X/%08X/C=%v", name, a, b, carryIn), func(t *testing.T) {
						v := vm.NewVM()
						v.CPU.R[1], v.CPU.R[2] = a, b
						v.CPU.CPSR.C = carryIn
						stepOne(t, v, encodeDPRegister(op))

						wantResult, wantFlags := refArithmetic(op, a, b, carryIn)
						if got := flagsOf(v); got != wantFlags {
							t.Errorf("flags = %+v, want %+v", got, wantFlags)
						}
						if writesResult && v.CPU.R[0] != wantResult {
							t.Errorf("R0 = 0x%08X, want 0x%08X", v.CPU.R[0], wantResult)
						}
					})
				}
			}
		}
	}
}
