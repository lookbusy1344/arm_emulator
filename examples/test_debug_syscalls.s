; test_debug_syscalls.s - Demonstrates debugging syscalls (0xF0-0xF4)
; Shows: DEBUG_PRINT, DUMP_REGISTERS, DUMP_MEMORY, ASSERT

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print header
        LDR     R0, =msg_header
        SWI     #0x02           ; WRITE_STRING

        ; Test 1: DEBUG_PRINT (0xF0)
        LDR     R0, =msg_test1
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =msg_debug1
        SWI     #0xF0           ; DEBUG_PRINT syscall
        LDR     R0, =msg_done
        SWI     #0x02           ; WRITE_STRING

        ; Test 2: Set up some registers for DUMP_REGISTERS
        LDR     R0, =msg_test2
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, #42         ; Set R0 = 42
        MOV     R1, #100        ; Set R1 = 100
        MOV     R2, #255        ; Set R2 = 255
        MOV     R3, #222        ; Set R3 = 222
        MOV     R4, #111        ; Set R4 = 111

        ; DUMP_REGISTERS (0xF2)
        SWI     #0xF2           ; DUMP_REGISTERS syscall

        LDR     R0, =msg_done
        SWI     #0x02           ; WRITE_STRING

        ; Test 3: Prepare memory for DUMP_MEMORY
        LDR     R0, =msg_test3
        SWI     #0x02           ; WRITE_STRING

        ; Write test pattern to memory
        LDR     R5, =test_data
        MOV     R6, #0x00
        STRB    R6, [R5, #0]
        MOV     R6, #0x11
        STRB    R6, [R5, #1]
        MOV     R6, #0x22
        STRB    R6, [R5, #2]
        MOV     R6, #0x33
        STRB    R6, [R5, #3]
        MOV     R6, #0x44
        STRB    R6, [R5, #4]
        MOV     R6, #0x55
        STRB    R6, [R5, #5]
        MOV     R6, #0x66
        STRB    R6, [R5, #6]
        MOV     R6, #0x77
        STRB    R6, [R5, #7]
        MOV     R6, #0x88
        STRB    R6, [R5, #8]
        MOV     R6, #0x99
        STRB    R6, [R5, #9]
        MOV     R6, #0xAA
        STRB    R6, [R5, #10]
        MOV     R6, #0xBB
        STRB    R6, [R5, #11]
        MOV     R6, #0xCC
        STRB    R6, [R5, #12]
        MOV     R6, #0xDD
        STRB    R6, [R5, #13]
        MOV     R6, #0xEE
        STRB    R6, [R5, #14]
        MOV     R6, #0xFF
        STRB    R6, [R5, #15]

        ; DUMP_MEMORY (0xF3)
        LDR     R0, =test_data  ; Address
        MOV     R1, #16         ; Length (16 bytes)
        SWI     #0xF3           ; DUMP_MEMORY syscall

        LDR     R0, =msg_done
        SWI     #0x02           ; WRITE_STRING

        ; Test 4: ASSERT with true condition (0xF4)
        LDR     R0, =msg_test4
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, #1          ; True condition
        LDR     R1, =msg_assert_pass
        SWI     #0xF4           ; ASSERT syscall (should not halt)

        LDR     R0, =msg_assert_ok
        SWI     #0x02           ; WRITE_STRING

        ; Test 5: More debug messages
        LDR     R0, =msg_test5
        SWI     #0x02           ; WRITE_STRING
        LDR     R0, =msg_debug2
        SWI     #0xF0           ; DEBUG_PRINT syscall
        LDR     R0, =msg_debug3
        SWI     #0xF0           ; DEBUG_PRINT syscall
        LDR     R0, =msg_done
        SWI     #0x02           ; WRITE_STRING

        ; Print final success message
        SWI     #0x07           ; WRITE_NEWLINE
        LDR     R0, =msg_success
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit program with code 0
        MOV     R0, #0
        SWI     #0x00           ; EXIT syscall

        ; Data section
msg_header:
        .asciz  "=== Debug Syscalls Test (0xF0-0xF4) ==="
msg_test1:
        .asciz  "Test 1: DEBUG_PRINT... "
msg_test2:
        .asciz  "Test 2: DUMP_REGISTERS...\n"
msg_test3:
        .asciz  "Test 3: DUMP_MEMORY...\n"
msg_test4:
        .asciz  "Test 4: ASSERT (pass)... "
msg_test5:
        .asciz  "Test 5: Multiple DEBUG_PRINT... "
msg_done:
        .asciz  "DONE\n"
msg_debug1:
        .asciz  "This is a debug message from DEBUG_PRINT syscall"
msg_debug2:
        .asciz  "Debug message 2: Everything is working"
msg_debug3:
        .asciz  "Debug message 3: All debug syscalls functional"
msg_assert_pass:
        .asciz  "Assertion should pass"
msg_assert_ok:
        .asciz  "PASSED\n"
msg_success:
        .asciz  "All debug syscalls working - Test PASSED"

        ; Reserve space for test data
        .align  4
test_data:
        .skip   16              ; 16 bytes for memory dump test
