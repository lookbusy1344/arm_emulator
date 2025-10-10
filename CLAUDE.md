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

## Test Command

```bash
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
- `tests/` - Test files (509 tests, 99.6% pass rate)
- `examples/` - Example ARM assembly programs (17 complete programs)
- `docs/` - User and developer documentation

## Development Guidelines

**IMPORTANT:** After implementing each phase of development, update `PROGRESS.md` to reflect the completed work, including:
- Mark the phase as completed
- Document any implementation details or deviations from the original plan
- Update the status of related tasks

**IMPORTANT:** Always run `go fmt ./...` and `go test ./...` after making changes to ensure code quality and correctness.

**IMPORTANT:** Do not delete failing tests without explicit instructions.

**IMPORTANT:** Anything that cannot be implemented should be noted in `TODO.md` with details so work can result later.

## Current Status

**Phase 11 (Production Hardening) - In Progress**
- 509 tests passing (99.6% pass rate)
- Go vet warnings fixed (method renames to avoid interface conflicts)
- CI updated to Go 1.25
- Build artifacts added to .gitignore
- Parser limitations resolved (debugger expression parser rewritten)

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
