# MemoryView Bug Analysis

## Issue Description

When running `string_copy_manual.s` in the Swift GUI, the memory view has a UX problem:

**Current behavior:** As each character is written into memory during the copy loop, the memory view updates its start address to exactly the write address. This causes previously copied characters to scroll out of view.

**Desired behavior:** The memory view should only scroll if the write address is completely outside the visible range (256 bytes from baseAddress). If the write is visible, just refresh the data in place while keeping the view stable.

## Working Functionality (Original Code)

The original code correctly:
- Shows green highlighting on written bytes
- Updates memory data when writes occur
- Forces view refresh with `refreshID = UUID()`

The ONLY issue is that `loadMemoryAsync(at: writeAddr)` is called on every write, which sets `baseAddress = writeAddr`, causing the view to scroll.

## Root Cause

In `MemoryView.swift` line 118:
```swift
await loadMemoryAsync(at: writeAddr)
```

This calls `loadMemoryAsync` which sets `baseAddress = loadedAddress` (the write address), causing the view to scroll to start at that address.

## Failed Fix Attempts

### Attempt 1: Simple Visibility Check

**Change:**
```swift
if autoScrollEnabled {
    let isVisible = writeAddr >= baseAddress && writeAddr < baseAddress + UInt32(totalBytes)
    if isVisible {
        await refreshMemoryAsync()  // Don't change baseAddress
    } else {
        await loadMemoryAsync(at: writeAddr)  // Scroll to it
    }
}
```

**Result:** Green highlighting vanished, written bytes not visible.

**Analysis from logs:**
- `refreshMemoryAsync()` was called correctly
- `highlightedWriteAddress` was set correctly
- `refreshID = UUID()` was called to force re-render
- But immediately after, `lastMemoryWrite` became `nil`, triggering another onChange
- This cleared `highlightedWriteAddress = nil` before the view could render

### Attempt 2: Don't Clear Highlight on Nil

**Additional change:**
```swift
} else {
    // No write this step - keep the highlight until next write
    return  // Instead of clearing highlightedWriteAddress
}
```

**Result:** Still no green highlighting, still no visible written bytes.

**Analysis:** The highlight should persist now, but something else is preventing the view from showing the updated data.

### Attempt 3: Extract MemoryGridView

**Change:** Created a separate `MemoryGridView` struct that takes `memoryData` as an explicit `let` parameter, hoping SwiftUI would better track changes.

**Result:** No improvement.

### Attempt 4: Add refreshCount with .id()

**Change:** Added `@State private var refreshCount = 0` and `.id(refreshCount)` on the grid, incrementing on refresh.

**Result:** No improvement.

## Key Observations from Logs

When the fix is applied:
```
🔵 [MemoryView] Processing write at 0x000080E4
🔵 [ViewModel] loadMemory: address=0x00008000, length=256
🔵 [ViewModel] loadMemory: Got 256 bytes at 0x00008000
🔵 [MemoryView] Memory refreshed at 0x00008000, 256 bytes
🔵 [MemoryView] Highlight set to 0x000080E4, baseAddress=0x00008000, memoryData.count=256
🔵 [MemoryView] Refreshing view with new UUID
```

This shows:
- Memory IS being loaded correctly (256 bytes at 0x8000)
- Highlight IS being set correctly (0x000080E4)
- View refresh IS being triggered (new UUID)

Yet the UI shows nothing.

## Unanswered Questions

1. **Why does the original code work for highlighting?**
   - Original: `loadMemoryAsync(at: writeAddr)` changes `baseAddress`
   - The same highlight clearing happens in original code
   - Yet highlighting works in original

2. **What's different about `refreshMemoryAsync()` vs `loadMemoryAsync()`?**
   - Both update `memoryData`
   - Both trigger `refreshID = UUID()`
   - Only difference: `loadMemoryAsync` also changes `baseAddress`

3. **Is SwiftUI somehow dependent on `baseAddress` changing?**
   - The view uses `baseAddress` in the ForEach to calculate row addresses
   - Maybe SwiftUI's diffing doesn't detect changes when only `memoryData` changes?

4. **Is there a race condition?**
   - The onChange handler spawns async Tasks
   - Multiple state updates could be in flight
   - Could state be getting overwritten?

## Potential Unexplored Solutions

1. **Make `baseAddress` change slightly then change back** - Force SwiftUI to see a change
2. **Use `@ObservedObject` or `@StateObject` for memory data** - Different state management
3. **Add explicit equatable conformance** - Help SwiftUI detect changes
4. **Debug SwiftUI rendering** - Add logging inside MemoryRowView to see if it's being called
5. **Check if memoryData actually contains the written bytes** - Log the actual byte values after refresh

## Current State

The code has been reset to original via `git checkout`. The bug (view scrolling to write address) persists, but highlighting works.

## Files Involved

- `swift-gui/ARMEmulator/Views/MemoryView.swift` - Main file with the issue
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift` - Provides `lastMemoryWrite` and `loadMemory()`
