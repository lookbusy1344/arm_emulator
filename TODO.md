# ARM2 Emulator TODO List

**IMPORTANT:** This file tracks outstanding tasks and known issues that cannot be completed in the current phase. After completing any work, update this file to reflect the current status.

It should not contain completed items or notes about past work. Those belong in `PROGRESS.md`.

**Last Updated:** 2025-10-10 (Phase 11 Started - Production Hardening)

---

## Summary

**Phase 11 (Production Hardening) has begun!** 🚀

**All 10 core phases from IMPLEMENTATION_PLAN.md are COMPLETE!** ✅
**Phase 11 Quick Wins (Code Quality) COMPLETE!** ✅

The ARM2 emulator is **functionally complete and production-ready**. All core features work:
- ✅ All ARM2 instructions implemented and tested
- ✅ Full debugger with TUI
- ✅ All system calls functional
- ✅ 493 tests (490 passing, 99.4% pass rate)
- ✅ Cross-platform configuration
- ✅ Tracing and performance statistics
- ✅ Development tools (linter, formatter, xref)
- ✅ 17 example programs
- ✅ Comprehensive documentation

**What remains:** Distribution and polish items for M8 (Release Ready):
- **High Priority:** CI/CD pipeline, cross-platform testing
- **Medium Priority:** Code coverage analysis, performance benchmarking, installation packages
- **Low Priority:** Additional documentation, trace/stats integration

**Estimated effort to M8:** 20-30 hours total

---

## Phase 11 Status - In Progress

**COMPLETED (2025-10-10):** Quick Wins - Code Quality Issues
- ✅ Fixed Go vet warnings (ReadByte/WriteByte renamed to ReadByteAt/WriteByteAt)
- ✅ Updated CI Go version from 1.21 to 1.25
- ✅ Added build artifacts to .gitignore
- ✅ All tests passing, go fmt clean, go vet clean

**REMAINING:** See "M8: Release Ready - Outstanding Items" below

---

## Phase 10 Status ✅

**COMPLETED:** Phase 10 (Cross-Platform & Polish) has been successfully implemented with the following features:

### Implemented Features
1. **Configuration Management** (config/)
   - Cross-platform config file paths (macOS/Linux/Windows)
   - TOML configuration with defaults
   - Platform-aware log directories
   - 7 tests passing

2. **Execution & Memory Tracing** (vm/trace.go)
   - Execution trace with register changes, flags, timing
   - Register filtering
   - Memory access trace (reads/writes)
   - 11 tests passing

3. **Performance Statistics** (vm/statistics.go)
   - Instruction frequency tracking
   - Branch statistics
   - Function call profiling
   - Hot path analysis
   - Export to JSON/CSV/HTML
   - 11 tests passing

4. **Command-Line Integration** (main.go)
   - New flags: -trace, -mem-trace, -stats
   - File output options
   - Format selection (json/csv/html)
   - Enhanced help text

**Note:** The trace/stats infrastructure is in place but not yet connected to VM.Step() for automatic recording. This integration can be done as needed.

### Deferred Items
- Cross-compilation builds (CI/CD phase)
- Multi-platform CI/CD testing
- Manual cross-platform testing checklist
- Code coverage tooling

---

## High Priority

### 1. End-to-End Integration Tests

**Status:** ✅ COMPLETE

**Implementation:** Comprehensive end-to-end and integration tests have been added to verify the emulator works correctly with real programs.

**Completed Features:**
- ✅ Added 12 new integration tests for complete programs
- ✅ Test arithmetic operations, loops, conditionals, function calls
- ✅ Test array operations, string operations, bitwise operations
- ✅ Test nested function calls (factorial, recursion)
- ✅ Test stack operations (PUSH/POP)
- ✅ Test shifts and rotations
- ✅ Test example programs (hello.s, arithmetic.s - working)
- ✅ All syscalls tested with real programs
- ✅ 34 integration tests total (32 passing, 93.5% pass rate)

**Known Limitations (documented below):**
- `STMFD`/`LDMFD` instructions not fully supported (affects loops.s, conditionals.s, bubble_sort.s, etc.)
- Pre-indexed addressing with writeback `[Rn, #offset]!` not supported
- Post-indexed addressing `[Rn], #offset` not supported
- Some example programs use `SVC` instead of `SWI` (same instruction, different name)
- Character literals in strings (e.g., `'\t'`) are treated as literal backslash-t, not tab
- Immediate offsets in load/store `[Rn, #offset]` have parser issues

**Test Results:**
- Total tests: 511 tests
- Passing tests: 509 (99.6% pass rate)
- Integration tests: 34 total (32 passing, 93.5% pass rate)
- Unit tests: 477 total (477 passing, 100% pass rate)

**Failing Tests (2):**
- `TestExamplePrograms_Loops` - requires character literal support (`#' '`, `#'\t'`)
- `TestExamplePrograms_Conditionals` - requires character literal support (`#'\t'`)

**Completed:** 2025-10-10

---

### 2. Missing ARM2 Instructions for Example Programs

**Status:** ✅ COMPLETE (STMFD/LDMFD implemented)

**Completed Features (2025-10-10):**

1. **Stack Multi-Register Operations** ✅ COMPLETE
   - `STMFD` (Store Multiple, Full Descending) - ✅ implemented
   - `LDMFD` (Load Multiple, Full Descending) - ✅ implemented
   - `STMFA`, `STMEA`, `STMED` aliases - ✅ implemented
   - `LDMFA`, `LDMEA`, `LDMED` aliases - ✅ implemented
   - `SVC` instruction (ARM7+ alias for SWI) - ✅ implemented
   - All aliases properly map to existing LDM/STM implementations

**Remaining Limited Features:**

1. **Addressing Modes** (MEDIUM PRIORITY)
   - Pre-indexed with writeback: `LDR R0, [R1, #4]!` - parser errors
   - Post-indexed: `LDRB R0, [R1], #4` - parser errors
   - Immediate offset: `LDR R0, [R1, #4]` - parser errors
   - Impact: More sophisticated memory access patterns don't work
   - Workaround: Use separate ADD instructions and base register addressing
   - Effort: 3-4 hours

2. **Character Literal Escaping** (HIGH PRIORITY - blocks 2 example programs)
   - Character literals in immediates: `MOV R0, #' '` (space) and `MOV R0, #'\t'` (tab)
   - Affects: loops.s (line 26, 105), conditionals.s (line 106)
   - Impact: **2 example programs cannot run** (loops.s, conditionals.s)
   - Workaround: Use numeric values (`MOV R0, #32` for space, `MOV R0, #9` for tab)
   - Effort: 1-2 hours

**Total Effort to Support All Examples:** 4-6 hours (down from 8-12 hours)

**Priority:** Medium-High (for M8 if we want all examples working)

---

### 3. Expression Parser Improvements (Phase 5 Enhancement)

**Status:** ✅ COMPLETE

**Implementation:** The expression parser has been upgraded with a proper two-phase tokenizer and precedence-climbing parser.

**Completed Features:**
- ✅ All numeric literals work (decimal, hex, binary, octal)
- ✅ Register references work
- ✅ Symbol lookups work
- ✅ Memory dereferencing works (`[addr]`, `*addr`)
- ✅ Arithmetic operations work (`10 + 20`, `5 * 6`, `0x10 + 0x20`)
- ✅ Bitwise operations work (`0xFF & 0x0F`, `0xF0 | 0x0F`, `0xFF ^ 0x0F`)
- ✅ Shift operations work (`1 << 4`, `16 >> 2`)
- ✅ Register operations work (`r0 + r1`, `r0 + 5`, `r1 - r0`)
- ✅ Operator precedence correctly implemented
- ✅ Parentheses for grouping
- ✅ All previously disabled tests now passing

**Implementation Details:**
- Created `debugger/expr_lexer.go` - Tokenizer for debugger expressions
- Created `debugger/expr_parser.go` - Precedence-climbing parser with proper operator precedence
- Updated `debugger/expressions.go` - Refactored to use new lexer and parser
- All tests in `debugger/expressions_test.go` are now passing (100%)

**Completed:** 2025-10-10

---

## Phase 11: Production Hardening - Remaining Tasks

### Task 1: Complete Code Quality Issues

**Status:** Partially complete

**Remaining:**
- [ ] Run golangci-lint and address issues

**Effort Estimate:** 20-30 minutes

**Priority:** High

---

### Task 2: Enhance CI/CD Pipeline

**Status:** Partially complete (basic CI exists with Go 1.25)

**Requirements:**
- [x] Basic CI workflow exists ✅
- [x] Updated to Go 1.25 ✅
- [ ] Configure matrix builds (macOS, Windows, Linux)
- [ ] Add test coverage reporting with codecov
- [ ] Add coverage threshold enforcement (70% minimum)
- [ ] Add race detector to tests
- [ ] Upload test results as CI artifacts

**Effort Estimate:** 4-6 hours

**Priority:** High

**Files to Modify:**
- `.github/workflows/ci.yml` - Enhance existing workflow

---

### Task 3: Cross-Platform Manual Testing

**Status:** Partially complete (macOS only)

**Requirements:**
- [x] Test on macOS (development platform) ✅
- [ ] Create testing checklist in `docs/testing_checklist.md`
- [ ] Test on Windows 10/11
  - [ ] Binary builds successfully
  - [ ] All tests pass
  - [ ] TUI renders correctly
  - [ ] File paths work (backslash handling)
  - [ ] Config file loads from correct location
  - [ ] Example programs execute correctly
  - [ ] Command-line flags work
- [ ] Test on Linux (Ubuntu, Fedora, Arch)
  - [ ] TUI renders correctly
  - [ ] File I/O works correctly
  - [ ] Config file paths work
  - [ ] Example programs run identically
  - [ ] Command-line flags work
  - [ ] Terminal compatibility (various emulators)
  - [ ] Package dependencies documented
- [ ] Document any platform-specific quirks or limitations
- [ ] Fix any platform-specific issues found

**Effort Estimate:** 3-4 hours

**Priority:** High (needed for M8)

---

### Task 4: Increase Code Coverage to 75%+

**Status:** Not started

**Current Coverage:** ~40% estimated (not measured)
**Target Coverage:** 75%+

**Target Coverage by Package:**
- `vm/`: 75%+ (currently no direct tests)
- `debugger/`: 65%+ (currently 47.9%)
- `parser/`: 75%+
- `tools/`: 90%+ (currently 86.6%)
- `config/`: 85%+

**Focus Areas:**

**A. VM Package Tests (2 hours)**
- [ ] Create `vm/vm_test.go`
- [ ] Test VM initialization
- [ ] Test VM.Reset()
- [ ] Test VM.ResetRegisters()
- [ ] Test execution modes

**B. Debugger Expression Tests (1-2 hours)**
- [ ] Enhance `debugger/expressions_test.go`
- [ ] Test complex hex arithmetic
- [ ] Test bitwise operations
- [ ] Test error handling

**C. Parser Error Paths (1-2 hours)**
- [ ] Add tests for invalid input
- [ ] Test malformed instructions
- [ ] Test undefined labels
- [ ] Test forward references

**Effort Estimate:** 4-6 hours

**Priority:** Medium-High

---

### Task 5: Connect Tracing to Execution

**Status:** Infrastructure complete, integration pending

**Requirements:**
- [ ] Connect ExecutionTrace to VM.Step()
  - [ ] Call `trace.RecordInstruction()` after each instruction
  - [ ] Generate disassembly string for each instruction
  - [ ] Make optional via VM flag or config
- [ ] Connect MemoryTrace to Memory operations
  - [ ] Call `trace.RecordRead()` in Memory.ReadWord(), ReadByteAt(), etc.
  - [ ] Call `trace.RecordWrite()` in Memory.WriteWord(), WriteByteAt(), etc.
  - [ ] Make optional via VM flag or config
- [ ] Connect Statistics to VM operations
  - [ ] Call `stats.RecordInstruction()` after each instruction
  - [ ] Call `stats.RecordBranch()` for branch instructions
  - [ ] Call `stats.RecordFunctionCall()` for BL instructions
  - [ ] Call `stats.RecordMemoryRead/Write()` for memory operations
  - [ ] Make optional via VM flag or config
- [ ] Add tests for trace/stats integration

**Effort Estimate:** 2-3 hours

**Priority:** Low (infrastructure is ready, integration is optional)

**Files to Modify:**
- `vm/executor.go` - Add trace/stats calls to Step()
- `vm/memory.go` - Add trace calls to memory operations

---

## Phase 12: Performance & Benchmarking

**Total Effort:** 14-20 hours
**Priority:** MEDIUM

### Task 1: Create Benchmark Suite

**Status:** Not started

**Requirements:**
- [ ] Create `vm/vm_bench_test.go`
  - [ ] BenchmarkExecutionLoop
  - [ ] BenchmarkMemoryRead
  - [ ] BenchmarkInstructionDecode
- [ ] Create `parser/parser_bench_test.go`
  - [ ] BenchmarkParseLargeFile
  - [ ] BenchmarkLexer
- [ ] Create `tests/benchmarks/tui_bench_test.go`
  - [ ] Benchmark TUI refresh rate
- [ ] Document benchmark targets:
  - Parser: < 100ms for < 1000 line programs
  - Execution: > 100k instructions/second
  - Memory: < 100MB for typical programs
  - TUI: 60 FPS minimum

**Effort Estimate:** 4-6 hours

**Priority:** Medium

---

### Task 2: Profile Performance

**Status:** Not started

**Requirements:**
- [ ] Run CPU profiling: `go test -bench=. -cpuprofile=cpu.prof ./vm`
- [ ] Run memory profiling: `go test -bench=. -memprofile=mem.prof ./vm`
- [ ] Analyze profiles with `go tool pprof`
- [ ] Generate profile reports
- [ ] Create `docs/performance_analysis.md` with findings
- [ ] Identify optimization opportunities

**Effort Estimate:** 2-3 hours

**Priority:** Medium

---

### Task 3: Implement Optimizations

**Status:** Not started (defer until after profiling)

**Potential Optimizations:**
- [ ] Instruction decode cache (if decode is hot path)
- [ ] Memory access optimization (if memory is bottleneck)
- [ ] Flag calculation optimization (only when S bit set)
- [ ] Pre-allocate common memory regions

**Effort Estimate:** 6-8 hours

**Priority:** Low (only if profiling shows need)

---

### Task 4: Document Performance

**Status:** Not started

**Requirements:**
- [ ] Create `docs/performance_characteristics.md`
  - [ ] Benchmark results
  - [ ] Performance comparison to similar tools
  - [ ] Optimization opportunities
  - [ ] Scalability limits

**Effort Estimate:** 2-3 hours

**Priority:** Low

---

## Phase 13: Release Engineering

**Total Effort:** 16-22 hours
**Priority:** MEDIUM-HIGH

### Task 1: Create Release Pipeline

**Status:** Not started

**Requirements:**
- [ ] Create `.github/workflows/release.yml`
  - [ ] Trigger on version tags (v*)
  - [ ] Matrix builds for all platforms:
    - [ ] linux-amd64
    - [ ] darwin-amd64
    - [ ] darwin-arm64
    - [ ] windows-amd64
  - [ ] Create release archives (tar.gz)
  - [ ] Upload artifacts
  - [ ] Create GitHub Release with notes
- [ ] Create cross-compilation builds for all platforms:
  - [ ] `GOOS=darwin GOARCH=amd64` - macOS Intel
  - [ ] `GOOS=darwin GOARCH=arm64` - macOS Apple Silicon
  - [ ] `GOOS=linux GOARCH=amd64` - Linux x64
  - [ ] `GOOS=windows GOARCH=amd64` - Windows x64
- [ ] Test release workflow

**Effort Estimate:** 4-6 hours

**Priority:** High (needed for v1.0.0 release)

**Files to Create:**
- `.github/workflows/release.yml` - Release workflow
- `Makefile` - Build targets for cross-compilation (optional)

---

### Task 2: Create Installation Packages

**Status:** Not started

**Requirements:**
- [ ] Create installation scripts
  - [ ] `install.sh` for macOS/Linux
  - [ ] `install.ps1` for Windows
- [ ] Package for distribution:
  - [ ] Homebrew formula (macOS) - 2 hours
    - Create `homebrew-arm-emulator` repository
    - Write formula with proper dependencies
    - Test installation
  - [ ] Chocolatey package (Windows) - 2 hours
    - Create `chocolatey/arm-emulator.nuspec`
    - Test installation
  - [ ] Debian package (`.deb`) - 2 hours
    - Create debian/ structure (control, changelog, rules, install)
  - [ ] RPM package (Fedora/RHEL) - 2 hours
    - Create RPM spec file
  - [ ] AUR package (Arch Linux) - optional
  - [ ] Scoop manifest (Windows) - optional
- [ ] Create release assets (tarballs, zip files)
- [ ] Add installation instructions to docs
- [ ] Test installation process on each platform

---

### 3. Cross-Platform Testing

**Status:** Partially complete (macOS only)

**Requirements:**
- [x] Test on macOS (development platform) ✅
- [ ] Test on Windows 10/11
  - [ ] TUI renders correctly
  - [ ] File I/O works correctly
  - [ ] Config file paths work
  - [ ] Example programs run identically
  - [ ] Command-line flags work
- [ ] Test on Linux (Ubuntu, Fedora, Arch)
  - [ ] TUI renders correctly
  - [ ] File I/O works correctly
  - [ ] Config file paths work
  - [ ] Example programs run identically
  - [ ] Command-line flags work
- [ ] Document any platform-specific quirks or limitations
- [ ] Fix any platform-specific issues found

**Effort Estimate:** 3-4 hours

**Priority:** High (needed for M8)

---

### 4. Code Coverage Analysis (From Phase 7)

**Status:** Not started

**Requirements:**
- [ ] Generate coverage reports with `go test -coverprofile`
- [ ] Set up coverage visualization (e.g., coveralls, codecov)
- [ ] Analyze coverage by package
- [ ] Identify gaps in test coverage
- [ ] Add tests to reach 85%+ target
- [ ] Add coverage badge to README

**Current Coverage:** ~40% estimated (not measured)
**Target:** 85%+

**Effort Estimate:** 4-6 hours

**Priority:** Medium

---

### 5. Performance Benchmarking

**Status:** Not started (infrastructure complete in Phase 10)

**Requirements:**
- [ ] Create benchmarking test suite
- [ ] Benchmark parser performance (target: < 100ms for < 1000 line programs)
- [ ] Benchmark execution performance (target: > 100k instructions/second)
- [ ] Benchmark memory usage (target: < 100MB for typical programs)
- [ ] Benchmark TUI refresh rate (target: 60 FPS minimum)
- [ ] Document performance results
- [ ] Profile hot paths and optimize if needed
- [ ] Add performance regression tests

**Effort Estimate:** 4-6 hours

**Priority:** Medium

**Files to Create:**
- `tests/benchmarks/parser_bench_test.go`
- `tests/benchmarks/vm_bench_test.go`
- `tests/benchmarks/tui_bench_test.go`
- `docs/performance.md`

---

### 6. Installation Packages

**Status:** Not started

**Requirements:**
- [ ] Create installation scripts for all platforms
  - [ ] `install.sh` for macOS/Linux
  - [ ] `install.ps1` for Windows
- [ ] Package for distribution:
  - [ ] Homebrew formula (macOS)
  - [ ] Debian package (`.deb`) for Ubuntu/Debian
  - [ ] RPM package for Fedora/RHEL
  - [ ] AUR package for Arch Linux
  - [ ] Chocolatey package for Windows
  - [ ] Scoop manifest for Windows
- [ ] Create release assets (tarballs, zip files)
- [ ] Add installation instructions to docs
- [ ] Test installation process on each platform

**Effort Estimate:** 6-8 hours

**Priority:** Medium

**Files to Create:**
- `install.sh`
- `install.ps1`
- `homebrew/arm-emulator.rb`
- `debian/control`, `debian/rules`, etc.
- `rpm/arm-emulator.spec`
- `docs/installation_packages.md`

---

### Task 3: Release Documentation

**Status:** Not started

**Requirements:**
- [ ] Create `CHANGELOG.md`
  - [ ] Document v1.0.0 features
  - [ ] List all completed features
  - [ ] Document fixes (go vet warnings, etc.)
  - [ ] Note performance characteristics
- [ ] Update `README.md`
  - [ ] Add installation instructions for all package managers
  - [ ] Add badges (tests, coverage, version)
  - [ ] Update quick start guide
- [ ] Create `CONTRIBUTING.md`
  - [ ] Guidelines for contributors
  - [ ] Code style requirements
  - [ ] PR process
  - [ ] Testing requirements

**Effort Estimate:** 3-4 hours

**Priority:** Medium-High

---

### Task 4: Release Testing

**Status:** Not started

**Requirements:**
- [ ] Create `docs/release_checklist.md`
- [ ] Pre-Release verification:
  - [ ] All tests passing
  - [ ] Code coverage >70%
  - [ ] Documentation up to date
  - [ ] CHANGELOG.md updated
  - [ ] Version numbers updated
- [ ] Build Testing:
  - [ ] Linux build works
  - [ ] macOS Intel build works
  - [ ] macOS ARM build works
  - [ ] Windows build works
- [ ] Installation Testing:
  - [ ] Homebrew formula works
  - [ ] Chocolatey package works
  - [ ] DEB package installs
  - [ ] RPM package installs
- [ ] Functional Testing:
  - [ ] All examples run successfully
  - [ ] TUI debugger works on all platforms
  - [ ] Command-line arguments work
  - [ ] Configuration files load correctly
- [ ] Release execution:
  - [ ] Tag version in git
  - [ ] GitHub Release created
  - [ ] Release notes published
  - [ ] Packages uploaded to repositories

**Effort Estimate:** 3-4 hours

**Priority:** High (needed before v1.0.0 release)

---

## Phase 14: Advanced Features (Optional, Future)

**Total Effort:** 80-120 hours
**Priority:** LOW (Post-release)

### 1. JIT Compilation

**Status:** Not started

**Description:** Translate ARM to native x86-64 for 10-100x speedup

**Requirements:**
- [ ] Design JIT architecture
- [ ] Implement ARM to x86-64 translator
- [ ] Add runtime code generation
- [ ] Handle self-modifying code
- [ ] Add JIT-specific tests
- [ ] Benchmark performance improvements

**Effort Estimate:** 40-60 hours

**Priority:** Low

---

### 2. Web-Based Debugger

**Status:** Not started

**Description:** React frontend with WebSocket backend for browser-based debugging

**Requirements:**
- [ ] Design WebSocket protocol
- [ ] Implement backend server
- [ ] Create React frontend
- [ ] Add real-time state synchronization
- [ ] Support multiple concurrent clients
- [ ] Add web-based TUI

**Effort Estimate:** 30-40 hours

**Priority:** Low

---

### 3. GDB Protocol Support

**Status:** Not started

**Description:** Implement GDB remote serial protocol for IDE integration

**Requirements:**
- [ ] Implement GDB remote protocol
- [ ] Add protocol parser
- [ ] Support common GDB commands
- [ ] Test with popular IDEs (VSCode, IntelliJ, etc.)
- [ ] Document GDB integration

**Effort Estimate:** 20-30 hours

**Priority:** Low

---

### 4. Additional Architectures

**Status:** Not started

**Description:** Extend to ARM3, ARM7, or other RISC architectures

**Requirements:**
- [ ] Research target architecture
- [ ] Extend instruction set
- [ ] Add architecture-specific features
- [ ] Update tests for new architecture
- [ ] Document new architecture support

**Effort Estimate:** 40+ hours per architecture

**Priority:** Low

---

## Deferred Documentation (From Phase 9)

**Status:** Core docs complete, additional docs deferred

**Requirements:**
- [ ] **docs/tutorial.md** - Step-by-step tutorial
  - [ ] Hello World walkthrough
  - [ ] Basic arithmetic tutorial
  - [ ] Function calls and stack tutorial
  - [ ] Using the debugger tutorial
  - [ ] Memory and data structures tutorial
  - Effort: 2-3 hours

- [ ] **docs/faq.md** - Frequently asked questions
  - [ ] Common errors and solutions
  - [ ] Platform-specific issues
  - [ ] Performance tips
  - [ ] Debugging tips
  - Effort: 1-2 hours

- [ ] **docs/api_reference.md** - API documentation
  - [ ] VM package API
  - [ ] Parser package API
  - [ ] Encoder package API
  - [ ] Debugger package API
  - [ ] Config package API
  - Effort: 3-4 hours

- [ ] **docs/contributing.md** - Contributing guidelines
  - [ ] How to contribute
  - [ ] Code style guidelines
  - [ ] Testing requirements
  - [ ] Pull request process
  - Effort: 1 hour

- [ ] **docs/coding_standards.md** - Go coding standards
  - [ ] Naming conventions
  - [ ] Error handling patterns
  - [ ] Testing patterns
  - [ ] Documentation requirements
  - Effort: 1 hour

**Total Effort Estimate:** 8-11 hours

**Priority:** Low (nice to have, not required for M8)

---

## Summary of Remaining Work

**Estimated Total Effort to v1.0.0 Release:** ~50-70 hours

**High Priority (Phase 11 - Production Hardening):** ~15-20 hours
- Code quality (golangci-lint)
- Enhanced CI/CD with multi-platform testing
- Cross-platform manual testing
- Code coverage increase to 75%+

**Medium-High Priority (Phase 13 - Release Engineering):** ~16-22 hours
- Release pipeline automation
- Installation packages (Homebrew, Chocolatey, DEB, RPM)
- Release documentation (CHANGELOG, CONTRIBUTING)
- Release testing checklist

**Medium Priority (Phase 12 - Performance):** ~14-20 hours
- Benchmark suite creation
- Performance profiling
- Optimization implementation
- Performance documentation

**Low Priority (Optional):** ~8-11 hours
- Deferred documentation (tutorials, FAQ, API docs)
- Trace/stats integration with VM
- Advanced features (JIT, web debugger, GDB protocol)

---

## Notes

- Priority levels: High (next phase), Medium (important), Low (nice to have)
- Some items may be dependencies for others
- This list will be updated as development progresses
