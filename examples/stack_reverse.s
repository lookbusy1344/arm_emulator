; stack_reverse.s - Reverse an array using the hardware stack
; Demonstrates: PUSH, POP, LIFO ordering, SP movement

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Push 1, 2, 3 then pop them: prints 3 2 1
        LDR     R0, =msg_basic
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, #1
        PUSH    {R0}
        MOV     R0, #2
        PUSH    {R0}
        MOV     R0, #3
        PUSH    {R0}

        POP     {R0}
        BL      print_item
        POP     {R0}
        BL      print_item
        POP     {R0}
        BL      print_item
        SWI     #0x07           ; WRITE_NEWLINE

        ; Print the original array
        LDR     R0, =msg_before
        SWI     #0x02           ; WRITE_STRING
        LDR     R4, =array
        LDR     R5, =array_len
        LDR     R5, [R5]
        BL      print_array

        ; Push every element in order
        MOV     R6, #0          ; Index
push_loop:
        CMP     R6, R5
        BGE     push_done
        LDR     R0, [R4, R6, LSL #2]
        PUSH    {R0}
        ADD     R6, R6, #1
        B       push_loop
push_done:

        ; Pop every element back into the array: last pushed lands first
        MOV     R6, #0
pop_loop:
        CMP     R6, R5
        BGE     pop_done
        POP     {R0}
        STR     R0, [R4, R6, LSL #2]
        ADD     R6, R6, #1
        B       pop_loop
pop_done:

        ; Print the reversed array
        LDR     R0, =msg_after
        SWI     #0x02           ; WRITE_STRING
        BL      print_array

        MOV     R0, #0
        SWI     #0x00           ; EXIT

; print_item: print R0 in decimal followed by a space
print_item:
        PUSH    {R0, LR}
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        MOV     R0, #' '
        SWI     #0x01           ; WRITE_CHAR
        POP     {R0, PC}

; print_array: print R5 words starting at R4, then a newline
print_array:
        PUSH    {R6, LR}
        MOV     R6, #0
print_array_loop:
        CMP     R6, R5
        BGE     print_array_done
        LDR     R0, [R4, R6, LSL #2]
        BL      print_item
        ADD     R6, R6, #1
        B       print_array_loop
print_array_done:
        SWI     #0x07           ; WRITE_NEWLINE
        POP     {R6, PC}

        .align  2
array:
        .word   10, 20, 30, 40, 50
array_len:
        .word   5

msg_basic:
        .asciz  "Push 1,2,3 then pop: "
msg_before:
        .asciz  "Before: "
msg_after:
        .asciz  "After:  "
