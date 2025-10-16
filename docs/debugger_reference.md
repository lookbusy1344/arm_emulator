# Debugger Reference

The ARM2 Emulator includes a powerful built-in debugger with both command-line and text-based user interface (TUI) modes.

## Starting the Debugger

```bash
# Command-line mode
./arm-emulator --debug program.s

# TUI mode (graphical interface)
./arm-emulator --tui program.s
```

## TUI Mode

The TUI (Text User Interface) provides a visual debugging environment with multiple panels:

### Layout

```
┌────────────────────────────────────────────────────────────┐
│ Source View         │ Registers     │ Memory View          │
│                     │ R0:  00000000 │ 0x8000: 00 00 00 00 │
│ 10: _start:         │ R1:  00000000 │ 0x8010: 00 00 00 00 │
│ 11:   MOV R0, #42  │ ...           │ ...                  │
│ 12: > ADD R1, R0    │ PC:  00008004 │                      │
│ 13:   SWI #0x00     │ CPSR: ----    │                      │
├────────────────────────────────────────────────────────────┤
│ Stack View          │ Watchpoints   │ Breakpoints          │
│ SP: 0xFFFF0000      │               │ 1: 0x8000 (enabled)  │
│ [SP]: 00000000      │               │ 2: main (enabled)    │
├────────────────────────────────────────────────────────────┤
│ Command Input                                              │
│ (debugger) step                                            │
├────────────────────────────────────────────────────────────┤
│ Output / Console                                           │
│ Breakpoint 1 at 0x8000 (_start)                           │
│ Stepped to 0x8004                                          │
└────────────────────────────────────────────────────────────┘
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `F1` | Help |
| `F5` | Continue |
| `F9` | Toggle breakpoint at current line |
| `F10` | Step over (next) |
| `F11` | Step into (step) |
| `Ctrl+L` | Refresh display |
| `Ctrl+C` | Quit debugger |
| `Tab` | Switch between panels |
| `↑/↓` | Navigate history / scroll |
| `PgUp/PgDn` | Scroll active panel |

## Command Reference

### Execution Control

#### run / r
Start or restart program execution.

```
(debugger) run
(debugger) r
```

Runs until a breakpoint is hit, program exits, or an error occurs.

#### step / s
Execute one instruction (step into function calls).

```
(debugger) step
(debugger) s
```

#### next / n
Execute one instruction (step over function calls).

```
(debugger) next
(debugger) n
```

#### continue / c
Continue execution from current position.

```
(debugger) continue
(debugger) c
```

#### finish
Run until the current function returns.

```
(debugger) finish
```

#### until <location>
Run until reaching a specific location.

```
(debugger) until loop_end
(debugger) until 0x8020
```

### Breakpoints

#### break / b <location>
Set a breakpoint.

```
(debugger) break _start          # Break at label
(debugger) break 0x8000          # Break at address
(debugger) b main                # Abbreviated form
```

#### break <location> if <condition>
Set a conditional breakpoint.

```
(debugger) break loop if R0 == 10
(debugger) break process if R1 > 100
```

**Supported conditions:**
- Comparisons: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Register values: `R0`, `R1`, ..., `PC`, `SP`, `LR`
- Memory values: `[address]`
- Arithmetic: `+`, `-`, `*`, `/`

#### tbreak <location>
Set a temporary breakpoint (removed after first hit).

```
(debugger) tbreak init
```

#### delete / d <id>
Delete a breakpoint.

```
(debugger) delete 1              # Delete breakpoint 1
(debugger) d 2                   # Abbreviated form
```

#### disable <id>
Disable a breakpoint without deleting it.

```
(debugger) disable 1
```

#### enable <id>
Re-enable a disabled breakpoint.

```
(debugger) enable 1
```

#### info breakpoints / i b
List all breakpoints.

```
(debugger) info breakpoints
(debugger) i b

Output:
Num  Type      Enabled  Address   Location     Condition
1    BP        yes      0x8000    _start
2    BP        yes      0x8010    loop         R0 == 10
3    BP        no       0x8020    done
```

### Watchpoints

#### watch <expression>
Set a write watchpoint (break when memory/register is written).

```
(debugger) watch R0              # Break when R0 changes
(debugger) watch [0x8100]        # Break when memory at 0x8100 changes
(debugger) watch counter         # Break when variable changes
```

#### rwatch <expression>
Set a read watchpoint (break when memory/register is read).

```
(debugger) rwatch [0x8200]
```

#### awatch <expression>
Set an access watchpoint (break on read OR write).

```
(debugger) awatch R5
```

#### info watchpoints
List all watchpoints.

```
(debugger) info watchpoints

Output:
Num  Type      Expression       Condition
1    Write     R0
2    Access    [0x8100]
```

### Inspection

#### print / p <expression>
Evaluate and print an expression.

```
(debugger) print R0              # Print R0
(debugger) print R1 + R2         # Print sum
(debugger) p [0x8000]            # Print memory at address
(debugger) p counter             # Print variable value
```

**Expression syntax:**
- Registers: `R0`, `R1`, ..., `PC`, `SP`, `LR`, `CPSR`
- Memory: `[address]`, `[label]`, `[R0]`
- Symbols: `main`, `counter`, `data`
- Arithmetic: `R0 + 4`, `R1 * 2`, `[SP] - 8`
- Hex: `0x1000`, `0xFF`
- Binary: `0b1010`, `0b11110000`

#### x / examine
Examine memory in various formats.

```
x/<count><format> <address>
```

**Formats:**
- `x` - Hexadecimal
- `d` - Decimal
- `u` - Unsigned decimal
- `o` - Octal
- `t` - Binary
- `c` - Character
- `s` - String

**Examples:**
```
(debugger) x/4x 0x8000           # 4 words in hex
(debugger) x/16b SP              # 16 bytes from SP
(debugger) x/s msg               # String at msg
(debugger) x/1d R0               # R0 as decimal
```

#### info registers / i r
Display all registers.

```
(debugger) info registers
(debugger) i r

Output:
R0:  0x0000002A  (42)         R8:  0x00000000  (0)
R1:  0x00000010  (16)         R9:  0x00000000  (0)
R2:  0x0000003A  (58)         R10: 0x00000000  (0)
R3:  0x00000000  (0)          R11: 0x00000000  (0)
R4:  0x00000000  (0)          R12: 0x00000000  (0)
R5:  0x00000000  (0)          SP:  0xFFFF0000
R6:  0x00000000  (0)          LR:  0x00000000
R7:  0x00000000  (0)          PC:  0x00008004

CPSR: N=0 Z=0 C=0 V=0
Cycles: 42
```

#### info stack / i s
Display stack information.

```
(debugger) info stack
(debugger) i s

Output:
SP:   0xFFFF0000
Base: 0xFFFF0000
Size: 1048576 bytes (1 MB)
Used: 64 bytes

Stack contents:
0xFFFEFFF0: 0x00008010
0xFFFEFFF4: 0x0000002A
0xFFFEFFF8: 0x00000000
0xFFFEFFFC: 0x00000000
```

#### backtrace / bt
Show call stack.

```
(debugger) backtrace
(debugger) bt

Output:
#0  0x8020 in process
#1  0x8010 in calculate
#2  0x8000 in main
```

#### list / l
List source code around current location.

```
(debugger) list                  # List around PC
(debugger) list 20               # List around line 20
(debugger) l _start              # List around label
```

### State Modification

#### set
Set register or memory value.

```
(debugger) set R0 = 42           # Set R0 to 42
(debugger) set PC = 0x8000       # Set PC (dangerous!)
(debugger) set [0x8100] = 100    # Set memory
(debugger) set CPSR.Z = 1        # Set zero flag
```

**CPSR flags:**
```
(debugger) set CPSR.N = 1        # Set negative flag
(debugger) set CPSR.Z = 0        # Clear zero flag
(debugger) set CPSR.C = 1        # Set carry flag
(debugger) set CPSR.V = 0        # Clear overflow flag
```

### Program Control

#### load <file>
Load a new program.

```
(debugger) load program.s
```

#### reset
Reset the VM to initial state.

```
(debugger) reset
```

Resets all registers, clears memory, and reloads the program.

#### quit / q
Exit the debugger.

```
(debugger) quit
(debugger) q
```

### Information Commands

#### help / h
Display help information.

```
(debugger) help                  # General help
(debugger) help break            # Help on specific command
```

#### info / i
Display various information.

```
(debugger) info breakpoints      # List breakpoints
(debugger) info watchpoints      # List watchpoints
(debugger) info registers        # Show registers
(debugger) info stack            # Show stack info
(debugger) info program          # Show program info
```

## Expression Evaluator

The debugger includes a powerful expression evaluator for use in `print`, `watch`, and conditional breakpoint commands.

### Supported Features

#### Literals
```
42                  Decimal
0x2A                Hexadecimal
0b101010            Binary
0o52                Octal
'A'                 Character (ASCII value)
```

#### Registers
```
R0, R1, ..., R15    General purpose registers
PC                  Program counter
SP                  Stack pointer
LR                  Link register
CPSR                Status register
```

#### Memory Access
```
[0x8000]            Dereference address
[R0]                Dereference register
[label]             Dereference label
[SP + 4]            Dereference with offset
```

#### Symbols
```
main                Label address
counter             Variable address
```

#### Operators

**Arithmetic:**
```
+   Addition
-   Subtraction
*   Multiplication
/   Division (integer)
%   Modulo
```

**Bitwise:**
```
&   AND
|   OR
^   XOR
~   NOT
<<  Left shift
>>  Right shift
```

**Comparison:**
```
==  Equal
!=  Not equal
<   Less than
>   Greater than
<=  Less than or equal
>=  Greater than or equal
```

**Logical:**
```
&&  Logical AND
||  Logical OR
!   Logical NOT
```

### Example Expressions

```
R0 + R1                          Sum of registers
[SP] * 2                         Stack top times 2
(R0 + R1) / 2                    Average
R0 & 0xFF                        Mask lower byte
[0x8000] == 42                   Check memory value
R0 > 100 && R1 < 50              Compound condition
```

## Command History

The debugger maintains command history:

- **↑ (Up Arrow)**: Previous command
- **↓ (Down Arrow)**: Next command
- **Ctrl+R**: Search history
- **!!**: Repeat last command

History is saved to `~/.config/arm-emu/history` and persists between sessions.

## Configuration

Create `~/.config/arm-emu/debugger.toml`:

```toml
[debugger]
# Enable TUI by default
default_tui = true

# History settings
history_size = 1000
save_history = true

# Display settings
show_registers = true
show_disassembly = true
assembly_lines = 10

# Colors (for TUI)
[colors]
breakpoint = "red"
current_line = "yellow"
comment = "green"
```

## Tips and Tricks

### 1. Conditional Debugging

Set breakpoints that only trigger under specific conditions:

```
(debugger) break loop if R0 == 100
(debugger) break process if [counter] > threshold
```

### 2. Watchpoints for Data

Use watchpoints to find where data is modified:

```
(debugger) watch [array_base]
(debugger) run
```

### 3. Quick Register Inspection

Use print with multiple registers:

```
(debugger) p R0
(debugger) p R1
(debugger) p R2
```

Or use `info registers` to see all at once.

### 4. Memory Examination

Examine arrays with:

```
(debugger) x/10x array_base      # 10 words in hex
```

### 5. Function Stepping

Use `finish` to quickly exit a function:

```
(debugger) finish                # Run until return
```

### 6. Temporary Breakpoints

Use `tbreak` for one-time breakpoints:

```
(debugger) tbreak init           # Break once at init
```

## Common Workflows

### Debugging a Loop

```
(debugger) break loop            # Set breakpoint at loop start
(debugger) watch R0              # Watch loop counter
(debugger) run                   # Start execution
(debugger) continue              # Continue through iterations
```

### Finding a Bug

```
(debugger) break suspicious_function
(debugger) run
# When breakpoint hits:
(debugger) info registers        # Check register state
(debugger) x/10x SP              # Check stack
(debugger) step                  # Step through code
(debugger) p [data]              # Check data values
```

### Analyzing Memory

```
(debugger) x/16x 0x8000          # Examine code
(debugger) x/s msg               # View string
(debugger) x/10d array           # View array as decimals
```

## Troubleshooting

### Breakpoint Not Hitting

- Check that address/label is correct
- Verify breakpoint is enabled: `info breakpoints`
- Ensure program reaches that location

### Expression Errors

- Use parentheses for complex expressions
- Check register/label names
- Use `0x` prefix for hex numbers

### TUI Display Issues

- Press `Ctrl+L` to refresh
- Check terminal size (minimum 80x24)
- Use `--debug` instead of `--tui` if problems persist

## See Also

- [Assembly Reference](assembly_reference.md) - ARM2 instruction set
- [Debugging Tutorial](debugging_tutorial.md) - Step-by-step debugging guide
- [Examples](../examples/README.md) - Sample programs
