# Swift Native GUI Planning Document

## Executive Summary

This document outlines the plan for building a native Swift macOS GUI for the ARM2 emulator, along with a cross-platform architecture that enables native front-ends on Windows (.NET), Linux, and other platforms. The recommended approach uses a **Go-based API server** that exposes the emulator engine through a well-defined interface, allowing multiple native front-end implementations while keeping all core logic in Go.

**Key Benefits:**
- Native macOS experience with SwiftUI
- Cross-platform capability (.NET, Electron, web)
- Clean separation of concerns
- Reuses 100% of existing Go codebase
- Enables headless automation and testing
- Better performance than Wails for native UI responsiveness

**Estimated Timeline:** 6-8 weeks for full implementation across all platforms

## Implementation Status

**Current Progress:** Stage 1 Complete (1/7 stages)

| Stage | Status | Completion |
|-------|--------|------------|
| Stage 1: Backend API Foundation | ✅ Complete | 2026-01-02 |
| Stage 2: WebSocket Real-Time Updates | 🔜 Next | - |
| Stage 3: Swift macOS App Foundation | ⏸️ Pending | - |
| Stage 4: Advanced Swift UI Features | ⏸️ Pending | - |
| Stage 5: Backend Enhancements | ⏸️ Pending | - |
| Stage 6: Polish & Testing | ⏸️ Pending | - |
| Stage 7: Cross-Platform Foundation | ⏸️ Pending | - |

**Latest Achievement:** Production-ready HTTP REST API with 16 endpoints, 17 passing integration tests, zero linting issues, and comprehensive documentation. Fully tested and ready for Swift/Web/.NET clients.

---

## 1. Technical Options Analysis

### Option A: Direct Swift-Go Interop via C Bridge

**Approach:** Export Go functions via `cgo`, create C header, import into Swift

**Pros:**
- Single process (lower latency)
- No network overhead
- Direct memory sharing possible

**Cons:**
- Complex marshaling between Swift ↔ C ↔ Go
- Platform-specific builds (fat binaries for universal macOS)
- Difficult cross-platform support (doesn't help Windows/.NET)
- cgo limitations (goroutine scheduling, callbacks)
- Memory management complexity
- Hard to support multiple concurrent clients
- Not suitable for .NET on Windows

**Verdict:** ❌ Not recommended due to cross-platform limitations

### Option B: Go Shared Library (.dylib/.dll/.so)

**Approach:** Compile Go as C-compatible shared library, load dynamically

**Pros:**
- Language-agnostic interface
- Works for Swift, .NET, Python, etc.
- Single-process deployment possible

**Cons:**
- C FFI complexity for all languages
- Callback handling is difficult
- State management across language boundaries
- Real-time updates require polling or complex callbacks
- Still requires per-platform builds

**Verdict:** ❌ Not recommended due to complexity and callback limitations

### Option C: HTTP/WebSocket API Server (RECOMMENDED)

**Approach:** Go backend runs as API server, native GUIs connect as clients

**Pros:**
- ✅ Clean separation of concerns
- ✅ Cross-platform by design
- ✅ Easy real-time updates via WebSocket
- ✅ Multiple concurrent clients (GUI + CLI + automation)
- ✅ Headless server mode for testing
- ✅ Standard HTTP/JSON/WebSocket protocols
- ✅ Can run backend on remote machine
- ✅ Enables web-based GUI as well
- ✅ Simple debugging (inspect traffic, curl, Postman)
- ✅ Natural authentication/authorization if needed

**Cons:**
- Slight overhead from serialization (negligible for emulator use case)
- Two processes to manage (mitigated by launcher)
- Network port binding (use localhost)

**Verdict:** ✅ **RECOMMENDED** - Best balance of simplicity, flexibility, and cross-platform support

---

## 2. Recommended Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   SwiftUI    │  │ .NET WPF/    │  │   Wails      │      │
│  │   (macOS)    │  │  Avalonia    │  │ (Existing)   │      │
│  │              │  │  (Windows)   │  │              │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                 │                 │              │
│         └─────────────────┼─────────────────┘              │
│                           │                                │
└───────────────────────────┼────────────────────────────────┘
                            │
                    HTTP/REST + WebSocket
                            │
┌───────────────────────────┼────────────────────────────────┐
│                           │                                │
│  ┌────────────────────────▼───────────────────────────┐    │
│  │          API Server Layer (Go)                     │    │
│  │  - HTTP/REST endpoints                             │    │
│  │  - WebSocket for real-time updates                 │    │
│  │  - Session management                              │    │
│  │  - JSON serialization                              │    │
│  └────────────────────────┬───────────────────────────┘    │
│                           │                                │
│  ┌────────────────────────▼───────────────────────────┐    │
│  │      Service Layer (Go) - NEW                      │    │
│  │  - EmulatorService: VM lifecycle, execution        │    │
│  │  - DebuggerService: Breakpoints, stepping          │    │
│  │  - FileService: Load/save programs                 │    │
│  │  - ConfigService: Settings management              │    │
│  │  - TraceService: Diagnostics, statistics           │    │
│  └────────────────────────┬───────────────────────────┘    │
│                           │                                │
│  ┌────────────────────────▼───────────────────────────┐    │
│  │         Core Engine (Go) - EXISTING                │    │
│  │  - vm/ - Virtual machine                           │    │
│  │  - parser/ - Assembly parser                       │    │
│  │  - debugger/ - Debugger logic                      │    │
│  │  - instructions/ - Instruction implementations     │    │
│  │  - encoder/ - Machine code encoder/decoder         │    │
│  └────────────────────────────────────────────────────┘    │
│                                                            │
│                  Go ARM Emulator Backend                   │
└────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### Go Backend Components

1. **API Server Layer** (`api/server.go`)
   - HTTP server (Gin or standard library)
   - WebSocket upgrade handler
   - Request routing
   - CORS configuration
   - Request validation

2. **Service Layer** (new package: `service/`)
   - **EmulatorService**: VM creation, program loading, execution control
   - **DebuggerService**: Breakpoint management, stepping, state inspection
   - **FileService**: File I/O, recent files, examples
   - **ConfigService**: Configuration management
   - **TraceService**: Execution tracing, statistics, diagnostics
   - Thread-safe access to VM instances
   - Session management (multi-user support)

3. **Core Engine** (existing packages: `vm/`, `parser/`, `debugger/`, etc.)
   - No changes required
   - Accessed exclusively through service layer

#### Native Client Components

1. **Swift macOS App**
   - SwiftUI for native UI
   - Combine for reactive state management
   - URLSession for HTTP, WebSocket
   - Sandboxed app with security entitlements

2. **.NET Windows/Linux App** (future)
   - WPF (Windows) or Avalonia (cross-platform)
   - HttpClient, WebSocket client
   - MVVM architecture

---

## 3. API Design

### REST Endpoints

#### Session Management

```
POST   /api/v1/session              Create new emulator session
DELETE /api/v1/session/:id          Destroy session
GET    /api/v1/session/:id/status   Get session status
```

#### Program Management

```
POST   /api/v1/session/:id/load     Load assembly program
GET    /api/v1/session/:id/program  Get current program source
POST   /api/v1/session/:id/assemble Assemble program
GET    /api/v1/session/:id/symbols  Get symbol table
```

#### Execution Control

```
POST   /api/v1/session/:id/run      Start execution
POST   /api/v1/session/:id/stop     Stop execution
POST   /api/v1/session/:id/step     Step single instruction
POST   /api/v1/session/:id/reset    Reset VM state
POST   /api/v1/session/:id/stdin    Send stdin data
```

#### State Inspection

```
GET    /api/v1/session/:id/registers    Get all registers
GET    /api/v1/session/:id/memory       Get memory range
GET    /api/v1/session/:id/stack        Get stack view
GET    /api/v1/session/:id/disassembly  Get disassembly
```

#### Debugging

```
POST   /api/v1/session/:id/breakpoint       Add breakpoint
DELETE /api/v1/session/:id/breakpoint/:addr Remove breakpoint
GET    /api/v1/session/:id/breakpoints      List breakpoints
POST   /api/v1/session/:id/watchpoint       Add watchpoint
DELETE /api/v1/session/:id/watchpoint/:addr Remove watchpoint
```

#### Configuration & Files

```
GET    /api/v1/config                Get configuration
PUT    /api/v1/config                Update configuration
GET    /api/v1/examples              List example programs
GET    /api/v1/examples/:name        Get example program
GET    /api/v1/recent                Get recent files
```

### WebSocket Events

**Client → Server:**
```json
{
  "type": "subscribe",
  "sessionId": "abc123",
  "events": ["state", "output", "trace"]
}
```

**Server → Client:**

State updates:
```json
{
  "type": "state",
  "sessionId": "abc123",
  "data": {
    "status": "running",
    "pc": 32768,
    "registers": {...},
    "flags": {...}
  }
}
```

Console output:
```json
{
  "type": "output",
  "sessionId": "abc123",
  "data": {
    "stream": "stdout",
    "content": "Hello, World!\n"
  }
}
```

Execution events:
```json
{
  "type": "event",
  "sessionId": "abc123",
  "data": {
    "event": "breakpoint_hit",
    "address": 32780,
    "symbol": "main+12"
  }
}
```

---

## 4. Implementation Stages

### Stage 1: Backend API Foundation (Week 1-2) ✅ **COMPLETED**

**Status:** ✅ Completed on 2026-01-02 (Initial implementation: 2026-01-01, Tests fixed: 2026-01-02)

**Goals:**
- ✅ Create service layer abstraction
- ✅ Implement HTTP API server
- ✅ Basic session management
- ✅ Core endpoints (load, run, step, stop)

**Deliverables:**
1. ✅ ~~New `service/` package~~ **Used existing service/DebuggerService**
2. ✅ `api/session_manager.go` - Multi-session support with crypto-secure IDs
3. ✅ `api/server.go` - HTTP server (standard library, no Gin)
4. ✅ `api/handlers.go` - REST endpoint handlers (16 endpoints)
5. ✅ `api/models.go` - Request/response DTOs
6. ✅ `api/api_test.go` - Comprehensive integration tests (17 tests)
7. ✅ `API.md` - Complete API documentation with examples

**Files Created:**
```
api/
  ├── server.go            # HTTP server setup (192 lines)
  ├── handlers.go          # Endpoint handlers (483 lines)
  ├── models.go            # JSON models (191 lines)
  ├── session_manager.go   # Session lifecycle (134 lines)
  └── api_test.go          # API tests (545 lines)

API.md                     # API documentation (608 lines)
```

**Implementation Notes:**
- Used existing `service/DebuggerService` instead of creating new service layer
- Standard library `net/http` instead of Gin (zero external HTTP dependencies)
- Thread-safe session management with RWMutex
- Crypto-secure session IDs (16-byte random hex)
- Security limits: 1MB request size, 1MB memory reads, 1000 instruction disassembly
- CORS-enabled for web clients
- Localhost-only binding (127.0.0.1) for security

**Endpoints Implemented (16 total):**
- ✅ GET /health - Health check
- ✅ POST /api/v1/session - Create session
- ✅ GET /api/v1/session - List sessions
- ✅ GET /api/v1/session/{id} - Get status
- ✅ DELETE /api/v1/session/{id} - Destroy session
- ✅ POST /api/v1/session/{id}/load - Load program
- ✅ POST /api/v1/session/{id}/run - Start execution
- ✅ POST /api/v1/session/{id}/stop - Stop execution
- ✅ POST /api/v1/session/{id}/step - Single step
- ✅ POST /api/v1/session/{id}/reset - Reset VM
- ✅ GET /api/v1/session/{id}/registers - Read registers
- ✅ GET /api/v1/session/{id}/memory - Read memory
- ✅ GET /api/v1/session/{id}/disassembly - Disassemble
- ✅ POST/DELETE /api/v1/session/{id}/breakpoint - Manage breakpoints
- ✅ GET /api/v1/session/{id}/breakpoints - List breakpoints
- ✅ POST /api/v1/session/{id}/stdin - Send input

**Success Criteria:**
- ✅ Can create session via API
- ✅ Can load and execute program via API
- ✅ Can retrieve registers and memory via API
- ✅ All endpoints return proper HTTP status codes
- ✅ Error handling with JSON error responses
- ✅ Comprehensive test coverage (17 integration tests, all passing)
- ✅ Full documentation with JavaScript, Swift, and curl examples
- ✅ Zero linting issues (golangci-lint)
- ✅ All tests passing across entire codebase (1,024+ tests)

**Commits:**
- f91c11d - "Implement HTTP REST API backend for cross-platform GUI support" (2026-01-01)
- TBD - "Fix API compilation errors, add proper error handling, and ensure all tests pass" (2026-01-02)

**Fixes Applied (2026-01-02):**
- Fixed method signature mismatches (GetRegisterState, Continue/Pause, GetMemory, GetDisassembly, SendInput)
- Implemented assembly parsing in LoadProgram endpoint with proper entry point detection
- Added comprehensive error handling for Reset, AddBreakpoint, RemoveBreakpoint
- Fixed CORS middleware application for proper OPTIONS handling
- Added proper integer overflow guards with security annotations
- Removed unused code (session mutex, memSize variable)
- Fixed test programs to include `.org 0x8000` directives
- Corrected ARM assembly syntax (MOVE → MOV)

### Stage 2: WebSocket Real-Time Updates (Week 2-3)

**Goals:**
- Implement WebSocket server
- Event broadcasting system
- Real-time state updates during execution

**Deliverables:**
1. `api/websocket.go` - WebSocket upgrade and handler
2. `api/broadcaster.go` - Event broadcasting to subscribed clients
3. State change notifications (PC, registers, flags)
4. Output streaming (stdout, stderr)
5. Event notifications (breakpoints, errors)

**Technical Details:**
- Use `gorilla/websocket` library
- One goroutine per WebSocket connection
- Broadcast channel for events
- Subscription filtering by session ID

**Success Criteria:**
- Client connects via WebSocket
- Receives real-time updates during execution
- Output appears immediately as program runs
- Breakpoint events trigger notifications

### Stage 3: Swift macOS App Foundation (Week 3-4)

**Goals:**
- Create SwiftUI project
- API client implementation
- Basic UI structure

**Deliverables:**
1. Xcode project with SwiftUI
2. `APIClient.swift` - HTTP REST client
3. `WebSocketClient.swift` - WebSocket client
4. `EmulatorSession.swift` - Session model
5. `MainView.swift` - Main window layout
6. `EditorView.swift` - Assembly editor
7. `RegistersView.swift` - Register display
8. `ConsoleView.swift` - Output console

**UI Structure:**
```
┌─────────────────────────────────────────────────┐
│  ARM Emulator                      [□] [◊] [✕]  │
├─────────────────────────────────────────────────┤
│  File  Edit  Run  Debug  View  Help             │
├──────────────────┬──────────────────────────────┤
│                  │                              │
│  Source Editor   │   Registers & Flags          │
│  (Assembly)      │   ┌──────────────────────┐   │
│                  │   │ R0:  0x00000000      │   │
│  Line numbers    │   │ R1:  0x00000000      │   │
│  Syntax          │   │ ...                  │   │
│  highlighting    │   │ PC:  0x00008000      │   │
│                  │   │ CPSR: ----           │   │
│                  │   └──────────────────────┘   │
│                  │                              │
│                  │   Memory View                │
│                  │   ┌──────────────────────┐   │
│                  │   │ 0x00008000: E3A0...  │   │
│                  │   └──────────────────────┘   │
│                  │                              │
├──────────────────┴──────────────────────────────┤
│  Console Output                                 │
│  > Hello, World!                                │
│  > Program exited with code 0                   │
│                                                 │
└─────────────────────────────────────────────────┘
│  [▶ Run] [⏸ Pause] [⏹ Stop] [⏭ Step] ● Running │
└─────────────────────────────────────────────────┘
```

**Swift Project Structure:**
```
ARMEmulator/
  ├── ARMEmulatorApp.swift     # App entry point
  ├── Models/
  │   ├── EmulatorSession.swift
  │   ├── Register.swift
  │   ├── MemoryRegion.swift
  │   └── ProgramState.swift
  ├── Services/
  │   ├── APIClient.swift
  │   ├── WebSocketClient.swift
  │   └── FileManager.swift
  ├── Views/
  │   ├── MainView.swift
  │   ├── EditorView.swift
  │   ├── RegistersView.swift
  │   ├── MemoryView.swift
  │   ├── ConsoleView.swift
  │   └── ToolbarView.swift
  ├── ViewModels/
  │   ├── EmulatorViewModel.swift
  │   └── EditorViewModel.swift
  └── Resources/
      └── Info.plist
```

**Success Criteria:**
- Swift app launches and shows UI
- Connects to Go backend API
- Can load assembly program
- Can execute and see output
- Registers update in real-time

### Stage 4: Advanced Swift UI Features (Week 4-5)

**Goals:**
- Complete feature parity with Wails GUI
- Debugging features
- Syntax highlighting
- File management

**Deliverables:**
1. Syntax highlighting for assembly
2. Breakpoint gutter in editor
3. Disassembly view
4. Stack visualization
5. Memory hex dump view
6. File open/save dialogs
7. Recent files menu
8. Examples browser
9. Preferences window
10. Toolbar with controls

**Features:**
- **Syntax Highlighting**: Custom TextEditor with NSTextView
- **Breakpoints**: Click gutter to add/remove, visual indicators
- **Disassembly**: Side-by-side source and machine code
- **Memory View**: Hex dump with ASCII, scrollable regions
- **Stack View**: SP visualization, push/pop tracking
- **File Dialogs**: Native macOS open/save panels
- **Drag & Drop**: Drop .s files into editor

**Success Criteria:**
- All Wails features available in Swift
- Native macOS look and feel
- Keyboard shortcuts work (Cmd+R, Cmd+S, etc.)
- Preferences persist across launches

### Stage 5: Backend Enhancements (Week 5-6)

**Goals:**
- Complete remaining API endpoints
- Debugging API
- Trace/statistics API
- Configuration API

**Deliverables:**
1. Debugger endpoints (breakpoints, watchpoints, step)
2. Trace endpoints (execution trace, coverage, stack trace)
3. Statistics endpoints (performance stats)
4. Configuration endpoints (get/set config)
5. File management endpoints (recent files, examples)
6. Input handling (stdin queue for interactive programs)

**API Additions:**
```go
// Debugger
POST   /api/v1/session/:id/breakpoint
DELETE /api/v1/session/:id/breakpoint/:addr
GET    /api/v1/session/:id/breakpoints
POST   /api/v1/session/:id/watchpoint
DELETE /api/v1/session/:id/watchpoint/:addr

// Tracing
POST   /api/v1/session/:id/trace/enable
POST   /api/v1/session/:id/trace/disable
GET    /api/v1/session/:id/trace/data
GET    /api/v1/session/:id/stats

// Input
POST   /api/v1/session/:id/stdin
```

**Success Criteria:**
- Can set/remove breakpoints via API
- Can enable/disable tracing via API
- Statistics available as JSON
- Interactive programs work with stdin queue

### Stage 6: Polish & Testing (Week 6-7)

**Goals:**
- End-to-end testing
- Performance optimization
- Error handling
- Documentation

**Deliverables:**
1. Integration tests for all API endpoints
2. Swift UI tests
3. Performance benchmarks
4. Error scenario testing
5. API documentation (OpenAPI/Swagger)
6. Swift app documentation
7. User guide updates

**Testing Focus:**
- Concurrent sessions
- Long-running programs
- Large programs (memory pressure)
- Network failures (reconnection)
- Backend crash recovery
- Memory leak detection

**Success Criteria:**
- All tests pass
- No memory leaks in Swift or Go
- API latency < 10ms for most operations
- WebSocket updates < 16ms (60fps)
- Swift app feels snappy and responsive

### Stage 7: Cross-Platform Foundation (Week 7-8)

**Goals:**
- Prepare for .NET client
- Launcher/installer
- Documentation

**Deliverables:**
1. Cross-platform API client library (Go)
2. .NET client library (C#) - basic implementation
3. Launcher app (manages backend process)
4. macOS app bundle with embedded backend
5. Installation guide
6. API reference documentation

**Launcher Functionality:**
- Start Go backend on launch
- Health check (wait for server ready)
- Auto-restart on crash
- Graceful shutdown
- Log file management

**macOS App Bundle:**
```
ARMEmulator.app/
  Contents/
    MacOS/
      ARMEmulator          # Swift binary
      arm-emulator-server  # Go backend
    Resources/
      examples/
      docs/
    Info.plist
```

**Success Criteria:**
- Swift app starts backend automatically
- Backend dies when app quits
- Windows user can connect .NET client to backend
- Cross-platform API documentation complete

---

## 5. Detailed Component Design

### Go Service Layer

#### EmulatorService Interface

```go
package service

type EmulatorService interface {
    // Session management
    CreateSession(opts SessionOptions) (sessionID string, err error)
    DestroySession(sessionID string) error
    GetSession(sessionID string) (*Session, error)

    // Program management
    LoadProgram(sessionID string, source string) error
    AssembleProgram(sessionID string) (*AssembleResult, error)
    GetSymbols(sessionID string) (map[string]uint32, error)

    // Execution control
    Run(sessionID string) error
    Stop(sessionID string) error
    Step(sessionID string) error
    Reset(sessionID string) error
    SendStdin(sessionID string, data string) error

    // State inspection
    GetRegisters(sessionID string) (*RegisterState, error)
    GetMemory(sessionID string, addr, length uint32) ([]byte, error)
    GetDisassembly(sessionID string, addr, count uint32) ([]*Instruction, error)
    GetStatus(sessionID string) (*VMStatus, error)

    // Event subscription
    Subscribe(sessionID string, eventTypes []EventType) (<-chan Event, error)
    Unsubscribe(sessionID string, subscription <-chan Event) error
}

type Session struct {
    ID        string
    VM        *vm.VM
    Debugger  *debugger.Debugger
    Source    string
    CreatedAt time.Time
    Status    VMStatus
    mu        sync.RWMutex
}

type VMStatus struct {
    State       string // "idle", "running", "paused", "halted", "error"
    PC          uint32
    Instruction string
    CycleCount  uint64
    Error       string
}
```

#### DebuggerService Interface

```go
package service

type DebuggerService interface {
    AddBreakpoint(sessionID string, addr uint32) error
    RemoveBreakpoint(sessionID string, addr uint32) error
    ListBreakpoints(sessionID string) ([]uint32, error)

    AddWatchpoint(sessionID string, addr uint32, condition WatchCondition) error
    RemoveWatchpoint(sessionID string, addr uint32) error
    ListWatchpoints(sessionID string) ([]*Watchpoint, error)

    StepOver(sessionID string) error
    StepInto(sessionID string) error
    StepOut(sessionID string) error
    Continue(sessionID string) error
}
```

### Swift Client Architecture

#### APIClient

```swift
import Foundation
import Combine

class APIClient: ObservableObject {
    private let baseURL: URL
    private let session: URLSession

    init(baseURL: URL = URL(string: "http://localhost:8080")!) {
        self.baseURL = baseURL
        self.session = URLSession.shared
    }

    // Session management
    func createSession(options: SessionOptions) async throws -> String {
        let url = baseURL.appendingPathComponent("/api/v1/session")
        return try await post(url: url, body: options)
    }

    func destroySession(sessionID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)")
        try await delete(url: url)
    }

    // Program management
    func loadProgram(sessionID: String, source: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/load")
        try await post(url: url, body: ["source": source])
    }

    // Execution control
    func run(sessionID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/run")
        try await post(url: url, body: [:])
    }

    func step(sessionID: String) async throws {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/step")
        try await post(url: url, body: [:])
    }

    // State inspection
    func getRegisters(sessionID: String) async throws -> RegisterState {
        let url = baseURL.appendingPathComponent("/api/v1/session/\(sessionID)/registers")
        return try await get(url: url)
    }

    // Generic helpers
    private func get<T: Decodable>(url: URL) async throws -> T { ... }
    private func post<T: Encodable, R: Decodable>(url: URL, body: T) async throws -> R { ... }
    private func delete(url: URL) async throws { ... }
}
```

#### WebSocketClient

```swift
import Foundation
import Combine

class WebSocketClient: ObservableObject {
    private var webSocket: URLSessionWebSocketTask?
    private let eventSubject = PassthroughSubject<EmulatorEvent, Never>()

    var events: AnyPublisher<EmulatorEvent, Never> {
        eventSubject.eraseToAnyPublisher()
    }

    func connect(sessionID: String) {
        let url = URL(string: "ws://localhost:8080/api/v1/ws")!
        webSocket = URLSession.shared.webSocketTask(with: url)
        webSocket?.resume()

        // Subscribe to events
        let subscription = SubscriptionMessage(
            type: "subscribe",
            sessionId: sessionID,
            events: ["state", "output", "event"]
        )
        send(subscription)

        // Start receiving
        receiveMessage()
    }

    func disconnect() {
        webSocket?.cancel(with: .goingAway, reason: nil)
    }

    private func receiveMessage() {
        webSocket?.receive { [weak self] result in
            switch result {
            case .success(let message):
                if case .string(let text) = message,
                   let data = text.data(using: .utf8),
                   let event = try? JSONDecoder().decode(EmulatorEvent.self, from: data) {
                    self?.eventSubject.send(event)
                }
                self?.receiveMessage() // Continue receiving
            case .failure(let error):
                print("WebSocket error: \(error)")
            }
        }
    }

    private func send<T: Encodable>(_ message: T) {
        guard let data = try? JSONEncoder().encode(message),
              let string = String(data: data, encoding: .utf8) else { return }
        webSocket?.send(.string(string)) { _ in }
    }
}
```

#### EmulatorViewModel

```swift
import Foundation
import Combine

@MainActor
class EmulatorViewModel: ObservableObject {
    @Published var registers: RegisterState = .empty
    @Published var consoleOutput: String = ""
    @Published var status: VMStatus = .idle
    @Published var breakpoints: Set<UInt32> = []

    private let apiClient: APIClient
    private let wsClient: WebSocketClient
    private var sessionID: String?
    private var cancellables = Set<AnyCancellable>()

    init(apiClient: APIClient = APIClient(), wsClient: WebSocketClient = WebSocketClient()) {
        self.apiClient = apiClient
        self.wsClient = wsClient

        // Subscribe to WebSocket events
        wsClient.events
            .receive(on: DispatchQueue.main)
            .sink { [weak self] event in
                self?.handleEvent(event)
            }
            .store(in: &cancellables)
    }

    func loadProgram(source: String) async throws {
        if sessionID == nil {
            sessionID = try await apiClient.createSession(options: .default)
            wsClient.connect(sessionID: sessionID!)
        }

        try await apiClient.loadProgram(sessionID: sessionID!, source: source)
    }

    func run() async throws {
        guard let sessionID = sessionID else { return }
        try await apiClient.run(sessionID: sessionID)
    }

    func step() async throws {
        guard let sessionID = sessionID else { return }
        try await apiClient.step(sessionID: sessionID)

        // Fetch updated state
        registers = try await apiClient.getRegisters(sessionID: sessionID)
    }

    private func handleEvent(_ event: EmulatorEvent) {
        switch event.type {
        case "state":
            if let state = event.data as? StateUpdate {
                registers = state.registers
                status = state.status
            }
        case "output":
            if let output = event.data as? OutputUpdate {
                consoleOutput += output.content
            }
        case "event":
            if let evt = event.data as? ExecutionEvent {
                // Handle breakpoint, error, etc.
            }
        default:
            break
        }
    }
}
```

---

## 6. Cross-Platform Considerations

### Windows (.NET) Client

**Technology Stack:**
- WPF (Windows-only) or Avalonia (cross-platform)
- C# with async/await
- HttpClient for REST
- ClientWebSocket for real-time updates

**Architecture:**
```
ARMEmulatorWPF/
  ├── App.xaml              # Application
  ├── MainWindow.xaml       # Main UI
  ├── Services/
  │   ├── ApiClient.cs      # HTTP REST client
  │   └── WebSocketClient.cs
  ├── ViewModels/
  │   └── EmulatorViewModel.cs
  ├── Views/
  │   ├── EditorView.xaml
  │   ├── RegistersView.xaml
  │   └── ConsoleView.xaml
  └── Models/
      └── EmulatorSession.cs
```

**Similar API Client Pattern:**
```csharp
public class ApiClient
{
    private readonly HttpClient _httpClient;

    public ApiClient(string baseUrl = "http://localhost:8080")
    {
        _httpClient = new HttpClient { BaseAddress = new Uri(baseUrl) };
    }

    public async Task<string> CreateSessionAsync(SessionOptions options)
    {
        var response = await _httpClient.PostAsJsonAsync("/api/v1/session", options);
        return await response.Content.ReadFromJsonAsync<string>();
    }

    public async Task LoadProgramAsync(string sessionId, string source)
    {
        await _httpClient.PostAsJsonAsync($"/api/v1/session/{sessionId}/load",
            new { source });
    }
}
```

### Linux Client

**Options:**
1. **Avalonia** - Cross-platform .NET UI framework
2. **Electron** - Reuse existing Wails web UI
3. **GTK with Python/Go bindings**

**Recommendation:** Avalonia for native .NET experience

---

## 7. Backend Process Management

### Launcher Application

The Swift/WPF app needs to manage the Go backend process lifecycle.

#### Swift Launcher

```swift
import Foundation

class BackendLauncher: ObservableObject {
    @Published var isReady = false
    @Published var error: String?

    private var process: Process?
    private let executablePath: String

    init() {
        // Path to Go backend in app bundle
        self.executablePath = Bundle.main.path(forResource: "arm-emulator-server",
                                                ofType: nil) ?? ""
    }

    func start() {
        process = Process()
        process?.executableURL = URL(fileURLWithPath: executablePath)
        process?.arguments = ["--api-server", "--port", "8080"]

        do {
            try process?.run()

            // Wait for server to be ready
            Task {
                await waitForBackend()
            }
        } catch {
            self.error = "Failed to start backend: \(error)"
        }
    }

    func stop() {
        process?.terminate()
        process?.waitUntilExit()
    }

    private func waitForBackend() async {
        for _ in 0..<30 { // 3 seconds max
            if await checkHealth() {
                isReady = true
                return
            }
            try? await Task.sleep(nanoseconds: 100_000_000) // 100ms
        }
        error = "Backend failed to start"
    }

    private func checkHealth() async -> Bool {
        guard let url = URL(string: "http://localhost:8080/health") else { return false }

        do {
            let (_, response) = try await URLSession.shared.data(from: url)
            return (response as? HTTPURLResponse)?.statusCode == 200
        } catch {
            return false
        }
    }
}
```

#### App Entry Point

```swift
@main
struct ARMEmulatorApp: App {
    @StateObject private var launcher = BackendLauncher()

    var body: some Scene {
        WindowGroup {
            if launcher.isReady {
                MainView()
            } else if let error = launcher.error {
                ErrorView(message: error)
            } else {
                LoadingView()
            }
        }
        .onAppear {
            launcher.start()
        }
        .onDisappear {
            launcher.stop()
        }
    }
}
```

---

## 8. Testing Strategy

### Backend Testing

1. **Unit Tests** (service layer)
   - Test each service method
   - Mock VM/debugger dependencies
   - Verify thread safety

2. **Integration Tests** (API endpoints)
   - Test HTTP handlers with test server
   - Verify JSON serialization
   - Test error responses

3. **E2E Tests**
   - Start real server
   - Execute full workflows (load, run, debug)
   - Test WebSocket events

### Swift App Testing

1. **Unit Tests**
   - Test ViewModels with mock API client
   - Verify state management
   - Test business logic

2. **UI Tests**
   - Test user interactions
   - Verify UI updates
   - Test keyboard shortcuts

3. **Integration Tests**
   - Test with real backend
   - Verify full workflows
   - Performance testing

### Test Example (Go)

```go
func TestEmulatorService_LoadAndRun(t *testing.T) {
    svc := service.NewEmulatorService()

    // Create session
    sessionID, err := svc.CreateSession(service.SessionOptions{})
    require.NoError(t, err)
    defer svc.DestroySession(sessionID)

    // Load program
    program := `
        MOVE R0, #65
        SWI #1    ; WRITE_CHAR
        SWI #0    ; EXIT
    `
    err = svc.LoadProgram(sessionID, program)
    require.NoError(t, err)

    // Run
    err = svc.Run(sessionID)
    require.NoError(t, err)

    // Verify state
    status, err := svc.GetStatus(sessionID)
    require.NoError(t, err)
    assert.Equal(t, "halted", status.State)
}
```

---

## 9. Migration from Wails

### Coexistence Strategy

The API server and Wails GUI can coexist:

1. **Wails continues to work** - No breaking changes
2. **API server is optional** - New `--api-server` flag
3. **Shared codebase** - Both use same VM/parser/debugger

### Migration Path

**Phase 1:** API server alongside Wails
- Users can choose GUI (Wails) or native app (Swift)
- Both maintained in parallel

**Phase 2:** Swift becomes primary macOS experience
- Wails remains for cross-platform web UI
- Windows/Linux use Wails until native clients ready

**Phase 3:** Native clients on all platforms
- Swift (macOS)
- WPF/Avalonia (Windows/Linux)
- Wails deprecated or becomes "lite" web UI

---

## 10. Performance Considerations

### Latency Analysis

**REST API Latency:**
- JSON serialization: < 1ms (small payloads)
- HTTP overhead: 1-2ms (localhost)
- Total: < 5ms per request

**WebSocket Latency:**
- Event serialization: < 1ms
- WebSocket send: < 1ms
- Total: < 2ms for real-time updates

**Comparison:**
- Wails (in-process): ~0.1ms
- API (localhost): ~2-5ms
- **Impact:** Negligible for human interaction (< 60fps requirement = 16ms)

### Optimization Strategies

1. **Batch Updates**
   - Send register updates at 60Hz max (not per instruction)
   - Debounce output streaming

2. **Incremental State**
   - Send only changed registers
   - Delta compression for large memory regions

3. **Connection Pooling**
   - Reuse HTTP connections
   - Keep WebSocket alive

4. **Efficient Serialization**
   - Use JSON for simplicity
   - Consider MessagePack/Protocol Buffers if needed

---

## 11. Security Considerations

### Localhost Binding

- Bind to `127.0.0.1` only (not `0.0.0.0`)
- Prevent network access by default
- Optional `--bind` flag for remote access (with warning)

### Authentication

- Not needed for local-only use
- If network access: add API key or OAuth

### Sandboxing

- Swift app: macOS sandbox with file access entitlements
- Go backend: existing filesystem security (`-fsroot`)

### Input Validation

- Validate all API inputs
- Limit request sizes (prevent DoS)
- Sanitize file paths

---

## 12. Documentation Plan

### API Documentation

- OpenAPI/Swagger specification
- Interactive API explorer (Swagger UI)
- Code examples (curl, Swift, C#)

### User Documentation

- "Getting Started" guide for Swift app
- Feature comparison (Wails vs Native)
- Troubleshooting guide

### Developer Documentation

- Architecture overview
- Service layer guide
- Adding new endpoints
- Client implementation guide

---

## 13. Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| API latency too high | Medium | Low | Profile early, optimize if needed |
| Complexity of managing two processes | Medium | Medium | Robust launcher, auto-recovery |
| Swift development expertise | High | Medium | Start small, iterate, leverage SwiftUI |
| Cross-platform API compatibility | Medium | Low | Test on all platforms early |
| Feature creep | High | High | Stick to stage plan, prioritize MVP |
| Breaking changes in Wails | Low | Low | API decouples from Wails |

---

## 14. Success Metrics

### Technical Metrics
- API response time < 10ms (p95)
- WebSocket latency < 5ms (p95)
- Memory usage < 100MB for Swift app
- Zero crashes in 1-hour stress test
- 100% API test coverage

### User Experience Metrics
- App launch time < 2 seconds
- UI feels "snappy" (< 16ms frame time)
- Native macOS look and feel
- Feature parity with Wails

### Development Metrics
- All stages completed on schedule
- All tests passing
- Documentation complete
- Cross-platform clients feasible

---

## 15. Future Enhancements

### Phase 2 Features (Post-Launch)

1. **Remote Debugging**
   - Connect to emulator on another machine
   - Collaborative debugging sessions

2. **Plugin System**
   - External tools via API
   - Custom debugger extensions

3. **Cloud Sync**
   - Sync programs across devices
   - Cloud-based examples library

4. **Performance Profiling**
   - Flame graphs
   - Hotspot analysis
   - Bottleneck detection

5. **Mobile Clients**
   - iOS/iPad app (Swift)
   - Android app (Kotlin)

6. **Web-Based Client**
   - Reuse API for web UI
   - No Wails dependency
   - Pure HTML/JS/CSS

---

## 16. Conclusion

Building a Swift native macOS GUI backed by a Go API server is **highly practical and recommended**. This architecture provides:

✅ **Native Performance:** SwiftUI delivers 60fps responsiveness
✅ **Cross-Platform Ready:** API enables .NET, web, mobile clients
✅ **Clean Architecture:** Clear separation of concerns
✅ **Maintainability:** Service layer encapsulates business logic
✅ **Flexibility:** Multiple clients, headless mode, automation
✅ **Future-Proof:** Extensible for plugins, remote access, cloud features

The **8-week staged implementation plan** is achievable with one developer, with the first usable Swift app available by week 4. The API server provides a foundation for native clients on all platforms, far exceeding the capabilities of the current Wails implementation.

**Recommendation:** Proceed with Stage 1 (Backend API Foundation) immediately.

---

## Appendix A: Technology Stack Summary

### Backend (Go)
- **HTTP Server:** Gin or `net/http`
- **WebSocket:** `gorilla/websocket`
- **JSON:** Standard library `encoding/json`
- **Testing:** Standard library `testing`
- **Existing:** All current packages (vm, parser, debugger, etc.)

### Frontend (Swift/macOS)
- **UI:** SwiftUI
- **Networking:** URLSession
- **State Management:** Combine
- **Persistence:** UserDefaults / FileManager
- **Testing:** XCTest

### Frontend (.NET/Windows)
- **UI:** WPF or Avalonia
- **Networking:** HttpClient, ClientWebSocket
- **Serialization:** System.Text.Json
- **Testing:** xUnit

### Development Tools
- **API Testing:** Postman, curl
- **API Docs:** Swagger/OpenAPI
- **Version Control:** Git
- **CI/CD:** GitHub Actions

---

## Appendix B: Example API Requests

### Create Session
```bash
curl -X POST http://localhost:8080/api/v1/session \
  -H "Content-Type: application/json" \
  -d '{"memorySize": 1048576}'
```

Response:
```json
{
  "sessionId": "abc123",
  "createdAt": "2025-01-01T12:00:00Z"
}
```

### Load Program
```bash
curl -X POST http://localhost:8080/api/v1/session/abc123/load \
  -H "Content-Type: application/json" \
  -d '{"source": "MOVE R0, #42\nSWI #0"}'
```

### Run Program
```bash
curl -X POST http://localhost:8080/api/v1/session/abc123/run
```

### Get Registers
```bash
curl http://localhost:8080/api/v1/session/abc123/registers
```

Response:
```json
{
  "r0": 42,
  "r1": 0,
  "r2": 0,
  "r3": 0,
  "r4": 0,
  "r5": 0,
  "r6": 0,
  "r7": 0,
  "r8": 0,
  "r9": 0,
  "r10": 0,
  "r11": 0,
  "r12": 0,
  "sp": 327680,
  "lr": 0,
  "pc": 32776,
  "cpsr": {
    "n": false,
    "z": false,
    "c": false,
    "v": false
  }
}
```

---

## Appendix C: File Checklist

### New Files to Create

**Go Backend:**
- [ ] `service/service.go`
- [ ] `service/emulator_service.go`
- [ ] `service/debugger_service.go`
- [ ] `service/file_service.go`
- [ ] `service/config_service.go`
- [ ] `service/trace_service.go`
- [ ] `service/session_manager.go`
- [ ] `service/service_test.go`
- [ ] `api/server.go`
- [ ] `api/handlers.go`
- [ ] `api/models.go`
- [ ] `api/websocket.go`
- [ ] `api/broadcaster.go`
- [ ] `api/middleware.go`
- [ ] `api/api_test.go`
- [ ] `cmd/api-server/main.go`

**Swift macOS App:**
- [ ] `ARMEmulator/ARMEmulatorApp.swift`
- [ ] `ARMEmulator/Models/EmulatorSession.swift`
- [ ] `ARMEmulator/Models/Register.swift`
- [ ] `ARMEmulator/Models/MemoryRegion.swift`
- [ ] `ARMEmulator/Models/ProgramState.swift`
- [ ] `ARMEmulator/Services/APIClient.swift`
- [ ] `ARMEmulator/Services/WebSocketClient.swift`
- [ ] `ARMEmulator/Services/BackendLauncher.swift`
- [ ] `ARMEmulator/Services/FileManager.swift`
- [ ] `ARMEmulator/Views/MainView.swift`
- [ ] `ARMEmulator/Views/EditorView.swift`
- [ ] `ARMEmulator/Views/RegistersView.swift`
- [ ] `ARMEmulator/Views/MemoryView.swift`
- [ ] `ARMEmulator/Views/ConsoleView.swift`
- [ ] `ARMEmulator/Views/ToolbarView.swift`
- [ ] `ARMEmulator/ViewModels/EmulatorViewModel.swift`
- [ ] `ARMEmulator/ViewModels/EditorViewModel.swift`

**Documentation:**
- [ ] `docs/API.md` - API reference
- [ ] `docs/SWIFT_DEVELOPMENT.md` - Swift app guide
- [ ] `docs/ARCHITECTURE.md` - System architecture
- [ ] Update `README.md` with Swift app info

**Configuration:**
- [ ] `.github/workflows/swift-build.yml` - CI for Swift app
- [ ] `ARMEmulator.xcodeproj` - Xcode project
- [ ] `openapi.yaml` - API specification

---

*End of Planning Document*
