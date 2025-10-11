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

	// Write header
	header := "Flag Change Trace Report\n"
	header += "========================\n\n"

	header += fmt.Sprintf("Statistics:\n")
	header += fmt.Sprintf("  Total Changes:    %d\n", f.totalChanges)
	header += fmt.Sprintf("  N flag changes:   %d\n", f.nChanges)
	header += fmt.Sprintf("  Z flag changes:   %d\n", f.zChanges)
	header += fmt.Sprintf("  C flag changes:   %d\n", f.cChanges)
	header += fmt.Sprintf("  V flag changes:   %d\n\n", f.vChanges)

	if _, err := f.Writer.Write([]byte(header)); err != nil {
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

	line := fmt.Sprintf("[%06d] 0x%04X: %-30s  %s -> %s  (changed: %s)\n",
		entry.Sequence,
		entry.PC,
		entry.Instruction,
		oldStr,
		highlightedNew,
		entry.Changed)

	return line
}

// formatFlags formats CPSR flags as a string
func (f *FlagTrace) formatFlags(flags CPSR) string {
	result := ""
	if flags.N {
		result += "N"
	} else {
		result += "-"
	}
	if flags.Z {
		result += "Z"
	} else {
		result += "-"
	}
	if flags.C {
		result += "C"
	} else {
		result += "-"
	}
	if flags.V {
		result += "V"
	} else {
		result += "-"
	}
	return result
}

// highlightChanges highlights changed flags in the new flags string
func (f *FlagTrace) highlightChanges(flags CPSR, changed string) string {
	result := ""

	// N flag
	if flags.N {
		if strings.Contains(changed, "N") {
			result += "N*"
		} else {
			result += "N"
		}
	} else {
		if strings.Contains(changed, "N") {
			result += "-*"
		} else {
			result += "-"
		}
	}

	// Z flag
	if flags.Z {
		if strings.Contains(changed, "Z") {
			result += "Z*"
		} else {
			result += "Z"
		}
	} else {
		if strings.Contains(changed, "Z") {
			result += "-*"
		} else {
			result += "-"
		}
	}

	// C flag
	if flags.C {
		if strings.Contains(changed, "C") {
			result += "C*"
		} else {
			result += "C"
		}
	} else {
		if strings.Contains(changed, "C") {
			result += "-*"
		} else {
			result += "-"
		}
	}

	// V flag
	if flags.V {
		if strings.Contains(changed, "V") {
			result += "V*"
		} else {
			result += "V"
		}
	} else {
		if strings.Contains(changed, "V") {
			result += "-*"
		} else {
			result += "-"
		}
	}

	return result
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
