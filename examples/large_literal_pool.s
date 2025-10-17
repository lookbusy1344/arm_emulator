; Large Literal Pool Stress Test
;
; This program heavily stresses the dynamic literal pool implementation
; with very large pools that exceed the default 16-literal estimate.
;
; It uses 80+ different constants across 4 pools.
; This tests:
; - Dynamic pool sizing with pools larger than 16 literals
; - Multiple literal pools via .ltorg directive
; - Cumulative address adjustments across many pools
; - Correct address calculation for instructions after large pools
;
; Expected Output: Result of XOR of all constants

.org 0x8000

_start:
    ; ===== POOL 1: 25 literals =====
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
    LDR R12, =0xDDDDDDDD

    ; More constants for pool 1
    MOV R1, R0          ; Save R0
    LDR R0, =0xEEEEEEEE
    EOR R1, R1, R0      ; XOR into R1
    LDR R0, =0xFFFFFFFF
    EOR R1, R1, R0
    LDR R0, =0x10101010
    EOR R1, R1, R0
    LDR R0, =0x20202020
    EOR R1, R1, R0
    LDR R0, =0x30303030
    EOR R1, R1, R0
    LDR R0, =0x40404040
    EOR R1, R1, R0
    LDR R0, =0x50505050
    EOR R1, R1, R0
    LDR R0, =0x60606060
    EOR R1, R1, R0
    LDR R0, =0x70707070
    EOR R1, R1, R0
    LDR R0, =0x80808080
    EOR R1, R1, R0
    LDR R0, =0x90909090
    EOR R1, R1, R0

    MOV R0, R1          ; Result back to R0

    B   pool1_done
    .ltorg

pool1_done:
    ; XOR all from first pool
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

    ; ===== POOL 2: 22 literals =====
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

    ; More for pool 2
    MOV R1, R0
    LDR R0, =0x11223344
    EOR R1, R1, R0
    LDR R0, =0x55667788
    EOR R1, R1, R0
    LDR R0, =0x99AABBCC
    EOR R1, R1, R0
    LDR R0, =0xDDEEFF00
    EOR R1, R1, R0
    LDR R0, =0x13579BDF
    EOR R1, R1, R0
    LDR R0, =0x2468ACE0
    EOR R1, R1, R0
    LDR R0, =0xFEDCBA98
    EOR R1, R1, R0
    LDR R0, =0x76543210
    EOR R1, R1, R0
    LDR R0, =0xAAAA5555
    EOR R1, R1, R0

    MOV R0, R1

    B   pool2_done
    .ltorg

pool2_done:
    ; Continue XOR with second pool
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

    ; ===== POOL 3: 18 literals =====
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

    ; More for pool 3
    MOV R1, R0
    LDR R0, =0x11111000
    EOR R1, R1, R0
    LDR R0, =0x22222000
    EOR R1, R1, R0
    LDR R0, =0x33333000
    EOR R1, R1, R0
    LDR R0, =0x44444000
    EOR R1, R1, R0
    LDR R0, =0x55555000
    EOR R1, R1, R0
    LDR R0, =0x66666000
    EOR R1, R1, R0
    LDR R0, =0x77777000
    EOR R1, R1, R0

    MOV R0, R1

    B   pool3_done
    .ltorg

pool3_done:
    ; Final XOR operations
    EOR R0, R0, R2
    EOR R0, R0, R3
    EOR R0, R0, R4
    EOR R0, R0, R5
    EOR R0, R0, R6
    EOR R0, R0, R7
    EOR R0, R0, R8
    EOR R0, R0, R9
    EOR R0, R0, R10

    ; ===== POOL 4: 20 literals =====
    LDR R1, =0x11110000
    LDR R2, =0x22220000
    LDR R3, =0x33330000
    LDR R4, =0x44440000
    LDR R5, =0x55550000
    LDR R6, =0x66660000
    LDR R7, =0x77770000
    LDR R8, =0x88880000
    LDR R9, =0x99990000
    LDR R10, =0xAAAA0000

    MOV R1, R0
    LDR R0, =0xBBBB0000
    EOR R1, R1, R0
    LDR R0, =0xCCCC0000
    EOR R1, R1, R0
    LDR R0, =0xDDDD0000
    EOR R1, R1, R0
    LDR R0, =0xEEEE0000
    EOR R1, R1, R0
    LDR R0, =0xFFFF0000
    EOR R1, R1, R0
    LDR R0, =0x12340000
    EOR R1, R1, R0
    LDR R0, =0x56780000
    EOR R1, R1, R0
    LDR R0, =0x9ABC0000
    EOR R1, R1, R0
    LDR R0, =0xDEF00000
    EOR R1, R1, R0

    MOV R0, R1

    B   pool4_done
    .ltorg

pool4_done:
    ; Final XOR with pool 4
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
    SWI #0x07

    ; Exit
    MOV R0, #0
    SWI #0x00
