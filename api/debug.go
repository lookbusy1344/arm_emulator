package api

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var apiLog *log.Logger

func init() {
	// Check if debug logging is enabled via environment variable
	if os.Getenv("ARM_EMULATOR_DEBUG") != "" {
		// Create debug log file.
		// Note: File handle intentionally not closed - kept open for process lifetime.
		// This is acceptable for debug logging; the OS cleans up on process exit.
		logPath := filepath.Join(os.TempDir(), "arm-emulator-api-debug.log")
		f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600) // #nosec G304 -- fixed filename in temp dir
		if err != nil {
			apiLog = log.New(os.Stderr, "API: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
		} else {
			apiLog = log.New(f, "API: ", log.Ltime|log.Lmicroseconds|log.Lshortfile)
		}
	} else {
		// Disable logging by default
		apiLog = log.New(io.Discard, "", 0)
	}
}

// debugLog logs a message if debug logging is enabled
func debugLog(format string, args ...interface{}) {
	apiLog.Printf(format, args...)
}
