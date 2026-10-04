package encoder_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/loader"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestEncodePSRTransfer(t *testing.T) {
	enc := newTestEncoder()
	tests := []struct {
		mnemonic string
		operands []string
		want     uint32
	}{
		{"MRS", []string{"R0", "CPSR"}, 0xE10F0000},
		{"MRS", []string{"R12", "cpsr"}, 0xE10FC000},
		{"MSR", []string{"CPSR_f", "R0"}, 0xE128F000},
		{"MSR", []string{"CPSR_flg", "R1"}, 0xE128F001},
		{"MSR", []string{"CPSR_fc", "R2"}, 0xE129F002},
		{"MSR", []string{"CPSR", "R3"}, 0xE129F003},
		{"MSR", []string{"CPSR_all", "R3"}, 0xE129F003},
		{"MSR", []string{"CPSR_f", "#0xF0000000"}, 0xE328F4F0}, // 0xF0 ROR 8
		{"MSR", []string{"CPSR_f", "#0"}, 0xE328F000},
	}
	for _, tt := range tests {
		if got := encodeInstruction(t, enc, tt.mnemonic, tt.operands, 0x8000); got != tt.want {
			t.Errorf("%s %v = 0x%08X, want 0x%08X", tt.mnemonic, tt.operands, got, tt.want)
		}
	}
}

func TestEncodePSRTransferErrors(t *testing.T) {
	enc := newTestEncoder()
	tests := []struct {
		mnemonic string
		operands []string
	}{
		{"MRS", []string{"R0"}},
		{"MRS", []string{"R0", "SPSR"}},
		{"MRS", []string{"PC", "CPSR"}},
		{"MRS", []string{"R0", "R1"}},
		{"MSR", []string{"CPSR_f"}},
		{"MSR", []string{"SPSR_f", "R0"}},
		{"MSR", []string{"CPSR_q", "R0"}},
		{"MSR", []string{"CPSR_f", "PC"}},
		{"MSR", []string{"CPSR_f", "#0x101"}},
	}
	for _, tt := range tests {
		inst := &parser.Instruction{Mnemonic: tt.mnemonic, Operands: tt.operands}
		if _, err := enc.EncodeInstruction(inst, 0x8000); err == nil {
			t.Errorf("%s %v: expected an error", tt.mnemonic, tt.operands)
		}
	}
}

func TestPSRTransferRoundTrip(t *testing.T) {
	program, err := parser.NewParser(`.org 0x8000
_start:
	MSR CPSR_f, #0xF0000000
	MRS R0, CPSR
	MOV R1, #0
	MSR CPSR_f, R1
	MRS R2, CPSR
	CMP R1, #0
	MRSEQ R3, CPSR
	SWI #0x00
`, "test.s").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	machine := vm.NewVM()
	if err := loader.LoadProgramIntoVM(machine, program, 0x8000); err != nil {
		t.Fatalf("load error: %v", err)
	}
	machine.CPU.PC = 0x8000
	_ = machine.Run()

	const zFlag = 0x40000000
	want := map[int]uint32{0: 0xF0000000, 2: 0, 3: zFlag | 0x20000000}
	for reg, w := range want {
		if got := machine.CPU.R[reg]; got != w {
			t.Errorf("R%d = 0x%08X, want 0x%08X", reg, got, w)
		}
	}
}
