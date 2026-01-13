# Backend Lifecycle Management Design

**Date:** 2026-01-02
**Status:** Approved for implementation

## Overview

Add automatic backend lifecycle management to the Swift macOS app. The app will auto-start the Go API backend if not running and clean it up on exit, providing a seamless user experience.

## Requirements

1. **Auto-start behavior:** When app launches and backend isn't running, start it automatically with brief loading indicator
2. **Cleanup behavior:** Stop backend on exit only if the app started it (preserve manually-started backends)
3. **Error handling:** Gracefully handle backend crashes, port conflicts, and missing binaries

## Architecture

### New Component: BackendManager

Singleton service responsible for backend lifecycle management:

```swift
@MainActor
class BackendManager: ObservableObject {
    @Published var backendStatus: BackendStatus = .unknown
    private var backendProcess: Process?
    private var didStartBackend = false
    private let baseURL = URL(string: "http://localhost:8080")!

    enum BackendStatus {
        case unknown, starting, running, stopped, error(String)
    }
}
```

**Responsibilities:**
- Health check via HTTP to detect if backend is running
- Process spawning and monitoring
- Ownership tracking (did we start it?)
- Graceful shutdown

### Launch Sequence

```
App Launch
   → BackendManager.ensureBackendRunning()
      → Try health check request (timeout: 500ms)
      → If fails: Start process, wait for ready (~1-2s), set didStartBackend=true
      → If succeeds: Set didStartBackend=false
   → EmulatorViewModel.initialize()
   → Main UI appears
```

### Shutdown Sequence

```
App Exit
   → EmulatorViewModel.cleanup() (destroy session, disconnect WebSocket)
   → BackendManager.shutdown()
      → If didStartBackend==true: Terminate process gracefully (SIGTERM)
      → If didStartBackend==false: Do nothing
```

## Implementation Details

### Binary Location Strategy

1. **Bundled with app (primary):** `ARMEmulator.app/Contents/Resources/arm-emulator`
   - Include binary in app bundle via build phase
   - Most reliable, no external dependencies

2. **Fallback to development location:** `./arm-emulator` (relative to project root)
   - Development convenience
   - Check during Xcode runs

3. **System PATH fallback:** Check `/usr/local/bin/arm-emulator`
   - For installed versions

### Health Check

```swift
func checkBackendHealth() async -> Bool {
    let healthURL = baseURL.appendingPathComponent("/api/health")
    var request = URLRequest(url: healthURL, timeoutInterval: 0.5)
    request.httpMethod = "GET"

    do {
        let (_, response) = try await URLSession.shared.data(for: request)
        if let httpResponse = response as? HTTPURLResponse {
            return (200...299).contains(httpResponse.statusCode)
        }
        return false
    } catch {
        return false
    }
}
```

**Note:** The Go backend doesn't currently have a `/api/health` endpoint. We'll use `/api/v1/session` as a fallback (it returns 405 Method Not Allowed for GET, but that proves the server is alive).

### Backend Startup

```swift
func startBackend() async throws {
    backendStatus = .starting

    guard let binaryPath = findBinaryPath() else {
        throw BackendError.binaryNotFound
    }

    let process = Process()
    process.executableURL = binaryPath
    process.arguments = ["--api-server", "--port", "8080"]

    // Capture output for debugging
    let outputPipe = Pipe()
    process.standardOutput = outputPipe
    process.standardError = outputPipe

    try process.run()
    backendProcess = process
    didStartBackend = true

    // Poll for readiness
    try await waitForBackendReady(timeout: 10.0)
    backendStatus = .running
}
```

### Ready-Wait Polling

Poll health check every 200ms for up to 10 seconds:

```swift
private func waitForBackendReady(timeout: TimeInterval) async throws {
    let deadline = Date().addingTimeInterval(timeout)

    while Date() < deadline {
        if await checkBackendHealth() {
            return
        }
        try await Task.sleep(nanoseconds: 200_000_000) // 200ms
    }

    throw BackendError.startupTimeout
}
```

### Graceful Shutdown

```swift
func shutdown() async {
    guard didStartBackend, let process = backendProcess else {
        return
    }

    if process.isRunning {
        process.terminate() // SIGTERM

        // Wait up to 3 seconds for graceful exit
        for _ in 0..<15 {
            if !process.isRunning { break }
            try? await Task.sleep(nanoseconds: 200_000_000)
        }

        // Force kill if still running
        if process.isRunning {
            kill(process.processIdentifier, SIGKILL)
        }
    }

    backendProcess = nil
}
```

### App Lifecycle Integration

**App startup:**

```swift
@main
struct ARMEmulatorApp: App {
    @StateObject private var backendManager = BackendManager()

    var body: some Scene {
        WindowGroup {
            MainView()
                .environmentObject(backendManager)
                .task {
                    await backendManager.ensureBackendRunning()
                }
        }
    }
}
```

**App termination:**

Implement `NSApplicationDelegate` to handle cleanup:

```swift
class AppDelegate: NSObject, NSApplicationDelegate {
    var backendManager: BackendManager?

    func applicationWillTerminate(_ notification: Notification) {
        Task {
            await backendManager?.shutdown()
        }
    }
}
```

## Error Handling

### Binary Not Found
- Show alert: "ARM Emulator backend not found. Please rebuild the app or check installation."
- App remains open but non-functional
- Future enhancement: File picker to locate binary manually

### Backend Crashes During Use
- Detect via `Process.terminationHandler`
- Show alert: "Backend process crashed. Restart?"
- Preserve loaded program to reload after restart

### Port Conflict (8080 in use)
- Health check fails, process launch fails
- Show alert: "Port 8080 is in use. Close other apps using this port."
- Future enhancement: Try alternative ports (8081, 8082)

### Backend Becomes Unresponsive
- Periodic health checks during app lifetime (every 30s)
- If 3 consecutive checks fail, show warning
- Offer to restart backend

## UI Integration

### Loading States

```swift
// In MainView
@EnvironmentObject var backendManager: BackendManager

var body: some View {
    ZStack {
        if backendManager.backendStatus == .running {
            EmulatorContentView() // Normal UI
        } else {
            BackendStatusView(status: backendManager.backendStatus)
        }
    }
}
```

### BackendStatusView

Shows appropriate UI for each state:
- `.starting`: Spinner + "Starting ARM Emulator backend..."
- `.error(message)`: Error icon + message + "Retry" button
- `.stopped`: "Backend stopped" + "Start" button
- `.running`: Hidden (main UI visible)

## Testing Strategy

### Unit Tests
- Mock `Process` to test startup/shutdown logic
- Test health check with mock URLSession
- Test ownership tracking (didStartBackend flag)

### Manual Testing Scenarios
1. Fresh launch (no backend running)
2. Launch with backend already running (from CLI)
3. Kill backend while app running (test crash recovery)
4. Launch with port 8080 blocked
5. Force quit app (Cmd+Q) - verify cleanup

## Edge Cases

- **Multiple app instances:** Each checks health, only first starts backend (others detect it's running)
- **Binary permissions:** Check executable bit, show helpful error if not set
- **Architecture mismatch:** Ensure binary architecture matches system (universal binary recommended)
- **Backend already running from CLI:** Health check succeeds, `didStartBackend=false`, don't stop on exit

## Build Integration

### Copy Backend Binary to App Bundle

Add to `project.yml` build phases:

```yaml
postBuildScripts:
  - name: Copy Backend Binary
    script: |
      cp "${PROJECT_DIR}/../arm-emulator" "${BUILT_PRODUCTS_DIR}/${PRODUCT_NAME}.app/Contents/Resources/"
      chmod +x "${BUILT_PRODUCTS_DIR}/${PRODUCT_NAME}.app/Contents/Resources/arm-emulator"
```

This ensures the backend binary is bundled with the app for distribution.

## Future Enhancements

1. **Preference for custom binary path:** Allow users to specify backend location
2. **Auto-update backend:** Check for newer binary versions
3. **Multiple port support:** Try alternative ports if 8080 is busy
4. **Backend log viewer:** Show backend stdout/stderr in app for debugging
5. **Health endpoint in Go backend:** Add dedicated `/api/health` endpoint (currently using `/api/v1/session` as fallback)
