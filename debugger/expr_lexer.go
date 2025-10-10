package debugger

import (
	"fmt"
	"strings"
	"unicode"
)

// ExprTokenType represents the type of expression token
type ExprTokenType int

const (
	ExprTokenEOF ExprTokenType = iota
	ExprTokenNumber
	ExprTokenRegister
	ExprTokenSymbol
	ExprTokenOperator
	ExprTokenLParen
	ExprTokenRParen
	ExprTokenLBracket
	ExprTokenRBracket
	ExprTokenStar     // for memory dereference
	ExprTokenValueRef // $1, $2, etc.
)

// ExprToken represents a token in an expression
type ExprToken struct {
	Type  ExprTokenType
	Value string
	Pos   int
}

// ExprLexer tokenizes debugger expressions
type ExprLexer struct {
	input string
	pos   int
	ch    rune
}

// NewExprLexer creates a new expression lexer
func NewExprLexer(input string) *ExprLexer {
	l := &ExprLexer{
		input: input,
		pos:   0,
	}
	l.readChar()
	return l
}

// readChar reads the next character
func (l *ExprLexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.pos])
	}
	l.pos++
}

// peekChar returns the next character without advancing
func (l *ExprLexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

// skipWhitespace skips whitespace characters
func (l *ExprLexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readNumber reads a complete number (hex, binary, octal, or decimal)
func (l *ExprLexer) readNumber() string {
	start := l.pos - 1

	// Check for hex (0x), binary (0b), or octal (0)
	if l.ch == '0' {
		next := l.peekChar()
		if next == 'x' || next == 'X' {
			// Hexadecimal
			l.readChar() // consume 0
			l.readChar() // consume x
			for isHexDigit(l.ch) {
				l.readChar()
			}
			return l.input[start : l.pos-1]
		} else if next == 'b' || next == 'B' {
			// Binary
			l.readChar() // consume 0
			l.readChar() // consume b
			for l.ch == '0' || l.ch == '1' {
				l.readChar()
			}
			return l.input[start : l.pos-1]
		}
	}

	// Decimal or octal number
	for unicode.IsDigit(l.ch) {
		l.readChar()
	}

	return l.input[start : l.pos-1]
}

// isHexDigit checks if a character is a hex digit
func isHexDigit(ch rune) bool {
	return unicode.IsDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// readIdentifier reads an identifier (register or symbol name)
func (l *ExprLexer) readIdentifier() string {
	start := l.pos - 1
	for unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[start : l.pos-1]
}

// isRegisterName checks if a string is a register name
func isRegisterName(s string) bool {
	s = strings.ToLower(s)

	// Check for r0-r15
	if strings.HasPrefix(s, "r") && len(s) >= 2 {
		for i := 1; i < len(s); i++ {
			if !unicode.IsDigit(rune(s[i])) {
				return false
			}
		}
		return true
	}

	// Check for aliases
	switch s {
	case "sp", "lr", "pc":
		return true
	}

	return false
}

// NextToken returns the next token from the input
func (l *ExprLexer) NextToken() ExprToken {
	l.skipWhitespace()

	pos := l.pos - 1
	tok := ExprToken{Pos: pos}

	switch l.ch {
	case 0:
		tok.Type = ExprTokenEOF
		tok.Value = ""
		return tok

	case '(':
		tok.Type = ExprTokenLParen
		tok.Value = "("
		l.readChar()

	case ')':
		tok.Type = ExprTokenRParen
		tok.Value = ")"
		l.readChar()

	case '[':
		tok.Type = ExprTokenLBracket
		tok.Value = "["
		l.readChar()

	case ']':
		tok.Type = ExprTokenRBracket
		tok.Value = "]"
		l.readChar()

	case '*':
		// Could be multiply or memory dereference
		// It's a dereference if it's at the start or after an operator/bracket
		// Otherwise it's multiply
		tok.Type = ExprTokenOperator
		tok.Value = "*"
		l.readChar()

	case '+':
		tok.Type = ExprTokenOperator
		tok.Value = "+"
		l.readChar()

	case '-':
		// Could be minus or negative number
		next := l.peekChar()
		if unicode.IsDigit(next) {
			// Negative number
			l.readChar() // consume -
			num := l.readNumber()
			tok.Type = ExprTokenNumber
			tok.Value = "-" + num
		} else {
			tok.Type = ExprTokenOperator
			tok.Value = "-"
			l.readChar()
		}

	case '/':
		tok.Type = ExprTokenOperator
		tok.Value = "/"
		l.readChar()

	case '&':
		tok.Type = ExprTokenOperator
		tok.Value = "&"
		l.readChar()

	case '|':
		tok.Type = ExprTokenOperator
		tok.Value = "|"
		l.readChar()

	case '^':
		tok.Type = ExprTokenOperator
		tok.Value = "^"
		l.readChar()

	case '<':
		if l.peekChar() == '<' {
			tok.Type = ExprTokenOperator
			tok.Value = "<<"
			l.readChar()
			l.readChar()
		} else {
			return ExprToken{Type: ExprTokenEOF, Value: "", Pos: pos} // Error
		}

	case '>':
		if l.peekChar() == '>' {
			tok.Type = ExprTokenOperator
			tok.Value = ">>"
			l.readChar()
			l.readChar()
		} else {
			return ExprToken{Type: ExprTokenEOF, Value: "", Pos: pos} // Error
		}

	case '$':
		// Value reference ($1, $2, etc.)
		l.readChar() // consume $
		if unicode.IsDigit(l.ch) {
			num := l.readNumber()
			tok.Type = ExprTokenValueRef
			tok.Value = "$" + num
		} else {
			return ExprToken{Type: ExprTokenEOF, Value: "", Pos: pos} // Error
		}

	default:
		if unicode.IsDigit(l.ch) {
			// Number
			num := l.readNumber()
			tok.Type = ExprTokenNumber
			tok.Value = num
		} else if unicode.IsLetter(l.ch) || l.ch == '_' {
			// Identifier (register or symbol)
			ident := l.readIdentifier()
			if isRegisterName(ident) {
				tok.Type = ExprTokenRegister
				tok.Value = strings.ToLower(ident)
			} else {
				tok.Type = ExprTokenSymbol
				tok.Value = ident
			}
		} else {
			return ExprToken{Type: ExprTokenEOF, Value: "", Pos: pos} // Error
		}
	}

	return tok
}

// TokenizeAll tokenizes the entire input
func (l *ExprLexer) TokenizeAll() []ExprToken {
	var tokens []ExprToken
	for {
		tok := l.NextToken()
		if tok.Type == ExprTokenEOF {
			tokens = append(tokens, tok)
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func (t ExprTokenType) String() string {
	switch t {
	case ExprTokenEOF:
		return "EOF"
	case ExprTokenNumber:
		return "NUMBER"
	case ExprTokenRegister:
		return "REGISTER"
	case ExprTokenSymbol:
		return "SYMBOL"
	case ExprTokenOperator:
		return "OPERATOR"
	case ExprTokenLParen:
		return "LPAREN"
	case ExprTokenRParen:
		return "RPAREN"
	case ExprTokenLBracket:
		return "LBRACKET"
	case ExprTokenRBracket:
		return "RBRACKET"
	case ExprTokenStar:
		return "STAR"
	case ExprTokenValueRef:
		return "VALUEREF"
	default:
		return fmt.Sprintf("ExprTokenType(%d)", t)
	}
}
