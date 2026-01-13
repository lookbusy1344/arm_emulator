# Wails GUI Documentation

> **⚠️ DEPRECATED**: The Wails GUI is deprecated in favor of the native Swift macOS app. It remains available for reference and cross-platform testing, but is **no longer actively developed**. For new development, use the Swift GUI (see [docs/SWIFT_APP.md](docs/SWIFT_APP.md)).

## Overview

The Wails GUI is a cross-platform graphical interface for the ARM Emulator, built with:
- **Backend**: Go (using the existing emulator service layer)
- **Frontend**: React + TypeScript + Vite
- **Framework**: Wails v2.9+ (combines Go and web technologies)

The GUI provides a visual debugging environment with code editor, register view, memory inspector, and execution control.

## Quick Reference

### Prerequisites

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Verify installation
wails doctor
```

### Development Mode

```bash
cd gui
wails dev -nocolour  # IMPORTANT: Always use -nocolour flag
```

**Note**: The `-nocolour` flag prevents ANSI escape codes in output, which can cause issues with log processing and CI/CD pipelines.

### Load File at Startup

Launch the GUI with a specific assembly file pre-loaded:

```bash
cd gui
wails dev -nocolour -appargs "../examples/stack.s"
```

The window title will show "ARM Emulator - filename.s" when a file is loaded.

### Production Build

```bash
cd gui
wails build -nocolour
./build/bin/arm-emulator
```

Binary location: `gui/build/bin/arm-emulator`

### Check Environment

```bash
wails doctor
```

Validates Go version, Node.js version, platform dependencies, and Wails installation.

## Features

- **Code Editor**: Write and edit ARM assembly with syntax highlighting
- **Register View**: Real-time display of all 16 registers, PC, and CPSR flags
- **Memory View**: Hex dump and ASCII representation of memory
- **Execution Control**: Step, run, pause, and reset program execution
- **Breakpoints**: Set and manage breakpoints (double-click on source or disassembly lines)
- **Symbol Resolution**: View and navigate to labels and symbols
- **Keyboard Shortcuts**: Fast debugging with F5, F9, F10, F11

See [docs/GUI.md](docs/GUI.md) for complete feature documentation.

## E2E Testing

End-to-end tests validate the complete GUI workflow using Playwright.

**CRITICAL**: E2E tests require the Wails dev server to be running first. Tests will hang indefinitely if the backend is not available.

### Running E2E Tests

```bash
# Terminal 1: Start Wails dev server
cd gui
wails dev -nocolour

# Terminal 2: Wait for server to start (watch for "API server starting on http://localhost:34115")
# Then run E2E tests
cd gui/frontend
npm run test:e2e                           # Run all tests (all browsers)
npm run test:e2e -- --project=chromium    # Run chromium only
npm run test:e2e -- --project=firefox     # Run firefox only
npm run test:e2e -- --project=webkit      # Run webkit only
npm run test:e2e:headed                    # Run with visible browser
```

### E2E Test Coverage

The test suite validates:
- Program loading and parsing
- Code execution (step, continue, reset)
- Breakpoint management
- Register updates during execution
- Memory view updates
- Symbol resolution
- Error handling

See [gui/frontend/e2e/README.md](gui/frontend/e2e/README.md) for complete testing documentation.

## Architecture

### Backend (Go)

The Wails backend uses the existing `service` package to provide a thread-safe interface to the emulator:

- `service.DebuggerService`: Core emulator interface (shared with TUI and API)
- `gui/app.go`: Wails application bindings
- `gui/main.go`: Application entry point

All business logic resides in Go. The frontend handles only UI concerns.

### Frontend (React/TypeScript)

Built with modern React + TypeScript + Vite:

- `App.tsx`: Main application container
- `RegisterView.tsx`: Register and CPSR display
- `MemoryView.tsx`: Memory hex dump
- `useEmulator.ts`: State management hook
- `wails.ts`: Typed API bindings

### Communication

Wails provides bidirectional Go ↔ JavaScript communication:
- Frontend calls Go methods via `window.go.main.App.*`
- Go can emit events to frontend (for real-time updates)
- All data serialized as JSON

## Development Workflow

### Frontend Development

```bash
cd gui/frontend

# Install dependencies
npm install

# Run type checking
npm run type-check

# Run linter
npm run lint

# Run tests
npm test

# Run tests with coverage
npm test -- --coverage
```

### Backend Development

```bash
# Run Go tests
go test ./service/...

# Run GUI integration tests
go test ./tests/integration/gui_test.go -v

# Format Go code
go fmt ./gui/...

# Lint Go code
golangci-lint run ./gui/...
```

### Adding New Features

1. **Backend**: Add method to `service/debugger_service.go` with tests
2. **Wails Binding**: Expose method in `gui/app.go`
3. **Frontend**: Add TypeScript binding in `wails.ts`
4. **State Management**: Update `useEmulator` hook if needed
5. **UI Component**: Create/update React component
6. **Tests**: Write component tests and E2E tests

## Why Deprecated?

The Wails GUI was deprecated in favor of the native Swift macOS app for several reasons:

**Swift GUI Advantages:**
- True native macOS experience (no webview)
- Better performance and resource usage
- Native keyboard shortcuts and menu integration
- SwiftUI provides modern, declarative UI
- Automatic backend lifecycle management
- Superior debugging with Xcode

**Wails GUI Limitations:**
- Webview performance overhead
- Non-native look and feel
- More complex build process
- Harder to debug (split between Go and JS)
- Cross-platform promise not realized (only developed for macOS)

## When to Use Wails GUI

The Wails GUI is still useful for:
- **Cross-platform development**: If targeting Linux or Windows
- **Web technology familiarity**: If team knows React but not Swift
- **Reference implementation**: Demonstrating alternative GUI approach
- **Testing**: Validating service layer works with different UI frameworks

For macOS development, **always prefer the Swift GUI**.

## Complete Documentation

For comprehensive documentation including detailed architecture, testing strategies, performance optimization, and troubleshooting, see:

- **[docs/GUI.md](docs/GUI.md)** - Complete Wails GUI guide
- **[gui/frontend/e2e/README.md](gui/frontend/e2e/README.md)** - E2E testing documentation
- **[API.md](API.md)** - REST API reference (shared backend)
- **[docs/SWIFT_APP.md](docs/SWIFT_APP.md)** - Recommended Swift GUI guide

## Common Issues

**Build fails with "command not found: wails"**
- Ensure `~/go/bin` is in your PATH
- Run: `export PATH=$PATH:~/go/bin`

**Frontend doesn't update in dev mode**
- Clear Vite cache: `rm -rf gui/frontend/.vite`
- Reinstall deps: `cd gui/frontend && npm ci`

**E2E tests hang indefinitely**
- Ensure Wails dev server is running first
- Check server is accessible at http://localhost:34115
- Look for "API server starting" message in terminal

**"Cannot read property 'go' of undefined"**
- Wails bindings only available in built app, not browser
- Use mock for testing: see `wails.test.ts`

**ANSI escape codes in output**
- Always use `-nocolour` flag: `wails dev -nocolour`

## Migration to Swift GUI

If you're currently using the Wails GUI and want to migrate to the Swift GUI:

1. The Swift GUI uses the same REST API backend (port 8080)
2. All emulator functionality is identical
3. No changes needed to example programs or test files
4. Swift GUI has feature parity plus native macOS benefits

See [docs/SWIFT_APP.md](docs/SWIFT_APP.md) for Swift GUI setup and usage.
