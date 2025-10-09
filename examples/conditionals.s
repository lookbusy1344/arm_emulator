; conditionals.s - Conditional execution demonstration
; Demonstrates: If/else, nested conditions, switch/case

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print intro
        LDR     R0, =msg_intro
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 1: Simple if/else
        LDR     R0, =msg_ex1
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, #15
        MOV     R1, #10
        CMP     R0, R1
        BGT     ex1_greater

        LDR     R0, =msg_not_greater
        B       ex1_done

ex1_greater:
        LDR     R0, =msg_greater

ex1_done:
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 2: If/else if/else chain
        LDR     R0, =msg_ex2
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, #75         ; Score
        BL      grade_calculator

        MOV     R1, R0          ; Save grade character

        LDR     R0, =msg_grade
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R1
        SWI     #0x01           ; WRITE_CHAR
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 3: Nested conditions
        LDR     R0, =msg_ex3
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, #25         ; Age
        MOV     R1, #1          ; Has license (1 = yes)
        BL      can_drive

        CMP     R0, #1
        BEQ     ex3_can

        LDR     R0, =msg_cannot
        B       ex3_done2

ex3_can:
        LDR     R0, =msg_can

ex3_done2:
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE
        SWI     #0x07           ; WRITE_NEWLINE

        ; Example 4: Switch/case (day of week)
        LDR     R0, =msg_ex4
        SWI     #0x02           ; WRITE_STRING

        MOV     R0, #3          ; Wednesday (0=Monday)
        BL      day_name

        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Exit
        MOV     R0, #0
        SWI     #0x00           ; EXIT

; Calculate grade based on score
; Input:  R0 = score (0-100)
; Output: R0 = grade character ('A', 'B', 'C', 'D', 'F')
grade_calculator:
        STMFD   SP!, {R1, LR}

        ; Check for A (90-100)
        CMP     R0, #90
        BGE     grade_a

        ; Check for B (80-89)
        CMP     R0, #80
        BGE     grade_b

        ; Check for C (70-79)
        CMP     R0, #70
        BGE     grade_c

        ; Check for D (60-69)
        CMP     R0, #60
        BGE     grade_d

        ; F (below 60)
        MOV     R0, #'F'
        B       grade_done

grade_a:
        MOV     R0, #'A'
        B       grade_done

grade_b:
        MOV     R0, #'B'
        B       grade_done

grade_c:
        MOV     R0, #'C'
        B       grade_done

grade_d:
        MOV     R0, #'D'

grade_done:
        LDMFD   SP!, {R1, PC}

; Check if person can drive
; Input:  R0 = age
;         R1 = has license (1 = yes, 0 = no)
; Output: R0 = 1 if can drive, 0 otherwise
can_drive:
        STMFD   SP!, {R2, LR}

        ; Must be at least 16
        CMP     R0, #16
        BLT     cannot_drive

        ; Must have license
        CMP     R1, #1
        BNE     cannot_drive

        ; Can drive
        MOV     R0, #1
        B       can_drive_done

cannot_drive:
        MOV     R0, #0

can_drive_done:
        LDMFD   SP!, {R2, PC}

; Get day name from day number
; Input:  R0 = day number (0-6, 0=Monday)
; Output: R0 = address of day name string
day_name:
        STMFD   SP!, {R1-R2, LR}

        ; Bounds check
        CMP     R0, #6
        BGT     day_invalid

        ; Use jump table
        LDR     R1, =day_table
        MOV     R2, R0, LSL #2  ; Multiply by 4 (word size)
        ADD     R1, R1, R2
        LDR     R0, [R1]        ; Load address from table

        LDMFD   SP!, {R1-R2, PC}

day_invalid:
        LDR     R0, =day_unknown
        LDMFD   SP!, {R1-R2, PC}

        ; Data section
msg_intro:
        .asciz  "Conditional Execution Demo"
msg_ex1:
        .asciz  "Example 1: Simple if/else (15 > 10?)"
msg_ex2:
        .asciz  "Example 2: Grade calculator (score = 75)"
msg_ex3:
        .asciz  "Example 3: Nested conditions (age=25, has_license=yes)"
msg_ex4:
        .asciz  "Example 4: Switch/case (day 3): "

msg_greater:
        .asciz  "15 is greater than 10"
msg_not_greater:
        .asciz  "15 is not greater than 10"
msg_grade:
        .asciz  "Grade: "
msg_can:
        .asciz  "Can drive"
msg_cannot:
        .asciz  "Cannot drive"

day_monday:
        .asciz  "Monday"
day_tuesday:
        .asciz  "Tuesday"
day_wednesday:
        .asciz  "Wednesday"
day_thursday:
        .asciz  "Thursday"
day_friday:
        .asciz  "Friday"
day_saturday:
        .asciz  "Saturday"
day_sunday:
        .asciz  "Sunday"
day_unknown:
        .asciz  "Unknown day"

        .align  2
day_table:
        .word   day_monday
        .word   day_tuesday
        .word   day_wednesday
        .word   day_thursday
        .word   day_friday
        .word   day_saturday
        .word   day_sunday
