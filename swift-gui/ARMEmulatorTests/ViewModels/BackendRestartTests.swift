import XCTest
@testable import ARMEmulator

@MainActor
final class BackendRestartTests: XCTestCase {
    private var viewModel: EmulatorViewModel!
    private var mockAPIClient: MockAPIClient!
    private var mockWSClient: MockWebSocketClient!

    private let program = "_start:\n  MOV R0, #1\n  MOV R1, #2\n  SWI #0\n"

    override func setUp() async throws {
        mockAPIClient = MockAPIClient()
        mockAPIClient.generateUniqueSessionIDs = true
        mockWSClient = MockWebSocketClient()
        viewModel = EmulatorViewModel(apiClient: mockAPIClient, wsClient: mockWSClient)
        await viewModel.initialize()
    }

    override func tearDown() async throws {
        viewModel = nil
        mockAPIClient = nil
        mockWSClient = nil
    }

    private func entry(_ address: UInt32, _ line: Int) -> SourceMapEntry {
        SourceMapEntry(address: address, lineNumber: line, line: "")
    }

    private func openProgram() async {
        mockAPIClient.mockSourceMap = [entry(0x8000, 2), entry(0x8004, 3), entry(0x8008, 4)]
        await viewModel.openProgram(source: program)
    }

    // MARK: - Session

    func testRestartOpensANewSessionAndConnectsToIt() async {
        await viewModel.restartBackend { true }

        XCTAssertEqual(viewModel.sessionID, "mock-session-id-2")
        XCTAssertEqual(mockWSClient.currentSessionID, "mock-session-id-2")
        XCTAssertTrue(viewModel.isConnected)
    }

    func testRestartDoesNotDestroyTheOldSessionOnTheDeadBackend() async {
        await viewModel.restartBackend { true }

        XCTAssertFalse(mockAPIClient.destroySessionCalled)
    }

    func testRestartDisconnectsTheWebSocketBeforeRestarting() async {
        var connectedDuringRestart: Bool?

        await viewModel.restartBackend { [mockWSClient] in
            connectedDuringRestart = mockWSClient?.isConnected
            return true
        }

        XCTAssertEqual(connectedDuringRestart, false)
    }

    func testFailedRestartLeavesNoSession() async {
        await viewModel.restartBackend { false }

        XCTAssertNil(viewModel.sessionID)
        XCTAssertFalse(viewModel.isConnected)
        XCTAssertEqual(mockAPIClient.createSessionCallCount, 1)
    }

    func testRestartWhenSessionCreationFailsReportsTheErrorAndLoadsNothing() async {
        await openProgram()
        mockAPIClient.shouldFailCreateSession = true

        await viewModel.restartBackend { true }

        XCTAssertEqual(viewModel.errorMessage, "Failed to initialize: Mock session creation failed")
        XCTAssertFalse(viewModel.isConnected)
        XCTAssertEqual(mockAPIClient.loadProgramCallCount, 1)
    }

    // MARK: - Program reload

    func testRestartReloadsTheAssembledProgramIntoTheNewSession() async {
        await openProgram()

        await viewModel.restartBackend { true }

        XCTAssertEqual(mockAPIClient.loadedSessionIDs, ["mock-session-id-1", "mock-session-id-2"])
        XCTAssertEqual(mockAPIClient.lastLoadedSource, program)
        XCTAssertFalse(viewModel.needsAssembly)
    }

    func testRestartWithNoProgramAssembledLoadsNothing() async {
        viewModel.sourceCode = program

        await viewModel.restartBackend { true }

        XCTAssertEqual(mockAPIClient.loadProgramCallCount, 0)
        XCTAssertEqual(viewModel.sourceCode, program)
    }

    func testRestartKeepsBreakpointsOnTheirLines() async {
        await openProgram()
        await viewModel.toggleBreakpoint(at: 0x8008)
        mockAPIClient.addedBreakpoints = []

        await viewModel.restartBackend { true }

        XCTAssertEqual(mockAPIClient.removedBreakpoints, [])
        XCTAssertEqual(mockAPIClient.addedBreakpoints, [0x8008])
        XCTAssertEqual(viewModel.breakpoints, [0x8008])
    }

    func testRestartClearsWatchpointsOfTheOldSession() async {
        await openProgram()
        await viewModel.addWatchpoint(at: 0x9000, type: "write")

        await viewModel.restartBackend { true }

        XCTAssertEqual(viewModel.watchpoints, [])
    }

    // MARK: - Concurrent initialization

    func testInitializeDuringAnotherInitializeWaitsForTheSession() async {
        viewModel.cleanup()
        let (release, releaser) = AsyncStream<Void>.makeStream()
        mockAPIClient.createSessionGate = {
            for await _ in release {
                return
            }
        }
        let first = Task { await viewModel.initialize() }
        while mockAPIClient.createSessionCallCount < 2 {
            await Task.yield()
        }
        // Runs once the second initialize below suspends
        Task {
            releaser.yield()
            releaser.finish()
        }

        await viewModel.initialize()

        XCTAssertTrue(viewModel.isConnected)
        XCTAssertEqual(mockAPIClient.createSessionCallCount, 2)
        await first.value
    }
}
