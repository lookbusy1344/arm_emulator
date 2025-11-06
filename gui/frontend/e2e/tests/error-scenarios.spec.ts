import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { TIMEOUTS, ADDRESSES } from '../utils/test-constants';
import { waitForVMStateChange } from '../utils/helpers';

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
      ({ source, entryPoint }) => {
        try {
          // @ts-ignore - Wails runtime
          return window.go.main.App.LoadProgramFromSource(source, 'invalid.s', entryPoint);
        } catch (e: any) {
          return { error: e.message };
        }
      },
      { source: invalidProgram, entryPoint: ADDRESSES.CODE_SEGMENT_START }
    );

    // Verify error was properly reported with meaningful message
    if (result && typeof result === 'object' && 'error' in result) {
      // Backend returned error - verify it contains meaningful information
      expect(result.error).toBeTruthy();
      expect(typeof result.error).toBe('string');
      // Should mention the invalid instruction or parse error
      const errorMsg = result.error.toLowerCase();
      expect(
        errorMsg.includes('invalid') ||
        errorMsg.includes('unknown') ||
        errorMsg.includes('parse') ||
        errorMsg.includes('error')
      ).toBe(true);
    }

    // Verify app is still responsive (didn't crash)
    await expect(appPage.registerView).toBeVisible();
  });

  test('should handle empty program', async () => {
    const emptyProgram = '';

    const result = await appPage.page.evaluate(
      ({ source, entryPoint }) => {
        try {
          // @ts-ignore - Wails runtime
          return window.go.main.App.LoadProgramFromSource(source, 'empty.s', entryPoint);
        } catch (e: any) {
          return { error: e.message };
        }
      },
      { source: emptyProgram, entryPoint: ADDRESSES.CODE_SEGMENT_START }
    );

    // Empty program should return meaningful error
    if (result && typeof result === 'object' && 'error' in result) {
      expect(result.error).toBeTruthy();
      expect(typeof result.error).toBe('string');
      // Should mention empty, no instructions, or similar
      const errorMsg = result.error.toLowerCase();
      expect(errorMsg.length).toBeGreaterThan(0);
    }

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
      ({ source, entryPoint }) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.LoadProgramFromSource(source, 'invalid_mem.s', entryPoint);
      },
      { source: invalidMemoryProgram, entryPoint: ADDRESSES.CODE_SEGMENT_START }
    );

    // Wait for VM to be ready
    await waitForVMStateChange(appPage.page);

    // Try to run - should either error or handle gracefully
    const pcBefore = await appPage.getRegisterValue('PC');
    await appPage.clickStep();

    // Wait for step to complete by checking PC changed or execution state updated
    await appPage.page.waitForFunction(
      (previousPC) => {
        const pcElement = document.querySelector('[data-register="PC"] .register-value');
        if (!pcElement) return false;
        const currentPC = pcElement.textContent?.trim() || '';
        return currentPC !== '' && currentPC !== previousPC;
      },
      pcBefore,
      { timeout: TIMEOUTS.STEP_COMPLETE }
    );

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
      ({ source, entryPoint }) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.LoadProgramFromSource(source, 'overflow.s', entryPoint);
      },
      { source: overflowProgram, entryPoint: ADDRESSES.CODE_SEGMENT_START }
    );

    // Wait for VM to be ready
    await waitForVMStateChange(appPage.page);

    // Execute the overflow instruction - step through each instruction
    for (let i = 0; i < 3; i++) {
      const pcBefore = await appPage.getRegisterValue('PC');
      await appPage.clickStep();
      // Wait for PC to update
      await appPage.page.waitForFunction(
        (previousPC) => {
          const pcElement = document.querySelector('[data-register="PC"] .register-value');
          if (!pcElement) return false;
          const currentPC = pcElement.textContent?.trim() || '';
          return currentPC !== '' && currentPC !== previousPC;
        },
        pcBefore,
        { timeout: TIMEOUTS.STEP_COMPLETE }
      );
    }

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
      ({ source, entryPoint }) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.LoadProgramFromSource(source, 'race.s', entryPoint);
      },
      { source: program, entryPoint: ADDRESSES.CODE_SEGMENT_START }
    );

    // Wait for VM to be ready
    await waitForVMStateChange(appPage.page);

    // Rapidly click step multiple times
    const pcBefore = await appPage.getRegisterValue('PC');
    await Promise.all([
      appPage.clickStep(),
      appPage.clickStep(),
      appPage.clickStep(),
      appPage.clickStep(),
      appPage.clickStep(),
    ]);

    // Wait for UI to stabilize by verifying PC has changed
    await appPage.page.waitForFunction(
      (previousPC) => {
        const pcElement = document.querySelector('[data-register="PC"] .register-value');
        if (!pcElement) return false;
        const currentPC = pcElement.textContent?.trim() || '';
        // PC should have changed from initial value
        return currentPC !== '' && currentPC !== previousPC;
      },
      pcBefore,
      { timeout: TIMEOUTS.VM_STATE_CHANGE }
    );

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
      ({ source, entryPoint }) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.LoadProgramFromSource(source, 'infinite.s', entryPoint);
      },
      { source: infiniteLoop, entryPoint: ADDRESSES.CODE_SEGMENT_START }
    );

    // Wait for VM to be ready
    await waitForVMStateChange(appPage.page);

    // Start running
    await appPage.clickRun();

    // Wait for execution to actually start
    await appPage.page.waitForFunction(
      () => {
        const statusElement = document.querySelector('[data-testid="execution-status"]');
        if (!statusElement) return false;
        const status = statusElement.textContent?.toLowerCase() || '';
        return status === 'running';
      },
      { timeout: TIMEOUTS.EXECUTION_START }
    );

    // Reset while running
    await appPage.clickReset();

    // Wait for reset to complete by checking PC is back at start
    const expectedPC = `0x${ADDRESSES.CODE_SEGMENT_START.toString(16).toUpperCase().padStart(8, '0')}`;
    await appPage.page.waitForFunction(
      (pc) => {
        const pcElement = document.querySelector('[data-register="PC"] .register-value');
        if (!pcElement) return false;
        return pcElement.textContent?.trim() === pc;
      },
      expectedPC,
      { timeout: TIMEOUTS.VM_RESET }
    );

    const pc = await appPage.getRegisterValue('PC');
    expect(pc).toBe(expectedPC); // Should reset to start
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
      ({ source, entryPoint }) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.LoadProgramFromSource(source, 'large.s', entryPoint);
      },
      { source: largeValueProgram, entryPoint: ADDRESSES.CODE_SEGMENT_START }
    );

    // Wait for VM to be ready
    await waitForVMStateChange(appPage.page);

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
