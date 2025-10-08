; Times Table Generator
; Reads a number from input and displays its multiplication table (1-12)
; Example: Input 5 produces 5x1=5, 5x2=10, ..., 5x12=60

.org 0x0000

main:
    ; Print prompt message
    LDR r0, =prompt_msg
    BL print_string

    ; Read number from input
    BL read_int
    MOV r4, r0              ; Store the input number in r4

    ; Initialize counter
    MOV r5, #1              ; r5 = counter (1 to 12)

loop:
    ; Check if counter > 12
    CMP r5, #12
    BGT end

    ; Calculate multiplication: r4 * r5
    MUL r6, r4, r5          ; r6 = r4 * r5

    ; Print: "N x M = Result"
    MOV r0, r4
    BL print_int

    LDR r0, =times_msg
    BL print_string

    MOV r0, r5
    BL print_int

    LDR r0, =equals_msg
    BL print_string

    MOV r0, r6
    BL print_int

    LDR r0, =newline
    BL print_string

    ; Increment counter
    ADD r5, r5, #1
    B loop

end:
    ; Exit program
    MOV r0, #0
    BL exit

; Helper functions (syscall wrappers)
print_string:
    ; Print null-terminated string in r0
    PUSH {r7, lr}
    MOV r7, #4              ; syscall: write
    SVC #0
    POP {r7, pc}

print_int:
    ; Print integer in r0
    PUSH {r7, lr}
    MOV r7, #1              ; syscall: print_int
    SVC #0
    POP {r7, pc}

read_int:
    ; Read integer into r0
    PUSH {r7, lr}
    MOV r7, #3              ; syscall: read_int
    SVC #0
    POP {r7, pc}

exit:
    MOV r7, #0              ; syscall: exit
    SVC #0

; Data section
.align 4
prompt_msg:     .asciz "Enter a number (1-12): "
times_msg:      .asciz " x "
equals_msg:     .asciz " = "
newline:        .asciz "\n"
