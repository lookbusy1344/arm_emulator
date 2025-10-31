package service_test

import (
	"bytes"
	"testing"

	"github.com/lookbusy1344/arm-emulator/service"
)

func TestEventEmittingWriter_Write(t *testing.T) {
	buffer := &bytes.Buffer{}
	// Use nil context for tests (event emission will be skipped)
	writer := service.NewEventEmittingWriter(buffer, nil)

	// Write some data
	data := []byte("Hello, World!")
	n, err := writer.Write(data)

	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	// Check buffer contains data
	if buffer.String() != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", buffer.String())
	}
}

func TestEventEmittingWriter_GetBufferAndClear(t *testing.T) {
	buffer := &bytes.Buffer{}
	// Use nil context for tests (event emission will be skipped)
	writer := service.NewEventEmittingWriter(buffer, nil)

	// Write data
	writer.Write([]byte("Test output"))

	// Get buffer contents and clear
	output := writer.GetBufferAndClear()

	if output != "Test output" {
		t.Errorf("Expected 'Test output', got '%s'", output)
	}

	// Buffer should be empty now
	if buffer.Len() != 0 {
		t.Errorf("Expected empty buffer, got %d bytes", buffer.Len())
	}
}
