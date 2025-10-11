; bit_operations.s - Comprehensive bit manipulation tests
; Demonstrates: Bitwise operations, shifts, rotations, bit counting

        .org    0x8000

_start:
        LDR     R0, =msg_intro
        SWI     #0x02
        SWI     #0x07
        SWI     #0x07

        ; Test 1: Count bits in a number
        LDR     R0, =msg_test1
        SWI     #0x02
        LDR     R0, =0xF0F0F0F0
        BL      count_bits
        MOV     R5, R0
        LDR     R0, =msg_bits
        SWI     #0x02
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03
        LDR     R0, =msg_expect16
        SWI     #0x02
        SWI     #0x07

        ; Test 2: Reverse bits in a byte
        LDR     R0, =msg_test2
        SWI     #0x02
        MOV     R0, #0b10110010
        BL      reverse_byte
        MOV     R5, R0
        LDR     R0, =msg_reversed
        SWI     #0x02
        MOV     R0, R5
        MOV     R1, #2          ; Binary output
        SWI     #0x03
        LDR     R0, =msg_expect_rev
        SWI     #0x02
        SWI     #0x07

        ; Test 3: Check if power of 2
        LDR     R0, =msg_test3
        SWI     #0x02
        MOV     R0, #64
        BL      is_power_of_2
        CMP     R0, #1
        BEQ     test3_pass
        LDR     R0, =msg_fail
        SWI     #0x02
        B       test3_done
test3_pass:
        LDR     R0, =msg_pass
        SWI     #0x02
test3_done:
        SWI     #0x07

        ; Test 4: Find first set bit (LSB)
        LDR     R0, =msg_test4
        SWI     #0x02
        MOV     R0, #0b10110000
        BL      find_first_set
        MOV     R5, R0
        LDR     R0, =msg_position
        SWI     #0x02
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03
        LDR     R0, =msg_expect4
        SWI     #0x02
        SWI     #0x07

        ; Test 5: Rotate left
        LDR     R0, =msg_test5
        SWI     #0x02
        MOV     R0, #0b00000001
        MOV     R1, #4
        BL      rotate_left
        MOV     R5, R0
        LDR     R0, =msg_result
        SWI     #0x02
        MOV     R0, R5
        MOV     R1, #2
        SWI     #0x03
        LDR     R0, =msg_expect16_bin
        SWI     #0x02
        SWI     #0x07

        ; Test 6: Extract bit field
        LDR     R0, =msg_test6
        SWI     #0x02
        LDR     R0, =0x12345678
        MOV     R1, #8          ; Start bit
        MOV     R2, #8          ; Length
        BL      extract_bits
        MOV     R5, R0
        LDR     R0, =msg_extracted
        SWI     #0x02
        MOV     R0, R5
        MOV     R1, #16
        SWI     #0x03
        SWI     #0x07

        ; Done message
        BL      print_done

        MOV     R0, #0
        SWI     #0x00

; Helper to print done message
print_done:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg_done
        SWI     #0x02
        SWI     #0x07
        LDMFD   SP!, {R0, PC}

; count_bits - Count number of 1 bits in a word
; Input: R0 = value
; Output: R0 = number of 1 bits
count_bits:
        STMFD   SP!, {R1, R2, LR}
        MOV     R1, R0          ; Value to check
        MOV     R0, #0          ; Bit count
        MOV     R2, #32         ; Number of bits to check

count_loop:
        CMP     R2, #0
        BEQ     count_done
        TST     R1, #1          ; Test LSB
        ADDNE   R0, R0, #1      ; Increment if set
        MOV     R1, R1, LSR #1  ; Shift right
        SUB     R2, R2, #1
        B       count_loop

count_done:
        LDMFD   SP!, {R1, R2, PC}

; reverse_byte - Reverse bits in a byte
; Input: R0 = byte (only lower 8 bits used)
; Output: R0 = reversed byte
reverse_byte:
        STMFD   SP!, {R1-R3, LR}
        MOV     R1, R0          ; Input
        MOV     R0, #0          ; Result
        MOV     R2, #8          ; Bit counter

reverse_loop:
        CMP     R2, #0
        BEQ     reverse_done
        MOV     R0, R0, LSL #1  ; Shift result left
        TST     R1, #1          ; Test LSB of input
        ORRNE   R0, R0, #1      ; Set LSB of result if set
        MOV     R1, R1, LSR #1  ; Shift input right
        SUB     R2, R2, #1
        B       reverse_loop

reverse_done:
        LDMFD   SP!, {R1-R3, PC}

; is_power_of_2 - Check if number is power of 2
; Input: R0 = number
; Output: R0 = 1 if power of 2, 0 otherwise
is_power_of_2:
        STMFD   SP!, {R1, LR}
        CMP     R0, #0
        MOVEQ   R0, #0          ; 0 is not power of 2
        BEQ     power2_done

        MOV     R1, R0
        SUB     R1, R1, #1      ; n-1
        AND     R1, R1, R0      ; n & (n-1)
        CMP     R1, #0
        MOVEQ   R0, #1          ; Is power of 2
        MOVNE   R0, #0          ; Not power of 2

power2_done:
        LDMFD   SP!, {R1, PC}

; find_first_set - Find position of first set bit (LSB = 0)
; Input: R0 = value
; Output: R0 = position (0-31), or 32 if no bits set
find_first_set:
        STMFD   SP!, {R1, R2, LR}
        MOV     R1, R0          ; Value
        MOV     R0, #0          ; Position counter
        MOV     R2, #32         ; Max iterations

ffs_loop:
        CMP     R2, #0
        BEQ     ffs_not_found
        TST     R1, #1
        BNE     ffs_done
        MOV     R1, R1, LSR #1
        ADD     R0, R0, #1
        SUB     R2, R2, #1
        B       ffs_loop

ffs_not_found:
        MOV     R0, #32

ffs_done:
        LDMFD   SP!, {R1, R2, PC}

; rotate_left - Rotate bits left
; Input: R0 = value, R1 = positions
; Output: R0 = rotated value (only lower 8 bits)
rotate_left:
        STMFD   SP!, {R1-R3, LR}
        AND     R0, R0, #0xFF   ; Keep only 8 bits
        AND     R1, R1, #7      ; Modulo 8
        MOV     R2, #8
        SUB     R2, R2, R1      ; 8 - positions = right rotate amount

        MOV     R3, R0, LSL R1  ; Shift left
        AND     R3, R3, #0xFF   ; Mask to 8 bits
        MOV     R0, R0, LSR R2  ; Shift right (wraparound bits)
        ORR     R0, R3, R0      ; Combine

        LDMFD   SP!, {R1-R3, PC}

; extract_bits - Extract bit field from word
; Input: R0 = value, R1 = start bit, R2 = length
; Output: R0 = extracted bits
extract_bits:
        STMFD   SP!, {R1-R3, LR}
        MOV     R3, R0          ; Save value
        MOV     R0, R3, LSR R1  ; Shift right by start
        MOV     R3, #1
        MOV     R3, R3, LSL R2  ; 1 << length
        SUB     R3, R3, #1      ; (1 << length) - 1 = mask
        AND     R0, R0, R3      ; Apply mask

        LDMFD   SP!, {R1-R3, PC}

msg_intro:
        .asciz  "Bit Operations Test Suite"
msg_test1:
        .asciz  "Test 1: Count bits in 0xF0F0F0F0"
msg_bits:
        .asciz  "  Bits set: "
msg_expect16:
        .asciz  " (expected 16)"
msg_test2:
        .asciz  "Test 2: Reverse byte 0b10110010"
msg_reversed:
        .asciz  "  Reversed: 0b"
msg_expect_rev:
        .asciz  " (expected 0b01001101)"
msg_test3:
        .asciz  "Test 3: Is 64 a power of 2? "
msg_pass:
        .asciz  "PASS"
msg_fail:
        .asciz  "FAIL"
msg_test4:
        .asciz  "Test 4: Find first set bit in 0b10110000"
msg_position:
        .asciz  "  Position: "
msg_expect4:
        .asciz  " (expected 4)"
msg_test5:
        .asciz  "Test 5: Rotate 0b00000001 left by 4"
msg_result:
        .asciz  "  Result: 0b"
msg_expect16_bin:
        .asciz  " (expected 0b00010000)"
msg_test6:
        .asciz  "Test 6: Extract bits [15:8] from 0x12345678"
msg_extracted:
        .asciz  "  Extracted: 0x"
msg_expect56:
        .asciz  " (expected 0x56)"
msg_done:
        .asciz  "All bit operation tests completed!"
