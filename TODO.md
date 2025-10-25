# ARM2 Emulator TODO List

**Last Updated:** 2025-10-23

This file tracks outstanding work only. Completed items are in `PROGRESS.md`.

---

## Summary

**Status:** Project is production-ready with comprehensive test coverage and all critical features implemented.

**Test Status:** 1,024 tests, 100% pass rate, 0 lint issues, 75.0% code coverage
**Code Size:** 46,257 lines of Go code
**Example Programs:** 49 programs, all fully functional (100% success rate)

---

## High Priority Tasks

None

---

## Medium Priority Tasks

### String Building Performance in Trace Output
**Priority:** MEDIUM
**Effort:** 2-3 hours
**Type:** Performance Optimization

**Problem:**
Multiple trace output files use string concatenation with `+=` in loops:
- `vm/flag_trace.go` - Statistics output formatting
- `vm/register_trace.go` - Report generation
- `vm/statistics.go` - HTML/text output building

Each concatenation allocates a new string and copies previous content, creating O(n²) performance characteristics for large trace outputs.

**Impact:** Trace file generation can be 10-50x slower than necessary for programs with 100,000+ instructions.

**Solution:** Replace string concatenation with `strings.Builder`:
```go
var sb strings.Builder
sb.WriteString("header\n")
sb.WriteString(fmt.Sprintf("value: %d\n", val))
output := sb.String()
```

**Files to Check:**
- `vm/flag_trace.go` (lines with fmt.Sprintf + +=)
- `vm/register_trace.go` (report generation)
- `vm/statistics.go` (HTML/text building)

---

### Memory Allocation Pressure in Hot Path (trace.go)
**Priority:** MEDIUM
**Effort:** 3-4 hours
**Type:** Performance Optimization

**Problem:**
`vm/trace.go` `RecordInstruction()` creates a new 16-entry map for every instruction:
```go
currentRegs := map[string]uint32{
    "R0": vm.CPU.R[0],
    "R1": vm.CPU.R[1],
    // ... repeated per instruction
}
```

With 1M+ instructions per program, this creates excessive garbage collection pressure.

**Impact:** Measurable GC overhead in long-running programs. Profile before/after optimization.

**Solution:** Reuse register snapshot structure or use array-based lookups instead of maps.

**Files:**
- `vm/trace.go` (RecordInstruction method, lines 99-119)

---

### Duplicate Register State Tracking
**Priority:** MEDIUM
**Effort:** 4-5 hours
**Type:** Refactoring/Code Quality

**Problem:**
Three independent systems track "last register state":
- `vm/trace.go` - `lastSnapshot` map
- `vm/register_trace.go` - `lastRegValues` map
- `debugger/tui.go` - `PrevRegisters` array

Code duplication creates maintenance burden and potential inconsistency in register change detection.

**Solution:** Extract into shared `RegisterSnapshot` type with methods:
- `ChangedRegs(other *RegisterSnapshot) []string`
- `Capture(cpu *CPU)`
- `GetRegister(name string) uint32`

Use consistently across all three locations.

**Files:**
- `vm/trace.go`
- `vm/register_trace.go`
- `debugger/tui.go`

---

### Missing Error Context in Encoder
**Priority:** MEDIUM
**Effort:** 3-4 hours
**Type:** Error Handling

**Problem:**
Encoder errors lack file:line information, making it hard to locate problematic instructions:
```go
return 0, fmt.Errorf("unknown instruction: %s", mnemonic)
// Missing: which file? which line? what was the actual instruction?
```

Users must manually search through assembly to find the error.

**Solution:** Create `EncodingError` type that includes instruction context:
```go
type EncodingError struct {
    Instruction *parser.Instruction
    Message     string
    Wrapped     error
}
```

Propagate source location information through encoder pipeline.

**Files:**
- `encoder/encoder.go`
- `encoder/data_processing.go`
- `encoder/memory.go`

---

### Syscall Error Handling Asymmetry
**Priority:** MEDIUM
**Effort:** 3-4 hours
**Type:** Error Handling

**Problem:**
Syscall documentation distinguishes between:
- "VM integrity errors" (should return Go error and halt)
- "Expected failures" (should return 0xFFFFFFFF in R0, continue)

However, this isn't consistently applied. Some handlers propagate Go errors for validation failures when they should return 0xFFFFFFFF.

**Example:** File operations like `handleOpen()` may panic or return errors instead of 0xFFFFFFFF on failure.

**Solution:** Create error classification system:
```go
type SyscallError struct {
    IsVMError bool // true = halt, false = return 0xFFFFFFFF
    Message   string
}
```

Audit all syscall handlers for consistency.

**Files:**
- `vm/syscall.go` (all handler functions)
- Create new file: `vm/syscall_error.go`

---

### TUI Help Command Display Issue
**Status:** BLOCKED - Needs Investigation
**Priority:** MEDIUM

**Problem:**
When typing `help` (or pressing F1) in the TUI debugger, the help text appears as black-on-black (invisible until highlighted). The text IS being written to the OutputView (confirmed via debug logging showing 1040 bytes), but is not visible.

**What Works:**
- Welcome message at TUI startup displays correctly with colors (green and white)
- All other TUI views display correctly
- Help command works fine in non-TUI debugger mode
- Color tags like `[green]`, `[white]` work in welcome message

**What Doesn't Work:**
- Help text written from within `QueueUpdateDraw` callback appears black-on-black
- Text written via `WriteOutput()` after TUI starts running is invisible

**Attempted Fixes (All Failed):**
1. Added `SetTextColor(tcell.ColorWhite)` to OutputView - no effect
2. Wrapped output in `[white]` color tags - no effect
3. Tried `[yellow]` tags to test if any colors work - no effect
4. Changed from `Write()` to `SetText()` - no effect
5. Used `GetText(true)` to preserve color tags - no effect
6. Tried `QueueUpdate` vs `QueueUpdateDraw` - no effect
7. Set explicit background color with `SetBackgroundColor(tcell.ColorBlack)` - no effect

**Technical Details:**
- File: `debugger/tui.go`
- Function: `executeCommand()` line 234
- OutputView config: line 136, has `SetDynamicColors(true)`
- Text written via: `go t.App.QueueUpdateDraw(func() { t.WriteOutput(...) })`

**Next Steps:**
- Investigate tview library documentation for `Write()` vs `SetText()` color handling
- Check if `QueueUpdateDraw` from goroutine has threading issues
- Test simpler approach without goroutines
- Consider filing tview library issue if bug is in library


### Additional Diagnostic Modes

**Proposed Extensions:**
- [ ] **Data Flow Tracing** (6-8 hours) - Track data movement between registers/memory, value provenance, data dependency tracking, taint analysis
- [ ] **Cycle-Accurate Timing Simulation** (8-10 hours) - Estimate ARM2 instruction timing, pipeline stall simulation, memory access latency, performance bottleneck identification
- [ ] **Memory Region Heatmap Visualization** (4-6 hours) - Track access frequency per region, HTML/graphical output, color-coded visualization
- [ ] **Reverse Execution Log** (10-12 hours) - Record state for backwards stepping, circular buffer of previous N instructions, time-travel debugging


### Performance & Benchmarking
**Effort:** 10-15 hours

- [ ] Create benchmark tests (VM, parser, TUI)
- [ ] Document performance targets
- [ ] Run CPU and memory profiling
- [ ] Create `docs/performance_analysis.md`
- [ ] Implement optimizations if needed

---

## Low Priority Tasks (Optional)

### Symbol Resolution Caching
**Priority:** LOW-MEDIUM
**Effort:** 2-3 hours
**Type:** Performance Optimization

**Problem:**
`ResolveAddress()` does binary search for every trace entry. With 100,000+ trace entries, this is 100,000 binary searches. The binary search is already efficient, but locality of reference is ignored.

**Solution:** Add simple cache for recently resolved symbols:
```go
type SymbolResolver struct {
    // existing fields...
    cacheAddr   uint32
    cacheName   string
    cacheOffset uint32
}

// Check cache (within 256 byte window) before binary search
```

Expected improvement: 5-15% speedup in trace output generation.

**Files:**
- `vm/symbol_resolver.go` (ResolveAddress method)

---

### RegisterTrace Memory Bounds
**Priority:** LOW
**Effort:** 2-3 hours
**Type:** Robustness

**Problem:**
`vm/register_trace.go` `RecordWrite()` can accumulate unlimited entries in `valuesSeen` map. In pathological cases with high register write variety, this could consume excessive memory.

**Solution:** Cap unique values tracking:
```go
const maxTrackedUniqueValues = 10000

if len(r.valuesSeen) < maxTrackedUniqueValues && !r.valuesSeen[value] {
    r.valuesSeen[value] = true
    r.UniqueValues++
}
```

Document the limit in output.

**Files:**
- `vm/register_trace.go` (RecordWrite method)

---

### TestAssert_MessageWraparound Test Investigation
**Status:** SKIPPED - Low Priority
**Priority:** LOW
**Effort:** 2-4 hours

**Current Behavior:**
The `TestAssert_MessageWraparound` test in `tests/unit/vm/security_fixes_test.go` is currently skipped. The ASSERT syscall does have wraparound protection and stops reading at address boundaries (reads 15 chars instead of wrapping), which is safe behavior. However, the test expects an explicit "wraparound" error message.

**Issue:**
When ASSERT reads a message starting at 0xFFFFFFF0 with no null terminator:
- Expects: Error containing "wraparound"
- Actual: Reads 15 characters (ABCDEFGHIJKLMNO) and returns "ASSERTION FAILED" error
- The 16th character at 0xFFFFFFFF is not being read

**Investigation Needed:**
- Determine why reading from address 0xFFFFFFFF fails or returns 0
- Verify memory segment boundary handling at maximum address
- Consider if current behavior (stopping at 15 chars) is acceptable
- Update test expectations or fix ASSERT handler accordingly

**Files:**
- `tests/unit/vm/security_fixes_test.go` (line 223)
- `vm/syscall.go` `handleAssert()` function (line 940)

### Later ARM Architecture Extensions (Optional)

These are **not** part of ARM2 but could be added for broader compatibility:

**ARMv2a Atomic Operations** (Effort: 4-6 hours)
- [ ] SWP (Swap Word) - Atomically swap 32-bit value between register and memory
- [ ] SWPB (Swap Byte) - Atomically swap 8-bit value between register and memory
- **Purpose:** Synchronization primitives for multi-threaded/multi-processor systems (spinlocks, semaphores, mutex)
- **Note:** Introduced in ARMv2a (ARM3), not original ARM2. ARM2 was single-processor without multi-threading support.

**ARMv2 Coprocessor Interface** (Effort: 20-30 hours)
- [ ] CDP, LDC, STC, MCR, MRC - Coprocessor operations
- [ ] Full coprocessor emulation framework

**Note:** The emulator has complete ARM2 instruction set coverage. All planned ARMv3/ARMv3M extensions have been completed. These remaining extensions are from later architectures.

### Enhanced CI/CD Pipeline (Optional)
**Effort:** 2-3 hours (partially complete)

**Completed:**
- ✅ Matrix builds for multiple platforms (build-release.yml: Linux AMD64, macOS ARM64, Windows AMD64/ARM64)
- ✅ Build artifact uploads with 30-day retention
- ✅ Race detector works locally (`go test -race ./...`)

**Remaining:**
- [ ] Add test coverage reporting (codecov integration)
- [ ] Add coverage threshold enforcement in CI (currently 75% local)
- [ ] Add race detector to CI pipeline (works locally but not in ci.yml)
