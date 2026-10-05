package parser

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// EvaluateExpression evaluates a constant expression over 32-bit unsigned values.
// Operands are numbers in any ParseNumber form and symbols that lookup resolves.
// Operators, from lowest to highest precedence: |, ^, &, << >>, + -, * / %, and
// unary - and ~. Parentheses group.
func EvaluateExpression(expr string, lookup func(name string) (uint32, bool)) (uint32, error) {
	e := &exprEvaluator{tokens: tokenizeExpression(expr), lookup: lookup}
	if len(e.tokens) == 0 {
		return 0, fmt.Errorf("empty expression")
	}
	value, err := e.binary(0)
	if err != nil {
		return 0, err
	}
	if e.pos < len(e.tokens) {
		return 0, fmt.Errorf("unexpected %q in expression %q", e.tokens[e.pos], expr)
	}
	return value, nil
}

const bitsInWord = 32

// binaryLevels lists binary operators by precedence, lowest first.
var binaryLevels = [][]string{{"|"}, {"^"}, {"&"}, {"<<", ">>"}, {"+", "-"}, {"*", "/", "%"}}

type exprEvaluator struct {
	tokens []string
	pos    int
	lookup func(string) (uint32, bool)
}

func (e *exprEvaluator) peek() string {
	if e.pos < len(e.tokens) {
		return e.tokens[e.pos]
	}
	return ""
}

func (e *exprEvaluator) binary(level int) (uint32, error) {
	if level == len(binaryLevels) {
		return e.unary()
	}
	left, err := e.binary(level + 1)
	if err != nil {
		return 0, err
	}
	for {
		op := e.peek()
		if !slices.Contains(binaryLevels[level], op) {
			return left, nil
		}
		e.pos++
		right, err := e.binary(level + 1)
		if err != nil {
			return 0, err
		}
		if left, err = applyBinary(op, left, right); err != nil {
			return 0, err
		}
	}
}

func (e *exprEvaluator) unary() (uint32, error) {
	switch tok := e.peek(); tok {
	case "-", "~":
		e.pos++
		v, err := e.unary()
		if tok == "-" {
			return -v, err
		}
		return ^v, err
	case "(":
		e.pos++
		v, err := e.binary(0)
		if err != nil {
			return 0, err
		}
		if e.peek() != ")" {
			return 0, fmt.Errorf("missing ')'")
		}
		e.pos++
		return v, nil
	case "":
		return 0, fmt.Errorf("expression ends where a value is expected")
	default:
		e.pos++
		return e.operand(tok)
	}
}

func (e *exprEvaluator) operand(tok string) (uint32, error) {
	if unicode.IsDigit(rune(tok[0])) {
		return ParseNumber(tok)
	}
	if isIdentStart(rune(tok[0])) {
		if v, ok := e.lookup(tok); ok {
			return v, nil
		}
		return 0, fmt.Errorf("undefined symbol %q", tok)
	}
	return 0, fmt.Errorf("unexpected %q in expression", tok)
}

func applyBinary(op string, a, b uint32) (uint32, error) {
	switch op {
	case "|":
		return a | b, nil
	case "^":
		return a ^ b, nil
	case "&":
		return a & b, nil
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/", "%":
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		if op == "/" {
			return a / b, nil
		}
		return a % b, nil
	}
	// << and >>
	if b >= bitsInWord {
		return 0, fmt.Errorf("shift amount %d out of range (0-31)", b)
	}
	if op == "<<" {
		return a << b, nil
	}
	return a >> b, nil
}

// tokenizeExpression splits an expression into numbers, identifiers and operators.
func tokenizeExpression(expr string) []string {
	var tokens []string
	runes := []rune(expr)
	for i := 0; i < len(runes); {
		r := runes[i]
		switch {
		case unicode.IsSpace(r):
			i++
		case unicode.IsDigit(r) || isIdentStart(r):
			j := i + 1
			for j < len(runes) && (unicode.IsLetter(runes[j]) || unicode.IsDigit(runes[j]) || runes[j] == '_' || runes[j] == '.') {
				j++
			}
			tokens = append(tokens, string(runes[i:j]))
			i = j
		case strings.HasPrefix(string(runes[i:]), "<<"), strings.HasPrefix(string(runes[i:]), ">>"):
			tokens = append(tokens, string(runes[i:i+2]))
			i += 2
		default:
			tokens = append(tokens, string(r))
			i++
		}
	}
	return tokens
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_' || r == '.'
}
