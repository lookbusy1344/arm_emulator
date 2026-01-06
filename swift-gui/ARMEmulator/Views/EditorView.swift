import AppKit
import SwiftUI

struct EditorView: View {
    @Binding var text: String
    @State private var breakpoints: Set<Int> = []
    @State private var textView: NSTextView?
    @EnvironmentObject var viewModel: EmulatorViewModel

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            Text("Assembly Editor")
                .font(.headline)
                .padding(.horizontal)
                .padding(.vertical, 8)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color(NSColor.controlBackgroundColor))

            EditorWithGutterView(
                text: $text,
                breakpoints: $breakpoints,
                onBreakpointToggle: { lineNumber in
                    toggleBreakpoint(at: lineNumber)
                },
                onTextViewCreated: { textView in
                    self.textView = textView
                }
            )
            .background(Color(NSColor.textBackgroundColor))
            .border(Color.gray.opacity(0.3))
        }
        .onChange(of: viewModel.breakpoints) { newBreakpoints in
            // Sync breakpoints from ViewModel (address-based) to line-based
            // Convert addresses back to line numbers using same heuristic
            self.breakpoints = Set(newBreakpoints.map { address in
                let offset = Int(address) - 0x8000
                let lineNumber = (offset / 4) + 1
                return lineNumber
            })
        }
    }

    private func toggleBreakpoint(at lineNumber: Int) {
        // Convert line number to address
        // For now, use a simple heuristic: assume code starts at 0x8000
        // and each instruction is 4 bytes
        // TODO: Get actual line-to-address mapping from backend
        let address = UInt32(0x8000 + (lineNumber - 1) * 4)
        
        // Validate address exists in source map before attempting to set breakpoint
        guard viewModel.sourceMap[address] != nil || breakpoints.contains(lineNumber) else {
            // Address not in source map - invalid location for breakpoint
            print("Cannot set breakpoint on line \(lineNumber) - address 0x\(String(format: "%X", address)) is not executable code")
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
    let onBreakpointToggle: (Int) -> Void
    let onTextViewCreated: (NSTextView) -> Void

    func makeNSView(context: Context) -> NSView {
        let containerView = NSView()

        let scrollView = NSScrollView()
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = true
        scrollView.autohidesScrollers = true

        let textView = NSTextView()
        textView.isEditable = true
        textView.isSelectable = true
        textView.font = NSFont.monospacedSystemFont(ofSize: 13, weight: .regular)
        textView.textColor = NSColor.labelColor
        textView.backgroundColor = NSColor.textBackgroundColor
        textView.textContainerInset = NSSize(width: 5, height: 5)
        textView.isAutomaticQuoteSubstitutionEnabled = false
        textView.isAutomaticDashSubstitutionEnabled = false
        textView.isAutomaticTextReplacementEnabled = false
        textView.delegate = context.coordinator

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

        scrollView.documentView = textView

        // Create custom gutter view (not NSRulerView - it breaks text rendering)
        let gutterView = CustomGutterView(textView: textView, scrollView: scrollView)
        gutterView.configure(onBreakpointToggle: onBreakpointToggle)

        // Add gutter and scroll view to container
        containerView.addSubview(gutterView)
        containerView.addSubview(scrollView)

        // Layout gutter and scroll view
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

        // Store gutter reference for updates
        context.coordinator.gutterView = gutterView

        // Notify parent that text view was created
        DispatchQueue.main.async {
            onTextViewCreated(textView)
        }

        return containerView
    }

    func updateNSView(_ containerView: NSView, context: Context) {
        // Find scroll view and text view
        guard let scrollView = containerView.subviews.first(where: { $0 is NSScrollView }) as? NSScrollView,
              let textView = scrollView.documentView as? NSTextView
        else {
            return
        }

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

        // Update gutter breakpoints
        context.coordinator.gutterView?.setBreakpoints(breakpoints)
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
        .frame(width: 400, height: 300)
    }
}
