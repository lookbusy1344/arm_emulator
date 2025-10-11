; Factorial Calculator
; Computes the factorial of a number using recursion
; Example: Input 5 produces 120 (5! = 5*4*3*2*1)

.org 0x0000

main:
    ; Print prompt
    LDR r0, =prompt_msg
    BL print_string

    ; Read input number
    BL read_int
    MOV r4, r0              ; Store input in r4

    ; Check for negative or too large
    CMP r4, #0
    BLT error_negative
    CMP r4, #12             ; Limit to 12! to prevent overflow
    BGT error_too_large

    ; Calculate factorial
    MOV r0, r4
    BL factorial

    ; Store result
    MOV r5, r0

    ; Print result message
    LDR r0, =result_msg
    BL print_string

    MOV r0, r4
    BL print_int

    LDR r0, =factorial_symbol
    BL print_string

    MOV r0, r5
    BL print_int

    LDR r0, =newline
    BL print_string

    ; Exit
    MOV r0, #0
    BL exit

error_negative:
    LDR r0, =error_neg_msg
    BL print_string
    MOV r0, #1
    BL exit

error_too_large:
    LDR r0, =error_large_msg
    BL print_string
    MOV r0, #1
    BL exit

; Recursive factorial function
; Input: r0 = n
; Output: r0 = n!
factorial:
    PUSH {r4, lr}
    MOV r4, r0              ; Save n

    ; Base case: if n <= 1, return 1
    CMP r4, #1
    BLE factorial_base

    ; Recursive case: n * factorial(n-1)
    SUB r0, r4, #1          ; r0 = n - 1
    BL factorial            ; r0 = factorial(n-1)
    MUL r0, r4, r0          ; r0 = n * factorial(n-1)
    POP {r4, pc}

factorial_base:
    MOV r0, #1
    POP {r4, pc}

; Helper functions
print_string:
    PUSH {lr}
    SWI #0x02               ; SWI_WRITE_STRING
    POP {pc}

print_int:
    PUSH {lr}
    SWI #0x03               ; SWI_WRITE_INT
    POP {pc}

read_int:
    PUSH {lr}
    SWI #0x06               ; SWI_READ_INT
    POP {pc}

exit:
    SWI #0x00               ; SWI_EXIT

; Data section
.align 4
prompt_msg:         .asciz "Enter a number (0-12): "
result_msg:         .asciz "Result: "
factorial_symbol:   .asciz "! = "
error_neg_msg:      .asciz "Error: Negative numbers not supported\n"
error_large_msg:    .asciz "Error: Number too large (max 12)\n"
newline:            .asciz "\n"
