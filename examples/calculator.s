; Simple Calculator
; Performs basic arithmetic operations: +, -, *, /
; Example: "15 + 7" produces "22"

.org 0x8000

main:
    ; Print welcome message
    LDR r0, =welcome_msg
    BL print_string

menu_loop:
    ; Print menu
    LDR r0, =menu_msg
    BL print_string

    ; Read first number
    LDR r0, =num1_prompt
    BL print_string
    BL read_int
    MOV r4, r0              ; r4 = first number

    ; Read operation
    LDR r0, =op_prompt
    BL print_string
    BL read_char
    MOV r5, r0              ; r5 = operation

    ; Check for quit
    CMP r5, #'q'
    BEQ exit_program
    CMP r5, #'Q'
    BEQ exit_program

    ; Read second number
    LDR r0, =num2_prompt
    BL print_string
    BL read_int
    MOV r6, r0              ; r6 = second number

    ; Perform operation based on operator
    CMP r5, #'+'
    BEQ do_add
    CMP r5, #'-'
    BEQ do_sub
    CMP r5, #'*'
    BEQ do_mul
    CMP r5, #'/'
    BEQ do_div

    ; Invalid operation
    LDR r0, =error_op_msg
    BL print_string
    B menu_loop

do_add:
    ADD r7, r4, r6
    B print_result

do_sub:
    SUB r7, r4, r6
    B print_result

do_mul:
    MUL r7, r4, r6
    B print_result

do_div:
    ; Check for division by zero
    CMP r6, #0
    BEQ error_div_zero

    ; Perform division (r7 = r4 / r6)
    MOV r0, r4
    MOV r1, r6
    BL divide
    MOV r7, r0              ; r7 = quotient
    MOV r8, r1              ; r8 = remainder

    ; Print result with remainder
    LDR r0, =result_msg
    BL print_string

    MOV r0, r4
    BL print_int

    LDR r0, =space
    BL print_string
    MOV r0, r5
    BL print_char
    LDR r0, =space
    BL print_string

    MOV r0, r6
    BL print_int

    LDR r0, =equals
    BL print_string

    MOV r0, r7
    BL print_int

    ; Print remainder if non-zero
    CMP r8, #0
    BEQ skip_remainder

    LDR r0, =remainder_msg
    BL print_string
    MOV r0, r8
    BL print_int

skip_remainder:
    LDR r0, =newline
    BL print_string
    B menu_loop

print_result:
    ; Print: num1 op num2 = result
    LDR r0, =result_msg
    BL print_string

    MOV r0, r4
    BL print_int

    LDR r0, =space
    BL print_string
    MOV r0, r5
    BL print_char
    LDR r0, =space
    BL print_string

    MOV r0, r6
    BL print_int

    LDR r0, =equals
    BL print_string

    MOV r0, r7
    BL print_int

    LDR r0, =newline
    BL print_string
    B menu_loop

error_div_zero:
    LDR r0, =error_div_msg
    BL print_string
    B menu_loop

exit_program:
    LDR r0, =goodbye_msg
    BL print_string
    MOV r0, #0
    BL exit

; Integer division function
; Input: r0 = dividend, r1 = divisor
; Output: r0 = quotient, r1 = remainder
divide:
    PUSH {r4, r5}
    MOV r4, #0              ; r4 = quotient
    MOV r5, r0              ; r5 = remainder

    ; Handle negative numbers (simplified - just use absolute values)
    CMP r5, #0
    RSBLT r5, r5, #0

    CMP r1, #0
    RSBLT r1, r1, #0

div_loop:
    CMP r5, r1
    BLT div_done

    SUB r5, r5, r1
    ADD r4, r4, #1
    B div_loop

div_done:
    MOV r0, r4              ; quotient
    MOV r1, r5              ; remainder
    POP {r4, r5}
    MOV pc, lr

; Helper functions
print_string:
    PUSH {lr}
    SWI #0x02               ; SWI_WRITE_STRING
    POP {pc}

print_int:
    PUSH {lr}
    PUSH {r1}               ; Save R1
    MOV r1, #10             ; Use decimal base
    SWI #0x03               ; SWI_WRITE_INT
    POP {r1}                ; Restore R1
    POP {pc}

print_char:
    PUSH {lr}
    SWI #0x01               ; SWI_WRITE_CHAR
    POP {pc}

read_int:
    PUSH {lr}
    SWI #0x06               ; SWI_READ_INT
    POP {pc}

read_char:
    PUSH {lr}
    SWI #0x04               ; SWI_READ_CHAR
    POP {pc}

exit:
    SWI #0x00               ; SWI_EXIT

; Data section
.align 4
welcome_msg:        .asciz "=== Simple Calculator ===\n\n"
menu_msg:           .asciz "Enter calculation (or 'q' to quit):\n"
num1_prompt:        .asciz "First number: "
op_prompt:          .asciz "Operation (+, -, *, /): "
num2_prompt:        .asciz "Second number: "
result_msg:         .asciz "Result: "
equals:             .asciz " = "
space:              .asciz " "
remainder_msg:      .asciz " remainder "
error_op_msg:       .asciz "Error: Invalid operation. Use +, -, *, or /\n\n"
error_div_msg:      .asciz "Error: Division by zero\n\n"
goodbye_msg:        .asciz "\nGoodbye!\n"
newline:            .asciz "\n"
