package api

import (
	"fmt"
	"net/http"
	"strconv"

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
	regs := session.Service.GetRegisters()
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

	// Load program using service
	loadErr := session.Service.LoadProgram(req.Source, "api")
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

	// Run the program asynchronously
	go func() {
		_ = session.Service.Run()
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

	session.Service.Stop()

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

	// Return updated registers
	regs := session.Service.GetRegisters()
	response := ToRegisterResponse(regs)

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

	session.Service.Reset()

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

	regs := session.Service.GetRegisters()
	response := ToRegisterResponse(regs)

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
	data := session.Service.ReadMemory(uint32(address), uint32(length))

	response := MemoryResponse{
		Address: uint32(address),
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
	lines := session.Service.Disassemble(uint32(address), uint32(count))

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

		session.Service.AddBreakpoint(req.Address)

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

		session.Service.RemoveBreakpoint(req.Address)

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

	response := BreakpointsResponse{
		Breakpoints: breakpoints,
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

	stdinErr := session.Service.SendStdin(req.Data)
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
