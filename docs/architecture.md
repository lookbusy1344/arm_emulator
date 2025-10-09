# Architecture Overview

This document describes the internal architecture of the ARM2 Emulator.

## Project Structure

```
arm_emulator/
├── main.go              # Entry point and CLI
├── vm/                  # Virtual machine core
│   ├── cpu.go          # CPU state and registers
│   ├── memory.go       # Memory management
│   ├── executor.go     # Fetch-decode-execute cycle
│   ├── flags.go        # CPSR flag operations
│   └── syscall.go      # System call handler
├── parser/              # Assembly parser
│   ├── lexer.go        # Tokenization
│   ├── parser.go       # Syntax analysis
│   ├── symbols.go      # Symbol table
│   ├── preprocessor.go # Includes and conditionals
│   └── macros.go       # Macro expansion
├── instructions/        # Instruction implementations
│   ├── data_processing.go  # MOV, ADD, SUB, etc.
│   ├── memory.go           # LDR, STR
│   ├── memory_multi.go     # LDM, STM
│   ├── branch.go           # B, BL, BX
│   └── multiply.go         # MUL, MLA
├── debugger/            # Debugging support
│   ├── debugger.go     # Main debugger logic
│   ├── commands.go     # Command interpreter
│   ├── breakpoints.go  # Breakpoint management
│   ├── watchpoints.go  # Watchpoint management
│   ├── expressions.go  # Expression evaluator
│   ├── history.go      # Command history
│   └── tui.go          # Text UI
├── tools/               # Development tools
│   ├── lint.go         # Assembly linter
│   ├── format.go       # Code formatter
│   └── xref.go         # Cross-reference generator
├── tests/               # Test files
└── examples/            # Example programs
```

## Module Overview

### 1. Virtual Machine (vm/)

The VM module implements the ARM2 processor and memory system.

#### CPU (cpu.go)

**Core Components:**
- 16 general-purpose registers (R0-R15)
- CPSR (Current Program Status Register) with N, Z, C, V flags
- Cycle counter for performance analysis

**Key Types:**
```go
type VM struct {
    R     [16]uint32        // Registers
    CPSR  CPSRFlags         // Status flags
    PC    uint32            // Program counter (alias for R[15])
    Mem   *Memory           // Memory subsystem
    Cycles uint64           // Cycle counter
}

type CPSRFlags struct {
    N bool  // Negative
    Z bool  // Zero
    C bool  // Carry
    V bool  // Overflow
}
```

#### Memory (memory.go)

**Architecture:**
- 4GB address space (32-bit)
- Segmented memory model:
  - Code segment (read-only)
  - Data segment (read-write)
  - Heap segment (dynamic)
  - Stack segment (grows downward)

**Features:**
- Alignment checking
- Permission enforcement
- Bounds checking
- Little-endian byte order

**Key Types:**
```go
type Memory struct {
    Data      map[uint32]byte  // Sparse memory
    Segments  []MemorySegment  // Segment definitions
    PageSize  uint32           // Page size for allocation
}

type MemorySegment struct {
    Start  uint32
    End    uint32
    Name   string
    Perms  Permissions  // Read, Write, Execute
}
```

#### Executor (executor.go)

**Fetch-Decode-Execute Cycle:**
1. Fetch instruction from memory at PC
2. Decode opcode and operands
3. Execute instruction
4. Update PC
5. Increment cycle counter

**Execution Modes:**
- **Run**: Execute until termination or breakpoint
- **Step**: Execute single instruction (step into)
- **Next**: Execute single instruction (step over)
- **Finish**: Execute until function return

#### Flags (flags.go)

**Flag Operations:**
- Calculation helpers for all arithmetic/logical operations
- Condition code evaluation (all 16 ARM condition codes)
- Shift operations (LSL, LSR, ASR, ROR, RRX)

**Key Functions:**
```go
func UpdateNZ(vm *VM, result uint32)
func UpdateCarryAdd(vm *VM, a, b, result uint32)
func UpdateOverflowAdd(vm *VM, a, b, result uint32)
func EvaluateCondition(vm *VM, cond ConditionCode) bool
```

### 2. Parser (parser/)

The parser converts ARM assembly source code into executable form.

#### Lexer (lexer.go)

**Responsibilities:**
- Tokenize source code
- Handle comments (`;`, `//`, `/* */`)
- Recognize keywords, registers, labels, directives
- Track line and column positions for error reporting

**Token Types:**
- Instructions (MOV, ADD, etc.)
- Registers (R0-R15, SP, LR, PC)
- Literals (numbers, strings)
- Operators (`,`, `[`, `]`, `#`, etc.)
- Directives (.org, .word, etc.)

#### Parser (parser.go)

**Two-Pass Assembly:**

**Pass 1: Symbol Collection**
- Collect all labels and their addresses
- Expand macros
- Process .equ/.set directives
- Build symbol table

**Pass 2: Code Generation**
- Resolve label references
- Generate instruction encodings
- Process data directives
- Create relocations

**Key Types:**
```go
type Instruction struct {
    Address    uint32
    Opcode     uint32
    Mnemonic   string
    Operands   []string
    LineNumber int
}

type Directive struct {
    Type   DirectiveType
    Args   []string
    Line   int
}
```

#### Symbol Table (symbols.go)

**Features:**
- Forward reference resolution
- Duplicate detection
- Scope management (global vs. local labels)
- Constant definitions (.equ)

**Symbol Types:**
- Code labels
- Data labels
- Constants (.equ)
- External symbols (.extern)

### 3. Instructions (instructions/)

Each instruction category has its own module.

#### Data Processing (data_processing.go)

**Implemented:**
- Move: MOV, MVN
- Arithmetic: ADD, ADC, SUB, SBC, RSB, RSC
- Logical: AND, ORR, EOR, BIC
- Compare: CMP, CMN, TST, TEQ

**Common Pattern:**
```go
func ExecuteADD(vm *VM, cond, s, rd, rn, op2 uint32) {
    if !EvaluateCondition(vm, cond) {
        return
    }

    result := rn + op2
    vm.R[rd] = result

    if s != 0 {
        UpdateFlags(vm, rn, op2, result)
    }

    vm.Cycles += 1
}
```

#### Memory Access (memory.go, memory_multi.go)

**Addressing Modes:**
- Offset: `[Rn, #offset]`
- Pre-indexed: `[Rn, #offset]!`
- Post-indexed: `[Rn], #offset`
- Register offset: `[Rn, Rm]`
- Scaled: `[Rn, Rm, LSL #n]`

**Load/Store Multiple:**
- Increment After (IA)
- Increment Before (IB)
- Decrement After (DA)
- Decrement Before (DB)

### 4. Debugger (debugger/)

The debugger provides program analysis and control.

#### Architecture

```
User Input → Command Parser → Debugger Core → VM
                                    ↓
                              Breakpoints
                              Watchpoints
                              Expressions
```

#### Breakpoints (breakpoints.go)

**Types:**
- Address breakpoints
- Label breakpoints
- Conditional breakpoints
- Temporary breakpoints

**Data Structure:**
```go
type Breakpoint struct {
    ID        int
    Address   uint32
    Condition string        // Optional
    Enabled   bool
    HitCount  int
    Temporary bool
}
```

#### Watchpoints (watchpoints.go)

**Watch Types:**
- Write watchpoint (break on modification)
- Read watchpoint (break on access)
- Access watchpoint (break on read OR write)

**Targets:**
- Registers
- Memory addresses
- Expressions

#### TUI (tui.go)

**Built with:**
- github.com/rivo/tview - UI components
- github.com/gdamore/tcell - Terminal handling

**Panels:**
- Source view (assembly listing)
- Register view (R0-R15, CPSR)
- Memory view (hex dump)
- Stack view (SP region)
- Disassembly view (decoded instructions)
- Command input (debugger commands)
- Output console (results)
- Breakpoints/Watchpoints list

### 5. Tools (tools/)

#### Linter (lint.go)

**Checks:**
- Syntax errors (via parser)
- Undefined labels
- Duplicate labels
- Unused labels (with exceptions)
- Unreachable code
- Register restrictions (MUL, PC usage)
- Best practices

**Algorithm:**
```
1. Parse program
2. Build symbol table
3. Check each instruction:
   - Validate operands
   - Check register usage
   - Detect unreachable code
4. Check symbols:
   - Find undefined references
   - Find unused definitions
5. Generate report
```

#### Formatter (format.go)

**Formatting:**
- Consistent indentation (tabs vs. spaces)
- Column alignment for:
  - Labels
  - Mnemonics
  - Operands
  - Comments
- Multiple styles (default, compact, expanded)

**Algorithm:**
```
1. Parse program
2. Calculate column widths
3. For each line:
   - Format label
   - Format instruction
   - Align operands
   - Align comments
4. Output formatted code
```

#### Cross-Reference (xref.go)

**Analysis:**
- Symbol definitions
- Symbol uses
- Reference types (call, branch, load, store, data)
- Call graph (function relationships)

**Output:**
```
Symbol: main
  Defined at: line 10
  References:
    line 5:  BL main (call)
    line 15: B main (branch)

Symbol: process
  Defined at: line 20
  References:
    line 12: BL process (call)
```

## Data Flow

### Program Loading

```
Source File (.s)
    ↓
Lexer (tokenize)
    ↓
Parser (pass 1: collect symbols)
    ↓
Parser (pass 2: generate code)
    ↓
VM Memory (load code and data)
    ↓
VM Registers (initialize)
    ↓
Ready for execution
```

### Program Execution

```
VM.Run()
    ↓
While not terminated:
    ↓
Fetch instruction at PC
    ↓
Decode opcode and operands
    ↓
Check breakpoints/watchpoints
    ↓
Execute instruction
    ↓
Update PC
    ↓
Increment cycles
```

### Debugging Session

```
User Command
    ↓
Command Parser
    ↓
Debugger Core
    ↓
VM Control (step, continue, etc.)
    ↓
Update TUI/Display results
```

## Design Patterns

### 1. Strategy Pattern

Used for instruction execution:

```go
type InstructionExecutor interface {
    Execute(vm *VM, encoding uint32) error
}
```

### 2. Observer Pattern

Used for watchpoints:

```go
type Watchpoint struct {
    Expression string
    OnTrigger  func()
}
```

### 3. Command Pattern

Used for debugger commands:

```go
type Command interface {
    Execute(debugger *Debugger, args []string) error
}
```

### 4. Factory Pattern

Used for creating instructions from opcodes:

```go
func InstructionFactory(opcode uint32) Instruction {
    switch opcode {
    case 0b0000: return &DataProcessing{}
    case 0b0001: return &MemoryAccess{}
    // ...
    }
}
```

## Performance Considerations

### Memory Management

- **Sparse arrays** for memory (map-based)
- Only allocate pages as needed
- Segment-based permissions reduce checks

### Instruction Dispatch

- **Direct function calls** (not reflection)
- Minimal allocations in execute loop
- Inline condition checks

### Optimization Opportunities

1. **Instruction caching**: Cache decoded instructions
2. **JIT compilation**: Translate ARM to native code
3. **Profile-guided optimization**: Optimize hot paths
4. **Memory pooling**: Reuse allocations

## Testing Strategy

### Unit Tests

- Each instruction tested individually
- Flag calculations tested exhaustively
- Memory operations tested for alignment, permissions
- Parser tested with valid and invalid inputs

### Integration Tests

- Complete programs tested end-to-end
- Example programs verify correctness
- Regression tests prevent bugs from recurring

### Coverage Goals

- Instructions: 95%+
- VM core: 90%+
- Parser: 85%+
- Overall: 85%+

## Extension Points

### Adding New Instructions

1. Define opcode in `instructions/` module
2. Implement execution function
3. Add to instruction decoder
4. Add tests
5. Update documentation

### Adding System Calls

1. Define syscall number in `vm/syscall.go`
2. Implement handler function
3. Add to syscall dispatcher
4. Document in syscall reference
5. Add tests

### Adding Debugger Commands

1. Add command in `debugger/commands.go`
2. Implement handler function
3. Update help text
4. Add tests

## Dependencies

### Runtime Dependencies

- **tview**: TUI components
- **tcell**: Terminal handling
- **cobra**: CLI framework
- **toml**: Configuration files

### Development Dependencies

- **testify**: Testing assertions
- Go standard library (testing, benchmark)

## Future Enhancements

### Planned Features

1. **JIT Compilation**: Translate ARM to native code for speed
2. **Remote Debugging**: GDB protocol support
3. **Profiling**: Performance analysis tools
4. **Disassembler**: Binary to assembly conversion
5. **ARM3/ARM6**: Extended instruction sets

### Architecture Changes

1. **Plugin System**: Load external instruction sets
2. **Scripting**: Lua/JavaScript for automation
3. **Network**: Remote execution and debugging
4. **Visualization**: Call graphs, memory maps

## See Also

- [Assembly Reference](assembly_reference.md)
- [Debugger Reference](debugger_reference.md)
- [Contributing Guide](contributing.md)
- [API Documentation](api_reference.md)
