; functions.s - Function calling conventions demonstration
; Demonstrates: Function calls, parameter passing, return values, stack usage

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print intro
        LDR     R0, =msg_intro
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 1: Simple function with return value
        LDR     R0, =msg_ex1
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, #15
        MOV     R1, #7
        BL      add_two_numbers

        MOV     R4, R0          ; Save result

        LDR     R0, =msg_result
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 2: Function with multiple parameters
        LDR     R0, =msg_ex2
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, #10         ; a
        MOV     R1, #20         ; b
        MOV     R2, #30         ; c
        MOV     R3, #40         ; d
        BL      sum_four_numbers

        MOV     R4, R0

        LDR     R0, =msg_result
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 3: Nested function calls
        LDR     R0, =msg_ex3
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, #3
        MOV     R1, #4
        BL      calculate_hypotenuse

        MOV     R4, R0

        LDR     R0, =msg_result
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 4: Function modifying parameters by reference
        LDR     R0, =msg_ex4
        SWI     #0x02           ; WRITE_STRING

        LDR     R0, =value1
        MOV     R1, #100
        STR     R1, [R0]

        LDR     R0, =value2
        MOV     R1, #200
        STR     R1, [R0]

        LDR     R0, =value1
        LDR     R1, =value2
        BL      swap_values

        LDR     R0, =msg_after_swap
        SWI     #0x02           ; WRITE_STRING

        LDR     R0, =value1
        LDR     R0, [R0]
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT

        LDR     R0, =msg_comma
        SWI     #0x02           ; WRITE_STRING

        LDR     R0, =value2
        LDR     R0, [R0]
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit
        MOV     R0, #0
        SWI     #0x00           ; EXIT

; Add two numbers
; Input:  R0 = first number
;         R1 = second number
; Output: R0 = sum
add_two_numbers:
        ADD     R0, R0, R1
        MOV     PC, LR          ; Return

; Sum four numbers
; Input:  R0 = a, R1 = b, R2 = c, R3 = d
; Output: R0 = a + b + c + d
sum_four_numbers:
        ADD     R0, R0, R1
        ADD     R0, R0, R2
        ADD     R0, R0, R3
        MOV     PC, LR          ; Return

; Calculate hypotenuse (sqrt(a^2 + b^2)) - approximation
; Input:  R0 = a
;         R1 = b
; Output: R0 = hypotenuse (approximate)
calculate_hypotenuse:
        STMFD   SP!, {R1-R5, LR}

        ; Calculate a^2
        MOV     R2, R0
        MUL     R2, R0, R2      ; R2 = a * a

        ; Calculate b^2
        MOV     R3, R1
        MUL     R3, R1, R3      ; R3 = b * b

        ; Sum = a^2 + b^2
        ADD     R4, R2, R3

        ; Simple integer square root using binary search
        MOV     R0, R4
        BL      isqrt

        LDMFD   SP!, {R1-R5, PC}

; Integer square root
; Input:  R0 = number
; Output: R0 = floor(sqrt(number))
isqrt:
        STMFD   SP!, {R1-R4, LR}

        MOV     R1, R0          ; R1 = n
        MOV     R2, #0          ; R2 = low
        MOV     R3, R1          ; R3 = high

isqrt_loop:
        CMP     R2, R3
        BGT     isqrt_done

        ; mid = (low + high) / 2
        ADD     R4, R2, R3
        MOV     R4, R4, LSR #1

        ; mid * mid
        MOV     R0, R4
        MUL     R0, R4, R0

        CMP     R0, R1
        BEQ     isqrt_found
        BLT     isqrt_increase

        ; mid^2 > n, search lower half
        SUB     R3, R4, #1
        B       isqrt_loop

isqrt_increase:
        ; mid^2 < n, search upper half
        ADD     R2, R4, #1
        B       isqrt_loop

isqrt_found:
        MOV     R0, R4
        LDMFD   SP!, {R1-R4, PC}

isqrt_done:
        MOV     R0, R3
        LDMFD   SP!, {R1-R4, PC}

; Swap two values by reference
; Input:  R0 = address of first value
;         R1 = address of second value
swap_values:
        STMFD   SP!, {R2-R3, LR}

        LDR     R2, [R0]        ; Load first value
        LDR     R3, [R1]        ; Load second value

        STR     R3, [R0]        ; Store second value to first location
        STR     R2, [R1]        ; Store first value to second location

        LDMFD   SP!, {R2-R3, PC}

        ; Data section
msg_intro:
        .asciz  "Function Calling Conventions Demo"
msg_ex1:
        .asciz  "Example 1: Adding 15 + 7"
msg_ex2:
        .asciz  "Example 2: Sum of 10, 20, 30, 40"
msg_ex3:
        .asciz  "Example 3: Hypotenuse of 3 and 4"
msg_ex4:
        .asciz  "Example 4: Swapping 100 and 200"
msg_result:
        .asciz  "Result: "
msg_after_swap:
        .asciz  "After swap: "
msg_comma:
        .asciz  ", "

        .align  2
value1:
        .word   0
value2:
        .word   0
