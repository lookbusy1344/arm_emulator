# ARM2 Programming Reference

This document provides reference material for ARM2 assembly programming including condition codes, addressing modes, and conventions.

**Related Documentation:**
- [Instruction Set Reference](INSTRUCTIONS.md) - ARM2 CPU instructions and system calls
- [Assembler Directives](ASSEMBLER.md) - Directives for organizing code and data (.text, .data, .word, .ltorg, etc.)

---

## Table of Contents
1. [Condition Codes](#condition-codes)
2. [Addressing Modes](#addressing-modes)
3. [Shift Operations](#shift-operations)
4. [Pseudo-Instructions](#pseudo-instructions)
5. [Register Usage Conventions](#register-usage-conventions)
6. [Notes](#notes)
7. [See Also](#see-also)

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
