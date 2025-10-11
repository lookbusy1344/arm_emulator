; Fibonacci Sequence Generator
; Generates the first N fibonacci numbers
; Example: Input 10 produces 0, 1, 1, 2, 3, 5, 8, 13, 21, 34

.org 0x0000

main:
    ; Print prompt
    LDR r0, =prompt_msg
    BL print_string

    ; Read how many numbers to generate
    BL read_int
    MOV r4, r0              ; r4 = count

    ; Validate input
    CMP r4, #0
    BLE error_invalid
    CMP r4, #20             ; Limit to 20 numbers
    BGT error_too_many

    ; Print header
    LDR r0, =header_msg
    BL print_string

    ; Initialize first two fibonacci numbers
    MOV r5, #0              ; r5 = fib(n-2)
    MOV r6, #1              ; r6 = fib(n-1)
    MOV r7, #0              ; r7 = counter

    ; Special case: if count >= 1, print first number
    CMP r4, #1
    BLT done

    MOV r0, r5
    BL print_int
    LDR r0, =comma_space
    BL print_string
    ADD r7, r7, #1

    ; Special case: if count >= 2, print second number
    CMP r4, #2
    BLT done

    MOV r0, r6
    BL print_int
    ADD r7, r7, #1

    ; Check if we need more numbers
    CMP r7, r4
    BGE done

    LDR r0, =comma_space
    BL print_string

loop:
    ; Check if we've printed enough
    CMP r7, r4
    BGE done

    ; Calculate next fibonacci: r8 = r5 + r6
    ADD r8, r5, r6

    ; Print the number
    MOV r0, r8
    BL print_int

    ; Print comma if not last number
    ADD r7, r7, #1
    CMP r7, r4
    BGE skip_comma

    LDR r0, =comma_space
    BL print_string

skip_comma:
    ; Shift numbers: r5 = r6, r6 = r8
    MOV r5, r6
    MOV r6, r8

    B loop

done:
    ; Print newline
    LDR r0, =newline
    BL print_string

    ; Exit
    MOV r0, #0
    BL exit

error_invalid:
    LDR r0, =error_invalid_msg
    BL print_string
    MOV r0, #1
    BL exit

error_too_many:
    LDR r0, =error_many_msg
    BL print_string
    MOV r0, #1
    BL exit

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
prompt_msg:         .asciz "How many Fibonacci numbers to generate (1-20)? "
header_msg:         .asciz "Fibonacci sequence: "
comma_space:        .asciz ", "
error_invalid_msg:  .asciz "Error: Please enter a positive number\n"
error_many_msg:     .asciz "Error: Too many numbers (max 20)\n"
newline:            .asciz "\n"
