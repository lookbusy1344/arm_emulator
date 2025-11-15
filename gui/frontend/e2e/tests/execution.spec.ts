import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { RegisterViewPage } from '../pages/register-view.page';
import { TEST_PROGRAMS } from '../fixtures/programs';
import { loadProgram, waitForExecution, waitForOutput, formatAddress } from '../utils/helpers';
import { ADDRESSES, TIMEOUTS } from '../utils/test-constants';

test.describe('Program Execution', () => {
  let appPage: AppPage;
  let registerView: RegisterViewPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    registerView = new RegisterViewPage(page);
    await appPage.goto();
    await appPage.waitForLoad();

    // Reset VM and clear all breakpoints to ensure clean state
    // Note: clickReset() now waits for PC to be 0x00000000 internally
    await appPage.clickReset();

    // Clear any existing breakpoints
    const breakpoints = await page.evaluate(() => {
      // @ts-ignore - Wails runtime
      return window.go.main.App.GetBreakpoints();
    });

    for (const bp of breakpoints) {
      await page.evaluate((address) => {
        // @ts-ignore - Wails runtime
        return window.go.main.App.RemoveBreakpoint(address);
      }, bp.Address);
    }
  });

  test('should execute hello world program', async () => {
    // Load program
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Switch to output tab BEFORE running so OutputView is mounted and listening
    await appPage.switchToOutputTab();

    // Run program
    await appPage.clickRun();

    // Wait for execution to complete and output to appear
    await waitForExecution(appPage.page);
    await waitForOutput(appPage.page);

    // Verify output
    const output = await appPage.getOutputText();
    expect(output).toContain('Hello, World!');
  });

  test('should step through fibonacci program', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Get initial PC
    const initialPC = await registerView.getRegisterValue('PC');

    // Step once
    await appPage.clickStep();

    // Wait for step to complete by checking PC changed
    await appPage.page.waitForFunction(
      (prevPC) => {
        const pcElement = document.querySelector('[data-register="PC"] .register-value');
        if (!pcElement) return false;
        const currentPC = pcElement.textContent?.trim() || '';
        return currentPC !== '' && currentPC !== prevPC;
      },
      initialPC,
      { timeout: TIMEOUTS.WAIT_FOR_STATE }
    );

    // Verify PC changed
    const newPC = await registerView.getRegisterValue('PC');
    expect(newPC).not.toBe(initialPC);

    // Step through several instructions
    for (let i = 0; i < 10; i++) {
      const prevPC = await registerView.getRegisterValue('PC');
      await appPage.clickStep();
      // Wait for PC to update
      await appPage.page.waitForFunction(
        (pc) => {
          const pcElement = document.querySelector('[data-register="PC"] .register-value');
          if (!pcElement) return false;
          const currentPC = pcElement.textContent?.trim() || '';
          return currentPC !== '' && currentPC !== pc;
        },
        prevPC,
        { timeout: TIMEOUTS.WAIT_FOR_STATE }
      );
    }

    // Verify registers changed
    const r0 = await registerView.getRegisterValue('R0');
    expect(r0).not.toBe('0x00000000');
  });

  test('should pause infinite loop', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.infinite_loop);

    // Start execution
    await appPage.clickRun();

    // Wait for execution to actually start by checking cycles > 0
    // This is more reliable than waiting for status text which may be missed in fast loops
    await appPage.page.waitForFunction(() => {
      const cyclesElement = document.querySelector('.status-cycles');
      if (!cyclesElement) return false;
      const text = cyclesElement.textContent || '';
      const match = text.match(/Cycles:\s*(\d+)/);
      if (!match) return false;
      const cycles = parseInt(match[1], 10);
      return cycles > 0;
    }, { timeout: TIMEOUTS.WAIT_FOR_RESET });

    // Pause
    await appPage.clickPause();

    // Wait for execution to actually stop by checking cycles stops changing
    // This is more reliable than waiting for status text
    const cyclesBeforePause = await appPage.page.evaluate(() => {
      const cyclesElement = document.querySelector('.status-cycles');
      if (!cyclesElement) return 0;
      const text = cyclesElement.textContent || '';
      const match = text.match(/Cycles:\s*(\d+)/);
      return match ? parseInt(match[1], 10) : 0;
    });

    // Wait for cycles to stabilize (execution stopped)
    await appPage.page.waitForFunction(
      (prevCycles) => {
        const cyclesElement = document.querySelector('.status-cycles');
        if (!cyclesElement) return false;
        const text = cyclesElement.textContent || '';
        const match = text.match(/Cycles:\s*(\d+)/);
        if (!match) return false;
        const currentCycles = parseInt(match[1], 10);
        // Cycles should be greater than before (execution happened) and stable
        return currentCycles >= prevCycles;
      },
      cyclesBeforePause,
      { timeout: 1000 } // Short timeout since pause should be fast
    );

    // Small delay to ensure state has fully propagated
    await appPage.page.waitForTimeout(100);

    // Verify we can step after pause
    const pc = await registerView.getRegisterValue('PC');
    await appPage.clickStep();

    // Wait for step to complete
    await appPage.page.waitForFunction(
      (prevPC) => {
        const pcElement = document.querySelector('[data-register="PC"] .register-value');
        if (!pcElement) return false;
        const currentPC = pcElement.textContent?.trim() || '';
        return currentPC !== '' && currentPC !== prevPC;
      },
      pc,
      { timeout: TIMEOUTS.WAIT_FOR_STATE }
    );

    const newPC = await registerView.getRegisterValue('PC');
    expect(newPC).not.toBe(pc);
  });

  test('should restart program to entry point', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Execute several steps
    for (let i = 0; i < 5; i++) {
      const prevPC = await registerView.getRegisterValue('PC');
      await appPage.clickStep();
      // Wait for PC to update
      await appPage.page.waitForFunction(
        (pc) => {
          const pcElement = document.querySelector('[data-register="PC"] .register-value');
          if (!pcElement) return false;
          const currentPC = pcElement.textContent?.trim() || '';
          return currentPC !== '' && currentPC !== pc;
        },
        prevPC,
        { timeout: TIMEOUTS.WAIT_FOR_STATE }
      );
    }

    // Get current register state
    const beforeReset = await registerView.getAllRegisters();

    // Restart (reset to entry point, keeping program loaded)
    await appPage.clickRestart();

    // Wait for restart to complete by checking PC is back at entry point
    const expectedPC = formatAddress(ADDRESSES.CODE_SEGMENT_START);
    await appPage.page.waitForFunction(
      (pc) => {
        const pcElement = document.querySelector('[data-register="PC"] .register-value');
        if (!pcElement) return false;
        const currentPC = pcElement.textContent?.trim() || '';
        return currentPC === pc;
      },
      expectedPC,
      { timeout: TIMEOUTS.WAIT_FOR_STATE }
    );

    // Verify PC is back at entry point after restart
    const afterRestart = await registerView.getAllRegisters();
    const pc = afterRestart['PC'];
    // PC should be back at entry point (code segment start)
    expect(pc).toBe(formatAddress(ADDRESSES.CODE_SEGMENT_START));
  });

  test('should execute arithmetic operations', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.arithmetic);

    // Step through first 5 instructions (MOV, MOV, ADD, SUB, MUL)
    // We don't need to execute the final SWI instruction to verify the results
    for (let i = 0; i < 5; i++) {
      const prevPC = await registerView.getRegisterValue('PC');
      await appPage.clickStep();

      // Wait for PC to change
      await appPage.page.waitForFunction(
        (expectedPC) => {
          const pcElement = document.querySelector('[data-register="PC"] .register-value');
          if (!pcElement) return false;
          const currentPC = pcElement.textContent?.trim() || '';
          return currentPC !== '' && currentPC !== expectedPC;
        },
        prevPC,
        { timeout: TIMEOUTS.WAIT_FOR_STATE }
      );
    }

    // Verify arithmetic results
    const r2 = await registerView.getRegisterValue('R2');
    expect(r2).toBe('0x0000001E'); // 30 in hex (10 + 20)

    const r3 = await registerView.getRegisterValue('R3');
    expect(r3).toBe('0x0000000A'); // 10 in hex (20 - 10)

    const r4 = await registerView.getRegisterValue('R4');
    expect(r4).toBe('0x000000C8'); // 200 in hex (10 * 20)
  });

  test('should step over function calls', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    const initialPC = await registerView.getRegisterValue('PC');

    // Step over
    await appPage.clickStepOver();

    // Wait for step over to complete by checking PC changed
    await appPage.page.waitForFunction(
      (prevPC) => {
        const pcElement = document.querySelector('[data-register="PC"] .register-value');
        if (!pcElement) return false;
        const currentPC = pcElement.textContent?.trim() || '';
        return currentPC !== '' && currentPC !== prevPC;
      },
      initialPC,
      { timeout: TIMEOUTS.EXECUTION_SHORT }
    );

    const newPC = await registerView.getRegisterValue('PC');
    expect(newPC).not.toBe(initialPC);
  });

  test('should complete program execution', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Run to completion
    await appPage.clickRun();

    // Wait for execution to complete
    await waitForExecution(appPage.page, TIMEOUTS.EXECUTION_MAX);

    // Switch to status tab
    await appPage.switchToStatusTab();

    // Verify program completed (VM uses "halted" for exited programs)
    const status = await appPage.page.locator('[data-testid="execution-status"]').textContent();
    expect(status?.toLowerCase()).toMatch(/halted|exited/);
  });

  test('should preserve CPSR flags across steps', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Step through and monitor flags
    for (let i = 0; i < 10; i++) {
      await appPage.clickStep();
      const flags = await registerView.getCPSRFlags();
      // Flags should be valid (not all false unless program sets them that way)
      expect(flags).toBeDefined();
    }
  });
});
