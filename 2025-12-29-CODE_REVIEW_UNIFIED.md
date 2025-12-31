# Unified Code Review - ARM Emulator
**Date:** 2025-12-29
**Sources:** GPT (o3) review, Gemini (2.5 Pro) review
**Validation & Additional Analysis:** Claude (Opus 4.5)

## Executive Summary

This is a well-structured ARM2 emulator with strong test coverage (1,024 tests, 100% pass rate) and clear separation of concerns.

**Total findings:** 16 (12 from original reviews + 4 additional)
**Race detector:** CLEAN (no data races detected)

### Critical Issues
1. **Parser correctness bugs** - string directive sizing uses raw length, literal pool counting doesn't deduplicate
2. **Preprocessor not integrated** - `.include`/`.ifdef` directives are dead code

### Notable Issues
3. **API confusion** - `MaxCycles` vs `CycleLimit` dual fields with different enforcement points
4. **Documentation mismatches** - sandbox behavior, ValidatePath comments
5. **Edge cases** - `parseNumber` rejects valid `-2147483648`, `.byte` doesn't support escapes

### Positive Findings
- Race detector passes all tests
- File descriptor cleanup properly implemented in `VM.Reset()`
- `RegisterSnapshot` refactoring centralizes state tracking
- Syscall error handling improved (returns error code, not halt)

---

## Findings

### HIGH Priority

#### 1. String Directives Use Raw Length Instead of Processed Length
**Status:** ✅ FIXED
**Location:** `parser/parser.go:356-371`

The parser reserves bytes for `.ascii`, `.asciz`, `.string` using `len(str)` on the raw string after quote removal, but `main.go` writes `ProcessEscapeSequences(str)` to memory. This causes address drift:

```asm
.ascii "A\nB"   ; reserves 4 bytes (raw), writes 3 bytes (processed)
.asciz "\x41"   ; reserves 4 bytes (raw), writes 2 bytes (1 + NUL)
```

Labels and branch targets after these directives will have incorrect addresses.

**Fix:** In `handleDirective` for string directives, reserve `len(ProcessEscapeSequences(str))` bytes.

---

#### 2. Literal Pool Counting Does Not Deduplicate
**Status:** ✅ FIXED
**Location:** `parser/parser.go:843-888`

`countLiteralsPerPool()` claims to count "unique literals" but actually just counts pseudo-instructions:

```go
count := len(literalsBeforePool[poolIdx])
literalsBeforePool[poolIdx][uint32(count)] = true  // Key is just set size!
```

If two instructions use `LDR R0, =0x1234` and `LDR R1, =0x1234`, they're counted as 2 literals instead of 1.

**Fix:** Use the literal expression string (e.g., `inst.Operands[1]` after trimming `=`) as the map key for deduplication.

---

#### 3. Preprocessor Is Implemented But Never Called
**Status:** ✅ FIXED
**Location:** `parser/file.go` (new), `main.go`, `gui/app.go`, `gui/main.go`

The parser has a full preprocessor with `.include`, `.ifdef`, `.ifndef`, `.else`, `.endif` support. However:
- `main.go` reads the file directly: `os.ReadFile(asmFile)` → `parser.NewParser(string(input), ...)`
- The preprocessor is never invoked, making these directives dead features

**Fix:** Create a unified `parser.ParseFile(path, opts)` function that runs preprocessing, then parse. Use this in CLI, GUI, and TUI.

---

### MEDIUM Priority

#### 4. `MaxCycles` vs `CycleLimit` Confusion
**Status:** ✅ FIXED
**Location:** `vm/executor.go:63, 111, 221-225`

Two fields exist with identical default values but different enforcement points:
- `CycleLimit` - checked in `VM.Step()` (line 222)
- `MaxCycles` - checked in `VM.Run()` (line 446)
- `main.go` sets `CycleLimit` from `--max-cycles` flag but never sets `MaxCycles`

This causes inconsistent behavior between CLI stepping and `VM.Run()`.

**Fix:** Collapse to a single field, or derive one from the other with clear semantics.

---

#### 5. `parseNumber` Rejects `-2147483648`
**Status:** ✅ FIXED
**Location:** `parser/parser.go:725-734`

```go
if result > uint32(math.MaxInt32) {
    return 0, fmt.Errorf("negative value %d is out of range for int32", result)
}
```

This rejects `-2147483648` because `2147483648 > 2147483647`. However, `-2147483648` is a valid `int32` (`math.MinInt32`).

**Fix:** Change condition to `result > uint32(math.MaxInt32) + 1` or handle `math.MinInt32` as a special case.

---

#### 6. `.byte` Character Literals Don't Support Escapes
**Status:** ✅ FIXED
**Location:** `main.go:715-739`, `service/debugger_service.go:204-232`

`.byte 'A'` works (3-character check), but `.byte '\n'`, `.byte '\x41'`, `.byte '\123'` fail.
The project has `parser.ParseEscapeChar()` that could handle these.

**Fix:** Use `ParseEscapeChar` for character literals in `.byte` directive handling.

---

#### 7. Documentation Says Sandbox Violations Halt; Code Returns Error
**Status:** ✅ FIXED (docs updated to match implementation)
**Location:** `docs/SECURITY.md`, `docs/CHANGELOG.md`

CLAUDE.md says: "Attempts to escape using `..` or symlinks will halt the VM with a security error."

Actual behavior: logs warning to stderr, sets `R0 = 0xFFFFFFFF`, continues execution.

**Fix:** Update documentation to match implementation (returning error is more flexible).

---

#### 8. `ValidatePath` Implementation Issues
**Status:** ✅ FIXED
**Location:** `vm/syscall.go:685-691, 704-705`

1. Comment says "EvalSymlinks returns error if any component is a symlink" - incorrect; it **resolves** symlinks
2. `strings.Contains(path, "..")` blocks legitimate names like `foo..bar`
3. Returns pre-resolution path, not resolved path (small TOCTOU surface)

**Fix:**
- Use component-based check after `filepath.Clean()` (split on separator, reject `..` components)
- Consider returning resolved path for actual file operations

---

#### 9. Debug Logging Hardcoded to `/tmp`
**Status:** ✅ FIXED
**Location:** `service/debugger_service.go:26-27`, `gui/app.go:28-29`

```go
f, err := os.OpenFile("/tmp/arm-emulator-service-debug.log", ...)
```

On Windows, `/tmp` doesn't exist.

**Fix:** Use `os.TempDir()` + filename.

---

#### 10. Program Loading Logic Duplicated
**Status:** DEFERRED (refactoring task, not a bug)
**Location:** `main.go:loadProgramIntoVM()`, `service/debugger_service.go:loadProgramIntoVM()`

Substantial duplication that will drift over time.

**Fix:** Extract to shared package function `vmloader.Load(program, vm, opts)`.

---

### LOW Priority

#### 11. `MakeCodeReadOnly()` Never Called
**Status:** ✅ RESOLVED - Writable code is intentional for ARM2 authenticity
**Location:** `vm/memory.go`

Code segment is writable by default. This is an **intentional design choice** for historical accuracy:

**ARM2 Hardware Reality:**
- ARM2 (1986) had **no MMU or memory protection** - no concept of read-only code or W^X
- Code and data were freely intermixed in real ARM2 programs
- The hardware allowed writes to any memory, including code

**Emulator Impact:**
- **37% of example programs** (18/49) embed writable data in code segment using `.space`/`.word`
- Parser doesn't implement `.data`/`.text` sections - all data after code stays in code segment
- Moving data to separate segment breaks ARM literal pools (4KB max offset from PC)
- Enforcing W^X would break historical accuracy and many example programs

**Resolution:** Documented writable code segment as intentional ARM2-authentic behavior in `vm/memory.go`.

---

#### 12. TUI Threading Model
**Status:** ✅ VERIFIED CORRECT
**Location:** `debugger/tui.go:66-71`, `debugger/debugger.go:45-47`

The threading model is correctly implemented with multiple layers of protection:

1. **Debugger.mu** - Protects execution state (Running, StepMode, etc.) with all access mutex-guarded
2. **TUI.stateMu** - Protects change-tracking state (ChangedRegs, RecentWrites, PrevState)
3. **Single-writer pattern** - VM state modified exclusively by background execution goroutine
4. **Queued reads** - UI reads VM state via `QueueUpdateDraw`, serialized on UI thread
5. **Race detector clean** - `go test -race ./...` passes with zero data races

The architecture ensures VM doesn't need additional mutex protection because writes are single-threaded and reads are queued. This is a valid and safe threading model.

---

## Additional Findings (Claude)

### 13. Debug Log File Handles Never Closed
**Status:** ✅ DOCUMENTED (intentional behavior)
**Location:** `service/debugger_service.go:25-28`, `gui/app.go:27-29`

When `ARM_EMULATOR_DEBUG` is set, a log file is opened in `init()` but never closed:

```go
f, err := os.OpenFile("/tmp/arm-emulator-service-debug.log", ...)
// f is never closed - leaked until process exit
```

This is a minor issue since it's one file per process and gets cleaned up on exit, but it's technically a resource leak.

**Fix:** Store the file handle and close it in a cleanup function, or accept this as intentional lifetime behavior and document it.

---

### 14. ValidatePath Comment Is Inaccurate
**Status:** ✅ FIXED (addressed in #8)
**Location:** `vm/syscall.go:704-705`

Comment says:
```go
// 6. Check for symlinks - EvalSymlinks returns error if any component is a symlink
```

This is incorrect. `filepath.EvalSymlinks` **resolves** symlinks to their targets, returning the final path. It only returns an error if the path doesn't exist or there's a permission issue.

**Fix:** Update comment to accurately describe behavior: "EvalSymlinks resolves symlinks and returns the canonical path".

---

### 15. Encoder Expression Evaluation Handles Only One Operator
**Status:** ✅ FIXED
**Location:** `encoder/encoder.go:326-360`

`evaluateExpression` scanned left-to-right for `+` or `-` and stopped at the first operator found. Expressions like `buffer+255-17` would fail with "invalid immediate value: 255-17".

**Verification:** Parser tests exist for multi-operator expressions (e.g., `buffer + 255 - 17`) but they weren't catching the encoder bug because they only tested parsing, not encoding.

**Fix:** Rewrote `evaluateExpression` to scan right-to-left for the last operator and recursively evaluate the left side. This maintains correct left-to-right evaluation order for expressions like `a+b-c` → `(a+b)-c`. Verified with test showing `buffer+255-17` correctly evaluates to `buffer+238`.

---

### 16. Race Detector Results
**Status:** VERIFIED CLEAN

Ran `go test -race ./...` and all tests pass with no race conditions detected. The TUI mutex implementation appears to be working correctly for the test scenarios covered.

```
ok  	github.com/lookbusy1344/arm-emulator/tests/unit/vm	6.108s
(all packages pass)
```

---

## Fixed/Completed Items

### A. Syscall Error Handling (Fixed)
**Source:** Gemini review
Sandbox violations now return error code to guest instead of halting VM. Security warnings logged to stderr.

### B. RegisterSnapshot Refactoring (Fixed)
**Source:** Gemini review
Created `vm/state.go` with `RegisterSnapshot` struct for centralized register state capture and comparison. Used by `trace.go` and `tui.go`.

---

## Invalid/Incorrect Claims

### TestAssert_MessageWraparound
**Source:** Gemini review
**Claim:** "The TODO.md listed TestAssert_MessageWraparound as a skipped/failing test"
**Reality:** This test is NOT in TODO.md. The test exists in `tests/unit/vm/security_fixes_test.go` and passes.

---

## Recommendations (Prioritized)

### Critical (Correctness)
1. **Fix string directive sizing** - causes address drift after escape sequences
2. **Fix literal pool deduplication** - wastes pool space, may cause layout issues

### High (Missing Features / Documentation)
3. **Integrate preprocessor** or remove/document as unfinished feature
4. **Update docs** to match sandbox implementation (returns error, not halt)
5. **Fix inaccurate ValidatePath comments** - misleading for future maintainers

### Medium (API / Edge Cases)
6. **Unify cycle limit fields** - `MaxCycles` vs `CycleLimit` confusion
7. **Fix parseNumber for MinInt32** - rejects valid `-2147483648`
8. **Add .byte escape support** - use existing `ParseEscapeChar()`

### Low (Portability / Cleanup)
9. **Use os.TempDir()** for debug logs (Windows compatibility)
10. **Extract shared program loading** - reduce code duplication
11. **Consider closing debug log handles** or document as intentional

### Already Verified
- Race detector passes - TUI threading appears sound
- File descriptor cleanup works correctly in `VM.Reset()`

---

## Test Commands

```bash
# Full test suite
go build -o arm-emulator && go clean -testcache && go test ./...

# With race detector (recommended after threading fixes)
go test -race ./...

# Lint
golangci-lint run ./...
```
