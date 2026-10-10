package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func TestMemoryResetDropsAddedSegments(t *testing.T) {
	m := vm.NewMemory()
	const lowSize = 0x1000
	m.AddSegment("low-memory", 0, lowSize, vm.PermRead|vm.PermWrite|vm.PermExecute)
	if err := m.WriteByteAt(0, 1); err != nil {
		t.Fatalf("setup: %v", err)
	}

	m.Reset()

	if _, err := m.ReadByteAt(0); err == nil {
		t.Error("address 0 still mapped after reset")
	}
	if got := len(m.Segments); got != len(vm.NewMemory().Segments) {
		t.Errorf("segments = %d after reset, want %d", got, len(vm.NewMemory().Segments))
	}
}

func TestMemoryResetRestoresCodePermissions(t *testing.T) {
	m := vm.NewMemory()
	m.MakeCodeReadOnly()

	m.Reset()

	if err := m.WriteWord(vm.CodeSegmentStart, 1); err != nil {
		t.Errorf("code segment not writable after reset: %v", err)
	}
}
