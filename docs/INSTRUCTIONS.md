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

**Description:** Performs 32-bit integer addition of two operands and stores the result in the destination register.
This is the fundamental arithmetic operation used for incrementing counters, calculating addresses, and general arithmetic.
When the S suffix is used, it updates the CPSR flags enabling conditional execution based on the result.

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

**Description:** Performs addition of two operands plus the current value of the carry flag (C bit in CPSR).
Essential for multi-precision arithmetic operations where results exceed 32 bits, such as 64-bit or 128-bit integer addition.
The carry flag from a previous addition is propagated to add the upper words correctly.

**Operation:** `Rd = Rn + operand2 + C`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
ADC R0, R1, R2        ; R0 = R1 + R2 + C
ADCS R3, R3, #0       ; R3 = R3 + C, update flags (for multi-precision)
```

#### SUB - Subtract

**Syntax:** `SUB{cond}{S} Rd, Rn, <operand2>`

**Description:** Performs 32-bit integer subtraction by subtracting the second operand from the first operand.
Used for decrementing counters, calculating differences, and comparing values when combined with conditional execution.
The carry flag is set when the subtraction does not require a borrow (i.e., when Rn >= operand2 for unsigned comparison).

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

**Description:** Performs subtraction with borrow, subtracting both operand2 and the inverse of the carry flag from Rn.
Used in multi-precision subtraction to propagate borrows between words, similar to how ADC propagates carries in addition.
The NOT(C) in the operation means a borrow is subtracted when C=0 (borrow present) and nothing when C=1 (no borrow).

**Operation:** `Rd = Rn - operand2 - NOT(C)`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
SBC R0, R1, R2        ; R0 = R1 - R2 - NOT(C)
SBCS R3, R3, #0       ; For multi-precision subtraction
```

#### RSB - Reverse Subtract

**Syntax:** `RSB{cond}{S} Rd, Rn, <operand2>`

**Description:** Performs subtraction in reverse order, subtracting Rn from operand2 instead of the usual order.
Particularly useful for negating values (e.g., RSB Rd, Rn, #0 computes -Rn) and calculating constants minus variables.
This instruction eliminates the need for separate negation or constant loading in many algorithms.

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

**Description:** Performs reverse subtraction with borrow, computing operand2 minus Rn minus the borrow (NOT C).
Used in multi-precision reverse subtraction operations where the order of operands needs to be reversed.
Combines the reverse order of RSB with the borrow propagation of SBC for extended precision calculations.

**Operation:** `Rd = operand2 - Rn - NOT(C)`

**Flags:** Updates N, Z, C, V when S bit is set

**Example:**
```arm
RSC R0, R1, R2        ; R0 = R2 - R1 - NOT(C)
```

### Logical Operations

#### AND - Logical AND

**Syntax:** `AND{cond}{S} Rd, Rn, <operand2>`

**Description:** Performs bitwise AND operation between two operands, setting result bits to 1 only where both input bits are 1.
Commonly used for masking operations to isolate specific bits, clearing unwanted bits, and testing bit patterns.
Essential for bit manipulation, flag checking, and extracting bit fields from packed data structures.

**Operation:** `Rd = Rn AND operand2`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
AND R0, R1, R2        ; R0 = R1 & R2
ANDS R3, R3, #0xFF    ; R3 = R3 & 0xFF, update flags
```

#### ORR - Logical OR

**Syntax:** `ORR{cond}{S} Rd, Rn, <operand2>`

**Description:** Performs bitwise OR operation between two operands, setting result bits to 1 where either or both input bits are 1.
Used for setting specific bits to 1 while preserving others, combining bit flags, and merging bit fields.
Common in hardware control registers, flag manipulation, and building composite values from multiple sources.

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

**Description:** Performs bitwise XOR operation, setting result bits to 1 only where input bits differ (one is 1, the other is 0).
Used for toggling specific bits, comparing bit patterns, and implementing simple encryption or checksums.
The idiom EOR Rd, Rn, Rn is a fast way to zero a register (commonly seen as EORS R3, R3, R3 to clear R3 and set Z flag).

**Operation:** `Rd = Rn EOR operand2`

**Flags:** Updates N, Z, C when S bit is set (V unaffected)

**Example:**
```arm
EOR R0, R1, R2        ; R0 = R1 ^ R2
EORS R3, R3, R3       ; R3 = 0, update flags
```

#### BIC - Bit Clear

**Syntax:** `BIC{cond}{S} Rd, Rn, <operand2>`

**Description:** Clears (sets to 0) specific bits in Rn by ANDing with the complement of operand2.
Each bit position where operand2 has a 1 will be cleared in the result; bits where operand2 has 0 are preserved.
Commonly used for clearing flag bits in control registers and removing specific bits from a value without affecting others.

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

**Description:** Copies a value from the source operand into the destination register, performing no arithmetic or logical operation.
Used for register-to-register transfers, loading immediate constants, and applying shift operations to values.
The special case MOVS PC, LR is used to return from subroutines while restoring the CPSR flags from SPSR.

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

**Description:** Loads the bitwise complement (NOT) of the source operand into the destination register, inverting all bits.
Useful for creating inverted bit patterns, generating all-ones values (MVN Rd, #0 yields 0xFFFFFFFF), and complementing masks.
Often used when the desired immediate constant is more easily expressed in inverted form due to ARM encoding limitations.

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

**Description:** Compares two values by performing subtraction (Rn - operand2) and setting the condition flags based on the result, but discarding the actual result.
Essential for conditional execution, typically followed by conditional branch or conditional data processing instructions.
The flags enable both signed comparisons (using N, V flags) and unsigned comparisons (using C, Z flags).

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

**Description:** Compares a register with the negative of operand2 by performing addition (Rn + operand2) and updating flags without storing the result.
Equivalent to comparing Rn with -operand2, useful when you want to test if a register equals the negation of a value.
Less commonly used than CMP but valuable when working with negated constants or when addition flags are needed for comparison.

**Operation:** `Rn + operand2` (result discarded)

**Flags:** Always updates N, Z, C, V

**Example:**
```arm
CMN R0, R1            ; Compare R0 with -R1
CMN R2, #-5           ; Test if R2 is 5
```

#### TST - Test Bits

**Syntax:** `TST{cond} Rn, <operand2>`

**Description:** Tests whether specific bits are set by performing a bitwise AND and updating flags based on the result without storing it.
Commonly used to check if one or more bits are set in a value, test for zero, or verify bit masks.
Sets the Z flag if the AND result is zero (no bits in common), making it ideal for bit testing in conditional logic.

**Operation:** `Rn AND operand2` (result discarded)

**Flags:** Always updates N, Z, C (V unaffected)

**Example:**
```arm
TST R0, #0x01         ; Test if bit 0 is set
TST R1, R2            ; Test bits in common
```

#### TEQ - Test Equivalence

**Syntax:** `TEQ{cond} Rn, <operand2>`

**Description:** Tests whether two values are equal by performing exclusive OR and setting flags based on the result without storing it.
Sets the Z flag if the values are identical (XOR yields zero), making it useful for equality testing without affecting the C or V flags.
Preferred over CMP for equality tests when you need to preserve carry and overflow flags for subsequent operations.

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

**Description:** Loads a 32-bit word (4 bytes) from memory into the destination register, performing a read from the computed address.
Supports flexible addressing modes including base register with offset, pre-indexed with writeback, and post-indexed with automatic base register update.
Essential for accessing variables, array elements, and data structures, with the address automatically aligned to 4-byte boundaries in most implementations.

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

**Description:** Stores a 32-bit word (4 bytes) from the source register to memory at the computed address, performing a write operation.
Like LDR, supports various addressing modes with optional base register update for efficient array traversal and pointer manipulation.
Used for writing variables, storing array elements, and updating data structures in memory with automatic address alignment.

**Operation:** `Memory[address] = Rd`

**Example:**
```arm
STR R0, [R1]          ; [R1] = R0
STR R2, [R3, #-4]     ; [R3 - 4] = R2
STR R4, [R5, R6, LSL #2]!  ; [R5 + (R6 << 2)] = R4, writeback
```

#### LDRB - Load Byte

**Syntax:** `LDRB{cond} Rd, <addressing_mode>`

**Description:** Loads a single 8-bit byte from memory and zero-extends it to 32 bits in the destination register (upper 24 bits set to 0).
Used for accessing character data, byte arrays, and packed structures where each element is one byte.
No alignment restrictions apply - bytes can be loaded from any memory address, making this instruction essential for string and byte buffer operations.

**Operation:** `Rd = ZeroExtend(Memory[address])`

**Example:**
```arm
LDRB R0, [R1]         ; R0 = byte at [R1]
LDRB R2, [R3, #1]     ; R2 = byte at [R3 + 1]
```

#### STRB - Store Byte

**Syntax:** `STRB{cond} Rd, <addressing_mode>`

**Description:** Stores the least significant byte (bits 7:0) of the source register to memory, discarding the upper 24 bits.
Used for writing character data, updating byte arrays, and modifying individual bytes in packed data structures.
Can write to any memory address without alignment requirements, making it perfect for string manipulation and byte-oriented I/O operations.

**Operation:** `Memory[address] = Rd[7:0]`

**Example:**
```arm
STRB R0, [R1]         ; [R1] = R0[7:0]
STRB R2, [R3, #10]    ; [R3 + 10] = R2[7:0]
```

#### LDRH - Load Halfword
**Syntax:** `LDRH{cond} Rd, <addressing_mode>`

**Description:** Loads a 16-bit halfword (2 bytes) from memory and zero-extends it to 32 bits in the destination register (upper 16 bits set to 0).
Used for accessing 16-bit data types such as Unicode characters, short integers, and halfword arrays.
Memory address should typically be 2-byte aligned for optimal performance, though some implementations allow unaligned access.

**Operation:** `Rd = ZeroExtend(Memory[address])`

**Example:**
```arm
LDRH R0, [R1]         ; R0 = halfword at [R1]
LDRH R2, [R3, #2]     ; R2 = halfword at [R3 + 2]
```

#### STRH - Store Halfword
**Syntax:** `STRH{cond} Rd, <addressing_mode>`

**Description:** Stores the least significant halfword (bits 15:0) of the source register to memory as 2 bytes, discarding the upper 16 bits.
Used for writing 16-bit data types, Unicode characters, and short integer values to memory.
Like LDRH, addresses should be 2-byte aligned for best performance in accessing 16-bit data structures and arrays.

**Operation:** `Memory[address] = Rd[15:0]`

**Example:**
```arm
STRH R0, [R1]         ; [R1] = R0[15:0]
STRH R2, [R3, #6]     ; [R3 + 6] = R2[15:0]
```

### Multiple Data Transfer

#### LDM - Load Multiple

**Syntax:** `LDM{cond}{mode} Rn{!}, {register_list}{^}`

**Description:** Efficiently loads multiple registers from consecutive 32-bit memory locations in a single instruction, starting from the base address.
Primarily used for function returns, context restoration, and bulk data loading from memory with automatic address incrementing or decrementing.
The optional writeback (!) updates the base register, and the caret (^) suffix enables CPSR restoration for exception returns when PC is in the list.

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

**Description:** Efficiently stores multiple registers to consecutive 32-bit memory locations in a single instruction, starting from the base address.
Primarily used for function entry (saving registers to stack), context saving, and bulk data storage with automatic address management.
The writeback (!) automatically updates the base register, making STM ideal for push operations when combined with appropriate addressing modes.

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

**Description:** Performs an unconditional or conditional branch to a target address specified by a label, changing program flow by updating the PC.
The most fundamental control flow instruction, used for loops, if-then-else logic, and jumping to different code sections.
Supports all condition codes for conditional execution, enabling efficient implementation of complex control structures without separate compare and jump instructions.

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

**Description:** Calls a subroutine by branching to the target label while saving the return address (PC+4) in the link register (LR).
The fundamental instruction for function calls, enabling modular code with automatic return address management.
Functions return by moving LR back to PC (typically using MOV PC, LR or BX LR), and nested calls require saving LR to the stack first.

**Operation:** `LR = PC + 4, PC = PC + offset`

**Range:** ±32MB from current instruction

**Example:**
```arm
BL function           ; Call function
BLEQ conditional_fn   ; Call if equal
```

#### BX - Branch and Exchange

**Syntax:** `BX{cond} Rm`

**Description:** Branches to the address contained in a register, enabling computed jumps and register-based returns from subroutines.
Originally designed for ARM/Thumb interworking (bit 0 of Rm indicates mode), but in ARM2 emulation simply branches to the register value.
The standard way to return from functions (BX LR) and implement jump tables or function pointers for dynamic dispatch.

**Operation:** `PC = Rm & 0xFFFFFFFE` (bit 0 would indicate Thumb mode in later ARM)

**Example:**
```arm
BX LR                 ; Return from subroutine
BX R0                 ; Branch to address in R0
```

#### BLX - Branch with Link and Exchange

**Syntax:** `BLX{cond} Rm`

**Description:** Calls a function at the address in a register while saving the return address to LR, combining BL's link behavior with BX's register branching.
Enables indirect function calls through function pointers, virtual method dispatch, and callback mechanisms where the target address is computed at runtime.
Essential for implementing dynamic dispatch, plugin architectures, and any scenario requiring computed subroutine calls rather than compile-time fixed addresses.

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

**Description:** Performs 32-bit integer multiplication of two operands, storing only the lower 32 bits of the 64-bit result in the destination register.
Used for basic multiplication where overflow beyond 32 bits is acceptable or expected to be handled separately.
Note that Rd and Rm must be different registers due to ARM2 architectural restrictions, and execution time varies (2-16 cycles) based on the multiplier value.

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

**Description:** Multiplies two 32-bit values and adds a third value (accumulator) to the result, storing the lower 32 bits in the destination register.
Particularly efficient for dot products, matrix operations, and polynomial evaluation where multiply-add patterns are common.
Like MUL, Rd and Rm must differ, and the instruction provides significant performance benefits over separate multiply and add operations.

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

**Description:** Performs unsigned multiplication of two 32-bit values, producing a full 64-bit result stored in two registers (RdHi:RdLo).
Essential for multi-precision arithmetic, large integer calculations, and situations where the full product of two numbers must be preserved.
The lower 32 bits go to RdLo and upper 32 bits to RdHi, with all three output registers (RdLo, RdHi, Rm) required to be distinct.

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

**Description:** Performs unsigned multiplication of two 32-bit values and adds the 64-bit result to the existing 64-bit accumulator in RdHi:RdLo.
Used for multi-precision multiply-accumulate operations, implementing 128-bit or larger arithmetic, and accumulating products in computational algorithms.
Particularly valuable in cryptographic operations, numerical computing, and any algorithm requiring extended precision accumulation of multiple products.

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

**Description:** Performs signed (two's complement) multiplication of two 32-bit values, producing a full 64-bit signed result in two registers.
Essential for signed multi-precision arithmetic where negative numbers must be handled correctly, preserving sign extension into the upper word.
Used in fixed-point arithmetic, financial calculations, and any signed numerical computation requiring the complete product without overflow.

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

**Description:** Performs signed multiplication of two 32-bit values and adds the signed 64-bit product to the existing signed 64-bit accumulator.
Combines signed long multiplication with accumulation for algorithms requiring repeated signed multiply-add operations with extended precision.
Critical for digital signal processing, matrix operations with signed values, and numerical algorithms needing signed multi-precision accumulation.

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

**Description:** Generates a software interrupt to invoke system services, transferring control to the OS or emulator's syscall handler with the immediate value indicating the requested service.
This is the ARM2 method for performing I/O, memory allocation, file operations, and other privileged operations that require OS intervention.
The syscall number is encoded directly in the instruction's immediate field (0-16777215), with arguments passed in registers R0-R2 and results returned in R0.

**Operation:** Transfers control to system call handler

**Example:**
```arm
SWI #0x00             ; Exit program
SWI #0x02             ; Write string to console
SWI #0x11             ; Write character
```

#### System Call Numbers (SWI)

##### Console I/O (0x00-0x07)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x00 | EXIT | Exit program | R0: exit code | - |
| 0x01 | WRITE_CHAR | Write character to stdout | R0: character | - |
| 0x02 | WRITE_STRING | Write null-terminated string | R0: string address | - |
| 0x03 | WRITE_INT | Write integer in specified base | R0: value, R1: base (2/8/10/16, default 10) | - |
| 0x04 | READ_CHAR | Read character from stdin (skips whitespace) | - | R0: character or 0xFFFFFFFF on error |
| 0x05 | READ_STRING | Read string from stdin (until newline) | R0: buffer address, R1: max length (default 256) | R0: bytes written or 0xFFFFFFFF on error |
| 0x06 | READ_INT | Read integer from stdin | - | R0: integer value or 0 on error |
| 0x07 | WRITE_NEWLINE | Write newline to stdout | - | - |

##### File Operations (0x10-0x16)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x10 | OPEN | Open file | R0: filename address, R1: mode (0=read, 1=write, 2=append) | R0: file descriptor or 0xFFFFFFFF on error |
| 0x11 | CLOSE | Close file | R0: file descriptor | R0: 0 on success, 0xFFFFFFFF on error |
| 0x12 | READ | Read from file | R0: fd, R1: buffer address, R2: length | R0: bytes read or 0xFFFFFFFF on error |
| 0x13 | WRITE | Write to file | R0: fd, R1: buffer address, R2: length | R0: bytes written or 0xFFFFFFFF on error |
| 0x14 | SEEK | Seek in file | R0: fd, R1: offset, R2: whence (0=start, 1=current, 2=end) | R0: new position or 0xFFFFFFFF on error |
| 0x15 | TELL | Get current file position | R0: file descriptor | R0: position or 0xFFFFFFFF on error |
| 0x16 | FILE_SIZE | Get file size | R0: file descriptor | R0: size or 0xFFFFFFFF on error |

##### Memory Operations (0x20-0x22)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x20 | ALLOCATE | Allocate memory from heap | R0: size in bytes | R0: address or 0 (NULL) on failure |
| 0x21 | FREE | Free allocated memory | R0: address | R0: 0 on success, 0xFFFFFFFF on error |
| 0x22 | REALLOCATE | Resize memory allocation | R0: old address, R1: new size | R0: new address or 0 (NULL) on failure |

##### System Information (0x30-0x33)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x30 | GET_TIME | Get time in milliseconds since Unix epoch | - | R0: timestamp (lower 32 bits) |
| 0x31 | GET_RANDOM | Get random 32-bit number | - | R0: random value |
| 0x32 | GET_ARGUMENTS | Get program arguments | - | R0: argc, R1: argv pointer (0 in current impl) |
| 0x33 | GET_ENVIRONMENT | Get environment variables | - | R0: envp pointer (0 in current impl) |

##### Error Handling (0x40-0x42)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x40 | GET_ERROR | Get last error code | - | R0: error code (0 in current impl) |
| 0x41 | SET_ERROR | Set error code | R0: error code | - |
| 0x42 | PRINT_ERROR | Print error message to stderr | R0: error code | - |

##### Debugging Support (0xF0-0xF4)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0xF0 | DEBUG_PRINT | Print debug message to stderr | R0: string address | - |
| 0xF1 | BREAKPOINT | Trigger debugger breakpoint | - | - |
| 0xF2 | DUMP_REGISTERS | Print all registers to stdout | - | - |
| 0xF3 | DUMP_MEMORY | Dump memory region as hex dump | R0: address, R1: length (max 1KB) | - |
| 0xF4 | ASSERT | Assert condition is true | R0: condition (0=fail), R1: message address | Halts if condition is 0 |

**Note:** CPSR flags (N, Z, C, V) are preserved across all syscalls to prevent unintended side effects on conditional logic.
### MRS - Move PSR to Register
**Syntax:** `MRS{cond} Rd, PSR`

**Description:** Reads the Current Program Status Register (CPSR) or Saved Program Status Register (SPSR) and copies all 32 bits into the destination register.
Used to examine processor flags (N, Z, C, V) and other status bits, typically before modifying them or for context preservation in interrupt handlers.
Essential for implementing atomic operations, critical sections, and any code that needs to inspect or preserve the processor state.

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

**Description:** Writes to specific fields of the CPSR or SPSR from a register or immediate value, allowing direct manipulation of processor status flags.
The _f (flags) field specifier controls which bits are updated, typically targeting the condition flags (N, Z, C, V) in bits 31-24.
Critical for restoring saved processor state, manually setting flags for testing, implementing context switches, and controlling processor modes in operating systems.

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
**Description:** Marks the beginning of a code section containing executable instructions, directing the assembler to place subsequent statements in the executable program area.
Essential for organizing assembly programs by separating executable code from data, allowing multiple code sections to be interleaved with data sections.
The first `.text` directive typically starts at address 0 unless overridden with `.org`, and multiple text sections can appear throughout the source file.

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
**Description:** Marks the beginning of a data section for defining initialized variables, constants, strings, and arrays that will be stored in memory.
Used to organize program data separately from executable code, making the assembly source more readable and maintainable.
Like `.text`, multiple data sections can be scattered throughout the source file and will be assembled sequentially based on their position and any `.org` directives.

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
**Description:** Declares a symbol as globally visible, making it accessible from other object modules during linking in multi-file programs.
Commonly used to export entry points (like `_start` or `main`), public functions, and shared data that need to be referenced from other compilation units.
While this emulator treats all symbols as visible (single-module execution), the directive is accepted for compatibility with standard ARM assembly conventions.

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
**Description:** Defines a named constant that associates a symbol with a numeric value, functioning like `#define` in C but processed by the assembler.
Improves code readability and maintainability by replacing magic numbers with meaningful names, supporting decimal, hexadecimal, binary, and character literal values.
Constants can reference other previously defined constants, and negative values are supported, making it ideal for configuration values, bit flags, and memory addresses.

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
**Description:** Sets the current memory address where the assembler will place subsequent instructions and data, effectively controlling the memory layout.
Used to position code at specific addresses (like interrupt vectors), create memory gaps, or organize data at predetermined locations in memory.
Can be used multiple times to create non-contiguous sections, with the first `.org` establishing the program's entry point when no explicit origin is set.

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
**Description:** Allocates and initializes one or more 32-bit words (4 bytes each) in memory with specified values, supporting multiple data formats.
The fundamental data allocation directive for integers, pointers, addresses, and lookup tables in ARM assembly.
Values are stored in little-endian format and can be specified as decimal, hexadecimal, binary, character literals, or label addresses.

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
**Description:** Allocates and initializes one or more 16-bit halfwords (2 bytes each) in memory, useful for 16-bit data types and space-efficient storage.
Commonly used for Unicode characters (UTF-16), short integers, packed data structures, and graphics data like RGB565 color values.
Values larger than 16 bits are truncated, and data is stored in little-endian format like all ARM data.

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
**Description:** Allocates and initializes one or more 8-bit bytes (1 byte each) in memory, providing the finest granularity for data storage.
Essential for ASCII characters, byte arrays, status flags, bitmap data, and any raw byte sequences that need explicit control.
Values larger than 8 bits are truncated to their lower 8 bits, making this directive perfect for packed data and character manipulation.

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
**Description:** Allocates a string in memory as a sequence of bytes without appending a null terminator, storing the exact characters specified.
Used when you need precise control over string length, when concatenating multiple strings, or when the null terminator will be added manually.
Supports standard escape sequences (\n, \t, \\, etc.) for special characters, with each character stored as a single ASCII byte.

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
**Description:** Allocates a null-terminated C-style string in memory by storing the specified characters followed by an automatic null byte (0x00).
The standard way to define strings for system calls, C library functions, and any code expecting null-terminated strings.
Both `.asciz` and `.string` are equivalent and interchangeable, with full support for escape sequences for special characters.

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
**Description:** Reserves a specified number of bytes in memory, initializing all bytes to zero, creating uninitialized buffers and arrays.
Both `.space` and `.skip` are equivalent and commonly used for allocating working memory, input/output buffers, stack space, and large data structures.
Size can be specified as immediate values or constants defined with `.equ`, making it easy to allocate memory based on configuration parameters.

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
**Description:** Aligns the current memory address to a power-of-2 byte boundary (2^n), padding with zero bytes as needed to reach the alignment.
Critical for performance optimization and hardware requirements, ensuring data structures start at addresses compatible with load/store instructions and cache line boundaries.
Common uses include word alignment (`.align 2` for 4-byte), double-word alignment (`.align 3` for 8-byte), and cache line alignment (`.align 4` for 16-byte).

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
**Description:** Aligns the current memory address to an exact byte boundary specified directly (not as a power of 2), padding with zeros as necessary.
More intuitive than `.align` when you want specific byte alignments like 4, 8, or 16 bytes without calculating powers of 2.
Commonly used for word alignment (4 bytes), double-word alignment (8 bytes), and cache line optimization (16 bytes) with straightforward numeric values.

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
**Description:** Explicitly places a literal pool at the current location, storing 32-bit constants used by `LDR Rd, =value` pseudo-instructions within the ±4095 byte range constraint.
Essential for large programs where automatic literal pool placement at the end might exceed the addressing range, requiring manual pool placement near the instructions that reference them.
The assembler automatically deduplicates identical constants, aligns the pool to 4-byte boundaries, and reserves the necessary space for all accumulated literals.

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

All instructions can be conditionally executed based on CPSR flags. This is one of ARM's most powerful features - nearly every instruction can include a condition code suffix that determines whether it executes based on the current flag values.

### Condition Code Table

| Code | Suffix | Description | Condition | Typical Use Case |
|------|--------|-------------|-----------|------------------|
| 0000 | EQ | Equal | Z == 1 | After comparison, values are equal |
| 0001 | NE | Not Equal | Z == 0 | After comparison, values differ |
| 0010 | CS/HS | Carry Set / Unsigned Higher or Same | C == 1 | Unsigned comparisons: A >= B |
| 0011 | CC/LO | Carry Clear / Unsigned Lower | C == 0 | Unsigned comparisons: A < B |
| 0100 | MI | Minus / Negative | N == 1 | After operation, result is negative |
| 0101 | PL | Plus / Positive or Zero | N == 0 | After operation, result is non-negative |
| 0110 | VS | Overflow Set | V == 1 | Signed arithmetic overflow occurred |
| 0111 | VC | Overflow Clear | V == 0 | No signed arithmetic overflow |
| 1000 | HI | Unsigned Higher | C == 1 AND Z == 0 | Unsigned comparisons: A > B |
| 1001 | LS | Unsigned Lower or Same | C == 0 OR Z == 1 | Unsigned comparisons: A <= B |
| 1010 | GE | Signed Greater or Equal | N == V | Signed comparisons: A >= B |
| 1011 | LT | Signed Less Than | N != V | Signed comparisons: A < B |
| 1100 | GT | Signed Greater Than | Z == 0 AND N == V | Signed comparisons: A > B |
| 1101 | LE | Signed Less or Equal | Z == 1 OR N != V | Signed comparisons: A <= B |
| 1110 | AL | Always (default) | Always true | No condition (default if omitted) |
| 1111 | NV | Never (deprecated) | Never true | Deprecated, do not use |

### Understanding Signed vs Unsigned Conditions

**Signed Conditions (GE, LT, GT, LE):**
- Use after comparing signed integers (values that can be negative)
- Check N and V flags to determine signed relationship
- Example: Comparing -5 and 10, or testing if counter < limit

**Unsigned Conditions (HS/CS, LO/CC, HI, LS):**
- Use after comparing unsigned integers (values >= 0)
- Check C and Z flags to determine unsigned relationship
- Example: Comparing addresses, array indices, or absolute quantities

**Equality Conditions (EQ, NE):**
- Work for both signed and unsigned
- Only check Z flag
- Example: Testing if value equals specific number

### Practical Examples

#### Example 1: Comparison with Conditional Branch

**Basic equality check:**
```arm
    ; Check if R0 equals 10
    CMP R0, #10         ; Compare R0 with 10 (sets Z flag)
    BEQ is_equal        ; Branch if Z==1 (R0 == 10)
    BNE not_equal       ; Branch if Z==0 (R0 != 10)

is_equal:
    MOV R1, #1          ; R1 = 1 (true)
    B continue

not_equal:
    MOV R1, #0          ; R1 = 0 (false)

continue:
    ; Continue with R1 set
```

**Signed comparison (less than):**
```arm
    ; Check if R0 < R1 (signed)
    CMP R0, R1          ; Compare R0 with R1
    BLT r0_less         ; Branch if R0 < R1 (signed)
    BGE r0_greater_eq   ; Branch if R0 >= R1 (signed)

r0_less:
    ; Handle case where R0 < R1
    MOV R2, R0          ; Use smaller value
    B done

r0_greater_eq:
    ; Handle case where R0 >= R1
    MOV R2, R1          ; Use R1

done:
    ; R2 contains appropriate value
```

**Signed comparison (greater than):**
```arm
    ; Find maximum of R0 and R1 (signed)
    CMP R0, R1          ; Compare R0 with R1
    BGT r0_is_bigger    ; Branch if R0 > R1 (signed)
    MOV R2, R1          ; R1 is max (R0 <= R1)
    B continue

r0_is_bigger:
    MOV R2, R0          ; R0 is max

continue:
    ; R2 now contains max(R0, R1)
```

**Unsigned comparison (array bounds check):**
```arm
    ; Check if index R0 < array_size (unsigned)
    LDR R1, =array_size
    LDR R1, [R1]        ; R1 = array size
    CMP R0, R1          ; Compare index with size
    BLO valid_index     ; Branch if R0 < R1 (unsigned)

    ; Invalid index - handle error
    MOV R0, #0xFFFFFFFF ; Return error code
    B exit

valid_index:
    ; Access array[R0]
    LDR R2, =array_base
    LDR R2, [R2, R0, LSL #2]  ; Load array[R0]

exit:
```

**Loop with counter:**
```arm
    ; Loop 10 times using counter
    MOV R0, #0          ; Initialize counter
loop_start:
    ; Loop body here
    ADD R1, R1, R0      ; Example operation

    ; Increment and test
    ADD R0, R0, #1      ; counter++
    CMP R0, #10         ; Compare with limit
    BLT loop_start      ; Continue if counter < 10
    BEQ loop_start      ; Alternative: continue if counter == 10
    BGE loop_done       ; Exit when counter >= 10

loop_done:
    ; R0 = 10, loop complete
```

**Range check (value between min and max):**
```arm
    ; Check if MIN <= R0 <= MAX
    CMP R0, #MIN        ; Compare R0 with MIN
    BLT out_of_range    ; Branch if R0 < MIN

    CMP R0, #MAX        ; Compare R0 with MAX
    BGT out_of_range    ; Branch if R0 > MAX

    ; R0 is in valid range [MIN, MAX]
    MOV R1, #1          ; Valid
    B continue

out_of_range:
    MOV R1, #0          ; Invalid

continue:
```

#### Example 2: Comparison with Conditional Data Processing

**Conditional arithmetic:**
```arm
    ; Add bonus only if score >= 100
    LDR R0, =score
    LDR R0, [R0]        ; Load score
    CMP R0, #100        ; Compare with threshold
    ADDGE R0, R0, #10   ; Add bonus if score >= 100
    STR R0, [R1]        ; Store updated score
```

**Conditional absolute value:**
```arm
    ; R0 = abs(R0)
    CMP R0, #0          ; Compare with zero
    RSBLT R0, R0, #0    ; If negative, R0 = 0 - R0 (negate)
    ; If R0 was positive or zero, RSBLT doesn't execute
```

**Conditional clamping (limit value to range):**
```arm
    ; Clamp R0 to range [0, 255]
    CMP R0, #0          ; Compare with minimum
    MOVLT R0, #0        ; If R0 < 0, set to 0

    CMP R0, #255        ; Compare with maximum
    MOVGT R0, #255      ; If R0 > 255, set to 255
    ; R0 now guaranteed to be in [0, 255]
```

**Conditional max/min without branches:**
```arm
    ; R2 = max(R0, R1) using conditional moves
    CMP R0, R1          ; Compare R0 and R1
    MOVGT R2, R0        ; If R0 > R1, R2 = R0
    MOVLE R2, R1        ; If R0 <= R1, R2 = R1

    ; R3 = min(R0, R1)
    CMP R0, R1          ; Compare R0 and R1
    MOVLT R3, R0        ; If R0 < R1, R3 = R0
    MOVGE R3, R1        ; If R0 >= R1, R3 = R1
```

**Conditional increment/decrement:**
```arm
    ; Increment R0 only if R1 != 0
    CMP R1, #0          ; Test R1
    ADDNE R0, R0, #1    ; Increment if R1 != 0

    ; Decrement counter but don't go below zero
    CMP R2, #0          ; Test if counter > 0
    SUBGT R2, R2, #1    ; Decrement only if > 0
```

**Sign extension:**
```arm
    ; Sign-extend byte in R0 to 32-bit
    AND R0, R0, #0xFF   ; Ensure only lower 8 bits
    TST R0, #0x80       ; Test sign bit (bit 7)
    ORRNE R0, R0, #0xFFFFFF00  ; If negative, extend sign
```

#### Example 3: Comparison with Conditional Memory Access

**Conditional store:**
```arm
    ; Store result only if computation succeeded (R1 == 0)
    BL compute_value    ; Call function, returns status in R1
    CMP R1, #0          ; Check status
    STREQ R0, [R2]      ; Store R0 only if status == 0 (success)
```

**Conditional load for optional value:**
```arm
    ; Load configuration if flag is set
    LDR R0, =config_flag
    LDR R0, [R0]        ; Load flag value
    CMP R0, #1          ; Check if enabled
    LDREQ R1, =config_value
    LDREQ R1, [R1]      ; Load config only if flag == 1
```

**Array update with condition:**
```arm
    ; Update array[i] only if new_value > old_value
    LDR R0, =array_base
    MOV R1, #5          ; Index
    LDR R2, [R0, R1, LSL #2]  ; Load array[5]
    LDR R3, =new_value
    LDR R3, [R3]        ; Load new value

    CMP R3, R2          ; Compare new with old
    STRGT R3, [R0, R1, LSL #2]  ; Store if new > old
```

**Conditional pointer dereference:**
```arm
    ; Load value from pointer only if pointer is non-null
    LDR R0, =data_ptr
    LDR R0, [R0]        ; Load pointer
    CMP R0, #0          ; Check if NULL
    LDRNE R1, [R0]      ; Dereference only if not NULL
    MOVEQ R1, #0        ; Use default value if NULL
```

**Table lookup with bounds check:**
```arm
    ; Safe table lookup with conditional load
    LDR R0, =index      ; Load index
    LDR R0, [R0]
    CMP R0, #TABLE_SIZE ; Check bounds
    LDRLO R1, =table    ; Load table base if valid
    LDRLO R1, [R1, R0, LSL #2]  ; Load table[index]
    MOVHS R1, #0        ; Return 0 if out of bounds
```

#### Example 4: Multi-Condition Logic

**Nested conditions:**
```arm
    ; if (x > 0 && x < 100)
    LDR R0, =x
    LDR R0, [R0]        ; Load x

    CMP R0, #0          ; Check x > 0
    BLE else_case       ; Branch if x <= 0

    CMP R0, #100        ; Check x < 100
    BGE else_case       ; Branch if x >= 100

    ; Both conditions true: 0 < x < 100
    MOV R1, #1
    B done

else_case:
    MOV R1, #0

done:
```

**Multiple comparisons in sequence:**
```arm
    ; Grade assignment based on score
    ; A: >= 90, B: >= 80, C: >= 70, D: >= 60, F: < 60
    LDR R0, =score
    LDR R0, [R0]        ; Load score

    CMP R0, #90
    BGE grade_A

    CMP R0, #80
    BGE grade_B

    CMP R0, #70
    BGE grade_C

    CMP R0, #60
    BGE grade_D

grade_F:
    MOV R1, #'F'
    B done

grade_D:
    MOV R1, #'D'
    B done

grade_C:
    MOV R1, #'C'
    B done

grade_B:
    MOV R1, #'B'
    B done

grade_A:
    MOV R1, #'A'

done:
    ; R1 contains grade letter
```

**Switch/case pattern:**
```arm
    ; Switch on value in R0
    LDR R0, =command    ; Load command value
    LDR R0, [R0]

    CMP R0, #1
    BEQ case_1

    CMP R0, #2
    BEQ case_2

    CMP R0, #3
    BEQ case_3

    ; Default case
    B case_default

case_1:
    MOV R1, #10
    B end_switch

case_2:
    MOV R1, #20
    B end_switch

case_3:
    MOV R1, #30
    B end_switch

case_default:
    MOV R1, #0

end_switch:
```

#### Example 5: Loop Patterns with Conditions

**While loop (test at top):**
```arm
    ; while (count > 0) { ... }
    LDR R0, =count
    LDR R0, [R0]        ; Load count

while_loop:
    CMP R0, #0          ; Test condition
    BLE while_done      ; Exit if count <= 0

    ; Loop body
    ; ... process ...

    SUB R0, R0, #1      ; count--
    B while_loop

while_done:
```

**Do-while loop (test at bottom):**
```arm
    ; do { ... } while (count > 0);
    LDR R0, =count
    LDR R0, [R0]        ; Load count

do_loop:
    ; Loop body executes at least once
    ; ... process ...

    SUBS R0, R0, #1     ; count-- and set flags
    BGT do_loop         ; Continue if count > 0
```

**For loop:**
```arm
    ; for (i = 0; i < 10; i++)
    MOV R0, #0          ; i = 0

for_loop:
    CMP R0, #10         ; Test i < 10
    BGE for_done        ; Exit if i >= 10

    ; Loop body
    ; ... use R0 as index ...

    ADD R0, R0, #1      ; i++
    B for_loop

for_done:
```

**Countdown loop with early exit:**
```arm
    ; Search array for zero, stop early if found
    LDR R0, =array
    MOV R1, #ARRAY_SIZE ; Counter

search_loop:
    CMP R1, #0          ; Check if done
    BLE not_found       ; Exit if counter reached 0

    LDR R2, [R0], #4    ; Load value, post-increment pointer
    CMP R2, #0          ; Check if zero
    BEQ found_zero      ; Exit early if found

    SUB R1, R1, #1      ; counter--
    B search_loop

found_zero:
    ; R0 points to element after the zero
    ; R1 contains remaining count
    MOV R3, #1          ; Success
    B done

not_found:
    MOV R3, #0          ; Not found

done:
```

#### Example 6: Conditional Function Calls

**Call function only if condition met:**
```arm
    ; Call error handler only if error occurred
    BL operation        ; Perform operation
    CMP R0, #0          ; Check return value
    BLNE error_handler  ; Call error handler if R0 != 0
```

**Conditional recursive call:**
```arm
factorial:
    ; factorial(n) = n <= 1 ? 1 : n * factorial(n-1)
    CMP R0, #1          ; Check base case
    MOVLE R0, #1        ; If n <= 1, return 1
    MOVLE PC, LR        ; Return

    ; Recursive case
    PUSH {R0, LR}       ; Save n and return address
    SUB R0, R0, #1      ; n - 1
    BL factorial        ; factorial(n-1)
    POP {R1, LR}        ; Restore n to R1
    MUL R0, R1, R0      ; n * factorial(n-1)
    MOV PC, LR          ; Return
```

**Conditional callback:**
```arm
    ; Call callback function pointer if not NULL
    LDR R0, =callback_ptr
    LDR R0, [R0]        ; Load function pointer
    CMP R0, #0          ; Check if NULL
    MOVNE LR, PC        ; Set return address if not NULL
    MOVNE PC, R0        ; Call function if not NULL
    ; Continues here (or after callback returns)
```

### Practical Tips

**1. Prefer conditional execution over branches when possible:**
```arm
    ; Less efficient (branch):
    CMP R0, #0
    BEQ skip
    ADD R1, R1, #1
skip:

    ; More efficient (conditional execution):
    CMP R0, #0
    ADDNE R1, R1, #1    ; No branch, fewer pipeline stalls
```

**2. Remember CMP is just SUBS that discards the result:**
```arm
    CMP R0, R1          ; Same as SUBS (discarded), R0 - R1
    SUBS R2, R0, R1     ; R2 = R0 - R1, sets same flags as CMP
```

**3. TST for bit testing:**
```arm
    ; Check if bit 5 is set
    TST R0, #(1<<5)     ; Test bit 5
    BNE bit_is_set      ; Branch if bit was set (Z=0)
    BEQ bit_is_clear    ; Branch if bit was clear (Z=1)
```

**4. Combine SUBS with conditional branch:**
```arm
    ; Decrement and loop if not zero
    SUBS R0, R0, #1     ; Decrement and set flags
    BNE loop_start      ; Continue if result != 0 (no separate CMP needed)
```

**5. Use appropriate signed/unsigned conditions:**
```arm
    ; For array indices (unsigned):
    CMP R0, R1
    BLO valid           ; Branch if R0 < R1 (unsigned)

    ; For temperatures (signed):
    CMP R0, #0
    BLT below_freezing  ; Branch if R0 < 0 (signed)
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

### Saving and Restoring Flags on the Stack

When calling functions that modify flags, you may need to preserve the current flag state. This is done by reading the CPSR, pushing it to the stack, and later popping and restoring it.

**Basic pattern:**
```arm
; Save flags to stack
MRS R0, CPSR        ; Read current flags into R0
PUSH {R0}           ; Push flags to stack

; ... code that modifies flags ...
ADDS R1, R2, R3     ; This will change N, Z, C, V flags

; Restore flags from stack
POP {R0}            ; Pop saved flags from stack
MSR CPSR_f, R0      ; Restore flags
```

**Function call with flag preservation:**
```arm
calculate:
    ; Function that needs to preserve caller's flags
    MRS R12, CPSR        ; Save flags in R12 (IP register)
    PUSH {R12, LR}       ; Save flags and return address

    ; Function body - modifies flags freely
    CMP R0, #0
    BEQ zero_case
    ADDS R0, R0, R1
    MULS R0, R0, R2

zero_case:
    ; Restore flags before returning
    POP {R12, LR}        ; Restore saved flags and return address
    MSR CPSR_f, R12      ; Restore caller's flags
    MOV PC, LR           ; Return
```

**Nested function calls:**
```arm
outer_function:
    ; Save flags and registers
    MRS R12, CPSR
    PUSH {R4-R6, R12, LR}

    ; Do some work
    MOV R4, R0
    ADDS R5, R1, R2

    ; Call inner function (which preserves flags)
    MOV R0, R4
    BL inner_function

    ; Continue with preserved flags from before BL
    ; (inner_function preserved them)
    ADDS R6, R5, R0

    ; Restore and return
    POP {R4-R6, R12, LR}
    MSR CPSR_f, R12
    MOV PC, LR

inner_function:
    ; This function also preserves caller's flags
    MRS R12, CPSR
    PUSH {R12, LR}

    ; Function body
    CMP R0, #10
    MOVGT R0, #10

    ; Restore and return
    POP {R12, LR}
    MSR CPSR_f, R12
    MOV PC, LR
```

**Critical section example:**
```arm
critical_operation:
    ; Disable interrupts and save flags
    MRS R12, CPSR        ; Save current CPSR
    PUSH {R12}           ; Save to stack

    ; Set flags to disable interrupts (if interrupt bits were in flags)
    MSR CPSR_f, #0xC0000000  ; Set N and Z flags (example)

    ; Critical code here
    LDR R0, =shared_data
    LDR R1, [R0]
    ADD R1, R1, #1
    STR R1, [R0]

    ; Restore original flags (and interrupt state)
    POP {R12}
    MSR CPSR_f, R12
    MOV PC, LR
```

**Why preserve flags?**
- Calling conventions may require preserving condition codes
- Interrupts or function calls in the middle of conditional logic
- Multi-step operations where intermediate flag states matter
- Implementing reentrant or interrupt-safe code

**Important notes:**
- Only the flag bits (N, Z, C, V) in bits 31-28 are typically preserved
- `MSR CPSR_f` writes only the flag field (bits 31-24), leaving other CPSR bits unchanged
- Use R12 (IP) as a temporary register for flags since it's caller-saved
- The stack pattern (PUSH/POP) ensures proper nesting of flag preservation

---

## Pseudo-Instructions

Pseudo-instructions are assembler conveniences that map to real instructions.

### ADR - Load PC-Relative Address

**Syntax:** `ADR{cond} Rd, label`

**Description:** Calculates and loads the address of a label relative to the program counter, providing position-independent address computation.
This pseudo-instruction is assembled into either ADD or SUB with PC as the base register, depending on whether the target is ahead or behind the current instruction.
Essential for position-independent code, accessing nearby data structures, and obtaining addresses of labels without using the literal pool, as long as the offset fits in an ARM immediate value.

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

---

## See Also

- [docs/TUTORIAL.md](docs/TUTORIAL.md) - Learn ARM2 assembly from scratch
- [docs/assembly_reference.md](docs/assembly_reference.md) - Assembler directives and syntax
- [examples/README.md](examples/README.md) - 44 example programs demonstrating all instructions
- [docs/debugger_reference.md](docs/debugger_reference.md) - Debugging commands and features
- [docs/FAQ.md](docs/FAQ.md) - Common questions and troubleshooting
