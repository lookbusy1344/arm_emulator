package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/lookbusy1344/arm-emulator/api"
)

func createSessionWith(t *testing.T, server *api.Server, body any, origin string) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader(data))
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	return w
}

func TestFSRootFromBrowserIsForbidden(t *testing.T) {
	server := testServer()

	w := createSessionWith(t, server, api.SessionCreateRequest{FSRoot: t.TempDir()}, "http://localhost:3000")

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403: %s", w.Code, w.Body.String())
	}
	if n := sessionCount(t, server); n != 0 {
		t.Errorf("sessions = %d, want 0", n)
	}
}

func TestSessionWithoutFSRootFromBrowserIsAllowed(t *testing.T) {
	server := testServer()

	w := createSessionWith(t, server, api.SessionCreateRequest{}, "http://localhost:3000")

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201: %s", w.Code, w.Body.String())
	}
}

func TestFSRootMustBeExistingAbsoluteDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	tests := map[string]string{
		"relative": "examples",
		"missing":  filepath.Join(dir, "missing"),
		"file":     file,
		"unclean":  dir + "/../" + filepath.Base(dir),
	}
	for name, root := range tests {
		t.Run(name, func(t *testing.T) {
			server := testServer()

			w := createSessionWith(t, server, api.SessionCreateRequest{FSRoot: root}, "")

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400: %s", w.Code, w.Body.String())
			}
			if n := sessionCount(t, server); n != 0 {
				t.Errorf("sessions = %d, want 0", n)
			}
		})
	}
}

func TestFSRootFromNativeClientIsUsed(t *testing.T) {
	server := testServer()
	dir := t.TempDir()

	w := createSessionWith(t, server, api.SessionCreateRequest{FSRoot: dir}, "")

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", w.Code, w.Body.String())
	}
	var resp api.SessionCreateResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	session, err := server.GetSession(resp.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if got := session.Service.GetVM().FilesystemRoot; got != dir {
		t.Errorf("filesystem root = %q, want %q", got, dir)
	}
	if n := sessionCount(t, server); n != 1 {
		t.Errorf("sessions = %d, want 1", n)
	}
}
