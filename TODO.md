# ARM2 Emulator TODO List

**Last Updated:** 2025-10-15

This file tracks outstanding work only. Completed items are in `PROGRESS.md`.

---

## Summary

**Current Phase:** Phase 11 (Production Hardening) - Core Complete

**Status:** Project is production-ready with comprehensive test coverage and all critical features implemented. ADR pseudo-instruction verified working. `.ltorg` directive implemented and tested. Remaining work focuses on release engineering, CI/CD improvements, and optional enhancements.

**Estimated effort to v1.0.0:** 35-50 hours (includes planned instruction extensions)

---

## High Priority Tasks

### Planned ARM2 Instruction Extensions
**Effort:** 20-30 hours

From INSTRUCTIONS.md - these instructions are documented but not yet implemented:

#### ARMv3M Long Multiply Instructions (8-12 hours)
- [ ] **UMULL** - Unsigned Multiply Long (`UMULL{cond}{S} RdLo, RdHi, Rm, Rs`)
  - Multiplies two 32-bit unsigned values producing 64-bit result
  - Operation: `RdHi:RdLo = Rm * Rs`
  - INSTRUCTIONS.md lines 612-620

- [ ] **UMLAL** - Unsigned Multiply-Accumulate Long (`UMLAL{cond}{S} RdLo, RdHi, Rm, Rs`)
  - Unsigned multiply and accumulate with 64-bit result
  - Operation: `RdHi:RdLo = (Rm * Rs) + RdHi:RdLo`
  - INSTRUCTIONS.md lines 622-630

- [ ] **SMULL** - Signed Multiply Long (`SMULL{cond}{S} RdLo, RdHi, Rm, Rs`)
  - Multiplies two 32-bit signed values producing 64-bit result
  - Operation: `RdHi:RdLo = Rm * Rs (signed)`
  - INSTRUCTIONS.md lines 632-640

- [ ] **SMLAL** - Signed Multiply-Accumulate Long (`SMLAL{cond}{S} RdLo, RdHi, Rm, Rs`)
  - Signed multiply and accumulate with 64-bit result
  - Operation: `RdHi:RdLo = (Rm * Rs) + RdHi:RdLo (signed)`
  - INSTRUCTIONS.md lines 642-650

#### ARMv3 PSR Transfer Instructions (4-6 hours)
- [ ] **MRS** - Move PSR to Register (`MRS{cond} Rd, PSR`)
  - Moves CPSR or SPSR to a register
  - Example: `MRS R0, CPSR` (R0 = CPSR)
  - INSTRUCTIONS.md lines 717-728

- [ ] **MSR** - Move Register to PSR (`MSR{cond} PSR, Rm`)
  - Moves a register or immediate to CPSR or SPSR
  - Example: `MSR CPSR, R0` (CPSR = R0)
  - INSTRUCTIONS.md lines 730-741

#### Pseudo-Instructions (4-6 hours)
- [ ] **NOP** - No operation (pseudo-instruction)
  - Maps to: `MOV R0, R0`
  - INSTRUCTIONS.md line 1242

- [ ] **LDR Rd, =value** - Load 32-bit constant (pseudo-instruction)
  - Maps to: `LDR Rd, [PC, #offset]` with literal pool
  - Allows loading any 32-bit constant value
  - INSTRUCTIONS.md line 1243

**Note:** These are documented in INSTRUCTIONS.md as planned but not yet implemented. They extend beyond core ARM2 into ARMv3/ARMv3M territory.

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

**Total estimated effort to v1.0.0:** 35-50 hours

- **High Priority:** 32-48 hours (planned instructions, CI/CD, release engineering)
  - Planned instruction extensions: 20-30 hours
  - CI/CD and release engineering: 12-18 hours
- **Medium Priority:** 18-27 hours (coverage, benchmarking, documentation)
- **Low Priority:** Variable (optional enhancements)
