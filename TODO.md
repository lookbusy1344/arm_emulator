# ARM2 Emulator TODO List

**Last Updated:** 2025-10-17

This file tracks outstanding work only. Completed items are in `PROGRESS.md`.

---

## Summary

**Status:** Project is production-ready with comprehensive test coverage and all critical features implemented.

**Test Status:** 969 tests, 100% pass rate, 0 lint issues, 75.0% code coverage

---

## High Priority Tasks

None

---

## Medium Priority Tasks

### TUI Memory Write Highlighting Not Visible
**Status:** INVESTIGATING - Debug instrumentation added
**Priority:** MEDIUM

**Problem:**
When a STR (store) instruction executes in the TUI debugger, the stored memory location should be highlighted in green in the Memory window, but the highlighting is not visible to the user.

**Expected Behavior:**
After executing a STR instruction (e.g., `STR R1, [R0]`), the 4 bytes written to memory should appear in green in the Memory window for one step, making it easy to see what memory was modified.

**What Works:**
- MemoryTrace is enabled and recording writes (confirmed via debug output: "RecentWrites=4, MemTrace=true, Entries=3")
- `DetectMemoryWrites()` is detecting writes and populating the `RecentWrites` map with 4 addresses
- Stack highlighting works (green highlighting appears in Stack window for PUSH operations)
- Register highlighting works (changed registers appear in green)

**What Doesn't Work:**
- Memory window doesn't show green highlighting for written bytes
- User reports no visible green color after stepping past STR instructions

**Investigation Steps Completed:**
1. ✅ Fixed stack pointer preservation bug (`ResetRegisters()` now restores SP from `vm.StackTop`)
2. ✅ Implemented MemoryTrace tracking (enabled automatically in TUI mode)
3. ✅ Added `CaptureMemoryTraceState()` and `DetectMemoryWrites()` functions
4. ✅ Fixed source view truncation (square brackets escaped with `tview.Escape()`)
5. ✅ Fixed memory hex rendering (manually build hex string instead of using `strings.Join()`)
6. ✅ Added `DetectMemoryWrites()` calls in breakpoint hit paths
7. ✅ Confirmed MemoryTrace.RecordWrite() is called by STR instruction execution
8. ✅ Added debug instrumentation to trace RecentWrites map population

**Current Debug Instrumentation:**
- `DetectMemoryWrites()`: Reports LastCount, CurrentCount, NewWrites, RecentWrites count, FirstAddr
- `UpdateMemoryView()`: Reports RecentWrites map size

**Technical Details:**
- File: `debugger/tui.go`
- Key functions: `DetectMemoryWrites()` (line 966), `UpdateMemoryView()` (line 588)
- Memory highlighting logic: lines 630-634 (checks `t.RecentWrites[byteAddr]` and renders `[green]XX[white]`)
- MemoryTrace recording: `vm/inst_memory.go` line 147

**Code Paths to Check:**
1. Run mode breakpoint hit: lines 307-318 (now calls DetectMemoryWrites)
2. Step mode breakpoint hit: lines 349-360 (now calls DetectMemoryWrites)
3. Normal step: lines 345-346 (calls DetectMemoryWrites)
4. Memory view rendering: lines 630-634 (applies green color)

**Next Steps:**
1. Run test with debug output: `./arm-emulator --tui examples/test_store_highlight.s`
2. Execute: `b str_test1`, `r`, `s`
3. Analyze debug output to determine:
   - Is `DetectMemoryWrites()` finding the writes?
   - Is `UpdateMemoryView()` seeing the populated RecentWrites map?
   - Is the rendering logic being executed?
4. Possible root causes still to investigate:
   - tview color tag rendering issue in MemoryView
   - QueueUpdateDraw timing causing RecentWrites to be stale
   - Memory window scrolling/focus preventing visibility
   - Color scheme or terminal compatibility issue

**Test Program:**
`examples/test_store_highlight.s` - Contains labeled STR instructions for easy breakpoint testing

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
