package vm_test

import (
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// Halfword-group encodings with Rd=R0, Rn=R1 and a zero immediate offset.
const (
	ldrsbR0R1  = 0xE1D100D0 // LDRSB R0, [R1]
	ldrshR0R1  = 0xE1D100F0 // LDRSH R0, [R1]
	swpR0R2R1  = 0xE1010092 // SWP R0, R2, [R1]
	swpbR0R2R1 = 0xE1410092 // SWPB R0, R2, [R1]
	strsbForm  = 0xE1C100D0 // S=1 with L=0: no ARMv2/v3 instruction
	strshForm  = 0xE1C100F0
)

func TestSignedLoads(t *testing.T) {
	tests := []struct {
		name   string
		opcode uint32
		offset uint32 // byte offset of the access from the data start
		want   uint32
	}{
		{"LDRSB negative", ldrsbR0R1, 0, 0xFFFFFF80},
		{"LDRSB positive max", ldrsbR0R1, 1, 0x0000007F},
		{"LDRSB minus one", ldrsbR0R1, 2, 0xFFFFFFFF},
		{"LDRSB zero", ldrsbR0R1, 3, 0},
		{"LDRSH positive", ldrshR0R1, 0, 0x00007F80},
		{"LDRSH minus one", ldrshR0R1, 2, 0x000000FF},
		{"LDRSH positive max", ldrshR0R1, 4, 0x00007FFF},
		{"LDRSH most negative", ldrshR0R1, 6, 0xFFFF8000},
	}
	data := []byte{0x80, 0x7F, 0xFF, 0x00, 0xFF, 0x7F, 0x00, 0x80}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			if err := v.Memory.LoadBytes(vm.DataSegmentStart, data); err != nil {
				t.Fatal(err)
			}
			v.CPU.R[1] = vm.DataSegmentStart + tt.offset
			stepOne(t, v, tt.opcode)
			if v.CPU.R[0] != tt.want {
				t.Errorf("R0 = 0x%08X, want 0x%08X", v.CPU.R[0], tt.want)
			}
		})
	}
}

func TestLDRSH_Unaligned(t *testing.T) {
	v := vm.NewVM()
	v.CPU.R[1] = vm.DataSegmentStart + 1
	v.CPU.PC = vm.CodeSegmentStart
	if err := v.Memory.WriteWord(vm.CodeSegmentStart, ldrshR0R1); err != nil {
		t.Fatal(err)
	}
	if err := v.Step(); err == nil || !strings.Contains(err.Error(), "unaligned") {
		t.Errorf("Step() error = %v, want unaligned halfword error", err)
	}
}

func TestUnsupportedHalfwordGroupEncodingsFailToDecode(t *testing.T) {
	tests := []struct {
		name   string
		opcode uint32
	}{
		{"SWP", swpR0R2R1},
		{"SWPB", swpbR0R2R1},
		{"store signed byte", strsbForm},
		{"store signed halfword", strshForm},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vm.NewVM()
			v.CPU.R[1] = vm.DataSegmentStart
			v.CPU.R[0] = 0x12345678
			v.CPU.PC = vm.CodeSegmentStart
			if err := v.Memory.WriteWord(vm.CodeSegmentStart, tt.opcode); err != nil {
				t.Fatal(err)
			}
			err := v.Step()
			if err == nil || !strings.Contains(err.Error(), "decode failed") {
				t.Fatalf("Step() error = %v, want decode failure", err)
			}
			if v.State != vm.StateError {
				t.Errorf("State = %v, want StateError", v.State)
			}
			if got, _ := v.Memory.ReadWord(vm.DataSegmentStart); got != 0 {
				t.Errorf("memory changed to 0x%08X", got)
			}
		})
	}
}
