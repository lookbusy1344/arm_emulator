; division.s - Integer division demonstration
; ARM2 has no hardware division, so we implement it in software
; Demonstrates division by repeated subtraction algorithm
;
; The div_simple function implements:
;   quotient = 0
;   while dividend >= divisor:
;       dividend -= divisor
;       quotient++
;   remainder = dividend
;
; Note: WRITE_INT syscall uses R0=value, R1=base (2/8/10/16)

.text
.global _start

_start:
    ; Test 1: Simple division - 100 / 7 = 14 remainder 2
    MOV     r0, #100
    MOV     r1, #7
    BL      div_simple
    ; r0 = quotient (14), r1 = remainder (2)
    
    MOV     r4, r0              ; Save quotient
    MOV     r5, r1              ; Save remainder
    
    ; Print result
    LDR     r0, =msg_test1
    SWI     #0x02               ; WRITE_STRING
    
    MOV     r0, r4
    MOV     r1, #10             ; Base 10
    SWI     #0x03               ; WRITE_INT
    LDR     r0, =msg_remainder
    SWI     #0x02               ; WRITE_STRING
    MOV     r0, r5
    MOV     r1, #10             ; Base 10
    SWI     #0x03               ; WRITE_INT
    SWI     #0x07               ; WRITE_NEWLINE
    
    ; Test 2: Larger division - 1000 / 17 = 58 remainder 14
    MOV     r0, #1000
    MOV     r1, #17
    BL      div_simple
    
    MOV     r4, r0
    MOV     r5, r1
    
    LDR     r0, =msg_test2
    SWI     #0x02
    MOV     r0, r4
    MOV     r1, #10
    SWI     #0x03
    LDR     r0, =msg_remainder
    SWI     #0x02
    MOV     r0, r5
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 3: Exact division - 144 / 12 = 12 remainder 0
    MOV     r0, #144
    MOV     r1, #12
    BL      div_simple
    
    MOV     r4, r0
    MOV     r5, r1
    
    LDR     r0, =msg_test3
    SWI     #0x02
    MOV     r0, r4
    MOV     r1, #10
    SWI     #0x03
    LDR     r0, =msg_remainder
    SWI     #0x02
    MOV     r0, r5
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 4: Division by 1 - 42 / 1 = 42 remainder 0
    MOV     r0, #42
    MOV     r1, #1
    BL      div_simple
    
    MOV     r4, r0
    MOV     r5, r1
    
    LDR     r0, =msg_test4
    SWI     #0x02
    MOV     r0, r4
    MOV     r1, #10
    SWI     #0x03
    LDR     r0, =msg_remainder
    SWI     #0x02
    MOV     r0, r5
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 5: Dividend less than divisor - 5 / 10 = 0 remainder 5
    MOV     r0, #5
    MOV     r1, #10
    BL      div_simple
    
    MOV     r4, r0
    MOV     r5, r1
    
    LDR     r0, =msg_test5
    SWI     #0x02
    MOV     r0, r4
    MOV     r1, #10
    SWI     #0x03
    LDR     r0, =msg_remainder
    SWI     #0x02
    MOV     r0, r5
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 6: Edge case - 0 / 5 = 0 remainder 0
    MOV     r0, #0
    MOV     r1, #5
    BL      div_simple
    
    MOV     r4, r0
    MOV     r5, r1
    
    LDR     r0, =msg_test6
    SWI     #0x02
    MOV     r0, r4
    MOV     r1, #10
    SWI     #0x03
    LDR     r0, =msg_remainder
    SWI     #0x02
    MOV     r0, r5
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Exit
    MOV     r0, #0
    SWI     #0x00               ; sys_exit

; Simple division by repeated subtraction
; Input: r0 = dividend, r1 = divisor
; Output: r0 = quotient, r1 = remainder
; Corrupts: r2
div_simple:
    MOV     r2, #0              ; quotient = 0
    CMP     r1, #0              ; check for division by zero
    MOVEQ   r0, #0
    MOVEQ   r1, #0
    MOVEQ   pc, lr              ; return for div by zero
    
div_loop:
    CMP     r0, r1              ; while dividend >= divisor
    BLT     div_done
    SUB     r0, r0, r1          ; dividend -= divisor
    ADD     r2, r2, #1          ; quotient++
    B       div_loop
    
div_done:
    MOV     r1, r0              ; remainder = dividend
    MOV     r0, r2              ; return quotient
    MOV     pc, lr

.data
msg_test1:      .asciz "100 / 7 = "
msg_test2:      .asciz "1000 / 17 = "
msg_test3:      .asciz "144 / 12 = "
msg_test4:      .asciz "42 / 1 = "
msg_test5:      .asciz "5 / 10 = "
msg_test6:      .asciz "0 / 5 = "
msg_remainder:  .asciz " remainder "
