package integration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/loader"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// assembleAndRun parses src, loads it at _start and runs it to the exit SWI.
func assembleAndRun(t *testing.T, src string) *vm.VM {
	t.Helper()
	const cycleLimit = 100000
	program, err := parser.NewParser(src, "test.s").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	start, ok := program.SymbolTable.Lookup("_start")
	if !ok {
		t.Fatal("no _start")
	}
	machine := vm.NewVM()
	machine.CycleLimit = cycleLimit
	if err := machine.InitializeStack(vm.StackSegmentStart + vm.StackSegmentSize); err != nil {
		t.Fatal(err)
	}
	if err := loader.LoadProgramIntoVM(machine, program, start.Value); err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := machine.Run(); err != nil && machine.State != vm.StateHalted {
		t.Fatalf("run: %v", err)
	}
	return machine
}

func checkRegisters(t *testing.T, machine *vm.VM, want map[int]uint32) {
	t.Helper()
	for reg, value := range want {
		if machine.CPU.R[reg] != value {
			t.Errorf("R%d = 0x%08X, want 0x%08X", reg, machine.CPU.R[reg], value)
		}
	}
}

// The code segment is 64 KB, so 8000 bytes of padding keeps everything mapped while
// putting the end of the program out of LDR's 4 KB reach.
const farPadding = "    .space 8000\n"

func TestLiteralGoesToNearbyLtorgPool(t *testing.T) {
	src := ".org 0x8000\n_start:\n    LDR R0, =0x12345678\n    B over\n    .ltorg\n" +
		farPadding + "over:\n    SWI #0\n"
	checkRegisters(t, assembleAndRun(t, src), map[int]uint32{0: 0x12345678})
}

func TestSameLiteralInTwoDistantPools(t *testing.T) {
	src := ".org 0x8000\n_start:\n    LDR R0, =0x12345678\n    B over\n    .ltorg\n" +
		farPadding + "over:\n    LDR R1, =0x12345678\n    SWI #0\n    .ltorg\n"
	checkRegisters(t, assembleAndRun(t, src), map[int]uint32{0: 0x12345678, 1: 0x12345678})
}

func TestBackwardReferenceToLtorgPool(t *testing.T) {
	src := ".org 0x8000\n_start:\n    B main\n    .ltorg\nmain:\n    LDR R0, =0xCAFEF00D\n    B over\n" +
		farPadding + "over:\n    SWI #0\n"
	checkRegisters(t, assembleAndRun(t, src), map[int]uint32{0: 0xCAFEF00D})
}

func TestFullLtorgPoolLeavesFollowingCodeIntact(t *testing.T) {
	const literalCount = 40 // more than the parser's default per-pool estimate
	var b strings.Builder
	b.WriteString(".org 0x8000\n_start:\n")
	for i := range literalCount {
		fmt.Fprintf(&b, "    LDR R0, =0x%08X\n", 0x10000001+uint32(i)*0x10001)
	}
	b.WriteString("    B after\n    .ltorg\nafter:\n    MOV R5, #7\n    MOV R6, #9\n    SWI #0\n")
	last := 0x10000001 + uint32(literalCount-1)*0x10001
	checkRegisters(t, assembleAndRun(t, b.String()), map[int]uint32{0: last, 5: 7, 6: 9})
}

func TestLiteralsWithoutLtorgGoAfterProgram(t *testing.T) {
	src := ".org 0x8000\n_start:\n    LDR R0, =0x12345678\n    LDR R1, =0x9ABCDEF0\n    SWI #0\n"
	checkRegisters(t, assembleAndRun(t, src), map[int]uint32{0: 0x12345678, 1: 0x9ABCDEF0})
}

func TestDataAfterLtorgPoolMatchesItsLabel(t *testing.T) {
	src := ".org 0x8000\n_start:\n    LDR R0, =0x12345678\n    B over\n    .ltorg\n" +
		"value:\n    .word 0x0BADF00D\nmsg:\n    .asciz \"ok\"\n    .align 2\n" +
		"over:\n    LDR R1, =value\n    LDR R2, [R1]\n    LDR R3, =msg\n    LDRB R4, [R3]\n    SWI #0\n"
	checkRegisters(t, assembleAndRun(t, src), map[int]uint32{0: 0x12345678, 2: 0x0BADF00D, 4: 'o'})
}
