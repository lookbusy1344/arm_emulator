package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Instruction represents a parsed ARM instruction
type Instruction struct {
	Label      string
	Mnemonic   string
	Condition  string // EQ, NE, CS, etc.
	SetFlags   bool   // S bit
	Operands   []string
	Comment    string
	Pos        Position
	RawLine    string
	EncodedLen int // Length in bytes (4 for ARM instructions)
}

// Directive represents an assembler directive
type Directive struct {
	Name    string
	Args    []string
	Pos     Position
	RawLine string
	Label   string // Optional label before directive
	Comment string
}

// Program represents a parsed assembly program
type Program struct {
	Instructions []*Instruction
	Directives   []*Directive
	SymbolTable  *SymbolTable
	MacroTable   *MacroTable
	Origin       uint32 // Current assembly address (.org)
}

// Parser parses ARM assembly language
type Parser struct {
	lexer          *Lexer
	tokens         []Token
	pos            int
	currentToken   Token
	peekToken      Token
	errors         *ErrorList
	symbolTable    *SymbolTable
	macroTable     *MacroTable
	numericLabels  *NumericLabelTable
	macroExpander  *MacroExpander
	preprocessor   *Preprocessor
	currentAddress uint32
}

// NewParser creates a new parser
func NewParser(input, filename string) *Parser {
	lexer := NewLexer(input, filename)
	p := &Parser{
		lexer:          lexer,
		tokens:         make([]Token, 0),
		pos:            0,
		errors:         &ErrorList{},
		symbolTable:    NewSymbolTable(),
		macroTable:     NewMacroTable(),
		numericLabels:  NewNumericLabelTable(),
		currentAddress: 0,
	}
	p.macroExpander = NewMacroExpander(p.macroTable)
	p.preprocessor = NewPreprocessor("")

	// Tokenize all input
	p.tokens = lexer.TokenizeAll()

	// Merge lexer errors
	for _, err := range lexer.Errors().Errors {
		p.errors.AddError(err)
	}

	// Initialize current and peek tokens
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken advances to the next token
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	if p.pos < len(p.tokens) {
		p.peekToken = p.tokens[p.pos]
		p.pos++
	} else {
		p.peekToken = Token{Type: TokenEOF, Literal: "", Pos: p.currentToken.Pos}
	}
}

// expectToken checks if current token matches expected type and advances
func (p *Parser) expectToken(t TokenType) bool {
	if p.currentToken.Type == t {
		p.nextToken()
		return true
	}
	p.errors.AddError(NewError(
		p.currentToken.Pos,
		ErrorSyntax,
		fmt.Sprintf("expected %s, got %s", t, p.currentToken.Type),
	))
	return false
}

// skipNewlines skips newline and comment tokens
func (p *Parser) skipNewlines() {
	for p.currentToken.Type == TokenNewline || p.currentToken.Type == TokenComment {
		p.nextToken()
	}
}

// Parse parses the entire program (two-pass assembly)
func (p *Parser) Parse() (*Program, error) {
	program := &Program{
		Instructions: make([]*Instruction, 0),
		Directives:   make([]*Directive, 0),
		SymbolTable:  p.symbolTable,
		MacroTable:   p.macroTable,
		Origin:       0,
	}

	// First pass: collect labels and directives
	err := p.firstPass(program)
	if err != nil {
		return nil, err
	}

	// Check for errors before second pass
	if p.errors.HasErrors() {
		return nil, p.errors
	}

	// Resolve forward references
	if err := p.symbolTable.ResolveForwardReferences(); err != nil {
		return nil, err
	}

	// Second pass: generate final instructions (would happen during execution)
	// For now, we've collected all the parsed instructions

	if p.errors.HasErrors() {
		return nil, p.errors
	}

	return program, nil
}

// firstPass performs the first pass of two-pass assembly
func (p *Parser) firstPass(program *Program) error {
	p.currentAddress = 0

	for p.currentToken.Type != TokenEOF {
		p.skipNewlines()

		if p.currentToken.Type == TokenEOF {
			break
		}

		// Check for label at start of line
		var label string
		if p.currentToken.Type == TokenIdentifier && p.peekToken.Type == TokenColon {
			label = p.currentToken.Literal
			p.nextToken() // consume identifier
			p.nextToken() // consume colon

			// Define label in symbol table at current address
			err := p.symbolTable.Define(label, SymbolLabel, p.currentAddress, p.currentToken.Pos)
			if err != nil {
				p.errors.AddError(NewError(p.currentToken.Pos, ErrorDuplicateLabel, err.Error()))
			}

			// Skip whitespace/newlines after label
			p.skipNewlines()
		}

		// After processing label, check what comes next
		if p.currentToken.Type == TokenEOF {
			break
		}

		// Check for directive
		if p.currentToken.Type == TokenDirective {
			directive := p.parseDirective()
			if directive != nil {
				directive.Label = label
				program.Directives = append(program.Directives, directive)
				p.handleDirective(directive)
			}
		} else if p.currentToken.Type == TokenIdentifier {
			// Parse instruction
			inst := p.parseInstruction()
			if inst != nil {
				inst.Label = label
				inst.EncodedLen = 4 // ARM instructions are 4 bytes
				program.Instructions = append(program.Instructions, inst)
				p.currentAddress += uint32(inst.EncodedLen)
			}
		} else if p.currentToken.Type != TokenNewline && p.currentToken.Type != TokenComment {
			// Skip unknown tokens (but not newlines/comments)
			p.errors.AddError(NewError(
				p.currentToken.Pos,
				ErrorSyntax,
				fmt.Sprintf("unexpected token: %s", p.currentToken.Type),
			))
			p.nextToken()
		}

		p.skipNewlines()
	}

	return nil
}

// parseDirective parses an assembler directive
func (p *Parser) parseDirective() *Directive {
	directive := &Directive{
		Name:    p.currentToken.Literal,
		Args:    make([]string, 0),
		Pos:     p.currentToken.Pos,
		RawLine: "",
	}

	p.nextToken() // consume directive name

	// Parse arguments
	for p.currentToken.Type != TokenNewline && p.currentToken.Type != TokenEOF && p.currentToken.Type != TokenComment {
		if p.currentToken.Type == TokenComma {
			p.nextToken()
			continue
		}

		arg := p.currentToken.Literal
		directive.Args = append(directive.Args, arg)
		p.nextToken()
	}

	// Consume comment if present
	if p.currentToken.Type == TokenComment {
		directive.Comment = p.currentToken.Literal
		p.nextToken()
	}

	return directive
}

// handleDirective processes directives that affect assembly state
func (p *Parser) handleDirective(d *Directive) {
	switch d.Name {
	case ".org":
		// Set origin address
		if len(d.Args) > 0 {
			if addr, err := parseNumber(d.Args[0]); err == nil {
				p.currentAddress = addr
			} else {
				p.errors.AddError(NewError(d.Pos, ErrorSyntax, fmt.Sprintf("invalid .org address: %s", d.Args[0])))
			}
		}

	case ".equ", ".set":
		// Define constant
		if len(d.Args) >= 2 {
			name := d.Args[0]
			if value, err := parseNumber(d.Args[1]); err == nil {
				if err := p.symbolTable.Define(name, SymbolConstant, value, d.Pos); err != nil {
					p.errors.AddError(NewError(d.Pos, ErrorDuplicateLabel, err.Error()))
				}
			} else {
				p.errors.AddError(NewError(d.Pos, ErrorSyntax, fmt.Sprintf("invalid constant value: %s", d.Args[1])))
			}
		}

	case ".word":
		// Reserve 4 bytes per word
		p.currentAddress += uint32(len(d.Args) * 4)

	case ".half":
		// Reserve 2 bytes per halfword
		p.currentAddress += uint32(len(d.Args) * 2)

	case ".byte":
		// Reserve 1 byte per byte
		p.currentAddress += uint32(len(d.Args))

	case ".ascii", ".asciz", ".string":
		// Reserve bytes for string
		if len(d.Args) > 0 {
			str := d.Args[0]
			// Remove quotes
			if len(str) >= 2 && (str[0] == '"' || str[0] == '\'') {
				str = str[1 : len(str)-1]
			}
			p.currentAddress += uint32(len(str))
			if d.Name == ".asciz" || d.Name == ".string" {
				p.currentAddress++ // Null terminator
			}
		}

	case ".space", ".skip":
		// Reserve specified number of bytes
		if len(d.Args) > 0 {
			if size, err := parseNumber(d.Args[0]); err == nil {
				p.currentAddress += size
			}
		}

	case ".align":
		// Align to power of 2
		if len(d.Args) > 0 {
			if align, err := parseNumber(d.Args[0]); err == nil {
				mask := align - 1
				p.currentAddress = (p.currentAddress + mask) & ^mask
			}
		}

	case ".balign":
		// Align to specified boundary
		if len(d.Args) > 0 {
			if align, err := parseNumber(d.Args[0]); err == nil {
				if p.currentAddress%align != 0 {
					p.currentAddress += align - (p.currentAddress % align)
				}
			}
		}
	}
}

// parseInstruction parses an ARM instruction
func (p *Parser) parseInstruction() *Instruction {
	inst := &Instruction{
		Mnemonic: p.currentToken.Literal,
		Operands: make([]string, 0),
		Pos:      p.currentToken.Pos,
		RawLine:  "",
	}

	// Parse mnemonic (might include condition and S flag)
	mnemonic := strings.ToUpper(p.currentToken.Literal)

	// Check for condition code suffix
	inst.Condition, inst.SetFlags, inst.Mnemonic = parseInstructionMnemonic(mnemonic)

	p.nextToken() // consume mnemonic

	// Parse operands
	for p.currentToken.Type != TokenNewline && p.currentToken.Type != TokenEOF && p.currentToken.Type != TokenComment {
		operand := p.parseOperand()
		if operand != "" {
			inst.Operands = append(inst.Operands, operand)
		}

		if p.currentToken.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Consume comment if present
	if p.currentToken.Type == TokenComment {
		inst.Comment = p.currentToken.Literal
		p.nextToken()
	}

	return inst
}

// parseOperand parses a single operand
func (p *Parser) parseOperand() string {
	var parts []string

	// Handle different operand types
	switch p.currentToken.Type {
	case TokenHash:
		// Immediate value: #123
		parts = append(parts, "#")
		p.nextToken()
		if p.currentToken.Type == TokenNumber || p.currentToken.Type == TokenIdentifier || p.currentToken.Type == TokenMinus {
			if p.currentToken.Type == TokenMinus {
				parts = append(parts, "-")
				p.nextToken()
			}
			parts = append(parts, p.currentToken.Literal)
			p.nextToken()
		}

	case TokenLBracket:
		// Memory address: [Rn], [Rn, #offset], etc.
		parts = append(parts, "[")
		p.nextToken()

		for p.currentToken.Type != TokenRBracket && p.currentToken.Type != TokenNewline && p.currentToken.Type != TokenEOF {
			if p.currentToken.Type == TokenComma {
				parts = append(parts, ",")
			} else {
				parts = append(parts, p.currentToken.Literal)
			}
			p.nextToken()
		}

		if p.currentToken.Type == TokenRBracket {
			parts = append(parts, "]")
			p.nextToken()

			// Check for post-indexed addressing: ]!
			if p.currentToken.Type == TokenExclaim {
				parts = append(parts, "!")
				p.nextToken()
			}
		}

	case TokenRegister, TokenIdentifier, TokenNumber:
		// Register or label
		parts = append(parts, p.currentToken.Literal)
		p.nextToken()

		// Check for shift operations: Rm, LSL #shift
		if p.currentToken.Type == TokenComma {
			// Peek ahead to see if this is a shift
			if p.peekToken.Type == TokenIdentifier {
				shiftOp := strings.ToUpper(p.peekToken.Literal)
				if shiftOp == "LSL" || shiftOp == "LSR" || shiftOp == "ASR" || shiftOp == "ROR" || shiftOp == "RRX" {
					parts = append(parts, ",")
					p.nextToken() // consume comma
					parts = append(parts, p.currentToken.Literal)
					p.nextToken() // consume shift op

					// Parse shift amount
					if p.currentToken.Type == TokenHash {
						parts = append(parts, "#")
						p.nextToken()
						parts = append(parts, p.currentToken.Literal)
						p.nextToken()
					} else if p.currentToken.Type == TokenRegister {
						parts = append(parts, p.currentToken.Literal)
						p.nextToken()
					}
					return strings.Join(parts, " ")
				}
			}
		}

	default:
		parts = append(parts, p.currentToken.Literal)
		p.nextToken()
	}

	return strings.Join(parts, " ")
}

// parseInstructionMnemonic parses the mnemonic and extracts condition and S flag
func parseInstructionMnemonic(mnemonic string) (condition string, setFlags bool, baseMnemonic string) {
	// List of condition codes
	conditions := []string{"EQ", "NE", "CS", "HS", "CC", "LO", "MI", "PL", "VS", "VC", "HI", "LS", "GE", "LT", "GT", "LE", "AL"}

	// Check if mnemonic ends with S (set flags)
	if len(mnemonic) > 1 && mnemonic[len(mnemonic)-1] == 'S' {
		// Check if it's really a set flags indicator (not part of the instruction like BLS)
		possibleBase := mnemonic[:len(mnemonic)-1]
		if !isInstructionName(mnemonic) && isInstructionName(possibleBase) {
			setFlags = true
			mnemonic = possibleBase
		}
	}

	// Check for condition code
	for _, cond := range conditions {
		if strings.HasSuffix(mnemonic, cond) {
			possibleBase := mnemonic[:len(mnemonic)-len(cond)]
			if isInstructionName(possibleBase) {
				condition = cond
				baseMnemonic = possibleBase
				return
			}
		}
	}

	return "", setFlags, mnemonic
}

// isInstructionName checks if a string is a valid instruction name
func isInstructionName(s string) bool {
	instructions := []string{
		"MOV", "MVN", "ADD", "ADC", "SUB", "SBC", "RSB", "RSC",
		"AND", "ORR", "EOR", "BIC", "CMP", "CMN", "TST", "TEQ",
		"LDR", "STR", "LDRB", "STRB", "LDRH", "STRH",
		"LDM", "STM", "LDMIA", "LDMIB", "LDMDA", "LDMDB",
		"STMIA", "STMIB", "STMDA", "STMDB",
		"PUSH", "POP",
		"B", "BL", "BX",
		"MUL", "MLA",
		"SWI",
	}

	for _, inst := range instructions {
		if s == inst {
			return true
		}
	}
	return false
}

// parseNumber parses a number in various formats (decimal, hex, binary, octal)
func parseNumber(s string) (uint32, error) {
	s = strings.TrimSpace(s)

	// Handle negative numbers
	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	var value uint64
	var err error

	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		// Hexadecimal
		value, err = strconv.ParseUint(s[2:], 16, 32)
	} else if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") {
		// Binary
		value, err = strconv.ParseUint(s[2:], 2, 32)
	} else if strings.HasPrefix(s, "0o") || strings.HasPrefix(s, "0O") {
		// Octal
		value, err = strconv.ParseUint(s[2:], 8, 32)
	} else {
		// Decimal
		value, err = strconv.ParseUint(s, 10, 32)
	}

	if err != nil {
		return 0, err
	}

	result := uint32(value)
	if negative {
		result = uint32(-int32(result))
	}

	return result, nil
}

// Errors returns the error list
func (p *Parser) Errors() *ErrorList {
	return p.errors
}
