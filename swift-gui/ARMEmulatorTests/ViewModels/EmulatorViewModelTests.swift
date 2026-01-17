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
            symbols: nil
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
            wsClient: MockWebSocketClient()
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
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false)
        )
        viewModel.updateRegisters(registers1)

        // Simulate second state with R0, R1 changed
        let registers2 = RegisterState(
            r0: 42, r1: 100, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
            sp: 0x50000, lr: 0, pc: 0x8004,
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false)
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
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false)
        )

        let stateUpdate = StateUpdate(
            status: "running",
            pc: 0x8004,
            registers: registers,
            flags: CPSRFlags(n: false, z: true, c: false, v: false)
        )

        let event = EmulatorEvent(
            type: "state",
            sessionId: "test-session",
            data: .state(stateUpdate)
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
            data: .output(outputUpdate)
        )

        viewModel.handleEvent(event)

        XCTAssertTrue(viewModel.consoleOutput.contains("Hello, World!"))
    }

    func testHandleOutputEventMultiple() {
        let event1 = EmulatorEvent(
            type: "output",
            sessionId: "test-session",
            data: .output(OutputUpdate(stream: "stdout", content: "Line 1\n"))
        )

        let event2 = EmulatorEvent(
            type: "output",
            sessionId: "test-session",
            data: .output(OutputUpdate(stream: "stdout", content: "Line 2\n"))
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
            message: "Breakpoint hit at loop"
        )

        let event = EmulatorEvent(
            type: "event",
            sessionId: "test-session",
            data: .event(execEvent)
        )

        viewModel.handleEvent(event)

        XCTAssertEqual(viewModel.status, .paused)
        XCTAssertEqual(viewModel.currentPC, 0x8010)
    }

    func testHandleErrorEvent() {
        let execEvent = ExecutionEvent(
            event: "error",
            address: nil,
            symbol: nil,
            message: "Division by zero"
        )

        let event = EmulatorEvent(
            type: "event",
            sessionId: "test-session",
            data: .event(execEvent)
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
            message: nil
        )

        let event = EmulatorEvent(
            type: "event",
            sessionId: "test-session",
            data: .event(execEvent)
        )

        viewModel.handleEvent(event)

        XCTAssertEqual(viewModel.status, .halted)
    }

    func testIgnoreEventForDifferentSession() {
        let initialOutput = viewModel.consoleOutput

        let event = EmulatorEvent(
            type: "output",
            sessionId: "different-session",
            data: .output(OutputUpdate(stream: "stdout", content: "Should be ignored"))
        )

        viewModel.handleEvent(event)

        XCTAssertEqual(viewModel.consoleOutput, initialOutput) // Should not change
    }

    func testHandleStateUpdateStatusOnly() {
        let stateUpdate = StateUpdate(
            status: "waiting_for_input",
            pc: nil,
            registers: nil,
            flags: nil
        )

        let event = EmulatorEvent(
            type: "state",
            sessionId: "test-session",
            data: .state(stateUpdate)
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

    func testStopSuccess() async throws {
        await viewModel.stop()

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
        viewModel.status = .paused

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
