# Avalonia GUI - Development Guidelines

Cross-platform ARM emulator GUI built with Avalonia and .NET 10.

## Prerequisites

- .NET SDK 10.0+ ([https://dot.net](https://dot.net))
- macOS 26.2 / Windows / Linux

## Architecture

- **Pattern:** MVVM with ReactiveUI
- **Backend Connection:** HTTP REST API (port 8080) + WebSocket
- **Language:** C# 13 with modern idioms

## Development Workflow

### Before Every Commit (MANDATORY)

```bash
# 1. Format code
dotnet format

# 2. Build (must succeed)
dotnet build

# 3. Run tests (must pass)
dotnet test
```

**All three must complete successfully before committing.**

### Build & Run

```bash
# Build
dotnet build

# Run application
dotnet run --project ARMEmulator

# Run tests
dotnet test
```

### Line Endings

**CRITICAL:** Always use LF (Unix) line endings, never CRLF.

- Configure your editor to use LF for this project
- Git is configured to enforce this via `.gitattributes`

## Code Style

### Modern C# 13 Features (Use These)

- **Primary constructors** for dependency injection
- **Collection expressions** `[item1, item2]` instead of `new List<T> { ... }`
- **Records** for immutable data models
- **Pattern matching** (switch expressions, property patterns)
- **Immutable collections** (`ImmutableArray<T>`, `ImmutableList<T>`)
- **File-scoped namespaces** (`namespace Foo;` not `namespace Foo { }`)
- **Nullable reference types** (enabled, treat warnings as errors)
- **Target-typed new** (`Thing x = new();`)

### Reactive Extensions (Rx)

Use ReactiveUI patterns:
- `ObservableAsPropertyHelper<T>` for derived properties
- `ReactiveCommand` for commands
- `WhenAnyValue()` for property change subscriptions
- Proper disposal of subscriptions

### Error Handling

Use **idiomatic .NET exception-based error handling** (not Result/Either monads):

- **Let exceptions propagate** - don't catch just to log and rethrow
- **Catch only at boundaries** where you can meaningfully handle errors (ViewModels for UI feedback)
- **Use domain-specific exceptions** (`ApiException`, `SessionNotFoundException`, etc.)
- **Don't catch-and-wrap** without adding useful context
- **Avoid anti-patterns:**
  - ❌ `catch (Exception) { return null; }` - hides failures
  - ❌ `catch (Exception ex) { Log(ex); throw; }` - noise
  - ❌ Pokemon exception handling (catch 'em all) at low levels

**See:** `../docs/AVALONIA_IMPLEMENTATION_PLAN.md` section "Exception Handling Philosophy" for detailed examples

### Naming Conventions

- **Public properties/methods:** PascalCase
- **Private fields:** _camelCase with underscore
- **Local variables:** camelCase
- **Constants:** PascalCase
- **Interfaces:** IPrefixed

### Code Organization

- **ViewModels:** One per view, inherit from `ViewModelBase`
- **Services:** Stateless, injected via constructor
- **Models:** Immutable records when possible
- **Views:** XAML with code-behind minimal (logic in ViewModel)

## Backend API Integration

**⚠️ CRITICAL:** Backend API is shared with Swift GUI.

- **NO breaking changes** to API contracts
- Only additive changes allowed
- Coordinate any API modifications with Swift GUI team
- Test against running Go backend (`make build && ./arm-emulator`)

### API Communication

- **REST API:** Synchronous operations (load program, step, reset)
- **WebSocket:** Real-time updates (execution state, register changes)
- **Base URL:** `http://localhost:8080`

## Testing

- Write tests for ViewModels and Services
- Mock backend services for unit tests
- Integration tests should use real WebSocket/HTTP (with backend running)
- Aim for high coverage of business logic

## Common Pitfalls

- ❌ Don't forget to dispose subscriptions
- ❌ Don't put logic in code-behind (use ViewModel)
- ❌ Don't use CRLF line endings
- ❌ Don't commit without running `dotnet format && dotnet build && dotnet test`
- ❌ Don't make breaking API changes

## Additional Documentation

- **Implementation Plan:** `../docs/AVALONIA_IMPLEMENTATION_PLAN.md`
- **API Reference:** See Go backend `api/` directory
- **Main Project Docs:** `../CLAUDE.md`
