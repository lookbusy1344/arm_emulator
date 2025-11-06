import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { TIMEOUTS } from '../utils/test-constants';

test.describe('Error Scenarios', () => {
  let appPage: AppPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    await appPage.goto();
    await appPage.waitForLoad();
  });

  test('should handle program with syntax errors', async () => {
    const invalidProgram = `
      .text
      .global _start
    _start:
      INVALID_INSTRUCTION R0, R1  // This should cause a parse error
      MOV R0, #10
      SWI #0x00
    `;

    // Attempt to load invalid program
    const result = await appPage.page.evaluate(
      ({ source }) => {
        try {
          // @ts-ignore - Wails runtime
          return window.go.main.App.LoadProgramFromSource(source, 'invalid.s', 0x00008000);
        } catch (e: any) {
          return { error: e.message };
        }
      },
      { source: invalidProgram }
    );

    // Should either return an error or backend should handle gracefully
    // At minimum, verify app doesn't crash
    await expect(appPage.registerView).toBeVisible();
  });

  test('should handle empty program', async () => {
    const emptyProgram = '';

    const result = await appPage.page.evaluate(
      ({ source }) => {
        try {
          // @ts-ignore - Wails runtime
          return window.go.main.App.LoadProgramFromSource(source, 'empty.s', 0x00008000);
        } catch (e: any) {
          return { error: e.message };
        }
      },
      { source: emptyProgram }
    );

    // Verify UI remains stable
    await expect(appPage.toolbar).toBeVisible();
  });

  test('should handle program that attempts invalid memory access', async () => {
    const invalidMemoryProgram = `
      .text
      .global _start
    _start:
      MOV R0, #0xFFFFFFFF   // Max address
      LDR R1, [R0]          // Attempt to load from invalid address
      SWI #0x00
    `;

    await appPage.page.evaluate(
      ({ source }) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.LoadProgramFromSource(source, 'invalid_mem.s', 0x00008000);
      },
      { source: invalidMemoryProgram }
    );

    await appPage.page.waitForTimeout(TIMEOUTS.VM_STATE_CHANGE);

    // Try to run - should either error or handle gracefully
    await appPage.clickStep();
    await appPage.page.waitForTimeout(TIMEOUTS.STEP_COMPLETE);

    // Verify app is still responsive
    const pc = await appPage.getRegisterValue('PC');
    expect(pc).toBeTruthy();
  });

  test('should handle division by zero', async () => {
    // Note: ARM doesn't have native division, but we can test multiply overflow
    const overflowProgram = `
      .text
      .global _start
    _start:
      MOV R0, #0xFFFFFFFF   // Max value
      MOV R1, #0xFFFFFFFF   // Max value
      MUL R2, R0, R1        // This will overflow
      SWI #0x00
    `;

    await appPage.page.evaluate(
      ({ source }) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.LoadProgramFromSource(source, 'overflow.s', 0x00008000);
      },
      { source: overflowProgram }
    );

    await appPage.page.waitForTimeout(TIMEOUTS.VM_STATE_CHANGE);

    // Execute the overflow instruction
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.page.waitForTimeout(TIMEOUTS.STEP_COMPLETE);

    // Should complete without crashing
    const r2 = await appPage.getRegisterValue('R2');
    expect(r2).toBeTruthy();
  });

  test('should handle clicking step without program loaded', async () => {
    // No program loaded - try to step
    await appPage.clickStep();

    // Should not crash
    await expect(appPage.registerView).toBeVisible();
  });

  test('should handle clicking run without program loaded', async () => {
    // No program loaded - try to run
    await appPage.clickRun();

    // Should not crash
    await expect(appPage.toolbar).toBeVisible();
  });

  test('should handle setting breakpoint without program loaded', async () => {
    // Try to set breakpoint without program
    await appPage.pressF9();

    // Should not crash
    await expect(appPage.breakpointsTab).toBeVisible();
  });

  test('should handle rapid button clicks (race conditions)', async () => {
    const program = `
      .text
      .global _start
    _start:
      MOV R0, #0
    loop:
      ADD R0, R0, #1
      CMP R0, #100
      BLT loop
      SWI #0x00
    `;

    await appPage.page.evaluate(
      ({ source }) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.LoadProgramFromSource(source, 'race.s', 0x00008000);
      },
      { source: program }
    );

    await appPage.page.waitForTimeout(TIMEOUTS.VM_STATE_CHANGE);

    // Rapidly click step multiple times
    await Promise.all([
      appPage.clickStep(),
      appPage.clickStep(),
      appPage.clickStep(),
      appPage.clickStep(),
      appPage.clickStep(),
    ]);

    // Wait for UI to stabilize
    await appPage.page.waitForTimeout(TIMEOUTS.UI_STABILIZE * 3);

    // Should still be responsive
    const pc = await appPage.getRegisterValue('PC');
    expect(pc).toBeTruthy();
  });

  test('should handle switching tabs rapidly', async () => {
    // Rapidly switch between tabs
    for (let i = 0; i < 5; i++) {
      await appPage.switchToSourceView();
      await appPage.switchToDisassemblyView();
      await appPage.switchToOutputTab();
      await appPage.switchToBreakpointsTab();
      await appPage.switchToStatusTab();
    }

    // Should not crash
    await expect(appPage.toolbar).toBeVisible();
  });

  test('should handle resetting during execution', async () => {
    const infiniteLoop = `
      .text
      .global _start
    _start:
      MOV R0, #0
    loop:
      ADD R0, R0, #1
      B loop
    `;

    await appPage.page.evaluate(
      ({ source }) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.LoadProgramFromSource(source, 'infinite.s', 0x00008000);
      },
      { source: infiniteLoop }
    );

    await appPage.page.waitForTimeout(TIMEOUTS.VM_STATE_CHANGE);

    // Start running
    await appPage.clickRun();
    await appPage.page.waitForTimeout(TIMEOUTS.EXECUTION_START);

    // Reset while running
    await appPage.clickReset();

    // Should handle gracefully
    await appPage.page.waitForTimeout(TIMEOUTS.VM_RESET);

    const pc = await appPage.getRegisterValue('PC');
    expect(pc).toBe('0x00008000'); // Should reset to start
  });

  test('should handle very large immediate values', async () => {
    const largeValueProgram = `
      .text
      .global _start
    _start:
      LDR R0, =0xFFFFFFFF   // Max 32-bit value
      LDR R1, =0x80000000   // Min signed 32-bit value
      ADD R2, R0, R1
      SWI #0x00
    `;

    await appPage.page.evaluate(
      ({ source }) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.LoadProgramFromSource(source, 'large.s', 0x00008000);
      },
      { source: largeValueProgram }
    );

    await appPage.page.waitForTimeout(TIMEOUTS.VM_STATE_CHANGE);

    // Execute
    await appPage.clickRun();
    await appPage.page.waitForFunction(
      () => {
        const statusElement = document.querySelector('[data-testid="execution-status"]');
        if (!statusElement) return false;
        const status = statusElement.textContent?.toLowerCase() || '';
        return status === 'halted' || status === 'exited';
      },
      { timeout: TIMEOUTS.EXECUTION_SHORT }
    );

    // Should complete successfully
    const r2 = await appPage.getRegisterValue('R2');
    expect(r2).toBeTruthy();
  });
});
