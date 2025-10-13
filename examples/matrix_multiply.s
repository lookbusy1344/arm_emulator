; Matrix Multiplication
; Computes C = A × B where A is m×n and B is n×p, resulting in C being m×p
; Tests: 2D array indexing, nested loops, multiply-accumulate operations
; Non-interactive: Uses predefined matrices for automated testing

.org 0x8000

main:
    ; Print header
    LDR r0, =header_msg
    BL print_string

    ; Initialize matrices
    BL init_matrices

    ; Print matrix A
    LDR r0, =matrix_a_msg
    BL print_string
    LDR r0, =matrix_a
    LDR r1, =m_size
    LDR r1, [r1]
    LDR r2, =n_size
    LDR r2, [r2]
    BL print_matrix

    ; Print matrix B
    LDR r0, =matrix_b_msg
    BL print_string
    LDR r0, =matrix_b
    LDR r1, =n_size
    LDR r1, [r1]
    LDR r2, =p_size
    LDR r2, [r2]
    BL print_matrix

    ; Perform multiplication: C = A × B
    BL multiply_matrices

    ; Print result matrix C
    LDR r0, =matrix_c_msg
    BL print_string
    LDR r0, =matrix_c
    LDR r1, =m_size
    LDR r1, [r1]
    LDR r2, =p_size
    LDR r2, [r2]
    BL print_matrix

    ; Print completion message
    LDR r0, =done_msg
    BL print_string

    ; Exit
    MOV r0, #0
    BL exit

; Initialize test matrices
; Matrix A (3×4):
;   1  2  3  4
;   5  6  7  8
;   9 10 11 12
;
; Matrix B (4×2):
;   1  2
;   3  4
;   5  6
;   7  8
;
; Expected result C (3×2):
;   50  60
;  114 140
;  178 220
init_matrices:
    PUSH {r4-r6, lr}

    ; Initialize matrix A (3×4)
    LDR r4, =matrix_a
    MOV r5, #1              ; Counter for values

    MOV r6, #12             ; 3 × 4 = 12 elements
init_a_loop:
    CMP r6, #0
    BLE init_a_done
    STR r5, [r4], #4
    ADD r5, r5, #1
    SUB r6, r6, #1
    B init_a_loop

init_a_done:
    ; Initialize matrix B (4×2)
    LDR r4, =matrix_b
    MOV r5, #1              ; Reset counter

    MOV r6, #8              ; 4 × 2 = 8 elements
init_b_loop:
    CMP r6, #0
    BLE init_b_done
    STR r5, [r4], #4
    ADD r5, r5, #1
    SUB r6, r6, #1
    B init_b_loop

init_b_done:
    POP {r4-r6, pc}

; Multiply matrices: C = A × B
; A is m×n, B is n×p, C is m×p
multiply_matrices:
    PUSH {r4-r12, lr}

    ; Load dimensions
    LDR r4, =m_size
    LDR r4, [r4]            ; r4 = m (rows of A)
    LDR r5, =n_size
    LDR r5, [r5]            ; r5 = n (cols of A, rows of B)
    LDR r6, =p_size
    LDR r6, [r6]            ; r6 = p (cols of B)

    ; Load matrix addresses
    LDR r7, =matrix_a       ; r7 = base address of A
    LDR r8, =matrix_b       ; r8 = base address of B
    LDR r9, =matrix_c       ; r9 = base address of C

    MOV r10, #0             ; r10 = i (row of A/C)

outer_i:
    CMP r10, r4
    BGE mult_done
    MOV r11, #0             ; r11 = j (col of B/C)

outer_j:
    CMP r11, r6
    BGE outer_j_done

    ; Compute C[i][j] = sum of A[i][k] * B[k][j] for k=0 to n-1
    MOV r0, #0              ; r0 = accumulator for C[i][j]
    MOV r12, #0             ; r12 = k

inner_k:
    CMP r12, r5
    BGE inner_k_done

    ; Calculate A[i][k] address: A + (i * n + k) * 4
    MUL r1, r10, r5         ; i * n
    ADD r1, r1, r12         ; i * n + k
    LDR r2, [r7, r1, LSL #2] ; r2 = A[i][k]

    ; Calculate B[k][j] address: B + (k * p + j) * 4
    MUL r1, r12, r6         ; k * p
    ADD r1, r1, r11         ; k * p + j
    LDR r3, [r8, r1, LSL #2] ; r3 = B[k][j]

    ; Multiply and accumulate
    MLA r0, r2, r3, r0      ; r0 = r0 + (A[i][k] * B[k][j])

    ADD r12, r12, #1
    B inner_k

inner_k_done:
    ; Store C[i][j]: C + (i * p + j) * 4
    MUL r1, r10, r6         ; i * p
    ADD r1, r1, r11         ; i * p + j
    STR r0, [r9, r1, LSL #2] ; C[i][j] = accumulator

    ADD r11, r11, #1
    B outer_j

outer_j_done:
    ADD r10, r10, #1
    B outer_i

mult_done:
    POP {r4-r12, pc}

; Print matrix
; r0 = matrix address, r1 = rows, r2 = cols
print_matrix:
    PUSH {r4-r8, lr}

    MOV r4, r0              ; r4 = matrix base
    MOV r5, r1              ; r5 = rows
    MOV r6, r2              ; r6 = cols
    MOV r7, #0              ; r7 = row index

print_row_loop:
    CMP r7, r5
    BGE print_matrix_done

    MOV r8, #0              ; r8 = col index

print_col_loop:
    CMP r8, r6
    BGE print_row_done

    ; Calculate element address: matrix + (row * cols + col) * 4
    MUL r0, r7, r6          ; row * cols
    ADD r0, r0, r8          ; row * cols + col
    LDR r0, [r4, r0, LSL #2] ; Load element

    ; Print element with width padding
    PUSH {r0-r3}
    BL print_int
    LDR r0, =space_msg
    BL print_string
    POP {r0-r3}

    ADD r8, r8, #1
    B print_col_loop

print_row_done:
    ; Print newline after each row
    PUSH {r0-r3}
    LDR r0, =newline
    BL print_string
    POP {r0-r3}

    ADD r7, r7, #1
    B print_row_loop

print_matrix_done:
    POP {r4-r8, pc}

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
header_msg:     .asciz "Matrix Multiplication: C = A × B\n\n"
matrix_a_msg:   .asciz "Matrix A (3×4):\n"
matrix_b_msg:   .asciz "\nMatrix B (4×2):\n"
matrix_c_msg:   .asciz "\nResult C = A × B (3×2):\n"
done_msg:       .asciz "\nMultiplication complete!\n"
space_msg:      .asciz "  "
newline:        .asciz "\n"

.align 4
; Matrix dimensions
m_size:         .word 3     ; rows of A
n_size:         .word 4     ; cols of A, rows of B
p_size:         .word 2     ; cols of B

; Matrix storage (row-major order)
.align 4
matrix_a:       .space 48   ; 3 × 4 × 4 bytes
matrix_b:       .space 32   ; 4 × 2 × 4 bytes
matrix_c:       .space 24   ; 3 × 2 × 4 bytes
