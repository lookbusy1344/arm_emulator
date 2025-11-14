# Stack Bounds Validation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add proactive stack pointer bounds validation to prevent stack overflow/underflow corruption.

**Architecture:** Modify `SetSP()` and `SetSPWithTrace()` in `vm/cpu.go` to validate SP stays within stack segment bounds (0x00040000 - 0x00050000). Return errors for violations following the VM's integrity error pattern. Update all call sites to handle errors and propagate them appropriately. Update ~40 test files to use valid stack addresses.

**Tech Stack:** Go 1.25, ARM2 emulator, TDD with table-driven tests

**Current State:**
- Stack segment: 0x00040000 - 0x00050000 (64KB, grows downward)
- `SetSP(value uint32)` - No validation, simple assignment
- `SetSPWithTrace(vm, value, pc)` - No validation, calls stack trace recorder
- Main call sites: `InitializeStack()`, `Bootstrap()`, `Reset()`

**Error Handling Strategy:**
Following VM's two-tier error philosophy (`vm/syscall.go:16-34`), stack bounds violations are **Tier 1: VM Integrity Errors** that halt execution immediately by returning `fmt.Errorf()`.

**ARCHITECTURAL DECISION (Made During Implementation):**
During Task 6 implementation, the stack validation boundary semantics were corrected from exclusive to inclusive upper bound. The original plan specified `value >= StackSegmentStart+StackSegmentSize` (exclusive), but ARM Full Descending stack convention requires `value > StackSegmentStart+StackSegmentSize` (inclusive). This allows SP to point to the empty stack position (one word above the allocated segment), which is standard ARM behavior. Empty stack SP never gets dereferenced - STMFD pre-decrements before storing. Memory access layer still protects against invalid access at the boundary address.

**Valid SP Range:** `[0x00040000, 0x00050000]` inclusive (not exclusive as originally planned)

---

## Task 1: Write Failing Tests for SetSP Bounds Validation

**Files:**
- Create: `tests/unit/vm/cpu_stack_bounds_test.go`

**Step 1: Create test file with bounds validation tests**

Create `tests/unit/vm/cpu_stack_bounds_test.go`:

```go
package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm_emulator/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCPU_SetSP_ValidRange(t *testing.T) {
	cpu := vm.NewCPU()

	tests := []struct {
		name  string
		value uint32
	}{
		{"stack start (minimum)", vm.StackSegmentStart},
		{"stack middle", vm.StackSegmentStart + vm.StackSegmentSize/2},
		{"stack end (maximum)", vm.StackSegmentStart + vm.StackSegmentSize},
		{"stack end minus 4", vm.StackSegmentStart + vm.StackSegmentSize - 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cpu.SetSP(tt.value)
			assert.NoError(t, err, "Valid SP value should not error")
			assert.Equal(t, tt.value, cpu.GetSP(), "SP should be set to requested value")
		})
	}
}

func TestCPU_SetSP_Underflow(t *testing.T) {
	cpu := vm.NewCPU()

	tests := []struct {
		name  string
		value uint32
	}{
		{"one below minimum", vm.StackSegmentStart - 1},
		{"far below minimum", vm.StackSegmentStart - 0x1000},
		{"zero address", 0x00000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cpu.SetSP(tt.value)
			require.Error(t, err, "SP below stack segment should error")
			assert.Contains(t, err.Error(), "stack underflow", "Error should mention underflow")
			assert.Contains(t, err.Error(), "0x00040000", "Error should show stack minimum")
		})
	}
}

func TestCPU_SetSP_Overflow(t *testing.T) {
	cpu := vm.NewCPU()

	tests := []struct {
		name  string
		value uint32
	}{
		{"one above maximum", vm.StackSegmentStart + vm.StackSegmentSize + 1},
		{"far above maximum", vm.StackSegmentStart + vm.StackSegmentSize + 0x1000},
		{"max address", 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cpu.SetSP(tt.value)
			require.Error(t, err, "SP above stack segment should error")
			assert.Contains(t, err.Error(), "stack overflow", "Error should mention overflow")
			assert.Contains(t, err.Error(), "0x00050000", "Error should show stack maximum")
		})
	}
}
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./tests/unit/vm/cpu_stack_bounds_test.go -v
```

Expected output:
```
# compilation error: cpu.SetSP signature is SetSP(uint32), not SetSP(uint32) error
```

This is expected - the function doesn't return an error yet.

**Step 3: Commit the failing tests**

```bash
git add tests/unit/vm/cpu_stack_bounds_test.go
git commit -m "test: add failing tests for SetSP bounds validation"
```

---

## Task 2: Update SetSP Function Signature and Implementation

**Files:**
- Modify: `vm/cpu.go:108-110`

**Step 1: Update SetSP to validate bounds and return error**

In `vm/cpu.go`, replace the existing `SetSP` function (lines 108-110):

```go
// SetSP sets the stack pointer (R13) with bounds validation.
// Returns error if newSP is outside the valid stack segment range.
func (c *CPU) SetSP(value uint32) error {
	// Validate SP is within stack segment bounds
	// Stack segment: 0x00040000 - 0x00050000 (64KB, grows downward)
	// Uses exclusive upper bound: [Start, Start+Size)
	if value < StackSegmentStart {
		return fmt.Errorf("stack underflow: SP=0x%08X is below stack minimum (0x%08X)",
			value, StackSegmentStart)
	}
	if value >= StackSegmentStart+StackSegmentSize {
		return fmt.Errorf("stack overflow: SP=0x%08X exceeds stack maximum (0x%08X)",
			value, StackSegmentStart+StackSegmentSize)
	}

	c.R[SP] = value
	return nil
}
```

**Step 2: Run unit tests to verify SetSP validation works**

Run:
```bash
go test ./tests/unit/vm/cpu_stack_bounds_test.go -v
```

Expected output:
```
=== RUN   TestCPU_SetSP_ValidRange
    --- PASS: TestCPU_SetSP_ValidRange (0.00s)
=== RUN   TestCPU_SetSP_Underflow
    --- PASS: TestCPU_SetSP_Underflow (0.00s)
=== RUN   TestCPU_SetSP_Overflow
    --- PASS: TestCPU_SetSP_Overflow (0.00s)
PASS
```

**Step 3: Commit SetSP implementation**

```bash
git add vm/cpu.go
git commit -m "feat: add bounds validation to SetSP function

- Validate SP stays within stack segment (0x00040000-0x00050000)
- Return descriptive errors for underflow/overflow
- Follows VM integrity error pattern"
```

---

## Task 3: Write Failing Tests for SetSPWithTrace Bounds Validation

**Files:**
- Modify: `tests/unit/vm/cpu_stack_bounds_test.go`

**Step 1: Add SetSPWithTrace test cases**

Append to `tests/unit/vm/cpu_stack_bounds_test.go`:

```go
func TestCPU_SetSPWithTrace_ValidRange(t *testing.T) {
	v := vm.NewVM()
	v.StackTrace = vm.NewStackTrace(nil, vm.StackSegmentStart+vm.StackSegmentSize, vm.StackSegmentStart)
	pc := uint32(0x00008000)

	tests := []struct {
		name  string
		value uint32
	}{
		{"stack start", vm.StackSegmentStart},
		{"stack middle", vm.StackSegmentStart + vm.StackSegmentSize/2},
		{"stack end", vm.StackSegmentStart + vm.StackSegmentSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.CPU.SetSPWithTrace(v, tt.value, pc)
			assert.NoError(t, err, "Valid SP value should not error")
			assert.Equal(t, tt.value, v.CPU.GetSP(), "SP should be set to requested value")
		})
	}
}

func TestCPU_SetSPWithTrace_Underflow(t *testing.T) {
	v := vm.NewVM()
	v.StackTrace = vm.NewStackTrace(nil, vm.StackSegmentStart+vm.StackSegmentSize, vm.StackSegmentStart)
	pc := uint32(0x00008000)

	tests := []struct {
		name  string
		value uint32
	}{
		{"one below minimum", vm.StackSegmentStart - 1},
		{"far below minimum", vm.StackSegmentStart - 0x1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.CPU.SetSPWithTrace(v, tt.value, pc)
			require.Error(t, err, "SP below stack segment should error")
			assert.Contains(t, err.Error(), "stack underflow")
		})
	}
}

func TestCPU_SetSPWithTrace_Overflow(t *testing.T) {
	v := vm.NewVM()
	v.StackTrace = vm.NewStackTrace(nil, vm.StackSegmentStart+vm.StackSegmentSize, vm.StackSegmentStart)
	pc := uint32(0x00008000)

	tests := []struct {
		name  string
		value uint32
	}{
		{"one above maximum", vm.StackSegmentStart + vm.StackSegmentSize + 1},
		{"far above maximum", vm.StackSegmentStart + vm.StackSegmentSize + 0x1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.CPU.SetSPWithTrace(v, tt.value, pc)
			require.Error(t, err, "SP above stack segment should error")
			assert.Contains(t, err.Error(), "stack overflow")
		})
	}
}
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./tests/unit/vm/cpu_stack_bounds_test.go -v
```

Expected output:
```
# compilation error: SetSPWithTrace signature mismatch
```

**Step 3: Commit the failing tests**

```bash
git add tests/unit/vm/cpu_stack_bounds_test.go
git commit -m "test: add failing tests for SetSPWithTrace bounds validation"
```

---

## Task 4: Update SetSPWithTrace Function

**Files:**
- Modify: `vm/cpu.go:113-121`

**Step 1: Update SetSPWithTrace to validate bounds**

In `vm/cpu.go`, replace the existing `SetSPWithTrace` function (lines 113-121):

```go
// SetSPWithTrace sets the stack pointer with tracing support and bounds validation.
// Returns error if newSP is outside the valid stack segment range.
func (c *CPU) SetSPWithTrace(vm *VM, value uint32, pc uint32) error {
	// Validate bounds before modifying SP
	// Uses exclusive upper bound: [Start, Start+Size)
	if value < StackSegmentStart {
		return fmt.Errorf("stack underflow: SP=0x%08X is below stack minimum (0x%08X)",
			value, StackSegmentStart)
	}
	if value >= StackSegmentStart+StackSegmentSize {
		return fmt.Errorf("stack overflow: SP=0x%08X exceeds stack maximum (0x%08X)",
			value, StackSegmentStart+StackSegmentSize)
	}

	oldSP := c.R[SP]
	c.R[SP] = value

	// Record stack trace if enabled
	if vm.StackTrace != nil {
		vm.StackTrace.RecordSPMove(vm.CPU.Cycles, pc, oldSP, value)
	}

	return nil
}
```

**Step 2: Run tests to verify SetSPWithTrace validation works**

Run:
```bash
go test ./tests/unit/vm/cpu_stack_bounds_test.go -v
```

Expected output: All tests passing (6 test functions)

**Step 3: Run full VM test suite to see cascading failures**

Run:
```bash
go test ./tests/unit/vm/ -v 2>&1 | head -50
```

Expected output: Many compilation errors for call sites that don't handle errors yet.

**Step 4: Commit SetSPWithTrace implementation**

```bash
git add vm/cpu.go
git commit -m "feat: add bounds validation to SetSPWithTrace

- Validate SP stays within stack segment before modification
- Return descriptive errors for underflow/overflow
- Preserve stack trace recording after validation"
```

---

## Task 5: Update InitializeStack Call Site

**Files:**
- Modify: `vm/executor.go:196-205`

**Step 1: Update InitializeStack to return and handle errors**

In `vm/executor.go`, find the `InitializeStack` function (around line 196-205) and update it:

```go
// InitializeStack sets up the stack pointer at the provided top address.
// Returns error if stackTop is outside valid stack bounds.
func (vm *VM) InitializeStack(stackTop uint32) error {
	vm.StackTop = stackTop
	if err := vm.CPU.SetSP(stackTop); err != nil {
		return fmt.Errorf("failed to initialize stack: %w", err)
	}
	return nil
}
```

**Step 2: Write test for InitializeStack error handling**

Create `tests/unit/vm/executor_stack_init_test.go`:

```go
package vm_test

import (
	"testing"

	"github.com/lookbusy1344/arm_emulator/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVM_InitializeStack_ValidAddress(t *testing.T) {
	v := vm.NewVM()

	validStackTop := vm.StackSegmentStart + vm.StackSegmentSize
	err := v.InitializeStack(validStackTop)

	assert.NoError(t, err)
	assert.Equal(t, validStackTop, v.CPU.GetSP())
	assert.Equal(t, validStackTop, v.StackTop)
}

func TestVM_InitializeStack_InvalidAddress(t *testing.T) {
	v := vm.NewVM()

	tests := []struct {
		name      string
		stackTop  uint32
		expectErr string
	}{
		{"underflow", vm.StackSegmentStart - 1, "stack underflow"},
		{"overflow", vm.StackSegmentStart + vm.StackSegmentSize + 1, "stack overflow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.InitializeStack(tt.stackTop)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectErr)
			assert.Contains(t, err.Error(), "failed to initialize stack")
		})
	}
}
```

**Step 3: Run test to verify**

Run:
```bash
go test ./tests/unit/vm/executor_stack_init_test.go -v
```

Expected: All tests pass

**Step 4: Commit InitializeStack changes**

```bash
git add vm/executor.go tests/unit/vm/executor_stack_init_test.go
git commit -m "feat: add error handling to InitializeStack

- Return error if stackTop is outside valid bounds
- Wrap underlying SetSP errors with context"
```

---

## Task 6: Update Bootstrap Call Site

**Files:**
- Modify: `vm/executor.go:476-485`

**Step 1: Update Bootstrap to handle InitializeStack errors**

In `vm/executor.go`, find the `Bootstrap` function (around line 476-485). Update the stack initialization section:

```go
// Initialize stack (around line 483-484)
stackTop := StackSegmentStart + StackSegmentSize
if err := vm.InitializeStack(stackTop); err != nil {
	return fmt.Errorf("failed to bootstrap VM: %w", err)
}
```

**Step 2: Verify Bootstrap signature already returns error**

Check that `Bootstrap` already has signature `func (vm *VM) Bootstrap(...) error`. If not, update it.

**Step 3: Run Bootstrap tests**

Run:
```bash
go test ./tests/unit/vm/ -run TestVM_Bootstrap -v
```

Expected: Tests should still pass (Bootstrap calculates valid stackTop)

**Step 4: Commit Bootstrap changes**

```bash
git add vm/executor.go
git commit -m "feat: propagate stack initialization errors in Bootstrap

- Handle InitializeStack errors and wrap with context
- Return error to caller for proper error propagation"
```

---

## Task 7: Update Reset Call Site

**Files:**
- Modify: `vm/executor.go:174-180`

**Step 1: Update Reset to handle SetSP errors**

In `vm/executor.go`, find the `Reset` function (around line 174-180). Update the SP restoration:

```go
// Reset the stack pointer (around line 177)
if err := vm.CPU.SetSP(vm.StackTop); err != nil {
	// Should never happen - StackTop was validated during Bootstrap
	// But handle defensively
	return fmt.Errorf("failed to reset stack pointer: %w", err)
}
```

**Step 2: Verify Reset signature returns error**

Check that `Reset` has signature `func (vm *VM) Reset() error`. If it currently returns nothing, update:

```go
func (vm *VM) Reset() error {
	// ... existing code ...

	if err := vm.CPU.SetSP(vm.StackTop); err != nil {
		return fmt.Errorf("failed to reset stack pointer: %w", err)
	}

	// ... existing code ...
	return nil
}
```

**Step 3: Update Reset call sites**

Search for calls to `Reset()` and update them to handle errors:

```bash
grep -rn "\.Reset()" vm/ --include="*.go"
```

For each call site, add error handling:
```go
if err := vm.Reset(); err != nil {
	return err  // or handle appropriately
}
```

**Step 4: Run VM tests**

Run:
```bash
go test ./tests/unit/vm/ -v
```

**Step 5: Commit Reset changes**

```bash
git add vm/executor.go
git commit -m "feat: add error handling to Reset function

- Propagate SetSP errors during stack pointer restoration
- Update function signature to return error
- Update call sites to handle errors"
```

---

## Task 8: Find and Update Remaining VM Call Sites

**Files:**
- Various files in `vm/` directory

**Step 1: Find all remaining SetSP/SetSPWithTrace call sites in VM**

Run:
```bash
grep -rn "\.SetSP\|\.SetSPWithTrace" vm/ --include="*.go" | grep -v "// " | grep -v "test"
```

**Step 2: For each call site, add error handling**

Pattern to follow:
```go
// Before:
vm.CPU.SetSP(value)

// After:
if err := vm.CPU.SetSP(value); err != nil {
	vm.State = StateError
	vm.LastError = err
	return err
}
```

Or for instruction execution (if inside `executeInstruction`):
```go
if err := vm.CPU.SetSPWithTrace(vm, newSP, pc); err != nil {
	vm.State = StateError
	vm.LastError = err
	return err
}
```

**Step 3: Build to check for remaining compilation errors**

Run:
```bash
go build -o arm-emulator
```

**Step 4: Fix any remaining compilation errors**

Continue pattern until build succeeds.

**Step 5: Run VM unit tests**

Run:
```bash
go test ./tests/unit/vm/ -v
```

**Step 6: Commit production code changes**

```bash
git add vm/
git commit -m "fix: update all VM call sites for SetSP error handling

- Add error handling to all SetSP/SetSPWithTrace calls in vm/ package
- Propagate errors appropriately based on context
- Set VM.State = StateError on stack bounds violations"
```

---

## Task 9: Update Instruction Implementation Call Sites

**Files:**
- `instructions/memory.go` (LDMIA, STMIA, LDM, STM with SP base)
- `instructions/data_processing.go` (MOV/ADD/SUB with SP as destination)
- Any other instruction files that modify SP

**Step 1: Find instruction implementations that modify SP**

Run:
```bash
grep -rn "SetSPWithTrace\|SetReg.*SP\|SetRegister.*13" instructions/ --include="*.go"
```

**Step 2: Update each instruction to handle SetSPWithTrace errors**

Common pattern for data processing instructions:
```go
// In executeDataProcessing or similar
if destReg == vm.SP {
	if err := vm.CPU.SetSPWithTrace(vm, result, pc); err != nil {
		vm.State = vm.StateError
		vm.LastError = err
		return err
	}
} else {
	vm.CPU.SetRegister(destReg, result)
}
```

**Step 3: Build and test instructions**

Run:
```bash
go build -o arm-emulator
go test ./tests/unit/instructions/ -v
```

**Step 4: Commit instruction changes**

```bash
git add instructions/
git commit -m "fix: add error handling for SP modifications in instructions

- Update data processing instructions to handle SetSPWithTrace errors
- Update memory instructions (LDM/STM) to handle SP updates
- Propagate errors to VM state"
```

---

## Task 10: Fix Test Files - Batch 1 (tests/unit/vm/)

**Files:**
- All test files in `tests/unit/vm/` that use SetSP with invalid addresses

**Step 1: Find test files using SetSP**

Run:
```bash
grep -rn "\.SetSP\|\.SetSPWithTrace" tests/unit/vm/ --include="*_test.go" -l
```

**Step 2: For each test file, update to use valid stack addresses**

Common patterns:

**Pattern 1: Tests using arbitrary addresses (e.g., 0x1000, 0x5000)**
```go
// Before:
cpu.SetSP(0x1000)

// After: Use valid stack address
cpu.SetSP(vm.StackSegmentStart + 0x1000)  // 0x00041000
```

**Pattern 2: Tests expecting SetSP to succeed**
```go
// Before:
cpu.SetSP(address)

// After:
err := cpu.SetSP(address)
require.NoError(t, err)
```

**Pattern 3: Tests that don't care about SP value**
```go
// Before:
cpu.SetSP(0x2000)  // arbitrary

// After: Use valid default
cpu.SetSP(vm.StackSegmentStart + vm.StackSegmentSize)  // stack top
```

**Step 3: Run tests incrementally**

After updating each file:
```bash
go test ./tests/unit/vm/<filename>_test.go -v
```

**Step 4: Commit batch 1 changes**

```bash
git add tests/unit/vm/
git commit -m "test: update vm/ tests to use valid stack addresses (batch 1)

- Replace arbitrary SP values with valid stack segment addresses
- Add error handling to SetSP/SetSPWithTrace calls
- Ensure tests use addresses in range [0x00040000, 0x00050000]"
```

---

## Task 11: Fix Test Files - Batch 2 (tests/unit/instructions/)

**Files:**
- All test files in `tests/unit/instructions/` that use SetSP

**Step 1: Find test files using SetSP**

Run:
```bash
grep -rn "\.SetSP" tests/unit/instructions/ --include="*_test.go" -l
```

**Step 2: Update each file with valid stack addresses**

Same patterns as Task 10.

**Step 3: Run tests**

```bash
go test ./tests/unit/instructions/ -v
```

**Step 4: Commit batch 2 changes**

```bash
git add tests/unit/instructions/
git commit -m "test: update instructions/ tests to use valid stack addresses (batch 2)

- Ensure all instruction tests use valid SP values
- Add error handling where needed"
```

---

## Task 12: Fix Test Files - Batch 3 (tests/integration/)

**Files:**
- Integration test files that might use SetSP

**Step 1: Check integration tests**

Run:
```bash
grep -rn "\.SetSP" tests/integration/ --include="*_test.go" -l
```

**Step 2: Update if needed**

Most integration tests use Bootstrap which calculates valid stackTop, so may not need changes.

**Step 3: Run integration tests**

```bash
go test ./tests/integration/ -v
```

**Step 4: Commit if changes were needed**

```bash
git add tests/integration/
git commit -m "test: update integration tests for stack bounds validation (batch 3)"
```

---

## Task 13: Update Stack Trace Tests

**Files:**
- `tests/unit/vm/stack_trace_test.go`

**Step 1: Review stack trace test patterns**

Stack trace tests specifically test overflow/underflow detection. These tests intentionally use invalid SP values to verify StackTrace.RecordSPMove detects violations.

**Step 2: Add comment explaining intentional violations**

For tests that use invalid SP values intentionally:

```go
// TestStackTrace_OverflowDetection verifies stack trace can detect violations
// even though SetSPWithTrace now validates proactively
func TestStackTrace_OverflowDetection(t *testing.T) {
	// This test uses intentionally invalid SP values to verify
	// the stack trace's detection logic still works

	// ... test code ...
}
```

**Step 3: Update test setup to bypass validation if needed**

If tests need to set invalid SP for testing StackTrace detection, consider:

Option A: Use direct register assignment for test purposes:
```go
// Bypass validation for testing overflow detection
vm.CPU.R[vm.SP] = invalidAddress
vm.StackTrace.RecordSPMove(cycle, pc, oldSP, invalidAddress)
```

Option B: Keep tests as-is if they're testing the StackTrace detection logic independently.

**Step 4: Run stack trace tests**

```bash
go test ./tests/unit/vm/stack_trace_test.go -v
```

**Step 5: Commit changes**

```bash
git add tests/unit/vm/stack_trace_test.go
git commit -m "test: update stack trace tests for bounds validation

- Add comments explaining intentional invalid SP usage
- Ensure tests still verify StackTrace detection logic"
```

---

## Task 14: Run Full Test Suite and Fix Remaining Issues

**Files:**
- Various test files that may still have issues

**Step 1: Clear test cache and run full test suite**

```bash
go clean -testcache
go test ./... -v 2>&1 | tee test_results.txt
```

**Step 2: Review failures and categorize**

Look for patterns in failures:
- Compilation errors (missed call sites)
- Test failures (using invalid SP values)
- Unexpected behavior

**Step 3: Fix remaining issues one by one**

For each issue:
1. Identify the root cause
2. Apply appropriate fix (error handling or valid address)
3. Run affected test to verify fix
4. Commit the fix

**Step 4: Run full suite again**

```bash
go clean -testcache
go test ./...
```

Expected: All tests pass

**Step 5: Commit any remaining fixes**

```bash
git add <files>
git commit -m "fix: resolve remaining test issues for stack bounds validation

- Fix <specific issues>
- All 1,024+ tests now passing"
```

---

## Task 15: Update Documentation

**Files:**
- Modify: `CODE_REVIEW.md`
- Modify: `PROGRESS.md`

**Step 1: Update CODE_REVIEW.md to mark stack bounds as complete**

In `CODE_REVIEW.md`, find section 8.1 and update Fix #4:

```markdown
#### Fix #4: Stack Bounds Validation (§2.2.2) ✅ **COMPLETE**
**Status:** Implemented in stack-bounds-validation branch

**Problem:** VM allocated fixed 64KB stack segment but didn't enforce bounds checking on stack pointer changes. Programs could move SP outside stack segment, potentially corrupting code/data segments.

**Solution:** Implemented proactive bounds validation in SetSP and SetSPWithTrace:
- Validate SP stays within [0x00040000, 0x00050000] range
- Return descriptive errors for underflow/overflow
- Updated all call sites to handle errors
- Updated ~40 test files to use valid stack addresses

**Testing:**
- 6 new unit tests for bounds validation (3 for SetSP, 3 for SetSPWithTrace)
- 2 new tests for InitializeStack error handling
- All 1,024+ existing tests updated and passing
- Verified error propagation through VM state

**Impact:** Prevents stack pointer from corrupting other memory segments. Stack violations now halt VM immediately with descriptive error messages.
```

**Step 2: Update PROGRESS.md**

Add entry to PROGRESS.md:

```markdown
## November 14, 2025 - Stack Bounds Validation

**Feature:** Proactive stack pointer bounds validation

**Implementation:**
- Modified `SetSP()` and `SetSPWithTrace()` to validate SP stays within stack segment (0x00040000-0x00050000)
- Return descriptive errors for stack underflow/overflow
- Updated `InitializeStack()`, `Bootstrap()`, and `Reset()` to propagate errors
- Updated all instruction implementations that modify SP
- Fixed ~40 test files to use valid stack addresses

**Testing:**
- Added 8 new unit tests for bounds validation and error handling
- All 1,024+ existing tests passing
- Verified error propagation through all code paths

**Impact:**
- Prevents stack pointer corruption of other memory segments
- Stack violations now halt VM with descriptive errors
- Follows VM's integrity error handling pattern
```

**Step 3: Commit documentation updates**

```bash
git add CODE_REVIEW.md PROGRESS.md
git commit -m "docs: update documentation for stack bounds validation

- Mark Fix #4 as complete in CODE_REVIEW.md
- Add implementation summary to PROGRESS.md"
```

---

## Task 16: Run Final Validation

**Files:**
- None (verification only)

**Step 1: Clear cache and run full test suite**

```bash
go clean -testcache
go test ./... -v
```

Expected: All tests pass (1,024+)

**Step 2: Run linter**

```bash
golangci-lint run ./...
```

Expected: No issues

**Step 3: Format code**

```bash
go fmt ./...
```

**Step 4: Build project**

```bash
go build -o arm-emulator
```

Expected: Clean build

**Step 5: Run example program to verify runtime behavior**

```bash
./arm-emulator examples/hello.s
```

Expected: Program runs successfully

**Step 6: Test error case with a program that corrupts stack**

Create `test_stack_corruption.s`:
```assembly
.text
.global _start

_start:
    ; Try to set SP to invalid address (below stack segment)
    MOV SP, #0x1000    ; Invalid: below 0x00040000
    SWI #0x00          ; Exit - should never reach this
```

Run:
```bash
./arm-emulator test_stack_corruption.s
```

Expected output:
```
Error: stack underflow: SP=0x00001000 is below stack minimum (0x00040000)
```

---

## Task 17: Final Commit and Summary

**Step 1: Review all commits**

```bash
git log --oneline
```

Verify commits follow conventional commit format:
- `test:` for test-only changes
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation

**Step 2: Create summary commit if needed**

If there are loose changes, commit them:

```bash
git add -A
git commit -m "chore: final cleanup for stack bounds validation"
```

**Step 3: Tag the completion**

```bash
git tag -a stack-bounds-validation-complete -m "Complete implementation of proactive stack bounds validation

- SetSP and SetSPWithTrace now validate SP stays in [0x00040000, 0x00050000]
- All call sites updated with error handling
- All 1,024+ tests passing
- Resolves CODE_REVIEW.md Fix #4 (§2.2.2)"
```

**Step 4: Generate summary for PR/merge**

Create summary in terminal:

```bash
cat << 'EOF'
# Stack Bounds Validation - Implementation Complete

## Summary
Implemented proactive stack pointer bounds validation to prevent stack overflow/underflow memory corruption.

## Changes
- Modified SetSP() and SetSPWithTrace() to validate SP range [0x00040000, 0x00050000]
- Updated InitializeStack(), Bootstrap(), Reset() with error propagation
- Updated all instruction implementations that modify SP
- Fixed ~40 test files to use valid stack addresses

## Testing
- Added 8 new unit tests for bounds validation
- All 1,024+ tests passing
- Verified error messages and propagation

## Impact
- Stack violations now halt VM immediately with descriptive errors
- Prevents stack from corrupting code/data segments
- Follows VM's integrity error handling pattern
- Resolves CODE_REVIEW.md Fix #4 (§2.2.2)

## Commands to Verify
```bash
go clean -testcache && go test ./...
golangci-lint run ./...
go build -o arm-emulator
```
EOF
```

---

## Execution Checklist

Use this checklist to track progress:

- [ ] Task 1: Write failing tests for SetSP
- [ ] Task 2: Implement SetSP with bounds validation
- [ ] Task 3: Write failing tests for SetSPWithTrace
- [ ] Task 4: Implement SetSPWithTrace with bounds validation
- [ ] Task 5: Update InitializeStack call site
- [ ] Task 6: Update Bootstrap call site
- [ ] Task 7: Update Reset call site
- [ ] Task 8: Update remaining VM call sites
- [ ] Task 9: Update instruction implementation call sites
- [ ] Task 10: Fix test files - batch 1 (tests/unit/vm/)
- [ ] Task 11: Fix test files - batch 2 (tests/unit/instructions/)
- [ ] Task 12: Fix test files - batch 3 (tests/integration/)
- [ ] Task 13: Update stack trace tests
- [ ] Task 14: Run full test suite and fix remaining issues
- [ ] Task 15: Update documentation (CODE_REVIEW.md, PROGRESS.md)
- [ ] Task 16: Run final validation (tests, lint, build, runtime)
- [ ] Task 17: Final commit and summary

---

## Troubleshooting

### Issue: Compilation errors about SetSP signature mismatch

**Cause:** Call sites not updated to handle error return value

**Fix:** Search for the call site and add error handling:
```go
if err := vm.CPU.SetSP(value); err != nil {
    return err  // or handle appropriately
}
```

### Issue: Test fails with "stack underflow" or "stack overflow"

**Cause:** Test using invalid SP address

**Fix:** Update test to use valid address in range [0x00040000, 0x00050000]:
```go
// Use stack segment start + offset
validSP := vm.StackSegmentStart + 0x1000
```

### Issue: Stack trace tests fail after validation added

**Cause:** Stack trace tests may intentionally use invalid SP to test detection

**Fix:** Either:
1. Update tests to expect validation errors earlier in the call chain
2. Use direct register assignment to bypass validation for testing

### Issue: Integration test fails with stack error

**Cause:** Test setup not using Bootstrap, or using manual SP initialization

**Fix:** Ensure test uses Bootstrap or InitializeStack:
```go
vm := NewVM()
err := vm.Bootstrap(program, options...)
require.NoError(t, err)
```

---

## Expected Outcomes

After completing this plan:

1. **Security:** Stack pointer cannot corrupt other memory segments
2. **Reliability:** Stack violations halt VM immediately with clear error messages
3. **Consistency:** All SetSP operations follow same validation pattern
4. **Testing:** All 1,024+ tests passing with proper error handling
5. **Documentation:** CODE_REVIEW.md updated to reflect completion
6. **Code Quality:** Clean build, no lint issues, follows project conventions

---

## Estimated Time

- Tasks 1-7: Core implementation - 2 hours
- Tasks 8-9: VM and instruction call sites - 1 hour
- Tasks 10-14: Test file updates - 3-4 hours
- Tasks 15-17: Documentation and validation - 1 hour

**Total: 7-8 hours** for complete implementation

---

## References

- **CODE_REVIEW.md §2.2.2:** Original issue description with detailed analysis
- **CODE_REVIEW.md §8.1:** Fix #4 implementation status (currently deferred)
- **vm/syscall.go:16-34:** VM's two-tier error handling philosophy
- **vm/memory.go:425-473:** Example of bounds validation pattern (Allocate function)
- **vm/constants.go:207-216:** Stack segment constants

---

**Plan created:** 2025-11-14
**For feature:** Stack Bounds Validation
**Resolves:** CODE_REVIEW.md Fix #4 (§2.2.2)
