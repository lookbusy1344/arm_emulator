; Test .ltorg directive - Literal Pool Management
; This program demonstrates using .ltorg to place literal pools
; within the ±4095 byte addressing range

.org 0x0000

main:
    ; Load some constants using literal pool
    LDR R0, =0x12345678
    LDR R1, =0xDEADBEEF
    LDR R2, =0xCAFEBABE
    
    ; Add them
    ADD R3, R0, R1
    ADD R3, R3, R2
    
    ; Print result
    MOV R0, R3
    SWI #0x03               ; WRITE_INT
    
    ; Print newline
    MOV R0, #10
    SWI #0x01               ; WRITE_CHAR
    
    ; Branch to next section
    B section2

    ; Place literal pool here (close to the loads above)
    .ltorg

section2:
    ; More code that's far from main
    LDR R4, =0x11111111
    LDR R5, =0x22222222
    LDR R6, =0x33333333
    
    ; Multiply
    ADD R7, R4, R5
    ADD R7, R7, R6
    
    ; Print result
    MOV R0, R7
    SWI #0x03               ; WRITE_INT
    
    ; Print newline
    MOV R0, #10
    SWI #0x01               ; WRITE_CHAR
    
    ; Another literal pool for this section
    .ltorg
    
    ; Exit
    MOV R0, #0
    SWI #0x00               ; EXIT
