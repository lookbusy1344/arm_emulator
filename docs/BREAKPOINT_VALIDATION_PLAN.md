# Breakpoint Validation Plan for Swift GUI

## Problem Statement

The Swift GUI allows setting breakpoints on invalid lines (blank lines, comments, directives that don't generate code). Commits `4c81f13..116dfcc` attempted to fix this but the solution doesn't work.

## Root Cause Analysis

### Current Implementation (Broken)

The current approach in `EditorView.swift` line 43-60:

```swift
private func toggleBreakpoint(at lineNumber: Int) {
    // Line number to address: assumes linear mapping starting at 0x8000
    let address = UInt32(0x8000 + (lineNumber - 1) * 4)
    
    // Check if address exists in source map
    guard viewModel.sourceMap[address] != nil || breakpoints.contains(lineNumber) else {
        print("Cannot set breakpoint...")
        return
    }
    // ...
}
```

### Why It Fails

1. **The line-to-address formula is wrong**: `address = 0x8000 + (lineNumber - 1) * 4`
   
   This assumes every source line generates a 4-byte instruction starting at line 1. In reality:
   
   Example `hello.s`:
   ```
   Line 1:  ; hello.s - Classic "Hello World"    <- comment, no code
   Line 2:  ; Demonstrates...                     <- comment, no code  
   Line 3:  (blank)                               <- blank, no code
   Line 4:          .org    0x8000                <- directive, sets origin
   Line 5:  (blank)                               <- blank, no code
   Line 6:  _start:                               <- label only, no code
   Line 7:          LDR     R0, =msg_hello        <- FIRST instruction at 0x8000
   Line 8:          SWI     #0x02                 <- 0x8004
   ...
   ```

   With the broken formula:
   - Line 7 → address `0x8000 + (7-1)*4 = 0x8018` (WRONG - actual is 0x8000)
   - Line 1 → address `0x8000 + (1-1)*4 = 0x8000` (WRONG - line 1 is a comment)

2. **Source map is keyed by ADDRESS, not LINE NUMBER**: The source map from the backend maps `address -> source_line_text`, not `line_number -> address`. This inverts the needed lookup.

3. **No line number tracking in backend**: The backend builds the source map from parsed instructions which have addresses and raw line text, but **not the original source file line numbers**.

## Solution Design

### Option A: Backend Provides Line-to-Address Mapping (Recommended)

Modify the backend to track and expose source line numbers during parsing.

#### Backend Changes

1. **Parser Enhancement** (`parser/parser.go`):
   - Track the current source line number during parsing
   - Store `LineNumber` field in `Instruction` and relevant `Directive` structs

2. **Source Map Enhancement** (`service/debugger_service.go`):
   - Change source map to include line numbers: `map[uint32]SourceMapEntry` where entry has `{LineNumber, RawLine}`
   - Or add a separate `lineToAddress` map: `map[int]uint32`

3. **API Enhancement** (`api/handlers.go`):
   - Extend `/sourcemap` endpoint to return line numbers
   - Response: `{sourceMap: [{address, lineNumber, line}]}`

#### Frontend Changes (Swift GUI)

1. **APIClient** (`APIClient.swift`):
   - Update `SourceMapEntry` to include `lineNumber: Int`

2. **EmulatorViewModel** (`EmulatorViewModel.swift`):
   - Add `validBreakpointLines: Set<Int>` derived from source map
   - Populate after loading source map

3. **EditorView** (`EditorView.swift`):
   - Replace broken address calculation with lookup in `validBreakpointLines`
   - Check `viewModel.validBreakpointLines.contains(lineNumber)` instead of computing address

4. **CustomGutterView** (`CustomGutterView.swift`):
   - Optionally show visual indicator (e.g., dimmed gutter) for lines that can't have breakpoints

### Option B: Frontend Parses Source (Not Recommended)

Have the Swift frontend parse the assembly source to determine valid lines.

**Pros**: No backend changes needed
**Cons**: 
- Duplicates parsing logic
- Must stay in sync with backend parser
- Complex to implement correctly

### Option C: Send Line Number in Breakpoint Request, Let Backend Validate

Frontend sends line number, backend does the line→address mapping and validates.

**Pros**: Single source of truth for validation
**Cons**: 
- Requires backend to track line numbers (same as Option A)
- Adds round-trip latency for validation feedback

## Recommended Implementation: Option A

### Phase 1: Backend Changes

1. **Update Parser** to track line numbers:
   ```go
   type Instruction struct {
       Address    uint32
       Opcode     uint32
       RawLine    string
       LineNumber int  // NEW: source file line number
   }
   ```

2. **Update Source Map API** to return line numbers:
   ```go
   type SourceMapEntry struct {
       Address    uint32 `json:"address"`
       Line       string `json:"line"`
       LineNumber int    `json:"lineNumber"`  // NEW
   }
   ```

3. **Add Validation Endpoint** (optional but useful):
   ```
   GET /api/v1/session/{id}/valid-breakpoint-lines
   Response: {lines: [7, 8, 9, 10, 14, 15]}
   ```

### Phase 2: Frontend Changes

1. **Update SourceMapEntry model**:
   ```swift
   struct SourceMapEntry: Codable {
       let address: UInt32
       let line: String
       let lineNumber: Int
   }
   ```

2. **Add valid lines tracking**:
   ```swift
   // In EmulatorViewModel
   @Published var validBreakpointLines: Set<Int> = []
   @Published var lineToAddress: [Int: UInt32] = [:]
   
   // After fetching source map:
   validBreakpointLines = Set(sourceMapEntries.map { $0.lineNumber })
   lineToAddress = Dictionary(uniqueKeysWithValues: 
       sourceMapEntries.map { ($0.lineNumber, $0.address) })
   ```

3. **Fix toggleBreakpoint**:
   ```swift
   private func toggleBreakpoint(at lineNumber: Int) {
       // Check if line can have breakpoint
       guard viewModel.validBreakpointLines.contains(lineNumber) ||
             breakpoints.contains(lineNumber) else {
           print("Cannot set breakpoint on line \(lineNumber) - not executable code")
           return
       }
       
       // Get actual address for this line
       guard let address = viewModel.lineToAddress[lineNumber] else {
           return
       }
       
       Task {
           await viewModel.toggleBreakpoint(at: address)
       }
   }
   ```

### Phase 3: Visual Feedback (Optional Enhancement)

1. **Dim non-breakable lines** in gutter
2. **Show tooltip** explaining why breakpoint can't be set
3. **Highlight breakable lines** on hover

## Testing Plan

1. **Unit Tests**:
   - Parser correctly tracks line numbers through comments, blanks, labels
   - Source map API returns correct line numbers
   - Frontend correctly identifies valid/invalid lines

2. **Integration Tests**:
   - Load program with various line types
   - Verify only executable lines accept breakpoints
   - Verify breakpoints still work on valid lines

3. **Manual Testing**:
   - Try setting breakpoints on comments → should fail
   - Try setting breakpoints on blank lines → should fail  
   - Try setting breakpoints on labels only → should fail
   - Try setting breakpoints on `.org`, `.equ` → should fail
   - Try setting breakpoints on instructions → should succeed
   - Try setting breakpoints on `.word`, `.byte` (data) → decision needed

## Edge Cases to Consider

1. **Multiple instructions per line**: Not supported in this assembler
2. **Labels on same line as instruction**: `label: MOV R0, #1` - line is valid
3. **Data directives**: `.word`, `.byte` generate addresses but aren't executable - probably shouldn't allow breakpoints
4. **LTORG-generated literals**: Have addresses but are data, not code

## Migration Notes

- Source map API change is backward-compatible (adding new field)
- Frontend should handle missing `lineNumber` field gracefully during transition
- Backend validation in `AddBreakpoint` still works as additional safety net

## Files to Modify

### Backend
- `parser/parser.go` - add line number tracking
- `parser/types.go` - add LineNumber to Instruction
- `service/debugger_service.go` - update source map building
- `api/handlers.go` - update sourcemap response

### Frontend (Swift GUI)
- `ARMEmulator/Services/APIClient.swift` - update SourceMapEntry
- `ARMEmulator/ViewModels/EmulatorViewModel.swift` - add validBreakpointLines, lineToAddress
- `ARMEmulator/Views/EditorView.swift` - fix toggleBreakpoint logic
- `ARMEmulator/Views/CustomGutterView.swift` - optional visual feedback

### Tests
- `tests/unit/api/api_test.go` - update source map test
- `swift-gui/ARMEmulatorTests/` - add breakpoint validation tests
