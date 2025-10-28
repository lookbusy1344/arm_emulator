# GUI Debugging Views Design

**Date:** 2025-10-28
**Status:** In Design
**Architecture:** Wails event-driven with extended service API

## Overview

Add complete debugging view parity between TUI and GUI by implementing missing views (Source, Disassembly, Stack, Breakpoints, Status, Output) with a GUI-optimized layout.

## Design Decisions

### Scope
- **Full TUI parity:** Implement all missing debugging views (Source, Disassembly, Stack, Breakpoints, Status, Output)
- **Not just** Source and Disassembly - complete debugging experience

### Layout Approach
- **GUI-optimized layout:** Leverage GUI advantages (tabs, collapsible panels, drag-and-drop, resizable areas)
- **Not** matching TUI's fixed panel arrangement
- Modern, flexible UX appropriate for graphical interface

### Architecture
- **Wails event-driven approach:** Use Wails' EventEmit/EventOn for state updates
- Service API extensions for data access
- Components auto-update on VM state changes
- Leverages Wails framework capabilities

### Output Capture
- **Service API extension:** Add output buffer to DebuggerService
- `GetOutput()` API returns captured output since last call
- Thread-safe with mutex
- VM writes to service's output buffer

### Backend API Extensions
All new APIs added to `service/debugger_service.go`:
- `GetSourceMap()` - Complete source map access
- `GetDisassembly(address, count)` - Disassembly generation
- `GetStack(offset, count)` - Stack data access
- `GetSymbolForAddress(addr)` - Address-to-symbol resolution

## Backend Service API Design

### New APIs in `service/debugger_service.go`

```go
// GetSourceMap returns the complete source map
func (s *DebuggerService) GetSourceMap() map[uint32]string

// GetDisassembly returns disassembled instructions starting at address
func (s *DebuggerService) GetDisassembly(startAddr uint32, count int) []DisassemblyLine

// GetStack returns stack contents from SP+offset
func (s *DebuggerService) GetStack(offset int, count int) []StackEntry

// GetSymbolForAddress resolves an address to a symbol name
func (s *DebuggerService) GetSymbolForAddress(addr uint32) string

// GetOutput returns captured program output (clears buffer)
func (s *DebuggerService) GetOutput() string

// GetWatchpoints returns all watchpoints (type already exists)
func (s *DebuggerService) GetWatchpoints() []WatchpointInfo
```

### New Types in `service/types.go`

```go
// DisassemblyLine represents a single disassembled instruction
type DisassemblyLine struct {
    Address uint32
    Opcode  uint32
    Symbol  string // Symbol at this address, if any
}

// StackEntry represents a single stack location
type StackEntry struct {
    Address uint32
    Value   uint32
    Symbol  string // If value points to a symbol
}
```

### Output Capture Implementation

- Add `outputBuffer *bytes.Buffer` and `outputMutex sync.Mutex` to DebuggerService
- Initialize buffer in NewDebuggerService
- Set `VM.Output = &outputBuffer` during LoadProgram
- `GetOutput()` locks mutex, reads buffer, clears it, returns string
- Thread-safe for concurrent GUI access

---

## Frontend Component Architecture

### New React Components

All components located in `gui/frontend/src/components/`:

- **SourceView.tsx** - Displays source code with current line highlighting, breakpoint markers
- **DisassemblyView.tsx** - Shows disassembled machine code with PC highlighting
- **StackView.tsx** - Displays stack contents with SP marker and recent write highlighting
- **BreakpointsView.tsx** - List of breakpoints/watchpoints with enable/disable controls
- **OutputView.tsx** - Program output (stdout from SWI syscalls)
- **StatusView.tsx** - Debugger status messages and errors

### Layout Structure

The GUI uses a flexible panel layout:

- **Top:** Toolbar (Load, Step, Run, Pause, Reset buttons) - already exists
- **Left panel:** Tabbed interface for Source / Disassembly views
- **Right panel:** Stacked panels for Registers / Memory / Stack
- **Bottom panel:** Tabbed interface for Output / Breakpoints / Status
- **All panels resizable** with split panes

### Event Integration

Each component subscribes to Wails events for automatic updates:

```typescript
// Subscribe to VM state changes
runtime.EventsOn("vm:state-changed", () => {
    // Refresh component data by calling backend APIs
})

// Subscribe to output events
runtime.EventsOn("vm:output", (data) => {
    // Append to OutputView
})
```

Components call backend APIs to fetch data when events fire:
- SourceView calls `GetSourceMap()`, `GetRegisterState()` for PC
- DisassemblyView calls `GetDisassembly(pc, 15)`, `GetBreakpoints()`
- StackView calls `GetStack(0, 16)`, `GetRegisterState()` for SP
- etc.

---

## Event Emission Strategy & Wails Integration

### Event Types

Events emitted from backend to frontend:

- **vm:state-changed** - Emitted after Step(), Continue() completes, Reset(), Pause()
- **vm:output** - Emitted when output buffer receives data (real-time)
- **vm:breakpoint-hit** - Emitted when execution stops at breakpoint
- **vm:error** - Emitted when execution error occurs

### Output Streaming Implementation (Option 1: Custom Writer)

Create a custom `io.Writer` that wraps the output buffer and emits events:

```go
type eventEmittingWriter struct {
    buffer *bytes.Buffer
    ctx    context.Context
    mutex  sync.Mutex
}

func (w *eventEmittingWriter) Write(p []byte) (n int, err error) {
    w.mutex.Lock()
    defer w.mutex.Unlock()

    n, err = w.buffer.Write(p)
    if err == nil && n > 0 {
        // Emit event with the new output
        runtime.EventsEmit(w.ctx, "vm:output", string(p))
    }
    return n, err
}
```

- Clean design, immediate feedback, no polling overhead
- DebuggerService uses this writer instead of plain buffer
- Frontend receives output in real-time as it's generated

### Wails App.go Changes

Emit events after state-changing operations:

```go
func (a *App) Step() error {
    err := a.service.Step()
    runtime.EventsEmit(a.ctx, "vm:state-changed")
    if err != nil {
        runtime.EventsEmit(a.ctx, "vm:error", err.Error())
    }
    return err
}

func (a *App) Continue() error {
    err := a.service.RunUntilHalt()
    runtime.EventsEmit(a.ctx, "vm:state-changed")
    if a.service.GetExecutionState() == service.StateBreakpoint {
        runtime.EventsEmit(a.ctx, "vm:breakpoint-hit")
    }
    return err
}
```

### Breakpoint Management Wrappers

Add App.go wrappers that emit events:

```go
func (a *App) ToggleBreakpoint(address uint32) error {
    bps := a.service.GetBreakpoints()
    exists := false
    for _, bp := range bps {
        if bp.Address == address {
            exists = true
            break
        }
    }

    var err error
    if exists {
        err = a.service.RemoveBreakpoint(address)
    } else {
        err = a.service.AddBreakpoint(address)
    }

    if err == nil {
        runtime.EventsEmit(a.ctx, "vm:state-changed")
    }
    return err
}
```

---

## Component Implementation Details

### SourceView Component

**Functionality:**
- Displays source lines from `GetSourceMap()`
- Highlights current line (PC address) in yellow
- Shows breakpoint markers (*) next to line numbers
- Click on line number to toggle breakpoint via `ToggleBreakpoint(address)`
- Auto-scrolls to keep PC in view
- Shows symbol labels above their addresses in green

**Data updates:** Subscribes to `vm:state-changed`, calls `GetSourceMap()` and `GetRegisterState()` for PC

### DisassemblyView Component

**Functionality:**
- Calls `GetDisassembly(pc - 20, 15)` to show context around PC
- Displays: address, opcode (hex), symbol name
- Highlights current PC line in yellow
- Breakpoint markers same as SourceView
- Click address to toggle breakpoint

**Data updates:** Subscribes to `vm:state-changed`, refreshes disassembly around current PC

### StackView Component

**Functionality:**
- Displays 16 words from `GetStack(0, 16)`
- Shows: address, value (hex), symbol resolution if value is pointer
- Highlights SP position with arrow marker (->)
- Future enhancement: highlighting for recently modified stack locations

**Data updates:** Subscribes to `vm:state-changed`, calls `GetStack()` and `GetRegisterState()` for SP

### BreakpointsView Component

**Functionality:**
- Table/list of all breakpoints from `GetBreakpoints()`
- Columns: Address (hex), Symbol (if any), Enabled checkbox
- Click checkbox to toggle enabled state (future: need enable/disable API)
- Delete button calls `RemoveBreakpoint(address)`
- Add button to create new breakpoint with address input → `AddBreakpoint(address)`

**Data updates:** Subscribes to `vm:state-changed`, refreshes breakpoint list

### OutputView Component

**Functionality:**
- Scrollable text area showing program output
- Appends output from `vm:output` events in real-time
- Auto-scrolls to bottom on new output
- Clear button to reset output buffer
- Monospace font for formatted output

**Data updates:** Subscribes to `vm:output`, appends to display

### StatusView Component

**Functionality:**
- Shows debugger status messages
- Execution state (Running/Halted/Breakpoint/Error) from `GetExecutionState()`
- Error messages from `vm:error` events
- Breakpoint hit notifications from `vm:breakpoint-hit` events
- Step count / cycle count display from `GetRegisterState().Cycles`

**Data updates:** Subscribes to `vm:state-changed`, `vm:error`, `vm:breakpoint-hit`

---

## Layout Implementation & UI Library

### Split Pane Library Selection

**Chosen library:** **allotment** (https://github.com/johnwalley/allotment)

**Rationale:**
- Modern, well-maintained
- Excellent TypeScript support
- Based on VS Code's layout system
- Provides VS Code-like debugger experience
- Resizable panels with min/max size constraints

**Installation:**
```bash
npm install allotment
```

### Layout Structure

```tsx
<Allotment vertical>
  {/* Top toolbar - fixed height */}
  <Allotment.Pane minSize={60} maxSize={60}>
    <Toolbar />
  </Allotment.Pane>

  {/* Main content area */}
  <Allotment.Pane>
    <Allotment>
      {/* Left: Source/Disassembly tabs */}
      <Allotment.Pane minSize={300}>
        <Tabs>
          <Tab label="Source"><SourceView /></Tab>
          <Tab label="Disassembly"><DisassemblyView /></Tab>
        </Tabs>
      </Allotment.Pane>

      {/* Right: Registers/Memory/Stack */}
      <Allotment.Pane minSize={300}>
        <Allotment vertical>
          <RegisterView />
          <MemoryView />
          <StackView />
        </Allotment>
      </Allotment.Pane>
    </Allotment>
  </Allotment.Pane>

  {/* Bottom: Output/Breakpoints/Status tabs */}
  <Allotment.Pane minSize={150} preferredSize={200}>
    <Tabs>
      <Tab label="Output"><OutputView /></Tab>
      <Tab label="Breakpoints"><BreakpointsView /></Tab>
      <Tab label="Status"><StatusView /></Tab>
    </Tabs>
  </Allotment.Pane>
</Allotment>
```

**Key features:**
- All panels user-resizable
- Min/max size constraints prevent UI breaking
- Persistent layout state (future: save/restore panel sizes)
- Vertical and horizontal split support

---

## Testing Strategy

### Backend Service API Tests

**Location:** `tests/unit/service/debugger_service_test.go`

**Coverage:**
- Unit tests for each new API method
- `GetSourceMap()` - verify complete source map returned
- `GetDisassembly(address, count)` - verify correct instructions decoded
- `GetStack(offset, count)` - verify stack data extraction
- `GetSymbolForAddress(addr)` - verify symbol resolution
- `GetOutput()` - verify output capture and clearing
- Event-emitting writer - verify events emitted on Write()
- Mock VM state for predictable test scenarios

### Frontend Component Tests

**Location:** `gui/frontend/src/components/*.test.tsx`

**Coverage:**
- Component tests using React Testing Library
- Test event subscription/unsubscription (no memory leaks)
- Test data fetching on state changes
- Test user interactions (breakpoint toggle, address navigation)
- Test rendering with mock data
- Test error states (no data, API errors)

**Integration tests:** `gui/frontend/src/test/integration.test.tsx`
- Full flow: Load program → Step → Verify all views update
- Breakpoint workflow: Set → Hit → Verify UI state
- Output streaming: Program writes → Verify OutputView updates

---

## Implementation Phases

### Phase 1: Backend Service Extensions
- Add new APIs to `service/debugger_service.go`
- Add new types to `service/types.go`
- Write unit tests for all new methods
- Verify tests pass

### Phase 2: Event-Emitting Writer
- Implement custom `io.Writer` for output streaming
- Integrate with DebuggerService
- Pass context to writer for event emission
- Test event emission

### Phase 3: App.go Event Emission
- Add `runtime.EventsEmit()` calls to Step(), Continue(), Reset(), Pause()
- Add `ToggleBreakpoint()` wrapper method
- Emit appropriate events (vm:state-changed, vm:error, vm:breakpoint-hit)

### Phase 4: Frontend Components (Core)
- Implement SourceView component
- Implement DisassemblyView component
- Add event subscriptions
- Test with mock data

### Phase 5: Frontend Components (Debugging)
- Implement StackView component
- Implement BreakpointsView component
- Implement OutputView component
- Implement StatusView component
- Add event subscriptions for all components

### Phase 6: Layout Integration
- Install allotment library: `npm install allotment`
- Create tab components for view switching
- Integrate all components into App.tsx with allotment layout
- Remove old simple layout

### Phase 7: Styling & Polish
- Add CSS for consistent styling across components
- Syntax highlighting for source/disassembly (future enhancement)
- Color scheme matching debugger theme (yellow for current line, green for symbols)
- Monospace fonts for code views

### Phase 8: Testing & Refinement
- Full integration testing with real programs
- Test with example programs (hello.s, factorial.s, etc.)
- Bug fixes and performance optimization
- Documentation updates

---

---

## Additional TUI Features (Full Parity)

### Command Input Interface

**Component:** `CommandInput.tsx`
- Text input field at bottom of GUI (similar to TUI layout)
- Accepts all debugger commands that TUI supports
- Command history (up/down arrows)
- Auto-completion for commands and symbols
- Output displayed in StatusView

**Backend API:**
```go
func (a *App) ExecuteCommand(command string) (string, error)
```
- Calls `debugger.ExecuteCommand(command)`
- Returns command output as string
- Emits `vm:state-changed` for state-modifying commands

### Advanced Stepping Modes

**New buttons in toolbar:**
- **Step Over** (Next) - Execute function call without stepping into it
- **Step Out** (Finish) - Complete current function and return to caller

**Backend APIs:**
```go
func (a *App) StepOver() error  // Calls debugger.cmdNext()
func (a *App) StepOut() error   // Calls debugger.cmdFinish()
```

### Watchpoints Support

**Extend BreakpointsView to show watchpoints:**
- Table showing both breakpoints and watchpoints
- Watchpoint columns: Address, Type (read/write/readwrite), Enabled
- Add/remove/toggle watchpoints via UI

**Backend APIs (already exist in service):**
```go
func (a *App) AddWatchpoint(address uint32, watchType string) error
func (a *App) RemoveWatchpoint(address uint32) error
func (s *DebuggerService) GetWatchpoints() []WatchpointInfo  // Already defined
```

### Conditional Breakpoints

**Enhanced breakpoint management:**
- Breakpoint dialog/modal for setting conditions
- Condition expression input field
- Display condition in BreakpointsView table

**Extended BreakpointInfo type:**
```go
type BreakpointInfo struct {
    Address   uint32
    Enabled   bool
    Condition string  // NEW: expression that must evaluate to true
}
```

**Backend API update:**
```go
func (a *App) AddBreakpointWithCondition(address uint32, condition string) error
```

### Expression Evaluation

**New component:** `ExpressionEvaluator.tsx`
- Input field for entering expressions
- Evaluate button
- Result display area
- Support for: registers, memory addresses, arithmetic, symbols

**Backend API:**
```go
func (a *App) EvaluateExpression(expr string) (uint32, error)
```
- Uses `debugger.Evaluator.Evaluate(expr, vm, symbols)`

### Additional Inspection Commands

All accessible via CommandInput:

**Memory examination:**
- `x` command - Examine memory at address/symbol
- Display in MemoryView or separate modal

**Info commands:**
- `info registers` - Show all registers (updates RegisterView)
- `info breakpoints` - Show breakpoints (updates BreakpointsView)
- `info watchpoints` - Show watchpoints
- `info symbols` - Show symbol table

**List command:**
- `list` - Show source around PC (updates SourceView)
- `list <address>` - Show source at address

**Backtrace:**
- `backtrace` / `bt` - Show call stack
- New component or display in StatusView

**Print command:**
- `print <expr>` - Evaluate and print expression
- Display result in StatusView

### State Modification

**Set command:**
- `set reg R0 = 0x1234` - Modify register value
- `set mem 0x8000 = 0xFF` - Modify memory value
- Backend uses existing debugger.cmdSet()

---

## Updated Component List

### All Components (Complete List)

1. **SourceView.tsx** - Source code with highlighting ✓ (already designed)
2. **DisassemblyView.tsx** - Disassembled instructions ✓ (already designed)
3. **RegisterView.tsx** - Registers and CPSR ✓ (already exists)
4. **MemoryView.tsx** - Memory contents ✓ (already exists)
5. **StackView.tsx** - Stack contents ✓ (already designed)
6. **BreakpointsView.tsx** - Breakpoints AND watchpoints with conditions
7. **OutputView.tsx** - Program output ✓ (already designed)
8. **StatusView.tsx** - Debugger messages ✓ (already designed)
9. **CommandInput.tsx** - Command line interface (NEW)
10. **ExpressionEvaluator.tsx** - Expression evaluation panel (NEW)

---

## Updated Backend Service APIs

### Complete API List

**Existing APIs:**
- `GetRegisterState()` ✓
- `GetMemory(address, size)` ✓
- `GetBreakpoints()` - Extended to include conditions
- `GetWatchpoints()` ✓
- `Step()` ✓
- `Continue()` ✓
- `Pause()` ✓
- `Reset()` ✓
- `AddBreakpoint(address)` - Extended for conditions
- `RemoveBreakpoint(address)` ✓

**New APIs:**
- `GetSourceMap()` ✓
- `GetDisassembly(address, count)` ✓
- `GetStack(offset, count)` ✓
- `GetSymbolForAddress(addr)` ✓
- `GetOutput()` ✓
- `ExecuteCommand(command string) (string, error)` (NEW)
- `StepOver()` (NEW)
- `StepOut()` (NEW)
- `AddBreakpointWithCondition(address, condition)` (NEW)
- `AddWatchpoint(address, type)` (NEW)
- `RemoveWatchpoint(address)` (NEW)
- `EvaluateExpression(expr)` (NEW)
- `GetSymbols()` ✓ (already exists)

---

## Updated Layout Structure

```tsx
<Allotment vertical>
  {/* Top toolbar - Step, Step Over, Step Out, Run, Pause, Reset */}
  <Allotment.Pane minSize={60} maxSize={60}>
    <Toolbar />
  </Allotment.Pane>

  {/* Main content area */}
  <Allotment.Pane>
    <Allotment>
      {/* Left: Source/Disassembly tabs */}
      <Allotment.Pane minSize={300}>
        <Tabs>
          <Tab label="Source"><SourceView /></Tab>
          <Tab label="Disassembly"><DisassemblyView /></Tab>
        </Tabs>
      </Allotment.Pane>

      {/* Right: Registers/Memory/Stack/Expression Evaluator */}
      <Allotment.Pane minSize={300}>
        <Allotment vertical>
          <RegisterView />
          <MemoryView />
          <StackView />
          <ExpressionEvaluator />
        </Allotment>
      </Allotment.Pane>
    </Allotment>
  </Allotment.Pane>

  {/* Bottom: Output/Breakpoints+Watchpoints/Status tabs */}
  <Allotment.Pane minSize={150} preferredSize={200}>
    <Tabs>
      <Tab label="Output"><OutputView /></Tab>
      <Tab label="Breakpoints"><BreakpointsView /></Tab>
      <Tab label="Status"><StatusView /></Tab>
    </Tabs>
  </Allotment.Pane>

  {/* Command Input - fixed height at very bottom */}
  <Allotment.Pane minSize={40} maxSize={40}>
    <CommandInput />
  </Allotment.Pane>
</Allotment>
```

---

## Updated Implementation Phases

### Phase 1: Backend Service Extensions (Enhanced)
- Add all new APIs to `service/debugger_service.go`
- `ExecuteCommand()` wrapper
- `StepOver()`, `StepOut()` wrappers
- `AddBreakpointWithCondition()`, `AddWatchpoint()`, `RemoveWatchpoint()`
- `EvaluateExpression()` wrapper
- Extend `BreakpointInfo` type to include `Condition` field
- Write unit tests for all new methods

### Phase 2: Event-Emitting Writer (Same)
- Implement custom `io.Writer` for output streaming
- Integrate with DebuggerService
- Test event emission

### Phase 3: App.go Event Emission & Command Interface (Enhanced)
- Add all event emission to state-changing methods
- Add `ExecuteCommand()` wrapper that calls debugger.ExecuteCommand()
- Add `StepOver()`, `StepOut()` wrappers
- Add watchpoint management wrappers
- Add `EvaluateExpression()` wrapper

### Phase 4: Frontend Components (Core)
- SourceView
- DisassemblyView
- Event subscriptions

### Phase 5: Frontend Components (Debugging - Enhanced)
- StackView
- BreakpointsView with watchpoints and conditions support
- OutputView
- StatusView
- CommandInput with history and auto-complete (NEW)
- ExpressionEvaluator (NEW)

### Phase 6: Layout Integration (Same)
- Install allotment
- Create tabs
- Integrate all components including CommandInput

### Phase 7: Advanced Features
- Command history implementation
- Command auto-completion
- Conditional breakpoint dialog/modal
- Watchpoint type selection UI
- Backtrace/call stack display

### Phase 8: Styling & Polish (Same)
- CSS styling
- Color scheme
- Monospace fonts

### Phase 9: Testing & Refinement (Enhanced)
- Test all debugger commands via CommandInput
- Test watchpoints trigger correctly
- Test conditional breakpoints
- Test expression evaluation
- Test step over/out functionality
- Integration tests with example programs

---

## Updated Success Criteria

**All original criteria PLUS:**
- ✓ Command input interface functional with all TUI commands
- ✓ Command history and auto-completion working
- ✓ Step Over and Step Out buttons functional
- ✓ Watchpoints can be added, removed, and trigger correctly
- ✓ Conditional breakpoints can be set and evaluate correctly
- ✓ Expression evaluator parses and evaluates expressions
- ✓ All inspection commands (info, list, x, print, backtrace) work
- ✓ State modification commands (set) work
- ✓ Full TUI command parity achieved

---

## Dependencies & Risks

### Dependencies
- **allotment** npm package for resizable panels
- Wails runtime for event system (`@wailsapp/runtime`)
- React Testing Library (already in project)
- Existing service layer and debugger infrastructure
- Existing debugger.Evaluator for expression parsing
- Existing debugger command handlers (cmdStep, cmdNext, cmdFinish, etc.)

### Potential Risks

1. **Event flooding** - If Step() is called rapidly, could flood frontend with events
   - Mitigation: Debounce or throttle event emission
   - Or: Batch state updates in frontend

2. **Memory leaks** - Event listeners not cleaned up
   - Mitigation: Proper useEffect cleanup in React components
   - Test: Verify EventsOff called on component unmount

3. **Performance** - GetDisassembly() on every step might be slow
   - Mitigation: Cache disassembly results, only refresh on PC change
   - Or: Lazy load disassembly only when DisassemblyView tab is active

4. **Context passing** - Event-emitting writer needs context reference
   - Solution: Store context in App struct, pass to service during initialization

5. **Command execution complexity** - ExecuteCommand() needs to handle all debugger commands
   - Mitigation: Leverage existing debugger.ExecuteCommand() implementation from TUI
   - Already tested and working in TUI

6. **Watchpoint implementation** - Watchpoints require memory access hooks
   - Note: Already implemented in debugger package
   - Need to expose via service layer

7. **Expression parsing** - Expression evaluator complexity
   - Mitigation: Use existing debugger.Evaluator from TUI
   - Already handles registers, symbols, memory, arithmetic

8. **Command auto-completion** - Building completion suggestions
   - Mitigation: Start with simple command list, add symbol completion later
   - Can be enhanced incrementally

9. **Increased scope** - Full TUI parity adds significant complexity
   - Mitigation: Phased implementation approach
   - Core features first, advanced features later
   - Thorough testing at each phase

---

## Design Complete

This design now achieves full feature parity with the TUI, providing a comprehensive debugging experience in the GUI.
