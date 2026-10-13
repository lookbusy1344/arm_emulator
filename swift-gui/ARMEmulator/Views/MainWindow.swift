import AppKit
import SwiftUI

/// Formats the main window title from the open file and its unsaved state
enum WindowTitle {
    private static let applicationName = "ARM Emulator"
    private static let untitledName = "Untitled"
    private static let dirtyMarker = "•"

    /// "ARM Emulator", then " — name" for an open file or unsaved edits, then " •" when there are unsaved edits
    static func format(fileURL: URL?, isDirty: Bool) -> String {
        guard fileURL != nil || isDirty else { return applicationName }
        let name = fileURL?.lastPathComponent ?? untitledName
        return "\(applicationName) — \(name)\(isDirty ? " \(dirtyMarker)" : "")"
    }
}

/// Asks before a window with unsaved edits closes. Stands in as the window delegate and forwards every other
/// delegate call to the delegate it replaced.
@MainActor
final class WindowCloseGuard: NSObject, NSWindowDelegate {
    private let hasUnsavedChanges: () -> Bool
    private let confirm: () async -> Bool
    private weak var original: NSWindowDelegate?

    init(hasUnsavedChanges: @escaping () -> Bool, confirm: @escaping () async -> Bool) {
        self.hasUnsavedChanges = hasUnsavedChanges
        self.confirm = confirm
    }

    /// Makes this guard the delegate of `window`. The window holds its delegate weakly; the caller keeps the guard.
    func install(on window: NSWindow) {
        guard window.delegate !== self else { return }
        original = window.delegate
        window.delegate = self
    }

    func windowShouldClose(_ window: NSWindow) -> Bool {
        guard hasUnsavedChanges() else {
            return original?.windowShouldClose?(window) ?? true
        }
        Task {
            if await confirm() {
                window.close()
            }
        }
        return false
    }

    override func responds(to selector: Selector!) -> Bool {
        super.responds(to: selector) || original?.responds(to: selector) == true
    }

    override func forwardingTarget(for _: Selector!) -> Any? {
        original
    }
}

/// Configures the hosting window: a stable frame autosave name, the close guard, and the edited dot
struct MainWindowSetup: NSViewRepresentable {
    static let frameAutosaveName = "MainWindow"

    let isDocumentEdited: Bool
    let makeCloseGuard: () -> WindowCloseGuard

    func makeNSView(context _: Context) -> NSView {
        WindowObservingView(closeGuard: makeCloseGuard())
    }

    func updateNSView(_ view: NSView, context _: Context) {
        view.window?.isDocumentEdited = isDocumentEdited
    }

    /// Owns the close guard, which the window holds only weakly
    final class WindowObservingView: NSView {
        private let closeGuard: WindowCloseGuard

        init(closeGuard: WindowCloseGuard) {
            self.closeGuard = closeGuard
            super.init(frame: .zero)
        }

        @available(*, unavailable)
        required init?(coder _: NSCoder) {
            fatalError("init(coder:) is not supported")
        }

        override func viewDidMoveToWindow() {
            super.viewDidMoveToWindow()
            guard let window else { return }
            window.setFrameAutosaveName(MainWindowSetup.frameAutosaveName)
            closeGuard.install(on: window)
        }
    }
}
