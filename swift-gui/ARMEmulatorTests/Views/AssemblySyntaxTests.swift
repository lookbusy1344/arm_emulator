import XCTest
@testable import ARMEmulator

final class AssemblySyntaxTests: XCTestCase {
    /// Each token as its text and kind, in order
    private func tokens(_ text: String) -> [String] {
        AssemblySyntax.tokens(in: text).map { token in
            "\((text as NSString).substring(with: token.range))=\(token.kind)"
        }
    }

    func testInstructionLineWithLabelRegistersNumberAndComment() {
        XCTAssertEqual(
            tokens("loop:  MOV R0, #0x10 ; done"),
            ["loop:=label", "MOV=instruction", "R0=register", "#0x10=number", "; done=comment"],
        )
    }

    func testKeywordsIgnoreCaseAndPlainSymbolsStayUncoloured() {
        XCTAssertEqual(tokens("ldr r1, =msg"), ["ldr=instruction", "r1=register"])
    }

    func testCommentMarkersInsideAStringStayInTheString() {
        XCTAssertEqual(
            tokens(#".asciz "a;b@c" @ note"#),
            [".asciz=directive", #""a;b@c"=string"#, "@ note=comment"],
        )
    }

    func testEscapedQuoteDoesNotEndTheString() {
        XCTAssertEqual(tokens(#".ascii "say \"hi\"""#), [".ascii=directive", #""say \"hi\""=string"#])
    }

    func testUnterminatedStringRunsToTheEndOfTheLine() {
        XCTAssertEqual(tokens("\"abc ; x\nMOV"), ["\"abc ; x=string", "MOV=instruction"])
    }

    func testDoubleSlashStartsAComment() {
        XCTAssertEqual(tokens("SWI #0 // exit"), ["SWI=instruction", "#0=number", "// exit=comment"])
    }

    func testConditionAndFlagSuffixesKeepTheInstruction() {
        XCTAssertEqual(
            tokens("BNE x\nADDS r0\nMOVEQ r1\nLDRB r2\nSUBGES r3"),
            [
                "BNE=instruction",
                "ADDS=instruction",
                "r0=register",
                "MOVEQ=instruction",
                "r1=register",
                "LDRB=instruction",
                "r2=register",
                "SUBGES=instruction",
                "r3=register",
            ],
        )
    }

    func testWordsThatOnlyStartWithAMnemonicAreNotInstructions() {
        XCTAssertEqual(tokens("B loopback\nmove addr"), ["B=instruction"])
    }

    func testRegisterNamesAndAliases() {
        XCTAssertEqual(
            tokens("PUSH {sp, lr, pc, r15, r16, cpsr}"),
            ["PUSH=instruction", "sp=register", "lr=register", "pc=register", "r15=register", "cpsr=register"],
        )
    }

    func testNumberForms() {
        XCTAssertEqual(
            tokens("MOV r0, #-5\n.word 0b101, 42, 0xFF"),
            [
                "MOV=instruction",
                "r0=register",
                "#-5=number",
                ".word=directive",
                "0b101=number",
                "42=number",
                "0xFF=number",
            ],
        )
    }

    func testDigitsInsideAnIdentifierAreNotANumber() {
        XCTAssertEqual(tokens("B label2"), ["B=instruction"])
    }

    func testRangesAreUTF16OffsetsAcrossLines() {
        let text = "; é😀\nMOV"
        let instruction = AssemblySyntax.tokens(in: text).last

        XCTAssertEqual(instruction?.range, NSRange(location: 6, length: 3))
    }

    func testTokensInARangeStayInsideIt() {
        let text = "MOV r0, #1\nADD r1, r2"
        let secondLine = NSRange(location: 11, length: 10)

        let found = AssemblySyntax.tokens(in: text, range: secondLine).map(\.range.location)

        XCTAssertEqual(found, [11, 15, 19])
    }
}

final class CPSRFlagItemsTests: XCTestCase {
    func testFlagsInNZCVOrderWithTheirState() {
        let flags = CPSRFlags(n: true, z: false, c: true, v: false)

        XCTAssertEqual(flags.items.map(\.name), ["N", "Z", "C", "V"])
        XCTAssertEqual(flags.items.map(\.isSet), [true, false, true, false])
    }

    func testEachFlagMapsToItsOwnBit() {
        XCTAssertEqual(CPSRFlags(n: false, z: true, c: false, v: false).items.map(\.isSet), [false, true, false, false])
        XCTAssertEqual(CPSRFlags(n: false, z: false, c: false, v: true).items.map(\.isSet), [false, false, false, true])
    }
}
