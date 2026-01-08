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

## Next Steps

1. Add debug output to confirm `loadStack()` is called and completes
2. Add debug output to show `stackData.count` after populating
3. Add temporary UI text to show data fetch status
4. Verify the backend memory API is working correctly
5. Check if there's a SwiftUI view update issue preventing ForEach from rendering

## Notes

- The SP value updates correctly, so register state is working
- The header displays correctly, so the view itself renders
- This suggests the issue is specifically with the **async data loading** or **array population**
- The `MemoryView` works correctly, so the pattern should be similar
