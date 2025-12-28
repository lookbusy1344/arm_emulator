package debugger

import (
	"fmt"
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

// Evaluate evaluates an expression and returns a boolean result (for conditions).
// Unlike EvaluateExpression, this method does NOT store the result in value history.
// This is intentional for breakpoint condition evaluation, which should not pollute
// the user's $1, $2, etc. history references.
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

	// Use the new tokenizer and parser
	lexer := NewExprLexer(expr)
	tokens := lexer.TokenizeAll()

	// Check for lexer errors (empty token list or only EOF)
	if len(tokens) == 0 || (len(tokens) == 1 && tokens[0].Type == ExprTokenEOF) {
		return 0, fmt.Errorf("invalid expression: %s", expr)
	}

	parser := NewExprParser(tokens, machine, symbols, e)
	result, err := parser.Parse()
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return result, nil
}

// Reset clears the value history
func (e *ExpressionEvaluator) Reset() {
	e.valueHistory = e.valueHistory[:0]
	e.valueNumber = 0
}
