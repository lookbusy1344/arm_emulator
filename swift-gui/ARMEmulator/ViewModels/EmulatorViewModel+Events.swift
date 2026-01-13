import Foundation

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
