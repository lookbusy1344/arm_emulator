; test_get_arguments.s - Demonstrates GET_ARGUMENTS syscall (0x32)
; Shows: Retrieving program argument count

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print header
        LDR     R0, =msg_header
        SWI     #0x02           ; WRITE_STRING

        ; Get program arguments
        SWI     #0x32           ; GET_ARGUMENTS syscall
        ; R0 now contains argc
        ; R1 now contains argv pointer (currently 0 in implementation)

        MOV     R4, R0          ; Save argc
        MOV     R5, R1          ; Save argv pointer

        ; Print argument count
        LDR     R0, =msg_argc
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        MOV     R1, #10         ; Decimal format
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Print argv pointer (for demonstration)
        LDR     R0, =msg_argv
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R5
        MOV     R1, #16         ; Hex format
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Check if we have arguments
        CMP     R4, #0
        BEQ     no_args

        ; We have arguments
        LDR     R0, =msg_has_args
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        LDR     R0, =msg_args_suffix
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        B       done

no_args:
        ; No arguments
        LDR     R0, =msg_no_args
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

done:
        ; Print success message
        SWI     #0x07           ; WRITE_NEWLINE
        LDR     R0, =msg_success
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit program with code 0
        MOV     R0, #0
        SWI     #0x00           ; EXIT syscall

        ; Data section
msg_header:
        .asciz  "=== GET_ARGUMENTS Syscall Test ==="
msg_argc:
        .asciz  "Argument count (argc): "
msg_argv:
        .asciz  "Argument pointer (argv): 0x"
msg_has_args:
        .asciz  "Program has "
msg_args_suffix:
        .asciz  " argument(s)"
msg_no_args:
        .asciz  "Program has no arguments"
msg_success:
        .asciz  "GET_ARGUMENTS syscall working - Test PASSED"
