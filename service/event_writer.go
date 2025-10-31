package service

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// EventEmittingWriter wraps a buffer and emits events when written to
type EventEmittingWriter struct {
	buffer *bytes.Buffer
	ctx    context.Context
	mutex  sync.Mutex
}

// NewEventEmittingWriter creates a new event-emitting writer
func NewEventEmittingWriter(buffer *bytes.Buffer, ctx context.Context) *EventEmittingWriter {
	return &EventEmittingWriter{
		buffer: buffer,
		ctx:    ctx,
	}
}

// Write implements io.Writer interface
func (w *EventEmittingWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	n, err = w.buffer.Write(p)
	if err == nil && n > 0 && w.ctx != nil {
		// Emit event with the new output
		runtime.EventsEmit(w.ctx, "vm:output", string(p))
	}
	return n, err
}

// GetBufferAndClear returns buffer contents and clears it
func (w *EventEmittingWriter) GetBufferAndClear() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	output := w.buffer.String()
	w.buffer.Reset()
	return output
}

// Ensure EventEmittingWriter implements io.Writer
var _ io.Writer = (*EventEmittingWriter)(nil)
