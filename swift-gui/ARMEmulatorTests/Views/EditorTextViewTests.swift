import AppKit
import XCTest
@testable import ARMEmulator

@MainActor
final class EditorTextViewTests: XCTestCase {
    private var window: NSWindow!
    private var textView: EditorTextView!
    private var reports: [Bool] = []

    override func setUp() async throws {
        try await super.setUp()
        window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 200, height: 200),
            styleMask: [.titled],
            backing: .buffered,
            defer: true,
        )
        window.isReleasedWhenClosed = false
        textView = EditorTextView()
        textView.onFocusChange = { [unowned self] in reports.append($0) }
        window.contentView = textView
        reports = []
    }

    override func tearDown() async throws {
        window.close()
        window = nil
        textView = nil
        try await super.tearDown()
    }

    func testGainingFocusReportsFocused() {
        XCTAssertTrue(window.makeFirstResponder(textView))
        XCTAssertEqual(reports, [true])
    }

    func testLosingFocusReportsUnfocused() {
        window.makeFirstResponder(textView)
        XCTAssertTrue(window.makeFirstResponder(nil))
        XCTAssertEqual(reports, [true, false])
    }

    func testFocusMovingToAnotherViewAndBackReportsEachChange() {
        let container = NSView()
        let field = NSTextField()
        container.addSubview(textView)
        container.addSubview(field)
        window.contentView = container

        window.makeFirstResponder(textView)
        window.makeFirstResponder(field)
        window.makeFirstResponder(textView)
        XCTAssertEqual(reports, [true, false, true])
    }
}
