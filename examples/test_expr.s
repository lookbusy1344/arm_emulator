; Test constant expressions

.data
buffer: .space 12

.text
_start:
    LDR r0, =buffer + 12
    SWI #0xF2               ; DUMP_REGISTERS
    MOV r0, #0
    SWI #0x00               ; EXIT
