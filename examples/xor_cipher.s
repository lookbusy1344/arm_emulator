; XOR Cipher Implementation
; Implements symmetric encryption/decryption using XOR with a repeating key
; Tests: Bitwise operations, string manipulation, encryption/decryption symmetry
; Non-interactive: Uses predefined plaintext and key for automated testing

.org 0x0000

main:
    ; Print header
    LDR r0, =header_msg
    BL print_string

    ; Display original plaintext
    LDR r0, =plaintext_msg
    BL print_string
    LDR r0, =plaintext
    BL print_string
    LDR r0, =newline
    BL print_string

    ; Display encryption key
    LDR r0, =key_msg
    BL print_string
    LDR r0, =key
    BL print_string
    LDR r0, =newline
    BL print_string

    ; Encrypt the plaintext
    LDR r0, =plaintext
    LDR r1, =ciphertext
    LDR r2, =key
    BL xor_encrypt

    ; Display ciphertext (in hex)
    LDR r0, =encrypted_msg
    BL print_string
    LDR r0, =ciphertext
    LDR r1, =plaintext
    BL print_hex_string

    ; Decrypt the ciphertext (XOR is symmetric)
    LDR r0, =ciphertext
    LDR r1, =decrypted
    LDR r2, =key
    BL xor_encrypt           ; Same operation for decrypt

    ; Display decrypted plaintext
    LDR r0, =decrypted_msg
    BL print_string
    LDR r0, =decrypted
    BL print_string
    LDR r0, =newline
    BL print_string

    ; Verify decryption matches original
    LDR r0, =plaintext
    LDR r1, =decrypted
    BL string_compare
    CMP r0, #1
    BEQ decrypt_success

    ; Decryption failed
    LDR r0, =fail_msg
    BL print_string
    MOV r0, #1
    BL exit

decrypt_success:
    ; Display advanced crypto operations
    LDR r0, =advanced_msg
    BL print_string

    ; Demonstrate key rotation
    LDR r0, =rotation_msg
    BL print_string
    LDR r0, =key
    MOV r1, #3              ; Rotate by 3 positions
    BL rotate_key
    LDR r0, =rotated_key
    BL print_string
    LDR r0, =newline
    BL print_string

    ; Demonstrate byte frequency analysis
    LDR r0, =freq_msg
    BL print_string
    LDR r0, =plaintext
    BL analyze_frequency

    ; Success
    LDR r0, =success_msg
    BL print_string
    MOV r0, #0
    BL exit

; XOR encrypt/decrypt function
; r0 = source string (null-terminated)
; r1 = destination buffer
; r2 = key string (null-terminated)
xor_encrypt:
    PUSH {r4-r9, lr}

    MOV r4, r0              ; r4 = source
    MOV r5, r1              ; r5 = destination
    MOV r6, r2              ; r6 = key

    ; Get key length
    MOV r0, r6
    BL string_length
    MOV r7, r0              ; r7 = key length

    MOV r8, #0              ; r8 = source index
    MOV r9, #0              ; r9 = key index

xor_loop:
    ; Load source byte
    LDRB r0, [r4, r8]

    ; Check for null terminator
    CMP r0, #0
    BEQ xor_done

    ; Load key byte (with wrapping)
    LDRB r1, [r6, r9]

    ; XOR source with key
    EOR r0, r0, r1

    ; Store encrypted byte
    STRB r0, [r5, r8]

    ; Advance indices
    ADD r8, r8, #1
    ADD r9, r9, #1

    ; Wrap key index if needed
    CMP r9, r7
    MOVGE r9, #0

    B xor_loop

xor_done:
    ; Add null terminator to destination
    MOV r0, #0
    STRB r0, [r5, r8]

    POP {r4-r9, pc}

; Rotate key by specified positions
; r0 = key string
; r1 = rotation count
rotate_key:
    PUSH {r4-r8, lr}

    MOV r4, r0              ; r4 = key
    MOV r5, r1              ; r5 = rotation

    ; Get key length
    BL string_length
    MOV r6, r0              ; r6 = length

    ; Normalize rotation (mod length)
    CMP r5, r6
    SUBGE r5, r5, r6

    ; Copy rotated key
    LDR r7, =rotated_key
    MOV r8, #0              ; r8 = index

rotate_loop:
    CMP r8, r6
    BGE rotate_done

    ; Calculate source index: (i + rotation) % length
    ADD r0, r8, r5
    CMP r0, r6
    SUBGE r0, r0, r6

    ; Copy byte
    LDRB r1, [r4, r0]
    STRB r1, [r7, r8]

    ADD r8, r8, #1
    B rotate_loop

rotate_done:
    ; Null terminate
    MOV r0, #0
    STRB r0, [r7, r8]

    POP {r4-r8, pc}

; Analyze byte frequency (simple histogram)
; r0 = string
analyze_frequency:
    PUSH {r4-r7, lr}

    MOV r4, r0              ; r4 = string

    ; Count distinct characters
    MOV r5, #0              ; r5 = index
    MOV r6, #0              ; r6 = distinct count

freq_count_loop:
    LDRB r7, [r4, r5]
    CMP r7, #0
    BEQ freq_count_done

    ; Check if printable
    CMP r7, #32
    BLT freq_skip
    CMP r7, #126
    BGT freq_skip

    ADD r6, r6, #1

freq_skip:
    ADD r5, r5, #1
    B freq_count_loop

freq_count_done:
    ; Print character count
    LDR r0, =char_count_msg
    BL print_string
    MOV r0, r6
    BL print_int
    LDR r0, =newline
    BL print_string

    POP {r4-r7, pc}

; Print string as hexadecimal bytes
; r0 = string, r1 = reference for length
print_hex_string:
    PUSH {r4-r7, lr}

    MOV r4, r0              ; r4 = string

    ; Get length from reference
    MOV r0, r1
    BL string_length
    MOV r5, r0              ; r5 = length

    MOV r6, #0              ; r6 = index

hex_loop:
    CMP r6, r5
    BGE hex_done

    ; Load byte
    LDRB r7, [r4, r6]

    ; Print high nibble
    MOV r0, r7
    LSR r0, r0, #4
    BL print_hex_digit

    ; Print low nibble
    MOV r0, r7
    AND r0, r0, #15
    BL print_hex_digit

    ; Print space between bytes
    LDR r0, =space_msg
    BL print_string

    ADD r6, r6, #1
    B hex_loop

hex_done:
    LDR r0, =newline
    BL print_string
    POP {r4-r7, pc}

; Print single hex digit (0-15)
; r0 = value (0-15)
print_hex_digit:
    PUSH {r4, lr}
    MOV r4, r0

    CMP r4, #10
    BLT print_hex_num

    ; Print A-F
    SUB r4, r4, #10
    ADD r4, r4, #'A'
    B print_hex_char

print_hex_num:
    ; Print 0-9
    ADD r4, r4, #'0'

print_hex_char:
    LDR r0, =hex_buffer
    STRB r4, [r0]
    BL print_string

    POP {r4, pc}

; Get string length
; r0 = string
; Returns: r0 = length (excluding null terminator)
string_length:
    PUSH {r4, r5}
    MOV r4, r0
    MOV r5, #0

len_loop:
    LDRB r0, [r4, r5]
    CMP r0, #0
    BEQ len_done
    ADD r5, r5, #1
    B len_loop

len_done:
    MOV r0, r5
    POP {r4, r5}
    BX lr

; Compare two strings
; r0 = string1, r1 = string2
; Returns: r0 = 1 if equal, 0 if different
string_compare:
    PUSH {r4-r6}
    MOV r4, r0
    MOV r5, r1
    MOV r6, #0

cmp_loop:
    LDRB r0, [r4, r6]
    LDRB r1, [r5, r6]

    CMP r0, r1
    BNE cmp_different

    ; Check if we reached end
    CMP r0, #0
    BEQ cmp_equal

    ADD r6, r6, #1
    B cmp_loop

cmp_equal:
    MOV r0, #1
    POP {r4-r6}
    BX lr

cmp_different:
    MOV r0, #0
    POP {r4-r6}
    BX lr

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
header_msg:         .asciz "XOR Cipher - Symmetric Encryption\n\n"
plaintext_msg:      .asciz "Plaintext:  "
key_msg:            .asciz "Key:        "
encrypted_msg:      .asciz "Encrypted:  "
decrypted_msg:      .asciz "Decrypted:  "
advanced_msg:       .asciz "\nAdvanced Operations:\n"
rotation_msg:       .asciz "Key rotation (3 positions): "
freq_msg:           .asciz "Frequency analysis:\n  "
char_count_msg:     .asciz "Printable characters: "
success_msg:        .asciz "\nEncryption/Decryption cycle verified!\n"
fail_msg:           .asciz "\nError: Decryption mismatch!\n"
space_msg:          .asciz " "
newline:            .asciz "\n"

.align 4
plaintext:          .asciz "ARM Assembly is powerful and efficient!"
key:                .asciz "SECRET"

.align 4
; Buffers
hex_buffer:         .asciz "X"
ciphertext:         .space 64
decrypted:          .space 64
rotated_key:        .space 32
