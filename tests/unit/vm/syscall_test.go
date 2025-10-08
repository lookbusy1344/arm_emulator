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
	setupCodeWrite(v); v.Memory.WriteWord(0x8000, opcode)
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
	setupCodeWrite(v); v.Memory.WriteWord(0x8000, opcode)
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
	setupCodeWrite(v); v.Memory.WriteWord(0x8000, 0xEF000020)
	v.Step()

	addr := v.CPU.R[0]
	if addr == 0 {
		t.Fatal("allocation failed")
	}

	// Now free the memory
	v.CPU.R[0] = addr
	v.CPU.PC = 0x8004

	// SWI #0x21 (free)
	setupCodeWrite(v); v.Memory.WriteWord(0x8004, 0xEF000021)
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
	setupCodeWrite(v); v.Memory.WriteWord(0x8000, 0xEF000021)
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
	setupCodeWrite(v); v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	if err != nil {
		t.Fatalf("get time failed: %v", err)
	}

	// R0 should contain a timestamp (non-zero)
	if v.CPU.R[0] == 0 {
		t.Error("expected non-zero timestamp")
	}
}

func TestSWI_GetRandom(t *testing.T) {
	// SWI #0x31 (get random number)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// SWI #0x31 (EF000031)
	opcode := uint32(0xEF000031)
	setupCodeWrite(v); v.Memory.WriteWord(0x8000, opcode)
	err := v.Step()

	if err != nil {
		t.Fatalf("get random failed: %v", err)
	}

	// R0 should contain a random number
	random1 := v.CPU.R[0]

	// Call again
	v.CPU.PC = 0x8004
	setupCodeWrite(v); v.Memory.WriteWord(0x8004, opcode)
	v.Step()

	random2 := v.CPU.R[0]

	// Two calls should (very likely) return different values
	if random1 == random2 {
		t.Log("Warning: two random calls returned same value (rare but possible)")
	}
}

func TestSWI_Breakpoint(t *testing.T) {
	// SWI #0xF1 (breakpoint)
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// SWI #0xF1 (EF0000F1)
	opcode := uint32(0xEF0000F1)
	setupCodeWrite(v); v.Memory.WriteWord(0x8000, opcode)
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

func TestSWI_MultipleAllocations(t *testing.T) {
	// Test multiple allocations don't overlap
	v := vm.NewVM()
	addresses := make([]uint32, 5)

	for i := 0; i < 5; i++ {
		v.CPU.R[0] = 32 // Allocate 32 bytes each
		v.CPU.PC = 0x8000 + uint32(i*4)

		setupCodeWrite(v); v.Memory.WriteWord(v.CPU.PC, 0xEF000020)
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

	setupCodeWrite(v); v.Memory.WriteWord(0x8000, 0xEF000020)
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
	setupCodeWrite(v); v.Memory.WriteWord(0x8000, 0xEF0000FF)
	err := v.Step()

	// Should return error
	if err == nil {
		t.Error("expected error for unimplemented syscall")
	}
}
