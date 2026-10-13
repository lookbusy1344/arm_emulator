package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/api"
)

// Clients cannot choose the session's filesystem root; an fsRoot field is ignored.
func TestSessionFilesystemRootIsItsTemporaryDirectory(t *testing.T) {
	server := testServer(t)
	chosen := t.TempDir()
	body, err := json.Marshal(map[string]string{"fsRoot": chosen})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(string(body)))
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

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
	if session.TempDir == "" {
		t.Fatal("session has no temporary directory")
	}
	if got := session.Service.GetVM().FilesystemRoot(); got != session.TempDir {
		t.Errorf("filesystem root = %q, want the session's temporary directory %q (client sent %q)", got, session.TempDir, chosen)
	}
}
