# ARM2 Assembly Language Emulator Specification

## Overview

This document specifies the design and implementation of an ARM2 assembly language emulator with an integrated debugger. The emulator interprets ARM2 assembly code directly without producing machine code, executing instructions in a virtual machine environment.

## 1. Core Components

### 1.1 Virtual Machine (VM)

The VM provides the runtime environment for executing ARM2 assembly instructions.

#### 1.1.1 Register Set
- **General Purpose Registers**: R0-R14 (32-bit each)
- **Program Counter (PC)**: R15
- **Current Program Status Register (CPSR)**: Contains condition flags
  - N (Negative)
  - Z (Zero)
  - C (Carry)
  - V (Overflow)
- **Stack Pointer (SP)**: R13 by convention
- **Link Register (LR)**: R14 by convention

#### 1.1.2 Memory System
- **Address Space**: 32-bit addressable (4GB virtual space)
- **Memory Layout**:
  - Code segment (read-only after load)
  - Data segment (read-write)
  - Stack segment (grows downward)
  - Heap segment (grows upward)
- **Endianness**: Little-endian by default
- **Access Granularity**: Byte, halfword (16-bit), word (32-bit)

#### 1.1.3 Execution Engine
- **Fetch-Decode-Execute Cycle**: Interpret instructions sequentially
- **Instruction Pipeline**: Single-stage (no pipelining initially)
- **Execution Modes**:
  - Run: Execute until halt or breakpoint
  - Step: Execute single instruction
  - Step Over: Execute until next instruction at same call level
  - Step Into: Execute single instruction, following branches/calls

### 1.2 Assembly Language Parser

#### 1.2.1 Syntax Support
- **Instruction Format**: `LABEL: MNEMONIC{COND}{S} operands ; comment`
- **Labels**: Alphanumeric identifiers followed by colon
  - Global labels: Start at column 0 or with alphanumeric characters
  - Local labels: Start with `.` (dot) for local scope within functions
  - Numeric labels: Support for backward/forward numeric labels (e.g., `1:`, `1b`, `1f`)
- **Comments**:
  - Line comments: `; comment` or `// comment`
  - Block comments: `/* comment */`
  - Inline comments: Comments after instructions
- **Case Sensitivity**: Instructions are case-insensitive, labels are case-sensitive
- **Directives**:
  - `.org <address>`: Set origin address
  - `.equ <name>, <value>`: Define constant
  - `.set <name>, <value>`: Alias for .equ
  - `.word <value>`: Allocate word (32-bit)
  - `.half <value>`: Allocate halfword (16-bit)
  - `.byte <value>`: Allocate byte (8-bit)
  - `.ascii "<string>"`: Allocate ASCII string
  - `.asciz "<string>"`: Allocate null-terminated ASCII string
  - `.string "<string>"`: Alias for .asciz
  - `.space <size>`: Reserve space
  - `.skip <size>`: Alias for .space
  - `.align <boundary>`: Align to boundary (power of 2)
  - `.balign <boundary>`: Byte alignment
  - `.include "<file>"`: Include another assembly file
  - `.macro <name>` / `.endm`: Define macro
  - `.if` / `.ifdef` / `.ifndef` / `.endif`: Conditional assembly
  - `.global <symbol>`: Mark symbol as global (for linking)
  - `.extern <symbol>`: Declare external symbol
  - `.section <name>`: Define section (.text, .data, .bss)
  - `.text`, `.data`, `.bss`: Shorthand section directives

#### 1.2.2 Supported Instructions

**Data Processing**:
- `MOV, MVN`: Move/Move Not
- `ADD, ADC, SUB, SBC, RSB, RSC`: Arithmetic
- `AND, ORR, EOR, BIC`: Logical
- `CMP, CMN, TST, TEQ`: Comparison (set flags only)

**Multiply**:
- `MUL`: Multiply
- `MLA`: Multiply-Accumulate

**Memory Access**:
- `LDR, STR`: Load/Store word
- `LDRB, STRB`: Load/Store byte
- `LDRH, STRH`: Load/Store halfword (if supporting ARM2a extensions)
- **Load/Store Multiple** (missing from original):
  - `LDM{mode}`: Load multiple registers
  - `STM{mode}`: Store multiple registers
  - Modes: IA (Increment After), IB (Increment Before), DA (Decrement After), DB (Decrement Before)
  - Variants: FD, ED, FA, EA (Full/Empty, Descending/Ascending)

**Branch**:
- `B{cond}`: Branch
- `BL{cond}`: Branch with Link
- **Branch and Exchange** (ARM architecture completeness):
  - `BX{cond}`: Branch and Exchange (for ARM/Thumb interworking if future extension)

**System**:
- `SWI <number>`: Software Interrupt (for system calls)

**Condition Codes**:
- `EQ, NE, CS/HS, CC/LO, MI, PL, VS, VC, HI, LS, GE, LT, GT, LE, AL`
- **Note**: `NV` (Never) condition exists in ARM2 but is deprecated and should generate a warning

**Addressing Modes**:
- Immediate: `#value`
- Register: `Rn`
- Register with shift: `Rm, LSL #shift`, `Rm, LSR #shift`, `Rm, ASR #shift`, `Rm, ROR #shift`
- Register shift by register: `Rm, LSL Rs`, `Rm, LSR Rs`, `Rm, ASR Rs`, `Rm, ROR Rs`
- Rotate right extended: `Rm, RRX` (shift right with carry)
- Memory addressing:
  - Offset: `[Rn, #offset]`
  - Pre-indexed: `[Rn, #offset]!`
  - Post-indexed: `[Rn], #offset`
  - Register offset: `[Rn, Rm]`
  - Register offset with shift: `[Rn, Rm, LSL #shift]`
  - Scaled register offset: `[Rn, Rm, shift_type #amount]`

#### 1.2.3 Parser Output
- **Intermediate Representation (IR)**:
  - Instruction type
  - Operands (registers, immediates, addresses)
  - Condition code
  - Flags update (S bit)
  - Addressing mode details
- **Symbol Table**: Maps labels to addresses, with forward reference resolution
- **Relocation Table**: Track symbols requiring address resolution
- **Error Handling**: Line numbers, column positions, error messages, warnings, suggestions
- **Metadata**: Source file mapping for debugging, macro expansion tracking

### 1.3 Debugger

The debugger provides interactive control and inspection of the VM state.

#### 1.3.1 TUI (Text User Interface)

**Layout Components**:
1. **Source View**: Display assembly source with current instruction highlighted
2. **Register View**: Show all registers (R0-R15, CPSR) with real-time updates
3. **Memory View**: Hexadecimal and ASCII display of memory regions
4. **Stack View**: Display stack contents with SP indicator
5. **Disassembly View**: Show instructions around PC
6. **Command Input**: Command line for debugger commands
7. **Output/Console**: Display program output and debugger messages
8. **Watchpoints Panel**: List active memory/register watches

**Navigation**:
- Arrow keys: Navigate views
- Tab/Shift-Tab: Switch between panels
- Page Up/Down: Scroll in active view
- Home/End: Jump to top/bottom

#### 1.3.2 Debugger Commands

**Execution Control**:
- `run` / `r`: Run program until completion or breakpoint
- `step` / `s`: Execute single instruction (step into)
- `next` / `n`: Execute single instruction (step over calls)
- `stepi` / `si`: Step one instruction
- `continue` / `c`: Continue execution
- `finish`: Execute until current function returns
- `until <location>`: Run until location reached

**Breakpoints**:
- `break <location>` / `b <location>`: Set breakpoint (address, label, or line)
- `tbreak <location>`: Set temporary breakpoint
- `delete <id>`: Delete breakpoint
- `disable <id>`: Disable breakpoint
- `enable <id>`: Enable breakpoint
- `info breakpoints` / `i b`: List all breakpoints

**Watchpoints**:
- `watch <expression>`: Set write watchpoint
- `rwatch <expression>`: Set read watchpoint
- `awatch <expression>`: Set access watchpoint (read/write)
- `watch register <Rn>`: Watch register changes
- `watch memory <address>`: Watch memory location
- `info watchpoints`: List all watchpoints

**Inspection**:
- `print <expression>` / `p <expression>`: Evaluate and print expression
- `x/<format> <address>`: Examine memory
  - Format: `[count][format][size]`
  - Formats: `x` (hex), `d` (decimal), `u` (unsigned), `o` (octal), `t` (binary), `a` (address), `c` (char), `s` (string)
  - Sizes: `b` (byte), `h` (halfword), `w` (word)
- `info registers` / `i r`: Display all registers
- `info stack` / `i s`: Display stack information
- `backtrace` / `bt`: Show call stack
- `list` / `l`: List source code around current location

**State Modification**:
- `set $Rn = <value>`: Set register value
- `set {<type>}<address> = <value>`: Set memory value
- `set var <name> = <value>`: Set variable value

**Program Control**:
- `load <file>`: Load assembly file
- `reset`: Reset VM to initial state
- `quit` / `q`: Exit debugger

#### 1.3.3 Debugger Features

**Call Stack Tracking**:
- Automatic detection of `BL` (function calls)
- Track return addresses (stored in LR)
- Display call hierarchy with frame information
- Frame selection for inspection

**Symbol Resolution**:
- Display labels instead of addresses when available
- Show function names in call stack
- Resolve addresses to nearest symbol

**Memory Inspection**:
- Multiple memory windows with different views
- Follow pointers (dereference)
- Search memory for patterns
- Memory change highlighting

**Watch Expressions**:
- Monitor register values
- Monitor memory locations
- Evaluate complex expressions
- Trigger on value changes

**Breakpoint Types**:
- Address breakpoints
- Conditional breakpoints (e.g., break when R0 == 5)
- Hardware breakpoints (limited number)
- One-shot breakpoints

**History**:
- Command history (up/down arrows)
- Execution history (reverse step - optional advanced feature)
- Register value history

## 2. File Formats

### 2.1 Assembly Source File (.s or .asm)

**Format**:
```
; Example ARM2 assembly program
        .org 0x8000         ; Start at address 0x8000

start:  MOV R0, #10         ; Load immediate value
        MOV R1, #20
        BL  add_nums        ; Call function
        B   end

add_nums:
        ADD R0, R0, R1      ; Add R0 and R1
        MOV PC, LR          ; Return

end:    B end               ; Infinite loop (halt)
```

### 2.2 Configuration File (.toml or .json)

**Purpose**: Store debugger settings, memory layout, etc.

**Example**:
```toml
[memory]
code_start = 0x8000
code_size = 0x10000
data_start = 0x20000
data_size = 0x10000
stack_start = 0x40000
stack_size = 0x10000

[debugger]
default_breakpoints = ["main", "error_handler"]
show_machine_code = false
syntax_highlighting = true

[execution]
max_cycles = 1000000
halt_on_undefined = true
```

### 2.3 Saved State File (.state)

**Purpose**: Save/restore VM state for debugging sessions

**Contents**:
- All register values
- Memory snapshot
- Breakpoint list
- Watchpoint list
- Current PC

## 3. Implementation Architecture

### 3.1 Module Structure

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

### 3.2 Data Flow

1. **Load Phase**:
   - Parse assembly source file
   - Build symbol table
   - Generate IR
   - Load into VM memory

2. **Execution Phase**:
   - Fetch instruction at PC
   - Decode instruction (from IR)
   - Check breakpoints/watchpoints
   - Execute instruction
   - Update registers/memory
   - Update PC
   - Check watchpoint triggers

3. **Debug Phase**:
   - Display current state in TUI
   - Wait for user command
   - Process command
   - Update displays

### 3.3 TUI Framework

**Recommended Libraries** (for Go):
- `tview`: Rich TUI components
- `tcell`: Terminal handling
- `lipgloss`: Styling and layout (alternative)

**Key Requirements**:
- Responsive layout (resize handling)
- Syntax highlighting
- Real-time updates
- Minimal flicker

## 4. System Calls and I/O Interface

### 4.1 Software Interrupt (SWI) Mechanism

The emulator provides system call functionality through the `SWI` instruction, allowing programs to interact with the host environment.

#### 4.1.1 SWI Instruction Format
```
SWI #number
```
When executed:
1. Program execution pauses
2. Emulator handles the system call based on the number
3. Results are returned in registers (typically R0)
4. Execution continues at the instruction following SWI

#### 4.1.2 Standard System Call Numbers

**Console I/O**:
- `0x00 - Exit`: Terminate program (R0 = exit code)
- `0x01 - Write Char`: Output character (R0 = char value)
- `0x02 - Write String`: Output null-terminated string (R0 = string address)
- `0x03 - Write Int`: Output integer (R0 = value, R1 = base 2/8/10/16)
- `0x04 - Read Char`: Input character (returns in R0)
- `0x05 - Read String`: Input string (R0 = buffer address, R1 = max length)
- `0x06 - Read Int`: Input integer (returns in R0)
- `0x07 - Write Newline`: Output newline character

**File Operations**:
- `0x10 - Open`: Open file (R0 = filename ptr, R1 = mode) → returns file descriptor in R0
- `0x11 - Close`: Close file (R0 = file descriptor)
- `0x12 - Read`: Read from file (R0 = fd, R1 = buffer, R2 = count) → returns bytes read
- `0x13 - Write`: Write to file (R0 = fd, R1 = buffer, R2 = count) → returns bytes written
- `0x14 - Seek`: Seek in file (R0 = fd, R1 = offset, R2 = whence)
- `0x15 - Tell`: Get current position (R0 = fd) → returns offset in R0
- `0x16 - FileSize`: Get file size (R0 = fd) → returns size in R0

**Memory Operations**:
- `0x20 - Allocate`: Allocate heap memory (R0 = size) → returns address in R0
- `0x21 - Free`: Free heap memory (R0 = address)
- `0x22 - Reallocate`: Resize allocation (R0 = address, R1 = new size) → returns new address in R0

**System Information**:
- `0x30 - Get Time`: Get current time in milliseconds → returns in R0
- `0x31 - Get Random`: Get random number → returns in R0
- `0x32 - Get Arguments`: Get command-line arguments (R0 = argc, R1 = argv pointer)
- `0x33 - Get Environment`: Get environment variable (R0 = name ptr) → returns value ptr in R0

**Error Handling**:
- `0x40 - Get Error`: Get last error code → returns in R0
- `0x41 - Set Error`: Set error code (R0 = error code)
- `0x42 - Print Error`: Print last error message

**Debugging Support**:
- `0xF0 - Debug Print`: Print debug message (R0 = string address)
- `0xF1 - Breakpoint`: Trigger debugger breakpoint
- `0xF2 - Dump Registers`: Print all register values
- `0xF3 - Dump Memory`: Print memory region (R0 = address, R1 = length)
- `0xF4 - Assert`: Assert condition (R0 = condition, R1 = message ptr)

#### 4.1.3 Example Usage
```asm
; Print "Hello, World!" example
        .org 0x8000

start:  LDR R0, =hello_msg
        SWI #0x02           ; Write string syscall
        MOV R0, #0
        SWI #0x00           ; Exit with code 0

hello_msg:
        .asciz "Hello, World!\n"
```

#### 4.1.4 System Call Error Handling

System calls return error codes to indicate success or failure:

**Return Values**:
- **R0**: Return value or error code
  - Positive values or zero: Success (may contain result)
  - Negative values: Error code (errno-style)

**Common Error Codes**:
```
0  - Success
-1 - EPERM (Operation not permitted)
-2 - ENOENT (No such file or directory)
-3 - EIO (I/O error)
-4 - ENOMEM (Out of memory)
-5 - EACCES (Permission denied)
-6 - EFAULT (Bad address)
-9 - EBADF (Bad file descriptor)
-11 - EAGAIN (Try again)
-12 - ENOMEM (Out of memory)
-14 - EFAULT (Bad address)
-22 - EINVAL (Invalid argument)
-28 - ENOSPC (No space left on device)
```

**Error Checking Example**:
```asm
        ; Try to open a file
        LDR     R0, =filename
        MOV     R1, #0              ; Read-only mode
        SWI     #0x10               ; Open syscall
        CMP     R0, #0
        BMI     handle_error        ; Branch if negative (error)
        ; R0 now contains file descriptor
        MOV     R4, R0              ; Save fd
        ; ... use file ...
        B       continue

handle_error:
        ; R0 contains error code
        SWI     #0x42               ; Print error message
        MOV     R0, #1
        SWI     #0x00               ; Exit with error code
```

### 4.2 Startup and Runtime Environment

#### 4.2.1 Bootstrap Sequence

When a program is loaded, the emulator performs the following initialization:

1. **Memory Layout Setup**:
   - Load code into code segment
   - Initialize data segment with static data
   - Set up stack at high memory address
   - Initialize heap space

2. **Register Initialization**:
   - PC = Entry point (default: first instruction or `_start` label)
   - SP (R13) = Stack top address
   - LR (R14) = 0 (no return address)
   - R0-R12 = 0 (or argument values if supported)
   - CPSR = 0 (all flags clear)

3. **Argument Passing** (optional feature):
   - R0 = argc (argument count)
   - R1 = argv pointer (array of string pointers in memory)
   - Format similar to C `main(int argc, char **argv)`

#### 4.2.2 Program Entry Points

The emulator searches for entry points in this order:
1. `_start` label (if present)
2. `main` label (if present)
3. First instruction at origin address

#### 4.2.3 Program Termination

Programs can terminate through:
- `SWI #0x00` with exit code in R0
- Reaching `_exit` label
- Execution of `B _exit` or infinite loop (debugger detects this)
- Maximum instruction count exceeded
- Fatal error (invalid instruction, memory violation)

### 4.3 Standard Library Macros

To simplify common operations, provide a standard library of assembly macros:

```asm
; stdlib.inc - Standard library macros

.macro PRINT_STR reg
        MOV R0, \reg
        SWI #0x02
.endm

.macro PRINT_CHAR char
        MOV R0, #\char
        SWI #0x01
.endm

.macro PRINT_INT reg
        MOV R0, \reg
        MOV R1, #10         ; Base 10
        SWI #0x03
.endm

.macro EXIT code
        MOV R0, #\code
        SWI #0x00
.endm

.macro ALLOC size, dest
        MOV R0, #\size
        SWI #0x20
        MOV \dest, R0
.endm
```

## 5. Example Programs Suite

The emulator should include a comprehensive set of example programs demonstrating various features and serving as both tests and learning resources.

### 5.1 Basic Examples

#### 5.1.1 Hello World (`examples/hello.s`)
```asm
; Simplest program - print a message
        .org 0x8000
start:  LDR R0, =message
        SWI #0x02
        MOV R0, #0
        SWI #0x00
message:
        .asciz "Hello, World!\n"
```

#### 5.1.2 Simple Arithmetic (`examples/arithmetic.s`)
```asm
; Demonstrate basic arithmetic operations
        .org 0x8000
start:  MOV R0, #10
        MOV R1, #20
        ADD R2, R0, R1      ; R2 = 30
        SUB R3, R1, R0      ; R3 = 10
        MUL R4, R0, R1      ; R4 = 200
        ; Print results...
```

### 5.2 Algorithm Examples

#### 5.2.1 Fibonacci Sequence (`examples/fibonacci.s`)
Iterative implementation of Fibonacci number generation.

#### 5.2.2 Factorial (`examples/factorial.s`)
Both iterative and recursive implementations.

#### 5.2.3 Bubble Sort (`examples/bubble_sort.s`)
Sort an array using bubble sort algorithm, demonstrating memory access and loops.

#### 5.2.4 Binary Search (`examples/binary_search.s`)
Search sorted array using binary search algorithm.

#### 5.2.5 GCD (`examples/gcd.s`)
Calculate greatest common divisor using Euclidean algorithm.

### 5.3 Data Structure Examples

#### 5.3.1 Array Operations (`examples/arrays.s`)
- Array initialization
- Element access
- Array traversal
- Finding min/max

#### 5.3.2 Linked List (`examples/linked_list.s`)
- Node structure
- Insert/delete operations
- List traversal
- Dynamic memory allocation

#### 5.3.3 Stack Implementation (`examples/stack.s`)
- Push/pop operations
- Stack-based calculator

#### 5.3.4 String Manipulation (`examples/strings.s`)
- String length
- String copy
- String comparison
- String concatenation

### 5.4 Advanced Examples

#### 5.4.1 Function Calls (`examples/functions.s`)
Demonstrate proper function calling convention:
- Parameter passing
- Local variables
- Return values
- Nested calls
- Recursion

#### 5.4.2 Conditional Logic (`examples/conditionals.s`)
- If/else statements
- Switch/case structures
- Conditional execution

#### 5.4.3 Loop Constructs (`examples/loops.s`)
- For loops
- While loops
- Do-while loops
- Loop unrolling

### 5.5 Test Programs

#### 5.5.1 Instruction Tests (`tests/instructions/`)
Individual test for each instruction verifying correct behavior.

#### 5.5.2 Flag Tests (`tests/flags/`)
Verify correct CPSR flag setting for all operations.

#### 5.5.3 Edge Cases (`tests/edge_cases/`)
- Integer overflow/underflow
- Division edge cases
- Memory boundary conditions
- Maximum recursion depth

## 6. User Workflow

### 6.1 Typical Session

1. **Start**: `arm-emu program.s`
2. **Load**: Program automatically loaded and parsed
3. **Set Breakpoints**: `break main`, `break loop_end`
4. **Run**: `run` or `r`
5. **Inspect**: Examine registers and memory in TUI
6. **Step**: Use `step` or `next` to step through code
7. **Watch**: `watch R5` to monitor R5 changes
8. **Continue**: `continue` to run to next breakpoint
9. **Exit**: `quit`

### 6.2 Error Handling

**Parse Errors**:
- Display line number and column
- Highlight error in source
- Suggest corrections

**Runtime Errors**:
- Undefined instruction
- Invalid memory access
- Division by zero (if implemented)
- Stack overflow
- Maximum cycle limit exceeded

**Debugger Errors**:
- Invalid command
- Invalid address/register
- Breakpoint not found

### 6.3 Error Recovery and Safety

The emulator includes comprehensive safety mechanisms to prevent infinite loops, runaway execution, and resource exhaustion.

#### 6.3.1 Execution Limits

**Cycle Limit**:
- Default: 1,000,000 instructions per run
- Configurable via config file or command-line flag
- Prevents infinite loops in non-interactive mode
- Warning at 90% of limit
- Halt with error message when exceeded

**Time Limit**:
- Wall-clock timeout (default: 10 seconds)
- Useful for detecting performance issues
- Can be disabled in debug mode

**Memory Limits**:
- Maximum heap allocation (default: 16MB)
- Stack depth limit (default: 1MB)
- Prevent memory exhaustion

#### 6.3.2 Memory Protection

**Access Violations**:
- Read/write/execute permissions by segment
- Code segment: read-only after load
- Null pointer detection (address 0x0000)
- Uninitialized memory detection (optional)
- Out-of-bounds access detection

**Alignment Checking**:
- Word access must be 4-byte aligned
- Halfword access must be 2-byte aligned
- Optional strict mode (halt on misalignment)
- Permissive mode (allow unaligned with warning)

#### 6.3.3 Stack Protection

**Stack Overflow Detection**:
- Monitor SP relative to stack bounds
- Warning when stack usage exceeds 75%
- Error when stack overflows into other segments

**Stack Underflow Detection**:
- Detect SP going above initial stack pointer
- Catch excessive pops

**Stack Canary** (advanced):
- Place guard value at stack base
- Check on function return
- Detect buffer overflows

#### 6.3.4 Infinite Loop Detection

**Simple Detection**:
- Detect `B <current_address>` (branch to self)
- Detect small loops without system calls (< 10 instructions)
- Prompt user in debugger mode

**Heuristic Detection**:
- Track PC history
- Detect tight loops with no I/O or memory writes
- Configurable threshold

#### 6.3.5 Crash Reports

When a fatal error occurs, generate detailed crash report:

```
=== ARM2 Emulator Crash Report ===
Error: Memory access violation (read from 0x00000000)
Program: examples/test.s
Location: 0x8024 (test.s:42) in function 'process_data'

Registers:
  R0: 0x00000000  R1: 0x00008100  R2: 0x0000002A  R3: 0x00000001
  R4: 0x00000000  R5: 0x00000000  R6: 0x00000000  R7: 0x00000000
  R8: 0x00000000  R9: 0x00000000  R10: 0x00000000 R11: 0x00000000
  R12: 0x00000000 SP: 0x0003FFF8  LR: 0x00008010  PC: 0x00008024
  CPSR: 0x00000000 [----]

Call Stack:
  #0  0x00008024 in process_data (test.s:42)
  #1  0x00008010 in main (test.s:15)

Recent Instructions:
  0x00008018: MOV R2, #42
  0x0000801C: BL process_data
  0x00008020: B end
  0x00008024: LDR R0, [R0]     <-- CRASH

Memory at SP (stack):
  0x0003FFF8: 00008010 00000000 00000000 00000000

Execution Statistics:
  Instructions executed: 156
  Function calls: 2
  Memory accesses: 45
  Runtime: 0.0012 seconds

Save crash dump to 'crash-20250107-143022.dump'? [Y/n]
```

#### 6.3.6 Graceful Degradation

**On Warning**:
- Log warning to console/debugger
- Highlight issue in TUI
- Continue execution (unless strict mode)
- Accumulate warnings for report

**On Error**:
- Stop execution immediately
- Enter debugger (if in debug mode)
- Display crash report
- Offer to save state
- Exit gracefully (no panic)

## 7. Development Tools

### 7.1 Assembly Validator/Linter

A standalone tool to check assembly syntax without running the program.

**Features**:
- Syntax validation
- Undefined label detection
- Unreachable code detection
- Register usage warnings
- Best practice recommendations

**Usage**:
```bash
arm-lint program.s
arm-lint --strict program.s         # Enable all warnings
arm-lint --fix program.s            # Auto-fix formatting issues
```

**Example Output**:
```
program.s:15: warning: Label 'unused_func' defined but never referenced
program.s:23: error: Undefined label 'typo_label'
program.s:31: warning: Register R7 clobbered without saving
program.s:45: info: Consider using 'SUB SP, SP, #8' instead of multiple pushes

1 error, 2 warnings, 1 info
```

### 7.2 Assembly Formatter

Automatically format assembly code for consistency.

**Features**:
- Consistent indentation
- Align operands in columns
- Normalize spacing
- Comment alignment
- Label formatting

**Usage**:
```bash
arm-format program.s                # Print formatted version
arm-format -w program.s             # Write changes to file
arm-format --style=compact program.s # Use compact style
```

**Example**:
```asm
; Before
start:MOV R0,#10
ADD R1,R0,#5
  BL func

; After
start:  MOV     R0, #10
        ADD     R1, R0, #5
        BL      func
```

### 7.3 Symbol Cross-Reference Generator

Generate a cross-reference of all symbols in the program.

**Output**:
```
Symbol Cross-Reference for program.s
=====================================

_start                  Defined: line 10
                        Referenced: (entry point)

add_numbers            Defined: line 25
                        Referenced: line 15, line 42, line 58

data_buffer            Defined: line 100 (data)
                        Referenced: line 30, line 35

error_msg              Defined: line 105 (data)
                        Referenced: line 67
```

### 7.4 Interactive Instruction Reference

Built into the TUI debugger, accessible via `help` command.

**Features**:
- Instruction syntax
- Flags affected
- Cycle count (for timing simulation)
- Usage examples
- Common pitfalls

**Usage in Debugger**:
```
(arm-emu) help MOV
(arm-emu) help addressing modes
(arm-emu) help syscalls
```

### 7.5 Debugger Command Scripts

Support for batch command files to automate debugging workflows.

**Script Format** (`.gdb` style):
```
# script.cmd - Automated test script
load program.s
break main
run
print $R0
step 10
x/16x 0x8000
continue
quit
```

**Usage**:
```bash
arm-emu -x script.cmd program.s     # Execute script
arm-emu --batch -x test.cmd test.s  # Non-interactive mode
```

### 7.6 Conditional Breakpoints

Advanced breakpoint system with expression evaluation.

**Examples**:
```
(arm-emu) break loop if $R0 == 100
(arm-emu) break *0x8050 if $R1 > $R2
(arm-emu) break func if [$SP] != 0
```

**Supported Expressions**:
- Register comparisons
- Memory dereferencing
- Arithmetic operations
- Logical operations

## 8. Compatibility and Standards

### 8.1 ARM Architecture Version

**Target**: ARM2 (ARMv2 architecture)
- Specific variant: ARM2 or ARM2aS (document which)
- 32-bit architecture
- 26-bit addressing mode (PC[25:0] for address, PC[31:26] for flags)
- Note: This emulator uses full 32-bit addressing for simplicity

### 8.2 Instruction Set Compliance

**Fully Supported**:
- All ARM2 data processing instructions
- Basic load/store operations
- Branch instructions
- Multiply instructions

**Partially Supported**:
- ARM2a extensions (LDRH, STRH) - optional
- No support for ARM2 coprocessor instructions initially

**Deviations from Hardware**:
- No instruction timing (all instructions take 1 cycle)
- No pipeline simulation
- Simplified flag handling (where appropriate)
- Extended addressing (32-bit vs 26-bit)

### 8.3 Memory Alignment

**Default Behavior**:
- Word (32-bit) accesses: must be 4-byte aligned
- Halfword (16-bit) accesses: must be 2-byte aligned
- Byte (8-bit) accesses: no alignment required

**Configuration Options**:
- `strict_alignment`: Halt on misalignment (default: true)
- `permissive_alignment`: Allow unaligned with performance warning

### 8.4 Undefined Behavior

**Undefined Instructions**:
- Halt with error message
- Display instruction encoding
- Suggest similar instructions

**Unpredictable Results**:
Document behavior for edge cases:
- Using R15 (PC) as destination with writeback
- Simultaneous register modification
- Self-modifying code

### 8.5 Endianness

**Default**: Little-endian
**Configuration**: Option to switch to big-endian for compatibility

### 8.6 Calling Convention

Document recommended ARM calling convention:
- **Arguments**: R0-R3 (first 4 arguments)
- **Return value**: R0 (R1 for 64-bit returns)
- **Callee-saved**: R4-R11 (must preserve)
- **Caller-saved**: R0-R3, R12 (scratch)
- **Special**: R13=SP, R14=LR, R15=PC
- **Stack**: Full descending (grows downward)

### 8.7 Instruction Encoding Format

Document ARM2 instruction encoding patterns for educational purposes:

**Data Processing Instructions** (bits 27-26 = 00):
```
31-28: Condition
27-26: 00 (data processing)
25:    I (immediate operand)
24-21: Opcode (ADD=0100, SUB=0010, MOV=1101, etc.)
20:    S (set condition codes)
19-16: Rn (first operand register)
15-12: Rd (destination register)
11-0:  Operand 2 (immediate or shifted register)
```

**Multiply Instructions** (bits 27-22 = 000000, bits 7-4 = 1001):
```
31-28: Condition
27-22: 000000
21:    A (accumulate)
20:    S (set condition codes)
19-16: Rd (destination register)
15-12: Rn (accumulate register, MLA only)
11-8:  Rs (shift register)
7-4:   1001 (multiply signature)
3-0:   Rm (multiply register)
```

**Load/Store Instructions** (bits 27-26 = 01):
```
31-28: Condition
27-26: 01 (load/store)
25:    I (immediate offset vs register offset)
24:    P (pre/post indexing)
23:    U (up/down)
22:    B (byte/word)
21:    W (write-back)
20:    L (load/store)
19-16: Rn (base register)
15-12: Rd (source/dest register)
11-0:  Offset (immediate or register)
```

**Branch Instructions** (bits 27-25 = 101):
```
31-28: Condition
27-25: 101 (branch)
24:    L (link)
23-0:  Signed 24-bit offset (in words)
```

This encoding information helps implementers understand the binary format and aids in debugging, disassembly, and validation testing.

## 9. Logging and Diagnostics

### 9.1 Execution Trace

Record every instruction executed for analysis.

**Trace Format**:
```
[000001] 0x8000: MOV R0, #10           | R0=0x0000000A CPSR=---- | 0.001ms
[000002] 0x8004: MOV R1, #20           | R1=0x00000014 CPSR=---- | 0.002ms
[000003] 0x8008: ADD R2, R0, R1        | R2=0x0000001E CPSR=---- | 0.003ms
[000004] 0x800C: BL 0x8020              | LR=0x00008010 PC=0x8020 | 0.004ms
```

**Usage**:
```bash
arm-emu --trace trace.log program.s
arm-emu --trace-filter="R0,R1" program.s  # Only log R0, R1 changes
```

### 9.2 Memory Access Log

Track all memory reads and writes.

**Log Format**:
```
[READ ] 0x00008024: LDR R0, [R1] <- [0x00020000] = 0x12345678
[WRITE] 0x00008028: STR R0, [R2] -> [0x00020010] = 0x12345678
[READ ] 0x0000802C: LDRB R3, [R4] <- [0x00020004] = 0x42
```

**Usage**:
```bash
arm-emu --mem-trace mem.log program.s
arm-emu --mem-watch=0x8000-0x9000 program.s  # Watch specific range
```

### 9.3 Performance Statistics

Collect and display execution statistics.

**Metrics**:
- Total instructions executed
- Instructions per second
- Instruction type breakdown (data processing, memory, branch)
- Branch taken/not-taken ratio
- Function call count
- Memory access count
- Cache hit rate (if simulated)

**Output**:
```
=== Execution Statistics ===
Total Instructions: 1,234,567
Execution Time: 2.456 seconds
Instructions/sec: 502,686

Instruction Breakdown:
  Data Processing: 45.2% (558,000)
  Memory Access:   32.1% (396,000)
  Branch:          18.7% (231,000)
  System:           4.0% (49,567)

Branch Prediction:
  Taken: 67.3%
  Not Taken: 32.7%

Function Calls: 8,523
Average Call Depth: 3.2
Max Call Depth: 12

Memory Access:
  Reads: 245,678 (62%)
  Writes: 150,322 (38%)
  Total Bytes: 1.52 MB
```

**Usage**:
```bash
arm-emu --stats program.s
arm-emu --profile hotspots.txt program.s  # Identify hot code paths
```

### 9.4 Hot Path Analysis

Identify most frequently executed code regions.

**Output**:
```
=== Hot Paths (Top 10) ===
 1. loop_body (0x8040-0x8060):     45.2% (558,234 executions)
 2. multiply (0x8100-0x8120):      12.3% (151,876 executions)
 3. compare (0x8200-0x8210):        8.1% (100,045 executions)
...
```

### 9.5 Code Coverage

Track which instructions have been executed.

**Report**:
```
=== Code Coverage Report ===
Total Instructions: 1,523
Executed: 1,245 (81.7%)
Not Executed: 278 (18.3%)

Uncovered Functions:
  - error_handler (0x8500-0x8530) [never called]
  - debug_dump (0x8600-0x8650) [never called]

Partially Covered Functions:
  - main (0x8000-0x8100): 95.0%
  - process_data (0x8200-0x8300): 76.5%
```

### 9.6 Execution Export Formats

Export trace and statistics in various formats.

**Formats**:
- **JSON**: Machine-readable for analysis tools
- **CSV**: Import into spreadsheet programs
- **HTML**: Interactive visualization
- **VCD**: Value Change Dump (for waveform viewers)

**Example JSON**:
```json
{
  "program": "examples/sort.s",
  "total_cycles": 1234567,
  "execution_time_ms": 2456,
  "instructions": [
    {
      "cycle": 1,
      "address": "0x8000",
      "instruction": "MOV R0, #10",
      "registers": {"R0": "0x0000000A"},
      "flags": "----"
    }
  ],
  "statistics": {
    "instruction_count": 1234567,
    "function_calls": 8523
  }
}
```

## 10. Testing Strategy

**Testing Philosophy**: The emulator's correctness is critical since it serves as both an educational tool and a development platform. Comprehensive unit testing is essential to ensure each component behaves correctly in isolation before integration. Every instruction, every addressing mode, and every edge case must be thoroughly tested.

### 10.1 Unit Testing Requirements

Unit tests form the foundation of quality assurance. Each component must have comprehensive test coverage before integration.

#### 10.1.1 Instruction-Level Tests

**Test Coverage**: Every instruction must have a dedicated test suite.

**Test Structure**:
```go
// Example: test_mov.go
func TestMOV_Immediate(t *testing.T) {
    vm := NewVM()
    vm.Execute("MOV R0, #42")
    assert.Equal(t, uint32(42), vm.R[0])
    assert.Equal(t, false, vm.CPSR.N)
    assert.Equal(t, false, vm.CPSR.Z)
}

func TestMOV_Register(t *testing.T) {
    vm := NewVM()
    vm.R[1] = 100
    vm.Execute("MOV R0, R1")
    assert.Equal(t, uint32(100), vm.R[0])
}

func TestMOV_WithShift(t *testing.T) {
    vm := NewVM()
    vm.R[1] = 4
    vm.Execute("MOV R0, R1, LSL #2")
    assert.Equal(t, uint32(16), vm.R[0])
}
```

**Required Tests Per Instruction**:
1. **Basic operation**: Verify instruction performs correct computation
2. **Flag updates**: Test all CPSR flags (N, Z, C, V) are set correctly
3. **All addressing modes**: Test every supported addressing mode
4. **Condition codes**: Test all condition codes (EQ, NE, etc.)
5. **Edge cases**:
   - Zero operands
   - Maximum/minimum values
   - Overflow/underflow conditions
   - Negative numbers
6. **Register aliases**: Test SP, LR, PC special cases
7. **Boundary conditions**: Test register range limits (R0-R15)

**Example Test Categories**:

```
tests/instructions/
├── data_processing/
│   ├── test_mov.go          # 20+ test cases
│   ├── test_add.go          # 30+ test cases (with carry, overflow)
│   ├── test_sub.go          # 30+ test cases
│   ├── test_and.go          # 15+ test cases
│   ├── test_orr.go          # 15+ test cases
│   ├── test_eor.go          # 15+ test cases
│   ├── test_cmp.go          # 25+ test cases
│   └── ...
├── memory/
│   ├── test_ldr.go          # 40+ test cases (all addressing modes)
│   ├── test_str.go          # 40+ test cases
│   ├── test_ldrb.go         # 30+ test cases
│   ├── test_strb.go         # 30+ test cases
│   ├── test_ldm.go          # 30+ test cases (load multiple)
│   ├── test_stm.go          # 30+ test cases (store multiple)
│   └── ...
├── branch/
│   ├── test_b.go            # 20+ test cases (all conditions)
│   ├── test_bl.go           # 15+ test cases
│   └── ...
└── multiply/
    ├── test_mul.go          # 20+ test cases
    └── test_mla.go          # 20+ test cases
```

**Minimum Target**: 600+ unit tests for instruction execution alone.

#### 10.1.2 Flag Calculation Tests

Flags (N, Z, C, V) must be tested exhaustively as they control conditional execution.

**Test Cases**:
- **Negative flag (N)**: Result has bit 31 set
- **Zero flag (Z)**: Result equals zero
- **Carry flag (C)**:
  - Addition: Unsigned overflow
  - Subtraction: Unsigned borrow
  - Shifts: Last bit shifted out
- **Overflow flag (V)**: Signed overflow
  - Positive + Positive = Negative
  - Negative + Negative = Positive
  - Positive - Negative = Negative
  - Negative - Positive = Positive

**Example**:
```go
func TestADD_CarryFlag(t *testing.T) {
    vm := NewVM()
    vm.R[0] = 0xFFFFFFFF
    vm.R[1] = 1
    vm.Execute("ADDS R2, R0, R1")
    assert.Equal(t, uint32(0), vm.R[2])
    assert.Equal(t, true, vm.CPSR.Z)   // Zero result
    assert.Equal(t, true, vm.CPSR.C)   // Carry occurred
}

func TestADD_OverflowFlag(t *testing.T) {
    vm := NewVM()
    vm.R[0] = 0x7FFFFFFF  // Max positive int32
    vm.R[1] = 1
    vm.Execute("ADDS R2, R0, R1")
    assert.Equal(t, uint32(0x80000000), vm.R[2])
    assert.Equal(t, true, vm.CPSR.N)   // Negative result
    assert.Equal(t, true, vm.CPSR.V)   // Overflow occurred
    assert.Equal(t, false, vm.CPSR.C)  // No carry
}
```

**Target**: 100+ flag calculation tests.

#### 10.1.3 Memory System Tests

Memory operations must handle alignment, permissions, and boundaries correctly.

**Test Categories**:
```go
// Alignment tests
func TestMemory_AlignedWordAccess(t *testing.T)
func TestMemory_UnalignedWordAccess(t *testing.T)
func TestMemory_AlignedHalfwordAccess(t *testing.T)
func TestMemory_UnalignedHalfwordAccess(t *testing.T)

// Permission tests
func TestMemory_WriteToCodeSegment(t *testing.T)      // Should fail
func TestMemory_ReadFromDataSegment(t *testing.T)     // Should succeed
func TestMemory_ExecuteFromDataSegment(t *testing.T)  // Optional: should warn

// Boundary tests
func TestMemory_ReadAtZero(t *testing.T)              // Null pointer
func TestMemory_ReadBeyondMemory(t *testing.T)        // Out of bounds
func TestMemory_StackOverflow(t *testing.T)
func TestMemory_StackUnderflow(t *testing.T)

// Endianness tests
func TestMemory_LittleEndianLoad(t *testing.T)
func TestMemory_LittleEndianStore(t *testing.T)
```

**Target**: 50+ memory system tests.

#### 10.1.4 Parser Tests

The parser must handle all syntax variations, directives, and error conditions.

**Test Cases**:
```go
// Instruction parsing
func TestParser_BasicInstruction(t *testing.T)
func TestParser_InstructionWithCondition(t *testing.T)
func TestParser_InstructionWithSFlag(t *testing.T)
func TestParser_AllAddressingModes(t *testing.T)

// Label parsing
func TestParser_LabelDefinition(t *testing.T)
func TestParser_LabelReference(t *testing.T)
func TestParser_ForwardReference(t *testing.T)
func TestParser_LocalLabels(t *testing.T)
func TestParser_NumericLabels(t *testing.T)

// Directive parsing
func TestParser_OrgDirective(t *testing.T)
func TestParser_EquDirective(t *testing.T)
func TestParser_DataDirectives(t *testing.T)
func TestParser_AlignDirective(t *testing.T)
func TestParser_IncludeDirective(t *testing.T)
func TestParser_MacroDirective(t *testing.T)
func TestParser_ConditionalDirective(t *testing.T)

// Comment handling
func TestParser_LineComments(t *testing.T)
func TestParser_BlockComments(t *testing.T)
func TestParser_InlineComments(t *testing.T)

// Error handling
func TestParser_InvalidInstruction(t *testing.T)
func TestParser_InvalidRegister(t *testing.T)
func TestParser_UndefinedLabel(t *testing.T)
func TestParser_SyntaxError(t *testing.T)
func TestParser_DuplicateLabel(t *testing.T)
func TestParser_InvalidDirective(t *testing.T)

// Edge cases
func TestParser_EmptyFile(t *testing.T)
func TestParser_OnlyComments(t *testing.T)
func TestParser_VeryLongLines(t *testing.T)
func TestParser_NestedIncludes(t *testing.T)
func TestParser_CircularIncludes(t *testing.T)
```

**Target**: 90+ parser tests.

#### 10.1.5 Addressing Mode Tests

Each addressing mode must be tested with multiple instructions.

**Addressing Modes to Test**:
1. Immediate: `#value`
2. Register: `Rn`
3. Register with shift: `Rm, LSL #n`, `Rm, LSR #n`, `Rm, ASR #n`, `Rm, ROR #n`
4. Register shift by register: `Rm, LSL Rs`
5. Memory offset: `[Rn, #offset]`
6. Memory pre-indexed: `[Rn, #offset]!`
7. Memory post-indexed: `[Rn], #offset`
8. Memory register offset: `[Rn, Rm]`
9. Memory register offset with shift: `[Rn, Rm, LSL #n]`

**Example**:
```go
func TestAddressingMode_ImmediateShift(t *testing.T) {
    tests := []struct {
        shift    string
        value    uint32
        amount   uint32
        expected uint32
    }{
        {"LSL", 1, 4, 16},
        {"LSR", 16, 2, 4},
        {"ASR", 0x80000000, 1, 0xC0000000},
        {"ROR", 0x00000003, 1, 0x80000001},
    }

    for _, tt := range tests {
        vm := NewVM()
        vm.R[1] = tt.value
        vm.Execute(fmt.Sprintf("MOV R0, R1, %s #%d", tt.shift, tt.amount))
        assert.Equal(t, tt.expected, vm.R[0], tt.shift)
    }
}
```

**Target**: 60+ addressing mode tests.

#### 10.1.6 System Call Tests

Each SWI syscall must be tested for correct behavior.

**Test Cases**:
```go
func TestSyscall_Exit(t *testing.T)
func TestSyscall_WriteChar(t *testing.T)
func TestSyscall_WriteString(t *testing.T)
func TestSyscall_WriteNewline(t *testing.T)
func TestSyscall_ReadChar(t *testing.T)
func TestSyscall_Allocate(t *testing.T)
func TestSyscall_Free(t *testing.T)
func TestSyscall_Reallocate(t *testing.T)
func TestSyscall_GetTime(t *testing.T)
func TestSyscall_GetRandom(t *testing.T)
func TestSyscall_InvalidNumber(t *testing.T)
func TestSyscall_FileOperations(t *testing.T)
func TestSyscall_ErrorHandling(t *testing.T)
```

**Target**: 30+ syscall tests.

#### 10.1.7 Test Coverage Metrics

**Minimum Coverage Requirements**:
- Instruction execution: **95%** code coverage
- Memory system: **90%** code coverage
- Parser: **85%** code coverage
- VM core: **90%** code coverage
- Overall: **85%** code coverage

**Tools**:
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 10.2 Integration Tests

Integration tests verify that components work correctly together.

#### 10.2.1 Complete Program Tests

**Test Programs**:
```
tests/programs/
├── factorial.s          # Recursive function calls
├── fibonacci.s          # Iterative loops
├── bubble_sort.s        # Array manipulation
├── string_ops.s         # String operations
├── linked_list.s        # Dynamic memory
├── nested_calls.s       # Deep call stack
└── ...
```

**Each test verifies**:
- Program completes successfully
- Correct output produced
- Final register/memory state as expected
- No memory leaks (if using heap)
- Stack properly maintained

#### 10.2.2 Cross-Component Tests

Test interactions between VM components:
- Parser → VM: Parsed instructions execute correctly
- VM → Memory: Instructions access memory correctly
- VM → Syscalls: System calls work during execution
- Debugger → VM: Breakpoints interrupt correctly

#### 10.2.3 Regression Tests

Maintain a suite of tests for previously found bugs:
```go
// Bug #42: LDR post-indexed addressing incorrect
func TestRegression_Bug42_LDRPostIndexed(t *testing.T) {
    vm := NewVM()
    vm.Memory[0x8000] = 0x12345678
    vm.R[1] = 0x8000
    vm.Execute("LDR R0, [R1], #4")
    assert.Equal(t, uint32(0x12345678), vm.R[0])
    assert.Equal(t, uint32(0x8004), vm.R[1]) // R1 should be updated
}
```

**Target**: 30+ regression tests.

### 10.3 Debugger Tests

#### 10.3.1 Breakpoint Tests
- Set breakpoint at address
- Set breakpoint at label
- Conditional breakpoints
- Temporary breakpoints
- Multiple breakpoints

#### 10.3.2 Execution Control Tests
- Step single instruction
- Step over function calls
- Step into function calls
- Run to completion
- Continue from breakpoint

#### 10.3.3 State Inspection Tests
- Read register values
- Read memory values
- Evaluate expressions
- Display call stack
- Symbol resolution

#### 10.3.4 Watchpoint Tests
- Watch register changes
- Watch memory writes
- Watch memory reads
- Multiple watchpoints

**Target**: 40+ debugger tests.

### 10.4 Test Organization

```
tests/
├── unit/
│   ├── instructions/      # 600+ tests
│   ├── memory/            # 50+ tests
│   ├── parser/            # 90+ tests
│   ├── flags/             # 100+ tests
│   ├── addressing/        # 60+ tests
│   └── syscall/           # 30+ tests
├── integration/
│   ├── programs/          # 20+ tests
│   ├── cross_component/   # 15+ tests
│   └── regression/        # 30+ tests
├── debugger/              # 40+ tests
└── performance/           # 10+ benchmarks
```

**Total Target**: 1000+ tests minimum.

### 10.5 Continuous Integration

**Automated Testing**:
- Run all tests on every commit
- Enforce minimum coverage thresholds
- Generate coverage reports
- Run performance benchmarks
- Detect performance regressions

**CI Configuration** (example):
```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.25
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 85" | bc -l) )); then
            echo "Coverage $coverage% is below 85% threshold"
            exit 1
          fi
      - name: Run benchmarks
        run: go test -bench=. -benchmem ./...
```

### 10.6 Test-Driven Development

**Recommended Workflow**:
1. Write test for new instruction/feature (test fails)
2. Implement minimum code to pass test
3. Refactor for clarity
4. Add edge case tests
5. Verify all tests pass
6. Check coverage increased

**Example**:
```go
// Step 1: Write test first
func TestRSB_BasicOperation(t *testing.T) {
    vm := NewVM()
    vm.R[0] = 5
    vm.R[1] = 3
    vm.Execute("RSB R2, R0, R1")  // R2 = R1 - R0 = 3 - 5 = -2
    assert.Equal(t, uint32(0xFFFFFFFE), vm.R[2])  // -2 in two's complement
}

// Step 2: Implement RSB instruction
// Step 3: Test passes
// Step 4: Add edge cases
func TestRSB_WithOverflow(t *testing.T) { ... }
func TestRSB_WithCarry(t *testing.T) { ... }
```

### 10.7 Documentation Requirements

Every test should include:
- **Clear name**: Describes what is being tested
- **Comment**: Explains why the test is important
- **Assertions**: Verify all relevant outputs
- **Error messages**: Helpful failure messages

**Example**:
```go
// TestADD_SignedOverflow verifies that adding two positive numbers
// that exceed INT32_MAX correctly sets the overflow flag while
// NOT setting the carry flag (carry is for unsigned overflow).
// This is critical for conditional execution based on signed comparisons.
func TestADD_SignedOverflow(t *testing.T) {
    vm := NewVM()
    vm.R[0] = 0x7FFFFFFF  // INT32_MAX
    vm.R[1] = 1
    vm.Execute("ADDS R2, R0, R1")

    assert.Equal(t, uint32(0x80000000), vm.R[2],
        "Result should be 0x80000000")
    assert.Equal(t, true, vm.CPSR.V,
        "Overflow flag must be set for signed overflow")
    assert.Equal(t, false, vm.CPSR.C,
        "Carry flag should NOT be set (no unsigned overflow)")
    assert.Equal(t, true, vm.CPSR.N,
        "Negative flag must be set (result is negative)")
}
```

## 11. Future Enhancements

### 11.1 Phase 2 Features
- Assembly optimization hints
- Performance profiling
- Code coverage analysis
- Trace logging

### 11.2 Phase 3 Features
- Time-travel debugging (record/replay)
- Remote debugging protocol
- Scripting support (Python/Lua)
- Machine code generation
- ARM3/ARM6 instruction support

### 11.3 Nice-to-Have
- GUI frontend (web-based or native)
- Visual memory map
- Register diff view
- Instruction timing simulation
- Interrupt support
- Co-processor support

## 12. Documentation Requirements

### 12.1 User Documentation
- Installation guide
- Assembly language reference
- Debugger command reference
- Tutorial with examples
- FAQ

### 12.2 Developer Documentation
- API reference
- Architecture overview
- Contributing guidelines
- Coding standards

## 13. Cross-Platform Compatibility

**Platform Support**: The emulator must be fully cross-platform, working seamlessly on macOS, Windows, and Linux. This is a critical requirement for both educational use and development accessibility.

### 13.1 Platform Requirements

**Primary Platforms**:
- **macOS**: 17.0 (Sequoia) or later (Apple Silicon)
- **Windows**: Windows 10 or later (64-bit)
- **Linux**: Modern distributions (Ubuntu 20.04+, Fedora 34+, Arch, etc.)

**Architecture Support**:
- x86_64 (AMD64)
- ARM64 (Apple Silicon, ARM servers)

### 13.2 Cross-Platform Implementation Guidelines

#### 13.2.1 File System Handling

**Path Separators**:
```go
// ✓ CORRECT: Use filepath.Join for cross-platform paths
configPath := filepath.Join(homeDir, ".config", "arm-emu", "config.toml")

// ✗ WRONG: Hard-coded separators
configPath := homeDir + "/.config/arm-emu/config.toml"  // Fails on Windows
```

**Home Directory**:
```go
// ✓ CORRECT: Cross-platform home directory
homeDir, err := os.UserHomeDir()

// Platform-specific config locations
// Linux/macOS: ~/.config/arm-emu/
// Windows: %APPDATA%\arm-emu\
```

**File Permissions**:
- Use os.FileMode consistently
- Test permission handling on all platforms
- Windows doesn't support Unix permissions (handle gracefully)

**Line Endings**:
- Accept both CRLF (Windows) and LF (Unix) in assembly files
- Use `bufio.Scanner` which handles both automatically
- Write output with platform-appropriate line endings when needed

#### 13.2.2 Terminal/Console Handling

**TUI Library Requirements**:
- Must work on all three platforms
- Handle terminal resize events
- Support different terminal emulators:
  - macOS: Terminal.app, iTerm2
  - Windows: cmd.exe, PowerShell, Windows Terminal
  - Linux: gnome-terminal, konsole, xterm, etc.

**Color Support**:
- Detect terminal capabilities (ANSI color support)
- Graceful fallback to no-color mode
- Test with `NO_COLOR` environment variable

**Keyboard Input**:
- Handle platform differences in key codes
- Arrow keys, function keys, modifiers
- Ctrl+C signal handling (cross-platform)

#### 13.2.3 Process Management

**Signal Handling**:
```go
// ✓ CORRECT: Cross-platform signal handling
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

// Handle Windows differences (no SIGTERM on Windows)
```

**Exit Codes**:
- Use standard exit codes (0 = success, non-zero = error)
- Document exit code meanings

#### 13.2.4 Build System

**Build Tags** (if needed):
```go
// +build windows

// platform_windows.go - Windows-specific code

// +build !windows

// platform_unix.go - Unix-specific code
```

**Cross-Compilation**:
```bash
# Build for all platforms
make build-all

# Produces:
# arm-emu-darwin-amd64
# arm-emu-darwin-arm64
# arm-emu-linux-amd64
# arm-emu-linux-arm64
# arm-emu-windows-amd64.exe
```

**Makefile/Build Scripts**:
```makefile
# Cross-platform Makefile
BINARY_NAME=arm-emu
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

ifeq ($(GOOS),windows)
    BINARY_NAME := $(BINARY_NAME).exe
endif

build:
	go build -o $(BINARY_NAME) .

build-all:
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64.exe .
```

#### 13.2.5 Testing on All Platforms

**CI/CD Requirements**:
```yaml
# .github/workflows/test.yml
name: Cross-Platform Tests

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: [1.25]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v2
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: ${{ matrix.go-version }}
      - name: Run tests
        run: go test -v ./...
      - name: Build
        run: go build -v .
```

**Manual Testing Checklist**:
- [ ] Install and run on macOS
- [ ] Install and run on Windows 10/11
- [ ] Install and run on Linux (Ubuntu, Fedora)
- [ ] TUI renders correctly on all platforms
- [ ] File I/O works correctly
- [ ] Path handling works
- [ ] Config file loading works
- [ ] Example programs run identically

### 13.3 Platform-Specific Considerations

#### 13.3.1 macOS

**Considerations**:
- Gatekeeper and code signing (for distribution)
- Apple Silicon support
- Case-sensitive vs case-insensitive filesystems
- Terminal.app vs iTerm2 differences

**Testing**:
- Test on Apple Silicon Macs if possible
- Test with Terminal.app

#### 13.3.2 Windows

**Considerations**:
- Backslash path separators
- Different terminal capabilities (cmd vs PowerShell vs Windows Terminal)
- No ANSI color support in older cmd.exe (Windows < 10)
- File locking behavior differs
- Case-insensitive filesystem
- Line endings (CRLF)

**Testing**:
- Test on Windows 10 and 11
- Test in cmd.exe, PowerShell, and Windows Terminal
- Verify file paths with spaces work
- Test with both forward and back slashes in paths

#### 13.3.3 Linux

**Considerations**:
- Wide variety of distributions
- Different terminal emulators
- Case-sensitive filesystem
- Different package managers

**Testing**:
- Test on at least Ubuntu and Fedora
- Test in multiple terminal emulators
- Verify no hardcoded /usr paths
- Test with different locale settings

### 13.4 Platform-Agnostic Coding Practices

**Use Standard Library**:
```go
// ✓ CORRECT: Use os and filepath packages
import (
    "os"
    "path/filepath"
)

// ✗ WRONG: Don't use platform-specific packages unnecessarily
import "syscall"  // Only if absolutely needed
```

**Avoid Platform-Specific Commands**:
```go
// ✗ WRONG: Calling shell commands
exec.Command("ls", "-la")       // Unix only
exec.Command("dir")             // Windows only

// ✓ CORRECT: Use Go's built-in functions
filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
    // Platform-agnostic directory traversal
})
```

**Handle Permissions Carefully**:
```go
// ✓ CORRECT: Use standard permissions
file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)

// Note: Windows will handle these differently but gracefully
```

### 13.5 Distribution and Installation

**Binary Distribution**:
- Provide pre-built binaries for all three platforms
- Use GitHub Releases with platform-specific downloads
- Include installation instructions for each platform

**Package Managers**:
- **macOS**: Homebrew formula
- **Linux**: DEB/RPM packages, or distribution-specific repos
- **Windows**: Chocolatey package (optional)
- **Cross-platform**: Go install (for Go developers)

**Installation Methods**:
```bash
# macOS (Homebrew)
brew install arm-emu

# Linux (from source)
git clone https://github.com/user/arm-emu
cd arm-emu
make install

# Windows (pre-built binary)
# Download arm-emu-windows-amd64.exe
# Add to PATH

# Go developers (all platforms)
go install github.com/user/arm-emu@latest
```

### 13.6 Documentation for Platform Differences

**User Documentation Must Include**:
- Platform-specific installation instructions
- Known platform limitations
- Platform-specific configuration examples
- Troubleshooting section for each platform

**Example**:
```markdown
## Platform-Specific Notes

### macOS
- Config file location: `~/.config/arm-emu/config.toml`
- Log file location: `~/Library/Logs/arm-emu/`

### Windows
- Config file location: `%APPDATA%\arm-emu\config.toml`
- Log file location: `%LOCALAPPDATA%\arm-emu\logs\`

### Linux
- Config file location: `~/.config/arm-emu/config.toml`
- Log file location: `~/.local/share/arm-emu/logs/`
```

## 14. Performance Considerations

### 14.1 Optimization Targets
- Parser: < 100ms for typical programs (< 1000 lines)
- Execution: > 100k instructions/second interpreted
- Memory: < 100MB for typical programs
- TUI refresh: 60 FPS minimum

### 14.2 Profiling Points
- Instruction dispatch overhead
- Memory access latency
- TUI render time
- Symbol lookup time

### 14.3 Instruction Timing (Optional Feature)

For educational purposes and performance simulation, the emulator can optionally simulate ARM2 instruction timing:

**Typical ARM2 Cycle Counts**:
- Data processing (register): 1 cycle
- Data processing (immediate): 1 cycle
- Data processing (shifted register): 1 cycle + shift cycles
- Load/Store word: 2-3 cycles (depending on cache/memory)
- Load/Store byte: 2-3 cycles
- Load/Store multiple: 2 + n cycles (n = number of registers)
- Multiply: 2-16 cycles (depends on operand values)
- Multiply-accumulate: 2-16 cycles
- Branch: 2-3 cycles (pipeline flush)
- Branch with link: 2-3 cycles

**Shift Operation Cycles**:
- LSL/LSR/ASR/ROR by immediate: 0 additional cycles (part of instruction)
- LSL/LSR/ASR/ROR by register: 1 additional cycle

**Configuration**:
```toml
[performance]
enable_timing = true           # Enable cycle counting
simulate_cache = false         # Simulate cache hits/misses (advanced)
memory_latency = 3            # Memory access cycles
```

**Usage**:
- Track total cycles executed
- Identify hot paths by cycle count
- Compare algorithm efficiency
- Educational tool for understanding performance

## 15. Dependencies

### 15.1 Core Dependencies
- Go 1.25+ (or target language)
- TUI library (tview, bubbletea, or equivalent)
- CLI library (cobra, or equivalent)
- Configuration parser (TOML/JSON)

**Cross-Platform Requirements**:
- All dependencies must support macOS, Windows, and Linux
- TUI library must handle different terminal emulators correctly
- No platform-specific dependencies in core functionality

### 15.2 Development Dependencies
- Testing framework
- Benchmarking tools
- Linter/formatter
- Documentation generator

## 16. Milestones

### Milestone 1: Core VM
- Basic VM implementation
- Essential instruction set (MOV, ADD, SUB, B)
- Simple memory model
- Basic execution loop

### Milestone 2: Parser
- Complete lexer and parser
- All ARM2 instructions
- Directives support
- Error reporting

### Milestone 3: Debugger Foundation
- Command processor
- Breakpoints
- Basic TUI layout
- Step execution

### Milestone 4: Full TUI
- All TUI panels
- Syntax highlighting
- Memory inspection
- Register display

### Milestone 5: Advanced Debugging
- Watchpoints
- Call stack
- Conditional breakpoints
- Expression evaluation

### Milestone 6: System Integration
- System calls (SWI instruction)
- I/O functionality
- Standard library macros
- Startup/shutdown sequences

### Milestone 7: Development Tools
- Assembly linter
- Code formatter
- Symbol cross-reference
- Instruction reference

### Milestone 8: Diagnostics
- Execution trace
- Memory access logging
- Performance statistics
- Code coverage

### Milestone 9: Polish
- Documentation
- Complete test suite
- Performance optimization
- Example programs suite

## 17. Security and Sandboxing

### 17.1 Isolation Considerations

The emulator must safely execute untrusted assembly code without compromising the host system:

**File System Access**:
- Limit file operations to designated sandbox directory
- Reject absolute paths outside sandbox
- Prevent access to sensitive system files
- Optional: Disable file syscalls entirely with `--no-file-io` flag

**Resource Limits**:
- Memory allocation cap (default: 16MB heap)
- Maximum file descriptors (default: 10)
- Maximum file size (default: 10MB)
- CPU cycle limit to prevent infinite loops
- Wall-clock timeout

**System Calls**:
- Whitelist of allowed syscalls
- No access to network operations
- No process spawning
- No system command execution

**Configuration**:
```toml
[security]
enable_sandbox = true
sandbox_directory = "./sandbox"
max_heap_size = 16777216        # 16MB
max_open_files = 10
max_file_size = 10485760        # 10MB
allow_file_io = true
allowed_file_extensions = [".txt", ".dat"]
```

### 17.2 Educational Use Safety

For classroom/learning environments:

**Safe Mode**:
```bash
arm-emu --safe-mode program.s
```
Enables:
- Strict sandboxing
- No file I/O
- Limited syscalls (console I/O only)
- Reduced resource limits
- Automatic timeout

**Automated Grading Integration**:
- JSON output format for test results
- Non-interactive batch mode
- Timeout protection
- Memory leak detection
- Deterministic execution (disable randomness)

### 17.3 Untrusted Code Warnings

Display warnings when potentially dangerous patterns detected:
- Self-modifying code
- Excessive memory allocation
- Suspicious file operations
- Unusual syscall patterns

---

## Appendix A: Command-Line Interface

### A.1 Main Emulator Command

```bash
arm-emu [OPTIONS] <program.s>

Options:
  -h, --help                Show help message
  -v, --version             Show version information
  -d, --debug               Start in debug mode (default)
  -r, --run                 Run without debugger
  -c, --config <file>       Use configuration file
  -x, --exec <script>       Execute debugger script
  --batch                   Non-interactive mode
  --trace <file>            Enable execution trace
  --mem-trace <file>        Enable memory access trace
  --stats                   Show execution statistics
  --profile <file>          Generate profile report
  --strict                  Enable strict error checking
  --max-cycles <n>          Set instruction limit (default: 1000000)
  --timeout <n>             Set timeout in seconds (default: 10)
  -o, --output <file>       Redirect program output to file

Examples:
  arm-emu program.s                       # Load and debug
  arm-emu -r program.s                    # Run directly
  arm-emu --trace trace.log program.s     # Run with tracing
  arm-emu -x test.cmd --batch test.s      # Automated testing
```

### A.2 Linter Command

```bash
arm-lint [OPTIONS] <program.s>

Options:
  -h, --help                Show help message
  --strict                  Enable all warnings
  --fix                     Auto-fix formatting issues
  -o, --output <file>       Write fixed code to file
  --format <style>          Output format (text|json|gcc)

Examples:
  arm-lint program.s                      # Check syntax
  arm-lint --strict program.s             # All warnings
  arm-lint --fix -o fixed.s program.s     # Fix and save
```

### A.3 Formatter Command

```bash
arm-format [OPTIONS] <program.s>

Options:
  -h, --help                Show help message
  -w, --write               Write changes to file
  -o, --output <file>       Write to different file
  --style <name>            Formatting style (default|compact|aligned)
  --indent <n>              Indentation size (default: 8)
  --check                   Check if formatted (exit code)

Examples:
  arm-format program.s                    # Print formatted
  arm-format -w program.s                 # Format in-place
  arm-format --style compact program.s    # Use compact style
```

## Appendix B: Configuration File Examples

### B.1 Basic Configuration

```toml
# arm-emu.toml - Basic configuration

[emulator]
variant = "ARM2aS"              # ARM2, ARM2a, or ARM2aS
strict_mode = true              # Halt on warnings
max_cycles = 1000000            # Instruction limit
timeout_seconds = 10            # Execution timeout

[memory]
code_start = 0x00008000
code_size = 0x00010000          # 64KB
data_start = 0x00020000
data_size = 0x00010000          # 64KB
stack_start = 0x00040000
stack_size = 0x00010000         # 64KB
heap_start = 0x00030000
heap_size = 0x00010000          # 64KB

alignment_checking = true
little_endian = true

[debugger]
auto_start = true               # Start in debugger
show_source = true
show_registers = true
show_memory = true
show_stack = true
syntax_highlighting = true

default_breakpoints = [
    "main",
    "_start"
]

[logging]
trace_execution = false
trace_memory = false
log_file = "arm-emu.log"
```

### B.2 Development Configuration

```toml
# arm-emu-dev.toml - Development/testing configuration

[emulator]
variant = "ARM2aS"
strict_mode = false             # Allow warnings
max_cycles = 10000000           # Higher limit
timeout_seconds = 0             # No timeout

[memory]
code_start = 0x00008000
code_size = 0x00100000          # 1MB for large programs
data_start = 0x00200000
data_size = 0x00100000          # 1MB
stack_start = 0x00400000
stack_size = 0x00100000         # 1MB
heap_start = 0x00300000
heap_size = 0x00100000          # 1MB

[debugger]
auto_start = false              # Run directly
command_history_size = 1000

[logging]
trace_execution = true
trace_memory = true
trace_file = "trace.log"
stats_file = "stats.json"
coverage_file = "coverage.html"

[performance]
enable_profiling = true
enable_coverage = true
```

## Appendix C: Assembly Language Style Guide

### C.1 Naming Conventions

**Labels**:
- Functions: `lowercase_with_underscores`
- Local labels: `.local_label` (starts with dot)
- Constants: `UPPERCASE_WITH_UNDERSCORES`
- Global data: `lowercase_with_underscores`

**Comments**:
- File header: Describe purpose, author, date
- Function header: Describe parameters, return value, side effects
- Inline: Explain non-obvious operations

### C.2 Code Organization

```asm
; =============================================================================
; Program: example.s
; Description: Demonstrates proper assembly style
; Author: Your Name
; Date: 2025-10-07
; =============================================================================

        .org 0x8000

; =============================================================================
; Constants
; =============================================================================

        .equ MAX_SIZE, 100
        .equ BUFFER_SIZE, 256

; =============================================================================
; Entry Point
; =============================================================================

_start:
        ; Initialize system
        BL      init_system

        ; Main program logic
        BL      main

        ; Clean shutdown
        MOV     R0, #0
        SWI     #0x00

; =============================================================================
; Function: main
; Description: Main program entry point
; Parameters: None
; Returns: R0 = exit code
; Clobbers: R0-R3, R12
; =============================================================================

main:
        STMFD   SP!, {R4-R11, LR}    ; Save callee-saved registers

        ; Function body
        MOV     R0, #42
        BL      process_value

        MOV     R0, #0               ; Success
        LDMFD   SP!, {R4-R11, PC}    ; Restore and return

; =============================================================================
; Data Section
; =============================================================================

        .align 4
message:
        .asciz  "Hello, World!\n"

buffer:
        .space  BUFFER_SIZE
```

### C.3 Formatting Rules

1. **Indentation**: 8 spaces for instructions (or 1 tab)
2. **Alignment**: Align operands for readability
3. **Spacing**: One space after commas
4. **Comments**: Start at column 40 or after instruction
5. **Labels**: No indentation for global, indent for local

### C.4 Common Pitfalls and Gotchas

**PC-Relative Addressing**:
- ARM2 reads PC as current instruction + 8 (pipeline effect)
- Be careful with PC-relative loads: `LDR R0, [PC, #offset]`
- Offset calculation must account for pipeline

**Condition Code Usage**:
- Remember that condition codes persist across instructions
- Don't assume flags are clear unless you explicitly set them
- TST/TEQ/CMP/CMN set flags but don't write results

**Register R15 (PC) Caveats**:
- Writing to R15 causes a branch
- Reading R15 gives address + 8 (not current instruction)
- Using R15 as destination with writeback: behavior is unpredictable
- Avoid: `LDR R15, [R0], #4` (post-indexed with PC)

**Stack Management**:
- Always balance pushes and pops
- STMFD/LDMFD for full descending stack (most common)
- Remember to save LR before calling other functions
- Don't forget to restore LR before returning

**Immediate Value Limitations**:
- ARM immediate values are 8-bit rotated
- Not all 32-bit values can be immediate operands
- Use `LDR Rd, =constant` for arbitrary 32-bit values (literal pool)
- Example: `MOV R0, #0x1FF` works, but `MOV R0, #0x101` doesn't

**Shift Operations**:
- LSL #0 is a no-op (doesn't shift)
- Shift by 32 or more: result is often unpredictable
- ROR #0 is actually RRX (rotate right with extend)

**Memory Alignment**:
- Word loads/stores must be 4-byte aligned
- Halfword loads/stores must be 2-byte aligned
- Unaligned access causes error (or silent misalignment on some systems)
- Use `.align 4` before data that needs word alignment

**Label Scope**:
- Global labels are visible everywhere
- Local labels (starting with .) are scoped to previous global label
- Forward references must resolve in second pass

**Multiply Instruction Restrictions**:
- MUL/MLA: Rd and Rm must be different registers
- Result is lower 32 bits only (no 64-bit result)
- Timing depends on value of Rs (2-16 cycles)

**Branch Range**:
- Branch offset is 24-bit signed (± 32MB range)
- Offset is in words (x4 for byte address)
- Long branches may need multiple instructions

**Load/Store Multiple Order**:
- Registers always transferred in numerical order (R0, R1, R2...)
- Register list in instruction doesn't affect order
- `STMFD SP!, {R0, R5, R2}` stores in order: R0, R2, R5

**Carry Flag in Shifts**:
- Carry flag gets last bit shifted out
- This is useful for multi-word arithmetic
- ADC (add with carry) continues carry chain

## Appendix D: Sample Programs

See the `examples/` directory for complete sample programs demonstrating:

- Basic I/O operations
- Arithmetic and logic
- Control flow
- Function calls
- Data structures
- Algorithm implementations

## Appendix E: References

### E.1 ARM Architecture Documentation
- ARM Architecture Reference Manual (ARMv2)
- ARM2 Data Sheet
- ARM Assembly Language Programming (books)

### E.2 Similar Projects
- GNU ARM Assembler (as)
- ARM Development Studio
- QEMU ARM Emulation
- VisUAL (ARM simulator with visualization)

### E.3 Learning Resources
- ARM Assembly Language Tutorial
- Low-Level Programming University
- Computer Organization and Design (Patterson & Hennessy)
