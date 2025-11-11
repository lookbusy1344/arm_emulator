; Test program that attempts to escape fsroot using ..
; This should HALT the VM with a security error

.text
.global _start

_start:
    ; Try to open ../forbidden.txt (should halt VM)
    LDR     R0, =filename
    MOV     R1, #0          ; Read mode
    SWI     #0x10           ; OPEN

    ; Should never reach here - VM should halt
    LDR     R0, =error_msg
    SWI     #0x02           ; WRITE_STRING
    MOV     R0, #1
    SWI     #0x00           ; EXIT

.data
filename:       .asciz  "../forbidden.txt"
error_msg:      .asciz  "ERROR: VM should have halted!\n"
