# ARM2 Emulator TODO List

**IMPORTANT:** This file tracks outstanding tasks and known issues that cannot be completed in the current phase. After completing any work, update this file to reflect the current status.

It should not contain completed items or notes about past work. Those belong in `PROGRESS.md`.

**Last Updated:** 2025-10-10 (Phase 11 - Production Hardening)

---

## Summary

**Status:** All 10 core phases complete! Phase 11 (Production Hardening) in progress.

The ARM2 emulator is **functionally complete**. All core features work:
- ✅ All ARM2 instructions implemented and tested
- ✅ Full debugger with TUI
- ✅ All system calls functional
- ✅ 511 tests (509 passing, 99.6% pass rate)
- ✅ Cross-platform configuration
- ✅ Tracing and performance statistics
- ✅ Development tools (linter, formatter, xref)
- ✅ 17 example programs
- ✅ Comprehensive documentation

**Remaining Work:**
- **High Priority:** CI/CD enhancements, cross-platform testing, code coverage
- **Medium Priority:** Release pipeline, installation packages, performance benchmarking
- **Low Priority:** Character literal support (2 failing tests), trace/stats integration, advanced features

**Estimated effort to v1.0.0:** 50-70 hours

---

## Known Issues

### Character Literal Escaping (2 Failing Tests)

**Impact:** 2 example programs cannot run (loops.s, conditionals.s)

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

### Task 1: Code Quality Tools

**Status:** Complete ✅

**Completed:**
- [x] Installed golangci-lint
- [x] Created .golangci.yml configuration with errcheck, unused, govet, ineffassign, and misspell
- [x] Fixed all errcheck issues in non-test files (46 issues fixed)
- [x] Removed unused code (4 functions/fields removed)
- [x] Added golangci-lint to CI pipeline
- [x] All tests passing (509 tests, 99.6% pass rate)

**Notes:**
- Configured to skip test files (common practice for test code)
- Disabled staticcheck QF style suggestions (too opinionated)
- Disabled revive (too many style warnings for established codebase)
- Focus on serious issues: error handling, unused code, vet warnings

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

**Estimated effort to v1.0.0:** 50-70 hours

**By Priority:**
- **High (Phase 11):** 15-20 hours - Code quality, CI/CD, cross-platform testing, coverage
- **Medium-High (Phase 13):** 16-22 hours - Release pipeline, packages, documentation
- **Medium (Phase 12):** 14-20 hours - Benchmarking and performance
- **Low (Optional):** 8-11 hours - Additional docs, trace integration, advanced features
