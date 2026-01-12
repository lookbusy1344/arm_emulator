# Swift GUI Focus & Shortcut Issue Analysis

## Problem Description
In the Swift native macOS app, the "Step Over" shortcut (F10, and later F8) stops working and produces a system beep after the application loses and regains focus (e.g., via Cmd+Tab), specifically during interactive sessions (like `waiting_for_input`). While the user can type into the source code editor, the debug shortcuts fail to trigger the associated commands.

## Attempts & Hypotheses

### 1. Local Focus State (Failed)
*   **Attempt:** Implemented local `@FocusState` in `ConsoleView` to auto-focus the input field when waiting for input and release it afterwards.
*   **Hypothesis:** The input field was retaining focus, "stealing" the keystrokes from the main window context.
*   **Result:** No improvement. Beeping persisted.

### 2. Shortcut & Event Handling Fixes (Failed)
*   **Attempt:**
    *   Removed `[.function]` modifier from shortcuts (which logged warnings).
    *   Added F8 as an alternative shortcut.
    *   Updated `EmulatorViewModel` to explicitly handle `waiting_for_input` events to sync UI state.
    *   Removed manual `refreshState()` calls in `sendInput` to avoid race conditions with WebSocket events.
*   **Hypothesis:** Invalid modifier flags or state desynchronization (UI thinking VM is running when it's waiting) were causing the commands to be disabled.
*   **Result:** "Step Over" works initially but still fails after focus switches. Logs confirmed `MainView` focus was active, but shortcuts still beeped.

### 3. Unified Focus Management (Failed)
*   **Attempt:** Lifted focus state to `MainView` using a `FocusField` enum (`.editor`, `.console`). Explicitly passed this binding to `ConsoleView` and `EditorView` to programmatically force focus back to the editor after input.
*   **Hypothesis:** SwiftUI's focus management was getting lost between the two split panes. Forcing it back to the editor would restore the responder chain.
*   **Result:** Did not resolve the issue.

## Root Cause Analysis

The core issue appears to be a **Responder Chain disconnection**.

1.  **The Beep:** A system beep on macOS indicates that a key event was passed down the entire Responder Chain and **no object handled it**. This means the `DebugCommands` (defined in SwiftUI) are effectively "disconnected" from the current First Responder.
2.  **`focusedSceneValue` Fragility:** The debug commands rely on `@FocusedValue(\.viewModel)`. This property wrapper only works if the view *currently holding focus* (or one of its ancestors) has explicitly provided this value using `.focusedSceneValue`.
3.  **AppKit Interop:** The `EditorView` wraps an AppKit `NSTextView`. When the user clicks into the editor, the `NSTextView` becomes the `NSWindow`'s first responder. If the SwiftUI framework doesn't correctly bridge the `focusedSceneValue` from the wrapping `NSViewRepresentable` down to the active `NSTextView`, the focus chain breaks.
4.  **Focus Loss on App Switch:** When switching apps (Cmd+Tab), macOS might reset the First Responder. If it restores focus to a container view (like the `NSHostingView`) instead of the specific `NSTextView` or `TextField`, the focused value might not be published correctly.

## Recommended Next Steps

### 1. Robust Global Shortcuts (Bypass SwiftUI Focus)
Instead of relying on fragile `focusedSceneValue`, implemented global keyboard handling at the `AppDelegate` or `NSWindow` level.
*   **Action:** Implement `performKeyEquivalent:` in `AppDelegate` or subclass `NSWindow` to trap F8/F10 events globally for the window, regardless of which specific view has focus.
*   **Why:** This guarantees the shortcut works as long as the window is active, ignoring the complexities of the SwiftUI/AppKit responder chain mix.

### 2. Force AppKit Focus
Use AppKit internals to force focus restoration.
*   **Action:** In `MainView.onAppear` or `.onChange(of: scenePhase)`, find the underlying `NSTextView` and call `window?.makeFirstResponder(textView)`.
*   **Why:** SwiftUI's `@FocusState` is sometimes insufficient for wrapped AppKit views.

### 3. Debug the Responder Chain
Add code to print the responder chain when the issue occurs.
*   **Action:** Add a temporary debug command or timer that prints `NSApp.keyWindow?.firstResponder`.
*   **Why:** This will definitively reveal *what* object holds focus when the beep occurs (e.g., is it `NSTextView`, `SwiftUI.FocusView`, or `nil`?).

### 4. Re-architect Command Availability
Remove the dependency on `@FocusedValue`.
*   **Action:** Inject `EmulatorViewModel` as an `@EnvironmentObject` into the top-level `Commands` builder (if possible in newer SwiftUI versions) or use a singleton/delegate pattern for the menu bar.
*   **Why:** Decouples the menu bar from the specific focus state of the window content.
