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

- [SPECIFICATION.md](SPECIFICATION.md) - Detailed specification for the ARM2 emulator
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) - Implementation roadmap and plan

## Features

- ARM2 instruction set implementation
- Assembly parser for ARM assembly programs
- Debugging utilities
- Virtual machine execution environment

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

Run the emulator:

```bash
./arm-emulator
```

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
