package api

import (
	"bytes"
	"io"
	"sync"
)

// EventWriter is an io.Writer that broadcasts output to WebSocket clients
// It replaces the Wails-specific EventEmittingWriter with a generic broadcaster-based implementation
type EventWriter struct {
	broadcaster *Broadcaster
	sessionID   string
	stream      string // "stdout" or "stderr"
	buffer      *bytes.Buffer
	mutex       sync.Mutex
}

// NewEventWriter creates a new event-broadcasting writer
func NewEventWriter(broadcaster *Broadcaster, sessionID string, stream string) *EventWriter {
	return &EventWriter{
		broadcaster: broadcaster,
		sessionID:   sessionID,
		stream:      stream,
		buffer:      &bytes.Buffer{},
	}
}

// Write implements io.Writer interface
// It broadcasts the written data as an output event to all subscribed WebSocket clients
func (w *EventWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	n, err = w.buffer.Write(p)
	if err == nil && n > 0 && w.broadcaster != nil {
		// Broadcast the output event
		w.broadcaster.BroadcastOutput(w.sessionID, w.stream, string(p))
	}
	return n, err
}

// GetBufferAndClear returns the buffer contents and clears it
// This is useful for retrieving accumulated output
func (w *EventWriter) GetBufferAndClear() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	output := w.buffer.String()
	w.buffer.Reset()
	return output
}

// GetBuffer returns the current buffer contents without clearing
func (w *EventWriter) GetBuffer() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.buffer.String()
}

// Ensure EventWriter implements io.Writer
var _ io.Writer = (*EventWriter)(nil)
