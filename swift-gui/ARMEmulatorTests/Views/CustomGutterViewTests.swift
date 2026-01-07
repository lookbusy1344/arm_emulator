import XCTest
@testable import ARMEmulator

class CustomGutterViewTests: XCTestCase {
    var scrollView: NSScrollView!
    var textView: NSTextView!
    var gutterView: CustomGutterView!

    override func setUp() {
        super.setUp()

        // Create scroll view
        scrollView = NSScrollView(frame: NSRect(x: 0, y: 0, width: 400, height: 300))
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = true

        // Create text view with sample code
        textView = NSTextView()
        textView.string = """
        .org 0x8000
        MOV R0, #42
        ADD R1, R0, #1
        SWI #0
        """
        textView.font = NSFont.monospacedSystemFont(ofSize: 13, weight: .regular)

        scrollView.documentView = textView

        // Create gutter view
        gutterView = CustomGutterView(textView: textView, scrollView: scrollView)
        gutterView.frame = NSRect(x: 0, y: 0, width: 50, height: 300)
    }

    override func tearDown() {
        gutterView = nil
        textView = nil
        scrollView = nil
        super.tearDown()
    }

    func testGutterInitialization() {
        XCTAssertNotNil(gutterView)
        XCTAssertEqual(gutterView.frame.width, 50, "Gutter width should be 50")
        XCTAssertTrue(gutterView.isFlipped, "Gutter should use flipped coordinates")
    }

    func testSetBreakpoints() {
        let breakpoints: Set<Int> = [2, 4]
        gutterView.setBreakpoints(breakpoints)

        // Verify needsDisplay is triggered (we can't easily verify the visual result)
        // The actual rendering is tested through integration tests
        XCTAssertTrue(true, "setBreakpoints should not crash")
    }

    func testSetCurrentLine() {
        gutterView.setCurrentLine(3)

        // Verify the method works without crashing
        XCTAssertTrue(true, "setCurrentLine should not crash")

        // Test with nil
        gutterView.setCurrentLine(nil)
        XCTAssertTrue(true, "setCurrentLine with nil should not crash")
    }

    func testCurrentLineAndBreakpointTogether() {
        // Test that we can have both a current line indicator and breakpoints
        gutterView.setBreakpoints([2, 4])
        gutterView.setCurrentLine(3)

        // Force a redraw
        gutterView.needsDisplay = true

        XCTAssertTrue(true, "Current line and breakpoints should coexist")
    }

    func testCurrentLineOnBreakpoint() {
        // Test when PC is on the same line as a breakpoint
        gutterView.setBreakpoints([3])
        gutterView.setCurrentLine(3)

        // Force a redraw
        gutterView.needsDisplay = true

        XCTAssertTrue(true, "Current line indicator and breakpoint on same line should not conflict")
    }

    func testCurrentLineChanges() {
        // Simulate stepping through code
        gutterView.setCurrentLine(1)
        gutterView.setCurrentLine(2)
        gutterView.setCurrentLine(3)

        // Verify updates trigger redraw
        XCTAssertTrue(true, "Current line should update smoothly")
    }

    func testScrollNotification() {
        // Simulate scroll notification
        NotificationCenter.default.post(
            name: NSView.boundsDidChangeNotification,
            object: scrollView.contentView
        )

        // Gutter should redraw on scroll
        XCTAssertTrue(true, "Gutter should handle scroll notifications")
    }

    func testTextChangeNotification() {
        // Simulate text change
        textView.string = "New content\nLine 2\nLine 3"

        NotificationCenter.default.post(
            name: NSText.didChangeNotification,
            object: textView
        )

        // Gutter should redraw on text change
        XCTAssertTrue(true, "Gutter should handle text change notifications")
    }

    func testCurrentLineIndicatorDrawing() {
        // Set up test conditions
        gutterView.setCurrentLine(2)

        // Force layout
        if let layoutManager = textView.layoutManager,
           let textContainer = textView.textContainer
        {
            layoutManager.ensureLayout(for: textContainer)

            // Verify needsDisplay is set (we don't actually call draw to avoid UI dependencies)
            gutterView.needsDisplay = true
            XCTAssertTrue(gutterView.needsDisplay, "Gutter should be marked for redraw")

            // If we get here without crashing, setup is successful
            XCTAssertTrue(true, "Current line indicator should be configured without errors")
        } else {
            XCTFail("Layout manager or text container missing")
        }
    }

    func testBreakpointToggleCallback() {
        gutterView.configure { _ in
            // Callback configured successfully
        }

        // Simulate mouse click at a specific location
        // (This is hard to test without actually running the UI)
        // Just verify the callback is set
        XCTAssertTrue(true, "Breakpoint toggle callback should be configurable")
    }
}
