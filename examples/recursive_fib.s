; Recursive Fibonacci Calculator
; Tests: Deep recursion, stack management, function calls
; Calculates Fibonacci numbers using pure recursion (no memoization)

.text
.global _start

_start:
    ; Calculate fib(20)
    MOV R0, #20
    BL fibonacci

    ; Print result
    MOV R1, #10
    SWI #0x03       ; Write int syscall
    SWI #0x07       ; Write newline

    ; Exit
    MOV R0, #0
    SWI #0x00

; Fibonacci function
; Input: R0 = n
; Output: R0 = fib(n)
; Preserves: R4-R11
fibonacci:
    ; Base case: if n <= 1, return n
    CMP R0, #1
    MOVLE PC, LR    ; Return n if n <= 1

    ; Save registers
    STMFD SP!, {R4, R5, LR}

    ; Save n in R4
    MOV R4, R0

    ; Calculate fib(n-1)
    SUB R0, R4, #1
    BL fibonacci
    MOV R5, R0      ; Save fib(n-1) in R5

    ; Calculate fib(n-2)
    SUB R0, R4, #2
    BL fibonacci

    ; Add results: fib(n) = fib(n-1) + fib(n-2)
    ADD R0, R5, R0

    ; Restore registers and return
    LDMFD SP!, {R4, R5, PC}
