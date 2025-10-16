# ARM2 Emulator TODO List

**Last Updated:** 2025-10-16

This file tracks outstanding work only. Completed items are in `PROGRESS.md`.

---

## Summary

**Current Phase:** Phase 11 (Production Hardening) - Core Complete + ARMv3 Extensions

**Status:** Project is production-ready with comprehensive test coverage and all critical features implemented. All planned ARMv3/ARMv3M instruction extensions have been completed (long multiply, PSR transfer, NOP, LDR Rd, =value). Register access pattern analysis diagnostic mode has been added. Remaining work focuses on optional enhancements and improvements.

**Test Status:** 1185+ tests, 100% pass rate, 0 lint issues, 75.0% code coverage

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

---

## Medium Priority Tasks

### Additional Diagnostic Modes

**Currently Implemented (Phase 11):**
- ✅ Code Coverage (`--coverage`) - Track executed vs unexecuted instructions
- ✅ Stack Trace (`--stack-trace`) - Monitor stack operations, detect overflow/underflow
- ✅ Flag Trace (`--flag-trace`) - Record CPSR flag changes (N, Z, C, V)
- ✅ Execution Trace (`--trace`) - Basic instruction execution logging
- ✅ Memory Trace (`--mem-trace`) - Memory access tracking
- ✅ Register Access Pattern Analysis (`--register-trace`) - Track read/write patterns per register, identify hot registers, detect unused registers, read-before-write detection
- ✅ Symbol-Aware Trace Output - All diagnostic traces show function/label names (e.g., `main+4`, `calculate`) instead of raw hex addresses

**Proposed Extensions:**
- [ ] **Data Flow Tracing** (6-8 hours) - Track data movement between registers/memory, value provenance, data dependency tracking, taint analysis
- [ ] **Cycle-Accurate Timing Simulation** (8-10 hours) - Estimate ARM2 instruction timing, pipeline stall simulation, memory access latency, performance bottleneck identification
- [ ] **Memory Region Heatmap Visualization** (4-6 hours) - Track access frequency per region, HTML/graphical output, color-coded visualization
- [ ] **Reverse Execution Log** (10-12 hours) - Record state for backwards stepping, circular buffer of previous N instructions, time-travel debugging

### ✅ ~~Symbol-Aware Trace Output~~ - COMPLETED (2025-10-16)
**Actual Effort:** ~2 hours

- [x] Created `SymbolResolver` for address-to-symbol lookups with offset calculation
- [x] Enhanced ExecutionTrace, MemoryTrace, StackTrace, FlagTrace, RegisterTrace with symbol support
- [x] Updated main.go to load symbols into all trace modules
- [x] Created comprehensive unit tests (22 tests) for SymbolResolver
- [x] Verified all traces show function/label names in output

**Result:** All diagnostic traces now display human-readable symbol names (e.g., `main+4`, `calculate`, `helper1+20`) instead of raw hex addresses. This dramatically improves debugging experience. The `SymbolResolver` provides efficient binary search-based address resolution with offset calculation. All existing tests pass with zero lint issues.

### ✅ ~~Re-enable TUI Automated Tests~~ - COMPLETED (2025-10-16)
**Actual Effort:** ~2 hours

- [x] Refactor `NewTUI` to accept optional `tcell.Screen` parameter for testing
- [x] Use `tcell.SimulationScreen` in tests instead of real terminal
- [x] Re-enable `tui_manual_test.go.disabled` as `tui_test.go`
- [x] Verify tests run without hanging in CI/CD

**Result:** 18 TUI tests implemented and passing successfully using `tcell.SimulationScreen`. Created `NewTUIWithScreen()` function that accepts an optional screen parameter, allowing tests to inject a simulation screen while production code uses the default screen. All tests run without hanging. Tests are located in `tests/unit/debugger/tui_test.go` following project structure conventions.

### ✅ ~~Code Coverage Improvements~~ - COMPLETED (2025-10-16)
**Actual Effort:** ~4 hours

**Current:** 75.0% (achieved) | **Target:** 75%+

Completed areas:
- [x] VM package tests (initialization, reset, execution modes)
- [x] VM CPU trace tests (register tracing, condition codes)
- [x] VM memory helper tests (byte operations, CPSR serialization)
- [x] Parser error path tests (error handling, warnings)
- [x] Parser macro tests (definition, expansion, substitution)
- [x] Parser preprocessor tests (symbol management, file processing)
- [x] Parser lexer tests (token string formatting)
- [x] Parser symbol tests (relocations, address updates)

**Result:** Achieved exactly 75.0% code coverage (up from 71.7%). Added 105 new tests across 8 test files covering previously untested functions in VM, parser, and debugger packages.

### Performance & Benchmarking
**Effort:** 10-15 hours

- [ ] Create benchmark tests (VM, parser, TUI)
- [ ] Document performance targets
- [ ] Run CPU and memory profiling
- [ ] Create `docs/performance_analysis.md`
- [ ] Implement optimizations if needed

### ✅ ~~Additional Documentation~~ - COMPLETED (2025-10-16)
**Actual Effort:** ~4 hours

- [x] Tutorial guide (step-by-step walkthrough) - `docs/TUTORIAL.md`
- [x] FAQ (common errors, platform issues, tips) - `docs/FAQ.md`
- [x] API reference documentation - `docs/API.md`

**Result:** Created comprehensive documentation for users and developers:
- **TUTORIAL.md**: Complete step-by-step guide covering basics through advanced topics with hands-on examples
- **FAQ.md**: Extensive FAQ with 50+ questions covering common errors, troubleshooting, tips, and platform-specific issues
- **API.md**: Full API reference for Go package interfaces with usage examples for programmatic use
- Updated README.md and installation.md to reference new documentation

---

## Low Priority Tasks (Optional)

### Later ARM Architecture Extensions (Optional)

These are **not** part of ARM2 but could be added for broader compatibility:

**ARMv2a Atomic Operations** (Effort: 4-6 hours)
- [ ] SWP, SWPB - Atomic swap operations

**ARMv2 Coprocessor Interface** (Effort: 20-30 hours)
- [ ] CDP, LDC, STC, MCR, MRC - Coprocessor operations
- [ ] Full coprocessor emulation framework

**Note:** The emulator has complete ARM2 instruction set coverage. All planned ARMv3/ARMv3M extensions have been completed. These remaining extensions are from later architectures.

### Enhanced CI/CD Pipeline (Optional)
**Effort:** 4-6 hours

- [ ] Configure matrix builds (macOS, Windows, Linux)
- [ ] Add test coverage reporting (codecov)
- [ ] Add coverage threshold enforcement (70% minimum)
- [ ] Add race detector to tests
- [ ] Upload test results as CI artifacts

### Release Engineering (Optional)
**Effort:** 8-12 hours

- [ ] Create `.github/workflows/release.yml` with matrix builds
  - linux-amd64, darwin-amd64, darwin-arm64, windows-amd64
- [ ] Create release archives and GitHub Release automation
- [ ] Create `CHANGELOG.md`
- [ ] Create `CONTRIBUTING.md`
- [ ] Create `docs/release_checklist.md`
- [ ] Pre-release verification checklist


---

## Effort Summary

- **High Priority:** All completed ✅
  - ~~Planned instruction extensions: 20-30 hours~~ ✅ COMPLETED
  - ~~Code coverage improvements: 4-6 hours~~ ✅ COMPLETED
- **Medium Priority:** 10-15 hours (benchmarking, documentation)
- **Low Priority:** Variable (optional enhancements, CI/CD, release engineering)

**Completed in this session (2025-10-16):**
- ARMv3/ARMv3M instruction extensions (~6 hours actual vs 20-30 hours estimated)
- Code coverage improvements to 75.0% (~4 hours actual vs 4-6 hours estimated)
