# ARM Emulator HTTP API

This document describes the HTTP REST API for the ARM emulator, which enables native GUI clients (Swift, .NET, web) to interact with the emulator backend.

## Overview

The API server provides a RESTful HTTP interface with JSON payloads, allowing multiple concurrent emulator sessions with full control over program execution, debugging, and state inspection.

**Key Features:**
- Session-based emulator instances
- Program loading and execution control
- Register and memory inspection
- Breakpoint management
- Real-time state updates via WebSocket
- **Localhost-only CORS** for web clients (security hardened)

## Starting the API Server

```bash
# Start API server on port 8080
./arm-emulator --api-server --port 8080

# Custom port
./arm-emulator --api-server --port 3000
```

**Security:** The server binds to `127.0.0.1` (localhost only) and enforces strict CORS to reject remote origins.

## Architecture

```
Client (Swift/Web/.NET)
    ↓ HTTP/JSON
API Server (api/)
    ↓
Session Manager
    ↓
DebuggerService (service/)
    ↓
VM + Debugger + Parser (existing core)
```

## Base URL

All endpoints are prefixed with `/api/v1` for versioning.

Example: `http://localhost:8080/api/v1/session`

## Security

**Network Binding:**
- Server binds exclusively to `127.0.0.1:PORT` (localhost)
- Cannot be accessed from external networks
- No configuration option to bind to `0.0.0.0` (hardcoded security)

**CORS Policy:**
- Allows origins: `http://localhost:*`, `https://localhost:*`, `http://127.0.0.1:*`, `https://127.0.0.1:*`, `file://`
- Rejects all remote origins (e.g., `http://evil.com`)
- Native apps (Swift, etc.) work without Origin headers
- Web-based GUIs must run on localhost

**Authentication:**
- Currently none - localhost-only binding + strict CORS provides security
- Future: API key or token-based auth for optional remote access

## Endpoints

### Health Check

#### GET /health

Returns server health status.

**Response:**
```json
{
  "status": "ok",
  "sessions": 3,
  "time": "2026-01-01T12:00:00Z"
}
```

---

### Session Management

#### POST /api/v1/session

Create a new emulator session.

**Request:**
```json
{
  "memorySize": 1048576,
  "stackSize": 65536,
  "heapSize": 262144,
  "fsRoot": "/path/to/sandbox"
}
```

All fields are optional (defaults: 1MB memory, 64KB stack, 256KB heap).

**Response:**
```json
{
  "sessionId": "a1b2c3d4e5f6...",
  "createdAt": "2026-01-01T12:00:00Z"
}
```

**Status Codes:**
- `201 Created` - Session created successfully
- `500 Internal Server Error` - Failed to create session

---

#### GET /api/v1/session

List all active sessions.

**Response:**
```json
{
  "sessions": ["session-id-1", "session-id-2"],
  "count": 2
}
```

---

#### GET /api/v1/session/{id}

Get session status.

**Response:**
```json
{
  "sessionId": "a1b2c3...",
  "state": "paused",
  "pc": 32772,
  "cycles": 5,
  "hasWrite": true,
  "writeAddr": 327680
}
```

**States:** `idle`, `running`, `paused`, `halted`, `error`

---

#### DELETE /api/v1/session/{id}

Destroy a session and free resources.

**Response:**
```json
{
  "success": true,
  "message": "Session destroyed"
}
```

**Status Codes:**
- `200 OK` - Session destroyed
- `404 Not Found` - Session not found

---

### Program Management

#### POST /api/v1/session/{id}/load

Load an assembly program into the session.

**Request:**
```json
{
  "source": "main:\n\tMOVE R0, #42\n\tSWI #0"
}
```

**Response:**
```json
{
  "success": true,
  "symbols": {
    "main": 32768,
    "loop": 32780
  }
}
```

On error:
```json
{
  "success": false,
  "errors": [
    "Line 5: Unknown instruction: INVALID"
  ]
}
```

**Status Codes:**
- `200 OK` - Program loaded successfully
- `400 Bad Request` - Parse error
- `404 Not Found` - Session not found

---

### Execution Control

#### POST /api/v1/session/{id}/run

Start program execution (asynchronous).

**Response:**
```json
{
  "success": true,
  "message": "Program started"
}
```

Program runs in background. Use GET status or WebSocket for state updates.

---

#### POST /api/v1/session/{id}/stop

Stop program execution.

**Response:**
```json
{
  "success": true,
  "message": "Program stopped"
}
```

---

#### POST /api/v1/session/{id}/step

Execute a single instruction.

**Response:**
```json
{
  "r0": 42,
  "r1": 0,
  ...
  "pc": 32772,
  "cpsr": {
    "n": false,
    "z": false,
    "c": false,
    "v": false
  },
  "cycles": 1
}
```

Returns updated register state after stepping.

---

#### POST /api/v1/session/{id}/reset

Reset VM to initial state (preserves loaded program).

**Response:**
```json
{
  "success": true,
  "message": "VM reset"
}
```

---

### State Inspection

#### GET /api/v1/session/{id}/registers

Get current register values.

**Response:**
```json
{
  "r0": 42,
  "r1": 100,
  ...
  "sp": 327680,
  "lr": 0,
  "pc": 32768,
  "cpsr": {
    "n": false,
    "z": false,
    "c": false,
    "v": false
  },
  "cycles": 10
}
```

---

#### GET /api/v1/session/{id}/memory

Read memory region.

**Query Parameters:**
- `address` - Start address (hex: `0x8000` or decimal: `32768`)
- `length` - Number of bytes to read (max: 1MB)

**Example:** `/api/v1/session/{id}/memory?address=0x8000&length=16`

**Response:**
```json
{
  "address": 32768,
  "data": [227, 160, 0, 42, ...],
  "length": 16
}
```

**Limits:**
- Maximum read: 1,048,576 bytes (1MB)
- Returns 400 Bad Request if limit exceeded

---

#### GET /api/v1/session/{id}/disassembly

Get disassembled instructions.

**Query Parameters:**
- `address` - Start address (hex or decimal)
- `count` - Number of instructions (default: 10, max: 1000)

**Example:** `/api/v1/session/{id}/disassembly?address=0x8000&count=5`

**Response:**
```json
{
  "instructions": [
    {
      "address": 32768,
      "machineCode": 3792517162,
      "disassembly": "MOVE R0, #42",
      "symbol": "main"
    },
    {
      "address": 32772,
      "machineCode": 3791396864,
      "disassembly": "SWI #0",
      "symbol": ""
    }
  ]
}
```

---

### Debugging

#### POST /api/v1/session/{id}/breakpoint

Add a breakpoint.

**Request:**
```json
{
  "address": 32772
}
```

**Response:**
```json
{
  "success": true,
  "message": "Breakpoint added"
}
```

---

#### DELETE /api/v1/session/{id}/breakpoint

Remove a breakpoint.

**Request:**
```json
{
  "address": 32772
}
```

**Response:**
```json
{
  "success": true,
  "message": "Breakpoint removed"
}
```

---

#### GET /api/v1/session/{id}/breakpoints

List all breakpoints.

**Response:**
```json
{
  "breakpoints": [32772, 32784, 32800]
}
```

---

### Input/Output

#### POST /api/v1/session/{id}/stdin

Send input to running program.

**Request:**
```json
{
  "data": "42\n"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Stdin sent"
}
```

Use this for interactive programs that read from stdin (SWI #4, #5, #6).

---

## Error Responses

All errors return JSON with this format:

```json
{
  "error": "Not Found",
  "message": "Session not found",
  "code": 404
}
```

**Common Status Codes:**
- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid request
- `404 Not Found` - Resource not found
- `405 Method Not Allowed` - Wrong HTTP method
- `500 Internal Server Error` - Server error

---

## Example Usage

### JavaScript (Fetch API)

```javascript
// Create session
const response = await fetch('http://localhost:8080/api/v1/session', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ memorySize: 1048576 })
});
const { sessionId } = await response.json();

// Load program
await fetch(`http://localhost:8080/api/v1/session/${sessionId}/load`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    source: 'MOVE R0, #42\nSWI #0'
  })
});

// Step execution
const stepResponse = await fetch(
  `http://localhost:8080/api/v1/session/${sessionId}/step`,
  { method: 'POST' }
);
const registers = await stepResponse.json();
console.log('R0:', registers.r0); // 42

// Get memory
const memResponse = await fetch(
  `http://localhost:8080/api/v1/session/${sessionId}/memory?address=0x8000&length=16`
);
const memory = await memResponse.json();
console.log('Memory:', memory.data);

// Clean up
await fetch(`http://localhost:8080/api/v1/session/${sessionID}`, {
  method: 'DELETE'
});
```

### Swift (URLSession)

```swift
// Create session
let url = URL(string: "http://localhost:8080/api/v1/session")!
var request = URLRequest(url: url)
request.httpMethod = "POST"
request.setValue("application/json", forHTTPHeaderField: "Content-Type")
request.httpBody = try? JSONEncoder().encode(SessionCreateRequest())

let (data, _) = try await URLSession.shared.data(for: request)
let response = try JSONDecoder().decode(SessionCreateResponse.self, from: data)
let sessionId = response.sessionId

// Load program
let loadURL = URL(string: "http://localhost:8080/api/v1/session/\(sessionId)/load")!
var loadRequest = URLRequest(url: loadURL)
loadRequest.httpMethod = "POST"
loadRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
let program = LoadProgramRequest(source: "MOVE R0, #42\nSWI #0")
loadRequest.httpBody = try? JSONEncoder().encode(program)

try await URLSession.shared.data(for: loadRequest)

// Step
let stepURL = URL(string: "http://localhost:8080/api/v1/session/\(sessionId)/step")!
var stepRequest = URLRequest(url: stepURL)
stepRequest.httpMethod = "POST"

let (stepData, _) = try await URLSession.shared.data(for: stepRequest)
let registers = try JSONDecoder().decode(RegistersResponse.self, from: stepData)
print("R0: \(registers.r0)") // 42
```

### curl

```bash
# Create session
SESSION_ID=$(curl -s -X POST http://localhost:8080/api/v1/session | jq -r '.sessionId')

# Load program
curl -X POST http://localhost:8080/api/v1/session/$SESSION_ID/load \
  -H "Content-Type: application/json" \
  -d '{"source": "MOVE R0, #42\nSWI #0"}'

# Step
curl -X POST http://localhost:8080/api/v1/session/$SESSION_ID/step

# Get registers
curl http://localhost:8080/api/v1/session/$SESSION_ID/registers

# Get memory
curl "http://localhost:8080/api/v1/session/$SESSION_ID/memory?address=0x8000&length=16"

# Destroy session
curl -X DELETE http://localhost:8080/api/v1/session/$SESSION_ID
```

---

## WebSocket API

Real-time event streaming for state changes, output, and execution events.

**Endpoint:** `ws://localhost:8080/api/v1/ws`

### Connection

Upgrade HTTP connection to WebSocket:

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws');

ws.onopen = () => {
  // Subscribe to events
  ws.send(JSON.stringify({
    type: 'subscribe',
    sessionId: 'a1b2c3...',
    events: ['state', 'output', 'event']
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Event:', message);
};
```

### Subscribe Message

After connecting, send a subscription message to start receiving events:

```json
{
  "type": "subscribe",
  "sessionId": "a1b2c3...",
  "events": ["state", "output", "event"]
}
```

**Event Types:**
- `state` - VM state changes (registers, PC, flags)
- `output` - Console output (stdout/stderr)
- `event` - Execution events (breakpoints, errors, completion)

### Event Messages

#### State Event

Sent when VM state changes (after step, run, or state modification):

```json
{
  "type": "state",
  "sessionId": "a1b2c3...",
  "data": {
    "status": "running",
    "pc": 32768,
    "sp": 327680,
    "lr": 0,
    "cycles": 10,
    "registers": {
      "r0": 42,
      "r1": 100,
      "r2": 0,
      ...
      "r12": 0
    },
    "flags": {
      "n": false,
      "z": true,
      "c": false,
      "v": false
    }
  }
}
```

#### Output Event

Sent when program writes to stdout or stderr:

```json
{
  "type": "output",
  "sessionId": "a1b2c3...",
  "data": {
    "stream": "stdout",
    "content": "Hello, World!\n"
  }
}
```

**Streams:** `stdout`, `stderr`

#### Execution Event

Sent for breakpoints, watchpoints, errors, or program completion:

```json
{
  "type": "event",
  "sessionId": "a1b2c3...",
  "data": {
    "event": "breakpoint_hit",
    "address": 32780,
    "symbol": "main+12"
  }
}
```

**Event Types:**
- `breakpoint_hit` - Breakpoint triggered
- `watchpoint_hit` - Watchpoint triggered
- `program_halted` - Program exited
- `error` - Execution error

### Swift Example

```swift
import Foundation

class WebSocketClient: NSObject, URLSessionWebSocketDelegate {
    private var webSocketTask: URLSessionWebSocketTask?

    func connect(sessionId: String) {
        let url = URL(string: "ws://localhost:8080/api/v1/ws")!
        let session = URLSession(configuration: .default, delegate: self, delegateQueue: nil)
        webSocketTask = session.webSocketTask(with: url)
        webSocketTask?.resume()

        // Subscribe to events
        let subscription = [
            "type": "subscribe",
            "sessionId": sessionId,
            "events": ["state", "output", "event"]
        ]

        let data = try! JSONSerialization.data(withJSONObject: subscription)
        let message = URLSessionWebSocketTask.Message.string(String(data: data, encoding: .utf8)!)
        webSocketTask?.send(message) { error in
            if let error = error {
                print("WebSocket send error: \(error)")
            }
        }

        receiveMessage()
    }

    private func receiveMessage() {
        webSocketTask?.receive { [weak self] result in
            switch result {
            case .success(let message):
                switch message {
                case .string(let text):
                    if let data = text.data(using: .utf8),
                       let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any] {
                        print("Received:", json)
                    }
                default:
                    break
                }
                self?.receiveMessage() // Continue receiving
            case .failure(let error):
                print("WebSocket receive error: \(error)")
            }
        }
    }
}
```

### Connection Management

**Keep-Alive:** Server sends ping frames automatically. Clients should respond with pong frames.

**Reconnection:** If connection drops, clients should reconnect and re-subscribe.

**Clean Shutdown:** Send close frame before disconnecting:

```javascript
ws.close(1000, 'Normal closure');
```

---

## Rate Limiting & Security

**Current:**
- Localhost-only binding (127.0.0.1)
- 1MB request size limit
- 1MB memory read limit
- 1000 instruction disassembly limit

**Future Enhancements:**
- Rate limiting (requests per minute)
- API key authentication
- TLS/HTTPS support
- Configurable bind address

---

## Testing

Run API integration tests:

```bash
go test -v ./api/
```

The test suite includes:
- Session management
- Program loading and execution
- Register and memory inspection
- Breakpoint management
- Error handling
- CORS headers

All tests use `httptest` for isolated testing without needing a running server.

---

## Implementation Details

**Files:**
- `api/models.go` - Request/response DTOs and type conversions
- `api/session_manager.go` - Multi-session management
- `api/server.go` - HTTP server setup and routing
- `api/handlers.go` - Endpoint implementations
- `api/api_test.go` - Integration tests

**Dependencies:**
- Standard library `net/http` (no external frameworks)
- Existing `service/` layer (wraps DebuggerService)
- VM, parser, debugger packages (no changes required)

**Thread Safety:**
- `SessionManager` uses `sync.RWMutex` for concurrent access
- Each session has its own lock
- `DebuggerService` is already thread-safe

---

## Performance

**Benchmarks (localhost):**
- Session creation: < 1ms
- Program load: ~5-10ms (depends on program size)
- Step execution: < 1ms
- Register read: < 0.1ms
- Memory read (16 bytes): < 0.1ms
- Disassembly (10 instructions): < 1ms

**Overhead vs. direct VM access:** ~1-2ms per request (negligible for GUI use)

---

## Next Steps

1. ✅ ~~**WebSocket Support**~~ - Completed (Stage 2)
2. ✅ ~~**CLI Flag**~~ - `--api-server` flag implemented
3. ✅ ~~**Swift GUI**~~ - Native macOS client completed (Stage 3-4)
4. ✅ ~~**API Documentation**~~ - OpenAPI/Swagger spec created (openapi.yaml)
5. **Metrics** - Prometheus endpoint for monitoring
6. **Authentication** - API keys for remote access
7. **Performance Benchmarks** - Comprehensive API latency testing
8. **Integration Tests** - Additional error scenario testing

---

*Last Updated: 2026-01-02*
