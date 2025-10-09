; hello.s - Classic "Hello World" program
; Demonstrates: String output, program termination

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print "Hello, World!" message
        LDR     R0, =msg_hello
        SWI     #0x02           ; WRITE_STRING syscall

        ; Print newline
        SWI     #0x07           ; WRITE_NEWLINE syscall

        ; Exit program with code 0
        MOV     R0, #0
        SWI     #0x00           ; EXIT syscall

        ; Data section
msg_hello:
        .asciz  "Hello, World!"
