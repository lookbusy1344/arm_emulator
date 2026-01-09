# Swift GUI Test Quality Review

**Date:** 2026-01-09  
**Scope:** ARMEmulatorTests test suite for the Swift GUI app

## Executive Summary

The Swift GUI has minimal test coverage. The test suite consists of ~691 lines of test code covering approximately 12% of the ~5,400 lines of production code. The tests focus primarily on view-related components (gutter views) with placeholder tests for core functionality. Critical components like `EmulatorViewModel`, `APIClient`, `WebSocketClient`, and most Views lack test coverage entirely.

---

## Current Test Inventory

### File Structure
```
ARMEmulatorTests/
├── ARMEmulatorTests.swift              (131 lines)
└── Views/
    ├── CustomGutterViewTests.swift     (154 lines)
    ├── EditorViewIntegrationTests.swift (181 lines)
    └── LineNumberGutterViewTests.swift  (225 lines)
```

### Test Coverage by Component

| Component | Lines of Code | Tests | Coverage |
|-----------|--------------|-------|----------|
| **Models** | | | |
| `RegisterState` / `CPSRFlags` | ~42 | 2 tests | ⚠️ Partial |
| `EmulatorSession` | ~72 | 0 tests | ❌ None |
| `ProgramState` | ~41 | 0 tests | ❌ None |
| `AppSettings` | ~24 | 0 tests | ❌ None |
| `Watchpoint` | ~16 | 0 tests | ❌ None |
| **Services** | | | |
| `APIClient` | ~462 | 0 tests | ❌ None |
| `BackendManager` | ~234 | 0 tests | ❌ None |
| `FileService` | ~238 | 0 tests | ❌ None |
| `WebSocketClient` | ~97 | 0 tests | ❌ None |
| **ViewModels** | | | |
| `EmulatorViewModel` | ~576 | 0 tests | ❌ None |
| **Views** | | | |
| `LineNumberGutterView` | ~200 (est.) | 6 tests | ✅ Good |
| `CustomGutterView` | ~150 (est.) | 9 tests | ✅ Good |
| `EditorView` | ~300 (est.) | 5 tests | ⚠️ Integration only |
| Other 15 Views | ~1500 (est.) | 0 tests | ❌ None |
| **Utilities** | | | |
| `DebugLog` | ~64 | 0 tests | ❌ None |
| **Other** | | | |
| Command-line arg parsing | (in tests) | 6 tests | ✅ Good |

### Existing Test Quality

#### ✅ Strengths
1. **LineNumberGutterViewTests**: Comprehensive coordinate calculation tests, scroll behavior, and breakpoint functionality
2. **CustomGutterViewTests**: Good coverage of gutter state management, notifications, and drawing setup
3. **EditorViewIntegrationTests**: Validates text wrapping, horizontal scrolling, and gutter integration
4. **CommandLineArgumentParsingTests**: Thorough edge case coverage for `.s` file extraction

#### ⚠️ Weaknesses
1. **Placeholder test exists**: `testPlaceholder()` is explicitly a stub with `XCTAssertTrue(true)`
2. **Many "crash tests"**: Tests like `XCTAssertTrue(true, "should not crash")` verify no exceptions but not correct behavior
3. **No mocking infrastructure**: No protocol abstractions for dependencies like `APIClient`, `URLSession`
4. **No async test coverage**: Despite heavy use of async/await in production code

---

## Critical Gaps

### 1. EmulatorViewModel (Highest Priority)
The core ViewModel (~576 lines) has **zero tests**. This is the central component that:
- Manages all emulator state (registers, breakpoints, watchpoints, memory)
- Coordinates between APIClient and WebSocket events
- Handles complex state transitions (idle → running → paused → halted)
- Processes WebSocket events and execution events

### 2. APIClient
The API layer (~462 lines) has **zero tests**. It handles:
- 25+ API endpoints
- Error handling for network failures, server errors, decoding errors
- Request/response encoding

### 3. WebSocketClient
Real-time communication (~97 lines) has **zero tests**:
- Connection lifecycle
- Event parsing and routing
- Reconnection behavior

### 4. BackendManager
Process management (~234 lines) has **zero tests**:
- Backend process spawning
- Health checks
- Shutdown handling

### 5. FileService
File operations (~238 lines) has **zero tests**:
- Recent files management
- File save/open logic
- Examples directory discovery

---

## Improvement Plan

### Phase 1: Test Infrastructure (Priority: High, Effort: 2-3 days)

1. **Create mock/protocol infrastructure**
   ```
   - [ ] Extract `APIClientProtocol` from `APIClient`
   - [ ] Create `MockAPIClient` implementing the protocol
   - [ ] Extract `WebSocketClientProtocol` from `WebSocketClient`
   - [ ] Create `MockWebSocketClient` for event simulation
   - [ ] Extract `FileServiceProtocol` for file operations
   ```

2. **Add async test utilities**
   ```
   - [ ] Create `XCTestCase` extension for async test helpers
   - [ ] Add timeout helpers for async operations
   ```

### Phase 2: Model Tests (Priority: High, Effort: 1 day)

3. **EmulatorSession models**
   ```
   - [ ] Test `EventData` decoding for all variants (state/output/event)
   - [ ] Test `EmulatorEvent` parsing edge cases
   - [ ] Test `StateUpdate` partial updates
   ```

4. **ProgramState models**
   ```
   - [ ] Test `VMState` rawValue mapping
   - [ ] Test `VMStatus.vmState` computed property
   - [ ] Test `MemoryData` and `DisassemblyInstruction` decoding
   ```

### Phase 3: ViewModel Tests (Priority: Critical, Effort: 3-4 days)

5. **EmulatorViewModel unit tests**
   ```
   - [ ] Test `initialize()` - session creation and WebSocket connection
   - [ ] Test `loadProgram()` - success path with source map
   - [ ] Test `loadProgram()` - failure with error messages
   - [ ] Test `run()`, `stop()`, `step()` - state transitions
   - [ ] Test `stepOver()`, `stepOut()` - including program exit handling
   - [ ] Test `reset()` - console clearing and state reset
   - [ ] Test `toggleBreakpoint()` - add/remove logic
   - [ ] Test `handleEvent()` - all event types (state/output/event)
   - [ ] Test register change detection (`detectRegisterChanges`)
   - [ ] Test `sendInput()` - buffered vs waiting states
   - [ ] Test `cleanup()` - session teardown
   ```

### Phase 4: Service Tests (Priority: High, Effort: 2-3 days)

6. **APIClient unit tests**
   ```
   - [ ] Test request encoding for each endpoint
   - [ ] Test response decoding for each endpoint
   - [ ] Test error handling (network, server, decoding errors)
   - [ ] Test URL construction for memory/disassembly queries
   ```

7. **WebSocketClient tests**
   ```
   - [ ] Test connection/disconnection lifecycle
   - [ ] Test message parsing and event emission
   - [ ] Test subscription message format
   ```

8. **FileService tests**
   ```
   - [ ] Test `addToRecentFiles()` deduplication and limit
   - [ ] Test `extractDescription()` from source comments
   - [ ] Test examples directory discovery logic
   ```

9. **BackendManager tests**
   ```
   - [ ] Test `findBinaryPath()` search locations
   - [ ] Test `checkBackendHealth()` HTTP handling
   - [ ] Test `waitForBackendReady()` timeout behavior
   ```

### Phase 5: View Tests (Priority: Medium, Effort: 2-3 days)

10. **Expand view testing**
    ```
    - [ ] Add `RegistersView` tests - value formatting
    - [ ] Add `MemoryView` tests - hex display logic
    - [ ] Add `ConsoleView` tests - input handling
    - [ ] Add `DisassemblyView` tests - PC highlighting
    - [ ] Add `BreakpointsListView` tests
    ```

### Phase 6: Integration Tests (Priority: Medium, Effort: 2-3 days)

11. **End-to-end workflow tests**
    ```
    - [ ] Test full "load → run → step → halt" cycle
    - [ ] Test breakpoint hit and resume
    - [ ] Test stdin/stdout workflow
    - [ ] Test error recovery scenarios
    ```

---

## Recommended Test File Structure

```
ARMEmulatorTests/
├── Mocks/
│   ├── MockAPIClient.swift
│   ├── MockWebSocketClient.swift
│   └── MockFileService.swift
├── Models/
│   ├── EmulatorSessionTests.swift
│   ├── ProgramStateTests.swift
│   └── WatchpointTests.swift
├── Services/
│   ├── APIClientTests.swift
│   ├── BackendManagerTests.swift
│   ├── FileServiceTests.swift
│   └── WebSocketClientTests.swift
├── ViewModels/
│   └── EmulatorViewModelTests.swift
├── Views/
│   ├── CustomGutterViewTests.swift      (existing)
│   ├── EditorViewIntegrationTests.swift (existing)
│   ├── LineNumberGutterViewTests.swift  (existing)
│   ├── MemoryViewTests.swift
│   ├── RegistersViewTests.swift
│   └── ConsoleViewTests.swift
├── Integration/
│   └── EmulatorWorkflowTests.swift
└── ARMEmulatorTests.swift               (existing - migrate to specific files)
```

---

## Priority Summary

| Priority | Component | Effort | Impact |
|----------|-----------|--------|--------|
| 🔴 Critical | EmulatorViewModel tests | 3-4 days | Core logic validation |
| 🔴 Critical | Test infrastructure (mocks) | 2-3 days | Enables all other tests |
| 🟡 High | APIClient tests | 2 days | Network reliability |
| 🟡 High | Model decoding tests | 1 day | Data integrity |
| 🟢 Medium | FileService tests | 1 day | File handling |
| 🟢 Medium | View tests | 2 days | UI correctness |
| 🔵 Low | Integration tests | 2 days | E2E confidence |

**Total estimated effort:** 13-18 days

---

## Quick Wins (< 1 day each)

1. **Remove placeholder test** - Replace `testPlaceholder()` with real tests or delete
2. **Add model decoding tests** - `EmulatorSession` types are pure data, easy to test
3. **Test `VMState` enum** - Simple rawValue mapping tests
4. **Test `CPSRFlags.displayString`** - Already partially tested, add edge cases
5. **Test `RegisterState.empty`** - Verify static factory produces correct defaults

---

## Metrics to Track

- **Line coverage**: Current ~12%, target 60%+
- **Branch coverage**: Unknown, should be tracked
- **Test count**: Current 28 tests, target 100+
- **Test execution time**: Monitor for CI pipeline efficiency

---

## Notes

- The codebase uses `@MainActor` extensively, requiring careful async test handling
- Views are SwiftUI-based; consider ViewInspector for deeper view testing
- Backend dependency complicates true unit testing; mocks are essential
- Consider snapshot testing for complex views like `MemoryView` or `DisassemblyView`
