# GUI Apps

The emulator has two desktop GUI apps. Both drive the same Go backend (`arm-emulator -api-server`) over the HTTP REST API and a WebSocket, and offer the same debugging features. Each starts the backend on launch and stops it on exit.

| | Swift app | Avalonia app |
|---|---|---|
| Platforms | macOS 26.2+ (Apple Silicon release build) | Windows 10+, macOS 13+, Linux |
| Toolkit | SwiftUI | Avalonia UI, ReactiveUI, .NET 10 |
| Source | `swift-gui/` | `avalonia-gui/` |
| Release artifact | `arm-emulator-swift-gui.app.tar.gz` | `arm-emulator-avalonia-<platform>` archive |
| Guide | [SWIFT_APP.md](SWIFT_APP.md) | [avalonia-gui/README.md](../avalonia-gui/README.md) |
| Shortcuts | [SWIFT_APP.md](SWIFT_APP.md#keyboard-shortcuts) | [KEYBOARD_SHORTCUTS.md](../avalonia-gui/KEYBOARD_SHORTCUTS.md) |

On macOS, use either. On Windows and Linux, use the Avalonia app.

## Features

- Editor with ARM syntax highlighting, line numbers, a breakpoint gutter and the current PC line
- Run, pause, step, step over, step out and restart; the editor source is assembled on demand when it has changed
- Registers with change highlighting and CPSR flags
- Memory, stack and disassembly views that follow execution
- Breakpoints, watchpoints and an expression evaluator
- Console for program output and input, including programs that wait for input
- Examples browser, recent files and persistent preferences

## Releases

Each release archive contains the app and the `arm-emulator` backend. See [installation.md](installation.md#method-1-download-pre-built-binaries-recommended) for the file list and the macOS first-launch step.

## API

Both apps use the same API. Backend changes must stay additive so neither app breaks. See [HTTP_API.md](HTTP_API.md) and [openapi.yaml](../openapi.yaml).
