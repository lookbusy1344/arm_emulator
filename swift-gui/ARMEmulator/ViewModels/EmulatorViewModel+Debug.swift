import Foundation

// MARK: - Debug Operations Extension

extension EmulatorViewModel {
    /// Toggles the breakpoint on a 1-based source line, assembling edited source first so the line maps to its address
    func toggleBreakpoint(atLine line: Int) async {
        guard await ensureAssembled(), let address = lineToAddress[line] else { return }
        await toggleBreakpoint(at: address)
    }

    /// Toggles the breakpoint on the editor caret line, reporting a line that holds no instruction
    func toggleBreakpointAtCaret() async {
        let line = caretLine
        guard await ensureAssembled() else { return }
        guard lineToAddress[line] != nil else {
            errorMessage = "Line \(line) has no instruction for a breakpoint"
            return
        }
        await toggleBreakpoint(atLine: line)
    }

    func toggleBreakpoint(at address: UInt32) async {
        guard let sessionID else {
            errorMessage = "No active session"
            print("toggleBreakpoint: No active session")
            return
        }

        print(
            "toggleBreakpoint: sessionID=\(sessionID), address=0x\(String(format: "%X", address)), current breakpoints=\(breakpoints)",
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
        guard let sessionID else {
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
        guard let sessionID else {
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
        guard let sessionID else { return }

        do {
            watchpoints = try await apiClient.getWatchpoints(sessionID: sessionID)
        } catch {
            errorMessage = "Failed to refresh watchpoints: \(error.localizedDescription)"
        }
    }
}
