package encoder

import (
	"fmt"

	"github.com/lookbusy1344/arm-emulator/parser"
)

// EncodingError provides detailed context for encoding failures.
// It includes the original instruction's source location (file, line, column),
// the raw source line, and the underlying error message.
type EncodingError struct {
	Instruction *parser.Instruction // Original instruction that failed to encode
	Message     string              // Error description
	Wrapped     error               // Underlying error (may be nil)
}

// Error implements the error interface.
// Returns a formatted error message with source location context.
func (e *EncodingError) Error() string {
	if e.Instruction == nil {
		if e.Wrapped != nil {
			return fmt.Sprintf("encoding error: %s: %v", e.Message, e.Wrapped)
		}
		return fmt.Sprintf("encoding error: %s", e.Message)
	}

	pos := e.Instruction.Pos
	location := ""
	if pos.Filename != "" {
		location = fmt.Sprintf("%s:%d:%d: ", pos.Filename, pos.Line, pos.Column)
	} else if pos.Line > 0 {
		location = fmt.Sprintf("line %d: ", pos.Line)
	}

	// Build the error message with context
	var msg string
	if e.Wrapped != nil {
		msg = fmt.Sprintf("%s%s: %v", location, e.Message, e.Wrapped)
	} else {
		msg = fmt.Sprintf("%s%s", location, e.Message)
	}

	// Include the raw source line if available for easier debugging
	if e.Instruction.RawLine != "" {
		msg = fmt.Sprintf("%s\n  source: %s", msg, e.Instruction.RawLine)
	}

	return msg
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *EncodingError) Unwrap() error {
	return e.Wrapped
}

// NewEncodingError creates a new EncodingError with instruction context.
func NewEncodingError(inst *parser.Instruction, message string) *EncodingError {
	return &EncodingError{
		Instruction: inst,
		Message:     message,
		Wrapped:     nil,
	}
}

// WrapEncodingError wraps an existing error with instruction context.
// If the error is already an EncodingError, it returns it unchanged.
// If err is nil, returns nil.
func WrapEncodingError(inst *parser.Instruction, err error) error {
	if err == nil {
		return nil
	}

	// Don't double-wrap EncodingErrors
	if _, ok := err.(*EncodingError); ok {
		return err
	}

	return &EncodingError{
		Instruction: inst,
		Message:     "failed to encode instruction",
		Wrapped:     err,
	}
}
