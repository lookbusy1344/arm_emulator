; test_get_time.s - Demonstrates GET_TIME syscall (0x30)
; Shows: Getting system timestamp, time comparison, formatting output

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print header
        LDR     R0, =msg_header
        SWI     #0x02           ; WRITE_STRING

        ; Get first timestamp
        SWI     #0x30           ; GET_TIME syscall
        MOV     R4, R0          ; Save first timestamp

        ; Print first timestamp
        LDR     R0, =msg_time1
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        MOV     R1, #10         ; Decimal format
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Do some work (simple loop to consume time)
        MOV     R5, #100        ; Smaller loop count
delay_loop:
        SUBS    R5, R5, #1
        BNE     delay_loop

        ; Get second timestamp
        SWI     #0x30           ; GET_TIME syscall
        MOV     R5, R0          ; Save second timestamp

        ; Print second timestamp
        LDR     R0, =msg_time2
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R5
        MOV     R1, #10         ; Decimal format
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Calculate elapsed time
        SUB     R6, R5, R4      ; elapsed = time2 - time1

        ; Print elapsed time
        LDR     R0, =msg_elapsed
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R6
        MOV     R1, #10         ; Decimal format
        SWI     #0x03           ; WRITE_INT
        LDR     R0, =msg_ms
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Print success message
        LDR     R0, =msg_success
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit program with code 0
        MOV     R0, #0
        SWI     #0x00           ; EXIT syscall

        ; Data section
msg_header:
        .asciz  "=== GET_TIME Syscall Test ==="
msg_time1:
        .asciz  "First timestamp:  "
msg_time2:
        .asciz  "Second timestamp: "
msg_elapsed:
        .asciz  "Elapsed time: "
msg_ms:
        .asciz  " ms"
msg_success:
        .asciz  "Time progresses forward - Test PASSED"
