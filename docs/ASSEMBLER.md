# ARM2 Assembler Directives

This document details the assembler directives and syntax for writing ARM2 assembly programs.

**Related Documentation:**
- [Instruction Set Reference](INSTRUCTIONS.md) - ARM2 CPU instructions and system calls
- [Programming Reference](REFERENCE.md) - Condition codes, addressing modes, shifts, and conventions

---

## Table of Contents
1. [Directive Quick Reference](#directive-quick-reference)
2. [Section Directives](#section-directives)
3. [Symbol Directives](#symbol-directives)
4. [Memory Allocation Directives](#memory-allocation-directives)
5. [Character Literals](#character-literals)
6. [Alignment Directives](#alignment-directives)
7. [Literal Pool Directive](#literal-pool-directive)
8. [Directive Usage Examples](#directive-usage-examples)

---

## Assembler Directives

Assembler directives control how the assembler processes your source code. They don't generate instructions but affect memory layout, symbol definitions, and code organization.

### Directive Quick Reference

| Directive | Category | Description |
|-----------|----------|-------------|
| `.text` | Section | Mark beginning of code section |
| `.data` | Section | Mark beginning of data section |
| `.global` | Symbol | Declare symbol as global/exported |
| `.equ` / `.set` | Symbol | Define a constant value |
| `.org` | Memory | Set assembly origin address |
| `.word` | Data | Allocate 32-bit words (4 bytes each) |
| `.half` | Data | Allocate 16-bit halfwords (2 bytes each) |
| `.byte` | Data | Allocate 8-bit bytes (1 byte each) |
| `.ascii` | String | Allocate string without null terminator |
| `.asciz` / `.string` | String | Allocate null-terminated string |
| `.space` / `.skip` | Memory | Reserve bytes (initialized to zero) |
| `.align` | Alignment | Align to 2^n byte boundary |
| `.balign` | Alignment | Align to specified byte boundary |
| `.ltorg` | Literal Pool | Place literal pool at current location |

### Section Directives
#### .text
**Description:** Marks the beginning of a code section containing executable instructions, directing the assembler to place subsequent statements in the executable program area.
Essential for organizing assembly programs by separating executable code from data, allowing multiple code sections to be interleaved with data sections.
The first `.text` directive typically starts at address 0 unless overridden with `.org`, and multiple text sections can appear throughout the source file.

**Syntax:** `.text`

**Details:**
- Indicates that subsequent lines contain executable code
- Multiple `.text` sections can appear in the same file
- If no `.org` directive has been set, the first `.text` section starts at address 0
- Sections can be interleaved (`.text`, `.data`, `.text`, etc.)
- The assembler tracks addresses sequentially across all sections

**Example:**
```arm
.text
.global _start
_start:
    MOV R0, #1
    BL main
    SWI #0x00

main:
    MOV R0, #42
    MOV PC, LR
```

**Multiple Sections:**
```arm
.text
function1:
    MOV R0, #1
    MOV PC, LR

.data
value: .word 100

.text           ; Second code section
function2:
    LDR R0, =value
    LDR R0, [R0]
    MOV PC, LR
```

#### .data
**Description:** Marks the beginning of a data section for defining initialized variables, constants, strings, and arrays that will be stored in memory.
Used to organize program data separately from executable code, making the assembly source more readable and maintainable.
Like `.text`, multiple data sections can be scattered throughout the source file and will be assembled sequentially based on their position and any `.org` directives.

**Syntax:** `.data`

**Details:**
- Indicates that subsequent lines contain data definitions
- Used for variables, constants, strings, and arrays
- Multiple `.data` sections can appear in the same file
- If no `.org` directive has been set, the first `.data` section starts at address 0
- Data section can be interleaved with `.text` sections
- The assembler tracks addresses sequentially across all sections

**Example:**
```arm
.data
counter:    .word 0
message:    .asciz "Hello, World!\n"
buffer:     .space 256
array:      .word 1, 2, 3, 4, 5
```

**Organized Data Layout:**
```arm
.data
; String constants
msg1:       .asciz "Ready\n"
msg2:       .asciz "Done\n"

; Numeric constants
.align 2
max_val:    .word 1000
min_val:    .word 0

; Arrays
.align 2
lookup:     .word 0, 1, 4, 9, 16, 25

; Buffers
.align 2
input_buf:  .space 512
```

**Note:** In this emulator, `.data` and `.text` sections can be freely interleaved. The assembler tracks addresses sequentially regardless of section type.

### Symbol Directives
#### .global
**Description:** Declares a symbol as globally visible, making it accessible from other object modules during linking in multi-file programs.
Commonly used to export entry points (like `_start` or `main`), public functions, and shared data that need to be referenced from other compilation units.
While this emulator treats all symbols as visible (single-module execution), the directive is accepted for compatibility with standard ARM assembly conventions.

**Syntax:** `.global symbol_name`

**Details:**
- Marks a label or symbol as globally visible
- In a multi-module program, global symbols can be referenced from other files
- Commonly used for entry points (like `_start`) and public functions
- In this single-module emulator, all symbols are visible, but `.global` is still accepted for compatibility
- Multiple symbols can be declared global with separate `.global` directives

**Example:**
```arm
.global _start
.global my_function
.global my_data

.text
_start:
    BL my_function
    SWI #0x00

my_function:
    MOV R0, #42
    MOV PC, LR

.data
my_data:
    .word 100
```

**Common Pattern:**
```arm
; Declare all public symbols at the top
.global _start
.global add_numbers
.global multiply

.text
_start:
    ; Entry point
    MOV R0, #5
    MOV R1, #10
    BL add_numbers
    SWI #0x00

add_numbers:
    ADD R0, R0, R1
    MOV PC, LR

multiply:
    MUL R0, R0, R1
    MOV PC, LR
```

#### .equ / .set
**Description:** Defines a named constant that associates a symbol with a numeric value, functioning like `#define` in C but processed by the assembler.
Improves code readability and maintainability by replacing magic numbers with meaningful names, supporting decimal, hexadecimal, binary, and character literal values.
Constants can reference other previously defined constants, and negative values are supported, making it ideal for configuration values, bit flags, and memory addresses.

**Syntax:** `.equ symbol, value` or `.set symbol, value`

**Details:**
- Creates a named constant that can be used throughout the program
- `.equ` and `.set` are equivalent (both define constants)
- The constant can be used in place of immediate values or addresses
- Values can be decimal, hexadecimal, binary, or character literals
- Constants can reference previously defined constants
- Negative values are supported

**Supported value formats:**
- Decimal: `256`, `-10`, `1000`
- Hexadecimal: `0x100`, `0xFF`, `0xDEADBEEF`
- Binary: `0b11111111`, `0b1010`
- Character literals: `'A'`, `'\n'`
- Expressions: Can reference other constants

**Example:**
```arm
; Basic constants
.equ BUFFER_SIZE, 256
.equ MAX_COUNT, 100
.set STACK_SIZE, 0x1000

; Using constants
.data
buffer:     .space BUFFER_SIZE

.text
MOV R0, #MAX_COUNT
LDR SP, =STACK_SIZE
```

**Hexadecimal and Binary:**
```arm
.equ STATUS_READY,  0b00000001
.equ STATUS_BUSY,   0b00000010
.equ STATUS_ERROR,  0b00000100

.equ PERIPHERAL_BASE, 0x40000000
.equ GPIO_OFFSET,     0x1000

.text
MOV R0, #STATUS_READY
LDR R1, =PERIPHERAL_BASE
```

**Character Literals and Negative Values:**
```arm
.equ NEWLINE,     '\n'
.equ SPACE,       ' '
.equ MINUS_ONE,   -1
.equ NEG_OFFSET,  -16

.text
MOV R0, #NEWLINE
MOV R1, #MINUS_ONE
```

**Referencing Other Constants:**
```arm
.equ KB,          1024
.equ BUFFER_SIZE, 16 * KB    ; 16KB
.equ STACK_TOP,   0x10000
.equ STACK_SIZE,  4 * KB
```

### Memory Allocation Directives
#### .org
**Description:** Sets the current memory address where the assembler will place subsequent instructions and data, effectively controlling the memory layout.
Used to position code at specific addresses (like interrupt vectors), create memory gaps, or organize data at predetermined locations in memory.
Can be used multiple times to create non-contiguous sections, with the first `.org` establishing the program's entry point when no explicit origin is set.

**Syntax:** `.org address`

**Details:**
- Sets the starting memory address for subsequent instructions and data
- Can be used multiple times to relocate code/data segments
- Address can be in decimal, hexadecimal (0x prefix), or binary (0b prefix)
- If not specified, the first section (.text or .data) defaults to address 0
- The first `.org` directive also sets the program's entry point origin

**Example:**
```arm
.org 0x8000        ; Start code at address 0x8000

.text
.org 0x8000
main:
    MOV R0, #1
    BL function
    SWI #0x00

function:
    MOV R0, #42
    MOV PC, LR

.data
.org 0x9000        ; Data section starts at 0x9000
buffer: .space 256
value:  .word 100
```

**Multiple .org Example:**
```arm
.text
.org 0x8000
vector_table:
    B reset_handler
    B irq_handler

.org 0x8100
reset_handler:
    MOV SP, #0x10000
    B main
```

#### .word
**Description:** Allocates and initializes one or more 32-bit words (4 bytes each) in memory with specified values, supporting multiple data formats.
The fundamental data allocation directive for integers, pointers, addresses, and lookup tables in ARM assembly.
Values are stored in little-endian format and can be specified as decimal, hexadecimal, binary, character literals, or label addresses.

**Syntax:** `.word value1, value2, ...`

**Details:**
- Each value is stored as a 32-bit (4-byte) word
- Values can be numbers, character literals, or label addresses
- Multiple values can be specified separated by commas
- Values are stored in little-endian format on ARM
- Commonly used for arrays, lookup tables, and constants

**Supported value formats:**
- Decimal: `42`, `-10`, `1000`
- Hexadecimal: `0x1234`, `0xABCDEF00`, `0xFF`
- Binary: `0b11010101`, `0b1111`
- Character literals: `'A'`, `'\n'`

**Example:**
```arm
.data
; Simple array
array:      .word 10, 20, 30, 40

; Hexadecimal values
table:      .word 0x12345678, 0xABCDEF00, 0xDEADBEEF

; Mixed formats
mixed:      .word 100, 0xFF, 0b11110000, 'A'

; Single value
counter:    .word 0

; Large array
fibonacci:  .word 0, 1, 1, 2, 3, 5, 8, 13, 21, 34
```

#### .half
**Description:** Allocates and initializes one or more 16-bit halfwords (2 bytes each) in memory, useful for 16-bit data types and space-efficient storage.
Commonly used for Unicode characters (UTF-16), short integers, packed data structures, and graphics data like RGB565 color values.
Values larger than 16 bits are truncated, and data is stored in little-endian format like all ARM data.

**Syntax:** `.half value1, value2, ...`

**Details:**
- Each value is stored as a 16-bit (2-byte) halfword
- Values are truncated to 16 bits if larger
- Multiple values can be specified separated by commas
- Values are stored in little-endian format
- Useful for 16-bit data arrays and smaller constants

**Supported value formats:**
- Decimal: `42`, `-10`, `1000`
- Hexadecimal: `0x1234`, `0xFFFF`
- Binary: `0b1101010101010101`
- Character literals: `'A'`, `'0'`

**Example:**
```arm
.data
; 16-bit array
shorts:     .half 100, 200, 300, 400

; Hexadecimal halfwords
colors:     .half 0xF800, 0x07E0, 0x001F  ; RGB565 colors

; Mixed formats
data:       .half 1000, 0x1234, 0b1111000011110000

; Single halfword
port_value: .half 0x5555
```

#### .byte
**Description:** Allocates and initializes one or more 8-bit bytes (1 byte each) in memory, providing the finest granularity for data storage.
Essential for ASCII characters, byte arrays, status flags, bitmap data, and any raw byte sequences that need explicit control.
Values larger than 8 bits are truncated to their lower 8 bits, making this directive perfect for packed data and character manipulation.

**Syntax:** `.byte value1, value2, ...`

**Details:**
- Each value is stored as an 8-bit (1-byte) byte
- Values are truncated to 8 bits if larger
- Multiple values can be specified separated by commas
- Useful for byte arrays, flags, and character data
- Character literals are commonly used with `.byte`

**Supported value formats:**
- Decimal: `42`, `255`, `0`
- Hexadecimal: `0x01`, `0xFF`, `0xAB`
- Binary: `0b11010101`, `0b1111`
- Character literals: `'A'`, `'B'`, `'\n'`, `'\0'`

**Example:**
```arm
.data
; Byte array
bytes:      .byte 0x01, 0x02, 0x03, 0xFF

; Character data (without null terminator)
flags:      .byte 'A', 'B', 'C', 'D'

; Status flags
status:     .byte 0b00000001, 0b00000010, 0b00000100

; Hex dump
dump:       .byte 0xDE, 0xAD, 0xBE, 0xEF

; Mixed formats
mixed:      .byte 65, 0x42, 0b01000011, 'D'  ; All are ASCII letters

; Null-terminated string (manual)
msg:        .byte 'H', 'i', '\n', 0  ; "Hi\n" with null terminator
```

#### .ascii
**Description:** Allocates a string in memory as a sequence of bytes without appending a null terminator, storing the exact characters specified.
Used when you need precise control over string length, when concatenating multiple strings, or when the null terminator will be added manually.
Supports standard escape sequences (\n, \t, \\, etc.) for special characters, with each character stored as a single ASCII byte.

**Syntax:** `.ascii "string"`

**Details:**
- Stores the string bytes without adding a null terminator
- Each character is stored as a single byte (ASCII/UTF-8)
- Useful when you need exact byte sequences or will add null manually
- Supports escape sequences for special characters
- The string length equals the number of characters (escape sequences count as 1)

**Example:**
```arm
.data
; String without null terminator (5 bytes)
msg:        .ascii "Hello"

; Multiple strings can be concatenated
banner:     .ascii "======"
            .ascii " ARM "
            .ascii "======"

; String with escape sequences
formatted:  .ascii "Line1\nLine2\tTabbed"

; Using with manual null terminator
cstring:    .ascii "Manual"
            .byte 0           ; Add null terminator manually
```

#### .asciz / .string
**Description:** Allocates a null-terminated C-style string in memory by storing the specified characters followed by an automatic null byte (0x00).
The standard way to define strings for system calls, C library functions, and any code expecting null-terminated strings.
Both `.asciz` and `.string` are equivalent and interchangeable, with full support for escape sequences for special characters.

**Syntax:** `.asciz "string"` or `.string "string"`

**Details:**
- Stores the string bytes and automatically adds a null terminator (0x00)
- `.asciz` and `.string` are equivalent (both add null terminator)
- String length is (number of characters + 1) for the null byte
- Ideal for C-style strings used with syscalls and string functions
- Supports escape sequences for special characters

**Example:**
```arm
.data
; Null-terminated string (6 bytes: 'H','e','l','l','o',0)
msg:        .asciz "Hello"

; Equivalent to .asciz
prompt:     .string "Enter name: "

; String with newline (syscall-ready)
greeting:   .asciz "Hello, World!\n"

; Multiple strings
error1:     .asciz "File not found"
error2:     .asciz "Access denied"

; Empty string (just the null terminator)
empty:      .asciz ""           ; 1 byte: 0
```

**Escape Sequences:** Both `.ascii` and `.asciz` support standard escape sequences:

| Escape | Description | Hex Value |
|--------|-------------|-----------|
| `\n` | Newline (LF) | 0x0A |
| `\r` | Carriage return (CR) | 0x0D |
| `\t` | Tab | 0x09 |
| `\b` | Backspace | 0x08 |
| `\\` | Backslash | 0x5C |
| `\"` | Double quote | 0x22 |
| `\'` | Single quote | 0x27 |
| `\0` | Null character | 0x00 |

**Escape Sequence Examples:**
```arm
.data
; Multi-line string
greeting:   .asciz "Hello\nWorld\n"

; Windows-style path with backslashes
path:       .asciz "C:\\Users\\Name\\file.txt"

; String with quotes
quoted:     .asciz "He said, \"Hello!\""

; Tab-separated values
tsv:        .asciz "Name\tAge\tCity\n"

; Mixed escape sequences
mixed:      .asciz "Line1\r\nLine2\tTab\0Extra"
```

#### .space / .skip
**Description:** Reserves a specified number of bytes in memory, initializing all bytes to zero, creating uninitialized buffers and arrays.
Both `.space` and `.skip` are equivalent and commonly used for allocating working memory, input/output buffers, stack space, and large data structures.
Size can be specified as immediate values or constants defined with `.equ`, making it easy to allocate memory based on configuration parameters.

**Syntax:** `.space size` or `.skip size`

**Details:**
- Reserves the specified number of bytes in memory
- All bytes are initialized to zero (0x00)
- `.space` and `.skip` are equivalent
- Size can be a decimal, hexadecimal, or binary number
- Size can also reference a constant defined with `.equ`
- Useful for buffers, arrays, and uninitialized data

**Example:**
```arm
.data
; 256-byte buffer (all zeros)
buffer:     .space 256

; 4KB stack space
stack:      .skip 0x1000

; Using constants
.equ BUFFER_SIZE, 512
input_buf:  .space BUFFER_SIZE

; Aligned buffer allocation
.align 2
.equ ARRAY_SIZE, 100
array:      .space ARRAY_SIZE * 4  ; 100 words = 400 bytes

; Multiple buffers
tx_buffer:  .space 128
rx_buffer:  .space 128

; Large memory region
heap:       .space 0x10000    ; 64KB
```

**Usage Pattern with Initialization:**
```arm
.data
; Define buffer size
.equ BUF_SIZE, 256

; Reserve buffer space
read_buffer:    .space BUF_SIZE

; Define pointer to buffer
buffer_ptr:     .word read_buffer

.text
; Use buffer in code
LDR R0, =read_buffer
MOV R1, #BUF_SIZE
BL clear_buffer
```

### Character Literals
Character literals can be used anywhere an immediate value is expected. They are enclosed in single quotes and evaluate to the ASCII/Unicode value of the character.

**Syntax:** `'c'` where c is any character

**Supported in:**
- Immediate operands in data processing instructions
- `.byte` directive values
- `.equ` constant definitions
- Comparison values

**Example:**
```arm
MOV R0, #'A'           ; R0 = 65 (ASCII value of 'A')
CMP R1, #'0'           ; Compare R1 with 48 (ASCII '0')
SUB R2, R2, #' '       ; Subtract space character (32)

.equ NEWLINE, '\n'     ; Define constant from character
.byte 'H', 'i', 0      ; Byte array with characters
```

**Escape Sequences:** Character literals support the same escape sequences as strings:
```arm
MOV R0, #'\n'          ; Newline (10)
MOV R1, #'\t'          ; Tab (9)
MOV R2, #'\\'          ; Backslash (92)
MOV R3, #'\''          ; Single quote (39)
```

### Alignment Directives
#### .align
**Description:** Aligns the current memory address to a power-of-2 byte boundary (2^n), padding with zero bytes as needed to reach the alignment.
Critical for performance optimization and hardware requirements, ensuring data structures start at addresses compatible with load/store instructions and cache line boundaries.
Common uses include word alignment (`.align 2` for 4-byte), double-word alignment (`.align 3` for 8-byte), and cache line alignment (`.align 4` for 16-byte).

**Syntax:** `.align n`

**Details:**
- Aligns to 2^n byte boundary (power of 2)
- Pads with zero bytes to reach the alignment boundary
- Commonly used values:
  - `.align 0` = 1-byte alignment (2^0, no effect)
  - `.align 1` = 2-byte alignment (2^1)
  - `.align 2` = 4-byte alignment (2^2, word alignment)
  - `.align 3` = 8-byte alignment (2^3, double-word)
  - `.align 4` = 16-byte alignment (2^4, cache line)

**Example:**
```arm
.data
.align 2          ; Align to 4-byte boundary (2^2)
value1: .word 100

.byte 1, 2, 3     ; 3 bytes
.align 2          ; Pad with 1 byte to reach 4-byte boundary
value2: .word 200 ; Now word-aligned

.text
.align 2          ; Ensure instructions are word-aligned
function:
    MOV R0, #1
    MOV PC, LR
```

#### .balign
**Description:** Aligns the current memory address to an exact byte boundary specified directly (not as a power of 2), padding with zeros as necessary.
More intuitive than `.align` when you want specific byte alignments like 4, 8, or 16 bytes without calculating powers of 2.
Commonly used for word alignment (4 bytes), double-word alignment (8 bytes), and cache line optimization (16 bytes) with straightforward numeric values.

**Syntax:** `.balign boundary`

**Details:**
- Aligns to the exact byte boundary specified (not a power of 2)
- Pads with zero bytes to reach the alignment boundary
- More intuitive than `.align` for specific byte boundaries
- Common values: 4 (word), 8 (double-word), 16 (cache line)

**Example:**
```arm
.data
.balign 4         ; Align to 4-byte boundary
array: .word 1, 2, 3, 4

.byte 0xFF        ; 1 byte
.balign 4         ; Pad with 3 bytes to reach 4-byte boundary
next_word: .word 0x12345678

.text
.balign 16        ; Align to 16-byte boundary (cache line)
critical_loop:
    ; Performance-critical code
    CMP R0, #0
    BNE critical_loop
```

**Alignment Comparison:**
```arm
; These are equivalent:
.align 2          ; 2^2 = 4 bytes
.balign 4         ; 4 bytes

; These are equivalent:
.align 3          ; 2^3 = 8 bytes
.balign 8         ; 8 bytes
```

### Literal Pool Directive
#### .ltorg
**Description:** Explicitly places a literal pool at the current location, storing 32-bit constants used by `LDR Rd, =value` pseudo-instructions within the ±4095 byte range constraint.
Essential for large programs where automatic literal pool placement at the end might exceed the addressing range, requiring manual pool placement near the instructions that reference them.
The assembler automatically deduplicates identical constants, aligns the pool to 4-byte boundaries, and reserves the necessary space for all accumulated literals.

**Syntax:** `.ltorg`

**Purpose:** Used with the `LDR Rd, =value` pseudo-instruction to control where 32-bit constants are stored in memory

**Details:**
- Literals must be within ±4095 bytes of the LDR instruction
- Multiple `.ltorg` directives can be used in large programs
- Values are automatically deduplicated
- Pool is 4-byte aligned automatically
- If no `.ltorg` is specified, a pool is placed at end of program

**Example:**
```arm
.text
.org 0x8000

main:
    LDR R0, =0x12345678   ; Load large constant
    LDR R1, =0xDEADBEEF   ; Load another constant
    ADD R2, R0, R1
    B end

    .ltorg                ; Place literal pool here

end:
    MOV R0, #0
    SWI #0x00
```

**Multiple Pools Example:**
```arm
section1:
    LDR R0, =0x11111111
    LDR R1, =0x22222222
    .ltorg                ; First pool

section2:
    LDR R2, =0x33333333
    LDR R3, =0x44444444
    .ltorg                ; Second pool
```

### Directive Usage Examples

**Complete program demonstrating all directives:**
```arm
; ============================================
; ARM Assembly Program - All Directives Demo
; ============================================

; Define constants using .equ and .set
.equ BUFFER_SIZE, 256
.equ EXIT_SYSCALL, 0x00
.equ WRITE_STRING, 0x02
.set STACK_SIZE, 0x1000
.equ NEWLINE, '\n'

; Declare global symbols
.global _start
.global process_data

; ============================================
; Code Section
; ============================================
.text
.org 0x8000        ; Set origin to 0x8000

_start:
    ; Initialize stack pointer
    LDR SP, =stack_top

    ; Print welcome message
    LDR R0, =welcome_msg
    SWI #WRITE_STRING

    ; Process some data
    BL process_data

    ; Exit program
    MOV R0, #0
    SWI #EXIT_SYSCALL

; Align function to 4-byte boundary
.align 2
process_data:
    ; Save registers
    PUSH {R4-R6, LR}

    ; Load data array address
    LDR R4, =data_array
    LDR R5, =result
    MOV R6, #0

    ; Sum array values
    LDR R0, [R4]
    LDR R1, [R4, #4]
    LDR R2, [R4, #8]
    ADD R6, R0, R1
    ADD R6, R6, R2

    ; Store result
    STR R6, [R5]

    ; Restore and return
    POP {R4-R6, PC}

; Place literal pool here
.ltorg

; ============================================
; Data Section
; ============================================
.data

; String data with .asciz (null-terminated)
welcome_msg:    .asciz "ARM Emulator Demo\n"
prompt:         .asciz "Enter value: "
done_msg:       .string "Processing complete\n"

; String without null using .ascii
banner:         .ascii "======\n"

; Word data (32-bit)
.align 2
data_array:     .word 10, 20, 30, 40, 50
result:         .word 0
counter:        .word 0

; Halfword data (16-bit)
.align 1
port_values:    .half 0x1234, 0x5678, 0xABCD

; Byte data (8-bit)
status_flags:   .byte 0x01, 0x02, 0x04, 0x08
char_array:     .byte 'A', 'R', 'M', '2'

; Mixed format data
mixed_data:     .word 100, 0xFF, 0b11110000, 'X'

; Reserved buffer space (initialized to zero)
.align 2
read_buffer:    .space BUFFER_SIZE
temp_buffer:    .skip 128

; Stack space
.balign 16      ; Align stack to 16-byte boundary
stack_bottom:   .space STACK_SIZE
stack_top:      ; Label marks top of stack

; ============================================
; Additional Code Section (interleaved)
; ============================================
.text

; Helper function
.align 2
clear_buffer:
    PUSH {R0-R2, LR}
    LDR R0, =read_buffer
    MOV R1, #0
    MOV R2, #BUFFER_SIZE
clear_loop:
    STRB R1, [R0], #1
    SUBS R2, R2, #1
    BNE clear_loop
    POP {R0-R2, PC}

; Final literal pool
.ltorg
```

**Simple program structure:**
```arm
; Constants
.equ EXIT, 0x00

; Entry point
.text
.org 0x8000
.global _start

_start:
    MOV R0, #42
    SWI #EXIT

; Data
.data
value:      .word 100
```

**Mixed code and data:**
```arm
.text
function1:
    MOV R0, R1
    MOV PC, LR

.data
value:      .word 42

.text
function2:
    LDR R0, =value
    LDR R0, [R0]
    MOV PC, LR
```

**Using alignment directives:**
```arm
.data
; Byte data (may be at odd address)
.byte 0x01, 0x02, 0x03

; Align to 4-byte boundary before word
.align 2
word_value: .word 0x12345678

; Align to 16-byte boundary
.balign 16
cache_aligned: .word 1, 2, 3, 4
```

---
