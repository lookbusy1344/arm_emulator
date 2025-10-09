; gcd.s - Greatest Common Divisor using Euclidean algorithm
; Demonstrates: Loops, modulo operation, algorithm implementation

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print intro message
        LDR     R0, =msg_intro
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Read first number
        LDR     R0, =msg_first
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x06           ; READ_INT
        MOV     R4, R0          ; Save first number

        ; Read second number
        LDR     R0, =msg_second
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x06           ; READ_INT
        MOV     R5, R0          ; Save second number

        ; Calculate GCD
        MOV     R0, R4
        MOV     R1, R5
        BL      gcd

        ; R0 now contains the GCD
        MOV     R6, R0          ; Save result

        ; Print result
        LDR     R0, =msg_result
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT

        LDR     R0, =msg_and
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT

        LDR     R0, =msg_is
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R6
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit
        MOV     R0, #0
        SWI     #0x00           ; EXIT

; GCD function using Euclidean algorithm
; Input:  R0 = first number (a)
;         R1 = second number (b)
; Output: R0 = GCD(a, b)
; Uses:   R2, R3 as scratch
gcd:
        STMFD   SP!, {R2-R3, LR}

gcd_loop:
        ; While b != 0
        CMP     R1, #0
        BEQ     gcd_done

        ; temp = a % b (using repeated subtraction)
        MOV     R2, R0          ; R2 = a
        MOV     R3, R1          ; R3 = b

mod_loop:
        CMP     R2, R3
        BLT     mod_done
        SUB     R2, R2, R3
        B       mod_loop
mod_done:
        ; R2 now contains a % b

        ; a = b
        MOV     R0, R1
        ; b = temp
        MOV     R1, R2

        B       gcd_loop

gcd_done:
        ; R0 contains the GCD
        LDMFD   SP!, {R2-R3, PC}

        ; Data section
msg_intro:
        .asciz  "GCD Calculator (Euclidean Algorithm)"
msg_first:
        .asciz  "Enter first number: "
msg_second:
        .asciz  "Enter second number: "
msg_result:
        .asciz  "GCD of "
msg_and:
        .asciz  " and "
msg_is:
        .asciz  " is "
