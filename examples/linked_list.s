; linked_list.s - Linked list implementation
; Demonstrates: Dynamic memory, pointers, data structures

        .org    0x8000          ; Program starts at address 0x8000

_start:
        ; Print intro
        LDR     R0, =msg_intro
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        ; Initialize list head to NULL
        MOV     R4, #0          ; R4 = list head

        ; Insert values: 10, 20, 30, 40, 50
        MOV     R0, #10
        MOV     R1, R4
        BL      insert_front
        MOV     R4, R0          ; Update head

        MOV     R0, #20
        MOV     R1, R4
        BL      insert_front
        MOV     R4, R0

        MOV     R0, #30
        MOV     R1, R4
        BL      insert_front
        MOV     R4, R0

        MOV     R0, #40
        MOV     R1, R4
        BL      insert_front
        MOV     R4, R0

        MOV     R0, #50
        MOV     R1, R4
        BL      insert_front
        MOV     R4, R0

        ; Print the list
        LDR     R0, =msg_list
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        BL      print_list
        SWI     #0x07           ; WRITE_NEWLINE

        ; Count elements
        MOV     R0, R4
        BL      count_list
        MOV     R5, R0

        LDR     R0, =msg_count
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R5
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT
        SWI     #0x07           ; WRITE_NEWLINE

        ; Delete value 30
        LDR     R0, =msg_delete
        SWI     #0x02           ; WRITE_STRING
        SWI     #0x07           ; WRITE_NEWLINE

        MOV     R0, R4
        MOV     R1, #30
        BL      delete_value
        MOV     R4, R0          ; Update head

        ; Print the list again
        LDR     R0, =msg_list
        SWI     #0x02           ; WRITE_STRING
        MOV     R0, R4
        BL      print_list
        SWI     #0x07           ; WRITE_NEWLINE

        ; Free the list
        MOV     R0, R4
        BL      free_list

        ; Exit
        MOV     R0, #0
        SWI     #0x00           ; EXIT

; Insert node at front of list
; Input:  R0 = value
;         R1 = current head
; Output: R0 = new head
insert_front:
        STMFD   SP!, {R1-R3, LR}

        MOV     R2, R0          ; Save value
        MOV     R3, R1          ; Save current head

        ; Allocate node (8 bytes: 4 for data, 4 for next)
        MOV     R0, #8
        SWI     #0x20           ; ALLOCATE

        ; Store data and next pointer
        STR     R2, [R0]        ; node->data = value
        STR     R3, [R0, #4]    ; node->next = old head

        LDMFD   SP!, {R1-R3, PC}

; Print linked list
; Input: R0 = head
print_list:
        STMFD   SP!, {R1-R3, LR}
        MOV     R1, R0          ; R1 = current node

print_loop:
        CMP     R1, #0
        BEQ     print_list_done

        ; Print node data
        LDR     R0, [R1]        ; Load data
        MOV     R2, R1          ; Save current node pointer
        MOV     R1, #10
        SWI     #0x03           ; WRITE_INT

        ; Move to next node
        LDR     R1, [R2, #4]    ; R1 = node->next

        ; Print arrow if not last
        CMP     R1, #0
        BEQ     print_loop

        LDR     R0, =msg_arrow
        SWI     #0x02           ; WRITE_STRING
        B       print_loop

print_list_done:
        LDMFD   SP!, {R1-R3, PC}

; Count list elements
; Input:  R0 = head
; Output: R0 = count
count_list:
        STMFD   SP!, {R1-R2, LR}
        MOV     R1, R0          ; R1 = current
        MOV     R2, #0          ; count = 0

count_loop:
        CMP     R1, #0
        BEQ     count_done

        ADD     R2, R2, #1      ; count++
        LDR     R1, [R1, #4]    ; current = current->next
        B       count_loop

count_done:
        MOV     R0, R2
        LDMFD   SP!, {R1-R2, PC}

; Delete node with specific value
; Input:  R0 = head
;         R1 = value to delete
; Output: R0 = new head
delete_value:
        STMFD   SP!, {R1-R5, LR}

        MOV     R2, R0          ; R2 = current
        MOV     R3, R1          ; R3 = value to delete
        MOV     R4, #0          ; R4 = previous
        MOV     R5, R0          ; R5 = original head

        ; Check if head needs to be deleted
        CMP     R2, #0
        BEQ     delete_done

        LDR     R1, [R2]        ; Load head data
        CMP     R1, R3
        BNE     delete_loop

        ; Delete head
        LDR     R5, [R2, #4]    ; new head = head->next
        MOV     R0, R2
        SWI     #0x21           ; FREE(old head)
        B       delete_done

delete_loop:
        CMP     R2, #0
        BEQ     delete_done

        LDR     R1, [R2]        ; Load current data
        CMP     R1, R3
        BEQ     delete_found

        MOV     R4, R2          ; previous = current
        LDR     R2, [R2, #4]    ; current = current->next
        B       delete_loop

delete_found:
        ; prev->next = current->next
        LDR     R1, [R2, #4]
        STR     R1, [R4, #4]

        ; Free current node
        MOV     R0, R2
        SWI     #0x21           ; FREE

delete_done:
        MOV     R0, R5          ; Return head (original or updated)
        LDMFD   SP!, {R1-R5, PC}

; Free entire list
; Input: R0 = head
free_list:
        STMFD   SP!, {R1-R2, LR}
        MOV     R1, R0          ; R1 = current

free_loop:
        CMP     R1, #0
        BEQ     free_done

        LDR     R2, [R1, #4]    ; R2 = next
        MOV     R0, R1
        SWI     #0x21           ; FREE(current)
        MOV     R1, R2          ; current = next
        B       free_loop

free_done:
        LDMFD   SP!, {R1-R2, PC}

        ; Data section
msg_intro:
        .asciz  "Linked List Operations"
msg_list:
        .asciz  "List: "
msg_arrow:
        .asciz  " -> "
msg_count:
        .asciz  "Element count: "
msg_delete:
        .asciz  "Deleting value 30..."
