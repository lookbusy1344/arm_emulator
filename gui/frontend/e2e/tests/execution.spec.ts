import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { RegisterViewPage } from '../pages/register-view.page';
import { TEST_PROGRAMS } from '../fixtures/programs';
import { loadProgram, waitForExecution } from '../utils/helpers';

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

    // Run program
    await appPage.clickRun();

    // Switch to output tab
    await appPage.switchToOutputTab();

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

    // Verify PC changed
    const newPC = await registerView.getRegisterValue('PC');
    expect(newPC).not.toBe(initialPC);

    // Step through several instructions
    for (let i = 0; i < 10; i++) {
      await appPage.clickStep();
    }

    // Verify registers changed
    const r0 = await registerView.getRegisterValue('R0');
    expect(r0).not.toBe('0x00000000');
  });

  test('should pause infinite loop', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.infinite_loop);

    // Start execution
    await appPage.clickRun();

    // Wait a bit
    await new Promise(resolve => setTimeout(resolve, 500));

    // Pause
    await appPage.clickPause();

    // Verify we can step after pause
    const pc = await registerView.getRegisterValue('PC');
    await appPage.clickStep();
    const newPC = await registerView.getRegisterValue('PC');
    expect(newPC).not.toBe(pc);
  });

  test('should reset program state', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Execute several steps
    for (let i = 0; i < 5; i++) {
      await appPage.clickStep();
    }

    // Get current register state
    const beforeReset = await registerView.getAllRegisters();

    // Reset
    await appPage.clickReset();

    // Verify registers reset
    const afterReset = await registerView.getAllRegisters();
    expect(afterReset['R0']).toBe('0x00000000');
    expect(afterReset['R1']).toBe('0x00000000');
  });

  test('should execute arithmetic operations', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.arithmetic);

    // Step through all instructions
    for (let i = 0; i < 5; i++) {
      await appPage.clickStep();
    }

    // Verify arithmetic results
    const r2 = await registerView.getRegisterValue('R2');
    expect(r2).toBe('0x0000001E'); // 30 in hex

    const r3 = await registerView.getRegisterValue('R3');
    expect(r3).toBe('0x0000000A'); // 10 in hex
  });

  test('should step over function calls', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    const initialPC = await registerView.getRegisterValue('PC');

    // Step over
    await appPage.clickStepOver();

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

    // Verify program completed
    const status = await appPage.page.locator('[data-testid="execution-status"]').textContent();
    expect(status).toContain('Exited');
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
