# ARM2 Emulator TODO List

**IMPORTANT:** This file tracks outstanding tasks and known issues. After completing any work, update this file to reflect the current status.

Completed items and past work belong in `PROGRESS.md`.

**Last Updated:** 2025-10-11 (Phase 11 - Production Hardening)

---

## Summary

**Status:** All 10 core phases complete! Phase 11 (Production Hardening) in progress.

**Remaining Work:**
- **High Priority:** Literal pool bug fix, CI/CD enhancements, cross-platform testing
- **Medium Priority:** Memory trace integration, release pipeline, performance benchmarking
- **Low Priority:** Additional documentation, advanced features

**Estimated effort to v1.0.0:** 40-60 hours

---

## Known Issues

### Literal Pool Memory Corruption Bug 🐛 OPEN

**Status:** Active bug affecting programs with many `LDR Rx, =label` instructions

**Symptoms:**
- Programs with multiple `LDR Rx, =label` instructions can corrupt memory at offsets 8 and 12
- Reading from stack offsets +8 and +12 returns incorrect values (-1073741806 and 25)
- Error at program end: "unimplemented SWI: 0xNNNNNN" with invalid SWI numbers
- The bug does not occur when using direct register addressing without literal pools

**Evidence:**
- Created test program that stores values at SP+0, SP+4, SP+8, SP+12
- Stores work correctly: 100, 200, 300, 400
- Reads return: 100, 200, -1073741806, 25 (offsets 8 and 12 are corrupted)
- Integration test `TestStackFile` reproduces the bug when using `LDR R0, =msg`
- Same test passes when avoiding literal pool loads

**Root Cause (Updated after investigation):**
- Programs with many `LDR Rx, =label` instructions execute correctly but fail to halt properly
- After producing correct output, execution continues past the final `SWI #0x00` (exit)
- PC advances into memory containing literal pool entries or data
- These values are interpreted as invalid SWI instructions (e.g., SWI 0x04FFC4, SWI 0xCD)
- The literal pool placement formula `(currentAddr & 0xFFFFF000) + 0x1000` may cause overlap issues
- Simple programs (1-2 literals) work; complex programs (8+ literals) fail

**Investigation Notes (2025-10-11):**
- Added `LiteralPoolStart` field to encoder to control literal pool placement
- Tried placing literal pool between instructions and data → still fails
- Tried placing literal pool after all data → still fails
- Added PC-relative offset validation (must be within 4KB) → helped but didn't solve it
- Integration test `TestStackFile` actually PASSES (stack corruption may be different manifestation)
- The issue may be related to execution not halting properly, not just literal placement

**Workaround:**
- Minimize use of `LDR Rx, =label` in programs
- Use direct values or register-to-register operations where possible
- Note: `examples/addressing_modes.s` still exhibits the bug despite workarounds

**Files Modified:**
- `encoder/memory.go` - Added `LiteralPoolStart` support and offset validation
- `main.go` - Set literal pool start address during assembly
- `tests/integration/syscalls_test.go` - Updated test helper with same logic

**Test Case:**
- Example programs: `examples/addressing_modes.s` (8 literals), `examples/arrays.s` (16 literals)
- Both produce correct output but then hit runtime errors
- Integration test: `tests/integration/test_stack_file_test.go` (PASSES - may not reproduce this specific issue)

**Next Steps:**
- Investigate why execution continues after `SWI #0x00` in programs with many literals
- Check if branch target calculations are affected by literal pool presence
- Verify SWI #0x00 handler properly sets VM state to halt
- Consider implementing proper multi-pass assembly with literal pool size calculation

**Effort:** 8-12 hours (complex issue requiring deeper investigation)

**Priority:** High (affects program correctness)

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

### Task 5: Memory Trace Integration (BROKEN)

**Status:** Infrastructure exists but not functional - `--mem-trace` flag does not work

**Problem:**
- MemoryTrace infrastructure is set up correctly (created, started, flushed in main.go)
- RecordRead() and RecordWrite() methods exist in vm/trace.go but are never called
- Result: Empty trace files are created with no memory access data

**Root Cause:**
- Load/store instruction handlers in `vm/inst_memory.go` call memory operations (ReadWord, WriteWord, etc.) but don't record traces
- Memory read/write functions in `vm/memory.go` don't have access to VM's MemoryTrace
- Same issue likely exists in multi-register transfers (`vm/memory_multi.go`) and other memory operations

**Fix Required:**
- [ ] Add MemoryTrace recording calls after each memory operation in `vm/inst_memory.go`
- [ ] Handle all memory access types: word (lines 79, 102), byte (lines 74, 99), halfword (lines 69, 95)
- [ ] Add recording for multi-register transfers in `vm/memory_multi.go` (LDM/STM instructions)
- [ ] Check if MemoryTrace is enabled before recording (nil check)
- [ ] Pass correct parameters: sequence number (vm.CPU.Cycles), PC, address, value, size ("WORD", "BYTE", "HALF")
- [ ] Add integration test to verify memory trace output is generated
- [ ] Optionally connect ExecutionTrace to VM.Step()
- [ ] Optionally connect Statistics to VM operations

**Files to Modify:**
- `vm/inst_memory.go` - Add RecordRead/RecordWrite calls after memory operations (lines 69-104)
- `vm/memory_multi.go` - Add RecordRead/RecordWrite calls in LDM/STM handlers

**Effort:** 2-3 hours

**Priority:** Medium (advertised feature that doesn't work)

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

## Additional Documentation (Optional)

**Effort:** 8-11 hours
**Priority:** Low

### Nice-to-Have Docs

- [ ] Tutorial guide (step-by-step walkthrough)
- [ ] FAQ (common errors, platform issues, tips)
- [ ] API reference (all packages)

---

## Effort Summary

**Total estimated effort to v1.0.0:** 40-60 hours

**By Priority:**
- **High (Phase 11):** 15-20 hours - Literal pool bug (4-6h), CI/CD enhancements, cross-platform testing, code coverage
- **Medium (Phase 12-13):** 18-26 hours - Memory trace integration (2-3h), release pipeline, performance benchmarking
- **Low (Optional):** 8-11 hours - Additional documentation, advanced features

**Completed:**
- ✅ Integer conversion issues fixed (1h)
