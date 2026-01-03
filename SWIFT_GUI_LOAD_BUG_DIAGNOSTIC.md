# Swift GUI Load Bug Diagnostic Report

**Date:** 2026-01-03
**Issue:** Swift GUI fails to load examples/hello.s with decode error
**Error Message:** "Failed to load program: Failed to decode response: The data couldn't be read because it is missing."

## Problem Description

When attempting to load `examples/hello.s` via the Swift GUI (using Examples browser or File > Open), the app displays an error dialog:
```
Failed to load program: Failed to decode response: The data couldn't be read because it is missing.
```

## Investigation Results

### Backend Verification ✅

The Go backend API is working correctly:

```bash
# Backend is running on port 8080
$ lsof -i :8080
arm-emula 74229 xc  6u  IPv4 ... TCP localhost:http-alt (LISTEN)

# Backend returns correct response
$ curl -X POST http://localhost:8080/api/v1/session -d '{}'
{"sessionId":"86a943f58d65723e88a5324561c61276","createdAt":"2026-01-03T10:15:47.633213Z"}

$ curl -X POST http://localhost:8080/api/v1/session/SESS_ID/load \
  -H "Content-Type: application/json" \
  -d '{"source": "        .org 0x8000\n_start:\n        MOV R0, #0\n        SWI #0x00\n"}'
{"success":true,"symbols":{"_start":32768}}
```

**Conclusion:** Backend returns valid JSON with structure:
```json
{
  "success": true,
  "symbols": {"_start": 32768}
}
```

### Root Cause Identified ✅

**Type Mismatch Between Swift Client and Go Backend:**

**Go Backend** (`api/handlers.go:149-154`):
```go
response := LoadProgramResponse{
    Success: true,
    Symbols: symbols,
}
writeJSON(w, http.StatusOK, response)
```

**Go Model** (`api/models.go:40-44`):
```go
type LoadProgramResponse struct {
    Success bool              `json:"success"`
    Errors  []string          `json:"errors,omitempty"`
    Symbols map[string]uint32 `json:"symbols,omitempty"`
}
```

**Swift Client** (`ARMEmulator/Services/APIClient.swift:62-68` - BEFORE FIX):
```swift
func loadProgram(sessionID: String, source: String) async throws {
    struct LoadProgramRequest: Codable {
        let source: String
    }
    let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/load")
    try await post(url: url, body: LoadProgramRequest(source: source))
}
```

This calls the void-returning `post` function which expects `EmptyResponse` (empty struct `{}`), but receives `{"success": true, "symbols": {...}}`, causing JSONDecoder to fail.

## Attempted Fixes

### Fix #1: Add LoadProgramResponse Type ❌ FAILED

**Changes Made:**

1. Added `LoadProgramResponse` struct to Swift (`APIClient.swift:332-336`):
```swift
struct LoadProgramResponse: Codable {
    let success: Bool
    let errors: [String]?
    let symbols: [String: UInt32]?
}
```

2. Updated function signature (`APIClient.swift:62`):
```swift
func loadProgram(sessionID: String, source: String) async throws -> LoadProgramResponse {
    // ...
    return try await post(url: url, body: LoadProgramRequest(source: source))
}
```

3. Updated ViewModel to handle response (`EmulatorViewModel.swift:66-73`):
```swift
let response = try await apiClient.loadProgram(sessionID: sessionID, source: source)
if !response.success {
    let errors = response.errors?.joined(separator: "\n") ?? "Unknown error"
    errorMessage = "Failed to load program:\n\(errors)"
    return
}
```

**Build:**
```bash
$ xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator clean build
Build Succeeded ✅
```

**Result:** User still gets exact same error ❌

**Problem:** Despite successful build, app continues using old binary. Multiple rebuild attempts didn't resolve.

### Fix #2: Add Debug Logging ❌ FAILED

**Attempt A: File Logging**

Added file logging to `/tmp/arm-api-debug.log` in `performRequest()`:
```swift
let logPath = "/tmp/arm-api-debug.log"
// ... write HTTP status, response body, expected type
```

**Result:** Log file never created. Suspected sandboxing issue.

**Attempt B: NSLog**

Replaced file logging with NSLog:
```swift
NSLog("🔴 API REQUEST: \(request.httpMethod ?? "?") \(request.url?.path ?? "unknown")")
NSLog("🔴 HTTP \(httpResponse.statusCode)")
NSLog("🔴 Response body: \(String(data: data, encoding: .utf8) ?? "<binary>")")
```

**Result:** No console output captured via `log stream` or Console.app

**Problem:** Debug output never appears, suggesting:
- App is running old binary
- Logging code not being executed
- Output being suppressed

### Fix #3: Multiple Rebuild Attempts ❌ FAILED

**Attempts:**
1. `xcodebuild clean build` (multiple times)
2. Killed all ARMEmulator processes before rebuild
3. Verified binary timestamps
4. Launched specific build from DerivedData path

**Timestamp Issues Discovered:**
```bash
# Source file modified AFTER binary was built:
$ stat ARMEmulator/Services/APIClient.swift
2026-01-03 10:34:52

$ stat .../DerivedData/.../ARMEmulator.app/Contents/MacOS/ARMEmulator
2026-01-03 10:28:00
```

**Problem:** Changes keep being made to source files, but rebuilds don't seem to incorporate them into running app.

## Current Status

### What We Know ✅
1. Backend API works correctly and returns proper JSON
2. Root cause is type mismatch (EmptyResponse vs LoadProgramResponse)
3. Correct fix has been implemented in source code
4. Code compiles successfully with no errors

### What's Not Working ❌
1. Running app doesn't reflect source code changes
2. Debug logging produces no output
3. User continues getting same decode error
4. Binary timestamps show old builds being executed

### Possible Causes
1. **Xcode caching issue** - DerivedData corruption
2. **Multiple app instances** - Old version cached somewhere
3. **Code signing/entitlements** - Sandbox preventing file writes
4. **Build configuration** - Debug vs Release mismatch
5. **App location** - User running app from different location than DerivedData

## Next Steps

### Recommended Actions

1. **Clean DerivedData completely:**
   ```bash
   rm -rf ~/Library/Developer/Xcode/DerivedData/ARMEmulator-*
   xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator clean build
   ```

2. **Verify single app instance:**
   ```bash
   killall ARMEmulator
   ps aux | grep ARMEmulator  # Should show nothing
   ```

3. **Run from Xcode directly:**
   - Open project in Xcode
   - Product > Clean Build Folder (Cmd+Shift+K)
   - Product > Run (Cmd+R)
   - Try loading file within Xcode-launched app

4. **Check app bundle location:**
   ```bash
   # Find all ARMEmulator.app bundles
   find ~ -name "ARMEmulator.app" -type d 2>/dev/null
   ```

5. **Verify build settings:**
   - Check if Debug configuration is being used
   - Verify code signing settings
   - Check for any custom build scripts

### Alternative Approaches

If rebuild issues persist:

1. **Direct binary replacement:**
   - Build fresh binary
   - Manually copy to known app location
   - Verify timestamps match

2. **Test with minimal reproduction:**
   - Create simple Swift script that tests LoadProgramResponse decoding
   - Verify type definitions are correct

3. **Check for framework caching:**
   - Swift may be caching compiled modules
   - Try `swift package clean` if using SPM

## Files Modified

- `swift-gui/ARMEmulator/Services/APIClient.swift` - Lines 62-68, 332-336, 285-319
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift` - Lines 66-73

## Related Files

- `api/handlers.go` - Backend load endpoint (line 88-155)
- `api/models.go` - LoadProgramResponse definition (line 40-44)
- `swift-gui/ARMEmulator/Views/MainView.swift` - File loading UI (line 63-70)
- `swift-gui/ARMEmulator/Views/FileCommands.swift` - File menu commands (line 12-18)

## Test Commands

```bash
# Verify backend response format
curl -s http://localhost:8080/api/v1/session -X POST -d '{}'
SESSION_ID=$(curl -s http://localhost:8080/api/v1/session -X POST -d '{}' | grep -o '"sessionId":"[^"]*"' | cut -d'"' -f4)
curl -s "http://localhost:8080/api/v1/session/$SESSION_ID/load" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"source": "        .org 0x8000\n_start:\n        MOV R0, #0\n        SWI #0x00\n"}'

# Check running processes
ps aux | grep "[A]RMEmulator"
lsof -i :8080

# Verify binary timestamps
stat -f "%Sm %N" -t "%Y-%m-%d %H:%M:%S" .../ARMEmulator.app/Contents/MacOS/ARMEmulator
stat -f "%Sm %N" -t "%Y-%m-%d %H:%M:%S" ARMEmulator/Services/APIClient.swift
```

## Conclusion

The fix is correct but isn't being executed. The issue appears to be with the build/deploy process rather than the code itself. The app is running an old binary that doesn't include the LoadProgramResponse changes.

**Priority:** Ensure fresh builds are actually being used before attempting further code changes.
