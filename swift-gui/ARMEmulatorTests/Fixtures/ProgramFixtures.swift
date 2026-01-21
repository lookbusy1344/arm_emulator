import Foundation

/// Test fixtures providing ARM assembly programs for testing
enum ProgramFixtures {
    /// Simple hello world program
    static let helloWorld = """
    .text
    .global _start
    _start:
        LDR R0, =message
        SWI #0x02        ; WRITE_STRING
        MOV R0, #0
        SWI #0x00        ; EXIT

    .data
    message: .asciz "Hello, World!\\n"
    """

    /// Program that exits with code 42
    static let exitCode42 = """
    MOV R0, #42
    SWI #0x00        ; EXIT
    """

    /// Simple loop that counts to 5
    static let simpleLoop = """
    .text
    .global _start
    _start:
        MOV R0, #0       ; counter
    loop:
        ADD R0, R0, #1
        CMP R0, #5
        BLT loop
        SWI #0x00        ; EXIT with R0=5
    """

    /// Program with a breakpoint opportunity (useful for debugging tests)
    static let withBreakpoint = """
    .text
    .global _start
    _start:
        MOV R0, #1       ; 0x8000 - good breakpoint spot
        MOV R1, #2       ; 0x8004 - another breakpoint spot
        ADD R2, R0, R1   ; 0x8008 - result in R2
        MOV R0, #0
        SWI #0x00        ; EXIT
    """

    /// Program that writes to memory (for watchpoint tests)
    static let memoryWrite = """
    .text
    .global _start
    _start:
        LDR R0, =data_area
        MOV R1, #42
        STR R1, [R0]     ; Write 42 to memory
        LDR R2, [R0]     ; Read it back
        MOV R0, #0
        SWI #0x00        ; EXIT

    .data
    data_area: .word 0
    """

    /// Program with syntax error
    static let syntaxError = """
    INVALID_INSTRUCTION R0, #42
    MOV R0, #0
    SWI #0x00
    """

    /// Program that runs for a while (for timeout/stop tests)
    static let longRunning = """
    .text
    .global _start
    _start:
        MOV R0, #0
    loop:
        ADD R0, R0, #1
        CMP R0, #1000000
        BLT loop
        SWI #0x00        ; EXIT
    """

    /// Fibonacci program (from examples)
    static let fibonacci = """
    .text
    .global _start

    _start:
        ; Print prompt
        LDR R0, =prompt
        SWI #0x02        ; WRITE_STRING

        ; Read count
        SWI #0x06        ; READ_INT
        MOV R3, R0       ; R3 = count

        ; Initialize
        MOV R0, #0       ; fib(0) = 0
        MOV R1, #1       ; fib(1) = 1
        MOV R2, #0       ; counter

    fib_loop:
        CMP R2, R3
        BGE done

        ; Print current number
        PUSH {R0-R3}
        SWI #0x07        ; WRITE_INT
        MOV R0, #' '
        SWI #0x01        ; WRITE_CHAR
        POP {R0-R3}

        ; Calculate next
        MOV R4, R0
        ADD R0, R0, R1
        MOV R1, R4

        ADD R2, R2, #1
        B fib_loop

    done:
        ; Print newline
        MOV R0, #10
        SWI #0x01        ; WRITE_CHAR

        MOV R0, #0
        SWI #0x00        ; EXIT

    .data
    prompt: .asciz "Enter count: "
    """

    /// Function call example (for step over/out tests)
    static let functionCall = """
    .text
    .global _start

    _start:
        MOV R0, #5
        MOV R1, #3
        BL add_numbers   ; Call function
        MOV R0, #0
        SWI #0x00        ; EXIT

    add_numbers:
        ADD R2, R0, R1
        MOV PC, LR       ; Return
    """

    /// Load program from example file
    /// - Parameter filename: Name of the example file (e.g., "fibonacci.s")
    /// - Returns: Program source code or nil if file not found
    static func loadExample(_ filename: String) -> String? {
        let examplesPath = "../examples/\(filename)"
        guard let url = URL(string: "file://\(FileManager.default.currentDirectoryPath)/\(examplesPath)") else {
            return nil
        }
        return try? String(contentsOf: url, encoding: .utf8)
    }
}
