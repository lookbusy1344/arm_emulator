# ARM Emulator - Avalonia .NET GUI

Cross-platform desktop GUI for the ARM Emulator built with Avalonia UI and .NET 10.

## Prerequisites

- .NET 10 SDK
- Platform-specific requirements:
  - **Windows:** Windows 10 or later
  - **macOS:** macOS 13 or later
  - **Linux:** Ubuntu 22.04+ or equivalent
- The Go backend binary (`arm-emulator`, built with `make build` at the repository root) for running from source

## Build

```bash
# Restore dependencies
dotnet restore

# Build the project
dotnet build

# Build in Release mode
dotnet build -c Release
```

## Run

```bash
# Run from source (Debug)
dotnet run --project ARMEmulator

# Run the built binary (Release)
dotnet run --project ARMEmulator -c Release
```

## Test

```bash
# Run all tests
dotnet test

# Run tests with coverage
dotnet test --collect:"XPlat Code Coverage"

# Run specific test project
dotnet test ARMEmulator.Tests
```

## Project Structure

```
avalonia-gui/
├── ARMEmulator/              # Main application
│   ├── Models/               # Data models (VMState, RegisterState, etc.)
│   ├── Services/             # Backend communication (ApiClient, WebSocketClient)
│   ├── ViewModels/           # MVVM ViewModels with ReactiveUI
│   ├── Views/                # Avalonia UI views (.axaml files)
│   ├── Controls/             # Custom controls (CodeEditor, HexDump, etc.)
│   ├── Converters/           # Value converters for data binding
│   ├── Themes/               # Custom theme resources
│   └── Assets/               # Icons and other static assets
├── ARMEmulator.Tests/        # Unit and integration tests
│   ├── Models/               # Model tests
│   ├── Services/             # Service tests (with mocks)
│   ├── ViewModels/           # ViewModel tests
│   ├── Ui/                   # Headless UI tests and screenshot baselines
│   └── Integration/          # End-to-end tests against a running backend
├── Directory.Build.props     # Shared build configuration
└── README.md                 # This file
```

## Technology Stack

| Component | Technology |
|-----------|------------|
| **Runtime** | .NET 10 with C# 13 |
| **UI Framework** | Avalonia UI 12.1.x |
| **MVVM** | ReactiveUI 26.x |
| **Text Editor** | AvaloniaEdit 12.x |
| **Testing** | xUnit v3, NSubstitute, AwesomeAssertions, Avalonia.Headless |

## Architecture

- **MVVM Pattern:** Strict separation of concerns with ReactiveUI
- **Functional Patterns:** Immutable data models using records, collection expressions
- **Modern C# 13:** Primary constructors, pattern matching, file-scoped namespaces
- **Exception-Based Error Handling:** Idiomatic .NET exceptions (no Result/Either monads)
- **Backend Communication:** REST API + WebSocket for real-time updates

## Backend Integration

This GUI connects to the ARM Emulator Go backend via:
- **REST API:** Port 8080 (configurable in preferences)
- **WebSocket:** Real-time state updates and console output

On start the GUI uses a healthy backend already listening on the configured port. Otherwise it starts the bundled `arm-emulator` with `-api-server -port N` and stops it on exit. The Restart Backend menu item, and the Retry button on the connection screen, start it again.

### Backend location

| Platform | Searched, in order |
|----------|--------------------|
| macOS app bundle | `Contents/Resources/arm-emulator`, next to the executable, its parent directory |
| Windows | `arm-emulator.exe` next to the executable, then its parent directory |
| Linux | next to the executable, `/usr/local/bin/arm-emulator`, `/usr/share/arm-emulator/arm-emulator` |

### Settings

Preferences, recent files, the selected inspector panel and the window geometry are stored as JSON in `ARMEmulator/settings.json` under the user's application data directory. An unreadable file is replaced by defaults and kept as `settings.json.bak`.

### Command line

`ARMEmulator prog.s` opens and loads the first argument that ends in `.s`.

## Tests

- Unit and headless UI tests run with `dotnet test`. Integration tests need a backend and are skipped otherwise; set `ARM_EMULATOR_URL` to point them at one.
- Screenshot tests compare against reviewed baselines in `ARMEmulator.Tests/Ui/Baselines/<os>/`. Run with `ARM_UPDATE_BASELINES=1` to refresh them, `ARM_SCREENSHOT_DIR=<dir>` to write review copies, and `ARM_VERIFY_SCREENSHOTS=1` to fail when a platform has no baselines.

See [KEYBOARD_SHORTCUTS.md](KEYBOARD_SHORTCUTS.md) for the key bindings.

## Development

### Code Style

- Follow `.editorconfig` settings for consistent formatting
- Use modern C# 13 features (primary constructors, collection expressions, pattern matching)
- Prefer immutable data models (records with `with` expressions)
- Use nullable reference types (`T?`) for optional values
- Follow TDD practices (red-green-refactor)

### Adding Dependencies

```bash
# Add package to main project
dotnet add ARMEmulator package PackageName

# Add package to test project
dotnet add ARMEmulator.Tests package PackageName
```

### Platform-Specific Builds

```bash
# Windows x64
dotnet publish -c Release -r win-x64 --self-contained

# macOS ARM64 (Apple Silicon)
dotnet publish -c Release -r osx-arm64 --self-contained

# macOS x64 (Intel)
dotnet publish -c Release -r osx-x64 --self-contained

# Linux x64
dotnet publish -c Release -r linux-x64 --self-contained
```

## License

See main project LICENSE file.
