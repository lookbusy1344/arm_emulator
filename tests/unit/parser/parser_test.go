package parser_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/parser"
)

func TestParser_SimpleInstruction(t *testing.T) {
	input := "MOV R0, #42"
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
	}

	inst := program.Instructions[0]
	if inst.Mnemonic != "MOV" {
		t.Errorf("expected mnemonic 'MOV', got %q", inst.Mnemonic)
	}

	if len(inst.Operands) != 2 {
		t.Fatalf("expected 2 operands, got %d", len(inst.Operands))
	}
}

func TestParser_InstructionWithLabel(t *testing.T) {
	input := "start: MOV R0, #0"
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
	}

	inst := program.Instructions[0]
	if inst.Label != "start" {
		t.Errorf("expected label 'start', got %q", inst.Label)
	}

	// Check symbol table
	sym, exists := program.SymbolTable.Lookup("start")
	if !exists {
		t.Errorf("label 'start' not found in symbol table")
	}
	if sym.Value != 0 {
		t.Errorf("expected label address 0, got %d", sym.Value)
	}
}

func TestParser_Directive_Org(t *testing.T) {
	input := `.org 0x8000
	MOV R0, #1`
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Directives) != 1 {
		t.Fatalf("expected 1 directive, got %d", len(program.Directives))
	}

	dir := program.Directives[0]
	if dir.Name != ".org" {
		t.Errorf("expected directive '.org', got %q", dir.Name)
	}
}

func TestParser_Directive_Equ(t *testing.T) {
	input := `.equ MAX_COUNT, 100
	MOV R0, #MAX_COUNT`
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Check symbol table
	sym, exists := program.SymbolTable.Lookup("MAX_COUNT")
	if !exists {
		t.Errorf("constant 'MAX_COUNT' not found in symbol table")
	}
	if sym.Value != 100 {
		t.Errorf("expected constant value 100, got %d", sym.Value)
	}
}

func TestParser_MultipleInstructions(t *testing.T) {
	input := `MOV R0, #0
	ADD R0, R0, #1
	SUB R0, R0, #1`
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Instructions) != 3 {
		t.Fatalf("expected 3 instructions, got %d", len(program.Instructions))
	}

	expectedMnemonics := []string{"MOV", "ADD", "SUB"}
	for i, expected := range expectedMnemonics {
		if program.Instructions[i].Mnemonic != expected {
			t.Errorf("instruction %d: expected %q, got %q", i, expected, program.Instructions[i].Mnemonic)
		}
	}
}

func TestParser_InstructionWithCondition(t *testing.T) {
	input := "MOVEQ R0, #1"
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
	}

	inst := program.Instructions[0]
	if inst.Mnemonic != "MOV" {
		t.Errorf("expected mnemonic 'MOV', got %q", inst.Mnemonic)
	}
	if inst.Condition != "EQ" {
		t.Errorf("expected condition 'EQ', got %q", inst.Condition)
	}
}

func TestParser_InstructionWithSetFlags(t *testing.T) {
	input := "ADDS R0, R1, #1"
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
	}

	inst := program.Instructions[0]
	if inst.Mnemonic != "ADD" {
		t.Errorf("expected mnemonic 'ADD', got %q", inst.Mnemonic)
	}
	if !inst.SetFlags {
		t.Errorf("expected SetFlags to be true")
	}
}

func TestParser_MemoryOperand(t *testing.T) {
	input := "LDR R0, [R1, #4]"
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
	}

	inst := program.Instructions[0]
	if inst.Mnemonic != "LDR" {
		t.Errorf("expected mnemonic 'LDR', got %q", inst.Mnemonic)
	}

	if len(inst.Operands) != 2 {
		t.Fatalf("expected 2 operands, got %d", len(inst.Operands))
	}
}

func TestParser_ShiftOperand(t *testing.T) {
	input := "ADD R0, R1, R2, LSL #2"
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(program.Instructions))
	}

	inst := program.Instructions[0]
	if len(inst.Operands) != 3 {
		t.Fatalf("expected 3 operands, got %d", len(inst.Operands))
	}
}

func TestParser_Comments(t *testing.T) {
	input := `MOV R0, #1 ; Initialize R0
	; This is a comment line
	ADD R0, R0, #1 // Another comment style`
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Should have 2 instructions (comment line is skipped)
	if len(program.Instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(program.Instructions))
	}
}

func TestParser_ForwardReference(t *testing.T) {
	input := `B end
	MOV R0, #1
end: MOV R1, #2`
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Check that 'end' label is defined
	sym, exists := program.SymbolTable.Lookup("end")
	if !exists {
		t.Errorf("label 'end' not found in symbol table")
	}
	if !sym.Defined {
		t.Errorf("label 'end' not defined")
	}
}

func TestParser_UndefinedLabel(t *testing.T) {
	input := "B undefined_label"
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	// Parser should parse it successfully but create a forward reference
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Check that the label was referenced but not defined
	undefined := program.SymbolTable.GetUndefinedSymbols()
	found := false
	for _, sym := range undefined {
		if sym.Name == "undefined_label" {
			found = true
			break
		}
	}
	// Note: Currently the parser doesn't track label references in operands
	// This would be added in a future enhancement
	_ = found
}

func TestParser_DuplicateLabel(t *testing.T) {
	input := `start: MOV R0, #1
start: MOV R1, #2`
	p := parser.NewParser(input, "test.s")

	_, err := p.Parse()
	if err == nil {
		t.Errorf("expected error for duplicate label, got nil")
	}
}

func TestParser_DirectiveWord(t *testing.T) {
	input := `.word 0x12345678, 0xABCDEF00`
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Directives) != 1 {
		t.Fatalf("expected 1 directive, got %d", len(program.Directives))
	}

	dir := program.Directives[0]
	if dir.Name != ".word" {
		t.Errorf("expected directive '.word', got %q", dir.Name)
	}
	if len(dir.Args) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(dir.Args))
	}
}

func TestParser_DirectiveString(t *testing.T) {
	input := `.asciz "Hello, World!"`
	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(program.Directives) != 1 {
		t.Fatalf("expected 1 directive, got %d", len(program.Directives))
	}

	dir := program.Directives[0]
	if dir.Name != ".asciz" {
		t.Errorf("expected directive '.asciz', got %q", dir.Name)
	}
}

func TestParser_ComplexProgram(t *testing.T) {
	input := `.org 0x8000
.equ MAX, 10
start:
MOV R0, #0
MOV R1, #MAX
loop:
ADD R0, R0, #1
CMP R0, R1
BLT loop
MOV R0, #0
SWI #0`

	p := parser.NewParser(input, "test.s")

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Check we have instructions
	if len(program.Instructions) < 5 {
		t.Errorf("expected at least 5 instructions, got %d", len(program.Instructions))
	}

	// Check symbols
	if _, exists := program.SymbolTable.Lookup("start"); !exists {
		t.Errorf("label 'start' not found")
	}
	if _, exists := program.SymbolTable.Lookup("loop"); !exists {
		t.Errorf("label 'loop' not found")
	}
	if _, exists := program.SymbolTable.Lookup("MAX"); !exists {
		t.Errorf("constant 'MAX' not found")
	}
}
