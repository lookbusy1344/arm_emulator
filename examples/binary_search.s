; binary_search.s - Binary search algorithm
; Demonstrates: Binary search, array access, function calls

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print search message
        LDR     R0, =msg_intro
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Search for value 25
        MOV     R0, #25         ; Target value
        LDR     R1, =array      ; Array address
        MOV     R2, #0          ; Left index
        MOV     R3, #9          ; Right index (array has 10 elements)
        BL      binary_search

        ; R0 now contains the index (or -1 if not found)
        MOV     R4, R0          ; Save result

        ; Print result
        CMP     R4, #0
        BGE     found

        ; Not found
        LDR     R0, =msg_not_found
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        B       exit

found:
        LDR     R0, =msg_found
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

exit:
        MOV     R0, #0
        SWI     #0x00           ; EXIT

; Binary search function
; Input:  R0 = target value
;         R1 = array address
;         R2 = left index
;         R3 = right index
; Output: R0 = index (or -1 if not found)
; Uses:   R4-R7 as scratch registers
binary_search:
        STMFD   SP!, {R4-R7, LR}

        MOV     R4, R0          ; Save target
        MOV     R5, R1          ; Save array address
        MOV     R6, R2          ; Save left
        MOV     R7, R3          ; Save right

bs_loop:
        ; Check if left > right
        CMP     R6, R7
        BGT     bs_not_found

        ; Calculate mid = (left + right) / 2
        ADD     R0, R6, R7
        MOV     R0, R0, LSR #1  ; Divide by 2

        ; Load array[mid]
        MOV     R1, R0
        MOV     R1, R1, LSL #2  ; Multiply by 4 (word size)
        ADD     R1, R5, R1
        LDR     R2, [R1]

        ; Compare array[mid] with target
        CMP     R2, R4
        BEQ     bs_found
        BLT     bs_right

        ; Search left half
        SUB     R7, R0, #1      ; right = mid - 1
        B       bs_loop

bs_right:
        ; Search right half
        ADD     R6, R0, #1      ; left = mid + 1
        B       bs_loop

bs_found:
        ; Return mid index in R0
        LDMFD   SP!, {R4-R7, PC}

bs_not_found:
        MOV     R0, #-1         ; Return -1
        LDMFD   SP!, {R4-R7, PC}

        ; Data section
        .align  2
array:
        .word   5, 10, 15, 20, 25, 30, 35, 40, 45, 50

msg_intro:
        .asciz  "Searching for value 25 in sorted array [5, 10, 15, 20, 25, 30, 35, 40, 45, 50]"
msg_found:
        .asciz  "Value found at index "
msg_not_found:
        .asciz  "Value not found in array"
