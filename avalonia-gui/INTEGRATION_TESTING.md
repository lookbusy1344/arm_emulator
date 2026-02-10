# Integration Testing Guide

Guide to running integration tests for the ARM Emulator Avalonia GUI.

## Overview

The Avalonia GUI includes comprehensive integration tests that verify end-to-end communication with the Go backend. These tests are **skipped by default** because they require a running backend instance.

**Test Suite**: 9 integration tests covering:
- Session management
- Program loading (success and error cases)
- Execution control (step, run, pause)
- Breakpoint management
- Memory operations
- Disassembly
- Expression evaluation
- Error handling

## Prerequisites

### 1. Build the Go Backend

```bash
cd /path/to/arm_emulator
make build
```

This creates the `arm-emulator` binary in the project root.

### 2. Start the Backend

```bash
./arm-emulator
```

**Expected Output**:
```
Starting ARM Emulator backend on http://localhost:8080
Press Ctrl+C to stop
```

Keep this terminal open while running integration tests.

### 3. Verify Backend is Running

```bash
curl http://localhost:8080/api/v1/version
```

**Expected Response**:
```json
{"version":"1.0.0","commit":"abc123","buildDate":"2024-01-01"}
```

## Running Integration Tests

### Option 1: Enable All Integration Tests

Edit the integration test file to remove `Skip` attributes:

```bash
# Open the file
$EDITOR avalonia-gui/ARMEmulator.Tests/Integration/BackendIntegrationTests.cs

# Remove all lines containing:
# Skip = "Requires running backend at localhost:8080 - remove Skip to enable"

# Also remove the corresponding #pragma warning disable/restore lines
```

Then run tests:

```bash
cd avalonia-gui
dotnet test --logger "console;verbosity=normal"
```

### Option 2: Run Individual Integration Tests

Use xUnit's test filtering to run specific tests:

```bash
cd avalonia-gui

# Run a specific test
dotnet test --filter "FullyQualifiedName~BackendIntegrationTests.HealthCheck_BackendAvailable_ReturnsVersion"

# Run all integration tests (they will be skipped by default)
dotnet test --filter "FullyQualifiedName~BackendIntegrationTests"
```

### Option 3: Programmatically Remove Skip Attribute

Temporarily modify tests in your IDE:

```csharp
// Before (skipped)
[Fact(Skip = "Requires running backend...")]
public async Task HealthCheck_BackendAvailable_ReturnsVersion()

// After (enabled)
[Fact]
public async Task HealthCheck_BackendAvailable_ReturnsVersion()
```

## Integration Test Descriptions

### 1. Health Check

**Test**: `HealthCheck_BackendAvailable_ReturnsVersion`
**Purpose**: Verify backend is reachable and returns version info
**Validates**:
- Backend connectivity
- Version endpoint response format

### 2. Full Execution Cycle

**Test**: `FullExecutionCycle_LoadStepRun_CompletesSuccessfully`
**Purpose**: Complete program execution from load to halt
**Validates**:
- Session creation
- Program loading
- Single-step execution
- Register state updates
- Run to completion
- Halt state

**Test Program**:
```assembly
MOV R0, #5      ; R0 = 5
MOV R1, #3      ; R1 = 3
ADD R2, R0, R1  ; R2 = 8
SWI 0           ; exit
```

### 3. Parse Error Handling

**Test**: `LoadProgram_WithSyntaxError_ThrowsProgramLoadException`
**Purpose**: Verify parse errors are properly reported
**Validates**:
- Invalid instruction detection
- Error message propagation
- ProgramLoadException thrown

### 4. Breakpoint Management

**Test**: `Breakpoints_AddAndRemove_WorksCorrectly`
**Purpose**: Verify breakpoint add/remove operations
**Validates**:
- Add breakpoint at specific address
- List breakpoints
- Remove breakpoint
- Breakpoint list updates

### 5. Memory Operations

**Test**: `Memory_ReadAndWrite_ReturnsCorrectData`
**Purpose**: Verify memory read operations
**Validates**:
- Memory read at specific address
- Correct byte count returned
- Memory content accuracy

### 6. Disassembly

**Test**: `Disassembly_GetInstructions_ReturnsFormattedCode`
**Purpose**: Verify disassembly endpoint
**Validates**:
- Disassembly around PC
- Instruction count
- Address correctness
- Format consistency

### 7. Expression Evaluation (Valid)

**Test**: `ExpressionEvaluation_ValidExpression_ReturnsValue`
**Purpose**: Verify expression evaluator with valid expressions
**Validates**:
- Register value evaluation
- Numeric result
- Expression parsing

**Example Expressions**: `r0`, `r0+r1`, `[r0]`, `0x8000`

### 8. Expression Evaluation (Invalid)

**Test**: `ExpressionEvaluation_InvalidExpression_ThrowsException`
**Purpose**: Verify error handling for invalid expressions
**Validates**:
- Syntax error detection
- ExpressionEvaluationException thrown
- Error message clarity

### 9. Session Not Found

**Test**: `SessionNotFound_ThrowsSessionNotFoundException`
**Purpose**: Verify error handling for invalid session IDs
**Validates**:
- HTTP 404 response
- SessionNotFoundException thrown
- Error message includes session ID

## Expected Output

### All Tests Pass

```
Test Run Successful.
Total tests: 244
     Passed: 244
    Skipped: 0
 Total time: 5.2 seconds
```

### Some Tests Skipped (Default)

```
Test Run Successful.
Total tests: 244
     Passed: 235
    Skipped: 9
 Total time: 4.5 seconds
```

### Test Failures (Backend Not Running)

```
Failed!  - Failed: 9, Passed: 235, Skipped: 0, Total: 244

  ARMEmulator.Tests.Integration.BackendIntegrationTests.HealthCheck_BackendAvailable_ReturnsVersion
    System.Net.Http.HttpRequestException: Connection refused
```

## Troubleshooting

### Connection Refused

**Symptom**: Tests fail with "Connection refused" or "No connection could be made"
**Cause**: Backend not running or wrong URL
**Solution**:
1. Start backend: `./arm-emulator`
2. Verify URL: `http://localhost:8080`
3. Check firewall settings

### Timeout Errors

**Symptom**: Tests fail with timeout after 5 seconds
**Cause**: Backend slow to respond or overloaded
**Solution**:
1. Check backend logs for errors
2. Restart backend
3. Increase timeout in test constructor (if needed)

### Port Already in Use

**Symptom**: Backend fails to start: "address already in use"
**Cause**: Another process using port 8080
**Solution**:
```bash
# Find process using port 8080
lsof -ti:8080

# Kill process (macOS/Linux)
kill -9 $(lsof -ti:8080)

# Or use different port
./arm-emulator --port 8081
```

Then update tests to use custom port:
```csharp
_httpClient = new HttpClient {
    BaseAddress = new Uri("http://localhost:8081")
};
```

### Tests Pass Locally, Fail in CI

**Symptom**: Integration tests fail in CI pipeline
**Cause**: CI environment doesn't have running backend
**Solution**: CI should skip integration tests by default (they use `Skip` attribute)

## CI/CD Integration

### Skip Integration Tests in CI

Integration tests are skipped by default using the `Skip` attribute. No special configuration needed for CI.

### Run Integration Tests in CI

To run integration tests in CI:

1. **Start Backend in CI Pipeline**:
```yaml
# GitHub Actions example
- name: Start Backend
  run: |
    make build
    ./arm-emulator &
    sleep 2  # Wait for backend to start

- name: Wait for Backend
  run: |
    timeout 30 bash -c 'until curl -s http://localhost:8080/api/v1/version; do sleep 1; done'
```

2. **Enable Integration Tests**:
```yaml
- name: Run Integration Tests
  run: |
    cd avalonia-gui
    # Remove Skip attributes programmatically
    sed -i 's/\[Fact(Skip = ".*")\]/[Fact]/' ARMEmulator.Tests/Integration/*.cs
    dotnet test --filter "FullyQualifiedName~BackendIntegrationTests"
```

## Manual Testing Workflow

### Full Integration Test Run

```bash
# Terminal 1: Start backend
cd /path/to/arm_emulator
make build
./arm-emulator

# Terminal 2: Run tests
cd avalonia-gui

# Enable integration tests (remove Skip attributes)
# Edit ARMEmulator.Tests/Integration/BackendIntegrationTests.cs

# Run all tests
dotnet test

# Or run just integration tests
dotnet test --filter "Category=Integration"
```

### Quick Smoke Test

```bash
# Terminal 1: Backend
./arm-emulator

# Terminal 2: Single test
cd avalonia-gui
dotnet test --filter "HealthCheck_BackendAvailable_ReturnsVersion"
```

## Test Maintenance

### Adding New Integration Tests

1. Add test method to `BackendIntegrationTests.cs`
2. Mark with `[Fact(Skip = "...")]` attribute
3. Add `#pragma warning disable/restore xUnit1004` around attribute
4. Use `_apiClient` for API calls
5. Follow existing test patterns
6. Include cleanup in `finally` block

**Template**:
```csharp
#pragma warning disable xUnit1004
[Fact(Skip = "Requires running backend at localhost:8080 - remove Skip to enable")]
public async Task NewTest_Description()
#pragma warning restore xUnit1004
{
    var session = await _apiClient.CreateSessionAsync(_cts.Token);

    try {
        // Test code here

        // Assertions
        result.Should().Be(expected);
    }
    finally {
        await _apiClient.DestroySessionAsync(session.SessionId, _cts.Token);
    }
}
```

### Updating Existing Tests

1. Maintain backward compatibility
2. Update expected outputs if API changes
3. Keep tests independent (no shared state)
4. Test both success and failure cases

## Performance Considerations

- **Test Duration**: Full suite runs in ~5 seconds with backend
- **Resource Usage**: Lightweight, suitable for frequent runs
- **Parallelization**: Tests run sequentially (session-based)

## See Also

- [README.md](README.md) - Build and run instructions
- [KEYBOARD_SHORTCUTS.md](KEYBOARD_SHORTCUTS.md) - Keyboard shortcuts
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration guide
- [../docs/AVALONIA_IMPLEMENTATION_PLAN.md](../docs/AVALONIA_IMPLEMENTATION_PLAN.md) - Architecture details
