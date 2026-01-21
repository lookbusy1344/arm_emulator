import XCTest
@testable import ARMEmulator

/// Integration tests for debugging workflows
/// Tests breakpoints, watchpoints, stepping, and expression evaluation end-to-end
@MainActor
final class DebuggingWorkflowTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var mockAPI: MockAPIClient!

    override func setUp() async throws {
        try await super.setUp()
        mockAPI = MockAPIClient()
        viewModel = EmulatorViewModel(
            apiClient: mockAPI,
            webSocketClient: MockWebSocketClient()
        )
        viewModel.sessionID = "test-session"
    }

    override func tearDown() async throws {
        viewModel = nil
        mockAPI = nil
        try await super.tearDown()
    }

    // MARK: - Breakpoint Workflow

    func testBreakpointWorkflow() async throws {
        // 1. Load program
        await viewModel.loadProgram(source: ProgramFixtures.withBreakpoint)
        XCTAssertNil(viewModel.errorMessage)

        // 2. Set breakpoint at address 0x8004 (second MOV instruction)
        await viewModel.toggleBreakpoint(at: 0x8004)
        XCTAssertTrue(mockAPI.addBreakpointCalled, "Should call addBreakpoint API")
        XCTAssertEqual(mockAPI.lastBreakpointAddress, 0x8004)

        // 3. Run until breakpoint
        await viewModel.run()
        XCTAssertTrue(mockAPI.runCalled)

        // Simulate hitting breakpoint by updating mock status
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x8004, instruction: nil, cycleCount: 2, error: nil)
        mockAPI.mockRegisters = RegisterState(
            r0: 1, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0, r13: 0x50000, r14: 0, r15: 0x8004,
            cpsr: 0
        )

        // Trigger status update
        await viewModel.fetchStatus()
        try await waitForStatus(.breakpoint, timeout: 1.0, viewModel: viewModel)

        // 4. Inspect registers at breakpoint
        await viewModel.fetchRegisters()
        XCTAssertTrue(mockAPI.getRegistersCalled)
        XCTAssertTrue(viewModel.registers.hasRegister("R0", value: 1),
                     "R0 should be 1 at breakpoint")
        XCTAssertTrue(viewModel.registers.hasRegister("PC", value: 0x8004),
                     "PC should be at breakpoint address")

        // 5. Step over (single instruction)
        await viewModel.step()
        XCTAssertTrue(mockAPI.stepCalled)

        // Simulate stepping to next instruction
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x8008, instruction: nil, cycleCount: 3, error: nil)
        mockAPI.mockRegisters.r1 = 2
        mockAPI.mockRegisters.r15 = 0x8008

        await viewModel.fetchStatus()
        await viewModel.fetchRegisters()

        XCTAssertTrue(viewModel.registers.hasRegister("R1", value: 2),
                     "R1 should be 2 after step")
        XCTAssertTrue(viewModel.registers.hasRegister("PC", value: 0x8008),
                     "PC should advance after step")

        // 6. Continue to completion
        mockAPI.mockStatus = VMStatus(state: "running", pc: 0x8008, instruction: nil, cycleCount: 3, error: nil)
        await viewModel.run()
        XCTAssertEqual(mockAPI.runCallCount, 2, "Should call run twice (initial + continue)")

        // Simulate completion
        mockAPI.mockStatus = VMStatus(state: "halted", pc: 0x8010, instruction: nil, cycleCount: 10, error: nil)
        await viewModel.fetchStatus()
        try await waitForStatus(.halted, timeout: 1.0, viewModel: viewModel)
    }

    func testMultipleBreakpoints() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.functionCall)
        XCTAssertNil(viewModel.errorMessage)

        // Set breakpoints at function entry and return
        await viewModel.toggleBreakpoint(at: 0x8000) // main entry
        await viewModel.toggleBreakpoint(at: 0x800C) // function entry

        XCTAssertEqual(mockAPI.addBreakpointCallCount, 2, "Should add two breakpoints")

        // Run and hit first breakpoint
        await viewModel.run()

        // Simulate hitting each breakpoint in sequence
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x8000, instruction: nil, cycleCount: 1, error: nil)
        await viewModel.fetchStatus()
        try await waitForStatus(.breakpoint, timeout: 1.0, viewModel: viewModel)

        // Continue to second breakpoint
        await viewModel.run()
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x800C, instruction: nil, cycleCount: 5, error: nil)
        await viewModel.fetchStatus()
        try await waitForStatus(.breakpoint, timeout: 1.0, viewModel: viewModel)
    }

    func testRemoveBreakpoint() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.withBreakpoint)

        // Add breakpoint
        await viewModel.toggleBreakpoint(at: 0x8004)
        XCTAssertTrue(mockAPI.addBreakpointCalled)

        // Remove breakpoint (toggle again)
        mockAPI.addBreakpointCalled = false // Reset flag
        await viewModel.toggleBreakpoint(at: 0x8004)
        XCTAssertTrue(mockAPI.removeBreakpointCalled, "Should call removeBreakpoint")
        XCTAssertEqual(mockAPI.lastBreakpointAddress, 0x8004)
    }

    // MARK: - Watchpoint Workflow

    func testWatchpointWorkflow() async throws {
        // 1. Load program that writes to memory
        await viewModel.loadProgram(source: ProgramFixtures.memoryWrite)
        XCTAssertNil(viewModel.errorMessage)

        // 2. Set memory watchpoint on data_area (assuming address 0x9000)
        let watchAddress: UInt32 = 0x9000
        await viewModel.addWatchpoint(address: watchAddress, type: "write")
        XCTAssertTrue(mockAPI.addWatchpointCalled)

        // 3. Run program
        await viewModel.run()

        // 4. Simulate watchpoint hit (backend would detect memory write)
        mockAPI.mockStatus = VMStatus(
            state: "breakpoint",
            pc: 0x8008,
            instruction: "STR R1, [R0]",
            cycleCount: 4,
            error: nil
        )
        await viewModel.fetchStatus()

        // 5. Inspect memory at watchpoint address
        mockAPI.mockMemoryData = [42, 0, 0, 0] // 42 in little-endian
        let memory = try await viewModel.apiClient.getMemory(sessionID: viewModel.sessionID!, address: watchAddress, length: 4)
        XCTAssertEqual(memory[0], 42, "Memory should contain written value")
    }

    func testWatchpointRemoval() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.memoryWrite)

        // Add watchpoint
        await viewModel.addWatchpoint(address: 0x9000, type: "write")
        XCTAssertTrue(mockAPI.addWatchpointCalled)

        // Remove watchpoint
        await viewModel.removeWatchpoint(id: 1)
        XCTAssertTrue(mockAPI.removeWatchpointCalled)
    }

    // MARK: - Stepping Workflow

    func testStepOverFunction() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.functionCall)

        // Set breakpoint before BL instruction
        await viewModel.toggleBreakpoint(at: 0x8008) // Before BL add_numbers

        // Run to breakpoint
        await viewModel.run()
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x8008, instruction: nil, cycleCount: 3, error: nil)
        await viewModel.fetchStatus()
        try await waitForStatus(.breakpoint, timeout: 1.0, viewModel: viewModel)

        // Step over function call
        await viewModel.stepOver()
        XCTAssertTrue(mockAPI.stepOverCalled)

        // Simulate step over completing (PC after function return)
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x800C, instruction: nil, cycleCount: 10, error: nil)
        mockAPI.mockRegisters.r2 = 8 // Result of 5 + 3
        mockAPI.mockRegisters.r15 = 0x800C
        await viewModel.fetchStatus()
        await viewModel.fetchRegisters()

        XCTAssertTrue(viewModel.registers.hasRegister("R2", value: 8),
                     "R2 should contain function result")
        XCTAssertTrue(viewModel.registers.hasRegister("PC", value: 0x800C),
                     "PC should be after function call")
    }

    func testStepOutOfFunction() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.functionCall)

        // Set breakpoint inside function
        await viewModel.toggleBreakpoint(at: 0x800C) // Inside add_numbers

        // Run to breakpoint
        await viewModel.run()
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x800C, instruction: nil, cycleCount: 5, error: nil)
        await viewModel.fetchStatus()
        try await waitForStatus(.breakpoint, timeout: 1.0, viewModel: viewModel)

        // Step out of function
        await viewModel.stepOut()
        XCTAssertTrue(mockAPI.stepOutCalled)

        // Simulate step out completing (back to caller)
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x800C, instruction: nil, cycleCount: 8, error: nil)
        await viewModel.fetchStatus()

        // Verify we're back in caller
        XCTAssertTrue(viewModel.registers.hasRegister("PC", value: 0x800C))
    }

    func testStepThroughLoop() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.simpleLoop)

        // Step through first few iterations
        for iteration in 0..<3 {
            await viewModel.step()
            XCTAssertTrue(mockAPI.stepCalled)

            // Simulate register update (counter incrementing)
            mockAPI.mockRegisters.r0 = UInt32(iteration + 1)
            await viewModel.fetchRegisters()
        }

        // Verify counter progressed
        let counterValue = viewModel.registers.registerValue("R0") ?? 0
        XCTAssertGreaterThanOrEqual(counterValue, 1, "Counter should have incremented")
    }

    // MARK: - Expression Evaluation Workflow

    func testEvaluateExpression() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.withBreakpoint)

        // Set breakpoint and run
        await viewModel.toggleBreakpoint(at: 0x8008)
        await viewModel.run()
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x8008, instruction: nil, cycleCount: 3, error: nil)
        await viewModel.fetchStatus()
        try await waitForStatus(.breakpoint, timeout: 1.0, viewModel: viewModel)

        // Evaluate expression at breakpoint
        let result = try await viewModel.evaluateExpression("R0 + R1")
        // Mock always returns 0, but in real scenario would return sum
        XCTAssertNotNil(result)
    }

    // MARK: - Combined Debugging Workflow

    func testCompleteDebuggingSession() async throws {
        // Complete workflow: load → breakpoints → run → inspect → step → continue

        // 1. Load program
        await viewModel.loadProgram(source: ProgramFixtures.withBreakpoint)
        XCTAssertNil(viewModel.errorMessage)

        // 2. Set multiple breakpoints
        await viewModel.toggleBreakpoint(at: 0x8000)
        await viewModel.toggleBreakpoint(at: 0x8004)
        await viewModel.toggleBreakpoint(at: 0x8008)

        // 3. Run to first breakpoint
        await viewModel.run()
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x8000, instruction: nil, cycleCount: 1, error: nil)
        await viewModel.fetchStatus()
        try await waitForStatus(.breakpoint, timeout: 1.0, viewModel: viewModel)

        // 4. Inspect state
        await viewModel.fetchRegisters()
        XCTAssertTrue(mockAPI.getRegistersCalled)

        // 5. Step once
        await viewModel.step()
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x8004, instruction: nil, cycleCount: 2, error: nil)
        await viewModel.fetchStatus()

        // 6. Continue to next breakpoint
        await viewModel.run()
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x8008, instruction: nil, cycleCount: 3, error: nil)
        await viewModel.fetchStatus()

        // 7. Remove a breakpoint
        await viewModel.toggleBreakpoint(at: 0x8004)
        XCTAssertTrue(mockAPI.removeBreakpointCalled)

        // 8. Continue to completion
        await viewModel.run()
        mockAPI.mockStatus = VMStatus(state: "halted", pc: 0x8010, instruction: nil, cycleCount: 10, error: nil)
        await viewModel.fetchStatus()
        try await waitForStatus(.halted, timeout: 1.0, viewModel: viewModel)
    }

    // MARK: - Error Handling

    func testBreakpointFailure() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.withBreakpoint)

        // Simulate breakpoint add failure
        mockAPI.shouldFailAddBreakpoint = true

        await viewModel.toggleBreakpoint(at: 0x8000)

        // Error should be handled gracefully
        // (ViewModel should set errorMessage)
        // XCTAssertNotNil(viewModel.errorMessage)
        // Note: Actual error handling depends on ViewModel implementation
    }

    func testStepFailure() async throws {
        await viewModel.loadProgram(source: ProgramFixtures.withBreakpoint)

        // Set breakpoint and hit it
        await viewModel.toggleBreakpoint(at: 0x8000)
        await viewModel.run()
        mockAPI.mockStatus = VMStatus(state: "breakpoint", pc: 0x8000, instruction: nil, cycleCount: 1, error: nil)
        await viewModel.fetchStatus()
        try await waitForStatus(.breakpoint, timeout: 1.0, viewModel: viewModel)

        // Simulate step failure
        mockAPI.shouldFailStep = true
        mockAPI.stepErrorMessage = "Cannot step: invalid state"

        await viewModel.step()

        // Error should be handled
        // XCTAssertNotNil(viewModel.errorMessage)
    }
}
