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

    // Callback for scrolling editor to current PC
    var scrollToCurrentPC: (() -> Void)?

    // Memory state
    @Published var memoryData: [UInt8] = []
    @Published var memoryAddress: UInt32 = 0x8000
    @Published var lastMemoryWrite: UInt32?

    // Disassembly state
    @Published var disassembly: [DisassemblyInstruction] = []

    // Source map: address -> source line (for display)
    @Published var sourceMap: [UInt32: String] = [:]
    // Valid breakpoint lines (1-based line numbers that can have breakpoints)
    @Published var validBreakpointLines: Set<Int> = []
    // Line number to address mapping for breakpoint setting
    @Published var lineToAddress: [Int: UInt32] = [:]
    // Address to line number mapping for breakpoint display
    @Published var addressToLine: [UInt32: Int] = [:]

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

            // Fetch source map after loading program
            DebugLog.log("Fetching source map...", category: "ViewModel")
            let sourceMapEntries = try await apiClient.getSourceMap(sessionID: sessionID)

            // Build address->line map for display
            sourceMap = Dictionary(uniqueKeysWithValues: sourceMapEntries.map { ($0.address, $0.line) })

            // Build valid breakpoint lines and bidirectional line<->address mappings
            validBreakpointLines = Set(sourceMapEntries.map { $0.lineNumber })
            lineToAddress = Dictionary(uniqueKeysWithValues: sourceMapEntries.map { ($0.lineNumber, $0.address) })
            addressToLine = Dictionary(uniqueKeysWithValues: sourceMapEntries.map { ($0.address, $0.lineNumber) })

            DebugLog.log(
                "Loaded source map with \(sourceMap.count) entries, \(validBreakpointLines.count) valid breakpoint lines",
                category: "ViewModel"
            )

            // Force UI update now that source map is loaded
            // We need to trigger the onChange even though PC value hasn't changed
            // by briefly changing it and then setting it back
            let savedPC = currentPC
            currentPC = 0xFFFF_FFFF // Temporary different value
            currentPC = savedPC // Restore actual PC, triggering onChange with valid mapping

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
            try await refreshState()
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
            // Check if program exited normally (not an error, just completion)
            let errorDesc = error.localizedDescription
            if errorDesc.contains("program exited with code") {
                // Program terminated - refresh state to show halted status
                try? await refreshState()
                errorMessage = nil
            } else {
                errorMessage = "Failed to step: \(errorDesc)"
            }
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
            // Check if program exited normally (not an error, just completion)
            let errorDesc = error.localizedDescription
            if errorDesc.contains("program exited with code") {
                // Program terminated - refresh state to show halted status
                try? await refreshState()
                errorMessage = nil
            } else {
                errorMessage = "Failed to step over: \(errorDesc)"
            }
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
            // Check if program exited normally (not an error, just completion)
            let errorDesc = error.localizedDescription
            if errorDesc.contains("program exited with code") {
                // Program terminated - refresh state to show halted status
                try? await refreshState()
                errorMessage = nil
            } else {
                errorMessage = "Failed to step out: \(errorDesc)"
            }
        }
    }

    func reset() async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            try await apiClient.restart(sessionID: sessionID)
            consoleOutput = ""
            try await refreshState()
            errorMessage = nil
        } catch {
            errorMessage = "Failed to restart: \(error.localizedDescription)"
        }
    }

    private func refreshState() async throws {
        guard let sessionID = sessionID else { return }

        let newRegisters = try await apiClient.getRegisters(sessionID: sessionID)
        updateRegisters(newRegisters)

        let vmStatus = try await apiClient.getStatus(sessionID: sessionID)
        status = vmStatus.vmState

        // Track last memory write - clear when no write occurs
        if let hasWrite = vmStatus.hasWrite, hasWrite, let writeAddr = vmStatus.writeAddr {
            DebugLog.log("Memory write detected at 0x\(String(format: "%08X", writeAddr))", category: "ViewModel")
            lastMemoryWrite = writeAddr
        } else {
            // Clear the write flag when no write happened
            lastMemoryWrite = nil
        }
    }

    private func updateRegisters(_ newRegisters: RegisterState) {
        // Track changes
        var changed = Set<String>()

        if let prev = previousRegisters {
            changed = detectRegisterChanges(previous: prev, new: newRegisters)
        }

        previousRegisters = registers
        changedRegisters = changed
        registers = newRegisters
        currentPC = newRegisters.pc
    }

    private func detectRegisterChanges(previous: RegisterState, new: RegisterState) -> Set<String> {
        var changed = Set<String>()

        // Compare general-purpose registers
        struct RegisterComparison {
            let name: String
            let prev: UInt32
            let new: UInt32
        }

        let registers: [RegisterComparison] = [
            RegisterComparison(name: "R0", prev: previous.r0, new: new.r0),
            RegisterComparison(name: "R1", prev: previous.r1, new: new.r1),
            RegisterComparison(name: "R2", prev: previous.r2, new: new.r2),
            RegisterComparison(name: "R3", prev: previous.r3, new: new.r3),
            RegisterComparison(name: "R4", prev: previous.r4, new: new.r4),
            RegisterComparison(name: "R5", prev: previous.r5, new: new.r5),
            RegisterComparison(name: "R6", prev: previous.r6, new: new.r6),
            RegisterComparison(name: "R7", prev: previous.r7, new: new.r7),
            RegisterComparison(name: "R8", prev: previous.r8, new: new.r8),
            RegisterComparison(name: "R9", prev: previous.r9, new: new.r9),
            RegisterComparison(name: "R10", prev: previous.r10, new: new.r10),
            RegisterComparison(name: "R11", prev: previous.r11, new: new.r11),
            RegisterComparison(name: "R12", prev: previous.r12, new: new.r12),
            RegisterComparison(name: "SP", prev: previous.sp, new: new.sp),
            RegisterComparison(name: "LR", prev: previous.lr, new: new.lr),
            RegisterComparison(name: "PC", prev: previous.pc, new: new.pc),
        ]

        for reg in registers where reg.prev != reg.new {
            changed.insert(reg.name)
        }

        if previous.cpsr != new.cpsr {
            changed.insert("CPSR")
        }

        return changed
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
}

// MARK: - Input Operations Extension

extension EmulatorViewModel {
    func sendInput(_ input: String) async {
        DebugLog.log("sendInput() called with input: \(input.prefix(20))...", category: "ViewModel")
        DebugLog.log("Current status: \(status)", category: "ViewModel")

        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        // Capture current status before sending input
        // If VM is waiting for input, the input will unblock an existing step() call
        // If VM is NOT waiting, the backend buffers input and we need to step() to consume it
        let wasWaitingForInput = (status == .waitingForInput)
        DebugLog.log("Was waiting for input: \(wasWaitingForInput)", category: "ViewModel")

        do {
            DebugLog.log("Sending stdin to backend...", category: "ViewModel")
            try await apiClient.sendStdin(sessionID: sessionID, data: input)
            DebugLog.success("Stdin sent successfully", category: "ViewModel")
            errorMessage = nil

            if wasWaitingForInput {
                // VM was waiting for input - the step() that triggered the input request
                // is still in progress and will complete now that we've provided input.
                // DO NOT call step() again or we'll execute an extra instruction!
                DebugLog.log(
                    "VM was waiting - input will unblock existing step, skipping auto-step",
                    category: "ViewModel"
                )
                try await refreshState()
            } else {
                // VM was not waiting - the backend buffered the input for later.
                // Call step() to consume the buffered input.
                DebugLog.log(
                    "VM was not waiting - stepping to consume buffered input...",
                    category: "ViewModel"
                )
                try await apiClient.step(sessionID: sessionID)
                try await refreshState()
                DebugLog.success("Step after input completed", category: "ViewModel")
            }
        } catch {
            DebugLog.error("sendInput() failed: \(error.localizedDescription)", category: "ViewModel")
            errorMessage = "Failed to send input: \(error.localizedDescription)"
        }
    }
}

// MARK: - Event Handling Extension

extension EmulatorViewModel {
    func handleEvent(_ event: EmulatorEvent) {
        guard event.sessionId == sessionID else {
            DebugLog.warning("Ignoring event for different session", category: "ViewModel")
            return
        }

        DebugLog.log("WebSocket event received: \(event.type)", category: "ViewModel")

        switch event.type {
        case "state":
            if let data = event.data, case let .state(stateUpdate) = data {
                DebugLog.log(
                    "State update - status: \(stateUpdate.status), PC: \(stateUpdate.pc.map { String(format: "0x%08X", $0) } ?? "nil")",
                    category: "ViewModel"
                )

                // Update registers if provided (full state update)
                if let registers = stateUpdate.registers {
                    updateRegisters(registers)
                }

                // Always update status (even for status-only updates like waiting_for_input)
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

    func handleExecutionEvent(_ event: ExecutionEvent) {
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
}

// MARK: - Debug Operations Extension

extension EmulatorViewModel {
    func toggleBreakpoint(at address: UInt32) async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            print("toggleBreakpoint: No active session")
            return
        }

        print(
            "toggleBreakpoint: sessionID=\(sessionID), address=0x\(String(format: "%X", address)), current breakpoints=\(breakpoints)"
        )

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
            print("toggleBreakpoint error: \(error)")
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
}

// MARK: - Memory Operations Extension

extension EmulatorViewModel {
    func loadMemory(at address: UInt32, length: Int) async {
        guard let sessionID = sessionID else {
            DebugLog.error("loadMemory: No session ID", category: "ViewModel")
            return
        }

        DebugLog.log(
            "loadMemory: address=0x\(String(format: "%08X", address)), length=\(length)",
            category: "ViewModel"
        )

        do {
            memoryData = try await apiClient.getMemory(sessionID: sessionID, address: address, length: length)
            memoryAddress = address
            DebugLog.log(
                "loadMemory: Got \(memoryData.count) bytes at 0x\(String(format: "%08X", memoryAddress))",
                category: "ViewModel"
            )
        } catch {
            DebugLog.error("loadMemory failed: \(error.localizedDescription)", category: "ViewModel")
            memoryData = []
        }
    }

    func fetchMemory(at address: UInt32, length: Int) async throws -> [UInt8] {
        guard let sessionID = sessionID else {
            throw NSError(
                domain: "EmulatorViewModel",
                code: 1,
                userInfo: [NSLocalizedDescriptionKey: "No active session"]
            )
        }

        return try await apiClient.getMemory(sessionID: sessionID, address: address, length: length)
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
