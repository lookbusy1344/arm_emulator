# Swift GUI Debugging Guide

## Overview

This document captures best practices for debugging the Swift native macOS GUI, including lessons learned from real debugging sessions. The Swift app integrates SwiftUI with AppKit components (NSTextView, NSRulerView) and communicates with a Go backend via HTTP REST API and WebSocket.

## General Debugging Principles

### 1. Simplify First, Add Complexity Later

**The Golden Rule:** When a complex view doesn't work, strip it down to the simplest possible version.

- Comment out custom components (ruler views, overlays, custom drawing)
- Get the basic AppKit view working first
- Add features back incrementally, testing after each addition
- This isolates the problematic component quickly

**Example:** NSTextView not rendering? Comment out the NSRulerView first.

### 2. Verify Build Freshness

Before testing any Swift code changes:

```bash
# Check binary timestamp vs source files
stat -f "%Sm %N" -t "%Y-%m-%d %H:%M:%S" \
  ~/Library/Developer/Xcode/DerivedData/ARMEmulator-*/Build/Products/Debug/ARMEmulator.app/Contents/MacOS/ARMEmulator

stat -f "%Sm %N" -t "%Y-%m-%d %H:%M:%S" \
  swift-gui/ARMEmulator/Services/APIClient.swift
```

**If source is newer than binary, you're testing old code.** Clean DerivedData immediately.

### 3. DerivedData Issues: Nuclear Option First

When code changes don't take effect despite rebuilds:

```bash
# Don't waste time with incremental fixes
# Go nuclear immediately:
rm -rf ~/Library/Developer/Xcode/DerivedData/ARMEmulator-*
cd swift-gui
xcodegen generate
open ARMEmulator.xcodeproj
# Then Cmd+R in Xcode
```

Xcode's build caching can be stubborn. Hours can be wasted on "mysterious" issues that are simply stale builds.

### 4. Backend Verification First

Before debugging Swift GUI issues, verify the backend works:

```bash
# Test backend directly
curl -s http://localhost:8080/api/v1/session -X POST -d '{}'

# Test specific endpoints
SESSION_ID="..."
curl -s "http://localhost:8080/api/v1/session/$SESSION_ID/load" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"source": "..."}'
```

Don't assume the frontend is broken if the backend might be returning bad data.

## Common Issues and Solutions

### Issue 1: NSViewRepresentable Views Don't Render

**Symptoms:**
- View appears blank
- Data is confirmed present (logging shows correct values)
- Frame/bounds are correct
- Text storage has content

**Diagnosis Approach:**

1. **Add diagnostics to updateNSView:**
```swift
func updateNSView(_ view: NSScrollView, context: Context) {
    guard let textView = view.documentView as? NSTextView else { return }

    print("Frame: \(textView.frame)")
    print("Text storage length: \(textView.textStorage?.length ?? 0)")

    if let layoutManager = textView.layoutManager,
       let container = textView.textContainer {
        print("Glyphs: \(layoutManager.numberOfGlyphs)")
        print("Used rect: \(layoutManager.usedRect(for: container))")
    }
}
```

2. **Check for common causes:**
   - Zero-width frame → Add sizing configuration
   - Custom ruler views → Comment out temporarily
   - Missing text color → Set explicitly
   - Layout not invalidated → Force layout refresh

3. **Apply divide-and-conquer:**
```swift
// Comment out suspect components one by one:
// let gutterView = LineNumberGutterView(...)
// scrollView.verticalRulerView = gutterView
// scrollView.hasVerticalRuler = true
```

4. **Verify basic AppKit functionality:**
```swift
// Minimal working configuration
textView.string = text
textView.textColor = .labelColor
textView.backgroundColor = .textBackgroundColor
textView.needsDisplay = true
```

**Solution Pattern:**
Once basic rendering works, add components back one at a time until you find the culprit.

### Issue 2: JSON Decoding Errors

**Symptom:**
```
Failed to decode response: The data couldn't be read because it is missing.
```

**Diagnosis:**

This cryptic error means **type mismatch**, not missing data. The decoder expected one type but received another.

1. **Log the actual response:**
```swift
private func performRequest<T: Decodable>(request: URLRequest) async throws -> T {
    let (data, response) = try await session.data(for: request)

    // Temporary diagnostic
    print("Response: \(String(data: data, encoding: .utf8) ?? "<binary>")")
    print("Expected: \(T.self)")

    return try JSONDecoder().decode(T.self, from: data)
}
```

2. **Compare with backend:**
```bash
# What does backend actually return?
curl -s http://localhost:8080/api/v1/session/$SESSION_ID/load \
  -X POST -d '{"source": "..."}' | jq
```

3. **Fix the type mismatch:**
```swift
// Before: Expected EmptyResponse {}
func loadProgram(...) async throws {
    try await post(url: url, body: request)  // Returns EmptyResponse
}

// After: Match backend response
struct LoadProgramResponse: Codable {
    let success: Bool
    let errors: [String]?
    let symbols: [String: UInt32]?
}

func loadProgram(...) async throws -> LoadProgramResponse {
    return try await post(url: url, body: request)
}
```

### Issue 3: NSTextView Zero-Width Frame

**Symptoms:**
- Frame shows (0.0, 0.0, 0.0, height)
- Container size has negative width
- Text storage has content but nothing renders

**Solution:**
Configure NSTextView sizing properly for NSScrollView:

```swift
// For wrapping text editor:
textView.minSize = NSSize(width: 0.0, height: scrollView.contentSize.height)
textView.maxSize = NSSize(width: CGFloat.greatestFiniteMagnitude,
                          height: CGFloat.greatestFiniteMagnitude)
textView.isVerticallyResizable = true
textView.isHorizontallyResizable = false
textView.autoresizingMask = [.width]

if let textContainer = textView.textContainer {
    textContainer.containerSize = NSSize(
        width: scrollView.contentSize.width,
        height: CGFloat.greatestFiniteMagnitude
    )
    textContainer.widthTracksTextView = true
}
```

### Issue 4: Optional Field Decoding Failures

**Symptom:**
Backend sometimes omits fields, causing decoding failures.

**Example:**
```json
// Backend returns this on initial state:
{"state": "halted", "pc": 32768, "cycles": 0}

// But Swift model expects:
struct VMStatus: Codable {
    var instruction: String  // ❌ Not always present!
}
```

**Solution:**
Make fields optional when backend doesn't guarantee them:

```swift
struct VMStatus: Codable {
    var state: String
    var pc: UInt32
    var instruction: String?  // ✅ Optional - backend doesn't always return this
    var cycleCount: UInt64?
    var error: String?
}
```

## Case Study: Program Loading and Rendering Bug

### Problem Description

When loading `examples/hello.s` via the Swift GUI, users encountered:
1. Initial error: "Failed to load program: Failed to decode response: The data couldn't be read because it is missing."
2. After fixing type mismatch: Program loaded successfully (HTTP 200) but text didn't appear in editor
3. Editor showed blank despite text storage containing 550 characters

### Investigation Timeline

#### Phase 1: Type Mismatch (2 hours)

**Issue:** Swift client expected `EmptyResponse` but backend returned `LoadProgramResponse`.

**How we found it:**
```bash
# Backend actually returns:
curl http://localhost:8080/api/v1/session/$ID/load -X POST -d '{"source":"..."}'
{"success":true,"symbols":{"_start":32768}}

# But Swift expected: {}
```

**Fix:**
```swift
// Added matching response type
struct LoadProgramResponse: Codable {
    let success: Bool
    let errors: [String]?
    let symbols: [String: UInt32]?
}

func loadProgram(...) async throws -> LoadProgramResponse {
    return try await post(url: url, body: request)
}
```

**Lesson:** Always verify backend response format matches Swift model exactly.

#### Phase 2: Xcode Build Caching (1 hour)

**Issue:** Changes weren't taking effect despite successful builds.

**How we detected it:**
```bash
# Source modified AFTER binary was built
$ stat APIClient.swift     # 2026-01-03 10:34:52
$ stat ARMEmulator.app/... # 2026-01-03 10:28:00  ❌ Older!
```

**Fix:**
```bash
rm -rf ~/Library/Developer/Xcode/DerivedData/ARMEmulator-*
xcodegen generate
# Rebuild in Xcode
```

**Lesson:** Check binary timestamps before debugging. Go nuclear with DerivedData early.

#### Phase 3: NSTextView Rendering Investigation (3 hours)

**Issue:** Text loaded into storage but didn't display visually.

**Diagnostics showed:**
```
✅ sourceCode set to 550 characters
✅ updateNSView called
✅ Text storage length: 550
✅ Container size: (439.5, 1.7976931348623157e+308)
✅ Frame: (0.0, 0.0, 449.5, 247.5)
❌ But text area completely blank!
```

**Attempts that didn't work:**
1. Adding explicit text color/background color
2. Configuring text view sizing (fixed zero-width frame, but still blank)
3. Forcing layout invalidation and redraw
4. Switching between wrapping and non-wrapping text containers

**What finally worked:**
Commenting out the `LineNumberGutterView`:

```swift
// Create and add gutter view
// let gutterView = LineNumberGutterView(scrollView: scrollView, orientation: .verticalRuler)
// gutterView.configure(textView: textView, onBreakpointToggle: onBreakpointToggle)

// Add gutter as a ruler view
// scrollView.verticalRulerView = gutterView
// scrollView.hasVerticalRuler = true
// scrollView.rulersVisible = true
```

**Immediately:** Text appeared!

**Lesson:** When complex views don't render, simplify by removing custom components first.

#### Phase 4: Root Cause Analysis

**Why did LineNumberGutterView break rendering?**

Potential causes:
1. Ruler view incorrectly calculating bounds/drawing area
2. Ruler view interfering with NSClipView layout
3. Ruler view's `ruleThickness` affecting scroll view geometry
4. Custom drawing code covering text view content area

**The fix is temporary** - line numbers are now disabled. The gutter view needs debugging separately to restore functionality.

### Complete Fix Summary

**Changes made:**

1. **APIClient.swift:**
   - Added `LoadProgramResponse` struct matching backend
   - Updated `loadProgram()` to return `LoadProgramResponse`
   - Fixed memory API parameter (`addr` → `address`)

2. **EmulatorViewModel.swift:**
   - Handle `LoadProgramResponse` and check `success` field
   - Only set `sourceCode` if load succeeded

3. **ProgramState.swift:**
   - Made `VMStatus.instruction` optional (backend doesn't always return it)

4. **EditorView.swift:**
   - Added text color and background color settings
   - Configured NSTextView sizing to prevent zero-width frames
   - Commented out LineNumberGutterView (temporary fix)
   - Added layout invalidation and redraw logic

**Total debugging time:** ~6 hours
**Time that could have been saved:** ~4 hours with immediate DerivedData clean and simplification approach

## Debugging Toolkit

### Essential Commands

```bash
# Clean build
rm -rf ~/Library/Developer/Xcode/DerivedData/ARMEmulator-*
cd swift-gui
xcodegen generate
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator clean build | xcbeautify

# Verify binary freshness
find ~/Library/Developer/Xcode/DerivedData -name "ARMEmulator.app" -type d -exec stat -f "%Sm %N" -t "%Y-%m-%d %H:%M:%S" {}/Contents/MacOS/ARMEmulator \;

# Test backend directly
curl -s http://localhost:8080/api/v1/session -X POST -d '{}'
lsof -i :8080

# Run with debugging
open ~/Library/Developer/Xcode/DerivedData/.../ARMEmulator.app

# Check logs (if using os_log)
log stream --predicate 'subsystem == "com.example.ARMEmulator"' --level debug
```

### Diagnostic Code Patterns

**NSTextView rendering diagnostics:**
```swift
func updateNSView(_ scrollView: NSScrollView, context: Context) {
    guard let textView = scrollView.documentView as? NSTextView else { return }

    if textView.string != text {
        textView.string = text

        // Diagnostics
        print("TextView frame: \(textView.frame)")
        print("TextView bounds: \(textView.bounds)")
        print("ScrollView frame: \(scrollView.frame)")

        if let storage = textView.textStorage {
            print("Text storage length: \(storage.length)")
        }

        if let container = textView.textContainer {
            print("Container size: \(container.containerSize)")
            print("Width tracks: \(container.widthTracksTextView)")
        }

        if let layoutManager = textView.layoutManager,
           let textContainer = textView.textContainer {
            print("Number of glyphs: \(layoutManager.numberOfGlyphs)")
            print("Used rect: \(layoutManager.usedRect(for: textContainer))")
        }
    }
}
```

**API response logging:**
```swift
private func performRequest<T: Decodable>(request: URLRequest) async throws -> T {
    let (data, response) = try await session.data(for: request)

    guard let httpResponse = response as? HTTPURLResponse else {
        throw APIError.invalidResponse
    }

    // Temporary diagnostics
    print("API: \(request.httpMethod ?? "?") \(request.url?.path ?? "?")")
    print("Status: \(httpResponse.statusCode)")
    print("Body: \(String(data: data, encoding: .utf8) ?? "<binary>")")
    print("Expected: \(T.self)")

    guard (200...299).contains(httpResponse.statusCode) else {
        let errorMessage = String(data: data, encoding: .utf8) ?? "Unknown error"
        throw APIError.serverError(httpResponse.statusCode, errorMessage)
    }

    do {
        return try JSONDecoder().decode(T.self, from: data)
    } catch {
        print("Decode error: \(error)")
        throw APIError.decodingError(error)
    }
}
```

## Debugging Checklist

When facing a Swift GUI issue, work through this checklist:

### Pre-Debugging
- [ ] Is the Go backend running? (`lsof -i :8080`)
- [ ] Does backend endpoint work? (test with `curl`)
- [ ] Is binary newer than source files? (check timestamps)

### Build Issues
- [ ] Clean DerivedData: `rm -rf ~/Library/Developer/Xcode/DerivedData/ARMEmulator-*`
- [ ] Regenerate project: `xcodegen generate`
- [ ] Kill all running instances before rebuild
- [ ] Verify you're testing the correct build

### View Rendering Issues
- [ ] Add diagnostic logging (frame, bounds, content)
- [ ] Verify data is present (storage, models)
- [ ] Check for zero-width/height frames
- [ ] Comment out custom components (ruler views, overlays)
- [ ] Test with minimal AppKit configuration
- [ ] Add components back incrementally

### API/Data Issues
- [ ] Log actual API responses vs expected types
- [ ] Check for optional fields (backend may omit)
- [ ] Verify JSON structure matches Swift model
- [ ] Test endpoint isolation with `curl`
- [ ] Check WebSocket connection state

### Last Resorts
- [ ] Create minimal reproduction case
- [ ] Test in standalone Swift project (isolate from SwiftUI)
- [ ] Review similar AppKit examples
- [ ] Check for known NSViewRepresentable issues

## Best Practices

### 1. Start Simple

When creating new SwiftUI/AppKit integrations:
```swift
// Step 1: Minimal working version
func makeNSView(context: Context) -> NSScrollView {
    let scrollView = NSScrollView()
    let textView = NSTextView()
    textView.string = "test"
    scrollView.documentView = textView
    return scrollView
}

// Step 2: Verify it works

// Step 3: Add configuration
textView.font = NSFont.monospacedSystemFont(ofSize: 13, weight: .regular)
// Test again

// Step 4: Add custom components
let gutterView = LineNumberGutterView(...)
// Test again
```

### 2. Log Liberally During Development

Add logging, test, then remove it:
```swift
print("🔵 updateNSView called - current: \(textView.string.count)")
// Remove after confirming it works
```

### 3. Verify Backend Contracts

When adding new API endpoints, document and verify:
```swift
/// Load a program into the emulator
///
/// Backend returns:
/// ```json
/// {
///   "success": true,
///   "symbols": {"_start": 32768},
///   "errors": ["optional error messages"]
/// }
/// ```
func loadProgram(...) async throws -> LoadProgramResponse
```

### 4. Make Fields Optional Appropriately

If backend might omit a field, make it optional:
```swift
struct VMStatus: Codable {
    var state: String              // Always present
    var pc: UInt32                 // Always present
    var instruction: String?       // Optional - not always returned
    var cycleCount: UInt64?        // Optional
}
```

### 5. Test Incrementally

Don't implement multiple features before testing:
```swift
// ❌ Bad: Implement everything then test
func makeNSView(context: Context) -> NSScrollView {
    // 100 lines of configuration
    // Custom ruler view
    // Custom drawing
    // Complex gestures
    return scrollView
}

// ✅ Good: Add and test incrementally
func makeNSView(context: Context) -> NSScrollView {
    let scrollView = NSScrollView()
    let textView = NSTextView()
    scrollView.documentView = textView
    // Test - verify text appears

    // Then add ruler view
    // Test - verify text still appears

    // Then add gestures
    // Test - verify everything works

    return scrollView
}
```

## Conclusion

The key to efficient Swift GUI debugging:

1. **Simplify first** - Remove complexity to isolate issues
2. **Verify builds** - Ensure you're testing fresh code
3. **Go nuclear early** - DerivedData issues waste hours
4. **Test backend independently** - Don't assume it's the frontend
5. **Add incrementally** - Test after each addition

Time investment in systematic debugging pays dividends. The program loading bug took 6 hours but could have been resolved in 2 hours with these practices.

## Future Work

To restore full functionality:

1. **Debug LineNumberGutterView separately:**
   - Create minimal test case with NSTextView + NSRulerView
   - Verify `ruleThickness` and bounds calculations
   - Test `drawHashMarksAndLabels(in:)` implementation
   - Check for interference with NSClipView

2. **Add unit tests for NSViewRepresentable components:**
   - Test text rendering
   - Test ruler view integration
   - Test breakpoint toggle functionality

3. **Document known AppKit/SwiftUI integration gotchas:**
   - NSRulerView requirements
   - NSTextView sizing in NSScrollView
   - Layout manager lifecycle
