; string_copy_manual.s - Manual string copy program
; Reads a string from input and copies it to another memory location byte-by-byte
; Demonstrates: Memory addressing, byte-level operations, loops

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print prompt
        LDR     R0, =msg_prompt
        SWI     #0x02           ; WRITE_STRING

        ; Read string into source_buffer
        LDR     R0, =src_buffer
        MOV     R1, #100        ; Max length
        SWI     #0x05           ; READ_STRING

        ; Initialize pointers for copy
        LDR     R1, =src_buffer ; Source pointer
        LDR     R2, =dst_buffer ; Destination pointer

copy_loop:
        LDRB    R3, [R1], #1    ; Load byte from source, increment pointer
        STRB    R3, [R2], #1    ; Store byte to dest, increment pointer
        CMP     R3, #0          ; Check for null terminator
        BNE     copy_loop       ; Continue if not null

        ; Print confirmation
        LDR     R0, =msg_result
        SWI     #0x02           ; WRITE_STRING

        ; Print the copied string from destination buffer
        LDR     R0, =dst_buffer
        SWI     #0x02           ; WRITE_STRING

        ; Print newline
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit
        MOV     R0, #0
        SWI     #0x00           ; EXIT

        ; Data section
msg_prompt:
        .asciz  "Enter a string to copy: "
msg_result:
        .asciz  "Copied string: "

        .align  4
src_buffer:
        .space  100
dst_buffer:
        .space  100
