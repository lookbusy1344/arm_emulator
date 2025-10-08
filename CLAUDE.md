# ARM Emulator Project

This is an ARM emulator written in Go that implements a subset of the ARMv8 instruction set.

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
