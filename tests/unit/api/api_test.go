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

// TestCORS tests CORS headers
func TestCORS(t *testing.T) {
	server := testServer()

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/session", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS headers")
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
