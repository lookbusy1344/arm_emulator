package debugger

import (
	"sync"
)

// CommandHistory maintains a history of executed commands
type CommandHistory struct {
	mu       sync.RWMutex
	commands []string
	maxSize  int
	position int // Current position in history for navigation
}

// NewCommandHistory creates a new command history
func NewCommandHistory() *CommandHistory {
	return &CommandHistory{
		commands: make([]string, 0, 100),
		maxSize:  1000, // Keep last 1000 commands
		position: 0,
	}
}

// Add adds a command to history
func (h *CommandHistory) Add(cmd string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Don't add empty commands or duplicates of the last command
	if cmd == "" {
		return
	}

	if len(h.commands) > 0 && h.commands[len(h.commands)-1] == cmd {
		h.position = len(h.commands)
		return
	}

	// Add command
	h.commands = append(h.commands, cmd)

	// Trim if exceeds max size
	if len(h.commands) > h.maxSize {
		h.commands = h.commands[len(h.commands)-h.maxSize:]
	}

	// Reset position to end
	h.position = len(h.commands)
}

// Previous returns the previous command in history
func (h *CommandHistory) Previous() string {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.commands) == 0 || h.position == 0 {
		return ""
	}

	h.position--
	return h.commands[h.position]
}

// Next returns the next command in history
func (h *CommandHistory) Next() string {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.commands) == 0 {
		return ""
	}

	if h.position >= len(h.commands)-1 {
		h.position = len(h.commands)
		return ""
	}

	h.position++
	return h.commands[h.position]
}

// GetLast returns the last command (without changing position)
func (h *CommandHistory) GetLast() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.commands) == 0 {
		return ""
	}

	return h.commands[len(h.commands)-1]
}

// GetAll returns all commands in history
func (h *CommandHistory) GetAll() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy
	result := make([]string, len(h.commands))
	copy(result, h.commands)
	return result
}

// Clear clears the command history
func (h *CommandHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.commands = h.commands[:0]
	h.position = 0
}

// Size returns the number of commands in history
func (h *CommandHistory) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.commands)
}

// Search searches for commands matching a prefix
func (h *CommandHistory) Search(prefix string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var results []string
	for _, cmd := range h.commands {
		if len(cmd) >= len(prefix) && cmd[:len(prefix)] == prefix {
			results = append(results, cmd)
		}
	}

	return results
}
