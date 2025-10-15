; Test program with .org 0x0000 and .ltorg
; This would fail without .ltorg due to literal pool distance

.org 0x0000

main:
    ; Load several large constants
    LDR R0, =0xDEADBEEF
    LDR R1, =0xCAFEBABE
    LDR R2, =0x12345678
    LDR R3, =0xFEEDFACE
    
    ; Perform calculations
    ADD R4, R0, R1
    ADD R5, R2, R3
    ADD R6, R4, R5
    
    ; Print result
    MOV R0, R6
    SWI #0x03           ; WRITE_INT
    
    ; Print newline
    MOV R0, #10
    SWI #0x01           ; WRITE_CHAR
    
    ; Place literal pool within 4095 bytes of the loads above
    .ltorg
    
    ; Exit
    MOV R0, #0
    SWI #0x00           ; EXIT
