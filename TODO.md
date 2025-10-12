# ARM2 Emulator TODO List

**IMPORTANT:** This file tracks outstanding tasks and known issues. After completing any work, update this file to reflect the current status.

Completed items and past work belong in `PROGRESS.md`.

**Last Updated:** 2025-10-12

---

## Summary

**Status:** Phase 11 (Production Hardening) - Most tasks complete!

**Test Status:** 660 tests passing (100% pass rate)
- Unit tests: 575 tests
- Integration tests: 85 tests

**Example Programs:** 22 of 30 fully functional (73% functional rate)

**Remaining Work:**
- **High Priority:** CI/CD enhancements (matrix builds, coverage reporting)
- **Medium Priority:** Code coverage improvements, release pipeline
- **Low Priority:** Performance benchmarking, additional documentation

**Estimated effort to v1.0.0:** 20-30 hours

---

## Known Issues

### Example Program Issues (Non-Critical)

**Test Summary (30 programs total):**
- ✅ **22 programs fully working** (73%)
- ⚠️ **3 programs with input issues** (10%)
- ❌ **5 programs with errors** (17%)

#### Working Programs (22)
1. ✅ addressing_modes.s - All addressing mode tests passed
2. ✅ arithmetic.s - All arithmetic operations work correctly
3. ✅ arrays.s - Array operations demo works
4. ✅ binary_search.s - Binary search works correctly
5. ✅ bit_operations.s - All bit operation tests passed
6. ✅ conditionals.s - All conditional execution tests passed
7. ✅ factorial.s - Factorial calculation works
8. ✅ fibonacci.s - Fibonacci sequence generation works
9. ✅ functions.s - Function calling conventions work
10. ✅ gcd.s - GCD calculation works correctly
11. ✅ hello.s - Hello world works
12. ✅ linked_list.s - Linked list operations work
13. ✅ loops.s - All loop constructs work correctly
14. ✅ memory_stress.s - All memory tests passed
15. ✅ nested_calls.s - Deep nested calls work correctly
16. ✅ recursive_factorial.s - Recursive factorial works
17. ✅ stack.s - Stack-based calculator works
18. ✅ string_reverse.s - String reversal works
19. ✅ strings.s - String operations work
20. ✅ times_table.s - Times table generation works

#### Programs with Input Issues (3)
21. ⚠️ **bubble_sort.s** - Expects interactive input, runs but shows zeros with stdin input
22. ⚠️ **calculator.s** - Input reading issue (operation character not read correctly, infinite loop)

#### Programs with Errors (5)
23. ❌ **hash_table.s** - Parse error: "invalid constant value: -" at line 10
   - Parser doesn't support "-" as a constant value placeholder

24. ❌ **matrix_multiply.s** - Runtime error: memory access violation at 0x000081D4
   - Attempts to read string at invalid address 0x00000002

25. ❌ **quicksort.s** - Runtime error: memory access violation at 0x000081EC
   - Attempts to read string at invalid address 0x00000011

26. ❌ **recursive_fib.s** - Parse errors: multiple syntax issues
   - Contains '@' characters (comments?) and parentheses in unexpected places
   - Appears to use non-ARM2 syntax

27. ❌ **reverse_chatgpt.s** - Parse errors: unexpected NUMBER tokens (lines 8, 13, 25, 32, 37)
   - Syntax issues preventing parsing

28. ❌ **sieve_of_eratosthenes.s** - Parse errors: extensive syntax issues
   - Contains '@' characters, parentheses, operators in unexpected places
   - Appears to use non-ARM2 syntax

29. ❌ **state_machine.s** - Runtime error: cycle limit exceeded (1000000 cycles)
   - Program enters infinite loop during email validation

30. ❌ **xor_cipher.s** - Encoding error: unknown instruction "LSR" at 0x00008220
   - LSR instruction not implemented in encoder

---

## Outstanding Tasks

### Phase 11: Production Hardening

#### Enhanced CI/CD Pipeline
**Priority:** High | **Effort:** 4-6 hours

- [ ] Configure matrix builds (macOS, Windows, Linux)
- [ ] Add test coverage reporting (codecov)
- [ ] Add coverage threshold enforcement (70% minimum)
- [ ] Add race detector to tests
- [ ] Upload test results as CI artifacts

#### Code Coverage Analysis
**Priority:** Medium-High | **Effort:** 4-6 hours

**Current:** ~40% (estimated) | **Target:** 75%+

Focus areas:
- [ ] VM package tests (initialization, reset, execution modes)
- [ ] Parser error path tests
- [ ] Debugger expression tests (error handling)

---

### Phase 12: Performance & Benchmarking
**Priority:** Medium | **Effort:** 14-20 hours

- [ ] Create benchmark tests (VM, parser, TUI)
- [ ] Document performance targets
- [ ] Run CPU and memory profiling
- [ ] Create `docs/performance_analysis.md`
- [ ] Implement optimizations if needed

---

### Phase 13: Release Engineering
**Priority:** High | **Effort:** 12-16 hours

#### Release Pipeline
- [ ] Create `.github/workflows/release.yml`
- [ ] Matrix builds (linux-amd64, darwin-amd64, darwin-arm64, windows-amd64)
- [ ] Create release archives and GitHub Release

#### Release Documentation
- [ ] Create `CHANGELOG.md`
- [ ] Create `CONTRIBUTING.md`

#### Release Testing
- [ ] Create `docs/release_checklist.md`
- [ ] Pre-release verification (tests, coverage, docs)
- [ ] Build testing (all platforms)
- [ ] Functional testing (examples, TUI, CLI)

---

## Future Enhancements (Optional)

### Additional Diagnostic Modes
**Priority:** Low

- [ ] Register Access Pattern Analysis - Track register read/write frequency
- [ ] Data Flow Trace - Track value provenance through registers
- [ ] Cycle-Accurate Timing - Per-instruction timing breakdown
- [ ] Symbol-Aware Trace - Show function names instead of addresses
- [ ] Diff Mode - Compare register/memory state between execution points
- [ ] Memory Region Heatmap - Visualize memory access patterns
- [ ] Assertion/Invariant Checking - Verify user-defined conditions
- [ ] Reverse Execution Log - Enable stepping backwards

### Additional Documentation
**Priority:** Low

- [ ] Tutorial guide (step-by-step walkthrough)
- [ ] FAQ (common errors, platform issues, tips)
- [ ] API reference (all packages)

---

## Effort Summary

**Total estimated effort to v1.0.0:** 20-30 hours

**By Priority:**
- **High:** 10-16 hours - CI/CD enhancements, code coverage, release pipeline
- **Medium:** 14-20 hours - Performance benchmarking
- **Low (Optional):** 10-15 hours - Future enhancements, additional documentation

---

## Recently Completed (See PROGRESS.md for details)

- ✅ Integer conversion issues fixed (gosec G115 warnings)
- ✅ ARM immediate encoding rotation bug fixed (fibonacci.s, calculator.s)
- ✅ Memory trace integration completed
- ✅ Diagnostic modes implemented (code coverage, stack trace, flag trace)
- ✅ CLI diagnostic flags with integration tests
- ✅ All lint issues resolved (golangci-lint clean)
- ✅ Go vet warnings fixed (method renames)
- ✅ CI updated to Go 1.25
- ✅ Parser limitations resolved (debugger expression parser rewritten)
- ✅ All example programs working (21 of 23)
