# ARM2 Emulator API Reference

This document provides a comprehensive API reference for developers who want to use or extend the ARM2 Emulator programmatically.

## Table of Contents

- [Overview](#overview)
- [Core Packages](#core-packages)
- [VM Package](#vm-package)
- [Parser Package](#parser-package)
- [Debugger Package](#debugger-package)
- [Encoder Package](#encoder-package)
- [Tools Package](#tools-package)
- [Config Package](#config-package)
- [Usage Examples](#usage-examples)

---

## Overview

The ARM2 Emulator is organized into several Go packages, each responsible for a specific aspect of the emulator:

| Package | Purpose |
|---------|---------|
| `vm` | Virtual machine execution, CPU, memory, syscalls |
| `parser` | Assembly language parser, preprocessor, macros |
| `debugger` | Debugging utilities, breakpoints, TUI |
| `encoder` | Machine code encoding/decoding |
| `tools` | Development tools (linter, formatter, xref) |
| `config` | Configuration management |

---

## Core Packages

### Package Import Paths

```go
import (
    "github.com/lookbusy1344/arm-emulator/vm"
    "github.com/lookbusy1344/arm-emulator/parser"
    "github.com/lookbusy1344/arm-emulator/debugger"
    "github.com/lookbusy1344/arm-emulator/encoder"
    "github.com/lookbusy1344/arm-emulator/tools"
    "github.com/lookbusy1344/arm-emulator/config"
)
```

---

## VM Package

The `vm` package provides the virtual machine that executes ARM2 programs.

### CPU

#### Type: CPU

```go
type CPU struct {
    R      [15]uint32  // General purpose registers R0-R14
    PC     uint32      // Program Counter (R15)
    CPSR   CPSR        // Current Program Status Register
    Cycles uint64      // Instruction cycle counter
}
```

**Description**: Represents the ARM2 processor state.

**Register Constants**:
```go
const (
    R0  = 0
    R1  = 1
    R2  = 2
    R3  = 3
    R4  = 4
    R5  = 5
    R6  = 6
    R7  = 7
    R8  = 8
    R9  = 9
    R10 = 10
    R11 = 11
    R12 = 12
    SP  = 13  // Stack Pointer
    LR  = 14  // Link Register
)
```

#### Function: NewCPU

```go
func NewCPU() *CPU
```

**Description**: Creates and initializes a new CPU instance with all registers set to zero.

**Returns**: Pointer to a new CPU.

**Example**:
```go
cpu := vm.NewCPU()
cpu.R[0] = 42
cpu.PC = 0x8000
```

#### Method: GetRegister

```go
func (c *CPU) GetRegister(reg int) uint32
```

**Description**: Returns the value of a register (R0-R14 or PC). When reading R15 (PC), returns PC+8 to simulate ARM pipeline effect.

**Parameters**:
- `reg`: Register number (0-15)

**Returns**: Register value

**Example**:
```go
value := cpu.GetRegister(vm.R0)
pc := cpu.GetRegister(15)  // Returns PC+8
```

#### Method: SetRegister

```go
func (c *CPU) SetRegister(reg int, value uint32)
```

**Description**: Sets the value of a register (R0-R14 or PC).

**Parameters**:
- `reg`: Register number (0-15)
- `value`: Value to set

**Example**:
```go
cpu.SetRegister(vm.R1, 100)
cpu.SetRegister(15, 0x8000)  // Set PC
```

#### Method: Reset

```go
func (c *CPU) Reset()
```

**Description**: Resets the CPU to initial state (all registers zero, cycle count zero).

**Example**:
```go
cpu.Reset()
```

### CPSR (Current Program Status Register)

#### Type: CPSR

```go
type CPSR struct {
    N bool  // Negative flag (bit 31)
    Z bool  // Zero flag (bit 30)
    C bool  // Carry flag (bit 29)
    V bool  // Overflow flag (bit 28)
}
```

**Description**: Represents condition flags.

#### Method: ToUint32

```go
func (c *CPSR) ToUint32() uint32
```

**Description**: Converts CPSR flags to a 32-bit value (NZCV in bits 31-28).

**Returns**: 32-bit representation

**Example**:
```go
cpsr := vm.CPSR{N: true, Z: false, C: true, V: false}
value := cpsr.ToUint32()  // 0xA0000000
```

#### Method: FromUint32

```go
func (c *CPSR) FromUint32(value uint32)
```

**Description**: Sets CPSR flags from a 32-bit value.

**Parameters**:
- `value`: 32-bit value with flags in bits 31-28

**Example**:
```go
cpsr.FromUint32(0x80000000)  // Sets N flag only
```

### Virtual Machine

#### Type: VM

```go
type VM struct {
    CPU    *CPU
    Memory *Memory

    // Execution control
    Running      bool
    MaxCycles    uint64

    // Tracing and diagnostics
    ExecutionTrace *ExecutionTrace
    MemoryTrace    *MemoryTrace
    StackTrace     *StackTrace
    FlagTrace      *FlagTrace
    RegisterTrace  *RegisterTrace
    Coverage       *Coverage
    Statistics     *Statistics

    // Symbol resolution
    SymbolResolver *SymbolResolver
}
```

**Description**: The main virtual machine that executes ARM2 programs.

#### Function: NewVM

```go
func NewVM() *VM
```

**Description**: Creates a new virtual machine with initialized CPU and memory.

**Returns**: Pointer to a new VM.

**Example**:
```go
machine := vm.NewVM()
```

#### Method: LoadProgram

```go
func (vm *VM) LoadProgram(program *parser.Program) error
```

**Description**: Loads a parsed assembly program into memory.

**Parameters**:
- `program`: Parsed program from parser

**Returns**: Error if load fails

**Example**:
```go
p := parser.NewParser(source, "program.s")
program, _ := p.Parse()
err := machine.LoadProgram(program)
```

#### Method: SetEntryPoint

```go
func (vm *VM) SetEntryPoint(address uint32)
```

**Description**: Sets the program entry point (PC).

**Parameters**:
- `address`: Entry point address

**Example**:
```go
machine.SetEntryPoint(0x8000)
```

#### Method: Run

```go
func (vm *VM) Run() error
```

**Description**: Runs the program until completion or error.

**Returns**: Error if execution fails

**Example**:
```go
err := machine.Run()
if err != nil {
    log.Fatal(err)
}
```

#### Method: Step

```go
func (vm *VM) Step() error
```

**Description**: Executes a single instruction.

**Returns**: Error if execution fails

**Example**:
```go
for i := 0; i < 10; i++ {
    if err := machine.Step(); err != nil {
        break
    }
}
```

#### Method: Reset

```go
func (vm *VM) Reset()
```

**Description**: Resets the VM to initial state.

**Example**:
```go
machine.Reset()
```

### Memory

#### Type: Memory

```go
type Memory struct {
    data   []byte
    size   uint32
    origin uint32
}
```

**Description**: Represents the emulator's memory.

#### Function: NewMemory

```go
func NewMemory(size uint32) *Memory
```

**Description**: Creates a new memory instance.

**Parameters**:
- `size`: Memory size in bytes

**Returns**: Pointer to new Memory

**Example**:
```go
mem := vm.NewMemory(0x100000)  // 1MB
```

#### Method: ReadWord

```go
func (m *Memory) ReadWord(address uint32) (uint32, error)
```

**Description**: Reads a 32-bit word from memory.

**Parameters**:
- `address`: Memory address (must be 4-byte aligned)

**Returns**: Word value and error if address invalid

**Example**:
```go
value, err := mem.ReadWord(0x8000)
```

#### Method: WriteWord

```go
func (m *Memory) WriteWord(address uint32, value uint32) error
```

**Description**: Writes a 32-bit word to memory.

**Parameters**:
- `address`: Memory address (must be 4-byte aligned)
- `value`: Value to write

**Returns**: Error if address invalid

**Example**:
```go
err := mem.WriteWord(0x8000, 0xE3A00042)  // MOV R0, #42
```

#### Method: ReadByte

```go
func (m *Memory) ReadByte(address uint32) (byte, error)
```

**Description**: Reads a byte from memory.

**Parameters**:
- `address`: Memory address

**Returns**: Byte value and error if address invalid

**Example**:
```go
b, err := mem.ReadByte(0x8000)
```

#### Method: WriteByte

```go
func (m *Memory) WriteByte(address uint32, value byte) error
```

**Description**: Writes a byte to memory.

**Parameters**:
- `address`: Memory address
- `value`: Byte value

**Returns**: Error if address invalid

**Example**:
```go
err := mem.WriteByte(0x8000, 0x42)
```

### Syscalls

#### Type: SyscallHandler

```go
type SyscallHandler func(vm *VM) error
```

**Description**: Function signature for syscall handlers.

#### Function: RegisterSyscall

```go
func (vm *VM) RegisterSyscall(number uint32, handler SyscallHandler)
```

**Description**: Registers a custom syscall handler.

**Parameters**:
- `number`: Syscall number
- `handler`: Handler function

**Example**:
```go
machine.RegisterSyscall(0x100, func(vm *vm.VM) error {
    value := vm.CPU.R[0]
    fmt.Printf("Custom syscall: %d\n", value)
    return nil
})
```

### Diagnostics

#### Type: ExecutionTrace

```go
type ExecutionTrace struct {
    Enabled   bool
    Entries   []TraceEntry
    Filter    func(entry TraceEntry) bool
}
```

**Description**: Execution trace for debugging.

#### Function: NewExecutionTrace

```go
func NewExecutionTrace() *ExecutionTrace
```

**Description**: Creates a new execution trace.

**Example**:
```go
trace := vm.NewExecutionTrace()
trace.Enabled = true
machine.ExecutionTrace = trace
```

#### Type: Coverage

```go
type Coverage struct {
    Enabled       bool
    ExecutedAddrs map[uint32]*CoverageInfo
}
```

**Description**: Code coverage tracking.

#### Function: NewCoverage

```go
func NewCoverage() *Coverage
```

**Description**: Creates a new coverage tracker.

**Example**:
```go
coverage := vm.NewCoverage()
coverage.Enabled = true
machine.Coverage = coverage
```

---

## Parser Package

The `parser` package parses ARM2 assembly language.

### Parser

#### Type: Parser

```go
type Parser struct {
    // Internal fields
}
```

**Description**: Parses ARM assembly language into a Program.

#### Function: NewParser

```go
func NewParser(input, filename string) *Parser
```

**Description**: Creates a new parser.

**Parameters**:
- `input`: Assembly source code
- `filename`: Source filename (for error messages)

**Returns**: Pointer to new Parser

**Example**:
```go
source := `
    .org 0x8000
_start:
    MOV R0, #42
    SWI #0x00
`
p := parser.NewParser(source, "example.s")
```

#### Method: Parse

```go
func (p *Parser) Parse() (*Program, error)
```

**Description**: Parses the input and returns a Program.

**Returns**: Parsed program and error if parsing fails

**Example**:
```go
program, err := p.Parse()
if err != nil {
    log.Fatal(err)
}
```

### Program

#### Type: Program

```go
type Program struct {
    Instructions    []*Instruction
    Directives      []*Directive
    SymbolTable     *SymbolTable
    MacroTable      *MacroTable
    Origin          uint32
    OriginSet       bool
    LiteralPoolLocs []uint32
}
```

**Description**: Represents a parsed assembly program.

**Example**:
```go
for _, inst := range program.Instructions {
    fmt.Printf("%s %s\n", inst.Mnemonic, strings.Join(inst.Operands, ", "))
}
```

### Instruction

#### Type: Instruction

```go
type Instruction struct {
    Label      string
    Mnemonic   string
    Condition  string
    SetFlags   bool
    Operands   []string
    Comment    string
    Pos        Position
    RawLine    string
    EncodedLen int
    Address    uint32
}
```

**Description**: Represents a parsed instruction.

**Example**:
```go
inst := &parser.Instruction{
    Mnemonic: "MOV",
    Operands: []string{"R0", "#42"},
    Address:  0x8000,
}
```

### Symbol Table

#### Type: SymbolTable

```go
type SymbolTable struct {
    // Internal fields
}
```

**Description**: Manages program symbols (labels, constants).

#### Function: NewSymbolTable

```go
func NewSymbolTable() *SymbolTable
```

**Description**: Creates a new symbol table.

**Example**:
```go
symTable := parser.NewSymbolTable()
```

#### Method: Define

```go
func (st *SymbolTable) Define(name string, value uint32, typ SymbolType) error
```

**Description**: Defines a symbol.

**Parameters**:
- `name`: Symbol name
- `value`: Symbol value (address or constant)
- `typ`: Symbol type (Label, Constant, etc.)

**Returns**: Error if symbol already defined

**Example**:
```go
err := symTable.Define("main", 0x8000, parser.SymbolTypeLabel)
```

#### Method: Lookup

```go
func (st *SymbolTable) Lookup(name string) (*Symbol, bool)
```

**Description**: Looks up a symbol.

**Parameters**:
- `name`: Symbol name

**Returns**: Symbol and true if found

**Example**:
```go
if sym, ok := symTable.Lookup("main"); ok {
    fmt.Printf("main at 0x%08X\n", sym.Value)
}
```

---

## Debugger Package

The `debugger` package provides debugging capabilities.

### Debugger

#### Type: Debugger

```go
type Debugger struct {
    VM          *vm.VM
    Breakpoints *BreakpointManager
    Watchpoints *WatchpointManager
    History     *CommandHistory
    Evaluator   *ExpressionEvaluator
    Running     bool
    StepMode    StepMode
    Symbols     map[string]uint32
    SourceMap   map[uint32]string
}
```

**Description**: Main debugger interface.

#### Function: NewDebugger

```go
func NewDebugger(machine *vm.VM) *Debugger
```

**Description**: Creates a new debugger.

**Parameters**:
- `machine`: VM instance to debug

**Returns**: Pointer to new Debugger

**Example**:
```go
machine := vm.NewVM()
dbg := debugger.NewDebugger(machine)
```

#### Method: LoadSymbols

```go
func (d *Debugger) LoadSymbols(symbols map[string]uint32)
```

**Description**: Loads symbol table for label resolution.

**Parameters**:
- `symbols`: Map of symbol names to addresses

**Example**:
```go
symbols := map[string]uint32{
    "main": 0x8000,
    "loop": 0x8010,
}
dbg.LoadSymbols(symbols)
```

#### Method: Run

```go
func (d *Debugger) Run() error
```

**Description**: Runs the program under debugger control.

**Returns**: Error if execution fails

**Example**:
```go
err := dbg.Run()
```

### Breakpoints

#### Type: BreakpointManager

```go
type BreakpointManager struct {
    // Internal fields
}
```

**Description**: Manages breakpoints.

#### Function: NewBreakpointManager

```go
func NewBreakpointManager() *BreakpointManager
```

**Description**: Creates a new breakpoint manager.

**Example**:
```go
bp := debugger.NewBreakpointManager()
```

#### Method: Set

```go
func (bm *BreakpointManager) Set(address uint32) int
```

**Description**: Sets a breakpoint at an address.

**Parameters**:
- `address`: Breakpoint address

**Returns**: Breakpoint ID

**Example**:
```go
id := bp.Set(0x8000)
```

#### Method: Remove

```go
func (bm *BreakpointManager) Remove(id int) bool
```

**Description**: Removes a breakpoint.

**Parameters**:
- `id`: Breakpoint ID

**Returns**: True if removed

**Example**:
```go
bp.Remove(id)
```

#### Method: IsBreakpoint

```go
func (bm *BreakpointManager) IsBreakpoint(address uint32) bool
```

**Description**: Checks if address has a breakpoint.

**Parameters**:
- `address`: Address to check

**Returns**: True if breakpoint exists

**Example**:
```go
if bp.IsBreakpoint(machine.CPU.PC) {
    fmt.Println("Hit breakpoint!")
}
```

### TUI (Text User Interface)

#### Type: TUI

```go
type TUI struct {
    Debugger *Debugger
    Screen   tcell.Screen
    // Internal fields
}
```

**Description**: Terminal user interface for debugging.

#### Function: NewTUI

```go
func NewTUI(dbg *Debugger) (*TUI, error)
```

**Description**: Creates a new TUI instance.

**Parameters**:
- `dbg`: Debugger instance

**Returns**: Pointer to new TUI and error if initialization fails

**Example**:
```go
tui, err := debugger.NewTUI(dbg)
if err != nil {
    log.Fatal(err)
}
defer tui.Close()
```

#### Method: Run

```go
func (t *TUI) Run() error
```

**Description**: Starts the TUI event loop.

**Returns**: Error if TUI fails

**Example**:
```go
err := tui.Run()
```

---

## Encoder Package

The `encoder` package encodes/decodes ARM machine code.

### Encoder

#### Type: Encoder

```go
type Encoder struct {
    // Internal fields
}
```

**Description**: Encodes ARM instructions to machine code.

#### Function: NewEncoder

```go
func NewEncoder() *Encoder
```

**Description**: Creates a new encoder.

**Example**:
```go
enc := encoder.NewEncoder()
```

#### Method: Encode

```go
func (e *Encoder) Encode(inst *parser.Instruction) (uint32, error)
```

**Description**: Encodes an instruction to 32-bit machine code.

**Parameters**:
- `inst`: Instruction to encode

**Returns**: Encoded instruction and error if encoding fails

**Example**:
```go
inst := &parser.Instruction{
    Mnemonic: "MOV",
    Operands: []string{"R0", "#42"},
}
encoded, err := enc.Encode(inst)
// encoded = 0xE3A0002A
```

### Decoder

#### Type: Decoder

```go
type Decoder struct {
    // Internal fields
}
```

**Description**: Decodes ARM machine code to instructions.

#### Function: NewDecoder

```go
func NewDecoder() *Decoder
```

**Description**: Creates a new decoder.

**Example**:
```go
dec := encoder.NewDecoder()
```

#### Method: Decode

```go
func (d *Decoder) Decode(encoded uint32) (string, error)
```

**Description**: Decodes machine code to assembly string.

**Parameters**:
- `encoded`: 32-bit machine code

**Returns**: Assembly string and error if decoding fails

**Example**:
```go
asm, err := dec.Decode(0xE3A0002A)
// asm = "MOV R0, #42"
```

---

## Tools Package

The `tools` package provides development utilities.

### Linter

#### Type: Linter

```go
type Linter struct {
    // Internal fields
}
```

**Description**: Analyzes assembly code for issues.

#### Function: NewLinter

```go
func NewLinter(program *parser.Program) *Linter
```

**Description**: Creates a new linter.

**Parameters**:
- `program`: Program to analyze

**Returns**: Pointer to new Linter

**Example**:
```go
linter := tools.NewLinter(program)
issues := linter.Analyze()
```

### Formatter

#### Type: Formatter

```go
type Formatter struct {
    Style FormatStyle
}
```

**Description**: Formats assembly code.

#### Function: NewFormatter

```go
func NewFormatter(style FormatStyle) *Formatter
```

**Description**: Creates a new formatter.

**Parameters**:
- `style`: Format style (Default, Compact, Expanded)

**Returns**: Pointer to new Formatter

**Example**:
```go
fmt := tools.NewFormatter(tools.StyleDefault)
formatted := fmt.Format(source)
```

---

## Config Package

The `config` package manages configuration.

### Config

#### Type: Config

```go
type Config struct {
    Emulator EmulatorConfig
    Debugger DebuggerConfig
    Output   OutputConfig
}
```

**Description**: Application configuration.

#### Function: LoadConfig

```go
func LoadConfig() (*Config, error)
```

**Description**: Loads configuration from file.

**Returns**: Config and error if load fails

**Example**:
```go
cfg, err := config.LoadConfig()
if err != nil {
    cfg = config.DefaultConfig()
}
```

#### Function: SaveConfig

```go
func (c *Config) Save() error
```

**Description**: Saves configuration to file.

**Returns**: Error if save fails

**Example**:
```go
err := cfg.Save()
```

---

## Usage Examples

### Complete Example: Load and Run Program

```go
package main

import (
    "fmt"
    "log"

    "github.com/lookbusy1344/arm-emulator/parser"
    "github.com/lookbusy1344/arm-emulator/vm"
)

func main() {
    // Create parser
    source := `
        .org 0x8000
    _start:
        MOV R0, #42
        SWI #0x00
    `
    p := parser.NewParser(source, "example.s")

    // Parse program
    program, err := p.Parse()
    if err != nil {
        log.Fatal(err)
    }

    // Create VM
    machine := vm.NewVM()

    // Load program
    err = machine.LoadProgram(program)
    if err != nil {
        log.Fatal(err)
    }

    // Find and set entry point
    if sym, ok := program.SymbolTable.Lookup("_start"); ok {
        machine.SetEntryPoint(sym.Value)
    }

    // Run
    err = machine.Run()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Exit code: %d\n", machine.CPU.R[0])
}
```

### Example: Debugging with Breakpoints

```go
package main

import (
    "log"

    "github.com/lookbusy1344/arm-emulator/debugger"
    "github.com/lookbusy1344/arm-emulator/parser"
    "github.com/lookbusy1344/arm-emulator/vm"
)

func main() {
    // Parse and load program (as above)
    source := "..." // Your assembly code
    p := parser.NewParser(source, "program.s")
    program, _ := p.Parse()

    machine := vm.NewVM()
    machine.LoadProgram(program)

    // Create debugger
    dbg := debugger.NewDebugger(machine)

    // Load symbols
    symbols := make(map[string]uint32)
    for name, sym := range program.SymbolTable.Symbols {
        symbols[name] = sym.Value
    }
    dbg.LoadSymbols(symbols)

    // Set breakpoint at 'main'
    if addr, ok := symbols["main"]; ok {
        dbg.Breakpoints.Set(addr)
    }

    // Run with debugger
    err := dbg.Run()
    if err != nil {
        log.Fatal(err)
    }
}
```

### Example: Code Coverage Analysis

```go
package main

import (
    "fmt"

    "github.com/lookbusy1344/arm-emulator/parser"
    "github.com/lookbusy1344/arm-emulator/vm"
)

func main() {
    // Parse and load program
    p := parser.NewParser(source, "program.s")
    program, _ := p.Parse()

    machine := vm.NewVM()
    machine.LoadProgram(program)

    // Enable coverage tracking
    coverage := vm.NewCoverage()
    coverage.Enabled = true
    machine.Coverage = coverage

    // Run program
    machine.Run()

    // Analyze coverage
    total := len(program.Instructions)
    executed := len(coverage.ExecutedAddrs)
    percent := float64(executed) / float64(total) * 100

    fmt.Printf("Coverage: %d/%d (%.1f%%)\n", executed, total, percent)

    // Find unexecuted code
    for _, inst := range program.Instructions {
        if _, executed := coverage.ExecutedAddrs[inst.Address]; !executed {
            fmt.Printf("Not executed: 0x%08X: %s\n", inst.Address, inst.RawLine)
        }
    }
}
```

### Example: Custom Syscall

```go
package main

import (
    "fmt"

    "github.com/lookbusy1344/arm-emulator/vm"
)

func main() {
    machine := vm.NewVM()

    // Register custom syscall 0x100
    machine.RegisterSyscall(0x100, func(vm *vm.VM) error {
        // Read parameters from registers
        x := vm.CPU.R[0]
        y := vm.CPU.R[1]

        // Perform operation
        result := x + y

        // Return result in R0
        vm.CPU.R[0] = result

        fmt.Printf("Custom add: %d + %d = %d\n", x, y, result)
        return nil
    })

    // Now assembly can call: SWI #0x100
}
```

---

## Best Practices

### Error Handling

Always check errors from VM and parser operations:

```go
if err := machine.LoadProgram(program); err != nil {
    return fmt.Errorf("failed to load program: %w", err)
}
```

### Memory Management

The VM automatically manages memory. No cleanup needed for normal operation.

### Thread Safety

The VM is **not thread-safe**. Don't access VM state from multiple goroutines simultaneously.

### Performance

For maximum performance:
- Disable tracing and diagnostics when not needed
- Use `Run()` instead of `Step()` for full execution
- Minimize syscall overhead

---

## See Also

- [Tutorial](TUTORIAL.md) - Learn assembly programming
- [Assembly Reference](assembly_reference.md) - Instruction reference
- [Debugger Reference](debugger_reference.md) - Debugging guide
- [FAQ](FAQ.md) - Common questions

---

**API Version**: 1.0
**Last Updated**: October 2025
