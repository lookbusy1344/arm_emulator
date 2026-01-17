import Combine
import Foundation
import XCTest
@testable import ARMEmulator

// MARK: - Mock Clients

final class MockAPIClient: APIClient, @unchecked Sendable {
    override func createSession() async throws -> String {
        return "mock-session-id"
    }

    override func destroySession(sessionID: String) async throws {
        // No-op
    }
}

final class MockWebSocketClient: WebSocketClient, @unchecked Sendable {
    override func connect(sessionID: String) {
        // No-op
    }

    override func disconnect() {
        // No-op
    }
}

// MARK: - Highlight Tests

@MainActor
final class HighlightTests: XCTestCase {
    var viewModel: EmulatorViewModel!

    override func setUp() async throws {
        viewModel = EmulatorViewModel(
            apiClient: MockAPIClient(),
            wsClient: MockWebSocketClient()
        )
    }

    func testRegisterHighlightAdded() {
        viewModel.highlightRegister("R0")
        XCTAssertNotNil(viewModel.registerHighlights["R0"])
    }

    func testMemoryHighlightAdded() {
        viewModel.highlightMemoryAddress(0x8000, size: 1)
        XCTAssertNotNil(viewModel.memoryHighlights[0x8000])
    }
}
