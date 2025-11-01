# Magic Numbers Report

**Date:** November 1, 2025
**Issue:** #37
**Scope:** Analysis of magic numbers across the ARM emulator codebase

## Executive Summary

This report identifies and categorizes magic numbers throughout the ARM emulator Go codebase (122 source files). Magic numbers are numeric literals that appear in code without clear semantic meaning, reducing code readability and maintainability.

The analysis reveals **extensive use of magic numbers** across all major subsystems:
- Memory segment addresses and sizes
- Syscall numbers and error codes
- ARM instruction encoding bit patterns
- CPSR flag bit positions
- File descriptor limits and buffer sizes
- Register numbers and counts

A rationalization plan is provided to replace these with a coherent set of named constants.

---

## Categories of Magic Numbers

### 1. Memory Architecture Constants

**Location:** `vm/memory.go`

#### Current State
```go
const (
    CodeSegmentStart  = 0x00008000 // 32KB offset
    CodeSegmentSize   = 0x00010000 // 64KB
    DataSegmentStart  = 0x00020000
    DataSegmentSize   = 0x00010000 // 64KB
    HeapSegmentStart  = 0x00030000
    HeapSegmentSize   = 0x00010000 // 64KB
    StackSegmentStart = 0x00040000
    StackSegmentSize  = 0x00010000 // 64KB
)
```

**Magic numbers in code:**
- `0x3`, `0x1` - Alignment masks and checks (lines 116, 120, 450)
- `0xFFFFFFFF`, `0xFFFFFFFC` - Address overflow checks (lines 445, 455, 685, 738, 783, 811, 844)
- `8`, `16`, `24` - Bit shift positions for endianness (lines 188, 190, 254-256, 258-261)

**Issue:** While memory segments have named constants, alignment checks, bit positions, and overflow boundaries use raw numbers.

---

### 2. Syscall Numbers

**Location:** `vm/syscall.go`

#### Current State
```go
const (
    // Console I/O
    SWI_EXIT          = 0x00
    SWI_WRITE_CHAR    = 0x01
    SWI_WRITE_STRING  = 0x02
    // ... (complete list lines 51-93)
)
```

**Good practice:** Syscall numbers are already well-defined as constants.

**Magic numbers in code:**
- `0x00FFFFFF` - Syscall number mask (line 171)
- `0xFFFFFFFF` - Error return value (appears ~40 times throughout syscall.go)
- `1024 * 1024` - Size limits (lines 36, 39-40)
- `4096` - Filename limit (line 37)
- `1024` - Various buffer limits (lines 38, 41, 144, 555)
- `256` - Default string buffer size (line 375)
- `3` - Standard file descriptor boundary (lines 114, 135)
- `0`, `1`, `2` - stdin/stdout/stderr (lines 116-121)
- `2`, `8`, `10`, `16` - Number bases for integer output (lines 326-338)
- `32`, `126` - ASCII printable range (line 578)

**Issue:** Many operational constants (error codes, limits, magic values) are hardcoded.

---

### 3. CPU and CPSR Constants

**Location:** `vm/cpu.go`

#### Current State
```go
const (
    R0  = 0
    R1  = 1
    // ... R0-R14 (lines 61-76)
    SP  = 13
    LR  = 14
)
```

**Good practice:** Register numbers are named constants.

**Magic numbers in code:**
- `31`, `30`, `29`, `28` - CPSR flag bit positions (lines 35-44, 53-56)
- `15` - Array size for general registers (line 6)
- `8` - PC pipeline offset (lines 137, 166)
- `4` - Instruction size in bytes (lines 156, 166)
- `0`, `14`, `15` - Register boundary checks (lines 136-152, 204-240)

**Issue:** CPSR bit positions and architectural constants (pipeline offset, instruction width) are hardcoded.

---

### 4. Instruction Encoding Constants

**Location:** `encoder/*.go`

#### Current State - Data Processing (encoder/data_processing.go)
```go
const (
    opAND = 0x0
    opEOR = 0x1
    opSUB = 0x2
    // ... (lines 11-28)
)
```

**Magic numbers in encoding:**
- `28`, `26`, `25`, `24`, `23`, `22`, `21`, `20`, `16`, `12`, `7`, `5`, `4` - Bit shift positions (throughout encoder files)
- `0xFFFFFF` - 24-bit mask for branch offsets (branch.go:77)
- `0x12FFF1`, `0x12FFF3` - BX/BLX encoding patterns (branch.go:104, 122)
- `0xFFF` - 12-bit offset limit (memory.go:149)
- `0x800000`, `0x7FFFFF` - 24-bit signed range (branch.go:70)
- Various instruction format masks and patterns

**Issue:** Bit positions and ARM-specific encoding patterns are scattered throughout encoding logic.

---

### 5. File and I/O Limits

**Location:** `vm/syscall.go`

#### Current State
```go
const (
    maxStringLength   = 1024 * 1024 // 1MB
    maxFilenameLength = 4096        // 4KB
    maxAssertMsgLen   = 1024        // 1KB
    maxReadSize       = 1024 * 1024 // 1MB
    maxWriteSize      = 1024 * 1024 // 1MB
    maxFDs            = 1024        // Max file descriptors
)
```

**Good practice:** Size limits are already defined as constants.

**Magic numbers in code:**
- `0644` - File permissions (lines 638, 641)
- `0`, `1`, `2` - File open modes (read/write/append) (lines 633-641)
- `0`, `1`, `2` - Seek whence values (io.SeekStart/Current/End) (line 802)

**Issue:** File operation modes and standard Unix values are hardcoded.

---

### 6. Execution and VM Constants

**Location:** `vm/executor.go`

**Magic numbers:**
- `1000000` - Default max cycles (line 111)
- `1000` - Initial instruction log capacity (line 113)
- `3` - File descriptor array size (line 118)

**Issue:** VM operational limits use raw numbers.

---

### 7. Bit Manipulation and Masks

**Scattered throughout codebase:**
- `0x1`, `0x3`, `0x7`, `0xF`, `0xFF`, `0xFFF`, `0xFFFF`, `0xFFFFFF`, `0xFFFFFFFF` - Various bit masks
- `1 << 0` through `1 << 31` - Bit position flags
- Shift amounts: `4`, `5`, `7`, `8`, `12`, `16`, `20`, `21`, `22`, `23`, `24`, `25`, `26`, `28`

**Issue:** No centralized definitions for common bit manipulation patterns.

---

## Impact Assessment

### Readability Issues
- Bit shift positions (e.g., `<< 28`) lack context about which field is being set
- Memory addresses like `0xFFFFFFFF` appear both as error codes and overflow checks
- Register counts (`15`, `14`) scattered throughout code

### Maintainability Issues
- Changing ARM architecture parameters requires hunting through multiple files
- Syscall error handling uses `0xFFFFFFFF` ~40 times
- Instruction encoding bit positions duplicated across encoder files

### Type Confusion Issues
- Easy to confuse different uses of the same magic number (e.g., `1024` as buffer size vs FD limit)
- No semantic meaning to distinguish between different numeric contexts
- Note: Named constants improve readability but don't add compile-time type safety in Go (constants are untyped)

---

## Rationalization Plan

### Phase 1: ARM Architecture Constants Package

**Create:** `vm/arch_constants.go`

```go
package vm

// ARM2 Architecture Constants
// These values are defined by the ARM2 specification and should not be modified

const (
    // Instruction encoding
    ARMInstructionSize = 4 // bytes
    ARMPipelineOffset  = 8 // PC is instruction address + 8

    // Register counts
    ARMGeneralRegisterCount = 15 // R0-R14
    ARMTotalRegisterCount   = 16 // Including PC (R15)
    ARMRegisterPC           = 15

    // CPSR flag bit positions (bits 31-28)
    CPSRBitN = 31 // Negative flag
    CPSRBitZ = 30 // Zero flag
    CPSRBitC = 29 // Carry flag
    CPSRBitV = 28 // Overflow flag

    // Instruction encoding bit positions
    InstructionConditionShift      = 28 // bits 31-28
    InstructionTypeShift           = 26 // bits 27-26
    InstructionImmediateShift      = 25 // bit 25
    InstructionPreIndexShift       = 24 // bit 24 (P bit)
    InstructionUpShift             = 23 // bit 23 (U bit)
    InstructionByteShift           = 22 // bit 22 (B bit)
    InstructionWritebackShift      = 21 // bit 21 (W bit)
    InstructionLoadShift           = 20 // bit 20 (L bit)
    InstructionRnShift             = 16 // bits 19-16
    InstructionRdShift             = 12 // bits 15-12
    InstructionRsShift             = 8  // bits 11-8
    InstructionShiftTypeShift      = 5  // bits 6-5
    InstructionShiftAmountShift    = 7  // bits 11-7

    // Bit masks
    Mask4Bit   = 0xF
    Mask8Bit   = 0xFF
    Mask12Bit  = 0xFFF
    Mask16Bit  = 0xFFFF
    Mask24Bit  = 0xFFFFFF
    Mask32Bit  = 0xFFFFFFFF

    // Alignment
    // Alignment constants (grouped together for discoverability)
    AlignmentWord        = 4          // 4-byte word alignment
    AlignmentHalfword    = 2          // 2-byte halfword alignment
    AlignmentByte        = 1          // no alignment required
    AlignMaskWord        = 0x3        // mask for word alignment check (address & mask == 0 means aligned)
    AlignMaskHalfword    = 0x1        // mask for halfword alignment check
    AlignRoundUpMaskWord = 0xFFFFFFFC // mask to round up to word alignment (~0x3)

    // Signed integer ranges (for branch offsets, etc.)
    Int24Max = 0x7FFFFF   // Maximum positive 24-bit signed value
    Int24Min = -0x800000  // Minimum negative 24-bit signed value
)

// Note: Byte shift positions (8, 16, 24) are used directly as literals - they are self-documenting

// Special instruction encoding patterns
const (
    // BX/BLX patterns are pre-positioned (NOT shift results - the trailing 0 is part of the hex value)
    // These are the actual bit patterns used directly in instruction encoding
    // Usage: instruction := (cond << 28) | BXEncodingBase | rm
    BXEncodingBase  = 0x12FFF10  // BX instruction base pattern (binary: 0001 0010 1111 1111 0001 0000)
    BLXEncodingBase = 0x12FFF30  // BLX instruction base pattern (binary: 0001 0010 1111 1111 0011 0000)
    NOPEncoding     = 0xE1A00000 // MOV R0, R0 (unconditional)
)
```

### Phase 2: System Call Constants Package

**Create:** `vm/syscall_constants.go`

```go
package vm

// Syscall Return Values
const (
    SyscallSuccess      = 0
    SyscallErrorGeneral = 0xFFFFFFFF // -1 in two's complement
    SyscallNull         = 0          // NULL pointer
)

// Syscall number extraction
const (
    SWIMask = 0x00FFFFFF // Bottom 24 bits contain syscall number
)

// File operation modes
const (
    FileModeRead   = 0 // Read-only
    FileModeWrite  = 1 // Write (create/truncate)
    FileModeAppend = 2 // Append (create/read-write)
)

// File permissions (Unix-style)
const (
    FilePermDefault = 0644 // rw-r--r--
)

// Note: For seek operations, use io.SeekStart, io.SeekCurrent, io.SeekEnd from the standard library

// Standard file descriptors
const (
    StdIn  = 0
    StdOut = 1
    StdErr = 2
    FirstUserFD = 3 // First available user FD
)

// Buffer size limits
const (
    MaxStringLength   = 1024 * 1024 // 1MB for general strings
    MaxFilenameLength = 4096        // 4KB (typical filesystem limit)
    MaxAssertMsgLen   = 1024        // 1KB for assertion messages
    MaxReadSize       = 1024 * 1024 // 1MB maximum file read
    MaxWriteSize      = 1024 * 1024 // 1MB maximum file write
    MaxFileDescriptors = 1024       // Maximum number of open FDs
    DefaultStringBuffer = 256       // Default buffer for READ_STRING
    MaxMemoryDump      = 1024       // 1KB limit for memory dumps
)

// Note: Number bases (2, 8, 10, 16) are used directly as literals - they are self-documenting

// ASCII character ranges
const (
    ASCIIPrintableMin = 32  // Space
    ASCIIPrintableMax = 126 // Tilde (~)
)
```

### Phase 3: VM Configuration Constants

**Create:** `vm/vm_constants.go`

```go
package vm

// VM Execution Limits
const (
    DefaultMaxCycles    = 1000000 // Default instruction limit
    DefaultLogCapacity  = 1000    // Initial capacity for instruction log
    DefaultFDTableSize  = 3       // stdin, stdout, stderr
)

// Memory overflow protection
const (
    Address32BitMax     = 0xFFFFFFFF // Maximum 32-bit address (also wraps on increment)
    Address32BitMaxSafe = 0xFFFFFFFC // Max address allowing 4-byte access without overflow
)
```

### Phase 4: Data Processing Opcode Constants

**Enhance:** `encoder/data_processing.go`

```go
// Data processing instruction opcodes (4-bit values)
const (
    OpcodeAND = 0x0 // Logical AND
    OpcodeEOR = 0x1 // Logical XOR
    OpcodeSUB = 0x2 // Subtract
    OpcodeRSB = 0x3 // Reverse Subtract
    OpcodeADD = 0x4 // Add
    OpcodeADC = 0x5 // Add with Carry
    OpcodeSBC = 0x6 // Subtract with Carry
    OpcodeRSC = 0x7 // Reverse Subtract with Carry
    OpcodeTST = 0x8 // Test (AND, sets flags only)
    OpcodeTEQ = 0x9 // Test Equivalence (XOR, sets flags only)
    OpcodeCMP = 0xA // Compare (SUB, sets flags only)
    OpcodeCMN = 0xB // Compare Negative (ADD, sets flags only)
    OpcodeORR = 0xC // Logical OR
    OpcodeMOV = 0xD // Move
    OpcodeBIC = 0xE // Bit Clear (AND NOT)
    OpcodeMVN = 0xF // Move Not
)
```

### Phase 5: Update Usage Throughout Codebase

**Priority files to update:**
1. `vm/cpu.go` - Replace CPSR bit positions
2. `vm/memory.go` - Replace alignment checks and endianness shifts
3. `vm/syscall.go` - Replace error codes, limits, and FD constants
4. `encoder/*.go` - Replace bit shift positions and masks
5. `vm/executor.go` - Replace VM configuration values

---

## Migration Strategy

### Step 1: Create Constants Packages (No Code Changes)
- Add new files with constant definitions
- No existing code is modified
- Verify code still builds and tests pass

### Step 2: Gradual Migration (File by File)
- Update one source file at a time
- Run tests after each file
- Use find/replace with caution for common patterns

### Step 3: Example Migration Pattern

**Before:**
```go
if address&0x3 != 0 {
    return fmt.Errorf("unaligned word access at 0x%08X", address)
}
```

**After:**
```go
if address&AlignMaskWord != 0 {
    return fmt.Errorf("unaligned word access at 0x%08X", address)
}
```

**Before:**
```go
c.N = (value & (1 << 31)) != 0
c.Z = (value & (1 << 30)) != 0
c.C = (value & (1 << 29)) != 0
c.V = (value & (1 << 28)) != 0
```

**After:**
```go
c.N = (value & (1 << CPSRBitN)) != 0
c.Z = (value & (1 << CPSRBitZ)) != 0
c.C = (value & (1 << CPSRBitC)) != 0
c.V = (value & (1 << CPSRBitV)) != 0
```

### Step 4: Verification
- Run full test suite after each phase
- Use `golangci-lint` to catch unused constants
- Review git diff to ensure no behavioral changes

---

## Benefits of Rationalization

### Immediate Benefits
1. **Improved Readability**: `CPSRBitN` is clearer than `31`
2. **Self-Documenting Code**: Constants explain their purpose
3. **Centralized Maintenance**: Change limits in one place
4. **Error Prevention**: Compiler catches typos in constant names (though not type mismatches)

### Long-Term Benefits
1. **Easier Architecture Changes**: Update constants, not scattered literals
2. **Better Tooling Support**: IDEs can navigate to constant definitions
3. **Reduced Bugs**: Less chance of using wrong magic number
4. **Compliance**: Follows Go best practices (avoid magic numbers)

---

## Excluded Numbers (Not Magic)

Some numbers are **not considered magic** and should remain as literals:

1. **Zero values**: `0` for initialization or null checks
2. **Boolean values**: `0` and `1` in clear boolean contexts
3. **Array indices**: `0`, `1`, `2` when iterating
4. **Small counting numbers**: `1`, `2`, `3` in clear arithmetic contexts
5. **Powers of 2**: `2`, `4`, `8`, `16` in very obvious contexts (though masks should still be constants)

---

## Estimated Effort

- **Phase 1-2 (Create constants)**: 2-3 hours
- **Phase 3 (VM constants)**: 1 hour
- **Phase 4 (Opcode constants)**: 1 hour
- **Phase 5 (Migration)**: 10-15 hours (gradual, file-by-file)
- **Testing and verification**: 3-4 hours

**Total:** ~20-25 hours of focused development work

---

## Recommendations

1. **Start with Phase 1-2**: Create architecture and syscall constants
2. **Migrate vm/cpu.go first**: High-impact, clear improvements
3. **Migrate vm/syscall.go second**: Many duplicated error codes
4. **Encoder files last**: Complex bit manipulation, need careful testing
5. **Use automated testing**: Run full test suite after each file migration
6. **Code review each phase**: Ensure constants make sense

---

## Conclusion

The ARM emulator codebase contains **hundreds of magic numbers** across all subsystems. While some areas (syscall numbers, memory segments) already use constants, many operational details (bit positions, limits, error codes) use raw literals.

A phased rationalization approach will significantly improve code quality without disrupting functionality. The proposed constant packages provide a clear, maintainable foundation for the codebase.

**Next Steps:**
1. Review this report
2. Approve the proposed constant organization
3. Begin Phase 1 implementation
4. Establish testing protocol for each phase

---

## Implementation Review (Post-Merge)

**Date:** November 1, 2025  
**Status:** Partial Implementation

### What Was Actually Implemented

The implementation completed phases 1-4 and partially migrated 5 files in phase 5:

#### ✅ Successfully Implemented
1. **Created constant files:**
   - `vm/arch_constants.go` (56 lines) - ARM2 architecture constants
   - `vm/syscall_constants.go` (55 lines) - Syscall operation constants
   - `vm/vm_constants.go` (14 lines) - VM configuration constants
   - `encoder/constants.go` (enhanced) - Instruction encoding constants

2. **Files successfully migrated:**
   - `vm/cpu.go` - CPSR bit positions, register counts, pipeline offset
   - `vm/memory.go` - Alignment masks (partial)
   - `vm/syscall.go` - Standard FDs, syscall mask, size limits
   - `vm/executor.go` - VM configuration defaults
   - `encoder/data_processing.go`, `encoder/memory.go`, `encoder/other.go` - Opcode constants

3. **Code quality improvements:**
   - CPSR flags: `31/30/29/28` → `CPSRBitN/Z/C/V` ✅
   - Alignment masks: `0x3/0x1` → `AlignMaskWord/Halfword` ✅
   - Standard FDs: `0/1/2/3` → `StdIn/Out/Err/FirstUserFD` ✅
   - Pipeline offset: `8` → `ARMPipelineOffset` ✅
   - Syscall mask: `0x00FFFFFF` → `SWIMask` ✅
   - Size limits: raw literals → named constants ✅

### ❌ Known Issues and Gaps

#### 1. **Missing Byte Shift Constants**
**Severity:** Medium  
**Location:** `vm/memory.go` lines 188, 190, 254-261

The PROGRESS.md claims "Byte shifts: `8/16/24` → `ByteShift8/16/24`" but these constants were never created. The literals remain:

```go
// Lines 254-256 - still using magic numbers
value = uint32(seg.Data[offset]) |
    uint32(seg.Data[offset+1])<<8 |
    uint32(seg.Data[offset+2])<<16 |
    uint32(seg.Data[offset+3])<<24
```

**Should be:** `ByteShift8`, `ByteShift16`, `ByteShift24` constants in `arch_constants.go`

#### 2. **Incomplete Migration Scope**
**Severity:** High  
**Impact:** Documentation/implementation mismatch

The MAGIC_NUMBERS.md document analyzes **122 Go files** across the entire codebase, but the implementation only touched **5 files**. Many files still contain magic numbers:

- `vm/inst_memory.go` - shift positions, alignment checks
- `vm/data_processing.go` - bit manipulation constants
- `vm/multiply.go` - bit masks and positions
- `vm/branch.go` - offset calculations
- `vm/memory_multi.go` - register masks
- `vm/psr.go` - CPSR bit operations
- `parser/*.go` - parsing constants
- `debugger/*.go` - display formatting constants
- `gui/*.go` - UI layout constants

**Actual coverage:** ~4% of files (5 out of 122)  
**Claimed coverage:** "throughout codebase"

#### 3. **Questionable Design Choices**

**a) Overly Obvious Constants**
```go
Address32BitMax = 0xFFFFFFFF  // Just ^uint32(0)
Mask32Bit = 0xFFFFFFFF         // Same as above, duplicate purpose
```
These don't add clarity and could confuse.

**b) Redundant Inverse Masks**
```go
AlignMaskWord = 0x3            // Test: address & mask == 0
AlignRoundUpMaskWord = 0xFFFFFFFC  // Round: address & mask
```
Computing `^AlignMaskWord` at runtime would be clearer than having both.

**c) Legacy Aliases**
The encoder constants kept legacy aliases "for backward compatibility" in an internal codebase that could just be updated in one pass.

#### 4. **Organizational Fragmentation**
**Severity:** Low  
**Impact:** Developer experience

Three separate constant files in `vm/` package feels fragmented:
- `arch_constants.go`
- `syscall_constants.go`
- `vm_constants.go`

Could consolidate into `vm/constants.go` with clear subsections, or keep separate but ensure better thematic boundaries (some constants feel arbitrarily split).

### 📊 Impact Assessment

**Positive impacts:**
- The 5 files touched are significantly more readable
- CPSR operations are self-documenting
- Syscall code is clearer with named FDs and limits
- Encoder constants are comprehensive and well-documented

**Negative impacts:**
- Documentation oversells implementation scope
- PROGRESS.md claims features that don't exist (byte shift constants)
- Future developers may be confused about partial implementation
- Technical debt remains in 95%+ of codebase

### 🎯 Recommendations

#### Short-term (Next PR)
1. **Fix byte shift constants** - Add `ByteShift8/16/24` to `arch_constants.go` and migrate `memory.go`
2. **Update PROGRESS.md** - Accurately describe what was implemented (don't claim byte shifts)
3. **Add implementation status section** - Be transparent about partial coverage

#### Medium-term
4. **Continue migration systematically:**
   - Phase 5a: `vm/inst_memory.go`, `vm/data_processing.go`, `vm/multiply.go`
   - Phase 5b: `vm/branch.go`, `vm/memory_multi.go`, `vm/psr.go`
   - Phase 5c: Parser and debugger packages
5. **Remove questionable constants** - Drop `Address32BitMax`, consolidate masks
6. **Consolidate constant files** - Consider single `vm/constants.go`

#### Long-term
7. **Complete the original vision** - Migrate all 122 files over time
8. **Establish constant policy** - Document when to create constants vs use literals
9. **Add linter rules** - Detect new magic numbers in code review

### ✅ Overall Assessment

**Status:** Partial success with documentation/implementation mismatch

The work completed is high quality for the files touched. However, the PR oversells its scope - claiming "replaced magic numbers throughout codebase" when only ~5 files were migrated. The byte shift constants documented in PROGRESS.md were never implemented.

This creates technical debt where future developers will:
- Expect features that don't exist (byte shift constants)
- Be confused about migration status
- Wonder why only some files use constants

**Recommendation:** Accept the PR for the value it delivers, but immediately follow up with:
1. Fix documentation to match reality
2. Add implementation status tracking
3. Create roadmap for completing remaining 95% of codebase

---

**Report prepared by:** Claude Code  
**Related Issue:** #37  
**Review Date:** November 1, 2025
