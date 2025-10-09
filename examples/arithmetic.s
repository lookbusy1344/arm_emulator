; arithmetic.s - Basic arithmetic operations demonstration
; Demonstrates: Addition, subtraction, multiplication, division

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Addition: 15 + 7 = 22
        MOV     R0, #15
        MOV     R1, #7
        ADD     R2, R0, R1      ; R2 = 15 + 7 = 22

        ; Print result
        LDR     R0, =msg_add
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R2
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Subtraction: 20 - 8 = 12
        MOV     R0, #20
        MOV     R1, #8
        SUB     R2, R0, R1      ; R2 = 20 - 8 = 12

        ; Print result
        LDR     R0, =msg_sub
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R2
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Multiplication: 6 * 7 = 42
        MOV     R0, #6
        MOV     R1, #7
        MUL     R2, R0, R1      ; R2 = 6 * 7 = 42

        ; Print result
        LDR     R0, =msg_mul
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R2
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Division: 35 / 5 = 7
        ; Note: ARM2 doesn't have hardware division, so we use repeated subtraction
        MOV     R0, #35         ; Dividend
        MOV     R1, #5          ; Divisor
        MOV     R2, #0          ; Quotient
div_loop:
        CMP     R0, R1
        BLT     div_done
        SUB     R0, R0, R1
        ADD     R2, R2, #1
        B       div_loop
div_done:
        ; R2 now contains quotient, R0 contains remainder

        ; Print result
        LDR     R0, =msg_div
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R2
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit program
        MOV     R0, #0
        SWI     #0x00           ; EXIT

        ; Data section
msg_add:
        .asciz  "Addition: 15 + 7 = "
msg_sub:
        .asciz  "Subtraction: 20 - 8 = "
msg_mul:
        .asciz  "Multiplication: 6 * 7 = "
msg_div:
        .asciz  "Division: 35 / 5 = "
