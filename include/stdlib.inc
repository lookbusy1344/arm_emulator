; ARM2 Emulator Standard Library
; Helper macros for system calls
; Include this file at the top of your assembly programs: .include "include/stdlib.inc"

; =============================================================================
; CONSOLE I/O MACROS
; =============================================================================

; Exit program with code in r0
.macro exit
    SWI #0x00
.endm

; Write character in r0 to console
.macro write_char
    SWI #0x01
.endm

; Write null-terminated string at address in r0
.macro write_string
    SWI #0x02
.endm

; Write integer in r0 in base r1 (default base 10)
.macro write_int
    SWI #0x03
.endm

; Read single character into r0
.macro read_char
    SWI #0x04
.endm

; Read string into buffer at r0, max length r1
.macro read_string
    SWI #0x05
.endm

; Read integer into r0, error flag in r1
.macro read_int
    SWI #0x06
.endm

; Write newline to console
.macro write_newline
    SWI #0x07
.endm

; =============================================================================
; FILE OPERATION MACROS
; =============================================================================

; Open file at address r0, mode r1 (0=read, 1=write, 2=append)
; Returns file descriptor in r0 (0xFFFFFFFF on error)
.macro file_open
    SWI #0x10
.endm

; Close file descriptor in r0
.macro file_close
    SWI #0x11
.endm

; Read from fd r0 into buffer r1, length r2
; Returns bytes read in r0
.macro file_read
    SWI #0x12
.endm

; Write to fd r0 from buffer r1, length r2
; Returns bytes written in r0
.macro file_write
    SWI #0x13
.endm

; Seek in fd r0 to offset r1, whence r2
.macro file_seek
    SWI #0x14
.endm

; Get current position in fd r0
.macro file_tell
    SWI #0x15
.endm

; Get file size of fd r0
.macro file_size
    SWI #0x16
.endm

; =============================================================================
; MEMORY MANAGEMENT MACROS
; =============================================================================

; Allocate memory of size r0
; Returns address in r0 (0 on failure)
.macro malloc
    SWI #0x20
.endm

; Free memory at address r0
; Returns 0 on success, -1 on error
.macro free
    SWI #0x21
.endm

; Reallocate memory at r0 to new size r1
; Returns new address in r0 (0 on failure)
.macro realloc
    SWI #0x22
.endm

; =============================================================================
; SYSTEM INFORMATION MACROS
; =============================================================================

; Get current time in milliseconds
; Returns time in r0
.macro get_time
    SWI #0x30
.endm

; Get random 32-bit number
; Returns random value in r0
.macro get_random
    SWI #0x31
.endm

; Get program arguments
; Returns argc in r0, argv pointer in r1
.macro get_arguments
    SWI #0x32
.endm

; Get environment variables
; Returns envp pointer in r0
.macro get_environment
    SWI #0x33
.endm

; =============================================================================
; ERROR HANDLING MACROS
; =============================================================================

; Get last error code
; Returns error code in r0
.macro get_error
    SWI #0x40
.endm

; Set error code from r0
.macro set_error
    SWI #0x41
.endm

; Print error message for code in r0
.macro print_error
    SWI #0x42
.endm

; =============================================================================
; DEBUGGING MACROS
; =============================================================================

; Print debug message at address r0 to stderr
.macro debug_print
    SWI #0xF0
.endm

; Trigger breakpoint
.macro breakpoint
    SWI #0xF1
.endm

; Dump all registers to console
.macro dump_registers
    SWI #0xF2
.endm

; Dump memory at address r0, length r1
.macro dump_memory
    SWI #0xF3
.endm

; Assert condition in r0, message at address r1
; Halts program if condition is 0
.macro assert
    SWI #0xF4
.endm

; =============================================================================
; UTILITY MACROS
; =============================================================================

; Print a compile-time string constant
.macro print msg
    LDR r0, =.print_msg_\@
    write_string
    B .print_end_\@
.print_msg_\@:
    .asciz "\msg"
.align 4
.print_end_\@:
.endm

; Print a compile-time string constant followed by newline
.macro println msg
    print "\msg"
    write_newline
.endm

; Push a single register
.macro push reg
    STR \reg, [sp, #-4]!
.endm

; Pop a single register
.macro pop reg
    LDR \reg, [sp], #4
.endm

; Function prologue - save lr and frame pointer
.macro function_enter
    PUSH {fp, lr}
    MOV fp, sp
.endm

; Function epilogue - restore and return
.macro function_return
    POP {fp, pc}
.endm

; =============================================================================
; CONSTANTS
; =============================================================================

; File open modes
.equ FILE_MODE_READ,   0
.equ FILE_MODE_WRITE,  1
.equ FILE_MODE_APPEND, 2

; Standard file descriptors
.equ STDIN,  0
.equ STDOUT, 1
.equ STDERR, 2

; Common constants
.equ NULL,   0
.equ TRUE,   1
.equ FALSE,  0

; Error codes
.equ E_OK,      0
.equ E_GENERIC, 1
.equ E_NOMEM,   2
.equ E_NOFILE,  3
.equ E_INVALID, 4
