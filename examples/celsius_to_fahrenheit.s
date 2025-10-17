; Celsius to Fahrenheit Converter
; Reads a Celsius temperature from user input and converts to Fahrenheit
; Formula: F = C * 9 / 5 + 32

.text
.global _start

_start:
    ; Print prompt
    LDR     r0, =prompt
    SWI     #0x02               ; WRITE_STRING

    ; Read Celsius value from stdin
    SWI     #0x06               ; READ_INT - reads integer into r0
    MOV     r4, r0              ; Save Celsius value in r4

    ; Calculate Fahrenheit: F = C * 9 / 5 + 32

    ; Step 1: C * 9
    MOV     r5, #9
    MUL     r6, r4, r5          ; r6 = C * 9

    ; Step 2: C * 9 / 5
    MOV     r0, r6              ; Dividend in r0
    MOV     r1, #5              ; Divisor in r1
    BL      divide              ; Result in r0
    MOV     r6, r0              ; r6 = C * 9 / 5

    ; Step 3: C * 9 / 5 + 32
    ADD     r6, r6, #32         ; r6 = Fahrenheit

    ; Print result message
    LDR     r0, =result_msg
    SWI     #0x02               ; WRITE_STRING

    ; Print Fahrenheit value
    MOV     r0, r6
    MOV     r1, #10             ; Base 10
    SWI     #0x03               ; WRITE_INT

    ; Print newline
    SWI     #0x07               ; WRITE_NEWLINE

    ; Exit
    MOV     r0, #0
    SWI     #0x00               ; sys_exit

; Integer division subroutine
; Input: r0 = dividend, r1 = divisor
; Output: r0 = quotient, r1 = remainder
; Corrupts: r2
divide:
    MOV     r2, #0              ; quotient = 0
    CMP     r1, #0              ; check for division by zero
    MOVEQ   r0, #0
    MOVEQ   r1, #0
    MOVEQ   pc, lr              ; return for div by zero

divide_loop:
    CMP     r0, r1              ; while dividend >= divisor
    BLT     divide_done
    SUB     r0, r0, r1          ; dividend -= divisor
    ADD     r2, r2, #1          ; quotient++
    B       divide_loop

divide_done:
    MOV     r1, r0              ; remainder = dividend
    MOV     r0, r2              ; return quotient
    MOV     pc, lr

.data
prompt:     .asciz "Enter temperature in Celsius: "
result_msg: .asciz "Temperature in Fahrenheit: "
