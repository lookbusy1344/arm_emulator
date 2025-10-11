# ARM2 Emulator TODO List

**IMPORTANT:** This file tracks outstanding tasks and known issues. After completing any work, update this file to reflect the current status.

Completed items and past work belong in `PROGRESS.md`.

**Last Updated:** 2025-10-11

---

## Summary

**Status:** Phase 11 (Production Hardening) - Most tasks complete!

**Test Status:** 660 tests passing (100% pass rate)
- Unit tests: 575 tests
- Integration tests: 85 tests

**Example Programs:** 21 of 23 fully functional (91% functional rate)

**Remaining Work:**
- **High Priority:** CI/CD enhancements (matrix builds, coverage reporting)
- **Medium Priority:** Code coverage improvements, release pipeline
- **Low Priority:** Performance benchmarking, additional documentation

**Estimated effort to v1.0.0:** 20-30 hours

---

## Known Issues

### Example Program Issues (Non-Critical)

1. **linked_list.s** - Unaligned word access at 0x0000000E
   - Pre-existing bug in the example program
   - Attempting unaligned memory access (must be 4-byte aligned)
   - Needs fix in the assembly code itself

2. **reverse_chatgpt.s** - Parse errors (syntax issues)
   - Multiple syntax errors preventing parsing
   - Pre-existing issues with assembly syntax
   - Needs syntax corrections

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
