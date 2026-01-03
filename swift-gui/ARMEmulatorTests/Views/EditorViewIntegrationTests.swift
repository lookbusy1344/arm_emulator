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

    func testTextViewHorizontalScrollingEnabled() {
        // Configure for horizontal scrolling
        textView.isVerticallyResizable = true
        textView.isHorizontallyResizable = true
        textView.autoresizingMask = []
        textView.maxSize = NSSize(width: CGFloat.greatestFiniteMagnitude,
                                  height: CGFloat.greatestFiniteMagnitude)
        textView.minSize = NSSize(width: 0, height: scrollView.contentSize.height)

        XCTAssertTrue(textView.isHorizontallyResizable,
                      "Text view should be horizontally resizable")
        XCTAssertTrue(textView.isVerticallyResizable,
                      "Text view should be vertically resizable")
        XCTAssertEqual(textView.autoresizingMask, [],
                       "Auto-resizing mask should be empty")
    }

    func testTextContainerUnlimitedWidth() {
        // Configure text container
        if let textContainer = textView.textContainer {
            textContainer.containerSize = NSSize(
                width: CGFloat.greatestFiniteMagnitude,
                height: CGFloat.greatestFiniteMagnitude
            )
            textContainer.widthTracksTextView = false

            XCTAssertEqual(textContainer.containerSize.width,
                           CGFloat.greatestFiniteMagnitude,
                           "Container width should be unlimited")
            XCTAssertFalse(textContainer.widthTracksTextView,
                           "Container should not track text view width")
        } else {
            XCTFail("Text container should exist")
        }
    }

    func testNoTextWrapping() {
        // Configure for horizontal scrolling
        textView.isHorizontallyResizable = true
        textView.maxSize = NSSize(width: CGFloat.greatestFiniteMagnitude,
                                  height: CGFloat.greatestFiniteMagnitude)

        if let textContainer = textView.textContainer {
            textContainer.containerSize = NSSize(
                width: CGFloat.greatestFiniteMagnitude,
                height: CGFloat.greatestFiniteMagnitude
            )
            textContainer.widthTracksTextView = false
        }

        // Set long text that would wrap if wrapping enabled
        let longLine = String(repeating: "x", count: 200)
        textView.string = longLine

        // Force layout
        if let layoutManager = textView.layoutManager,
           let textContainer = textView.textContainer {
            layoutManager.ensureLayout(for: textContainer)

            // Count number of lines (should be 1 - no wrapping)
            var lineCount = 0
            let text = textView.string as NSString
            var index = 0

            while index < text.length {
                let lineRange = text.lineRange(for: NSRange(location: index, length: 0))
                lineCount += 1
                index = NSMaxRange(lineRange)
            }

            XCTAssertEqual(lineCount, 1, "Long line should not wrap")
        }
    }
}
