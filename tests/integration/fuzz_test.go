package integration

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/loader"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// FuzzAssembleAndRun pushes source through the parser, loader and VM. No stage may
// panic, and a run may end only by halting, a breakpoint or the error state.
func FuzzAssembleAndRun(f *testing.F) {
	const cycleLimit = 2000
	paths, err := filepath.Glob(filepath.Join("..", "..", "examples", "*.s"))
	if err != nil {
		f.Fatal(err)
	}
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(string(src))
	}
	for _, seed := range []string{
		"_start:\n LDR R0, =0x12345678\n B x\n .ltorg\nx:\n SWI #0\n",
		"_start:\n LDRH R0, [PC, #-4]\n STRH R0, [SP, #-2]!\n SWI #0\n",
		".equ K, 3\n_start:\n MOV R0, #-K\n MUL R1, R1, R0\n SWI #0\n",
		"_start:\n LDR R0, v\n SWI #0\nv: .half 1, 2\n .byte -1\n .align 2\n",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, src string) {
		program, err := parser.NewParser(src, "fuzz.s").Parse()
		if err != nil {
			return
		}
		machine := vm.NewVM()
		machine.CycleLimit = cycleLimit
		machine.OutputWriter = discardWriter{}
		machine.SetStdinReader(strings.NewReader("1\nab\n"))
		if err := machine.InitializeStack(vm.StackSegmentStart + vm.StackSegmentSize); err != nil {
			t.Fatal(err)
		}
		entry := uint32(vm.CodeSegmentStart)
		if start, ok := program.SymbolTable.Lookup("_start"); ok && start.Defined {
			entry = start.Value
		}
		if err := loader.LoadProgramIntoVM(machine, program, entry); err != nil {
			return
		}
		err = machine.Run()
		switch {
		case errors.Is(err, vm.ErrProgramExited):
			if machine.State != vm.StateHalted {
				t.Fatalf("exit left State = %v", machine.State)
			}
		case errors.Is(err, vm.ErrBreakpointHit):
		case err != nil && machine.State != vm.StateError:
			t.Fatalf("run failed (%v) but State = %v", err, machine.State)
		}
	})
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
