# ARM Assembly Examples

This directory contains example ARM assembly programs that demonstrate various features of the ARM emulator.

## Examples

### 1. times_table.s
**Multiplication Table Generator**
- Reads a number from input (1-12)
- Displays the complete multiplication table for that number
- Demonstrates: loops, multiplication, I/O operations

**Example:**
```
Input: 5
Output:
5 x 1 = 5
5 x 2 = 10
...
5 x 12 = 60
```

### 2. factorial.s
**Factorial Calculator**
- Computes factorial using recursion
- Input validation (0-12 to prevent overflow)
- Demonstrates: recursive function calls, stack usage, error handling

**Example:**
```
Input: 5
Output: 5! = 120
```

### 3. fibonacci.s
**Fibonacci Sequence Generator**
- Generates the first N Fibonacci numbers
- Supports up to 20 numbers
- Demonstrates: loops, variable management, formatted output

**Example:**
```
Input: 10
Output: 0, 1, 1, 2, 3, 5, 8, 13, 21, 34
```

### 4. string_reverse.s
**String Reversal**
- Reads a string and prints it reversed
- In-place reversal algorithm
- Demonstrates: string operations, byte access, pointer manipulation

**Example:**
```
Input: "Hello"
Output: "olleH"
```

### 5. bubble_sort.s
**Bubble Sort Algorithm**
- Sorts an array of integers in ascending order
- Interactive array input
- Demonstrates: nested loops, array access, memory operations, comparison

**Example:**
```
Input: [64, 34, 25, 12, 22]
Output: [12, 22, 25, 34, 64]
```

### 6. calculator.s
**Simple Calculator**
- Interactive calculator with basic operations (+, -, *, /)
- Menu-driven interface
- Division includes remainder
- Demonstrates: branching, arithmetic operations, user interaction

**Example:**
```
Input: 15 + 7
Output: 15 + 7 = 22
```

## Running Examples

Once the ARM emulator is complete, you can run these examples with:

```bash
./arm-emulator examples/times_table.s
./arm-emulator examples/factorial.s
# etc.
```

## Features Tested

These examples collectively test:
- **Arithmetic**: ADD, SUB, MUL, DIV operations
- **Memory**: LDR, STR, LDRB, STRB instructions
- **Control Flow**: Conditional branches, loops, function calls
- **Stack**: PUSH, POP operations
- **Comparison**: CMP instruction and condition flags
- **I/O**: System calls for reading/writing integers, strings, characters
- **Addressing Modes**: Immediate, register, indexed, shifted
- **Directives**: .org, .asciz, .space, .word, .align
- **Recursion**: Stack management for nested calls
- **String Operations**: Byte-level manipulation
- **Array Operations**: Indexed access, sorting algorithms
