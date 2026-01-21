import Foundation
@testable import ARMEmulator

/// Matchers and assertions for RegisterState testing
extension RegisterState {
    /// Check if a specific register has a specific value
    /// - Parameters:
    ///   - name: Register name (e.g., "R0", "R1", "SP", "PC")
    ///   - value: Expected value
    /// - Returns: True if the register matches the value
    func hasRegister(_ name: String, value: UInt32) -> Bool {
        switch name.uppercased() {
        case "R0": return r0 == value
        case "R1": return r1 == value
        case "R2": return r2 == value
        case "R3": return r3 == value
        case "R4": return r4 == value
        case "R5": return r5 == value
        case "R6": return r6 == value
        case "R7": return r7 == value
        case "R8": return r8 == value
        case "R9": return r9 == value
        case "R10": return r10 == value
        case "R11": return r11 == value
        case "R12": return r12 == value
        case "R13", "SP": return sp == value
        case "R14", "LR": return lr == value
        case "R15", "PC": return pc == value
        default: return false
        }
    }

    /// Get register value by name
    /// - Parameter name: Register name (e.g., "R0", "R1", "SP", "PC")
    /// - Returns: Register value or nil if invalid name
    func registerValue(_ name: String) -> UInt32? {
        switch name.uppercased() {
        case "R0": return r0
        case "R1": return r1
        case "R2": return r2
        case "R3": return r3
        case "R4": return r4
        case "R5": return r5
        case "R6": return r6
        case "R7": return r7
        case "R8": return r8
        case "R9": return r9
        case "R10": return r10
        case "R11": return r11
        case "R12": return r12
        case "R13", "SP": return sp
        case "R14", "LR": return lr
        case "R15", "PC": return pc
        default: return nil
        }
    }

    /// Check if CPSR has specific flag set
    /// - Parameter flag: Flag name ("N", "Z", "C", "V")
    /// - Returns: True if flag is set
    func hasFlag(_ flag: String) -> Bool {
        switch flag.uppercased() {
        case "N": return cpsr.n
        case "Z": return cpsr.z
        case "C": return cpsr.c
        case "V": return cpsr.v
        default: return false
        }
    }
}

// ProgramState matchers removed - model doesn't exist in current architecture
