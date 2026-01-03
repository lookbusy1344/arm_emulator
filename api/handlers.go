package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/service"
)

// handleCreateSession handles POST /api/v1/session
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req SessionCreateRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session, err := s.sessions.CreateSession(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %v", err))
		return
	}

	response := SessionCreateResponse{
		SessionID: session.ID,
		CreatedAt: session.CreatedAt,
	}

	writeJSON(w, http.StatusCreated, response)
}

// handleListSessions handles GET /api/v1/session
func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	ids := s.sessions.ListSessions()

	response := map[string]interface{}{
		"sessions": ids,
		"count":    len(ids),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleGetSessionStatus handles GET /api/v1/session/{id}
func (s *Server) handleGetSessionStatus(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Get current state from service
	regs := session.Service.GetRegisterState()
	state := session.Service.GetExecutionState()
	memWrite := session.Service.GetLastMemoryWrite()

	response := SessionStatusResponse{
		SessionID: sessionID,
		State:     string(state),
		PC:        regs.PC,
		Cycles:    regs.Cycles,
		HasWrite:  memWrite.HasWrite,
		WriteAddr: memWrite.Address,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleDestroySession handles DELETE /api/v1/session/{id}
func (s *Server) handleDestroySession(w http.ResponseWriter, r *http.Request, sessionID string) {
	err := s.sessions.DestroySession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Session destroyed",
	})
}

// handleLoadProgram handles POST /api/v1/session/{id}/load
func (s *Server) handleLoadProgram(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	var req LoadProgramRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse assembly source
	p := parser.NewParser(req.Source, "api")
	program, parseErr := p.Parse()
	if parseErr != nil {
		// Collect all parse errors
		errorList := p.Errors()
		errors := make([]string, len(errorList.Errors))
		for i, e := range errorList.Errors {
			errors[i] = e.Error()
		}
		response := LoadProgramResponse{
			Success: false,
			Errors:  errors,
		}
		writeJSON(w, http.StatusBadRequest, response)
		return
	}

	// Determine entry point (same logic as main.go)
	var entryAddr uint32
	if startSym, exists := program.SymbolTable.Lookup("_start"); exists {
		entryAddr = startSym.Value
	} else if program.OriginSet {
		entryAddr = program.Origin
	} else {
		entryAddr = 0x8000 // Default ARM entry point
	}

	// Load program using service
	loadErr := session.Service.LoadProgram(program, entryAddr)
	if loadErr != nil {
		response := LoadProgramResponse{
			Success: false,
			Errors:  []string{loadErr.Error()},
		}
		writeJSON(w, http.StatusBadRequest, response)
		return
	}

	// Get symbols
	symbols := session.Service.GetSymbols()

	response := LoadProgramResponse{
		Success: true,
		Symbols: symbols,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleRun handles POST /api/v1/session/{id}/run
func (s *Server) handleRun(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Set running state synchronously BEFORE launching goroutine
	// This ensures the frontend can immediately observe the state change
	// and RunUntilHalt() will proceed with execution
	session.Service.SetRunning(true)

	// Run the program asynchronously
	go func() {
		_ = session.Service.RunUntilHalt()
	}()

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Program started",
	})
}

// handleStop handles POST /api/v1/session/{id}/stop
func (s *Server) handleStop(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	session.Service.Pause()

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Program stopped",
	})
}

// handleStep handles POST /api/v1/session/{id}/step
func (s *Server) handleStep(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	stepErr := session.Service.Step()
	if stepErr != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Step failed: %v", stepErr))
		return
	}

	// Get updated state
	regs := session.Service.GetRegisterState()
	state := session.Service.GetExecutionState()

	// Broadcast state change to WebSocket clients
	s.broadcastStateChange(sessionID, &regs, state)

	// Return updated registers
	response := ToRegisterResponse(&regs)
	writeJSON(w, http.StatusOK, response)
}

// handleReset handles POST /api/v1/session/{id}/reset
func (s *Server) handleReset(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	if err := session.Service.Reset(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Reset failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "VM reset",
	})
}

// handleGetRegisters handles GET /api/v1/session/{id}/registers
func (s *Server) handleGetRegisters(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	regs := session.Service.GetRegisterState()
	response := ToRegisterResponse(&regs)

	writeJSON(w, http.StatusOK, response)
}

// handleGetMemory handles GET /api/v1/session/{id}/memory
func (s *Server) handleGetMemory(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	address, err := parseHexOrDec(query.Get("address"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid address parameter")
		return
	}

	length, err := strconv.ParseUint(query.Get("length"), 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid length parameter")
		return
	}

	// Limit memory reads
	const maxMemoryRead = 1024 * 1024 // 1MB
	if length > maxMemoryRead {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Length too large (max %d bytes)", maxMemoryRead))
		return
	}

	// Read memory
	data, err := session.Service.GetMemory(uint32(address), uint32(length)) // #nosec G115 -- parseHexOrDec validates input fits in uint32
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read memory: %v", err))
		return
	}

	response := MemoryResponse{
		Address: uint32(address), // #nosec G115 -- parseHexOrDec validates input fits in uint32
		Data:    data,
		Length:  uint32(length),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleGetDisassembly handles GET /api/v1/session/{id}/disassembly
func (s *Server) handleGetDisassembly(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	address, err := parseHexOrDec(query.Get("address"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid address parameter")
		return
	}

	count, err := strconv.ParseUint(query.Get("count"), 10, 32)
	if err != nil || count == 0 {
		count = 10 // Default to 10 instructions
	}

	// Limit disassembly
	const maxDisassembly = 1000
	if count > maxDisassembly {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Count too large (max %d)", maxDisassembly))
		return
	}

	// Get disassembly
	lines := session.Service.GetDisassembly(uint32(address), int(count)) // #nosec G115 -- parseHexOrDec validates input fits in uint32

	instructions := make([]InstructionInfo, len(lines))
	for i, line := range lines {
		instructions[i] = ToInstructionInfo(&line)
	}

	response := DisassemblyResponse{
		Instructions: instructions,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleBreakpoint handles POST/DELETE /api/v1/session/{id}/breakpoint
func (s *Server) handleBreakpoint(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	switch r.Method {
	case http.MethodPost:
		// Add breakpoint
		var req BreakpointRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := session.Service.AddBreakpoint(req.Address); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to add breakpoint: %v", err))
			return
		}

		writeJSON(w, http.StatusOK, SuccessResponse{
			Success: true,
			Message: "Breakpoint added",
		})

	case http.MethodDelete:
		// Remove breakpoint
		var req BreakpointRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if err := session.Service.RemoveBreakpoint(req.Address); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to remove breakpoint: %v", err))
			return
		}

		writeJSON(w, http.StatusOK, SuccessResponse{
			Success: true,
			Message: "Breakpoint removed",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListBreakpoints handles GET /api/v1/session/{id}/breakpoints
func (s *Server) handleListBreakpoints(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	breakpoints := session.Service.GetBreakpoints()

	// Extract just the addresses from BreakpointInfo array
	addresses := make([]uint32, len(breakpoints))
	for i, bp := range breakpoints {
		addresses[i] = bp.Address
	}

	response := BreakpointsResponse{
		Breakpoints: addresses,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleSendStdin handles POST /api/v1/session/{id}/stdin
func (s *Server) handleSendStdin(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	var req StdinRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	stdinErr := session.Service.SendInput(req.Data)
	if stdinErr != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to send stdin: %v", stdinErr))
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Stdin sent",
	})
}

// parseHexOrDec parses a string as either hexadecimal (0x prefix) or decimal
func parseHexOrDec(s string) (uint64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	if len(s) > 2 && s[:2] == "0x" {
		return strconv.ParseUint(s[2:], 16, 32)
	}

	return strconv.ParseUint(s, 10, 32)
}

// handleWatchpoint handles POST/DELETE /api/v1/session/{id}/watchpoint
func (s *Server) handleWatchpoint(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	switch r.Method {
	case http.MethodPost:
		// Add watchpoint
		var req WatchpointRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate watchpoint type
		watchType := req.Type
		if watchType == "" {
			watchType = "readwrite" // Default
		}
		if watchType != "read" && watchType != "write" && watchType != "readwrite" {
			writeError(w, http.StatusBadRequest, "Invalid watchpoint type (must be 'read', 'write', or 'readwrite')")
			return
		}

		if err := session.Service.AddWatchpoint(req.Address, watchType); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to add watchpoint: %v", err))
			return
		}

		// Get the watchpoints to find the newly created one
		watchpoints := session.Service.GetWatchpoints()
		var newWatchpoint *service.WatchpointInfo
		for i := range watchpoints {
			if watchpoints[i].Address == req.Address {
				newWatchpoint = &watchpoints[i]
				break
			}
		}

		if newWatchpoint == nil {
			writeError(w, http.StatusInternalServerError, "Failed to retrieve created watchpoint")
			return
		}

		response := WatchpointResponse{
			ID:      newWatchpoint.ID,
			Address: newWatchpoint.Address,
			Type:    newWatchpoint.Type,
		}

		writeJSON(w, http.StatusOK, response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleDeleteWatchpoint handles DELETE /api/v1/session/{id}/watchpoint/{watchpointID}
func (s *Server) handleDeleteWatchpoint(w http.ResponseWriter, r *http.Request, sessionID string, watchpointID int) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	if err := session.Service.RemoveWatchpoint(watchpointID); err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Failed to remove watchpoint: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Watchpoint removed",
	})
}

// handleListWatchpoints handles GET /api/v1/session/{id}/watchpoints
func (s *Server) handleListWatchpoints(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	watchpoints := session.Service.GetWatchpoints()

	response := WatchpointsResponse{
		Watchpoints: watchpoints,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleTraceControl handles POST /api/v1/session/{id}/trace/{enable|disable}
func (s *Server) handleTraceControl(w http.ResponseWriter, r *http.Request, sessionID string, action string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	switch action {
	case "enable":
		if err := session.Service.EnableExecutionTrace(); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to enable trace: %v", err))
			return
		}
		writeJSON(w, http.StatusOK, SuccessResponse{
			Success: true,
			Message: "Execution trace enabled",
		})
	case "disable":
		session.Service.DisableExecutionTrace()
		writeJSON(w, http.StatusOK, SuccessResponse{
			Success: true,
			Message: "Execution trace disabled",
		})
	default:
		writeError(w, http.StatusBadRequest, "Invalid action (must be 'enable' or 'disable')")
	}
}

// handleTraceData handles GET /api/v1/session/{id}/trace/data
func (s *Server) handleTraceData(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	entries, err := session.Service.GetExecutionTraceData()
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get trace data: %v", err))
		return
	}

	// Convert vm.TraceEntry to API TraceEntryInfo
	apiEntries := make([]TraceEntryInfo, len(entries))
	for i, entry := range entries {
		apiEntries[i] = TraceEntryInfo{
			Sequence:        entry.Sequence,
			Address:         entry.Address,
			Opcode:          entry.Opcode,
			Disassembly:     entry.Disassembly,
			RegisterChanges: entry.RegisterChanges,
			Flags: CPSRFlags{
				N: entry.Flags.N,
				Z: entry.Flags.Z,
				C: entry.Flags.C,
				V: entry.Flags.V,
			},
			DurationNs: entry.Duration.Nanoseconds(),
		}
	}

	response := TraceDataResponse{
		Entries: apiEntries,
		Count:   len(apiEntries),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleStatsControl handles POST /api/v1/session/{id}/stats/{enable|disable}
func (s *Server) handleStatsControl(w http.ResponseWriter, r *http.Request, sessionID string, action string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	switch action {
	case "enable":
		if err := session.Service.EnableStatistics(); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to enable statistics: %v", err))
			return
		}
		writeJSON(w, http.StatusOK, SuccessResponse{
			Success: true,
			Message: "Statistics collection enabled",
		})
	case "disable":
		session.Service.DisableStatistics()
		writeJSON(w, http.StatusOK, SuccessResponse{
			Success: true,
			Message: "Statistics collection disabled",
		})
	default:
		writeError(w, http.StatusBadRequest, "Invalid action (must be 'enable' or 'disable')")
	}
}

// handleStats handles GET /api/v1/session/{id}/stats
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request, sessionID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := s.sessions.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Session not found")
		return
	}

	stats, err := session.Service.GetStatistics()
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to get statistics: %v", err))
		return
	}

	response := StatisticsResponse{
		TotalInstructions:  stats.TotalInstructions,
		TotalCycles:        stats.TotalCycles,
		ExecutionTimeMs:    stats.ExecutionTime.Milliseconds(),
		InstructionsPerSec: stats.InstructionsPerSec,
		InstructionCounts:  stats.InstructionCounts,
		BranchCount:        stats.BranchCount,
		BranchTakenCount:   stats.BranchTakenCount,
		BranchMissedCount:  stats.BranchMissedCount,
		MemoryReads:        stats.MemoryReads,
		MemoryWrites:       stats.MemoryWrites,
		BytesRead:          stats.BytesRead,
		BytesWritten:       stats.BytesWritten,
	}

	writeJSON(w, http.StatusOK, response)
}

// handleGetConfig handles GET /api/v1/config
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, return default configuration
	// In a full implementation, this would load from a config file or store
	cfg := s.getDefaultConfig()
	writeJSON(w, http.StatusOK, cfg)
}

// handleUpdateConfig handles PUT /api/v1/config
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cfg ConfigResponse
	if err := readJSON(r, &cfg); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// For now, just acknowledge the update
	// In a full implementation, this would save to config file/store
	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Configuration updated",
	})
}

// handleListExamples handles GET /api/v1/examples
func (s *Server) handleListExamples(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read examples directory
	examplesDir := "examples"
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read examples directory: %v", err))
		return
	}

	// Build example list
	examples := make([]ExampleInfo, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only include .s files
		name := entry.Name()
		if !strings.HasSuffix(name, ".s") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		examples = append(examples, ExampleInfo{
			Name: name,
			Size: info.Size(),
		})
	}

	response := ExamplesResponse{
		Examples: examples,
		Count:    len(examples),
	}

	writeJSON(w, http.StatusOK, response)
}

// handleGetExample handles GET /api/v1/examples/{name}
func (s *Server) handleGetExample(w http.ResponseWriter, r *http.Request, exampleName string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Security: prevent path traversal
	if strings.Contains(exampleName, "..") || strings.Contains(exampleName, "/") {
		writeError(w, http.StatusBadRequest, "Invalid example name")
		return
	}

	// Read example file
	examplePath := filepath.Join("examples", exampleName)
	content, err := os.ReadFile(examplePath) // #nosec G304 -- path is validated above
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Example not found: %s", exampleName))
		return
	}

	// Get file info for size
	info, err := os.Stat(examplePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get file info")
		return
	}

	response := ExampleContentResponse{
		Name:    exampleName,
		Content: string(content),
		Size:    info.Size(),
	}

	writeJSON(w, http.StatusOK, response)
}

// getDefaultConfig returns default configuration as API response
func (s *Server) getDefaultConfig() ConfigResponse {
	// In a full implementation, this would load from config package
	return ConfigResponse{
		Execution: ExecutionConfig{
			MaxCycles:      1000000,
			StackSize:      65536,
			DefaultEntry:   "0x8000",
			EnableTrace:    false,
			EnableMemTrace: false,
			EnableStats:    false,
		},
		Debugger: DebuggerConfig{
			HistorySize:    1000,
			AutoSaveBreaks: true,
			ShowSource:     true,
			ShowRegisters:  true,
		},
		Display: DisplayConfig{
			ColorOutput:   true,
			BytesPerLine:  16,
			DisasmContext: 5,
			SourceContext: 5,
			NumberFormat:  "hex",
		},
		Trace: TraceConfig{
			OutputFile:    "trace.log",
			FilterRegs:    "",
			IncludeFlags:  true,
			IncludeTiming: true,
			MaxEntries:    100000,
		},
		Statistics: StatisticsConfig{
			OutputFile:     "stats.json",
			Format:         "json",
			CollectHotPath: true,
			TrackCalls:     true,
		},
	}
}

// broadcastStateChange broadcasts VM state changes to WebSocket clients
func (s *Server) broadcastStateChange(sessionID string, regs *service.RegisterState, state service.ExecutionState) {
	if s.broadcaster == nil {
		return
	}

	// Convert register state to map for broadcasting
	// ARM register mapping: R0-R12, R13=SP, R14=LR, R15=PC
	data := map[string]interface{}{
		"status": string(state),
		"pc":     regs.PC,
		"sp":     regs.Registers[13],
		"lr":     regs.Registers[14],
		"cycles": regs.Cycles,
		"registers": map[string]uint32{
			"r0":  regs.Registers[0],
			"r1":  regs.Registers[1],
			"r2":  regs.Registers[2],
			"r3":  regs.Registers[3],
			"r4":  regs.Registers[4],
			"r5":  regs.Registers[5],
			"r6":  regs.Registers[6],
			"r7":  regs.Registers[7],
			"r8":  regs.Registers[8],
			"r9":  regs.Registers[9],
			"r10": regs.Registers[10],
			"r11": regs.Registers[11],
			"r12": regs.Registers[12],
		},
		"flags": map[string]bool{
			"n": regs.CPSR.N,
			"z": regs.CPSR.Z,
			"c": regs.CPSR.C,
			"v": regs.CPSR.V,
		},
	}

	s.broadcaster.BroadcastState(sessionID, data)
}
