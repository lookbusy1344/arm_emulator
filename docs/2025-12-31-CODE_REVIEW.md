# ARM Emulator Code Review (Dec 31, 2025)

**Reviewer:** Claude Opus 4.5
**Method:** 4 parallel code review agents covering VM, Parser/Encoder, Debugger/TUI, Main/Service/Tools/Config

## Current Status

- **Build**: Passing
- **Lint**: 0 issues
- **Tests**: 1,024 tests passing
- **Code Size**: ~54K lines of Go
- **Previous Review (Dec 29)**: Most critical issues already fixed

---

## Critical Issues (Must Fix)

### 1. Security: Path Traversal in Preprocessor Include

**Location**: `parser/preprocessor.go:51-66`

The preprocessor allows `.include "../../../etc/passwd"` style path traversal. Current code:

```go
path := filepath.Join(p.baseDir, filename)
content, err := os.ReadFile(path) // #nosec G304
```

**Problem**: An attacker could read arbitrary files if the emulator is run with elevated privileges or in a sensitive directory.

**Fix**: Validate resolved path stays within baseDir:

```go
absPath, _ := filepath.Abs(filepath.Join(p.baseDir, filename))
absBase, _ := filepath.Abs(p.baseDir)
if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) {
    return "", fmt.Errorf("include path escapes base directory")
}
```

---

### 2. DoS: Unbounded Include Depth

**Location**: `parser/preprocessor.go:50-79`

No limit on nested includes. A deeply nested but non-circular include chain (A includes B, B includes C, ...) could exhaust stack space or cause excessive memory allocation.

**Fix**: Add `MaxIncludeDepth = 100` constant and check `len(p.includeStack)`:

```go
const MaxIncludeDepth = 100

func (p *Preprocessor) ProcessFile(filename string) (string, error) {
    if len(p.includeStack) >= MaxIncludeDepth {
        return "", fmt.Errorf("include depth exceeds maximum (%d)", MaxIncludeDepth)
    }
    // ... rest of function
}
```

---

### 3. Race Condition: Breakpoint HitCount Modification

**Location**: `debugger/debugger.go:263-287`

After `GetBreakpoint()` releases its RLock, `bp.HitCount++` modifies the breakpoint without any lock protection. If another goroutine modifies or deletes the breakpoint between the `GetBreakpoint()` call and the subsequent operations, this could cause a data race.

```go
if bp := d.Breakpoints.GetBreakpoint(pc); bp != nil {  // Lock released here
    // ...
    bp.HitCount++  // Race condition: no lock held

    if bp.Temporary {
        _ = d.Breakpoints.DeleteBreakpoint(bp.ID)  // Race condition
    }
}
```

**Fix**: Add atomic method `BreakpointManager.ProcessHit(addr uint32) (*Breakpoint, bool)` that atomically increments hit count and handles temporary deletion while holding the lock.

---

### 4. Race Condition: SendInput Without Lock

**Location**: `service/debugger_service.go:1010-1026`

The `SendInput` method accesses `s.vm.OutputWriter` without synchronization. A concurrent reset could set it to nil between the check and the write, causing a nil pointer dereference.

```go
func (s *DebuggerService) SendInput(input string) error {
    // NOTE: No mutex lock here!
    if s.vm.OutputWriter != nil {
        _, _ = s.vm.OutputWriter.Write([]byte(input + "\n"))
    }
}
```

**Fix**: Use `RLock()` when accessing OutputWriter:

```go
s.mu.RLock()
outputWriter := s.vm.OutputWriter
s.mu.RUnlock()

if outputWriter != nil {
    _, _ = outputWriter.Write([]byte(input + "\n"))
}
```

---

### 5. Resource Leak: stdin Pipe Never Closed

**Location**: `debugger/tui.go:127-128`

The `io.Pipe()` is created in `NewTUIWithScreen()`, but neither `stdinPipeReader` nor `stdinPipeWriter` are ever closed. This causes a goroutine leak if the guest program is blocked on stdin read when the TUI exits.

**Fix**: Close pipe in `Stop()` method:

```go
func (t *TUI) Stop() {
    if t.stdinPipeWriter != nil {
        t.stdinPipeWriter.Close()
    }
    t.App.Stop()
}
```

---

### 6. Integer Overflow: Stack Size Calculation

**Location**: `main.go:140-141`

The stack size from CLI flag is not validated before addition, which could wrap around:

```go
stackTop := uint32(vm.StackSegmentStart + *stackSize) // #nosec G115
```

**Fix**: Add validation before calculation:

```go
const maxStackSize = 0x10000000 // 256MB reasonable maximum
if *stackSize > maxStackSize {
    fmt.Fprintf(os.Stderr, "Error: stack size %d exceeds maximum allowed %d\n", *stackSize, maxStackSize)
    os.Exit(1)
}
```

---

## Important Issues (Should Fix)

### 7. Recursive RLock Acquisition

**Location**: `service/debugger_service.go:738-739`

`GetDisassembly` holds `s.mu.RLock()` then calls `GetSymbolForAddress()` which also attempts to acquire RLock. While Go allows multiple readers, if someone changes `GetSymbolForAddress` to acquire a write lock in the future, this would deadlock.

**Fix**: Use existing `getSymbolForAddressUnsafe()` method instead.

---

### 8. Duplicate loadProgramIntoVM Code

**Location**: `main.go` and `service/debugger_service.go`

~200 lines of `loadProgramIntoVM` logic duplicated between files. Bug fixes or changes need to be made in two places.

**Fix**: Extract to `vm/loader.go` as shared function.

---

### 9. ADC/SBC Overflow Flag Calculation

**Location**: `vm/data_processing.go:112-139`

The overflow calculation for ADC uses only `op1` and `op2`, ignoring the carry-in:

```go
case OpADC:
    overflow = CalculateAddOverflow(op1, op2, result)  // Should include carryIn
```

The ARM architecture specifies that the V flag for ADC should be computed considering all three inputs (op1, op2, and carry).

**Fix**: Use correct three-operand overflow calculation:

```go
temp := op1 + op2
tempOverflow := CalculateAddOverflow(op1, op2, temp)
finalOverflow := CalculateAddOverflow(temp, carryIn, result)
overflow = tempOverflow || finalOverflow
```

---

### 10. MRS/MSR Double Cycle Increment

**Location**: `vm/psr.go:49-51, 96-98`

Both MRS and MSR functions call `IncrementCycles(1)` after `IncrementPC()`, but the main `Step()` function in `executor.go` also calls `IncrementCycles(1)`. This results in MRS/MSR counting 2 cycles instead of 1.

**Fix**: Remove the `IncrementCycles(1)` calls from `executeMRS()` and `executeMSR()`.

---

### 11. Inefficient O(n) Symbol Lookup in TUI

**Location**: `debugger/tui.go:1077-1084`

`findSymbolForAddress()` iterates through all symbols for every address lookup. This is called multiple times per refresh.

```go
func (t *TUI) findSymbolForAddress(addr uint32) string {
    for sym, symAddr := range t.Debugger.Symbols {
        if symAddr == addr {
            return sym
        }
    }
    return ""
}
```

**Fix**: Create reverse map `map[uint32]string` at load time for O(1) lookups.

---

### 12. Shift Amount Range Not Validated

**Location**: `encoder/encoder.go:279-324`

The `parseShift` function parses shift amounts but doesn't validate the range. ARM shift amounts should be 0-31.

**Fix**: Add validation:

```go
if amount > 31 {
    return 0, 0, -1, fmt.Errorf("shift amount out of range: %d (max 31)", amount)
}
```

---

### 13. Inefficient String Splitting in Parser

**Location**: `parser/parser.go:894-906`

The `getRawLineFromInput` function splits the entire input on newlines every time it's called. For large files, this creates O(n * m) complexity.

**Fix**: Cache the split lines in the parser struct:

```go
type Parser struct {
    inputLines []string // cached split lines
}

func (p *Parser) getRawLineFromInput(lineNum int) string {
    if p.inputLines == nil {
        p.inputLines = strings.Split(p.lexer.input, "\n")
    }
    // ...
}
```

---

### 14. Bubble Sort for Symbols

**Location**: `main.go:1012-1019`

O(n^2) bubble sort used for symbol sorting:

```go
for i := 0; i < len(entries); i++ {
    for j := i + 1; j < len(entries); j++ {
        if entries[i].symbol.Value > entries[j].symbol.Value {
            entries[i], entries[j] = entries[j], entries[i]
        }
    }
}
```

**Fix**: Use `sort.Slice()` for O(n log n):

```go
sort.Slice(entries, func(i, j int) bool {
    return entries[i].symbol.Value < entries[j].symbol.Value
})
```

---

### 15. Unbounded Expression Value History

**Location**: `debugger/expressions.go:30-33`

The `valueHistory` slice grows unbounded with each `EvaluateExpression()` call.

**Fix**: Add `maxHistorySize` limit and trim old entries when exceeded.

---

## Minor Issues

| # | Location | Issue |
|---|----------|-------|
| 16 | `parser/lexer.go:512-530` | Register validation incomplete - `R999` accepted as valid |
| 17 | `debugger/interface.go:90` | Magic number `148` for terminal width should be constant |
| 18 | `service/debugger_service.go:711,775` | Hardcoded validator limits (1000, 100000) should be named constants |
| 19 | `tools/lint.go:529` | Custom `min()` function shadows Go 1.21+ builtin |
| 20 | `tools/lint.go:440-475` | Levenshtein allocates full (n+1)x(m+1) matrix; use 2-row version |
| 21 | `service/debugger_service.go:597` | `stepsBeforeYield = 1000` should be package-level constant |
| 22 | `vm/constants.go` vs `vm/arch_constants.go` | Duplicate register constants (`SPRegister` = `ARMRegisterSP`) |
| 23 | `main.go:440-462` | Exit code not propagated from debugger/TUI modes |
| 24 | `debugger/expr_lexer.go:22` | Unused `ExprTokenStar` token type defined but never produced |
| 25 | `config/config.go:185-189` | Config parse errors logged but silently ignored; may confuse users |

---

## Implementation Order

### Phase 1 - Security (Immediate)

1. Path traversal fix (#1)
2. Include depth limit (#2)
3. Stack size validation (#6)

### Phase 2 - Concurrency (High Priority)

4. Breakpoint race condition (#3)
5. SendInput race condition (#4)
6. Recursive RLock (#7)
7. stdin pipe cleanup (#5)

### Phase 3 - Correctness

8. ADC/SBC overflow flags (#9)
9. MRS/MSR double cycle (#10)
10. Shift amount validation (#12)

### Phase 4 - Performance/Maintainability

11. Duplicate loadProgramIntoVM (#8)
12. Symbol lookup optimization (#11)
13. Parser string splitting (#13)
14. Bubble sort replacement (#14)
15. Expression history limit (#15)

### Phase 5 - Polish

- Minor issues #16-25

---

## Summary

| Category | Count |
|----------|-------|
| Critical | 6 |
| Important | 9 |
| Minor | 10 |
| **Total** | **25** |

---

## Positive Observations

1. **Strong test coverage**: 1,024 tests with 100% pass rate
2. **Clean linting**: golangci-lint reports 0 issues
3. **Good mutex discipline**: TUI correctly uses `sync.RWMutex` with documented locking strategy
4. **Named constants**: Minimal magic numbers throughout codebase
5. **Security awareness**: Filesystem sandboxing, bounded input sizes, proper escape sequence handling
6. **Clear separation of concerns**: Well-organized package structure
7. **Previous review addressed**: Most Dec 29th critical issues already fixed
