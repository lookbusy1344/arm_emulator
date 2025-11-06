import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { loadProgram, waitForVMStateChange } from '../utils/helpers';
import { TIMEOUTS } from '../utils/test-constants';

test.describe('Smoke Tests', () => {
  let appPage: AppPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    await appPage.goto();
    await appPage.waitForLoad();
  });

  test('should load the application', async ({ page }) => {
    await expect(page).toHaveTitle(/ARM Emulator/);
  });

  test('should display all main UI elements', async () => {
    // Check toolbar buttons
    await expect(appPage.loadButton).toBeVisible();
    await expect(appPage.stepButton).toBeVisible();
    await expect(appPage.stepOverButton).toBeVisible();
    await expect(appPage.runButton).toBeVisible();
    await expect(appPage.resetButton).toBeVisible();

    // Check tabs
    await expect(appPage.sourceTab).toBeVisible();
    await expect(appPage.disassemblyTab).toBeVisible();
    await expect(appPage.outputTab).toBeVisible();
    await expect(appPage.breakpointsTab).toBeVisible();

    // Check views
    await expect(appPage.registerView).toBeVisible();
    await expect(appPage.memoryView).toBeVisible();
    await expect(appPage.stackView).toBeVisible();
  });

  test('should switch between tabs', async () => {
    await appPage.switchToDisassemblyView();
    await expect(appPage.disassemblyView).toBeVisible();

    await appPage.switchToSourceView();
    await expect(appPage.sourceView).toBeVisible();

    await appPage.switchToBreakpointsTab();
    await expect(appPage.breakpointsTab).toHaveClass(/active/);
  });

  test('should respond to keyboard shortcuts', async () => {
    // Load a simple program to test keyboard shortcuts
    const simpleProgram = `
      .text
      .global _start
    _start:
      MOV R0, #10
      MOV R1, #20
      ADD R2, R0, R1
      SWI #0x00
    `;

    await loadProgram(appPage, simpleProgram);

    // Get initial PC
    const initialPC = await appPage.getRegisterValue('PC');

    // F11 = Step - verify PC changed
    await appPage.pressF11();
    await waitForVMStateChange(appPage.page);
    const pcAfterF11 = await appPage.getRegisterValue('PC');
    expect(pcAfterF11).not.toBe(initialPC);

    // F10 = Step Over - verify PC changed
    const pcBeforeF10 = await appPage.getRegisterValue('PC');
    await appPage.pressF10();
    await waitForVMStateChange(appPage.page);
    const pcAfterF10 = await appPage.getRegisterValue('PC');
    expect(pcAfterF10).not.toBe(pcBeforeF10);

    // Reset before testing F5
    await appPage.clickReset();
    await waitForVMStateChange(appPage.page, TIMEOUTS.VM_RESET);

    // F5 = Run - verify execution started and completed
    await appPage.pressF5();

    // Wait for execution to complete
    await appPage.page.waitForFunction(
      () => {
        const statusElement = document.querySelector('[data-testid="execution-status"]');
        if (!statusElement) return false;
        const status = statusElement.textContent?.toLowerCase() || '';
        return status === 'halted' || status === 'exited';
      },
      { timeout: TIMEOUTS.EXECUTION_SHORT }
    );

    // Verify program ran (should be at exit)
    const pcAfterRun = await appPage.getRegisterValue('PC');
    expect(pcAfterRun).not.toBe('0x00008000');
  });
});
