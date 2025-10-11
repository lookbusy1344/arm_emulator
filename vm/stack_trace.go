package vm

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// StackOperation represents a stack operation type
type StackOperation string

const (
	StackPush StackOperation = "PUSH"
	StackPop  StackOperation = "POP"
	StackMove StackOperation = "MOVE" // SP register update
)

// StackTraceEntry represents a single stack operation
type StackTraceEntry struct {
	Sequence  uint64         // Instruction sequence number
	PC        uint32         // Program counter
	Operation StackOperation // Operation type
	OldSP     uint32         // Stack pointer before operation
	NewSP     uint32         // Stack pointer after operation
	Value     uint32         // Value pushed/popped (if applicable)
	Address   uint32         // Memory address accessed
	Size      uint32         // Number of bytes in operation
	Register  string         // Register involved (if applicable)
}

// StackTrace tracks stack operations and detects issues
type StackTrace struct {
	Enabled bool
	Writer  io.Writer

	// Stack configuration
	StackBase uint32 // Bottom of stack (highest address)
	StackTop  uint32 // Top of stack (lowest valid address)

	// Tracking
	entries    []StackTraceEntry
	maxEntries int
	currentSP  uint32
	minSP      uint32 // Lowest SP seen (stack growth limit)
	maxSP      uint32 // Highest SP seen

	// Statistics
	totalPushes    uint64
	totalPops      uint64
	totalBytes     uint64
	overflowCount  uint64 // Number of stack overflow events
	underflowCount uint64 // Number of stack underflow events
}

// NewStackTrace creates a new stack trace tracker
func NewStackTrace(writer io.Writer, stackBase, stackTop uint32) *StackTrace {
	return &StackTrace{
		Enabled:    true,
		Writer:     writer,
		StackBase:  stackBase,
		StackTop:   stackTop,
		entries:    make([]StackTraceEntry, 0, 1000),
		maxEntries: 100000,
		currentSP:  stackBase,
		minSP:      stackBase,
		maxSP:      stackBase,
	}
}

// Start starts stack tracing
func (s *StackTrace) Start(initialSP uint32) {
	s.entries = s.entries[:0]
	s.currentSP = initialSP
	s.minSP = initialSP
	s.maxSP = initialSP
	s.totalPushes = 0
	s.totalPops = 0
	s.totalBytes = 0
	s.overflowCount = 0
	s.underflowCount = 0
}

// RecordPush records a push operation
func (s *StackTrace) RecordPush(sequence uint64, pc, oldSP, newSP, value, address uint32, register string) {
	if !s.Enabled {
		return
	}

	if s.maxEntries > 0 && len(s.entries) >= s.maxEntries {
		return
	}

	size := oldSP - newSP

	entry := StackTraceEntry{
		Sequence:  sequence,
		PC:        pc,
		Operation: StackPush,
		OldSP:     oldSP,
		NewSP:     newSP,
		Value:     value,
		Address:   address,
		Size:      size,
		Register:  register,
	}

	s.entries = append(s.entries, entry)
	s.updateTracking(newSP, size)
	s.totalPushes++

	// Check for overflow (SP went below valid range)
	if newSP < s.StackTop {
		s.overflowCount++
	}
}

// RecordPop records a pop operation
func (s *StackTrace) RecordPop(sequence uint64, pc, oldSP, newSP, value, address uint32, register string) {
	if !s.Enabled {
		return
	}

	if s.maxEntries > 0 && len(s.entries) >= s.maxEntries {
		return
	}

	size := newSP - oldSP

	entry := StackTraceEntry{
		Sequence:  sequence,
		PC:        pc,
		Operation: StackPop,
		OldSP:     oldSP,
		NewSP:     newSP,
		Value:     value,
		Address:   address,
		Size:      size,
		Register:  register,
	}

	s.entries = append(s.entries, entry)
	s.updateTracking(newSP, size)
	s.totalPops++

	// Check for underflow (SP went above base)
	if newSP > s.StackBase {
		s.underflowCount++
	}
}

// RecordSPMove records a direct SP register update
func (s *StackTrace) RecordSPMove(sequence uint64, pc, oldSP, newSP uint32) {
	if !s.Enabled {
		return
	}

	if s.maxEntries > 0 && len(s.entries) >= s.maxEntries {
		return
	}

	var size uint32
	if newSP < oldSP {
		size = oldSP - newSP
	} else {
		size = newSP - oldSP
	}

	entry := StackTraceEntry{
		Sequence:  sequence,
		PC:        pc,
		Operation: StackMove,
		OldSP:     oldSP,
		NewSP:     newSP,
		Size:      size,
	}

	s.entries = append(s.entries, entry)
	s.updateTracking(newSP, size)
}

// updateTracking updates internal tracking state
func (s *StackTrace) updateTracking(newSP, bytes uint32) {
	s.currentSP = newSP
	s.totalBytes += uint64(bytes)

	if newSP < s.minSP {
		s.minSP = newSP
	}
	if newSP > s.maxSP {
		s.maxSP = newSP
	}
}

// GetStackUsage returns the maximum stack usage in bytes
func (s *StackTrace) GetStackUsage() uint32 {
	if s.StackBase >= s.minSP {
		return s.StackBase - s.minSP
	}
	return 0
}

// GetStackDepth returns current stack depth in bytes
func (s *StackTrace) GetStackDepth() uint32 {
	if s.StackBase >= s.currentSP {
		return s.StackBase - s.currentSP
	}
	return 0
}

// HasOverflow returns whether stack overflow was detected
func (s *StackTrace) HasOverflow() bool {
	return s.overflowCount > 0
}

// HasUnderflow returns whether stack underflow was detected
func (s *StackTrace) HasUnderflow() bool {
	return s.underflowCount > 0
}

// GetEntries returns all stack trace entries
func (s *StackTrace) GetEntries() []StackTraceEntry {
	return s.entries
}

// Flush writes stack trace report to the writer
func (s *StackTrace) Flush() error {
	if s.Writer == nil {
		return nil
	}

	// Write header
	header := "Stack Trace Report\n"
	header += "==================\n\n"

	header += fmt.Sprintf("Stack Configuration:\n")
	header += fmt.Sprintf("  Base (high):      0x%08X\n", s.StackBase)
	header += fmt.Sprintf("  Top (low):        0x%08X\n", s.StackTop)
	header += fmt.Sprintf("  Total Size:       %d bytes\n\n", s.StackBase-s.StackTop)

	header += fmt.Sprintf("Stack Usage:\n")
	header += fmt.Sprintf("  Max Depth:        %d bytes\n", s.GetStackUsage())
	header += fmt.Sprintf("  Current Depth:    %d bytes\n", s.GetStackDepth())
	header += fmt.Sprintf("  Min SP:           0x%08X\n", s.minSP)
	header += fmt.Sprintf("  Max SP:           0x%08X\n\n", s.maxSP)

	header += fmt.Sprintf("Operations:\n")
	header += fmt.Sprintf("  Total Pushes:     %d\n", s.totalPushes)
	header += fmt.Sprintf("  Total Pops:       %d\n", s.totalPops)
	header += fmt.Sprintf("  Total Bytes:      %d\n\n", s.totalBytes)

	if s.overflowCount > 0 || s.underflowCount > 0 {
		header += "WARNINGS:\n"
		if s.overflowCount > 0 {
			header += fmt.Sprintf("  ⚠️  Stack Overflow detected: %d times (SP < 0x%08X)\n", s.overflowCount, s.StackTop)
		}
		if s.underflowCount > 0 {
			header += fmt.Sprintf("  ⚠️  Stack Underflow detected: %d times (SP > 0x%08X)\n", s.underflowCount, s.StackBase)
		}
		header += "\n"
	}

	if _, err := s.Writer.Write([]byte(header)); err != nil {
		return err
	}

	// Write detailed trace
	if _, err := s.Writer.Write([]byte("Stack Operations:\n")); err != nil {
		return err
	}
	if _, err := s.Writer.Write([]byte("-----------------\n")); err != nil {
		return err
	}

	for _, entry := range s.entries {
		line := s.formatEntry(entry)
		if _, err := s.Writer.Write([]byte(line)); err != nil {
			return err
		}
	}

	return nil
}

// formatEntry formats a stack trace entry for output
func (s *StackTrace) formatEntry(entry StackTraceEntry) string {
	var line string

	switch entry.Operation {
	case StackPush:
		line = fmt.Sprintf("[%06d] 0x%04X: PUSH %-3s  SP: 0x%08X -> 0x%08X  [0x%08X] = 0x%08X  (%d bytes)",
			entry.Sequence, entry.PC, entry.Register,
			entry.OldSP, entry.NewSP, entry.Address, entry.Value, entry.Size)

		// Warn if overflow
		if entry.NewSP < s.StackTop {
			line += " ⚠️ OVERFLOW"
		}

	case StackPop:
		line = fmt.Sprintf("[%06d] 0x%04X: POP  %-3s  SP: 0x%08X -> 0x%08X  [0x%08X] = 0x%08X  (%d bytes)",
			entry.Sequence, entry.PC, entry.Register,
			entry.OldSP, entry.NewSP, entry.Address, entry.Value, entry.Size)

		// Warn if underflow
		if entry.NewSP > s.StackBase {
			line += " ⚠️ UNDERFLOW"
		}

	case StackMove:
		direction := "grow"
		if entry.NewSP > entry.OldSP {
			direction = "shrink"
		}
		line = fmt.Sprintf("[%06d] 0x%04X: MOVE      SP: 0x%08X -> 0x%08X  (%s by %d bytes)",
			entry.Sequence, entry.PC, entry.OldSP, entry.NewSP, direction, entry.Size)
	}

	line += "\n"
	return line
}

// ExportJSON exports stack trace data as JSON
func (s *StackTrace) ExportJSON(w io.Writer) error {
	data := map[string]interface{}{
		"stack_base":      s.StackBase,
		"stack_top":       s.StackTop,
		"stack_size":      s.StackBase - s.StackTop,
		"max_usage":       s.GetStackUsage(),
		"current_depth":   s.GetStackDepth(),
		"min_sp":          s.minSP,
		"max_sp":          s.maxSP,
		"total_pushes":    s.totalPushes,
		"total_pops":      s.totalPops,
		"total_bytes":     s.totalBytes,
		"overflow_count":  s.overflowCount,
		"underflow_count": s.underflowCount,
		"entries":         s.entries,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// String returns a formatted string representation
func (s *StackTrace) String() string {
	var sb strings.Builder

	sb.WriteString("Stack Trace Summary\n")
	sb.WriteString("===================\n\n")

	sb.WriteString(fmt.Sprintf("Max Stack Usage:    %d bytes\n", s.GetStackUsage()))
	sb.WriteString(fmt.Sprintf("Current Depth:      %d bytes\n", s.GetStackDepth()))
	sb.WriteString(fmt.Sprintf("Total Pushes:       %d\n", s.totalPushes))
	sb.WriteString(fmt.Sprintf("Total Pops:         %d\n", s.totalPops))

	if s.overflowCount > 0 || s.underflowCount > 0 {
		sb.WriteString("\nWARNINGS:\n")
		if s.overflowCount > 0 {
			sb.WriteString(fmt.Sprintf("  ⚠️  Stack Overflow:  %d times\n", s.overflowCount))
		}
		if s.underflowCount > 0 {
			sb.WriteString(fmt.Sprintf("  ⚠️  Stack Underflow: %d times\n", s.underflowCount))
		}
	}

	return sb.String()
}
