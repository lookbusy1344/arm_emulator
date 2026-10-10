package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func mustAlloc(t *testing.T, m *vm.Memory, size uint32) uint32 {
	t.Helper()
	addr, err := m.Allocate(size)
	if err != nil {
		t.Fatalf("Allocate(%d): %v", size, err)
	}
	return addr
}

func mustFree(t *testing.T, m *vm.Memory, addr uint32) {
	t.Helper()
	if err := m.Free(addr); err != nil {
		t.Fatalf("Free(0x%08X): %v", addr, err)
	}
}

func expectAddr(t *testing.T, what string, got, want uint32) {
	t.Helper()
	if got != want {
		t.Errorf("%s at 0x%08X, want 0x%08X", what, got, want)
	}
}

func TestHeapAllocationCanFillHeap(t *testing.T) {
	m := vm.NewMemory()

	expectAddr(t, "whole-heap block", mustAlloc(t, m, vm.HeapSegmentSize), vm.HeapSegmentStart)

	if _, err := m.Allocate(4); err == nil {
		t.Error("allocation succeeded with the heap full")
	}
}

func TestHeapReusesFreedBlock(t *testing.T) {
	m := vm.NewMemory()
	a := mustAlloc(t, m, 16)
	mustAlloc(t, m, 16)
	mustFree(t, m, a)

	expectAddr(t, "reused block", mustAlloc(t, m, 16), vm.HeapSegmentStart)
}

func TestHeapMallocFreeLoopDoesNotExhaustHeap(t *testing.T) {
	m := vm.NewMemory()
	const blockSize = 0x1000
	const iterations = 4 * vm.HeapSegmentSize / blockSize
	for i := range iterations {
		addr, err := m.Allocate(blockSize)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		mustFree(t, m, addr)
	}
}

func TestHeapSplitsFreeBlock(t *testing.T) {
	m := vm.NewMemory()
	a := mustAlloc(t, m, 16) // 0x30000
	mustAlloc(t, m, 16)      // 0x30010
	mustFree(t, m, a)

	expectAddr(t, "first half", mustAlloc(t, m, 8), vm.HeapSegmentStart)
	expectAddr(t, "second half", mustAlloc(t, m, 8), vm.HeapSegmentStart+8)
	expectAddr(t, "fresh block", mustAlloc(t, m, 8), vm.HeapSegmentStart+32)
}

func TestHeapCoalescesAdjacentFreeBlocks(t *testing.T) {
	m := vm.NewMemory()
	a := mustAlloc(t, m, 16) // 0x30000
	b := mustAlloc(t, m, 16) // 0x30010
	mustAlloc(t, m, 16)      // 0x30020, keeps the pair below the top
	mustFree(t, m, b)
	mustFree(t, m, a)

	expectAddr(t, "merged block", mustAlloc(t, m, 32), vm.HeapSegmentStart)
}

func TestHeapFreeingTopBlockReturnsSpace(t *testing.T) {
	m := vm.NewMemory()
	mustAlloc(t, m, 16)      // 0x30000
	b := mustAlloc(t, m, 16) // 0x30010
	mustFree(t, m, b)

	expectAddr(t, "block over the freed top", mustAlloc(t, m, 32), vm.HeapSegmentStart+16)
	expectAddr(t, "next heap address", m.NextHeapAddress, vm.HeapSegmentStart+48)
}

func TestHeapLargerRequestSkipsSmallFreeBlock(t *testing.T) {
	m := vm.NewMemory()
	a := mustAlloc(t, m, 8) // 0x30000
	mustAlloc(t, m, 8)      // 0x30008
	mustFree(t, m, a)

	expectAddr(t, "large block", mustAlloc(t, m, 16), vm.HeapSegmentStart+16)
	expectAddr(t, "small block", mustAlloc(t, m, 8), vm.HeapSegmentStart)
}

func TestHeapReusedBlockIsZeroed(t *testing.T) {
	m := vm.NewMemory()
	a := mustAlloc(t, m, 4)
	mustAlloc(t, m, 4)
	if err := m.WriteWord(a, 0xDEADBEEF); err != nil {
		t.Fatal(err)
	}
	mustFree(t, m, a)

	b := mustAlloc(t, m, 4)
	got, err := m.ReadWord(b)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Errorf("reused block holds 0x%08X, want 0", got)
	}
}

func TestHeapDoubleFreeFails(t *testing.T) {
	m := vm.NewMemory()
	a := mustAlloc(t, m, 16)
	mustAlloc(t, m, 16)
	mustFree(t, m, a)

	if err := m.Free(a); err == nil {
		t.Error("second free of the same block succeeded")
	}
}

func TestHeapResetForgetsFreeBlocks(t *testing.T) {
	m := vm.NewMemory()
	mustAlloc(t, m, 16)
	b := mustAlloc(t, m, 16)
	mustAlloc(t, m, 16)
	mustFree(t, m, b)

	m.ResetHeap()

	expectAddr(t, "first block after reset", mustAlloc(t, m, 32), vm.HeapSegmentStart)
	expectAddr(t, "second block after reset", mustAlloc(t, m, 16), vm.HeapSegmentStart+32)
}

func TestHeapCoalescesWithPrecedingFreeBlock(t *testing.T) {
	m := vm.NewMemory()
	a := mustAlloc(t, m, 16) // 0x30000
	b := mustAlloc(t, m, 16) // 0x30010
	mustAlloc(t, m, 16)      // 0x30020, keeps the pair below the top
	mustFree(t, m, a)
	mustFree(t, m, b)

	expectAddr(t, "merged block", mustAlloc(t, m, 32), vm.HeapSegmentStart)
}
