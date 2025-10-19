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
test1:
        LDR     R0, =value1     ; R0 = address of value1
        MOV     R1, #42         ; R1 = 42
str_test1:
        STR     R1, [R0]        ; Store 42 to value1 (breakpoint: str_test1)

        ; Test 2: STR with offset
test2:
        LDR     R0, =array      ; R0 = address of array
        MOV     R1, #100
str_test2:
        STR     R1, [R0, #0]    ; Store 100 to array[0] (breakpoint: str_test2)

        MOV     R1, #200
str_test3:
        STR     R1, [R0, #4]    ; Store 200 to array[1] (breakpoint: str_test3)

        MOV     R1, #300
str_test4:
        STR     R1, [R0, #8]    ; Store 300 to array[2] (breakpoint: str_test4)

        ; Test 3: PUSH onto stack (store multiple)
test3:
        MOV     R1, #11
        MOV     R2, #22
        MOV     R3, #33
        MOV     R4, #44
push_test:
        STMFD   SP!, {R1-R4}    ; Push R1-R4 onto stack (breakpoint: push_test)

        ; Test 4: POP from stack (load multiple)
pop_test:
        LDMFD   SP!, {R5-R8}    ; Pop into R5-R8 (breakpoint: pop_test)

        ; Test 5: Store byte
test5:
        LDR     R0, =byte_val
        MOV     R1, #0xFF
strb_test:
        STRB    R1, [R0]        ; Store byte (breakpoint: strb_test)

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
