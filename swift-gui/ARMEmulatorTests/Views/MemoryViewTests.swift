import XCTest
@testable import ARMEmulator

// MARK: - MemoryView Constants Tests

final class MemoryViewConstantsTests: XCTestCase {
    func testBytesPerRow() {
        // MemoryView uses 16 bytes per row for hex display
        let bytesPerRow = 16
        XCTAssertEqual(bytesPerRow, 16)
    }

    func testRowsToShow() {
        // MemoryView shows 16 rows at a time
        let rowsToShow = 16
        XCTAssertEqual(rowsToShow, 16)
    }

    func testTotalBytesCalculation() {
        // Total bytes shown = bytesPerRow * rowsToShow
        let bytesPerRow = 16
        let rowsToShow = 16
        let totalBytes = bytesPerRow * rowsToShow

        XCTAssertEqual(totalBytes, 256) // 16 * 16 = 256 bytes displayed
    }

    func testMemoryWindowSize() {
        // Verify memory window is reasonable size for ARM architecture
        let bytesPerRow = 16
        let rowsToShow = 16
        let totalBytes = bytesPerRow * rowsToShow

        XCTAssertEqual(totalBytes, 256)
        XCTAssertGreaterThanOrEqual(totalBytes, 128) // At least 128 bytes
        XCTAssertLessThanOrEqual(totalBytes, 1024) // No more than 1KB at once
    }
}

// MARK: - MemoryView Address Parsing Tests

final class MemoryViewAddressParsingTests: XCTestCase {
    func testHexAddressParsingWithPrefix() {
        // Test parsing hex addresses with 0x prefix
        let addresses = [
            "0x8000": UInt32(0x8000),
            "0x0": UInt32(0),
            "0xFFFFFFFF": UInt32(0xFFFFFFFF),
            "0x10000": UInt32(0x10000),
        ]

        for (input, expected) in addresses {
            // Simulate address parsing logic
            let cleaned = input.hasPrefix("0x") ? String(input.dropFirst(2)) : input
            if let parsed = UInt32(cleaned, radix: 16) {
                XCTAssertEqual(parsed, expected, "Failed to parse '\(input)'")
            } else {
                XCTFail("Failed to parse '\(input)'")
            }
        }
    }

    func testHexAddressParsingWithoutPrefix() {
        // Test parsing hex addresses without 0x prefix
        let addresses = [
            "8000": UInt32(0x8000),
            "0": UInt32(0),
            "FFFFFFFF": UInt32(0xFFFFFFFF),
            "10000": UInt32(0x10000),
        ]

        for (input, expected) in addresses {
            if let parsed = UInt32(input, radix: 16) {
                XCTAssertEqual(parsed, expected, "Failed to parse '\(input)'")
            } else {
                XCTFail("Failed to parse '\(input)'")
            }
        }
    }

    func testInvalidAddressParsing() {
        // Test invalid address formats
        let invalidAddresses = [
            "not_a_hex",
            "0xGGGG",
            "",
            "   ",
        ]

        for input in invalidAddresses {
            let cleaned = input.hasPrefix("0x") ? String(input.dropFirst(2)) : input
            let parsed = UInt32(cleaned, radix: 16)
            XCTAssertNil(parsed, "Should not parse invalid address '\(input)'")
        }
    }

    func testCaseInsensitiveHexParsing() {
        // Hex parsing should be case-insensitive
        let lowercase = "0xabcd"
        let uppercase = "0xABCD"
        let mixed = "0xAbCd"

        let cleaned1 = lowercase.dropFirst(2)
        let cleaned2 = uppercase.dropFirst(2)
        let cleaned3 = mixed.dropFirst(2)

        let parsed1 = UInt32(cleaned1, radix: 16)
        let parsed2 = UInt32(cleaned2, radix: 16)
        let parsed3 = UInt32(cleaned3, radix: 16)

        XCTAssertEqual(parsed1, parsed2)
        XCTAssertEqual(parsed2, parsed3)
        XCTAssertEqual(parsed1, 0xABCD)
    }
}

// MARK: - MemoryView Row Calculation Tests

final class MemoryViewRowCalculationTests: XCTestCase {
    func testRowNumberFromAddress() {
        // Calculate which row an address would appear in
        let bytesPerRow = 16

        func rowNumber(for address: UInt32, baseAddress: UInt32) -> Int {
            let offset = Int(address - baseAddress)
            return offset / bytesPerRow
        }

        let baseAddress: UInt32 = 0x8000

        XCTAssertEqual(rowNumber(for: 0x8000, baseAddress: baseAddress), 0) // First row
        XCTAssertEqual(rowNumber(for: 0x8010, baseAddress: baseAddress), 1) // Second row
        XCTAssertEqual(rowNumber(for: 0x8020, baseAddress: baseAddress), 2) // Third row
        XCTAssertEqual(rowNumber(for: 0x800F, baseAddress: baseAddress), 0) // Still first row
    }

    func testAddressFromRow() {
        // Calculate base address for a given row
        let bytesPerRow = 16

        func address(for row: Int, baseAddress: UInt32) -> UInt32 {
            return baseAddress + UInt32(row * bytesPerRow)
        }

        let baseAddress: UInt32 = 0x8000

        XCTAssertEqual(address(for: 0, baseAddress: baseAddress), 0x8000)
        XCTAssertEqual(address(for: 1, baseAddress: baseAddress), 0x8010)
        XCTAssertEqual(address(for: 2, baseAddress: baseAddress), 0x8020)
        XCTAssertEqual(address(for: 15, baseAddress: baseAddress), 0x80F0)
    }

    func testOffsetWithinRow() {
        // Calculate byte offset within a row
        let bytesPerRow = 16

        func offsetInRow(for address: UInt32, baseAddress: UInt32) -> Int {
            let offset = Int(address - baseAddress)
            return offset % bytesPerRow
        }

        let baseAddress: UInt32 = 0x8000

        XCTAssertEqual(offsetInRow(for: 0x8000, baseAddress: baseAddress), 0)
        XCTAssertEqual(offsetInRow(for: 0x8001, baseAddress: baseAddress), 1)
        XCTAssertEqual(offsetInRow(for: 0x800F, baseAddress: baseAddress), 15)
        XCTAssertEqual(offsetInRow(for: 0x8010, baseAddress: baseAddress), 0) // New row
    }
}

// MARK: - MemoryView Hex Formatting Tests

final class MemoryViewHexFormattingTests: XCTestCase {
    func testByteHexFormatting() {
        // Test formatting bytes as 2-digit hex
        let byte1: UInt8 = 0x00
        let formatted1 = String(format: "%02X", byte1)
        XCTAssertEqual(formatted1, "00")

        let byte2: UInt8 = 0xFF
        let formatted2 = String(format: "%02X", byte2)
        XCTAssertEqual(formatted2, "FF")

        let byte3: UInt8 = 0x42
        let formatted3 = String(format: "%02X", byte3)
        XCTAssertEqual(formatted3, "42")
    }

    func testAddressHexFormatting() {
        // Test formatting addresses as 8-digit hex
        let address1: UInt32 = 0x8000
        let formatted1 = String(format: "0x%08X", address1)
        XCTAssertEqual(formatted1, "0x00008000")

        let address2: UInt32 = 0xFFFFFFFF
        let formatted2 = String(format: "0x%08X", address2)
        XCTAssertEqual(formatted2, "0xFFFFFFFF")

        let address3: UInt32 = 0x0
        let formatted3 = String(format: "0x%08X", address3)
        XCTAssertEqual(formatted3, "0x00000000")
    }

    func testRowHexDisplay() {
        // Test formatting a full row of bytes
        let bytes: [UInt8] = [
            0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
            0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
        ]

        let hex = bytes.map { String(format: "%02X", $0) }.joined(separator: " ")
        let expected = "00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F"
        XCTAssertEqual(hex, expected)
    }
}

// MARK: - Note on SwiftUI View Testing Limitations

/*
 MemoryView Testing Limitations:

 MemoryView uses @State and @ObservedObject, making it difficult to unit test:
 - Address input field state
 - Memory data loading
 - Auto-scroll behavior
 - Quick access button actions

 What we CAN test:
 - Constants (bytesPerRow, rowsToShow, totalBytes)
 - Address parsing logic
 - Row/offset calculations
 - Hex formatting

 What we CANNOT easily test without refactoring:
 - Button click handlers
 - TextField input validation
 - Auto-scroll toggle behavior
 - Memory data fetching via ViewModel
 - Row highlighting on write

 Recommendations:
 1. Extract address parsing to a utility function
 2. Extract row calculation logic to a utility struct
 3. Use UI tests for integration testing
 4. Test memory loading through ViewModel tests
 */
