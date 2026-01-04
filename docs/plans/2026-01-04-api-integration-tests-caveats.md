# API Integration Tests - Implementation Caveats

**Status:** In Progress (9/18 tasks complete)
**Date:** 2026-01-04

This document tracks architectural issues and technical debt discovered during implementation that must be addressed before the integration tests are fully functional.

---

## Task 7: WebSocket Client Connection - Concurrency Issues

**Status:** Spec compliant but has inherent concurrency issues from specification

### Issue 1: Race Condition in Close Method
- **Severity:** Important
- **Location:** `tests/integration/api_example_programs_test.go:78-90`
- **Problem:** The mutex only protects the `Close()` method, but `c.conn` is accessed concurrently by both `receiveLoop()` (line 67, `c.conn.ReadJSON`) and `Close()` (lines 84-86) without coordination
- **Impact:** Race detector will flag this; undefined behavior if Close() called during ReadJSON()
- **Root cause:** WebSocket connections are not safe for concurrent reads/writes from multiple goroutines
- **Proposed fixes:**
  - Use separate `closed` flag protected by mutex, checked at start of `receiveLoop`
  - Use `sync.Once` to ensure Close only executes once
  - Add connection read/write deadlines and rely on error handling
- **When to fix:** Before implementing actual WebSocket tests that will trigger concurrent access patterns

### Issue 2: Error Channel May Block During Shutdown
- **Severity:** Important
- **Location:** `tests/integration/api_example_programs_test.go:71`
- **Problem:** If the `errors` channel (capacity 10) fills up and `receiveLoop()` tries to send another error, it will block indefinitely instead of exiting
- **Impact:** During abnormal shutdown with multiple errors, goroutine may leak; `<-c.done` wait in Close will hang forever; tests could timeout
- **Proposed fix:**
  ```go
  select {
  case c.errors <- err:
  default:
      // Channel full, drop error
  }
  return
  ```
- **When to fix:** Before implementing tests that might generate multiple errors

### Issue 3: Updates Channel May Block During Shutdown
- **Severity:** Important
- **Location:** `tests/integration/api_example_programs_test.go:74`
- **Problem:** If `updates` channel fills up (capacity 10), sending will block and prevent clean shutdown
- **Impact:** If test code stops consuming from updates channel but connection keeps receiving messages, goroutine blocks forever; Close hangs waiting for receiveLoop
- **Proposed fix:**
  ```go
  select {
  case c.updates <- update:
  case <-time.After(5 * time.Second):
      return // Shutdown timeout
  }
  ```
- **When to fix:** Before implementing tests that might not consume all updates

### Note
These issues are **inherent in the Task 7 specification**, not implementation bugs. The code implements what was requested. However, they should be addressed before writing actual WebSocket tests that will exercise these code paths.

---

## Task 9: Real HTTP Server - Not Yet Functional

**Status:** Infrastructure scaffolding complete, but function cannot be used yet

### Issue 1: Port Exposure Limitation (BLOCKING)
- **Severity:** Important (blocks WebSocket test implementation)
- **Location:** `tests/integration/api_example_programs_test.go:162-164`
- **Problem:** Function creates server with `port: 0` (random port) but returns hardcoded `http://localhost:8080` URL
  - Server runs on random port (e.g., 54321)
  - Tests try to connect to port 8080
  - **Connection will fail**
- **Impact:** First test that uses `createTestServerWithWebSocket()` will fail mysteriously
- **TODO comment:** Already acknowledged in code: "TODO: need to expose port from server"
- **Proposed fixes:**
  1. **Add GetPort() to api.Server** - Extract port from listener after Start()
  2. **Change Server.Start() to return listener** - Update with actual port from `listener.Addr()`
  3. **Use fixed test port** - Keep current approach but add warning guards
- **Recommended fix:** Option 2 (modify api.Server.Start())
- **When to fix:** MUST be fixed before Task 11 (first actual WebSocket test)
- **Guard needed:** Add `t.Skip()` to function until fixed to prevent accidental usage

### Issue 2: Race Condition in Server Startup
- **Severity:** Important (will cause flaky tests)
- **Location:** `tests/integration/api_example_programs_test.go:159-160`
- **Problem:** Uses timing-based synchronization (`time.Sleep(50 * time.Millisecond)`)
  - No guarantee server is ready when function returns
  - On slow systems or under load, 50ms might not be enough
- **Impact:** Tests become flaky with intermittent failures on CI systems
- **Proposed fixes:**
  1. **Use channel to signal readiness** (requires modifying api.Server)
  2. **Poll health endpoint** (simpler, no api.Server changes needed):
     ```go
     for i := 0; i < 50; i++ {  // Try for ~5 seconds
         resp, err := http.Get(baseURL + "/health")
         if err == nil && resp.StatusCode == 200 {
             resp.Body.Close()
             return server, baseURL
         }
         time.Sleep(100 * time.Millisecond)
     }
     t.Fatal("Server failed to respond to health checks")
     ```
- **Recommended fix:** Health check polling (option 2)
- **When to fix:** Before writing WebSocket tests that depend on server being ready

### Issue 3: Missing Shutdown Timeout
- **Severity:** Important (could cause test hangs)
- **Location:** `tests/integration/api_example_programs_test.go:166-168`
- **Problem:** Passing `nil` context to `server.Shutdown()` means it will block indefinitely if connections don't close cleanly
- **Impact:** Tests could hang during cleanup, especially with active WebSocket connections; hard to debug
- **Proposed fix:**
  ```go
  t.Cleanup(func() {
      ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
      defer cancel()
      if err := server.Shutdown(ctx); err != nil {
          t.Logf("Server shutdown error: %v", err)
      }
  })
  ```
- **Required import:** Add `"context"` to imports
- **When to fix:** Before writing WebSocket tests that create persistent connections

### Usage Warning
The `createTestServerWithWebSocket()` function is **infrastructure scaffolding only**. It compiles and passes review but **cannot be used** until the port exposure issue is fixed. Consider adding this guard:

```go
func createTestServerWithWebSocket(t *testing.T) (*api.Server, string) {
    t.Skip("createTestServerWithWebSocket requires port exposure (not yet implemented)")
    // ... rest of function
}
```

Remove the skip when Issues 1-3 are addressed.

---

## Summary

**Total caveats:** 6 issues across 2 tasks
- **Blocking:** 1 (Task 9 port exposure - must fix before Task 11)
- **Important:** 5 (concurrency, race conditions, timeouts)
- **When to address:** Before implementing actual WebSocket tests (Tasks 11+)

**Recommendation:** Address Task 9 issues (port exposure, startup race, shutdown timeout) in Task 10 or Task 11 before writing first WebSocket test. Task 7 concurrency issues can be addressed when they cause actual test failures or race detector warnings.
