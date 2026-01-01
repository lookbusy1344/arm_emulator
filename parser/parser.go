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
	Instructions       []*Instruction
	Directives         []*Directive
	SymbolTable        *SymbolTable
	MacroTable         *MacroTable
	Origin             uint32         // Current assembly address (.org)
	OriginSet          bool           // Whether .org directive was explicitly used
	LiteralPoolLocs    []uint32       // Addresses where .ltorg directives appear
	LiteralPoolCounts  []int          // Number of unique literals needed for each pool
	LiteralPoolIndices map[uint32]int // Maps pool address to index in LiteralPoolCounts
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
	originSet      bool     // Track if .org directive has been encountered
	inputLines     []string // Cached split lines for getRawLineFromInput
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
		Instructions:       make([]*Instruction, 0),
		Directives:         make([]*Directive, 0),
		SymbolTable:        p.symbolTable,
		MacroTable:         p.macroTable,
		Origin:             0,
		LiteralPoolIndices: make(map[uint32]int),
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

	// Count literals for each pool location
	if len(program.LiteralPoolLocs) > 0 {
		p.countLiteralsPerPool(program)
	}

	// Adjust addresses after calculating actual literal pool needs
	// This is needed because we might have reserved more space than necessary
	if len(program.LiteralPoolLocs) > 0 {
		p.adjustAddressesForDynamicPools(program)
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
				directive.Address = p.currentAddress                          // Record address before processing
				directive.RawLine = p.getRawLineFromInput(directive.Pos.Line) // Capture raw source line
				program.Directives = append(program.Directives, directive)
				p.handleDirective(directive, program)
			}
		} else if p.currentToken.Type == TokenIdentifier {
			// Parse instruction
			inst := p.parseInstruction()
			if inst != nil {
				inst.Label = label
				inst.EncodedLen = 4                                 // ARM instructions are 4 bytes
				inst.Address = p.currentAddress                     // Record address
				inst.RawLine = p.getRawLineFromInput(inst.Pos.Line) // Capture raw source line
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
		} else if p.currentToken.Type == TokenString {
			// Preserve quotes for character literals
			arg = "'" + p.currentToken.Literal + "'"
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
		// Reserve bytes for string (use processed length to account for escape sequences)
		if len(d.Args) > 0 {
			str := d.Args[0]
			// Remove quotes
			if len(str) >= 2 && (str[0] == '"' || str[0] == '\'') {
				str = str[1 : len(str)-1]
			}
			// Process escape sequences to get actual byte count (e.g., "\n" = 1 byte, "\x41" = 1 byte)
			processedStr := ProcessEscapeSequences(str)
			// Safe: string length is limited by available memory
			p.currentAddress += uint32(len(processedStr)) // #nosec G115 -- reasonable string length
			if d.Name == ".asciz" || d.Name == ".string" {
				p.currentAddress++ // Null terminator
			}
		}

	case ".space", ".skip":
		// Reserve specified number of bytes
		if len(d.Args) > 0 {
			var size uint32
			var err error

			// Try to parse as number first
			size, err = parseNumber(d.Args[0])
			if err != nil {
				// If not a number, try to resolve as symbol (e.g., .equ constant)
				size, err = p.symbolTable.Get(d.Args[0])
				if err != nil {
					p.errors.AddError(NewError(d.Pos, ErrorInvalidOperand,
						fmt.Sprintf("invalid size for .space: %s", d.Args[0])))
					return
				}
			}
			p.currentAddress += size
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

	case ".ltorg":
		// Literal pool directive - mark this location for literal pool emission
		// Align to 4-byte boundary
		if p.currentAddress%4 != 0 {
			p.currentAddress += 4 - (p.currentAddress % 4)
		}
		// Record this location as a literal pool position
		program.LiteralPoolLocs = append(program.LiteralPoolLocs, p.currentAddress)

		// Reserve space for the literal pool
		// We reserve a reasonable fixed amount (16 literals = 64 bytes) for each .ltorg
		// This is a conservative estimate that handles typical usage while not being excessive
		// The encoder will place actual literals within this space
		p.currentAddress += EstimatedLiteralsPerPool * 4
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

// parseOperand parses a single operand by dispatching to type-specific parsers
func (p *Parser) parseOperand() string {
	switch p.currentToken.Type {
	case TokenHash:
		return p.parseImmediateOperand()
	case TokenLBracket:
		return p.parseMemoryOperand()
	case TokenLBrace:
		return p.parseRegisterListOperand()
	case TokenEqual:
		return p.parsePseudoOperand()
	case TokenRegister, TokenIdentifier, TokenNumber:
		return p.parseRegisterOrLabelOperand()
	default:
		lit := p.currentToken.Literal
		p.nextToken()
		return lit
	}
}

// parseImmediateOperand parses immediate values: #123, #-45, #'A'
func (p *Parser) parseImmediateOperand() string {
	var parts []string
	parts = append(parts, "#")
	p.nextToken()

	if p.currentToken.Type == TokenNumber || p.currentToken.Type == TokenIdentifier ||
		p.currentToken.Type == TokenMinus || p.currentToken.Type == TokenString {
		if p.currentToken.Type == TokenMinus {
			parts = append(parts, "-")
			p.nextToken()
		}
		if p.currentToken.Type == TokenString {
			parts = append(parts, "'"+p.currentToken.Literal+"'")
		} else {
			parts = append(parts, p.currentToken.Literal)
		}
		p.nextToken()
	}
	return strings.Join(parts, "")
}

// parseMemoryOperand parses memory addresses: [Rn], [Rn, #offset], [Rn, Rm, LSL #2]
func (p *Parser) parseMemoryOperand() string {
	var parts []string
	parts = append(parts, "[")
	p.nextToken()

	for p.currentToken.Type != TokenRBracket && p.currentToken.Type != TokenNewline && p.currentToken.Type != TokenEOF {
		switch {
		case p.currentToken.Type == TokenComma:
			parts = append(parts, ",")
		case p.currentToken.Type == TokenHash:
			parts = append(parts, " #")
		case p.currentToken.Literal != "" && strings.TrimSpace(p.currentToken.Literal) != "":
			lit := p.currentToken.Literal
			if isShiftOperator(lit) {
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
		if p.currentToken.Type == TokenExclaim {
			parts = append(parts, "!")
			p.nextToken()
		}
	}
	return strings.Join(parts, "")
}

// parseRegisterListOperand parses register lists: {R0, R1, R2}, {R0-R3}
func (p *Parser) parseRegisterListOperand() string {
	var parts []string
	parts = append(parts, "{")
	p.nextToken()

	for p.currentToken.Type != TokenRBrace && p.currentToken.Type != TokenNewline && p.currentToken.Type != TokenEOF {
		switch p.currentToken.Type {
		case TokenComma:
			parts = append(parts, ",")
		case TokenMinus:
			parts = append(parts, "-")
		default:
			parts = append(parts, p.currentToken.Literal)
		}
		p.nextToken()
	}

	if p.currentToken.Type == TokenRBrace {
		parts = append(parts, "}")
		p.nextToken()
	}
	return strings.Join(parts, "")
}

// parsePseudoOperand parses pseudo-instruction operands: =label, =value, =label+offset
func (p *Parser) parsePseudoOperand() string {
	var parts []string
	parts = append(parts, "=")
	p.nextToken()

	if p.currentToken.Type == TokenIdentifier || p.currentToken.Type == TokenNumber {
		parts = append(parts, p.currentToken.Literal)
		p.nextToken()

		// Parse arithmetic expressions: =label+12, =label-4
		for p.currentToken.Type == TokenPlus || p.currentToken.Type == TokenMinus {
			parts = append(parts, p.currentToken.Literal)
			p.nextToken()
			if p.currentToken.Type == TokenNumber || p.currentToken.Type == TokenIdentifier {
				parts = append(parts, p.currentToken.Literal)
				p.nextToken()
			} else {
				break
			}
		}
	}
	return strings.Join(parts, "")
}

// parseRegisterOrLabelOperand parses registers, labels, and shifted registers: R0, label, R0!, R0,LSL #2
func (p *Parser) parseRegisterOrLabelOperand() string {
	base := p.currentToken.Literal
	p.nextToken()

	// Writeback: R13! or SP!
	if p.currentToken.Type == TokenExclaim {
		p.nextToken()
		return base + "!"
	}

	// Shifted register: Rm, LSL #shift or Rm, LSL Rs
	if p.currentToken.Type == TokenComma && p.peekToken.Type == TokenIdentifier {
		if isShiftOperator(p.peekToken.Literal) {
			p.nextToken() // consume comma
			shiftOp := p.currentToken.Literal
			p.nextToken() // consume shift operator

			var shiftAmount string
			if p.currentToken.Type == TokenHash {
				p.nextToken()
				shiftAmount = " #" + p.currentToken.Literal
				p.nextToken()
			} else if p.currentToken.Type == TokenRegister {
				shiftAmount = " " + p.currentToken.Literal
				p.nextToken()
			}
			return base + "," + shiftOp + shiftAmount
		}
	}

	return base
}

// isShiftOperator returns true if the given string is a shift operator
func isShiftOperator(s string) bool {
	switch strings.ToUpper(s) {
	case "LSL", "LSR", "ASR", "ROR", "RRX":
		return true
	}
	return false
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
		"PUSH", "POP", "NOP",
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
		// Allow up to 2147483648 (which represents MinInt32 = -2147483648)
		// MaxInt32 is 2147483647, so we allow MaxInt32+1 for the MinInt32 case
		if result > uint32(math.MaxInt32)+1 {
			return 0, fmt.Errorf("negative value -%d is out of range for int32", result)
		}
		// Safe: bounds checked above to be within int32 range (including MinInt32)
		result = uint32(-int32(result)) // #nosec G115 -- bounds checked
	}

	return result, nil
}

// adjustAddressesForDynamicPools adjusts addresses after determining actual literal pool sizes
// This function updates the addresses of all items after each pool to reflect the actual
// space needed (based on literal counts) rather than the fixed 16-literal estimate
func (p *Parser) adjustAddressesForDynamicPools(program *Program) {
	if len(program.LiteralPoolLocs) == 0 || len(program.LiteralPoolCounts) == 0 {
		return
	}

	// Calculate the difference between estimated and actual space for each pool
	// Estimated: N literals = N*4 bytes
	// Actual: LiteralPoolCounts[i] literals = LiteralPoolCounts[i] * 4 bytes
	estimatedBytes := EstimatedLiteralsPerPool * 4

	// Track cumulative address offset due to differences
	cumulativeOffset := int32(0)

	// Store original pool locations before adjustments
	originalPoolLocs := make([]uint32, len(program.LiteralPoolLocs))
	copy(originalPoolLocs, program.LiteralPoolLocs)

	// Store cumulative offset at each pool location for later address adjustments
	offsetAtPool := make([]int32, len(program.LiteralPoolLocs))

	for i, poolLoc := range program.LiteralPoolLocs {
		// Calculate actual space needed for this pool
		actualCount := program.LiteralPoolCounts[i]
		actualBytes := actualCount * 4

		// Difference (can be negative if we reserved more space than needed)
		difference := actualBytes - estimatedBytes

		// Update pool location with cumulative offset
		// #nosec G115 -- poolLoc and cumulativeOffset are within reasonable bounds for address calculations
		program.LiteralPoolLocs[i] = uint32(int32(poolLoc) + cumulativeOffset)

		// Add this pool's difference to cumulative offset for subsequent items
		// #nosec G115 -- difference is bounded by pool size estimates, safe to convert to int32
		cumulativeOffset += int32(difference)

		// Store cumulative offset after this pool
		offsetAtPool[i] = cumulativeOffset
	}

	// Now update all instruction and symbol addresses that come after pools
	// IMPORTANT: We must calculate all adjustments first, THEN apply them
	// Otherwise we'd be comparing already-adjusted addresses against original pool locations
	if len(originalPoolLocs) > 0 {
		// Helper function to determine adjustment for a given ORIGINAL address
		getAdjustmentForAddress := func(addr uint32) int32 {
			// Find the last pool that this address comes after
			for i := len(originalPoolLocs) - 1; i >= 0; i-- {
				// #nosec G115 -- estimatedBytes is a const 64, safe to convert to uint32
				poolEndLoc := originalPoolLocs[i] + uint32(estimatedBytes)
				if addr >= poolEndLoc {
					// This address comes after pool i, use cumulative offset at pool i
					return offsetAtPool[i]
				}
			}
			// Address comes before all pools, no adjustment needed
			return 0
		}

		// First pass: calculate adjustments for all instructions
		instAdjustments := make([]int32, len(program.Instructions))
		for i, inst := range program.Instructions {
			instAdjustments[i] = getAdjustmentForAddress(inst.Address)
		}

		// Second pass: apply adjustments
		for i, adjustment := range instAdjustments {
			if adjustment != 0 {
				// #nosec G115 -- adjustment is bounded, safe conversion
				program.Instructions[i].Address = uint32(int32(program.Instructions[i].Address) + adjustment)
			}
		}

		// Same for symbols
		symbolAdjustments := make(map[string]int32)
		for name, symbol := range program.SymbolTable.symbols {
			adjustmentOffset := getAdjustmentForAddress(symbol.Value)
			if adjustmentOffset != 0 {
				symbolAdjustments[name] = adjustmentOffset
			}
		}

		for name, adjustment := range symbolAdjustments {
			symbol := program.SymbolTable.symbols[name]
			// #nosec G115 -- adjustment is bounded, safe conversion
			symbol.Value = uint32(int32(symbol.Value) + adjustment)
			program.SymbolTable.symbols[name] = symbol
		}
	}
}

// countLiteralsPerPool counts the number of unique literals needed for each literal pool
// This dynamically reserves space based on actual literal usage rather than a fixed estimate
func (p *Parser) countLiteralsPerPool(program *Program) {
	if len(program.LiteralPoolLocs) == 0 {
		return
	}

	// Initialize map for quick lookup of pool indices
	for i, poolLoc := range program.LiteralPoolLocs {
		program.LiteralPoolIndices[poolLoc] = i
	}

	// Scan all instructions to count LDR pseudo-instructions for each pool
	// An LDR Rd, =value instruction needs one literal pool entry
	literalsBeforePool := make(map[int]map[string]bool) // pool index -> set of unique literal expressions

	// Initialize literal tracking for each pool
	for i := range program.LiteralPoolLocs {
		literalsBeforePool[i] = make(map[string]bool)
	}

	// Track which pool each instruction belongs to
	for _, inst := range program.Instructions {
		if inst.Mnemonic == "LDR" && len(inst.Operands) >= 2 {
			operand := strings.TrimSpace(inst.Operands[1])
			if strings.HasPrefix(operand, "=") {
				// This is a pseudo-instruction that needs a literal pool entry
				// Find the nearest pool location AFTER this instruction
				poolIdx := -1
				for i, poolLoc := range program.LiteralPoolLocs {
					if poolLoc > inst.Address {
						poolIdx = i
						break
					}
				}

				// If no pool found after instruction, use the last pool
				if poolIdx == -1 && len(program.LiteralPoolLocs) > 0 {
					poolIdx = len(program.LiteralPoolLocs) - 1
				}

				// Record that this pool will need this literal
				if poolIdx >= 0 {
					// Add to the set of unique literals for this pool
					// Use the literal expression string as key for deduplication
					// (e.g., "=0x1234" used twice only needs one pool entry)
					literalExpr := operand[1:] // Remove the '=' prefix
					literalsBeforePool[poolIdx][literalExpr] = true
				}
			}
		}
	}

	// Calculate actual literal counts per pool
	program.LiteralPoolCounts = make([]int, len(program.LiteralPoolLocs))
	for i := range program.LiteralPoolLocs {
		program.LiteralPoolCounts[i] = len(literalsBeforePool[i])
	}
}

// getRawLineFromInput extracts the raw source line for a given line number
func (p *Parser) getRawLineFromInput(lineNum int) string {
	if p.lexer == nil || p.lexer.input == "" {
		return ""
	}

	// Cache split lines on first access
	if p.inputLines == nil {
		p.inputLines = strings.Split(p.lexer.input, "\n")
	}

	if lineNum < 1 || lineNum > len(p.inputLines) {
		return ""
	}

	// Line numbers are 1-based
	return p.inputLines[lineNum-1]
}

// Errors returns the error list
func (p *Parser) Errors() *ErrorList {
	return p.errors
}
