package vm_test

import (
	"strings"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

func guestRead(t *testing.T, v *vm.VM, fd, length uint32) uint32 {
	t.Helper()
	v.CPU.R[0], v.CPU.R[1], v.CPU.R[2] = fd, readBufAddr, length
	stepSWI(t, v, swiRead)
	return v.CPU.R[0]
}

func TestReadReturnsZeroAtEndOfFile(t *testing.T) {
	const requestMoreThanFile = 16
	v, fd := openTestFile(t, "abcdefgh")
	if got := guestRead(t, v, fd, requestMoreThanFile); got != 8 {
		t.Fatalf("first READ = %d, want 8", got)
	}
	for i := range 2 {
		if got := guestRead(t, v, fd, requestMoreThanFile); got != 0 {
			t.Errorf("READ %d at EOF = 0x%08X, want 0", i, got)
		}
	}
}

func TestReadEmptyFileReturnsZero(t *testing.T) {
	v, fd := openTestFile(t, "")
	if got := guestRead(t, v, fd, 4); got != 0 {
		t.Errorf("READ of empty file = 0x%08X, want 0", got)
	}
}

func TestReadStdinAtEndOfInputReturnsZero(t *testing.T) {
	v := vm.NewVM()
	v.SetStdinReader(strings.NewReader(""))
	v.CPU.PC = vm.CodeSegmentStart
	if got := guestRead(t, v, vm.StdIn, 4); got != 0 {
		t.Errorf("READ of exhausted stdin = 0x%08X, want 0", got)
	}
}

func TestReadErrorsStillReturnMinusOne(t *testing.T) {
	const unopenedFD = 9
	t.Run("unopened descriptor", func(t *testing.T) {
		v := vm.NewVM()
		v.CPU.PC = vm.CodeSegmentStart
		if got := guestRead(t, v, unopenedFD, 4); got != vm.SyscallErrorGeneral {
			t.Errorf("READ = 0x%08X, want error", got)
		}
	})
	t.Run("write-only file", func(t *testing.T) {
		v := vm.NewVM()
		v.FilesystemRoot = t.TempDir()
		v.CPU.PC = vm.CodeSegmentStart
		if err := v.Memory.LoadBytes(sandboxNameAddr, []byte("out.txt\x00")); err != nil {
			t.Fatal(err)
		}
		v.CPU.R[0], v.CPU.R[1] = sandboxNameAddr, vm.FileModeWrite
		stepSWI(t, v, swiOpen)
		if got := guestRead(t, v, v.CPU.R[0], 4); got != vm.SyscallErrorGeneral {
			t.Errorf("READ = 0x%08X, want error", got)
		}
	})
}
