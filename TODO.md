# ARM2 Emulator TODO List

**IMPORTANT:** This file tracks outstanding tasks and known issues. After completing any work, update this file to reflect the current status.

Completed items and past work belong in `PROGRESS.md`.

**Last Updated:** 2025-10-11 (Phase 11 - Production Hardening)

---

## Summary

**Status:** All 10 core phases complete! Phase 11 (Production Hardening) in progress.

**Remaining Work:**
- **High Priority:** CI/CD enhancements, cross-platform testing, code coverage analysis
- **Medium Priority:** Release pipeline, performance benchmarking
- **Low Priority:** Additional documentation, advanced features

**Estimated effort to v1.0.0:** 30-45 hours

---

## Known Issues

No known critical issues at this time. All 531 tests pass (100%).

---

## Phase 11: Production Hardening

### Task 1: Fix Integer Conversion Issues ✅ COMPLETED

**Status:** ✅ All G115 warnings resolved (2025-10-11)

**Issues Found:** 4 integer conversions in test files flagged by gosec (G115 rule)
- All were safe loop index conversions (int → uint32)
- Loop indices are always non-negative and bounded

**Resolution:**
- Added `#nosec G115` directives with justification comments
- Verified all conversions are mathematically safe (loop indices [0, N))
- Added clear documentation explaining why each conversion is safe

**Files Fixed:**
- tests/unit/parser/character_literals_test.go (2 instances)
- tests/unit/vm/memory_system_test.go (2 instances)
- tests/unit/vm/syscall_test.go (1 instance)

**Verification:**
- ✅ All tests pass (531 tests, 100%)
- ✅ golangci-lint reports 0 issues
- ✅ No G115 warnings remain

**Note:** The original estimate of 56 instances appears to have been from an earlier scan. Current codebase only had 4 instances, all in test files.

**Effort:** 1 hour (much less than estimated 6-10 hours)

**Priority:** High (security/correctness issue) - COMPLETED

---

### Task 2: Enhanced CI/CD Pipeline

**Status:** Basic CI exists with Go 1.25

**Requirements:**
- [ ] Configure matrix builds (macOS, Windows, Linux)
- [ ] Add test coverage reporting (codecov)
- [ ] Add coverage threshold enforcement (70% minimum)
- [ ] Add race detector to tests
- [ ] Upload test results as CI artifacts

**Effort:** 4-6 hours

**Priority:** High

---

### Task 4: Code Coverage Analysis

**Current:** ~40% (estimated)
**Target:** 75%+

**Focus Areas:**
- [ ] VM package tests (initialization, reset, execution modes)
- [ ] Parser error path tests
- [ ] Debugger expression tests (error handling)

**Effort:** 4-6 hours

**Priority:** Medium-High

---

### Task 5: Memory Trace Integration ✅ COMPLETED

**Status:** ✅ Completed (2025-10-11)

**Implementation:**
- Added MemoryTrace recording calls after each memory operation in `vm/inst_memory.go`
- Handled all memory access types: WORD, BYTE, and HALF (halfword)
- Added recording for multi-register transfers in `vm/memory_multi.go` (LDM/STM instructions)
- Added nil checks before recording to prevent crashes when tracing is disabled
- Pass correct parameters: sequence number (vm.CPU.Cycles), PC, address, value, size

**Files Modified:**
- `vm/inst_memory.go` - Added RecordRead/RecordWrite calls after load/store operations
- `vm/memory_multi.go` - Added RecordRead/RecordWrite calls in LDM/STM handlers
- `CLAUDE.md` - Removed "Known Issue" note about non-functional memory tracing

**Verification:**
- ✅ All 531 tests pass (100%)
- ✅ golangci-lint reports 0 issues
- ✅ Tested with examples/arrays.s - generated 93 memory trace entries
- ✅ Memory trace output format: `[seq] [TYPE] PC <- [addr] = value (SIZE)`
- ✅ Captures both READ and WRITE operations with correct addresses and values

**Effort:** 2 hours (as estimated)

**Priority:** Medium (advertised feature that wasn't working) - COMPLETED

---

## Phase 12: Performance & Benchmarking

**Effort:** 14-20 hours
**Priority:** Medium

### Benchmarking Suite

**Requirements:**
- [ ] Create benchmark tests (VM, parser, TUI)
- [ ] Document performance targets
- [ ] Run CPU and memory profiling
- [ ] Create `docs/performance_analysis.md`
- [ ] Implement optimizations if needed

**Effort:** 14-20 hours

---

## Phase 13: Release Engineering

**Effort:** 16-22 hours
**Priority:** Medium-High

### Release Pipeline

**Requirements:**
- [ ] Create `.github/workflows/release.yml`
- [ ] Matrix builds (linux-amd64, darwin-amd64, darwin-arm64, windows-amd64)
- [ ] Create release archives and GitHub Release

**Effort:** 4-6 hours
**Priority:** High

---

### Release Documentation

**Requirements:**
- [ ] Create `CHANGELOG.md`
- [ ] Update `README.md` with installation instructions and badges
- [ ] Create `CONTRIBUTING.md`

**Effort:** 3-4 hours
**Priority:** Medium-High

---

### Release Testing

**Requirements:**
- [ ] Create `docs/release_checklist.md`
- [ ] Pre-release verification (tests, coverage, docs)
- [ ] Build testing (all platforms)
- [ ] Installation testing (all package managers)
- [ ] Functional testing (examples, TUI, CLI)

**Effort:** 3-4 hours
**Priority:** High

---

## Additional Diagnostic Modes ✅ COMPLETED (Phase 11)

**Status:** ✅ Completed (2025-10-11)
**Actual Effort:** 4 hours
**Priority:** Low-Medium

### Implemented Diagnostic Features

**Completed:**
- ✅ Code Coverage Mode - Track which instructions/addresses were executed vs not executed
- ✅ Stack Trace Mode - Track stack operations (push/pop/SP modifications) and detect overflow/underflow
- ✅ Flag Change Trace - Dedicated mode to track CPSR flag changes for debugging conditional logic

**Implementation Details:**
- Created `vm/coverage.go` - Code coverage tracker with execution counts and coverage percentage
- Created `vm/stack_trace.go` - Stack trace tracker with overflow/underflow detection
- Created `vm/flag_trace.go` - Flag change tracker for CPSR flag modifications
- Integrated all trackers into VM execution flow
- Added command-line flags: `--coverage`, `--stack-trace`, `--flag-trace`
- All modes support both text and JSON output formats
- Added integration with data processing and memory operations
- All 531 tests still pass (100%)
- Zero lint issues

**Files Modified:**
- `vm/executor.go` - Added diagnostic tracking fields to VM struct
- `vm/cpu.go` - Added SetSPWithTrace for stack tracking
- `vm/data_processing.go` - Added stack trace recording for SP modifications
- `vm/memory_multi.go` - Added stack trace recording for LDM/STM with SP
- `main.go` - Added command-line flags and initialization code
- `CLAUDE.md` - Updated documentation with diagnostic modes section

**Future Considerations:**
- [ ] Register Access Pattern Analysis - Track which registers are most frequently read/written
- [ ] Data Flow Trace - Track how data flows through registers (value provenance)
- [ ] Cycle-Accurate Timing - Per-instruction timing breakdown for performance analysis
- [ ] Symbol-Aware Trace - Enhanced trace showing function names instead of just addresses
- [ ] Diff Mode - Compare register/memory state between two execution points
- [ ] Memory Region Heatmap - Visualize which memory regions are most accessed
- [ ] Assertion/Invariant Checking - Verify user-defined conditions during execution
- [ ] Reverse Execution Log - Log sufficient state to enable stepping backwards
- [ ] Pipeline Visualization - Show instruction flow through execution stages

---

## Additional Documentation (Optional)

**Effort:** 8-11 hours
**Priority:** Low

### Nice-to-Have Docs

- [ ] Tutorial guide (step-by-step walkthrough)
- [ ] FAQ (common errors, platform issues, tips)
- [ ] API reference (all packages)

---

## Effort Summary

**Total estimated effort to v1.0.0:** 30-45 hours

**By Priority:**
- **High (Phase 11):** 7-12 hours - CI/CD enhancements, cross-platform testing, code coverage
- **Medium (Phase 12-13):** 16-23 hours - Release pipeline, performance benchmarking
- **Low (Optional):** 8-11 hours - Additional documentation, advanced features

**Completed:**
- ✅ Integer conversion issues fixed (1h)
- ✅ Literal pool bug fixed (0h - was already fixed in commit b6c59e2)
- ✅ Memory trace integration fixed (2h)
