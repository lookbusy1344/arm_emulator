# Memory Pane Issues and Fixes

## Original Issues Reported

### 1. Memory Pane Width
**Problem**: The memory pane only shows 12 bytes instead of all 16 bytes in the character view on the right.

**Fix Applied**:
- Reduced source pane minimum width from 300px to 150px in `MainView.swift:93`
- Increased right panel maximum width from 500px to 800px in `MainView.swift:149`
- Added horizontal scrolling to memory view in `MemoryView.swift:75`

**Status**: ✅ FIXED - User can now resize panes more flexibly and horizontal scroll prevents clipping.

### 2. Memory Write Highlighting
**Problem**: Edited memory locations should be highlighted in green (like registers are).

**Root Cause Found**: Race condition between two `onChange` handlers:
1. `onChange(of: viewModel.currentPC)` - refreshed memory at old address (0x8000)
2. `onChange(of: viewModel.lastMemoryWrite)` - tried to scroll to new address (0x0004FFC0)

Both started async Tasks that competed to update shared `viewModel.memoryData` state. Depending on which completed last, the wrong data would be displayed.

**Fix Applied** (2026-01-09):
1. Added local `highlightedWriteAddress` state to track highlighting separately from navigation
2. Added `refreshID` state with `.id(refreshID)` modifier to force SwiftUI re-renders
3. Coordinated onChange handlers - PC handler skips refresh when write is pending
4. Made `loadMemoryAsync(at:)` a proper async function that can be awaited
5. Updated ViewModel to clear `lastMemoryWrite` when no write occurs

**Status**: ❌ NOT FIXED - Multiple attempts failed. See "Fix Attempts" section below.

### 3. Auto-scroll to Memory Writes
**Problem**: When a memory location is altered, it should automatically be scrolled into view.

**Status**: ❌ NOT FIXED - Auto-scroll code exists but memory loading itself is broken.

## Technical Details

### Analysis (2026-01-09)

**Original Problem**: Two `onChange` handlers competed for shared state:

```swift
// BROKEN - Race condition between handlers
.onChange(of: viewModel.currentPC) { _ in
    // This checked viewModel.lastMemoryWrite, but timing was unreliable
    if viewModel.lastMemoryWrite == nil {
        Task { await refreshMemory() }  // Refreshed at baseAddress (0x8000)
    }
}
.onChange(of: viewModel.lastMemoryWrite) { newWriteAddress in
    Task { await loadMemory(at: writeAddr) }  // Tried to change baseAddress
}
```

**Why it failed**: In `refreshState()`, `currentPC` is set BEFORE `lastMemoryWrite`:
1. `updateRegisters()` → sets `currentPC` → triggers PC onChange
2. `lastMemoryWrite = writeAddr` → triggers its onChange

The PC onChange might fire before `lastMemoryWrite` is updated, seeing the OLD value (nil from previous step). This caused it to refresh at 0x8000 (code area) instead of the write address (stack area).

**Key Insight**: The user saw the **row highlighted blue** (PC highlight working) but **no green bytes** (write highlight missing). This confirms the view was still showing the CODE area (where PC is), not the STACK area (where writes happen). The auto-scroll wasn't actually scrolling.

### Fix Attempts (All Failed)

#### Attempt 1: Fix Race Condition Between onChange Handlers
- **Theory**: Two onChange handlers (PC and lastMemoryWrite) were competing
- **Fix**: Removed PC onChange handler entirely, simplified to single handler
- **Result**: ❌ FAILED - Same behavior

#### Attempt 2: Add Local Highlighting State
- **Theory**: Using `viewModel.lastMemoryWrite` directly might cause timing issues
- **Fix**: Added `highlightedWriteAddress` as local `@State` property
- **Result**: ❌ FAILED - `Write` shows value but `Base` doesn't update

#### Attempt 3: Force View Refresh
- **Theory**: SwiftUI might not be detecting state changes
- **Fix**: Added `refreshID = UUID()` with `.id(refreshID)` modifier
- **Result**: ❌ FAILED - No visible change

#### Attempt 4: Add Debug Logging
- **Theory**: Need visibility into what's happening
- **Fix**: Added extensive DebugLog calls throughout the code path
- **Result**: Logs not visible when running from Finder (only in Xcode console)

#### Attempt 5: Add Debug Bar and Refresh Button
- **Theory**: Need visual confirmation of state values
- **Fix**: Added yellow debug bar showing `Base`, `Write`, `Data`, `AutoScroll`
- **Result**: Debug bar reveals the problem but doesn't fix it

### Current State (2026-01-09)

**Debug bar shows:**
- `Base: 0x00008000` - NOT updating when write occurs
- `Write: 0x0004FFC0` - IS being set correctly
- `Data: 0b` - NO memory data is loading at all
- `AutoScroll: ON` (presumably)

**Key observations:**
1. `viewModel.lastMemoryWrite` IS being set (Write shows value)
2. `highlightedWriteAddress` IS being set (same value appears)
3. `memoryData` is EMPTY (no hex bytes displayed, Data shows 0b)
4. `baseAddress` is NOT updating (stays at 0x8000)
5. Even the initial `.task` load appears to fail (no data on view appear)
6. Manual Refresh button also doesn't load data

**Root cause likely:**
- `viewModel.loadMemory()` is failing silently
- Either `sessionID` is nil, or the API call is failing
- The error handling silently sets `memoryData = []`

### Files Modified

- `swift-gui/ARMEmulator/Views/MemoryView.swift` - Added debug bar, refresh button, logging
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift` - Added logging to loadMemory

### What Needs Investigation

1. **Why is `loadMemory` failing?**
   - Is `sessionID` nil when Memory tab is shown?
   - Is the API returning an error?
   - Run from Xcode to see DebugLog output

2. **Session timing issue?**
   - Memory tab might be displayed before session is fully initialized
   - The `.task` runs on view appear, but session might not be ready

3. **API issue?**
   - Check if `/api/v1/session/{id}/memory` endpoint is working
   - Test with curl to isolate Swift vs API issue

### How to Debug Further

1. **Run from Xcode** to see console output with DebugLog messages
2. **Check session initialization** - is sessionID set when Memory view appears?
3. **Test API directly**:
   ```bash
   # Get session ID from app, then:
   curl http://localhost:8080/api/v1/session/{SESSION_ID}/memory?address=0x8000&length=256
   ```
4. **Add NSLog for terminal debugging** - DebugLog uses print() which doesn't show in terminal
