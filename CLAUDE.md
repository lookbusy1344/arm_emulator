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
