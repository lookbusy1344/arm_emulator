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

    static let gutterWidth: CGFloat = 64
    private var gutterWidth: CGFloat {
        Self.gutterWidth
    }

    private let arrowMargin: CGFloat = 4
    private let arrowSize: CGFloat = 8
    private let lineNumberMinX: CGFloat = 14
    private let lineNumberTrailingInset: CGFloat = 16
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
            object: scrollView.contentView,
        )

        // Register for text changes
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(textDidChange),
            name: NSText.didChangeNotification,
            object: textView,
        )
    }

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) not implemented")
    }

    /// Use flipped coordinates (origin at top-left) to match text view
    override var isFlipped: Bool {
        true
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

        let attributes = lineNumberAttributes()

        for line in gutterLines() where line.minY + line.height >= 0 && line.minY < bounds.height {
            drawCurrentLineIndicatorIfNeeded(line.number, yPos: line.minY, lineHeight: line.height)
            drawLineNumber(line.number, yPos: line.minY, lineHeight: line.height, attributes: attributes)
            drawBreakpointIfNeeded(line.number, yPos: line.minY, lineHeight: line.height)
        }
    }

    /// One row per logical line, in gutter coordinates.
    /// `height` is the first visual fragment, where the number and indicators sit.
    /// `hitHeight` spans every fragment of a wrapped line.
    struct GutterLine: Equatable {
        let number: Int
        let minY: CGFloat
        let height: CGFloat
        let hitHeight: CGFloat
    }

    func gutterLines() -> [GutterLine] {
        guard let textView,
              let layoutManager = textView.layoutManager,
              let textContainer = textView.textContainer,
              let scrollView
        else {
            return []
        }

        layoutManager.ensureLayout(for: textContainer)
        let text = textView.string as NSString
        let offsetY = textView.textContainerInset.height - scrollView.documentVisibleRect.origin.y

        var lines: [GutterLine] = []
        var glyphIndex = 0
        while glyphIndex < layoutManager.numberOfGlyphs {
            let characterIndex = layoutManager.characterIndexForGlyph(at: glyphIndex)
            let lineRange = text.lineRange(for: NSRange(location: characterIndex, length: 0))
            let lineGlyphRange = layoutManager.glyphRange(forCharacterRange: lineRange, actualCharacterRange: nil)
            let firstFragment = layoutManager.lineFragmentRect(forGlyphAt: glyphIndex, effectiveRange: nil)
            let lastFragment = layoutManager.lineFragmentRect(
                forGlyphAt: NSMaxRange(lineGlyphRange) - 1,
                effectiveRange: nil,
            )

            lines.append(GutterLine(
                number: lines.count + 1,
                minY: firstFragment.minY + offsetY,
                height: firstFragment.height,
                hitHeight: lastFragment.maxY - firstFragment.minY,
            ))
            glyphIndex = NSMaxRange(lineGlyphRange)
        }

        // The empty line after a trailing newline (or the only line of empty text) has no glyphs.
        let extra = layoutManager.extraLineFragmentRect
        if !extra.isEmpty {
            lines.append(GutterLine(
                number: lines.count + 1,
                minY: extra.minY + offsetY,
                height: extra.height,
                hitHeight: extra.height,
            ))
        }
        return lines
    }

    func lineNumber(atY y: CGFloat) -> Int? {
        gutterLines().first { y >= $0.minY && y < $0.minY + $0.hitHeight }?.number
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
        attributes: [NSAttributedString.Key: Any],
    ) {
        let lineNumberString = "\(lineNumber)" as NSString
        let rect = NSRect(
            x: lineNumberMinX,
            y: yPos,
            width: gutterWidth - lineNumberMinX - lineNumberTrailingInset,
            height: lineHeight,
        )
        lineNumberString.draw(in: rect, withAttributes: attributes)
    }

    private func drawCurrentLineIndicatorIfNeeded(_ lineNumber: Int, yPos: CGFloat, lineHeight: CGFloat) {
        guard let currentLine, currentLine == lineNumber else { return }

        #if DEBUG
            DebugLog.log(
                "Drawing PC indicator at line \(lineNumber), yPos: \(yPos), lineHeight: \(lineHeight)",
                category: "CustomGutterView",
            )
        #endif

        // Draw arrow pointing to current line
        let arrowX = arrowMargin
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
            height: breakpointSize,
        )

        let path = NSBezierPath(ovalIn: rect)
        NSColor.systemRed.setFill()
        path.fill()
    }

    override func mouseDown(with event: NSEvent) {
        let location = convert(event.locationInWindow, from: nil)
        if let number = lineNumber(atY: location.y) {
            onBreakpointToggle?(number)
        }
    }

    deinit {
        NotificationCenter.default.removeObserver(self)
    }
}
