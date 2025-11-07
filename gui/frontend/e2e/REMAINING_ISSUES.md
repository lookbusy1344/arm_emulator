# E2E Testing Infrastructure - Remaining Issues

This document tracks issues identified in the e2e testing infrastructure audit that require follow-up work.

## HIGH PRIORITY

### 1. Replace Remaining Hardcoded Waits
**Status:** Partially complete
**Files affected:**
- `execution.spec.ts` - ~10 instances of `waitForTimeout()`
- `breakpoints.spec.ts` - ~8 instances
- `memory.spec.ts` - ~6 instances
- `visual.spec.ts` - ~5 instances with 1-2 second waits

**Action needed:**
Replace all `page.waitForTimeout()` calls with proper `waitForFunction()` or event-based waiting.

**Estimate:** 4-6 hours

---

### 2. Add Verification to Cleanup Operations
**Status:** Not started
**Files affected:** All `beforeEach` hooks in test files

**Problem:**
Tests attempt to reset VM and clear breakpoints but don't verify operations succeeded:
```typescript
await appPage.clickReset();
await page.waitForTimeout(200);  // Hope it worked!
```

**Action needed:**
- Verify reset succeeded by checking PC returned to entry point
- Verify breakpoints cleared by querying breakpoint list
- Add `verifyCleanState()` helper function

**Estimate:** 2-3 hours

---

### 3. Improve Test Isolation
**Status:** Not started
**Risk:** Medium

**Problem:**
Tests run serially but share same VM instance. State corruption in one test could affect subsequent tests.

**Action needed:**
- Add comprehensive state verification in `beforeEach`
- Consider adding VM health check before each test
- Add test-level timeouts to prevent hung tests from blocking suite

**Estimate:** 3-4 hours

---

## MEDIUM PRIORITY

### 4. Complete Skipped Tests
**Status:** Not started
**Tests to implement:**
1. `memory.spec.ts:48` - "should scroll through memory"
2. `breakpoints.spec.ts:180` - "should disable/enable breakpoint"
3. `breakpoints.spec.ts:212` - "should clear all breakpoints"
4. `visual.spec.ts:267,285` - Dark/light mode tests (blocked by feature implementation)

**Action needed:**
- Implement scroll test with virtual scrolling support
- Add breakpoint enable/disable UI and tests
- Add clear all breakpoints button and test
- Implement dark mode and visual tests

**Estimate:** 6-8 hours

---

### 5. Strengthen Assertions
**Status:** Not started
**Files affected:** Multiple test files

**Problem:**
Weak assertions that don't verify actual behavior:
```typescript
expect(flags).toBeDefined();  // Checks nothing meaningful
expect(pcAfterContinue).toBeTruthy();  // Just checks not empty
```

**Action needed:**
Replace with specific value checks:
```typescript
expect(flags).toMatchObject({ N: false, Z: true, C: false, V: false });
expect(pcAfterContinue).toMatch(/0x[0-9A-F]{8}/);
```

**Estimate:** 2-3 hours

---

### 6. Replace Magic Numbers in Test Files
**Status:** Constants created, not yet used in all files
**Files affected:**
- `execution.spec.ts`
- `breakpoints.spec.ts`
- `memory.spec.ts`
- `visual.spec.ts`
- `examples.spec.ts`

**Action needed:**
Import and use constants from `test-constants.ts` throughout test files.

**Estimate:** 2-3 hours

---

### 7. Add Backend Health Checks to CI
**Status:** Not started
**File:** `.github/workflows/e2e-tests.yml:65-77`

**Problem:**
Workflow only checks if port 34115 responds, not if Wails is actually ready.

**Action needed:**
- Add Wails health check endpoint
- Verify endpoint returns valid state before running tests
- Add retry logic with exponential backoff

**Estimate:** 2-3 hours

---

## LOW PRIORITY

### 8. Add Missing Test Categories

#### Accessibility Tests
**Status:** Not started
**Tools:** `@axe-core/playwright`
**Estimate:** 4-6 hours

Add WCAG compliance testing for:
- Keyboard navigation
- Screen reader support
- Color contrast
- Focus indicators

#### Performance Tests
**Status:** Not started
**Estimate:** 3-4 hours

Add tests for:
- Initial load time
- Step execution speed
- Memory view rendering performance
- Large program handling

#### Security Tests
**Status:** Not started
**Estimate:** 4-6 hours

Add tests for:
- XSS in program source display
- Malicious assembly input
- Memory boundary violations
- Buffer overflow attempts

---

### 9. Parameterized Register Tests
**Status:** Not started
**File:** Various test files
**Estimate:** 1-2 hours

Replace individual register checks with parameterized tests:
```typescript
for (const reg of REGISTERS.GENERAL) {
  test(`should update ${reg} correctly`, async () => {
    // test logic
  });
}
```

---

### 10. Expand Test Fixtures
**Status:** Only 4 programs exist
**File:** `e2e/fixtures/programs.ts`
**Estimate:** 2-3 hours

Add test programs for:
- Syntax errors (intentional)
- Programs that crash
- Memory corruption scenarios
- Stack overflow
- Large programs (stress test with 1000+ instructions)

---

### 11. Add Performance Monitoring
**Status:** Not started
**Estimate:** 4-6 hours

Track test health metrics:
- Flakiness rate per test
- Average execution time
- Retry frequency
- Failure patterns

Tools: Custom reporter or test analytics service

---

### 12. Improve Page Object Return Values
**Status:** Not started
**Files:** `e2e/pages/*.page.ts`
**Estimate:** 2-3 hours

**Problem:**
Page object methods return void, making it hard to verify operations.

**Action needed:**
Return useful values:
```typescript
async clickStep(): Promise<boolean> {
  await this.stepButton.click();
  return await this.verifyStepCompleted();
}
```

---

## TECHNICAL DEBT

### 13. CI Browser Matrix
**Status:** Fixed for macOS and Linux
**Remaining:** Windows testing

**Current matrix:** 6 combinations (macOS + Linux × 3 browsers)
**Missing:** Windows runners (adds 3 more combinations)

**Note:** Windows runners are more expensive in CI. Consider adding only if needed.

**Estimate:** 1 hour to add, but increases CI time/cost

---

### 14. Visual Regression Baselines
**Status:** May need regeneration
**Files:** `e2e/tests/visual.spec.ts-snapshots/`

**Action after tightening tolerances:**
Visual tests may fail with tighter tolerances. Need to:
1. Run visual tests locally
2. Review differences
3. Update baselines if changes are acceptable
4. Commit new baselines

**Estimate:** 1-2 hours

---

### 15. Test Documentation
**Status:** Minimal
**Estimate:** 2-3 hours

Add comprehensive test documentation:
- What each test suite covers
- How to debug failing tests
- How to update visual baselines
- How to add new tests
- Common pitfalls and solutions

---

## SUMMARY

**Total estimated work:** 50-70 hours

**Breakdown by priority:**
- High priority: 9-13 hours
- Medium priority: 18-24 hours
- Low priority: 19-27 hours
- Technical debt: 4-6 hours

**Recommended approach:**
1. Complete high priority items first (hardcoded waits, verification, isolation)
2. Address medium priority items as time permits
3. Low priority items can be addressed incrementally over time

**Next steps:**
1. Create GitHub issues for each item
2. Prioritize based on team capacity
3. Schedule work across sprints
4. Track progress and update this document
