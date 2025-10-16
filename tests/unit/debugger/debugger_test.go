package debugger_test

import (
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/debugger"
	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestNewDebugger tests debugger creation
func TestNewDebugger(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	if dbg == nil {
		t.Fatal("NewDebugger returned nil")
	}

	if dbg.VM != machine {
		t.Error("VM not set correctly")
	}

	if dbg.Breakpoints == nil {
		t.Error("Breakpoints not initialized")
	}

	if dbg.Watchpoints == nil {
		t.Error("Watchpoints not initialized")
	}

	if dbg.History == nil {
		t.Error("History not initialized")
	}

	if dbg.Evaluator == nil {
		t.Error("Evaluator not initialized")
	}
}

// TestLoadSymbols tests symbol loading
func TestLoadSymbols(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	symbols := map[string]uint32{
		"main":   0x1000,
		"_start": 0x2000,
		"loop":   0x3000,
	}

	dbg.LoadSymbols(symbols)

	if len(dbg.Symbols) != 3 {
		t.Errorf("Expected 3 symbols, got %d", len(dbg.Symbols))
	}

	if dbg.Symbols["main"] != 0x1000 {
		t.Errorf("Expected main at 0x1000, got 0x%08X", dbg.Symbols["main"])
	}
}

// TestResolveAddress tests address resolution
func TestResolveAddress(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Add test symbols
	dbg.LoadSymbols(map[string]uint32{
		"main": 0x1000,
		"loop": 0x2000,
	})

	tests := []struct {
		name    string
		input   string
		want    uint32
		wantErr bool
	}{
		{"Symbol", "main", 0x1000, false},
		{"Hex address", "0x3000", 0x3000, false},
		{"Decimal address", "4096", 4096, false},
		{"Invalid symbol", "nonexistent", 0, true},
		{"Invalid hex", "0xGGGG", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dbg.ResolveAddress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ResolveAddress() = 0x%08X, want 0x%08X", got, tt.want)
			}
		})
	}
}

// TestExecuteCommand tests command execution
func TestExecuteCommand(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	tests := []struct {
		name      string
		command   string
		wantErr   bool
		checkFunc func(*testing.T, *debugger.Debugger)
	}{
		{
			name:    "Help command",
			command: "help",
			wantErr: false,
			checkFunc: func(t *testing.T, d *debugger.Debugger) {
				output := d.GetOutput()
				if !strings.Contains(output, "ARM2 Debugger Commands") {
					t.Error("Help output not found")
				}
			},
		},
		{
			name:    "Reset command",
			command: "reset",
			wantErr: false,
			checkFunc: func(t *testing.T, d *debugger.Debugger) {
				if d.VM.CPU.PC != 0 {
					t.Error("VM not reset")
				}
			},
		},
		{
			name:    "Invalid command",
			command: "invalidcmd",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dbg.ExecuteCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, dbg)
			}
		})
	}
}

// TestBreakpointCommands tests breakpoint commands
func TestBreakpointCommands(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Set breakpoint
	err := dbg.ExecuteCommand("break 0x1000")
	if err != nil {
		t.Fatalf("Failed to set breakpoint: %v", err)
	}

	output := dbg.GetOutput()
	if !strings.Contains(output, "Breakpoint") {
		t.Error("Breakpoint not confirmed in output")
	}

	// Check breakpoint was created
	bp := dbg.Breakpoints.GetBreakpoint(0x1000)
	if bp == nil {
		t.Fatal("Breakpoint not created")
	}

	if !bp.Enabled {
		t.Error("Breakpoint not enabled")
	}

	// Disable breakpoint
	err = dbg.ExecuteCommand("disable 1")
	if err != nil {
		t.Fatalf("Failed to disable breakpoint: %v", err)
	}

	if bp.Enabled {
		t.Error("Breakpoint still enabled after disable")
	}

	// Enable breakpoint
	err = dbg.ExecuteCommand("enable 1")
	if err != nil {
		t.Fatalf("Failed to enable breakpoint: %v", err)
	}

	if !bp.Enabled {
		t.Error("Breakpoint not enabled after enable")
	}

	// Delete breakpoint
	err = dbg.ExecuteCommand("delete 1")
	if err != nil {
		t.Fatalf("Failed to delete breakpoint: %v", err)
	}

	bp = dbg.Breakpoints.GetBreakpoint(0x1000)
	if bp != nil {
		t.Error("Breakpoint not deleted")
	}
}

// TestTemporaryBreakpoint tests temporary breakpoints
func TestTemporaryBreakpoint(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Set temporary breakpoint
	err := dbg.ExecuteCommand("tbreak 0x2000")
	if err != nil {
		t.Fatalf("Failed to set temporary breakpoint: %v", err)
	}

	bp := dbg.Breakpoints.GetBreakpoint(0x2000)
	if bp == nil {
		t.Fatal("Temporary breakpoint not created")
	}

	if !bp.Temporary {
		t.Error("Breakpoint not marked as temporary")
	}

	// Set PC to breakpoint address
	machine.CPU.PC = 0x2000

	// Check if should break (this will delete the temporary breakpoint)
	shouldBreak, reason := dbg.ShouldBreak()
	if !shouldBreak {
		t.Error("Should break at temporary breakpoint")
	}

	if !strings.Contains(reason, "breakpoint") {
		t.Errorf("Wrong break reason: %s", reason)
	}

	// Verify breakpoint was deleted
	bp = dbg.Breakpoints.GetBreakpoint(0x2000)
	if bp != nil {
		t.Error("Temporary breakpoint not deleted after hit")
	}
}

// TestInfoRegisters tests the info registers command
func TestInfoRegisters(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Set some register values
	machine.CPU.R[0] = 0x12345678
	machine.CPU.R[1] = 0xABCDEF00
	machine.CPU.SetSP(0xFFFF0000)
	machine.CPU.PC = 0x1000

	err := dbg.ExecuteCommand("info registers")
	if err != nil {
		t.Fatalf("Failed to execute info registers: %v", err)
	}

	output := dbg.GetOutput()

	// Check that output contains register values
	if !strings.Contains(output, "R0") {
		t.Error("Output missing R0")
	}

	if !strings.Contains(output, "0x12345678") {
		t.Error("Output missing R0 value")
	}

	if !strings.Contains(output, "SP") {
		t.Error("Output missing SP")
	}

	if !strings.Contains(output, "CPSR") {
		t.Error("Output missing CPSR")
	}
}

// TestInfoBreakpoints tests the info breakpoints command
func TestInfoBreakpoints(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Add breakpoints
	dbg.Breakpoints.AddBreakpoint(0x1000, false, "")
	dbg.Breakpoints.AddBreakpoint(0x2000, false, "r0 == 5")

	err := dbg.ExecuteCommand("info breakpoints")
	if err != nil {
		t.Fatalf("Failed to execute info breakpoints: %v", err)
	}

	output := dbg.GetOutput()

	if !strings.Contains(output, "0x00001000") {
		t.Error("Output missing first breakpoint")
	}

	if !strings.Contains(output, "0x00002000") {
		t.Error("Output missing second breakpoint")
	}

	if !strings.Contains(output, "r0 == 5") {
		t.Error("Output missing condition")
	}
}

// TestPrintCommand tests the print command
func TestPrintCommand(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Set register value
	machine.CPU.R[5] = 42

	err := dbg.ExecuteCommand("print r5")
	if err != nil {
		t.Fatalf("Failed to execute print: %v", err)
	}

	output := dbg.GetOutput()

	if !strings.Contains(output, "42") {
		t.Errorf("Output missing value 42: %s", output)
	}
}

// TestExamineMemory tests the examine memory command
func TestExamineMemory(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Write test data to memory - use data segment address
	testAddr := uint32(0x00020000) // Data segment start
	machine.Memory.WriteWord(testAddr, 0x12345678)

	err := dbg.ExecuteCommand("x 0x00020000")
	if err != nil {
		t.Fatalf("Failed to execute examine: %v", err)
	}

	output := dbg.GetOutput()

	if !strings.Contains(output, "0x12345678") {
		t.Errorf("Output missing memory value: %s", output)
	}
}

// TestSetRegister tests the set register command
func TestSetRegister(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	err := dbg.ExecuteCommand("set r3 = 0x100")
	if err != nil {
		t.Fatalf("Failed to set register: %v", err)
	}

	if machine.CPU.R[3] != 0x100 {
		t.Errorf("Register not set correctly: got 0x%08X, want 0x100", machine.CPU.R[3])
	}
}

// TestStepMode tests stepping modes
func TestStepMode(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Test step command
	err := dbg.ExecuteCommand("step")
	if err != nil {
		t.Fatalf("Failed to execute step: %v", err)
	}

	if dbg.StepMode != debugger.StepSingle {
		t.Error("Step mode not set to debugger.StepSingle")
	}

	if !dbg.Running {
		t.Error("Running flag not set")
	}

	// Check that ShouldBreak returns true for single step
	shouldBreak, reason := dbg.ShouldBreak()
	if !shouldBreak {
		t.Error("Should break after single step")
	}

	if !strings.Contains(reason, "single step") {
		t.Errorf("Wrong break reason: %s", reason)
	}

	// Verify step mode was cleared
	if dbg.StepMode != debugger.StepNone {
		t.Error("Step mode not cleared after break")
	}
}

// TestCommandHistory tests command history functionality
func TestCommandHistory(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Execute some commands
	cmds := []string{"break 0x1000", "step", "continue"}
	for _, cmd := range cmds {
		dbg.ExecuteCommand(cmd)
	}

	// Check history
	history := dbg.History.GetAll()
	if len(history) != len(cmds) {
		t.Errorf("Expected %d commands in history, got %d", len(cmds), len(history))
	}

	// Check last command
	last := dbg.History.GetLast()
	if last != cmds[len(cmds)-1] {
		t.Errorf("Last command = %s, want %s", last, cmds[len(cmds)-1])
	}
}

// TestShouldBreak tests breakpoint detection
func TestShouldBreak(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Set breakpoint
	dbg.Breakpoints.AddBreakpoint(0x1000, false, "")

	// PC not at breakpoint
	machine.CPU.PC = 0x2000
	shouldBreak, _ := dbg.ShouldBreak()
	if shouldBreak {
		t.Error("Should not break when PC not at breakpoint")
	}

	// PC at breakpoint
	machine.CPU.PC = 0x1000
	shouldBreak, reason := dbg.ShouldBreak()
	if !shouldBreak {
		t.Error("Should break when PC at breakpoint")
	}

	if !strings.Contains(reason, "breakpoint") {
		t.Errorf("Wrong break reason: %s", reason)
	}

	// Check hit count
	bp := dbg.Breakpoints.GetBreakpoint(0x1000)
	if bp.HitCount != 1 {
		t.Errorf("Hit count = %d, want 1", bp.HitCount)
	}
}

// TestConditionalBreakpoint tests breakpoints with conditions
func TestConditionalBreakpoint(t *testing.T) {
	machine := vm.NewVM()
	dbg := debugger.NewDebugger(machine)

	// Set conditional breakpoint
	dbg.Breakpoints.AddBreakpoint(0x1000, false, "r0")
	machine.CPU.PC = 0x1000

	// Condition false (r0 == 0)
	machine.CPU.R[0] = 0
	shouldBreak, _ := dbg.ShouldBreak()
	if shouldBreak {
		t.Error("Should not break when condition is false")
	}

	// Condition true (r0 != 0)
	machine.CPU.R[0] = 5
	shouldBreak, reason := dbg.ShouldBreak()
	if !shouldBreak {
		t.Error("Should break when condition is true")
	}

	if !strings.Contains(reason, "breakpoint") {
		t.Errorf("Wrong break reason: %s", reason)
	}
}
