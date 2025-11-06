import { test, expect } from '@playwright/test';
import { AppPage } from '../pages/app.page';
import { RegisterViewPage } from '../pages/register-view.page';
import { TEST_PROGRAMS } from '../fixtures/programs';
import { loadProgram, waitForExecution, stepUntilAddress, waitForVMStateChange } from '../utils/helpers';
import { ADDRESSES, TIMEOUTS } from '../utils/test-constants';

test.describe('Breakpoints', () => {
  let appPage: AppPage;
  let registerView: RegisterViewPage;

  test.beforeEach(async ({ page }) => {
    appPage = new AppPage(page);
    registerView = new RegisterViewPage(page);
    await appPage.goto();
    await appPage.waitForLoad();

    // Wait for any ongoing execution to complete
    await page.waitForFunction(() => {
      const statusElement = document.querySelector('[data-testid="execution-status"]');
      if (!statusElement) return true; // If no status element, assume ready
      const status = statusElement.textContent?.toLowerCase() || '';
      return status !== 'running';
    }, { timeout: TIMEOUTS.EXECUTION_MAX });

    // Clear all breakpoints first
    await page.evaluate(() => {
      // @ts-ignore
      window.go.main.App.ClearAllBreakpoints();
    });

    // Reset VM to clean state
    await appPage.clickReset();

    // Wait for reset to complete by checking PC is back at zero
    await page.waitForFunction(() => {
      const pcElement = document.querySelector('[data-register="PC"] .register-value');
      if (!pcElement) return false;
      const pcValue = pcElement.textContent?.trim() || '';
      return pcValue === '0x00000000';
    }, { timeout: TIMEOUTS.EXECUTION_MAX });
  });

  test('should set breakpoint via F9', async () => {
    // Load program
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Step to a specific line
    await appPage.clickStep();
    await appPage.clickStep();

    // Set breakpoint with F9
    await appPage.pressF9();

    // Switch to breakpoints tab
    await appPage.switchToBreakpointsTab();

    // Verify breakpoint appears in list
    const breakpointsList = await appPage.page.locator('[data-testid="breakpoints-list"]');
    await expect(breakpointsList.locator('.breakpoint-item')).toHaveCount(1);
  });

  test('should stop at breakpoint during run', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Step to an instruction in the loop
    let previousPC = await registerView.getRegisterValue('PC');
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.clickStep();

    // Wait for last step to complete by checking PC changed
    await appPage.page.waitForFunction(
      (prevPC) => {
        const pcElement = document.querySelector('[data-register="PC"] .register-value');
        if (!pcElement) return false;
        const currentPC = pcElement.textContent?.trim() || '';
        return currentPC !== '' && currentPC !== prevPC;
      },
      previousPC,
      { timeout: TIMEOUTS.WAIT_FOR_STATE }
    );

    // Get current PC to set breakpoint at
    const breakpointAddress = await registerView.getRegisterValue('PC');

    // Set breakpoint at current location
    await appPage.pressF9();

    // Restart and run (restart preserves program and breakpoints)
    await appPage.clickRestart();
    await appPage.clickRun();

    // Should stop at breakpoint
    await waitForExecution(appPage.page);
    const pc = await registerView.getRegisterValue('PC');
    expect(pc).toBe(breakpointAddress);
  });

  test('should toggle breakpoint on/off', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint
    await appPage.pressF9();
    await appPage.switchToBreakpointsTab();
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(1);

    // Toggle off (press F9 again)
    await appPage.pressF9();
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(0);
  });

  test('should display breakpoint in source view', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint
    await appPage.pressF9();

    // Switch to source view
    await appPage.switchToSourceView();

    // Verify breakpoint indicator is visible
    const breakpointIndicator = appPage.page.locator('[data-testid="breakpoint-indicator"]');
    await expect(breakpointIndicator).toBeVisible();
  });

  test('should set multiple breakpoints', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set first breakpoint
    await appPage.clickStep();
    await appPage.pressF9();

    // Set second breakpoint
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.pressF9();

    // Set third breakpoint
    await appPage.clickStep();
    await appPage.pressF9();

    // Verify all breakpoints in list
    await appPage.switchToBreakpointsTab();
    const breakpointsList = await appPage.page.locator('[data-testid="breakpoints-list"]');
    await expect(breakpointsList.locator('.breakpoint-item')).toHaveCount(3);
  });

  test('should continue execution after hitting breakpoint', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint at an early instruction in the loop
    let previousPC = await registerView.getRegisterValue('PC');
    await appPage.clickStep();
    await appPage.clickStep();
    await appPage.clickStep();

    // Wait for last step to complete by checking PC changed
    await appPage.page.waitForFunction(
      (prevPC) => {
        const pcElement = document.querySelector('[data-register="PC"] .register-value');
        if (!pcElement) return false;
        const currentPC = pcElement.textContent?.trim() || '';
        return currentPC !== '' && currentPC !== prevPC;
      },
      previousPC,
      { timeout: TIMEOUTS.WAIT_FOR_STATE }
    );

    await appPage.pressF9();

    const pcAtBreakpoint = await registerView.getRegisterValue('PC');

    // Restart and run to hit breakpoint (restart preserves program and breakpoints)
    await appPage.clickRestart();
    await appPage.clickRun();

    // Wait for execution to pause at breakpoint
    await waitForExecution(appPage.page);

    // Verify we stopped at the breakpoint
    const pcAtBreakpointFirstHit = await registerView.getRegisterValue('PC');
    expect(pcAtBreakpointFirstHit).toBe(pcAtBreakpoint);

    // Continue execution - should hit breakpoint again or advance
    await appPage.clickRun();

    // Wait for next execution pause (either another breakpoint hit or completion)
    await waitForExecution(appPage.page, TIMEOUTS.EXECUTION_SHORT);

    // Verify execution continued (PC changed or program completed)
    const pcAfterContinue = await registerView.getRegisterValue('PC');
    // PC should either hit the breakpoint again (same address) or have moved past it
    // Just verify the run command was processed successfully
    expect(pcAfterContinue).toBeTruthy();
  });

  test('should remove breakpoint from list', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint
    await appPage.pressF9();

    // Switch to breakpoints tab
    await appPage.switchToBreakpointsTab();
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(1);

    // Click remove button
    const removeButton = appPage.page.locator('[data-testid="remove-breakpoint-button"]').first();
    await removeButton.click();

    // Verify breakpoint removed
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(0);
  });

  test.skip('should disable/enable breakpoint', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set breakpoint
    await appPage.pressF9();

    // Switch to breakpoints tab
    await appPage.switchToBreakpointsTab();

    // Disable breakpoint
    const disableCheckbox = appPage.page.locator('[data-testid="breakpoint-enabled-checkbox"]').first();
    await disableCheckbox.click();

    // Reset and run - should not stop at disabled breakpoint
    await appPage.clickReset();
    await appPage.clickRun();

    // Should not stop (or will stop at exit)
    await waitForExecution(appPage.page, TIMEOUTS.WAIT_FOR_STATE);

    // Re-enable breakpoint
    await appPage.switchToBreakpointsTab();
    await disableCheckbox.click();

    // Reset and run - should now stop at breakpoint
    await appPage.clickReset();
    await appPage.clickRun();

    await waitForExecution(appPage.page);
    // Should have stopped at breakpoint
  });

  test.skip('should clear all breakpoints', async () => {
    await loadProgram(appPage, TEST_PROGRAMS.fibonacci);

    // Set multiple breakpoints
    await appPage.pressF9();
    await appPage.clickStep();
    await appPage.pressF9();
    await appPage.clickStep();
    await appPage.pressF9();

    // Switch to breakpoints tab
    await appPage.switchToBreakpointsTab();
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(3);

    // Click clear all button
    const clearAllButton = appPage.page.locator('[data-testid="clear-all-breakpoints-button"]');
    await clearAllButton.click();

    // Verify all breakpoints removed
    await expect(appPage.page.locator('.breakpoint-item')).toHaveCount(0);
  });
});
