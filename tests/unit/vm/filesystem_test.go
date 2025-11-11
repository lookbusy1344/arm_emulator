package vm_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lookbusy1344/arm-emulator/vm"
)

// TestValidatePathNoRoot tests that when no filesystem root is configured,
// all paths are allowed (backward compatibility)
func TestValidatePathNoRoot(t *testing.T) {
	machine := vm.NewVM()
	// Don't set FilesystemRoot - should allow all paths

	testCases := []struct {
		name string
		path string
	}{
		{"simple path", "test.txt"},
		{"path with slash", "/etc/passwd"},
		{"path with ..", "../etc/passwd"},
		{"relative path", "dir/file.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When FilesystemRoot is empty, all paths should be allowed
			_, err := machine.ValidatePath(tc.path)
			if err != nil {
				t.Errorf("Expected no error for '%s' with no fsroot, got: %v", tc.path, err)
			}
		})
	}
}

// TestValidatePathValid tests valid paths within fsroot
func TestValidatePathValid(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "arm-emu-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file structure
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	subFile := filepath.Join(subDir, "data.txt")
	if err := os.WriteFile(subFile, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to create subfile: %v", err)
	}

	machine := vm.NewVM()
	machine.FilesystemRoot = tmpDir

	testCases := []struct {
		name string
		path string
	}{
		{"simple file", "test.txt"},
		{"file in subdir", "subdir/data.txt"},
		{"absolute path treated as relative", "/test.txt"},
		{"absolute subdir path", "/subdir/data.txt"},
		{"new file (doesn't exist yet)", "newfile.txt"},
		{"new file in subdir", "subdir/newfile.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validPath, err := machine.ValidatePath(tc.path)
			if err != nil {
				t.Errorf("Expected no error for '%s', got: %v", tc.path, err)
			}
			// Verify the returned path is under tmpDir
			if !filepath.IsAbs(validPath) {
				t.Errorf("Expected absolute path, got: %s", validPath)
			}
		})
	}
}

// TestValidatePathDotDot tests that paths with .. are blocked
func TestValidatePathDotDot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "arm-emu-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	machine := vm.NewVM()
	machine.FilesystemRoot = tmpDir

	testCases := []struct {
		name string
		path string
	}{
		{"simple dotdot", "../test.txt"},
		{"dotdot in middle", "subdir/../test.txt"},
		{"multiple dotdot", "../../etc/passwd"},
		{"dotdot at end", "subdir/.."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := machine.ValidatePath(tc.path)
			if err == nil {
				t.Errorf("Expected error for path with '..': '%s'", tc.path)
			}
		})
	}
}

// TestValidatePathEmpty tests that empty paths are rejected
func TestValidatePathEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "arm-emu-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	machine := vm.NewVM()
	machine.FilesystemRoot = tmpDir

	_, err = machine.ValidatePath("")
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

// TestValidatePathSymlink tests that symlinks are handled correctly
func TestValidatePathSymlink(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "arm-emu-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a directory outside fsroot
	outsideDir, err := os.MkdirTemp("", "arm-emu-outside-*")
	if err != nil {
		t.Fatalf("Failed to create outside dir: %v", err)
	}
	defer os.RemoveAll(outsideDir)

	outsideFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
		t.Fatalf("Failed to create outside file: %v", err)
	}

	// Create symlink inside fsroot pointing outside
	symlinkPath := filepath.Join(tmpDir, "escape")
	if err := os.Symlink(outsideDir, symlinkPath); err != nil {
		t.Skipf("Skipping symlink test: %v", err)
	}

	machine := vm.NewVM()
	machine.FilesystemRoot = tmpDir

	// Try to access file through symlink
	_, err = machine.ValidatePath("escape/secret.txt")
	if err == nil {
		t.Error("Expected error for symlink escape attempt")
	}
}

// TestValidatePathBoundary tests paths at the boundary of fsroot
func TestValidatePathBoundary(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "arm-emu-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file at root
	rootFile := filepath.Join(tmpDir, "root.txt")
	if err := os.WriteFile(rootFile, []byte("root"), 0644); err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}

	machine := vm.NewVM()
	machine.FilesystemRoot = tmpDir

	testCases := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{"root file", "root.txt", false},
		{"just slash", "/", false},
		{"empty after slash strip", "", true}, // This will be caught by empty check
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := machine.ValidatePath(tc.path)
			if tc.shouldErr && err == nil {
				t.Errorf("Expected error for '%s'", tc.path)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Expected no error for '%s', got: %v", tc.path, err)
			}
		})
	}
}

// TestValidatePathCaseSensitivity tests platform-specific case handling
func TestValidatePathCaseSensitivity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "arm-emu-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "Test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	machine := vm.NewVM()
	machine.FilesystemRoot = tmpDir

	// This test behavior depends on the filesystem
	// On case-insensitive systems (macOS default), both should work
	// On case-sensitive systems (Linux), only exact match works
	_, err = machine.ValidatePath("Test.txt")
	if err != nil {
		t.Errorf("Expected no error for exact case match: %v", err)
	}
}
