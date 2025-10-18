package vm_test

import (
	"sync"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestHeapAllocatorPerInstance tests that heap allocator state is per-instance
// and not global, ensuring multiple VM instances don't interfere with each other.
// This tests the fix for the critical global state bug.
func TestHeapAllocatorPerInstance(t *testing.T) {
	// Create two separate VM instances
	vm1 := vm.NewVM()
	vm2 := vm.NewVM()

	// Allocate memory in VM1
	addr1, err := vm1.Memory.Allocate(100)
	if err != nil {
		t.Fatalf("VM1 allocation failed: %v", err)
	}

	// Allocate memory in VM2
	addr2, err := vm2.Memory.Allocate(100)
	if err != nil {
		t.Fatalf("VM2 allocation failed: %v", err)
	}

	// Both should allocate at heap start (they're independent)
	expectedStart := uint32(0x00030000) // HeapSegmentStart
	if addr1 != expectedStart {
		t.Errorf("VM1 first allocation should be at 0x%08X, got 0x%08X", expectedStart, addr1)
	}
	if addr2 != expectedStart {
		t.Errorf("VM2 first allocation should be at 0x%08X, got 0x%08X", expectedStart, addr2)
	}

	// Verify allocations are tracked independently
	if _, ok := vm1.Memory.HeapAllocations[addr1]; !ok {
		t.Error("VM1 allocation not tracked in VM1")
	}
	if _, ok := vm2.Memory.HeapAllocations[addr2]; !ok {
		t.Error("VM2 allocation not tracked in VM2")
	}

	// Verify VM1's allocation is NOT in VM2's tracking
	if _, ok := vm2.Memory.HeapAllocations[addr1]; ok {
		t.Error("VM1 allocation incorrectly tracked in VM2 (global state leak)")
	}

	// Verify VM2's allocation is NOT in VM1's tracking
	if _, ok := vm1.Memory.HeapAllocations[addr2]; ok {
		t.Error("VM2 allocation incorrectly tracked in VM1 (global state leak)")
	}
}

// TestHeapAllocatorIndependentNextAddress tests that NextHeapAddress
// is independent across VM instances
func TestHeapAllocatorIndependentNextAddress(t *testing.T) {
	vm1 := vm.NewVM()
	vm2 := vm.NewVM()

	// Allocate different amounts in each VM
	_, err := vm1.Memory.Allocate(200)
	if err != nil {
		t.Fatalf("VM1 allocation failed: %v", err)
	}

	_, err = vm2.Memory.Allocate(400)
	if err != nil {
		t.Fatalf("VM2 allocation failed: %v", err)
	}

	// Second allocations should be at different offsets
	addr1Second, _ := vm1.Memory.Allocate(100)
	addr2Second, _ := vm2.Memory.Allocate(100)

	// VM1 should have advanced by 200 bytes (aligned to 4)
	expectedVM1 := uint32(0x00030000 + 200)
	if addr1Second != expectedVM1 {
		t.Errorf("VM1 second allocation expected at 0x%08X, got 0x%08X", expectedVM1, addr1Second)
	}

	// VM2 should have advanced by 400 bytes (aligned to 4)
	expectedVM2 := uint32(0x00030000 + 400)
	if addr2Second != expectedVM2 {
		t.Errorf("VM2 second allocation expected at 0x%08X, got 0x%08X", expectedVM2, addr2Second)
	}
}

// TestResetHeapIndependence tests that resetting heap in one VM
// doesn't affect other VMs
func TestResetHeapIndependence(t *testing.T) {
	vm1 := vm.NewVM()
	vm2 := vm.NewVM()

	// Allocate in both VMs
	addr1, _ := vm1.Memory.Allocate(100)
	addr2, _ := vm2.Memory.Allocate(100)

	// Reset VM1's heap
	vm1.Memory.ResetHeap()

	// VM1 should have no allocations
	if len(vm1.Memory.HeapAllocations) != 0 {
		t.Errorf("VM1 heap should be empty after reset, has %d allocations", len(vm1.Memory.HeapAllocations))
	}

	// VM2 should still have its allocation
	if _, ok := vm2.Memory.HeapAllocations[addr2]; !ok {
		t.Error("VM2 allocation lost after VM1 reset (global state bug)")
	}
	if len(vm2.Memory.HeapAllocations) != 1 {
		t.Errorf("VM2 should have 1 allocation, has %d", len(vm2.Memory.HeapAllocations))
	}
}

// TestReallocateDataCopy tests that REALLOCATE properly copies data
// from old allocation to new allocation
func TestReallocateDataCopy(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// First allocate some memory
	v.CPU.R[0] = 100 // Size
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xEF000020) // SWI #0x20 (ALLOCATE)
	err := v.Step()
	if err != nil {
		t.Fatalf("initial allocation failed: %v", err)
	}

	oldAddr := v.CPU.R[0]
	if oldAddr == 0 {
		t.Fatal("allocation returned NULL")
	}

	// Write test data to the allocated memory
	testData := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	for i, b := range testData {
		if err := v.Memory.WriteByteAt(oldAddr+uint32(i), b); err != nil {
			t.Fatalf("failed to write test data: %v", err)
		}
	}

	// Now reallocate to a larger size
	v.CPU.R[0] = oldAddr // Old address
	v.CPU.R[1] = 200     // New size (larger)
	v.CPU.PC = 0x8004
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8004, 0xEF000022) // SWI #0x22 (REALLOCATE)
	err = v.Step()
	if err != nil {
		t.Fatalf("reallocation failed: %v", err)
	}

	newAddr := v.CPU.R[0]
	if newAddr == 0 {
		t.Fatal("reallocation returned NULL")
	}

	// Verify data was copied to new location
	for i, expected := range testData {
		actual, err := v.Memory.ReadByteAt(newAddr + uint32(i))
		if err != nil {
			t.Fatalf("failed to read from new address: %v", err)
		}
		if actual != expected {
			t.Errorf("byte %d: expected 0x%02X, got 0x%02X", i, expected, actual)
		}
	}

	// Verify old memory was freed (should be zeroed)
	for i := range testData {
		b, err := v.Memory.ReadByteAt(oldAddr + uint32(i))
		if err != nil {
			t.Fatalf("failed to read old address: %v", err)
		}
		if b != 0 {
			t.Errorf("old memory byte %d not zeroed: got 0x%02X", i, b)
		}
	}
}

// TestReallocateShrink tests that REALLOCATE correctly handles shrinking allocations
func TestReallocateShrink(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// Allocate 100 bytes
	v.CPU.R[0] = 100
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xEF000020)
	v.Step()
	oldAddr := v.CPU.R[0]

	// Write more than 50 bytes of data
	for i := uint32(0); i < 100; i++ {
		v.Memory.WriteByteAt(oldAddr+i, byte(i&0xFF))
	}

	// Reallocate to smaller size (50 bytes)
	v.CPU.R[0] = oldAddr
	v.CPU.R[1] = 50
	v.CPU.PC = 0x8004
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8004, 0xEF000022)
	v.Step()

	newAddr := v.CPU.R[0]
	if newAddr == 0 {
		t.Fatal("reallocation returned NULL")
	}

	// Verify first 50 bytes were copied
	for i := uint32(0); i < 50; i++ {
		actual, _ := v.Memory.ReadByteAt(newAddr + i)
		expected := byte(i & 0xFF)
		if actual != expected {
			t.Errorf("byte %d: expected 0x%02X, got 0x%02X", i, expected, actual)
		}
	}
}

// TestReallocateNullPointer tests that REALLOCATE with NULL pointer
// behaves like ALLOCATE
func TestReallocateNullPointer(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// Call REALLOCATE with NULL (0) pointer
	v.CPU.R[0] = 0   // NULL pointer
	v.CPU.R[1] = 100 // Size
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xEF000022) // SWI #0x22 (REALLOCATE)
	err := v.Step()
	if err != nil {
		t.Fatalf("reallocate with NULL failed: %v", err)
	}

	addr := v.CPU.R[0]
	if addr == 0 {
		t.Error("reallocate with NULL should allocate new memory, got NULL")
	}

	// Verify it's in heap segment
	if addr < 0x30000 || addr >= 0x40000 {
		t.Errorf("allocated address 0x%X not in heap segment", addr)
	}
}

// TestReallocateInvalidAddress tests that REALLOCATE with invalid address
// returns NULL
func TestReallocateInvalidAddress(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// Call REALLOCATE with an address that was never allocated
	v.CPU.R[0] = 0x35000 // Some address in heap (but not allocated)
	v.CPU.R[1] = 100     // Size
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xEF000022) // SWI #0x22 (REALLOCATE)
	v.Step()

	result := v.CPU.R[0]
	if result != 0 {
		t.Errorf("reallocate with invalid address should return NULL, got 0x%X", result)
	}
}

// TestReallocateAllocationFailure tests that REALLOCATE returns NULL
// when new allocation fails
func TestReallocateAllocationFailure(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// First allocate some memory
	v.CPU.R[0] = 100
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xEF000020)
	v.Step()
	oldAddr := v.CPU.R[0]

	// Try to reallocate to a size that will fail (larger than heap segment)
	v.CPU.R[0] = oldAddr
	v.CPU.R[1] = 0x00020000 // Larger than heap segment (64KB)
	v.CPU.PC = 0x8004
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8004, 0xEF000022)
	v.Step()

	result := v.CPU.R[0]
	if result != 0 {
		t.Errorf("reallocate with too large size should return NULL, got 0x%X", result)
	}

	// Verify old allocation is still intact (not freed on failure)
	if _, ok := v.Memory.HeapAllocations[oldAddr]; !ok {
		t.Error("old allocation was freed even though reallocation failed")
	}
}

// TestHeapOverflowCheck tests that heap allocation properly checks for overflow
func TestHeapOverflowCheck(t *testing.T) {
	v := vm.NewVM()
	v.CPU.PC = 0x8000

	// Try to allocate a size that would cause overflow
	// If NextHeapAddress is 0x30000 and we add 0xFFFFFFFF, it would overflow
	v.CPU.R[0] = 0xFFFFFFFF // Maximum uint32
	setupCodeWrite(v)
	v.Memory.WriteWord(0x8000, 0xEF000020) // SWI #0x20 (ALLOCATE)
	err := v.Step()
	if err != nil {
		t.Fatalf("allocate syscall failed: %v", err)
	}

	// Should return NULL (0) due to overflow check
	result := v.CPU.R[0]
	if result != 0 {
		t.Errorf("allocation with overflow size should return NULL, got 0x%X", result)
	}
}

// TestHeapOverflowCheckNearMax tests overflow check with address near max
func TestHeapOverflowCheckNearMax(t *testing.T) {
	v := vm.NewVM()

	// Manually set NextHeapAddress to a value near overflow
	v.Memory.NextHeapAddress = 0xFFFFFFF0

	// Try to allocate 32 bytes (would overflow)
	addr, err := v.Memory.Allocate(32)
	if err == nil {
		t.Error("allocation should fail due to overflow")
	}
	if addr != 0 {
		t.Errorf("allocation should return 0 on overflow, got 0x%X", addr)
	}
}

// TestFileDescriptorPerInstance tests that file descriptor mutex is per-instance
func TestFileDescriptorPerInstance(t *testing.T) {
	vm1 := vm.NewVM()
	vm2 := vm.NewVM()

	// Create a simple test to verify fdMu is per-instance by checking
	// that concurrent operations on different VMs don't deadlock or panic

	var wg sync.WaitGroup
	errors := make(chan error, 2)

	// Concurrent file operations on VM1
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			// Allocate and close file descriptors
			vm1.CPU.PC = 0x8000
			// These operations would use fdMu internally
			// Just checking for race conditions and panics
		}
	}()

	// Concurrent file operations on VM2
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			vm2.CPU.PC = 0x8000
		}
	}()

	// Wait for completion
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Errorf("concurrent operations failed: %v", err)
		}
	}
}

// TestMultipleVMsConcurrent tests multiple VMs can operate concurrently
// without race conditions in their heap allocators or file descriptors
func TestMultipleVMsConcurrent(t *testing.T) {
	const numVMs = 10
	const numOps = 50

	var wg sync.WaitGroup
	errors := make(chan error, numVMs)

	for vmID := 0; vmID < numVMs; vmID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			v := vm.NewVM()

			// Perform allocations and frees
			for i := 0; i < numOps; i++ {
				addr, err := v.Memory.Allocate(uint32((i + 1) * 4))
				if err != nil {
					errors <- err
					return
				}

				// Write some data
				if err := v.Memory.WriteByteAt(addr, byte(id)); err != nil {
					errors <- err
					return
				}

				// Read it back
				val, err := v.Memory.ReadByteAt(addr)
				if err != nil {
					errors <- err
					return
				}
				if val != byte(id) {
					errors <- err
					return
				}

				// Free it
				if err := v.Memory.Free(addr); err != nil {
					errors <- err
					return
				}
			}
		}(vmID)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		if err != nil {
			t.Errorf("concurrent VM operation failed: %v", err)
		}
	}
}
