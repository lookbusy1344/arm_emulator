package loader_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/loader"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

func load(t *testing.T, source string) (*vm.VM, *parser.Program) {
	t.Helper()
	program, err := parser.NewParser(source, "test.s").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	machine := vm.NewVM()
	if err := loader.LoadProgramIntoVM(machine, program, 0x8000); err != nil {
		t.Fatalf("load error: %v", err)
	}
	return machine, program
}

func symbol(t *testing.T, program *parser.Program, name string) uint32 {
	t.Helper()
	addr, err := program.SymbolTable.Get(name)
	if err != nil {
		t.Fatal(err)
	}
	return addr
}

func TestWordDirectiveValues(t *testing.T) {
	machine, program := load(t, `.org 0x8000
_start:
	SWI #0x00
words:
	.word -1, -2147483648, 0b101, 0o17, 0xFFFFFFFF, 4294967295, 42, words
`)
	base := symbol(t, program, "words")
	want := []uint32{0xFFFFFFFF, 0x80000000, 5, 15, 0xFFFFFFFF, 0xFFFFFFFF, 42, base}
	for i, w := range want {
		got, err := machine.Memory.ReadWord(base + uint32(i)*4)
		if err != nil {
			t.Fatal(err)
		}
		if got != w {
			t.Errorf(".word element %d = 0x%08X, want 0x%08X", i, got, w)
		}
	}
}

func TestByteDirectiveValues(t *testing.T) {
	machine, program := load(t, `.org 0x8000
_start:
	SWI #0x00
bytes:
	.byte -1, -128, 0b11, 0x7F, 255, 'A'
`)
	base := symbol(t, program, "bytes")
	want := []byte{0xFF, 0x80, 3, 0x7F, 0xFF, 'A'}
	got, err := machine.Memory.GetBytes(base, uint32(len(want)))
	if err != nil {
		t.Fatal(err)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf(".byte element %d = 0x%02X, want 0x%02X", i, got[i], want[i])
		}
	}
}

func TestWordDirectiveRejectsInvalidValue(t *testing.T) {
	program, err := parser.NewParser(".org 0x8000\n_start:\n\tSWI #0\n\t.word undefined_label\n", "test.s").Parse()
	if err != nil {
		return // rejected at parse time
	}
	if err := loader.LoadProgramIntoVM(vm.NewVM(), program, 0x8000); err == nil {
		t.Error("expected an error for an undefined .word symbol")
	}
}
