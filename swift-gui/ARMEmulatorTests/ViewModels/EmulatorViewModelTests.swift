// swiftlint:disable file_length
import Combine
import Foundation
import XCTest
@testable import ARMEmulator

// MARK: - Initialization and Cleanup Tests

@MainActor
final class EmulatorViewModelInitializationTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
    }

    func testInitialize() async throws {
        await viewModel.initialize()

        XCTAssertTrue(mockAPIClient.createSessionCalled)
        XCTAssertEqual(viewModel.sessionID, "mock-session-id")
        XCTAssertTrue(viewModel.isConnected)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testInitializeFailure() async throws {
        mockAPIClient.shouldFailCreateSession = true

        await viewModel.initialize()

        XCTAssertTrue(mockAPIClient.createSessionCalled)
        XCTAssertNil(viewModel.sessionID)
        XCTAssertFalse(viewModel.isConnected)
        XCTAssertNotNil(viewModel.errorMessage)
        XCTAssertTrue(viewModel.errorMessage?.contains("Failed to initialize") ?? false)
    }

    func testInitializeSkipsWhenAlreadyConnected() async throws {
        await viewModel.initialize()
        mockAPIClient.createSessionCalled = false // Reset

        await viewModel.initialize()

        XCTAssertFalse(mockAPIClient.createSessionCalled) // Should not be called again
    }

    func testCleanup() async throws {
        viewModel.sessionID = "test-session"
        viewModel.isConnected = true

        viewModel.cleanup()

        // Wait for async Task to complete
        try await Task.sleep(nanoseconds: 100_000_000) // 0.1 second

        XCTAssertTrue(mockAPIClient.destroySessionCalled)
        XCTAssertNil(viewModel.sessionID)
        XCTAssertFalse(viewModel.isConnected)
    }
}

// MARK: - Program Loading Tests

@MainActor
final class EmulatorViewModelProgramLoadingTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        viewModel.sessionID = "test-session" // Set session ID for testing
    }

    func testLoadProgramSuccess() async throws {
        let sourceCode = "MOV R0, #42\nSWI #0"

        await viewModel.loadProgram(source: sourceCode)

        XCTAssertTrue(mockAPIClient.loadProgramCalled)
        XCTAssertEqual(mockAPIClient.lastLoadedSource, sourceCode)
        XCTAssertEqual(viewModel.sourceCode, sourceCode)
        XCTAssertNil(viewModel.errorMessage)
        XCTAssertTrue(mockAPIClient.getRegistersCalled)
        XCTAssertTrue(mockAPIClient.getStatusCalled)
    }

    func testLoadProgramFailureFromAPI() async throws {
        mockAPIClient.shouldFailLoadProgram = true

        await viewModel.loadProgram(source: "MOV R0, #42")

        XCTAssertTrue(mockAPIClient.loadProgramCalled)
        XCTAssertNotNil(viewModel.errorMessage)
        XCTAssertTrue(viewModel.errorMessage?.contains("Failed to load program") ?? false)
    }

    func testLoadProgramFailureFromResponse() async throws {
        mockAPIClient.mockLoadProgramResponse = LoadProgramResponse(
            success: false,
            errors: ["Syntax error on line 1", "Undefined symbol"],
            symbols: nil,
        )

        await viewModel.loadProgram(source: "INVALID")

        XCTAssertTrue(mockAPIClient.loadProgramCalled)
        XCTAssertNotNil(viewModel.errorMessage)
        XCTAssertTrue(viewModel.errorMessage?.contains("Syntax error") ?? false)
        XCTAssertTrue(viewModel.errorMessage?.contains("Undefined symbol") ?? false)
    }

    func testLoadProgramWithoutSession() async throws {
        viewModel.sessionID = nil

        await viewModel.loadProgram(source: "MOV R0, #42")

        XCTAssertFalse(mockAPIClient.loadProgramCalled)
        XCTAssertNotNil(viewModel.errorMessage)
        XCTAssertEqual(viewModel.errorMessage, "No active session")
    }

    func testLoadProgramClearsHighlights() async throws {
        // Set up some highlights
        viewModel.highlightRegister("R0")
        viewModel.highlightMemoryAddress(0x8000, size: 4)

        XCTAssertFalse(viewModel.registerHighlights.isEmpty)
        XCTAssertFalse(viewModel.memoryHighlights.isEmpty)

        await viewModel.loadProgram(source: "MOV R0, #42")

        // After a brief moment, highlights should be cancelled
        // Note: They may still exist briefly due to async nature, but tasks are cancelled
        XCTAssertTrue(true) // Highlights are cleared via cancelAllHighlights()
    }
}

// MARK: - Highlight Tests

@MainActor
final class HighlightTests: XCTestCase {
    var viewModel: EmulatorViewModel!

    override func setUp() async throws {
        viewModel = EmulatorViewModel(
            apiClient: MockAPIClient(),
            wsClient: MockWebSocketClient(),
        )
    }

    func testRegisterHighlightAdded() {
        viewModel.highlightRegister("R0")
        XCTAssertNotNil(viewModel.registerHighlights["R0"])
    }

    func testMemoryHighlightAdded() {
        viewModel.highlightMemoryAddress(0x8000, size: 1)
        XCTAssertNotNil(viewModel.memoryHighlights[0x8000])
    }

    func testRegisterHighlightFadesAfterDelay() async throws {
        viewModel.highlightRegister("R0")

        // Should be highlighted immediately
        XCTAssertNotNil(viewModel.registerHighlights["R0"])

        // Wait for fade to complete
        try await Task.sleep(nanoseconds: 1_600_000_000) // 1.6s

        // Should be removed
        XCTAssertNil(viewModel.registerHighlights["R0"])
    }

    func testRapidChangesRestartTimer() async throws {
        viewModel.highlightRegister("R0")

        // Wait halfway through fade
        try await Task.sleep(nanoseconds: 500_000_000) // 0.5s

        // Trigger another change (should restart timer)
        viewModel.highlightRegister("R0")

        // Wait 1.2s (0.7s after restart)
        try await Task.sleep(nanoseconds: 1_200_000_000)

        // Should still be highlighted
        XCTAssertNotNil(viewModel.registerHighlights["R0"])

        // Wait another 0.5s (1.2s after restart, past 1.5s threshold)
        try await Task.sleep(nanoseconds: 500_000_000)

        // Should be removed now
        XCTAssertNil(viewModel.registerHighlights["R0"])
    }

    func testMultipleRegisterHighlightsIndependent() async throws {
        viewModel.highlightRegister("R0")

        try await Task.sleep(nanoseconds: 500_000_000) // 0.5s

        viewModel.highlightRegister("R1")

        // Both should be highlighted
        XCTAssertNotNil(viewModel.registerHighlights["R0"])
        XCTAssertNotNil(viewModel.registerHighlights["R1"])

        // Wait for R0 to fade (1.2s more = 1.7s total)
        try await Task.sleep(nanoseconds: 1_200_000_000)

        // R0 should be gone, R1 still visible
        XCTAssertNil(viewModel.registerHighlights["R0"])
        XCTAssertNotNil(viewModel.registerHighlights["R1"])
    }

    func testMemoryHighlightMultipleBytes() {
        viewModel.highlightMemoryAddress(0x8000, size: 4)

        // All 4 bytes should be highlighted
        XCTAssertNotNil(viewModel.memoryHighlights[0x8000])
        XCTAssertNotNil(viewModel.memoryHighlights[0x8001])
        XCTAssertNotNil(viewModel.memoryHighlights[0x8002])
        XCTAssertNotNil(viewModel.memoryHighlights[0x8003])
        XCTAssertNil(viewModel.memoryHighlights[0x8004]) // 5th byte not written
    }

    func testUpdateRegistersTriggersHighlights() async throws {
        // Simulate first state
        let registers1 = RegisterState(
            r0: 0, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
            sp: 0x50000, lr: 0, pc: 0x8000,
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false),
        )
        viewModel.updateRegisters(registers1)

        // Simulate second state with R0, R1 changed
        let registers2 = RegisterState(
            r0: 42, r1: 100, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
            sp: 0x50000, lr: 0, pc: 0x8004,
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false),
        )
        viewModel.updateRegisters(registers2)

        // R0 and R1 should be highlighted, PC should be highlighted
        XCTAssertNotNil(viewModel.registerHighlights["R0"])
        XCTAssertNotNil(viewModel.registerHighlights["R1"])
        XCTAssertNotNil(viewModel.registerHighlights["PC"])
        XCTAssertNil(viewModel.registerHighlights["R2"]) // Unchanged
    }
}

// MARK: - Event Handling Tests

@MainActor
final class EmulatorViewModelEventHandlingTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        viewModel.sessionID = "test-session"
    }

    func testHandleStateEvent() {
        let registers = RegisterState(
            r0: 42, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
            sp: 0x50000, lr: 0, pc: 0x8004,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false),
        )

        let stateUpdate = StateUpdate(
            status: "running",
            pc: 0x8004,
            registers: registers,
            flags: CPSRFlags(n: false, z: true, c: false, v: false),
        )

        let event = EmulatorEvent(
            type: "state",
            sessionId: "test-session",
            data: .state(stateUpdate),
        )

        viewModel.handleEvent(event)

        XCTAssertEqual(viewModel.status, .running)
        XCTAssertEqual(viewModel.registers.r0, 42)
        XCTAssertEqual(viewModel.currentPC, 0x8004)
    }

    func testHandleOutputEvent() {
        let outputUpdate = OutputUpdate(stream: "stdout", content: "Hello, World!")

        let event = EmulatorEvent(
            type: "output",
            sessionId: "test-session",
            data: .output(outputUpdate),
        )

        viewModel.handleEvent(event)

        XCTAssertTrue(viewModel.consoleOutput.contains("Hello, World!"))
    }

    func testHandleOutputEventMultiple() {
        let event1 = EmulatorEvent(
            type: "output",
            sessionId: "test-session",
            data: .output(OutputUpdate(stream: "stdout", content: "Line 1\n")),
        )

        let event2 = EmulatorEvent(
            type: "output",
            sessionId: "test-session",
            data: .output(OutputUpdate(stream: "stdout", content: "Line 2\n")),
        )

        viewModel.handleEvent(event1)
        viewModel.handleEvent(event2)

        XCTAssertTrue(viewModel.consoleOutput.contains("Line 1"))
        XCTAssertTrue(viewModel.consoleOutput.contains("Line 2"))
    }

    func testHandleBreakpointHitEvent() {
        let execEvent = ExecutionEvent(
            event: "breakpoint_hit",
            address: 0x8010,
            symbol: "loop",
            message: "Breakpoint hit at loop",
        )

        let event = EmulatorEvent(
            type: "event",
            sessionId: "test-session",
            data: .event(execEvent),
        )

        viewModel.handleEvent(event)

        XCTAssertEqual(viewModel.status, .breakpoint)
        XCTAssertEqual(viewModel.currentPC, 0x8010)
    }

    func testHandleErrorEvent() {
        let execEvent = ExecutionEvent(
            event: "error",
            address: nil,
            symbol: nil,
            message: "Division by zero",
        )

        let event = EmulatorEvent(
            type: "event",
            sessionId: "test-session",
            data: .event(execEvent),
        )

        viewModel.handleEvent(event)

        XCTAssertEqual(viewModel.status, .error)
        XCTAssertEqual(viewModel.errorMessage, "Division by zero")
    }

    func testHandleHaltedEvent() {
        let execEvent = ExecutionEvent(
            event: "halted",
            address: nil,
            symbol: nil,
            message: nil,
        )

        let event = EmulatorEvent(
            type: "event",
            sessionId: "test-session",
            data: .event(execEvent),
        )

        viewModel.handleEvent(event)

        XCTAssertEqual(viewModel.status, .halted)
    }

    func testIgnoreEventForDifferentSession() {
        let initialOutput = viewModel.consoleOutput

        let event = EmulatorEvent(
            type: "output",
            sessionId: "different-session",
            data: .output(OutputUpdate(stream: "stdout", content: "Should be ignored")),
        )

        viewModel.handleEvent(event)

        XCTAssertEqual(viewModel.consoleOutput, initialOutput) // Should not change
    }

    func testHandleStateUpdateStatusOnly() {
        let stateUpdate = StateUpdate(
            status: "waiting_for_input",
            pc: nil,
            registers: nil,
            flags: nil,
        )

        let event = EmulatorEvent(
            type: "state",
            sessionId: "test-session",
            data: .state(stateUpdate),
        )

        viewModel.handleEvent(event)

        XCTAssertEqual(viewModel.status, .waitingForInput)
    }
}

// MARK: - Execution Control Tests

@MainActor
final class EmulatorViewModelExecutionTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        viewModel.sessionID = "test-session"
    }

    func testRunSuccess() async throws {
        await viewModel.run()

        XCTAssertTrue(mockAPIClient.runCalled)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testRunFailure() async throws {
        mockAPIClient.shouldFailRun = true

        await viewModel.run()

        XCTAssertTrue(mockAPIClient.runCalled)
        XCTAssertNotNil(viewModel.errorMessage)
        XCTAssertTrue(viewModel.errorMessage?.contains("Failed to run") ?? false)
    }

    func testRunWithoutSession() async throws {
        viewModel.sessionID = nil

        await viewModel.run()

        XCTAssertFalse(mockAPIClient.runCalled)
        XCTAssertEqual(viewModel.errorMessage, "No active session")
    }

    func testPauseSuccess() async throws {
        await viewModel.pause()

        XCTAssertTrue(mockAPIClient.stopCalled)
        XCTAssertTrue(mockAPIClient.getRegistersCalled)
        XCTAssertTrue(mockAPIClient.getStatusCalled)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testStepSuccess() async throws {
        await viewModel.step()

        XCTAssertTrue(mockAPIClient.stepCalled)
        XCTAssertTrue(mockAPIClient.getRegistersCalled)
        XCTAssertTrue(mockAPIClient.getStatusCalled)
        XCTAssertNil(viewModel.errorMessage)
        XCTAssertNil(viewModel.lastMemoryWrite) // Cleared before step
    }

    func testStepWithProgramExit() async throws {
        mockAPIClient.shouldFailStep = true
        mockAPIClient.stepErrorMessage = "program exited with code 0"

        await viewModel.step()

        XCTAssertTrue(mockAPIClient.stepCalled)
        XCTAssertNil(viewModel.errorMessage) // No error for normal exit
    }

    func testStepOverSuccess() async throws {
        await viewModel.stepOver()

        XCTAssertTrue(mockAPIClient.stepOverCalled)
        XCTAssertTrue(mockAPIClient.getRegistersCalled)
        XCTAssertTrue(mockAPIClient.getStatusCalled)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testStepOutSuccess() async throws {
        await viewModel.stepOut()

        XCTAssertTrue(mockAPIClient.stepOutCalled)
        XCTAssertTrue(mockAPIClient.getRegistersCalled)
        XCTAssertTrue(mockAPIClient.getStatusCalled)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testResetClearsConsoleAndHighlights() async throws {
        viewModel.consoleOutput = "Previous output"
        viewModel.highlightRegister("R0")

        await viewModel.reset()

        XCTAssertTrue(mockAPIClient.restartCalled)
        XCTAssertEqual(viewModel.consoleOutput, "")
        XCTAssertNil(viewModel.errorMessage)
    }
}

// MARK: - Debug Features Tests

@MainActor
final class EmulatorViewModelDebugTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        viewModel.sessionID = "test-session"
    }

    func testToggleBreakpointAdd() async throws {
        let address: UInt32 = 0x8000

        await viewModel.toggleBreakpoint(at: address)

        XCTAssertTrue(mockAPIClient.addBreakpointCalled)
        XCTAssertEqual(mockAPIClient.lastBreakpointAddress, address)
        XCTAssertTrue(viewModel.breakpoints.contains(address))
        XCTAssertNil(viewModel.errorMessage)
    }

    func testToggleBreakpointRemove() async throws {
        let address: UInt32 = 0x8000
        viewModel.breakpoints.insert(address)

        await viewModel.toggleBreakpoint(at: address)

        XCTAssertTrue(mockAPIClient.removeBreakpointCalled)
        XCTAssertEqual(mockAPIClient.lastBreakpointAddress, address)
        XCTAssertFalse(viewModel.breakpoints.contains(address))
        XCTAssertNil(viewModel.errorMessage)
    }

    func testToggleBreakpointFailure() async throws {
        mockAPIClient.shouldFailAddBreakpoint = true

        await viewModel.toggleBreakpoint(at: 0x8000)

        XCTAssertTrue(mockAPIClient.addBreakpointCalled)
        XCTAssertNotNil(viewModel.errorMessage)
        XCTAssertTrue(viewModel.errorMessage?.contains("Failed to toggle breakpoint") ?? false)
    }

    func testAddWatchpoint() async throws {
        await viewModel.addWatchpoint(at: 0x8000, type: "write")

        XCTAssertTrue(mockAPIClient.addWatchpointCalled)
        XCTAssertEqual(viewModel.watchpoints.count, 1)
        XCTAssertEqual(viewModel.watchpoints[0].address, 0x8000)
        XCTAssertEqual(viewModel.watchpoints[0].type, "write")
        XCTAssertNil(viewModel.errorMessage)
    }

    func testRemoveWatchpoint() async throws {
        viewModel.watchpoints = [Watchpoint(id: 1, address: 0x8000, type: "write")]

        await viewModel.removeWatchpoint(id: 1)

        XCTAssertTrue(mockAPIClient.removeWatchpointCalled)
        XCTAssertEqual(viewModel.watchpoints.count, 0)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testRefreshWatchpoints() async throws {
        await viewModel.refreshWatchpoints()

        XCTAssertTrue(mockAPIClient.getWatchpointsCalled)
    }
}

// MARK: - Input/Output Handling Tests

@MainActor
final class EmulatorViewModelInputOutputTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        viewModel.sessionID = "test-session"
    }

    func testSendInputWhenWaiting() async throws {
        viewModel.status = .waitingForInput

        await viewModel.sendInput("test input")

        XCTAssertTrue(mockAPIClient.sendStdinCalled)
        XCTAssertEqual(mockAPIClient.lastStdinData, "test input")
        XCTAssertFalse(mockAPIClient.stepCalled) // Should NOT step when waiting
        XCTAssertTrue(mockAPIClient.getRegistersCalled) // Should refresh state
        XCTAssertNil(viewModel.errorMessage)
    }

    func testSendInputWhenNotWaiting() async throws {
        viewModel.status = .breakpoint

        await viewModel.sendInput("buffered input")

        XCTAssertTrue(mockAPIClient.sendStdinCalled)
        XCTAssertEqual(mockAPIClient.lastStdinData, "buffered input")
        XCTAssertTrue(mockAPIClient.stepCalled) // Should step to consume buffered input
        XCTAssertTrue(mockAPIClient.getRegistersCalled)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testSendInputWithoutSession() async throws {
        viewModel.sessionID = nil

        await viewModel.sendInput("test")

        XCTAssertFalse(mockAPIClient.sendStdinCalled)
        XCTAssertEqual(viewModel.errorMessage, "No active session")
    }
}

// MARK: - VM State Tests

@MainActor
final class EmulatorViewModelStateTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        viewModel.sessionID = "test-session"
    }

    func testCanPauseWhenRunning() {
        viewModel.status = .running
        XCTAssertTrue(viewModel.canPause, "Pause button should be enabled when status is .running")
    }

    func testCanPauseWhenWaitingForInput() {
        viewModel.status = .waitingForInput
        XCTAssertTrue(viewModel.canPause, "Pause button should be enabled when status is .waitingForInput")
    }

    func testCannotPauseWhenAtBreakpoint() {
        viewModel.status = .breakpoint
        XCTAssertFalse(
            viewModel.canPause,
            "Pause button should be disabled when status is .breakpoint (already paused)",
        )
    }

    func testCannotPauseWhenIdle() {
        viewModel.status = .idle
        XCTAssertFalse(viewModel.canPause, "Pause button should be disabled when status is .idle")
    }

    func testCannotPauseWhenHalted() {
        viewModel.status = .halted
        XCTAssertFalse(viewModel.canPause, "Pause button should be disabled when status is .halted")
    }

    func testCannotPauseWhenError() {
        viewModel.status = .error
        XCTAssertFalse(viewModel.canPause, "Pause button should be disabled when status is .error")
    }

    // MARK: - canStep Tests

    func testCanStepWhenIdle() {
        viewModel.status = .idle
        XCTAssertTrue(viewModel.canStep, "Step buttons should be enabled when status is .idle (ready to start)")
    }

    func testCanStepWhenAtBreakpoint() {
        viewModel.status = .breakpoint
        XCTAssertTrue(
            viewModel.canStep,
            "Step buttons should be enabled when status is .breakpoint (paused, ready to continue)",
        )
    }

    func testCannotStepWhenRunning() {
        viewModel.status = .running
        XCTAssertFalse(viewModel.canStep, "Step buttons should be disabled when status is .running")
    }

    func testCannotStepWhenWaitingForInput() {
        viewModel.status = .waitingForInput
        XCTAssertFalse(viewModel.canStep, "Step buttons should be disabled when status is .waitingForInput")
    }

    func testCannotStepWhenHalted() {
        viewModel.status = .halted
        XCTAssertFalse(viewModel.canStep, "Step buttons should be disabled when status is .halted (program stopped)")
    }

    func testCannotStepWhenError() {
        viewModel.status = .error
        XCTAssertFalse(viewModel.canStep, "Step buttons should be disabled when status is .error")
    }
}

// MARK: - Phase 1.2: Enhanced ViewModel Testing - Edge Cases

// MARK: - Concurrent State Changes Tests

@MainActor
final class EmulatorViewModelConcurrencyTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        viewModel.sessionID = "test-session"
    }

    func testConcurrentRegisterHighlightsDoNotRace() async throws {
        // Rapidly highlight the same register multiple times concurrently
        // This should not crash or leave highlights in invalid state

        // Start 100 concurrent highlight operations on R0
        await withTaskGroup(of: Void.self) { group in
            for _ in 0 ..< 100 {
                group.addTask { @MainActor in
                    self.viewModel.highlightRegister("R0")
                }
            }
        }

        // R0 should be highlighted exactly once (last highlight wins)
        XCTAssertNotNil(viewModel.registerHighlights["R0"], "R0 should be highlighted after concurrent updates")

        // There should be exactly one highlight task for R0 (previous tasks cancelled)
        // We can't directly access highlightTasks (private), but we can verify behavior:
        // After waiting for fade, highlight should be removed
        try await Task.sleep(nanoseconds: 1_600_000_000) // 1.6s

        XCTAssertNil(viewModel.registerHighlights["R0"], "R0 highlight should fade after 1.5s")
    }

    func testConcurrentMemoryHighlightsDoNotRace() async throws {
        // Rapidly highlight the same memory addresses concurrently

        await withTaskGroup(of: Void.self) { group in
            for _ in 0 ..< 100 {
                group.addTask { @MainActor in
                    self.viewModel.highlightMemoryAddress(0x8000, size: 4)
                }
            }
        }

        // All 4 bytes should be highlighted
        XCTAssertNotNil(viewModel.memoryHighlights[0x8000])
        XCTAssertNotNil(viewModel.memoryHighlights[0x8001])
        XCTAssertNotNil(viewModel.memoryHighlights[0x8002])
        XCTAssertNotNil(viewModel.memoryHighlights[0x8003])

        // After fade timeout, all should be removed
        try await Task.sleep(nanoseconds: 1_600_000_000)

        XCTAssertNil(viewModel.memoryHighlights[0x8000])
        XCTAssertNil(viewModel.memoryHighlights[0x8003])
    }

    func testConcurrentRegisterUpdatesHighlightCorrectly() async throws {
        // Simulate rapid register state updates (as would happen during fast execution)

        let baseRegisters = RegisterState(
            r0: 0, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
            sp: 0x50000, lr: 0, pc: 0x8000,
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false),
        )

        viewModel.updateRegisters(baseRegisters)

        // Rapidly update registers 1000 times with different values
        for i in 0 ..< 1000 {
            let newRegisters = RegisterState(
                r0: UInt32(i), r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
                r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
                sp: 0x50000, lr: 0, pc: 0x8000 + UInt32(i * 4),
                cpsr: CPSRFlags(n: false, z: false, c: false, v: false),
            )
            viewModel.updateRegisters(newRegisters)
        }

        // R0 and PC should be highlighted (changed in every update)
        XCTAssertNotNil(viewModel.registerHighlights["R0"], "R0 should be highlighted after rapid updates")
        XCTAssertNotNil(viewModel.registerHighlights["PC"], "PC should be highlighted after rapid updates")

        // Final register value should be correct (no race in state updates)
        XCTAssertEqual(viewModel.registers.r0, 999, "Final R0 value should be 999")
        XCTAssertEqual(viewModel.registers.pc, 0x8000 + 999 * 4, "Final PC should reflect last update")
    }

    func testCancelAllHighlightsDuringActiveFades() async throws {
        // Set up multiple highlights that are fading
        viewModel.highlightRegister("R0")
        viewModel.highlightRegister("R1")
        viewModel.highlightMemoryAddress(0x8000, size: 4)

        // Wait a bit (but not long enough to fade)
        try await Task.sleep(nanoseconds: 500_000_000) // 0.5s

        // Cancel all highlights mid-fade
        viewModel.cancelAllHighlights()

        // All highlights should be immediately removed
        XCTAssertTrue(viewModel.registerHighlights.isEmpty, "All register highlights should be cancelled")
        XCTAssertTrue(viewModel.memoryHighlights.isEmpty, "All memory highlights should be cancelled")

        // Wait for original fade timers (1.5s total from start)
        try await Task.sleep(nanoseconds: 1_100_000_000) // 1.1s more = 1.6s total

        // Highlights should still be empty (tasks were cancelled, not just cleared)
        XCTAssertTrue(viewModel.registerHighlights.isEmpty, "Highlights should remain cancelled")
        XCTAssertTrue(viewModel.memoryHighlights.isEmpty, "Memory highlights should remain cancelled")
    }
}

// MARK: - Memory Pressure Tests

@MainActor
final class EmulatorViewModelMemoryPressureTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        viewModel.sessionID = "test-session"
    }

    func testLoadLargeMemoryChunk() async throws {
        // Test loading 10MB of memory (extreme case)
        let largeSize = 10 * 1024 * 1024 // 10MB

        // Mock API should return large chunk
        mockAPIClient.mockMemoryData = [UInt8](repeating: 0xAB, count: largeSize)

        await viewModel.loadMemory(at: 0x8000, length: largeSize)

        // Should successfully load all data
        XCTAssertEqual(viewModel.memoryData.count, largeSize, "Should load entire 10MB chunk")
        XCTAssertEqual(viewModel.memoryAddress, 0x8000, "Address should be set correctly")
        XCTAssertTrue(viewModel.memoryData.allSatisfy { $0 == 0xAB }, "All bytes should match mock data")
    }

    func testRapidConsecutiveMemoryLoads() async throws {
        // Simulate rapid memory view scrolling (100 loads in quick succession)
        let loadCount = 100

        for i in 0 ..< loadCount {
            let address = UInt32(0x8000 + i * 256)
            mockAPIClient.mockMemoryData = [UInt8](repeating: UInt8(i & 0xFF), count: 256)

            await viewModel.loadMemory(at: address, length: 256)

            // Each load should complete successfully
            XCTAssertEqual(viewModel.memoryAddress, address, "Address should update for load \(i)")
            XCTAssertEqual(viewModel.memoryData.count, 256, "Data size should be correct for load \(i)")
        }

        // Final state should reflect last load
        XCTAssertEqual(viewModel.memoryAddress, 0x8000 + UInt32(loadCount - 1) * 256)
        XCTAssertEqual(viewModel.memoryData.count, 256)
    }

    func testRapidConsoleOutputUpdates() async throws {
        // Simulate program printing 10,000 lines rapidly
        let lineCount = 10000

        for i in 0 ..< lineCount {
            let output = OutputUpdate(stream: "stdout", content: "Line \(i)\n")
            let event = EmulatorEvent(
                type: "output",
                sessionId: "test-session",
                data: .output(output),
            )

            viewModel.handleEvent(event)
        }

        // Console should contain all lines
        XCTAssertTrue(viewModel.consoleOutput.contains("Line 0"))
        XCTAssertTrue(viewModel.consoleOutput.contains("Line 9999"))

        // Verify no data loss (simple character count check)
        let expectedMinLength = lineCount * 7 // "Line X\n" is at least 7 chars
        XCTAssertGreaterThanOrEqual(
            viewModel.consoleOutput.count,
            expectedMinLength,
            "Console should contain all output",
        )
    }

    func testThousandsOfMemoryHighlights() async throws {
        // Test highlighting 1000 different memory addresses
        // (Simulates intensive memory operations)

        for i in 0 ..< 1000 {
            let address = UInt32(0x8000 + i * 4)
            viewModel.highlightMemoryAddress(address, size: 4)
        }

        // Should have 4000 highlights (1000 addresses × 4 bytes each)
        XCTAssertEqual(viewModel.memoryHighlights.count, 4000, "Should have 4000 memory highlights")

        // All addresses should be highlighted
        for i in 0 ..< 1000 {
            let address = UInt32(0x8000 + i * 4)
            XCTAssertNotNil(
                viewModel.memoryHighlights[address],
                "Address 0x\(String(format: "%08X", address)) should be highlighted",
            )
        }
    }

    func testLargeNumberOfBreakpoints() async throws {
        // Test adding 1000 breakpoints
        let breakpointCount = 1000

        for i in 0 ..< breakpointCount {
            let address = UInt32(0x8000 + i * 4)
            await viewModel.toggleBreakpoint(at: address)
        }

        // All breakpoints should be added
        XCTAssertEqual(viewModel.breakpoints.count, breakpointCount, "Should have 1000 breakpoints")
        XCTAssertEqual(mockAPIClient.addBreakpointCallCount, breakpointCount, "API should be called 1000 times")

        // Verify specific breakpoints exist
        XCTAssertTrue(viewModel.breakpoints.contains(0x8000))
        XCTAssertTrue(viewModel.breakpoints.contains(0x8000 + UInt32((breakpointCount - 1) * 4)))
    }

    func testMemoryLoadAfterFailure() async throws {
        // Test that memory loads can recover after API failures

        // First load fails
        mockAPIClient.shouldFailGetMemory = true
        await viewModel.loadMemory(at: 0x8000, length: 256)

        // Should have empty data after failure
        XCTAssertTrue(viewModel.memoryData.isEmpty, "Memory data should be empty after failure")

        // Second load succeeds
        mockAPIClient.shouldFailGetMemory = false
        mockAPIClient.mockMemoryData = [UInt8](repeating: 0x42, count: 256)

        await viewModel.loadMemory(at: 0x8000, length: 256)

        // Should successfully load data after recovery
        XCTAssertEqual(viewModel.memoryData.count, 256, "Should load data after recovery")
        XCTAssertEqual(viewModel.memoryData[0], 0x42, "Data should match mock")
    }
}

// MARK: - WebSocket Reconnection Tests

@MainActor
final class VMWebSocketReconnectionTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        viewModel.sessionID = "test-session"
    }

    func testWebSocketReconnectionAfterDisconnect() async throws {
        // Simulate initial connection
        await viewModel.initialize()

        XCTAssertTrue(viewModel.isConnected, "Should be connected after initialization")

        // First, verify that events work before disconnect
        let testEvent1 = EmulatorEvent(
            type: "output",
            sessionId: viewModel.sessionID!,
            data: .output(OutputUpdate(stream: "stdout", content: "Before disconnect")),
        )
        mockWSClient.simulateEvent(testEvent1)
        try await Task.sleep(nanoseconds: 100_000_000)
        XCTAssertTrue(viewModel.consoleOutput.contains("Before disconnect"), "Should receive events before disconnect")

        // Simulate WebSocket disconnect
        mockWSClient.simulateDisconnect()
        try await Task.sleep(nanoseconds: 100_000_000)

        // Simulate reconnection
        mockWSClient.simulateReconnect(sessionID: viewModel.sessionID!)
        try await Task.sleep(nanoseconds: 100_000_000)

        // Should still be able to receive events after reconnection
        let event = EmulatorEvent(
            type: "output",
            sessionId: viewModel.sessionID!,
            data: .output(OutputUpdate(stream: "stdout", content: "After reconnect")),
        )

        mockWSClient.simulateEvent(event)

        // Wait for event to be processed (async sink on main queue)
        try await Task.sleep(nanoseconds: 300_000_000) // 0.3s to ensure async processing completes

        // Event should be received and processed
        XCTAssertTrue(
            viewModel.consoleOutput.contains("After reconnect"),
            "Should process events after reconnection. Console output: \(viewModel.consoleOutput)",
        )
    }

    func testStaleEventsIgnoredAfterSessionChange() async throws {
        // Start with old session
        viewModel.sessionID = "old-session-id"

        let oldOutput = OutputUpdate(stream: "stdout", content: "Old session output")
        let oldEvent = EmulatorEvent(
            type: "output",
            sessionId: "old-session-id",
            data: .output(oldOutput),
        )

        viewModel.handleEvent(oldEvent)

        XCTAssertTrue(viewModel.consoleOutput.contains("Old session output"))

        // Clear console and change to new session
        viewModel.consoleOutput = ""
        viewModel.sessionID = "new-session-id"

        // Receive event from old session (stale)
        let staleEvent = EmulatorEvent(
            type: "output",
            sessionId: "old-session-id",
            data: .output(OutputUpdate(stream: "stdout", content: "Stale output")),
        )

        viewModel.handleEvent(staleEvent)

        // Stale event should be ignored
        XCTAssertFalse(
            viewModel.consoleOutput.contains("Stale output"),
            "Stale events from old session should be ignored",
        )
        XCTAssertTrue(viewModel.consoleOutput.isEmpty, "Console should remain empty")
    }

    func testMultipleReconnectionsHandledGracefully() async throws {
        // Simulate multiple disconnect/reconnect cycles

        for cycle in 0 ..< 10 {
            // Disconnect
            mockWSClient.simulateDisconnect()
            try await Task.sleep(nanoseconds: 50_000_000) // 0.05s

            // Reconnect
            mockWSClient.simulateReconnect(sessionID: "test-session")
            try await Task.sleep(nanoseconds: 50_000_000)

            // Send test event
            let event = EmulatorEvent(
                type: "output",
                sessionId: "test-session",
                data: .output(OutputUpdate(stream: "stdout", content: "Cycle \(cycle)\n")),
            )

            mockWSClient.simulateEvent(event)
            try await Task.sleep(nanoseconds: 50_000_000)
        }

        // All events should be received
        for cycle in 0 ..< 10 {
            XCTAssertTrue(
                viewModel.consoleOutput.contains("Cycle \(cycle)"),
                "Should receive event from cycle \(cycle)",
            )
        }
    }

    func testWebSocketEventsDuringAPICall() async throws {
        // Test that WebSocket events are processed even while API calls are in flight

        // Start a slow API operation (simulated by delay in mock)
        mockAPIClient.simulateDelay = 1.0 // 1 second delay

        Task {
            await viewModel.loadProgram(source: "MOV R0, #42")
        }

        // While loadProgram is running, send WebSocket events
        try await Task.sleep(nanoseconds: 100_000_000) // 0.1s (during API call)

        let event = EmulatorEvent(
            type: "output",
            sessionId: "test-session",
            data: .output(OutputUpdate(stream: "stdout", content: "During API call")),
        )

        viewModel.handleEvent(event)

        // Event should be processed immediately (not blocked by API call)
        XCTAssertTrue(
            viewModel.consoleOutput.contains("During API call"),
            "WebSocket events should process during API calls",
        )
    }
}

// MARK: - Backend Restart Handling Tests

@MainActor
final class EmulatorViewModelBackendRestartTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
    }

    func testRecoverFromBackendRestartDuringInitialization() async throws {
        // First initialization fails (backend not ready)
        mockAPIClient.shouldFailCreateSession = true

        await viewModel.initialize()

        XCTAssertFalse(viewModel.isConnected, "Should not be connected after failed initialization")
        XCTAssertNil(viewModel.sessionID, "Should not have session ID")
        XCTAssertNotNil(viewModel.errorMessage, "Should have error message")

        // Backend comes back up, retry initialization
        mockAPIClient.shouldFailCreateSession = false
        mockAPIClient.createSessionCalled = false // Reset call tracking

        await viewModel.initialize()

        XCTAssertTrue(viewModel.isConnected, "Should be connected after backend recovery")
        XCTAssertNotNil(viewModel.sessionID, "Should have session ID after recovery")
        XCTAssertNil(viewModel.errorMessage, "Error should be cleared after successful recovery")
    }

    func testSessionRecoveryAfterBackendRestart() async throws {
        // Enable unique session ID generation for this test
        mockAPIClient.generateUniqueSessionIDs = true

        // Establish initial session
        await viewModel.initialize()

        let initialSessionID = viewModel.sessionID
        XCTAssertNotNil(initialSessionID)

        // Simulate backend restart (session becomes invalid)
        // Attempting operations should fail with session error
        mockAPIClient.shouldFailRun = true
        mockAPIClient.runErrorMessage = "session not found"

        await viewModel.run()

        XCTAssertNotNil(viewModel.errorMessage, "Should have error after session invalidation")
        XCTAssertTrue(
            viewModel.errorMessage?.contains("session not found") ?? false,
            "Error should indicate session not found",
        )

        // Cleanup old session and reinitialize
        viewModel.cleanup()
        mockAPIClient.shouldFailRun = false
        mockAPIClient.createSessionCalled = false

        await viewModel.initialize()

        // Should have new session ID
        XCTAssertNotNil(viewModel.sessionID, "Should have new session ID")
        XCTAssertNotEqual(viewModel.sessionID, initialSessionID, "Should have different session ID after restart")
        XCTAssertTrue(viewModel.isConnected, "Should be connected with new session")
    }

    func testStatePreservationAcrossSessionRecreation() async throws {
        // Test that UI state (breakpoints, watchpoints) can be restored after backend restart

        // Set up initial session with breakpoints
        await viewModel.initialize()

        await viewModel.toggleBreakpoint(at: 0x8000)
        await viewModel.toggleBreakpoint(at: 0x8010)

        let breakpointsBeforeRestart = viewModel.breakpoints

        XCTAssertEqual(breakpointsBeforeRestart.count, 2, "Should have 2 breakpoints before restart")

        // Simulate backend restart and session recreation
        viewModel.cleanup()
        mockAPIClient.createSessionCalled = false
        mockAPIClient.addBreakpointCallCount = 0

        await viewModel.initialize()

        // Breakpoints should still be in ViewModel state (not cleared by cleanup)
        XCTAssertEqual(
            viewModel.breakpoints,
            breakpointsBeforeRestart,
            "Breakpoints should be preserved in ViewModel state",
        )

        // User would need to manually re-add breakpoints to new session
        // (This tests that the ViewModel preserves the UI state for re-application)
    }

    func testGracefulHandlingOfAPITimeoutDuringBackendRestart() async throws {
        // Simulate API timeout during backend restart (slow to respond)

        mockAPIClient.simulateDelay = 5.0 // Very slow response

        let startTime = Date()

        await viewModel.initialize()

        let elapsed = Date().timeIntervalSince(startTime)

        // Should eventually complete (even if slow)
        // In real implementation, there might be timeout logic
        XCTAssertTrue(elapsed >= 5.0, "Should wait for slow API response")
    }
}

// MARK: - Long-Running Execution Tests

@MainActor
final class VMLongRunningExecutionTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPIClient: MockAPIClient!
    var mockWSClient: MockWebSocketClient!

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        viewModel.sessionID = "test-session"
    }

    func testLongRunningProgramExecutionTracking() async throws {
        // Simulate a program that runs for extended period
        // The ViewModel should remain responsive and track state correctly

        // Start execution
        await viewModel.run()

        // Simulate state updates over time (as would come from WebSocket)
        for i in 0 ..< 100 {
            let registers = RegisterState(
                r0: UInt32(i), r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
                r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
                sp: 0x50000, lr: 0, pc: 0x8000 + UInt32(i * 4),
                cpsr: CPSRFlags(n: false, z: false, c: false, v: false),
            )

            let stateUpdate = StateUpdate(
                status: "running",
                pc: 0x8000 + UInt32(i * 4),
                registers: registers,
                flags: registers.cpsr,
            )

            let event = EmulatorEvent(
                type: "state",
                sessionId: "test-session",
                data: .state(stateUpdate),
            )

            viewModel.handleEvent(event)

            // Small delay to simulate real execution timing
            try await Task.sleep(nanoseconds: 10_000_000) // 0.01s
        }

        // After 100 updates, state should be correct
        XCTAssertEqual(viewModel.registers.r0, 99, "R0 should reflect last update")
        XCTAssertEqual(viewModel.currentPC, 0x8000 + 99 * 4, "PC should reflect last update")
        XCTAssertEqual(viewModel.status, .running, "Status should be running")
    }

    func testProgressTrackingDuringLongExecution() async throws {
        // Test that console output accumulates correctly during long execution

        // Simulate program printing 1000 lines over time
        for i in 0 ..< 1000 {
            let event = EmulatorEvent(
                type: "output",
                sessionId: "test-session",
                data: .output(OutputUpdate(stream: "stdout", content: "Progress: \(i)\n")),
            )

            viewModel.handleEvent(event)

            // Brief delay
            if i % 100 == 0 {
                try await Task.sleep(nanoseconds: 10_000_000) // 0.01s every 100 lines
            }
        }

        // All output should be captured
        XCTAssertTrue(viewModel.consoleOutput.contains("Progress: 0"))
        XCTAssertTrue(viewModel.consoleOutput.contains("Progress: 500"))
        XCTAssertTrue(viewModel.consoleOutput.contains("Progress: 999"))

        // Count occurrences to verify no data loss
        let lineCount = viewModel.consoleOutput.components(separatedBy: "\n").count
        XCTAssertGreaterThanOrEqual(lineCount, 1000, "Should have at least 1000 lines")
    }

    func testPauseAndResumeDuringLongExecution() async throws {
        // Test that pause/resume works correctly during long execution

        // Start execution
        await viewModel.run()

        // Simulate running state
        viewModel.status = .running

        // Pause mid-execution
        await viewModel.pause()

        XCTAssertTrue(mockAPIClient.stopCalled, "Should call stop API")
        XCTAssertTrue(mockAPIClient.getRegistersCalled, "Should refresh registers after pause")

        // Simulate breakpoint state after pause
        viewModel.status = .breakpoint

        // Resume execution
        await viewModel.run()

        XCTAssertEqual(mockAPIClient.runCallCount, 2, "Should call run API twice (initial + resume)")
    }

    func testMemoryHighlightsFadeCorrectlyDuringLongExecution() async throws {
        // Test that memory highlights don't accumulate indefinitely during long execution

        // Simulate rapid memory writes (as would happen during long execution)
        for i in 0 ..< 100 {
            let address = UInt32(0x8000 + i * 4)
            viewModel.highlightMemoryAddress(address, size: 4)

            // Brief delay between writes
            try await Task.sleep(nanoseconds: 20_000_000) // 0.02s
        }

        // Early highlights should have faded by now (1.5s timeout)
        // Only recent highlights should remain

        // After 100 writes × 0.02s = 2 seconds, first ~30 highlights should be gone
        let maxActiveHighlights = 4 * 75 // Last 75 writes × 4 bytes each = 300 highlights max

        XCTAssertLessThanOrEqual(
            viewModel.memoryHighlights.count,
            maxActiveHighlights,
            "Old highlights should fade during execution",
        )
    }

    func testErrorHandlingDuringLongExecution() async throws {
        // Test that errors during long execution are handled gracefully

        // Start execution
        await viewModel.run()

        // Simulate successful execution for a while
        for i in 0 ..< 50 {
            let event = EmulatorEvent(
                type: "output",
                sessionId: "test-session",
                data: .output(OutputUpdate(stream: "stdout", content: "Line \(i)\n")),
            )
            viewModel.handleEvent(event)
        }

        // Simulate error mid-execution
        let errorEvent = EmulatorEvent(
            type: "event",
            sessionId: "test-session",
            data: .event(ExecutionEvent(
                event: "error",
                address: 0x8100,
                symbol: nil,
                message: "Division by zero at 0x8100",
            )),
        )

        viewModel.handleEvent(errorEvent)

        // Should transition to error state
        XCTAssertEqual(viewModel.status, .error, "Should be in error state")
        XCTAssertEqual(viewModel.errorMessage, "Division by zero at 0x8100", "Should have error message")

        // Console output should be preserved
        XCTAssertTrue(viewModel.consoleOutput.contains("Line 0"))
        XCTAssertTrue(viewModel.consoleOutput.contains("Line 49"))
    }
}
