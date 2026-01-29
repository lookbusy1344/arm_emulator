# ARM Emulator - Avalonia .NET GUI

Cross-platform desktop GUI for the ARM Emulator built with Avalonia UI and .NET 10.

## Prerequisites

- .NET 10 SDK
- Platform-specific requirements:
  - **Windows:** Windows 10 or later
  - **macOS:** macOS 13 or later
  - **Linux:** Ubuntu 22.04+ or equivalent

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
│   ├── Views/                # Headless UI tests
│   ├── Integration/          # End-to-end integration tests
│   └── Mocks/                # Mock implementations
├── Directory.Build.props     # Shared build configuration
└── README.md                 # This file
```

## Technology Stack

| Component | Technology |
|-----------|------------|
| **Runtime** | .NET 10 with C# 13 |
| **UI Framework** | Avalonia UI 11.3.x |
| **MVVM** | ReactiveUI 20.x with source generators |
| **Text Editor** | AvaloniaEdit 0.10.x |
| **Testing** | xUnit, NSubstitute, FluentAssertions, Avalonia.Headless |

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

The backend must be running before launching the GUI. See the main project README for backend setup.

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
