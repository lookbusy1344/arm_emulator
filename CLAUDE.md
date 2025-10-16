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

## Run Command

```bash
./arm-emulator
```

## Project Structure

- `main.go` - Entry point and CLI interface
- `vm/` - Virtual machine implementation (CPU, memory, execution, syscalls, tracing, statistics)
- `parser/` - Assembly parser with preprocessor and macros
- `instructions/` - Instruction implementations (data processing, memory, branch, multiply)
- `encoder/` - Machine code encoder/decoder for binary ARM instructions
- `debugger/` - Debugging utilities with TUI (breakpoints, watchpoints, expression evaluation)
- `config/` - Cross-platform configuration management
- `tools/` - Development tools (linter, formatter, cross-reference generator)
- `tests/` - Test files (660 tests, 100% pass rate)
- `examples/` - Example ARM assembly programs (23 programs, 21 fully functional)
- `docs/` - User and developer documentation

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

## Current Status

**Phase 11 (Production Hardening) - Complete ✅**
- All tests passing (100% pass rate)
  - Unit tests: ~900 tests
  - Integration tests: 62 tests (full end-to-end execution pipeline + CLI diagnostic flags + example programs)
- Code quality tools implemented (golangci-lint with errcheck, unused, govet, ineffassign, misspell)
- All lint issues resolved (0 issues reported)
- Go vet warnings fixed (method renames to avoid interface conflicts)
- CI updated to Go 1.25 with automated linting
- Build artifacts added to .gitignore
- Parser limitations resolved (debugger expression parser rewritten)
- Example programs: 34 programs with comprehensive integration tests
  - 32 programs with expected output files
  - Table-driven test framework for easy test maintenance
- Recent improvements (Oct 2025):
  - Negative constants in .equ directives now supported
  - MOVW automatic encoding for 16-bit immediates
  - CMP/CMN instruction handling for un-encodable immediates
  - Shift operations work correctly as operand modifiers (LSR, LSL, ASR, ROR in MOV/ADD/etc.)
  - Data section ordering fixed (.data before .text)
  - Comprehensive integration test coverage for all examples
- Character literal support complete (basic chars + escape sequences)
- ARM immediate encoding bug fixed (fibonacci.s, calculator.s now work correctly)
- Diagnostic modes complete (code coverage, stack trace, flag trace)

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

# Combine multiple diagnostic modes
./arm-emulator --coverage --stack-trace --flag-trace --verbose program.s
```

Features:
- **Code Coverage**: Tracks executed vs unexecuted instructions, reports coverage percentage
- **Stack Trace**: Monitors all stack operations (push/pop/SP modifications), detects overflow/underflow
- **Flag Trace**: Records CPSR flag changes (N, Z, C, V) for each instruction that modifies flags

All diagnostic modes support both text and JSON output formats.

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
