# ARM2 Emulator TODO List

**IMPORTANT:** This file tracks outstanding tasks and known issues. After completing any work, update this file to reflect the current status.

Completed items and past work belong in `PROGRESS.md`.

**Last Updated:** 2025-10-13

---

## Summary

**Status:** Phase 11 (Production Hardening) - All Test Priorities Complete ✅

**Test Status:** 1040/1040 tests passing (100% ✅)
- Baseline tests: 660 passing
- Priority 1-5 tests: 356 added (all passing)
- Tool tests: 24 added (including 4 new standalone label tests)
- Comprehensive ARM2 instruction coverage achieved

**Example Programs:** 22 of 30 fully functional (73% functional rate)

**Recent Completion (2025-10-12):**
- ✅ **All Test Priorities Complete:** 1016/1016 tests passing
  - Priority 1: Critical tests (LDRH/STRH, BX, conditionals) - 24 tests ✅
  - Priority 2: Memory addressing modes - 35 tests ✅
  - Priority 3: Data processing with register shifts - 56 tests ✅
  - Priority 4: Edge cases (special registers, flags, immediates) - 65 tests ✅
  - Priority 5: Instruction-condition matrix - 160 tests ✅
  - **See MISSING_TESTS.md for full details**

**Recent Additions (2025-10-13):**
- ✅ **Formatter and XRef Tools - Standalone Labels Fixed:** Tools now correctly handle labels on their own line
  - **Status:** FIXED - Both formatter and xref tools properly process standalone labels
  - Bug: Formatter was outputting all standalone labels at beginning of file instead of in source order
  - Bug: Both tools were collecting standalone labels but not interleaving them properly with instructions/directives
  - Fix: Modified `tools/format.go` to interleave standalone labels from symbol table in proper source order based on line numbers
  - XRef was already working correctly but added comprehensive tests to verify
  - 4 new comprehensive unit tests added:
    - `TestFormat_StandaloneLabel` - single standalone label positioning
    - `TestFormat_MultipleStandaloneLabels` - multiple standalone labels in order
    - `TestXRef_StandaloneLabel` - xref tracking of standalone label
    - `TestXRef_MultipleStandaloneLabels` - xref tracking of multiple standalone labels
  - All 1040 tests passing (64 tool tests, 976 other tests)
- ✅ **Standalone Label Parser Bug Fixed:** Parser now correctly handles labels on their own line
  - **Status:** FIXED - All labels now parsed correctly
  - Bug: When a standalone label (nothing after it on the line) was followed by another labeled directive, the second label would be misparsed as an instruction
  - Fix: Removed premature `skipNewlines()` call after label processing in parser/parser.go
  - 7 new comprehensive unit tests added (tests/unit/parser/space_directive_test.go)
  - All 1023 tests passing
- ✅ **Constant Expression Support:** Parser and encoder now support arithmetic expressions in pseudo-instructions
  - **Status:** WORKING - Parsing, evaluation, and execution all functional
  - Syntax: `LDR r0, =label + 12`, `LDR r1, =symbol - 4`, `LDR r2, =0x8000 + 8`
  - Supports addition and subtraction with immediate values (decimal and hex)
  - Unit tests added and passing (tests/unit/parser/constant_expressions_test.go)
  - Real-world validation: `LDR r0, =buffer` followed by `LDR r1, =buffer + 4` correctly evaluates
- ✅ **Division Example:** Added examples/division.s demonstrating software division (ARM2 lacks hardware division)
  - Implements division by repeated subtraction
  - Tests multiple edge cases (exact division, dividend < divisor, etc.)
  - Uses proper syscall conventions (WRITE_INT with base parameter)

**Remaining Work:**
- **High Priority:** CI/CD enhancements (matrix builds, coverage reporting)
- **Medium Priority:** Code coverage improvements, release pipeline
- **Low Priority:** Performance benchmarking, additional documentation

**Estimated effort to v1.0.0:** 20-30 hours

---

## Known Issues

### Example Program Issues (Non-Critical)

**Test Summary (30 programs total):**
- ✅ **22 programs fully working** (73%)
- ⚠️ **3 programs with input issues** (10%)
- ❌ **5 programs with errors** (17%)

#### Working Programs (22)
1. ✅ addressing_modes.s - All addressing mode tests passed
2. ✅ arithmetic.s - All arithmetic operations work correctly
3. ✅ arrays.s - Array operations demo works
4. ✅ binary_search.s - Binary search works correctly
5. ✅ bit_operations.s - All bit operation tests passed
6. ✅ conditionals.s - All conditional execution tests passed
7. ✅ factorial.s - Factorial calculation works
8. ✅ fibonacci.s - Fibonacci sequence generation works
9. ✅ functions.s - Function calling conventions work
10. ✅ gcd.s - GCD calculation works correctly
11. ✅ hello.s - Hello world works
12. ✅ linked_list.s - Linked list operations work
13. ✅ loops.s - All loop constructs work correctly
14. ✅ memory_stress.s - All memory tests passed
15. ✅ nested_calls.s - Deep nested calls work correctly
16. ✅ recursive_factorial.s - Recursive factorial works
17. ✅ stack.s - Stack-based calculator works
18. ✅ string_reverse.s - String reversal works
19. ✅ strings.s - String operations work
20. ✅ times_table.s - Times table generation works

#### Programs with Input Issues (3)
21. ⚠️ **bubble_sort.s** - Expects interactive input, runs but shows zeros with stdin input
22. ⚠️ **calculator.s** - Input reading issue (operation character not read correctly, infinite loop)

#### Programs with Errors (5)
23. ❌ **hash_table.s** - Parse error: "invalid constant value: -" at line 10
   - Parser doesn't support "-" as a constant value placeholder

24. ❌ **matrix_multiply.s** - Runtime error: memory access violation at 0x000081D4
   - Attempts to read string at invalid address 0x00000002

25. ❌ **quicksort.s** - Runtime error: memory access violation at 0x000081EC
   - Attempts to read string at invalid address 0x00000011

26. ❌ **recursive_fib.s** - Parse errors: multiple syntax issues
   - Contains '@' characters (comments?) and parentheses in unexpected places
   - Appears to use non-ARM2 syntax

27. ❌ **reverse_chatgpt.s** - Parse errors: unexpected NUMBER tokens (lines 8, 13, 25, 32, 37)
   - Syntax issues preventing parsing

28. ❌ **sieve_of_eratosthenes.s** - Parse errors: extensive syntax issues
   - Contains '@' characters, parentheses, operators in unexpected places
   - Appears to use non-ARM2 syntax

29. ❌ **state_machine.s** - Runtime error: cycle limit exceeded (1000000 cycles)
   - Program enters infinite loop during email validation

30. ❌ **xor_cipher.s** - Encoding error: unknown instruction "LSR" at 0x00008220
   - LSR instruction not implemented in encoder

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

**Total estimated effort to v1.0.0:** 20-30 hours

**By Priority:**
- **High:** 10-16 hours - CI/CD enhancements, code coverage, release pipeline
- **Medium:** 14-20 hours - Performance benchmarking
- **Low (Optional):** 10-15 hours - Future enhancements, additional documentation

---

## Recently Completed (See PROGRESS.md for details)

- ✅ All test priorities complete (1016/1016 tests passing)
- ✅ Priority 1-5 tests: 356 new tests added and passing
- ✅ Comprehensive ARM2 instruction coverage achieved
- ✅ Halfword detection bug fixed (literal pool loads)
- ✅ Integer conversion issues fixed (gosec G115 warnings)
- ✅ ARM immediate encoding rotation bug fixed (fibonacci.s, calculator.s)
- ✅ Memory trace integration completed
- ✅ Diagnostic modes implemented (code coverage, stack trace, flag trace)
- ✅ CLI diagnostic flags with integration tests
- ✅ All lint issues resolved (golangci-lint clean)
- ✅ Go vet warnings fixed (method renames)
- ✅ CI updated to Go 1.25
- ✅ Parser limitations resolved (debugger expression parser rewritten)
