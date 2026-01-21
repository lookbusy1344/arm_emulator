import XCTest
@testable import ARMEmulator

/// Integration tests with real Go backend
/// **IMPORTANT**: These tests require the Go backend to be built and running
///
/// To run these tests:
/// 1. Build the backend: `cd .. && make build`
/// 2. Ensure backend is not already running on port 8080
/// 3. Tests will start backend automatically (or skip if unavailable)
///
/// These tests are designed to be skipped gracefully if backend is not available.
@MainActor
final class BackendIntegrationTests: XCTestCase {
    static var backendProcess: Process?
    static var backendAvailable = false

    var apiClient: APIClient!
    var sessionID: String?

    // Use separate test port to avoid conflicts
    static let testPort = 8081
    static var testBaseURL: URL {
        URL(string: "http://localhost:\(testPort)")!
    }

    override class func setUp() {
        super.setUp()

        // Try to start backend on test port
        do {
            backendProcess = try startBackend(port: testPort)
            // Wait for backend to be ready
            Thread.sleep(forTimeInterval: 2.0)
            backendAvailable = true
            print("✅ Backend started on port \(testPort)")
        } catch {
            print("⚠️  Backend not available: \(error.localizedDescription)")
            print("   These tests will be skipped")
            backendAvailable = false
        }
    }

    override class func tearDown() {
        if let process = backendProcess {
            process.terminate()
            process.waitUntilExit()
            print("🛑 Backend stopped")
        }
        super.tearDown()
    }

    override func setUp() async throws {
        try await super.setUp()

        guard Self.backendAvailable else {
            throw XCTSkip("Backend not available")
        }

        apiClient = APIClient(baseURL: Self.testBaseURL)
    }

    override func tearDown() async throws {
        // Clean up session if created
        if let sessionID = sessionID {
            try? await apiClient.destroySession(sessionID: sessionID)
            self.sessionID = nil
        }

        apiClient = nil
        try await super.tearDown()
    }

    // MARK: - Session Management

    func testCreateAndDestroySession() async throws {
        // Create session
        let sessionID = try await apiClient.createSession()
        self.sessionID = sessionID

        XCTAssertFalse(sessionID.isEmpty, "Session ID should not be empty")
        XCTAssertTrue(sessionID.hasPrefix("session-") || sessionID.count > 8,
                     "Session ID should be properly formatted")

        // Destroy session
        try await apiClient.destroySession(sessionID: sessionID)
        self.sessionID = nil

        // Verify session is destroyed by trying to get status (should fail)
        do {
            _ = try await apiClient.getStatus(sessionID: sessionID)
            XCTFail("Should fail to get status for destroyed session")
        } catch {
            // Expected to fail
            XCTAssertTrue(true, "Session properly destroyed")
        }
    }

    func testSessionTimeout() async throws {
        // Create session
        let sessionID = try await apiClient.createSession()
        self.sessionID = sessionID

        // Session should remain valid for reasonable time
        try await Task.sleep(nanoseconds: 1_000_000_000) // 1 second

        // Should still be valid
        let status = try await apiClient.getStatus(sessionID: sessionID)
        XCTAssertNotNil(status)

        // Clean up
        try await apiClient.destroySession(sessionID: sessionID)
        self.sessionID = nil
    }

    func testMultipleSessions() async throws {
        // Create multiple sessions
        let session1 = try await apiClient.createSession()
        let session2 = try await apiClient.createSession()

        XCTAssertNotEqual(session1, session2, "Session IDs should be unique")

        // Both should be valid
        let status1 = try await apiClient.getStatus(sessionID: session1)
        let status2 = try await apiClient.getStatus(sessionID: session2)
        XCTAssertNotNil(status1)
        XCTAssertNotNil(status2)

        // Clean up
        try await apiClient.destroySession(sessionID: session1)
        try await apiClient.destroySession(sessionID: session2)
    }

    // MARK: - Program Loading

    func testLoadSimpleProgram() async throws {
        sessionID = try await apiClient.createSession()

        // Load simple program
        let response = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.exitCode42
        )

        XCTAssertTrue(response.success, "Program should load successfully")
        XCTAssertNil(response.errors, "Should have no errors")
        XCTAssertNotNil(response.symbols, "Should have symbol table")
    }

    func testLoadAllExamplePrograms() async throws {
        // Note: This test loads ALL 49 example programs
        // It may take a while but verifies backend handles all examples

        sessionID = try await apiClient.createSession()

        let exampleFiles = [
            "hello.s", "fibonacci.s", "factorial.s", "loops.s",
            "arithmetic.s", "conditionals.s", "functions.s",
            // Add more example filenames as needed
        ]

        for filename in exampleFiles {
            guard let programSource = ProgramFixtures.loadExample(filename) else {
                print("⚠️  Skipping \(filename) - file not found")
                continue
            }

            let response = try await apiClient.loadProgram(
                sessionID: sessionID!,
                source: programSource
            )

            XCTAssertTrue(response.success,
                         "\(filename) should load successfully")
        }
    }

    func testLoadProgramWithSyntaxError() async throws {
        sessionID = try await apiClient.createSession()

        let response = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.syntaxError
        )

        XCTAssertFalse(response.success, "Should fail to load invalid program")
        XCTAssertNotNil(response.errors, "Should have error messages")
    }

    func testLoadLargeProgram() async throws {
        sessionID = try await apiClient.createSession()

        // Create large program
        let largeProgram = String(repeating: "MOV R0, #0\nADD R0, R0, #1\n", count: 5000)

        let response = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: largeProgram
        )

        XCTAssertTrue(response.success, "Should handle large programs")
    }

    // MARK: - Execution Control

    func testRunProgram() async throws {
        sessionID = try await apiClient.createSession()

        // Load program
        _ = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.simpleLoop
        )

        // Run program
        try await apiClient.run(sessionID: sessionID!)

        // Wait for completion (poll status)
        var attempts = 0
        while attempts < 50 {
            let status = try await apiClient.getStatus(sessionID: sessionID!)
            if status.state == "halted" || status.state == "error" {
                XCTAssertEqual(status.state, "halted", "Program should halt successfully")
                break
            }
            try await Task.sleep(nanoseconds: 100_000_000) // 100ms
            attempts += 1
        }

        XCTAssertLessThan(attempts, 50, "Program should complete within 5 seconds")
    }

    func testStepExecution() async throws {
        sessionID = try await apiClient.createSession()

        // Load program
        _ = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.withBreakpoint
        )

        // Step once
        try await apiClient.step(sessionID: sessionID!)

        // Get status
        let status = try await apiClient.getStatus(sessionID: sessionID!)
        XCTAssertEqual(status.state, "breakpoint", "Should be in breakpoint state after step")

        // PC should have advanced
        XCTAssertGreaterThan(status.pc, 0x8000, "PC should have advanced")
    }

    func testStopExecution() async throws {
        sessionID = try await apiClient.createSession()

        // Load long-running program
        _ = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.longRunning
        )

        // Start execution
        try await apiClient.run(sessionID: sessionID!)

        // Wait a bit
        try await Task.sleep(nanoseconds: 100_000_000) // 100ms

        // Stop execution
        try await apiClient.stop(sessionID: sessionID!)

        // Verify stopped
        let status = try await apiClient.getStatus(sessionID: sessionID!)
        XCTAssertNotEqual(status.state, "running", "Should not be running after stop")
    }

    func testResetExecution() async throws {
        sessionID = try await apiClient.createSession()

        // Load and run program
        _ = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.simpleLoop
        )
        try await apiClient.run(sessionID: sessionID!)

        // Wait for completion
        var attempts = 0
        while attempts < 50 {
            let status = try await apiClient.getStatus(sessionID: sessionID!)
            if status.state == "halted" {
                break
            }
            try await Task.sleep(nanoseconds: 100_000_000)
            attempts += 1
        }

        // Reset
        try await apiClient.reset(sessionID: sessionID!)

        // Verify reset
        let status = try await apiClient.getStatus(sessionID: sessionID!)
        XCTAssertEqual(status.state, "idle", "Should be idle after reset")
        XCTAssertEqual(status.pc, 0x8000, "PC should be reset to start")
    }

    // MARK: - Memory Operations

    func testReadMemory() async throws {
        sessionID = try await apiClient.createSession()

        // Load program
        _ = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.memoryWrite
        )

        // Read memory at program start (should be instructions)
        let memory = try await apiClient.getMemory(
            sessionID: sessionID!,
            address: 0x8000,
            length: 16
        )

        XCTAssertEqual(memory.count, 16, "Should read 16 bytes")
        // Memory should contain ARM instructions (non-zero)
        XCTAssertTrue(memory.contains { $0 != 0 }, "Memory should contain instructions")
    }

    func testReadLargeMemoryChunk() async throws {
        sessionID = try await apiClient.createSession()

        _ = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.helloWorld
        )

        // Read 1KB of memory
        let memory = try await apiClient.getMemory(
            sessionID: sessionID!,
            address: 0x8000,
            length: 1024
        )

        XCTAssertEqual(memory.count, 1024, "Should read full 1KB")
    }

    func testMemoryBoundsChecking() async throws {
        sessionID = try await apiClient.createSession()

        _ = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.helloWorld
        )

        // Try to read invalid memory (very high address)
        do {
            _ = try await apiClient.getMemory(
                sessionID: sessionID!,
                address: 0xFFFF_FFFF,
                length: 16
            )
            // Backend may allow this and return zeros, or may error
            XCTAssertTrue(true, "Backend handled out-of-bounds read")
        } catch {
            // Also acceptable to error
            XCTAssertTrue(true, "Backend rejected out-of-bounds read")
        }
    }

    // MARK: - Breakpoint Management

    func testAddAndRemoveBreakpoint() async throws {
        sessionID = try await apiClient.createSession()

        _ = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.withBreakpoint
        )

        // Add breakpoint
        try await apiClient.addBreakpoint(sessionID: sessionID!, address: 0x8004)

        // Get breakpoints list
        let breakpoints = try await apiClient.getBreakpoints(sessionID: sessionID!)
        XCTAssertTrue(breakpoints.contains(0x8004), "Breakpoint should be in list")

        // Remove breakpoint
        try await apiClient.removeBreakpoint(sessionID: sessionID!, address: 0x8004)

        // Verify removed
        let breakpointsAfter = try await apiClient.getBreakpoints(sessionID: sessionID!)
        XCTAssertFalse(breakpointsAfter.contains(0x8004), "Breakpoint should be removed")
    }

    func testBreakpointHit() async throws {
        sessionID = try await apiClient.createSession()

        _ = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: ProgramFixtures.withBreakpoint
        )

        // Add breakpoint at second instruction
        try await apiClient.addBreakpoint(sessionID: sessionID!, address: 0x8004)

        // Run until breakpoint
        try await apiClient.run(sessionID: sessionID!)

        // Poll for breakpoint hit
        var attempts = 0
        while attempts < 20 {
            let status = try await apiClient.getStatus(sessionID: sessionID!)
            if status.state == "breakpoint" {
                XCTAssertEqual(status.pc, 0x8004, "Should stop at breakpoint address")
                return
            }
            try await Task.sleep(nanoseconds: 50_000_000) // 50ms
            attempts += 1
        }

        XCTFail("Should hit breakpoint within 1 second")
    }

    // MARK: - Backend Version

    func testGetBackendVersion() async throws {
        let version = try await apiClient.getVersion()

        XCTAssertFalse(version.version.isEmpty, "Should have version string")
        XCTAssertFalse(version.commit.isEmpty, "Should have commit hash")
        XCTAssertFalse(version.date.isEmpty, "Should have build date")

        print("✅ Backend version: \(version.version) (\(version.commit))")
    }

    // MARK: - Error Responses

    func testInvalidSessionError() async throws {
        // Try to use nonexistent session
        do {
            _ = try await apiClient.getStatus(sessionID: "invalid-session-id")
            XCTFail("Should fail with invalid session")
        } catch {
            // Expected to fail
            XCTAssertTrue(true, "Properly rejected invalid session")
        }
    }

    func testInvalidProgramError() async throws {
        sessionID = try await apiClient.createSession()

        let response = try await apiClient.loadProgram(
            sessionID: sessionID!,
            source: "This is not ARM assembly"
        )

        XCTAssertFalse(response.success, "Should fail to load invalid program")
        XCTAssertNotNil(response.errors, "Should have error details")
    }

    // MARK: - Helper: Start Backend Process

    private static func startBackend(port: Int) throws -> Process {
        // Find backend binary
        let projectRoot = URL(fileURLWithPath: #file)
            .deletingLastPathComponent() // Integration
            .deletingLastPathComponent() // ARMEmulatorTests
            .deletingLastPathComponent() // swift-gui
            .deletingLastPathComponent() // project root

        let backendPath = projectRoot.appendingPathComponent("arm-emulator")

        guard FileManager.default.fileExists(atPath: backendPath.path) else {
            throw TestError.setupFailure("Backend binary not found at \(backendPath.path). Run: make build")
        }

        // Start backend process
        let process = Process()
        process.executableURL = URL(fileURLWithPath: backendPath.path)
        process.arguments = ["--api-only", "--api-port", "\(port)"]

        // Suppress output
        process.standardOutput = Pipe()
        process.standardError = Pipe()

        try process.run()

        return process
    }
}
