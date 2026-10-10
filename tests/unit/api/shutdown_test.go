package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/lookbusy1344/arm-emulator/api"
	"github.com/lookbusy1344/arm-emulator/service"
)

func TestShutdownDestroysSessions(t *testing.T) {
	server := testServer()
	sessionID := createTestSession(t, server)
	loadProgram(t, server, sessionID, spinLoop)
	session, err := server.GetSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if session.TempDir == "" {
		t.Fatal("setup: session has no temporary directory")
	}
	if w := post(t, server.Handler(), fmt.Sprintf("/api/v1/session/%s/run", sessionID)); w.Code != http.StatusOK {
		t.Fatalf("run: status %d", w.Code)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if _, err := os.Stat(session.TempDir); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("temporary directory %s remains after shutdown (stat: %v)", session.TempDir, err)
	}
	if _, err := server.GetSession(sessionID); !errors.Is(err, api.ErrSessionNotFound) {
		t.Errorf("GetSession after shutdown = %v, want ErrSessionNotFound", err)
	}
	if err := session.Service.Step(); !errors.Is(err, service.ErrClosed) {
		t.Errorf("Step after shutdown = %v, want ErrClosed", err)
	}
}

func TestShutdownKeepsClientFSRoot(t *testing.T) {
	server := testServer()
	dir := t.TempDir()
	w := createSessionWith(t, server, api.SessionCreateRequest{FSRoot: dir}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status %d", w.Code)
	}
	var resp api.SessionCreateResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Errorf("client fsRoot %s removed by shutdown (stat: %v)", dir, err)
	}
}
