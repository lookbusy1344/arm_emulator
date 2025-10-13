; const_expressions.s - Demonstrates constant expression evaluation
; Shows arithmetic operations in constant expressions: +, -, *, /, %, <<, >>
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
msg5:           .asciz "Buffer - 5: "  ; Invalid but demonstrates negative
msg6:           .asciz "Complex expr: "

; Array of numbers for demonstration
numbers:        .word 10, 20, 30, 40, 50, 60, 70, 80, 90, 100

msg_array:      .asciz "Array element "
msg_equals:     .asciz " = "

newline:        .asciz "\n"

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
    
    ; Test 2: Load buffer + 11
    LDR     r0, =msg2
    SWI     #0x02
    
    LDR     r0, =buffer + 11    ; Constant expression: addition
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 3: Load buffer + 33
    LDR     r0, =msg3
    SWI     #0x02
    
    LDR     r0, =buffer + 33    ; Another addition
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 4: Load buffer + 67
    LDR     r0, =msg4
    SWI     #0x02
    
    LDR     r0, =buffer + 67
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 5: Complex expression - buffer + (100 - 33)
    LDR     r0, =msg6
    SWI     #0x02
    
    LDR     r0, =buffer + (100 - 33)  ; = buffer + 67
    MOV     r1, #16
    SWI     #0x03
    SWI     #0x07
    
    ; Test 6: Array access with constant expressions
    ; Access array[3] using offset calculation (3 * 4 bytes)
    LDR     r0, =msg_array
    SWI     #0x02
    
    MOV     r0, #3              ; Index
    MOV     r1, #10
    SWI     #0x03
    
    LDR     r0, =msg_equals
    SWI     #0x02
    
    LDR     r0, =numbers + 3 * 4    ; Load address of array[3]
    LDR     r0, [r0]                ; Dereference
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 7: Array[7] = numbers + 7 * 4 = numbers + 28
    LDR     r0, =msg_array
    SWI     #0x02
    
    MOV     r0, #7
    MOV     r1, #10
    SWI     #0x03
    
    LDR     r0, =msg_equals
    SWI     #0x02
    
    LDR     r0, =numbers + 7 * 4
    LDR     r0, [r0]
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 8: More complex arithmetic - numbers + (5 + 3) * 4
    LDR     r0, =msg_array
    SWI     #0x02
    
    MOV     r0, #8              ; 5 + 3 = 8
    MOV     r1, #10
    SWI     #0x03
    
    LDR     r0, =msg_equals
    SWI     #0x02
    
    LDR     r0, =numbers + (5 + 3) * 4
    LDR     r0, [r0]
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 9: Division in constant expression - numbers + 20 / 4
    ; 20 / 4 = 5, so this is numbers[1.25] but truncated to byte offset 5
    LDR     r0, =msg_array
    SWI     #0x02
    
    MOV     r0, #1              ; offset 5 bytes = words[1] + 1 byte
    MOV     r1, #10
    SWI     #0x03
    
    LDR     r0, =msg_equals
    SWI     #0x02
    
    LDR     r0, =numbers + 20 / 4   ; = numbers + 5
    LDR     r0, [r0]
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 10: Shift operations - numbers + (1 << 3)
    ; 1 << 3 = 8, so numbers[2]
    LDR     r0, =msg_array
    SWI     #0x02
    
    MOV     r0, #2
    MOV     r1, #10
    SWI     #0x03
    
    LDR     r0, =msg_equals
    SWI     #0x02
    
    LDR     r0, =numbers + (1 << 3)  ; = numbers + 8
    LDR     r0, [r0]
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 11: Right shift - numbers + (32 >> 2)
    ; 32 >> 2 = 8, so numbers[2] again
    LDR     r0, =msg_array
    SWI     #0x02
    
    MOV     r0, #2
    MOV     r1, #10
    SWI     #0x03
    
    LDR     r0, =msg_equals
    SWI     #0x02
    
    LDR     r0, =numbers + (32 >> 2)
    LDR     r0, [r0]
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 12: Modulo - numbers + (13 % 5) * 4
    ; 13 % 5 = 3, so 3 * 4 = 12, numbers[3]
    LDR     r0, =msg_array
    SWI     #0x02
    
    MOV     r0, #3
    MOV     r1, #10
    SWI     #0x03
    
    LDR     r0, =msg_equals
    SWI     #0x02
    
    LDR     r0, =numbers + (13 % 5) * 4
    LDR     r0, [r0]
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Test 13: Complex nested expression
    ; numbers + ((10 + 5) * 2 - 6) / 3 * 4
    ; = numbers + (15 * 2 - 6) / 3 * 4
    ; = numbers + (30 - 6) / 3 * 4
    ; = numbers + 24 / 3 * 4
    ; = numbers + 8 * 4
    ; = numbers + 32 = numbers[8]
    LDR     r0, =msg_array
    SWI     #0x02
    
    MOV     r0, #8
    MOV     r1, #10
    SWI     #0x03
    
    LDR     r0, =msg_equals
    SWI     #0x02
    
    LDR     r0, =numbers + ((10 + 5) * 2 - 6) / 3 * 4
    LDR     r0, [r0]
    MOV     r1, #10
    SWI     #0x03
    SWI     #0x07
    
    ; Exit successfully
    MOV     r0, #0
    SWI     #0x00
