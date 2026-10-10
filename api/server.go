package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Server represents the HTTP API server
type Server struct {
	sessions    *SessionManager
	broadcaster *Broadcaster
	mux         *http.ServeMux
	serverMu    sync.Mutex
	server      *http.Server
	tokenPath   string // token file Start wrote; Shutdown removes it
	port        int
	version     string
	commit      string
	date        string
}

// NewServer creates a new API server
func NewServer(port int) *Server {
	return NewServerWithVersion(port, "dev", "unknown", "unknown")
}

// NewServerWithVersion creates a new API server with version information
func NewServerWithVersion(port int, version, commit, date string) *Server {
	broadcaster := NewBroadcaster()

	s := &Server{
		sessions:    NewSessionManager(broadcaster),
		broadcaster: broadcaster,
		mux:         http.NewServeMux(),
		port:        port,
		version:     version,
		commit:      commit,
		date:        date,
	}

	// Register routes
	s.registerRoutes()

	return s
}

// GetSession returns a session by ID (for testing)
func (s *Server) GetSession(sessionID string) (*Session, error) {
	return s.sessions.GetSession(sessionID)
}

// Handler returns the HTTP handler with CORS and security middleware applied
func (s *Server) Handler() http.Handler {
	return s.corsMiddleware(s.rejectPathTraversal(s.mux))
}

// rejectPathTraversal rejects requests with ".." in their path before the ServeMux
// can issue a redirect. Without this, Go's ServeMux path-cleaning redirects path
// traversal attempts with 307 rather than rejecting them outright.
func (s *Server) rejectPathTraversal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.RawPath
		if path == "" {
			path = r.URL.Path
		}
		if strings.Contains(path, "..") {
			writeError(w, http.StatusBadRequest, "Invalid path")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// registerRoutes sets up all HTTP routes
func (s *Server) registerRoutes() {
	// Health check
	s.mux.HandleFunc("/health", s.handleHealth)

	// Version information
	s.mux.HandleFunc("/api/v1/version", s.handleGetVersion)

	// WebSocket endpoint for real-time updates
	s.mux.HandleFunc("/api/v1/ws", s.handleWebSocket)

	// Session management
	s.mux.HandleFunc("/api/v1/session", s.handleSession)
	s.mux.HandleFunc("/api/v1/session/", s.handleSessionRoute)

	// Configuration
	s.mux.HandleFunc("/api/v1/config", s.handleConfig)

	// Examples
	s.mux.HandleFunc("/api/v1/examples", s.handleExamples)
	s.mux.HandleFunc("/api/v1/examples/", s.handleExamplesRoute)
}

// Start binds the port, writes a fresh API token to the token file for the port, and
// serves requests that carry the token. The port is bound first, so a backend that
// cannot bind leaves the running backend's token file alone.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.port))
	if err != nil {
		return err
	}
	token, err := GenerateToken()
	if err != nil {
		_ = ln.Close()
		return err
	}
	tokenPath, err := TokenPath(s.port)
	if err != nil {
		_ = ln.Close()
		return err
	}

	srv := &http.Server{
		Handler:      LoopbackHostOnly(RequireToken(token, s.Handler())),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Clients may call as soon as the token file exists, so Shutdown must already
	// see the server; holding the lock across the write orders the two.
	s.serverMu.Lock()
	err = WriteTokenFile(tokenPath, token)
	if err == nil {
		s.server = srv
		s.tokenPath = tokenPath
	}
	s.serverMu.Unlock()
	if err != nil {
		_ = ln.Close()
		return err
	}

	log.Printf("API server starting on http://127.0.0.1:%d (version: %s, commit: %s, built: %s), token file %s", s.port, s.version, s.commit, s.date, tokenPath)
	return srv.Serve(ln)
}

// Shutdown gracefully shuts down the server, then destroys every session
func (s *Server) Shutdown(ctx context.Context) error {
	// Close broadcaster to disconnect all WebSocket clients
	if s.broadcaster != nil {
		s.broadcaster.Close()
	}

	s.serverMu.Lock()
	srv, tokenPath := s.server, s.tokenPath
	s.serverMu.Unlock()
	var err error
	if srv != nil {
		err = srv.Shutdown(ctx)
	}
	if tokenPath != "" {
		_ = os.Remove(tokenPath)
	}
	s.sessions.DestroyAll()
	return err
}

// GetBroadcaster returns the broadcaster (for testing)
func (s *Server) GetBroadcaster() *Broadcaster {
	return s.broadcaster
}

// corsMiddleware rejects requests from browser pages not served by this machine and
// adds CORS headers for the rest.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := isAllowedOrigin(origin)
		if allowed && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// A preflight without Access-Control-Allow-Origin fails in the browser.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Simple requests (e.g. a text/plain POST) skip preflight, so reject them here.
		if !allowed {
			writeError(w, http.StatusForbidden, "Origin not allowed")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":   "ok",
		"sessions": s.sessions.Count(),
		"time":     time.Now().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleSession handles session creation and listing
func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateSession(w, r)
	case http.MethodGet:
		s.handleListSessions(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSessionRoute handles session-specific routes
func (s *Server) handleSessionRoute(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from path: /api/v1/session/{id}/action
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/session/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "Session ID required")
		return
	}

	sessionID := parts[0]

	// Route to appropriate handler based on action
	if len(parts) == 1 {
		// /api/v1/session/{id}
		switch r.Method {
		case http.MethodGet:
			s.handleGetSessionStatus(w, r, sessionID)
		case http.MethodDelete:
			s.handleDestroySession(w, r, sessionID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	action := parts[1]
	switch action {
	case "load":
		s.handleLoadProgram(w, r, sessionID)
	case "run":
		s.handleRun(w, r, sessionID)
	case "stop":
		s.handleStop(w, r, sessionID)
	case "step":
		s.handleStep(w, r, sessionID)
	case "step-over":
		s.handleStepOver(w, r, sessionID)
	case "step-out":
		s.handleStepOut(w, r, sessionID)
	case "reset":
		s.handleReset(w, r, sessionID)
	case "restart":
		s.handleRestart(w, r, sessionID)
	case "registers":
		s.handleGetRegisters(w, r, sessionID)
	case "memory":
		s.handleGetMemory(w, r, sessionID)
	case "disassembly":
		s.handleGetDisassembly(w, r, sessionID)
	case "console":
		s.handleGetConsoleOutput(w, r, sessionID)
	case "breakpoint":
		s.handleBreakpoint(w, r, sessionID)
	case "breakpoints":
		s.handleListBreakpoints(w, r, sessionID)
	case "sourcemap":
		s.handleGetSourceMap(w, r, sessionID)
	case "watchpoint":
		// Handle DELETE /api/v1/session/{id}/watchpoint/{watchpointID}
		if len(parts) == 3 && r.Method == http.MethodDelete {
			// Parse watchpoint ID
			watchpointID := 0
			if _, err := fmt.Sscanf(parts[2], "%d", &watchpointID); err != nil {
				writeError(w, http.StatusBadRequest, "Invalid watchpoint ID")
				return
			}
			s.handleDeleteWatchpoint(w, r, sessionID, watchpointID)
		} else {
			// POST /api/v1/session/{id}/watchpoint
			s.handleWatchpoint(w, r, sessionID)
		}
	case "watchpoints":
		s.handleListWatchpoints(w, r, sessionID)
	case "evaluate":
		s.handleEvaluateExpression(w, r, sessionID)
	case "trace":
		// Handle /api/v1/session/{id}/trace/{enable|disable|data}
		if len(parts) < 3 {
			writeError(w, http.StatusBadRequest, "Trace action required (enable, disable, or data)")
			return
		}
		traceAction := parts[2]
		if traceAction == "data" {
			s.handleTraceData(w, r, sessionID)
		} else {
			s.handleTraceControl(w, r, sessionID, traceAction)
		}
	case "stats":
		// Handle /api/v1/session/{id}/stats or /api/v1/session/{id}/stats/{enable|disable}
		if len(parts) == 2 {
			// GET /api/v1/session/{id}/stats
			s.handleStats(w, r, sessionID)
		} else if len(parts) == 3 {
			// POST /api/v1/session/{id}/stats/{enable|disable}
			statsAction := parts[2]
			s.handleStatsControl(w, r, sessionID, statsAction)
		} else {
			writeError(w, http.StatusBadRequest, "Invalid stats endpoint")
		}
	case "stdin":
		s.handleSendStdin(w, r, sessionID)
	default:
		writeError(w, http.StatusNotFound, fmt.Sprintf("Unknown action: %s", action))
	}
}

// handleConfig handles GET/PUT /api/v1/config
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetConfig(w, r)
	case http.MethodPut:
		s.handleUpdateConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleExamples handles GET /api/v1/examples
func (s *Server) handleExamples(w http.ResponseWriter, r *http.Request) {
	s.handleListExamples(w, r)
}

// handleExamplesRoute handles GET /api/v1/examples/{name}
func (s *Server) handleExamplesRoute(w http.ResponseWriter, r *http.Request) {
	// Extract example name from path: /api/v1/examples/{name}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/examples/")

	if path == "" {
		writeError(w, http.StatusBadRequest, "Example name required")
		return
	}

	s.handleGetExample(w, r, path)
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
		Code:    status,
	})
}

func readJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1024*1024)) // 1MB limit
	return decoder.Decode(v)
}
