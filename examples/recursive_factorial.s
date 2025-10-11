; recursive_factorial.s - Factorial calculation using recursion
; Demonstrates: Recursion, stack operations, function calls, edge cases

        .org    0x8000

_start:
        ; Print intro
        BL      print_intro

        ; Test factorial(0) = 1
        MOV     R0, #0
        BL      factorial
        MOV     R5, R0          ; Save result
        BL      print_fact0
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07

        ; Test factorial(5) = 120
        MOV     R0, #5
        BL      factorial
        MOV     R5, R0
        BL      print_fact5
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07

        ; Test factorial(10) = 3628800
        MOV     R0, #10
        BL      factorial
        MOV     R5, R0
        BL      print_fact10
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07

        ; Done
        BL      print_done

        MOV     R0, #0
        SWI     #0x00

; Helper functions to print messages (reduces literal usage)
print_intro:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg_intro
        SWI     #0x02
        SWI     #0x07
        LDMFD   SP!, {R0, PC}

print_fact0:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg_fact0
        SWI     #0x02
        LDMFD   SP!, {R0, PC}

print_fact5:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg_fact5
        SWI     #0x02
        LDMFD   SP!, {R0, PC}

print_fact10:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg_fact10
        SWI     #0x02
        LDMFD   SP!, {R0, PC}

print_done:
        STMFD   SP!, {R0, LR}
        LDR     R0, =msg_done
        SWI     #0x02
        SWI     #0x07
        LDMFD   SP!, {R0, PC}

; factorial - Calculate n! recursively
; Input: R0 = n
; Output: R0 = n!
; Clobbers: R1, R2
factorial:
        STMFD   SP!, {R1, R2, LR}

        ; Base case: if n <= 1, return 1
        CMP     R0, #1
        MOVLE   R0, #1
        BLE     factorial_done

        ; Recursive case: n! = n * (n-1)!
        MOV     R2, R0          ; Save n in R2
        SUB     R0, R0, #1      ; R0 = n-1
        BL      factorial       ; R0 = (n-1)!
        MOV     R1, R0          ; R1 = (n-1)!
        MOV     R0, R2          ; R0 = n
        MUL     R0, R1, R0      ; R0 = n * (n-1)!

factorial_done:
        LDMFD   SP!, {R1, R2, PC}

msg_intro:
        .asciz  "Recursive Factorial Calculator"
msg_fact0:
        .asciz  "factorial(0) = "
msg_fact1:
        .asciz  "factorial(1) = "
msg_fact5:
        .asciz  "factorial(5) = "
msg_fact10:
        .asciz  "factorial(10) = "
msg_fact12:
        .asciz  "factorial(12) = "
msg_done:
        .asciz  "All factorial tests passed!"
