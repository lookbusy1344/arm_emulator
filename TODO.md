# ARM2 Emulator TODO List

**Last Updated:** 2025-10-16

This file tracks outstanding work only. Completed items are in `PROGRESS.md`.

---

## Summary

**Current Phase:** Phase 11 (Production Hardening) - Core Complete + ARMv3 Extensions

**Status:** Project is production-ready with comprehensive test coverage and all critical features implemented. All planned ARMv3/ARMv3M instruction extensions have been completed (long multiply, PSR transfer, NOP, LDR Rd, =value). Remaining work focuses on release engineering, CI/CD improvements, and optional enhancements.

**Test Status:** 1106 tests, 100% pass rate, 0 lint issues

**Estimated effort to v1.0.0:** 12-18 hours (CI/CD and release engineering only)

---

## High Priority Tasks

### ✅ ~~Planned ARM2 Instruction Extensions~~ - COMPLETED (2025-10-16)
**Actual Effort:** ~6 hours total

**All planned instructions have been implemented and tested:**

#### ✅ ARMv3M Long Multiply Instructions - COMPLETED
- [x] **UMULL** - Unsigned Multiply Long (vm/multiply.go)
- [x] **UMLAL** - Unsigned Multiply-Accumulate Long
- [x] **SMULL** - Signed Multiply Long
- [x] **SMLAL** - Signed Multiply-Accumulate Long
- **Tests:** 14 unit tests in tests/unit/vm/multiply_test.go
- **Status:** All tests passing

#### ✅ ARMv3 PSR Transfer Instructions - COMPLETED
- [x] **MRS** - Move PSR to Register (vm/psr.go)
- [x] **MSR** - Move Register to PSR (supports register and immediate forms)
- **Tests:** 13 unit tests in tests/unit/vm/psr_test.go
- **Status:** All tests passing

#### ✅ Pseudo-Instructions - COMPLETED
- [x] **NOP** - No operation (encoder/other.go:encodeNOP)
- [x] **LDR Rd, =value** - Load 32-bit constant (already implemented)
  - Smart encoding: MOV/MVN for small values, literal pool for large values
  - Automatic value deduplication
  - Multiple pool support via `.ltorg` directive
- [x] **`.ltorg` directive** - Literal pool placement (already implemented)
- **Tests:** 5 integration tests in tests/integration/ltorg_test.go
- **Status:** All tests passing

**Note:** These instructions extend beyond core ARM2 into ARMv3/ARMv3M territory, but are useful for enhanced compatibility. See PROGRESS.md for full implementation details.

**Documentation:** INSTRUCTIONS.md updated with complete syntax, examples, and usage notes for all new instructions.

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

**ARMv2a Atomic Operations** (Effort: 4-6 hours)
- [ ] SWP, SWPB - Atomic swap operations

**ARMv2 Coprocessor Interface** (Effort: 20-30 hours)
- [ ] CDP, LDC, STC, MCR, MRC - Coprocessor operations
- [ ] Full coprocessor emulation framework

**Note:** The emulator has complete ARM2 instruction set coverage. The long multiply and PSR transfer instructions listed in High Priority are now being prioritized for implementation. These remaining extensions are from later architectures.

### Additional Diagnostic Modes

- [ ] Register access pattern analysis
- [ ] Data flow tracing
- [ ] Cycle-accurate timing simulation
- [ ] Symbol-aware trace output
- [ ] Memory region heatmap visualization
- [ ] Reverse execution log

---

## Effort Summary

**Total estimated effort to v1.0.0:** 12-18 hours

- **High Priority:** 12-18 hours (CI/CD and release engineering only)
  - ~~Planned instruction extensions: 20-30 hours~~ ✅ COMPLETED
  - CI/CD and release engineering: 12-18 hours
- **Medium Priority:** 18-27 hours (coverage, benchmarking, documentation)
- **Low Priority:** Variable (optional enhancements)

**Completed in this session (2025-10-16):** ARMv3/ARMv3M instruction extensions (~6 hours actual vs 20-30 hours estimated)
