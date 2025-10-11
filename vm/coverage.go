package vm

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// CoverageEntry represents coverage information for an address
type CoverageEntry struct {
	Address        uint32 // Instruction address
	ExecutionCount uint64 // Number of times executed
	FirstExecution uint64 // Cycle number of first execution
	LastExecution  uint64 // Cycle number of last execution
}

// CodeCoverage tracks which instructions have been executed
type CodeCoverage struct {
	Enabled bool
	Writer  io.Writer

	// Coverage data
	executed  map[uint32]*CoverageEntry // address -> execution info
	codeStart uint32                    // Start of code segment
	codeEnd   uint32                    // End of code segment

	// Symbol information (optional)
	symbols         map[string]uint32 // label -> address
	addressToSymbol map[uint32]string // address -> label
}

// NewCodeCoverage creates a new code coverage tracker
func NewCodeCoverage(writer io.Writer) *CodeCoverage {
	return &CodeCoverage{
		Enabled:         true,
		Writer:          writer,
		executed:        make(map[uint32]*CoverageEntry),
		symbols:         make(map[string]uint32),
		addressToSymbol: make(map[uint32]string),
	}
}

// SetCodeRange sets the range of code addresses to track
func (c *CodeCoverage) SetCodeRange(start, end uint32) {
	c.codeStart = start
	c.codeEnd = end
}

// LoadSymbols loads symbol information for better reporting
func (c *CodeCoverage) LoadSymbols(symbols map[string]uint32) {
	c.symbols = symbols
	// Build reverse map
	for name, addr := range symbols {
		c.addressToSymbol[addr] = name
	}
}

// Start starts coverage tracking
func (c *CodeCoverage) Start() {
	c.executed = make(map[uint32]*CoverageEntry)
}

// RecordExecution records that an instruction was executed
func (c *CodeCoverage) RecordExecution(address uint32, cycle uint64) {
	if !c.Enabled {
		return
	}

	// Only track if address is in code range (if range is set)
	if c.codeStart != 0 || c.codeEnd != 0 {
		if address < c.codeStart || address >= c.codeEnd {
			return
		}
	}

	if entry, exists := c.executed[address]; exists {
		entry.ExecutionCount++
		entry.LastExecution = cycle
	} else {
		c.executed[address] = &CoverageEntry{
			Address:        address,
			ExecutionCount: 1,
			FirstExecution: cycle,
			LastExecution:  cycle,
		}
	}
}

// GetCoverage returns the coverage percentage
func (c *CodeCoverage) GetCoverage() float64 {
	if c.codeStart == 0 && c.codeEnd == 0 {
		return 0.0
	}

	totalInstructions := (c.codeEnd - c.codeStart) / 4
	if totalInstructions == 0 {
		return 0.0
	}

	// Safe conversion: map size is bounded by memory constraints
	executedCount := uint32(len(c.executed)) // #nosec G115 -- map size limited by available memory
	return float64(executedCount) / float64(totalInstructions) * 100.0
}

// GetExecutedAddresses returns all executed addresses sorted
func (c *CodeCoverage) GetExecutedAddresses() []uint32 {
	addresses := make([]uint32, 0, len(c.executed))
	for addr := range c.executed {
		addresses = append(addresses, addr)
	}
	sort.Slice(addresses, func(i, j int) bool {
		return addresses[i] < addresses[j]
	})
	return addresses
}

// GetUnexecutedAddresses returns addresses in code range that were not executed
func (c *CodeCoverage) GetUnexecutedAddresses() []uint32 {
	if c.codeStart == 0 && c.codeEnd == 0 {
		return nil
	}

	unexecuted := make([]uint32, 0)
	for addr := c.codeStart; addr < c.codeEnd; addr += 4 {
		if _, exists := c.executed[addr]; !exists {
			unexecuted = append(unexecuted, addr)
		}
	}
	return unexecuted
}

// GetEntry returns coverage entry for an address
func (c *CodeCoverage) GetEntry(address uint32) *CoverageEntry {
	return c.executed[address]
}

// Flush writes coverage report to the writer
func (c *CodeCoverage) Flush() error {
	if c.Writer == nil {
		return nil
	}

	// Write header
	header := "Code Coverage Report\n"
	header += "====================\n\n"

	if c.codeStart != 0 || c.codeEnd != 0 {
		totalInstructions := (c.codeEnd - c.codeStart) / 4
		executedCount := len(c.executed)
		coverage := c.GetCoverage()

		header += fmt.Sprintf("Code Range:           0x%08X - 0x%08X\n", c.codeStart, c.codeEnd)
		header += fmt.Sprintf("Total Instructions:   %d\n", totalInstructions)
		header += fmt.Sprintf("Executed:             %d\n", executedCount)
		// Safe conversion: executedCount is already bounded by map size
		header += fmt.Sprintf("Not Executed:         %d\n", totalInstructions-uint32(executedCount)) // #nosec G115 -- executedCount bounded by memory
		header += fmt.Sprintf("Coverage:             %.2f%%\n\n", coverage)
	} else {
		header += fmt.Sprintf("Total Executed:       %d unique addresses\n\n", len(c.executed))
	}

	if _, err := c.Writer.Write([]byte(header)); err != nil {
		return err
	}

	// Write executed addresses
	if _, err := c.Writer.Write([]byte("Executed Addresses:\n")); err != nil {
		return err
	}
	if _, err := c.Writer.Write([]byte("-------------------\n")); err != nil {
		return err
	}

	executedAddrs := c.GetExecutedAddresses()
	for _, addr := range executedAddrs {
		entry := c.executed[addr]
		line := fmt.Sprintf("0x%08X: executed %6d times (first: cycle %6d, last: cycle %6d)",
			addr, entry.ExecutionCount, entry.FirstExecution, entry.LastExecution)

		// Add symbol if available
		if symbol, exists := c.addressToSymbol[addr]; exists {
			line += fmt.Sprintf(" [%s]", symbol)
		}

		line += "\n"
		if _, err := c.Writer.Write([]byte(line)); err != nil {
			return err
		}
	}

	// Write unexecuted addresses if code range is set
	unexecuted := c.GetUnexecutedAddresses()
	if len(unexecuted) > 0 {
		if _, err := c.Writer.Write([]byte("\nNot Executed:\n")); err != nil {
			return err
		}
		if _, err := c.Writer.Write([]byte("-------------\n")); err != nil {
			return err
		}

		for _, addr := range unexecuted {
			line := fmt.Sprintf("0x%08X", addr)

			// Add symbol if available
			if symbol, exists := c.addressToSymbol[addr]; exists {
				line += fmt.Sprintf(" [%s]", symbol)
			}

			line += "\n"
			if _, err := c.Writer.Write([]byte(line)); err != nil {
				return err
			}
		}
	}

	return nil
}

// ExportJSON exports coverage data as JSON
func (c *CodeCoverage) ExportJSON(w io.Writer) error {
	data := map[string]interface{}{
		"code_start":           c.codeStart,
		"code_end":             c.codeEnd,
		"coverage_percent":     c.GetCoverage(),
		"executed_count":       len(c.executed),
		"unexecuted_count":     len(c.GetUnexecutedAddresses()),
		"executed_addresses":   c.executed,
		"unexecuted_addresses": c.GetUnexecutedAddresses(),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// String returns a formatted string representation
func (c *CodeCoverage) String() string {
	var sb strings.Builder

	sb.WriteString("Code Coverage Summary\n")
	sb.WriteString("=====================\n\n")

	if c.codeStart != 0 || c.codeEnd != 0 {
		totalInstructions := (c.codeEnd - c.codeStart) / 4
		executedCount := len(c.executed)
		coverage := c.GetCoverage()

		sb.WriteString(fmt.Sprintf("Code Range:         0x%08X - 0x%08X\n", c.codeStart, c.codeEnd))
		sb.WriteString(fmt.Sprintf("Total Instructions: %d\n", totalInstructions))
		sb.WriteString(fmt.Sprintf("Executed:           %d\n", executedCount))
		// Safe conversion: executedCount is already bounded by map size
		sb.WriteString(fmt.Sprintf("Not Executed:       %d\n", totalInstructions-uint32(executedCount))) // #nosec G115 -- executedCount bounded by memory
		sb.WriteString(fmt.Sprintf("Coverage:           %.2f%%\n", coverage))
	} else {
		sb.WriteString(fmt.Sprintf("Executed:           %d unique addresses\n", len(c.executed)))
	}

	return sb.String()
}
