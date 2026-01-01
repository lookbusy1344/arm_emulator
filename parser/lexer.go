package parser

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	TokenEOF TokenType = iota
	TokenNewline
	TokenComment

	// Identifiers and literals
	TokenIdentifier // labels, instruction mnemonics
	TokenRegister   // R0-R15, SP, LR, PC
	TokenNumber     // immediate values
	TokenString     // string literals

	// Operators and punctuation
	TokenComma     // ,
	TokenColon     // :
	TokenHash      // # (immediate prefix)
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenLBrace    // {
	TokenRBrace    // }
	TokenExclaim   // !
	TokenPlus      // +
	TokenMinus     // -
	TokenStar      // *
	TokenSlash     // /
	TokenPercent   // %
	TokenAmpersand // &
	TokenPipe      // |
	TokenCaret     // ^
	TokenTilde     // ~
	TokenLShift    // <<
	TokenRShift    // >>
	TokenEqual     // =

	// Directives
	TokenDirective // .org, .equ, .word, etc.

	// Condition codes (as part of instruction)
	TokenCondition // EQ, NE, CS, CC, etc.
)

var tokenNames = map[TokenType]string{
	TokenEOF:        "EOF",
	TokenNewline:    "NEWLINE",
	TokenComment:    "COMMENT",
	TokenIdentifier: "IDENTIFIER",
	TokenRegister:   "REGISTER",
	TokenNumber:     "NUMBER",
	TokenString:     "STRING",
	TokenComma:      ",",
	TokenColon:      ":",
	TokenHash:       "#",
	TokenLBracket:   "[",
	TokenRBracket:   "]",
	TokenLBrace:     "{",
	TokenRBrace:     "}",
	TokenExclaim:    "!",
	TokenPlus:       "+",
	TokenMinus:      "-",
	TokenStar:       "*",
	TokenSlash:      "/",
	TokenPercent:    "%",
	TokenAmpersand:  "&",
	TokenPipe:       "|",
	TokenCaret:      "^",
	TokenTilde:      "~",
	TokenLShift:     "<<",
	TokenRShift:     ">>",
	TokenEqual:      "=",
	TokenDirective:  "DIRECTIVE",
	TokenCondition:  "CONDITION",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TokenType(%d)", t)
}

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Pos     Position
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q) at %s", t.Type, t.Literal, t.Pos)
}

// Lexer tokenizes ARM assembly source code
type Lexer struct {
	input    string
	filename string
	pos      int  // current position in input
	line     int  // current line number
	column   int  // current column number
	ch       rune // current character
	errors   *ErrorList
}

// NewLexer creates a new lexer for the given input
func NewLexer(input, filename string) *Lexer {
	l := &Lexer{
		input:    input,
		filename: filename,
		pos:      0,
		line:     1,
		column:   0,
		errors:   &ErrorList{},
	}
	l.readChar() // Initialize first character
	return l
}

// readChar reads the next character
func (l *Lexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = rune(l.input[l.pos])
	}
	l.pos++
	l.column++
}

// peekChar returns the next character without advancing
func (l *Lexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

// currentPos returns the current position
func (l *Lexer) currentPos() Position {
	return Position{
		Filename: l.filename,
		Line:     l.line,
		Column:   l.column,
	}
}

// skipWhitespace skips spaces and tabs (but not newlines)
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}
}

// skipLineComment skips a line comment starting with ;, @, or //
func (l *Lexer) skipLineComment() string {
	start := l.pos - 1
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[start : l.pos-1]
}

// skipBlockComment skips a block comment /* ... */
func (l *Lexer) skipBlockComment() string {
	start := l.pos - 1
	startLine := l.line
	startCol := l.column

	for {
		if l.ch == 0 {
			l.errors.AddError(NewError(
				Position{l.filename, startLine, startCol},
				ErrorSyntax,
				"unterminated block comment",
			))
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // consume *
			l.readChar() // consume /
			break
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}

	return l.input[start : l.pos-1]
}

// readIdentifier reads an identifier (label, instruction, etc.)
func (l *Lexer) readIdentifier() string {
	start := l.pos - 1 // Current character position
	// Identifiers can start with letter, underscore, or dot (for local labels)
	// and contain letters, digits, underscores
	for isIdentifierChar(l.ch) {
		l.readChar()
	}
	return l.input[start : l.pos-1]
}

// isIdentifierChar returns true if the character can be part of an identifier
func isIdentifierChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '.'
}

// readNumber reads a number (decimal, hex, binary, octal)
func (l *Lexer) readNumber() string {
	start := l.pos - 1

	// Check for hex (0x), binary (0b), octal (0o)
	if l.ch == '0' {
		next := l.peekChar()
		if next == 'x' || next == 'X' {
			l.readChar() // consume 0
			l.readChar() // consume x
			for isHexDigit(l.ch) {
				l.readChar()
			}
			return l.input[start : l.pos-1]
		} else if next == 'b' || next == 'B' {
			l.readChar() // consume 0
			l.readChar() // consume b
			for l.ch == '0' || l.ch == '1' {
				l.readChar()
			}
			return l.input[start : l.pos-1]
		} else if next == 'o' || next == 'O' {
			l.readChar() // consume 0
			l.readChar() // consume o
			for l.ch >= '0' && l.ch <= '7' {
				l.readChar()
			}
			return l.input[start : l.pos-1]
		}
	}

	// Decimal number
	for unicode.IsDigit(l.ch) {
		l.readChar()
	}

	return l.input[start : l.pos-1]
}

// isHexDigit returns true if the character is a hex digit
func isHexDigit(ch rune) bool {
	return unicode.IsDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// readString reads a string literal
func (l *Lexer) readString(quote rune) string {
	start := l.pos - 1 // position after opening quote (we've already consumed it)
	startLine := l.line
	startCol := l.column - 1

	for {
		if l.ch == 0 {
			l.errors.AddError(NewError(
				Position{l.filename, startLine, startCol},
				ErrorSyntax,
				"unterminated string literal",
			))
			break
		}
		if l.ch == quote {
			result := l.input[start : l.pos-1]
			l.readChar() // consume closing quote
			return result
		}
		if l.ch == '\\' {
			l.readChar() // consume backslash
			if l.ch != 0 {
				l.readChar() // consume escaped character
			}
		} else {
			if l.ch == '\n' {
				l.line++
				l.column = 0
			}
			l.readChar()
		}
	}

	return l.input[start : l.pos-1]
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	pos := l.currentPos()
	tok := Token{Pos: pos}

	switch l.ch {
	case 0:
		tok.Type = TokenEOF
		tok.Literal = ""
		return tok

	case '\n', '\r':
		tok.Type = TokenNewline
		tok.Literal = "\n"
		if l.ch == '\r' && l.peekChar() == '\n' {
			l.readChar() // consume \r
		}
		l.readChar()
		l.line++
		l.column = 0
		return tok

	case ';', '@':
		tok.Type = TokenComment
		tok.Literal = l.skipLineComment()
		return tok

	case '/':
		if l.peekChar() == '/' {
			l.readChar() // consume first /
			tok.Type = TokenComment
			tok.Literal = l.skipLineComment()
			return tok
		} else if l.peekChar() == '*' {
			l.readChar() // consume /
			l.readChar() // consume *
			tok.Type = TokenComment
			tok.Literal = l.skipBlockComment()
			return tok
		} else {
			tok.Type = TokenSlash
			tok.Literal = "/"
			l.readChar()
			return tok
		}

	case ',':
		tok.Type = TokenComma
		tok.Literal = ","
		l.readChar()

	case ':':
		tok.Type = TokenColon
		tok.Literal = ":"
		l.readChar()

	case '#':
		tok.Type = TokenHash
		tok.Literal = "#"
		l.readChar()

	case '[':
		tok.Type = TokenLBracket
		tok.Literal = "["
		l.readChar()

	case ']':
		tok.Type = TokenRBracket
		tok.Literal = "]"
		l.readChar()

	case '{':
		tok.Type = TokenLBrace
		tok.Literal = "{"
		l.readChar()

	case '}':
		tok.Type = TokenRBrace
		tok.Literal = "}"
		l.readChar()

	case '!':
		tok.Type = TokenExclaim
		tok.Literal = "!"
		l.readChar()

	case '+':
		tok.Type = TokenPlus
		tok.Literal = "+"
		l.readChar()

	case '-':
		tok.Type = TokenMinus
		tok.Literal = "-"
		l.readChar()

	case '*':
		tok.Type = TokenStar
		tok.Literal = "*"
		l.readChar()

	case '%':
		tok.Type = TokenPercent
		tok.Literal = "%"
		l.readChar()

	case '&':
		tok.Type = TokenAmpersand
		tok.Literal = "&"
		l.readChar()

	case '|':
		tok.Type = TokenPipe
		tok.Literal = "|"
		l.readChar()

	case '^':
		tok.Type = TokenCaret
		tok.Literal = "^"
		l.readChar()

	case '~':
		tok.Type = TokenTilde
		tok.Literal = "~"
		l.readChar()

	case '<':
		if l.peekChar() == '<' {
			l.readChar()
			tok.Type = TokenLShift
			tok.Literal = "<<"
			l.readChar()
		} else {
			l.errors.AddError(NewError(pos, ErrorSyntax, fmt.Sprintf("unexpected character: %q", l.ch)))
			l.readChar()
			return l.NextToken()
		}

	case '>':
		if l.peekChar() == '>' {
			l.readChar()
			tok.Type = TokenRShift
			tok.Literal = ">>"
			l.readChar()
		} else {
			l.errors.AddError(NewError(pos, ErrorSyntax, fmt.Sprintf("unexpected character: %q", l.ch)))
			l.readChar()
			return l.NextToken()
		}

	case '=':
		tok.Type = TokenEqual
		tok.Literal = "="
		l.readChar()

	case '"', '\'':
		quote := l.ch
		l.readChar() // consume opening quote
		tok.Type = TokenString
		tok.Literal = l.readString(quote)

	case '.':
		// Could be a directive or a local label
		if unicode.IsLetter(l.peekChar()) {
			// Directive
			ident := l.readIdentifier() // This will read . and following chars
			tok.Type = TokenDirective
			tok.Literal = ident
		} else if unicode.IsDigit(l.peekChar()) || l.peekChar() == '_' {
			// Local label like .L1
			ident := l.readIdentifier()
			tok.Type = TokenIdentifier
			tok.Literal = ident
		} else {
			l.errors.AddError(NewError(pos, ErrorSyntax, fmt.Sprintf("unexpected character after '.': %q", l.peekChar())))
			l.readChar()
			return l.NextToken()
		}

	default:
		if unicode.IsLetter(l.ch) || l.ch == '_' {
			// Identifier (label, instruction, register)
			ident := l.readIdentifier()

			// Check if it's a register
			upperIdent := strings.ToUpper(ident)
			if isRegister(upperIdent) {
				tok.Type = TokenRegister
				tok.Literal = upperIdent
			} else {
				tok.Type = TokenIdentifier
				tok.Literal = ident
			}

		} else if unicode.IsDigit(l.ch) {
			// Number
			number := l.readNumber()
			tok.Type = TokenNumber
			tok.Literal = number

		} else {
			l.errors.AddError(NewError(pos, ErrorSyntax, fmt.Sprintf("unexpected character: %q", l.ch)))
			l.readChar()
			return l.NextToken()
		}
	}

	return tok
}

// isRegister checks if a string is a valid register name
func isRegister(s string) bool {
	// Check R0-R15
	if len(s) >= 2 && s[0] == 'R' {
		numStr := s[1:]
		for i := 0; i < len(numStr); i++ {
			if !unicode.IsDigit(rune(numStr[i])) {
				return false
			}
		}
		// Parse and validate range 0-15
		num := 0
		for i := 0; i < len(numStr); i++ {
			num = num*10 + int(numStr[i]-'0')
			if num > 15 {
				return false
			}
		}
		return true
	}

	// Check aliases
	switch s {
	case "SP", "LR", "PC":
		return true
	}

	return false
}

// Errors returns the error list
func (l *Lexer) Errors() *ErrorList {
	return l.errors
}

// TokenizeAll tokenizes the entire input and returns all tokens
func (l *Lexer) TokenizeAll() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}
