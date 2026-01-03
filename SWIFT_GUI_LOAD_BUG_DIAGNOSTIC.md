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

### Fix #3: Nuclear DerivedData Clean ✅ PARTIALLY SUCCESSFUL

**Actions Taken:**

1. Deleted all DerivedData:
```bash
rm -rf ~/Library/Developer/Xcode/DerivedData/ARMEmulator-*
```

2. Regenerated Xcode project:
```bash
cd swift-gui
xcodegen generate
```

3. Opened in Xcode and rebuilt (Product > Run)

**Result:**
- ✅ LoadProgramResponse fix now working - HTTP 200 response received
- ✅ No more decode error dialog
- ✅ sourceCode property being set (550 characters confirmed)
- ❌ New issue: Text loaded but not visible in editor

**New Error Discovered:**
```
🔴 HTTP 200
🔴 Response body: {"sessionId":"...","state":"halted","pc":32768,"cycles":0,"hasWrite":false}
```
Backend doesn't return `instruction` field, but Swift `VMStatus` struct expected it as non-optional.

**Fix Applied:**
Made `instruction` optional in `ProgramState.swift`:
```swift
struct VMStatus: Codable {
    var instruction: String?  // Optional - backend doesn't always return this
    ...
}
```

### Fix #4: NSTextView Rendering Investigation ❌ IN PROGRESS

**Problem:** Text loads successfully into storage but doesn't render visually.

**Diagnostic Evidence:**
```
🟢 Setting sourceCode to 550 characters
🟢 sourceCode is now: ; hello.s - Classic "Hello World" program...
🔵 updateNSView called - current: 0 chars, new: 550 chars
🔵 Updating text view content
🔵 Text storage length: 550
🔵 Text storage string: ; hello.s - Classic "Hello World" program...
```

**Initial Frame Diagnostics (before fix):**
```
🔵 TextView frame: (0.0, 0.0, 0.0, 264.5)      ❌ Zero width!
🔵 TextView bounds: (0.0, 0.0, 0.0, 264.5)
🔵 ScrollView frame: (0.0, 0.0, 449.5, 264.5)  ✅ Correct
🔵 Container size: (-10.0, 10000000.0)         ❌ Negative width!
```

**Root Cause #1:** NSTextView had zero width, preventing rendering.

**Attempted Fix #4A: Add NSTextView Sizing Configuration**

Added proper text view configuration in `EditorView.swift` `makeNSView()`:
```swift
// Configure text view sizing for scroll view
textView.minSize = NSSize(width: 0.0, height: scrollView.contentSize.height)
textView.maxSize = NSSize(width: CGFloat.greatestFiniteMagnitude, height: CGFloat.greatestFiniteMagnitude)
textView.isVerticallyResizable = true
textView.isHorizontallyResizable = false
textView.autoresizingMask = [.width]

// Configure text container
if let textContainer = textView.textContainer {
    textContainer.containerSize = NSSize(width: scrollView.contentSize.width, height: CGFloat.greatestFiniteMagnitude)
    textContainer.widthTracksTextView = true
}
```

**Result:** Frame fixed but text still not visible ❌

**Frame Diagnostics (after Fix #4A):**
```
🔵 TextView frame: (0.0, 0.0, 449.5, 247.5)    ✅ Correct width!
🔵 TextView bounds: (0.0, 0.0, 449.5, 247.5)
🔵 ScrollView frame: (0.0, 0.0, 449.5, 264.5)
🔵 Container size: (439.5, 1.7976931348623157e+308)  ✅ Correct!
🔵 Text storage length: 550                    ✅ Has content
```

**Visual Evidence:** Screenshot shows:
- Line numbers 1-14 visible in gutter
- Main text area completely blank
- Horizontal scrollbar present but not functional
- Line numbers don't scroll when scrolling editor pane

**Attempted Fix #4B: Non-Wrapping Text Editor Configuration**

Changed configuration for code editor style (non-wrapping, horizontally scrollable):
```swift
// Configure text view sizing for scroll view (non-wrapping, horizontally scrollable)
textView.isVerticallyResizable = true
textView.isHorizontallyResizable = true        // Changed from false
textView.autoresizingMask = []                 // Changed from [.width]

// Configure text container for non-wrapping text
if let textContainer = textView.textContainer {
    textContainer.containerSize = NSSize(width: CGFloat.greatestFiniteMagnitude, height: CGFloat.greatestFiniteMagnitude)
    textContainer.widthTracksTextView = false  // Changed from true
}
```

**Result:** Text still not visible ❌

**Hypothesis:** Despite correct frame sizes and text storage containing content, the NSTextView is not rendering glyphs. Possible causes:
1. Layout manager not generating glyphs
2. Font configuration issue
3. Text color transparency
4. Glyph rendering disabled
5. View hierarchy z-order issue

## Current Status

### What We Know ✅
1. Backend API works correctly and returns proper JSON
2. Type mismatch fix (LoadProgramResponse) is working
3. VMStatus optional instruction field fix is working
4. Program loading succeeds (HTTP 200)
5. sourceCode property is set correctly (550 characters)
6. Text storage contains correct content
7. TextView frame/bounds are correct (449.5 x 247.5)
8. Text container size is correct (439.5 width)
9. updateNSView is called and updates text
10. NSLog diagnostics are working

### What's Not Working ❌
1. NSTextView not rendering text despite having content
2. Text area appears completely blank
3. Line numbers don't scroll with content
4. Horizontal scrollbar non-functional

### Remaining Issues
1. **Primary:** NSTextView rendering failure (text invisible)
2. **Secondary:** Memory/Disassembly endpoints return 400 errors (invalid address parameter)
3. **Minor:** Excessive debug logging needs cleanup after fix

## Next Steps

### Immediate Actions for NSTextView Rendering

1. **Add layout manager diagnostics:**
   - Check if layout manager has glyphs generated
   - Verify glyph range for text
   - Check if layout is being invalidated

2. **Add font/color diagnostics:**
   - Log actual font being used (could be nil)
   - Log actual text color (could be transparent)
   - Log text view's drawing rect

3. **Test with minimal text view:**
   - Create standalone NSTextView test without SwiftUI wrapper
   - Verify text renders in simple AppKit context
   - If works, issue is SwiftUI integration

4. **Check layout manager generation:**
   ```swift
   if let layoutManager = textView.layoutManager {
       NSLog("🔵 Layout manager: \(layoutManager)")
       NSLog("🔵 Number of glyphs: \(layoutManager.numberOfGlyphs)")
       NSLog("🔵 Used rect: \(layoutManager.usedRect(for: textView.textContainer!))")
   }
   ```

5. **Try explicit sizeToFit and layout invalidation:**
   ```swift
   textView.string = text
   textView.sizeToFit()
   textView.layoutManager?.invalidateLayout(forCharacterRange: NSRange(location: 0, length: text.count), actualCharacterRange: nil)
   textView.layoutManager?.ensureLayout(for: textView.textContainer!)
   ```

### Alternative Approaches

1. **Replace NSViewRepresentable with simpler approach:**
   - Use UIViewRepresentable/NSViewRepresentable without complex updateNSView logic
   - Set text only once in makeNSView
   - Use Coordinator to handle text changes

2. **Use TextEditor instead of custom NSTextView:**
   - SwiftUI's native TextEditor might work better
   - Less control but more reliable rendering

3. **Check if LineNumberGutterView is interfering:**
   - Temporarily remove gutter view
   - Test if text renders without it
   - Ruler view might be covering content area

## Files Modified

- `swift-gui/ARMEmulator/Services/APIClient.swift` - Lines 62-68, 285-310, 332-336
  - Added LoadProgramResponse struct
  - Updated loadProgram to return LoadProgramResponse
  - Added extensive NSLog debugging

- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift` - Lines 66-73, 27, 41-47, 229-237, 244-256
  - Updated loadProgram to handle LoadProgramResponse
  - Check success field before setting sourceCode
  - Added isInitializing flag to prevent concurrent initialization
  - Fixed cleanup to reset state properly
  - Made memory/disassembly errors silent (benign failures)

- `swift-gui/ARMEmulator/Models/ProgramState.swift` - Line 14
  - Made instruction field optional in VMStatus

- `swift-gui/ARMEmulator/Views/EditorView.swift` - Lines 75-76, 83-94, 105-139
  - Added textColor and backgroundColor settings
  - Added NSTextView sizing configuration (minSize, maxSize, resizing)
  - Added text container configuration
  - Changed to non-wrapping text editor configuration
  - Added extensive frame/bounds/container diagnostics

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

**Progress Made:**
1. ✅ Fixed Xcode build caching issue (DerivedData clean)
2. ✅ Fixed type mismatch (LoadProgramResponse)
3. ✅ Fixed VMStatus optional field
4. ✅ Program loading now succeeds (HTTP 200)
5. ✅ Text correctly loaded into storage
6. ✅ Fixed zero-width NSTextView frame issue

**Current Blocker:**
Despite all diagnostics showing correct values (frame, bounds, container size, text storage), the NSTextView is not rendering any glyphs. The text area appears completely blank even though:
- Frame is 449.5 x 247.5 (correct)
- Container size is 439.5 (correct)
- Text storage has 550 characters (correct)
- updateNSView is being called (confirmed)

**Next Priority:** Investigate layout manager and glyph generation to determine why text with proper frame and storage is not rendering visually.
