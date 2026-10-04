package vm_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

const (
	swiOpen  = 0xEF000010
	swiWrite = 0xEF000013
)

// stepSWI writes an SWI at the current PC and executes it.
func stepSWI(t *testing.T, v *vm.VM, opcode uint32) {
	t.Helper()
	setupCodeWrite(v)
	if err := v.Memory.WriteWord(v.CPU.PC, opcode); err != nil {
		t.Fatal(err)
	}
	if err := v.Step(); err != nil {
		t.Fatalf("SWI 0x%08X failed: %v", opcode, err)
	}
}

func TestResetKeepsStandardFileDescriptors(t *testing.T) {
	v := vm.NewVM()
	v.Reset()
	v.CPU.PC = 0x8000

	// Zero-length write to stderr goes through the descriptor table.
	v.CPU.R[0] = vm.StdErr
	v.CPU.R[1] = 0x20000
	v.CPU.R[2] = 0
	stepSWI(t, v, swiWrite)
	if v.CPU.R[0] != 0 {
		t.Errorf("write to stderr after Reset: expected R0=0, got 0x%08X", v.CPU.R[0])
	}
}

func TestResetOpenReturnsUserDescriptor(t *testing.T) {
	dir := t.TempDir()
	const name = "f.txt"
	if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	v := vm.NewVM()
	v.Reset()
	v.FilesystemRoot = dir
	v.CPU.PC = 0x8000

	const nameAddr = 0x20000
	setupDataWrite(v)
	if err := v.Memory.LoadBytes(nameAddr, append([]byte(name), 0)); err != nil {
		t.Fatal(err)
	}
	v.CPU.R[0] = nameAddr
	v.CPU.R[1] = vm.FileModeRead
	stepSWI(t, v, swiOpen)

	if fd := v.CPU.R[0]; fd == vm.SyscallErrorGeneral || fd < vm.FirstUserFD {
		t.Errorf("open after Reset: expected fd >= %d, got 0x%08X", vm.FirstUserFD, fd)
	}
}
