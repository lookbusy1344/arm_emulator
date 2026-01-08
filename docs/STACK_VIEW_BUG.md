# Stack View Display Bug

## Issue Summary

The Stack View in the Swift GUI (`swift-gui/ARMEmulator/Views/StackView.swift`) displays the stack pointer (SP) value in the header but **does not display any stack entries** below it.

## Expected Behavior

When a program is loaded and running:
1. The header shows: `Stack (SP = 0x00050000)` (or current SP value)
2. Below the header, a scrollable list of stack entries should appear showing:
   - Offset from SP (e.g., `SP+0`, `SP-4`, `SP-8`, etc.)
   - Memory address (e.g., `0x0004FFFC`)
   - 32-bit value at that address (e.g., `0x00008018`)
   - ASCII representation of the bytes
   - Annotations (e.g., `← LR`, `← R0?`, `← Code?`)

## Actual Behavior

- The header displays correctly with the current SP value
- **The stack entries list is completely empty** - no rows appear
- The `stackData` array appears to be empty or not rendering

## Reproduction Steps

1. Build and run the Swift GUI app: `cd swift-gui && xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator build`
2. Launch the app with a test program: `open ARMEmulator.app --args ../examples/fibonacci.s`
3. Step through the program (the SP value will change as function calls happen)
4. Observe the Stack View panel on the right side of the UI
5. **Bug**: Stack entries do not appear despite SP changing

## Test Program

Use `examples/fibonacci.s` - it has function calls (`print_string`, `read_int`) that push/pop the stack with `PUSH {lr}` and `POP {pc}`, making stack activity easy to observe.

## Code Location

**File**: `swift-gui/ARMEmulator/Views/StackView.swift`

### Key Components

1. **View State**:
   ```swift
   @State private var stackData: [StackEntry] = []
   @State private var localMemoryData: [UInt8] = []
   ```

2. **Load Trigger**:
   ```swift
   .task {
       await loadStack()
   }
   .onChange(of: viewModel.registers.sp) { _ in
       Task {
           await loadStack()
       }
   }
   ```

3. **Data Loading Function**:
   ```swift
   private func loadStack() async {
       let sp = viewModel.registers.sp
       guard sp >= 0x1000 else { return }

       let offset = UInt32(wordsToShow / 2 * 4)
       let startAddress = sp >= offset ? sp - offset : 0

       do {
           localMemoryData = try await viewModel.fetchMemory(at: startAddress, length: wordsToShow * 4)
       } catch {
           stackData = []
           return
       }

       // ... builds stackData array ...
       stackData = data
   }
   ```

4. **Rendering**:
   ```swift
   ForEach(stackData) { entry in
       StackRowView(...)
   }
   ```

## Attempted Fixes (Current Dirty Code)

Recent changes attempted to fix the issue:
- Added `viewModel.fetchMemory()` method to fetch memory asynchronously
- Changed from using `viewModel.memoryData` directly to `localMemoryData` state
- Added `.task` and `.onChange` to trigger loading

These changes were meant to ensure proper async data fetching but **the display is still empty**.

## Debugging Strategy

1. **Check if `loadStack()` is being called**:
   - Added debug logging (DebugLog.log calls)
   - Check Xcode console output when app runs

2. **Check if memory fetch succeeds**:
   - Does `viewModel.fetchMemory()` return data?
   - Does the API call to backend succeed?
   - Verify backend is running and responding

3. **Check if `stackData` array is populated**:
   - Log the count of entries after building array
   - Verify the array is not being cleared somewhere

4. **Check for threading/async issues**:
   - Is the UI update happening on @MainActor?
   - Is there a race condition between loads?

5. **Check ForEach rendering**:
   - Does the `ForEach(stackData)` block execute?
   - Is `stackData.isEmpty` true when it shouldn't be?

## Related Code

- **ViewModel**: `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift`
  - `fetchMemory()` method (lines 516-522)
  - `loadMemory()` method (lines 503-514)

- **API Client**: `swift-gui/ARMEmulator/Services/APIClient.swift`
  - `getMemory()` API call

- **Backend API**: Runs on `http://localhost:8080`
  - Endpoint: `GET /api/v1/session/{id}/memory?address={addr}&length={len}`

## ROOT CAUSE IDENTIFIED

**Date**: 2026-01-08

Debug logging revealed the actual problem:

```
🔵 [StackView] loadStack() called, SP = 0x00050000
🔵 [StackView] Fetching memory from 0x0004FFC0, length: 128 bytes
❌ [StackView] Failed to fetch stack memory: Server error (500):
    "Failed to read memory: memory access violation: address 0x00050000 is not mapped"
```

### The Problem

The `loadStack()` function tries to read memory **centered around** the SP:
- Calculates: `startAddress = SP - 64 bytes`
- Fetches 128 bytes (±64 bytes around SP)
- This means it reads from `(SP - 64)` to `(SP + 64)`

**But the stack grows downward!**
- SP = `0x00050000` (top of stack, not mapped)
- Valid stack memory is **below** SP: `0x0004FFFF`, `0x0004FFFE`, etc.
- Addresses **at or above** SP (`0x00050000+`) are unmapped memory

When the fetch tries to read up to `0x00050040`, the backend correctly returns:
```
"memory access violation: address 0x00050000 is not mapped"
```

### Why It Fails

```swift
let offset = UInt32(wordsToShow / 2 * 4)  // 64 bytes
let startAddress = sp >= offset ? sp - offset : 0  // SP - 64

// Tries to fetch 128 bytes:
localMemoryData = try await viewModel.fetchMemory(at: startAddress, length: wordsToShow * 4)
// This reads from (SP - 64) to (SP - 64 + 128) = (SP + 64)
//                                                   ^^^^^^^^ UNMAPPED!
```

### The Fix

Only read memory **below** the stack pointer:
- Read from `(SP - 128)` to `SP` (or `SP - 4` to avoid the exact boundary)
- This shows the "top" 128 bytes of the stack (most recent pushes)
- All addresses will be valid stack memory

Alternative: Read from `(SP - 64)` to `(SP - 1)` (only 63 bytes, but safer)

## Next Steps

1. ✅ **DONE**: Identified root cause - memory fetch crosses unmapped region
2. Modify `loadStack()` to only read memory below SP
3. Adjust the display to show stack entries from `SP-128` to `SP-4`
4. Test with fibonacci.s to verify stack entries appear
5. Update UI to clearly indicate "stack grows down" direction

## Notes

- The SP value updates correctly, so register state is working ✓
- The header displays correctly, so the view itself renders ✓
- `loadStack()` IS being called on SP changes ✓
- The issue is specifically with the **memory address calculation** crossing unmapped memory
- The `MemoryView` works because it reads arbitrary addresses, not relative to SP
