import Foundation
@testable import ARMEmulator

/// Mock implementation of APIClient for testing
/// Implements APIClientProtocol to enable proper dependency injection
final class MockAPIClient: APIClientProtocol, @unchecked Sendable {
    // Call tracking
    var createSessionCalled = false
    var destroySessionCalled = false
    var loadProgramCalled = false
    var lastLoadedSource: String?
    var runCalled = false
    var runCallCount = 0 // Track multiple run calls
    var stopCalled = false
    var stepCalled = false
    var stepOverCalled = false
    var stepOutCalled = false
    var resetCalled = false
    var restartCalled = false
    var getRegistersCalled = false
    var getStatusCalled = false
    var sendStdinCalled = false
    var lastStdinData: String?
    var addBreakpointCalled = false
    var addBreakpointCallCount = 0 // Track multiple breakpoint additions
    var removeBreakpointCalled = false
    var lastBreakpointAddress: UInt32?
    var addWatchpointCalled = false
    var removeWatchpointCalled = false
    var getWatchpointsCalled = false

    // Error simulation
    var shouldFailCreateSession = false
    var shouldFailLoadProgram = false
    var shouldFailRun = false
    var runErrorMessage: String? // Custom error message for run failures
    var shouldFailStep = false
    var stepErrorMessage: String?
    var shouldFailAddBreakpoint = false
    var shouldFailRemoveBreakpoint = false
    var shouldFailGetMemory = false

    // Response customization
    var mockSessionID = "mock-session-id"
    var mockLoadProgramResponse = LoadProgramResponse(success: true, errors: nil, symbols: ["main": 0x8000])
    var mockRegisters = RegisterState.empty
    var mockStatus = VMStatus(state: "idle", pc: 0x8000, instruction: nil, cycleCount: nil, error: nil)
    var mockMemoryData: [UInt8]? // Custom memory data for tests

    /// Performance simulation
    var simulateDelay: TimeInterval = 0 // Simulate slow API responses

    /// Stdin simulation (for interactive program testing)
    var mockedStdinInput: String? // Simulate stdin input for interactive programs

    // Session ID generation (for session restart testing)
    var generateUniqueSessionIDs = false // Set to true to generate unique IDs per call
    private var sessionIDCounter = 0

    func createSession() async throws -> String {
        if simulateDelay > 0 {
            try await Task.sleep(nanoseconds: UInt64(simulateDelay * 1_000_000_000))
        }
        createSessionCalled = true
        if shouldFailCreateSession {
            throw NSError(
                domain: "MockAPIClient",
                code: -1,
                userInfo: [NSLocalizedDescriptionKey: "Mock session creation failed"],
            )
        }
        // Generate unique session ID for each call if requested (for session restart testing)
        if generateUniqueSessionIDs {
            sessionIDCounter += 1
            return "mock-session-id-\(sessionIDCounter)"
        }
        return mockSessionID
    }

    func destroySession(sessionID: String) async throws {
        destroySessionCalled = true
    }

    func getStatus(sessionID: String) async throws -> VMStatus {
        getStatusCalled = true
        return mockStatus
    }

    func loadProgram(sessionID: String, source: String) async throws -> LoadProgramResponse {
        if simulateDelay > 0 {
            try await Task.sleep(nanoseconds: UInt64(simulateDelay * 1_000_000_000))
        }
        loadProgramCalled = true
        lastLoadedSource = source
        if shouldFailLoadProgram {
            throw NSError(
                domain: "MockAPIClient",
                code: -1,
                userInfo: [NSLocalizedDescriptionKey: "Mock load program failed"],
            )
        }
        return mockLoadProgramResponse
    }

    func run(sessionID: String) async throws {
        if simulateDelay > 0 {
            try await Task.sleep(nanoseconds: UInt64(simulateDelay * 1_000_000_000))
        }
        runCalled = true
        runCallCount += 1
        if shouldFailRun {
            let message = runErrorMessage ?? "Mock run failed"
            throw NSError(domain: "MockAPIClient", code: -1, userInfo: [NSLocalizedDescriptionKey: message])
        }
    }

    func stop(sessionID: String) async throws {
        stopCalled = true
    }

    func step(sessionID: String) async throws {
        stepCalled = true
        if shouldFailStep {
            throw NSError(
                domain: "MockAPIClient",
                code: -1,
                userInfo: [NSLocalizedDescriptionKey: stepErrorMessage ?? "Mock step failed"],
            )
        }
    }

    func stepOver(sessionID: String) async throws {
        stepOverCalled = true
        if shouldFailStep {
            throw NSError(
                domain: "MockAPIClient",
                code: -1,
                userInfo: [NSLocalizedDescriptionKey: stepErrorMessage ?? "Mock step over failed"],
            )
        }
    }

    func stepOut(sessionID: String) async throws {
        stepOutCalled = true
        if shouldFailStep {
            throw NSError(
                domain: "MockAPIClient",
                code: -1,
                userInfo: [NSLocalizedDescriptionKey: stepErrorMessage ?? "Mock step out failed"],
            )
        }
    }

    func reset(sessionID: String) async throws {
        resetCalled = true
    }

    func restart(sessionID: String) async throws {
        restartCalled = true
    }

    func sendStdin(sessionID: String, data: String) async throws {
        sendStdinCalled = true
        lastStdinData = data
    }

    func getRegisters(sessionID: String) async throws -> RegisterState {
        getRegistersCalled = true
        return mockRegisters
    }

    func addBreakpoint(sessionID: String, address: UInt32) async throws {
        addBreakpointCalled = true
        addBreakpointCallCount += 1
        lastBreakpointAddress = address
        if shouldFailAddBreakpoint {
            throw NSError(
                domain: "MockAPIClient",
                code: -1,
                userInfo: [NSLocalizedDescriptionKey: "Mock add breakpoint failed"],
            )
        }
    }

    func removeBreakpoint(sessionID: String, address: UInt32) async throws {
        removeBreakpointCalled = true
        lastBreakpointAddress = address
        if shouldFailRemoveBreakpoint {
            throw NSError(
                domain: "MockAPIClient",
                code: -1,
                userInfo: [NSLocalizedDescriptionKey: "Mock remove breakpoint failed"],
            )
        }
    }

    func getSourceMap(sessionID: String) async throws -> [SourceMapEntry] {
        []
    }

    func getBreakpoints(sessionID: String) async throws -> [UInt32] {
        []
    }

    func evaluateExpression(sessionID: String, expression: String) async throws -> UInt32 {
        0
    }

    func addWatchpoint(sessionID: String, address: UInt32, type: String) async throws -> Watchpoint {
        addWatchpointCalled = true
        return Watchpoint(id: 1, address: address, type: type)
    }

    func removeWatchpoint(sessionID: String, watchpointID: Int) async throws {
        removeWatchpointCalled = true
    }

    func getWatchpoints(sessionID: String) async throws -> [Watchpoint] {
        getWatchpointsCalled = true
        return []
    }

    func getMemory(sessionID: String, address: UInt32, length: Int) async throws -> [UInt8] {
        if shouldFailGetMemory {
            throw NSError(
                domain: "MockAPIClient",
                code: -1,
                userInfo: [NSLocalizedDescriptionKey: "Mock get memory failed"],
            )
        }
        if let customData = mockMemoryData {
            return customData
        }
        return Array(repeating: 0, count: length)
    }

    func getDisassembly(sessionID: String, address: UInt32, count: Int) async throws -> [DisassemblyInstruction] {
        []
    }

    func getVersion() async throws -> BackendVersion {
        BackendVersion(version: "1.0.0", commit: "mock", date: "2026-01-17")
    }
}
