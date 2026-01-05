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
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lookbusy1344/arm-emulator/api"
)

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
	SessionID string                 `json:"sessionId"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// GetStatus extracts the status from the nested data structure
func (s *StateUpdate) GetStatus() string {
	if s.Data != nil {
		if status, ok := s.Data["status"].(string); ok {
			return status
		}
	}
	return ""
}

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

	// Extract session ID from URL query parameter
	// wsURL format: ws://host/api/v1/ws?session=SESSION_ID
	sessionID := ""
	if idx := strings.Index(wsURL, "session="); idx != -1 {
		sessionID = wsURL[idx+8:]
	}

	// Send subscription request to receive all events for this session
	if sessionID != "" {
		subReq := map[string]interface{}{
			"type":      "subscribe",
			"sessionId": sessionID,
			"events":    []string{}, // Empty = all events
		}
		if err := conn.WriteJSON(subReq); err != nil {
			t.Fatalf("Failed to send subscription: %v", err)
		}
		// Give subscription time to register
		time.Sleep(50 * time.Millisecond)
	}

	return client
}

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

// WaitForState waits for specific state value with timeout
func (c *WebSocketTestClient) WaitForState(targetState string, timeout time.Duration) (StateUpdate, error) {
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return StateUpdate{}, fmt.Errorf("timeout waiting for state %q", targetState)
		}

		update, err := c.WaitForStateUpdate(remaining)
		if err != nil {
			return StateUpdate{}, err
		}

		if update.GetStatus() == targetState {
			return update, nil
		}
		// Continue looping to wait for the target state
	}
}

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
		skip           bool   // skip test if true
	}{
		{
			name:           "Hello_API",
			programFile:    "hello.s",
			expectedOutput: "hello.txt",
			stdinMode:      "",
		},
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
			expectedOutput: "calculator_interactive.txt", // Interactive mode echoes input
			stdin:          "15\n+\n7\n0\nq\n",           // Need 5 inputs: num1, op, num2, (dummy)num1, quit-op
			stdinMode:      "interactive",
		},
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
	}

	server, baseURL := createTestServerWithWebSocket(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Test skipped - see caveats documentation")
			}

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

// createTestServerWithWebSocket creates and starts a real HTTP server for WebSocket testing
func createTestServerWithWebSocket(t *testing.T) (*api.Server, string) {
	t.Helper()

	server := api.NewServer(8080)
	testServer := httptest.NewServer(server.Handler())

	t.Cleanup(func() {
		testServer.Close()
	})

	return server, testServer.URL
}

// createTestServer creates test server without WebSocket (for simple REST tests)
func createTestServer() *api.Server {
	return api.NewServer(8080)
}

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

// loadProgramViaAPI loads a program via REST API
func loadProgramViaAPI(t *testing.T, server *api.Server, sessionID, source string) {
	t.Helper()

	reqBody := api.LoadProgramRequest{Source: source}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

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

// sendStdinBatch sends all stdin upfront via REST API
func sendStdinBatch(t *testing.T, server *api.Server, sessionID, stdin string) {
	t.Helper()

	reqBody := api.StdinRequest{Data: stdin}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal stdin request: %v", err)
	}

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
		// Wait for state update
		update, err := wsClient.WaitForStateUpdate(5 * time.Second)
		if err != nil {
			return fmt.Errorf("waiting for state update: %w", err)
		}

		// Check if program completed
		status := update.GetStatus()
		if status == "halted" || status == "error" {
			// Program halted - this is success even if inputs remain unused
			break
		}

		// Check if VM is waiting for input
		if status == "waiting_for_input" || update.Type == "stdin_request" {
			if inputIndex >= len(inputs) {
				return fmt.Errorf("program requested more input than provided")
			}

			// Send next input line
			sendStdinBatch(t, server, sessionID, inputs[inputIndex]+"\n")
			inputIndex++
		}
		// Continue monitoring
	}

	return nil
}

// runProgramViaAPI runs a program via REST API with optional stdin
func runProgramViaAPI(t *testing.T, server *api.Server, baseURL, programFile, stdin, stdinMode string) (string, error) {
	t.Helper()

	// Create session
	sessionID := createAPISession(t, server)
	defer destroySession(t, server, sessionID)

	// Establish WebSocket if interactive mode or if we need to wait for halt
	var wsClient *WebSocketTestClient
	if stdinMode == "interactive" || stdinMode == "batch" {
		wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + fmt.Sprintf("/api/v1/ws?session=%s", sessionID)
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

func TestCreateAPISession(t *testing.T) {
	server := createTestServer()
	sessionID := createAPISession(t, server)

	if sessionID == "" {
		t.Fatal("Expected non-empty session ID")
	}
}

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
