import AppKit
import SwiftUI

class LineNumberGutterView: NSRulerView {
    private var textView: NSTextView?
    private var breakpoints: Set<Int> = []
    private var onBreakpointToggle: ((Int) -> Void)?

    private let gutterWidth: CGFloat = 50
    private let breakpointMargin: CGFloat = 8
    private let breakpointSize: CGFloat = 10

    override init(scrollView: NSScrollView?, orientation: NSRulerView.Orientation) {
        super.init(scrollView: scrollView, orientation: orientation)
        ruleThickness = gutterWidth
        wantsLayer = true
        layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
    }

    required init(coder: NSCoder) {
        super.init(coder: coder)
        ruleThickness = gutterWidth
        wantsLayer = true
        layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
    }

    func configure(textView: NSTextView, onBreakpointToggle: @escaping (Int) -> Void) {
        self.textView = textView
        self.onBreakpointToggle = onBreakpointToggle

        // Register for text change notifications to update line numbers
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(textDidChange),
            name: NSText.didChangeNotification,
            object: textView
        )
    }

    func setBreakpoints(_ breakpoints: Set<Int>) {
        self.breakpoints = breakpoints
        needsDisplay = true
    }

    @objc private func textDidChange(_ notification: Notification) {
        needsDisplay = true
    }

    override func draw(_ dirtyRect: NSRect) {
        super.draw(dirtyRect)

        guard let textView = textView,
              let layoutManager = textView.layoutManager,
              let textContainer = textView.textContainer
        else {
            return
        }

        drawBackground(dirtyRect)
        drawLineNumbers(textView: textView, layoutManager: layoutManager, textContainer: textContainer)
    }

    private func drawBackground(_ dirtyRect: NSRect) {
        NSColor.controlBackgroundColor.setFill()
        dirtyRect.fill()

        NSColor.separatorColor.setStroke()
        let separatorPath = NSBezierPath()
        separatorPath.move(to: NSPoint(x: bounds.width - 0.5, y: bounds.minY))
        separatorPath.line(to: NSPoint(x: bounds.width - 0.5, y: bounds.maxY))
        separatorPath.lineWidth = 1
        separatorPath.stroke()
    }

    private func drawLineNumbers(textView: NSTextView, layoutManager: NSLayoutManager, textContainer: NSTextContainer) {
        let text = textView.string as NSString
        guard text.length > 0 else { return }

        let glyphRange = layoutManager.glyphRange(for: textContainer)
        var lineNumber = 1
        var glyphIndex = glyphRange.location

        let attributes = lineNumberAttributes()

        while glyphIndex < glyphRange.upperBound {
            let characterIndex = layoutManager.characterIndexForGlyph(at: glyphIndex)
            let lineRange = text.lineRange(for: NSRange(location: characterIndex, length: 0))
            let glyphRange = layoutManager.glyphRange(forCharacterRange: lineRange, actualCharacterRange: nil)
            let lineRect = layoutManager.boundingRect(forGlyphRange: glyphRange, in: textContainer)
            let yPos = lineRect.minY - textView.textContainerInset.height + textView.bounds.minY

            drawLineNumber(lineNumber, yPos: yPos, lineHeight: lineRect.height, attributes: attributes)
            drawBreakpointIfNeeded(lineNumber, yPos: yPos, lineHeight: lineRect.height)

            glyphIndex = NSMaxRange(glyphRange)
            lineNumber += 1
        }
    }

    private func lineNumberAttributes() -> [NSAttributedString.Key: Any] {
        let paragraphStyle = NSMutableParagraphStyle()
        paragraphStyle.alignment = .right

        return [
            .font: NSFont.monospacedSystemFont(ofSize: 11, weight: .regular),
            .foregroundColor: NSColor.secondaryLabelColor,
            .paragraphStyle: paragraphStyle,
        ]
    }

    private func drawLineNumber(
        _ lineNumber: Int,
        yPos: CGFloat,
        lineHeight: CGFloat,
        attributes: [NSAttributedString.Key: Any]
    ) {
        let lineNumberString = "\(lineNumber)" as NSString
        let rect = NSRect(x: 5, y: yPos, width: gutterWidth - 20, height: lineHeight)
        lineNumberString.draw(in: rect, withAttributes: attributes)
    }

    private func drawBreakpointIfNeeded(_ lineNumber: Int, yPos: CGFloat, lineHeight: CGFloat) {
        guard breakpoints.contains(lineNumber) else { return }

        let rect = NSRect(
            x: gutterWidth - breakpointMargin - breakpointSize,
            y: yPos + (lineHeight - breakpointSize) / 2,
            width: breakpointSize,
            height: breakpointSize
        )

        let path = NSBezierPath(ovalIn: rect)
        NSColor.systemRed.setFill()
        path.fill()
    }

    override func mouseDown(with event: NSEvent) {
        let location = convert(event.locationInWindow, from: nil)

        guard let textView = textView,
              let layoutManager = textView.layoutManager,
              let textContainer = textView.textContainer
        else {
            return
        }

        // Determine which line was clicked
        let text = textView.string as NSString
        let textLength = text.length

        guard textLength > 0 else { return }

        let glyphRange = layoutManager.glyphRange(for: textContainer)
        var lineNumber = 1
        var glyphIndex = glyphRange.location

        while glyphIndex < glyphRange.upperBound {
            let characterIndex = layoutManager.characterIndexForGlyph(at: glyphIndex)
            let lineRange = text.lineRange(
                for: NSRange(location: characterIndex, length: 0)
            )

            let glyphRange = layoutManager.glyphRange(forCharacterRange: lineRange, actualCharacterRange: nil)
            let lineRect = layoutManager.boundingRect(forGlyphRange: glyphRange, in: textContainer)

            let yPos = lineRect.minY - textView.textContainerInset.height + textView.bounds.minY

            // Check if click is within this line
            if location.y >= yPos, location.y < yPos + lineRect.height {
                onBreakpointToggle?(lineNumber)
                return
            }

            glyphIndex = NSMaxRange(glyphRange)
            lineNumber += 1
        }
    }

    deinit {
        NotificationCenter.default.removeObserver(self)
    }
}

struct LineNumberGutterViewWrapper: NSViewRepresentable {
    let textView: NSTextView
    let breakpoints: Set<Int>
    let onBreakpointToggle: (Int) -> Void

    func makeNSView(context: Context) -> LineNumberGutterView {
        let gutterView = LineNumberGutterView(frame: .zero)
        gutterView.configure(textView: textView, onBreakpointToggle: onBreakpointToggle)
        return gutterView
    }

    func updateNSView(_ nsView: LineNumberGutterView, context: Context) {
        nsView.setBreakpoints(breakpoints)
    }
}
