; State Machine: Email Validator
; Implements a finite state machine to validate email addresses
; Tests: Jump tables, state transitions, character classification, pattern matching
; Non-interactive: Uses predefined test strings for automated testing

.org 0x8000

; State definitions
.equ STATE_START, 0         ; Initial state
.equ STATE_LOCAL, 1         ; Reading local part (before @)
.equ STATE_AT, 2            ; Found @ symbol
.equ STATE_DOMAIN, 3        ; Reading domain part
.equ STATE_DOT, 4           ; Found . in domain
.equ STATE_TLD, 5           ; Reading top-level domain
.equ STATE_ACCEPT, 6        ; Valid email
.equ STATE_REJECT, 7        ; Invalid email

main:
    ; Print header
    LDR r0, =header_msg
    BL print_string

    ; Test case 1: "user@example.com" (valid)
    LDR r0, =test1_msg
    BL print_string
    LDR r0, =test1
    BL validate_email
    BL print_result

    ; Test case 2: "john.doe@company.co.uk" (valid)
    LDR r0, =test2_msg
    BL print_string
    LDR r0, =test2
    BL validate_email
    BL print_result

    ; Test case 3: "invalid@@example.com" (invalid - double @)
    LDR r0, =test3_msg
    BL print_string
    LDR r0, =test3
    BL validate_email
    BL print_result

    ; Test case 4: "missing_at.com" (invalid - no @)
    LDR r0, =test4_msg
    BL print_string
    LDR r0, =test4
    BL validate_email
    BL print_result

    ; Test case 5: "user@no_tld" (invalid - no .)
    LDR r0, =test5_msg
    BL print_string
    LDR r0, =test5
    BL validate_email
    BL print_result

    ; Test case 6: "a@b.c" (valid - minimal)
    LDR r0, =test6_msg
    BL print_string
    LDR r0, =test6
    BL validate_email
    BL print_result

    ; Test case 7: "@example.com" (invalid - no local part)
    LDR r0, =test7_msg
    BL print_string
    LDR r0, =test7
    BL validate_email
    BL print_result

    ; Test case 8: "user@.com" (invalid - domain starts with .)
    LDR r0, =test8_msg
    BL print_string
    LDR r0, =test8
    BL validate_email
    BL print_result

    ; All tests complete
    LDR r0, =done_msg
    BL print_string
    MOV r0, #0
    BL exit

; Validate email address using state machine
; r0 = pointer to null-terminated string
; Returns: r0 = 1 if valid, 0 if invalid
validate_email:
    PUSH {r4-r7, lr}

    MOV r4, r0              ; r4 = string pointer
    MOV r5, #STATE_START    ; r5 = current state
    MOV r6, #0              ; r6 = character index

state_loop:
    ; Load current character
    LDRB r7, [r4, r6]

    ; Check for end of string
    CMP r7, #0
    BEQ check_final_state

    ; Process character based on current state
    CMP r5, #STATE_START
    BEQ process_start
    CMP r5, #STATE_LOCAL
    BEQ process_local
    CMP r5, #STATE_AT
    BEQ process_at
    CMP r5, #STATE_DOMAIN
    BEQ process_domain
    CMP r5, #STATE_DOT
    BEQ process_dot
    CMP r5, #STATE_TLD
    BEQ process_tld

    ; Unknown state - reject
    B reject_email

process_start:
    ; From START, alphanumeric goes to LOCAL
    MOV r0, r7
    BL is_alnum
    CMP r0, #1
    BEQ trans_to_local
    B reject_email

process_local:
    ; In LOCAL state
    CMP r7, #'@'
    BEQ trans_to_at
    ; Alphanumeric or dot stays in LOCAL
    MOV r0, r7
    BL is_alnum
    CMP r0, #1
    BEQ next_char
    CMP r7, #'.'
    BEQ next_char
    B reject_email

process_at:
    ; After @, need alphanumeric for domain
    MOV r0, r7
    BL is_alnum
    CMP r0, #1
    BEQ trans_to_domain
    B reject_email

process_domain:
    ; In DOMAIN state
    CMP r7, #'.'
    BEQ trans_to_dot
    ; Alphanumeric stays in DOMAIN
    MOV r0, r7
    BL is_alnum
    CMP r0, #1
    BEQ next_char
    B reject_email

process_dot:
    ; After dot in domain, need alphanumeric for TLD
    MOV r0, r7
    BL is_alnum
    CMP r0, #1
    BEQ trans_to_tld
    B reject_email

process_tld:
    ; In TLD state
    CMP r7, #'.'
    BEQ trans_to_dot     ; Allow multiple dots (e.g., co.uk)
    ; Alphanumeric stays in TLD
    MOV r0, r7
    BL is_alnum
    CMP r0, #1
    BEQ next_char
    B reject_email

trans_to_local:
    MOV r5, #STATE_LOCAL
    B next_char

trans_to_at:
    MOV r5, #STATE_AT
    B next_char

trans_to_domain:
    MOV r5, #STATE_DOMAIN
    B next_char

trans_to_dot:
    MOV r5, #STATE_DOT
    B next_char

trans_to_tld:
    MOV r5, #STATE_TLD
    B next_char

next_char:
    ADD r6, r6, #1
    B state_loop

check_final_state:
    ; Email is valid only if we end in STATE_TLD
    CMP r5, #STATE_TLD
    BEQ accept_email
    B reject_email

accept_email:
    MOV r0, #1
    POP {r4-r7, pc}

reject_email:
    MOV r0, #0
    POP {r4-r7, pc}

; Check if character is alphanumeric (a-z, A-Z, 0-9)
; r0 = character
; Returns: r0 = 1 if alphanumeric, 0 otherwise
is_alnum:
    ; Check lowercase a-z
    CMP r0, #'a'
    BLT check_upper
    CMP r0, #'z'
    BLE is_alnum_yes

check_upper:
    ; Check uppercase A-Z
    CMP r0, #'A'
    BLT check_digit
    CMP r0, #'Z'
    BLE is_alnum_yes

check_digit:
    ; Check digit 0-9
    CMP r0, #'0'
    BLT is_alnum_no
    CMP r0, #'9'
    BLE is_alnum_yes

is_alnum_no:
    MOV r0, #0
    BX lr

is_alnum_yes:
    MOV r0, #1
    BX lr

; Print validation result
; r0 = result (1=valid, 0=invalid)
print_result:
    PUSH {r4, lr}
    MOV r4, r0

    LDR r0, =result_msg
    BL print_string

    CMP r4, #1
    BEQ print_valid

    LDR r0, =invalid_msg
    BL print_string
    B print_result_done

print_valid:
    LDR r0, =valid_msg
    BL print_string

print_result_done:
    POP {r4, pc}

; Helper functions
print_string:
    PUSH {lr}
    SWI #0x02               ; SWI_WRITE_STRING
    POP {pc}

exit:
    SWI #0x00               ; SWI_EXIT

; Data section
.align 4
header_msg:     .asciz "Email Validator - State Machine\n\n"
test1_msg:      .asciz "Test 1: 'user@example.com'\n"
test2_msg:      .asciz "Test 2: 'john.doe@company.co.uk'\n"
test3_msg:      .asciz "Test 3: 'invalid@@example.com'\n"
test4_msg:      .asciz "Test 4: 'missing_at.com'\n"
test5_msg:      .asciz "Test 5: 'user@no_tld'\n"
test6_msg:      .asciz "Test 6: 'a@b.c'\n"
test7_msg:      .asciz "Test 7: '@example.com'\n"
test8_msg:      .asciz "Test 8: 'user@.com'\n"
result_msg:     .asciz "  Result: "
valid_msg:      .asciz "VALID\n\n"
invalid_msg:    .asciz "INVALID\n\n"
done_msg:       .asciz "All tests completed!\n"

.align 4
; Test strings
test1:          .asciz "user@example.com"
test2:          .asciz "john.doe@company.co.uk"
test3:          .asciz "invalid@@example.com"
test4:          .asciz "missing_at.com"
test5:          .asciz "user@no_tld"
test6:          .asciz "a@b.c"
test7:          .asciz "@example.com"
test8:          .asciz "user@.com"
