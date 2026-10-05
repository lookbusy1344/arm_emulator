package integration_test

import (
	"strings"
	"testing"
)

// The output exceeds a pipe buffer (64 KiB on macOS and Linux).
func TestRunAssemblyCapturesLargeOutput(t *testing.T) {
	const lines = 20000 // "0123456789\n" per line: 220,000 bytes
	src := `.org 0x8000
_start:
    LDR R4, =20000
loop:
    LDR R0, =line
    SWI #0x02
    SUBS R4, R4, #1
    BNE loop
    MOV R0, #0
    SWI #0x00
line:
    .asciz "0123456789\n"
`
	stdout, _, exitCode, err := runAssembly(t, src)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0", exitCode)
	}
	if got := strings.Count(stdout, "0123456789\n"); got != lines {
		t.Errorf("captured %d lines, want %d", got, lines)
	}
}
