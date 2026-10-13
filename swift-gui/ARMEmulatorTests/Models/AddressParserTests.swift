import XCTest
@testable import ARMEmulator

final class AddressParserTests: XCTestCase {
    func testRejectsSignsInsideTheNumber() {
        XCTAssertNil(AddressParser.parse("+42"))
        XCTAssertNil(AddressParser.parse("0x+FF"))
        XCTAssertNil(AddressParser.parse("0x-1"))
    }

    func testRejectsValuesAbove32Bits() {
        XCTAssertNil(AddressParser.parse("4294967296"))
        XCTAssertNil(AddressParser.parse("0x100000000"))
    }

    func testRejectsEmbeddedSpaceAndSeparators() {
        XCTAssertNil(AddressParser.parse("1 2"))
        XCTAssertNil(AddressParser.parse("0x 10"))
        XCTAssertNil(AddressParser.parse("1,000"))
        XCTAssertNil(AddressParser.parse("9000h"))
    }
}
