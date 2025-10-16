package vm

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// RegisterAccessType represents the type of register access
type RegisterAccessType string

const (
	RegisterRead  RegisterAccessType = "READ"
	RegisterWrite RegisterAccessType = "WRITE"
)

// RegisterAccessEntry represents a single register access
type RegisterAccessEntry struct {
	Sequence   uint64             // Instruction sequence number
	PC         uint32             // Program counter
	Register   string             // Register name (R0-R15, CPSR)
	AccessType RegisterAccessType // READ or WRITE
	Value      uint32             // Value read or written
	OldValue   uint32             // Previous value (for writes)
}

// RegisterStats contains statistics for a single register
type RegisterStats struct {
	RegisterName string // Register name
	ReadCount    uint64 // Number of reads
	WriteCount   uint64 // Number of writes
	FirstRead    uint64 // Sequence number of first read (0 if never read)
	FirstWrite   uint64 // Sequence number of first write (0 if never written)
	LastRead     uint64 // Sequence number of last read
	LastWrite    uint64 // Sequence number of last write
	LastValue    uint32 // Most recent value
	UniqueValues uint64 // Number of unique values written
	valuesSeen   map[uint32]bool
}

// NewRegisterStats creates a new register statistics tracker
func NewRegisterStats(name string) *RegisterStats {
	return &RegisterStats{
		RegisterName: name,
		valuesSeen:   make(map[uint32]bool),
	}
}

// RecordRead records a read operation
func (r *RegisterStats) RecordRead(sequence uint64, value uint32) {
	r.ReadCount++
	if r.FirstRead == 0 {
		r.FirstRead = sequence
	}
	r.LastRead = sequence
	r.LastValue = value
}

// RecordWrite records a write operation
func (r *RegisterStats) RecordWrite(sequence uint64, value uint32) {
	r.WriteCount++
	if r.FirstWrite == 0 {
		r.FirstWrite = sequence
	}
	r.LastWrite = sequence
	r.LastValue = value

	if !r.valuesSeen[value] {
		r.valuesSeen[value] = true
		r.UniqueValues++
	}
}

// RegisterTrace tracks register access patterns
type RegisterTrace struct {
	Enabled bool
	Writer  io.Writer

	// Tracking
	entries       []RegisterAccessEntry
	maxEntries    int
	registerStats map[string]*RegisterStats // register name -> stats

	// Current state (for detecting changes)
	lastRegValues map[string]uint32 // register name -> last known value

	// Statistics
	totalReads  uint64
	totalWrites uint64

	// Symbol resolution
	symbols *SymbolResolver // Symbol resolver for address annotation
}

// NewRegisterTrace creates a new register trace tracker
func NewRegisterTrace(writer io.Writer) *RegisterTrace {
	return &RegisterTrace{
		Enabled:       true,
		Writer:        writer,
		entries:       make([]RegisterAccessEntry, 0, 1000),
		maxEntries:    100000,
		registerStats: make(map[string]*RegisterStats),
		lastRegValues: make(map[string]uint32),
	}
}

// LoadSymbols loads a symbol table for address annotation
func (r *RegisterTrace) LoadSymbols(symbols map[string]uint32) {
	r.symbols = NewSymbolResolver(symbols)
}

// Start starts register tracing
func (r *RegisterTrace) Start() {
	r.entries = r.entries[:0]
	r.registerStats = make(map[string]*RegisterStats)
	r.lastRegValues = make(map[string]uint32)
	r.totalReads = 0
	r.totalWrites = 0
}

// RecordRead records a register read
func (r *RegisterTrace) RecordRead(sequence uint64, pc uint32, registerName string, value uint32) {
	if !r.Enabled {
		return
	}

	// Update statistics
	stats := r.getOrCreateStats(registerName)
	stats.RecordRead(sequence, value)
	r.totalReads++

	// Record entry if within limit
	if r.maxEntries > 0 && len(r.entries) >= r.maxEntries {
		return
	}

	entry := RegisterAccessEntry{
		Sequence:   sequence,
		PC:         pc,
		Register:   registerName,
		AccessType: RegisterRead,
		Value:      value,
	}
	r.entries = append(r.entries, entry)

	// Update last known value
	r.lastRegValues[registerName] = value
}

// RecordWrite records a register write
func (r *RegisterTrace) RecordWrite(sequence uint64, pc uint32, registerName string, oldValue, newValue uint32) {
	if !r.Enabled {
		return
	}

	// Update statistics
	stats := r.getOrCreateStats(registerName)
	stats.RecordWrite(sequence, newValue)
	r.totalWrites++

	// Record entry if within limit
	if r.maxEntries > 0 && len(r.entries) >= r.maxEntries {
		return
	}

	entry := RegisterAccessEntry{
		Sequence:   sequence,
		PC:         pc,
		Register:   registerName,
		AccessType: RegisterWrite,
		Value:      newValue,
		OldValue:   oldValue,
	}
	r.entries = append(r.entries, entry)

	// Update last known value
	r.lastRegValues[registerName] = newValue
}

// getOrCreateStats gets or creates statistics for a register
func (r *RegisterTrace) getOrCreateStats(registerName string) *RegisterStats {
	if stats, exists := r.registerStats[registerName]; exists {
		return stats
	}
	stats := NewRegisterStats(registerName)
	r.registerStats[registerName] = stats
	return stats
}

// GetStats returns statistics for a register
func (r *RegisterTrace) GetStats(registerName string) *RegisterStats {
	return r.registerStats[registerName]
}

// GetAllStats returns statistics for all registers, sorted by name
func (r *RegisterTrace) GetAllStats() []*RegisterStats {
	stats := make([]*RegisterStats, 0, len(r.registerStats))
	for _, s := range r.registerStats {
		stats = append(stats, s)
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].RegisterName < stats[j].RegisterName
	})
	return stats
}

// GetHotRegisters returns the most frequently accessed registers
func (r *RegisterTrace) GetHotRegisters(limit int) []*RegisterStats {
	stats := r.GetAllStats()

	// Sort by total accesses (reads + writes)
	sort.Slice(stats, func(i, j int) bool {
		totalI := stats[i].ReadCount + stats[i].WriteCount
		totalJ := stats[j].ReadCount + stats[j].WriteCount
		return totalI > totalJ
	})

	if limit > 0 && limit < len(stats) {
		return stats[:limit]
	}
	return stats
}

// GetUnusedRegisters returns registers that were never accessed
func (r *RegisterTrace) GetUnusedRegisters() []string {
	// Standard ARM registers
	allRegisters := []string{
		"R0", "R1", "R2", "R3", "R4", "R5", "R6", "R7",
		"R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15",
	}

	unused := make([]string, 0)
	for _, reg := range allRegisters {
		if stats := r.registerStats[reg]; stats == nil || (stats.ReadCount == 0 && stats.WriteCount == 0) {
			unused = append(unused, reg)
		}
	}
	return unused
}

// DetectReadBeforeWrite detects registers that were read before being written
func (r *RegisterTrace) DetectReadBeforeWrite() []string {
	result := make([]string, 0)
	for _, stats := range r.registerStats {
		if stats.FirstRead > 0 && (stats.FirstWrite == 0 || stats.FirstRead < stats.FirstWrite) {
			result = append(result, stats.RegisterName)
		}
	}
	sort.Strings(result)
	return result
}

// GetEntries returns all register access entries
func (r *RegisterTrace) GetEntries() []RegisterAccessEntry {
	return r.entries
}

// Flush writes register trace report to the writer
func (r *RegisterTrace) Flush() error {
	if r.Writer == nil {
		return nil
	}

	var sb strings.Builder

	// Header
	sb.WriteString("Register Access Pattern Analysis\n")
	sb.WriteString("=================================\n\n")

	// Summary statistics
	sb.WriteString(fmt.Sprintf("Total Reads:  %d\n", r.totalReads))
	sb.WriteString(fmt.Sprintf("Total Writes: %d\n", r.totalWrites))
	sb.WriteString(fmt.Sprintf("Total Entries: %d\n", len(r.entries)))
	sb.WriteString(fmt.Sprintf("Registers Tracked: %d\n\n", len(r.registerStats)))

	// Hot registers
	sb.WriteString("Hot Registers (Top 10 by Total Accesses):\n")
	sb.WriteString("------------------------------------------\n")
	hotRegs := r.GetHotRegisters(10)
	for i, stats := range hotRegs {
		total := stats.ReadCount + stats.WriteCount
		sb.WriteString(fmt.Sprintf("%2d. %-4s: %6d accesses (R:%6d W:%6d) [%d unique values]\n",
			i+1, stats.RegisterName, total, stats.ReadCount, stats.WriteCount, stats.UniqueValues))
	}
	sb.WriteString("\n")

	// Unused registers
	unused := r.GetUnusedRegisters()
	if len(unused) > 0 {
		sb.WriteString("Unused Registers:\n")
		sb.WriteString("-----------------\n")
		sb.WriteString(strings.Join(unused, ", "))
		sb.WriteString("\n\n")
	}

	// Read-before-write detection
	rbw := r.DetectReadBeforeWrite()
	if len(rbw) > 0 {
		sb.WriteString("Registers Read Before Write (potential uninitialized use):\n")
		sb.WriteString("----------------------------------------------------------\n")
		for _, reg := range rbw {
			stats := r.registerStats[reg]
			sb.WriteString(fmt.Sprintf("  %s: first read at #%d", reg, stats.FirstRead))
			if stats.FirstWrite > 0 {
				sb.WriteString(fmt.Sprintf(", first write at #%d", stats.FirstWrite))
			} else {
				sb.WriteString(", never written")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Detailed statistics per register
	sb.WriteString("Detailed Register Statistics:\n")
	sb.WriteString("-----------------------------\n")
	allStats := r.GetAllStats()
	for _, stats := range allStats {
		if stats.ReadCount > 0 || stats.WriteCount > 0 {
			sb.WriteString(fmt.Sprintf("%-4s: R:%6d W:%6d",
				stats.RegisterName, stats.ReadCount, stats.WriteCount))

			if stats.FirstRead > 0 {
				sb.WriteString(fmt.Sprintf(" | First R:#%d", stats.FirstRead))
			}
			if stats.FirstWrite > 0 {
				sb.WriteString(fmt.Sprintf(" | First W:#%d", stats.FirstWrite))
			}
			if stats.UniqueValues > 0 {
				sb.WriteString(fmt.Sprintf(" | Unique:%d", stats.UniqueValues))
			}
			sb.WriteString(fmt.Sprintf(" | Last:0x%08X", stats.LastValue))
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	// Write to output
	_, err := r.Writer.Write([]byte(sb.String()))
	return err
}

// ExportJSON exports register trace data as JSON
func (r *RegisterTrace) ExportJSON(w io.Writer) error {
	// Prepare stats for JSON export
	statsMap := make(map[string]interface{})
	for name, stats := range r.registerStats {
		statsMap[name] = map[string]interface{}{
			"read_count":    stats.ReadCount,
			"write_count":   stats.WriteCount,
			"first_read":    stats.FirstRead,
			"first_write":   stats.FirstWrite,
			"last_read":     stats.LastRead,
			"last_write":    stats.LastWrite,
			"last_value":    stats.LastValue,
			"unique_values": stats.UniqueValues,
		}
	}

	data := map[string]interface{}{
		"total_reads":       r.totalReads,
		"total_writes":      r.totalWrites,
		"total_entries":     len(r.entries),
		"registers_tracked": len(r.registerStats),
		"register_stats":    statsMap,
		"hot_registers":     r.GetHotRegisters(10),
		"unused_registers":  r.GetUnusedRegisters(),
		"read_before_write": r.DetectReadBeforeWrite(),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// String returns a formatted string representation
func (r *RegisterTrace) String() string {
	var sb strings.Builder

	sb.WriteString("Register Access Summary\n")
	sb.WriteString("=======================\n\n")
	sb.WriteString(fmt.Sprintf("Total Reads:  %d\n", r.totalReads))
	sb.WriteString(fmt.Sprintf("Total Writes: %d\n", r.totalWrites))
	sb.WriteString(fmt.Sprintf("Registers Tracked: %d\n", len(r.registerStats)))

	return sb.String()
}
