# Fix: Pause Button Behavior During Stepping

## Problem Summary

The "Stop" button has incorrect semantics and availability:

1. **Mislabeled**: Called "Stop" but actually pauses execution (doesn't reset)
2. **Wrong icon**: Uses stop icon (■) instead of pause icon (❚❚)
3. **Available when it shouldn't be**: Enabled during `.breakpoint` state (F10 stepping) when the program is already paused

### Observed Behavior

From execution log:
```
🔵 [ViewModel] stop() called - current status: breakpoint, canStep: true
```

User pressed "Stop" while stepping with F10. Expected: button disabled (already paused). Actual: button enabled and callable.

After pressing stop during stepping, F10 continues from current PC instead of being blocked. Stale WebSocket events also override the halted state.

## Root Cause Analysis

### 1. Incorrect `canStop` Logic

**Current** (`EmulatorViewModel.swift:56-58`):
```swift
var canStop: Bool {
    status == .running || status == .waitingForInput || status == .breakpoint
}
```

**Problem**: Includes `.breakpoint` state, but during stepping the VM is already paused at a breakpoint - there's nothing to pause.

### 2. Semantic Confusion

Two distinct operations exist:
- **Pause**: Interrupt continuous execution, keep PC where it is
- **Reset**: Return to entry point, start fresh

Current "Stop" button does Pause but is labeled/iconified as Stop.

### 3. Stale WebSocket Events

WebSocket events arriving after `stop()` can override the halted state:
```
⚠️ [ViewModel] Ignoring stale WebSocket state transition from halted to breakpoint
```

The filter works for some events but not all, and the PC still updates via other code paths.

## Solution

### Conceptual Model

| Button | Icon | Action | Enabled When |
|--------|------|--------|--------------|
| **Pause** | ❚❚ | Interrupt continuous run, keep PC | `.running`, `.waitingForInput` |
| **Reset** | ↻ | Reset to entry point | Always (already exists) |
| **Step (F10)** | → | Execute one instruction | `.idle`, `.breakpoint` |
| **Run** | ▶ | Continuous execution | `.idle`, `.breakpoint` |

### Code Changes

#### 1. Rename `canStop` → `canPause` and fix logic

**File**: `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift`

```swift
// BEFORE
var canStop: Bool {
    status == .running || status == .waitingForInput || status == .breakpoint
}

// AFTER
var canPause: Bool {
    status == .running || status == .waitingForInput
}
```

#### 2. Rename `stop()` → `pause()` for clarity

**File**: `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel+Execution.swift`

Rename function from `stop()` to `pause()` to match semantics.

#### 3. Update UI references

**Files**:
- `swift-gui/ARMEmulator/Views/MainViewToolbar.swift`
- `swift-gui/ARMEmulator/Views/DebugCommands.swift`

Changes:
- Change button label from "Stop" to "Pause"
- Change icon from `stop.fill` to `pause.fill`
- Change `viewModel.stop()` to `viewModel.pause()`
- Change `.disabled(!viewModel.canStop)` to `.disabled(!viewModel.canPause)`
- Update help text from "Stop execution" to "Pause execution"

#### 4. Update tests

**File**: `swift-gui/ARMEmulatorTests/ViewModels/EmulatorViewModelTests.swift`

- Rename `canStop` tests to `canPause`
- Remove test `testCanStopWhenAtBreakpoint` (no longer valid)
- Update test assertions to reflect new logic

### Backend Changes

None required. The backend `Pause()` function already does the right thing (stops execution without reset). The `ResetToEntryPoint()` function handles reset and is called by the existing Reset button.

## Test Plan

1. **Pause during continuous run**
   - Load program, press Run (▶)
   - Press Pause (❚❚) → execution stops, PC stays where it was
   - Pause button becomes disabled

2. **Pause unavailable during stepping**
   - Load program, press F10 to step
   - Verify Pause button is disabled (grayed out)
   - F10 continues to work for stepping

3. **Reset always available**
   - At any state, Reset (↻) should reset PC to entry point
   - After reset, F10 starts from beginning

4. **Stale WebSocket events**
   - Pause during continuous run
   - Verify status stays `.halted` (not overridden by stale events)

## Files to Modify

| File | Changes |
|------|---------|
| `EmulatorViewModel.swift` | Rename `canStop` → `canPause`, fix logic |
| `EmulatorViewModel+Execution.swift` | Rename `stop()` → `pause()` |
| `MainViewToolbar.swift` | Update button label, icon, action, disabled state |
| `DebugCommands.swift` | Update menu item label, action, disabled state |
| `EmulatorViewModelTests.swift` | Update tests for new naming and logic |

## Complexity

Low - primarily renaming and adjusting one boolean condition.
