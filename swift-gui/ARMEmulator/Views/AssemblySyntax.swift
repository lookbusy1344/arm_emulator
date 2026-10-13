import AppKit

/// Tokenises ARM assembly source for highlighting. Keywords ignore case.
enum AssemblySyntax {
    enum Kind: String {
        case comment, string, label, directive, instruction, register, number
    }

    struct Token: Equatable {
        let range: NSRange
        let kind: Kind
    }

    private static let mnemonics: Set = [
        "mov", "mvn", "add", "adc", "sub", "sbc", "rsb", "rsc", "mul", "mla", "umull", "umlal", "smull", "smlal",
        "and", "orr", "eor", "bic", "lsl", "lsr", "asr", "ror", "rrx", "cmp", "cmn", "tst", "teq",
        "b", "bl", "bx", "blx", "ldr", "ldrb", "ldrh", "ldrsb", "ldrsh", "str", "strb", "strh",
        "ldm", "ldmia", "ldmib", "ldmda", "ldmdb", "ldmfd", "ldmfa", "ldmed", "ldmea",
        "stm", "stmia", "stmib", "stmda", "stmdb", "stmfd", "stmfa", "stmed", "stmea",
        "push", "pop", "swi", "svc", "bkpt", "nop", "mrs", "msr", "swp", "swpb", "adr",
    ]

    private static let conditions = [
        "eq", "ne", "cs", "hs", "cc", "lo", "mi", "pl", "vs", "vc", "hi", "ls", "ge", "lt", "gt", "le", "al",
    ]

    /// What may follow a mnemonic: a condition code, then an S that sets the flags
    private static let suffixes = Set(conditions + conditions.map { $0 + "s" } + ["s"])

    private static let registers: Set = [
        "r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15",
        "sp", "lr", "pc", "cpsr", "spsr",
    ]

    /// One alternative per kind. A match consumes its text, so markers inside strings and comments are not seen;
    /// the label alternative precedes the plain word. `token(_:in:)` tells the kinds apart by their characters.
    /// Computed because Regex is not Sendable.
    private static var lexer: some RegexComponent<Substring> {
        #/
            " (?: \\. | [^"\\\n] )* "?
          | (?: ; | @ | // ) [^\n]*
          | \b [A-Za-z_] \w* :
          | \. [A-Za-z_] \w*
          | \#? 0x [0-9A-Fa-f]+ \b | \#? 0b [01]+ \b | \#? -? \b [0-9]+ \b
          | \b [A-Za-z_] \w* \b
        /#.wordBoundaryKind(.simple)
    }

    /// Tokens in `range` of `text`, or in all of it. Ranges are UTF-16 offsets into `text`.
    static func tokens(in text: String, range: NSRange? = nil) -> [Token] {
        let whole = text.startIndex ..< text.endIndex
        let searchRange = range.flatMap { Range($0, in: text) } ?? whole
        return text[searchRange].matches(of: lexer).compactMap { match in
            token(match.range, in: text)
        }
    }

    private static func token(_ range: Range<String.Index>, in text: String) -> Token? {
        let lexeme = text[range]
        let kind: Kind? = switch lexeme.first {
        case "\"": .string
        case ";", "@", "/": .comment
        case ".": .directive
        case "#", "-", .some("0" ... "9"): .number
        default: lexeme.last == ":" ? .label : wordKind(lexeme.lowercased())
        }
        return kind.map { Token(range: NSRange(range, in: text), kind: $0) }
    }

    private static func wordKind(_ word: String) -> Kind? {
        if registers.contains(word) {
            return .register
        }
        let isInstruction = mnemonics.contains(word) || mnemonics.contains { mnemonic in
            word.hasPrefix(mnemonic) && suffixes.contains(String(word.dropFirst(mnemonic.count)))
        }
        return isInstruction ? .instruction : nil
    }
}

/// Colours the text storage of an editor as it changes, restyling only the edited lines
final class AssemblyHighlighter: NSObject, NSTextStorageDelegate {
    private static func color(for kind: AssemblySyntax.Kind) -> NSColor {
        switch kind {
        case .comment: .systemGreen
        case .string: .systemBrown
        case .label: .systemOrange
        case .directive: .systemPurple
        case .instruction: .systemBlue
        case .register: .systemTeal
        case .number: .systemRed
        }
    }

    func textStorage(
        _ textStorage: NSTextStorage,
        didProcessEditing editedMask: NSTextStorageEditActions,
        range editedRange: NSRange,
        changeInLength _: Int,
    ) {
        guard editedMask.contains(.editedCharacters) else { return }
        let lines = (textStorage.string as NSString).lineRange(for: editedRange)
        textStorage.addAttribute(.foregroundColor, value: NSColor.labelColor, range: lines)
        for token in AssemblySyntax.tokens(in: textStorage.string, range: lines) {
            textStorage.addAttribute(.foregroundColor, value: Self.color(for: token.kind), range: token.range)
        }
    }
}
