import Combine
import Foundation
import SwiftUI

@MainActor
class EmulatorViewModel: ObservableObject {
    @Published var registers: RegisterState = .empty
    @Published var consoleOutput = ""
    @Published var status: VMState = .idle
    @Published var breakpoints: Set<UInt32> = []
    @Published var sourceCode = ""
    @Published var currentPC: UInt32 = 0
    @Published var errorMessage: String?
    @Published var isConnected = false

    // Memory state
    @Published var memoryData: [UInt8] = []
    @Published var memoryAddress: UInt32 = 0x8000

    // Disassembly state
    @Published var disassembly: [DisassembledInstruction] = []

    private let apiClient: APIClient
    private let wsClient: WebSocketClient
    private var sessionID: String?
    private var cancellables = Set<AnyCancellable>()

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
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            try await apiClient.loadProgram(sessionID: sessionID, source: source)
            sourceCode = source
            errorMessage = nil

            try await refreshState()
        } catch {
            errorMessage = "Failed to load program: \(error.localizedDescription)"
        }
    }

    func run() async {
        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            try await apiClient.run(sessionID: sessionID)
            errorMessage = nil
        } catch {
            errorMessage = "Failed to run: \(error.localizedDescription)"
        }
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

    private func refreshState() async throws {
        guard let sessionID = sessionID else { return }

        registers = try await apiClient.getRegisters(sessionID: sessionID)
        currentPC = registers.pc

        let vmStatus = try await apiClient.getStatus(sessionID: sessionID)
        status = vmStatus.vmState
    }

    private func handleEvent(_ event: EmulatorEvent) {
        guard event.sessionId == sessionID else { return }

        switch event.type {
        case "state":
            if let data = event.data, case let .state(stateUpdate) = data {
                registers = stateUpdate.registers
                currentPC = stateUpdate.pc
                status = VMState(rawValue: stateUpdate.status) ?? .idle
            }
        case "output":
            if let data = event.data, case let .output(outputUpdate) = data {
                consoleOutput += outputUpdate.content
            }
        case "event":
            if let data = event.data, case let .event(execEvent) = data {
                handleExecutionEvent(execEvent)
            }
        default:
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
    }

    // MARK: - Memory Operations

    func loadMemory(at address: UInt32, length: Int) async {
        guard let sessionID = sessionID else { return }

        do {
            memoryData = try await apiClient.getMemory(sessionID: sessionID, address: address, length: length)
            memoryAddress = address
        } catch {
            errorMessage = "Failed to load memory: \(error.localizedDescription)"
        }
    }

    func loadDisassembly(around address: UInt32, count: Int) async {
        guard let sessionID = sessionID else { return }

        do {
            disassembly = try await apiClient.getDisassembly(sessionID: sessionID, address: address, count: count)
        } catch {
            errorMessage = "Failed to load disassembly: \(error.localizedDescription)"
        }
    }
}
