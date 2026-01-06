import Foundation

struct Watchpoint: Codable, Identifiable, Hashable {
    let id: Int
    let address: UInt32
    let type: String // "read", "write", "readwrite"

    func hash(into hasher: inout Hasher) {
        hasher.combine(id)
    }

    static func == (lhs: Watchpoint, rhs: Watchpoint) -> Bool {
        lhs.id == rhs.id
    }
}
