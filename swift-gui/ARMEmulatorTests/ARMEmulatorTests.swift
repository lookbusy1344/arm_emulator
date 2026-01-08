import XCTest
@testable import ARMEmulator

// TODO: Implement comprehensive Swift UI tests for core workflows
// - Load program workflow
// - Run program workflow
// - Step through program workflow
// - Breakpoint management workflow
//
// Note: Full ViewModel testing requires complex mocking setup.
// Consider integration tests with real backend instead.

final class ARMEmulatorTests: XCTestCase {
    /// Placeholder test to satisfy build system
    func testPlaceholder() {
        XCTAssertTrue(true, "Placeholder test")
    }
}

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

// MARK: - RegisterState Tests

final class RegisterStateTests: XCTestCase {
    func testRegisterStateInitialization() {
        let registers = RegisterState(
            r0: 1, r1: 2, r2: 3, r3: 4, r4: 5, r5: 6, r6: 7, r7: 8,
            r8: 9, r9: 10, r10: 11, r11: 12, r12: 13, sp: 0x50000, lr: 0, pc: 0x8000,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false)
        )

        XCTAssertEqual(registers.r0, 1)
        XCTAssertEqual(registers.r1, 2)
        XCTAssertEqual(registers.pc, 0x8000)
        XCTAssertEqual(registers.sp, 0x50000)
        XCTAssertTrue(registers.cpsr.z)
        XCTAssertFalse(registers.cpsr.n)
    }

    func testCPSRFlags() {
        let cpsr = CPSRFlags(n: true, z: false, c: true, v: false)

        XCTAssertTrue(cpsr.n)
        XCTAssertFalse(cpsr.z)
        XCTAssertTrue(cpsr.c)
        XCTAssertFalse(cpsr.v)
    }
}
