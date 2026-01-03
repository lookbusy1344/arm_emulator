# ARM Emulator Swift macOS App

Native macOS application for the ARM2 emulator, built with SwiftUI and connected to the Go backend via HTTP REST API and WebSocket.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Building](#building)
- [Running](#running)
- [Development Workflow](#development-workflow)
- [Project Structure](#project-structure)
- [Key Components](#key-components)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Troubleshooting](#troubleshooting)

## Overview

The Swift app provides a native macOS experience for the ARM2 emulator with:

- **Native SwiftUI Interface** - Responsive, 60fps UI following macOS design guidelines
- **MVVM Architecture** - Clean separation of concerns with ViewModels managing state
- **Real-time Updates** - WebSocket connection for live register/output updates
- **Keyboard Shortcuts** - Standard macOS shortcuts (⌘L, ⌘R, ⌘T, etc.)
- **Zero External Dependencies** - Uses only Foundation, SwiftUI, and Combine
- **CLI-First Development** - XcodeGen-based project with swiftlint/swiftformat integration

**Important:** The Swift app automatically manages the Go API backend lifecycle. It finds, starts, and monitors the backend process - no manual startup required.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Swift macOS App                         │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                   Views (SwiftUI)                     │ │
│  │  MainView, EditorView, RegistersView, ConsoleView    │ │
│  └─────────────────────┬─────────────────────────────────┘ │
│                        │                                   │
│  ┌─────────────────────▼─────────────────────────────────┐ │
│  │              ViewModels (MVVM)                        │ │
│  │           EmulatorViewModel (@MainActor)              │ │
│  │  - Manages app state (@Published properties)          │ │
│  │  - Coordinates API/WebSocket clients                  │ │
│  │  - Handles user actions                               │ │
│  └─────────────┬───────────────────┬─────────────────────┘ │
│                │                   │                       │
│  ┌─────────────▼──────┐  ┌─────────▼──────────┐           │
│  │   APIClient        │  │  WebSocketClient   │           │
│  │  (HTTP/REST)       │  │  (Real-time)       │           │
│  │  - URLSession      │  │  - URLSession WS   │           │
│  │  - async/await     │  │  - Combine events  │           │
│  └─────────────┬──────┘  └─────────┬──────────┘           │
│                │                   │                       │
└────────────────┼───────────────────┼───────────────────────┘
                 │                   │
        HTTP/REST│                   │WebSocket
                 │                   │
┌────────────────▼───────────────────▼───────────────────────┐
│                  Go API Backend                            │
│         (http://localhost:8080)                            │
│                                                             │
│  - Session Management                                      │
│  - Program Loading/Execution                               │
│  - Register/Memory Inspection                              │
│  - Debugging (Breakpoints/Watchpoints)                     │
│  - Real-time Event Broadcasting                            │
└─────────────────────────────────────────────────────────────┘
```

### MVVM Pattern

- **Models** - Data structures (EmulatorSession, Register, ProgramState)
- **Views** - SwiftUI views (MainView, EditorView, etc.)
- **ViewModels** - State management and business logic (EmulatorViewModel)
- **Services** - API/WebSocket communication (APIClient, WebSocketClient)

### Data Flow

1. **User Action** → View → ViewModel method
2. **ViewModel** → APIClient (HTTP request)
3. **Backend** → WebSocket broadcast
4. **WebSocketClient** → ViewModel event handler
5. **ViewModel** → @Published property update
6. **View** → Automatic UI refresh (SwiftUI)

## Prerequisites

Install required tools via Homebrew:

```bash
brew install xcodegen swiftlint swiftformat xcbeautify
```

**System Requirements:**
- macOS 13.0+ (Ventura or later)
- Xcode Command Line Tools
- Go 1.21+ (for API backend)

## Getting Started

### 1. Generate Xcode Project

The Xcode project is generated from `project.yml` using XcodeGen:

```bash
cd swift-gui
xcodegen generate
```

**Important:** You must regenerate the project whenever you modify `project.yml` (e.g., adding files, changing settings).

### 2. Build the API Backend Binary

The Swift app needs the backend binary to be available:

```bash
# From project root
go build -o arm-emulator
```

The Swift app will automatically find and start this binary when launched. The backend lifecycle is fully managed - it starts on app launch and shuts down when the app quits.

### 3. Open in Xcode

```bash
cd swift-gui
open ARMEmulator.xcodeproj
```

Press `⌘R` to build and run.

## Building

### Build via Xcode

```bash
cd swift-gui
open ARMEmulator.xcodeproj
# Press Cmd+R to build and run
```

### Build via CLI

```bash
cd swift-gui

# Debug build
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator build | xcbeautify

# Release build
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator -configuration Release build | xcbeautify

# Clean build
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator clean build | xcbeautify
```

**Built App Location:**
```
~/Library/Developer/Xcode/DerivedData/ARMEmulator-*/Build/Products/Debug/ARMEmulator.app
```

### Build Artifacts

```
DerivedData/ARMEmulator-*/Build/Products/
├── Debug/
│   └── ARMEmulator.app        # Debug build (development)
└── Release/
    └── ARMEmulator.app        # Release build (optimized)
```

## Running

### Run from Xcode

1. Ensure backend binary is built: `go build -o arm-emulator`
2. Open Xcode: `cd swift-gui && open ARMEmulator.xcodeproj`
3. Press `⌘R` to run (backend starts automatically)

### Run from CLI

```bash
# Find and open the built app
find ~/Library/Developer/Xcode/DerivedData -name "ARMEmulator.app" -type d -exec open {} \; -quit
```

### Pre-load a File

To launch with a specific assembly file already loaded:

```bash
cd gui
wails dev -nocolour -appargs "../examples/hello.s"
```

Note: This feature is currently only available in the Wails GUI. Swift app file loading will be added in future releases.

## Development Workflow

### Typical Development Cycle

```bash
# 1. Ensure backend binary is built
go build -o arm-emulator

# 2. Open Xcode (backend will start automatically when app runs)
cd swift-gui
open ARMEmulator.xcodeproj

# 3. Make code changes in Xcode

# 4. Format and lint before committing
swiftformat .
swiftlint

# 5. Run tests
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify

# 6. Commit changes
git add .
git commit -m "feat: Add feature description"
```

### Hot Reload Workflow

Xcode supports SwiftUI hot reload:

1. Run app with `⌘R`
2. Make UI changes in SwiftUI files
3. Press `⌘R` again to see updates immediately

For logic changes in ViewModels/Services, full rebuild is required.

### Adding New Files

When adding new Swift files:

1. Add the file to the appropriate group in `project.yml`:

```yaml
targets:
  ARMEmulator:
    sources:
      - path: ARMEmulator/NewFeature
        name: NewFeature
```

2. Regenerate the Xcode project:

```bash
xcodegen generate
```

3. Open the project in Xcode to verify the file is included.

## Project Structure

```
swift-gui/
├── project.yml                      # XcodeGen configuration
├── .swiftlint.yml                   # SwiftLint rules
├── .swiftformat                     # SwiftFormat rules
├── .gitignore                       # Git ignore (Xcode user files)
│
├── ARMEmulator/                     # Main app target
│   ├── ARMEmulatorApp.swift         # App entry point
│   │
│   ├── Models/                      # Data models
│   │   ├── EmulatorSession.swift    # Event/session models
│   │   ├── ProgramState.swift       # VM state models
│   │   └── Register.swift           # Register/CPSR models
│   │
│   ├── Services/                    # Backend communication
│   │   ├── APIClient.swift          # HTTP REST client (247 lines)
│   │   └── WebSocketClient.swift    # WebSocket client (96 lines)
│   │
│   ├── ViewModels/                  # MVVM state management
│   │   └── EmulatorViewModel.swift  # Main ViewModel (213 lines)
│   │
│   ├── Views/                       # SwiftUI views
│   │   ├── MainView.swift           # Main window + toolbar
│   │   ├── EditorView.swift         # Assembly editor
│   │   ├── RegistersView.swift      # Register display
│   │   └── ConsoleView.swift        # Console output
│   │
│   └── Resources/
│       └── Info.plist               # App metadata
│
└── ARMEmulatorTests/                # Test target
    └── ARMEmulatorTests.swift       # Unit tests
```

### File Responsibilities

#### Models

- **EmulatorSession.swift** - WebSocket event models (state, output, event)
- **ProgramState.swift** - VM execution state
- **Register.swift** - Register values and CPSR flags

#### Services

- **APIClient.swift** - HTTP REST communication (create session, load program, run, step, etc.)
- **WebSocketClient.swift** - Real-time event streaming

#### ViewModels

- **EmulatorViewModel.swift** - Central state management, coordinates API/WebSocket, handles user actions

#### Views

- **MainView.swift** - Main window layout with split views and toolbar
- **EditorView.swift** - Assembly source editor with line numbers
- **RegistersView.swift** - Register display with CPSR flags
- **ConsoleView.swift** - Output console with stdin support

## Key Components

### APIClient

HTTP REST client using URLSession and async/await:

```swift
class APIClient {
    func createSession() async throws -> String
    func loadProgram(sessionId: String, source: String) async throws
    func run(sessionId: String) async throws
    func step(sessionId: String) async throws
    func stop(sessionId: String) async throws
    func reset(sessionId: String) async throws
    func getRegisters(sessionId: String) async throws -> RegisterState
}
```

### WebSocketClient

Real-time event streaming using URLSessionWebSocketTask:

```swift
class WebSocketClient: ObservableObject {
    @Published var isConnected = false
    let eventPublisher: AnyPublisher<EmulatorEvent, Never>

    func connect(sessionId: String)
    func disconnect()
}
```

### EmulatorViewModel

State management with @MainActor for UI thread safety:

```swift
@MainActor
class EmulatorViewModel: ObservableObject {
    @Published var registers: RegisterState
    @Published var consoleOutput: String
    @Published var executionState: ExecutionState
    @Published var error: String?

    func loadProgram(source: String) async
    func run() async
    func step() async
    func stop() async
    func reset() async
}
```

## Testing

### Run Tests via Xcode

1. Press `⌘U` to run all tests
2. View results in Test Navigator (⌘6)

### Run Tests via CLI

```bash
cd swift-gui

# Run all tests
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' | xcbeautify

# Run with coverage
xcodebuild test -project ARMEmulator.xcodeproj -scheme ARMEmulator -destination 'platform=macOS' -enableCodeCoverage YES | xcbeautify
```

### Test Status

**Current State:** Test infrastructure added, comprehensive tests delayed (Stage 6).

**Reason:** Complex async/await mocking for APIClient/WebSocketClient not time-effective at this stage.

**Recommendation:** Manual testing covers core workflows effectively. Integration tests with real backend planned for future releases.

See `SWIFT_GUI_PLANNING.md` Stage 6 for details.

## Code Quality

### Format Code

```bash
cd swift-gui
swiftformat .
```

Configuration: `.swiftformat`

### Lint Code

```bash
cd swift-gui
swiftlint
```

Auto-fix issues:

```bash
swiftlint --fix
```

Configuration: `.swiftlint.yml`

**Current Standard:** 0 violations required before commit.

### Pre-Commit Checklist

Before committing Swift code changes:

```bash
# 1. Format
swiftformat .

# 2. Lint
swiftlint

# 3. Build
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator build | xcbeautify

# 4. Verify linting passed (0 violations)
# 5. Commit
git add .
git commit -m "feat: Description"
```

## Troubleshooting

### "No such module 'SwiftUI'"

**Cause:** Xcode Command Line Tools not installed or misconfigured.

**Fix:**
```bash
xcode-select --install
sudo xcode-select -s /Applications/Xcode.app/Contents/Developer
```

### "Failed to connect to backend" or "Backend binary not found"

**Cause:** Backend binary not built or not in expected location.

**Fix:**
```bash
# Build the backend binary
cd /path/to/arm_emulator
go build -o arm-emulator
```

The app searches for `arm-emulator` in:
1. App bundle resources (production builds)
2. Current directory
3. Parent directory (for development from `swift-gui/` folder)

Verify the binary exists and is executable:
```bash
ls -la arm-emulator
# Should show executable permissions
```

### "Xcode project out of sync"

**Cause:** `project.yml` modified but Xcode project not regenerated.

**Fix:**
```bash
cd swift-gui
xcodegen generate
```

### "SwiftLint not found"

**Cause:** SwiftLint not installed or not in PATH.

**Fix:**
```bash
brew install swiftlint
```

### "DerivedData issues"

**Cause:** Corrupted build artifacts.

**Fix:**
```bash
# Clean DerivedData
rm -rf ~/Library/Developer/Xcode/DerivedData

# Rebuild
cd swift-gui
xcodebuild -project ARMEmulator.xcodeproj -scheme ARMEmulator clean build | xcbeautify
```

### "Code signing errors"

**Cause:** Certificate issues (rare with ad-hoc signing).

**Fix:**
The project uses automatic ad-hoc code signing for development. If you encounter issues:

1. Open `project.yml`
2. Verify `DEVELOPMENT_TEAM` is not set (uses ad-hoc)
3. Regenerate: `xcodegen generate`

### WebSocket Connection Drops

**Cause:** Backend restarted or network issues.

**Solution:** The WebSocket client automatically attempts reconnection. If it fails, restart the app.

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| ⌘L | Load program |
| ⌘R | Run program |
| ⌘T | Step instruction |
| ⌘. | Stop execution |
| ⌘⇧R | Reset VM |

## Performance

**Target Metrics:**
- UI frame time: < 16ms (60fps)
- API latency: < 10ms (localhost)
- WebSocket event latency: < 5ms
- Memory usage: < 100MB

**Optimization:**
- @MainActor ensures UI updates on main thread
- Async/await prevents blocking
- WebSocket events use Combine for reactive updates

## Related Documentation

- **API Documentation:** `API.md` - Complete REST/WebSocket API reference
- **OpenAPI Spec:** `openapi.yaml` - Machine-readable API specification
- **CLI Automation:** `docs/SWIFT_CLI_AUTOMATION.md` - General Swift CLI development guide
- **Planning:** `SWIFT_GUI_PLANNING.md` - Implementation roadmap and status
- **Project Guide:** `CLAUDE.md` - Project-specific development guidelines

## Future Enhancements

Planned features (see `SWIFT_GUI_PLANNING.md` Stage 4):

- Syntax highlighting for assembly code
- Breakpoint gutter in editor
- Disassembly view
- Memory hex dump
- Stack visualization
- File open/save dialogs
- Recent files menu
- Examples browser
- Preferences window

---

*Last Updated: 2026-01-02*
