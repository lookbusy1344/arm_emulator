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
                let newStatus = VMState(rawValue: stateUpdate.status) ?? .idle

                // Prevent stale WebSocket events from overriding explicit stop
                // Check this BEFORE logging or updating anything
                if status == .halted, newStatus == .breakpoint || newStatus == .running {
                    DebugLog.warning(
                        "Ignoring stale WebSocket state transition from halted to \(newStatus) at PC \(stateUpdate.pc.map { String(format: "0x%08X", $0) } ?? "nil")",
                        category: "ViewModel"
                    )
                    return // Exit early - don't update registers or PC
                }

                DebugLog.log(
                    "State update - status: \(stateUpdate.status), PC: \(stateUpdate.pc.map { String(format: "0x%08X", $0) } ?? "nil")",
                    category: "ViewModel"
                )

                // Update registers if provided (full state update)
                if let registers = stateUpdate.registers {
                    updateRegisters(registers)
                }

                // Update status
                status = newStatus
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
            // Prevent stale WebSocket events from overriding explicit stop
            if status == .halted {
                DebugLog.warning(
                    "Ignoring stale breakpoint_hit event while halted (PC: \(event.address.map { String(format: "0x%08X", $0) } ?? "nil"))",
                    category: "ViewModel"
                )
                return
            }
            status = .breakpoint
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
