package vm_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

const sandboxNameAddr = vm.DataSegmentStart

// guestOpen runs SWI OPEN on name with mode and returns R0.
func guestOpen(t *testing.T, root, name string, mode uint32) uint32 {
	t.Helper()
	v := vm.NewVM()
	v.FilesystemRoot = root
	v.CPU.PC = vm.CodeSegmentStart
	if err := v.Memory.LoadBytes(sandboxNameAddr, append([]byte(name), 0)); err != nil {
		t.Fatal(err)
	}
	v.CPU.R[0], v.CPU.R[1] = sandboxNameAddr, mode
	stepSWI(t, v, swiOpen)
	return v.CPU.R[0]
}

func TestOpenThroughDanglingSymlinkStaysInRoot(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	target := filepath.Join(outside, "created.txt")
	if err := os.Symlink(target, filepath.Join(root, "link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	m := vm.NewVM()
	m.FilesystemRoot = root
	if _, err := m.ValidatePath("link"); err == nil {
		t.Error("ValidatePath accepted a dangling symlink that points outside the root")
	}

	for _, mode := range []uint32{vm.FileModeWrite, vm.FileModeAppend} {
		if fd := guestOpen(t, root, "link", mode); fd != vm.SyscallErrorGeneral {
			t.Errorf("OPEN mode %d returned fd %d, want error", mode, fd)
		}
		if _, err := os.Lstat(target); err == nil {
			t.Fatalf("OPEN mode %d created %s outside the root", mode, target)
		}
	}
}

func TestOpenThroughSymlinkedDirectoryOutsideRootFails(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	for _, mode := range []uint32{vm.FileModeRead, vm.FileModeWrite} {
		if fd := guestOpen(t, root, "escape/secret.txt", mode); fd != vm.SyscallErrorGeneral {
			t.Errorf("OPEN mode %d returned fd %d, want error", mode, fd)
		}
	}
	if data, _ := os.ReadFile(filepath.Join(outside, "secret.txt")); string(data) != "secret" {
		t.Errorf("outside file changed to %q", data)
	}
}

func TestOpenInsideRootSucceeds(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "dir", "in.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("dir/in.txt", filepath.Join(root, "inner-link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	tests := []struct {
		name string
		path string
		mode uint32
	}{
		{"read existing", "dir/in.txt", vm.FileModeRead},
		{"leading slash is root-relative", "/dir/in.txt", vm.FileModeRead},
		{"symlink within root", "inner-link", vm.FileModeRead},
		{"create new file", "new.txt", vm.FileModeWrite},
		{"append creates", "dir/log.txt", vm.FileModeAppend},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if fd := guestOpen(t, root, tt.path, tt.mode); fd == vm.SyscallErrorGeneral || fd < vm.FirstUserFD {
				t.Errorf("OPEN %q returned 0x%08X, want a user fd", tt.path, fd)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(root, "new.txt")); err != nil {
		t.Errorf("new.txt not created in root: %v", err)
	}
}
