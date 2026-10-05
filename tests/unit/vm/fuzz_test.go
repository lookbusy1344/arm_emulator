package vm_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// FuzzStep executes one arbitrary instruction word with arbitrary register contents.
// No word may panic, and every failure must leave the VM in the error state unless it
// is an exit or breakpoint SWI.
func FuzzStep(f *testing.F) {
	for _, seed := range []uint32{
		0xE0910002, 0xE1B00312, 0xE5910000, 0xE1D100F0, 0xE8BD8000, 0xE92D4000,
		0xE12FFF1E, 0xEB000000, 0xE0810392, 0xE0C10392, 0xE10F0000, 0xE129F000,
		0xEF000000, 0xEF0000F1, 0xE1010092, 0xEE000000, 0xF0000000,
	} {
		f.Add(seed, uint32(vm.DataSegmentStart), uint32(4), uint32(0xFFFFFFFF), uint8(0))
	}
	f.Fuzz(func(t *testing.T, opcode, r1, r2, r3 uint32, flags uint8) {
		const nzcvBits = 4
		v := vm.NewVM()
		v.CycleLimit = 1
		v.OutputWriter = discard{}
		v.SetStdinReader(strings.NewReader("12\nabc\n"))
		for i := range vm.SP {
			v.CPU.R[i] = []uint32{r1, r2, r3}[i%3]
		}
		v.CPU.R[vm.SP] = vm.StackSegmentStart + vm.StackSegmentSize/2
		v.CPU.CPSR.FromUint32(uint32(flags) << (32 - nzcvBits))
		v.CPU.PC = vm.CodeSegmentStart
		if err := v.Memory.WriteWord(vm.CodeSegmentStart, opcode); err != nil {
			t.Fatal(err)
		}

		err := v.Step()
		switch {
		case err == nil, errors.Is(err, vm.ErrProgramExited), errors.Is(err, vm.ErrBreakpointHit):
		case errors.Is(err, vm.ErrInputInterrupted):
			t.Fatalf("opcode 0x%08X: input interrupted without an input source", opcode)
		default:
			if v.State != vm.StateError {
				t.Fatalf("opcode 0x%08X failed (%v) but State = %v", opcode, err, v.State)
			}
		}
	})
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
