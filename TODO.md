## CRITICAL REGRESSION: Example Programs Broken After 715390c5bb9143f75d87eebf7b166f25f55e57b7

- Multiple example programs that previously worked in commit 17cafc7b8c79bce40ced5fa5ae410bd11ceff04b now fail with memory access violations or unmapped PC errors.
- Affected: fibonacci.s, calculator.s, factorial.s, bubble_sort.s, and likely others.
- Symptoms: PC jumps to unmapped memory (e.g., PC=0x000000D8 not mapped), stack/return address issues, program halts unexpectedly.
- Root cause is likely a regression in control flow, stack handling, or section mapping logic introduced in 715390c5bb9143f75d87eebf7b166f25f55e57b7.
- **Priority: CRITICAL** — Must bisect, diagnose, and fix before further development.
- Do not modify example programs to "work around" this; fix the emulator logic.
- See also: Known Issues section and integration test failures for details.
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

**Example Programs:** 15 of 35 fully functional (43% functional rate)
- 15 working programs ✅
- 17 failing programs ❌
- 1 timeout (infinite loop) ⏱️
- 2 hanging (awaiting input) ⌛

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

### Example Program Issues CRITICAL - Must Fix Before Further Development

Something broke several example programs in 715390c5bb9143f75d87eebf7b166f25f55e57b7, they worked in 17cafc7b8c79bce40ced5fa5ae410bd11ceff04b. They continue to fail in the current branch. Eg: fibonacci.s, calculator.s, factorial.s, bubble_sort.s

**Test Summary (35 programs total):**
- ✅ **15 programs fully working** (43%)
- ❌ **17 programs with runtime errors** (49%) - Memory access violations
- ⏱️ **1 program timeout** (3%) - Infinite loop
- ⌛ **2 programs hang awaiting input** (6%) - Need stdin piping

#### Working Programs (15/35)
1. ✅ addressing_modes.s - All addressing mode tests passed
2. ✅ arithmetic.s - All arithmetic operations work correctly
3. ✅ arrays.s - Array operations demo works
4. ✅ binary_search.s - Binary search works correctly
5. ✅ bit_operations.s - All bit operation tests passed
6. ✅ conditionals.s - All conditional execution tests passed
7. ✅ const_expressions.s - Constant expression evaluation works
8. ✅ functions.s - Function calling conventions work
9. ✅ hello.s - Hello world works
10. ✅ linked_list.s - Linked list operations work
11. ✅ loops.s - All loop constructs work correctly
12. ✅ memory_stress.s - All memory tests passed
13. ✅ nested_calls.s - Deep nested calls work correctly
14. ✅ recursive_factorial.s - Recursive factorial works
15. ✅ stack.s - Stack-based calculator works
16. ✅ strings.s - String operations work
17. ✅ test_expr.s - Expression evaluation tests pass

#### Programs with Runtime Errors - Memory Access Violations (17/35)
All of these fail with "memory access violation: address 0xXXXX is not mapped" - the PC attempts to execute from an unmapped memory region, suggesting control flow issues (bad function returns, stack corruption, or missing code sections).

18. ❌ **bubble_sort.s** - PC=0x00000148 not mapped
19. ❌ **calculator.s** - PC=0x000001BC not mapped
20. ❌ **division.s** - PC=0x00000188 not mapped
21. ❌ **factorial.s** - PC=0x000000A4 not mapped
22. ❌ **fibonacci.s** - PC=0x000000D8 not mapped
23. ❌ **matrix_multiply.s** - PC=0x000001D0 not mapped
24. ❌ **quicksort.s** - PC=0x000001E8 not mapped
25. ❌ **standalone_labels.s** - PC=0x00000004 not mapped
26. ❌ **state_machine.s** - PC=0x0000025C not mapped
27. ❌ **string_reverse.s** - PC=0x000000B4 not mapped
28. ❌ **test_const_expr.s** - PC=0x00000058 not mapped (after register dump)
29. ❌ **times_table.s** - PC=0x00000060 not mapped

#### Programs with Parse/Encoding Errors (5/35)
30. ❌ **hash_table.s** - Parse error: "invalid constant value: -" at line 10
   - `.equ EMPTY_KEY, -1` not supported - parser rejects negative constant values

31. ❌ **recursive_fib.s** - Parse errors: unexpected '@' characters throughout
   - Uses GNU assembler syntax with '@' for comments instead of ';'
   - Also has syntax issues with parentheses and other non-ARM2 constructs

32. ❌ **reverse_chatgpt.s** - Parse errors: unexpected NUMBER tokens (lines 8, 13, 25, 32, 37)
   - Uses AREA directive and & prefix for syscall numbers (Acorn/RISC OS syntax)
   - Not compatible with current parser

33. ❌ **sieve_of_eratosthenes.s** - Parse errors: extensive syntax issues
   - Uses GNU assembler syntax with '@' for comments
   - Multiple syntax incompatibilities

34. ❌ **xor_cipher.s** - Encoding error: "unknown instruction: LSR" at 0x00008220
   - LSR instruction not implemented in encoder (logical shift right)
   - This is a standard ARM2 instruction that should be supported

#### Programs with Infinite Loop (1/35)
35. ⏱️ **gcd.s** - Timeout after 2 seconds (exit code 124)
   - Program enters infinite loop, exceeds execution time limit

#### Programs Awaiting Input (2/35)
Note: These programs may work correctly with proper stdin input but cannot be tested non-interactively.
- ⌛ Programs that require user input but hang when run without stdin
- May work correctly if piped with appropriate input data

---

**Analysis Summary:**

The majority of failures (17/35 = 49%) are memory access violations where the PC jumps to unmapped memory regions. This pattern suggests:
1. **Stack/Return Address Issues** - Functions may not be preserving/restoring LR correctly
2. **Missing Code Sections** - Programs may be organized with `.text`/`.data` sections that aren't being loaded properly
3. **Incorrect Branch Targets** - Label resolution or branch encoding issues causing jumps to wrong addresses

Key actionable issues identified:
- **LSR/LSL/ASR/ROR as standalone instructions** - Currently only supported as shift modifiers (e.g., `MOV r0, r1, LSR #2`), not as standalone instructions (e.g., `LSR r0, r1, #2`). Affects xor_cipher.s
- **Negative constants not supported** - Parser rejects `.equ SYMBOL, -1` (affects hash_table.s)
- **Alternative syntax support** - Several programs use GNU assembler or RISC OS syntax (@ comments, AREA directive, & syscalls)

---

## Outstanding Tasks

### Example Program Fixes
**Priority:** Medium | **Effort:** 8-12 hours

**Important Note:** The low success rate (43%) was not detected earlier because **only 4 of 35 example programs have automated tests**:
- `hello.s` ✅ (TestExamplePrograms_Hello)
- `arithmetic.s` ✅ (TestExamplePrograms_Arithmetic)
- `loops.s` ✅ (TestExamplePrograms_Loops)
- `conditionals.s` ✅ (TestExamplePrograms_Conditionals)

All 4 tested programs are in the "working" category. The remaining 31 programs (89%) have no automated verification, allowing regressions to go undetected. The previous claim of "22/30 working" appears unverified and likely inaccurate.

**Recommendation:** Before fixing individual programs, add comprehensive integration tests for all examples to:
1. Prevent future regressions
2. Provide a baseline for measuring improvements
3. Enable CI/CD to catch breaking changes

#### Comprehensive Example Program Testing (PREREQUISITE)
**Priority:** CRITICAL | **Effort:** 6-8 hours

**Rationale:** Currently only 4/35 (11%) of example programs have automated tests. This allowed the low success rate to go undetected and makes it impossible to track improvements or prevent regressions.

**Tasks:**
- [ ] **Create integration test suite for all 35 example programs**
  - Add `tests/integration/examples_test.go` with test for each program
  - For working programs: verify exit code 0 and expected output
  - For failing programs: mark as `t.Skip()` with failure reason (e.g., "Known issue: LSR instruction")
  - For parse errors: verify specific error message
  
- [ ] **Document expected behavior**
  - Create `tests/integration/testdata/expected_outputs.json` with:
    - Expected exit code
    - Expected stdout (substring matching for non-deterministic output)
    - Known issues/skip reasons
  
- [ ] **CI/CD Integration**
  - Add example tests to CI pipeline
  - Configure to fail on newly broken examples
  - Generate test coverage report for examples

**Benefits:**
- Track exact regression when changes break examples
- Measure progress as failures are fixed (e.g., "17/35 working" → "18/35 working")
- Prevent future breaking changes
- Document expected behavior formally

#### Critical Instructions Missing
- [ ] **Implement LSR/LSL/ASR/ROR as standalone instructions**
  - Currently only work as shift modifiers (e.g., `MOV r0, r1, LSR #2`)
  - Need to support standalone form (e.g., `LSR r0, r1, #2`)
  - These are aliases that should expand to `MOV Rd, Rm, SHIFT #n`
  - Required by xor_cipher.s (line 256: `LSR r0, r0, #4`)
  - ARM documentation shows these as pseudo-instructions that expand to MOV

#### Parser Enhancements  
- [ ] **Support negative constant values** in `.equ` directives
  - Currently rejects `.equ EMPTY_KEY, -1`
  - Affects hash_table.s and potentially other programs
  
- [ ] **Support alternative comment syntax** (optional)
  - GNU assembler uses '@' for comments (we use ';')
  - Affects recursive_fib.s, sieve_of_eratosthenes.s
  - Low priority - can document syntax requirements instead

#### Memory Access Violation Investigation
- [ ] **Debug control flow issues** causing PC jumps to unmapped memory
  - Affects 17 programs with similar failure pattern
  - May indicate systemic issue with:
    - Function return address handling
    - Stack frame management
    - Section loading (.text/.data organization)
  - Recommend debugging one representative case (e.g., factorial.s, times_table.s)

---

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
