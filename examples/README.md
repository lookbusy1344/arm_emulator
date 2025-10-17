# ARM Assembly Examples

This directory contains example ARM assembly programs that demonstrate various features of the ARM2 emulator.

**Total Examples:** 42 programs covering basic operations, algorithms, data structures, advanced features, and comprehensive stress tests.

## Quick Navigation

- [Basic Examples](#basic-examples) - Simple programs for beginners
- [Algorithm Examples](#algorithm-examples) - Classic algorithms and data processing
- [Data Structure Examples](#data-structure-examples) - Arrays, lists, stacks, hash tables
- [Advanced Examples](#advanced-examples) - Functions, conditionals, loops, recursion
- [Bit Manipulation Examples](#bit-manipulation-examples) - Bit operations and ciphers
- [System & I/O Examples](#system--io-examples) - File I/O, system calls, stress tests
- [Test Programs](#test-programs) - Assembler feature and edge case tests
- [Running Examples](#running-examples) - How to run these programs
- [Features Demonstrated](#features-demonstrated) - Complete feature reference

## Basic Examples

### 1. hello.s
**Hello World Program**
- Classic first program in any language
- Demonstrates: String output, program termination
- Simple introduction to ARM assembly and syscalls

**Example Output:**
```
Hello, World!
```

### 2. arithmetic.s
**Basic Arithmetic Operations**
- Demonstrates addition, subtraction, multiplication, and division
- Shows how to perform basic math in ARM assembly
- Division implemented using repeated subtraction (ARM2 has no hardware divide)

**Example Output:**
```
Addition: 15 + 7 = 22
Subtraction: 20 - 8 = 12
Multiplication: 6 * 7 = 42
Division: 35 / 5 = 7
```

### 3. times_table.s
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

### 4. celsius_to_fahrenheit.s
**Temperature Converter**
- Converts Celsius to Fahrenheit
- Formula: F = C * 9 / 5 + 32
- Demonstrates: user input, arithmetic, division algorithm

**Example:**
```
Input: Enter temperature in Celsius: 25
Output: Temperature in Fahrenheit: 77
```

### 5. calculator.s
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

## Algorithm Examples

### 6. fibonacci.s
**Fibonacci Sequence Generator**
- Generates the first N Fibonacci numbers
- Supports up to 20 numbers
- Demonstrates: loops, variable management, formatted output

**Example:**
```
Input: 10
Output: 0, 1, 1, 2, 3, 5, 8, 13, 21, 34
```

### 7. factorial.s
**Factorial Calculator**
- Computes factorial using recursion
- Input validation (0-12 to prevent overflow)
- Demonstrates: recursive function calls, stack usage, error handling

**Example:**
```
Input: 5
Output: 5! = 120
```

### 8. recursive_factorial.s
**Recursive Factorial Implementation**
- Alternative factorial implementation emphasizing recursion
- Clean recursive algorithm demonstration
- Demonstrates: proper stack frame management, base cases, recursive calls

### 9. recursive_fib.s
**Recursive Fibonacci Calculator**
- Classic recursive Fibonacci implementation
- Demonstrates recursive function calls
- Shows call tree depth and stack usage

### 10. bubble_sort.s
**Bubble Sort Algorithm**
- Sorts an array of integers in ascending order
- Interactive array input
- Demonstrates: nested loops, array access, memory operations, comparison

**Example:**
```
Input: [64, 34, 25, 12, 22]
Output: [12, 22, 25, 34, 64]
```

### 11. quicksort.s
**Quicksort Algorithm**
- Efficient O(n log n) sorting algorithm
- Partition-based divide-and-conquer approach
- Demonstrates: recursion, in-place sorting, advanced algorithm implementation

### 12. binary_search.s
**Binary Search Algorithm**
- Searches for a value in a sorted array
- Efficient O(log n) search algorithm
- Demonstrates: binary search, function calls, array indexing

**Example:**
```
Searching for value 25 in sorted array [5, 10, 15, 20, 25, 30, 35, 40, 45, 50]
Value found at index 4
```

### 13. gcd.s
**Greatest Common Divisor**
- Calculates GCD using Euclidean algorithm
- Interactive input for two numbers
- Demonstrates: loops, modulo operation (via repeated subtraction), algorithm implementation

**Example:**
```
Input: 48, 18
Output: GCD of 48 and 18 is 6
```

### 14. division.s
**Integer Division Demonstration**
- Software division using repeated subtraction
- Returns both quotient and remainder
- Demonstrates: ARM2 lacks hardware division, loop-based algorithms

**Example Output:**
```
100 / 7 = 14 remainder 2
1000 / 17 = 58 remainder 14
```

### 15. sieve_of_eratosthenes.s
**Prime Number Generator**
- Classic sieve algorithm for finding prime numbers
- Efficient prime generation up to a limit
- Demonstrates: array manipulation, nested loops, mathematical algorithms

## Data Structure Examples

### 16. arrays.s
**Array Operations**
- Array initialization, access, and traversal
- Finding minimum, maximum, and sum
- Demonstrates: array manipulation, iteration, aggregate operations

**Example Output:**
```
Array: 10, 25, 5, 42, 18, 33, 7, 50, 12, 3
Minimum value: 3
Maximum value: 50
Sum of all elements: 205
```

### 17. linked_list.s
**Linked List Implementation**
- Dynamic linked list with insertion and deletion
- Uses dynamic memory allocation (syscall 0x20/0x21)
- Demonstrates: pointers, dynamic memory, data structure operations

**Example Output:**
```
List: 50 -> 40 -> 30 -> 20 -> 10
Element count: 5
Deleting value 30...
List: 50 -> 40 -> 20 -> 10
```

### 18. stack.s
**Stack-Based Calculator**
- Stack implementation with push/pop operations
- Evaluates postfix (reverse Polish notation) expressions
- Demonstrates: LIFO data structure, expression evaluation

**Example:**
```
Evaluating: (5 + 3) * 2
Result: 16
Evaluating: (10 - 6) / 2
Result: 2
```

### 19. hash_table.s
**Hash Table Implementation**
- Hash table with linear probing for collision resolution
- Insertion and lookup operations
- Demonstrates: hash functions, collision handling, modulo operations

**Example Output:**
```
Inserted: 100 -> 1000
Inserted: 33 -> 330 (collision resolved)
Found: 42 -> 420
```

### 20. strings.s
**String Manipulation**
- String operations: length, copy, compare, concatenate
- Byte-level memory access
- Demonstrates: string functions, byte operations, pointer manipulation

**Example Output:**
```
String: "Hello" has length 5
Copied string: "World"
Comparing "Hello" with "World": first string is less
Concatenated: "ARM Assembly"
```

### 21. string_reverse.s
**String Reversal**
- Reads a string and prints it reversed
- In-place reversal algorithm
- Demonstrates: string operations, byte access, pointer manipulation

**Example:**
```
Input: "Hello"
Output: "olleH"
```

### 22. matrix_multiply.s
**Matrix Multiplication**
- Computes C = A × B where A is m×n and B is n×p
- 2D array indexing with row-major order
- Demonstrates: nested loops, multiply-accumulate operations, complex array indexing

**Example Output:**
```
Matrix A (3×4) × Matrix B (4×2) = Matrix C (3×2)
Result: [[50, 60], [114, 140], [178, 220]]
```

## Advanced Examples

### 23. functions.s
**Function Calling Conventions**
- Demonstrates proper function call syntax
- Parameter passing in registers
- Return values and stack preservation
- Nested function calls

**Features:**
- Simple functions with return values
- Functions with multiple parameters
- Nested function calls
- Pass-by-reference parameter modification

**Example Output:**
```
Example 1: Adding 15 + 7
Result: 22

Example 2: Sum of 10, 20, 30, 40
Result: 100

Example 3: Hypotenuse of 3 and 4
Result: 5

Example 4: Swapping 100 and 200
After swap: 200, 100
```

### 24. conditionals.s
**Conditional Execution**
- If/else statements
- Nested conditions
- Switch/case using jump tables
- Multiple comparison types

**Features:**
- Simple if/else
- If/else if/else chains
- Nested conditionals
- Jump table implementation for switch/case

**Example Output:**
```
Example 1: Simple if/else (15 > 10?)
15 is greater than 10

Example 2: Grade calculator (score = 75)
Grade: C

Example 3: Nested conditions (age=25, has_license=yes)
Can drive

Example 4: Switch/case (day 3): Wednesday
```

### 25. loops.s
**Loop Constructs**
- For loops (counted iteration)
- While loops (pre-test loops)
- Do-while loops (post-test loops)
- Nested loops
- Break and continue simulation

**Features:**
- Simple counting loops
- Accumulator loops (sum)
- Product loops (factorial)
- Nested loops (multiplication table)
- Search loops with early exit

**Example Output:**
```
Example 1: For loop (1 to 5): 1 2 3 4 5

Example 2: While loop (sum 1 to 10)
Sum = 55

Example 3: Do-while loop (5!)
5! = 120

Example 4: Nested loops (3x3 multiplication table):
1	2	3
2	4	6
3	6	9

Example 5: Break/continue (find first even number)
Found even number at index 5 with value 18
```

### 26. nested_calls.s
**Deep Function Call Nesting**
- Tests deep call stack functionality
- Multiple levels of function nesting
- Demonstrates: stack frame management, return address handling, stack depth

### 27. state_machine.s
**Finite State Machine**
- Email address validator using state machine pattern
- State transitions based on input characters
- Demonstrates: state-based logic, character processing, validation algorithms

### 28. task_scheduler.s
**Cooperative Task Scheduler**
- Simple round-robin task scheduler
- Context switching between tasks
- Demonstrates: cooperative multitasking, register state management, scheduling algorithms

## Bit Manipulation Examples

### 29. bit_operations.s
**Comprehensive Bit Manipulation**
- Bit counting (popcount)
- Bit reversal
- Power of 2 detection
- Find first set bit
- Rotate operations
- Bit field extraction

**Features Tested:**
- Bitwise AND, OR, EOR operations
- Left and right shifts
- Rotation operations
- Bit masking and extraction

**Example Output:**
```
Test 1: Count bits in 0xF0F0F0F0
  Bits set: 16
Test 2: Reverse byte 0b10110010
  Reversed: 0b01001101
Test 3: Is 64 a power of 2? PASS
```

### 30. xor_cipher.s
**XOR Encryption/Decryption**
- Simple XOR cipher implementation
- Demonstrates encryption and decryption
- Shows: bitwise XOR operations, byte manipulation, symmetric encryption

## System & I/O Examples

### 31. addressing_modes.s
**ARM Addressing Modes Demonstration**
- Comprehensive demonstration of ARM2 addressing modes
- Tests immediate offset, pre-indexed with writeback, post-indexed
- Tests register offset and scaled register offset
- Demonstrates byte access with post-indexed addressing
- Self-testing program that validates each mode

**Example Output:**
```
Testing ARM Addressing Modes...
All addressing mode tests passed!
```

### 32. adr_demo.s
**ADR Pseudo-Instruction Demo**
- Position-independent address loading
- PC-relative addressing
- Loading addresses of labels and data
- Demonstrates: ADR instruction, position-independent code

### 33. memory_stress.s
**Memory Operations Stress Test**
- Sequential access patterns
- Strided access patterns
- Byte-level operations
- Pre/post indexed addressing
- LDM/STM operations

**Example Output:**
```
Test 1: Sequential access... PASS
Test 2: Strided access... PASS
Test 3: Byte operations... PASS
Test 4: Indexed addressing... PASS
Test 5: LDM/STM operations... PASS
```

### 34. multi_precision_arith.s
**Multi-Precision Arithmetic**
- 128-bit arithmetic operations
- ADC/SBC carry chain operations
- Flag preservation and testing
- Demonstrates: multi-word arithmetic, carry propagation, shift operations

**Example Output:**
```
Test 1 PASSED: A + B
Test 2 PASSED: (A+B) - B = A
Test 3 PASSED: (A+B) + C
Test 4 PASSED: Shift
ALL TESTS PASSED
```

### 35. file_io.s
**File I/O Operations**
- File open, read, write, close syscalls
- Seek and tell operations
- Round-trip write/read verification
- Demonstrates: file I/O syscalls, error handling, data verification

**Example Output:**
```
[file_io] File I/O round-trip test starting
64
64
[file_io] PASS
```

## Test Programs

These programs test specific assembler features and edge cases:

### 36. test_ltorg.s
**Literal Pool Management**
- Tests `.ltorg` directive for placing literal pools
- Demonstrates: literal pool placement, LDR pseudo-instruction, constant loading

### 37. test_org_0_with_ltorg.s
**Origin 0x0000 with Literal Pool**
- Tests programs starting at address 0x0000
- Combined with literal pool management
- Edge case testing for address calculations

### 38. nop_test.s
**NOP Pseudo-Instruction**
- Tests NOP (No Operation) instruction
- Demonstrates: pseudo-instruction expansion

### 39. const_expressions.s
**Constant Expression Evaluation**
- Tests constant arithmetic in assembly time
- Label arithmetic (label + offset, label - offset)
- Non-power-of-2 constants
- Demonstrates: assembler constant evaluation, compile-time arithmetic

### 40. test_const_expr.s
**Constant Expression Tests**
- Additional constant expression test cases
- Edge cases for constant evaluation

### 41. test_expr.s
**Expression Evaluation Tests**
- Tests expression parser in assembler
- Complex constant expressions

### 42. standalone_labels.s
**Standalone Label Syntax**
- Tests labels on their own lines (not followed by instructions)
- Demonstrates: label definition syntax variations

## Running Examples

To run these examples with the ARM emulator:

```bash
# Build the emulator first
go build -o arm-emulator

# Run any example
./arm-emulator examples/hello.s
./arm-emulator examples/fibonacci.s
./arm-emulator examples/bubble_sort.s
# etc.

# Run with diagnostic modes
./arm-emulator --trace examples/loops.s
./arm-emulator --coverage examples/fibonacci.s
./arm-emulator --stack-trace examples/factorial.s
./arm-emulator --flag-trace examples/conditionals.s
```

## Features Demonstrated

These examples collectively demonstrate:

### Instructions
- **Data Processing**: MOV, MVN, ADD, ADC, SUB, SBC, RSB, RSC, AND, ORR, EOR, BIC
- **Comparison**: CMP, CMN, TST, TEQ
- **Memory**: LDR, STR, LDRB, STRB, LDRH, STRH, LDM, STM
- **Branch**: B, BL, BX with all condition codes
- **Multiply**: MUL, MLA
- **Stack**: PUSH (STMFD), POP (LDMFD)
- **Pseudo-instructions**: ADR, NOP, LDR =constant

### Addressing Modes
- Immediate: `#value`
- Register: `Rn`
- Register with shift: `Rm, LSL #shift`
- Memory offset: `[Rn, #offset]`
- Pre-indexed: `[Rn, #offset]!`
- Post-indexed: `[Rn], #offset`
- Register offset: `[Rn, Rm]`
- Scaled offset: `[Rn, Rm, LSL #shift]`

### Directives
- `.org` - Set origin address
- `.word` - Define 32-bit word
- `.byte` - Define byte
- `.asciz` - Define null-terminated string
- `.space` - Reserve space
- `.align` - Align to boundary
- `.equ` - Define constant
- `.ltorg` - Place literal pool
- `.text` - Code section
- `.data` - Data section
- `.global` - Global symbol declaration

### Condition Codes
- `EQ` (Equal), `NE` (Not Equal)
- `GT` (Greater Than), `GE` (Greater or Equal)
- `LT` (Less Than), `LE` (Less or Equal)
- `CS`/`HS` (Carry Set/Higher or Same)
- `CC`/`LO` (Carry Clear/Lower)
- `MI` (Minus/Negative), `PL` (Plus/Positive)
- `VS` (Overflow Set), `VC` (Overflow Clear)
- `HI` (Higher), `LS` (Lower or Same)
- `AL` (Always)

### System Calls
- **Console I/O**:
  - `0x02` WRITE_STRING - Print null-terminated string
  - `0x03` WRITE_INT - Print integer (with base selection)
  - `0x04` WRITE_CHAR - Print single character
  - `0x06` READ_INT - Read integer from stdin
  - `0x07` WRITE_NEWLINE - Print newline
- **Memory Management**:
  - `0x20` ALLOCATE - Allocate memory
  - `0x21` FREE - Free memory
  - `0x22` REALLOCATE - Reallocate memory
- **File I/O**:
  - `0x10` OPEN - Open file
  - `0x11` CLOSE - Close file
  - `0x12` READ - Read from file
  - `0x13` WRITE - Write to file
  - `0x14` SEEK - Seek in file
  - `0x15` TELL - Get file position
- **Control**:
  - `0x00` EXIT - Exit program

### Programming Concepts
- Function calls and returns
- Recursive functions
- Parameter passing (registers and stack)
- Stack management and frame pointers
- Local variables
- Data structures (arrays, linked lists, stacks, hash tables)
- Algorithms (sorting, searching, mathematical)
- Control flow (loops, conditionals, branches)
- String manipulation
- Dynamic memory allocation
- Bit manipulation
- State machines
- Multi-precision arithmetic
- File I/O and error handling

## Difficulty Levels

- **Beginner** (Start here):
  - hello.s, arithmetic.s, times_table.s, celsius_to_fahrenheit.s

- **Intermediate** (Core concepts):
  - fibonacci.s, factorial.s, bubble_sort.s, calculator.s, string_reverse.s,
  - arrays.s, strings.s, conditionals.s, loops.s, functions.s, division.s

- **Advanced** (Complex algorithms and data structures):
  - binary_search.s, gcd.s, linked_list.s, stack.s, hash_table.s,
  - quicksort.s, matrix_multiply.s, bit_operations.s, sieve_of_eratosthenes.s,
  - recursive_factorial.s, recursive_fib.s, state_machine.s, task_scheduler.s

- **Expert** (System-level and stress tests):
  - memory_stress.s, multi_precision_arith.s, file_io.s, nested_calls.s,
  - addressing_modes.s, xor_cipher.s

- **Test/Edge Cases** (Assembler feature testing):
  - test_ltorg.s, test_org_0_with_ltorg.s, nop_test.s, const_expressions.s,
  - test_const_expr.s, test_expr.s, standalone_labels.s, adr_demo.s

## Learning Path

Recommended order for learning:

1. **hello.s** - Get familiar with basic program structure
2. **arithmetic.s** - Learn basic operations
3. **conditionals.s** - Learn branching and conditions
4. **loops.s** - Understand iteration
5. **functions.s** - Master function calls
6. **arrays.s** - Work with data structures
7. **strings.s** - Manipulate text data
8. **fibonacci.s** - Combine loops and variables
9. **factorial.s** - Learn recursion basics
10. **recursive_factorial.s** - Deep dive into recursion
11. **bubble_sort.s** - Implement algorithms
12. **binary_search.s** - Efficient algorithms
13. **stack.s** - Advanced data structures
14. **linked_list.s** - Dynamic memory
15. **hash_table.s** - Complex data structures
16. **bit_operations.s** - Low-level operations
17. **addressing_modes.s** - Master ARM addressing
18. **memory_stress.s** - Advanced memory operations

## Testing

All example programs have integration tests that verify their output. Run the tests with:

```bash
go clean -testcache
go test ./tests/integration/...
```

32 programs have expected output files for automated testing.

## Notes

- All examples use the ARM2 instruction set
- Examples are designed for educational purposes
- Code includes extensive comments explaining each operation
- Programs demonstrate best practices for ARM assembly
- Error handling is included where appropriate
- The emulator uses ARM2 syscall conventions (not Linux-style syscalls)
- All file operations use syscall numbers 0x10-0x15
- Memory allocation uses syscalls 0x20-0x22

## Additional Resources

For more information about the emulator and its features, see:
- `../README.md` - Main project documentation
- `../docs/` - Comprehensive documentation
- `../CLAUDE.md` - Development guidelines and project structure
