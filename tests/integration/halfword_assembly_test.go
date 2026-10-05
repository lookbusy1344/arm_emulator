package integration

import "testing"

func TestAssembledHalfwordLoadsAndStores(t *testing.T) {
	src := ".org 0x8000\n_start:\n    LDR R1, =val\n    LDRH R0, [R1]\n    LDRH R4, [R1, #2]\n" +
		"    LDR R3, =0xBEEF\n    STRH R3, [R1, #2]\n    LDR R2, [R1]\n    SWI #0\n" +
		"val:\n    .word 0x56781234\n"
	checkRegisters(t, assembleAndRun(t, src), map[int]uint32{0: 0x1234, 4: 0x5678, 2: 0xBEEF1234})
}
