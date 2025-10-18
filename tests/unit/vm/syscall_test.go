package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestSWI_Exit(t *testing.T) {
	// SWI #0 (exit)
	v := vm.NewVM()
	v.CPU.R[0] = 42 // Exit code
	v.CPU.PC = 0x8000

	// SWI #0 (EF000000)
	opcode := uint32(0xEF000000)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	// Should return error with exit code
	if err == nil {
		t.Error("expected error for exit syscall")
	}

	// VM should be halted
	if v.State != vm.StateHalted {
		t.Errorf("expected state=Halted, got state=%v", v.State)
	}
}

func TestSWI_Allocate(t *testing.T) {
	// SWI #0x20 (allocate memory)
	v := vm.NewVM()
	v.CPU.R[0] = 100 // Size to allocate
	v.CPU.PC = 0x8000

	// SWI #0x20 (EF000020)
	opcode := uint32(0xEF000020)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	if err != nil {
		t.Fatalf("allocate failed: %v", err)
	}

	// R0 should contain allocated address
	addr := v.CPU.R[0]
	if addr == 0 {
		t.Error("expected non-zero address")
	}

	// Address should be in heap segment
	if addr < 0x30000 || addr >= 0x40000 {
		t.Errorf("allocated address 0x%X not in heap segment", addr)
	}
}

func TestSWI_AllocateAndFree(t *testing.T) {
	// Allocate then free
	v := vm.NewVM()
	v.CPU.R[0] = 64
	v.CPU.PC = 0x8000

	// SWI #0x20 (allocate)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xEF000020)
	v.Step()

	addr := v.CPU.R[0]
	if addr == 0 {
		t.Fatal("allocation failed")
	}

	// Now free the memory
	v.CPU.R[0] = addr
	v.CPU.PC = 0x8004

	// SWI #0x21 (free)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8004, 0xEF000021)
	err := v.Step()

	if err != nil {
		t.Fatalf("free failed: %v", err)
	}

	// R0 should be 0 (success)
	if v.CPU.R[0] != 0 {
		t.Errorf("expected R0=0 (success), got R0=%d", v.CPU.R[0])
	}
}

func TestSWI_FreeInvalidAddress(t *testing.T) {
	// Try to free an address that wasn't allocated
	v := vm.NewVM()
	v.CPU.R[0] = 0x50000 // Invalid address
	v.CPU.PC = 0x8000

	// SWI #0x21 (free)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xEF000021)
	v.Step()

	// R0 should be 0xFFFFFFFF (error)
	if v.CPU.R[0] != 0xFFFFFFFF {
		t.Errorf("expected R0=0xFFFFFFFF (error), got R0=0x%X", v.CPU.R[0])
	}
}

func TestSWI_GetTime(t *testing.T) {
	// SWI #0x30 (get time)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// SWI #0x30 (EF000030)
	opcode := uint32(0xEF000030)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	if err != nil {
		t.Fatalf("get time failed: %v", err)
	}

	// R0 should contain a timestamp (non-zero)
	timestamp1 := v.CPU.R[0]
	if timestamp1 == 0 {
		t.Error("expected non-zero timestamp")
	}

	// Call again and verify time advances
	v.CPU.PC = 0x8004
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8004, opcode)
	v.Step()

	timestamp2 := v.CPU.R[0]
	if timestamp2 < timestamp1 {
		t.Errorf("time went backwards: first=%d, second=%d", timestamp1, timestamp2)
	}

	// Verify PC advanced correctly
	if v.CPU.PC != 0x8008 {
		t.Errorf("expected PC=0x8008, got PC=0x%08X", v.CPU.PC)
	}
}

func TestSWI_GetRandom(t *testing.T) {
	// SWI #0x31 (get random number)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// SWI #0x31 (EF000031)
	opcode := uint32(0xEF000031)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	if err != nil {
		t.Fatalf("get random failed: %v", err)
	}

	// R0 should contain a random number
	random1 := v.CPU.R[0]

	// Call again
	v.CPU.PC = 0x8004
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8004, opcode)
	v.Step()

	random2 := v.CPU.R[0]

	// Two calls should (very likely) return different values
	if random1 == random2 {
		t.Log("Warning: two random calls returned same value (rare but possible)")
	}
}

func TestSWI_DebugPrint(t *testing.T) {
	// SWI #0xF0 (debug print)
	v := vm.NewVM()
	setupDataWrite(v)

	// Write debug message to memory
	msgAddr := uint32(0x10000)
	msg := "Debug test message"
	for i, c := range msg {
		v.Memory.WriteByteAt(msgAddr+uint32(i), byte(c))
	}
	v.Memory.WriteByteAt(msgAddr+uint32(len(msg)), 0) // Null terminator

	v.CPU.R[0] = msgAddr
	v.CPU.PC = 0x8000

	// SWI #0xF0 (EF0000F0)
	opcode := uint32(0xEF0000F0)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)

	// Note: This will print to stderr, but we can't easily capture it in tests
	// We just verify it doesn't error
	err := v.Step()

	if err != nil {
		t.Errorf("debug_print failed: %v", err)
	}

	// PC should have advanced
	if v.CPU.PC != 0x8004 {
		t.Errorf("expected PC=0x8004, got PC=0x%08X", v.CPU.PC)
	}
}

func TestSWI_DebugPrint_InvalidAddress(t *testing.T) {
	// Test DEBUG_PRINT with invalid memory address
	v := vm.NewVM()
	v.CPU.R[0] = 0xFFFFFFFF // Invalid address
	v.CPU.PC = 0x8000

	// SWI #0xF0 (EF0000F0)
	opcode := uint32(0xEF0000F0)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	// Should return error for invalid address
	if err == nil {
		t.Error("expected error for invalid address")
	}
}

func TestSWI_Breakpoint(t *testing.T) {
	// SWI #0xF1 (breakpoint)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// SWI #0xF1 (EF0000F1)
	opcode := uint32(0xEF0000F1)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	// Should return error for breakpoint
	if err == nil {
		t.Error("expected error for breakpoint")
	}

	// VM should be in breakpoint state
	if v.State != vm.StateBreakpoint {
		t.Errorf("expected state=Breakpoint, got state=%v", v.State)
	}
}

func TestSWI_DumpRegisters(t *testing.T) {
	// SWI #0xF2 (dump registers)
	v := vm.NewVM()

	// Set some register values to verify they're dumped
	v.CPU.R[0] = 0x12345678
	v.CPU.R[1] = 0xABCDEF00
	v.CPU.R[13] = 0x50000 // SP
	v.CPU.R[14] = 0x8100  // LR
	v.CPU.PC = 0x8000
	v.CPU.CPSR.N = true
	v.CPU.CPSR.Z = false

	// SWI #0xF2 (EF0000F2)
	opcode := uint32(0xEF0000F2)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)

	// Note: This will print to stdout, but we can't easily capture it in tests
	// We just verify it doesn't error
	err := v.Step()

	if err != nil {
		t.Errorf("dump_registers failed: %v", err)
	}

	// PC should have advanced
	if v.CPU.PC != 0x8004 {
		t.Errorf("expected PC=0x8004, got PC=0x%08X", v.CPU.PC)
	}
}

func TestSWI_DumpMemory(t *testing.T) {
	// SWI #0xF3 (dump memory)
	v := vm.NewVM()
	setupDataWrite(v)

	// Write some test data to memory
	testAddr := uint32(0x10000)
	testData := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
		0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	for i, b := range testData {
		v.Memory.WriteByteAt(testAddr+uint32(i), b)
	}

	v.CPU.R[0] = testAddr
	v.CPU.R[1] = uint32(len(testData))
	v.CPU.PC = 0x8000

	// SWI #0xF3 (EF0000F3)
	opcode := uint32(0xEF0000F3)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)

	// Note: This will print to stdout, but we can't easily capture it in tests
	// We just verify it doesn't error
	err := v.Step()

	if err != nil {
		t.Errorf("dump_memory failed: %v", err)
	}

	// PC should have advanced
	if v.CPU.PC != 0x8004 {
		t.Errorf("expected PC=0x8004, got PC=0x%08X", v.CPU.PC)
	}
}

func TestSWI_DumpMemory_LargeLength(t *testing.T) {
	// Test DUMP_MEMORY with length > 1KB (should be clamped)
	v := vm.NewVM()
	setupDataWrite(v)

	testAddr := uint32(0x10000)
	v.CPU.R[0] = testAddr
	v.CPU.R[1] = 2048 // Larger than 1KB limit
	v.CPU.PC = 0x8000

	// SWI #0xF3 (EF0000F3)
	opcode := uint32(0xEF0000F3)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)

	// Should still work, but only dump up to 1KB
	err := v.Step()

	if err != nil {
		t.Errorf("dump_memory with large length failed: %v", err)
	}

	// PC should have advanced
	if v.CPU.PC != 0x8004 {
		t.Errorf("expected PC=0x8004, got PC=0x%08X", v.CPU.PC)
	}
}

func TestSWI_MultipleAllocations(t *testing.T) {
	// Test multiple allocations don't overlap
	v := vm.NewVM()
	addresses := make([]uint32, 5)

	for i := 0; i < 5; i++ {
		v.CPU.R[0] = 32 // Allocate 32 bytes each
		// #nosec G115 -- i is loop index [0,5), guaranteed non-negative and within bounds
		v.CPU.PC = 0x8000 + uint32(i*4)

		setupCodeWrite(v)
		v.Memory.WriteWord(v.CPU.PC, 0xEF000020)
		err := v.Step()

		if err != nil {
			t.Fatalf("allocation %d failed: %v", i, err)
		}

		addresses[i] = v.CPU.R[0]
		if addresses[i] == 0 {
			t.Fatalf("allocation %d returned null", i)
		}
	}

	// Verify all addresses are unique and properly spaced
	for i := 0; i < 4; i++ {
		if addresses[i] >= addresses[i+1] {
			t.Errorf("addresses not increasing: [%d]=0x%X, [%d]=0x%X",
				i, addresses[i], i+1, addresses[i+1])
		}

		// Should be at least 32 bytes apart (aligned to 4)
		diff := addresses[i+1] - addresses[i]
		if diff < 32 {
			t.Errorf("allocations too close: diff=%d", diff)
		}
	}
}

func TestSWI_AllocateZero(t *testing.T) {
	// Try to allocate 0 bytes (should fail)
	v := vm.NewVM()
	v.CPU.R[0] = 0
	v.CPU.PC = 0x8000

	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xEF000020)
	v.Step()

	// Should return NULL (0) for invalid size
	if v.CPU.R[0] != 0 {
		t.Errorf("expected NULL for zero-size allocation, got 0x%X", v.CPU.R[0])
	}
}

func TestSWI_UnimplementedSyscall(t *testing.T) {
	// Try to call an unimplemented syscall
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// SWI #0xFF (unimplemented)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xEF0000FF)
	err := v.Step()

	// Should return error
	if err == nil {
		t.Error("expected error for unimplemented syscall")
	}
}

func TestSWI_Reallocate(t *testing.T) {
	// Test REALLOCATE syscall (0x22)
	v := vm.NewVM()

	// First allocate some memory
	v.CPU.R[0] = 100
	v.CPU.PC = 0x8000
	allocOpcode := uint32(0xEF000020) // SWI #0x20
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, allocOpcode)
	v.Step()

	oldAddr := v.CPU.R[0]
	if oldAddr == 0 {
		t.Fatal("initial allocation failed")
	}

	// Now reallocate
	v.CPU.R[0] = oldAddr
	v.CPU.R[1] = 200 // New size
	v.CPU.PC = 0x8004

	reallocOpcode := uint32(0xEF000022) // SWI #0x22
	v.Memory.WriteWord(0x8004, reallocOpcode)
	err := v.Step()

	if err != nil {
		t.Fatalf("reallocate failed: %v", err)
	}

	newAddr := v.CPU.R[0]
	if newAddr == 0 {
		t.Error("reallocate returned NULL")
	}
}

func TestSWI_GetArguments(t *testing.T) {
	// Test GET_ARGUMENTS syscall (0x32)
	v := vm.NewVM()
	v.ProgramArguments = []string{"program", "arg1", "arg2"}
	v.CPU.PC = 0x8000

	// SWI #0x32 (EF000032)
	opcode := uint32(0xEF000032)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	if err != nil {
		t.Fatalf("get_arguments failed: %v", err)
	}

	argc := v.CPU.R[0]
	if argc != 3 {
		t.Errorf("expected argc=3, got argc=%d", argc)
	}

	// R1 should contain argv pointer (currently 0 in implementation)
	argv := v.CPU.R[1]
	if argv != 0 {
		t.Logf("argv pointer = 0x%08X (currently unimplemented, returns 0)", argv)
	}
}

func TestSWI_GetArguments_Empty(t *testing.T) {
	// Test GET_ARGUMENTS with no arguments
	v := vm.NewVM()
	v.ProgramArguments = []string{}
	v.CPU.PC = 0x8000

	// SWI #0x32 (EF000032)
	opcode := uint32(0xEF000032)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	if err != nil {
		t.Fatalf("get_arguments failed: %v", err)
	}

	argc := v.CPU.R[0]
	if argc != 0 {
		t.Errorf("expected argc=0 for empty args, got argc=%d", argc)
	}
}

func TestSWI_GetEnvironment(t *testing.T) {
	// Test GET_ENVIRONMENT syscall (0x33)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// SWI #0x33 (EF000033)
	opcode := uint32(0xEF000033)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	if err != nil {
		t.Fatalf("get_environment failed: %v", err)
	}

	// R0 should contain environment pointer (currently 0 in simplified implementation)
	envp := v.CPU.R[0]
	if envp != 0 {
		t.Errorf("expected envp=0 (unimplemented), got envp=0x%08X", envp)
	}

	// PC should have advanced
	if v.CPU.PC != 0x8004 {
		t.Errorf("expected PC=0x8004, got PC=0x%08X", v.CPU.PC)
	}
}

func TestSWI_Assert_Pass(t *testing.T) {
	// Test ASSERT syscall (0xF4) with passing condition
	v := vm.NewVM()
	setupDataWrite(v)

	// Write assertion message to memory
	msgAddr := uint32(0x10000)
	msg := "Test assertion"
	for i, c := range msg {
		v.Memory.WriteByteAt(msgAddr+uint32(i), byte(c))
	}
	v.Memory.WriteByteAt(msgAddr+uint32(len(msg)), 0) // Null terminator

	v.CPU.R[0] = 1 // Condition is true
	v.CPU.R[1] = msgAddr
	v.CPU.PC = 0x8000

	// SWI #0xF4 (EF0000F4)
	opcode := uint32(0xEF0000F4)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	if err != nil {
		t.Errorf("assert with true condition should not fail: %v", err)
	}

	if v.State == vm.StateError {
		t.Error("VM should not be in error state for passing assertion")
	}
}

func TestSWI_Assert_Fail(t *testing.T) {
	// Test ASSERT syscall (0xF4) with failing condition
	v := vm.NewVM()
	setupDataWrite(v)

	// Write assertion message to memory
	msgAddr := uint32(0x10000)
	msg := "Assertion failed message"
	for i, c := range msg {
		v.Memory.WriteByteAt(msgAddr+uint32(i), byte(c))
	}
	v.Memory.WriteByteAt(msgAddr+uint32(len(msg)), 0) // Null terminator

	v.CPU.R[0] = 0 // Condition is false
	v.CPU.R[1] = msgAddr
	v.CPU.PC = 0x8000

	// SWI #0xF4 (EF0000F4)
	opcode := uint32(0xEF0000F4)
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	if err == nil {
		t.Error("assert with false condition should return error")
	}

	if v.State != vm.StateError {
		t.Errorf("expected state=Error, got state=%v", v.State)
	}
}

func TestVM_Bootstrap(t *testing.T) {
	// Test bootstrap functionality
	v := vm.NewVM()
	args := []string{"program", "arg1", "arg2"}

	err := v.Bootstrap(args)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	// Check that arguments were stored
	if len(v.ProgramArguments) != 3 {
		t.Errorf("expected 3 arguments, got %d", len(v.ProgramArguments))
	}

	// Check that stack pointer was initialized
	sp := v.CPU.GetSP()
	expectedSP := uint32(vm.StackSegmentStart + vm.StackSegmentSize)
	if sp != expectedSP {
		t.Errorf("expected SP=0x%08X, got SP=0x%08X", expectedSP, sp)
	}

	// Check that LR was set to halt address
	lr := v.CPU.GetLR()
	if lr != 0xFFFFFFFF {
		t.Errorf("expected LR=0xFFFFFFFF, got LR=0x%08X", lr)
	}

	// Check that PC was set to entry point
	if v.CPU.PC != v.EntryPoint {
		t.Errorf("expected PC=0x%08X, got PC=0x%08X", v.EntryPoint, v.CPU.PC)
	}
}

func TestVM_FindEntryPoint(t *testing.T) {
	// Test entry point detection
	v := vm.NewVM()

	// Test with _start symbol
	symbols := map[string]uint32{
		"_start": 0x8000,
		"foo":    0x8100,
	}

	addr, err := v.FindEntryPoint(symbols)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if addr != 0x8000 {
		t.Errorf("expected entry point=0x8000, got 0x%08X", addr)
	}

	// Test with main symbol
	symbols2 := map[string]uint32{
		"main": 0x9000,
		"foo":  0x8100,
	}

	addr2, err2 := v.FindEntryPoint(symbols2)
	if err2 != nil {
		t.Errorf("unexpected error: %v", err2)
	}
	if addr2 != 0x9000 {
		t.Errorf("expected entry point=0x9000, got 0x%08X", addr2)
	}

	// Test with no entry point (should default to code segment start)
	symbols3 := map[string]uint32{
		"foo": 0x8100,
	}

	addr3, err3 := v.FindEntryPoint(symbols3)
	if err3 == nil {
		t.Error("expected error when no entry point found")
	}
	if addr3 != vm.CodeSegmentStart {
		t.Errorf("expected default entry point=0x%08X, got 0x%08X", vm.CodeSegmentStart, addr3)
	}
}
