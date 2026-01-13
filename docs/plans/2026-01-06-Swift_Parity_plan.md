# Swift GUI Feature Parity Plan

This document identifies features missing from the Swift GUI compared to the Wails GUI and TUI interfaces, and provides an implementation plan to achieve feature equality.

## Status Summary

**Implementation Progress:** ✅ **COMPLETE** - All high and medium priority features have been implemented!

### Completed in this session (2026-01-06):
- ✅ Restart Command verification (already implemented)
- ✅ Keyboard Shortcuts (F5/F9/F10/F11)
- ✅ Breakpoints List View

### Previously completed:
- ✅ Step Over / Step Out
- ✅ Watchpoints (add/remove/list)
- ✅ Expression Evaluator
- ✅ Register Change Highlighting
- ✅ Memory Write Highlighting

### Phase 3 (Low Priority) remaining:
- ⚠️ Breakpoint conditions (requires backend API support)
- ⚠️ Symbol resolution polish (partial implementation)

### Not Applicable:
- N/A Debugger command input (no backend API, Expression Evaluator covers use cases)

## Feature Comparison Matrix

| Feature | Wails GUI | TUI | Swift GUI | Priority |
|---------|-----------|-----|-----------|----------|
| **Execution Control** |
| Step (single instruction) | ✅ | ✅ | ✅ | - |
| Step Over (skip function calls) | ✅ | ✅ | ✅ | - |
| Step Out (run until return) | ✅ | ✅ | ✅ | - |
| Run/Continue | ✅ | ✅ | ✅ | - |
| Pause/Stop | ✅ | ✅ | ✅ | - |
| Reset | ✅ | ✅ | ✅ | - |
| Restart (preserve program) | ✅ | ✅ | ✅ | - |
| **Debugging** |
| Breakpoints (add/remove/toggle) | ✅ | ✅ | ✅ | - |
| Breakpoint conditions | ✅ | ✅ | ❌ | Low |
| Watchpoints (memory watch) | ✅ | ✅ | ✅ | - |
| Expression evaluator | ✅ | ✅ | ✅ | - |
| Debugger command input | ✅ | ✅ | N/A | - |
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
| Register change highlighting | ✅ | ✅ | ✅ | - |
| Memory write highlighting | ✅ | ✅ | ✅ | - |
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

### Phase 1: Core Debugging Features (High Priority) ✅

#### 1.1 Step Over / Step Out ✅
**Status:** COMPLETE - Already implemented prior to this session.

**Implementation:**
1. ✅ API endpoints exist
2. ✅ `APIClient.swift` has methods
3. ✅ `EmulatorViewModel.swift` has methods
4. ✅ `MainView.swift` toolbar has buttons and shortcuts (⌘⇧T, ⌘⌥T)

#### 1.2 Watchpoints ✅
**Status:** COMPLETE - Already implemented prior to this session.

**Implementation:**
1. ✅ API calls in `APIClient.swift`
2. ✅ `Watchpoint` model exists
3. ✅ `EmulatorViewModel.swift` has watchpoint methods
4. ✅ `WatchpointsView.swift` with add/remove UI

#### 1.3 Expression Evaluator ✅
**Status:** COMPLETE - Already implemented prior to this session.

**Implementation:**
1. ✅ API call exists in `APIClient.swift`
2. ✅ `ExpressionEvaluatorView.swift` with full UI
3. ✅ Added as tab in `MainView.swift`

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

#### 2.2 Debugger Command Input ⚠️
**Status:** NOT APPLICABLE - No API endpoint exists for generic command execution.

**Note:** The Expression Evaluator (Phase 1.3) provides similar functionality for evaluating expressions and inspecting values. A generic command input would require backend API support that doesn't currently exist.

#### 2.3 Register Change Highlighting ✅
**Status:** COMPLETE - Already implemented prior to this session.

**Implementation:**
1. ✅ State tracking in `EmulatorViewModel` with `previousRegisters` and `changedRegisters`
2. ✅ `RegistersView` uses `isChanged` parameter to highlight changed registers in green

#### 2.4 Memory Write Highlighting ✅
**Status:** COMPLETE - Already implemented prior to this session.

**Implementation:**
1. ✅ State tracking in `EmulatorViewModel` with `previousRegisters` and `changedRegisters`
2. ✅ `RegistersView` uses `isChanged` parameter to highlight changed registers in green

**Files modified:**
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift`
- `swift-gui/ARMEmulator/Views/RegistersView.swift`

#### 2.4 Memory Write Highlighting ✅
**Status:** COMPLETE - Already implemented prior to this session.

**Implementation:**
1. ✅ Tracks recent writes via `lastMemoryWrite` in `EmulatorViewModel`
2. ✅ `MemoryView` highlights recent write addresses

**Files modified:**
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

All required API endpoints for Phase 1 and Phase 2 features are implemented:

| Endpoint | Swift Client | Status |
|----------|--------------|--------|
| `/api/v1/session/{id}/step` | ✅ | Done |
| `/api/v1/session/{id}/step-over` | ✅ | Done |
| `/api/v1/session/{id}/step-out` | ✅ | Done |
| `/api/v1/session/{id}/watchpoint` (POST) | ✅ | Done |
| `/api/v1/session/{id}/watchpoint/{id}` (DELETE) | ✅ | Done |
| `/api/v1/session/{id}/watchpoints` (GET) | ✅ | Done |
| `/api/v1/session/{id}/evaluate` (POST) | ✅ | Done |
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
