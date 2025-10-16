# Debugging Tutorial

This tutorial walks through practical debugging sessions using the ARM2 Emulator's debugger with real example programs. It covers both command-line debug mode (`--debug`) and TUI mode (`--tui`).

## Prerequisites

- Build the emulator: `go build -o arm-emulator`
- Example programs are in the `examples/` directory
- See [Debugger Reference](debugger_reference.md) for complete command documentation

## Tutorial 1: Debugging a Simple Loop (times_table.s)

The times table program reads a number and prints its multiplication table from 1 to 12. We'll use the debugger to understand how it works.

### Starting the Debugger

```bash
# Command-line mode
./arm-emulator --debug examples/times_table.s

# Or TUI mode (recommended for beginners)
./arm-emulator --tui examples/times_table.s
```

### Command-Line Debug Session

```
(debugger) list
```

This shows the source code around the entry point. You'll see the `main` label and initial instructions.

#### Step 1: Set a Breakpoint at Main

```
(debugger) break main
Breakpoint 1 at 0x0000 (main)

(debugger) info breakpoints
Num  Type  Enabled  Address   Location
1    BP    yes      0x0000    main
```

#### Step 2: Run to the Breakpoint

```
(debugger) run
Breakpoint 1 hit at 0x0000 (main)

(debugger) list
  7: main:
  8:     ; Print prompt message
  9:     LDR r0, =prompt_msg
 10:     BL print_string
 11:
 12:     ; Read number from input
 13:     BL read_int
```

#### Step 3: Examine Initial Register State

```
(debugger) info registers
R0:  0x00000000  (0)          R8:  0x00000000  (0)
R1:  0x00000000  (0)          R9:  0x00000000  (0)
R2:  0x00000000  (0)          R10: 0x00000000  (0)
R3:  0x00000000  (0)          R11: 0x00000000  (0)
R4:  0x00000000  (0)          R12: 0x00000000  (0)
R5:  0x00000000  (0)          SP:  0xFFFF0000
R6:  0x00000000  (0)          LR:  0x00000000
R7:  0x00000000  (0)          PC:  0x00000000

CPSR: N=0 Z=0 C=0 V=0
```

All registers start at zero, SP is at the top of the stack.

#### Step 4: Step Through the Prompt

```
(debugger) step
PC: 0x0004
(debugger) print R0
R0 = 0x0060 (96)    ; Address of prompt_msg string

(debugger) x/s R0
"Enter a number (1-12): "
```

The `LDR r0, =prompt_msg` loaded the string address into R0.

#### Step 5: Step Over the Print Function

```
(debugger) next
Enter a number (1-12): _
```

The `BL print_string` call was executed, printing the prompt. Using `next` (not `step`) stepped *over* the function call.

#### Step 6: Set a Breakpoint at the Loop

```
(debugger) break loop
Breakpoint 2 at 0x001C (loop)

(debugger) continue
5
```

After typing "5" and pressing Enter, execution continues to the loop breakpoint.

#### Step 7: Watch the Loop Counter

```
(debugger) print R4
R4 = 0x00000005 (5)     ; The input number

(debugger) print R5
R5 = 0x00000001 (1)     ; Loop counter

(debugger) watch R5
Watchpoint 1: R5 (write)
```

We're watching R5, which is the loop counter (1-12).

#### Step 8: Continue Through Loop Iterations

```
(debugger) continue
5 x 1 = 5

Watchpoint 1 hit: R5 changed from 1 to 2
PC: 0x0048 (after ADD r5, r5, #1)

(debugger) continue
5 x 2 = 10

Watchpoint 1 hit: R5 changed from 2 to 3
```

The watchpoint triggers every time R5 is incremented.

#### Step 9: Use a Conditional Breakpoint

Let's break when we reach the 10th iteration:

```
(debugger) delete 2              # Remove old loop breakpoint
(debugger) break loop if R5 == 10
Breakpoint 3 at 0x001C (loop) if R5 == 10

(debugger) continue
5 x 3 = 15
5 x 4 = 20
...
5 x 9 = 45
5 x 10 = 50

Breakpoint 3 hit at 0x001C (loop)

(debugger) print R5
R5 = 0x0000000A (10)
```

The conditional breakpoint only triggers when R5 equals 10.

#### Step 10: Examine the Multiplication

```
(debugger) print R4
R4 = 0x00000005 (5)     ; Input number

(debugger) print R5
R5 = 0x0000000A (10)    ; Counter

(debugger) step
(debugger) step
(debugger) step

(debugger) print R6
R6 = 0x00000032 (50)    ; Result: 5 * 10 = 50
```

#### Step 11: Let It Finish

```
(debugger) delete 1              # Remove watchpoint
(debugger) delete 3              # Remove conditional breakpoint
(debugger) continue
5 x 11 = 55
5 x 12 = 60

Program exited with code 0
```

### TUI Mode Session

The TUI mode provides a visual interface that makes debugging easier:

```bash
./arm-emulator --tui examples/times_table.s
```

#### TUI Layout

```
┌─────────────────────────────────────────────────────────────┐
│ Source View                                                  │
│   7: main:                                                   │
│   8:     ; Print prompt message                              │
│   9: >   LDR r0, =prompt_msg         <- Current line        │
│  10:     BL print_string                                     │
│  11:                                                         │
├──────────────────────────┬──────────────────────────────────┤
│ Registers                │ Memory View                       │
│ R0:  0x00000000  (0)    │ 0x0000: E3 A0 00 60             │
│ R4:  0x00000000  (0)    │ 0x0004: EB 00 00 10             │
│ R5:  0x00000000  (0)    │                                  │
│ R6:  0x00000000  (0)    │                                  │
│ PC:  0x00000000         │                                  │
│ CPSR: ----              │                                  │
├──────────────────────────┴──────────────────────────────────┤
│ Breakpoints              │ Watchpoints                      │
│ 1: 0x001C (loop)        │                                  │
├─────────────────────────────────────────────────────────────┤
│ Console Output                                              │
│ Debugger started. Type 'help' for commands.                │
└─────────────────────────────────────────────────────────────┘
(debugger) _
```

#### Using TUI Keyboard Shortcuts

1. **F9** at line 19 (loop:) to set a breakpoint
2. **F5** to run the program
3. Enter "5" when prompted
4. **F10** to step over instructions
5. **F11** to step into function calls
6. **Tab** to switch between panels
7. Type commands at the bottom prompt

#### Advantages of TUI Mode

- See registers update in real-time
- Visual indication of current line
- No need to type `info registers` repeatedly
- Easy breakpoint management with F9
- Split-screen view of code, registers, and memory

## Tutorial 2: Debugging a Calculator (calculator.s)

Let's debug a more complex program that evaluates expressions.

```bash
./arm-emulator --debug examples/calculator.s
```

### Understanding the Program Structure

```
(debugger) list
  1: ; Simple calculator
  2: ; Reads two numbers and an operator, performs calculation
  3:
  4: main:
  5:     LDR r0, =prompt1
  6:     BL print_string
```

### Setting Multiple Breakpoints

```
(debugger) break main
(debugger) break read_int
(debugger) break add_op
(debugger) break sub_op
(debugger) break mul_op
(debugger) break div_op

(debugger) info breakpoints
Num  Type  Enabled  Address   Location
1    BP    yes      0x0000    main
2    BP    yes      0x0020    read_int
3    BP    yes      0x0040    add_op
4    BP    yes      0x0050    sub_op
5    BP    yes      0x0060    mul_op
6    BP    yes      0x0070    div_op
```

### Running with Multiple Breakpoints

```
(debugger) run
Breakpoint 1 hit at 0x0000 (main)

(debugger) continue
Enter first number: 10
Breakpoint 2 hit at 0x0020 (read_int)

(debugger) continue
Enter operator (+,-,*,/): +
Enter second number: 25

Breakpoint 3 hit at 0x0040 (add_op)

(debugger) print R4
R4 = 0x0000000A (10)    ; First number

(debugger) print R5
R5 = 0x00000019 (25)    ; Second number

(debugger) step
(debugger) step

(debugger) print R6
R6 = 0x00000023 (35)    ; Result: 10 + 25 = 35
```

### Using the Backtrace

```
(debugger) backtrace
#0  0x0040 in add_op
#1  0x0018 in main
```

This shows the call stack - we're currently in `add_op`, called from `main`.

### Testing Division by Zero

```
(debugger) reset
(debugger) run
Enter first number: 10
Enter operator (+,-,*,/): /
Enter second number: 0

(debugger) break div_op
(debugger) continue
Breakpoint at 0x0070 (div_op)

(debugger) print R5
R5 = 0x00000000 (0)     ; Divisor is zero!

(debugger) step
Error: Division by zero!
```

## Tutorial 3: Debugging Memory Issues (array_sum.s)

This program sums an array of numbers.

```bash
./arm-emulator --debug examples/array_sum.s
```

### Examining Memory

```
(debugger) break main
(debugger) run

(debugger) print array
array = 0x0100

(debugger) x/10d array
0x0100: 5        ; array[0]
0x0104: 10       ; array[1]
0x0108: 15       ; array[2]
0x010C: 20       ; array[3]
0x0110: 25       ; array[4]
0x0114: 30       ; array[5]
0x0118: 35       ; array[6]
0x011C: 40       ; array[7]
0x0120: 45       ; array[8]
0x0124: 50       ; array[9]
```

### Setting a Memory Watchpoint

```
(debugger) watch [0x0108]    ; Watch array[2]
Watchpoint 1: [0x0108] (write)

(debugger) continue
```

If any instruction writes to address 0x0108, the watchpoint will trigger.

### Examining the Sum Loop

```
(debugger) break sum_loop
(debugger) run

(debugger) print R0
R0 = 0x00000000 (0)     ; Accumulator

(debugger) print R1
R1 = 0x00000100 (256)   ; Array pointer

(debugger) print R2
R2 = 0x0000000A (10)    ; Array length

(debugger) step
(debugger) step
(debugger) step

(debugger) print R0
R0 = 0x00000005 (5)     ; Sum after first element

(debugger) print R1
R1 = 0x00000104 (260)   ; Pointer advanced by 4 bytes
```

### Using Expressions

```
(debugger) print [R1]
[R1] = 0x0000000A (10)  ; Value at current pointer

(debugger) print R0 + [R1]
35                      ; Sum + next value

(debugger) print R2 - 1
9                       ; Remaining iterations
```

## Tutorial 4: Debugging Recursion (fibonacci.s)

Fibonacci is a recursive function, perfect for practicing call stack debugging.

```bash
./arm-emulator --debug examples/fibonacci.s
```

### Setting Up

```
(debugger) break fib
Breakpoint 1 at fib

(debugger) run
Enter N: 5

Breakpoint 1 hit at fib
```

### Watching the Call Stack Grow

```
(debugger) backtrace
#0  0x0020 in fib
#1  0x0010 in main

(debugger) print R0
R0 = 0x00000005 (5)     ; fib(5)

(debugger) continue
Breakpoint 1 hit at fib

(debugger) backtrace
#0  0x0020 in fib
#1  0x0028 in fib
#2  0x0010 in main

(debugger) print R0
R0 = 0x00000004 (4)     ; fib(4) - recursive call

(debugger) continue
Breakpoint 1 hit at fib

(debugger) backtrace
#0  0x0020 in fib
#1  0x0028 in fib
#2  0x0028 in fib
#3  0x0010 in main

(debugger) print R0
R0 = 0x00000003 (3)     ; fib(3) - deeper recursion
```

### Examining the Stack

```
(debugger) info stack
SP:   0xFFFEFFE0
Base: 0xFFFF0000
Used: 32 bytes

Stack contents:
0xFFFEFFE0: 0x00000028   ; Return address (fib)
0xFFFEFFE4: 0x00000003   ; Saved R0
0xFFFEFFE8: 0x00000028   ; Return address (fib)
0xFFFEFFEC: 0x00000004   ; Saved R0
0xFFFEFFF0: 0x00000028   ; Return address (fib)
0xFFFEFFF4: 0x00000005   ; Saved R0
```

### Using 'finish' to Return

```
(debugger) finish
Returned to 0x0028

(debugger) print R0
R0 = 0x00000002 (2)     ; fib(3) returned 2

(debugger) backtrace
#0  0x0028 in fib       ; Now back up one level
#1  0x0028 in fib
#2  0x0010 in main
```

## Tutorial 5: Finding a Bug

Let's say you have a buggy program that's producing incorrect output.

### The Bug Hunt Strategy

1. **Identify the symptom**: Wrong output value
2. **Set breakpoints**: At key calculation points
3. **Examine state**: Check registers and memory
4. **Step through**: Execute line by line
5. **Find the cause**: Compare expected vs actual values

### Example: Buggy Average Calculator

```
(debugger) break calculate_average
(debugger) run

(debugger) print R0
R0 = 0x00000064 (100)   ; Sum

(debugger) print R1
R1 = 0x00000005 (5)     ; Count

(debugger) step
(debugger) step
(debugger) step

(debugger) print R2
R2 = 0x00000014 (20)    ; Result: 100 / 5 = 20 ✓

; But the output shows 25! Let's keep investigating...

(debugger) step
(debugger) print R2
R2 = 0x00000019 (25)    ; Aha! Something changed it!

(debugger) list
50:     DIV R2, R0, R1   ; Calculate average
51:     ADD R2, R2, #5   ; ← BUG! Accidentally adding 5
52:     MOV R0, R2
```

Found the bug! There's an extra `ADD R2, R2, #5` that shouldn't be there.

## Tips for Effective Debugging

### 1. Start with TUI Mode

TUI mode is more intuitive for learning:
```bash
./arm-emulator --tui program.s
```

### 2. Use Conditional Breakpoints

Instead of hitting a breakpoint 100 times:
```
(debugger) break loop if R0 >= 100
```

### 3. Combine Watchpoints and Breakpoints

Watch a critical variable and break when conditions are met:
```
(debugger) watch result
(debugger) break check_result if [result] == 0
```

### 4. Use 'next' vs 'step'

- `step` (F11): Goes *into* function calls
- `next` (F10): Steps *over* function calls

Use `next` when you trust the function; use `step` when you need to debug it.

### 5. Save Time with 'finish'

Instead of stepping through an entire function:
```
(debugger) finish
```

### 6. Examine Memory in Different Formats

```
(debugger) x/4x array    ; Hexadecimal
(debugger) x/4d array    ; Decimal
(debugger) x/s string    ; String
(debugger) x/4t flags    ; Binary
```

### 7. Use Expressions

```
(debugger) print (R0 + R1) * 2
(debugger) print [SP + 8]
(debugger) print R0 & 0xFF
```

### 8. Check the Call Stack

When execution is in an unexpected place:
```
(debugger) backtrace
```

### 9. Reset and Retry

Made a mistake? Reset the VM:
```
(debugger) reset
(debugger) run
```

### 10. Combine with Tracing

For complex issues, use tracing alongside debugging:
```bash
./arm-emulator --debug --trace --trace-file debug.log program.s
```

### 11. Use Diagnostic Modes

The emulator provides several diagnostic modes to help analyze program behavior:

```bash
# Code coverage - see which instructions were executed
./arm-emulator --coverage program.s

# Stack trace - monitor stack operations and detect issues
./arm-emulator --stack-trace program.s

# Flag trace - track CPSR flag changes
./arm-emulator --flag-trace program.s

# Register trace - analyze register usage patterns
./arm-emulator --register-trace program.s

# Combine multiple modes
./arm-emulator --coverage --stack-trace --flag-trace --register-trace program.s
```

These modes can help identify:
- Dead code (coverage)
- Stack overflow/underflow (stack trace)
- Incorrect conditional logic (flag trace)
- Uninitialized registers or inefficient register usage (register trace)

## Common Debugging Scenarios

### Scenario 1: Loop Not Terminating

```
(debugger) break loop_start
(debugger) watch loop_counter
(debugger) continue
; Check if loop_counter is being updated correctly
```

### Scenario 2: Wrong Calculation Result

```
(debugger) break before_calculation
(debugger) step
; Step through the calculation line by line
; Print intermediate values
```

### Scenario 3: Unexpected Program Exit

```
(debugger) break main
(debugger) run
(debugger) continue
; Check if program exits normally or crashes
; Examine exit code with: print R0
```

### Scenario 4: Function Not Being Called

```
(debugger) break function_name
(debugger) run
; If breakpoint never hits, check the caller
(debugger) break caller
; Examine the branch/call instruction
```

### Scenario 5: Memory Corruption

```
(debugger) watch [suspicious_address]
(debugger) run
; When watchpoint hits, check what wrote to memory
(debugger) backtrace
```

## Keyboard Shortcuts Reference (TUI Mode)

| Shortcut | Action |
|----------|--------|
| **F1** | Show help |
| **F5** | Continue execution (run/continue) |
| **F9** | Toggle breakpoint at current line |
| **F10** | Step over (next) |
| **F11** | Step into (step) |
| **Ctrl+C** | Stop/interrupt program |
| **Ctrl+L** | Refresh display |
| **Tab** | Switch between panels |
| **↑/↓** | Command history / scroll |
| **PgUp/PgDn** | Scroll active panel |

## Next Steps

- Read the [Debugger Reference](debugger_reference.md) for complete command documentation
- Try debugging the example programs in `examples/`
- Experiment with watchpoints and conditional breakpoints
- Practice using the TUI mode keyboard shortcuts

## See Also

- [Debugger Reference](debugger_reference.md) - Complete command reference
- [Assembly Reference](assembly_reference.md) - ARM2 instruction set
- [Examples](../examples/README.md) - Sample programs to practice with
