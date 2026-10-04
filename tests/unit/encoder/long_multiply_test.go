package encoder_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/loader"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestEncodeLongMultiply(t *testing.T) {
	tests := []struct {
		line string
		want uint32
	}{
		{"UMULL R0, R1, R2, R3", 0xE0810392},
		{"UMULLS R0, R1, R2, R3", 0xE0910392},
		{"UMLAL R4, R5, R6, R7", 0xE0A54796},
		{"SMULL R0, R1, R2, R3", 0xE0C10392},
		{"SMLAL R0, R1, R2, R3", 0xE0E10392},
		{"SMLALS R0, R1, R2, R3", 0xE0F10392},
		{"UMULLNE R0, R1, R2, R3", 0x10810392},
		{"SMLALEQ R0, R1, R2, R3", 0x00E10392},
	}
	for _, tt := range tests {
		got, err := encodeSource(t, tt.line)
		if err != nil {
			t.Errorf("%s: %v", tt.line, err)
			continue
		}
		if got != tt.want {
			t.Errorf("%s = 0x%08X, want 0x%08X", tt.line, got, tt.want)
		}
	}
}

func TestEncodeLongMultiplyErrors(t *testing.T) {
	for _, line := range []string{
		"UMULL R0, R1, R2",     // too few operands
		"UMULL R0, R0, R2, R3", // RdLo == RdHi
		"SMULL R2, R1, R2, R3", // RdLo == Rm
		"UMULL R0, R2, R2, R3", // RdHi == Rm
		"UMULL R0, PC, R2, R3", // PC
		"UMULL R0, R1, R2, #3", // immediate
	} {
		if _, err := encodeSource(t, line); err == nil {
			t.Errorf("%s: expected an error", line)
		}
	}
}

func TestLongMultiplyRoundTrip(t *testing.T) {
	program, err := parser.NewParser(`.org 0x8000
_start:
	MVN R2, #0          ; 0xFFFFFFFF
	MOV R3, #2
	UMULL R0, R1, R2, R3
	SMULL R4, R5, R2, R3
	MOV R6, #10
	MOV R7, #0
	UMLAL R6, R7, R2, R3
	MOV R8, #10
	MOV R9, #0
	SMLAL R8, R9, R2, R3
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

	want := map[int]uint32{
		0: 0xFFFFFFFE, 1: 0x00000001, // 0xFFFFFFFF * 2 unsigned
		4: 0xFFFFFFFE, 5: 0xFFFFFFFF, // -1 * 2 signed
		6: 0x00000008, 7: 0x00000002, // 0x1FFFFFFFE + 10
		8: 0x00000008, 9: 0x00000000, // -2 + 10
	}
	for reg, w := range want {
		if got := machine.CPU.R[reg]; got != w {
			t.Errorf("R%d = 0x%08X, want 0x%08X", reg, got, w)
		}
	}
}
