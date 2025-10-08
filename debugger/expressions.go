package debugger

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// ExpressionEvaluator evaluates expressions in debugger commands
type ExpressionEvaluator struct {
	valueHistory []uint32 // History of evaluated values
	valueNumber  int      // Current value number for $1, $2, etc.
}

// NewExpressionEvaluator creates a new expression evaluator
func NewExpressionEvaluator() *ExpressionEvaluator {
	return &ExpressionEvaluator{
		valueHistory: make([]uint32, 0),
		valueNumber:  0,
	}
}

// EvaluateExpression evaluates an expression and returns the result
func (e *ExpressionEvaluator) EvaluateExpression(expr string, machine *vm.VM, symbols map[string]uint32) (uint32, error) {
	result, err := e.evaluate(expr, machine, symbols)
	if err != nil {
		return 0, err
	}

	// Store in history
	e.valueHistory = append(e.valueHistory, result)
	e.valueNumber = len(e.valueHistory)

	return result, nil
}

// Evaluate evaluates an expression and returns a boolean result (for conditions)
func (e *ExpressionEvaluator) Evaluate(expr string, machine *vm.VM, symbols map[string]uint32) (bool, error) {
	result, err := e.evaluate(expr, machine, symbols)
	if err != nil {
		return false, err
	}

	return result != 0, nil
}

// GetValueNumber returns the current value number
func (e *ExpressionEvaluator) GetValueNumber() int {
	return e.valueNumber
}

// GetValue returns a value from history by number
func (e *ExpressionEvaluator) GetValue(number int) (uint32, error) {
	if number < 1 || number > len(e.valueHistory) {
		return 0, fmt.Errorf("value $%d not in history", number)
	}

	return e.valueHistory[number-1], nil
}

// evaluate is the main evaluation logic
func (e *ExpressionEvaluator) evaluate(expr string, machine *vm.VM, symbols map[string]uint32) (uint32, error) {
	expr = strings.TrimSpace(expr)

	// Handle empty expression
	if expr == "" {
		return 0, fmt.Errorf("empty expression")
	}

	// Try to evaluate as simple atom first
	if val, err := e.trySimpleEval(expr, machine, symbols); err == nil {
		return val, nil
	}

	// Handle binary operations (simplified parser)
	// Support: +, -, *, /, &, |, ^, <<, >>
	// Look for operators with whitespace around them to avoid matching inside hex literals
	operators := []string{"<<", ">>", "&", "|", "^", "+", "-", "*", "/"}
	for _, op := range operators {
		// Look for operator with at least one space before or after
		// This ensures we don't match inside hex numbers like "0xFF"
		patterns := []string{
			" " + op + " ", // spaces on both sides
			" " + op,       // space before only
			op + " ",       // space after only
		}

		for _, pattern := range patterns {
			idx := strings.Index(expr, pattern)
			if idx < 0 {
				continue
			}

			// Calculate actual operator position
			opPos := idx
			if pattern[0] == ' ' {
				opPos++
			}

			left := strings.TrimSpace(expr[:opPos])
			right := strings.TrimSpace(expr[opPos+len(op):])

			// Skip empty left or right
			if left == "" || right == "" {
				continue
			}

			leftVal, err := e.evaluate(left, machine, symbols)
			if err != nil {
				continue // Try next pattern
			}

			rightVal, err := e.evaluate(right, machine, symbols)
			if err != nil {
				continue // Try next pattern
			}

			return e.applyOperator(leftVal, rightVal, op)
		}
	}

	return 0, fmt.Errorf("invalid expression: %s", expr)
}

// trySimpleEval tries to evaluate a simple expression (number, register, memory, symbol)
func (e *ExpressionEvaluator) trySimpleEval(expr string, machine *vm.VM, symbols map[string]uint32) (uint32, error) {
	expr = strings.TrimSpace(expr)

	// Check for memory dereference [addr] or *addr
	if strings.HasPrefix(expr, "[") && strings.HasSuffix(expr, "]") {
		addrExpr := strings.TrimSpace(expr[1 : len(expr)-1])
		addr, err := e.evaluate(addrExpr, machine, symbols)
		if err != nil {
			return 0, err
		}

		value, err := machine.Memory.ReadWord(addr)
		if err != nil {
			return 0, fmt.Errorf("failed to read memory at 0x%08X: %w", addr, err)
		}

		return value, nil
	}

	if strings.HasPrefix(expr, "*") {
		addrExpr := strings.TrimSpace(expr[1:])
		addr, err := e.evaluate(addrExpr, machine, symbols)
		if err != nil {
			return 0, err
		}

		value, err := machine.Memory.ReadWord(addr)
		if err != nil {
			return 0, fmt.Errorf("failed to read memory at 0x%08X: %w", addr, err)
		}

		return value, nil
	}

	// Check for value history reference ($1, $2, etc.)
	if strings.HasPrefix(expr, "$") {
		numStr := expr[1:]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid value reference: %s", expr)
		}

		return e.GetValue(num)
	}

	// Check for register
	if val, err := e.evalRegister(expr, machine); err == nil {
		return val, nil
	}

	// Check for symbol
	if addr, exists := symbols[expr]; exists {
		return addr, nil
	}

	// Try to parse as number
	if val, err := e.parseNumber(expr); err == nil {
		return val, nil
	}

	return 0, fmt.Errorf("unknown identifier: %s", expr)
}

// evalRegister evaluates a register reference
func (e *ExpressionEvaluator) evalRegister(expr string, machine *vm.VM) (uint32, error) {
	expr = strings.ToLower(expr)

	// Special registers
	switch expr {
	case "pc", "r15":
		return machine.CPU.PC, nil
	case "sp", "r13":
		return machine.CPU.GetSP(), nil
	case "lr", "r14":
		return machine.CPU.GetLR(), nil
	}

	// General registers
	if strings.HasPrefix(expr, "r") {
		var regNum int
		_, err := fmt.Sscanf(expr, "r%d", &regNum)
		if err == nil && regNum >= 0 && regNum <= 14 {
			return machine.CPU.R[regNum], nil
		}
	}

	return 0, fmt.Errorf("not a register")
}

// parseNumber parses a numeric literal
func (e *ExpressionEvaluator) parseNumber(expr string) (uint32, error) {
	expr = strings.TrimSpace(expr)

	// Hexadecimal
	if strings.HasPrefix(strings.ToLower(expr), "0x") {
		var val uint32
		_, err := fmt.Sscanf(strings.ToLower(expr), "0x%x", &val)
		if err != nil {
			return 0, err
		}
		return val, nil
	}

	// Binary
	if strings.HasPrefix(expr, "0b") || strings.HasPrefix(expr, "0B") {
		val, err := strconv.ParseUint(expr[2:], 2, 32)
		if err != nil {
			return 0, err
		}
		return uint32(val), nil
	}

	// Octal
	if strings.HasPrefix(expr, "0") && len(expr) > 1 {
		val, err := strconv.ParseUint(expr, 8, 32)
		if err != nil {
			return 0, err
		}
		return uint32(val), nil
	}

	// Decimal (including negative)
	val, err := strconv.ParseInt(expr, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(val), nil
}

// applyOperator applies a binary operator to two values
func (e *ExpressionEvaluator) applyOperator(left, right uint32, op string) (uint32, error) {
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

// Reset clears the value history
func (e *ExpressionEvaluator) Reset() {
	e.valueHistory = e.valueHistory[:0]
	e.valueNumber = 0
}
