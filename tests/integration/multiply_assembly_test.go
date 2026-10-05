package integration

import "testing"

func TestMultiplyWithDestinationAsFirstOperandRuns(t *testing.T) {
	src := ".org 0x8000\n_start:\n    MOV R0, #6\n    MOV R1, #7\n    MUL R0, R0, R1\n" +
		"    MOV R2, #3\n    MOV R3, #5\n    MOV R4, #100\n    MLA R2, R2, R3, R4\n    SWI #0\n"
	checkRegisters(t, assembleAndRun(t, src), map[int]uint32{0: 42, 2: 115})
}
