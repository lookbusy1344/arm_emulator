package debugger_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/debugger"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// FuzzDebuggerExpression checks that breakpoint conditions and watch expressions
// never panic, whatever the user types.
func FuzzDebuggerExpression(f *testing.F) {
	for _, seed := range []string{
		"R0", "R0 + R1", "R0 & R1", "[R1]", "*0x20000", "[R1+4] == 0x10", "main", "main+4",
		"PC", "SP", "CPSR", "R0 == 1 && R1 != 2", "(", ")", "[", "]", "R99", "0x", "1/0", "1 % 0",
		"-", "~", "!R0", "R0 << 40", "R0 >> 33", "''", "'a'",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, expr string) {
		machine := vm.NewVM()
		machine.CPU.R[1] = vm.DataSegmentStart
		symbols := map[string]uint32{"main": vm.CodeSegmentStart}
		eval := debugger.NewExpressionEvaluator()
		_, _ = eval.EvaluateExpression(expr, machine, symbols)
		_, _ = eval.Evaluate(expr, machine, symbols)
	})
}
