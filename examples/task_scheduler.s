; task_scheduler.s - Simple cooperative round-robin task scheduler
; Demonstrates: LDM/STM for context save/restore, multiple stacks,
;               function pointers (indirect branches), state machine,
;               use of SWI for output, conditional execution.
;
; We simulate three tasks:
;   Task 0: Print sequence 'A', 'B', 'C', ... wraps at 'Z'
;   Task 1: Increment 32-bit counter
;   Task 2: Fibonacci generator (32-bit) with overflow detection
;
; Each task yields after N iterations. Scheduler cycles fixed number of rounds
; then prints summaries and exits. Context = R0-R12, LR, CPSR (simplified; CPSR ignored).
;
; Layout:
;   task_stacks: individual downward-growing stacks
;   tcb array: {sp, entry, state}
;
; NOTE: This stresses multiple register save/restore patterns and indirect branches.
;
        .org    0x8000

.equ    NUM_TASKS, 3
.equ    ROUNDS, 10          ; scheduler rounds
.equ    ITER_PER_SLICE, 5

_start:
        LDR     R0, =msg_intro
        SWI     #0x02
        SWI     #0x07

        BL      init_tasks
        BL      run_scheduler

        LDR     R0, =msg_done
        SWI     #0x02
        SWI     #0x07
        MOV     R0, #0
        SWI     #0x00

; init_tasks: initialize task control blocks and stacks
init_tasks:
        STMFD   SP!, {R4-R8, LR}
        LDR     R4, =tcb
        LDR     R5, =task_stacks
        MOV     R6, #0
init_loop:
        CMP     R6, #NUM_TASKS
        BGE     init_done
        ; Calculate: task_stacks + (task_index + 1) * STACK_SIZE
        ; Each task's SP points to the top of its allocated stack region
        ADD     R7, R6, #1           ; task_index + 1
        LDR     R8, =STACK_SIZE
        MUL     R7, R8, R7           ; STACK_SIZE * (task_index + 1)
        ADD     R7, R5, R7           ; task_stacks + offset = top of this task's stack
        ; Write initial SP
        STR     R7, [R4]          ; tcb[i].sp
        ; Entry function pointer
        ADR     R0, task0_entry
        CMP     R6, #0
        BEQ     store_entry
        ADR     R0, task1_entry
        CMP     R6, #1
        BEQ     store_entry
        ADR     R0, task2_entry
store_entry:
        STR     R0, [R4, #4]      ; tcb[i].entry
        MOV     R0, #0
        STR     R0, [R4, #8]      ; tcb[i].state (unused general state)
        ADD     R4, R4, #12
        ADD     R6, R6, #1
        B       init_loop
init_done:
        LDMFD   SP!, {R4-R8, PC}

; run_scheduler: simple round-robin
run_scheduler:
        STMFD   SP!, {R4-R11, LR}
        MOV     R10, #0          ; current task index
        MOV     R11, #0          ; round counter
sched_outer:
        CMP     R11, #ROUNDS
        BGE     sched_done
        MOV     R9, #0           ; tasks processed this round
sched_round:
        CMP     R9, #NUM_TASKS
        BGE     next_round
        ; Load TCB base = tcb + i*12
        LDR     R4, =tcb
        MOV     R5, R10, LSL #2      ; R5 = i * 4
        ADD     R5, R5, R10, LSL #3  ; R5 = i*4 + i*8 = i*12
        ADD     R4, R4, R5
        ; Load SP and entry
        LDR     R6, [R4]         ; saved SP
        LDR     R7, [R4, #4]     ; entry
        ; Restore minimal context (only LR for first run)
        ; If first run we push a fake LR = task_yield_exit label
        CMP     R6, #0           ; (never zero) just placeholder
        ; Switch stack: save current SP, switch to task SP
        MOV     R8, SP
        MOV     SP, R6
        BL      task_slice        ; run slice with R7=entry, R10=task index
        MOV     R6, SP           ; updated task SP
        MOV     SP, R8           ; restore scheduler stack
        STR     R6, [R4]         ; store updated SP
        ; Next task
        ADD     R10, R10, #1
        CMP     R10, #NUM_TASKS
        BLT     skip_mod
        MOV     R10, #0
skip_mod:
        ADD     R9, R9, #1
        B       sched_round
next_round:
        ADD     R11, R11, #1
        B       sched_outer
sched_done:
        ; Print summaries
        BL      print_summaries
        LDMFD   SP!, {R4-R11, PC}

; task_slice: run ITER_PER_SLICE iterations of task entry function
; Inputs: R7 = entry function pointer, R10 = task index
; Uses task private stack for locals
task_slice:
        STMFD   SP!, {R0-R3, R4-R8, LR}
        MOV     R4, #0
slice_loop:
        CMP     R4, #ITER_PER_SLICE
        BGE     slice_done
        MOV     R0, R10          ; pass task id
        BLX     R7               ; call task entry
        ADD     R4, R4, #1
        B       slice_loop
slice_done:
        LDMFD   SP!, {R0-R3, R4-R8, PC}

; Task 0: cyclic alphabet printer
; R0 = task id
task0_entry:
        STMFD   SP!, {R1-R4, LR}
        LDR     R1, =t0_char
        LDRB    R2, [R1]
        SWI     #0x01            ; write char
        ; Advance
        ADD     R2, R2, #1
        CMP     R2, #'Z'
        BLE     t0_store
        MOV     R2, #'A'
 t0_store:
        STRB    R2, [R1]
        LDMFD   SP!, {R1-R4, PC}

; Task 1: increment 32-bit counter
; R0 = task id
task1_entry:
        STMFD   SP!, {R1-R3, LR}
        LDR     R1, =t1_counter
        LDR     R2, [R1]
        ADD     R2, R2, #1
        STR     R2, [R1]
        LDMFD   SP!, {R1-R3, PC}

; Task 2: Fibonacci sequence generator with overflow detect
; R0 = task id
task2_entry:
        STMFD   SP!, {R1-R5, LR}
        LDR     R1, =t2_prev
        LDR     R2, =t2_curr
        LDR     R3, [R1]
        LDR     R4, [R2]
        ADD     R5, R3, R4
        CMP     R5, R4           ; overflow if result < one operand
        BCS     t2_no_of
        ; overflow -> reset sequence
        MOV     R3, #0
        MOV     R4, #1
        STR     R3, [R1]
        STR     R4, [R2]
        B       t2_exit
 t2_no_of:
        STR     R4, [R1]
        STR     R5, [R2]
 t2_exit:
        LDMFD   SP!, {R1-R5, PC}

print_summaries:
        STMFD   SP!, {R4-R7, LR}
        SWI     #0x07
        LDR     R0, =msg_summary
        SWI     #0x02
        SWI     #0x07
        ; Task1 counter
        LDR     R0, =msg_counter
        SWI     #0x02
        LDR     R0, =t1_counter
        LDR     R0, [R0]
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07
        ; Task2 fib current
        LDR     R0, =msg_fib
        SWI     #0x02
        LDR     R0, =t2_curr
        LDR     R0, [R0]
        MOV     R1, #10
        SWI     #0x03
        SWI     #0x07
        LDMFD   SP!, {R4-R7, PC}

; Data & control blocks ------------------------------------------------

.equ STACK_SIZE, 128
.equ TCB_SIZE, 36          ; NUM_TASKS * 12 = 3 * 12
.equ STACKS_SIZE, 384      ; NUM_TASKS * STACK_SIZE = 3 * 128
        .align 4

; TCB array: NUM_TASKS * 12 bytes (sp, entry, state)
tcb:    .space  TCB_SIZE

; Stacks (simple blob). Provide contiguous region.
        .align 4
 task_stacks:
        .space  STACKS_SIZE

; Task private data
        .align 4
 t0_char:       .byte   'A'
                .space  3
 t1_counter:    .word   0
 t2_prev:       .word   0
 t2_curr:       .word   1

; Messages
msg_intro:      .asciz  "[task_scheduler] Starting cooperative scheduler demo"
msg_done:       .asciz  "[task_scheduler] Scheduler finished"
msg_summary:    .asciz  "[task_scheduler] Summary:" 
msg_counter:    .asciz  "  Task1 counter = "
msg_fib:        .asciz  "  Task2 fib current = "
