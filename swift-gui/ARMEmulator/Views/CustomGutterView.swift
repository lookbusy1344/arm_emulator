import AppKit
import SwiftUI

/// Custom gutter view that doesn't use NSRulerView
/// NSRulerView has a rendering bug that causes NSTextView to not display text
class CustomGutterView: NSView {
    private weak var textView: NSTextView?
    private weak var scrollView: NSScrollView?
    private var breakpoints: Set<Int> = []
    private var currentLine: Int?
    private var onBreakpointToggle: ((Int) -> Void)?

    private let gutterWidth: CGFloat = 50
    private let breakpointMargin: CGFloat = 8
    private let breakpointSize: CGFloat = 10

    init(textView: NSTextView, scrollView: NSScrollView) {
        self.textView = textView
        self.scrollView = scrollView
        super.init(frame: .zero)

        wantsLayer = true
        layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor

        // Register for scroll notifications
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(scrollViewDidScroll),
            name: NSView.boundsDidChangeNotification,
            object: scrollView.contentView
        )

        // Register for text changes
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(textDidChange),
            name: NSText.didChangeNotification,
            object: textView
        )
    }

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) not implemented")
    }

    // Use flipped coordinates (origin at top-left) to match text view
    override var isFlipped: Bool {
        return true
    }

    func configure(onBreakpointToggle: @escaping (Int) -> Void) {
        self.onBreakpointToggle = onBreakpointToggle
    }

    func setBreakpoints(_ breakpoints: Set<Int>) {
        self.breakpoints = breakpoints
        needsDisplay = true
    }

    func setCurrentLine(_ currentLine: Int?) {
        self.currentLine = currentLine
        #if DEBUG
            if let line = currentLine {
                DebugLog.log("CustomGutterView: Setting current line to \(line)", category: "CustomGutterView")
            } else {
                DebugLog.log("CustomGutterView: Clearing current line", category: "CustomGutterView")
            }
        #endif
        needsDisplay = true
    }

    @objc private func scrollViewDidScroll(_ notification: Notification) {
        needsDisplay = true
    }

    @objc private func textDidChange(_ notification: Notification) {
        needsDisplay = true
    }

    override func draw(_ dirtyRect: NSRect) {
        super.draw(dirtyRect)

        // Draw background
        NSColor.controlBackgroundColor.setFill()
        bounds.fill()

        // Draw separator line
        NSColor.separatorColor.setStroke()
        let separatorPath = NSBezierPath()
        separatorPath.move(to: NSPoint(x: bounds.width - 0.5, y: bounds.minY))
        separatorPath.line(to: NSPoint(x: bounds.width - 0.5, y: bounds.maxY))
        separatorPath.lineWidth = 1
        separatorPath.stroke()

        // Draw line numbers
        guard let textView = textView,
              let layoutManager = textView.layoutManager,
              let textContainer = textView.textContainer,
              let scrollView = scrollView
        else {
            return
        }

        let text = textView.string as NSString
        guard text.length > 0 else { return }

        let visibleRect = scrollView.documentVisibleRect
        let glyphRange = layoutManager.glyphRange(for: textContainer)
        var lineNumber = 1
        var glyphIndex = glyphRange.location

        let attributes = lineNumberAttributes()

        while glyphIndex < glyphRange.upperBound {
            let characterIndex = layoutManager.characterIndexForGlyph(at: glyphIndex)
            let lineRange = text.lineRange(for: NSRange(location: characterIndex, length: 0))
            let glyphRange = layoutManager.glyphRange(forCharacterRange: lineRange, actualCharacterRange: nil)
            let lineRect = layoutManager.boundingRect(forGlyphRange: glyphRange, in: textContainer)

            // Convert from text container coordinates to gutter view coordinates
            // Account for text view's textContainerInset (5pt top padding)
            let textInset = textView.textContainerInset.height
            let yPos = lineRect.minY - visibleRect.origin.y + textInset

            // Only draw if visible
            if yPos + lineRect.height >= 0, yPos < bounds.height {
                drawCurrentLineIndicatorIfNeeded(lineNumber, yPos: yPos, lineHeight: lineRect.height)
                drawLineNumber(lineNumber, yPos: yPos, lineHeight: lineRect.height, attributes: attributes)
                drawBreakpointIfNeeded(lineNumber, yPos: yPos, lineHeight: lineRect.height)
            }

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

    private func drawCurrentLineIndicatorIfNeeded(_ lineNumber: Int, yPos: CGFloat, lineHeight: CGFloat) {
        guard let currentLine = currentLine, currentLine == lineNumber else { return }

        #if DEBUG
            DebugLog.log(
                "Drawing PC indicator at line \(lineNumber), yPos: \(yPos), lineHeight: \(lineHeight)",
                category: "CustomGutterView"
            )
        #endif

        // Draw arrow pointing to current line
        let arrowSize: CGFloat = 8
        let arrowX = gutterWidth - breakpointMargin - breakpointSize - arrowSize - 2
        let arrowY = yPos + (lineHeight - arrowSize) / 2

        let arrow = NSBezierPath()
        arrow.move(to: NSPoint(x: arrowX, y: arrowY))
        arrow.line(to: NSPoint(x: arrowX + arrowSize, y: arrowY + arrowSize / 2))
        arrow.line(to: NSPoint(x: arrowX, y: arrowY + arrowSize))
        arrow.close()

        NSColor.systemBlue.setFill()
        arrow.fill()
    }

    private func drawBreakpointIfNeeded(_ lineNumber: Int, yPos: CGFloat, lineHeight: CGFloat) {
        guard breakpoints.contains(lineNumber) else { return }

        let rect = NSRect(
            x: gutterWidth - breakpointMargin - breakpointSize + 5,
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
              let textContainer = textView.textContainer,
              let scrollView = scrollView
        else {
            return
        }

        let text = textView.string as NSString
        guard text.length > 0 else { return }

        let visibleRect = scrollView.documentVisibleRect
        let glyphRange = layoutManager.glyphRange(for: textContainer)
        var lineNumber = 1
        var glyphIndex = glyphRange.location

        while glyphIndex < glyphRange.upperBound {
            let characterIndex = layoutManager.characterIndexForGlyph(at: glyphIndex)
            let lineRange = text.lineRange(for: NSRange(location: characterIndex, length: 0))
            let glyphRange = layoutManager.glyphRange(forCharacterRange: lineRange, actualCharacterRange: nil)
            let lineRect = layoutManager.boundingRect(forGlyphRange: glyphRange, in: textContainer)

            // Account for text view's textContainerInset
            let textInset = textView.textContainerInset.height
            let yPos = lineRect.minY - visibleRect.origin.y + textInset

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
