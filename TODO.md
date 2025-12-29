# ARM2 Emulator TODO List

**Last Updated:** 2025-12-29

This file tracks outstanding work only. Completed items are in `PROGRESS.md`.

---

## Summary

**Status:** Project is production-ready with comprehensive test coverage and all critical features implemented.

**Test Status:** 1,024 tests, 100% pass rate, 0 lint issues, 75.0% code coverage
**Code Size:** 46,257 lines of Go code
**Example Programs:** 49 programs, all fully functional (100% success rate)

---

## Medium Priority Tasks

### Additional Diagnostic Modes

**Proposed Extensions:**
- [ ] **Data Flow Tracing** (6-8 hours) - Track data movement between registers/memory, value provenance, data dependency tracking, taint analysis
- [ ] **Cycle-Accurate Timing Simulation** (8-10 hours) - Estimate ARM2 instruction timing, pipeline stall simulation, memory access latency, performance bottleneck identification
- [ ] **Memory Region Heatmap Visualization** (4-6 hours) - Track access frequency per region, HTML/graphical output, color-coded visualization
- [ ] **Reverse Execution Log** (10-12 hours) - Record state for backwards stepping, circular buffer of previous N instructions, time-travel debugging

---

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

---

### Enhanced CI/CD Pipeline (Optional)
**Effort:** 2-3 hours (partially complete)

**Remaining:**
- [ ] Add test coverage reporting (codecov integration)
- [ ] Add coverage threshold enforcement in CI (currently 75% local)
- [ ] Add race detector to CI pipeline (works locally but not in ci.yml)

---

### GUI E2E Test Quality Improvements (Optional)

**Remaining Work (UI Features Not Implemented):**
- [ ] Implement theme toggle UI (2 skipped tests in visual.spec.ts)
- [ ] Implement breakpoint enable/disable checkbox (1 skipped test in breakpoints.spec.ts)
- [ ] Implement clear-all-breakpoints button (1 skipped test in breakpoints.spec.ts)
- [ ] Scroll test for memory view (1 skipped test - memory view is virtualized)

**Test Quality Improvements:**
- [ ] **Error message verification in error-scenarios.spec.ts** - Currently tests only check errors exist (`toBeTruthy()`), not actual error message content
- [ ] **Remove hardcoded waits from visual.spec.ts** - 5 `waitForTimeout()` calls should be replaced with proper state checks
- [ ] **Remove hardcoded waits from memory.spec.ts** - 2 `waitForTimeout(200)` calls should use state-based assertions
- [ ] **Remove hardcoded waits from breakpoints.spec.ts** - 3 `waitForTimeout()` calls should use `waitForFunction()`
- [ ] **Remove hardcoded waits from execution.spec.ts** - 12 `waitForTimeout()` calls should be replaced with proper state checks
