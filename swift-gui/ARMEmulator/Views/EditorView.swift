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
            // For now, we'll keep them separate until we have proper address mapping
            self.breakpoints = Set(newBreakpoints.map { Int($0) })
        }
    }

    private func toggleBreakpoint(at lineNumber: Int) {
        if breakpoints.contains(lineNumber) {
            breakpoints.remove(lineNumber)
        } else {
            breakpoints.insert(lineNumber)
        }

        // Convert line number to address
        // For now, use a simple heuristic: assume code starts at 0x8000
        // and each instruction is 4 bytes
        // TODO: Get actual line-to-address mapping from backend
        let address = UInt32(0x8000 + (lineNumber - 1) * 4)

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

    func makeNSView(context: Context) -> NSScrollView {
        let scrollView = NSScrollView()
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = true
        scrollView.autohidesScrollers = true

        let textView = NSTextView()
        textView.isEditable = true
        textView.isSelectable = true
        textView.font = NSFont.monospacedSystemFont(ofSize: 13, weight: .regular)
        textView.textColor = NSColor.labelColor // Ensure text is visible
        textView.backgroundColor = NSColor.textBackgroundColor
        textView.textContainerInset = NSSize(width: 5, height: 5)
        textView.isAutomaticQuoteSubstitutionEnabled = false
        textView.isAutomaticDashSubstitutionEnabled = false
        textView.isAutomaticTextReplacementEnabled = false
        textView.delegate = context.coordinator

        // Configure text view sizing for scroll view (standard wrapping)
        textView.minSize = NSSize(width: 0.0, height: scrollView.contentSize.height)
        textView.maxSize = NSSize(width: CGFloat.greatestFiniteMagnitude, height: CGFloat.greatestFiniteMagnitude)
        textView.isVerticallyResizable = true
        textView.isHorizontallyResizable = false
        textView.autoresizingMask = [.width]

        // Configure text container for wrapping text
        if let textContainer = textView.textContainer {
            textContainer.containerSize = NSSize(
                width: scrollView.contentSize.width,
                height: CGFloat.greatestFiniteMagnitude
            )
            textContainer.widthTracksTextView = true
        }

        scrollView.documentView = textView

        // Create and add gutter view
        // let gutterView = LineNumberGutterView(scrollView: scrollView, orientation: .verticalRuler)
        // gutterView.configure(textView: textView, onBreakpointToggle: onBreakpointToggle)

        // Add gutter as a ruler view
        // scrollView.verticalRulerView = gutterView
        // scrollView.hasVerticalRuler = true
        // scrollView.rulersVisible = true

        // Notify parent that text view was created
        DispatchQueue.main.async {
            onTextViewCreated(textView)
        }

        return scrollView
    }

    func updateNSView(_ scrollView: NSScrollView, context: Context) {
        guard let textView = scrollView.documentView as? NSTextView else { return }

        if textView.string != text {
            textView.string = text

            // Force layout and redraw
            if let layoutManager = textView.layoutManager, let textContainer = textView.textContainer {
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
        if let gutterView = scrollView.verticalRulerView as? LineNumberGutterView {
            gutterView.setBreakpoints(breakpoints)
        }
    }

    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }

    class Coordinator: NSObject, NSTextViewDelegate {
        var parent: EditorWithGutterView

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
