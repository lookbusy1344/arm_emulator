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

**FIXED** - Root cause confirmed and fixed. See fix details below.

# Fix: Stepping After User Input Issue

## Problem
The `swift-gui` (and other clients) experienced a hang when stepping through code that requires user input (e.g., `READ_INT`).
The issue was a deadlock in `DebuggerService` and a logic error in `SendInput`.

1. **Deadlock**: `Step()` held the service mutex (`s.mu`) while calling `vm.Step()`. `vm.Step()` blocks on `stdin` read. `SendInput` attempts to acquire `s.mu` to check VM status, blocking indefinitely.
2. **Logic Error**: `SendInput` checked `IsRunning()` to decide whether to write to the pipe or buffer. During stepping, `IsRunning()` is false, so input was buffered instead of sent to the blocked VM.

## Fix
1. **Unlock during Step**: Modified `Step()` and `StepOver()` in `service/debugger_service.go` to release `s.mu` before calling `vm.Step()`. This allows other service methods (specifically `SendInput`) to execute while the VM is blocked on I/O.
2. **Allow Input when Waiting**: Modified `SendInput()` to check `vm.State == StateWaitingForInput`. If the VM is waiting for input, we write to the pipe immediately, even if `IsRunning()` is false.

## Verification
This fix ensures that:
1. `Step()` does not hold the lock while waiting for input.
2. `SendInput()` correctly detects the "Waiting for Input" state and delivers data to the VM pipe.
3. The VM unblocks, consumes the input, and `Step()` completes.

# Issue: Double-Step After User Input

## Problem
After the initial fix (commit 00a7731), a new issue was discovered: when stepping through code that requires user input, the GUI would execute TWO instructions instead of one after the user provided input.

**Example:**
```
SWI #0x06    ; READ_INT - cursor here
POP {pc}     ; Expected cursor position after input
```

**Expected behavior:** After user sends input, cursor moves to `POP {pc}`
**Actual behavior:** Cursor jumps PAST `POP {pc}` - the POP has already executed

## Root Cause
The Swift GUI's `sendInput()` method unconditionally called `step()` after sending input:

```swift
func sendInput(_ input: String) async {
    try await apiClient.sendStdin(sessionID: sessionID, data: input)

    // Always step after sending input to consume the buffered input
    try await apiClient.step(sessionID: sessionID)  // ← PROBLEM!
}
```

This caused a double-step when the VM was waiting for input during a step operation:
1. User clicks Step → Backend's `vm.Step()` blocks on `SWI #0x06` (READ_INT)
2. User sends input → Backend unblocks from the FIRST step (SWI completes)
3. GUI calls `step()` again → Backend executes SECOND step (`POP {pc}`)

## Fix
Modified `EmulatorViewModel.swift` to check the VM state before deciding whether to auto-step:

```swift
func sendInput(_ input: String) async {
    // Capture status BEFORE sending input
    let wasWaitingForInput = (status == .waitingForInput)

    try await apiClient.sendStdin(sessionID: sessionID, data: input)

    if wasWaitingForInput {
        // VM was waiting - the step() that triggered the input request
        // is still in progress and will complete. DON'T call step() again!
        try await refreshState()
    } else {
        // VM was not waiting - backend buffered the input.
        // Call step() to consume the buffered input.
        try await apiClient.step(sessionID: sessionID)
        try await refreshState()
    }
}
```

## Logic
The backend has two modes for handling input:
- **Buffered mode**: VM is NOT running AND NOT waiting → buffer input for later
- **Immediate mode**: VM IS running OR waiting for input → write to pipe immediately

The Swift GUI now matches this logic:
- If `waiting_for_input` state: Input unblocks existing step, no additional step needed
- Otherwise: Input is buffered, call step() to consume it

## Verification
This fix ensures:
1. Single-step behavior works correctly when VM is waiting for input
2. Batch input functionality is preserved (auto-step when buffering)
3. No double-execution of instructions after user input
