import XCTest
@testable import ARMEmulator

// MARK: - ExampleProgram Tests

final class ExampleProgramTests: XCTestCase {
    func testExampleProgramInitialization() {
        let url = URL(fileURLWithPath: "/examples/fibonacci.s")
        let program = ExampleProgram(
            name: "fibonacci",
            filename: "fibonacci.s",
            description: "Calculate Fibonacci sequence",
            size: 1234,
            url: url,
        )

        XCTAssertEqual(program.name, "fibonacci")
        XCTAssertEqual(program.filename, "fibonacci.s")
        XCTAssertEqual(program.description, "Calculate Fibonacci sequence")
        XCTAssertEqual(program.size, 1234)
        XCTAssertEqual(program.url, url)
    }

    func testExampleProgramFormattedSize() {
        let program = ExampleProgram(
            name: "test",
            filename: "test.s",
            description: "Test",
            size: 1024,
            url: URL(fileURLWithPath: "/test.s"),
        )

        // ByteCountFormatter may vary by locale, so just verify it's not empty
        XCTAssertFalse(program.formattedSize.isEmpty)
        XCTAssertTrue(program.formattedSize.contains("KB") || program.formattedSize.contains("bytes"))
    }

    func testExampleProgramHashable() {
        let url = URL(fileURLWithPath: "/test.s")
        let program1 = ExampleProgram(name: "test", filename: "test.s", description: "Test", size: 100, url: url)
        let program2 = ExampleProgram(name: "test", filename: "test.s", description: "Test", size: 100, url: url)

        // Different instances should have different IDs
        XCTAssertNotEqual(program1.id, program2.id)
        XCTAssertNotEqual(program1, program2)

        // Should be usable in Set
        var set: Set<ExampleProgram> = [program1]
        set.insert(program2)
        XCTAssertEqual(set.count, 2)
    }

    func testExampleProgramEquality() {
        let program1 = ExampleProgram(
            name: "test",
            filename: "test.s",
            description: "Test",
            size: 100,
            url: URL(fileURLWithPath: "/test.s"),
        )
        let program2 = ExampleProgram(
            name: "test",
            filename: "test.s",
            description: "Test",
            size: 100,
            url: URL(fileURLWithPath: "/test.s"),
        )

        // Equality based on ID only
        XCTAssertEqual(program1, program1) // Same instance
        XCTAssertNotEqual(program1, program2) // Different instances, different IDs
    }
}

// MARK: - FileService Recent Files Tests

@MainActor
final class FileServiceRecentFilesTests: XCTestCase {
    var fileService: FileService!

    override func setUp() async throws {
        try await super.setUp()
        // Note: FileService is a singleton, so we can't easily create a fresh instance
        // We'll clear recent files before each test
        fileService = FileService.shared
        fileService.clearRecentFiles()
    }

    override func tearDown() async throws {
        fileService.clearRecentFiles()
        try await super.tearDown()
    }

    func testAddToRecentFiles() {
        let url1 = URL(fileURLWithPath: "/test1.s")
        fileService.addToRecentFiles(url1)

        XCTAssertEqual(fileService.recentFiles.count, 1)
        XCTAssertEqual(fileService.recentFiles[0], url1)
    }

    func testAddToRecentFilesDeduplication() {
        let url1 = URL(fileURLWithPath: "/test1.s")
        let url2 = URL(fileURLWithPath: "/test2.s")

        fileService.addToRecentFiles(url1)
        fileService.addToRecentFiles(url2)
        fileService.addToRecentFiles(url1) // Add url1 again

        // Should have 2 files, with url1 at the front
        XCTAssertEqual(fileService.recentFiles.count, 2)
        XCTAssertEqual(fileService.recentFiles[0], url1)
        XCTAssertEqual(fileService.recentFiles[1], url2)
    }

    func testAddToRecentFilesOrdering() {
        let url1 = URL(fileURLWithPath: "/test1.s")
        let url2 = URL(fileURLWithPath: "/test2.s")
        let url3 = URL(fileURLWithPath: "/test3.s")

        fileService.addToRecentFiles(url1)
        fileService.addToRecentFiles(url2)
        fileService.addToRecentFiles(url3)

        // Most recent should be first
        XCTAssertEqual(fileService.recentFiles[0], url3)
        XCTAssertEqual(fileService.recentFiles[1], url2)
        XCTAssertEqual(fileService.recentFiles[2], url1)
    }

    func testAddToRecentFilesMaxLimit() {
        let maxFiles = AppSettings.shared.maxRecentFiles

        // Add more files than the limit (maxFiles + 5)
        for i in 0 ..< maxFiles + 5 {
            let url = URL(fileURLWithPath: "/test\(i).s")
            fileService.addToRecentFiles(url)
        }

        // Should not exceed max
        XCTAssertEqual(fileService.recentFiles.count, maxFiles)

        // Most recent file should be the last one added
        XCTAssertEqual(fileService.recentFiles[0].lastPathComponent, "test\(maxFiles + 4).s")

        // Oldest file kept should be the one at index (maxFiles + 4 - maxFiles + 1) = 5
        XCTAssertEqual(fileService.recentFiles[maxFiles - 1].lastPathComponent, "test5.s")
    }

    func testClearRecentFiles() {
        let url1 = URL(fileURLWithPath: "/test1.s")
        let url2 = URL(fileURLWithPath: "/test2.s")

        fileService.addToRecentFiles(url1)
        fileService.addToRecentFiles(url2)
        XCTAssertEqual(fileService.recentFiles.count, 2)

        fileService.clearRecentFiles()
        XCTAssertEqual(fileService.recentFiles.count, 0)
    }
}

// MARK: - FileService Path Resolution Tests

final class FileServicePathResolutionTests: XCTestCase {
    // Note: findExamplesDirectory() is a private method and depends on actual filesystem.
    // Testing it would require:
    // 1. Making it internal (with @testable import)
    // 2. Creating a temporary directory structure for testing
    // 3. Or testing it indirectly through loadExamples()

    // Note: extractDescription() is also private and cannot be tested directly.
    // To test these methods, we would need to refactor FileService to:
    // 1. Mark them as internal for testing
    // 2. Extract them to a separate, testable utility class
    // 3. Or inject them as dependencies with protocols

    // Current untestable private methods:
    // - findExamplesDirectory() -> URL? (filesystem-dependent)
    // - extractDescription(from: String) -> String (pure logic, could be made testable)
}
