import XCTest
@testable import ARMEmulator

// MARK: - Command-Line Argument Parsing Tests

final class CommandLineArgumentParsingTests: XCTestCase {
    /// Test that command-line argument parsing filters out Xcode debug flags and their values
    func testFiltersOutDebugFlagsAndValues() {
        // Simulate command-line arguments with Xcode debug flags
        let testArgs = [
            "/path/to/app",
            "-NSDocumentRevisionsDebugMode",
            "YES",
            "-NSShowNonLocalizedStrings",
            "1",
        ]

        // Extract first .s file (mimicking the logic in AppDelegate)
        let filePath = testArgs.dropFirst().first(where: { $0.hasSuffix(".s") })

        // Should return nil because no .s files are present (filters out flags and values)
        XCTAssertNil(filePath)
    }

    /// Test that valid assembly files are correctly identified even with debug flags
    func testExtractsValidFilePath() {
        let testArgs = [
            "/path/to/app",
            "-NSDocumentRevisionsDebugMode",
            "YES",
            "/Users/test/example.s",
        ]

        let filePath = testArgs.dropFirst().first(where: { $0.hasSuffix(".s") })

        XCTAssertEqual(filePath, "/Users/test/example.s")
    }

    /// Test that no assembly file returns nil
    func testNoAssemblyFileReturnsNil() {
        let testArgs = [
            "/path/to/app",
            "-NSDocumentRevisionsDebugMode",
            "-NSShowNonLocalizedStrings",
            "random.txt",
        ]

        let filePath = testArgs.dropFirst().first(where: { $0.hasSuffix(".s") })

        XCTAssertNil(filePath)
    }

    /// Test that empty arguments (just the app path) returns nil
    func testEmptyArgumentsReturnsNil() {
        let testArgs = ["/path/to/app"]

        let filePath = testArgs.dropFirst().first(where: { $0.hasSuffix(".s") })

        XCTAssertNil(filePath)
    }

    /// Test that relative paths are preserved
    func testRelativePathPreserved() {
        let testArgs = [
            "/path/to/app",
            "examples/fibonacci.s",
        ]

        let filePath = testArgs.dropFirst().first(where: { $0.hasSuffix(".s") })

        XCTAssertEqual(filePath, "examples/fibonacci.s")
    }

    /// Test that only .s files are accepted (not other file types)
    func testOnlyAcceptsAssemblyFiles() {
        let testArgs = [
            "/path/to/app",
            "README.md",
            "program.txt",
            "test.s",
        ]

        let filePath = testArgs.dropFirst().first(where: { $0.hasSuffix(".s") })

        XCTAssertEqual(filePath, "test.s")
    }
}
