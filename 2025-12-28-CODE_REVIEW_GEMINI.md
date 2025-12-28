# Code Review - ARM Emulator
**Date:** 2025-12-28
**Reviewer:** Gemini (AI Assistant)

## 1. Executive Summary

The ARM emulator project is a well-structured and comprehensive implementation of an ARM2-compatible virtual machine. It features a clean separation of concerns between the VM core, parser, encoder, and debugger. The codebase demonstrates good Go practices, including effective use of interfaces and error handling.

However, there are a few areas that require attention, particularly regarding input validation in system calls (potential DoS), incomplete support for escape sequences, and some architectural duplication in string processing.

## 2. Critical Issues & Bugs

### 2.1. Unbounded Memory Allocation in Syscalls (DoS Risk)
**Severity:** High
**Location:** `vm/syscall.go` (`handleReadString`, `handleReadInt`)

The `handleReadString` and `handleReadInt` functions use `vm.stdinReader.ReadString('\n')` to read input from the standard input. `bufio.Reader.ReadString` reads until the delimiter is found, expanding the buffer as needed.
*   **Risk:** A malicious actor or a large input file (e.g., via pipe) without newlines could cause the VM to allocate an unbounded amount of memory, leading to a Denial of Service (OOM crash).
*   **Recommendation:** Use a bounded reader or `ReadSlice` with a fixed buffer size, or implement a limit on the number of bytes read before a newline is found.

### 2.2. Incomplete Escape Sequence Support
**Severity:** Medium
**Location:** `main.go` (`processEscapeSequences`), `encoder/encoder.go` (`parseImmediate`)

The project implements custom escape sequence parsing in multiple places.
*   **Issue:** The current implementation supports basic escapes (`\n`, `\t`, `\\`, etc.) but lacks support for hex (`\xNN`) and octal (`\NNN`) escape sequences, which are standard in assembly and C.
*   **Impact:** Users cannot easily embed arbitrary byte values in strings or character literals if they don't map to standard escapes.
*   **Recommendation:** Implement a centralized `ParseEscapeSequence` utility in the `parser` or `tools` package that supports full C-style escape sequences, and use it consistently across the codebase.

## 3. Architectural Observations

### 3.1. Stack/Heap Collision Risk
**Location:** `vm/memory.go`, `vm/constants.go`

*   **Observation:** The Heap segment (0x30000-0x40000) and Stack segment (0x40000-0x50000) are adjacent.
*   **Behavior:** The heap allocator correctly checks bounds and won't grow beyond 0x40000. However, the stack pointer (SP) is not bounds-checked (by design, to simulate hardware).
*   **Risk:** If the stack grows downwards past 0x40000, it will silently overwrite heap data. While this mimics real hardware behavior, it can lead to difficult-to-debug corruption in the emulator.
*   **Recommendation:** Consider adding a "Stack Guard" feature (optional/debug mode) that warns or halts if SP crosses into the Heap segment.

### 3.2. Split Literal Pool Logic
**Location:** `parser/parser.go`, `encoder/encoder.go`

*   **Observation:** The literal pool handling is split. The parser estimates pool sizes and adjusts addresses (`adjustAddressesForDynamicPools`), while the encoder places the actual literals.
*   **Complexity:** This two-pass adjustment logic is complex and fragile. If the parser's estimation differs from the encoder's reality in unexpected ways (beyond the adjustment logic), addresses could be misaligned.
*   **Recommendation:** Ensure strict synchronization between parser estimation and encoder generation. Ideally, the assembler would be single-pass with backpatching or strictly two-pass where the first pass calculates exact sizes.

### 3.3. Code Duplication
*   **Escape Sequences:** Logic is duplicated between `main.go` and `encoder/encoder.go`.
*   **Address Calculation:** Both `parser` and `main.go` (during loading) calculate addresses for directives.

## 4. Code Quality & Style

*   **Go Idioms:** The code generally follows good Go idioms.
*   **Error Handling:** Error handling is robust, with clear distinction between VM errors (halting) and syscall errors (returning error codes).
*   **Documentation:** The code is well-commented.
*   **Safety:** Use of `SafeIntToUint32` and similar helpers in `vm/safeconv.go` is a good practice.

## 5. Missing Features

*   **Full Escape Sequences:** As mentioned above (`\xNN`, `\uNNNN`).
*   **Memory Protection:** No MMU or MPU simulation (expected for ARM2, but limits OS simulation).
*   **Coprocessor Support:** Stubs exist but no implementation (e.g., for floating point).

## 6. Recommendations

1.  **Fix the DoS vulnerability** in `vm/syscall.go` immediately.
2.  **Refactor escape sequence parsing** into a shared utility function.
3.  **Add hex escape support** to the new utility.
4.  **Add a stack overflow warning** if SP enters the heap segment.
