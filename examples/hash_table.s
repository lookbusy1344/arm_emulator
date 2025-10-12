; Hash Table Implementation
; Implements a simple hash table with linear probing for collision resolution
; Tests: Hash functions, modulo operations, collision handling, dynamic lookups
; Non-interactive: Uses predefined key-value pairs for automated testing

.org 0x0000

; Hash table constants
.equ TABLE_SIZE, 16         ; Power of 2 for efficient modulo
.equ EMPTY_KEY, -1          ; Marker for empty slots
.equ MAX_PROBES, 16         ; Maximum probes for linear probing

main:
    ; Print header
    LDR r0, =header_msg
    BL print_string

    ; Initialize hash table
    BL init_hash_table

    ; Insert test key-value pairs
    LDR r0, =insert_msg
    BL print_string

    ; Insert: key=100, value=1000
    MOV r0, #100
    MOV r1, #1000
    BL hash_insert
    CMP r0, #0
    BNE insert_fail
    LDR r0, =inserted_msg
    BL print_string
    MOV r0, #100
    BL print_int
    LDR r0, =arrow_msg
    BL print_string
    MOV r0, #1000
    BL print_int
    LDR r0, =newline
    BL print_string

    ; Insert: key=42, value=420
    MOV r0, #42
    MOV r1, #420
    BL hash_insert
    CMP r0, #0
    BNE insert_fail
    LDR r0, =inserted_msg
    BL print_string
    MOV r0, #42
    BL print_int
    LDR r0, =arrow_msg
    BL print_string
    MOV r0, #420
    BL print_int
    LDR r0, =newline
    BL print_string

    ; Insert: key=17, value=170
    MOV r0, #17
    MOV r1, #170
    BL hash_insert
    CMP r0, #0
    BNE insert_fail
    LDR r0, =inserted_msg
    BL print_string
    MOV r0, #17
    BL print_int
    LDR r0, =arrow_msg
    BL print_string
    MOV r0, #170
    BL print_int
    LDR r0, =newline
    BL print_string

    ; Insert: key=33, value=330 (will collide with key=17)
    MOV r0, #33
    MOV r1, #330
    BL hash_insert
    CMP r0, #0
    BNE insert_fail
    LDR r0, =inserted_msg
    BL print_string
    MOV r0, #33
    BL print_int
    LDR r0, =arrow_msg
    BL print_string
    MOV r0, #330
    BL print_int
    LDR r0, =collision_msg
    BL print_string

    ; Insert: key=255, value=2550
    MOV r0, #255
    MOV r1, #2550
    BL hash_insert
    CMP r0, #0
    BNE insert_fail
    LDR r0, =inserted_msg
    BL print_string
    MOV r0, #255
    BL print_int
    LDR r0, =arrow_msg
    BL print_string
    MOV r0, #2550
    BL print_int
    LDR r0, =newline
    BL print_string

    ; Print hash table
    LDR r0, =table_msg
    BL print_string
    BL print_hash_table

    ; Lookup test keys
    LDR r0, =lookup_msg
    BL print_string

    ; Lookup key=42
    MOV r0, #42
    BL hash_lookup
    CMP r0, #-1
    BEQ lookup_not_found
    PUSH {r0}
    LDR r0, =found_msg
    BL print_string
    MOV r0, #42
    BL print_int
    LDR r0, =arrow_msg
    BL print_string
    POP {r0}
    BL print_int
    LDR r0, =newline
    BL print_string

    ; Lookup key=33 (collision case)
    MOV r0, #33
    BL hash_lookup
    CMP r0, #-1
    BEQ lookup_not_found
    PUSH {r0}
    LDR r0, =found_msg
    BL print_string
    MOV r0, #33
    BL print_int
    LDR r0, =arrow_msg
    BL print_string
    POP {r0}
    BL print_int
    LDR r0, =newline
    BL print_string

    ; Lookup key=99 (not in table)
    MOV r0, #99
    BL hash_lookup
    CMP r0, #-1
    BNE lookup_found
    LDR r0, =not_found_msg
    BL print_string
    MOV r0, #99
    BL print_int
    LDR r0, =newline
    BL print_string

    ; Success
    LDR r0, =success_msg
    BL print_string
    MOV r0, #0
    BL exit

insert_fail:
    LDR r0, =insert_fail_msg
    BL print_string
    MOV r0, #1
    BL exit

lookup_not_found:
    LDR r0, =lookup_fail_msg
    BL print_string
    MOV r0, #1
    BL exit

lookup_found:
    LDR r0, =lookup_fail2_msg
    BL print_string
    MOV r0, #1
    BL exit

; Initialize hash table (mark all slots as empty)
init_hash_table:
    PUSH {r4-r6, lr}

    LDR r4, =hash_keys
    LDR r5, =hash_values
    MOV r6, #TABLE_SIZE

init_loop:
    CMP r6, #0
    BLE init_done

    ; Set key to EMPTY_KEY
    MOV r0, #EMPTY_KEY
    STR r0, [r4], #4

    ; Set value to 0
    MOV r0, #0
    STR r0, [r5], #4

    SUB r6, r6, #1
    B init_loop

init_done:
    POP {r4-r6, pc}

; Hash function: h(key) = key mod TABLE_SIZE
; r0 = key
; Returns: r0 = hash value (0 to TABLE_SIZE-1)
hash_function:
    ; Since TABLE_SIZE is 16 (power of 2), we can use bitwise AND
    AND r0, r0, #15         ; key & (TABLE_SIZE - 1)
    BX lr

; Insert key-value pair into hash table
; r0 = key, r1 = value
; Returns: r0 = 0 on success, -1 on failure (table full)
hash_insert:
    PUSH {r4-r9, lr}

    MOV r4, r0              ; r4 = key
    MOV r5, r1              ; r5 = value

    ; Compute initial hash
    BL hash_function
    MOV r6, r0              ; r6 = hash index

    ; Linear probing
    MOV r7, #0              ; r7 = probe count
    LDR r8, =hash_keys
    LDR r9, =hash_values

insert_probe:
    CMP r7, #MAX_PROBES
    BGE insert_table_full

    ; Check if slot is empty
    LDR r0, [r8, r6, LSL #2]
    CMP r0, #EMPTY_KEY
    BEQ insert_at_slot

    ; Slot occupied, try next slot (linear probing)
    ADD r6, r6, #1
    AND r6, r6, #15         ; Wrap around using modulo
    ADD r7, r7, #1
    B insert_probe

insert_at_slot:
    ; Store key and value
    STR r4, [r8, r6, LSL #2]
    STR r5, [r9, r6, LSL #2]
    MOV r0, #0              ; Success
    POP {r4-r9, pc}

insert_table_full:
    MOV r0, #-1             ; Failure
    POP {r4-r9, pc}

; Lookup value by key
; r0 = key
; Returns: r0 = value if found, -1 if not found
hash_lookup:
    PUSH {r4-r8, lr}

    MOV r4, r0              ; r4 = key

    ; Compute initial hash
    BL hash_function
    MOV r5, r0              ; r5 = hash index

    ; Linear probing
    MOV r6, #0              ; r6 = probe count
    LDR r7, =hash_keys
    LDR r8, =hash_values

lookup_probe:
    CMP r6, #MAX_PROBES
    BGE lookup_not_found_ret

    ; Load key at current slot
    LDR r0, [r7, r5, LSL #2]

    ; Check if slot is empty
    CMP r0, #EMPTY_KEY
    BEQ lookup_not_found_ret

    ; Check if key matches
    CMP r0, r4
    BEQ lookup_found_ret

    ; Try next slot
    ADD r5, r5, #1
    AND r5, r5, #15         ; Wrap around
    ADD r6, r6, #1
    B lookup_probe

lookup_found_ret:
    ; Return value
    LDR r0, [r8, r5, LSL #2]
    POP {r4-r8, pc}

lookup_not_found_ret:
    MOV r0, #-1
    POP {r4-r8, pc}

; Print hash table contents
print_hash_table:
    PUSH {r4-r7, lr}

    LDR r4, =hash_keys
    LDR r5, =hash_values
    MOV r6, #0

print_table_loop:
    CMP r6, #TABLE_SIZE
    BGE print_table_done

    ; Print index
    LDR r0, =index_msg
    BL print_string
    MOV r0, r6
    BL print_int
    LDR r0, =colon_space
    BL print_string

    ; Load key
    LDR r7, [r4, r6, LSL #2]
    CMP r7, #EMPTY_KEY
    BEQ print_empty

    ; Print key -> value
    MOV r0, r7
    BL print_int
    LDR r0, =arrow_msg
    BL print_string
    LDR r0, [r5, r6, LSL #2]
    BL print_int
    LDR r0, =newline
    BL print_string
    B print_table_next

print_empty:
    LDR r0, =empty_msg
    BL print_string

print_table_next:
    ADD r6, r6, #1
    B print_table_loop

print_table_done:
    POP {r4-r7, pc}

; Helper functions
print_string:
    PUSH {lr}
    SWI #0x02               ; SWI_WRITE_STRING
    POP {pc}

print_int:
    PUSH {lr}
    SWI #0x03               ; SWI_WRITE_INT
    POP {pc}

exit:
    SWI #0x00               ; SWI_EXIT

; Data section
.align 4
header_msg:         .asciz "Hash Table Implementation\n"
insert_msg:         .asciz "\nInserting key-value pairs:\n"
inserted_msg:       .asciz "  Inserted: "
collision_msg:      .asciz " (collision resolved)\n"
table_msg:          .asciz "\nHash Table Contents:\n"
lookup_msg:         .asciz "\nLookup Tests:\n"
found_msg:          .asciz "  Found: "
not_found_msg:      .asciz "  Not found: "
success_msg:        .asciz "\nAll operations completed successfully!\n"
insert_fail_msg:    .asciz "\nError: Insert failed (table full)\n"
lookup_fail_msg:    .asciz "\nError: Lookup failed for existing key\n"
lookup_fail2_msg:   .asciz "\nError: Lookup found non-existent key\n"
arrow_msg:          .asciz " -> "
colon_space:        .asciz ": "
index_msg:          .asciz "  ["
empty_msg:          .asciz "empty]\n"
newline:            .asciz "\n"

.align 4
; Hash table storage (parallel arrays)
hash_keys:      .space 64   ; TABLE_SIZE * 4 bytes
hash_values:    .space 64   ; TABLE_SIZE * 4 bytes
