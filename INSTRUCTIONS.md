# ARM2 Instruction Set Reference

This document details the ARM assembly instructions implemented and planned for this ARM2 emulator.

## Status Legend
- ✅ **Implemented** - Fully functional and tested
- ⏸️ **Planned** - Documented but not yet implemented
- 🔧 **Partial** - Partially implemented or has limitations

---

## Table of Contents
1. [Data Processing Instructions](#data-processing-instructions)
2. [Memory Access Instructions](#memory-access-instructions)
3. [Branch Instructions](#branch-instructions)
4. [Multiply Instructions](#multiply-instructions)
5. [System Instructions](#system-instructions)
6. [Assembler Directives](#assembler-directives)
7. [Condition Codes](#condition-codes)
8. [Addressing Modes](#addressing-modes)
9. [Shift Operations](#shift-operations)

---

## Data Processing Instructions

### Arithmetic Operations

#### ADD - Add ✅

**Status:** Implemented

**Syntax:** `ADD{cond}{S} Rd, Rn, <operand2>`

**Description:** Adds two values and stores the result

**Operation:** `Rd = Rn + operand2`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
ADD R0, R1, R2        ; R0 = R1 + R2
ADDS R3, R3, #1       ; R3 = R3 + 1, update flags
ADDEQ R4, R5, R6, LSL #2  ; If equal, R4 = R5 + (R6 << 2)
```

#### ADC - Add with Carry ✅

**Status:** Implemented

**Syntax:** `ADC{cond}{S} Rd, Rn, <operand2>`

**Description:** Adds two values plus the carry flag

**Operation:** `Rd = Rn + operand2 + C`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
ADC R0, R1, R2        ; R0 = R1 + R2 + C
ADCS R3, R3, #0       ; R3 = R3 + C, update flags (for multi-precision)
```

#### SUB - Subtract ✅

**Status:** Implemented

**Syntax:** `SUB{cond}{S} Rd, Rn, <operand2>`

**Description:** Subtracts operand2 from Rn

**Operation:** `Rd = Rn - operand2`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
SUB R0, R1, R2        ; R0 = R1 - R2
SUBS R3, R3, #1       ; R3 = R3 - 1, update flags
```

#### SBC - Subtract with Carry ✅

**Status:** Implemented

**Syntax:** `SBC{cond}{S} Rd, Rn, <operand2>`

**Description:** Subtracts operand2 from Rn with borrow

**Operation:** `Rd = Rn - operand2 - NOT(C)`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
SBC R0, R1, R2        ; R0 = R1 - R2 - NOT(C)
SBCS R3, R3, #0       ; For multi-precision subtraction
```

#### RSB - Reverse Subtract ✅

**Status:** Implemented

**Syntax:** `RSB{cond}{S} Rd, Rn, <operand2>`

**Description:** Subtracts Rn from operand2

**Operation:** `Rd = operand2 - Rn`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
RSB R0, R1, #10       ; R0 = 10 - R1
RSBS R2, R2, #0       ; R2 = -R2 (negate)
```

#### RSC - Reverse Subtract with Carry ✅

**Status:** Implemented

**Syntax:** `RSC{cond}{S} Rd, Rn, <operand2>`

**Description:** Subtracts Rn from operand2 with borrow

**Operation:** `Rd = operand2 - Rn - NOT(C)`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
RSC R0, R1, R2        ; R0 = R2 - R1 - NOT(C)
```

### Logical Operations

#### AND - Logical AND ✅

**Status:** Implemented

**Syntax:** `AND{cond}{S} Rd, Rn, <operand2>`

**Description:** Performs bitwise AND

**Operation:** `Rd = Rn AND operand2`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
AND R0, R1, R2        ; R0 = R1 & R2
ANDS R3, R3, #0xFF    ; R3 = R3 & 0xFF, update flags
```

#### ORR - Logical OR ✅

**Status:** Implemented

**Syntax:** `ORR{cond}{S} Rd, Rn, <operand2>`

**Description:** Performs bitwise OR

**Operation:** `Rd = Rn OR operand2`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
ORR R0, R1, R2        ; R0 = R1 | R2
ORRS R3, R3, #0x80    ; Set bit 7, update flags
```

#### EOR - Logical Exclusive OR ✅

**Status:** Implemented

**Syntax:** `EOR{cond}{S} Rd, Rn, <operand2>`

**Description:** Performs bitwise exclusive OR

**Operation:** `Rd = Rn EOR operand2`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
EOR R0, R1, R2        ; R0 = R1 ^ R2
EORS R3, R3, R3       ; R3 = 0, update flags
```

#### BIC - Bit Clear ✅

**Status:** Implemented

**Syntax:** `BIC{cond}{S} Rd, Rn, <operand2>`

**Description:** Clears bits in Rn specified by operand2

**Operation:** `Rd = Rn AND NOT(operand2)`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
BIC R0, R1, R2        ; R0 = R1 & ~R2
BICS R3, R3, #0x0F    ; Clear lower 4 bits, update flags
```

### Move Operations

#### MOV - Move ✅

**Status:** Implemented

**Syntax:** `MOV{cond}{S} Rd, <operand2>`

**Description:** Moves a value into a register

**Operation:** `Rd = operand2`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
MOV R0, R1            ; R0 = R1
MOV R2, #100          ; R2 = 100
MOV R3, R4, LSL #2    ; R3 = R4 << 2
MOVS PC, LR           ; Return from subroutine with flag restore
```

#### MVN - Move NOT ✅

**Status:** Implemented

**Syntax:** `MVN{cond}{S} Rd, <operand2>`

**Description:** Moves the bitwise complement of a value

**Operation:** `Rd = NOT(operand2)`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
MVN R0, R1            ; R0 = ~R1
MVN R2, #0            ; R2 = 0xFFFFFFFF (-1)
```

### Comparison Operations

#### CMP - Compare ✅

**Status:** Implemented

**Syntax:** `CMP{cond} Rn, <operand2>`

**Description:** Compares two values by subtraction

**Operation:** `Rn - operand2` (result discarded)

**Flags:** Always updates N, Z, C, V

**Example:**
```arm
CMP R0, R1            ; Compare R0 with R1
CMP R2, #0            ; Test if R2 is zero
CMP R3, R4, LSL #1    ; Compare R3 with R4 << 1
```

#### CMN - Compare Negative ✅

**Status:** Implemented

**Syntax:** `CMN{cond} Rn, <operand2>`

**Description:** Compares two values by addition

**Operation:** `Rn + operand2` (result discarded)

**Flags:** Always updates N, Z, C, V

**Example:**
```arm
CMN R0, R1            ; Compare R0 with -R1
CMN R2, #-5           ; Test if R2 is 5
```

#### TST - Test Bits ✅

**Status:** Implemented

**Syntax:** `TST{cond} Rn, <operand2>`

**Description:** Tests bits by AND operation

**Operation:** `Rn AND operand2` (result discarded)

**Flags:** Always updates N, Z, C (V unaffected)

**Example:**
```arm
TST R0, #0x01         ; Test if bit 0 is set
TST R1, R2            ; Test bits in common
```

#### TEQ - Test Equivalence ✅

**Status:** Implemented

**Syntax:** `TEQ{cond} Rn, <operand2>`

**Description:** Tests equality by EOR operation

**Operation:** `Rn EOR operand2` (result discarded)

**Flags:** Always updates N, Z, C (V unaffected)

**Example:**
```arm
TEQ R0, R1            ; Test if R0 equals R1
TEQ R2, #0            ; Test if R2 is zero
```

---

## Memory Access Instructions

### Single Data Transfer

#### LDR - Load Word ✅

**Status:** Implemented

**Syntax:** `LDR{cond} Rd, <addressing_mode>`

**Description:** Loads a 32-bit word from memory

**Operation:** `Rd = Memory[address]`

**Example:**
```arm
LDR R0, [R1]          ; R0 = [R1]
LDR R2, [R3, #4]      ; R2 = [R3 + 4]
LDR R4, [R5], #8      ; R4 = [R5], then R5 = R5 + 8
LDR R6, [R7, R8]      ; R6 = [R7 + R8]
```

#### STR - Store Word ✅

**Status:** Implemented

**Syntax:** `STR{cond} Rd, <addressing_mode>`

**Description:** Stores a 32-bit word to memory

**Operation:** `Memory[address] = Rd`

**Example:**
```arm
STR R0, [R1]          ; [R1] = R0
STR R2, [R3, #-4]     ; [R3 - 4] = R2
STR R4, [R5, R6, LSL #2]!  ; [R5 + (R6 << 2)] = R4, writeback
```

#### LDRB - Load Byte ✅

**Status:** Implemented

**Syntax:** `LDRB{cond} Rd, <addressing_mode>`

**Description:** Loads an 8-bit byte from memory (zero-extended)

**Operation:** `Rd = ZeroExtend(Memory[address])`

**Example:**
```arm
LDRB R0, [R1]         ; R0 = byte at [R1]
LDRB R2, [R3, #1]     ; R2 = byte at [R3 + 1]
```

#### STRB - Store Byte ✅

**Status:** Implemented

**Syntax:** `STRB{cond} Rd, <addressing_mode>`

**Description:** Stores an 8-bit byte to memory

**Operation:** `Memory[address] = Rd[7:0]`

**Example:**
```arm
STRB R0, [R1]         ; [R1] = R0[7:0]
STRB R2, [R3, #10]    ; [R3 + 10] = R2[7:0]
```

#### LDRH - Load Halfword ✅

**Status:** Implemented (ARM2a extension)

**Syntax:** `LDRH{cond} Rd, <addressing_mode>`

**Description:** Loads a 16-bit halfword from memory (zero-extended)

**Operation:** `Rd = ZeroExtend(Memory[address])`

**Example:**
```arm
LDRH R0, [R1]         ; R0 = halfword at [R1]
LDRH R2, [R3, #2]     ; R2 = halfword at [R3 + 2]
```

#### STRH - Store Halfword ✅

**Status:** Implemented (ARM2a extension)

**Syntax:** `STRH{cond} Rd, <addressing_mode>`

**Description:** Stores a 16-bit halfword to memory

**Operation:** `Memory[address] = Rd[15:0]`

**Example:**
```arm
STRH R0, [R1]         ; [R1] = R0[15:0]
STRH R2, [R3, #6]     ; [R3 + 6] = R2[15:0]
```

### Multiple Data Transfer

#### LDM - Load Multiple ✅

**Status:** Implemented

**Syntax:** `LDM{cond}{mode} Rn{!}, {register_list}`

**Description:** Loads multiple registers from consecutive memory locations

**Modes:** IA (Increment After), IB (Increment Before), DA (Decrement After), DB (Decrement Before)

**Stack Modes:** FD (Full Descending), ED (Empty Descending), FA (Full Ascending), EA (Empty Ascending)

**Example:**
```arm
LDMIA R13!, {R0-R3}   ; Load R0-R3 from stack, increment R13
LDMFD SP!, {R4-R6, PC}  ; Pop R4-R6 and return
```

#### STM - Store Multiple ✅

**Status:** Implemented

**Syntax:** `STM{cond}{mode} Rn{!}, {register_list}`

**Description:** Stores multiple registers to consecutive memory locations

**Modes:** IA (Increment After), IB (Increment Before), DA (Decrement After), DB (Decrement Before)

**Stack Modes:** FD (Full Descending), ED (Empty Descending), FA (Full Ascending), EA (Empty Ascending)

**Example:**
```arm
STMDB SP!, {R0-R3, LR}  ; Push R0-R3 and LR to stack
STMFD SP!, {R4-R11}   ; Push R4-R11 to stack
```

---

## Branch Instructions

#### B - Branch ✅

**Status:** Implemented

**Syntax:** `B{cond} label`

**Description:** Branches to a label/address

**Operation:** `PC = PC + offset`

**Range:** ±32MB from current instruction

**Example:**
```arm
B loop                ; Branch to loop
BEQ equal_case        ; Branch if equal
BNE not_zero          ; Branch if not zero
```

#### BL - Branch with Link ✅

**Status:** Implemented

**Syntax:** `BL{cond} label`

**Description:** Branches to a subroutine and saves return address

**Operation:** `LR = PC + 4, PC = PC + offset`

**Range:** ±32MB from current instruction

**Example:**
```arm
BL function           ; Call function
BLEQ conditional_fn   ; Call if equal
```

#### BX - Branch and Exchange ✅

**Status:** Implemented

**Syntax:** `BX{cond} Rm`

**Description:** Branches to address in register (ARM/Thumb interworking)

**Operation:** `PC = Rm & 0xFFFFFFFE` (bit 0 would indicate Thumb mode in later ARM)

**Example:**
```arm
BX LR                 ; Return from subroutine
BX R0                 ; Branch to address in R0
```

---

## Multiply Instructions

#### MUL - Multiply ✅

**Status:** Implemented

**Syntax:** `MUL{cond}{S} Rd, Rm, Rs`

**Description:** Multiplies two 32-bit values (lower 32 bits of result)

**Operation:** `Rd = (Rm * Rs)[31:0]`

**Flags:** Updates N, Z when S bit is set (C meaningless, V unaffected)

**Restrictions:** Rd and Rm must be different registers, R15 (PC) cannot be used

**Cycles:** 2-16 cycles depending on multiplier value

**Example:**
```arm
MUL R0, R1, R2        ; R0 = R1 * R2
MULS R3, R4, R5       ; R3 = R4 * R5, update flags
```

#### MLA - Multiply-Accumulate ✅

**Status:** Implemented

**Syntax:** `MLA{cond}{S} Rd, Rm, Rs, Rn`

**Description:** Multiplies and adds to accumulator

**Operation:** `Rd = (Rm * Rs + Rn)[31:0]`

**Flags:** Updates N, Z when S bit is set (C meaningless, V unaffected)

**Restrictions:** Rd and Rm must be different registers, R15 (PC) cannot be used

**Cycles:** 2-16 cycles depending on multiplier value

**Example:**
```arm
MLA R0, R1, R2, R3    ; R0 = R1 * R2 + R3
MLAS R4, R5, R6, R7   ; R4 = R5 * R6 + R7, update flags
```

#### UMULL - Unsigned Multiply Long ⏸️

**Status:** Planned

**Syntax:** `UMULL{cond}{S} RdLo, RdHi, Rm, Rs`

**Description:** Multiplies two 32-bit unsigned values producing 64-bit result

**Operation:** `RdHi:RdLo = Rm * Rs`

#### UMLAL - Unsigned Multiply-Accumulate Long ⏸️

**Status:** Planned

**Syntax:** `UMLAL{cond}{S} RdLo, RdHi, Rm, Rs`

**Description:** Unsigned multiply and accumulate with 64-bit result

**Operation:** `RdHi:RdLo = (Rm * Rs) + RdHi:RdLo`

#### SMULL - Signed Multiply Long ⏸️

**Status:** Planned

**Syntax:** `SMULL{cond}{S} RdLo, RdHi, Rm, Rs`

**Description:** Multiplies two 32-bit signed values producing 64-bit result

**Operation:** `RdHi:RdLo = Rm * Rs (signed)`

#### SMLAL - Signed Multiply-Accumulate Long ⏸️

**Status:** Planned

**Syntax:** `SMLAL{cond}{S} RdLo, RdHi, Rm, Rs`

**Description:** Signed multiply and accumulate with 64-bit result

**Operation:** `RdHi:RdLo = (Rm * Rs) + RdHi:RdLo (signed)`

---

## System Instructions

### SWI - Software Interrupt ✅

**Status:** Implemented

**Syntax:** `SWI{cond} #immediate`

**Description:** Generates a software interrupt (system call)

**Operation:** Transfers control to system call handler

**Example:**
```arm
SWI #0x00             ; Exit program
SWI #0x02             ; Write string to console
SWI #0x11             ; Write character
```

#### System Call Numbers (SWI)

##### Console I/O (0x00-0x07) ✅
- `0x00` - **EXIT** - Exit program (R0 = exit code) ✅
- `0x01` - **WRITE_CHAR** - Write character (R0 = character) ✅
- `0x02` - **WRITE_STRING** - Write null-terminated string (R0 = address) ✅
- `0x03` - **WRITE_INT** - Write integer (R0 = value, R1 = base) ✅
- `0x04` - **READ_CHAR** - Read character (returns in R0) ✅
- `0x05` - **READ_STRING** - Read string (R0 = buffer, R1 = max length) ✅
- `0x06` - **READ_INT** - Read integer (returns in R0) ✅
- `0x07` - **WRITE_NEWLINE** - Write newline ✅

##### File Operations (0x10-0x16) ✅
- `0x10` - **OPEN** - Open file (R0 = filename, R1 = mode) ✅
- `0x11` - **CLOSE** - Close file (R0 = file handle) ✅
- `0x12` - **READ** - Read from file (R0 = handle, R1 = buffer, R2 = size) ✅
- `0x13` - **WRITE** - Write to file (R0 = handle, R1 = buffer, R2 = size) ✅
- `0x14` - **SEEK** - Seek in file (R0 = handle, R1 = offset, R2 = whence) ✅
- `0x15` - **TELL** - Get file position (R0 = handle) ✅
- `0x16` - **FILE_SIZE** - Get file size (R0 = handle) ✅

##### Memory Operations (0x20-0x22) ✅
- `0x20` - **ALLOCATE** - Allocate memory (R0 = size, returns address in R0) ✅
- `0x21` - **FREE** - Free memory (R0 = address) ✅
- `0x22` - **REALLOCATE** - Reallocate memory (R0 = address, R1 = new size) ✅

##### System Information (0x30-0x33) ✅
- `0x30` - **GET_TIME** - Get time in milliseconds (returns in R0) ✅
- `0x31` - **GET_RANDOM** - Get random 32-bit value (returns in R0) ✅
- `0x32` - **GET_ARGUMENTS** - Get command-line arguments ✅
- `0x33` - **GET_ENVIRONMENT** - Get environment variables ✅

##### Error Handling (0x40-0x42) ✅
- `0x40` - **GET_ERROR** - Get last error code ✅
- `0x41` - **SET_ERROR** - Set error code ✅
- `0x42` - **PRINT_ERROR** - Print error message ✅

##### Debugging Support (0xF0-0xF4) ✅
- `0xF0` - **DEBUG_PRINT** - Print debug message (R0 = string address) ✅
- `0xF1` - **BREAKPOINT** - Trigger breakpoint ✅
- `0xF2` - **DUMP_REGISTERS** - Dump all registers ✅
- `0xF3` - **DUMP_MEMORY** - Dump memory region (R0 = address, R1 = length) ✅
- `0xF4` - **ASSERT** - Assert condition (R0 = condition) ✅

### MRS - Move PSR to Register ⏸️

**Status:** Planned

**Syntax:** `MRS{cond} Rd, PSR`

**Description:** Moves CPSR or SPSR to a register

**Example:**
```arm
MRS R0, CPSR          ; R0 = CPSR
```

### MSR - Move Register to PSR ⏸️

**Status:** Planned

**Syntax:** `MSR{cond} PSR, Rm`

**Description:** Moves a register or immediate to CPSR or SPSR

**Example:**
```arm
MSR CPSR, R0          ; CPSR = R0
```

---

## Assembler Directives

Assembler directives control how the assembler processes your source code. They don't generate instructions but affect memory layout, symbol definitions, and code organization.

### Section Directives ✅

#### .text
**Description:** Marks the beginning of the code section

**Syntax:** `.text`

**Example:**
```arm
.text
.global _start
_start:
    MOV R0, #1
```

#### .data
**Description:** Marks the beginning of the data section for initialized data

**Syntax:** `.data`

**Example:**
```arm
.data
counter:    .word 0
message:    .asciz "Hello"
```

**Note:** In this emulator, `.data` and `.text` sections can be interleaved. The assembler tracks addresses sequentially regardless of section.

### Symbol Directives ✅

#### .global
**Description:** Declares a symbol as global (visible to other modules)

**Syntax:** `.global symbol_name`

**Example:**
```arm
.global _start
.global my_function
```

#### .equ / .set
**Description:** Defines a constant value for a symbol

**Syntax:** `.equ symbol, value` or `.set symbol, value`

**Example:**
```arm
.equ BUFFER_SIZE, 256
.equ MAX_COUNT, 100
.set STACK_SIZE, 0x1000
```

### Memory Allocation Directives ✅

#### .org
**Description:** Sets the assembly origin address

**Syntax:** `.org address`

**Example:**
```arm
.org 0x8000        ; Start code at address 0x8000
```

#### .word
**Description:** Allocates and initializes 32-bit words (4 bytes each)

**Syntax:** `.word value1, value2, ...`

**Example:**
```arm
array:      .word 10, 20, 30, 40
table:      .word 0x12345678, 0xABCDEF00
```

#### .half
**Description:** Allocates and initializes 16-bit halfwords (2 bytes each)

**Syntax:** `.half value1, value2, ...`

**Example:**
```arm
shorts:     .half 100, 200, 300
```

#### .byte
**Description:** Allocates and initializes 8-bit bytes (1 byte each)

**Syntax:** `.byte value1, value2, ...`

**Example:**
```arm
bytes:      .byte 0x01, 0x02, 0x03, 0xFF
flags:      .byte 'A', 'B', 'C'
```

#### .ascii
**Description:** Allocates a string without null terminator

**Syntax:** `.ascii "string"`

**Example:**
```arm
msg:        .ascii "Hello"        ; 5 bytes, no null
```

#### .asciz / .string
**Description:** Allocates a null-terminated string (C-style)

**Syntax:** `.asciz "string"` or `.string "string"`

**Example:**
```arm
msg:        .asciz "Hello"        ; 6 bytes, includes null
prompt:     .string "Enter name: "
```

**Escape Sequences:** Both `.ascii` and `.asciz` support standard escape sequences:
- `\n` - Newline (0x0A)
- `\r` - Carriage return (0x0D)
- `\t` - Tab (0x09)
- `\b` - Backspace (0x08)
- `\\` - Backslash
- `\"` - Double quote
- `\'` - Single quote
- `\0` - Null character

**Example:**
```arm
greeting:   .asciz "Hello\nWorld\n"
path:       .asciz "C:\\Users\\Name"
```

#### .space / .skip
**Description:** Reserves specified number of bytes (initialized to zero)

**Syntax:** `.space size` or `.skip size`

**Example:**
```arm
buffer:     .space 256        ; Reserve 256 bytes
stack:      .skip 0x1000      ; Reserve 4KB
```

### Character Literals ✅

Character literals can be used anywhere an immediate value is expected. They are enclosed in single quotes and evaluate to the ASCII/Unicode value of the character.

**Syntax:** `'c'` where c is any character

**Supported in:**
- Immediate operands in data processing instructions
- `.byte` directive values
- `.equ` constant definitions
- Comparison values

**Example:**
```arm
MOV R0, #'A'           ; R0 = 65 (ASCII value of 'A')
CMP R1, #'0'           ; Compare R1 with 48 (ASCII '0')
SUB R2, R2, #' '       ; Subtract space character (32)

.equ NEWLINE, '\n'     ; Define constant from character
.byte 'H', 'i', 0      ; Byte array with characters
```

**Escape Sequences:** Character literals support the same escape sequences as strings:
```arm
MOV R0, #'\n'          ; Newline (10)
MOV R1, #'\t'          ; Tab (9)
MOV R2, #'\\'          ; Backslash (92)
MOV R3, #'\''          ; Single quote (39)
```

### Alignment Directives ✅

#### .align
**Description:** Aligns to 2^n bytes boundary

**Syntax:** `.align n`

**Example:**
```arm
.align 2          ; Align to 4-byte boundary (2^2)
.align 3          ; Align to 8-byte boundary (2^3)
```

#### .balign
**Description:** Aligns to specified byte boundary

**Syntax:** `.balign boundary`

**Example:**
```arm
.balign 4         ; Align to 4-byte boundary
.balign 16        ; Align to 16-byte boundary
```

### Directive Usage Examples

**Complete program structure:**
```arm
; Constants and symbols
.equ BUFFER_SIZE, 256
.equ EXIT_SYSCALL, 0x00

; Code section
.text
.org 0x8000
.global _start

_start:
    .align 4
    LDR R0, =message
    BL print_string
    SWI #EXIT_SYSCALL

print_string:
    ; Function implementation
    MOV PC, LR

; Data section
.data
.align 4
message:    .asciz "Hello, World!\n"
counter:    .word 0
array:      .word 1, 2, 3, 4, 5

; Uninitialized data
buffer:     .space BUFFER_SIZE
temp:       .word 0
```

**Mixed code and data:**
```arm
.text
function:
    MOV R0, R1
    MOV PC, LR

.data
value:      .word 42

.text
another_fn:
    LDR R0, =value
    LDR R0, [R0]
    MOV PC, LR
```

---

## Condition Codes

All instructions can be conditionally executed based on CPSR flags. ✅

| Code | Suffix | Description | Condition |
|------|--------|-------------|-----------|
| 0000 | EQ | Equal | Z == 1 |
| 0001 | NE | Not Equal | Z == 0 |
| 0010 | CS/HS | Carry Set / Unsigned Higher or Same | C == 1 |
| 0011 | CC/LO | Carry Clear / Unsigned Lower | C == 0 |
| 0100 | MI | Minus / Negative | N == 1 |
| 0101 | PL | Plus / Positive or Zero | N == 0 |
| 0110 | VS | Overflow Set | V == 1 |
| 0111 | VC | Overflow Clear | V == 0 |
| 1000 | HI | Unsigned Higher | C == 1 AND Z == 0 |
| 1001 | LS | Unsigned Lower or Same | C == 0 OR Z == 1 |
| 1010 | GE | Signed Greater or Equal | N == V |
| 1011 | LT | Signed Less Than | N != V |
| 1100 | GT | Signed Greater Than | Z == 0 AND N == V |
| 1101 | LE | Signed Less or Equal | Z == 1 OR N != V |
| 1110 | AL | Always (default) | Always true |
| 1111 | NV | Never (deprecated) | Never true |

**Examples:**
```arm
ADDEQ R0, R1, R2      ; Add if equal (Z == 1)
STRNE R3, [R4]        ; Store if not equal (Z == 0)
BLLT function         ; Call if less than (N != V)
MOVGT R5, #1          ; Move if greater than
```

---

## Addressing Modes

### Data Processing Operand2 ✅

1. **Immediate Value with Rotation**
   ```arm
   MOV R0, #255          ; R0 = 255
   ADD R1, R2, #0x100    ; R1 = R2 + 256
   ```

2. **Register**
   ```arm
   ADD R0, R1, R2        ; R0 = R1 + R2
   ```

3. **Register with Logical Shift Left (LSL)**
   ```arm
   ADD R0, R1, R2, LSL #2    ; R0 = R1 + (R2 << 2)
   ```

4. **Register with Logical Shift Right (LSR)**
   ```arm
   SUB R0, R1, R2, LSR #4    ; R0 = R1 - (R2 >> 4)
   ```

5. **Register with Arithmetic Shift Right (ASR)**
   ```arm
   MOV R0, R1, ASR #8        ; R0 = R1 >> 8 (signed)
   ```

6. **Register with Rotate Right (ROR)**
   ```arm
   ORR R0, R1, R2, ROR #16   ; R0 = R1 | rotate_right(R2, 16)
   ```

7. **Register with Rotate Right Extended (RRX)**
   ```arm
   MOV R0, R1, RRX           ; R0 = rotate_right_with_carry(R1)
   ```

8. **Register-specified Shift**
   ```arm
   MOV R0, R1, LSL R2        ; R0 = R1 << R2
   ADD R3, R4, R5, LSR R6    ; R3 = R4 + (R5 >> R6)
   ```

### Memory Addressing Modes ✅

1. **Register Indirect**
   ```arm
   LDR R0, [R1]              ; R0 = [R1]
   ```

2. **Pre-indexed with Immediate Offset**
   ```arm
   LDR R0, [R1, #4]          ; R0 = [R1 + 4]
   STR R2, [R3, #-8]         ; [R3 - 8] = R2
   ```

3. **Pre-indexed with Register Offset**
   ```arm
   LDR R0, [R1, R2]          ; R0 = [R1 + R2]
   ```

4. **Pre-indexed with Scaled Register Offset**
   ```arm
   LDR R0, [R1, R2, LSL #2]  ; R0 = [R1 + (R2 << 2)]
   ```

5. **Pre-indexed with Writeback**
   ```arm
   LDR R0, [R1, #4]!         ; R0 = [R1 + 4], R1 = R1 + 4
   ```

6. **Post-indexed with Immediate**
   ```arm
   LDR R0, [R1], #4          ; R0 = [R1], then R1 = R1 + 4
   ```

7. **Post-indexed with Register**
   ```arm
   STR R0, [R1], R2          ; [R1] = R0, then R1 = R1 + R2
   ```

---

## Shift Operations

All shift operations are available in data processing instructions. ✅

### LSL - Logical Shift Left ✅

**Operation:** Shifts bits left, fills with zeros

**Special Cases:**
- LSL #0: No shift performed (identity operation)

**Example:** `MOV R0, R1, LSL #4` shifts R1 left by 4 bits

### LSR - Logical Shift Right ✅

**Operation:** Shifts bits right, fills with zeros

**Special Cases:**
- LSR #0: In ARM encoding, this means LSR #32 (all bits shifted out, result is 0)
- LSR #32: All bits shifted out, result is 0, carry flag = bit 31 of original value

**Example:** `ADD R0, R1, R2, LSR #8` adds R1 + (R2 >> 8)

### ASR - Arithmetic Shift Right ✅

**Operation:** Shifts bits right, preserves sign bit (fills with bit 31)

**Special Cases:**
- ASR #0: In ARM encoding, this means ASR #32 (sign bit extended across all positions)
- ASR #32: Result is 0 if positive, -1 (0xFFFFFFFF) if negative

**Example:** `MOV R0, R1, ASR #2` performs signed divide by 4

### ROR - Rotate Right ✅

**Operation:** Rotates bits right, wraps around

**Special Cases:**
- ROR #0: In ARM encoding, this means RRX (rotate right extended through carry)

**Example:** `ORR R0, R1, R2, ROR #16` rotates R2 by 16 bits

### RRX - Rotate Right Extended ✅

**Operation:** Rotates right by 1 bit through carry flag (33-bit rotation with carry)

**Details:**
- Encoded as ROR #0 in ARM instruction format
- Bit 0 goes to carry flag
- Carry flag goes to bit 31
- Useful for multi-precision shifts

**Example:** `MOV R0, R1, RRX` rotates R1 right through carry

### Register-Specified Shifts ✅

**Operation:** Shift amount specified in register (bottom 8 bits used)

**Details:**
- Only the bottom 8 bits of the register are used for shift amount
- If shift amount is 0, no shift is performed
- If shift amount >= 32, result depends on shift type (LSL/LSR: 0, ASR: sign-extended)

**Example:** `MOV R0, R1, LSL R2` shifts R1 left by amount in R2

---

## CPSR Flags

The Current Program Status Register (CPSR) contains condition flags. ✅

| Flag | Name | Description |
|------|------|-------------|
| N | Negative | Set when result is negative (bit 31 = 1) |
| Z | Zero | Set when result is zero |
| C | Carry | Set on unsigned overflow (addition) or no borrow (subtraction) |
| V | Overflow | Set on signed overflow |

**Flag Update Rules:**
- Arithmetic operations (ADD, ADC, SUB, SBC, RSB, RSC): Update N, Z, C, V
- Logical operations (AND, ORR, EOR, BIC, MOV, MVN): Update N, Z, C (V unaffected)
- Comparison operations (CMP, CMN, TST, TEQ): Always update flags
- Multiply operations (MUL, MLA): Update N, Z only (C meaningless, V unaffected)
- S suffix required for most instructions to update flags
- Comparison instructions always update flags regardless of S bit

---

## Pseudo-Instructions

Pseudo-instructions are assembler conveniences that map to real instructions. ⏸️

| Pseudo | Real Instruction | Description |
|--------|------------------|-------------|
| NOP | MOV R0, R0 | No operation |
| LDR Rd, =value | LDR Rd, [PC, #offset] | Load 32-bit constant |
| ADR Rd, label | ADD Rd, PC, #offset | Load address |
| PUSH {regs} | STMDB SP!, {regs} | Push registers |
| POP {regs} | LDMIA SP!, {regs} | Pop registers |

---

## Register Usage Conventions

| Register | Alias | Purpose |
|----------|-------|---------|
| R0-R3 | - | Argument/result registers |
| R4-R11 | - | Local variables (callee-saved) |
| R12 | IP | Intra-procedure-call scratch register |
| R13 | SP | Stack pointer |
| R14 | LR | Link register (return address) |
| R15 | PC | Program counter |

---

## Notes

- **ARM2 Compatibility:** This emulator targets the ARM2 instruction set with select ARM2a extensions (halfword load/store)

- **Current Status (2025-10-14):**
  - Production hardening and comprehensive testing complete
  - 784 total tests (100% pass rate)
  - All 23 example programs fully tested with expected output verification
  - Integration test framework covers entire example suite
  - All core ARM2 instructions implemented and tested

- **Phase 11 Complete (2025-10-14):**
  - Comprehensive integration testing for all example programs
  - Expected output files for systematic regression testing
  - Fixed negative constant support in .equ directives
  - Fixed data section ordering bug
  - Added standalone shift instruction support (LSL, LSR, ASR, ROR)
  - Fixed 16-bit immediate encoding edge cases

- **Phase 10 Complete (2025-10-09):**
  - Cross-platform configuration management (config/) with TOML support
  - Execution and memory tracing (vm/trace.go) with register filtering
  - Performance statistics tracking (vm/statistics.go) with JSON/CSV/HTML export
  - Command-line flags: -trace, -mem-trace, -stats with file and format options

- **Assembler Features:**
  - Full directive support: `.text`, `.data`, `.bss`, `.global`, `.equ`, `.set`
  - Memory allocation: `.word`, `.half`, `.byte`, `.ascii`, `.asciz`, `.space`
  - Alignment directives: `.align`, `.balign`
  - Character literals with escape sequences
  - Immediate value support in multiple formats (decimal, hex, binary)

- **Recent Fixes (2025-10-09):**
  - PC Pipeline Handling: GetRegister() now returns PC+8 when reading R15 to simulate ARM pipeline effect
  - ROR #0 to RRX Conversion: Fixed encoding where ROR with shift amount 0 means RRX (rotate right extended)
  - LSR #0 Special Case: Now correctly treated as LSR #32 (shifts all bits out to 0)
  - ASR #0 Special Case: Now correctly treated as ASR #32 (preserves sign bit across all positions)
  - RRX Carry Calculation: Fixed carry flag handling for rotate right extended operations
  - Debugger Run Command: Fixed to preserve program memory using ResetRegisters() instead of Reset()

- **Test Coverage:** 784 unit and integration tests (100% pass rate) covering:
  - Debugger tests: 60 tests
  - Parser tests: 74 tests (including 39 character literal tests)
  - VM/Unit tests: 400 tests
  - Tools tests: 73 tests (linter: 25, formatter: 27, xref: 21)
  - Integration tests: 32 tests (complete example suite)
  - Config tests: 12 tests
  - Encoder tests: included in integration tests

- **Development Tools:**
  - Assembly Linter (tools/lint.go) - Code analysis with 25 tests
  - Code Formatter (tools/format.go) - Professional formatting with 27 tests
  - Cross-Reference Generator (tools/xref.go) - Symbol analysis with 21 tests

- **Future Extensions:** Long multiply instructions (UMULL, UMLAL, SMULL, SMLAL) and PSR transfer instructions (MRS, MSR) are planned

- **Cycle Accuracy:** Multiply instructions use cycle-accurate timing (2-16 cycles based on multiplier)

- **Memory Alignment:** Word accesses should be 4-byte aligned, halfword 2-byte aligned

---

## References

- ARM Architecture Reference Manual (ARMv2)
- ARM2 Datasheet (Acorn RISC Machine)
- Implementation files: `/vm/*.go`
- Test suite: `/tests/unit/vm/*_test.go`
- Progress tracking: `PROGRESS.md`
