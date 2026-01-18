import Foundation

// MARK: - Execution Control Extension

extension EmulatorViewModel {
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
        DebugLog.log("stop() called - current status: \(status), canStep: \(canStep)", category: "ViewModel")

        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        do {
            DebugLog.log("Calling apiClient.stop()...", category: "ViewModel")
            try await apiClient.stop(sessionID: sessionID)
            DebugLog.success("apiClient.stop() succeeded", category: "ViewModel")

            DebugLog.log("Refreshing state after stop...", category: "ViewModel")
            try await refreshState()
            DebugLog.log("After stop - status: \(status), canStep: \(canStep)", category: "ViewModel")
            errorMessage = nil
        } catch {
            DebugLog.error("stop() failed: \(error.localizedDescription)", category: "ViewModel")
            errorMessage = "Failed to stop: \(error.localizedDescription)"
        }
    }

    func step() async {
        DebugLog.log("step() called - status: \(status), canStep: \(canStep)", category: "ViewModel")

        guard let sessionID = sessionID else {
            errorMessage = "No active session"
            return
        }

        // Clear memory write tracking before step to ensure onChange triggers
        lastMemoryWrite = nil

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

        // Clear memory write tracking before step to ensure onChange triggers
        lastMemoryWrite = nil

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

        // Clear memory write tracking before step to ensure onChange triggers
        lastMemoryWrite = nil

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
        // Clear highlights when restarting
        cancelAllHighlights()

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
}
