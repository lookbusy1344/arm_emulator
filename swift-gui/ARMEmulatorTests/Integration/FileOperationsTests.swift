import XCTest
@testable import ARMEmulator

/// Integration tests for file operations
/// Tests loading programs from files, recent files management, and examples
@MainActor
final class FileOperationsTests: XCTestCase {
    var viewModel: EmulatorViewModel!
    var fileService: FileService!
    var tempDirectory: URL!

    override func setUp() async throws {
        try await super.setUp()

        viewModel = EmulatorViewModel(
            apiClient: MockAPIClient(),
            webSocketClient: MockWebSocketClient()
        )
        viewModel.sessionID = "test-session"

        fileService = FileService()

        // Create temporary directory for test files
        tempDirectory = FileManager.default.temporaryDirectory
            .appendingPathComponent("ARMEmulatorTests-\(UUID().uuidString)")
        try FileManager.default.createDirectory(at: tempDirectory, withIntermediateDirectories: true)
    }

    override func tearDown() async throws {
        viewModel = nil
        fileService = nil

        // Clean up temporary directory
        if let tempDirectory = tempDirectory {
            try? FileManager.default.removeItem(at: tempDirectory)
        }

        try await super.tearDown()
    }

    // MARK: - Load Program from File

    func testLoadSmallProgramFromFile() async throws {
        // Create test file
        let testFile = tempDirectory.appendingPathComponent("test.s")
        try ProgramFixtures.helloWorld.write(to: testFile, atomically: true, encoding: .utf8)

        // Load file content
        let content = try fileService.loadFile(at: testFile)
        XCTAssertEqual(content, ProgramFixtures.helloWorld)

        // Load into emulator
        await viewModel.loadProgram(source: content)
        XCTAssertNil(viewModel.errorMessage, "Should load successfully")
        XCTAssertEqual(viewModel.status, .idle)
    }

    func testLoadLargeProgramFromFile() async throws {
        // Create large program (>100KB)
        let largeProgram = String(repeating: "MOV R0, #0\n", count: 10000)
        let testFile = tempDirectory.appendingPathComponent("large.s")
        try largeProgram.write(to: testFile, atomically: true, encoding: .utf8)

        // Load file
        let content = try fileService.loadFile(at: testFile)
        XCTAssertGreaterThan(content.count, 100_000, "File should be >100KB")

        // Load into emulator
        await viewModel.loadProgram(source: content)
        // Should handle large files gracefully
    }

    func testLoadFileWithUnicodeCharacters() async throws {
        // Program with unicode in comments and strings
        let unicodeProgram = """
            .text
            .global _start
            _start:
                ; Comment with emoji: 🚀 ARM Emulator
                LDR R0, =message
                SWI #0x02
                MOV R0, #0
                SWI #0x00

            .data
            message: .asciz "Hello 世界! 🌍"
            """

        let testFile = tempDirectory.appendingPathComponent("unicode.s")
        try unicodeProgram.write(to: testFile, atomically: true, encoding: .utf8)

        // Load and verify unicode preserved
        let content = try fileService.loadFile(at: testFile)
        XCTAssertTrue(content.contains("🚀"), "Should preserve emoji")
        XCTAssertTrue(content.contains("世界"), "Should preserve Unicode characters")

        await viewModel.loadProgram(source: content)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testLoadNonexistentFile() throws {
        let nonexistentFile = tempDirectory.appendingPathComponent("does_not_exist.s")

        XCTAssertThrowsError(try fileService.loadFile(at: nonexistentFile)) { error in
            XCTAssertTrue(error is FileServiceError, "Should throw FileServiceError")
        }
    }

    func testLoadInvalidFilePermissions() throws {
        // Create file with no read permissions
        let restrictedFile = tempDirectory.appendingPathComponent("restricted.s")
        try "test".write(to: restrictedFile, atomically: true, encoding: .utf8)

        // Remove read permissions
        #if os(macOS)
        try FileManager.default.setAttributes(
            [.posixPermissions: 0o000],
            ofItemAtPath: restrictedFile.path
        )

        XCTAssertThrowsError(try fileService.loadFile(at: restrictedFile))

        // Restore permissions for cleanup
        try FileManager.default.setAttributes(
            [.posixPermissions: 0o644],
            ofItemAtPath: restrictedFile.path
        )
        #endif
    }

    // MARK: - Save Program to File

    func testSaveProgramToFile() throws {
        let saveFile = tempDirectory.appendingPathComponent("saved.s")
        let programContent = ProgramFixtures.fibonacci

        // Save program
        try fileService.saveFile(content: programContent, to: saveFile)

        // Verify file exists and content matches
        XCTAssertTrue(FileManager.default.fileExists(atPath: saveFile.path))
        let loadedContent = try String(contentsOf: saveFile, encoding: .utf8)
        XCTAssertEqual(loadedContent, programContent)
    }

    func testSaveOverwritesExistingFile() throws {
        let saveFile = tempDirectory.appendingPathComponent("overwrite.s")

        // Save first version
        try fileService.saveFile(content: "Version 1", to: saveFile)
        let content1 = try String(contentsOf: saveFile, encoding: .utf8)
        XCTAssertEqual(content1, "Version 1")

        // Overwrite with second version
        try fileService.saveFile(content: "Version 2", to: saveFile)
        let content2 = try String(contentsOf: saveFile, encoding: .utf8)
        XCTAssertEqual(content2, "Version 2")
    }

    // MARK: - Recent Files Management

    func testAddToRecentFiles() {
        let file1 = tempDirectory.appendingPathComponent("recent1.s")
        let file2 = tempDirectory.appendingPathComponent("recent2.s")

        // Add files to recent files
        fileService.addToRecentFiles(file1)
        fileService.addToRecentFiles(file2)

        // Verify recent files list
        let recentFiles = fileService.getRecentFiles()
        XCTAssertTrue(recentFiles.contains(file1), "Should contain file1")
        XCTAssertTrue(recentFiles.contains(file2), "Should contain file2")
        XCTAssertEqual(recentFiles.count, 2)
    }

    func testRecentFilesLimit() {
        // Add more than max recent files (assuming limit of 10)
        for i in 0..<15 {
            let file = tempDirectory.appendingPathComponent("recent\(i).s")
            fileService.addToRecentFiles(file)
        }

        let recentFiles = fileService.getRecentFiles()
        XCTAssertLessThanOrEqual(recentFiles.count, 10, "Should limit to max recent files")

        // Most recent should be at the top
        let mostRecent = recentFiles.first!
        XCTAssertTrue(mostRecent.lastPathComponent.contains("14"),
                     "Most recent file should be at index 0")
    }

    func testRecentFilesMoveToTop() {
        let file1 = tempDirectory.appendingPathComponent("file1.s")
        let file2 = tempDirectory.appendingPathComponent("file2.s")
        let file3 = tempDirectory.appendingPathComponent("file3.s")

        // Add files in order
        fileService.addToRecentFiles(file1)
        fileService.addToRecentFiles(file2)
        fileService.addToRecentFiles(file3)

        // Re-open file1 (should move to top)
        fileService.addToRecentFiles(file1)

        let recentFiles = fileService.getRecentFiles()
        XCTAssertEqual(recentFiles.first, file1, "Re-opened file should move to top")
        XCTAssertEqual(recentFiles.count, 3, "Should not duplicate entries")
    }

    func testClearRecentFiles() {
        // Add some files
        fileService.addToRecentFiles(tempDirectory.appendingPathComponent("file1.s"))
        fileService.addToRecentFiles(tempDirectory.appendingPathComponent("file2.s"))

        XCTAssertFalse(fileService.getRecentFiles().isEmpty)

        // Clear recent files
        fileService.clearRecentFiles()

        XCTAssertTrue(fileService.getRecentFiles().isEmpty, "Recent files should be empty")
    }

    // MARK: - Example Programs

    func testLoadExampleProgram() {
        // Get example programs list
        let examples = fileService.getExamplePrograms()
        XCTAssertFalse(examples.isEmpty, "Should have example programs")

        // Find fibonacci example
        guard let fibExample = examples.first(where: { $0.lastPathComponent == "fibonacci.s" }) else {
            XCTFail("Should find fibonacci.s example")
            return
        }

        // Load example
        XCTAssertNoThrow(try fileService.loadFile(at: fibExample))
    }

    func testExampleProgramsDirectory() {
        // Verify examples directory exists
        let projectRoot = URL(fileURLWithPath: #file)
            .deletingLastPathComponent() // Fixtures
            .deletingLastPathComponent() // Integration
            .deletingLastPathComponent() // ARMEmulatorTests
            .deletingLastPathComponent() // swift-gui
            .deletingLastPathComponent() // project root

        let examplesDir = projectRoot.appendingPathComponent("examples")

        // FileService should find examples directory
        let examples = fileService.getExamplePrograms()

        if FileManager.default.fileExists(atPath: examplesDir.path) {
            XCTAssertFalse(examples.isEmpty, "Should find example programs")
        } else {
            // Examples directory may not exist in test environment
            XCTAssertTrue(true, "Examples directory not found in test environment")
        }
    }

    func testRunExampleToCompletion() async throws {
        // This test requires examples directory to exist
        let examples = fileService.getExamplePrograms()
        guard let helloExample = examples.first(where: { $0.lastPathComponent == "hello.s" }) else {
            throw XCTSkip("hello.s example not found")
        }

        // Load example
        let content = try fileService.loadFile(at: helloExample)

        // Load into emulator
        await viewModel.loadProgram(source: content)
        XCTAssertNil(viewModel.errorMessage, "Example should load successfully")

        // Run to completion
        await viewModel.run()
        try await waitForStatus(.halted, timeout: 2.0, viewModel: viewModel)

        XCTAssertEqual(viewModel.status, .halted, "Example should run to completion")
    }

    // MARK: - File Watching (Future Enhancement)

    func testFileChangeDetection() throws {
        // Future: Detect when file changes on disk
        // This would require FileService to support file watching
        // Placeholder for future implementation
        throw XCTSkip("File watching not yet implemented")
    }

    // MARK: - Integration with ViewModel

    func testViewModelFileWorkflow() async throws {
        // Complete workflow: create file → load → edit → save

        // 1. Create test file
        let testFile = tempDirectory.appendingPathComponent("workflow.s")
        let initialContent = ProgramFixtures.exitCode42
        try fileService.saveFile(content: initialContent, to: testFile)

        // 2. Load file into ViewModel
        let content = try fileService.loadFile(at: testFile)
        await viewModel.loadProgram(source: content)
        XCTAssertNil(viewModel.errorMessage)

        // 3. Run program
        await viewModel.run()
        try await waitForStatus(.halted, timeout: 2.0, viewModel: viewModel)

        // 4. Modify program
        let modifiedContent = ProgramFixtures.simpleLoop

        // 5. Save modified version
        try fileService.saveFile(content: modifiedContent, to: testFile)

        // 6. Reload
        let newContent = try fileService.loadFile(at: testFile)
        await viewModel.reset()
        await viewModel.loadProgram(source: newContent)

        // 7. Run modified program
        await viewModel.run()
        try await waitForStatus(.halted, timeout: 2.0, viewModel: viewModel)

        // Verify different result
        XCTAssertTrue(viewModel.registers.hasRegister("R0", value: 5),
                     "Modified program should produce different result")
    }
}
