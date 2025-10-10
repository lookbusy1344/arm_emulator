; loops.s - Loop constructs demonstration
; Demonstrates: For loops, while loops, do-while loops, nested loops

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print intro
        LDR     R0, =msg_intro
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 1: For loop (count 1 to 5)
        LDR     R0, =msg_ex1
        SWI     #0x02           ; WRITE_STRING

        MOV     R4, #1          ; i = 1
for_loop:
        CMP     R4, #5
        BGT     for_done

        ; Print i
        MOV     R0, R4
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        MOV     R0, #' '
        SWI     #0x01           ; WRITE_CHAR

        ADD     R4, R4, #1      ; i++
        B       for_loop

for_done:
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 2: While loop (sum 1 to 10)
        LDR     R0, =msg_ex2
        SWI     #0x02           ; WRITE_STRING

        MOV     R4, #1          ; i = 1
        MOV     R5, #0          ; sum = 0

while_loop:
        CMP     R4, #10
        BGT     while_done

        ADD     R5, R5, R4      ; sum += i
        ADD     R4, R4, #1      ; i++
        B       while_loop

while_done:
        LDR     R0, =msg_sum
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 3: Do-while loop (factorial of 5)
        LDR     R0, =msg_ex3
        SWI     #0x02           ; WRITE_STRING

        MOV     R4, #5          ; n = 5
        MOV     R5, #1          ; result = 1
        MOV     R6, R4          ; i = n

do_while_loop:
        MUL     R5, R6, R5      ; result *= i (R5 = R6 * R5)
        SUB     R6, R6, #1      ; i--
        CMP     R6, #0
        BGT     do_while_loop

        LDR     R0, =msg_factorial
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 4: Nested loops (multiplication table 3x3)
        LDR     R0, =msg_ex4
        SWI     #0x02           ; WRITE_STRING

        MOV     R4, #1          ; row = 1

outer_loop:
        CMP     R4, #3
        BGT     outer_done

        MOV     R5, #1          ; col = 1

inner_loop:
        CMP     R5, #3
        BGT     inner_done

        ; Calculate row * col
        MUL     R6, R4, R5

        ; Print result
        MOV     R0, R6
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        MOV     R0, #'\t'
        SWI     #0x01           ; WRITE_CHAR

        ADD     R5, R5, #1      ; col++
        B       inner_loop

inner_done:
        SWI     #0x07           ; WRITE_NEWLINE
        ADD     R4, R4, #1      ; row++
        B       outer_loop

outer_done:
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 5: Break and continue simulation (find first even number)
        LDR     R0, =msg_ex5
        SWI     #0x02           ; WRITE_STRING

        LDR     R4, =numbers
        MOV     R5, #0          ; index = 0

search_loop:
        CMP     R5, #10
        BGE     search_not_found

        ; Load number
        MOV     R0, R5, LSL #2
        ADD     R0, R4, R0
        LDR     R6, [R0]

        ; Check if even (bit 0 == 0)
        ANDS    R0, R6, #1
        BNE     search_continue ; If odd, continue

        ; Found even number
        B       search_found

search_continue:
        ADD     R5, R5, #1
        B       search_loop

search_found:
        LDR     R0, =msg_found_at
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        LDR     R0, =msg_value
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R6
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE
        B       search_done

search_not_found:
        LDR     R0, =msg_not_found
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

search_done:
        ; Exit
        MOV     R0, #0
        SWI     #0x00           ; EXIT

        ; Data section
msg_intro:
        .asciz  "Loop Constructs Demo"
msg_ex1:
        .asciz  "Example 1: For loop (1 to 5): "
msg_ex2:
        .asciz  "Example 2: While loop (sum 1 to 10)"
msg_ex3:
        .asciz  "Example 3: Do-while loop (5!)"
msg_ex4:
        .asciz  "Example 4: Nested loops (3x3 multiplication table):"
msg_ex5:
        .asciz  "Example 5: Break/continue (find first even number)"

msg_sum:
        .asciz  "Sum = "
msg_factorial:
        .asciz  "5! = "
msg_found_at:
        .asciz  "Found even number at index "
msg_value:
        .asciz  " with value "
msg_not_found:
        .asciz  "No even number found"

        .align  2
numbers:
        .word   7, 13, 21, 9, 33, 18, 25, 41, 51, 17
