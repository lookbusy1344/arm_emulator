# Fading Highlights Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement time-based fading highlights for register and memory changes in Swift GUI with 1.5 second ease-out animations.

**Architecture:** Add UUID-based highlight tracking to EmulatorViewModel with Task-based cleanup timers. Views receive highlight UUIDs and use SwiftUI's `.animation()` modifier to automatically fade colors over 1.5 seconds.

**Tech Stack:** SwiftUI animations, Swift async/await tasks, XCTest for unit tests

---

## Task 1: Add ViewModel Highlight Properties

**Files:**
- Modify: `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift:5-25`
- Test: `swift-gui/ARMEmulatorTests/ViewModels/EmulatorViewModelTests.swift`

**Step 1: Write failing test for register highlight tracking**

Create new test file if it doesn't exist. Add test:

```swift
@MainActor
final class HighlightTests: XCTestCase {
    var viewModel: EmulatorViewModel!

    override func setUp() async throws {
        viewModel = EmulatorViewModel(
            apiClient: MockAPIClient(),
            wsClient: MockWebSocketClient()
        )
    }

    func testRegisterHighlightAdded() {
        viewModel.highlightRegister("R0")
        XCTAssertNotNil(viewModel.registerHighlights["R0"])
    }

    func testMemoryHighlightAdded() {
        viewModel.highlightMemoryAddress(0x8000, size: 1)
        XCTAssertNotNil(viewModel.memoryHighlights[0x8000])
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd swift-gui && xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify`

Expected: FAIL with "Value of type 'EmulatorViewModel' has no member 'highlightRegister'"

**Step 3: Add highlight properties to EmulatorViewModel**

In `EmulatorViewModel.swift`, add after line 8 (`@Published var changedRegisters`):

```swift
// Highlight tracking with UUIDs for animation
@Published var registerHighlights: [String: UUID] = [:]
@Published var memoryHighlights: [UInt32: UUID] = [:]

// Task tracking for cleanup
private var highlightTasks: [String: Task<Void, Never>] = [:]
private var memoryHighlightTasks: [UInt32: Task<Void, Never>] = [:]
```

**Step 4: Implement highlightRegister() function**

Add before the `init()` function in `EmulatorViewModel.swift`:

```swift
func highlightRegister(_ name: String) {
    // Cancel existing fade task for this register
    highlightTasks[name]?.cancel()

    // Add new highlight (triggers animation to green)
    registerHighlights[name] = UUID()

    // Schedule removal after 1.5 seconds
    highlightTasks[name] = Task { @MainActor in
        try? await Task.sleep(nanoseconds: 1_500_000_000)
        registerHighlights[name] = nil
        highlightTasks[name] = nil
    }
}

func highlightMemoryAddress(_ address: UInt32, size: UInt32) {
    // Highlight each byte in the write
    for offset in 0..<size {
        let addr = address + offset

        // Cancel existing fade task
        memoryHighlightTasks[addr]?.cancel()

        // Add new highlight
        memoryHighlights[addr] = UUID()

        // Schedule removal after 1.5 seconds
        memoryHighlightTasks[addr] = Task { @MainActor in
            try? await Task.sleep(nanoseconds: 1_500_000_000)
            memoryHighlights[addr] = nil
            memoryHighlightTasks[addr] = nil
        }
    }
}
```

**Step 5: Run tests to verify they pass**

Run: `cd swift-gui && xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify`

Expected: PASS for `testRegisterHighlightAdded` and `testMemoryHighlightAdded`

**Step 6: Commit**

```bash
git add swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift swift-gui/ARMEmulatorTests/ViewModels/EmulatorViewModelTests.swift
git commit -m "feat: add highlight tracking with UUID-based timers

Add registerHighlights and memoryHighlights dictionaries to track
which items should be highlighted. Each highlight gets a unique UUID
and a 1.5s cleanup task.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Test Highlight Timer Behavior

**Files:**
- Test: `swift-gui/ARMEmulatorTests/ViewModels/EmulatorViewModelTests.swift`

**Step 1: Write failing test for highlight fade timing**

Add to `HighlightTests`:

```swift
func testRegisterHighlightFadesAfterDelay() async throws {
    viewModel.highlightRegister("R0")

    // Should be highlighted immediately
    XCTAssertNotNil(viewModel.registerHighlights["R0"])

    // Wait for fade to complete
    try await Task.sleep(nanoseconds: 1_600_000_000)  // 1.6s

    // Should be removed
    XCTAssertNil(viewModel.registerHighlights["R0"])
}

func testRapidChangesRestartTimer() async throws {
    viewModel.highlightRegister("R0")

    // Wait halfway through fade
    try await Task.sleep(nanoseconds: 500_000_000)  // 0.5s

    // Trigger another change (should restart timer)
    viewModel.highlightRegister("R0")

    // Wait 1.2s (0.7s after restart)
    try await Task.sleep(nanoseconds: 1_200_000_000)

    // Should still be highlighted
    XCTAssertNotNil(viewModel.registerHighlights["R0"])

    // Wait another 0.5s (1.2s after restart, past 1.5s threshold)
    try await Task.sleep(nanoseconds: 500_000_000)

    // Should be removed now
    XCTAssertNil(viewModel.registerHighlights["R0"])
}

func testMultipleRegisterHighlightsIndependent() async throws {
    viewModel.highlightRegister("R0")

    try await Task.sleep(nanoseconds: 500_000_000)  // 0.5s

    viewModel.highlightRegister("R1")

    // Both should be highlighted
    XCTAssertNotNil(viewModel.registerHighlights["R0"])
    XCTAssertNotNil(viewModel.registerHighlights["R1"])

    // Wait for R0 to fade (1.2s more = 1.7s total)
    try await Task.sleep(nanoseconds: 1_200_000_000)

    // R0 should be gone, R1 still visible
    XCTAssertNil(viewModel.registerHighlights["R0"])
    XCTAssertNotNil(viewModel.registerHighlights["R1"])
}

func testMemoryHighlightMultipleBytes() {
    viewModel.highlightMemoryAddress(0x8000, size: 4)

    // All 4 bytes should be highlighted
    XCTAssertNotNil(viewModel.memoryHighlights[0x8000])
    XCTAssertNotNil(viewModel.memoryHighlights[0x8001])
    XCTAssertNotNil(viewModel.memoryHighlights[0x8002])
    XCTAssertNotNil(viewModel.memoryHighlights[0x8003])
    XCTAssertNil(viewModel.memoryHighlights[0x8004])  // 5th byte not written
}
```

**Step 2: Run tests to verify they pass**

Run: `cd swift-gui && xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify`

Expected: PASS for all new timer tests (implementation from Task 1 should already handle these correctly)

**Step 3: Commit**

```bash
git add swift-gui/ARMEmulatorTests/ViewModels/EmulatorViewModelTests.swift
git commit -m "test: add highlight timer behavior tests

Test that highlights fade after 1.5s, rapid changes restart timers,
and multiple highlights work independently.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Integrate Highlights into updateRegisters()

**Files:**
- Modify: `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift:303-315`
- Test: Manual verification with stepping

**Step 1: Write integration test**

Add to test file:

```swift
func testUpdateRegistersTriggersHighlights() async throws {
    // Set up mock API client with register data
    let mockClient = MockAPIClient()
    viewModel = EmulatorViewModel(apiClient: mockClient, wsClient: MockWebSocketClient())

    // Simulate first state
    let registers1 = RegisterState(
        r0: 0, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
        r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
        sp: 0x50000, lr: 0, pc: 0x8000,
        cpsr: CPSRFlags(n: false, z: false, c: false, v: false)
    )
    viewModel.updateRegisters(registers1)

    // Simulate second state with R0, R1 changed
    let registers2 = RegisterState(
        r0: 42, r1: 100, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
        r8: 0, r9: 0, r10: 0, r11: 0, r12: 0,
        sp: 0x50000, lr: 0, pc: 0x8004,
        cpsr: CPSRFlags(n: false, z: false, c: false, v: false)
    )
    viewModel.updateRegisters(registers2)

    // R0 and R1 should be highlighted, PC should be highlighted
    XCTAssertNotNil(viewModel.registerHighlights["R0"])
    XCTAssertNotNil(viewModel.registerHighlights["R1"])
    XCTAssertNotNil(viewModel.registerHighlights["PC"])
    XCTAssertNil(viewModel.registerHighlights["R2"])  // Unchanged
}
```

**Step 2: Run test to verify it fails**

Run: `cd swift-gui && xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify`

Expected: FAIL with "XCTAssertNotNil failed" - highlights not triggered

**Step 3: Modify updateRegisters() to call highlightRegister()**

Find `updateRegisters()` in `EmulatorViewModel.swift` (around line 303) and modify:

```swift
func updateRegisters(_ newRegisters: RegisterState) {
    // Track changes
    var changed = Set<String>()

    if let prev = previousRegisters {
        changed = detectRegisterChanges(previous: prev, new: newRegisters)
    }

    // Highlight each changed register (independent timers)
    for registerName in changed {
        highlightRegister(registerName)
    }

    previousRegisters = registers
    registers = newRegisters
    currentPC = newRegisters.pc
}
```

**Step 4: Run test to verify it passes**

Run: `cd swift-gui && xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify`

Expected: PASS for `testUpdateRegistersTriggersHighlights`

**Step 5: Commit**

```bash
git add swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift swift-gui/ARMEmulatorTests/ViewModels/EmulatorViewModelTests.swift
git commit -m "feat: integrate highlights into updateRegisters()

Call highlightRegister() for each changed register to trigger
1.5s fade animation.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: Update RegistersView to Accept Highlights Map

**Files:**
- Modify: `swift-gui/ARMEmulator/Views/RegistersView.swift:3-10`
- Modify: `swift-gui/ARMEmulator/Views/MainView.swift` (RegistersView instantiation)

**Step 1: Modify RegistersView to accept highlights**

In `RegistersView.swift`, change the properties and initializer:

```swift
struct RegistersView: View {
    let registers: RegisterState
    let registerHighlights: [String: UUID]  // Changed from changedRegisters

    init(registers: RegisterState, registerHighlights: [String: UUID] = [:]) {
        self.registers = registers
        self.registerHighlights = registerHighlights
    }

    // ... rest of struct
```

**Step 2: Update RegisterRow calls to pass highlightID**

In `RegistersView.swift`, update all `RegisterRow` calls (lines 39-62) to pass the highlight UUID instead of boolean:

```swift
// General-purpose registers
LazyVGrid(columns: gridColumns(for: geometry.size.width), alignment: .leading, spacing: 4) {
    RegisterRow(name: "R0", value: registers.r0, highlightID: registerHighlights["R0"])
    RegisterRow(name: "R1", value: registers.r1, highlightID: registerHighlights["R1"])
    RegisterRow(name: "R2", value: registers.r2, highlightID: registerHighlights["R2"])
    RegisterRow(name: "R3", value: registers.r3, highlightID: registerHighlights["R3"])
    RegisterRow(name: "R4", value: registers.r4, highlightID: registerHighlights["R4"])
    RegisterRow(name: "R5", value: registers.r5, highlightID: registerHighlights["R5"])
    RegisterRow(name: "R6", value: registers.r6, highlightID: registerHighlights["R6"])
    RegisterRow(name: "R7", value: registers.r7, highlightID: registerHighlights["R7"])
    RegisterRow(name: "R8", value: registers.r8, highlightID: registerHighlights["R8"])
    RegisterRow(name: "R9", value: registers.r9, highlightID: registerHighlights["R9"])
    RegisterRow(name: "R10", value: registers.r10, highlightID: registerHighlights["R10"])
    RegisterRow(name: "R11", value: registers.r11, highlightID: registerHighlights["R11"])
    RegisterRow(name: "R12", value: registers.r12, highlightID: registerHighlights["R12"])
}
.padding(.horizontal, 8)

Divider()
    .padding(.vertical, 4)

// Special registers
LazyVGrid(columns: gridColumns(for: geometry.size.width), alignment: .leading, spacing: 4) {
    RegisterRow(name: "SP", value: registers.sp, highlightID: registerHighlights["SP"])
    RegisterRow(name: "LR", value: registers.lr, highlightID: registerHighlights["LR"])
    RegisterRow(name: "PC", value: registers.pc, highlightID: registerHighlights["PC"])
}
.padding(.horizontal, 8)
```

**Step 3: Update CPSR highlighting**

Around line 70-78, update CPSR to use UUID:

```swift
HStack {
    Text("CPSR:")
        .font(.system(size: 10, design: .monospaced))
        .fontWeight(.bold)
        .frame(width: 60, alignment: .leading)

    Text(registers.cpsr.displayString)
        .font(.system(size: 10, design: .monospaced))
        .foregroundColor(registerHighlights["CPSR"] != nil ? .green : .primary)
}
.padding(.horizontal)
.padding(.vertical, 2)
```

**Step 4: Update MainView to pass registerHighlights**

Find the `RegistersView` instantiation in `MainView.swift` and update:

```swift
RegistersView(
    registers: viewModel.registers,
    registerHighlights: viewModel.registerHighlights
)
```

**Step 5: Build to verify no compile errors**

Run: `cd swift-gui && xcodebuild build -project ARMEmulator.xcodeproj -scheme ARMEmulator | xcbeautify`

Expected: Build succeeds (RegisterRow will have compile errors until next task)

**Step 6: Commit**

```bash
git add swift-gui/ARMEmulator/Views/RegistersView.swift swift-gui/ARMEmulator/Views/MainView.swift
git commit -m "refactor: update RegistersView to use highlights map

Change from boolean isChanged to UUID-based highlightID for
animation support.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Update RegisterRow with Animation

**Files:**
- Modify: `swift-gui/ARMEmulator/Views/RegistersView.swift:90-122`

**Step 1: Update RegisterRow signature**

Change `RegisterRow` struct (around line 90):

```swift
struct RegisterRow: View {
    let name: String
    let value: UInt32
    let highlightID: UUID?  // Changed from isChanged: Bool

    init(name: String, value: UInt32, highlightID: UUID? = nil) {
        self.name = name
        self.value = value
        self.highlightID = highlightID
    }

    var body: some View {
        HStack {
            Text("\(name):")
                .font(.system(size: 10, design: .monospaced))
                .fontWeight(.bold)
                .frame(width: 60, alignment: .leading)
                .foregroundColor(highlightID != nil ? .green : .primary)
                .animation(.easeOut(duration: 1.5), value: highlightID)

            Text(String(format: "0x%08X", value))
                .font(.system(size: 10, design: .monospaced))
                .foregroundColor(highlightID != nil ? .green : .primary)
                .animation(.easeOut(duration: 1.5), value: highlightID)

            Spacer()

            Text(String(value))
                .font(.system(size: 10, design: .monospaced))
                .foregroundColor(highlightID != nil ? .green : .secondary)
                .animation(.easeOut(duration: 1.5), value: highlightID)
        }
        .padding(.horizontal)
        .padding(.vertical, 2)
    }
}
```

**Step 2: Update preview to use highlightID**

Update the preview at the bottom of the file:

```swift
struct RegistersView_Previews: PreviewProvider {
    static var previews: some View {
        RegistersView(
            registers: RegisterState(
                r0: 0x0000_0042, r1: 0x0000_0001, r2: 0x0000_0002, r3: 0x0000_0003,
                r4: 0, r5: 0, r6: 0, r7: 0,
                r8: 0, r9: 0, r10: 0, r11: 0,
                r12: 0, sp: 0x0005_0000, lr: 0, pc: 0x0000_8004,
                cpsr: CPSRFlags(n: false, z: false, c: true, v: false)
            ),
            registerHighlights: ["R0": UUID(), "PC": UUID()]  // Show R0 and PC highlighted
        )
        .frame(width: 300, height: 500)
    }
}
```

**Step 3: Build and verify no errors**

Run: `cd swift-gui && xcodebuild build -project ARMEmulator.xcodeproj -scheme ARMEmulator | xcbeautify`

Expected: Build succeeds

**Step 4: Commit**

```bash
git add swift-gui/ARMEmulator/Views/RegistersView.swift
git commit -m "feat: add fade animations to RegisterRow

Use .animation(.easeOut(duration: 1.5), value: highlightID) to
automatically fade register highlights.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 6: Update MemoryView to Call highlightMemoryAddress()

**Files:**
- Modify: `swift-gui/ARMEmulator/Views/MemoryView.swift:94-129`

**Step 1: Update onChange handler**

Find the `.onChange(of: viewModel.lastMemoryWrite)` handler (around line 94) and modify:

```swift
.onChange(of: viewModel.lastMemoryWrite) {
    // Handle memory write highlighting and auto-scroll
    guard let writeAddr = viewModel.lastMemoryWrite else {
        return
    }

    Task {
        // Highlight the written address (triggers 1.5s fade)
        viewModel.highlightMemoryAddress(writeAddr, size: viewModel.lastMemoryWriteSize)

        // Check if write is within currently visible range
        let visibleEnd = baseAddress + UInt32(totalBytes)
        let isVisible = writeAddr >= baseAddress && writeAddr < visibleEnd

        if autoScrollEnabled && !isVisible {
            // Write is outside visible range - scroll to it
            let alignedAddress = writeAddr & ~UInt32(0xF)
            await loadMemoryAsync(at: alignedAddress)
        } else {
            // Write is visible - just refresh data in place
            await refreshMemoryAsync()
        }

        // Trigger scroll to the row containing the write (if auto-scroll enabled)
        if autoScrollEnabled {
            let rowOffset = Int((writeAddr - baseAddress) / UInt32(bytesPerRow))
            if rowOffset >= 0 && rowOffset < rowsToShow && rowOffset != lastScrolledRow {
                scrollToRow = rowOffset
                lastScrolledRow = rowOffset
            }
        }
    }
}
```

**Step 2: Remove old highlightedWriteAddress state**

Remove the `@State private var highlightedWriteAddress: UInt32?` declaration (around line 9) as it's no longer needed.

**Step 3: Build to verify**

Run: `cd swift-gui && xcodebuild build -project ARMEmulator.xcodeproj -scheme ARMEmulator | xcbeautify`

Expected: Build succeeds (MemoryRowView will have errors until next task)

**Step 4: Commit**

```bash
git add swift-gui/ARMEmulator/Views/MemoryView.swift
git commit -m "refactor: use highlightMemoryAddress() in onChange

Call viewModel.highlightMemoryAddress() to trigger timed fades
instead of setting single highlightedWriteAddress.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 7: Update MemoryRowView with Animation

**Files:**
- Modify: `swift-gui/ARMEmulator/Views/MemoryView.swift:193-279` and `283-334`

**Step 1: Update MemoryRowView signature**

Change `MemoryRowView` struct (around line 193):

```swift
struct MemoryRowView: View {
    let address: UInt32
    let bytes: [UInt8]
    let highlightAddress: UInt32?  // PC highlight (keep for now)
    let memoryHighlights: [UInt32: UUID]  // Changed from lastWriteAddress
    let lastWriteSize: UInt32  // Keep for compatibility

    private var isHighlighted: Bool {
        guard let highlight = highlightAddress else { return false }
        return address <= highlight && highlight < address + UInt32(bytes.count)
    }

    private func highlightID(for byteIndex: Int) -> UUID? {
        let byteAddr = address + UInt32(byteIndex)
        return memoryHighlights[byteAddr]
    }

    var body: some View {
        HStack(spacing: 4) {
            // Address
            Text(String(format: "0x%08X", address))
                .foregroundColor(Color(red: 0.5, green: 0.6, blue: 0.7))
                .frame(width: 75, alignment: .leading)

            // Hex bytes
            HStack(spacing: 2) {
                ForEach(0 ..< 16) { i in
                    if i < bytes.count {
                        let highlight = highlightID(for: i)
                        Text(String(format: "%02X", bytes[i]))
                            .frame(width: 20)
                            .foregroundColor(highlight != nil ? .white : .primary)
                            .fontWeight(highlight != nil ? .bold : .regular)
                            .background(highlight != nil ? Color.green : Color.clear)
                            .cornerRadius(2)
                            .animation(.easeOut(duration: 1.5), value: highlight)
                    } else {
                        Text("  ")
                            .frame(width: 20)
                    }
                }
            }

            // ASCII representation
            HStack(spacing: 0) {
                ForEach(0 ..< 16) { i in
                    if i < bytes.count {
                        let char = bytes[i]
                        let displayChar = (32 ... 126).contains(char) ? String(UnicodeScalar(char)) : "."
                        let highlight = highlightID(for: i)
                        Text(displayChar)
                            .frame(width: 10)
                            .foregroundColor(highlight != nil ? .white : .primary)
                            .fontWeight(highlight != nil ? .bold : .regular)
                            .background(highlight != nil ? Color.green : Color.clear)
                            .cornerRadius(2)
                            .animation(.easeOut(duration: 1.5), value: highlight)
                    } else {
                        Text(" ")
                            .frame(width: 10)
                    }
                }
            }

            Spacer()
        }
        .padding(.vertical, 2)
        .background(isHighlighted ? Color.accentColor.opacity(0.2) : Color.clear)
        .cornerRadius(2)
    }
}
```

**Step 2: Update MemoryDisplayView to pass highlights**

Find `MemoryDisplayView` (around line 283) and update the `MemoryRowView` instantiation:

```swift
struct MemoryDisplayView: View {
    let rowsToShow: Int
    let bytesPerRow: Int
    let baseAddress: UInt32
    let memoryData: [UInt8]
    @ObservedObject var viewModel: EmulatorViewModel
    let scrollToRow: Int?
    let refreshID: UUID

    var body: some View {
        ScrollView([.vertical, .horizontal]) {
            ScrollViewReader { proxy in
                VStack(alignment: .leading, spacing: 0) {
                    ForEach(0 ..< rowsToShow, id: \.self) { row in
                        MemoryRowView(
                            address: baseAddress + UInt32(row * bytesPerRow),
                            bytes: bytesForRow(row),
                            highlightAddress: viewModel.currentPC,
                            memoryHighlights: viewModel.memoryHighlights,  // Pass highlights map
                            lastWriteSize: viewModel.lastMemoryWriteSize
                        )
                        .id("row_\(row)")
                    }
                }
                .font(.system(size: 10, design: .monospaced))
                .padding(.vertical, 8)
                .padding(.horizontal, 4)
                .task(id: scrollToRow) {
                    if let row = scrollToRow {
                        try? await Task.sleep(nanoseconds: 10_000_000)
                        withAnimation(.easeInOut(duration: 0.3)) {
                            proxy.scrollTo("row_\(row)", anchor: .center)
                        }
                    }
                }
            }
        }
        .id(refreshID)
    }

    private func bytesForRow(_ row: Int) -> [UInt8] {
        let startIndex = row * bytesPerRow
        let endIndex = min((row + 1) * bytesPerRow, memoryData.count)

        guard startIndex < memoryData.count else {
            return []
        }

        return Array(memoryData[startIndex ..< endIndex])
    }
}
```

**Step 3: Update MemoryView to pass memoryHighlights**

Find the `MemoryDisplayView` call in `MemoryView` and update:

```swift
MemoryDisplayView(
    rowsToShow: rowsToShow,
    bytesPerRow: bytesPerRow,
    baseAddress: baseAddress,
    memoryData: memoryData,
    viewModel: viewModel,
    scrollToRow: scrollToRow,
    refreshID: refreshID
)
```

**Step 4: Build and verify no errors**

Run: `cd swift-gui && xcodebuild build -project ARMEmulator.xcodeproj -scheme ARMEmulator | xcbeautify`

Expected: Build succeeds

**Step 5: Commit**

```bash
git add swift-gui/ARMEmulator/Views/MemoryView.swift
git commit -m "feat: add fade animations to memory highlights

Use memoryHighlights map with per-byte UUID tracking and
.animation(.easeOut(duration: 1.5)) for smooth fades.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Add Session Cleanup

**Files:**
- Modify: `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift` (loadProgram and restart methods)

**Step 1: Add cleanup to loadProgram()**

Find `loadProgram()` method and add cleanup at the start:

```swift
func loadProgram(source: String) async {
    DebugLog.log("loadProgram() called", category: "ViewModel")
    DebugLog.log("Source length: \(source.count) chars", category: "ViewModel")

    // Clear highlights when loading new program
    cancelAllHighlights()

    guard let sessionID = sessionID else {
        // ... rest of function
```

**Step 2: Add cleanup to restart()**

Find `restart()` method and add cleanup:

```swift
func restart() async {
    DebugLog.log("restart() called", category: "ViewModel")

    // Clear highlights when restarting
    cancelAllHighlights()

    guard let sessionID = sessionID else {
        // ... rest of function
```

**Step 3: Implement cancelAllHighlights()**

Add this private method to `EmulatorViewModel`:

```swift
private func cancelAllHighlights() {
    // Cancel all pending highlight tasks
    for task in highlightTasks.values {
        task.cancel()
    }
    for task in memoryHighlightTasks.values {
        task.cancel()
    }

    // Clear all highlights
    highlightTasks.removeAll()
    memoryHighlightTasks.removeAll()
    registerHighlights.removeAll()
    memoryHighlights.removeAll()
}
```

**Step 4: Build and verify**

Run: `cd swift-gui && xcodebuild build -project ARMEmulator.xcodeproj -scheme ARMEmulator | xcbeautify`

Expected: Build succeeds

**Step 5: Commit**

```bash
git add swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift
git commit -m "feat: cancel highlights on program load and restart

Clear all highlight tasks and dictionaries when loading new program
or restarting to prevent stale highlights.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 9: Manual Testing

**Files:**
- Test: Manual testing with `examples/fibonacci.s`

**Step 1: Build the app**

Run: `cd swift-gui && xcodebuild build -project ARMEmulator.xcodeproj -scheme ARMEmulator -configuration Debug | xcbeautify`

Expected: Build succeeds

**Step 2: Launch app and load fibonacci.s**

Run:
```bash
find ~/Library/Developer/Xcode/DerivedData -name "ARMEmulator.app" -type d -exec open {} \; -quit
```

Then load `examples/fibonacci.s`

**Step 3: Test single step highlights**

- Click "Step" button
- Verify: Changed registers turn green
- Wait 1.5 seconds
- Verify: Green fades smoothly back to normal color

**Step 4: Test rapid stepping**

- Click "Step" several times quickly (< 0.5s between clicks)
- Verify: Multiple registers can be green simultaneously
- Verify: If same register changes twice, green "restarts" (stays bright)
- Wait 1.5 seconds after last step
- Verify: All highlights fade

**Step 5: Test memory highlights**

- Load a program with memory writes (e.g., `examples/arrays.s`)
- Step through execution
- Verify: Memory bytes turn green when written
- Verify: Highlights fade after 1.5 seconds
- Verify: Multi-byte writes (STR = 4 bytes) highlight all bytes

**Step 6: Test program reload**

- Step through a few instructions (create highlights)
- Load a different program
- Verify: All highlights clear immediately

**Step 7: Document results**

Create test report in commit message.

**Step 8: Commit**

```bash
git commit --allow-empty -m "test: manual testing of fading highlights

Verified:
- Single step highlights fade over 1.5s with ease-out curve
- Rapid steps show multiple simultaneous highlights
- Same register changing twice restarts timer
- Memory highlights fade correctly (1, 2, 4 byte writes)
- Program reload clears all highlights

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 10: Run All Tests

**Step 1: Run full test suite**

Run: `cd swift-gui && xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify`

Expected: All tests pass

**Step 2: Check for warnings**

Review build output for any warnings. Fix if found.

**Step 3: Run linter**

Run: `cd swift-gui && swiftlint`

Expected: No violations

**Step 4: Format code**

Run: `cd swift-gui && swiftformat .`

Expected: No changes (code already formatted)

**Step 5: Final commit if any fixes**

```bash
git add .
git commit -m "fix: address lint warnings and formatting

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Success Criteria Checklist

After completing all tasks, verify:

- [ ] Multiple registers can be highlighted simultaneously
- [ ] Highlights fade over 1.5 seconds with ease-out curve
- [ ] Rapid changes to same register restart the timer (stays highlighted for 1.5s after *last* change)
- [ ] Memory writes highlight 1, 2, or 4 bytes correctly
- [ ] Highlights clear when loading new program or restarting
- [ ] All unit tests pass
- [ ] Manual testing with `fibonacci.s` shows smooth fading
- [ ] No SwiftLint violations
- [ ] No compiler warnings
- [ ] Code is formatted with SwiftFormat

---

## Notes

**Testing Challenges:**
- SwiftUI animation testing is difficult - rely on manual verification for visual behavior
- Use async tests with `Task.sleep()` to verify timer-based cleanup
- Unit tests focus on data model correctness, manual tests verify visual behavior

**Performance:**
- Each highlight creates a 1.5s Task (negligible overhead)
- SwiftUI coalesces published changes (no frame rate issues)
- Dictionary lookups are O(1) for highlight checks

**Mock Requirements:**
If `MockAPIClient` or `MockWebSocketClient` don't exist, create minimal mocks:

```swift
class MockAPIClient: APIClient {
    override func createSession() async throws -> String { "mock-session" }
    override func getRegisters(sessionID: String) async throws -> RegisterState {
        RegisterState(r0: 0, r1: 0, r2: 0, r3: 0, r4: 0, r5: 0, r6: 0, r7: 0,
                     r8: 0, r9: 0, r10: 0, r11: 0, r12: 0, sp: 0x50000, lr: 0, pc: 0x8000,
                     cpsr: CPSRFlags(n: false, z: false, c: false, v: false))
    }
}

class MockWebSocketClient: WebSocketClient {
    override func connect(sessionID: String) {}
}
```
