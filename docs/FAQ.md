# Frequently Asked Questions (FAQ)

This FAQ covers common questions, errors, and troubleshooting tips for the ARM2 Emulator.

## Table of Contents

- [General Questions](#general-questions)
- [Installation & Setup](#installation--setup)
- [Assembly Programming](#assembly-programming)
- [Common Errors](#common-errors)
- [Memory & Addressing](#memory--addressing)
- [Functions & Stack](#functions--stack)
- [Debugger](#debugger)
- [Performance](#performance)
- [Platform-Specific Issues](#platform-specific-issues)

---

## General Questions

### Q: What is ARM2?

**A:** ARM2 is a 32-bit RISC processor from 1986 by Acorn Computers. It's the ancestor of modern ARM processors found in smartphones, tablets, and many embedded devices. This emulator faithfully reproduces the ARM2 instruction set for educational purposes.

### Q: Is this a full ARM emulator?

**A:** This emulator implements the complete ARM2 instruction set as found in the original 1986 processor, plus some useful ARMv3/ARMv3M extensions (long multiply, PSR transfer instructions). It does NOT support:
- Thumb mode (introduced in ARMv4T)
- ARMv7/ARMv8 instructions
- NEON/SIMD instructions
- Hardware floating point

### Q: Can I run Linux or other operating systems?

**A:** No. This is a user-mode emulator for ARM2 assembly programs. It doesn't emulate MMU, interrupts, or privileged modes needed for operating systems.

### Q: What's the difference between ARM2 and modern ARM?

**A:** Modern ARM (ARMv7/ARMv8/ARMv9) adds:
- Thumb instruction set (16-bit instructions)
- NEON SIMD instructions
- Hardware floating point
- Security extensions (TrustZone)
- 64-bit support (AArch64)
- Advanced exception handling

The core concepts (load-store architecture, conditional execution, registers) remain similar.

---

## Installation & Setup

### Q: What Go version do I need?

**A:** Go 1.25 or later. Check with:
```bash
go version
```

Install the latest version from https://go.dev/dl/

### Q: The build fails with "package not found" errors

**A:** Make sure you're in the correct directory and dependencies are up to date:
```bash
cd arm_emulator
go mod tidy
go build -o arm-emulator
```

### Q: Where is the configuration file?

**A:** Platform-specific locations:
- **macOS/Linux:** `~/.config/arm-emu/config.toml`
- **Windows:** `%APPDATA%\arm-emu\config.toml`

The emulator creates it automatically with defaults on first run.

### Q: How do I uninstall the emulator?

**A:** Delete the binary and configuration:
```bash
# macOS/Linux
rm /usr/local/bin/arm-emulator  # if installed globally
rm -rf ~/.config/arm-emu

# Windows
# Delete arm-emulator.exe and %APPDATA%\arm-emu
```

---

## Assembly Programming

### Q: Where does my program start?

**A:** The emulator searches for entry points in this order:
1. `_start`
2. `main`
3. `__start`
4. `start`
5. First instruction at origin (`.org` address)

**Recommendation:** Always use `_start` as your entry point.

### Q: What's the difference between `.org 0x8000` and `.org 0x0000`?

**A:**
- **`.org 0x8000`**: Traditional ARM load address, gives you 32KB of space below for stack/heap. Recommended for most programs.
- **`.org 0x0000`**: Start at address zero. Useful for specific scenarios but may require explicit `.ltorg` directives for large constants.

### Q: Why does `LDR R0, =0x12345678` work but `MOV R0, #0x12345678` doesn't?

**A:** ARM immediate values are limited to 8-bit values with rotation. The `LDR Rd, =constant` pseudo-instruction:
- Uses `MOV` or `MVN` if the constant can be encoded
- Otherwise creates a literal pool entry and loads from memory

For large constants, always use `LDR Rd, =constant`.

### Q: How do I print output?

**A:** Use syscalls:

```asm
; Print a string
LDR     R0, =msg
SWI     #0x02           ; WRITE_STRING

; Print a number
MOV     R0, #42
MOV     R1, #10         ; Base 10
SWI     #0x03           ; WRITE_INT

; Print newline
SWI     #0x07           ; WRITE_NEWLINE
```

### Q: How do I read input?

**A:** Use input syscalls:

```asm
; Read a character
SWI     #0x04           ; READ_CHAR
; Result in R0

; Read a string
LDR     R0, =buffer
MOV     R1, #100        ; Max length
SWI     #0x05           ; READ_STRING

; Read an integer
SWI     #0x06           ; READ_INT
; Result in R0
```

### Q: What's the difference between `.asciz` and `.ascii`?

**A:**
- **`.asciz`**: Null-terminated string (adds `\0` at end) - use with `SWI #0x02`
- **`.ascii`**: Raw string without null terminator

```asm
msg1:   .asciz  "Hello"    ; 6 bytes: H e l l o \0
msg2:   .ascii  "Hello"    ; 5 bytes: H e l l o
```

### Q: Can I use C-style comments?

**A:** Yes! Multiple comment styles are supported:

```asm
; Semicolon comments
// Double-slash comments
/* Multi-line
   comments */
MOV R0, #10     ; Inline comments
```

---

## Common Errors

### Q: "Error: entry point '_start' not found"

**A:** Your program is missing an entry point. Add a `_start` label:

```asm
        .org    0x8000

_start:
        ; Your code here
        MOV     R0, #0
        SWI     #0x00
```

### Q: "Error: undefined symbol 'label_name'"

**A:** You referenced a label that doesn't exist. Check for:
- Typos in label names (case-sensitive!)
- Missing label definitions
- Labels defined after use (forward references usually work, but check)

```asm
        B       my_label    ; Error if my_label not defined

my_label:
        MOV     R0, #0
```

### Q: "Error: immediate value out of range"

**A:** ARM immediate values must be 8-bit values with even rotation. Solutions:

```asm
; Problem
MOV     R0, #1000       ; Error: can't encode 1000

; Solution 1: Use LDR with literal pool
LDR     R0, =1000       ; OK: loads from memory

; Solution 2: Build from smaller values
MOV     R0, #1000       ; If encodable, assembler does this:
; becomes: MOV R0, #250, then LSL or similar

; Solution 3: Use multiple instructions
MOV     R0, #3
MOV     R1, #232
ADD     R0, R0, R1      ; R0 = 1000
```

### Q: "Error: Rd and Rm must be different in MUL"

**A:** ARM2 MUL restriction: destination and first operand must differ:

```asm
; Wrong
MUL     R0, R0, R1      ; Error: Rd == Rm

; Correct
MUL     R0, R1, R0      ; OK: Rd != Rm (but note operand order!)
; or
MOV     R2, R0
MUL     R0, R2, R1      ; OK: use temporary register
```

### Q: "Segmentation fault" or "memory access violation"

**A:** Common causes:
1. **Uninitialized register**: Forgot to initialize a pointer register
2. **Stack overflow**: Recursion too deep or missing stack setup
3. **Invalid pointer arithmetic**: Calculated wrong address
4. **Corrupted stack**: Forgot to preserve LR or corrupted SP

**Debug with:**
```bash
./arm-emulator --stack-trace --debug program.s
```

### Q: Program exits immediately without output

**A:** Check:
1. Entry point exists and has correct name (`_start`)
2. Not hitting `SWI #0x00` (EXIT) too early
3. Logic errors causing immediate branch to exit

**Debug with:**
```bash
./arm-emulator --debug program.s
(debug) break _start
(debug) run
(debug) step
```

### Q: "Error: literal pool out of range"

**A:** PC-relative addressing has ±4095 byte range. For `.org 0x0000` or very large programs:

```asm
.org 0x0000

main:
    LDR R0, =0x12345678
    ; ... lots of code ...
    B   next_section

    .ltorg              ; Force literal pool emission here

next_section:
    ; ... more code ...
```

---

## Memory & Addressing

### Q: What's the memory layout?

**A:**
```
0x00000000 ┌──────────────┐
           │ Low Memory   │
           │ (if .org 0)  │
0x00008000 ├──────────────┤
           │ Program Code │ (.org 0x8000 default)
           │ and Data     │
0x00010000 ├──────────────┤
           │ Heap         │
           │ (grows up ↑) │
           ├──────────────┤
           │ Free Space   │
           ├──────────────┤
           │ Stack        │
           │ (grows dn ↓) │
0x00050000 └──────────────┘ SP initialized here
```

### Q: How do I access array elements?

**A:** Calculate address = base + (index * element_size):

```asm
        ; Word array (4 bytes per element)
        LDR     R0, =array
        MOV     R1, #2              ; index
        LDR     R2, [R0, R1, LSL #2]  ; array[2]
        ; LSL #2 means multiply index by 4

        ; Byte array
        LDR     R0, =bytes
        MOV     R1, #5
        LDRB    R2, [R0, R1]        ; bytes[5]
```

### Q: What's the difference between `LDR` and `LDRB`?

**A:**
- **`LDR`**: Load 32-bit word (4 bytes)
- **`LDRB`**: Load byte (8 bits), zero-extended to 32 bits
- **`LDRH`**: Load halfword (16 bits), zero-extended

```asm
        ; Memory contains: 0x12 0x34 0x56 0x78 at address 0x8000
        LDR     R0, [R1]    ; R0 = 0x78563412 (little-endian word)
        LDRB    R0, [R1]    ; R0 = 0x00000012 (byte)
        LDRH    R0, [R1]    ; R0 = 0x00003412 (halfword)
```

### Q: How do I allocate dynamic memory?

**A:** Use the ALLOCATE syscall:

```asm
        ; Allocate 100 bytes
        MOV     R0, #100
        SWI     #0x20           ; ALLOCATE
        ; R0 now contains address

        MOV     R4, R0          ; Save address

        ; Use memory...

        ; Free when done
        MOV     R0, R4
        SWI     #0x21           ; FREE
```

### Q: Why is my data misaligned?

**A:** ARM requires proper alignment:
- Words (`.word`): must be 4-byte aligned
- Halfwords (`.half`): must be 2-byte aligned
- Bytes (`.byte`): no alignment requirement

**Use `.align`:**
```asm
        .asciz  "Hello"     ; 6 bytes (including null)
        .align  2           ; Align to next 4-byte boundary
data:   .word   42          ; OK: properly aligned
```

---

## Functions & Stack

### Q: How do I call a function?

**A:** Use `BL` (Branch with Link):

```asm
        MOV     R0, #5      ; First argument
        MOV     R1, #7      ; Second argument
        BL      my_function ; Call (LR = return address)
        ; R0 contains return value
```

### Q: What's the function calling convention?

**A:**
- **Arguments**: R0-R3 (additional args on stack)
- **Return value**: R0
- **Preserved**: R4-R11, SP, LR must be saved by callee
- **Scratch**: R0-R3, R12 can be modified

**Function template:**
```asm
my_function:
        STMFD   SP!, {R4-R7, LR}    ; Save registers

        ; Function body
        ; ...

        LDMFD   SP!, {R4-R7, PC}    ; Restore and return
```

### Q: How do I return from a function?

**A:** Two methods:

```asm
; Method 1: Direct return
        MOV     PC, LR

; Method 2: Pop PC from stack (if you pushed LR)
        LDMFD   SP!, {... , PC}
```

### Q: What's the difference between `STMFD` and `STMIA`?

**A:**
- **`STMFD SP!`**: Full Descending - standard stack push (decrement before store)
- **`STMIA Rn!`**: Increment After - for building structures upward

**Stack operations (most common):**
```asm
STMFD   SP!, {R0-R3, LR}    ; Push (Full Descending)
LDMFD   SP!, {R0-R3, PC}    ; Pop (Full Descending)
```

### Q: "Stack overflow" error - how do I fix it?

**A:** Causes:
1. **Infinite recursion**: Missing base case
2. **Too much stack allocation**: Large local arrays
3. **Forgot to pop**: Unbalanced push/pop

**Solutions:**
1. Check recursion base cases
2. Use heap allocation for large data
3. Verify every push has matching pop
4. Monitor with: `./arm-emulator --stack-trace program.s`

### Q: Can I pass more than 4 arguments?

**A:** Yes, push extras onto stack:

```asm
; Call function with 6 arguments
        MOV     R0, #1      ; Arg 1
        MOV     R1, #2      ; Arg 2
        MOV     R2, #3      ; Arg 3
        MOV     R3, #4      ; Arg 4
        LDR     R4, =5
        LDR     R5, =6
        STMFD   SP!, {R4-R5}  ; Push args 5-6
        BL      my_function
        ADD     SP, SP, #8    ; Clean up stack (2 words)

my_function:
        STMFD   SP!, {R4-R5, LR}
        ; Args 1-4 in R0-R3
        ; Arg 5 at [SP, #12]
        ; Arg 6 at [SP, #16]
        ; ...
        LDMFD   SP!, {R4-R5, PC}
```

---

## Debugger

### Q: How do I use the debugger?

**A:** Start with `--debug` or `--tui`:

```bash
# Command-line debugger
./arm-emulator --debug program.s

# TUI (visual) mode
./arm-emulator --tui program.s
```

**Essential commands:**
```
run               - Start program
step              - Execute one instruction
next              - Step over function calls
continue          - Run until breakpoint
break main        - Set breakpoint
print R0          - Show register
info registers    - Show all registers
```

### Q: How do I set a breakpoint?

**A:** Use `break` command with label or address:

```
(debug) break _start        # At label
(debug) break main          # At function
(debug) break 0x8000        # At address
(debug) break my_loop       # At any label
```

### Q: How do I examine memory?

**A:** Use `print` with memory syntax:

```
(debug) print [R0]          # Word at address in R0
(debug) print [0x8000]      # Word at specific address
(debug) print [R1 + 4]      # Word at R1 + 4
(debug) x/10 R0             # Examine 10 words starting at R0
```

### Q: TUI mode doesn't display correctly

**A:** Requirements:
- Terminal with ANSI color support
- Minimum terminal size: 80x24
- On Windows: Use Windows Terminal or PowerShell (not cmd.exe)

**Fix:**
```bash
# Check terminal size
echo $COLUMNS x $LINES

# Resize terminal if needed
# Use command-line debugger as alternative
./arm-emulator --debug program.s
```

### Q: How do I trace execution?

**A:** Multiple tracing options:

```bash
# Basic execution trace
./arm-emulator --trace program.s

# Save to file
./arm-emulator --trace --trace-file trace.txt program.s

# Memory access trace
./arm-emulator --mem-trace program.s

# Code coverage
./arm-emulator --coverage program.s

# Stack monitoring
./arm-emulator --stack-trace program.s

# Flag tracking
./arm-emulator --flag-trace program.s
```

---

## Performance

### Q: How fast is the emulator?

**A:** The emulator can execute millions of instructions per second on modern hardware, far faster than the original 6-10 MIPS ARM2 processor. Performance depends on:
- Host CPU speed
- Tracing/debugging enabled (overhead)
- Complexity of syscalls used

### Q: Do tracing modes slow down execution?

**A:** Yes, diagnostics add overhead:
- **No tracing**: Full speed
- **Basic trace**: ~10-20% slower
- **All diagnostics**: ~30-50% slower
- **TUI mode**: Additional rendering overhead

For maximum speed, run without debug/trace flags.

### Q: Can I profile my program?

**A:** Yes! Use statistics mode:

```bash
# Generate performance statistics
./arm-emulator --stats --stats-format html program.s

# View instruction frequency
./arm-emulator --stats --stats-format json program.s | jq '.instruction_frequency'
```

Statistics include:
- Instruction frequency
- Branch statistics
- Function call profiling
- Hot path analysis

### Q: How much memory can I use?

**A:** The emulator provides:
- **Program space**: 4GB address space (32-bit)
- **Default stack**: 1MB (configurable)
- **Heap**: Limited by available host memory

Practical limits depend on your host system.

---

## Platform-Specific Issues

### Q: macOS: "command not found: arm-emulator"

**A:** Add to PATH or use full path:

```bash
# Use full path
./arm-emulator program.s

# Or install globally
sudo mv arm-emulator /usr/local/bin/
arm-emulator program.s
```

### Q: Windows: "arm-emulator is not recognized"

**A:** Either:
1. Use full path: `.\arm-emulator.exe program.s`
2. Add directory to PATH environment variable

### Q: Linux: "permission denied"

**A:** Make executable:

```bash
chmod +x arm-emulator
./arm-emulator program.s
```

### Q: macOS: "cannot be opened because the developer cannot be verified"

**A:** Security workaround:

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine arm-emulator

# Or build from source
go build -o arm-emulator
```

### Q: File paths with spaces don't work

**A:** Use quotes:

```bash
# Wrong
./arm-emulator my program.s

# Correct
./arm-emulator "my program.s"
```

---

## Tips & Best Practices

### Q: How do I write better assembly code?

**A:**
1. **Comment extensively**: Explain the "why", not the "what"
2. **Use meaningful labels**: `calculate_total` not `label1`
3. **Consistent style**: Pick a style and stick to it
4. **Test incrementally**: Don't write everything at once
5. **Use the debugger**: Step through code to understand behavior
6. **Preserve registers**: Follow calling conventions
7. **Check edge cases**: Test with 0, -1, maximum values
8. **Initialize everything**: Don't assume register values

### Q: How do I debug a crash?

**A:**
1. Run with stack trace: `--stack-trace`
2. Use debugger to find crash location
3. Check register values before crash
4. Verify pointers are valid
5. Look for uninitialized registers
6. Check stack balance (push/pop pairs)

### Q: How do I learn ARM assembly?

**A:**
1. Start with [TUTORIAL.md](TUTORIAL.md)
2. Study example programs in `examples/` (49 programs covering all concepts)
3. Reference [INSTRUCTIONS.md](INSTRUCTIONS.md) for CPU instructions and syscalls
4. Learn [ASSEMBLER.md](ASSEMBLER.md) for directives (.text, .data, .word, .ltorg, etc.)
5. Check [REFERENCE.md](REFERENCE.md) for condition codes, addressing modes, and shifts
6. Read [assembly_reference.md](assembly_reference.md) for comprehensive syntax guide
7. Write small programs and gradually increase complexity
8. Use debugger (see [debugger_reference.md](debugger_reference.md)) to understand instruction behavior
9. Practice, practice, practice!

### Q: Where can I find example programs?

**A:** The `examples/` directory contains 49 programs:
- **hello.s**: Basic output
- **arithmetic.s**: Math operations
- **functions.s**: Function calls and conventions
- **arrays.s**: Array operations
- **fibonacci.s**: Recursive algorithms
- **bubble_sort.s**: Sorting
- And many more!

### Q: How do I report a bug?

**A:** Create an issue with:
1. Emulator version
2. Platform (OS and version)
3. Assembly code that reproduces the bug
4. Expected vs actual behavior
5. Error messages or output

---

## Still Need Help?

- **Documentation**: Check `docs/` directory for comprehensive guides
- **Tutorial**: Work through [TUTORIAL.md](TUTORIAL.md) for hands-on learning
- **Instruction Reference**: [INSTRUCTIONS.md](INSTRUCTIONS.md) - ARM2 CPU instructions and syscalls
- **Assembler Directives**: [ASSEMBLER.md](ASSEMBLER.md) - .text, .data, .word, .ltorg, etc.
- **Programming Reference**: [REFERENCE.md](REFERENCE.md) - Condition codes, addressing modes, shifts
- **Assembly Reference**: [assembly_reference.md](assembly_reference.md) for comprehensive syntax guide
- **Examples**: Browse `examples/` (49 programs) for reference code
- **Debugger**: [debugger_reference.md](debugger_reference.md) and [debugging_tutorial.md](debugging_tutorial.md)

**Remember**: The best way to learn is by doing. Write code, make mistakes, debug, and learn!
