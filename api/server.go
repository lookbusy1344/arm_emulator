package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Server represents the HTTP API server
type Server struct {
	sessions *SessionManager
	mux      *http.ServeMux
	server   *http.Server
	port     int
}

// NewServer creates a new API server
func NewServer(port int) *Server {
	s := &Server{
		sessions: NewSessionManager(),
		mux:      http.NewServeMux(),
		port:     port,
	}

	// Register routes
	s.registerRoutes()

	return s
}

// registerRoutes sets up all HTTP routes
func (s *Server) registerRoutes() {
	// Health check
	s.mux.HandleFunc("/health", s.handleHealth)

	// Session management
	s.mux.HandleFunc("/api/v1/session", s.handleSession)
	s.mux.HandleFunc("/api/v1/session/", s.handleSessionRoute)

	// Enable CORS for all routes
	s.mux.Handle("/", s.corsMiddleware(s.mux))
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("API server starting on http://127.0.0.1:%d", s.port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
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
	case "reset":
		s.handleReset(w, r, sessionID)
	case "registers":
		s.handleGetRegisters(w, r, sessionID)
	case "memory":
		s.handleGetMemory(w, r, sessionID)
	case "disassembly":
		s.handleGetDisassembly(w, r, sessionID)
	case "breakpoint":
		s.handleBreakpoint(w, r, sessionID)
	case "breakpoints":
		s.handleListBreakpoints(w, r, sessionID)
	case "stdin":
		s.handleSendStdin(w, r, sessionID)
	default:
		writeError(w, http.StatusNotFound, fmt.Sprintf("Unknown action: %s", action))
	}
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
