# GUI Documentation

> **⚠️ DEPRECATED**: The Wails GUI is deprecated in favor of the native Swift macOS app. See [SWIFT_APP.md](SWIFT_APP.md) for the recommended GUI. For quick reference on Wails commands and development, see [../Wails.md](../Wails.md).

## Overview

The ARM Emulator GUI provides a graphical interface for debugging ARM assembly programs. Built with Wails, it combines a Go backend with a React/TypeScript frontend for cross-platform compatibility.

## Features

- **Code Editor**: Write and edit ARM assembly with syntax highlighting
- **Register View**: Real-time display of all 16 registers, PC, and CPSR flags
- **Memory View**: Hex dump and ASCII representation of memory
- **Execution Control**: Step, run, pause, and reset program execution
- **Breakpoints**: Set and manage breakpoints (double-click on source or disassembly lines)
- **Symbol Resolution**: View and navigate to labels and symbols
- **Keyboard Shortcuts**: Fast debugging with function key shortcuts

### Keyboard Shortcuts

The GUI supports keyboard shortcuts matching the TUI interface:

- **F5** - Continue (run until breakpoint or halt)
- **F9** - Toggle breakpoint at current PC
- **F10** - Step Over (next instruction)
- **F11** - Step (single instruction)

These shortcuts work when focus is not in an input field, providing fast debugging without reaching for the mouse.

## Building

### Prerequisites

- Go 1.25+
- Node.js 18+
- Wails CLI v2.9+

Install Wails:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Development Build

```bash
cd gui
wails dev
```

This starts the development server with hot reload.

### Production Build

```bash
cd gui
wails build
```

Binary will be created in `gui/build/bin/`.

## Architecture

### Backend (Go)

The backend uses the existing `service` package to provide a thread-safe interface to the emulator. All business logic resides in Go, with the frontend handling only UI concerns.

Key components:
- `service.DebuggerService`: Core emulator interface
- `gui/app.go`: Wails application bindings
- `gui/main.go`: Application entry point

### Frontend (React/TypeScript)

The frontend is a React application built with TypeScript and Vite.

Key components:
- `App.tsx`: Main application container
- `RegisterView.tsx`: Register and CPSR display
- `MemoryView.tsx`: Memory hex dump
- `useEmulator.ts`: State management hook
- `wails.ts`: Typed API bindings

### Communication

Wails provides bidirectional communication between Go and JavaScript:
- Frontend calls Go methods via `window.go.main.App.*`
- Go can emit events to frontend (for real-time updates)
- All data serialized as JSON

## Testing

### Backend Tests

```bash
# Unit tests
go test ./service/...

# Integration tests
go test ./tests/integration/gui_test.go -v
```

### Frontend Tests

```bash
cd gui/frontend

# Run all tests
npm test

# Run with coverage
npm test -- --coverage

# Run specific test
npm test -- RegisterView
```

## Development Tips

### Hot Reload

In development mode (`wails dev`), both frontend and backend support hot reload. Frontend changes reload instantly, Go changes trigger app restart.

### Debugging

**Frontend:**
- Open DevTools in the app (Cmd+Option+I on macOS)
- Console logs visible in DevTools
- React DevTools extension works

**Backend:**
- Add breakpoints in Go code
- Use `wails dev -debug` for additional logging
- Check terminal output for Go logs

### Adding New Features

1. Add method to `service/debugger_service.go`
2. Add test to `tests/unit/service/`
3. Expose method in `gui/app.go`
4. Add binding in `gui/frontend/src/services/wails.ts`
5. Update `useEmulator` hook if needed
6. Add UI component
7. Write component tests

## Common Issues

**Build fails with "command not found: wails"**
- Ensure `~/go/bin` is in your PATH
- Run: `export PATH=$PATH:~/go/bin`

**Frontend doesn't update**
- Clear Vite cache: `rm -rf gui/frontend/.vite`
- Reinstall deps: `cd gui/frontend && npm ci`

**"Cannot read property 'go' of undefined"**
- Wails bindings not injected (only available in built app, not in browser)
- Use mock for testing: see `wails.test.ts`

## Performance

The GUI is designed for smooth performance even with large programs:

- Register updates are throttled during continuous execution
- Memory view uses virtual scrolling (planned)
- Execution runs on separate goroutine to avoid blocking UI
- Frontend uses React.memo and useMemo for expensive renders

## Accessibility

- Keyboard navigation supported throughout
- Function key shortcuts (F5, F9, F10, F11) for common debugging operations
- ARIA labels on all interactive elements
- High contrast color scheme
- Screen reader compatible (via semantic HTML)
