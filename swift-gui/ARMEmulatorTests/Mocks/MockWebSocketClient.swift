import Combine
import Foundation
@testable import ARMEmulator

/// Mock implementation of WebSocketClient for testing
/// Implements WebSocketClientProtocol to enable proper dependency injection
final class MockWebSocketClient: WebSocketClientProtocol, @unchecked Sendable {
    private let eventSubject = PassthroughSubject<EmulatorEvent, Never>()

    var events: AnyPublisher<EmulatorEvent, Never> {
        eventSubject.eraseToAnyPublisher()
    }

    func connect(sessionID: String) {
        // No-op
    }

    func disconnect() {
        // No-op
    }

    // Helper method for tests to emit events
    func emitEvent(_ event: EmulatorEvent) {
        eventSubject.send(event)
    }
}
