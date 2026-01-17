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

    func testRegisterStateEmpty() {
        let empty = RegisterState.empty

        XCTAssertEqual(empty.r0, 0)
        XCTAssertEqual(empty.r1, 0)
        XCTAssertEqual(empty.r2, 0)
        XCTAssertEqual(empty.r3, 0)
        XCTAssertEqual(empty.r4, 0)
        XCTAssertEqual(empty.r5, 0)
        XCTAssertEqual(empty.r6, 0)
        XCTAssertEqual(empty.r7, 0)
        XCTAssertEqual(empty.r8, 0)
        XCTAssertEqual(empty.r9, 0)
        XCTAssertEqual(empty.r10, 0)
        XCTAssertEqual(empty.r11, 0)
        XCTAssertEqual(empty.r12, 0)
        XCTAssertEqual(empty.sp, 0)
        XCTAssertEqual(empty.lr, 0)
        XCTAssertEqual(empty.pc, 0)
        XCTAssertFalse(empty.cpsr.n)
        XCTAssertFalse(empty.cpsr.z)
        XCTAssertFalse(empty.cpsr.c)
        XCTAssertFalse(empty.cpsr.v)
    }

    func testRegisterStateEquality() {
        let registers1 = RegisterState(
            r0: 1, r1: 2, r2: 3, r3: 4, r4: 5, r5: 6, r6: 7, r7: 8,
            r8: 9, r9: 10, r10: 11, r11: 12, r12: 13, sp: 0x50000, lr: 0, pc: 0x8000,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false)
        )

        let registers2 = RegisterState(
            r0: 1, r1: 2, r2: 3, r3: 4, r4: 5, r5: 6, r6: 7, r7: 8,
            r8: 9, r9: 10, r10: 11, r11: 12, r12: 13, sp: 0x50000, lr: 0, pc: 0x8000,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false)
        )

        let registers3 = RegisterState(
            r0: 42, r1: 2, r2: 3, r3: 4, r4: 5, r5: 6, r6: 7, r7: 8,
            r8: 9, r9: 10, r10: 11, r11: 12, r12: 13, sp: 0x50000, lr: 0, pc: 0x8000,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false)
        )

        XCTAssertEqual(registers1, registers2)
        XCTAssertNotEqual(registers1, registers3)
    }

    func testCPSRFlags() {
        let cpsr = CPSRFlags(n: true, z: false, c: true, v: false)

        XCTAssertTrue(cpsr.n)
        XCTAssertFalse(cpsr.z)
        XCTAssertTrue(cpsr.c)
        XCTAssertFalse(cpsr.v)
    }

    func testCPSRFlagsDisplayStringAllSet() {
        let cpsr = CPSRFlags(n: true, z: true, c: true, v: true)
        XCTAssertEqual(cpsr.displayString, "NZCV")
    }

    func testCPSRFlagsDisplayStringAllClear() {
        let cpsr = CPSRFlags(n: false, z: false, c: false, v: false)
        XCTAssertEqual(cpsr.displayString, "----")
    }

    func testCPSRFlagsDisplayStringMixed() {
        let cpsr1 = CPSRFlags(n: true, z: false, c: true, v: false)
        XCTAssertEqual(cpsr1.displayString, "N-C-")

        let cpsr2 = CPSRFlags(n: false, z: true, c: false, v: false)
        XCTAssertEqual(cpsr2.displayString, "-Z--")

        let cpsr3 = CPSRFlags(n: false, z: false, c: false, v: true)
        XCTAssertEqual(cpsr3.displayString, "---V")
    }

    func testCPSRFlagsEquality() {
        let flags1 = CPSRFlags(n: true, z: false, c: true, v: false)
        let flags2 = CPSRFlags(n: true, z: false, c: true, v: false)
        let flags3 = CPSRFlags(n: false, z: false, c: true, v: false)

        XCTAssertEqual(flags1, flags2)
        XCTAssertNotEqual(flags1, flags3)
    }
}
