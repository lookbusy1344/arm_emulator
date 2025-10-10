# ARM2 Emulator TODO List

**IMPORTANT:** This file tracks outstanding tasks and known issues that cannot be completed in the current phase. After completing any work, update this file to reflect the current status.

It should not contain completed items or notes about past work. Those belong in `PROGRESS.md`.

**Last Updated:** 2025-10-10 (Phase 11 - Production Hardening)

---

## Summary

**Status:** All 10 core phases complete! Phase 11 (Production Hardening) in progress.

The ARM2 emulator is **mostly functional** but has critical issues in example programs:
- ✅ All ARM2 instructions implemented and tested
- ✅ Full debugger with TUI
- ⚠️ System calls have input handling issues (SYS_READ_INT returns 0/-1 without user input)
- ✅ 511 tests (509 passing, 99.6% pass rate)
- ✅ Cross-platform configuration
- ✅ Tracing and performance statistics
- ✅ Development tools (linter, formatter, xref)
- ⚠️ 17 example programs (only 6/17 passing, 11 with critical issues)
- ✅ Comprehensive documentation
- ✅ Code quality tools (golangci-lint integrated, 0 lint issues)

**Remaining Work:**
- **Critical Priority:** Fix example program malfunctions (input handling, memory permissions, encoding errors)
- **High Priority:** CI/CD enhancements (matrix builds, code coverage), cross-platform testing
- **Medium Priority:** Release pipeline, installation packages, performance benchmarking
- **Low Priority:** Character literal support (2 failing tests), trace/stats integration, advanced features

**Estimated effort to v1.0.0:** 55-80 hours (increased from 45-65 due to example program issues)

---

## Known Issues

### Example Program Malfunctions (11 of 17 programs have issues)

**Testing Date:** 2025-10-10

**Passing Programs (6/17):**
- ✅ hello.s - Works correctly
- ✅ arithmetic.s - Works correctly
- ✅ gcd.s - Works correctly (but reads 0,0 without user input)
- ✅ loops.s - Works correctly
- ✅ conditionals.s - Works correctly
- ✅ arrays.s - Partial success (crashes with unimplemented SWI 0xCD near end)

**Failing Programs (11/17):**

1. **times_table.s** - Runtime error: "unsupported base: 4294967295"
   - Error at PC=0x00008078
   - Appears to be related to input parsing

2. **factorial.s** - Runtime error: "unsupported base: 4294967295"
   - Error at PC=0x000080BC
   - Same input parsing issue as times_table.s

3. **fibonacci.s** - Input validation failure
   - Prints "Error: Please enter a positive number"
   - Never receives valid input (reads 0 without user input)

4. **string_reverse.s** - Input failure
   - Prints "Error: Empty string"
   - Never receives string input

5. **functions.s** - Memory access error
   - Error at PC=0x000080C4: "write permission denied for segment 'code' at 0x0000827C"
   - Attempting to write to code segment

6. **strings.s** - Infinite loop
   - Runtime error: "cycle limit exceeded (1000000 cycles)"
   - Error at PC=0x00008144

7. **stack.s** - Memory access error
   - Error at PC=0x00008128: "write permission denied for segment 'code' at 0x00008304"
   - Attempting to write to code segment

8. **linked_list.s** - Encoding error
   - "failed to encode instruction at 0x000080F0 (STR): invalid immediate value:  4"
   - STR instruction with invalid immediate offset

9. **bubble_sort.s** - Encoding error
   - "failed to encode instruction at 0x0000805C (STR): invalid immediate value"
   - STR instruction with invalid immediate offset

10. **binary_search.s** - Encoding error
    - "failed to encode instruction at 0x000080B0 (MOV): immediate value 0xFFFFFFFF cannot be encoded as ARM immediate"
    - Attempting to load -1 with MOV instead of MVN

11. **calculator.s** - Infinite loop
    - Runtime error: "cycle limit exceeded (1000000 cycles)"
    - Error at PC=0x000081C0
    - Infinite loop reading invalid operations

**Root Causes Identified:**

1. **Input Handling Issues:**
   - SYS_READ_INT appears to return 0 or -1 (0xFFFFFFFF) without prompting user
   - Programs expecting user input fail or loop infinitely
   - Affects: times_table.s, factorial.s, fibonacci.s, string_reverse.s, gcd.s, calculator.s

2. **Memory Permissions:**
   - Programs attempting to write to code segment
   - May be related to data section placement or stack initialization
   - Affects: functions.s, stack.s

3. **Instruction Encoding:**
   - STR with immediate offsets not properly validated/encoded
   - MOV with 0xFFFFFFFF should use MVN R0, #0 instead
   - Affects: linked_list.s, bubble_sort.s, binary_search.s

4. **Unimplemented Syscalls:**
   - SWI 0xCD not implemented
   - Affects: arrays.s (partial)

5. **Infinite Loops:**
   - Likely due to input reading failures causing loop conditions to never be met
   - Affects: strings.s, calculator.s

**Effort:** 10-15 hours to fix all issues

**Priority:** High (example programs are key demonstration of emulator capabilities)

---

### Character Literal Escaping (2 Failing Tests)

**Impact:** Character literals in assembly not supported

**Issue:** Character literals in immediates not supported:
- `MOV R0, #' '` (space)
- `MOV R0, #'\t'` (tab)

**Workaround:** Use numeric values (`MOV R0, #32` for space, `MOV R0, #9` for tab)

**Effort:** 1-2 hours

**Priority:** Medium (nice to have, not blocking release)

---

## Parser Enhancements (Optional)

### Advanced Addressing Modes

**Status:** Not required for ARM2, but would enable more sophisticated patterns

**Missing Features:**
- Pre-indexed with writeback: `LDR R0, [R1, #4]!`
- Post-indexed: `LDRB R0, [R1], #4`
- Immediate offset: `LDR R0, [R1, #4]`

**Workaround:** Use separate ADD instructions and base register addressing

**Effort:** 3-4 hours

**Priority:** Low (not needed for current examples)

---

## Phase 11: Production Hardening

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

### Task 5: Trace/Stats Integration (Optional)

**Status:** Infrastructure complete, integration optional

**Requirements:**
- [ ] Connect ExecutionTrace to VM.Step()
- [ ] Connect MemoryTrace to Memory operations
- [ ] Connect Statistics to VM operations
- [ ] Add integration tests

**Effort:** 2-3 hours

**Priority:** Low

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

## Summary

**Estimated effort to v1.0.0:** 55-80 hours

**By Priority:**
- **Critical (Example Programs):** 10-15 hours - Fix input handling, memory permissions, encoding errors
- **High (Phase 11):** 15-20 hours - Code quality, CI/CD, cross-platform testing, coverage
- **Medium-High (Phase 13):** 16-22 hours - Release pipeline, packages, documentation
- **Medium (Phase 12):** 14-20 hours - Benchmarking and performance
- **Low (Optional):** 8-11 hours - Additional docs, trace integration, advanced features
