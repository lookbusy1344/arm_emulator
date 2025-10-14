; Sieve of Eratosthenes
; Tests: Complex algorithm, bit manipulation, memory operations
; Finds all prime numbers up to N using the ancient algorithm

.data
limit:
    .word 100       ; Find primes up to 100

sieve:
    .space 101      ; Byte array for sieve (0-100)

.align 4
header_msg:
    .ascii "Prime numbers up to 100:\n"
header_end:

.align 4
.text
.global _start

_start:
    ; Initialize sieve
    LDR R0, =sieve
    LDR R1, =limit
    LDR R1, [R1]
    BL init_sieve

    ; Run sieve algorithm
    LDR R0, =sieve
    LDR R1, =limit
    LDR R1, [R1]
    BL sieve_algorithm

    ; Print header
    MOV R0, #1
    LDR R1, =header_msg
    MOV R2, #25
    SWI #0x13

    ; Print primes
    LDR R0, =sieve
    LDR R1, =limit
    LDR R1, [R1]
    BL print_primes

    ; Exit
    MOV R0, #0
    SWI #0x00

; Initialize sieve - mark all numbers as potentially prime
; Input: R0 = sieve address, R1 = limit
init_sieve:
    STMFD SP!, {R4-R6, LR}

    MOV R4, R0      ; R4 = sieve base
    MOV R5, #2      ; R5 = index (start at 2)
    MOV R6, R1      ; R6 = limit

init_loop:
    CMP R5, R6
    BGT init_done

    ; Mark as prime (1)
    MOV R2, #1
    STRB R2, [R4, R5]

    ADD R5, R5, #1
    B init_loop

init_done:
    LDMFD SP!, {R4-R6, PC}

; Sieve of Eratosthenes algorithm
; Input: R0 = sieve address, R1 = limit
sieve_algorithm:
    STMFD SP!, {R4-R9, LR}

    MOV R4, R0      ; R4 = sieve base
    MOV R5, R1      ; R5 = limit
    MOV R6, #2      ; R6 = current number

outer_sieve:
    ; Calculate R6 * R6
    MUL R7, R6, R6
    CMP R7, R5
    BGT sieve_done

    ; Check if R6 is prime
    LDRB R8, [R4, R6]
    CMP R8, #0
    BEQ next_outer

    ; Mark multiples of R6 as composite
    MUL R7, R6, R6  ; Start at R6^2
    MOV R8, R7      ; R8 = current multiple

mark_multiples:
    CMP R8, R5
    BGT next_outer

    ; Mark as composite (0)
    MOV R9, #0
    STRB R9, [R4, R8]

    ; Next multiple
    ADD R8, R8, R6
    B mark_multiples

next_outer:
    ADD R6, R6, #1
    B outer_sieve

sieve_done:
    LDMFD SP!, {R4-R9, PC}

; Print all prime numbers
; Input: R0 = sieve address, R1 = limit
print_primes:
    STMFD SP!, {R4-R7, LR}

    MOV R4, R0      ; R4 = sieve base
    MOV R5, R1      ; R5 = limit
    MOV R6, #2      ; R6 = current number
    MOV R7, #0      ; R7 = count

print_loop:
    CMP R6, R5
    BGT print_done

    ; Check if prime
    LDRB R2, [R4, R6]
    CMP R2, #0
    BEQ skip_print

    ; Print the prime number
    MOV R0, R6
    SWI #0x03

    ; Print space
    MOV R0, #32     ; ' '
    SWI #0x01

    ; Count and print newline every 10 primes
    ADD R7, R7, #1
    MOV R2, R7
    AND R2, R2, #15  ; Check if count % 16 == 0
    CMP R2, #0
    BNE skip_print

    MOV R0, #10     ; '\n'
    SWI #0x01

skip_print:
    ADD R6, R6, #1
    B print_loop

print_done:
    ; Print final newline
    MOV R0, #10
    SWI #0x01

    LDMFD SP!, {R4-R7, PC}
