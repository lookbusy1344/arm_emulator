package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func runShift(t *testing.T, opcode, r1, r2 uint32, carryIn bool) *vm.VM {
	t.Helper()
	v := vm.NewVM()
	loadWords(t, v, opcode)
	v.CPU.R[1] = r1
	v.CPU.R[2] = r2
	v.CPU.CPSR.C = carryIn
	if err := v.Step(); err != nil {
		t.Fatalf("Step failed: %v", err)
	}
	return v
}

func TestImmediateShiftEncodedAsThirtyTwo(t *testing.T) {
	tests := []struct {
		name      string
		opcode    uint32
		r1        uint32
		wantR0    uint32
		wantCarry bool
	}{
		{"LSR #32 negative", 0xE1B00021, 0x80000000, 0, true},
		{"LSR #32 positive", 0xE1B00021, 0x7FFFFFFF, 0, false},
		{"ASR #32 negative", 0xE1B00041, 0x80000000, 0xFFFFFFFF, true},
		{"ASR #32 positive", 0xE1B00041, 0x40000000, 0, false},
		{"RRX", 0xE1B00061, 0x00000003, 0x80000001, true}, // carry in = 1
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Carry in is the opposite of the expected carry, except RRX which needs C=1.
			carryIn := !tt.wantCarry || tt.name == "RRX"
			v := runShift(t, tt.opcode, tt.r1, 0, carryIn)
			if v.CPU.R[0] != tt.wantR0 || v.CPU.CPSR.C != tt.wantCarry {
				t.Errorf("R0=0x%08X C=%v, want R0=0x%08X C=%v", v.CPU.R[0], v.CPU.CPSR.C, tt.wantR0, tt.wantCarry)
			}
		})
	}
}

func TestRegisterShiftByZeroLeavesOperand(t *testing.T) {
	const value = 0x80000001
	for _, tc := range []struct {
		name   string
		opcode uint32
	}{
		{"LSL R2", 0xE1B00211},
		{"LSR R2", 0xE1B00231},
		{"ASR R2", 0xE1B00251},
		{"ROR R2", 0xE1B00271},
	} {
		for _, carryIn := range []bool{false, true} {
			v := runShift(t, tc.opcode, value, 0, carryIn)
			if v.CPU.R[0] != value || v.CPU.CPSR.C != carryIn {
				t.Errorf("%s by 0 (C=%v): R0=0x%08X C=%v, want R0=0x%08X with carry unchanged",
					tc.name, carryIn, v.CPU.R[0], v.CPU.CPSR.C, value)
			}
		}
	}
}

func TestRegisterShiftByThirtyTwo(t *testing.T) {
	v := runShift(t, 0xE1B00231, 0x80000000, 32, false) // MOVS R0, R1, LSR R2
	if v.CPU.R[0] != 0 || !v.CPU.CPSR.C {
		t.Errorf("LSR by 32: R0=0x%08X C=%v, want 0 and C=1", v.CPU.R[0], v.CPU.CPSR.C)
	}
}

func TestLoadOffsetRRX(t *testing.T) {
	const base = 0x20000
	v := vm.NewVM()
	setupDataWrite(v)
	if err := v.Memory.WriteWord(base+4, 0xCAFEF00D); err != nil {
		t.Fatal(err)
	}
	loadWords(t, v, 0xE7910062) // LDR R0, [R1, R2, RRX]
	v.CPU.R[1] = base
	v.CPU.R[2] = 8 // RRX with C=0 -> 4
	if err := v.Step(); err != nil {
		t.Fatalf("Step failed: %v", err)
	}
	if v.CPU.R[0] != 0xCAFEF00D {
		t.Errorf("expected load from base+4, got R0=0x%08X", v.CPU.R[0])
	}
}
