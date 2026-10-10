// Package sessiondirs checks that tests destroy the API sessions they create.
package sessiondirs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// sessionDirGlob matches the temporary directory the session manager creates for each
// session without an fsRoot.
const sessionDirGlob = "arm-emulator-session-*"

// tempDirVars are the variables os.TempDir reads: TMPDIR on Unix, TMP and TEMP on
// Windows.
var tempDirVars = []string{"TMPDIR", "TMP", "TEMP"}

// Run runs the tests with a private temporary directory and fails the run if any
// session directory remains in it afterwards. Call it from TestMain.
func Run(m *testing.M) int {
	dir, err := os.MkdirTemp("", "arm-emulator-tests-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sessiondirs: %v\n", err)
		return 1
	}
	defer func() { _ = os.RemoveAll(dir) }()
	for _, v := range tempDirVars {
		if err := os.Setenv(v, dir); err != nil {
			fmt.Fprintf(os.Stderr, "sessiondirs: %v\n", err)
			return 1
		}
	}

	code := m.Run()

	leaked, err := filepath.Glob(filepath.Join(dir, sessionDirGlob))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sessiondirs: %v\n", err)
		return 1
	}
	if len(leaked) > 0 {
		fmt.Fprintf(os.Stderr, "%d session directories left behind; destroy sessions or shut down the server in t.Cleanup\n", len(leaked))
		if code == 0 {
			code = 1
		}
	}
	return code
}
