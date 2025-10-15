# ARM2 Emulator TODO List

**Last Updated:** 2025-10-14

This file tracks outstanding work only. Completed items are in `PROGRESS.md`.

---

## Summary

**Current Phase:** Phase 11 (Production Hardening) - Core Complete

**Status:** Project is production-ready with comprehensive test coverage and all critical features implemented. Remaining work focuses on release engineering, CI/CD improvements, and optional enhancements.

**Estimated effort to v1.0.0:** 15-20 hours

---

## High Priority Tasks

### Enhanced CI/CD Pipeline
**Effort:** 4-6 hours

- [ ] Configure matrix builds (macOS, Windows, Linux)
- [ ] Add test coverage reporting (codecov)
- [ ] Add coverage threshold enforcement (70% minimum)
- [ ] Add race detector to tests
- [ ] Upload test results as CI artifacts

### Release Engineering
**Effort:** 8-12 hours

- [ ] Create `.github/workflows/release.yml` with matrix builds
  - linux-amd64, darwin-amd64, darwin-arm64, windows-amd64
- [ ] Create release archives and GitHub Release automation
- [ ] Create `CHANGELOG.md`
- [ ] Create `CONTRIBUTING.md`
- [ ] Create `docs/release_checklist.md`
- [ ] Pre-release verification checklist

---

## Medium Priority Tasks

### Re-enable TUI Automated Tests
**Effort:** 2-3 hours

- [ ] Refactor `NewTUI` to accept optional `tcell.Screen` parameter for testing
- [ ] Use `tcell.SimulationScreen` in tests instead of real terminal
- [ ] Re-enable `tui_manual_test.go.disabled` as `tui_test.go`
- [ ] Verify tests run without hanging in CI/CD

**Context:** Current TUI tests are disabled because `tview.NewApplication()` tries to initialize a real terminal, causing hangs during `go test`. Using tcell's SimulationScreen allows testing without a terminal.

### Code Coverage Improvements
**Effort:** 4-6 hours

**Current:** ~60% (estimated) | **Target:** 75%+

Focus areas:
- [ ] VM package tests (initialization, reset, execution modes)
- [ ] Parser error path tests
- [ ] Debugger expression tests (error handling)

### Performance & Benchmarking
**Effort:** 10-15 hours

- [ ] Create benchmark tests (VM, parser, TUI)
- [ ] Document performance targets
- [ ] Run CPU and memory profiling
- [ ] Create `docs/performance_analysis.md`
- [ ] Implement optimizations if needed

### Additional Documentation
**Effort:** 4-6 hours

- [ ] Tutorial guide (step-by-step walkthrough)
- [ ] FAQ (common errors, platform issues, tips)
- [ ] API reference documentation

---

## Low Priority Tasks (Optional)

### Later ARM Architecture Extensions (Optional)

These are **not** part of ARM2 but could be added for broader compatibility:

**ARMv3M Long Multiply** (Effort: 8-12 hours)
- [ ] UMULL, UMLAL, SMULL, SMLAL - 64-bit multiply operations

**ARMv3 PSR Transfers** (Effort: 4-6 hours)
- [ ] MRS, MSR - Direct PSR register access

**ARMv2a Atomic Operations** (Effort: 4-6 hours)
- [ ] SWP, SWPB - Atomic swap operations

**ARMv2 Coprocessor Interface** (Effort: 20-30 hours)
- [ ] CDP, LDC, STC, MCR, MRC - Coprocessor operations
- [ ] Full coprocessor emulation framework

**Note:** The emulator has complete ARM2 instruction set coverage. These extensions are from later architectures.

### Additional Diagnostic Modes

- [ ] Register access pattern analysis
- [ ] Data flow tracing
- [ ] Cycle-accurate timing simulation
- [ ] Symbol-aware trace output
- [ ] Memory region heatmap visualization
- [ ] Reverse execution log

---

## Effort Summary

**Total estimated effort to v1.0.0:** 15-20 hours

- **High Priority:** 12-18 hours (CI/CD, release engineering)
- **Medium Priority:** 18-27 hours (coverage, benchmarking, documentation)
- **Low Priority:** Variable (optional enhancements)
