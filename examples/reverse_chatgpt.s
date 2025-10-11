; reverse a string - secondary example from ChatGPT
        AREA    Reverse, CODE, READONLY
        ENTRY

start
        ; Print prompt
        ADR     R0, prompt
        SWI     &02              ; OS_Write0

        ; Read input string
        ADR     R0, buffer       ; R0 = address of buffer
        MOV     R1, #80          ; R1 = max length (80 chars)
        SWI     &0A              ; OS_ReadLine

        ; Now find end of string (look for NULL)
        ADR     R0, buffer
find_end
        LDRB    R1, [R0], #1
        CMP     R1, #0
        BNE     find_end
        SUB     R0, R0, #2       ; Step back before the NULL (and newline)

        ; Print header
        ADR     R2, msg_out
        SWI     &02              ; OS_Write0

print_back
        LDRB    R1, [R0], #-1    ; Load byte and move backwards
        CMP     R1, #13          ; Stop if CR (Enter)
        BEQ     done
        MOV     R0, R1
        SWI     &00              ; OS_WriteC
        CMP     R0, #' '         ; check still in range
        BGE     print_back

done
        SWI     &11              ; OS_Exit

; -------------------------------
;prompt
;        =       "Enter text: ",0
;msg_out
;        =       "Reversed: ",0
;buffer
;        SPACE   80               ; Reserve 80 bytes for input
;        END
