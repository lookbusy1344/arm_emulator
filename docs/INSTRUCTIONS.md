# ARM2 Instruction Set Reference

This document details the ARM2 CPU instructions supported by this emulator.

**Related Documentation:**
- [Assembler Directives](ASSEMBLER.md) - Directives for organizing code and data (.text, .data, .word, .ltorg, etc.)
- [Programming Reference](REFERENCE.md) - Condition codes, addressing modes, shifts, and conventions

---

## Table of Contents
1. [CPSR Flags](#cpsr-flags)
2. [Data Processing Instructions](#data-processing-instructions)
3. [Memory Access Instructions](#memory-access-instructions)
4. [Branch Instructions](#branch-instructions)
5. [Multiply Instructions](#multiply-instructions)
6. [System Instructions](#system-instructions)
7. [Unsupported Instructions](#unsupported-instructions)

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

## Unsupported Instructions

The following ARM instructions are intentionally not supported in this ARM2 emulator:

### Coprocessor Instructions (Not Implemented)

The ARM2 architecture included coprocessor support for floating-point units (FPA) and other external processors, but this emulator does not implement coprocessor functionality. The following instructions are not supported:

| Mnemonic | Description |
|----------|-------------|
| CDP | Coprocessor Data Processing |
| MCR | Move ARM register to Coprocessor register |
| MRC | Move Coprocessor register to ARM register |
| LDC | Load Coprocessor register from memory |
| STC | Store Coprocessor register to memory |

**Rationale:** Coprocessor support was optional in the ARM2 architecture and was primarily used for floating-point operations (FPA10, FPA11). This emulator focuses on the core ARM2 integer instruction set. Programs requiring floating-point operations should implement software floating-point routines or use fixed-point arithmetic.

**Behavior:** Attempting to execute a coprocessor instruction will halt the VM with the error: "coprocessor instructions not supported".

---
