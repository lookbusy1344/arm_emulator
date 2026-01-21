import XCTest
@testable import ARMEmulator

final class RegisterStateTests: XCTestCase {
    // MARK: - RegisterState Factory Tests

    func testEmptyRegisterState() {
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

        // All flags should be false
        XCTAssertFalse(empty.cpsr.n)
        XCTAssertFalse(empty.cpsr.z)
        XCTAssertFalse(empty.cpsr.c)
        XCTAssertFalse(empty.cpsr.v)
    }

    // MARK: - RegisterState Codable Tests

    func testRegisterStateEncoding() throws {
        let state = RegisterState(
            r0: 42, r1: 100, r2: 200, r3: 300,
            r4: 400, r5: 500, r6: 600, r7: 700,
            r8: 800, r9: 900, r10: 1000, r11: 1100,
            r12: 1200, sp: 327_680, lr: 32768, pc: 32772,
            cpsr: CPSRFlags(n: true, z: false, c: true, v: false),
        )

        let data = try JSONEncoder().encode(state)
        let decoded = try JSONDecoder().decode(RegisterState.self, from: data)

        XCTAssertEqual(decoded, state)
    }

    func testRegisterStateDecoding() throws {
        let json = """
        {
            "r0": 42, "r1": 100, "r2": 200, "r3": 300,
            "r4": 400, "r5": 500, "r6": 600, "r7": 700,
            "r8": 800, "r9": 900, "r10": 1000, "r11": 1100,
            "r12": 1200, "sp": 327680, "lr": 32768, "pc": 32772,
            "cpsr": {"n": true, "z": false, "c": true, "v": false}
        }
        """

        let data = json.data(using: .utf8)!
        let state = try JSONDecoder().decode(RegisterState.self, from: data)

        XCTAssertEqual(state.r0, 42)
        XCTAssertEqual(state.r1, 100)
        XCTAssertEqual(state.sp, 327_680)
        XCTAssertEqual(state.lr, 32768)
        XCTAssertEqual(state.pc, 32772)
        XCTAssertTrue(state.cpsr.n)
        XCTAssertFalse(state.cpsr.z)
        XCTAssertTrue(state.cpsr.c)
        XCTAssertFalse(state.cpsr.v)
    }

    // MARK: - Special Register Value Tests

    func testStackPointerMaxValue() {
        let state = RegisterState(
            r0: 0, r1: 0, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: UInt32.max, lr: 0, pc: 0,
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false),
        )

        XCTAssertEqual(state.sp, UInt32.max)
    }

    func testProgramCounterEdgeCases() {
        // PC at 0 (program start)
        var state = RegisterState(
            r0: 0, r1: 0, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 0, lr: 0, pc: 0,
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false),
        )
        XCTAssertEqual(state.pc, 0)

        // PC at typical start address (0x8000 = 32768)
        state.pc = 32768
        XCTAssertEqual(state.pc, 32768)

        // PC at max value
        state.pc = UInt32.max
        XCTAssertEqual(state.pc, UInt32.max)
    }

    func testLinkRegisterEdgeCases() {
        let state = RegisterState(
            r0: 0, r1: 0, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 0, lr: 32772, pc: 32768,
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false),
        )

        // LR typically holds return address (4 bytes after PC)
        XCTAssertEqual(state.lr, 32772)
        XCTAssertEqual(state.pc + 4, state.lr)
    }

    func testAllRegistersMaxValue() {
        let state = RegisterState(
            r0: UInt32.max, r1: UInt32.max, r2: UInt32.max, r3: UInt32.max,
            r4: UInt32.max, r5: UInt32.max, r6: UInt32.max, r7: UInt32.max,
            r8: UInt32.max, r9: UInt32.max, r10: UInt32.max, r11: UInt32.max,
            r12: UInt32.max, sp: UInt32.max, lr: UInt32.max, pc: UInt32.max,
            cpsr: CPSRFlags(n: true, z: true, c: true, v: true),
        )

        XCTAssertEqual(state.r0, UInt32.max)
        XCTAssertEqual(state.r7, UInt32.max) // R7 has no special meaning in ARM2
        XCTAssertEqual(state.sp, UInt32.max)
        XCTAssertEqual(state.lr, UInt32.max)
        XCTAssertEqual(state.pc, UInt32.max)
    }

    // MARK: - Register Change Detection Tests

    func testRegisterStateEquality() {
        let state1 = RegisterState(
            r0: 42, r1: 100, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 327_680, lr: 0, pc: 32768,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false),
        )

        let state2 = RegisterState(
            r0: 42, r1: 100, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 327_680, lr: 0, pc: 32768,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false),
        )

        XCTAssertEqual(state1, state2)
    }

    func testRegisterStateInequalityFromRegisterChange() {
        let state1 = RegisterState(
            r0: 42, r1: 100, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 327_680, lr: 0, pc: 32768,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false),
        )

        let state2 = RegisterState(
            r0: 43, r1: 100, r2: 0, r3: 0, // R0 changed
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 327_680, lr: 0, pc: 32768,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false),
        )

        XCTAssertNotEqual(state1, state2)
    }

    func testRegisterStateInequalityFromFlagChange() {
        let state1 = RegisterState(
            r0: 42, r1: 100, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 327_680, lr: 0, pc: 32768,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false),
        )

        let state2 = RegisterState(
            r0: 42, r1: 100, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 327_680, lr: 0, pc: 32768,
            cpsr: CPSRFlags(n: true, z: true, c: false, v: false), // N flag changed
        )

        XCTAssertNotEqual(state1, state2)
    }

    func testRegisterStateInequalityFromPCChange() {
        let state1 = RegisterState(
            r0: 42, r1: 100, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 327_680, lr: 0, pc: 32768,
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false),
        )

        let state2 = RegisterState(
            r0: 42, r1: 100, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 327_680, lr: 0, pc: 32772, // PC changed
            cpsr: CPSRFlags(n: false, z: true, c: false, v: false),
        )

        XCTAssertNotEqual(state1, state2)
    }

    // MARK: - CPSRFlags Display Tests

    func testCPSRFlagsDisplayAllClear() {
        let flags = CPSRFlags(n: false, z: false, c: false, v: false)
        XCTAssertEqual(flags.displayString, "----")
    }

    func testCPSRFlagsDisplayAllSet() {
        let flags = CPSRFlags(n: true, z: true, c: true, v: true)
        XCTAssertEqual(flags.displayString, "NZCV")
    }

    func testCPSRFlagsDisplayOnlyNegative() {
        let flags = CPSRFlags(n: true, z: false, c: false, v: false)
        XCTAssertEqual(flags.displayString, "N---")
    }

    func testCPSRFlagsDisplayOnlyZero() {
        let flags = CPSRFlags(n: false, z: true, c: false, v: false)
        XCTAssertEqual(flags.displayString, "-Z--")
    }

    func testCPSRFlagsDisplayOnlyCarry() {
        let flags = CPSRFlags(n: false, z: false, c: true, v: false)
        XCTAssertEqual(flags.displayString, "--C-")
    }

    func testCPSRFlagsDisplayOnlyOverflow() {
        let flags = CPSRFlags(n: false, z: false, c: false, v: true)
        XCTAssertEqual(flags.displayString, "---V")
    }

    func testCPSRFlagsDisplayNegativeAndZero() {
        let flags = CPSRFlags(n: true, z: true, c: false, v: false)
        XCTAssertEqual(flags.displayString, "NZ--")
    }

    func testCPSRFlagsDisplayCarryAndOverflow() {
        let flags = CPSRFlags(n: false, z: false, c: true, v: true)
        XCTAssertEqual(flags.displayString, "--CV")
    }

    func testCPSRFlagsDisplayNegativeAndCarry() {
        let flags = CPSRFlags(n: true, z: false, c: true, v: false)
        XCTAssertEqual(flags.displayString, "N-C-")
    }

    func testCPSRFlagsDisplayZeroAndOverflow() {
        let flags = CPSRFlags(n: false, z: true, c: false, v: true)
        XCTAssertEqual(flags.displayString, "-Z-V")
    }

    func testCPSRFlagsDisplayThreeFlags() {
        let flags = CPSRFlags(n: true, z: true, c: true, v: false)
        XCTAssertEqual(flags.displayString, "NZC-")
    }

    // MARK: - CPSRFlags Codable Tests

    func testCPSRFlagsEncoding() throws {
        let flags = CPSRFlags(n: true, z: false, c: true, v: false)

        let data = try JSONEncoder().encode(flags)
        let decoded = try JSONDecoder().decode(CPSRFlags.self, from: data)

        XCTAssertEqual(decoded, flags)
    }

    func testCPSRFlagsDecoding() throws {
        let json = """
        {
            "n": true,
            "z": false,
            "c": true,
            "v": false
        }
        """

        let data = json.data(using: .utf8)!
        let flags = try JSONDecoder().decode(CPSRFlags.self, from: data)

        XCTAssertTrue(flags.n)
        XCTAssertFalse(flags.z)
        XCTAssertTrue(flags.c)
        XCTAssertFalse(flags.v)
    }

    // MARK: - CPSRFlags Equality Tests

    func testCPSRFlagsEquality() {
        let flags1 = CPSRFlags(n: true, z: false, c: true, v: false)
        let flags2 = CPSRFlags(n: true, z: false, c: true, v: false)

        XCTAssertEqual(flags1, flags2)
    }

    func testCPSRFlagsInequality() {
        let flags1 = CPSRFlags(n: true, z: false, c: true, v: false)
        let flags2 = CPSRFlags(n: true, z: true, c: true, v: false) // Z flag different

        XCTAssertNotEqual(flags1, flags2)
    }

    // MARK: - Comprehensive Flag Combination Tests (All 16 Combinations)

    func testAllCPSRFlagCombinations() {
        struct FlagTest {
            let n: Bool
            let z: Bool
            let c: Bool
            let v: Bool
            let expected: String
        }

        let combinations: [FlagTest] = [
            FlagTest(n: false, z: false, c: false, v: false, expected: "----"),
            FlagTest(n: false, z: false, c: false, v: true, expected: "---V"),
            FlagTest(n: false, z: false, c: true, v: false, expected: "--C-"),
            FlagTest(n: false, z: false, c: true, v: true, expected: "--CV"),
            FlagTest(n: false, z: true, c: false, v: false, expected: "-Z--"),
            FlagTest(n: false, z: true, c: false, v: true, expected: "-Z-V"),
            FlagTest(n: false, z: true, c: true, v: false, expected: "-ZC-"),
            FlagTest(n: false, z: true, c: true, v: true, expected: "-ZCV"),
            FlagTest(n: true, z: false, c: false, v: false, expected: "N---"),
            FlagTest(n: true, z: false, c: false, v: true, expected: "N--V"),
            FlagTest(n: true, z: false, c: true, v: false, expected: "N-C-"),
            FlagTest(n: true, z: false, c: true, v: true, expected: "N-CV"),
            FlagTest(n: true, z: true, c: false, v: false, expected: "NZ--"),
            FlagTest(n: true, z: true, c: false, v: true, expected: "NZ-V"),
            FlagTest(n: true, z: true, c: true, v: false, expected: "NZC-"),
            FlagTest(n: true, z: true, c: true, v: true, expected: "NZCV"),
        ]

        for test in combinations {
            let flags = CPSRFlags(n: test.n, z: test.z, c: test.c, v: test.v)
            XCTAssertEqual(
                flags.displayString, test.expected,
                "Flag combination N=\(test.n) Z=\(test.z) C=\(test.c) V=\(test.v) should display as '\(test.expected)'",
            )
        }
    }
}
