# Memory Pane Issues and Failed Fixes

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

**Attempts**:
1. Added green foreground color to written bytes in `MemoryView.swift:173`
2. Added bold font weight to written bytes in `MemoryView.swift:174`
3. Applied same highlighting to ASCII column in `MemoryView.swift:193-194`
4. Added debug logging to track when writes are detected

**Current Code**:
```swift
// In MemoryRowView
private func isRecentWrite(byteIndex: Int) -> Bool {
    guard let writeAddr = lastWriteAddress else { return false }
    let byteAddr = address + UInt32(byteIndex)
    // Highlight the written byte and the next 3 bytes (for 32-bit writes)
    return writeAddr <= byteAddr && byteAddr < writeAddr + 4
}

// In hex bytes rendering
Text(String(format: "%02X", bytes[i]))
    .frame(width: 20)
    .foregroundColor(isRecentWrite(byteIndex: i) ? .green : .primary)
    .fontWeight(isRecentWrite(byteIndex: i) ? .bold : .regular)
```

**Status**: ❌ FAILED - Highlighting code is in place but not visible to user. Debug logs confirm writes are detected:
```
🔵 [ViewModel] Memory write detected at 0x0004FFC0
🔵 [MemoryView] Auto-scrolling to memory write at 0x0004FFC0
```

### 3. Auto-scroll to Memory Writes
**Problem**: When a memory location is altered, it should automatically be scrolled into view.

**Attempts**:
1. Added `onChange(of: viewModel.lastMemoryWrite)` handler in `MemoryView.swift:99-106`
2. Auto-scroll jumps to write address when `autoScrollEnabled` is true
3. Removed PC auto-scroll (was causing view to jump on every step)
4. Added `refreshMemory()` function to update memory data without changing base address

**Current Code**:
```swift
.onChange(of: viewModel.currentPC) { _ in
    // Refresh memory at current address to show updates
    Task {
        await refreshMemory()
    }
}
.onChange(of: viewModel.lastMemoryWrite) { newWriteAddress in
    if autoScrollEnabled, let writeAddr = newWriteAddress {
        DebugLog.log("Auto-scrolling to memory write at 0x\(String(format: "%08X", writeAddr))", category: "MemoryView")
        Task {
            await loadMemory(at: writeAddr)
        }
    }
}
```

**Status**: ⚠️ PARTIAL - Debug logs show auto-scroll is triggering, but user reports no visible changes.

## Debug Evidence

Logs from stepping through `addressing_modes.s` show:
- Memory writes ARE being detected by backend
- ViewModel IS receiving write notifications
- Auto-scroll handler IS being triggered
- Memory refresh IS being called

Example log sequence:
```
🔵 [ViewModel] Memory write detected at 0x0004FFC0
🔵 [MemoryView] Auto-scrolling to memory write at 0x0004FFC0
[...next step...]
🔵 [ViewModel] Memory write detected at 0x0004FFC4
🔵 [MemoryView] Auto-scrolling to memory write at 0x0004FFC4
```

## Possible Root Causes

### Theory 1: Timing/Race Condition
The `lastMemoryWrite` might be getting cleared or overwritten before the view can render with the highlighting.

**Evidence**:
- Changed registers use similar mechanism and DO work
- No code clears `lastMemoryWrite` between steps
- But highlighting persists for registers, not memory

### Theory 2: View Update Not Triggering
SwiftUI might not be re-rendering `MemoryRowView` when `lastWriteAddress` changes.

**Evidence**:
- `lastWriteAddress` is passed as a parameter to `MemoryRowView`
- Changes to `@Published var lastMemoryWrite` should trigger view update
- But user reports no visible highlighting

### Theory 3: Memory Data Not Refreshing
The memory bytes themselves might not be updating, making the highlighting moot.

**Evidence**:
- `refreshMemory()` is called on every PC change
- User reports "no memory changing"
- But logs don't show memory fetch failures

### Theory 4: Wrong Memory Address Range
The memory view might be showing address 0x8000 (code) while writes are at 0x0004FFC0 (stack).

**Evidence**:
- Memory view defaults to 0x8000
- Stack writes are at 0x0004FFC0-0x0004FFC8
- User would need to click "SP" button to see stack
- Auto-scroll should handle this, but might not be working

### Theory 5: Background/Async Issue
The async memory loading might complete after the highlight state is cleared.

**Evidence**:
- All memory operations use `Task { await ... }`
- `refreshMemory()` is async
- Timing between state update and view render is unclear

## What Works

1. ✅ Memory pane resizing
2. ✅ Horizontal scroll bar
3. ✅ Detection of memory writes (backend)
4. ✅ Propagation of write events to ViewModel
5. ✅ Triggering of auto-scroll handler
6. ✅ Debug logging throughout the chain

## What Doesn't Work

1. ❌ Visual highlighting of written bytes
2. ❌ Visible memory value changes
3. ❌ Observable auto-scroll behavior

## Next Steps to Debug

1. **Add visual confirmation**: Add a `Text()` view showing current `lastMemoryWrite` value to confirm it's being set
2. **Log memory data**: Add logging to show actual memory bytes being fetched and displayed
3. **Force view refresh**: Try using `@State private var refreshID = UUID()` and `.id(refreshID)` to force re-render
4. **Check Stack tab**: Compare with Stack view which DOES update and show memory changes
5. **Simplify test**: Create minimal reproduction with single memory write and manual address entry
6. **Profile timing**: Add timestamps to logs to check for race conditions
7. **Check API response**: Log the actual memory bytes returned by `loadMemory()` API call

## Files Modified

- `swift-gui/ARMEmulator/Views/MemoryView.swift` - Main view logic
- `swift-gui/ARMEmulator/Views/MainView.swift` - Layout constraints
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift` - Write detection logging

## User Test Instructions

To verify the highlighting SHOULD work (if it were working):

1. Load `addressing_modes.s`
2. Click **Memory** tab
3. Click **SP** button (jumps to stack at ~0x0004FFC0)
4. Enable **Auto-scroll** toggle
5. Step through with **Step** button
6. Expected: Green bold bytes at write addresses, view jumping to writes
7. Actual: User reports nothing visible changes

## Conclusion

The memory write detection and auto-scroll infrastructure is in place and functioning at the code level (confirmed by logs), but the visual manifestation (highlighting and scrolling) is not appearing to the user. This suggests either:
- A SwiftUI view update issue
- A timing/async problem
- The memory data itself not updating
- Or the user is looking at the wrong address range

Needs further investigation with more targeted debugging.
