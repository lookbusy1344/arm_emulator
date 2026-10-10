package api

import "github.com/lookbusy1344/arm-emulator/config"

var apiLog = config.DebugLogger("arm-emulator-api-debug.log", "API: ")

// debugLog logs a message if debug logging is enabled
func debugLog(format string, args ...interface{}) {
	apiLog.Printf(format, args...)
}
