# Filesystem Restriction Design

**Date:** 2025-11-11
**Status:** Approved for implementation

## Overview

Restrict guest program file access to a specified directory to prevent unrestricted host filesystem access.

## Problem Statement

The ARM emulator currently grants guest programs full access to the host filesystem through file operation syscalls (OPEN, READ, WRITE, etc.). This is a security concern as malicious or buggy assembly programs can read/write arbitrary files on the host system.

## Solution

Implement filesystem sandboxing by restricting all file operations to a specified root directory.

## Design

### 1. Command-Line Interface

**New flag:**
```
-fsroot <path>     Restrict file operations to this directory (default: current working directory)
```

**Behavior:**
- If `-fsroot` is not specified, defaults to current working directory
- The fsroot path is resolved to absolute canonical path at startup
- Stored in VM instance (`vm.FilesystemRoot string`)
- Examples:
  - `./arm-emulator -fsroot /tmp/sandbox program.s` - restrict to /tmp/sandbox
  - `./arm-emulator -fsroot ./test_files program.s` - restrict to ./test_files (absolute)
  - `./arm-emulator program.s` - restrict to current working directory

**Implementation location:**
- Add flag to `main.go` flag definitions (line ~27-64)
- Resolve to absolute path using `filepath.Abs()`
- Pass to VM during initialization

### 2. Path Validation and Security

**Validation function:**
```go
// In vm/syscall.go
func (vm *VM) ValidatePath(path string) (string, error)
```

**Validation rules (in order):**
0. Check FilesystemRoot is configured (required - no unrestricted access mode)
1. Check path is non-empty
2. Block paths containing `..` components (anywhere)
3. Strip leading `/` if present (treat absolute paths as relative to fsroot)
4. Join with vm.FilesystemRoot using `filepath.Join()`
5. Canonicalize with `filepath.Clean()`
6. Check for symlinks using `filepath.EvalSymlinks()` - reject if symlink detected
7. Verify canonical path starts with fsroot (has fsroot as prefix)
8. Return validated absolute path or error

**Security considerations:**
- Validation happens before any file open operation
- `filepath.EvalSymlinks()` detects symlinks and returns error
- Race condition between check-then-open is acceptable for emulator use case
- Empty filename returns error
- Relative paths from guest are resolved relative to fsroot

### 3. Integration with Syscalls

**Affected syscall:**
- `SWI_OPEN (0x10)` - Only syscall that takes a path parameter

**Modified flow in handleOpen():**
```
1. Read filename from memory (existing code, line ~615-646)
2. **NEW: Validate path**
   - Call vm.validatePath(filename)
   - If validation fails: return fmt.Errorf("filesystem access denied: %s outside %s", path, vm.FilesystemRoot)
   - VM halts immediately (security violation, not guest-recoverable)
3. Get mode from R1 (existing code)
4. Open validated file with os.Open/os.OpenFile (existing code)
5. Allocate fd and return (existing code)
```

**Unchanged syscalls** (operate on file descriptors only):
- SWI_CLOSE, SWI_READ, SWI_WRITE, SWI_SEEK, SWI_TELL, SWI_FILE_SIZE

**Standard file descriptors:**
- stdin/stdout/stderr remain unrestricted (no path validation needed)

### 4. Error Handling

**Validation failures:**
- Type: VM-level security violation
- Action: Halt execution with error (like memory corruption)
- Error format: `"filesystem access denied: attempted to access %s outside allowed root %s"`
- Not guest-recoverable (doesn't return error code to R0)

**Rationale:**
Filesystem escape attempts are security violations, not expected runtime errors. Similar to address wraparound errors in string reading.

### 5. Testing Strategy

**Unit tests** (`tests/unit/vm/filesystem_test.go`):
- Valid paths within fsroot
- Paths with `..` components (VM halts)
- Symlinks (VM halts)
- Empty paths (VM halts)
- Absolute paths treated as relative to fsroot
- Paths at fsroot boundary
- Standard fds remain accessible

**Integration tests** (`tests/integration/`):
- Assembly program with file operations
- Test valid access within fsroot
- Test escape attempts with `..`
- Verify error messages

**Existing test updates:**
- Example programs without file I/O: no changes needed
- Example programs with file I/O: update test harness to pass appropriate `-fsroot`
- Table-driven tests in `examples_test.go` need fsroot configuration

### 6. Documentation Updates

**Files to update:**
1. `README.md` - Add filesystem security section
2. `CLAUDE.md` - Document `-fsroot` flag and testing requirements
3. `main.go` - Update help text and examples

## Implementation Phases

### Phase 1: Core validation
- Add `FilesystemRoot` field to VM struct
- Implement `validatePath()` function
- Add unit tests for path validation

### Phase 2: Syscall integration
- Modify `handleOpen()` to use validation
- Update error messages
- Test OPEN syscall with various paths

### Phase 3: CLI and configuration
- Add `-fsroot` flag to main.go
- Resolve and pass to VM
- Update help text

### Phase 4: Testing and documentation
- Integration tests
- Update existing test framework
- Update documentation

## Security Properties

**Guarantees:**
- Guest programs cannot access files outside fsroot
- Path traversal with `..` is blocked
- Symlink escape attempts are blocked
- Absolute paths are treated as relative to fsroot

**Non-goals:**
- This doesn't sandbox other operations (network, process spawning, etc.)
- This doesn't provide resource limits (disk space, file size - already limited to 1MB)
- This doesn't prevent timing attacks or covert channels

## Security Policy

**Filesystem sandboxing is mandatory:**
- All file operations require FilesystemRoot to be configured
- No backward compatibility mode with unrestricted access
- Default behavior: restrict to current working directory (CWD)
- CLI always sets fsroot (defaults to CWD if not specified)

**For programmatic use (direct VM instantiation):**
- Must explicitly set vm.FilesystemRoot before any file operations
- File operations without FilesystemRoot will halt the VM with error
- Test suite automatically sets appropriate fsroot for each test

## Open Questions

None - design approved for implementation.
