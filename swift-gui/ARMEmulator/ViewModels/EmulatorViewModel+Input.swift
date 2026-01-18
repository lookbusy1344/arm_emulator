import Foundation

// MARK: - Input Operations Extension

extension EmulatorViewModel {
    func sendInput(_ input: String) async {
        DebugLog.log("sendInput() called with input: \(input.prefix(20))...", category: "ViewModel")
        DebugLog.log("Current status: \(status)", category: "ViewModel")

        guard let sessionID else {
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
                    category: "ViewModel",
                )
                try await refreshState()
            } else {
                // VM was not waiting - the backend buffered the input for later.
                // Call step() to consume the buffered input.
                DebugLog.log(
                    "VM was not waiting - stepping to consume buffered input...",
                    category: "ViewModel",
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
