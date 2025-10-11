; nested_calls.s - Deep function call nesting test
; Demonstrates: Deep recursion, stack management, register preservation
; Tests the stack with deeply nested function calls

        .org    0x8000

_start:
        LDR     R0, =msg_intro
        SWI     #0x02
        SWI     #0x07

        ; Test with depth 10
        MOV     R0, #10
        BL      nested_call
        MOV     R5, R0

        LDR     R0, =msg_result
        SWI     #0x02
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07

        ; Test sum function that calls itself
        MOV     R0, #1
        MOV     R1, #100
        BL      sum_range
        MOV     R5, R0

        LDR     R0, =msg_sum
        SWI     #0x02
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07

        LDR     R0, =msg_done
        SWI     #0x02
        SWI     #0x07

        MOV     R0, #0
        SWI     #0x00

; nested_call - Recursively call function to test stack depth
; Input: R0 = depth remaining
; Output: R0 = depth reached
nested_call:
        STMFD   SP!, {R1-R4, LR}

        ; Base case: if depth == 0, return 0
        CMP     R0, #0
        BEQ     nested_done

        ; Save current depth
        MOV     R4, R0

        ; Call helper functions to use more stack
        BL      helper1
        BL      helper2
        BL      helper3

        ; Recursive call with depth-1
        SUB     R0, R4, #1
        BL      nested_call

        ; Add 1 to result
        ADD     R0, R0, #1

nested_done:
        LDMFD   SP!, {R1-R4, PC}

helper1:
        STMFD   SP!, {R0-R3, LR}
        MOV     R0, #1
        MOV     R1, #2
        MOV     R2, #3
        MOV     R3, #4
        LDMFD   SP!, {R0-R3, PC}

helper2:
        STMFD   SP!, {R0-R3, LR}
        MOV     R0, #5
        MOV     R1, #6
        MOV     R2, #7
        MOV     R3, #8
        LDMFD   SP!, {R0-R3, PC}

helper3:
        STMFD   SP!, {R0-R3, LR}
        MOV     R0, #9
        MOV     R1, #10
        MOV     R2, #11
        MOV     R3, #12
        LDMFD   SP!, {R0-R3, PC}

; sum_range - Calculate sum from start to end (inclusive)
; Input: R0 = start, R1 = end
; Output: R0 = sum
sum_range:
        STMFD   SP!, {R1-R4, LR}

        ; Base case: if start > end, return 0
        CMP     R0, R1
        MOVGT   R0, #0
        BGT     sum_done

        ; Base case: if start == end, return start
        CMP     R0, R1
        BEQ     sum_done

        ; Recursive case: start + sum(start+1, end)
        MOV     R4, R0          ; Save start
        ADD     R0, R0, #1      ; start+1
        BL      sum_range       ; sum(start+1, end)
        ADD     R0, R0, R4      ; start + sum(start+1, end)

sum_done:
        LDMFD   SP!, {R1-R4, PC}

msg_intro:
        .asciz  "Deep Nested Function Call Test"
msg_result:
        .asciz  "Nested call depth reached: "
msg_sum:
        .asciz  "Sum(1..100) = "
msg_done:
        .asciz  "Stack depth test passed!"
