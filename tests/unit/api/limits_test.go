package api

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

	"github.com/lookbusy1344/arm-emulator/api"
	"github.com/lookbusy1344/arm-emulator/service"
)

func createSessionStatus(t *testing.T, server *api.Server) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	return w.Code
}

func TestSessionLimit(t *testing.T) {
	server := testServer(t)
	ids := make([]string, api.MaxSessions)
	for i := range ids {
		ids[i] = createTestSession(t, server)
	}

	if code := createSessionStatus(t, server); code != http.StatusTooManyRequests {
		t.Fatalf("session %d: status %d, want %d", api.MaxSessions+1, code, http.StatusTooManyRequests)
	}
	if n := sessionCount(t, server); n != api.MaxSessions {
		t.Fatalf("session count after refusal = %d, want %d", n, api.MaxSessions)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/session/"+ids[0], nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("destroy: status %d", w.Code)
	}
	if code := createSessionStatus(t, server); code != http.StatusCreated {
		t.Fatalf("create after destroy: status %d, want %d", code, http.StatusCreated)
	}
}

func sendStdin(t *testing.T, server *api.Server, sessionID, data string) int {
	t.Helper()
	body, err := json.Marshal(api.StdinRequest{Data: data})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/session/%s/stdin", sessionID), bytes.NewReader(body))
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	return w.Code
}

func TestStdinQueueLimit(t *testing.T) {
	server := testServer(t)
	sessionID := createTestSession(t, server)
	half := strings.Repeat("a", service.MaxQueuedInput/2)

	for i := range 2 {
		if code := sendStdin(t, server, sessionID, half); code != http.StatusOK {
			t.Fatalf("stdin %d: status %d, want %d", i, code, http.StatusOK)
		}
	}
	if code := sendStdin(t, server, sessionID, "b"); code != http.StatusTooManyRequests {
		t.Fatalf("stdin over limit: status %d, want %d", code, http.StatusTooManyRequests)
	}
}

// Concurrent creates must not overshoot the limit or leave directories behind.
func TestSessionLimitUnderConcurrentCreates(t *testing.T) {
	const extra = 16
	server := testServer(t)
	before := sessionDirs(t)

	var wg sync.WaitGroup
	for range api.MaxSessions + extra {
		wg.Go(func() { createSessionStatus(t, server) })
	}
	wg.Wait()

	if n := sessionCount(t, server); n != api.MaxSessions {
		t.Fatalf("session count = %d, want %d", n, api.MaxSessions)
	}
	if created := sessionDirs(t) - before; created != api.MaxSessions {
		t.Fatalf("session directories created = %d, want %d", created, api.MaxSessions)
	}
}

// sessionDirs counts the session directories in the temporary directory.
func sessionDirs(t *testing.T) int {
	t.Helper()
	dirs, err := filepath.Glob(filepath.Join(os.TempDir(), "arm-emulator-session-*"))
	if err != nil {
		t.Fatal(err)
	}
	return len(dirs)
}
