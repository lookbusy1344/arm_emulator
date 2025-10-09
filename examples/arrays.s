; arrays.s - Array operations demonstration
; Demonstrates: Array initialization, access, traversal, min/max

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print intro
        LDR     R0, =msg_intro
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Initialize array with values 10, 25, 5, 42, 18, 33, 7, 50, 12, 3
        LDR     R4, =array      ; R4 = array address
        LDR     R5, =array_size
        LDR     R5, [R5]        ; R5 = array size

        ; Print array
        LDR     R0, =msg_array
        SWI     #0x02           ; WRITE_STRING

        MOV     R6, #0          ; Index
print_loop:
        CMP     R6, R5
        BGE     print_done

        ; Load and print array[i]
        MOV     R0, R6, LSL #2  ; Offset = index * 4
        ADD     R0, R4, R0
        LDR     R0, [R0]
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT

        ; Print comma if not last element
        ADD     R6, R6, #1
        CMP     R6, R5
        BGE     print_loop
        MOV     R0, #','
        SWI     #0x01           ; WRITE_CHAR
        MOV     R0, #' '
        SWI     #0x01           ; WRITE_CHAR
        B       print_loop

print_done:
        SWI     #0x07           ; WRITE_NEWLINE

        ; Find minimum value
        BL      find_min
        MOV     R7, R0          ; Save min

        ; Find maximum value
        BL      find_max
        MOV     R8, R0          ; Save max

        ; Print min
        LDR     R0, =msg_min
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R7
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Print max
        LDR     R0, =msg_max
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R8
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Calculate sum
        BL      array_sum
        MOV     R7, R0          ; Save sum

        ; Print sum
        LDR     R0, =msg_sum
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R7
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit
        MOV     R0, #0
        SWI     #0x00           ; EXIT

; Find minimum value in array
; Output: R0 = minimum value
find_min:
        STMFD   SP!, {R1-R5, LR}

        LDR     R1, =array
        LDR     R2, =array_size
        LDR     R2, [R2]
        LDR     R0, [R1]        ; min = array[0]
        MOV     R3, #1          ; i = 1

min_loop:
        CMP     R3, R2
        BGE     min_done

        MOV     R4, R3, LSL #2
        ADD     R4, R1, R4
        LDR     R4, [R4]        ; array[i]

        CMP     R4, R0
        MOVLT   R0, R4          ; if array[i] < min, min = array[i]

        ADD     R3, R3, #1
        B       min_loop

min_done:
        LDMFD   SP!, {R1-R5, PC}

; Find maximum value in array
; Output: R0 = maximum value
find_max:
        STMFD   SP!, {R1-R5, LR}

        LDR     R1, =array
        LDR     R2, =array_size
        LDR     R2, [R2]
        LDR     R0, [R1]        ; max = array[0]
        MOV     R3, #1          ; i = 1

max_loop:
        CMP     R3, R2
        BGE     max_done

        MOV     R4, R3, LSL #2
        ADD     R4, R1, R4
        LDR     R4, [R4]        ; array[i]

        CMP     R4, R0
        MOVGT   R0, R4          ; if array[i] > max, max = array[i]

        ADD     R3, R3, #1
        B       max_loop

max_done:
        LDMFD   SP!, {R1-R5, PC}

; Calculate array sum
; Output: R0 = sum of all elements
array_sum:
        STMFD   SP!, {R1-R5, LR}

        LDR     R1, =array
        LDR     R2, =array_size
        LDR     R2, [R2]
        MOV     R0, #0          ; sum = 0
        MOV     R3, #0          ; i = 0

sum_loop:
        CMP     R3, R2
        BGE     sum_done

        MOV     R4, R3, LSL #2
        ADD     R4, R1, R4
        LDR     R4, [R4]        ; array[i]

        ADD     R0, R0, R4      ; sum += array[i]

        ADD     R3, R3, #1
        B       sum_loop

sum_done:
        LDMFD   SP!, {R1-R5, PC}

        ; Data section
        .align  2
array:
        .word   10, 25, 5, 42, 18, 33, 7, 50, 12, 3
array_size:
        .word   10

msg_intro:
        .asciz  "Array Operations Demo"
msg_array:
        .asciz  "Array: "
msg_min:
        .asciz  "Minimum value: "
msg_max:
        .asciz  "Maximum value: "
msg_sum:
        .asciz  "Sum of all elements: "
