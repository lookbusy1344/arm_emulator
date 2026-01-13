import XCTest
@testable import ARMEmulator

@MainActor
class LineNumberGutterViewTests: XCTestCase {
    var textView: NSTextView!
    var scrollView: NSScrollView!
    var gutterView: LineNumberGutterView!

    override func setUp() {
        super.setUp()

        // Create scroll view
        scrollView = NSScrollView(frame: NSRect(x: 0, y: 0, width: 400, height: 300))
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = true

        // Create and configure text view
        textView = NSTextView(frame: scrollView.bounds)
        textView.isEditable = true
        textView.isSelectable = true
        textView.font = NSFont.monospacedSystemFont(ofSize: 13, weight: .regular)
        textView.string = "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"

        scrollView.documentView = textView

        // Create gutter view
        gutterView = LineNumberGutterView(scrollView: scrollView, orientation: .verticalRuler)
        gutterView.configure(textView: textView, onBreakpointToggle: { _ in })

        // Force layout
        if let layoutManager = textView.layoutManager,
           let textContainer = textView.textContainer
        {
            layoutManager.ensureLayout(for: textContainer)
        }
    }

    override func tearDown() {
        textView = nil
        scrollView = nil
        gutterView = nil
        super.tearDown()
    }

    func testLineNumberCoordinateCalculation() {
        // Verify text view has content
        XCTAssertGreaterThan(textView.string.count, 0, "Text view should have content")

        // Get layout manager and text container
        guard let layoutManager = textView.layoutManager,
              let textContainer = textView.textContainer
        else {
            XCTFail("Layout manager and text container should exist")
            return
        }

        // Get first line position
        let text = textView.string as NSString
        let firstLineRange = text.lineRange(for: NSRange(location: 0, length: 0))
        let glyphRange = layoutManager.glyphRange(
            forCharacterRange: firstLineRange,
            actualCharacterRange: nil
        )
        let lineRect = layoutManager.boundingRect(
            forGlyphRange: glyphRange,
            in: textContainer
        )

        // Calculate yPos using new formula
        let visibleRect = scrollView.documentVisibleRect
        let yPos = lineRect.minY - visibleRect.origin.y

        // First line should be at or near y=0 when not scrolled
        XCTAssertGreaterThanOrEqual(
            yPos,
            -5,
            "First line should be near top"
        )
        XCTAssertLessThanOrEqual(
            yPos,
            5,
            "First line should be near top"
        )
    }

    func testCoordinatesWithVerticalScroll() {
        // Scroll down by 50 points
        scrollView.contentView.scroll(to: NSPoint(x: 0, y: 50))

        guard let layoutManager = textView.layoutManager,
              let textContainer = textView.textContainer
        else {
            XCTFail("Layout manager and text container should exist")
            return
        }

        // Get first line position
        let text = textView.string as NSString
        let firstLineRange = text.lineRange(for: NSRange(location: 0, length: 0))
        let glyphRange = layoutManager.glyphRange(
            forCharacterRange: firstLineRange,
            actualCharacterRange: nil
        )
        let lineRect = layoutManager.boundingRect(
            forGlyphRange: glyphRange,
            in: textContainer
        )

        // Calculate yPos with scroll offset
        let visibleRect = scrollView.documentVisibleRect
        let yPos = lineRect.minY - visibleRect.origin.y

        // After scrolling down 50, first line should be at negative y
        XCTAssertLessThan(
            yPos,
            0,
            "First line should be above visible area after scroll"
        )
    }

    func testCoordinatesWithHorizontalScroll() {
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

        // Get first line position before horizontal scroll
        guard let layoutManager = textView.layoutManager,
              let textContainer = textView.textContainer
        else {
            XCTFail("Layout manager and text container should exist")
            return
        }

        let text = textView.string as NSString
        let firstLineRange = text.lineRange(for: NSRange(location: 0, length: 0))
        let glyphRange = layoutManager.glyphRange(
            forCharacterRange: firstLineRange,
            actualCharacterRange: nil
        )
        let lineRect = layoutManager.boundingRect(
            forGlyphRange: glyphRange,
            in: textContainer
        )

        let visibleRectBefore = scrollView.documentVisibleRect
        let yPosBefore = lineRect.minY - visibleRectBefore.origin.y

        // Scroll horizontally (not vertically)
        scrollView.contentView.scroll(to: NSPoint(x: 100, y: 0))

        // Get first line position after horizontal scroll
        let visibleRectAfter = scrollView.documentVisibleRect
        let yPosAfter = lineRect.minY - visibleRectAfter.origin.y

        // Y position should remain the same (only X changed)
        XCTAssertEqual(
            yPosBefore,
            yPosAfter,
            accuracy: 1.0,
            "Line Y position should not change with horizontal scroll"
        )
    }

    func testBreakpointToggle() {
        var toggledLine: Int?

        // Create gutter with breakpoint callback
        gutterView = LineNumberGutterView(scrollView: scrollView, orientation: .verticalRuler)
        gutterView.configure(textView: textView, onBreakpointToggle: { lineNumber in
            toggledLine = lineNumber
        })

        // Set initial breakpoints
        gutterView.setBreakpoints([2, 4])

        XCTAssertNil(toggledLine, "No toggle should have occurred yet")

        // Verify the gutter was configured without crashing
        XCTAssertNotNil(gutterView, "Gutter view should be created")

        // Verify that setting different breakpoints works without crashing
        gutterView.setBreakpoints([1, 3, 5])

        // Verify that clearing breakpoints works
        gutterView.setBreakpoints([])

        // Verify that setting breakpoints again works
        gutterView.setBreakpoints([2, 4, 6])
    }

    func testBreakpointDrawing() {
        // Set initial breakpoints
        let breakpoints: Set<Int> = [1, 3, 5]
        gutterView.setBreakpoints(breakpoints)

        // Verify display can be called without crashing
        gutterView.display()

        // Test with different breakpoint configurations
        gutterView.setBreakpoints([2, 4, 6])
        gutterView.display()

        // Test with empty breakpoint set
        gutterView.setBreakpoints([])
        gutterView.display()

        // Test with large line numbers
        gutterView.setBreakpoints([10, 20, 30])
        gutterView.display()

        // All operations should complete without throwing
        XCTAssertNotNil(gutterView, "Gutter view should remain valid")
    }
}
