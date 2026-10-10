package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lookbusy1344/arm-emulator/api"
)

const printA = ".org 0x8000\n_start:\n\tMOV R0, #65\n\tSWI #0x01\n\tMOV R0, #0\n\tSWI #0x00\n"

func consoleOutput(t *testing.T, server *api.Server, sessionID string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/session/%s/console", sessionID), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("console: status %d", w.Code)
	}
	var resp api.ConsoleOutputResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	return resp.Output
}

func runToHalted(t *testing.T, server *api.Server, sessionID string) {
	t.Helper()
	if w := post(t, server.Handler(), fmt.Sprintf("/api/v1/session/%s/run", sessionID)); w.Code != http.StatusOK {
		t.Fatalf("run: status %d", w.Code)
	}
	waitForState(t, server, sessionID, "halted")
}

func TestConsoleClearedOnLoad(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)
	loadProgram(t, server, sessionID, printA)
	runToHalted(t, server, sessionID)

	loadProgram(t, server, sessionID, printA)

	if out := consoleOutput(t, server, sessionID); out != "" {
		t.Errorf("console after reload = %q, want empty", out)
	}
}

func TestConsoleHoldsOnlyLatestRun(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)
	loadProgram(t, server, sessionID, printA)
	runToHalted(t, server, sessionID)

	runToHalted(t, server, sessionID)

	if out := consoleOutput(t, server, sessionID); out != "A" {
		t.Errorf("console after second run = %q, want %q", out, "A")
	}
}

func TestConsoleClearedOnRestartAndReset(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)
	loadProgram(t, server, sessionID, printA)

	// Reset also unloads the program, so it comes last.
	for _, action := range []string{"restart", "reset"} {
		runToHalted(t, server, sessionID)
		if w := post(t, server.Handler(), fmt.Sprintf("/api/v1/session/%s/%s", sessionID, action)); w.Code != http.StatusOK {
			t.Fatalf("%s: status %d", action, w.Code)
		}
		if out := consoleOutput(t, server, sessionID); out != "" {
			t.Errorf("console after %s = %q, want empty", action, out)
		}
	}
}
