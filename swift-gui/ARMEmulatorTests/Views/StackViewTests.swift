import XCTest
@testable import ARMEmulator

// MARK: - StackView Logic Tests

final class StackViewCalculationsTests: XCTestCase {
    func testStackSizeCalculation() {
        // Initial SP value from vm/constants.go
        let initialSP: UInt32 = 0x0005_0000

        // Stack grows downward (SP decreases as stack grows)
        func calculateStackSize(currentSP: UInt32) -> UInt32 {
            currentSP < initialSP ? initialSP - currentSP : 0
        }

        // No stack usage yet (SP at or above initial value)
        XCTAssertEqual(calculateStackSize(currentSP: 0x0005_0000), 0)
        XCTAssertEqual(calculateStackSize(currentSP: 0x0005_0001), 0)

        // Stack usage after pushing
        XCTAssertEqual(calculateStackSize(currentSP: 0x0004_FFFC), 4) // Pushed 1 word
        XCTAssertEqual(calculateStackSize(currentSP: 0x0004_FFF8), 8) // Pushed 2 words
        XCTAssertEqual(calculateStackSize(currentSP: 0x0004_FF00), 256) // Pushed 64 words
    }

    func testStackOffsetCalculation() {
        // Offset from current SP (SP+0, SP+4, SP+8, etc.)
        func calculateOffset(address: UInt32, currentSP: UInt32) -> Int {
            Int(address) - Int(currentSP)
        }

        let currentSP: UInt32 = 0x0004_FFF0

        XCTAssertEqual(calculateOffset(address: 0x0004_FFF0, currentSP: currentSP), 0) // SP+0
        XCTAssertEqual(calculateOffset(address: 0x0004_FFF4, currentSP: currentSP), 4) // SP+4
        XCTAssertEqual(calculateOffset(address: 0x0004_FFF8, currentSP: currentSP), 8) // SP+8
        XCTAssertEqual(calculateOffset(address: 0x0004_FFFC, currentSP: currentSP), 12) // SP+12
    }

    func testStackSanityCheck() {
        // Stack should not exceed 64KB (max stack size from constants)
        let maxStackSize = 65536

        func isValidStackSize(bytesToRead: Int) -> Bool {
            bytesToRead > 0 && bytesToRead <= maxStackSize
        }

        XCTAssertTrue(isValidStackSize(bytesToRead: 4))
        XCTAssertTrue(isValidStackSize(bytesToRead: 1024))
        XCTAssertTrue(isValidStackSize(bytesToRead: 65536))
        XCTAssertFalse(isValidStackSize(bytesToRead: 0))
        XCTAssertFalse(isValidStackSize(bytesToRead: 65537))
        XCTAssertFalse(isValidStackSize(bytesToRead: 1_000_000))
    }
}

// MARK: - StackEntry Tests

final class StackEntryTests: XCTestCase {
    func testStackEntryCreation() {
        let entry = StackEntry(
            offset: 0,
            address: 0x0004_FFF0,
            value: 0x0000_002A,
            annotation: "← R0?",
        )

        XCTAssertEqual(entry.offset, 0)
        XCTAssertEqual(entry.address, 0x0004_FFF0)
        XCTAssertEqual(entry.value, 0x0000_002A)
        XCTAssertEqual(entry.annotation, "← R0?")
        XCTAssertNotNil(entry.id) // UUID should be generated
    }

    func testStackEntryIsIdentifiable() {
        // Verify StackEntry conforms to Identifiable
        let entry1 = StackEntry(offset: 0, address: 0x1000, value: 0, annotation: "")
        let entry2 = StackEntry(offset: 4, address: 0x1004, value: 1, annotation: "")

        // Each entry should have a unique ID
        XCTAssertNotEqual(entry1.id, entry2.id)
    }
}

// MARK: - StackRow Format Tests

final class StackRowFormatTests: XCTestCase {
    func testStackOffsetFormatting() {
        // Stack offsets are displayed as "SP+N" or "SP-N"
        func formatOffset(_ offset: Int) -> String {
            String(format: "SP%+d", offset)
        }

        XCTAssertEqual(formatOffset(0), "SP+0")
        XCTAssertEqual(formatOffset(4), "SP+4")
        XCTAssertEqual(formatOffset(8), "SP+8")
        XCTAssertEqual(formatOffset(-4), "SP-4") // Negative offsets possible
    }

    func testStackAddressFormatting() {
        let address1: UInt32 = 0x0004_FFF0
        let formatted1 = String(format: "0x%08X", address1)
        XCTAssertEqual(formatted1, "0x0004FFF0")

        let address2: UInt32 = 0x0005_0000
        let formatted2 = String(format: "0x%08X", address2)
        XCTAssertEqual(formatted2, "0x00050000")
    }

    func testStackValueFormatting() {
        let value1: UInt32 = 0x0000_002A
        let formatted1 = String(format: "0x%08X", value1)
        XCTAssertEqual(formatted1, "0x0000002A")

        let value2: UInt32 = 0xFFFF_FFFF
        let formatted2 = String(format: "0x%08X", value2)
        XCTAssertEqual(formatted2, "0xFFFFFFFF")
    }

    func testASCIIRepresentation() {
        // Convert 4 bytes to ASCII representation (printable chars or dots)
        func asciiRepresentation(value: UInt32) -> String {
            let bytes = [
                UInt8(value & 0xFF),
                UInt8((value >> 8) & 0xFF),
                UInt8((value >> 16) & 0xFF),
                UInt8((value >> 24) & 0xFF),
            ]

            return bytes.map { byte in
                (32 ... 126).contains(byte) ? String(UnicodeScalar(byte)) : "."
            }.joined()
        }

        // Test printable ASCII
        XCTAssertEqual(asciiRepresentation(value: 0x4142_4344), "DCBA") // Little-endian

        // Test non-printable bytes
        XCTAssertEqual(asciiRepresentation(value: 0x0000_0000), "....")
        XCTAssertEqual(asciiRepresentation(value: 0x0100_0002), "....") // Control characters

        // Test mixed printable/non-printable
        XCTAssertEqual(asciiRepresentation(value: 0x4100_0042), "B.A") // 'B', 0, 0, 'A'
    }
}

// MARK: - Annotation Detection Tests

final class StackAnnotationTests: XCTestCase {
    func testCodeAddressDetection() {
        // Detect likely code addresses (typical range: 0x8000-0x10000)
        func isLikelyCodeAddress(_ value: UInt32) -> Bool {
            (0x8000 ... 0x10000).contains(value)
        }

        XCTAssertTrue(isLikelyCodeAddress(0x8000))
        XCTAssertTrue(isLikelyCodeAddress(0x9000))
        XCTAssertTrue(isLikelyCodeAddress(0x10000))
        XCTAssertFalse(isLikelyCodeAddress(0x7FFF))
        XCTAssertFalse(isLikelyCodeAddress(0x10001))
        XCTAssertFalse(isLikelyCodeAddress(0x50000))
    }

    func testStackAddressDetection() {
        // Detect values that might be stack pointers (within 4KB of current SP)
        func isLikelyStackAddress(value: UInt32, currentSP: UInt32) -> Bool {
            abs(Int(value) - Int(currentSP)) < 4096
        }

        let currentSP: UInt32 = 0x0004_F000

        XCTAssertTrue(isLikelyStackAddress(value: 0x0004_F000, currentSP: currentSP))
        XCTAssertTrue(isLikelyStackAddress(value: 0x0004_F100, currentSP: currentSP))
        XCTAssertTrue(isLikelyStackAddress(value: 0x0004_EFFF, currentSP: currentSP))
        XCTAssertFalse(isLikelyStackAddress(value: 0x0005_0000, currentSP: currentSP))
        XCTAssertFalse(isLikelyStackAddress(value: 0x0000_8000, currentSP: currentSP))
    }

    func testRegisterValueMatching() {
        // Detect values that match register contents (saved register values)
        func matchesRegister(value: UInt32, registerValues: [(String, UInt32)]) -> String? {
            for (name, regValue) in registerValues {
                if value == regValue, regValue != 0 {
                    return name
                }
            }
            return nil
        }

        let registers: [(String, UInt32)] = [
            ("R0", 42),
            ("R1", 100),
            ("R2", 0), // Zero values not annotated
        ]

        XCTAssertEqual(matchesRegister(value: 42, registerValues: registers), "R0")
        XCTAssertEqual(matchesRegister(value: 100, registerValues: registers), "R1")
        XCTAssertNil(matchesRegister(value: 0, registerValues: registers)) // Zero not matched
        XCTAssertNil(matchesRegister(value: 999, registerValues: registers)) // Not found
    }
}

// MARK: - Little-Endian Byte Conversion Tests

final class StackByteConversionTests: XCTestCase {
    func testLittleEndianConversion() {
        // Convert 4 bytes (little-endian) to UInt32
        func littleEndianToUInt32(_ bytes: [UInt8]) -> UInt32 {
            bytes.enumerated().reduce(UInt32(0)) { result, item in
                result | (UInt32(item.element) << (item.offset * 8))
            }
        }

        // Test simple value: 0x00000042
        XCTAssertEqual(littleEndianToUInt32([0x42, 0x00, 0x00, 0x00]), 0x0000_0042)

        // Test complex value: 0x12345678
        XCTAssertEqual(littleEndianToUInt32([0x78, 0x56, 0x34, 0x12]), 0x1234_5678)

        // Test all bytes set: 0xFFFFFFFF
        XCTAssertEqual(littleEndianToUInt32([0xFF, 0xFF, 0xFF, 0xFF]), 0xFFFF_FFFF)

        // Test zero
        XCTAssertEqual(littleEndianToUInt32([0x00, 0x00, 0x00, 0x00]), 0x0000_0000)
    }

    func testByteSlicing() {
        // Verify we can extract 4-byte words from memory data
        let memoryData: [UInt8] = [0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08]

        let word1 = Array(memoryData[0 ..< 4])
        XCTAssertEqual(word1, [0x01, 0x02, 0x03, 0x04])

        let word2 = Array(memoryData[4 ..< 8])
        XCTAssertEqual(word2, [0x05, 0x06, 0x07, 0x08])
    }
}

// MARK: - StackView Initialization Tests

@MainActor
final class StackViewInitializationTests: XCTestCase {
    func testInitWithEmptyStack() {
        let viewModel = EmulatorViewModel()
        viewModel.registers.sp = 0x0005_0000 // Initial SP (no stack usage)

        let view = StackView(viewModel: viewModel)

        XCTAssertNotNil(view)
    }

    func testInitWithStackData() {
        let viewModel = EmulatorViewModel()
        viewModel.registers.sp = 0x0004_FFF0 // 16 bytes of stack used

        let view = StackView(viewModel: viewModel)

        XCTAssertNotNil(view)
    }

    func testInitWithMaxStackUsage() {
        let viewModel = EmulatorViewModel()
        viewModel.registers.sp = 0x0004_0000 // 64KB stack used (max)

        let view = StackView(viewModel: viewModel)

        XCTAssertNotNil(view)
    }
}

// MARK: - StackView State Tests

@MainActor
final class StackViewStateTests: XCTestCase {
    func testEmptyStackMessage() {
        // When SP is at initial value, display "Stack is empty" message
        let initialSP: UInt32 = 0x0005_0000

        func shouldShowEmptyMessage(currentSP: UInt32) -> Bool {
            currentSP >= initialSP
        }

        XCTAssertTrue(shouldShowEmptyMessage(currentSP: 0x0005_0000))
        XCTAssertTrue(shouldShowEmptyMessage(currentSP: 0x0005_0001))
        XCTAssertFalse(shouldShowEmptyMessage(currentSP: 0x0004_FFFF))
    }

    func testStackHeaderFormatting() {
        // Test header text format: "Stack ↓ (SP = 0x...) - N bytes used"
        func formatHeader(sp: UInt32, stackSize: UInt32) -> String {
            "Stack ↓ (SP = \(String(format: "0x%08X", sp))) - \(stackSize) bytes used"
        }

        let header1 = formatHeader(sp: 0x0005_0000, stackSize: 0)
        XCTAssertEqual(header1, "Stack ↓ (SP = 0x00050000) - 0 bytes used")

        let header2 = formatHeader(sp: 0x0004_FFF0, stackSize: 16)
        XCTAssertEqual(header2, "Stack ↓ (SP = 0x0004FFF0) - 16 bytes used")
    }
}

// MARK: - Note on SwiftUI View Testing Limitations

/*
 StackView Testing Limitations:

 StackView uses @State for stackData and localMemoryData, which are not directly
 accessible in unit tests. The loadStack() method is private and async.

 What we CAN test:
 - Stack size calculation logic
 - Offset calculation from SP
 - Byte-to-word conversion (little-endian)
 - ASCII representation formatting
 - Annotation detection logic
 - Initialization with various SP values

 What we CANNOT easily test:
 - Async memory fetching via viewModel.fetchMemory()
 - ScrollView behavior
 - Row highlighting (isCurrent)
 - onChange handler triggering
 - Error handling during memory fetch

 Recommendations:
 1. Test calculation logic in isolation (done above)
 2. Use integration tests for memory fetching
 3. Use UI tests for visual verification
 4. Extract annotation logic to testable utility if needed
 */
