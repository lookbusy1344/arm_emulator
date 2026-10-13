import AppKit
import XCTest
@testable import ARMEmulator

@MainActor
final class EditorFontTests: XCTestCase {
    func testApplyFontSetsMonospacedFontOfTheGivenSize() {
        let textView = NSTextView()

        EditorWithGutterView.applyFont(size: 18, to: textView)

        XCTAssertEqual(textView.font?.pointSize, 18)
        XCTAssertTrue(textView.font?.isFixedPitch ?? false, "Editor font should be monospaced")
    }

    func testApplyFontChangesAnExistingFont() {
        let textView = NSTextView()
        EditorWithGutterView.applyFont(size: 12, to: textView)

        EditorWithGutterView.applyFont(size: 22, to: textView)

        XCTAssertEqual(textView.font?.pointSize, 22)
    }

    func testApplyFontRestylesExistingText() throws {
        let textView = NSTextView()
        textView.string = "MOV R0, #1\nSWI #0"

        EditorWithGutterView.applyFont(size: 16, to: textView)

        let storage = try XCTUnwrap(textView.textStorage)
        let lastCharacterFont = storage.attribute(.font, at: storage.length - 1, effectiveRange: nil) as? NSFont
        XCTAssertEqual(lastCharacterFont?.pointSize, 16, "Existing text should take the new size")
    }
}
