# ARM Emulator Project

This is an ARM emulator written in Go that implements a subset of the ARM2 instruction set.

## Build Command

```bash
go build -o arm-emulator
```

## Format Command

```bash
go fmt ./...
```

## Lint Command

```bash
golangci-lint run ./...
```

## Test Command

```bash
go clean -testcache
go test ./...
```

## Update dependencies

```bash
go get -u ./...
go mod tidy
go mod verify
```

## Run Command

```bash
./arm-emulator program.s
```

## GUI Commands (Wails)

**IMPORTANT:** Always use the `-nocolour` flag with `wails build` and `wails dev` to prevent ANSI escape codes in output.

### Build GUI

```bash
wails build -nocolour
```

### Run GUI in Development Mode

```bash
wails dev -nocolour
```

### Check Wails Environment

```bash
wails doctor
```

### E2E Testing

**IMPORTANT:** E2E tests require the Wails dev server to be running first. Tests will hang indefinitely if the backend is not available.

```bash
# Terminal 1: Start Wails dev server
cd gui
wails dev -nocolour

# Terminal 2: Run E2E tests
cd gui/frontend
npm run test:e2e                    # Run all tests
npm run test:e2e -- --project=chromium  # Run chromium only
npm run test:e2e:headed             # Run with visible browser
```

The Wails backend must be running on http://localhost:34115 before any E2E tests can execute.

## Project Structure

- `main.go` - Entry point and CLI interface
- `vm/` - Virtual machine implementation (CPU, memory, execution, syscalls, tracing, statistics)
- `parser/` - Assembly parser with preprocessor and macros
- `instructions/` - Instruction implementations (data processing, memory, branch, multiply)
- `encoder/` - Machine code encoder/decoder for binary ARM instructions
- `debugger/` - Debugging utilities with TUI (breakpoints, watchpoints, expression evaluation)
- `config/` - Cross-platform configuration management
- `tools/` - Development tools (linter, formatter, cross-reference generator)
- `tests/` - Test files (1,024 tests, 100% pass rate)
  - `tests/unit/` - Unit tests for all packages
  - `tests/integration/` - Integration tests for complete programs
- `examples/` - Example ARM assembly programs (49 programs, all fully functional including 3 interactive)
- `docs/` - User and developer documentation

## SWI Syscall Reference

For the complete syscall reference, see [docs/INSTRUCTIONS.md](docs/INSTRUCTIONS.md#system-instructions).

### Quick Reference

The emulator implements traditional ARM2 syscall convention: `SWI #immediate_value` where the syscall number is encoded directly in the instruction. Arguments and return values use registers R0-R2.

### Console I/O (0x00-0x07)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x00 | EXIT | Exit program | R0: exit code | - |
| 0x01 | WRITE_CHAR | Write character to stdout | R0: character | - |
| 0x02 | WRITE_STRING | Write null-terminated string | R0: string address | - |
| 0x03 | WRITE_INT | Write integer in specified base | R0: value, R1: base (2/8/10/16, default 10) | - |
| 0x04 | READ_CHAR | Read character from stdin (skips whitespace) | - | R0: character or 0xFFFFFFFF on error |
| 0x05 | READ_STRING | Read string from stdin (until newline) | R0: buffer address, R1: max length (default 256) | R0: bytes written or 0xFFFFFFFF on error |
| 0x06 | READ_INT | Read integer from stdin | - | R0: integer value or 0 on error |
| 0x07 | WRITE_NEWLINE | Write newline to stdout | - | - |

### File Operations (0x10-0x16)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x10 | OPEN | Open file | R0: filename address, R1: mode (0=read, 1=write, 2=append) | R0: file descriptor or 0xFFFFFFFF on error |
| 0x11 | CLOSE | Close file | R0: file descriptor | R0: 0 on success, 0xFFFFFFFF on error |
| 0x12 | READ | Read from file | R0: fd, R1: buffer address, R2: length | R0: bytes read or 0xFFFFFFFF on error |
| 0x13 | WRITE | Write to file | R0: fd, R1: buffer address, R2: length | R0: bytes written or 0xFFFFFFFF on error |
| 0x14 | SEEK | Seek in file | R0: fd, R1: offset, R2: whence (0=start, 1=current, 2=end) | R0: new position or 0xFFFFFFFF on error |
| 0x15 | TELL | Get current file position | R0: file descriptor | R0: position or 0xFFFFFFFF on error |
| 0x16 | FILE_SIZE | Get file size | R0: file descriptor | R0: size or 0xFFFFFFFF on error |

### Memory Operations (0x20-0x22)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x20 | ALLOCATE | Allocate memory from heap | R0: size in bytes | R0: address or 0 (NULL) on failure |
| 0x21 | FREE | Free allocated memory | R0: address | R0: 0 on success, 0xFFFFFFFF on error |
| 0x22 | REALLOCATE | Resize memory allocation | R0: old address, R1: new size | R0: new address or 0 (NULL) on failure |

### System Information (0x30-0x33)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x30 | GET_TIME | Get time in milliseconds since Unix epoch | - | R0: timestamp (lower 32 bits) |
| 0x31 | GET_RANDOM | Get random 32-bit number | - | R0: random value |
| 0x32 | GET_ARGUMENTS | Get program arguments | - | R0: argc, R1: argv pointer (0 in current impl) |
| 0x33 | GET_ENVIRONMENT | Get environment variables | - | R0: envp pointer (0 in current impl) |

### Error Handling (0x40-0x42)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x40 | GET_ERROR | Get last error code | - | R0: error code (0 in current impl) |
| 0x41 | SET_ERROR | Set error code | R0: error code | - |
| 0x42 | PRINT_ERROR | Print error message to stderr | R0: error code | - |

### Debugging Support (0xF0-0xF4)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0xF0 | DEBUG_PRINT | Print debug message to stderr | R0: string address | - |
| 0xF1 | BREAKPOINT | Trigger debugger breakpoint | - | - |
| 0xF2 | DUMP_REGISTERS | Print all registers to stdout | - | - |
| 0xF3 | DUMP_MEMORY | Dump memory region as hex dump | R0: address, R1: length (max 1KB) | - |
| 0xF4 | ASSERT | Assert condition is true | R0: condition (0=fail), R1: message address | Halts if condition is 0 |

**Note:** CPSR flags (N, Z, C, V) are preserved across all syscalls to prevent unintended side effects on conditional logic.

## Development Guidelines

**IMPORTANT:** After implementing each phase of development, update `PROGRESS.md` to reflect the completed work, including:
- Mark the phase as completed
- Document any implementation details or deviations from the original plan
- Update the status of related tasks
- Note outstanding work and issues in `TODO.md`

**IMPORTANT:** Always run `go fmt ./...`, `golangci-lint run ./...`, and `go build -o arm-emulator && go clean -testcache && go test ./...` after making changes to ensure code quality and correctness.

**IMPORTANT:** Do not delete tests without explicit instructions. Do not simplify tests because they fail. If you think a test is malfunctioning, think about it carefully and ask me before making any changes to the tests.

**IMPORTANT:** Anything that cannot be implemented should be noted in `TODO.md` with details so work can result later. TODO.md should not contain completed work, that should go in PROGRESS.md.

**IMPORTANT:** Do not modify example programs just to make them work without explicit permission, unless they are actually broken. Instead, fix the emulator to run the programs properly. Example programs are test cases that demonstrate expected behavior.

**IMPORTANT:** To ensure tests are up to date, recompile and clear the test cache before running tests `go build -o arm-emulator && go clean -testcache && go test ./...`

**IMPORTANT:** This emulator implements the classic ARM2 architecture. Do NOT implement Linux-style syscalls (using `SVC #0` with syscall number in R7 register). The emulator uses only traditional ARM2 syscall convention: `SWI #immediate_value` where the syscall number is encoded directly in the instruction. R7 is just a general-purpose register with no special meaning for syscalls.

**IMPORTANT:** All tests belong in the `tests/` directory structure, not in the main package directories. TUI tests use `tcell.SimulationScreen` to avoid terminal initialization issues. The `debugger.NewTUIWithScreen()` function accepts an optional screen parameter for testing while production code uses `debugger.NewTUI()` with the default screen.

**IMPORTANT:** Avoid embedding magic numbers directly in the code. Use named constants or enums for clarity and maintainability.

**IMPORTANT:** When doing code reviews, look at it with fresh eyes. Assume the engineer implemented it suspiciously quickly and is not to be trusted.

## Current Status

**Phase 11 (Production Hardening) - Complete ✅**
- All tests passing (100% pass rate, 1,024 total tests)
  - Unit tests: ~960 tests (includes 18 TUI tests using tcell.SimulationScreen, 12 register trace tests, security tests)
  - Integration tests: 64 tests (full end-to-end execution pipeline + CLI diagnostic flags + example programs)
- Code quality tools implemented (golangci-lint with errcheck, unused, govet, ineffassign, misspell)
- All lint issues resolved (0 issues reported)
- Go vet warnings fixed (method renames to avoid interface conflicts)
- CI updated to Go 1.25 with automated linting
- Build artifacts added to .gitignore
- Parser limitations resolved (debugger expression parser rewritten)
- Example programs: 49 programs total, all fully functional (100% success rate) ✅
  - 38+ programs with expected output files
  - Table-driven test framework for easy test maintenance
  - 46 non-interactive programs fully functional
  - 3 interactive programs work correctly with stdin (bubble_sort.s, calculator.s, fibonacci.s)
  - All bugs fixed (calculator.s EOF handling, test_ltorg.s and test_org_0_with_ltorg.s literal pool issues)
- Recent improvements (Oct 2025):
  - **Security hardening (Oct 22-23):** Comprehensive buffer overflow protection, address wraparound validation, file size limits (1MB default, 16MB max), thread-safety fixes (stdin reader moved to VM instance), file descriptor table size limit (1024), enhanced validation across all syscalls, wraparound protection verified with additional tests
  - Negative constants in .equ directives now supported
  - MOVW automatic encoding for 16-bit immediates
  - CMP/CMN instruction handling for un-encodable immediates
  - Shift operations work correctly as operand modifiers (LSR, LSL, ASR, ROR in MOV/ADD/etc.)
  - Data section ordering fixed (.data before .text)
  - Comprehensive integration test coverage for all examples
  - Comment support: `;`, `@`, `//` line comments and `/* */` block comments (GNU Assembler compatible)
  - **All example programs now working (100%):**
    - Fixed calculator.s EOF handling (no more infinite loop)
    - Fixed .ltorg literal pool space reservation in parser
    - Fixed test_org_0_with_ltorg.s missing branch instruction
- Character literal support complete (basic chars + escape sequences)
- ARM immediate encoding bug fixed (fibonacci.s, calculator.s now work correctly)
- Diagnostic modes complete (code coverage, stack trace, flag trace, register access pattern analysis)
- Symbol-aware trace output (Oct 2025): All diagnostic traces now show function/label names instead of raw addresses for improved readability

## Additional Features

### Performance Analysis

Run programs with tracing and statistics:

```bash
# Execution tracing
./arm-emulator --trace --trace-file trace.txt program.s

# Memory tracing
./arm-emulator --mem-trace --mem-trace-file mem_trace.txt program.s

# Performance statistics
./arm-emulator --stats --stats-file stats.html --stats-format html program.s
```

### Diagnostic Modes

Run programs with advanced diagnostic tracking:

```bash
# Code coverage - track which instructions were executed
./arm-emulator --coverage --coverage-format text program.s

# Stack trace - monitor stack operations and detect overflow/underflow
./arm-emulator --stack-trace --stack-trace-format text program.s

# Flag trace - track CPSR flag changes for debugging conditional logic
./arm-emulator --flag-trace --flag-trace-format text program.s

# Register access pattern analysis - track register usage patterns
./arm-emulator --register-trace --register-trace-format text program.s

# Combine multiple diagnostic modes
./arm-emulator --coverage --stack-trace --flag-trace --register-trace --verbose program.s
```

Features:
- **Code Coverage**: Tracks executed vs unexecuted instructions, reports coverage percentage
- **Stack Trace**: Monitors all stack operations (push/pop/SP modifications), detects overflow/underflow
- **Flag Trace**: Records CPSR flag changes (N, Z, C, V) for each instruction that modifies flags
- **Register Trace**: Analyzes register access patterns, identifies hot registers, detects unused registers, and flags read-before-write issues
- **Symbol-Aware Output**: All traces automatically show function/label names (e.g., `main+4`, `calculate`) instead of raw hex addresses for easier debugging

All diagnostic modes support both text and JSON output formats.

Example symbol-aware output:
```
# Stack trace showing function names
[000005] nested_call         : MOVE      SP: 0x00050000 -> 0x0004FFEC  (grow by 20 bytes)
[000010] helper1             : MOVE      SP: 0x0004FFEC -> 0x0004FFD8  (grow by 20 bytes)

# Flag trace showing symbol names
[000012] loop                : 0xE355000C                      ---- -> N*---  (changed: N)

# Coverage showing functions with symbols
0x00008000: executed      1 times (first: cycle      1, last: cycle      1) [main]
0x00008014: executed     13 times (first: cycle     12, last: cycle    462) [loop]
```

### Example Programs Status

#### All Example Programs Working! (49 total, 100%) ✅

All 49 example programs execute successfully:

**Non-Interactive Programs (46):**
- hello.s, loops.s, arithmetic.s, conditionals.s, functions.s
- factorial.s, recursive_fib.s, recursive_factorial.s
- string operations: strings.s, string_reverse.s (with stdin)
- data structures: arrays.s, linked_list.s, hash_table.s
- sorting algorithms: quicksort.s, bubble_sort.s
- literal pools: test_ltorg.s, test_org_0_with_ltorg.s
- multi-precision: add_128bit.s (128-bit integer addition with carry propagation)
- And 27+ more fully functional examples

**Interactive Programs (3):**
These programs work correctly when provided with stdin input:
- **bubble_sort.s** - Prompts for array size and elements (fully working)
- **calculator.s** - Interactive calculator with +, -, *, / operations (fully working)
- **fibonacci.s** - Prompts for count of Fibonacci numbers (fully working)

#### Recent Fixes (Oct 2025)
- **calculator.s** - Fixed infinite loop bug when stdin exhausted (EOF handling)
- **test_ltorg.s** - Fixed literal pool space reservation in parser
- **test_org_0_with_ltorg.s** - Fixed literal pool space reservation + added missing branch instruction

### Development Tools

Located in `tools/` directory:

- **Linter** - Analyze assembly code for issues (`tools/lint.go`)
  - Undefined label detection with suggestions
  - Unreachable code detection
  - Register usage warnings
  - 25 tests
- **Formatter** - Format assembly code consistently (`tools/format.go`)
  - Multiple format styles (default, compact, expanded)
  - Configurable alignment and spacing
  - 27 tests
- **Cross-Reference** - Generate symbol usage reports (`tools/xref.go`)
  - Symbol cross-reference with usage tracking
  - Function and data label identification
  - 21 tests

### Machine Code Encoder/Decoder

Located in `encoder/` directory:

- **Encoder** - Convert assembly to ARM machine code
- **Decoder** - Disassemble ARM machine code to assembly
- Supports all ARM2 instruction formats
- 1148 lines across 5 files
- Complete encoding/decoding for data processing, memory, branch, and multiply instructions

### Configuration

Configuration files are stored in platform-specific locations:
- **macOS/Linux:** `~/.config/arm-emu/config.toml`
- **Windows:** `%APPDATA%\arm-emu\config.toml`

See `config/config.go` for all available options.

### Note

Coreutils is installed on MacOS, so commands like `gtimeout` are available.
