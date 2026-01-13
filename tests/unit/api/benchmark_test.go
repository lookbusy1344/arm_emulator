package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/lookbusy1344/arm-emulator/api"
)

// BenchmarkCreateSession benchmarks session creation performance
func BenchmarkCreateSession(b *testing.B) {
	server := testServerBench()
	defer shutdownServer(server)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/v1/session", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			b.Fatalf("Expected 201, got %d", w.Code)
		}
	}
}

// BenchmarkLoadProgram benchmarks program loading performance
func BenchmarkLoadProgram(b *testing.B) {
	server := testServerBench()
	defer shutdownServer(server)

	// Create session
	sessionID := createBenchSession(b, server)

	program := `.org 0x8000
main:
    MOV R0, #42
    SWI #0
`

	reqBody := api.LoadProgramRequest{Source: program}
	bodyBytes, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/session/%s/load", sessionID),
			bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
		}
	}
}

// BenchmarkStepExecution benchmarks single-step execution performance
func BenchmarkStepExecution(b *testing.B) {
	server := testServerBench()
	defer shutdownServer(server)

	sessionID := createBenchSession(b, server)

	program := `.org 0x8000
main:
    MOV R0, #42
    MOV R1, #1
    ADD R2, R0, R1
    SWI #0
`

	loadBenchProgram(b, server, sessionID, program)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/session/%s/step", sessionID), nil)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Reset for next iteration
		if i < b.N-1 {
			resetReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/session/%s/reset", sessionID), nil)
			resetW := httptest.NewRecorder()
			server.Handler().ServeHTTP(resetW, resetReq)
		}
	}
}

// BenchmarkGetRegisters benchmarks register state retrieval performance
func BenchmarkGetRegisters(b *testing.B) {
	server := testServerBench()
	defer shutdownServer(server)

	sessionID := createBenchSession(b, server)

	program := `.org 0x8000
main:
    MOV R0, #42
    SWI #0
`

	loadBenchProgram(b, server, sessionID, program)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", w.Code)
		}
	}
}

// BenchmarkGetMemory benchmarks memory read performance
func BenchmarkGetMemory(b *testing.B) {
	server := testServerBench()
	defer shutdownServer(server)

	sessionID := createBenchSession(b, server)

	program := `.org 0x8000
main:
    MOV R0, #42
    SWI #0
`

	loadBenchProgram(b, server, sessionID, program)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET",
			fmt.Sprintf("/api/v1/session/%s/memory?address=32768&length=256", sessionID), nil)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", w.Code)
		}
	}
}

// BenchmarkBreakpointOperations benchmarks breakpoint add/remove performance
func BenchmarkBreakpointOperations(b *testing.B) {
	server := testServerBench()
	defer shutdownServer(server)

	sessionID := createBenchSession(b, server)

	program := `.org 0x8000
main:
    MOV R0, #42
    SWI #0
`

	loadBenchProgram(b, server, sessionID, program)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Add breakpoint
		reqBody := api.BreakpointRequest{Address: 0x8000}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/session/%s/breakpoint", sessionID),
			bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", w.Code)
		}

		// Remove breakpoint
		deleteReq := httptest.NewRequest("DELETE",
			fmt.Sprintf("/api/v1/session/%s/breakpoint?address=32768", sessionID), nil)
		deleteW := httptest.NewRecorder()
		server.Handler().ServeHTTP(deleteW, deleteReq)

		if deleteW.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", deleteW.Code)
		}
	}
}

// BenchmarkConcurrentSessions benchmarks concurrent session handling
func BenchmarkConcurrentSessions(b *testing.B) {
	server := testServerBench()
	defer shutdownServer(server)

	program := `.org 0x8000
main:
    MOV R0, #42
    SWI #0
`

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Create session
			req := httptest.NewRequest("POST", "/api/v1/session", bytes.NewReader([]byte("{}")))
			w := httptest.NewRecorder()
			server.Handler().ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				b.Fatalf("Expected 201, got %d", w.Code)
				return
			}

			var resp api.SessionCreateResponse
			json.Unmarshal(w.Body.Bytes(), &resp)

			// Load program
			reqBody := api.LoadProgramRequest{Source: program}
			bodyBytes, _ := json.Marshal(reqBody)
			loadReq := httptest.NewRequest("POST",
				fmt.Sprintf("/api/v1/session/%s/load", resp.SessionID),
				bytes.NewBuffer(bodyBytes))
			loadReq.Header.Set("Content-Type", "application/json")
			loadW := httptest.NewRecorder()
			server.Handler().ServeHTTP(loadW, loadReq)

			// Step
			stepReq := httptest.NewRequest("POST",
				fmt.Sprintf("/api/v1/session/%s/step", resp.SessionID), nil)
			stepW := httptest.NewRecorder()
			server.Handler().ServeHTTP(stepW, stepReq)

			// Clean up
			deleteReq := httptest.NewRequest("DELETE",
				fmt.Sprintf("/api/v1/session/%s", resp.SessionID), nil)
			deleteW := httptest.NewRecorder()
			server.Handler().ServeHTTP(deleteW, deleteReq)
		}
	})
}

// BenchmarkJSONSerialization benchmarks JSON encoding/decoding performance
func BenchmarkJSONSerialization(b *testing.B) {
	server := testServerBench()
	defer shutdownServer(server)

	sessionID := createBenchSession(b, server)

	program := `.org 0x8000
main:
    MOV R0, #42
    MOV R1, #1
    MOV R2, #2
    MOV R3, #3
    SWI #0
`

	loadBenchProgram(b, server, sessionID, program)

	// Step once to populate registers
	stepReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/session/%s/step", sessionID), nil)
	stepW := httptest.NewRecorder()
	server.Handler().ServeHTTP(stepW, stepReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", w.Code)
		}

		// Decode response
		var resp api.RegistersResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
	}
}

// TestConcurrentSessionsStressTest tests many concurrent sessions
func TestConcurrentSessionsStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	server := testServerBench()
	defer shutdownServer(server)

	const numSessions = 20
	const numOperations = 4 // Program has 5 instructions, so 4 steps before SWI exit

	program := `.org 0x8000
main:
    MOV R0, #42
    ADD R1, R0, #1
    SUB R2, R1, #1
    MOV R3, R0
    SWI #0
`

	var wg sync.WaitGroup
	errors := make(chan error, numSessions*numOperations)

	start := time.Now()

	for i := 0; i < numSessions; i++ {
		wg.Add(1)
		go func(sessionNum int) {
			defer wg.Done()

			// Create session
			req := httptest.NewRequest("POST", "/api/v1/session", bytes.NewReader([]byte("{}")))
			w := httptest.NewRecorder()
			server.Handler().ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				errors <- fmt.Errorf("session %d: create failed with %d", sessionNum, w.Code)
				return
			}

			var resp api.SessionCreateResponse
			json.Unmarshal(w.Body.Bytes(), &resp)
			sessionID := resp.SessionID

			// Load program
			reqBody := api.LoadProgramRequest{Source: program}
			bodyBytes, _ := json.Marshal(reqBody)
			loadReq := httptest.NewRequest("POST",
				fmt.Sprintf("/api/v1/session/%s/load", sessionID),
				bytes.NewBuffer(bodyBytes))
			loadReq.Header.Set("Content-Type", "application/json")
			loadW := httptest.NewRecorder()
			server.Handler().ServeHTTP(loadW, loadReq)

			if loadW.Code != http.StatusOK {
				errors <- fmt.Errorf("session %d: load failed with %d", sessionNum, loadW.Code)
				return
			}

			// Wait for program to load
			time.Sleep(10 * time.Millisecond)

			// Perform multiple operations
			for j := 0; j < numOperations; j++ {
				// Step
				stepReq := httptest.NewRequest("POST",
					fmt.Sprintf("/api/v1/session/%s/step", sessionID), nil)
				stepW := httptest.NewRecorder()
				server.Handler().ServeHTTP(stepW, stepReq)

				if stepW.Code != http.StatusOK {
					errors <- fmt.Errorf("session %d, op %d: step failed with %d",
						sessionNum, j, stepW.Code)
					continue
				}

				// Get registers
				regReq := httptest.NewRequest("GET",
					fmt.Sprintf("/api/v1/session/%s/registers", sessionID), nil)
				regW := httptest.NewRecorder()
				server.Handler().ServeHTTP(regW, regReq)

				if regW.Code != http.StatusOK {
					errors <- fmt.Errorf("session %d, op %d: registers failed with %d",
						sessionNum, j, regW.Code)
				}
			}

			// Clean up
			deleteReq := httptest.NewRequest("DELETE",
				fmt.Sprintf("/api/v1/session/%s", sessionID), nil)
			deleteW := httptest.NewRecorder()
			server.Handler().ServeHTTP(deleteW, deleteReq)
		}(i)
	}

	wg.Wait()
	close(errors)

	elapsed := time.Since(start)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("Stress test failed with %d errors", errorCount)
	}

	t.Logf("Stress test completed: %d sessions, %d ops/session, %v elapsed",
		numSessions, numOperations, elapsed)
	t.Logf("Average time per session: %v", elapsed/numSessions)
	t.Logf("Average time per operation: %v", elapsed/(numSessions*numOperations))
}

// TestNetworkFailureScenarios tests error handling and recovery
func TestNetworkFailureScenarios(t *testing.T) {
	server := testServerBench()
	defer shutdownServer(server)

	t.Run("invalid session ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/session/invalid-id/registers", nil)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		sessionID := createBenchSession(t, server)

		req := httptest.NewRequest("POST",
			fmt.Sprintf("/api/v1/session/%s/load", sessionID),
			bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("empty source program", func(t *testing.T) {
		sessionID := createBenchSession(t, server)

		req := httptest.NewRequest("POST",
			fmt.Sprintf("/api/v1/session/%s/load", sessionID),
			bytes.NewBufferString(`{"source":""}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		// Empty source is accepted (results in empty program)
		if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
			t.Errorf("Expected 200 or 400, got %d", w.Code)
		}
	})

	t.Run("oversized request", func(t *testing.T) {
		sessionID := createBenchSession(t, server)

		// Create a large payload (> 1MB)
		largeSource := make([]byte, 2*1024*1024)
		for i := range largeSource {
			largeSource[i] = 'A'
		}

		reqBody := api.LoadProgramRequest{Source: string(largeSource)}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST",
			fmt.Sprintf("/api/v1/session/%s/load", sessionID),
			bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		// MaxBytesReader returns an error that results in 400, not 413
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid memory range", func(t *testing.T) {
		sessionID := createBenchSession(t, server)

		program := `.org 0x8000
main:
    MOV R0, #42
    SWI #0
`

		loadBenchProgram(t, server, sessionID, program)

		// Request more than 1MB
		req := httptest.NewRequest("GET",
			fmt.Sprintf("/api/v1/session/%s/memory?address=0&length=2000000", sessionID), nil)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})
}

// Helper functions

func testServerBench() *api.Server {
	return api.NewServer(8080)
}

func shutdownServer(server *api.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func createBenchSession(tb testing.TB, server *api.Server) string {
	req := httptest.NewRequest("POST", "/api/v1/session", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		tb.Fatalf("Failed to create session: %d %s", w.Code, w.Body.String())
	}

	var response api.SessionCreateResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		tb.Fatalf("Failed to decode session response: %v", err)
	}

	return response.SessionID
}

func loadBenchProgram(tb testing.TB, server *api.Server, sessionID, program string) {
	reqBody := api.LoadProgramRequest{Source: program}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST",
		fmt.Sprintf("/api/v1/session/%s/load", sessionID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		tb.Fatalf("Failed to load program: %d %s", w.Code, w.Body.String())
	}

	// Wait for program to load
	time.Sleep(10 * time.Millisecond)
}
