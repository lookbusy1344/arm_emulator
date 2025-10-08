# ARM Emulator - a vibe codeing project

This is an example of using vibe codeing to re-create my first commercial project, from 1992, which implemented an ARM2 emulator. The original code was written in Turbo Pascal for MS-DOS and is completely lost.

Here I am attempting to use Claude Code to generate a Go implementation of an ARM2 emulator, with a simple TUI debugger.

## Initial prompt to Claude

*"Write a markdown file outline specification for a ARM2 assembly language emulator. Actually producing machine code is not initially important, the assembly language file should be interpreted and run by a simple virtual machine environment. We also need a debugger with a TUI, allowing single step, step over/into, and watching memory locations and registers and viewing the call stack."*

## Overview

An ARM emulator written in Go that implements a subset of the ARMv8 instruction set.

## Documentation

- [SPECIFICATION.md](SPECIFICATION.md) - Detailed specification for the ARM2 emulator
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) - Implementation roadmap and plan

## Features

- ARMv8 instruction set implementation
- Assembly parser for ARM assembly programs
- Debugging utilities
- Virtual machine execution environment

## Prerequisites

- Go 1.24 or higher

## Installation

Clone the repository and build the project:

```bash
git clone <repository-url>
cd arm_emulator
go build -o arm-emulator
```

## Usage

Run the emulator:

```bash
./arm-emulator
```

## Development

### Building

```bash
go build -o arm-emulator
```

### Formatting

```bash
go fmt ./...
```

### Testing

```bash
go test ./...
```

## Project Structure

```
.
├── main.go              # Entry point
├── vm/                  # Virtual machine implementation
├── parser/              # Assembly parser
├── instructions/        # Instruction implementations
├── debugger/            # Debugging utilities
├── tests/               # Test files
├── examples/            # Example ARM assembly programs
└── tools/               # Utility tools
```

## License

MIT License. See `LICENSE` file for details.
