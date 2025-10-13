; Test constant expression evaluation

.data
buffer:     .space 12
buffer_end:

.text
.global _start

_start:
    ; Test that buffer+12 equals buffer_end
    LDR r0, =buffer + 12
    LDR r1, =buffer_end
    SWI #0xF2               ; Debug
    CMP r0, r1
    BNE fail
    
    ; Test hex offset
    LDR r2, =buffer + 0x0C
    SWI #0xF2               ; Debug
    CMP r0, r2
    BNE fail
    
    ; Test subtraction: buffer_end-12 should equal buffer
    LDR r3, =buffer_end - 12
    LDR r4, =buffer
    SWI #0xF2               ; Debug
    CMP r3, r4
    BNE fail
    
    ; Success
    LDR r0, =msg_success
    SWI #0x02
    SWI #0x07
    MOV r0, #0
    SWI #0x00

fail:
    LDR r0, =msg_fail
    SWI #0x02
    SWI #0x07
    MOV r0, #1
    SWI #0x00

msg_success: .asciz "PASS: Constant expressions work correctly"
msg_fail:    .asciz "FAIL: Expression mismatch"
