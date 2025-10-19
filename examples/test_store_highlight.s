; test_store_highlight.s - Simple test for memory store highlighting in TUI
; This program performs various store operations to test TUI memory highlighting
; Set breakpoints at the STR instructions to see memory locations highlighted

        .org    0x8000

_start:
        ; Print a simple message
        LDR     R0, =msg_start
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Test 1: Simple STR to a known memory location
        LDR     R0, =value1     ; R0 = address of value1
        MOV     R1, #42         ; R1 = 42
        STR     R1, [R0]        ; Store 42 to value1 (BREAKPOINT HERE - line 17)

        ; Test 2: STR with offset
        LDR     R0, =array      ; R0 = address of array
        MOV     R1, #100
        STR     R1, [R0, #0]    ; Store 100 to array[0] (BREAKPOINT HERE - line 22)

        MOV     R1, #200
        STR     R1, [R0, #4]    ; Store 200 to array[1] (BREAKPOINT HERE - line 25)

        MOV     R1, #300
        STR     R1, [R0, #8]    ; Store 300 to array[2] (BREAKPOINT HERE - line 28)

        ; Test 3: PUSH onto stack (store multiple)
        MOV     R1, #11
        MOV     R2, #22
        MOV     R3, #33
        MOV     R4, #44
        STMFD   SP!, {R1-R4}    ; Push R1-R4 onto stack (BREAKPOINT HERE - line 35)

        ; Test 4: POP from stack (load multiple)
        LDMFD   SP!, {R5-R8}    ; Pop into R5-R8 (BREAKPOINT HERE - line 38)

        ; Test 5: Store byte
        LDR     R0, =byte_val
        MOV     R1, #0xFF
        STRB    R1, [R0]        ; Store byte (BREAKPOINT HERE - line 43)

        ; Print success message
        LDR     R0, =msg_done
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit
        MOV     R0, #0
        SWI     #0x00           ; EXIT

; Data section
msg_start:
        .asciz  "Testing memory store highlighting..."
msg_done:
        .asciz  "All stores complete!"

        .align  2
value1:
        .word   0
value2:
        .word   0
array:
        .word   0, 0, 0, 0, 0
byte_val:
        .byte   0
