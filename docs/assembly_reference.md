# ARM2 Assembly Language Reference

This document provides a complete reference for the ARM2 assembly language supported by this emulator.

## Table of Contents

- [Program Structure](#program-structure)
- [Registers](#registers)
- [Data Types](#data-types)
- [Addressing Modes](#addressing-modes)
- [Instructions](#instructions)
- [Directives](#directives)
- [Condition Codes](#condition-codes)
- [System Calls](#system-calls)

## Program Structure

### Basic Structure

```asm
        .org    0x8000          ; Set origin address

_start:                         ; Entry point (required)
        ; Your code here
        MOV     R0, #0          ; Exit code
        SWI     #0x00           ; Exit syscall

        ; Data section
msg:    .asciz  "Hello, World!"
data:   .word   42, 100, 255
```

### Entry Points

The emulator searches for entry points in this order:
1. `_start`
2. `main`
3. `__start`
4. `start`
5. First instruction at origin

## Registers

### General Purpose Registers

- **R0-R12**: General purpose registers (32-bit)
- **R13 (SP)**: Stack Pointer
- **R14 (LR)**: Link Register (stores return address)
- **R15 (PC)**: Program Counter

### Status Register

**CPSR** (Current Program Status Register) contains:
- **N**: Negative flag (bit 31)
- **Z**: Zero flag (bit 30)
- **C**: Carry flag (bit 29)
- **V**: Overflow flag (bit 28)

## Data Types

- **Word**: 32-bit (4 bytes) - default
- **Halfword**: 16-bit (2 bytes)
- **Byte**: 8-bit (1 byte)

All multi-byte values are **little-endian**.

## Addressing Modes

### 1. Immediate

Value specified directly in instruction:
```asm
MOV     R0, #42             ; R0 = 42
ADD     R1, R2, #10         ; R1 = R2 + 10
```

### 2. Register

Value from another register:
```asm
MOV     R0, R1              ; R0 = R1
ADD     R2, R3, R4          ; R2 = R3 + R4
```

### 3. Register with Shift

Register value shifted before use:
```asm
MOV     R0, R1, LSL #2      ; R0 = R1 << 2 (multiply by 4)
ADD     R2, R3, R4, LSR #1  ; R2 = R3 + (R4 >> 1)
```

**Shift types:**
- `LSL #n` - Logical Shift Left
- `LSR #n` - Logical Shift Right
- `ASR #n` - Arithmetic Shift Right
- `ROR #n` - Rotate Right
- `RRX` - Rotate Right Extended (through carry)

### 4. Memory Offset

Access memory at base + offset:
```asm
LDR     R0, [R1, #4]        ; R0 = memory[R1 + 4]
STR     R2, [R3, #-8]       ; memory[R3 - 8] = R2
```

### 5. Pre-Indexed

Update base register before access:
```asm
LDR     R0, [R1, #4]!       ; R1 += 4; R0 = memory[R1]
```

### 6. Post-Indexed

Update base register after access:
```asm
LDR     R0, [R1], #4        ; R0 = memory[R1]; R1 += 4
```

### 7. Register Offset

Offset from another register:
```asm
LDR     R0, [R1, R2]        ; R0 = memory[R1 + R2]
STR     R3, [R4, -R5]       ; memory[R4 - R5] = R3
```

### 8. Scaled Register Offset

Register offset with shift:
```asm
LDR     R0, [R1, R2, LSL #2]  ; R0 = memory[R1 + (R2 << 2)]
```

## Instructions

### Data Movement

#### MOV - Move
```asm
MOV{cond}{S}  Rd, <operand>
```
Move value to register.

**Examples:**
```asm
MOV     R0, #42             ; R0 = 42
MOV     R1, R2              ; R1 = R2
MOVEQ   R3, #0              ; R3 = 0 if Z flag set
MOVS    R4, R5              ; R4 = R5 and update flags
```

#### MVN - Move Not
```asm
MVN{cond}{S}  Rd, <operand>
```
Move bitwise NOT of value to register.

**Example:**
```asm
MVN     R0, #0              ; R0 = 0xFFFFFFFF
```

### Arithmetic

#### ADD - Addition
```asm
ADD{cond}{S}  Rd, Rn, <operand>
```

**Examples:**
```asm
ADD     R0, R1, #10         ; R0 = R1 + 10
ADD     R2, R3, R4          ; R2 = R3 + R4
ADDS    R5, R6, R7          ; R5 = R6 + R7 and update flags
```

#### SUB - Subtraction
```asm
SUB{cond}{S}  Rd, Rn, <operand>
```

**Example:**
```asm
SUB     R0, R1, #5          ; R0 = R1 - 5
```

#### RSB - Reverse Subtraction
```asm
RSB{cond}{S}  Rd, Rn, <operand>
```

**Example:**
```asm
RSB     R0, R1, #100        ; R0 = 100 - R1
```

#### ADC - Add with Carry
```asm
ADC{cond}{S}  Rd, Rn, <operand>
```

#### SBC - Subtract with Carry
```asm
SBC{cond}{S}  Rd, Rn, <operand>
```

#### RSC - Reverse Subtract with Carry
```asm
RSC{cond}{S}  Rd, Rn, <operand>
```

### Logical

#### AND - Bitwise AND
```asm
AND{cond}{S}  Rd, Rn, <operand>
```

**Example:**
```asm
AND     R0, R1, #0xFF       ; R0 = R1 & 0xFF (mask lower byte)
```

#### ORR - Bitwise OR
```asm
ORR{cond}{S}  Rd, Rn, <operand>
```

#### EOR - Bitwise Exclusive OR
```asm
EOR{cond}{S}  Rd, Rn, <operand>
```

#### BIC - Bit Clear
```asm
BIC{cond}{S}  Rd, Rn, <operand>
```

**Example:**
```asm
BIC     R0, R1, #0x0F       ; R0 = R1 & ~0x0F (clear lower 4 bits)
```

### Comparison

These instructions update flags but don't store the result.

#### CMP - Compare
```asm
CMP{cond}  Rn, <operand>
```

**Example:**
```asm
CMP     R0, #10             ; Compare R0 with 10
BEQ     equal               ; Branch if equal
```

#### CMN - Compare Negative
```asm
CMN{cond}  Rn, <operand>
```

#### TST - Test Bits
```asm
TST{cond}  Rn, <operand>
```

Performs AND operation and updates flags.

#### TEQ - Test Equivalence
```asm
TEQ{cond}  Rn, <operand>
```

Performs EOR operation and updates flags.

### Memory Access

#### LDR - Load Register
```asm
LDR{cond}  Rd, <address>
```

**Examples:**
```asm
LDR     R0, [R1]            ; R0 = memory[R1]
LDR     R2, [R3, #4]        ; R2 = memory[R3 + 4]
LDR     R4, =label          ; R4 = address of label
```

#### STR - Store Register
```asm
STR{cond}  Rd, <address>
```

**Example:**
```asm
STR     R0, [R1, #8]        ; memory[R1 + 8] = R0
```

#### LDRB - Load Byte
```asm
LDRB{cond}  Rd, <address>
```

Loads a byte and zero-extends to 32 bits.

#### STRB - Store Byte
```asm
STRB{cond}  Rd, <address>
```

Stores lower 8 bits of register.

#### LDRH - Load Halfword
```asm
LDRH{cond}  Rd, <address>
```

#### STRH - Store Halfword
```asm
STRH{cond}  Rd, <address>
```

### Multiple Load/Store

#### LDM - Load Multiple
```asm
LDM{mode}{cond}  Rn{!}, {register-list}
```

**Modes:**
- `IA` - Increment After
- `IB` - Increment Before
- `DA` - Decrement After
- `DB` - Decrement Before

**Stack aliases:**
- `FD` - Full Descending (= LDMIA)
- `ED` - Empty Descending (= LDMDB)
- `FA` - Full Ascending (= LDMDA)
- `EA` - Empty Ascending (= LDMIB)

**Examples:**
```asm
LDMIA   R13!, {R0-R3}       ; Pop R0-R3 from stack
LDMFD   SP!, {R4-R7, PC}    ; Pop R4-R7 and return
```

#### STM - Store Multiple
```asm
STM{mode}{cond}  Rn{!}, {register-list}
```

**Example:**
```asm
STMFD   SP!, {R0-R3, LR}    ; Push R0-R3 and LR onto stack
```

### Branch

#### B - Branch
```asm
B{cond}  label
```

**Examples:**
```asm
B       loop                ; Unconditional branch
BEQ     equal               ; Branch if equal
BNE     not_equal           ; Branch if not equal
```

#### BL - Branch with Link
```asm
BL{cond}  label
```

Calls a subroutine (saves return address in LR).

**Example:**
```asm
BL      function            ; Call function
```

#### BX - Branch and Exchange
```asm
BX{cond}  Rn
```

Branch to address in register.

### Multiply

#### MUL - Multiply
```asm
MUL{cond}{S}  Rd, Rm, Rs
```

**Example:**
```asm
MUL     R0, R1, R2          ; R0 = R1 * R2 (lower 32 bits)
```

**Restrictions:**
- Rd and Rm must be different registers
- Rd cannot be R15 (PC)

#### MLA - Multiply-Accumulate
```asm
MLA{cond}{S}  Rd, Rm, Rs, Rn
```

**Example:**
```asm
MLA     R0, R1, R2, R3      ; R0 = (R1 * R2) + R3
```

### System Call

#### SWI - Software Interrupt
```asm
SWI{cond}  #immediate
```

Invokes a system call. See [System Calls](#system-calls) section.

## Directives

### .org - Set Origin
```asm
.org    address
```

Sets the base address for the program.

**Example:**
```asm
.org    0x8000              ; Program starts at 0x8000
```

### .equ / .set - Define Constant
```asm
.equ    name, value
.set    name, value
```

**Example:**
```asm
.equ    MAX_SIZE, 100
MOV     R0, #MAX_SIZE
```

### .word - Define Word
```asm
.word   value1, value2, ...
```

**Example:**
```asm
data:   .word   10, 20, 30
```

### .half - Define Halfword
```asm
.half   value1, value2, ...
```

### .byte - Define Byte
```asm
.byte   value1, value2, ...
```

### .asciz / .string - Define String
```asm
.asciz  "string"            ; Null-terminated
.string "string"            ; Same as .asciz
```

**Example:**
```asm
msg:    .asciz  "Hello, World!"
```

### .ascii - Define String (no null)
```asm
.ascii  "string"
```

### .space / .skip - Reserve Space
```asm
.space  size
.skip   size
```

**Example:**
```asm
buffer: .space  256         ; Reserve 256 bytes
```

### .align - Align Address
```asm
.align  power_of_2
```

**Example:**
```asm
.align  2                   ; Align to 4-byte boundary
data:   .word   42
```

### .balign - Byte Align
```asm
.balign boundary
```

**Example:**
```asm
.balign 4                   ; Align to 4-byte boundary
```

### .ltorg - Literal Pool
```asm
.ltorg
```

Forces emission of the literal pool at the current location. Literals are values loaded using the `LDR Rd, =constant` pseudo-instruction that cannot be encoded as immediate values.

**Why Use .ltorg:**
- ARM PC-relative addressing has a ±4095 byte range
- Programs with `.org 0x0000` or many constants may exceed this range
- `.ltorg` places literals within reachable distance

**Smart Pool Sizing:**
- Pools are sized dynamically based on actual literal count
- No wasted space for small pools (e.g., 5 literals saves 44 bytes vs. estimate)
- Supports 20+ literals per pool (tested up to 33)
- Address adjustments account for cumulative pool size differences
- Optional validation warnings when `ARM_WARN_POOLS` environment variable is set

**Alignment:**
- Automatically aligns to 4-byte boundary
- Space is reserved for accumulated literals (dynamic calculation)

**Example:**
```asm
.org 0x0000

main:
    LDR R0, =0x12345678     ; Large constant needs literal pool
    LDR R1, =0xDEADBEEF
    LDR R2, =0xCAFEBABE
    ADD R3, R0, R1
    ADD R3, R3, R2
    B   next_section

    .ltorg                  ; Place literals here (within 4095 bytes)
                            ; Pool sized for actual literal count (3 literals)

next_section:
    ; More code far from main...
    LDR R4, =0x11111111
    LDR R5, =0x22222222
    LDR R6, =0x33333333
    LDR R7, =0x44444444
    LDR R8, =0x55555555

    .ltorg                  ; Another pool for distant code
                            ; Pool sized for 5 literals (only 20 bytes needed)
```

**Notes:**
- Multiple `.ltorg` directives allowed
- Literals automatically deduplicated (same value reused)
- Dynamic sizing wastes no space on small pools
- If no `.ltorg` specified, literals placed at end of program
- For programs using `.org 0x8000`, `.ltorg` is usually unnecessary
- Use `ARM_WARN_POOLS=1 ./arm-emulator program.s` to see pool utilization warnings

## Condition Codes

All instructions can be conditionally executed by appending a condition code.

| Code | Meaning | Flags |
|------|---------|-------|
| `EQ` | Equal | Z = 1 |
| `NE` | Not Equal | Z = 0 |
| `CS/HS` | Carry Set / Unsigned Higher or Same | C = 1 |
| `CC/LO` | Carry Clear / Unsigned Lower | C = 0 |
| `MI` | Minus / Negative | N = 1 |
| `PL` | Plus / Positive or Zero | N = 0 |
| `VS` | Overflow Set | V = 1 |
| `VC` | Overflow Clear | V = 0 |
| `HI` | Unsigned Higher | C = 1 AND Z = 0 |
| `LS` | Unsigned Lower or Same | C = 0 OR Z = 1 |
| `GE` | Signed Greater or Equal | N = V |
| `LT` | Signed Less Than | N ≠ V |
| `GT` | Signed Greater Than | Z = 0 AND N = V |
| `LE` | Signed Less or Equal | Z = 1 OR N ≠ V |
| `AL` | Always | (unconditional) |

**Examples:**
```asm
CMP     R0, #10
ADDEQ   R1, R1, #1          ; R1++ if R0 == 10
MOVGT   R2, #5              ; R2 = 5 if R0 > 10 (signed)
BLLT    negative            ; Call if R0 < 10 (signed)
```

## System Calls

System calls are invoked using `SWI #number`.

### Console I/O

| Number | Name | Description | Inputs | Outputs |
|--------|------|-------------|--------|---------|
| 0x00 | EXIT | Exit program | R0 = exit code | - |
| 0x01 | WRITE_CHAR | Write character | R0 = char | - |
| 0x02 | WRITE_STRING | Write string | R0 = string address | - |
| 0x03 | WRITE_INT | Write integer | R0 = value, R1 = base | - |
| 0x04 | READ_CHAR | Read character | - | R0 = char |
| 0x05 | READ_STRING | Read string | R0 = buffer, R1 = max length | R0 = length |
| 0x06 | READ_INT | Read integer | - | R0 = value |
| 0x07 | WRITE_NEWLINE | Write newline | - | - |

### File Operations

| Number | Name | Description | Inputs | Outputs |
|--------|------|-------------|--------|---------|
| 0x10 | OPEN | Open file | R0 = filename, R1 = mode | R0 = file descriptor |
| 0x11 | CLOSE | Close file | R0 = fd | R0 = status |
| 0x12 | READ | Read from file | R0 = fd, R1 = buffer, R2 = count | R0 = bytes read |
| 0x13 | WRITE | Write to file | R0 = fd, R1 = buffer, R2 = count | R0 = bytes written |
| 0x14 | SEEK | Seek in file | R0 = fd, R1 = offset, R2 = whence | R0 = position |
| 0x15 | TELL | Get position | R0 = fd | R0 = position |
| 0x16 | FILE_SIZE | Get file size | R0 = fd | R0 = size |

### Memory Operations

| Number | Name | Description | Inputs | Outputs |
|--------|------|-------------|--------|---------|
| 0x20 | ALLOCATE | Allocate memory | R0 = size | R0 = address |
| 0x21 | FREE | Free memory | R0 = address | - |
| 0x22 | REALLOCATE | Reallocate memory | R0 = address, R1 = new size | R0 = new address |

**Example:**
```asm
; Allocate 100 bytes
MOV     R0, #100
SWI     #0x20               ; ALLOCATE
MOV     R4, R0              ; Save address

; Use the memory...

; Free the memory
MOV     R0, R4
SWI     #0x21               ; FREE
```

## Comments

```asm
; Single line comment

// Also single line comment

/* Multi-line
   comment */

MOV     R0, #10             ; Inline comment
```

## Labels

### Global Labels
Start at column 0:
```asm
main:
loop:
data:
```

### Local Labels
Start with a dot:
```asm
.loop:
.skip:
.done:
```

### Numeric Labels
```asm
1:                          ; Define label 1
        B       1f          ; Branch forward to label 1
1:                          ; Another label 1
        B       1b          ; Branch backward to label 1
```

## Best Practices

1. **Always initialize registers** before use
2. **Use meaningful label names**
3. **Comment your code** - explain the "why", not the "what"
4. **Preserve registers** in functions using PUSH/POP
5. **Check for overflow** when necessary
6. **Align data** for better performance
7. **Free allocated memory** when done
8. **Use constants** (.equ) for magic numbers

## See Also

- [TUTORIAL.md](TUTORIAL.md) - Step-by-step learning guide
- [Instruction Set Reference](INSTRUCTIONS.md) - Detailed documentation for every ARM2 instruction
- [Examples](../examples/README.md) - 44 sample programs
- [Debugger Reference](debugger_reference.md) - Debugging commands and features
- [Debugging Tutorial](debugging_tutorial.md) - Hands-on debugging walkthroughs
- [FAQ](FAQ.md) - Common questions and troubleshooting
