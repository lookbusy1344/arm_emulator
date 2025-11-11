; Test program that accesses files within the allowed fsroot
; This should succeed when run with -fsroot pointing to tests/integration/fsroot_test/allowed

.text
.global _start

_start:
    ; Try to open test.txt (should succeed)
    LDR     R0, =filename1
    MOV     R1, #0          ; Read mode
    SWI     #0x10           ; OPEN
    CMP     R0, #0xFFFFFFFF
    BEQ     error

    ; Close the file
    SWI     #0x11           ; CLOSE

    ; Try to open file in subdir (should succeed)
    LDR     R0, =filename2
    MOV     R1, #0          ; Read mode
    SWI     #0x10           ; OPEN
    CMP     R0, #0xFFFFFFFF
    BEQ     error

    ; Close the file
    SWI     #0x11           ; CLOSE

    ; Success - print message and exit
    LDR     R0, =success_msg
    SWI     #0x02           ; WRITE_STRING
    MOV     R0, #0
    SWI     #0x00           ; EXIT

error:
    ; Error - print message and exit with error code
    LDR     R0, =error_msg
    SWI     #0x02           ; WRITE_STRING
    MOV     R0, #1
    SWI     #0x00           ; EXIT

.data
filename1:      .asciz  "test.txt"
filename2:      .asciz  "subdir/data.txt"
success_msg:    .asciz  "File access succeeded\n"
error_msg:      .asciz  "File access failed\n"
