import AppKit
import SwiftUI

struct EditorView: View {
    @Binding var text: String
    @State private var breakpoints: Set<Int> = []
    @State private var currentLine: Int?
    @State private var textView: NSTextView?
    @State private var scrollView: NSScrollView?
    @EnvironmentObject var viewModel: EmulatorViewModel

    // Compute editor editability based on VM state
    // Editor is editable only when VM is completely stopped (idle, halted, error)
    // Editor is read-only during any form of execution (running, paused, breakpoint, waitingForInput)
    private var isEditable: Bool {
        viewModel.status == .idle || viewModel.status == .halted || viewModel.status == .error
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("Assembly Editor")
                .font(.system(size: 11, weight: .semibold))
                .padding(.horizontal)
                .padding(.vertical, 8)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color(NSColor.controlBackgroundColor))

            EditorWithGutterView(
                text: $text,
                breakpoints: $breakpoints,
                currentLine: $currentLine,
                isEditable: isEditable,
                onBreakpointToggle: { lineNumber in
                    toggleBreakpoint(at: lineNumber)
                },
                onTextViewCreated: { textView in
                    self.textView = textView
                },
                onScrollViewCreated: { scrollView in
                    self.scrollView = scrollView
                }
            )
            .background(Color(NSColor.textBackgroundColor))
            .border(Color.gray.opacity(0.3))
        }
        .onChange(of: viewModel.breakpoints) { _, newBreakpoints in
            // Sync breakpoints from ViewModel (address-based) to line-based
            // Use addressToLine mapping from source map
            self.breakpoints = Set(newBreakpoints.compactMap { address in
                viewModel.addressToLine[address]
            })
        }
        .onChange(of: viewModel.currentPC) { _, newPC in
            // Update current line from PC address
            self.currentLine = viewModel.addressToLine[newPC]
            #if DEBUG
                DebugLog.log(
                    "PC changed to 0x\(String(format: "%08X", newPC)), mapped to line \(self.currentLine?.description ?? "nil")",
                    category: "EditorView"
                )
                DebugLog.log("addressToLine has \(viewModel.addressToLine.count) entries", category: "EditorView")
                if !viewModel.addressToLine.isEmpty {
                    let samples = Array(viewModel.addressToLine.prefix(3))
                    DebugLog.log("Sample mappings: \(samples)", category: "EditorView")
                }
            #endif

            // Scroll to bring PC into view
            scrollToCurrentLine()
        }
        .onAppear {
            // Register scroll callback with view model
            // Capture the necessary state explicitly since we're in a struct
            let lineGetter: () -> Int? = { currentLine }
            let textViewGetter: () -> NSTextView? = { textView }
            let scrollViewGetter: () -> NSScrollView? = { scrollView }

            viewModel.scrollToCurrentPC = {
                if let currentLine = lineGetter(),
                   let textView = textViewGetter(),
                   let scrollView = scrollViewGetter()
                {
                    scrollToLine(currentLine, in: textView, scrollView: scrollView)
                }
            }
        }
    }

    private func scrollToLine(_ lineNumber: Int, in textView: NSTextView, scrollView: NSScrollView) {
        guard let layoutManager = textView.layoutManager,
              let textContainer = textView.textContainer
        else {
            return
        }

        let text = textView.string as NSString
        guard text.length > 0 else { return }

        // Find the character range for the target line
        var currentLine = 1
        var characterIndex = 0

        while characterIndex < text.length, currentLine < lineNumber {
            let lineRange = text.lineRange(for: NSRange(location: characterIndex, length: 0))
            characterIndex = NSMaxRange(lineRange)
            currentLine += 1
        }

        if currentLine == lineNumber, characterIndex < text.length {
            let lineRange = text.lineRange(for: NSRange(location: characterIndex, length: 0))
            let glyphRange = layoutManager.glyphRange(forCharacterRange: lineRange, actualCharacterRange: nil)
            let lineRect = layoutManager.boundingRect(forGlyphRange: glyphRange, in: textContainer)

            // Convert to scroll view coordinates with text inset
            let textInset = textView.textContainerInset.height
            var scrollRect = lineRect
            scrollRect.origin.y += textInset

            // Center the line in the visible area
            let visibleHeight = scrollView.documentVisibleRect.height
            scrollRect.origin.y -= (visibleHeight - lineRect.height) / 2
            scrollRect.size.height = visibleHeight

            textView.scrollToVisible(scrollRect)

            #if DEBUG
                DebugLog.log(
                    "Scrolled to line \(lineNumber), rect: \(scrollRect)",
                    category: "EditorView"
                )
            #endif
        }
    }

    func scrollToCurrentLine() {
        guard let currentLine = currentLine,
              let textView = textView,
              let scrollView = scrollView
        else {
            return
        }

        scrollToLine(currentLine, in: textView, scrollView: scrollView)
    }

    private func toggleBreakpoint(at lineNumber: Int) {
        // Check if this line already has a breakpoint (allow removal)
        if breakpoints.contains(lineNumber) {
            // Get address for this line to remove breakpoint
            if let address = viewModel.lineToAddress[lineNumber] {
                Task {
                    await viewModel.toggleBreakpoint(at: address)
                }
            }
            return
        }

        // Validate line can have a breakpoint (must be executable code)
        guard viewModel.validBreakpointLines.contains(lineNumber) else {
            print("Cannot set breakpoint on line \(lineNumber) - not executable code")
            return
        }

        // Get actual address for this line from the backend-provided mapping
        guard let address = viewModel.lineToAddress[lineNumber] else {
            print("Cannot set breakpoint on line \(lineNumber) - no address mapping found")
            return
        }

        Task {
            await viewModel.toggleBreakpoint(at: address)
        }
    }
}

// Custom text view wrapper for editor
struct EditorWithGutterView: NSViewRepresentable {
    @Binding var text: String
    @Binding var breakpoints: Set<Int>
    @Binding var currentLine: Int?
    let isEditable: Bool
    let onBreakpointToggle: (Int) -> Void
    let onTextViewCreated: (NSTextView) -> Void
    let onScrollViewCreated: (NSScrollView) -> Void

    func makeNSView(context: Context) -> NSView {
        let containerView = NSView()

        let scrollView = configureScrollView()
        let textView = configureTextView(scrollView: scrollView, coordinator: context.coordinator)

        scrollView.documentView = textView

        // Create custom gutter view (not NSRulerView - it breaks text rendering)
        let gutterView = CustomGutterView(textView: textView, scrollView: scrollView)
        gutterView.configure(onBreakpointToggle: onBreakpointToggle)

        // Add gutter and scroll view to container
        containerView.addSubview(gutterView)
        containerView.addSubview(scrollView)

        // Layout gutter and scroll view
        setupLayout(containerView: containerView, gutterView: gutterView, scrollView: scrollView)

        // Store gutter reference for updates
        context.coordinator.gutterView = gutterView

        // Notify parent that text view was created
        DispatchQueue.main.async {
            onTextViewCreated(textView)
            onScrollViewCreated(scrollView)
        }

        return containerView
    }

    private func configureScrollView() -> NSScrollView {
        let scrollView = NSScrollView()
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = true
        scrollView.autohidesScrollers = true
        return scrollView
    }

    private func configureTextView(scrollView: NSScrollView, coordinator: Coordinator) -> NSTextView {
        let textView = NSTextView()
        textView.isEditable = true
        textView.isSelectable = true
        textView.font = NSFont.monospacedSystemFont(ofSize: 10, weight: .regular)
        textView.textColor = NSColor.labelColor
        textView.backgroundColor = NSColor.textBackgroundColor
        textView.textContainerInset = NSSize(width: 5, height: 5)
        textView.isAutomaticQuoteSubstitutionEnabled = false
        textView.isAutomaticDashSubstitutionEnabled = false
        textView.isAutomaticTextReplacementEnabled = false
        textView.delegate = coordinator

        // Configure for horizontal scrolling (no wrapping)
        textView.isVerticallyResizable = true
        textView.isHorizontallyResizable = true
        textView.autoresizingMask = []
        textView.maxSize = NSSize(
            width: CGFloat.greatestFiniteMagnitude,
            height: CGFloat.greatestFiniteMagnitude
        )
        textView.minSize = NSSize(width: 0, height: scrollView.contentSize.height)

        // Configure text container for unlimited width (no wrapping)
        if let textContainer = textView.textContainer {
            textContainer.containerSize = NSSize(
                width: CGFloat.greatestFiniteMagnitude,
                height: CGFloat.greatestFiniteMagnitude
            )
            textContainer.widthTracksTextView = false
        }

        return textView
    }

    private func setupLayout(containerView: NSView, gutterView: CustomGutterView, scrollView: NSScrollView) {
        gutterView.translatesAutoresizingMaskIntoConstraints = false
        scrollView.translatesAutoresizingMaskIntoConstraints = false

        NSLayoutConstraint.activate([
            // Gutter on the left
            gutterView.leadingAnchor.constraint(equalTo: containerView.leadingAnchor),
            gutterView.topAnchor.constraint(equalTo: containerView.topAnchor),
            gutterView.bottomAnchor.constraint(equalTo: containerView.bottomAnchor),
            gutterView.widthAnchor.constraint(equalToConstant: 50),

            // Scroll view fills remaining space
            scrollView.leadingAnchor.constraint(equalTo: gutterView.trailingAnchor),
            scrollView.trailingAnchor.constraint(equalTo: containerView.trailingAnchor),
            scrollView.topAnchor.constraint(equalTo: containerView.topAnchor),
            scrollView.bottomAnchor.constraint(equalTo: containerView.bottomAnchor),
        ])
    }

    func updateNSView(_ containerView: NSView, context: Context) {
        // Find scroll view and text view
        guard let scrollView = containerView.subviews.first(where: { $0 is NSScrollView }) as? NSScrollView,
              let textView = scrollView.documentView as? NSTextView
        else {
            return
        }

        // Update editability based on VM state
        textView.isEditable = isEditable

        if textView.string != text {
            textView.string = text

            // Force layout and redraw
            if let layoutManager = textView.layoutManager,
               let textContainer = textView.textContainer
            {
                layoutManager.invalidateLayout(
                    forCharacterRange: NSRange(location: 0, length: text.count),
                    actualCharacterRange: nil
                )
                layoutManager.ensureLayout(for: textContainer)
            }

            textView.sizeToFit()
            textView.needsDisplay = true
            scrollView.needsDisplay = true
        }

        // Update gutter breakpoints and current line
        context.coordinator.gutterView?.setBreakpoints(breakpoints)
        context.coordinator.gutterView?.setCurrentLine(currentLine)
    }

    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }

    class Coordinator: NSObject, NSTextViewDelegate {
        var parent: EditorWithGutterView
        var gutterView: CustomGutterView?

        init(_ parent: EditorWithGutterView) {
            self.parent = parent
        }

        func textDidChange(_ notification: Notification) {
            guard let textView = notification.object as? NSTextView else { return }
            parent.text = textView.string
        }
    }
}

struct EditorView_Previews: PreviewProvider {
    static var previews: some View {
        EditorView(text: .constant("""
        .org 0x8000
        MOV R0, #42
        SWI #0
        """))
        .environmentObject(EmulatorViewModel())
        .frame(width: 400, height: 300)
    }
}
