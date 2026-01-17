import Foundation
@testable import ARMEmulator

/// Mock implementation of APIClient for testing
/// Implements APIClientProtocol to enable proper dependency injection
final class MockAPIClient: APIClientProtocol, @unchecked Sendable {
    func createSession() async throws -> String {
        return "mock-session-id"
    }

    func destroySession(sessionID: String) async throws {
        // No-op
    }

    func getStatus(sessionID: String) async throws -> VMStatus {
        return VMStatus(state: "idle", pc: 0, instruction: nil, cycleCount: nil, error: nil)
    }

    func loadProgram(sessionID: String, source: String) async throws -> LoadProgramResponse {
        return LoadProgramResponse(success: true, errors: nil, symbols: nil)
    }

    func run(sessionID: String) async throws {
        // No-op
    }

    func stop(sessionID: String) async throws {
        // No-op
    }

    func step(sessionID: String) async throws {
        // No-op
    }

    func stepOver(sessionID: String) async throws {
        // No-op
    }

    func stepOut(sessionID: String) async throws {
        // No-op
    }

    func reset(sessionID: String) async throws {
        // No-op
    }

    func restart(sessionID: String) async throws {
        // No-op
    }

    func sendStdin(sessionID: String, data: String) async throws {
        // No-op
    }

    func getRegisters(sessionID: String) async throws -> RegisterState {
        return .empty
    }

    func addBreakpoint(sessionID: String, address: UInt32) async throws {
        // No-op
    }

    func removeBreakpoint(sessionID: String, address: UInt32) async throws {
        // No-op
    }

    func getSourceMap(sessionID: String) async throws -> [SourceMapEntry] {
        return []
    }

    func getBreakpoints(sessionID: String) async throws -> [UInt32] {
        return []
    }

    func evaluateExpression(sessionID: String, expression: String) async throws -> UInt32 {
        return 0
    }

    func addWatchpoint(sessionID: String, address: UInt32, type: String) async throws -> Watchpoint {
        return Watchpoint(id: 1, address: address, type: type)
    }

    func removeWatchpoint(sessionID: String, watchpointID: Int) async throws {
        // No-op
    }

    func getWatchpoints(sessionID: String) async throws -> [Watchpoint] {
        return []
    }

    func getMemory(sessionID: String, address: UInt32, length: Int) async throws -> [UInt8] {
        return Array(repeating: 0, count: length)
    }

    func getDisassembly(sessionID: String, address: UInt32, count: Int) async throws -> [DisassemblyInstruction] {
        return []
    }

    func getVersion() async throws -> BackendVersion {
        return BackendVersion(version: "1.0.0", commit: "mock", date: "2026-01-17")
    }
}
