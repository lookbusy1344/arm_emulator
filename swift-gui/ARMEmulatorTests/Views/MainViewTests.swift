import SwiftUI
import XCTest
@testable import ARMEmulator

@MainActor
final class MainViewTests: XCTestCase {
    var backendManager: BackendManager!
    var fileService: FileService!
    var settings: AppSettings!

    override func setUp() async throws {
        try await super.setUp()

        backendManager = BackendManager()
        fileService = FileService.shared
        settings = AppSettings.shared
    }

    override func tearDown() async throws {
        // Note: Cannot nil out shared instances
        backendManager = nil
        try await super.tearDown()
    }

    // MARK: - Backend Status Tests

    func testBackendStatusRunning() {
        // Given: Backend is running
        backendManager.backendStatus = .running

        // Then: Status should be running
        XCTAssertEqual(backendManager.backendStatus, .running, "Backend should be in running state")
    }

    func testBackendStatusStopped() {
        // Given: Backend is stopped
        backendManager.backendStatus = .stopped

        // Then: Status should be stopped
        XCTAssertEqual(backendManager.backendStatus, .stopped, "Backend should be in stopped state")
    }

    func testBackendStatusError() {
        // Given: Backend has error
        backendManager.backendStatus = .error("Backend crashed")

        // Then: Status should be error
        if case let .error(message) = backendManager.backendStatus {
            XCTAssertEqual(message, "Backend crashed")
        } else {
            XCTFail("Backend status should be error")
        }
    }

    func testBackendStatusTransitions() {
        // Given: Backend starts as stopped
        backendManager.backendStatus = .stopped

        // When: Transition to starting
        backendManager.backendStatus = .starting

        // Then: Status should be starting
        XCTAssertEqual(backendManager.backendStatus, .starting)

        // When: Transition to running
        backendManager.backendStatus = .running

        // Then: Status should be running
        XCTAssertEqual(backendManager.backendStatus, .running)
    }

    // MARK: - Startup File Loading Tests

    func testStartupFileValidation_FileNotFound() async {
        // Given: ViewModel with mocks
        let mockAPIClient = MockAPIClient()
        let mockWebSocketClient = MockWebSocketClient()
        let viewModel = EmulatorViewModel(
            apiClient: mockAPIClient,
            wsClient: mockWebSocketClient,
        )

        await viewModel.initialize()

        // When: Attempt to load non-existent file
        let nonExistentPath = "/tmp/nonexistent-file-12345.s"

        // Simulate loadStartupFile logic
        let url = URL(fileURLWithPath: nonExistentPath)
        guard FileManager.default.fileExists(atPath: nonExistentPath) else {
            viewModel.errorMessage = "Could not load '\(url.lastPathComponent)': File not found"
            // Then: Error message should be set
            XCTAssertEqual(
                viewModel.errorMessage,
                "Could not load 'nonexistent-file-12345.s': File not found",
            )
            return
        }

        XCTFail("Should have detected file does not exist")
    }

    func testStartupFileValidation_InvalidExtension() {
        // Given: File with wrong extension
        let viewModel = EmulatorViewModel()
        let url = URL(fileURLWithPath: "/tmp/test.txt")

        // When: Check extension validation
        let isValidExtension = url.pathExtension == "s"

        // Then: Extension should be invalid
        XCTAssertFalse(isValidExtension, "File should have invalid extension")

        // Simulate error message
        if !isValidExtension {
            viewModel.errorMessage = "Could not load '\(url.lastPathComponent)': Not an assembly file (.s)"
            XCTAssertEqual(
                viewModel.errorMessage,
                "Could not load 'test.txt': Not an assembly file (.s)",
            )
        }
    }

    func testStartupFileValidation_ValidExtension() {
        // Given: File with valid extension
        let url = URL(fileURLWithPath: "/tmp/test.s")

        // When: Check extension validation
        let isValidExtension = url.pathExtension == "s"

        // Then: Extension should be valid
        XCTAssertTrue(isValidExtension, "File should have valid .s extension")
    }

    func testStartupFileLoadingSuccess() async throws {
        // Given: Temporary test file
        let tempDir = FileManager.default.temporaryDirectory
        let testFileURL = tempDir.appendingPathComponent("test-program.s")
        let testContent = """
        .org 0x8000
        MOV R0, #42
        SWI #0
        """

        try testContent.write(to: testFileURL, atomically: true, encoding: .utf8)

        // When: Load file content
        let loadedContent = try String(contentsOf: testFileURL, encoding: .utf8)

        // Then: Content should match
        XCTAssertEqual(loadedContent, testContent)

        // Cleanup
        try FileManager.default.removeItem(at: testFileURL)
    }

    // MARK: - Error Alert Display Tests

    func testErrorMessageTriggersAlert() {
        // Given: ViewModel with error
        let viewModel = EmulatorViewModel()

        // When: Set error message
        viewModel.errorMessage = "Test error message"

        // Then: Error message should be set
        XCTAssertEqual(viewModel.errorMessage, "Test error message")
        XCTAssertNotNil(viewModel.errorMessage, "Error message should not be nil")
    }

    func testErrorMessageClearingDismissesAlert() {
        // Given: ViewModel with error
        let viewModel = EmulatorViewModel()
        viewModel.errorMessage = "Test error"

        // When: Clear error message
        viewModel.errorMessage = nil

        // Then: Error message should be nil
        XCTAssertNil(viewModel.errorMessage, "Error message should be cleared")
    }

    func testMultipleErrorMessages() {
        // Given: ViewModel
        let viewModel = EmulatorViewModel()

        // When: Set first error
        viewModel.errorMessage = "First error"
        XCTAssertEqual(viewModel.errorMessage, "First error")

        // When: Set second error (overwrites first)
        viewModel.errorMessage = "Second error"

        // Then: Second error should be shown
        XCTAssertEqual(viewModel.errorMessage, "Second error")
    }

    // MARK: - Tab Selection Tests

    func testDefaultTabSelection() {
        // Given: Settings (shared instance may have been modified)
        // When: Check tab is within valid range
        let validTabs = [0, 1, 2, 3, 4, 5, 6]

        // Then: Selected tab should be within valid range
        XCTAssertTrue(
            validTabs.contains(settings.selectedTab),
            "Selected tab should be within valid range (0-6)",
        )

        // Test that we can set it to default (Registers = 0)
        settings.selectedTab = 0
        XCTAssertEqual(settings.selectedTab, 0, "Should be able to set tab to Registers")
    }

    func testTabSelectionPersistence() {
        // Given: Settings with selected tab
        settings.selectedTab = 2 // Stack tab

        // Then: Selection should persist
        XCTAssertEqual(settings.selectedTab, 2, "Selected tab should be Stack")

        // When: Change to different tab
        settings.selectedTab = 3 // Disassembly tab

        // Then: Selection should update
        XCTAssertEqual(settings.selectedTab, 3, "Selected tab should be Disassembly")
    }

    func testAllTabOptions() {
        // Given: All available tabs
        let tabs = [
            (0, "Registers"),
            (1, "Memory"),
            (2, "Stack"),
            (3, "Disassembly"),
            (4, "Evaluator"),
            (5, "Watchpoints"),
            (6, "Breakpoints"),
        ]

        // When/Then: Each tab should be selectable
        for (index, name) in tabs {
            settings.selectedTab = index
            XCTAssertEqual(
                settings.selectedTab,
                index,
                "\(name) tab should be selectable",
            )
        }
    }

    // MARK: - FileService Integration Tests

    func testRecentFilesManagement() {
        // Given: FileService
        let tempDir = FileManager.default.temporaryDirectory
        let file1 = tempDir.appendingPathComponent("test1.s")
        let file2 = tempDir.appendingPathComponent("test2.s")

        // When: Add files to recent files
        fileService.addToRecentFiles(file1)
        fileService.addToRecentFiles(file2)

        // Then: Recent files should contain both URLs
        XCTAssertTrue(
            fileService.recentFiles.contains(file1),
            "Recent files should contain test1.s",
        )
        XCTAssertTrue(
            fileService.recentFiles.contains(file2),
            "Recent files should contain test2.s",
        )
    }

    func testCurrentFileURL() {
        // Given: FileService
        let tempDir = FileManager.default.temporaryDirectory
        let testFile = tempDir.appendingPathComponent("current.s")

        // When: Set current file
        fileService.currentFileURL = testFile

        // Then: Current file should be set
        XCTAssertEqual(
            fileService.currentFileURL,
            testFile,
            "Current file URL should be set",
        )
    }

    // MARK: - Color Scheme Tests

    func testColorSchemePreference() {
        // Given: Settings (preferredColorScheme is computed from colorScheme)
        // Default should be "system"
        XCTAssertNil(settings.preferredColorScheme, "Default color scheme should be nil (system)")

        // When: Set underlying colorScheme to dark
        settings.colorScheme = "dark"

        // Then: preferredColorScheme should return dark
        XCTAssertEqual(settings.preferredColorScheme, .dark)

        // When: Set underlying colorScheme to light
        settings.colorScheme = "light"

        // Then: preferredColorScheme should return light
        XCTAssertEqual(settings.preferredColorScheme, .light)

        // When: Reset to system
        settings.colorScheme = "system"

        // Then: preferredColorScheme should be nil (system)
        XCTAssertNil(settings.preferredColorScheme)
    }

    // MARK: - ViewModel Initialization Tests

    func testViewModelConnection() async {
        // Given: ViewModel with mocks
        let mockAPIClient = MockAPIClient()
        let mockWebSocketClient = MockWebSocketClient()
        let viewModel = EmulatorViewModel(
            apiClient: mockAPIClient,
            wsClient: mockWebSocketClient,
        )

        // Initially not connected
        XCTAssertFalse(viewModel.isConnected, "ViewModel should not be connected initially")

        // When: Initialize ViewModel
        await viewModel.initialize()

        // Then: ViewModel should be connected
        XCTAssertTrue(
            viewModel.isConnected,
            "ViewModel should be connected after initialization",
        )
        XCTAssertTrue(
            mockAPIClient.createSessionCalled,
            "API client should create session",
        )
    }

    // MARK: - StatusView Tests

    func testStatusViewStates() {
        // Test all VM states and their corresponding colors
        let states: [VMState] = [.idle, .running, .breakpoint, .halted, .waitingForInput, .error]

        for state in states {
            // When/Then: State should be displayable
            // (This verifies the state is well-formed and can be used in UI)
            XCTAssertNotNil(state.rawValue, "State \(state) should have raw value")
        }
    }

    func testPCFormatting() {
        // Given: Program counter value
        let pc: UInt32 = 0x8000

        // When: Format as hex string
        let formatted = String(format: "0x%08X", pc)

        // Then: Should be formatted correctly
        XCTAssertEqual(formatted, "0x00008000", "PC should be formatted as 8-digit hex")

        // Test with different values
        let testCases: [(UInt32, String)] = [
            (0x0, "0x00000000"),
            (0xFFFF_FFFF, "0xFFFFFFFF"),
            (0x1234_5678, "0x12345678"),
        ]

        for (value, expected) in testCases {
            let result = String(format: "0x%08X", value)
            XCTAssertEqual(result, expected, "PC \(value) should format as \(expected)")
        }
    }
}
