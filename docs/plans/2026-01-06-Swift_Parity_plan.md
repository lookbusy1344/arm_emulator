# Swift GUI Feature Parity Plan

This document identifies features missing from the Swift GUI compared to the Wails GUI and TUI interfaces, and provides an implementation plan to achieve feature equality.

## Feature Comparison Matrix

| Feature | Wails GUI | TUI | Swift GUI | Priority |
|---------|-----------|-----|-----------|----------|
| **Execution Control** |
| Step (single instruction) | ✅ | ✅ | ✅ | - |
| Step Over (skip function calls) | ✅ | ✅ | ❌ | High |
| Step Out (run until return) | ✅ | ✅ | ❌ | High |
| Run/Continue | ✅ | ✅ | ✅ | - |
| Pause/Stop | ✅ | ✅ | ✅ | - |
| Reset | ✅ | ✅ | ✅ | - |
| Restart (preserve program) | ✅ | ✅ | ✅ | - |
| **Debugging** |
| Breakpoints (add/remove/toggle) | ✅ | ✅ | ✅ | - |
| Breakpoint conditions | ✅ | ✅ | ❌ | Low |
| Watchpoints (memory watch) | ✅ | ✅ | ❌ | High |
| Expression evaluator | ✅ | ✅ | ❌ | High |
| Debugger command input | ✅ | ✅ | ❌ | Medium |
| **Views** |
| Registers view | ✅ | ✅ | ✅ | - |
| Memory view | ✅ | ✅ | ✅ | - |
| Stack view | ✅ | ✅ | ✅ | - |
| Disassembly view | ✅ | ✅ | ✅ | - |
| Source/Editor view | ✅ | ✅ | ✅ | - |
| Console/Output view | ✅ | ✅ | ✅ | - |
| Status view | ✅ | ✅ | ✅ | - |
| Breakpoints list view | ✅ | ✅ | ✅ | - |
| Watchpoints list view | ✅ | ✅ | ✅ | - |
| **Visual Feedback** |
| Register change highlighting | ✅ | ✅ | ❌ | Medium |
| Memory write highlighting | ✅ | ✅ | ❌ | Medium |
| Symbol resolution in views | ✅ | ✅ | Partial | Low |
| Current line indicator | ✅ | ✅ | Partial | Low |
| **File Operations** |
| Open file | ✅ | N/A | ✅ | - |
| Save file | ✅ | N/A | ✅ | - |
| Recent files | ✅ | N/A | ✅ | - |
| Examples browser | ✅ | N/A | ✅ | - |
| **Keyboard Shortcuts** |
| F5 (Run/Continue) | ✅ | ✅ | ✅ | - |
| F9 (Toggle breakpoint) | ✅ | ✅ | ✅ | - |
| F10 (Step Over) | ✅ | ✅ | ✅ | - |
| F11 (Step) | ✅ | ✅ | ✅ | - |

## Missing Features - Detailed Implementation Plan

### Phase 1: Core Debugging Features (High Priority)

#### 1.1 Step Over / Step Out
**Gap:** Swift GUI only has basic step, missing step-over and step-out functionality.

**Implementation:**
1. Add API endpoints (if not existing):
   - `POST /api/v1/session/{id}/step-over`
   - `POST /api/v1/session/{id}/step-out`
   
2. Update `APIClient.swift`:
   ```swift
   func stepOver(sessionID: String) async throws
   func stepOut(sessionID: String) async throws
   ```

3. Update `EmulatorViewModel.swift`:
   ```swift
   func stepOver() async
   func stepOut() async
   ```

4. Update `MainView.swift` toolbar:
   - Add "Step Over" and "Step Out" buttons
   - Add keyboard shortcuts (⌘⇧T for step over, ⌘⌥T for step out)

**Files to modify:**
- `swift-gui/ARMEmulator/Services/APIClient.swift`
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift`
- `swift-gui/ARMEmulator/Views/MainView.swift`

#### 1.2 Watchpoints
**Gap:** No watchpoint support in Swift GUI.

**Implementation:**
1. Add API calls in `APIClient.swift`:
   ```swift
   func addWatchpoint(sessionID: String, address: UInt32, type: String) async throws -> Int
   func removeWatchpoint(sessionID: String, watchpointId: Int) async throws
   func getWatchpoints(sessionID: String) async throws -> [Watchpoint]
   ```

2. Add `Watchpoint` model in `Models/`:
   ```swift
   struct Watchpoint: Codable, Identifiable {
       let id: Int
       let address: UInt32
       let type: String // "read", "write", "readwrite"
   }
   ```

3. Update `EmulatorViewModel.swift`:
   ```swift
   @Published var watchpoints: [Watchpoint] = []
   func addWatchpoint(at address: UInt32, type: String) async
   func removeWatchpoint(id: Int) async
   ```

4. Create `WatchpointsView.swift` or integrate into existing `BreakpointsListView`

**Files to create/modify:**
- `swift-gui/ARMEmulator/Models/Watchpoint.swift` (new)
- `swift-gui/ARMEmulator/Services/APIClient.swift`
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift`
- `swift-gui/ARMEmulator/Views/WatchpointsView.swift` (new)

#### 1.3 Expression Evaluator
**Gap:** Cannot evaluate expressions (register values, memory, arithmetic).

**Implementation:**
1. Add API call:
   ```swift
   func evaluateExpression(sessionID: String, expression: String) async throws -> UInt32
   ```

2. Create `ExpressionEvaluatorView.swift`:
   ```swift
   struct ExpressionEvaluatorView: View {
       @State private var expression = ""
       @State private var results: [EvaluationResult] = []
       // Input field + history of evaluations
   }
   ```

3. Add to tab view in `MainView.swift`

**Files to create/modify:**
- `swift-gui/ARMEmulator/Services/APIClient.swift`
- `swift-gui/ARMEmulator/Views/ExpressionEvaluatorView.swift` (new)
- `swift-gui/ARMEmulator/Views/MainView.swift`

### Phase 2: Enhanced Debugging Views (Medium Priority)

#### 2.1 Breakpoints/Watchpoints List View ✅
**Status:** COMPLETE - Dedicated view showing all breakpoints and watchpoints.

**Implementation:**
1. ✅ Created `BreakpointsListView.swift` with:
   - List of all breakpoints with addresses
   - List of all watchpoints with addresses and types
   - Remove buttons for each item
   - Empty state view with instructions
2. ✅ Added as new tab in `MainView.swift`

**Files created/modified:**
- `swift-gui/ARMEmulator/Views/BreakpointsListView.swift` (new)
- `swift-gui/ARMEmulator/Views/MainView.swift`

#### 2.2 Debugger Command Input
**Gap:** No raw debugger command interface.

**Implementation:**
1. Add API call for command execution:
   ```swift
   func executeCommand(sessionID: String, command: String) async throws -> String
   ```

2. Create `CommandInputView.swift`:
   ```swift
   struct CommandInputView: View {
       @State private var command = ""
       @State private var history: [String] = []
       @State private var output = ""
       // Text field with up/down for history navigation
   }
   ```

3. Add to bottom of main view or as separate tab

**Files to create/modify:**
- `swift-gui/ARMEmulator/Services/APIClient.swift`
- `swift-gui/ARMEmulator/Views/CommandInputView.swift` (new)
- `swift-gui/ARMEmulator/Views/MainView.swift`

#### 2.3 Register Change Highlighting
**Gap:** Changed registers not highlighted after step.

**Implementation:**
1. Add state tracking in `EmulatorViewModel`:
   ```swift
   @Published var previousRegisters: RegisterState?
   @Published var changedRegisters: Set<String> = []
   ```

2. Update `RegistersView.swift` to highlight changed registers:
   ```swift
   func registerColor(for name: String) -> Color {
       changedRegisters.contains(name) ? .green : .primary
   }
   ```

**Files to modify:**
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift`
- `swift-gui/ARMEmulator/Views/RegistersView.swift`

#### 2.4 Memory Write Highlighting
**Gap:** Recently written memory addresses not highlighted.

**Implementation:**
1. Track recent writes via WebSocket events or last write info
2. Update `MemoryView.swift` to highlight recent writes in green

**Files to modify:**
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift`
- `swift-gui/ARMEmulator/Views/MemoryView.swift`

#### 2.5 Restart Command ✅
**Status:** COMPLETE - The `/reset` API endpoint preserves the loaded program and resets VM to initial state, which is the restart functionality. Already fully implemented with UI button and keyboard shortcut (⌘⇧R).

**Implementation:**
1. ✅ API call exists: `apiClient.reset(sessionID:)` 
2. ✅ ViewModel method: `viewModel.reset()`
3. ✅ UI button in toolbar with icon and keyboard shortcut

**Files modified:**
- `swift-gui/ARMEmulator/Services/APIClient.swift` (already has reset)
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift` (already has reset)
- `swift-gui/ARMEmulator/Views/MainView.swift` (already has button)

#### 2.6 Keyboard Shortcuts (Function Keys) ✅
**Status:** COMPLETE - Function key shortcuts added via Debug menu.

**Implementation:**
1. ✅ Created `DebugCommands.swift` with Debug menu
2. ✅ Added function key shortcuts:
   - F5 → Run/Continue
   - F9 → Toggle breakpoint at current PC
   - F10 → Step Over
   - F11 → Step
3. ✅ Integrated into app commands

**Files created/modified:**
- `swift-gui/ARMEmulator/Views/DebugCommands.swift` (new)
- `swift-gui/ARMEmulator/ARMEmulatorApp.swift`

### Phase 3: Polish and Parity (Low Priority)

#### 3.1 Conditional Breakpoints
**Gap:** Cannot set conditions on breakpoints.

**Implementation:**
1. Extend breakpoint model with condition field
2. Update UI to allow condition entry

#### 3.2 Symbol Resolution in All Views
**Gap:** Disassembly and source views don't always show symbols.

**Implementation:**
1. Ensure API returns symbol info
2. Update views to display symbol names

## API Endpoint Status

The following API endpoints exist in `openapi.yaml` and need Swift client implementation:

| Endpoint | Swift Client | Status |
|----------|--------------|--------|
| `/api/v1/session/{id}/step` | ✅ | Done |
| `/api/v1/session/{id}/watchpoint` (POST) | ❌ | Needed |
| `/api/v1/session/{id}/watchpoint/{id}` (DELETE) | ❌ | Needed |
| `/api/v1/session/{id}/watchpoints` (GET) | ❌ | Needed |
| `/api/v1/session/{id}/trace/*` | ❌ | Nice to have |
| `/api/v1/session/{id}/stats/*` | ❌ | Nice to have |

## Estimated Effort

| Phase | Features | Estimated Time |
|-------|----------|----------------|
| Phase 1 | Step Over/Out, Watchpoints, Expression Evaluator | 2-3 days |
| Phase 2 | List views, Command input, Highlighting, Shortcuts | 2-3 days |
| Phase 3 | Conditional breakpoints, Symbol polish | 1 day |
| **Total** | | **5-7 days** |

## Implementation Order

1. **Step Over / Step Out** - Basic debugging necessity
2. **Expression Evaluator** - High value for debugging
3. **Watchpoints** - Memory debugging capability
4. **Register/Memory Highlighting** - Visual feedback
5. **Keyboard Shortcuts** - Muscle memory from TUI/Wails
6. **Breakpoints List View** - Visibility of debugging state
7. **Command Input** - Power user feature
8. **Conditional Breakpoints** - Advanced debugging
