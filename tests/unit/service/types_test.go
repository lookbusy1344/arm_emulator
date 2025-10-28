package service_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/service"
)

func TestDisassemblyLine_Creation(t *testing.T) {
	line := service.DisassemblyLine{
		Address: 0x00008000,
		Opcode:  0xE3A00001,
		Symbol:  "main",
	}

	if line.Address != 0x00008000 {
		t.Errorf("Expected address 0x00008000, got 0x%08X", line.Address)
	}
	if line.Opcode != 0xE3A00001 {
		t.Errorf("Expected opcode 0xE3A00001, got 0x%08X", line.Opcode)
	}
	if line.Symbol != "main" {
		t.Errorf("Expected symbol 'main', got '%s'", line.Symbol)
	}
}

func TestStackEntry_Creation(t *testing.T) {
	entry := service.StackEntry{
		Address: 0x00050000,
		Value:   0xDEADBEEF,
		Symbol:  "data_label",
	}

	if entry.Address != 0x00050000 {
		t.Errorf("Expected address 0x00050000, got 0x%08X", entry.Address)
	}
	if entry.Value != 0xDEADBEEF {
		t.Errorf("Expected value 0xDEADBEEF, got 0x%08X", entry.Value)
	}
	if entry.Symbol != "data_label" {
		t.Errorf("Expected symbol 'data_label', got '%s'", entry.Symbol)
	}
}

func TestBreakpointInfo_WithCondition(t *testing.T) {
	bp := service.BreakpointInfo{
		Address:   0x00008010,
		Enabled:   true,
		Condition: "R0 > 10",
	}

	if bp.Condition != "R0 > 10" {
		t.Errorf("Expected condition 'R0 > 10', got '%s'", bp.Condition)
	}
}
