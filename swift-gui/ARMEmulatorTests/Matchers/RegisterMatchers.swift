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
        case "R13", "SP": return r13 == value
        case "R14", "LR": return r14 == value
        case "R15", "PC": return r15 == value
        case "CPSR": return cpsr == value
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
        case "R13", "SP": return r13
        case "R14", "LR": return r14
        case "R15", "PC": return r15
        case "CPSR": return cpsr
        default: return nil
        }
    }

    /// Check if CPSR has specific flag set
    /// - Parameter flag: Flag name ("N", "Z", "C", "V")
    /// - Returns: True if flag is set
    func hasFlag(_ flag: String) -> Bool {
        switch flag.uppercased() {
        case "N": return (cpsr & 0x8000_0000) != 0  // Bit 31
        case "Z": return (cpsr & 0x4000_0000) != 0  // Bit 30
        case "C": return (cpsr & 0x2000_0000) != 0  // Bit 29
        case "V": return (cpsr & 0x1000_0000) != 0  // Bit 28
        default: return false
        }
    }
}

/// Matchers for ProgramState testing
extension ProgramState {
    /// Check if console output contains a specific string
    /// - Parameter text: Text to search for
    /// - Returns: True if console output contains the text
    func consoleContains(_ text: String) -> Bool {
        consoleOutput.contains(text)
    }

    /// Check if console output matches a pattern
    /// - Parameter pattern: Regex pattern to match
    /// - Returns: True if console output matches the pattern
    func consoleMatches(_ pattern: String) -> Bool {
        guard let regex = try? NSRegularExpression(pattern: pattern) else {
            return false
        }
        let range = NSRange(consoleOutput.startIndex..., in: consoleOutput)
        return regex.firstMatch(in: consoleOutput, range: range) != nil
    }

    /// Check if any breakpoint exists at a specific address
    /// - Parameter address: Memory address
    /// - Returns: True if breakpoint exists at address
    func hasBreakpoint(at address: UInt32) -> Bool {
        breakpoints.contains { $0.address == address }
    }

    /// Check if breakpoint is enabled
    /// - Parameter address: Memory address
    /// - Returns: True if breakpoint exists and is enabled
    func isBreakpointEnabled(at address: UInt32) -> Bool {
        breakpoints.first { $0.address == address }?.enabled ?? false
    }
}
