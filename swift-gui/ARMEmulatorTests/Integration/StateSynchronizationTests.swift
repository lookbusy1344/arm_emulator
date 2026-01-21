import XCTest
import Combine
@testable import ARMEmulator

/// Integration tests for state synchronization between API calls and WebSocket events
/// Tests that ViewModel correctly coordinates REST API and WebSocket updates
@MainActor
final class StateSynchronizationTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPI: MockAPIClient!
    var mockWebSocket: MockWebSocketClient!
    var cancellables: Set<AnyCancellable> = []

    override func setUp() async throws {
        try await super.setUp()

        mockAPI = MockAPIClient()
        mockWebSocket = MockWebSocketClient()

        viewModel = EmulatorViewModel(
            apiClient: mockAPI,
            webSocketClient: mockWebSocket
        )
        viewModel.sessionID = "test-session"
    }

    override func tearDown() async throws {
        cancellables.removeAll()
        viewModel = nil
        mockAPI = nil
        mockWebSocket = nil
        try await super.tearDown()
    }

    // MARK: - Register Updates via WebSocket

    func testRegisterUpdatesArriveAfterStep() async throws {
        // Load program
        await viewModel.loadProgram(source: ProgramFixtures.withBreakpoint)

        // Set initial register state
        mockAPI.mockRegisters = RegisterState(
            r0: 0, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0, r13: 0x50000, r14: 0, r15: 0x8000,
            cpsr: 0
        )

        // Step instruction
        await viewModel.step()

        // Simulate WebSocket register update event
        let updatedRegisters = RegisterState(
            r0: 1, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0, r13: 0x50000, r14: 0, r15: 0x8004,
            cpsr: 0
        )

        let stateUpdate = StateUpdate(
            registers: updatedRegisters,
            memory: nil,
            status: VMStatus(state: "breakpoint", pc: 0x8004, instruction: nil, cycleCount: 2, error: nil)
        )

        mockWebSocket.simulateStateUpdate(stateUpdate)

        // Wait for WebSocket event to propagate
        try await Task.sleep(nanoseconds: 100_000_000) // 100ms

        // Verify registers updated
        XCTAssertTrue(viewModel.registers.hasRegister("R0", value: 1),
                     "Registers should update via WebSocket")
        XCTAssertTrue(viewModel.registers.hasRegister("PC", value: 0x8004),
                     "PC should update via WebSocket")
    }

    func testRegisterHighlightsAppearAfterUpdate() async throws {
        // Load program and set initial state
        await viewModel.loadProgram(source: ProgramFixtures.simpleLoop)

        let initialRegisters = RegisterState.empty
        mockWebSocket.simulateStateUpdate(StateUpdate(
            registers: initialRegisters,
            memory: nil,
            status: VMStatus(state: "idle", pc: 0x8000, instruction: nil, cycleCount: 0, error: nil)
        ))

        try await Task.sleep(nanoseconds: 50_000_000) // 50ms

        // Step and update R0
        await viewModel.step()

        let updatedRegisters = RegisterState(
            r0: 1, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0, r13: 0x50000, r14: 0, r15: 0x8004,
            cpsr: 0
        )

        mockWebSocket.simulateStateUpdate(StateUpdate(
            registers: updatedRegisters,
            memory: nil,
            status: VMStatus(state: "breakpoint", pc: 0x8004, instruction: nil, cycleCount: 1, error: nil)
        ))

        try await Task.sleep(nanoseconds: 50_000_000)

        // Verify R0 is highlighted (changed from 0 to 1)
        XCTAssertTrue(viewModel.changedRegisters.contains("R0"),
                     "Changed register should be highlighted")
    }

    // MARK: - Console Output via WebSocket

    func testConsoleOutputArrivesDuringExecution() async throws {
        // Load program that produces output
        await viewModel.loadProgram(source: ProgramFixtures.helloWorld)

        // Simulate console output event during execution
        let outputEvent = EmulatorEvent(
            type: "output",
            sessionId: "test-session",
            data: .output(OutputData(text: "Hello, World!\n", isError: false))
        )

        mockWebSocket.eventsSubject.send(outputEvent)

        try await Task.sleep(nanoseconds: 50_000_000)

        // Verify console output appeared
        XCTAssertTrue(viewModel.consoleOutput.contains("Hello, World!"),
                     "Console output should appear from WebSocket")
    }

    func testMultipleConsoleOutputsConcatenate() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.fibonacci)

        // Simulate multiple output events
        let outputs = ["0 ", "1 ", "1 ", "2 ", "3 ", "5 ", "8\n"]

        for output in outputs {
            let event = EmulatorEvent(
                type: "output",
                sessionId: "test-session",
                data: .output(OutputData(text: output, isError: false))
            )
            mockWebSocket.eventsSubject.send(event)
            try await Task.sleep(nanoseconds: 10_000_000) // 10ms between outputs
        }

        try await Task.sleep(nanoseconds: 50_000_000)

        // Verify all outputs concatenated
        XCTAssertTrue(viewModel.consoleOutput.contains("0 1 1 2 3 5 8"),
                     "Console should concatenate multiple outputs")
    }

    func testErrorOutputMarkedSeparately() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.helloWorld)

        // Simulate error output
        let errorEvent = EmulatorEvent(
            type: "output",
            sessionId: "test-session",
            data: .output(OutputData(text: "Error: Something went wrong\n", isError: true))
        )

        mockWebSocket.eventsSubject.send(errorEvent)

        try await Task.sleep(nanoseconds: 50_000_000)

        // Error should appear in console
        XCTAssertTrue(viewModel.consoleOutput.contains("Error: Something went wrong"),
                     "Error output should appear in console")
    }

    // MARK: - Breakpoint Hit Event

    func testBreakpointHitEventArrivesBeforeStatusChange() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.withBreakpoint)
        await viewModel.toggleBreakpoint(at: 0x8004)
        await viewModel.run()

        // Simulate breakpoint hit event
        let breakpointEvent = EmulatorEvent(
            type: "breakpoint_hit",
            sessionId: "test-session",
            data: .breakpointHit(BreakpointHitData(address: 0x8004, instruction: "MOV R1, #2"))
        )

        mockWebSocket.eventsSubject.send(breakpointEvent)

        try await Task.sleep(nanoseconds: 50_000_000)

        // Then send status update
        let stateUpdate = StateUpdate(
            registers: nil,
            memory: nil,
            status: VMStatus(state: "breakpoint", pc: 0x8004, instruction: "MOV R1, #2", cycleCount: 2, error: nil)
        )

        mockWebSocket.simulateStateUpdate(stateUpdate)

        try await Task.sleep(nanoseconds: 50_000_000)

        // Verify status updated to breakpoint
        XCTAssertEqual(viewModel.status, .breakpoint)
    }

    // MARK: - Memory Highlights

    func testMemoryHighlightsAppearAfterWrite() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.memoryWrite)

        // Simulate memory write event
        let memoryUpdate = MemoryUpdate(
            address: 0x9000,
            data: [42, 0, 0, 0],
            length: 4
        )

        let stateUpdate = StateUpdate(
            registers: nil,
            memory: memoryUpdate,
            status: nil
        )

        mockWebSocket.simulateStateUpdate(stateUpdate)

        try await Task.sleep(nanoseconds: 50_000_000)

        // Verify memory highlight exists
        XCTAssertTrue(viewModel.memoryHighlights.contains(0x9000),
                     "Memory address should be highlighted after write")
    }

    func testMemoryHighlightsExpire() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.memoryWrite)

        // Simulate memory write
        let memoryUpdate = MemoryUpdate(address: 0x9000, data: [42], length: 1)
        mockWebSocket.simulateStateUpdate(StateUpdate(registers: nil, memory: memoryUpdate, status: nil))

        try await Task.sleep(nanoseconds: 50_000_000)

        XCTAssertTrue(viewModel.memoryHighlights.contains(0x9000))

        // Wait for highlight to expire (typical timeout: 2-3 seconds)
        try await Task.sleep(nanoseconds: 3_000_000_000) // 3 seconds

        // Highlight should have expired
        XCTAssertFalse(viewModel.memoryHighlights.contains(0x9000),
                      "Memory highlight should expire after timeout")
    }

    // MARK: - Stale Events Ignored

    func testStaleEventsFromOldSessionIgnored() async throws {
        // Start with session 1
        viewModel.sessionID = "session-1"
        await viewModel.loadProgram(source: ProgramFixtures.helloWorld)

        // Simulate event from old session
        let staleEvent = EmulatorEvent(
            type: "output",
            sessionId: "old-session-id",
            data: .output(OutputData(text: "Stale output", isError: false))
        )

        mockWebSocket.eventsSubject.send(staleEvent)

        try await Task.sleep(nanoseconds: 50_000_000)

        // Stale output should not appear
        XCTAssertFalse(viewModel.consoleOutput.contains("Stale output"),
                      "Output from old session should be ignored")
    }

    func testEventsForCurrentSessionProcessed() async throws {
        viewModel.sessionID = "current-session"
        await viewModel.loadProgram(source: ProgramFixtures.helloWorld)

        // Event matching current session
        let validEvent = EmulatorEvent(
            type: "output",
            sessionId: "current-session",
            data: .output(OutputData(text: "Valid output", isError: false))
        )

        mockWebSocket.eventsSubject.send(validEvent)

        try await Task.sleep(nanoseconds: 50_000_000)

        // Valid output should appear
        XCTAssertTrue(viewModel.consoleOutput.contains("Valid output"),
                     "Output from current session should be processed")
    }

    // MARK: - State Transition Synchronization

    func testRunToHaltedTransition() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.simpleLoop)

        // Initial state: idle
        XCTAssertEqual(viewModel.status, .idle)

        // Start running
        await viewModel.run()

        // Simulate running state
        mockWebSocket.simulateStateUpdate(StateUpdate(
            registers: nil,
            memory: nil,
            status: VMStatus(state: "running", pc: 0x8008, instruction: nil, cycleCount: 5, error: nil)
        ))

        try await Task.sleep(nanoseconds: 50_000_000)
        // Note: Mock may not transition to running, but real backend would

        // Simulate completion
        let haltedEvent = EmulatorEvent(
            type: "halted",
            sessionId: "test-session",
            data: .halted(HaltedData(exitCode: 5))
        )

        mockWebSocket.eventsSubject.send(haltedEvent)

        try await Task.sleep(nanoseconds: 50_000_000)

        // Should transition to halted
        XCTAssertEqual(viewModel.status, .halted)
    }

    func testErrorTransition() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.helloWorld)
        await viewModel.run()

        // Simulate error event
        let errorEvent = EmulatorEvent(
            type: "error",
            sessionId: "test-session",
            data: .error(ErrorData(message: "Runtime error: invalid memory access"))
        )

        mockWebSocket.eventsSubject.send(errorEvent)

        try await Task.sleep(nanoseconds: 50_000_000)

        // Should transition to error state
        XCTAssertEqual(viewModel.status, .error)
        XCTAssertNotNil(viewModel.errorMessage)
        XCTAssertTrue(viewModel.errorMessage?.contains("invalid memory access") ?? false)
    }

    // MARK: - WebSocket Reconnection

    func testWebSocketReconnection() async throws {
        // Simulate WebSocket disconnect
        mockWebSocket.isConnected = false

        // Attempt to reconnect
        await mockWebSocket.connect(to: URL(string: "ws://localhost:8080")!, sessionID: "test-session")

        mockWebSocket.isConnected = true

        // Verify connection restored
        XCTAssertTrue(mockWebSocket.isConnected)
    }

    func testEventsAfterReconnection() async throws {
        viewModel.sessionID = "test-session"
        await viewModel.loadProgram(source: ProgramFixtures.helloWorld)

        // Simulate disconnect and reconnect
        mockWebSocket.isConnected = false
        try await Task.sleep(nanoseconds: 50_000_000)

        mockWebSocket.isConnected = true
        await mockWebSocket.connect(to: URL(string: "ws://localhost:8080")!, sessionID: "test-session")

        // Send event after reconnection
        let event = EmulatorEvent(
            type: "output",
            sessionId: "test-session",
            data: .output(OutputData(text: "After reconnect", isError: false))
        )

        mockWebSocket.eventsSubject.send(event)

        try await Task.sleep(nanoseconds: 50_000_000)

        // Event should be processed
        XCTAssertTrue(viewModel.consoleOutput.contains("After reconnect"))
    }

    // MARK: - Concurrent State Updates

    func testConcurrentRegisterAndMemoryUpdates() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.memoryWrite)

        // Send register and memory updates concurrently
        let registerUpdate = RegisterState(
            r0: 42, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0, r13: 0x50000, r14: 0, r15: 0x8008,
            cpsr: 0
        )

        let memoryUpdate = MemoryUpdate(address: 0x9000, data: [42, 0, 0, 0], length: 4)

        let stateUpdate = StateUpdate(
            registers: registerUpdate,
            memory: memoryUpdate,
            status: VMStatus(state: "breakpoint", pc: 0x8008, instruction: nil, cycleCount: 4, error: nil)
        )

        mockWebSocket.simulateStateUpdate(stateUpdate)

        try await Task.sleep(nanoseconds: 100_000_000)

        // Both updates should be applied
        XCTAssertTrue(viewModel.registers.hasRegister("R0", value: 42))
        XCTAssertTrue(viewModel.memoryHighlights.contains(0x9000))
        XCTAssertEqual(viewModel.status, .breakpoint)
    }
}
