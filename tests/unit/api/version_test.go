package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lookbusy1344/arm-emulator/api"
)

func TestVersionEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		version string
		commit  string
		date    string
	}{
		{
			name:    "production version",
			version: "v1.0.0",
			commit:  "abc123def456",
			date:    "2026-01-07 12:00:00 UTC",
		},
		{
			name:    "development version",
			version: "dev",
			commit:  "unknown",
			date:    "unknown",
		},
		{
			name:    "git describe version",
			version: "v1.1.2-123-g1e713a3-dirty",
			commit:  "1e713a3006ca790974eb44d22691a192f2ab98c1",
			date:    "2026-01-07T09:34:45Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create server with specific version info
			server := api.NewServerWithVersion(8080, tt.version, tt.commit, tt.date)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
			w := httptest.NewRecorder()

			// Call handler directly
			server.Handler().ServeHTTP(w, req)

			// Check response code
			if w.Code != http.StatusOK {
				t.Fatalf("Expected status 200, got %d", w.Code)
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			// Decode response
			var response struct {
				Version string `json:"version"`
				Commit  string `json:"commit"`
				Date    string `json:"date"`
			}

			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Verify response fields
			if response.Version != tt.version {
				t.Errorf("Expected version %q, got %q", tt.version, response.Version)
			}

			if response.Commit != tt.commit {
				t.Errorf("Expected commit %q, got %q", tt.commit, response.Commit)
			}

			if response.Date != tt.date {
				t.Errorf("Expected date %q, got %q", tt.date, response.Date)
			}
		})
	}
}

func TestVersionEndpoint_MethodNotAllowed(t *testing.T) {
	server := api.NewServerWithVersion(8080, "v1.0.0", "abc123", "2026-01-07")

	methods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/version", nil)
			w := httptest.NewRecorder()

			server.Handler().ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

func TestVersionEndpoint_CORS(t *testing.T) {
	server := api.NewServerWithVersion(8080, "v1.0.0", "abc123", "2026-01-07")

	tests := []struct {
		name           string
		origin         string
		expectAllowed  bool
		expectedOrigin string
	}{
		{
			name:           "localhost with port",
			origin:         "http://localhost:3000",
			expectAllowed:  true,
			expectedOrigin: "http://localhost:3000",
		},
		{
			name:           "127.0.0.1",
			origin:         "http://127.0.0.1:8080",
			expectAllowed:  true,
			expectedOrigin: "http://127.0.0.1:8080",
		},
		{
			name:           "file protocol",
			origin:         "file://",
			expectAllowed:  true,
			expectedOrigin: "file://",
		},
		{
			name:          "no origin",
			origin:        "",
			expectAllowed: true, // No origin is allowed (native apps)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			server.Handler().ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("Expected status 200, got %d", w.Code)
			}

			if tt.expectAllowed {
				allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
				if tt.origin != "" && allowOrigin != tt.expectedOrigin {
					t.Errorf("Expected Access-Control-Allow-Origin %q, got %q", tt.expectedOrigin, allowOrigin)
				}
			}
		})
	}
}

func TestVersionEndpoint_JSONFormat(t *testing.T) {
	server := api.NewServerWithVersion(8080, "v1.0.0", "abc123", "2026-01-07")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&jsonData); err != nil {
		t.Fatalf("Response is not valid JSON: %v", err)
	}

	// Verify all required fields are present
	requiredFields := []string{"version", "commit", "date"}
	for _, field := range requiredFields {
		if _, ok := jsonData[field]; !ok {
			t.Errorf("Missing required field: %s", field)
		}
	}

	// Verify all fields are strings
	for field, value := range jsonData {
		if _, ok := value.(string); !ok {
			t.Errorf("Field %s should be string, got %T", field, value)
		}
	}
}

func TestVersionEndpoint_NewServerBackwardsCompatibility(t *testing.T) {
	// Test that NewServer (without version) still works and returns defaults
	server := api.NewServer(8080)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
		Date    string `json:"date"`
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should have default values
	if response.Version != "dev" {
		t.Errorf("Expected default version 'dev', got %q", response.Version)
	}

	if response.Commit != "unknown" {
		t.Errorf("Expected default commit 'unknown', got %q", response.Commit)
	}

	if response.Date != "unknown" {
		t.Errorf("Expected default date 'unknown', got %q", response.Date)
	}
}
