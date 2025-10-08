; String Reversal
; Reads a string and prints it in reverse order
; Example: "Hello" becomes "olleH"

.org 0x0000

main:
    ; Print prompt
    LDR r0, =prompt_msg
    BL print_string

    ; Read string input
    LDR r0, =input_buffer
    MOV r1, #100            ; Max length
    BL read_string

    ; Find string length
    LDR r0, =input_buffer
    BL strlen
    MOV r4, r0              ; r4 = string length

    ; Check if empty
    CMP r4, #0
    BEQ error_empty

    ; Print header
    LDR r0, =result_msg
    BL print_string

    ; Reverse the string in-place
    LDR r5, =input_buffer   ; r5 = start pointer
    ADD r6, r5, r4          ; r6 = end pointer + 1
    SUB r6, r6, #1          ; r6 = last char pointer

reverse_loop:
    ; Check if pointers have met or crossed
    CMP r5, r6
    BGE reverse_done

    ; Swap characters
    LDRB r7, [r5]           ; r7 = *start
    LDRB r8, [r6]           ; r8 = *end
    STRB r8, [r5]           ; *start = r8
    STRB r7, [r6]           ; *end = r7

    ; Move pointers
    ADD r5, r5, #1
    SUB r6, r6, #1
    B reverse_loop

reverse_done:
    ; Print reversed string
    LDR r0, =input_buffer
    BL print_string

    LDR r0, =newline
    BL print_string

    ; Exit
    MOV r0, #0
    BL exit

error_empty:
    LDR r0, =error_empty_msg
    BL print_string
    MOV r0, #1
    BL exit

; Calculate string length
; Input: r0 = string pointer
; Output: r0 = length
strlen:
    PUSH {r4, r5}
    MOV r4, r0              ; r4 = string pointer
    MOV r5, #0              ; r5 = counter

strlen_loop:
    LDRB r0, [r4, r5]       ; Load byte at position
    CMP r0, #0              ; Check for null terminator
    BEQ strlen_done
    ADD r5, r5, #1
    B strlen_loop

strlen_done:
    MOV r0, r5
    POP {r4, r5}
    MOV pc, lr

; Helper functions
print_string:
    PUSH {r7, lr}
    MOV r7, #4
    SVC #0
    POP {r7, pc}

read_string:
    ; r0 = buffer, r1 = max length
    PUSH {r7, lr}
    MOV r7, #5              ; syscall: read_string
    SVC #0
    POP {r7, pc}

exit:
    MOV r7, #0
    SVC #0

; Data section
.align 4
prompt_msg:         .asciz "Enter a string to reverse: "
result_msg:         .asciz "Reversed: "
error_empty_msg:    .asciz "Error: Empty string\n"
newline:            .asciz "\n"

.align 4
input_buffer:       .space 100      ; Buffer for input string
