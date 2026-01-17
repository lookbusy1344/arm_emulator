import Foundation
@testable import ARMEmulator

/// Mock implementation of WebSocketClient for testing
/// Currently subclasses concrete WebSocketClient class - should implement protocol once extracted
final class MockWebSocketClient: WebSocketClient, @unchecked Sendable {
    override func connect(sessionID: String) {
        // No-op
    }

    override func disconnect() {
        // No-op
    }
}
