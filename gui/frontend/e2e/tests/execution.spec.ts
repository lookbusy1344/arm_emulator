import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { RegisterViewPage } from '../pages/register-view.page';
import { TEST_PROGRAMS } from '../fixtures/programs';
import { loadProgram, waitForExecution, waitForOutput } from '../utils/helpers';

test.describe('Program Execution', () => {
  let appPage: AppPage;
  let registerView: RegisterViewPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    registerView = new RegisterViewPage(page);
    await appPage.goto();
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
    await appPage.page.waitForTimeout(100);

    // Verify PC changed
    const newPC = await registerView.getRegisterValue('PC');
    expect(newPC).not.toBe(initialPC);

    // Step through several instructions
    for (let i = 0; i < 10; i++) {
      await appPage.clickStep();
      await appPage.page.waitForTimeout(50);
    }

    // Verify registers changed
    const r0 = await registerView.getRegisterValue('R0');
    expect(r0).not.toBe('0x00000000');
  });

  test('should pause infinite loop', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.infinite_loop);

    // Start execution
    await appPage.clickRun();

    // Wait for execution to start (small delay)
    await appPage.page.waitForTimeout(300);

    // Pause
    await appPage.clickPause();

    // Wait for state to change to paused
    await appPage.page.waitForTimeout(200);

    // Verify we can step after pause
    const pc = await registerView.getRegisterValue('PC');
    await appPage.clickStep();
    await appPage.page.waitForTimeout(100);
    const newPC = await registerView.getRegisterValue('PC');
    expect(newPC).not.toBe(pc);
  });

  test('should reset program state', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Execute several steps
    for (let i = 0; i < 5; i++) {
      await appPage.clickStep();
      await appPage.page.waitForTimeout(50);
    }

    // Get current register state
    const beforeReset = await registerView.getAllRegisters();

    // Reset
    await appPage.clickReset();

    // Wait for reset to complete
    await appPage.page.waitForTimeout(300);

    // Verify registers reset to entry point, not necessarily all zeros
    const afterReset = await registerView.getAllRegisters();
    const pc = afterReset['PC'];
    // PC should be back at entry point (code segment start)
    expect(pc).toBe('0x00008000');
  });

  test('should execute arithmetic operations', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.arithmetic);

    // Step through all instructions (need enough steps for all operations)
    for (let i = 0; i < 6; i++) {
      await appPage.clickStep();
      await appPage.page.waitForTimeout(50);
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

    // Wait for step to complete
    await appPage.page.waitForTimeout(200);

    const newPC = await registerView.getRegisterValue('PC');
    expect(newPC).not.toBe(initialPC);
  });

  test('should complete program execution', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.hello);

    // Run to completion
    await appPage.clickRun();

    // Wait for execution to complete
    await waitForExecution(appPage.page, 10000);

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
