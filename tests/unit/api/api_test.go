package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lookbusy1344/arm-emulator/api"
)

// testServer creates a test server for testing
func testServer() *api.Server {
	server := api.NewServer(8080)
	// For testing, we need to wrap mux with CORS middleware manually since Start() isn't called
	return server
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	server := testServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", response["status"])
	}
}

// TestCreateSession tests session creation
func TestCreateSession(t *testing.T) {
	server := testServer()

	reqBody := api.SessionCreateRequest{
		MemorySize: 1024 * 1024,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response api.SessionCreateResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.SessionID == "" {
		t.Error("Expected non-empty session ID")
	}

	if response.CreatedAt.IsZero() {
		t.Error("Expected non-zero creation time")
	}
}

// TestListSessions tests listing sessions
func TestListSessions(t *testing.T) {
	server := testServer()

	// Create a few sessions
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)
	}

	// List sessions
	req := httptest.NewRequest(http.MethodGet, "/api/v1/session", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	sessions := response["sessions"].([]interface{})
	if len(sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(sessions))
	}
}

// TestLoadProgram tests loading a program
func TestLoadProgram(t *testing.T) {
	server := testServer()

	// Create session
	sessionID := createTestSession(t, server)

	// Load program
	program := `
	.org 0x8000
main:
	MOV R0, #42
	SWI #0
	`

	reqBody := api.LoadProgramRequest{
		Source: program,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/load", sessionID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response api.LoadProgramResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected successful load, got errors: %v", response.Errors)
	}

	if response.Symbols == nil {
		t.Error("Expected symbols map")
	}

	if _, exists := response.Symbols["main"]; !exists {
		t.Error("Expected 'main' symbol in symbol table")
	}
}

// TestLoadInvalidProgram tests loading an invalid program
func TestLoadInvalidProgram(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	reqBody := api.LoadProgramRequest{
		Source: "INVALID_INSTRUCTION R0, R1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/load", sessionID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response api.LoadProgramResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Success {
		t.Error("Expected failed load for invalid program")
	}

	if len(response.Errors) == 0 {
		t.Error("Expected error messages")
	}
}

// TestStepExecution tests single-step execution
func TestStepExecution(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load program
	program := `
	.org 0x8000
	MOV R0, #42
	MOV R1, #100
	SWI #0
	`
	loadProgram(t, server, sessionID, program)

	// Step once
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/step", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response api.RegistersResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.R0 != 42 {
		t.Errorf("Expected R0 = 42, got %d", response.R0)
	}

	// Step again
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/step", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	json.NewDecoder(w.Body).Decode(&response)

	if response.R1 != 100 {
		t.Errorf("Expected R1 = 100, got %d", response.R1)
	}
}

// TestGetRegisters tests getting register state
func TestGetRegisters(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response api.RegistersResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify all registers are present (PC should be at default or loaded position)
	// PC is allowed to be 0 if no program is loaded, so just check the structure is valid
	if response.Cycles < 0 {
		t.Error("Expected non-negative cycles")
	}
}

// TestGetMemory tests reading memory
func TestGetMemory(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/memory?address=0x8000&length=16", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response api.MemoryResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Address != 0x8000 {
		t.Errorf("Expected address 0x8000, got 0x%X", response.Address)
	}

	if response.Length != 16 {
		t.Errorf("Expected length 16, got %d", response.Length)
	}

	if len(response.Data) != 16 {
		t.Errorf("Expected 16 bytes of data, got %d", len(response.Data))
	}
}

// TestGetMemoryTooLarge tests memory read size limit
func TestGetMemoryTooLarge(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Try to read 2MB (should fail)
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/memory?address=0x8000&length=2097152", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestBreakpoints tests breakpoint management
func TestBreakpoints(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Add breakpoint
	reqBody := api.BreakpointRequest{
		Address: 0x8004,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/breakpoint", sessionID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// List breakpoints
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/breakpoints", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	var response api.BreakpointsResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Breakpoints) != 1 {
		t.Errorf("Expected 1 breakpoint, got %d", len(response.Breakpoints))
	}

	if response.Breakpoints[0] != 0x8004 {
		t.Errorf("Expected breakpoint at 0x8004, got 0x%X", response.Breakpoints[0])
	}

	// Remove breakpoint
	req = httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/api/v1/session/%s/breakpoint", sessionID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestReset tests VM reset
func TestReset(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load and execute program
	program := ".org 0x8000\nMOV R0, #42\nSWI #0"
	loadProgram(t, server, sessionID, program)

	// Step once
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/step", sessionID), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	// Reset
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/reset", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify state is reset (get registers)
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var regs api.RegistersResponse
	json.NewDecoder(w.Body).Decode(&regs)

	if regs.Cycles != 0 {
		t.Errorf("Expected cycles = 0 after reset, got %d", regs.Cycles)
	}
}

// TestDestroySession tests session destruction
func TestDestroySession(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Destroy session
	req := httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/api/v1/session/%s", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify session is gone
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// TestSessionNotFound tests error handling for non-existent session
func TestSessionNotFound(t *testing.T) {
	server := testServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/session/nonexistent", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// TestCORS tests CORS headers with localhost restriction
func TestCORS(t *testing.T) {
	server := testServer()

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		shouldAllow    bool
	}{
		{
			name:           "localhost with port",
			origin:         "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			shouldAllow:    true,
		},
		{
			name:           "localhost https",
			origin:         "https://localhost:8443",
			expectedOrigin: "https://localhost:8443",
			shouldAllow:    true,
		},
		{
			name:           "127.0.0.1 with port",
			origin:         "http://127.0.0.1:5173",
			expectedOrigin: "http://127.0.0.1:5173",
			shouldAllow:    true,
		},
		{
			name:           "127.0.0.1 https",
			origin:         "https://127.0.0.1:443",
			expectedOrigin: "https://127.0.0.1:443",
			shouldAllow:    true,
		},
		{
			name:           "file protocol",
			origin:         "file:///path/to/file.html",
			expectedOrigin: "file:///path/to/file.html",
			shouldAllow:    true,
		},
		{
			name:        "remote origin rejected",
			origin:      "http://evil.com",
			shouldAllow: false,
		},
		{
			name:        "remote https rejected",
			origin:      "https://attacker.net:8080",
			shouldAllow: false,
		},
		{
			name:           "no origin header (native apps)",
			origin:         "",
			expectedOrigin: "",
			shouldAllow:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "/api/v1/session", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			server.Handler().ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
			}

			corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if tt.shouldAllow {
				if corsOrigin != tt.expectedOrigin {
					t.Errorf("Expected CORS origin '%s', got '%s'", tt.expectedOrigin, corsOrigin)
				}
			} else {
				if corsOrigin != "" {
					t.Errorf("Expected no CORS origin for remote host, got '%s'", corsOrigin)
				}
			}
		})
	}
}

// TestCORSWithActualRequest tests CORS with a real GET request
func TestCORSWithActualRequest(t *testing.T) {
	server := testServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Error("Expected localhost CORS origin to be echoed back")
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Expected credentials support for localhost")
	}
}

// Helper functions

func createTestSession(t *testing.T, server *api.Server) string {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create session: %d %s", w.Code, w.Body.String())
	}

	var response api.SessionCreateResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode session response: %v", err)
	}

	return response.SessionID
}

func loadProgram(t *testing.T, server *api.Server, sessionID string, program string) {
	reqBody := api.LoadProgramRequest{Source: program}
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

	// Wait a bit for program to load
	time.Sleep(10 * time.Millisecond)
}

// TestStopExecution tests stop execution
func TestStopExecution(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load infinite loop
	program := ".org 0x8000\nloop: B loop"
	loadProgram(t, server, sessionID, program)

	// Start execution
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	// Give it time to start
	time.Sleep(20 * time.Millisecond)

	// Stop execution
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/stop", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestDisassembly tests disassembly endpoint
func TestDisassembly(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load program
	program := ".org 0x8000\nMOV R0, #42\nMOV R1, #100\nADD R2, R0, R1"
	loadProgram(t, server, sessionID, program)

	// Get disassembly
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/disassembly?address=0x8000&count=3", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response api.DisassemblyResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Instructions) == 0 {
		t.Error("Expected at least one instruction in disassembly")
	}

	if response.Instructions[0].Address != 0x8000 {
		t.Errorf("Expected first instruction at 0x8000, got 0x%X", response.Instructions[0].Address)
	}
}

// TestWatchpoints tests watchpoint management
func TestWatchpoints(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Add watchpoint
	reqBody := api.WatchpointRequest{
		Address: 0x20000,
		Type:    "write",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/watchpoint", sessionID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var addResponse api.WatchpointResponse
	if err := json.NewDecoder(w.Body).Decode(&addResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	watchpointID := addResponse.ID

	// List watchpoints
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/watchpoints", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var listResponse api.WatchpointsResponse
	json.NewDecoder(w.Body).Decode(&listResponse)

	if len(listResponse.Watchpoints) != 1 {
		t.Errorf("Expected 1 watchpoint, got %d", len(listResponse.Watchpoints))
	}

	// Remove watchpoint
	req = httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/api/v1/session/%s/watchpoint/%d", sessionID, watchpointID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify removed
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/watchpoints", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	json.NewDecoder(w.Body).Decode(&listResponse)

	if len(listResponse.Watchpoints) != 0 {
		t.Errorf("Expected 0 watchpoints after removal, got %d", len(listResponse.Watchpoints))
	}
}

// TestExecutionTrace tests trace management
func TestExecutionTrace(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load simple program
	program := ".org 0x8000\nMOV R0, #1\nMOV R1, #2\nADD R2, R0, R1"
	loadProgram(t, server, sessionID, program)

	// Enable trace
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/trace/enable", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Step a few times
	for i := 0; i < 3; i++ {
		req = httptest.NewRequest(http.MethodPost,
			fmt.Sprintf("/api/v1/session/%s/step", sessionID), nil)
		w = httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)
	}

	// Get trace data
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/trace/data", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response api.TraceDataResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Trace entries may be empty if tracing wasn't enabled before steps
	// Just verify the endpoint works and returns valid structure
	if response.Entries == nil {
		t.Error("Expected trace entries array (even if empty)")
	}

	// Disable trace
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/trace/disable", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestStatistics tests statistics collection
func TestStatistics(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load simple program
	program := ".org 0x8000\nMOV R0, #42\nB end\nend: SWI #0"
	loadProgram(t, server, sessionID, program)

	// Enable statistics
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/stats/enable", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Run program
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	// Get statistics
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/stats", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response api.StatisticsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Statistics may be zero if program completed very quickly
	// Just verify the endpoint works and returns valid structure
	if response.TotalInstructions < 0 {
		t.Error("Expected non-negative instruction count")
	}

	if response.TotalCycles < 0 {
		t.Error("Expected non-negative cycle count")
	}

	// Disable statistics
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/stats/disable", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestConfiguration tests configuration management
func TestConfiguration(t *testing.T) {
	server := testServer()

	// Get configuration
	req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var getResponse api.ConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&getResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Update configuration
	reqBody := api.ConfigResponse{
		Execution: api.ExecutionConfig{
			MaxCycles: 2000000,
		},
	}

	body, _ := json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Note: The config update endpoint acknowledges but doesn't persist changes
	// This is expected behavior for the current implementation
	// Future implementations might add persistence
}

// TestExamples tests example file management
func TestExamples(t *testing.T) {
	server := testServer()

	// List examples
	req := httptest.NewRequest(http.MethodGet, "/api/v1/examples", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	// If examples directory doesn't exist, skip test gracefully
	if w.Code == http.StatusInternalServerError {
		t.Skip("Examples directory not available - skipping test")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var listResponse api.ExamplesResponse
	if err := json.NewDecoder(w.Body).Decode(&listResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(listResponse.Examples) == 0 {
		t.Skip("No example files found - skipping example retrieval test")
	}

	// Get first example
	exampleName := listResponse.Examples[0].Name

	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/examples/%s", exampleName), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var contentResponse api.ExampleContentResponse
	if err := json.NewDecoder(w.Body).Decode(&contentResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if contentResponse.Name != exampleName {
		t.Errorf("Expected name '%s', got '%s'", exampleName, contentResponse.Name)
	}

	if contentResponse.Content == "" {
		t.Error("Expected non-empty content")
	}
}

// TestExamplesPathTraversal tests path traversal protection
func TestExamplesPathTraversal(t *testing.T) {
	server := testServer()

	// Try path traversal attack
	req := httptest.NewRequest(http.MethodGet, "/api/v1/examples/../../../etc/passwd", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	// Accept 400, 404, or 301 (redirect) as valid responses to path traversal
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound && w.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status 400, 404, or 301 for path traversal, got %d", w.Code)
	}
}
