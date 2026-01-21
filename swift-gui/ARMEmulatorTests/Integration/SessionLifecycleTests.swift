import XCTest
@testable import ARMEmulator

/// Integration tests for complete emulator session lifecycle
/// Tests workflows from session creation through program execution to cleanup
@MainActor
final class SessionLifecycleTests: XCTestCase {
    var viewModel: EmulatorViewModel!

    override func setUp() async throws {
        try await super.setUp()
        viewModel = EmulatorViewModel(
            apiClient: MockAPIClient(),
            webSocketClient: MockWebSocketClient()
        )
    }

    override func tearDown() async throws {
        viewModel = nil
        try await super.tearDown()
    }

    // MARK: - Complete Session Lifecycle

    func testCompleteSessionLifecycle() async throws {
        // 1. Initialize ViewModel (simulates backend connection)
        viewModel.sessionID = "test-session"

        // 2. Load program
        await viewModel.loadProgram(source: ProgramFixtures.exitCode42)
        XCTAssertNil(viewModel.errorMessage, "Program load should succeed")
        XCTAssertEqual(viewModel.status, .idle)

        // 3. Execute program
        await viewModel.run()

        // 4. Wait for halted state
        try await waitForStatus(.halted, timeout: 2.0, viewModel: viewModel)
        XCTAssertEqual(viewModel.status, .halted)

        // 5. Verify registers updated
        XCTAssertTrue(viewModel.registers.hasRegister("R0", value: 42))

        // 6. Cleanup
        viewModel.cleanup()
        XCTAssertNil(viewModel.sessionID)
    }

    // MARK: - Program with Breakpoints

    func testProgramWithBreakpoints() async throws {
        viewModel.sessionID = "test-session"

        // Load program
        await viewModel.loadProgram(source: ProgramFixtures.withBreakpoint)
        XCTAssertNil(viewModel.errorMessage)

        // Set breakpoint at address 0x8004 (second instruction)
        await viewModel.toggleBreakpoint(at: 0x8004)

        // Run until breakpoint
        await viewModel.run()

        // Should hit breakpoint
        try await waitForStatus(.breakpoint, timeout: 2.0, viewModel: viewModel)
        XCTAssertEqual(viewModel.status, .breakpoint)

        // Verify we stopped at the right place
        XCTAssertTrue(viewModel.registers.hasRegister("PC", value: 0x8004))

        // Inspect registers at breakpoint
        XCTAssertTrue(viewModel.registers.hasRegister("R0", value: 1),
                     "R0 should be 1 after first instruction")

        // Continue execution
        await viewModel.run()
        try await waitForStatus(.halted, timeout: 2.0, viewModel: viewModel)
        XCTAssertEqual(viewModel.status, .halted)
    }

    // MARK: - Interactive Program

    func testInteractiveProgram() async throws {
        // Note: Interactive programs with stdin require special handling
        // This test uses the mock API which can simulate stdin responses

        viewModel.sessionID = "test-session"

        // Mock API client should be configured to provide stdin
        guard let mockAPI = viewModel.apiClient as? MockAPIClient else {
            XCTFail("Expected MockAPIClient")
            return
        }

        // Configure mock to simulate stdin input "5"
        mockAPI.mockedStdinInput = "5\n"

        // Load fibonacci program
        await viewModel.loadProgram(source: ProgramFixtures.fibonacci)
        XCTAssertNil(viewModel.errorMessage)

        // Run program (will read from mocked stdin)
        await viewModel.run()

        // Wait for completion
        try await waitForStatus(.halted, timeout: 5.0, viewModel: viewModel)

        // Verify output contains fibonacci numbers
        XCTAssertTrue(viewModel.consoleOutput.contains("0 1 1 2 3"),
                     "Console should contain fibonacci sequence")
    }

    // MARK: - Program with Errors

    func testProgramWithSyntaxError() async throws {
        viewModel.sessionID = "test-session"

        // Load program with syntax error
        await viewModel.loadProgram(source: ProgramFixtures.syntaxError)

        // Should have error message
        XCTAssertNotNil(viewModel.errorMessage, "Should have syntax error")
        XCTAssertTrue(viewModel.errorMessage?.contains("INVALID") ?? false,
                     "Error should mention invalid instruction")
    }

    func testProgramWithRuntimeError() async throws {
        // Program that causes runtime error (divide by zero or invalid memory access)
        let divideByZero = """
            MOV R0, #42
            MOV R1, #0
            ; ARM2 doesn't have DIV instruction, so simulate error with invalid memory
            LDR R2, [R1]     ; Invalid memory read at address 0
            SWI #0x00
            """

        viewModel.sessionID = "test-session"

        await viewModel.loadProgram(source: divideByZero)
        XCTAssertNil(viewModel.errorMessage, "Program should load successfully")

        await viewModel.run()

        // Wait for error or halted state
        try await waitForCondition(timeout: 2.0) {
            self.viewModel.status == .error || self.viewModel.status == .halted
        }

        // Should either be in error state or have error message
        if viewModel.status == .error {
            XCTAssertNotNil(viewModel.errorMessage)
        }
    }

    // MARK: - Long-Running Program

    func testLongRunningProgram() async throws {
        viewModel.sessionID = "test-session"

        // Load long-running program
        await viewModel.loadProgram(source: ProgramFixtures.longRunning)
        XCTAssertNil(viewModel.errorMessage)

        // Start execution
        await viewModel.run()

        // Wait a bit for it to start running
        try await Task.sleep(nanoseconds: 100_000_000) // 100ms

        // Should be in running state
        // (Note: Mock might transition to halted immediately - adjust expectations for real backend)

        // Stop execution manually
        await viewModel.stop()

        // Should transition to idle or breakpoint state
        try await waitForCondition(timeout: 2.0) {
            self.viewModel.status != .running
        }

        XCTAssertNotEqual(viewModel.status, .running, "Program should have stopped")
    }

    // MARK: - Reset Functionality

    func testResetProgram() async throws {
        viewModel.sessionID = "test-session"

        // Load and run program
        await viewModel.loadProgram(source: ProgramFixtures.simpleLoop)
        XCTAssertNil(viewModel.errorMessage)

        await viewModel.run()
        try await waitForStatus(.halted, timeout: 2.0, viewModel: viewModel)

        // Registers should be modified
        XCTAssertTrue(viewModel.registers.hasRegister("R0", value: 5),
                     "R0 should be 5 after loop completes")

        // Reset
        await viewModel.reset()

        // Should return to idle state
        XCTAssertEqual(viewModel.status, .idle)

        // Registers should be reset (or at initial values)
        // PC should be back at start
        let pc = viewModel.registers.registerValue("PC") ?? 0
        XCTAssertEqual(pc, 0x8000, "PC should be reset to start address")
    }

    // MARK: - Multiple Programs in Same Session

    func testMultipleProgramsInSession() async throws {
        viewModel.sessionID = "test-session"

        // Load and run first program
        await viewModel.loadProgram(source: ProgramFixtures.exitCode42)
        await viewModel.run()
        try await waitForStatus(.halted, timeout: 2.0, viewModel: viewModel)
        XCTAssertTrue(viewModel.registers.hasRegister("R0", value: 42))

        // Reset
        await viewModel.reset()

        // Load and run second program
        await viewModel.loadProgram(source: ProgramFixtures.simpleLoop)
        await viewModel.run()
        try await waitForStatus(.halted, timeout: 2.0, viewModel: viewModel)
        XCTAssertTrue(viewModel.registers.hasRegister("R0", value: 5))

        // Both programs should work independently
        XCTAssertNil(viewModel.errorMessage)
    }

    // MARK: - Session Cleanup

    func testSessionCleanup() async throws {
        viewModel.sessionID = "test-session"

        // Load program
        await viewModel.loadProgram(source: ProgramFixtures.helloWorld)
        XCTAssertNotNil(viewModel.sessionID)

        // Cleanup
        viewModel.cleanup()

        // Session should be cleared
        XCTAssertNil(viewModel.sessionID)

        // State should be reset
        XCTAssertEqual(viewModel.status, .idle)
        XCTAssertTrue(viewModel.consoleOutput.isEmpty)
    }
}
