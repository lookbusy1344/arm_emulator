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

Claude can sometimes suffer from groupthink, just confirming the code is excellent without looking afresh. This prompt helps *"Look at it with fresh eyes. The engineer implemented it suspiciously quickly and I do not trust their work"*

## Daily progress

**Day 1 - 8 Oct** - Claude has written a specification, and a staged implementation plan. It's made good progress with phases 1-5 completed.

**Day 2 - 9 Oct** - Phases 6-10 completed. From the original plan the project should be essentially complete, but there is actually much more to do, including the parser. I have directed Claude to note in TODO.md anything it cannot complete, as frequently it will truncate a complex task and then 'forget' about the more difficult features left unfinished.

**Day 3 - 10 Oct** - Go code is about 25,000 lines using command:

```
find . -name "*.go" -type f -exec cat {} + | wc -l
```

**Day 4 - 11 Oct** - Whats becoming increasingly clear is that, although Claude is very impressive and has done great things in only 4 days (3 hours per day), it can get lost in the weeds. It has a tendency to 'fix' tests by removing them, or simplifying them. It sometimes loses sight of the big picture and currently several of the test programs to not operate correctly (but Claude hasn't noticed).

Today I have focused on getting the example programs (written by Claude) to run properly, which acts as good integration testing.

Go code is 28,331 lines long at the end of day. Weekly Claude usages limits are being approached.

**Day 5 - 12 Oct** - Asking Claude if any core ARM2 instructions are left to implement (apparently not), and to write comprehensive tests for all instructions. This has found several instructions that are malfunctioning, and it’s important to keep reminding Claude not to simplify or delete tests that fail, but to address the underlying issue. However the project is beginning to look impressive now, and I still haven't written or edited one line of Go here, just markdown files and prompts to Claude.

Today I have focused on asking Claude to write challenging example programs, but not at this stage getting them to run (to avoid issues with Claude just deleting things that don't work). I have also asked it to write comprehensive tests for all instructions, again without focusing too much on whether they pass.

Go code is 33,461 lines. Weekly Claude usage limits reached, can resume again on Thursday 16 Oct. In the interim I will use Sonnet 4.5 with Copilot.

It's important to have clear daily (or at least progressive) goals, so we can keep Claude focused on them.

**Day 6 - 13 Oct - Copilot VSCode** - Switched to Sonnet 4.5 in Copilot interactively with VSCode, because Claude Code weekly limits have been reached. While making progress, several more example programs are failing without detection from Copilot. The automated tests are clearly not thorough enough.

I've been working through the failing example programs, and getting them to run without Copilot editing out the parts that don't work. This is a slow but I'm making progress.

Last step of the day was to take more detailed control, and tell Copilot to include integration tests that run whole example programs, comparing the output to expected output files. This is a good way to catch problems that unit tests miss (although some example programs are still failing). This is the first code I have actually looked at in detail and directed more closely, eg 9262a29b2373970592ad.

Go code now is 34,735 lines.

**Day 7 - 14 Oct - Copilot CLI** - Today I’m trying Copilot CLI. https://github.com/github/copilot-cli with the default Sonnet 4.5 model, from commit c3d1c0ada6fbf073e0. Not as slick as Claude Code, and needs more confirmations.

Commit f19616250600ed4ed9883 breaks integration tests (the script runs fine by hand), and Sonnet 4.5 completely failed to fix it. Switching to GPT-5 however seemed to fixed very elegantly, although it is slow and uncommunicative. The solution just appeared after minutes of silence!

By 2001523f91760431f078e we have all the example programs fixed, and running as integration tests. Something I've closely pushed for in the latest phase of development.

**Day 8 - 15 Oct - Copilot CLI** - More integration tests. Updating the documents. Although not as polished as Claude, Copilot has the choice of models which can be useful.

Go code now is 35,206 lines.

**Day 9 - 16 Oct - Claude Code** - Back on Claude Code, far superior to Copilot. Adding some missing assembly instructions, and automated testing the TUI interface, and Register Access Pattern Analysis.

Go code is now 40,352 lines and we have 75% code coverage. Note the big jump in output.

**Day 10 - 17 Oct** - The project is almost finished and ready for detailed review, a substantial task for over 40k lines of code! Today I have mainly focused on docs, checked through the tutorial and making a list of specific fixes and clarifications I’d like Claude to make. Also worked on a solution for the CPSR issue with 32 bit addresses (this is a small departure from strict ARM2 that used 26 bit addressing, and stored CPU flags in the remaining bits of the PC register). This emulator uses full 32 bit addressing, so we need to store the CPSR flags separately. This is only a theoretical issue for most example programs, because there are no hardware interrupts.

Go code is now 42,481 lines. Estimated total vibing time is around 32 hours, so an average of 1,300 lines of code per hour!

All unit and integration tests are passing, and the test system runs all the example programs and confirms their output.

Release automation added, and v0.9.0 tagged. We now have automatic builds for 4 platfroms. Perhaps at this point I should actually try it with some assembly I have written myself 😂

**Day 11 - 18 Oct** - Telling Claude to take a fresh look with prompt *"Look at it with fresh eyes. The engineer implemented it suspiciously quickly and I do not trust their work"* created a good PR that apparently fixes some serious bugs from commits 233b2d5 to 93e7fa0. Interesting!

Also working on TUI debugging, which had a number of problems and doesn't seem to be well tested. By end of day it's much more usable.

Go code is now 44,073 lines.

**Day 12 - 19 Oct** - Final polish on the TUI, highlighting altered memory locations in green and showing labels properly in the Source window. UI testing is much slower because I need to manually run it, observe changes I'd like, ask Claude and repeat the loop.

Go code is now 44,276 lines. About 3,600 lines per day, and that's only in a few hours per day!

## Summary, some thoughts

At this stage I’ve probably taken vibing as far as I can go without actually writing some ARM2 assembly and trying to run and debug it by hand. So far every test program has been written by Claude (or Copilot, when Claude limits were exceeded).

In terms of developer experience, *Claude Code* is amazing for backend/API and when testing can be automated. It’s easy to get jaded and blasé about current AI progress, but if you told me 2 years ago this would be possible I would have said you were dreaming! It's not so good when work is highly visual (the TUI debugger), and automated testing doesn't seem very useful. This would probably extend to websites.

My criticisms are minor. You need to lean on Claude to preserve tests that fail, and actually fix the issue not delete the test. Instructions in CLAUDE.md should be strong in this regard. The only time I’ve applied detailed pressure and closely monitored code was when instructing Claude to write comprehensive integration tests, running the example programs and checking the output against expected results. Even then, some failing examples survived unnoticed between days 7-10. This might be less of an issue when Claude is writing something less esoteric and more easily seen, like a website. In this project, the goals were always loosely defined and many example programs were pretty open-ended.

Giving Claude strong instructions to "Look at it with fresh eyes" is vital for getting good code reviews.

Claude continues to evolve quickly. Even during these 2 weeks, it has had 2 significant updates: first with Sonnet 4.5 and then today the faster Haiku 4.5 for simple tasks. Haiku completed small doc updates and automatic release workflows almost instantly.

Exciting times. Perhaps I should think of a second more challenging vibe-coding project!

**The rest of this document is AI generated.**

## Documentation

### Project Documentation
- [SPECIFICATION.md](SPECIFICATION.md) - Detailed specification for the ARM2 emulator
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) - Implementation roadmap and plan
- [PROGRESS.md](PROGRESS.md) - Development progress and completed phases

### User Documentation
- [docs/installation.md](docs/installation.md) - Installation guide and setup
- [docs/TUTORIAL.md](docs/TUTORIAL.md) - Step-by-step tutorial for learning ARM2 assembly
- [INSTRUCTIONS.md](INSTRUCTIONS.md) - Complete ARM2 instruction set reference
- [docs/assembly_reference.md](docs/assembly_reference.md) - ARM2 assembly language reference (directives, syntax)
- [examples/README.md](examples/README.md) - Example programs and usage instructions (48 programs)
- [docs/debugger_reference.md](docs/debugger_reference.md) - Complete debugger command reference and guide
- [docs/debugging_tutorial.md](docs/debugging_tutorial.md) - Hands-on debugging tutorials with examples
- [docs/FAQ.md](docs/FAQ.md) - Frequently asked questions and troubleshooting

### Developer Documentation
- [docs/API.md](docs/API.md) - API reference for developers
- [docs/architecture.md](docs/architecture.md) - System architecture and design
- [docs/ltorg_implementation.md](docs/ltorg_implementation.md) - Literal pool implementation details

## Features

- **Complete ARM2 instruction set implementation** with 969 passing tests (100% pass rate, 75% code coverage)
  - All 16 data processing instructions (AND, EOR, SUB, RSB, ADD, ADC, SBC, RSC, TST, TEQ, CMP, CMN, ORR, MOV, BIC, MVN)
  - All memory operations (LDR/STR/LDRB/STRB/LDM/STM + halfword extensions)
  - All branch instructions (B/BL/BX)
  - Multiply instructions (MUL/MLA)
  - All ARM2 addressing modes (immediate, register, shifted, pre/post-indexed)
  - Software interrupts with 35+ syscalls (console I/O, file operations, memory management, system information, debugging support)
- Assembly parser for ARM assembly programs with macros and preprocessor
- **Dynamic literal pool sizing**: Smart allocation based on actual literal usage, not fixed estimates
  - Counts LDR pseudo-instructions per `.ltorg` directive
  - Adjusts pool addresses for optimal space utilization
  - Validation warnings for pools exceeding capacity
  - Support for 20+ literals per pool (tested up to 33)
- Machine code encoder/decoder for binary ARM instruction formats
- Interactive debugger with TUI (Text User Interface)
- Virtual machine execution environment
- Cross-platform configuration management (TOML)
- Execution and memory tracing with filtering
- Performance statistics (JSON/CSV/HTML export)
- **Diagnostic modes: code coverage, stack trace, flag trace, register access pattern analysis**
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
- `F9` - Toggle breakpoint at current line
- `F10` - Step over
- `F11` - Step into
- `Ctrl+L` - Refresh display
- `Tab` - Switch between panels

**TUI visual features:**
- **Register highlighting** - Changed registers shown in green
- **Memory write highlighting** - Written memory bytes shown in green (auto-scrolls to written address)
- **Stack highlighting** - PUSH/POP operations highlighted in green
- **Symbol-aware display** - Function/label names shown instead of raw addresses
- **Source view** - Shows current line with `>` indicator, handles labels and comments properly
- **Multi-panel layout** - Source, Registers, Memory, Stack, Breakpoints, Watchpoints, Console

For complete debugger documentation including conditional breakpoints, watchpoints, memory examination, and expression syntax, see [docs/debugger_reference.md](docs/debugger_reference.md).

### Symbol Table Dump

Inspect the parsed symbols from your assembly program:

```bash
# Dump symbol table to stdout
./arm-emulator --dump-symbols program.s

# Save symbol table to a file
./arm-emulator --dump-symbols --symbols-file symbols.txt program.s
```

The symbol dump displays all labels, constants, and variables with their addresses, types, and definition status. This is useful for understanding program layout and debugging symbol resolution issues.

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

# Register trace - analyze register access patterns
./arm-emulator --register-trace program.s

# Combine multiple diagnostic modes with verbose output
./arm-emulator --coverage --stack-trace --flag-trace --register-trace --verbose program.s
```

**Diagnostic features:**

**Code Coverage:**
- Tracks executed vs unexecuted instructions with symbol names
- Reports coverage percentage
- Shows execution counts for each address
- Records first and last execution cycle
- Identifies dead code and untested paths
- Symbol-aware output (e.g., `0x00008000: executed 1 times [main]`)

**Stack Trace:**
- Monitors all stack operations (PUSH, POP, SP modifications)
- Tracks stack depth and maximum usage
- **Detects and warns on stack overflow/underflow**
- Detailed trace with addresses and values
- Symbol-aware output showing function names (e.g., `[000005] nested_call : MOVE SP: 0x00050000 -> 0x0004FFEC`)
- Helps identify stack-related bugs

**Flag Trace:**
- Tracks CPSR flag changes (N, Z, C, V)
- Only records actual changes for efficiency
- Shows before/after states with highlights
- Statistics on flag change frequency
- Symbol-aware output showing labels (e.g., `[000012] loop : 0xE355000C ---- -> N*---`)
- Helps debug conditional logic issues

**Register Access Pattern Analysis:**
- Tracks read/write patterns for all registers
- Identifies "hot" registers (most frequently accessed)
- Detects unused registers
- Flags read-before-write issues (potential uninitialized use)
- Shows unique value counts and access sequences
- Helps optimize register allocation and find bugs

All diagnostic modes support both text and JSON output formats:
```bash
# JSON output for programmatic analysis
./arm-emulator --coverage --coverage-format json program.s
./arm-emulator --stack-trace --stack-trace-format json program.s
./arm-emulator --flag-trace --flag-trace-format json program.s
./arm-emulator --register-trace --register-trace-format json program.s
```

### Example Programs

The `examples/` directory contains 44 sample ARM assembly programs that demonstrate various features (100% working):

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
- **add_128bit.s** - 128-bit integer addition with carry propagation

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
go clean -testcache
go test ./...
```

### Release Builds

Create optimized release builds for distribution:

**Local optimized build:**
```bash
go build -ldflags="-s -w" -o arm-emulator
```

The `-ldflags="-s -w"` flags strip symbol tables and debug information, producing smaller, faster-loading binaries suitable for distribution (~30-40% size reduction).

**Automated release builds:**

The project includes automated GitHub Actions workflows that create optimized release builds for multiple platforms:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This triggers the `Build Release` workflow which:
- Builds optimized binaries for **linux-amd64**, **darwin-arm64**, **windows-amd64**, and **windows-arm64**
- Generates SHA256 checksums for each binary
- Creates a GitHub Release with pre-built binaries, individual checksums, and a combined SHA256SUMS file
- Users can download platform-specific binaries directly from the [Releases](https://github.com/lookbusy1344/arm_emulator/releases) page

**Verifying downloads:**

To verify the integrity of a downloaded binary, use the SHA256 checksums provided in the release:

```bash
# On Linux/macOS - verify using the combined SHA256SUMS file
sha256sum -c SHA256SUMS --ignore-missing

# On Linux/macOS - verify a specific binary
sha256sum arm-emulator-linux-amd64
# Compare the output with the checksum in the .sha256 file

# On Windows (PowerShell)
Get-FileHash arm-emulator-windows-amd64.exe -Algorithm SHA256
# Compare the output with the checksum in the .sha256 file
```

Each release includes:
- Individual `.sha256` files for each binary (e.g., `arm-emulator-linux-amd64.sha256`)
- A combined `SHA256SUMS` file containing all checksums for easy verification

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
├── tests/               # Test files (969 tests, 100% passing, 75% coverage)
├── examples/            # Example ARM assembly programs (44 programs)
└── docs/                # User and developer documentation
```

## Instruction Set Completeness

This emulator provides **complete ARM2 instruction set coverage** as implemented in the original 1986 Acorn ARM2 processor. All core ARM2 instructions and addressing modes are fully functional and tested.

**Beyond ARM2 - Additional instructions implemented:**
- **Long multiply instructions (UMULL/UMLAL/SMULL/SMLAL)** - introduced in ARMv3M (ARM6), fully implemented with 64-bit results
- **PSR transfer instructions (MRS/MSR)** - introduced in ARMv3, implemented for CPSR flag manipulation

**What's NOT implemented:**
- Atomic swap instructions (SWP/SWPB) - introduced in ARMv2a (ARM3), not original ARM2
- Coprocessor instructions (CDP/LDC/STC/MCR/MRC) - optional in ARMv2, rarely used

## License

MIT License. See `LICENSE` file for details.
