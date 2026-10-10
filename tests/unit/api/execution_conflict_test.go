package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lookbusy1344/arm-emulator/service"
)

const spinLoop = ".org 0x8000\nloop: B loop"

func post(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func waitUntil(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestStepWhileRunningConflicts(t *testing.T) {
	server := testServer(t)
	sessionID := createTestSession(t, server)
	loadProgram(t, server, sessionID, spinLoop)
	session, err := server.GetSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	svc := session.Service

	if w := post(t, server.Handler(), fmt.Sprintf("/api/v1/session/%s/run", sessionID)); w.Code != http.StatusOK {
		t.Fatalf("run: status %d", w.Code)
	}
	waitUntil(t, "execution to start", func() bool { return svc.GetRegisterState().Cycles > 0 })

	for _, action := range []string{"step", "step-over"} {
		w := post(t, server.Handler(), fmt.Sprintf("/api/v1/session/%s/%s", sessionID, action))
		if w.Code != http.StatusConflict {
			t.Errorf("%s while running: expected 409, got %d: %s", action, w.Code, w.Body.String())
		}
	}

	if w := post(t, server.Handler(), fmt.Sprintf("/api/v1/session/%s/stop", sessionID)); w.Code != http.StatusOK {
		t.Fatalf("stop: status %d", w.Code)
	}
}

func TestDestroySessionStopsExecution(t *testing.T) {
	server := testServer(t)
	sessionID := createTestSession(t, server)
	loadProgram(t, server, sessionID, spinLoop)
	session, err := server.GetSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	svc := session.Service

	if w := post(t, server.Handler(), fmt.Sprintf("/api/v1/session/%s/run", sessionID)); w.Code != http.StatusOK {
		t.Fatalf("run: status %d", w.Code)
	}
	waitUntil(t, "execution to start", func() bool { return svc.GetRegisterState().Cycles > 0 })

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/session/%s", sessionID), nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("destroy: status %d", w.Code)
	}

	if svc.IsRunning() {
		t.Error("execution still running after session destroy")
	}
	before := svc.GetRegisterState().Cycles
	time.Sleep(10 * time.Millisecond)
	if after := svc.GetRegisterState().Cycles; after != before {
		t.Errorf("guest kept executing after destroy: cycles %d -> %d", before, after)
	}
	if err := svc.Step(); !errors.Is(err, service.ErrClosed) {
		t.Errorf("expected ErrClosed after destroy, got %v", err)
	}
}

func TestRunWhileRunningConflicts(t *testing.T) {
	server := testServer(t)
	sessionID := createTestSession(t, server)
	loadProgram(t, server, sessionID, spinLoop)
	session, err := server.GetSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	svc := session.Service
	runPath := fmt.Sprintf("/api/v1/session/%s/run", sessionID)

	if w := post(t, server.Handler(), runPath); w.Code != http.StatusOK {
		t.Fatalf("first run: status %d", w.Code)
	}
	w := post(t, server.Handler(), runPath)

	if w.Code != http.StatusConflict {
		t.Errorf("second run: expected 409, got %d: %s", w.Code, w.Body.String())
	}
	if !svc.IsRunning() {
		t.Error("first run stopped after the rejected second run")
	}
	if w := post(t, server.Handler(), fmt.Sprintf("/api/v1/session/%s/stop", sessionID)); w.Code != http.StatusOK {
		t.Fatalf("stop: status %d", w.Code)
	}
}

func TestRunAfterDestroyIsGone(t *testing.T) {
	server := testServer(t)
	sessionID := createTestSession(t, server)
	loadProgram(t, server, sessionID, spinLoop)
	session, err := server.GetSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	session.Service.Close()

	w := post(t, server.Handler(), fmt.Sprintf("/api/v1/session/%s/run", sessionID))

	if w.Code != http.StatusGone {
		t.Errorf("run after close: expected 410, got %d: %s", w.Code, w.Body.String())
	}
}
