package api

import (
	"bytes"
	"encoding/json"
	"errors"
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

// testServerWithFSRootBase returns a server that accepts client fsRoots under base.
func testServerWithFSRootBase(t *testing.T, base string) *api.Server {
	t.Helper()
	server := testServer(t)
	if err := server.AllowFSRootsUnder(base); err != nil {
		t.Fatalf("AllowFSRootsUnder(%q): %v", base, err)
	}
	return server
}

// filesystemRootOf creates a session with fsRoot and returns the VM's filesystem root.
func filesystemRootOf(t *testing.T, server *api.Server, fsRoot string) string {
	t.Helper()
	w := createSessionWith(t, server, api.SessionCreateRequest{FSRoot: fsRoot}, "")
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
	return session.Service.GetVM().FilesystemRoot
}

func TestFSRootFromBrowserIsForbidden(t *testing.T) {
	base := t.TempDir()
	server := testServerWithFSRootBase(t, base)

	w := createSessionWith(t, server, api.SessionCreateRequest{FSRoot: base}, "http://localhost:3000")

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403: %s", w.Code, w.Body.String())
	}
	if n := sessionCount(t, server); n != 0 {
		t.Errorf("sessions = %d, want 0", n)
	}
}

func TestSessionWithoutFSRootFromBrowserIsAllowed(t *testing.T) {
	server := testServer(t)

	w := createSessionWith(t, server, api.SessionCreateRequest{}, "http://localhost:3000")

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201: %s", w.Code, w.Body.String())
	}
}

func TestFSRootWithoutServerBaseIsForbidden(t *testing.T) {
	server := testServer(t)

	w := createSessionWith(t, server, api.SessionCreateRequest{FSRoot: t.TempDir()}, "")

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403: %s", w.Code, w.Body.String())
	}
	if n := sessionCount(t, server); n != 0 {
		t.Errorf("sessions = %d, want 0", n)
	}
}

func TestFSRootMustBeExistingDirectoryUnderBase(t *testing.T) {
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	outside := filepath.Join(parent, "outside")
	for _, dir := range []string{filepath.Join(base, "sub"), outside} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(base, "file.txt")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	escape := filepath.Join(base, "escape")
	if err := os.Symlink(outside, escape); err != nil {
		t.Fatal(err)
	}
	tests := map[string]string{
		"relative":         "sub",
		"missing":          filepath.Join(base, "missing"),
		"file":             file,
		"unclean":          base + "/sub/../sub",
		"parent of base":   parent,
		"sibling of base":  outside,
		"base name prefix": base + "x",
		"filesystem root":  "/",
		"symlink escape":   escape,
	}
	for name, root := range tests {
		t.Run(name, func(t *testing.T) {
			server := testServerWithFSRootBase(t, base)

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

func TestFSRootEqualToBaseIsUsed(t *testing.T) {
	base := t.TempDir()
	server := testServerWithFSRootBase(t, base)

	if got := filesystemRootOf(t, server, base); got != base {
		t.Errorf("filesystem root = %q, want %q", got, base)
	}
}

func TestFSRootUnderBaseIsUsed(t *testing.T) {
	base := t.TempDir()
	sub := filepath.Join(base, "a", "b")
	if err := os.MkdirAll(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	server := testServerWithFSRootBase(t, base)

	if got := filesystemRootOf(t, server, sub); got != sub {
		t.Errorf("filesystem root = %q, want %q", got, sub)
	}
}

func TestAllowFSRootsUnderRejectsMissingDirectory(t *testing.T) {
	server := testServer(t)
	missing := filepath.Join(t.TempDir(), "missing")

	if err := server.AllowFSRootsUnder(missing); !errors.Is(err, api.ErrInvalidFSRoot) {
		t.Errorf("AllowFSRootsUnder(missing) = %v, want ErrInvalidFSRoot", err)
	}
}

func TestAllowFSRootsUnderResolvesRelativeBase(t *testing.T) {
	base := t.TempDir()
	t.Chdir(filepath.Dir(base))
	server := testServer(t)
	if err := server.AllowFSRootsUnder(filepath.Base(base)); err != nil {
		t.Fatal(err)
	}

	if got := filesystemRootOf(t, server, base); got != base {
		t.Errorf("filesystem root = %q, want %q", got, base)
	}
}
