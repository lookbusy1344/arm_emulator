import Combine
import Foundation
import SwiftUI

@MainActor
class EmulatorViewModel: ObservableObject {
    @Published var registers: RegisterState = .empty
    @Published var previousRegisters: RegisterState?
    @Published var changedRegisters: Set<String> = []
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
    @Published var lastMemoryWrite: UInt32?

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
            DebugLog.success(
                "Program loaded successfully, PC: 0x\(String(format: "%08X", currentPC))",
                category: "ViewModel"
            )
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

        let newRegisters = try await apiClient.getRegisters(sessionID: sessionID)
        updateRegisters(newRegisters)

        let vmStatus = try await apiClient.getStatus(sessionID: sessionID)
        status = vmStatus.vmState

        // Track last memory write
        if let hasWrite = vmStatus.hasWrite, hasWrite, let writeAddr = vmStatus.writeAddr {
            lastMemoryWrite = writeAddr
        }
    }

    private func updateRegisters(_ newRegisters: RegisterState) {
        // Track changes
        var changed = Set<String>()

        if let prev = previousRegisters {
            if prev.r0 != newRegisters.r0 { changed.insert("R0") }
            if prev.r1 != newRegisters.r1 { changed.insert("R1") }
            if prev.r2 != newRegisters.r2 { changed.insert("R2") }
            if prev.r3 != newRegisters.r3 { changed.insert("R3") }
            if prev.r4 != newRegisters.r4 { changed.insert("R4") }
            if prev.r5 != newRegisters.r5 { changed.insert("R5") }
            if prev.r6 != newRegisters.r6 { changed.insert("R6") }
            if prev.r7 != newRegisters.r7 { changed.insert("R7") }
            if prev.r8 != newRegisters.r8 { changed.insert("R8") }
            if prev.r9 != newRegisters.r9 { changed.insert("R9") }
            if prev.r10 != newRegisters.r10 { changed.insert("R10") }
            if prev.r11 != newRegisters.r11 { changed.insert("R11") }
            if prev.r12 != newRegisters.r12 { changed.insert("R12") }
            if prev.sp != newRegisters.sp { changed.insert("SP") }
            if prev.lr != newRegisters.lr { changed.insert("LR") }
            if prev.pc != newRegisters.pc { changed.insert("PC") }
            if prev.cpsr != newRegisters.cpsr { changed.insert("CPSR") }
        }

        previousRegisters = registers
        changedRegisters = changed
        registers = newRegisters
        currentPC = newRegisters.pc
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
                DebugLog.log(
                    "State update - PC: 0x\(String(format: "%08X", stateUpdate.pc)), status: \(stateUpdate.status)",
                    category: "ViewModel"
                )
                updateRegisters(stateUpdate.registers)
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
