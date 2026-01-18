import Foundation

struct RegisterState: Codable, Equatable {
    var r0: UInt32
    var r1: UInt32
    var r2: UInt32
    var r3: UInt32
    var r4: UInt32
    var r5: UInt32
    var r6: UInt32
    var r7: UInt32
    var r8: UInt32
    var r9: UInt32
    var r10: UInt32
    var r11: UInt32
    var r12: UInt32
    var sp: UInt32
    var lr: UInt32
    var pc: UInt32
    var cpsr: CPSRFlags

    static var empty: RegisterState {
        RegisterState(
            r0: 0, r1: 0, r2: 0, r3: 0,
            r4: 0, r5: 0, r6: 0, r7: 0,
            r8: 0, r9: 0, r10: 0, r11: 0,
            r12: 0, sp: 0, lr: 0, pc: 0,
            cpsr: CPSRFlags(n: false, z: false, c: false, v: false),
        )
    }
}

struct CPSRFlags: Codable, Equatable {
    var n: Bool // Negative
    var z: Bool // Zero
    var c: Bool // Carry
    var v: Bool // Overflow

    var displayString: String {
        "\(n ? "N" : "-") \(z ? "Z" : "-") \(c ? "C" : "-") \(v ? "V" : "-")"
    }
}
