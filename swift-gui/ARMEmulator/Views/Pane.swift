import SwiftUI

/// The regions of the main window that F6 and Shift+F6 move focus between, in order
enum Pane: CaseIterable {
    case editor
    case inspector
    case console

    /// The pane after (or before) `current`; the first (or last) when no pane has focus
    static func next(after current: Pane?, forward: Bool) -> Pane {
        guard let current, let index = allCases.firstIndex(of: current) else {
            return forward ? allCases.first! : allCases.last!
        }
        let step = forward ? 1 : -1
        return allCases[(index + step + allCases.count) % allCases.count]
    }
}

extension View {
    /// Binds this view's focus to `pane` when there is a binding
    @ViewBuilder
    func focused(_ binding: FocusState<Pane?>.Binding?, as pane: Pane) -> some View {
        if let binding {
            focused(binding, equals: pane)
        } else {
            self
        }
    }

    /// Tints a pane title with the accent colour while its pane has keyboard focus
    func paneTitle(focused: Bool, unfocused: Color = .primary) -> some View {
        foregroundStyle(focused ? Color.accentColor : unfocused)
    }
}

/// The inspector panels the picker selects by index
enum InspectorPanel {
    static let count = 7

    /// The panel after (or before) `selected`, wrapping at the ends. A selection out of range counts as the first.
    static func cycle(from selected: Int, forward: Bool) -> Int {
        let current = (0 ..< count).contains(selected) ? selected : 0
        return (current + (forward ? 1 : -1) + count) % count
    }
}

/// Moves focus to the next or previous pane of the focused window
struct PaneNavigator {
    let move: (_ forward: Bool) -> Void
}

extension FocusedValues {
    @Entry var paneNavigator: PaneNavigator?
}

struct PaneCommands: Commands {
    @FocusedValue(\.paneNavigator) var navigator: PaneNavigator?

    private static let paneKey = KeyEquivalent(Character(UnicodeScalar(NSF6FunctionKey)!))

    var body: some Commands {
        CommandGroup(after: .toolbar) {
            Button("Next Pane") {
                navigator?.move(true)
            }
            .keyboardShortcut(Self.paneKey, modifiers: [])
            .disabled(navigator == nil)

            Button("Previous Pane") {
                navigator?.move(false)
            }
            .keyboardShortcut(Self.paneKey, modifiers: .shift)
            .disabled(navigator == nil)
        }
    }
}
