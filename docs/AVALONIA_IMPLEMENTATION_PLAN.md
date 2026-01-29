# Avalonia .NET Frontend Implementation Plan

This document outlines a phased implementation plan for creating an Avalonia .NET frontend for the ARM Emulator that matches the existing Swift macOS app capabilities.

## Executive Summary

| Aspect | Details |
|--------|---------|
| **Target Framework** | .NET 10 with C# 13 |
| **UI Framework** | Avalonia UI 11.3.x |
| **Platforms** | Windows (primary), macOS, Linux |
| **Architecture** | MVVM with ReactiveUI + functional patterns |
| **Backend Communication** | REST API + WebSocket (same as Swift app) |
| **Feature Parity Target** | 100% Swift GUI feature coverage |

## Critical Constraints

### Test-Driven Development (TDD)

**All development must follow strict TDD practices with the red-green-refactor cycle.**

| Phase | Action |
|-------|--------|
| **Red** | Write a failing test first that defines the expected behavior |
| **Green** | Write the minimum code necessary to make the test pass |
| **Refactor** | Clean up the code while keeping tests green |

**TDD Rules:**
- **Never write implementation code without a failing test first**
- **Tests define the specification** - they document expected behavior
- **Do NOT delete or simplify tests because they fail** - fix the implementation instead
- **Run tests frequently** during development
- **All new code must have test coverage** (unit tests in `ARMEmulator.Tests/`)

**What to Test:**

| Layer | Test Type | Tools |
|-------|-----------|-------|
| **Models** | Unit tests for serialization, equality, edge cases | xUnit, FluentAssertions |
| **Services** | Unit tests with mocked HTTP/WebSocket | xUnit, NSubstitute |
| **ViewModels** | Unit tests for state transitions, commands | xUnit, ReactiveUI.Testing |
| **Views** | Headless UI tests for rendering/interaction | Avalonia.Headless.XUnit |
| **Integration** | End-to-end with real backend | xUnit, running backend |

**Example TDD Workflow:**

```csharp
// 1. RED: Write failing test first
[Fact]
public void RegisterState_Diff_ReturnsChangedRegisters()
{
    var before = RegisterState.Create(r0: 0, r1: 100);
    var after = RegisterState.Create(r0: 42, r1: 100);  // R0 changed

    var diff = after.Diff(before);

    diff.Should().BeEquivalentTo(["R0"]);
}

// 2. GREEN: Implement minimum code to pass
public ImmutableHashSet<string> Diff(RegisterState other) =>
    Enumerable.Range(0, 16)
        .Where(i => Registers[i] != other.Registers[i])
        .Select(i => $"R{i}")
        .ToImmutableHashSet();

// 3. REFACTOR: Clean up while tests stay green
```

**Test Organization:**

```
ARMEmulator.Tests/
├── Models/
│   ├── RegisterStateTests.cs
│   ├── VMStatusTests.cs
│   └── WatchpointTests.cs
├── Services/
│   ├── ApiClientTests.cs
│   ├── WebSocketClientTests.cs
│   └── BackendManagerTests.cs
├── ViewModels/
│   ├── MainWindowViewModelTests.cs
│   └── ExecutionCommandTests.cs
├── Views/
│   └── HeadlessUITests.cs
├── Integration/
│   └── FullExecutionCycleTests.cs
└── Mocks/
    ├── MockApiClient.cs
    └── MockWebSocketClient.cs
```

---

### Backend API Stability

**The Go backend API is shared between the Swift GUI and this Avalonia app. The API must remain stable.**

| Constraint | Rationale |
|------------|-----------|
| **No breaking changes to existing endpoints** | Swift GUI depends on current API contract |
| **Additive changes only** | New endpoints/fields are safe; removing or renaming breaks Swift |
| **Version the API if breaking changes needed** | Use `/api/v2/` prefix for incompatible changes |
| **Document any API additions** | Update `docs/` so both frontends stay synchronized |
| **Test both frontends after backend changes** | CI should validate Swift GUI still works |

**If you need backend changes:**
1. Check if the change is additive (safe) or breaking (dangerous)
2. For breaking changes, discuss with team before proceeding
3. Consider adding new endpoints alongside old ones during transition
4. Update both `AVALONIA_IMPLEMENTATION_PLAN.md` and Swift GUI documentation

**API Contract Location:** The definitive API specification lives in `api/` Go source files. The Swift GUI's `APIClient.swift` serves as the reference implementation.

---

## Design Principles

### Modern C# Style (C# 13)

- **Expression-bodied members** for concise implementations
- **Primary constructors** for records and classes
- **Collection expressions** (`[1, 2, 3]` syntax)
- **Pattern matching** exhaustively in switch expressions
- **`required` properties** for initialization safety
- **`file`-scoped types** for internal implementation details
- **Raw string literals** for multi-line strings
- **`init`-only setters** for immutable-after-construction
- **`params` collections** (new in C# 13) for flexible method signatures

### Functional Programming Patterns

- **Immutable data models** using records with `with` expressions
- **Pure functions** where possible (no side effects)
- **Pipeline-style composition** using LINQ and extension methods
- **Nullable reference types** (`T?`) with exhaustive handling for optional values
- **Exceptions** for error handling (idiomatic .NET) with domain-specific exception types
- **Higher-order functions** for callbacks and event handling
- **Discriminated unions** via abstract records with sealed derived types (for domain events, not errors)

### Code Organization

- **Vertical slice architecture** within feature folders
- **Extension methods** for cross-cutting concerns
- **Static factory methods** over constructors for complex creation
- **Prefer composition** over inheritance

### Exception Handling Philosophy

Exceptions in .NET are designed to propagate up the call stack until caught by code that can meaningfully handle them. Follow these principles:

**Let exceptions propagate freely:**
```csharp
// GOOD: Let the exception bubble up naturally
public async Task<RegisterState> StepAsync(string sessionId, CancellationToken ct)
{
    var response = await _http.PostAsync($"/api/v1/session/{sessionId}/step", null, ct);
    return await ParseResponseOrThrowAsync<RegisterState>(response, ct, sessionId);
}

// BAD: Catching just to log and rethrow adds noise
public async Task<RegisterState> StepAsync(string sessionId, CancellationToken ct)
{
    try
    {
        var response = await _http.PostAsync($"/api/v1/session/{sessionId}/step", null, ct);
        return await ParseResponseOrThrowAsync<RegisterState>(response, ct, sessionId);
    }
    catch (Exception ex)
    {
        _logger.LogError(ex, "Step failed");  // Noise - the UI layer will log this
        throw;  // Pointless rethrow
    }
}
```

**Catch only at boundaries where you can act:**
```csharp
// ViewModel: The UI boundary where we display errors to users
public async Task ExecuteStepCommand()
{
    try
    {
        var registers = await _api.StepAsync(SessionId, _cts.Token);
        UpdateRegisters(registers);
    }
    catch (SessionNotFoundException)
    {
        await ReconnectSessionAsync();  // Meaningful recovery action
    }
    catch (ApiException ex)
    {
        ErrorMessage = ex.Message;  // Display to user - this is the boundary
    }
}
```

**Don't catch-and-wrap without adding information:**
```csharp
// BAD: Wrapping adds nothing useful
catch (HttpRequestException ex)
{
    throw new ApiException("HTTP error occurred", ex);  // "HTTP error" is less useful than the original
}

// GOOD: Wrap only when adding domain context
catch (HttpRequestException ex)
{
    throw new BackendUnavailableException(
        $"Cannot reach backend at {_baseUrl} - is the emulator running?", ex);
}
```

**Where to catch exceptions:**

| Layer | Catch? | Action |
|-------|--------|--------|
| **ApiClient** | Selective | Only to translate HTTP errors into domain exceptions |
| **Services** | Rarely | Only for retry logic or resource cleanup |
| **ViewModels** | Yes | Display errors, trigger recovery, update UI state |
| **Views** | No | Let ViewModels handle it |

**Anti-patterns to avoid:**
- `catch (Exception) { return null; }` — Hides failures, causes NullReferenceException elsewhere
- `catch (Exception ex) { Log(ex); throw; }` — Noise; the eventual handler will log it
- `catch (Exception) { throw new Exception("Error"); }` — Destroys stack trace and original message
- Pokemon exception handling (`catch 'em all`) at low levels — Masks bugs

---

## Phase 0: Project Setup & Infrastructure

**Duration:** Foundation work
**Goal:** Establish project structure, tooling, and build configuration

### 0.1 Project Structure

```
avalonia-gui/
├── ARMEmulator.sln
├── ARMEmulator/
│   ├── App.axaml
│   ├── App.axaml.cs
│   ├── Program.cs
│   ├── Models/
│   │   ├── VMState.cs
│   │   ├── VMStatus.cs
│   │   ├── RegisterState.cs
│   │   ├── CPSRFlags.cs
│   │   ├── Watchpoint.cs
│   │   ├── DisassemblyInstruction.cs
│   │   ├── SourceMapEntry.cs
│   │   ├── SessionInfo.cs
│   │   └── AppSettings.cs
│   ├── Services/
│   │   ├── IApiClient.cs
│   │   ├── ApiClient.cs
│   │   ├── IWebSocketClient.cs
│   │   ├── WebSocketClient.cs
│   │   ├── IBackendManager.cs
│   │   ├── BackendManager.cs
│   │   ├── IFileService.cs
│   │   └── FileService.cs
│   ├── ViewModels/
│   │   ├── MainWindowViewModel.cs
│   │   ├── EditorViewModel.cs
│   │   ├── RegistersViewModel.cs
│   │   ├── MemoryViewModel.cs
│   │   ├── StackViewModel.cs
│   │   ├── DisassemblyViewModel.cs
│   │   ├── ConsoleViewModel.cs
│   │   ├── WatchpointsViewModel.cs
│   │   ├── BreakpointsViewModel.cs
│   │   ├── ExpressionEvaluatorViewModel.cs
│   │   └── ExamplesBrowserViewModel.cs
│   ├── Views/
│   │   ├── MainWindow.axaml(.cs)
│   │   ├── EditorView.axaml(.cs)
│   │   ├── RegistersView.axaml(.cs)
│   │   ├── MemoryView.axaml(.cs)
│   │   ├── StackView.axaml(.cs)
│   │   ├── DisassemblyView.axaml(.cs)
│   │   ├── ConsoleView.axaml(.cs)
│   │   ├── WatchpointsView.axaml(.cs)
│   │   ├── BreakpointsView.axaml(.cs)
│   │   ├── ExpressionEvaluatorView.axaml(.cs)
│   │   ├── ExamplesBrowserWindow.axaml(.cs)
│   │   ├── PreferencesWindow.axaml(.cs)
│   │   └── AboutWindow.axaml(.cs)
│   ├── Controls/
│   │   ├── CodeEditor.axaml(.cs)
│   │   ├── HexDump.axaml(.cs)
│   │   └── StatusIndicator.axaml(.cs)
│   ├── Converters/
│   │   ├── HexValueConverter.cs
│   │   ├── StateToColorConverter.cs
│   │   └── BoolToVisibilityConverter.cs
│   ├── Themes/
│   │   ├── Light.axaml
│   │   └── Dark.axaml
│   └── Assets/
│       └── Icons/
├── ARMEmulator.Tests/
│   ├── ViewModels/
│   ├── Services/
│   └── Mocks/
└── Directory.Build.props
```

### 0.2 Dependencies

```xml
<!-- Core Avalonia 11.3 -->
<PackageReference Include="Avalonia" Version="11.3.*" />
<PackageReference Include="Avalonia.Desktop" Version="11.3.*" />
<PackageReference Include="Avalonia.Themes.Fluent" Version="11.3.*" />
<PackageReference Include="Avalonia.Fonts.Inter" Version="11.3.*" />

<!-- MVVM - ReactiveUI with source generators (no Fody needed in .NET 10) -->
<PackageReference Include="ReactiveUI" Version="20.*" />
<PackageReference Include="ReactiveUI.SourceGenerators" Version="20.*" />
<PackageReference Include="Avalonia.ReactiveUI" Version="11.3.*" />

<!-- Syntax Highlighting -->
<PackageReference Include="AvaloniaEdit" Version="11.*" />

<!-- HTTP & WebSocket (built into .NET 10, no extra packages needed) -->
<PackageReference Include="System.Reactive" Version="6.*" />

<!-- DI & Configuration -->
<PackageReference Include="Microsoft.Extensions.DependencyInjection" Version="10.*" />
<PackageReference Include="Microsoft.Extensions.Configuration.Json" Version="10.*" />

<!-- Functional helpers (optional - for advanced patterns like Seq, Option beyond T?) -->
<!-- <PackageReference Include="LanguageExt.Core" Version="5.*" /> -->

<!-- Testing -->
<PackageReference Include="xunit" Version="2.*" />
<PackageReference Include="NSubstitute" Version="5.*" />
<PackageReference Include="FluentAssertions" Version="7.*" />
<PackageReference Include="Avalonia.Headless.XUnit" Version="11.3.*" />
```

> **Note:** We use `NSubstitute` instead of `Moq` for a more functional mocking syntax. Error handling uses idiomatic .NET exceptions with domain-specific exception types rather than Result/Either monads.

### 0.3 Build Configuration

Create `Directory.Build.props`:
```xml
<Project>
  <PropertyGroup>
    <TargetFramework>net10.0</TargetFramework>
    <LangVersion>13</LangVersion>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
    <EnforceCodeStyleInBuild>true</EnforceCodeStyleInBuild>
    <AnalysisLevel>latest-recommended</AnalysisLevel>

    <!-- Enable AOT-friendly patterns -->
    <EnableTrimAnalyzer>true</EnableTrimAnalyzer>
    <JsonSerializerIsReflectionEnabledByDefault>false</JsonSerializerIsReflectionEnabledByDefault>
  </PropertyGroup>

  <!-- Global usings for functional style -->
  <ItemGroup>
    <Using Include="System.Collections.Immutable" />
    <Using Include="System.Reactive.Linq" />
  </ItemGroup>
</Project>
```

Create `.editorconfig` for consistent style:
```ini
[*.cs]
# Prefer expression bodies
csharp_style_expression_bodied_methods = when_on_single_line:suggestion
csharp_style_expression_bodied_properties = true:suggestion
csharp_style_expression_bodied_lambdas = true:suggestion

# Prefer pattern matching
csharp_style_pattern_matching_over_is_with_cast_check = true:suggestion
csharp_style_pattern_matching_over_as_with_null_check = true:suggestion
csharp_style_prefer_switch_expression = true:suggestion
csharp_style_prefer_pattern_matching = true:suggestion

# Prefer modern syntax
csharp_style_prefer_primary_constructors = true:suggestion
csharp_style_namespace_declarations = file_scoped:warning
csharp_style_prefer_collection_expression = true:suggestion

# Prefer immutability
dotnet_style_readonly_field = true:warning
```

Platform-specific targets in `.csproj`:
```xml
<PropertyGroup Condition="'$(RuntimeIdentifier)' == 'win-x64'">
  <PublishSingleFile>true</PublishSingleFile>
  <SelfContained>true</SelfContained>
  <PublishAot>false</PublishAot>
</PropertyGroup>
<PropertyGroup Condition="'$(RuntimeIdentifier)' == 'osx-x64' Or '$(RuntimeIdentifier)' == 'osx-arm64'">
  <PublishSingleFile>false</PublishSingleFile>
  <CreatePackage>true</CreatePackage>
</PropertyGroup>
```

### 0.4 Tasks

- [ ] Create solution and project files
- [ ] Configure Avalonia with Fluent theme
- [ ] Set up ReactiveUI infrastructure
- [ ] Configure multi-platform builds (win-x64, osx-arm64, osx-x64, linux-x64)
- [ ] Create basic App.axaml with theme switching support
- [ ] Set up unit test project with mocking infrastructure
- [ ] Create README.md with build instructions
- [ ] Add EditorConfig for consistent code style

---

## Phase 1: Data Models & Services

**Goal:** Implement all data models and backend communication

### 1.1 Data Models

Map Swift models to C# equivalents:

| Swift Model | C# Model | Notes |
|-------------|----------|-------|
| `VMState` enum | `VMState` enum | idle, running, breakpoint, halted, error, waitingForInput |
| `VMStatus` | `VMStatus` record | State, PC, cycles, error, memory write info |
| `RegisterState` | `RegisterState` record | R0-R15, SP, LR, PC, CPSR |
| `CPSRFlags` | `CPSRFlags` record | N, Z, C, V flags |
| `Watchpoint` | `Watchpoint` record | ID, address, type |
| `DisassemblyInstruction` | `DisassemblyInstruction` record | Address, machine code, mnemonic, symbol |
| `SourceMapEntry` | `SourceMapEntry` record | Address, line number, source text |
| `AppSettings` | `AppSettings` class | Backend URL, font size, theme, recent files |

Example models using modern C# 13 patterns:

```csharp
namespace ARMEmulator.Models;

// Enum with extension methods for behavior
public enum VMState
{
    Idle,
    Running,
    Breakpoint,
    Halted,
    Error,
    WaitingForInput
}

public static class VMStateExtensions
{
    public static bool IsEditorEditable(this VMState state) =>
        state is VMState.Idle or VMState.Halted or VMState.Error;

    public static bool CanStep(this VMState state) =>
        state is VMState.Idle or VMState.Breakpoint;

    public static bool CanPause(this VMState state) =>
        state is VMState.Running or VMState.WaitingForInput;
}

// Immutable record with optional memory write info
public sealed record VMStatus(
    VMState State,
    uint PC,
    ulong Cycles,
    string? Error = null,
    MemoryWrite? LastWrite = null
);

// Nested record for memory write tracking
public sealed record MemoryWrite(uint Address, uint Size);

// Use a struct for small, frequently-copied value types
public readonly record struct CPSRFlags(bool N, bool Z, bool C, bool V)
{
    public string DisplayString =>
        $"{(N ? 'N' : '-')}{(Z ? 'Z' : '-')}{(C ? 'C' : '-')}{(V ? 'V' : '-')}";

    // Factory for parsing from JSON
    public static CPSRFlags FromJson(JsonElement json) => new(
        N: json.GetProperty("n").GetBoolean(),
        Z: json.GetProperty("z").GetBoolean(),
        C: json.GetProperty("c").GetBoolean(),
        V: json.GetProperty("v").GetBoolean()
    );
}

// Use ImmutableArray for register storage (efficient + immutable)
public sealed record RegisterState
{
    public required ImmutableArray<uint> Registers { get; init; }  // R0-R15
    public required CPSRFlags CPSR { get; init; }

    // Named accessors using collection expression indexing
    public uint R0 => Registers[0];
    public uint R1 => Registers[1];
    public uint R2 => Registers[2];
    public uint R3 => Registers[3];
    public uint R4 => Registers[4];
    public uint R5 => Registers[5];
    public uint R6 => Registers[6];
    public uint R7 => Registers[7];
    public uint R8 => Registers[8];
    public uint R9 => Registers[9];
    public uint R10 => Registers[10];
    public uint R11 => Registers[11];
    public uint R12 => Registers[12];
    public uint SP => Registers[13];
    public uint LR => Registers[14];
    public uint PC => Registers[15];

    // Get register by name (functional lookup)
    public uint this[string name] => name.ToUpperInvariant() switch
    {
        "R0" => R0, "R1" => R1, "R2" => R2, "R3" => R3,
        "R4" => R4, "R5" => R5, "R6" => R6, "R7" => R7,
        "R8" => R8, "R9" => R9, "R10" => R10, "R11" => R11,
        "R12" => R12, "SP" or "R13" => SP, "LR" or "R14" => LR,
        "PC" or "R15" => PC,
        _ => throw new ArgumentException($"Unknown register: {name}")
    };

    // Factory method using collection expression
    public static RegisterState Create(
        uint r0 = 0, uint r1 = 0, uint r2 = 0, uint r3 = 0,
        uint r4 = 0, uint r5 = 0, uint r6 = 0, uint r7 = 0,
        uint r8 = 0, uint r9 = 0, uint r10 = 0, uint r11 = 0,
        uint r12 = 0, uint sp = 0, uint lr = 0, uint pc = 0,
        CPSRFlags? cpsr = null
    ) => new()
    {
        Registers = [r0, r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12, sp, lr, pc],
        CPSR = cpsr ?? default
    };

    // Diff with another state (returns changed register names)
    public ImmutableHashSet<string> Diff(RegisterState other)
    {
        var names = new[] { "R0", "R1", "R2", "R3", "R4", "R5", "R6", "R7",
                           "R8", "R9", "R10", "R11", "R12", "SP", "LR", "PC" };
        return names
            .Where((name, i) => Registers[i] != other.Registers[i])
            .Concat(CPSR != other.CPSR ? ["CPSR"] : [])
            .ToImmutableHashSet();
    }
}

// Discriminated union for emulator events using abstract record
public abstract record EmulatorEvent(string SessionId);

public sealed record StateEvent(
    string SessionId,
    VMStatus Status,
    RegisterState Registers
) : EmulatorEvent(SessionId);

public sealed record OutputEvent(
    string SessionId,
    OutputStream Stream,
    string Content
) : EmulatorEvent(SessionId);

public sealed record ExecutionEvent(
    string SessionId,
    ExecutionEventType EventType,
    uint? Address = null,
    string? Symbol = null,
    string? Message = null
) : EmulatorEvent(SessionId);

// Enums for type safety
public enum OutputStream { Stdout, Stderr }

public enum ExecutionEventType { BreakpointHit, Halted, Error }

// Watchpoint with type enum
public sealed record Watchpoint(int Id, uint Address, WatchpointType Type);

public enum WatchpointType { Read, Write, ReadWrite }
```

### 1.2 Domain Exceptions

Define domain-specific exceptions for clear error handling:

```csharp
namespace ARMEmulator.Services;

/// <summary>Base exception for all API-related errors.</summary>
public class ApiException : Exception
{
    public HttpStatusCode? StatusCode { get; }

    public ApiException(string message, HttpStatusCode? statusCode = null, Exception? inner = null)
        : base(message, inner) => StatusCode = statusCode;
}

/// <summary>Thrown when a session is not found or has expired.</summary>
public class SessionNotFoundException(string sessionId)
    : ApiException($"Session '{sessionId}' not found or expired", HttpStatusCode.NotFound);

/// <summary>Thrown when program loading fails due to parse/assembly errors.</summary>
public class ProgramLoadException : ApiException
{
    public IReadOnlyList<ParseError> Errors { get; }

    public ProgramLoadException(IReadOnlyList<ParseError> errors)
        : base($"Program failed to load: {errors.Count} error(s)") => Errors = errors;
}

/// <summary>Thrown when the backend is unreachable.</summary>
public class BackendUnavailableException(string message, Exception? inner = null)
    : ApiException(message, null, inner);

/// <summary>Thrown when an expression evaluation fails.</summary>
public class ExpressionEvaluationException(string expression, string error)
    : ApiException($"Failed to evaluate '{expression}': {error}");

/// <summary>Parse error details from the assembler.</summary>
public sealed record ParseError(int Line, int Column, string Message);
```

### 1.3 API Client

Implement `IApiClient` interface using idiomatic .NET exception-based error handling:

```csharp
namespace ARMEmulator.Services;

/// <summary>
/// Client for the ARM Emulator REST API.
/// All methods throw <see cref="ApiException"/> or derived types on failure.
/// </summary>
public interface IApiClient
{
    // Session Management
    /// <exception cref="BackendUnavailableException">Backend not reachable</exception>
    /// <exception cref="ApiException">Request failed</exception>
    Task<SessionInfo> CreateSessionAsync(CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<VMStatus> GetStatusAsync(string sessionId, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task DestroySessionAsync(string sessionId, CancellationToken ct = default);

    // Program Loading
    /// <exception cref="SessionNotFoundException">Session not found</exception>
    /// <exception cref="ProgramLoadException">Assembly/parse errors</exception>
    Task<LoadProgramResponse> LoadProgramAsync(string sessionId, string source, CancellationToken ct = default);

    // Execution Control
    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task RunAsync(string sessionId, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task StopAsync(string sessionId, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<RegisterState> StepAsync(string sessionId, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<RegisterState> StepOverAsync(string sessionId, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<RegisterState> StepOutAsync(string sessionId, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task ResetAsync(string sessionId, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task RestartAsync(string sessionId, CancellationToken ct = default);

    // State Inspection
    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<RegisterState> GetRegistersAsync(string sessionId, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<ImmutableArray<byte>> GetMemoryAsync(string sessionId, uint address, int length, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<ImmutableArray<DisassemblyInstruction>> GetDisassemblyAsync(string sessionId, uint address, int count, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<ImmutableArray<SourceMapEntry>> GetSourceMapAsync(string sessionId, CancellationToken ct = default);

    // Breakpoints
    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task AddBreakpointAsync(string sessionId, uint address, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task RemoveBreakpointAsync(string sessionId, uint address, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<ImmutableArray<uint>> GetBreakpointsAsync(string sessionId, CancellationToken ct = default);

    // Watchpoints
    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<Watchpoint> AddWatchpointAsync(string sessionId, uint address, WatchpointType type, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task RemoveWatchpointAsync(string sessionId, int watchpointId, CancellationToken ct = default);

    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task<ImmutableArray<Watchpoint>> GetWatchpointsAsync(string sessionId, CancellationToken ct = default);

    // Expression Evaluation
    /// <exception cref="SessionNotFoundException">Session not found</exception>
    /// <exception cref="ExpressionEvaluationException">Invalid expression</exception>
    Task<uint> EvaluateExpressionAsync(string sessionId, string expression, CancellationToken ct = default);

    // Input
    /// <exception cref="SessionNotFoundException">Session not found</exception>
    Task SendStdinAsync(string sessionId, string data, CancellationToken ct = default);

    // Version
    /// <exception cref="BackendUnavailableException">Backend not reachable</exception>
    Task<BackendVersion> GetVersionAsync(CancellationToken ct = default);

    // Examples
    Task<ImmutableArray<ExampleInfo>> GetExamplesAsync(CancellationToken ct = default);

    /// <exception cref="ApiException">Example not found</exception>
    Task<string> GetExampleContentAsync(string name, CancellationToken ct = default);
}

// Source-generated JSON context for AOT compatibility
[JsonSourceGenerationOptions(
    PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase,
    DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull
)]
[JsonSerializable(typeof(SessionInfo))]
[JsonSerializable(typeof(VMStatus))]
[JsonSerializable(typeof(RegisterState))]
[JsonSerializable(typeof(LoadProgramResponse))]
[JsonSerializable(typeof(Watchpoint))]
[JsonSerializable(typeof(BackendVersion))]
[JsonSerializable(typeof(ExampleInfo))]
[JsonSerializable(typeof(DisassemblyInstruction))]
[JsonSerializable(typeof(SourceMapEntry))]
[JsonSerializable(typeof(ApiErrorResponse))]
internal partial class ApiJsonContext : JsonSerializerContext { }

/// <summary>Standard error response from the backend.</summary>
internal sealed record ApiErrorResponse(string Error, IReadOnlyList<ParseError>? ParseErrors = null);

// Implementation using primary constructor
public sealed class ApiClient(HttpClient http, ILogger<ApiClient> logger) : IApiClient
{
    private readonly JsonSerializerOptions _jsonOptions = new()
    {
        TypeInfoResolver = ApiJsonContext.Default
    };

    public async Task<SessionInfo> CreateSessionAsync(CancellationToken ct = default)
    {
        try
        {
            var response = await http.PostAsync("/api/v1/session", null, ct);
            return await ParseResponseOrThrowAsync<SessionInfo>(response, ct);
        }
        catch (HttpRequestException ex)
        {
            logger.LogError(ex, "Failed to connect to backend");
            throw new BackendUnavailableException("Cannot connect to backend", ex);
        }
    }

    public async Task<LoadProgramResponse> LoadProgramAsync(string sessionId, string source, CancellationToken ct = default)
    {
        var content = new StringContent(source, Encoding.UTF8, "text/plain");
        var response = await http.PostAsync($"/api/v1/session/{sessionId}/load", content, ct);

        // Special handling for parse errors (400 with error details)
        if (response.StatusCode == HttpStatusCode.BadRequest)
        {
            var errorResponse = await response.Content.ReadFromJsonAsync<ApiErrorResponse>(_jsonOptions, ct);
            if (errorResponse?.ParseErrors is { Count: > 0 } errors)
                throw new ProgramLoadException(errors);
        }

        return await ParseResponseOrThrowAsync<LoadProgramResponse>(response, ct, sessionId);
    }

    public async Task<uint> EvaluateExpressionAsync(string sessionId, string expression, CancellationToken ct = default)
    {
        var response = await http.PostAsJsonAsync(
            $"/api/v1/session/{sessionId}/evaluate",
            new { expression },
            _jsonOptions,
            ct);

        if (response.StatusCode == HttpStatusCode.BadRequest)
        {
            var errorResponse = await response.Content.ReadFromJsonAsync<ApiErrorResponse>(_jsonOptions, ct);
            throw new ExpressionEvaluationException(expression, errorResponse?.Error ?? "Unknown error");
        }

        return await ParseResponseOrThrowAsync<uint>(response, ct, sessionId);
    }

    // Helper for consistent response parsing with appropriate exceptions
    private async Task<T> ParseResponseOrThrowAsync<T>(
        HttpResponseMessage response,
        CancellationToken ct,
        string? sessionId = null)
    {
        if (response.StatusCode == HttpStatusCode.NotFound && sessionId is not null)
            throw new SessionNotFoundException(sessionId);

        if (!response.IsSuccessStatusCode)
        {
            var error = await response.Content.ReadAsStringAsync(ct);
            throw new ApiException($"API error: {error}", response.StatusCode);
        }

        var content = await response.Content.ReadFromJsonAsync<T>(_jsonOptions, ct);
        return content ?? throw new ApiException("Response deserialized to null");
    }

    // ... other methods follow same pattern: call API, throw on failure
}
```

### 1.3 WebSocket Client

Implement event streaming with auto-reconnection:

```csharp
public interface IWebSocketClient : IDisposable
{
    IObservable<EmulatorEvent> Events { get; }
    Task ConnectAsync(string sessionId, CancellationToken ct = default);
    Task DisconnectAsync();
    bool IsConnected { get; }
}

public abstract record EmulatorEvent(string SessionId);
public record StateEvent(string SessionId, VMStatus Status, RegisterState Registers) : EmulatorEvent(SessionId);
public record OutputEvent(string SessionId, string Stream, string Content) : EmulatorEvent(SessionId);
public record ExecutionEvent(string SessionId, string EventType, uint? Address, string? Symbol, string? Message) : EmulatorEvent(SessionId);
```

Features:
- Auto-reconnection with exponential backoff (up to 5 retries)
- Session-specific event filtering
- Rx.NET observable for event streaming
- Thread-safe connection state management

### 1.4 Backend Manager

Platform-specific backend process management:

```csharp
public interface IBackendManager
{
    BackendStatus Status { get; }
    IObservable<BackendStatus> StatusChanged { get; }
    Task StartAsync(CancellationToken ct = default);
    Task StopAsync();
    Task<bool> HealthCheckAsync(CancellationToken ct = default);
}

public enum BackendStatus
{
    Unknown,
    Starting,
    Running,
    Stopped,
    Error
}
```

Platform considerations:
- **Windows:** Find `arm-emulator.exe` in app directory
- **macOS:** Find `arm-emulator` in `Contents/Resources` for .app bundle
- **Linux:** Find `arm-emulator` in app directory or `/usr/local/bin`

### 1.5 File Service

```csharp
public interface IFileService
{
    Task<string?> OpenFileAsync(Window parent);
    Task<string?> SaveFileAsync(Window parent, string content, string? currentPath);
    IReadOnlyList<RecentFile> RecentFiles { get; }
    void AddRecentFile(string path);
    void ClearRecentFiles();
    string? CurrentFilePath { get; }
}
```

### 1.6 Tasks (TDD Order)

**Models (test first, then implement):**
- [ ] Write tests for `RegisterState` (creation, diff, indexer, serialization)
- [ ] Implement `RegisterState` to pass tests
- [ ] Write tests for `VMStatus`, `Watchpoint`, `EmulatorEvent` records
- [ ] Implement remaining models to pass tests

**Services (test first with mocks, then implement):**
- [ ] Write `ApiClient` tests with mocked `HttpClient`
- [ ] Implement `ApiClient` to pass tests
- [ ] Write `WebSocketClient` tests with mocked WebSocket
- [ ] Implement `WebSocketClient` with Rx.NET event streaming
- [ ] Write `BackendManager` tests (process lifecycle)
- [ ] Implement `BackendManager` for all platforms
- [ ] Write `FileService` tests
- [ ] Implement `FileService` with recent files persistence

**Integration:**
- [ ] Write integration tests against running backend
- [ ] Verify all endpoints match Swift `APIClient.swift` behavior

---

## Phase 2: Core ViewModels

**Goal:** Implement main ViewModel with reactive state management

### 2.1 Main ViewModel

Central state hub matching Swift's `EmulatorViewModel`, using modern C# patterns:

```csharp
namespace ARMEmulator.ViewModels;

// Use partial class with source generators for [Reactive] properties
public partial class MainWindowViewModel : ReactiveObject, IDisposable
{
    private readonly IApiClient _api;
    private readonly IWebSocketClient _ws;
    private readonly CompositeDisposable _disposables = new();

    // Primary constructor with DI
    public MainWindowViewModel(IApiClient api, IWebSocketClient ws)
    {
        _api = api;
        _ws = ws;

        // Initialize commands using expression-bodied definitions
        RunCommand = CreateCommand(RunAsync, this.WhenAnyValue(x => x.Status).Select(s => !s.CanPause()));
        PauseCommand = CreateCommand(PauseAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanPause()));
        StepCommand = CreateCommand(StepAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanStep()));
        StepOverCommand = CreateCommand(StepOverAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanStep()));
        StepOutCommand = CreateCommand(StepOutAsync, this.WhenAnyValue(x => x.Status).Select(s => s.CanStep()));
        ResetCommand = CreateCommand(ResetAsync);
        LoadProgramCommand = CreateCommand(LoadProgramAsync);

        // Subscribe to WebSocket events functionally
        _ws.Events
            .ObserveOn(RxApp.MainThreadScheduler)
            .Subscribe(HandleEvent)
            .DisposeWith(_disposables);

        // Computed properties using WhenAnyValue
        _canPauseHelper = this.WhenAnyValue(x => x.Status)
            .Select(s => s.CanPause())
            .ToProperty(this, x => x.CanPause)
            .DisposeWith(_disposables);

        _canStepHelper = this.WhenAnyValue(x => x.Status)
            .Select(s => s.CanStep())
            .ToProperty(this, x => x.CanStep)
            .DisposeWith(_disposables);

        _isEditorEditableHelper = this.WhenAnyValue(x => x.Status)
            .Select(s => s.IsEditorEditable())
            .ToProperty(this, x => x.IsEditorEditable)
            .DisposeWith(_disposables);
    }

    // Reactive properties via source generators (no Fody!)
    [Reactive] public RegisterState Registers { get; set; } = RegisterState.Create();
    [Reactive] public RegisterState? PreviousRegisters { get; set; }
    [Reactive] public ImmutableHashSet<string> ChangedRegisters { get; set; } = [];
    [Reactive] public VMState Status { get; set; } = VMState.Idle;
    [Reactive] public string ConsoleOutput { get; set; } = "";
    [Reactive] public string? ErrorMessage { get; set; }

    // Debugging state (immutable collections)
    [Reactive] public ImmutableHashSet<uint> Breakpoints { get; set; } = [];
    [Reactive] public ImmutableArray<Watchpoint> Watchpoints { get; set; } = [];

    // Source mapping (immutable dictionaries)
    [Reactive] public string SourceCode { get; set; } = "";
    [Reactive] public ImmutableDictionary<uint, int> AddressToLine { get; set; } = ImmutableDictionary<uint, int>.Empty;
    [Reactive] public ImmutableDictionary<int, uint> LineToAddress { get; set; } = ImmutableDictionary<int, uint>.Empty;
    [Reactive] public ImmutableHashSet<int> ValidBreakpointLines { get; set; } = [];

    // Memory state
    [Reactive] public ImmutableArray<byte> MemoryData { get; set; } = [];
    [Reactive] public uint MemoryAddress { get; set; }
    [Reactive] public MemoryWrite? LastMemoryWrite { get; set; }

    // Disassembly
    [Reactive] public ImmutableArray<DisassemblyInstruction> Disassembly { get; set; } = [];

    // Connection state
    [Reactive] public bool IsConnected { get; set; }
    public string? SessionId { get; private set; }

    // Computed properties (ObservableAsPropertyHelper)
    private readonly ObservableAsPropertyHelper<bool> _canPauseHelper;
    private readonly ObservableAsPropertyHelper<bool> _canStepHelper;
    private readonly ObservableAsPropertyHelper<bool> _isEditorEditableHelper;

    public bool CanPause => _canPauseHelper.Value;
    public bool CanStep => _canStepHelper.Value;
    public bool IsEditorEditable => _isEditorEditableHelper.Value;

    // Commands
    public ReactiveCommand<Unit, Unit> RunCommand { get; }
    public ReactiveCommand<Unit, Unit> PauseCommand { get; }
    public ReactiveCommand<Unit, Unit> StepCommand { get; }
    public ReactiveCommand<Unit, Unit> StepOverCommand { get; }
    public ReactiveCommand<Unit, Unit> StepOutCommand { get; }
    public ReactiveCommand<Unit, Unit> ResetCommand { get; }
    public ReactiveCommand<Unit, Unit> LoadProgramCommand { get; }

    // Helper to create commands with consistent error handling
    private ReactiveCommand<Unit, Unit> CreateCommand(
        Func<CancellationToken, Task> execute,
        IObservable<bool>? canExecute = null
    ) => ReactiveCommand.CreateFromTask(
        execute,
        canExecute,
        outputScheduler: RxApp.MainThreadScheduler
    ).DisposeWith(_disposables);

    public void Dispose() => _disposables.Dispose();
}
```

### 2.2 Register Change Highlighting

Implement the same highlight system as Swift, using functional reactive patterns:

```csharp
// In MainWindowViewModel - use Rx for timed highlights (no manual CancellationTokenSource)
private readonly Subject<string> _registerHighlightTrigger = new();
private readonly TimeSpan HighlightDuration = TimeSpan.FromSeconds(1.5);

// In constructor, set up highlight pipeline
private void SetupHighlightPipeline()
{
    // Each register gets its own debounced removal stream
    _registerHighlightTrigger
        .GroupBy(register => register)
        .SelectMany(group =>
            group.Select(register => (register, action: "add"))
                .Merge(group
                    .Throttle(HighlightDuration)
                    .Select(register => (register, action: "remove"))
                )
        )
        .ObserveOn(RxApp.MainThreadScheduler)
        .Subscribe(x =>
        {
            ChangedRegisters = x.action == "add"
                ? ChangedRegisters.Add(x.register)
                : ChangedRegisters.Remove(x.register);
        })
        .DisposeWith(_disposables);
}

// Update registers and trigger highlights for changes
private void UpdateRegisters(RegisterState newRegisters)
{
    if (PreviousRegisters is not null)
    {
        // Functional diff - get changed registers and trigger highlights
        newRegisters.Diff(PreviousRegisters)
            .ForEach(register => _registerHighlightTrigger.OnNext(register));
    }

    PreviousRegisters = Registers;
    Registers = newRegisters;
}

// Extension method for ForEach on IEnumerable
public static class EnumerableExtensions
{
    public static void ForEach<T>(this IEnumerable<T> source, Action<T> action)
    {
        foreach (var item in source) action(item);
    }
}
```

### 2.3 WebSocket Event Handling

Using exhaustive pattern matching with switch expressions:

```csharp
// Handle all event types with exhaustive pattern matching
private void HandleEvent(EmulatorEvent evt)
{
    // Guard against stale events when already halted
    if (Status == VMState.Halted && evt is StateEvent) return;

    _ = evt switch
    {
        StateEvent { Status: var status, Registers: var regs } =>
            ApplyStateUpdate(status, regs),

        OutputEvent { Stream: var stream, Content: var content } =>
            AppendOutput(stream, content),

        ExecutionEvent { EventType: var type, Message: var msg } =>
            ApplyExecutionEvent(type, msg),

        _ => false // Unreachable with sealed record hierarchy
    };
}

// Pure-ish function that returns success for chaining
private bool ApplyStateUpdate(VMStatus status, RegisterState registers)
{
    UpdateRegisters(registers);
    Status = status.State;
    LastMemoryWrite = status.LastWrite;
    return true;
}

private bool AppendOutput(OutputStream stream, string content)
{
    // Could differentiate stdout/stderr styling here
    ConsoleOutput += content;
    return true;
}

// Exhaustive switch expression on enum
private bool ApplyExecutionEvent(ExecutionEventType type, string? message) =>
    type switch
    {
        ExecutionEventType.BreakpointHit => SetStatus(VMState.Breakpoint),
        ExecutionEventType.Halted => SetStatus(VMState.Halted),
        ExecutionEventType.Error => SetStatusWithError(VMState.Error, message),
    };

private bool SetStatus(VMState state) { Status = state; return true; }
private bool SetStatusWithError(VMState state, string? msg) { Status = state; ErrorMessage = msg; return true; }
```

### 2.4 Tasks (TDD Order)

**Write tests first, then implement:**
- [ ] Write tests for ViewModel initial state
- [ ] Write tests for `RunCommand` execution and state changes
- [ ] Write tests for `StepCommand` with register change detection
- [ ] Write tests for `PauseCommand` enabled/disabled states
- [ ] Write tests for WebSocket event handling (state, output, execution events)
- [ ] Write tests for register highlighting (add, timeout removal)
- [ ] Write tests for session lifecycle (create, destroy, cleanup)

**Implement to pass tests:**
- [ ] Implement `MainWindowViewModel` with reactive properties
- [ ] Implement all commands (Run, Pause, Step, StepOver, StepOut, Reset)
- [ ] Implement register change detection and highlighting pipeline
- [ ] Implement WebSocket event subscription and handling
- [ ] Wire up dependency injection for services

---

## Phase 3: Main Window Layout

**Goal:** Create the main window with split-panel layout

### 3.1 Main Window Structure

```
┌──────────────────────────────────────────────────────────────┐
│ [Toolbar: Status | Load | Run | Pause | Step | StepOver | StepOut | Reset | ShowPC] │
├────────────────────────────┬─────────────────────────────────┤
│                            │  [Tab Bar: Registers | Memory  │
│                            │   Stack | Disasm | Eval |      │
│     Code Editor            │   Watchpoints | Breakpoints]   │
│     (with gutter)          │                                 │
│                            │  [Selected Tab Content]         │
│                            │                                 │
├────────────────────────────┴─────────────────────────────────┤
│                     Console Output                           │
│  ─────────────────────────────────────────────────────────── │
│  Input: [                                        ] [Send]    │
└──────────────────────────────────────────────────────────────┘
```

### 3.2 MainWindow.axaml

```xml
<Window xmlns="https://github.com/avaloniaui"
        xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml"
        xmlns:vm="using:ARMEmulator.ViewModels"
        x:DataType="vm:MainWindowViewModel"
        Title="ARM Emulator">

    <DockPanel>
        <!-- Toolbar -->
        <views:ToolbarView DockPanel.Dock="Top" />

        <!-- Main Content -->
        <Grid>
            <Grid.RowDefinitions>
                <RowDefinition Height="3*" MinHeight="200" />
                <RowDefinition Height="Auto" />
                <RowDefinition Height="*" MinHeight="100" />
            </Grid.RowDefinitions>

            <!-- Editor + Right Panel (Horizontal Split) -->
            <Grid Grid.Row="0">
                <Grid.ColumnDefinitions>
                    <ColumnDefinition Width="2*" MinWidth="300" />
                    <ColumnDefinition Width="Auto" />
                    <ColumnDefinition Width="*" MinWidth="250" />
                </Grid.ColumnDefinitions>

                <views:EditorView Grid.Column="0" />
                <GridSplitter Grid.Column="1" Width="5" />
                <views:RightPanelView Grid.Column="2" />
            </Grid>

            <!-- Vertical Splitter -->
            <GridSplitter Grid.Row="1" Height="5" HorizontalAlignment="Stretch" />

            <!-- Console -->
            <views:ConsoleView Grid.Row="2" />
        </Grid>
    </DockPanel>
</Window>
```

### 3.3 Toolbar

Implement matching Swift toolbar with keyboard shortcuts:

| Action | Icon | Keyboard Shortcut | Enabled When |
|--------|------|-------------------|--------------|
| Load | 📂 | Ctrl+L | Always |
| Run/Continue | ▶️ | F5, Ctrl+R | Not running |
| Pause | ⏸️ | Ctrl+. | Running or waiting |
| Step | ➡️ | F11, Ctrl+T | Idle or breakpoint |
| Step Over | ⤵️ | F10, Ctrl+Shift+T | Idle or breakpoint |
| Step Out | ⤴️ | Ctrl+Alt+T | Idle or breakpoint |
| Reset | 🔄 | Ctrl+Shift+R | Always |
| Show PC | 📍 | Ctrl+J | Always |

### 3.4 Tasks

- [ ] Create MainWindow.axaml with split-panel layout
- [ ] Implement ToolbarView with all buttons
- [ ] Implement GridSplitter for resizable panels
- [ ] Wire up keyboard shortcuts (KeyBindings)
- [ ] Implement status indicator with color coding
- [ ] Create placeholder views for all panels
- [ ] Test layout responsiveness on different window sizes

---

## Phase 4: Editor View

**Goal:** Full-featured assembly editor with gutter and breakpoints

### 4.1 Editor Requirements

| Feature | Implementation |
|---------|----------------|
| Syntax highlighting | AvaloniaEdit with ARM assembly syntax definition |
| Line numbers | Custom gutter with AvaloniaEdit |
| Breakpoint markers | Red circles in gutter |
| Current PC indicator | Blue arrow in gutter |
| Breakpoint toggle | Click gutter to toggle |
| Read-only mode | Locked during execution |
| Scroll to PC | Auto-scroll on step |
| Monospace font | Configurable size (10-24pt) |

### 4.2 ARM Assembly Syntax Highlighting

Create custom syntax definition for ARM assembly:

```csharp
public class ARMAssemblySyntaxHighlighting : IHighlightingDefinition
{
    // Keywords: MOV, ADD, SUB, MUL, LDR, STR, B, BL, CMP, etc.
    // Registers: R0-R15, SP, LR, PC, CPSR
    // Directives: .data, .text, .global, .word, .ascii, .asciz
    // Comments: ; or @
    // Numbers: #0x..., #..., 0x...
    // Labels: identifier followed by :
}
```

### 4.3 Custom Gutter

Using modern patterns with styled properties and functional rendering:

```csharp
namespace ARMEmulator.Controls;

public class EditorGutterMargin : AbstractMargin
{
    // Avalonia styled properties for reactive binding
    public static readonly StyledProperty<ImmutableHashSet<int>> BreakpointLinesProperty =
        AvaloniaProperty.Register<EditorGutterMargin, ImmutableHashSet<int>>(
            nameof(BreakpointLines), []);

    public static readonly StyledProperty<int?> CurrentPCLineProperty =
        AvaloniaProperty.Register<EditorGutterMargin, int?>(nameof(CurrentPCLine));

    public ImmutableHashSet<int> BreakpointLines
    {
        get => GetValue(BreakpointLinesProperty);
        set => SetValue(BreakpointLinesProperty, value);
    }

    public int? CurrentPCLine
    {
        get => GetValue(CurrentPCLineProperty);
        set => SetValue(CurrentPCLineProperty, value);
    }

    // Event for breakpoint toggle (functional callback)
    public event Action<int>? LineClicked;

    // Colors as static readonly for performance
    private static readonly IBrush BreakpointBrush = Brushes.Red;
    private static readonly IBrush PCArrowBrush = Brushes.DodgerBlue;
    private static readonly IBrush LineNumberBrush = Brushes.Gray;

    public override void Render(DrawingContext context)
    {
        if (TextView is not { VisualLinesValid: true } textView) return;

        var lineHeight = textView.DefaultLineHeight;
        var gutterWidth = Bounds.Width;

        // Render each visible line functionally
        foreach (var visualLine in textView.VisualLines)
        {
            var lineNumber = visualLine.FirstDocumentLine.LineNumber;
            var y = visualLine.GetTextLineVisualYPosition(
                visualLine.TextLines[0], VisualYPosition.LineTop) - textView.VerticalOffset;

            RenderLineNumber(context, lineNumber, y, gutterWidth, lineHeight);

            if (BreakpointLines.Contains(lineNumber))
                RenderBreakpoint(context, y, lineHeight);

            if (CurrentPCLine == lineNumber)
                RenderPCArrow(context, y, lineHeight);
        }
    }

    private static void RenderLineNumber(
        DrawingContext ctx, int line, double y, double width, double height)
    {
        var text = new FormattedText(
            line.ToString(),
            CultureInfo.CurrentCulture,
            FlowDirection.LeftToRight,
            new Typeface("JetBrains Mono", FontStyle.Normal, FontWeight.Normal),
            12,
            LineNumberBrush);

        ctx.DrawText(text, new Point(width - text.Width - 8, y + (height - text.Height) / 2));
    }

    private static void RenderBreakpoint(DrawingContext ctx, double y, double height)
    {
        var radius = height * 0.35;
        var center = new Point(12, y + height / 2);
        ctx.DrawEllipse(BreakpointBrush, null, center, radius, radius);
    }

    private static void RenderPCArrow(DrawingContext ctx, double y, double height)
    {
        var arrowY = y + height / 2;
        var geometry = new PolylineGeometry(
            [new Point(4, arrowY - 4), new Point(10, arrowY), new Point(4, arrowY + 4)],
            isFilled: true);
        ctx.DrawGeometry(PCArrowBrush, null, geometry);
    }

    protected override void OnPointerPressed(PointerPressedEventArgs e)
    {
        base.OnPointerPressed(e);

        if (TextView is not { } textView) return;

        var pos = e.GetPosition(textView);
        var line = textView.GetVisualLineFromVisualTop(pos.Y + textView.VerticalOffset);

        if (line is not null)
            LineClicked?.Invoke(line.FirstDocumentLine.LineNumber);

        e.Handled = true;
    }
}
```

### 4.4 Tasks

- [ ] Integrate AvaloniaEdit text editor control
- [ ] Create ARM assembly syntax highlighting definition
- [ ] Implement custom gutter margin for breakpoints
- [ ] Implement breakpoint toggle on gutter click
- [ ] Implement current PC line indicator
- [ ] Implement auto-scroll to current PC
- [ ] Implement read-only mode binding
- [ ] Bind font size to settings
- [ ] Test with various assembly files

---

## Phase 5: Register & Status Views

**Goal:** Display registers and CPSR flags with highlighting

### 5.1 Registers View Layout

```
┌────────────────────────────────────────┐
│ Registers                              │
├──────────────┬──────────────┬──────────┤
│ R0  0x000000 │ R1  0x000000 │ R2  ...  │
│ R3  0x000000 │ R4  0x000000 │ R5  ...  │
│ ...          │ ...          │ ...      │
├──────────────┴──────────────┴──────────┤
│ SP  0x00050000    LR  0x00000000       │
│ PC  0x00008000    CPSR  N-Z-C-V-       │
└────────────────────────────────────────┘
```

### 5.2 Register Display

Each register shows:
- Name (R0, R1, ..., SP, LR, PC)
- Hex value (0x00000000)
- Decimal value (optional tooltip)
- Green highlight when changed (1.5s fade)

### 5.3 CPSR Flags

Display as: `N Z C V` with each letter highlighted when set.

### 5.4 Tasks

- [ ] Create RegistersView.axaml with responsive grid
- [ ] Implement register cell control with hex/decimal display
- [ ] Implement highlight animation (green fade)
- [ ] Display CPSR flags with individual indicators
- [ ] Bind to ChangedRegisters for highlighting
- [ ] Test with register state changes

---

## Phase 6: Memory, Stack & Disassembly Views

**Goal:** Memory inspection tools matching Swift GUI

### 6.1 Memory View

Features:
- Hexdump display (16 bytes per row, 16 rows)
- Address navigation input
- Quick jump buttons (PC, SP, R0-R3)
- Memory write highlighting
- Auto-scroll to writes (configurable)

Layout:
```
┌─────────────────────────────────────────────────────┐
│ Address: [0x00000000] [Go] [PC][SP][R0][R1][R2][R3] │
│ [ ] Auto-scroll to writes                          │
├─────────────────────────────────────────────────────┤
│ 00000000  00 00 00 00 00 00 00 00  ........         │
│ 00000010  00 00 00 00 00 00 00 00  ........         │
│ ...                                                 │
└─────────────────────────────────────────────────────┘
```

### 6.2 Stack View

Features:
- Stack contents display (grows downward)
- Address, value (hex), ASCII
- Stack offset relative to SP
- Annotations (code address, stack address, LR)
- Current SP indicator
- Stack size display

### 6.3 Disassembly View

Features:
- Centered around current PC (±32 instructions)
- Address | Machine Code | Mnemonic columns
- Breakpoint indicators
- PC indicator
- Refresh button

Layout:
```
┌──────────────────────────────────────────────────────┐
│ Disassembly                            [🔄 Refresh]  │
├─────────┬──────────┬─────────────────────────────────┤
│ Address │ Code     │ Instruction                     │
├─────────┼──────────┼─────────────────────────────────┤
│ ● 8000  │ E3A00001 │ MOV R0, #1                     │
│ → 8004  │ E2811001 │ ADD R1, R1, #1                 │
│   8008  │ E1510002 │ CMP R1, R2                     │
└─────────┴──────────┴─────────────────────────────────┘
```

### 6.4 Tasks

- [ ] Create MemoryView with hexdump control
- [ ] Implement address navigation
- [ ] Implement quick jump buttons
- [ ] Implement memory write highlighting
- [ ] Create StackView with annotations
- [ ] Create DisassemblyView with instruction table
- [ ] Implement breakpoint/PC indicators in disassembly
- [ ] Test with various memory states

---

## Phase 7: Debugging Features

**Goal:** Expression evaluator, watchpoints, breakpoints list

### 7.1 Expression Evaluator

Features:
- Expression input field
- Evaluate button
- Result display (hex, decimal, binary)
- History with timestamps
- Error display
- Examples hint ("r0", "r0+r1", "[r0]", "0x8000")

### 7.2 Watchpoints View

Features:
- Add watchpoint form (address, type)
- Type selector (read/write/readwrite)
- List of active watchpoints
- Remove button per watchpoint
- Type icons

### 7.3 Breakpoints List View

Features:
- Combined breakpoints + watchpoints list
- Section headers
- Remove buttons
- Address display with symbols
- Empty state message

### 7.4 Tasks

- [ ] Create ExpressionEvaluatorView
- [ ] Implement expression history storage
- [ ] Implement result formatting (hex/dec/binary)
- [ ] Create WatchpointsView with add form
- [ ] Create BreakpointsListView with sections
- [ ] Implement remove functionality for both
- [ ] Test evaluation with various expressions

---

## Phase 8: Console & Input

**Goal:** Console output with interactive input

### 8.1 Console View

Features:
- Monospace output display
- Auto-scroll to bottom
- Text selection support
- Input field (when enabled)
- Send button with Enter shortcut
- Visual feedback when waiting for input (orange border)

### 8.2 Input Handling

Match Swift logic:
- Detect if VM is waiting for input
- If waiting: send and wait (unblocks pending step)
- If not waiting: send and step (consumes buffered input)
- Auto-append newline

### 8.3 Tasks

- [ ] Create ConsoleView with output TextBox
- [ ] Implement auto-scroll behavior
- [ ] Implement input field with Send button
- [ ] Implement waiting-for-input visual state
- [ ] Implement smart input sending logic
- [ ] Test with interactive programs

---

## Phase 9: File Operations & Dialogs

**Goal:** File management and dialogs

### 9.1 File Commands

| Command | Shortcut | Action |
|---------|----------|--------|
| Open | Ctrl+O | File picker for .s files |
| Save | Ctrl+S | Save current file |
| Save As | Ctrl+Shift+S | Save with new name |
| Recent Files | - | Submenu with history |
| Open Example | Ctrl+Shift+E | Show examples browser |

### 9.2 Examples Browser

Features:
- Split view: list + preview
- Searchable list
- Search highlighting
- File size display
- Load button

### 9.3 Preferences Window

Tabs:
- **General:** Backend URL, color scheme, recent files limit
- **Editor:** Font size with live preview

### 9.4 About Dialog

Display:
- App icon and name
- Backend version info (async load)
- Commit hash and build date

### 9.5 Tasks

- [ ] Implement Open/Save/SaveAs file dialogs
- [ ] Implement recent files menu
- [ ] Create ExamplesBrowserWindow
- [ ] Create PreferencesWindow with tabs
- [ ] Create AboutWindow with version info
- [ ] Persist settings to user config file
- [ ] Test file operations on all platforms

---

## Phase 10: Platform Integration

**Goal:** Platform-specific polish and packaging

### 10.1 Windows

- Native file dialogs
- System tray support (optional)
- Windows installer (MSIX or WiX)
- Start menu integration

### 10.2 macOS

- Native menu bar integration
- .app bundle with backend binary
- Code signing
- DMG installer
- System appearance (light/dark) detection

### 10.3 Linux

- XDG desktop file
- AppImage or Flatpak packaging
- System theme integration

### 10.4 Backend Bundling

The Go backend binary must be bundled with the app:
- **Windows:** `arm-emulator.exe` in app directory
- **macOS:** `arm-emulator` in `Contents/Resources/`
- **Linux:** `arm-emulator` in app directory or `/usr/share/arm-emulator/`

### 10.5 Tasks

- [ ] Configure Windows native dialogs
- [ ] Create Windows installer (MSIX)
- [ ] Configure macOS .app bundle
- [ ] Create macOS DMG installer
- [ ] Create Linux AppImage
- [ ] Implement backend binary bundling for each platform
- [ ] Test backend auto-start on each platform
- [ ] Test health check and recovery

---

## Phase 11: Testing & Documentation

**Goal:** Comprehensive testing and documentation

### 11.1 Unit Tests

| Area | Tests |
|------|-------|
| Models | Serialization, equality, edge cases |
| Services | API client, WebSocket, file service (mocked) |
| ViewModels | State transitions, commands, highlighting |

### 11.2 Integration Tests

- Backend communication (requires running backend)
- Full execution cycle (load, run, step, breakpoint)
- WebSocket event handling

### 11.3 UI Tests (Optional)

- Avalonia Headless testing for views
- Screenshot comparison tests

### 11.4 Documentation

- README.md with build instructions
- User guide
- Developer guide
- API documentation

### 11.5 Tasks

- [ ] Write unit tests for all models
- [ ] Write unit tests for services with mocks
- [ ] Write unit tests for ViewModels
- [ ] Write integration tests for backend communication
- [ ] Create README.md with full instructions
- [ ] Document keyboard shortcuts
- [ ] Document configuration options

---

## Phase 12: Polish & Release

**Goal:** Final polish and release preparation

### 12.1 UI Polish

- Consistent spacing and alignment
- Loading indicators
- Error messages and alerts
- Tooltips for all buttons
- Empty states for lists
- Responsive layout testing

### 12.2 Performance

- Profile memory usage
- Optimize large memory display
- Optimize disassembly rendering
- Test with large programs

### 12.3 Accessibility

- Keyboard navigation
- Screen reader support
- High contrast theme support

### 12.4 Release Checklist

- [ ] All features implemented and tested
- [ ] No compiler warnings
- [ ] All tests passing
- [ ] Performance acceptable
- [ ] Documentation complete
- [ ] Installers built for all platforms
- [ ] Version number set
- [ ] Release notes written

---

## Feature Parity Checklist

### Editor Features
- [ ] Syntax-aware assembly editing (monospace font, configurable size)
- [ ] Line numbers in gutter
- [ ] Breakpoint visual indicators (red circles)
- [ ] Current PC indicator (blue arrow)
- [ ] Breakpoint toggle (click gutter or F9)
- [ ] Editor read-only during execution
- [ ] Editor editable when idle/halted/error
- [ ] Scroll to current PC on step
- [ ] Text selection

### Execution Control
- [ ] Run/Continue (F5)
- [ ] Pause
- [ ] Step single instruction (F11)
- [ ] Step over functions (F10)
- [ ] Step out of function
- [ ] Reset VM
- [ ] Load program
- [ ] Compilation with error reporting

### Register Inspection
- [ ] Display all 16 registers (R0-R12, SP, LR, PC)
- [ ] Display CPSR flags (N, Z, C, V)
- [ ] Hex and decimal value display
- [ ] Changed register highlighting with animation
- [ ] Register change detection
- [ ] Responsive column layout

### Memory Operations
- [ ] Hexdump display (16 bytes per row)
- [ ] Navigate to address
- [ ] Quick jump to PC, SP, R0-R3
- [ ] Memory write tracking and highlighting
- [ ] Auto-scroll to memory writes (configurable)

### Stack Inspection
- [ ] Display stack contents
- [ ] Address, value (hex), ASCII
- [ ] Stack offset calculation
- [ ] Stack size calculation
- [ ] Annotation detection
- [ ] Current SP indicator

### Disassembly
- [ ] Live disassembly around current PC
- [ ] Address | Machine Code | Mnemonic format
- [ ] Breakpoint markers
- [ ] PC indicator
- [ ] Manual refresh

### Debugging Features
- [ ] Breakpoints (add/remove from gutter and disassembly)
- [ ] Breakpoints list view
- [ ] Watchpoints (read/write/readwrite)
- [ ] Watchpoint list view
- [ ] Expression evaluation
- [ ] Expression history

### Console I/O
- [ ] Program output display
- [ ] Interactive input field
- [ ] Input waiting state indicator
- [ ] Auto-scroll

### File Management
- [ ] Open file dialog
- [ ] Save/Save As
- [ ] Recent files list
- [ ] Example programs browser with search
- [ ] Command-line argument file loading

### Settings & Preferences
- [ ] Backend URL configuration
- [ ] Editor font size
- [ ] Color scheme selection (auto/light/dark)
- [ ] Recent files limit

### Application Features
- [ ] Backend lifecycle management
- [ ] Backend status monitoring
- [ ] About dialog with version info
- [ ] Global keyboard shortcuts
- [ ] Connection status indicator
- [ ] Error message display
- [ ] Tab-based right panel navigation

---

## Technology Stack Summary

| Component | Technology |
|-----------|------------|
| **Runtime** | .NET 10 with C# 13 |
| **UI Framework** | Avalonia UI 11.3.x |
| **MVVM Framework** | ReactiveUI 20.x with source generators |
| **Text Editor** | AvaloniaEdit 11.x |
| **HTTP Client** | HttpClient (built-in, with IHttpClientFactory) |
| **WebSocket** | ClientWebSocket (System.Net.WebSockets) |
| **Reactive Extensions** | System.Reactive 6.x |
| **JSON Serialization** | System.Text.Json (source-generated) |
| **Dependency Injection** | Microsoft.Extensions.DependencyInjection 10.x |
| **Testing** | xUnit 2.x, NSubstitute 5.x, FluentAssertions 7.x |
| **UI Testing** | Avalonia.Headless.XUnit |
| **Packaging** | MSIX (Windows), DMG (macOS), AppImage (Linux) |

### C# 13 Features Used

| Feature | Usage |
|---------|-------|
| **Primary constructors** | ViewModels, Services |
| **Collection expressions** | `[1, 2, 3]`, `[]` for empty |
| **Pattern matching** | Exhaustive switch expressions |
| **Records** | All data models (immutable by default) |
| **Required members** | Enforced initialization |
| **File-scoped namespaces** | Cleaner files |
| **Raw string literals** | Multi-line JSON, XML |
| **`params` collections** | Flexible APIs |
| **Extension types** (preview) | Behavior on enums |

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| **Breaking Swift GUI with API changes** | Treat API as frozen; additive changes only; test both frontends |
| AvaloniaEdit limitations | Fall back to plain TextBox with manual gutter if needed |
| Platform-specific issues | Test early and often on all three platforms |
| WebSocket reliability | Implement robust reconnection with exponential backoff |
| Backend process management | Platform-specific process handling with health checks |
| Performance with large programs | Virtual scrolling, lazy loading, profiling |
| API drift between frontends | Use Swift `APIClient.swift` as reference; sync changes to both |

---

## Success Criteria

1. **Feature Parity:** 100% of Swift GUI features implemented
2. **Platform Support:** Works on Windows 10+, macOS 13+, Ubuntu 22.04+
3. **Performance:** Responsive UI with programs up to 10,000 instructions
4. **Reliability:** Graceful handling of backend failures with recovery
5. **Usability:** Consistent UX across platforms
6. **Maintainability:** Clean architecture, comprehensive tests, documentation
7. **Code Quality:** Modern C# 13 idioms, functional patterns, zero warnings

---

## Appendix: Modern C# Patterns Reference

### Collection Expressions (C# 12+)

```csharp
// Empty collections
ImmutableArray<int> empty = [];
ImmutableHashSet<string> emptySet = [];

// Inline initialization
ImmutableArray<int> numbers = [1, 2, 3, 4, 5];

// Spread operator
ImmutableArray<int> combined = [..numbers, 6, 7, 8];

// In method calls
DoSomething([1, 2, 3]);
```

### Primary Constructors (C# 12+)

```csharp
// Class with primary constructor (captures parameters as fields)
public class ApiClient(HttpClient http, ILogger logger) : IApiClient
{
    public async Task<T> GetAsync<T>(string path) =>
        await http.GetFromJsonAsync<T>(path) ?? throw new ApiException("Null response");
}

// Record with primary constructor (auto-generates properties)
public sealed record Watchpoint(int Id, uint Address, WatchpointType Type);
```

### Pattern Matching (C# 11+)

```csharp
// List patterns
int[] arr = [1, 2, 3];
var result = arr switch
{
    [] => "empty",
    [var single] => $"one: {single}",
    [var first, .., var last] => $"first: {first}, last: {last}",
};

// Property patterns with nested matching
var description = status switch
{
    { State: VMState.Error, Error: var msg } => $"Error: {msg}",
    { State: VMState.Running, PC: var pc } => $"Running at 0x{pc:X8}",
    { State: VMState.Halted } => "Program completed",
    _ => "Unknown state"
};
```

### Extension Methods for Fluent APIs

```csharp
public static class ObservableExtensions
{
    public static IDisposable DisposeWith(this IDisposable disposable, CompositeDisposable composite)
    {
        composite.Add(disposable);
        return disposable;
    }

    public static IObservable<T> WhereNotNull<T>(this IObservable<T?> source) where T : class =>
        source.Where(x => x is not null).Select(x => x!);
}
```

### Async/Await with LINQ

```csharp
// Process items concurrently, collecting results (exceptions propagate naturally)
var results = await addresses
    .ToAsyncEnumerable()
    .SelectAwaitWithCancellation(async (addr, ct) =>
        await _api.GetMemoryAsync(sessionId, addr, 16, ct))
    .ToListAsync(ct);

// Or with exception handling per item (when partial failure is acceptable)
var partialResults = new List<ImmutableArray<byte>>();
foreach (var addr in addresses)
{
    try { partialResults.Add(await _api.GetMemoryAsync(sessionId, addr, 16, ct)); }
    catch (ApiException ex) { logger.LogWarning(ex, "Skipping address {Address}", addr); }
}
```
