# ARM Emulator Project

ARM emulator in Go with HTTP REST API backend. Two GUI frontends (Swift + Avalonia) connect to the same API.

**⚠️ CRITICAL: API SYNCHRONIZATION**
- Go backend API is shared by **both Swift GUI and Avalonia GUI**
- **DO NOT make breaking API changes** - only additive changes allowed
- Any API modifications must work with both frontends
- Test both GUIs after backend changes

## Go Backend (Core Emulator + API)

### Build & Test

```bash
# Build with version info
make build

# Format, lint, test (MANDATORY before commit)
go fmt ./...
golangci-lint run ./...
go clean -testcache && go test ./...

# Run emulator
./arm-emulator program.s
```

**Test Organization:** Tests in `./tests/unit/` and `./tests/integration/` (exceptions: `gui/app_test.go`, `debugger/tui_internal_test.go`)

## Swift GUI (macOS, Primary GUI)

**Prerequisites:** macOS 26.2, Swift 6.2, Xcode 26.2
```bash
brew install xcodegen swiftlint swiftformat xcbeautify
```

### Build & Test

```bash
cd swift-gui

# Generate Xcode project (after modifying project.yml)
xcodegen generate

# Build (requires Go backend: make build)
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator build | xcbeautify

# Test
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify

# Format & lint (MANDATORY before commit - 0 violations required)
swiftformat .
swiftlint
```

**Architecture:** MVVM with SwiftUI. Connects via HTTP REST API + WebSocket to Go backend. Uses modern Swift 6.2 features.

**Docs:** `SWIFT_GUI_PLANNING.md`, `docs/SWIFT_CLI_AUTOMATION.md`, `docs/MCP_UI_DEBUGGING.md`

## Avalonia GUI (Cross-Platform: Windows/macOS/Linux)

**Prerequisites:** .NET SDK 10.0+ from [https://dot.net](https://dot.net)

### Build & Test

```bash
cd avalonia-gui

# Build & run
dotnet build
dotnet run --project ARMEmulator

# Test
dotnet test

# Format (MANDATORY before commit - must build and pass tests)
# Note: dotnet format is run automatically after every file change
dotnet format
dotnet build
dotnet test
```

**Architecture:** MVVM with ReactiveUI. Connects via HTTP REST API + WebSocket to Go backend. Uses C# 13 features (primary constructors, collection expressions, records, pattern matching, immutable collections).

**Note:** `dotnet format` runs automatically after every change to the Avalonia project.

**Docs:** `docs/AVALONIA_IMPLEMENTATION_PLAN.md`

## Project Structure

```
├── api/           - HTTP REST API backend (port 8080)
├── service/       - Service layer for API/GUI integration
├── vm/            - Virtual machine implementation
├── parser/        - Assembly parser with macros
├── instructions/  - Instruction implementations
├── swift-gui/     - Swift native macOS GUI
├── avalonia-gui/  - Avalonia .NET cross-platform GUI
├── tests/         - Unit and integration tests
├── examples/      - 49 example ARM assembly programs
└── docs/          - Documentation
```

## Additional Documentation

- **Instructions & Syscalls:** [docs/INSTRUCTIONS.md](docs/INSTRUCTIONS.md)
- **Swift GUI Architecture:** `SWIFT_GUI_PLANNING.md`, `docs/SWIFT_CLI_AUTOMATION.md`, `docs/MCP_UI_DEBUGGING.md`
- **Avalonia GUI Plan:** `docs/AVALONIA_IMPLEMENTATION_PLAN.md`
- **Diagnostic Modes:** Code coverage, stack trace, flag trace, register trace (see `--help`)
