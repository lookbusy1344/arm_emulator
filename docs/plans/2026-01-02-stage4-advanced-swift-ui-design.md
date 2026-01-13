# Stage 4: Advanced Swift UI Features - Design Document

**Date:** 2026-01-02
**Author:** Claude (with user collaboration)
**Status:** Approved for implementation

## Overview

This document details the design for Stage 4 of the Swift macOS GUI, which adds advanced UI features to achieve feature parity with the Wails GUI. The stage implements 9 major features organized into 3 implementation phases.

## Design Decisions Summary

| Feature | Approach | Rationale |
|---------|----------|-----------|
| **Layout** | Tabbed right panel | Simple, predictable, keeps current layout familiar |
| **Examples Browser** | Sheet modal with preview | Good discoverability, doesn't complicate main UI |
| **Syntax Highlighting** | Custom NSTextView with TextKit | Native feel, full control, no dependencies |
| **Breakpoints** | Click gutter to toggle | Industry standard, meets user expectations |
| **Preferences** | Essential settings only (4-5) | YAGNI principle, can expand later |
| **Memory View** | Address input + auto-scroll | Flexibility with convenient shortcuts |
| **Stack View** | Specialized vertical inspector | Stack-specific features aid debugging |
| **Disassembly** | Live with PC tracking | Leverages existing API, professional debugger UX |

## Architecture

### File Structure

**New Files (9 total):**

**Views (5 files):**
- `swift-gui/ARMEmulator/Views/MemoryView.swift` - Hex dump viewer
- `swift-gui/ARMEmulator/Views/StackView.swift` - Stack visualization
- `swift-gui/ARMEmulator/Views/DisassemblyView.swift` - Live disassembly
- `swift-gui/ARMEmulator/Views/ExamplesBrowserView.swift` - Examples picker
- `swift-gui/ARMEmulator/Views/PreferencesView.swift` - Settings window

**Editor Components (2 files):**
- `swift-gui/ARMEmulator/Views/SyntaxHighlightingTextView.swift` - Custom editor
- `swift-gui/ARMEmulator/Views/BreakpointGutter.swift` - Gutter rendering/interaction

**Services (1 file):**
- `swift-gui/ARMEmulator/Services/FileService.swift` - File I/O, recent files

**Models (1 file):**
- `swift-gui/ARMEmulator/Models/AppSettings.swift` - Preferences storage

**Modified Files (3):**
- `swift-gui/ARMEmulator/Views/MainView.swift` - Add tabs, menus
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift` - Memory/stack/disassembly state
- `swift-gui/ARMEmulator/Services/APIClient.swift` - New API methods

### Component Interactions

```
MainView
├── EditorView (SyntaxHighlightingTextView + BreakpointGutter)
├── TabView (Right Panel)
│   ├── RegistersView + StatusView [Tab 1]
│   ├── MemoryView [Tab 2]
│   ├── StackView [Tab 3]
│   └── DisassemblyView [Tab 4]
└── ConsoleView

FileService ←→ MainView (Open/Save/Recent/Examples)
AppSettings ←→ PreferencesView (Persist via UserDefaults)
EmulatorViewModel ←→ All Views (Shared state)
APIClient ←→ EmulatorViewModel (Backend communication)
```

## Feature Designs

### 1. Tabbed Right Panel

**Implementation:**
Replace current right panel VStack with TabView containing 4 tabs:

```swift
TabView {
    // Tab 1: Registers (existing)
    VStack(spacing: 0) {
        RegistersView(registers: viewModel.registers)
        Divider()
        StatusView(status: viewModel.status, pc: viewModel.currentPC)
    }
    .tabItem { Label("Registers", systemImage: "cpu") }

    // Tab 2: Memory
    MemoryView(viewModel: viewModel)
        .tabItem { Label("Memory", systemImage: "memorychip") }

    // Tab 3: Stack
    StackView(viewModel: viewModel)
        .tabItem { Label("Stack", systemImage: "square.stack.3d.down.right") }

    // Tab 4: Disassembly
    DisassemblyView(viewModel: viewModel)
        .tabItem { Label("Disassembly", systemImage: "hammer") }
}
.frame(minWidth: 250, maxWidth: 500)
```

**Tab Persistence:**
- Use `@AppStorage("selectedTab")` to remember last selected tab
- Restores user's preferred view on app relaunch

### 2. File Management

**FileService.swift:**

```swift
class FileService: ObservableObject {
    @Published var recentFiles: [URL] = []
    private let maxRecentFiles = 10

    func openFile() async -> String?
    func saveFile(content: String, suggestedName: String) async -> Bool
    func loadExamples() -> [ExampleProgram]
    func addToRecentFiles(_ url: URL)
    private func persistRecentFiles()
    private func loadRecentFiles()
}

struct ExampleProgram: Identifiable {
    let id = UUID()
    let name: String
    let filename: String
    let description: String
    let size: Int
    let content: String
}
```

**Menu Integration:**
- File → Open... (⌘O) → NSOpenPanel for .s files
- File → Save... (⌘S) → NSSavePanel
- File → Open Recent → Dynamic submenu (last 10 files, "Clear Menu" option)
- File → Open Example... → Shows ExamplesBrowserView sheet

**ExamplesBrowserView:**
- Modal sheet with search field
- List of examples with: filename, description (parsed from first comment), size
- Preview pane showing first 10 lines
- Double-click or "Open" button loads example
- Search filters by filename or description text

### 3. Memory Hex Dump View

**MemoryView.swift:**

**UI Layout:**
```
┌─────────────────────────────────────────────────┐
│ Address: [0x00008000] [Go] [PC] [SP] [R0]...[R15]│
├─────────────────────────────────────────────────┤
│ 0x00008000 | E3 A0 00 42 E3 A0 10 00 ... | ...B... │
│ 0x00008010 | EF 00 00 01 EF 00 00 00 ... | ....... │
│ 0x00008020 | ...                                   │
│ (16 bytes per row, 16 rows = 256 bytes)          │
└─────────────────────────────────────────────────┘
```

**Features:**
- Address text field accepts hex input (0x prefix optional)
- Quick access buttons jump to PC, SP, or register values
- Shows 256 bytes (configurable)
- Format: Address | 16 hex bytes | 16 ASCII chars
- Highlight changed bytes (compare with previous state)
- Auto-scroll to PC toggle (follows execution)
- Updates via WebSocket state events

**Data Flow:**
- User enters address → API call `GET /api/v1/session/{id}/memory?addr={addr}&length=256`
- WebSocket state event → Refresh if auto-scroll enabled
- Register button clicked → Load memory at register value

### 4. Stack Visualization

**StackView.swift:**

**UI Layout:**
```
┌─────────────────────────────────────────────────┐
│ Stack (SP = 0x00050000)                          │
├─────────────────────────────────────────────────┤
│ SP-16  | 0x0004FFF0 | 0x00000000 | ....         │
│ SP-12  | 0x0004FFF4 | 0x00000000 | ....         │
│ SP-8   | 0x0004FFF8 | 0x00000042 | ...B         │
│ SP-4   | 0x0004FFFC | 0x00008010 | ....  ← LR?  │
│ SP+0 → | 0x00050000 | 0xDEADBEEF | .... ◄ SP    │
│ SP+4   | 0x00050004 | 0x00000000 | ....         │
│ SP+8   | 0x00050008 | 0x00000000 | ....         │
└─────────────────────────────────────────────────┘
```

**Features:**
- Vertical layout (stack grows down visually)
- Shows SP ± 128 bytes (32 words)
- Format: `Offset | Address | Value | ASCII | Annotation`
- Highlights current SP row
- Detects patterns:
  - Saved LR (value in code range)
  - Saved FP (value in stack range)
  - Return addresses
- Auto-updates when SP changes
- Click row to jump to that address in Memory view

**Pattern Detection:**
- LR pattern: Value in range [0x8000, 0x10000] likely code address
- FP pattern: Value in stack range, aligned to 4 bytes
- Annotations appear on right: "← LR?", "← FP?", "← R4"

### 5. Disassembly View

**DisassemblyView.swift:**

**UI Layout:**
```
┌─────────────────────────────────────────────────┐
│ Disassembly (PC = 0x00008004)                    │
├─────────────────────────────────────────────────┤
│   0x00008000 | E3A00042 | MOV    R0, #66  | main  │
│ ► 0x00008004 | E3A01000 | MOV    R1, #0   | main+4│
│ ● 0x00008008 | EF000001 | SWI    #1       | main+8│
│   0x0000800C | EF000000 | SWI    #0       | main+C│
└─────────────────────────────────────────────────┘
```

**Features:**
- Shows ±32 instructions around PC (64 total)
- Format: `[Indicator] Address | Machine Code | Mnemonic | Symbol`
- Indicators:
  - `►` = Current PC
  - `●` = Breakpoint
  - ` ` = Normal
- Highlights PC line (background color)
- Shows breakpoint dots in left margin
- Auto-scrolls to PC when stepping
- Symbol names from API (if available)

**Data Flow:**
- API call: `GET /api/v1/session/{id}/disassembly?addr={pc-128}&count=64`
- Returns array of: `{address, machineCode, mnemonic, symbol}`
- Updates on step/run via WebSocket state events
- Breakpoints synced with editor's breakpoint list

### 6. Syntax Highlighting

**SyntaxHighlightingTextView.swift:**

Custom NSTextView wrapper using NSViewRepresentable:

```swift
struct SyntaxHighlightingTextView: NSViewRepresentable {
    @Binding var text: String
    @Binding var breakpoints: Set<UInt32>

    func makeNSView(context: Context) -> NSScrollView
    func updateNSView(_ nsView: NSScrollView, context: Context)

    class Coordinator: NSObject, NSTextStorageDelegate {
        // Handle syntax highlighting on text change
        func textStorage(_ textStorage: NSTextStorage,
                        didProcessEditing editedMask: NSTextStorageEditActions,
                        range editedRange: NSRange,
                        changeInLength delta: Int)
    }
}
```

**Syntax Coloring Rules:**

| Token Type | Color | Pattern |
|------------|-------|---------|
| Instructions | Blue (#0000FF) | MOV, ADD, SUB, LDR, STR, B, BL, etc. |
| Registers | Purple (#8B00FF) | R0-R15, SP, LR, PC, CPSR |
| Labels | Brown (#8B4513) | `word:` at start of line |
| Comments | Green (#008000) | `;` to end of line |
| Numbers | Orange (#FF8C00) | #decimal, 0xhex, 0binary |
| Directives | Teal (#008B8B) | .org, .equ, .data, .text, LTORG |
| Strings | Red (#DC143C) | "..." |

**Implementation:**
- Use `NSTextStorage` subclass with custom `processEditing()`
- Apply attributes via `addAttribute(.foregroundColor, value: NSColor, range:)`
- Regex patterns for each token type
- Re-highlight on text change (debounced for performance)

### 7. Breakpoint Gutter

**BreakpointGutter.swift:**

Manages line number gutter with breakpoint interaction:

```swift
class BreakpointGutterView: NSView {
    var breakpoints: Set<Int> = [] // Line numbers
    var lineNumbersToAddresses: [Int: UInt32] = [:] // Map lines to addresses

    override func draw(_ dirtyRect: NSRect) {
        // Draw background
        // Draw line numbers
        // Draw breakpoint indicators (red circles)
    }

    override func mouseDown(with event: NSEvent) {
        let clickPoint = convert(event.locationInWindow, from: nil)
        let lineNumber = calculateLine(from: clickPoint.y)
        toggleBreakpoint(at: lineNumber)
    }

    private func toggleBreakpoint(at line: Int) {
        guard let address = lineNumbersToAddresses[line] else { return }
        // Call viewModel.toggleBreakpoint(at: address)
    }
}
```

**Visual Indicators:**
- Breakpoint enabled: Solid red circle (●)
- Breakpoint disabled: Hollow red circle (○)
- No breakpoint: Just line number
- Hover effect: Faint circle preview

**Integration:**
- Embedded in SyntaxHighlightingTextView as left accessory view
- Width: 40pt (line numbers + breakpoint indicator)
- Syncs with backend: `POST/DELETE /api/v1/session/{id}/breakpoint`
- Syncs with DisassemblyView (both show breakpoints)

### 8. Preferences Window

**AppSettings.swift:**

```swift
@MainActor
class AppSettings: ObservableObject {
    @AppStorage("backendURL") var backendURL = "http://localhost:8080"
    @AppStorage("editorFontSize") var editorFontSize = 14
    @AppStorage("colorScheme") var colorScheme = "auto" // "light", "dark", "auto"
    @AppStorage("maxRecentFiles") var maxRecentFiles = 10
}
```

**PreferencesView.swift:**

```swift
struct PreferencesView: View {
    @StateObject private var settings = AppSettings()

    var body: some View {
        TabView {
            GeneralPreferences(settings: settings)
                .tabItem { Label("General", systemImage: "gear") }

            EditorPreferences(settings: settings)
                .tabItem { Label("Editor", systemImage: "doc.text") }
        }
        .frame(width: 500, height: 300)
    }
}
```

**Settings:**

**General Tab:**
- Backend URL (text field, default: http://localhost:8080)
- Color scheme (Picker: Light/Dark/Auto)
- Max recent files (Stepper: 5-20, default: 10)

**Editor Tab:**
- Font size (Stepper: 10-24, default: 14)
- (Future: Syntax colors customization)

**Window Management:**
- Shown via Window → Preferences (⌘,)
- Uses `.Settings` window style on macOS
- Changes apply immediately (no OK/Cancel)
- Persists via @AppStorage to UserDefaults

## API Changes Required

### New API Methods in APIClient.swift

```swift
// Memory
func getMemory(sessionID: String, address: UInt32, length: Int) async throws -> [UInt8]

// Disassembly
func getDisassembly(sessionID: String, address: UInt32, count: Int) async throws -> [DisassembledInstruction]

// Examples (if endpoint exists, otherwise read from bundle)
func listExamples() async throws -> [String]
func getExample(name: String) async throws -> String

struct DisassembledInstruction: Codable {
    let address: UInt32
    let machineCode: UInt32
    let mnemonic: String
    let symbol: String?
}
```

### ViewModel Extensions

**EmulatorViewModel.swift additions:**

```swift
// Memory state
@Published var memoryData: [UInt8] = []
@Published var memoryAddress: UInt32 = 0x00008000

// Stack state (derived from memory around SP)
var stackData: [(offset: Int, address: UInt32, value: UInt32)] { ... }

// Disassembly state
@Published var disassembly: [DisassembledInstruction] = []

// Breakpoints
@Published var breakpoints: Set<UInt32> = []
func toggleBreakpoint(at address: UInt32) async
func syncBreakpointsFromAPI() async

// Methods
func loadMemory(at address: UInt32, length: Int) async
func loadDisassembly(around address: UInt32, count: Int) async
```

## Implementation Plan

### Phase A: File Management (Commit 1-2)

**Commit 1: File operations and recent files**
1. Create `FileService.swift`
2. Create `AppSettings.swift`
3. Modify `MainView.swift`:
   - Add File menu items (Open, Save, Recent)
   - Connect to FileService
   - Handle file open/save via NSOpenPanel/NSSavePanel
4. Test: Open file, save file, recent files menu

**Commit 2: Examples browser**
1. Create `ExamplesBrowserView.swift`
2. Add File → Open Example menu item
3. Load examples from `examples/` directory (via API or bundle)
4. Implement search, preview, and load functionality
5. Test: Browse examples, search, load into editor

### Phase B: Debugging Views (Commit 3-4)

**Commit 3: Memory and stack views**
1. Create `MemoryView.swift`
2. Create `StackView.swift`
3. Add `getMemory()` to APIClient
4. Modify `MainView.swift`: Add TabView for right panel
5. Modify `EmulatorViewModel.swift`: Add memory/stack state and methods
6. Test: View memory, navigate to addresses, see stack updates

**Commit 4: Disassembly view**
1. Create `DisassemblyView.swift`
2. Add `getDisassembly()` to APIClient
3. Modify `EmulatorViewModel.swift`: Add disassembly state and methods
4. Test: View disassembly, auto-scroll to PC, see breakpoints

### Phase C: Advanced Editor (Commit 5-6)

**Commit 5: Breakpoint gutter**
1. Create `BreakpointGutter.swift`
2. Integrate into `EditorView.swift`
3. Modify `EmulatorViewModel.swift`: Add breakpoint management
4. Test: Click gutter to add/remove breakpoints, sync with API

**Commit 6: Syntax highlighting**
1. Create `SyntaxHighlightingTextView.swift`
2. Replace EditorView's TextEditor with custom view
3. Implement syntax coloring for ARM assembly
4. Test: Type code, see syntax highlighting, verify colors

**Commit 7: Preferences window**
1. Create `PreferencesView.swift` (already have AppSettings)
2. Add Window → Preferences menu item
3. Implement General and Editor tabs
4. Test: Change settings, verify persistence across launches

### Phase D: Update Planning Document (Commit 8)

**Commit 8: Update SWIFT_GUI_PLANNING.md**
1. Mark Stage 4 as complete
2. Document implementation details
3. Update success criteria
4. Note any deviations from original plan

## Success Criteria

- ✅ All 9 features implemented and functional
- ✅ Tabbed right panel with 4 tabs (Registers, Memory, Stack, Disassembly)
- ✅ File open/save dialogs work with .s files
- ✅ Recent files menu populated and functional
- ✅ Examples browser shows all 49 examples with search
- ✅ Memory view shows hex dump, responds to address input
- ✅ Stack view displays stack with SP offsets and annotations
- ✅ Disassembly view shows live disassembly, auto-scrolls to PC
- ✅ Breakpoint gutter allows click-to-toggle, syncs with backend
- ✅ Syntax highlighting colors ARM assembly correctly
- ✅ Preferences window saves settings persistently
- ✅ All keyboard shortcuts work (⌘O, ⌘S, ⌘,, etc.)
- ✅ Zero SwiftLint violations
- ✅ Code formatted with SwiftFormat
- ✅ App builds and runs without errors
- ✅ Native macOS look and feel maintained

## Testing Strategy

For each feature:
1. Manual testing during development
2. Edge case testing (empty files, invalid addresses, etc.)
3. Integration testing (features work together)
4. Performance testing (syntax highlighting doesn't lag on large files)

## Future Enhancements (Post-Stage 4)

- Customizable syntax colors in preferences
- Multi-region memory view (show code + stack simultaneously)
- Stack frame detection with local variable inference
- Conditional breakpoints (when backend supports it)
- Watch expressions
- Instruction tooltips (hover for opcode details)

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Syntax highlighting performance on large files | Implement incremental highlighting, only visible range |
| NSTextView complexity | Start simple, iterate, use existing examples from Apple docs |
| Breakpoint address calculation from line numbers | Parse source during load, maintain line→address map |
| Examples directory structure varies | Flexible parser, handle missing descriptions gracefully |

---

**End of Design Document**
