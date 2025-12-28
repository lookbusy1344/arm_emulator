# ARM Emulator Code Review

**Date:** 2025-12-27
**Reviewer:** Gemini (via Copilot CLI)
**Codebase:** ARM2 Emulator in Go
**Scope:** Full codebase analysis, verifying previous findings, and identifying new issues.

---

## Executive Summary

The ARM2 emulator is a robust and well-structured project, demonstrating high code quality and a strong test suite (100% pass rate). However, critical issues identified in previous reviews remain unaddressed, specifically regarding thread safety in the TUI and complexity in the parser. Additionally, the `encoder` package completely lacks unit tests, which is a significant risk for a core component.

### Key Findings

1.  ~~**CRITICAL: TUI Race Conditions Persist**~~ - **FIXED 2025-12-28** - Added `sync.RWMutex` to protect shared state in `debugger/tui.go`.
2.  ~~**CRITICAL: Missing Encoder Tests**~~ - **FIXED 2025-12-28** - Added comprehensive encoder unit tests in `tests/unit/encoder/`.
3.  **High Complexity in Parser** - `parseOperand` remains a monolithic function that is difficult to maintain.
4.  ~~**Feature Creep/Inaccuracy**~~ - **FIXED 2025-12-28** - Removed `MOVW` (ARMv7) support for strict ARM2 compliance. Updated `hash_table.s` to use literal pools.

---

## 1. Verification of Previous Findings (Opus Review)

I have verified the findings from the `CODE_REVIEW_OPUS.md` (dated 2025-11-26) and found that the recommended fixes were **NOT implemented**.

### 1.1 TUI Thread Safety (Confirmed Critical)
**File:** `debugger/tui.go` (Lines 420-496)
The `executeUntilBreak` method spawns a goroutine that modifies `t.Debugger.Running` and calls methods like `t.CaptureRegisterState()` and `t.DetectRegisterChanges()` without any mutex locking. These methods modify fields that are also accessed by the main thread during `RefreshAll()`.
*   **Status:** ✅ **FIXED 2025-12-28** - Added `sync.RWMutex` (`stateMu`) to protect shared state. Writers (Capture/Detect methods) use `Lock()`, readers (Update methods) use `RLock()` and copy state to minimize lock hold time.

### 1.2 Parser Complexity (Confirmed)
**File:** `parser/parser.go` (Lines 468-630)
The `parseOperand` function is approximately 160 lines long, containing a large switch statement that handles immediate values, memory addresses, register lists, and pseudo-instructions.
*   **Status:** 🔴 **Unfixed**. Refactoring into smaller functions (`parseImmediateOperand`, `parseMemoryOperand`, etc.) is highly recommended.

---

## 2. New Findings

### 2.1 Missing Encoder Tests (Critical)
**Location:** `tests/unit/encoder/` (Missing)
The `encoder` package is responsible for generating machine code. While the `vm` and `parser` have extensive tests, the `encoder` has **no unit tests**.
*   **Status:** ✅ **FIXED 2025-12-28** - Created `tests/unit/encoder/encoder_test.go` with comprehensive tests:
    - Condition code encoding (all 16 conditions)
    - Immediate value encoding with rotation
    - Register parsing (R0-R15, SP, LR, PC aliases)
    - Data processing instructions (MOV, ADD, SUB, AND, ORR, CMP, etc.)
    - Shift encoding (LSL, LSR, ASR, ROR)
    - Memory addressing modes
    - Branch, multiply, SWI, NOP encoding
    - Error handling for invalid inputs

### 2.2 ARM2 Incompatibility (MOVW)
**File:** `encoder/data_processing.go` (Line 237)
The encoder attempts to use `MOVW` (Move Wide) if an immediate value fits in 16 bits but cannot be encoded as a standard ARM immediate (8-bit rotated).
*   **Status:** ✅ **FIXED 2025-12-28** - Removed `MOVW` support for strict ARM2 compliance:
    - Encoder now returns clear error: "use LDR Rd, =value with literal pool"
    - Removed `MOVWOpcodeValue` constant
    - Updated `hash_table.s` example to use `LDR Rd, =value` for large immediates (330, 420, 1000, 2550)
    - Updated expected test output

### 2.3 Encoder Immediate Rotation Logic
**File:** `encoder/encoder.go` (Line 260)
The `encodeImmediate` function correctly implements the ARM immediate encoding logic (finding a rotation such that `value = imm8 ROR (2*rot)`).
*   **Note:** The logic is correct, but the lack of tests makes it fragile.

---

## 3. Recommendations

### 3.1 Immediate Actions
1.  ~~**Fix TUI Concurrency:**~~ ✅ **DONE** - Added `sync.RWMutex` to the `TUI` struct.
2.  ~~**Add Encoder Tests:**~~ ✅ **DONE** - Created comprehensive test suite in `tests/unit/encoder/`.

### 3.2 Refactoring
1.  **Refactor `parseOperand`:** Split this function into `parseImmediate`, `parseMemory`, `parseRegisterList`, etc.
2.  ~~**Remove `MOVW`:**~~ ✅ **DONE** - Encoder now adheres to ARM2 specification.

### 3.3 Documentation
1.  **Update `TODO.md`:** Add these findings to the TODO list to ensure they are tracked.

---

## 4. Conclusion

The project is in a good state overall, but the lingering concurrency issue and the lack of encoder tests are significant blind spots. Addressing these will bring the project to a production-ready standard.
