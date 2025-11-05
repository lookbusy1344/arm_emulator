export const TEST_PROGRAMS = {
  hello: `
    .text
    .global _start
_start:
    LDR R0, =msg
    SWI #0x02          ; WRITE_STRING
    SWI #0x00          ; EXIT

.data
msg: .ascii "Hello, World!\\n"
     .byte 0
`,

  fibonacci: `
    .text
    .global _start
_start:
    MOV R0, #10        ; Calculate 10 Fibonacci numbers
    MOV R1, #0         ; First number
    MOV R2, #1         ; Second number
loop:
    CMP R0, #0
    BEQ done
    MOV R3, R1
    ADD R1, R1, R2
    MOV R2, R3
    SUB R0, R0, #1
    B loop
done:
    SWI #0x00          ; EXIT
`,

  infinite_loop: `
    .text
    .global _start
_start:
    MOV R0, #0
loop:
    ADD R0, R0, #1
    B loop
`,

  arithmetic: `
    .text
    .global _start
_start:
    MOV R0, #10
    MOV R1, #20
    ADD R2, R0, R1     ; R2 = 30
    SUB R3, R1, R0     ; R3 = 10
    MUL R4, R0, R1     ; R4 = 200
    SWI #0x00          ; EXIT
`,
};
