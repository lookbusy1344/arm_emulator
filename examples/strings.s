; strings.s - String manipulation operations
; Demonstrates: String length, copy, compare, concatenation

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print intro
        LDR     R0, =msg_intro
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Test 1: String length
        LDR     R0, =msg_test1
        SWI     #0x02           ; WRITE_STRING

        LDR     R0, =str1
        BL      strlen
        MOV     R4, R0

        LDR     R0, =msg_str1
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =str1
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =msg_length
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Test 2: String copy
        LDR     R0, =msg_test2
        SWI     #0x02           ; WRITE_STRING

        LDR     R0, =buffer
        LDR     R1, =str2
        BL      strcpy

        LDR     R0, =msg_copied
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =buffer
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =msg_quote
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Test 3: String compare
        LDR     R0, =msg_test3
        SWI     #0x02           ; WRITE_STRING

        LDR     R0, =str1
        LDR     R1, =str2
        BL      strcmp
        MOV     R4, R0

        LDR     R0, =msg_compare1
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =str1
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =msg_compare2
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =str2
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =msg_compare3
        SWI     #0x02           ; WRITE_STRING

        CMP     R4, #0
        BEQ     str_equal
        BLT     str_less
        BGT     str_greater

str_equal:
        LDR     R0, =msg_equal
        B       str_cmp_done
str_less:
        LDR     R0, =msg_less
        B       str_cmp_done
str_greater:
        LDR     R0, =msg_greater
str_cmp_done:
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Test 4: String concatenation
        LDR     R0, =msg_test4
        SWI     #0x02           ; WRITE_STRING

        ; Copy str3 to concat_buf
        LDR     R0, =concat_buf
        LDR     R1, =str3
        BL      strcpy

        ; Concatenate str4
        LDR     R0, =concat_buf
        LDR     R1, =str4
        BL      strcat

        LDR     R0, =msg_concat
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =concat_buf
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =msg_quote
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit
        MOV     R0, #0
        SWI     #0x00           ; EXIT

; Calculate string length
; Input:  R0 = string address
; Output: R0 = length (excluding null terminator)
strlen:
        STMFD   SP!, {R1-R2, LR}
        MOV     R1, R0          ; R1 = string pointer
        MOV     R2, #0          ; length = 0

strlen_loop:
        LDRB    R0, [R1], #1    ; Load byte and increment
        CMP     R0, #0
        BEQ     strlen_done
        ADD     R2, R2, #1
        B       strlen_loop

strlen_done:
        MOV     R0, R2
        LDMFD   SP!, {R1-R2, PC}

; Copy string
; Input:  R0 = destination
;         R1 = source
strcpy:
        STMFD   SP!, {R1-R3, LR}
        MOV     R2, R0          ; R2 = dest
        MOV     R3, R1          ; R3 = src

strcpy_loop:
        LDRB    R0, [R3], #1    ; Load byte from src
        STRB    R0, [R2], #1    ; Store byte to dest
        CMP     R0, #0          ; Check for null terminator
        BNE     strcpy_loop

        LDMFD   SP!, {R1-R3, PC}

; Compare strings
; Input:  R0 = string1
;         R1 = string2
; Output: R0 = 0 if equal, <0 if str1 < str2, >0 if str1 > str2
strcmp:
        STMFD   SP!, {R1-R4, LR}
        MOV     R2, R0          ; R2 = str1
        MOV     R3, R1          ; R3 = str2

strcmp_loop:
        LDRB    R0, [R2], #1    ; Load byte from str1
        LDRB    R1, [R3], #1    ; Load byte from str2

        CMP     R0, R1
        BNE     strcmp_diff

        ; Check if both are null
        CMP     R0, #0
        BEQ     strcmp_equal

        B       strcmp_loop

strcmp_diff:
        SUB     R0, R0, R1      ; Return difference
        LDMFD   SP!, {R1-R4, PC}

strcmp_equal:
        MOV     R0, #0          ; Strings are equal
        LDMFD   SP!, {R1-R4, PC}

; Concatenate strings
; Input:  R0 = destination (must have enough space)
;         R1 = source to append
strcat:
        STMFD   SP!, {R1-R4, LR}
        MOV     R2, R0          ; R2 = dest
        MOV     R3, R1          ; R3 = src

        ; Find end of destination string
strcat_find_end:
        LDRB    R4, [R2]
        CMP     R4, #0
        ADDNE   R2, R2, #1
        BNE     strcat_find_end

        ; Now R2 points to null terminator of dest
        ; Copy src to end of dest
strcat_copy:
        LDRB    R4, [R3], #1
        STRB    R4, [R2], #1
        CMP     R4, #0
        BNE     strcat_copy

        LDMFD   SP!, {R1-R4, PC}

        ; Data section
msg_intro:
        .asciz  "String Operations Demo"
msg_test1:
        .asciz  "Test 1: String Length"
msg_test2:
        .asciz  "Test 2: String Copy"
msg_test3:
        .asciz  "Test 3: String Compare"
msg_test4:
        .asciz  "Test 4: String Concatenation"

msg_str1:
        .asciz  "String: \""
msg_length:
        .asciz  "\" has length "
msg_copied:
        .asciz  "Copied string: \""
msg_compare1:
        .asciz  "Comparing \""
msg_compare2:
        .asciz  "\" with \""
msg_compare3:
        .asciz  "\": "
msg_equal:
        .asciz  "strings are equal"
msg_less:
        .asciz  "first string is less"
msg_greater:
        .asciz  "first string is greater"
msg_concat:
        .asciz  "Concatenated: \""
msg_quote:
        .asciz  "\""

str1:
        .asciz  "Hello"
str2:
        .asciz  "World"
str3:
        .asciz  "ARM "
str4:
        .asciz  "Assembly"

        .align  2
buffer:
        .space  64
concat_buf:
        .space  128
