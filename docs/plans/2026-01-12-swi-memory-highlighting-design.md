# SWI Memory Highlighting Design

**Date:** 2026-01-12
**Status:** Approved, Ready for Implementation

## Problem Statement

The Swift GUI memory view currently highlights memory writes from STR/LDR/STM/LDM instructions, but does NOT highlight memory writes from SWI syscalls. This means user input (READ_STRING), file I/O (READ), and memory operations (REALLOCATE) don't show visual feedback in the memory view.

## Current Architecture

### Memory Write Tracking

The memory view highlighting relies on three VM fields:
- `LastMemoryWrite uint32` - address that was written
- `LastMemoryWriteSize uint32` - bytes written (1, 2, or 4)
- `HasMemoryWrite bool` - boolean flag indicating a write occurred

### Where Tracking is Currently Set

**Instructions (working):**
- `vm/inst_memory.go:168-170` - STR/STRB/STRH instructions
- `vm/memory_multi.go:126-127` - STM/LDM instructions

**Syscalls (NOT working):**
- `vm/syscall.go` - SWI handlers use `vm.Memory.WriteByteAt()` directly without setting tracking flags

### GUI Integration

The Swift GUI reads these flags via:
1. `service/debugger_service.go:GetLastMemoryWrite()` - Returns `MemoryWriteInfo` struct
2. `api/handlers.go` - Includes write info in status response
3. `swift-gui/ARMEmulator/ViewModels/EmulatorViewModel.swift:94-130` - `onChange(of: viewModel.lastMemoryWrite)` triggers highlighting

## Solution Design

### Approach: Explicit Syscall-Level Tracking

**Chosen approach:** Add tracking flags explicitly in each syscall handler after memory writes.

**Rationale:**
- **Control**: Explicit choice of which operations to highlight
- **Size accuracy**: Each syscall knows exactly how many bytes it wrote
- **Simplicity**: No API changes to Memory layer, just add 3 lines after memory operations
- **Consistency**: Matches existing pattern in `inst_memory.go`

**Rejected approach:** Automatic tracking in `WriteByteAt()`
- Would require API changes to Memory layer
- Might track unwanted internal operations
- Less control over what gets highlighted

## Implementation Plan

### Syscalls to Modify

**1. READ_STRING (SWI 0x05) - handleReadString()**
- **Location:** `vm/syscall.go:459-511`
- **Trigger:** After writing user input string and null terminator
- **Tracking:**
  ```go
  vm.LastMemoryWrite = addr
  vm.LastMemoryWriteSize = bytesToWrite + 1  // Include null terminator
  vm.HasMemoryWrite = true
  ```
- **Notes:** Only set on success path, after state restored from `StateWaitingForInput`

**2. READ (SWI 0x12) - handleRead()**
- **Location:** `vm/syscall.go:856-919`
- **Trigger:** After reading file data into buffer
- **Tracking:**
  ```go
  vm.LastMemoryWrite = bufferAddr
  vm.LastMemoryWriteSize = uint32(n)  // Actual bytes read
  vm.HasMemoryWrite = true
  ```
- **Notes:** Only set on success path (when n > 0), after state restored from `StateWaitingForInput`

**3. REALLOCATE (SWI 0x22) - handleReallocate()**
- **Location:** `vm/syscall.go:1073-1136`
- **Trigger:** After copying old data to new allocation
- **Tracking:**
  ```go
  vm.LastMemoryWrite = newAddr
  vm.LastMemoryWriteSize = copySize  // Actual bytes copied
  vm.HasMemoryWrite = true
  ```
- **Notes:** Track NEW allocation address, not old; only on success path

### Edge Cases

1. **Error paths**: Only set tracking flags on SUCCESS paths (when R0 != 0xFFFFFFFF)
2. **State transitions**: Tracking should happen AFTER state is restored to `StateRunning`
3. **Size = 0 writes**: Do NOT set `HasMemoryWrite = true` when no bytes written
4. **Null terminator**: For READ_STRING, include null byte in size so it gets highlighted
5. **Partial writes**: For READ, track actual bytes read (n), not requested length

### Consistency with Instructions

Current instruction tracking:
- STR: size = 4 (word)
- STRB: size = 1 (byte)
- STRH: size = 2 (halfword)

Syscall tracking:
- READ_STRING: Variable size (bytesToWrite + 1)
- READ: Variable size (actual bytes read)
- REALLOCATE: Variable size (copySize)

**Consistent pattern:** Always track the actual number of bytes written.

## Testing Strategy

### Unit Tests (Required)

**Location:** `tests/unit/vm/syscall_test.go`

**Test cases:**
1. `TestReadStringSetsMemoryTracking` - Verify flags set correctly
2. `TestReadStringErrorNoTracking` - Verify flags NOT set on error
3. `TestReadSetsMemoryTracking` - Verify flags set for file reads
4. `TestReadFromStdinSetsMemoryTracking` - Verify flags set for stdin reads
5. `TestReallocateSetsMemoryTracking` - Verify flags set, tracks new address
6. `TestReallocateErrorNoTracking` - Verify flags NOT set on allocation failure

**Test structure:**
```go
func TestReadStringSetsMemoryTracking(t *testing.T) {
    vm := setupTestVM()
    // Setup: Write test string to stdin pipe
    // Execute: Call handleReadString via SWI #0x05
    // Assert: vm.HasMemoryWrite == true
    // Assert: vm.LastMemoryWrite == bufferAddr
    // Assert: vm.LastMemoryWriteSize == len(input) + 1
}
```

### Integration Tests (Optional)

Manual testing via Swift GUI:
1. Load program with READ_STRING → verify input highlights buffer
2. Load program with READ → verify file data highlights buffer
3. Load program with REALLOCATE → verify new allocation highlights
4. Verify highlighting clears on next step without writes

## Implementation Checklist

- [ ] Add tracking to `handleReadString()` in `vm/syscall.go`
- [ ] Add tracking to `handleRead()` in `vm/syscall.go`
- [ ] Add tracking to `handleReallocate()` in `vm/syscall.go`
- [ ] Write unit tests for READ_STRING tracking
- [ ] Write unit tests for READ tracking
- [ ] Write unit tests for REALLOCATE tracking
- [ ] Write unit tests for error paths (no tracking)
- [ ] Run all existing tests to ensure no regressions
- [ ] Manual testing in Swift GUI
- [ ] Update `PROGRESS.md` with completed work

## Success Criteria

1. All three syscalls set memory tracking flags on success
2. No tracking on error paths
3. All unit tests pass (new + existing)
4. Swift GUI memory view highlights SWI memory writes
5. Zero regressions in existing tests

## Notes

- This design follows the existing pattern from `inst_memory.go` and `memory_multi.go`
- No changes needed to GUI/API layer - they already support variable-size writes
- The `GetLastMemoryWrite()` function in `debugger_service.go` already clears the flag after reading, so no changes needed there
- Future syscalls that write memory should follow this same pattern

## References

- Memory view implementation: `swift-gui/ARMEmulator/Views/MemoryView.swift`
- Existing tracking: `vm/inst_memory.go:168-170`, `vm/memory_multi.go:126-127`
- Service layer: `service/debugger_service.go:384-397`
- VM tracking fields: `vm/executor.go:101-103`
