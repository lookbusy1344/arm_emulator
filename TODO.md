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
