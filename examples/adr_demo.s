; ADR Demo - Position-Independent Address Loading
; Demonstrates the ADR pseudo-instruction for loading PC-relative addresses

        .org 0x8000

_start:
        ; Example 1: Load address of a string
        ADR R0, hello_msg
        SWI #0x02           ; Print string at address in R0
        SWI #0x07           ; Newline

        ; Example 2: Load address of data, then dereference
        ADR R1, magic_number
        LDR R2, [R1]        ; Load the actual value
        MOV R0, R2
        MOV R1, #16         ; Hex format
        SWI #0x03           ; Print integer
        SWI #0x07           ; Newline

        ; Example 3: Calculate array element address
        ADR R3, array_start
        MOV R4, #2          ; Index 2
        MOV R5, #4          ; Element size (word = 4 bytes)
        MUL R6, R4, R5      ; Offset = index * size = 8 (use R6 for result)
        ADD R3, R3, R6      ; address = base + offset
        LDR R0, [R3]        ; Load array[2]
        MOV R1, #10         ; Decimal format
        SWI #0x03           ; Print integer
        SWI #0x07           ; Newline

        ; Example 4: Backward reference (ADR to earlier label)
        ADR R0, hello_msg   ; Reference the same label again
        SWI #0x02           ; Print it again
        
        ; Exit
        MOV R0, #0
        SWI #0x00

; Data section
hello_msg:
        .asciz "Hello from ADR!"

        .align 4
magic_number:
        .word 0xDEADBEEF

array_start:
        .word 10, 20, 30, 40, 50
