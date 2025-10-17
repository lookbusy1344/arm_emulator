; add_128bit.s - Add two 128-bit integers
; Demonstrates: Multi-word arithmetic with carry propagation
; A 128-bit integer is stored across 4 32-bit registers (little-endian)
;
; Example: 0x123456789ABCDEF0FEDCBA9876543210 + 0x0000000011111111222222 2222222222
;        = 0x123456789ABCDF02111111110EDD765432221098FE765432

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; First 128-bit number: 0x123456789ABCDEF0FEDCBA9876543210
        ; Stored in R0-R3 (little-endian: R0=low, R3=high)
        LDR     R0, =0x76543210  ; Bits 0-31 (lowest)
        LDR     R1, =0xFEDCBA98  ; Bits 32-63
        LDR     R2, =0x9ABCDEF0  ; Bits 64-95
        LDR     R3, =0x12345678  ; Bits 96-127 (highest)

        ; Second 128-bit number: 0x00000000111111112222222222222222
        ; Stored in R4-R7 (little-endian: R4=low, R7=high)
        LDR     R4, =0x22222222  ; Bits 0-31 (lowest)
        LDR     R5, =0x22222222  ; Bits 32-63
        LDR     R6, =0x11111111  ; Bits 64-95
        LDR     R7, =0x00000000  ; Bits 96-127 (highest)

        ; Print header
        LDR     R0, =msg_header
        SWI     #0x02           ; WRITE_STRING

        ; Print first number parts (in hex format manually)
        LDR     R0, =msg_num1
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R3
        BL      print_hex
        MOV     R0, R2
        BL      print_hex
        MOV     R0, R1
        BL      print_hex
        LDR     R0, =0x76543210
        BL      print_hex
        SWI     #0x07           ; WRITE_NEWLINE

        ; Print second number
        LDR     R0, =msg_num2
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R7
        BL      print_hex
        MOV     R0, R6
        BL      print_hex
        MOV     R0, R5
        BL      print_hex
        MOV     R0, R4
        BL      print_hex
        SWI     #0x07           ; WRITE_NEWLINE

        ; Restore register values for addition
        LDR     R0, =0x76543210
        LDR     R1, =0xFEDCBA98
        LDR     R2, =0x9ABCDEF0
        LDR     R3, =0x12345678
        LDR     R4, =0x22222222
        LDR     R5, =0x22222222
        LDR     R6, =0x11111111
        LDR     R7, =0x00000000

        ; Perform 128-bit addition with carry propagation
        ; Result will be in R8-R11

        ; Add lowest 32 bits (R0 + R4)
        ADDS    R8, R0, R4      ; R8 = R0 + R4, set carry flag

        ; Add next 32 bits with carry (R1 + R5 + carry)
        ADCS    R9, R1, R5      ; R9 = R1 + R5 + carry, set carry flag

        ; Add next 32 bits with carry (R2 + R6 + carry)
        ADCS    R10, R2, R6     ; R10 = R2 + R6 + carry, set carry flag

        ; Add highest 32 bits with carry (R3 + R7 + carry)
        ADCS    R11, R3, R7     ; R11 = R3 + R7 + carry, set carry flag

        ; Print result
        LDR     R0, =msg_result
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R11
        BL      print_hex
        MOV     R0, R10
        BL      print_hex
        MOV     R0, R9
        BL      print_hex
        MOV     R0, R8
        BL      print_hex
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit program
        MOV     R0, #0
        SWI     #0x00           ; EXIT

; Print a 32-bit value in hexadecimal (8 hex digits)
; Input: R0 = value to print
; Preserves: All registers except R0
print_hex:
        PUSH    {R1-R4, LR}
        MOV     R1, R0          ; Save value
        MOV     R2, #8          ; 8 hex digits to print

print_hex_loop:
        ; Extract highest 4 bits
        MOV     R0, R1, LSR #28

        ; Convert to ASCII hex digit
        CMP     R0, #10
        BLT     print_hex_digit

        ; A-F: subtract 10, then add 'A' (65)
        SUB     R0, R0, #10
        ADD     R0, R0, #65
        B       print_hex_char

print_hex_digit:
        ; 0-9: add '0' (48)
        ADD     R0, R0, #48

print_hex_char:
        ; Print the character
        SWI     #0x01           ; WRITE_CHAR

        ; Shift value left by 4 bits
        MOV     R1, R1, LSL #4

        ; Loop for all 8 digits
        SUBS    R2, R2, #1
        BNE     print_hex_loop

        POP     {R1-R4, PC}

        ; Data section
msg_header:
        .asciz  "128-bit Integer Addition Demo\n\n"
msg_num1:
        .asciz  "First number:  0x"
msg_num2:
        .asciz  "Second number: 0x"
msg_result:
        .asciz  "Result:        0x"
