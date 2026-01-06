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
	sessions    *SessionManager
	broadcaster *Broadcaster
	mux         *http.ServeMux
	server      *http.Server
	port        int
}

// NewServer creates a new API server
func NewServer(port int) *Server {
	broadcaster := NewBroadcaster()

	s := &Server{
		sessions:    NewSessionManager(broadcaster),
		broadcaster: broadcaster,
		mux:         http.NewServeMux(),
		port:        port,
	}

	// Register routes
	s.registerRoutes()

	return s
}

// Handler returns the HTTP handler with CORS middleware applied
func (s *Server) Handler() http.Handler {
	return s.corsMiddleware(s.mux)
}

// registerRoutes sets up all HTTP routes
func (s *Server) registerRoutes() {
	// Health check
	s.mux.HandleFunc("/health", s.handleHealth)

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

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler:      s.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("API server starting on http://127.0.0.1:%d", s.port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Close broadcaster to disconnect all WebSocket clients
	if s.broadcaster != nil {
		s.broadcaster.Close()
	}

	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// GetBroadcaster returns the broadcaster (for testing)
func (s *Server) GetBroadcaster() *Broadcaster {
	return s.broadcaster
}

// corsMiddleware adds CORS headers restricted to localhost origins for security
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow localhost origins only (various forms and ports)
		// Allowed: http://localhost:*, http://127.0.0.1:*, https://localhost:*, file://
		// Rejected: any remote origin
		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isAllowedOrigin checks if the origin is from localhost
func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return true // No origin header (native apps, curl, etc.)
	}

	// Allow file:// for local HTML files
	if strings.HasPrefix(origin, "file://") {
		return true
	}

	// Allow localhost and 127.0.0.1 with http/https on any port
	if strings.HasPrefix(origin, "http://localhost") ||
		strings.HasPrefix(origin, "https://localhost") ||
		strings.HasPrefix(origin, "http://127.0.0.1") ||
		strings.HasPrefix(origin, "https://127.0.0.1") {
		return true
	}

	return false
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
