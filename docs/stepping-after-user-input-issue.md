# Issue: Stepping After User Input Not Working

## Summary

When the Swift GUI is stepping through code and hits a syscall that requires user input (e.g., `SWI #0x06` for READ_INT), the program correctly enters `waiting_for_input` state and highlights the input field. However, after the user enters input and presses Send, the program does not resume execution and further stepping becomes impossible.

## Current Behavior

1. User steps through code until reaching `SWI #0x06` (READ_INT)
2. VM enters `waiting_for_input` state
3. UI correctly highlights input field with orange border/background ✅
4. Status shows "Waiting for Input" with keyboard icon ✅
5. User types "5" and presses Send button
6. **`sendStdin` API call hangs indefinitely** ❌
7. No error messages, but stepping no longer works

## Log Evidence

### Successful State Change to waiting_for_input
```
🔵 [ViewModel] WebSocket event received: state
🔵 [ViewModel] State update - status: waiting_for_input, PC: nil
```

### Input Send Hangs
```
🔵 [ViewModel] sendInput() called with input: 5
🔵 [ViewModel] Current status: waitingForInput
🔵 [ViewModel] Sending stdin to backend...
[HANGS HERE - never reaches "Stdin sent successfully"]
```

## Technical Analysis

### Swift GUI Flow

```swift
func sendInput(_ input: String) async {
    // ... validation ...
    
    // This call hangs:
    try await apiClient.sendStdin(sessionID: sessionID, data: input)
    
    // Never reaches here:
    try await apiClient.step(sessionID: sessionID)
    try await refreshState()
}
```

The `sendStdin` call in `APIClient.swift` uses:
```swift
let (data, response) = try await session.data(for: request)  // HANGS HERE
```

### Backend Flow

From `service/debugger_service.go`:

```go
func (s *DebuggerService) SendInput(input string) error {
    // Check if VM is running
    if !s.IsRunning() {
        // Buffer input for later
        s.stdinBuffer.WriteString(input)
        return nil  // Returns immediately
    }
    // ... write to pipe if running ...
}
```

**Key Issue:** When VM is in `waiting_for_input` state during stepping:
- `IsRunning()` returns `false` (VM is paused waiting for input)
- Input gets buffered to `stdinBuffer`
- Function returns `nil` immediately
- **But the HTTP handler never sends a response back to the client!**

### HTTP Handler

From `api/handlers.go`:

```go
func (s *Server) handleSendStdin(w http.ResponseWriter, r *http.Request, sessionID string) {
    // ... get session ...
    
    stdinErr := session.Service.SendInput(req.Data)
    if stdinErr != nil {
        writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to send stdin: %v", stdinErr))
        return
    }
    
    // This should be reached, but HTTP response might not be written
    writeJSON(w, http.StatusOK, SuccessResponse{
        Success: true,
        Message: "Stdin sent",
    })
}
```

## Root Cause Hypothesis

The `handleSendStdin` API endpoint might be:

1. **Blocked/deadlocked** - Possibly holding a mutex that prevents the response from being written
2. **Not flushing response** - The HTTP response might not be getting flushed to the client
3. **Connection issue** - WebSocket or HTTP keep-alive interference
4. **Goroutine deadlock** - Some background goroutine might be blocked waiting for the main execution thread

## Expected Behavior

1. User enters input → `sendStdin` API call completes successfully
2. Input is buffered by backend (since VM is not running)
3. Swift GUI automatically calls `step` API
4. Backend step consumes buffered stdin from `stdinBuffer`
5. Syscall completes, PC advances
6. User can continue stepping normally

## Workaround Attempts

### Attempt 1: Check status before auto-stepping
```swift
if status == .waitingForInput {
    try await apiClient.step(sessionID: sessionID)
}
```
**Result:** Still hangs at `sendStdin` call, never reaches the status check.

### Attempt 2: Always step after stdin
```swift
// Always step after sending input to consume buffered input
try await apiClient.sendStdin(sessionID: sessionID, data: input)
try await apiClient.step(sessionID: sessionID)
```
**Result:** Still hangs at `sendStdin` call.

## Related Code

### Backend
- `service/debugger_service.go` - `SendInput()` method
- `api/handlers.go` - `handleSendStdin()` endpoint
- `vm/syscall.go` - `StateWaitingForInput` handling

### Swift GUI
- `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift` - `sendInput()` method
- `swift-gui/ARMEmulator/Services/APIClient.swift` - `sendStdin()` method
- `swift-gui/ARMEmulator/Views/ConsoleView.swift` - Input field UI

## Testing

### Manual Test
1. Load `examples/fibonacci.s`
2. Step through until "How many Fibonacci numbers to generate (1-20)?" prompt
3. Observe orange highlight on input field (works)
4. Type "5" and press Send
5. **Observe hang** - no further stepping possible

### Expected in TUI
The TUI debugger handles this correctly - after typing input, stepping continues normally.

## Next Steps

1. Add debug logging to backend `handleSendStdin` to see if request is received
2. Check if there's a mutex deadlock in the service layer
3. Investigate if HTTP response is being written but not flushed
4. Consider adding timeout to `sendStdin` API call to prevent indefinite hang
5. Review if WebSocket connection interferes with HTTP request/response

## Workaround for Users

Currently **no workaround** - the GUI becomes unresponsive after sending input during step mode. Users must:
1. Restart the app
2. Use "Run" mode instead of "Step" mode (run mode might handle input differently)
3. Use the TUI debugger instead of the GUI

## Priority

**HIGH** - This blocks a core debugging workflow (stepping through programs with user input).

## Related Issues

- Input waiting indicator works correctly (fixed in commit 0462489)
- PC scrolling works correctly (fixed in commit 9ce1bb2)
- The issue is specifically with the stdin API endpoint not responding

## Date

2026-01-07

## Status

**INVESTIGATING** - Root cause not yet confirmed. Needs backend debugging to determine why HTTP response is not being sent.
