package parser

import (
	"fmt"
	"strings"
)

// Position represents a location in the source file
type Position struct {
	Filename string
	Line     int
	Column   int
}

func (p Position) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}

// Error represents a parse error with position information
type Error struct {
	Pos     Position
	Message string
	Context string // The line of code where the error occurred
	Kind    ErrorKind
}

// ErrorKind categorizes the type of error
type ErrorKind int

const (
	ErrorSyntax ErrorKind = iota
	ErrorUndefinedLabel
	ErrorDuplicateLabel
	ErrorInvalidDirective
	ErrorInvalidInstruction
	ErrorInvalidOperand
	ErrorCircularInclude
	ErrorMacroExpansion
	ErrorFileIO
)

func (e *Error) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s: error: %s\n", e.Pos, e.Message))

	if e.Context != "" {
		sb.WriteString(fmt.Sprintf("    %s\n", e.Context))
	}

	return sb.String()
}

// NewError creates a new parser error
func NewError(pos Position, kind ErrorKind, message string) *Error {
	return &Error{
		Pos:     pos,
		Message: message,
		Kind:    kind,
	}
}

// NewErrorWithContext creates a new parser error with source context
func NewErrorWithContext(pos Position, kind ErrorKind, message, context string) *Error {
	return &Error{
		Pos:     pos,
		Message: message,
		Context: context,
		Kind:    kind,
	}
}

// Warning represents a non-fatal parse warning
type Warning struct {
	Pos     Position
	Message string
}

func (w *Warning) String() string {
	return fmt.Sprintf("%s: warning: %s", w.Pos, w.Message)
}

// ErrorList collects multiple errors and warnings
type ErrorList struct {
	Errors   []*Error
	Warnings []*Warning
}

// AddError adds an error to the list
func (el *ErrorList) AddError(err *Error) {
	el.Errors = append(el.Errors, err)
}

// AddWarning adds a warning to the list
func (el *ErrorList) AddWarning(warn *Warning) {
	el.Warnings = append(el.Warnings, warn)
}

// HasErrors returns true if there are any errors
func (el *ErrorList) HasErrors() bool {
	return len(el.Errors) > 0
}

// Error implements the error interface
func (el *ErrorList) Error() string {
	if !el.HasErrors() {
		return ""
	}

	var sb strings.Builder
	for _, err := range el.Errors {
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// PrintWarnings prints all warnings
func (el *ErrorList) PrintWarnings() string {
	if len(el.Warnings) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, warn := range el.Warnings {
		sb.WriteString(warn.String())
		sb.WriteString("\n")
	}
	return sb.String()
}
