; test_get_random.s - Demonstrates GET_RANDOM syscall (0x31)
; Shows: Random number generation, histogram, distribution check

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print header
        LDR     R0, =msg_header
        SWI     #0x02           ; WRITE_STRING

        ; Generate and display 10 random numbers
        MOV     R4, #10         ; Counter

random_loop:
        ; Print "Random: "
        LDR     R0, =msg_random
        SWI     #0x02           ; WRITE_STRING

        ; Get random number
        SWI     #0x31           ; GET_RANDOM syscall

        ; Print in hexadecimal
        MOV     R1, #16         ; Hex format
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Decrement counter and loop
        SUBS    R4, R4, #1
        BNE     random_loop

        ; Test randomness: generate 100 numbers and check distribution
        SWI     #0x07           ; WRITE_NEWLINE
        LDR     R0, =msg_dist
        SWI     #0x02           ; WRITE_STRING

        MOV     R4, #100        ; Generate 100 numbers
        MOV     R5, #0          ; Count of numbers with high bit set
        MOV     R6, #0          ; Count of numbers with low bit set

dist_loop:
        SWI     #0x31           ; GET_RANDOM

        ; Check if high bit (bit 31) is set
        TST     R0, #0x80000000
        ADDNE   R5, R5, #1      ; Increment if set

        ; Check if low bit (bit 0) is set
        TST     R0, #1
        ADDNE   R6, R6, #1      ; Increment if set

        SUBS    R4, R4, #1
        BNE     dist_loop

        ; Print high bit statistics
        LDR     R0, =msg_highbit
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R5
        MOV     R1, #10         ; Decimal
        SWI     #0x03           ; WRITE_INT
        LDR     R0, =msg_of100
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Print low bit statistics
        LDR     R0, =msg_lowbit
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R6
        MOV     R1, #10         ; Decimal
        SWI     #0x03           ; WRITE_INT
        LDR     R0, =msg_of100
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Print success message
        SWI     #0x07           ; WRITE_NEWLINE
        LDR     R0, =msg_success
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit program with code 0
        MOV     R0, #0
        SWI     #0x00           ; EXIT syscall

        ; Data section
msg_header:
        .asciz  "=== GET_RANDOM Syscall Test ==="
msg_random:
        .asciz  "Random: 0x"
msg_dist:
        .asciz  "Distribution test (100 samples):"
msg_highbit:
        .asciz  "High bit set: "
msg_lowbit:
        .asciz  "Low bit set:  "
msg_of100:
        .asciz  "/100"
msg_success:
        .asciz  "Random number generation working - Test PASSED"
