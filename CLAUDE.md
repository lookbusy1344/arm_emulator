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

**Note:** The Swift app is the primary GUI for this project and automatically manages the Go HTTP API backend lifecycle - it finds, starts, and monitors the backend process. No manual backend startup required.

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

**When DebugLog Doesn't Work (Terminal Output)**
- `DebugLog` uses `print()` which writes to stdout
- When running macOS GUI apps from the terminal, stdout may not be captured
- **For debugging GUI apps from terminal, use `NSLog()`** - it writes to stderr and system log
- Example: `NSLog("🔵 [Category] Message: %@", value)`
- `NSLog()` output appears in Console.app and terminal stderr
- Once debugging is complete, remove `NSLog()` calls - they bypass Swift's type system
- For production code, use `DebugLog` (works in Xcode) or `os_log` for system-level logging

### Common Pitfalls

1. **"No such module 'SwiftUI'" error**: Ensure Xcode Command Line Tools are installed: `xcode-select --install`
2. **Xcode project out of sync**: Run `xcodegen generate` after modifying `project.yml`
3. **App can't find backend binary**: Ensure `arm-emulator` is built in the project root: `go build -o arm-emulator`
4. **SwiftLint/SwiftFormat not found**: Install via Homebrew: `brew install swiftlint swiftformat`
5. **Code signing errors**: The project uses automatic code signing (ad-hoc for development)
6. **DerivedData issues**: Clean with `rm -rf ~/Library/Developer/Xcode/DerivedData` if builds behave strangely
7. **Git tracking Xcode user files**: The `swift-gui/.gitignore` excludes user-specific files like `*.xcuserstate`, `xcuserdata/`, and build artifacts. These should never be committed.
8. **SwiftUI/AppKit rendering issues**: When NSViewRepresentable views don't render content (blank view despite data being present), use divide-and-conquer debugging: comment out complex custom components (ruler views, overlays, custom drawing) to isolate the issue. Get the basic AppKit view working first, then add complexity back incrementally. Example: NSTextView not displaying text may be caused by NSRulerView interference.

### Debugging with MCP Servers

The XcodeBuild and Playwright MCP servers enable automated debugging and testing of both native macOS apps and web UIs. For comprehensive documentation, see [docs/MCP_UI_DEBUGGING.md](docs/MCP_UI_DEBUGGING.md).

#### MCP Setup

**XcodeBuild MCP** (iOS/macOS automation):
```bash
# Correct installation command
claude mcp add --transport stdio XcodeBuildMCP -- npx -y xcodebuildmcp@latest

# Install AXe for UI automation
brew install cameroncooke/axe/axe
```

**Playwright MCP** (web UI automation):
```bash
# Installation
claude mcp add playwright npx @playwright/mcp@latest
```

#### Critical Prerequisite: Check Schema First

**MANDATORY**: Always check tool schema using `mcp-cli info <server>/<tool>` before any `mcp-cli call` command. This is non-negotiable.

```bash
# ALWAYS do this first
mcp-cli info XcodeBuildMCP/build_macos

# Then make the call
mcp-cli call XcodeBuildMCP/build_macos '{}'
```

#### XcodeBuild MCP Workflows

**1. Set Session Defaults (Critical First Step)**

Set defaults once to avoid repeating parameters in every call:

```bash
# For macOS app
mcp-cli call XcodeBuildMCP/session_set_defaults '{
  "projectPath": "/path/to/ARMEmulator.xcodeproj",
  "scheme": "ARMEmulator"
}'

# For iOS app with simulator
mcp-cli call XcodeBuildMCP/session_set_defaults '{
  "projectPath": "/path/to/MyApp.xcodeproj",
  "scheme": "MyApp",
  "simulatorName": "iPhone 16"
}'
```

**2. Build and Launch**

```bash
# Build the app
mcp-cli call XcodeBuildMCP/build_macos '{}'

# Get app bundle path
mcp-cli call XcodeBuildMCP/get_mac_app_path '{}'

# Launch with arguments
mcp-cli call XcodeBuildMCP/launch_mac_app '{
  "appPath": "/path/to/ARMEmulator.app",
  "args": ["examples/fibonacci.s"]
}'
```

**3. Log Capture (Programmatic Debug Access)**

Capture app logs programmatically for automated debugging:

```bash
# Start log capture with console output (captures print/NSLog/DebugLog)
mcp-cli call XcodeBuildMCP/start_sim_log_cap '{
  "bundleId": "com.example.ARMEmulator",
  "captureConsole": true
}'
# Returns: { "sessionId": "abc123" }

# Interact with app (trigger the bug, etc.)
# ...

# Stop capture and retrieve logs
mcp-cli call XcodeBuildMCP/stop_sim_log_cap '{
  "logSessionId": "abc123"
}'
# Returns logs with all debug output
```

**IMPORTANT:** Use `captureConsole: true` to capture `print()`, `NSLog()`, and custom debug utilities like `DebugLog`. Without it, you only get structured `os_log` entries.

**4. UI Automation with AXe**

```bash
# Get UI hierarchy (critical - don't guess coordinates!)
mcp-cli call XcodeBuildMCP/describe_ui '{}'

# Tap at coordinates
mcp-cli call XcodeBuildMCP/tap '{"x": 100, "y": 200}'

# Tap by accessibility label
mcp-cli call XcodeBuildMCP/tap '{"accessibilityId": "RunButton"}'

# Type text
mcp-cli call XcodeBuildMCP/type_text '{"text": "hello"}'

# Gesture presets
mcp-cli call XcodeBuildMCP/gesture '{"gesture": "scroll-down"}'

# Screenshot for verification
mcp-cli call XcodeBuildMCP/screenshot '{}'
```

**5. Direct AXe CLI Usage**

AXe can be used directly for shell scripts and CI pipelines:

```bash
# Get simulator UDID
UDID=$(xcrun simctl list devices booted -j | jq -r '.devices[][] | select(.state=="Booted") | .udid' | head -1)

# Tap at coordinates
axe tap -x 100 -y 200 --udid $UDID

# Gesture presets (easier than custom swipes!)
axe gesture scroll-up --udid $UDID
axe gesture swipe-from-left-edge --udid $UDID  # Back navigation

# Type text
echo "user@example.com" | axe type --stdin --udid $UDID

# Get UI hierarchy
axe describe-ui --udid $UDID

# Screenshot
axe screenshot --output ~/Desktop/test.png --udid $UDID
```

#### Playwright MCP Workflows (Wails GUI)

**1. Navigate and Capture Snapshot**

Key workflow: **navigate → snapshot → interact using refs**

```bash
# Navigate to page
mcp-cli call playwright/browser_navigate '{"url": "http://localhost:34115"}'

# Get accessibility tree (provides element refs)
mcp-cli call playwright/browser_snapshot '{}'
# Returns: - button "Run" [ref=e5]
#          - textbox "Code" [ref=e4]

# Click using ref
mcp-cli call playwright/browser_click '{"element": "Run button", "ref": "e5"}'
```

**2. Inspect Network & Console**

```bash
# Get console errors
mcp-cli call playwright/browser_console_messages '{"level": "error"}'

# List network requests
mcp-cli call playwright/browser_network_requests '{}'

# Execute JavaScript
mcp-cli call playwright/browser_evaluate '{
  "function": "() => document.querySelector(\"#result\").textContent"
}'
```

**3. Screenshots and Testing**

```bash
# Viewport screenshot
mcp-cli call playwright/browser_take_screenshot '{}'

# Full page screenshot
mcp-cli call playwright/browser_take_screenshot '{"fullPage": true}'

# Element screenshot
mcp-cli call playwright/browser_take_screenshot '{
  "element": "Chart",
  "ref": "e8"
}'
```

#### Automated Debugging Workflow Example

Complete automation for reproducing Swift GUI bugs:

```bash
#!/bin/bash
# debug_swift_gui.sh - Automated build-launch-debug cycle

set -e

echo "=== Building Swift GUI ==="
mcp-cli call XcodeBuildMCP/build_macos '{}'

echo "=== Stopping old instances ==="
killall ARMEmulator 2>/dev/null || true
sleep 1

echo "=== Launching app with test file ==="
APP_PATH=$(mcp-cli call XcodeBuildMCP/get_mac_app_path '{}' | jq -r '.appPath')
mcp-cli call XcodeBuildMCP/launch_mac_app "{
  \"appPath\": \"$APP_PATH\",
  \"args\": [\"$(pwd)/examples/fibonacci.s\"]
}"

# Wait for backend to start
sleep 3

echo "=== Starting log capture ==="
SESSION=$(mcp-cli call XcodeBuildMCP/start_sim_log_cap '{
  "bundleId": "com.example.ARMEmulator",
  "captureConsole": true
}' | jq -r '.sessionId')

echo "=== Testing via API ==="
# Create API session and trigger bug
# ...

echo "=== Retrieving logs ==="
LOGS=$(mcp-cli call XcodeBuildMCP/stop_sim_log_cap "{\"logSessionId\": \"$SESSION\"}")
echo "$LOGS" | jq -r '.logs' > debug.log

echo "=== Analyzing errors ==="
grep -E "❌|Error" debug.log
```

**Benefits of Automation:**
- Reproducibility: Exact same steps every time
- Speed: Full cycle takes ~10 seconds vs 2+ minutes manually
- Isolation: Test GUI and backend independently
- Documentation: Script serves as executable documentation

#### Best Practices

1. **Always check schema first** - Use `mcp-cli info` before any `mcp-cli call`
2. **Set session defaults early** - Call `session_set_defaults` once at start
3. **Use describe_ui before interactions** - Never guess coordinates from screenshots
4. **Capture logs with console output** - Use `captureConsole: true` for debug logs
5. **Start with snapshots** - Get current state before interacting (both Playwright and XcodeBuild)
6. **Clean builds for fresh state** - Use `clean` when debugging build issues

#### When to Use MCP Debugging

**Use Automation When:**
- Reproducing specific bug scenarios repeatedly
- Testing multiple edge cases quickly
- Verifying fixes across different inputs
- Running regression tests
- CI/CD integration

**Use Manual Debugging When:**
- Exploring unknown behavior
- UI layout/visual issues
- Initial bug triage

#### Troubleshooting

**Problem:** App doesn't launch
- Check backend binary built: `ls -la arm-emulator`
- Fix: `make build`

**Problem:** API connection refused
- Check backend started (look for "API server starting")
- Fix: Add `sleep 3` after launch

**Problem:** No debug logs appear
- Check building in Debug configuration (Release strips DebugLog)
- Fix: `mcp-cli call XcodeBuildMCP/build_macos '{"configuration": "Debug"}'`

**Problem:** "Element not found" (Playwright)
- Capture fresh snapshot to get updated refs
- Ensure page has loaded completely

**Problem:** UI automation not working
- Install AXe: `brew install cameroncooke/axe/axe`
- Ensure simulator is booted and app is running

See [docs/MCP_UI_DEBUGGING.md](docs/MCP_UI_DEBUGGING.md) for complete documentation with detailed examples, advanced techniques, and comprehensive troubleshooting.

## GUI Commands (Wails) - ⚠️ DEPRECATED

> **IMPORTANT:** The Wails GUI is deprecated in favor of the native Swift app. It remains available for reference and cross-platform testing, but is no longer actively developed. Use the Swift GUI (above) for all new development.

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

**IMPORTANT:** After implementing each phase of development, update `PROGRESS.md` to reflect the completed work, including:
- Mark the phase as completed
- Document any implementation details or deviations from the original plan
- Update the status of related tasks
- Note outstanding work and issues in `TODO.md`

**IMPORTANT:** Always run `go fmt ./...`, `golangci-lint run ./...`, and `go build -o arm-emulator && go clean -testcache && go test ./...` after making changes and **BEFORE committing** to ensure code quality and correctness. Linting must pass with 0 issues before any commit.

### Test-Driven Development (TDD)

**CRITICAL:** This project follows strict Test-Driven Development (TDD) practices. All development must follow the red-green-refactor cycle:

1. **Red**: Write a failing test that defines the desired behavior
2. **Green**: Write the minimal code to make the test pass
3. **Refactor**: Clean up the code while keeping tests passing

**TDD Requirements:**
- **Write tests FIRST**: Before implementing any feature or fix, write the test that validates the expected behavior
- **Comprehensive coverage**: All new code must be covered by tests
- **Test types**:
  - Unit tests for individual functions and components (in `tests/unit/`)
  - Integration tests for complete workflows and programs (in `tests/integration/`)
- **Tests as documentation**: Tests should clearly demonstrate how the code is intended to be used
- **Run tests frequently**: Execute tests after each small change to catch regressions immediately
- **Test quality**: Tests should be clear, focused, and test one thing at a time

**IMPORTANT:** Do not delete tests without explicit instructions. Do not simplify tests because they fail. If you think a test is malfunctioning, think about it carefully and ask me before making any changes to the tests.

**IMPORTANT:** Tests serve as both validation and documentation. When reviewing code, always check that corresponding tests exist and properly validate the intended behavior.

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
