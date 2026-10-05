package parser_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// addExampleSeeds adds every example program to the fuzz corpus.
func addExampleSeeds(f *testing.F) {
	f.Helper()
	paths, err := filepath.Glob(filepath.Join("..", "..", "..", "examples", "*.s"))
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
}

// FuzzParse checks that no input makes the parser panic.
func FuzzParse(f *testing.F) {
	addExampleSeeds(f)
	for _, seed := range []string{
		"", "\n", ".balign 0", ".align 40", ".space -4", "LDR R0, [R1", "PUSH {R0",
		".equ X, (1<<4)|2", ".macro m\n.endm\n", "x: .word x+4", "MOV R0, #'", ".asciz \"\\x",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, src string) {
		_, _ = parser.NewParser(src, "fuzz.s").Parse()
	})
}

// FuzzEvaluateExpression checks that the constant evaluator never panics and agrees
// with ParseNumber on plain numbers.
func FuzzEvaluateExpression(f *testing.F) {
	for _, seed := range []string{
		"0", "42", "-1", "0x10+0b11", "(1<<4)|(1<<1)", "A*2+1", "~0", "17%5", "4/0",
		"1<<32", "((((1))))", "(", ")", "--1", "A+", "0x", "1 2",
	} {
		f.Add(seed)
	}
	lookup := func(name string) (uint32, bool) {
		if name == "A" {
			return 4, true
		}
		return 0, false
	}
	f.Fuzz(func(t *testing.T, expr string) {
		got, err := parser.EvaluateExpression(expr, lookup)
		if want, perr := parser.ParseNumber(expr); perr == nil && isPlainNumber(expr) {
			if err != nil || got != want {
				t.Errorf("EvaluateExpression(%q) = %d, %v; ParseNumber = %d", expr, got, err, want)
			}
		}
	})
}

// isPlainNumber reports whether s is an unsigned decimal literal without spaces.
func isPlainNumber(s string) bool {
	_, err := strconv.ParseUint(s, 10, 32)
	return err == nil && s[0] != '+'
}
