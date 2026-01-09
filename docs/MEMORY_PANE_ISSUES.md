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

**Fix Applied** (2026-01-09):
1. Added local `highlightedWriteAddress` state to track highlighting separately from navigation
2. Added `refreshID` state with `.id(refreshID)` modifier to force SwiftUI re-renders
3. Coordinated onChange handlers - PC handler skips refresh when write is pending
4. Made `loadMemoryAsync(at:)` a proper async function that can be awaited
5. Updated ViewModel to clear `lastMemoryWrite` when no write occurs

**Status**: ✅ FIXED - See below for final resolution.

### 3. Auto-scroll to Memory Writes
**Problem**: When a memory location is altered, it should automatically be scrolled into view.

**Root Cause**: The memory write was occurring near the end of the stack segment (e.g. `0x0004FFC0`). The memory view requests a fixed block of 256 bytes. This request extended beyond the mapped stack segment (ending at `0x00050000`). The backend `GetMemory` implementation was failing the entire request if *any* byte in the range was unreadable. This caused the API to return 500 Error, resulting in `memoryData` being cleared to `[]`.

**Fix Applied** (2026-01-09):
- Updated `service/debugger_service.go` to gracefully handle unmapped memory in `GetMemory`. It now returns `0` for unmapped bytes instead of failing the request.
- This allows `MemoryView` to display memory regions that partially overlap with valid segments (e.g. end of stack).

**Status**: ✅ FIXED - Memory now loads correctly even at segment boundaries, enabling auto-scroll and highlighting to work.

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

**Why it failed**: In `refreshState()`, `currentPC` is set BEFORE `lastMemoryWrite`. The PC onChange fired first, seeing `nil` for `lastMemoryWrite`, and refreshed at the old address.

**Secondary Issue**: Even after fixing the race condition, the memory view showed `Data: 0b`. This was because the auto-scroll target (`0x0004FFC0`) caused a read request that spanned into unmapped memory, causing the backend to fail the request.

### Final Resolution

1. **Backend Fix**: Modified `GetMemory` in `service/debugger_service.go` to return partial results (zeros for unmapped bytes) instead of error.
2. **Frontend Cleanup**: Removed temporary debug bar and refresh button from `MemoryView.swift`.

The combination of the earlier race condition fix (removing the conflicting PC handler) and the backend fix (allowing boundary reads) solves all reported issues.
