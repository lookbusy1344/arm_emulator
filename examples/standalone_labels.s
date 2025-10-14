; Example demonstrating standalone labels
; Labels that appear on their own line with no instruction/directive
; This tests the formatter and xref tool fix

_start:
		MOV R0, #0		; Initialize counter

; Standalone label on its own line
loop1:
		ADD R0, R0, #1		; Increment
		CMP R0, #5		; Compare with 5
		BNE loop1		; Loop if not equal

		SWI #3			; Output the result (5)
		SWI #7			; Newline

; Another standalone label
loop2:
		SUB R0, R0, #1		; Decrement
		CMP R0, #0		; Compare with 0
		BNE loop2		; Loop if not equal

		SWI #3			; Output the result (0)
		SWI #7			; Newline

; Final standalone label
done:
		SWI #0			; Exit
