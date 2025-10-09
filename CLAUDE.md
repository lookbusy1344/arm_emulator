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

- `main.go` - Entry point
- `vm/` - Virtual machine implementation
- `parser/` - Assembly parser
- `instructions/` - Instruction implementations
- `debugger/` - Debugging utilities
- `tests/` - Test files
- `examples/` - Example ARM assembly programs
- `tools/` - Utility tools

## Development Guidelines

**IMPORTANT:** After implementing each phase of development, update `PROGRESS.md` to reflect the completed work, including:
- Mark the phase as completed
- Document any implementation details or deviations from the original plan
- Update the status of related tasks

**IMPORTANT:** Always run `go fmt ./...` and `go test ./...` after making changes to ensure code quality and correctness.

**IMPORTANT:** Do not delete failing tests without explicit instructions.

**IMPORTANT:** Anything that cannot be implemented should be noted in `TODO.md` with details so work can result later.

## Additional Features (Phase 10)

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
- **Formatter** - Format assembly code consistently (`tools/format.go`)
- **Cross-Reference** - Generate symbol usage reports (`tools/xref.go`)

### Configuration

Configuration files are stored in platform-specific locations:
- **macOS/Linux:** `~/.config/arm-emu/config.toml`
- **Windows:** `%APPDATA%\arm-emu\config.toml`

See `config/config.go` for all available options.
