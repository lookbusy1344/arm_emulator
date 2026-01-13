import Foundation

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
