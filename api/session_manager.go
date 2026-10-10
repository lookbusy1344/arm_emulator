package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lookbusy1344/arm-emulator/service"
	"github.com/lookbusy1344/arm-emulator/vm"
)

var (
	// ErrSessionNotFound is returned when a session is not found
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionAlreadyExists is returned when trying to create a session with an existing ID
	ErrSessionAlreadyExists = errors.New("session already exists")
	// ErrInvalidFSRoot is returned when a requested filesystem root is not a clean
	// absolute path to an existing directory
	ErrInvalidFSRoot = errors.New("fsRoot must be a clean absolute path to an existing directory")
)

// validateFSRoot checks a client-supplied filesystem root.
func validateFSRoot(root string) error {
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return ErrInvalidFSRoot
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return ErrInvalidFSRoot
	}
	return nil
}

// Session represents an active emulator session
type Session struct {
	ID        string
	Service   *service.DebuggerService
	CreatedAt time.Time
	TempDir   string // Temporary directory for filesystem operations (cleaned up on destroy)
}

// SessionManager manages multiple emulator sessions
type SessionManager struct {
	sessions    map[string]*Session
	broadcaster *Broadcaster
	mu          sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager(broadcaster *Broadcaster) *SessionManager {
	return &SessionManager{
		sessions:    make(map[string]*Session),
		broadcaster: broadcaster,
	}
}

// CreateSession creates a new session with a unique ID
func (sm *SessionManager) CreateSession(opts SessionCreateRequest) (*Session, error) {
	// Generate unique session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	// Create VM instance (note: opts.MemorySize is currently unused, VM uses default size)
	// TODO: Future enhancement - configure VM memory size based on opts.MemorySize
	machine := vm.NewVM()

	// Configure filesystem root for security and file operations
	// If FSRoot is provided, use it; otherwise create a temporary directory for this session
	var tempDir string
	if opts.FSRoot != "" {
		if err := validateFSRoot(opts.FSRoot); err != nil {
			return nil, err
		}
		machine.FilesystemRoot = opts.FSRoot
	} else {
		// Create a temporary directory for this session's file operations
		var err error
		tempDir, err = os.MkdirTemp("", "arm-emulator-session-*")
		if err != nil {
			return nil, err
		}
		machine.FilesystemRoot = tempDir
	}

	// Set up output broadcasting if broadcaster is available
	if sm.broadcaster != nil {
		outputWriter := NewEventWriter(sm.broadcaster, sessionID, "stdout")
		machine.OutputWriter = outputWriter
		debugLog("Session %s: EventWriter set up for stdout broadcasting", sessionID)

		// Set up VM state change callback to broadcast state changes (e.g., waiting_for_input)
		broadcaster := sm.broadcaster
		sid := sessionID
		machine.OnStateChange = func(state vm.ExecutionState) {
			serviceState := service.VMStateToExecution(state)
			data := map[string]interface{}{
				"status": string(serviceState),
			}
			broadcaster.BroadcastState(sid, data)
		}
	} else {
		debugLog("Session %s: WARNING - no broadcaster available for output", sessionID)
	}

	// Create debugger service
	debugService := service.NewDebuggerService(machine)

	session := &Session{
		ID:        sessionID,
		Service:   debugService,
		CreatedAt: time.Now(),
		TempDir:   tempDir,
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[sessionID]; exists {
		return nil, ErrSessionAlreadyExists
	}

	sm.sessions[sessionID] = session
	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// DestroySession removes a session by ID
func (sm *SessionManager) DestroySession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	// Handlers may still hold the session, so stop execution rather than clearing Service.
	session.Service.Close()

	// Clean up temporary directory if it was created
	if session.TempDir != "" {
		_ = os.RemoveAll(session.TempDir)
	}

	delete(sm.sessions, sessionID)
	return nil
}

// ListSessions returns a list of all session IDs
func (sm *SessionManager) ListSessions() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]string, 0, len(sm.sessions))
	for id := range sm.sessions {
		ids = append(ids, id)
	}
	return ids
}

// Count returns the number of active sessions
func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.sessions)
}

// generateSessionID generates a unique session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
