# ARM2 Emulator Implementation Plan

## Overview

This document outlines the implementation plan for the ARM2 Assembly Language Emulator with integrated debugger, as specified in SPECIFICATION.md. The project will be implemented in Go with a focus on cross-platform compatibility (macOS, Windows, Linux).

## Phase 1: Foundation (Weeks 1-2)

### 1. Project Setup

**Goals:**
- Initialize Go module with cross-platform support
- Set up dependencies
- Create directory structure
- Configure CI/CD pipeline

**Dependencies:**
- `github.com/rivo/tview` - TUI framework
- `github.com/gdamore/tcell/v2` - Terminal handling
- `github.com/spf13/cobra` - CLI framework
- `github.com/BurntSushi/toml` - Configuration file parsing

**Directory Structure:** (per SPECIFICATION.md:308-340)
```
emulator/
├── vm/
│   ├── cpu.go           # CPU state and register management
│   ├── memory.go        # Memory management
│   ├── executor.go      # Instruction execution engine
│   └── flags.go         # CPSR flag operations
├── parser/
│   ├── lexer.go         # Tokenization
│   ├── parser.go        # Syntax analysis
│   ├── assembler.go     # Convert to IR
│   ├── symbols.go       # Symbol table management
│   ├── macros.go        # Macro processing
│   └── preprocessor.go  # Handle includes and conditional assembly
├── debugger/
│   ├── debugger.go      # Main debugger logic
│   ├── tui.go           # Text UI implementation
│   ├── commands.go      # Command parser and handlers
│   ├── breakpoints.go   # Breakpoint management
│   ├── watchpoints.go   # Watchpoint management
│   ├── expressions.go   # Expression evaluator for conditions
│   └── history.go       # Command and execution history
├── instructions/
│   ├── data_processing.go
│   ├── multiply.go
│   ├── memory.go
│   ├── memory_multi.go  # Load/store multiple
│   ├── branch.go
│   └── syscall.go       # System call implementations
├── tools/
│   ├── lint.go          # Assembly linter
│   ├── format.go        # Code formatter
│   ├── xref.go          # Cross-reference generator
│   └── disassembler.go  # Disassembler (for binary to assembly)
└── main.go
```

**CI/CD Setup:**
- GitHub Actions workflow for all platforms (macOS, Windows, Linux)
- Automated testing on every commit
- Coverage reporting
- Cross-compilation builds

### 2. Core VM Implementation

**Components:**

**vm/cpu.go:**
- 16 general-purpose registers (R0-R15)
- R13 = Stack Pointer (SP)
- R14 = Link Register (LR)
- R15 = Program Counter (PC)
- CPSR (Current Program Status Register) with flags:
  - N (Negative)
  - Z (Zero)
  - C (Carry)
  - V (Overflow)

**vm/memory.go:**
- 4GB addressable space (32-bit addressing)
- Memory segments:
  - Code segment (read-only after load)
  - Data segment (read-write)
  - Stack segment (grows downward)
  - Heap segment (grows upward)
- Little-endian by default
- Access granularity: byte, halfword (16-bit), word (32-bit)

**vm/flags.go:**
- Flag calculation helpers for all operations
- N flag: bit 31 of result
- Z flag: result == 0
- C flag: carry out for arithmetic, last bit shifted out for shifts
- V flag: signed overflow detection

**vm/executor.go:**
- Fetch-decode-execute cycle skeleton
- Execution modes: run, step, step over, step into

**Memory Protection Features:**
- Alignment checking (word=4-byte, halfword=2-byte)
- Bounds checking
- Null pointer detection (address 0x0000)
- Read/write/execute permissions by segment
- Stack overflow/underflow detection

---

## Phase 2: Parser & Assembler (Weeks 3-4)

### 3. Lexer & Parser

**parser/lexer.go:**
- Tokenize assembly syntax
- Handle comments: `; comment`, `// comment`, `/* block */`
- Recognize keywords, registers, directives, labels
- Support case-insensitive instructions, case-sensitive labels

**parser/parser.go:**
- Parse instruction format: `LABEL: MNEMONIC{COND}{S} operands ; comment`
- Label types:
  - Global labels (start at column 0)
  - Local labels (start with `.`)
  - Numeric labels (`1:`, `1b`, `1f`)
- Parse all directives:
  - `.org`, `.equ`, `.set`
  - `.word`, `.half`, `.byte`
  - `.ascii`, `.asciz`, `.string`
  - `.space`, `.skip`, `.align`, `.balign`
  - `.include`, `.macro`/`.endm`
  - `.if`, `.ifdef`, `.ifndef`, `.endif`
  - `.global`, `.extern`
  - `.section`, `.text`, `.data`, `.bss`

**parser/symbols.go:**
- Symbol table management
- Forward reference resolution (two-pass assembly)
- Relocation table for address resolution

**parser/preprocessor.go:**
- Handle `.include` directives
- Conditional assembly (`.if`, `.ifdef`, `.ifndef`)
- Detect circular includes
- Track nested includes

**parser/macros.go:**
- Macro definition and expansion
- Parameter substitution
- Macro expansion tracking for debugging

### 4. Error Handling

**Features:**
- Line and column position tracking
- Error messages with context
- Syntax error suggestions
- Undefined label detection
- Duplicate label warnings
- Invalid directive/instruction reporting

**Output Format:**
```
program.s:15:5: error: Undefined label 'typo_label'
        BL      typo_label
                ^~~~~~~~~~
Did you mean 'type_label'?
```

---

## Phase 3: Instruction Set (Weeks 5-7)

### 5. Data Processing Instructions

**instructions/data_processing.go:**

**Move Instructions:**
- `MOV` - Move
- `MVN` - Move Not

**Arithmetic Instructions:**
- `ADD` - Add
- `ADC` - Add with Carry
- `SUB` - Subtract
- `SBC` - Subtract with Carry
- `RSB` - Reverse Subtract
- `RSC` - Reverse Subtract with Carry

**Logical Instructions:**
- `AND` - Bitwise AND
- `ORR` - Bitwise OR
- `EOR` - Bitwise Exclusive OR
- `BIC` - Bit Clear

**Comparison Instructions:**
- `CMP` - Compare (SUB without result)
- `CMN` - Compare Negative (ADD without result)
- `TST` - Test (AND without result)
- `TEQ` - Test Equivalence (EOR without result)

**Addressing Modes:**
1. Immediate: `#value`
2. Register: `Rn`
3. Register with shift:
   - `Rm, LSL #shift` (Logical Shift Left)
   - `Rm, LSR #shift` (Logical Shift Right)
   - `Rm, ASR #shift` (Arithmetic Shift Right)
   - `Rm, ROR #shift` (Rotate Right)
4. Register shift by register: `Rm, LSL Rs`, etc.
5. Rotate right extended: `Rm, RRX`

**Condition Codes:**
- `EQ` (Equal, Z=1)
- `NE` (Not Equal, Z=0)
- `CS/HS` (Carry Set/Unsigned Higher or Same, C=1)
- `CC/LO` (Carry Clear/Unsigned Lower, C=0)
- `MI` (Minus/Negative, N=1)
- `PL` (Plus/Positive or Zero, N=0)
- `VS` (Overflow Set, V=1)
- `VC` (Overflow Clear, V=0)
- `HI` (Unsigned Higher, C=1 and Z=0)
- `LS` (Unsigned Lower or Same, C=0 or Z=1)
- `GE` (Signed Greater or Equal, N=V)
- `LT` (Signed Less Than, N≠V)
- `GT` (Signed Greater Than, Z=0 and N=V)
- `LE` (Signed Less or Equal, Z=1 or N≠V)
- `AL` (Always, unconditional)

**Flag Updates:**
- S bit support: instructions can optionally set flags
- Flag calculation for all operations
- Correct C and V flag handling for overflow cases

### 6. Memory Access Instructions

**instructions/memory.go:**

**Load/Store Word:**
- `LDR` - Load Register
- `STR` - Store Register

**Load/Store Byte:**
- `LDRB` - Load Register Byte
- `STRB` - Store Register Byte

**Load/Store Halfword:** (ARM2a extensions)
- `LDRH` - Load Register Halfword
- `STRH` - Store Register Halfword

**Addressing Modes:**
1. Offset: `[Rn, #offset]`
2. Pre-indexed: `[Rn, #offset]!` (update Rn before access)
3. Post-indexed: `[Rn], #offset` (update Rn after access)
4. Register offset: `[Rn, Rm]`
5. Register offset with shift: `[Rn, Rm, LSL #shift]`
6. Scaled register offset: `[Rn, Rm, shift_type #amount]`

**instructions/memory_multi.go:**

**Load/Store Multiple:**
- `LDM{mode}` - Load Multiple
- `STM{mode}` - Store Multiple

**Modes:**
- `IA` (Increment After)
- `IB` (Increment Before)
- `DA` (Decrement After)
- `DB` (Decrement Before)

**Stack Variants:**
- `FD` (Full Descending) - most common
- `ED` (Empty Descending)
- `FA` (Full Ascending)
- `EA` (Empty Ascending)

### 7. Branch & Multiply Instructions

**instructions/branch.go:**

**Branch Instructions:**
- `B{cond}` - Branch
- `BL{cond}` - Branch with Link (stores return address in LR)
- `BX{cond}` - Branch and Exchange (for future ARM/Thumb interworking)

**Features:**
- 24-bit signed offset (±32MB range)
- All condition codes supported
- Call stack tracking for debugger

**instructions/multiply.go:**

**Multiply Instructions:**
- `MUL` - Multiply
- `MLA` - Multiply-Accumulate

**Restrictions:**
- Rd and Rm must be different registers
- Result is lower 32 bits only
- Timing varies (2-16 cycles) based on operand values

---

## Phase 4: System Integration (Week 8)

### 8. System Calls (SWI Mechanism)

**instructions/syscall.go:**

**Console I/O:**
- `0x00 - Exit`: Terminate program (R0 = exit code)
- `0x01 - Write Char`: Output character (R0 = char value)
- `0x02 - Write String`: Output null-terminated string (R0 = string address)
- `0x03 - Write Int`: Output integer (R0 = value, R1 = base)
- `0x04 - Read Char`: Input character (returns in R0)
- `0x05 - Read String`: Input string (R0 = buffer address, R1 = max length)
- `0x06 - Read Int`: Input integer (returns in R0)
- `0x07 - Write Newline`: Output newline

**File Operations:**
- `0x10 - Open`: Open file (R0 = filename ptr, R1 = mode) → returns fd in R0
- `0x11 - Close`: Close file (R0 = fd)
- `0x12 - Read`: Read from file (R0 = fd, R1 = buffer, R2 = count)
- `0x13 - Write`: Write to file (R0 = fd, R1 = buffer, R2 = count)
- `0x14 - Seek`: Seek in file (R0 = fd, R1 = offset, R2 = whence)
- `0x15 - Tell`: Get current position (R0 = fd)
- `0x16 - FileSize`: Get file size (R0 = fd)

**Memory Operations:**
- `0x20 - Allocate`: Allocate heap memory (R0 = size) → returns address
- `0x21 - Free`: Free heap memory (R0 = address)
- `0x22 - Reallocate`: Resize allocation (R0 = address, R1 = new size)

**System Information:**
- `0x30 - Get Time`: Get current time in milliseconds
- `0x31 - Get Random`: Get random number
- `0x32 - Get Arguments`: Get command-line arguments
- `0x33 - Get Environment`: Get environment variable

**Error Handling:**
- `0x40 - Get Error`: Get last error code
- `0x41 - Set Error`: Set error code
- `0x42 - Print Error`: Print last error message

**Debugging Support:**
- `0xF0 - Debug Print`: Print debug message (R0 = string address)
- `0xF1 - Breakpoint`: Trigger debugger breakpoint
- `0xF2 - Dump Registers`: Print all register values
- `0xF3 - Dump Memory`: Print memory region (R0 = address, R1 = length)
- `0xF4 - Assert`: Assert condition (R0 = condition, R1 = message ptr)

**Error Codes:**
- Return values in R0: positive/zero = success, negative = error
- Standard errno-style error codes

### 9. Runtime Environment

**Bootstrap Sequence:**
1. Load code into code segment
2. Initialize data segment with static data
3. Set up stack at high memory address
4. Initialize heap space
5. Register initialization:
   - PC = Entry point (`_start`, `main`, or first instruction)
   - SP (R13) = Stack top address
   - LR (R14) = 0
   - R0-R12 = 0 (or argument values)
   - CPSR = 0

**Entry Point Detection:**
- Search order: `_start` → `main` → first instruction at origin

**Program Termination:**
- `SWI #0x00` with exit code in R0
- Reaching `_exit` label
- Infinite loop detection (debugger)
- Maximum instruction count exceeded
- Fatal error (invalid instruction, memory violation)

**Standard Library Macros:**
Create `stdlib.inc` with common macros:
- `PRINT_STR`, `PRINT_CHAR`, `PRINT_INT`
- `EXIT`, `ALLOC`, `FREE`
- Stack manipulation helpers

---

## Phase 5: Debugger Core (Weeks 9-10)

### 10. Debugger Foundation

**debugger/debugger.go:**
- Main debugger state machine
- VM integration
- Execution control flow
- State inspection interface

**debugger/commands.go:**

**Execution Control:**
- `run` / `r`: Run program until completion or breakpoint
- `step` / `s`: Execute single instruction (step into)
- `next` / `n`: Execute single instruction (step over calls)
- `stepi` / `si`: Step one instruction
- `continue` / `c`: Continue execution
- `finish`: Execute until current function returns
- `until <location>`: Run until location reached

**Breakpoint Commands:**
- `break <location>` / `b <location>`: Set breakpoint
- `tbreak <location>`: Set temporary breakpoint
- `delete <id>`: Delete breakpoint
- `disable <id>`: Disable breakpoint
- `enable <id>`: Enable breakpoint
- `info breakpoints` / `i b`: List all breakpoints

**Watchpoint Commands:**
- `watch <expression>`: Set write watchpoint
- `rwatch <expression>`: Set read watchpoint
- `awatch <expression>`: Set access watchpoint
- `watch register <Rn>`: Watch register changes
- `watch memory <address>`: Watch memory location
- `info watchpoints`: List all watchpoints

**Inspection Commands:**
- `print <expression>` / `p <expression>`: Evaluate and print expression
- `x/<format> <address>`: Examine memory
- `info registers` / `i r`: Display all registers
- `info stack` / `i s`: Display stack information
- `backtrace` / `bt`: Show call stack
- `list` / `l`: List source code around current location

**State Modification:**
- `set $Rn = <value>`: Set register value
- `set {<type>}<address> = <value>`: Set memory value

**Program Control:**
- `load <file>`: Load assembly file
- `reset`: Reset VM to initial state
- `quit` / `q`: Exit debugger

**debugger/breakpoints.go:**
- Address breakpoints
- Label breakpoints
- Conditional breakpoints (e.g., `break loop if $R0 == 100`)
- Temporary breakpoints (one-shot)
- Breakpoint enable/disable
- Breakpoint hit counting

**debugger/watchpoints.go:**
- Register watchpoints
- Memory watchpoints (read/write/access)
- Watch expression evaluation
- Trigger on value changes

**debugger/expressions.go:**
- Expression parser and evaluator
- Support for:
  - Register values (`$R0`, `$PC`)
  - Memory dereferencing (`[$SP]`)
  - Arithmetic operations
  - Comparisons
  - Logical operations

**debugger/history.go:**
- Command history (up/down arrows)
- History file persistence
- Execution history tracking
- Register value history

### 11. Call Stack Tracking

**Features:**
- Automatic detection of `BL` (function calls)
- Track return addresses (stored in LR)
- Display call hierarchy with frame information
- Frame selection for inspection
- Unwind stack on errors

---

## Phase 6: TUI Interface (Weeks 11-12)

### 12. TUI Implementation

**debugger/tui.go:**

**Layout Components:**
1. **Source View**: Display assembly source with current instruction highlighted
2. **Register View**: Show all registers (R0-R15, CPSR) with real-time updates
3. **Memory View**: Hexadecimal and ASCII display of memory regions
4. **Stack View**: Display stack contents with SP indicator
5. **Disassembly View**: Show instructions around PC
6. **Command Input**: Command line for debugger commands
7. **Output/Console**: Display program output and debugger messages
8. **Watchpoints Panel**: List active memory/register watches

**Features:**
- Responsive layout with resize handling
- Syntax highlighting for assembly code
- Real-time updates on execution
- Minimal flicker
- Color support with graceful fallback

**Navigation:**
- Arrow keys: Navigate views
- Tab/Shift-Tab: Switch between panels
- Page Up/Down: Scroll in active view
- Home/End: Jump to top/bottom

**Platform Support:**
- macOS: Terminal.app, iTerm2
- Windows: cmd.exe, PowerShell, Windows Terminal
- Linux: gnome-terminal, konsole, xterm, etc.

**Implementation with tview:**
```go
// Panels using tview primitives:
// - Flex for layout
// - TextView for source/registers/console
// - Table for memory view
// - InputField for command input
```

---

## Phase 7: Testing (Weeks 13-15)

### 13. Unit Tests (Target: 1000+ tests)

**Instruction Tests (600+ tests):**

Location: `tests/unit/instructions/`

**Categories:**
- Data processing: `test_mov.go`, `test_add.go`, `test_sub.go`, etc.
- Memory: `test_ldr.go`, `test_str.go`, `test_ldrb.go`, etc.
- Branch: `test_b.go`, `test_bl.go`
- Multiply: `test_mul.go`, `test_mla.go`

**Test Requirements per Instruction:**
1. Basic operation verification
2. All addressing modes
3. All condition codes
4. Flag updates (N, Z, C, V)
5. Edge cases:
   - Zero operands
   - Maximum/minimum values
   - Overflow/underflow
   - Negative numbers
6. Register aliases (SP, LR, PC)
7. Boundary conditions

**Example Test Structure:**
```go
func TestADD_SignedOverflow(t *testing.T) {
    vm := NewVM()
    vm.R[0] = 0x7FFFFFFF  // INT32_MAX
    vm.R[1] = 1
    vm.Execute("ADDS R2, R0, R1")

    assert.Equal(t, uint32(0x80000000), vm.R[2])
    assert.Equal(t, true, vm.CPSR.V)   // Overflow
    assert.Equal(t, false, vm.CPSR.C)  // No carry
    assert.Equal(t, true, vm.CPSR.N)   // Negative
}
```

**Flag Calculation Tests (100+ tests):**

Location: `tests/unit/flags/`

**Test Cases:**
- N flag: bit 31 set
- Z flag: result equals zero
- C flag:
  - Addition: unsigned overflow
  - Subtraction: unsigned borrow
  - Shifts: last bit shifted out
- V flag: signed overflow (all cases)

**Memory System Tests (50+ tests):**

Location: `tests/unit/memory/`

**Categories:**
- Alignment tests (aligned/unaligned word/halfword/byte access)
- Permission tests (write to code segment, execute from data)
- Boundary tests (null pointer, out of bounds, stack overflow/underflow)
- Endianness tests (little-endian load/store)

**Parser Tests (90+ tests):**

Location: `tests/unit/parser/`

**Categories:**
- Instruction parsing (basic, with conditions, with S flag)
- Label parsing (global, local, numeric, forward references)
- Directive parsing (all directives)
- Comment handling (line, block, inline)
- Error handling (invalid instruction, undefined label, syntax errors)
- Edge cases (empty file, only comments, long lines, nested/circular includes)

**Addressing Mode Tests (60+ tests):**

Location: `tests/unit/addressing/`

**Test all 9 addressing modes:**
1. Immediate
2. Register
3. Register with shift (LSL, LSR, ASR, ROR)
4. Register shift by register
5. Memory offset
6. Memory pre-indexed
7. Memory post-indexed
8. Memory register offset
9. Memory register offset with shift

**System Call Tests (30+ tests):**

Location: `tests/unit/syscall/`

**Test each syscall:**
- Console I/O (write/read char/string/int)
- File operations (open/close/read/write/seek)
- Memory operations (allocate/free/reallocate)
- System info (time, random, arguments)
- Error handling (invalid syscall number, error codes)

**Coverage Requirements:**
- Instruction execution: 95%
- Memory system: 90%
- Parser: 85%
- VM core: 90%
- Overall: 85%

### 14. Integration Tests

Location: `tests/integration/`

**Complete Program Tests (20+ tests):**
- `factorial.s` - Recursive function calls
- `fibonacci.s` - Iterative loops
- `bubble_sort.s` - Array manipulation
- `string_ops.s` - String operations
- `linked_list.s` - Dynamic memory allocation
- `nested_calls.s` - Deep call stack

**Each test verifies:**
- Program completes successfully
- Correct output produced
- Final register/memory state as expected
- No memory leaks
- Stack properly maintained

**Cross-Component Tests (15+ tests):**
- Parser → VM: Parsed instructions execute correctly
- VM → Memory: Instructions access memory correctly
- VM → Syscalls: System calls work during execution
- Debugger → VM: Breakpoints interrupt correctly

**Regression Tests (30+ tests):**

Maintain tests for all previously found bugs:
```go
// Bug #42: LDR post-indexed addressing incorrect
func TestRegression_Bug42_LDRPostIndexed(t *testing.T) {
    // Test case to prevent regression
}
```

### 15. Debugger Tests (40+ tests)

Location: `tests/debugger/`

**Categories:**
- Breakpoint tests (set at address/label, conditional, temporary, multiple)
- Execution control (step, next, continue, finish)
- State inspection (registers, memory, expressions, call stack)
- Watchpoint tests (register/memory read/write/access)

**Test Coverage Goal: 85%+ for all components**

---

## Phase 8: Development Tools (Week 16)

### 16. Tools

**tools/lint.go - Assembly Linter**

**Features:**
- Syntax validation
- Undefined label detection
- Unreachable code detection
- Register usage warnings
- Best practice recommendations

**Usage:**
```bash
arm-lint program.s
arm-lint --strict program.s
arm-lint --fix program.s
```

**Example Output:**
```
program.s:15: warning: Label 'unused_func' defined but never referenced
program.s:23: error: Undefined label 'typo_label'
program.s:31: warning: Register R7 clobbered without saving
```

**tools/format.go - Code Formatter**

**Features:**
- Consistent indentation
- Align operands in columns
- Normalize spacing
- Comment alignment
- Label formatting

**Usage:**
```bash
arm-format program.s          # Print formatted version
arm-format -w program.s       # Write changes to file
arm-format --style=compact program.s
```

**tools/xref.go - Cross-Reference Generator**

**Features:**
- Symbol cross-reference
- Show all symbol definitions and uses
- Function reference tracking

**Output:**
```
Symbol Cross-Reference for program.s
=====================================

_start                  Defined: line 10
                        Referenced: (entry point)

add_numbers            Defined: line 25
                        Referenced: line 15, line 42, line 58
```

**tools/disassembler.go - Disassembler**

Future feature: Convert binary machine code back to assembly

---

## Phase 9: Examples & Documentation (Week 17)

### 17. Example Programs

Location: `examples/`

**Basic Examples:**
- `hello.s` - Hello World
- `arithmetic.s` - Basic arithmetic operations

**Algorithm Examples:**
- `fibonacci.s` - Fibonacci sequence (iterative)
- `factorial.s` - Factorial (iterative and recursive)
- `bubble_sort.s` - Bubble sort algorithm
- `binary_search.s` - Binary search
- `gcd.s` - Greatest common divisor (Euclidean algorithm)

**Data Structure Examples:**
- `arrays.s` - Array operations (init, access, traversal, min/max)
- `linked_list.s` - Linked list (insert, delete, traversal)
- `stack.s` - Stack implementation and calculator
- `strings.s` - String manipulation (length, copy, compare, concat)

**Advanced Examples:**
- `functions.s` - Function calling conventions
- `conditionals.s` - If/else, switch/case
- `loops.s` - For, while, do-while loops

**Test Programs:**

Location: `tests/programs/`

- Instruction tests (one per instruction)
- Flag tests (verify CPSR flag setting)
- Edge case tests (overflow, underflow, boundary conditions)

### 18. Documentation

**User Documentation:**

- `README.md` - Overview, quick start
- `docs/installation.md` - Platform-specific installation
- `docs/assembly_reference.md` - ARM2 assembly language reference
- `docs/debugger_reference.md` - Debugger commands
- `docs/tutorial.md` - Step-by-step tutorial
- `docs/faq.md` - Frequently asked questions
- `docs/syscalls.md` - System call reference

**Developer Documentation:**

- `docs/api_reference.md` - API documentation
- `docs/architecture.md` - Architecture overview
- `docs/contributing.md` - Contributing guidelines
- `docs/coding_standards.md` - Go coding standards

**Platform-Specific Notes:**

Include platform differences in documentation:
- Config file locations (macOS/Windows/Linux)
- Log file locations
- Terminal compatibility notes
- Known limitations per platform

---

## Phase 10: Cross-Platform & Polish (Week 18)

### 19. Cross-Platform Features

**File System Handling:**
```go
// ✓ CORRECT: Use filepath.Join for cross-platform paths
configPath := filepath.Join(homeDir, ".config", "arm-emu", "config.toml")

// ✗ WRONG: Hard-coded separators
configPath := homeDir + "/.config/arm-emu/config.toml"
```

**Platform-Specific Config Paths:**
- **macOS/Linux**: `~/.config/arm-emu/config.toml`
- **Windows**: `%APPDATA%\arm-emu\config.toml`

**Terminal Handling:**
- Support different terminal emulators on all platforms
- Handle terminal resize events
- Detect ANSI color support
- Graceful fallback to no-color mode

**Cross-Compilation:**

Makefile targets:
```makefile
build-all:
	GOOS=darwin GOARCH=amd64 go build -o bin/arm-emu-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o bin/arm-emu-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o bin/arm-emu-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o bin/arm-emu-windows-amd64.exe .
```

**CI/CD Testing:**

GitHub Actions workflow:
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go-version: [1.21]
```

**Manual Testing Checklist:**
- [ ] Install and run on macOS (Apple Silicon)
- [ ] Install and run on Windows 10/11
- [ ] Install and run on Linux (Ubuntu, Fedora)
- [ ] TUI renders correctly on all platforms
- [ ] File I/O works correctly
- [ ] Path handling works
- [ ] Config file loading works
- [ ] Example programs run identically

### 20. Performance & Diagnostics

**Performance Targets:**
- Parser: < 100ms for programs < 1000 lines
- Execution: > 100k instructions/second
- Memory: < 100MB for typical programs
- TUI refresh: 60 FPS minimum

**Execution Trace:**

Format:
```
[000001] 0x8000: MOV R0, #10      | R0=0x0000000A CPSR=---- | 0.001ms
[000002] 0x8004: ADD R1, R0, R1   | R1=0x00000014 CPSR=---- | 0.002ms
```

Usage:
```bash
arm-emu --trace trace.log program.s
arm-emu --trace-filter="R0,R1" program.s
```

**Memory Access Log:**

Format:
```
[READ ] 0x8024: LDR R0, [R1] <- [0x20000] = 0x12345678
[WRITE] 0x8028: STR R0, [R2] -> [0x20010] = 0x12345678
```

**Performance Statistics:**

Metrics:
- Total instructions executed
- Instructions per second
- Instruction type breakdown
- Branch prediction stats
- Function call count
- Memory access count

**Export Formats:**
- JSON (machine-readable)
- CSV (spreadsheet import)
- HTML (interactive visualization)

---

## Delivery Milestones

### M1: Core VM (Week 2)
**Deliverables:**
- Basic VM with registers, memory, flags
- Executes MOV, ADD, SUB, B instructions
- Simple test suite passing

**Success Criteria:**
- Can execute simple arithmetic program
- Registers update correctly
- Flags set correctly
- 50+ unit tests passing

### M2: Parser Complete (Week 4)
**Deliverables:**
- Full lexer and parser
- All directives supported
- Symbol table with forward references
- Error reporting with line/column

**Success Criteria:**
- Parses all valid ARM2 syntax
- Handles all directives
- Useful error messages
- 90+ parser tests passing

### M3: Complete Instruction Set (Week 7)
**Deliverables:**
- All ARM2 instructions implemented
- All addressing modes working
- All condition codes working
- Multiply instructions

**Success Criteria:**
- All 30+ instructions working
- 600+ instruction tests passing
- Flag calculation 100% correct
- Can execute complex programs

### M4: System Calls (Week 8)
**Deliverables:**
- SWI instruction handler
- All console I/O syscalls
- File operation syscalls
- Memory allocation syscalls
- Standard library macros

**Success Criteria:**
- Programs can print output
- Programs can read input
- File I/O works
- Dynamic memory allocation works
- Example programs run successfully

### M5: Debugger Core (Week 10)
**Deliverables:**
- Command processor
- Breakpoints (address, label, conditional)
- Execution control (run, step, continue)
- State inspection (registers, memory)
- Watchpoints

**Success Criteria:**
- Can set/delete breakpoints
- Step through code
- Inspect registers and memory
- Conditional breakpoints work
- 40+ debugger tests passing

### M6: Full TUI (Week 12)
**Deliverables:**
- Complete TUI with all panels
- Source view with syntax highlighting
- Register view with live updates
- Memory view (hex + ASCII)
- Stack view
- Command input with history

**Success Criteria:**
- TUI renders correctly
- All panels functional
- Responsive to window resize
- Works on all platforms
- Smooth 60 FPS updates

### M7: Testing Complete (Week 15)
**Deliverables:**
- 1000+ unit tests
- 20+ integration tests
- 30+ regression tests
- 85%+ code coverage
- CI/CD running on all platforms

**Success Criteria:**
- All tests passing
- Coverage thresholds met
- No memory leaks
- Example programs all work
- Cross-platform compatibility verified

### M8: Release Ready (Week 18)
**Deliverables:**
- Complete documentation
- Development tools (linter, formatter, xref)
- Example program suite
- Cross-platform binaries
- Installation packages

**Success Criteria:**
- All documentation complete
- Installable on all platforms
- Performance targets met
- Zero known critical bugs
- Ready for public release

---

## Key Technical Decisions

### Language & Runtime
- **Language**: Go 1.25+ for cross-platform support and strong standard library
- **Target Platforms**: macOS (Apple Silicon), Windows 10/11, Linux (Ubuntu, Fedora, Arch)
- **Architecture Support**: x86_64 (AMD64), ARM64

### Libraries & Frameworks
- **TUI**: `tview` + `tcell` (mature, cross-platform, well-documented)
- **CLI**: `cobra` (standard for Go CLIs)
- **Config**: TOML (human-readable, `BurntSushi/toml`)
- **Testing**: Go's built-in testing framework

### Performance Targets
- **Execution Speed**: > 100,000 instructions/second (interpreted)
- **Memory Usage**: < 100MB for typical programs
- **Parser Speed**: < 100ms for < 1000 line programs
- **TUI Refresh**: 60 FPS minimum

### Code Quality
- **Test Coverage**: 85%+ overall, 95%+ for instruction execution
- **Test Count**: 1000+ unit tests minimum
- **Documentation**: Complete user and developer docs
- **Cross-Platform**: CI/CD testing on all platforms

---

## Risk Mitigation

### TUI Complexity
**Risk**: Terminal UI is complex and platform-dependent

**Mitigation**:
- Use mature, well-tested library (tview)
- Start with simple layout, iterate
- Test early and often on all platforms
- Have fallback to simple CLI mode

### Parser Edge Cases
**Risk**: Assembly syntax has many edge cases

**Mitigation**:
- Build comprehensive test suite from day 1
- Use two-pass assembly for forward references
- Thorough error handling and reporting
- Reference existing assemblers (GNU as)

### Cross-Platform Issues
**Risk**: Behavior differs across platforms

**Mitigation**:
- Use Go's standard library (avoid platform-specific code)
- Weekly testing on all platforms via CI
- Document platform-specific differences
- Use `filepath` package for all path operations

### Performance Bottlenecks
**Risk**: Interpreted execution may be too slow

**Mitigation**:
- Profile early and often
- Optimize hot paths (instruction dispatch, memory access)
- Use efficient data structures
- Consider JIT compilation in future phase

### Scope Creep
**Risk**: Feature requests expand scope indefinitely

**Mitigation**:
- Stick to ARM2 specification
- Defer ARM3/ARM6/Thumb to Phase 3
- Focus on core functionality first
- Maintain clear milestone goals

---

## Dependencies

### Core Dependencies
- `github.com/rivo/tview` - Rich TUI components
- `github.com/gdamore/tcell/v2` - Terminal handling
- `github.com/spf13/cobra` - CLI framework
- `github.com/BurntSushi/toml` - TOML parser

### Development Dependencies
- `github.com/stretchr/testify` - Testing assertions
- Go's built-in testing framework
- Go's built-in benchmarking
- `golangci-lint` - Linter
- `gofmt` - Formatter

**Note**: All dependencies must support macOS, Windows, and Linux

---

## Timeline Summary

| Phase | Weeks | Focus Area | Deliverable |
|-------|-------|------------|-------------|
| 1 | 1-2 | Foundation | Core VM + Project Setup |
| 2 | 3-4 | Parser | Complete Parser + Symbol Table |
| 3 | 5-7 | Instructions | All ARM2 Instructions |
| 4 | 8 | System | System Calls + Runtime |
| 5 | 9-10 | Debugger | Debugger Core + Commands |
| 6 | 11-12 | TUI | Full Text UI |
| 7 | 13-15 | Testing | 1000+ Tests + 85% Coverage |
| 8 | 16 | Tools | Linter + Formatter + Xref |
| 9 | 17 | Examples | Example Programs + Docs |
| 10 | 18 | Polish | Cross-Platform + Release |

**Total Duration: 18 weeks**

---

## Success Criteria

The ARM2 emulator will be considered complete when:

1. **Functionality**:
   - All ARM2 instructions implemented and tested
   - Full debugger with TUI working
   - System calls functional
   - Example programs run successfully

2. **Quality**:
   - 1000+ tests passing
   - 85%+ code coverage
   - Zero known critical bugs
   - Performance targets met

3. **Cross-Platform**:
   - Works on macOS (Apple Silicon)
   - Works on Windows 10/11
   - Works on Linux (Ubuntu, Fedora)
   - Identical behavior across platforms

4. **Documentation**:
   - Complete user documentation
   - Complete developer documentation
   - API reference
   - Tutorial with examples

5. **Usability**:
   - Installable via standard methods
   - Clear error messages
   - Intuitive debugger commands
   - Responsive TUI

---

## Next Steps

1. Review and approve this implementation plan
2. Set up initial project structure (Week 1)
3. Begin Phase 1: Core VM implementation
4. Establish weekly review cadence
5. Track progress against milestones

---

*Document Version: 1.0*
*Last Updated: 2025-10-08*
