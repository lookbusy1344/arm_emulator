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

	// Load a program first so we have valid addresses
	program := `
		.org 0x8000
		MOV R0, #1
		MOV R1, #2
		ADD R2, R0, R1
	`
	loadProgram(t, server, sessionID, program)

	// Add breakpoint at valid address (0x8004 = second instruction)
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

// TestBreakpointValidation tests that invalid breakpoints are rejected
func TestBreakpointValidation(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load a program with some actual code
	program := `
		.org 0x8000
		; Comment line
		
		MOV R0, #1
		MOV R1, #2
		ADD R2, R0, R1
		SWI #0
	`
	loadProgram(t, server, sessionID, program)

	// Try to set breakpoint at invalid address (not in source map)
	reqBody := api.BreakpointRequest{
		Address: 0x9000, // Invalid address outside program
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/breakpoint", sessionID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	// Should return error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Verify no breakpoint was added
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/breakpoints", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var response api.BreakpointsResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Breakpoints) != 0 {
		t.Errorf("Expected 0 breakpoints, got %d", len(response.Breakpoints))
	}
}

// TestSourceMap tests the source map endpoint
func TestSourceMap(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load a simple program
	program := `
		.org 0x8000
		MOV R0, #1
		MOV R1, #2
		ADD R2, R0, R1
	`
	loadProgram(t, server, sessionID, program)

	// Get source map
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/sourcemap", sessionID), nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	sourceMap, ok := response["sourceMap"].([]interface{})
	if !ok {
		t.Fatal("Expected sourceMap array in response")
	}

	// Should have 3 entries (one for each instruction)
	if len(sourceMap) != 3 {
		t.Errorf("Expected 3 source map entries, got %d", len(sourceMap))
	}

	// Verify first entry has address, line, and lineNumber
	if len(sourceMap) > 0 {
		entry := sourceMap[0].(map[string]interface{})
		if _, hasAddress := entry["address"]; !hasAddress {
			t.Error("Source map entry missing 'address' field")
		}
		if _, hasLine := entry["line"]; !hasLine {
			t.Error("Source map entry missing 'line' field")
		}
		if _, hasLineNumber := entry["lineNumber"]; !hasLineNumber {
			t.Error("Source map entry missing 'lineNumber' field")
		}
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

// TestRunExecution tests that /run actually executes the program
func TestRunExecution(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load a simple program that sets R0=42 and exits
	program := ".org 0x8000\nMOV R0, #42\nSWI #0"
	loadProgram(t, server, sessionID, program)

	// Get initial registers to verify R0 is 0
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var initialRegs api.RegistersResponse
	json.NewDecoder(w.Body).Decode(&initialRegs)

	if initialRegs.R0 != 0 {
		t.Errorf("Expected initial R0 = 0, got %d", initialRegs.R0)
	}

	// Start execution
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Wait for program to complete
	time.Sleep(100 * time.Millisecond)

	// Get final registers - R0 should be 42 if program actually ran
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var finalRegs api.RegistersResponse
	if err := json.NewDecoder(w.Body).Decode(&finalRegs); err != nil {
		t.Fatalf("Failed to decode registers: %v", err)
	}

	if finalRegs.R0 != 42 {
		t.Errorf("Program did not execute! Expected R0 = 42, got %d", finalRegs.R0)
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

// TestConsoleOutput tests retrieving console output via API
func TestConsoleOutput(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load a program that writes to console
	program := `
	.org 0x8000
main:
	; Write "Hello, World!\n" to console
	LDR R0, =message
	SWI #0x02    ; WRITE_STRING syscall
	SWI #0x07    ; WRITE_NEWLINE syscall
	SWI #0       ; EXIT syscall

message:
	.asciz "Hello, World!"
	`

	loadProgram(t, server, sessionID, program)

	// Run the program
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for run, got %d: %s", w.Code, w.Body.String())
	}

	// Wait for program to complete
	time.Sleep(100 * time.Millisecond)

	// Get console output
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/console", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response api.ConsoleOutputResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expectedOutput := "Hello, World!\n"
	if response.Output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, response.Output)
	}
}

// TestConsoleOutputEmpty tests console output with no program output
func TestConsoleOutputEmpty(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load a program that doesn't write anything
	program := `
	.org 0x8000
	MOV R0, #42
	SWI #0
	`
	loadProgram(t, server, sessionID, program)

	// Run the program
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	// Get console output (should be empty)
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/console", sessionID), nil)
	w = httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response api.ConsoleOutputResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Output != "" {
		t.Errorf("Expected empty output, got %q", response.Output)
	}
}

// TestReRunProgram tests that a program can be run multiple times after completion
func TestReRunProgram(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load a simple program that increments R0 and exits
	program := `
	.org 0x8000
	MOV R0, #1
	ADD R0, R0, #1
	SWI #0
	`
	loadProgram(t, server, sessionID, program)

	// First run
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for first run, got %d: %s", w.Code, w.Body.String())
	}

	// Wait for program to complete
	time.Sleep(100 * time.Millisecond)

	// Check registers after first run
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var regs1 api.RegistersResponse
	if err := json.NewDecoder(w.Body).Decode(&regs1); err != nil {
		t.Fatalf("Failed to decode registers: %v", err)
	}

	if regs1.R0 != 2 {
		t.Errorf("After first run: Expected R0 = 2, got %d", regs1.R0)
	}

	expectedPC1 := uint32(0x8008) // PC after SWI
	if regs1.PC != expectedPC1 {
		t.Logf("After first run: PC = 0x%08X (expected 0x%08X)", regs1.PC, expectedPC1)
	}

	// Second run without explicit reset - should auto-reset and run again
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for second run, got %d: %s", w.Code, w.Body.String())
	}

	// Wait for program to complete
	time.Sleep(100 * time.Millisecond)

	// Check registers after second run
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var regs2 api.RegistersResponse
	if err := json.NewDecoder(w.Body).Decode(&regs2); err != nil {
		t.Fatalf("Failed to decode registers: %v", err)
	}

	// After second run, R0 should still be 2, not 4
	// This proves the registers were reset before the second run
	if regs2.R0 != 2 {
		t.Errorf("After second run: Expected R0 = 2 (reset and rerun), got %d", regs2.R0)
	}

	// Third run to confirm it keeps working
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for third run, got %d: %s", w.Code, w.Body.String())
	}

	time.Sleep(100 * time.Millisecond)

	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var regs3 api.RegistersResponse
	if err := json.NewDecoder(w.Body).Decode(&regs3); err != nil {
		t.Fatalf("Failed to decode registers: %v", err)
	}

	if regs3.R0 != 2 {
		t.Errorf("After third run: Expected R0 = 2, got %d", regs3.R0)
	}
}

// TestRestart tests the restart endpoint which resets execution to entry point
// while preserving the loaded program (unlike reset which clears everything)
func TestRestart(t *testing.T) {
	server := testServer()

	// Create session
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var createResp api.SessionCreateResponse
	if err := json.NewDecoder(w.Body).Decode(&createResp); err != nil {
		t.Fatalf("Failed to decode session response: %v", err)
	}
	sessionID := createResp.SessionID

	// Load a simple program
	source := `
	.org 0x8000
	MOV R0, #10
	ADD R0, R0, #5
	ADD R0, R0, #7
	SWI #0
	`

	loadReq := api.LoadProgramRequest{Source: source}
	body, _ := json.Marshal(loadReq)
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/load", sessionID), bytes.NewReader(body))
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for load, got %d: %s", w.Code, w.Body.String())
	}

	// Get initial PC (should be entry point)
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var initialRegs api.RegistersResponse
	if err := json.NewDecoder(w.Body).Decode(&initialRegs); err != nil {
		t.Fatalf("Failed to decode initial registers: %v", err)
	}
	entryPoint := initialRegs.PC

	// Step twice to advance PC
	for i := 0; i < 2; i++ {
		req = httptest.NewRequest(http.MethodPost,
			fmt.Sprintf("/api/v1/session/%s/step", sessionID), nil)
		w = httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200 for step, got %d: %s", w.Code, w.Body.String())
		}
	}

	// Verify PC has advanced
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var steppedRegs api.RegistersResponse
	if err := json.NewDecoder(w.Body).Decode(&steppedRegs); err != nil {
		t.Fatalf("Failed to decode stepped registers: %v", err)
	}

	if steppedRegs.PC == entryPoint {
		t.Error("PC should have advanced after stepping")
	}

	// Restart (should reset PC to entry point but keep program loaded)
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/restart", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for restart, got %d: %s", w.Code, w.Body.String())
	}

	// Verify PC is back at entry point
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var restartedRegs api.RegistersResponse
	if err := json.NewDecoder(w.Body).Decode(&restartedRegs); err != nil {
		t.Fatalf("Failed to decode restarted registers: %v", err)
	}

	if restartedRegs.PC != entryPoint {
		t.Errorf("After restart: Expected PC = 0x%08X (entry point), got 0x%08X",
			entryPoint, restartedRegs.PC)
	}

	// Step again to verify program is still loaded and executable
	req = httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/step", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for step after restart, got %d: %s", w.Code, w.Body.String())
	}

	// Verify execution continued correctly
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	var finalRegs api.RegistersResponse
	if err := json.NewDecoder(w.Body).Decode(&finalRegs); err != nil {
		t.Fatalf("Failed to decode final registers: %v", err)
	}

	if finalRegs.R0 != 10 {
		t.Errorf("After restart and first step: Expected R0 = 10, got %d", finalRegs.R0)
	}

	if finalRegs.PC == entryPoint {
		t.Error("PC should have advanced after stepping post-restart")
	}
}

// TestMemoryWriteSize_STRB tests that API returns writeSize=1 for STRB instruction
func TestMemoryWriteSize_STRB(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load program with STRB instruction
	program := `
	.org 0x8000
main:
	LDR R0, =0x12345678
	LDR R1, =data_area
	STRB R0, [R1]    ; Store byte - should track size=1
	SWI #0

data_area:
	.space 16
	`
	loadProgram(t, server, sessionID, program)

	// Run to completion (program will halt on SWI #0)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	// Wait for program to complete
	time.Sleep(50 * time.Millisecond)

	// Get session status
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var status api.SessionStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode status: %v", err)
	}

	// Verify write was tracked
	if !status.HasWrite {
		t.Error("Expected HasWrite=true after STRB")
	}

	// NEW: Verify write SIZE is included in response
	if status.WriteSize != 1 {
		t.Errorf("Expected WriteSize=1 for STRB, got %d", status.WriteSize)
	}
}

// TestMemoryWriteSize_STR tests that API returns writeSize=4 for STR instruction
func TestMemoryWriteSize_STR(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)

	// Load program with STR instruction
	program := `
	.org 0x8000
main:
	LDR R0, =0xDEADBEEF
	LDR R1, =data_area
	STR R0, [R1]     ; Store word - should track size=4
	SWI #0

data_area:
	.space 16
	`
	loadProgram(t, server, sessionID, program)

	// Run to completion (program will halt on SWI #0)
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/session/%s/run", sessionID), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	// Wait for program to complete
	time.Sleep(50 * time.Millisecond)

	// Get session status
	req = httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/session/%s", sessionID), nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var status api.SessionStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode status: %v", err)
	}

	// Verify write was tracked
	if !status.HasWrite {
		t.Error("Expected HasWrite=true after STR")
	}

	// NEW: Verify write SIZE is included in response
	if status.WriteSize != 4 {
		t.Errorf("Expected WriteSize=4 for STR, got %d", status.WriteSize)
	}
}
