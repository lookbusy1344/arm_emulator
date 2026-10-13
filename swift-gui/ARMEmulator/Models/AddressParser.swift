import Foundation

/// Reads a 32-bit address typed by the user: `0x` and hex digits, or a bare decimal number
enum AddressParser {
    private static let hexPrefix = "0x"
    private static let hexRadix = 16

    /// The address `text` names, ignoring surrounding whitespace, or nil when it names none
    static func parse(_ text: String) -> UInt32? {
        let trimmed = text.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.lowercased().hasPrefix(hexPrefix) {
            let digits = trimmed.dropFirst(hexPrefix.count)
            guard digits.allSatisfy(\.isHexDigit) else { return nil }
            return UInt32(digits, radix: hexRadix)
        }
        guard trimmed.allSatisfy(\.isASCIIDigitCharacter) else { return nil }
        return UInt32(trimmed)
    }
}

private extension Character {
    var isASCIIDigitCharacter: Bool {
        isASCII && isNumber
    }
}
