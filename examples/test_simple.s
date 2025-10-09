; Simple test program
.org 0x8000

_start:
    MOV R0, #42
    MOV R1, #10
    ADD R2, R0, R1
    SWI #0x00
