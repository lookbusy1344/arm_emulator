import Combine
import Foundation
@testable import ARMEmulator

/// Mock implementation of WebSocketClient for testing
/// Implements WebSocketClientProtocol to enable proper dependency injection
final class MockWebSocketClient: WebSocketClientProtocol, @unchecked Sendable {
    private let eventSubject = PassthroughSubject<EmulatorEvent, Never>()

    // Connection state tracking
    var isConnected = false
    var currentSessionID: String?

    var events: AnyPublisher<EmulatorEvent, Never> {
        eventSubject.eraseToAnyPublisher()
    }

    func connect(sessionID: String) {
        isConnected = true
        currentSessionID = sessionID
    }

    func disconnect() {
        isConnected = false
        currentSessionID = nil
    }

    // MARK: - Test Helper Methods

    /// Helper method for tests to emit events
    func emitEvent(_ event: EmulatorEvent) {
        eventSubject.send(event)
    }

    /// Simulate a WebSocket disconnect (connection loss)
    func simulateDisconnect() {
        isConnected = false
        // Don't clear currentSessionID - simulates temporary connection loss
    }

    /// Simulate a WebSocket reconnection
    func simulateReconnect(sessionID: String) {
        isConnected = true
        currentSessionID = sessionID
    }

    /// Simulate receiving an event (alias for emitEvent for clarity in tests)
    func simulateEvent(_ event: EmulatorEvent) {
        eventSubject.send(event)
    }
}
