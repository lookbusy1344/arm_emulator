# ARM Emulator GUI

## About

This is the Wails-based graphical user interface for the ARM2 Emulator. It provides a modern, cross-platform debugging environment with source view, register inspection, memory viewer, and breakpoint management.

The GUI is built with:
- **Backend:** Go with Wails v2.10.2
- **Frontend:** React 18 + TypeScript + Vite
- **Shared Service Layer:** Reuses the same `DebuggerService` as the TUI interface

For more information about the project configuration, see `wails.json` or visit: https://wails.io/docs/reference/project-config

## Prerequisites

- Go 1.25+
- Node.js 18+
- Wails CLI v2.9+ (for development mode only)

Install Wails CLI:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Quick Start

### Using Makefile (Recommended)

```bash
# Install all dependencies
make install

# Build production binary
make build

# Run all tests
make test

# Start development mode (requires Wails CLI)
make dev

# Clean build artifacts
make clean

# Show all available targets
make help
```

### Manual Build

```bash
# Install dependencies
cd frontend && npm install && cd ..
go mod download

# Build frontend
cd frontend && npm run build && cd ..

# Build Go application
go build -o ../build/arm-emulator-gui
```

## Live Development

To run in live development mode with hot reload:

```bash
# Using Makefile
make dev

# Or directly with Wails
wails dev
```

This will run a Vite development server with very fast hot reload of frontend changes. A dev server also runs on http://localhost:34115 where you can connect in your browser and call Go methods from devtools.

## Building

To build a redistributable, production mode package:

```bash
# Using Makefile
make build

# Or directly with Wails
wails build
```

The binary will be created in `../build/arm-emulator-gui`.

## Testing

```bash
# Run all tests (Go + frontend)
make test

# Run frontend tests in watch mode
make test-watch

# Run Go tests only
go test ./...

# Run frontend tests only
cd frontend && npm test
```

**Note:** Go tests require the frontend to be built first (frontend/dist must exist).

## Project Structure

```
gui/
├── wails.json          # Wails configuration
├── Makefile            # Build automation
├── main.go             # Application entry point
├── app.go              # Wails app bindings
├── app_test.go         # Go tests
├── go.mod              # Go dependencies
└── frontend/           # React frontend
    ├── src/
    │   ├── App.tsx           # Main application
    │   ├── components/       # React components
    │   ├── hooks/            # Custom hooks
    │   ├── services/         # API wrappers
    │   └── types/            # TypeScript types
    ├── package.json    # npm dependencies
    └── vite.config.ts  # Vite configuration
```

## Documentation

- [GUI Comprehensive Review](../docs/GUI_COMPREHENSIVE_REVIEW.md) - Detailed code review and analysis
- [GUI Executive Summary](../docs/GUI_REVIEW_EXECUTIVE_SUMMARY.md) - Quick overview and ratings
- [GUI Documentation](../docs/GUI.md) - User guide and features

## Troubleshooting

### Build fails with "pattern all:frontend/dist: no matching files found"

This means the frontend hasn't been built yet. Run:
```bash
cd frontend && npm install && npm run build
```

### Tests fail with context errors

This is expected - Wails context is only available when running the full application, not in unit tests. The frontend build must exist for tests to compile.

### "wails: command not found" in development mode

Install the Wails CLI:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Debug Logging

To enable detailed debug logging for troubleshooting GUI issues:

```bash
# Enable debug logging
export ARM_EMULATOR_DEBUG=1

# Run the GUI
./gui/gui

# Or in development mode
ARM_EMULATOR_DEBUG=1 wails dev
```

Debug logs are written to:
- `/tmp/arm-emulator-gui-debug.log` - GUI layer events and method calls
- `/tmp/arm-emulator-service-debug.log` - Service layer execution and state changes

This is useful for diagnosing issues with program execution, event handling, or state synchronization.

## Contributing

When making changes to the GUI:
1. Test both frontend and backend: `make test`
2. Ensure frontend builds successfully: `make frontend`
3. Run the application to verify UI changes: `make dev`
4. Update tests if adding new features
