# Code Review – ARM Emulator (Fresh Eyes)
**Date:** 2025-12-29  
**Reviewer:** GPT  

## What I ran
- `go build -o arm-emulator && go clean -testcache && go test ./...` ✅ (all passing)

## Executive summary
This is an impressively feature-complete ARM2-ish emulator with strong tests and a generally sane package layout (`vm/`, `parser/`, `encoder/`, `debugger/`, `service/`, `gui/`). The main risks I see are **(a)** a few **latent correctness bugs in the assembler/parser around literals/strings**, **(b)** **feature/documentation mismatch** (notably preprocessing / `.include`), and **(c)** **concurrency/data-race hazards in the TUI**.

## Strengths
- **Clear separation of concerns**: VM execution vs. parsing vs. encoding vs. debugging.
- **Security awareness**: syscall layer has explicit wraparound checks, size caps, and a sandbox model (`-fsroot`).
- **Good “product” features**: tracing, coverage, stack/flag/register diagnostics, GUI service layer.
- **Test depth**: unit + integration, example programs as integration tests.

---

## Findings (bugs / missing features / architecture issues)

### 1) Parser / assembler: preprocessor exists but is not integrated (feature mismatch)
**Severity: High (user-facing missing feature + doc mismatch)**
- `parser/preprocessor.go` is implemented and thoroughly tested, but **`main.go` never runs it**.
  - `main.go` does: `input, _ := os.ReadFile(asmFile)` → `parser.NewParser(string(input), ...)` → `Parse()`.
  - `Parser` allocates `p.preprocessor = NewPreprocessor("")` but **`Parse()` doesn’t use it**.
- README/CLAUDE claim “parser with preprocessor” and `.include` support. As-is, `.include`, `.ifdef`, `.ifndef`, `.else`, `.endif` are effectively dead features for the CLI.

**Recommendation:**
- Add a single entrypoint like `parser.ParseFile(path, opts)` (or a `PreprocessAndParse(...)` function) and make **CLI/GUI/TUI all use it**, so `.include` works consistently.

### 2) Parser: `.ascii/.asciz` address accounting ignores escape expansion
**Severity: High (silent misassembly / incorrect addresses)**
In `parser/parser.go`, directive address reservation uses raw string length:
- `.ascii`, `.asciz`, `.string`:
  - Strips quotes and then does `p.currentAddress += uint32(len(str))`.
  - But when loading into memory (`main.go` and `service/debugger_service.go`), you do:
    - `processedStr := parser.ProcessEscapeSequences(str)` and write the processed bytes.

This means directives like:
```asm
.ascii "A\nB"   ; raw len=4, processed len=3
.asciz "\x41"   ; raw len=4, processed len=1 (+NUL)
```
will likely produce **address drift** after the directive, breaking labels, pools, and branches.

**Recommendation:**
- In `handleDirective` for string directives, reserve `len(ProcessEscapeSequences(str))` (plus NUL for `.asciz/.string`).

### 3) Parser: “dynamic literal pool sizing” bookkeeping looks incorrect
**Severity: High (layout errors, especially with many `LDR =...` + `.ltorg`)**
`parser.Parser.countLiteralsPerPool()` claims to count **unique** literals per pool, but the current implementation does:
```go
count := len(literalsBeforePool[poolIdx])
literalsBeforePool[poolIdx][uint32(count)] = true
```
This does **not** deduplicate by literal value/expression; it effectively counts one per pseudo-instruction (and the key is just the current set size).

If the intent is “unique literal values per pool”, you need a stable key:
- The operand string itself (e.g. `inst.Operands[1]` after trimming `=`), or
- A parsed constant (when resolvable), falling back to string key for forward refs.

**Recommendation:**
- Either:
  1) Rename the feature to “count pseudo-literals (one per LDR =...)” and ensure encoder/layout matches, or
  2) Actually dedupe by literal expression.

### 4) Parser: number parsing rejects `-2147483648`
**Severity: Medium (edge-case correctness)**
`parseNumber()` treats negative values as:
- parse absolute into `uint32`, then checks `result > math.MaxInt32` and errors.
This rejects `-2147483648` even though it is a valid int32.

**Recommendation:**
- Allow `result == 2147483648` as a special-case (`math.MinInt32`).

### 5) Parser: `.byte` character literal handling doesn’t support escapes
**Severity: Medium (assembler quality)**
Loader logic in `main.go`/`service` handles `.byte 'A'` only when `len(arg)==3`; it rejects `'.\n'`, `'\x41'`, etc.
Yet the project already has `parser.ParseEscapeChar()` which could be used.

**Recommendation:**
- Support `.byte '\n'`, `.byte '\x0A'`, `.byte '\123'` (octal) using `ParseEscapeChar`.

---

## VM / execution layer

### 6) Cycle limiting: two knobs (`MaxCycles` vs `CycleLimit`) with inconsistent enforcement
**Severity: Medium (operational correctness / surprising behavior)**
- `VM.Step()` enforces `CycleLimit`.
- `VM.Run()` enforces `MaxCycles` (and also separately checks a placeholder “halt” condition).
- `main.go` sets `machine.CycleLimit = *maxCycles` but never sets `MaxCycles`.

This is confusing and can lead to:
- CLI path vs. `vm.Run()` path behaving differently.

**Recommendation:**
- Collapse to a single field, or make one clearly derived from the other.

### 7) `Memory.MakeCodeReadOnly()` exists but is never called in production code
**Severity: Low/Medium (spec mismatch / security posture)**
The code segment is writable by default (“to support .word/.byte data and self-modifying code”). That might be intentional, but it’s a sharp edge:
- It weakens “W^X” style assumptions.
- It complicates debugging/trace expectations.

**Recommendation:**
- Decide and document:
  - If self-modifying is supported: keep as-is.
  - If not: call `MakeCodeReadOnly()` after loading.

---

## Syscalls / filesystem sandbox

### 8) Docs say sandbox violations halt the VM; code generally returns an error code
**Severity: Medium (doc mismatch + user expectation)**
README/CLAUDE state:
- “Attempts to escape using `..` or symlinks will halt the VM”.
But `vm/syscall.go` `handleOpen` behavior is:
- `ValidatePath` failure → log to stderr → set `R0 = 0xFFFFFFFF` → continue execution.

**Recommendation:**
- Either update docs to match the implementation, or change implementation to match docs.

### 9) `ValidatePath` implementation/comment mismatch + conservative/naive checks
**Severity: Medium (security hardening / portability)**
- Comment says “EvalSymlinks returns error if any component is a symlink” — that’s not how `filepath.EvalSymlinks` works; it **resolves** symlinks.
- `strings.Contains(path, "..")` is a blunt instrument:
  - blocks safe names containing `..` (e.g. `foo..bar`),
  - but doesn’t model actual path components.
- `ValidatePath` returns `fullPath` (pre-resolution) rather than `resolvedPath`.
  - This can reintroduce a small TOCTOU surface if symlinks/paths change after validation.

**Recommendation:**
- Replace substring check with a component-based check after `Clean` (split on separators and reject `..`).
- Consider returning `resolvedPath` for actual open (or explicitly document the TOCTOU trade-off).

---

## Debugger / UI

### 10) TUI likely has data races (VM stepped in one goroutine, UI reads in another)
**Severity: High (heisenbugs; race detector will likely flag)**
`debugger/tui.go` runs execution in a goroutine and frequently calls `QueueUpdateDraw` to refresh views. Those view refreshes read VM state while the execution goroutine is mutating it; there is **no shared mutex around VM** in the TUI path.

You already solved this problem in the GUI via `service/DebuggerService` (single lock protecting VM + debugger interactions), but the TUI bypasses it.

**Recommendation:**
- Either:
  1) Run *all* VM state access on one goroutine (event-loop model), or
  2) Introduce a TUI-level mutex that wraps VM.Step and all reads during refresh, or
  3) Reuse `DebuggerService` for TUI as well.

---

## Service / GUI

### 11) Debug logging path hardcoded to `/tmp/...`
**Severity: Low (portability)**
`service/debugger_service.go` writes debug logs to `/tmp/arm-emulator-service-debug.log` when `ARM_EMULATOR_DEBUG` is set.
- On Windows, `/tmp` isn’t generally valid.

**Recommendation:**
- Use `os.TempDir()` + a stable filename.

### 12) Program loading logic is duplicated (CLI vs service)
**Severity: Medium (maintenance risk)**
There is substantial duplication between:
- `main.go: loadProgramIntoVM(...)`
- `service/debugger_service.go: loadProgramIntoVM(...)`

Even with comments like “matches main.go”, this will drift over time.

**Recommendation:**
- Move program loading into a shared package function (e.g. `vmloader.Load(program, vm, opts)`), and have both call it.

---

## Closing recommendations (prioritized)
1. **Fix assembler layout correctness**: string directive sizing + literal pool counting.
2. **Integrate the preprocessor** (or remove it / correct docs).
3. **Make TUI execution/threading race-free** (prefer single-threaded VM access).
4. **Unify cycle limit semantics** (`MaxCycles` vs `CycleLimit`).
5. Tighten/clarify **filesystem sandbox** semantics and align docs with behavior.

## Notes
- I intentionally did not propose “big refactors” beyond reducing duplication where it directly affects correctness/maintenance (loader + preprocessor integration).
- The codebase’s test suite is strong; I’d strongly recommend adding `go test -race ./...` to CI once the TUI/service locking story is resolved (your TODO already mentions this).
