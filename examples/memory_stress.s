; memory_stress.s - Stress test memory operations and addressing modes
; Demonstrates: Complex memory patterns, all addressing modes, edge cases

        .org    0x8000

_start:
        LDR     R0, =msg_intro
        SWI     #0x02
        SWI     #0x07

        ; Allocate test buffer on stack (smaller to avoid stack overflow)
        SUB     SP, SP, #128

        ; Test 1: Sequential writes and reads
        LDR     R0, =msg_test1
        SWI     #0x02
        MOV     R4, SP
        BL      test_sequential
        CMP     R0, #0
        BEQ     test1_pass
        LDR     R0, =msg_fail
        SWI     #0x02
        B       cleanup
test1_pass:
        LDR     R0, =msg_pass
        SWI     #0x02
        SWI     #0x07

        ; Test 2: Strided access pattern
        LDR     R0, =msg_test2
        SWI     #0x02
        MOV     R4, SP
        BL      test_strided
        CMP     R0, #0
        BEQ     test2_pass
        LDR     R0, =msg_fail
        SWI     #0x02
        B       cleanup
test2_pass:
        LDR     R0, =msg_pass
        SWI     #0x02
        SWI     #0x07

        ; Test 3: Byte operations
        LDR     R0, =msg_test3
        SWI     #0x02
        MOV     R4, SP
        BL      test_bytes
        CMP     R0, #0
        BEQ     test3_pass
        LDR     R0, =msg_fail
        SWI     #0x02
        B       cleanup
test3_pass:
        LDR     R0, =msg_pass
        SWI     #0x02
        SWI     #0x07

        ; Test 4: Pre/post indexed addressing
        LDR     R0, =msg_test4
        SWI     #0x02
        MOV     R4, SP
        BL      test_indexed
        CMP     R0, #0
        BEQ     test4_pass
        LDR     R0, =msg_fail
        SWI     #0x02
        B       cleanup
test4_pass:
        LDR     R0, =msg_pass
        SWI     #0x02
        SWI     #0x07

        ; Test 5: Multiple register transfer
        LDR     R0, =msg_test5
        SWI     #0x02
        MOV     R4, SP
        BL      test_ldm_stm
        CMP     R0, #0
        BEQ     test5_pass
        LDR     R0, =msg_fail
        SWI     #0x02
        B       cleanup
test5_pass:
        LDR     R0, =msg_pass
        SWI     #0x02
        SWI     #0x07

        LDR     R0, =msg_done
        SWI     #0x02
        SWI     #0x07

cleanup:
        ; Restore stack
        ADD     SP, SP, #128

        MOV     R0, #0
        SWI     #0x00

; test_sequential - Write and read sequential values
; Input: R4 = buffer address
; Output: R0 = 0 if pass, 1 if fail
test_sequential:
        STMFD   SP!, {R1-R5, LR}
        MOV     R5, R4          ; Save buffer address

        ; Write values 0-31 (32 words = 128 bytes)
        MOV     R1, #0          ; Counter
seq_write:
        CMP     R1, #32
        BGE     seq_write_done
        STR     R1, [R5], #4    ; Post-indexed
        ADD     R1, R1, #1
        B       seq_write

seq_write_done:
        ; Read and verify
        MOV     R5, R4          ; Reset pointer
        MOV     R1, #0          ; Counter
seq_read:
        CMP     R1, #32
        BGE     seq_pass
        LDR     R2, [R5], #4    ; Post-indexed
        CMP     R2, R1
        BNE     seq_fail
        ADD     R1, R1, #1
        B       seq_read

seq_pass:
        MOV     R0, #0
        B       seq_done
seq_fail:
        MOV     R0, #1
seq_done:
        LDMFD   SP!, {R1-R5, PC}

; test_strided - Write and read with stride pattern
; Input: R4 = buffer address
; Output: R0 = 0 if pass, 1 if fail
test_strided:
        STMFD   SP!, {R1-R5, LR}
        MOV     R5, R4

        ; Write every 4th word with pattern
        MOV     R1, #0          ; Counter
        MOV     R2, #100        ; Base value
stride_write:
        CMP     R1, #16
        BGE     stride_write_done
        MOV     R3, R1, LSL #4  ; Offset = i * 16
        ADD     R3, R3, R2      ; Value = 100 + offset
        MOV     R6, R1, LSL #4  ; Calculate address offset
        STR     R3, [R5, R6]    ; Indexed addressing
        ADD     R1, R1, #1
        B       stride_write

stride_write_done:
        ; Read and verify
        MOV     R1, #0
stride_read:
        CMP     R1, #16
        BGE     stride_pass
        MOV     R6, R1, LSL #4
        LDR     R3, [R5, R6]
        MOV     R7, R1, LSL #4
        ADD     R7, R7, #100
        CMP     R3, R7
        BNE     stride_fail
        ADD     R1, R1, #1
        B       stride_read

stride_pass:
        MOV     R0, #0
        B       stride_done
stride_fail:
        MOV     R0, #1
stride_done:
        LDMFD   SP!, {R1-R5, PC}

; test_bytes - Test byte-level operations
; Input: R4 = buffer address
; Output: R0 = 0 if pass, 1 if fail
test_bytes:
        STMFD   SP!, {R1-R5, LR}
        MOV     R5, R4

        ; Write byte pattern
        MOV     R1, #0
byte_write:
        CMP     R1, #128
        BGE     byte_write_done
        AND     R2, R1, #0xFF   ; Byte value
        STRB    R2, [R5, R1]    ; Store byte
        ADD     R1, R1, #1
        B       byte_write

byte_write_done:
        ; Read and verify
        MOV     R1, #0
byte_read:
        CMP     R1, #128
        BGE     byte_pass
        LDRB    R2, [R5, R1]    ; Load byte
        AND     R3, R1, #0xFF
        CMP     R2, R3
        BNE     byte_fail
        ADD     R1, R1, #1
        B       byte_read

byte_pass:
        MOV     R0, #0
        B       byte_done
byte_fail:
        MOV     R0, #1
byte_done:
        LDMFD   SP!, {R1-R5, PC}

; test_indexed - Test pre/post indexed addressing
; Input: R4 = buffer address
; Output: R0 = 0 if pass, 1 if fail
test_indexed:
        STMFD   SP!, {R1-R6, LR}
        MOV     R5, R4

        ; Test post-indexed write
        MOV     R6, R5          ; Copy address
        MOV     R1, #10
        MOV     R2, #20
        MOV     R3, #30
        STR     R1, [R6], #4    ; Post-increment
        STR     R2, [R6], #4
        STR     R3, [R6], #4

        ; Test pre-indexed read
        MOV     R6, R5          ; Reset
        LDR     R1, [R6, #4]!   ; Pre-increment, writeback
        CMP     R1, #20
        BNE     indexed_fail

        ; Verify R6 was updated
        SUB     R7, R6, R5
        CMP     R7, #4
        BNE     indexed_fail

        ; Test post-indexed read
        MOV     R6, R5
        LDR     R1, [R6], #8    ; Post-increment by 8
        CMP     R1, #10
        BNE     indexed_fail
        LDR     R1, [R6]        ; Should read offset 8 (30)
        CMP     R1, #30
        BNE     indexed_fail

indexed_pass:
        MOV     R0, #0
        B       indexed_done
indexed_fail:
        MOV     R0, #1
indexed_done:
        LDMFD   SP!, {R1-R6, PC}

; test_ldm_stm - Test multiple register transfers
; Input: R4 = buffer address
; Output: R0 = 0 if pass, 1 if fail
test_ldm_stm:
        STMFD   SP!, {R1-R8, LR}
        MOV     R5, R4

        ; Setup test values in registers (use values that can be encoded as immediates)
        MOV     R1, #10
        MOV     R2, #20
        MOV     R3, #30
        MOV     R4, #40
        MOV     R6, #50
        MOV     R7, #60
        MOV     R8, #70

        ; Store multiple registers
        STMIA   R5, {R1-R4, R6-R8}

        ; Clear registers
        MOV     R1, #0
        MOV     R2, #0
        MOV     R3, #0
        MOV     R4, #0
        MOV     R6, #0
        MOV     R7, #0
        MOV     R8, #0

        ; Load multiple registers
        LDMIA   R5, {R1-R4, R6-R8}

        ; Verify values
        CMP     R1, #10
        BNE     ldm_fail
        CMP     R2, #20
        BNE     ldm_fail
        CMP     R3, #30
        BNE     ldm_fail
        CMP     R4, #40
        BNE     ldm_fail
        CMP     R6, #50
        BNE     ldm_fail
        CMP     R7, #60
        BNE     ldm_fail
        CMP     R8, #70
        BNE     ldm_fail

ldm_pass:
        MOV     R0, #0
        B       ldm_done
ldm_fail:
        MOV     R0, #1
ldm_done:
        LDMFD   SP!, {R1-R8, PC}

msg_intro:
        .asciz  "Memory Operations Stress Test"
msg_test1:
        .asciz  "Test 1: Sequential access... "
msg_test2:
        .asciz  "Test 2: Strided access... "
msg_test3:
        .asciz  "Test 3: Byte operations... "
msg_test4:
        .asciz  "Test 4: Indexed addressing... "
msg_test5:
        .asciz  "Test 5: LDM/STM operations... "
msg_pass:
        .asciz  "PASS"
msg_fail:
        .asciz  "FAIL"
msg_done:
        .asciz  "All memory tests completed!"
