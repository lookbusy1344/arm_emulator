# GUI Debugging Views Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement complete debugging view parity between TUI and GUI with all missing views (Source, Disassembly, Memory, Stack, Breakpoints, Output, Status) plus advanced features (command input, expression evaluation, watchpoints, conditional breakpoints). Memory view includes full-featured scrollable hex display with yellow highlighting for recently written bytes, matching TUI functionality.

**Architecture:** Wails event-driven approach with extended service APIs. Backend exposes new APIs for source maps, disassembly, memory data with write tracking, stack access, and symbol resolution. Frontend React components subscribe to Wails events and auto-update on VM state changes. Custom event-emitting writer provides real-time output streaming. Memory write detection uses VM MemoryTrace to track changes.

**Tech Stack:** Go (backend services), Wails runtime (event system), React + TypeScript (frontend), allotment (resizable panels), tcell.SimulationScreen (TUI testing)

**Implementation Status:** COMPLETE (All 23 tasks complete)

**Last Updated:** 2025-10-30

---

## Implementation Progress

### ✅ Backend Service APIs (Tasks 1-9) - COMPLETE
- [x] Task 1: Core backend types (DisassemblyLine, StackEntry, BreakpointInfo.Condition)
- [x] Task 2: GetSourceMap API with defensive copy and tests
- [x] Task 3: GetSymbolForAddress API for symbol resolution
- [x] Task 4: GetDisassembly API with input validation and edge case tests
- [x] Task 5: GetStack API with integer overflow protection and security fixes
- [x] Task 6: EventEmittingWriter for real-time output streaming
- [x] Task 7: Integration of EventEmittingWriter with DebuggerService
- [x] Task 8: StepOver, StepOut, watchpoint management APIs (with thread-safety fixes)
- [x] Task 9: ExecuteCommand and EvaluateExpression APIs (with comprehensive test coverage)

### ✅ GUI Integration (Tasks 10-11) - COMPLETE
- [x] Task 10: Event emission in App.go (vm:state-changed, vm:error, vm:breakpoint-hit)
- [x] Task 11: Wrapper methods for all service APIs with event emission (with critical fixes)

### ✅ Frontend Components (Tasks 12-20) - COMPLETE
- [x] Task 12: Install allotment library (v1.20.4)
- [x] Task 13: SourceView component with breakpoint toggling
- [x] Task 14: DisassemblyView component with PC highlighting
- [x] Task 15: StackView component with SP marker
- [x] Task 16: OutputView component with real-time streaming
- [x] Task 17: StatusView component with execution state
- [x] Task 18: BreakpointsView component with watchpoints (with bug fixes)
- [x] Task 19: Allotment layout integration with resizable panels
- [x] Task 20: Toolbar button functionality (Step, StepOver, StepOut, Run, Pause, Reset)

### ✅ Advanced UI Components (Tasks 21-22) - COMPLETE
- [x] Task 21: CommandInput component with history (command execution, arrow key navigation, result display)
- [x] Task 22: ExpressionEvaluator component with result display (dual format hex/decimal, last 10 results, error handling)

### ✅ Final Tasks (Task 23) - COMPLETE
- [x] Task 23: Final testing and documentation (Go build success, 1024+ tests passing, 0 lint issues)

### Code Quality Summary
- **Total Tests:** 1,024+ tests, 100% pass rate
- **Lint Issues:** 0 (golangci-lint clean)
- **Build Status:** Success (both Go backend and Wails frontend)
- **Critical Bugs Fixed:** 5 (integer overflow, thread-safety, API mismatches, parameter types)
- **Security Improvements:** Integer wraparound protection, input validation

---

## Task 1: Add Core Backend Types

**Files:**
- Modify: `service/types.go`
- Test: `tests/unit/service/types_test.go`

**Step 1: Write test for DisassemblyLine type**

Create `tests/unit/service/types_test.go`:

```go
package service_test

import (
	"testing"
	"github.com/yourusername/arm-emulator/service"
)

func TestDisassemblyLine_Creation(t *testing.T) {
	line := service.DisassemblyLine{
		Address: 0x00008000,
		Opcode:  0xE3A00001,
		Symbol:  "main",
	}

	if line.Address != 0x00008000 {
		t.Errorf("Expected address 0x00008000, got 0x%08X", line.Address)
	}
	if line.Opcode != 0xE3A00001 {
		t.Errorf("Expected opcode 0xE3A00001, got 0x%08X", line.Opcode)
	}
	if line.Symbol != "main" {
		t.Errorf("Expected symbol 'main', got '%s'", line.Symbol)
	}
}

func TestStackEntry_Creation(t *testing.T) {
	entry := service.StackEntry{
		Address: 0x00050000,
		Value:   0xDEADBEEF,
		Symbol:  "data_label",
	}

	if entry.Address != 0x00050000 {
		t.Errorf("Expected address 0x00050000, got 0x%08X", entry.Address)
	}
	if entry.Value != 0xDEADBEEF {
		t.Errorf("Expected value 0xDEADBEEF, got 0x%08X", entry.Value)
	}
	if entry.Symbol != "data_label" {
		t.Errorf("Expected symbol 'data_label', got '%s'", entry.Symbol)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDisassemblyLine`

Expected: FAIL with "undefined: service.DisassemblyLine"

**Step 3: Add new types to service/types.go**

Add to `service/types.go`:

```go
// DisassemblyLine represents a single disassembled instruction
type DisassemblyLine struct {
	Address uint32 `json:"address"`
	Opcode  uint32 `json:"opcode"`
	Symbol  string `json:"symbol"` // Symbol at this address, if any
}

// StackEntry represents a single stack location
type StackEntry struct {
	Address uint32 `json:"address"`
	Value   uint32 `json:"value"`
	Symbol  string `json:"symbol"` // If value points to a symbol
}
```

**Step 4: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDisassemblyLine`

Expected: PASS (2 tests)

**Step 5: Extend BreakpointInfo type for conditions**

Add test to `tests/unit/service/types_test.go`:

```go
func TestBreakpointInfo_WithCondition(t *testing.T) {
	bp := service.BreakpointInfo{
		Address:   0x00008010,
		Enabled:   true,
		Condition: "R0 > 10",
	}

	if bp.Condition != "R0 > 10" {
		t.Errorf("Expected condition 'R0 > 10', got '%s'", bp.Condition)
	}
}
```

Run test: `go clean -testcache && go test ./tests/unit/service -v -run TestBreakpointInfo_WithCondition`

Expected: FAIL with "unknown field 'Condition'"

**Step 6: Add Condition field to BreakpointInfo**

Modify `service/types.go` - find `BreakpointInfo` struct and add:

```go
type BreakpointInfo struct {
	Address   uint32 `json:"address"`
	Enabled   bool   `json:"enabled"`
	Condition string `json:"condition"` // Expression that must evaluate to true
}
```

Run test: `go clean -testcache && go test ./tests/unit/service -v -run TestBreakpointInfo_WithCondition`

Expected: PASS

**Step 7: Commit types**

```bash
git add service/types.go tests/unit/service/types_test.go
git commit -m "feat(service): add DisassemblyLine, StackEntry types and BreakpointInfo.Condition"
```

---

## Task 2: Implement GetSourceMap API

**Files:**
- Modify: `service/debugger_service.go`
- Test: `tests/unit/service/debugger_service_test.go`

**Step 1: Write failing test for GetSourceMap**

Add to `tests/unit/service/debugger_service_test.go`:

```go
func TestDebuggerService_GetSourceMap(t *testing.T) {
	// Create service with mock VM
	s := service.NewDebuggerService()

	// Load a simple program
	program := `
main:
    MOV R0, #42
    SWI #0x00
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Get source map
	sourceMap := s.GetSourceMap()

	// Should have entries for the instructions
	if len(sourceMap) == 0 {
		t.Error("Expected non-empty source map")
	}

	// Check that main label exists at 0x8000
	if source, ok := sourceMap[0x00008000]; ok {
		if source != "    MOV R0, #42" {
			t.Errorf("Expected '    MOV R0, #42', got '%s'", source)
		}
	} else {
		t.Error("Expected source line at address 0x00008000")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetSourceMap`

Expected: FAIL with "s.GetSourceMap undefined"

**Step 3: Implement GetSourceMap method**

Add to `service/debugger_service.go`:

```go
// GetSourceMap returns the complete source map (address -> source line)
func (s *DebuggerService) GetSourceMap() map[uint32]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil || s.debugger.VM == nil {
		return make(map[uint32]string)
	}

	// Return copy of source map to prevent external modification
	sourceMap := make(map[uint32]string)
	for addr, line := range s.debugger.VM.SourceMap {
		sourceMap[addr] = line
	}

	return sourceMap
}
```

**Step 4: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetSourceMap`

Expected: PASS

**Step 5: Commit GetSourceMap**

```bash
git add service/debugger_service.go tests/unit/service/debugger_service_test.go
git commit -m "feat(service): add GetSourceMap API for source code access"
```

---

## Task 3: Implement GetSymbolForAddress API

**Files:**
- Modify: `service/debugger_service.go`
- Test: `tests/unit/service/debugger_service_test.go`

**Step 1: Write failing test for GetSymbolForAddress**

Add to `tests/unit/service/debugger_service_test.go`:

```go
func TestDebuggerService_GetSymbolForAddress(t *testing.T) {
	s := service.NewDebuggerService()

	program := `
main:
    MOV R0, #1
loop:
    ADD R0, R0, #1
    B loop
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Get symbol for main (should be at 0x8000)
	symbol := s.GetSymbolForAddress(0x00008000)
	if symbol != "main" {
		t.Errorf("Expected symbol 'main', got '%s'", symbol)
	}

	// Get symbol for loop (should be at 0x8004)
	symbol = s.GetSymbolForAddress(0x00008004)
	if symbol != "loop" {
		t.Errorf("Expected symbol 'loop', got '%s'", symbol)
	}

	// Get symbol for address without label
	symbol = s.GetSymbolForAddress(0x00008008)
	if symbol != "" {
		t.Errorf("Expected empty string, got '%s'", symbol)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetSymbolForAddress`

Expected: FAIL with "s.GetSymbolForAddress undefined"

**Step 3: Implement GetSymbolForAddress method**

Add to `service/debugger_service.go`:

```go
// GetSymbolForAddress resolves an address to a symbol name
func (s *DebuggerService) GetSymbolForAddress(addr uint32) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil {
		return ""
	}

	// Check if there's a symbol at this address
	for name, symbolAddr := range s.debugger.Symbols {
		if symbolAddr == addr {
			return name
		}
	}

	return ""
}
```

**Step 4: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetSymbolForAddress`

Expected: PASS

**Step 5: Commit GetSymbolForAddress**

```bash
git add service/debugger_service.go tests/unit/service/debugger_service_test.go
git commit -m "feat(service): add GetSymbolForAddress API for address-to-symbol resolution"
```

---

## Task 4: Implement GetDisassembly API

**Files:**
- Modify: `service/debugger_service.go`
- Test: `tests/unit/service/debugger_service_test.go`

**Step 1: Write failing test for GetDisassembly**

Add to `tests/unit/service/debugger_service_test.go`:

```go
func TestDebuggerService_GetDisassembly(t *testing.T) {
	s := service.NewDebuggerService()

	program := `
main:
    MOV R0, #42
    MOV R1, #10
    ADD R2, R0, R1
    SWI #0x00
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Get disassembly starting at main
	lines := s.GetDisassembly(0x00008000, 3)

	if len(lines) != 3 {
		t.Errorf("Expected 3 disassembly lines, got %d", len(lines))
	}

	// Check first line is at main
	if lines[0].Address != 0x00008000 {
		t.Errorf("Expected address 0x00008000, got 0x%08X", lines[0].Address)
	}
	if lines[0].Symbol != "main" {
		t.Errorf("Expected symbol 'main', got '%s'", lines[0].Symbol)
	}

	// Check opcodes are valid (non-zero)
	for i, line := range lines {
		if line.Opcode == 0 {
			t.Errorf("Line %d has zero opcode", i)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetDisassembly`

Expected: FAIL with "s.GetDisassembly undefined"

**Step 3: Implement GetDisassembly method**

Add to `service/debugger_service.go`:

```go
// GetDisassembly returns disassembled instructions starting at address
func (s *DebuggerService) GetDisassembly(startAddr uint32, count int) []DisassemblyLine {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil || s.debugger.VM == nil {
		return []DisassemblyLine{}
	}

	lines := make([]DisassemblyLine, 0, count)
	addr := startAddr

	for i := 0; i < count; i++ {
		// Read instruction from memory
		opcode, err := s.debugger.VM.Memory.ReadWord(addr)
		if err != nil {
			break
		}

		// Get symbol at this address if any
		symbol := s.GetSymbolForAddress(addr)

		line := DisassemblyLine{
			Address: addr,
			Opcode:  opcode,
			Symbol:  symbol,
		}

		lines = append(lines, line)
		addr += 4 // ARM instructions are 4 bytes
	}

	return lines
}
```

**Step 4: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetDisassembly`

Expected: PASS

**Step 5: Commit GetDisassembly**

```bash
git add service/debugger_service.go tests/unit/service/debugger_service_test.go
git commit -m "feat(service): add GetDisassembly API for disassembly generation"
```

---

## Task 5: Implement GetStack API

**Files:**
- Modify: `service/debugger_service.go`
- Test: `tests/unit/service/debugger_service_test.go`

**Step 1: Write failing test for GetStack**

Add to `tests/unit/service/debugger_service_test.go`:

```go
func TestDebuggerService_GetStack(t *testing.T) {
	s := service.NewDebuggerService()

	program := `
main:
    MOV SP, #0x50000
    MOV R0, #0x1234
    PUSH {R0}
    MOV R1, #0x5678
    PUSH {R1}
    SWI #0x00
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Execute until we've pushed values
	for i := 0; i < 5; i++ {
		s.Step()
	}

	// Get stack contents
	stack := s.GetStack(0, 4)

	if len(stack) == 0 {
		t.Error("Expected non-empty stack")
	}

	// Stack should contain pushed values
	// Note: Stack grows downward, so most recent push is at lower address
	foundValue := false
	for _, entry := range stack {
		if entry.Value == 0x5678 || entry.Value == 0x1234 {
			foundValue = true
			break
		}
	}

	if !foundValue {
		t.Error("Expected to find pushed values on stack")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetStack`

Expected: FAIL with "s.GetStack undefined"

**Step 3: Implement GetStack method**

Add to `service/debugger_service.go`:

```go
// GetStack returns stack contents from SP+offset
func (s *DebuggerService) GetStack(offset int, count int) []StackEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil || s.debugger.VM == nil {
		return []StackEntry{}
	}

	entries := make([]StackEntry, 0, count)
	sp := s.debugger.VM.Registers[13] // R13 is SP

	// Calculate starting address
	startAddr := sp + uint32(offset*4)

	for i := 0; i < count; i++ {
		addr := startAddr + uint32(i*4)

		// Read value from memory
		value, err := s.debugger.VM.Memory.ReadWord(addr)
		if err != nil {
			break
		}

		// Check if value points to a symbol
		symbol := s.GetSymbolForAddress(value)

		entry := StackEntry{
			Address: addr,
			Value:   value,
			Symbol:  symbol,
		}

		entries = append(entries, entry)
	}

	return entries
}
```

**Step 4: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetStack`

Expected: PASS

**Step 5: Commit GetStack**

```bash
git add service/debugger_service.go tests/unit/service/debugger_service_test.go
git commit -m "feat(service): add GetStack API for stack data access"
```

---

## Task 6: Implement Event-Emitting Writer

**Files:**
- Create: `service/event_writer.go`
- Test: `tests/unit/service/event_writer_test.go`

**Step 1: Write failing test for event-emitting writer**

Create `tests/unit/service/event_writer_test.go`:

```go
package service_test

import (
	"bytes"
	"context"
	"testing"
	"github.com/yourusername/arm-emulator/service"
)

func TestEventEmittingWriter_Write(t *testing.T) {
	buffer := &bytes.Buffer{}
	ctx := context.Background()

	// Create event-emitting writer
	writer := service.NewEventEmittingWriter(buffer, ctx)

	// Write some data
	data := []byte("Hello, World!")
	n, err := writer.Write(data)

	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	// Check buffer contains data
	if buffer.String() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", buffer.String())
	}
}

func TestEventEmittingWriter_GetBufferAndClear(t *testing.T) {
	buffer := &bytes.Buffer{}
	ctx := context.Background()

	writer := service.NewEventEmittingWriter(buffer, ctx)

	// Write data
	writer.Write([]byte("Test output"))

	// Get buffer contents and clear
	output := writer.GetBufferAndClear()

	if output != "Test output" {
		t.Errorf("Expected 'Test output', got '%s'", output)
	}

	// Buffer should be empty now
	if buffer.Len() != 0 {
		t.Errorf("Expected empty buffer, got %d bytes", buffer.Len())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestEventEmittingWriter`

Expected: FAIL with "undefined: service.NewEventEmittingWriter"

**Step 3: Implement event-emitting writer**

Create `service/event_writer.go`:

```go
package service

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// EventEmittingWriter wraps a buffer and emits events when written to
type EventEmittingWriter struct {
	buffer *bytes.Buffer
	ctx    context.Context
	mutex  sync.Mutex
}

// NewEventEmittingWriter creates a new event-emitting writer
func NewEventEmittingWriter(buffer *bytes.Buffer, ctx context.Context) *EventEmittingWriter {
	return &EventEmittingWriter{
		buffer: buffer,
		ctx:    ctx,
	}
}

// Write implements io.Writer interface
func (w *EventEmittingWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	n, err = w.buffer.Write(p)
	if err == nil && n > 0 && w.ctx != nil {
		// Emit event with the new output
		runtime.EventsEmit(w.ctx, "vm:output", string(p))
	}
	return n, err
}

// GetBufferAndClear returns buffer contents and clears it
func (w *EventEmittingWriter) GetBufferAndClear() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	output := w.buffer.String()
	w.buffer.Reset()
	return output
}

// Ensure EventEmittingWriter implements io.Writer
var _ io.Writer = (*EventEmittingWriter)(nil)
```

**Step 4: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestEventEmittingWriter`

Expected: PASS (2 tests)

**Step 5: Commit event-emitting writer**

```bash
git add service/event_writer.go tests/unit/service/event_writer_test.go
git commit -m "feat(service): add event-emitting writer for real-time output streaming"
```

---

## Task 7: Integrate Event-Emitting Writer with DebuggerService

**Files:**
- Modify: `service/debugger_service.go`
- Test: `tests/unit/service/debugger_service_test.go`

**Step 1: Write failing test for GetOutput**

Add to `tests/unit/service/debugger_service_test.go`:

```go
func TestDebuggerService_GetOutput(t *testing.T) {
	s := service.NewDebuggerService()

	program := `
main:
    MOV R0, #42
    SWI #0x03  ; Write integer
    SWI #0x00
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Execute program
	s.RunUntilHalt()

	// Get output
	output := s.GetOutput()

	if output == "" {
		t.Error("Expected non-empty output")
	}

	// Second call should return empty (buffer cleared)
	output2 := s.GetOutput()
	if output2 != "" {
		t.Errorf("Expected empty output after clear, got '%s'", output2)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetOutput`

Expected: FAIL with "s.GetOutput undefined"

**Step 3: Add output writer field to DebuggerService**

Modify `service/debugger_service.go` - add field to struct:

```go
type DebuggerService struct {
	debugger     *debugger.Debugger
	mu           sync.Mutex
	outputWriter *EventEmittingWriter
	ctx          context.Context
}
```

**Step 4: Add SetContext method for Wails context**

Add to `service/debugger_service.go`:

```go
// SetContext sets the Wails context for event emission
func (s *DebuggerService) SetContext(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctx = ctx
}
```

**Step 5: Modify LoadProgram to use event-emitting writer**

Modify `LoadProgram` in `service/debugger_service.go`:

Find the section where output is set and replace with:

```go
	// Create output buffer with event emission
	outputBuffer := &bytes.Buffer{}
	s.outputWriter = NewEventEmittingWriter(outputBuffer, s.ctx)
	vm.Output = s.outputWriter
```

**Step 6: Implement GetOutput method**

Add to `service/debugger_service.go`:

```go
// GetOutput returns captured program output (clears buffer)
func (s *DebuggerService) GetOutput() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.outputWriter == nil {
		return ""
	}

	return s.outputWriter.GetBufferAndClear()
}
```

**Step 7: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetOutput`

Expected: PASS

**Step 8: Commit output integration**

```bash
git add service/debugger_service.go tests/unit/service/debugger_service_test.go
git commit -m "feat(service): integrate event-emitting writer with DebuggerService"
```

---

## Task 8: Add Advanced Debugger APIs (StepOver, StepOut, Watchpoints)

**Files:**
- Modify: `service/debugger_service.go`
- Test: `tests/unit/service/debugger_service_test.go`

**Step 1: Write failing tests for advanced stepping**

Add to `tests/unit/service/debugger_service_test.go`:

```go
func TestDebuggerService_StepOver(t *testing.T) {
	s := service.NewDebuggerService()

	program := `
main:
    BL function
    MOV R0, #1
    SWI #0x00
function:
    MOV R1, #2
    MOV PC, LR
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Step over should execute the function and stop at next line
	err = s.StepOver()
	if err != nil {
		t.Errorf("StepOver failed: %v", err)
	}

	// PC should be at "MOV R0, #1" not inside function
	regs := s.GetRegisterState()
	if regs.PC != 0x00008004 {
		t.Errorf("Expected PC at 0x00008004 after step over, got 0x%08X", regs.PC)
	}
}

func TestDebuggerService_StepOut(t *testing.T) {
	s := service.NewDebuggerService()

	program := `
main:
    BL function
    MOV R0, #1
    SWI #0x00
function:
    MOV R1, #2
    MOV R2, #3
    MOV PC, LR
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Step into function
	s.Step()

	// Now step out
	err = s.StepOut()
	if err != nil {
		t.Errorf("StepOut failed: %v", err)
	}

	// Should be back in main after function call
	regs := s.GetRegisterState()
	if regs.PC < 0x00008004 {
		t.Errorf("Expected PC back in main, got 0x%08X", regs.PC)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run "TestDebuggerService_Step(Over|Out)"`

Expected: FAIL with "s.StepOver undefined"

**Step 3: Implement StepOver and StepOut methods**

Add to `service/debugger_service.go`:

```go
// StepOver executes one instruction, stepping over function calls
func (s *DebuggerService) StepOver() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil {
		return fmt.Errorf("no program loaded")
	}

	return s.debugger.CmdNext()
}

// StepOut executes until the current function returns
func (s *DebuggerService) StepOut() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil {
		return fmt.Errorf("no program loaded")
	}

	return s.debugger.CmdFinish()
}
```

**Step 4: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run "TestDebuggerService_Step(Over|Out)"`

Expected: PASS (2 tests)

**Step 5: Write failing tests for watchpoint management**

Add to `tests/unit/service/debugger_service_test.go`:

```go
func TestDebuggerService_AddWatchpoint(t *testing.T) {
	s := service.NewDebuggerService()

	program := `
main:
    MOV R0, #0x10000
    MOV R1, #42
    STR R1, [R0]
    SWI #0x00
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Add watchpoint
	err = s.AddWatchpoint(0x10000, "write")
	if err != nil {
		t.Errorf("AddWatchpoint failed: %v", err)
	}

	// Get watchpoints
	watchpoints := s.GetWatchpoints()
	if len(watchpoints) == 0 {
		t.Error("Expected watchpoint to be added")
	}
}

func TestDebuggerService_RemoveWatchpoint(t *testing.T) {
	s := service.NewDebuggerService()

	program := `main: SWI #0x00`
	s.LoadProgram([]byte(program), []string{})

	s.AddWatchpoint(0x10000, "write")

	err := s.RemoveWatchpoint(0x10000)
	if err != nil {
		t.Errorf("RemoveWatchpoint failed: %v", err)
	}

	watchpoints := s.GetWatchpoints()
	if len(watchpoints) != 0 {
		t.Error("Expected watchpoint to be removed")
	}
}
```

**Step 6: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run "TestDebuggerService_.*Watchpoint"`

Expected: FAIL with "s.AddWatchpoint undefined"

**Step 7: Implement watchpoint management methods**

Add to `service/debugger_service.go`:

```go
// AddWatchpoint adds a watchpoint at the specified address
func (s *DebuggerService) AddWatchpoint(address uint32, watchType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil {
		return fmt.Errorf("no program loaded")
	}

	// Convert string type to debugger type
	var wpType int
	switch watchType {
	case "read":
		wpType = 1 // Assuming debugger package constants
	case "write":
		wpType = 2
	case "readwrite":
		wpType = 3
	default:
		return fmt.Errorf("invalid watchpoint type: %s", watchType)
	}

	return s.debugger.AddWatchpoint(address, wpType)
}

// RemoveWatchpoint removes a watchpoint at the specified address
func (s *DebuggerService) RemoveWatchpoint(address uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil {
		return fmt.Errorf("no program loaded")
	}

	return s.debugger.RemoveWatchpoint(address)
}
```

**Step 8: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run "TestDebuggerService_.*Watchpoint"`

Expected: PASS (2 tests)

**Step 9: Commit advanced debugger APIs**

```bash
git add service/debugger_service.go tests/unit/service/debugger_service_test.go
git commit -m "feat(service): add StepOver, StepOut, and watchpoint management APIs"
```

---

## Task 9: Add ExecuteCommand and EvaluateExpression APIs

**Files:**
- Modify: `service/debugger_service.go`
- Test: `tests/unit/service/debugger_service_test.go`

**Step 1: Write failing test for ExecuteCommand**

Add to `tests/unit/service/debugger_service_test.go`:

```go
func TestDebuggerService_ExecuteCommand(t *testing.T) {
	s := service.NewDebuggerService()

	program := `
main:
    MOV R0, #42
    SWI #0x00
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Execute "info registers" command
	output, err := s.ExecuteCommand("info registers")
	if err != nil {
		t.Errorf("ExecuteCommand failed: %v", err)
	}

	if output == "" {
		t.Error("Expected non-empty command output")
	}
}

func TestDebuggerService_EvaluateExpression(t *testing.T) {
	s := service.NewDebuggerService()

	program := `
main:
    MOV R0, #42
    MOV R1, #10
    SWI #0x00
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Execute first two instructions
	s.Step()
	s.Step()

	// Evaluate "R0 + R1"
	result, err := s.EvaluateExpression("R0 + R1")
	if err != nil {
		t.Errorf("EvaluateExpression failed: %v", err)
	}

	expected := uint32(52) // 42 + 10
	if result != expected {
		t.Errorf("Expected result %d, got %d", expected, result)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run "TestDebuggerService_(ExecuteCommand|EvaluateExpression)"`

Expected: FAIL with "s.ExecuteCommand undefined"

**Step 3: Implement ExecuteCommand method**

Add to `service/debugger_service.go`:

```go
// ExecuteCommand executes a debugger command and returns output
func (s *DebuggerService) ExecuteCommand(command string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil {
		return "", fmt.Errorf("no program loaded")
	}

	// Capture output in buffer
	outputBuf := &bytes.Buffer{}
	oldOutput := s.debugger.VM.Output
	s.debugger.VM.Output = outputBuf

	// Execute command
	err := s.debugger.ExecuteCommand(command)

	// Restore original output
	s.debugger.VM.Output = oldOutput

	return outputBuf.String(), err
}
```

**Step 4: Implement EvaluateExpression method**

Add to `service/debugger_service.go`:

```go
// EvaluateExpression evaluates an expression and returns the result
func (s *DebuggerService) EvaluateExpression(expr string) (uint32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil || s.debugger.Evaluator == nil {
		return 0, fmt.Errorf("no program loaded")
	}

	return s.debugger.Evaluator.Evaluate(expr, s.debugger.VM, s.debugger.Symbols)
}
```

**Step 5: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run "TestDebuggerService_(ExecuteCommand|EvaluateExpression)"`

Expected: PASS (2 tests)

**Step 6: Commit command execution APIs**

```bash
git add service/debugger_service.go tests/unit/service/debugger_service_test.go
git commit -m "feat(service): add ExecuteCommand and EvaluateExpression APIs"
```

---

## Task 10: Add Event Emission to App.go

**Files:**
- Modify: `gui/app.go`

**Step 1: Store context in App struct**

Modify `gui/app.go` - update App struct:

```go
type App struct {
	ctx     context.Context
	service *service.DebuggerService
}
```

**Step 2: Update NewApp to store context**

Modify `NewApp` function in `gui/app.go`:

```go
func NewApp() *App {
	return &App{
		service: service.NewDebuggerService(),
	}
}
```

**Step 3: Update startup to set context**

Modify `startup` method in `gui/app.go`:

```go
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.service.SetContext(ctx)
}
```

**Step 4: Add event emission to Step method**

Modify `Step` method in `gui/app.go`:

```go
func (a *App) Step() error {
	err := a.service.Step()
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	} else {
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
	}
	return err
}
```

**Step 5: Add event emission to Continue method**

Modify `Continue` method in `gui/app.go`:

```go
func (a *App) Continue() error {
	err := a.service.RunUntilHalt()
	runtime.EventsEmit(a.ctx, "vm:state-changed")

	if err != nil {
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
	}

	// Check if stopped at breakpoint
	if a.service.GetExecutionState() == service.StateBreakpoint {
		runtime.EventsEmit(a.ctx, "vm:breakpoint-hit")
	}

	return err
}
```

**Step 6: Add event emission to Reset method**

Modify `Reset` method in `gui/app.go`:

```go
func (a *App) Reset() error {
	err := a.service.Reset()
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}
	return err
}
```

**Step 7: Add event emission to Pause method**

Modify `Pause` method in `gui/app.go`:

```go
func (a *App) Pause() error {
	err := a.service.Pause()
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}
	return err
}
```

**Step 8: Test manually**

Run: `go build -o arm-emulator`

Expected: Builds successfully

**Step 9: Commit event emission**

```bash
git add gui/app.go
git commit -m "feat(gui): add Wails event emission to state-changing methods"
```

---

## Task 11: Add ToggleBreakpoint Wrapper to App.go

**Files:**
- Modify: `gui/app.go`

**Step 1: Add ToggleBreakpoint method**

Add to `gui/app.go`:

```go
// ToggleBreakpoint toggles a breakpoint at the specified address
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

**Step 2: Add wrappers for new service methods**

Add to `gui/app.go`:

```go
// GetSourceMap returns the complete source map
func (a *App) GetSourceMap() map[uint32]string {
	return a.service.GetSourceMap()
}

// GetDisassembly returns disassembled instructions
func (a *App) GetDisassembly(startAddr uint32, count int) []service.DisassemblyLine {
	return a.service.GetDisassembly(startAddr, count)
}

// GetStack returns stack contents
func (a *App) GetStack(offset int, count int) []service.StackEntry {
	return a.service.GetStack(offset, count)
}

// GetSymbolForAddress resolves address to symbol
func (a *App) GetSymbolForAddress(addr uint32) string {
	return a.service.GetSymbolForAddress(addr)
}

// GetOutput returns captured output
func (a *App) GetOutput() string {
	return a.service.GetOutput()
}

// GetMemoryData returns memory data with write tracking
func (a *App) GetMemoryData(startAddr uint32, length int) service.MemoryData {
	return a.service.GetMemoryData(startAddr, length)
}

// StepOver steps over function calls
func (a *App) StepOver() error {
	err := a.service.StepOver()
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	} else {
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
	}
	return err
}

// StepOut steps out of current function
func (a *App) StepOut() error {
	err := a.service.StepOut()
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	} else {
		runtime.EventsEmit(a.ctx, "vm:error", err.Error())
	}
	return err
}

// AddWatchpoint adds a watchpoint
func (a *App) AddWatchpoint(address uint32, watchType string) error {
	err := a.service.AddWatchpoint(address, watchType)
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}
	return err
}

// RemoveWatchpoint removes a watchpoint
func (a *App) RemoveWatchpoint(address uint32) error {
	err := a.service.RemoveWatchpoint(address)
	if err == nil {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}
	return err
}

// ExecuteCommand executes a debugger command
func (a *App) ExecuteCommand(command string) (string, error) {
	output, err := a.service.ExecuteCommand(command)

	// Check if command modified state
	if isStateModifyingCommand(command) {
		runtime.EventsEmit(a.ctx, "vm:state-changed")
	}

	return output, err
}

// EvaluateExpression evaluates an expression
func (a *App) EvaluateExpression(expr string) (uint32, error) {
	return a.service.EvaluateExpression(expr)
}

// isStateModifyingCommand checks if command modifies VM state
func isStateModifyingCommand(command string) bool {
	stateCommands := []string{"step", "next", "finish", "continue", "set", "break", "delete"}
	for _, cmd := range stateCommands {
		if strings.HasPrefix(strings.ToLower(command), cmd) {
			return true
		}
	}
	return false
}
```

**Step 3: Test build**

Run: `go build -o arm-emulator`

Expected: Builds successfully

**Step 4: Commit App.go wrappers**

```bash
git add gui/app.go
git commit -m "feat(gui): add wrapper methods for new service APIs with event emission"
```

---

## Task 12: Install Frontend Dependencies

**Files:**
- Modify: `gui/frontend/package.json`

**Step 1: Navigate to frontend directory**

Run: `cd gui/frontend`

**Step 2: Install allotment library**

Run: `npm install allotment`

Expected: Package installed successfully

**Step 3: Verify installation**

Run: `npm list allotment`

Expected: Shows installed version

**Step 4: Return to project root**

Run: `cd ../..`

**Step 5: Commit package.json and package-lock.json**

```bash
git add gui/frontend/package.json gui/frontend/package-lock.json
git commit -m "feat(gui): install allotment library for resizable panels"
```

---

## Task 13: Create SourceView Component

**Files:**
- Create: `gui/frontend/src/components/SourceView.tsx`
- Create: `gui/frontend/src/components/SourceView.css`

**Step 1: Create SourceView component file**

Create `gui/frontend/src/components/SourceView.tsx`:

```tsx
import React, { useEffect, useState, useRef } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetSourceMap, GetRegisterState, ToggleBreakpoint, GetBreakpoints } from '../../wailsjs/go/main/App';
import './SourceView.css';

interface SourceLine {
  address: number;
  source: string;
  hasBreakpoint: boolean;
  isCurrent: boolean;
  symbol: string;
}

export const SourceView: React.FC = () => {
  const [lines, setLines] = useState<SourceLine[]>([]);
  const [currentPC, setCurrentPC] = useState<number>(0);
  const containerRef = useRef<HTMLDivElement>(null);

  const loadSourceData = async () => {
    try {
      const sourceMap = await GetSourceMap();
      const registerState = await GetRegisterState();
      const breakpoints = await GetBreakpoints();

      const pc = registerState.pc;
      setCurrentPC(pc);

      // Convert source map to sorted array
      const sourceLines: SourceLine[] = [];
      const breakpointAddresses = new Set(breakpoints.map(bp => bp.address));

      for (const [addrStr, source] of Object.entries(sourceMap)) {
        const address = parseInt(addrStr);
        sourceLines.push({
          address,
          source,
          hasBreakpoint: breakpointAddresses.has(address),
          isCurrent: address === pc,
          symbol: '', // Will be populated if needed
        });
      }

      // Sort by address
      sourceLines.sort((a, b) => a.address - b.address);

      setLines(sourceLines);

      // Auto-scroll to current PC
      setTimeout(() => scrollToCurrentLine(), 100);
    } catch (error) {
      console.error('Failed to load source data:', error);
    }
  };

  const scrollToCurrentLine = () => {
    if (containerRef.current) {
      const currentLine = containerRef.current.querySelector('.source-line-current');
      if (currentLine) {
        currentLine.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
    }
  };

  const handleLineClick = async (address: number) => {
    try {
      await ToggleBreakpoint(address);
      // State will update via event
    } catch (error) {
      console.error('Failed to toggle breakpoint:', error);
    }
  };

  useEffect(() => {
    // Initial load
    loadSourceData();

    // Subscribe to VM state changes
    EventsOn('vm:state-changed', loadSourceData);

    return () => {
      EventsOff('vm:state-changed');
    };
  }, []);

  return (
    <div className="source-view" ref={containerRef}>
      <div className="source-header">Source Code</div>
      <div className="source-content">
        {lines.map((line, index) => (
          <div
            key={index}
            className={`source-line ${line.isCurrent ? 'source-line-current' : ''} ${line.hasBreakpoint ? 'source-line-breakpoint' : ''}`}
            onClick={() => handleLineClick(line.address)}
          >
            <span className="source-line-number">
              {line.hasBreakpoint && <span className="breakpoint-marker">●</span>}
              {line.address.toString(16).padStart(8, '0')}
            </span>
            <span className="source-line-text">{line.source}</span>
          </div>
        ))}
      </div>
    </div>
  );
};
```

**Step 2: Create SourceView CSS**

Create `gui/frontend/src/components/SourceView.css`:

```css
.source-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', monospace;
}

.source-header {
  padding: 8px 12px;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  font-weight: bold;
  font-size: 12px;
}

.source-content {
  flex: 1;
  overflow-y: auto;
  padding: 4px;
}

.source-line {
  display: flex;
  padding: 2px 4px;
  cursor: pointer;
  font-size: 13px;
  line-height: 1.5;
}

.source-line:hover {
  background: #2a2a2a;
}

.source-line-current {
  background: #4a4a00 !important;
}

.source-line-breakpoint {
  background: #3e1e1e;
}

.source-line-number {
  min-width: 100px;
  color: #858585;
  user-select: none;
  padding-right: 12px;
}

.breakpoint-marker {
  color: #e51400;
  margin-right: 4px;
  font-weight: bold;
}

.source-line-text {
  flex: 1;
  white-space: pre;
}
```

**Step 3: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 4: Commit SourceView component**

```bash
git add gui/frontend/src/components/SourceView.tsx gui/frontend/src/components/SourceView.css
git commit -m "feat(gui): add SourceView component with breakpoint support"
```

---

## Task 14: Create DisassemblyView Component

**Files:**
- Create: `gui/frontend/src/components/DisassemblyView.tsx`
- Create: `gui/frontend/src/components/DisassemblyView.css`

**Step 1: Create DisassemblyView component**

Create `gui/frontend/src/components/DisassemblyView.tsx`:

```tsx
import React, { useEffect, useState } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetDisassembly, GetRegisterState, GetBreakpoints, ToggleBreakpoint } from '../../wailsjs/go/main/App';
import { service } from '../../wailsjs/go/models';
import './DisassemblyView.css';

interface DisassemblyLineView extends service.DisassemblyLine {
  hasBreakpoint: boolean;
  isCurrent: boolean;
}

export const DisassemblyView: React.FC = () => {
  const [lines, setLines] = useState<DisassemblyLineView[]>([]);

  const loadDisassembly = async () => {
    try {
      const registerState = await GetRegisterState();
      const breakpoints = await GetBreakpoints();
      const pc = registerState.pc;

      // Get disassembly around PC
      const startAddr = Math.max(0, pc - 20 * 4);
      const disasm = await GetDisassembly(startAddr, 15);

      const breakpointAddresses = new Set(breakpoints.map(bp => bp.address));

      const linesWithState = disasm.map(line => ({
        ...line,
        hasBreakpoint: breakpointAddresses.has(line.address),
        isCurrent: line.address === pc,
      }));

      setLines(linesWithState);
    } catch (error) {
      console.error('Failed to load disassembly:', error);
    }
  };

  const handleLineClick = async (address: number) => {
    try {
      await ToggleBreakpoint(address);
    } catch (error) {
      console.error('Failed to toggle breakpoint:', error);
    }
  };

  useEffect(() => {
    loadDisassembly();
    EventsOn('vm:state-changed', loadDisassembly);

    return () => {
      EventsOff('vm:state-changed');
    };
  }, []);

  return (
    <div className="disassembly-view">
      <div className="disassembly-header">Disassembly</div>
      <div className="disassembly-content">
        {lines.map((line, index) => (
          <div
            key={index}
            className={`disasm-line ${line.isCurrent ? 'disasm-line-current' : ''} ${line.hasBreakpoint ? 'disasm-line-breakpoint' : ''}`}
            onClick={() => handleLineClick(line.address)}
          >
            <span className="disasm-address">
              {line.hasBreakpoint && <span className="breakpoint-marker">●</span>}
              {line.address.toString(16).padStart(8, '0')}
            </span>
            <span className="disasm-opcode">
              {line.opcode.toString(16).padStart(8, '0')}
            </span>
            {line.symbol && (
              <span className="disasm-symbol">{line.symbol}:</span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};
```

**Step 2: Create DisassemblyView CSS**

Create `gui/frontend/src/components/DisassemblyView.css`:

```css
.disassembly-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', monospace;
}

.disassembly-header {
  padding: 8px 12px;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  font-weight: bold;
  font-size: 12px;
}

.disassembly-content {
  flex: 1;
  overflow-y: auto;
  padding: 4px;
}

.disasm-line {
  display: flex;
  padding: 2px 4px;
  cursor: pointer;
  font-size: 13px;
  line-height: 1.5;
  gap: 12px;
}

.disasm-line:hover {
  background: #2a2a2a;
}

.disasm-line-current {
  background: #4a4a00 !important;
}

.disasm-line-breakpoint {
  background: #3e1e1e;
}

.disasm-address {
  min-width: 100px;
  color: #858585;
  user-select: none;
}

.disasm-opcode {
  min-width: 80px;
  color: #9cdcfe;
}

.disasm-symbol {
  color: #4ec9b0;
  font-weight: bold;
}

.breakpoint-marker {
  color: #e51400;
  margin-right: 4px;
  font-weight: bold;
}
```

**Step 3: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 4: Commit DisassemblyView component**

```bash
git add gui/frontend/src/components/DisassemblyView.tsx gui/frontend/src/components/DisassemblyView.css
git commit -m "feat(gui): add DisassemblyView component with instruction display"
```

---

## Task 15: Create StackView Component

**Files:**
- Create: `gui/frontend/src/components/StackView.tsx`
- Create: `gui/frontend/src/components/StackView.css`

**Step 1: Create StackView component**

Create `gui/frontend/src/components/StackView.tsx`:

```tsx
import React, { useEffect, useState } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetStack, GetRegisterState } from '../../wailsjs/go/main/App';
import { service } from '../../wailsjs/go/models';
import './StackView.css';

interface StackEntryView extends service.StackEntry {
  isSP: boolean;
}

export const StackView: React.FC = () => {
  const [entries, setEntries] = useState<StackEntryView[]>([]);

  const loadStack = async () => {
    try {
      const registerState = await GetRegisterState();
      const sp = registerState.registers[13]; // R13 is SP
      const stackData = await GetStack(0, 16);

      const entriesWithSP = stackData.map(entry => ({
        ...entry,
        isSP: entry.address === sp,
      }));

      setEntries(entriesWithSP);
    } catch (error) {
      console.error('Failed to load stack:', error);
    }
  };

  useEffect(() => {
    loadStack();
    EventsOn('vm:state-changed', loadStack);

    return () => {
      EventsOff('vm:state-changed');
    };
  }, []);

  return (
    <div className="stack-view">
      <div className="stack-header">Stack</div>
      <div className="stack-content">
        {entries.map((entry, index) => (
          <div key={index} className={`stack-entry ${entry.isSP ? 'stack-entry-sp' : ''}`}>
            {entry.isSP && <span className="sp-marker">→</span>}
            <span className="stack-address">
              {entry.address.toString(16).padStart(8, '0')}
            </span>
            <span className="stack-value">
              {entry.value.toString(16).padStart(8, '0')}
            </span>
            {entry.symbol && (
              <span className="stack-symbol">{entry.symbol}</span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};
```

**Step 2: Create StackView CSS**

Create `gui/frontend/src/components/StackView.css`:

```css
.stack-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', monospace;
}

.stack-header {
  padding: 8px 12px;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  font-weight: bold;
  font-size: 12px;
}

.stack-content {
  flex: 1;
  overflow-y: auto;
  padding: 4px;
}

.stack-entry {
  display: flex;
  padding: 2px 4px;
  font-size: 13px;
  line-height: 1.5;
  gap: 12px;
  align-items: center;
}

.stack-entry-sp {
  background: #2a3a2a;
  font-weight: bold;
}

.sp-marker {
  color: #4ec9b0;
  font-weight: bold;
  min-width: 16px;
}

.stack-address {
  min-width: 80px;
  color: #858585;
}

.stack-value {
  min-width: 80px;
  color: #9cdcfe;
}

.stack-symbol {
  color: #4ec9b0;
  font-style: italic;
}
```

**Step 3: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 4: Commit StackView component**

```bash
git add gui/frontend/src/components/StackView.tsx gui/frontend/src/components/StackView.css
git commit -m "feat(gui): add StackView component with SP highlighting"
```

---

## Task 16: Create OutputView Component

**Files:**
- Create: `gui/frontend/src/components/OutputView.tsx`
- Create: `gui/frontend/src/components/OutputView.css`

**Step 1: Create OutputView component**

Create `gui/frontend/src/components/OutputView.tsx`:

```tsx
import React, { useEffect, useState, useRef } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import './OutputView.css';

export const OutputView: React.FC = () => {
  const [output, setOutput] = useState<string>('');
  const contentRef = useRef<HTMLPreElement>(null);

  const handleOutputEvent = (data: string) => {
    setOutput(prev => prev + data);

    // Auto-scroll to bottom
    setTimeout(() => {
      if (contentRef.current) {
        contentRef.current.scrollTop = contentRef.current.scrollHeight;
      }
    }, 10);
  };

  const handleClear = () => {
    setOutput('');
  };

  useEffect(() => {
    EventsOn('vm:output', handleOutputEvent);

    return () => {
      EventsOff('vm:output');
    };
  }, []);

  return (
    <div className="output-view">
      <div className="output-header">
        <span>Program Output</span>
        <button className="output-clear-btn" onClick={handleClear}>Clear</button>
      </div>
      <pre className="output-content" ref={contentRef}>
        {output || '(no output)'}
      </pre>
    </div>
  );
};
```

**Step 2: Create OutputView CSS**

Create `gui/frontend/src/components/OutputView.css`:

```css
.output-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', monospace;
}

.output-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  font-weight: bold;
  font-size: 12px;
}

.output-clear-btn {
  background: #3e3e3e;
  color: #d4d4d4;
  border: none;
  padding: 4px 8px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 11px;
}

.output-clear-btn:hover {
  background: #4e4e4e;
}

.output-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-wrap: break-word;
}
```

**Step 3: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 4: Commit OutputView component**

```bash
git add gui/frontend/src/components/OutputView.tsx gui/frontend/src/components/OutputView.css
git commit -m "feat(gui): add OutputView component with real-time output streaming"
```

---

## Task 17: Create StatusView Component

**Files:**
- Create: `gui/frontend/src/components/StatusView.tsx`
- Create: `gui/frontend/src/components/StatusView.css`

**Step 1: Create StatusView component**

Create `gui/frontend/src/components/StatusView.tsx`:

```tsx
import React, { useEffect, useState } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetExecutionState, GetRegisterState } from '../../wailsjs/go/main/App';
import './StatusView.css';

interface StatusMessage {
  type: 'info' | 'error' | 'breakpoint';
  message: string;
  timestamp: Date;
}

export const StatusView: React.FC = () => {
  const [messages, setMessages] = useState<StatusMessage[]>([]);
  const [executionState, setExecutionState] = useState<string>('');
  const [cycles, setCycles] = useState<number>(0);

  const addMessage = (type: StatusMessage['type'], message: string) => {
    setMessages(prev => [
      ...prev,
      { type, message, timestamp: new Date() }
    ].slice(-50)); // Keep last 50 messages
  };

  const loadState = async () => {
    try {
      const state = await GetExecutionState();
      const registerState = await GetRegisterState();

      setExecutionState(state);
      setCycles(registerState.cycles || 0);
    } catch (error) {
      console.error('Failed to load state:', error);
    }
  };

  useEffect(() => {
    loadState();

    EventsOn('vm:state-changed', () => {
      loadState();
      addMessage('info', 'VM state changed');
    });

    EventsOn('vm:error', (errorMsg: string) => {
      addMessage('error', errorMsg);
    });

    EventsOn('vm:breakpoint-hit', () => {
      addMessage('breakpoint', 'Breakpoint hit');
    });

    return () => {
      EventsOff('vm:state-changed');
      EventsOff('vm:error');
      EventsOff('vm:breakpoint-hit');
    };
  }, []);

  return (
    <div className="status-view">
      <div className="status-header">
        <span>Debugger Status</span>
        <div className="status-info">
          <span className="status-state">{executionState}</span>
          <span className="status-cycles">Cycles: {cycles}</span>
        </div>
      </div>
      <div className="status-content">
        {messages.map((msg, index) => (
          <div key={index} className={`status-message status-message-${msg.type}`}>
            <span className="status-timestamp">
              {msg.timestamp.toLocaleTimeString()}
            </span>
            <span className="status-text">{msg.message}</span>
          </div>
        ))}
      </div>
    </div>
  );
};
```

**Step 2: Create StatusView CSS**

Create `gui/frontend/src/components/StatusView.css`:

```css
.status-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', monospace;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  font-weight: bold;
  font-size: 12px;
}

.status-info {
  display: flex;
  gap: 16px;
  font-weight: normal;
}

.status-state {
  color: #4ec9b0;
}

.status-cycles {
  color: #9cdcfe;
}

.status-content {
  flex: 1;
  overflow-y: auto;
  padding: 4px;
}

.status-message {
  display: flex;
  gap: 8px;
  padding: 4px 8px;
  font-size: 12px;
  line-height: 1.4;
  border-left: 3px solid transparent;
}

.status-message-info {
  border-left-color: #4ec9b0;
}

.status-message-error {
  border-left-color: #e51400;
  background: #3e1e1e;
}

.status-message-breakpoint {
  border-left-color: #ce9178;
  background: #3e3e1e;
}

.status-timestamp {
  color: #858585;
  min-width: 80px;
}

.status-text {
  flex: 1;
}
```

**Step 3: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 4: Commit StatusView component**

```bash
git add gui/frontend/src/components/StatusView.tsx gui/frontend/src/components/StatusView.css
git commit -m "feat(gui): add StatusView component with execution state and messages"
```

---

## Task 18: Create BreakpointsView Component

**Files:**
- Create: `gui/frontend/src/components/BreakpointsView.tsx`
- Create: `gui/frontend/src/components/BreakpointsView.css`

**Step 1: Create BreakpointsView component**

Create `gui/frontend/src/components/BreakpointsView.tsx`:

```tsx
import React, { useEffect, useState } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetBreakpoints, RemoveBreakpoint, GetWatchpoints, RemoveWatchpoint } from '../../wailsjs/go/main/App';
import { service } from '../../wailsjs/go/models';
import './BreakpointsView.css';

export const BreakpointsView: React.FC = () => {
  const [breakpoints, setBreakpoints] = useState<service.BreakpointInfo[]>([]);
  const [watchpoints, setWatchpoints] = useState<service.WatchpointInfo[]>([]);

  const loadBreakpoints = async () => {
    try {
      const bps = await GetBreakpoints();
      const wps = await GetWatchpoints();

      setBreakpoints(bps || []);
      setWatchpoints(wps || []);
    } catch (error) {
      console.error('Failed to load breakpoints:', error);
    }
  };

  const handleRemoveBreakpoint = async (address: number) => {
    try {
      await RemoveBreakpoint(address);
    } catch (error) {
      console.error('Failed to remove breakpoint:', error);
    }
  };

  const handleRemoveWatchpoint = async (address: number) => {
    try {
      await RemoveWatchpoint(address);
    } catch (error) {
      console.error('Failed to remove watchpoint:', error);
    }
  };

  useEffect(() => {
    loadBreakpoints();
    EventsOn('vm:state-changed', loadBreakpoints);

    return () => {
      EventsOff('vm:state-changed');
    };
  }, []);

  return (
    <div className="breakpoints-view">
      <div className="breakpoints-header">Breakpoints & Watchpoints</div>

      <div className="breakpoints-section">
        <div className="section-title">Breakpoints ({breakpoints.length})</div>
        {breakpoints.length === 0 ? (
          <div className="empty-message">No breakpoints set</div>
        ) : (
          <table className="breakpoints-table">
            <thead>
              <tr>
                <th>Address</th>
                <th>Condition</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {breakpoints.map((bp, index) => (
                <tr key={index}>
                  <td className="bp-address">0x{bp.address.toString(16).padStart(8, '0')}</td>
                  <td className="bp-condition">{bp.condition || '(always)'}</td>
                  <td className="bp-actions">
                    <button
                      className="btn-remove"
                      onClick={() => handleRemoveBreakpoint(bp.address)}
                    >
                      Remove
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div className="breakpoints-section">
        <div className="section-title">Watchpoints ({watchpoints.length})</div>
        {watchpoints.length === 0 ? (
          <div className="empty-message">No watchpoints set</div>
        ) : (
          <table className="breakpoints-table">
            <thead>
              <tr>
                <th>Address</th>
                <th>Type</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {watchpoints.map((wp, index) => (
                <tr key={index}>
                  <td className="bp-address">0x{wp.address.toString(16).padStart(8, '0')}</td>
                  <td className="bp-type">{wp.type}</td>
                  <td className="bp-actions">
                    <button
                      className="btn-remove"
                      onClick={() => handleRemoveWatchpoint(wp.address)}
                    >
                      Remove
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
```

**Step 2: Create BreakpointsView CSS**

Create `gui/frontend/src/components/BreakpointsView.css`:

```css
.breakpoints-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', monospace;
}

.breakpoints-header {
  padding: 8px 12px;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  font-weight: bold;
  font-size: 12px;
}

.breakpoints-section {
  padding: 12px;
  border-bottom: 1px solid #3e3e3e;
}

.section-title {
  font-size: 11px;
  font-weight: bold;
  color: #858585;
  text-transform: uppercase;
  margin-bottom: 8px;
}

.empty-message {
  color: #858585;
  font-size: 12px;
  font-style: italic;
  padding: 8px 0;
}

.breakpoints-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.breakpoints-table th {
  text-align: left;
  padding: 6px 8px;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  font-weight: bold;
  font-size: 11px;
}

.breakpoints-table td {
  padding: 6px 8px;
  border-bottom: 1px solid #2e2e2e;
}

.bp-address {
  color: #9cdcfe;
  font-family: 'Consolas', 'Monaco', monospace;
}

.bp-condition,
.bp-type {
  color: #ce9178;
  font-style: italic;
}

.bp-actions {
  text-align: right;
}

.btn-remove {
  background: #e51400;
  color: white;
  border: none;
  padding: 4px 8px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 11px;
}

.btn-remove:hover {
  background: #c51400;
}
```

**Step 3: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 4: Commit BreakpointsView component**

```bash
git add gui/frontend/src/components/BreakpointsView.tsx gui/frontend/src/components/BreakpointsView.css
git commit -m "feat(gui): add BreakpointsView component with watchpoint support"
```

---

## Task 19: Integrate Layout with Allotment

**Files:**
- Modify: `gui/frontend/src/App.tsx`
- Modify: `gui/frontend/src/App.css`

**Step 1: Import allotment and new components in App.tsx**

Modify `gui/frontend/src/App.tsx` - add imports at top:

```tsx
import { Allotment } from 'allotment';
import 'allotment/dist/style.css';
import { SourceView } from './components/SourceView';
import { DisassemblyView } from './components/DisassemblyView';
import { MemoryView } from './components/MemoryView';
import { StackView } from './components/StackView';
import { OutputView } from './components/OutputView';
import { StatusView } from './components/StatusView';
import { BreakpointsView } from './components/BreakpointsView';
import { useState } from 'react';
```

**Step 2: Replace App component layout**

Modify the App component in `gui/frontend/src/App.tsx`:

```tsx
function App() {
  const [leftTab, setLeftTab] = useState<'source' | 'disassembly'>('source');
  const [bottomTab, setBottomTab] = useState<'output' | 'breakpoints' | 'status'>('output');

  return (
    <div className="app-container">
      <Allotment vertical>
        {/* Top toolbar - fixed height */}
        <Allotment.Pane snap minSize={60} maxSize={60}>
          <div className="toolbar">
            <button onClick={() => {}}>Load</button>
            <button onClick={() => {}}>Step</button>
            <button onClick={() => {}}>Step Over</button>
            <button onClick={() => {}}>Step Out</button>
            <button onClick={() => {}}>Run</button>
            <button onClick={() => {}}>Pause</button>
            <button onClick={() => {}}>Reset</button>
          </div>
        </Allotment.Pane>

        {/* Main content area */}
        <Allotment.Pane>
          <Allotment>
            {/* Left: Source/Disassembly tabs */}
            <Allotment.Pane minSize={300} preferredSize={500}>
              <div className="tabbed-panel">
                <div className="tabs">
                  <button
                    className={leftTab === 'source' ? 'tab active' : 'tab'}
                    onClick={() => setLeftTab('source')}
                  >
                    Source
                  </button>
                  <button
                    className={leftTab === 'disassembly' ? 'tab active' : 'tab'}
                    onClick={() => setLeftTab('disassembly')}
                  >
                    Disassembly
                  </button>
                </div>
                <div className="tab-content">
                  {leftTab === 'source' && <SourceView />}
                  {leftTab === 'disassembly' && <DisassemblyView />}
                </div>
              </div>
            </Allotment.Pane>

            {/* Right: Registers/Memory/Stack */}
            <Allotment.Pane minSize={300} preferredSize={400}>
              <Allotment vertical>
                <Allotment.Pane>
                  <div className="placeholder-view">Register View (existing)</div>
                </Allotment.Pane>
                <Allotment.Pane>
                  <MemoryView />
                </Allotment.Pane>
                <Allotment.Pane>
                  <StackView />
                </Allotment.Pane>
              </Allotment>
            </Allotment.Pane>
          </Allotment>
        </Allotment.Pane>

        {/* Bottom: Output/Breakpoints/Status tabs */}
        <Allotment.Pane snap minSize={150} preferredSize={200}>
          <div className="tabbed-panel">
            <div className="tabs">
              <button
                className={bottomTab === 'output' ? 'tab active' : 'tab'}
                onClick={() => setBottomTab('output')}
              >
                Output
              </button>
              <button
                className={bottomTab === 'breakpoints' ? 'tab active' : 'tab'}
                onClick={() => setBottomTab('breakpoints')}
              >
                Breakpoints
              </button>
              <button
                className={bottomTab === 'status' ? 'tab active' : 'tab'}
                onClick={() => setBottomTab('status')}
              >
                Status
              </button>
            </div>
            <div className="tab-content">
              {bottomTab === 'output' && <OutputView />}
              {bottomTab === 'breakpoints' && <BreakpointsView />}
              {bottomTab === 'status' && <StatusView />}
            </div>
          </div>
        </Allotment.Pane>
      </Allotment>
    </div>
  );
}

export default App;
```

**Step 3: Update App.css for new layout**

Modify `gui/frontend/src/App.css`:

```css
.app-container {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: #1e1e1e;
}

.toolbar {
  display: flex;
  gap: 8px;
  padding: 12px;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  align-items: center;
}

.toolbar button {
  background: #3e3e3e;
  color: #d4d4d4;
  border: none;
  padding: 6px 12px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
}

.toolbar button:hover {
  background: #4e4e4e;
}

.tabbed-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
}

.tabs {
  display: flex;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
}

.tab {
  background: transparent;
  color: #858585;
  border: none;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 12px;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.tab:hover {
  color: #d4d4d4;
  background: #2a2a2a;
}

.tab.active {
  color: #d4d4d4;
  border-bottom-color: #4ec9b0;
}

.tab-content {
  flex: 1;
  overflow: hidden;
}

.placeholder-view {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background: #1e1e1e;
  color: #858585;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
}
```

**Step 4: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 5: Commit layout integration**

```bash
git add gui/frontend/src/App.tsx gui/frontend/src/App.css
git commit -m "feat(gui): integrate new debugging views with allotment resizable layout"
```

---

## Task 20: Add Toolbar Button Functionality

**Files:**
- Modify: `gui/frontend/src/App.tsx`

**Step 1: Import Wails bindings**

Add imports to `gui/frontend/src/App.tsx`:

```tsx
import {
  Step,
  StepOver,
  StepOut,
  Continue,
  Pause,
  Reset
} from '../wailsjs/go/main/App';
```

**Step 2: Add button click handlers**

Add handler functions in App component:

```tsx
  const handleStep = async () => {
    try {
      await Step();
    } catch (error) {
      console.error('Step failed:', error);
    }
  };

  const handleStepOver = async () => {
    try {
      await StepOver();
    } catch (error) {
      console.error('Step Over failed:', error);
    }
  };

  const handleStepOut = async () => {
    try {
      await StepOut();
    } catch (error) {
      console.error('Step Out failed:', error);
    }
  };

  const handleRun = async () => {
    try {
      await Continue();
    } catch (error) {
      console.error('Continue failed:', error);
    }
  };

  const handlePause = async () => {
    try {
      await Pause();
    } catch (error) {
      console.error('Pause failed:', error);
    }
  };

  const handleReset = async () => {
    try {
      await Reset();
    } catch (error) {
      console.error('Reset failed:', error);
    }
  };
```

**Step 3: Connect buttons to handlers**

Update toolbar buttons in App.tsx:

```tsx
<div className="toolbar">
  <button onClick={handleStep}>Step</button>
  <button onClick={handleStepOver}>Step Over</button>
  <button onClick={handleStepOut}>Step Out</button>
  <button onClick={handleRun}>Run</button>
  <button onClick={handlePause}>Pause</button>
  <button onClick={handleReset}>Reset</button>
</div>
```

**Step 4: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 5: Commit toolbar functionality**

```bash
git add gui/frontend/src/App.tsx
git commit -m "feat(gui): connect toolbar buttons to debugger actions"
```

---

## Task 25: Create CommandInput Component

**Files:**
- Create: `gui/frontend/src/components/CommandInput.tsx`
- Create: `gui/frontend/src/components/CommandInput.css`

**Step 1: Create CommandInput component**

Create `gui/frontend/src/components/CommandInput.tsx`:

```tsx
import React, { useState, useRef, useEffect } from 'react';
import { ExecuteCommand } from '../../wailsjs/go/main/App';
import './CommandInput.css';

export const CommandInput: React.FC = () => {
  const [command, setCommand] = useState<string>('');
  const [history, setHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState<number>(-1);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!command.trim()) return;

    try {
      const output = await ExecuteCommand(command);
      console.log('Command output:', output);

      // Add to history
      setHistory(prev => [...prev, command]);
      setHistoryIndex(-1);
      setCommand('');
    } catch (error) {
      console.error('Command execution failed:', error);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (history.length > 0) {
        const newIndex = historyIndex === -1 ? history.length - 1 : Math.max(0, historyIndex - 1);
        setHistoryIndex(newIndex);
        setCommand(history[newIndex]);
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIndex >= 0) {
        const newIndex = historyIndex + 1;
        if (newIndex >= history.length) {
          setHistoryIndex(-1);
          setCommand('');
        } else {
          setHistoryIndex(newIndex);
          setCommand(history[newIndex]);
        }
      }
    }
  };

  return (
    <div className="command-input">
      <form onSubmit={handleSubmit} className="command-form">
        <span className="command-prompt">&gt;</span>
        <input
          ref={inputRef}
          type="text"
          className="command-field"
          value={command}
          onChange={e => setCommand(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Enter debugger command..."
          autoComplete="off"
        />
      </form>
    </div>
  );
};
```

**Step 2: Create CommandInput CSS**

Create `gui/frontend/src/components/CommandInput.css`:

```css
.command-input {
  display: flex;
  align-items: center;
  background: #1e1e1e;
  border-top: 1px solid #3e3e3e;
  padding: 8px 12px;
  font-family: 'Consolas', 'Monaco', monospace;
}

.command-form {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 8px;
}

.command-prompt {
  color: #4ec9b0;
  font-weight: bold;
  font-size: 14px;
}

.command-field {
  flex: 1;
  background: #252525;
  color: #d4d4d4;
  border: 1px solid #3e3e3e;
  padding: 6px 8px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  border-radius: 3px;
}

.command-field:focus {
  outline: none;
  border-color: #4ec9b0;
}

.command-field::placeholder {
  color: #858585;
}
```

**Step 3: Add CommandInput to App layout**

Modify `gui/frontend/src/App.tsx` - add import:

```tsx
import { CommandInput } from './components/CommandInput';
```

Add at bottom of Allotment before closing vertical Allotment:

```tsx
        {/* Command Input - fixed height at very bottom */}
        <Allotment.Pane snap minSize={40} maxSize={40}>
          <CommandInput />
        </Allotment.Pane>
      </Allotment>
    </div>
  );
```

**Step 4: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 5: Commit CommandInput component**

```bash
git add gui/frontend/src/components/CommandInput.tsx gui/frontend/src/components/CommandInput.css gui/frontend/src/App.tsx
git commit -m "feat(gui): add CommandInput component with command history"
```

---

## Task 22: Add Memory Read/Write Tracking API

**Files:**
- Modify: `service/types.go`
- Modify: `service/debugger_service.go`
- Test: `tests/unit/service/debugger_service_test.go`

**Step 1: Add MemoryData type**

Add to `service/types.go`:

```go
// MemoryData represents a chunk of memory with metadata
type MemoryData struct {
	Address      uint32   `json:"address"`       // Start address
	Data         []byte   `json:"data"`          // Raw bytes
	RecentWrites []uint32 `json:"recent_writes"` // Addresses written since last step
}
```

**Step 2: Write failing test for GetMemoryData**

Add to `tests/unit/service/debugger_service_test.go`:

```go
func TestDebuggerService_GetMemoryData(t *testing.T) {
	s := service.NewDebuggerService()

	program := `
main:
    MOV R0, #0x10000
    MOV R1, #0x42
    STR R1, [R0]
    SWI #0x00
`
	err := s.LoadProgram([]byte(program), []string{})
	if err != nil {
		t.Fatalf("Failed to load program: %v", err)
	}

	// Execute first 3 instructions (STR will write to memory)
	s.Step() // MOV R0
	s.Step() // MOV R1
	s.Step() // STR R1, [R0]

	// Get memory data at 0x10000
	memData := s.GetMemoryData(0x10000, 256)

	if len(memData.Data) != 256 {
		t.Errorf("Expected 256 bytes, got %d", len(memData.Data))
	}

	// Check that write was tracked
	if len(memData.RecentWrites) == 0 {
		t.Error("Expected recent writes to be tracked")
	}

	// Verify write address is in recent writes
	foundWrite := false
	for _, addr := range memData.RecentWrites {
		if addr >= 0x10000 && addr <= 0x10003 {
			foundWrite = true
			break
		}
	}

	if !foundWrite {
		t.Error("Expected write at 0x10000 to be in recent writes")
	}
}
```

**Step 3: Run test to verify it fails**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetMemoryData`

Expected: FAIL with "s.GetMemoryData undefined"

**Step 4: Add memory tracking fields to DebuggerService**

Modify `service/debugger_service.go` - add to struct:

```go
type DebuggerService struct {
	debugger            *debugger.Debugger
	mu                  sync.Mutex
	outputWriter        *EventEmittingWriter
	ctx                 context.Context
	recentWrites        map[uint32]bool // Memory addresses written in last step
	lastTraceEntryCount int             // Number of trace entries before last step
}
```

Update `NewDebuggerService`:

```go
func NewDebuggerService() *DebuggerService {
	return &DebuggerService{
		recentWrites: make(map[uint32]bool),
	}
}
```

**Step 5: Implement GetMemoryData method**

Add to `service/debugger_service.go`:

```go
// GetMemoryData returns a chunk of memory with recent write tracking
func (s *DebuggerService) GetMemoryData(startAddr uint32, length int) MemoryData {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil || s.debugger.VM == nil {
		return MemoryData{
			Address:      startAddr,
			Data:         []byte{},
			RecentWrites: []uint32{},
		}
	}

	// Read memory data
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		addr := startAddr + uint32(i)
		b, err := s.debugger.VM.Memory.ReadByteAt(addr)
		if err != nil {
			data[i] = 0
		} else {
			data[i] = b
		}
	}

	// Collect recent write addresses
	recentWrites := make([]uint32, 0, len(s.recentWrites))
	for addr := range s.recentWrites {
		if addr >= startAddr && addr < startAddr+uint32(length) {
			recentWrites = append(recentWrites, addr)
		}
	}

	return MemoryData{
		Address:      startAddr,
		Data:         data,
		RecentWrites: recentWrites,
	}
}
```

**Step 6: Implement memory write detection**

Add to `service/debugger_service.go`:

```go
// detectMemoryWrites tracks memory writes from the last step
func (s *DebuggerService) detectMemoryWrites() {
	// Clear previous writes
	s.recentWrites = make(map[uint32]bool)

	// Check memory trace for new writes
	if s.debugger.VM.MemoryTrace != nil && s.debugger.VM.MemoryTrace.Enabled {
		entries := s.debugger.VM.MemoryTrace.GetEntries()

		// Only look at new entries since last step
		for i := s.lastTraceEntryCount; i < len(entries); i++ {
			if entries[i].Type == "WRITE" {
				addr := entries[i].Address
				// Mark 4 bytes (word write)
				s.recentWrites[addr] = true
				s.recentWrites[addr+1] = true
				s.recentWrites[addr+2] = true
				s.recentWrites[addr+3] = true
			}
		}

		s.lastTraceEntryCount = len(entries)
	}
}
```

**Step 7: Call detectMemoryWrites in Step method**

Modify existing `Step` method in `service/debugger_service.go`:

```go
func (s *DebuggerService) Step() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.debugger == nil {
		return fmt.Errorf("no program loaded")
	}

	err := s.debugger.CmdStep()
	if err == nil {
		s.detectMemoryWrites()
	}
	return err
}
```

**Step 8: Run test to verify it passes**

Run: `go clean -testcache && go test ./tests/unit/service -v -run TestDebuggerService_GetMemoryData`

Expected: PASS

**Step 9: Commit memory tracking API**

```bash
git add service/types.go service/debugger_service.go tests/unit/service/debugger_service_test.go
git commit -m "feat(service): add GetMemoryData API with write tracking"
```

---

## Task 23: Create Advanced MemoryView Component

**Files:**
- Create: `gui/frontend/src/components/MemoryView.tsx`
- Create: `gui/frontend/src/components/MemoryView.css`

**Step 1: Create MemoryView component with scrolling and highlighting**

Create `gui/frontend/src/components/MemoryView.tsx`:

```tsx
import React, { useEffect, useState, useRef } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { GetMemoryData, GetRegisterState } from '../../wailsjs/go/main/App';
import { service } from '../../wailsjs/go/models';
import './MemoryView.css';

interface MemoryRow {
  address: number;
  bytes: number[];
  ascii: string;
  hasRecentWrite: boolean[];
}

export const MemoryView: React.FC = () => {
  const [rows, setRows] = useState<MemoryRow[]>([]);
  const [baseAddress, setBaseAddress] = useState<number>(0x00008000);
  const [addressInput, setAddressInput] = useState<string>('00008000');
  const [bytesPerRow] = useState<number>(16);
  const [rowCount] = useState<number>(32); // Show 32 rows = 512 bytes
  const containerRef = useRef<HTMLDivElement>(null);

  const loadMemoryData = async () => {
    try {
      const memData: service.MemoryData = await GetMemoryData(
        baseAddress,
        bytesPerRow * rowCount
      );

      // Build set of recently written addresses for fast lookup
      const recentWriteSet = new Set(memData.recent_writes || []);

      // Convert flat byte array to rows
      const newRows: MemoryRow[] = [];
      for (let row = 0; row < rowCount; row++) {
        const rowAddr = baseAddress + row * bytesPerRow;
        const rowBytes: number[] = [];
        const hasWrite: boolean[] = [];
        let ascii = '';

        for (let col = 0; col < bytesPerRow; col++) {
          const idx = row * bytesPerRow + col;
          const byte = idx < memData.data.length ? memData.data[idx] : 0;
          const byteAddr = rowAddr + col;

          rowBytes.push(byte);
          hasWrite.push(recentWriteSet.has(byteAddr));

          // ASCII representation (printable chars only)
          if (byte >= 32 && byte <= 126) {
            ascii += String.fromCharCode(byte);
          } else {
            ascii += '.';
          }
        }

        newRows.push({
          address: rowAddr,
          bytes: rowBytes,
          ascii,
          hasRecentWrite: hasWrite,
        });
      }

      setRows(newRows);
    } catch (error) {
      console.error('Failed to load memory data:', error);
    }
  };

  const handleAddressChange = (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const addr = parseInt(addressInput, 16);
      if (!isNaN(addr)) {
        setBaseAddress(addr);
      }
    } catch (error) {
      console.error('Invalid address:', error);
    }
  };

  const handleScroll = (delta: number) => {
    const newAddr = Math.max(0, baseAddress + delta * bytesPerRow);
    setBaseAddress(newAddr);
    setAddressInput(newAddr.toString(16).padStart(8, '0'));
  };

  const jumpToPC = async () => {
    try {
      const regs = await GetRegisterState();
      const pc = regs.pc;
      setBaseAddress(pc);
      setAddressInput(pc.toString(16).padStart(8, '0'));
    } catch (error) {
      console.error('Failed to jump to PC:', error);
    }
  };

  useEffect(() => {
    loadMemoryData();
    EventsOn('vm:state-changed', loadMemoryData);

    return () => {
      EventsOff('vm:state-changed');
    };
  }, [baseAddress]);

  return (
    <div className="memory-view">
      <div className="memory-header">
        <span>Memory View</span>
        <div className="memory-controls">
          <form onSubmit={handleAddressChange} className="address-form">
            <span className="address-label">Addr:</span>
            <input
              type="text"
              className="address-input"
              value={addressInput}
              onChange={e => setAddressInput(e.target.value)}
              placeholder="00008000"
              maxLength={8}
            />
            <button type="submit" className="btn-go">Go</button>
          </form>
          <button className="btn-jump" onClick={jumpToPC}>Jump to PC</button>
          <div className="scroll-buttons">
            <button className="btn-scroll" onClick={() => handleScroll(-16)}>
              ▲▲
            </button>
            <button className="btn-scroll" onClick={() => handleScroll(-1)}>
              ▲
            </button>
            <button className="btn-scroll" onClick={() => handleScroll(1)}>
              ▼
            </button>
            <button className="btn-scroll" onClick={() => handleScroll(16)}>
              ▼▼
            </button>
          </div>
        </div>
      </div>
      <div className="memory-content" ref={containerRef}>
        <div className="memory-grid">
          {/* Header row */}
          <div className="memory-row memory-row-header">
            <span className="memory-address">Address</span>
            <span className="memory-hex-header">
              {Array.from({ length: bytesPerRow }, (_, i) => (
                <span key={i} className="hex-col-label">
                  {i.toString(16).toUpperCase()}
                </span>
              ))}
            </span>
            <span className="memory-ascii-header">ASCII</span>
          </div>

          {/* Data rows */}
          {rows.map((row, rowIdx) => (
            <div key={rowIdx} className="memory-row">
              <span className="memory-address">
                {row.address.toString(16).padStart(8, '0').toUpperCase()}
              </span>
              <span className="memory-hex">
                {row.bytes.map((byte, colIdx) => (
                  <span
                    key={colIdx}
                    className={`hex-byte ${
                      row.hasRecentWrite[colIdx] ? 'hex-byte-changed' : ''
                    }`}
                  >
                    {byte.toString(16).padStart(2, '0').toUpperCase()}
                  </span>
                ))}
              </span>
              <span className="memory-ascii">{row.ascii}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
```

**Step 2: Create MemoryView CSS with yellow highlighting**

Create `gui/frontend/src/components/MemoryView.css`:

```css
.memory-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.memory-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  font-weight: bold;
  font-size: 12px;
}

.memory-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.address-form {
  display: flex;
  align-items: center;
  gap: 6px;
}

.address-label {
  font-size: 11px;
  color: #858585;
}

.address-input {
  width: 80px;
  background: #1e1e1e;
  color: #9cdcfe;
  border: 1px solid #3e3e3e;
  padding: 4px 6px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 11px;
  text-transform: uppercase;
  border-radius: 3px;
}

.address-input:focus {
  outline: none;
  border-color: #4ec9b0;
}

.btn-go,
.btn-jump,
.btn-scroll {
  background: #3e3e3e;
  color: #d4d4d4;
  border: none;
  padding: 4px 8px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 10px;
  font-weight: 500;
}

.btn-go:hover,
.btn-jump:hover,
.btn-scroll:hover {
  background: #4e4e4e;
}

.scroll-buttons {
  display: flex;
  gap: 2px;
}

.btn-scroll {
  padding: 2px 6px;
  font-size: 9px;
  line-height: 1;
}

.memory-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  background: #1e1e1e;
  padding: 4px;
}

.memory-grid {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.memory-row {
  display: flex;
  gap: 12px;
  padding: 2px 4px;
  font-size: 12px;
  line-height: 1.6;
  align-items: center;
}

.memory-row-header {
  position: sticky;
  top: 0;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  font-weight: bold;
  font-size: 10px;
  color: #858585;
  z-index: 1;
  padding: 4px;
}

.memory-address {
  min-width: 80px;
  color: #858585;
  font-weight: normal;
  user-select: none;
}

.memory-hex-header {
  display: flex;
  gap: 8px;
  font-family: 'Consolas', 'Monaco', monospace;
  letter-spacing: 0.5px;
}

.hex-col-label {
  display: inline-block;
  width: 20px;
  text-align: center;
  font-size: 9px;
}

.memory-hex {
  display: flex;
  gap: 8px;
  font-family: 'Consolas', 'Monaco', monospace;
  letter-spacing: 0.5px;
}

.hex-byte {
  display: inline-block;
  width: 20px;
  text-align: center;
  color: #9cdcfe;
  transition: background-color 0.3s ease, color 0.3s ease;
}

.hex-byte-changed {
  background-color: #ffeb3b;
  color: #000000;
  font-weight: bold;
  border-radius: 2px;
  animation: highlight-fade 2s ease-out;
}

@keyframes highlight-fade {
  0% {
    background-color: #ffeb3b;
  }
  100% {
    background-color: #b8860b;
  }
}

.memory-ascii-header {
  min-width: 140px;
  padding-left: 8px;
}

.memory-ascii {
  min-width: 140px;
  color: #ce9178;
  font-family: 'Consolas', 'Monaco', monospace;
  letter-spacing: 1px;
  padding-left: 8px;
}

/* Scrollbar styling for webkit browsers */
.memory-content::-webkit-scrollbar {
  width: 8px;
}

.memory-content::-webkit-scrollbar-track {
  background: #1e1e1e;
}

.memory-content::-webkit-scrollbar-thumb {
  background: #3e3e3e;
  border-radius: 4px;
}

.memory-content::-webkit-scrollbar-thumb:hover {
  background: #4e4e4e;
}
```

**Step 3: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 4: Commit MemoryView component**

```bash
git add gui/frontend/src/components/MemoryView.tsx gui/frontend/src/components/MemoryView.css
git commit -m "feat(gui): add advanced MemoryView component with scrolling and write highlighting"
```

---

## Task 24: Create ExpressionEvaluator Component

**Files:**
- Create: `gui/frontend/src/components/ExpressionEvaluator.tsx`
- Create: `gui/frontend/src/components/ExpressionEvaluator.css`

**Step 1: Create ExpressionEvaluator component**

Create `gui/frontend/src/components/ExpressionEvaluator.tsx`:

```tsx
import React, { useState } from 'react';
import { EvaluateExpression } from '../../wailsjs/go/main/App';
import './ExpressionEvaluator.css';

interface EvaluationResult {
  expression: string;
  result: string;
  error?: string;
}

export const ExpressionEvaluator: React.FC = () => {
  const [expression, setExpression] = useState<string>('');
  const [results, setResults] = useState<EvaluationResult[]>([]);

  const handleEvaluate = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!expression.trim()) return;

    try {
      const result = await EvaluateExpression(expression);
      setResults(prev => [
        ...prev,
        {
          expression,
          result: `0x${result.toString(16).padStart(8, '0')} (${result})`,
        }
      ].slice(-10)); // Keep last 10 results

      setExpression('');
    } catch (error) {
      setResults(prev => [
        ...prev,
        {
          expression,
          result: '',
          error: String(error),
        }
      ].slice(-10));

      setExpression('');
    }
  };

  return (
    <div className="expression-evaluator">
      <div className="evaluator-header">Expression Evaluator</div>

      <form onSubmit={handleEvaluate} className="evaluator-form">
        <input
          type="text"
          className="evaluator-input"
          value={expression}
          onChange={e => setExpression(e.target.value)}
          placeholder="Enter expression (e.g., R0 + R1)"
        />
        <button type="submit" className="evaluator-button">Evaluate</button>
      </form>

      <div className="evaluator-results">
        {results.map((result, index) => (
          <div key={index} className={`result-entry ${result.error ? 'result-error' : ''}`}>
            <div className="result-expression">{result.expression}</div>
            <div className="result-value">
              {result.error ? `Error: ${result.error}` : result.result}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
```

**Step 2: Create ExpressionEvaluator CSS**

Create `gui/frontend/src/components/ExpressionEvaluator.css`:

```css
.expression-evaluator {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', monospace;
}

.evaluator-header {
  padding: 8px 12px;
  background: #252525;
  border-bottom: 1px solid #3e3e3e;
  font-weight: bold;
  font-size: 12px;
}

.evaluator-form {
  display: flex;
  gap: 8px;
  padding: 8px;
  border-bottom: 1px solid #3e3e3e;
}

.evaluator-input {
  flex: 1;
  background: #252525;
  color: #d4d4d4;
  border: 1px solid #3e3e3e;
  padding: 6px 8px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  border-radius: 3px;
}

.evaluator-input:focus {
  outline: none;
  border-color: #4ec9b0;
}

.evaluator-button {
  background: #3e3e3e;
  color: #d4d4d4;
  border: none;
  padding: 6px 12px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 11px;
}

.evaluator-button:hover {
  background: #4e4e4e;
}

.evaluator-results {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.result-entry {
  padding: 6px 8px;
  margin-bottom: 4px;
  background: #252525;
  border-left: 3px solid #4ec9b0;
  border-radius: 3px;
}

.result-error {
  border-left-color: #e51400;
}

.result-expression {
  font-size: 12px;
  color: #9cdcfe;
  margin-bottom: 2px;
}

.result-value {
  font-size: 11px;
  color: #d4d4d4;
}

.result-error .result-value {
  color: #e51400;
}
```

**Step 3: Add ExpressionEvaluator to App layout**

Modify `gui/frontend/src/App.tsx` - add import:

```tsx
import { ExpressionEvaluator } from './components/ExpressionEvaluator';
```

Add to right panel vertical Allotment:

```tsx
            <Allotment.Pane minSize={300} preferredSize={400}>
              <Allotment vertical>
                <Allotment.Pane>
                  <div className="placeholder-view">Register View (existing)</div>
                </Allotment.Pane>
                <Allotment.Pane>
                  <div className="placeholder-view">Memory View (existing)</div>
                </Allotment.Pane>
                <Allotment.Pane>
                  <StackView />
                </Allotment.Pane>
                <Allotment.Pane>
                  <ExpressionEvaluator />
                </Allotment.Pane>
              </Allotment>
            </Allotment.Pane>
```

**Step 4: Test build**

Run: `cd gui/frontend && npm run build && cd ../..`

Expected: Builds successfully

**Step 5: Commit ExpressionEvaluator component**

```bash
git add gui/frontend/src/components/ExpressionEvaluator.tsx gui/frontend/src/components/ExpressionEvaluator.css gui/frontend/src/App.tsx
git commit -m "feat(gui): add ExpressionEvaluator component for debugging expressions"
```

---

## Task 25: Final Testing and Documentation

**Files:**
- Test: Manual testing with example programs
- Update: `docs/plans/2025-10-28-gui-debugging-views-design.md`

**Step 1: Build complete application**

Run: `go build -o arm-emulator`

Expected: Builds successfully

**Step 2: Test with hello.s example**

Run: `./arm-emulator --gui examples/hello.s`

Expected: GUI opens with new layout and components visible

**Step 3: Test stepping functionality**

1. Click Step button
2. Verify SourceView highlights current line
3. Verify DisassemblyView shows current instruction
4. Verify RegisterView updates (if implemented)
5. Verify StatusView shows state changes

**Step 4: Test breakpoint functionality**

1. Click on a source line to set breakpoint
2. Verify breakpoint marker appears
3. Click Run button
4. Verify execution stops at breakpoint
5. Verify BreakpointsView shows breakpoint

**Step 5: Test memory view functionality**

Run program that writes to memory

1. Step through program
2. Verify MemoryView shows hex dump
3. Verify bytes written are highlighted in yellow
4. Type new address in address input and click Go
5. Verify memory view scrolls to new address
6. Click scroll buttons to navigate
7. Click "Jump to PC" button
8. Verify memory view shows PC location

**Step 6: Test output capture**

Run program that produces output (hello.s)

1. Click Run
2. Verify OutputView shows "Hello, World!"
3. Click Clear button
4. Verify output cleared

**Step 7: Test command input**

1. Type "info registers" in CommandInput
2. Press Enter
3. Verify command executes (check browser console for output)

**Step 8: Test expression evaluator**

1. Step through program to set register values
2. Enter "R0 + R1" in ExpressionEvaluator
3. Click Evaluate
4. Verify result shows correct value

**Step 9: Update design document status**

Modify `docs/plans/2025-10-28-gui-debugging-views-design.md`:

Change:
```markdown
**Status:** In Design
```

To:
```markdown
**Status:** Implemented
**Implementation Date:** 2025-10-28
```

**Step 9: Run full test suite**

Run: `go clean -testcache && go test ./...`

Expected: All tests pass

**Step 10: Run linting**

Run: `golangci-lint run ./...`

Expected: No issues

**Step 11: Final commit**

```bash
git add docs/plans/2025-10-28-gui-debugging-views-design.md
git commit -m "docs: mark GUI debugging views as implemented"
```

---

## Success Criteria

All tasks completed when:

- ✓ All backend service APIs implemented and tested
- ✓ Memory write tracking implemented with trace integration
- ✓ GetMemoryData API returns memory with recent writes
- ✓ Event-emitting writer functional with real-time output
- ✓ Wails events emitted for all state changes
- ✓ All frontend components created and styled
- ✓ Components subscribe to events and auto-update
- ✓ Allotment layout integrated with resizable panels
- ✓ Breakpoints can be toggled from SourceView
- ✓ DisassemblyView shows machine code with symbols
- ✓ MemoryView displays scrollable hex dump with yellow highlighting for recent writes
- ✓ MemoryView supports address navigation and jump to PC
- ✓ MemoryView shows ASCII representation alongside hex
- ✓ StackView displays stack with SP marker
- ✓ OutputView captures and displays program output
- ✓ StatusView shows execution state and messages
- ✓ BreakpointsView lists breakpoints and watchpoints
- ✓ CommandInput accepts debugger commands with history
- ✓ ExpressionEvaluator parses and evaluates expressions
- ✓ Toolbar buttons functional (Step, Step Over, Step Out, Run, Pause, Reset)
- ✓ All tests pass (go test ./...)
- ✓ No lint issues (golangci-lint run)
- ✓ GUI builds and runs successfully
- ✓ Manual testing with example programs successful

---

## Execution Complete

**Plan saved to:** `docs/plans/2025-10-28-gui-debugging-views-implementation.md`

This plan provides bite-sized tasks (2-5 minutes each) following TDD principles with exact file paths, complete code examples, and verification steps. Each task includes test-first development, incremental commits, and clear success criteria.
