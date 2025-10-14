; file_io.s - Comprehensive file I/O syscall exercise
; Demonstrates: Open, Write, Seek, Tell, Read, Close, Error handling
; Uses SWI codes: 0x10(Open) 0x13(Write) 0x14(Seek) 0x15(Tell) 0x12(Read) 0x11(Close)
; Falls back to printing an error message if any syscall returns negative.
; Verifies round-trip integrity by comparing written + read buffers.
;
; File contents pattern: incremental bytes 0..255 trimmed to LENGTH
;
        .org    0x8000

.equ    LENGTH, 64
.equ    SEEK_SET, 0

_start:
        LDR     R0, =msg_intro
        SWI     #0x02
        SWI     #0x07

        BL      build_pattern
        BL      write_file
        CMP     R0, #0
        BNE     fail
        BL      read_file
        CMP     R0, #0
        BNE     fail
        BL      verify_buffers
        CMP     R0, #0
        BNE     fail
        LDR     R0, =msg_pass
        SWI     #0x02
        B       done
fail:
        LDR     R0, =msg_fail
        SWI     #0x02
        B       done

done:
        SWI     #0x07
        MOV     R0, #0
        SWI     #0x00

; build_pattern: fill write_buf with ascending bytes
build_pattern:
        STMFD   SP!, {R1-R4, LR}
        LDR     R1, =write_buf
        MOV     R2, #0
bp_loop:
        CMP     R2, #LENGTH
        BGE     bp_done
        STRB    R2, [R1, R2]
        ADD     R2, R2, #1
        B       bp_loop
bp_done:
        MOV     R0, #0
        LDMFD   SP!, {R1-R4, PC}

; write_file: open + write pattern
; Returns R0=0 on success else 1
write_file:
        STMFD   SP!, {R1-R7, LR}
        LDR     R0, =filename
        MOV     R1, #1            ; mode (1=write/create) - mode semantics depend on emulator
        SWI     #0x10             ; Open
        CMP     R0, #0
        BMI     wf_err
        MOV     R4, R0            ; fd
        LDR     R1, =write_buf
        MOV     R2, #LENGTH
        MOV     R0, R4
        SWI     #0x13             ; Write
        CMP     R0, #LENGTH
        BNE     wf_err
        ; Seek to start
        MOV     R0, R4
        MOV     R1, #0            ; offset
        MOV     R2, #SEEK_SET
        SWI     #0x14             ; Seek
        ; Close
        MOV     R0, R4
        SWI     #0x11             ; Close
        MOV     R0, #0
        B       wf_done
wf_err:
        MOV     R0, #1
wf_done:
        LDMFD   SP!, {R1-R7, PC}

; read_file: reopen and read back into read_buf
read_file:
        STMFD   SP!, {R1-R6, LR}
        LDR     R0, =filename
        MOV     R1, #0            ; mode read-only
        SWI     #0x10             ; Open
        CMP     R0, #0
        BMI     rf_err
        MOV     R4, R0            ; fd
        MOV     R0, R4
        LDR     R1, =read_buf
        MOV     R2, #LENGTH
        SWI     #0x12             ; Read
        CMP     R0, #LENGTH
        BNE     rf_err
        ; Tell (optional)
        MOV     R0, R4
        SWI     #0x15             ; Tell
        ; Close
        MOV     R0, R4
        SWI     #0x11
        MOV     R0, #0
        B       rf_done
rf_err:
        MOV     R0, #1
rf_done:
        LDMFD   SP!, {R1-R6, PC}

; verify_buffers: compare write_buf and read_buf
verify_buffers:
        STMFD   SP!, {R1-R5, LR}
        LDR     R1, =write_buf
        LDR     R2, =read_buf
        MOV     R3, #0
vb_loop:
        CMP     R3, #LENGTH
        BGE     vb_pass
        LDRB    R4, [R1, R3]
        LDRB    R5, [R2, R3]
        CMP     R4, R5
        BNE     vb_fail
        ADD     R3, R3, #1
        B       vb_loop
vb_pass:
        MOV     R0, #0
        B       vb_done
vb_fail:
        MOV     R0, #1
vb_done:
        LDMFD   SP!, {R1-R5, PC}

; Data ------------------------------------------------------------------

filename:       .asciz  "test_io.txt"
write_buf:      .space  LENGTH
read_buf:       .space  LENGTH

msg_intro:      .asciz  "[file_io] File I/O round-trip test starting"
msg_pass:       .asciz  "[file_io] PASS"
msg_fail:       .asciz  "[file_io] FAIL"
