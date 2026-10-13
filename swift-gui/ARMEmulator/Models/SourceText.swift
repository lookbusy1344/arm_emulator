import Foundation

/// Positions in editor source text
enum SourceText {
    private static let newline = UInt16(UInt8(ascii: "\n"))

    /// The 1-based line holding the UTF-16 `offset`, the unit NSTextView uses for caret positions.
    /// Offsets outside the text clamp to its bounds.
    static func line(atOffset offset: Int, in text: String) -> Int {
        let units = text.utf16
        let end = units.index(units.startIndex, offsetBy: min(max(offset, 0), units.count))
        return units[..<end].count { $0 == newline } + 1
    }
}
