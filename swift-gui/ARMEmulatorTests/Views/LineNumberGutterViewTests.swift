import XCTest
@testable import ARMEmulator

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
           let textContainer = textView.textContainer {
            layoutManager.ensureLayout(for: textContainer)
        }
    }

    override func tearDown() {
        textView = nil
        scrollView = nil
        gutterView = nil
        super.tearDown()
    }
}
