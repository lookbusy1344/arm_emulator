# Swift GUI Focus Fix Plan

## Problem Summary

After the app loses and regains focus (Cmd+Tab), F8/F10 shortcuts beep instead of executing debug commands. This occurs specifically during `waiting_for_input` state after `SWI #0x05` (READ_STRING).

## Architecture Analysis

### Current Flow
```
ARMEmulatorApp
  └── .commands { DebugCommands() }  ← Uses @FocusedValue(\.viewModel)
  └── WindowGroup
        └── MainView
              ├── .focusedSceneValue(\.viewModel, viewModel)  ← Publishes viewModel
              ├── EditorWithGutterView (NSViewRepresentable → NSTextView)
              └── ConsoleView (TextField for input)
```

### Root Cause: Responder Chain Disconnection

1. **The beep** = macOS saying "key event unhandled by responder chain"
2. **`@FocusedValue` dependency** = `DebugCommands` only works if a view publishing `.focusedSceneValue` is in the focus hierarchy
3. **NSTextView complication** = `EditorWithGutterView` wraps an AppKit `NSTextView`. When focused, it becomes `NSWindow.firstResponder`, but the SwiftUI focus bridge may not propagate `focusedSceneValue` correctly
4. **App switch reset** = Cmd+Tab can reset `firstResponder` to a container (`NSHostingView`) rather than the specific view that published the focused value

### Failed Attempt Analysis (a889b1e)

The commit added:
```swift
@FocusState private var isWindowFocused: Bool
// ...
.focusable()
.focused($isWindowFocused)
.focusEffectDisabled()
.onAppear { isWindowFocused = true }
```

**Why it failed:** `@FocusState` manages SwiftUI focus, but:
- The `NSTextView` inside `EditorWithGutterView` takes AppKit-level focus
- SwiftUI's `focused($isWindowFocused)` doesn't override AppKit's `firstResponder`
- Setting `isWindowFocused = true` in `onAppear` only fires once

## Proposed Solutions (Priority Order)

### Solution 1: Global Key Handler via NSApplication (Recommended)

**Bypass SwiftUI focus entirely** by intercepting function keys at the application level.

```swift
// In AppDelegate.swift
@MainActor
class AppDelegate: NSObject, NSApplicationDelegate {
    weak var viewModel: EmulatorViewModel?
    
    func applicationDidFinishLaunching(_ notification: Notification) {
        NSEvent.addLocalMonitorForEvents(matching: .keyDown) { [weak self] event in
            if self?.handleFunctionKey(event) == true {
                return nil  // Consume the event
            }
            return event  // Pass through
        }
    }
    
    private func handleFunctionKey(_ event: NSEvent) -> Bool {
        guard let viewModel = viewModel,
              event.modifierFlags.contains(.function) else { return false }
        
        switch Int(event.keyCode) {
        case kVK_F5:
            Task { await viewModel.run() }
            return true
        case kVK_F8, kVK_F10:
            Task { await viewModel.stepOver() }
            return true
        case kVK_F11:
            Task { await viewModel.step() }
            return true
        case kVK_F9:
            Task { await viewModel.toggleBreakpoint(at: viewModel.currentPC) }
            return true
        default:
            return false
        }
    }
}
```

**Requires:** Pass `viewModel` to `AppDelegate` after initialization.

**Pros:**
- Works regardless of focus state
- Zero dependency on SwiftUI focus machinery
- Most reliable solution

**Cons:**
- Must manage viewModel lifecycle carefully
- Global handler—must check window is key window

### Solution 2: Override performKeyEquivalent in Custom NSWindow

Create a custom `NSWindow` subclass that handles function keys:

```swift
class EmulatorWindow: NSWindow {
    var viewModel: EmulatorViewModel?
    
    override func performKeyEquivalent(with event: NSEvent) -> Bool {
        // Handle F-keys even if responder chain doesn't
        if handleDebugShortcut(event) {
            return true
        }
        return super.performKeyEquivalent(with: event)
    }
    
    private func handleDebugShortcut(_ event: NSEvent) -> Bool {
        // Similar to Solution 1
    }
}
```

**Then in ARMEmulatorApp:**
```swift
WindowGroup {
    MainView()
        .background(WindowAccessor { window in
            // Configure custom window
        })
}
```

### Solution 3: Scene Phase + Force First Responder

React to app activation and force focus restoration:

```swift
// In MainView.swift
@Environment(\.scenePhase) private var scenePhase

// In body, add:
.onChange(of: scenePhase) { newPhase in
    if newPhase == .active {
        // Force the window to re-establish responder chain
        DispatchQueue.main.async {
            NSApp.keyWindow?.makeFirstResponder(NSApp.keyWindow?.contentView)
            // Re-trigger focusedSceneValue by toggling focus
            isWindowFocused = false
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.05) {
                isWindowFocused = true
            }
        }
    }
}
```

**Pros:** Uses existing SwiftUI patterns  
**Cons:** Timing-dependent, may flash focus

### Solution 4: Remove @FocusedValue Dependency

Store the viewModel reference outside the focus system:

```swift
// Singleton or ObservableObject at app level
class DebugCommandsHandler: ObservableObject {
    static let shared = DebugCommandsHandler()
    weak var activeViewModel: EmulatorViewModel?
}

// In MainView
.onAppear {
    DebugCommandsHandler.shared.activeViewModel = viewModel
}

// In DebugCommands
struct DebugCommands: Commands {
    var viewModel: EmulatorViewModel? {
        DebugCommandsHandler.shared.activeViewModel
    }
}
```

**Pros:** Decouples from focus entirely  
**Cons:** Manual lifecycle management, won't auto-nil on window close

## Debugging Steps

Before implementing, confirm the hypothesis with diagnostic code:

```swift
// Add to MainView temporarily
.onReceive(NotificationCenter.default.publisher(for: NSWindow.didBecomeKeyNotification)) { _ in
    print("🔑 Window became key")
    print("   firstResponder: \(String(describing: NSApp.keyWindow?.firstResponder))")
    print("   focusedViewModel: \(viewModel)")
}

.onReceive(NotificationCenter.default.publisher(for: NSWindow.didResignKeyNotification)) { _ in
    print("🔓 Window resigned key")
}
```

Also add to `DebugCommands`:
```swift
var body: some Commands {
    CommandMenu("Debug") {
        Button("Step Over") {
            print("🎯 stepOver triggered, viewModel: \(viewModel != nil)")
            // ...
        }
    }
}
```

## Recommended Implementation Order

1. **Add diagnostic logging** to confirm `@FocusedValue` becomes `nil` after focus switch
2. **Implement Solution 1** (NSEvent monitor) as it's most robust
3. **Keep existing SwiftUI shortcuts** as fallback for menu bar interaction
4. **Test** with the READ_STRING scenario specifically

## Test Case

```arm
; Test program - run this, let it wait for input, Cmd+Tab away and back
        LDR     R0, =src_buffer
        MOV     R1, #100
        SWI     #0x05           ; READ_STRING - triggers waiting_for_input
        SWI     #0x00           ; EXIT
        
.data
src_buffer: .space 100
```

Expected: F8 should execute stepOver after Cmd+Tab return  
Actual (bug): System beep, no action
