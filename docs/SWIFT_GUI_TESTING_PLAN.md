# Swift GUI Testing Plan: UI and Integration Testing

## Executive Summary

This document outlines a comprehensive, staged plan to implement high-quality UI and integration testing for the ARM Emulator Swift macOS app. The plan builds on existing unit test infrastructure (13 test files covering ViewModels, Services, Models, and basic Views) to add UI-level testing and complete end-to-end integration testing.

## Current Testing State

### Strengths
- **Strong unit test foundation** (13 test files)
  - ViewModels: `EmulatorViewModelTests.swift` (653 lines, comprehensive coverage)
  - Services: `APIClientTests.swift`, `FileServiceTests.swift`
  - Models: `ProgramStateTests.swift`, `WatchpointTests.swift`, `EmulatorSessionTests.swift`
  - Views: `RegistersViewTests.swift`, `MemoryViewTests.swift`, `ConsoleViewTests.swift`
  - Mock infrastructure: `MockAPIClient.swift`, `MockWebSocketClient.swift`

- **Protocol-based architecture** enables testability
  - `APIClientProtocol`, `WebSocketClientProtocol`
  - Dependency injection through initializers
  - Clean separation of concerns (MVVM)

- **Integration test example**: `EditorViewIntegrationTests.swift` demonstrates NSTextView integration testing

### Gaps
- **No UI testing framework** (XCTest UI Testing not configured)
- **No end-to-end integration tests** (complete user workflows)
- **Limited view testing** (only RegistersView, MemoryView, ConsoleView have dedicated tests)
- **No accessibility testing** (critical for macOS apps)
- **No performance/stress testing** (important for real-time emulator UI)
- **No automated visual regression testing**
- **Backend integration tests require manual setup** (backend must be running)

## Testing Strategy

### Test Pyramid

```
                    ▲
                   / \
                  /   \
                 /  E2E \ ←─── Phase 4 (Manual + Automated UI tests)
                /_______\
               /         \
              / UI Tests  \ ←─ Phase 3 (XCTest UI Testing)
             /___________\
            /             \
           / Integration   \ ←─ Phase 2 (Full workflow tests)
          /_________________\
         /                   \
        /    Unit Tests       \ ←─ Phase 1 (Existing + Gaps)
       /_______________________\
```

### Testing Philosophy

1. **Test at the right level**: Unit tests for logic, integration tests for workflows, UI tests for user interactions
2. **Fast feedback**: Unit tests should be instant, integration tests <1s, UI tests <5s
3. **Deterministic**: All tests must be repeatable and isolated
4. **Maintainable**: Tests should be easy to understand and update
5. **Valuable**: Focus on high-value scenarios (critical paths, edge cases, error handling)

## Staged Implementation Plan

---

## Phase 1: Complete Unit Test Coverage (1-2 weeks)

**Goal**: Achieve 90%+ unit test coverage for all testable components.

### 1.1 Missing View Tests

**Current**: Only 3 views tested (RegistersView, MemoryView, ConsoleView)
**Needed**: 15+ views require unit tests

#### Priority 1 (Critical Views)
- [ ] **EditorView** - Code editor with syntax highlighting, breakpoints, line numbers
  - Test breakpoint toggle UI
  - Test current PC highlighting
  - Test horizontal scrolling configuration
  - Test gutter interaction (builds on existing `EditorViewIntegrationTests`)

- [ ] **MainView** - Application shell and orchestration
  - Test backend status transitions
  - Test startup file loading
  - Test error alert display
  - Test examples browser integration

- [ ] **DisassemblyView** - Assembly instruction display
  - Test instruction formatting
  - Test address highlighting
  - Test symbol resolution display

#### Priority 2 (Feature Views)
- [ ] **BreakpointsListView** - Breakpoint management
  - Test breakpoint list display
  - Test breakpoint enable/disable
  - Test breakpoint removal
  - Test empty state

- [ ] **WatchpointsView** - Memory watchpoint management
  - Test watchpoint creation UI
  - Test watchpoint removal
  - Test watchpoint hit display

- [ ] **ExpressionEvaluatorView** - Debug expression evaluation
  - Test expression input
  - Test result display
  - Test error handling

#### Priority 3 (Supporting Views)
- [ ] **StackView** - Stack visualization
- [ ] **BackendStatusView** - Backend health display
- [ ] **ExamplesBrowserView** - Example program browser
- [ ] **PreferencesView** - Settings UI
- [ ] **AboutView** - Application info
- [ ] **MainViewToolbar** - Toolbar actions

### 1.2 Enhanced ViewModel Testing

**Current**: `EmulatorViewModelTests.swift` has strong coverage but missing edge cases.

- [ ] **Concurrent state changes** - Race conditions in highlight system
- [ ] **Memory pressure scenarios** - Large memory dumps, rapid updates
- [ ] **WebSocket reconnection** - Connection loss and recovery
- [ ] **Backend restart handling** - Session recovery after backend restart
- [ ] **Long-running execution** - Timeout handling, progress tracking

### 1.3 Service Layer Completion

**Current**: `APIClientTests.swift` uses MockURLProtocol but incomplete.

- [ ] **APIClient real network tests** (separate target, requires running backend)
  - Session lifecycle (create, destroy)
  - Program loading (success, errors, large files)
  - Execution control (run, step, stop, reset)
  - Memory operations (read, write, large chunks)
  - Breakpoint/watchpoint management

- [ ] **WebSocketClient tests** (currently no dedicated test file)
  - Connection establishment
  - Message parsing (state, output, events)
  - Reconnection logic
  - Error handling
  - Message ordering/buffering

- [ ] **BackendManager tests** (currently no dedicated test file)
  - Backend process lifecycle
  - Health monitoring
  - Restart logic
  - Version detection
  - Port conflict handling

### 1.4 Model Tests Enhancement

**Current**: Basic model tests exist.

- [ ] **RegisterState edge cases**
  - CPSR flag decoding
  - Special register values (SP, PC, LR)
  - Register change detection

- [ ] **ProgramState transitions**
  - All VMState values
  - Invalid state transitions

- [ ] **AppSettings persistence**
  - UserDefaults integration
  - Default values
  - Migration logic (for future schema changes)

### Phase 1 Success Criteria
- ✅ 90%+ line coverage for ViewModels, Services, Models
- ✅ 70%+ line coverage for Views (where applicable)
- ✅ All critical paths have unit tests
- ✅ All edge cases documented and tested
- ✅ Test execution time <10 seconds total

---

## Phase 2: Integration Testing (2-3 weeks)

**Goal**: Test complete workflows across multiple components without UI interaction.

### 2.1 Session Lifecycle Integration

Test complete emulator session from creation to destruction.

```swift
// Example: SessionLifecycleTests.swift
@MainActor
final class SessionLifecycleTests: XCTestCase {
    func testCompleteSessionLifecycle() async throws {
        // 1. Backend starts
        let backend = BackendManager()
        await backend.startBackend()
        XCTAssertEqual(backend.backendStatus, .running)

        // 2. ViewModel connects
        let viewModel = EmulatorViewModel()
        await viewModel.initialize()
        XCTAssertTrue(viewModel.isConnected)

        // 3. Load program
        await viewModel.loadProgram(source: fixtureProgram)
        XCTAssertNil(viewModel.errorMessage)

        // 4. Execute program
        await viewModel.run()
        try await waitForStatus(.halted, timeout: 5.0)

        // 5. Cleanup
        viewModel.cleanup()
        await backend.stopBackend()
        XCTAssertEqual(backend.backendStatus, .stopped)
    }
}
```

#### Test Scenarios
- [ ] Complete session lifecycle (create → load → run → cleanup)
- [ ] Program with breakpoints (hit → inspect → resume)
- [ ] Interactive program (stdin/stdout handling)
- [ ] Program with errors (syntax error, runtime error)
- [ ] Long-running program (timeout, manual stop)
- [ ] Multiple sessions (isolation, cleanup)

### 2.2 Debugging Workflow Integration

Test debugger features end-to-end.

- [ ] **Breakpoint workflow**
  1. Load program
  2. Set breakpoint at address
  3. Run until breakpoint
  4. Inspect registers
  5. Step over/into
  6. Continue to next breakpoint

- [ ] **Watchpoint workflow**
  1. Set memory watchpoint
  2. Run program
  3. Detect memory write
  4. Pause execution
  5. Inspect memory changes

- [ ] **Expression evaluation workflow**
  1. Pause execution
  2. Evaluate register expressions
  3. Evaluate memory expressions
  4. Display results

### 2.3 File Operations Integration

- [ ] **Load program from file**
  - Small file (<1KB)
  - Large file (>100KB)
  - File with syntax errors
  - File with unicode characters

- [ ] **Recent files management**
  - Add to recent files
  - Limit to N recent files
  - Clear recent files

- [ ] **Example programs**
  - Browse examples
  - Load example
  - Run example to completion

### 2.4 Backend Integration Tests

**IMPORTANT**: These tests require the Go backend to be built and running.

```swift
// Example: BackendIntegrationTests.swift
final class BackendIntegrationTests: XCTestCase {
    static var backendProcess: Process?

    override class func setUp() {
        super.setUp()
        // Start real backend on test port
        backendProcess = startBackend(port: 8081)
    }

    override class func tearDown() {
        backendProcess?.terminate()
        super.tearDown()
    }

    func testRealAPISession() async throws {
        let client = APIClient(baseURL: URL(string: "http://localhost:8081")!)

        // Real HTTP calls
        let session = try await client.createSession()
        XCTAssertFalse(session.sessionId.isEmpty)

        // Real program load
        let response = try await client.loadProgram(
            sessionID: session.sessionId,
            source: "MOV R0, #42\nSWI #0"
        )
        XCTAssertTrue(response.success)

        // Real execution
        try await client.run(sessionID: session.sessionId)

        // Cleanup
        try await client.destroySession(sessionID: session.sessionId)
    }
}
```

#### Scenarios
- [ ] API session management (create, destroy, timeout)
- [ ] Program loading (all example programs)
- [ ] Execution control (run, step, stop, reset)
- [ ] Memory operations (read, write, bounds checking)
- [ ] Breakpoint management (add, remove, hit detection)
- [ ] WebSocket state streaming (register updates, console output)
- [ ] Error responses (invalid session, invalid program, execution errors)

### 2.5 State Synchronization Tests

Test coordination between API calls and WebSocket events.

- [ ] **Register updates arrive via WebSocket after step**
- [ ] **Console output arrives via WebSocket during execution**
- [ ] **Breakpoint hit event arrives before status change**
- [ ] **Memory highlights appear after memory write**
- [ ] **Stale events ignored** (from old session)

### Phase 2 Success Criteria
- ✅ All critical workflows tested end-to-end
- ✅ Backend integration tests pass with real Go backend
- ✅ State synchronization verified
- ✅ Error scenarios handled gracefully
- ✅ Test execution time <30 seconds (excluding backend startup)

---

## Phase 3: UI Testing with XCTest UI Testing (3-4 weeks)

**Goal**: Automate user interactions and verify UI behavior.

### 3.1 XCTest UI Testing Setup

```swift
// New target: ARMEmulatorUITests
final class ARMEmulatorUITests: XCTestCase {
    var app: XCUIApplication!

    override func setUp() {
        super.setUp()
        continueAfterFailure = false

        app = XCUIApplication()
        app.launchArguments = ["--uitesting"]

        // Set environment for testing
        app.launchEnvironment = [
            "BACKEND_PORT": "8082",
            "SKIP_STARTUP_FILE": "1"
        ]

        app.launch()
    }

    func testBasicWorkflow() throws {
        // Wait for backend to start
        let statusLabel = app.staticTexts["Backend: Running"]
        XCTAssertTrue(statusLabel.waitForExistence(timeout: 10))

        // Load example program
        app.buttons["Examples"].click()
        app.outlines.staticTexts["fibonacci.s"].click()
        app.buttons["Load"].click()

        // Verify program loaded
        let editor = app.textViews["CodeEditor"]
        XCTAssertTrue(editor.exists)

        // Run program
        app.buttons["Run"].click()

        // Wait for completion
        let haltedStatus = app.staticTexts["Status: Halted"]
        XCTAssertTrue(haltedStatus.waitForExistence(timeout: 5))

        // Verify console output
        let console = app.textViews["ConsoleOutput"]
        XCTAssertTrue(console.exists)
        XCTAssertTrue(console.value as! String)
            .contains("1 1 2 3 5 8")
    }
}
```

### 3.2 Accessibility Identifiers

**CRITICAL**: Add accessibility identifiers to all interactive elements.

```swift
// Example: In MainView.swift
Button("Run") {
    // ...
}
.accessibilityIdentifier("RunButton")
```

#### Required Identifiers
- [ ] **Toolbar buttons**: Run, Stop, Step, Step Over, Step Out, Reset
- [ ] **Code editor**: CodeEditor
- [ ] **Console view**: ConsoleOutput
- [ ] **Register list**: RegisterList
- [ ] **Memory view**: MemoryView
- [ ] **Breakpoints list**: BreakpointsList
- [ ] **Examples browser**: ExamplesBrowser
- [ ] **Status labels**: BackendStatus, ExecutionStatus
- [ ] **Error alerts**: ErrorAlert

### 3.3 UI Test Scenarios

#### 3.3.1 Program Execution Tests
- [ ] **Run simple program** (fibonacci.s)
  - Load → Run → Verify output → Verify halted state

- [ ] **Step through program**
  - Load → Step → Verify PC increments → Verify registers update

- [ ] **Stop running program**
  - Load → Run → Stop mid-execution → Verify paused state

- [ ] **Reset program**
  - Load → Run → Reset → Verify registers reset → Verify PC reset

#### 3.3.2 Debugging Tests
- [ ] **Set and hit breakpoint**
  - Load → Click gutter to set breakpoint → Run → Verify paused at breakpoint

- [ ] **Remove breakpoint**
  - Set breakpoint → Click gutter again → Verify removed

- [ ] **Step over function**
  - Load → Set breakpoint before BL → Run → Step Over → Verify skipped function

- [ ] **Step out of function**
  - Load → Set breakpoint in function → Run → Step Out → Verify returned

#### 3.3.3 Memory and Register Tests
- [ ] **View memory**
  - Load → Run → Navigate to Memory tab → Verify memory displayed

- [ ] **Memory highlights on write**
  - Load → Step → Verify memory highlights appear for writes

- [ ] **Register highlights on change**
  - Load → Step → Verify changed registers highlighted

- [ ] **Register values update**
  - Load → Step → Verify register values change

#### 3.3.4 Error Handling Tests
- [ ] **Syntax error on load**
  - Enter invalid code → Load → Verify error alert → Verify error message

- [ ] **Runtime error**
  - Load program with divide by zero → Run → Verify error status

- [ ] **Backend connection failure**
  - Kill backend → Verify connection error → Verify retry UI

#### 3.3.5 File Operations Tests
- [ ] **Load from file**
  - File → Open → Select file → Verify loaded

- [ ] **Recent files**
  - Load file → File menu → Verify recent file appears

- [ ] **Examples browser**
  - View → Examples → Select example → Load → Verify code appears

### 3.4 Accessibility Testing

**macOS apps must be accessible.** Use Xcode's Accessibility Inspector.

- [ ] **VoiceOver navigation**
  - Enable VoiceOver → Navigate UI → Verify all elements accessible

- [ ] **Keyboard navigation**
  - Tab through all controls → Verify focus order
  - Verify keyboard shortcuts (⌘R, ⌘., ⌘T)

- [ ] **High contrast mode**
  - Enable high contrast → Verify colors readable

- [ ] **Dynamic type**
  - Increase text size → Verify layout adapts

- [ ] **Reduced motion**
  - Enable reduced motion → Verify animations disabled

### Phase 3 Success Criteria
- ✅ All critical user workflows automated
- ✅ Accessibility testing passes
- ✅ UI tests run on CI (GitHub Actions)
- ✅ All interactive elements have accessibility identifiers
- ✅ Test execution time <2 minutes

---

## Phase 4: End-to-End Testing with MCP (2-3 weeks)

**Goal**: Automate complete user scenarios using MCP (Model Context Protocol) for UI automation.

### 4.1 MCP XcodeBuild Integration

The project already has MCP support documented in `CLAUDE.md`. Use MCP for advanced scenarios.

```bash
# Example: Automated bug reproduction
#!/bin/bash

# Build app
mcp-cli call XcodeBuildMCP/build_macos '{}'

# Launch app with test file
APP_PATH=$(mcp-cli call XcodeBuildMCP/get_mac_app_path '{}' | jq -r '.appPath')
mcp-cli call XcodeBuildMCP/launch_mac_app "{
  \"appPath\": \"$APP_PATH\",
  \"args\": [\"examples/fibonacci.s\"]
}"

# Start log capture
SESSION=$(mcp-cli call XcodeBuildMCP/start_sim_log_cap '{
  "bundleId": "com.lookbusy1344.ARMEmulator",
  "captureConsole": true
}' | jq -r '.sessionId')

# Interact with UI
sleep 2
mcp-cli call XcodeBuildMCP/describe_ui '{}'  # Get UI hierarchy
mcp-cli call XcodeBuildMCP/tap '{"accessibilityId": "RunButton"}'
sleep 5

# Capture screenshot
mcp-cli call XcodeBuildMCP/screenshot '{}'

# Retrieve logs
LOGS=$(mcp-cli call XcodeBuildMCP/stop_sim_log_cap "{\"logSessionId\": \"$SESSION\"}")
echo "$LOGS" | jq -r '.logs' > test-output.log

# Analyze results
grep "Halted" test-output.log && echo "SUCCESS" || echo "FAILURE"
```

### 4.2 Automated Test Scenarios (MCP)

#### 4.2.1 Long-Running Program Tests
- [ ] **Stress test: 1000 iterations**
  - Load loop program → Run → Monitor performance → Verify completion

- [ ] **Memory pressure test**
  - Load program with large allocations → Monitor memory → Verify no leaks

#### 4.2.2 Visual Regression Tests
- [ ] **Screenshot comparison**
  - Capture screenshots at key points → Compare with baseline → Detect regressions

- [ ] **UI layout validation**
  - Verify toolbar layout
  - Verify split view proportions
  - Verify console scrolling

#### 4.2.3 Interactive Program Tests
- [ ] **calculator.s** (interactive)
  - Launch → Type expression → Verify result

- [ ] **bubble_sort.s** (interactive)
  - Launch → Enter array size → Enter elements → Verify sorted output

#### 4.2.4 Multi-Session Tests
- [ ] **Multiple windows**
  - Open 2 windows → Load different programs → Run independently → Verify isolation

### 4.3 Performance Testing

**Use XCTest's `measure` blocks for performance benchmarks.**

```swift
func testRegisterUpdatePerformance() {
    let viewModel = EmulatorViewModel()

    measure {
        for _ in 0..<1000 {
            viewModel.updateRegisters(mockRegisterState)
        }
    }

    // Baseline: <100ms for 1000 updates
}
```

#### Performance Benchmarks
- [ ] **Register updates** (<100ms for 1000 updates)
- [ ] **Memory view scrolling** (60fps for 1MB memory dump)
- [ ] **Console output** (no lag for 10,000 lines)
- [ ] **Breakpoint hit** (<50ms pause time)
- [ ] **Program load** (<1s for 10KB file)

### Phase 4 Success Criteria
- ✅ All 49 example programs tested via automation
- ✅ Performance benchmarks established
- ✅ Visual regression testing in place
- ✅ MCP automation scripts documented
- ✅ CI/CD pipeline includes E2E tests

---

## Phase 5: Continuous Integration & Monitoring (1 week)

**Goal**: Integrate all tests into CI/CD pipeline with quality gates.

### 5.1 GitHub Actions Workflow

```yaml
# .github/workflows/swift-tests.yml
name: Swift GUI Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Go Backend
        run: make build

      - name: Generate Xcode Project
        run: |
          cd swift-gui
          xcodegen generate

      - name: Run Unit Tests
        run: |
          cd swift-gui
          xcodebuild test \
            -project ARMEmulator.xcodeproj \
            -scheme ARMEmulator \
            -destination 'platform=macOS' \
            -enableCodeCoverage YES | xcbeautify

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: swift-gui/coverage.xml

  integration-tests:
    runs-on: macos-latest
    needs: unit-tests
    steps:
      - uses: actions/checkout@v3

      - name: Build Go Backend
        run: make build

      - name: Run Integration Tests
        run: |
          # Start backend
          ./arm-emulator --api-only &
          BACKEND_PID=$!

          # Run tests
          cd swift-gui
          xcodebuild test \
            -project ARMEmulator.xcodeproj \
            -scheme ARMEmulatorIntegration \
            -destination 'platform=macOS' | xcbeautify

          # Cleanup
          kill $BACKEND_PID

  ui-tests:
    runs-on: macos-latest
    needs: integration-tests
    steps:
      - uses: actions/checkout@v3

      - name: Build Go Backend
        run: make build

      - name: Run UI Tests
        run: |
          cd swift-gui
          xcodebuild test \
            -project ARMEmulator.xcodeproj \
            -scheme ARMEmulatorUITests \
            -destination 'platform=macOS' | xcbeautify
```

### 5.2 Quality Gates

**Prevent merging if tests fail or coverage drops.**

- [ ] Unit tests: 100% pass rate, >90% coverage
- [ ] Integration tests: 100% pass rate
- [ ] UI tests: 100% pass rate
- [ ] Performance tests: No regressions >10%
- [ ] SwiftLint: 0 violations
- [ ] SwiftFormat: All files formatted

### 5.3 Test Reporting

- [ ] **Code coverage reports** (Codecov)
- [ ] **Test result dashboard** (GitHub Actions summary)
- [ ] **Performance trend tracking** (store metrics over time)
- [ ] **Flaky test detection** (run tests 3x, flag inconsistent results)

### Phase 5 Success Criteria
- ✅ All tests run on every PR
- ✅ Coverage reports published
- ✅ Quality gates enforced
- ✅ Total CI time <10 minutes

---

## Testing Infrastructure & Tools

### Test Targets

```
ARMEmulator/                    # Main app target
ARMEmulatorTests/               # Unit tests (existing)
ARMEmulatorIntegrationTests/    # NEW: Integration tests
ARMEmulatorUITests/             # NEW: UI tests
ARMEmulatorPerformanceTests/    # NEW: Performance tests
```

### Test Utilities

#### Fixtures
```swift
// ARMEmulatorTests/Fixtures/ProgramFixtures.swift
enum ProgramFixtures {
    static let helloWorld = """
        .text
        .global _start
        _start:
            LDR R0, =message
            SWI #0x02        ; WRITE_STRING
            MOV R0, #0
            SWI #0x00        ; EXIT

        .data
        message: .asciz "Hello, World!\\n"
        """

    static let fibonacci = """
        ; Load from examples/fibonacci.s
        """

    static let syntaxError = """
        INVALID_INSTRUCTION R0, #42
        """
}
```

#### Helpers
```swift
// ARMEmulatorTests/Helpers/AsyncHelpers.swift
extension XCTestCase {
    func waitForStatus(
        _ status: VMState,
        timeout: TimeInterval,
        viewModel: EmulatorViewModel
    ) async throws {
        let deadline = Date().addingTimeInterval(timeout)
        while viewModel.status != status {
            if Date() > deadline {
                throw TestError.timeout("Status did not change to \(status)")
            }
            try await Task.sleep(nanoseconds: 100_000_000)
        }
    }

    func waitForBackend(
        _ manager: BackendManager,
        timeout: TimeInterval
    ) async throws {
        let deadline = Date().addingTimeInterval(timeout)
        while manager.backendStatus != .running {
            if Date() > deadline {
                throw TestError.timeout("Backend did not start")
            }
            try await Task.sleep(nanoseconds: 100_000_000)
        }
    }
}
```

#### Matchers
```swift
// ARMEmulatorTests/Matchers/RegisterMatchers.swift
extension RegisterState {
    func hasRegister(_ name: String, value: UInt32) -> Bool {
        switch name {
        case "R0": return r0 == value
        case "R1": return r1 == value
        // ...
        default: return false
        }
    }
}

// Usage in tests:
XCTAssertTrue(viewModel.registers.hasRegister("R0", value: 42))
```

### Mock Enhancements

#### Enhanced MockAPIClient
```swift
// Add delay simulation for realistic timing
class MockAPIClient: APIClientProtocol {
    var simulateDelay: TimeInterval = 0

    func run(sessionID: String) async throws {
        if simulateDelay > 0 {
            try await Task.sleep(nanoseconds: UInt64(simulateDelay * 1_000_000_000))
        }
        runCalled = true
        if shouldFailRun {
            throw APIError.serverError(500, "Mock error")
        }
    }
}
```

#### Enhanced MockWebSocketClient
```swift
// Add event simulation
class MockWebSocketClient: WebSocketClientProtocol {
    func simulateStateUpdate(_ state: StateUpdate) {
        let event = EmulatorEvent(
            type: "state",
            sessionId: "mock-session-id",
            data: .state(state)
        )
        eventsSubject.send(event)
    }
}
```

---

## Test Data Management

### Example Programs

All 49 example programs should be tested:

```swift
// ARMEmulatorTests/Integration/ExampleProgramsTests.swift
final class ExampleProgramsTests: XCTestCase {
    static let examplePrograms = [
        "hello.s",
        "loops.s",
        "fibonacci.s",
        // ... all 49 programs
    ]

    func testAllExamplePrograms() async throws {
        for program in Self.examplePrograms {
            try await testProgram(named: program)
        }
    }

    func testProgram(named filename: String) async throws {
        let viewModel = EmulatorViewModel()
        await viewModel.initialize()

        let source = try loadExample(filename)
        await viewModel.loadProgram(source: source)

        guard viewModel.errorMessage == nil else {
            XCTFail("Failed to load \(filename): \(viewModel.errorMessage!)")
            return
        }

        await viewModel.run()

        // Wait for halted or error (max 10s)
        try await waitForStatus(.halted, timeout: 10, viewModel: viewModel)

        XCTAssertNil(viewModel.errorMessage,
                     "\(filename) failed: \(viewModel.errorMessage ?? "")")
    }
}
```

### Expected Outputs

Create expected output files for verification:

```
tests/fixtures/expected_outputs/
  fibonacci.txt       # Expected console output for fibonacci.s
  hello.txt           # Expected console output for hello.s
  calculator.txt      # Expected prompts for calculator.s
  ...
```

### Test Snapshots

For visual regression testing, store baseline screenshots:

```
tests/snapshots/
  main-view-idle.png
  main-view-running.png
  main-view-paused.png
  registers-highlighted.png
  memory-view.png
  ...
```

---

## Risk Mitigation

### Flaky Tests

**Problem**: Async operations, WebSocket timing, backend startup can cause flakes.

**Solutions**:
1. Use proper async/await patterns (not sleep)
2. Add configurable timeouts with generous defaults
3. Retry failed tests 3x before marking as failed
4. Isolate tests (each test gets fresh session)
5. Mock network operations where possible

### Slow Tests

**Problem**: UI tests and integration tests can be slow.

**Solutions**:
1. Run unit tests in parallel
2. Cache built backend binary in CI
3. Use XCTest's `addTeardownBlock` for efficient cleanup
4. Optimize test data (small programs where possible)
5. Run expensive tests only on main branch, not PRs

### CI/CD Resource Limits

**Problem**: GitHub Actions has limited macOS runners.

**Solutions**:
1. Run unit tests on every PR (fast)
2. Run integration tests only on main (slower)
3. Run UI tests nightly (slowest)
4. Consider self-hosted runners for better performance

### Test Maintenance

**Problem**: Tests become outdated as code evolves.

**Solutions**:
1. Write clear test names (`testLoadFibonacciProgramAndRunToCompletion`)
2. Add comments explaining "why" not "what"
3. Use fixtures/helpers to reduce duplication
4. Review tests during code review
5. Refactor tests alongside production code

---

## Success Metrics

### Coverage Targets

| Component | Unit | Integration | UI | Total |
|-----------|------|-------------|----|----|
| ViewModels | 95% | 90% | 80% | **90%** |
| Services | 90% | 95% | N/A | **93%** |
| Models | 95% | 80% | N/A | **90%** |
| Views | 60% | 70% | 90% | **75%** |
| **Overall** | **85%** | **85%** | **85%** | **85%+** |

### Quality Metrics

- **Defect escape rate**: <5% (bugs found in production vs testing)
- **Test reliability**: >99% (flaky test rate <1%)
- **Test execution time**: <10 minutes total CI time
- **Test maintenance**: <10% of development time

### Project Milestones

| Phase | Duration | Completion Date | Status |
|-------|----------|-----------------|--------|
| Phase 1: Unit Tests | 1-2 weeks | TBD | 🔴 Not Started |
| Phase 2: Integration | 2-3 weeks | TBD | 🔴 Not Started |
| Phase 3: UI Testing | 3-4 weeks | TBD | 🔴 Not Started |
| Phase 4: E2E/MCP | 2-3 weeks | TBD | 🔴 Not Started |
| Phase 5: CI/CD | 1 week | TBD | 🔴 Not Started |
| **Total** | **9-13 weeks** | **TBD** | **🔴 0% Complete** |

---

## Conclusion

This plan provides a comprehensive roadmap for implementing world-class testing infrastructure for the ARM Emulator Swift GUI. By following the staged approach:

1. **Phase 1** builds on existing unit tests to achieve comprehensive coverage
2. **Phase 2** adds integration testing for complete workflows
3. **Phase 3** implements UI testing with XCTest UI Testing framework
4. **Phase 4** adds advanced E2E testing with MCP automation
5. **Phase 5** ensures continuous quality through CI/CD integration

The result will be a robust, maintainable test suite that provides confidence in every release, catches bugs early, and serves as living documentation for the application.

### Next Steps

1. Review this plan with stakeholders
2. Set target completion dates for each phase
3. Begin Phase 1 implementation (missing view tests)
4. Establish code coverage baseline
5. Configure CI/CD pipeline
6. Track progress using GitHub Projects

---

**Document Version**: 1.0
**Last Updated**: 2026-01-17
**Author**: Development Team
**Status**: Draft - Ready for Review
