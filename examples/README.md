# ARM Assembly Examples

This directory contains example ARM assembly programs that demonstrate various features of the ARM2 emulator.

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

## Algorithm Examples

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

### 4. factorial.s
**Factorial Calculator**
- Computes factorial using recursion
- Input validation (0-12 to prevent overflow)
- Demonstrates: recursive function calls, stack usage, error handling

**Example:**
```
Input: 5
Output: 5! = 120
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

### 6. binary_search.s
**Binary Search Algorithm**
- Searches for a value in a sorted array
- Efficient O(log n) search algorithm
- Demonstrates: binary search, function calls, array indexing

**Example:**
```
Searching for value 25 in sorted array [5, 10, 15, 20, 25, 30, 35, 40, 45, 50]
Value found at index 4
```

### 7. gcd.s
**Greatest Common Divisor**
- Calculates GCD using Euclidean algorithm
- Interactive input for two numbers
- Demonstrates: loops, modulo operation (via repeated subtraction), algorithm implementation

**Example:**
```
Input: 48, 18
Output: GCD of 48 and 18 is 6
```

## Data Structure Examples

### 8. arrays.s
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

### 9. linked_list.s
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

### 10. stack.s
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

### 11. strings.s
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

## Advanced Examples

### 12. functions.s
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

### 13. conditionals.s
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

### 14. loops.s
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

## Utility Examples

### 15. times_table.s
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

### 16. string_reverse.s
**String Reversal**
- Reads a string and prints it reversed
- In-place reversal algorithm
- Demonstrates: string operations, byte access, pointer manipulation

**Example:**
```
Input: "Hello"
Output: "olleH"
```

### 17. calculator.s
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

To run these examples with the ARM emulator:

```bash
# Build the emulator first
go build -o arm-emulator

# Run any example
./arm-emulator examples/hello.s
./arm-emulator examples/fibonacci.s
./arm-emulator examples/bubble_sort.s
# etc.
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

### Condition Codes
- `EQ` (Equal), `NE` (Not Equal)
- `GT` (Greater Than), `GE` (Greater or Equal)
- `LT` (Less Than), `LE` (Less or Equal)
- `AL` (Always), and many more

### System Calls
- **Console I/O**: WRITE_STRING, WRITE_INT, WRITE_CHAR, READ_INT, WRITE_NEWLINE
- **Memory**: ALLOCATE, FREE, REALLOCATE
- **Control**: EXIT

### Programming Concepts
- Function calls and returns
- Recursive functions
- Parameter passing
- Stack management
- Local variables
- Data structures (arrays, linked lists, stacks)
- Algorithms (sorting, searching, mathematical)
- Control flow (loops, conditionals, branches)
- String manipulation
- Dynamic memory allocation

## Difficulty Levels

- **Beginner**: hello.s, arithmetic.s, times_table.s
- **Intermediate**: fibonacci.s, factorial.s, bubble_sort.s, calculator.s, string_reverse.s, arrays.s, strings.s
- **Advanced**: binary_search.s, gcd.s, linked_list.s, stack.s, functions.s, conditionals.s, loops.s

## Learning Path

Recommended order for learning:

1. **hello.s** - Get familiar with basic program structure
2. **arithmetic.s** - Learn basic operations
3. **loops.s** - Understand iteration
4. **conditionals.s** - Learn branching and conditions
5. **functions.s** - Master function calls
6. **arrays.s** - Work with data structures
7. **strings.s** - Manipulate text data
8. **fibonacci.s** - Combine loops and arrays
9. **factorial.s** - Learn recursion
10. **bubble_sort.s** - Implement algorithms
11. **binary_search.s** - Efficient algorithms
12. **stack.s** - Advanced data structures
13. **linked_list.s** - Dynamic memory

## Notes

- All examples use the ARM2 instruction set
- Examples are designed for educational purposes
- Code includes extensive comments explaining each operation
- Programs demonstrate best practices for ARM assembly
- Error handling is included where appropriate
