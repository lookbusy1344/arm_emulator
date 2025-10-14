## ~~CRITICAL REGRESSION: Example Programs Broken~~ FIXED ✅

- **Status:** FIXED (2025-10-14)
- **Root Cause:** Multiple issues fixed:
  1. Programs using `.org 0x0000` failed (fixed 2025-10-13)
  2. Negative constants in .equ not supported (fixed 2025-10-14)
  3. Data section ordering bug causing overlaps (fixed 2025-10-14)
  4. Standalone shift instructions not supported (fixed 2025-10-14)
  5. 16-bit immediate encoding failures (fixed 2025-10-14)
- **Solution:** Comprehensive bug fixes and integration test coverage
- **Test Coverage:** 32 of 34 example programs now have integration tests (94%)
- **Success Rate:** 32 of 32 tested programs pass ✅ (100%)
- See PROGRESS.md for full details

# ARM2 Emulator TODO List

**IMPORTANT:** This file tracks outstanding tasks and known issues. After completing any work, update this file to reflect the current status.

Completed items and past work belong in `PROGRESS.md`.

**Last Updated:** 2025-10-14

---

## Summary

**Status:** Phase 11 (Production Hardening) - Complete ✅

**Test Status:** All tests passing (100% ✅)
- Unit tests: ~900 tests
- Integration tests: 62 tests (including 32 example program tests)
- Comprehensive ARM2 instruction coverage achieved

**Example Programs:** 32 of 32 tested programs working (100% ✅)
- 34 total programs in examples/
- 32 programs with integration tests and expected outputs
- 2 programs require interactive input (times_table.s, calculator.s with specific inputs)
- Table-driven test framework makes adding tests trivial

**Recent Completion (2025-10-14):**
- ✅ **Comprehensive Integration Testing:** 32 example programs now tested
- ✅ **Parser Enhancements:** Negative constants, standalone shifts, 16-bit immediates
- ✅ **Bug Fixes:** Data section ordering, immediate encoding, syntax issues
- ✅ **Success Rate:** From ~50% to 100% of tested programs working

**Remaining Work:**
- **High Priority:** CI/CD enhancements (matrix builds, coverage reporting)
- **Medium Priority:** Additional documentation, performance benchmarking
- **Low Priority:** Future enhancements (see below)

**Estimated effort to v1.0.0:** 15-20 hours

---

## Known Issues

### None - All Critical Issues Resolved ✅

All previously reported critical issues have been fixed:
- ✅ Memory access violations - Fixed via proper section handling
- ✅ Parse errors - Fixed via parser enhancements
- ✅ Encoding errors - Fixed via automatic MOVW/CMN fallbacks
- ✅ Negative constants - Now supported
- ✅ Standalone shift instructions - Now supported as pseudo-instructions
- ✅ Data section ordering - Fixed

**Testing Status:** 32 of 32 example programs with tests passing (100%)

---

## Outstanding Tasks

### Phase 11: Production Hardening - Remaining Items

#### Enhanced CI/CD Pipeline
**Priority:** High | **Effort:** 4-6 hours

- [ ] Configure matrix builds (macOS, Windows, Linux)
- [ ] Add test coverage reporting (codecov)
- [ ] Add coverage threshold enforcement (70% minimum)
- [ ] Add race detector to tests
- [ ] Upload test results as CI artifacts

#### Code Coverage Analysis
**Priority:** Medium-High | **Effort:** 4-6 hours

**Current:** ~60% (estimated) | **Target:** 75%+

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

## Missing ARM2/ARMv2 Instructions

### Long Multiply Instructions (ARMv3M - Not in ARM2)
**Priority:** Low | **Effort:** 8-12 hours

These instructions were introduced in ARMv3M (ARM6 and later), not present in ARM2:
- [ ] UMULL - Unsigned Multiply Long (64-bit result)
- [ ] UMLAL - Unsigned Multiply-Accumulate Long (64-bit result)
- [ ] SMULL - Signed Multiply Long (64-bit result)
- [ ] SMLAL - Signed Multiply-Accumulate Long (64-bit result)

**Note:** Already documented in INSTRUCTIONS.md as "Planned". Not historically accurate for ARM2.

### Coprocessor Instructions (ARMv2 - Optional)
**Priority:** Very Low | **Effort:** 20-30 hours

ARMv2 included coprocessor interface support, but these are rarely needed for typical ARM2 programs:
- [ ] CDP - Coprocessor Data Processing
- [ ] LDC - Load Coprocessor register
- [ ] STC - Store Coprocessor register
- [ ] MCR - Move to Coprocessor Register
- [ ] MRC - Move from Coprocessor Register

**Note:** Would require full coprocessor emulation framework. Low priority for ARM2 emulation.

### PSR Transfer Instructions (ARMv3 - Not in ARM2)
**Priority:** Very Low | **Effort:** 4-6 hours

These were introduced in ARMv3, not present in ARM2:
- [ ] MRS - Move PSR to Register
- [ ] MSR - Move Register to PSR

**Note:** Already documented in INSTRUCTIONS.md as "Planned". ARM2 stored PSR flags in R15.

### Atomic Swap Instructions (ARMv2a - Not in ARM2)
**Priority:** Very Low | **Effort:** 4-6 hours

These were added in ARMv2a (ARM3), not present in original ARM2:
- [ ] SWP - Swap word (atomic load-store)
- [ ] SWPB - Swap byte (atomic load-store)

**Note:** ARMv2a extension, part of ARM3. Not in original ARM2.

### Summary: Missing Instructions Status

**Implemented in this emulator:**
- ✅ All ARM2 data processing instructions (16 opcodes)
- ✅ All ARM2 memory instructions (LDR/STR/LDRB/STRB/LDM/STM)
- ✅ ARM2a halfword extensions (LDRH/STRH)
- ✅ All ARM2 branch instructions (B/BL/BX)
- ✅ ARM2 multiply instructions (MUL/MLA)
- ✅ Software interrupts (SWI with 30+ syscalls)

**Not implemented (and historically accurate for ARM2):**
- ❌ Long multiply (ARMv3M only - UMULL/UMLAL/SMULL/SMLAL)
- ❌ PSR transfers (ARMv3 only - MRS/MSR)
- ❌ Atomic swap (ARMv2a only - SWP/SWPB)
- ❌ Coprocessor interface (ARMv2 optional - CDP/LDC/STC/MCR/MRC)

**Conclusion:** This emulator has **complete ARM2 instruction set coverage**. All missing instructions are from later ARM versions (ARMv2a, ARMv3, ARMv3M) or optional coprocessor support.

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

**Total estimated effort to v1.0.0:** 15-20 hours

**By Priority:**
- **High:** 8-12 hours - CI/CD enhancements, code coverage, release pipeline
- **Medium:** 10-15 hours - Performance benchmarking, additional documentation
- **Low (Optional):** 5-10 hours - Future enhancements

---

## Recently Completed (See PROGRESS.md for details)

### October 2025 - Production Hardening Complete ✅
- ✅ Comprehensive integration test coverage (32 of 34 example programs tested)
- ✅ All critical parser bugs fixed (negative constants, standalone shifts, data section ordering)
- ✅ All critical encoder bugs fixed (16-bit immediates, CMP/CMN fallbacks)
- ✅ All example program syntax issues resolved
- ✅ 100% of tested programs passing
- ✅ Table-driven test framework for easy maintenance
- ✅ Expected output files for all programs
- ✅ All lint issues resolved (golangci-lint clean)
- ✅ Go vet warnings fixed (method renames)
- ✅ CI updated to Go 1.25
