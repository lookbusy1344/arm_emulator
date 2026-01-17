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

    func testRegisterHighlightFadesAfterDelay() async throws {
        viewModel.highlightRegister("R0")

        // Should be highlighted immediately
        XCTAssertNotNil(viewModel.registerHighlights["R0"])

        // Wait for fade to complete
        try await Task.sleep(nanoseconds: 1_600_000_000)  // 1.6s

        // Should be removed
        XCTAssertNil(viewModel.registerHighlights["R0"])
    }

    func testRapidChangesRestartTimer() async throws {
        viewModel.highlightRegister("R0")

        // Wait halfway through fade
        try await Task.sleep(nanoseconds: 500_000_000)  // 0.5s

        // Trigger another change (should restart timer)
        viewModel.highlightRegister("R0")

        // Wait 1.2s (0.7s after restart)
        try await Task.sleep(nanoseconds: 1_200_000_000)

        // Should still be highlighted
        XCTAssertNotNil(viewModel.registerHighlights["R0"])

        // Wait another 0.5s (1.2s after restart, past 1.5s threshold)
        try await Task.sleep(nanoseconds: 500_000_000)

        // Should be removed now
        XCTAssertNil(viewModel.registerHighlights["R0"])
    }

    func testMultipleRegisterHighlightsIndependent() async throws {
        viewModel.highlightRegister("R0")

        try await Task.sleep(nanoseconds: 500_000_000)  // 0.5s

        viewModel.highlightRegister("R1")

        // Both should be highlighted
        XCTAssertNotNil(viewModel.registerHighlights["R0"])
        XCTAssertNotNil(viewModel.registerHighlights["R1"])

        // Wait for R0 to fade (1.2s more = 1.7s total)
        try await Task.sleep(nanoseconds: 1_200_000_000)

        // R0 should be gone, R1 still visible
        XCTAssertNil(viewModel.registerHighlights["R0"])
        XCTAssertNotNil(viewModel.registerHighlights["R1"])
    }

    func testMemoryHighlightMultipleBytes() {
        viewModel.highlightMemoryAddress(0x8000, size: 4)

        // All 4 bytes should be highlighted
        XCTAssertNotNil(viewModel.memoryHighlights[0x8000])
        XCTAssertNotNil(viewModel.memoryHighlights[0x8001])
        XCTAssertNotNil(viewModel.memoryHighlights[0x8002])
        XCTAssertNotNil(viewModel.memoryHighlights[0x8003])
        XCTAssertNil(viewModel.memoryHighlights[0x8004])  // 5th byte not written
    }
}
