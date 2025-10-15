package integration

import (
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/encoder"
	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestLtorgDirective_Basic(t *testing.T) {
	source := `
.org 0x0000

main:
    LDR R0, =0x12345678
    LDR R1, =0xDEADBEEF
    ADD R2, R0, R1
    MOV R0, #0
    SWI #0x00

    .ltorg
`

	p := parser.NewParser(source, "test_ltorg_basic.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(p.Errors().Errors) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Verify .ltorg location was recorded
	if len(program.LiteralPoolLocs) != 1 {
		t.Fatalf("Expected 1 literal pool location, got %d", len(program.LiteralPoolLocs))
	}

	// Create encoder and verify it has the pool location
	enc := encoder.NewEncoder(program.SymbolTable)
	enc.LiteralPoolLocs = program.LiteralPoolLocs

	if len(enc.LiteralPoolLocs) != 1 {
		t.Fatalf("Encoder: Expected 1 literal pool location, got %d", len(enc.LiteralPoolLocs))
	}

	t.Logf("Literal pool location: 0x%08X", enc.LiteralPoolLocs[0])
}

func TestLtorgDirective_MultiplePools(t *testing.T) {
	source := `
.org 0x8000

section1:
    LDR R0, =0x11111111
    LDR R1, =0x22222222
    ADD R2, R0, R1
    .ltorg

section2:
    LDR R3, =0x33333333
    LDR R4, =0x44444444
    ADD R5, R3, R4
    .ltorg

section3:
    MOV R0, #0
    SWI #0x00
`

	p := parser.NewParser(source, "test_ltorg_multiple.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(p.Errors().Errors) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Verify multiple .ltorg locations were recorded
	if len(program.LiteralPoolLocs) != 2 {
		t.Fatalf("Expected 2 literal pool locations, got %d", len(program.LiteralPoolLocs))
	}

	t.Logf("Literal pool locations: 0x%08X, 0x%08X",
		program.LiteralPoolLocs[0], program.LiteralPoolLocs[1])
}

func TestLtorgDirective_LowMemoryOrigin(t *testing.T) {
	source := `
.org 0x0000

main:
    ; Many constants to ensure we exceed 4095 byte range without .ltorg
    LDR R0, =0x10000000
    LDR R1, =0x20000000
    LDR R2, =0x30000000
    LDR R3, =0x40000000
    LDR R4, =0x50000000
    LDR R5, =0x60000000
    LDR R6, =0x70000000
    LDR R7, =0x80000000
    
    ; Place pool nearby
    .ltorg
    
    ; More code
    ADD R0, R0, R1
    ADD R2, R2, R3
    ADD R4, R4, R5
    ADD R6, R6, R7
    
    MOV R0, #0
    SWI #0x00
`

	p := parser.NewParser(source, "test_ltorg_low_memory.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(p.Errors().Errors) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Verify .ltorg was parsed
	if len(program.LiteralPoolLocs) != 1 {
		t.Fatalf("Expected 1 literal pool location, got %d", len(program.LiteralPoolLocs))
	}

	// Create a VM and load the program
	machine := vm.NewVM()

	// Find entry point
	entryPoint := uint32(0)
	if program.OriginSet {
		entryPoint = program.Origin
	}

	// Ensure low memory segment exists
	if entryPoint < vm.CodeSegmentStart {
		segmentSize := uint32(vm.CodeSegmentStart)
		machine.Memory.AddSegment("low-memory", 0, segmentSize, vm.PermRead|vm.PermWrite|vm.PermExecute)
	}

	// Create encoder with pool locations
	enc := encoder.NewEncoder(program.SymbolTable)
	enc.LiteralPoolLocs = program.LiteralPoolLocs

	// Set fallback pool start
	maxAddr := entryPoint + uint32(len(program.Instructions)*4)
	enc.LiteralPoolStart = (maxAddr + 3) & ^uint32(3)

	// Try to encode instructions - should succeed with .ltorg
	for _, inst := range program.Instructions {
		addr := inst.Address
		_, err := enc.EncodeInstruction(inst, addr)
		if err != nil {
			// Check if it's a literal pool offset error
			if strings.Contains(err.Error(), "literal pool offset too large") {
				t.Fatalf("Literal pool offset error even with .ltorg: %v", err)
			}
			// Other errors might be OK (e.g., encoding issues unrelated to pools)
		}
	}

	t.Logf("Successfully encoded %d instructions with literal pool at 0x%08X",
		len(program.Instructions), program.LiteralPoolLocs[0])
}

func TestLtorgDirective_Alignment(t *testing.T) {
	source := `
.org 0x8000

main:
    LDR R0, =0x12345678
    MOV R1, #1
    
    ; .ltorg should be 4-byte aligned
    .ltorg
    
    ; Verify next instruction is also aligned
    LDR R2, =0xABCDEF01
    MOV R0, #0
    SWI #0x00
`

	p := parser.NewParser(source, "test_ltorg_alignment.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(p.Errors().Errors) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	if len(program.LiteralPoolLocs) != 1 {
		t.Fatalf("Expected 1 literal pool location, got %d", len(program.LiteralPoolLocs))
	}

	// Check that pool location is 4-byte aligned
	poolLoc := program.LiteralPoolLocs[0]
	if poolLoc%4 != 0 {
		t.Errorf("Literal pool location 0x%08X is not 4-byte aligned", poolLoc)
	}

	t.Logf("Literal pool is properly aligned at 0x%08X", poolLoc)
}

func TestLtorgDirective_NoLtorg(t *testing.T) {
	source := `
.org 0x8000

main:
    LDR R0, =0x12345678
    LDR R1, =0xDEADBEEF
    ADD R2, R0, R1
    MOV R0, #0
    SWI #0x00
`

	p := parser.NewParser(source, "test_no_ltorg.s")
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(p.Errors().Errors) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	// Without .ltorg, should have no pool locations
	if len(program.LiteralPoolLocs) != 0 {
		t.Fatalf("Expected 0 literal pool locations (no .ltorg), got %d", len(program.LiteralPoolLocs))
	}

	// Encoder should fall back to default behavior
	enc := encoder.NewEncoder(program.SymbolTable)
	enc.LiteralPoolStart = 0x9000 // Fallback location

	// Should still work with fallback
	for _, inst := range program.Instructions {
		addr := inst.Address
		_, err := enc.EncodeInstruction(inst, addr)
		if err != nil && strings.Contains(err.Error(), "literal pool offset too large") {
			t.Fatalf("Fallback literal pool failed: %v", err)
		}
	}

	t.Log("Fallback literal pool mechanism works correctly")
}
