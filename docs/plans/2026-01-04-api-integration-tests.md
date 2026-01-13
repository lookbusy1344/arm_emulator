# API Integration Tests Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement comprehensive API integration tests for all 49 example programs with WebSocket state monitoring and hybrid stdin handling.

**Architecture:** Create `tests/integration/api_example_programs_test.go` with WebSocket test client, REST API helpers, and hybrid stdin strategy (batch upfront + interactive incremental). Reuse existing `expected_outputs/*.txt` files.

**Tech Stack:** Go 1.21+, net/http/httptest, gorilla/websocket, existing api package

---

## Task 1: Create Test File and Basic Structure

**Files:**
- Create: `tests/integration/api_example_programs_test.go`

**Step 1: Create test file with package and imports**

```go
package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lookbusy1344/arm-emulator/api"
)
```

**Step 2: Add test skeleton**

```go
// TestAPIExamplePrograms runs integration tests for example programs via REST API
func TestAPIExamplePrograms(t *testing.T) {
	// Placeholder - will add test cases in later tasks
	t.Skip("Test framework not yet implemented")
}
```

**Step 3: Run test to verify it compiles**

Run: `go test ./tests/integration -run TestAPIExamplePrograms -v`
Expected: SKIP with "Test framework not yet implemented"

**Step 4: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add API integration test skeleton"
```

---

## Task 2: Create API Test Server Helper

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add createTestServer helper**

Add at bottom of file:

```go
// createTestServer creates a new API server for testing
func createTestServer() *api.Server {
	server := api.NewServer(8080)
	return server
}
```

**Step 2: Add createAPISession helper**

Add after createTestServer:

```go
// createAPISession creates a new session via REST API
func createAPISession(t *testing.T, server *api.Server) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/session",
		bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create session: %d %s", w.Code, w.Body.String())
	}

	var resp api.SessionCreateResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode session response: %v", err)
	}

	return resp.SessionID
}
```

**Step 3: Write test for createAPISession**

Add new test function:

```go
func TestCreateAPISession(t *testing.T) {
	server := createTestServer()
	sessionID := createAPISession(t, server)

	if sessionID == "" {
		t.Fatal("Expected non-empty session ID")
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./tests/integration -run TestCreateAPISession -v`
Expected: PASS

**Step 5: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add API session creation helper"
```

---

## Task 3: Add Program Loading Helper

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add loadProgramViaAPI helper**

Add after createAPISession:

```go
// loadProgramViaAPI loads a program via REST API
func loadProgramViaAPI(t *testing.T, server *api.Server, sessionID, source string) {
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
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode load response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("Program load errors: %v", resp.Errors)
	}
}
```

**Step 2: Write test for loadProgramViaAPI**

Add new test:

```go
func TestLoadProgramViaAPI(t *testing.T) {
	server := createTestServer()
	sessionID := createAPISession(t, server)

	program := `.org 0x8000
main:
	MOV R0, #42
	SWI #0
`
	loadProgramViaAPI(t, server, sessionID, program)
	// If we get here without panic, load succeeded
}
```

**Step 3: Run test to verify it passes**

Run: `go test ./tests/integration -run TestLoadProgramViaAPI -v`
Expected: PASS

**Step 4: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add program loading helper"
```

---

## Task 4: Add Execution and Output Helpers

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add startExecution helper**

Add after loadProgramViaAPI:

```go
// startExecution starts program execution via REST API
func startExecution(t *testing.T, server *api.Server, sessionID string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to start execution: %d %s", w.Code, w.Body.String())
	}
}
```

**Step 2: Add getConsoleOutput helper**

Add after startExecution:

```go
// getConsoleOutput retrieves console output via REST API
func getConsoleOutput(t *testing.T, server *api.Server, sessionID string) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/console", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get console output: %d %s", w.Code, w.Body.String())
	}

	var resp api.ConsoleOutputResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode console response: %v", err)
	}

	return resp.Output
}
```

**Step 3: Add destroySession helper**

Add after getConsoleOutput:

```go
// destroySession destroys a session via REST API
func destroySession(t *testing.T, server *api.Server, sessionID string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/api/v1/session/%s", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	// Don't fail test if session already gone
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Logf("Warning: Failed to destroy session: %d", w.Code)
	}
}
```

**Step 4: Write test for execution flow**

Add new test:

```go
func TestExecutionFlow(t *testing.T) {
	server := createTestServer()
	sessionID := createAPISession(t, server)
	defer destroySession(t, server, sessionID)

	program := `.org 0x8000
	LDR R0, =msg
	SWI #0x02
	SWI #0
msg:
	.asciz "Hello"
`
	loadProgramViaAPI(t, server, sessionID, program)
	startExecution(t, server, sessionID)

	// Wait for execution to complete
	time.Sleep(100 * time.Millisecond)

	output := getConsoleOutput(t, server, sessionID)
	if output != "Hello" {
		t.Errorf("Expected 'Hello', got %q", output)
	}
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./tests/integration -run TestExecutionFlow -v`
Expected: PASS

**Step 6: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add execution and output helpers"
```

---

## Task 5: Add Batch Stdin Helper

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add sendStdinBatch helper**

Add after destroySession:

```go
// sendStdinBatch sends all stdin upfront via REST API
func sendStdinBatch(t *testing.T, server *api.Server, sessionID, stdin string) {
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

**Step 2: Write test for batch stdin**

Add new test:

```go
func TestBatchStdin(t *testing.T) {
	server := createTestServer()
	sessionID := createAPISession(t, server)
	defer destroySession(t, server, sessionID)

	// Load fibonacci.s from examples
	programPath := filepath.Join("..", "..", "examples", "fibonacci.s")
	programBytes, err := os.ReadFile(programPath)
	if err != nil {
		t.Skipf("fibonacci.s not found: %v", err)
	}

	loadProgramViaAPI(t, server, sessionID, string(programBytes))
	sendStdinBatch(t, server, sessionID, "10\n")
	startExecution(t, server, sessionID)

	// Wait for execution
	time.Sleep(200 * time.Millisecond)

	output := getConsoleOutput(t, server, sessionID)
	if !strings.Contains(output, "Fibonacci sequence") {
		t.Errorf("Expected Fibonacci output, got: %q", output)
	}
}
```

**Step 3: Run test to verify it passes**

Run: `go test ./tests/integration -run TestBatchStdin -v`
Expected: PASS (or SKIP if fibonacci.s not found)

**Step 4: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add batch stdin helper"
```

---

## Task 6: Create WebSocket Test Client Structure

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add WebSocket imports**

Update imports section to add:

```go
import (
	// ... existing imports ...
	"sync"

	"github.com/gorilla/websocket"
)
```

**Step 2: Add WebSocketTestClient type**

Add after imports, before test functions:

```go
// WebSocketTestClient manages WebSocket connection for tests
type WebSocketTestClient struct {
	conn    *websocket.Conn
	updates chan StateUpdate
	errors  chan error
	done    chan struct{}
	mu      sync.Mutex
}

// StateUpdate represents a state update from WebSocket
type StateUpdate struct {
	Type      string                 `json:"type"`
	SessionID string                 `json:"session_id"`
	State     string                 `json:"state"`
	Registers map[string]interface{} `json:"registers,omitempty"`
	PC        uint32                 `json:"pc,omitempty"`
	Cycles    int64                  `json:"cycles,omitempty"`
}
```

**Step 3: Run build to verify it compiles**

Run: `go build ./tests/integration`
Expected: Success (or error if gorilla/websocket not imported - install with `go get github.com/gorilla/websocket`)

**Step 4: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add WebSocket test client structure"
```

---

## Task 7: Implement WebSocket Client Connection

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add NewWebSocketTestClient function**

Add after StateUpdate type:

```go
// NewWebSocketTestClient creates a WebSocket test client
func NewWebSocketTestClient(t *testing.T, wsURL string) *WebSocketTestClient {
	t.Helper()

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect WebSocket: %v", err)
	}

	client := &WebSocketTestClient{
		conn:    conn,
		updates: make(chan StateUpdate, 10),
		errors:  make(chan error, 10),
		done:    make(chan struct{}),
	}

	// Start receiving messages
	go client.receiveLoop()

	return client
}
```

**Step 2: Add receiveLoop method**

Add after NewWebSocketTestClient:

```go
// receiveLoop receives WebSocket messages in background
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
```

**Step 3: Add Close method**

Add after receiveLoop:

```go
// Close closes the WebSocket connection
func (c *WebSocketTestClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
		<-c.done // Wait for receive loop to finish
	}
	return nil
}
```

**Step 4: Run build to verify it compiles**

Run: `go build ./tests/integration`
Expected: Success

**Step 5: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: implement WebSocket client connection"
```

---

## Task 8: Add WebSocket Wait Helpers

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add WaitForStateUpdate method**

Add after Close method:

```go
// WaitForStateUpdate waits for next state update with timeout
func (c *WebSocketTestClient) WaitForStateUpdate(timeout time.Duration) (StateUpdate, error) {
	select {
	case update := <-c.updates:
		return update, nil
	case err := <-c.errors:
		return StateUpdate{}, fmt.Errorf("WebSocket error: %w", err)
	case <-time.After(timeout):
		return StateUpdate{}, fmt.Errorf("timeout waiting for state update")
	}
}
```

**Step 2: Add WaitForState method**

Add after WaitForStateUpdate:

```go
// WaitForState waits for specific state with timeout
func (c *WebSocketTestClient) WaitForState(targetState string, timeout time.Duration) (StateUpdate, error) {
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return StateUpdate{}, fmt.Errorf("timeout waiting for state: %s", targetState)
		}

		update, err := c.WaitForStateUpdate(remaining)
		if err != nil {
			return StateUpdate{}, err
		}

		if update.State == targetState {
			return update, nil
		}
	}
}
```

**Step 3: Run build to verify it compiles**

Run: `go build ./tests/integration`
Expected: Success

**Step 4: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add WebSocket wait helpers"
```

---

## Task 9: Create Real HTTP Server for WebSocket Tests

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Update createTestServer to start real server**

Replace createTestServer function:

```go
// createTestServerWithWebSocket creates and starts a real HTTP server for WebSocket testing
func createTestServerWithWebSocket(t *testing.T) (*api.Server, string) {
	t.Helper()

	server := api.NewServer(0) // Port 0 = random available port

	// Start server in background
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Get actual port (TODO: need to expose port from server)
	// For now, use fixed test port
	baseURL := "http://localhost:8080"

	t.Cleanup(func() {
		server.Shutdown(nil)
	})

	return server, baseURL
}

// createTestServer creates test server without WebSocket (for simple REST tests)
func createTestServer() *api.Server {
	return api.NewServer(8080)
}
```

**Step 2: Run build to verify it compiles**

Run: `go build ./tests/integration`
Expected: Success

**Step 3: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add real HTTP server for WebSocket tests"
```

---

## Task 10: Implement Interactive Stdin Handler

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add sendStdinInteractive function**

Add after sendStdinBatch:

```go
// sendStdinInteractive sends stdin incrementally based on WebSocket state
func sendStdinInteractive(t *testing.T, server *api.Server, sessionID, stdin string,
	wsClient *WebSocketTestClient) error {
	t.Helper()

	// Split stdin into lines
	inputs := strings.Split(stdin, "\n")
	// Remove empty trailing line if present
	if len(inputs) > 0 && inputs[len(inputs)-1] == "" {
		inputs = inputs[:len(inputs)-1]
	}

	inputIndex := 0

	// Monitor WebSocket for stdin requests
	for {
		update, err := wsClient.WaitForStateUpdate(5 * time.Second)
		if err != nil {
			return fmt.Errorf("waiting for state update: %w", err)
		}

		// Check if program completed
		if update.State == "halted" {
			break
		}

		// Check if VM waiting for stdin
		if update.State == "waiting_for_input" || update.Type == "stdin_request" {
			if inputIndex >= len(inputs) {
				return fmt.Errorf("program waiting for stdin but all input exhausted")
			}

			// Send next input line
			sendStdinBatch(t, server, sessionID, inputs[inputIndex]+"\n")
			inputIndex++
		}
	}

	return nil
}
```

**Step 2: Run build to verify it compiles**

Run: `go build ./tests/integration`
Expected: Success

**Step 3: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add interactive stdin handler"
```

---

## Task 11: Create Main Test Runner

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add runProgramViaAPI function**

Add after sendStdinInteractive:

```go
// runProgramViaAPI runs a program via REST API with optional stdin
func runProgramViaAPI(t *testing.T, server *api.Server, baseURL, programFile, stdin, stdinMode string) (string, error) {
	t.Helper()

	// Create session
	sessionID := createAPISession(t, server)
	defer destroySession(t, server, sessionID)

	// Establish WebSocket if interactive mode or if we need to wait for halt
	var wsClient *WebSocketTestClient
	if stdinMode == "interactive" || stdinMode == "batch" {
		wsURL := fmt.Sprintf("ws://%s/api/v1/ws?session=%s",
			strings.TrimPrefix(baseURL, "http://"), sessionID)
		wsClient = NewWebSocketTestClient(t, wsURL)
		defer wsClient.Close()
	}

	// Load program
	programPath := filepath.Join("..", "..", "examples", programFile)
	programBytes, err := os.ReadFile(programPath)
	if err != nil {
		return "", fmt.Errorf("reading program: %w", err)
	}
	loadProgramViaAPI(t, server, sessionID, string(programBytes))

	// Handle stdin based on mode
	if stdinMode == "batch" {
		// Send all stdin upfront
		if stdin != "" {
			sendStdinBatch(t, server, sessionID, stdin)
		}
		// Start execution
		startExecution(t, server, sessionID)
		// Wait for halt via WebSocket
		_, err := wsClient.WaitForState("halted", 10*time.Second)
		if err != nil {
			return "", fmt.Errorf("waiting for halt: %w", err)
		}
	} else if stdinMode == "interactive" {
		// Start execution first
		startExecution(t, server, sessionID)
		// Send stdin incrementally
		if err := sendStdinInteractive(t, server, sessionID, stdin, wsClient); err != nil {
			return "", fmt.Errorf("interactive stdin: %w", err)
		}
	} else {
		// No stdin, just run
		startExecution(t, server, sessionID)
		// Wait a bit for execution to complete
		time.Sleep(200 * time.Millisecond)
	}

	// Get console output
	output := getConsoleOutput(t, server, sessionID)
	return output, nil
}
```

**Step 2: Run build to verify it compiles**

Run: `go build ./tests/integration`
Expected: Success

**Step 3: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add main test runner function"
```

---

## Task 12: Add First Test Case (Hello World)

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Update TestAPIExamplePrograms with first test case**

Replace the TestAPIExamplePrograms function:

```go
// TestAPIExamplePrograms runs integration tests for example programs via REST API
func TestAPIExamplePrograms(t *testing.T) {
	// Note: These tests require a real HTTP server for WebSocket support
	// They cannot use httptest.ResponseRecorder

	tests := []struct {
		name           string
		programFile    string
		expectedOutput string
		stdin          string
		stdinMode      string // "batch", "interactive", or ""
	}{
		{
			name:           "Hello_API",
			programFile:    "hello.s",
			expectedOutput: "hello.txt",
			stdinMode:      "",
		},
	}

	server, baseURL := createTestServerWithWebSocket(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runProgramViaAPI(t, server, baseURL,
				tt.programFile, tt.stdin, tt.stdinMode)
			if err != nil {
				t.Fatalf("execution failed: %v", err)
			}

			// Load expected output
			expectedPath := filepath.Join("expected_outputs", tt.expectedOutput)
			expectedBytes, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("failed to read expected output: %v", err)
			}
			expected := string(expectedBytes)

			// Compare output
			if output != expected {
				t.Errorf("output mismatch\nExpected (%d bytes):\n%q\nGot (%d bytes):\n%q",
					len(expected), expected, len(output), output)
			}
		})
	}
}
```

**Step 2: Run test to verify**

Run: `go test ./tests/integration -run TestAPIExamplePrograms/Hello_API -v`
Expected: PASS (or informative failure showing what needs fixing)

**Step 3: Fix any issues and commit**

Debug and fix any failures, then:

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add first API test case (hello.s)"
```

---

## Task 13: Add Fibonacci Test Case (Batch Stdin)

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add fibonacci test case**

Add to tests slice in TestAPIExamplePrograms:

```go
{
	name:           "Fibonacci_API",
	programFile:    "fibonacci.s",
	expectedOutput: "fibonacci.txt",
	stdin:          "10\n",
	stdinMode:      "batch",
},
```

**Step 2: Run test to verify**

Run: `go test ./tests/integration -run TestAPIExamplePrograms/Fibonacci_API -v`
Expected: PASS

**Step 3: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add fibonacci.s API test (batch stdin)"
```

---

## Task 14: Add Calculator Test Case (Interactive Stdin)

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add calculator test case**

Add to tests slice:

```go
{
	name:           "Calculator_API",
	programFile:    "calculator.s",
	expectedOutput: "calculator.txt",
	stdin:          "15\n+\n7\nq\n",
	stdinMode:      "interactive",
},
```

**Step 2: Run test to verify**

Run: `go test ./tests/integration -run TestAPIExamplePrograms/Calculator_API -v`
Expected: PASS (or failure indicating WebSocket state detection needs work)

**Step 3: Debug and fix WebSocket state detection if needed**

If test fails because WebSocket isn't detecting stdin requests properly, check:
- Is the API server broadcasting stdin_request events?
- Is the StateUpdate type correctly unmarshaling the WebSocket messages?

Make necessary fixes and re-test.

**Step 4: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add calculator.s API test (interactive stdin)"
```

---

## Task 15: Add All Remaining Test Cases

**Files:**
- Modify: `tests/integration/api_example_programs_test.go`

**Step 1: Add all 46 remaining test cases**

Add to tests slice (copy from existing `example_programs_test.go`, converting to API format):

```go
{
	name:           "Loops_API",
	programFile:    "loops.s",
	expectedOutput: "loops.txt",
	stdinMode:      "",
},
{
	name:           "MatrixMultiply_API",
	programFile:    "matrix_multiply.s",
	expectedOutput: "matrix_multiply.txt",
	stdinMode:      "",
},
{
	name:           "MemoryStress_API",
	programFile:    "memory_stress.s",
	expectedOutput: "memory_stress.txt",
	stdinMode:      "",
},
{
	name:           "GCD_API",
	programFile:    "gcd.s",
	expectedOutput: "gcd.txt",
	stdin:          "48\n18\n",
	stdinMode:      "batch",
},
{
	name:           "StateMachine_API",
	programFile:    "state_machine.s",
	expectedOutput: "state_machine.txt",
	stdinMode:      "",
},
{
	name:           "StringReverse_API",
	programFile:    "string_reverse.s",
	expectedOutput: "string_reverse.txt",
	stdin:          "Hello World\n",
	stdinMode:      "batch",
},
{
	name:           "Strings_API",
	programFile:    "strings.s",
	expectedOutput: "strings.txt",
	stdinMode:      "",
},
{
	name:           "Stack_API",
	programFile:    "stack.s",
	expectedOutput: "stack.txt",
	stdinMode:      "",
},
{
	name:           "NestedCalls_API",
	programFile:    "nested_calls.s",
	expectedOutput: "nested_calls.txt",
	stdinMode:      "",
},
{
	name:           "HashTable_API",
	programFile:    "hash_table.s",
	expectedOutput: "hash_table.txt",
	stdinMode:      "",
},
{
	name:           "ConstExpressions_API",
	programFile:    "const_expressions.s",
	expectedOutput: "const_expressions.txt",
	stdinMode:      "",
},
{
	name:           "RecursiveFactorial_API",
	programFile:    "recursive_factorial.s",
	expectedOutput: "recursive_factorial.txt",
	stdinMode:      "",
},
{
	name:           "RecursiveFibonacci_API",
	programFile:    "recursive_fib.s",
	expectedOutput: "recursive_fib.txt",
	stdinMode:      "",
},
{
	name:           "SieveOfEratosthenes_API",
	programFile:    "sieve_of_eratosthenes.s",
	expectedOutput: "sieve_of_eratosthenes.txt",
	stdinMode:      "",
},
{
	name:           "StandaloneLabels_API",
	programFile:    "standalone_labels.s",
	expectedOutput: "standalone_labels.txt",
	stdinMode:      "",
},
{
	name:           "XORCipher_API",
	programFile:    "xor_cipher.s",
	expectedOutput: "xor_cipher.txt",
	stdinMode:      "",
},
{
	name:           "FileIO_API",
	programFile:    "file_io.s",
	expectedOutput: "file_io.txt",
	stdinMode:      "",
},
{
	name:           "MultiPrecisionArith_API",
	programFile:    "multi_precision_arith.s",
	expectedOutput: "multi_precision_arith.txt",
	stdinMode:      "",
},
{
	name:           "TaskScheduler_API",
	programFile:    "task_scheduler.s",
	expectedOutput: "task_scheduler.txt",
	stdinMode:      "",
},
{
	name:           "ADRDemo_API",
	programFile:    "adr_demo.s",
	expectedOutput: "adr_demo.txt",
	stdinMode:      "",
},
{
	name:           "TestLtorg_API",
	programFile:    "test_ltorg.s",
	expectedOutput: "test_ltorg.txt",
	stdinMode:      "",
},
{
	name:           "TestOrg0WithLtorg_API",
	programFile:    "test_org_0_with_ltorg.s",
	expectedOutput: "test_org_0_with_ltorg.txt",
	stdinMode:      "",
},
{
	name:           "LargeLiteralPool_API",
	programFile:    "large_literal_pool.s",
	expectedOutput: "large_literal_pool.txt",
	stdinMode:      "",
},
{
	name:           "NOPTest_API",
	programFile:    "nop_test.s",
	expectedOutput: "nop_test.txt",
	stdinMode:      "",
},
{
	name:           "CelsiusToFahrenheit_0_API",
	programFile:    "celsius_to_fahrenheit.s",
	expectedOutput: "celsius_to_fahrenheit_0.txt",
	stdin:          "0\n",
	stdinMode:      "batch",
},
{
	name:           "CelsiusToFahrenheit_25_API",
	programFile:    "celsius_to_fahrenheit.s",
	expectedOutput: "celsius_to_fahrenheit_25.txt",
	stdin:          "25\n",
	stdinMode:      "batch",
},
{
	name:           "CelsiusToFahrenheit_100_API",
	programFile:    "celsius_to_fahrenheit.s",
	expectedOutput: "celsius_to_fahrenheit_100.txt",
	stdin:          "100\n",
	stdinMode:      "batch",
},
{
	name:           "AddressingModes_API",
	programFile:    "addressing_modes.s",
	expectedOutput: "addressing_modes.txt",
	stdinMode:      "",
},
{
	name:           "Arithmetic_API",
	programFile:    "arithmetic.s",
	expectedOutput: "arithmetic.txt",
	stdinMode:      "",
},
{
	name:           "Add128Bit_API",
	programFile:    "add_128bit.s",
	expectedOutput: "add_128bit.txt",
	stdinMode:      "",
},
{
	name:           "Arrays_API",
	programFile:    "arrays.s",
	expectedOutput: "arrays.txt",
	stdinMode:      "",
},
{
	name:           "BinarySearch_API",
	programFile:    "binary_search.s",
	expectedOutput: "binary_search.txt",
	stdinMode:      "",
},
{
	name:           "BitOperations_API",
	programFile:    "bit_operations.s",
	expectedOutput: "bit_operations.txt",
	stdinMode:      "",
},
{
	name:           "BubbleSort_API",
	programFile:    "bubble_sort.s",
	expectedOutput: "bubble_sort.txt",
	stdin:          "7\n5\n1\n4\n2\n8\n3\n6\n",
	stdinMode:      "batch",
},
{
	name:           "Conditionals_API",
	programFile:    "conditionals.s",
	expectedOutput: "conditionals.txt",
	stdinMode:      "",
},
{
	name:           "Division_API",
	programFile:    "division.s",
	expectedOutput: "division.txt",
	stdinMode:      "",
},
{
	name:           "Factorial_API",
	programFile:    "factorial.s",
	expectedOutput: "factorial.txt",
	stdin:          "5\n",
	stdinMode:      "batch",
},
{
	name:           "Functions_API",
	programFile:    "functions.s",
	expectedOutput: "functions.txt",
	stdinMode:      "",
},
{
	name:           "LinkedList_API",
	programFile:    "linked_list.s",
	expectedOutput: "linked_list.txt",
	stdinMode:      "",
},
{
	name:           "Quicksort_API",
	programFile:    "quicksort.s",
	expectedOutput: "quicksort.txt",
	stdinMode:      "",
},
{
	name:           "TimesTable_API",
	programFile:    "times_table.s",
	expectedOutput: "times_table.txt",
	stdin:          "7\n",
	stdinMode:      "batch",
},
{
	name:           "TestConstExpr_API",
	programFile:    "test_const_expr.s",
	expectedOutput: "test_const_expr.txt",
	stdinMode:      "",
},
{
	name:           "TestExpr_API",
	programFile:    "test_expr.s",
	expectedOutput: "test_expr.txt",
	stdinMode:      "",
},
{
	name:           "TestGetArguments_API",
	programFile:    "test_get_arguments.s",
	expectedOutput: "test_get_arguments.txt",
	stdinMode:      "",
},
```

**Step 2: Run all tests to verify**

Run: `go test ./tests/integration -run TestAPIExamplePrograms -v`
Expected: All tests PASS

**Step 3: Commit**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: add all 49 API integration test cases"
```

---

## Task 16: Update CLAUDE.md Documentation

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Add API integration test section**

Find the test section and add:

```markdown
## API Integration Tests

API-level integration tests validate the complete HTTP REST API + WebSocket stack:

```bash
# Run all API integration tests
go test ./tests/integration -run TestAPIExamplePrograms -v

# Run specific API test
go test ./tests/integration -run TestAPIExamplePrograms/Fibonacci_API -v

# Run with race detector
go test -race ./tests/integration -run TestAPIExamplePrograms
```

**Test Coverage:**
- All 49 example programs tested via REST API
- WebSocket state monitoring for execution tracking
- Hybrid stdin strategy (batch + interactive)
- Reuses existing `expected_outputs/*.txt` files

**Note:** API tests start a real HTTP server for WebSocket support (cannot use httptest.ResponseRecorder).
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add API integration tests documentation"
```

---

## Task 17: Run Full Test Suite and Verify

**Files:**
- None (verification only)

**Step 1: Clear test cache**

Run: `go clean -testcache`

**Step 2: Run all tests**

Run: `go test ./...`
Expected: All tests PASS including new API integration tests

**Step 3: Run with race detector**

Run: `go test -race ./tests/integration -run TestAPIExamplePrograms`
Expected: No race conditions detected

**Step 4: Run formatting and linting**

Run: `go fmt ./... && golangci-lint run ./...`
Expected: No issues

**Step 5: Commit if any fixes were needed**

If any issues found and fixed:

```bash
git add .
git commit -m "fix: resolve test suite issues"
```

---

## Task 18: Final Integration and Cleanup

**Files:**
- Various (cleanup)

**Step 1: Remove temporary test functions**

Remove these test functions that were used for incremental development:
- `TestCreateAPISession`
- `TestLoadProgramViaAPI`
- `TestExecutionFlow`
- `TestBatchStdin`

They've served their purpose validating individual helpers.

**Step 2: Run tests to ensure removal didn't break anything**

Run: `go test ./tests/integration -v`
Expected: All tests still PASS

**Step 3: Commit cleanup**

```bash
git add tests/integration/api_example_programs_test.go
git commit -m "test: remove temporary helper test functions"
```

**Step 4: Create final commit**

```bash
git commit --allow-empty -m "feat: complete API integration tests

- 49 programs tested via REST API + WebSocket
- Hybrid stdin strategy (batch + interactive)
- Comprehensive state monitoring
- Full stack validation"
```

---

## Completion Checklist

- [ ] Task 1: Test file skeleton created
- [ ] Task 2: API session helper implemented
- [ ] Task 3: Program loading helper implemented
- [ ] Task 4: Execution and output helpers implemented
- [ ] Task 5: Batch stdin helper implemented
- [ ] Task 6: WebSocket client structure added
- [ ] Task 7: WebSocket connection implemented
- [ ] Task 8: WebSocket wait helpers implemented
- [ ] Task 9: Real HTTP server for WebSocket created
- [ ] Task 10: Interactive stdin handler implemented
- [ ] Task 11: Main test runner implemented
- [ ] Task 12: First test case (hello.s) added
- [ ] Task 13: Fibonacci test case added
- [ ] Task 14: Calculator test case added
- [ ] Task 15: All 49 test cases added
- [ ] Task 16: Documentation updated
- [ ] Task 17: Full test suite verified
- [ ] Task 18: Cleanup completed

---

## Success Criteria

- [ ] All 49 example programs pass via API tests
- [ ] No race conditions detected
- [ ] Zero linting issues
- [ ] Documentation updated in CLAUDE.md
- [ ] Test execution time reasonable (<2 minutes for full suite)
- [ ] WebSocket state monitoring working correctly
- [ ] Interactive stdin programs (calculator.s) working

---

## Notes

- **WebSocket Limitation:** Tests require real HTTP server, cannot use httptest.ResponseRecorder
- **State Detection:** If WebSocket state updates don't include "waiting_for_input", may need to modify API server to broadcast this state
- **Timing:** Some tests may need adjusted timeouts based on system performance
- **Parallelization:** Consider adding `t.Parallel()` after verifying all tests pass serially
