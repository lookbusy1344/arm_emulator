# MemoryView Diagnostic Plan

## Executive Summary

After 11 failed fix attempts, the MemoryView bug remains unsolved. The core symptom is that during byte-by-byte memory writes (e.g., `string_copy_manual.s`), the memory view either:
1. Scrolls unnecessarily on each write (original behavior), or
2. Fails to show updates/highlighting when "fixed" to not scroll

**Key finding**: Backend memory data IS correct. The bug is purely in UI presentation.

This document outlines a systematic diagnostic approach to isolate the root cause.

---

## Problem Statement

### Current Behavior
When `autoScrollEnabled=true` and stepping through a string copy loop:
- Each byte write triggers `loadMemoryAsync(at: writeAddr)`
- This sets `baseAddress = writeAddr`, causing the view to "jump" to that address
- Previously copied bytes scroll out of view

### Desired Behavior
- If the write address is within the visible 256-byte window: refresh data in place, highlight the write, DO NOT change `baseAddress`
- If the write address is outside the window: scroll to show it

### Why Previous Fixes Failed
All attempts to implement the desired behavior resulted in the view appearing "blank" or not updating, despite logs showing:
- Correct memory data fetched
- Correct highlight address set
- View refresh triggered (`refreshID = UUID()`)

---

## Diagnostic Hypotheses

### Hypothesis A: SwiftUI Identity/Diffing Failure
SwiftUI's view diffing may not detect changes when only `memoryData` changes but `baseAddress` remains constant.

**Evidence for:**
- Original code (which changes `baseAddress`) shows updates
- Fixed code (constant `baseAddress`) does not show updates
- `ForEach(0..<rowsToShow, id: \.self)` uses index-based identity, not content-based

**Evidence against:**
- `.id(refreshID)` should force complete rebuild
- Attempt 4 (refreshCount + .id()) didn't help

### Hypothesis B: State Update Race Condition
Multiple async Tasks spawned by rapid `onChange` triggers may overwrite each other's state updates.

**Evidence for:**
- `lastMemoryWrite` transitions: `writeAddr` → `nil` in rapid succession
- Each transition spawns a new `Task {}`
- Logs show highlight being set then immediately cleared

**Evidence against:**
- Attempt 7 (MainActor enforcement) didn't help
- Attempt 9 (task cancellation) didn't help

### Hypothesis C: ScrollView Position Reset
The ScrollView may be resetting its scroll position to the top when `.id(refreshID)` changes, making it appear that data isn't visible when it's actually scrolled out of view.

**Evidence for:**
- `.id()` changes force complete view reconstruction
- ScrollView would naturally reset to top on reconstruction
- User might be seeing row 0x8000 when expecting to see 0x80E0

**Evidence against:**
- Attempt 6 (removing `.id(refreshID)`) didn't help

### Hypothesis D: Stale Closure Capture
The `Task {}` closures in `onChange` may capture stale values of `baseAddress` or other state, causing writes to the wrong memory locations.

**Evidence for:**
- Swift closures capture values at creation time
- Multiple async operations could see inconsistent state
- Pattern: `Task { await loadMemoryAsync(at: writeAddr) }` captures current `writeAddr`

**Evidence against:**
- Logs show correct addresses being used

### Hypothesis E: View Not In Visible Hierarchy
The MemoryView might be getting recreated or removed from the hierarchy during rapid updates, causing state loss.

**Evidence for:**
- Tab-based UI could unmount views
- SwiftUI may optimize away "unchanged" views

**Evidence against:**
- User is actively viewing the Memory tab

### Hypothesis F: Local @State vs @Published Conflict
The local `@State memoryData` copy may not sync correctly with `viewModel.memoryData`, or SwiftUI may not observe changes to the local copy.

**Evidence for:**
- Two sources of truth: `@State memoryData` and `viewModel.memoryData`
- Attempt 8 (direct viewModel access) still failed, but implementation may have been incomplete

**Evidence against:**
- Assignment `memoryData = viewModel.memoryData` should trigger @State observation

---

## Diagnostic Protocol

### Phase 1: Establish Ground Truth (30 min)

**Objective**: Determine exactly what the user sees vs what the state contains.

#### Diagnostic 1.1: Add Persistent Debug Overlay

Add a non-scrolling overlay to MemoryView that shows current state:

```swift
// Add at top of VStack, outside ScrollView
VStack(alignment: .leading, spacing: 2) {
    Text("baseAddress: 0x\(String(format: "%08X", baseAddress))")
    Text("highlightAddr: \(highlightedWriteAddress.map { String(format: "0x%08X", $0) } ?? "nil")")
    Text("memoryData.count: \(memoryData.count)")
    Text("refreshID: \(refreshID.uuidString.prefix(8))")
    if memoryData.count > 0 {
        Text("First byte: 0x\(String(format: "%02X", memoryData[0]))")
    }
}
.font(.system(size: 9, design: .monospaced))
.padding(4)
.background(Color.yellow.opacity(0.3))
```

**What to observe:**
- Does `baseAddress` change on each write?
- Does `highlightedWriteAddress` get set then cleared?
- Does `memoryData.count` stay at 256?
- Does `refreshID` change on each write?

#### Diagnostic 1.2: Row Render Tracking

Add `onAppear` logging to `MemoryRowView`:

```swift
var body: some View {
    HStack(spacing: 4) {
        // existing content
    }
    .onAppear {
        DebugLog.log("MemoryRowView APPEAR: 0x\(String(format: "%08X", address))", category: "MemoryRow")
    }
    .onChange(of: bytes) { newBytes in
        DebugLog.log("MemoryRowView BYTES CHANGED: 0x\(String(format: "%08X", address)) -> \(newBytes.prefix(4).map { String(format: "%02X", $0) })", category: "MemoryRow")
    }
}
```

**What to observe:**
- Do rows appear/disappear during stepping?
- Does `onChange(of: bytes)` fire when memory changes?

### Phase 2: Isolate ScrollView Behavior (30 min)

**Objective**: Determine if ScrollView position is the issue.

#### Diagnostic 2.1: Track Scroll Position

Add a `GeometryReader` with `PreferenceKey` to track scroll offset:

```swift
struct ScrollOffsetPreferenceKey: PreferenceKey {
    static var defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = nextValue()
    }
}

// Inside ScrollView, wrap content:
GeometryReader { geo in
    Color.clear.preference(
        key: ScrollOffsetPreferenceKey.self,
        value: geo.frame(in: .named("scroll")).minY
    )
}
.frame(height: 0)

// On ScrollView:
.coordinateSpace(name: "scroll")
.onPreferenceChange(ScrollOffsetPreferenceKey.self) { offset in
    DebugLog.log("Scroll offset: \(offset)", category: "MemoryScroll")
}
```

**What to observe:**
- Does scroll offset jump to 0 when `refreshID` changes?
- Does scroll offset remain stable when only `memoryData` changes?

#### Diagnostic 2.2: Programmatic Scroll Anchor Test

Wrap in `ScrollViewReader` and force scroll position after refresh:

```swift
ScrollViewReader { proxy in
    ScrollView([.vertical, .horizontal]) {
        VStack(alignment: .leading, spacing: 0) {
            ForEach(0..<rowsToShow, id: \.self) { row in
                MemoryRowView(...)
                    .id("row_\(row)")
            }
        }
    }
    .onChange(of: refreshID) { _ in
        // Always scroll to row containing highlight
        if let highlight = highlightedWriteAddress {
            let targetRow = Int((highlight - baseAddress) / UInt32(bytesPerRow))
            if targetRow >= 0 && targetRow < rowsToShow {
                proxy.scrollTo("row_\(targetRow)", anchor: .center)
                DebugLog.log("Scrolled to row \(targetRow) for highlight", category: "MemoryScroll")
            }
        }
    }
}
```

**What to observe:**
- Does forced scrolling make the highlight visible?
- If yes: confirms scroll position reset is the issue

### Phase 3: Isolate State Propagation (30 min)

**Objective**: Verify state changes flow correctly through the view hierarchy.

#### Diagnostic 3.1: Direct ViewModel Observation Test

Temporarily replace local `@State memoryData` with direct `viewModel.memoryData` access everywhere:

```swift
// Remove: @State private var memoryData: [UInt8] = []
// Change all references from memoryData to viewModel.memoryData
// Change bytesForRow to use viewModel.memoryData directly
```

**What to observe:**
- Does the view update when viewModel.memoryData changes?
- Are there any type mismatches or async issues?

#### Diagnostic 3.2: State Change Timestamp

Add timestamp tracking to verify state update sequence:

```swift
@State private var stateUpdateLog: [(Date, String)] = []

// In onChange handler:
stateUpdateLog.append((Date(), "highlight=\(writeAddr), base=\(baseAddress)"))

// Display in debug overlay:
ForEach(stateUpdateLog.suffix(5), id: \.0) { entry in
    Text("\(entry.0.timeIntervalSince1970): \(entry.1)")
}
```

**What to observe:**
- Are state updates happening in expected order?
- Is there evidence of overwrites?

### Phase 4: Isolate Task/Concurrency Issues (30 min)

**Objective**: Determine if async task scheduling causes state corruption.

#### Diagnostic 4.1: Synchronous State Update Test

Replace async task with synchronous MainActor call:

```swift
.onChange(of: viewModel.lastMemoryWrite) { newWriteAddress in
    // Remove Task wrapper - execute synchronously
    MainActor.assumeIsolated {
        if let writeAddr = newWriteAddress {
            highlightedWriteAddress = writeAddr
            // Don't load memory - just set highlight
            refreshID = UUID()
        }
    }
}
```

**What to observe:**
- Does highlight appear without memory load?
- Is the issue in async memory loading or state update?

#### Diagnostic 4.2: Single Task Guard

Ensure only one task handles writes at a time:

```swift
@State private var currentWriteTask: Task<Void, Never>?

.onChange(of: viewModel.lastMemoryWrite) { newWriteAddress in
    currentWriteTask?.cancel()
    currentWriteTask = Task {
        guard !Task.isCancelled else { return }
        // existing logic
    }
}
```

**What to observe:**
- Does cancelling previous tasks improve behavior?
- Are there still race conditions?

### Phase 5: Minimal Reproduction (45 min)

**Objective**: Create the simplest possible test case that reproduces the bug.

#### Diagnostic 5.1: Standalone Test View

Create a minimal view that simulates the issue:

```swift
struct MemoryViewTestHarness: View {
    @State private var memoryData: [UInt8] = Array(repeating: 0, count: 256)
    @State private var baseAddress: UInt32 = 0x8000
    @State private var highlightAddress: UInt32? = nil
    @State private var refreshID = UUID()
    @State private var writeCounter = 0

    var body: some View {
        VStack {
            Button("Simulate Write") {
                // Simulate a byte write at offset from base
                let offset = writeCounter % 256
                memoryData[offset] = UInt8(writeCounter & 0xFF)
                highlightAddress = baseAddress + UInt32(offset)
                refreshID = UUID()
                writeCounter += 1
            }

            // Same ScrollView structure as MemoryView
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    ForEach(0..<16, id: \.self) { row in
                        // Same row structure
                    }
                }
            }
            .id(refreshID)
        }
    }
}
```

**What to observe:**
- Does this minimal version exhibit the same bug?
- If not, what's different about the real MemoryView?

#### Diagnostic 5.2: External State Mutation Test

Test if the issue is specific to `onChange` triggering:

```swift
// Add manual refresh button
Button("Force Refresh") {
    Task {
        await loadMemoryAsync(at: baseAddress)
        refreshID = UUID()
    }
}
```

**What to observe:**
- Does manual refresh show correct data?
- Is the issue specific to `onChange` code path?

### Phase 6: Backend Verification (15 min)

**Objective**: Confirm backend is not the issue (likely already ruled out, but verify).

#### Diagnostic 6.1: Direct API Call

During reproduction, curl the API directly:

```bash
curl "http://localhost:8080/api/v1/session/{id}/memory?address=0x80E0&length=64" | xxd
```

**What to observe:**
- Does API return the expected bytes?
- Are bytes changing as the program executes?

#### Diagnostic 6.2: WebSocket State Inspection

Log all WebSocket state events during stepping:

```swift
// In handleEvent:
DebugLog.log("WS EVENT RAW: \(event)", category: "WebSocket")
```

**What to observe:**
- Are `hasWrite` and `writeAddr` being sent correctly?
- Is there any event ordering issue?

---

## Decision Tree

Based on diagnostic results, follow this decision tree:

```
Q1: Does debug overlay show correct state?
├─ NO: State is not being updated correctly
│   └─ Focus on: Task/concurrency issues (Phase 4)
│
└─ YES: State is correct but view doesn't reflect it
    │
    Q2: Does row onChange(of: bytes) fire?
    ├─ NO: SwiftUI not detecting changes
    │   └─ Focus on: View identity/diffing (Phase 1.2, 3.1)
    │
    └─ YES: Changes detected but not rendered
        │
        Q3: Does scroll offset jump on refresh?
        ├─ YES: ScrollView position reset
        │   └─ Focus on: ScrollView anchoring (Phase 2.2)
        │
        └─ NO: Rendering issue
            └─ Focus on: Minimal reproduction (Phase 5)
```

---

## Likely Root Causes (Ranked by Probability)

### 1. ScrollView Position Reset (60% likely)
When `.id(refreshID)` changes, SwiftUI destroys and recreates the ScrollView, resetting scroll position to top. User is at address 0x8000 when they should be at 0x80E0.

**Fix approach**: Remove `.id(refreshID)`, use `ScrollViewReader` for explicit scroll control, ensure row identities are stable.

### 2. Race Between onChange Callbacks (25% likely)
The `lastMemoryWrite` signal transitions `writeAddr` → `nil` rapidly. Each transition spawns a Task. The second Task (for `nil`) clears the highlight before the view renders.

**Fix approach**: Debounce the onChange handler, ignore `nil` transitions, or use a different signaling mechanism that doesn't send `nil`.

### 3. SwiftUI Diffing Optimization (10% likely)
SwiftUI may determine that `MemoryRowView` inputs haven't changed (same `address`, `bytes` array reference even if contents changed) and skip re-render.

**Fix approach**: Ensure `bytes` array is a new allocation on each change, add explicit `Equatable` conformance, use `.id()` on individual rows.

### 4. View Lifecycle Issues (5% likely)
Tab switching or view reconstruction during rapid updates loses state.

**Fix approach**: Verify with `.onAppear`/`.onDisappear` logging, consider `@StateObject` for more stable ownership.

---

## Implementation Order

1. **Phase 1.1** (Debug overlay) - Establishes ground truth quickly
2. **Phase 2.1** (Scroll tracking) - Tests most likely hypothesis
3. **Phase 2.2** (Scroll anchor) - Confirms/fixes scroll issue if found
4. **Phase 4.1** (Synchronous test) - Rules out async issues
5. **Phase 1.2** (Row tracking) - Fine-grained render observation
6. **Phase 5** (Minimal repro) - Only if above phases inconclusive

---

## Success Criteria

The bug is fixed when:
1. Running `string_copy_manual.s` in step mode
2. Memory view shows addresses 0x80E0+ (destination buffer) staying visible
3. Green highlighting appears on each written byte
4. View does NOT jump/scroll on each write when the write is within visible range
5. View DOES scroll when write falls outside the 256-byte window

---

## Notes for Implementation

- Keep ALL diagnostic code behind `#if DEBUG` or a feature flag
- Log to both console AND the debug overlay for easy correlation
- After each diagnostic phase, commit findings before proceeding
- Don't implement fixes during diagnostic phase - only observe and document
