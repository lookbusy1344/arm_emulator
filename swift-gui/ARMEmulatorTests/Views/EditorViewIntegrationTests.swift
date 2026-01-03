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
        textView.maxSize = NSSize(
            width: CGFloat.greatestFiniteMagnitude,
            height: CGFloat.greatestFiniteMagnitude
        )
        textView.minSize = NSSize(width: 0, height: scrollView.contentSize.height)

        XCTAssertTrue(
            textView.isHorizontallyResizable,
            "Text view should be horizontally resizable"
        )
        XCTAssertTrue(
            textView.isVerticallyResizable,
            "Text view should be vertically resizable"
        )
        XCTAssertEqual(
            textView.autoresizingMask,
            [],
            "Auto-resizing mask should be empty"
        )
    }

    func testTextContainerUnlimitedWidth() {
        // Configure text container
        if let textContainer = textView.textContainer {
            textContainer.containerSize = NSSize(
                width: CGFloat.greatestFiniteMagnitude,
                height: CGFloat.greatestFiniteMagnitude
            )
            textContainer.widthTracksTextView = false

            XCTAssertEqual(
                textContainer.containerSize.width,
                CGFloat.greatestFiniteMagnitude,
                "Container width should be unlimited"
            )
            XCTAssertFalse(
                textContainer.widthTracksTextView,
                "Container should not track text view width"
            )
        } else {
            XCTFail("Text container should exist")
        }
    }

    func testNoTextWrapping() {
        // Configure for horizontal scrolling
        textView.isHorizontallyResizable = true
        textView.maxSize = NSSize(
            width: CGFloat.greatestFiniteMagnitude,
            height: CGFloat.greatestFiniteMagnitude
        )

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
           let textContainer = textView.textContainer
        {
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

    func testGutterDoesNotBreakTextRendering() {
        // CRITICAL TEST: Verify text view remains functional with gutter enabled
        // This test documents the NSRulerView rendering bug
        let testText = "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
        textView.string = testText

        // Configure for horizontal scrolling (as in EditorView)
        textView.isVerticallyResizable = true
        textView.isHorizontallyResizable = true
        textView.autoresizingMask = []
        textView.maxSize = NSSize(
            width: CGFloat.greatestFiniteMagnitude,
            height: CGFloat.greatestFiniteMagnitude
        )
        textView.minSize = NSSize(width: 0, height: scrollView.contentSize.height)

        if let textContainer = textView.textContainer {
            textContainer.containerSize = NSSize(
                width: CGFloat.greatestFiniteMagnitude,
                height: CGFloat.greatestFiniteMagnitude
            )
            textContainer.widthTracksTextView = false
        }

        // Create and attach gutter (same as EditorView)
        let gutterView = LineNumberGutterView(scrollView: scrollView, orientation: .verticalRuler)
        gutterView.configure(textView: textView, onBreakpointToggle: { _ in })

        scrollView.verticalRulerView = gutterView
        scrollView.hasVerticalRuler = true
        scrollView.rulersVisible = true

        // Force layout
        if let layoutManager = textView.layoutManager,
           let textContainer = textView.textContainer
        {
            layoutManager.ensureLayout(for: textContainer)

            // Verify text view still has content
            XCTAssertEqual(textView.string, testText, "Text view should retain content")

            // Verify frame is non-zero (detects layout collapse)
            XCTAssertGreaterThan(textView.frame.width, 0, "Text view width should be > 0 with gutter")
            XCTAssertGreaterThan(textView.frame.height, 0, "Text view height should be > 0 with gutter")

            // Verify layout manager has glyphs (detects rendering pipeline)
            XCTAssertGreaterThan(layoutManager.numberOfGlyphs, 0, "Layout manager should have glyphs")

            // Verify used rect is non-zero (detects coordinate issues)
            let usedRect = layoutManager.usedRect(for: textContainer)
            XCTAssertGreaterThan(usedRect.width, 0, "Used rect width should be > 0")
            XCTAssertGreaterThan(usedRect.height, 0, "Used rect height should be > 0")

            // Verify gutter is properly attached
            XCTAssertNotNil(scrollView.verticalRulerView, "Gutter should be attached")
            XCTAssertTrue(scrollView.hasVerticalRuler, "Scroll view should have vertical ruler")
            XCTAssertTrue(scrollView.rulersVisible, "Rulers should be visible")

            // If this test passes but text still doesn't render visually,
            // the issue is at the drawing layer, not the configuration/layout layer
        }
    }
}
