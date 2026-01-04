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
	_ = time.Now
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
