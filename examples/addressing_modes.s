; addressing_modes_simple.s - ARM addressing modes demonstration
; This version works around a literal pool bug by minimizing LDR =label usage

        .org    0x8000

_start:
        ; Print intro
        LDR     R0, =msg_intro
        SWI     #0x02
        SWI     #0x07
        SWI     #0x07

        ; =====================================================
        ; Immediate Offset: [Rn, #offset]
        ; =====================================================
        SUB     SP, SP, #64     ; Allocate stack space

        MOV     R2, #100
        STR     R2, [SP]
        MOV     R2, #200
        STR     R2, [SP, #4]
        MOV     R2, #300
        STR     R2, [SP, #8]

        LDR     R3, [SP, #4]    ; Load from offset 4
        CMP     R3, #200
        BNE     fail1

        ; =====================================================
        ; Pre-indexed with Writeback: [Rn, #offset]!
        ; =====================================================
        MOV     R4, SP
        LDR     R5, [R4, #8]!   ; R4 updated to SP+8, load value
        CMP     R5, #300
        BNE     fail2

        ; Verify R4 was updated
        SUB     R0, R4, SP
        CMP     R0, #8
        BNE     fail2

        ; =====================================================
        ; Post-indexed: [Rn], #offset
        ; =====================================================
        MOV     R7, SP
        LDR     R8, [R7], #4    ; Load from R7, then R7=R7+4
        CMP     R8, #100
        BNE     fail3

        ; Verify R7 was updated
        SUB     R0, R7, SP
        CMP     R0, #4
        BNE     fail3

        ; =====================================================
        ; Register Offset: [Rn, Rm]
        ; =====================================================
        MOV     R9, SP
        MOV     R10, #8
        LDR     R11, [R9, R10]  ; Load from SP+8
        CMP     R11, #300
        BNE     fail4

        ; =====================================================
        ; Scaled Register Offset: [Rn, Rm, LSL #shift]
        ; =====================================================
        MOV     R12, SP
        MOV     R1, #1          ; Index 1
        LDR     R2, [R12, R1, LSL #2]  ; Load from SP+(1<<2)=SP+4
        CMP     R2, #200
        BNE     fail5

        ; =====================================================
        ; Byte Access with Post-indexed
        ; =====================================================
        SUB     SP, SP, #16
        MOV     R3, SP
        MOV     R4, #'A'
        STRB    R4, [R3], #1
        MOV     R4, #'R'
        STRB    R4, [R3], #1
        MOV     R4, #'M'
        STRB    R4, [R3]

        MOV     R5, SP
        LDRB    R6, [R5]
        CMP     R6, #'A'
        BNE     fail6

        ; =====================================================
        ; Success!
        ; =====================================================
        LDR     R0, =msg_success
        SWI     #0x02
        SWI     #0x07
        MOV     R0, #0
        SWI     #0x00

fail1:
        LDR     R0, =msg_fail1
        B       fail_exit
fail2:
        LDR     R0, =msg_fail2
        B       fail_exit
fail3:
        LDR     R0, =msg_fail3
        B       fail_exit
fail4:
        LDR     R0, =msg_fail4
        B       fail_exit
fail5:
        LDR     R0, =msg_fail5
        B       fail_exit
fail6:
        LDR     R0, =msg_fail6

fail_exit:
        SWI     #0x02
        SWI     #0x07
        MOV     R0, #1
        SWI     #0x00

msg_intro:
        .asciz  "Testing ARM Addressing Modes..."
msg_success:
        .asciz  "All addressing mode tests passed!"
msg_fail1:
        .asciz  "FAIL: Immediate offset [Rn, #offset]"
msg_fail2:
        .asciz  "FAIL: Pre-indexed with writeback [Rn, #offset]!"
msg_fail3:
        .asciz  "FAIL: Post-indexed [Rn], #offset"
msg_fail4:
        .asciz  "FAIL: Register offset [Rn, Rm]"
msg_fail5:
        .asciz  "FAIL: Scaled register offset [Rn, Rm, LSL #shift]"
msg_fail6:
        .asciz  "FAIL: Byte access"
