package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/lookbusy1344/arm-emulator/api"
)

const (
	testToken      = "abc123"
	ownerOnlyFile  = 0o600
	startupTimeout = 5 * time.Second
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
}

func TestRequireToken(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		authorization string
		want          int
	}{
		{"correct token", "/api/v1/session", "Bearer " + testToken, http.StatusOK},
		{"websocket with token", "/api/v1/ws", "Bearer " + testToken, http.StatusOK},
		{"health without token", "/health", "", http.StatusOK},
		{"missing header", "/api/v1/session", "", http.StatusUnauthorized},
		{"websocket without token", "/api/v1/ws", "", http.StatusUnauthorized},
		{"wrong token", "/api/v1/session", "Bearer wrong", http.StatusUnauthorized},
		{"token prefix only", "/api/v1/session", "Bearer abc", http.StatusUnauthorized},
		{"token with suffix", "/api/v1/session", "Bearer " + testToken + "x", http.StatusUnauthorized},
		{"wrong scheme", "/api/v1/session", "Basic " + testToken, http.StatusUnauthorized},
		{"bare token", "/api/v1/session", testToken, http.StatusUnauthorized},
		{"health subpath is not exempt", "/health/x", "", http.StatusUnauthorized},
	}
	handler := api.RequireToken(testToken, okHandler())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Fatalf("status = %d, want %d", w.Code, tt.want)
			}
			if tt.want == http.StatusUnauthorized && w.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Errorf("WWW-Authenticate = %q, want %q", w.Header().Get("WWW-Authenticate"), "Bearer")
			}
		})
	}
}

func TestGenerateTokenIsRandomHex(t *testing.T) {
	const hexLen = 64 // 32 random bytes
	a, err := api.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := api.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != hexLen || strings.Trim(a, "0123456789abcdef") != "" {
		t.Errorf("token %q is not %d hex digits", a, hexLen)
	}
	if a == b {
		t.Error("two tokens are equal")
	}
}

// isolateHome points the per-user directories at a fresh directory and returns it.
func isolateHome(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("token location test covers the Unix layout")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

func TestTokenPath(t *testing.T) {
	home := isolateHome(t)

	path, err := api.TokenPath(8080)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".config", "arm-emu", "api-token-8080"); path != want {
		t.Errorf("TokenPath = %q, want %q", path, want)
	}
}

func TestWriteTokenFileReplacesContentOwnerOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arm-emu", "api-token-8080")

	for _, token := range []string{"first", testToken} {
		if err := api.WriteTokenFile(path, token); err != nil {
			t.Fatal(err)
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != testToken {
		t.Errorf("token file = %q, want %q", content, testToken)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != ownerOnlyFile {
		t.Errorf("token file mode = %o, want %o", mode, ownerOnlyFile)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("token directory holds %d entries, want 1 (no temporary files left)", len(entries))
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	return port
}

// startServer runs a real server and returns it once its token file exists.
func startServer(t *testing.T, port int) (*api.Server, string, <-chan error) {
	t.Helper()
	server := api.NewServer(port)
	done := make(chan error, 1)
	go func() { done <- server.Start() }()
	path, err := api.TokenPath(port)
	if err != nil {
		t.Fatal(err)
	}
	waitUntil(t, "token file", func() bool {
		_, err := os.Stat(path)
		return err == nil
	})
	return server, path, done
}

// noReuseClient opens a connection per request. A pooled or spare connection would keep
// Server.Shutdown waiting until the connection is old enough to count as idle.
var noReuseClient = &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}

func getSessions(t *testing.T, port int, token string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/api/v1/session", port), nil)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := noReuseClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	return resp.StatusCode
}

func TestStartRequiresTokenFromFile(t *testing.T) {
	isolateHome(t)
	port := freePort(t)
	server, path, done := startServer(t, port)

	token, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if code := getSessions(t, port, ""); code != http.StatusUnauthorized {
		t.Errorf("without token: status %d, want %d", code, http.StatusUnauthorized)
	}
	if code := getSessions(t, port, string(token)); code != http.StatusOK {
		t.Errorf("with token: status %d, want %d", code, http.StatusOK)
	}

	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, http.ErrServerClosed) {
		t.Errorf("Start returned %v, want %v", err, http.ErrServerClosed)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("token file after shutdown: %v, want removed", err)
	}
}

// A second backend that cannot bind the port must leave the running backend's token alone.
func TestStartOnBusyPortKeepsExistingToken(t *testing.T) {
	isolateHome(t)
	port := freePort(t)
	server, path, _ := startServer(t, port)
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	second := api.NewServer(port)
	if err := second.Start(); err == nil {
		t.Fatal("second Start on a busy port succeeded")
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("token file after failed start: %v", err)
	}
	if string(after) != string(before) {
		t.Errorf("token file changed from %q to %q", before, after)
	}
	if code := getSessions(t, port, string(before)); code != http.StatusOK {
		t.Errorf("first backend with its token: status %d, want %d", code, http.StatusOK)
	}
}
