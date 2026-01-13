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

## Subsequent Fix Attempts (2026-01)

These were attempted after the initial analysis above; none resolved the user-visible behavior.

### Attempt 5: “Only scroll if out-of-view” (overlap-based)

**Change:** In `onChange(of: lastMemoryWrite)`, detect whether the write overlaps the current `[baseAddress, baseAddress+256)` window.
- If in-view: `refreshMemoryAsync()` at the existing base
- If out-of-view: `loadMemoryAsync(at: targetBase)`

**Result:** Logs show the intended branch is taken (in-view refresh) and `baseAddress` stays stable, but the UI still behaves as if the view is “jumping” / not showing the updated region.

### Attempt 6: Remove forced view rebuild (`.id(refreshID)`)

**Hypothesis:** forcing a full view identity change on each write makes SwiftUI recreate the ScrollView content and resets scroll offset.

**Change:** Remove `.id(refreshID)` and any `refreshID = UUID()` usage.

**Result:** No change in observed behavior.

### Attempt 7: MainActor / highlight race mitigation

**Hypothesis:** `lastMemoryWrite` is emitted as `writeAddr` then quickly as `nil` on the next tick (`refreshState()` clears it when no write), racing SwiftUI.

**Changes:**
- Ensure write handling runs on the main actor (`Task { @MainActor in ... }`)
- Avoid clearing highlight on `lastMemoryWrite == nil`

**Result:** Still no visible highlight / update per user report.

### Attempt 8: Remove local `@State` memory cache

**Hypothesis:** local `@State memoryData` copy prevents SwiftUI from observing the real published changes.

**Change:** Render directly from `viewModel.memoryData`.

**Result:** No change.

### Attempt 9: Coalesce write-handling tasks

**Hypothesis:** overlapping async tasks (one per byte write) stomp state and/or create inconsistent UI updates.

**Change:** track and cancel a previous `writeHandlingTask` before starting a new one.

**Result:** No change.

### Attempt 10: Prove backend bytes are changing (Post-refresh diagnostics)

**Change:** After each refresh/load, log:
- `baseAddress`, `writeAddr`, computed `offset`
- the specific `memoryData[offset]`
- the full 16-byte row containing the write

**Result (key discovery):** Backend memory fetch is correct; bytes *do* change as expected.
Example:
- write `0x80E4` -> byte `0x48` ('H')
- write `0x80E5` -> byte `0x65` ('e')

So this is not “stale memory data”; it’s UI presentation (scroll/window/visibility) diverging from the underlying model.

### Attempt 11: Stabilize row identity + virtualization

**Hypothesis:** `ForEach(0..<rowsToShow, id: \.self)` + frequent updates causes SwiftUI to rebuild in a way that resets scroll position; also row identity should be by address not index.

**Changes:**
- use `LazyVStack`
- build `rowAddresses` and `ForEach(rowAddresses, id: \.self)`

**Result:** Still reproduces per user.

## Updated Understanding

- The memory bytes *are* changing in the fetched window (proved by `Post-refresh:` logs).
- The remaining bug is likely one of:
  1) Scroll position being reset (user loses the `0x80E0` area and ends up back near `0x8000`), or
  2) SwiftUI diffing/invalidations not repainting the visible rows reliably under rapid updates, or
  3) A second view/state path is overwriting the window/scroll outside the code we’ve been focusing on.

## Further Diagnostic Steps (Next)

1. **Confirm what address range is actually visible when the user reports “no update”.**
   - Add a small, always-visible overlay in MemoryView showing:
     - `baseAddress` (window start)
     - the currently highlighted write address
     - the last `Post-refresh` row start address
   - This distinguishes “data updated but user is scrolled elsewhere” vs “data not painted”.

2. **Track ScrollView offset explicitly (SwiftUI PreferenceKey).**
   - Use a `GeometryReader` + custom `PreferenceKey` to record the ScrollView’s content offset.
   - Log when the offset jumps (especially back toward top) during byte writes.

3. **Instrument row rendering.
   - Add an `onAppear` / `onChange(of: bytes)` log inside `MemoryRowView` for the specific row containing `dst_buffer` (e.g. `0x000080E0`).
   - If the row body is not re-evaluated when bytes change, it’s a diffing/identity issue.

4. **Programmatic scroll anchoring (diagnostic, not final UX).
   - Wrap in `ScrollViewReader` and, after refresh, `scrollTo(addressRow, anchor: .top)` ONLY when we detect the view jumped unexpectedly.
   - If this “fixes” visibility, it confirms an offset-reset problem.

5. **Rule out tab/view re-creation.
   - Add a `DebugLog` in `MemoryView.init` and/or `.onAppear/.onDisappear` to see if the Memory tab is being reconstructed during execution.

6. **Backend cross-check (already effectively confirmed).
   - Optionally curl `GET /api/v1/session/{id}/memory?address=0x80E0&length=64` during reproduction to validate outside SwiftUI.

## Current State

- UX target remains: keep the current visible window stable and only scroll when the write is fully outside.
- Multiple SwiftUI-side attempts (in-view refresh, identity fixes, MainActor, task coalescing, LazyVStack + address IDs) did not resolve the user-visible behavior.
- Latest evidence strongly indicates the remaining issue is ScrollView/visibility (not incorrect backend memory reads).

## Files Involved

- `swift-gui/ARMEmulator/Views/MemoryView.swift` - Main file with the issue
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift` - Provides `lastMemoryWrite` and `loadMemory()`
