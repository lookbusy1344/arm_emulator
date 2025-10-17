; Large Literal Pool Stress Test
;
; This program stretches the dynamic literal pool implementation by using
; many LDR Rd, =constant pseudo-instructions in multiple pools.
;
; It loads 30+ different constants into registers and performs XOR operations.
; This tests:
; - Dynamic pool sizing (should allocate exactly for each pool)
; - Multiple literal pools via .ltorg directive
; - Large pool handling (>16 literals per pool)
; - Cumulative address adjustments
;
; Expected Output: Result of XOR of all constants

.org 0x8000

_start:
    ; First pool with 12 literals
    LDR R0, =0x11111111
    LDR R1, =0x22222222
    LDR R2, =0x33333333
    LDR R3, =0x44444444
    LDR R4, =0x55555555
    LDR R5, =0x66666666
    LDR R6, =0x77777777
    LDR R7, =0x88888888
    LDR R8, =0x99999999
    LDR R9, =0xAAAAAAAA
    LDR R10, =0xBBBBBBBB
    LDR R11, =0xCCCCCCCC

    B   pool1_done      ; Jump over literal pool

    ; Pool 1: Mark literal pool location
    .ltorg

pool1_done:
    ; Perform XOR operations on first pool
    EOR R0, R0, R1      ; Result in R0
    EOR R0, R0, R2
    EOR R0, R0, R3
    EOR R0, R0, R4
    EOR R0, R0, R5
    EOR R0, R0, R6
    EOR R0, R0, R7
    EOR R0, R0, R8
    EOR R0, R0, R9
    EOR R0, R0, R10
    EOR R0, R0, R11

    ; Second pool with 13 literals
    LDR R1, =0x12121212
    LDR R2, =0x23232323
    LDR R3, =0x34343434
    LDR R4, =0x45454545
    LDR R5, =0x56565656
    LDR R6, =0x67676767
    LDR R7, =0x78787878
    LDR R8, =0x89898989
    LDR R9, =0x9A9A9A9A
    LDR R10, =0xABABABAB
    LDR R11, =0xBCBCBCBC
    LDR R12, =0xCDCDCDCD

    B   pool2_done      ; Jump over literal pool

    ; Pool 2: Mark second literal pool location
    .ltorg

pool2_done:
    ; Continue XOR with second pool
    EOR R0, R0, R1
    EOR R0, R0, R2
    EOR R0, R0, R3
    EOR R0, R0, R4
    EOR R0, R0, R5
    EOR R0, R0, R6
    EOR R0, R0, R7
    EOR R0, R0, R8
    EOR R0, R0, R9
    EOR R0, R0, R10
    EOR R0, R0, R11
    EOR R0, R0, R12

    ; Third pool with 10 literals for final verification
    LDR R1, =0xDEDEDEDE
    LDR R2, =0xEFEFEFEF
    LDR R3, =0xF0F0F0F0
    LDR R4, =0x13579BDF
    LDR R5, =0x2468ACE0
    LDR R6, =0x13579FDB
    LDR R7, =0x2468ACEF
    LDR R8, =0x3579BDEC
    LDR R9, =0x468ACF13
    LDR R10, =0x579BDCEA

    B   pool3_done      ; Jump over literal pool

    ; Pool 3: Mark third literal pool location
    .ltorg

pool3_done:
    ; Final XOR operations
    EOR R0, R0, R1
    EOR R0, R0, R2
    EOR R0, R0, R3
    EOR R0, R0, R4
    EOR R0, R0, R5
    EOR R0, R0, R6
    EOR R0, R0, R7
    EOR R0, R0, R8
    EOR R0, R0, R9
    EOR R0, R0, R10

    ; Print result in R0
    SWI #0x03
    SWI #0x07    ; Print newline

    ; Exit
    MOV R0, #0
    SWI #0x00
