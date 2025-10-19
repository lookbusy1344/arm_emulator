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

### GUI Debugger Enhancements
**Status:** IN PROGRESS - Core GUI implemented, enhancements needed
**Priority:** MEDIUM

**Current State:**
- ✅ Basic GUI debugger implemented with Fyne framework (`--gui` flag)
- ✅ Core panels: Source, Registers, Memory, Stack, Breakpoints, Console
- ✅ Control toolbar: Run, Step, Continue, Stop, Add/Clear Breakpoints
- ✅ Initial automated tests in `debugger/gui_test.go`
- ✅ Comprehensive documentation (`docs/gui_assessment.md`, `docs/gui_debugger_guide.md`, `docs/gui_testing.md`)

**CRITICAL REQUIREMENT:** **Full automated testing is VITAL for every GUI feature.** All new GUI functionality must include comprehensive automated tests using Fyne's `fyne.io/fyne/v2/test` package.

**Planned Enhancements (Each requires automated tests):**

**Phase 1: Core Improvements** (8-12 hours)
- [ ] **Syntax highlighting in source view** - Color-code ARM instructions, registers, labels, comments
  - **Testing:** Automated tests verifying color tags are applied correctly
- [ ] **Click-to-set breakpoints** - Click in source view to toggle breakpoints
  - **Testing:** Automated tests using `test.Tap()` to simulate clicks and verify breakpoint addition/removal
- [ ] **Keyboard shortcuts** - F5 (Run), F9 (Toggle BP), F10 (Step Over), F11 (Step Into), Ctrl+R (Refresh)
  - **Testing:** Automated tests using key event simulation to verify shortcuts work
- [ ] **Register change highlighting** - Color-code registers that changed in last step (green)
  - **Testing:** Automated tests verifying highlight state after step operations

**Phase 2: Advanced Features** (10-15 hours)
- [ ] **Memory editing** - Click to edit memory bytes in hex view
  - **Testing:** Automated tests for edit operations and value validation
- [ ] **Memory search** - Find byte patterns or strings in memory
  - **Testing:** Automated tests for search functionality and result navigation
- [ ] **Watch expressions** - Add custom expressions to watch panel
  - **Testing:** Automated tests for expression evaluation and display updates
- [ ] **Goto address** - Jump to specific memory/stack address
  - **Testing:** Automated tests verifying address navigation

**Phase 3: Polish & UX** (6-8 hours)
- [ ] **Dark/Light theme toggle** - Switch between color schemes
  - **Testing:** Automated tests verifying theme changes apply to all panels
- [ ] **Window state persistence** - Save/restore window size, position, panel sizes
  - **Testing:** Automated tests for state serialization/deserialization
- [ ] **Preferences dialog** - Configure GUI settings (font size, colors, etc.)
  - **Testing:** Automated tests for preference changes and persistence
- [ ] **Improved source view** - Better handling of long files, line wrapping options
  - **Testing:** Automated tests for scrolling, wrapping, navigation

**Phase 4: Advanced Debugging** (12-16 hours)
- [ ] **Conditional breakpoints UI** - Set conditions when creating breakpoints
  - **Testing:** Automated tests for condition parsing and evaluation
- [ ] **Breakpoint hit counts** - Track how many times breakpoint was hit
  - **Testing:** Automated tests verifying hit count tracking
- [ ] **Mixed source/disassembly view** - Show assembly alongside source
  - **Testing:** Automated tests for view synchronization
- [ ] **Instruction tooltips** - Hover to see instruction documentation
  - **Testing:** Automated tests using mouse hover simulation

**Testing Requirements:**
- **Every feature MUST have automated tests** before being merged
- Use `fyne.io/fyne/v2/test` package for headless testing
- Tests must work in CI/CD without display server
- Minimum test coverage: 80% for all GUI code
- Include both unit tests (individual functions) and integration tests (complete workflows)
- Visual regression tests for UI changes (compare screenshots)

**Example Test Pattern:**
```go
func TestFeature(t *testing.T) {
    app := test.NewApp()  // Headless test app
    defer app.Quit()
    
    gui := newGUI(debugger)
    
    // Test setup
    // Perform action
    // Verify result
}
```

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
