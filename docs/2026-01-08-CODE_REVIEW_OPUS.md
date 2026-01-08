# Swift GUI Code Review - 8 January 2026

**Reviewer:** Claude Opus (via Copilot CLI)  
**Scope:** `/swift-gui/ARMEmulator/` - Complete codebase review

---

## Executive Summary

The Swift GUI is a well-structured macOS application using SwiftUI with proper MVVM architecture. The codebase demonstrates good async/await patterns and error handling, but contains several issues ranging from duplicate code to missing reconnection logic. Test coverage is weak with many placeholder assertions.

---

## 1. Features Implemented ✅

### Core Emulation
- Session management (create/destroy via REST API)
- Program loading and assembly execution
- Full register inspection (R0-R12, SP, LR, PC, CPSR flags)
- Real-time state updates via WebSocket
- Step execution (single-step, step-over, step-out)
- Run/stop/reset program controls
- Breakpoint management (add/remove at memory addresses)
- Watchpoint support (read/write/readwrite memory watches)
- Memory viewer with hex display and quick navigation (PC, SP, R0-R3)
- Disassembly viewer with live PC tracking
- Console I/O (stdout/stderr, stdin with blocking detection)
- Expression evaluator for register/memory values

### Editor & UI
- Syntax-aware assembly editor with line numbers
- Gutter-based breakpoint toggling with visual indicators
- Current PC highlighting in editor
- Changed register highlighting (green text)
- Tab-based layout (Registers, Memory, Disassembly, Stack, Console)
- Preferences panel (font size, color scheme, backend URL)
- File operations (open, save, save-as, recent files)
- Example browser for loading sample programs
- Command-line file loading support

### Backend Integration
- Automatic backend process management (spawn, health check, shutdown)
- Debug logging with conditional compilation (DEBUG only)
- Parent process monitoring (Go backend auto-terminates)

---

## 2. Incomplete/TODO Features ⚠️

| Feature | Status | Notes |
|---------|--------|-------|
| WebSocket reconnection | Missing | No auto-reconnect if connection drops |
| Expression evaluator | Basic | Limited functionality, needs enhancement |
| Stack view | Minimal | Present but sparse implementation |
| Examples browser | Fragile | Multiple fallback paths suggest brittle directory discovery |
| Disassembly cache | Basic | Fixed ±32 instruction window, could be smarter |

---

## 3. Bugs & Issues 🐛

### Critical/High Priority

#### 3.1 Duplicate API Methods with Different Signatures

**File:** `Services/APIClient.swift`

Two implementations of `getDisassembly()` exist with conflicting return types:

```swift
// Lines 141-161 - Returns [DisassemblyInstruction]
func getDisassembly(sessionID: String, address: UInt32, count: UInt32) async throws -> [DisassemblyInstruction]

// Lines 361-381 - Returns [DisassembledInstruction]  
func getDisassembly(sessionID: String, address: UInt32, count: Int) async throws -> [DisassembledInstruction]
```

**Issues:**
- Different return types: `DisassemblyInstruction` vs `DisassembledInstruction`
- Different parameter types: `count: UInt32` vs `count: Int`
- Different query param names: `address` vs `addr`
- Swift will compile this (function overloading) but calling code may use wrong variant

Similarly, two `getMemory()` implementations exist:
- Lines 124-139: Returns `MemoryData` struct
- Lines 327-359: Returns `[UInt8]` with base64 decoding

**Impact:** Potential runtime confusion, inconsistent behavior  
**Fix:** Consolidate to single implementation per endpoint

#### 3.2 WebSocket Stops Receiving on Error

**File:** `Services/WebSocketClient.swift` (lines 51-54)

```swift
case let .failure(error):
    print("WebSocket receive error: \(error)")
    // ❌ Does NOT call receiveMessage() again - stops all reception
```

**Impact:** Silent data loss if any receive error occurs  
**Fix:** Add reconnection logic with exponential backoff

#### 3.3 Race Condition in Session Initialization

**File:** `ViewModels/EmulatorViewModel.swift` (lines ~57-73)

The `isInitializing` flag check combined with `!isConnected` could allow concurrent session creation during edge cases.

**Fix:** Use a proper lock or actor isolation

### Medium Priority

#### 3.4 Hardcoded URLs Scattered

URLs are hardcoded in multiple places instead of using centralized config:

```swift
// APIClient.swift:33
init(baseURL: URL = URL(string: "http://localhost:8080")!)

// WebSocketClient.swift:14
guard let url = URL(string: "ws://localhost:8080/api/v1/ws")
```

`AppSettings.backendURL` exists but isn't used consistently.

#### 3.5 No Input Validation for File Paths

**File:** `ARMEmulatorApp.swift` (lines 25-30)

Command-line arguments are taken without validation:
```swift
if CommandLine.arguments.count > 1 {
    startupFilePath = CommandLine.arguments[1]
}
```

No check for file existence, extension, or type.

#### 3.6 Stale Bookmark Silent Failure

**File:** `Services/FileService.swift` (lines 115-123)

Detects stale bookmarks but silently ignores them - user gets no feedback.

#### 3.7 Editor Scroll Callback Fragility

**File:** `Views/EditorView.swift` (lines 75-86)

Closure-based scroll callback with state capture is fragile. Better to use proper `@State` variables or `PreferenceKey`.

#### 3.8 PC Update Hack

**File:** `ViewModels/EmulatorViewModel.swift` (lines 134-136)

```swift
currentPC = 0xFFFF_FFFF // Temporary different value
currentPC = savedPC     // Restore - triggers onChange twice
```

Hacky workaround to force UI refresh. Should use explicit trigger mechanism.

### Low Priority

#### 3.9 Magic Numbers Throughout

- `0x8000` - default memory address
- `15` - kill retry attempts
- `200ms` - sleep duration
- `10s` - connection timeout
- `32` - disassembly instruction window

These should be named constants.

#### 3.10 Inconsistent Logging

Some code uses `DebugLog`, some uses `print()`. Should standardize.

#### 3.11 Manual Register Comparison

16 individual if statements to compare registers. Could use reflection or generate equality.

---

## 4. Type Confusion Issue

Two different disassembly instruction types exist:

```swift
// Models/ProgramState.swift:36
struct DisassemblyInstruction: Codable {
    var address: UInt32
    var machineCode: UInt32
    var disassembly: String      // ← Field name
    var symbol: String?
}

// Services/APIClient.swift:418
struct DisassembledInstruction: Codable, Identifiable, Hashable {
    let address: UInt32
    let machineCode: UInt32
    let mnemonic: String         // ← Different field name
    let symbol: String?
}
```

This is confusing and error-prone. Should be unified.

---

## 5. Test Coverage Analysis

### What Is Tested
- `RegisterState` & `CPSRFlags` initialization
- `LineNumberGutterView` coordinate calculations, scrolling
- `CustomGutterView` initialization, notification handling
- `EditorViewIntegration` horizontal scrolling, text container config

### What Is Untested
- ViewModels (load, run, step, breakpoint workflows)
- API client network calls
- WebSocket message handling
- Error handling paths
- File I/O operations
- Backend process management

### Test Quality Issues

Many tests use placeholder assertions:
```swift
XCTAssertTrue(true)  // ← Not actually testing anything
```

Comments acknowledge inability to verify actual rendering. No mocks for dependencies.

---

## 6. Code Quality Observations

### Strengths ✅
- Excellent `LocalizedError` conformance for errors
- Proper async/await patterns throughout
- `@MainActor` used correctly for thread safety
- Clean MVVM architecture separation
- Good Combine usage for WebSocket events
- Conditional DEBUG logging (zero Release overhead)

### Weaknesses ❌
- Code duplication (duplicate API methods)
- Hardcoded URLs in 3+ files
- Weak test coverage
- Magic numbers throughout
- Inconsistent logging approach
- No structured logging (all strings)

---

## 7. Recommendations

| Priority | Action |
|----------|--------|
| **HIGH** | Consolidate duplicate `getDisassembly()` and `getMemory()` methods |
| **HIGH** | Add WebSocket reconnection with exponential backoff |
| **HIGH** | Unify `DisassemblyInstruction` / `DisassembledInstruction` types |
| **MEDIUM** | Extract hardcoded URLs to centralized config |
| **MEDIUM** | Add input validation for file paths and addresses |
| **MEDIUM** | Improve test coverage for ViewModels and API client |
| **LOW** | Replace magic numbers with named constants |
| **LOW** | Standardize on `DebugLog` for all debug output |

---

## 8. Files Reviewed

```
ARMEmulator/
├── ARMEmulatorApp.swift
├── Models/
│   ├── AppSettings.swift
│   ├── EmulatorSession.swift
│   ├── ProgramState.swift
│   ├── Register.swift
│   └── Watchpoint.swift
├── Services/
│   ├── APIClient.swift
│   ├── BackendManager.swift
│   ├── FileService.swift
│   └── WebSocketClient.swift
├── Utilities/
│   └── DebugLog.swift
├── ViewModels/
│   └── EmulatorViewModel.swift
└── Views/
    ├── AboutView.swift
    ├── BackendStatusView.swift
    ├── BreakpointsListView.swift
    ├── ConsoleView.swift
    ├── CustomGutterView.swift
    ├── DebugCommands.swift
    ├── DisassemblyView.swift
    ├── EditorView.swift
    ├── ExamplesBrowserView.swift
    ├── ExpressionEvaluatorView.swift
    ├── FileCommands.swift
    ├── LineNumberGutterView.swift
    ├── MainView.swift
    ├── MemoryView.swift
    ├── PreferencesView.swift
    ├── RegistersView.swift
    ├── StackView.swift
    └── WatchpointsView.swift

ARMEmulatorTests/
├── ARMEmulatorTests.swift
└── Views/
    ├── CustomGutterViewTests.swift
    ├── EditorViewIntegrationTests.swift
    └── LineNumberGutterViewTests.swift
```

---

*Review completed: 8 January 2026*
