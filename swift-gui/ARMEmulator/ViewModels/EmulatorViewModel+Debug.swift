import Foundation

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
