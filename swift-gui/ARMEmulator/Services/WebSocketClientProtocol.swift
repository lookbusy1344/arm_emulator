import Combine
import Foundation

/// Protocol for WebSocket client to enable dependency injection and testing
protocol WebSocketClientProtocol: ObservableObject, Sendable {
    /// Publisher for emulator events from the WebSocket connection
    var events: AnyPublisher<EmulatorEvent, Never> { get }

    /// Connect to the WebSocket server for the given session
    func connect(sessionID: String)

    /// Disconnect from the WebSocket server
    func disconnect()
}
