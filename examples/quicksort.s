; Quicksort Algorithm
; Implements the classic divide-and-conquer sorting algorithm
; Tests: Recursive partitioning, stack usage, in-place sorting
; Non-interactive: Uses a predefined array for automated testing

.org 0x8000

main:
    ; Print header
    LDR r0, =header_msg
    BL print_string

    ; Initialize test array
    BL init_array

    ; Print original array
    LDR r0, =original_msg
    BL print_string
    BL print_array

    ; Perform quicksort
    LDR r0, =array          ; r0 = array base
    MOV r1, #0              ; r1 = low (start index)
    LDR r2, =array_size
    LDR r2, [r2]
    SUB r2, r2, #1          ; r2 = high (end index)
    BL quicksort

    ; Print sorted array
    LDR r0, =sorted_msg
    BL print_string
    BL print_array

    ; Verify array is sorted
    BL verify_sorted
    CMP r0, #1
    BEQ sort_success

    ; Sort failed
    LDR r0, =fail_msg
    BL print_string
    MOV r0, #1
    BL exit

sort_success:
    LDR r0, =success_msg
    BL print_string
    MOV r0, #0
    BL exit

; Initialize array with unsorted values
; Test data: [42, 17, 93, 8, 55, 31, 76, 12, 68, 25, 89, 4, 61, 38, 14]
init_array:
    PUSH {r4-r6, lr}

    LDR r4, =array
    LDR r5, =test_data
    LDR r6, =array_size
    LDR r6, [r6]            ; r6 = count

copy_loop:
    CMP r6, #0
    BLE copy_done
    LDR r0, [r5], #4
    STR r0, [r4], #4
    SUB r6, r6, #1
    B copy_loop

copy_done:
    POP {r4-r6, pc}

; Quicksort implementation
; r0 = array base, r1 = low index, r2 = high index
quicksort:
    PUSH {r4-r7, lr}

    ; Base case: if low >= high, return
    CMP r1, r2
    BGE qs_return

    ; Save parameters
    MOV r4, r0              ; r4 = array base
    MOV r5, r1              ; r5 = low
    MOV r6, r2              ; r6 = high

    ; Partition the array
    MOV r0, r4
    MOV r1, r5
    MOV r2, r6
    BL partition            ; r0 = pivot index

    MOV r7, r0              ; r7 = pivot index

    ; Recursively sort left partition: quicksort(array, low, pivot-1)
    MOV r0, r4
    MOV r1, r5
    SUB r2, r7, #1
    BL quicksort

    ; Recursively sort right partition: quicksort(array, pivot+1, high)
    MOV r0, r4
    ADD r1, r7, #1
    MOV r2, r6
    BL quicksort

qs_return:
    POP {r4-r7, pc}

; Partition function (Lomuto partition scheme)
; r0 = array base, r1 = low, r2 = high
; Returns: r0 = pivot index
partition:
    PUSH {r4-r10, lr}

    MOV r4, r0              ; r4 = array base
    MOV r5, r1              ; r5 = low
    MOV r6, r2              ; r6 = high

    ; Choose pivot as last element: pivot = array[high]
    LDR r7, [r4, r6, LSL #2] ; r7 = pivot value

    ; i = low - 1 (index of smaller element)
    SUB r8, r5, #1          ; r8 = i

    ; j = low (scanning index)
    MOV r9, r5              ; r9 = j

partition_loop:
    CMP r9, r6
    BGE partition_done

    ; Load array[j]
    LDR r10, [r4, r9, LSL #2] ; r10 = array[j]

    ; If array[j] < pivot, swap array[i+1] with array[j]
    CMP r10, r7
    BGE skip_swap

    ; Increment i
    ADD r8, r8, #1

    ; Swap array[i] and array[j]
    ; Load array[i]
    LDR r0, [r4, r8, LSL #2]
    LDR r1, [r4, r9, LSL #2]

    ; Store swapped values
    STR r1, [r4, r8, LSL #2]
    STR r0, [r4, r9, LSL #2]

skip_swap:
    ADD r9, r9, #1
    B partition_loop

partition_done:
    ; Place pivot in correct position: swap array[i+1] with array[high]
    ADD r8, r8, #1

    LDR r0, [r4, r8, LSL #2]
    LDR r1, [r4, r6, LSL #2]

    STR r1, [r4, r8, LSL #2]
    STR r0, [r4, r6, LSL #2]

    ; Return pivot index
    MOV r0, r8

    POP {r4-r10, pc}

; Print array contents
print_array:
    PUSH {r4-r6, lr}

    LDR r4, =array
    LDR r5, =array_size
    LDR r5, [r5]
    MOV r6, #0

print_loop:
    CMP r6, r5
    BGE print_done

    LDR r0, [r4, r6, LSL #2]
    BL print_int

    ; Print comma and space if not last element
    ADD r0, r6, #1
    CMP r0, r5
    BGE skip_comma

    LDR r0, =comma_space
    BL print_string

skip_comma:
    ADD r6, r6, #1
    B print_loop

print_done:
    LDR r0, =newline
    BL print_string
    POP {r4-r6, pc}

; Verify array is sorted in ascending order
; Returns: r0 = 1 if sorted, 0 otherwise
verify_sorted:
    PUSH {r4-r7, lr}

    LDR r4, =array
    LDR r5, =array_size
    LDR r5, [r5]
    SUB r5, r5, #1          ; Last index to check
    MOV r6, #0              ; Current index

verify_loop:
    CMP r6, r5
    BGE verify_success

    ; Load array[i] and array[i+1]
    LDR r0, [r4, r6, LSL #2]
    ADD r7, r6, #1
    LDR r1, [r4, r7, LSL #2]

    ; Check if array[i] <= array[i+1]
    CMP r0, r1
    BGT verify_fail

    ADD r6, r6, #1
    B verify_loop

verify_success:
    MOV r0, #1
    POP {r4-r7, pc}

verify_fail:
    MOV r0, #0
    POP {r4-r7, pc}

; Helper functions
print_string:
    PUSH {lr}
    SWI #0x02               ; SWI_WRITE_STRING
    POP {pc}

print_int:
    PUSH {lr}
    SWI #0x03               ; SWI_WRITE_INT
    POP {pc}

exit:
    SWI #0x00               ; SWI_EXIT

; Data section
.align 4
header_msg:     .asciz "Quicksort Algorithm\n\n"
original_msg:   .asciz "Original array: "
sorted_msg:     .asciz "Sorted array:   "
success_msg:    .asciz "\nVerification: Array is correctly sorted!\n"
fail_msg:       .asciz "\nVerification: Sort failed!\n"
comma_space:    .asciz ", "
newline:        .asciz "\n"

.align 4
array_size:     .word 15

; Test data: mix of values to test partitioning
test_data:
    .word 42, 17, 93, 8, 55
    .word 31, 76, 12, 68, 25
    .word 89, 4, 61, 38, 14

.align 4
array:          .space 60   ; Space for 15 integers (15 * 4 bytes)
