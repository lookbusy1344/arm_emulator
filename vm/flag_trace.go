package vm

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// FlagChangeEntry represents a single flag change event
type FlagChangeEntry struct {
	Sequence    uint64 // Instruction sequence number
	PC          uint32 // Program counter
	Instruction string // Instruction that changed flags
	OldFlags    CPSR   // Flags before instruction
	NewFlags    CPSR   // Flags after instruction
	Changed     string // Which flags changed (e.g., "NZ")
}

// FlagTrace tracks CPSR flag changes
type FlagTrace struct {
	Enabled bool
	Writer  io.Writer

	// Tracking
	entries    []FlagChangeEntry
	maxEntries int
	lastFlags  CPSR

	// Statistics
	totalChanges uint64
	nChanges     uint64 // Negative flag changes
	zChanges     uint64 // Zero flag changes
	cChanges     uint64 // Carry flag changes
	vChanges     uint64 // Overflow flag changes

	// Symbol resolution
	symbols *SymbolResolver // Symbol resolver for address annotation
}

// NewFlagTrace creates a new flag trace tracker
func NewFlagTrace(writer io.Writer) *FlagTrace {
	return &FlagTrace{
		Enabled:    true,
		Writer:     writer,
		entries:    make([]FlagChangeEntry, 0, 1000),
		maxEntries: 100000,
	}
}

// LoadSymbols loads a symbol table for address annotation
func (f *FlagTrace) LoadSymbols(symbols map[string]uint32) {
	f.symbols = NewSymbolResolver(symbols)
}

// Start starts flag tracing
func (f *FlagTrace) Start(initialFlags CPSR) {
	f.entries = f.entries[:0]
	f.lastFlags = initialFlags
	f.totalChanges = 0
	f.nChanges = 0
	f.zChanges = 0
	f.cChanges = 0
	f.vChanges = 0
}

// RecordFlags records the current flag state
func (f *FlagTrace) RecordFlags(sequence uint64, pc uint32, instruction string, newFlags CPSR) {
	if !f.Enabled {
		return
	}

	// Check if any flags changed
	changed := f.detectChanges(f.lastFlags, newFlags)
	if changed == "" {
		// No changes, don't record
		return
	}

	if f.maxEntries > 0 && len(f.entries) >= f.maxEntries {
		return
	}

	entry := FlagChangeEntry{
		Sequence:    sequence,
		PC:          pc,
		Instruction: instruction,
		OldFlags:    f.lastFlags,
		NewFlags:    newFlags,
		Changed:     changed,
	}

	f.entries = append(f.entries, entry)
	f.updateStatistics(f.lastFlags, newFlags)
	f.lastFlags = newFlags
	f.totalChanges++
}

// detectChanges returns a string indicating which flags changed
func (f *FlagTrace) detectChanges(old, new CPSR) string {
	var changes []string

	if old.N != new.N {
		changes = append(changes, "N")
	}
	if old.Z != new.Z {
		changes = append(changes, "Z")
	}
	if old.C != new.C {
		changes = append(changes, "C")
	}
	if old.V != new.V {
		changes = append(changes, "V")
	}

	return strings.Join(changes, "")
}

// updateStatistics updates flag change statistics
func (f *FlagTrace) updateStatistics(old, new CPSR) {
	if old.N != new.N {
		f.nChanges++
	}
	if old.Z != new.Z {
		f.zChanges++
	}
	if old.C != new.C {
		f.cChanges++
	}
	if old.V != new.V {
		f.vChanges++
	}
}

// GetEntries returns all flag trace entries
func (f *FlagTrace) GetEntries() []FlagChangeEntry {
	return f.entries
}

// Flush writes flag trace report to the writer
func (f *FlagTrace) Flush() error {
	if f.Writer == nil {
		return nil
	}

	// Write header using strings.Builder for efficiency
	var header strings.Builder
	header.WriteString("Flag Change Trace Report\n")
	header.WriteString("========================\n\n")
	header.WriteString("Statistics:\n")
	header.WriteString(fmt.Sprintf("  Total Changes:    %d\n", f.totalChanges))
	header.WriteString(fmt.Sprintf("  N flag changes:   %d\n", f.nChanges))
	header.WriteString(fmt.Sprintf("  Z flag changes:   %d\n", f.zChanges))
	header.WriteString(fmt.Sprintf("  C flag changes:   %d\n", f.cChanges))
	header.WriteString(fmt.Sprintf("  V flag changes:   %d\n\n", f.vChanges))

	if _, err := f.Writer.Write([]byte(header.String())); err != nil {
		return err
	}

	// Write detailed trace
	if _, err := f.Writer.Write([]byte("Flag Changes:\n")); err != nil {
		return err
	}
	if _, err := f.Writer.Write([]byte("-------------\n")); err != nil {
		return err
	}

	for _, entry := range f.entries {
		line := f.formatEntry(entry)
		if _, err := f.Writer.Write([]byte(line)); err != nil {
			return err
		}
	}

	return nil
}

// formatEntry formats a flag change entry for output
func (f *FlagTrace) formatEntry(entry FlagChangeEntry) string {
	oldStr := f.formatFlags(entry.OldFlags)

	// Highlight changed flags
	highlightedNew := f.highlightChanges(entry.NewFlags, entry.Changed)

	// Use symbol-aware formatting if symbols are available
	pcStr := fmt.Sprintf("0x%04X", entry.PC)
	if f.symbols != nil && f.symbols.HasSymbols() {
		pcStr = f.symbols.FormatAddressCompact(entry.PC)
	}

	line := fmt.Sprintf("[%06d] %-20s: %-30s  %s -> %s  (changed: %s)\n",
		entry.Sequence,
		pcStr,
		entry.Instruction,
		oldStr,
		highlightedNew,
		entry.Changed)

	return line
}

// formatFlags formats CPSR flags as a string
func (f *FlagTrace) formatFlags(flags CPSR) string {
	// Use a fixed-size byte slice for efficiency (4 flags)
	result := make([]byte, 4)
	if flags.N {
		result[0] = 'N'
	} else {
		result[0] = '-'
	}
	if flags.Z {
		result[1] = 'Z'
	} else {
		result[1] = '-'
	}
	if flags.C {
		result[2] = 'C'
	} else {
		result[2] = '-'
	}
	if flags.V {
		result[3] = 'V'
	} else {
		result[3] = '-'
	}
	return string(result)
}

// highlightChanges highlights changed flags in the new flags string
func (f *FlagTrace) highlightChanges(flags CPSR, changed string) string {
	var sb strings.Builder
	sb.Grow(8) // Max 4 flags * 2 chars each

	// Helper to check if flag changed
	hasN := strings.Contains(changed, "N")
	hasZ := strings.Contains(changed, "Z")
	hasC := strings.Contains(changed, "C")
	hasV := strings.Contains(changed, "V")

	// N flag
	if flags.N {
		sb.WriteByte('N')
	} else {
		sb.WriteByte('-')
	}
	if hasN {
		sb.WriteByte('*')
	}

	// Z flag
	if flags.Z {
		sb.WriteByte('Z')
	} else {
		sb.WriteByte('-')
	}
	if hasZ {
		sb.WriteByte('*')
	}

	// C flag
	if flags.C {
		sb.WriteByte('C')
	} else {
		sb.WriteByte('-')
	}
	if hasC {
		sb.WriteByte('*')
	}

	// V flag
	if flags.V {
		sb.WriteByte('V')
	} else {
		sb.WriteByte('-')
	}
	if hasV {
		sb.WriteByte('*')
	}

	return sb.String()
}

// ExportJSON exports flag trace data as JSON
func (f *FlagTrace) ExportJSON(w io.Writer) error {
	data := map[string]interface{}{
		"total_changes": f.totalChanges,
		"n_changes":     f.nChanges,
		"z_changes":     f.zChanges,
		"c_changes":     f.cChanges,
		"v_changes":     f.vChanges,
		"entries":       f.entries,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// String returns a formatted string representation
func (f *FlagTrace) String() string {
	var sb strings.Builder

	sb.WriteString("Flag Change Summary\n")
	sb.WriteString("===================\n\n")

	sb.WriteString(fmt.Sprintf("Total Changes:      %d\n", f.totalChanges))
	sb.WriteString(fmt.Sprintf("N flag changes:     %d\n", f.nChanges))
	sb.WriteString(fmt.Sprintf("Z flag changes:     %d\n", f.zChanges))
	sb.WriteString(fmt.Sprintf("C flag changes:     %d\n", f.cChanges))
	sb.WriteString(fmt.Sprintf("V flag changes:     %d\n", f.vChanges))

	return sb.String()
}
