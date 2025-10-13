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

Once I checked the specification, I asked Claude to produce a staged implementation plan, breaking the project down into manageable phases in `IMPLEMENTATION_PLAN.md`. It produced 10 phases, which looked reasonable.

The prompt for each phase was *"Let’s implement phase X from IMPLEMENTATION_PLAN.md, documenting completed work in PROGRESS.md, and implement appropriate tests. Anything that you cannot fix, note in TODO.md"*

## Daily progress

**Day 1 - 8 Oct** - Claude has written a specification, and a staged implemenation plan. It's made good progress with phases 1-5 completed.

**Day 2 - 9 Oct** - Phases 6-10 completed. From the original plan the project should be essentially complete, but there is actually much more to do, including the parser. I have directed Claude to note in TODO.md anything it cannot complete, as frequently it will truncate a complex task and then 'forget' about the more difficult features left unfinished.

**Day 3 - 10 Oct** - Go code is about 25,000 lines using command:

```
find . -name "*.go" -type f -exec cat {} + | wc -l
```

**Day 4 - 11 Oct** - Whats becoming increasingly clear is that, although Claude is very impressive and has done great things in only 4 days (3 hours per day), it can get lost in the weeds. It has a tendency to 'fix' tests by removing them, or simplifying them. It sometimes loses sight of the big picture and currently several of the test programs to not operate correctly (but Claude hasn't noticed).

Today I have focused on getting the example programs (written by Claude) to run properly, which acts as good integration testing.

Go code is 28,331 lines long at the end of day. Weekly Claude usages limits are being approached.

**Day 5 - 12 Oct** - Asking Claude if any core ARM2 instructions are left to implement (apparently not), and to write comprehensive tests for all instructions. This has found several instructions that are malfunctioning, and it’s important to keep reminding Claude not to simplify or delete tests that fail, but to address the underlying issue. However the project is beginning to look impressive now, and I still haven't written or edited one line of Go here, just markdown files and prompts to Claude.

Today I have focused on asking Claude to write challanging example programs, but not at this stage getting them to run (to avoid issues with Claude just deleting things that don't work). I have also asked it to write comprehensive tests for all instructions, again without focusing too much on whether they pass.

Go code is 33,461 lines. Weekly Claude usage limits reached, can resume again on Thursday 16 Oct. In the interim I will use Sonnet 4.5 with Copilot.

It's important to have clear daily (or at least progressive) goals, so we can keep Claude focused on them.

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

- **Complete ARM2 instruction set implementation** with 1016 passing tests (100% pass rate)
  - All 16 data processing instructions (AND, EOR, SUB, RSB, ADD, ADC, SBC, RSC, TST, TEQ, CMP, CMN, ORR, MOV, BIC, MVN)
  - All memory operations (LDR/STR/LDRB/STRB/LDM/STM + halfword extensions)
  - All branch instructions (B/BL/BX)
  - Multiply instructions (MUL/MLA)
  - All ARM2 addressing modes (immediate, register, shifted, pre/post-indexed)
  - Software interrupts with 30+ syscalls
- Assembly parser for ARM assembly programs with macros and preprocessor
- Machine code encoder/decoder for binary ARM instruction formats
- Interactive debugger with TUI (Text User Interface)
- Virtual machine execution environment
- Cross-platform configuration management (TOML)
- Execution and memory tracing with filtering
- Performance statistics (JSON/CSV/HTML export)
- **Diagnostic modes: code coverage, stack trace, flag trace**
- Development tools (linter, formatter, cross-reference generator)

## Prerequisites

- Go 1.25 or higher
- Supported platforms: macOS, Linux, Windows

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

### Diagnostic Modes

Advanced debugging tools to help identify and fix issues:

```bash
# Code coverage - track which instructions were executed
./arm-emulator --coverage program.s

# Stack trace - monitor stack operations and detect overflow/underflow
./arm-emulator --stack-trace program.s

# Flag trace - track CPSR flag changes for debugging conditional logic
./arm-emulator --flag-trace program.s

# Combine multiple diagnostic modes with verbose output
./arm-emulator --coverage --stack-trace --flag-trace --verbose program.s
```

**Diagnostic features:**

**Code Coverage:**
- Tracks executed vs unexecuted instructions
- Reports coverage percentage
- Shows execution counts for each address
- Records first and last execution cycle
- Identifies dead code and untested paths

**Stack Trace:**
- Monitors all stack operations (PUSH, POP, SP modifications)
- Tracks stack depth and maximum usage
- **Detects and warns on stack overflow/underflow**
- Detailed trace with addresses and values
- Helps identify stack-related bugs

**Flag Trace:**
- Tracks CPSR flag changes (N, Z, C, V)
- Only records actual changes for efficiency
- Shows before/after states with highlights
- Statistics on flag change frequency
- Helps debug conditional logic issues

All diagnostic modes support both text and JSON output formats:
```bash
# JSON output for programmatic analysis
./arm-emulator --coverage --coverage-format json program.s
./arm-emulator --stack-trace --stack-trace-format json program.s
./arm-emulator --flag-trace --flag-trace-format json program.s
```

### Example Programs

The `examples/` directory contains 23 sample ARM assembly programs that demonstrate various features:

**Basic Examples:**
- **hello.s** - Hello World program
- **arithmetic.s** - Basic arithmetic operations

**Algorithm Examples:**
- **fibonacci.s** - Fibonacci sequence generator
- **factorial.s** - Recursive factorial calculator
- **bubble_sort.s** - Bubble sort algorithm
- **binary_search.s** - Binary search implementation
- **gcd.s** - Greatest common divisor

**Data Structure Examples:**
- **arrays.s** - Array operations
- **linked_list.s** - Linked list implementation
- **stack.s** - Stack implementation
- **strings.s** - String manipulation

**Advanced Examples:**
- **functions.s** - Function calling conventions
- **conditionals.s** - If/else, switch/case patterns
- **loops.s** - For, while, do-while loops
- **addressing_modes.s** - ARM2 addressing modes demonstration

And more! See [examples/README.md](examples/README.md) for detailed descriptions and usage instructions.

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
├── main.go              # Entry point and CLI
├── vm/                  # Virtual machine implementation
├── parser/              # Assembly parser with preprocessor
├── instructions/        # Instruction implementations
├── encoder/             # Machine code encoder/decoder
├── debugger/            # Debugging utilities with TUI
├── config/              # Cross-platform configuration
├── tools/               # Development tools (lint, format, xref)
├── tests/               # Test files (1016 tests, 100% passing)
├── examples/            # Example ARM assembly programs (23 programs)
└── docs/                # User and developer documentation
```

## Instruction Set Completeness

This emulator provides **complete ARM2 instruction set coverage** as implemented in the original 1986 Acorn ARM2 processor. All core ARM2 instructions and addressing modes are fully functional and tested.

**What's NOT implemented (and why):**
- Long multiply instructions (UMULL/UMLAL/SMULL/SMLAL) - introduced in ARMv3M (ARM6), not ARM2
- PSR transfer instructions (MRS/MSR) - introduced in ARMv3, not ARM2
- Atomic swap instructions (SWP/SWPB) - introduced in ARMv2a (ARM3), not original ARM2
- Coprocessor instructions (CDP/LDC/STC/MCR/MRC) - optional in ARMv2, rarely used

For detailed analysis, see the "Missing ARM2/ARMv2 Instructions" section in [TODO.md](TODO.md).

## License

MIT License. See `LICENSE` file for details.
