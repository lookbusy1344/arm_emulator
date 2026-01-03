# Backend Lifecycle Management Design
**Date:** 2026-01-03
**Status:** Approved
**Target:** Swift GUI backend process management

## Problem Statement

The Swift macOS GUI launches a Go backend process (`arm-emulator --api-server`) but can leave it orphaned when:
- Swift app crashes
- App is force-quit (Cmd+Q, Activity Monitor)
- Xcode stops debugging session
- System shutdown interrupts normal cleanup

**Current Workaround:** `killOrphanedBackends()` uses `pkill` to kill existing backends on startup. This is a symptom of the real problem: backends don't detect when their parent dies.

## Solution: Parent Process Monitoring

The Go backend will monitor its parent process ID (PPID) and automatically shut down when the parent dies.

### Architecture Overview

**How It Works:**
- When launched, backend records its parent PID via `os.Getppid()`
- Dedicated goroutine checks PPID every 2 seconds
- When PPID changes (parent died → OS re-parents to init/PID 1), trigger graceful shutdown
- Independent of signal handlers, catches all failure modes

**Key Benefits:**
- Zero network overhead (no HTTP heartbeat needed)
- Fast detection (~2-3 second maximum delay)
- Crash-proof (detects all failure modes)
- Development-friendly (works with Xcode debugging)
- Eliminates need for `killOrphanedBackends()` workaround

## Implementation Details

### 1. Go Backend: Process Monitor

**New file:** `api/process_monitor.go`

```go
package api

import (
    "log"
    "os"
    "time"
)

// ProcessMonitor watches the parent process and triggers shutdown when it dies
type ProcessMonitor struct {
    parentPID     int
    checkInterval time.Duration
    shutdownFunc  func()
    stopChan      chan struct{}
}

// NewProcessMonitor creates a monitor that calls shutdownFunc when parent dies
func NewProcessMonitor(shutdownFunc func()) *ProcessMonitor {
    return &ProcessMonitor{
        parentPID:     os.Getppid(),
        checkInterval: 2 * time.Second,
        shutdownFunc:  shutdownFunc,
        stopChan:      make(chan struct{}),
    }
}

// Start begins monitoring the parent process
func (pm *ProcessMonitor) Start() {
    go pm.monitorLoop()
}

// Stop gracefully stops the monitor
func (pm *ProcessMonitor) Stop() {
    close(pm.stopChan)
}

// monitorLoop runs in a goroutine and checks parent PID periodically
func (pm *ProcessMonitor) monitorLoop() {
    ticker := time.NewTicker(pm.checkInterval)
    defer ticker.Stop()

    log.Printf("Process monitor started (parent PID: %d)", pm.parentPID)

    for {
        select {
        case <-ticker.C:
            currentPPID := os.Getppid()
            if currentPPID != pm.parentPID {
                log.Printf("Parent process died (PPID changed: %d -> %d), initiating shutdown",
                    pm.parentPID, currentPPID)
                pm.shutdownFunc()
                return
            }
        case <-pm.stopChan:
            log.Println("Process monitor stopped")
            return
        }
    }
}
```

**Integration into `main.go`:**

```go
// In main.go, API server mode section (around line 98-128)

if *apiServer {
    server := api.NewServer(*apiPort)

    // Setup graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Create shutdown function (ensure it runs only once)
    var shutdownOnce sync.Once
    performShutdown := func() {
        shutdownOnce.Do(func() {
            fmt.Println("\nShutting down API server...")

            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancel()

            if err := server.Shutdown(ctx); err != nil {
                fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
                os.Exit(1)
            }

            fmt.Println("API server stopped")
            os.Exit(0)
        })
    }

    // Start process monitor to detect parent death
    monitor := api.NewProcessMonitor(performShutdown)
    monitor.Start()

    // Start server in goroutine
    go func() {
        if err := server.Start(); err != nil && err != http.ErrServerClosed {
            fmt.Fprintf(os.Stderr, "API server error: %v\n", err)
            os.Exit(1)
        }
    }()

    // Wait for shutdown signal
    <-sigChan
    performShutdown()
}
```

### 2. Swift Frontend: Simplified BackendManager

**Changes to `BackendManager.swift`:**

1. **Remove `killOrphanedBackends()`** (lines 81-94) - No longer needed!
2. **Remove call to `killOrphanedBackends()` from `ensureBackendRunning()`** (line 66)
3. **Simplify `shutdown()`** - Backend handles crashes automatically

```swift
func ensureBackendRunning() async {
    // REMOVED: await killOrphanedBackends()

    if await checkBackendHealth() {
        backendStatus = .running
        didStartBackend = false
        return
    }

    do {
        try await startBackend()
    } catch {
        backendStatus = .error(error.localizedDescription)
    }
}

// shutdown() stays mostly the same, but we know backend will
// self-terminate on crash, so this only handles normal quit
```

**Key Simplification:**
- No more `pkill` workaround
- Backend self-manages its lifecycle
- Swift only handles normal shutdown path
- Cleaner separation of concerns

## Scenario Analysis

| Scenario | Behavior | Result |
|----------|----------|--------|
| **Normal Quit** | Swift sends SIGTERM → backend graceful shutdown | Clean exit ~100-500ms |
| **App Crash** | Monitor detects PPID change → auto shutdown | Self-terminate in ~2-3s |
| **Force Quit** | Monitor detects parent death → auto shutdown | Self-terminate in ~2-3s |
| **Xcode Stop** | Xcode kills Swift app → monitor triggers | Self-terminate in ~2-3s |
| **System Shutdown** | OS sends SIGTERM to all processes | Clean shutdown via signal or monitor |
| **Manual Backend** | Parent is terminal/shell, not Swift app | Runs until terminal closes |

## Testing Strategy

### Unit Tests

**File:** `api/process_monitor_test.go`

```go
func TestProcessMonitor_DetectsParentDeath(t *testing.T) {
    shutdownCalled := false
    shutdown := func() { shutdownCalled = true }

    monitor := NewProcessMonitor(shutdown)
    // Test with mocked getppid that simulates parent death
    // Verify shutdown callback is invoked
}

func TestProcessMonitor_GracefulStop(t *testing.T) {
    monitor := NewProcessMonitor(func() {})
    monitor.Start()
    monitor.Stop()
    // Verify monitor goroutine exits cleanly
}
```

### Manual Testing

1. **Normal Quit Test:**
   - Start Swift app from Xcode
   - Monitor: `ps aux | grep "arm-emulator.*api-server"`
   - Quit app normally
   - Verify backend exits within 1 second

2. **Force Quit Test:**
   - Start Swift app
   - Note backend PID
   - Force quit Swift (Activity Monitor or `kill -9 <swift-pid>`)
   - Verify backend exits within 2-3 seconds

3. **Crash Test:**
   - Add crash button to Swift app: `fatalError("Test crash")`
   - Verify backend auto-terminates

4. **Development Workflow:**
   - Run from Xcode
   - Stop/restart rapidly 10+ times
   - Verify: `ps aux | grep arm-emulator` shows no orphans

### Success Criteria

- ✅ Zero orphaned processes after any termination scenario
- ✅ Backend shutdown within 3 seconds of parent death
- ✅ Graceful shutdown: sessions closed, logs flushed
- ✅ Can remove `killOrphanedBackends()` workaround
- ✅ Works seamlessly during Xcode development
- ✅ All existing tests continue to pass

## Implementation Checklist

### Go Backend
- [ ] Create `api/process_monitor.go`
- [ ] Add unit tests in `api/process_monitor_test.go`
- [ ] Integrate into `main.go` (API server mode)
- [ ] Add `sync.Once` for shutdown deduplication
- [ ] Test manually: force quit, crash, normal exit

### Swift Frontend
- [ ] Remove `killOrphanedBackends()` function
- [ ] Remove call from `ensureBackendRunning()`
- [ ] Update comments explaining new behavior
- [ ] Test all shutdown scenarios

### Documentation
- [ ] Update `SWIFT_GUI_PLANNING.md` with lifecycle changes
- [ ] Update `docs/SWIFT_APP.md` backend management section
- [ ] Add process monitoring notes to API.md

### Verification
- [ ] Run all Go tests: `go test ./...`
- [ ] Run all Swift tests
- [ ] Manual testing: 5 scenarios above
- [ ] Verify zero linting issues: `golangci-lint run ./...`
- [ ] Verify Swift formatting: `swiftformat --lint .`

## Alternative Approaches Considered

### HTTP Heartbeat (user's original suggestion)
- Swift sends periodic `/heartbeat` requests every 5s
- Backend shuts down after 10s timeout
- **Rejected:** Higher overhead, more complex, unnecessary for macOS-only deployment

### WebSocket Connection Monitoring
- Monitor WebSocket connection state
- Shutdown when connection drops
- **Rejected:** Too aggressive, breaks during normal reconnection

### Session Activity Timeout
- Track last activity on any session
- Shutdown after 30 minutes idle
- **Rejected:** Too slow for crash detection, doesn't solve the problem

## Risk Analysis

| Risk | Impact | Mitigation |
|------|--------|------------|
| `os.Getppid()` not available on platform | High | macOS/Linux/BSD all support this; Windows would need alternative |
| Backend shuts down during normal operation | High | Only triggers on PPID change (parent death), not normal conditions |
| 2-second delay allows orphan to accept connections | Low | Orphaned backend can't harm anything; sessions isolated by ID |
| Race condition: signal + monitor both trigger | Medium | Use `sync.Once` to ensure shutdown runs exactly once |

## Future Enhancements

**Optional HTTP Heartbeat (Stage 7):**
- If cross-platform clients (.NET, web) are added later
- Implement heartbeat as optional fallback mechanism
- Keep process monitoring as primary for macOS

**Configurable Check Interval:**
- Allow `--parent-check-interval` flag for development
- Faster checks (500ms) for rapid development cycles
- Slower checks (5s) for production if desired

## Conclusion

Parent process monitoring provides a robust, zero-overhead solution for backend lifecycle management on macOS. It eliminates orphaned processes without requiring coordination between frontend and backend, simplifies the Swift app code, and works reliably across all failure scenarios including crashes, force-quits, and development workflows.

**Recommendation:** Proceed with implementation.
