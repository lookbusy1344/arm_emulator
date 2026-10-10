package config_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/lookbusy1344/arm-emulator/config"
)

const (
	debugEnv     = "ARM_EMULATOR_DEBUG"
	debugLogName = "test-debug.log"
	ownerOnly    = 0o600
)

// debugLine is one log line: prefix, time with microseconds, file:line, message.
var debugLine = regexp.MustCompile(`^TEST: \d{2}:\d{2}:\d{2}\.\d{6} debug_log_test\.go:\d+: hello\n$`)

// isolateHome points the per-user directories at a fresh directory.
func isolateHome(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("log location test covers the Unix layout")
	}
	t.Setenv("HOME", t.TempDir())
}

func TestDebugLoggerWritesToUserLogDirectory(t *testing.T) {
	isolateHome(t)
	t.Setenv(debugEnv, "1")

	config.DebugLogger(debugLogName, "TEST: ").Print("hello")

	path := filepath.Join(os.Getenv("HOME"), ".local", "share", "arm-emu", "logs", debugLogName)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("debug log not at %s: %v", path, err)
	}
	if !debugLine.Match(content) {
		t.Errorf("debug log = %q, want one line matching %s", content, debugLine)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != ownerOnly {
		t.Errorf("debug log mode = %o, want %o", mode, ownerOnly)
	}
	if _, err := os.Stat(filepath.Join(os.TempDir(), debugLogName)); !os.IsNotExist(err) {
		t.Errorf("debug log also in temp dir: %v", err)
	}
}

func TestDebugLoggerDiscardsWhenDisabled(t *testing.T) {
	isolateHome(t)
	t.Setenv(debugEnv, "")

	config.DebugLogger(debugLogName, "TEST: ").Print("hello")

	path := filepath.Join(os.Getenv("HOME"), ".local", "share", "arm-emu", "logs", debugLogName)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("debug log created while disabled: %v", err)
	}
}
