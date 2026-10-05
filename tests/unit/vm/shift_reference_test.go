package vm_test

import (
	"fmt"
	"math/bits"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// refRegisterShift is the ARM shift-by-register rule: the amount is Rs[7:0], and an
// amount of 0 leaves both the value and the carry unchanged.
func refRegisterShift(value uint32, shiftType vm.ShiftType, amount uint32, carryIn bool) (uint32, bool) {
	const wordBits = 32
	bit := func(n uint32) bool { return (value>>n)&1 == 1 }
	if amount == 0 {
		return value, carryIn
	}
	switch shiftType {
	case vm.ShiftLSL:
		switch {
		case amount < wordBits:
			return value << amount, bit(wordBits - amount)
		case amount == wordBits:
			return 0, bit(0)
		default:
			return 0, false
		}
	case vm.ShiftLSR:
		switch {
		case amount < wordBits:
			return value >> amount, bit(amount - 1)
		case amount == wordBits:
			return 0, bit(wordBits - 1)
		default:
			return 0, false
		}
	case vm.ShiftASR:
		if amount < wordBits {
			return uint32(int32(value) >> amount), bit(amount - 1)
		}
		return uint32(int32(value) >> (wordBits - 1)), bit(wordBits - 1)
	case vm.ShiftROR:
		effective := amount % wordBits
		if effective == 0 {
			return value, bit(wordBits - 1)
		}
		return bits.RotateLeft32(value, -int(effective)), bit(effective - 1)
	}
	panic(fmt.Sprintf("not a register shift type: %d", shiftType))
}

func TestRegisterShiftMatchesReferenceModel(t *testing.T) {
	const (
		movsR0R2ShiftR3 = 0xE1B00312 // MOVS R0, R2, <type> R3 with type bits zero
		shiftTypeShift  = 5
		shiftAmountMask = 0xFF // Rs[7:0]
	)
	values := []uint32{0x80000001, 0x7FFFFFFE, 0xFFFFFFFF, 1}
	// Rs above 255 checks that only the bottom byte counts.
	amounts := []uint32{0, 1, 31, 32, 33, 64, 96, 255, 256, 0x120}
	types := []vm.ShiftType{vm.ShiftLSL, vm.ShiftLSR, vm.ShiftASR, vm.ShiftROR}
	for _, shiftType := range types {
		for _, amount := range amounts {
			for _, value := range values {
				for _, carryIn := range []bool{false, true} {
					t.Run(fmt.Sprintf("type%d/%d/%08X/C=%v", shiftType, amount, value, carryIn), func(t *testing.T) {
						v := vm.NewVM()
						v.CPU.R[2], v.CPU.R[3] = value, amount
						v.CPU.CPSR.C = carryIn
						stepOne(t, v, movsR0R2ShiftR3|uint32(shiftType)<<shiftTypeShift)

						want, wantCarry := refRegisterShift(value, shiftType, amount&shiftAmountMask, carryIn)
						if v.CPU.R[0] != want || v.CPU.CPSR.C != wantCarry {
							t.Errorf("got 0x%08X C=%v, want 0x%08X C=%v", v.CPU.R[0], v.CPU.CPSR.C, want, wantCarry)
						}
					})
				}
			}
		}
	}
}
