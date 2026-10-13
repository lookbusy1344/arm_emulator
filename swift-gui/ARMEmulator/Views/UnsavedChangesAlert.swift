import AppKit

/// The Save / Don't Save / Cancel alert shown before unsaved edits are replaced
@MainActor
enum UnsavedChangesAlert {
    static func ask() -> UnsavedChangesChoice {
        let alert = NSAlert()
        alert.messageText = "Do you want to save the changes to this program?"
        alert.informativeText = "Your changes will be lost if you don't save them."
        alert.addButton(withTitle: "Save")
        alert.addButton(withTitle: "Don't Save")
        alert.addButton(withTitle: "Cancel")

        return switch alert.runModal() {
        case .alertFirstButtonReturn: .save
        case .alertSecondButtonReturn: .discard
        default: .cancel
        }
    }
}
