package vm

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"sort"
	"strings"
	"time"
)

// InstructionStats tracks statistics for a single instruction type
type InstructionStats struct {
	Mnemonic string
	Count    uint64
	Cycles   uint64
}

// FunctionStats tracks statistics for a function
type FunctionStats struct {
	Name        string
	Address     uint32
	CallCount   uint64
	TotalCycles uint64
}

// HotPathEntry represents a frequently executed address
type HotPathEntry struct {
	Address uint32
	Count   uint64
}

// PerformanceStatistics tracks execution statistics
type PerformanceStatistics struct {
	Enabled bool

	// Execution metrics
	TotalInstructions  uint64
	TotalCycles        uint64
	ExecutionTime      time.Duration
	InstructionsPerSec float64

	// Instruction breakdown
	InstructionCounts map[string]uint64 // mnemonic -> count

	// Branch statistics
	BranchCount       uint64
	BranchTakenCount  uint64
	BranchMissedCount uint64

	// Function call tracking
	FunctionCalls map[uint32]*FunctionStats // address -> stats

	// Hot path (most frequently executed addresses)
	HotPath map[uint32]uint64 // address -> count

	// Memory access statistics
	MemoryReads  uint64
	MemoryWrites uint64
	BytesRead    uint64
	BytesWritten uint64

	// Internal
	startTime      time.Time
	collectHotPath bool
	trackCalls     bool
}

// NewPerformanceStatistics creates a new statistics tracker
func NewPerformanceStatistics() *PerformanceStatistics {
	return &PerformanceStatistics{
		Enabled:           true,
		InstructionCounts: make(map[string]uint64),
		FunctionCalls:     make(map[uint32]*FunctionStats),
		HotPath:           make(map[uint32]uint64),
		collectHotPath:    true,
		trackCalls:        true,
	}
}

// Start starts statistics collection
func (s *PerformanceStatistics) Start() {
	s.startTime = time.Now()
	s.TotalInstructions = 0
	s.TotalCycles = 0
	s.InstructionCounts = make(map[string]uint64)
	s.BranchCount = 0
	s.BranchTakenCount = 0
	s.BranchMissedCount = 0
	s.FunctionCalls = make(map[uint32]*FunctionStats)
	s.HotPath = make(map[uint32]uint64)
	s.MemoryReads = 0
	s.MemoryWrites = 0
	s.BytesRead = 0
	s.BytesWritten = 0
}

// RecordInstruction records an executed instruction
func (s *PerformanceStatistics) RecordInstruction(mnemonic string, address uint32, cycles uint64) {
	if !s.Enabled {
		return
	}

	s.TotalInstructions++
	s.TotalCycles += cycles
	s.InstructionCounts[mnemonic]++

	// Track hot path
	if s.collectHotPath {
		s.HotPath[address]++
	}
}

// RecordBranch records a branch instruction
func (s *PerformanceStatistics) RecordBranch(taken bool) {
	if !s.Enabled {
		return
	}

	s.BranchCount++
	if taken {
		s.BranchTakenCount++
	} else {
		s.BranchMissedCount++
	}
}

// RecordFunctionCall records a function call
func (s *PerformanceStatistics) RecordFunctionCall(address uint32, name string) {
	if !s.Enabled || !s.trackCalls {
		return
	}

	if stats, exists := s.FunctionCalls[address]; exists {
		stats.CallCount++
	} else {
		s.FunctionCalls[address] = &FunctionStats{
			Name:      name,
			Address:   address,
			CallCount: 1,
		}
	}
}

// RecordMemoryRead records a memory read
func (s *PerformanceStatistics) RecordMemoryRead(bytes uint64) {
	if !s.Enabled {
		return
	}

	s.MemoryReads++
	s.BytesRead += bytes
}

// RecordMemoryWrite records a memory write
func (s *PerformanceStatistics) RecordMemoryWrite(bytes uint64) {
	if !s.Enabled {
		return
	}

	s.MemoryWrites++
	s.BytesWritten += bytes
}

// Finalize finalizes statistics collection
func (s *PerformanceStatistics) Finalize() {
	s.ExecutionTime = time.Since(s.startTime)
	if s.ExecutionTime.Seconds() > 0 {
		s.InstructionsPerSec = float64(s.TotalInstructions) / s.ExecutionTime.Seconds()
	}
}

// GetTopInstructions returns the most frequently executed instructions
func (s *PerformanceStatistics) GetTopInstructions(n int) []InstructionStats {
	stats := make([]InstructionStats, 0, len(s.InstructionCounts))
	for mnemonic, count := range s.InstructionCounts {
		stats = append(stats, InstructionStats{
			Mnemonic: mnemonic,
			Count:    count,
		})
	}

	// Sort by count descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	if n > 0 && n < len(stats) {
		return stats[:n]
	}
	return stats
}

// GetTopHotPath returns the most frequently executed addresses
func (s *PerformanceStatistics) GetTopHotPath(n int) []HotPathEntry {
	entries := make([]HotPathEntry, 0, len(s.HotPath))
	for addr, count := range s.HotPath {
		entries = append(entries, HotPathEntry{
			Address: addr,
			Count:   count,
		})
	}

	// Sort by count descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})

	if n > 0 && n < len(entries) {
		return entries[:n]
	}
	return entries
}

// GetTopFunctions returns the most frequently called functions
func (s *PerformanceStatistics) GetTopFunctions(n int) []*FunctionStats {
	functions := make([]*FunctionStats, 0, len(s.FunctionCalls))
	for _, stats := range s.FunctionCalls {
		functions = append(functions, stats)
	}

	// Sort by call count descending
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].CallCount > functions[j].CallCount
	})

	if n > 0 && n < len(functions) {
		return functions[:n]
	}
	return functions
}

// ExportJSON exports statistics as JSON
func (s *PerformanceStatistics) ExportJSON(w io.Writer) error {
	s.Finalize()

	data := map[string]interface{}{
		"total_instructions":   s.TotalInstructions,
		"total_cycles":         s.TotalCycles,
		"execution_time_ms":    s.ExecutionTime.Milliseconds(),
		"instructions_per_sec": s.InstructionsPerSec,
		"branch_count":         s.BranchCount,
		"branch_taken":         s.BranchTakenCount,
		"branch_missed":        s.BranchMissedCount,
		"memory_reads":         s.MemoryReads,
		"memory_writes":        s.MemoryWrites,
		"bytes_read":           s.BytesRead,
		"bytes_written":        s.BytesWritten,
		"top_instructions":     s.GetTopInstructions(20),
		"hot_path":             s.GetTopHotPath(20),
		"top_functions":        s.GetTopFunctions(20),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// ExportCSV exports statistics as CSV
func (s *PerformanceStatistics) ExportCSV(w io.Writer) error {
	s.Finalize()

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"Metric", "Value"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write summary metrics
	rows := [][]string{
		{"Total Instructions", fmt.Sprintf("%d", s.TotalInstructions)},
		{"Total Cycles", fmt.Sprintf("%d", s.TotalCycles)},
		{"Execution Time (ms)", fmt.Sprintf("%d", s.ExecutionTime.Milliseconds())},
		{"Instructions/Sec", fmt.Sprintf("%.2f", s.InstructionsPerSec)},
		{"Branch Count", fmt.Sprintf("%d", s.BranchCount)},
		{"Branch Taken", fmt.Sprintf("%d", s.BranchTakenCount)},
		{"Branch Missed", fmt.Sprintf("%d", s.BranchMissedCount)},
		{"Memory Reads", fmt.Sprintf("%d", s.MemoryReads)},
		{"Memory Writes", fmt.Sprintf("%d", s.MemoryWrites)},
		{"Bytes Read", fmt.Sprintf("%d", s.BytesRead)},
		{"Bytes Written", fmt.Sprintf("%d", s.BytesWritten)},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write blank line
	writer.Write([]string{})

	// Write instruction breakdown
	writer.Write([]string{"Instruction", "Count"})
	for _, stat := range s.GetTopInstructions(0) {
		if err := writer.Write([]string{stat.Mnemonic, fmt.Sprintf("%d", stat.Count)}); err != nil {
			return err
		}
	}

	return nil
}

// ExportHTML exports statistics as HTML
func (s *PerformanceStatistics) ExportHTML(w io.Writer) error {
	s.Finalize()

	tmpl := template.Must(template.New("stats").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>ARM Emulator Performance Statistics</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        h2 { color: #666; margin-top: 30px; }
        table { border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
        tr:nth-child(even) { background-color: #f2f2f2; }
        .metric { font-weight: bold; }
    </style>
</head>
<body>
    <h1>ARM Emulator Performance Statistics</h1>

    <h2>Execution Summary</h2>
    <table>
        <tr><td class="metric">Total Instructions</td><td>{{.TotalInstructions}}</td></tr>
        <tr><td class="metric">Total Cycles</td><td>{{.TotalCycles}}</td></tr>
        <tr><td class="metric">Execution Time</td><td>{{.ExecutionTime}}</td></tr>
        <tr><td class="metric">Instructions/Second</td><td>{{printf "%.2f" .InstructionsPerSec}}</td></tr>
    </table>

    <h2>Branch Statistics</h2>
    <table>
        <tr><td class="metric">Total Branches</td><td>{{.BranchCount}}</td></tr>
        <tr><td class="metric">Branches Taken</td><td>{{.BranchTakenCount}}</td></tr>
        <tr><td class="metric">Branches Not Taken</td><td>{{.BranchMissedCount}}</td></tr>
        <tr><td class="metric">Branch Rate</td><td>{{printf "%.1f%%" .BranchRate}}</td></tr>
    </table>

    <h2>Memory Access Statistics</h2>
    <table>
        <tr><td class="metric">Memory Reads</td><td>{{.MemoryReads}}</td></tr>
        <tr><td class="metric">Memory Writes</td><td>{{.MemoryWrites}}</td></tr>
        <tr><td class="metric">Bytes Read</td><td>{{.BytesRead}}</td></tr>
        <tr><td class="metric">Bytes Written</td><td>{{.BytesWritten}}</td></tr>
    </table>

    <h2>Top Instructions (by frequency)</h2>
    <table>
        <tr><th>Instruction</th><th>Count</th><th>Percentage</th></tr>
        {{range .TopInstructions}}
        <tr><td>{{.Mnemonic}}</td><td>{{.Count}}</td><td>{{.Percentage}}%</td></tr>
        {{end}}
    </table>

    <h2>Hot Path (most executed addresses)</h2>
    <table>
        <tr><th>Address</th><th>Executions</th></tr>
        {{range .HotPath}}
        <tr><td>0x{{printf "%04X" .Address}}</td><td>{{.Count}}</td></tr>
        {{end}}
    </table>

    {{if .TopFunctions}}
    <h2>Function Call Statistics</h2>
    <table>
        <tr><th>Function</th><th>Address</th><th>Call Count</th></tr>
        {{range .TopFunctions}}
        <tr><td>{{.Name}}</td><td>0x{{printf "%04X" .Address}}</td><td>{{.CallCount}}</td></tr>
        {{end}}
    </table>
    {{end}}
</body>
</html>
`))

	// Prepare data for template
	data := struct {
		TotalInstructions  uint64
		TotalCycles        uint64
		ExecutionTime      time.Duration
		InstructionsPerSec float64
		BranchCount        uint64
		BranchTakenCount   uint64
		BranchMissedCount  uint64
		BranchRate         float64
		MemoryReads        uint64
		MemoryWrites       uint64
		BytesRead          uint64
		BytesWritten       uint64
		TopInstructions    []struct {
			Mnemonic   string
			Count      uint64
			Percentage float64
		}
		HotPath      []HotPathEntry
		TopFunctions []*FunctionStats
	}{
		TotalInstructions:  s.TotalInstructions,
		TotalCycles:        s.TotalCycles,
		ExecutionTime:      s.ExecutionTime,
		InstructionsPerSec: s.InstructionsPerSec,
		BranchCount:        s.BranchCount,
		BranchTakenCount:   s.BranchTakenCount,
		BranchMissedCount:  s.BranchMissedCount,
		MemoryReads:        s.MemoryReads,
		MemoryWrites:       s.MemoryWrites,
		BytesRead:          s.BytesRead,
		BytesWritten:       s.BytesWritten,
		HotPath:            s.GetTopHotPath(20),
		TopFunctions:       s.GetTopFunctions(20),
	}

	// Calculate branch rate
	if s.BranchCount > 0 {
		data.BranchRate = float64(s.BranchTakenCount) / float64(s.BranchCount) * 100
	}

	// Convert top instructions with percentages
	topInsts := s.GetTopInstructions(20)
	for _, inst := range topInsts {
		percentage := float64(inst.Count) / float64(s.TotalInstructions) * 100
		data.TopInstructions = append(data.TopInstructions, struct {
			Mnemonic   string
			Count      uint64
			Percentage float64
		}{
			Mnemonic:   inst.Mnemonic,
			Count:      inst.Count,
			Percentage: percentage,
		})
	}

	return tmpl.Execute(w, data)
}

// String returns a formatted string representation
func (s *PerformanceStatistics) String() string {
	s.Finalize()

	var sb strings.Builder

	sb.WriteString("Performance Statistics\n")
	sb.WriteString("======================\n\n")

	sb.WriteString(fmt.Sprintf("Total Instructions:  %d\n", s.TotalInstructions))
	sb.WriteString(fmt.Sprintf("Total Cycles:        %d\n", s.TotalCycles))
	sb.WriteString(fmt.Sprintf("Execution Time:      %v\n", s.ExecutionTime))
	sb.WriteString(fmt.Sprintf("Instructions/Sec:    %.2f\n\n", s.InstructionsPerSec))

	sb.WriteString(fmt.Sprintf("Branch Count:        %d\n", s.BranchCount))
	sb.WriteString(fmt.Sprintf("Branches Taken:      %d\n", s.BranchTakenCount))
	sb.WriteString(fmt.Sprintf("Branches Not Taken:  %d\n\n", s.BranchMissedCount))

	sb.WriteString(fmt.Sprintf("Memory Reads:        %d (%d bytes)\n", s.MemoryReads, s.BytesRead))
	sb.WriteString(fmt.Sprintf("Memory Writes:       %d (%d bytes)\n\n", s.MemoryWrites, s.BytesWritten))

	sb.WriteString("Top Instructions:\n")
	for i, stat := range s.GetTopInstructions(10) {
		percentage := float64(stat.Count) / float64(s.TotalInstructions) * 100
		sb.WriteString(fmt.Sprintf("  %2d. %-8s %8d (%.1f%%)\n", i+1, stat.Mnemonic, stat.Count, percentage))
	}

	return sb.String()
}
