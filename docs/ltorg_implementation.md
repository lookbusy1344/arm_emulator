# .ltorg Directive Implementation Summary

## Overview

Implemented the `.ltorg` (literal pool organization) directive to solve the literal pool addressing range limitation in ARM2 programs, particularly those using low memory origins like `.org 0x0000`.

## Problem Statement

**Issue:** ARM's PC-relative addressing has a ±4095 byte range (12-bit offset). When programs use `.org 0x0000` or other low addresses, and have many `LDR Rd, =constant` pseudo-instructions, the default behavior of placing all literals at the end of the program can cause "literal pool offset too large" errors.

**Example Scenario:**
```
Code starts at: 0x0000
Code ends at:   0x1000
Data ends at:   0x2000
Literals at:    0x2000+

Instruction at PC=0x0008 needs literal at 0x2000
Offset = 0x2000 - 0x0008 - 8 = 8184 bytes
Error: Exceeds 4095 byte maximum!
```

## Solution

The `.ltorg` directive allows programmers to manually place literal pools at strategic locations within the ±4095 byte range of instructions that need them.

## Implementation Details

### 1. Parser Changes (`parser/parser.go`)

**Added to `Program` struct:**
```go
LiteralPoolLocs []uint32 // Addresses where .ltorg directives appear
```

**Added directive handler:**
- Recognizes `.ltorg` directive
- Aligns to 4-byte boundary
- Records location in `Program.LiteralPoolLocs`
- Space reservation happens during encoding (size unknown at parse time)

### 2. Encoder Changes (`encoder/encoder.go`, `encoder/memory.go`)

**Added to `Encoder` struct:**
```go
LiteralPoolLocs  []uint32          // Addresses of .ltorg directives
pendingLiterals  map[uint32]uint32 // value -> address for dedup
```

**New methods:**
- `findNearestLiteralPoolLocation(pc, value uint32) uint32`
  - Finds closest `.ltorg` location within ±4095 bytes
  - Considers how many literals already allocated at each pool
  - Returns 0 if no suitable location (falls back to old behavior)
  
- `countLiteralsAtPool(poolLoc uint32) int`
  - Counts literals already assigned near a pool location
  - Used to calculate where next literal would go

**Modified `encodeLDRPseudo()`:**
- Checks for existing literal (deduplication)
- Calls `findNearestLiteralPoolLocation()` if `.ltorg` exists
- Falls back to `LiteralPoolStart` if no `.ltorg` directives
- Validates offset is within ±4095 bytes

### 3. Loader Changes (`main.go`)

**Program loading:**
- Copies `program.LiteralPoolLocs` to encoder
- `.ltorg` directive handled in directive processing (continues without action)
- Literals written to memory after all instructions encoded

## Features

✅ **Multiple Pools:** Support multiple `.ltorg` directives for large programs  
✅ **Automatic Alignment:** Pools aligned to 4-byte boundaries  
✅ **Deduplication:** Same constant value shared across pools  
✅ **Smart Selection:** Chooses nearest pool within addressing range  
✅ **Backward Compatible:** Falls back to old behavior if no `.ltorg`  
✅ **Low Memory Support:** Works with `.org 0x0000` and other low origins  

## Usage Example

```asm
.org 0x0000

main:
    LDR R0, =0x12345678    ; Large constant
    LDR R1, =0xDEADBEEF    ; Another constant
    ADD R2, R0, R1
    B   section2
    
    .ltorg                 ; Pool #1: close to loads above

section2:
    ; Code far from main
    LDR R3, =0x11111111
    LDR R4, =0x22222222
    
    .ltorg                 ; Pool #2: close to distant loads
    
    MOV R0, #0
    SWI #0x00
```

## Testing

**Integration Tests (`tests/integration/ltorg_test.go`):**
1. `TestLtorgDirective_Basic` - Single `.ltorg` directive
2. `TestLtorgDirective_MultiplePools` - Multiple pools
3. `TestLtorgDirective_LowMemoryOrigin` - `.org 0x0000` with many constants
4. `TestLtorgDirective_Alignment` - 4-byte alignment verification
5. `TestLtorgDirective_NoLtorg` - Fallback behavior

**Example Programs:**
- `examples/test_ltorg.s` - Demonstrates multiple pools
- `examples/test_org_0_with_ltorg.s` - Low memory origin

**All tests passing:** ✅

## Documentation

- **Assembly Reference:** `docs/assembly_reference.md` updated with `.ltorg` section
- **Examples README:** `examples/README.md` updated to mention `.ltorg`
- **Progress Log:** `PROGRESS.md` entry created (2025-10-15)
- **TODO:** High-priority item removed from `TODO.md`

## Algorithm: Finding Nearest Pool

```
For each .ltorg location:
  1. Calculate distance from PC to pool location
  2. If distance > 4095 bytes: skip
  3. Count existing literals at this pool
  4. Calculate where new literal would go
  5. Verify new literal still within ±4095 bytes from PC
  6. Track as best candidate if closer than previous best
Return best candidate (or 0 if none found)
```

## Performance Impact

- **Negligible:** Pool location search is O(n) where n = number of `.ltorg` directives
- Typical programs have 0-3 pools
- No impact on programs without `.ltorg`

## Compatibility

- **Standard ARM:** `.ltorg` is a standard ARM assembler directive
- **GNU AS:** Compatible with GNU assembler syntax
- **Backward Compatible:** Programs without `.ltorg` work as before
- **No Breaking Changes:** All existing tests pass

## Files Modified

1. `parser/parser.go` - Parse `.ltorg`, track locations
2. `encoder/encoder.go` - Add pool tracking fields
3. `encoder/memory.go` - Add pool selection logic
4. `main.go` - Copy pool locations to encoder
5. `docs/assembly_reference.md` - Document `.ltorg`
6. `examples/README.md` - Note `.ltorg` availability

## Files Added

1. `tests/integration/ltorg_test.go` - 5 integration tests
2. `examples/test_ltorg.s` - Demonstration program
3. `examples/test_org_0_with_ltorg.s` - Low memory test

## Verification

```bash
# Build
go build -o arm-emulator

# Run all tests
go test ./...
# Result: All tests pass ✅

# Test example program
./arm-emulator examples/test_ltorg.s
# Output: -1142894555, 1717986918 ✅

# Test low memory origin
./arm-emulator examples/test_org_0_with_ltorg.s
# Output: -1160852749 ✅
```

## Future Enhancements

Possible improvements (not required for v1.0):
- Automatic `.ltorg` insertion when offset would exceed range
- Warning if literal pool might be unreachable
- Statistics on literal pool usage
- Pool size optimization

## Conclusion

The `.ltorg` directive implementation successfully solves the literal pool addressing limitation, enabling programs with low memory origins to use multiple constants without exceeding the ±4095 byte addressing range. The solution is standard-compliant, fully tested, and backward compatible.
