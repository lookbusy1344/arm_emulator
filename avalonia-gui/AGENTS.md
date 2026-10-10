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

- Applies to **all files** including `.cs`, `.csproj`, `.md`, `.json`, `.xml`
- Configure your editor to use LF for this project
- Git is configured to enforce this via `.gitattributes`
- `.editorconfig` enforces `end_of_line = lf` for all file types

## Code Style

### Modern C# 13 Features (Use These)

- **Primary constructors** for dependency injection
- **Collection expressions** - modern initialization syntax:
  - Empty: `[]` instead of `new List<T>()` or `Array.Empty<T>()`
  - With items: `[item1, item2]` instead of `new List<T> { item1, item2 }`
  - Spread: `[..existingCollection, newItem]`
  - Works with arrays, lists, immutable collections, and any collection type
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

### Naming Conventions

- **Public properties/methods:** PascalCase
- **Private fields:** camelCase (no underscore prefix)
- **Local variables:** camelCase
- **Constants:** PascalCase
- **Interfaces:** IPrefixed

### Code Organization

- **ViewModels:** One per view, inherit from `ViewModelBase`
- **Services:** Stateless, injected via constructor
- **Models:** Immutable records when possible
- **Views:** XAML with code-behind minimal (logic in ViewModel)

### Analyzer Suppressions

**Never suppress JSV01** (RecordValueAnalyser: record member without value semantics). Fix the member type instead. For collections in records, use `EquatableArray<T>` or `EquatableDictionary<TKey, TValue>` from `ARMEmulator/Collections/`. They compare by contents and serialise as plain JSON arrays and objects. Register their element types in `ApiJsonContext` when they appear in wire records.

**Use inline suppressions, not central suppression files.**

- Suppress warnings at the specific location using `#pragma warning disable` or `[SuppressMessage]`
- Include a justification comment explaining why the suppression is necessary
- Avoid `GlobalSuppressions.cs` or `.editorconfig` suppressions unless truly project-wide

```csharp
// Justification: WebSocket library requires async void event handler
#pragma warning disable VSTHRD100
private async void OnWebSocketMessage(object? sender, MessageEventArgs e)
#pragma warning restore VSTHRD100
{
    // ...
}
```

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
- Give every test that touches sockets, processes or timers `[Fact(Timeout = ...)]` of a few seconds, and pass `TestContext.Current.CancellationToken` to every await. A hang then fails in seconds with the test's name.

### Running Tests

Do not make the user wait on long, silent runs. A filtered test run takes under a few seconds once built; a run that takes much longer is hung, not slow.

- **Build and test separately.** Build first (`dotnet build -v q`); compile errors surface in seconds and do not hide behind a test timeout. Then run `dotnet test --no-build`.
- **Run targeted tests while iterating:** `dotnet test --no-build --filter "FullyQualifiedName~ClassName"`. Run the full suite once, before commit.
- **Set tight timeouts:** wrap commands in `gtimeout` sized to the expected time plus a margin (about 60 s for a filtered run, 180 s for the full suite). Never set a ten-minute ceiling.
- **Log full output to a file**, then read its tail. Do not pipe test output through `grep`; a filter that matches nothing hides both hangs and failures.
- **After a timeout, find the hung test** (run the suspect classes alone) before rerunning anything. Never rerun the same command unchanged.

## Common Pitfalls

- ❌ Don't forget to dispose subscriptions
- ❌ Don't put logic in code-behind (use ViewModel)
- ❌ Don't use CRLF line endings
- ❌ Don't commit without running `dotnet format && dotnet build && dotnet test`
- ❌ Don't make breaking API changes
- ❌ Don't use `GlobalSuppressions.cs` - use inline suppressions with justifications

## Additional Documentation

- **Plan and status:** `../docs/plans/2026-10-06-avalonia-gui-plan.md`
- **API Reference:** See Go backend `api/` directory
- **Main Project Docs:** `../AGENTS.md`
