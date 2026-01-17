import Foundation

/// Protocol for API client to enable dependency injection and testing
protocol APIClientProtocol: ObservableObject, Sendable {
    // MARK: - Session Management

    func createSession() async throws -> String
    func destroySession(sessionID: String) async throws
    func getStatus(sessionID: String) async throws -> VMStatus

    // MARK: - Program Management

    func loadProgram(sessionID: String, source: String) async throws -> LoadProgramResponse

    // MARK: - Execution Control

    func run(sessionID: String) async throws
    func stop(sessionID: String) async throws
    func step(sessionID: String) async throws
    func stepOver(sessionID: String) async throws
    func stepOut(sessionID: String) async throws
    func reset(sessionID: String) async throws
    func restart(sessionID: String) async throws
    func sendStdin(sessionID: String, data: String) async throws

    // MARK: - State Inspection

    func getRegisters(sessionID: String) async throws -> RegisterState

    // MARK: - Debugging

    func addBreakpoint(sessionID: String, address: UInt32) async throws
    func removeBreakpoint(sessionID: String, address: UInt32) async throws
    func getSourceMap(sessionID: String) async throws -> [SourceMapEntry]
    func getBreakpoints(sessionID: String) async throws -> [UInt32]
    func evaluateExpression(sessionID: String, expression: String) async throws -> UInt32

    // MARK: - Watchpoints

    func addWatchpoint(sessionID: String, address: UInt32, type: String) async throws -> Watchpoint
    func removeWatchpoint(sessionID: String, watchpointID: Int) async throws
    func getWatchpoints(sessionID: String) async throws -> [Watchpoint]

    // MARK: - Memory Operations

    func getMemory(sessionID: String, address: UInt32, length: Int) async throws -> [UInt8]
    func getDisassembly(sessionID: String, address: UInt32, count: Int) async throws -> [DisassemblyInstruction]

    // MARK: - Version

    func getVersion() async throws -> BackendVersion
}
