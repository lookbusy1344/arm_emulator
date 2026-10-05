package vm_test

import (
	"os"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

const (
	swiClose = 0xEF000011
	swiTell  = 0xEF000015
)

// swapStdFile replaces the host standard file for fd with a pipe for the test's duration
// and returns the write end.
func swapStdFile(t *testing.T, fd uint32) (*os.File, *os.File) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	target := map[uint32]**os.File{vm.StdIn: &os.Stdin, vm.StdOut: &os.Stdout, vm.StdErr: &os.Stderr}[fd]
	saved := *target
	if fd == vm.StdIn {
		*target = r
	} else {
		*target = w
	}
	t.Cleanup(func() {
		*target = saved
		_ = r.Close()
		_ = w.Close()
	})
	return r, w
}

func TestCloseStandardDescriptorLeavesHostFileOpen(t *testing.T) {
	for _, fd := range []uint32{vm.StdIn, vm.StdOut, vm.StdErr} {
		t.Run(map[uint32]string{vm.StdIn: "stdin", vm.StdOut: "stdout", vm.StdErr: "stderr"}[fd], func(t *testing.T) {
			r, w := swapStdFile(t, fd)
			v := vm.NewVM()
			v.CPU.PC = vm.CodeSegmentStart

			// TELL fails on a pipe but installs the host file in the descriptor table.
			v.CPU.R[0] = fd
			stepSWI(t, v, swiTell)

			v.CPU.R[0] = fd
			stepSWI(t, v, swiClose)
			if v.CPU.R[0] != vm.SyscallErrorGeneral {
				t.Errorf("CLOSE(%d) R0 = 0x%08X, want 0x%08X", fd, v.CPU.R[0], uint32(vm.SyscallErrorGeneral))
			}

			hostFile := w
			if fd == vm.StdIn {
				hostFile = r
			}
			if _, err := hostFile.Stat(); err != nil {
				t.Errorf("host file closed by guest: %v", err)
			}
		})
	}
}
