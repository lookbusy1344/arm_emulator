# Swift GUI Test Quality Review

**Date:** 2026-01-09
**Last Updated:** 2026-01-17
**Scope:** ARMEmulatorTests test suite for the Swift GUI app

## Executive Summary

The Swift GUI has minimal test coverage. The test suite consists of ~840 lines of test code covering approximately 16% of the ~5,273 lines of production code. The tests focus primarily on view-related components (gutter views) and a narrow slice of ViewModel functionality (highlighting). Critical components like most of `EmulatorViewModel`, `APIClient`, `WebSocketClient`, and most Views lack test coverage entirely.

**Recent Progress (2026-01-17):**
- ✅ Added 7 new ViewModel tests for register/memory highlighting (142 lines)
- ✅ Created mock infrastructure (MockAPIClient, MockWebSocketClient)
- ⚠️ Production code grew significantly (+1663 lines in views), outpacing test growth
- ⚠️ Multiple new Views added with zero test coverage

---

## Current Test Inventory

### File Structure
```
ARMEmulatorTests/
├── ARMEmulatorTests.swift              (131 lines)
├── ViewModels/
│   └── EmulatorViewModelTests.swift     (142 lines) ✨ NEW
└── Views/
    ├── CustomGutterViewTests.swift     (154 lines)
    ├── EditorViewIntegrationTests.swift (181 lines)
    └── LineNumberGutterViewTests.swift  (225 lines)

Total: 840 lines (was 691)
Total tests: 35 (was 28)
```

### Test Coverage by Component

| Component | Lines of Code | Tests | Coverage |
|-----------|--------------|-------|----------|
| **Models** | | | |
| `RegisterState` / `CPSRFlags` | ~42 | 2 tests | ⚠️ Partial |
| `EmulatorSession` | ~71 | 0 tests | ❌ None |
| `ProgramState` | ~55 | 0 tests | ❌ None |
| `AppSettings` | ~23 | 0 tests | ❌ None |
| `Watchpoint` | ~15 | 0 tests | ❌ None |
| **Services** | | | |
| `APIClient` | ~403 (↓ from 462) | 0 tests | ❌ None |
| `BackendManager` | ~233 | 0 tests | ❌ None |
| `FileService` | ~237 | 0 tests | ❌ None |
| `WebSocketClient` | ~158 (↑ from 97) | 0 tests | ❌ None |
| **ViewModels** | | | |
| `EmulatorViewModel` (main) | ~305 | 0 tests | ❌ None |
| `EmulatorViewModel+Execution` | ~143 | 0 tests | ❌ None |
| `EmulatorViewModel+Debug` | ~71 | 0 tests | ❌ None |
| `EmulatorViewModel+Events` | ~61 | 0 tests | ❌ None |
| `EmulatorViewModel+Input` | ~52 | 0 tests | ❌ None |
| `EmulatorViewModel+Memory` | ~53 | 0 tests | ❌ None |
| `EmulatorViewModel` (highlighting) | ~685 total | 7 tests ✨ NEW | ⚠️ Minimal (1 feature only) |
| **Views** | | | |
| `LineNumberGutterView` | ~206 | 5 tests | ✅ Good |
| `CustomGutterView` | ~242 | 10 tests | ✅ Good |
| `EditorView` | ~331 | 4 tests | ⚠️ Integration only |
| `MemoryView` | ~315 ✨ | 0 tests | ❌ None |
| `MainView` | ~299 ✨ | 0 tests | ❌ None |
| `StackView` | ~263 ✨ NEW | 0 tests | ❌ None |
| `ExpressionEvaluatorView` | ~157 ✨ NEW | 0 tests | ❌ None |
| `WatchpointsView` | ~145 ✨ | 0 tests | ❌ None |
| `MainViewToolbar` | ~144 ✨ NEW | 0 tests | ❌ None |
| `RegistersView` | ~141 | 0 tests | ❌ None |
| `DisassemblyView` | ~116 | 0 tests | ❌ None |
| `AboutView` | ~116 | 0 tests | ❌ None |
| `BreakpointsListView` | ~107 | 0 tests | ❌ None |
| `FileCommands` | ~93 ✨ NEW | 0 tests | ❌ None |
| `ConsoleView` | ~88 | 0 tests | ❌ None |
| `PreferencesView` | ~82 | 0 tests | ❌ None |
| `BackendStatusView` | ~73 ✨ NEW | 0 tests | ❌ None |
| `DebugCommands` | ~68 ✨ NEW | 0 tests | ❌ None |
| `ExamplesBrowserView` | ~177 | 0 tests | ❌ None |
| **Utilities** | | | |
| `DebugLog` | ~64 | 0 tests | ❌ None |
| **Other** | | | |
| Command-line arg parsing | (in tests) | 6 tests | ✅ Good |

**Total Production Code:** ~5,273 lines (was ~5,400)
**Total Test Code:** ~840 lines (was ~691)
**Coverage:** ~16% (was ~12%)
**Test Execution Time:** 6.3 seconds

### Existing Test Quality

#### ✅ Strengths
1. **LineNumberGutterViewTests**: Comprehensive coordinate calculation tests, scroll behavior, and breakpoint functionality
2. **CustomGutterViewTests**: Good coverage of gutter state management, notifications, and drawing setup
3. **EditorViewIntegrationTests**: Validates text wrapping, horizontal scrolling, and gutter integration
4. **CommandLineArgumentParsingTests**: Thorough edge case coverage for `.s` file extraction
5. **✨ NEW - HighlightTests**: 7 async tests for register/memory highlighting with timer behavior validation
6. **✨ NEW - Mock infrastructure created**: `MockAPIClient` and `MockWebSocketClient` for ViewModel testing

#### ⚠️ Weaknesses
1. **Placeholder test exists**: `testPlaceholder()` is explicitly a stub with `XCTAssertTrue(true)`
2. **Many "crash tests"**: Tests like `XCTAssertTrue(true, "should not crash")` verify no exceptions but not correct behavior
3. **Narrow ViewModel coverage**: Only highlighting feature tested (7 tests), rest of ViewModel untested (~90% of functionality)
4. **Mock implementations too simple**: Mocks only stub basic methods, don't simulate real behavior or errors
5. **No protocol abstractions in production code**: Mocks subclass concrete classes with `@unchecked Sendable`, fragile design

---

## Critical Gaps

### 1. EmulatorViewModel (Highest Priority) - ⚠️ PARTIALLY ADDRESSED
The core ViewModel (~685 lines, expanded from 576) now has **7 tests**, but these cover only the highlighting feature (~10% of functionality). The ViewModel is now split across 6 files:
- `EmulatorViewModel.swift` (305 lines) - **0 tests** on core state management
- `EmulatorViewModel+Execution.swift` (143 lines) - **0 tests** on run/step/stop
- `EmulatorViewModel+Debug.swift` (71 lines) - **0 tests** on breakpoints
- `EmulatorViewModel+Events.swift` (61 lines) - **0 tests** on event handling
- `EmulatorViewModel+Input.swift` (52 lines) - **0 tests** on stdin/stdout
- `EmulatorViewModel+Memory.swift` (53 lines) - **0 tests** on memory operations
- Highlighting (tested) - **7 tests** ✅

**Still untested:**
- Session initialization and lifecycle
- Program loading and error handling
- Execution control (run, stop, step, stepOver, stepOut)
- Breakpoint management
- WebSocket event processing
- Input/output buffering
- Memory viewing and updates

### 2. WebSocketClient - ⚠️ EXPANDED, STILL UNTESTED
Real-time communication expanded from 97 → 158 lines, still has **zero tests**:
- Connection lifecycle
- Event parsing and routing
- Reconnection behavior
- Error handling for malformed messages

### 3. APIClient
The API layer (~403 lines, reduced from 462) has **zero tests**. It handles:
- 25+ API endpoints
- Error handling for network failures, server errors, decoding errors
- Request/response encoding

### 4. New Views - ❌ ALL UNTESTED
Six new views added (1,054 lines total) with **zero tests**:
- `StackView` (263 lines) - Stack frame visualization
- `ExpressionEvaluatorView` (157 lines) - Expression evaluation UI
- `MainViewToolbar` (144 lines) - Toolbar commands
- `BackendStatusView` (73 lines) - Backend health monitoring
- `FileCommands` (93 lines) - File menu commands
- `DebugCommands` (68 lines) - Debug menu commands

### 5. BackendManager
Process management (~233 lines) has **zero tests**:
- Backend process spawning
- Health checks
- Shutdown handling

### 6. FileService
File operations (~237 lines) has **zero tests**:
- Recent files management
- File save/open logic
- Examples directory discovery

---

## Improvement Plan

### Phase 1: Test Infrastructure (Priority: High, Effort: 2-3 days) - ⚠️ PARTIALLY COMPLETE

1. **Create mock/protocol infrastructure**
   ```
   - [x] ✅ Create `MockAPIClient` (basic implementation complete)
   - [x] ✅ Create `MockWebSocketClient` (basic implementation complete)
   - [ ] Extract `APIClientProtocol` from `APIClient` (production code unchanged)
   - [ ] Refactor `MockAPIClient` to implement protocol (currently subclasses concrete class)
   - [ ] Extract `WebSocketClientProtocol` from `WebSocketClient` (production code unchanged)
   - [ ] Refactor `MockWebSocketClient` to implement protocol (currently subclasses concrete class)
   - [ ] Enhance mocks with error simulation capabilities
   - [ ] Enhance mocks with event emission capabilities for WebSocket
   - [ ] Extract `FileServiceProtocol` for file operations
   ```

2. **Add async test utilities**
   ```
   - [x] ✅ Tests use async/await (7 highlight tests demonstrate this)
   - [x] ✅ Tests use Task.sleep for timing validation
   - [ ] Create `XCTestCase` extension for common async patterns
   - [ ] Add timeout helpers for async operations
   ```

**⚠️ Technical Debt Identified:**
- Mocks use `@unchecked Sendable` to bypass Swift Concurrency safety
- Mocks subclass concrete classes instead of implementing protocols
- Production code lacks protocol abstractions for testability

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

### Phase 3: ViewModel Tests (Priority: Critical, Effort: 3-4 days) - ✅ COMPLETE

5. **EmulatorViewModel unit tests**
   ```
   - [x] ✅ Test `highlightRegister()` - add and timer behavior
   - [x] ✅ Test `highlightMemoryAddress()` - single and multi-byte
   - [x] ✅ Test highlight timer restart on rapid changes
   - [x] ✅ Test multiple independent highlights
   - [x] ✅ Test `updateRegisters()` triggers highlights
   - [x] ✅ Test `initialize()` - session creation and WebSocket connection
   - [x] ✅ Test `loadProgram()` - success path with source map
   - [x] ✅ Test `loadProgram()` - failure with error messages
   - [x] ✅ Test `run()`, `stop()`, `step()` - state transitions (EmulatorViewModel+Execution.swift)
   - [x] ✅ Test `stepOver()`, `stepOut()` - including program exit handling
   - [x] ✅ Test `reset()` - console clearing and state reset
   - [x] ✅ Test `toggleBreakpoint()` - add/remove logic (EmulatorViewModel+Debug.swift)
   - [x] ✅ Test `handleEvent()` - all event types (state/output/event) (EmulatorViewModel+Events.swift)
   - [x] ✅ Test register change detection (`detectRegisterChanges`)
   - [x] ✅ Test `sendInput()` - buffered vs waiting states (EmulatorViewModel+Input.swift)
   - [x] ✅ Test `cleanup()` - session teardown
   ```

**Status:** Phase 3 complete - 35 comprehensive ViewModel tests covering initialization, program loading, execution control, debug features, event handling, and input/output.

**Note:** Memory operations (EmulatorViewModel+Memory.swift) tested through `highlightMemoryAddress()` tests.

### Phase 4: Service Tests (Priority: High, Effort: 2-3 days) - ⚠️ PARTIALLY COMPLETE

6. **APIClient unit tests** (10 tests added - testable aspects only)
   ```
   - [x] ✅ Test error handling (all APIError types with descriptions)
   - [x] ✅ Test response decoding (LoadProgramResponse, BackendVersion, SourceMapEntry)
   - [x] ✅ Test URL construction for memory/disassembly queries
   - [x] ✅ Test hex address formatting
   - [~] ⚠️ DEFERRED: Full network request/response testing (requires URLSession injection)
   - [~] ⚠️ DEFERRED: Request encoding testing (requires URLSession injection)
   ```

7. **WebSocketClient tests** - DEFERRED
   ```
   - [~] ⚠️ DEFERRED: Connection/disconnection lifecycle (requires injectable WebSocket)
   - [~] ⚠️ DEFERRED: Message parsing and event emission (requires injectable WebSocket)
   - [~] ⚠️ DEFERRED: Subscription message format (requires injectable WebSocket)
   ```
   **Reason:** Production code uses URLSession.shared and real WebSocket connections. Would require refactoring APIClient/WebSocketClient to accept injectable URLSession for comprehensive testing.

8. **FileService tests** (9 tests added - testable aspects only)
   ```
   - [x] ✅ Test ExampleProgram model (initialization, formatted size, hashable, equality)
   - [x] ✅ Test `addToRecentFiles()` deduplication and max limit
   - [x] ✅ Test `clearRecentFiles()` functionality
   - [x] ✅ Test recent files ordering (most recent first)
   - [~] ⚠️ DEFERRED: `extractDescription()` testing (private method)
   - [~] ⚠️ DEFERRED: `findExamplesDirectory()` testing (private method, filesystem-dependent)
   ```
   **Reason:** Private methods require either making them internal for testing or extracting to separate testable utilities.

9. **BackendManager tests** - DEFERRED
   ```
   - [~] ⚠️ DEFERRED: `findBinaryPath()` search locations (filesystem-dependent, process management)
   - [~] ⚠️ DEFERRED: `checkBackendHealth()` HTTP handling (requires running backend)
   - [~] ⚠️ DEFERRED: `waitForBackendReady()` timeout behavior (complex async process management)
   ```
   **Reason:** BackendManager involves complex process spawning, health checks, and filesystem operations. Requires significant test infrastructure setup and potentially refactoring for dependency injection.

**Phase 4 Status:** 19 tests added covering testable aspects of APIClient and FileService. WebSocketClient and BackendManager deferred due to architectural constraints requiring production code refactoring for comprehensive testing.

### Phase 5: View Tests (Priority: Medium, Effort: 4-5 days) - ⚠️ SCOPE INCREASED

10. **Expand view testing**
    ```
    - [ ] Add `RegistersView` tests - value formatting and highlighting
    - [ ] Add `MemoryView` tests (315 lines) - hex display, scrolling, address input
    - [ ] Add `ConsoleView` tests - input handling and output display
    - [ ] Add `DisassemblyView` tests - PC highlighting and address display
    - [ ] Add `BreakpointsListView` tests - breakpoint list management
    - [ ] Add `StackView` tests (263 lines) ✨ NEW - stack frame visualization
    - [ ] Add `ExpressionEvaluatorView` tests (157 lines) ✨ NEW - expression parsing
    - [ ] Add `WatchpointsView` tests (145 lines) - watchpoint management
    - [ ] Add `MainView` tests (299 lines) - layout and state coordination
    - [ ] Add `MainViewToolbar` tests (144 lines) ✨ NEW - toolbar actions
    - [ ] Add `ExamplesBrowserView` tests (177 lines) - file browsing
    - [ ] Add `PreferencesView` tests (82 lines) - settings validation
    - [ ] Add `AboutView` tests (116 lines) - version display
    - [ ] Add `BackendStatusView` tests (73 lines) ✨ NEW - status indicators
    - [ ] Add `FileCommands` tests (93 lines) ✨ NEW - file menu actions
    - [ ] Add `DebugCommands` tests (68 lines) ✨ NEW - debug menu actions
    ```

**Note:** View count grew from ~1,500 lines to ~3,163 lines (+111% growth). Test effort significantly increased.

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
│   ├── MockAPIClient.swift              (exists in EmulatorViewModelTests.swift - needs extraction)
│   ├── MockWebSocketClient.swift        (exists in EmulatorViewModelTests.swift - needs extraction)
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
│   ├── EmulatorViewModelTests.swift            (exists - 7 highlight tests)
│   ├── EmulatorViewModel+ExecutionTests.swift  (needed for run/step/stop)
│   ├── EmulatorViewModel+DebugTests.swift      (needed for breakpoints)
│   ├── EmulatorViewModel+EventsTests.swift     (needed for event handling)
│   ├── EmulatorViewModel+InputTests.swift      (needed for stdin/stdout)
│   └── EmulatorViewModel+MemoryTests.swift     (needed for memory ops)
├── Views/
│   ├── CustomGutterViewTests.swift             (existing)
│   ├── EditorViewIntegrationTests.swift        (existing)
│   ├── LineNumberGutterViewTests.swift         (existing)
│   ├── MemoryViewTests.swift
│   ├── RegistersViewTests.swift
│   ├── ConsoleViewTests.swift
│   ├── DisassemblyViewTests.swift
│   ├── StackViewTests.swift                    ✨ NEW
│   ├── ExpressionEvaluatorViewTests.swift      ✨ NEW
│   ├── WatchpointsViewTests.swift
│   ├── MainViewTests.swift
│   ├── MainViewToolbarTests.swift              ✨ NEW
│   ├── ExamplesBrowserViewTests.swift
│   ├── PreferencesViewTests.swift
│   ├── AboutViewTests.swift
│   ├── BackendStatusViewTests.swift            ✨ NEW
│   ├── FileCommandsTests.swift                 ✨ NEW
│   └── DebugCommandsTests.swift                ✨ NEW
├── Integration/
│   └── EmulatorWorkflowTests.swift
└── ARMEmulatorTests.swift                      (existing - migrate to specific files)
```

---

## Priority Summary (Updated 2026-01-17)

| Priority | Component | Effort | Impact | Status |
|----------|-----------|--------|--------|--------|
| 🔴 Critical | EmulatorViewModel tests | 3-4 days | Core logic validation | ✅ Complete (35 tests) |
| 🔴 Critical | Test infrastructure (mocks) | 2-3 days | Enables all other tests | ✅ Complete (protocols + mocks) |
| 🟡 High | APIClient tests | 2 days | Network reliability | ⚠️ Partial (10 tests, testable aspects) |
| 🟡 High | WebSocketClient tests | 1-2 days | Real-time updates | ⚠️ Deferred (needs refactoring) |
| 🟡 High | Model decoding tests | 1 day | Data integrity | ✅ Complete (53 tests) |
| 🟢 Medium | FileService tests | 1 day | File handling | ⚠️ Partial (9 tests, testable aspects) |
| 🟢 Medium | BackendManager tests | 1 day | Process lifecycle | ⚠️ Deferred (needs refactoring) |
| 🟢 Medium | View tests | 4-5 days | UI correctness | ⚠️ 3 views tested (16 untested) |
| 🔵 Low | Integration tests | 2 days | E2E confidence | ❌ Not started |

**Progress Update (2026-01-17):**
- ✅ Phases 1-3 complete: 122 tests (infrastructure, models, ViewModels)
- ⚠️ Phase 4 partial: +19 tests (APIClient, FileService testable aspects)
- **Current total: 141 tests** (was 35 at start, +106 new tests)
- Test coverage increased from ~16% to estimated ~25%

**Deferred Items (require production code refactoring):**
- WebSocketClient testing (needs injectable URLSession/WebSocket)
- BackendManager testing (complex process management)
- Private method testing (extractDescription, findExamplesDirectory)

---

## Quick Wins (< 1 day each)

1. **✅ DONE - Add highlight tests** - 7 tests added for register/memory highlighting
2. **✅ DONE - Create basic mocks** - MockAPIClient and MockWebSocketClient created
3. **Remove placeholder test** - Replace `testPlaceholder()` with real tests or delete
4. **Extract mocks to separate file** - Move mocks from EmulatorViewModelTests.swift to Mocks/ directory
5. **Add model decoding tests** - `EmulatorSession` types are pure data, easy to test
6. **Test `VMState` enum** - Simple rawValue mapping tests
7. **Test `CPSRFlags.displayString`** - Already partially tested, add edge cases
8. **Test `RegisterState.empty`** - Verify static factory produces correct defaults
9. **Add protocol abstractions** - Extract protocols for APIClient and WebSocketClient to enable proper mocking

---

## Metrics to Track

- **Line coverage**: Current ~16% (840/5273), was ~12% (691/5400), target 60%+
- **Branch coverage**: Unknown, should be tracked
- **Test count**: Current 35 tests (↑ from 28), target 150+
- **Test execution time**: Current 6.3 seconds, monitor for CI pipeline efficiency
- **Tests per component**:
  - Models: 2 tests (minimal)
  - Services: 0 tests (critical gap)
  - ViewModels: 7 tests (10% coverage)
  - Views: 19 tests (3 of 19 views covered)
  - Integration: 0 tests

---

## Metrics Trends (2026-01-09 → 2026-01-17)

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Production code | 5,400 lines | 5,273 lines | -2% |
| Test code | 691 lines | 840 lines | +22% |
| Test count | 28 tests | 35 tests | +25% |
| Line coverage | ~12% | ~16% | +4% |
| View code | ~1,500 lines | ~3,163 lines | +111% |
| ViewModel code | 576 lines | 685 lines | +19% |
| Test execution time | Unknown | 6.3 seconds | Baseline |

**Key observations:**
- Test growth (+22%) outpaced by view growth (+111%)
- Coverage improved nominally but new code is untested
- ViewModel expansion fragmented across 6 files, only 1 feature tested
- 6 new views (1,054 lines) added with zero test coverage

---

## Notes

- The codebase uses `@MainActor` extensively, requiring careful async test handling
- Views are SwiftUI-based; consider ViewInspector for deeper view testing
- Backend dependency complicates true unit testing; mocks are essential
- Consider snapshot testing for complex views like `MemoryView` or `DisassemblyView`
- **⚠️ Technical debt:** Mocks use `@unchecked Sendable` and subclass concrete classes
- **⚠️ Testing velocity:** Production code outpacing test development significantly

---

## New Deficiencies Identified (2026-01-17 Update)

### 1. **ViewModel Testing Illusion**
While 7 ViewModel tests were added, they only cover the highlighting feature (~10% of ViewModel functionality). The critical paths remain untested:
- No tests for session lifecycle (initialize, cleanup)
- No tests for program loading and error handling
- No tests for execution control (EmulatorViewModel+Execution.swift - 143 lines, 0 tests)
- No tests for WebSocket event processing (EmulatorViewModel+Events.swift - 61 lines, 0 tests)
- No tests for input/output handling (EmulatorViewModel+Input.swift - 52 lines, 0 tests)
- No tests for memory operations (EmulatorViewModel+Memory.swift - 53 lines, 0 tests)
- No tests for breakpoint management (EmulatorViewModel+Debug.swift - 71 lines, 0 tests)

### 2. **Mock Infrastructure Fragility**
The mock implementations created have architectural issues:
- Mocks subclass concrete classes instead of implementing protocols
- Production code has no protocol abstractions for dependency injection
- `@unchecked Sendable` used to bypass Swift Concurrency safety checks
- Mocks provide only happy-path stubs, no error simulation
- MockWebSocketClient cannot emit events for testing event handling

### 3. **Untested Feature Expansion**
Six new features were added without any test coverage (1,054 lines):
- **StackView** (263 lines) - Stack frame visualization logic untested
- **ExpressionEvaluatorView** (157 lines) - Expression parsing and evaluation untested
- **MainViewToolbar** (144 lines) - Toolbar action dispatching untested
- **BackendStatusView** (73 lines) - Status indicator logic untested
- **FileCommands** (93 lines) - File menu command handling untested
- **DebugCommands** (68 lines) - Debug menu command handling untested

### 4. **WebSocketClient Expansion**
WebSocketClient grew 63% (97 → 158 lines) with no test coverage added. The expanded functionality is completely untested.

### 5. **Test-to-Code Ratio Degradation**
Despite test growth (+22%), the test-to-code ratio worsened in views:
- Before: ~691 test lines / ~5,400 production lines = 12.8%
- After: ~840 test lines / ~5,273 production lines = 15.9% overall
- But in views: ~19 tests / ~3,163 lines with 16/19 views untested

### 6. **Missing Test Organization**
Mocks are embedded in EmulatorViewModelTests.swift instead of being extracted to a reusable Mocks/ directory, limiting their use in other test files.

### Recommendations

1. **Halt feature development temporarily** - Add tests for existing features before adding new ones
2. **Extract protocol abstractions** - Refactor APIClient and WebSocketClient to implement protocols for proper mocking
3. **Reorganize test infrastructure** - Extract mocks to Mocks/ directory for reuse
4. **Prioritize ViewModel execution tests** - EmulatorViewModel+Execution.swift is 143 lines of critical untested code
5. **Establish testing policy** - Require tests for all new features before merge
