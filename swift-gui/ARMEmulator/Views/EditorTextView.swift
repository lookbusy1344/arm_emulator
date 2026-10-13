import AppKit

/// The editor's text view. Reports gaining and losing first responder so SwiftUI can show the pane's focus cue.
final class EditorTextView: NSTextView {
    var onFocusChange: (Bool) -> Void = { _ in }

    override func becomeFirstResponder() -> Bool {
        let became = super.becomeFirstResponder()
        if became {
            onFocusChange(true)
        }
        return became
    }

    override func resignFirstResponder() -> Bool {
        let resigned = super.resignFirstResponder()
        if resigned {
            onFocusChange(false)
        }
        return resigned
    }
}
