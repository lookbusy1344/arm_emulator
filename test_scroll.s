.org 0x8000

; Test program to verify scrolling
start:
    MOV R0, #0
    MOV R1, #10
    MOV R2, #0
    MOV R3, #0
    MOV R4, #0
    MOV R5, #0
    MOV R6, #0
    MOV R7, #0
    MOV R8, #0
    MOV R9, #0
    MOV R10, #0
    MOV R11, #0
    MOV R12, #0

loop:
    ADD R0, R0, #1
    CMP R0, R1
    BLT loop

; This BL instruction should scroll the editor when stepped into
    BL read_int
    
    SWI #0x00

; Function far down in the code - should require scrolling
read_int:
    MOV R0, #42
    MOV PC, LR

; More padding to make scrolling necessary
    MOV R0, #0
    MOV R1, #0
    MOV R2, #0
    MOV R3, #0
    MOV R4, #0
    MOV R5, #0
    MOV R6, #0
    MOV R7, #0
    MOV R8, #0
    MOV R9, #0
    MOV R10, #0
    MOV R11, #0
    MOV R12, #0
