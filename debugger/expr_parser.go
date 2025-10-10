package debugger

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ExprParser parses debugger expressions using precedence climbing
type ExprParser struct {
	tokens  []ExprToken
	pos     int
	vm      *vm.VM
	symbols map[string]uint32
	eval    *ExpressionEvaluator
}

// NewExprParser creates a new expression parser
func NewExprParser(tokens []ExprToken, machine *vm.VM, symbols map[string]uint32, eval *ExpressionEvaluator) *ExprParser {
	return &ExprParser{
		tokens:  tokens,
		pos:     0,
		vm:      machine,
		symbols: symbols,
		eval:    eval,
	}
}

// currentToken returns the current token
func (p *ExprParser) currentToken() ExprToken {
	if p.pos >= len(p.tokens) {
		return ExprToken{Type: ExprTokenEOF}
	}
	return p.tokens[p.pos]
}

// advance moves to the next token
func (p *ExprParser) advance() {
	p.pos++
}

// operatorPrecedence returns the precedence of an operator
// Higher numbers = higher precedence
func operatorPrecedence(op string) int {
	switch op {
	case "|":
		return 1
	case "^":
		return 2
	case "&":
		return 3
	case "<<", ">>":
		return 4
	case "+", "-":
		return 5
	case "*", "/":
		return 6
	default:
		return 0
	}
}

// Parse parses the expression and returns the result
func (p *ExprParser) Parse() (uint32, error) {
	result, err := p.parseExpression(0)
	if err != nil {
		return 0, err
	}

	// Should be at EOF
	if p.currentToken().Type != ExprTokenEOF {
		return 0, fmt.Errorf("unexpected token: %s", p.currentToken().Value)
	}

	return result, nil
}

// parseExpression parses an expression with precedence climbing
func (p *ExprParser) parseExpression(minPrecedence int) (uint32, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return 0, err
	}

	for {
		tok := p.currentToken()
		if tok.Type != ExprTokenOperator {
			break
		}

		precedence := operatorPrecedence(tok.Value)
		if precedence < minPrecedence {
			break
		}

		op := tok.Value
		p.advance() // consume operator

		right, err := p.parseExpression(precedence + 1)
		if err != nil {
			return 0, err
		}

		left, err = p.applyOperator(left, right, op)
		if err != nil {
			return 0, err
		}
	}

	return left, nil
}

// parsePrimary parses a primary expression (number, register, memory access, etc.)
func (p *ExprParser) parsePrimary() (uint32, error) {
	tok := p.currentToken()

	switch tok.Type {
	case ExprTokenNumber:
		p.advance()
		return p.parseNumberValue(tok.Value)

	case ExprTokenRegister:
		p.advance()
		return p.parseRegisterValue(tok.Value)

	case ExprTokenSymbol:
		p.advance()
		if addr, exists := p.symbols[tok.Value]; exists {
			return addr, nil
		}
		return 0, fmt.Errorf("unknown symbol: %s", tok.Value)

	case ExprTokenValueRef:
		p.advance()
		// Parse $1, $2, etc.
		numStr := strings.TrimPrefix(tok.Value, "$")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid value reference: %s", tok.Value)
		}
		return p.eval.GetValue(num)

	case ExprTokenLParen:
		// Parenthesized expression
		p.advance() // consume (
		result, err := p.parseExpression(0)
		if err != nil {
			return 0, err
		}
		if p.currentToken().Type != ExprTokenRParen {
			return 0, fmt.Errorf("expected ')', got %s", p.currentToken().Value)
		}
		p.advance() // consume )
		return result, nil

	case ExprTokenLBracket:
		// Memory access [expr]
		p.advance() // consume [
		addr, err := p.parseExpression(0)
		if err != nil {
			return 0, err
		}
		if p.currentToken().Type != ExprTokenRBracket {
			return 0, fmt.Errorf("expected ']', got %s", p.currentToken().Value)
		}
		p.advance() // consume ]

		value, err := p.vm.Memory.ReadWord(addr)
		if err != nil {
			return 0, fmt.Errorf("failed to read memory at 0x%08X: %w", addr, err)
		}
		return value, nil

	case ExprTokenOperator:
		// Handle prefix operators like *addr for memory dereference
		if tok.Value == "*" {
			// Memory dereference *expr
			p.advance() // consume *
			addr, err := p.parsePrimary()
			if err != nil {
				return 0, err
			}

			value, err := p.vm.Memory.ReadWord(addr)
			if err != nil {
				return 0, fmt.Errorf("failed to read memory at 0x%08X: %w", addr, err)
			}
			return value, nil
		}
		return 0, fmt.Errorf("unexpected operator: %s", tok.Value)

	case ExprTokenStar:
		// Memory dereference *expr
		p.advance() // consume *
		addr, err := p.parsePrimary()
		if err != nil {
			return 0, err
		}

		value, err := p.vm.Memory.ReadWord(addr)
		if err != nil {
			return 0, fmt.Errorf("failed to read memory at 0x%08X: %w", addr, err)
		}
		return value, nil

	default:
		return 0, fmt.Errorf("unexpected token: %s (%s)", tok.Value, tok.Type)
	}
}

// parseNumberValue parses a number string to uint32
func (p *ExprParser) parseNumberValue(s string) (uint32, error) {
	s = strings.TrimSpace(s)

	// Hexadecimal
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		var val uint32
		_, err := fmt.Sscanf(strings.ToLower(s), "0x%x", &val)
		if err != nil {
			return 0, err
		}
		return val, nil
	}

	// Binary
	if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") {
		val, err := strconv.ParseUint(s[2:], 2, 32)
		if err != nil {
			return 0, err
		}
		return uint32(val), nil
	}

	// Octal
	if strings.HasPrefix(s, "0") && len(s) > 1 && !strings.ContainsAny(s, "89") {
		val, err := strconv.ParseUint(s, 8, 32)
		if err != nil {
			return 0, err
		}
		return uint32(val), nil
	}

	// Decimal (including negative)
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(val), nil
}

// parseRegisterValue gets the value of a register
func (p *ExprParser) parseRegisterValue(reg string) (uint32, error) {
	reg = strings.ToLower(reg)

	// Special registers
	switch reg {
	case "pc", "r15":
		return p.vm.CPU.PC, nil
	case "sp", "r13":
		return p.vm.CPU.GetSP(), nil
	case "lr", "r14":
		return p.vm.CPU.GetLR(), nil
	}

	// General registers
	if strings.HasPrefix(reg, "r") {
		var regNum int
		_, err := fmt.Sscanf(reg, "r%d", &regNum)
		if err == nil && regNum >= 0 && regNum <= 14 {
			return p.vm.CPU.R[regNum], nil
		}
	}

	return 0, fmt.Errorf("invalid register: %s", reg)
}

// applyOperator applies a binary operator to two values
func (p *ExprParser) applyOperator(left, right uint32, op string) (uint32, error) {
	switch op {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return left / right, nil
	case "&":
		return left & right, nil
	case "|":
		return left | right, nil
	case "^":
		return left ^ right, nil
	case "<<":
		return left << right, nil
	case ">>":
		return left >> right, nil
	default:
		return 0, fmt.Errorf("unknown operator: %s", op)
	}
}
