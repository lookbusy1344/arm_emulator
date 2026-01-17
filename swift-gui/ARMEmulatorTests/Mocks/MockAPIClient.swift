import Foundation
@testable import ARMEmulator

/// Mock implementation of APIClient for testing
/// Currently subclasses concrete APIClient class - should implement protocol once extracted
final class MockAPIClient: APIClient, @unchecked Sendable {
    override func createSession() async throws -> String {
        return "mock-session-id"
    }

    override func destroySession(sessionID: String) async throws {
        // No-op
    }
}
