package vm_test

import (
	"fmt"
	"math/big"
	"math/bits"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// refImmediateShift is the ARM shift-by-immediate rule: amount 0 means LSL #0 (no
// shift, carry unchanged), LSR #32, ASR #32 or RRX.
func refImmediateShift(value uint32, shiftType vm.ShiftType, amount uint32, carryIn bool) (uint32, bool) {
	const wordBits = 32
	bit := func(n uint32) bool { return (value>>n)&1 == 1 }
	if amount == 0 {
		switch shiftType {
		case vm.ShiftLSL:
			return value, carryIn
		case vm.ShiftLSR:
			return 0, bit(wordBits - 1)
		case vm.ShiftASR:
			return uint32(int32(value) >> (wordBits - 1)), bit(wordBits - 1)
		default: // RRX
			c := uint32(0)
			if carryIn {
				c = 1
			}
			return c<<(wordBits-1) | value>>1, bit(0)
		}
	}
	switch shiftType {
	case vm.ShiftLSL:
		return value << amount, bit(wordBits - amount)
	case vm.ShiftLSR:
		return value >> amount, bit(amount - 1)
	case vm.ShiftASR:
		return uint32(int32(value) >> amount), bit(amount - 1)
	default:
		return bits.RotateLeft32(value, -int(amount)), bit(amount - 1)
	}
}

func TestImmediateShiftMatchesReferenceModel(t *testing.T) {
	const (
		movsR0R2        = 0xE1B00002 // MOVS R0, R2, <type> #<amount> with type and amount zero
		shiftTypeShift  = 5
		shiftAmountPos  = 7
		maxShiftAmount  = 31
		shiftTypeCount  = 4
		andsR0R1R2Shift = 0xE0110002 // ANDS R0, R1, R2, <type> #<amount>
	)
	values := []uint32{0x80000001, 0x7FFFFFFE, 0xFFFFFFFF, 1, 0}
	for typ := range vm.ShiftType(shiftTypeCount) {
		for amount := range uint32(maxShiftAmount + 1) {
			for _, value := range values {
				for _, carryIn := range []bool{false, true} {
					name := fmt.Sprintf("type%d/#%d/%08X/C=%v", typ, amount, value, carryIn)
					t.Run(name, func(t *testing.T) {
						field := uint32(typ)<<shiftTypeShift | amount<<shiftAmountPos
						want, wantCarry := refImmediateShift(value, typ, amount, carryIn)

						v := vm.NewVM()
						v.CPU.R[2] = value
						v.CPU.CPSR.C = carryIn
						stepOne(t, v, movsR0R2|field)
						if v.CPU.R[0] != want || v.CPU.CPSR.C != wantCarry {
							t.Errorf("MOVS: got 0x%08X C=%v, want 0x%08X C=%v", v.CPU.R[0], v.CPU.CPSR.C, want, wantCarry)
						}

						// A logical operation takes its carry from the shifter too.
						v = vm.NewVM()
						v.CPU.R[1], v.CPU.R[2] = 0xFFFFFFFF, value
						v.CPU.CPSR.C = carryIn
						stepOne(t, v, andsR0R1R2Shift|field)
						if v.CPU.R[0] != want || v.CPU.CPSR.C != wantCarry {
							t.Errorf("ANDS: got 0x%08X C=%v, want 0x%08X C=%v", v.CPU.R[0], v.CPU.CPSR.C, want, wantCarry)
						}
					})
				}
			}
		}
	}
}

// Rotated immediates: the value is imm8 ROR (2 * rot), and a non-zero rotation sets C
// to bit 31 of the result.
func TestRotatedImmediateMatchesReferenceModel(t *testing.T) {
	const (
		movsR0Imm     = 0xE3B00000 // MOVS R0, #imm8 ROR (2*rot)
		rotationShift = 8
		rotations     = 16
	)
	for rot := range uint32(rotations) {
		for _, imm8 := range []uint32{0, 1, 0x80, 0xFF, 0x7F} {
			for _, carryIn := range []bool{false, true} {
				t.Run(fmt.Sprintf("rot%d/%02X/C=%v", rot, imm8, carryIn), func(t *testing.T) {
					want := bits.RotateLeft32(imm8, -int(2*rot))
					wantCarry := carryIn
					if rot != 0 {
						wantCarry = want>>31 == 1
					}
					v := vm.NewVM()
					v.CPU.CPSR.C = carryIn
					stepOne(t, v, movsR0Imm|rot<<rotationShift|imm8)
					if v.CPU.R[0] != want || v.CPU.CPSR.C != wantCarry {
						t.Errorf("got 0x%08X C=%v, want 0x%08X C=%v", v.CPU.R[0], v.CPU.CPSR.C, want, wantCarry)
					}
				})
			}
		}
	}
}

var multiplyBoundaryValues = []uint32{0, 1, 2, 0x7FFFFFFF, 0x80000000, 0x80000001, 0xFFFFFFFE, 0xFFFFFFFF, 0x12345678}

// refMultiply returns the 64-bit result of a long multiply computed with math/big.
func refMultiply(a, b uint32, signed bool, acc uint64, accumulate bool) uint64 {
	const wordBits = 64
	toBig := func(x uint32) *big.Int {
		if signed {
			return big.NewInt(int64(int32(x)))
		}
		return new(big.Int).SetUint64(uint64(x))
	}
	product := new(big.Int).Mul(toBig(a), toBig(b))
	if accumulate {
		product.Add(product, new(big.Int).SetUint64(acc))
	}
	modulus := new(big.Int).Lsh(big.NewInt(1), wordBits)
	return new(big.Int).Mod(product, modulus).Uint64()
}

func TestMultiplyMatchesReferenceModel(t *testing.T) {
	// Registers: Rm=R2, Rs=R3, Rn (MLA accumulator)=R4; Rd=R0, or RdLo=R0 and RdHi=R1.
	const (
		mulsR0   = 0xE0100392 // MULS R0, R2, R3
		mlasR0   = 0xE0304392 // MLAS R0, R2, R3, R4
		umullsR0 = 0xE0910392 // UMULLS R0, R1, R2, R3
		umlalsR0 = 0xE0B10392 // UMLALS R0, R1, R2, R3
		smullsR0 = 0xE0D10392 // SMULLS R0, R1, R2, R3
		smlalsR0 = 0xE0F10392 // SMLALS R0, R1, R2, R3
		accLow   = 0xFFFFFFFF
		accHigh  = 0x7FFFFFFF
		accWord  = 0x9
		hiShift  = 32
	)
	type longOp struct {
		name               string
		opcode             uint32
		signed, accumulate bool
	}
	longOps := []longOp{
		{"UMULLS", umullsR0, false, false}, {"UMLALS", umlalsR0, false, true},
		{"SMULLS", smullsR0, true, false}, {"SMLALS", smlalsR0, true, true},
	}
	for _, a := range multiplyBoundaryValues {
		for _, b := range multiplyBoundaryValues {
			t.Run(fmt.Sprintf("%08X*%08X", a, b), func(t *testing.T) {
				// C and V are preset and must not change.
				preset := vm.CPSR{C: true, V: true}

				for _, op := range []struct {
					name   string
					opcode uint32
					want   uint32
				}{
					{"MULS", mulsR0, a * b},
					{"MLAS", mlasR0, a*b + accWord},
				} {
					v := vm.NewVM()
					v.CPU.R[2], v.CPU.R[3], v.CPU.R[4] = a, b, accWord
					v.CPU.CPSR = preset
					stepOne(t, v, op.opcode)
					wantFlags := nzcv{N: op.want>>31 == 1, Z: op.want == 0, C: true, V: true}
					if v.CPU.R[0] != op.want || flagsOf(v) != wantFlags {
						t.Errorf("%s = 0x%08X %+v, want 0x%08X %+v", op.name, v.CPU.R[0], flagsOf(v), op.want, wantFlags)
					}
				}

				for _, op := range longOps {
					acc := uint64(accHigh)<<hiShift | accLow
					want := refMultiply(a, b, op.signed, acc, op.accumulate)
					v := vm.NewVM()
					v.CPU.R[2], v.CPU.R[3] = a, b
					v.CPU.R[0], v.CPU.R[1] = accLow, accHigh
					v.CPU.CPSR = preset
					stepOne(t, v, op.opcode)
					got := uint64(v.CPU.R[1])<<hiShift | uint64(v.CPU.R[0])
					wantFlags := nzcv{N: want>>63 == 1, Z: want == 0, C: true, V: true}
					if got != want || flagsOf(v) != wantFlags {
						t.Errorf("%s = 0x%016X %+v, want 0x%016X %+v", op.name, got, flagsOf(v), want, wantFlags)
					}
				}
			})
		}
	}
}

// refBlockTransfer returns the addresses an LDM/STM touches in register order and the
// written-back base.
func refBlockTransfer(base uint32, count int, increment, before bool) ([]uint32, uint32) {
	const word = 4
	size := uint32(count) * word
	var start, newBase uint32
	switch {
	case increment && !before: // IA
		start, newBase = base, base+size
	case increment && before: // IB
		start, newBase = base+word, base+size
	case !increment && !before: // DA
		start, newBase = base-size+word, base-size
	default: // DB
		start, newBase = base-size, base-size
	}
	addrs := make([]uint32, count)
	for i := range addrs {
		addrs[i] = start + uint32(i)*word
	}
	return addrs, newBase
}

func TestBlockTransferMatchesReferenceModel(t *testing.T) {
	const (
		ldmBase        = 0xE8100000 // LDM R0, {...} with P, U, W zero
		stmBase        = 0xE8000000
		pBit           = 1 << 24
		uBit           = 1 << 23
		wBit           = 1 << 21
		base           = vm.DataSegmentStart + 0x100
		fillPattern    = 0xA0000000
		storedPattern  = 0x50000000
		registerR0Base = 0 // Rn = R0
	)
	lists := map[string][]int{
		"one":      {5},
		"two":      {1, 2},
		"sparse":   {1, 4, 9, 12},
		"low-high": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 14},
	}
	modes := []struct {
		name              string
		increment, before bool
	}{{"IA", true, false}, {"IB", true, true}, {"DA", false, false}, {"DB", false, true}}

	for listName, regs := range lists {
		var mask uint32
		for _, r := range regs {
			mask |= 1 << r
		}
		for _, mode := range modes {
			for _, writeBack := range []bool{false, true} {
				var bitsPUW uint32
				if mode.before {
					bitsPUW |= pBit
				}
				if mode.increment {
					bitsPUW |= uBit
				}
				if writeBack {
					bitsPUW |= wBit
				}
				addrs, newBase := refBlockTransfer(base, len(regs), mode.increment, mode.before)
				wantBase := uint32(base)
				if writeBack {
					wantBase = newBase
				}

				t.Run(fmt.Sprintf("LDM%s/%s/W=%v", mode.name, listName, writeBack), func(t *testing.T) {
					v := vm.NewVM()
					for i := range uint32(64) {
						mustWriteWord(t, v, base-0x80+i*4, fillPattern|(base-0x80+i*4))
					}
					v.CPU.R[registerR0Base] = base
					stepOne(t, v, ldmBase|bitsPUW|mask)
					for i, r := range regs {
						if want := fillPattern | addrs[i]; v.CPU.R[r] != want {
							t.Errorf("R%d = 0x%08X, want word at 0x%08X", r, v.CPU.R[r], addrs[i])
						}
					}
					if v.CPU.R[0] != wantBase {
						t.Errorf("base = 0x%08X, want 0x%08X", v.CPU.R[0], wantBase)
					}
				})

				t.Run(fmt.Sprintf("STM%s/%s/W=%v", mode.name, listName, writeBack), func(t *testing.T) {
					v := vm.NewVM()
					for _, r := range regs {
						v.CPU.R[r] = storedPattern | uint32(r)
					}
					v.CPU.R[registerR0Base] = base
					stepOne(t, v, stmBase|bitsPUW|mask)
					for i, r := range regs {
						if got, _ := v.Memory.ReadWord(addrs[i]); got != storedPattern|uint32(r) {
							t.Errorf("[0x%08X] = 0x%08X, want R%d", addrs[i], got, r)
						}
					}
					// Nothing outside the block changes.
					lo, hi := addrs[0], addrs[len(addrs)-1]
					for addr := uint32(base - 0x80); addr < base+0x80; addr += 4 {
						if addr >= lo && addr <= hi {
							continue
						}
						if got, _ := v.Memory.ReadWord(addr); got != 0 {
							t.Errorf("[0x%08X] = 0x%08X outside the block", addr, got)
						}
					}
					if v.CPU.R[0] != wantBase {
						t.Errorf("base = 0x%08X, want 0x%08X", v.CPU.R[0], wantBase)
					}
				})
			}
		}
	}
}
