# ARM2 Instruction Set Reference

This document details the ARM assembly instructions supported by this ARM2 emulator.

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

#### ADD - Add

**Syntax:** `ADD{cond}{S} Rd, Rn, <operand2>`

**Description:** Adds two values and stores the result

**Operation:** `Rd = Rn + operand2`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
ADD R0, R1, R2        ; R0 = R1 + R2
ADDS R3, R3, #1       ; R3 = R3 + 1, update flags
ADDEQ R4, R5, R6, LSL #2  ; If equal, R4 = R5 + (R6 << 2)
ADDNE R0, R0, #1      ; If not equal, R0 = R0 + 1
```

#### ADC - Add with Carry

**Syntax:** `ADC{cond}{S} Rd, Rn, <operand2>`

**Description:** Adds two values plus the carry flag

**Operation:** `Rd = Rn + operand2 + C`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
ADC R0, R1, R2        ; R0 = R1 + R2 + C
ADCS R3, R3, #0       ; R3 = R3 + C, update flags (for multi-precision)
```

#### SUB - Subtract

**Syntax:** `SUB{cond}{S} Rd, Rn, <operand2>`

**Description:** Subtracts operand2 from Rn

**Operation:** `Rd = Rn - operand2`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
SUB R0, R1, R2        ; R0 = R1 - R2
SUBS R3, R3, #1       ; R3 = R3 - 1, update flags
SUBGE R4, R5, #10     ; If greater or equal, R4 = R5 - 10
```

#### SBC - Subtract with Carry

**Syntax:** `SBC{cond}{S} Rd, Rn, <operand2>`

**Description:** Subtracts operand2 from Rn with borrow

**Operation:** `Rd = Rn - operand2 - NOT(C)`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
SBC R0, R1, R2        ; R0 = R1 - R2 - NOT(C)
SBCS R3, R3, #0       ; For multi-precision subtraction
```

#### RSB - Reverse Subtract

**Syntax:** `RSB{cond}{S} Rd, Rn, <operand2>`

**Description:** Subtracts Rn from operand2

**Operation:** `Rd = operand2 - Rn`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
RSB R0, R1, #10       ; R0 = 10 - R1
RSBS R2, R2, #0       ; R2 = -R2 (negate)
RSBLT R3, R4, #100    ; If less than, R3 = 100 - R4
```

#### RSC - Reverse Subtract with Carry

**Syntax:** `RSC{cond}{S} Rd, Rn, <operand2>`

**Description:** Subtracts Rn from operand2 with borrow

**Operation:** `Rd = operand2 - Rn - NOT(C)`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
RSC R0, R1, R2        ; R0 = R2 - R1 - NOT(C)
```

### Logical Operations

#### AND - Logical AND

**Syntax:** `AND{cond}{S} Rd, Rn, <operand2>`

**Description:** Performs bitwise AND

**Operation:** `Rd = Rn AND operand2`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
AND R0, R1, R2        ; R0 = R1 & R2
ANDS R3, R3, #0xFF    ; R3 = R3 & 0xFF, update flags
```

#### ORR - Logical OR

**Syntax:** `ORR{cond}{S} Rd, Rn, <operand2>`

**Description:** Performs bitwise OR

**Operation:** `Rd = Rn OR operand2`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
ORR R0, R1, R2        ; R0 = R1 | R2
ORRS R3, R3, #0x80    ; Set bit 7, update flags
ORRNE R4, R5, #1      ; If not equal, R4 = R5 | 1
```

#### EOR - Logical Exclusive OR

**Syntax:** `EOR{cond}{S} Rd, Rn, <operand2>`

**Description:** Performs bitwise exclusive OR

**Operation:** `Rd = Rn EOR operand2`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
EOR R0, R1, R2        ; R0 = R1 ^ R2
EORS R3, R3, R3       ; R3 = 0, update flags
```

#### BIC - Bit Clear

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

#### MOV - Move

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
MOVEQ R5, #1          ; If equal, R5 = 1
MOVNE R6, #0          ; If not equal, R6 = 0
MOVGT R7, R8          ; If greater than, R7 = R8
MOVLT R9, #-1         ; If less than, R9 = -1
MOVGE R10, #5         ; If greater or equal, R10 = 5
MOVLE R11, #0         ; If less or equal, R11 = 0
```

#### MVN - Move NOT

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

#### CMP - Compare

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

#### CMN - Compare Negative

**Syntax:** `CMN{cond} Rn, <operand2>`

**Description:** Compares two values by addition

**Operation:** `Rn + operand2` (result discarded)

**Flags:** Always updates N, Z, C, V

**Example:**
```arm
CMN R0, R1            ; Compare R0 with -R1
CMN R2, #-5           ; Test if R2 is 5
```

#### TST - Test Bits

**Syntax:** `TST{cond} Rn, <operand2>`

**Description:** Tests bits by AND operation

**Operation:** `Rn AND operand2` (result discarded)

**Flags:** Always updates N, Z, C (V unaffected)

**Example:**
```arm
TST R0, #0x01         ; Test if bit 0 is set
TST R1, R2            ; Test bits in common
```

#### TEQ - Test Equivalence

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

#### LDR - Load Word

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

#### STR - Store Word

**Syntax:** `STR{cond} Rd, <addressing_mode>`

**Description:** Stores a 32-bit word to memory

**Operation:** `Memory[address] = Rd`

**Example:**
```arm
STR R0, [R1]          ; [R1] = R0
STR R2, [R3, #-4]     ; [R3 - 4] = R2
STR R4, [R5, R6, LSL #2]!  ; [R5 + (R6 << 2)] = R4, writeback
```

#### LDRB - Load Byte

**Syntax:** `LDRB{cond} Rd, <addressing_mode>`

**Description:** Loads an 8-bit byte from memory (zero-extended)

**Operation:** `Rd = ZeroExtend(Memory[address])`

**Example:**
```arm
LDRB R0, [R1]         ; R0 = byte at [R1]
LDRB R2, [R3, #1]     ; R2 = byte at [R3 + 1]
```

#### STRB - Store Byte

**Syntax:** `STRB{cond} Rd, <addressing_mode>`

**Description:** Stores an 8-bit byte to memory

**Operation:** `Memory[address] = Rd[7:0]`

**Example:**
```arm
STRB R0, [R1]         ; [R1] = R0[7:0]
STRB R2, [R3, #10]    ; [R3 + 10] = R2[7:0]
```

#### LDRH - Load Halfword
**Syntax:** `LDRH{cond} Rd, <addressing_mode>`

**Description:** Loads a 16-bit halfword from memory (zero-extended)

**Operation:** `Rd = ZeroExtend(Memory[address])`

**Example:**
```arm
LDRH R0, [R1]         ; R0 = halfword at [R1]
LDRH R2, [R3, #2]     ; R2 = halfword at [R3 + 2]
```

#### STRH - Store Halfword
**Syntax:** `STRH{cond} Rd, <addressing_mode>`

**Description:** Stores a 16-bit halfword to memory

**Operation:** `Memory[address] = Rd[15:0]`

**Example:**
```arm
STRH R0, [R1]         ; [R1] = R0[15:0]
STRH R2, [R3, #6]     ; [R3 + 6] = R2[15:0]
```

### Multiple Data Transfer

#### LDM - Load Multiple

**Syntax:** `LDM{cond}{mode} Rn{!}, {register_list}{^}`

**Description:** Loads multiple registers from consecutive memory locations

**Modes:** IA (Increment After), IB (Increment Before), DA (Decrement After), DB (Decrement Before)

**Stack Modes:** FD (Full Descending), ED (Empty Descending), FA (Full Ascending), EA (Empty Ascending)

**Example:**
```arm
LDMIA R13!, {R0-R3}      ; Load R0-R3 from stack, increment R13
LDMFD SP!, {R4-R6, PC}   ; Pop R4-R6 and return
LDMIA R0, {R1-R4}        ; Load R1-R4 from memory at R0
LDMFD SP!, {R0-R12, LR, PC}^  ; Exception return (restore CPSR from SPSR)
```

**Usage Note:** `LDMFD SP!` (Load Multiple Full Descending) is the standard way to pop registers from the stack.

**S Bit (^ suffix):** When the `^` suffix is used with PC in the register list, the instruction performs an exception return by restoring the CPSR from SPSR. This simulates returning from an exception handler where the processor status needs to be restored along with the program counter. The S bit has no effect if PC is not in the register list.

**Example Exception Return:**
```arm
; Exception handler epilogue
LDMFD SP!, {R0-R12, LR, PC}^  ; Restore all registers and CPSR from SPSR
```

#### STM - Store Multiple

**Syntax:** `STM{cond}{mode} Rn{!}, {register_list}{^}`

**Description:** Stores multiple registers to consecutive memory locations

**Modes:** IA (Increment After), IB (Increment Before), DA (Decrement After), DB (Decrement Before)

**Stack Modes:** FD (Full Descending), ED (Empty Descending), FA (Full Ascending), EA (Empty Ascending)

**Example:**
```arm
STMDB SP!, {R0-R3, LR}  ; Push R0-R3 and LR to stack
STMFD SP!, {R4-R11}     ; Push R4-R11 to stack
STMIA R0!, {R1-R4}      ; Store R1-R4 to memory at R0, increment R0
```

**Usage Note:** `STMFD SP!` (Store Multiple Full Descending) is the standard way to push registers onto the stack.

**S Bit (^ suffix):** The `^` suffix sets the S bit (bit 22) in the instruction encoding. For STM instructions, this bit has no special behavior in this implementation - registers are stored normally with PC stored as PC+12 when included in the register list. The S bit is primarily used with LDM instructions for exception returns.

---

## Branch Instructions

#### B - Branch

**Syntax:** `B{cond} label`

**Description:** Branches to a label/address

**Operation:** `PC = PC + offset`

**Range:** ±32MB from current instruction

**Example:**
```arm
B loop                ; Branch to loop
BEQ equal_case        ; Branch if equal
BNE not_zero          ; Branch if not zero
BGT greater           ; Branch if greater than
BLT less_than         ; Branch if less than
BGE greater_equal     ; Branch if greater or equal
BLE less_equal        ; Branch if less or equal
BCS carry_set         ; Branch if carry set
BMI minus             ; Branch if minus/negative
```

#### BL - Branch with Link

**Syntax:** `BL{cond} label`

**Description:** Branches to a subroutine and saves return address

**Operation:** `LR = PC + 4, PC = PC + offset`

**Range:** ±32MB from current instruction

**Example:**
```arm
BL function           ; Call function
BLEQ conditional_fn   ; Call if equal
```

#### BX - Branch and Exchange

**Syntax:** `BX{cond} Rm`

**Description:** Branches to address in register (ARM/Thumb interworking)

**Operation:** `PC = Rm & 0xFFFFFFFE` (bit 0 would indicate Thumb mode in later ARM)

**Example:**
```arm
BX LR                 ; Return from subroutine
BX R0                 ; Branch to address in R0
```

#### BLX - Branch with Link and Exchange

**Syntax:** `BLX{cond} Rm`

**Description:** Branches to address in register and saves return address (ARM/Thumb interworking)

**Operation:** `LR = PC + 4, PC = Rm & 0xFFFFFFFE` (bit 0 would indicate Thumb mode in later ARM)

**Example:**
```arm
BLX R7                ; Call function at address in R7
BLX R0                ; Call function at address in R0
```

---

## Multiply Instructions

#### MUL - Multiply

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

#### MLA - Multiply-Accumulate

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

#### UMULL - Unsigned Multiply Long
**Syntax:** `UMULL{cond}{S} RdLo, RdHi, Rm, Rs`

**Description:** Multiplies two 32-bit unsigned values producing 64-bit result

**Operation:** `RdHi:RdLo = Rm * Rs`

**Flags:** Updates N, Z when S bit is set (C, V unaffected)

**Restrictions:** RdHi, RdLo, and Rm must all be different registers, R15 (PC) cannot be used

**Example:**
```arm
UMULL R0, R1, R2, R3   ; R1:R0 = R2 * R3 (unsigned)
UMULLS R4, R5, R6, R7  ; R5:R4 = R6 * R7, update flags
```

#### UMLAL - Unsigned Multiply-Accumulate Long
**Syntax:** `UMLAL{cond}{S} RdLo, RdHi, Rm, Rs`

**Description:** Unsigned multiply and accumulate with 64-bit result

**Operation:** `RdHi:RdLo = (Rm * Rs) + RdHi:RdLo`

**Flags:** Updates N, Z when S bit is set (C, V unaffected)

**Restrictions:** RdHi, RdLo, and Rm must all be different registers, R15 (PC) cannot be used

**Example:**
```arm
UMLAL R0, R1, R2, R3   ; R1:R0 = (R2 * R3) + R1:R0 (unsigned)
UMLALS R4, R5, R6, R7  ; R5:R4 += R6 * R7, update flags
```

#### SMULL - Signed Multiply Long
**Syntax:** `SMULL{cond}{S} RdLo, RdHi, Rm, Rs`

**Description:** Multiplies two 32-bit signed values producing 64-bit result

**Operation:** `RdHi:RdLo = Rm * Rs (signed)`

**Flags:** Updates N, Z when S bit is set (C, V unaffected)

**Restrictions:** RdHi, RdLo, and Rm must all be different registers, R15 (PC) cannot be used

**Example:**
```arm
SMULL R0, R1, R2, R3   ; R1:R0 = R2 * R3 (signed)
SMULLS R4, R5, R6, R7  ; R5:R4 = R6 * R7 (signed), update flags
```

#### SMLAL - Signed Multiply-Accumulate Long
**Syntax:** `SMLAL{cond}{S} RdLo, RdHi, Rm, Rs`

**Description:** Signed multiply and accumulate with 64-bit result

**Operation:** `RdHi:RdLo = (Rm * Rs) + RdHi:RdLo (signed)`

**Flags:** Updates N, Z when S bit is set (C, V unaffected)

**Restrictions:** RdHi, RdLo, and Rm must all be different registers, R15 (PC) cannot be used

**Example:**
```arm
SMLAL R0, R1, R2, R3   ; R1:R0 = (R2 * R3) + R1:R0 (signed)
SMLALS R4, R5, R6, R7  ; R5:R4 += R6 * R7 (signed), update flags
```

---

## System Instructions

### SWI - Software Interrupt

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

##### Console I/O (0x00-0x07)- `0x00` - **EXIT** - Exit program (R0 = exit code)- `0x01` - **WRITE_CHAR** - Write character (R0 = character)- `0x02` - **WRITE_STRING** - Write null-terminated string (R0 = address)- `0x03` - **WRITE_INT** - Write integer (R0 = value, R1 = base)- `0x04` - **READ_CHAR** - Read character (returns in R0)- `0x05` - **READ_STRING** - Read string (R0 = buffer, R1 = max length)- `0x06` - **READ_INT** - Read integer (returns in R0)- `0x07` - **WRITE_NEWLINE** - Write newline
##### File Operations (0x10-0x16)- `0x10` - **OPEN** - Open file (R0 = filename, R1 = mode)- `0x11` - **CLOSE** - Close file (R0 = file handle)- `0x12` - **READ** - Read from file (R0 = handle, R1 = buffer, R2 = size)- `0x13` - **WRITE** - Write to file (R0 = handle, R1 = buffer, R2 = size)- `0x14` - **SEEK** - Seek in file (R0 = handle, R1 = offset, R2 = whence)- `0x15` - **TELL** - Get file position (R0 = handle)- `0x16` - **FILE_SIZE** - Get file size (R0 = handle)
##### Memory Operations (0x20-0x22)- `0x20` - **ALLOCATE** - Allocate memory (R0 = size, returns address in R0)- `0x21` - **FREE** - Free memory (R0 = address)- `0x22` - **REALLOCATE** - Reallocate memory (R0 = address, R1 = new size)
##### System Information (0x30-0x33)- `0x30` - **GET_TIME** - Get time in milliseconds (returns in R0)- `0x31` - **GET_RANDOM** - Get random 32-bit value (returns in R0)- `0x32` - **GET_ARGUMENTS** - Get command-line arguments- `0x33` - **GET_ENVIRONMENT** - Get environment variables
##### Error Handling (0x40-0x42)- `0x40` - **GET_ERROR** - Get last error code- `0x41` - **SET_ERROR** - Set error code- `0x42` - **PRINT_ERROR** - Print error message
##### Debugging Support (0xF0-0xF4)- `0xF0` - **DEBUG_PRINT** - Print debug message (R0 = string address)- `0xF1` - **BREAKPOINT** - Trigger breakpoint- `0xF2` - **DUMP_REGISTERS** - Dump all registers- `0xF3` - **DUMP_MEMORY** - Dump memory region (R0 = address, R1 = length)- `0xF4` - **ASSERT** - Assert condition (R0 = condition)
### MRS - Move PSR to Register
**Syntax:** `MRS{cond} Rd, PSR`

**Description:** Moves CPSR or SPSR to a register

**Operation:** `Rd = CPSR` (reads all 32 bits of the status register)

**Restrictions:** Rd cannot be R15 (PC)

**Example:**
```arm
MRS R0, CPSR          ; R0 = CPSR (read current flags)
MRS R1, CPSR          ; R1 = CPSR
```

**Use Cases:**
- Reading current processor flags for later restoration
- Implementing critical sections in interrupt handlers
- Debugging flag states

### MSR - Move Register to PSR
**Syntax:** `MSR{cond} PSR_fields, Rm` or `MSR{cond} PSR_fields, #immediate`

**Description:** Moves a register or immediate to CPSR or SPSR flags

**Operation:** `CPSR_flags = Rm` or `CPSR_flags = immediate` (writes to flag bits only)

**Restrictions:** Rm cannot be R15 (PC)

**Fields:** `_f` indicates flag field (bits 31-24: N, Z, C, V)

**Example:**
```arm
MSR CPSR_f, R0        ; CPSR flags = R0 (register form)
MSR CPSR_f, #0xF0000000  ; Set all flags (immediate form)
MSR CPSR_f, R1        ; Restore saved flags from R1
```

**Use Cases:**
- Restoring processor flags after critical section
- Manually setting/clearing flags for testing
- Context switching in operating systems

---

## Assembler Directives

Assembler directives control how the assembler processes your source code. They don't generate instructions but affect memory layout, symbol definitions, and code organization.

### Directive Quick Reference

| Directive | Category | Description |
|-----------|----------|-------------|
| `.text` | Section | Mark beginning of code section |
| `.data` | Section | Mark beginning of data section |
| `.global` | Symbol | Declare symbol as global/exported |
| `.equ` / `.set` | Symbol | Define a constant value |
| `.org` | Memory | Set assembly origin address |
| `.word` | Data | Allocate 32-bit words (4 bytes each) |
| `.half` | Data | Allocate 16-bit halfwords (2 bytes each) |
| `.byte` | Data | Allocate 8-bit bytes (1 byte each) |
| `.ascii` | String | Allocate string without null terminator |
| `.asciz` / `.string` | String | Allocate null-terminated string |
| `.space` / `.skip` | Memory | Reserve bytes (initialized to zero) |
| `.align` | Alignment | Align to 2^n byte boundary |
| `.balign` | Alignment | Align to specified byte boundary |
| `.ltorg` | Literal Pool | Place literal pool at current location |

### Section Directives
#### .text
**Description:** Marks the beginning of the code section (executable instructions)

**Syntax:** `.text`

**Details:**
- Indicates that subsequent lines contain executable code
- Multiple `.text` sections can appear in the same file
- If no `.org` directive has been set, the first `.text` section starts at address 0
- Sections can be interleaved (`.text`, `.data`, `.text`, etc.)
- The assembler tracks addresses sequentially across all sections

**Example:**
```arm
.text
.global _start
_start:
    MOV R0, #1
    BL main
    SWI #0x00

main:
    MOV R0, #42
    MOV PC, LR
```

**Multiple Sections:**
```arm
.text
function1:
    MOV R0, #1
    MOV PC, LR

.data
value: .word 100

.text           ; Second code section
function2:
    LDR R0, =value
    LDR R0, [R0]
    MOV PC, LR
```

#### .data
**Description:** Marks the beginning of the data section for initialized data

**Syntax:** `.data`

**Details:**
- Indicates that subsequent lines contain data definitions
- Used for variables, constants, strings, and arrays
- Multiple `.data` sections can appear in the same file
- If no `.org` directive has been set, the first `.data` section starts at address 0
- Data section can be interleaved with `.text` sections
- The assembler tracks addresses sequentially across all sections

**Example:**
```arm
.data
counter:    .word 0
message:    .asciz "Hello, World!\n"
buffer:     .space 256
array:      .word 1, 2, 3, 4, 5
```

**Organized Data Layout:**
```arm
.data
; String constants
msg1:       .asciz "Ready\n"
msg2:       .asciz "Done\n"

; Numeric constants
.align 2
max_val:    .word 1000
min_val:    .word 0

; Arrays
.align 2
lookup:     .word 0, 1, 4, 9, 16, 25

; Buffers
.align 2
input_buf:  .space 512
```

**Note:** In this emulator, `.data` and `.text` sections can be freely interleaved. The assembler tracks addresses sequentially regardless of section type.

### Symbol Directives
#### .global
**Description:** Declares a symbol as global (visible to other modules/exported)

**Syntax:** `.global symbol_name`

**Details:**
- Marks a label or symbol as globally visible
- In a multi-module program, global symbols can be referenced from other files
- Commonly used for entry points (like `_start`) and public functions
- In this single-module emulator, all symbols are visible, but `.global` is still accepted for compatibility
- Multiple symbols can be declared global with separate `.global` directives

**Example:**
```arm
.global _start
.global my_function
.global my_data

.text
_start:
    BL my_function
    SWI #0x00

my_function:
    MOV R0, #42
    MOV PC, LR

.data
my_data:
    .word 100
```

**Common Pattern:**
```arm
; Declare all public symbols at the top
.global _start
.global add_numbers
.global multiply

.text
_start:
    ; Entry point
    MOV R0, #5
    MOV R1, #10
    BL add_numbers
    SWI #0x00

add_numbers:
    ADD R0, R0, R1
    MOV PC, LR

multiply:
    MUL R0, R0, R1
    MOV PC, LR
```

#### .equ / .set
**Description:** Defines a constant value for a symbol (similar to #define in C)

**Syntax:** `.equ symbol, value` or `.set symbol, value`

**Details:**
- Creates a named constant that can be used throughout the program
- `.equ` and `.set` are equivalent (both define constants)
- The constant can be used in place of immediate values or addresses
- Values can be decimal, hexadecimal, binary, or character literals
- Constants can reference previously defined constants
- Negative values are supported

**Supported value formats:**
- Decimal: `256`, `-10`, `1000`
- Hexadecimal: `0x100`, `0xFF`, `0xDEADBEEF`
- Binary: `0b11111111`, `0b1010`
- Character literals: `'A'`, `'\n'`
- Expressions: Can reference other constants

**Example:**
```arm
; Basic constants
.equ BUFFER_SIZE, 256
.equ MAX_COUNT, 100
.set STACK_SIZE, 0x1000

; Using constants
.data
buffer:     .space BUFFER_SIZE

.text
MOV R0, #MAX_COUNT
LDR SP, =STACK_SIZE
```

**Hexadecimal and Binary:**
```arm
.equ STATUS_READY,  0b00000001
.equ STATUS_BUSY,   0b00000010
.equ STATUS_ERROR,  0b00000100

.equ PERIPHERAL_BASE, 0x40000000
.equ GPIO_OFFSET,     0x1000

.text
MOV R0, #STATUS_READY
LDR R1, =PERIPHERAL_BASE
```

**Character Literals and Negative Values:**
```arm
.equ NEWLINE,     '\n'
.equ SPACE,       ' '
.equ MINUS_ONE,   -1
.equ NEG_OFFSET,  -16

.text
MOV R0, #NEWLINE
MOV R1, #MINUS_ONE
```

**Referencing Other Constants:**
```arm
.equ KB,          1024
.equ BUFFER_SIZE, 16 * KB    ; 16KB
.equ STACK_TOP,   0x10000
.equ STACK_SIZE,  4 * KB
```

### Memory Allocation Directives
#### .org
**Description:** Sets the assembly origin address (starting address for code/data)

**Syntax:** `.org address`

**Details:**
- Sets the starting memory address for subsequent instructions and data
- Can be used multiple times to relocate code/data segments
- Address can be in decimal, hexadecimal (0x prefix), or binary (0b prefix)
- If not specified, the first section (.text or .data) defaults to address 0
- The first `.org` directive also sets the program's entry point origin

**Example:**
```arm
.org 0x8000        ; Start code at address 0x8000

.text
.org 0x8000
main:
    MOV R0, #1
    BL function
    SWI #0x00

function:
    MOV R0, #42
    MOV PC, LR

.data
.org 0x9000        ; Data section starts at 0x9000
buffer: .space 256
value:  .word 100
```

**Multiple .org Example:**
```arm
.text
.org 0x8000
vector_table:
    B reset_handler
    B irq_handler

.org 0x8100
reset_handler:
    MOV SP, #0x10000
    B main
```

#### .word
**Description:** Allocates and initializes 32-bit words (4 bytes each)

**Syntax:** `.word value1, value2, ...`

**Details:**
- Each value is stored as a 32-bit (4-byte) word
- Values can be numbers, character literals, or label addresses
- Multiple values can be specified separated by commas
- Values are stored in little-endian format on ARM
- Commonly used for arrays, lookup tables, and constants

**Supported value formats:**
- Decimal: `42`, `-10`, `1000`
- Hexadecimal: `0x1234`, `0xABCDEF00`, `0xFF`
- Binary: `0b11010101`, `0b1111`
- Character literals: `'A'`, `'\n'`

**Example:**
```arm
.data
; Simple array
array:      .word 10, 20, 30, 40

; Hexadecimal values
table:      .word 0x12345678, 0xABCDEF00, 0xDEADBEEF

; Mixed formats
mixed:      .word 100, 0xFF, 0b11110000, 'A'

; Single value
counter:    .word 0

; Large array
fibonacci:  .word 0, 1, 1, 2, 3, 5, 8, 13, 21, 34
```

#### .half
**Description:** Allocates and initializes 16-bit halfwords (2 bytes each)

**Syntax:** `.half value1, value2, ...`

**Details:**
- Each value is stored as a 16-bit (2-byte) halfword
- Values are truncated to 16 bits if larger
- Multiple values can be specified separated by commas
- Values are stored in little-endian format
- Useful for 16-bit data arrays and smaller constants

**Supported value formats:**
- Decimal: `42`, `-10`, `1000`
- Hexadecimal: `0x1234`, `0xFFFF`
- Binary: `0b1101010101010101`
- Character literals: `'A'`, `'0'`

**Example:**
```arm
.data
; 16-bit array
shorts:     .half 100, 200, 300, 400

; Hexadecimal halfwords
colors:     .half 0xF800, 0x07E0, 0x001F  ; RGB565 colors

; Mixed formats
data:       .half 1000, 0x1234, 0b1111000011110000

; Single halfword
port_value: .half 0x5555
```

#### .byte
**Description:** Allocates and initializes 8-bit bytes (1 byte each)

**Syntax:** `.byte value1, value2, ...`

**Details:**
- Each value is stored as an 8-bit (1-byte) byte
- Values are truncated to 8 bits if larger
- Multiple values can be specified separated by commas
- Useful for byte arrays, flags, and character data
- Character literals are commonly used with `.byte`

**Supported value formats:**
- Decimal: `42`, `255`, `0`
- Hexadecimal: `0x01`, `0xFF`, `0xAB`
- Binary: `0b11010101`, `0b1111`
- Character literals: `'A'`, `'B'`, `'\n'`, `'\0'`

**Example:**
```arm
.data
; Byte array
bytes:      .byte 0x01, 0x02, 0x03, 0xFF

; Character data (without null terminator)
flags:      .byte 'A', 'B', 'C', 'D'

; Status flags
status:     .byte 0b00000001, 0b00000010, 0b00000100

; Hex dump
dump:       .byte 0xDE, 0xAD, 0xBE, 0xEF

; Mixed formats
mixed:      .byte 65, 0x42, 0b01000011, 'D'  ; All are ASCII letters

; Null-terminated string (manual)
msg:        .byte 'H', 'i', '\n', 0  ; "Hi\n" with null terminator
```

#### .ascii
**Description:** Allocates a string without null terminator

**Syntax:** `.ascii "string"`

**Details:**
- Stores the string bytes without adding a null terminator
- Each character is stored as a single byte (ASCII/UTF-8)
- Useful when you need exact byte sequences or will add null manually
- Supports escape sequences for special characters
- The string length equals the number of characters (escape sequences count as 1)

**Example:**
```arm
.data
; String without null terminator (5 bytes)
msg:        .ascii "Hello"

; Multiple strings can be concatenated
banner:     .ascii "======"
            .ascii " ARM "
            .ascii "======"

; String with escape sequences
formatted:  .ascii "Line1\nLine2\tTabbed"

; Using with manual null terminator
cstring:    .ascii "Manual"
            .byte 0           ; Add null terminator manually
```

#### .asciz / .string
**Description:** Allocates a null-terminated string (C-style string)

**Syntax:** `.asciz "string"` or `.string "string"`

**Details:**
- Stores the string bytes and automatically adds a null terminator (0x00)
- `.asciz` and `.string` are equivalent (both add null terminator)
- String length is (number of characters + 1) for the null byte
- Ideal for C-style strings used with syscalls and string functions
- Supports escape sequences for special characters

**Example:**
```arm
.data
; Null-terminated string (6 bytes: 'H','e','l','l','o',0)
msg:        .asciz "Hello"

; Equivalent to .asciz
prompt:     .string "Enter name: "

; String with newline (syscall-ready)
greeting:   .asciz "Hello, World!\n"

; Multiple strings
error1:     .asciz "File not found"
error2:     .asciz "Access denied"

; Empty string (just the null terminator)
empty:      .asciz ""           ; 1 byte: 0
```

**Escape Sequences:** Both `.ascii` and `.asciz` support standard escape sequences:

| Escape | Description | Hex Value |
|--------|-------------|-----------|
| `\n` | Newline (LF) | 0x0A |
| `\r` | Carriage return (CR) | 0x0D |
| `\t` | Tab | 0x09 |
| `\b` | Backspace | 0x08 |
| `\\` | Backslash | 0x5C |
| `\"` | Double quote | 0x22 |
| `\'` | Single quote | 0x27 |
| `\0` | Null character | 0x00 |

**Escape Sequence Examples:**
```arm
.data
; Multi-line string
greeting:   .asciz "Hello\nWorld\n"

; Windows-style path with backslashes
path:       .asciz "C:\\Users\\Name\\file.txt"

; String with quotes
quoted:     .asciz "He said, \"Hello!\""

; Tab-separated values
tsv:        .asciz "Name\tAge\tCity\n"

; Mixed escape sequences
mixed:      .asciz "Line1\r\nLine2\tTab\0Extra"
```

#### .space / .skip
**Description:** Reserves specified number of bytes (initialized to zero)

**Syntax:** `.space size` or `.skip size`

**Details:**
- Reserves the specified number of bytes in memory
- All bytes are initialized to zero (0x00)
- `.space` and `.skip` are equivalent
- Size can be a decimal, hexadecimal, or binary number
- Size can also reference a constant defined with `.equ`
- Useful for buffers, arrays, and uninitialized data

**Example:**
```arm
.data
; 256-byte buffer (all zeros)
buffer:     .space 256

; 4KB stack space
stack:      .skip 0x1000

; Using constants
.equ BUFFER_SIZE, 512
input_buf:  .space BUFFER_SIZE

; Aligned buffer allocation
.align 2
.equ ARRAY_SIZE, 100
array:      .space ARRAY_SIZE * 4  ; 100 words = 400 bytes

; Multiple buffers
tx_buffer:  .space 128
rx_buffer:  .space 128

; Large memory region
heap:       .space 0x10000    ; 64KB
```

**Usage Pattern with Initialization:**
```arm
.data
; Define buffer size
.equ BUF_SIZE, 256

; Reserve buffer space
read_buffer:    .space BUF_SIZE

; Define pointer to buffer
buffer_ptr:     .word read_buffer

.text
; Use buffer in code
LDR R0, =read_buffer
MOV R1, #BUF_SIZE
BL clear_buffer
```

### Character Literals
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

### Alignment Directives
#### .align
**Description:** Aligns the current address to a 2^n bytes boundary

**Syntax:** `.align n`

**Details:**
- Aligns to 2^n byte boundary (power of 2)
- Pads with zero bytes to reach the alignment boundary
- Commonly used values:
  - `.align 0` = 1-byte alignment (2^0, no effect)
  - `.align 1` = 2-byte alignment (2^1)
  - `.align 2` = 4-byte alignment (2^2, word alignment)
  - `.align 3` = 8-byte alignment (2^3, double-word)
  - `.align 4` = 16-byte alignment (2^4, cache line)

**Example:**
```arm
.data
.align 2          ; Align to 4-byte boundary (2^2)
value1: .word 100

.byte 1, 2, 3     ; 3 bytes
.align 2          ; Pad with 1 byte to reach 4-byte boundary
value2: .word 200 ; Now word-aligned

.text
.align 2          ; Ensure instructions are word-aligned
function:
    MOV R0, #1
    MOV PC, LR
```

#### .balign
**Description:** Aligns the current address to the specified byte boundary

**Syntax:** `.balign boundary`

**Details:**
- Aligns to the exact byte boundary specified (not a power of 2)
- Pads with zero bytes to reach the alignment boundary
- More intuitive than `.align` for specific byte boundaries
- Common values: 4 (word), 8 (double-word), 16 (cache line)

**Example:**
```arm
.data
.balign 4         ; Align to 4-byte boundary
array: .word 1, 2, 3, 4

.byte 0xFF        ; 1 byte
.balign 4         ; Pad with 3 bytes to reach 4-byte boundary
next_word: .word 0x12345678

.text
.balign 16        ; Align to 16-byte boundary (cache line)
critical_loop:
    ; Performance-critical code
    CMP R0, #0
    BNE critical_loop
```

**Alignment Comparison:**
```arm
; These are equivalent:
.align 2          ; 2^2 = 4 bytes
.balign 4         ; 4 bytes

; These are equivalent:
.align 3          ; 2^3 = 8 bytes
.balign 8         ; 8 bytes
```

### Literal Pool Directive
#### .ltorg
**Description:** Places the literal pool at the current location

**Syntax:** `.ltorg`

**Purpose:** Used with the `LDR Rd, =value` pseudo-instruction to control where 32-bit constants are stored in memory

**Details:**
- Literals must be within ±4095 bytes of the LDR instruction
- Multiple `.ltorg` directives can be used in large programs
- Values are automatically deduplicated
- Pool is 4-byte aligned automatically
- If no `.ltorg` is specified, a pool is placed at end of program

**Example:**
```arm
.text
.org 0x8000

main:
    LDR R0, =0x12345678   ; Load large constant
    LDR R1, =0xDEADBEEF   ; Load another constant
    ADD R2, R0, R1
    B end

    .ltorg                ; Place literal pool here

end:
    MOV R0, #0
    SWI #0x00
```

**Multiple Pools Example:**
```arm
section1:
    LDR R0, =0x11111111
    LDR R1, =0x22222222
    .ltorg                ; First pool

section2:
    LDR R2, =0x33333333
    LDR R3, =0x44444444
    .ltorg                ; Second pool
```

### Directive Usage Examples

**Complete program demonstrating all directives:**
```arm
; ============================================
; ARM Assembly Program - All Directives Demo
; ============================================

; Define constants using .equ and .set
.equ BUFFER_SIZE, 256
.equ EXIT_SYSCALL, 0x00
.equ WRITE_STRING, 0x02
.set STACK_SIZE, 0x1000
.equ NEWLINE, '\n'

; Declare global symbols
.global _start
.global process_data

; ============================================
; Code Section
; ============================================
.text
.org 0x8000        ; Set origin to 0x8000

_start:
    ; Initialize stack pointer
    LDR SP, =stack_top

    ; Print welcome message
    LDR R0, =welcome_msg
    SWI #WRITE_STRING

    ; Process some data
    BL process_data

    ; Exit program
    MOV R0, #0
    SWI #EXIT_SYSCALL

; Align function to 4-byte boundary
.align 2
process_data:
    ; Save registers
    PUSH {R4-R6, LR}

    ; Load data array address
    LDR R4, =data_array
    LDR R5, =result
    MOV R6, #0

    ; Sum array values
    LDR R0, [R4]
    LDR R1, [R4, #4]
    LDR R2, [R4, #8]
    ADD R6, R0, R1
    ADD R6, R6, R2

    ; Store result
    STR R6, [R5]

    ; Restore and return
    POP {R4-R6, PC}

; Place literal pool here
.ltorg

; ============================================
; Data Section
; ============================================
.data

; String data with .asciz (null-terminated)
welcome_msg:    .asciz "ARM Emulator Demo\n"
prompt:         .asciz "Enter value: "
done_msg:       .string "Processing complete\n"

; String without null using .ascii
banner:         .ascii "======\n"

; Word data (32-bit)
.align 2
data_array:     .word 10, 20, 30, 40, 50
result:         .word 0
counter:        .word 0

; Halfword data (16-bit)
.align 1
port_values:    .half 0x1234, 0x5678, 0xABCD

; Byte data (8-bit)
status_flags:   .byte 0x01, 0x02, 0x04, 0x08
char_array:     .byte 'A', 'R', 'M', '2'

; Mixed format data
mixed_data:     .word 100, 0xFF, 0b11110000, 'X'

; Reserved buffer space (initialized to zero)
.align 2
read_buffer:    .space BUFFER_SIZE
temp_buffer:    .skip 128

; Stack space
.balign 16      ; Align stack to 16-byte boundary
stack_bottom:   .space STACK_SIZE
stack_top:      ; Label marks top of stack

; ============================================
; Additional Code Section (interleaved)
; ============================================
.text

; Helper function
.align 2
clear_buffer:
    PUSH {R0-R2, LR}
    LDR R0, =read_buffer
    MOV R1, #0
    MOV R2, #BUFFER_SIZE
clear_loop:
    STRB R1, [R0], #1
    SUBS R2, R2, #1
    BNE clear_loop
    POP {R0-R2, PC}

; Final literal pool
.ltorg
```

**Simple program structure:**
```arm
; Constants
.equ EXIT, 0x00

; Entry point
.text
.org 0x8000
.global _start

_start:
    MOV R0, #42
    SWI #EXIT

; Data
.data
value:      .word 100
```

**Mixed code and data:**
```arm
.text
function1:
    MOV R0, R1
    MOV PC, LR

.data
value:      .word 42

.text
function2:
    LDR R0, =value
    LDR R0, [R0]
    MOV PC, LR
```

**Using alignment directives:**
```arm
.data
; Byte data (may be at odd address)
.byte 0x01, 0x02, 0x03

; Align to 4-byte boundary before word
.align 2
word_value: .word 0x12345678

; Align to 16-byte boundary
.balign 16
cache_aligned: .word 1, 2, 3, 4
```

---

## Condition Codes

All instructions can be conditionally executed based on CPSR flags.
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

ARM2 provides flexible addressing modes for both data processing instructions and memory access instructions.

### Data Processing Operand2
Data processing instructions (ADD, SUB, MOV, etc.) use the **Operand2** field, which can be:

1. **Immediate Value with Rotation**
   ```arm
   MOV R0, #255          ; R0 = 255
   ADD R1, R2, #0x100    ; R1 = R2 + 256
   ```
   - 8-bit immediate value rotated right by an even number of positions (0-30)
   - Allows encoding of many common constants efficiently
   - If value cannot be encoded, assembler will error or use literal pool

2. **Register**
   ```arm
   ADD R0, R1, R2        ; R0 = R1 + R2
   MOV R0, R1            ; R0 = R1
   ```
   - Simple register value, no shift applied

3. **Register with Logical Shift Left (LSL)**
   ```arm
   ADD R0, R1, R2, LSL #2    ; R0 = R1 + (R2 << 2)
   MOV R0, R1, LSL #4        ; R0 = R1 << 4
   ```
   - Shift amount: 0-31
   - LSL #0 means no shift
   - Commonly used for multiplication by powers of 2

4. **Register with Logical Shift Right (LSR)**
   ```arm
   SUB R0, R1, R2, LSR #4    ; R0 = R1 - (R2 >> 4)
   MOV R0, R1, LSR #8        ; R0 = R1 >> 8 (unsigned)
   ```
   - Shift amount: 1-32
   - LSR #32 means all bits shifted out (result = 0)
   - In encoding, LSR #0 is interpreted as LSR #32

5. **Register with Arithmetic Shift Right (ASR)**
   ```arm
   MOV R0, R1, ASR #8        ; R0 = R1 >> 8 (signed)
   SUB R0, R1, R2, ASR #2    ; R0 = R1 - (R2 >> 2, signed)
   ```
   - Shift amount: 1-32
   - Preserves sign bit (bit 31)
   - ASR #32 means sign-extend across all bits
   - In encoding, ASR #0 is interpreted as ASR #32

6. **Register with Rotate Right (ROR)**
   ```arm
   ORR R0, R1, R2, ROR #16   ; R0 = R1 | rotate_right(R2, 16)
   MOV R0, R1, ROR #8        ; R0 = R1 rotated right 8 bits
   ```
   - Rotate amount: 1-31
   - Bits rotate from LSB to MSB
   - In encoding, ROR #0 means RRX (rotate right extended)

7. **Register with Rotate Right Extended (RRX)**
   ```arm
   MOV R0, R1, RRX           ; R0 = rotate_right_with_carry(R1)
   ADD R0, R1, R2, RRX       ; R0 = R1 + rotate_right_extended(R2)
   ```
   - Encoded as ROR #0 in machine code
   - 33-bit rotate through carry flag
   - Bit 0 → carry, carry → bit 31
   - Useful for multi-precision shifts

8. **Register-specified Shift**
   ```arm
   MOV R0, R1, LSL R2        ; R0 = R1 << R2
   ADD R3, R4, R5, LSR R6    ; R3 = R4 + (R5 >> R6)
   SUB R0, R1, R2, ASR R3    ; R0 = R1 - (R2 >> R3, signed)
   ```
   - Shift amount taken from bottom 8 bits of register
   - If shift amount is 0, no shift performed
   - If shift amount >= 32, behavior depends on shift type

### Memory Addressing Modes
Memory access instructions (LDR, STR, LDRB, STRB, LDRH, STRH) support these addressing modes:

#### Addressing Mode Summary Table

| Mode | Syntax | Address Calculation | Base Register Update | Example |
|------|--------|---------------------|---------------------|---------|
| Register Indirect | `[Rn]` | address = Rn | No | `LDR R0, [R1]` |
| Pre-indexed Immediate | `[Rn, #±offset]` | address = Rn ± offset | No | `LDR R0, [R1, #4]` |
| Pre-indexed Register | `[Rn, ±Rm]` | address = Rn ± Rm | No | `LDR R0, [R1, R2]` |
| Pre-indexed Scaled | `[Rn, ±Rm, shift]` | address = Rn ± (Rm shift amount) | No | `LDR R0, [R1, R2, LSL #2]` |
| Pre-indexed Immediate Writeback | `[Rn, #±offset]!` | address = Rn ± offset | Rn = Rn ± offset | `LDR R0, [R1, #4]!` |
| Pre-indexed Register Writeback | `[Rn, ±Rm]!` | address = Rn ± Rm | Rn = Rn ± Rm | `LDR R0, [R1, R2]!` |
| Pre-indexed Scaled Writeback | `[Rn, ±Rm, shift]!` | address = Rn ± (Rm shift amount) | Rn = address | `LDR R0, [R1, R2, LSL #2]!` |
| Post-indexed Immediate | `[Rn], #±offset` | address = Rn | Rn = Rn ± offset | `LDR R0, [R1], #4` |
| Post-indexed Register | `[Rn], ±Rm` | address = Rn | Rn = Rn ± Rm | `LDR R0, [R1], R2` |
| Post-indexed Scaled | `[Rn], ±Rm, shift` | address = Rn | Rn = Rn ± (Rm shift amount) | `LDR R0, [R1], R2, LSL #2` |

#### Detailed Mode Descriptions

1. **Register Indirect**
   ```arm
   LDR R0, [R1]              ; R0 = memory[R1]
   STR R0, [R1]              ; memory[R1] = R0
   ```
   - Simplest mode: address is in base register
   - Base register unchanged

2. **Pre-indexed with Immediate Offset**
   ```arm
   LDR R0, [R1, #4]          ; R0 = memory[R1 + 4], R1 unchanged
   LDR R0, [R1, #-8]         ; R0 = memory[R1 - 8], R1 unchanged
   STR R2, [R3, #12]         ; memory[R3 + 12] = R2, R3 unchanged
   ```
   - Offset range: -4095 to +4095
   - Address = base + offset
   - Base register unchanged
   - **Use case:** Accessing struct fields at fixed offsets

3. **Pre-indexed with Register Offset**
   ```arm
   LDR R0, [R1, R2]          ; R0 = memory[R1 + R2], R1 unchanged
   LDR R0, [R1, -R2]         ; R0 = memory[R1 - R2], R1 unchanged
   STR R0, [R1, R2]          ; memory[R1 + R2] = R0, R1 unchanged
   ```
   - Address = base ± register
   - Base register unchanged
   - **Use case:** Variable offset access

4. **Pre-indexed with Scaled Register Offset**
   ```arm
   LDR R0, [R1, R2, LSL #2]  ; R0 = memory[R1 + (R2 << 2)], R1 unchanged
   STR R3, [R4, R5, LSL #2]  ; memory[R4 + (R5 << 2)] = R3, R4 unchanged
   LDR R6, [R7, R8, LSR #1]  ; R6 = memory[R7 + (R8 >> 1)], R7 unchanged
   LDRH R0, [R1, R2, LSL #1] ; Load halfword from array[R2]
   ```
   - Address = base ± (register shifted by amount)
   - Available shifts: LSL, LSR, ASR, ROR
   - Shift amount: 0-31
   - Base register unchanged
   - **Use case:** Efficient array indexing without separate multiply
   - **Note:** `LDR R0, [R1, R2, LSL #2]` accesses word array[R2] when R1 points to array base

5. **Pre-indexed with Immediate Writeback** (the `!` addressing)
   ```arm
   LDR R0, [R1, #4]!         ; R1 = R1 + 4, then R0 = memory[R1]
   LDR R0, [R1, #-4]!        ; R1 = R1 - 4, then R0 = memory[R1]
   STR R0, [R1, #8]!         ; R1 = R1 + 8, then memory[R1] = R0
   ```
   - Address = base + offset
   - **Base register updated** to effective address
   - **Use case:** Pre-increment/decrement pointer
   - **Important:** The `!` suffix is what triggers the writeback

6. **Pre-indexed Register Writeback** (the `!` addressing with register)
   ```arm
   LDR R0, [R1, R2]!         ; R1 = R1 + R2, then R0 = memory[R1]
   LDR R0, [R1, -R2]!        ; R1 = R1 - R2, then R0 = memory[R1]
   ```
   - Address = base ± register
   - **Base register updated** to effective address
   - **Use case:** Variable increment pointer update

7. **Pre-indexed Scaled Writeback** (the `!` addressing with scaled register)
   ```arm
   LDR R0, [R1, R2, LSL #2]! ; R1 = R1 + (R2 << 2), then R0 = memory[R1]
   STR R0, [R1, R2, LSL #1]! ; R1 = R1 + (R2 << 1), then memory[R1] = R0
   ```
   - Address = base ± (register shifted by amount)
   - **Base register updated** to effective address
   - **Use case:** Stepping through arrays with element size scaling

8. **Post-indexed with Immediate**
   ```arm
   LDR R0, [R1], #4          ; R0 = memory[R1], then R1 = R1 + 4
   LDR R0, [R1], #-4         ; R0 = memory[R1], then R1 = R1 - 4
   STR R0, [R1], #8          ; memory[R1] = R0, then R1 = R1 + 8
   ```
   - Address = base (current value)
   - **Base register updated** after access: base = base + offset
   - **Use case:** Iterate through array, pointer moves after each access
   - **Common pattern:** Sequential memory access in loops

9. **Post-indexed with Register**
   ```arm
   LDR R0, [R1], R2          ; R0 = memory[R1], then R1 = R1 + R2
   STR R0, [R1], R2          ; memory[R1] = R0, then R1 = R1 + R2
   LDR R0, [R1], -R2         ; R0 = memory[R1], then R1 = R1 - R2
   ```
   - Address = base (current value)
   - **Base register updated** after access: base = base ± register
   - **Use case:** Variable increment iteration

10. **Post-indexed with Scaled Register**
    ```arm
    LDR R0, [R1], R2, LSL #2  ; R0 = memory[R1], then R1 = R1 + (R2 << 2)
    STR R0, [R1], R2, LSL #1  ; memory[R1] = R0, then R1 = R1 + (R2 << 1)
    ```
    - Address = base (current value)
    - **Base register updated** after access: base = base ± (register shifted)
    - **Use case:** Complex iteration patterns

#### Understanding Writeback (`!`)

The **writeback** feature updates the base register with the effective address:

**Pre-indexed writeback:**
```arm
LDR R0, [R1, #4]!    ; Equivalent to: R1 = R1 + 4; R0 = memory[R1]
```

**Post-indexed (implicit writeback):**
```arm
LDR R0, [R1], #4     ; Equivalent to: R0 = memory[R1]; R1 = R1 + 4
```

**Key difference:**
- **Pre-indexed with `!`**: Update base, *then* use updated address
- **Post-indexed**: Use current base, *then* update base

**Common use cases:**
- **Stack operations:** `STR R0, [SP, #-4]!` (pre-decrement)
- **Array iteration:** `LDR R0, [R1], #4` (post-increment)
- **Linked list traversal:** `LDR R0, [R1]!` (follow pointer)

#### Addressing Mode Examples

**Struct field access:**
```arm
; struct { int x; int y; int z; } point;
; R0 = &point
LDR R1, [R0, #0]     ; R1 = point.x
LDR R2, [R0, #4]     ; R2 = point.y
LDR R3, [R0, #8]     ; R3 = point.z
```

**Array access:**
```arm
; int array[10];
; R0 = &array[0], R1 = index
LDR R2, [R0, R1, LSL #2]   ; R2 = array[index]
```

**Iteration with post-increment:**
```arm
; Copy 10 words from src to dst
; R0 = src, R1 = dst, R2 = count
loop:
    LDR R3, [R0], #4       ; Load and increment src
    STR R3, [R1], #4       ; Store and increment dst
    SUBS R2, R2, #1
    BNE loop
```

**Stack push (pre-decrement):**
```arm
STR R0, [SP, #-4]!     ; Push R0: SP = SP - 4, memory[SP] = R0
```

**Stack pop (post-increment):**
```arm
LDR R0, [SP], #4       ; Pop R0: R0 = memory[SP], SP = SP + 4
```

---

## Shift Operations

All shift operations are available in data processing instructions.
### LSL - Logical Shift Left
**Operation:** Shifts bits left, fills with zeros

**Special Cases:**
- LSL #0: No shift performed (identity operation)

**Example:** `MOV R0, R1, LSL #4` shifts R1 left by 4 bits

### LSR - Logical Shift Right
**Operation:** Shifts bits right, fills with zeros

**Special Cases:**
- LSR #0: In ARM encoding, this means LSR #32 (all bits shifted out, result is 0)
- LSR #32: All bits shifted out, result is 0, carry flag = bit 31 of original value

**Example:** `ADD R0, R1, R2, LSR #8` adds R1 + (R2 >> 8)

### ASR - Arithmetic Shift Right
**Operation:** Shifts bits right, preserves sign bit (fills with bit 31)

**Special Cases:**
- ASR #0: In ARM encoding, this means ASR #32 (sign bit extended across all positions)
- ASR #32: Result is 0 if positive, -1 (0xFFFFFFFF) if negative

**Example:** `MOV R0, R1, ASR #2` performs signed divide by 4

### ROR - Rotate Right
**Operation:** Rotates bits right, wraps around

**Special Cases:**
- ROR #0: In ARM encoding, this means RRX (rotate right extended through carry)

**Example:** `ORR R0, R1, R2, ROR #16` rotates R2 by 16 bits

### RRX - Rotate Right Extended
**Operation:** Rotates right by 1 bit through carry flag (33-bit rotation with carry)

**Details:**
- Encoded as ROR #0 in ARM instruction format
- Bit 0 goes to carry flag
- Carry flag goes to bit 31
- Useful for multi-precision shifts

**Example:** `MOV R0, R1, RRX` rotates R1 right through carry

### Register-Specified Shifts
**Operation:** Shift amount specified in register (bottom 8 bits used)

**Details:**
- Only the bottom 8 bits of the register are used for shift amount
- If shift amount is 0, no shift is performed
- If shift amount >= 32, result depends on shift type (LSL/LSR: 0, ASR: sign-extended)

**Example:** `MOV R0, R1, LSL R2` shifts R1 left by amount in R2

---

## CPSR Flags

The Current Program Status Register (CPSR) contains condition flags.
### Flag Overview

| Flag | Bit | Name | Description |
|------|-----|------|-------------|
| N | 31 | Negative | Set when result is negative (bit 31 = 1) |
| Z | 30 | Zero | Set when result is zero |
| C | 29 | Carry | Set on unsigned overflow (addition) or no borrow (subtraction) |
| V | 28 | Overflow | Set on signed overflow |

### Detailed Flag Descriptions

#### N Flag (Negative)
**When SET (N=1):**
- The result has bit 31 set (sign bit = 1)
- For signed operations, the result is negative
- Example: After `SUBS R0, R0, R1` where R0 < R1, N=1

**When CLEAR (N=0):**
- The result has bit 31 clear (sign bit = 0)
- For signed operations, the result is positive or zero
- Example: After `ADDS R0, R1, R2` with positive result, N=0

**Usage:**
- Use with signed comparisons (LT, GE, GT, LE)
- Test with MI (minus/negative) and PL (plus/positive) conditions

#### Z Flag (Zero)
**When SET (Z=1):**
- The result is exactly zero (all 32 bits are 0)
- Equality condition is true
- Example: After `SUBS R0, R1, R1`, Z=1 (result is 0)

**When CLEAR (Z=0):**
- The result is non-zero
- Equality condition is false
- Example: After `ADDS R0, R1, R2` with non-zero result, Z=0

**Usage:**
- Use with EQ (equal) and NE (not equal) conditions
- Test after CMP for equality checks
- Essential for loop termination conditions

#### C Flag (Carry)
**When SET (C=1):**

**For Addition (ADD, ADC, CMN):**
- Unsigned overflow occurred (result > 0xFFFFFFFF)
- The addition produced a carry out of bit 31
- Example: `ADDS R0, R1, R2` where R1=0xFFFFFFFF, R2=1 sets C=1

**For Subtraction (SUB, SBC, CMP, RSB, RSC):**
- **NO borrow occurred** (result >= 0 in unsigned terms)
- The subtraction did NOT require a borrow
- Example: `SUBS R0, R1, R2` where R1 >= R2 sets C=1
- **Important:** C=1 means "no borrow", C=0 means "borrow occurred"

**For Shifts:**
- Contains the last bit shifted out
- Example: `MOVS R0, R1, LSR #1` puts bit 0 of R1 into C

**When CLEAR (C=0):**

**For Addition:**
- No unsigned overflow (result <= 0xFFFFFFFF)

**For Subtraction:**
- Borrow occurred (unsigned underflow)
- Example: `SUBS R0, R1, R2` where R1 < R2 sets C=0

**Usage:**
- Use with unsigned comparisons (HI, LS, HS/CS, LO/CC)
- Use ADC/SBC for multi-precision arithmetic
- Test with CS/HS (carry set) and CC/LO (carry clear) conditions

#### V Flag (Overflow)
**When SET (V=1):**
- Signed overflow occurred
- The result cannot be represented in 32-bit two's complement
- The sign bit changed incorrectly

**Addition overflow occurs when:**
- Adding two positive numbers yields negative result
- Adding two negative numbers yields positive result
- Example: `ADDS R0, R1, R2` where R1=0x7FFFFFFF, R2=1 sets V=1
  (0x7FFFFFFF + 1 = 0x80000000, positive + positive = negative)

**Subtraction overflow occurs when:**
- Subtracting negative from positive yields negative result
- Subtracting positive from negative yields positive result
- Example: `SUBS R0, R1, R2` where R1=0x80000000, R2=1 sets V=1
  (0x80000000 - 1 = 0x7FFFFFFF, negative - positive = positive)

**When CLEAR (V=0):**
- No signed overflow occurred
- The result is valid in two's complement

**Usage:**
- Use with signed comparisons (LT, GE, GT, LE)
- Test with VS (overflow set) and VC (overflow clear) conditions
- Essential for detecting overflow in signed arithmetic

### Flag Update Rules

**Arithmetic Operations (ADD, ADC, SUB, SBC, RSB, RSC):**
- Update all four flags: N, Z, C, V
- C flag: Carry out for addition, NOT borrow for subtraction
- V flag: Signed overflow detection
- S suffix required (ADDS, SUBS, etc.)

**Logical Operations (AND, ORR, EOR, BIC, MOV, MVN):**
- Update N, Z, C only (V unaffected)
- N: Set from bit 31 of result
- Z: Set if result is zero
- C: Set from shifter carry out (if shift applied)
- S suffix required (ANDS, MOVS, etc.)

**Comparison Operations (CMP, CMN, TST, TEQ):**
- Always update flags (no S suffix needed)
- CMP/CMN: Update N, Z, C, V (like SUB/ADD)
- TST/TEQ: Update N, Z, C (like AND/EOR)

**Multiply Operations (MUL, MLA):**
- Update N, Z only when S suffix used
- C flag meaningless after multiply
- V flag unaffected
- Example: `MULS R0, R1, R2`

**Long Multiply Operations (UMULL, SMULL, etc.):**
- Update N, Z only when S suffix used
- C, V flags unaffected

### Carry Flag in Subtraction - Important Note

The carry flag behaves **inversely** for subtraction compared to addition:

```arm
; Addition: C=1 means carry occurred
ADDS R0, R1, R2     ; If overflow: C=1, else C=0

; Subtraction: C=1 means NO borrow (result >= 0)
SUBS R0, R1, R2     ; If R1 >= R2: C=1 (no borrow)
                    ; If R1 < R2:  C=0 (borrow occurred)

; Comparison (performs subtraction)
CMP R0, R1          ; Same as SUBS, but result discarded
                    ; If R0 >= R1: C=1
                    ; If R0 < R1:  C=0
```

This is why:
- `BHS` (branch if higher or same) checks C=1
- `BLO` (branch if lower) checks C=0

### Multi-Precision Arithmetic Using Flags

**64-bit Addition:**
```arm
; Add R1:R0 + R3:R2 -> R5:R4
ADDS R4, R0, R2     ; Low word, sets carry
ADC  R5, R1, R3     ; High word + carry
```

**64-bit Subtraction:**
```arm
; Subtract R1:R0 - R3:R2 -> R5:R4
SUBS R4, R0, R2     ; Low word, sets borrow in C
SBC  R5, R1, R3     ; High word - borrow
```

### Flag Usage Examples

**Signed comparison:**
```arm
CMP R0, #10         ; Compare R0 with 10
; If R0 = 5:  N=1 (negative result), Z=0, C=0 (borrow), V=0
; If R0 = 10: N=0, Z=1 (zero result), C=1 (no borrow), V=0
; If R0 = 15: N=0, Z=0, C=1 (no borrow), V=0

BGT greater         ; Branch if R0 > 10 (Z=0 AND N=V)
BLT less            ; Branch if R0 < 10 (N != V)
BEQ equal           ; Branch if R0 = 10 (Z=1)
```

**Unsigned comparison:**
```arm
CMP R0, #100        ; Compare R0 with 100 (unsigned)
BHI higher          ; Branch if R0 > 100 (C=1 AND Z=0)
BLO lower           ; Branch if R0 < 100 (C=0)
BHS higher_same     ; Branch if R0 >= 100 (C=1)
BLS lower_same      ; Branch if R0 <= 100 (C=0 OR Z=1)
```

**Testing bits:**
```arm
TST R0, #0x01       ; Test if bit 0 is set
; If bit 0 set: Z=0
; If bit 0 clear: Z=1
BNE bit_is_set      ; Branch if Z=0 (bit was set)
BEQ bit_is_clear    ; Branch if Z=1 (bit was clear)
```

**Overflow detection:**
```arm
MOV R0, #0x7FFFFFFF ; Max positive 32-bit signed int
ADDS R0, R0, #1     ; Add 1
; Result: R0 = 0x80000000 (looks negative)
; Flags: N=1 (negative), V=1 (overflow occurred)
BVS overflow_handler ; Branch if overflow (V=1)
```

### Reading and Writing Flags

**Reading CPSR:**
```arm
MRS R0, CPSR        ; Read CPSR into R0
; Bit 31 = N flag
; Bit 30 = Z flag
; Bit 29 = C flag
; Bit 28 = V flag
```

**Writing flags:**
```arm
MSR CPSR_f, R0      ; Write R0 to CPSR flags field
MSR CPSR_f, #0xF0000000  ; Set all flags (N=1, Z=1, C=1, V=1)
MSR CPSR_f, #0x00000000  ; Clear all flags
```

---

## Pseudo-Instructions

Pseudo-instructions are assembler conveniences that map to real instructions.

### ADR - Load PC-Relative Address

**Syntax:** `ADR{cond} Rd, label`

**Description:** Loads a PC-relative address into a register (pseudo-instruction that generates ADD or SUB)

**Operation:** `Rd = PC + offset` (generates ADD or SUB instruction based on offset sign)

**Range:** Offset must be encodable as an ARM immediate value

**Example:**
```arm
ADR R0, message       ; R0 = address of message
ADR R1, data_table    ; R1 = address of data_table
ADR R2, function      ; R2 = address of function
```

**Note:** This is a true pseudo-instruction. The assembler converts it to `ADD Rd, PC, #offset` or `SUB Rd, PC, #offset` based on whether the offset is positive or negative.

### Other Pseudo-Instructions

| Pseudo | Real Instruction | Description |
|--------|------------------|-------------|
| NOP | MOV R0, R0 | No operation |
| LDR Rd, =value | LDR Rd, [PC, #offset] or MOV/MVN | Load 32-bit constant |
| PUSH {regs} | STMDB SP!, {regs} | Push registers to stack |
| POP {regs} | LDMIA SP!, {regs} | Pop registers from stack |

**NOP Example:**
```arm
NOP                   ; No operation (encoded as MOV R0, R0)
NOP                   ; Used for timing, alignment, or placeholders
```

**LDR Rd, =value Example:**
```arm
LDR R0, =0x12345678   ; Load 32-bit constant into R0
LDR R1, =message      ; Load address of message label
LDR R2, =0xFF         ; Small values may use MOV R2, #0xFF
```

**LDR Rd, =value Details:**
The assembler intelligently chooses the most efficient encoding:
1. If the value fits in an ARM immediate (8-bit rotated), uses `MOV Rd, #value`
2. If ~value fits in an ARM immediate, uses `MVN Rd, #~value`
3. Otherwise, places the value in a literal pool and generates `LDR Rd, [PC, #offset]`

**Literal Pool Management:**
- Use `.ltorg` directive to place literal pool at specific locations
- Literal pool must be within ±4095 bytes of the LDR instruction
- Values are automatically deduplicated in the pool
- Multiple `.ltorg` directives can be used for large programs

**Literal Pool Example:**
```arm
.text
.org 0x8000

main:
    LDR R0, =0x12345678
    LDR R1, =0xDEADBEEF
    ADD R2, R0, R1
    MOV R0, #0
    SWI #0x00

    .ltorg              ; Place literal pool here
```

**PUSH Example:**
```arm
PUSH {R0-R3, LR}      ; Push R0-R3 and LR to stack
PUSH {R4-R11}         ; Push R4-R11 to stack
```

**POP Example:**
```arm
POP {R0-R3, PC}       ; Pop R0-R3 and return (PC=LR)
POP {R4-R11}          ; Pop R4-R11 from stack
```

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

- **Assembler Features:**
  - Full directive support: `.text`, `.data`, `.bss`, `.global`, `.equ`, `.set`
  - Memory allocation: `.word`, `.half`, `.byte`, `.ascii`, `.asciz`, `.space`
  - Alignment directives: `.align`, `.balign`
  - Character literals with escape sequences
  - Immediate value support in multiple formats (decimal, hex, binary)

- **Development Tools:**
  - Assembly Linter (tools/lint.go) - Code analysis with 25 tests
  - Code Formatter (tools/format.go) - Professional formatting with 27 tests
  - Cross-Reference Generator (tools/xref.go) - Symbol analysis with 21 tests

- **Memory Alignment:** Word accesses should be 4-byte aligned, halfword 2-byte aligned
