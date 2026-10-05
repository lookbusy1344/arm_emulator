package vm_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

const (
	swiSeek     = 0xEF000014
	swiFileSize = 0xEF000016

	seekStart   = 0
	seekCurrent = 1
	seekEnd     = 2
	badWhence   = 3

	readBufAddr = vm.DataSegmentStart + 0x100
)

// openTestFile writes content to a file in a fresh root and opens it in the guest.
func openTestFile(t *testing.T, content string) (*vm.VM, uint32) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "f.txt"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	v := vm.NewVM()
	v.FilesystemRoot = root
	v.CPU.PC = vm.CodeSegmentStart
	if err := v.Memory.LoadBytes(sandboxNameAddr, []byte("f.txt\x00")); err != nil {
		t.Fatal(err)
	}
	v.CPU.R[0], v.CPU.R[1] = sandboxNameAddr, vm.FileModeRead
	stepSWI(t, v, swiOpen)
	fd := v.CPU.R[0]
	if fd == vm.SyscallErrorGeneral {
		t.Fatal("OPEN failed")
	}
	return v, fd
}

func guestSeek(t *testing.T, v *vm.VM, fd uint32, offset int32, whence uint32) uint32 {
	t.Helper()
	v.CPU.R[0], v.CPU.R[1], v.CPU.R[2] = fd, uint32(offset), whence
	stepSWI(t, v, swiSeek)
	return v.CPU.R[0]
}

func TestSeek(t *testing.T) {
	const content = "abcdefgh"
	const errResult = vm.SyscallErrorGeneral
	tests := []struct {
		name   string
		offset int32
		whence uint32
		want   uint32
	}{
		{"start", 3, seekStart, 3},
		{"start zero", 0, seekStart, 0},
		{"end", 0, seekEnd, 8},
		{"end negative", -2, seekEnd, 6},
		{"end to start", -8, seekEnd, 0},
		{"past end", 4, seekEnd, 12},
		{"before start", -9, seekEnd, errResult},
		{"negative from start", -1, seekStart, errResult},
		{"invalid whence", 0, badWhence, errResult},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, fd := openTestFile(t, content)
			if got := guestSeek(t, v, fd, tt.offset, tt.whence); got != tt.want {
				t.Errorf("SEEK(%d, %d) = 0x%08X, want 0x%08X", tt.offset, tt.whence, got, tt.want)
			}
		})
	}
}

func TestSeekCurrentNegativeThenRead(t *testing.T) {
	v, fd := openTestFile(t, "abcdefgh")
	guestSeek(t, v, fd, 5, seekStart)
	if got := guestSeek(t, v, fd, -3, seekCurrent); got != 2 {
		t.Fatalf("SEEK(-3, current) = %d, want 2", got)
	}

	v.CPU.R[0] = fd
	stepSWI(t, v, swiTell)
	if v.CPU.R[0] != 2 {
		t.Errorf("TELL = %d, want 2", v.CPU.R[0])
	}

	v.CPU.R[0], v.CPU.R[1], v.CPU.R[2] = fd, readBufAddr, 1
	stepSWI(t, v, swiRead)
	got, err := v.Memory.ReadByteAt(readBufAddr)
	if err != nil || got != 'c' {
		t.Errorf("read after seek = %q (%v), want 'c'", got, err)
	}
}

func TestFileSizeKeepsPosition(t *testing.T) {
	v, fd := openTestFile(t, "abcdefgh")
	guestSeek(t, v, fd, 3, seekStart)

	v.CPU.R[0] = fd
	stepSWI(t, v, swiFileSize)
	if v.CPU.R[0] != 8 {
		t.Errorf("FILE_SIZE = %d, want 8", v.CPU.R[0])
	}

	v.CPU.R[0] = fd
	stepSWI(t, v, swiTell)
	if v.CPU.R[0] != 3 {
		t.Errorf("TELL after FILE_SIZE = %d, want 3", v.CPU.R[0])
	}
}

func TestSeekTellFileSizeBadDescriptor(t *testing.T) {
	const unopenedFD = 9
	for name, opcode := range map[string]uint32{"SEEK": swiSeek, "TELL": swiTell, "FILE_SIZE": swiFileSize} {
		t.Run(name, func(t *testing.T) {
			v := vm.NewVM()
			v.CPU.PC = vm.CodeSegmentStart
			v.CPU.R[0] = unopenedFD
			stepSWI(t, v, opcode)
			if v.CPU.R[0] != vm.SyscallErrorGeneral {
				t.Errorf("R0 = 0x%08X, want error", v.CPU.R[0])
			}
		})
	}
}
