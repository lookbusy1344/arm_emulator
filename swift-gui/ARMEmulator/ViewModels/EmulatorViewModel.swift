import Combine
import Foundation
import SwiftUI

@MainActor
class EmulatorViewModel: ObservableObject {
    @Published var registers: RegisterState = .empty
    @Published var consoleOutput = ""
    @Published var status: VMState = .idle
    @Published var breakpoints: Set<UInt32> = []
    @Published var watchpoints: [Watchpoint] = []
    @Published var sourceCode = ""
    @Published var currentPC: UInt32 = 0
    @Published var errorMessage: String?
    @Published var isConnected = false

    // Memory state
    @Published var memoryData: [UInt8] = []
    @Published var memoryAddress: UInt32 = 0x8000

    // Disassembly state
    @Published var disassembly: [DisassembledInstruction] = []

    let apiClient: APIClient
    private let wsClient: WebSocketClient
    var sessionID: String?
    private var cancellables = Set<AnyCancellable>()
    private var isInitializing = false

    init(apiClient: APIClient = APIClient(), wsClient: WebSocketClient = WebSocketClient()) {
        self.apiClient = apiClient
        self.wsClient = wsClient

        wsClient.events
            .receive(on: DispatchQueue.main)
            .sink { [weak self] event in
                self?.handleEvent(event)
            }
            .store(in: &cancellables)
    }

    func initialize() async {
        // Prevent concurrent initialization
        guard !isInitializing, !isConnected else { return }

        isInitializing = true
        defer { isInitializing = false }

        do {
            sessionID = try await apiClient.createSession()
            wsClient.connect(sessionID: sessionID!)
            isConnected = true
            errorMessage = nil
        } catch {
            errorMessage = "Failed to initialize: \(error.localizedDescription)"
            isConnected = false
        }
    }

    func loadProgram(source: String) async {
        DebugLog.log("loadProgram() called", category: "ViewModel")
        DebugLog.log("Source length: \(source.count) chars", category: "ViewModel")

        guard let sessionID = sessionID else {
            DebugLog.error("No active session for loadProgram", category: "ViewModel")
            errorMessage = "No active session"
            return
        }

        DebugLog.log("SessionID: \(sessionID)", category: "ViewModel")

        do {
            DebugLog.log("Calling apiClient.loadProgram()...", category: "ViewModel")
            let response = try await apiClient.loadProgram(sessionID: sessionID, source: source)

            DebugLog.log("Load response - success: \(response.success)", category: "ViewModel")

            // Check if load was successful
            if !response.success {
                let errors = response.errors?.joined(separator: "\n") ?? "Unknown error"
                DebugLog.error("Load failed with errors: \(errors)", category: "ViewModel")
                errorMessage = "Failed to load program:\n\(errors)"
                return
            }

            if let symbols = response.symbols {
                DebugLog.log("Loaded \(symbols.count) symbols", category: "ViewModel")
                for (name, addr) in symbols.prefix(5) {
                    DebugLog.log("  Symbol: \(name) -> 0x\(String(format: "%08X", addr))", category: "ViewModel")
                }
            }

            sourceCode = source
            errorMessage = nil

            DebugLog.log("Refreshing state...", category: "ViewModel")
            try await refreshState()
            DebugLog.success("Program loaded successfully, PC: 0x\(String(format: "%08X", currentPC))", category: "ViewModel")
        } catch {
            DebugLog.error("loadProgram() failed: \(error.localizedDescription)", category: "ViewModel")
            errorMessage = "Failed to load program: \(error.localizedDescription)"
        }
    }

    func run() async {
        DebugLog.log("run() called", category: "ViewModel")

        guard let sessionID = sessionID else {
            DebugLog.error("No active session", category: "ViewModel")
            errorMessage = "No active session"
            return
        }

        DebugLog.log("SessionID: \(sessionID)", category: "ViewModel")
        DebugLog.log("Current status: \(status)", category: "ViewModel")
        DebugLog.log("Current PC: 0x\(String(format: "%08X", currentPC))", category: "ViewModel")

        do {
            DebugLog.log("Calling apiClient.run()...", category: "ViewModel")
            try await apiClient.run(sessionID: sessionID)
            DebugLog.success("apiClient.run() succeeded", category: "ViewModel")
            errorMessage = nil
        } catch {
            DebugLog.error("apiClient.run() failed: \(error.localizedDescription)", category: "ViewModel")
            errorMessage = "Failed to run: \(error.localizedDescription)"
        }

        DebugLog.log("run() completed", category: "ViewModel")
    }

    func stop() async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            try await apiClient.stop(sessionID: sessionID)
            errorMessage = nil
        } catch {
            errorMessage = "Failed to stop: \(error.localizedDescription)"
        }
    }

    func step() async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            try await apiClient.step(sessionID: sessionID)
            try await refreshState()
            errorMessage = nil
        } catch {
            errorMessage = "Failed to step: \(error.localizedDescription)"
        }
    }

    func stepOver() async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            try await apiClient.stepOver(sessionID: sessionID)
            try await refreshState()
            errorMessage = nil
        } catch {
            errorMessage = "Failed to step over: \(error.localizedDescription)"
        }
    }

    func stepOut() async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            try await apiClient.stepOut(sessionID: sessionID)
            try await refreshState()
            errorMessage = nil
        } catch {
            errorMessage = "Failed to step out: \(error.localizedDescription)"
        }
    }

    func reset() async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            try await apiClient.reset(sessionID: sessionID)
            consoleOutput = ""
            try await refreshState()
            errorMessage = nil
        } catch {
            errorMessage = "Failed to reset: \(error.localizedDescription)"
        }
    }

    func sendInput(_ input: String) async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            try await apiClient.sendStdin(sessionID: sessionID, data: input)
            errorMessage = nil
        } catch {
            errorMessage = "Failed to send input: \(error.localizedDescription)"
        }
    }

    func toggleBreakpoint(at address: UInt32) async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            if breakpoints.contains(address) {
                try await apiClient.removeBreakpoint(sessionID: sessionID, address: address)
                breakpoints.remove(address)
            } else {
                try await apiClient.addBreakpoint(sessionID: sessionID, address: address)
                breakpoints.insert(address)
            }
            errorMessage = nil
        } catch {
            errorMessage = "Failed to toggle breakpoint: \(error.localizedDescription)"
        }
    }

    func addWatchpoint(at address: UInt32, type: String) async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            let watchpoint = try await apiClient.addWatchpoint(sessionID: sessionID, address: address, type: type)
            watchpoints.append(watchpoint)
            errorMessage = nil
        } catch {
            errorMessage = "Failed to add watchpoint: \(error.localizedDescription)"
        }
    }

    func removeWatchpoint(id: Int) async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            try await apiClient.removeWatchpoint(sessionID: sessionID, watchpointID: id)
            watchpoints.removeAll { $0.id == id }
            errorMessage = nil
        } catch {
            errorMessage = "Failed to remove watchpoint: \(error.localizedDescription)"
        }
    }

    func refreshWatchpoints() async {
        guard let sessionID = sessionID else { return }

        do {
            watchpoints = try await apiClient.getWatchpoints(sessionID: sessionID)
        } catch {
            errorMessage = "Failed to refresh watchpoints: \(error.localizedDescription)"
        }
    }

    private func refreshState() async throws {
        guard let sessionID = sessionID else { return }

        registers = try await apiClient.getRegisters(sessionID: sessionID)
        currentPC = registers.pc

        let vmStatus = try await apiClient.getStatus(sessionID: sessionID)
        status = vmStatus.vmState
    }

    private func handleEvent(_ event: EmulatorEvent) {
        guard event.sessionId == sessionID else {
            DebugLog.warning("Ignoring event for different session", category: "ViewModel")
            return
        }

        DebugLog.log("WebSocket event received: \(event.type)", category: "ViewModel")

        switch event.type {
        case "state":
            if let data = event.data, case let .state(stateUpdate) = data {
                DebugLog.log("State update - PC: 0x\(String(format: "%08X", stateUpdate.pc)), status: \(stateUpdate.status)", category: "ViewModel")
                registers = stateUpdate.registers
                currentPC = stateUpdate.pc
                status = VMState(rawValue: stateUpdate.status) ?? .idle
            }
        case "output":
            if let data = event.data, case let .output(outputUpdate) = data {
                DebugLog.log("Console output: \(outputUpdate.content.prefix(50))...", category: "ViewModel")
                consoleOutput += outputUpdate.content
            }
        case "event":
            if let data = event.data, case let .event(execEvent) = data {
                DebugLog.log("Execution event: \(execEvent.event)", category: "ViewModel")
                handleExecutionEvent(execEvent)
            }
        default:
            DebugLog.warning("Unknown event type: \(event.type)", category: "ViewModel")
            break
        }
    }

    private func handleExecutionEvent(_ event: ExecutionEvent) {
        switch event.event {
        case "breakpoint_hit":
            status = .paused
            if let address = event.address {
                currentPC = address
            }
        case "error":
            status = .error
            errorMessage = event.message ?? "Unknown error"
        case "halted":
            status = .halted
        default:
            break
        }
    }

    func cleanup() {
        wsClient.disconnect()

        if let sessionID = sessionID {
            Task {
                try? await apiClient.destroySession(sessionID: sessionID)
            }
        }

        isConnected = false
        isInitializing = false
        sessionID = nil
    }

    // MARK: - Memory Operations

    func loadMemory(at address: UInt32, length: Int) async {
        guard let sessionID = sessionID else { return }

        do {
            memoryData = try await apiClient.getMemory(sessionID: sessionID, address: address, length: length)
            memoryAddress = address
        } catch {
            // Silently fail - memory loading errors are benign (e.g., no program loaded)
            // Views will just show empty data
            memoryData = []
        }
    }

    func loadDisassembly(around address: UInt32, count: Int) async {
        guard let sessionID = sessionID else { return }

        do {
            disassembly = try await apiClient.getDisassembly(sessionID: sessionID, address: address, count: count)
        } catch {
            // Silently fail - disassembly errors are benign (e.g., no program loaded)
            // View will just show empty data
            disassembly = []
        }
    }
}
