# ARM Emulator Project

This is an ARM emulator written in Go that implements a subset of the ARM2 instruction set.

## Build Command

```bash
go build -o arm-emulator
```

## Format Command

```bash
go fmt ./...
```

## Lint Command

```bash
golangci-lint run ./...
```

## Test Command

```bash
go clean -testcache
go test ./...
```

**IMPORTANT:** All tests should be centralized in the `./tests` directory structure:
- `tests/unit/` - Unit tests for individual packages
- `tests/integration/` - Integration tests for complete workflows

Do NOT create test files in package directories (e.g., `api/api_test.go`). Instead, create them in the appropriate `./tests` subdirectory.

**Exceptions (due to Go language limitations):**
- `gui/app_test.go` - Tests for `package main` (cannot be imported from other directories)
- `debugger/tui_internal_test.go` - White-box tests requiring access to unexported methods

## Race Detection

Run tests with the race detector to check for data races (especially important for TUI/GUI code):

```bash
go test -race ./...
```

Run periodically, especially after modifying concurrent code (TUI, service layer, goroutines).

## API Integration Tests

API-level integration tests validate the complete HTTP REST API + WebSocket stack:

```bash
# Run all API integration tests
go test ./tests/integration -run TestAPIExamplePrograms -v

# Run specific API test
go test ./tests/integration -run TestAPIExamplePrograms/Fibonacci_API -v

# Run with race detector
go test -race ./tests/integration -run TestAPIExamplePrograms
```

**Test Coverage:**
- All 49 example programs tested via REST API
- WebSocket state monitoring for execution tracking
- Hybrid stdin strategy (batch + interactive)
- Reuses existing `expected_outputs/*.txt` files

**Note:** API tests start a real HTTP server for WebSocket support (cannot use httptest.ResponseRecorder).

## Update dependencies

```bash
go get -u ./...
go mod tidy
go mod verify
```

## Run Command

```bash
./arm-emulator program.s
```

### Filesystem Security

**IMPORTANT:** The emulator restricts file operations to a specified directory for security.

```bash
# Restrict to current directory (default)
./arm-emulator program.s

# Restrict to specific directory
./arm-emulator -fsroot /path/to/sandbox program.s

# Restrict to examples directory
./arm-emulator -fsroot ./examples program.s
```

By default, guest programs can only access files within the current working directory. Attempts to escape using `..` or symlinks will return an error code (`0xFFFFFFFF`) to the guest program and log a security warning to stderr.

When running tests, the filesystem root is automatically set appropriately for each test case.

## GUI Commands (Wails)

**IMPORTANT:** Always use the `-nocolour` flag with `wails build` and `wails dev` to prevent ANSI escape codes in output.

### Build GUI

```bash
wails build -nocolour
```

### Run GUI in Development Mode

```bash
wails dev -nocolour
```

### Run GUI with a File Pre-loaded

To launch the GUI with a specific assembly file already loaded:

```bash
cd gui
wails dev -nocolour -appargs "../examples/stack.s"
```

The window title will show "ARM Emulator - filename.s" when a file is loaded.

### Check Wails Environment

```bash
wails doctor
```

### E2E Testing

**IMPORTANT:** E2E tests require the Wails dev server to be running first. Tests will hang indefinitely if the backend is not available.

```bash
# Terminal 1: Start Wails dev server
cd gui
wails dev -nocolour

# Terminal 2: Run E2E tests
cd gui/frontend
npm run test:e2e                    # Run all tests
npm run test:e2e -- --project=chromium  # Run chromium only
npm run test:e2e:headed             # Run with visible browser
```

The Wails backend must be running on http://localhost:34115 before any E2E tests can execute.

## Swift Native macOS App Commands

**Note:** The Swift app automatically manages the Go HTTP API backend lifecycle - it finds, starts, and monitors the backend process. No manual backend startup required.

### Prerequisites

Install required tools via Homebrew:

```bash
brew install xcodegen swiftlint swiftformat xcbeautify
```

### Generate Xcode Project

The Xcode project is generated from `project.yml` using XcodeGen:

```bash
cd swift-gui
xcodegen generate
```

**IMPORTANT:** You must regenerate the Xcode project whenever you modify `project.yml` (e.g., adding files, changing settings, adding dependencies).

### Open in Xcode

```bash
# From project root
open swift-gui/ARMEmulator.xcodeproj

# Or from swift-gui directory
cd swift-gui
open ARMEmulator.xcodeproj
```

The project can be fully developed in Xcode with all standard features (visual debugging, Interface Builder for SwiftUI previews, breakpoints, etc.).

### Build Swift App (CLI)

```bash
cd swift-gui

# Debug build
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator build | xcbeautify

# Release build
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator -configuration Release build | xcbeautify

# Clean build
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator clean build | xcbeautify
```

Built app location: `~/Library/Developer/Xcode/DerivedData/ARMEmulator-*/Build/Products/Debug/ARMEmulator.app`

### Run Swift App

```bash
# Find and open the built app (backend starts automatically)
find ~/Library/Developer/Xcode/DerivedData -name "ARMEmulator.app" -type d -exec open {} \; -quit

# Or run from Xcode (Cmd+R)
cd swift-gui
open ARMEmulator.xcodeproj
# Then press Cmd+R in Xcode
```

The app automatically finds and starts the Go backend binary from the project root. The backend lifecycle is fully managed by the Swift app - it starts on launch and shuts down when the app quits.

### Test Swift App

```bash
cd swift-gui

# Run all tests
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify

# Run tests with coverage
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' -enableCodeCoverage YES | xcbeautify
```

### Code Quality (Swift)

```bash
cd swift-gui

# Format Swift code
swiftformat .

# Lint Swift code
swiftlint

# Auto-fix linting issues
swiftlint --fix
```

**IMPORTANT:** SwiftLint and SwiftFormat configurations are in `.swiftlint.yml` and `.swiftformat` respectively. Current configuration enforces 0 violations before commit.

### Swift App Architecture

The Swift app follows MVVM architecture and connects to the Go backend via:
- **HTTP REST API** (`APIClient.swift`) - For commands (run, step, reset, load program)
- **WebSocket** (`WebSocketClient.swift`) - For real-time updates (registers, console output, execution state)

See `SWIFT_GUI_PLANNING.md` for detailed architecture documentation and `docs/SWIFT_CLI_AUTOMATION.md` for comprehensive CLI development guide.

#### Debug Logging Best Practice

**IMPORTANT:** Use the `DebugLog` utility (in `swift-gui/ARMEmulator/Utilities/DebugLog.swift`) for all diagnostic logging. This provides:

- **Conditional compilation**: Logs are completely removed in RELEASE builds (no runtime overhead)
- **Runtime toggle**: Set `DebugLog.enabled = false` to silence logs without rebuilding
- **Categorized output**: Different log levels (log, success, error, warning, network, ui) with emoji icons

Example usage:
```swift
DebugLog.log("Loading program", category: "ViewModel")
DebugLog.network("POST /api/v1/session/\(sessionID)/run")
DebugLog.success("Program loaded successfully", category: "ViewModel")
DebugLog.error("Failed to connect: \(error)", category: "ViewModel")
DebugLog.ui("Run button clicked")
```

**Benefits:**
- Debug logs appear in Xcode console during development
- Zero overhead in production builds (code is stripped out)
- Easy to toggle on/off for specific debugging sessions
- Consistent formatting across the codebase

### Common Pitfalls

1. **"No such module 'SwiftUI'" error**: Ensure Xcode Command Line Tools are installed: `xcode-select --install`
2. **Xcode project out of sync**: Run `xcodegen generate` after modifying `project.yml`
3. **App can't find backend binary**: Ensure `arm-emulator` is built in the project root: `go build -o arm-emulator`
4. **SwiftLint/SwiftFormat not found**: Install via Homebrew: `brew install swiftlint swiftformat`
5. **Code signing errors**: The project uses automatic code signing (ad-hoc for development)
6. **DerivedData issues**: Clean with `rm -rf ~/Library/Developer/Xcode/DerivedData` if builds behave strangely
7. **Git tracking Xcode user files**: The `swift-gui/.gitignore` excludes user-specific files like `*.xcuserstate`, `xcuserdata/`, and build artifacts. These should never be committed.
8. **SwiftUI/AppKit rendering issues**: When NSViewRepresentable views don't render content (blank view despite data being present), use divide-and-conquer debugging: comment out complex custom components (ruler views, overlays, custom drawing) to isolate the issue. Get the basic AppKit view working first, then add complexity back incrementally. Example: NSTextView not displaying text may be caused by NSRulerView interference.

## Project Structure

- `main.go` - Entry point and CLI interface
- `vm/` - Virtual machine implementation (CPU, memory, execution, syscalls, tracing, statistics)
- `parser/` - Assembly parser with preprocessor and macros
- `instructions/` - Instruction implementations (data processing, memory, branch, multiply)
- `encoder/` - Machine code encoder/decoder for binary ARM instructions
- `debugger/` - Debugging utilities with TUI (breakpoints, watchpoints, expression evaluation)
- `config/` - Cross-platform configuration management
- `tools/` - Development tools (linter, formatter, cross-reference generator)
- `api/` - HTTP REST API backend for GUI frontends (runs on port 8080)
- `service/` - Service layer for API/GUI integration and emulator state management
- `gui/` - Wails cross-platform GUI (Go + Svelte web frontend)
- `swift-gui/` - Swift native macOS app (SwiftUI + MVVM, connects to API backend)
- `tests/` - Test files (1,024 tests, 100% pass rate)
  - `tests/unit/` - Unit tests for all packages
  - `tests/integration/` - Integration tests for complete programs
- `examples/` - Example ARM assembly programs (49 programs, all fully functional including 3 interactive)
- `docs/` - User and developer documentation

## SWI Syscall Reference

For the complete syscall reference, see [docs/INSTRUCTIONS.md](docs/INSTRUCTIONS.md#system-instructions).

### Quick Reference

The emulator implements traditional ARM2 syscall convention: `SWI #immediate_value` where the syscall number is encoded directly in the instruction. Arguments and return values use registers R0-R2.

### Console I/O (0x00-0x07)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x00 | EXIT | Exit program | R0: exit code | - |
| 0x01 | WRITE_CHAR | Write character to stdout | R0: character | - |
| 0x02 | WRITE_STRING | Write null-terminated string | R0: string address | - |
| 0x03 | WRITE_INT | Write integer in specified base | R0: value, R1: base (2/8/10/16, default 10) | - |
| 0x04 | READ_CHAR | Read character from stdin (skips whitespace) | - | R0: character or 0xFFFFFFFF on error |
| 0x05 | READ_STRING | Read string from stdin (until newline) | R0: buffer address, R1: max length (default 256) | R0: bytes written or 0xFFFFFFFF on error |
| 0x06 | READ_INT | Read integer from stdin | - | R0: integer value or 0 on error |
| 0x07 | WRITE_NEWLINE | Write newline to stdout | - | - |

### File Operations (0x10-0x16)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x10 | OPEN | Open file | R0: filename address, R1: mode (0=read, 1=write, 2=append) | R0: file descriptor or 0xFFFFFFFF on error |
| 0x11 | CLOSE | Close file | R0: file descriptor | R0: 0 on success, 0xFFFFFFFF on error |
| 0x12 | READ | Read from file | R0: fd, R1: buffer address, R2: length | R0: bytes read or 0xFFFFFFFF on error |
| 0x13 | WRITE | Write to file | R0: fd, R1: buffer address, R2: length | R0: bytes written or 0xFFFFFFFF on error |
| 0x14 | SEEK | Seek in file | R0: fd, R1: offset, R2: whence (0=start, 1=current, 2=end) | R0: new position or 0xFFFFFFFF on error |
| 0x15 | TELL | Get current file position | R0: file descriptor | R0: position or 0xFFFFFFFF on error |
| 0x16 | FILE_SIZE | Get file size | R0: file descriptor | R0: size or 0xFFFFFFFF on error |

### Memory Operations (0x20-0x22)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x20 | ALLOCATE | Allocate memory from heap | R0: size in bytes | R0: address or 0 (NULL) on failure |
| 0x21 | FREE | Free allocated memory | R0: address | R0: 0 on success, 0xFFFFFFFF on error |
| 0x22 | REALLOCATE | Resize memory allocation | R0: old address, R1: new size | R0: new address or 0 (NULL) on failure |

### System Information (0x30-0x33)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x30 | GET_TIME | Get time in milliseconds since Unix epoch | - | R0: timestamp (lower 32 bits) |
| 0x31 | GET_RANDOM | Get random 32-bit number | - | R0: random value |
| 0x32 | GET_ARGUMENTS | Get program arguments | - | R0: argc, R1: argv pointer (0 in current impl) |
| 0x33 | GET_ENVIRONMENT | Get environment variables | - | R0: envp pointer (0 in current impl) |

### Error Handling (0x40-0x42)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0x40 | GET_ERROR | Get last error code | - | R0: error code (0 in current impl) |
| 0x41 | SET_ERROR | Set error code | R0: error code | - |
| 0x42 | PRINT_ERROR | Print error message to stderr | R0: error code | - |

### Debugging Support (0xF0-0xF4)

| Code | Name | Description | Arguments | Return |
|------|------|-------------|-----------|--------|
| 0xF0 | DEBUG_PRINT | Print debug message to stderr | R0: string address | - |
| 0xF1 | BREAKPOINT | Trigger debugger breakpoint | - | - |
| 0xF2 | DUMP_REGISTERS | Print all registers to stdout | - | - |
| 0xF3 | DUMP_MEMORY | Dump memory region as hex dump | R0: address, R1: length (max 1KB) | - |
| 0xF4 | ASSERT | Assert condition is true | R0: condition (0=fail), R1: message address | Halts if condition is 0 |

**Note:** CPSR flags (N, Z, C, V) are preserved across all syscalls to prevent unintended side effects on conditional logic.

## Development Guidelines

**IMPORTANT:** After implementing each phase of development, update `PROGRESS.md` to reflect the completed work, including:
- Mark the phase as completed
- Document any implementation details or deviations from the original plan
- Update the status of related tasks
- Note outstanding work and issues in `TODO.md`

**IMPORTANT:** Always run `go fmt ./...`, `golangci-lint run ./...`, and `go build -o arm-emulator && go clean -testcache && go test ./...` after making changes and **BEFORE committing** to ensure code quality and correctness. Linting must pass with 0 issues before any commit.

**IMPORTANT:** Focus on Test-Driven Development (TDD). Write tests before implementing features. Ensure all new code is covered by tests. Use unit tests for individual components and integration tests for end-to-end functionality.

**IMPORTANT:** Do not delete tests without explicit instructions. Do not simplify tests because they fail. If you think a test is malfunctioning, think about it carefully and ask me before making any changes to the tests.

**IMPORTANT:** Anything that cannot be implemented should be noted in `TODO.md` with details so work can result later. TODO.md should not contain completed work, that should go in PROGRESS.md.

**IMPORTANT:** Do not modify example programs just to make them work without explicit permission, unless they are actually broken. Instead, fix the emulator to run the programs properly. Example programs are test cases that demonstrate expected behavior.

**IMPORTANT:** To ensure tests are up to date, recompile and clear the test cache before running tests `go build -o arm-emulator && go clean -testcache && go test ./...`

**IMPORTANT:** This emulator implements the classic ARM2 architecture. Do NOT implement Linux-style syscalls (using `SVC #0` with syscall number in R7 register). The emulator uses only traditional ARM2 syscall convention: `SWI #immediate_value` where the syscall number is encoded directly in the instruction. R7 is just a general-purpose register with no special meaning for syscalls.

**IMPORTANT:** All tests belong in the `tests/` directory structure, not in the main package directories. TUI tests use `tcell.SimulationScreen` to avoid terminal initialization issues. The `debugger.NewTUIWithScreen()` function accepts an optional screen parameter for testing while production code uses `debugger.NewTUI()` with the default screen.

**IMPORTANT:** Avoid embedding magic numbers directly in the code. Use named constants or enums for clarity and maintainability.

**IMPORTANT:** When doing code reviews, look at it with fresh eyes. Assume the engineer implemented it suspiciously quickly and is not to be trusted.

## Additional Features

### Performance Analysis

Run programs with tracing and statistics:

```bash
# Execution tracing
./arm-emulator --trace --trace-file trace.txt program.s

# Memory tracing
./arm-emulator --mem-trace --mem-trace-file mem_trace.txt program.s

# Performance statistics
./arm-emulator --stats --stats-file stats.html --stats-format html program.s
```

### Diagnostic Modes

Run programs with advanced diagnostic tracking:

```bash
# Code coverage - track which instructions were executed
./arm-emulator --coverage --coverage-format text program.s

# Stack trace - monitor stack operations and detect overflow/underflow
./arm-emulator --stack-trace --stack-trace-format text program.s

# Flag trace - track CPSR flag changes for debugging conditional logic
./arm-emulator --flag-trace --flag-trace-format text program.s

# Register access pattern analysis - track register usage patterns
./arm-emulator --register-trace --register-trace-format text program.s

# Combine multiple diagnostic modes
./arm-emulator --coverage --stack-trace --flag-trace --register-trace --verbose program.s
```

Features:
- **Code Coverage**: Tracks executed vs unexecuted instructions, reports coverage percentage
- **Stack Trace**: Monitors all stack operations (push/pop/SP modifications), detects overflow/underflow
- **Flag Trace**: Records CPSR flag changes (N, Z, C, V) for each instruction that modifies flags
- **Register Trace**: Analyzes register access patterns, identifies hot registers, detects unused registers, and flags read-before-write issues
- **Symbol-Aware Output**: All traces automatically show function/label names (e.g., `main+4`, `calculate`) instead of raw hex addresses for easier debugging

All diagnostic modes support both text and JSON output formats.

Example symbol-aware output:
```
# Stack trace showing function names
[000005] nested_call         : MOVE      SP: 0x00050000 -> 0x0004FFEC  (grow by 20 bytes)
[000010] helper1             : MOVE      SP: 0x0004FFEC -> 0x0004FFD8  (grow by 20 bytes)

# Flag trace showing symbol names
[000012] loop                : 0xE355000C                      ---- -> N*---  (changed: N)

# Coverage showing functions with symbols
0x00008000: executed      1 times (first: cycle      1, last: cycle      1) [main]
0x00008014: executed     13 times (first: cycle     12, last: cycle    462) [loop]
```

### Example Programs Status

#### All Example Programs Working! (49 total, 100%) ✅

All 49 example programs execute successfully:

**Non-Interactive Programs (46):**
- hello.s, loops.s, arithmetic.s, conditionals.s, functions.s
- factorial.s, recursive_fib.s, recursive_factorial.s
- string operations: strings.s, string_reverse.s (with stdin)
- data structures: arrays.s, linked_list.s, hash_table.s
- sorting algorithms: quicksort.s, bubble_sort.s
- literal pools: test_ltorg.s, test_org_0_with_ltorg.s
- multi-precision: add_128bit.s (128-bit integer addition with carry propagation)
- And 27+ more fully functional examples

**Interactive Programs (3):**
These programs work correctly when provided with stdin input:
- **bubble_sort.s** - Prompts for array size and elements (fully working)
- **calculator.s** - Interactive calculator with +, -, *, / operations (fully working)
- **fibonacci.s** - Prompts for count of Fibonacci numbers (fully working)

#### Recent Fixes (Dec 2025)
- **Security:** Fixed DoS vulnerability in stdin syscalls (bounded input to 4KB)
- **Parser:** Added octal escape sequence support (`\NNN` format)
- **Performance:** Optimized string building in trace output (strings.Builder)
- **Performance:** Eliminated per-call map allocation in trace.go
- **Performance:** Added memory bounds to RegisterTrace unique value tracking
- **TUI:** Fixed help command display (black-on-black text issue)

#### Previous Fixes (Oct 2025)
- **calculator.s** - Fixed infinite loop bug when stdin exhausted (EOF handling)
- **test_ltorg.s** - Fixed literal pool space reservation in parser
- **test_org_0_with_ltorg.s** - Fixed literal pool space reservation + added missing branch instruction

### Development Tools

Located in `tools/` directory:

- **Linter** - Analyze assembly code for issues (`tools/lint.go`)
  - Undefined label detection with suggestions
  - Unreachable code detection
  - Register usage warnings
  - 25 tests
- **Formatter** - Format assembly code consistently (`tools/format.go`)
  - Multiple format styles (default, compact, expanded)
  - Configurable alignment and spacing
  - 27 tests
- **Cross-Reference** - Generate symbol usage reports (`tools/xref.go`)
  - Symbol cross-reference with usage tracking
  - Function and data label identification
  - 21 tests

### Machine Code Encoder/Decoder

Located in `encoder/` directory:

- **Encoder** - Convert assembly to ARM machine code
- **Decoder** - Disassemble ARM machine code to assembly
- Supports all ARM2 instruction formats
- 1148 lines across 5 files
- Complete encoding/decoding for data processing, memory, branch, and multiply instructions

### Configuration

Configuration files are stored in platform-specific locations:
- **macOS/Linux:** `~/.config/arm-emu/config.toml`
- **Windows:** `%APPDATA%\arm-emu\config.toml`

See `config/config.go` for all available options.

### Note

Coreutils is installed on MacOS, so commands like `gtimeout` are available.
