; multi_precision_arith.s - Multi-precision (128-bit) arithmetic and flag stress test
; Demonstrates: ADC/SBC carry chain, CMP/CMN flag effects, conditional execution,
;               LDM/STM frame usage, mixed shifts, verification logic.
;
; Operations performed:
;   1. 128-bit A + B -> SUM
;   2. 128-bit SUM - B -> DIFF (should equal A)
;   3. 128-bit (A + B) + C with carry propagation
;   4. Simple multi-word left shift by 1 (SUM <<= 1)
; Verifies each step and prints PASS / FAIL messages.
;
; Register usage (within routines):
;   R0-R3  temporary / word operands
;   R4     pointer to operand A
;   R5     pointer to operand B
;   R6     pointer to operand C
;   R7     pointer to result buffers
;   R8     loop counter / scratch
;   R9     expected pointer
;   R10    status aggregate (0 = all pass)
;   R11    scratch
;   R12    scratch
;
; Word order: Little-endian limbs: [0] = least significant 32 bits
;
        .org    0x8000

_start:
        LDR     R0, =msg_intro
        SWI     #0x02
        SWI     #0x07

        ; Initialize status = 0
        MOV     R10, #0

        LDR     R4, =opA
        LDR     R5, =opB
        LDR     R6, =opC
        LDR     R7, =sum          ; SUM buffer

        BL      add128            ; SUM = A + B
        BL      verify_sum_ab
        CMP     R0, #0
        BNE     test1_fail
        LDR     R0, =msg_test1_pass
        SWI     #0x02
        SWI     #0x07
        MOV     R0, #0
test1_fail:
        ORR     R10, R10, R0      ; accumulate failures

        ; DIFF = SUM - B -> should recover A
        LDR     R0, =sum
        LDR     R1, =opB
        LDR     R2, =diff
        BL      sub128            ; DIFF = SUM - B
        BL      verify_diff_recovers_a
        CMP     R0, #0
        BNE     test2_fail
        LDR     R0, =msg_test2_pass
        SWI     #0x02
        SWI     #0x07
        MOV     R0, #0
test2_fail:
        ORR     R10, R10, R0

        ; SUM2 = SUM + C
        LDR     R0, =sum
        LDR     R1, =opC
        LDR     R2, =sum2
        BL      add128_generic
        BL      verify_sum2
        CMP     R0, #0
        BNE     test3_fail
        LDR     R0, =msg_test3_pass
        SWI     #0x02
        SWI     #0x07
        MOV     R0, #0
test3_fail:
        ORR     R10, R10, R0

        ; Shift SUM left by 1 bit -> SHIFTED
        LDR     R0, =sum
        LDR     R1, =shifted
        BL      shl128_1
        BL      verify_shifted
        CMP     R0, #0
        BNE     test4_fail
        LDR     R0, =msg_test4_pass
        SWI     #0x02
        SWI     #0x07
        MOV     R0, #0
test4_fail:
        ORR     R10, R10, R0

        ; Final result
        CMP     R10, #0
        BNE     overall_fail
        LDR     R0, =msg_all_pass
        SWI     #0x02
        B       done

overall_fail:
        LDR     R0, =msg_any_fail
        SWI     #0x02

done:
        SWI     #0x07
        MOV     R0, #0
        SWI     #0x00

; add128: (R4)=A, (R5)=B, R7=destination buffer SUM
; Clobbers R0-R3,R8, flags
add128:
        STMFD   SP!, {R4-R8, LR}
        ; Unrolled loop to preserve carry flag between ADCS
        LDR     R0, [R4, #0]
        LDR     R1, [R5, #0]
        ADDS    R2, R0, R1        ; First add, clear carry
        STR     R2, [R7, #0]
        
        LDR     R0, [R4, #4]
        LDR     R1, [R5, #4]
        ADCS    R2, R0, R1        ; Add with carry
        STR     R2, [R7, #4]
        
        LDR     R0, [R4, #8]
        LDR     R1, [R5, #8]
        ADCS    R2, R0, R1        ; Add with carry
        STR     R2, [R7, #8]
        
        LDR     R0, [R4, #12]
        LDR     R1, [R5, #12]
        ADCS    R2, R0, R1        ; Add with carry
        STR     R2, [R7, #12]
        
        ; Final carry in C flag (ignored)
        MOV     R0, #0
        LDMFD   SP!, {R4-R8, PC}

; add128_generic: R0=ptr X, R1=ptr Y, R2=dest -> dest = X + Y
add128_generic:
        STMFD   SP!, {R3-R8, LR}
        ; Unrolled loop to preserve carry flag
        LDR     R4, [R0, #0]
        LDR     R5, [R1, #0]
        ADDS    R6, R4, R5        ; First add
        STR     R6, [R2, #0]
        
        LDR     R4, [R0, #4]
        LDR     R5, [R1, #4]
        ADCS    R6, R4, R5        ; Add with carry
        STR     R6, [R2, #4]
        
        LDR     R4, [R0, #8]
        LDR     R5, [R1, #8]
        ADCS    R6, R4, R5        ; Add with carry
        STR     R6, [R2, #8]
        
        LDR     R4, [R0, #12]
        LDR     R5, [R1, #12]
        ADCS    R6, R4, R5        ; Add with carry
        STR     R6, [R2, #12]
        
        MOV     R0, #0
        LDMFD   SP!, {R3-R8, PC}

; sub128: R0=ptr X, R1=ptr Y, R2=dest -> dest = X - Y
sub128:
        STMFD   SP!, {R3-R8, LR}
        ; Unrolled loop to preserve carry flag
        ; For subtraction, C=1 means no borrow
        LDR     R4, [R0, #0]
        LDR     R5, [R1, #0]
        SUBS    R6, R4, R5        ; First subtract, sets carry
        STR     R6, [R2, #0]
        
        LDR     R4, [R0, #4]
        LDR     R5, [R1, #4]
        SBCS    R6, R4, R5        ; Subtract with borrow
        STR     R6, [R2, #4]
        
        LDR     R4, [R0, #8]
        LDR     R5, [R1, #8]
        SBCS    R6, R4, R5        ; Subtract with borrow
        STR     R6, [R2, #8]
        
        LDR     R4, [R0, #12]
        LDR     R5, [R1, #12]
        SBCS    R6, R4, R5        ; Subtract with borrow
        STR     R6, [R2, #12]
        
        MOV     R0, #0
        LDMFD   SP!, {R3-R8, PC}

; shl128_1: R0=src, R1=dest -> dest = src << 1
shl128_1:
        STMFD   SP!, {R2-R6, LR}
        MOV     R2, #0            ; carry in
        MOV     R3, #0
shl_loop:
        CMP     R3, #4
        BGE     shl_done
        LDR     R4, [R0, R3, LSL #2]
        MOV     R5, R4, LSR #31   ; next carry = top bit
        ADD     R4, R4, R4        ; shift left by 1
        ORR     R4, R4, R2        ; insert previous carry (always 0 here)
        STR     R4, [R1, R3, LSL #2]
        MOV     R2, R5
        ADD     R3, R3, #1
        B       shl_loop
shl_done:
        MOV     R0, #0
        LDMFD   SP!, {R2-R6, PC}

; verify_sum_ab: expects SUM == expected_sum_ab
verify_sum_ab:
        LDR     R8, =sum
        LDR     R9, =expect_sum_ab
        B       verify_block
; verify_diff_recovers_a: DIFF == opA
verify_diff_recovers_a:
        LDR     R8, =diff
        LDR     R9, =opA
        B       verify_block
; verify_sum2: sum2 == expect_sum2
verify_sum2:
        LDR     R8, =sum2
        LDR     R9, =expect_sum2
        B       verify_block
; verify_shifted: shifted == expect_shifted
verify_shifted:
        LDR     R8, =shifted
        LDR     R9, =expect_shifted

verify_block:
        MOV     R0, #0            ; assume pass
        MOV     R11, #0
vb_loop:
        CMP     R11, #4
        BGE     vb_done
        LDR     R1, [R8, R11, LSL #2]
        LDR     R2, [R9, R11, LSL #2]
        CMP     R1, R2
        BNE     vb_fail
        ADD     R11, R11, #1
        B       vb_loop
vb_fail:
        MOV     R0, #1
vb_done:
        MOV     PC, LR

; Data -----------------------------------------------------------------

opA:            ; 0xFFFFFFFFFFFFFFFF0000000100000001 (mixed pattern)
        .word   0x00000001, 0x00000001, 0xFFFFFFFF, 0xFFFFFFFF
opB:            ; 0x00000002000000030000000400000005
        .word   0x00000005, 0x00000004, 0x00000003, 0x00000002
opC:            ; 0x11111111222222223333333344444444
        .word   0x44444444, 0x33333333, 0x22222222, 0x11111111

sum:            .space  16
sum2:           .space  16
diff:           .space  16
shifted:        .space  16

; Precomputed expected results
expect_sum_ab:  ; opA + opB (check carry propagation)
        ; limb0: 1 + 5 = 6
        ; limb1: 1 + 4 = 5
        ; limb2: 0xFFFFFFFF + 3 + carry0? none -> 0x00000002 with carry 1
        ; Actually: FFFFFFFF + 3 = 100000002 -> low 32=0x00000002 carry=1
        ; limb3: 0xFFFFFFFF + 2 + carry=1 = 0x00000002 with carry out drop
        .word   0x00000006, 0x00000005, 0x00000002, 0x00000002

expect_sum2:    ; (opA+opB) + opC
        ; add expect_sum_ab + opC
        ; limb0: 6 + 0x44444444 = 0x4444444A
        ; limb1: 5 + 0x33333333 = 0x33333338
        ; limb2: 2 + 0x22222222 = 0x22222224
        ; limb3: 2 + 0x11111111 = 0x11111113
        .word   0x4444444A, 0x33333338, 0x22222224, 0x11111113

expect_shifted: ; (opA+opB) << 1
        ; From expect_sum_ab words
        ; limb0: 0x00000006 <<1 = 0x0000000C
        ; limb1: 0x00000005<<1 = 0x0000000A
        ; limb2: 0x00000002<<1 = 0x00000004
        ; limb3: 0x00000002<<1 = 0x00000004
        .word   0x0000000C, 0x0000000A, 0x00000004, 0x00000004

msg_intro:              .asciz  "[multi_precision_arith] Multi-precision arithmetic tests running"
msg_test1_pass:         .asciz  "[multi_precision_arith] Test 1 PASSED: A + B"
msg_test2_pass: .asciz  "[multi_precision_arith] Test 2 PASSED: (A+B) - B = A"
msg_test3_pass: .asciz  "[multi_precision_arith] Test 3 PASSED: (A+B) + C"
msg_test4_pass: .asciz  "[multi_precision_arith] Test 4 PASSED: Shift"
msg_all_pass:   .asciz  "[multi_precision_arith] ALL TESTS PASSED"
msg_any_fail:   .asciz  "[multi_precision_arith] ONE OR MORE TESTS FAILED"
