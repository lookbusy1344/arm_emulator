package parser

import (
	"fmt"
	"math"
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
	EncodedLen int    // Length in bytes (4 for ARM instructions)
	Address    uint32 // Address where this instruction should be placed
}

// Directive represents an assembler directive
type Directive struct {
	Name    string
	Args    []string
	Pos     Position
	RawLine string
	Label   string // Optional label before directive
	Comment string
	Address uint32 // Address where this directive's data should be placed
}

// Program represents a parsed assembly program
type Program struct {
	Instructions []*Instruction
	Directives   []*Directive
	SymbolTable  *SymbolTable
	MacroTable   *MacroTable
	Origin       uint32 // Current assembly address (.org)
	OriginSet    bool   // Whether .org directive was explicitly used
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
	originSet      bool // Track if .org directive has been encountered
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

			// NOTE: Don't skip newlines here. The lexer has already skipped horizontal whitespace.
			// If there's a directive/instruction on the same line, it will be in currentToken.
			// If the label is standalone, currentToken will be a newline, which will be handled
			// by the skipNewlines() at the end of the loop. This prevents the bug where standalone
			// labels cause the next line's label to be consumed as an instruction mnemonic.
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
				directive.Address = p.currentAddress // Record address before processing
				program.Directives = append(program.Directives, directive)
				p.handleDirective(directive, program)
			}
		} else if p.currentToken.Type == TokenIdentifier {
			// Parse instruction
			inst := p.parseInstruction()
			if inst != nil {
				inst.Label = label
				inst.EncodedLen = 4             // ARM instructions are 4 bytes
				inst.Address = p.currentAddress // Record address
				program.Instructions = append(program.Instructions, inst)
				// Safe: EncodedLen is always 4 for ARM instructions
				p.currentAddress += uint32(inst.EncodedLen) // #nosec G115 -- EncodedLen is always 4
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

		// Handle negative numbers by combining minus sign with following number
		if p.currentToken.Type == TokenMinus && p.peekToken.Type == TokenNumber {
			p.nextToken() // consume minus
			arg = "-" + p.currentToken.Literal
		}

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
func (p *Parser) handleDirective(d *Directive, program *Program) {
	switch d.Name {
	case ".text":
		// Text section - if currentAddress is 0 and origin not set,
		// this is the first section, so set origin to 0.
		// If currentAddress > 0 (data came first), keep the current address.
		if !p.originSet {
			if p.currentAddress == 0 {
				// First section is .text
				p.currentAddress = 0
			}
			// else: data came first, keep current address which has data
			program.Origin = p.currentAddress
			program.OriginSet = true
			p.originSet = true
		}
		// If .text appears after data, currentAddress is already positioned after data

	case ".data":
		// Data section - if origin isn't set yet and currentAddress is 0,
		// set origin to 0 for data segment
		if !p.originSet && p.currentAddress == 0 {
			program.Origin = 0
			program.OriginSet = true
			p.originSet = true
		}
		// Continue at current address (interleaved with code if that's the layout)

	case ".global":
		// Global symbol declaration - mark symbol as global (exported)
		// For now, we just note it but don't need special handling in a simple emulator
		// In a linker, this would make the symbol visible to other modules

	case ".org":
		// Set origin address
		if len(d.Args) > 0 {
			if addr, err := parseNumber(d.Args[0]); err == nil {
				p.currentAddress = addr
				// Set program origin if this is the first .org directive
				if !p.originSet {
					program.Origin = addr
					program.OriginSet = true
					p.originSet = true
				}
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
		// Safe: len(d.Args) is limited by available memory, multiplication by 4 won't overflow in practice
		p.currentAddress += uint32(len(d.Args) * 4) // #nosec G115 -- reasonable argument count

	case ".half":
		// Reserve 2 bytes per halfword
		// Safe: len(d.Args) is limited by available memory, multiplication by 2 won't overflow in practice
		p.currentAddress += uint32(len(d.Args) * 2) // #nosec G115 -- reasonable argument count

	case ".byte":
		// Reserve 1 byte per byte
		// Safe: len(d.Args) is limited by available memory
		p.currentAddress += uint32(len(d.Args)) // #nosec G115 -- reasonable argument count

	case ".ascii", ".asciz", ".string":
		// Reserve bytes for string
		if len(d.Args) > 0 {
			str := d.Args[0]
			// Remove quotes
			if len(str) >= 2 && (str[0] == '"' || str[0] == '\'') {
				str = str[1 : len(str)-1]
			}
			// Safe: string length is limited by available memory
			p.currentAddress += uint32(len(str)) // #nosec G115 -- reasonable string length
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
		// Align to power of 2 (e.g., .align 2 means align to 2^2 = 4 bytes)
		if len(d.Args) > 0 {
			if alignPower, err := parseNumber(d.Args[0]); err == nil {
				alignBytes := uint32(1 << alignPower) // 2^alignPower
				mask := alignBytes - 1
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
		// Immediate value: #123 or #'A'
		parts = append(parts, "#")
		p.nextToken()
		if p.currentToken.Type == TokenNumber || p.currentToken.Type == TokenIdentifier || p.currentToken.Type == TokenMinus || p.currentToken.Type == TokenString {
			if p.currentToken.Type == TokenMinus {
				parts = append(parts, "-")
				p.nextToken()
			}
			// Handle character literals: #'A'
			if p.currentToken.Type == TokenString {
				parts = append(parts, "'"+p.currentToken.Literal+"'")
			} else {
				parts = append(parts, p.currentToken.Literal)
			}
			p.nextToken()
		}
		// Return without joining with spaces for #value
		return strings.Join(parts, "")

	case TokenLBracket:
		// Memory address: [Rn], [Rn, #offset], etc.
		parts = append(parts, "[")
		p.nextToken()

		for p.currentToken.Type != TokenRBracket && p.currentToken.Type != TokenNewline && p.currentToken.Type != TokenEOF {
			if p.currentToken.Type == TokenComma {
				parts = append(parts, ",")
			} else if p.currentToken.Type == TokenHash {
				// Add space before # for shift amounts like "LSL #2"
				parts = append(parts, " #")
			} else if p.currentToken.Literal != "" && strings.TrimSpace(p.currentToken.Literal) != "" {
				lit := p.currentToken.Literal
				// Add space before shift operators for proper parsing
				shiftOp := strings.ToUpper(lit)
				if shiftOp == "LSL" || shiftOp == "LSR" || shiftOp == "ASR" || shiftOp == "ROR" || shiftOp == "RRX" {
					parts = append(parts, " "+lit)
				} else {
					parts = append(parts, lit)
				}
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
		// Return without joining with spaces for memory addressing
		return strings.Join(parts, "")

	case TokenLBrace:
		// Register list: {R0, R1, R2} or {R0-R3}
		parts = append(parts, "{")
		p.nextToken()

		for p.currentToken.Type != TokenRBrace && p.currentToken.Type != TokenNewline && p.currentToken.Type != TokenEOF {
			if p.currentToken.Type == TokenComma {
				parts = append(parts, ",")
			} else if p.currentToken.Type == TokenMinus {
				// Handle register range: R0-R3
				parts = append(parts, "-")
			} else {
				parts = append(parts, p.currentToken.Literal)
			}
			p.nextToken()
		}

		if p.currentToken.Type == TokenRBrace {
			parts = append(parts, "}")
			p.nextToken()
		}
		// Return without joining with spaces for register lists
		return strings.Join(parts, "")

	case TokenEqual:
		// Handle =label or =value for pseudo-instructions like LDR Rd, =label
		// Now also supports constant expressions: =label + 12, =label - 4
		parts = append(parts, "=")
		p.nextToken()

		// Get the first identifier or number after =
		if p.currentToken.Type == TokenIdentifier || p.currentToken.Type == TokenNumber {
			parts = append(parts, p.currentToken.Literal)
			p.nextToken()

			// Check for arithmetic operators to form expressions
			for p.currentToken.Type == TokenPlus || p.currentToken.Type == TokenMinus {
				// Append the operator
				parts = append(parts, p.currentToken.Literal)
				p.nextToken()

				// Append the operand (number or identifier)
				if p.currentToken.Type == TokenNumber || p.currentToken.Type == TokenIdentifier {
					parts = append(parts, p.currentToken.Literal)
					p.nextToken()
				} else {
					// Invalid expression
					break
				}
			}
		}
		// Return without joining with spaces for =value
		return strings.Join(parts, "")

	case TokenRegister, TokenIdentifier, TokenNumber:
		// Register or label
		parts = append(parts, p.currentToken.Literal)
		p.nextToken()

		// Check for writeback: R13! or SP!
		if p.currentToken.Type == TokenExclaim {
			parts = append(parts, "!")
			p.nextToken()
			// Return immediately for writeback syntax
			return strings.Join(parts, "")
		}

		// Check for shift operations: Rm, LSL #shift
		if p.currentToken.Type == TokenComma {
			// Peek ahead to see if this is a shift
			if p.peekToken.Type == TokenIdentifier {
				shiftOp := strings.ToUpper(p.peekToken.Literal)
				if shiftOp == "LSL" || shiftOp == "LSR" || shiftOp == "ASR" || shiftOp == "ROR" || shiftOp == "RRX" {
					p.nextToken() // consume comma
					// Build shift operand: R0,LSL #2
					shiftPart := p.currentToken.Literal // shift op (LSL/LSR/etc)
					p.nextToken()                       // consume shift op

					// Parse shift amount
					if p.currentToken.Type == TokenHash {
						p.nextToken() // consume hash
						shiftPart += " #" + p.currentToken.Literal
						p.nextToken() // consume number
					} else if p.currentToken.Type == TokenRegister {
						shiftPart += " " + p.currentToken.Literal
						p.nextToken()
					}
					// Return as "R0,LSL #2" format (register,shift)
					return parts[0] + "," + shiftPart
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
		"LDMFD", "LDMFA", "LDMEA", "LDMED", // Load Multiple aliases (FD=Full Descending, etc.)
		"STMFD", "STMFA", "STMEA", "STMED", // Store Multiple aliases
		"PUSH", "POP",
		"B", "BL", "BX",
		"MUL", "MLA",
		"SWI", "SVC", // SVC is ARM7+ name for SWI (Supervisor Call)
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
		if result > uint32(math.MaxInt32) {
			return 0, fmt.Errorf("negative value %d is out of range for int32", result)
		}
		// Safe: bounds checked above to be within int32 range
		result = uint32(-int32(result)) // #nosec G115 -- bounds checked
	}

	return result, nil
}

// Errors returns the error list
func (p *Parser) Errors() *ErrorList {
	return p.errors
}
