# API Integration Tests Design

**Date:** 2026-01-04
**Status:** Approved
**Scope:** Comprehensive API integration tests for all 49 example programs

## Overview

This design adds comprehensive API-level integration tests that validate the complete HTTP REST API stack (session management, program loading, stdin handling, execution, output retrieval) by running all 49 example programs through the API and comparing their output against expected results.

## Design Decisions

### 1. Scope
**Comprehensive coverage (all 49 programs)** - Test every example program via the API to ensure complete stack validation, including programs with and without stdin/stdout interaction.

### 2. Test Location
**`tests/integration/api_example_programs_test.go`** - Co-locate with existing VM-level integration tests, reuse `expected_outputs/*.txt` files for consistency.

### 3. Server Setup
**Direct handler invocation (`httptest.ResponseRecorder`)** - Use the same approach as existing API unit tests for speed and simplicity. No real network layer needed.

### 4. Stdin Strategy
**Hybrid approach:**
- **Batch mode:** Programs with predetermined input (fibonacci.s, bubble_sort.s) send all stdin upfront via POST `/api/v1/session/{id}/stdin`, then run
- **Interactive mode:** Programs with loops/menus (calculator.s) send stdin incrementally based on WebSocket state updates

### 5. State Monitoring
**WebSocket integration** - Use WebSocket connections to monitor execution state in real-time, detect when programs are waiting for stdin, and know when execution completes.

---

## Architecture

### Core Components

1. **Test Server Setup**
   - Create `api.Server` instance per test or shared across test suite
   - Invoke handlers via `httptest.ResponseRecorder` (no real HTTP server)
   - Each test case gets isolated session for clean state

2. **WebSocket Test Client**
   - Lightweight WebSocket client connects to `/api/v1/ws`
   - Monitors state updates (running, halted, waiting_for_input, breakpoint)
   - Channel-based message reception for non-blocking operation
   - Timeout protection to prevent hung tests

3. **Test Case Structure**
   ```go
   type apiTestCase struct {
       name           string // Test name (e.g., "Fibonacci_API")
       programFile    string // Assembly file in examples/
       expectedOutput string // Expected output file in expected_outputs/
       stdin          string // Optional stdin input
       stdinMode      string // "batch" or "interactive"
   }
   ```

4. **Stdin Strategy Handler**
   - Dispatches to batch or interactive stdin handling
   - Batch: Send all input upfront, then run
   - Interactive: Monitor WebSocket, send input incrementally

### Test Flow

```
1. Create session        → POST /api/v1/session
2. Establish WebSocket   → /api/v1/ws?session={id}
3. Load program          → POST /api/v1/session/{id}/load
4. Apply stdin strategy  → Batch or Interactive
5. Start execution       → POST /api/v1/session/{id}/run
6. Monitor completion    → WebSocket state updates
7. Retrieve output       → GET /api/v1/session/{id}/console
8. Compare output        → Against expected_outputs/*.txt
9. Cleanup               → DELETE /api/v1/session/{id}
```

---

## Implementation Details

### Test Structure

```go
// tests/integration/api_example_programs_test.go
package integration_test

func TestAPIExamplePrograms(t *testing.T) {
    tests := []struct {
        name           string
        programFile    string
        expectedOutput string
        stdin          string
        stdinMode      string
    }{
        {
            name:           "Fibonacci_API",
            programFile:    "fibonacci.s",
            expectedOutput: "fibonacci.txt",
            stdin:          "10\n",
            stdinMode:      "batch",
        },
        {
            name:           "Calculator_API",
            programFile:    "calculator.s",
            expectedOutput: "calculator.txt",
            stdin:          "15\n+\n7\nq\n",
            stdinMode:      "interactive",
        },
        // ... all 49 programs
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server := createTestServer()
            output, exitCode, err := runProgramViaAPI(
                t, server, tt.programFile, tt.stdin, tt.stdinMode)

            if err != nil {
                t.Fatalf("execution failed: %v", err)
            }
            if exitCode != 0 {
                t.Errorf("expected exit code 0, got %d", exitCode)
            }

            // Load and compare expected output
            expectedPath := filepath.Join("expected_outputs", tt.expectedOutput)
            expected, _ := os.ReadFile(expectedPath)

            if output != string(expected) {
                t.Errorf("output mismatch\nExpected:\n%s\nGot:\n%s",
                         expected, output)
            }
        })
    }
}
```

### Helper Functions

#### `runProgramViaAPI()`
Main orchestrator for API test flow:

```go
func runProgramViaAPI(t *testing.T, server *api.Server,
                      programFile, stdin, stdinMode string) (string, int32, error) {
    t.Helper()

    // 1. Create session
    sessionID := createAPISession(t, server)
    defer destroySession(t, server, sessionID)

    // 2. Establish WebSocket
    wsClient := NewWebSocketTestClient(t, server, sessionID)
    defer wsClient.Close()

    // 3. Load program
    source, _ := os.ReadFile(filepath.Join("../../examples", programFile))
    loadProgramViaAPI(t, server, sessionID, string(source))

    // 4. Handle stdin based on mode
    if stdinMode == "batch" {
        if stdin != "" {
            sendStdinBatch(t, server, sessionID, stdin)
        }
        startExecution(t, server, sessionID)
        wsClient.WaitForState("halted", 10*time.Second)
    } else { // interactive
        startExecution(t, server, sessionID)
        sendStdinInteractive(t, server, sessionID, stdin, wsClient)
    }

    // 5. Get output
    output := getConsoleOutput(t, server, sessionID)
    regs := getRegisters(t, server, sessionID)

    return output, regs.R0, nil
}
```

#### `createAPISession()`
Create session via REST API:

```go
func createAPISession(t *testing.T, server *api.Server) string {
    t.Helper()

    req := httptest.NewRequest(http.MethodPost, "/api/v1/session",
                               bytes.NewReader([]byte("{}")))
    w := httptest.NewRecorder()
    server.Handler().ServeHTTP(w, req)

    if w.Code != http.StatusCreated {
        t.Fatalf("Failed to create session: %d", w.Code)
    }

    var resp api.SessionCreateResponse
    json.NewDecoder(w.Body).Decode(&resp)
    return resp.SessionID
}
```

#### `loadProgramViaAPI()`
Load program via REST API:

```go
func loadProgramViaAPI(t *testing.T, server *api.Server,
                       sessionID, source string) {
    t.Helper()

    reqBody := api.LoadProgramRequest{Source: source}
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest(http.MethodPost,
        fmt.Sprintf("/api/v1/session/%s/load", sessionID),
        bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    server.Handler().ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("Failed to load program: %d %s", w.Code, w.Body.String())
    }

    var resp api.LoadProgramResponse
    json.NewDecoder(w.Body).Decode(&resp)
    if !resp.Success {
        t.Fatalf("Program load errors: %v", resp.Errors)
    }
}
```

### Stdin Handling

#### Batch Mode

```go
func sendStdinBatch(t *testing.T, server *api.Server,
                    sessionID, stdin string) {
    t.Helper()

    reqBody := api.StdinRequest{Data: stdin}
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest(http.MethodPost,
        fmt.Sprintf("/api/v1/session/%s/stdin", sessionID),
        bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    server.Handler().ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("Failed to send stdin: %d %s", w.Code, w.Body.String())
    }
}
```

**Flow:** Send all stdin → Call `/run` → Wait for halt → Get output

#### Interactive Mode

```go
func sendStdinInteractive(t *testing.T, server *api.Server,
                          sessionID, stdin string,
                          wsClient *WebSocketTestClient) {
    t.Helper()

    // Split stdin into chunks (lines)
    inputs := strings.Split(stdin, "\n")
    inputIndex := 0

    // Monitor WebSocket for stdin requests
    for inputIndex < len(inputs) {
        stateUpdate := wsClient.WaitForStateUpdate(5 * time.Second)

        // Check if VM waiting for stdin
        if stateUpdate.State == "waiting_for_input" {
            // Send next input chunk
            sendStdinChunk(t, server, sessionID, inputs[inputIndex]+"\n")
            inputIndex++
        }

        // Check if program completed
        if stateUpdate.State == "halted" {
            break
        }

        // Guard: all input sent but still waiting
        if inputIndex >= len(inputs) &&
           stateUpdate.State == "waiting_for_input" {
            t.Fatalf("Program waiting for stdin but all input exhausted")
        }
    }
}

func sendStdinChunk(t *testing.T, server *api.Server,
                    sessionID, chunk string) {
    // Same as sendStdinBatch but for single chunk
    sendStdinBatch(t, server, sessionID, chunk)
}
```

**Flow:** Call `/run` → Monitor WebSocket → Send stdin when requested → Repeat

### WebSocket Test Client

```go
type WebSocketTestClient struct {
    conn    *websocket.Conn
    updates chan StateUpdate
    errors  chan error
    done    chan struct{}
    mu      sync.Mutex
}

type StateUpdate struct {
    State     string // "running", "halted", "waiting_for_input", "breakpoint"
    Registers map[string]uint32
    PC        uint32
    Cycles    int64
}

func NewWebSocketTestClient(t *testing.T, server *api.Server,
                             sessionID string) *WebSocketTestClient {
    t.Helper()

    // Create WebSocket connection
    wsURL := fmt.Sprintf("/api/v1/ws?session=%s", sessionID)

    // Use gorilla/websocket testing utilities or httptest.NewServer
    // for establishing WebSocket connection in tests

    client := &WebSocketTestClient{
        updates: make(chan StateUpdate, 10),
        errors:  make(chan error, 10),
        done:    make(chan struct{}),
    }

    // Start background message receiver
    go client.receiveLoop()

    return client
}

func (c *WebSocketTestClient) WaitForStateUpdate(timeout time.Duration) StateUpdate {
    select {
    case update := <-c.updates:
        return update
    case err := <-c.errors:
        panic(fmt.Sprintf("WebSocket error: %v", err))
    case <-time.After(timeout):
        panic("Timeout waiting for state update")
    }
}

func (c *WebSocketTestClient) WaitForState(targetState string,
                                           timeout time.Duration) StateUpdate {
    deadline := time.Now().Add(timeout)
    for {
        remaining := time.Until(deadline)
        if remaining <= 0 {
            panic(fmt.Sprintf("Timeout waiting for state: %s", targetState))
        }

        update := c.WaitForStateUpdate(remaining)
        if update.State == targetState {
            return update
        }
    }
}

func (c *WebSocketTestClient) receiveLoop() {
    defer close(c.done)
    for {
        var update StateUpdate
        if err := c.conn.ReadJSON(&update); err != nil {
            if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
                return
            }
            c.errors <- err
            return
        }
        c.updates <- update
    }
}

func (c *WebSocketTestClient) Close() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.conn != nil {
        c.conn.Close()
        <-c.done // Wait for receive loop to finish
    }
    return nil
}
```

**Key Features:**
- Non-blocking message reception via channels
- Timeout protection for all waits
- Clean shutdown handling
- Buffered channels prevent blocking

---

## Execution Timelines

### Batch Mode (fibonacci.s)

```
[Session Create]
    ↓
[WebSocket Connect]
    ↓
[Load Program: fibonacci.s]
    ↓
[Send All Stdin: "10\n"]
    ↓
[POST /run]
    ↓
[WS: state=running]
    ↓
[WS: state=halted]
    ↓
[GET /console → "Fibonacci sequence: 0, 1, 1, 2, 3, 5, 8, 13, 21, 34"]
    ↓
[Compare vs expected_outputs/fibonacci.txt]
    ↓
[Cleanup Session]
```

### Interactive Mode (calculator.s)

```
[Session Create]
    ↓
[WebSocket Connect]
    ↓
[Load Program: calculator.s]
    ↓
[POST /run]
    ↓
[WS: state=waiting_for_input] → [Send "15\n"]
    ↓
[WS: state=running]
    ↓
[WS: state=waiting_for_input] → [Send "+\n"]
    ↓
[WS: state=running]
    ↓
[WS: state=waiting_for_input] → [Send "7\n"]
    ↓
[WS: state=running]
    ↓
[WS: state=waiting_for_input] → [Send "q\n"]
    ↓
[WS: state=halted]
    ↓
[GET /console → "Result: 15 + 7 = 22"]
    ↓
[Compare vs expected_outputs/calculator.txt]
    ↓
[Cleanup Session]
```

---

## Error Handling

### 1. Program Load Failures

```go
var loadResp api.LoadProgramResponse
json.NewDecoder(w.Body).Decode(&loadResp)
if !loadResp.Success {
    t.Fatalf("Failed to load %s: %v", programFile, loadResp.Errors)
}
```

### 2. Execution Timeouts

```go
const maxExecutionTime = 10 * time.Second

select {
case update := <-wsClient.updates:
    if update.State == "halted" {
        break
    }
case <-time.After(maxExecutionTime):
    t.Fatalf("Program execution timeout after %v", maxExecutionTime)
}
```

### 3. WebSocket Connection Failures

```go
func connectWithRetry(t *testing.T, server *api.Server,
                      sessionID string) *WebSocketTestClient {
    for attempt := 0; attempt < 3; attempt++ {
        client, err := tryConnect(server, sessionID)
        if err == nil {
            return client
        }
        time.Sleep(time.Duration(attempt*100) * time.Millisecond)
    }
    t.Fatal("Failed to establish WebSocket after retries")
    return nil
}
```

### 4. Stdin Overflow (Interactive Mode)

```go
if inputIndex >= len(inputs) &&
   stateUpdate.State == "waiting_for_input" {
    t.Fatalf("Program waiting for stdin but all input exhausted")
}
```

### 5. Output Mismatch Details

```go
if output != expected {
    t.Errorf("Output mismatch for %s:\n"+
             "Expected (%d bytes):\n%s\n"+
             "Got (%d bytes):\n%s\n"+
             "Diff:\n%s",
             programFile,
             len(expected), expected,
             len(output), output,
             diffStrings(expected, output))
}
```

---

## Test Parallelization

- Mark tests as `t.Parallel()` where safe
- Each test gets isolated session - no shared state
- WebSocket clients are per-session - no conflicts
- Server can be shared across parallel tests (stateless handler)

---

## Test Case Categorization

Programs will be categorized by stdin mode:

### Batch Mode (No Interaction)
- hello.s, loops.s, arithmetic.s, conditionals.s, functions.s
- factorial.s, recursive_fib.s, arrays.s, linked_list.s
- And 30+ more non-interactive programs

### Interactive Mode (Stdin Required)
- fibonacci.s (prompt for count)
- calculator.s (menu loop)
- bubble_sort.s (array input)
- gcd.s (two number input)
- celsius_to_fahrenheit.s (temperature input)
- times_table.s (multiplier input)
- string_reverse.s (string input)

---

## Benefits

1. **Complete Stack Validation** - Tests API server, session management, stdin/stdout handling, WebSocket state updates
2. **Regression Protection** - Catches API-level bugs that VM tests might miss
3. **Real-World Scenarios** - Tests how GUI clients actually use the API
4. **Comprehensive Coverage** - All 49 programs tested via production API paths
5. **Reuses Test Infrastructure** - Leverages existing expected output files
6. **Automated Quality Gate** - CI/CD can run these tests before deployment

---

## Testing the Tests

To validate the test framework itself:

1. Run against known-good programs (hello.s, fibonacci.s)
2. Introduce intentional bugs (wrong stdin, modified program)
3. Verify tests catch the bugs
4. Test timeout handling (infinite loop program)
5. Test WebSocket disconnection recovery
6. Test concurrent session isolation

---

## Future Enhancements

1. **Performance Benchmarking** - Track API response times across test runs
2. **Coverage Metrics** - Measure API endpoint coverage
3. **Stress Testing** - Run multiple sessions concurrently
4. **Failure Injection** - Test error handling paths (OOM, invalid opcodes)
5. **WebSocket Message Validation** - Verify state update accuracy
6. **API Versioning Tests** - Test backward compatibility

---

## Implementation Checklist

- [ ] Create `tests/integration/api_example_programs_test.go`
- [ ] Implement `WebSocketTestClient` with channel-based messaging
- [ ] Implement `runProgramViaAPI()` orchestrator
- [ ] Implement `sendStdinBatch()` for batch mode
- [ ] Implement `sendStdinInteractive()` for interactive mode
- [ ] Add helper functions (createAPISession, loadProgramViaAPI, etc.)
- [ ] Categorize all 49 programs into batch/interactive modes
- [ ] Add test cases for all 49 programs
- [ ] Implement error handling for all failure modes
- [ ] Add timeout protection for all waits
- [ ] Test the test framework with known-good programs
- [ ] Run full test suite and verify all programs pass
- [ ] Add CI/CD integration
- [ ] Document test execution in CLAUDE.md

---

## Conclusion

This design provides comprehensive API integration testing that validates the complete HTTP REST API + WebSocket stack by running all 49 example programs through production API paths. The hybrid stdin strategy (batch vs interactive) handles both simple and complex programs appropriately, while the WebSocket client enables proper state monitoring for interactive execution flows.
