package vm_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// openGuestFile runs SWI OPEN on v and returns R0.
func openGuestFile(t *testing.T, v *vm.VM, name string, mode uint32) uint32 {
	t.Helper()
	v.CPU.PC = vm.CodeSegmentStart
	if err := v.Memory.LoadBytes(sandboxNameAddr, append([]byte(name), 0)); err != nil {
		t.Fatal(err)
	}
	v.CPU.R[0], v.CPU.R[1] = sandboxNameAddr, mode
	stepSWI(t, v, swiOpen)
	return v.CPU.R[0]
}

// vmWithRoot returns a VM whose filesystem root is dir, closed when the test ends.
func vmWithRoot(t *testing.T, dir string) *vm.VM {
	t.Helper()
	v := vm.NewVM()
	if err := v.OpenFilesystemRoot(dir); err != nil {
		t.Fatalf("OpenFilesystemRoot(%q): %v", dir, err)
	}
	t.Cleanup(v.Close)
	return v
}

func TestFilesystemRootReportsDirectory(t *testing.T) {
	dir := t.TempDir()
	v := vmWithRoot(t, dir)

	if got := v.FilesystemRoot(); got != dir {
		t.Errorf("FilesystemRoot() = %q, want %q", got, dir)
	}
}

func TestFilesystemRootIsEmptyWithoutRoot(t *testing.T) {
	if got := vm.NewVM().FilesystemRoot(); got != "" {
		t.Errorf("FilesystemRoot() = %q, want empty", got)
	}
}

func TestOpenFilesystemRootRejectsMissingDirectory(t *testing.T) {
	v := vm.NewVM()
	missing := filepath.Join(t.TempDir(), "missing")

	if err := v.OpenFilesystemRoot(missing); err == nil {
		t.Fatal("OpenFilesystemRoot accepted a missing directory")
	}
	if got := v.FilesystemRoot(); got != "" {
		t.Errorf("FilesystemRoot() = %q, want empty", got)
	}
}

// Replacing the root directory with a symlink after setup must not redirect the guest.
func TestOpenUsesRootDirectoryHeldAtSetup(t *testing.T) {
	parent, outside := t.TempDir(), t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	v := vmWithRoot(t, root)

	moved := filepath.Join(parent, "moved")
	if err := os.Rename(root, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, root); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if fd := openGuestFile(t, v, "secret.txt", vm.FileModeRead); fd != vm.SyscallErrorGeneral {
		t.Errorf("OPEN secret.txt returned fd %d, want error", fd)
	}
	if fd := openGuestFile(t, v, "new.txt", vm.FileModeWrite); fd == vm.SyscallErrorGeneral {
		t.Fatal("OPEN new.txt for write failed")
	}
	if _, err := os.Stat(filepath.Join(moved, "new.txt")); err != nil {
		t.Errorf("new.txt not created in the original root directory: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(outside, "new.txt")); err == nil {
		t.Error("new.txt created outside the root")
	}
}

func TestCloseReleasesFilesystemRoot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "in.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	v := vmWithRoot(t, dir)

	v.Close()

	if got := v.FilesystemRoot(); got != "" {
		t.Errorf("FilesystemRoot() after Close = %q, want empty", got)
	}
	if fd := openGuestFile(t, v, "in.txt", vm.FileModeRead); fd != vm.SyscallErrorGeneral {
		t.Errorf("OPEN after Close returned fd %d, want error", fd)
	}
}

func TestResetKeepsFilesystemRoot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "in.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	v := vmWithRoot(t, dir)

	v.Reset()

	if fd := openGuestFile(t, v, "in.txt", vm.FileModeRead); fd < vm.FirstUserFD || fd == vm.SyscallErrorGeneral {
		t.Errorf("OPEN after Reset returned 0x%08X, want a user fd", fd)
	}
}

func TestValidatePathAcceptsNamesStartingWithDots(t *testing.T) {
	dir := t.TempDir()
	v := vmWithRoot(t, dir)

	got, err := v.ValidatePath("..foo")
	if err != nil {
		t.Fatalf("ValidatePath(\"..foo\"): %v", err)
	}
	if want := filepath.Join(dir, "..foo"); got != want {
		t.Errorf("ValidatePath(\"..foo\") = %q, want %q", got, want)
	}
}
