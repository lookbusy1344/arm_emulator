package integration

import (
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/loader"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestLDRPseudoNegativeValues(t *testing.T) {
	src := ".org 0x8000\n.equ K, 5\n_start:\n    LDR R0, =-1\n    LDR R1, =-0x12345678\n" +
		"    LDR R2, =-K\n    LDR R3, =-2147483648\n    SWI #0\n"
	checkRegisters(t, assembleAndRun(t, src), map[int]uint32{
		0: 0xFFFFFFFF, 1: 0xEDCBA988, 2: 0xFFFFFFFB, 3: 0x80000000,
	})
}

func TestPCRelativeLabelLoadsAndStores(t *testing.T) {
	src := ".org 0x8000\nbefore:\n    .word 0x0BADCAFE\n_start:\n    LDR R0, after\n    LDR R1, before\n" +
		"    LDRB R2, bytes\n    MOV R3, #0x7E\n    STR R3, after\n    LDR R4, after\n    SWI #0\n" +
		"after:\n    .word 0x12345678\nbytes:\n    .byte 0xAB\n"
	checkRegisters(t, assembleAndRun(t, src), map[int]uint32{
		0: 0x12345678, 1: 0x0BADCAFE, 2: 0xAB, 4: 0x7E,
	})
}

func TestPCRelativeLabelOutOfRange(t *testing.T) {
	src := ".org 0x8000\n_start:\n    LDR R0, far\n    SWI #0\n    .space 5000\nfar:\n    .word 1\n"
	program, err := parser.NewParser(src, "test.s").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	err = loader.LoadProgramIntoVM(vm.NewVM(), program, 0x8000)
	if err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Errorf("load error = %v, want out of range", err)
	}
}
