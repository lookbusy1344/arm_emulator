; Bubble Sort Algorithm
; Sorts an array of integers in ascending order
; Demonstrates nested loops, array access, and comparison operations

.org 0x0000

main:
    ; Print prompt
    LDR r0, =prompt_msg
    BL print_string

    ; Read array size
    BL read_int
    MOV r4, r0              ; r4 = array size

    ; Validate size
    CMP r4, #1
    BLE error_size
    CMP r4, #20
    BGT error_size

    ; Store size
    LDR r5, =array_size
    STR r4, [r5]

    ; Read array elements
    LDR r0, =input_msg
    BL print_string

    MOV r6, #0              ; r6 = index
    LDR r7, =array          ; r7 = array base address

input_loop:
    CMP r6, r4
    BGE input_done

    ; Print element prompt
    LDR r0, =element_msg
    BL print_string

    ADD r0, r6, #1          ; Element number (1-indexed)
    BL print_int

    LDR r0, =colon_space
    BL print_string

    ; Read element
    BL read_int
    STR r0, [r7, r6, LSL #2]  ; array[i] = input

    ADD r6, r6, #1
    B input_loop

input_done:
    ; Print original array
    LDR r0, =original_msg
    BL print_string
    BL print_array

    ; Perform bubble sort
    BL bubble_sort

    ; Print sorted array
    LDR r0, =sorted_msg
    BL print_string
    BL print_array

    ; Exit
    MOV r0, #0
    BL exit

; Bubble Sort Implementation
bubble_sort:
    PUSH {r4-r9, lr}

    LDR r4, =array_size
    LDR r4, [r4]            ; r4 = n
    LDR r5, =array          ; r5 = array base
    MOV r6, #0              ; r6 = i (outer loop)

outer_loop:
    SUB r7, r4, r6
    SUB r7, r7, #1          ; r7 = n - i - 1
    CMP r6, r4
    BGE sort_done

    MOV r8, #0              ; r8 = j (inner loop)

inner_loop:
    CMP r8, r7
    BGE inner_done

    ; Load array[j] and array[j+1]
    LDR r9, [r5, r8, LSL #2]      ; r9 = array[j]
    ADD r10, r8, #1
    LDR r11, [r5, r10, LSL #2]    ; r11 = array[j+1]

    ; Compare and swap if needed
    CMP r9, r11
    BLE no_swap

    ; Swap array[j] and array[j+1]
    STR r11, [r5, r8, LSL #2]
    STR r9, [r5, r10, LSL #2]

no_swap:
    ADD r8, r8, #1
    B inner_loop

inner_done:
    ADD r6, r6, #1
    B outer_loop

sort_done:
    POP {r4-r9, pc}

; Print array contents
print_array:
    PUSH {r4-r7, lr}

    LDR r4, =array_size
    LDR r4, [r4]            ; r4 = size
    LDR r5, =array          ; r5 = array base
    MOV r6, #0              ; r6 = index

print_loop:
    CMP r6, r4
    BGE print_done

    ; Load and print element
    LDR r0, [r5, r6, LSL #2]
    BL print_int

    ; Print comma if not last element
    ADD r7, r6, #1
    CMP r7, r4
    BGE skip_print_comma

    LDR r0, =comma_space
    BL print_string

skip_print_comma:
    ADD r6, r6, #1
    B print_loop

print_done:
    LDR r0, =newline
    BL print_string
    POP {r4-r7, pc}

error_size:
    LDR r0, =error_size_msg
    BL print_string
    MOV r0, #1
    BL exit

; Helper functions
print_string:
    PUSH {r7, lr}
    MOV r7, #4
    SVC #0
    POP {r7, pc}

print_int:
    PUSH {r7, lr}
    MOV r7, #1
    SVC #0
    POP {r7, pc}

read_int:
    PUSH {r7, lr}
    MOV r7, #3
    SVC #0
    POP {r7, pc}

exit:
    MOV r7, #0
    SVC #0

; Data section
.align 4
prompt_msg:         .asciz "Bubble Sort Demo\nEnter array size (2-20): "
input_msg:          .asciz "\nEnter array elements:\n"
element_msg:        .asciz "Element "
colon_space:        .asciz ": "
original_msg:       .asciz "\nOriginal array: "
sorted_msg:         .asciz "Sorted array:   "
comma_space:        .asciz ", "
error_size_msg:     .asciz "Error: Size must be between 2 and 20\n"
newline:            .asciz "\n"

.align 4
array_size:         .word 0
array:              .space 80       ; Space for 20 integers (20 * 4 bytes)
