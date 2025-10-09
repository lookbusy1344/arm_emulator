package vm

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// TraceEntry represents a single execution trace entry
type TraceEntry struct {
	Sequence        uint64            // Instruction sequence number
	Address         uint32            // Instruction address
	Opcode          uint32            // Instruction opcode
	Disassembly     string            // Disassembled instruction
	RegisterChanges map[string]uint32 // Register changes (name -> new value)
	Flags           CPSR              // CPSR flags after execution
	Duration        time.Duration     // Execution time
}

// ExecutionTrace manages execution tracing
type ExecutionTrace struct {
	Enabled       bool
	Writer        io.Writer
	FilterRegs    map[string]bool // Registers to track (empty = all)
	IncludeFlags  bool
	IncludeTiming bool
	MaxEntries    int

	entries      []TraceEntry
	startTime    time.Time
	lastSnapshot map[string]uint32 // Previous register values
}

// NewExecutionTrace creates a new execution trace
func NewExecutionTrace(writer io.Writer) *ExecutionTrace {
	return &ExecutionTrace{
		Enabled:       true,
		Writer:        writer,
		FilterRegs:    make(map[string]bool),
		IncludeFlags:  true,
		IncludeTiming: true,
		MaxEntries:    100000,
		entries:       make([]TraceEntry, 0, 1000),
		lastSnapshot:  make(map[string]uint32),
	}
}

// SetFilterRegisters sets which registers to track
// Pass empty slice or nil to track all registers
func (t *ExecutionTrace) SetFilterRegisters(regs []string) {
	t.FilterRegs = make(map[string]bool)
	for _, reg := range regs {
		t.FilterRegs[strings.ToUpper(reg)] = true
	}
}

// Start starts the trace
func (t *ExecutionTrace) Start() {
	t.startTime = time.Now()
	t.entries = t.entries[:0]
	t.lastSnapshot = make(map[string]uint32)
}

// RecordInstruction records an instruction execution
func (t *ExecutionTrace) RecordInstruction(vm *VM, disasm string) {
	if !t.Enabled {
		return
	}

	// Check if we've exceeded max entries
	if t.MaxEntries > 0 && len(t.entries) >= t.MaxEntries {
		return
	}

	entry := TraceEntry{
		Sequence:        vm.CPU.Cycles,
		Address:         vm.CPU.PC - 4, // PC has already advanced
		Opcode:          0,             // Will be filled by caller if needed
		Disassembly:     disasm,
		RegisterChanges: make(map[string]uint32),
		Flags:           vm.CPU.CPSR,
		Duration:        0,
	}

	// Record timing if enabled
	if t.IncludeTiming {
		entry.Duration = time.Since(t.startTime)
	}

	// Track register changes
	currentRegs := map[string]uint32{
		"R0":  vm.CPU.R[0],
		"R1":  vm.CPU.R[1],
		"R2":  vm.CPU.R[2],
		"R3":  vm.CPU.R[3],
		"R4":  vm.CPU.R[4],
		"R5":  vm.CPU.R[5],
		"R6":  vm.CPU.R[6],
		"R7":  vm.CPU.R[7],
		"R8":  vm.CPU.R[8],
		"R9":  vm.CPU.R[9],
		"R10": vm.CPU.R[10],
		"R11": vm.CPU.R[11],
		"R12": vm.CPU.R[12],
		"R13": vm.CPU.R[13],
		"R14": vm.CPU.R[14],
		"R15": vm.CPU.PC,
		"SP":  vm.CPU.R[13],
		"LR":  vm.CPU.R[14],
		"PC":  vm.CPU.PC,
	}

	// Find changed registers
	for name, value := range currentRegs {
		// Apply filter if set
		if len(t.FilterRegs) > 0 && !t.FilterRegs[name] {
			continue
		}

		// Check if register changed
		if oldValue, exists := t.lastSnapshot[name]; !exists || oldValue != value {
			entry.RegisterChanges[name] = value
			t.lastSnapshot[name] = value
		}
	}

	t.entries = append(t.entries, entry)
}

// Flush writes all trace entries to the writer
func (t *ExecutionTrace) Flush() error {
	if t.Writer == nil {
		return nil
	}

	for _, entry := range t.entries {
		if err := t.writeEntry(entry); err != nil {
			return err
		}
	}

	return nil
}

// writeEntry writes a single trace entry
func (t *ExecutionTrace) writeEntry(entry TraceEntry) error {
	// Format: [seq] addr: instruction | changes | flags | time
	line := fmt.Sprintf("[%06d] 0x%04X: %-30s",
		entry.Sequence,
		entry.Address,
		entry.Disassembly)

	// Add register changes
	if len(entry.RegisterChanges) > 0 {
		changes := make([]string, 0, len(entry.RegisterChanges))
		for name, value := range entry.RegisterChanges {
			changes = append(changes, fmt.Sprintf("%s=0x%08X", name, value))
		}
		line += " | " + strings.Join(changes, " ")
	} else {
		line += " | (no changes)"
	}

	// Add flags if enabled
	if t.IncludeFlags {
		flags := ""
		if entry.Flags.N {
			flags += "N"
		} else {
			flags += "-"
		}
		if entry.Flags.Z {
			flags += "Z"
		} else {
			flags += "-"
		}
		if entry.Flags.C {
			flags += "C"
		} else {
			flags += "-"
		}
		if entry.Flags.V {
			flags += "V"
		} else {
			flags += "-"
		}
		line += " | " + flags
	}

	// Add timing if enabled
	if t.IncludeTiming {
		line += fmt.Sprintf(" | %v", entry.Duration)
	}

	line += "\n"

	_, err := t.Writer.Write([]byte(line))
	return err
}

// GetEntries returns all trace entries
func (t *ExecutionTrace) GetEntries() []TraceEntry {
	return t.entries
}

// Clear clears all trace entries
func (t *ExecutionTrace) Clear() {
	t.entries = t.entries[:0]
	t.lastSnapshot = make(map[string]uint32)
}

// MemoryAccessEntry represents a memory access
type MemoryAccessEntry struct {
	Sequence  uint64        // Instruction sequence number
	Address   uint32        // Memory address accessed
	PC        uint32        // Program counter at time of access
	Type      string        // "READ" or "WRITE"
	Size      string        // "BYTE", "HALF", "WORD"
	Value     uint32        // Value read or written
	Timestamp time.Duration // Time since start
}

// MemoryTrace manages memory access tracing
type MemoryTrace struct {
	Enabled    bool
	Writer     io.Writer
	MaxEntries int

	entries   []MemoryAccessEntry
	startTime time.Time
}

// NewMemoryTrace creates a new memory trace
func NewMemoryTrace(writer io.Writer) *MemoryTrace {
	return &MemoryTrace{
		Enabled:    true,
		Writer:     writer,
		MaxEntries: 100000,
		entries:    make([]MemoryAccessEntry, 0, 1000),
	}
}

// Start starts the memory trace
func (t *MemoryTrace) Start() {
	t.startTime = time.Now()
	t.entries = t.entries[:0]
}

// RecordRead records a memory read
func (t *MemoryTrace) RecordRead(sequence uint64, pc, address, value uint32, size string) {
	if !t.Enabled {
		return
	}

	if t.MaxEntries > 0 && len(t.entries) >= t.MaxEntries {
		return
	}

	t.entries = append(t.entries, MemoryAccessEntry{
		Sequence:  sequence,
		Address:   address,
		PC:        pc,
		Type:      "READ",
		Size:      size,
		Value:     value,
		Timestamp: time.Since(t.startTime),
	})
}

// RecordWrite records a memory write
func (t *MemoryTrace) RecordWrite(sequence uint64, pc, address, value uint32, size string) {
	if !t.Enabled {
		return
	}

	if t.MaxEntries > 0 && len(t.entries) >= t.MaxEntries {
		return
	}

	t.entries = append(t.entries, MemoryAccessEntry{
		Sequence:  sequence,
		Address:   address,
		PC:        pc,
		Type:      "WRITE",
		Size:      size,
		Value:     value,
		Timestamp: time.Since(t.startTime),
	})
}

// Flush writes all memory trace entries to the writer
func (t *MemoryTrace) Flush() error {
	if t.Writer == nil {
		return nil
	}

	for _, entry := range t.entries {
		if err := t.writeEntry(entry); err != nil {
			return err
		}
	}

	return nil
}

// writeEntry writes a single memory trace entry
func (t *MemoryTrace) writeEntry(entry MemoryAccessEntry) error {
	// Format: [seq] [TYPE] PC: instruction <- [addr] = value (size)
	var line string
	if entry.Type == "READ" {
		line = fmt.Sprintf("[%06d] [%-5s] 0x%04X <- [0x%08X] = 0x%08X (%s)\n",
			entry.Sequence,
			entry.Type,
			entry.PC,
			entry.Address,
			entry.Value,
			entry.Size)
	} else {
		line = fmt.Sprintf("[%06d] [%-5s] 0x%04X -> [0x%08X] = 0x%08X (%s)\n",
			entry.Sequence,
			entry.Type,
			entry.PC,
			entry.Address,
			entry.Value,
			entry.Size)
	}

	_, err := t.Writer.Write([]byte(line))
	return err
}

// GetEntries returns all memory trace entries
func (t *MemoryTrace) GetEntries() []MemoryAccessEntry {
	return t.entries
}

// Clear clears all memory trace entries
func (t *MemoryTrace) Clear() {
	t.entries = t.entries[:0]
}

// OpenTraceFile opens a trace file for writing
func OpenTraceFile(filename string) (*os.File, error) {
	return os.Create(filename)
}
