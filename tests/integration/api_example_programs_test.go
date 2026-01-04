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
	SessionID string                 `json:"session_id"`
	State     string                 `json:"state"`
	Registers map[string]interface{} `json:"registers,omitempty"`
	PC        uint32                 `json:"pc,omitempty"`
	Cycles    int64                  `json:"cycles,omitempty"`
}

// TestAPIExamplePrograms runs integration tests for example programs via REST API
func TestAPIExamplePrograms(t *testing.T) {
	// Temporary usage to satisfy Go's unused import check
	// These will be used in subsequent tasks
	_ = bytes.Buffer{}
	_ = json.Marshal
	_ = fmt.Sprint
	_ = http.StatusOK
	_ = httptest.NewServer
	_ = os.Open
	_ = filepath.Join
	_ = strings.Join
	_ = sync.Mutex{}
	_ = time.Now
	_ = (*websocket.Conn)(nil)
	_ = api.NewServer

	// Placeholder - will add test cases in later tasks
	t.Skip("Test framework not yet implemented")
}

// createTestServer creates a new API server for simple REST tests.
// Note: Does not take *testing.T parameter since it cannot call t.Fatal.
// For WebSocket tests, use createTestServerWithWebSocket (Task 9) instead.
func createTestServer() *api.Server {
	server := api.NewServer(8080)
	return server
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
