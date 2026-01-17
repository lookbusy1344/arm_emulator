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
        try await Task.sleep(nanoseconds: 1_600_000_000) // 1.6s

        // Should be removed
        XCTAssertNil(viewModel.registerHighlights["R0"])
    }

    func testRapidChangesRestartTimer() async throws {
        viewModel.highlightRegister("R0")

        // Wait halfway through fade
        try await Task.sleep(nanoseconds: 500_000_000) // 0.5s

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

        try await Task.sleep(nanoseconds: 500_000_000) // 0.5s

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
        XCTAssertNil(viewModel.memoryHighlights[0x8004]) // 5th byte not written
    }

    func testUpdateRegistersTriggersHighlights() async throws {
        // Simulate first state
        let registers1 = RegisterState(
            r0: 0, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
            sp: 0x50000, lr: 0, pc: 0x8000,
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false)
        )
        viewModel.updateRegisters(registers1)

        // Simulate second state with R0, R1 changed
        let registers2 = RegisterState(
            r0: 42, r1: 100, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
            sp: 0x50000, lr: 0, pc: 0x8004,
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false)
        )
        viewModel.updateRegisters(registers2)

        // R0 and R1 should be highlighted, PC should be highlighted
        XCTAssertNotNil(viewModel.registerHighlights["R0"])
        XCTAssertNotNil(viewModel.registerHighlights["R1"])
        XCTAssertNotNil(viewModel.registerHighlights["PC"])
        XCTAssertNil(viewModel.registerHighlights["R2"]) // Unchanged
    }
}
