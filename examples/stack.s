; stack.s - Stack-based calculator implementation
; Demonstrates: Stack operations, LIFO data structure, expression evaluation

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print intro
        LDR     R0, =msg_intro
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Initialize custom stack pointer (not using system stack)
        LDR     R4, =stack_base
        LDR     R5, =stack_size
        LDR     R5, [R5]
        ADD     R4, R4, R5      ; R4 = stack top (grows downward)
        MOV     R6, #0          ; R6 = stack count

        ; Example: Evaluate (5 + 3) * 2
        ; In postfix: 5 3 + 2 *

        LDR     R0, =msg_expr
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Push 5
        MOV     R0, #5
        BL      stack_push

        ; Push 3
        MOV     R0, #3
        BL      stack_push

        ; Add (pop two, add, push result)
        BL      stack_pop
        MOV     R7, R0
        BL      stack_pop
        ADD     R0, R0, R7
        BL      stack_push

        ; Push 2
        MOV     R0, #2
        BL      stack_push

        ; Multiply
        BL      stack_pop
        MOV     R7, R0
        BL      stack_pop
        MUL     R0, R0, R7
        BL      stack_push

        ; Result is on top of stack
        BL      stack_pop
        MOV     R7, R0

        LDR     R0, =msg_result
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R7
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 2: Evaluate (10 - 6) / 2
        ; In postfix: 10 6 - 2 /

        LDR     R0, =msg_expr2
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Push 10
        MOV     R0, #10
        BL      stack_push

        ; Push 6
        MOV     R0, #6
        BL      stack_push

        ; Subtract
        BL      stack_pop
        MOV     R7, R0
        BL      stack_pop
        SUB     R0, R0, R7
        BL      stack_push

        ; Push 2
        MOV     R0, #2
        BL      stack_push

        ; Divide (using repeated subtraction)
        BL      stack_pop
        MOV     R7, R0          ; divisor
        BL      stack_pop
        MOV     R8, R0          ; dividend
        MOV     R0, #0          ; quotient

div_loop:
        CMP     R8, R7
        BLT     div_done
        SUB     R8, R8, R7
        ADD     R0, R0, #1
        B       div_loop
div_done:
        BL      stack_push

        ; Result
        BL      stack_pop
        MOV     R7, R0

        LDR     R0, =msg_result
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R7
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit
        MOV     R0, #0
        SWI     #0x00           ; EXIT

; Push value onto stack
; Input: R0 = value to push
; Uses: R4 = stack pointer, R6 = count
stack_push:
        STMFD   SP!, {R1, LR}

        ; Check for overflow
        CMP     R6, #100
        BGE     stack_overflow

        ; Push value
        SUB     R4, R4, #4
        STR     R0, [R4]
        ADD     R6, R6, #1

        LDMFD   SP!, {R1, PC}

stack_overflow:
        LDR     R0, =msg_overflow
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        MOV     R0, #1
        SWI     #0x00           ; EXIT

; Pop value from stack
; Output: R0 = popped value
; Uses: R4 = stack pointer, R6 = count
stack_pop:
        STMFD   SP!, {R1, LR}

        ; Check for underflow
        CMP     R6, #0
        BLE     stack_underflow

        ; Pop value
        LDR     R0, [R4]
        ADD     R4, R4, #4
        SUB     R6, R6, #1

        LDMFD   SP!, {R1, PC}

stack_underflow:
        LDR     R0, =msg_underflow
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        MOV     R0, #1
        SWI     #0x00           ; EXIT

        ; Data section
        .align  2
stack_base:
        .space  400             ; 100 words (4 bytes each)
stack_size:
        .word   400

msg_intro:
        .asciz  "Stack-Based Calculator"
msg_expr:
        .asciz  "Evaluating: (5 + 3) * 2"
msg_expr2:
        .asciz  "Evaluating: (10 - 6) / 2"
msg_result:
        .asciz  "Result: "
msg_overflow:
        .asciz  "Error: Stack overflow"
msg_underflow:
        .asciz  "Error: Stack underflow"
