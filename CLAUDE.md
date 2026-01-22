# ARM Emulator Project

This is an ARM emulator written in Go that implements a subset of the ARM2 instruction set.

## Build Command

```bash
# Build with version information (recommended)
make build

# Or use the build script
./build_with_version.sh

# Build and run tests with version info
./build_test.sh

# Manual build (basic, no version info)
go build -o arm-emulator
```

The Makefile, build script, and test script automatically inject version information from git (tag, commit hash, and build date) into the binary.

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

**Excluded Tests (non-deterministic output):**
- `test_get_random.s` - Uses GET_RANDOM syscall
- `test_get_time.s` - Uses GET_TIME syscall

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

## Swift Native macOS App Commands (Primary GUI)

**Note:** The Swift app automatically manages the Go backend lifecycle - no manual startup required.

### ⚠️ CRITICAL: ALWAYS FORMAT AND LINT SWIFT CODE ⚠️

After ANY Swift code changes, run `swiftformat .` and `swiftlint` immediately. See [Code Quality (Swift)](#code-quality-swift) section below for details. **0 violations required before any commit.**

### Platform Requirements

**IMPORTANT:** This Swift project targets modern platforms only - no backward compatibility required.

- **macOS:** 26.2
- **Swift:** 6.2
- **Xcode:** 26.2

The project uses the latest SwiftUI APIs and Swift language features. Always use modern Swift/SwiftUI capabilities without concern for older OS versions.

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

**IMPORTANT:** Before building the Swift app, ensure the Go backend is built with version information:

```bash
# Build Go backend first
cd ..
make build  # or ./build_with_version.sh
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

**MANDATORY - Run after ANY Swift code changes:**

```bash
cd swift-gui
swiftformat .        # Format code
swiftlint            # Check violations
swiftlint --fix      # Auto-fix if needed
swiftlint            # Verify 0 violations
```

**Requirements:**
- 0 violations (0 warnings, 0 errors) before any commit
- Run IMMEDIATELY after editing, BEFORE committing/completing work
- Exception: Test files >500 lines may use `// swiftlint:disable file_length`

### Swift App Architecture

The Swift app follows MVVM architecture and connects to the Go backend via:
- **HTTP REST API** (`APIClient.swift`) - For commands (run, step, reset, load program)
- **WebSocket** (`WebSocketClient.swift`) - For real-time updates (registers, console output, execution state)

See `SWIFT_GUI_PLANNING.md` for detailed architecture documentation and `docs/SWIFT_CLI_AUTOMATION.md` for comprehensive CLI development guide.

#### VM Execution States

Swift `VMState` enum must match backend `ExecutionState` values (see `swift-gui/ARMEmulator/Models/ProgramState.swift`):

| State | Description | Editor Editable? |
|-------|-------------|------------------|
| `idle` | No program loaded | ✅ Yes |
| `running` | Executing | ❌ No |
| `breakpoint` | Stopped at breakpoint/step | ❌ No |
| `halted` | Program finished | ✅ Yes |
| `error` | Error occurred | ✅ Yes |
| `waiting_for_input` | Blocked on stdin | ❌ No |

**Editor rule:** Editable only when fully stopped (idle/halted/error), read-only during execution.

**Note:** WebSocket events use `"breakpoint_hit"` but status responses use `"breakpoint"` (both map to `.breakpoint`).

#### Debug Logging

Use `DebugLog` utility (conditionally compiled, zero overhead in RELEASE):

```swift
DebugLog.log("Loading program", category: "ViewModel")
DebugLog.network("POST /api/v1/session/\(sessionID)/run")
DebugLog.error("Failed: \(error)", category: "ViewModel")
```

**For terminal debugging:** Use `NSLog()` (writes to stderr/system log) instead of `DebugLog.print()`. Remove after debugging.

### Common Pitfalls

1. **Xcode project out of sync**: Run `xcodegen generate` after modifying `project.yml`
2. **App can't find backend**: Build backend first: `make build`
3. **Tools not found**: `brew install xcodegen swiftlint swiftformat xcbeautify`
4. **DerivedData issues**: Clean with `rm -rf ~/Library/Developer/Xcode/DerivedData`
5. **NSViewRepresentable blank view**: Use divide-and-conquer - comment out complex components (rulers, overlays) and add back incrementally

### Debugging with MCP Servers

**Setup:**
```bash
# XcodeBuild MCP (macOS/iOS automation)
claude mcp add --transport stdio XcodeBuildMCP -- npx -y xcodebuildmcp@latest
brew install cameroncooke/axe/axe

# Playwright MCP (web UI automation)
claude mcp add playwright npx @playwright/mcp@latest
```

**MANDATORY:** Always check schema first: `mcp-cli info <server>/<tool>` before any `mcp-cli call`.

**Quick workflow:**
1. Set session defaults: `mcp-cli call XcodeBuildMCP/session_set_defaults`
2. Build: `mcp-cli call XcodeBuildMCP/build_macos '{}'`
3. Launch: `mcp-cli call XcodeBuildMCP/launch_mac_app`
4. Capture logs: `mcp-cli call XcodeBuildMCP/start_sim_log_cap` (use `captureConsole: true`)
5. UI automation: `mcp-cli call XcodeBuildMCP/describe_ui`, then tap/type/gesture

See [docs/MCP_UI_DEBUGGING.md](docs/MCP_UI_DEBUGGING.md) for comprehensive documentation, examples, and troubleshooting.

## GUI Commands (Wails) - ⚠️ DEPRECATED

> **DEPRECATED:** Use the native Swift app for all development. Wails GUI remains for reference only.

**Quick commands:**
```bash
# Development
cd gui
wails dev -nocolour

# Build
wails build -nocolour

# E2E tests (requires running dev server)
cd gui/frontend
npm run test:e2e -- --project=chromium
```

Always use `-nocolour` flag to prevent ANSI escape codes.

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
- `swift-gui/` - Swift native macOS app (SwiftUI + MVVM, connects to API backend) - **Primary GUI**
- `gui/` - Wails cross-platform GUI (Go + Svelte web frontend) - **DEPRECATED**
- `tests/` - Test files (1,024 tests, 100% pass rate)
  - `tests/unit/` - Unit tests for all packages
  - `tests/integration/` - Integration tests for complete programs
- `examples/` - Example ARM assembly programs (49 programs, all fully functional including 3 interactive)
- `docs/` - User and developer documentation

## SWI Syscall Reference

The emulator implements traditional ARM2 syscall convention: `SWI #immediate_value` where the syscall number is encoded directly in the instruction. Arguments and return values use registers R0-R2.

**Common syscalls:**
- `SWI #0x00` - EXIT (R0: exit code)
- `SWI #0x01` - WRITE_CHAR (R0: character)
- `SWI #0x02` - WRITE_STRING (R0: string address)
- `SWI #0x04` - READ_CHAR (returns R0: character)
- `SWI #0x06` - READ_INT (returns R0: integer)
- `SWI #0x10` - OPEN (R0: filename, R1: mode → R0: fd)
- `SWI #0x20` - ALLOCATE (R0: size → R0: address)

For the complete syscall reference including file operations, memory management, system information, error handling, and debugging support, see [docs/INSTRUCTIONS.md](docs/INSTRUCTIONS.md#system-instructions).

**Note:** CPSR flags (N, Z, C, V) are preserved across all syscalls.

## Development Guidelines

**IMPORTANT:** Always run `go fmt ./...`, `golangci-lint run ./...`, and `go build -o arm-emulator && go clean -testcache && go test ./...` after making changes and **BEFORE committing** to ensure code quality and correctness. Linting must pass with 0 issues before any commit.

### Test-Driven Development (TDD)

**CRITICAL:** Follow strict TDD practices (red-green-refactor cycle):

1. **Red**: Write failing test first
2. **Green**: Write minimal code to pass
3. **Refactor**: Clean up while keeping tests passing

**Requirements:**
- Write tests FIRST before implementing
- All new code must be covered by tests (unit in `tests/unit/`, integration in `tests/integration/`)
- Tests serve as both validation and documentation
- Run tests frequently: `go build -o arm-emulator && go clean -testcache && go test ./...`

**Rules:**
- Do NOT delete or simplify failing tests without permission - ask first
- Do NOT modify example programs to make them work - fix the emulator instead
- Use traditional ARM2 syscalls (`SWI #immediate`) NOT Linux-style (`SVC #0` with R7)
- Avoid magic numbers - use named constants
- Code reviews: assume suspicious implementation, verify thoroughly

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

All 49 example programs working (100%):
- **46 non-interactive:** hello.s, loops.s, arithmetic.s, factorial.s, sorting algorithms, data structures, literal pools, add_128bit.s, etc.
- **3 interactive:** bubble_sort.s, calculator.s, fibonacci.s (require stdin)
