import Foundation
@testable import ARMEmulator

/// Matchers and assertions for RegisterState testing
extension RegisterState {
    /// Check if a specific register has a specific value
    /// - Parameters:
    ///   - name: Register name (e.g., "R0", "R1", "SP", "PC")
    ///   - value: Expected value
    /// - Returns: True if the register matches the value
    func hasRegister(_ name: String, value: UInt32) -> Bool { // swiftlint:disable:this cyclomatic_complexity
        switch name.uppercased() {
        case "R0": r0 == value
        case "R1": r1 == value
        case "R2": r2 == value
        case "R3": r3 == value
        case "R4": r4 == value
        case "R5": r5 == value
        case "R6": r6 == value
        case "R7": r7 == value
        case "R8": r8 == value
        case "R9": r9 == value
        case "R10": r10 == value
        case "R11": r11 == value
        case "R12": r12 == value
        case "R13", "SP": sp == value
        case "R14", "LR": lr == value
        case "R15", "PC": pc == value
        default: false
        }
    }

    /// Get register value by name
    /// - Parameter name: Register name (e.g., "R0", "R1", "SP", "PC")
    /// - Returns: Register value or nil if invalid name
    func registerValue(_ name: String) -> UInt32? { // swiftlint:disable:this cyclomatic_complexity
        switch name.uppercased() {
        case "R0": r0
        case "R1": r1
        case "R2": r2
        case "R3": r3
        case "R4": r4
        case "R5": r5
        case "R6": r6
        case "R7": r7
        case "R8": r8
        case "R9": r9
        case "R10": r10
        case "R11": r11
        case "R12": r12
        case "R13", "SP": sp
        case "R14", "LR": lr
        case "R15", "PC": pc
        default: nil
        }
    }

    /// Check if CPSR has specific flag set
    /// - Parameter flag: Flag name ("N", "Z", "C", "V")
    /// - Returns: True if flag is set
    func hasFlag(_ flag: String) -> Bool {
        switch flag.uppercased() {
        case "N": cpsr.n
        case "Z": cpsr.z
        case "C": cpsr.c
        case "V": cpsr.v
        default: false
        }
    }
}

// ProgramState matchers removed - model doesn't exist in current architecture
