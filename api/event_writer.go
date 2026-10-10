package api

import (
	"io"
	"sync"
	"unicode/utf8"
)

const (
	// MaxConsoleOutput is how many bytes of the most recent output a session keeps.
	MaxConsoleOutput = 1024 * 1024
	// MaxOutputEventSize is the most output one broadcast event carries. A larger
	// write sends its last MaxOutputEventSize bytes; the console endpoint holds more.
	MaxOutputEventSize = 64 * 1024
)

// EventWriter is an io.Writer that keeps recent output and broadcasts each write to
// WebSocket clients.
type EventWriter struct {
	broadcaster *Broadcaster
	sessionID   string
	stream      string // "stdout" or "stderr"
	// buffer grows to twice MaxConsoleOutput before it is trimmed, so trimming
	// costs amortised constant time per byte.
	buffer []byte
	mutex  sync.Mutex
}

// NewEventWriter creates a new event-broadcasting writer
func NewEventWriter(broadcaster *Broadcaster, sessionID string, stream string) *EventWriter {
	return &EventWriter{
		broadcaster: broadcaster,
		sessionID:   sessionID,
		stream:      stream,
	}
}

// Write implements io.Writer. It keeps the most recent MaxConsoleOutput bytes and
// broadcasts the written data as an output event.
func (w *EventWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.buffer = append(w.buffer, lastBytes(p, 2*MaxConsoleOutput)...)
	if len(w.buffer) > 2*MaxConsoleOutput {
		w.buffer = append([]byte(nil), lastBytes(w.buffer, MaxConsoleOutput)...)
	}

	if len(p) > 0 && w.broadcaster != nil {
		content := string(lastBytes(p, MaxOutputEventSize))
		debugLog("EventWriter.Write: broadcasting %d of %d bytes to session %s", len(content), len(p), w.sessionID)
		w.broadcaster.BroadcastOutput(w.sessionID, w.stream, content)
	}
	return len(p), nil
}

// lastBytes returns at most limit bytes from the end of b, starting on a UTF-8
// character boundary.
func lastBytes(b []byte, limit int) []byte {
	if len(b) <= limit {
		return b
	}
	start := len(b) - limit
	for start < len(b) && !utf8.RuneStart(b[start]) {
		start++
	}
	return b[start:]
}

// GetBufferAndClear returns the buffer contents and clears it
func (w *EventWriter) GetBufferAndClear() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	output := string(lastBytes(w.buffer, MaxConsoleOutput))
	w.buffer = nil
	return output
}

// Reset discards the captured output
func (w *EventWriter) Reset() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.buffer = nil
}

// GetBuffer returns the most recent MaxConsoleOutput bytes of output without clearing
func (w *EventWriter) GetBuffer() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return string(lastBytes(w.buffer, MaxConsoleOutput))
}

// Ensure EventWriter implements io.Writer
var _ io.Writer = (*EventWriter)(nil)
