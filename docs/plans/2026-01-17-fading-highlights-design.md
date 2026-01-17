# Fading Highlights Design

## Overview

Implement time-based fading highlights for register and memory changes in the Swift GUI. Currently, highlights are binary (on/off) and each new change replaces the previous one. The new system allows multiple highlights to be visible simultaneously, each fading independently over 1.5 seconds with an ease-out animation curve.

## Requirements

- **Multiple concurrent highlights**: Register/memory changes should not replace existing highlights
- **Time-based fade**: Highlights fade over 1.5 seconds after the change occurs
- **Ease-out curve**: Highlights stay bright initially, then fade faster toward the end
- **Timer restart**: If the same item changes again before fade completes, restart the 1.5s timer
- **SwiftUI-native**: Leverage built-in animation system for optimal performance

## Data Model Changes

### Current State
```swift
// Binary on/off state
@Published var changedRegisters: Set<String>
@Published var highlightedWriteAddress: UInt32?
```

### New State
```swift
// Track when items changed using unique identifiers
@Published var registerHighlights: [String: UUID] = [:]  // RegisterName -> unique ID
@Published var memoryHighlights: [UInt32: UUID] = [:]    // Address -> unique ID

// Task tracking for cleanup
private var highlightTasks: [String: Task<Void, Never>] = [:]
private var memoryHighlightTasks: [UInt32: Task<Void, Never>] = [:]
```

### Why UUIDs?

SwiftUI's animation system works best with value identity changes:
- Adding highlight: `registerHighlights["R0"] = UUID()` triggers animation to green
- Removing highlight: `registerHighlights["R0"] = nil` triggers animation back to normal
- Replacing UUID: Restarts the animation

### Backward Compatibility

Keep `changedRegisters` as a computed property for existing code:
```swift
var changedRegisters: Set<String> {
    Set(registerHighlights.keys)
}
```

## View Implementation

### Current Approach
```swift
.foregroundColor(isChanged ? .green : .primary)
```

### New Approach
```swift
// RegisterRow receives optional UUID
struct RegisterRow: View {
    let name: String
    let value: UInt32
    let highlightID: UUID?

    var body: some View {
        HStack {
            Text("\(name):")
                .foregroundColor(highlightID != nil ? .green : .primary)
                .animation(.easeOut(duration: 1.5), value: highlightID)
            // ... rest of row
        }
    }
}
```

### Animation Mechanism

- `.animation(.easeOut(duration: 1.5), value: highlightID)` watches the UUID
- `nil` → `UUID()`: animates to green
- `UUID()` → `nil`: animates back to normal color
- `UUID()` → different `UUID()`: restarts green animation

### Multiple Highlights

If R0 changes at T=0s and T=0.5s:
- First change: green fading from T=0 to T=1.5
- Second change: green restarts at T=0.5, fades to T=2.0
- Both animations are visible and independent

## Timer-Based Cleanup

### Highlight Function
```swift
func highlightRegister(_ name: String) {
    // Cancel existing fade task (if any)
    highlightTasks[name]?.cancel()

    // Add new highlight (triggers animation to green)
    registerHighlights[name] = UUID()

    // Schedule removal after 1.5 seconds
    highlightTasks[name] = Task { @MainActor in
        try? await Task.sleep(nanoseconds: 1_500_000_000)
        registerHighlights[name] = nil  // Triggers fade-out animation
        highlightTasks[name] = nil
    }
}

func highlightMemoryAddress(_ address: UInt32, size: UInt32) {
    // Cancel existing fade task
    memoryHighlightTasks[address]?.cancel()

    // Add highlights for each byte in the write
    for offset in 0..<size {
        let addr = address + offset
        memoryHighlights[addr] = UUID()

        memoryHighlightTasks[addr] = Task { @MainActor in
            try? await Task.sleep(nanoseconds: 1_500_000_000)
            memoryHighlights[addr] = nil
            memoryHighlightTasks[addr] = nil
        }
    }
}
```

### Task Cancellation

If R0 changes twice rapidly:
1. First change at T=0: schedules removal at T=1.5
2. Second change at T=0.5: **cancels** T=1.5 task, schedules new removal at T=2.0
3. Result: R0 stays highlighted until T=2.0 (1.5s after the *last* change)

### MainActor Safety

All UI updates happen on `@MainActor`, ensuring thread safety. Task cleanup is automatic when the ViewModel is deallocated.

## Integration with Existing Code

### Register Updates

**Current:**
```swift
func updateRegisters(_ newRegisters: RegisterState) {
    var changed = Set<String>()
    if let prev = previousRegisters {
        changed = detectRegisterChanges(previous: prev, new: newRegisters)
    }

    previousRegisters = registers
    changedRegisters = changed  // Replaces entire set
    registers = newRegisters
}
```

**New:**
```swift
func updateRegisters(_ newRegisters: RegisterState) {
    // Detect changes
    var changed = Set<String>()
    if let prev = previousRegisters {
        changed = detectRegisterChanges(previous: prev, new: newRegisters)
    }

    // Highlight each changed register (independent timers)
    for registerName in changed {
        highlightRegister(registerName)
    }

    // Update state
    previousRegisters = registers
    registers = newRegisters
    currentPC = newRegisters.pc
}
```

### Memory Highlights

**Current:**
```swift
.onChange(of: viewModel.lastMemoryWrite) {
    guard let writeAddr = viewModel.lastMemoryWrite else { return }
    highlightedWriteAddress = writeAddr  // Single address
    // ... scroll logic
}
```

**New:**
```swift
.onChange(of: viewModel.lastMemoryWrite) {
    guard let writeAddr = viewModel.lastMemoryWrite else { return }

    // Highlight written bytes (triggers 1.5s fade for each)
    viewModel.highlightMemoryAddress(writeAddr, size: viewModel.lastMemoryWriteSize)

    // ... existing scroll logic
}
```

### View Pass-Through

**RegistersView:**
```swift
RegistersView(
    registers: viewModel.registers,
    registerHighlights: viewModel.registerHighlights  // Pass highlights map
)
```

**RegisterRow:**
```swift
RegisterRow(
    name: "R0",
    value: registers.r0,
    highlightID: registerHighlights["R0"]  // Extract specific highlight
)
```

**MemoryRowView:**
```swift
MemoryRowView(
    address: baseAddress + UInt32(row * bytesPerRow),
    bytes: bytesForRow(row),
    memoryHighlights: viewModel.memoryHighlights,  // Pass entire map
    lastWriteSize: viewModel.lastMemoryWriteSize
)
```

## Testing Strategy

### Unit Tests

Test highlight lifecycle in `EmulatorViewModelTests.swift`:

```swift
func testRegisterHighlightAdded() {
    viewModel.highlightRegister("R0")
    XCTAssertNotNil(viewModel.registerHighlights["R0"])
}

func testRegisterHighlightFadesAfterDelay() async {
    viewModel.highlightRegister("R0")
    try await Task.sleep(nanoseconds: 1_600_000_000)  // 1.6s
    XCTAssertNil(viewModel.registerHighlights["R0"])
}

func testRapidChangesRestartTimer() async {
    viewModel.highlightRegister("R0")
    try await Task.sleep(nanoseconds: 500_000_000)  // 0.5s
    viewModel.highlightRegister("R0")  // Restart timer
    try await Task.sleep(nanoseconds: 1_200_000_000)  // 1.2s total (0.7s after restart)
    XCTAssertNotNil(viewModel.registerHighlights["R0"])  // Should still be highlighted
}

func testMultipleRegisterHighlightsIndependent() async {
    viewModel.highlightRegister("R0")
    try await Task.sleep(nanoseconds: 500_000_000)
    viewModel.highlightRegister("R1")

    // Both should be highlighted
    XCTAssertNotNil(viewModel.registerHighlights["R0"])
    XCTAssertNotNil(viewModel.registerHighlights["R1"])

    // Wait for R0 to fade
    try await Task.sleep(nanoseconds: 1_200_000_000)  // 1.7s total
    XCTAssertNil(viewModel.registerHighlights["R0"])
    XCTAssertNotNil(viewModel.registerHighlights["R1"])  // R1 still visible
}

func testMemoryHighlightMultipleBytes() {
    viewModel.highlightMemoryAddress(0x8000, size: 4)

    // All 4 bytes should be highlighted
    XCTAssertNotNil(viewModel.memoryHighlights[0x8000])
    XCTAssertNotNil(viewModel.memoryHighlights[0x8001])
    XCTAssertNotNil(viewModel.memoryHighlights[0x8002])
    XCTAssertNotNil(viewModel.memoryHighlights[0x8003])
}
```

### Integration Tests

Test with actual program execution:

```swift
func testMultipleRegisterChangesShowIndependentHighlights() async {
    // Load program that modifies R0, R1, R2 in sequence
    let program = """
        MOV R0, #42
        MOV R1, #100
        MOV R2, #200
    """
    await viewModel.loadProgram(source: program)

    await viewModel.step()  // R0 changes
    XCTAssertNotNil(viewModel.registerHighlights["R0"])

    try await Task.sleep(nanoseconds: 500_000_000)
    await viewModel.step()  // R1 changes

    // Both should be highlighted simultaneously
    XCTAssertNotNil(viewModel.registerHighlights["R0"])
    XCTAssertNotNil(viewModel.registerHighlights["R1"])
}
```

### Manual Testing

Run `fibonacci.s` with stepping to verify:
1. Multiple registers highlight simultaneously when changed in rapid succession
2. Highlights fade smoothly over 1.5 seconds with ease-out curve
3. Rapid changes to the same register restart the timer correctly
4. Memory writes show per-byte highlights (1, 2, or 4 bytes depending on write size)
5. No performance degradation with many concurrent highlights

## Implementation Files

### Files to Modify

1. **swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift**
   - Add `registerHighlights`, `memoryHighlights`, task tracking
   - Implement `highlightRegister()` and `highlightMemoryAddress()`
   - Modify `updateRegisters()` to use new highlight system
   - Add computed property for backward compatibility

2. **swift-gui/ARMEmulator/Views/RegistersView.swift**
   - Change parameter from `changedRegisters: Set<String>` to `registerHighlights: [String: UUID]`
   - Update `RegisterRow` to accept `highlightID: UUID?`
   - Add `.animation(.easeOut(duration: 1.5), value: highlightID)` modifiers

3. **swift-gui/ARMEmulator/Views/MemoryView.swift**
   - Update `onChange(of: viewModel.lastMemoryWrite)` to call `highlightMemoryAddress()`
   - Pass `memoryHighlights` map to `MemoryRowView`

4. **swift-gui/ARMEmulator/Views/MemoryView.swift** (MemoryRowView)
   - Change from single `highlightedWriteAddress` to `memoryHighlights` map
   - Check map for each byte and apply animation modifier
   - Handle multi-byte writes (1, 2, 4 bytes)

5. **swift-gui/ARMEmulator/Views/MainView.swift**
   - Update `RegistersView` instantiation to pass `registerHighlights`

### Files to Create

1. **swift-gui/ARMEmulatorTests/ViewModels/HighlightTests.swift**
   - Unit tests for highlight lifecycle
   - Timer restart tests
   - Multiple concurrent highlight tests

## Migration Path

### Phase 1: Add New System (Non-Breaking)
- Add `registerHighlights` and `memoryHighlights` to ViewModel
- Implement `highlightRegister()` and `highlightMemoryAddress()`
- Keep `changedRegisters` computed property
- No view changes yet

### Phase 2: Update Views
- Modify `RegistersView` and `MemoryView` to use new system
- Add animation modifiers
- Test that fading works correctly

### Phase 3: Cleanup
- Remove old `changedRegisters` stored property (keep computed if needed elsewhere)
- Remove old `highlightedWriteAddress` state variable
- Remove any dead code

## Edge Cases

### Rapid Execution
During "Run" mode (not stepping), hundreds of changes per second:
- Each change creates a task (overhead negligible)
- Highlights update only when published values change (SwiftUI coalesces)
- Visual result: registers flash green continuously (expected behavior)

### Memory Pressure
If thousands of memory addresses are highlighted:
- Each has a 1.5s task (auto-cleaned)
- Dictionary overhead is minimal (<1KB for 1000 entries)
- SwiftUI only renders visible rows (no performance issue)

### Session Reset
When loading a new program or resetting:
- Cancel all highlight tasks explicitly in `loadProgram()` and `restart()`
- Clear highlight dictionaries
- Prevents stale highlights from previous execution

## Benefits

1. **Better debugging experience**: See multiple register changes simultaneously
2. **Visual tracking**: Easier to follow data flow through registers
3. **No lost information**: Rapid changes don't hide previous changes
4. **Smooth animations**: Native SwiftUI animations feel polished
5. **Minimal complexity**: Animation-driven approach is simpler than manual timers

## Alternative Approaches Considered

### Dictionary with Timestamps (Rejected)
Store `[RegisterName: Date]` and calculate opacity on each render. Rejected because:
- Requires manual opacity calculation and color interpolation
- Frame-rate dependent (inconsistent on slow machines)
- More complex implementation

### Timer-Based Opacity Updates (Rejected)
Single timer firing every 60ms to update opacity values. Rejected because:
- Adds timer management complexity
- Doesn't leverage SwiftUI's animation system
- Harder to test (timing dependencies)

### Animation-Driven (Selected)
Store triggers and let SwiftUI handle fade. Selected because:
- SwiftUI-native and optimized
- Built-in ease-out curve
- Automatic cleanup and memory management
- Simple implementation

## Success Criteria

- [ ] Multiple registers can be highlighted simultaneously
- [ ] Highlights fade over 1.5 seconds with ease-out curve
- [ ] Rapid changes to the same item restart the timer
- [ ] Memory writes highlight 1, 2, or 4 bytes correctly
- [ ] No performance degradation during rapid execution
- [ ] All unit tests pass
- [ ] Manual testing with `fibonacci.s` shows smooth fading
