# ARM Emulator - a vibe coding project

[![CI](https://github.com/lookbusy1344/arm_emulator/actions/workflows/ci.yml/badge.svg)](https://github.com/lookbusy1344/arm_emulator/actions/workflows/ci.yml)

This is an example of using vibe coding to re-create my first commercial project, from 1992, which implemented an ARM2 emulator. The original code was written in Turbo Pascal for 16-bit MS-DOS and is completely lost.

Here I am attempting to use Claude Code to generate a Go implementation of an ARM2 emulator, with a simple TUI debugger.

## Background

ARM2 is the earliest commercial precursor to the AARCH64 architecture we all use in our smartphones, Macs and low-power Windows laptops. It started life in the mid-1980’s at the UK’s Acorn Computers.

The ARM1 (Acorn RISC Machine 1) was Acorn Computers' first microprocessor design. The ARM1 was the initial result of the Advanced Research and Development division Acorn Computers formed in order to advance the development of their own RISC processor. Design started in 1983, and when it was finished in 1985 the ARM1 was the simplest RISC processor produced worldwide.

Introduced in 1986, the ARM2 was capable of exceeding 10 MIPS when not bottlenecked by memory with an average of around 6 MIPS. Unlike the ARM1 which was predominantly a research project, the ARM2 became the first commercially successful ARM microprocessor.

The Acorn Archimedes family of personal computers was built using the ARM2 along with a number of fully custom support chips that were also designed by Acorn Computers.

https://en.wikichip.org/wiki/acorn/microarchitectures/arm1

## Initial prompt to Claude

*"Write a markdown file outline specification for a ARM2 assembly language emulator. Actually producing machine code is not initially important, the assembly language file should be interpreted and run by a simple virtual machine environment. We also need a debugger with a TUI, allowing single step, step over/into, and watching memory locations and registers and viewing the call stack."*

## Later prompts

IMPLEMENTATION_PLAN.md

PROGRESS.md

TODO.md

Prompt for each phase: *"Let’s implement phase X from IMPLEMENTATION_PLAN.md considering PROGRESS.md, and implement appropriate tests. Anything that you cannot fix, note in TODO.md"*

## Overview

An ARM emulator written in Go that implements a subset of the ARM2 instruction set.

## Documentation

### Project Documentation
- [SPECIFICATION.md](SPECIFICATION.md) - Detailed specification for the ARM2 emulator
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) - Implementation roadmap and plan
- [PROGRESS.md](PROGRESS.md) - Development progress and completed phases

### User Documentation
- [docs/debugger_reference.md](docs/debugger_reference.md) - Complete debugger command reference and guide
- [docs/assembly_reference.md](docs/assembly_reference.md) - ARM2 assembly language reference
- [docs/architecture.md](docs/architecture.md) - System architecture and design
- [examples/README.md](examples/README.md) - Example programs and usage instructions

## Features

- ARM2 instruction set implementation with 493 passing tests
- Assembly parser for ARM assembly programs
- Interactive debugger with TUI (Text User Interface)
- Virtual machine execution environment
- Cross-platform configuration management (TOML)
- Execution and memory tracing with filtering
- Performance statistics (JSON/CSV/HTML export)
- Development tools (linter, formatter, cross-reference generator)

## Prerequisites

- Go 1.25 or higher

## Installation

Clone the repository and build the project:

```bash
git clone <repository-url>
cd arm_emulator
go build -o arm-emulator
```

## Usage

### Running Assembly Programs

Run an ARM assembly program directly:

```bash
./arm-emulator program.s
```

The emulator will execute the program starting from `_start` (or `main` if `_start` is not found). The program runs until it encounters a `SWI #0x00` (exit) instruction or an error occurs.

### Using the Debugger

The emulator includes a powerful debugger with both command-line and TUI (Text User Interface) modes:

```bash
# Command-line debugger mode
./arm-emulator --debug program.s

# TUI mode with visual panels for source, registers, memory, etc.
./arm-emulator --tui program.s
```

**Quick debugger commands:**
- `run` (r) - Start/restart program execution
- `step` (s) - Execute one instruction (step into)
- `next` (n) - Execute one instruction (step over)
- `continue` (c) - Continue until breakpoint or exit
- `break <location>` (b) - Set breakpoint at label or address
- `print <expr>` (p) - Evaluate expression (registers, memory, etc.)
- `info registers` (i r) - Show all registers
- `help` - Show all available commands

**TUI keyboard shortcuts:**
- `F5` - Continue execution
- `F9` - Toggle breakpoint
- `F10` - Step over
- `F11` - Step into

For complete debugger documentation including conditional breakpoints, watchpoints, memory examination, and expression syntax, see [docs/debugger_reference.md](docs/debugger_reference.md).

### Performance Analysis

The emulator includes built-in tracing and statistics capabilities:

```bash
# Enable execution tracing
./arm-emulator --trace --trace-file trace.txt program.s

# Enable memory access tracing
./arm-emulator --mem-trace --mem-trace-file mem_trace.txt program.s

# Generate performance statistics
./arm-emulator --stats --stats-file stats.html --stats-format html program.s
```

**Performance features:**
- Execution trace with register changes and timing
- Memory access tracking (reads/writes)
- Instruction frequency analysis
- Branch statistics and prediction
- Function call profiling
- Hot path analysis
- Export to JSON, CSV, or HTML formats

### Example Programs

The `examples/` directory contains sample ARM assembly programs that demonstrate various features:

- **times_table.s** - Multiplication table generator
- **factorial.s** - Recursive factorial calculator
- **fibonacci.s** - Fibonacci sequence generator
- **string_reverse.s** - String reversal program
- **bubble_sort.s** - Bubble sort algorithm
- **calculator.s** - Interactive calculator with basic operations

See [examples/README.md](examples/README.md) for detailed descriptions and usage instructions.

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
