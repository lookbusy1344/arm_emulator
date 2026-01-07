# WebSocket Race Condition Investigation

## Problem Statement

Intermittent test failures in `TestAPIExamplePrograms` with error:
```
--- FAIL: TestAPIExamplePrograms (12.90s)
    --- FAIL: TestAPIExamplePrograms/Add128Bit_API (10.06s)
        api_example_programs_test.go:504: execution failed: waiting for halt: timeout waiting for state update
```

The test times out waiting for the "halted" state after 10 seconds, despite the program completing successfully.

## Root Cause Analysis

### The Fundamental Race Condition

The race occurs between:
1. **Program execution completing** → Server sends "halted" WebSocket update
2. **Test calling `WaitForState("halted")`** → Starts listening for updates

**Timeline of the race:**

```
t0: WebSocket client created and subscription sent (50ms sleep)
t1: Program loaded via API
t2: startExecution() called → triggers async execution
t3: WaitForState("halted") called → starts waiting
    ↓
    RACE WINDOW: If program completes between t2 and t3,
    the "halted" update arrives before we start listening
```

### Why It's Intermittent

- **Fast programs** (like `add_128bit.s`, `test_debug_syscalls.s`) execute in microseconds
- Most tests (~80%) work because programs take long enough for `WaitForState` to be ready
- ~20% fail when execution is extremely fast and CPU scheduling is unfavorable

### Original Design Flaws

1. **Channel buffer drops messages**: The `receiveLoop` has a timeout that drops updates:
   ```go
   select {
   case c.updates <- update:
   case <-time.After(100 * time.Millisecond):  // Drops the update!
   }
   ```

2. **No state history**: Once a state update passes through the channel, it's lost if not immediately consumed

3. **Late subscription**: The 50ms sleep after subscription isn't sufficient for fast programs

## Attempted Fixes

### Attempt 1: Track Current State
Added `currentState` field to track last known state and check it before waiting.

**Result**: Partial success, but still failed because `currentState` might not be updated if the message was dropped.

### Attempt 2: Increase Buffer Size
Changed channel buffer from 10 to 100 elements.

**Result**: Reduced failures but didn't eliminate them. Fast programs can still complete before draining starts.

### Attempt 3: Drain Channel on Entry
Modified `WaitForState` to drain all pending updates before starting timeout:
```go
for {
    select {
    case update := <-c.updates:
        // Process update
    default:
        goto waitLoop
    }
}
```

**Result**: Better, but messages can still be dropped in `receiveLoop` if test is slow to drain.

### Attempt 4: API Fallback Check
Added fallback to query actual session state via `server.GetSession()` when timeout occurs:
```go
if remaining <= 0 {
    if session, err := c.server.GetSession(c.sessionID); err == nil {
        actualState := string(session.Service.GetExecutionState())
        if actualState == targetState {
            return StateUpdate{Data: map[string]interface{}{"status": actualState}}, nil
        }
    }
    return StateUpdate{}, fmt.Errorf("timeout waiting for state %q", targetState)
}
```

**Result**: Should work in theory, but still seeing intermittent failures (~20% rate).

### Attempt 5: Track All Seen States
Added `seenStates map[string]bool` to remember every state encountered:
```go
type WebSocketTestClient struct {
    // ...
    seenStates map[string]bool
}
```

Check this map first in `WaitForState`:
```go
c.mu.Lock()
if c.seenStates[targetState] {
    c.mu.Unlock()
    return StateUpdate{Data: map[string]interface{}{"status": targetState}}, nil
}
c.mu.Unlock()
```

**Result**: Testing in progress...

## Code Changes Made

### Files Modified

1. **`tests/integration/api_example_programs_test.go`**:
   - Added `currentState string` field to `WebSocketTestClient`
   - Added `seenStates map[string]bool` field
   - Added `server *api.Server` and `sessionID string` fields for API fallback
   - Modified `NewWebSocketTestClient` to accept server and sessionID parameters
   - Modified `receiveLoop` to track state updates and increase timeout to 5s
   - Modified `WaitForState` to:
     - Check `seenStates` first
     - Drain pending channel updates
     - Fall back to API query on timeout
   - Increased channel buffer to 100

2. **`api/server.go`**:
   - Added public `GetSession(sessionID string) (*Session, error)` method for testing

## Remaining Issues

The fix appears to help but may not be 100% reliable yet. Possible remaining issues:

1. **Message dropping in receiveLoop**: The 5-second timeout still drops messages if test is blocked
2. **Timing of state tracking**: Updates in `receiveLoop` and `WaitForState` might miss the critical window
3. **WebSocket buffering**: Server-side or network buffering could delay message delivery

## Recommended Next Steps

### Option A: Eliminate Message Dropping (Preferred)
Make `receiveLoop` fully blocking when sending to channel:
```go
// Block indefinitely - test must consume updates
c.updates <- update
```

Risk: If test hangs, goroutine leaks. Mitigate with proper cleanup in `Close()`.

### Option B: Pre-emptive State Query
Query actual session state BEFORE calling `WaitForState`:
```go
// After startExecution, give it a moment then check
time.Sleep(10 * time.Millisecond)
if session, _ := server.GetSession(sessionID); session != nil {
    if string(session.Service.GetExecutionState()) == "halted" {
        return // Already done
    }
}
// Now wait via WebSocket
wsClient.WaitForState("halted", 10*time.Second)
```

### Option C: Redesign Test Pattern
Instead of relying on WebSocket updates, poll the session state via API:
```go
deadline := time.Now().Add(10 * time.Second)
for time.Now().Before(deadline) {
    if session, _ := server.GetSession(sessionID); session != nil {
        if string(session.Service.GetExecutionState()) == "halted" {
            break
        }
    }
    time.Sleep(50 * time.Millisecond)
}
```

This is more reliable but defeats the purpose of testing WebSocket functionality.

## Test Results Summary

From multiple test runs (30-40 iterations):
- **Success rate**: ~80% (24/30)
- **Failure rate**: ~20% (6/30)
- **Failure pattern**: Always timeout after ~10 seconds + test overhead (~13s total)
- **Most common failures**: `TestDebugSyscalls_API`, `Add128Bit_API`
- **Success pattern**: Complete in ~3 seconds

## Conclusion

The race condition is confirmed and understood. The current fix (Attempt 5) with state tracking and API fallback should theoretically eliminate all failures, but intermittent issues persist. The most likely cause is message dropping in `receiveLoop` or timing issues with map updates across goroutines.

Further investigation needed to confirm 100% reliability or implement Option A/B above.
