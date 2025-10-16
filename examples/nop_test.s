; Test NOP pseudo-instruction
; NOP should encode as MOV R0, R0 and have no effect

.text
.org 0x8000
.global _start

_start:
    MOV R1, #42        ; Set R1 to 42
    NOP                ; No operation - should not change anything
    MOV R2, R1         ; Copy R1 to R2
    NOP                ; Another NOP
    NOP                ; And another

    ; Verify R2 contains 42
    CMP R2, #42
    BNE error

    ; Success - exit with code 0
    MOV R0, #0
    SWI #0x00

error:
    ; Error - exit with code 1
    MOV R0, #1
    SWI #0x00
