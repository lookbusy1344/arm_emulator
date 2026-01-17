import Combine
import Foundation
import SwiftUI

@MainActor
class EmulatorViewModel: ObservableObject {
    @Published var registers: RegisterState = .empty
    @Published var previousRegisters: RegisterState?
    @Published var changedRegisters: Set<String> = []

    // Highlight tracking with UUIDs for animation
    @Published var registerHighlights: [String: UUID] = [:]
    @Published var memoryHighlights: [UInt32: UUID] = [:]

    // Task tracking for cleanup
    private var highlightTasks: [String: Task<Void, Never>] = [:]
    private var memoryHighlightTasks: [UInt32: Task<Void, Never>] = [:]

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
    @Published var lastMemoryWriteSize: UInt32 = 4 // Size in bytes (1, 2, or 4)

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

    let apiClient: any APIClientProtocol
    private let wsClient: any WebSocketClientProtocol
    var sessionID: String?
    private var cancellables = Set<AnyCancellable>()
    private var isInitializing = false

    func highlightRegister(_ name: String) {
        // Cancel existing fade task for this register
        highlightTasks[name]?.cancel()

        // Add new highlight (triggers animation to green)
        registerHighlights[name] = UUID()

        // Schedule removal after 1.5 seconds
        highlightTasks[name] = Task { @MainActor in
            try? await Task.sleep(nanoseconds: 1_500_000_000)
            guard !Task.isCancelled else { return }
            registerHighlights[name] = nil
            highlightTasks[name] = nil
        }
    }

    func highlightMemoryAddress(_ address: UInt32, size: UInt32) {
        // Highlight each byte in the write
        for offset in 0 ..< size {
            let addr = address + offset

            // Cancel existing fade task
            memoryHighlightTasks[addr]?.cancel()

            // Add new highlight
            memoryHighlights[addr] = UUID()

            // Schedule removal after 1.5 seconds
            memoryHighlightTasks[addr] = Task { @MainActor in
                try? await Task.sleep(nanoseconds: 1_500_000_000)
                guard !Task.isCancelled else { return }
                memoryHighlights[addr] = nil
                memoryHighlightTasks[addr] = nil
            }
        }
    }

    init(apiClient: any APIClientProtocol = APIClient(), wsClient: any WebSocketClientProtocol = WebSocketClient()) {
        self.apiClient = apiClient
        self.wsClient = wsClient

        wsClient.events
            .receive(on: DispatchQueue.main)
            .sink { [weak self] event in
                self?.handleEvent(event)
            }
            .store(in: &cancellables)
    }

    func cancelAllHighlights() {
        // Cancel all pending highlight tasks
        for task in highlightTasks.values {
            task.cancel()
        }
        for task in memoryHighlightTasks.values {
            task.cancel()
        }

        // Clear all highlights
        highlightTasks.removeAll()
        memoryHighlightTasks.removeAll()
        registerHighlights.removeAll()
        memoryHighlights.removeAll()
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

        // Clear highlights when loading new program
        cancelAllHighlights()

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

    func refreshState() async throws {
        guard let sessionID = sessionID else { return }

        let newRegisters = try await apiClient.getRegisters(sessionID: sessionID)
        updateRegisters(newRegisters)

        let vmStatus = try await apiClient.getStatus(sessionID: sessionID)
        status = vmStatus.vmState

        // Track last memory write - value persists until next write or explicit clear in step()
        if let hasWrite = vmStatus.hasWrite, hasWrite, let writeAddr = vmStatus.writeAddr {
            let writeSize = vmStatus.writeSize ?? 4
            lastMemoryWrite = writeAddr
            lastMemoryWriteSize = writeSize
        }
        // Don't clear here - let the value persist so onChange can fire
        // It will be cleared at the start of the next step() operation
    }

    func updateRegisters(_ newRegisters: RegisterState) {
        // Track changes
        var changed = Set<String>()

        if let prev = previousRegisters {
            changed = detectRegisterChanges(previous: prev, new: newRegisters)
        }

        // Highlight each changed register (independent timers)
        for registerName in changed {
            highlightRegister(registerName)
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
