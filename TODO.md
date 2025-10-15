# ARM2 Emulator TODO List

**Last Updated:** 2025-10-15

This file tracks outstanding work only. Completed items are in `PROGRESS.md`.

---

## Summary

**Current Phase:** Phase 11 (Production Hardening) - Core Complete

**Status:** Project is production-ready with comprehensive test coverage and all critical features implemented. ADR pseudo-instruction verified working. One high-priority limitation identified: literal pool addressing range with low memory origins requires `.ltorg` directive support. Remaining work focuses on this fix, release engineering, CI/CD improvements, and optional enhancements.

**Estimated effort to v1.0.0:** 21-28 hours (includes `.ltorg` implementation)

---

## High Priority Tasks

### Literal Pool Addressing Range Limitation with Low Memory Origins
**Effort:** 6-8 hours
**Priority:** HIGH - Affects programs using `.org 0x0000` or low addresses with many constants

**Problem:**
Programs using `.org 0x0000` (or other low memory origins) with many `LDR Rd, =constant` pseudo-instructions can fail with "literal pool offset too large" errors. This is NOT a bug in the assembly programs - it's a limitation in the emulator's literal pool placement strategy.

**Root Cause:**
1. The emulator places all literal pool entries after code and data (`literalPoolStart = maxAddr`)
2. ARM's PC-relative addressing has a ±4095 byte range (12-bit offset)
3. With `.org 0x0000`, early instructions (e.g., PC=0x0008) may be >4095 bytes away from literals
4. Example: Code at 0x0000-0x1000, data at 0x1000-0x2000, literals start at 0x2000
   - Instruction at PC=0x0008 needs literal at 0x2000: offset = 0x2000 - 0x0008 - 8 = 8184 bytes
   - This exceeds ARM's 4095 byte maximum offset

**Examples Affected:**
- `bubble_sort.s` originally used `.org 0x0000` (now uses `.org 0x8000`)
- Any program with `.org <low address>` and multiple literal pool entries
- Large programs (>4KB code+data) using low memory origins

**Current Workaround:**
Programs use `.org 0x8000` to ensure adequate addressing range.

**Proposed Solution Options:**

**Option 1: Implement `.ltorg` Directive (Recommended)**
- Add parser support for `.ltorg` directive to force literal pool emission
- Allow programmers to manually place literal pools within 4095 bytes of usage
- Update encoder to support multiple literal pool regions
- Estimated effort: 6-8 hours

**Option 2: Automatic Literal Pool Insertion**
- Automatically insert literal pools when offset would exceed 4095 bytes
- Requires tracking forward references during encoding
- More complex implementation, may affect instruction addresses
- Estimated effort: 12-15 hours

**Option 3: Document Limitation (Current State)**
- Keep current implementation, document the limitation
- Recommend using `.org 0x8000` for programs with many literals
- Simplest but limits user flexibility

**Recommended:** Implement Option 1 (`.ltorg` directive) for compatibility with standard ARM assemblers.

**Files to Modify:**
- `parser/parser.go` - Add `.ltorg` directive parsing
- `parser/types.go` - Add `.ltorg` to directive types
- `encoder/memory.go` - Support multiple literal pool regions
- `main.go` - Handle `.ltorg` directive during program loading
- `tests/integration/literal_pool_test.go` - Add tests for `.ltorg`
- `docs/assembly_reference.md` - Document `.ltorg` directive

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

**Total estimated effort to v1.0.0:** 21-28 hours

- **High Priority:** 18-26 hours (`.ltorg` directive, CI/CD, release engineering)
- **Medium Priority:** 18-27 hours (coverage, benchmarking, documentation)
- **Low Priority:** Variable (optional enhancements)
