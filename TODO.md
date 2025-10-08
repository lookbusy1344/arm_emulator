# ARM2 Emulator TODO List

**Last Updated:** 2025-10-08

---

## High Priority

### 1. Expression Parser Improvements (Phase 5 Enhancement)

**Status:** Partially Complete - Basic expressions work, complex expressions need proper tokenization

**Problem:** The current expression evaluator uses simple string searching to find operators, which cannot properly distinguish between:
- Operators that are part of numeric literals (e.g., hex digits like 'F' in '0xFF')
- Actual binary operators between values
- Register names vs operators

**Disabled Tests:**
```go
// In debugger/expressions_test.go:

// Hex number arithmetic
{"Hex addition", "0x10 + 0x20", 0x30}

// Bitwise operations with hex numbers
{"AND", "0xFF & 0x0F", 0x0F}
{"OR", "0xF0 | 0x0F", 0xFF}
{"XOR", "0xFF ^ 0x0F", 0xF0}

// Register expression operations
{"Register addition", "r0 + r1", 30}
{"Register with constant", "r0 + 5", 15}
{"Register subtraction", "r1 - r0", 10}
```

**Current Workarounds:**
- ✅ Simple numeric literals work
- ✅ Register references work
- ✅ Symbol lookups work
- ✅ Memory dereferencing works
- ✅ Simple decimal arithmetic works (`10 + 20`, `5 * 6`)
- ✅ Shift operations work (`1 << 4`)

**Failures:**
- ❌ Arithmetic with hex numbers
- ❌ Bitwise operations with hex numbers
- ❌ Operations between registers

**Recommended Solution:**

Implement a proper two-phase expression parser:

#### Phase 1: Lexical Analysis (Tokenization)

Create a lexer in `debugger/lexer.go`:

```go
type Token struct {
    Type  TokenType
    Value string
    Pos   int
}

type TokenType int
const (
    TOKEN_NUMBER TokenType = iota
    TOKEN_HEX_NUMBER
    TOKEN_BINARY_NUMBER
    TOKEN_OCTAL_NUMBER
    TOKEN_REGISTER
    TOKEN_SYMBOL
    TOKEN_OPERATOR
    TOKEN_LPAREN
    TOKEN_RPAREN
    TOKEN_LBRACKET
    TOKEN_RBRACKET
    TOKEN_MEMORY_DEREF  // *
    TOKEN_EOF
)

type Lexer struct {
    input string
    pos   int
    tokens []Token
}

func (l *Lexer) Tokenize() []Token
func (l *Lexer) NextToken() Token
```

**Example tokenization:**
- `"0x10 + 0x20"` → `[TOKEN_HEX_NUMBER("0x10"), TOKEN_OPERATOR("+"), TOKEN_HEX_NUMBER("0x20")]`
- `"r0 + r1"` → `[TOKEN_REGISTER("r0"), TOKEN_OPERATOR("+"), TOKEN_REGISTER("r1")]`
- `"0xFF & 0x0F"` → `[TOKEN_HEX_NUMBER("0xFF"), TOKEN_OPERATOR("&"), TOKEN_HEX_NUMBER("0x0F")]`
- `"[r0 + 4]"` → `[TOKEN_LBRACKET, TOKEN_REGISTER("r0"), TOKEN_OPERATOR("+"), TOKEN_NUMBER("4"), TOKEN_RBRACKET]`

#### Phase 2: Parsing (Syntax Analysis)

Implement a proper expression parser using one of these approaches:

**Option A: Recursive Descent Parser**
```go
type Parser struct {
    tokens []Token
    pos    int
    vm     *vm.VM
    symbols map[string]uint32
}

func (p *Parser) Parse() (uint32, error)
func (p *Parser) parseExpression() (uint32, error)
func (p *Parser) parseTerm() (uint32, error)
func (p *Parser) parseFactor() (uint32, error)
func (p *Parser) parsePrimary() (uint32, error)
```

**Option B: Shunting-Yard Algorithm** (better for operator precedence)
```go
func (p *Parser) ShuntingYard() ([]Token, error)  // Convert to postfix
func (p *Parser) EvaluatePostfix(postfix []Token) (uint32, error)
```

**Option C: Precedence Climbing**
```go
func (p *Parser) ParseExpression(minPrecedence int) (uint32, error)
```

#### Operator Precedence (from highest to lowest)
1. Memory dereference: `[]`, `*`
2. Multiplication/Division: `*`, `/`
3. Addition/Subtraction: `+`, `-`
4. Bitwise Shift: `<<`, `>>`
5. Bitwise AND: `&`
6. Bitwise XOR: `^`
7. Bitwise OR: `|`

#### Implementation Steps

1. **Create `debugger/lexer.go`** (2-3 hours)
   - [ ] Implement Token struct and TokenType enum
   - [ ] Implement Lexer struct with Tokenize() method
   - [ ] Handle all token types (numbers, registers, operators, symbols)
   - [ ] Add whitespace skipping
   - [ ] Handle hex (0x), binary (0b), octal (0) number prefixes
   - [ ] Recognize all operators: `+`, `-`, `*`, `/`, `&`, `|`, `^`, `<<`, `>>`
   - [ ] Handle brackets: `[`, `]`, `(`, `)`
   - [ ] Handle memory deref operator: `*`

2. **Create `debugger/parser.go`** (2-3 hours)
   - [ ] Implement Parser struct
   - [ ] Implement expression parsing with operator precedence
   - [ ] Handle parentheses for grouping
   - [ ] Support memory dereferencing `[expr]` and `*expr`
   - [ ] Integrate with existing VM and symbol table

3. **Update `debugger/expressions.go`** (1 hour)
   - [ ] Replace current `evaluate()` method with tokenizer + parser approach
   - [ ] Keep existing helper methods (parseNumber, evalRegister, etc.) for use by parser
   - [ ] Maintain backward compatibility with value history

4. **Add Tests** (1-2 hours)
   - [ ] Re-enable all disabled tests in `expressions_test.go`
   - [ ] Add tests for complex expressions: `(r0 + r1) * 2`
   - [ ] Add tests for nested memory access: `[r0 + [r1]]`
   - [ ] Add tests for operator precedence: `2 + 3 * 4` should equal 14
   - [ ] Add tests for parentheses: `(2 + 3) * 4` should equal 20

5. **Documentation** (30 minutes)
   - [ ] Update debugger help text with expression syntax
   - [ ] Document operator precedence
   - [ ] Add expression examples to README

**Effort Estimate:** 6-8 hours total

**Priority:** Medium - Current implementation is functional for common use cases

**Assigned To:** Unassigned

**Related Files:**
- `debugger/expressions.go` (to be refactored)
- `debugger/expressions_test.go` (tests to re-enable)
- `debugger/lexer.go` (new file)
- `debugger/parser.go` (new file, or refactor expressions.go)

---

## Phase 6: TUI Interface (Not Started)

### 2. Implement Text User Interface

**Status:** Pending

**Requirements:**
- [ ] Source code view with current line highlighting
- [ ] Register view with live updates
- [ ] Memory view with hex/ASCII display
- [ ] Stack view
- [ ] Disassembly view
- [ ] Command input panel
- [ ] Output/console panel
- [ ] Breakpoints panel
- [ ] Watchpoints panel
- [ ] Responsive layout with resize handling
- [ ] Keyboard shortcuts
- [ ] Mouse support (optional)

**Dependencies:**
- `github.com/rivo/tview` (already in go.mod)
- `github.com/gdamore/tcell/v2` (already in go.mod)

**Files to Create:**
- `debugger/tui.go` - Main TUI implementation
- `debugger/tui_views.go` - Individual panel implementations
- `debugger/tui_test.go` - TUI tests

**Effort Estimate:** 12-16 hours

**Priority:** High (next phase)

---

## Phase 7: Testing & Coverage (Not Started)

### 3. Expand Test Coverage

**Current Status:** 161 tests passing

**Target:** 1000+ tests

**Required Coverage:**
- [ ] Instruction tests: 600+ tests (need ~500 more)
- [ ] Flag calculation tests: 100+ tests
- [ ] Memory system tests: 50+ tests
- [ ] Parser tests: 90+ tests (have 35, need ~55 more)
- [ ] Addressing mode tests: 60+ tests
- [ ] System call tests: 30+ tests (have 15, need ~15 more)
- [ ] Debugger tests: 100+ tests (have 60, need ~40 more)
- [ ] Integration tests: 50+ tests

**Coverage Targets:**
- [ ] Instruction execution: 95%
- [ ] Memory system: 90%
- [ ] Parser: 85%
- [ ] VM core: 90%
- [ ] Overall: 85%

**Effort Estimate:** 20-30 hours

**Priority:** Medium

---

## Phase 8: Development Tools (Not Started)

### 4. Assembly Linter

**Status:** Not started

**Features:**
- [ ] Syntax validation
- [ ] Undefined label detection
- [ ] Unreachable code detection
- [ ] Register usage warnings
- [ ] Dead code elimination suggestions

**File:** `tools/lint.go`

**Effort Estimate:** 6-8 hours

**Priority:** Low

### 5. Code Formatter

**Status:** Not started

**Features:**
- [ ] Consistent indentation
- [ ] Operand alignment
- [ ] Comment alignment
- [ ] Whitespace normalization

**File:** `tools/format.go`

**Effort Estimate:** 4-6 hours

**Priority:** Low

### 6. Cross-Reference Generator

**Status:** Not started

**Features:**
- [ ] Symbol cross-reference
- [ ] Function reference tracking
- [ ] Jump target analysis
- [ ] Call graph generation

**File:** `tools/xref.go`

**Effort Estimate:** 4-6 hours

**Priority:** Low

---

## Phase 9: Examples & Documentation (Not Started)

### 7. Example Programs

**Status:** Not started

**Basic Examples:**
- [ ] hello.s - Hello World
- [ ] arithmetic.s - Basic arithmetic

**Algorithm Examples:**
- [ ] fibonacci.s - Fibonacci sequence
- [ ] factorial.s - Factorial (iterative and recursive)
- [ ] bubble_sort.s - Bubble sort
- [ ] binary_search.s - Binary search
- [ ] gcd.s - Greatest common divisor

**Data Structure Examples:**
- [ ] arrays.s - Array operations
- [ ] linked_list.s - Linked list
- [ ] stack.s - Stack implementation
- [ ] strings.s - String manipulation

**Advanced Examples:**
- [ ] functions.s - Function calling conventions
- [ ] conditionals.s - If/else, switch/case
- [ ] loops.s - For, while, do-while

**Effort Estimate:** 8-12 hours

**Priority:** Medium

### 8. Documentation

**Status:** Partial (README exists, needs expansion)

**User Documentation:**
- [ ] README.md - Overview, quick start (expand current version)
- [ ] docs/installation.md
- [ ] docs/assembly_reference.md
- [ ] docs/debugger_reference.md
- [ ] docs/tutorial.md
- [ ] docs/faq.md
- [ ] docs/syscalls.md (partially done in code comments)

**Developer Documentation:**
- [ ] docs/api_reference.md
- [ ] docs/architecture.md
- [ ] docs/contributing.md
- [ ] docs/coding_standards.md

**Effort Estimate:** 12-16 hours

**Priority:** Medium

---

## Phase 10: Cross-Platform & Polish (Not Started)

### 9. Cross-Platform Testing

**Status:** Not started

**Requirements:**
- [ ] Test on macOS (development platform)
- [ ] Test on Windows
- [ ] Test on Linux
- [ ] Fix any platform-specific issues
- [ ] CI/CD pipeline for all platforms
- [ ] Cross-compilation builds

**Effort Estimate:** 4-6 hours

**Priority:** Medium

### 10. Performance Optimization

**Status:** Not started

**Features:**
- [ ] Execution trace
- [ ] Memory access log
- [ ] Performance statistics
- [ ] Hot path analysis
- [ ] Profiling support
- [ ] Export formats (JSON, CSV, HTML)

**Effort Estimate:** 6-8 hours

**Priority:** Low

---

## Future Enhancements

### 11. Advanced Call Stack Tracking

**Status:** Basic implementation complete

**Enhancements:**
- [ ] Frame selection with `up`/`down` commands
- [ ] Frame-relative variable inspection
- [ ] Stack unwinding for exception handling
- [ ] Call graph visualization

**Effort Estimate:** 3-4 hours

**Priority:** Low

### 12. Disassembler

**Status:** Not started

**Features:**
- [ ] Binary to assembly conversion
- [ ] Symbol recovery
- [ ] Control flow analysis
- [ ] Data section identification

**File:** `tools/disassembler.go`

**Effort Estimate:** 10-12 hours

**Priority:** Low

### 13. Remote Debugging

**Status:** Not started

**Features:**
- [ ] GDB remote protocol support
- [ ] Network debugging
- [ ] Multiple client support
- [ ] Secure connections

**Effort Estimate:** 12-16 hours

**Priority:** Low

---

## Bug Fixes & Technical Debt

### Known Issues

1. **Expression Parser** (see item #1 above)
   - Cannot parse hex numbers with arithmetic operators
   - Cannot parse register operations
   - Cannot parse hex numbers with bitwise operators

2. **None currently** - All other features working as expected

### Technical Debt

1. **Code Coverage**
   - Current: ~40% estimated
   - Target: 85%+
   - Action: Add more comprehensive tests

2. **Error Messages**
   - Some error messages could be more descriptive
   - Add suggestions for common mistakes
   - Improve error context

3. **Performance**
   - No profiling done yet
   - Potential optimization opportunities in fetch-decode-execute cycle
   - Memory allocations could be reduced

---

## Completed Items ✅

- ✅ Phase 1: Core VM Implementation
- ✅ Phase 2: Parser & Assembler
- ✅ Phase 3: Instruction Set
- ✅ Phase 4: System Integration
- ✅ Phase 5: Debugger Core (with noted expression parser limitations)

---

## Notes

- Priority levels: High (next phase), Medium (important), Low (nice to have)
- Effort estimates are approximate and may vary
- Some items may be dependencies for others
- This list will be updated as development progresses
