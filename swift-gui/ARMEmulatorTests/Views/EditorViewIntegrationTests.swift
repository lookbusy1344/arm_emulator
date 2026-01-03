import XCTest
@testable import ARMEmulator

class EditorViewIntegrationTests: XCTestCase {
    var scrollView: NSScrollView!
    var textView: NSTextView!

    override func setUp() {
        super.setUp()

        // Create scroll view
        scrollView = NSScrollView(frame: NSRect(x: 0, y: 0, width: 400, height: 300))
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = true

        // Create text view
        textView = NSTextView()
        textView.string = "Short line\nThis is a much longer line that should trigger horizontal scrolling\nAnother line"

        scrollView.documentView = textView
    }

    override func tearDown() {
        textView = nil
        scrollView = nil
        super.tearDown()
    }
}
