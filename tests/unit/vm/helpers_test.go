package vm_test

import (
	"io"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// Helper function to enable write permissions on code segment
func setupCodeWrite(v *vm.VM) {
	for _, seg := range v.Memory.Segments {
		if seg.Name == "code" {
			seg.Permissions = vm.PermRead | vm.PermWrite | vm.PermExecute
		}
	}
}

// Helper function to enable write permissions on data segment
func setupDataWrite(v *vm.VM) {
	for _, seg := range v.Memory.Segments {
		if seg.Name == "data" {
			seg.Permissions = vm.PermRead | vm.PermWrite
		}
	}
}

// Helper function to create a stdin pipe for testing
func createStdinPipe() (*io.PipeReader, *io.PipeWriter) {
	return io.Pipe()
}

// mustStep executes one instruction and fails the test if it returns an error.
func mustStep(t testing.TB, v *vm.VM) {
	t.Helper()
	if err := v.Step(); err != nil {
		t.Fatalf("Step at PC=0x%08X: %v", v.CPU.PC, err)
	}
}

func mustWriteWord(t testing.TB, v *vm.VM, addr, value uint32) {
	t.Helper()
	if err := v.Memory.WriteWord(addr, value); err != nil {
		t.Fatalf("WriteWord(0x%08X): %v", addr, err)
	}
}

func mustWriteHalfword(t testing.TB, v *vm.VM, addr uint32, value uint16) {
	t.Helper()
	if err := v.Memory.WriteHalfword(addr, value); err != nil {
		t.Fatalf("WriteHalfword(0x%08X): %v", addr, err)
	}
}

func mustWriteByte(t testing.TB, v *vm.VM, addr uint32, value byte) {
	t.Helper()
	if err := v.Memory.WriteByteAt(addr, value); err != nil {
		t.Fatalf("WriteByteAt(0x%08X): %v", addr, err)
	}
}
