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
### HIGH PRIORITY: File I/O example failure / CMP & conditional branch anomaly

Context: `examples/file_io.s` continues to report FAIL during write verification despite correct byte count written (64) and preserved flags. Investigation uncovered multiple defects and remaining unresolved issues.

Findings:
- Initial failure due to incomplete file syscall layer (READ/WRITE ignored fd). Implemented fd table and real syscalls.
- SWI handler originally clobbered CPSR flags; added save/restore in `ExecuteSWI`.
- CMP instruction caused infinite loop because PC not incremented when Rd field encoded as R15 for flag-only ops (CMP/TST/TEQ/CMN). Fixed by always incrementing PC for these opcodes (data_processing.go).
- After fixes, CMP sets result=0 (op1==op2==64) and eventually Z=1 appears in debug logs, yet BNE still taken in `write_file` length check.
- Debug output shows multiple CMP executions at same address prior to fix; resolved loop but logical compare still failing path.
- Current instrumentation shows:
  - R6 (saved write length) = 64
  - Immediate LENGTH = 64
  - CMP result = 0
  - Z flag transitions to true, but failure branch path executed (suggests conditional branch evaluation mismatch or stale flag read for that specific branch instruction instance).

Hypotheses:
1. Assembler encodes `BNE` with incorrect condition code bits (e.g., using condition before flag update due to PC pipeline offset mishandling in branching logic) — need to disassemble instruction word for offending BNE.
2. Race in execution order: branch instruction condition evaluated using flags before CMP flags visible (would require two sequential CMPs; not supported by current pipeline model but logs show delayed Z true occurrence).
3. Flag update logic correct; EvaluateCondition returns expected value, but decoded condition on BNE instruction not mapping to CondNE (e.g., mis-decoding opcode bits 31:28).
4. Memory corruption overwriting CPSR between CMP and branch (less likely after SWI fixes—no SWI in between now).

Required Diagnostics / Actions:
- [ ] Add disassembly / dump of instruction word at failing BNE PC (capture PC, opcode, decoded condition) when write_file length check executes.
- [ ] Instrument `EvaluateCondition` to log cond, current flags, and decision for that specific branch only.
- [ ] Confirm assembler output for `CMP R6,#LENGTH` (check Rd bits, ensure not accidentally writing to PC).
- [ ] Verify that after CMP execution, next fetched instruction address matches expected sequential PC (no accidental re-fetch of CMP due to write-to-PC detection).
- [ ] Create minimal reproduction assembly: MOV R0,#64; MOV R6,R0; CMP R6,#64; BNE fail; PASS: ... and verify pass inside same build (baseline already passes in cmp_test.s) — compare encoded BNE against failing one.
- [ ] After diagnostics, remove temporary debug code added to examples and VM.
- [ ] Add regression test covering CMP with Rd bits=15 case and conditional branch immediately following (ensuring PC increment and correct condition evaluation).

Blocking Release: Yes (affects correctness of conditional execution and example reliability).

Owner: TBD
Priority: HIGH
ETA: 2-4 hours (diagnostics + fix + tests)



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
