package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/lookbusy1344/arm-emulator/api"
)

func sessionCount(t *testing.T, server *api.Server) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/session", nil)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list sessions: status %d", w.Code)
	}
	var resp struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	return resp.Count
}

func TestOriginPolicy(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed bool
	}{
		{"no origin (native client)", "", true},
		{"localhost with port", "http://localhost:3000", true},
		{"localhost without port", "http://localhost", true},
		{"https localhost", "https://localhost:8443", true},
		{"ipv4 loopback", "http://127.0.0.1:5173", true},
		{"ipv6 loopback", "http://[::1]:8080", true},
		{"local file", "file://", true},
		{"file with remote host", "file://attacker.example", false},
		{"localhost prefix of attacker domain", "http://localhost.attacker.example", false},
		{"loopback prefix of attacker domain", "http://127.0.0.1.attacker.example", false},
		{"remote origin", "https://attacker.example", false},
		{"userinfo trick", "http://localhost@attacker.example", false},
		{"opaque null origin", "null", false},
		{"non-http scheme", "ftp://localhost", false},
		{"malformed", "http://%zz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := withCleanup(t, api.NewServer(8080))

			req := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader("{}"))
			req.Header.Set("Content-Type", "text/plain")
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()
			server.Handler().ServeHTTP(w, req)

			acao := w.Header().Get("Access-Control-Allow-Origin")
			if tt.allowed {
				if w.Code != http.StatusCreated {
					t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
				}
				if tt.origin != "" && acao != tt.origin {
					t.Errorf("expected ACAO %q, got %q", tt.origin, acao)
				}
				return
			}

			if w.Code != http.StatusForbidden {
				t.Fatalf("expected 403, got %d", w.Code)
			}
			if acao != "" {
				t.Errorf("expected no ACAO header, got %q", acao)
			}
			if n := sessionCount(t, server); n != 0 {
				t.Errorf("rejected request created %d session(s)", n)
			}
		})
	}
}

func TestWebSocketRejectsForeignOrigin(t *testing.T) {
	server := withCleanup(t, api.NewServer(8080))
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/ws"

	header := http.Header{"Origin": []string{"http://localhost.attacker.example"}}
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err == nil {
		_ = conn.Close()
		t.Fatal("expected WebSocket dial from foreign origin to fail")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 response, got %v", resp)
	}

	header = http.Header{"Origin": []string{"http://localhost:3000"}}
	conn, _, err = websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("expected WebSocket dial from localhost origin to succeed: %v", err)
	}
	_ = conn.Close()
}

func TestLoopbackHostOnly(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := api.LoopbackHostOnly(ok)

	tests := []struct {
		host    string
		allowed bool
	}{
		{"localhost:8080", true},
		{"localhost", true},
		{"LOCALHOST:8080", true},
		{"127.0.0.1:8080", true},
		{"[::1]:8080", true},
		{"attacker.example:8080", false},
		{"attacker.example", false},
		{"localhost.attacker.example:8080", false},
		{"127.0.0.2:8080", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			req.Host = tt.host
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			want := http.StatusForbidden
			if tt.allowed {
				want = http.StatusNoContent
			}
			if w.Code != want {
				t.Errorf("host %q: expected %d, got %d", tt.host, want, w.Code)
			}
		})
	}
}
