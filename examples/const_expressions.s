; const_expressions.s - Demonstrates constant expression evaluation
; Shows addition and subtraction in constant expressions with non-power-of-2 values
; These are evaluated at assembly time, not runtime
;
; Demonstrates loading addresses at various offsets from labels using
; the = pseudo-instruction with constant expressions

.data
; Create a buffer and demonstrate offset calculations
buffer:         .space 100      ; 100 byte buffer

msg1:           .asciz "Buffer base: "
msg2:           .asciz "Buffer + 11: "
msg3:           .asciz "Buffer + 33: "
msg4:           .asciz "Buffer + 67: "
msg5:           .asciz "Buffer + 99: "
msg6:           .asciz "Buffer - 7: "

; Array of numbers for demonstration
.align 4
numbers:        .word 10, 20, 30, 40, 50, 60, 70, 80, 90, 100

msg_arr_base:   .asciz "Array base: "
msg_arr_12:     .asciz "Array + 12 (numbers[3]): "
msg_arr_28:     .asciz "Array + 28 (numbers[7]): "
msg_arr_36:     .asciz "Array + 36 (numbers[9]): "
msg_value:      .asciz " -> value: "

; Label at end for subtraction tests
.align 4
end_marker:     .word 0xDEADBEEF
msg_end_sub13:  .asciz "end_marker - 13: "
msg_end_sub47:  .asciz "end_marker - 47: "
msg_end_sub91:  .asciz "end_marker - 91: "

.align 4
.text
.global _start

_start:
    ; Test 1: Load buffer base address
    LDR     r0, =msg1
    SWI     #0x02               ; WRITE_STRING
    
    LDR     r0, =buffer         ; Simple label reference
    MOV     r1, #16             ; Hexadecimal
    SWI     #0x03               ; WRITE_INT
    SWI     #0x07               ; WRITE_NEWLINE
    
    ; Test 2: Load buffer + 11 (non-power-of-2)
    LDR     r0, =msg2
    SWI     #0x02
    
    LDR     r0, =buffer + 11    ; Constant expression: addition with 11
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 3: Load buffer + 33 (non-power-of-2)
    LDR     r0, =msg3
    SWI     #0x02
    
    LDR     r0, =buffer + 33    ; Addition with 33
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 4: Load buffer + 67
    LDR     r0, =msg4
    SWI     #0x02
    
    LDR     r0, =buffer + 67    ; Addition with 67
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 5: Load buffer + 99 (almost at end)
    LDR     r0, =msg5
    SWI     #0x02
    
    LDR     r0, =buffer + 99    ; Addition with 99
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 6: Subtraction with non-power-of-2
    LDR     r0, =msg6
    SWI     #0x02
    
    LDR     r1, =buffer         ; Get base address
    LDR     r0, =buffer + 100   ; Go to end
    SUB     r0, r0, #107        ; Subtract 107 to go back 7 bytes from base
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 7: Array base address
    LDR     r0, =msg_arr_base
    SWI     #0x02
    
    LDR     r0, =numbers
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 8: Array element access - numbers + 12 = numbers[3]
    LDR     r0, =msg_arr_12
    SWI     #0x02
    
    LDR     r0, =numbers + 12   ; 12 bytes = 3 words
    MOV     r1, #16
    SWI     #0x03
    
    LDR     r0, =msg_value
    SWI     #0x02
    
    LDR     r0, =numbers + 12
    LDR     r0, [r0]            ; Dereference to get value
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 9: Array element access - numbers + 28 = numbers[7]
    LDR     r0, =msg_arr_28
    SWI     #0x02
    
    LDR     r0, =numbers + 28
    MOV     r1, #16
    SWI     #0x03
    
    LDR     r0, =msg_value
    SWI     #0x02
    
    LDR     r0, =numbers + 28
    LDR     r0, [r0]
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 10: Array element access - numbers + 36 = numbers[9]
    LDR     r0, =msg_arr_36
    SWI     #0x02
    
    LDR     r0, =numbers + 36
    MOV     r1, #16
    SWI     #0x03
    
    LDR     r0, =msg_value
    SWI     #0x02
    
    LDR     r0, =numbers + 36
    LDR     r0, [r0]
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 11: Verify addresses differ by expected amounts
    ; buffer + 67 - (buffer + 33) should equal 34
    LDR     r0, =buffer + 67
    LDR     r1, =buffer + 33
    SUB     r2, r0, r1          ; Should be 34
    
    LDR     r0, =msg_arr_base   ; Reuse message
    SWI     #0x02
    MOV     r0, r2
    MOV     r1, #10
    SWI     #0x03               ; Print difference
    SWI     #0x07
    
    ; Test 12: Subtraction - end_marker - 13
    LDR     r0, =msg_end_sub13
    SWI     #0x02
    
    LDR     r0, =end_marker - 13
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 13: Subtraction - end_marker - 47
    LDR     r0, =msg_end_sub47
    SWI     #0x02
    
    LDR     r0, =end_marker - 47
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 14: Subtraction - end_marker - 91
    LDR     r0, =msg_end_sub91
    SWI     #0x02
    
    LDR     r0, =end_marker - 91
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 15: Verify subtraction - should differ by 34
    LDR     r0, =end_marker - 13
    LDR     r1, =end_marker - 47
    SUB     r2, r0, r1          ; Should be 34 (47 - 13)
    
    LDR     r0, =msg_arr_base   ; Reuse message
    SWI     #0x02
    MOV     r0, r2
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Exit successfully
    MOV     r0, #0
    SWI     #0x00
