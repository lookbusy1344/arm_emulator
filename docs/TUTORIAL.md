# ARM2 Assembly Programming Tutorial

Welcome to the ARM2 Assembly Programming Tutorial! This guide will teach you ARM2 assembly language from scratch, with hands-on examples you can run in the emulator.

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Your First Program](#your-first-program)
4. [Understanding Registers](#understanding-registers)
5. [Basic Arithmetic](#basic-arithmetic)
6. [Working with Memory](#working-with-memory)
7. [Control Flow](#control-flow)
8. [Functions and the Stack](#functions-and-the-stack)
9. [Arrays and Data Structures](#arrays-and-data-structures)
10. [Advanced Topics](#advanced-topics)
11. [Debugging Your Programs](#debugging-your-programs)
12. [Next Steps](#next-steps)

---

## Introduction

### What is ARM2?

ARM2 is a 32-bit RISC (Reduced Instruction Set Computer) processor introduced in 1986 by Acorn Computers. It's the ancestor of the ARM processors found in modern smartphones, tablets, and laptops.

**Key characteristics:**
- **32-bit architecture**: Registers and addresses are 32 bits wide
- **Load-store architecture**: Only load/store instructions access memory
- **Simple instruction set**: About 30 core instructions
- **Conditional execution**: Every instruction can be conditionally executed
- **Efficient**: Originally achieved 6-10 MIPS while consuming minimal power

### What You'll Learn

By the end of this tutorial, you'll be able to:
- Write ARM2 assembly programs from scratch
- Understand ARM2's register architecture and instruction set
- Work with memory, arrays, and data structures
- Write functions with proper calling conventions
- Use the debugger to find and fix bugs
- Read and understand existing ARM assembly code

### Prerequisites

- Basic programming knowledge (variables, loops, functions)
- Emulator installed and working (see [installation.md](installation.md))
- A text editor for writing assembly code

---

## Getting Started

### Program Structure

Every ARM assembly program has this basic structure:

```asm
        .org    0x8000          ; Set starting address

_start:                         ; Entry point (required)
        ; Your code goes here

        MOV     R0, #0          ; Exit code
        SWI     #0x00           ; Exit syscall

        ; Data section (optional)
data:   .word   42
msg:    .asciz  "Hello"
```

**Key elements:**
1. **`.org` directive**: Sets where your program loads in memory (usually `0x8000`)
2. **`_start` label**: The entry point where execution begins
3. **Exit syscall**: `SWI #0x00` terminates your program
4. **Data section**: Define constants, strings, and variables

### Running Your Programs

Save your code to a file (e.g., `program.s`) and run:

```bash
./arm-emulator program.s
```

To use the debugger:

```bash
./arm-emulator --debug program.s
```

Or TUI mode for visual debugging:

```bash
./arm-emulator --tui program.s
```

---

## Your First Program

Let's write the classic "Hello, World!" program.

### Syscall Register Conventions

When calling system calls (SWI instructions), it's important to understand which registers are preserved and which are modified:

**Preserved Registers (Safe across SWI calls):**
- **R4-R11**: Always preserved - syscalls never modify these
- **SP (R13)**: Stack pointer - always preserved
- **LR (R14)**: Link register - always preserved
- **CPSR flags**: Condition flags (N, Z, C, V) are always preserved

**Volatile Registers (May be modified by SWI calls):**
- **R0**: Used for first parameter and return value - ALWAYS modified
- **R1-R3**: Used for additional parameters - may be modified depending on the syscall
- **R12 (IP)**: Intra-procedure scratch register - may be modified

**Register Usage by Common Syscalls:**

| Syscall | Number | Input | Output | Preserved |
|---------|--------|-------|--------|-----------|
| EXIT | 0x00 | R0=exit_code | - | R1-R14, flags |
| WRITE_CHAR | 0x01 | R0=char | - | R1-R14, flags |
| WRITE_STRING | 0x02 | R0=string_addr | - | R1-R14, flags |
| WRITE_INT | 0x03 | R0=value, R1=base | - | R2-R14, flags |
| READ_CHAR | 0x04 | - | R0=char | R1-R14, flags |
| READ_STRING | 0x05 | R0=buffer, R1=maxlen | R0=bytes_read | R2-R14, flags |
| READ_INT | 0x06 | - | R0=value | R1-R14, flags |
| WRITE_NEWLINE | 0x07 | - | - | R0-R14, flags |

**Important Notes:**
- Always assume R0-R3 may be modified by any syscall
- If you need to preserve R0-R3 across a syscall, save them to the stack or other registers (R4-R11)
- The CPSR flags are always preserved, so you don't need to worry about condition codes being corrupted
- For nested function calls with syscalls, use `PUSH {R4-R11, LR}` to preserve both work registers and return address

### Example: hello.s

Create a file called `hello.s`:

```asm
        .org    0x8000

_start:
        ; Load address of message into R0
        LDR     R0, =msg_hello

        ; Call WRITE_STRING syscall
        ; Note: R0 will be preserved (WRITE_STRING doesn't modify it for this syscall)
        ; but we should not rely on this - always assume R0-R3 may be modified
        SWI     #0x02

        ; Print a newline
        SWI     #0x07

        ; Exit with code 0
        MOV     R0, #0
        SWI     #0x00

        ; Data section
msg_hello:
        .asciz  "Hello, World!"
```

### Running It

```bash
./arm-emulator hello.s
```

**Output:**
```
Hello, World!
```

### Understanding the Code

1. **`LDR R0, =msg_hello`**: Loads the address of `msg_hello` into register R0
   - The `=` syntax tells the assembler to load a 32-bit constant (the address)

2. **`SWI #0x02`**: Software interrupt (syscall) for WRITE_STRING
   - Expects R0 to contain the string address
   - Prints the null-terminated string

3. **`SWI #0x07`**: WRITE_NEWLINE syscall
   - Prints a newline character

4. **`.asciz "Hello, World!"`**: Defines a null-terminated string

### Try It Yourself

Modify the program to:
1. Print your name instead of "Hello, World!"
2. Print multiple lines
3. Print the message twice

**Solution for printing twice:**
```asm
_start:
        LDR     R0, =msg_hello
        SWI     #0x02
        SWI     #0x07

        LDR     R0, =msg_hello  ; Load again
        SWI     #0x02
        SWI     #0x07

        MOV     R0, #0
        SWI     #0x00
```

### Example: Preserving Registers Across Syscalls

Here's an example showing how to preserve register values when making syscalls:

```asm
        .org    0x8000

_start:
        ; Calculate a value we want to keep
        MOV     R4, #42         ; R4 = 42 (safe - R4 preserved by syscalls)
        MOV     R0, #10         ; R0 = 10 (will be modified)

        ; Print first number
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT (modifies R0, preserves R4)

        SWI     #0x07           ; WRITE_NEWLINE

        ; R0 may have been modified, but R4 is still 42
        MOV     R0, R4          ; Move our preserved value to R0
        MOV     R1, #10
        SWI     #0x03           ; Print 42

        SWI     #0x07

        ; Example: When you need to preserve multiple values
        MOV     R0, #100
        MOV     R1, #200
        MOV     R2, #300

        ; Save R0-R2 to preserved registers before syscalls
        MOV     R4, R0          ; R4 = 100
        MOV     R5, R1          ; R5 = 200
        MOV     R6, R2          ; R6 = 300

        ; Now make multiple syscalls
        LDR     R0, =msg_values
        SWI     #0x02           ; WRITE_STRING

        ; Our values are still safe in R4-R6
        MOV     R0, R4
        MOV     R1, #10
        SWI     #0x03           ; Print 100

        ; Exit
        MOV     R0, #0
        SWI     #0x00

msg_values:
        .asciz  "Saved value: "
```

**Key Takeaways:**
- Use R4-R11 for values you need to keep across syscalls
- R0-R3 should be considered temporary when calling syscalls
- The CPSR flags are automatically preserved, so conditional logic remains valid after syscalls

---

## Understanding Registers

### The Register File

ARM2 has 16 registers, each 32 bits wide:

| Register | Alias | Purpose |
|----------|-------|---------|
| R0-R3 | - | General purpose, function arguments/return values |
| R4-R10 | - | General purpose, preserved across calls |
| R11 | FP | Frame Pointer (optional) |
| R12 | IP | Intra-procedure call scratch register |
| R13 | SP | Stack Pointer |
| R14 | LR | Link Register (return address) |
| R15 | PC | Program Counter |

### Special Registers

**CPSR (Current Program Status Register)**
- Contains condition flags and processor status
- **N**: Negative (bit 31) - Result has bit 31 set (negative in signed operations)
- **Z**: Zero (bit 30) - Result is zero
- **C**: Carry (bit 29) - Carry out (addition) or NO borrow (subtraction)
- **V**: Overflow (bit 28) - Signed overflow occurred

### Understanding Processor Flags

The CPSR flags are automatically set by arithmetic and logical operations when you use the `S` suffix (e.g., `ADDS`, `SUBS`). Comparison instructions (`CMP`, `TST`) always set flags.

#### The N Flag (Negative)
- **SET (1)**: Result is negative (bit 31 = 1)
- **CLEAR (0)**: Result is positive or zero (bit 31 = 0)
- **Use with**: Signed comparisons (MI, PL conditions)

```asm
SUBS R0, R1, R2     ; If result negative, N=1
BMI  negative_case  ; Branch if minus (N=1)
BPL  positive_case  ; Branch if plus (N=0)
```

#### The Z Flag (Zero)
- **SET (1)**: Result is exactly zero
- **CLEAR (0)**: Result is non-zero
- **Use with**: Equality tests (EQ, NE conditions)

```asm
CMP  R0, R1         ; Compare R0 with R1
BEQ  equal          ; Branch if equal (Z=1)
BNE  not_equal      ; Branch if not equal (Z=0)
```

#### The C Flag (Carry)
This flag behaves differently for addition vs subtraction:

**For Addition:**
- **SET (1)**: Carry occurred (unsigned overflow)
- **CLEAR (0)**: No carry

```asm
ADDS R0, R1, R2     ; Add with flags
BCS  overflow       ; Branch if carry set
```

**For Subtraction (Important!):**
- **SET (1)**: NO borrow occurred (result >= 0)
- **CLEAR (0)**: Borrow occurred (result < 0)

```asm
; This is counterintuitive but important!
SUBS R0, R1, R2     ; If R1 >= R2, C=1 (no borrow needed)
                    ; If R1 < R2,  C=0 (borrow needed)

CMP  R0, #10        ; Compare R0 with 10
BHS  greater_equal  ; Branch if higher or same (C=1)
BLO  lower          ; Branch if lower (C=0)
```

**Use with**: Unsigned comparisons (HI, LO, HS, LS), multi-precision arithmetic (ADC, SBC)

#### The V Flag (Overflow)
- **SET (1)**: Signed overflow occurred (result out of range for two's complement)
- **CLEAR (0)**: No signed overflow
- **Use with**: Signed overflow detection (VS, VC conditions)

**Overflow examples:**
```asm
; Positive + Positive = Negative (overflow!)
MOV  R0, #0x7FFFFFFF    ; Max positive int
ADDS R0, R0, #1         ; Add 1
                        ; Result: 0x80000000 (negative!)
                        ; V=1 (overflow occurred)

; Negative - Positive = Positive (overflow!)
MOV  R0, #0x80000000    ; Min negative int
SUBS R0, R0, #1         ; Subtract 1
                        ; Result: 0x7FFFFFFF (positive!)
                        ; V=1 (overflow occurred)
```

#### When Flags Are Updated

**With S suffix** (explicit flag update):
```asm
ADDS R0, R1, R2     ; Add and update N, Z, C, V
SUBS R0, R1, R2     ; Subtract and update N, Z, C, V
ANDS R0, R1, R2     ; AND and update N, Z, C (V unchanged)
MOVS R0, R1         ; Move and update N, Z, C (V unchanged)
```

**Without S suffix** (no flag update):
```asm
ADD  R0, R1, R2     ; Add but don't update flags
SUB  R0, R1, R2     ; Subtract but don't update flags
```

**Always update flags** (no S needed):
```asm
CMP  R0, R1         ; Compare (like SUBS, but discard result)
CMN  R0, R1         ; Compare negative (like ADDS)
TST  R0, R1         ; Test bits (like ANDS)
TEQ  R0, R1         ; Test equal (like EORS)
```

#### Practical Flag Usage

**Loop termination:**
```asm
        MOV  R0, #10    ; Counter
loop:
        SUBS R0, R0, #1 ; Decrement and set flags
        BNE  loop       ; Continue if not zero (Z=0)
```

**Range checking:**
```asm
        CMP  R0, #0     ; Check if R0 < 0
        BMI  error      ; Branch if negative
        CMP  R0, #100   ; Check if R0 >= 100
        BHS  error      ; Branch if higher or same
```

**Multi-precision arithmetic:**
```asm
        ; 64-bit addition: R1:R0 = R1:R0 + R3:R2
        ADDS R0, R0, R2 ; Add low words, set carry
        ADC  R1, R1, R3 ; Add high words + carry
```

### Using Registers

```asm
        MOV     R0, #42         ; R0 = 42
        MOV     R1, R0          ; R1 = R0 (R1 = 42)
        ADD     R2, R0, R1      ; R2 = R0 + R1 (R2 = 84)
```

### Example: Simple Calculator

```asm
        .org    0x8000

_start:
        ; Calculate (5 + 3) * 2
        MOV     R0, #5          ; R0 = 5
        MOV     R1, #3          ; R1 = 3
        ADD     R2, R0, R1      ; R2 = 5 + 3 = 8
        MOV     R3, #2          ; R3 = 2
        MUL     R4, R2, R3      ; R4 = 8 * 2 = 16

        ; Print result
        LDR     R0, =msg_result
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, R4          ; Result in R0
        MOV     R1, #10         ; Base 10
        SWI     #0x03           ; WRITE_INT

        SWI     #0x07           ; WRITE_NEWLINE

        MOV     R0, #0
        SWI     #0x00           ; EXIT

msg_result:
        .asciz  "Result: "
```

**Output:** `Result: 16`

---

## Basic Arithmetic

### Arithmetic Instructions

| Instruction | Operation | Example |
|-------------|-----------|---------|
| `ADD Rd, Rn, Op2` | Addition | `ADD R0, R1, #5` → R0 = R1 + 5 |
| `SUB Rd, Rn, Op2` | Subtraction | `SUB R0, R1, R2` → R0 = R1 - R2 |
| `RSB Rd, Rn, Op2` | Reverse Sub | `RSB R0, R1, #10` → R0 = 10 - R1 |
| `MUL Rd, Rm, Rs` | Multiply | `MUL R0, R1, R2` → R0 = R1 * R2 |

### Addition and Subtraction

```asm
        MOV     R0, #15
        MOV     R1, #7
        ADD     R2, R0, R1      ; R2 = 15 + 7 = 22
        SUB     R3, R0, R1      ; R3 = 15 - 7 = 8
```

### Multiplication

```asm
        MOV     R0, #6
        MOV     R1, #7
        MUL     R2, R0, R1      ; R2 = 6 * 7 = 42
```

**Important**: Rd and Rm must be different registers!

```asm
        MUL     R0, R0, R1      ; ERROR: Rd == Rm
        MUL     R0, R1, R1      ; OK: Rd != Rm
```

### Division (Software)

ARM2 has no division instruction. Use repeated subtraction:

```asm
        ; Divide R0 by R1, result in R2, remainder in R0
        MOV     R0, #35         ; Dividend
        MOV     R1, #5          ; Divisor
        MOV     R2, #0          ; Quotient

div_loop:
        CMP     R0, R1          ; Compare dividend with divisor
        BLT     div_done        ; If less, we're done
        SUB     R0, R0, R1      ; Subtract divisor
        ADD     R2, R2, #1      ; Increment quotient
        B       div_loop

div_done:
        ; R2 = 7 (quotient), R0 = 0 (remainder)
```

### Flags and Carry

Add the `S` suffix to update condition flags:

```asm
        MOV     R0, #5
        MOV     R1, #10
        SUBS    R2, R0, R1      ; R2 = -5, sets N flag

        ; Now N=1, Z=0, C=0 (borrow occurred)
```

### Exercise: Temperature Converter

Write a program that converts Celsius to Fahrenheit using: F = (C * 9 / 5) + 32

**Hint:** Calculate C * 9, then divide by 5, then add 32.

---

## Working with Memory

### Memory Addressing

ARM2 uses a **load-store architecture**: only LDR/STR instructions access memory.

### Loading from Memory

```asm
        LDR     R0, [R1]        ; R0 = memory[R1]
        LDR     R0, [R1, #4]    ; R0 = memory[R1 + 4]
        LDR     R0, [R1, R2]    ; R0 = memory[R1 + R2]
```

### Storing to Memory

```asm
        STR     R0, [R1]        ; memory[R1] = R0
        STR     R0, [R1, #8]    ; memory[R1 + 8] = R0
```

### Byte Access

```asm
        LDRB    R0, [R1]        ; Load byte (zero-extended)
        STRB    R0, [R1]        ; Store byte (lower 8 bits)
```

### Pre/Post-Indexed Addressing

**Pre-indexed** (update before access):
```asm
        LDR     R0, [R1, #4]!   ; R1 += 4; R0 = memory[R1]
```

**Post-indexed** (update after access):
```asm
        LDR     R0, [R1], #4    ; R0 = memory[R1]; R1 += 4
```

### Example: Array Sum

```asm
        .org    0x8000

_start:
        LDR     R0, =array      ; R0 = array address
        MOV     R1, #5          ; R1 = count
        MOV     R2, #0          ; R2 = sum

sum_loop:
        CMP     R1, #0
        BEQ     sum_done

        LDR     R3, [R0], #4    ; Load and increment
        ADD     R2, R2, R3      ; Add to sum
        SUB     R1, R1, #1      ; Decrement count
        B       sum_loop

sum_done:
        ; Print result
        LDR     R0, =msg_sum
        SWI     #0x02
        MOV     R0, R2
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07

        MOV     R0, #0
        SWI     #0x00

msg_sum:
        .asciz  "Sum: "

        .align  2
array:
        .word   10, 20, 30, 40, 50
```

**Output:** `Sum: 150`

---

## Control Flow

### Comparison and Branching

```asm
        CMP     R0, #10         ; Compare R0 with 10
        BEQ     equal           ; Branch if equal
        BNE     not_equal       ; Branch if not equal
        BGT     greater         ; Branch if greater (signed)
        BLT     less            ; Branch if less (signed)
```

### Condition Codes

| Code | Meaning | When to Use |
|------|---------|-------------|
| `EQ` | Equal | Z = 1 |
| `NE` | Not Equal | Z = 0 |
| `GT` | Greater Than (signed) | Z = 0 AND N = V |
| `LT` | Less Than (signed) | N ≠ V |
| `GE` | Greater or Equal (signed) | N = V |
| `LE` | Less or Equal (signed) | Z = 1 OR N ≠ V |
| `HI` | Higher (unsigned) | C = 1 AND Z = 0 |
| `LS` | Lower or Same (unsigned) | C = 0 OR Z = 1 |

### If-Else Statement

```asm
        ; if (R0 > 10) { R1 = 1; } else { R1 = 0; }
        CMP     R0, #10
        BLE     else_branch

        ; Then branch
        MOV     R1, #1
        B       endif

else_branch:
        MOV     R1, #0

endif:
        ; Continue...
```

### While Loop

```asm
        ; while (R0 < 10) { R0++; }
        MOV     R0, #0

while_loop:
        CMP     R0, #10
        BGE     while_done

        ADD     R0, R0, #1
        B       while_loop

while_done:
```

### For Loop

```asm
        ; for (i = 0; i < 10; i++) { ... }
        MOV     R0, #0          ; i = 0

for_loop:
        CMP     R0, #10         ; i < 10?
        BGE     for_done

        ; Loop body
        ; ...

        ADD     R0, R0, #1      ; i++
        B       for_loop

for_done:
```

### Example: Fibonacci

```asm
        .org    0x8000

_start:
        MOV     R0, #0          ; F(0) = 0
        MOV     R1, #1          ; F(1) = 1
        MOV     R2, #10         ; Count

fib_loop:
        CMP     R2, #0
        BEQ     fib_done

        ; Print current Fibonacci number
        STMFD   SP!, {R0-R2}
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE
        LDMFD   SP!, {R0-R2}

        ; Calculate next: F(n) = F(n-1) + F(n-2)
        ADD     R3, R0, R1      ; R3 = F(n-2) + F(n-1)
        MOV     R0, R1          ; Shift: R0 = F(n-1)
        MOV     R1, R3          ; Shift: R1 = F(n)

        SUB     R2, R2, #1
        B       fib_loop

fib_done:
        MOV     R0, #0
        SWI     #0x00
```

**Output:**
```
0
1
1
2
3
5
8
13
21
34
```

---

## Functions and the Stack

### Calling Convention

**Argument Passing:**
- First 4 arguments: R0-R3
- Additional arguments: push onto stack

**Return Value:**
- Result in R0

**Preserved Registers:**
- R4-R11, SP, LR must be preserved by callee

### Simple Function

```asm
; Add two numbers
; Input:  R0 = a, R1 = b
; Output: R0 = a + b
add_function:
        ADD     R0, R0, R1
        MOV     PC, LR          ; Return
```

**Calling it:**
```asm
        MOV     R0, #5
        MOV     R1, #7
        BL      add_function    ; R0 = 12
```

### The Stack

The stack grows **downward** (from high to low addresses).

**Stack Pointer (SP/R13):**
- Points to the top of the stack
- Initialized to a high address (e.g., 0x00050000)

### Push and Pop

**Pushing:**
```asm
        STMFD   SP!, {R0-R3, LR}    ; Push R0-R3 and LR
```

**Popping:**
```asm
        LDMFD   SP!, {R0-R3, PC}    ; Pop R0-R3 and return
```

### Function with Stack Usage

```asm
; Factorial function
; Input:  R0 = n
; Output: R0 = n!
factorial:
        STMFD   SP!, {R1-R2, LR}    ; Save registers

        CMP     R0, #1
        BLE     fact_base           ; if n <= 1, return 1

        ; Recursive case: n * factorial(n-1)
        MOV     R1, R0              ; Save n
        SUB     R0, R0, #1          ; n - 1
        BL      factorial           ; Call recursively
        MUL     R0, R1, R0          ; n * factorial(n-1)

        LDMFD   SP!, {R1-R2, PC}    ; Restore and return

fact_base:
        MOV     R0, #1
        LDMFD   SP!, {R1-R2, PC}
```

### Example: Function Call Demo

```asm
        .org    0x8000

_start:
        MOV     R0, #5
        BL      factorial

        ; Print result
        MOV     R4, R0
        LDR     R0, =msg_result
        SWI     #0x02
        MOV     R0, R4
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07

        MOV     R0, #0
        SWI     #0x00

; (factorial function from above)

msg_result:
        .asciz  "Factorial: "
```

---

## Arrays and Data Structures

### Defining Arrays

```asm
        .align  2
array:
        .word   10, 20, 30, 40, 50

byte_array:
        .byte   1, 2, 3, 4, 5

string_array:
        .asciz  "Hello"
        .asciz  "World"
```

### Accessing Array Elements

```asm
        LDR     R0, =array
        LDR     R1, [R0]        ; R1 = array[0] = 10
        LDR     R2, [R0, #4]    ; R2 = array[1] = 20
        LDR     R3, [R0, #8]    ; R3 = array[2] = 30
```

**Using index:**
```asm
        LDR     R0, =array
        MOV     R1, #2          ; index
        LDR     R2, [R0, R1, LSL #2]  ; R2 = array[2]
        ; LSL #2 multiplies index by 4 (word size)
```

### Example: Bubble Sort

```asm
        .org    0x8000

_start:
        LDR     R0, =array
        MOV     R1, #5          ; size
        BL      bubble_sort

        ; Print sorted array
        LDR     R4, =array
        MOV     R5, #5

print_loop:
        CMP     R5, #0
        BEQ     done

        LDR     R0, [R4], #4
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07

        SUB     R5, R5, #1
        B       print_loop

done:
        MOV     R0, #0
        SWI     #0x00

; Bubble sort
; Input:  R0 = array address, R1 = size
bubble_sort:
        STMFD   SP!, {R2-R7, LR}

        MOV     R2, R0          ; R2 = array
        MOV     R3, R1          ; R3 = size

outer_loop:
        CMP     R3, #1
        BLE     sort_done

        MOV     R4, R2          ; R4 = current
        SUB     R5, R3, #1      ; R5 = inner count

inner_loop:
        CMP     R5, #0
        BEQ     outer_next

        LDR     R6, [R4]        ; R6 = current element
        LDR     R7, [R4, #4]    ; R7 = next element

        CMP     R6, R7
        BLE     no_swap

        ; Swap
        STR     R7, [R4]
        STR     R6, [R4, #4]

no_swap:
        ADD     R4, R4, #4      ; Move to next element
        SUB     R5, R5, #1
        B       inner_loop

outer_next:
        SUB     R3, R3, #1
        B       outer_loop

sort_done:
        LDMFD   SP!, {R2-R7, PC}

        .align  2
array:
        .word   64, 34, 25, 12, 22
```

---

## Advanced Topics

### Bitwise Operations

```asm
        AND     R0, R1, #0xFF   ; Mask lower 8 bits
        ORR     R0, R1, R2      ; Bitwise OR
        EOR     R0, R1, R2      ; Bitwise XOR
        BIC     R0, R1, #0x0F   ; Bit clear (R1 & ~0x0F)
        MVN     R0, R1          ; Bitwise NOT
```

### Shifts and Rotates

```asm
        MOV     R0, R1, LSL #2  ; Logical shift left (multiply by 4)
        MOV     R0, R1, LSR #1  ; Logical shift right (divide by 2)
        MOV     R0, R1, ASR #3  ; Arithmetic shift right (signed divide)
        MOV     R0, R1, ROR #8  ; Rotate right
```

### Conditional Execution

Every instruction can be conditional:

```asm
        CMP     R0, #10
        ADDGT   R1, R1, #1      ; Only if R0 > 10
        MOVEQ   R2, #0          ; Only if R0 == 10
        STRNE   R3, [R4]        ; Only if R0 != 10
```

### Example: Bit Counting

```asm
; Count set bits in R0
count_bits:
        STMFD   SP!, {R1-R2, LR}

        MOV     R1, R0          ; R1 = value
        MOV     R2, #0          ; R2 = count

bit_loop:
        CMP     R1, #0
        BEQ     bit_done

        TST     R1, #1          ; Test lowest bit
        ADDNE   R2, R2, #1      ; Increment if set

        MOV     R1, R1, LSR #1  ; Shift right
        B       bit_loop

bit_done:
        MOV     R0, R2
        LDMFD   SP!, {R1-R2, PC}
```

### Character and String Operations

```asm
        ; Character literal
        MOV     R0, #'A'        ; R0 = 65 (ASCII)

        ; Escape sequences
        MOV     R0, #'\n'       ; Newline
        MOV     R0, #'\t'       ; Tab
        MOV     R0, #'\''       ; Single quote
```

**String comparison:**
```asm
strcmp:
        LDRB    R2, [R0], #1
        LDRB    R3, [R1], #1
        CMP     R2, R3
        BNE     not_equal
        CMP     R2, #0
        BNE     strcmp
        ; Equal
        MOV     R0, #0
        MOV     PC, LR
not_equal:
        MOV     R0, #1
        MOV     PC, LR
```

---

## Debugging Your Programs

### Using the Command-Line Debugger

Start with `--debug`:

```bash
./arm-emulator --debug program.s
```

**Essential commands:**
```
run               - Start execution
step              - Execute one instruction
next              - Step over function calls
continue          - Run until breakpoint
break main        - Set breakpoint at 'main'
break 0x8004      - Set breakpoint at address
print R0          - Show register value
print [R1]        - Show memory contents
info registers    - Show all registers
disassemble       - Show disassembly
```

### Using the TUI

Start with `--tui`:

```bash
./arm-emulator --tui program.s
```

**Keyboard shortcuts:**
- `F5` - Continue
- `F9` - Toggle breakpoint
- `F10` - Step over
- `F11` - Step into
- Arrow keys - Navigate source

### Diagnostic Modes

**Instruction coverage tracking:**

Track which ARM instructions in your assembly program were executed:

```bash
./arm-emulator --coverage program.s
```

This shows which instructions ran and which didn't, with execution counts and cycle timing. The output goes to `~/.local/share/arm-emu/logs/coverage.txt` by default.

Example with custom output:
```bash
# Text format (default)
./arm-emulator --coverage --coverage-file report.txt program.s

# JSON format for automation
./arm-emulator --coverage --coverage-format json --coverage-file report.json program.s
```

Coverage reports include:
- Executed vs unexecuted instruction addresses
- Execution counts per instruction
- Coverage percentage
- First and last execution cycle for each instruction
- Symbol names (function/label names) in the output

**Stack monitoring:**
```bash
./arm-emulator --stack-trace program.s
```

**Flag tracking:**
```bash
./arm-emulator --flag-trace program.s
```

**Register analysis:**
```bash
./arm-emulator --register-trace program.s
```

### Common Debugging Techniques

1. **Print debugging**: Use `SWI #0x03` to print values
2. **Breakpoints**: Set at suspicious locations
3. **Step through**: Watch register changes
4. **Memory inspection**: Verify array contents
5. **Stack inspection**: Check for overflow/corruption

### Example Debug Session

```bash
$ ./arm-emulator --debug factorial.s
(debug) break factorial
Breakpoint 1 at factorial
(debug) run
Hit breakpoint 1 at factorial
(debug) info registers
R0 = 0x00000005
...
(debug) step
(debug) print R0
R0 = 0x00000005
(debug) continue
```

---

## Next Steps

Congratulations! You've learned ARM2 assembly programming fundamentals.

### Practice Projects

1. **Calculator**: Implement a simple calculator with +, -, *, /
2. **Binary Search**: Write efficient binary search
3. **String Library**: Create strlen, strcpy, strcat functions
4. **Linked List**: Implement insert, delete, search
5. **Game**: Simple number guessing game

### Further Reading

- [Assembly Reference](assembly_reference.md) - Complete instruction reference
- [Debugger Reference](debugger_reference.md) - Advanced debugging
- [FAQ](FAQ.md) - Common questions and solutions
- [Examples](../examples/README.md) - 30+ example programs

### Resources

- ARM Architecture Reference Manual
- ARM System Developer's Guide
- Example programs in `examples/` directory

### Getting Help

- Check [FAQ.md](FAQ.md) for common issues
- Review example programs for patterns
- Use the debugger to understand behavior
- Experiment and learn by doing!

---

**Happy coding!** Remember: the best way to learn assembly is to write code, make mistakes, debug, and repeat. Don't be afraid to experiment!
