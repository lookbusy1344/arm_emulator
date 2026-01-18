import Foundation

/// VM execution states - these match the backend's ExecutionState values
/// and WebSocket event notifications
enum VMState: String, Codable {
    case idle // No execution, editor editable
    case running // Actively executing, editor read-only
    case breakpoint // Stopped at breakpoint (from step or run), editor read-only
    case halted // Program finished, editor editable
    case error // Error occurred, editor editable
    case waitingForInput = "waiting_for_input" // Blocked on input, editor read-only
}

struct VMStatus: Codable {
    var state: String
    var pc: UInt32
    var instruction: String? // Optional - backend doesn't always return this
    var cycleCount: UInt64?
    var error: String?
    var hasWrite: Bool?
    var writeAddr: UInt32?
    var writeSize: UInt32? // Size of last write in bytes (1, 2, or 4)

    var vmState: VMState {
        VMState(rawValue: state) ?? .idle
    }
}

struct SessionInfo: Codable {
    var sessionId: String
    var createdAt: String?
}

struct MemoryData: Codable {
    var address: UInt32
    var data: [UInt8]
}

struct DisassemblyInstruction: Codable, Identifiable, Hashable {
    var address: UInt32
    var machineCode: UInt32
    var disassembly: String
    var symbol: String?

    var id: UInt32 { address }

    // Alias for compatibility with code expecting 'mnemonic'
    var mnemonic: String { disassembly }

    func hash(into hasher: inout Hasher) {
        hasher.combine(address)
    }

    static func == (lhs: DisassemblyInstruction, rhs: DisassemblyInstruction) -> Bool {
        lhs.address == rhs.address
    }
}
