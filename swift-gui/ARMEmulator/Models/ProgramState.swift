import Foundation

enum VMState: String, Codable {
    case idle
    case running
    case paused
    case halted
    case error
}

struct VMStatus: Codable {
    var state: String
    var pc: UInt32
    var instruction: String? // Optional - backend doesn't always return this
    var cycleCount: UInt64?
    var error: String?
    var hasWrite: Bool?
    var writeAddr: UInt32?

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

struct DisassemblyInstruction: Codable {
    var address: UInt32
    var machineCode: UInt32
    var disassembly: String
    var symbol: String?
}
